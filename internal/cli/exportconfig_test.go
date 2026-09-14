package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"solace/internal/engine"
)

// Fixtures and doubles for `broker perform export-config` / `import-config`
// (exportconfig.go, plus the wiring in commands.go/helpers.go/ops_k8s.go/
// ops_container.go).
//
// Both commands talk to the broker over the CLI-over-exec channel (broker.Ops,
// internal/broker/importops.go): upload a script, run `cli -Apes`, read stdout.
// A canned opRunner (cli_test.go) answers the two probes that channel makes -- a
// `show current-config all` of the TARGET (its VPN list, read through the
// artifact's own parser rather than the `show message-vpn *` table, and led by the
// login banner the target's broker type is read off) and the broker-generated
// teardown script -- so ImportPlan/ImportApply run against real business logic
// without a live broker. The banner was a third probe of its own until it became
// free off the first.
// Anything this function does not recognise falls through to (nil, nil), the same
// empty answer engine.Echo
// gives, which is what makes an uncanned step fail exactly the way an unanswered
// real probe would (runCapture's "the broker returned no output" refusal).
func exportconfigTransportOutput(rr *opRunner, targetVPNs []string) func(opCall) []byte {
	return func(c opCall) []byte {
		switch {
		case opArgvMatch(c, ".plan-vpns.cli"):
			// The banner LEADS this transcript because a real `cli -Apes` echoes it
			// ahead of any output, and the target's broker type is now read off this
			// very response rather than from a probe of its own. A fixture without
			// one reads back as an unknown type and the import is refused before it
			// starts.
			return append([]byte(exportconfigBannerOutput), exportconfigVPNListOutput(targetVPNs...)...)
		case opArgvMatch(c, ".export-teardown.cli"):
			return []byte(exportconfigTeardownScript)
		case opArgvMatch(c, ".verify-export.cli"):
			return []byte(exportconfigVerifyCapture)
		case opArgvMatch(c, ".export-config.cli"):
			return []byte(exportconfigNoVPNArtifact)
		case opArgvMatch(c, exportconfigDriverName):
			return exportconfigDriverOutput(rr)
		}
		return nil
	}
}

// newExportconfigRunner builds the canned runner these tests drive import-config
// through.
//
// The two-step wiring is not incidental: the apply driver's answer can only be
// produced by reading the driver script back out of the runner's OWN recorded
// upload, so the responder needs a reference to the runner that will hold it.
func newExportconfigRunner(targetVPNs []string) *opRunner {
	rr := &opRunner{}
	rr.output = exportconfigTransportOutput(rr, targetVPNs)
	return rr
}

// newExportconfigRunnerRejectingChunk builds a runner whose broker REJECTS a line in
// the first chunk, so a test can drive the failure path end to end.
func newExportconfigRunnerRejectingChunk(targetVPNs []string) *opRunner {
	rr := &opRunner{}
	inner := exportconfigTransportOutput(rr, targetVPNs)
	rr.output = func(c opCall) []byte {
		if opArgvMatch(c, exportconfigDriverName) {
			// What the real driver prints when `stop-on-error` stopped chunk 0: the
			// echoed tail, the FAIL marker, and NO completion marker, because it
			// exits there.
			return []byte("SOLACE-UTIL-DETAIL-BEGIN\n" +
				"error: invalid command at line 3\n" +
				"SOLACE-UTIL-DETAIL-END\n" +
				"SOLACE-UTIL-CHUNK-FAIL 0\n")
		}
		return inner(c)
	}
	return rr
}

// exportconfigDriverName is the basename runDriver uploads the generated apply
// script to (internal/broker: shellScriptPath("solace-util-import")).
const exportconfigDriverName = ".solace-util-import"

// Marker lines the apply driver prints, mirroring the constants in
// internal/broker/driver.go. They are duplicated rather than exported because they
// are a protocol between that file and its own parser; if the two ever drift, every
// import test here fails loudly on the next run, which is the point.
const (
	exportconfigMarkerOK   = "SOLACE-UTIL-CHUNK-OK"
	exportconfigMarkerDone = "SOLACE-UTIL-DONE"
)

// exportconfigDriverOutput answers the one `sh <driver>` exec an import makes, with
// the markers a real driver prints when every chunk applied.
//
// It derives the chunk COUNT from the uploaded script rather than assuming one,
// because the parser rejects a marker naming a chunk that does not exist -- so a
// double that guessed would make these tests pass or fail on the guess rather than
// on the behaviour.
func exportconfigDriverOutput(rr *opRunner) []byte {
	var body string
	for i := len(rr.calls) - 1; i >= 0; i-- {
		if rr.calls[i].stdin != "" && strings.Contains(rr.calls[i].stdin, "# ---- chunk ") {
			body = rr.calls[i].stdin
			break
		}
	}
	var b strings.Builder
	for i := 0; i < strings.Count(body, "# ---- chunk "); i++ {
		fmt.Fprintf(&b, "%s %d\n", exportconfigMarkerOK, i)
	}
	b.WriteString(exportconfigMarkerDone + "\n")
	return []byte(b.String())
}

// exportconfigVerifyCapture is what the TARGET re-exports for ImportVerify, and it
// is deliberately the UNION of both import fixtures' content: the broker-level
// `Configure System` line from exportconfigNoVPNArtifact and the VPN blocks from
// exportconfigOneVPNArtifact. The double cannot know which artifact a given test
// imported, and the diff is directional (every artifact line must be present on the
// target, but the target may carry more), so a superset verifies clean for either.
//
// Verification is the step under which the whole feature stands or falls -- `cli
// -Apes` exits 0 on a rejected line -- so a confirmation test asserting "the import
// proceeded" is asserting that this capture was fetched at all.
const exportconfigVerifyCapture = " ! BEGIN\n" +
	"! Command: \"show current-config all\"\n" +
	"!   SEMP Schema Version: \"broker/99.99.9.9999\"\n" +
	"!   Redacted: \"False\"\n" +
	"\n" +
	"home\nenable\nconfigure\n\n" +
	"! Create All: \"Message Vpns\"\n" +
	"create message-vpn \"fake-import-vpn\"\n" +
	"  exit\n\n" +
	"! Configure System:\n" +
	"system health disk-latency-high-threshold 10000000\n\n" +
	"! Create Message Vpn: \"fake-import-vpn\"\n" +
	"message-vpn \"fake-import-vpn\"\n" +
	"  no shutdown\n" +
	"  radius-domain \"" + exportconfigFixtureSecret + "\"\n" +
	"  exit\n\n" +
	"! END\n"

// exportconfigBannerOutput is the CLI login banner ImportPlan reads the target's
// type off (importops.go's bannerType): the "Event Broker Software" product line
// and the license-software URL both read as a software broker, matching both
// fixtures below. It is prepended to the TARGET capture rather than answered on a
// probe of its own, because that is what a real `cli -Apes` does -- it echoes the
// banner ahead of whatever the script printed.
const exportconfigBannerOutput = "Solace Event Broker Software Enterprise Version 10.26.0.8827\n\n" +
	"located at http://www.solace.com/license-software\n"

// exportconfigVPNListOutput renders the capture ImportPlan reads the TARGET's
// message-VPNs from. It is a `show current-config all` capture, not a
// `show message-vpn *` table: readTarget (internal/broker/importops.go) reads the
// create-all-VPNs pass through the same quote-aware parser that reads the
// artifact, because the table parser it replaced split the name column on
// whitespace and so filed an existing VPN whose name contains spaces as new.
// No args means the target has no message-VPNs.
func exportconfigVPNListOutput(vpns ...string) []byte {
	var b strings.Builder
	b.WriteString(" ! BEGIN\n")
	b.WriteString("! Command: \"show current-config all\"\n\n")
	b.WriteString("home\nenable\nconfigure\n\n")
	b.WriteString("! Create All: \"Message Vpns\"\n")
	for _, v := range vpns {
		b.WriteString("create message-vpn \"" + v + "\"\n  exit\n")
	}
	b.WriteString("\n! END\n")
	return []byte(b.String())
}

// exportconfigTeardownScript stands in for the broker-generated
// `show current-config message-vpn <n> remove` output ImportApply replays
// verbatim; its own content is never parsed, only re-uploaded.
const exportconfigTeardownScript = " ! BEGIN\nhome\nenable\nconfigure\nno message-vpn \"fake-import-vpn\"\nexit\n! END\n"

// exportconfigFixtureSecret is a distinctive, obviously-fake credential-shaped
// value embedded in exportconfigOneVPNArtifact, so TestExportImportSecretsNeverEchoed
// can prove an artifact's own content never reaches stdout/stderr the same way
// smokeAdminPass proves the env file's admin password does not.
const exportconfigFixtureSecret = "FIXTURE-FAKE-SECRET-9f8e7d21"

// exportconfigNoVPNArtifact is a minimal, well-formed `show current-config`
// transcript with no message-VPN at all. It serves two roles: the canned answer
// exportconfigTransportOutput gives for the real export-config command (standing
// in for the broker's own response), and an import artifact whose broker scope
// carries no VPN (TestImportConfigBrokerScopeNeverPrompts) -- both are exactly
// the same shape, a raw current-config capture.
const exportconfigNoVPNArtifact = "! solace-util-export: 1\n! BEGIN\n" +
	"! Command: \"show current-config all\"\n" +
	"!   SEMP Schema Version: \"broker/99.99.9.9999\"\n" +
	"!   Redacted: \"False\"\n" +
	"\n" +
	"home\nenable\nconfigure\n\n" +
	"! Configure System:\n" +
	"system health disk-latency-high-threshold 10000000\n" +
	"\n! END\n"

// exportconfigOneVPNArtifact captures exactly one message-VPN, "fake-import-vpn",
// so a test can drive ImportPlan's Existing/NewVPNs split by varying only what
// the canned capture of the TARGET says it already has.
const exportconfigOneVPNArtifact = "! solace-util-export: 1\n! BEGIN\n" +
	"! Command: \"show current-config all\"\n" +
	"!   SEMP Schema Version: \"broker/99.99.9.9999\"\n" +
	"!   Redacted: \"False\"\n" +
	"\n" +
	"home\nenable\nconfigure\n\n" +
	"! Create All: \"Message Vpns\"\n" +
	"create message-vpn \"fake-import-vpn\"\n" +
	"  exit\n\n" +
	"! Create Message Vpn: \"fake-import-vpn\"\n" +
	"message-vpn \"fake-import-vpn\"\n" +
	"  no shutdown\n" +
	"  radius-domain \"" + exportconfigFixtureSecret + "\"\n" +
	"  exit\n\n" +
	"! END\n"

// writeExportconfigArtifact writes body to a temp file and returns its path, for
// `import-config <file>` to read.
func writeExportconfigArtifact(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "capture.cli")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write import artifact fixture: %v", err)
	}
	return path
}

// exportconfigReadCounter wraps a reader and counts Read calls, so a test can
// prove a confirmation prompt was never actually READ -- stronger than merely
// observing that the command did not block or error.
type exportconfigReadCounter struct {
	io.Reader
	n *int
}

func (r exportconfigReadCounter) Read(p []byte) (int, error) {
	*r.n++
	return r.Reader.Read(p)
}

// runImportConfirmation runs `import-config <artifact>` on the container
// (docker) platform -- the node-local transport, which needs no --pod and so
// keeps every confirmation test about the confirmation and nothing else -- with
// targetVPNs canning the VPNs the target's own capture reports. It returns
// captured stdout, stderr and the run's error alongside the opRunner, so a
// caller can inspect exactly which transport calls were (or were not) made.
func runImportConfirmation(t *testing.T, path, artifact string, targetVPNs []string, extraArgs []string, configure func(*App)) (string, string, error, *opRunner) {
	t.Helper()
	rr := newExportconfigRunner(targetVPNs)
	args := append([]string{"broker", "perform", "import-config", artifact, "--env", path, "--platform", "docker"}, extraArgs...)
	var out string
	var err error
	stderrText := captureStderr(t, func() {
		out, err = runRootWith(t, args, func(a *App) {
			a.NewRunner = func(*App) engine.Runner { return rr }
			if configure != nil {
				configure(a)
			}
		})
	})
	return out, stderrText, err, rr
}

// TestExportConfigScopeConflictIsUsageError covers the --vpn/--broker-only
// conflict end to end through the command, not just exportScope in isolation.
//
// RISK (CLAUDE.md S5): bad input must be rejected AND the downstream call must
// not be made -- a usage error that still reached the broker would be worse than
// no validation at all, since it would look like the flag combination was
// silently reinterpreted as something else.
func TestExportConfigScopeConflictIsUsageError(t *testing.T) {
	rr := &opRunner{}
	_, err := runRootWith(t, withEnv("broker", "perform", "export-config", "--vpn", "acme", "--broker-only", "--platform", "kubernetes"),
		func(a *App) { a.NewRunner = func(*App) engine.Runner { return rr } })
	if err == nil || ExitCode(err) != exitUsage {
		t.Fatalf("export-config --vpn x --broker-only err = %v (exit %d), want a usage error (exit %d)",
			err, ExitCode(err), exitUsage)
	}
	if len(rr.calls) != 0 {
		t.Errorf("the flag conflict still reached the transport:\n%s", rr.dump())
	}
}

// TestExportConfigLeafRunsOnAllPlatforms proves the leaf is wired -- reaches its
// real per-platform handler and issues the broker CLI invocation -- on all three
// platforms. Under engine.Echo the capture is necessarily empty (Echo's Output
// always returns (nil, nil)), so the command fails with runCapture's own
// "no output" refusal; that failure, reached identically on every platform, is
// itself the proof the wiring is not platform-specific.
func TestExportConfigLeafRunsOnAllPlatforms(t *testing.T) {
	for _, platform := range []string{"kubernetes", "docker", "podman"} {
		t.Run(platform, func(t *testing.T) {
			out, err := runRootWith(t, withEnv("broker", "perform", "export-config", "--platform", platform), echoRunner)
			if err == nil || !strings.Contains(err.Error(), "the broker returned no output") {
				t.Fatalf("%s export-config (echo runner) err = %v, want the empty-capture refusal", platform, err)
			}
			if !strings.Contains(out, "-Apes") {
				t.Errorf("%s export-config stdout = %q, want the echoed `cli -Apes` invocation", platform, out)
			}
		})
	}
}

// TestImportConfigLeafRunsOnAllPlatforms is export-config's counterpart, and it
// now fails in exactly the same way for exactly the same reason. Reading a
// well-formed artifact succeeds, but the first thing ImportPlan asks the TARGET
// for is its own capture, and under engine.Echo that answer is empty -- so
// runCapture's "no output" refusal is what comes back, identically on every
// platform, which is what proves each one reached the same real handler.
//
// It used to refuse with "cannot tell which broker type the target is" instead,
// because the target's type came from a dedicated banner round trip that ran
// FIRST and came back empty. That probe is gone -- the type is read off this very
// capture -- so the empty capture is now caught before anything can be concluded
// from it, one refusal earlier and strictly more accurate about what went wrong.
// The type refusal is still reachable and still tested, by a fake that answers
// the capture with output carrying no login banner.
func TestImportConfigLeafRunsOnAllPlatforms(t *testing.T) {
	artifact := writeExportconfigArtifact(t, exportconfigOneVPNArtifact)
	for _, platform := range []string{"kubernetes", "docker", "podman"} {
		t.Run(platform, func(t *testing.T) {
			out, err := runRootWith(t, withEnv("broker", "perform", "import-config", artifact, "--platform", platform), echoRunner)
			// readTarget's own wrapper as well as the refusal: the bare refusal is
			// what export-config's sibling test above asserts too, so pinning it
			// alone would leave this test satisfiable by the wrong round trip.
			if err == nil ||
				!strings.Contains(err.Error(), "read the target's configuration to list its message-VPNs") ||
				!strings.Contains(err.Error(), "the broker returned no output") {
				t.Fatalf("%s import-config (echo runner) err = %v, want readTarget's empty-capture refusal", platform, err)
			}
			if !strings.Contains(out, "-Apes") {
				t.Errorf("%s import-config stdout = %q, want the echoed `cli -Apes` invocation", platform, out)
			}
		})
	}
}

// TestExportImportPodFlagPlatformScope covers --pod's platform scoping
// (addPodFlag/checkFlagPlatforms) for both leaves: on Kubernetes it names the
// pod; on docker/podman there is one container per host and an
// accepted-and-ignored flag is worse than a refused one, so it is a usage error
// there instead.
func TestExportImportPodFlagPlatformScope(t *testing.T) {
	artifact := writeExportconfigArtifact(t, exportconfigOneVPNArtifact)

	rejected := []struct {
		name string
		args []string
	}{
		{"export-config --pod on docker", []string{"broker", "perform", "export-config", "--pod", "primary", "--platform", "docker"}},
		{"export-config --pod on podman", []string{"broker", "perform", "export-config", "--pod", "primary", "--platform", "podman"}},
		{"import-config --pod on docker", []string{"broker", "perform", "import-config", artifact, "--pod", "primary", "--platform", "docker"}},
		{"import-config --pod on podman", []string{"broker", "perform", "import-config", artifact, "--pod", "primary", "--platform", "podman"}},
	}
	for _, tc := range rejected {
		t.Run(tc.name, func(t *testing.T) {
			rr := &opRunner{}
			_, err := runRootWith(t, withEnv(tc.args...), func(a *App) { a.NewRunner = func(*App) engine.Runner { return rr } })
			if err == nil || ExitCode(err) != exitUsage || !strings.Contains(err.Error(), "--pod is not supported on") {
				t.Fatalf("%s err = %v (exit %d), want a --pod usage refusal (exit %d)", tc.name, err, ExitCode(err), exitUsage)
			}
			if len(rr.calls) != 0 {
				t.Errorf("%s: --pod was refused but the transport still saw %d call(s):\n%s", tc.name, len(rr.calls), rr.dump())
			}
		})
	}

	t.Run("export-config --pod is accepted on kubernetes", func(t *testing.T) {
		out, err := runRootWith(t, withEnv("broker", "perform", "export-config", "--pod", "primary", "--platform", "kubernetes"), echoRunner)
		if err != nil && strings.Contains(err.Error(), "--pod") {
			t.Fatalf("export-config --pod on kubernetes err = %v, want --pod accepted", err)
		}
		if !strings.Contains(out, "-Apes") {
			t.Errorf("export-config --pod on kubernetes stdout = %q, want it to have reached the transport", out)
		}
	})

	t.Run("import-config --pod is accepted on kubernetes", func(t *testing.T) {
		out, err := runRootWith(t, withEnv("broker", "perform", "import-config", artifact, "--pod", "primary", "--platform", "kubernetes"), echoRunner)
		if err != nil && strings.Contains(err.Error(), "--pod") {
			t.Fatalf("import-config --pod on kubernetes err = %v, want --pod accepted", err)
		}
		if !strings.Contains(out, "-Apes") {
			t.Errorf("import-config --pod on kubernetes stdout = %q, want it to have reached the transport", out)
		}
	})
}

// TestExportConfigWithoutOutWritesOnlyArtifactToStdout is the piped-artifact
// safety property: with no --out, the artifact IS stdout, so a Section header
// (the "=== ... ===" shape internal/output uses for every report) mixed in would
// corrupt anything piping it into `import-config` or a file. Narration goes to
// stderr unconditionally, so nothing is lost either way -- runExport's own
// comment states exactly this trade.
func TestExportConfigWithoutOutWritesOnlyArtifactToStdout(t *testing.T) {
	path := writeStandaloneEnv(t)
	rr := newExportconfigRunner(nil)

	var out string
	var err error
	stderrText := captureStderr(t, func() {
		out, err = runRootWith(t, []string{"broker", "perform", "export-config", "--env", path, "--platform", "kubernetes"},
			func(a *App) { a.NewRunner = func(*App) engine.Runner { return rr } })
	})
	if err != nil {
		t.Fatalf("export-config err = %v, want nil", err)
	}
	if out == "" {
		t.Fatal("export-config without --out produced empty stdout")
	}
	if strings.Contains(out, "===") {
		t.Errorf("export-config without --out put a report header on stdout, which would corrupt a piped artifact:\n%s", out)
	}
	if !strings.Contains(stderrText, "==>") {
		t.Errorf("export-config narration did not reach stderr: %q", stderrText)
	}
}

// TestExportConfigWithOutWritesFile0600AndReportsOnStdout is --out's own shape:
// once it has freed stdout, the report (Section + KVBlock) goes there instead,
// and the artifact itself -- as sensitive as the env file, per exportconfig.go's
// own comment -- is written to a file nothing but its owner can read.
func TestExportConfigWithOutWritesFile0600AndReportsOnStdout(t *testing.T) {
	path := writeStandaloneEnv(t)
	outFile := filepath.Join(t.TempDir(), "capture.cli")
	rr := newExportconfigRunner(nil)

	out, err := runRootWith(t, []string{"broker", "perform", "export-config", "--out", outFile, "--env", path, "--platform", "kubernetes"},
		func(a *App) { a.NewRunner = func(*App) engine.Runner { return rr } })
	if err != nil {
		t.Fatalf("export-config --out err = %v, want nil", err)
	}
	if !strings.Contains(out, "Configuration export") {
		t.Errorf("export-config --out stdout = %q, want the export report", out)
	}
	body, readErr := os.ReadFile(outFile)
	if readErr != nil {
		t.Fatalf("read %s: %v", outFile, readErr)
	}
	if len(body) == 0 {
		t.Fatal("export-config --out wrote an empty file")
	}
	if runtime.GOOS != "windows" {
		// Windows has no POSIX permission bits for os.WriteFile's mode argument to
		// set, so the assertion is meaningless there -- the repo's other file-mode
		// checks skip it the same way.
		info, statErr := os.Stat(outFile)
		if statErr != nil {
			t.Fatalf("stat %s: %v", outFile, statErr)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("export-config --out file mode = %o, want 0600", perm)
		}
	}
}

// TestExportConfigOutOverwriteGate covers emitOrWrite's overwrite question over
// an existing --out path, all three answers: refused (and the file left alone)
// when there is nobody to ask, replaced when the question is asked and answered
// yes, replaced with --no-prompt. The same gate examples's own --out takes
// (TestExamplesRefusesToOverwrite), here for the command whose artifact is as
// sensitive as the env file itself.
func TestExportConfigOutOverwriteGate(t *testing.T) {
	path := writeStandaloneEnv(t)
	outFile := filepath.Join(t.TempDir(), "existing.cli")
	if err := os.WriteFile(outFile, []byte("PLACEHOLDER"), 0o644); err != nil {
		t.Fatalf("seed existing --out file: %v", err)
	}
	rr := newExportconfigRunner(nil)

	t.Run("refuses without confirmation and leaves the file alone", func(t *testing.T) {
		var err error
		stderrText := captureStderr(t, func() {
			_, err = runRootWith(t, []string{"broker", "perform", "export-config", "--out", outFile, "--env", path, "--platform", "kubernetes"},
				func(a *App) {
					a.Interactive = func() bool { return false }
					a.NewRunner = func(*App) engine.Runner { return rr }
				})
		})
		// The refusal formats the path with %q, so on Windows the message carries
		// ESCAPED separators (C:\\Users\\...) while outFile holds single ones.
		// Comparing against the quoted form is what makes this assertion portable --
		// a raw Contains passes on Linux and fails here for no real reason.
		if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("%q", outFile)) {
			t.Fatalf("export-config --out over an existing file err = %v, want a refusal naming %q", err, outFile)
		}
		if !strings.Contains(stderrText, "--no-prompt") {
			t.Errorf("the refusal did not name --no-prompt: %q", stderrText)
		}
		if body, readErr := os.ReadFile(outFile); readErr != nil || string(body) != "PLACEHOLDER" {
			t.Errorf("a declined overwrite still changed the file: body=%q err=%v", body, readErr)
		}
	})

	// The half the operator actually hit: they were at a terminal, and the command
	// refused instead of asking. --no-prompt is not the answer to that -- being ASKED
	// and saying yes is, so it is pinned on its own rather than left implied by the
	// refusal case above.
	t.Run("an answered yes replaces it", func(t *testing.T) {
		if err := os.WriteFile(outFile, []byte("PLACEHOLDER"), 0o644); err != nil {
			t.Fatalf("reseed existing --out file: %v", err)
		}
		var err error
		stderrText := captureStderr(t, func() {
			_, err = runRootWith(t, []string{"broker", "perform", "export-config", "--out", outFile, "--env", path, "--platform", "kubernetes"},
				func(a *App) {
					a.Interactive = func() bool { return true }
					a.PromptIn = strings.NewReader("y\n")
					a.NewRunner = func(*App) engine.Runner { return rr }
				})
		})
		if err != nil {
			t.Fatalf("export-config --out with an answered overwrite err = %v, want nil", err)
		}
		if !strings.Contains(stderrText, "Overwrite") {
			t.Errorf("no overwrite question was asked: %q", stderrText)
		}
		if body, readErr := os.ReadFile(outFile); readErr != nil || string(body) == "PLACEHOLDER" {
			t.Errorf("an accepted overwrite did not replace the file: body=%q err=%v", body, readErr)
		}
	})

	// The artifact carries every configured secret in the broker's encrypted form,
	// so its mode is not cosmetic. os.WriteFile honours its perm argument only when
	// it CREATES the file, so overwriting an existing 0644 path left the capture
	// world-readable while the code two lines above claimed "0o600 on every artifact
	// this writes". POSIX only: Windows file permissions are ACL-based and Chmod
	// there only moves the read-only bit, so the assertion has nothing to check.
	t.Run("an overwritten artifact is still owner-only", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("file modes are not POSIX bits on Windows")
		}
		if err := os.WriteFile(outFile, []byte("PLACEHOLDER"), 0o644); err != nil {
			t.Fatalf("reseed a world-readable --out file: %v", err)
		}
		_, err := runRootWith(t, []string{"broker", "perform", "export-config", "--out", outFile, "--no-prompt", "--env", path, "--platform", "kubernetes"},
			func(a *App) { a.NewRunner = func(*App) engine.Runner { return rr } })
		if err != nil {
			t.Fatalf("export-config over an existing file err = %v, want nil", err)
		}
		info, statErr := os.Stat(outFile)
		if statErr != nil {
			t.Fatalf("stat %s: %v", outFile, statErr)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("mode = %04o, want 0600 -- the credential-bearing artifact stayed group/world readable on the overwrite path", perm)
		}
	})

	t.Run("--no-prompt replaces it", func(t *testing.T) {
		if err := os.WriteFile(outFile, []byte("PLACEHOLDER"), 0o644); err != nil {
			t.Fatalf("reseed existing --out file: %v", err)
		}
		_, err := runRootWith(t, []string{"broker", "perform", "export-config", "--out", outFile, "--no-prompt", "--env", path, "--platform", "kubernetes"},
			func(a *App) { a.NewRunner = func(*App) engine.Runner { return rr } })
		if err != nil {
			t.Fatalf("export-config --out --no-prompt err = %v, want nil", err)
		}
		body, readErr := os.ReadFile(outFile)
		if readErr != nil {
			t.Fatalf("read %s: %v", outFile, readErr)
		}
		if string(body) == "PLACEHOLDER" {
			t.Error("--no-prompt did not replace the existing file")
		}
	})
}

// TestImportConfigMissingFileIsUsageError covers runImport's own read failure:
// the operator chose the path, so a bad one is the operator's mistake, not the
// broker's, and nothing must be attempted against a broker over a file that was
// never read.
func TestImportConfigMissingFileIsUsageError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.cli")
	rr := &opRunner{}
	_, err := runRootWith(t, withEnv("broker", "perform", "import-config", missing, "--platform", "kubernetes"),
		func(a *App) { a.NewRunner = func(*App) engine.Runner { return rr } })
	if err == nil || ExitCode(err) != exitUsage || !strings.Contains(err.Error(), "read the configuration artifact") {
		t.Fatalf("import-config <missing file> err = %v (exit %d), want a usage error naming the read failure",
			err, ExitCode(err))
	}
	if len(rr.calls) != 0 {
		t.Errorf("a missing artifact still reached the transport:\n%s", rr.dump())
	}
}

// TestImportConfigConfirmationMatrix is confirmImport's whole contract
// (exportconfig.go): the strength of the question tracks what an import would
// actually destroy, not merely whether one is happening.
//
// RISK: tearing down an existing message-VPN destroys every message spooled in
// its queues -- the same class of irreversible loss --delete-data asks about --
// so a lenient "y" answering the wrong question here would let a typo empty a
// production VPN. The marker for "the operation proceeded" is the re-export
// ImportVerify always issues afterward (`.verify-export.cli` -- a name of its own,
// so it can no longer be confused with the export command's own capture); the
// marker for "a conflicting VPN was torn down" is the teardown capture ImportApply
// issues only for an existing VPN (`.export-teardown.cli`). A case that should NOT
// proceed is told apart from one that should by whether the call was attempted at
// all, never by whether the whole import finishes cleanly.
// TestImportConfigFailedApplySkipsVerification pins the sequencing rule: the apply
// stops at the first rejected chunk, reports, and the diff does NOT run.
//
// Verifying after a failed apply would be actively unhelpful, not merely wasteful.
// The apply stops at the failing chunk, so every chunk after it never ran -- a diff
// there reports each of them as missing, which is true and says nothing, burying the
// one line that names the actual failure under dozens describing work that was never
// attempted. The operator already has the specific answer.
//
// RISK if this inverts: the diff's verdict becomes the exit code and the chunk
// failure is reported as a verification mismatch, which points at completely the
// wrong recovery.
func TestImportConfigFailedApplySkipsVerification(t *testing.T) {
	path := writeCtrStandaloneEnv(t)
	artifact := writeExportconfigArtifact(t, exportconfigOneVPNArtifact)
	rr := newExportconfigRunnerRejectingChunk(nil)

	var out string
	var err error
	stderrText := captureStderr(t, func() {
		out, err = runRootWith(t, []string{"broker", "perform", "import-config", artifact,
			"--no-prompt", "--env", path, "--platform", "docker"},
			func(a *App) { a.NewRunner = func(*App) engine.Runner { return rr } })
	})

	if err == nil {
		t.Fatal("a rejected chunk must fail the command")
	}
	if rr.hasCall(".verify-export.cli") {
		t.Errorf("the verification re-export ran after a FAILED apply:\n%s", rr.dump())
	}
	// The failure has to be attributable and actionable: which chunk, and what the
	// broker said about it.
	if !strings.Contains(out, "invalid command at line 3") {
		t.Errorf("the broker's own words on the failure were not reported:\n%s", out)
	}
	if !strings.Contains(out+stderrText, "chunk") {
		t.Errorf("the report does not name the failing chunk:\n%s\n%s", out, stderrText)
	}
}

// TestExportConfigRejectsABadVPNNameAsUsage covers the exit code, which is the
// only thing separating "your request was wrong" from "the broker failed".
//
// `--vpn "$VPN"` with VPN unset is the realistic way to get here, and it is a typo,
// not a transient failure. The name used to be validated only inside
// broker.ExportConfig, whose error came back unwrapped -- so a wrapper that retries
// on 1 and refuses to retry on 2 retried a typo forever. Nothing may reach the
// broker either.
func TestExportConfigRejectsABadVPNNameAsUsage(t *testing.T) {
	path := writeCtrStandaloneEnv(t)
	for _, bad := range []string{"", "   ", `evil"vpn`} {
		rr := newExportconfigRunner(nil)
		_, err := runRootWith(t, []string{"broker", "perform", "export-config", "--vpn", bad, "--env", path, "--platform", "docker"},
			func(a *App) { a.NewRunner = func(*App) engine.Runner { return rr } })
		if err == nil {
			t.Fatalf("--vpn %q was accepted", bad)
		}
		if code := ExitCode(err); code != exitUsage {
			t.Errorf("--vpn %q exit code = %d, want %d (usage)", bad, code, exitUsage)
		}
		if len(rr.calls) != 0 {
			t.Errorf("--vpn %q reached the broker before being rejected:\n%s", bad, rr.dump())
		}
	}
}

func TestImportConfigConfirmationMatrix(t *testing.T) {
	path := writeCtrStandaloneEnv(t)
	artifact := writeExportconfigArtifact(t, exportconfigOneVPNArtifact)

	t.Run("new VPN: lenient y proceeds", func(t *testing.T) {
		_, _, _, rr := runImportConfirmation(t, path, artifact, nil, nil, func(a *App) {
			a.Interactive = func() bool { return true }
			a.PromptIn = strings.NewReader("y\n")
		})
		if !rr.hasCall(".verify-export.cli") {
			t.Errorf("a lenient 'y' over a new VPN did not proceed:\n%s", rr.dump())
		}
	})

	t.Run("existing VPN: lenient y is refused by the strict gate", func(t *testing.T) {
		_, _, err, rr := runImportConfirmation(t, path, artifact, []string{"fake-import-vpn"}, nil, func(a *App) {
			a.Interactive = func() bool { return true }
			a.PromptIn = strings.NewReader("y\n")
		})
		if err != nil {
			t.Fatalf("a declined import must not be an error: %v", err)
		}
		if rr.hasCall(".export-teardown.cli") || rr.hasCall(".verify-export.cli") {
			t.Errorf("a lenient 'y' must NOT tear down an existing VPN:\n%s", rr.dump())
		}
	})

	t.Run("existing VPN: exact yes proceeds", func(t *testing.T) {
		_, _, _, rr := runImportConfirmation(t, path, artifact, []string{"fake-import-vpn"}, nil, func(a *App) {
			a.Interactive = func() bool { return true }
			a.PromptIn = strings.NewReader("yes\n")
		})
		if !rr.hasCall(".export-teardown.cli") {
			t.Errorf("an exact 'yes' did not tear down the conflicting VPN:\n%s", rr.dump())
		}
	})

	t.Run("--no-prompt proceeds without a conflict", func(t *testing.T) {
		_, _, _, rr := runImportConfirmation(t, path, artifact, nil, []string{"--no-prompt"}, nil)
		if !rr.hasCall(".verify-export.cli") {
			t.Errorf("--no-prompt did not proceed:\n%s", rr.dump())
		}
	})

	t.Run("--no-prompt proceeds over a conflict", func(t *testing.T) {
		_, _, _, rr := runImportConfirmation(t, path, artifact, []string{"fake-import-vpn"}, []string{"--no-prompt"}, nil)
		if !rr.hasCall(".export-teardown.cli") {
			t.Errorf("--no-prompt did not tear down the conflicting VPN:\n%s", rr.dump())
		}
	})

	t.Run("non-interactive without --no-prompt warns and writes nothing", func(t *testing.T) {
		_, stderrText, err, rr := runImportConfirmation(t, path, artifact, []string{"fake-import-vpn"}, nil, func(a *App) {
			a.Interactive = func() bool { return false }
		})
		if err != nil {
			t.Fatalf("an unanswerable prompt must not fail the command: %v", err)
		}
		if !strings.Contains(stderrText, "--no-prompt") {
			t.Errorf("the refusal did not name --no-prompt: %q", stderrText)
		}
		if rr.hasCall(".export-teardown.cli") || rr.hasCall(".verify-export.cli") {
			t.Errorf("an unanswerable prompt must not touch the broker:\n%s", rr.dump())
		}
	})
}

// TestImportConfigBrokerScopeNeverPrompts covers confirmImport's default case:
// an artifact with no message-VPN at all asks nothing, because which
// broker-level sections apply is a fixed classification (docs/import.md), not a
// per-run decision. Reading the prompt at all -- even to a "y" that would have
// answered it the same way -- would be the wrong contract, so this counts reads
// rather than trusting the outcome.
func TestImportConfigBrokerScopeNeverPrompts(t *testing.T) {
	path := writeCtrStandaloneEnv(t)
	artifact := writeExportconfigArtifact(t, exportconfigNoVPNArtifact)
	rr := newExportconfigRunner(nil)
	reads := 0

	runRootWith(t, []string{"broker", "perform", "import-config", artifact, "--env", path, "--platform", "docker"},
		func(a *App) {
			a.NewRunner = func(*App) engine.Runner { return rr }
			a.Interactive = func() bool { return true }
			a.PromptIn = exportconfigReadCounter{Reader: strings.NewReader("yes\n"), n: &reads}
		})
	if reads != 0 {
		t.Errorf("a broker-scope import read the confirmation prompt %d time(s), want 0", reads)
	}
}

// TestExportImportSecretsNeverEchoed is the S3 smoke check for both leaves,
// mirroring TestSecretsNeverEchoed: a secret-bearing run over the echo seam must
// never put a secret value on stdout or stderr. export-config's own env carries
// smokeAdminPass; import-config's artifact additionally carries
// exportconfigFixtureSecret, standing in for the encrypted credential material a
// real un-redacted capture contains.
func TestExportImportSecretsNeverEchoed(t *testing.T) {
	path := writeStandaloneEnv(t)
	artifact := writeExportconfigArtifact(t, exportconfigOneVPNArtifact)

	var exportOut string
	exportErr := captureStderr(t, func() {
		exportOut, _ = runRootWith(t, []string{"broker", "perform", "export-config", "--env", path, "--platform", "kubernetes"}, echoRunner)
	})
	if strings.Contains(exportOut, smokeAdminPass) || strings.Contains(exportErr, smokeAdminPass) {
		t.Error("export-config leaked the admin password")
	}

	var importOut string
	importErr := captureStderr(t, func() {
		importOut, _ = runRootWith(t, []string{"broker", "perform", "import-config", artifact, "--env", path, "--platform", "kubernetes"}, echoRunner)
	})
	for _, secret := range []string{smokeAdminPass, exportconfigFixtureSecret} {
		if strings.Contains(importOut, secret) || strings.Contains(importErr, secret) {
			t.Errorf("import-config leaked %q", secret)
		}
	}
}

// TestExportScopeFlagCombination covers exportScope directly: the two flags pass
// through unchanged when at most one is set, and the conflict is refused as a
// usage error -- the same pair-reading shape wantRemove/wantEnable already use,
// so a caller need not re-derive it.
func TestExportScopeFlagCombination(t *testing.T) {
	cases := []struct {
		name       string
		vpns       []string
		brokerOnly bool
		wantErr    bool
	}{
		{"neither flag", nil, false, false},
		{"--vpn only", []string{"acme"}, false, false},
		{"--broker-only only", nil, true, false},
		{"both set: refused", []string{"acme"}, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vpns, brokerOnly, err := exportScope(&App{vpns: tc.vpns, brokerOnly: tc.brokerOnly})
			if tc.wantErr {
				if err == nil || !isUsage(err) {
					t.Fatalf("exportScope(%v, %v) err = %v, want a usage error", tc.vpns, tc.brokerOnly, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("exportScope(%v, %v) err = %v, want nil", tc.vpns, tc.brokerOnly, err)
			}
			if brokerOnly != tc.brokerOnly || len(vpns) != len(tc.vpns) {
				t.Errorf("exportScope(%v, %v) = (%v, %v), want the flags back unchanged", tc.vpns, tc.brokerOnly, vpns, brokerOnly)
			}
		})
	}
}

// TestPluralVPNWording covers the confirmation sentence's one piece of English
// that varies with count -- "1 VPNs" reads as broken in a prompt an operator is
// about to answer under pressure.
func TestPluralVPNWording(t *testing.T) {
	if got := pluralVPN(1); got != "it" {
		t.Errorf("pluralVPN(1) = %q, want %q", got, "it")
	}
	if got := pluralVPN(2); got != "them" {
		t.Errorf("pluralVPN(2) = %q, want %q", got, "them")
	}
}

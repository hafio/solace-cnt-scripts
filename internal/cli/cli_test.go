package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"solace/internal/config"
	"solace/internal/engine"
)

// TestMain severs the package from the real stdin for every test in it.
//
// interactive() is deliberately generous now -- it asks only whether stdin is usable,
// never what kind of stream it is (see stdinCanAnswer) -- so a test that reaches a
// confirm helper without setting App.PromptIn would read the console the suite was
// launched from and block there waiting for a human. Pointing stdin at an
// already-closed pipe makes every such read return EOF at once, which is exactly the
// unattended case every confirm helper is required to refuse. Tests that want a real
// answer swap os.Stdin themselves, one test at a time.
func TestMain(m *testing.M) {
	r, w, err := os.Pipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: os.Pipe: %v\n", err)
		os.Exit(1)
	}
	w.Close()
	os.Stdin = r
	os.Exit(m.Run())
}

// sampleEnv is the shared fixture: the user template doubles as a valid config
// for k8s, docker, and podman (same file render_test.go renders from). During
// `go test ./internal/cli` the cwd is the package dir, so this relative path
// resolves to the repo's env/sample.yaml.
const sampleEnv = "../../env/sample.yaml"

// withEnv appends the sample --env so a command reaching PreRunE loads
// a valid config instead of the missing env.yaml. The value carries a separator,
// so it resolves verbatim rather than through the base-dir/env lookup.
func withEnv(args ...string) []string {
	return append(append([]string{}, args...), "--env", sampleEnv)
}

// echoRunner installs engine.Echo as the App's runner. It replaces the retired
// --dry-run flag: the flag is gone from the CLI, but the property these tests
// assert -- which argv a command would issue -- is unchanged, so the seam that
// used to be a user-facing mode is now a test-only one.
func echoRunner(a *App) { a.NewRunner = func(*App) engine.Runner { return engine.Echo{W: os.Stdout} } }

// smokeAdminPass is a distinctive admin password used only by the standalone test
// env, so TestSecretsNeverEchoed can prove it never reaches stdout.
const smokeAdminPass = "SMOKE-PW-do-not-log-1234"

// writeStandaloneEnv writes a minimal single-broker (redundancy.enabled: false) env to a temp file
// and returns its path. The sample defines all three nodes, so the HA-only config and
// verify steps poll or exercise failover there -- unsuitable for a clean run over the echo
// seam. This env has no nodes, so those steps self-skip.
func writeStandaloneEnv(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "standalone.yaml")
	content := "redundancy:\n  enabled: false\n" +
		"image:\n" +
		"  repo: solace-pubsub-standard\n" +
		"  tag: \"10.10.1.128\"\n" +
		"semp:\n" +
		"  adminPass: " + smokeAdminPass + "\n" +
		"kubernetes:\n" +
		"  name: dev-broker\n" +
		"  namespace: solace\n" +
		"  adminSecret: solace-admin-secret\n" +
		"  updateStrategy: automatedRolling\n" +
		"  storage:\n" +
		"    class: standard\n" +
		"    msgNodeSize: 30Gi\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write standalone env: %v", err)
	}
	return path
}

// runStandalone runs a k8s command over the echo seam against the standalone env.
func runStandalone(t *testing.T, path string, args ...string) (string, error) {
	t.Helper()
	full := append(append([]string{}, args...), "--env", path)
	return runRootWith(t, full, echoRunner)
}

// runCtr runs a container (docker/podman) command over the echo seam against the
// given env path. It is the container-facing name for the same run that
// runStandalone performs, kept distinct for readability at container call sites
// (against both the HA sample and a container-standalone env).
func runCtr(t *testing.T, path string, args ...string) (string, error) {
	t.Helper()
	return runStandalone(t, path, args...)
}

// writeCtrStandaloneEnv writes a minimal single-broker (redundancy.enabled: false) env that is
// valid for docker and podman (data dir and network mode default). It carries no TLS,
// domainCerts, or productKeys, so the config steps that gate on those self-skip and none
// of them reach a poll loop -- keeping every run over the echo seam fast and
// deterministic. The k8s-shaped writeStandaloneEnv cannot be reused: it declares no
// docker/podman section, so resolvePlatform refuses it.
//
// It names the primary even though standalone no longer requires one: the routername ends
// up in every rendered artifact, and letting it default here would make each assertion
// depend on the name of the machine running the suite. The fallback has its own test
// (TestStandaloneRouternameFallsBackToTheHost), which injects a fixed hostname.
//
// An env file must declare its platform section -- even an empty one -- or
// resolvePlatform refuses it ("declares no platform section"). This fixture
// declares BOTH docker: {} and podman: {} so the one file still serves either
// platform; every caller is therefore ambiguous and must pass --platform.
func writeCtrStandaloneEnv(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ctr-standalone.yaml")
	content := "redundancy:\n  enabled: false\n" +
		"  primary:\n" +
		"    name: pri-host\n" +
		"image:\n" +
		"  repo: solace-pubsub-standard\n" +
		"  tag: \"10.10.1.128\"\n" +
		"semp:\n" +
		"  adminPass: " + smokeAdminPass + "\n" +
		"docker: {}\n" +
		"podman:\n" +
		"  baseDir: /opt/solace\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write container standalone env: %v", err)
	}
	return path
}

// loadDirect writes yamlBody to a temp file and loads it for platform p, for
// tests that need a *config.Config to build an App directly (bypassing App.load
// and its runner-selection entirely) so a fake Runner other than engine.Echo can
// be attached -- see opRunner below.
func loadDirect(t *testing.T, yamlBody string, p config.Platform) *config.Config {
	t.Helper()
	path := filepath.Join(t.TempDir(), "direct.yaml")
	if err := os.WriteFile(path, []byte(yamlBody), 0o644); err != nil {
		t.Fatalf("write direct env: %v", err)
	}
	cfg, err := config.Load(path, p)
	if err != nil {
		t.Fatalf("load direct env: %v", err)
	}
	return cfg
}

// --- fake engine.Runner for direct op-function calls ------------------------
//
// opK8sDeployAll/opK8sPrepAll/opK8sRemoveAll/opCtrVerifyRedundancy etc. are
// unexported handlers that take only an *App (or an *App plus a role arg), so a
// test in this package can call them directly rather than through cobra/runRoot.
// Doing so is what makes two things possible that engine.Echo cannot provide: (1)
// targeted fault injection -- Echo's Run/Output methods never return an error, so
// the ordering property "step N fails, step N+1 never runs" has no double to
// exercise it without this, mirroring internal/container's capRunner/failOn
// (transport_test.go); and (2) canned Output content -- Echo always returns (nil,
// nil), which makes a poll-based HA state machine (config-sync leader,
// redundancy) exhaust its real PollInterval x PollAttempts budget (broker.New's 2s
// x 60 defaults) before timing out, since the CLI layer has no seam to shorten it.
// Supplying a healthy transcript lets the state machine succeed on its first check
// instead.
type opCall struct {
	method string // Run | RunInput | RunEnv | RunInteractive | Output | OutputInput
	name   string
	args   []string
	stdin  string // RunInput/OutputInput only
}

// opRunner is a fake engine.Runner. fail, when set, is consulted for every call
// after it is recorded: a non-nil return is propagated as that command's error. It
// receives the whole opCall (not just name/args, unlike container's failOn) so a
// test can target a call by its stdin body as well as its argv -- see
// failDisableDefaultUsersUpload, which needs exactly that to disambiguate two
// steps that issue the identical script name. output, when set, supplies
// Output/OutputInput's returned bytes whenever fail did not fire.
type opRunner struct {
	calls  []opCall
	fail   func(opCall) error
	output func(opCall) []byte
}

func (r *opRunner) Run(_ context.Context, name string, args ...string) error {
	return r.do(opCall{method: "Run", name: name, args: args})
}
func (r *opRunner) RunInput(_ context.Context, in []byte, name string, args ...string) error {
	return r.do(opCall{method: "RunInput", name: name, args: args, stdin: string(in)})
}
func (r *opRunner) RunEnv(_ context.Context, _ []string, name string, args ...string) error {
	return r.do(opCall{method: "RunEnv", name: name, args: args})
}
func (r *opRunner) RunInteractive(_ context.Context, name string, args ...string) error {
	return r.do(opCall{method: "RunInteractive", name: name, args: args})
}
func (r *opRunner) do(c opCall) error {
	r.calls = append(r.calls, c)
	if r.fail != nil {
		return r.fail(c)
	}
	return nil
}

// opCanI reports whether a recorded call is the k8s read-only permission probe
// (Cluster.Preflight's `auth can-i`), which every mutating operation issues first.
// The double answers it "yes" out of band so the op-level tests below stay about
// the work they were written for; a test that needs a refusal sets output/fail for
// this call itself, as TestOpK8sDeployStopsOnPreflightRefusal does.
func opCanI(c opCall) bool {
	for _, a := range c.args {
		if a == "can-i" {
			return true
		}
	}
	return false
}

func (r *opRunner) Output(_ context.Context, name string, args ...string) ([]byte, error) {
	c := opCall{method: "Output", name: name, args: args}
	r.calls = append(r.calls, c)
	if r.fail != nil {
		if err := r.fail(c); err != nil {
			return nil, err
		}
	}
	if r.output != nil {
		if out := r.output(c); out != nil {
			return out, nil
		}
	}
	if opCanI(c) {
		return []byte("yes\n"), nil
	}
	return nil, nil
}
func (r *opRunner) OutputInput(_ context.Context, in []byte, name string, args ...string) ([]byte, error) {
	c := opCall{method: "OutputInput", name: name, args: args, stdin: string(in)}
	r.calls = append(r.calls, c)
	if r.fail != nil {
		if err := r.fail(c); err != nil {
			return nil, err
		}
	}
	if r.output != nil {
		return r.output(c), nil
	}
	return nil, nil
}

// opArgvMatch reports whether substr appears in c's command name or any of its args.
func opArgvMatch(c opCall, substr string) bool {
	if strings.Contains(c.name, substr) {
		return true
	}
	for _, a := range c.args {
		if strings.Contains(a, substr) {
			return true
		}
	}
	return false
}

// hasCall reports whether any recorded call's name or args contain substr.
func (r *opRunner) hasCall(substr string) bool {
	for _, c := range r.calls {
		if opArgvMatch(c, substr) {
			return true
		}
	}
	return false
}

// dump renders the recorded call sequence for a failure message: which command
// ran, in what order, is the only thing that distinguishes "the abort worked" from
// "the injected failure never landed". Stdin is deliberately omitted -- several of
// these uploads carry passwords.
func (r *opRunner) dump() string {
	var b strings.Builder
	for i, c := range r.calls {
		fmt.Fprintf(&b, "  %2d %-11s %s %s\n", i+1, c.method, c.name, strings.Join(c.args, " "))
	}
	return b.String()
}

// callCount counts recorded calls whose name or args contain substr.
func (r *opRunner) callCount(substr string) int {
	n := 0
	for _, c := range r.calls {
		if opArgvMatch(c, substr) {
			n++
		}
	}
	return n
}

// opFailOn builds an opRunner.fail hook that errors on every call whose command
// name or args contain substr -- for a marker that is unique to one step, where
// failing its first (chronological) match aborts before any later call is made.
func opFailOn(substr string) func(opCall) error {
	return func(c opCall) error {
		if opArgvMatch(c, substr) {
			return fmt.Errorf("injected failure for %q", substr)
		}
		return nil
	}
}

// opFailOnCount builds a fail hook that errors only on the nth (1-indexed) call
// whose name or args contain substr, succeeding every other one -- for
// opK8sDeployAll/opK8sRemoveAll's repeated, textually-identical steps (every
// `apply -f -` / `delete ... --ignore-not-found` call looks the same; only its
// position in the sequence identifies which step it belongs to).
func opFailOnCount(substr string, n int) func(opCall) error {
	count := 0
	return func(c opCall) error {
		if !opArgvMatch(c, substr) {
			return nil
		}
		count++
		if count == n {
			return fmt.Errorf("injected failure for %q (call #%d)", substr, n)
		}
		return nil
	}
}

// failDisableDefaultUsersUpload targets DisableDefaultUsers' own "show-vpn" probe
// without also aborting DisableDefaultVPN's: both steps run RunCLI with the
// identical script name "show-vpn" (broker/config_ops.go), so their argv is
// indistinguishable. What differs is the uploaded body: DisableDefaultVPN's own
// closing probe uploads the interactive-mode script (showVPNScript, which opens
// "home\nenable\nconfigure"), while DisableDefaultUsers uploads the bare one
// (showVPNBareScript, just "show message-vpn *"). Matching on the upload's stdin --
// invisible to argv-only matching -- is what makes the two distinguishable at all.
func failDisableDefaultUsersUpload(c opCall) error {
	if c.method == "RunInput" && opArgvMatch(c, ".show-vpn.cli") && !strings.Contains(c.stdin, "configure") {
		return fmt.Errorf("injected failure for disable-default-users' show-vpn upload")
	}
	return nil
}

// healthyShowRD is a canned `show redundancy` transcript satisfying
// primaryRedundancyUp (internal/broker/verify_ops.go), so a direct-call test can
// drive Leader's poll to succeed on the first check.
const healthyShowRD = "Configuration Status: Enabled\n" +
	"Redundancy Status: Up\n" +
	"Active-Standby Role: Primary\n" +
	"ADB Link To Mate: Up\n" +
	"ADB Hello To Mate: Up\n"

// capture redirects the given standard stream (os.Stdout or os.Stderr) through a
// pipe for the duration of fn and returns everything written.
func capture(t *testing.T, target **os.File, fn func()) string {
	t.Helper()
	old := *target
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	*target = w
	type res struct {
		b []byte
		e error
	}
	ch := make(chan res, 1)
	go func() {
		b, e := io.ReadAll(r)
		ch <- res{b, e}
	}()
	fn()
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe: %v", err)
	}
	*target = old
	out := <-ch
	if out.e != nil {
		t.Fatalf("read pipe: %v", out.e)
	}
	return string(out.b)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	return capture(t, &os.Stdout, fn)
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	return capture(t, &os.Stderr, fn)
}

// runRoot builds a fresh command tree with a fresh App, runs it with args, and
// returns captured stdout plus the Execute error. A fresh App per call means no
// flag scratch state leaks between cases.
func runRoot(t *testing.T, args []string) (string, error) {
	t.Helper()
	return runRootWith(t, args, nil)
}

// runRootWith is runRoot, but lets a test configure the App before Execute --
// e.g. setting Interactive/PromptIn to deterministically drive the confirm-prompt
// branches (confirmDelete/confirmLayer/confirmRestart) that a real terminal would
// otherwise gate on the test process's own, environment-dependent stdin, or
// installing echoRunner to capture the argv a command would issue instead of
// running it for real. configure may be nil, in which case this is exactly
// runRoot.
func runRootWith(t *testing.T, args []string, configure func(*App)) (string, error) {
	t.Helper()
	var runErr error
	out := captureStdout(t, func() {
		app := &App{}
		if configure != nil {
			configure(app)
		}
		root := newRootCmd(app)
		root.SetArgs(args)
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		runErr = root.Execute()
	})
	return out, runErr
}

// findCmd walks the tree from root by successive command names.
func findCmd(t *testing.T, root *cobra.Command, path ...string) *cobra.Command {
	t.Helper()
	cur := root
	for _, name := range path {
		var next *cobra.Command
		for _, c := range cur.Commands() {
			if c.Name() == name {
				next = c
				break
			}
		}
		if next == nil {
			t.Fatalf("command %q not found under %q", name, cur.CommandPath())
		}
		cur = next
	}
	return cur
}

func collectPaths(c *cobra.Command, acc *[]string) {
	*acc = append(*acc, c.CommandPath())
	for _, sub := range c.Commands() {
		collectPaths(sub, acc)
	}
}

// runStatusStderr runs `status broker` over the echo seam with the given flags and
// returns what reached stderr, where the resolved-env-file line is echoed.
func runStatusStderr(t *testing.T, args ...string) (string, error) {
	t.Helper()
	var runErr error
	errOut := captureStderr(t, func() {
		_, runErr = runRootWith(t, append([]string{"broker", "status", "--platform", "kubernetes"}, args...), echoRunner)
	})
	return errOut, runErr
}

// TestEnvFileLookup covers the resolver as the CLI wires it: -e names a file,
// searched in the base dir then <base-dir>/env, and the winner is echoed so a
// base-dir copy shadowing the env/ copy is visible.
func TestEnvFileLookup(t *testing.T) {
	body, err := os.ReadFile(writeStandaloneEnv(t))
	if err != nil {
		t.Fatalf("read standalone env: %v", err)
	}
	root := t.TempDir()
	put := func(rel string) string {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(p, body, 0o600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		return p
	}

	t.Run("falls back to the env dir", func(t *testing.T) {
		want := put("env/dev.yaml")
		out, err := runStatusStderr(t, "--base-dir", root, "-e", "dev.yaml")
		if err != nil {
			t.Fatalf("status err = %v", err)
		}
		if !strings.Contains(out, "==> env file: "+want) {
			t.Errorf("stderr = %q, want the env file line for %q", out, want)
		}
	})

	t.Run("base dir shadows the env dir", func(t *testing.T) {
		want := put("dev.yaml")
		out, err := runStatusStderr(t, "--base-dir", root, "-e", "dev.yaml")
		if err != nil {
			t.Fatalf("status err = %v", err)
		}
		if !strings.Contains(out, "==> env file: "+want) {
			t.Errorf("stderr = %q, want the env file line for %q", out, want)
		}
	})

	t.Run("no extension is inferred", func(t *testing.T) {
		_, err := runStatusStderr(t, "--base-dir", root, "-e", "dev")
		if err == nil || !strings.Contains(err.Error(), `env file "dev" not found`) {
			t.Fatalf("-e dev err = %v, want a not-found error", err)
		}
	})

	t.Run("long and short flags agree", func(t *testing.T) {
		want := filepath.Join(root, "dev.yaml")
		out, err := runStatusStderr(t, "--base-dir", root, "--env", "dev.yaml")
		if err != nil {
			t.Fatalf("status err = %v", err)
		}
		if !strings.Contains(out, "==> env file: "+want) {
			t.Errorf("stderr = %q, want the env file line for %q", out, want)
		}
	})
}

// TestFirstArg / TestFirstArgOr are gone with firstArg/firstArgOr: no command reads a
// positional argument that needs defaulting any more. The platform is --platform, the role
// is --pod, and the file-path positionals that remain are required by their Args validator
// rather than defaulted.

// TestNotImplemented is gone with the notImplemented placeholder, and so is every
// deliberately-unimplemented DIRECTION: `configure server-certs --remove` and
// `configure product-keys --remove` were both loud refusals until `no ssl
// server-certificate` and `no product-key <key>` were confirmed on a live broker. Every
// command in the tree, in both directions, is wired to a real handler. The refusal shape
// they shared (notImplementedRemoval) went with them rather than being left as dead code
// for a future placeholder to maybe reuse.

func TestEmit(t *testing.T) {
	var emitErr error
	out := captureStdout(t, func() { emitErr = emit([]byte("hello world")) })
	if emitErr != nil {
		t.Fatalf("emit err = %v, want nil", emitErr)
	}
	if out != "hello world" {
		t.Errorf("emit wrote %q, want %q", out, "hello world")
	}
}

func TestWarnAndStep(t *testing.T) {
	out := captureStderr(t, func() { warn("bad %s", "thing") })
	if !strings.Contains(out, "[WARN] bad thing") {
		t.Errorf("warn wrote %q, want it to contain '[WARN] bad thing'", out)
	}
	out = captureStderr(t, func() { step("doing %d", 5) })
	if !strings.Contains(out, "==> doing 5") {
		t.Errorf("step wrote %q, want it to contain '==> doing 5'", out)
	}
}

// TestTreeStructure covers the one flat tree: the platform is resolved rather than
// typed as the first word of a command (platform.go), so every command lives
// directly off root (or off a verb group) regardless of which platform it applies
// to. The sample below is representative, not exhaustive: a top-level leaf, a
// group's child, a couple of multi-level paths, and at least one command from
// each applicability class (shared, kubernetes-only, container-only).
func TestTreeStructure(t *testing.T) {
	root := newRootCmd(&App{})

	var paths []string
	collectPaths(root, &paths)
	have := make(map[string]bool, len(paths))
	for _, p := range paths {
		have[p] = true
	}
	wantLeaves := []string{
		"solace-util validate",                      // top level, shared
		"solace-util broker deploy",                 // noun's child, shared
		"solace-util broker validate",               // the scoped half
		"solace-util operator remove",               // noun's child, kubernetes-only
		"solace-util operator generate",             // noun's child, kubernetes-only
		"solace-util broker configure server-certs", // three-level path, shared
		"solace-util broker perform assert-leader",  // three-level path, shared
		"solace-util broker copy from",
		"solace-util broker copy into",
		"solace-util broker status",
		"solace-util convert",
		"solace-util examples",
		"solace-util auto-complete",
		"solace-util auto-complete powershell",
		"solace-util version",
	}
	for _, want := range wantLeaves {
		if !have[want] {
			t.Errorf("command path %q missing from tree", want)
		}
	}
}

// TestEveryRunnableCommandIsWired pins the wiring that replaced the two
// PersistentPreRunE hooks: a command that runs something must carry the shared
// pre-run (which resolves the platform and loads the env file) and the
// --allow-command flag that goes with executing. Missing either is invisible
// until that one command is run, so it is checked structurally instead. The verb
// groups (check, deploy, config, remove, ...) carry no RunE at all -- they print
// help and act on nothing -- so this walk never touches them.
func TestEveryRunnableCommandIsWired(t *testing.T) {
	root := newRootCmd(&App{})

	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		path := c.CommandPath()
		// A verb group carries a RunE only to print help or reject an unknown noun
		// (group(), commands.go) -- it reaches no external command, so it must NOT
		// carry the pre-run or --allow-command.
		excluded := c.Annotations[groupAnnotation] == "true" ||
			path == "solace-util convert" ||
			path == "solace-util examples" ||
			path == "solace-util version" ||
			path == "solace-util auto-complete" ||
			strings.HasPrefix(path, "solace-util auto-complete ")
		if c.RunE != nil && !excluded {
			if c.PreRunE == nil {
				t.Errorf("%s: has RunE but no PreRunE (missing wireExec)", path)
			}
			if c.Flags().Lookup("allow-command") == nil {
				t.Errorf("%s: has RunE but no --allow-command flag (missing wireExec)", path)
			}
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(root)

	if root.PersistentFlags().Lookup("allow-command") != nil {
		t.Error("root carries --allow-command as a persistent flag; it must be per-command (wireExec), never inherited")
	}
	convert := findCmd(t, root, "convert")
	if convert.Flags().Lookup("allow-command") != nil {
		t.Error("convert has an --allow-command flag; it loads no env file and executes nothing")
	}
	eg := findCmd(t, root, "examples")
	if eg.Flags().Lookup("allow-command") != nil {
		t.Error("examples has an --allow-command flag; it emits a template and executes nothing")
	}
	if eg.PreRunE != nil {
		t.Error("examples has a PreRunE; it must not load an env file -- it writes one")
	}
}

// TestGroupCommandsPrintHelpAndDoNothing pins the no-implicit-actions rule
// (commands.go): a verb that owns objects has no RunE, so running it bare prints
// its own help and touches nothing. Proof that it touches nothing is that it
// succeeds with no --env at all -- a runnable leaf would fail resolving the
// missing default env.yaml, but a bare group never reaches PreRunE.
func TestGroupCommandsPrintHelpAndDoNothing(t *testing.T) {
	for _, name := range []string{"broker", "operator"} {
		t.Run(name, func(t *testing.T) {
			if _, err := runRoot(t, []string{name}); err != nil {
				t.Errorf("%s (bare) err = %v, want nil (prints help, does not act)", name, err)
			}
		})
	}
	// The families under broker behave the same way.
	for _, path := range [][]string{{"broker", "copy"}, {"broker", "configure"}, {"broker", "perform"}} {
		t.Run(strings.Join(path, " "), func(t *testing.T) {
			if _, err := runRoot(t, path); err != nil {
				t.Errorf("%v (bare) err = %v, want nil (prints help, does not act)", path, err)
			}
		})
	}
}

// TestFlagsRegistered covers flag registration on the flat tree. Each command is
// one static shape on every platform (platform.go), so a platform-scoped flag
// like --restart is registered on `deploy broker` unconditionally and only its
// applicability (flagOnlyOn) narrows by platform.
func TestFlagsRegistered(t *testing.T) {
	root := newRootCmd(&App{})
	cases := []struct {
		path  []string
		flags []string
	}{
		{[]string{"broker", "deploy"}, []string{"restart"}},
		{[]string{"broker", "remove"}, []string{"delete-data", "no-prompt"}},
		{[]string{"operator", "remove"}, []string{"delete-crd", "no-prompt"}},
		{[]string{"broker", "perform", "gather-diagnostics"}, []string{"days"}},
		{[]string{"broker", "cli"}, []string{"pod"}},
		{[]string{"broker", "perform", "cli-script"}, []string{"pod"}},
		{[]string{"broker", "perform", "shell-script"}, []string{"pod"}},
		{[]string{"broker", "copy", "from"}, []string{"pod"}},
		{[]string{"broker", "copy", "into"}, []string{"pod", "dir"}},
		{[]string{"broker", "status"}, []string{"all", "detail", "pod"}},
		{[]string{"operator", "status"}, []string{"detail"}},
		// Every `broker configure` leaf whose disable/remove direction acts on a live
		// broker asks first, so all five carry the silencer.
		{[]string{"broker", "configure", "server-certs"}, []string{"apply", "remove", "no-prompt"}},
		{[]string{"broker", "configure", "product-keys"}, []string{"apply", "remove", "no-prompt"}},
		{[]string{"broker", "configure", "domain-certs"}, []string{"apply", "remove", "no-prompt"}},
		{[]string{"broker", "configure", "default-vpn"}, []string{"disable", "enable", "no-prompt"}},
		{[]string{"broker", "configure", "default-users"}, []string{"disable", "enable", "no-prompt"}},
		{[]string{"examples"}, []string{"out", "no-prompt"}},
		// Every command whose output is an artifact you keep takes the same pair.
		{[]string{"broker", "generate"}, []string{"out", "no-prompt"}},
		{[]string{"operator", "generate"}, []string{"out", "no-prompt"}},
		{[]string{"convert"}, []string{"out", "no-prompt"}},
	}
	for _, tc := range cases {
		t.Run(strings.Join(tc.path, "/"), func(t *testing.T) {
			cmd := findCmd(t, root, tc.path...)
			for _, name := range tc.flags {
				if cmd.Flags().Lookup(name) == nil {
					t.Errorf("%s: flag %q not registered", cmd.CommandPath(), name)
				}
			}
		})
	}
}

func TestHelpNoConfig(t *testing.T) {
	// --help short-circuits before PreRunE, so no env is needed.
	cases := [][]string{
		{"--help"},
		{"broker", "--help"},
		{"broker", "configure", "--help"},
		{"broker", "status", "--help"},
		{"operator", "--help"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			if _, err := runRoot(t, args); err != nil {
				t.Errorf("help %q err = %v, want nil", args, err)
			}
		})
	}
}

// TestGenerateWired covers the one render per noun that replaced the generate sub-tree.
//
// Each platform has one artifact and one command that prints it.
//
// The prefixes are what each renderer actually emits first. On k8s that is now a Secret,
// not the CR: secrets are emitted ahead of the resource that references them, so the stream
// applies in order to an empty namespace.
func TestGenerateWired(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		prefix string
	}{
		{"kubernetes broker generate", []string{"broker", "generate", "--platform", "kubernetes"}, "apiVersion: v1"},
		{"docker broker generate", []string{"broker", "generate", "--platform", "docker"}, "name:"},
		{"podman broker generate", []string{"broker", "generate", "--platform", "podman"}, "[Unit]"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runRoot(t, withEnv(tc.args...))
			if err != nil {
				t.Fatalf("%s err = %v, want nil", tc.name, err)
			}
			if out == "" {
				t.Fatalf("%s produced empty stdout", tc.name)
			}
			if !strings.HasPrefix(out, tc.prefix) {
				t.Errorf("%s stdout prefix = %q, want it to start with %q", tc.name, firstLine(out), tc.prefix)
			}
		})
	}
}

// TestGenerateEmitsSecretsBeforeTheCR is the ordering half of decision 6 on Kubernetes,
// end to end through the command rather than through k8s.GenBroker.
//
// `generate secrets broker` and `generate broker` used to be two commands whose outputs had
// to be applied in the right order by hand. One command prints all three documents now, so
// the order is this tool's responsibility: the Namespace before the namespaced objects, and
// a Secret before the CR that names it, or applying the stream to an empty cluster leaves
// the broker referencing something that does not exist yet.
func TestGenerateEmitsSecretsBeforeTheCR(t *testing.T) {
	out, err := runRoot(t, []string{"broker", "generate", "--env", writeStandaloneEnv(t), "--platform", "kubernetes"})
	if err != nil {
		t.Fatalf("broker generate err = %v, want nil", err)
	}
	ns := strings.Index(out, "kind: Namespace")
	secret := strings.Index(out, "kind: Secret")
	cr := strings.Index(out, "kind: PubSubPlusEventBroker")
	if ns < 0 || secret < 0 || cr < 0 {
		t.Fatalf("broker generate must emit the Namespace, a Secret and the CR, got:\n%s", out)
	}
	if !(ns < secret && secret < cr) {
		t.Errorf("broker generate emitted its documents out of apply order (ns=%d secret=%d cr=%d):\n%s",
			ns, secret, cr, out)
	}
}

// TestCtrWiredDryRun drives every container command that is safe to run against
// the HA sample env over the echo seam: each reaches its real handler and returns
// no error, with the expected "+ <runtime> ..." (or systemctl/mkdir/chown) echo
// landing on stdout. Poll-driven steps (config leader / smoke redundancy, which
// fail over or wait) and secret-bearing prep are covered by the guard, standalone,
// and error tests instead, so nothing here blocks on a poll loop. The sample is
// docker compose mode and rootful podman, so status/remove take the compose path
// and podman systemctl carries no --user.
func TestCtrWiredDryRun(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"docker check deploy", []string{"validate", "--platform", "docker"}, "+ docker version"},
		{"podman check deploy", []string{"validate", "--platform", "podman"}, "+ podman version"},
		{"docker status broker", []string{"broker", "status", "--platform", "docker"}, "+ docker ps"},
		{"podman status broker", []string{"broker", "status", "--platform", "podman"}, "+ podman ps"},
		{"docker copy from", []string{"broker", "copy", "from", "a.log", "--platform", "docker"}, "+ docker cp"},
		{"docker copy into", []string{"broker", "copy", "into", "a.cli", "--dir", "/tmp", "--platform", "docker"}, "+ docker cp"},
		// --follow is a flag on both platforms now. The container path used to hardcode -f,
		// which made the flag a no-op there and meant a plain `logs` never returned.
		{"docker logs broker", []string{"broker", "logs", "--platform", "docker"}, "+ docker logs solace"},
		{"docker logs broker --follow", []string{"broker", "logs", "--follow", "--platform", "docker"}, "+ docker logs -f"},
		{"docker cli", []string{"broker", "cli", "--platform", "docker"}, "+ docker exec -it"},
		{"docker shell", []string{"broker", "shell", "--platform", "docker"}, "+ docker exec -it"},
		// The host prerequisites are folded into deploy now, so the mkdir/chown that
		// `prepare host` used to issue must show up on the deploy path itself.
		{"docker deploy makes the data dir", []string{"broker", "deploy", "--pod", "primary", "--platform", "docker"}, "+ mkdir -p"},
		{"docker deploy owns the data dir", []string{"broker", "deploy", "--pod", "primary", "--platform", "docker"}, "+ chown"},
		{"docker deploy broker primary", []string{"broker", "deploy", "--pod", "primary", "--platform", "docker"}, "+ docker compose"},
		{"podman deploy broker primary", []string{"broker", "deploy", "--pod", "primary", "--platform", "podman"}, "+ systemctl daemon-reload"},
		{"docker remove broker", []string{"broker", "remove", "--no-prompt", "--platform", "docker"}, "+ docker compose"},
		{"podman remove broker", []string{"broker", "remove", "--no-prompt", "--platform", "podman"}, "+ systemctl"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runCtr(t, sampleEnv, tc.args...)
			if err != nil {
				t.Fatalf("%s err = %v, want nil", tc.name, err)
			}
			if !strings.Contains(out, tc.want) {
				t.Errorf("%s stdout = %q, want it to contain %q", tc.name, out, tc.want)
			}
		})
	}
}

// TestCtrRoleGuards covers the fail-loud / self-skip role guards on the two
// primary-driven HA operations (config leader, smoke redundancy). None reach a
// poll loop or the mate SEMP channel: the HA cases are rejected before anything
// runs (wrong node, unknown host, or bad role) and the standalone cases return
// nil via skipIfStandalone, so every case resolves immediately over the echo seam.
func TestCtrRoleGuards(t *testing.T) {
	ha := sampleEnv
	standalone := writeCtrStandaloneEnv(t)
	cases := []struct {
		name    string
		env     string
		args    []string
		wantErr string // "" -> expect nil (self-skip path)
	}{
		{"assert-leader on monitor", ha, []string{"broker", "perform", "assert-leader", "--pod", "monitor", "--platform", "docker"}, "must run on the primary node"},
		{"assert-leader on backup", ha, []string{"broker", "perform", "assert-leader", "--pod", "backup", "--platform", "podman"}, "this host is the backup node"},
		{"redundancy-test on monitor", ha, []string{"broker", "perform", "redundancy-test", "--pod", "monitor", "--platform", "docker"}, "this host is the monitor node"},
		{"redundancy-test on backup", ha, []string{"broker", "perform", "redundancy-test", "--pod", "backup", "--platform", "podman"}, "this host is the backup node"},
		{"redundancy-test unknown host", ha, []string{"broker", "perform", "redundancy-test", "--platform", "docker"}, "cannot determine node role"},
		{"assert-leader bad role", ha, []string{"broker", "perform", "assert-leader", "--pod", "bogus", "--platform", "docker"}, "invalid node role"},
		{"assert-leader standalone skip", standalone, []string{"broker", "perform", "assert-leader", "--platform", "docker"}, ""},
		{"redundancy-test standalone skip", standalone, []string{"broker", "perform", "redundancy-test", "--platform", "podman"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := runCtr(t, tc.env, tc.args...)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("%s err = %v, want nil (self-skip path)", tc.name, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("%s err = nil, want an error containing %q", tc.name, tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("%s err = %q, want it to contain %q", tc.name, err.Error(), tc.wantErr)
			}
		})
	}
}

// TestCtrConfigDryRun covers the post-deploy config steps that run cleanly on a
// container-standalone env over the echo seam: the VPN/user hardening and cli
// --input echo their exec commands, while the cert/key-gated steps self-skip
// (none configured). config leader is excluded -- it is HA-only and covered by
// TestCtrRoleGuards. None of these steps polls.
func TestCtrConfigDryRun(t *testing.T) {
	path := writeCtrStandaloneEnv(t)
	cases := []struct {
		name string
		args []string
	}{
		// The disable/remove directions now confirm like every other disruptive
		// `broker configure` leaf, so --no-prompt is what keeps these dry runs
		// actually reaching the handler instead of a silent decline.
		{"config disable default-vpn", []string{"broker", "configure", "default-vpn", "--no-prompt", "--platform", "docker"}},
		{"config disable default-users", []string{"broker", "configure", "default-users", "--no-prompt", "--platform", "docker"}},
		{"config apply domain-certs (skip)", []string{"broker", "configure", "domain-certs", "--platform", "docker"}},
		{"perform cli-script runs a script", []string{"broker", "perform", "cli-script", "setup.cli", "--platform", "docker"}},
		{"config delete domain-certs (skip)", []string{"broker", "configure", "domain-certs", "--remove", "--no-prompt", "--platform", "docker"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := runCtr(t, path, tc.args...); err != nil {
				t.Fatalf("%s (container-standalone) err = %v, want nil", tc.name, err)
			}
		})
	}
}

// TestCtrExecCLIPathSeparator covers opCtrExecCLI's used-as-is branch: a file
// argument containing a path separator is not joined under the cliScripts folder.
// The bare-filename (join) branch is covered by TestCtrConfigDryRun.
func TestCtrExecCLIPathSeparator(t *testing.T) {
	path := writeCtrStandaloneEnv(t)
	if _, err := runCtr(t, path, "broker", "perform", "cli-script", "sub/dir/x.cli", "--platform", "docker"); err != nil {
		t.Fatalf("perform cli-script with a path arg err = %v, want nil", err)
	}
}

// TestCtrErrorPaths covers the actionable failures of the container config/check
// steps on a container-standalone env over the echo seam: the cert and
// product-key steps demand configuration that is absent, and a login over the
// echo runner cannot succeed against a non-existent broker. None polls.
func TestCtrErrorPaths(t *testing.T) {
	path := writeCtrStandaloneEnv(t)
	cases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"config apply server-cert (no tls)", []string{"broker", "configure", "server-certs", "--platform", "docker"}, "must both be set"},
		{"config apply product-keys (none)", []string{"broker", "configure", "product-keys", "--platform", "docker"}, "no product keys configured"},
		{"check semp-login (echo runner)", []string{"broker", "perform", "semp-login-check", "--platform", "docker"}, "SEMP login failed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := runCtr(t, path, tc.args...)
			if err == nil {
				t.Fatalf("%s err = nil, want an error containing %q", tc.name, tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("%s err = %q, want it to contain %q", tc.name, err.Error(), tc.wantErr)
			}
		})
	}
}

// TestCtrDiagnosticsDryRun is isolated because Diagnostics does an
// os.MkdirAll(diagDir) side-effect; t.Chdir(t.TempDir()) keeps the created dir out
// of the package directory. The container-standalone env's path is absolute, so it
// survives the chdir. Over the echo seam it echoes the node-local gather/download
// sequence without polling.
func TestCtrDiagnosticsDryRun(t *testing.T) {
	path := writeCtrStandaloneEnv(t)
	t.Chdir(t.TempDir())
	out, err := runCtr(t, path, "broker", "perform", "gather-diagnostics", "--platform", "docker")
	if err != nil {
		t.Fatalf("diagnostics err = %v, want nil", err)
	}
	if !strings.Contains(out, "+ docker") {
		t.Errorf("diagnostics stdout = %q, want a '+ docker ...' echo", out)
	}
}

// TestCtrRoleArgCount pins that the role-taking commands reject a second
// positional (cobra.MaximumNArgs(1)). Arg validation runs before PreRunE, so no
// env is loaded.
func TestCtrRoleArgCount(t *testing.T) {
	cases := [][]string{
		{"broker", "perform", "assert-leader", "primary", "extra", "--platform", "docker"},
		{"broker", "perform", "redundancy-test", "primary", "extra", "--platform", "docker"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			if _, err := runCtr(t, sampleEnv, args...); err == nil {
				t.Fatalf("%q err = nil, want a too-many-args error", args)
			}
		})
	}
}

// TestCtrRoleHelp confirms the role-taking commands expose --help (which
// short-circuits before PreRunE, so no env is needed). runRoot discards
// cobra's help output, so the assertion is purely that Execute returns no error.
func TestCtrRoleHelp(t *testing.T) {
	cases := [][]string{
		{"broker", "perform", "assert-leader", "--help", "--platform", "docker"},
		{"broker", "perform", "redundancy-test", "--help", "--platform", "podman"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			if _, err := runRoot(t, args); err != nil {
				t.Errorf("help %q err = %v, want nil", args, err)
			}
		})
	}
}

// TestK8sWiredDryRun drives every k8s command that is safe to run against the HA
// sample env over the echo seam: each reaches its real handler and returns no
// error. wantEcho commands shell out to kubectl (so a `+ kubectl ...` line lands
// on stdout); the skip-path commands (no configured labels / domain certs) return
// cleanly without touching the runner. Steps that need a live cluster to make
// sense on the HA sample (config leader -> redundancy poll, smoke redundancy ->
// failover, server-cert/secrets -> absent cert files) are exercised in the
// standalone and error tests instead.
func TestK8sWiredDryRun(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantEcho bool
	}{
		{"check deploy", []string{"validate"}, true},
		{"status broker", []string{"broker", "status"}, true},
		{"status broker --all", []string{"broker", "status", "--all"}, true},
		{"status broker --detail", []string{"broker", "status", "--detail"}, true},
		{"logs broker", []string{"broker", "logs"}, true},
		{"cli", []string{"broker", "cli"}, true},
		{"shell", []string{"broker", "shell"}, true},
		{"start broker", []string{"broker", "start"}, true},
		{"stop broker", []string{"broker", "stop"}, true},
		// restart deletes pods, so it takes the same --no-prompt gate remove does.
		{"restart broker (all)", []string{"broker", "restart", "--no-prompt"}, true},
		{"restart broker backup", []string{"broker", "restart", "--pod", "backup", "--no-prompt"}, true},
		// roleWord's remaining two cases (backup is covered above): a swapped case
		// would misname which pod the prompt is about to bounce, so both must
		// actually be reached, not assumed from backup's.
		{"restart broker monitor", []string{"broker", "restart", "--pod", "monitor", "--no-prompt"}, true},
		{"restart broker primary explicit", []string{"broker", "restart", "--pod", "primary", "--no-prompt"}, true},
		{"deploy broker", []string{"broker", "deploy"}, true},
		{"deploy operator", []string{"operator", "deploy"}, true},
		{"restart operator", []string{"operator", "restart"}, true},
		{"status operator", []string{"operator", "status"}, true},
		{"status operator --detail", []string{"operator", "status", "--detail"}, true},
		{"logs operator", []string{"operator", "logs"}, true},
		// disable/remove confirm like every other disruptive `broker configure` leaf, so
		// --no-prompt is what keeps these reaching the handler instead of declining.
		{"config disable default-vpn", []string{"broker", "configure", "default-vpn", "--no-prompt"}, true},
		{"config disable default-users", []string{"broker", "configure", "default-users", "--no-prompt"}, true},
		{"config apply domain-certs", []string{"broker", "configure", "domain-certs"}, false}, // none configured
		{"perform cli-script --pod", []string{"broker", "perform", "cli-script", "setup.cli", "--pod", "p"}, true},
		{"copy from", []string{"broker", "copy", "from", "somefile", "--pod", "p"}, true},
		{"copy into", []string{"broker", "copy", "into", "somefile", "--pod", "p"}, true},
		{"config delete domain-certs", []string{"broker", "configure", "domain-certs", "--remove", "--no-prompt"}, false}, // none configured
		// Both removes carry --no-prompt so no confirm helper reads os.Stdin. There is one
		// remove per noun now: the secrets and namespace steps are folded into `broker remove`,
		// which deletes what this env file owns and then asks about the namespace only if
		// nothing else is left in it.
		{"remove broker", []string{"broker", "remove", "--no-prompt"}, true},
		{"remove operator", []string{"operator", "remove", "--no-prompt"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{}, tc.args...)
			args = append(args, "--platform", "kubernetes")
			out, err := runRootWith(t, withEnv(args...), echoRunner)
			if err != nil {
				t.Fatalf("%s err = %v, want nil", tc.name, err)
			}
			if tc.wantEcho && !strings.Contains(out, "+ kubectl") {
				t.Errorf("%s stdout = %q, want a '+ kubectl ...' echo", tc.name, out)
			}
			if !tc.wantEcho && strings.Contains(out, "+ kubectl") {
				t.Errorf("%s stdout = %q, want no kubectl echo (skip path)", tc.name, out)
			}
		})
	}
}

// TestK8sStandaloneDryRun covers the commands whose behavior branches on redundancy:
// on a standalone env the HA-only steps self-skip (assert-leader / redundancy-test)
// and the secret-bearing prepare steps have no TLS to guard, so config/prepare/deploy
// all run clean over the echo seam.
func TestK8sStandaloneDryRun(t *testing.T) {
	path := writeStandaloneEnv(t)
	cases := []struct {
		name     string
		args     []string
		wantEcho bool
	}{
		{"assert-leader (skipped)", []string{"broker", "perform", "assert-leader", "--platform", "kubernetes"}, false},
		{"redundancy-test (skipped)", []string{"broker", "perform", "redundancy-test", "--platform", "kubernetes"}, false},
		{"deploy", []string{"broker", "deploy", "--platform", "kubernetes"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runStandalone(t, path, tc.args...)
			if err != nil {
				t.Fatalf("%s (standalone) err = %v, want nil", tc.name, err)
			}
			if tc.wantEcho && !strings.Contains(out, "+ kubectl") {
				t.Errorf("%s (standalone) stdout = %q, want a '+ kubectl ...' echo", tc.name, out)
			}
			if !tc.wantEcho && strings.Contains(out, "+ kubectl") {
				t.Errorf("%s (standalone) stdout = %q, want no kubectl echo (skip path)", tc.name, out)
			}
		})
	}
}

// TestDeployAllDoesNotApplyOperator pins the deliberate behavior change from the
// old `up`: the operator is cluster-scoped and shared between brokers, so
// `broker deploy` must not install it -- only `operator deploy` does. The operator is
// cluster-scoped, so a broker deploy reaching for it would have one env file silently
// re-applying a controller that other namespaces depend on.
func TestDeployBrokerDoesNotApplyOperator(t *testing.T) {
	path := writeStandaloneEnv(t)
	out, err := runStandalone(t, path, "broker", "deploy", "--platform", "kubernetes")
	if err != nil {
		t.Fatalf("broker deploy err = %v, want nil", err)
	}
	// The marker is the CRD permission probe OperatorApply issues before anything
	// else. The operator's DEPLOYMENT NAME is not usable here: `check deploy` reads
	// it back to report whether the operator is installed, so it appears in the
	// echo of a run that installed nothing.
	if strings.Contains(out, "customresourcedefinitions") {
		t.Errorf("deploy all applied the operator bundle:\n%s", out)
	}
}

// TestCheckDeployWarnsWhenOperatorAbsent covers the operator probe inside the
// check report: `check deploy` is read-only, so a missing operator is a [WARN]
// row rather than a failure of the check itself -- a cluster where the operator
// is not installed YET is a reasonable thing to be checking, and installing it is
// the next command. The row stands in for the operator install `deploy all` no
// longer performs.
//
// It cannot go through the echo seam: Echo's Output never returns an error, so the
// probe would always read as "installed". This drives opK8sValidate directly with the
// fault-injecting opRunner instead, failing exactly the CRD lookup and canning the
// StorageClass answers the check needs to get that far.
func TestCheckDeployWarnsWhenOperatorAbsent(t *testing.T) {
	cfg := loadDirect(t, "redundancy:\n  enabled: false\n"+
		"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n"+
		"semp:\n  adminPass: "+smokeAdminPass+"\n"+
		"kubernetes:\n  name: dev-broker\n  namespace: solace\n"+
		"  storage:\n    class: standard\n    msgNodeSize: 30Gi\n", config.K8s)

	isCRDLookup := func(c opCall) bool {
		for i, a := range c.args {
			if a == "crd" && i+1 < len(c.args) {
				return true
			}
		}
		return false
	}
	rr := &opRunner{
		fail: func(c opCall) error {
			if isCRDLookup(c) {
				return fmt.Errorf("the server doesn't have a resource type \"crd\"")
			}
			return nil
		},
		output: func(c opCall) []byte {
			// The StorageClass probe reads one custom column at a time; answer both
			// with the values CheckStorageClass demands so it passes and the run
			// reaches the operator probe under test.
			for _, a := range c.args {
				if strings.Contains(a, "volumeBindingMode") {
					return []byte("WaitForFirstConsumer\n")
				}
				if strings.Contains(a, "allowVolumeExpansion") {
					return []byte("true\n")
				}
			}
			return nil
		},
	}
	a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr}

	// The report goes to stdout (it is report content); only the narration is on
	// stderr, so the row is captured here rather than in errOut.
	var err error
	out := captureStdout(t, func() { err = opK8sValidate(a) })
	if err != nil {
		t.Fatalf("check deploy err = %v, want nil: a missing operator warns, it does not fail the check", err)
	}
	if !strings.Contains(out, "[WARN] installed") {
		t.Errorf("check deploy report = %q, want a WARN row for the missing operator", out)
	}
	if !strings.Contains(out, "deploy operator") {
		t.Errorf("the row should name the command that installs it:\n%s", out)
	}
}

// TestStartStopRestartBroker covers the day-2 start/stop/restart verbs on both
// platform families: Kubernetes scales the statefulset(s), containers start/stop
// the container in place.
func TestStartStopRestartBroker(t *testing.T) {
	t.Run("kubernetes", func(t *testing.T) {
		for _, args := range [][]string{{"broker", "start"}, {"broker", "stop"}} {
			out, err := runRootWith(t, append(withEnv(args...), "--platform", "kubernetes"), echoRunner)
			if err != nil {
				t.Fatalf("%v err = %v, want nil", args, err)
			}
			if !strings.Contains(out, "+ kubectl") {
				t.Errorf("%v stdout = %q, want a kubectl echo", args, out)
			}
		}
		out, err := runRootWith(t, append(withEnv("broker", "restart", "--no-prompt"), "--platform", "kubernetes"), echoRunner)
		if err != nil {
			t.Fatalf("restart broker err = %v, want nil", err)
		}
		if !strings.Contains(out, "+ kubectl") {
			t.Errorf("restart broker stdout = %q, want a kubectl echo", out)
		}
	})
	t.Run("docker", func(t *testing.T) {
		path := writeCtrStandaloneEnv(t)
		for _, args := range [][]string{{"broker", "start"}, {"broker", "stop"}, {"broker", "restart", "--no-prompt"}} {
			full := append(append([]string{}, args...), "--env", path, "--platform", "docker")
			out, err := runRootWith(t, full, echoRunner)
			if err != nil {
				t.Fatalf("%v err = %v, want nil", args, err)
			}
			if !strings.Contains(out, "+ docker") {
				t.Errorf("%v stdout = %q, want a docker echo", args, out)
			}
		}
	})
}

// TestCLICommand covers `cli`'s two shapes: bare, it opens an interactive session;
// with --input, it uploads and runs a script instead. Both are the same command
// now, distinguished by a flag rather than by being separate subcommands.
func TestCLICommand(t *testing.T) {
	t.Run("bare cli opens a session", func(t *testing.T) {
		out, err := runRootWith(t, append(withEnv("broker", "cli"), "--platform", "kubernetes"), echoRunner)
		if err != nil {
			t.Fatalf("cli err = %v, want nil", err)
		}
		if !strings.Contains(out, "+ kubectl") {
			t.Errorf("cli stdout = %q, want a kubectl exec echo", out)
		}
	})
	t.Run("--input runs a script", func(t *testing.T) {
		out, err := runRootWith(t, append(withEnv("broker", "perform", "cli-script", "setup.cli"), "--platform", "kubernetes"), echoRunner)
		if err != nil {
			t.Fatalf("perform cli-script err = %v, want nil", err)
		}
		if !strings.Contains(out, "+ kubectl") {
			t.Errorf("perform cli-script stdout = %q, want a kubectl exec echo", out)
		}
	})
	// H2 regression, now split across two commands. `cli` used to read [role] and --input
	// through two disconnected branches, so "cli backup --input maint.cli" silently ran the
	// script on the PRIMARY pod: the positional was ignored once --input was set, and --pod
	// was unset. Running a script is its own command now, which removes the branch that
	// caused it -- but the property still has to hold, because both commands resolve their
	// target through the same podRole. env/sample.yaml names the broker "dev-broker", so its
	// pods are dev-broker-pubsubplus-<p|b|m>-0.
	t.Run("--input --pod targets the named pod, not the primary", func(t *testing.T) {
		out, err := runRootWith(t, append(withEnv("broker", "perform", "cli-script", "setup.cli", "--pod", "backup"),
			"--platform", "kubernetes"), echoRunner)
		if err != nil {
			t.Fatalf("perform cli-script --pod backup err = %v, want nil", err)
		}
		if !strings.Contains(out, "dev-broker-pubsubplus-b-0") {
			t.Errorf("perform cli-script --pod backup stdout = %q, want the exec targeting the backup pod", out)
		}
		if strings.Contains(out, "dev-broker-pubsubplus-p-0") {
			t.Errorf("perform cli-script --pod backup stdout = %q, want no primary-pod exec", out)
		}
	})
	t.Run("bare --pod opens the session against the named pod", func(t *testing.T) {
		out, err := runRootWith(t, append(withEnv("broker", "cli", "--pod", "backup"), "--platform", "kubernetes"), echoRunner)
		if err != nil {
			t.Fatalf("cli --pod backup err = %v, want nil", err)
		}
		if !strings.Contains(out, "dev-broker-pubsubplus-b-0") {
			t.Errorf("cli --pod backup stdout = %q, want the session opened against the backup pod", out)
		}
		if strings.Contains(out, "dev-broker-pubsubplus-p-0") {
			t.Errorf("cli --pod backup stdout = %q, want no primary-pod session", out)
		}
	})
}

// TestStatusBrokerFlags covers how --all and --detail compose on `status broker`:
// they widen the report along independent axes (every broker in the cluster vs.
// this env file's one; the static description vs. the running summary) rather
// than one replacing the other.
func TestStatusBrokerFlags(t *testing.T) {
	t.Run("kubernetes --all lists every broker", func(t *testing.T) {
		out, err := runRootWith(t, append(withEnv("broker", "status", "--all"), "--platform", "kubernetes"), echoRunner)
		if err != nil {
			t.Fatalf("status broker --all err = %v, want nil", err)
		}
		if !strings.Contains(out, "+ kubectl") {
			t.Errorf("status broker --all stdout = %q, want a kubectl echo", out)
		}
	})
	t.Run("kubernetes --detail adds the static description", func(t *testing.T) {
		out, err := runRootWith(t, append(withEnv("broker", "status", "--detail"), "--platform", "kubernetes"), echoRunner)
		if err != nil {
			t.Fatalf("status broker --detail err = %v, want nil", err)
		}
		if !strings.Contains(out, "+ kubectl") {
			t.Errorf("status broker --detail stdout = %q, want a kubectl echo", out)
		}
	})
	t.Run("container --detail adds the inspection", func(t *testing.T) {
		path := writeCtrStandaloneEnv(t)
		out, err := runRootWith(t, []string{"broker", "status", "--detail", "--env", path, "--platform", "docker"}, echoRunner)
		if err != nil {
			t.Fatalf("status broker --detail (docker) err = %v, want nil", err)
		}
		if !strings.Contains(out, "+ docker") {
			t.Errorf("status broker --detail (docker) stdout = %q, want a docker echo", out)
		}
	})
}

// TestRemoveBrokerLayerContract covers the retained-layer contract on `remove
// broker`: persistent data is kept unless asked for by name, and --no-prompt alone
// (which only answers the delete-the-broker question) must not also answer the
// delete-the-data question.
func TestRemoveBrokerLayerContract(t *testing.T) {
	path := writeStandaloneEnv(t)
	run := func(configure func(*App), args ...string) string {
		t.Helper()
		full := append(append([]string{}, args...), "--env", path, "--platform", "kubernetes")
		return captureStderr(t, func() {
			_, err := runRootWith(t, full, configure)
			if err != nil {
				t.Fatalf("%v err = %v, want nil", args, err)
			}
		})
	}
	t.Run("--no-prompt keeps data", func(t *testing.T) {
		errOut := run(echoRunner, "broker", "remove", "--no-prompt")
		if !strings.Contains(errOut, "PVCs kept") {
			t.Errorf("remove broker --no-prompt stderr = %q, want data kept", errOut)
		}
	})
	t.Run("--delete-data deletes it", func(t *testing.T) {
		errOut := run(echoRunner, "broker", "remove", "--no-prompt", "--delete-data")
		if !strings.Contains(errOut, "PVCs deleted") {
			t.Errorf("remove broker --delete-data stderr = %q, want data deleted", errOut)
		}
	})
	t.Run("non-interactive keeps data", func(t *testing.T) {
		errOut := run(func(a *App) {
			a.Interactive = func() bool { return false }
			echoRunner(a)
		}, "broker", "remove", "--no-prompt")
		if !strings.Contains(errOut, "PVCs kept") {
			t.Errorf("remove broker --no-prompt stderr = %q, want data kept by default", errOut)
		}
	})
}

// TestRemoveOperatorLayerContract mirrors TestRemoveBrokerLayerContract for the
// operator's CRDs: kept by default (their removal cascades to every broker in the
// cluster), deleted only when named.
func TestRemoveOperatorLayerContract(t *testing.T) {
	path := writeStandaloneEnv(t)
	run := func(args ...string) string {
		t.Helper()
		full := append(append([]string{}, args...), "--env", path, "--platform", "kubernetes")
		return captureStderr(t, func() {
			_, err := runRootWith(t, full, echoRunner)
			if err != nil {
				t.Fatalf("%v err = %v, want nil", args, err)
			}
		})
	}
	t.Run("kept by default", func(t *testing.T) {
		errOut := run("operator", "remove", "--no-prompt")
		if !strings.Contains(errOut, "CRDs kept") {
			t.Errorf("remove operator stderr = %q, want CRDs kept", errOut)
		}
	})
	t.Run("--delete-crd deletes them", func(t *testing.T) {
		errOut := run("operator", "remove", "--no-prompt", "--delete-crd")
		if !strings.Contains(errOut, "CRDs deleted") {
			t.Errorf("remove operator --delete-crd stderr = %q, want CRDs deleted", errOut)
		}
	})
}

// TestRemoveFlagsCompose pins that the two flags answer DIFFERENT questions and so
// must combine rather than conflict: --delete-data says what to do with the data,
// --no-prompt says not to ask about any of it, and a fully unattended removal that
// also drops the data needs both. They were briefly mutually exclusive, which made
// exactly that case impossible to express.
func TestRemoveFlagsCompose(t *testing.T) {
	path := writeStandaloneEnv(t)
	errOut := captureStderr(t, func() {
		_, err := runRootWith(t, []string{"broker", "remove", "--delete-data", "--no-prompt",
			"--env", path, "--platform", "kubernetes"}, func(a *App) {
			a.Interactive = func() bool { return false }
			echoRunner(a)
		})
		if err != nil {
			t.Fatalf("remove broker --delete-data --no-prompt err = %v, want nil", err)
		}
	})
	if !strings.Contains(errOut, "PVCs deleted") {
		t.Errorf("stderr = %q, want the data deleted with nothing asked", errOut)
	}
}

// TestConfirmLayerMatrix pins the whole removal-confirmation contract, which INVERTED
// with the command tree.
//
// It used to be: --delete-data deletes without asking, --no-prompt keeps without asking.
// The layer flag was the destructive one and --no-prompt was always safe.
//
// It is now: the layer flag RAISES a confirmation rather than answering one. Without it
// the layer is kept and nothing is asked; with it you are asked, and --no-prompt is what
// answers yes. So a fully unattended wipe is `--delete-data --no-prompt`, and the two
// flags still compose rather than conflict -- but --no-prompt alone no longer guarantees
// the layer survives, because it now answers every question that IS raised.
//
// All four cells are here because the dangerous one is only dangerous in combination, and
// a matrix with a hole in it is how the wrong cell ships.
func TestConfirmLayerMatrix(t *testing.T) {
	yes := func() bool { return true }
	no := func() bool { return false }
	for _, tc := range []struct {
		name        string
		app         *App
		want        bool
		wantWarning string
	}{
		{"no layer flag -> kept, nothing asked",
			&App{Interactive: yes}, false, ""},
		{"no layer flag + --no-prompt -> still kept",
			&App{noPrompt: true, Interactive: yes}, false, ""},
		{"--delete-data + --no-prompt -> deleted unattended",
			&App{deleteLayer: true, noPrompt: true, Interactive: no}, true, ""},
		{"--delete-data, interactive, exact yes -> deleted",
			&App{deleteLayer: true, Interactive: yes, PromptIn: strings.NewReader("yes\n")}, true, ""},
		{"--delete-data, interactive, anything else -> kept",
			&App{deleteLayer: true, Interactive: yes, PromptIn: strings.NewReader("y\n")}, false, ""},
		{"--delete-data, no terminal, no --no-prompt -> kept and said so",
			&App{deleteLayer: true, Interactive: no}, false, "--no-prompt"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got bool
			errOut := captureStderr(t, func() { got = confirmLayer(tc.app, layerData) })
			if got != tc.want {
				t.Errorf("confirmLayer = %v, want %v", got, tc.want)
			}
			if tc.wantWarning != "" && !strings.Contains(errOut, tc.wantWarning) {
				t.Errorf("stderr = %q, want it to name %q -- refusing silently leaves the "+
					"operator thinking the data went", errOut, tc.wantWarning)
			}
		})
	}
}

// TestConfirmDeleteShortcut covers the other half: --no-prompt confirms the REMOVAL
// itself, which is the question every destructive command asks before it does anything.
// That meaning did not change.
func TestConfirmDeleteShortcut(t *testing.T) {
	if !confirmDelete(&App{noPrompt: true}, "broker x") {
		t.Error("confirmDelete with --no-prompt = false, want true")
	}
}

// TestConfirmNonTTY covers the unattended branches. Stdin is a pipe whose write end is
// CLOSED, so the read returns EOF with no input -- the "nobody answered" case, which is
// what an unattended run actually looks like. confirmDelete refuses without --no-prompt
// and confirmLayer keeps the retained layer, on any host.
//
// The closed write end is the point, not an accident: an OPEN one would block here
// exactly as it would in production, which is the one cost of deciding by reading
// rather than by inspecting the descriptor (see stdinCanAnswer).
func TestConfirmNonTTY(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	w.Close()
	defer r.Close()
	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()

	var deleted bool
	_ = captureStderr(t, func() { deleted = confirmDelete(&App{}, "broker x") })
	if deleted {
		t.Error("confirmDelete non-TTY without --no-prompt = true, want false")
	}
	if confirmLayer(&App{}, layerData) {
		t.Error("confirmLayer non-TTY = true, want false (kept)")
	}
}

// TestConfirmDowngradeNeverUnattended pins the M9 policy: unlike every other
// confirm helper here, confirmDowngrade has no --no-prompt escape at all, because
// `deploy operator` never registers that flag -- an operator downgrade is always
// a human's call, made at a terminal, or not made.
//
// This is exercised directly rather than through runRootWith/echoRunner: the
// Echo runner's isEcho() short-circuit in k8s.confirmNoDowngrade returns before
// Cluster.Confirm (and so this closure) is ever invoked, so a real downgrade
// cannot be driven to this function from the CLI dispatch layer at all. Direct
// unit coverage of the function itself is what IS reachable.
func TestConfirmDowngradeNeverUnattended(t *testing.T) {
	t.Run("non-interactive declines and tells the operator to re-run interactively", func(t *testing.T) {
		var proceed bool
		errOut := captureStderr(t, func() {
			proceed = confirmDowngrade(&App{Interactive: func() bool { return false }}, "Downgrade the operator?")
		})
		if proceed {
			t.Error("a non-interactive run must decline the downgrade")
		}
		if strings.Contains(errOut, "--no-prompt") {
			t.Errorf("refusal must not advise a flag `deploy operator` does not register: %q", errOut)
		}
		if !strings.Contains(errOut, "interactively") {
			t.Errorf("refusal must tell the operator to re-run interactively: %q", errOut)
		}
	})
	t.Run("--no-prompt is not a side-door -- still declines non-interactively", func(t *testing.T) {
		if confirmDowngrade(&App{noPrompt: true, Interactive: func() bool { return false }}, "q") {
			t.Error("noPrompt must not have been re-added as an escape for confirmDowngrade")
		}
	})
	t.Run("interactive yes proceeds", func(t *testing.T) {
		a := &App{Interactive: func() bool { return true }, PromptIn: strings.NewReader("yes\n")}
		if !confirmDowngrade(a, "Downgrade the operator?") {
			t.Error("an interactive yes must let the downgrade through")
		}
	})
	t.Run("interactive no refuses", func(t *testing.T) {
		a := &App{Interactive: func() bool { return true }, PromptIn: strings.NewReader("no\n")}
		if confirmDowngrade(a, "Downgrade the operator?") {
			t.Error("an interactive no must refuse the downgrade")
		}
	})
}

// TestDeployOperatorNoPromptStaysUnknownFlag pins the other half of M9: the fix
// removes confirmDowngrade's dead a.noPrompt branch, not the flag's absence from
// `deploy operator` -- --no-prompt was never wired there (addRemoveFlags is not
// called on this command), so it must keep failing exactly the way it did before,
// as cobra's own unknown-flag error, rather than silently becoming valid.
func TestDeployOperatorNoPromptStaysUnknownFlag(t *testing.T) {
	_, err := runRootWith(t, append(withEnv("operator", "deploy", "--no-prompt"), "--platform", "kubernetes"), echoRunner)
	if err == nil {
		t.Fatal("deploy operator --no-prompt must still fail: the flag is not registered on this command")
	}
	if !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("err = %v, want an unknown-flag error", err)
	}
}

// TestPromptYesNo pins the lenient delete confirmation: y/yes (any case) accept,
// everything else declines.
//
// The second return is the distinction the whole prompt layer now rests on. An empty
// stream is UNANSWERED, not a "no": a reply of "n" is the operator's decision, while
// EOF means nobody was there to make one, and only the latter earns the refusal line
// that names --no-prompt. A bare "\n" is a real answer -- someone pressed Enter -- and
// declines. A final line with no trailing newline still counts.
func TestPromptYesNo(t *testing.T) {
	cases := []struct {
		in           string
		want         bool
		wantAnswered bool
	}{
		{"y\n", true, true}, {"yes\n", true, true}, {"Y\n", true, true}, {"YES\n", true, true},
		{"n\n", false, true}, {"no\n", false, true}, {"\n", false, true}, {"maybe\n", false, true},
		{"y", true, true},  // no trailing newline: still an answer
		{"", false, false}, // EOF with no input: nobody answered
	}
	for _, tc := range cases {
		var out strings.Builder
		got, answered := promptYesNo(strings.NewReader(tc.in), &out, "? ")
		if got != tc.want || answered != tc.wantAnswered {
			t.Errorf("promptYesNo(%q) = (%v, %v), want (%v, %v)", tc.in, got, answered, tc.want, tc.wantAnswered)
		}
	}
}

// TestPromptYes pins the strict layer-deletion confirmation: only an exact
// (trimmed, case-insensitive) "yes" accepts; a bare "y" is not enough. Same
// answered-vs-declined split as TestPromptYesNo.
func TestPromptYes(t *testing.T) {
	cases := []struct {
		in           string
		want         bool
		wantAnswered bool
	}{
		{"yes\n", true, true}, {"YES\n", true, true}, {"  yes  \n", true, true},
		{"y\n", false, true}, {"no\n", false, true}, {"\n", false, true}, {"yess\n", false, true},
		{"yes", true, true},
		{"", false, false},
	}
	for _, tc := range cases {
		var out strings.Builder
		got, answered := promptYes(strings.NewReader(tc.in), &out, "? ")
		if got != tc.want || answered != tc.wantAnswered {
			t.Errorf("promptYes(%q) = (%v, %v), want (%v, %v)", tc.in, got, answered, tc.want, tc.wantAnswered)
		}
	}
}

func TestErrorPaths(t *testing.T) {
	t.Run("bad env path", func(t *testing.T) {
		// A value with a separator names one file: no env/ retry, so the error
		// lists that single candidate.
		_, err := runRoot(t, []string{"broker", "status", "--env", "/no/such/file.yaml", "--platform", "kubernetes"})
		if err == nil || !strings.Contains(err.Error(), "not found: looked for") {
			t.Fatalf("bad --env err = %v, want a not-found error", err)
		}
	})
	t.Run("bad container generate role via --pod", func(t *testing.T) {
		_, err := runRoot(t, withEnv("broker", "generate", "--pod", "bogus", "--platform", "docker"))
		if err == nil || !strings.Contains(err.Error(), "invalid node role") {
			t.Fatalf("docker broker generate --pod bogus err = %v, want 'invalid node role'", err)
		}
	})
	t.Run("bad container deploy role via --pod", func(t *testing.T) {
		// ParseRole runs in RunE before any host operation, so the bogus role is rejected
		// without the (real) PrepHost/Deploy ever executing. The role is a flag now, but
		// the refusal has to survive the move: --pod is what names this host's role on a
		// container platform, and a typo there must not reach a mkdir or a compose write.
		_, err := runRoot(t, withEnv("broker", "deploy", "--pod", "bogus", "--platform", "docker"))
		if err == nil || !strings.Contains(err.Error(), "invalid node role") {
			t.Fatalf("docker broker deploy --pod bogus err = %v, want 'invalid node role'", err)
		}
	})
	t.Run("bad k8s role leaf", func(t *testing.T) {
		_, err := runRoot(t, withEnv("broker", "logs", "--pod", "bogus", "--platform", "kubernetes"))
		if err == nil || !strings.Contains(err.Error(), "invalid node role") {
			t.Fatalf("k8s logs broker --pod bogus err = %v, want 'invalid node role'", err)
		}
	})
	t.Run("unknown verb under a noun", func(t *testing.T) {
		// The refusal is group()'s own, not cobra's: cobra would print help and
		// exit 0 for an unknown word on a noun that owns verbs.
		_, err := runRoot(t, withEnv("broker", "bogus", "--platform", "kubernetes"))
		if err == nil || !strings.Contains(err.Error(), "bogus") {
			t.Fatalf("generate bogus err = %v, want a refusal naming the unknown word", err)
		}
	})
}

// TestK8sGenSecretsWired covers the Secret manifests `broker generate` emits ahead of the
// CR. It uses the standalone env because the assertions are about the admin Secret, which
// every env has; the sample ships with TLS commented out, so it would render the same one.
func TestK8sGenSecretsWired(t *testing.T) {
	path := writeStandaloneEnv(t)
	out, err := runRoot(t, []string{"broker", "generate", "--env", path, "--platform", "kubernetes"})
	if err != nil {
		t.Fatalf("broker generate err = %v, want nil", err)
	}
	if !strings.HasPrefix(out, "apiVersion: v1") {
		t.Errorf("broker generate should render Secret manifests, got %q", firstLine(out))
	}
	// The manifests carry the value base64-encoded, so the raw password must not
	// appear -- but the rendering is still secret-bearing by design.
	if !strings.Contains(out, "kind: Secret") {
		t.Errorf("broker generate output is not a Secret manifest:\n%s", out)
	}
}

// TestOperatorGenerateIsTheWholeInstallStream pins what `operator generate` emits and, more
// importantly, the ORDER it emits it in.
//
// The `generate secrets operator` split is gone: one command now prints everything the
// install applies -- the namespace document, the image-pull Secret when one is configured,
// then the bundle. The order is the part worth pinning rather than the contents: the Secret
// is namespaced and its namespace only exists inside the bundle, so a stream that put it
// first would not apply to an empty cluster. Splitting it across two commands made that
// ordering the operator's problem; folding it back in makes it ours.
func TestOperatorGenerateIsTheWholeInstallStream(t *testing.T) {
	out, err := runRoot(t, []string{"operator", "generate", "--env", sampleEnv, "--platform", "kubernetes"})
	if err != nil {
		t.Fatalf("operator generate err = %v, want nil", err)
	}
	for _, want := range []string{"kind: Namespace", "kind: Secret", "name: regcred", ".dockerconfigjson"} {
		if !strings.Contains(out, want) {
			t.Errorf("operator generate should render %q:\n%s", want, out)
		}
	}
	ns := strings.Index(out, "kind: Namespace")
	secret := strings.Index(out, "name: regcred")
	deployment := strings.Index(out, "kind: Deployment")
	if !(ns < secret && secret < deployment) {
		t.Errorf("operator generate must emit namespace -> regcred -> bundle, got offsets ns=%d secret=%d deployment=%d",
			ns, secret, deployment)
	}

	// With no pull secret configured there is simply no Secret document -- that is a
	// deployable stream, not an error. The split version had to fail here, because
	// `generate secrets operator` had nothing to print and printing nothing would have
	// looked like success.
	path := writeStandaloneEnv(t) // no kubernetes.imagePullSecret
	bare, err := runRoot(t, []string{"operator", "generate", "--env", path, "--platform", "kubernetes"})
	if err != nil {
		t.Fatalf("operator generate without a pull secret err = %v, want nil", err)
	}
	if strings.Contains(bare, ".dockerconfigjson") {
		t.Error("operator generate emitted a pull secret the env file does not configure")
	}
	if !strings.Contains(bare, "kind: Deployment") {
		t.Error("operator generate without a pull secret must still emit the bundle")
	}
}

// TestGenSecretsRefusesEmptyValue is gone with `generate secrets broker`, and its
// replacement attempt (asserting deploy refuses an empty redundancy.psk) was wrong about what
// deploy does: on a container platform `broker deploy` GENERATES a PSK on the first HA run
// and writes it back to the env file, so an empty one is the normal starting state rather
// than an error.
//
// render.SecretPreflight still refuses an empty secret before one is created from it, and
// internal/render's TestSecretPreflight drives that function directly -- which is the right
// level for it, since the refusal is reached only when nothing generated the value.

// TestGenNeverLeaksSecrets is the container half of decision 6, and it is a one-way
// property now: on docker and podman `broker generate` renders the deploy artifact and
// nothing else, so no secret value can appear in it at all.
//
// Neither engine can carry a value in the artifact -- a quadlet references podman store
// secrets by name, and a compose file names an environment variable -- so there is nothing
// for a secret-bearing render to print that would actually be applyable. The values reach
// the engine from `broker deploy`, which is the only command that handles them.
func TestGenNeverLeaksSecrets(t *testing.T) {
	path := writeCtrStandaloneEnv(t)
	for _, platform := range []string{"docker", "podman"} {
		out, err := runRoot(t, []string{"broker", "generate", "--env", path, "--platform", platform})
		if err != nil {
			t.Fatalf("%s broker generate: %v", platform, err)
		}
		if strings.Contains(out, smokeAdminPass) {
			t.Errorf("%s broker generate leaked the admin password:\n%s", platform, out)
		}
	}
}

// TestConfigStepsDoNotLeakSecrets drives each config apply/disable step against a
// container fixture carrying every optional value (server cert/key, a domain CA, a
// product key) and asserts none of them prints the private key material to
// stdout. `config` no longer aggregates these into one run-everything step (there
// is no re-runnable ordering to assume), so each is exercised through its own
// command instead of one combined call -- replacing the old direct-call coverage
// of the now-deleted opCtrConfigAll.
func TestConfigStepsDoNotLeakSecrets(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.ToSlash(filepath.Join(dir, "tls.crt"))
	keyPath := filepath.ToSlash(filepath.Join(dir, "tls.key"))
	const keyMaterial = "PRIVATE-KEY-MATERIAL-do-not-log"
	if err := os.WriteFile(certPath, []byte("CERT-PEM\n"), 0o644); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte(keyMaterial+"\n"), 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}
	path := filepath.Join(dir, "cfgsteps.yaml")
	content := "redundancy:\n  enabled: false\n" +
		"  primary:\n    name: pri-host\n" +
		"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n" +
		"semp:\n  adminPass: " + smokeAdminPass + "\n" +
		"tls:\n  cert: " + certPath + "\n  certKey: " + keyPath + "\n" +
		"docker: {}\n" +
		"broker:\n" +
		"  domainCerts:\n    folder: " + filepath.ToSlash(dir) + "\n    files:\n      myca: myca.pem\n" +
		"  productKeys:\n    - KEY-1\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	for _, args := range [][]string{
		{"broker", "configure", "server-certs"},
		{"broker", "configure", "domain-certs"},
		{"broker", "configure", "product-keys"},
		// default-vpn/default-users now confirm; --no-prompt keeps this exercising
		// the real disable path instead of a silent decline.
		{"broker", "configure", "default-vpn", "--no-prompt"},
		{"broker", "configure", "default-users", "--no-prompt"},
	} {
		full := append(append([]string{}, args...), "--platform", "docker")
		out, err := runCtr(t, path, full...)
		if err != nil {
			t.Fatalf("%v err = %v, want nil", args, err)
		}
		if strings.Contains(out, keyMaterial) {
			t.Errorf("%v leaked the server-certificate private key to stdout", args)
		}
	}
}

// writeK8sDeployAllEnv builds a minimal, valid k8s env with redundancy
// overridable, so opK8sDeployAll's HA-only final Leader() branch can be
// exercised (writeStandaloneEnv is fixed at redundancy.enabled: false).
func writeK8sDeployAllEnv(t *testing.T, redundancy string) *config.Config {
	t.Helper()
	yamlBody := "redundancy:\n  enabled: " + redundancy + "\n" +
		"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n" +
		"semp:\n  adminPass: " + smokeAdminPass + "\n" +
		"kubernetes:\n" +
		"  name: dev-broker\n" +
		"  namespace: solace\n" +
		"  adminSecret: solace-admin-secret\n" +
		"  updateStrategy: automatedRolling\n" +
		"  storage:\n    class: standard\n    msgNodeSize: 30Gi\n"
	return loadDirect(t, yamlBody, config.K8s)
}

// k8sDeployAllOutputHook supplies the canned Output content opK8sDeployAll's
// happy path needs once a fake (non-Echo) Runner is in play:
// k8s.Cluster.isEcho() is a concrete type assertion on engine.Echo, so a fake
// Runner takes CheckStorageClass's real validation branch (unlike engine.Echo,
// which skips it) -- it needs a WaitForFirstConsumer/true StorageClass answer to
// pass, and Leader (HA only) needs a healthy `show redundancy` transcript to
// avoid its real poll budget.
func k8sDeployAllOutputHook(c opCall) []byte {
	switch {
	case opArgvMatch(c, "volumeBindingMode"):
		return []byte("WaitForFirstConsumer")
	case opArgvMatch(c, "allowVolumeExpansion"):
		return []byte("true")
	case opArgvMatch(c, ".show-rd.cli"):
		return []byte(healthyShowRD)
	default:
		return nil
	}
}

// TestDeployDoesNotAssertLeader is the inverse of the test that used to live here.
//
// `deploy all` asserted the config-sync leader as its final step. `broker deploy` must NOT:
// asserting the leader reverts activity on the mate, which changes redundancy state, and a
// deploy has no business making that change unasked. It is `broker perform assert-leader`
// now, and the deploy's Long text names it as the next step on HA.
func TestDeployDoesNotAssertLeader(t *testing.T) {
	cfg := writeK8sDeployAllEnv(t, "true")
	rr := &opRunner{output: k8sDeployAllOutputHook}
	a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr}
	var deployErr error
	captureStdout(t, func() { deployErr = opK8sDeploy(a) })
	if deployErr != nil {
		t.Fatalf("opK8sDeploy (HA) err = %v, want nil", deployErr)
	}
	if rr.hasCall("assert-leader") {
		t.Error("broker deploy asserted the config-sync leader; that reverts activity on the mate " +
			"and belongs to `broker perform assert-leader`")
	}
}

// TestOpK8sDeployAborts covers opK8sDeploy's error-return arms -- CreateNamespace,
// CreateSecrets, DeployBroker -- each in a sub-test that fails exactly that step and asserts
// no later step's command was issued.
//
// Two arms the `deploy all` version had are gone. There is no Check step: `broker deploy`
// runs the per-operation preflight, not the whole validate report, so a broken StorageClass
// is reported by `validate` rather than re-checked on every deploy. And there is no
// operator-apply step, because the operator is cluster-scoped and has its own noun.
func TestOpK8sDeployAborts(t *testing.T) {
	steps := []string{"create-namespace", "create-secrets", "deploy-broker"}
	for i, step := range steps {
		n := i + 1
		t.Run(step+" fails -> the next step never runs", func(t *testing.T) {
			cfg := writeK8sDeployAllEnv(t, "false")
			rr := &opRunner{fail: opFailOnCount("apply", n), output: k8sDeployAllOutputHook}
			a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr}
			var err error
			captureStdout(t, func() { err = opK8sDeploy(a) })
			if err == nil {
				t.Fatalf("opK8sDeploy = nil, want the injected %s failure to abort", step)
			}
			if got := rr.callCount("apply"); got != n {
				t.Errorf("opK8sDeploy issued %d apply command(s) after %s failed, want exactly %d (no later step ran)", got, step, n)
			}
		})
	}
}

// TestPrepLabelsIsInteractiveOnly is gone with the command it pinned. This tool no longer
// labels cluster worker nodes at all: kubernetes.placement.labels are SELECTORS, which
// internal/render puts in the CR as nodeSelector and affinity terms, and whoever owns the
// cluster owns which node carries which label.
//
// That also removed the one prerequisite that could not be folded into an idempotent
// `broker deploy`: the picker needed a human at a terminal and recorded its choice nowhere,
// so a re-run could label a different node than the first run did.

// TestDeployNeverLabelsNodes keeps the property the labelling tests were really about: a
// deploy touches no node. It used to be true because `deploy all` deliberately left the
// picker out; it is now true because nothing in the tool labels nodes at all.
//
// The terminal is still made interactive here on purpose. That is what the old bug looked
// like -- `deploy all` ran the picker when placement was configured and stdin happened to be
// a terminal, so the same command was interactive or not depending on where it ran.
func TestDeployNeverLabelsNodes(t *testing.T) {
	cfg := loadDirect(t, "redundancy:\n  enabled: false\n"+
		"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n"+
		"semp:\n  adminPass: "+smokeAdminPass+"\n"+
		"kubernetes:\n  name: dev-broker\n  namespace: solace\n"+
		"  storage:\n    class: standard\n    msgNodeSize: 30Gi\n"+
		"  placement:\n    labelsPrimary: [\"solace-node: primary\"]\n", config.K8s)
	rr := &opRunner{output: func(c opCall) []byte {
		for _, a := range c.args {
			if strings.Contains(a, "volumeBindingMode") {
				return []byte("WaitForFirstConsumer\n")
			}
			if strings.Contains(a, "allowVolumeExpansion") {
				return []byte("true\n")
			}
		}
		return nil
	}}
	a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr,
		Interactive: func() bool { return true }}
	var err error
	captureStdout(t, func() { err = opK8sDeploy(a) })
	if err != nil {
		t.Fatalf("broker deploy err = %v, want nil", err)
	}
	for _, c := range rr.calls {
		if len(c.args) > 1 && c.args[0] == "label" {
			t.Errorf("broker deploy labelled a node: %v", c.args)
		}
	}
}

// TestOpK8sPrepAllAborts is gone: `prepare all` is gone. The namespace-then-secrets sequence
// it pinned is the first half of `broker deploy` now, and its abort ordering is covered by
// TestOpK8sDeployAborts above -- one test over one command, rather than the same property
// asserted twice because two commands ran the same steps.

// TestOpK8sRemoveAborts covers opK8sRemoveBroker's error-return arms (DeleteBroker,
// DeleteSecrets). The property is a real one about teardown ORDER, not error forwarding: a
// removal must not go on to ask about the namespace after a broker- or secrets-deletion
// actually failed, because that question is only meaningful once everything this env file
// owns is gone. Getting it wrong would offer to cascade-delete a namespace still holding a
// broker whose deletion had failed.
func TestOpK8sRemoveAborts(t *testing.T) {
	steps := []struct {
		name string
		n    int
	}{
		{"delete-broker", 1},
		{"delete-secrets", 2},
	}
	for _, st := range steps {
		t.Run(st.name+" fails -> the namespace is never inspected", func(t *testing.T) {
			cfg := writeK8sDeployAllEnv(t, "false")
			rr := &opRunner{fail: opFailOnCount("delete", st.n)}
			a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr, noPrompt: true}
			var err error
			captureStdout(t, func() { err = opK8sRemoveBroker(a) })
			if err == nil {
				t.Fatalf("opK8sRemoveBroker = nil, want the injected %s failure to abort", st.name)
			}
			if got := rr.callCount("delete"); got != st.n {
				t.Errorf("opK8sRemoveBroker issued %d delete command(s) after %s failed, want exactly %d",
					got, st.name, st.n)
			}
		})
	}
}

// TestOpCtrVerifyRedundancyRunsCoordinated covers opCtrVerifyRedundancy's
// actual failover exercise arm (RedundancyCoordinated) rather than only the
// skip/reject arms already covered by TestCtrRoleGuards. It sets this host as
// the primary and supplies a canned `show redundancy` transcript over a fake
// Runner, since engine.Echo's fixed empty output would otherwise send
// RedundancyCoordinated into its real poll loop (broker.New's 2s x 60 budget,
// which the CLI has no seam to shorten).
func TestOpCtrVerifyRedundancyRunsCoordinated(t *testing.T) {
	host, err := os.Hostname()
	if err != nil {
		t.Skipf("os.Hostname unavailable: %v", err)
	}
	yamlBody := "redundancy:\n  enabled: true\n" +
		"  primary:\n    name: '" + host + "'\n    addr: 10.0.0.11\n" +
		"  backup:\n    name: ctr-coordinated-backup\n    addr: 10.0.0.12\n" +
		"  monitor:\n    name: ctr-coordinated-monitor\n    addr: 10.0.0.13\n" +
		"  psk: SHARED-PRESHARED-KEY\n" +
		"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n" +
		"semp:\n  adminPass: " + smokeAdminPass + "\n"
	cfg := loadDirect(t, yamlBody, config.Docker)
	rr := &opRunner{output: func(c opCall) []byte {
		if opArgvMatch(c, ".show-rd.cli") {
			// Non-empty but deliberately unhealthy (no Configuration/Redundancy
			// Status lines): primaryRedundancyUp is false, so RedundancyCoordinated
			// returns its health error immediately -- proving it ran, without ever
			// entering a real poll loop or touching the mate SEMP channel.
			return []byte("Activity Status: Local Active\n")
		}
		return nil
	}}
	a := &App{Cfg: cfg, Platform: config.Docker, Runner: rr}
	captureStdout(t, func() { err = opCtrVerifyRedundancy(a) })
	if err == nil || !strings.Contains(err.Error(), "redundancy configuration/status is not healthy") {
		t.Fatalf("opCtrVerifyRedundancy (primary, active-but-unhealthy) err = %v, want the redundancy-unhealthy error", err)
	}
}

// TestContainerRoleDetectsFromHostname covers M1. `config.ParseRole("")` returns
// Primary whatever the redundancy setting is, so omitting the role on an HA
// BACKUP host used to deploy a second PRIMARY into the group -- silently, and
// while `deploy broker`'s own Short claimed the role was "required in HA", which
// nothing enforced. Two primaries surface much later as a redundancy failure that
// points nowhere near the command that caused it.
//
// containerRole now detects the node from this host's name, says which it chose,
// and fails loud rather than guessing. An explicit role still wins outright.
func TestContainerRoleDetectsFromHostname(t *testing.T) {
	const ha = "redundancy:\n  enabled: true\n" +
		"  primary:\n    name: pri-host\n    addr: 10.0.0.11\n" +
		"  backup:\n    name: bkp-host\n    addr: 10.0.0.12\n" +
		"  monitor:\n    name: mon-host\n    addr: 10.0.0.13\n" +
		"  psk: SHARED-PRESHARED-KEY\n" +
		"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n" +
		"semp:\n  adminPass: " + smokeAdminPass + "\n"

	appOn := func(t *testing.T, body, host string) *App {
		t.Helper()
		return &App{
			Cfg:      loadDirect(t, body, config.Docker),
			Platform: config.Docker,
			Runner:   engine.Echo{W: io.Discard},
			Hostname: func() (string, error) { return host, nil },
			// No interface addresses: detection's second pass must not depend on how
			// the machine running the suite is numbered. The pass itself is covered in
			// internal/broker (TestDetectRoleByAddress).
			LocalAddrs: func() (map[string]bool, error) { return nil, nil },
		}
	}

	t.Run("an omitted role is detected, and announced", func(t *testing.T) {
		a := appOn(t, ha, "bkp-host")
		var role config.Role
		var err error
		errOut := captureStderr(t, func() { role, err = containerRole(a) })
		if err != nil {
			t.Fatalf("containerRole: %v", err)
		}
		if role != config.Backup {
			t.Errorf("role = %q, want backup -- this host is the backup", role)
		}
		if !strings.Contains(errOut, "backup") {
			t.Errorf("stderr = %q, want the detected role announced", errOut)
		}
	})

	t.Run("an explicit role wins over the hostname, but is checked against it", func(t *testing.T) {
		// The escape hatch stays an escape hatch: --pod is obeyed. It is no longer
		// obeyed SILENTLY, though -- naming the wrong role is how a second primary
		// joins a group, and the mistake is invisible afterwards. A warning, never a
		// prompt: the operator said which node this is, and a scripted deploy across
		// three hosts must not stop to ask.
		a := appOn(t, ha, "bkp-host")
		a.pod = "monitor"
		var role config.Role
		var err error
		errOut := captureStderr(t, func() { role, err = containerRole(a) })
		if err != nil {
			t.Fatalf("containerRole: %v", err)
		}
		if role != config.Monitor {
			t.Errorf("role = %q, want monitor -- an explicit role is the escape hatch for a host whose "+
				"name does not match the env file", role)
		}
		for _, want := range []string{"monitor", "backup", "proceeding"} {
			if !strings.Contains(errOut, want) {
				t.Errorf("stderr = %q, want a mismatch warning naming %q", errOut, want)
			}
		}
	})

	t.Run("an explicit role that agrees with the host says nothing", func(t *testing.T) {
		a := appOn(t, ha, "bkp-host")
		a.pod = "backup"
		var err error
		errOut := captureStderr(t, func() { _, err = containerRole(a) })
		if err != nil {
			t.Fatalf("containerRole: %v", err)
		}
		if strings.Contains(errOut, "proceeding") {
			t.Errorf("stderr = %q, want no warning when --pod and the host agree", errOut)
		}
	})

	t.Run("an unrecognisable host does not turn an explicit role into a warning", func(t *testing.T) {
		// Detection failing is not a mismatch. A host whose name matches nothing is
		// exactly the case --pod exists for, so warning there would fire on every
		// correct use of the flag and teach the operator to ignore it.
		a := appOn(t, ha, "somewhere-else")
		a.pod = "monitor"
		var err error
		errOut := captureStderr(t, func() { _, err = containerRole(a) })
		if err != nil {
			t.Fatalf("containerRole: %v", err)
		}
		if strings.Contains(errOut, "proceeding") {
			t.Errorf("stderr = %q, want silence: undetectable is not disagreement", errOut)
		}
	})

	t.Run("an unrecognised host fails loud", func(t *testing.T) {
		a := appOn(t, ha, "somewhere-else")
		if _, err := containerRole(a); err == nil {
			t.Fatal("a hostname matching no configured node must not silently deploy a second primary")
		} else if !strings.Contains(err.Error(), "primary|backup|monitor") {
			t.Errorf("error = %v, want it to name the roles that can be passed instead", err)
		}
	})

	t.Run("rendering falls back where deploying refuses", func(t *testing.T) {
		// `generate broker` changes nothing, and the artifact it prints NAMES the
		// node it is for -- so a fallback is visible in the output the operator is
		// already reading. Refusing here would make rendering the backup's artifact
		// on a laptop impossible, which is an ordinary thing to want. The same
		// hostname is fatal for a deploy, which the subtest above pins.
		a := appOn(t, ha, "somewhere-else")
		var role config.Role
		var err error
		errOut := captureStderr(t, func() { role, err = containerRenderRole(a) })
		if err != nil {
			t.Fatalf("containerRenderRole must not refuse to render on an unrecognised host: %v", err)
		}
		if role != config.Primary {
			t.Errorf("role = %q, want primary as the rendering fallback", role)
		}
		if !strings.Contains(errOut, "primary") {
			t.Errorf("stderr = %q, want the fallback stated rather than silent", errOut)
		}
	})

	t.Run("standalone still defaults without detecting", func(t *testing.T) {
		standalone := strings.Replace(ha, "enabled: true", "enabled: false", 1)
		a := appOn(t, standalone, "somewhere-else")
		var role config.Role
		var err error
		errOut := captureStderr(t, func() { role, err = containerRole(a) })
		if err != nil {
			t.Fatalf("containerRole (standalone): %v", err)
		}
		if role != config.Primary {
			t.Errorf("role = %q, want primary: standalone has one node, so the argument means nothing", role)
		}
		if strings.Contains(errOut, "detected") {
			t.Errorf("stderr = %q, want no detection line when there is nothing to detect", errOut)
		}
	})
}

// TestStandaloneRouternameFallsBackToTheHost pins the load-time fill
// (App.fillStandaloneNodeName) end to end: a standalone container env file that names no
// routername must deploy a broker called after the HOST, not one called "".
//
// Through the real command rather than the helper, because the point of doing this at load
// is that every reader agrees -- the rendered artifact's hostname AND its routername here,
// and the check report and DNS check elsewhere, all off one filled value.
func TestStandaloneRouternameFallsBackToTheHost(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nameless.yaml")
	content := "redundancy:\n  enabled: false\n" +
		"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n" +
		"semp:\n  adminPass: " + smokeAdminPass + "\n" +
		"docker: {}\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write nameless env: %v", err)
	}

	var errOut string
	var out string
	var err error
	errOut = captureStderr(t, func() {
		out, err = runRootWith(t, []string{"broker", "generate", "--env", path, "--platform", "docker"},
			func(a *App) {
				a.Hostname = func() (string, error) { return "host-of-record", nil }
				echoRunner(a)
			})
	})
	if err != nil {
		t.Fatalf("broker generate on a nameless standalone env: %v", err)
	}
	for _, want := range []string{`hostname: "host-of-record"`, `routername: "host-of-record"`} {
		if !strings.Contains(out, want) {
			t.Errorf("artifact does not contain %s -- the routername did not fall back to the host:\n%s", want, out)
		}
	}
	if !strings.Contains(errOut, "host-of-record") {
		t.Errorf("stderr = %q, want the fallback announced rather than applied silently", errOut)
	}
}

// TestConfiguredRouternameSurvivesTheFallback is the other half: the fill must never
// overwrite a name the env file already carries.
func TestConfiguredRouternameSurvivesTheFallback(t *testing.T) {
	path := writeCtrStandaloneEnv(t) // names the primary "pri-host"
	out, err := runRootWith(t, []string{"broker", "generate", "--env", path, "--platform", "docker"},
		func(a *App) {
			a.Hostname = func() (string, error) { return "host-of-record", nil }
			echoRunner(a)
		})
	if err != nil {
		t.Fatalf("broker generate: %v", err)
	}
	if strings.Contains(out, "host-of-record") {
		t.Errorf("the host's name overwrote the configured routername:\n%s", out)
	}
	if !strings.Contains(out, `hostname: "pri-host"`) {
		t.Errorf("artifact does not carry the configured routername:\n%s", out)
	}
}

// TestK8sSmokeRedundancyUnhealthy covers opK8sVerifyRedundancy's error-return: on
// the HA sample over the echo seam, engine.Echo's empty `show redundancy` output
// makes primaryRedundancyUp false, so Redundancy fails on its first check (no poll).
func TestK8sSmokeRedundancyUnhealthy(t *testing.T) {
	_, err := runRootWith(t, withEnv("broker", "perform", "redundancy-test", "--platform", "kubernetes"), echoRunner)
	if err == nil || !strings.Contains(err.Error(), "redundancy configuration/status is not healthy") {
		t.Fatalf("k8s smoke redundancy (HA sample) err = %v, want the redundancy-unhealthy error", err)
	}
}

// TestK8sConfigDeleteDomainCertsConfigured covers domainCANames' loop body: every
// other test's domainCerts.files map is empty, so the map-to-slice conversion
// feeding RemoveDomainCerts is trivially correct by vacuity. This configures one
// CA and asserts `config delete domain-certs` actually issues a kubectl exec
// rather than self-skipping.
func TestK8sConfigDeleteDomainCertsConfigured(t *testing.T) {
	path := filepath.Join(t.TempDir(), "domaincerts.yaml")
	content := "redundancy:\n  enabled: false\n" +
		"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n" +
		"semp:\n  adminPass: " + smokeAdminPass + "\n" +
		"kubernetes:\n" +
		"  name: dev-broker\n" +
		"  namespace: solace\n" +
		"  adminSecret: solace-admin-secret\n" +
		"  updateStrategy: automatedRolling\n" +
		"  storage:\n    class: standard\n    msgNodeSize: 30Gi\n" +
		"broker:\n" +
		"  domainCerts:\n    folder: certs\n    files:\n      myca: myca.pem\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}
	out, err := runRootWith(t, []string{"broker", "configure", "domain-certs", "--remove", "--no-prompt",
		"--env", path, "--platform", "kubernetes"}, echoRunner)
	if err != nil {
		t.Fatalf("config delete domain-certs (configured) err = %v, want nil", err)
	}
	if !strings.Contains(out, "+ kubectl") {
		t.Errorf("config delete domain-certs (configured) stdout = %q, want a '+ kubectl ...' echo", out)
	}
}

// bashEnv is a minimal legacy container env file used by the convert tests.
const bashEnv = "#!/bin/bash\n" +
	"SOLBK_IMAGE=\"solace-pubsub-standard\"\n" +
	"SOLBK_IMG_TAG=\"10.10.1.128\"\n" +
	"SOLBK_ADM_PASS=\"" + smokeAdminPass + "\"\n" +
	"SOLBK_REDUNDANCY=\"no\"\n" +
	"SOLBK_NODE_PRI_NAME=\"pri-host\"\n" +
	"DOCKER_MODE=\"compose\"\n" +
	"SOMETHING_UNKNOWN=\"x\"\n"

// writeBashEnv writes bashEnv to a temp file and returns its path.
func writeBashEnv(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "legacy-env")
	if err := os.WriteFile(path, []byte(bashEnv), 0o600); err != nil {
		t.Fatalf("write bash env: %v", err)
	}
	return path
}

// TestConvertToStdout covers the default path: the YAML lands on stdout, and the
// unmapped-variable warning goes to stderr so it never pollutes the artifact.
func TestConvertToStdout(t *testing.T) {
	src := writeBashEnv(t)
	var out string
	errOut := captureStderr(t, func() {
		var err error
		out, err = runRoot(t, []string{"convert", src})
		if err != nil {
			t.Fatalf("convert err = %v, want nil", err)
		}
	})
	for _, want := range []string{"enabled: \"false\"", "repo: solace-pubsub-standard", "name: pri-host"} {
		if !strings.Contains(out, want) {
			t.Errorf("converted YAML missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "SOMETHING_UNKNOWN") {
		t.Errorf("the unmapped-variable note must not reach stdout:\n%s", out)
	}
	if !strings.Contains(errOut, "SOMETHING_UNKNOWN") {
		t.Errorf("stderr = %q, want the unmapped-variable warning", errOut)
	}
}

// TestConvertToFile covers -o, including the refusal to clobber an existing file and the
// --no-prompt that answers it. `convert` writes a file carrying the source's secrets
// verbatim, so it asks before replacing one, the same way `examples` and `generate` do.
func TestConvertToFile(t *testing.T) {
	src := writeBashEnv(t)
	dst := filepath.Join(t.TempDir(), "converted.yaml")

	_ = captureStderr(t, func() {
		if _, err := runRoot(t, []string{"convert", src, "-o", dst}); err != nil {
			t.Fatalf("convert -o err = %v, want nil", err)
		}
	})
	body, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read converted file: %v", err)
	}
	if !strings.Contains(string(body), "image:") {
		t.Errorf("converted file looks wrong:\n%s", body)
	}

	_ = captureStderr(t, func() {
		_, err := runRoot(t, []string{"convert", src, "-o", dst})
		if err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
			t.Fatalf("second convert err = %v, want a refusal", err)
		}
	})
	_ = captureStderr(t, func() {
		if _, err := runRoot(t, []string{"convert", src, "-o", dst, "--no-prompt"}); err != nil {
			t.Fatalf("convert --no-prompt err = %v, want nil", err)
		}
	})
}

// TestConvertRoundTrip proves the converter's output is loadable: convert, then
// drive a real command with -e against the result.
func TestConvertRoundTrip(t *testing.T) {
	src := writeBashEnv(t)
	dst := filepath.Join(t.TempDir(), "converted.yaml")
	_ = captureStderr(t, func() {
		if _, err := runRoot(t, []string{"convert", src, "-o", dst}); err != nil {
			t.Fatalf("convert err = %v, want nil", err)
		}
	})
	out, err := runCtr(t, dst, "broker", "status", "--platform", "docker")
	if err != nil {
		t.Fatalf("docker status broker against the converted env err = %v, want nil", err)
	}
	if !strings.Contains(out, "+ docker") {
		t.Errorf("converted env did not drive a real command:\n%s", out)
	}
}

func TestConvertErrorPaths(t *testing.T) {
	src := writeBashEnv(t)
	// convert reads the same --platform every other command does, so the rejection
	// is config.ParsePlatform's own and names the platform rather than the flag --
	// one word, one meaning, whichever command it was typed on.
	t.Run("bad platform", func(t *testing.T) {
		_, err := runRoot(t, []string{"convert", src, "--platform", "bogus"})
		if err == nil || !strings.Contains(err.Error(), "invalid platform") {
			t.Fatalf("err = %v, want an invalid-platform error", err)
		}
	})
	t.Run("missing source file", func(t *testing.T) {
		_, err := runRoot(t, []string{"convert", filepath.Join(t.TempDir(), "nope")})
		if err == nil || !strings.Contains(err.Error(), "read bash env file") {
			t.Fatalf("err = %v, want a read error", err)
		}
	})
	t.Run("no source file", func(t *testing.T) {
		if _, err := runRoot(t, []string{"convert"}); err == nil {
			t.Fatal("convert with no argument should fail")
		}
	})
}

// TestConvertParseError covers runConvert's convert.Convert error return: a
// malformed legacy env file (an unterminated array assignment) is a real,
// actionable failure a migrating user can actually hit.
func TestConvertParseError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad-array.env")
	if err := os.WriteFile(path, []byte("SOLBK_TLS_CERTCAS=(\n  \"/a\"\n"), 0o600); err != nil {
		t.Fatalf("write malformed env: %v", err)
	}
	_, err := runRoot(t, []string{"convert", path})
	if err == nil || !strings.Contains(err.Error(), "unterminated array") {
		t.Fatalf("convert (unterminated array) err = %v, want it to name the parse failure", err)
	}
}

// TestConvertWriteError covers runConvert's os.WriteFile error return: an -o path
// in a directory that does not exist is a real mistake a migrating user can make.
func TestConvertWriteError(t *testing.T) {
	src := writeBashEnv(t)
	dst := filepath.Join(t.TempDir(), "missing-subdir", "out.yaml")
	_, err := runRoot(t, []string{"convert", src, "-o", dst})
	// runConvert formats the path with %q, so the expectation is built the same way:
	// a raw path would match on Linux and never on Windows, where %q escapes every
	// separator.
	want := fmt.Sprintf("write %q", dst)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("convert -o (missing parent dir) err = %v, want it to contain %s", err, want)
	}
}

// TestVersionPrintsStampedValue: `version` reports whatever the linker (the
// dev scripts' -X flag) set the package var to, verbatim -- the contract the
// release binaries depend on to match the git tag.
func TestVersionPrintsStampedValue(t *testing.T) {
	old := version
	version = "v9.9.9-test"
	t.Cleanup(func() { version = old })

	out, err := runRoot(t, []string{"version"})
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if !strings.Contains(out, "v9.9.9-test") {
		t.Errorf("output = %q, want it to contain the stamped version", out)
	}
}

// TestVersionDefaultsToDev: an unstamped build -- this package's own `go test`,
// or a plain `go build .` -- reports "dev".
func TestVersionDefaultsToDev(t *testing.T) {
	out, err := runRoot(t, []string{"version"})
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if !strings.HasPrefix(out, "solace-util dev ") {
		t.Errorf("output = %q, want it to start with %q", out, "solace-util dev ")
	}
}

// TestVersionIncludesToolchainAndPlatform: support triage needs the Go
// toolchain and OS/arch that built the binary alongside the tag.
func TestVersionIncludesToolchainAndPlatform(t *testing.T) {
	out, err := runRoot(t, []string{"version"})
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	for _, want := range []string{runtime.Version(), runtime.GOOS, runtime.GOARCH} {
		if !strings.Contains(out, want) {
			t.Errorf("output = %q, want it to contain %q", out, want)
		}
	}
}

// TestVersionRejectsArgs: NoArgs is enforced like every other bare leaf.
func TestVersionRejectsArgs(t *testing.T) {
	if _, err := runRoot(t, []string{"version", "extra"}); err == nil {
		t.Error("version with an argument: want an error, got nil")
	}
}

// TestBashEnvGivenToEnvFlag is the other half of the migration story: pointing
// -e at a legacy bash file must say it is not YAML and name the converter.
func TestBashEnvGivenToEnvFlag(t *testing.T) {
	src := writeBashEnv(t)
	_, err := runRoot(t, []string{"broker", "status", "-e", src, "--platform", "kubernetes"})
	if err == nil {
		t.Fatal("a bash env file should not load")
	}
	for _, want := range []string{"not valid YAML", "this looks like a legacy bash env file", "solace-util convert"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err.Error(), want)
		}
	}
}

func TestExecute(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"solace-util", "--help"}
	var err error
	captureStdout(t, func() { err = Execute() })
	if err != nil {
		t.Fatalf("Execute(--help) err = %v, want nil", err)
	}
}

// TestK8sConfirmDeclined covers the confirm-declined branch of every removal: a
// non-interactive run without --no-prompt must issue zero cluster calls -- the
// actual safety default.
//
// There is one remove per noun now: the secrets and namespace steps that used to be their
// own commands are folded into `broker remove`. The property they were here to protect is
// unchanged and is the single worst outcome in the tree -- an unattended run reaching
// kubectl and deleting a namespace holding more than this env file put there -- but it is
// now guarded structurally as well: the namespace is only offered when nothing else is left
// in it (TestRemoveKeepsAnOccupiedNamespace).
func TestK8sConfirmDeclined(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"remove broker declined", []string{"broker", "remove", "--platform", "kubernetes"}},
		{"remove broker --delete-data declined", []string{"broker", "remove", "--delete-data", "--platform", "kubernetes"}},
		{"remove operator declined", []string{"operator", "remove", "--platform", "kubernetes"}},
		{"remove operator --delete-crd declined", []string{"operator", "remove", "--delete-crd", "--platform", "kubernetes"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runRootWith(t, withEnv(tc.args...), func(a *App) {
				a.Interactive = func() bool { return false }
				echoRunner(a)
			})
			if err != nil {
				t.Fatalf("%s err = %v, want nil", tc.name, err)
			}
			if strings.Contains(out, "+ kubectl") {
				t.Errorf("%s stdout = %q, want no kubectl echo (declined)", tc.name, out)
			}
		})
	}
}

// TestK8sPromptsNameNamespaceAndContext covers H5. A Kubernetes prompt used to
// name the broker and nothing else, leaving the two facts that decide blast
// radius unsaid: the namespace (in the env file, but not in the question, so an
// operator with several env files open had nothing to check it against) and the
// kubeconfig context, which is not in the env file AT ALL -- it alone decides
// which CLUSTER the call lands in, so a file saying "dev" against a context that
// drifted to prod reads exactly the same.
//
// The context is injected here the way the runner and the interactive seam are:
// announceKubeContext resolves it from `kubectl config current-context` during
// load, which a test-supplied runner deliberately skips.
func TestK8sPromptsNameNamespaceAndContext(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want []string
	}{
		{"remove broker", []string{"broker", "remove"}, []string{"broker dev-broker", "namespace solace", "context prod-cluster"}},
		// The operator is cluster-scoped and lives in ITS OWN namespace, so this
		// prompt must not repeat the broker's -- naming the wrong location is
		// worse than naming none.
		{"remove operator", []string{"operator", "remove"}, []string{"operator in namespace pubsubplus-operator-system", "context prod-cluster"}},
		{"restart broker", []string{"broker", "restart"}, []string{"namespace solace", "context prod-cluster"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			args := append(append([]string{}, tc.args...), "--platform", "kubernetes")
			errOut := captureStderr(t, func() {
				_, err := runRootWith(t, withEnv(args...), func(a *App) {
					a.Interactive = func() bool { return false } // declines, but still renders the refusal
					a.kubeContext = "prod-cluster"
					echoRunner(a)
				})
				if err != nil {
					t.Fatalf("%s err = %v, want nil", tc.name, err)
				}
			})
			for _, want := range tc.want {
				if !strings.Contains(errOut, want) {
					t.Errorf("%s prompt = %q, want it to name %q", tc.name, errOut, want)
				}
			}
		})
	}
}

// TestK8sPromptsOmitAnUnknownContext pins the other half: when the context could
// not be resolved (no kubeconfig current-context, or a non-Kubernetes run) the
// clause is dropped entirely. A prompt reading "(context )" would be worse than
// one that never mentioned it.
func TestK8sPromptsOmitAnUnknownContext(t *testing.T) {
	errOut := captureStderr(t, func() {
		_, err := runRootWith(t, withEnv("broker", "remove", "--platform", "kubernetes"), func(a *App) {
			a.Interactive = func() bool { return false }
			echoRunner(a) // leaves kubeContext empty, as a real unresolved lookup does
		})
		if err != nil {
			t.Fatalf("remove broker err = %v, want nil", err)
		}
	})
	if !strings.Contains(errOut, "namespace solace") {
		t.Errorf("prompt = %q, want it to still name the namespace", errOut)
	}
	if strings.Contains(errOut, "context") {
		t.Errorf("prompt = %q, want no context clause when none was resolved", errOut)
	}
}

// TestK8sRestartRollsOrTargetsOnePod is the regression guard for the CRITICAL
// distinction H2 depends on: opK8sRestart must keep reading app.pod RAW rather
// than through podRole, because podRole's empty-defaults-to-primary would
// silently turn "restart every pod" into "restart the primary pod alone". With
// --pod unset every pod (monitor, backup, primary) is bounced in that order;
// with --pod set, only that one pod is. env/sample.yaml (withEnv) is HA and
// names the broker "dev-broker", so its pods are dev-broker-pubsubplus-<p|b|m>-0.
func TestK8sRestartRollsOrTargetsOnePod(t *testing.T) {
	t.Run("no --pod rolls every pod", func(t *testing.T) {
		out, err := runRootWith(t, append(withEnv("broker", "restart", "--no-prompt"), "--platform", "kubernetes"), echoRunner)
		if err != nil {
			t.Fatalf("restart broker err = %v, want nil", err)
		}
		for _, pod := range []string{"dev-broker-pubsubplus-m-0", "dev-broker-pubsubplus-b-0", "dev-broker-pubsubplus-p-0"} {
			if !strings.Contains(out, pod) {
				t.Errorf("restart broker stdout = %q, want a rolling restart touching %s", out, pod)
			}
		}
	})
	t.Run("--pod backup restarts only that pod", func(t *testing.T) {
		out, err := runRootWith(t, append(withEnv("broker", "restart", "--pod", "backup", "--no-prompt"),
			"--platform", "kubernetes"), echoRunner)
		if err != nil {
			t.Fatalf("restart broker --pod backup err = %v, want nil", err)
		}
		if !strings.Contains(out, "dev-broker-pubsubplus-b-0") {
			t.Errorf("restart broker --pod backup stdout = %q, want the backup pod restarted", out)
		}
		for _, pod := range []string{"dev-broker-pubsubplus-m-0", "dev-broker-pubsubplus-p-0"} {
			if strings.Contains(out, pod) {
				t.Errorf("restart broker --pod backup stdout = %q, want no other pod touched (%s)", out, pod)
			}
		}
	})
}

// TestK8sRestartConfirmGate covers opK8sRestart's confirmation gate: a
// non-interactive run (no --no-prompt) must bounce nothing, whether restarting every pod
// or a single one, and a bad role is rejected before any prompt is even possible.
func TestK8sRestartConfirmGate(t *testing.T) {
	t.Run("restart broker (all) declined (no --no-prompt)", func(t *testing.T) {
		out, err := runRootWith(t, withEnv("broker", "restart", "--platform", "kubernetes"), func(a *App) {
			a.Interactive = func() bool { return false }
			echoRunner(a)
		})
		if err != nil {
			t.Fatalf("restart broker declined err = %v, want nil", err)
		}
		if strings.Contains(out, "+ kubectl") {
			t.Errorf("restart broker declined stdout = %q, want no kubectl echo", out)
		}
	})
	t.Run("restart broker one role declined (no --no-prompt)", func(t *testing.T) {
		out, err := runRootWith(t, withEnv("broker", "restart", "--pod", "backup", "--platform", "kubernetes"), func(a *App) {
			a.Interactive = func() bool { return false }
			echoRunner(a)
		})
		if err != nil {
			t.Fatalf("restart broker --pod backup declined err = %v, want nil", err)
		}
		if strings.Contains(out, "+ kubectl") {
			t.Errorf("restart broker --pod backup declined stdout = %q, want no kubectl echo", out)
		}
	})
	t.Run("bad role rejected before any prompt", func(t *testing.T) {
		_, err := runRoot(t, withEnv("broker", "restart", "--pod", "bogus", "--platform", "kubernetes"))
		if err == nil || !strings.Contains(err.Error(), "invalid node role") {
			t.Fatalf("restart broker --pod bogus err = %v, want 'invalid node role'", err)
		}
	})
}

// TestCtrConfirmDeclined covers opCtrRemoveBroker's confirm-declined branch, the
// container-side counterpart of TestK8sConfirmDeclined: a non-interactive removal
// (no --no-prompt) must issue zero runtime calls.
func TestCtrConfirmDeclined(t *testing.T) {
	path := writeCtrStandaloneEnv(t)
	out, err := runRootWith(t, []string{"broker", "remove", "--env", path, "--platform", "docker"}, func(a *App) {
		a.Interactive = func() bool { return false }
		echoRunner(a)
	})
	if err != nil {
		t.Fatalf("docker remove broker declined err = %v, want nil", err)
	}
	if strings.Contains(out, "+ docker") {
		t.Errorf("docker remove broker declined stdout = %q, want no docker echo", out)
	}
}

// TestCtrRestartConfirmGate covers opCtrRestartBroker's confirmAction gate, the
// container-side counterpart of TestK8sRestartConfirmGate: a non-interactive run
// (no --no-prompt) must bounce nothing, and --no-prompt must issue the restart, on
// both docker and podman.
func TestCtrRestartConfirmGate(t *testing.T) {
	path := writeCtrStandaloneEnv(t)
	cases := []struct {
		platform  string
		wantEcho  string
		noPrompt  bool
		wantIssue bool
	}{
		{"docker", "+ docker", false, false},
		{"docker", "+ docker", true, true},
		{"podman", "+ systemctl", false, false},
		{"podman", "+ systemctl", true, true},
	}
	for _, tc := range cases {
		name := fmt.Sprintf("%s restart broker no-prompt=%v", tc.platform, tc.noPrompt)
		t.Run(name, func(t *testing.T) {
			args := []string{"broker", "restart", "--env", path, "--platform", tc.platform}
			if tc.noPrompt {
				args = append(args, "--no-prompt")
			}
			out, err := runRootWith(t, args, func(a *App) {
				a.Interactive = func() bool { return false }
				echoRunner(a)
			})
			if err != nil {
				t.Fatalf("%s err = %v, want nil", name, err)
			}
			gotIssue := strings.Contains(out, tc.wantEcho)
			if gotIssue != tc.wantIssue {
				t.Errorf("%s stdout = %q, want issued=%v (%q)", name, out, tc.wantIssue, tc.wantEcho)
			}
		})
	}
}

// TestStdinCanAnswerClosedFile covers stdinCanAnswer's fail-safe branch: a stdin whose
// Stat() cannot be evaluated (a closed file, on both Windows and POSIX) is a stream no
// reply can arrive on, so the question must not even be asked.
func TestStdinCanAnswerClosedFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "closed-*")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}
	old := os.Stdin
	os.Stdin = f
	defer func() { os.Stdin = old }()

	if stdinCanAnswer() {
		t.Error("stdinCanAnswer with a closed stdin = true, want false (fail-safe)")
	}
	if interactive(&App{}) {
		t.Error("interactive with a closed stdin = true, want false")
	}
}

// TestPromptAsksWhenStdinIsNotATTY is the regression for the live failure that started
// this: `export-config -o <existing file>` refused to overwrite WITHOUT ASKING, on a
// terminal where the operator was sitting right there. The cause was that interactive()
// asked isTTY(os.Stdin), and that shell's stdin was a pipe rather than a character
// device -- which silently turned EVERY confirm in the tool (broker remove, broker
// restart, the import tear-down gate) into a refusal.
//
// A pipe carrying a reply is now exactly as good as a console: the question is asked,
// and the answer on stdin is honoured. Note what this test does NOT do -- it does not
// open a terminal. The first attempt at this fix read /dev/tty / CONIN$ directly, which
// hung forever in a shell where the keystrokes go to the pipe instead.
func TestPromptAsksWhenStdinIsNotATTY(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	defer r.Close()
	if _, err := w.WriteString("y\n"); err != nil {
		t.Fatalf("write answer: %v", err)
	}
	w.Close()
	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()

	a := &App{} // no seams: the production path, end to end
	if !interactive(a) {
		t.Fatal("interactive with a pipe on stdin = false, want true")
	}

	var ok bool
	stderrText := captureStderr(t, func() { ok = confirmAction(a, "Overwrite", "overwrite", "out.cli") })
	if !ok {
		t.Error(`confirmAction answered "y" on a piped stdin = false, want true`)
	}
	if !strings.Contains(stderrText, "Overwrite out.cli?") {
		t.Errorf("the question was not asked: %q", stderrText)
	}
}

// TestUnansweredPromptStillRefuses covers the genuinely unattended run. Stdin is
// readable but ends immediately, so the question is ASKED and then goes unanswered --
// which must refuse and name --no-prompt, exactly as the old non-TTY branch did.
//
// Asking before knowing whether a reply can arrive is the deliberate trade this design
// makes. The question now appears in an unattended run's stderr above its refusal,
// which is strictly more informative than the silence it replaced.
func TestUnansweredPromptStillRefuses(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	w.Close()
	defer r.Close()
	old := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = old }()

	a := &App{}
	var ok bool
	stderrText := captureStderr(t, func() { ok = confirmAction(a, "Overwrite", "overwrite", "out.cli") })
	if ok {
		t.Error("confirmAction with nobody to answer = true, want false")
	}
	if !strings.Contains(stderrText, "Overwrite out.cli?") {
		t.Errorf("the question was not asked before refusing: %q", stderrText)
	}
	if !strings.Contains(stderrText, "--no-prompt") {
		t.Errorf("the refusal did not name --no-prompt: %q", stderrText)
	}

	// interactive/promptSource have always tolerated a nil App; keep it that way.
	if promptSource(nil) == nil {
		t.Error("promptSource(nil) = nil, want a reader over os.Stdin")
	}
}

// TestPromptKeepsBufferedAnswersForLaterQuestions covers a command that asks TWO
// questions in one run, which `broker remove --delete-data` does.
//
// promptLine used to build a fresh bufio.Reader per call and discard it, losing
// every byte that reader had buffered past the newline it returned. A pipe
// delivers `y\nyes\n` in ONE read, so the first question consumed all of it,
// answered "y", and threw "yes\n" away -- and the second question saw EOF,
// reported itself unanswered, and refused. The operator had supplied both answers
// and was told they had supplied neither. One reader per run, memoised on the App,
// is the fix; this asserts the second answer survives the first question.
func TestPromptKeepsBufferedAnswersForLaterQuestions(t *testing.T) {
	// deleteLayer is what RAISES confirmLayer's question at all (see confirmLayer).
	a := &App{
		Interactive: func() bool { return true },
		PromptIn:    strings.NewReader("y\nyes\n"),
		deleteLayer: true,
	}

	var first, second bool
	_ = captureStderr(t, func() {
		first = confirmAction(a, "Delete", "delete", "broker x")
		second = confirmLayer(a, layerData)
	})
	if !first {
		t.Error(`first question answered "y" = false, want true`)
	}
	if !second {
		t.Error(`second question answered "yes" = false, want true -- the answer was buffered away by the first`)
	}
}

// TestPromptSeamsBeatStdin pins the precedence the ~200 wiring tests depend on: an App
// carrying explicit Interactive/PromptIn seams must never touch os.Stdin, or a test
// would inherit whatever the suite happens to be launched from.
func TestPromptSeamsBeatStdin(t *testing.T) {
	a := &App{Interactive: func() bool { return true }, PromptIn: strings.NewReader("y\n")}
	if src := promptSource(a); src == os.Stdin {
		t.Error("promptSource ignored App.PromptIn")
	}
	var ok bool
	_ = captureStderr(t, func() { ok = confirmAction(a, "Overwrite", "overwrite", "out.cli") })
	if !ok {
		t.Error("confirmAction through the App seams = false, want true")
	}
}

// TestCtrManagerConfirmWiring covers confirmRestart's wiring into ctrManager: a
// regression here (ctrManager forgetting to set m.Confirm, or confirmRestart
// losing its interactive guard) would silently let a non-interactive deploy bounce
// a live broker container unattended. It drives the closure through the App's
// Interactive seam rather than the ambient os.Stdin, so the outcome is
// deterministic on any host.
func TestCtrManagerConfirmWiring(t *testing.T) {
	a := &App{Interactive: func() bool { return false }}
	m := ctrManager(a)
	if m.Confirm == nil {
		t.Fatal("ctrManager did not wire Manager.Confirm")
	}
	if m.Confirm("restart now?") {
		t.Error("Confirm (non-interactive) = true, want false")
	}
}

// TestK8sLoginOutcomes covers k8sLogin's two real, distinct SEMP outcomes -- a
// transport failure (error propagates) vs. a successful login (nil) -- which
// engine.Echo's fixed (nil, nil) OutputInput return can never produce; every
// existing test only reaches the "no HTTP response" failure-but-not-error branch.
func TestK8sLoginOutcomes(t *testing.T) {
	cfg, err := config.Load(writeStandaloneEnv(t), config.K8s)
	if err != nil {
		t.Fatalf("load standalone env: %v", err)
	}
	t.Run("SEMP login succeeds", func(t *testing.T) {
		rr := &opRunner{output: func(opCall) []byte { return []byte("HTTP/1.1 200 OK\r\n\r\n") }}
		a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr}
		var loginErr error
		captureStdout(t, func() { loginErr = k8sLogin(a, k8sOps(a), config.Primary) })
		if loginErr != nil {
			t.Errorf("k8sLogin (canned 200 OK) err = %v, want nil", loginErr)
		}
	})
	t.Run("SEMP transport failure propagates", func(t *testing.T) {
		rr := &opRunner{fail: opFailOn("curl")}
		a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr}
		var loginErr error
		captureStdout(t, func() { loginErr = k8sLogin(a, k8sOps(a), config.Primary) })
		if loginErr == nil || !strings.Contains(loginErr.Error(), "SEMP request failed") {
			t.Errorf("k8sLogin (transport failure) err = %v, want it to contain 'SEMP request failed'", loginErr)
		}
	})
}

// TestCtrLoginOutcomes is ctrLogin's half of TestK8sLoginOutcomes.
func TestCtrLoginOutcomes(t *testing.T) {
	cfg, err := config.Load(writeCtrStandaloneEnv(t), config.Docker)
	if err != nil {
		t.Fatalf("load container standalone env: %v", err)
	}
	t.Run("SEMP login succeeds", func(t *testing.T) {
		rr := &opRunner{output: func(opCall) []byte { return []byte("HTTP/1.1 200 OK\r\n\r\n") }}
		a := &App{Cfg: cfg, Platform: config.Docker, Runner: rr}
		var loginErr error
		captureStdout(t, func() { loginErr = ctrLogin(a, ctrOps(a)) })
		if loginErr != nil {
			t.Errorf("ctrLogin (canned 200 OK) err = %v, want nil", loginErr)
		}
	})
	t.Run("SEMP transport failure propagates", func(t *testing.T) {
		rr := &opRunner{fail: opFailOn("curl")}
		a := &App{Cfg: cfg, Platform: config.Docker, Runner: rr}
		var loginErr error
		captureStdout(t, func() { loginErr = ctrLogin(a, ctrOps(a)) })
		if loginErr == nil || !strings.Contains(loginErr.Error(), "SEMP request failed") {
			t.Errorf("ctrLogin (transport failure) err = %v, want it to contain 'SEMP request failed'", loginErr)
		}
	})
}

// TestSecretsNeverEchoed is the S3 smoke check: a secret-bearing command over the
// echo seam must show its stdin as a byte count, never the secret value. `check
// semp-login` puts the admin credential on curl's stdin; the login itself fails
// under Echo (no broker), which is fine -- the assertion is purely that the
// password does not reach stdout.
func TestSecretsNeverEchoed(t *testing.T) {
	path := writeStandaloneEnv(t)
	out, _ := runStandalone(t, path, "broker", "perform", "semp-login-check", "--platform", "kubernetes")
	if !strings.Contains(out, "bytes on stdin") {
		t.Errorf("check semp-login stdout = %q, want a 'bytes on stdin' redaction", out)
	}
	if strings.Contains(out, smokeAdminPass) {
		t.Error("check semp-login leaked the admin password to stdout")
	}
}

// TestK8sGenSecretsMissingCertFile covers emitK8sArtifact's GenSecrets error
// return: kubernetes.tlsServerSecret is configured but the referenced cert file
// cannot be read -- a real broken-reference case, otherwise caught only at real
// deploy time.
func TestK8sGenSecretsMissingCertFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "badcert.yaml")
	content := "redundancy:\n  enabled: false\n" +
		"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n" +
		"semp:\n  adminPass: " + smokeAdminPass + "\n" +
		"tls:\n" +
		"  cert: " + filepath.ToSlash(filepath.Join(dir, "missing.crt")) + "\n" +
		"  certKey: " + filepath.ToSlash(filepath.Join(dir, "missing.key")) + "\n" +
		"kubernetes:\n" +
		"  name: dev-broker\n" +
		"  namespace: solace\n" +
		"  tlsServerSecret: solace-tls-secret\n" +
		"  adminSecret: solace-admin-secret\n" +
		"  updateStrategy: automatedRolling\n" +
		"  storage:\n    class: standard\n    msgNodeSize: 30Gi\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}
	_, err := runRoot(t, []string{"broker", "generate", "--env", path, "--platform", "kubernetes"})
	if err == nil || !strings.Contains(err.Error(), "read tls.cert") {
		t.Fatalf("k8s broker generate (missing cert file) err = %v, want it to wrap the tls.cert read failure", err)
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// fakeBinaryOnPath writes an executable stub named base into a fresh directory, puts
// that directory at the front of PATH for this test, and returns the bare name plus the
// absolute path engine.Resolve will find. It makes the announcement assertions hermetic:
// no test host needs kubectl or docker installed, and the expected path is exact rather
// than "something absolute".
func fakeBinaryOnPath(t *testing.T, base string) (name, path string) {
	t.Helper()
	dir := t.TempDir()
	file, body, mode := base, "#!/bin/sh\nexit 0\n", os.FileMode(0o755)
	if runtime.GOOS == "windows" {
		// LookPath resolves a bare name through PATHEXT, which includes .BAT -- the
		// same assumption engine's TestResolveRefusesCurrentDirectory makes.
		file, body, mode = base+".bat", "@echo off\r\nexit /b 0\r\n", 0o666
	}
	path = filepath.Join(dir, file)
	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return base, path
}

// allowRuntime approves a fake binary for this config the only way the allowlist can be
// widened -- the operator's own --allow-command, here through its one entry point.
func allowRuntime(t *testing.T, cfg *config.Config, names ...string) {
	t.Helper()
	if err := cfg.AllowCommands(names); err != nil {
		t.Fatalf("AllowCommands(%v): %v", names, err)
	}
}

// TestAnnounceCommandsNamesResolvedBinaries covers the preamble that replaced the
// per-call `exec:` line: the binaries an env file chose are resolved and named ONCE, up
// front, so the location the allowlist cannot guarantee is still visible without
// repeating itself between report lines on every command.
func TestAnnounceCommandsNamesResolvedBinaries(t *testing.T) {
	t.Run("k8s names the cluster CLI", func(t *testing.T) {
		name, path := fakeBinaryOnPath(t, "solace-fake-kube")
		cfg, err := config.Load(writeStandaloneEnv(t), config.K8s)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		cfg.K8s.Runtime = config.Command{name}
		allowRuntime(t, cfg, name)
		a := &App{Cfg: cfg, Platform: config.K8s}
		got := captureStderr(t, a.announceCommands)
		if want := "==> using " + name + ": " + path + "\n"; got != want {
			t.Errorf("announcement =\n%q\nwant\n%q", got, want)
		}
	})
	// docker.compose defaults to the runtime's own `compose` subcommand, so argv[0] is
	// one binary announced once -- naming it twice would read as two installs.
	t.Run("docker names one binary when compose is derived", func(t *testing.T) {
		name, path := fakeBinaryOnPath(t, "solace-fake-docker")
		cfg, err := config.Load(writeCtrStandaloneEnv(t), config.Docker)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		cfg.Docker.Runtime = config.Command{name}
		cfg.Docker.Compose = config.Command{name, "compose"}
		allowRuntime(t, cfg, name)
		a := &App{Cfg: cfg, Platform: config.Docker}
		got := captureStderr(t, a.announceCommands)
		if want := "==> using " + name + ": " + path + "\n"; got != want {
			t.Errorf("announcement =\n%q\nwant\n%q", got, want)
		}
	})
	// A host carrying only the standalone compose v1 binary runs two different
	// binaries, and both locations matter.
	t.Run("docker names a standalone compose binary too", func(t *testing.T) {
		name, path := fakeBinaryOnPath(t, "solace-fake-docker")
		compose, composePath := fakeBinaryOnPath(t, "solace-fake-compose")
		cfg, err := config.Load(writeCtrStandaloneEnv(t), config.Docker)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		cfg.Docker.Runtime = config.Command{name}
		cfg.Docker.Compose = config.Command{compose}
		allowRuntime(t, cfg, name, compose)
		a := &App{Cfg: cfg, Platform: config.Docker}
		got := captureStderr(t, a.announceCommands)
		want := "==> using " + name + ": " + path + "\n" +
			"==> using " + compose + ": " + composePath + "\n"
		if got != want {
			t.Errorf("announcement =\n%q\nwant\n%q", got, want)
		}
	})
	// A name that resolves nowhere is skipped in silence: this is a report, and turning
	// it into a failure would break a command whose runner the operator never reaches.
	// The first real execution still fails with engine.Resolve's own message.
	t.Run("an unresolvable binary is skipped silently", func(t *testing.T) {
		cfg, err := config.Load(writeStandaloneEnv(t), config.K8s)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		cfg.K8s.Runtime = config.Command{"solace-absent-kube"}
		allowRuntime(t, cfg, "solace-absent-kube")
		a := &App{Cfg: cfg, Platform: config.K8s}
		if got := captureStderr(t, a.announceCommands); got != "" {
			t.Errorf("announcement = %q, want nothing for a binary that resolves nowhere", got)
		}
	})
}

// TestBinaryAnnouncementWiring drives the preamble through the real command tree. The
// stub is named `kubectl` so it is the schema default and needs no --allow-command --
// which also keeps the negative cases honest, since --allow-command is itself refused on
// a command that renders without executing.
//
// The negative half matters as much as the positive: `generate` is documented to need no
// kubectl/docker/podman installed at all, since it never executes anything, and
// resolving one there would contradict the promise.
func TestBinaryAnnouncementWiring(t *testing.T) {
	const marker = "==> using "

	t.Run("a real run announces before it works", func(t *testing.T) {
		_, path := fakeBinaryOnPath(t, "kubectl")
		env := writeStandaloneEnv(t)
		// The stub answers every call with empty stdout, so `check deploy` fails at
		// the storage-class assertion -- after the announcement, which is what is
		// asserted.
		got := captureStderr(t, func() {
			_, _ = runRoot(t, []string{"validate", "--env", env, "--platform", "kubernetes"})
		})
		if want := marker + "kubectl: " + path; !strings.Contains(got, want) {
			t.Errorf("stderr = %q, want it to carry %q", got, want)
		}
	})
	t.Run("nothing is announced where nothing executes", func(t *testing.T) {
		fakeBinaryOnPath(t, "kubectl")
		env := writeStandaloneEnv(t)
		for _, args := range [][]string{
			{"broker", "generate"},
			{"operator", "generate"},
		} {
			full := append(append([]string{}, args...), "--env", env, "--platform", "kubernetes")
			got := captureStderr(t, func() { _, _ = runRoot(t, full) })
			if strings.Contains(got, marker) {
				t.Errorf("%v announced a binary it never runs: %q", args, got)
			}
		}
	})
}

// TestVerboseFlagTracesEveryCommand: -v is the opt-in per-call trail that replaced the
// unconditional `exec:` line. It answers "what exactly did this run issue?", which the
// once-per-binary preamble deliberately does not.
func TestVerboseFlagTracesEveryCommand(t *testing.T) {
	fakeBinaryOnPath(t, "kubectl")
	env := writeStandaloneEnv(t)

	traced := captureStderr(t, func() {
		_, _ = runRoot(t, []string{"validate", "-v", "--env", env, "--platform", "kubernetes"})
	})
	if !strings.Contains(traced, "==> exec: ") || !strings.Contains(traced, "version -o json") {
		t.Errorf("-v stderr = %q, want a `==> exec: <path> version -o json` line", traced)
	}
	// The default run stays quiet per call: the preamble already named the binary.
	quiet := captureStderr(t, func() {
		_, _ = runRoot(t, []string{"validate", "--env", env, "--platform", "kubernetes"})
	})
	if strings.Contains(quiet, "==> exec: ") {
		t.Errorf("a run without -v traced its commands: %q", quiet)
	}
	// And it is a no-op with the test-only echo seam installed too, where Echo
	// already prints every command it would run -- passing both must still work
	// rather than fight.
	out, err := runRootWith(t, []string{"validate", "-v", "--env", env, "--platform", "kubernetes"}, echoRunner)
	if err != nil {
		t.Fatalf("echoRunner + -v: %v", err)
	}
	if !strings.Contains(out, "+ kubectl version -o json") {
		t.Errorf("echoRunner + -v stdout = %q, want the echoed command", out)
	}
}

// TestConfigureServerCertsRefusesASecretItDoesNotOwn pins the three-way routing. Keying
// the Secret branch on ManagesTLSSecret sent a bring-your-own deployment down the pod-exec
// CLI path, where it died on a generic "tls.cert and tls.certKey must both be set" --
// pointing the operator at fields that, if they set them, would make this tool overwrite a
// cert-manager Secret. The named Secret is what selects the Secret route; whether we can
// rebuild it is UpdateServerCertSecret's question.
// TestRemoveServerCertsOverTheCLI covers the direction that was a loud placeholder until
// `no ssl server-certificate` was confirmed. Three properties, all of them the reason the
// command was held back:
//
//   - it runs the confirmed form, not something adjacent to it;
//   - it ASKS first, because it takes TLS down on a live broker, and --no-prompt is the
//     one way to answer that unattended;
//   - declining touches the broker in no way at all.
func TestRemoveServerCertsOverTheCLI(t *testing.T) {
	path := writeCtrStandaloneEnv(t)

	t.Run("--no-prompt runs the confirmed CLI form", func(t *testing.T) {
		out, err := runCtr(t, path, "broker", "configure", "server-certs", "--remove",
			"--no-prompt", "--platform", "docker")
		if err != nil {
			t.Fatalf("configure server-certs --remove err = %v, want nil", err)
		}
		if !strings.Contains(out, "remove-server-certs") {
			t.Errorf("output does not show the removal script being run:\n%s", out)
		}
	})

	t.Run("declining changes nothing", func(t *testing.T) {
		out, err := runRootWith(t,
			[]string{"broker", "configure", "server-certs", "--remove", "--env", path, "--platform", "docker"},
			func(a *App) {
				a.Interactive = func() bool { return true }
				a.PromptIn = strings.NewReader("n\n")
				echoRunner(a)
			})
		if err != nil {
			t.Fatalf("declining must not be an error: %v", err)
		}
		if strings.Contains(out, "remove-server-certs") {
			t.Errorf("a declined removal must issue nothing:\n%s", out)
		}
	})

	t.Run("no terminal and no --no-prompt keeps the certificate", func(t *testing.T) {
		out, err := runRootWith(t,
			[]string{"broker", "configure", "server-certs", "--remove", "--env", path, "--platform", "docker"},
			func(a *App) {
				a.Interactive = func() bool { return false }
				echoRunner(a)
			})
		if err != nil {
			t.Fatalf("an unanswerable prompt keeps the certificate rather than failing: %v", err)
		}
		if strings.Contains(out, "remove-server-certs") {
			t.Error("a removal nobody could confirm must not run")
		}
	})
}

// TestRemoveProductKeysOverTheCLI covers the last direction that was a placeholder.
// `no product-key <key>` can leave the broker UNLICENSED -- an outage whose cause points
// nowhere near the command that caused it -- so the same gate the server-certificate
// removal takes applies here, and the same three properties are asserted.
func TestRemoveProductKeysOverTheCLI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.yaml")
	content := "redundancy:\n  enabled: false\n" +
		"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n" +
		"semp:\n  adminPass: " + smokeAdminPass + "\n" +
		"broker:\n  productKeys:\n    - KEY-ONE\n" +
		"docker: {}\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	t.Run("--no-prompt revokes every configured key", func(t *testing.T) {
		out, err := runCtr(t, path, "broker", "configure", "product-keys", "--remove",
			"--no-prompt", "--platform", "docker")
		if err != nil {
			t.Fatalf("configure product-keys --remove err = %v, want nil", err)
		}
		if !strings.Contains(out, "remove-product-keys") {
			t.Errorf("output does not show the revocation script being run:\n%s", out)
		}
	})

	t.Run("declining changes nothing", func(t *testing.T) {
		out, err := runRootWith(t,
			[]string{"broker", "configure", "product-keys", "--remove", "--env", path, "--platform", "docker"},
			func(a *App) {
				a.Interactive = func() bool { return true }
				a.PromptIn = strings.NewReader("n\n")
				echoRunner(a)
			})
		if err != nil {
			t.Fatalf("declining must not be an error: %v", err)
		}
		if strings.Contains(out, "remove-product-keys") {
			t.Errorf("a declined revocation must issue nothing:\n%s", out)
		}
	})

	t.Run("nothing configured is refused rather than reported as done", func(t *testing.T) {
		bare := writeCtrStandaloneEnv(t) // no broker.productKeys at all
		_, err := runCtr(t, bare, "broker", "configure", "product-keys", "--remove",
			"--no-prompt", "--platform", "docker")
		if err == nil || !strings.Contains(err.Error(), "no product keys configured") {
			t.Errorf("err = %v, want the same empty-list refusal the apply path gives", err)
		}
	})
}

// TestRemoveDomainCertsOverTheCLI covers domain-certs' newly confirmed --remove
// direction: it deletes trust anchors from a live broker, so it now takes the same
// gate server-certs/product-keys already did. Three properties, the same ones those
// two pin: the confirmed form actually runs; declining changes nothing; and a
// non-interactive run with no --no-prompt keeps the certificates rather than
// removing them silently.
func TestRemoveDomainCertsOverTheCLI(t *testing.T) {
	path := filepath.Join(t.TempDir(), "domaincerts-ctr.yaml")
	content := "redundancy:\n  enabled: false\n" +
		"  primary:\n    name: pri-host\n" +
		"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n" +
		"semp:\n  adminPass: " + smokeAdminPass + "\n" +
		"docker: {}\n" +
		"broker:\n  domainCerts:\n    files:\n      myca: myca.pem\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	t.Run("--no-prompt removes the configured CAs", func(t *testing.T) {
		out, err := runCtr(t, path, "broker", "configure", "domain-certs", "--remove",
			"--no-prompt", "--platform", "docker")
		if err != nil {
			t.Fatalf("configure domain-certs --remove err = %v, want nil", err)
		}
		if !strings.Contains(out, "remove-domain-certs") {
			t.Errorf("output does not show the removal script being run:\n%s", out)
		}
	})

	t.Run("declining changes nothing", func(t *testing.T) {
		out, err := runRootWith(t,
			[]string{"broker", "configure", "domain-certs", "--remove", "--env", path, "--platform", "docker"},
			func(a *App) {
				a.Interactive = func() bool { return true }
				a.PromptIn = strings.NewReader("n\n")
				echoRunner(a)
			})
		if err != nil {
			t.Fatalf("declining must not be an error: %v", err)
		}
		if strings.Contains(out, "remove-domain-certs") {
			t.Errorf("a declined removal must issue nothing:\n%s", out)
		}
	})

	t.Run("no terminal and no --no-prompt keeps the certificates", func(t *testing.T) {
		out, err := runRootWith(t,
			[]string{"broker", "configure", "domain-certs", "--remove", "--env", path, "--platform", "docker"},
			func(a *App) {
				a.Interactive = func() bool { return false }
				echoRunner(a)
			})
		if err != nil {
			t.Fatalf("an unanswerable prompt keeps the certificates rather than failing: %v", err)
		}
		if strings.Contains(out, "remove-domain-certs") {
			t.Error("a removal nobody could confirm must not run")
		}
	})

	t.Run("--apply still runs with no prompt at all", func(t *testing.T) {
		out, err := runCtr(t, path, "broker", "configure", "domain-certs", "--platform", "docker")
		if err != nil {
			t.Fatalf("configure domain-certs err = %v, want nil", err)
		}
		if !strings.Contains(out, "myca.pem") {
			t.Errorf("the apply direction must still run unconfirmed:\n%s", out)
		}
	})
}

// TestDisableDefaultVPNConfirmGate covers default-vpn's newly confirmed disable
// direction: it stops every client connection using the VPN, so it now takes the
// same gate the other disruptive `broker configure` leaves do.
func TestDisableDefaultVPNConfirmGate(t *testing.T) {
	path := writeCtrStandaloneEnv(t)

	t.Run("--no-prompt shuts the VPN down", func(t *testing.T) {
		out, err := runCtr(t, path, "broker", "configure", "default-vpn", "--no-prompt", "--platform", "docker")
		if err != nil {
			t.Fatalf("configure default-vpn err = %v, want nil", err)
		}
		if !strings.Contains(out, "disable-default-vpn") {
			t.Errorf("output does not show the disable script being run:\n%s", out)
		}
	})

	t.Run("declining changes nothing", func(t *testing.T) {
		out, err := runRootWith(t,
			[]string{"broker", "configure", "default-vpn", "--env", path, "--platform", "docker"},
			func(a *App) {
				a.Interactive = func() bool { return true }
				a.PromptIn = strings.NewReader("n\n")
				echoRunner(a)
			})
		if err != nil {
			t.Fatalf("declining must not be an error: %v", err)
		}
		if strings.Contains(out, "disable-default-vpn") {
			t.Errorf("a declined shutdown must issue nothing:\n%s", out)
		}
	})

	t.Run("no terminal and no --no-prompt keeps the VPN up", func(t *testing.T) {
		out, err := runRootWith(t,
			[]string{"broker", "configure", "default-vpn", "--env", path, "--platform", "docker"},
			func(a *App) {
				a.Interactive = func() bool { return false }
				echoRunner(a)
			})
		if err != nil {
			t.Fatalf("an unanswerable prompt keeps the VPN up rather than failing: %v", err)
		}
		if strings.Contains(out, "disable-default-vpn") {
			t.Error("a shutdown nobody could confirm must not run")
		}
	})

	t.Run("--enable still runs with no prompt at all", func(t *testing.T) {
		out, err := runCtr(t, path, "broker", "configure", "default-vpn", "--enable", "--platform", "docker")
		if err != nil {
			t.Fatalf("configure default-vpn --enable err = %v, want nil", err)
		}
		if !strings.Contains(out, "enable-default-vpn") {
			t.Errorf("the enable direction must still run unconfirmed:\n%s", out)
		}
	})
}

// TestDisableDefaultUsersConfirmGate covers default-users' newly confirmed disable
// direction: it blocks any client still relying on the default username, so it now
// takes the same gate its default-vpn sibling does. DisableDefaultUsers reads the
// broker's VPN list before it can build the disable script, and the echo seam
// answers that read with nothing, so the script itself never runs either way --
// the property that distinguishes a confirmed run from a declined one is the
// upload-and-run of that "show-vpn" probe, which only a confirmed run reaches.
func TestDisableDefaultUsersConfirmGate(t *testing.T) {
	path := writeCtrStandaloneEnv(t)

	t.Run("--no-prompt reaches the broker", func(t *testing.T) {
		out, err := runCtr(t, path, "broker", "configure", "default-users", "--no-prompt", "--platform", "docker")
		if err != nil {
			t.Fatalf("configure default-users err = %v, want nil", err)
		}
		if !strings.Contains(out, "show-vpn") {
			t.Errorf("output does not show the VPN list being read:\n%s", out)
		}
	})

	t.Run("declining changes nothing", func(t *testing.T) {
		out, err := runRootWith(t,
			[]string{"broker", "configure", "default-users", "--env", path, "--platform", "docker"},
			func(a *App) {
				a.Interactive = func() bool { return true }
				a.PromptIn = strings.NewReader("n\n")
				echoRunner(a)
			})
		if err != nil {
			t.Fatalf("declining must not be an error: %v", err)
		}
		if strings.Contains(out, "show-vpn") {
			t.Errorf("a declined shutdown must issue nothing:\n%s", out)
		}
	})

	t.Run("no terminal and no --no-prompt keeps the users up", func(t *testing.T) {
		out, err := runRootWith(t,
			[]string{"broker", "configure", "default-users", "--env", path, "--platform", "docker"},
			func(a *App) {
				a.Interactive = func() bool { return false }
				echoRunner(a)
			})
		if err != nil {
			t.Fatalf("an unanswerable prompt keeps the users up rather than failing: %v", err)
		}
		if strings.Contains(out, "show-vpn") {
			t.Error("a shutdown nobody could confirm must not run")
		}
	})

	t.Run("--enable still runs with no prompt at all", func(t *testing.T) {
		out, err := runCtr(t, path, "broker", "configure", "default-users", "--enable", "--platform", "docker")
		if err != nil {
			t.Fatalf("configure default-users --enable err = %v, want nil", err)
		}
		if !strings.Contains(out, "show-vpn") {
			t.Errorf("the enable direction must still run unconfirmed:\n%s", out)
		}
	})
}

// TestRemoveServerCertsRefusedOnASecretManagedDeployment: with kubernetes.tlsServerSecret
// set, the operator mounts the certificate and would reconcile it straight back -- so a
// CLI removal would report success over a broker that still presents it. The refusal has
// to name the Secret and say what to do instead, since "not supported here" would leave
// an operator with no next move.
func TestRemoveServerCertsRefusedOnASecretManagedDeployment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret-managed.yaml")
	content := "redundancy:\n  enabled: false\n" +
		"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n" +
		"semp:\n  adminPass: " + smokeAdminPass + "\n" +
		"kubernetes:\n" +
		"  name: dev-broker\n" +
		"  namespace: solace\n" +
		"  adminSecret: solace-admin-secret\n" +
		"  tlsServerSecret: byo-tls-secret\n" +
		"  updateStrategy: automatedRolling\n" +
		"  storage:\n    class: standard\n    msgNodeSize: 30Gi\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	_, err := runRootWith(t, []string{"broker", "configure", "server-certs", "--remove",
		"--no-prompt", "--env", path, "--platform", "kubernetes"}, echoRunner)
	if err == nil {
		t.Fatal("a CLI removal on a Secret-managed deployment must be refused, not reported as done")
	}
	for _, want := range []string{"byo-tls-secret", "tlsServerSecret", "reconcile"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %q, want it to mention %q", err, want)
		}
	}
}

func TestConfigureServerCertsRefusesASecretItDoesNotOwn(t *testing.T) {
	path := filepath.Join(t.TempDir(), "byo.yaml")
	content := "redundancy:\n  enabled: false\n" +
		"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n" +
		"semp:\n  adminPass: " + smokeAdminPass + "\n" +
		"kubernetes:\n" +
		"  name: dev-broker\n" +
		"  namespace: solace\n" +
		"  adminSecret: solace-admin-secret\n" +
		"  tlsServerSecret: byo-tls-secret\n" + // named, but no tls.cert/certKey
		"  updateStrategy: automatedRolling\n" +
		"  storage:\n    class: standard\n    msgNodeSize: 30Gi\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}

	_, err := runRoot(t, []string{"broker", "configure", "server-certs", "--env", path, "--platform", "kubernetes"})
	if err == nil {
		t.Fatal("rotating a Secret this env file supplies no files for must fail loud")
	}
	// The tailored message, not the CLI path's generic bundle error.
	for _, want := range []string{"byo-tls-secret", "cert-manager"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %q, want it to mention %q", err, want)
		}
	}
	if strings.Contains(err.Error(), "must both be set to build a server certificate") {
		t.Errorf("that is the CLI path's error -- a Secret-backed deployment must not reach it: %v", err)
	}
}

// TestGenerateWritesTheArtifactItself covers the flag's whole reason for existing.
//
// `broker generate > x.yaml` is not portable: Windows PowerShell 5.1 re-encodes this
// tool's plain ASCII as UTF-16LE with a BOM, and kubectl then rejects the file with an
// error pointing at the YAML rather than at the shell. The corruption happens after this
// process exits, so nothing inside it can prevent it -- writing the file here is the only
// fix. The property that matters is therefore byte-for-byte identity between what -o
// writes and what stdout carries, on every platform.
func TestGenerateWritesTheArtifactItself(t *testing.T) {
	env := writeStandaloneEnv(t)
	path := filepath.Join(t.TempDir(), "solace.yaml")

	stdout, err := runRoot(t, []string{"broker", "generate", "--env", env, "--platform", "kubernetes"})
	if err != nil {
		t.Fatalf("broker generate err = %v, want nil", err)
	}
	_ = captureStderr(t, func() {
		if _, err := runRoot(t, []string{"broker", "generate", "--env", env, "--platform", "kubernetes", "-o", path}); err != nil {
			t.Fatalf("broker generate -o err = %v, want nil", err)
		}
	})

	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read the written artifact: %v", err)
	}
	if string(written) != stdout {
		t.Errorf("-o wrote something other than what stdout carries:\nfile:\n%s\nstdout:\n%s", written, stdout)
	}
	// Plain ASCII, no BOM: the whole point is that nothing re-encoded it.
	if len(written) > 2 && written[0] == 0xff && written[1] == 0xfe {
		t.Error("the written file carries a UTF-16 BOM")
	}
	for i, b := range written {
		if b == 0 {
			t.Fatalf("the written file has a NUL byte at offset %d -- it is not plain text", i)
		}
	}
}

// TestGenerateToFileConfirmsAnOverwrite: the artifact is a file you keep, so replacing one
// is confirmed like every other destructive step. A test run has no terminal, which is the
// non-interactive branch -- it must keep the file and say which flag would have proceeded.
func TestGenerateToFileConfirmsAnOverwrite(t *testing.T) {
	env := writeStandaloneEnv(t)
	path := filepath.Join(t.TempDir(), "solace.yaml")
	if err := os.WriteFile(path, []byte("keep me\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var err error
	_ = captureStderr(t, func() {
		_, err = runRoot(t, []string{"broker", "generate", "--env", env, "--platform", "kubernetes", "-o", path})
	})
	if err == nil {
		t.Fatal("an existing -o path must not be replaced without confirmation")
	}
	if body, _ := os.ReadFile(path); string(body) != "keep me\n" {
		t.Error("the refused write still replaced the file")
	}

	_ = captureStderr(t, func() {
		if _, err := runRoot(t, []string{"broker", "generate", "--env", env, "--platform", "kubernetes", "-o", path, "--no-prompt"}); err != nil {
			t.Fatalf("broker generate -o --no-prompt err = %v, want nil", err)
		}
	})
	if body, _ := os.ReadFile(path); string(body) == "keep me\n" {
		t.Error("--no-prompt did not replace the file")
	}
}

// TestGenerateNoPromptWithoutOutIsHarmless: --no-prompt answers the overwrite question, and
// with no -o there is no file and no question. Accepted rather than refused -- a script
// that passes it unconditionally is not making a mistake worth an exit 2, and nothing is
// silently skipped.
func TestGenerateNoPromptWithoutOutIsHarmless(t *testing.T) {
	env := writeStandaloneEnv(t)
	out, err := runRoot(t, []string{"broker", "generate", "--env", env, "--platform", "kubernetes", "--no-prompt"})
	if err != nil {
		t.Fatalf("broker generate --no-prompt err = %v, want nil", err)
	}
	if !strings.Contains(out, "kind: PubSubPlusEventBroker") {
		t.Errorf("the artifact must still reach stdout:\n%s", out)
	}
}

// TestMousetrapIsDisabled pins that cobra's Windows double-click guard stays off.
//
// It is a behavioural property first: cobra's pre-exec hook decides whether the binary
// was launched from Explorer and, if so, prints "this is a command line application" and
// declines to run. This tool is driven from terminals, scripts and CI, and refusing to
// work because of how the process was started is not a failure mode it should have.
//
// It is also the single largest cost in this package's own test suite. The hook walks the
// OS process table on every Execute() -- Getppid plus a snapshot scan for the parent --
// and a CPU profile measured it at 37% of the whole run, 10.6s of 13.8s, against 1.8s for
// the commands actually under test. Nothing else would notice it coming back: every test
// would still pass, just slower, which is exactly the kind of regression that survives.
func TestMousetrapIsDisabled(t *testing.T) {
	// Built rather than read from the package variable directly, because building the
	// root is what sets it -- a future refactor that moved the assignment out of
	// newRootCmd would leave the variable at its default in a real run.
	cobra.MousetrapHelpText = "put back by this test"
	_ = newRootCmd(&App{})
	if cobra.MousetrapHelpText != "" {
		t.Errorf("cobra.MousetrapHelpText = %q, want empty so the Explorer check is skipped: "+
			"it costs a process-table walk on every Execute and would refuse to run a "+
			"double-clicked binary that should simply work",
			cobra.MousetrapHelpText)
	}
}

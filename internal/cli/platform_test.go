package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"solace/internal/config"
)

// The platform used to be the first word of every command. It is now resolved --
// from --platform, or from the platform sections the env file declares -- so the
// resolution itself is the thing that has to be pinned: it decides what every
// later step in the run talks to, and getting it wrong silently would point a
// deploy at the wrong system.

// writePlatformEnv writes a minimal but VALID env carrying exactly the given
// platform sections, so a test can drive resolution without also tripping over
// schema validation. The kubernetes and container schemas need different
// mandatory fields, hence the two bodies.
func writePlatformEnv(t *testing.T, platforms ...config.Platform) string {
	t.Helper()
	body := "redundancy:\n  enabled: false\n"
	// A container env no longer NEEDS a primary name -- standalone falls back to the
	// host's -- but naming one here keeps these fixtures independent of the machine
	// running the suite. It rides in the same block as the switch.
	for _, p := range platforms {
		if p.IsContainer() {
			body += "  primary:\n    name: pri-host\n"
			break
		}
	}
	body += "image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n" +
		"semp:\n  adminPass: " + smokeAdminPass + "\n"
	for _, p := range platforms {
		switch p {
		case config.K8s:
			body += "kubernetes:\n  name: dev-broker\n  namespace: solace\n" +
				"  storage:\n    class: standard\n    msgNodeSize: 30Gi\n"
		case config.Podman:
			// A bare `podman: {}` no longer loads: baseDir is mandatory, because it
			// receives the server-certificate bundle and the schema will not pick a
			// location for a private key. Docker still needs no key of its own, which
			// is why only this platform gets a body.
			body += "podman:\n  baseDir: /opt/solace\n"
		default:
			body += string(p) + ": {}\n"
		}
	}
	path := filepath.Join(t.TempDir(), "platforms.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
	return path
}

// runPlatform runs `status broker` over the echo seam against path, which is the
// cheapest command that still goes all the way through resolution and config
// loading.
func runPlatform(t *testing.T, path string, configure func(*App), args ...string) (string, error) {
	t.Helper()
	full := append(append([]string{"broker", "status"}, args...), "--env", path)
	return runRootWith(t, full, func(a *App) {
		if configure != nil {
			configure(a)
		}
		echoRunner(a)
	})
}

// TestResolvesSinglePlatformSilently is the everyday case: an env file describes
// one deployment, so nothing needs to be said on the command line.
func TestResolvesSinglePlatformSilently(t *testing.T) {
	for _, p := range config.Platforms() {
		t.Run(string(p), func(t *testing.T) {
			out, err := runPlatform(t, writePlatformEnv(t, p), nil)
			if err != nil {
				t.Fatalf("status on a %s-only env: %v", p, err)
			}
			// The dry-run echo names the binary the platform drives, which is the
			// observable proof the right one was chosen.
			want := "kubectl"
			if p.IsContainer() {
				want = string(p)
			}
			if !strings.Contains(out, want) {
				t.Errorf("status on a %s-only env should drive %s, got:\n%s", p, want, out)
			}
		})
	}
}

// TestNoPlatformSectionIsRefused pins the marker requirement. A container env
// file needs no docker: keys at all -- every one of them defaults -- so without
// this rule such a file would be indistinguishable from a kubernetes one, and
// the tool would have to guess which system to deploy to.
func TestNoPlatformSectionIsRefused(t *testing.T) {
	path := filepath.Join(t.TempDir(), "no-platform.yaml")
	body := "redundancy:\n  enabled: false\nimage:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n" +
		"semp:\n  adminPass: " + smokeAdminPass + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
	_, err := runPlatform(t, path, nil)
	if err == nil {
		t.Fatal("an env file declaring no platform section should be refused")
	}
	for _, want := range []string{"kubernetes", "docker", "podman"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should name %q as a section to add", err, want)
		}
	}
}

// TestMultiPlatformNonInteractiveIsRefused: with no terminal to ask, guessing is
// the one thing that must not happen -- a scripted run that picked a platform on
// its own could deploy to the wrong one and look like it worked.
func TestMultiPlatformNonInteractiveIsRefused(t *testing.T) {
	path := writePlatformEnv(t, config.K8s, config.Docker)
	_, err := runPlatform(t, path, func(a *App) { a.Interactive = func() bool { return false } })
	if err == nil {
		t.Fatal("an ambiguous env file should be refused when nothing can be asked")
	}
	if !strings.Contains(err.Error(), "--platform") {
		t.Errorf("error %q should point at --platform", err)
	}
	for _, want := range []string{"kubernetes", "docker"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should name the declared platform %q", err, want)
		}
	}
}

// TestMultiPlatformPromptSelects drives the interactive branch through the App's
// prompt seams, the same way the destructive-confirmation tests do.
func TestMultiPlatformPromptSelects(t *testing.T) {
	path := writePlatformEnv(t, config.K8s, config.Docker)
	out, err := runPlatform(t, path, func(a *App) {
		a.Interactive = func() bool { return true }
		a.PromptIn = strings.NewReader("2\n")
	})
	if err != nil {
		t.Fatalf("selecting docker at the prompt: %v", err)
	}
	if !strings.Contains(out, "docker") {
		t.Errorf("selecting entry 2 should drive docker, got:\n%s", out)
	}
}

// TestMultiPlatformPromptRejectsBadAnswer: an unusable answer stops the run. It
// must not fall through to a default, for the same reason the non-interactive
// case refuses to guess.
func TestMultiPlatformPromptRejectsBadAnswer(t *testing.T) {
	path := writePlatformEnv(t, config.K8s, config.Docker)
	for _, answer := range []string{"\n", "9\n", "docker\n", "0\n"} {
		_, err := runPlatform(t, path, func(a *App) {
			a.Interactive = func() bool { return true }
			a.PromptIn = strings.NewReader(answer)
		})
		if err == nil {
			t.Errorf("answer %q should not select a platform", strings.TrimSpace(answer))
		}
	}
}

// TestPlatformFlagSilencesThePrompt: naming the platform is what makes an
// ambiguous file usable from a script.
func TestPlatformFlagSilencesThePrompt(t *testing.T) {
	path := writePlatformEnv(t, config.K8s, config.Docker)
	out, err := runPlatform(t, path, func(a *App) { a.Interactive = func() bool { return false } },
		"--platform", "docker")
	if err != nil {
		t.Fatalf("--platform docker on an ambiguous env: %v", err)
	}
	if !strings.Contains(out, "docker") {
		t.Errorf("--platform docker should drive docker, got:\n%s", out)
	}
}

// TestPlatformFlagAcceptsAbbreviations pins the short spellings end to end, not
// just in the parser: they have to survive the whole resolution path.
func TestPlatformFlagAcceptsAbbreviations(t *testing.T) {
	for abbrev, want := range map[string]config.Platform{"kube": config.K8s, "dk": config.Docker, "pm": config.Podman} {
		path := writePlatformEnv(t, config.K8s, config.Docker, config.Podman)
		out, err := runPlatform(t, path, func(a *App) { a.Interactive = func() bool { return false } },
			"--platform", abbrev)
		if err != nil {
			t.Errorf("--platform %s: %v", abbrev, err)
			continue
		}
		marker := "kubectl"
		if want.IsContainer() {
			marker = string(want)
		}
		if !strings.Contains(out, marker) {
			t.Errorf("--platform %s should drive %s, got:\n%s", abbrev, want, out)
		}
	}
}

// TestPlatformFlagRejectsUndeclaredSection: --platform names which of the file's
// deployments to drive, so naming one the file does not describe is a mistake
// worth stopping for -- the alternative is deploying from defaults the operator
// never wrote down.
func TestPlatformFlagRejectsUndeclaredSection(t *testing.T) {
	path := writePlatformEnv(t, config.K8s)
	_, err := runPlatform(t, path, nil, "--platform", "podman")
	if err == nil {
		t.Fatal("--platform podman against a kubernetes-only env should be refused")
	}
	if !strings.Contains(err.Error(), "podman") || !strings.Contains(err.Error(), "kubernetes") {
		t.Errorf("error %q should name both what was asked for and what the file declares", err)
	}
}

// TestPlatformFlagRejectsUnknownValue keeps the retired spellings out: `k8s`
// never came back, and `k8` was the abbreviation only until `kube` replaced it.
func TestPlatformFlagRejectsUnknownValue(t *testing.T) {
	path := writePlatformEnv(t, config.K8s)
	for _, bad := range []string{"k8s", "k8", "swarm"} {
		if _, err := runPlatform(t, path, nil, "--platform", bad); err == nil {
			t.Errorf("--platform %s should be refused", bad)
		}
	}
}

// TestUnsupportedCommandFailsLoud is the other half of the one-tree decision: the
// tree shows every command on every platform, so the refusal has to be the thing
// that tells an operator a command does not apply here -- and it has to name
// where it does apply, or the message is a dead end.
func TestUnsupportedCommandFailsLoud(t *testing.T) {
	for _, tc := range []struct {
		platform config.Platform
		args     []string
		applies  string
	}{
		{config.Docker, []string{"operator", "status"}, "kubernetes"},
		{config.Docker, []string{"operator", "restart"}, "kubernetes"},
		{config.Docker, []string{"operator", "deploy"}, "kubernetes"},
		{config.Docker, []string{"operator", "remove"}, "kubernetes"},
		{config.Docker, []string{"operator", "generate"}, "kubernetes"},
		// Every row is kubernetes-only, and that is now a property of the tree rather than a
		// gap in the table: the whole `operator` noun is the kubernetes-only surface, and no
		// COMMAND is container-only any more. `prepare host` was the last one, and its work
		// folded into `broker deploy`, which applies everywhere. What stayed container-only is
		// a flag (--restart) and a role reading (--pod), both covered by TestScopedFlagFailsLoud.
		//
		// `broker generate` is deliberately absent too: it is the one verb that means the same
		// thing on both families -- what `broker deploy` would apply -- so it is refused
		// nowhere. TestGenerateWired covers both renderings.
	} {
		t.Run(string(tc.platform)+" "+strings.Join(tc.args, " "), func(t *testing.T) {
			path := writePlatformEnv(t, config.K8s, config.Docker, config.Podman)
			args := append(append([]string{}, tc.args...), "--platform", string(tc.platform))
			_, err := runRootWith(t, append(args, "--env", path),
				func(a *App) { a.Interactive = func() bool { return false }; echoRunner(a) })
			if err == nil {
				t.Fatalf("%v should be refused on %s", tc.args, tc.platform)
			}
			if !strings.Contains(err.Error(), "not supported on "+string(tc.platform)) {
				t.Errorf("error %q should say it is not supported on %s", err, tc.platform)
			}
			if !strings.Contains(err.Error(), tc.applies) {
				t.Errorf("error %q should name %s, where it does apply", err, tc.applies)
			}
		})
	}
}

// TestScopedFlagFailsLoud: a flag that exists on every platform but means
// something on only one is refused where it means nothing, rather than accepted
// and ignored -- a --restart that did nothing would read as "already restarted".
func TestScopedFlagFailsLoud(t *testing.T) {
	for _, tc := range []struct {
		platform config.Platform
		args     []string
		flag     string
	}{
		{config.K8s, []string{"broker", "deploy", "--restart"}, "restart"},
		{config.Docker, []string{"broker", "cli", "--pod", "primary"}, "pod"},
		{config.Docker, []string{"broker", "logs", "--previous"}, "previous"},
		{config.Podman, []string{"broker", "logs", "--previous"}, "previous"},
		// --all is deliberately NOT here: it applies on every platform. On
		// Kubernetes it surveys the cluster, on a container host every Solace
		// container found by image -- the same question, asked of what that
		// platform has. TestStatusAllFindsBrokersByImage covers the container half.
	} {
		t.Run(string(tc.platform)+" --"+tc.flag, func(t *testing.T) {
			path := writePlatformEnv(t, config.K8s, config.Docker, config.Podman)
			args := append(append([]string{}, tc.args...), "--platform", string(tc.platform))
			_, err := runRootWith(t, append(args, "--env", path),
				func(a *App) { a.Interactive = func() bool { return false }; echoRunner(a) })
			if err == nil {
				t.Fatalf("--%s should be refused on %s", tc.flag, tc.platform)
			}
			if !strings.Contains(err.Error(), "--"+tc.flag) {
				t.Errorf("error %q should name --%s", err, tc.flag)
			}
		})
	}
}

// TestUnusableRoleFailsLoud: a role that was TYPED must never be silently
// dropped, whichever way it is unusable. Two different refusals meet here.
//
// The first six cases lost their positional entirely when the pod selector became
// --pod: `logs broker backup` is now a migration error (noRolePositional), because
// that spelling was the documented one until the change and cobra's own "unknown
// command" would hide the fact that it merely moved. On a container host the same
// message also says why no role applies there at all -- one broker per machine.
//
// The last four KEEP their positional, where the role is this host's node identity
// in a container deployment rather than a pod selector; on Kubernetes that is
// meaningless and is refused (rejectRole) rather than accepted and ignored.
//
// Both halves must name the role, which is all this asserts: `logs backup` quietly
// reading the local broker's logs is the exact mistake either refusal prevents.
func TestUnusableRoleFailsLoud(t *testing.T) {
	for _, tc := range []struct {
		platform config.Platform
		args     []string
	}{
		{config.Docker, []string{"broker", "logs", "backup"}},
		{config.Docker, []string{"broker", "cli", "backup"}},
		{config.Docker, []string{"broker", "shell", "monitor"}},
		{config.Docker, []string{"broker", "perform", "semp-login-check", "backup"}},
		{config.Docker, []string{"broker", "status", "backup"}},
		{config.Docker, []string{"broker", "restart", "backup"}},
		{config.K8s, []string{"broker", "deploy", "backup"}},
		{config.K8s, []string{"broker", "perform", "assert-leader", "backup"}},
		{config.K8s, []string{"broker", "perform", "redundancy-test", "backup"}},
	} {
		t.Run(string(tc.platform)+" "+strings.Join(tc.args, " "), func(t *testing.T) {
			path := writePlatformEnv(t, config.K8s, config.Docker, config.Podman)
			args := append(append([]string{}, tc.args...), "--platform", string(tc.platform))
			_, err := runRootWith(t, append(args, "--env", path),
				func(a *App) { a.Interactive = func() bool { return false }; echoRunner(a) })
			if err == nil {
				t.Fatalf("%v should be refused on %s", tc.args, tc.platform)
			}
			if !strings.Contains(err.Error(), "role") {
				t.Errorf("error %q should explain the role argument", err)
			}
		})
	}
}

// TestRolePositionalTeachesPodFlag is the other half of the migration, and the
// likelier one: an operator on KUBERNETES typing the spelling that worked until
// this change. `shell backup` was documented, so cobra's bare `unknown command
// "backup"` would be true and actively unhelpful -- it hides that the role merely
// moved to a flag. The error has to name the flag AND the role that was typed, so
// the corrected command can be read straight out of it.
//
// A word that is NOT a role keeps cobra's own wording: a typo is not a migration.
func TestRolePositionalTeachesPodFlag(t *testing.T) {
	for _, args := range [][]string{
		{"broker", "shell", "backup"},
		{"broker", "logs", "monitor"},
		{"broker", "cli", "primary"},
		{"broker", "status", "backup"},
		{"broker", "restart", "backup"},
		{"broker", "perform", "semp-login-check", "backup"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			path := writePlatformEnv(t, config.K8s, config.Docker, config.Podman)
			full := append(append([]string{}, args...), "--platform", "kubernetes", "--env", path)
			_, err := runRootWith(t, full, func(a *App) {
				a.Interactive = func() bool { return false }
				echoRunner(a)
			})
			if err == nil {
				t.Fatalf("%v should be refused: the role positional is gone", args)
			}
			role := args[len(args)-1]
			for _, want := range []string{"--pod", role} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q should name %q so the corrected command is readable from it", err, want)
				}
			}
		})
	}
	t.Run("a non-role word is still just unknown", func(t *testing.T) {
		path := writePlatformEnv(t, config.K8s, config.Docker, config.Podman)
		_, err := runRootWith(t, []string{"broker", "shell", "typo", "--platform", "kubernetes", "--env", path},
			func(a *App) { a.Interactive = func() bool { return false }; echoRunner(a) })
		if err == nil {
			t.Fatal("an unknown positional should still be refused")
		}
		if strings.Contains(err.Error(), "--pod") {
			t.Errorf("error %q should not offer the migration hint for a word that is not a role", err)
		}
	})
}

// TestPlatformIsAnnouncedInThePreamble: which system a command is about to talk
// to is now inferred rather than typed, so it has to be stated -- otherwise the
// one fact the operator no longer supplies is also the one they cannot see.
func TestPlatformIsAnnouncedInThePreamble(t *testing.T) {
	path := writePlatformEnv(t, config.Docker)
	stderr := captureStderr(t, func() {
		if _, err := runPlatform(t, path, nil); err != nil {
			t.Fatalf("status: %v", err)
		}
	})
	if !strings.Contains(stderr, "platform: docker") {
		t.Errorf("the preamble should announce the resolved platform, got:\n%s", stderr)
	}
}

// TestCompletionNeverReadsTheEnvFile is the invariant that decided where the
// pre-run hook lives. Cobra runs the NEAREST ancestor's PersistentPreRunE, and
// __complete is a child of root -- so a hook on root would parse an untrusted
// env file on every TAB press. Keeping the hook on each command instead is what
// prevents that, and this proves it: completion still works when the env file
// named on the command line does not even exist.
func TestCompletionNeverReadsTheEnvFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.yaml")
	for _, args := range [][]string{
		{"--env", missing, "broker", "logs", ""},
		{"--env", missing, "broker", "status", "--allow-command", ""},
		{"--env", missing, ""},
	} {
		if _, directive := runComplete(t, args...); directive == "" {
			t.Errorf("completion for %v should work without an env file", args)
		}
	}
}

// TestPlatformFlagIsOnRoot pins the flag's placement: it is inherited by every
// command, including convert, which is what lets one word mean one thing across
// the whole CLI.
func TestPlatformFlagIsOnRoot(t *testing.T) {
	root := newRootCmd(&App{})
	if root.PersistentFlags().Lookup("platform") == nil {
		t.Fatal("--platform should be a root persistent flag")
	}
	convert := findCmd(t, root, "convert")
	if convert.Flags().Lookup("platform") != nil {
		t.Error("convert should inherit --platform, not declare its own")
	}
	if convert.InheritedFlags().Lookup("platform") == nil {
		t.Error("convert should see the inherited --platform")
	}
}

// TestScopedCommandsSaySoInHelp: the tree is one static shape, so the help text
// is the only place an operator can learn that a command does not apply before
// running it.
func TestScopedCommandsSaySoInHelp(t *testing.T) {
	root := newRootCmd(&App{})
	for _, tc := range []struct {
		path []string
		want string
	}{
		{[]string{"operator", "status"}, "kubernetes only"},
		{[]string{"operator", "restart"}, "kubernetes only"},
		{[]string{"operator", "generate"}, "kubernetes only"},
	} {
		c := findCmd(t, root, tc.path...)
		if !strings.Contains(c.Short, tc.want) {
			t.Errorf("%q Short = %q, want it to mention %q", c.CommandPath(), c.Short, tc.want)
		}
	}
	// A command that applies everywhere carries no such tail.
	if c := findCmd(t, root, "broker", "status"); strings.Contains(c.Short, "only)") {
		t.Errorf("broker status applies everywhere, so its Short should carry no scope: %q", c.Short)
	}
}

// TestPlatformAnnotationsMatchDispatch guards the one place this design could rot
// silently: a command's applicability annotation is what help and the pre-run
// check read, while the ops map is what actually runs. If a command were tagged
// for a platform it has no implementation for, the refusal would come from the
// wrong place -- as an internal error at dispatch instead of an actionable one
// before anything loaded.
func TestPlatformAnnotationsMatchDispatch(t *testing.T) {
	root := newRootCmd(&App{})
	var walk func(*cobra.Command)
	walk = func(c *cobra.Command) {
		if v, ok := c.Annotations[platformAnnotation]; ok {
			if got := parsePlatformList(v); len(got) == 0 {
				t.Errorf("%q carries an unparseable platform annotation %q", c.CommandPath(), v)
			}
		}
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(root)
}

// TestPlatformOpsCoversEveryPlatform pins the builder against Platforms(). Every
// other consumer of an ops map already walks Platforms() -- supported() does, and
// so does the annotation it produces -- so platformOps is the one place a fourth
// platform would be dropped silently: with both halves non-nil it would still
// return three entries, supported() would omit the new name, onlyOn would tag the
// command for three platforms, and the new one would refuse every command that has
// a perfectly good container implementation. The refusal would look deliberate.
func TestPlatformOpsCoversEveryPlatform(t *testing.T) {
	stub := func(*App) error { return nil }
	m := platformOps(stub, stub)
	for _, p := range config.Platforms() {
		if m[p] == nil {
			t.Errorf("platformOps built with both halves non-nil has no implementation for %s: "+
				"the builder names platforms literally and this one was missed", p)
		}
	}
	if len(m) != len(config.Platforms()) {
		t.Errorf("platformOps returned %d entries, want %d (one per platform)", len(m), len(config.Platforms()))
	}
	// The nil halves must still be honoured, or "this operation does not exist
	// there" stops being expressible and every command becomes universal.
	if got := platformOps(stub, nil); len(got) != 1 || got[config.K8s] == nil {
		t.Errorf("platformOps(k8s, nil) = %v, want kubernetes only", supported(got))
	}
	if got := platformOps(nil, stub); got[config.K8s] != nil {
		t.Errorf("platformOps(nil, container) must not implement kubernetes, got %v", supported(got))
	}
}

// TestLogArgsBuildsOneSetForBothPlatforms pins the shared builder. The two argv
// builders must not disagree about what a flag means, which is why the tokens are
// produced once rather than assembled per platform.
func TestLogArgsBuildsOneSetForBothPlatforms(t *testing.T) {
	// Nothing set means nothing appended: `logs broker` prints a snapshot and exits,
	// which is the behaviour change from the old hard-coded follow.
	if got, err := logArgs(&App{}); err != nil || len(got) != 0 {
		t.Errorf("logArgs with no flags = %v, %v; want no tokens at all", got, err)
	}

	a := &App{follow: true, tail: "100", since: "90m", timestamps: true}
	got, err := logArgs(a)
	if err != nil {
		t.Fatalf("logArgs: %v", err)
	}
	want := []string{"-f", "--tail", "100", "--since", "1h30m0s", "--timestamps"}
	if len(got) != len(want) {
		t.Fatalf("logArgs = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("logArgs[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	// --since is canonicalised, not passed through: `90m` became `1h30m0s`, so what
	// reaches an argv is a string time.Duration produced.
	if got[4] == "90m" {
		t.Error("--since must be canonicalised through the duration parser, not forwarded verbatim")
	}

	// A bad duration fails here, before any argv is built.
	if _, err := logArgs(&App{since: "yesterday"}); err == nil {
		t.Error("an unparseable --since must be refused rather than reaching the engine")
	}

	// --previous is deliberately NOT in the shared set: it is kubernetes-only and
	// appended by the one handler that has it.
	for _, tok := range logArgs2(t, &App{previous: true}) {
		if tok == "-p" || tok == "--previous" {
			t.Error("logArgs must not carry --previous: it is kubernetes-only, so the handler appends it")
		}
	}
}

// logArgs2 is logArgs with the error asserted away, for the cases that are about the
// token list rather than the parsing.
func logArgs2(t *testing.T, a *App) []string {
	t.Helper()
	got, err := logArgs(a)
	if err != nil {
		t.Fatalf("logArgs: %v", err)
	}
	return got
}

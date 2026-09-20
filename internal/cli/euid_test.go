package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"solace/internal/engine"
)

// The euid guard cli.prepare runs (container.GuardPodmanEUID) is the one place the
// DECLARED podman.rootless is checked against the account running the command. It
// sits in prepare rather than in the Manager because fifteen container commands --
// every `broker configure *` and `broker perform *` -- reach the broker over
// `podman exec` and never build a Manager at all.
//
// Every case here drives a NON-Echo runner, because the property under test is that
// the runner is never reached; an Echo runner would pass the guard for its own
// reason and prove nothing. That is also why App.Geteuid exists: without it these
// would pass or fail on whichever account runs `go test`.

// writeRootlessPodmanEnv is a podman-only env file declaring rootless: true, so the
// guard has something to disagree with. Podman only, so no --platform is needed.
func writeRootlessPodmanEnv(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "rootless.yaml")
	content := "redundancy:\n  enabled: false\n" +
		"  primary:\n" +
		"    name: pri-host\n" +
		"image:\n" +
		"  repo: solace-pubsub-standard\n" +
		"  tag: \"10.10.1.128\"\n" +
		"semp:\n" +
		"  adminPass: " + smokeAdminPass + "\n" +
		"podman:\n" +
		"  rootless: true\n" +
		"  baseDir: /opt/solace\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write rootless podman env: %v", err)
	}
	return path
}

// runGuarded runs args with a capturing non-Echo runner at the given euid, and
// returns stdout, the error, and the runner so a caller can assert on what was NOT
// called.
func runGuarded(t *testing.T, euid int, args ...string) (string, error, *opRunner) {
	t.Helper()
	rr := &opRunner{}
	out, err := runRootWith(t, args, func(a *App) {
		a.Geteuid = func() int { return euid }
		a.NewRunner = func(*App) engine.EnvRunner { return rr }
	})
	return out, err, rr
}

// TestPodmanEUIDGuardRefusesBeforeAnyCommand is the point of the change: the refusal
// lands before the engine is touched, on the Manager path AND on the transport path
// that never builds one. `broker remove --delete-data` is the case that motivated
// it -- under sudo against a rootless deployment it used to stop nothing, remove
// nothing, and then delete the data directory's contents as root.
func TestPodmanEUIDGuardRefusesBeforeAnyCommand(t *testing.T) {
	rootless := writeRootlessPodmanEnv(t)
	sample := writeCtrStandaloneEnv(t)
	cases := []struct {
		name string
		euid int
		args []string
		want string
	}{
		{"rootless remove as root", 0,
			[]string{"broker", "remove", "--delete-data", "--no-prompt", "--env", rootless}, "running as root"},
		{"rootless status as root", 0, []string{"broker", "status", "--env", rootless}, "running as root"},
		{"rootless validate as root", 0, []string{"validate", "--env", rootless}, "running as root"},
		// The transport path: no Manager is built at all on this one.
		{"rootless configure as root", 0,
			[]string{"broker", "configure", "product-keys", "--env", rootless}, "running as root"},
		{"rootful status as non-root", 1000,
			[]string{"broker", "status", "--platform", "podman", "--env", sample}, "requires root"},
		{"rootful perform as non-root", 1000,
			[]string{"broker", "perform", "semp-login-check", "--platform", "podman", "--env", sample}, "requires root"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err, rr := runGuarded(t, tc.euid, tc.args...)
			if err == nil {
				t.Fatalf("a mismatched euid must refuse; stdout:\n%s", out)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error should say %q, got: %v", tc.want, err)
			}
			// Exit 2: a different invocation is what fixes it, and retrying the
			// same one cannot -- prepare's own rule (platform.go).
			if code := ExitCode(err); code != exitUsage {
				t.Errorf("exit code = %d, want %d (usage)", code, exitUsage)
			}
			if len(rr.calls) != 0 {
				t.Errorf("nothing may reach the engine once the euid is wrong:\n%+v", rr.calls)
			}
			if out != "" {
				t.Errorf("a refused command reports no rows, got stdout:\n%s", out)
			}
		})
	}
}

// TestPodmanEUIDGuardSkips: the three cases that pass without asking. Each asserts
// the runner was REACHED rather than a clean exit -- these commands go on to fail
// for their own reasons against a runner that answers nothing.
func TestPodmanEUIDGuardSkips(t *testing.T) {
	sample := writeCtrStandaloneEnv(t)

	t.Run("echo runner reaches no host", func(t *testing.T) {
		out, err := runCtr(t, sample, "broker", "status", "--platform", "podman")
		if err != nil {
			t.Fatalf("a preview must not be guarded: %v", err)
		}
		if !strings.Contains(out, "+ podman ps") {
			t.Errorf("the preview should still echo the probe:\n%s", out)
		}
	})

	t.Run("non-posix euid", func(t *testing.T) {
		_, _, rr := runGuarded(t, -1, "broker", "status", "--platform", "podman", "--env", sample)
		if len(rr.calls) == 0 {
			t.Error("with no POSIX euid there is nothing to compare, so the command proceeds")
		}
	})

	for _, euid := range []int{0, 1000} {
		t.Run("docker is never guarded", func(t *testing.T) {
			_, _, rr := runGuarded(t, euid, "broker", "status", "--platform", "docker", "--env", sample)
			if len(rr.calls) == 0 {
				t.Errorf("euid %d: docker has no rootless mode in this schema", euid)
			}
		})
	}
}

// TestGenerateIgnoresTheEUID is the exemption, and the reason the guard sits in
// prepare rather than in the two container constructors: `generate` is render-only,
// so an env file renders the same artifact from any account. A guard at
// ctrOps/ctrManager would have caught the Ops that detectContainerRole builds for
// role detection -- which issues no podman command -- and broken this.
func TestGenerateIgnoresTheEUID(t *testing.T) {
	for _, tc := range []struct {
		name string
		euid int
		args []string
	}{
		{"rootless file as root", 0, []string{"broker", "generate", "--env", writeRootlessPodmanEnv(t)}},
		{"rootful file as non-root", 1000,
			[]string{"broker", "generate", "--platform", "podman", "--env", writeCtrStandaloneEnv(t)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err, rr := runGuarded(t, tc.euid, tc.args...)
			if err != nil {
				t.Fatalf("generate renders from any account: %v", err)
			}
			if !strings.HasPrefix(out, "[Unit]") {
				t.Errorf("expected a quadlet unit on stdout, got:\n%s", out)
			}
			if len(rr.calls) != 0 {
				t.Errorf("generate reaches no host at all:\n%+v", rr.calls)
			}
		})
	}
}

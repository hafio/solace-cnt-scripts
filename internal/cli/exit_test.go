package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/exitcode"
	"github.com/spf13/cobra"

	"solace/internal/engine"
)

// k8sEnv appends the sample env plus --platform kubernetes. The sample declares all
// three platform sections, so an invocation that named none would be refused as
// ambiguous -- which is itself exit 2, and would make every case below pass for the
// wrong reason.
func k8sEnv(args ...string) []string {
	return append(withEnv(args...), "--platform", "kubernetes")
}

// TestExitCodeContract drives real invocations through the tree and asserts the code
// each produces. Classifying an error handed straight to exitcode.Code would prove only
// that the switch works; these prove the markers are actually in the paths an
// operator reaches.
//
// The documented gap is asserted too, deliberately: a mistyped TOP-LEVEL command
// exits 1 because cobra produces that error before any interception point, and
// pinning it means the day someone finds a clean fix, this is the case that tells
// them the fix worked (exitcode.Code, and the known gap in exit.go).
func TestExitCodeContract(t *testing.T) {
	k8sOnly := writeStandaloneEnv(t) // declares kubernetes: and nothing else

	cases := []struct {
		name string
		args []string
		want int
	}{
		{"unknown flag", k8sEnv("broker", "status", "--nope"), exitcode.Usage},
		{"flag missing its value", k8sEnv("broker", "status", "--pod"), exitcode.Usage},
		{"too many positionals", k8sEnv("validate", "extra"), exitcode.Usage},
		{"unknown noun under a verb", k8sEnv("broker", "bogus"), exitcode.Usage},
		{"role positional where none is taken", k8sEnv("broker", "logs", "primary"), exitcode.Usage},
		{"unknown word where none is taken", k8sEnv("broker", "logs", "bogus"), exitcode.Usage},
		{"bad --pod value", k8sEnv("broker", "status", "--pod", "nope"), exitcode.Usage},
		{"unparseable --since value", k8sEnv("broker", "logs", "--since", "half an hour"), exitcode.Usage},
		{"bad --platform value", withEnv("broker", "status", "--platform", "nope"), exitcode.Usage},
		{"--platform naming an undeclared section",
			[]string{"broker", "status", "--env", k8sOnly, "--platform", "docker"}, exitcode.Usage},
		{"env file that does not exist",
			[]string{"broker", "status", "--env", "no/such/file.yaml"}, exitcode.Usage},
		{"env file declaring no platform section",
			[]string{"broker", "status", "--env", writeNoPlatformEnv(t)}, exitcode.Usage},
		{"--allow-command where nothing executes",
			k8sEnv("broker", "generate", "--allow-command", "kubectl"), exitcode.Usage},
		{"kubernetes-only command on a container platform",
			[]string{"operator", "status", "--env", writeCtrStandaloneEnv(t), "--platform", "docker"}, exitcode.Usage},
		{"kubernetes-only flag on a container platform",
			[]string{"broker", "logs", "--previous", "--env", writeCtrStandaloneEnv(t), "--platform", "docker"}, exitcode.Usage},
		{"unknown example platform", []string{"examples", "--platform", "bogus"}, exitcode.Usage},
		// A script is a NAME in broker.cliScriptsDir: a path, or a name that folder does
		// not hold, is a command line to fix, not a broker to retry.
		{"cli script named by a path", k8sEnv("broker", "perform", "cli-script", "./setup.cli"), exitcode.Usage},
		{"shell script named by a path", k8sEnv("broker", "perform", "shell-script", "scripts/run.sh"), exitcode.Usage},
		{"script whose scripts folder does not exist",
			[]string{"broker", "perform", "cli-script", "setup.cli", "--env", k8sOnly}, exitcode.Usage},
		{"a noun run bare prints help", k8sEnv("broker"), exitcode.OK},
		{"--help", []string{"broker", "status", "--help"}, exitcode.OK},
		{"mistyped top-level command (documented gap)", []string{"depoy", "broker"}, exitcode.Failure},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// echoRunner throughout: the cases that get as far as a handler must not
			// reach a real kubectl, and installing it also skips the binary
			// resolution in App.load, so the suite does not depend on what is
			// installed on the machine running it.
			_, err := runRootWith(t, tc.args, echoRunner)
			if got := exitcode.Code(err); got != tc.want {
				t.Fatalf("`solace-util %s`: exit %d, want %d (err: %v)",
					strings.Join(tc.args, " "), got, tc.want, err)
			}
		})
	}
}

// writeNoPlatformEnv writes a syntactically valid env file that declares no platform
// section at all -- the "add kubernetes:, docker: or podman:" refusal, which is a
// wrong FILE rather than a wrong command line and is classified 2 for the same
// reason: no retry of the same invocation can fix it.
func writeNoPlatformEnv(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "no-platform.yaml")
	if err := os.WriteFile(path, []byte("redundancy:\n  enabled: false\n"), 0o644); err != nil {
		t.Fatalf("write env: %v", err)
	}
	return path
}

// TestEveryArgValidatorIsAUsageError is the drift guard behind markUsageArgs. The
// Args declarations are spread across four files; this walks the BUILT tree rather
// than trusting that the next one added was remembered.
//
// Each validator is offered nothing and then far too much. Every validator in this
// tree refuses one of those, and whichever it refuses must be a usage error. One
// that accepts both is reported too -- not because the command is wrong, but because
// this test has stopped saying anything about it.
func TestEveryArgValidatorIsAUsageError(t *testing.T) {
	root := newRootCmd(&App{})
	tooMany := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}

	var checked int
	walkTree(root, func(c *cobra.Command) {
		if c.Args == nil {
			return
		}
		var refused bool
		for _, args := range [][]string{nil, tooMany} {
			err := c.Args(c, args)
			if err == nil {
				continue
			}
			refused = true
			if !exitcode.IsUsage(err) {
				t.Errorf("%s: Args refused %d arguments with a non-usage error: %v",
					c.CommandPath(), len(args), err)
			}
		}
		if !refused {
			t.Errorf("%s: Args accepted both 0 and %d arguments, so this test proves nothing about it",
				c.CommandPath(), len(tooMany))
			return
		}
		checked++
	})
	if checked == 0 {
		t.Fatal("no argument validators found: the walk is not reaching the tree")
	}
}

// TestRootKeepsCobrasUnknownCommandError pins the reason for the documented gap.
// Cobra produces its unknown-command error inside Find only while root.Args is nil;
// give root a validator and the error is never produced, and a non-runnable root
// with no error prints help and exits 0 -- so a mistyped command would report
// success. This fails at that change rather than letting it ship.
func TestRootKeepsCobrasUnknownCommandError(t *testing.T) {
	root := newRootCmd(&App{})
	if root.Args != nil {
		t.Fatal("root has an Args validator: cobra will no longer produce its unknown-command error, " +
			"and a mistyped top-level command will print help and exit 0 (see the known gap in exit.go)")
	}
	_, err := runRootWith(t, []string{"depoy", "broker"}, echoRunner)
	if err == nil {
		t.Fatal("a mistyped top-level command reported success")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("expected cobra's unknown-command error, got %v", err)
	}
}

// interactiveFailRunner is engine.Echo with RunInteractive failing. Every other
// method still echoes, so the preflight and the argv assertions behave as they do
// under the plain Echo runner and only the handed-over session fails.
type interactiveFailRunner struct {
	engine.Echo
	err error
}

func (r interactiveFailRunner) RunInteractive(context.Context, string, ...string) error { return r.err }

// runFailRunner is engine.Echo with Run failing -- the non-interactive counterpart,
// for proving that the SAME error from a command that is not a handed-over session
// is an ordinary failure.
type runFailRunner struct {
	engine.Echo
	err error
}

func (r runFailRunner) Run(context.Context, string, ...string) error { return r.err }

// TestChildExitStatusIsScopedToInteractiveSessions is the property that matters
// about the passthrough: the same child error, carrying the same status, becomes
// this tool's exit code from `shell` and does NOT from `logs broker`.
//
// Without the scoping, every command would start reporting whatever kubectl or the
// engine happened to exit with, which is neither documentable nor usable in a
// script. Testing both halves against one error is what pins that.
func TestChildExitStatusIsScopedToInteractiveSessions(t *testing.T) {
	child := fmt.Errorf("kubectl: %w", childStatusError(t, 7))
	env := writeStandaloneEnv(t)

	_, err := runRootWith(t, []string{"broker", "shell", "--env", env}, func(a *App) {
		a.NewRunner = func(*App) engine.EnvRunner {
			return interactiveFailRunner{Echo: engine.Echo{W: io.Discard}, err: child}
		}
	})
	if got := exitcode.Code(err); got != 7 {
		t.Errorf("shell over a session that exited 7: got %d, want 7 (err: %v)", got, err)
	}

	_, err = runRootWith(t, []string{"broker", "logs", "--env", env}, func(a *App) {
		a.NewRunner = func(*App) engine.EnvRunner {
			return runFailRunner{Echo: engine.Echo{W: io.Discard}, err: child}
		}
	})
	if got := exitcode.Code(err); got != exitcode.Failure {
		t.Errorf("logs broker over the same error: got %d, want %d -- the engine's status must not "+
			"escape a command that is not a handed-over session (err: %v)", got, exitcode.Failure, err)
	}
}

// TestHelperExitProcess is not a test: it is the child process childStatusError
// starts. This is the standard os/exec helper-process pattern -- the test binary
// re-executes itself -- because a real *exec.ExitError cannot be built by hand (only
// os/exec fills in its ProcessState) and no one command spells "exit 7" on both
// Windows and Linux. The value is the status to exit with.
func TestHelperExitProcess(t *testing.T) {
	code := os.Getenv(exitCodeVar)
	if code == "" {
		t.Skip("helper process for childStatusError; nothing to do when run directly")
	}
	n, err := strconv.Atoi(code)
	if err != nil {
		t.Fatalf("%s=%q is not a number", exitCodeVar, code)
	}
	os.Exit(n)
}

const exitCodeVar = "SOLACE_TEST_CHILD_EXIT_CODE"

// childStatusError returns a real *exec.ExitError carrying the given status.
func childStatusError(t *testing.T, code int) error {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run", "^TestHelperExitProcess$")
	cmd.Env = append(os.Environ(), exitCodeVar+"="+strconv.Itoa(code))
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	err := cmd.Run()
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("helper process did not produce an exit status: %v", err)
	}
	if ee.ExitCode() != code {
		t.Fatalf("helper process exited %d, wanted %d", ee.ExitCode(), code)
	}
	return err
}

// walkTree visits c and every command beneath it.
func walkTree(c *cobra.Command, fn func(*cobra.Command)) {
	fn(c)
	for _, sub := range c.Commands() {
		walkTree(sub, fn)
	}
}

package engine

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/exitcode"
	"github.com/hafio/go-deploy-utility/runner"
	"github.com/hafio/go-deploy-utility/runner/runnertest"
)

// captureStdout swaps os.Stdout for a pipe while fn runs and returns whatever fn wrote to
// it. os.Stdout is always restored. Not safe with t.Parallel.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	done := make(chan []byte, 1)
	errc := make(chan error, 1)
	go func() {
		b, e := io.ReadAll(r)
		errc <- e
		done <- b
	}()

	fn()

	os.Stdout = old
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}
	if e := <-errc; e != nil {
		t.Fatalf("read pipe: %v", e)
	}
	out := <-done
	if err := r.Close(); err != nil {
		t.Fatalf("close pipe reader: %v", err)
	}
	return string(out)
}

// TestOnlyEnvRunnerCarriesRunEnv pins the narrowing, which is a security boundary rather
// than a tidiness one.
//
// Cmd.Env is how a secret reaches a child process without passing through an argv, and
// exactly one caller in the tool needs it: the container manager's compose deploy, for
// docker's environment-sourced secrets. Every other holder of a runner -- the Kubernetes
// cluster, both broker transports, the replication mate channel -- has no business putting
// a value in a child's environment, so each wraps its runner in NoEnv. The compose holder
// keeps the runner as given and still passes environment; NoEnv refuses it, naming the
// program and never the value, and never reaches the runner underneath. IsEcho sees through
// the wrapper, so a preview is still a preview.
func TestOnlyEnvRunnerCarriesRunEnv(t *testing.T) {
	const secret = "hunter2-secret-value"
	withEnv := runner.Cmd{Argv: []string{"docker", "compose", "up"}, Env: []string{"SOLACE_ADMIN_PASSWORD=" + secret}}
	plain := runner.Cmd{Argv: []string{"kubectl", "get", "pods"}}

	t.Run("the unwrapped runner passes environment", func(t *testing.T) {
		fake := &runnertest.Fake{}
		if err := fake.Run(t.Context(), withEnv, nil, nil); err != nil {
			t.Fatalf("Run: %v", err)
		}
		if calls := fake.Calls(); len(calls) != 1 || len(calls[0].Env) != 1 {
			t.Errorf("calls = %+v, want the one call with its environment", calls)
		}
	})

	t.Run("a wrapped runner refuses environment on Run", func(t *testing.T) {
		fake := &runnertest.Fake{}
		err := NoEnv(fake).Run(t.Context(), withEnv, nil, nil)
		if err == nil {
			t.Fatal("Run with Env was accepted")
		}
		if !strings.Contains(err.Error(), "docker") || !strings.Contains(err.Error(), "compose") {
			t.Errorf("error = %q, want it to name the program and say only compose may pass environment", err)
		}
		if strings.Contains(err.Error(), secret) || strings.Contains(err.Error(), "SOLACE_ADMIN_PASSWORD") {
			t.Errorf("error = %q, want no environment entry in it", err)
		}
		if n := len(fake.Calls()); n != 0 {
			t.Errorf("%d calls reached the runner underneath, want none", n)
		}
	})

	t.Run("a wrapped runner refuses environment on Attach", func(t *testing.T) {
		fake := &runnertest.Fake{}
		code, err := NoEnv(fake).Attach(withEnv, os.Stdin, os.Stdout, os.Stderr)
		if err == nil || code != 0 {
			t.Fatalf("Attach with Env = (%d, %v), want a refusal", code, err)
		}
		if strings.Contains(err.Error(), secret) {
			t.Errorf("error = %q, want no environment value in it", err)
		}
		if n := len(fake.Calls()); n != 0 {
			t.Errorf("%d calls reached the runner underneath, want none", n)
		}
	})

	t.Run("a wrapped runner refuses a malformed command first", func(t *testing.T) {
		fake := &runnertest.Fake{}
		if err := NoEnv(fake).Run(t.Context(), runner.Cmd{}, nil, nil); err == nil {
			t.Error("a Cmd with no program was accepted")
		}
	})

	t.Run("a wrapped runner passes everything else through", func(t *testing.T) {
		fake := &runnertest.Fake{Script: runnertest.Seq(runnertest.Reply{}, runnertest.Reply{Code: 5})}
		r := NoEnv(fake)
		if err := r.Run(t.Context(), plain, nil, nil); err != nil {
			t.Fatalf("Run: %v", err)
		}
		code, err := r.Attach(plain, os.Stdin, os.Stdout, os.Stderr)
		if err != nil || code != 5 {
			t.Fatalf("Attach = (%d, %v), want (5, nil)", code, err)
		}
		if calls := fake.Calls(); len(calls) != 2 || calls[0].Attach || !calls[1].Attach {
			t.Errorf("calls = %+v, want a Run then an Attach", calls)
		}
	})

	t.Run("IsEcho sees through the wrapper", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			r    runner.Runner
			want bool
		}{
			{"echo", runner.Echo{}, true},
			{"wrapped echo", NoEnv(runner.Echo{}), true},
			{"exec", runner.Exec{}, false},
			{"wrapped exec", NoEnv(runner.Exec{}), false},
			{"fake", &runnertest.Fake{}, false},
		} {
			if got := IsEcho(tc.r); got != tc.want {
				t.Errorf("IsEcho(%s) = %v, want %v", tc.name, got, tc.want)
			}
		}
	})
}

// TestRunStreamsToThisProcess: Run hands the child this process's stdout and stderr, so a
// command's output reaches the operator as it arrives, and a failure comes back as the
// runner returned it.
func TestRunStreamsToThisProcess(t *testing.T) {
	boom := errors.New("kubectl: exit status 1")
	fake := &runnertest.Fake{Script: runnertest.Seq(
		runnertest.Reply{Stdout: "applied\n"},
		runnertest.Reply{Err: boom},
	)}
	c := runner.Cmd{Argv: []string{"kubectl", "apply", "-f", "-"}, Stdin: []byte("kind: Namespace")}

	var err error
	got := captureStdout(t, func() { err = Run(t.Context(), fake, c) })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got != "applied\n" {
		t.Errorf("streamed %q, want the child's stdout", got)
	}
	if err := Run(t.Context(), fake, c); !errors.Is(err, boom) {
		t.Errorf("Run error = %v, want the runner's error", err)
	}
	for i, call := range fake.Calls() {
		// The first call ran while stdout was swapped for the capture pipe, so only the
		// second can be compared with the restored os.Stdout; the pipe's output above is
		// what proves the first one streamed.
		if i == 1 && call.Stdout != io.Writer(os.Stdout) {
			t.Errorf("call %d got stdout %v, want this process's own", i, call.Stdout)
		}
		if call.Stderr != io.Writer(os.Stderr) {
			t.Errorf("call %d got stderr %v, want this process's own", i, call.Stderr)
		}
		if string(call.Stdin) != "kind: Namespace" {
			t.Errorf("call %d stdin = %q, want the command's", i, call.Stdin)
		}
	}
}

// TestOutputCapturesStdoutAndStreamsStderr: Output returns what the child wrote to stdout,
// even when it failed -- a broker CLI or kubectl often says why in that stream -- while its
// stderr goes live to this process's.
func TestOutputCapturesStdoutAndStreamsStderr(t *testing.T) {
	boom := errors.New("kubectl: exit status 1")
	fake := &runnertest.Fake{Script: runnertest.Seq(
		runnertest.Reply{Stdout: "answer"},
		runnertest.Reply{Stdout: "reason", Err: boom},
	)}
	c := runner.Cmd{Argv: []string{"kubectl", "get", "pods"}}

	out, err := Output(t.Context(), fake, c)
	if err != nil || string(out) != "answer" {
		t.Fatalf("Output = (%q, %v), want (answer, nil)", out, err)
	}
	out, err = Output(t.Context(), fake, c)
	if !errors.Is(err, boom) || string(out) != "reason" {
		t.Errorf("failed Output = (%q, %v), want the captured text and the runner's error", out, err)
	}
	for i, call := range fake.Calls() {
		if call.Stdout == io.Writer(os.Stdout) {
			t.Errorf("call %d streamed stdout, want it captured", i)
		}
		if call.Stderr != io.Writer(os.Stderr) {
			t.Errorf("call %d stderr = %v, want this process's own", i, call.Stderr)
		}
	}
}

// TestInteractiveReturnsTheSessionStatus: Interactive attaches this process's terminal; a
// clean session is nil, a non-zero one is an exitcode.Exited error that carries the
// status (so the CLI can pass it through with ChildExit, and not otherwise), and an
// Attach error -- the session never ran -- is returned as it is.
func TestInteractiveReturnsTheSessionStatus(t *testing.T) {
	c := runner.Cmd{Argv: []string{"kubectl", "exec", "-it", "pod", "--", "bash"}}
	neverRan := errors.New("kubectl: not found on PATH")
	fake := &runnertest.Fake{Script: runnertest.Seq(
		runnertest.Reply{},
		runnertest.Reply{Code: 7},
		runnertest.Reply{Err: neverRan},
	)}

	if err := Interactive(fake, c); err != nil {
		t.Errorf("a session that exited 0: %v, want nil", err)
	}

	err := Interactive(fake, c)
	if err == nil || err.Error() != "kubectl: exit status 7" {
		t.Fatalf("a session that exited 7: %v, want %q", err, "kubectl: exit status 7")
	}
	if code := exitcode.Code(err); code != exitcode.Failure {
		t.Errorf("unmarked status exits %d, want %d: only ChildExit passes a session's status through",
			code, exitcode.Failure)
	}
	if code, ok := exitcode.ChildStatus(exitcode.ChildExit(err)); !ok || code != 7 {
		t.Errorf("ChildStatus(ChildExit(err)) = (%d, %v), want (7, true)", code, ok)
	}

	err = Interactive(fake, c)
	if !errors.Is(err, neverRan) {
		t.Errorf("an Attach error: %v, want it returned as is", err)
	}
	if _, ok := exitcode.ChildStatus(exitcode.ChildExit(err)); ok {
		t.Error("an Attach error carries an exit status, want none")
	}
	for i, call := range fake.Calls() {
		if !call.Attach {
			t.Errorf("call %d was not an Attach", i)
		}
	}
}

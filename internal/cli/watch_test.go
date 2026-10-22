package cli

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"solace/internal/engine"
)

// stopAfter is a Sleep seam that records each interval it was asked to wait and ends the
// watch on its nth call, so a test runs exactly n refreshes without waiting for any.
func stopAfter(n int, got *[]time.Duration) func(context.Context, time.Duration) error {
	return func(_ context.Context, d time.Duration) error {
		*got = append(*got, d)
		if len(*got) >= n {
			return context.Canceled
		}
		return nil
	}
}

// watchSection matches the timestamped section line an appended refresh opens with.
var watchSection = regexp.MustCompile(`(?m)^=== \d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} =+$`)

// TestWatchRedrawsInPlaceOnATerminal: on a terminal every refresh clears the screen, names
// what it is watching and how often, then reruns the same report -- here twice, at the
// --interval asked for.
func TestWatchRedrawsInPlaceOnATerminal(t *testing.T) {
	var slept []time.Duration
	out, err := runRootWith(t, []string{"broker", "status", "--all", "--watch", "--interval", "1s",
		"--platform", "docker", "--env", writeCtrStandaloneEnv(t)}, func(a *App) {
		echoRunner(a)
		a.Terminal = func() bool { return true }
		a.Sleep = stopAfter(2, &slept)
	})
	if err != nil {
		t.Fatalf("err = %v, want nil: the watch ends cleanly", err)
	}
	if got := strings.Count(out, clearScreen); got != 2 {
		t.Errorf("cleared the screen %d times, want once per refresh (2):\n%q", got, out)
	}
	if got := strings.Count(out, "Every 1s: solace-util broker status --all"); got != 2 {
		t.Errorf("header shown %d times, want 2:\n%s", got, out)
	}
	if got := strings.Count(out, "+ docker ps"); got != 2 {
		t.Errorf("the report ran %d times, want 2:\n%s", got, out)
	}
	if len(slept) != 2 || slept[0] != time.Second {
		t.Errorf("slept %v, want two waits of 1s", slept)
	}
}

// TestWatchAppendsWhenNotATerminal: piped or redirected, a refresh never writes an escape
// byte -- each report is appended under its own timestamped section line instead, so a log
// of a watch stays plain text.
func TestWatchAppendsWhenNotATerminal(t *testing.T) {
	var slept []time.Duration
	out, err := runRootWith(t, []string{"broker", "status", "--watch",
		"--platform", "docker", "--env", writeCtrStandaloneEnv(t)}, func(a *App) {
		echoRunner(a)
		a.Terminal = func() bool { return false }
		a.Sleep = stopAfter(2, &slept)
	})
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if strings.Contains(out, "\x1b") {
		t.Errorf("a redirected watch wrote an escape sequence:\n%q", out)
	}
	if got := len(watchSection.FindAllString(out, -1)); got != 2 {
		t.Errorf("found %d timestamped section lines, want one per refresh (2):\n%s", got, out)
	}
	if len(slept) != 2 || slept[0] != defaultWatchInterval {
		t.Errorf("slept %v, want the 5s default each time", slept)
	}
}

// TestWatchKeepsGoingAfterAFailedRefresh: a refresh that fails -- here the engine refusing
// the listing -- is shown, and the next one still runs. Someone watching wants to see an
// outage come and go, not restart the watch by hand.
func TestWatchKeepsGoingAfterAFailedRefresh(t *testing.T) {
	ps := 0
	rr := &opRunner{fail: func(c opCall) error {
		if len(c.args) > 0 && c.args[0] == "ps" {
			ps++
			if ps == 1 {
				return errors.New("engine unreachable")
			}
		}
		return nil
	}}
	var slept []time.Duration
	var err error
	errOut := captureStderr(t, func() {
		_, err = runRootWith(t, []string{"broker", "status", "--all", "--watch", "--interval", "2s",
			"--platform", "docker", "--env", writeCtrStandaloneEnv(t)}, func(a *App) {
			a.NewRunner = func(*App) engine.EnvRunner { return rr }
			a.Terminal = func() bool { return false }
			a.Sleep = stopAfter(2, &slept)
		})
	})
	if err != nil {
		t.Fatalf("err = %v, want nil: a failed refresh does not end the watch", err)
	}
	if !strings.Contains(errOut, "engine unreachable") || !strings.Contains(errOut, "trying again in 2s") {
		t.Errorf("stderr = %q, want the failure shown with the retry", errOut)
	}
	if ps != 2 {
		t.Errorf("the listing ran %d times, want 2 -- the refresh after the failure must still run", ps)
	}
}

// TestWatchEndsQuietlyWhenInterrupted: Ctrl-C mid-read makes the read fail too (its child
// got the same interrupt), and that failure is the watch ending, not something to report
// or retry.
func TestWatchEndsQuietlyWhenInterrupted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	runs := 0
	var err error
	errOut := captureStderr(t, func() {
		captureStdout(t, func() {
			err = watchLoop(ctx, &App{Terminal: func() bool { return false }}, "broker status", time.Second,
				func() error { runs++; return errors.New("signal: interrupt") })
		})
	})
	if err != nil || runs != 1 {
		t.Errorf("err = %v after %d run(s), want nil after exactly one", err, runs)
	}
	if strings.Contains(errOut, "trying again") {
		t.Errorf("an interrupted read was reported as a failure to retry: %q", errOut)
	}
}

// TestWatchSleepHonoursCtrlC covers the real timer: it waits out an interval, and returns
// at once when the context is already cancelled rather than finishing the wait.
func TestWatchSleepHonoursCtrlC(t *testing.T) {
	a := &App{}
	if err := a.sleep(context.Background(), time.Millisecond); err != nil {
		t.Errorf("a completed wait returned %v, want nil", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	start := time.Now()
	if err := a.sleep(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Errorf("a cancelled wait returned %v, want context.Canceled", err)
	}
	if time.Since(start) > time.Minute {
		t.Error("a cancelled wait did not return at once")
	}
}

// TestWatchNeverRedrawsIntoAPipe: with no seam, a redraw needs stdout to be a terminal, so
// output going into a pipe -- as a test's captured stdout does -- is never sent escape bytes.
func TestWatchNeverRedrawsIntoAPipe(t *testing.T) {
	var redraw bool
	captureStdout(t, func() { redraw = (&App{}).canRedraw() })
	if redraw {
		t.Error("canRedraw() = true with stdout a pipe; escape bytes would land in the pipe")
	}
}

// TestWatchIntervalIsValidated: each bad --interval is a usage error (exit 2) refused before
// anything runs -- the echo runner prints nothing -- including an --interval given without
// --watch, which a script would expect to start a watch it never gets.
func TestWatchIntervalIsValidated(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"below 1s", []string{"--watch", "--interval", "500ms"}, "below 1s"},
		{"not a duration", []string{"--watch", "--interval", "soon"}, "not a duration"},
		{"without --watch", []string{"--interval", "10s"}, "means nothing without --watch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"broker", "status", "--platform", "docker", "--env", writeCtrStandaloneEnv(t)}, tc.args...)
			out, err := runRootWith(t, args, func(a *App) {
				echoRunner(a)
				a.Sleep = func(context.Context, time.Duration) error {
					t.Fatal("a refused --interval must not start a watch")
					return nil
				}
			})
			if err == nil || ExitCode(err) != exitUsage || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v (exit %d), want a usage error containing %q", err, ExitCode(err), tc.want)
			}
			if out != "" {
				t.Errorf("a refused --interval still ran the report: %q", out)
			}
		})
	}
}

// TestStatusWithoutWatchRunsOnce: the flags change nothing unless --watch is given.
func TestStatusWithoutWatchRunsOnce(t *testing.T) {
	out, err := runRootWith(t, []string{"broker", "status", "--platform", "docker", "--env", writeCtrStandaloneEnv(t)},
		func(a *App) {
			echoRunner(a)
			a.Sleep = func(context.Context, time.Duration) error {
				t.Fatal("status without --watch must not wait for a second run")
				return nil
			}
		})
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got := strings.Count(out, "+ docker ps"); got != 1 {
		t.Errorf("the report ran %d times, want once", got)
	}
}

// TestOperatorStatusWatches: the operator report watches the same way, its header naming
// --detail when that widens it.
func TestOperatorStatusWatches(t *testing.T) {
	var slept []time.Duration
	out, err := runRootWith(t, append(withEnv("operator", "status", "--detail", "--watch", "--interval", "3s"),
		"--platform", "kubernetes"), func(a *App) {
		echoRunner(a)
		a.Terminal = func() bool { return true }
		a.Sleep = stopAfter(1, &slept)
	})
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if !strings.Contains(out, "Every 3s: solace-util operator status --detail") {
		t.Errorf("stdout = %q, want the header naming the watched command", out)
	}
	if len(slept) != 1 || slept[0] != 3*time.Second {
		t.Errorf("slept %v, want one wait of 3s", slept)
	}
}

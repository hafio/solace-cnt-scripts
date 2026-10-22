package cli

import (
	"context"
	"os"
	"os/signal"
	"time"

	"solace/internal/output"
)

// watch.go is `--watch` on the status commands: the same report, re-run every --interval
// until Ctrl-C.
//
// The report itself is untouched -- BrokerReport, ClusterReport, OperatorReport and the
// container Status/StatusAll run exactly as they do once -- so a watch cannot show
// anything a single run would not. What this adds is the loop around it, and the loop
// makes three decisions:
//
// Where stdout is a terminal that takes escape sequences, each refresh clears the screen
// and redraws in place, as `watch` does. That is the one escape sequence this tool
// writes, and it is why it is sent ONLY to such a terminal: everywhere else -- a pipe, a
// log file, Git Bash's mintty pipe, an old Windows console that refuses VT processing --
// each report is appended under a timestamped section line instead, so the output stays
// the plain ASCII the rest of the tool promises.
//
// A refresh that fails (an unreachable cluster or engine) is shown and the watch goes on:
// a transient outage is exactly what someone watching wants to see come and go, and
// ending the watch on it would make them restart it by hand.
//
// Ctrl-C ends the watch with exit 0, as `watch` does -- it is how a watch is meant to
// stop, not a failure. A kubectl or engine read in flight gets the same interrupt from
// the terminal and exits, and the loop, seeing its own context cancelled, reports nothing
// about it.

// clearScreen homes the cursor and clears the screen. Written only by watchLoop, and only
// when canRedraw says the terminal will interpret it.
const clearScreen = "\x1b[H\x1b[2J"

// watchTimeLayout is the timestamp a refresh carries: readable, and second-precise like
// the interval it repeats at.
const watchTimeLayout = "2006-01-02 15:04:05"

// withWatch runs fn once, or -- under --watch -- every --interval until Ctrl-C. The flags
// are read first either way, so a bad --interval is refused even without --watch, before
// anything reaches the cluster.
func withWatch(label string, fn opFunc) opFunc {
	return func(a *App) error {
		every, err := watchInterval(a)
		if err != nil {
			return err
		}
		if !a.watch {
			return fn(a)
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		return watchLoop(ctx, a, watchLabel(label, a), every, func() error { return fn(a) })
	}
}

// watchLoop is the loop itself, separate from the signal handling so a test can hand it a
// context of its own.
func watchLoop(ctx context.Context, a *App, label string, every time.Duration, run func() error) error {
	redraw := a.canRedraw()
	for {
		now := time.Now().Format(watchTimeLayout)
		if redraw {
			out := output.New(os.Stdout)
			out.Line("%sEvery %s: solace-util %s   %s   Ctrl-C to stop", clearScreen, every, label, now)
			out.Line("")
		} else {
			output.New(os.Stdout).Section(now)
		}
		if err := run(); err != nil {
			if ctx.Err() != nil {
				return nil // interrupted mid-read: the watch is over, not failed
			}
			warn("%v -- trying again in %s", err, every)
		}
		if err := a.sleep(ctx, every); err != nil {
			return nil
		}
	}
}

// watchLabel is the command the header names, with the flags that change what it shows.
func watchLabel(label string, a *App) string {
	if a.all {
		label += " --all"
	}
	if a.detail {
		label += " --detail"
	}
	return label
}

// canRedraw reports whether a refresh may clear the screen: stdout is a terminal, and on
// Windows its console accepted VT processing (enableVT). The App seam stands in for both.
func (a *App) canRedraw() bool {
	if a.Terminal != nil {
		return a.Terminal()
	}
	return stdoutIsTerminal() && enableVT()
}

// stdoutIsTerminal reports whether stdout is a character device -- a terminal or console,
// never a pipe or a file. Here the stream type IS the question, unlike the prompts (which
// decide by reading stdin): escape bytes written into a file or a pipe corrupt it.
func stdoutIsTerminal() bool {
	fi, err := os.Stdout.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// sleep waits one interval, or returns the context's error once Ctrl-C cancels it. The App
// seam lets a test count refreshes and stop the loop without waiting.
func (a *App) sleep(ctx context.Context, d time.Duration) error {
	if a.Sleep != nil {
		return a.Sleep(ctx, d)
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

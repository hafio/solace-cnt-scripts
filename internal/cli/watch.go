package cli

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/hafio/go-deploy-utility/watch"
)

// watch.go is `--watch` on the status commands: the same report, re-run every --interval
// until Ctrl-C.
//
// The report itself is untouched -- BrokerReport, ClusterReport, OperatorReport and the
// container Status/StatusAll run exactly as they do once -- so a watch cannot show
// anything a single run would not. What this adds is the loop around it, which is the
// shared watch package's (go-deploy-utility), the same one solmq-conn-util runs; this
// file wires it to the App seams and the status commands. The loop makes three decisions:
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

// watchLoop hands the shared loop this tool's title, seams and stdout, separate from the
// signal handling so a test can hand it a context of its own. A nil App.Sleep is the shared
// real timer.
func watchLoop(ctx context.Context, a *App, label string, every time.Duration, run func() error) error {
	return watch.Loop(ctx, watch.Options{
		Every:  every,
		Title:  "solace-util " + label,
		Redraw: a.canRedraw(),
		Out:    os.Stdout,
		Warn:   func(msg string) { warn("%s", msg) },
		Sleep:  a.Sleep,
	}, func(time.Time) error { return run() })
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

// canRedraw reports whether a refresh may clear the screen: stdout is a terminal that takes
// escape sequences, which watch.CanRedraw decides (on Windows it asks the console for VT
// processing). The App seam stands in for it.
func (a *App) canRedraw() bool {
	if a.Terminal != nil {
		return a.Terminal()
	}
	return watch.CanRedraw(os.Stdout)
}

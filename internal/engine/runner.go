// Package engine is this tool's wiring over the shared command runner
// (github.com/hafio/go-deploy-utility/runner): the verbose announcer, the holders'
// refusal to carry environment, and the helpers that point a command's streams at this
// process (or at a Redirect's writers). Execution, resolution, quoting and echoing all
// live in the shared module.
package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/hafio/go-deploy-utility/exitcode"
	"github.com/hafio/go-deploy-utility/runner"

	"solace/internal/output"
)

// NewExec builds the runner the CLI uses. With verbose, every command is announced to
// w as it runs, through the same internal/output Sink the CLI's own progress lines use
// -- output is a leaf package, so both sides share the one `==> ` definition instead of
// each spelling it out. Without verbose the runner says nothing: the preamble already
// reported what resolved.
func NewExec(w io.Writer, verbose bool) runner.Exec {
	if !verbose {
		return runner.Exec{}
	}
	sink := output.New(w)
	return runner.Exec{Announce: func(path string, args []string) {
		sink.Step("exec: %s", runner.Quote(path, args...))
	}}
}

// noEnv is the Runner NoEnv returns. It is a value type so IsEcho can see through it.
type noEnv struct{ runner.Runner }

// NoEnv narrows a Runner to one that refuses a Cmd carrying Env.
//
// Env is how a secret reaches a child without passing through an argv, and exactly ONE
// caller in the tool needs it -- container.Manager's compose deploy, which is how docker's
// environment-sourced secrets arrive. Every other holder of a runner (k8s.Cluster, both
// transports, the replication mate channel) has no business passing environment to a
// child, so each wraps what it is given in NoEnv and a builder that starts putting a value
// there fails loudly instead of quietly widening what the child inherits.
func NoEnv(r runner.Runner) runner.Runner { return noEnv{r} }

// Run refuses a Cmd with Env, naming the program but never the entries: they are secret
// values as often as not.
func (n noEnv) Run(ctx context.Context, c runner.Cmd, stdout, stderr io.Writer) error {
	if err := refuseEnv(c); err != nil {
		return err
	}
	return n.Runner.Run(ctx, c, stdout, stderr)
}

// Attach is Run's counterpart for an attached session.
func (n noEnv) Attach(c runner.Cmd, stdin, stdout, stderr *os.File) (int, error) {
	if err := refuseEnv(c); err != nil {
		return 0, err
	}
	return n.Runner.Attach(c, stdin, stdout, stderr)
}

func refuseEnv(c runner.Cmd) error {
	if err := c.Check(); err != nil {
		return err
	}
	if len(c.Env) > 0 {
		return fmt.Errorf("%s: refusing to pass environment to the child: only the compose deploy may do that, "+
			"so this command must carry its values another way (the entries are not shown, in case they are secrets)",
			c.Argv[0])
	}
	return nil
}

// redirect is the Runner Redirect returns. It is a value type so IsEcho and streams can
// see through it.
type redirect struct {
	runner.Runner
	out, errOut io.Writer
}

// Redirect makes r's children write what Run, Output, RunTee and OutputTee would send to
// this process's stdout and stderr to out and errOut instead. A `--watch` refresh uses it
// to collect a report into the frame it draws in one write, so a command's own output
// and kubectl's complaints land under the header with the rest of the report rather
// than on a screen the next frame clears. A buffer a helper captures into is unaffected.
func Redirect(r runner.Runner, out, errOut io.Writer) runner.Runner {
	return redirect{Runner: r, out: out, errOut: errOut}
}

// unwrap looks through the wrappers this package adds and reports the redirect among
// them, if any.
func unwrap(r runner.Runner) (runner.Runner, *redirect) {
	var d *redirect
	for {
		switch w := r.(type) {
		case noEnv:
			r = w.Runner
		case redirect:
			if d == nil {
				d = &w
			}
			r = w.Runner
		default:
			return r, d
		}
	}
}

// streams is where a child's output goes when no buffer captures it: the Redirect's
// writers, else this process's stdout and stderr.
func streams(r runner.Runner) (io.Writer, io.Writer) {
	if _, d := unwrap(r); d != nil {
		return d.out, d.errOut
	}
	return os.Stdout, os.Stderr
}

// IsEcho reports whether r only echoes (runner.Echo), looking through NoEnv and
// Redirect. Callers use it to skip work that only makes sense against a real engine:
// file writes, host probes, questions a preview cannot answer.
func IsEcho(r runner.Runner) bool {
	inner, _ := unwrap(r)
	_, ok := inner.(runner.Echo)
	return ok
}

// Run executes c with the child's output going to this process's stdout and stderr, or
// a Redirect's writers.
func Run(ctx context.Context, r runner.Runner, c runner.Cmd) error {
	out, errOut := streams(r)
	return r.Run(ctx, c, out, errOut)
}

// Output executes c and returns its captured stdout, with stderr streamed to this
// process's (or a Redirect's). It returns whatever the child managed to write even when
// it failed: a broker CLI or kubectl often says why in that same stream, and the callers
// that scan a transcript need the text more than they need the error alone.
func Output(ctx context.Context, r runner.Runner, c runner.Cmd) ([]byte, error) {
	_, errOut := streams(r)
	return runner.Output(ctx, r, c, errOut)
}

// RunTee is Run that also returns what the child wrote to stderr, which still reaches
// this process's stderr (or a Redirect's) as it arrives. It is for a step whose refusal
// must be told from its other failures by the shared kube.Forbidden: the operator sees
// kubectl's own words and the tool reads them too.
func RunTee(ctx context.Context, r runner.Runner, c runner.Cmd) (string, error) {
	var errBuf bytes.Buffer
	out, errOut := streams(r)
	err := r.Run(ctx, c, out, io.MultiWriter(errOut, &errBuf))
	return errBuf.String(), err
}

// OutputTee is Output that also returns what the child wrote to stderr, streamed to this
// process's stderr (or a Redirect's) as well, for a read whose refusal must be told from
// its other failures (kube.Forbidden).
func OutputTee(ctx context.Context, r runner.Runner, c runner.Cmd) ([]byte, string, error) {
	var out, errBuf bytes.Buffer
	_, errOut := streams(r)
	err := r.Run(ctx, c, &out, io.MultiWriter(errOut, &errBuf))
	return out.Bytes(), errBuf.String(), err
}

// Interactive wires this process's own terminal through to the child, for interactive
// sessions (`exec -it`, a Solace CLI, a shell). A non-zero session status comes back as
// exitcode.Exited, which is a plain failure until the CLI marks it with
// exitcode.ChildExit; an error from Attach means the session never ran and is returned
// as it is.
func Interactive(r runner.Runner, c runner.Cmd) error {
	code, err := r.Attach(c, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}
	return exitcode.Exited(c.Argv[0], code)
}

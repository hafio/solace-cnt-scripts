// Package engine is this tool's wiring over the shared command runner
// (github.com/hafio/go-deploy-utility/runner): the verbose announcer, the holders'
// refusal to carry environment, and the helpers that point a command's streams at this
// process. Execution, resolution, quoting and echoing all live in the shared module.
package engine

import (
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

// IsEcho reports whether r only echoes (runner.Echo), looking through NoEnv. Callers use
// it to skip work that only makes sense against a real engine: file writes, host probes,
// questions a preview cannot answer.
func IsEcho(r runner.Runner) bool {
	if n, ok := r.(noEnv); ok {
		r = n.Runner
	}
	_, ok := r.(runner.Echo)
	return ok
}

// Run executes c with the child's output going to this process's stdout and stderr.
func Run(ctx context.Context, r runner.Runner, c runner.Cmd) error {
	return r.Run(ctx, c, os.Stdout, os.Stderr)
}

// Output executes c and returns its captured stdout, with stderr streamed to this
// process's. It returns whatever the child managed to write even when it failed: a broker
// CLI or kubectl often says why in that same stream, and the callers that scan a
// transcript need the text more than they need the error alone.
func Output(ctx context.Context, r runner.Runner, c runner.Cmd) ([]byte, error) {
	return runner.Output(ctx, r, c, os.Stderr)
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

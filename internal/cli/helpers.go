package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"solace/internal/config"
)

// emit writes rendered bytes to stdout. Used by `gen` and `--gen` paths so the
// artifact is the command's only stdout (progress/warnings go to stderr).
func emit(b []byte) error {
	_, err := os.Stdout.Write(b)
	return err
}

// renderAnnotation marks a command that ONLY renders -- it never executes an
// external command, whatever flags it is given. `gen` is the whole set today. It
// exists so checkAllowCommand can refuse --allow-command where there is nothing to
// allow, the mirror of what genAnnotation does for the --gen-*-only trio.

const renderAnnotation = "solace_render_only"

// renderOnly tags a command as never executing, and returns it so registration can
// wrap inline: renderOnly(genCapable(...)).

func renderOnly(c *cobra.Command) *cobra.Command {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[renderAnnotation] = "true"
	return c
}

// withLong attaches extended help to a command built by leaf/roleLeaf/dispatchLeaf,
// which take a Short and nothing else. It returns the command so registration can
// wrap inline the way onlyOn and renderOnly do -- onlyOn(withLong(leaf(...), x), ...)
// -- rather than forcing a named variable and a separate assignment for every
// command that has something more to say than its one-line Short.
//
// A Long is a BRIEF summary -- one to four lines -- of what the command does and how
// to invoke it. Mechanics, constraints and history belong in docs/operations.md, not
// in help; a destructive command carries exactly one sentence saying what it
// destroys. The generator skips a Long equal to the Short (commanddoc_test.go), so a
// Long that would only restate the Short is not written at all.

func withLong(c *cobra.Command, long string) *cobra.Command {
	c.Long = long
	return c
}

// willExecute reports whether this invocation can reach an external command. Only
// the render-only commands cannot: they build an artifact from the env file and
// print it, touching nothing.

func (a *App) willExecute(cmd *cobra.Command) bool {
	return cmd.Annotations[renderAnnotation] != "true"
}

// addAllowCommandFlag wires --allow-command onto one command that executes. It is
// the operator's escape hatch for the execution-guard allowlist (config/execguard.go):
// a binary this tool does not drive by default -- a `microk8s kubectl`, a site
// wrapper -- runs only when the person at the keyboard names it, for that one
// invocation. It cannot approve a privilege-escalation wrapper at all: elevate this
// tool when you run it (`sudo solace-util ...`), never through an env file.
//
// It is a CLI flag and NOTHING else on purpose. There is no config key for it, no
// environment variable, and no binding layer that could give an env file a way to
// set it: an env file that could approve its own binary would make the allowlist
// decorative. wireExec adds it to each command that runs something rather than to
// root, so `solace-util convert --allow-command ...` is a usage error too.

type opFunc func(*App) error

// roleOpFunc is a leaf handler parameterized by a broker node role (p|b|m).

type roleOpFunc func(*App, config.Role) error

// leaf builds a no-arg subcommand that dispatches straight to fn. It takes no
// arguments, so NoFileCompletions is what it should offer -- cobra's default is
// to fall back to filenames, which would be wrong for every command built here.

func leaf(app *App, use, short string, fn opFunc) *cobra.Command {
	return wireExec(app, &cobra.Command{
		Use:               use,
		Short:             short,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              func(*cobra.Command, []string) error { return fn(app) },
	})
}

// layer describes the one thing a removal keeps by default: the part that is
// expensive to recreate and impossible to get back. Both removals ask about theirs
// the same way, through addLayerFlags/confirmLayer, so learning the contract on one
// teaches the other.

func emitOrWrite(a *App, body []byte, what string) error {
	if a.out == "" {
		return emit(body)
	}
	if _, err := os.Stat(a.out); err == nil {
		if !confirmAction(a, "Overwrite", "overwrite", fmt.Sprintf("%s with the %s", a.out, what)) {
			// Same verb confirmAction's own non-TTY warning uses, and no more: that line
			// has already named --no-prompt, and a declined prompt needs no explaining.
			return fmt.Errorf("refusing to overwrite %q", a.out)
		}
	}
	// 0o600 on every artifact this writes. On Kubernetes the stream carries the admin
	// password, the pre-shared key and the TLS private key base64-encoded, and a compose
	// file or quadlet unit names the secrets a broker reads -- neither is group-readable
	// material.
	if err := os.WriteFile(a.out, body, 0o600); err != nil {
		return fmt.Errorf("write %q: %w", a.out, err)
	}
	// WriteFile's perm argument is honoured only when it CREATES the file, so
	// re-exporting over an existing path would leave the artifact at whatever mode
	// that file already had -- 0644 under a default umask, world-readable, holding
	// the encrypted admin password, the config-sync pre-shared key and every client
	// secret. The chmod is what makes the 0600 above true on the overwrite path too.
	if err := os.Chmod(a.out, 0o600); err != nil {
		return fmt.Errorf("restrict %q to owner-only: %w", a.out, err)
	}
	step("wrote the %s to %s", what, a.out)
	return nil
}

// addRestartFlag wires --restart onto the deploy command. Deliberately separate
// from --no-prompt: bouncing a live broker to apply a changed artifact is its own
// explicit decision, the same way deleting its data is.

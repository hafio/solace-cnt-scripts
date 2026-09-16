package cli

import (
	"fmt"
	"os"
	"path/filepath"

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

// withLong attaches extended help to a command built by leaf or dispatchLeaf,
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

// opFunc is a leaf handler that needs only the app context.
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

// emitOrWrite sends body to stdout, or to app.out when --out named a file.
//
// The overwrite question is asked the same way every destructive question in this tree is
// asked (confirmAction): an exact prompt when there is a terminal, yes under --no-prompt,
// and a refusal that names the flag when there is neither. `what` describes the artifact
// for the confirmation line, since "overwrite x.yaml?" is more useful when it says what is
// about to be written over it.
func emitOrWrite(a *App, body []byte, what string) error {
	if a.out == "" {
		return emit(body)
	}
	if fi, err := os.Lstat(a.out); err == nil {
		// A symlink is refused, not followed: the artifact carries secrets, and writing
		// through a link an operator did not notice puts them wherever it points.
		if fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing to write %q: it is a symbolic link; name the real file", a.out)
		}
		if !confirmAction(a, "Overwrite", "overwrite", fmt.Sprintf("%s with the %s", a.out, what)) {
			// Same verb confirmAction's own non-TTY warning uses, and no more: that line
			// has already named --no-prompt, and a declined prompt needs no explaining.
			return fmt.Errorf("refusing to overwrite %q", a.out)
		}
	}
	// Written to a fresh 0600 file beside the target and RENAMED over it, rather than
	// opened and truncated in place. Three things that buys, all of which the plain
	// write got wrong:
	//
	// The mode is right before the first byte exists -- os.WriteFile honours its
	// permission only when it CREATES the file, so an overwrite left the artifact at
	// whatever mode was already there (0644 under a default umask) while the admin
	// password, pre-shared key and TLS private key were being written into it.
	//
	// A symlink at the target is REPLACED, not followed. The Lstat above refuses one it
	// can see, but a confirmation prompt sits between that check and the write, which is
	// an arbitrary amount of time for the path to become a link; rename overwrites the
	// link itself, so the window cannot be used.
	//
	// And a failed write leaves the previous file intact instead of a truncated one.
	tmp, err := os.CreateTemp(filepath.Dir(a.out), ".solace-*.tmp")
	if err != nil {
		return fmt.Errorf("write %q: %w", a.out, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once the rename below has succeeded
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return fmt.Errorf("restrict %q to owner-only: %w", a.out, err)
	}
	if _, err := tmp.Write(body); err != nil {
		tmp.Close()
		return fmt.Errorf("write %q: %w", a.out, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write %q: %w", a.out, err)
	}
	if err := os.Rename(tmpName, a.out); err != nil {
		return fmt.Errorf("write %q: %w", a.out, err)
	}
	step("wrote the %s to %s", what, a.out)
	return nil
}

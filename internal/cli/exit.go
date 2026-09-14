package cli

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

// Exit status is a three-way answer (the same shape the retired live-probe harness
// documents and tests the same three codes): 0 the command did what it was asked,
// 1 it was asked for something sensible and could not do it, 2 the request itself
// was wrong. A script that retries on a transient cluster failure must not retry a
// typo, and one code for both makes that undecidable.
//
// The line between 1 and 2 is "would a different command line have helped?" -- so
// an env file that fails to load is 2, not 1: the operator chose that file, and
// nothing about the cluster or the engine is going to change the outcome. That is
// the same reading a bad `--env` takes everywhere else.
const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
)

// usageError marks an error as being about the command line rather than about the
// world. It is a wrapper rather than a sentinel because every one of these errors
// says something specific -- which flag, which word, which platform -- and the
// classification must not cost that message.
type usageError struct{ err error }

func (e usageError) Error() string { return e.err.Error() }
func (e usageError) Unwrap() error { return e.err }

// usagef builds a usage error, for the refusals this package writes itself.
func usagef(format string, args ...any) error {
	return usageError{fmt.Errorf(format, args...)}
}

// asUsage marks an error produced elsewhere -- pflag, a cobra validator,
// config.ParseRole -- as a usage error, passing nil through and never
// double-wrapping.
func asUsage(err error) error {
	if err == nil || isUsage(err) {
		return err
	}
	return usageError{err}
}

func isUsage(err error) bool {
	var u usageError
	return errors.As(err, &u)
}

// childExitError marks an error whose OWN exit status is the answer, so this tool
// reports the child's code rather than its own 1. It is for the two commands that
// hand the terminal over -- `cli` and `shell` -- where the operator's last command
// inside the session is what they are asking about, and collapsing every non-zero
// session to 1 throws that away.
//
// It is deliberately not applied anywhere else. Every other command runs kubectl or
// a container engine as an implementation detail, and leaking those codes would make
// this tool's contract "whatever the engine felt like", which is unusable in a
// script and impossible to document.
//
// A passed-through status can of course BE 2, and then it is indistinguishable from
// this tool's own usage code. That is inherent to a passthrough rather than a defect:
// the operator asked for the session's answer, and these two commands are documented
// as giving it (docs/operations.md, Exit codes). It is also why the passthrough is
// two commands and not a general facility.
type childExitError struct{ err error }

func (e childExitError) Error() string { return e.err.Error() }
func (e childExitError) Unwrap() error { return e.err }

// childExit marks an interactive session's failure for status passthrough.
func childExit(err error) error {
	if err == nil {
		return nil
	}
	return childExitError{err}
}

// ExitCode classifies what Execute returned. main() is the only caller: keeping the
// mapping here rather than in main keeps it next to the errors it classifies and in
// a package that has tests.
//
// KNOWN GAP, deliberate: a mistyped TOP-LEVEL command exits 1, not 2. Cobra
// produces that error inside Find, before ExecuteC reaches any interception point,
// and the obvious fix is worse than the gap -- giving root an Args validator stops
// the error being produced at all, and a non-runnable command with no error prints
// help and exits 0, so `solace-util depoy broker` would look like success. Every
// mistyped SUB-command is 2: those are refused by group's RunE or by
// noRolePositional, both of which this package owns. See TestExitCodeContract.
func ExitCode(err error) int {
	switch {
	case err == nil:
		return exitOK
	case isUsage(err):
		return exitUsage
	default:
		if code, ok := childStatus(err); ok {
			return code
		}
		return exitFailure
	}
}

// childStatus reads the exit status out of a marked interactive session's error.
//
// A child killed by a signal reports -1, which os.Exit cannot take, so the status is
// returned with `code > 0` as the ok answer rather than through a branch: the
// signalled case is unreachable on Windows (TerminateProcess sets a real status), so
// a branch for it would be code no local coverage run could ever enter. One
// expression says the same thing and is exercised by every caller that gets here.
func childStatus(err error) (int, bool) {
	var marked childExitError
	if !errors.As(err, &marked) {
		return 0, false
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) {
		return 0, false
	}
	code := ee.ExitCode()
	return code, code > 0
}

// usageArgs wraps a positional-argument validator so its refusal carries the usage
// marker. A nil validator stays nil: cobra treats that as "use legacyArgs", which
// is what produces root's unknown-command error, and replacing it would silently
// turn a mistyped command into a successful help print (see ExitCode).
func usageArgs(v cobra.PositionalArgs) cobra.PositionalArgs {
	if v == nil {
		return nil
	}
	return func(cmd *cobra.Command, args []string) error { return asUsage(v(cmd, args)) }
}

// markUsageArgs marks every argument validator in the tree, walking it the way
// applyAliases does rather than wrapping at each declaration. There are fifteen
// declaration sites across four files; a walk cannot be forgotten at a sixteenth,
// and TestEveryArgValidatorIsAUsageError pins that it was not.
func markUsageArgs(c *cobra.Command) {
	c.Args = usageArgs(c.Args)
	for _, sub := range c.Commands() {
		markUsageArgs(sub)
	}
}

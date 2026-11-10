package cli

import (
	"github.com/hafio/go-deploy-utility/exitcode"
	"github.com/spf13/cobra"
)

// The exit-status contract (0 worked, 1 could not, 2 the request was wrong, a
// session's own status for `cli` and `shell`) lives in the shared exitcode package,
// which also holds the usage and child-exit markers. This file keeps only what is
// specific to this command tree.
//
// KNOWN GAP, deliberate: a mistyped TOP-LEVEL command exits 1, not 2. Cobra
// produces that error inside Find, before ExecuteC reaches any interception point,
// and the obvious fix is worse than the gap -- giving root an Args validator stops
// the error being produced at all, and a non-runnable command with no error prints
// help and exits 0, so `solace-util depoy broker` would look like success. Every
// mistyped SUB-command is 2: those are refused by group's RunE or by
// noRolePositional, both of which this package owns. See TestExitCodeContract.

// usageArgs wraps a positional-argument validator so its refusal carries the usage
// marker. A nil validator stays nil: cobra treats that as "use legacyArgs", which
// is what produces root's unknown-command error, and replacing it would silently
// turn a mistyped command into a successful help print (see the known gap above).
func usageArgs(v cobra.PositionalArgs) cobra.PositionalArgs {
	if v == nil {
		return nil
	}
	return func(cmd *cobra.Command, args []string) error { return exitcode.AsUsage(v(cmd, args)) }
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

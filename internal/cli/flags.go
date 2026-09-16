package cli

import (
	"github.com/spf13/cobra"

	"solace/internal/broker"
)

// flags.go registers the command-local flags and reads them back.
//
// Registrar and reader live together on purpose: a flag whose value is interpreted
// somewhere other than where it was declared is how a default drifts from what the help
// text promises. So addRemoveFlags sits beside nothing, addApplyRemoveFlags beside
// wantRemove, addDisableEnableFlags beside wantEnable, addExportFlags beside exportScope.
//
// The paired direction flags (--apply/--remove, --disable/--enable) are the reason the
// readers return an error rather than a bool: each leaf has a DEFAULT direction and a flag
// for the other one, both spellings are registered so a script can say which way it is
// going, and setting both at once is a usage error caught by the pair's own reader instead
// of by whichever handler happened to look first.

// addRemoveFlags wires the confirmation contract onto a command that asks before it acts.
// --no-prompt answers yes to every question that IS asked, so a script needs one thing
// switched off rather than one per question.
//
// l is the retained layer, or nil for a command that has none. The two flags answer
// DIFFERENT questions and compose rather than conflict:
//
//	--delete-data / --delete-crd   RAISES the layer question. Without it the layer is
//	                               kept and no question about it is asked at all.
//	--no-prompt                    answers yes to whatever was asked.
//
// So --no-prompt alone keeps the data, because no data question was raised; a fully
// unattended wipe is `--delete-data --no-prompt`. That is the inverse of the earlier
// design, where the layer flag SILENCED a question that was always asked -- which made
// --delete-data both the request and its own confirmation.
//
// The help text follows l for that reason. Seven commands pass nil (`broker restart`,
// `operator stop`, the four `broker configure` leaves), and telling their operator that
// "without a --delete-* flag the layer is still kept" names a flag those commands do not
// have, about a layer they do not touch.
func addRemoveFlags(c *cobra.Command, app *App, l *layer) {
	usage := "do not ask: answer yes to every question this command would ask"
	if l != nil {
		usage += ". It does not RAISE a question -- without --" + l.flag + ", " + l.what + " is still kept"
	}
	c.Flags().BoolVar(&app.noPrompt, "no-prompt", false, usage)
	if l != nil {
		c.Flags().BoolVar(&app.deleteLayer, l.flag, false, l.usage)
	}
}

// addApplyRemoveFlags wires the direction pair for a `broker configure` leaf whose
// default is to apply: --apply is the default and accepted explicitly, --remove is the
// other way. Both at once is a usage error.
func addApplyRemoveFlags(c *cobra.Command, app *App) {
	c.Flags().BoolVar(&app.flagApply, "apply", false,
		"apply what the env file configures (the default; accepted explicitly so a script can say so)")
	c.Flags().BoolVar(&app.flagRemove, "remove", false,
		"remove it from the broker instead of applying it")
}

// addDisableEnableFlags wires the direction pair for a hardening leaf, whose default is
// to disable: --disable is the default and accepted explicitly, --enable reverses it.
func addDisableEnableFlags(c *cobra.Command, app *App) {
	c.Flags().BoolVar(&app.flagDisable, "disable", false,
		"shut it down (the default; accepted explicitly so a script can say so)")
	c.Flags().BoolVar(&app.flagEnable, "enable", false,
		"start it back up instead of shutting it down")
}

// wantRemove reports whether an --apply/--remove leaf should remove. It is the one place
// the pair is read, so two leaves cannot decide the precedence differently.
func wantRemove(a *App) (bool, error) {
	if a.flagApply && a.flagRemove {
		return false, usagef("--apply and --remove ask for opposite things; pass one or neither " +
			"(neither means --apply)")
	}
	return a.flagRemove, nil
}

// wantEnable reports whether a --disable/--enable leaf should enable.
func wantEnable(a *App) (bool, error) {
	if a.flagDisable && a.flagEnable {
		return false, usagef("--disable and --enable ask for opposite things; pass one or neither " +
			"(neither means --disable)")
	}
	return a.flagEnable, nil
}

// addExportFlags wires the scope flags onto `broker perform export-config`.
//
// Scope is expressed by the PRESENCE of a value rather than by a --scope enum, which is
// the same grammar --pod already uses: a value narrows, absence means the default set. A
// `--scope total|broker|vpn` enum plus a --vpn list would leave three overlaps to memorise
// (--scope vpn with no --vpn, --vpn with --scope total, --scope broker --vpn x); this
// leaves exactly one, and exportScope refuses it.
//
// --vpn takes no completer that reads the broker, and cannot: completion never loads the
// env file, which is what stops a TAB press from parsing untrusted YAML
// (TestCompletionNeverReadsTheEnvFile). So it is NoFileCompletions -- a VPN name is not a
// path, and falling back to filename completion would offer nonsense.
func addExportFlags(c *cobra.Command, app *App) {
	c.Flags().StringArrayVar(&app.vpns, "vpn", nil,
		"capture only this message-VPN, repeatable for several; omit it and every VPN is captured")
	c.Flags().BoolVar(&app.brokerOnly, "broker-only", false,
		"capture only broker-level configuration, and no message-VPN at all")
	registerFlagCompletion(c, "vpn", cobra.NoFileCompletions)
}

// exportScope reads the two scope flags as one decision, the way wantRemove and
// wantEnable already read the configure direction flags: the refusal lives here
// once rather than in each caller.
func exportScope(a *App) ([]string, bool, error) {
	if a.brokerOnly && len(a.vpns) > 0 {
		return nil, false, usagef("--broker-only excludes every message-VPN, so naming --vpn %s "+
			"asks for opposite things; pass one or neither (neither captures both)", a.vpns[0])
	}
	// The value itself is checked here, at the flag, so an unusable name is the exit
	// code that means "the request was wrong" rather than the one that means "the
	// broker failed". broker.ExportConfig checks again; this decides the exit code.
	for _, v := range a.vpns {
		if err := broker.ValidVPNName(v); err != nil {
			return nil, false, asUsage(err)
		}
	}
	return a.vpns, a.brokerOnly, nil
}

// addOutFlags wires --out/-o (and the --no-prompt that answers its one question) onto a
// command whose output is an artifact you keep rather than read.
//
// It exists because redirection is not portable. `broker generate > x.yaml` in Windows
// PowerShell 5.1 re-encodes this tool's plain ASCII as UTF-16LE with a BOM, and kubectl
// then refuses the file with an error that points at the YAML rather than at the shell --
// the corruption happens after this process has exited, so nothing inside it can prevent
// it. Writing the file here is the only fix, and it buys three smaller things too: the
// file appears only if the render SUCCEEDED (a shell truncates it before the command even
// runs), the secret-bearing Kubernetes stream stays out of terminal scrollback, and the
// file is created 0600 rather than inheriting whatever the shell would have used.
func addOutFlags(c *cobra.Command, app *App) {
	c.Flags().StringVarP(&app.out, "out", "o", "", "write the artifact to this file instead of stdout")
	// Meaningful only alongside --out, and accepted without it rather than refused: a
	// script that passes it unconditionally is not making a mistake worth an exit 2, and
	// nothing is silently skipped -- with no --out there is no file and no question.
	c.Flags().BoolVar(&app.noPrompt, "no-prompt", false,
		"answer yes to the --out overwrite question (no effect without --out)")
	registerFlagCompletion(c, "out", cobra.FixedCompletions(nil, cobra.ShellCompDirectiveDefault))
}

// addRestartFlag wires --restart onto the deploy command. Deliberately separate from
// --no-prompt: bouncing a live broker to apply a changed artifact is its own explicit
// decision, the same way deleting its data is.
func addRestartFlag(c *cobra.Command, app *App) {
	c.Flags().BoolVar(&app.restart, "restart", false,
		"restart an already-running broker when the deploy artifact changed (otherwise you are asked, and a non-interactive run leaves it running)")
}

// addAllowCommandFlag wires --allow-command onto one command that executes. It is the
// operator's escape hatch for the execution-guard allowlist (config/execguard.go): a
// binary this tool does not drive by default -- a `microk8s kubectl`, a site wrapper --
// runs only when the person at the keyboard names it, for that one invocation. It cannot
// approve a privilege-escalation wrapper at all: elevate this tool when you run it
// (`sudo solace-util ...`), never through an env file.
//
// It is a CLI flag and NOTHING else on purpose. There is no config key for it, no
// environment variable, and no binding layer that could give an env file a way to set it:
// an env file that could approve its own binary would make the allowlist decorative.
// wireExec adds it to each command that runs something rather than to root, so
// `solace-util convert --allow-command ...` is a usage error too.
func addAllowCommandFlag(c *cobra.Command, app *App) {
	c.Flags().StringArrayVar(&app.AllowCommand, "allow-command", nil,
		"approve one extra binary for the config's platform command, for this run only "+
			"(repeatable; a bare name, never a path). The env file cannot grant this")
	// No file completion: the value is a bare binary name, and offering paths would
	// coach exactly the mistake the help text above warns against.
	registerFlagCompletion(c, "allow-command", cobra.NoFileCompletions)
}

// checkAllowCommand rejects --allow-command on an invocation that cannot execute
// anything. Silently accepting it there would teach the flag as harmless boilerplate
// -- exactly the habit that gets it pasted into a wrapper script, where it then
// applies to runs that DO execute. Hand-rolled rather than cobra's flag groups for
// the same reason checkGenFlags is: the flag is declared on the platform command and
// validated against the leaf that inherited it, which lets the error name the leaf.
func checkAllowCommand(cmd *cobra.Command, app *App) error {
	if len(app.AllowCommand) == 0 || app.willExecute(cmd) {
		return nil
	}
	return usagef("--allow-command is only valid on a command that runs something, and %q renders "+
		"without executing; drop the flag", cmd.CommandPath())
}

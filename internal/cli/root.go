package cli

import (
	"github.com/spf13/cobra"

	"solace/internal/config"
)

// Execute builds the command tree and runs it. main() calls this.
func Execute() error {
	app := &App{}
	root := newRootCmd(app)
	return root.Execute()
}

func newRootCmd(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:   "solace-util",
		Short: "Deploy and operate Solace PubSub+ brokers on Kubernetes, Docker, or Podman",
		Long: "Deploy and operate Solace PubSub+ Event Brokers from one YAML env file, with the\n" +
			"same commands on every platform. Name the thing, then the verb:\n" +
			"\n" +
			"  broker <verb>      the broker this env file describes\n" +
			"  operator <verb>    the cluster-scoped EventBroker Operator (kubernetes only)\n" +
			"  validate           check the whole env file\n" +
			"  examples           write a starting env file\n" +
			"\n" +
			"Every command takes -e/--env <file>; the platform comes from that file, or from\n" +
			"--platform kubernetes|docker|podman when it declares more than one.\n" +
			"\n" +
			"Exit status: 0 worked, 2 bad command line or env file, 1 anything else.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	// newCompletionCmd below replaces the one cobra would add during Execute, which
	// never appeared in the generated reference because that renders this tree
	// without executing it. Disabling cobra's is explicit rather than relying on it
	// standing down for a same-named command of ours.
	root.CompletionOptions.DisableDefaultCmd = true
	// Cobra runs a WINDOWS-ONLY pre-exec hook that decides whether this binary was
	// double-clicked in Explorer, so it can print "this is a command line application"
	// instead of running. Setting the text to "" is cobra's own switch for turning that
	// off, and it belongs off here for two independent reasons.
	//
	// It is wrong for this tool. `solace-util` is driven from a terminal, from scripts
	// and from CI; the message is for a GUI user who has no console, and a tool that
	// refuses to run because of how it was launched is a failure mode nobody here wants.
	//
	// And it is not free. The check walks the OS process table on every single
	// Execute(): `syscall.Getppid` plus a snapshot scan to find the parent. A CPU
	// profile of internal/cli put it at 37% of the whole suite -- 10.6 seconds of a
	// 13.8-second run, and 85% of everything cobra itself did, against 1.8 seconds for
	// the commands under test. Tests pay it once per command they run and a real
	// invocation pays it once, so this is a startup cost for operators too, not just a
	// test-suite one.
	cobra.MousetrapHelpText = ""
	// A flag pflag refuses -- unknown, or a bad value for its type -- is a usage
	// error, not a runtime failure, and must exit 2 like every other one. Cobra
	// resolves this hook through the nearest ancestor that sets it, so setting it
	// on root covers every command in the tree.
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return asUsage(err) })

	root.PersistentFlags().StringVarP(&app.EnvName, "env", "e", config.EnvFileDefault, "env file name, searched in the base dir then <base-dir>/env; a value with a directory is used as-is")
	root.PersistentFlags().StringVar(&app.BaseDir, "base-dir", "", "directory searched for the env file, and holding env/ (default: current directory)")
	root.PersistentFlags().BoolVarP(&app.Verbose, "verbose", "v", false, "announce every external command as it runs; by default the binaries this env file names are resolved and listed once, up front")
	root.PersistentFlags().StringVar(&app.PlatformFlag, "platform", "",
		"platform to drive: kubernetes (kube), docker (dk) or podman (pm). Default: the one the env file declares, "+
			"or a prompt if it declares several")

	// The persistent flags that take a value worth suggesting. An inherited
	// flag is the same *pflag.Flag in every subcommand, and cobra keys completion
	// functions by that pointer, so registering here covers the whole tree.
	registerFlagCompletion(root, "env", completeEnvFiles(app))
	registerFlagCompletion(root, "base-dir", completeDirs)
	registerFlagCompletion(root, "platform", completePlatforms)

	addCommands(root, app)
	root.AddCommand(
		newConvertCmd(app),
		newExamplesCmd(app),
		newCompletionCmd(),
		newVersionCmd(),
	)
	// Last, so it sees the whole tree: the abbreviations are attached by walking it,
	// and the collision check that comes with them can only be complete once every
	// command is in place.
	applyAliases(root)
	// Also last, and for the same reason: markUsageArgs wraps every Args validator in
	// the tree so a refusal from one exits 2 rather than 1. It has to run after every
	// command is attached, and after applyAliases, so nothing added later is missed.
	markUsageArgs(root)
	return root
}

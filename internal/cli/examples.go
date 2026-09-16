package cli

import (
	"github.com/spf13/cobra"

	"solace/internal/config"
	"solace/internal/examples"
)

// newExamplesCmd builds `solace-util examples`, the only way to get a starting env file
// out of a binary with no checkout beside it. A bare literal like newVersionCmd, not leaf:
// it needs no *App and must NOT load an env file -- it emits one. Not calling wireExec is
// the whole opt-out, and is also why --allow-command never appears here: nothing runs.
//
// It stays at the top level because it acts on no deployment: there is no cluster, no
// runtime and no existing env file, so every ancestor a noun would give it is dead weight.
//
// It takes no positional argument: --platform names the platform here as it does
// everywhere else in the tree, and bare `examples` gives the full schema.
func newExamplesCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "examples",
		Short: "Print a sample env file to start from",
		Long: "Bare `examples` prints the FULL annotated schema: every key the loader accepts and\n" +
			"the default each omitted one takes.\n" +
			"\n" +
			"--platform writes a minimal STARTER instead -- only the keys that platform cannot\n" +
			"default, declaring only its own section, so the file it writes needs no --platform\n" +
			"afterwards. It reads no env file and contacts nothing.\n" +
			"\n" +
			"Every secret is a CHANGE-ME placeholder the broker refuses until you edit it. Each\n" +
			"has a sibling *Env key naming a variable to read instead, which is what keeps a\n" +
			"committed env file secret-free.\n" +
			"\n" +
			"  solace-util examples --platform docker -o env/dev.yaml\n" +
			"  solace-util examples | less\n" +
			"  solace-util eg --platform kube -o env/prod.yaml && solace-util validate -e env/prod.yaml",
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		SilenceUsage:      true,
		SilenceErrors:     true,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runExample(app, app.PlatformFlag)
		},
	}
	// --force is gone: an existing --out path is CONFIRMED rather than requiring a second
	// flag, which is how every other question in this tree is asked, and --no-prompt is
	// the one silencer. Overwriting an env file someone has already edited is the worst
	// accident this command can have, so it deserves the prompt more than generate does.
	addOutFlags(cmd, app)
	return cmd
}

// The bare-listing helper is gone with the positional argument it served. There is no
// list to print any more: bare `examples` emits the full annotated schema, which is a
// better answer to "what can this configure" than a list of four names was, and
// --platform's own completion offers the three starters.

// runExample resolves which template to emit and writes it.
//
// An empty platform means the full annotated schema -- the useful default, since someone
// who has not said which platform they are on is usually reading rather than starting a
// deployment. A platform value is expanded by config.ParsePlatform, at the moment this
// command reads its own input: the same rule every other abbreviation site follows, so
// `eg --platform dk` and `--platform dk` mean the one thing and there is no second
// vocabulary to declare.
//
// It deliberately does NOT go through resolvePlatform, which requires the value to name a
// section of a declared env file. There is no env file yet; that is the point.
func runExample(a *App, platform string) error {
	name := examples.FullName
	if platform != "" {
		p, err := config.ParsePlatform(platform)
		if err != nil {
			return asUsage(err)
		}
		name = string(p)
	}
	ex, err := examples.Get(name)
	if err != nil {
		return asUsage(err)
	}
	// emitOrWrite owns stdout-vs-file, the overwrite confirmation and the 0600 mode, so
	// this command and `generate` cannot drift on any of the three. The file is a template
	// now, but it is about to hold real credentials.
	if err := emitOrWrite(a, ex.Body, ex.Name+" example"); err != nil {
		return err
	}
	if a.out != "" {
		step("replace every CHANGE-ME before running anything against it")
	}
	return nil
}

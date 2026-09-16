package cli

import (
	"os"

	"github.com/spf13/cobra"

	"solace/internal/config"
	"solace/internal/convert"
)

// newConvertCmd builds `solace-util convert`, the migration aid from the pre-Go bash
// env format to the YAML env file every other command reads. It sits at the root
// rather than under a platform because it loads no config of its own: the file
// it reads is the argument, not -e/--env, so the app context stays unused.
func newConvertCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert <bash-env-file>",
		Short: "Convert a legacy bash env file into a YAML env file",
		Long: "Convert a legacy bash env file -- the pre-Go format sourced by bash/000-env.sh --\n" +
			"into the YAML env file this CLI reads.\n\n" +
			"The target platform section is detected from the variables present; pass\n" +
			"--platform to choose it yourself. Variables with no YAML equivalent are\n" +
			"reported on stderr rather than dropped silently.\n\n" +
			"The output carries every secret from the source file verbatim, so treat it\n" +
			"like the source: write it with -o rather than through a shared terminal, and\n" +
			"never commit it.\n\n" +
			"  solace-util convert bash/env/prod -o prod.yaml\n" +
			"  solace-util convert bash/env/prod --platform podman -o prod.yaml\n" +
			"  solace-util check deploy -e prod.yaml",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, args []string) error {
			return runConvert(app, args[0], app.PlatformFlag)
		},
	}
	addOutFlags(cmd, app)
	return cmd
}

// runConvert reads --platform through the same parser the rest of the tree uses,
// so one word means one thing everywhere: it names a platform, in canonical or
// abbreviated form. What it SELECTS still differs by necessity -- here it is the
// section to write into a new file, elsewhere the section to read from an
// existing one -- but there is no second spelling to learn. Empty still means
// detect, which for a bash source is a question about its variable names
// (internal/convert), not about YAML sections.
func runConvert(a *App, src, platform string) error {
	p, err := config.ParsePlatform(platform)
	if err != nil {
		return asUsage(err)
	}
	raw, err := os.ReadFile(src)
	if err != nil {
		// The operator chose this path, so a bad one is exit 2 -- the same reading
		// import-config's own source-file failure takes. No retry and no different
		// broker state would help.
		return usagef("read bash env file %q: %w", src, err)
	}
	res, err := convert.Convert(raw, src, p)
	if err != nil {
		return err
	}
	for _, w := range res.Warnings {
		warn("%s", w)
	}
	step("converted %s for platform %s", src, res.Platform)
	// emitOrWrite owns stdout-vs-file, the overwrite confirmation and the 0600 mode. The
	// converted file carries the same secrets as the source, which is why the mode matters
	// and why replacing one is worth confirming.
	if a.out == "" {
		// Said BEFORE the write on the path where it matters most: stdout is a
		// terminal, a pipe or a log, none of which is 0600.
		warn("the converted file carries every secret from %s verbatim; prefer -o <file> to a shared terminal", src)
	}
	if err := emitOrWrite(a, res.YAML, "converted env file"); err != nil {
		return err
	}
	if a.out != "" {
		step("review it before use; it carries the secrets from %s verbatim", src)
	}
	return nil
}

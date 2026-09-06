package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"solace/internal/config"
	"solace/internal/examples"
	"solace/internal/output"
)

// newExamplesCmd builds `solace-util examples`, the only way to get a starting env
// file out of a binary with no checkout beside it. A bare literal like
// newVersionCmd, not leaf: it needs no *App and must NOT load an env file -- it
// emits one. Not calling wireExec is the whole opt-out, and is also why
// --allow-command never appears here: nothing runs.
//
// The one noun at the top level. It sits beside convert rather than under a verb
// because it acts on no deployment: there is no cluster, no runtime, and no
// platform to resolve, so every ancestor a verb would give it is dead weight.
func newExamplesCmd() *cobra.Command {
	var (
		out   string
		force bool
	)
	cmd := &cobra.Command{
		Use:   "examples [name]",
		Short: "Print a sample env file to start from",
		Long: "[name] is " + examples.List() + ". The three platform names give a minimal\n" +
			"standalone env file carrying only the keys that platform cannot default, and\n" +
			"declaring only its own section -- so the file it writes needs no --platform.\n" +
			"`full` is the complete annotated schema: every key the loader accepts, the\n" +
			"default each omitted one takes, and all three platform sections at once.\n" +
			"Omit [name] to list them.\n\n" +
			"Nothing here reads an env file or contacts anything, so -e/--env is ignored\n" +
			"and no cluster or container runtime has to exist yet.\n\n" +
			"Every secret in the output is a CHANGE-ME placeholder: an unedited file is\n" +
			"refused by the broker rather than deployed with a password that is public in\n" +
			"this repo. Each one has a sibling *Env key naming a variable to read instead,\n" +
			"which is what keeps a committed env file secret-free.\n\n" +
			"  solace-util examples docker -o env/dev.yaml\n" +
			"  solace-util examples full | less\n" +
			"  solace-util eg kubernetes -o env/prod.yaml && solace-util check deploy -e env/prod.yaml",
		Args:          cobra.MaximumNArgs(1),
		ValidArgs:     examples.Names(),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 0 {
				listExamples()
				return nil
			}
			return runExample(args[0], out, force)
		},
	}
	cmd.Flags().StringVarP(&out, "out", "o", "", "write the env file here instead of stdout")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite the --out file if it already exists")
	return cmd
}

// listExamples reports what the command can emit. A report, so stdout: `examples
// > list.txt` should capture the names.
func listExamples() {
	rows := make([]output.KV, 0, len(examples.All()))
	for _, e := range examples.All() {
		rows = append(rows, output.KV{Key: e.Name, Value: e.Short})
	}
	r := output.New(os.Stdout)
	r.Section("env file examples")
	r.KVBlock(rows)
}

// runExample resolves the name and emits the template. A platform abbreviation is
// expanded here, at the moment this command reads its own input, by the same
// parser --platform uses -- so `eg dk` and `--platform dk` mean the one thing,
// and there is no second vocabulary to declare. `full` is not a platform, so an
// unknown word falls through to the examples package's own error, which names
// every alternative.
func runExample(name, out string, force bool) error {
	ex, err := examples.Get(name)
	if err != nil {
		p, perr := config.ParsePlatform(name)
		if perr != nil || p == "" {
			return err
		}
		if ex, err = examples.Get(string(p)); err != nil {
			return err
		}
	}
	if out == "" {
		return emit(ex.Body)
	}
	if _, err := os.Stat(out); err == nil && !force {
		return fmt.Errorf("refusing to overwrite %q: pass --force to replace it, or choose another --out path", out)
	}
	// 0o600: the file is a template now, but it is about to hold real credentials.
	if err := os.WriteFile(out, ex.Body, 0o600); err != nil {
		return fmt.Errorf("write %q: %w", out, err)
	}
	step("wrote the %s example to %s", ex.Name, out)
	step("replace every CHANGE-ME before running anything against it")
	return nil
}

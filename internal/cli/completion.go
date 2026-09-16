package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"solace/internal/config"
)

// newCompletionCmd builds `solace-util auto-complete <shell>`. Cobra adds a command by
// this name on its own, but only from inside Execute -- and the command reference
// is rendered straight from newRootCmd, so a command that shipped and worked was
// missing from the generated docs. Registering it here makes it a command like any
// other: it lands in docs/commands.md, it is testable through the tree, and the
// help text is ours. That last part matters because writeCommand copies Long into
// the reference verbatim, and cobra's stock text carries markdown headings.
func newCompletionCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "auto-complete",
		Short: "Print the shell auto-completion script for solace-util",
		Long: "Print a shell's completion script on stdout. Load it to complete commands and\n" +
			"flags, plus the values they take: env files for -e/--env, primary|backup|monitor\n" +
			"for --pod, platform names for --platform, and directories for --base-dir and\n" +
			"--dir.\n\n" +
			"To load it for every new shell:\n\n" +
			"  bash        solace-util auto-complete bash > /etc/bash_completion.d/solace-util\n" +
			"  zsh         solace-util auto-complete zsh > \"${fpath[1]}/_solace-util\"\n" +
			"  fish        solace-util auto-complete fish > ~/.config/fish/completions/solace-util.fish\n" +
			"  powershell  solace-util auto-complete powershell > $HOME\\solace-util.ps1\n" +
			"              then add  . $HOME\\solace-util.ps1  to $PROFILE\n\n" +
			"Each shell's own help has the one-liner for loading into the CURRENT shell\n" +
			"instead, and the prerequisites where a shell has any.\n\n" +
			"Completion never reads the env file, so it stays inert -- a TAB press cannot\n" +
			"parse config or run anything.",
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		// A parent with no RunE is not Runnable, and cobra answers a non-runnable
		// command by printing help to STDOUT and exiting 0 -- so `solace-util completion
		// tcsh > solace-util.ps1` would write the help text into the profile script and
		// report success. Being runnable is what lets NoArgs reject the unknown shell
		// instead, and makes the bare command say what to pass. `--help` is unaffected:
		// the help flag short-circuits ahead of both.
		RunE: func(cmd *cobra.Command, _ []string) error {
			return fmt.Errorf("%s needs a shell: bash, zsh, fish, or powershell", cmd.CommandPath())
		},
	}
	c.AddCommand(
		completionShell("bash",
			"Load into the current shell:\n\n"+
				"  source <(solace-util auto-complete bash)\n\n"+
				"Load for every session. With the bash-completion package installed, write it\n"+
				"where that package looks:\n\n"+
				"  solace-util auto-complete bash > /etc/bash_completion.d/solace-util\n\n"+
				"Without that package there is no such directory, so source it from ~/.bashrc\n"+
				"instead -- this script needs nothing but bash itself:\n\n"+
				"  echo 'source <(solace-util auto-complete bash)' >> ~/.bashrc",
			func(root *cobra.Command, w io.Writer, desc bool) error {
				if err := root.GenBashCompletionV2(w, desc); err != nil {
					return err
				}
				_, err := io.WriteString(w, bashInitFallback+bashFiledirFallback)
				return err
			}),
		completionShell("zsh",
			"Load into the current shell:\n\n"+
				"  source <(solace-util auto-complete zsh)\n\n"+
				"Load for every session (compinit must be enabled in ~/.zshrc):\n\n"+
				"  solace-util auto-complete zsh > \"${fpath[1]}/_solace-util\"",
			func(root *cobra.Command, w io.Writer, desc bool) error {
				if !desc {
					return root.GenZshCompletionNoDesc(w)
				}
				return root.GenZshCompletion(w)
			}),
		completionShell("fish",
			"Load into the current shell:\n\n"+
				"  solace-util auto-complete fish | source\n\n"+
				"Load for every session:\n\n"+
				"  solace-util auto-complete fish > ~/.config/fish/completions/solace-util.fish",
			func(root *cobra.Command, w io.Writer, desc bool) error {
				return root.GenFishCompletion(w, desc)
			}),
		completionShell("powershell",
			"Load into the current shell:\n\n"+
				"  solace-util auto-complete powershell | Out-String | Invoke-Expression\n\n"+
				"Load for every session. Write the script once, then dot-source it from your\n"+
				"profile -- generating it once is what keeps shell start-up fast, since the\n"+
				"alternative runs this binary on every new shell:\n\n"+
				"  solace-util auto-complete powershell > $HOME\\solace-util.ps1\n"+
				"  Add-Content $PROFILE '. $HOME\\solace-util.ps1'\n\n"+
				"If $PROFILE does not exist yet, create it first:\n\n"+
				"  New-Item -ItemType File -Force $PROFILE",
			func(root *cobra.Command, w io.Writer, desc bool) error {
				if !desc {
					return root.GenPowerShellCompletion(w)
				}
				return root.GenPowerShellCompletionWithDesc(w)
			}),
	)
	return c
}

// completionShell builds one shell's generator. gen is handed the include-descriptions
// decision rather than reading the flag itself, which keeps the flag variable scoped
// to the command that declares it.
func completionShell(shell, long string, gen func(root *cobra.Command, w io.Writer, desc bool) error) *cobra.Command {
	var noDesc bool
	c := &cobra.Command{
		Use:               shell,
		Short:             "Print the " + shell + " completion script",
		Long:              long,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// os.Stdout, like emit: the script is this command's only stdout.
			return gen(cmd.Root(), os.Stdout, !noDesc)
		},
	}
	c.Flags().BoolVar(&noDesc, "no-descriptions", false, "omit the descriptions shown beside each completion")
	return c
}

// bashInitFallback is appended to cobra's bash script, and is the whole reason
// `auto-complete bash` works on a host with no bash-completion package.
//
// cobra's script asks for that package's `_init_completion` and falls back to its own
// `__solace-util_init_completion` when it is absent -- but that fallback is itself only a
// call to `_get_comp_words_by_ref`, which the package also owns. So the branch written for
// a host without the package is the one branch that cannot run there: every TAB press
// fails with `_get_comp_words_by_ref: command not found` and completes nothing.
//
// Redefining the function AFTER cobra's copy is what overrides it -- a later definition
// wins in bash, which is why this cannot be a prologue. It takes over one
// `__solace-util_*` name and defines no `_init_completion` or `_get_comp_words_by_ref` of
// its own, so a shell that HAS the package is unaffected in both directions: nothing
// global is shadowed, and `__start_solace-util` still picks the real `_init_completion`,
// which it tests for at completion time rather than when the script is sourced.
//
// The `-n <chars>` rejoining is why cobra passes `-n =:` at all. readline splits
// COMP_WORDS on every character in COMP_WORDBREAKS, `=` and `:` among them, so
// `--platform=docker` arrives as three words; leaving them split is what makes
// `--platform=d` complete to nothing. cur/prev/words/cword are the locals
// `__start_solace-util` declares, which is what lets a callee assign them.
const bashInitFallback = `
# The bash-completion package owns _get_comp_words_by_ref, which the fallback above
# still calls, so this self-contained replacement takes over. With that package
# installed, __start_solace-util prefers its _init_completion and never reaches here.
__solace-util_init_completion()
{
    COMPREPLY=()
    local _excl=""
    while (($#)); do
        case $1 in
            -n) shift; _excl=${1-}; shift ;;
            *)  shift ;;
        esac
    done

    # Rejoin what readline split on the excluded characters, so ` + "`--flag=value`" + `
    # and a value carrying a colon each arrive as the single word they were typed as.
    words=(); cword=0
    local _i _w _last
    for ((_i = 0; _i < ${#COMP_WORDS[@]}; _i++)); do
        _w=${COMP_WORDS[_i]}
        if ((_i > 0)) && [[ -n $_excl ]] && [[ $_w == [$_excl] || ${COMP_WORDS[_i-1]} == *[$_excl] ]]; then
            _last=$(( ${#words[@]} - 1 )); words[_last]+=$_w
        else
            words+=("$_w")
        fi
        ((_i == COMP_CWORD)) && cword=$(( ${#words[@]} - 1 ))
    done

    cur=${words[cword]}
    if ((cword > 0)); then prev=${words[cword-1]}; else prev=""; fi
}
`

// bashFiledirFallback is the second half of the same problem, and the one a reader is
// most likely to miss: cobra's script answers ShellCompDirectiveFilterDirs by calling
// `_filedir -d`, which the bash-completion package also owns. That is the directive
// `completeDirs` returns, so without the package `--base-dir` and `broker copy into
// --dir` print `_filedir: command not found` and offer nothing.
//
// Unlike the init function this name is not ours, so it is defined only when it is
// missing: a host WITH bash-completion keeps that package's own `_filedir`, which
// handles quoting, tilde expansion and `cur` shapes this does not. The guard is what
// makes taking over a name outside our namespace safe -- the real one is never shadowed,
// and on a host without the package there is nothing to shadow.
//
// Only the `-d` form matters here, because FilterDirs is the one file directive this CLI
// emits; plain file completion never reaches `_filedir` at all, since cobra leaves
// COMPREPLY empty and lets the `complete -o default` registration fall back to readline.
// The other form is answered with files and directories rather than left to fail, since
// that is what a caller asking to filter by extension would rather have.
//
// readarray -O appends without disturbing what the caller already collected, and keeps a
// name containing a space or a newline in one piece where word splitting would not.
const bashFiledirFallback = `
# The bash-completion package also owns _filedir, which the directory directive above
# calls. Define it only when absent: with the package installed its own version handles
# quoting and tilde expansion that this does not, and must not be shadowed.
if ! declare -F _filedir >/dev/null 2>&1; then
    _filedir()
    {
        if [[ $(type -t compopt) == builtin ]]; then
            compopt -o filenames
        fi
        if [[ ${1-} == -d ]]; then
            readarray -t -O "${#COMPREPLY[@]}" COMPREPLY < <(compgen -d -- "${cur-}")
        else
            readarray -t -O "${#COMPREPLY[@]}" COMPREPLY < <(compgen -f -- "${cur-}")
        fi
    }
fi
`

// completeEnvFiles offers the env files -e/--env can name. It mirrors
// config.ResolveEnvPath: the base dir first, then <base-dir>/env, bare names only,
// first one wins -- so a base-dir file shadowing the env/ copy of the same name is
// offered once, the way it resolves. A value that already carries a directory is
// used verbatim by ResolveEnvPath, so that case hands back to the shell's own file
// completion instead of guessing. The .yaml/.yml filter is a suggestion filter and
// not validation: ResolveEnvPath still accepts any regular file typed out in full.
func completeEnvFiles(app *App) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if strings.ContainsAny(toComplete, `/\`) {
			return nil, cobra.ShellCompDirectiveDefault
		}
		base := app.BaseDir
		if base == "" {
			base = "."
		}
		var names []string
		seen := map[string]bool{}
		for _, dir := range []string{base, filepath.Join(base, "env")} {
			entries, err := os.ReadDir(dir)
			if err != nil {
				continue // a missing env/ is the normal case, and a TAB press has nowhere to report it
			}
			for _, e := range entries {
				name := e.Name()
				if e.IsDir() || seen[name] || !isEnvFileName(name) || !strings.HasPrefix(name, toComplete) {
					continue
				}
				seen[name] = true
				names = append(names, name)
			}
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	}
}

// isEnvFileName reports whether name looks like a YAML env file.
func isEnvFileName(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".yaml", ".yml":
		return true
	}
	return false
}

// completeRoles completes a role value -- a [role] positional or --pod -- from the
// long names ParseRole accepts.
func completeRoles(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return matching(config.RoleNames(), toComplete), cobra.ShellCompDirectiveNoFileComp
}

// completePlatforms completes --platform. Only the canonical names are offered:
// the kube/dk/pm abbreviations exist to save typing something you already know,
// which is precisely what a completion removes the need for, and suggesting both
// spellings would put two names for one platform in front of the user. The empty
// value is left out too -- "detect it from the env file" is what omitting the
// flag already gives you.
func completePlatforms(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	all := make([]string, 0, len(config.Platforms()))
	for _, p := range config.Platforms() {
		all = append(all, string(p))
	}
	return matching(all, toComplete), cobra.ShellCompDirectiveNoFileComp
}

// completeDirs completes a flag whose value is a directory.
func completeDirs(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveFilterDirs
}

// matching keeps the candidates the partial word could still become. Cobra filters
// ValidArgs this way but leaves a completion function to filter its own.
func matching(candidates []string, toComplete string) []string {
	var out []string
	for _, c := range candidates {
		if strings.HasPrefix(c, toComplete) {
			out = append(out, c)
		}
	}
	return out
}

// registerFlagCompletion wires a completion function onto a flag. The only way
// RegisterFlagCompletionFunc fails is a flag name that was never declared -- a
// wiring bug, not a runtime condition -- and the tree builders have no error to
// return, so swallowing it would ship a flag that silently completes nothing.
// Every test that builds the tree runs this, so the panic cannot reach a user.
func registerFlagCompletion(c *cobra.Command, name string,
	fn func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective)) {
	if err := c.RegisterFlagCompletionFunc(name, fn); err != nil {
		panic(fmt.Sprintf("completion for --%s on %q: %v", name, c.CommandPath(), err))
	}
}

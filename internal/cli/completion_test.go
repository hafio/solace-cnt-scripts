package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// runComplete drives cobra's hidden __complete endpoint -- the same request a
// loaded completion script makes on every TAB press. runRoot cannot be reused:
// it points cobra's own writer at io.Discard, and __complete writes the
// candidates there rather than to os.Stdout. The last line is always the
// ":<directive>" marker.
func runComplete(t *testing.T, args ...string) (candidates []string, directive string) {
	t.Helper()
	var buf bytes.Buffer
	root := newRootCmd(&App{})
	root.SetArgs(append([]string{cobra.ShellCompRequestCmd}, args...))
	root.SetOut(&buf)
	root.SetErr(io.Discard)
	if err := root.Execute(); err != nil {
		t.Fatalf("__complete %v: %v", args, err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) == 0 {
		t.Fatalf("__complete %v produced no output, want at least a directive line", args)
	}
	return lines[:len(lines)-1], lines[len(lines)-1]
}

// wantDirective renders the ":<n>" marker cobra prints last, so the tests name the
// directive instead of a magic number.
func wantDirective(d cobra.ShellCompDirective) string { return fmt.Sprintf(":%d", d) }

// TestCompletionScriptsGenerate: every advertised shell emits its own script. The
// marker is the line that actually binds the completer to the `solace-util` command, so
// a script that generated but wired up nothing would still fail.
func TestCompletionScriptsGenerate(t *testing.T) {
	for _, tc := range []struct{ shell, marker string }{
		{"bash", "__start_solace-util"},
		{"zsh", "#compdef solace-util"},
		{"fish", "complete -c solace-util"},
		{"powershell", "Register-ArgumentCompleter -CommandName 'solace-util'"},
	} {
		t.Run(tc.shell, func(t *testing.T) {
			out, err := runRoot(t, []string{"auto-complete", tc.shell})
			if err != nil {
				t.Fatalf("completion %s: %v", tc.shell, err)
			}
			if !strings.Contains(out, tc.marker) {
				t.Errorf("completion %s output missing %q", tc.shell, tc.marker)
			}
		})
	}
}

// TestBashScriptDoesNotNeedBashCompletion: the emitted bash script completes on a host
// with no bash-completion package. Cobra's script calls that package in two places and
// neither survives its absence -- its own "minimal" fallback init is just
// `_get_comp_words_by_ref`, and the directory directive `completeDirs` returns is
// answered with `_filedir -d`. Unpatched, a TAB press there dies with a
// command-not-found and offers nothing.
//
// The init assertion is positional on purpose: a later definition is what wins in bash,
// so what has to hold is that the LAST `__solace-util_init_completion` in the file needs
// nothing the package provides. Checking only that our text is present somewhere would
// pass even if it were emitted first, where cobra's copy would override it instead.
//
// `_filedir` is checked differently because the name is not ours: it must be defined
// behind a `declare -F` guard, so a host that HAS bash-completion keeps that package's
// own version rather than being handed this cut-down one.
func TestBashScriptDoesNotNeedBashCompletion(t *testing.T) {
	const initFn = "__solace-util_init_completion()"
	const owned = "_get_comp_words_by_ref"

	out, err := runRoot(t, []string{"auto-complete", "bash"})
	if err != nil {
		t.Fatalf("completion bash: %v", err)
	}
	if n := strings.Count(out, initFn); n < 2 {
		t.Fatalf("the bash script defines %s %d time(s): cobra's copy has to be FOLLOWED by "+
			"ours, which is what overrides it", initFn, n)
	}
	if last := out[strings.LastIndex(out, initFn):]; strings.Contains(last, owned) {
		t.Errorf("the definition of %s that wins still calls %s, which only the bash-completion "+
			"package provides:\n%s", initFn, owned, last)
	}
	if !strings.Contains(out, "_filedir()") {
		t.Error("the bash script calls _filedir for the directory directive but never defines " +
			"it, so --base-dir and `broker copy into --dir` complete nothing without the " +
			"bash-completion package")
	}
	if !strings.Contains(out, "declare -F _filedir") {
		t.Error("_filedir is defined unguarded: the name belongs to bash-completion, and a host " +
			"that has that package must keep its version rather than this cut-down one")
	}

	// Scoped to bash: the other three generators are cobra's, untouched.
	for _, shell := range []string{"zsh", "fish", "powershell"} {
		other, err := runRoot(t, []string{"auto-complete", shell})
		if err != nil {
			t.Fatalf("completion %s: %v", shell, err)
		}
		for _, marker := range []string{initFn, "_filedir()"} {
			if strings.Contains(other, marker) {
				t.Errorf("the %s script carries bash's %s fallback", shell, marker)
			}
		}
	}
}

// TestBashInitFallbackRejoinsSplitWords: the replacement init function fills the four
// variables cobra's script reads, and rejoins the words readline split apart.
//
// That rejoining is the whole reason cobra passes `-n =:`. COMP_WORDBREAKS contains both
// characters, so `--platform=docker` reaches the completer as three words; a fallback
// that left them split still completes `--platform docker` but offers nothing at all for
// `--platform=d`, which is why asserting on the emitted text is not enough -- the text
// says nothing about whether the shell agrees with it.
//
// It calls the function directly rather than pressing TAB, so the result does not depend
// on whether the host has bash-completion (__start_solace-util would prefer the real
// _init_completion) nor on a built binary named solace-util being on PATH for the
// `${words[0]} __complete` round trip. Gated on bash being present rather than on GOOS:
// where a Windows runner has Git Bash, this is worth running there too.
// driveBash runs the generated script plus a driving snippet under bash, on STDIN
// rather than as a `-c` argument. The script is multi-line and full of quotes, and
// as one argv element Go quotes it by MSVCRT rules while Git Bash re-parses by MSYS
// rules -- the two disagree and the Windows runner sees an unterminated quote in a
// script that is valid. stdin has no such round-trip.
//
// Returns stdout alone so the exact comparisons below are not polluted, and folds
// stderr into the error, which is where bash reports a syntax error.
func driveBash(t *testing.T, bash, script, drive, dir string) ([]byte, error) {
	t.Helper()
	var stderr bytes.Buffer
	cmd := exec.Command(bash, "--norc", "-s")
	cmd.Stdin = strings.NewReader(script + drive)
	cmd.Stderr = &stderr
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return out, fmt.Errorf("%w\n%s", err, stderr.String())
	}
	return out, nil
}

func TestBashInitFallbackRejoinsSplitWords(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skipf("no bash on PATH, so the generated script cannot be driven here: %v", err)
	}
	script, err := runRoot(t, []string{"auto-complete", "bash"})
	if err != nil {
		t.Fatalf("completion bash: %v", err)
	}

	// What bash hands a completion function for `solace-util broker deploy --platform=d`.
	const drive = `
COMP_WORDS=(solace-util broker deploy --platform = d)
COMP_CWORD=5
cur=; prev=; words=(); cword=
__solace-util_init_completion -n =:
printf '%s|%s|%s|%s' "${words[*]}" "$cword" "$cur" "$prev"
`
	out, err := driveBash(t, bash, script, drive, "")
	if err != nil {
		t.Fatalf("driving the generated script: %v\n%s", err, out)
	}
	const want = "solace-util broker deploy --platform=d|3|--platform=d|deploy"
	if got := string(out); got != want {
		t.Errorf("the init fallback did not rejoin the split words\ngot  %s\nwant %s", got, want)
	}
}

// TestBashFiledirFallbackCompletesDirsOnly: the `_filedir` stand-in answers `-d` with
// directories and nothing else. That is the whole of what `--base-dir` and `broker copy
// into --dir` ask for, and offering the files beside them would be a worse answer than
// the one a host WITH bash-completion gets.
//
// The fixture holds a directory whose name contains a space, which is what an unquoted
// `$(compgen -d)` splits in half and the reason the fallback reads with readarray, and
// two files, which are what the `-d` has to exclude.
func TestBashFiledirFallbackCompletesDirsOnly(t *testing.T) {
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skipf("no bash on PATH, so the generated script cannot be driven here: %v", err)
	}
	script, err := runRoot(t, []string{"auto-complete", "bash"})
	if err != nil {
		t.Fatalf("completion bash: %v", err)
	}

	dir := t.TempDir()
	for _, sub := range []string{"alpha", "gamma space"} {
		if err := os.Mkdir(filepath.Join(dir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"afile.txt", "bfile.yaml"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// compopt's output is dropped because it refuses outside a real completion, which
	// this is not. What lands in COMPREPLY is the subject here.
	const drive = `
cur=""
COMPREPLY=()
_filedir -d 2>/dev/null
printf '%s\n' "${COMPREPLY[@]}"
`
	out, err := driveBash(t, bash, script, drive, dir)
	if err != nil {
		t.Fatalf("driving _filedir in the generated script: %v", err)
	}
	const want = "alpha\ngamma space\n"
	if got := string(out); got != want {
		t.Errorf("_filedir -d should offer the directories and only those\ngot  %q\nwant %q", got, want)
	}
}

// TestCompletionNoDescriptions: --no-descriptions is honoured on every shell. It
// switches the request the generated script makes from __complete to
// __completeNoDesc, which is the only externally visible difference.
func TestCompletionNoDescriptions(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			with, err := runRoot(t, []string{"auto-complete", shell, "--no-descriptions"})
			if err != nil {
				t.Fatalf("completion %s --no-descriptions: %v", shell, err)
			}
			if !strings.Contains(with, cobra.ShellCompNoDescRequestCmd) {
				t.Errorf("completion %s --no-descriptions does not request %s", shell, cobra.ShellCompNoDescRequestCmd)
			}
			without, err := runRoot(t, []string{"auto-complete", shell})
			if err != nil {
				t.Fatalf("completion %s: %v", shell, err)
			}
			if strings.Contains(without, cobra.ShellCompNoDescRequestCmd) {
				t.Errorf("completion %s requests %s without the flag", shell, cobra.ShellCompNoDescRequestCmd)
			}
		})
	}
}

// TestCompletionNeedsAShell: an unsupported shell, or none at all, fails loud with
// nothing on stdout. This is why the parent carries a RunE: cobra answers a
// non-runnable command by printing help to stdout and exiting 0, which would put
// the help text into `solace-util completion tcsh > solace-util.ps1` and call it a success.
func TestCompletionNeedsAShell(t *testing.T) {
	for _, args := range [][]string{{"auto-complete", "tcsh"}, {"auto-complete"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			out, err := runRoot(t, args)
			if err == nil {
				t.Fatalf("%v err = nil, want a loud failure naming the shells", args)
			}
			if out != "" {
				t.Errorf("%v wrote %q to stdout, want nothing", args, out)
			}
		})
	}
}

// TestCompletionHelpStillWorks: --help short-circuits ahead of the RunE above, so
// asking how to use the command is not itself an error.
func TestCompletionHelpStillWorks(t *testing.T) {
	if _, err := runRoot(t, []string{"auto-complete", "--help"}); err != nil {
		t.Errorf("completion --help err = %v, want nil", err)
	}
}

// TestEnvFlagCompletesEnvFiles: -e is completed from the two directories
// config.ResolveEnvPath searches, by bare name. The base-dir copy of a name that
// also exists under env/ is offered once, matching the shadowing the resolver
// applies, and a non-YAML file is not suggested.
func TestEnvFlagCompletesEnvFiles(t *testing.T) {
	base := t.TempDir()
	write := func(rel string) {
		t.Helper()
		p := filepath.Join(base, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", rel, err)
		}
		if err := os.WriteFile(p, []byte("redundancy:\n  enabled: false\n"), 0o600); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	write("prod.yaml")
	write("env/prod.yaml") // shadowed by the base-dir copy above
	write("env/dev.yml")
	write("notes.txt")

	got, directive := runComplete(t, "broker", "status", "--base-dir", base, "-e", "")
	want := []string{"prod.yaml", "dev.yml"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("-e completions = %v, want %v (base dir first, deduped, YAML only)", got, want)
	}
	if directive != wantDirective(cobra.ShellCompDirectiveNoFileComp) {
		t.Errorf("-e directive = %s, want no-file-completion", directive)
	}
}

// TestEnvFlagPrefixFilters: a partial name narrows the suggestions rather than
// dumping every env file back at the shell.
func TestEnvFlagPrefixFilters(t *testing.T) {
	base := t.TempDir()
	for _, name := range []string{"prod.yaml", "preprod.yaml", "dev.yaml"} {
		if err := os.WriteFile(filepath.Join(base, name), []byte("redundancy:\n  enabled: false\n"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	got, _ := runComplete(t, "broker", "status", "--base-dir", base, "-e", "pr")
	want := []string{"preprod.yaml", "prod.yaml"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("-e pr completions = %v, want %v", got, want)
	}
}

// TestEnvFlagWithPathDefersToShell: a value carrying a directory is used verbatim
// by ResolveEnvPath, so completion hands back to the shell's own file completion
// instead of offering bare names that would resolve somewhere else.
func TestEnvFlagWithPathDefersToShell(t *testing.T) {
	base := t.TempDir()
	if err := os.WriteFile(filepath.Join(base, "prod.yaml"), []byte("redundancy:\n  enabled: false\n"), 0o600); err != nil {
		t.Fatalf("write prod.yaml: %v", err)
	}
	got, directive := runComplete(t, "broker", "status", "--base-dir", base, "-e", "env/")
	if len(got) != 0 {
		t.Errorf("-e env/ completions = %v, want none", got)
	}
	if directive != wantDirective(cobra.ShellCompDirectiveDefault) {
		t.Errorf("-e env/ directive = %s, want the shell's default file completion", directive)
	}
}

// TestRoleArgsComplete is gone: no command takes a [role] positional any more.
//
// The role is --pod everywhere now, on both readings -- which pod to act on for a
// Kubernetes command, which node this host is for a container one -- so the completion
// that used to teach the positional lives in TestPodFlagCompletesRoles below. That is the
// whole convention: one flag, one completer, registered next to where the flag is declared.

// TestExampleArgsComplete is gone with the positional it completed. `examples` selects its
// template with --platform now, and TestPlatformFlagCompletes already pins that completer --
// including the part this test was really protecting, that the abbreviations are accepted
// but never offered.

// TestPodFlagCompletesRoles: --pod names the same roles as the positionals, so it
// completes to the same set rather than to filenames. copy from/into always named
// a pod this way; cli, shell, logs broker, check semp-login, status broker and
// restart broker are the six that used to take a [role] positional instead (H2)
// and now offer the same set through --pod -- see TestRoleArgsComplete's doc
// comment for the commands that kept a real positional.
func TestPodFlagCompletesRoles(t *testing.T) {
	for _, path := range [][]string{
		{"broker", "copy", "into"},
		{"broker", "copy", "from"},
		{"broker", "cli"},
		{"broker", "shell"},
		{"broker", "logs"},
		{"broker", "perform", "semp-login-check"},
		{"broker", "status"},
		{"broker", "restart"},
	} {
		t.Run(strings.Join(path, " "), func(t *testing.T) {
			got, directive := runComplete(t, append(append([]string{}, path...), "--pod", "")...)
			want := []string{"primary", "backup", "monitor"}
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Errorf("--pod completions = %v, want %v", got, want)
			}
			if directive != wantDirective(cobra.ShellCompDirectiveNoFileComp) {
				t.Errorf("--pod directive = %s, want no-file-completion", directive)
			}
		})
	}
}

// TestPlatformFlagCompletes: --platform completes to the three canonical platform
// names, on root and on every command that inherits it -- convert included, since
// it no longer declares its own copy of the flag and reads the root one instead.
// The empty "detect" value is left out -- omitting the flag is how you ask for it.
//
// Neither the retired k8s spelling nor the kube/dk/pm abbreviations are offered: the
// abbreviations exist to save typing something you already know, which is what a
// completion already does -- offering both would put two names for one platform
// in front of the user.
func TestPlatformFlagCompletes(t *testing.T) {
	for _, path := range [][]string{
		{"broker", "status"},
		{"convert", "old-env"},
		// examples reads --platform too, and resolves it WITHOUT an env file. The completer
		// is the root one either way, which is the point: one flag, one vocabulary, whether
		// or not there is a file to check the answer against.
		{"examples"},
	} {
		t.Run(strings.Join(path, " "), func(t *testing.T) {
			got, directive := runComplete(t, append(path, "--platform", "")...)
			want := []string{"kubernetes", "docker", "podman"}
			if strings.Join(got, ",") != strings.Join(want, ",") {
				t.Errorf("--platform completions = %v, want %v", got, want)
			}
			if directive != wantDirective(cobra.ShellCompDirectiveNoFileComp) {
				t.Errorf("--platform directive = %s, want no-file-completion", directive)
			}
			for _, retired := range []string{"k8s", "k8", "kube", "dk", "pm"} {
				for _, c := range got {
					if c == retired {
						t.Errorf("--platform completions = %v, must not offer %q", got, retired)
					}
				}
			}
		})
	}
}

// TestDirFlagCompletesDirectories: --dir takes a directory, so it asks the shell
// to filter to directories rather than offering every file. The tree is flat now,
// so `copy into` is one command shared by every platform rather than a separate
// copy per platform subtree.
func TestDirFlagCompletesDirectories(t *testing.T) {
	_, directive := runComplete(t, "broker", "copy", "into", "--dir", "")
	if directive != wantDirective(cobra.ShellCompDirectiveFilterDirs) {
		t.Errorf("--dir directive = %s, want directory filtering", directive)
	}
}

// TestNoArgsLeafOffersNoFiles: a command built by leaf takes no arguments, so it
// offers nothing. Cobra's fallback is filename completion, which is what this
// stops -- the wrong suggestion on the majority of commands in the tree.
//
// The cases below deliberately include every no-arg command that needs a flag on
// top (remove broker/all, diagnostics): those used to hand-roll the literal leaf
// already builds, which is how they lost NoFileCompletions while the plain leaves
// kept it. They now go through leaf and attach the extra afterwards, so there is
// one definition of "takes no arguments" for the whole tree.
//
// The verb groups themselves (check, status, remove, ...) are deliberately
// absent: they carry no RunE and no ValidArgsFunction of their own, so completing
// after one offers its object subcommands' names -- the group's own semantics,
// not "takes no arguments". `deploy broker`/`deploy all` are absent for the same
// reason as TestRoleArgsComplete's comment: their [role] positional is offered by
// ValidArgs unconditionally, since completion cannot tell here whether this run
// will land on Kubernetes (which rejects the role) or a container host (which
// uses it).
func TestNoArgsLeafOffersNoFiles(t *testing.T) {
	for _, path := range [][]string{
		{"validate"},
		{"operator", "status"},
		{"broker", "remove"},
		{"broker", "perform", "gather-diagnostics"},
		{"operator", "logs"},
		{"broker", "configure", "domain-certs", "--remove"},
		{"version"},
	} {
		t.Run(strings.Join(path, " "), func(t *testing.T) {
			got, directive := runComplete(t, append(path, "")...)
			if len(got) != 0 {
				t.Errorf("completions = %v, want none", got)
			}
			if directive != wantDirective(cobra.ShellCompDirectiveNoFileComp) {
				t.Errorf("directive = %s, want no-file-completion", directive)
			}
		})
	}
}

// TestAllowCommandOffersNoFiles: --allow-command takes a bare binary name, never a
// path. Offering files would coach the mistake its own help text warns against.
// wireExec wires the flag the same way on every command that carries it, so one
// case built through leaf (check) and one built inline (deploy) are enough to
// prove the registration, rather than enumerating every runnable command.
func TestAllowCommandOffersNoFiles(t *testing.T) {
	for _, path := range [][]string{{"validate"}, {"broker", "deploy"}} {
		t.Run(strings.Join(path, " "), func(t *testing.T) {
			got, directive := runComplete(t, append(path, "--allow-command", "")...)
			if len(got) != 0 {
				t.Errorf("completions = %v, want none", got)
			}
			if directive != wantDirective(cobra.ShellCompDirectiveNoFileComp) {
				t.Errorf("directive = %s, want no-file-completion", directive)
			}
		})
	}
}

// TestFlagCompletionsRegistered is the drift gate: renaming a flag without moving
// its completion registration leaves the flag completing filenames again, which no
// other test here would notice because it fails silently at a TAB press.
func TestFlagCompletionsRegistered(t *testing.T) {
	root := newRootCmd(&App{})
	cases := []struct {
		path []string
		flag string
	}{
		{nil, "env"},
		{nil, "base-dir"},
		{nil, "platform"},
		{[]string{"broker", "status"}, "allow-command"},
		{[]string{"broker", "deploy"}, "allow-command"},
		{[]string{"broker", "cli"}, "pod"},
		{[]string{"broker", "copy", "from"}, "pod"},
		{[]string{"broker", "copy", "into"}, "pod"},
		{[]string{"broker", "copy", "into"}, "dir"},
		{[]string{"broker", "perform", "gather-diagnostics"}, "days"},
		{[]string{"broker", "logs"}, "tail"},
		{[]string{"broker", "logs"}, "since"},
		{[]string{"operator", "logs"}, "tail"},
		{[]string{"operator", "logs"}, "since"},
	}
	for _, tc := range cases {
		cmd := findCmd(t, root, tc.path...)
		if _, ok := cmd.GetFlagCompletionFunc(tc.flag); !ok {
			t.Errorf("%s: --%s has no completion function registered", cmd.CommandPath(), tc.flag)
		}
	}
}

// TestCompletionHelpKeepsTheLoadingInstructions pins the one thing this command exists to
// tell you, which has gone missing before.
//
// A completion script is useless on its own: its whole value is the line you paste to load
// it, and the line differs per shell in a way nobody remembers. So both halves have to be
// in the help -- the one-liner for the CURRENT shell, and the permanent form for every new
// shell -- and the parent has to carry the permanent form for all four, so `auto-complete`
// on its own is enough without drilling into each subcommand.
//
// Nothing else would notice their removal: every other completion test drives the
// generator and checks the SCRIPT, which is unaffected by the help text around it.
func TestCompletionHelpKeepsTheLoadingInstructions(t *testing.T) {
	root := newRootCmd(&App{})
	var comp *cobra.Command
	for _, c := range root.Commands() {
		if c.Name() == "auto-complete" {
			comp = c
		}
	}
	if comp == nil {
		t.Fatal("no auto-complete command in the tree")
	}

	// The parent names the permanent form for every shell it supports, so one help
	// screen is enough to set this up.
	for _, want := range []string{
		"/etc/bash_completion.d/solace-util",
		"_solace-util",
		"~/.config/fish/completions/solace-util.fish",
		"$PROFILE",
	} {
		if !strings.Contains(comp.Long, want) {
			t.Errorf("`auto-complete` help does not say how to load permanently for one of the\n"+
				"shells it supports: missing %q\n\n%s", want, comp.Long)
		}
	}

	// And each shell says both how to load it NOW and how to keep it.
	shells := 0
	for _, sub := range comp.Commands() {
		shells++
		if !strings.Contains(sub.Long, "current shell") {
			t.Errorf("`auto-complete %s` help does not show how to load into the current shell:\n%s",
				sub.Name(), sub.Long)
		}
		if !strings.Contains(sub.Long, "every session") {
			t.Errorf("`auto-complete %s` help does not show how to load for every session:\n%s",
				sub.Name(), sub.Long)
		}
		// A heading promising a command, with no command under it, is worse than no
		// heading: the powershell help said "source it from your profile" for a while
		// and never gave the line to put there.
		if !strings.Contains(sub.Long, "solace-util auto-complete "+sub.Name()) {
			t.Errorf("`auto-complete %s` help explains loading without showing the command:\n%s",
				sub.Name(), sub.Long)
		}
	}
	if shells != 4 {
		t.Errorf("auto-complete offers %d shells, want 4 (bash, zsh, fish, powershell)", shells)
	}
}

package cli

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"solace/internal/config"
)

// update regenerates docs/commands.md from the live command tree. The doc is a
// golden, so this test is also the generator -- the same arrangement the render
// and k8s packages use for their manifest goldens.
var update = flag.Bool("update", false, "regenerate docs/commands.md from the command tree")

// commandDocPath is anchored on the package dir, which is `go test`'s cwd.
const commandDocPath = "../../docs/commands.md"

// TestCommandDocs is the drift gate: a new command, a renamed flag, or an edited
// Short string fails `test` until the reference is regenerated. Nothing else in
// the suite covers Short text or the whole tree.
func TestCommandDocs(t *testing.T) {
	got := renderCommandDocs(newRootCmd(&App{}))

	if *update {
		if err := os.WriteFile(commandDocPath, got, 0o644); err != nil {
			t.Fatalf("write %s: %v", commandDocPath, err)
		}
		return
	}
	want, err := os.ReadFile(commandDocPath)
	if err != nil {
		t.Fatalf("read %s (regenerate: go test ./internal/cli -update): %v", commandDocPath, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s is stale -- regenerate with: go test ./internal/cli -update\n%s",
			commandDocPath, firstDiff(got, want))
	}
}

// firstDiff names the first line that differs, so the failure says what changed
// instead of dumping two whole documents.
func firstDiff(got, want []byte) string {
	g, w := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(g) || i < len(w); i++ {
		gl, wl := "", ""
		if i < len(g) {
			gl = g[i]
		}
		if i < len(w) {
			wl = w[i]
		}
		if gl != wl {
			return fmt.Sprintf("first difference at line %d:\n  generated: %q\n  committed: %q", i+1, gl, wl)
		}
	}
	return ""
}

// renderCommandDocs renders the whole command tree as markdown. It is the single
// source of truth for docs/commands.md: every fact in the reference is read off
// the cobra tree, so the doc cannot describe a command that does not exist.
func renderCommandDocs(root *cobra.Command) []byte {
	var b strings.Builder

	b.WriteString("# Command reference\n\n")
	b.WriteString("Every command `solace-util` exposes, with its arguments and flags.\n\n")
	b.WriteString("**Generated from the command tree -- do not edit by hand.** Regenerate after any\n")
	b.WriteString("command, flag, or description change:\n\n")
	b.WriteString("```\ngo test ./internal/cli -update\n```\n\n")
	b.WriteString("The `test` task fails while this file is stale, so it cannot drift from the code.\n\n")

	b.WriteString("## Tree\n\n```\n")
	walkCommands(root, 0, func(c *cobra.Command, depth int) {
		fmt.Fprintf(&b, "%s%s\n", strings.Repeat("  ", depth), c.Use)
	})
	b.WriteString("```\n\n")

	writeIndex(&b, root)
	writeReadingNotes(&b)

	b.WriteString("## Global flags\n\n")
	b.WriteString("Inherited by every command.\n\n")
	writeFlagTable(&b, root.PersistentFlags())

	b.WriteString("## Commands\n")
	walkCommands(root, 0, func(c *cobra.Command, _ int) {
		writeCommand(&b, c, c == root)
	})
	return []byte(b.String())
}

// writeIndex links every command to the section writeCommand emits for it. The
// tree above is the shape; this is the way in, because a fenced block cannot
// carry links. The label is what you type, so the index doubles as a list of
// every valid command line.
func writeIndex(b *strings.Builder, root *cobra.Command) {
	b.WriteString("## Index\n\n")
	prefix := root.Name() + " "
	walkCommands(root, 0, func(c *cobra.Command, depth int) {
		if depth == 0 {
			return
		}
		label := strings.TrimPrefix(c.CommandPath(), prefix)
		fmt.Fprintf(b, "%s- [%s](#%s)\n", strings.Repeat("  ", depth-1), label, anchor(c.CommandPath()))
	})
	b.WriteString("\n")
}

// anchor mirrors the id a markdown renderer derives from the `### <path>`
// heading writeCommand emits.
func anchor(path string) string {
	return strings.ReplaceAll(strings.ToLower(path), " ", "-")
}

// writeReadingNotes states once what would otherwise repeat on every command:
// the rules the whole tree obeys, and the vocabulary its positionals share. A
// per-command line cannot say these things, because they are facts about the CLI
// rather than about any one command.
func writeReadingNotes(b *strings.Builder) {
	b.WriteString("## Reading this reference\n\n")
	b.WriteString("- **Global flags are inherited by every command** and are listed once, under\n")
	b.WriteString("  [Global flags](#global-flags), rather than repeated in each command's flag table.\n")
	b.WriteString("- **`--allow-command` approves one extra binary** for the env file's platform command,\n")
	b.WriteString("  for that run only; it is repeatable and takes a bare name, never a path. It is listed\n")
	b.WriteString("  on every command that executes something, and refused with a named error by the ones\n")
	b.WriteString("  that only render.\n")
	b.WriteString("- **A verb that owns more than one kind of object never acts when run bare.** Its entry\n")
	b.WriteString("  below says \"Runs nothing on its own\" and names the objects instead; you pick one to\n")
	b.WriteString("  make it do anything. `remove` alone removes nothing.\n")
	b.WriteString("- **`[role]` is `primary`, `backup` or `monitor`**, and the letters `p`, `b`, `m` work\n")
	b.WriteString("  everywhere the long names do. On Kubernetes it picks which pod a command targets,\n")
	b.WriteString("  defaulting to the primary. On docker and podman there is one container per host, so it\n")
	b.WriteString("  instead names which host in the redundancy group this invocation runs on -- required\n")
	b.WriteString("  where the artifact is per-host, and detected from the host name where it may be\n")
	b.WriteString("  omitted. Passing one where it means nothing is refused with a named error.\n")
	b.WriteString("- **Abbreviations mean the same thing at every depth.** A command's short form is on its\n")
	b.WriteString("  \"Also available as:\" line, and it works under whichever verb the command sits below --\n")
	b.WriteString("  `br` and `op` ride under every verb that takes `broker` and `operator`. Every short\n")
	b.WriteString("  form the tool accepts, including the role letters and the `--platform` spellings, is\n")
	b.WriteString("  in [abbreviation.md](abbreviation.md).\n")
	b.WriteString("- **\"Applies to:\" is enforced, not advisory.** A command scoped to some platforms refuses\n")
	b.WriteString("  on the others with a named error rather than silently doing nothing; a command with no\n")
	b.WriteString("  \"Applies to:\" line works on all of them. The tree is the same shape everywhere, because\n")
	b.WriteString("  help and completion render it without reading an env file.\n\n")
}

// walkCommands visits the tree depth-first. cobra sorts Commands() by name, so
// the order is stable without sorting here.
func walkCommands(c *cobra.Command, depth int, fn func(*cobra.Command, int)) {
	fn(c, depth)
	for _, sub := range c.Commands() {
		if !sub.IsAvailableCommand() {
			continue
		}
		walkCommands(sub, depth+1, fn)
	}
}

// writeCommand renders one command. isRoot suppresses the flag table, because
// the root's local flags are the global ones already listed above.
func writeCommand(b *strings.Builder, c *cobra.Command, isRoot bool) {
	fmt.Fprintf(b, "\n### %s\n\n", c.CommandPath())
	if c.Short != "" {
		fmt.Fprintf(b, "%s\n\n", c.Short)
	}
	if c.Long != "" && c.Long != c.Short {
		fmt.Fprintf(b, "%s\n\n", c.Long)
	}
	fmt.Fprintf(b, "```\n%s\n```\n\n", c.UseLine())

	if s := argumentLine(c); s != "" {
		fmt.Fprintf(b, "%s\n\n", s)
	}
	// group() is the only thing that sets groupAnnotation, and it is also what gives a
	// verb the RunE that refuses an unknown noun -- so the marker and the behaviour it
	// describes come from the same constructor and cannot drift apart.
	if subs := availableSubs(c); len(subs) > 0 {
		if c.Annotations[groupAnnotation] == "true" {
			fmt.Fprintf(b, "Runs nothing on its own. It names what `%s` can act on -- %s -- "+
				"and given a word it does not know it fails rather than reporting success.\n\n",
				c.CommandPath(), strings.Join(subs, ", "))
		} else {
			fmt.Fprintf(b, "Subcommands: %s\n\n", strings.Join(subs, ", "))
		}
	}
	if len(c.Aliases) > 0 {
		fmt.Fprintf(b, "Also available as: %s\n\n", strings.Join(c.Aliases, ", "))
	}
	// The tree is one shape on every platform, so applicability is a fact about a
	// command rather than something the reader can infer from where it sits. It is
	// read off the annotation that also drives the refusal, so the reference cannot
	// promise a platform the command would reject.
	if v, ok := c.Annotations[platformAnnotation]; ok && v != config.JoinPlatforms(config.Platforms()) {
		fmt.Fprintf(b, "Applies to: %s. On any other platform this command fails rather than doing nothing.\n\n", v)
	}
	// The same annotation that refuses --allow-command here, so the reference cannot
	// call a command harmless that would in fact run something.
	if _, ok := c.Annotations[renderAnnotation]; ok {
		b.WriteString("Renders to stdout and changes nothing: it runs no external command, so it needs " +
			"no cluster or runtime, runs no preflight, and refuses `--allow-command` -- there is " +
			"nothing here for it to approve.\n\n")
	}
	if !isRoot {
		writeFlagTable(b, c.NonInheritedFlags())
	}
}

// argumentLine describes a positional whose vocabulary the command already
// carries as ValidArgs. Every spelling is round-tripped through config.ParseRole,
// so the reference cannot offer one the parser would reject -- and a positional
// the tree does not describe gets no line rather than an invented one.
func argumentLine(c *cobra.Command) string {
	if len(c.ValidArgs) == 0 {
		return ""
	}
	forms := make([]string, 0, len(c.ValidArgs))
	for _, v := range c.ValidArgs {
		r, err := config.ParseRole(v)
		if err != nil {
			return ""
		}
		forms = append(forms, fmt.Sprintf("`%s` (`%s`)", v, r.Letter()))
	}
	name := "the argument"
	if f := strings.Fields(c.Use); len(f) > 1 {
		name = "`" + f[1] + "`"
	}
	return fmt.Sprintf("Arguments: %s is one of %s -- see [Reading this reference](#reading-this-reference).",
		name, strings.Join(forms, ", "))
}

func availableSubs(c *cobra.Command) []string {
	var out []string
	for _, sub := range c.Commands() {
		if sub.IsAvailableCommand() {
			out = append(out, "`"+sub.Name()+"`")
		}
	}
	return out
}

// writeFlagTable renders a flag set, or nothing when it is empty.
func writeFlagTable(b *strings.Builder, fs *pflag.FlagSet) {
	type row struct{ name, def, usage string }
	var rows []row
	fs.VisitAll(func(f *pflag.Flag) {
		if f.Hidden || f.Name == "help" {
			return
		}
		name := "`--" + f.Name + "`"
		if f.Shorthand != "" {
			name = "`-" + f.Shorthand + "`, " + name
		}
		def := "(none)"
		if f.DefValue != "" {
			def = "`" + f.DefValue + "`"
		}
		rows = append(rows, row{name, def, mdCell(f.Usage)})
	})
	if len(rows) == 0 {
		return
	}
	b.WriteString("| Flag | Default | Meaning |\n| --- | --- | --- |\n")
	for _, r := range rows {
		fmt.Fprintf(b, "| %s | %s | %s |\n", r.name, r.def, r.usage)
	}
	b.WriteString("\n")
}

// mdCell escapes the characters that would break a markdown table cell: a pipe
// ends the cell, and angle brackets are stripped as unknown HTML by renderers.
func mdCell(s string) string {
	return strings.NewReplacer("|", "\\|", "<", "&lt;", ">", "&gt;").Replace(s)
}

package cli

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"solace/internal/abbrev"
	"solace/internal/config"
)

// abbrevDocPath is anchored on the package dir, which is `go test`'s cwd.
const abbrevDocPath = "../../docs/abbreviation.md"

// TestAbbreviationDocs is the drift gate for the glossary. The short forms this
// tool accepts come from four places that never meet at run time -- the command
// set, the role set, the platform set and pflag's shorthands -- so a hand-written
// glossary would go stale at whichever of them changed first. This renders it from
// all four and fails while the committed file differs.
//
// It reuses commanddoc_test.go's -update flag, so the `regen` task already rewrites
// this file too: both goldens live in this package and one flag serves both.
func TestAbbreviationDocs(t *testing.T) {
	got := renderAbbrevDocs(newRootCmd(&App{}))

	if *update {
		if err := os.WriteFile(abbrevDocPath, got, 0o644); err != nil {
			t.Fatalf("write %s: %v", abbrevDocPath, err)
		}
		return
	}
	want, err := os.ReadFile(abbrevDocPath)
	if err != nil {
		t.Fatalf("read %s (regenerate: go test ./internal/cli -update): %v", abbrevDocPath, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s is stale -- regenerate with: go test ./internal/cli -update\n%s",
			abbrevDocPath, firstDiff(got, want))
	}
}

// TestFlagShorthandsAreConsistent is the one abbreviation rule abbrev cannot
// enforce, because pflag owns shorthand registration: pflag rejects a duplicate
// within one flag set, but nothing stops `-o` meaning --out on one command and
// --output on another. A letter that means two things is the failure the whole
// abbreviation scheme exists to avoid, so it is checked across the whole tree.
func TestFlagShorthandsAreConsistent(t *testing.T) {
	owner := map[string]string{}
	for _, f := range treeShorthands(newRootCmd(&App{})) {
		if prev, seen := owner[f.short]; seen && prev != f.name {
			t.Errorf("-%s means --%s on %s and --%s elsewhere; a shorthand must name one flag everywhere",
				f.short, f.name, f.where, prev)
		}
		owner[f.short] = f.name
	}
}

// shorthand is one flag that carries a single-letter form, and where it is
// declared. Read off the live flag sets rather than a list, so the reference
// cannot offer a letter pflag would not accept.
type shorthand struct {
	short, name, usage, where string
}

// treeShorthands collects every shorthand in the tree, root's persistent flags
// first. NonInheritedFlags is what keeps a global flag from being reported again
// under every command that inherits it.
func treeShorthands(root *cobra.Command) []shorthand {
	var out []shorthand
	collect := func(fs *pflag.FlagSet, where string) {
		fs.VisitAll(func(f *pflag.Flag) {
			if f.Hidden || f.Shorthand == "" {
				return
			}
			out = append(out, shorthand{short: f.Shorthand, name: f.Name, usage: f.Usage, where: where})
		})
	}
	collect(root.PersistentFlags(), "every command")
	walkCommands(root, 0, func(c *cobra.Command, depth int) {
		if depth == 0 {
			return
		}
		collect(c.NonInheritedFlags(), "`"+c.CommandPath()+"`")
	})
	sort.SliceStable(out, func(i, j int) bool { return out[i].short < out[j].short })
	return out
}

// renderAbbrevDocs renders docs/abbreviation.md. Every row is read off a set or
// off the live tree; only the prose is written here, and it says the things no
// table can -- how a short form is actually resolved, and the rules every set obeys.
func renderAbbrevDocs(root *cobra.Command) []byte {
	var b strings.Builder

	b.WriteString("# Abbreviations\n\n")
	b.WriteString("Every short form `solace-util` accepts, and the word it stands for.\n\n")
	b.WriteString("**Generated from the abbreviation sets and the command tree -- do not edit by hand.**\n")
	b.WriteString("Regenerate after adding or removing one:\n\n")
	b.WriteString("```\ngo test ./internal/cli -update\n```\n\n")
	b.WriteString("The `test` task fails while this file is stale, so it cannot drift from the code.\n\n")

	writeResolutionNotes(&b)

	writeAbbrevTable(&b, "Commands", "Command", commandAbbrev)
	writeUnabbreviated(&b, root)
	writeAbbrevTable(&b, "Node roles", "Role", config.RoleAbbrev())
	b.WriteString("The letter is not a second spelling: it is the role's internal form, the one that\n")
	b.WriteString("appears in Kubernetes pod-name suffixes. Both spellings work anywhere a `[role]`\n")
	b.WriteString("positional or `--pod` is taken.\n\n")
	writeAbbrevTable(&b, "Platforms", "`--platform`", config.PlatformAbbrev())
	b.WriteString("The canonical word is the env file's own section key, so it is the only spelling\n")
	b.WriteString("the file, the error messages and the completions use. A short form is accepted on\n")
	b.WriteString("`--platform` alone and is refused in the env file.\n\n")
	writeShorthandTable(&b, root)

	return []byte(b.String())
}

// writeResolutionNotes states what the tables cannot: there is no expansion pass
// at all. Each kind of short form is expanded by whatever already owns that input,
// at the moment it reads it, which is why one of them can be added without the
// others knowing -- and why the rules they share are declared in one package.
func writeResolutionNotes(b *strings.Builder) {
	b.WriteString("## How a short form is resolved\n\n")
	b.WriteString("Nothing rewrites the command line. There is no expansion pass in front of\n")
	b.WriteString("execution: each word is expanded by whatever already owns that input.\n\n")
	b.WriteString("| Kind | Expanded by | When |\n| --- | --- | --- |\n")
	b.WriteString("| Command name | cobra's own `Find`, from the aliases `applyAliases` attaches at construction | dispatch |\n")
	b.WriteString("| Node role | `config.ParseRole` | while the command runs |\n")
	b.WriteString("| `--platform` value | `config.ParsePlatform` | flag handling |\n")
	b.WriteString("| Flag shorthand | pflag | flag parsing |\n\n")
	b.WriteString("What the four share is `internal/abbrev`, which owns the rules rather than the\n")
	b.WriteString("words: a short form is lowercase letters or digits, is strictly shorter than the\n")
	b.WriteString("word it stands for, is claimed by one word only, and never shadows a canonical\n")
	b.WriteString("name. A declaration that breaks one of those fails at construction, so it cannot\n")
	b.WriteString("reach a release.\n\n")
	b.WriteString("Three consequences worth knowing:\n\n")
	b.WriteString("- **Only these exact words work.** Cobra's prefix matching is off, so `dep` is not\n")
	b.WriteString("  `deploy`.\n")
	b.WriteString("- **A word means the same thing at every depth.** `br` is `broker` under whichever\n")
	b.WriteString("  verb it sits below, because the table is keyed by the command's name rather than\n")
	b.WriteString("  by its path.\n")
	b.WriteString("- **Completion teaches the canonical spelling.** The short forms work everywhere but\n")
	b.WriteString("  are deliberately not offered by TAB completion -- a completion already saves the\n")
	b.WriteString("  typing an abbreviation is for.\n\n")
}

// writeAbbrevTable renders one set. The Notes column appears only when the set
// has something to say, so a set of plain abbreviations is not padded with an
// empty column.
func writeAbbrevTable(b *strings.Builder, heading, label string, s *abbrev.Set) {
	entries := s.Entries()
	notes := false
	for _, e := range entries {
		if e.Note != "" {
			notes = true
		}
	}
	header := []string{label, "Short form"}
	rule := []string{"---", "---"}
	if notes {
		header = append(header, "Notes")
		rule = append(rule, "---")
	}
	fmt.Fprintf(b, "## %s\n\n", heading)
	mdRow(b, header)
	mdRow(b, rule)
	for _, e := range entries {
		short := "none"
		if len(e.Short) > 0 {
			short = "`" + strings.Join(e.Short, "`, `") + "`"
		}
		cells := []string{"`" + e.Canonical + "`", short}
		if notes {
			cells = append(cells, e.Note)
		}
		mdRow(b, cells)
	}
	b.WriteString("\n")
}

// mdRow writes one markdown table row.
func mdRow(b *strings.Builder, cells []string) {
	fmt.Fprintf(b, "| %s |\n", strings.Join(cells, " | "))
}

// writeUnabbreviated names the verbs the command set says nothing about. An
// abbreviation that was never granted and one that was deliberately withheld read
// the same in a table of what exists, so the ones with no entry at all are listed
// here -- a new verb shows up in this line until someone decides which it is.
func writeUnabbreviated(b *strings.Builder, root *cobra.Command) {
	var missing []string
	walkCommands(root, 0, func(c *cobra.Command, depth int) {
		if depth != 1 {
			return
		}
		if _, declared := commandAbbrev.Expand(c.Name()); !declared {
			missing = append(missing, "`"+c.Name()+"`")
		}
	})
	if len(missing) == 0 {
		return
	}
	fmt.Fprintf(b, "Verbs with no short form and no recorded reason: %s.\n\n", strings.Join(missing, ", "))
}

// writeShorthandTable renders pflag's single-letter flags. `where` is read off the
// tree so a shorthand added to one command cannot be reported as global.
func writeShorthandTable(b *strings.Builder, root *cobra.Command) {
	b.WriteString("## Flag shorthands\n\n")
	b.WriteString("| Shorthand | Flag | Where | Meaning |\n| --- | --- | --- | --- |\n")
	for _, f := range treeShorthands(root) {
		fmt.Fprintf(b, "| `-%s` | `--%s` | %s | %s |\n", f.short, f.name, f.where, mdCell(f.usage))
	}
	b.WriteString("\n")
}

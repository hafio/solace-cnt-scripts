package broker

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
)

// update regenerates docs/import.md from the classification in sections.go.
// Run `go test ./internal/broker -update` only after reading the diff; the
// committed file is the reviewed expected output.
var update = flag.Bool("update", false, "regenerate docs/import.md from sections.go")

const importDocPath = "../../docs/import.md"

// TestImportDocs is both the drift gate and the generator, the shape
// internal/cli/commanddoc_test.go established.
//
// The point of generating this document rather than writing it is that an
// operator has no other way to know what `import-config` will touch: the
// classification is baked into the binary, so a hand-written page would be a
// second copy of it and the two would drift. `test` fails while the page is
// stale, so they cannot.
func TestImportDocs(t *testing.T) {
	got := renderImportDocs()
	if *update {
		if err := os.WriteFile(importDocPath, got, 0o644); err != nil {
			t.Fatalf("write %s: %v", importDocPath, err)
		}
		return
	}
	want, err := os.ReadFile(importDocPath)
	if err != nil {
		t.Fatalf("read %s (regenerate: go test ./internal/broker -update): %v", importDocPath, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s is stale -- regenerate with: go test ./internal/broker -update", importDocPath)
	}
}

// TestEverySectionRuleHasAReason guards the thing the document is for: a rule
// whose Reason is empty tells an operator nothing about why their configuration
// was skipped, which is the whole value of publishing the table.
func TestEverySectionRuleHasAReason(t *testing.T) {
	for _, r := range Rules() {
		if strings.TrimSpace(r.Reason) == "" {
			t.Errorf("section %q has no Reason; every disposition has to be explainable", r.Section)
		}
		if r.Disposition == ApplyFiltered && r.DropLines == nil {
			t.Errorf("section %q is apply-filtered but drops no lines", r.Section)
		}
		if r.Disposition != ApplyFiltered && r.DropLines != nil {
			t.Errorf("section %q has a DropLines predicate but is %s, where it would never run",
				r.Section, r.Disposition)
		}
	}
}

// TestOnlyOneSectionIsAppliedFirst pins the isolation contract: `Create logging`
// ends the CLI session, so it is applied alone. Two First sections would share an
// invocation and the second would silently not run.
func TestOnlyOneSectionIsAppliedFirst(t *testing.T) {
	var first []string
	for _, r := range Rules() {
		if r.First {
			first = append(first, r.Section)
		}
	}
	if len(first) > 1 {
		t.Errorf("First is set on %v; each must be applied alone, so more than one needs "+
			"ImportApply to give each its own invocation", first)
	}
}

func renderImportDocs() []byte {
	var b strings.Builder

	b.WriteString("# What `import-config` applies\n\n")
	b.WriteString("This file is GENERATED from `internal/broker/sections.go` by\n")
	b.WriteString("`internal/broker/importdoc_test.go`. Do not edit it by hand: run the `regen` task.\n\n")
	b.WriteString("`broker perform import-config` decides what to do with a captured configuration\n")
	b.WriteString("section by section, using the broker's own `! Configure X` / `! Create X` comments\n")
	b.WriteString("as the unit. The classification is compiled in -- there is no config key and no\n")
	b.WriteString("flag -- because an artifact that chose what to execute would be untrusted input\n")
	b.WriteString("doing what `internal/config/execguard.go` exists to prevent.\n\n")
	b.WriteString("A section marked **skip** is REMOVED by `export-config` -- the artifact never contains\n")
	b.WriteString("it, and records the omission as a `! solace-util-omitted:` line -- and cannot be opted\n")
	b.WriteString("into. `import-config` ignores such a section if an older or hand-edited artifact still\n")
	b.WriteString("carries one, and reports it only then. Recover a node by deploying it from its own\n")
	b.WriteString("env file and letting config-sync populate it.\n\n")

	writeDispositionTable(&b, "Applied", Apply)
	writeDispositionTable(&b, "Applied, minus some lines", ApplyFiltered)
	writeDispositionTable(&b, "Never applied", Skip)
	writeDispositionTable(&b, "Appliance only (skipped on a software broker)", SkipAppliance)

	b.WriteString("## Sections that interrupt a service\n\n")
	b.WriteString("These are applied, and applying them is service-affecting. `import-config` does\n")
	b.WriteString("not ask before doing so at broker scope.\n\n")
	b.WriteString("| Section | Interrupts |\n|---|---|\n")
	for _, r := range Rules() {
		if r.Interrupts != "" {
			fmt.Fprintf(&b, "| `%s` | %s |\n", r.Section, r.Interrupts)
		}
	}
	b.WriteString("\n")

	return []byte(b.String())
}

func writeDispositionTable(b *strings.Builder, title string, d Disposition) {
	var rows []SectionRule
	for _, r := range Rules() {
		if r.Disposition == d {
			rows = append(rows, r)
		}
	}
	if len(rows) == 0 {
		return
	}
	fmt.Fprintf(b, "## %s\n\n", title)
	b.WriteString("| Section | Why |\n|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(b, "| `%s` | %s |\n", r.Section, r.Reason)
	}
	b.WriteString("\n")
}

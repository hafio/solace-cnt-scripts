package broker

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"solace/internal/output"
)

// diff.go is the ONLY error detection the config-import feature has: `cli
// -Apes <script>` exits 0 even when the broker rejects an individual line, so
// the sole way an import learns a line was silently dropped is to read the
// target back and DiffBlocks it against what the artifact asked for. A false
// pass here -- want reported as satisfied when the target does not actually
// carry it, or a block that should have been skipped reported as a failure --
// means an operator is told an import succeeded when the broker rejected part
// of it, with nothing else in the tool positioned to catch that. Every test
// below is chosen to be a case a naive "compare the line sets" implementation
// gets wrong, per DiffBlocks' own doc comment.

// newDiffBlock builds a Block for these tests without every case having to
// name Index (DiffBlocks never keys on it -- TestDiffBlocksIdentityIgnoresIndex
// is what proves that) or spell out a Lines literal separately.
func newDiffBlock(section, kind, name, vpn string, lines ...string) Block {
	return Block{Section: section, Kind: kind, Name: name, VPN: vpn, Lines: lines}
}

// TestDiffBlocksDirectionalGotMaySupersetWant is the property a naive set-
// equality implementation gets wrong first: a target's own `show
// current-config` fills in defaults the import artifact never mentioned, and
// those extra lines must not read as a difference. Swapping the two arguments
// proves the check is actually directional and not accidentally symmetric --
// a bug that made it symmetric would pass the first half of this test too.
func TestDiffBlocksDirectionalGotMaySupersetWant(t *testing.T) {
	want := newDiffBlock("Configure Message VPN: vpn1", "message-vpn", "vpn1", "vpn1",
		`message-vpn "vpn1"`,
		"authentication basic",
	)
	got := newDiffBlock("Configure Message VPN: vpn1", "message-vpn", "vpn1", "vpn1",
		`message-vpn "vpn1"`,
		"authentication basic",
		"max-connections 100", // a default the target filled in that want never carried
	)

	result := DiffBlocks([]Block{want}, []Block{got}, nil)
	if !result.OK() || result.Matched != 1 {
		t.Fatalf("got superset of want: OK=%v Matched=%d Differing=%v Missing=%v, want OK with 1 match",
			result.OK(), result.Matched, result.Differing, result.Missing)
	}

	// Swap the roles: now the 3-line block is what's wanted, and the 2-line block
	// is what the target has. The extra "max-connections 100" line is a genuine
	// requirement this time, so it must be reported missing.
	reversed := DiffBlocks([]Block{got}, []Block{want}, nil)
	if reversed.OK() {
		t.Fatal("swapping want/got must break the match -- the comparison is directional, not a set-equality check")
	}
	if len(reversed.Differing) != 1 || !reflect.DeepEqual(reversed.Differing[0].Missing, []string{"max-connections 100"}) {
		t.Errorf("reversed.Differing = %v, want exactly the one line the swapped want side needs", reversed.Differing)
	}
}

// TestDiffBlocksReportsExactMissingLine protects the actionability of the
// report: a table row can say a block differs, but only the exact missing
// line tells an operator what to go fix. Reporting the whole block instead of
// the one absent line would send them re-reading a block that mostly landed.
func TestDiffBlocksReportsExactMissingLine(t *testing.T) {
	want := newDiffBlock("Configure Authentication", "authentication", "", "",
		"authentication",
		`ldap-profile "corp-ldap" enable`,
		"max-connections 100",
	)
	got := newDiffBlock("Configure Authentication", "authentication", "", "",
		"authentication",
		`ldap-profile "corp-ldap" enable`,
		// "max-connections 100" never landed -- the one line to report.
	)

	result := DiffBlocks([]Block{want}, []Block{got}, nil)
	if len(result.Missing) != 0 {
		t.Fatalf("a block the target has (partially) is a partial mismatch, not an absent block: Missing=%v", result.Missing)
	}
	if len(result.Differing) != 1 {
		t.Fatalf("Differing = %d block(s), want 1", len(result.Differing))
	}
	bd := result.Differing[0]
	if bd.Absent {
		t.Error("Absent must be false: the target DOES have a block under this identity")
	}
	if !reflect.DeepEqual(bd.Missing, []string{"max-connections 100"}) {
		t.Errorf("Missing = %v, want exactly the one line that did not land", bd.Missing)
	}
}

// TestDiffBlocksAbsentBlockLandsInMissingNotDiffering protects the distinction
// a report reader relies on: a block the target never created at all
// (Absent) is a different failure than one that landed short (Differing), and
// they sort into different result buckets so a caller can tell them apart
// without inspecting every BlockDiff.
func TestDiffBlocksAbsentBlockLandsInMissingNotDiffering(t *testing.T) {
	want := newDiffBlock("Create Queue: orders-q", "queue", "orders-q", "vpn1",
		`queue "orders-q" message-vpn "vpn1"`,
		"max-spool-usage 1000",
	)
	// got has nothing under this identity at all.
	got := newDiffBlock("Create Queue: other-q", "queue", "other-q", "vpn1",
		`queue "other-q" message-vpn "vpn1"`,
	)

	result := DiffBlocks([]Block{want}, []Block{got}, nil)
	if len(result.Differing) != 0 {
		t.Fatalf("a block absent entirely is not a partial mismatch: Differing=%v", result.Differing)
	}
	if len(result.Missing) != 1 {
		t.Fatalf("Missing = %d block(s), want 1", len(result.Missing))
	}
	bd := result.Missing[0]
	if !bd.Absent {
		t.Error("Absent must be true when the target carries no block under this identity")
	}
	wantLines := []string{`queue "orders-q" message-vpn "vpn1"`, "max-spool-usage 1000"}
	if !reflect.DeepEqual(bd.Missing, wantLines) {
		t.Errorf("Missing = %v, want every effective want line (%v) -- there is nothing on the target to compare against", bd.Missing, wantLines)
	}
}

// TestDiffBlocksIdentityIgnoresIndex protects the reason Index is excluded
// from blockKey: the target's own export numbers and orders its blocks
// independently of the artifact, so matching on Index would fail almost every
// real comparison. Conversely, two blocks that DO share an Index must still
// be treated as different objects when their Kind differs -- Index is not a
// usable substitute for identity in either direction.
func TestDiffBlocksIdentityIgnoresIndex(t *testing.T) {
	want := Block{Index: 5, Kind: "service", Lines: []string{"service", "smf listen-port 55555"}}

	sameIdentityDifferentIndex := Block{Index: 200, Kind: "service", Lines: []string{"service", "smf listen-port 55555"}}
	result := DiffBlocks([]Block{want}, []Block{sameIdentityDifferentIndex}, nil)
	if !result.OK() || result.Matched != 1 {
		t.Errorf("blocks with the same (Kind,Name,VPN) but different Index must match: OK=%v Matched=%d", result.OK(), result.Matched)
	}

	sameIndexDifferentIdentity := Block{Index: 5, Kind: "tls", Lines: []string{"tls", "smf listen-port 55555"}}
	result = DiffBlocks([]Block{want}, []Block{sameIndexDifferentIdentity}, nil)
	if result.OK() {
		t.Error("a shared Index must not stand in for a shared (Kind,Name,VPN) identity")
	}
	if len(result.Missing) != 1 || !result.Missing[0].Absent {
		t.Errorf("Missing = %v, want one Absent block -- the \"service\" identity has nothing on the target", result.Missing)
	}
}

// TestDiffBlocksPoolsRepeatedIdentityBeforeComparing is the scenario
// diff.go's own doc comment names as the failure mode of a naive
// implementation: semp/all.cli repeats "authentication" three times (lines
// 291, 320, 352) as an LDAP-profile pass, a Radius-profile pass, and the real
// attribute-carrying pass, all sharing one (Kind,Name,VPN) identity. A lookup
// that used "whichever block is found first under a key" would land on one of
// the two near-empty early passes and read the real attribute as missing.
// Merging all three before comparing is what avoids that false failure.
func TestDiffBlocksPoolsRepeatedIdentityBeforeComparing(t *testing.T) {
	ldapPass := newDiffBlock("Create LDAP Profile", "authentication", "", "", "authentication")
	radiusPass := newDiffBlock("Create Radius Profile", "authentication", "", "", "authentication")
	realPass := newDiffBlock("Create Authentication", "authentication", "", "",
		"authentication",
		`ldap-profile "corp-ldap" enable`,
	)
	want := newDiffBlock("Configure Authentication", "authentication", "", "",
		"authentication",
		`ldap-profile "corp-ldap" enable`,
	)

	result := DiffBlocks([]Block{want}, []Block{ldapPass, radiusPass, realPass}, nil)
	if !result.OK() || result.Matched != 1 {
		t.Fatalf("expected the three got occurrences to pool into one match: OK=%v Matched=%d Differing=%v Missing=%v",
			result.OK(), result.Matched, result.Differing, result.Missing)
	}
}

// TestDiffBlocksMergeKeepsEachLinesOwnSection pins the attribution that has to
// survive the merge, and the hole that appeared when it did not.
//
// Identity is (Kind, Name, VPN) with no Section (blockKey), so two column-0 blocks
// from genuinely different `! Create ...` sections merge into one -- bare,
// Name-less scope openers like `authentication` and `routing` are built up by
// several real sections over one capture, and semp/all.cli:134 and :143 are
// exactly that for `routing`. The merged block used to keep only the LAST Section
// seen, and importIgnore then applied that one section's disposition to EVERY line
// in it. So a skipped or unclassified sibling silenced the other sections' lines
// in the diff: they were excluded from the only error detection this feature has,
// and a rejected line among them read as a clean import.
//
// Now each line carries its own section, which this asserts in the way that
// matters -- not by inspecting the merged struct, but by showing that skipping one
// section still leaves the other's missing line visible.
func TestDiffBlocksMergeKeepsEachLinesOwnSection(t *testing.T) {
	ldap := newDiffBlock("Create LDAP Profile", "authentication", "", "",
		"authentication", `  ldap-profile "corp-ldap" enable`)
	radius := newDiffBlock("Create Radius Profile", "authentication", "", "",
		"authentication", `  radius-profile "corp-radius" enable`)

	merged := mergeByIdentity([]Block{ldap, radius})
	if len(merged) != 1 {
		t.Fatalf("two blocks sharing (Kind,Name,VPN) must merge into one, got %d", len(merged))
	}
	sections := map[string]bool{}
	for _, ml := range merged[0].lines {
		sections[ml.section] = true
	}
	if !sections["Create LDAP Profile"] || !sections["Create Radius Profile"] {
		t.Errorf("merged lines lost a section: got %v, want both contributing sections", sections)
	}

	// The target has neither profile line. Skipping the LDAP section must not
	// excuse the Radius line: that is the silencing this test exists to catch.
	target := newDiffBlock("Create Radius Profile", "authentication", "", "", "authentication")
	skipLDAP := func(b Block, _ string) bool { return b.Section == "Create LDAP Profile" }

	got := DiffBlocks([]Block{ldap, radius}, []Block{target}, skipLDAP)
	if got.OK() {
		t.Fatal("a missing Radius line was excused by the LDAP section's skip -- one section's disposition must not silence another's lines")
	}
	if len(got.Differing) != 1 {
		t.Fatalf("Differing = %d block(s), want 1", len(got.Differing))
	}
	bd := got.Differing[0]
	if bd.Section != "Create Radius Profile" {
		t.Errorf("Section = %q, want the section that actually failed", bd.Section)
	}
	for _, line := range bd.Missing {
		if strings.Contains(line, "corp-ldap") {
			t.Errorf("a skipped section's line was reported missing: %q", line)
		}
	}
}

// TestDiffBlocksSeesAMismatchInsideOneNestedSubObject is the regression for the
// second of two criticals found reviewing this feature, and it is the one that
// mattered most: the diff could not see a per-object failure at all.
//
// A column-0 block can hold hundreds of independent nested objects -- in the real
// capture all 901 queues of `VPN WITH MANY QUEUES` are indented children of ONE
// `message-spool message-vpn "..."` opener. The comparison was flat set membership
// over the whole block's lines, which asks only "does SOME line here have this
// text". So when the broker rejected one line of one queue while its siblings
// carried the identical setting, every wanted line was still found somewhere and
// the block counted as MATCHED. Since `cli -Apes` exits 0 on a rejected line and
// this diff is the only error detection there is, that reported a broken import as
// a clean one.
//
// Two queues are enough to prove it: sibling Q2 carries the exact line Q1 is
// missing, which is precisely what used to mask it.
func TestDiffBlocksSeesAMismatchInsideOneNestedSubObject(t *testing.T) {
	// The blank line between the two queues is deliberate: the parser keeps blank
	// lines inside a block (blocks.go) and they carry nothing to verify, so they
	// must drop out of the comparison rather than become a key of their own.
	spool := func(q1Egress string) Block {
		return newDiffBlock("Create Message Spool", "message-spool", "", "acme",
			`message-spool message-vpn "acme"`,
			`  create queue "Q1"`,
			`    permission all consume`,
			"    "+q1Egress,
			`    exit`,
			``,
			`  create queue "Q2"`,
			`    permission all consume`,
			`    no shutdown egress`,
			`    exit`,
		)
	}
	want := spool("no shutdown egress")
	// The broker rejected Q1's egress line, so the target re-exports Q1 shut down
	// while Q2 -- untouched -- still carries the wanted text.
	got := spool("shutdown egress")

	d := DiffBlocks([]Block{want}, []Block{got}, nil)
	if d.OK() {
		t.Fatal("a rejected line inside one nested sub-object was reported as a clean import; the sibling queue's identical line masked it")
	}
	if len(d.Differing) != 1 {
		t.Fatalf("Differing = %d, want 1", len(d.Differing))
	}
	missing := strings.Join(d.Differing[0].Missing, "\n")
	if !strings.Contains(missing, "no shutdown egress") {
		t.Errorf("missing lines = %q, want Q1's rejected egress line", missing)
	}
	// Q2 landed exactly as asked, so nothing of Q2's may be reported.
	if strings.Contains(missing, `"Q2"`) {
		t.Errorf("the intact sibling queue was reported as a problem: %q", missing)
	}
}

// TestDiffBlocksReportsTheHighestMissingAncestorOnly keeps the report readable
// once the diff is structural. When a whole sub-object never landed, naming its
// opener is the actionable fact; repeating each of its attribute lines underneath
// would turn one failure into a wall of rows that all describe it.
func TestDiffBlocksReportsTheHighestMissingAncestorOnly(t *testing.T) {
	want := newDiffBlock("Create Message Spool", "message-spool", "", "acme",
		`message-spool message-vpn "acme"`,
		`  create queue "Q1"`,
		`    permission all consume`,
		`    max-spool-usage 100`,
		`    exit`,
	)
	got := newDiffBlock("Create Message Spool", "message-spool", "", "acme",
		`message-spool message-vpn "acme"`,
	)

	d := DiffBlocks([]Block{want}, []Block{got}, nil)
	if len(d.Differing) != 1 {
		t.Fatalf("Differing = %d, want 1", len(d.Differing))
	}
	missing := d.Differing[0].Missing
	if len(missing) != 1 || !strings.Contains(missing[0], `create queue "Q1"`) {
		t.Errorf("missing = %v, want only the queue opener -- its children are suppressed under it", missing)
	}
}

// TestDiffBlocksIgnorePredicateDropsLinesAndWholeIgnoredBlocks protects two
// halves of the ignore contract. Without the first half, a line an
// apply-filtered rule deliberately dropped would read as a broken import.
// Without the second, a whole section the import never intended to apply
// (nothing on the target to find) would read as an ABSENT block -- turning
// every deliberately-skipped section in the report into noisy false failures.
func TestDiffBlocksIgnorePredicateDropsLinesAndWholeIgnoredBlocks(t *testing.T) {
	kept := newDiffBlock("Configure Service", "service", "", "", "service", "smf listen-port 55555")
	wholeSectionSkipped := newDiffBlock("Configure TLS", "tls", "", "", "tls", "some-setting X")
	oneLineFiltered := newDiffBlock("Configure LDAP", "ldap", "", "",
		"ldap", "secret-line password-goes-here", "public-line enabled",
	)

	ignore := func(b Block, line string) bool {
		if b.Kind == "tls" {
			return true // a whole section the import deliberately did not apply
		}
		return b.Kind == "ldap" && strings.Contains(line, "secret-line")
	}

	got := []Block{
		newDiffBlock("Configure Service", "service", "", "", "service", "smf listen-port 55555"),
		// no "tls" block at all -- it was never supposed to land.
		newDiffBlock("Configure LDAP", "ldap", "", "", "ldap", "public-line enabled"),
	}

	result := DiffBlocks([]Block{kept, wholeSectionSkipped, oneLineFiltered}, got, ignore)
	if result.Matched != 2 {
		t.Errorf("Matched = %d, want 2 (service and ldap; tls is ignored entirely, not one of the checked blocks)", result.Matched)
	}
	if len(result.Missing) != 0 {
		t.Errorf("Missing = %v, want none -- the ignored tls block must not read as absent even though the target has nothing under that identity", result.Missing)
	}
	if len(result.Differing) != 0 {
		t.Errorf("Differing = %v, want none -- the ignored secret-line must not read as a missing line", result.Differing)
	}
}

// TestDiffBlocksNormalizesWhitespaceButNotCaseOrQuotes protects the exact
// boundary normalizeLine draws. semp/all.cli genuinely double-spaces some
// lines ("system scaling  scale  max-connections 1000"), and nothing
// guarantees the target's own re-export reproduces the same run length, so
// that must not read as a difference. But `"55003"` and `55003` are different
// tokens to the broker -- collapsing that too would hide a real mismatch.
func TestDiffBlocksNormalizesWhitespaceButNotCaseOrQuotes(t *testing.T) {
	want := []Block{newDiffBlock("Configure System Scaling", "system", "", "",
		"system scaling",
		"scale max-connections 1000",
		`service smf listen-port "55003"`,
	)}
	got := []Block{newDiffBlock("Configure System Scaling", "system", "", "",
		"system  scaling",              // extra internal space: must still match
		"scale  max-connections 1000  ", // extra internal space + trailing: must still match
		"service smf listen-port 55003", // unquoted: a genuinely different token
	)}

	result := DiffBlocks(want, got, nil)
	if len(result.Differing) != 1 {
		t.Fatalf("Differing = %d block(s), want exactly 1 (the quoting difference); whitespace-only differences must not count", len(result.Differing))
	}
	bd := result.Differing[0]
	if !reflect.DeepEqual(bd.Missing, []string{`service smf listen-port "55003"`}) {
		t.Errorf("Missing = %v, want exactly the quoted line", bd.Missing)
	}
}

// TestDiffNormalizeLinePreservesIndentationCaseAndQuoting is the direct unit
// test of the normalisation rule the block-level test above exercises
// end-to-end: it pins each half (what collapses, what must not) against the
// function itself so a future change to one rule cannot hide behind the
// other still passing.
func TestDiffNormalizeLinePreservesIndentationCaseAndQuoting(t *testing.T) {
	if a, b := normalizeLine("scale  max-connections 1000"), normalizeLine("scale max-connections 1000"); a != b {
		t.Errorf("internal whitespace run did not collapse: %q != %q", a, b)
	}
	if a, b := normalizeLine("max-connections 1000   "), normalizeLine("max-connections 1000"); a != b {
		t.Errorf("trailing whitespace was not trimmed: %q != %q", a, b)
	}
	// Leading indentation is structural (blocks.go derives block structure from
	// it) and must survive exactly, not collapse to a single space like an
	// internal run would.
	if a, b := normalizeLine("  x"), normalizeLine("    x"); a == b {
		t.Errorf("differing indentation must not normalise to the same string, both gave %q", a)
	}
	if a, b := normalizeLine("Enable"), normalizeLine("enable"); a == b {
		t.Errorf("case must be preserved, both normalised to %q", a)
	}
	if a, b := normalizeLine(`"55003"`), normalizeLine("55003"); a == b {
		t.Errorf("a quoted and an unquoted operand are different tokens to the broker, both normalised to %q", a)
	}
}

// TestDiffResultOKOnlyWhenDifferingAndMissingEmpty pins the exact predicate a
// caller (RenderDiffResult, and the import command's own exit-status logic)
// relies on to decide pass/fail. Matched is deliberately excluded from the
// case table's expectations: a high Matched count must never paper over a
// non-empty Differing or Missing.
func TestDiffResultOKOnlyWhenDifferingAndMissingEmpty(t *testing.T) {
	cases := []struct {
		name string
		d    DiffResult
		want bool
	}{
		{"both empty", DiffResult{Matched: 5}, true},
		{"differing only", DiffResult{Differing: []BlockDiff{{}}}, false},
		{"missing only", DiffResult{Missing: []BlockDiff{{}}}, false},
		{"both populated", DiffResult{Differing: []BlockDiff{{}}, Missing: []BlockDiff{{}}}, false},
	}
	for _, c := range cases {
		if got := c.d.OK(); got != c.want {
			t.Errorf("%s: OK() = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestDiffBlockLabelFormatsIdentityForTheReport pins blockLabel's branches
// directly: it is the only place a report row's identity turns into the
// human-readable name RenderDiffResult prints beside a block's missing lines,
// so a wrong branch here silently mislabels every dirty report.
func TestDiffBlockLabelFormatsIdentityForTheReport(t *testing.T) {
	cases := []struct {
		name string
		bd   BlockDiff
		want string
	}{
		{
			"VPN-scoped object gets a qualifier",
			BlockDiff{Kind: "client-username", Name: "myuser", VPN: "vpn1"},
			`client-username "myuser" (message-vpn "vpn1")`,
		},
		{
			"the VPN block itself is not qualified against its own name",
			BlockDiff{Kind: "message-vpn", Name: "vpn1", VPN: "vpn1"},
			`message-vpn "vpn1"`,
		},
		{
			"a bare scope opener has no name and no qualifier",
			BlockDiff{Kind: "authentication"},
			"authentication",
		},
		{
			"no kind or name falls back to the section comment",
			BlockDiff{Section: "Configure Something Bare"},
			"Configure Something Bare",
		},
		{
			"the section fallback can still take a VPN qualifier",
			BlockDiff{VPN: "vpn1", Section: "whatever"},
			`whatever (message-vpn "vpn1")`,
		},
	}
	for _, c := range cases {
		if got := blockLabel(c.bd); got != c.want {
			t.Errorf("%s: blockLabel = %q, want %q", c.name, got, c.want)
		}
	}
}

// TestRenderDiffResultCleanReportHasNoFailureRows protects the quiet path: a
// clean verification must not print a table, a [FAIL] tag, or any per-block
// detail -- a report that always prints the same scaffolding regardless of
// outcome would bury the one signal (does anything need attention) an
// operator is reading the report for.
func TestRenderDiffResultCleanReportHasNoFailureRows(t *testing.T) {
	buf := &bytes.Buffer{}
	RenderDiffResult(output.New(buf), DiffResult{Matched: 3})
	out := buf.String()

	if !strings.Contains(out, "Configuration Verification") {
		t.Errorf("missing the report section header: %q", out)
	}
	if !strings.Contains(out, "3 block(s)") {
		t.Errorf("missing the matched count: %q", out)
	}
	for _, absent := range []string{"[FAIL]", "absent from the target", "missing these line(s)"} {
		if strings.Contains(out, absent) {
			t.Errorf("clean report must not contain %q: %q", absent, out)
		}
	}
}

// TestRenderDiffResultDirtyReportNamesOffendingSectionAndLines protects the
// only actionable part of a dirty report: the table says something failed,
// but only the per-block section label and the exact missing lines tell an
// operator what to go fix on the broker.
func TestRenderDiffResultDirtyReportNamesOffendingSectionAndLines(t *testing.T) {
	d := DiffResult{
		Matched: 1,
		Missing: []BlockDiff{{
			Section: "Create Message Vpn: xps-ps-01",
			Kind:    "message-vpn", Name: "xps-ps-01", VPN: "xps-ps-01",
			Absent:  true,
			Missing: []string{`message-vpn "xps-ps-01"`, "authentication basic"},
		}},
		Differing: []BlockDiff{{
			Section: "Configure Authentication",
			Kind:    "authentication",
			Missing: []string{`ldap-profile "corp-ldap" enable`},
		}},
	}

	buf := &bytes.Buffer{}
	RenderDiffResult(output.New(buf), d)
	out := buf.String()

	if n := strings.Count(out, "[FAIL]"); n != 2 {
		t.Errorf("expected one [FAIL] table row per offending block, got %d in: %q", n, out)
	}
	for _, want := range []string{
		"Create Message Vpn: xps-ps-01", // the absent block's table row names its section
		`message-vpn "xps-ps-01": absent from the target`,
		`    message-vpn "xps-ps-01"`,
		"    authentication basic",
		"authentication: missing these line(s) on the target",
		`    ldap-profile "corp-ldap" enable`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("dirty report missing %q, got: %q", want, out)
		}
	}
}

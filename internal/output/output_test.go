package output

import (
	"bytes"
	"strings"
	"testing"
)

// sink returns a Sink over a buffer plus the buffer, the shape every test here
// uses: assert on exactly what a terminal would have received.
func sink() (*Sink, *bytes.Buffer) {
	var buf bytes.Buffer
	return New(&buf), &buf
}

// TestTagsShareOneWidth is the invariant the whole tag set exists for: a column
// of mixed outcomes must not jog sideways. It is what stops a 7-character
// [ERROR] (the tag [FAIL] replaced) coming back.
func TestTagsShareOneWidth(t *testing.T) {
	tags := map[string]string{"ok": tagOK, "warn": tagWarn, "fail": tagFail, "info": tagInfo, "skip": tagSkip}
	want := len(tagOK)
	for name, tag := range tags {
		if len(tag) != want {
			t.Errorf("tag %s = %q, width %d, want width %d -- every tag must be one width",
				name, tag, len(tag), want)
		}
		if !strings.HasPrefix(tag, "[") || !strings.HasSuffix(tag, "]") {
			t.Errorf("tag %s = %q, want it bracketed", name, tag)
		}
	}
}

// TestStepKeepsTheArrowPrefix pins the one string README.md and the -v/--verbose
// help text both promise verbatim.
func TestStepKeepsTheArrowPrefix(t *testing.T) {
	s, buf := sink()
	s.Step("using %s: %s", "kubectl", "/usr/bin/kubectl")
	if got, want := buf.String(), "==> using kubectl: /usr/bin/kubectl\n"; got != want {
		t.Errorf("Step = %q, want %q", got, want)
	}
}

// TestLeveledLinesCarryTheTagAndNoArrow is the house rule made testable: a
// stderr line is `==> <phase>` OR `[TAG ] <msg>`, never both.
func TestLeveledLinesCarryTheTagAndNoArrow(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func(s *Sink)
		want string
	}{
		{"ok", func(s *Sink) { s.OK("dns: %s", "primary01") }, "[ OK ] dns: primary01\n"},
		{"warn", func(s *Sink) { s.Warn("no labels for %s", "backup") }, "[WARN] no labels for backup\n"},
		{"fail", func(s *Sink) { s.Fail("copy %s: %v", "a.txt", "denied") }, "[FAIL] copy a.txt: denied\n"},
		{"info", func(s *Sink) { s.Info("reverting %s", "mate") }, "[INFO] reverting mate\n"},
		{"skip", func(s *Sink) { s.Skip("p4: %s", "podman only") }, "[SKIP] p4: podman only\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, buf := sink()
			tc.call(s)
			got := buf.String()
			if got != tc.want {
				t.Errorf("%s = %q, want %q", tc.name, got, tc.want)
			}
			if strings.Contains(got, "==>") {
				t.Errorf("%s = %q, must not carry the ==> prefix as well as a tag", tc.name, got)
			}
		})
	}
}

// TestLineAddsNoPrefix proves the raw end of the chain stays raw -- it is what
// the CLI hands to Cluster/Manager/Ops as their line sink.
func TestLineAddsNoPrefix(t *testing.T) {
	s, buf := sink()
	s.Line("plain %d", 7)
	if got, want := buf.String(), "plain 7\n"; got != want {
		t.Errorf("Line = %q, want %q", got, want)
	}
}

// TestNewFuncRoutesThroughTheGivenLineSink is the seam Cluster/Manager/Ops use:
// a Sink built over another Sink's Line must still add its own prefixes exactly
// once.
func TestNewFuncRoutesThroughTheGivenLineSink(t *testing.T) {
	var buf bytes.Buffer
	outer := New(&buf)
	inner := NewFunc(outer.Line)
	inner.Step("phase")
	inner.Warn("careful")
	if got, want := buf.String(), "==> phase\n[WARN] careful\n"; got != want {
		t.Errorf("through line sink = %q, want %q", got, want)
	}
}

// TestNilSinksDiscard covers every "nobody wired a sink" path: a nil *Sink, a
// Sink over a nil writer, and a Sink over a nil func all stay silent instead of
// panicking, the same tolerance the logf helpers they replace already had.
func TestNilSinksDiscard(t *testing.T) {
	var nilSink *Sink
	for name, s := range map[string]*Sink{
		"nil sink":   nilSink,
		"nil writer": New(nil),
		"nil func":   NewFunc(nil),
	} {
		t.Run(name, func(t *testing.T) {
			// Every method, since each one is a separate chance to dereference.
			s.Line("x")
			s.Step("x")
			s.OK("x")
			s.Warn("x")
			s.Fail("x")
			s.Info("x")
			s.Skip("x")
			s.Section("x")
			s.KVBlock([]KV{{Key: "k", Value: "v"}})
			s.Table([]string{"A"}, [][]string{{"1"}})
		})
	}
}

// TestSectionPadsToWidth pins the report header shape, including the deliberate
// refusal to truncate: a clipped title names nothing.
func TestSectionPadsToWidth(t *testing.T) {
	s, buf := sink()
	s.Section("Broker deployment (kubernetes)")
	line := strings.TrimSuffix(buf.String(), "\n")
	if len(line) != SectionWidth {
		t.Errorf("Section line = %q, len %d, want %d", line, len(line), SectionWidth)
	}
	if !strings.HasPrefix(line, "=== Broker deployment (kubernetes) =") {
		t.Errorf("Section = %q, want the title between the borders", line)
	}

	buf.Reset()
	long := strings.Repeat("x", SectionWidth+10)
	s.Section(long)
	if got := strings.TrimSuffix(buf.String(), "\n"); got != "=== "+long+" " {
		t.Errorf("long Section = %q, want the title whole and untruncated", got)
	}
}

// TestKVBlockAlignsOnTheLongestKey is the point of the helper: the column is
// computed, so no caller hand-counts spaces again.
func TestKVBlockAlignsOnTheLongestKey(t *testing.T) {
	s, buf := sink()
	s.KVBlock([]KV{
		{Key: "name", Value: "solace-dev"},
		{Key: "storage class", Value: "gp3"},
		{Key: "ns", Value: "solace"},
	})
	want := "  name          : solace-dev\n" +
		"  storage class : gp3\n" +
		"  ns            : solace\n"
	if got := buf.String(); got != want {
		t.Errorf("KVBlock =\n%q\nwant\n%q", got, want)
	}

	// No rows must not emit a stray blank line into a report.
	buf.Reset()
	s.KVBlock(nil)
	if got := buf.String(); got != "" {
		t.Errorf("empty KVBlock = %q, want nothing", got)
	}
}

// TestKVRowHonorsAnExplicitWidth covers the multi-method report case: rows
// emitted from different call sites line up because they name one width, and a
// key wider than the column still gets its single separating space.
func TestKVRowHonorsAnExplicitWidth(t *testing.T) {
	s, buf := sink()
	s.KVRow(15, "operator ns", "pubsubplus-operator-system")
	s.KVRow(15, "storage class", "skipped (preview)")
	want := "  operator ns     : pubsubplus-operator-system\n" +
		"  storage class   : skipped (preview)\n"
	if got := buf.String(); got != want {
		t.Errorf("KVRow =\n%q\nwant\n%q", got, want)
	}

	buf.Reset()
	s.KVRow(3, "muchlongerkey", "v")
	if got, want := buf.String(), "  muchlongerkey : v\n"; got != want {
		t.Errorf("over-wide key = %q, want %q", got, want)
	}
}

// TestTableComputesColumnWidths covers the header/row width interaction and the
// deliberate lack of trailing whitespace on the last column.
func TestTableComputesColumnWidths(t *testing.T) {
	s, buf := sink()
	s.Table([]string{"NAME", "IMAGE", "STATUS"}, [][]string{
		{"solace-dev", "solace/pubsubplus-standard:10.10", "Up 3 minutes"},
		{"x", "y", "Exited"},
	})
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("Table wrote %d lines, want 3: %q", len(lines), buf.String())
	}
	for i, l := range lines {
		if strings.HasSuffix(l, " ") {
			t.Errorf("line %d = %q, want no trailing whitespace on the last column", i, l)
		}
	}
	// The widest cell in column 1 sets the column, so column 2 starts at the
	// same offset on every line.
	off := strings.Index(lines[0], "STATUS")
	if off <= 0 || strings.Index(lines[1], "Up 3 minutes") != off || strings.Index(lines[2], "Exited") != off {
		t.Errorf("column 3 offsets differ across\n%q", buf.String())
	}

	// A header with no rows is still a header.
	buf.Reset()
	s.Table([]string{"A", "B"}, nil)
	if got, want := buf.String(), "A B\n"; got != want {
		t.Errorf("header-only Table = %q, want %q", got, want)
	}

	// A row shorter than the header must not panic or pad past its own cells.
	buf.Reset()
	s.Table([]string{"A", "BBBB"}, [][]string{{"1"}})
	if got, want := buf.String(), "A BBBB\n1\n"; got != want {
		t.Errorf("ragged Table = %q, want %q", got, want)
	}
}

// TestLevelTagsMapToTheHouseTags pins the Level -> tag mapping, including that an
// unknown/zero Level reads as [INFO] rather than as a passing check -- a report
// that defaulted to [ OK ] would claim verification nobody did.
func TestLevelTagsMapToTheHouseTags(t *testing.T) {
	for _, tc := range []struct {
		l    Level
		want string
	}{
		{LevelInfo, tagInfo},
		{LevelOK, tagOK},
		{LevelWarn, tagWarn},
		{LevelFail, tagFail},
		{LevelSkip, tagSkip},
		{Level(99), tagInfo},
	} {
		if got := tc.l.tag(); got != tc.want {
			t.Errorf("Level(%d).tag() = %q, want %q", tc.l, got, tc.want)
		}
	}
	if TagWidth != len(tagOK)+1 {
		t.Errorf("TagWidth = %d, want the tag plus its trailing space (%d)", TagWidth, len(tagOK)+1)
	}
}

// TestLevelOrderKeepsSkipBelowFail pins the ordering the worst-of idiom relies
// on ("if row.Level > lvl", used across the k8s reports to fold a whole section
// down to its single worst row): LevelSkip must NOT outrank LevelFail, or one
// row nobody checked would mask a real failure in any summary built that way.
// A future reorder that reintroduces Skip above Fail must fail here, not
// silently in a report.
func TestLevelOrderKeepsSkipBelowFail(t *testing.T) {
	if !(LevelInfo < LevelOK && LevelOK < LevelWarn && LevelWarn < LevelSkip && LevelSkip < LevelFail) {
		t.Errorf("Level order = Info=%d OK=%d Warn=%d Skip=%d Fail=%d, want Info < OK < Warn < Skip < Fail",
			LevelInfo, LevelOK, LevelWarn, LevelSkip, LevelFail)
	}
}

// TestKVRowAtLeadsWithTheTag covers the tagged report row: the tag leads, the key
// column still aligns, and the separator stays the house " : ".
func TestKVRowAtLeadsWithTheTag(t *testing.T) {
	s, buf := sink()
	s.KVRowAt(LevelInfo, 14, "namespace", "solace")
	s.KVRowAt(LevelOK, 14, "storage class", "standard")
	want := "[INFO] namespace      : solace\n" +
		"[ OK ] storage class  : standard\n"
	if got := buf.String(); got != want {
		t.Errorf("KVRowAt =\n%q\nwant\n%q", got, want)
	}
}

// TestTaggedTableAlignsUnderTheTagColumn is the property that makes a tagged
// table readable: the header sits over the CELLS, not over the tags, and every
// column is computed from header and rows together.
func TestTaggedTableAlignsUnderTheTagColumn(t *testing.T) {
	s, buf := sink()
	s.TaggedTable([]string{"NAME", "ROLE", "READY"}, []Row{
		{LevelOK, []string{"solace-prod-pubsubplus-p-0", "primary", "3/3"}},
		{LevelWarn, []string{"x", "monitor", "0/1"}},
	})
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("wrote %d lines, want header + 2 rows:\n%s", len(lines), buf.String())
	}
	// The header is indented by exactly the tag column, so NAME starts where the
	// first cell of every row starts.
	nameCol := strings.Index(lines[0], "NAME")
	if nameCol != TagWidth {
		t.Errorf("header starts at %d, want TagWidth (%d)", nameCol, TagWidth)
	}
	for i, l := range lines[1:] {
		if strings.Index(l, "solace")+strings.Index(l, "x") < 0 {
			continue
		}
		if !strings.HasPrefix(l[TagWidth:], "solace") && !strings.HasPrefix(l[TagWidth:], "x") {
			t.Errorf("row %d does not start its first cell at TagWidth: %q", i, l)
		}
	}
	// ROLE lines up across header and both rows.
	off := strings.Index(lines[0], "ROLE")
	if strings.Index(lines[1], "primary") != off || strings.Index(lines[2], "monitor") != off {
		t.Errorf("ROLE column is ragged:\n%s", buf.String())
	}
	for i, l := range lines {
		if strings.HasSuffix(l, " ") {
			t.Errorf("line %d has trailing whitespace: %q", i, l)
		}
	}

	// No rows: a header with nothing under it is still a header (an empty section
	// must not silently vanish and read as "not checked").
	buf.Reset()
	s.TaggedTable([]string{"A"}, nil)
	if got := buf.String(); got != strings.Repeat(" ", TagWidth)+"A\n" {
		t.Errorf("header-only TaggedTable = %q", got)
	}

	// No header at all: rows only, no stray blank line.
	buf.Reset()
	s.TaggedTable(nil, []Row{{LevelOK, []string{"only"}}})
	if got, want := buf.String(), "[ OK ] only\n"; got != want {
		t.Errorf("headerless TaggedTable = %q, want %q", got, want)
	}
}

// TestWarningIsFindableInAScrollback pins the properties that make the loud block
// worth having: blank lines around it, a border that is NOT the report section's
// (an alarm must not look like a heading), the title in the header, and an
// indented body where an empty entry becomes a real blank line.
func TestWarningIsFindableInAScrollback(t *testing.T) {
	s, buf := sink()
	s.Warning("operator downgrade",
		"the operator running in this cluster is NEWER than the one",
		"this env file would install:",
		"",
		"  existing : 1.4.2",
		"  new      : 1.4.0",
	)
	got := buf.String()
	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")

	if lines[0] != "" || lines[len(lines)-1] != "" {
		t.Errorf("a loud warning must be surrounded by blank lines:\n%q", got)
	}
	head := lines[1]
	if !strings.HasPrefix(head, "!!! WARNING: operator downgrade ") {
		t.Errorf("header = %q, want the title in a !!! border", head)
	}
	if strings.Contains(head, "===") {
		t.Error("a warning must not wear the report section's border -- an alarm is not a heading")
	}
	if len(head) != SectionWidth {
		t.Errorf("header len = %d, want SectionWidth (%d)", len(head), SectionWidth)
	}
	// Prose sits at two spaces; a caller that pre-indents a key/value pair gets it
	// nested one level deeper, which is how the comparison reads as a sub-block.
	if !strings.Contains(got, "\n  the operator running") {
		t.Errorf("prose should be indented by two spaces:\n%q", got)
	}
	if !strings.Contains(got, "\n    existing : 1.4.2\n") {
		t.Errorf("a pre-indented body line should nest one level deeper:\n%q", got)
	}
	// The empty body entry became a real blank line inside the block.
	if !strings.Contains(got, "would install:\n\n    existing") {
		t.Errorf("an empty body entry should render as a blank separator:\n%q", got)
	}
}

// TestWarningTitleLongerThanTheBorder covers the same non-truncation rule Section
// has: a long title is printed whole, because a clipped alarm names nothing.
func TestWarningTitleLongerThanTheBorder(t *testing.T) {
	s, buf := sink()
	long := strings.Repeat("y", SectionWidth+5)
	s.Warning(long, "body")
	if !strings.Contains(buf.String(), long) {
		t.Errorf("long warning title was truncated:\n%q", buf.String())
	}
}

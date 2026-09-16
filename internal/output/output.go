// Package output is the one place this tool's stdout/stderr conventions live.
//
// Two voices, split by stream, because a script must be able to capture the
// report without the progress narration mixed into it:
//
//   - stderr, progress: `==> <phase>` for a phase announcement, `[TAG ] <msg>`
//     for a leveled status line. A line is one or the other, NEVER both -- an
//     `==> [WARN] ...` line says the same thing twice in two shapes.
//   - stdout, reports: `=== <title> ===...` section headers, aligned key/value
//     blocks and tables. This is the content `generate` and `status` produce,
//     so it is what a pipe is expected to carry.
//
// Every tag is the same width, so a column of mixed outcomes does not jog left
// and right by a character per line (the reason [ERROR] became [FAIL]). Output
// is plain ASCII with no ANSI styling: these lines are read in PowerShell 5.1
// consoles, captured into dev-script logs, and grepped.
package output

import (
	"fmt"
	"io"
	"strings"
)

// Tags are fixed-width on purpose -- see tagWidth and TestTagsShareOneWidth.
const (
	tagOK   = "[ OK ]"
	tagWarn = "[WARN]"
	tagFail = "[FAIL]"
	tagInfo = "[INFO]"
	tagSkip = "[SKIP]"
)

// SectionWidth is the total width of a Section header line, borders included.
const SectionWidth = 60

// Sink formats lines and hands them to one emitter. It holds nothing else, so
// constructing one is free and a caller never needs to cache it.
//
// The emitter is either a writer (New) or a line function (NewFunc). The second
// form exists because Cluster/Manager/Ops take their progress channel as a
// `func(string, ...any)` seam that ~80 call sites already pass: wrapping that
// func keeps every one of them compiling while still routing the formatting
// through this package.
type Sink struct {
	emit func(format string, args ...any)
}

// New builds a Sink that writes formatted lines to w. A nil w discards.
func New(w io.Writer) *Sink {
	if w == nil {
		return &Sink{}
	}
	return &Sink{emit: func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}}
}

// NewFunc builds a Sink over an existing line sink -- a func that takes a
// format string and its arguments and emits exactly one line. A nil f discards.
func NewFunc(f func(format string, args ...any)) *Sink { return &Sink{emit: f} }

// Line emits one line verbatim, adding no prefix. It is the raw end of the
// chain: the CLI hands this method to Cluster/Manager/Ops as their line sink so
// that the prefix decisions stay here, in the Sink the callee builds over it.
func (s *Sink) Line(format string, args ...any) {
	if s == nil || s.emit == nil {
		return
	}
	s.emit(format, args...)
}

// Step announces one phase of work: `==> <msg>`. README and the -v/--verbose
// help text both promise this prefix verbatim.
func (s *Sink) Step(format string, args ...any) { s.Line("==> "+format, args...) }

// OK, Warn, Fail, Info and Skip are the leveled status lines. They carry no
// `==> `: the tag already says what the line is. Skip reports work deliberately
// not done -- a step that does not apply here -- which is neither a success nor
// a failure and must not be read as either.
func (s *Sink) OK(format string, args ...any)   { s.Line(tagOK+" "+format, args...) }
func (s *Sink) Warn(format string, args ...any) { s.Line(tagWarn+" "+format, args...) }
func (s *Sink) Fail(format string, args ...any) { s.Line(tagFail+" "+format, args...) }
func (s *Sink) Info(format string, args ...any) { s.Line(tagInfo+" "+format, args...) }
func (s *Sink) Skip(format string, args ...any) { s.Line(tagSkip+" "+format, args...) }

// Level is the verdict a report row carries. It exists so a reader can tell what
// the tool actually CHECKED from what it merely read out of the env file -- a
// report where both look alike invites trusting a value nobody verified.
//
// The meanings are deliberately conservative, because a tag that cries wolf is
// worse than no tag:
//
//   - LevelInfo -- a neutral fact. Echoed config, or a resource whose state
//     carries no health meaning. NOT a judgement.
//   - LevelOK   -- checked, and in its expected steady state.
//   - LevelWarn -- checked, exists, but not in that state yet or no longer:
//     Pending, a LoadBalancer with no address, replicas short. Worth a look,
//     not necessarily wrong.
//   - LevelFail -- checked, and in a state that needs intervention.
//   - LevelSkip -- not checked, and the report says so rather than implying pass.
//
// LevelSkip sits between LevelWarn and LevelFail, NOT above LevelFail. The
// worst-of idiom used across the k8s reports ("if row.Level > lvl") folds a
// whole section down to its single worst row, and ranking Skip as the highest
// value would let one row nobody checked outrank a row that was checked and
// found broken -- a real failure hiding behind "not checked" is worse than the
// reverse. Skip still outranks Warn: a row skipped outright withholds more
// information than one that was checked and merely raised a caution.
type Level int

const (
	LevelInfo Level = iota
	LevelOK
	LevelWarn
	LevelSkip
	LevelFail
)

// tag renders the fixed-width tag for a level.
func (l Level) tag() string {
	switch l {
	case LevelOK:
		return tagOK
	case LevelWarn:
		return tagWarn
	case LevelFail:
		return tagFail
	case LevelSkip:
		return tagSkip
	default:
		return tagInfo
	}
}

// TagWidth is the width a tag plus its trailing space occupies, so callers that
// need to line untagged text up under tagged rows can indent by it.
var TagWidth = len(tagOK) + 1

// KVRowAt emits a key/value row led by its verdict tag:
//
//	[ OK ] storage class   : standard
//
// It is KVRow with the tag column, for the reports where the distinction between
// "checked" and "echoed" is the point.
func (s *Sink) KVRowAt(l Level, width int, key, value string) {
	s.Line("%s %-*s : %s", l.tag(), width, key, value)
}

// Row is one TaggedTable row: a verdict plus the cells.
type Row struct {
	Level Level
	Cells []string
}

// TaggedTable is Table with a leading verdict column. The header is indented by
// TagWidth so its titles line up over the cells rather than over the tags.
func (s *Sink) TaggedTable(header []string, rows []Row) {
	cells := make([][]string, len(rows))
	for i, r := range rows {
		cells[i] = r.Cells
	}
	widths := columnWidths(header, cells)
	if len(header) > 0 {
		s.Line("%s%s", strings.Repeat(" ", TagWidth), pad(taggedGap, header, widths))
	}
	for _, r := range rows {
		s.Line("%s %s", r.Level.tag(), pad(taggedGap, r.Cells, widths))
	}
}

// The column gap, which differs between the two tables and is stated here rather than
// being a literal at each call site. A tagged table's rows already start behind a tag,
// so its columns need the wider gap to stay separable; a plain table does not.
const (
	tableGap  = " "
	taggedGap = "  "
)

// columnWidths measures every column against the widest cell in it, the header
// included, so both tables size themselves by the same rule.
func columnWidths(header []string, rows [][]string) []int {
	widths := make([]int, len(header))
	for i, h := range header {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, c := range row {
			if i < len(widths) && len(c) > widths[i] {
				widths[i] = len(c)
			}
		}
	}
	return widths
}

// pad joins cells at the given widths, leaving the LAST one unpadded so no line carries
// trailing whitespace into a log or a diff.
func pad(gap string, cells []string, widths []int) string {
	parts := make([]string, len(cells))
	for i, c := range cells {
		if i < len(cells)-1 && i < len(widths) {
			parts[i] = fmt.Sprintf("%-*s", widths[i], c)
		} else {
			parts[i] = c
		}
	}
	return strings.Join(parts, gap)
}

// Warning prints a loud, bordered block for the handful of warnings that change
// what an operator should do NEXT -- a config that silently did not take effect,
// an HA group that will never form, an install that would downgrade something
// every other broker shares.
//
// It is deliberately reserved. Most warnings are one-liners and should stay that
// way: best-effort cleanup that failed, a teardown step that found nothing to
// stop, one malformed entry skipped in a loop. A block around every one of those
// would train the reader to scroll past blocks.
//
// The border is `!!!`, not the `===` a report Section uses, because the two land
// on different streams and can interleave in a terminal -- an alarm must not look
// like a heading. Blank lines above and below are part of the point: the block
// has to be findable in a scrollback after the fact.
//
// A body line that is empty renders as a blank line, so a caller can separate a
// prose paragraph from an indented key/value pair.
func (s *Sink) Warning(title string, body ...string) {
	head := "!!! WARNING: " + title + " "
	if pad := SectionWidth - len(head); pad > 0 {
		head += strings.Repeat("!", pad)
	}
	s.Line("")
	s.Line("%s", head)
	for _, l := range body {
		if l == "" {
			s.Line("")
			continue
		}
		s.Line("  %s", l)
	}
	s.Line("")
}

// Section prints a bordered report header padded to SectionWidth:
//
//	=== Broker deployment (kubernetes) =========================
//
// A title long enough to fill the width is printed whole rather than truncated:
// a header exists to name the block, and a clipped name names nothing.
func (s *Sink) Section(title string) {
	head := "=== " + title + " "
	if pad := SectionWidth - len(head); pad > 0 {
		head += strings.Repeat("=", pad)
	}
	s.Line("%s", head)
}

// KV is one row of a key/value report block.
type KV struct{ Key, Value string }

// KVBlock renders rows as `  <key> : <value>`, every key padded to the width of
// the longest key in THIS block. The alignment is computed, never hand-counted,
// so adding a longer field name cannot leave the column ragged.
func (s *Sink) KVBlock(rows []KV) {
	width := 0
	for _, r := range rows {
		if len(r.Key) > width {
			width = len(r.Key)
		}
	}
	for _, r := range rows {
		s.KVRow(width, r.Key, r.Value)
	}
}

// KVRow emits one key/value row at an explicit column width. It exists for the
// reports whose rows are produced by SEVERAL methods -- a check report that
// interleaves locally known config with lines that each cost a cluster
// round-trip -- where no single call site can see every key. Such a report names
// its width once, as a constant, so widening the column is still one edit rather
// than one per line. A report built in one place uses KVBlock and computes it.
func (s *Sink) KVRow(width int, key, value string) {
	s.Line("  %-*s : %s", width, key, value)
}

// Table renders a header row and its rows, each column padded to the widest
// cell in it, header included. The last column is never padded, so no line
// carries trailing whitespace into a log or a diff.
func (s *Sink) Table(header []string, rows [][]string) {
	widths := columnWidths(header, rows)
	s.Line("%s", pad(tableGap, header, widths))
	for _, row := range rows {
		s.Line("%s", pad(tableGap, row, widths))
	}
}

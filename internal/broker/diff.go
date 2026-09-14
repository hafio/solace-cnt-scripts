package broker

import (
	"fmt"
	"regexp"
	"strings"

	"solace/internal/output"
)

// This file is the ONLY error detection the config-import feature has.
//
// `cli -Apes <script>` runs every line of a script independently and keeps going
// past a rejected one -- ExecCLI's doc comment (config_ops.go) records that "a
// line the broker rejects does not stop the rest from running" -- and the
// process itself does not fail for that: ExecCLI has to scan the CAPTURED OUTPUT
// TEXT for "invalid" / "error" / "busy" rather than trust the exit code, and even
// that scan only says something in the transcript looked like a rejection, never
// which setting it was or whether it stuck. So the only way to know whether an
// imported block of configuration actually landed is to read the target back
// with another `show current-config` and compare the two -- which is what
// DiffBlocks does. If this comparison is wrong, the import feature has no other
// way to notice a silently-rejected line.

// BlockDiff is one block's verification result: identity, and either the lines
// the artifact wanted that the target does not have (a partial mismatch) or
// Absent, when the target has no block at all under this identity.
type BlockDiff struct {
	Section string
	Kind    string
	Name    string
	VPN     string
	Missing []string // lines in the artifact that the target does not have
	Absent  bool     // the whole block is absent from the target
}

// DiffResult is the outcome of comparing every block the artifact wants applied
// against what a fresh read of the target actually has.
type DiffResult struct {
	Matched   int
	Differing []BlockDiff
	Missing   []BlockDiff
}

// OK reports whether every block the artifact wanted applied was found on the
// target carrying every line it is supposed to.
func (r DiffResult) OK() bool {
	return len(r.Differing) == 0 && len(r.Missing) == 0
}

// blockKey identifies a block across the two sides of a diff. Index is
// deliberately excluded: the target's own export orders and numbers its blocks
// independently of the artifact's and may not carry every block the artifact
// does, so position can never be what says two blocks are "the same" one.
type blockKey struct{ Kind, Name, VPN string }

func keyOf(b Block) blockKey { return blockKey{b.Kind, b.Name, b.VPN} }

// mergedLine is one line of a merged block together with the Section of the block
// it actually came from.
//
// The section travels WITH THE LINE because the merge is by (Kind, Name, VPN) and
// several different sections legitimately produce blocks under one identity:
// `Configure Routing` and `Configure Routing MNR` both open with a bare `routing`
// line (semp/all.cli:134 and :143), and a DMR-configured broker adds a third. The
// merged block used to keep only the LAST section seen, and importIgnore then
// applied that one section's disposition to every line in it -- so one unclassified
// or skipped sibling silenced the other sections' lines in the diff, dropping them
// out of the only error detection this feature has.
type mergedLine struct {
	text    string
	section string
}

// mergedBlock is every block sharing one identity, concatenated in capture order.
type mergedBlock struct {
	key             blockKey
	Kind, Name, VPN string
	lines           []mergedLine
}

// texts is the merged block's lines without their section attribution, for the
// structural pass that does not care where a line came from.
func (m mergedBlock) texts() []string {
	out := make([]string, len(m.lines))
	for i, ml := range m.lines {
		out[i] = ml.text
	}
	return out
}

// DiffBlocks compares want (the parsed import artifact) against got (a Capture
// freshly read back from the target) and reports, per block identity, whether
// the target ended up carrying what the artifact asked for.
//
// ignore excludes one line of one want block from the comparison -- a whole
// section the import deliberately did not apply, or a single line an
// apply-filtered rule dropped (see internal/broker/sections.go) -- so that lines
// the artifact never intended to land do not read as a difference. A block whose
// every line is ignored is left out of the result entirely: it adds to neither
// Matched (nothing about it was checked) nor Differing/Missing (it was never
// supposed to land, so its absence on the target proves nothing). A nil ignore
// excludes nothing.
//
// Three properties make the comparison correct, and a naive implementation gets
// all three wrong:
//
//  1. It is DIRECTIONAL. Every non-ignored line of want's block must be present
//     in got's block, but got is free to carry lines want does not: the target's
//     own `show current-config` fills in defaults the artifact never mentioned,
//     and set equality would flag every one of those as a false failure.
//  2. Blocks sharing one identity are MERGED before comparing. The same
//     message-vpn appears as several separate column-0 blocks in a real capture
//     -- a one-line `create message-vpn "X"` in the broker-level "Create All"
//     summary, a second pragma-guarded create, and the real attribute-carrying
//     `message-vpn "X"` block later on -- and semp/all.cli:28,33468,33473 is
//     exactly that repetition for "xps-ps-01". Comparing against whichever block
//     is found first under a key would most likely land on one of the near-empty
//     ones and read every real attribute as missing.
//  3. Lines are compared WITHIN THEIR SUB-OBJECT, not as a flat bag over the whole
//     block. A column-0 block can hold hundreds of independent nested objects: in
//     the real capture, all 901 queues of `VPN WITH MANY QUEUES` are indented
//     children of ONE `message-spool message-vpn "VPN WITH MANY QUEUES"` opener.
//     Flat set membership asks only "does SOME line in this block have this text",
//     so if the broker rejected one line of QUEUE-042 while its 900 siblings
//     carried the identical setting, every wanted line was still found somewhere
//     and the block counted as matched. Qualifying each line with the openers
//     enclosing it makes that mismatch visible, because QUEUE-042's path differs
//     from its siblings'. Without this, the feature's only error detection is
//     blind to exactly the per-object failure an import produces.
func DiffBlocks(want, got []Block, ignore func(b Block, line string) bool) DiffResult {
	if ignore == nil {
		ignore = func(Block, string) bool { return false }
	}

	wantBlocks := mergeByIdentity(want)
	gotIndex := indexByIdentity(mergeByIdentity(got))

	var result DiffResult
	for _, w := range wantBlocks {
		keys := qualifiedKeys(w.texts())

		// Each line is judged by the disposition of ITS OWN section, not the merged
		// block's: see mergedLine.
		var effective []int
		for i, ml := range w.lines {
			if keys[i] == "" || ignore(Block{Section: ml.section, Kind: w.Kind, Name: w.Name, VPN: w.VPN}, ml.text) {
				continue
			}
			effective = append(effective, i)
		}
		if len(effective) == 0 {
			// Every line here was deliberately not applied -- a skipped section, or
			// a block whose only content an apply-filtered rule dropped. Its
			// absence from the target proves nothing, so it does not enter the
			// result at all: not Matched (nothing was checked), not Missing or
			// Differing (nothing was ever supposed to land).
			continue
		}

		g, found := gotIndex[w.key]
		bd := BlockDiff{Section: w.lines[effective[0]].section, Kind: w.Kind, Name: w.Name, VPN: w.VPN}

		if !found {
			bd.Absent = true
			for _, i := range effective {
				bd.Missing = append(bd.Missing, w.lines[i].text)
			}
			result.Missing = append(result.Missing, bd)
			continue
		}

		missing, failedIn := missingUnder(w, keys, effective, keySet(qualifiedKeys(g.texts())))
		if len(missing) > 0 {
			// The SECTION column names the section that actually failed, not
			// whichever one happened to merge last.
			bd.Missing, bd.Section = missing, failedIn
			result.Differing = append(result.Differing, bd)
			continue
		}
		result.Matched++
	}
	return result
}

// missingUnder returns the want lines (by index in effective) whose qualified key
// is absent from present.
//
// A line whose ENCLOSING sub-object is already reported is left out: when a
// `create queue "Q1" ...` opener never landed, naming it once is the actionable
// fact, and repeating its 25 attribute lines underneath would bury it. So the
// highest missing ancestor is reported and its descendants are suppressed --
// which is also what keeps a rejected queue from producing 25 rows in the report
// that all describe one failure.
// failedIn is the Section of the FIRST line reported missing, returned alongside
// rather than looked up again afterwards -- a second pass matching on line text
// would need a fallback for the case where it found nothing, which is unreachable
// and so could never be tested.
func missingUnder(w mergedBlock, keys []string, effective []int, present map[string]bool) (missing []string, failedIn string) {
	reported := make(map[string]bool, len(effective))
	for _, i := range effective {
		k := keys[i]
		if reported[k] || present[k] || ancestorReported(k, reported) {
			continue
		}
		if missing == nil {
			failedIn = w.lines[i].section
		}
		missing = append(missing, w.lines[i].text)
		reported[k] = true
	}
	return missing, failedIn
}

// ancestorReported reports whether any strict prefix of k -- an enclosing
// sub-object -- has already been named as missing.
func ancestorReported(k string, reported map[string]bool) bool {
	for i := strings.LastIndexByte(k, 0); i > 0; i = strings.LastIndexByte(k[:i], 0) {
		if reported[k[:i]] {
			return true
		}
	}
	return false
}

// mergeByIdentity combines blocks that share a (Kind, Name, VPN) identity into
// one, concatenating their lines in capture order and keeping each line's own
// originating Section (see mergedLine for why that attribution has to survive the
// merge). The returned slice preserves the order identities were first seen, so a
// report built from it reads in roughly the order the artifact does.
func mergeByIdentity(blocks []Block) []mergedBlock {
	order := make([]blockKey, 0, len(blocks))
	merged := make(map[blockKey]*mergedBlock, len(blocks))
	for _, b := range blocks {
		k := keyOf(b)
		m, ok := merged[k]
		if !ok {
			m = &mergedBlock{key: k, Kind: b.Kind, Name: b.Name, VPN: b.VPN}
			merged[k] = m
			order = append(order, k)
		}
		for _, line := range b.Lines {
			m.lines = append(m.lines, mergedLine{text: line, section: b.Section})
		}
	}
	out := make([]mergedBlock, 0, len(order))
	for _, k := range order {
		out = append(out, *merged[k])
	}
	return out
}

// indexByIdentity builds the (Kind, Name, VPN) lookup DiffBlocks matches against.
// blocks must already be merged (mergeByIdentity): two entries sharing a key here
// would silently let the second overwrite the first.
func indexByIdentity(blocks []mergedBlock) map[blockKey]mergedBlock {
	idx := make(map[blockKey]mergedBlock, len(blocks))
	for _, b := range blocks {
		idx[b.key] = b
	}
	return idx
}

// qualifiedKeys returns one comparison key per input line, or "" for a line that
// carries nothing to verify (blank, or whitespace-only after normalising).
//
// A key is the line's own normalised text prefixed by the normalised text of every
// opener enclosing it, NUL-joined. That prefix is what distinguishes
// ` no shutdown egress` under QUEUE-042 from the identical line under QUEUE-043,
// and it is derived the same way blocks.go derives block structure: from
// indentation, never by matching `exit`. A line at indent n is enclosed by the
// nearest preceding line at an indent strictly less than n.
//
// A `! pragma:` comment held at column 0 lands at top level and so never becomes
// the parent of the opener that follows it, which is what keeps a pragma-guarded
// block's children keyed identically to an unguarded one's.
func qualifiedKeys(lines []string) []string {
	out := make([]string, len(lines))
	type frame struct {
		indent int
		norm   string
	}
	var stack []frame
	var b strings.Builder
	for i, line := range lines {
		norm := normalizeLine(line)
		if strings.TrimSpace(norm) == "" {
			continue
		}
		indent := len(norm) - len(strings.TrimLeft(norm, " \t"))
		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}
		b.Reset()
		for _, f := range stack {
			b.WriteString(f.norm)
			b.WriteByte(0)
		}
		b.WriteString(norm)
		out[i] = b.String()
		stack = append(stack, frame{indent, norm})
	}
	return out
}

// keySet collapses qualifiedKeys output into the membership set the diff checks
// against, dropping the empty keys that stand for lines carrying no configuration.
func keySet(keys []string) map[string]bool {
	set := make(map[string]bool, len(keys))
	for _, k := range keys {
		if k != "" {
			set[k] = true
		}
	}
	return set
}

// internalSpaceRunRE matches a run of two or more spaces or tabs that is not the
// line's own leading indentation -- see normalizeLine.
var internalSpaceRunRE = regexp.MustCompile(`[ \t]{2,}`)

// normalizeLine puts a captured line into the form DiffBlocks compares: trailing
// whitespace trimmed, and runs of internal whitespace collapsed to one space.
// Real captures need this -- `system scaling  scale  max-connections ...` and
// `service smf listen-port "55003"  "compressed"` both carry double spaces
// between operands (semp/all.cli) -- and nothing guarantees a target's own
// export reproduces the same run length.
//
// Leading indentation is preserved exactly rather than collapsed: blocks.go
// derives block structure from indentation depth, so treating "  x" and "    x"
// as equal here would erase a real structural difference to fix an accidental
// formatting one. Case and quoting are left untouched on purpose -- `"55003"`
// and `55003` are different tokens to the broker, and a difference there may be
// real.
func normalizeLine(line string) string {
	trimmed := strings.TrimRight(line, " \t\r")
	lead := len(trimmed) - len(strings.TrimLeft(trimmed, " \t"))
	return trimmed[:lead] + internalSpaceRunRE.ReplaceAllString(trimmed[lead:], " ")
}

// ---- report rendering ----

// diffTableHeader is the column set RenderDiffResult's table uses for both the
// absent and the differing blocks, so the two sets of rows line up under one
// header rather than two.
var diffTableHeader = []string{"KIND", "NAME", "VPN", "SECTION"}

// RenderDiffResult writes the stdout verification report for d: a summary block,
// and -- only when something needs attention -- a tagged table naming every
// block that came back absent or short, followed by the exact lines each one is
// missing. The exact lines are what make the report actionable: a table row can
// say a block differs, but only the lines themselves say what to go fix.
func RenderDiffResult(r *output.Sink, d DiffResult) {
	r.Section("Configuration Verification")
	r.KVBlock([]output.KV{
		{Key: "matched", Value: fmt.Sprintf("%d block(s)", d.Matched)},
		{Key: "absent", Value: fmt.Sprintf("%d block(s)", len(d.Missing))},
		{Key: "differing", Value: fmt.Sprintf("%d block(s)", len(d.Differing))},
	})
	if d.OK() {
		return
	}
	r.Line("")

	rows := make([]output.Row, 0, len(d.Missing)+len(d.Differing))
	for _, bd := range d.Missing {
		rows = append(rows, diffRow(bd))
	}
	for _, bd := range d.Differing {
		rows = append(rows, diffRow(bd))
	}
	r.TaggedTable(diffTableHeader, rows)
	r.Line("")

	for _, bd := range d.Missing {
		reportBlockLines(r, bd, "absent from the target")
	}
	for _, bd := range d.Differing {
		reportBlockLines(r, bd, "missing these line(s) on the target")
	}
}

// diffRow renders one BlockDiff as a TaggedTable row. Both Absent and a partial
// mismatch tag [FAIL]: either way the target does not carry what the artifact
// asked for, and there is no third outcome a verification row could report.
func diffRow(bd BlockDiff) output.Row {
	return output.Row{Level: output.LevelFail, Cells: []string{
		orDash(bd.Kind), orDash(bd.Name), orDash(bd.VPN), orDash(bd.Section),
	}}
}

// reportBlockLines prints one problem block's exact missing lines, indented
// under a label naming the block and why it is here -- the detail a table row
// has no room for.
func reportBlockLines(r *output.Sink, bd BlockDiff, why string) {
	r.Line("%s: %s", blockLabel(bd), why)
	for _, line := range bd.Missing {
		r.Line("    %s", line)
	}
	r.Line("")
}

// blockLabel names a block for the report body: `<kind> "<name>"` with the VPN
// qualifier appended when it differs from the name (a VPN-scoped object rather
// than the VPN block itself), falling back to the section comment for a bare
// one-liner that carries neither a kind nor a name.
func blockLabel(bd BlockDiff) string {
	label := bd.Kind
	if bd.Name != "" {
		label += fmt.Sprintf(" %q", bd.Name)
	}
	if label == "" {
		label = bd.Section
	}
	if bd.VPN != "" && bd.VPN != bd.Name {
		label += fmt.Sprintf(" (message-vpn %q)", bd.VPN)
	}
	return label
}

// orDash fills an empty table cell so a broker-level block's blank Name/VPN
// columns read as "nothing here" rather than as a ragged gap.
func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

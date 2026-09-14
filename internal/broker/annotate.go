package broker

import (
	"fmt"
	"sort"
	"strings"
)

// This file is the artifact's own marker layer: the `! solace-util` lines
// export writes and import reads.
//
// Why mark at all. A raw `show current-config` is readable but tells you nothing
// about what an import will do with it, and it makes import infer three things it
// could simply be told: where the create-VPN pass ends and broker-level begins,
// where each section starts and stops, and which broker this came from. The
// markers say all three, so import stops guessing and an operator can read the
// plan straight out of the file.
//
// Two rules keep the layer safe, and both matter:
//
//   - BOUNDARIES ARE AUTHORITATIVE, DISPOSITION IS ADVISORY. A marker may say what
//     this build WOULD do with a section, and import shows that to the operator --
//     but import always re-resolves the disposition from sections.go and warns when
//     the two disagree. If the file's own disposition were honoured, editing one
//     line in an artifact would re-enable `Configure SEMP Service` and sever the
//     channel the import runs over mid-run. That is exactly the class of hazard
//     internal/config/execguard.go exists to prevent: config text choosing what
//     executes. It also means a corrected rule applies to artifacts already taken.
//   - MARKERS ARE COMMENTS IN A RESERVED NAMESPACE. The broker's own output is full
//     of `!` comments and its `! pragma:` lines are load-bearing, so ours are
//     namespaced to `! solace-util` and the parser routes them to metadata: never a
//     section, never a block boundary, never block content.
//
// Sections carry BEGIN and END markers rather than begin-only. A begin-only scheme
// makes the last section in a region run to whatever follows, so a truncated
// artifact and a complete one differ only in a terminator nobody checked -- the
// same failure `! END` exists to catch at the file level.

const (
	markerPrefix  = "! solace-util"
	markerVersion = 1

	markerExport       = "solace-util-export"
	markerExportEnd    = "solace-util-export-end"
	markerMeta         = "solace-util-meta"
	markerOmitted      = "solace-util-omitted"
	markerRegionBegin  = "solace-util-region-begin"
	markerRegionEnd    = "solace-util-region-end"
	markerSectionBegin = "solace-util-section-begin"
	markerSectionEnd   = "solace-util-section-end"
)

// isMarker reports whether a comment belongs to this tool's namespace.
func isMarker(trimmed string) bool { return strings.HasPrefix(trimmed, markerPrefix) }

// markerVerb splits a marker into its verb and payload: `! solace-util-x: rest`.
func markerVerb(trimmed string) (verb, payload string) {
	body := strings.TrimSpace(strings.TrimPrefix(trimmed, "!"))
	verb, payload, _ = strings.Cut(body, ":")
	return strings.TrimSpace(verb), strings.TrimSpace(payload)
}

// readMarker records what a marker says about the capture as a whole.
//
// Unknown verbs are ignored rather than refused: a newer build may write markers
// this one does not know, and the artifact is still perfectly parseable without
// them because the broker's own structure is intact underneath.
func readMarker(c *Capture, trimmed string) {
	verb, payload := markerVerb(trimmed)
	switch verb {
	case markerExport:
		c.Marked = true
		c.Meta["export-version"] = payload
	case markerMeta:
		for k, v := range parseKV(payload) {
			c.Meta[k] = v
			applyMetaField(c, k, v)
		}
	case markerOmitted:
		kv := parseKV(payload)
		c.Omitted = append(c.Omitted, Omission{Section: kv["section"], Reason: kv["reason"]})
	}
}

// applyMetaField lets our own header stand in for the broker's.
//
// It matters because the broker's header layout is not ours to depend on: reading
// the type from a field we wrote ourselves is what makes the cross-type refusal
// robust across broker versions. The broker's own header is still parsed first, so
// this only fills or confirms.
func applyMetaField(c *Capture, key, value string) {
	switch key {
	case "broker-type":
		c.Type = BrokerType(value)
	case "semp-schema":
		c.SEMPSchema = value
		if c.Type == BrokerUnknown {
			c.Type = brokerTypeFromSchema(value)
		}
	case "version":
		c.Version = value
	case "redacted":
		c.Redacted = strings.EqualFold(value, "true")
	}
}

// markerRegion reads a region fence. An end fence returns the empty region, which
// is what makes termination explicit rather than implied by the next begin.
func markerRegion(trimmed string) (Region, bool) {
	switch verb, payload := markerVerb(trimmed); verb {
	case markerRegionBegin:
		return Region(payload), true
	case markerRegionEnd:
		return "", true
	}
	return "", false
}

// markerSection reads a section fence, returning the section name a following
// block belongs to. An end fence clears it.
func markerSection(trimmed string) (string, bool) {
	switch verb, payload := markerVerb(trimmed); verb {
	case markerSectionBegin:
		return parseKV(payload)["name"], true
	case markerSectionEnd:
		return "", true
	}
	return "", false
}

// parseKV parses `key=value key="value with spaces"` into a map.
//
// It is quote-aware for the same reason the opener tokeniser is: a section name is
// a value here (`name="Configure SEMP Service"`), and splitting on whitespace would
// truncate it to its first word.
// It also undoes the %q escaping the writer applies, which matters because a real
// section name is itself quoted: `Create Redundancy PSK: "#redundancy-group"`
// (semp/all.cli:340) is written as `section="Create Redundancy PSK:
// \"#redundancy-group\""`. Trimming quotes off that without unescaping returned a
// truncated, backslash-bearing name, and that name is what the plan's "omitted at
// export" row shows the operator.
func parseKV(s string) map[string]string {
	out := map[string]string{}
	for _, tok := range splitKV(s) {
		if k, v, ok := strings.Cut(tok, "="); ok {
			out[strings.TrimSpace(k)] = unquoteValue(v)
		}
	}
	return out
}

// unquoteValue reverses renderKV's %q for a marker value: strip the surrounding
// quotes, then unescape `\"` and `\\`. strconv.Unquote is deliberately not used --
// it rejects a value that is not a fully valid Go literal, and a hand-edited
// artifact is exactly the input this has to survive rather than drop.
func unquoteValue(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && strings.HasPrefix(v, `"`) && strings.HasSuffix(v, `"`) {
		v = v[1 : len(v)-1]
	}
	var b strings.Builder
	for i := 0; i < len(v); i++ {
		if v[i] == '\\' && i+1 < len(v) && (v[i+1] == '"' || v[i+1] == '\\') {
			i++
		}
		b.WriteByte(v[i])
	}
	return b.String()
}

// splitKV splits on whitespace, except inside a double-quoted run.
//
// A `\"` is NOT a quote for this purpose. Toggling on it split a value at the
// first space after an escaped quote, so `name="Create All: \"Message Vpns\""`
// became three tokens and the name came back as `Create All: \"Message`. The
// writer escapes with %q, so the reader has to honour the same escape.
func splitKV(s string) []string {
	var (
		out    []string
		cur    strings.Builder
		inQuot bool
		esc    bool
	)
	for _, r := range s {
		switch {
		case esc:
			esc = false
			cur.WriteRune(r)
		case r == '\\' && inQuot:
			esc = true
			cur.WriteRune(r)
		case r == '"':
			inQuot = !inQuot
			cur.WriteRune(r)
		case (r == ' ' || r == '\t') && !inQuot:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

// StripMarkers removes every `! solace-util` line.
//
// Verification diffs an artifact against a freshly captured one, and only one of
// the two carries markers at the moment the comparison happens -- so they are
// stripped first, or every marker would read as a missing line.
// It joins rather than appending "\n" per line, which matters: splitLines yields a
// trailing empty element for input that ends in a newline, so appending added a
// second one every time and StripMarkers was not idempotent. Joining reproduces
// the input byte-for-byte when there is nothing to strip.
func StripMarkers(raw []byte) []byte {
	lines := splitLines(raw)
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if isMarker(strings.TrimSpace(line)) {
			continue
		}
		kept = append(kept, line)
	}
	return []byte(strings.Join(kept, "\n"))
}

// Annotate re-renders a parsed capture as a marked artifact.
//
// It rebuilds rather than splicing markers into the original bytes, which makes
// the output canonical: two exports of the same configuration produce
// byte-identical artifacts, so the verification diff compares like with like and a
// golden test can pin the whole thing. The cost is that the broker's own blank-line
// and trailing-colon habits are normalised away -- acceptable, because this
// artifact is consumed by this tool rather than replayed by hand.
//
// scope is recorded for the reader; import re-derives what it needs from the
// blocks themselves rather than trusting it.
func Annotate(c *Capture, scope string, stamp string) []byte {
	var b strings.Builder

	fmt.Fprintf(&b, "! %s: %d\n", markerExport, markerVersion)
	fmt.Fprintf(&b, "! %s: %s\n", markerMeta, renderKV(map[string]string{
		"broker-type": string(c.Type),
		"semp-schema": c.SEMPSchema,
		"version":     c.Version,
		"redacted":    fmt.Sprintf("%t", c.Redacted),
		"scope":       scope,
		"exported-at": stamp,
	}))

	// What export REMOVED, one line each, so the artifact states its own omissions.
	// These are the sections import would never apply (see omitAtExport); writing
	// them here is what stops a reader mistaking a deliberate absence for a broker
	// that simply had no such configuration.
	for _, om := range c.Omitted {
		fmt.Fprintf(&b, "! %s: %s\n", markerOmitted, renderKV(map[string]string{
			"section": om.Section,
			"reason":  om.Reason,
		}))
	}

	// The broker's own header is kept verbatim underneath ours. It is the provenance
	// Solace support would ask for, and re-deriving it later is impossible.
	for _, line := range c.Preamble {
		b.WriteString(line)
		b.WriteString("\n")
	}

	writeBlocks(&b, c.Blocks)
	b.WriteString("\n! END\n")
	fmt.Fprintf(&b, "! %s: %d\n", markerExportEnd, markerVersion)
	return []byte(b.String())
}

// writeBlocks emits the blocks wrapped in region and section fences, opening and
// closing each as the run changes.
func writeBlocks(b *strings.Builder, blocks []Block) {
	var region Region
	var section string

	closeSection := func() {
		if section != "" {
			fmt.Fprintf(b, "! %s: name=%q\n", markerSectionEnd, section)
			section = ""
		}
	}
	closeRegion := func() {
		closeSection()
		if region != "" {
			fmt.Fprintf(b, "! %s: %s\n", markerRegionEnd, region)
			region = ""
		}
	}

	for _, blk := range blocks {
		if blk.Region != region {
			closeRegion()
			region = blk.Region
			fmt.Fprintf(b, "\n! %s: %s\n", markerRegionBegin, region)
		}
		if blk.Section != section {
			closeSection()
			section = blk.Section
			fmt.Fprintf(b, "%s\n", sectionBeginMarker(section))
			fmt.Fprintf(b, "! %s:\n", section)
		}
		for _, line := range blk.Lines {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	closeRegion()
}

// sectionBeginMarker renders a section fence, carrying the ADVISORY disposition
// this build would apply. Import re-resolves it and warns on disagreement -- see
// the file comment for why it is never honoured from the file.
func sectionBeginMarker(section string) string {
	if rule, ok := RuleFor(section); ok {
		if rule.Interrupts != "" {
			return fmt.Sprintf("! %s: name=%q advisory=%s interrupts=%q",
				markerSectionBegin, section, rule.Disposition, rule.Interrupts)
		}
		return fmt.Sprintf("! %s: name=%q advisory=%s", markerSectionBegin, section, rule.Disposition)
	}
	return fmt.Sprintf("! %s: name=%q advisory=unclassified", markerSectionBegin, section)
}

// renderKV renders a metadata payload with keys in a stable order, so the artifact
// is reproducible and diffable.
func renderKV(kv map[string]string) string {
	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		if kv[k] == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%q", k, kv[k]))
	}
	return strings.Join(parts, " ")
}

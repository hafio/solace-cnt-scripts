package broker

import (
	"reflect"
	"strings"
	"testing"
)

// annotate_test.go exercises the `! solace-util` marker layer (annotate.go):
// Annotate/StripMarkers as the round-trip pair, and the small parsing/rendering
// helpers underneath them. The headline property is the round trip -- an export
// that cannot be re-parsed into the same blocks cannot be imported -- and the
// rest pin the design rule the file's own comment states: boundaries (the
// region/section fences) are authoritative, disposition (the advisory= field) is
// never trusted back from the file. loadSample/blockByIndex come from
// blocks_test.go; this file declares no fixture of its own.
//
// Where a test's name says BUG, the assertion documents the CORRECT behaviour
// per StripMarkers' own doc comment and is expected to fail against the current
// implementation -- this file owns only the test suite, not annotate.go, so the
// bug is reported rather than fixed.

// --- Annotate / ParseBlocks round trip --------------------------------------

// TestAnnotateRoundTripPreservesBlocks is the headline property: Annotate then
// ParseBlocks must yield the same blocks, same count and same
// Kind/Name/VPN/Section/Region per block. The sample fixture exercises every
// region (vpn-create, broker, vpn) and a section whose own comment embeds a
// quoted operand, so this is not just the happy path of one flat region.
func TestAnnotateRoundTripPreservesBlocks(t *testing.T) {
	orig := loadSample(t)
	out := Annotate(orig, "full", "2026-01-01T00:00:00Z")

	got, err := ParseBlocks(out)
	if err != nil {
		t.Fatalf("ParseBlocks(Annotate output): %v", err)
	}
	if len(got.Blocks) != len(orig.Blocks) {
		t.Fatalf("round trip changed block count: got %d, want %d (annotated output:\n%s)",
			len(got.Blocks), len(orig.Blocks), out)
	}
	for i := range orig.Blocks {
		o, g := orig.Blocks[i], got.Blocks[i]
		if o.Kind != g.Kind || o.Name != g.Name || o.VPN != g.VPN || o.Section != g.Section || o.Region != g.Region {
			t.Errorf("block %d round trip mismatch:\n  orig = {Kind:%q Name:%q VPN:%q Section:%q Region:%q}\n  got  = {Kind:%q Name:%q VPN:%q Section:%q Region:%q}",
				i+1, o.Kind, o.Name, o.VPN, o.Section, o.Region, g.Kind, g.Name, g.VPN, g.Section, g.Region)
		}
	}
}

// TestAnnotateRoundTripCaptureIsMarked covers the other half of the round trip:
// the re-parsed Capture must itself say it is a marked artifact (Marked, plus
// the Meta fields import reads), not just contain the same blocks. An importer
// that cannot tell a marked artifact from a raw capture cannot decide whether to
// trust the advisory disposition it carries at all.
func TestAnnotateRoundTripCaptureIsMarked(t *testing.T) {
	orig := loadSample(t)
	out := Annotate(orig, "broker-only", "2026-01-01T00:00:00Z")

	got, err := ParseBlocks(out)
	if err != nil {
		t.Fatalf("ParseBlocks(Annotate output): %v", err)
	}
	if !got.Marked {
		t.Error("a round-tripped artifact must parse back with Marked == true")
	}
	for _, key := range []string{"broker-type", "semp-schema", "redacted", "scope"} {
		if _, ok := got.Meta[key]; !ok {
			t.Errorf("Meta[%q] missing after round trip; Meta = %v", key, got.Meta)
		}
	}
	if got.Meta["scope"] != "broker-only" {
		t.Errorf("Meta[scope] = %q, want %q", got.Meta["scope"], "broker-only")
	}
}

// TestAnnotateIsDeterministic pins that two calls with the same input and stamp
// produce byte-identical output. Verification diffs a freshly captured artifact
// against a stored one, and a non-deterministic renderer would make that diff
// noise regardless of whether the configuration actually changed.
func TestAnnotateIsDeterministic(t *testing.T) {
	c := loadSample(t)
	a := Annotate(c, "full", "2026-01-01T00:00:00Z")
	b := Annotate(c, "full", "2026-01-01T00:00:00Z")
	if string(a) != string(b) {
		t.Error("Annotate must be deterministic for the same Capture and stamp, or a " +
			"verification diff against a re-export of the identical configuration " +
			"would show a spurious difference")
	}
}

// TestAnnotateRegionAndSectionFencesAreBalanced covers the fencing design rule
// directly: every region and every section this build opens must also close,
// nested correctly, with nothing left open at end of file. A begin-only scheme
// would let the LAST region or section run to whatever follows it -- the same
// failure the capture-level "! END" terminator exists to catch -- and that
// would only show up on a file whose last section is trimmed differently than
// its others, so it is checked structurally here rather than by eyeballing one
// fixture's tail.
func TestAnnotateRegionAndSectionFencesAreBalanced(t *testing.T) {
	c := loadSample(t)
	out := Annotate(c, "full", "2026-01-01T00:00:00Z")
	assertFencesBalanced(t, out)
}

// assertFencesBalanced walks marker lines in Annotate output and checks that
// region and section fences nest correctly: a section always closes before its
// enclosing region does, every begin has a matching end, and nothing is left
// open at end of file.
func assertFencesBalanced(t *testing.T, out []byte) {
	t.Helper()
	var regionOpen, sectionOpen bool
	nRegionBegin, nRegionEnd, nSectionBegin, nSectionEnd := 0, 0, 0, 0

	for _, line := range splitLines(out) {
		trimmed := strings.TrimSpace(line)
		if !isMarker(trimmed) {
			continue
		}
		verb, _ := markerVerb(trimmed)
		switch verb {
		case markerRegionBegin:
			if regionOpen {
				t.Fatalf("region-begin with a region already open: %q", trimmed)
			}
			regionOpen = true
			nRegionBegin++
		case markerRegionEnd:
			if !regionOpen {
				t.Fatalf("region-end with no region open: %q", trimmed)
			}
			if sectionOpen {
				t.Fatalf("region-end while a section is still open: %q", trimmed)
			}
			regionOpen = false
			nRegionEnd++
		case markerSectionBegin:
			if !regionOpen {
				t.Fatalf("section-begin outside any region: %q", trimmed)
			}
			if sectionOpen {
				t.Fatalf("section-begin with a section already open: %q", trimmed)
			}
			sectionOpen = true
			nSectionBegin++
		case markerSectionEnd:
			if !sectionOpen {
				t.Fatalf("section-end with no section open: %q", trimmed)
			}
			sectionOpen = false
			nSectionEnd++
		}
	}

	if regionOpen || sectionOpen {
		t.Error("a region or section fence is still open at end of file -- a begin-only " +
			"scheme would let the last one run to whatever follows")
	}
	if nRegionBegin == 0 || nSectionBegin == 0 {
		t.Fatal("expected at least one region and one section fence in the output")
	}
	if nRegionBegin != nRegionEnd {
		t.Errorf("region-begin count %d != region-end count %d", nRegionBegin, nRegionEnd)
	}
	if nSectionBegin != nSectionEnd {
		t.Errorf("section-begin count %d != section-end count %d", nSectionBegin, nSectionEnd)
	}
}

// --- writeBlocks (direct) ----------------------------------------------------

// TestWriteBlocksOpensAndClosesOnChange pins writeBlocks' own structural output
// for a hand-built two-block, two-region input: a region/section change closes
// the old fences before opening the new ones, and the broker's own section
// comment is re-emitted verbatim alongside the advisory marker rather than
// replaced by it -- that restated comment is what a raw broker capture (with no
// markers at all) would show, and it is also what a parser re-reading this
// output falls back to for the section boundary.
func TestWriteBlocksOpensAndClosesOnChange(t *testing.T) {
	blocks := []Block{
		{Section: "Configure System", Region: RegionBroker, Kind: "system", Lines: []string{"system scaling 1000"}},
		{Section: "Create Message Vpn", Region: RegionVPN, VPN: "acme", Kind: "message-vpn", Name: "acme",
			Lines: []string{`message-vpn "acme"`, "  exit"}},
	}
	var b strings.Builder
	writeBlocks(&b, blocks)
	out := b.String()

	for _, want := range []string{
		"! " + markerRegionBegin + ": " + string(RegionBroker),
		"! Configure System:",
		"system scaling 1000",
		"! " + markerSectionEnd + `: name="Configure System"`,
		"! " + markerRegionEnd + ": " + string(RegionBroker),
		"! " + markerRegionBegin + ": " + string(RegionVPN),
		"! Create Message Vpn:",
		`message-vpn "acme"`,
		"  exit",
		"! " + markerSectionEnd + `: name="Create Message Vpn"`,
		"! " + markerRegionEnd + ": " + string(RegionVPN),
	} {
		if !strings.Contains(out, want) {
			t.Errorf("writeBlocks output missing %q; full output:\n%s", want, out)
		}
	}
}

// --- StripMarkers -------------------------------------------------------------

// TestStripMarkersKeepsPragmaAndSectionComments covers the discrimination
// StripMarkers depends on: it must remove every "! solace-util" line and NOTHING
// else. Stripping a broker "! pragma:" line would detach ignore-already-exists
// from the create line it makes replayable; stripping a "! Configure X:" section
// comment would destroy the classification key sections.go matches against.
func TestStripMarkersKeepsPragmaAndSectionComments(t *testing.T) {
	const raw = "! pragma:interpreter:relaxed-db-restore\n" +
		"! solace-util-export: 1\n" +
		"! solace-util-meta: broker-type=\"software\"\n" +
		"! Configure SEMP Service:\n" +
		"ssl tls-session-timeout 86400\n" +
		"! solace-util-section-end: name=\"Configure SEMP Service\"\n" +
		"! END\n"

	got := string(StripMarkers([]byte(raw)))
	for _, want := range []string{
		"! pragma:interpreter:relaxed-db-restore",
		"! Configure SEMP Service:",
		"ssl tls-session-timeout 86400",
		"! END",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("StripMarkers dropped a non-marker line %q; got:\n%s", want, got)
		}
	}
	if strings.Contains(got, "solace-util") {
		t.Errorf("StripMarkers left a marker line behind; got:\n%s", got)
	}
}

// TestStripMarkersNoMarkersPassthrough documents what StripMarkers's own doc
// comment promises: it removes "! solace-util" lines and nothing else, so a
// capture carrying none must come back byte-for-byte unchanged, and stripping
// twice must equal stripping once.
//
// BUG: it does not. StripMarkers rebuilds its output as `line + "\n"` for every
// element of splitLines(raw) (blocks.go), and splitLines is built on
// strings.Split(raw, "\n") -- not strings.Join. Split(s, sep) on any s ending in
// "\n" yields a trailing "" element (Split("a\n", "\n") == []string{"a", ""}),
// and the write loop appends "\n" after that element too, so the output gains
// one extra trailing blank line the input did not have. Every real capture ends
// in a newline (the last line is always "! END\n"), so this fires on ordinary
// input, not an edge case: StripMarkers(raw) is one line longer than raw, and a
// second call is one line longer still, so the function is neither a no-op on
// marker-free input nor idempotent. It also means StripMarkers's own stated
// purpose is undermined by itself: it exists so a verification diff does not
// read a marker as a missing line, but the blank line it now adds is exactly
// that same false difference.
func TestStripMarkersNoMarkersPassthrough(t *testing.T) {
	const raw = "home\nenable\nconfigure\n! END\n"

	once := StripMarkers([]byte(raw))
	if string(once) != raw {
		t.Errorf("StripMarkers on marker-free input = %q, want it unchanged: %q", once, raw)
	}

	twice := StripMarkers(once)
	if string(twice) != string(once) {
		t.Errorf("StripMarkers is not idempotent: stripping again changed %q into %q", once, twice)
	}
}

// --- parseKV / splitKV --------------------------------------------------------

// TestParseKVQuotedValueKeepsSpaces is the trap named in annotate.go's own
// comment: a section name is a value here (name="Configure SEMP Service"), and
// a plain whitespace split would truncate it to "Configure".
func TestParseKVQuotedValueKeepsSpaces(t *testing.T) {
	got := parseKV(`name="Configure SEMP Service"`)
	want := map[string]string{"name": "Configure SEMP Service"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseKV(...) = %v, want %v", got, want)
	}
}

func TestParseKVMultipleTokens(t *testing.T) {
	got := parseKV("a=1 b=2")
	want := map[string]string{"a": "1", "b": "2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf(`parseKV("a=1 b=2") = %v, want %v`, got, want)
	}
}

func TestParseKVEmptyPayload(t *testing.T) {
	if got := parseKV(""); len(got) != 0 {
		t.Errorf(`parseKV("") = %v, want an empty map`, got)
	}
}

// TestParseKVIgnoresTokenWithNoEquals covers the malformed-token case: a bare
// word with no "=" is dropped rather than erroring, so one corrupt field in a
// meta line does not fail the whole read.
func TestParseKVIgnoresTokenWithNoEquals(t *testing.T) {
	got := parseKV("foo bar=1")
	want := map[string]string{"bar": "1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf(`parseKV("foo bar=1") = %v, want %v`, got, want)
	}
}

// TestSplitKVQuoteAware covers the tokenizer parseKV depends on: whitespace
// splits tokens except inside a double-quoted run, and a tab separates tokens
// exactly like a space does.
func TestSplitKVQuoteAware(t *testing.T) {
	got := splitKV(`a="b c" d=e`)
	want := []string{`a="b c"`, "d=e"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf(`splitKV(a="b c" d=e) = %v, want %v`, got, want)
	}

	got = splitKV("a=1\tb=2")
	want = []string{"a=1", "b=2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitKV with a tab separator = %v, want %v", got, want)
	}
}

// TestMarkerValueSurvivesAnEmbeddedQuote is the regression for a marker layer that
// could not carry a real section name.
//
// The writer renders values with %q, which escapes an embedded double quote as
// \", and a real capture's section names are full of them:
// `Create Redundancy PSK: "#redundancy-group"` (semp/all.cli:340) and
// `Create All: "Message Vpns"`. splitKV used to toggle its in-quote state on the
// `"` of a `\"` too, so the value split at the next space and parseKV's quote-trim
// returned a truncated, backslash-bearing string -- which is the name the import
// plan's "omitted at export" row shows the operator.
func TestMarkerValueSurvivesAnEmbeddedQuote(t *testing.T) {
	for _, value := range []string{
		`Create Redundancy PSK: "#redundancy-group"`,
		`Create All: "Message Vpns"`,
		`no quotes at all`,
		`trailing backslash \`,
	} {
		line := renderKV(map[string]string{"section": value, "reason": "because"})
		kv := parseKV(line)
		if kv["section"] != value {
			t.Errorf("round trip of %q gave %q", value, kv["section"])
		}
		if kv["reason"] != "because" {
			t.Errorf("the neighbouring key was corrupted by %q: reason = %q", value, kv["reason"])
		}
	}
}

// --- isMarker / markerVerb / markerRegion / markerSection --------------------

// TestIsMarkerDiscriminatesNamespace is the discrimination the parser depends
// on: our own lines must be recognised, and a broker pragma or section comment
// -- both of which also start with "!" -- must not be mistaken for one.
func TestIsMarkerDiscriminatesNamespace(t *testing.T) {
	for _, trimmed := range []string{
		"! solace-util-export: 1",
		`! solace-util-meta: broker-type="software"`,
		"! solace-util-region-begin: broker",
	} {
		if !isMarker(trimmed) {
			t.Errorf("isMarker(%q) = false, want true", trimmed)
		}
	}
	for _, trimmed := range []string{
		"! pragma:interpreter:relaxed-db-restore",
		"! Configure SEMP Service:",
	} {
		if isMarker(trimmed) {
			t.Errorf("isMarker(%q) = true, want false -- a broker comment must not be "+
				"routed into our own metadata handling", trimmed)
		}
	}
}

func TestMarkerVerbSplitsVerbAndPayload(t *testing.T) {
	verb, payload := markerVerb(`! solace-util-meta: broker-type="software" scope="full"`)
	if verb != markerMeta {
		t.Errorf("verb = %q, want %q", verb, markerMeta)
	}
	if payload != `broker-type="software" scope="full"` {
		t.Errorf("payload = %q", payload)
	}
}

// TestMarkerVerbNoPayload covers a marker with no ":" at all: strings.Cut
// leaves payload empty rather than panicking or misreading the verb.
func TestMarkerVerbNoPayload(t *testing.T) {
	verb, payload := markerVerb("! solace-util-region-end")
	if verb != markerRegionEnd || payload != "" {
		t.Errorf(`markerVerb("! solace-util-region-end") = (%q, %q), want (%q, "")`, verb, payload, markerRegionEnd)
	}
}

// TestMarkerRegionBeginAndEnd covers the region fence reader, including the
// property named in its doc comment: an end fence must return the empty region
// so termination is explicit, never implied by whatever begin comes next.
func TestMarkerRegionBeginAndEnd(t *testing.T) {
	r, ok := markerRegion("! solace-util-region-begin: broker")
	if !ok || r != RegionBroker {
		t.Errorf("markerRegion(begin) = (%q, %v), want (%q, true)", r, ok, RegionBroker)
	}
	r, ok = markerRegion("! solace-util-region-end: broker")
	if !ok || r != "" {
		t.Errorf(`markerRegion(end) = (%q, %v), want ("", true)`, r, ok)
	}
	if _, ok := markerRegion("! Configure System:"); ok {
		t.Error("markerRegion on a non-region marker should return ok=false")
	}
}

func TestMarkerSectionBeginAndEnd(t *testing.T) {
	name, ok := markerSection(`! solace-util-section-begin: name="Configure SEMP Service" advisory=skip`)
	if !ok || name != "Configure SEMP Service" {
		t.Errorf("markerSection(begin) = (%q, %v), want (%q, true)", name, ok, "Configure SEMP Service")
	}
	name, ok = markerSection(`! solace-util-section-end: name="Configure SEMP Service"`)
	if !ok || name != "" {
		t.Errorf(`markerSection(end) = (%q, %v), want ("", true)`, name, ok)
	}
}

// --- readMarker / applyMetaField ---------------------------------------------

func TestReadMarkerExportSetsMarked(t *testing.T) {
	c := &Capture{Meta: map[string]string{}}
	readMarker(c, "! solace-util-export: 1")
	if !c.Marked {
		t.Error("readMarker should set Marked on an export marker")
	}
	if c.Meta["export-version"] != "1" {
		t.Errorf("Meta[export-version] = %q, want %q", c.Meta["export-version"], "1")
	}
}

func TestReadMarkerMetaAppliesEveryField(t *testing.T) {
	c := &Capture{Meta: map[string]string{}}
	readMarker(c, `! solace-util-meta: broker-type="software" version="10.1.1.1" redacted="true" scope="full"`)
	if c.Type != BrokerSoftware {
		t.Errorf("Type = %q, want %q", c.Type, BrokerSoftware)
	}
	if c.Version != "10.1.1.1" {
		t.Errorf("Version = %q, want %q", c.Version, "10.1.1.1")
	}
	if !c.Redacted {
		t.Error("Redacted should be true")
	}
	if c.Meta["scope"] != "full" {
		t.Errorf("Meta[scope] = %q, want %q", c.Meta["scope"], "full")
	}
}

// TestReadMarkerUnknownVerbIgnored covers the forward-compatibility case named
// in readMarker's doc comment: a newer build's marker verb must be ignored
// rather than refused, since the broker structure underneath is still intact
// without it.
func TestReadMarkerUnknownVerbIgnored(t *testing.T) {
	c := &Capture{Meta: map[string]string{}}
	readMarker(c, "! solace-util-something-a-future-build-invented: x")
	if c.Marked {
		t.Error("an unrecognised marker verb must not set Marked")
	}
	if len(c.Meta) != 0 {
		t.Errorf("an unrecognised marker verb must not touch Meta, got %v", c.Meta)
	}
}

func TestApplyMetaFieldBrokerTypeSetsType(t *testing.T) {
	c := &Capture{Meta: map[string]string{}}
	applyMetaField(c, "broker-type", "software")
	if c.Type != BrokerSoftware {
		t.Errorf("Type = %q, want %q", c.Type, BrokerSoftware)
	}
}

// TestApplyMetaFieldSempSchemaFillsOnlyWhenUnknown pins the rule stated in
// applyMetaField's own comment: the broker's own header is parsed first and
// must win, so semp-schema may only fill Type in, never override it.
func TestApplyMetaFieldSempSchemaFillsOnlyWhenUnknown(t *testing.T) {
	unknown := &Capture{Meta: map[string]string{}}
	applyMetaField(unknown, "semp-schema", "broker/10.1.1.1")
	if unknown.Type != BrokerSoftware {
		t.Errorf("Type = %q, want %q when Type started Unknown", unknown.Type, BrokerSoftware)
	}
	if unknown.SEMPSchema != "broker/10.1.1.1" {
		t.Errorf("SEMPSchema = %q, want %q", unknown.SEMPSchema, "broker/10.1.1.1")
	}

	known := &Capture{Type: BrokerAppliance, Meta: map[string]string{}}
	applyMetaField(known, "semp-schema", "broker/10.1.1.1")
	if known.Type != BrokerAppliance {
		t.Errorf("Type = %q, want it to stay %q -- the broker's own header must win over semp-schema",
			known.Type, BrokerAppliance)
	}
	if known.SEMPSchema != "broker/10.1.1.1" {
		t.Errorf("SEMPSchema = %q, want it recorded even when Type is not derived from it", known.SEMPSchema)
	}
}

func TestApplyMetaFieldVersion(t *testing.T) {
	c := &Capture{Meta: map[string]string{}}
	applyMetaField(c, "version", "10.26.0.8827")
	if c.Version != "10.26.0.8827" {
		t.Errorf("Version = %q, want %q", c.Version, "10.26.0.8827")
	}
}

// TestApplyMetaFieldRedactedCaseInsensitive covers the comparison
// strings.EqualFold performs: any casing of "true" sets Redacted, and anything
// else -- including an empty value -- must not.
func TestApplyMetaFieldRedactedCaseInsensitive(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"false", false},
		{"", false},
	}
	for _, tc := range cases {
		c := &Capture{Meta: map[string]string{}}
		applyMetaField(c, "redacted", tc.value)
		if c.Redacted != tc.want {
			t.Errorf("applyMetaField(redacted=%q).Redacted = %v, want %v", tc.value, c.Redacted, tc.want)
		}
	}
}

// --- sectionBeginMarker / renderKV --------------------------------------------

// TestSectionBeginMarkerAdvisoryDisposition covers what the operator reads a
// section fence for: the classified disposition (with its Interrupts detail
// when it has one), and "unclassified" for a section sections.go has no rule
// for at all -- never a silent omission of the advisory= field.
func TestSectionBeginMarkerAdvisoryDisposition(t *testing.T) {
	cases := []struct {
		section string
		want    string
	}{
		{"Configure System", `! solace-util-section-begin: name="Configure System" advisory=skip`},
		{"Configure SMF Service",
			`! solace-util-section-begin: name="Configure SMF Service" advisory=apply interrupts="all messaging"`},
		{"Totally Unclassified Section",
			`! solace-util-section-begin: name="Totally Unclassified Section" advisory=unclassified`},
	}
	for _, tc := range cases {
		if got := sectionBeginMarker(tc.section); got != tc.want {
			t.Errorf("sectionBeginMarker(%q) = %q, want %q", tc.section, got, tc.want)
		}
	}
}

// TestRenderKVSortsKeys pins the reason renderKV sorts its keys: without a
// stable order, two exports of the same configuration would not be
// byte-identical, and the verification diff Annotate exists to support would
// be comparing arbitrary map iteration order instead of content.
func TestRenderKVSortsKeys(t *testing.T) {
	got := renderKV(map[string]string{"zeta": "1", "alpha": "2", "mid": "3"})
	want := `alpha="2" mid="3" zeta="1"`
	if got != want {
		t.Errorf("renderKV(...) = %q, want %q", got, want)
	}
}

// TestRenderKVSkipsEmptyValues covers the other half of renderKV: a field with
// no value (e.g. an empty scope) is omitted rather than rendered as key="".
func TestRenderKVSkipsEmptyValues(t *testing.T) {
	got := renderKV(map[string]string{"a": "", "b": "x"})
	want := `b="x"`
	if got != want {
		t.Errorf("renderKV(...) = %q, want %q", got, want)
	}
}

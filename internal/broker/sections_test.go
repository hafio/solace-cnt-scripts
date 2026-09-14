package broker

import (
	"reflect"
	"testing"
)

// sections_test.go covers sections.go, the baked-in classification
// `import-config` applies to each broker-level section of a `show
// current-config` capture. Every property here traces to a real risk: a
// misclassified or unmatched section either silently fails to apply
// configuration the operator chose to keep, or applies something (a service
// bounce, a credential overwrite, an HA-breaking change) the operator
// deliberately excluded. See semp/import-decisions.md for the operator's own
// reasoning per section and docs/import.md (generated from Rules() by
// importdoc_test.go) for the published classification this file must agree
// with.
//
// Real line text quoted below is copied by hand from semp/all.cli, cited by
// line number in each test's comment -- that file is never read or copied
// into testdata (see blocks_test.go's own note), and every line reproduced
// here is ordinary config text, never a credential.

// ruleFieldsEqual compares every SectionRule field except DropLines by value,
// and DropLines by function identity. reflect.DeepEqual cannot be used on a
// SectionRule directly: per its own doc, "Func values are deeply equal if
// both are nil; otherwise they are not deeply equal", so two copies of the
// SAME rule (same DropLines closure, copied by value) would report unequal.
func ruleFieldsEqual(a, b SectionRule) bool {
	if a.Section != b.Section || a.Disposition != b.Disposition || a.Reason != b.Reason ||
		a.Interrupts != b.Interrupts || a.First != b.First {
		return false
	}
	if (a.DropLines == nil) != (b.DropLines == nil) {
		return false
	}
	if a.DropLines == nil {
		return true
	}
	return reflect.ValueOf(a.DropLines).Pointer() == reflect.ValueOf(b.DropLines).Pointer()
}

// isZeroSectionRule reports whether r is the zero value RuleFor returns
// alongside ok=false. Written field-by-field because SectionRule embeds a
// func, which makes the struct type itself incomparable with == or !=.
func isZeroSectionRule(r SectionRule) bool {
	return r.Section == "" && r.Disposition == "" && r.Reason == "" &&
		r.Interrupts == "" && !r.First && r.DropLines == nil
}

// countDisposition counts how many rules carry disposition d, so a table test
// against a hand-typed list of section names can also catch one silently
// added to or removed from sectionRules without the list being updated.
func countDisposition(rules []SectionRule, d Disposition) int {
	n := 0
	for _, r := range rules {
		if r.Disposition == d {
			n++
		}
	}
	return n
}

// TestSectionsRuleForExactMatch: every rule Rules() advertises must be found
// again by RuleFor under its own Section, as itself. RISK: if RuleFor ever
// drifted from sectionRules (a rule appended but unreachable, or an earlier
// duplicate Section shadowing a later one) import would silently misclassify
// a real captured section while docs/import.md kept advertising the original
// rule.
func TestSectionsRuleForExactMatch(t *testing.T) {
	for _, want := range Rules() {
		got, ok := RuleFor(want.Section)
		if !ok {
			t.Errorf("RuleFor(%q) not found, but it is in Rules()", want.Section)
			continue
		}
		if !ruleFieldsEqual(got, want) {
			t.Errorf("RuleFor(%q) = %+v, want %+v", want.Section, got, want)
		}
	}
}

// TestSectionsRuleForPrefixMatch pins prefix matching against the full quoted
// section comments the broker actually emits (semp/all.cli:50, 290, 319, 333,
// 340, 351): the operand is deployment-chosen (an LDAP/RADIUS profile name, a
// domain CA name, an authentication profile name, or the one Solace-internal
// singleton name), so equality alone would need one rule per possible name.
// RISK: if prefix matching breaks, a real captured section becomes
// unclassified and import has nothing to do with it -- neither applied nor
// reported skipped, just silently dropped from the run.
func TestSectionsRuleForPrefixMatch(t *testing.T) {
	cases := []struct {
		section string // the full Block.Section text as the broker emits it
		want    string // the canonical (unsuffixed) rule it must resolve to
	}{
		{`Configure ip vrf: "management"`, "Configure ip vrf"},
		{`Create LDAP Profile: "default"`, "Create LDAP Profile"},
		{`Create Radius Profile: "default"`, "Create Radius Profile"},
		{`Create Domain Certificate Authority: "handy-ca"`, "Create Domain Certificate Authority"},
		{`Create Authentication: "miscellaneous"`, "Create Authentication"},
		{`Create Redundancy PSK: "#redundancy-group"`, "Create Redundancy PSK"},
	}
	for _, tc := range cases {
		got, ok := RuleFor(tc.section)
		if !ok {
			t.Errorf("RuleFor(%q) not found", tc.section)
			continue
		}
		if got.Section != tc.want {
			t.Errorf("RuleFor(%q).Section = %q, want %q", tc.section, got.Section, tc.want)
		}
	}
}

// TestSectionsIsPrefixSection pins the membership list RuleFor's second loop
// consults. RISK: a section wrongly added here starts matching every capture
// whose section name happens to start with it (see the shadow test below); a
// section wrongly missing here silently falls back to unclassified for every
// deployment where its operand varies.
func TestSectionsIsPrefixSection(t *testing.T) {
	for _, s := range prefixSections {
		if !isPrefixSection(s) {
			t.Errorf("isPrefixSection(%q) = false, want true (listed in prefixSections)", s)
		}
	}
	notPrefix := []string{
		"",
		"Configure System",
		"Configure Routing", // exact rule; also a strict text-prefix of "Configure Routing MNR"
		"Create Authentication ", // trailing space must not fuzzily match a real entry
	}
	for _, s := range notPrefix {
		if isPrefixSection(s) {
			t.Errorf("isPrefixSection(%q) = true, want false", s)
		}
	}
}

// TestSectionsExactRuleNotShadowedByPrefix exploits the one place in the real
// table where an exact rule's Section text is a strict prefix of a DIFFERENT
// rule's Section text: "Configure Routing" (ApplyFiltered, drops the host
// interface line) is a literal prefix of "Configure Routing MNR" (Apply, no
// filtering). Neither is in prefixSections, so RuleFor must resolve
// "Configure Routing MNR" by equality, not accidentally inherit "Configure
// Routing"'s disposition or line filter. RISK: if RuleFor ever tried prefix
// matching before equality, or "Configure Routing" were ever added to
// prefixSections, the CSPF/queue settings in "Configure Routing MNR"
// (semp/all.cli:142-152) would silently lose lines they never had, or worse,
// resolve to the wrong disposition entirely.
func TestSectionsExactRuleNotShadowedByPrefix(t *testing.T) {
	routing, ok := RuleFor("Configure Routing")
	if !ok || routing.Disposition != ApplyFiltered || routing.DropLines == nil {
		t.Fatalf(`RuleFor("Configure Routing") = %+v, ok=%v; want ApplyFiltered with DropLines set`, routing, ok)
	}
	mnr, ok := RuleFor("Configure Routing MNR")
	if !ok {
		t.Fatal(`RuleFor("Configure Routing MNR") not found`)
	}
	if mnr.Disposition != Apply {
		t.Errorf(`RuleFor("Configure Routing MNR").Disposition = %q, want %q`, mnr.Disposition, Apply)
	}
	if mnr.DropLines != nil {
		t.Error(`RuleFor("Configure Routing MNR") must not inherit "Configure Routing"'s line filter`)
	}
}

// TestSectionsRuleForUnknownAndEmpty: a caller that gets ok=false must be able
// to trust it completely. RISK: a zero-value SectionRule with Disposition ""
// returned alongside ok=true, or a non-zero leftover value alongside ok=false,
// would be neither apply nor skip -- a section that is neither reported nor
// applied, simply forgotten by the report the operator relies on to see
// nothing was dropped silently.
func TestSectionsRuleForUnknownAndEmpty(t *testing.T) {
	for _, section := range []string{
		"",
		"Configure Something Nobody Emits",
		"configure system",  // the broker's own casing is exact; lowercase must not match
		" Configure System", // leading whitespace must not fuzzily match
		"Configure Systems", // one letter off an exact rule; must not partial-match
	} {
		got, ok := RuleFor(section)
		if ok {
			t.Errorf("RuleFor(%q) ok = true, want false", section)
			continue
		}
		if !isZeroSectionRule(got) {
			t.Errorf("RuleFor(%q) = %+v on ok=false, want the zero SectionRule", section, got)
		}
	}
}

// TestSectionsSkipDispositions asserts the exact disposition, by name, for
// every section the operator decided must never be applied (docs/import.md
// "Never applied"; semp/import-decisions.md sections A-D, "absolutely
// unreachable" per Q1's answer). RISK: a mis-typed disposition on "Configure
// SEMP Service" would apply the very section that shuts SEMP down mid-import
// and severs the channel import runs over; on "Create Redundancy PSK" it
// would silently break an HA group the operator's own guide says the
// operator cannot repair.
func TestSectionsSkipDispositions(t *testing.T) {
	want := []string{
		"Configure System",
		"Configure SEMP Service",
		"Configure Matelink Service",
		"Configure Redundancy Service",
		"Configure Router Name",
		"Configure Redundancy",
		"Configure Config Sync",
		"Create Redundancy PSK",
		// Empty in both software captures, so it had no rule until a real appliance
		// capture showed it carrying `hostname "<name>" defer` -- node identity, like
		// Configure Router Name above.
		"Configure Hostname",
	}
	for _, section := range want {
		r, ok := RuleFor(section)
		if !ok {
			t.Errorf("RuleFor(%q) not found", section)
			continue
		}
		if r.Disposition != Skip {
			t.Errorf("RuleFor(%q).Disposition = %q, want %q", section, r.Disposition, Skip)
		}
	}
	if got := countDisposition(Rules(), Skip); got != len(want) {
		t.Errorf("table has %d Skip sections, want exactly %d named above -- a section was "+
			"added to or removed from sectionRules without updating this test", got, len(want))
	}
}

// TestSectionsSkipApplianceDispositions asserts the exact disposition, by
// name, for every appliance-only section (docs/import.md "Appliance only").
// RISK: a section wrongly left as SkipAppliance when it actually carries
// software-relevant content (or the reverse) either drops real configuration
// an operator captured on a software broker, or tries to apply hardware
// config no software target has anywhere to put.
func TestSectionsSkipApplianceDispositions(t *testing.T) {
	want := []string{
		"Configure Ethernet Interfaces",
		"Configure Lag Interfaces",
		"Configure SolOsPhy Interfaces",
		"Configure ip vrf",
		"Configure DNS",
		"Disk",
		"Configure Clock",
		"Configure Clock Synchronization",
		"Configure SNMP",
		// Two more `hardware` sections, found only once a real appliance capture
		// existed. Of the four hardware commands, `message-spool` is on both
		// platforms while these and Disk are appliance-only.
		"Power Redundancy",
		"Topic Routing",
	}
	for _, section := range want {
		r, ok := RuleFor(section)
		if !ok {
			t.Errorf("RuleFor(%q) not found", section)
			continue
		}
		if r.Disposition != SkipAppliance {
			t.Errorf("RuleFor(%q).Disposition = %q, want %q", section, r.Disposition, SkipAppliance)
		}
	}
	if got := countDisposition(Rules(), SkipAppliance); got != len(want) {
		t.Errorf("table has %d SkipAppliance sections, want exactly %d named above -- a section "+
			"was added to or removed from sectionRules without updating this test", got, len(want))
	}
}

// TestSectionsDropRoutingInterfaceLine pins dropRoutingInterfaceLine against
// the real "Configure Routing" block text (semp/all.cli:134-136). RISK: too
// broad and it also drops `mode dynamic-message-routing defer`, silently
// disabling DMR on import; too narrow (e.g. hard-coding the real 2-space
// indent) and it stops matching the moment a capture is re-indented, leaving
// host hardware (`interface "intf0"`, semp/all.cli:135) referenced on a
// target whose interface-creating sections are all skip-appliance.
func TestSectionsDropRoutingInterfaceLine(t *testing.T) {
	drop := []string{
		`  interface "intf0"`, // semp/all.cli:135, exactly as captured (2-space indent)
		`interface "intf0"`,   // the unindented form must also match
	}
	for _, line := range drop {
		if !dropRoutingInterfaceLine(line) {
			t.Errorf("dropRoutingInterfaceLine(%q) = false, want true", line)
		}
	}
	keep := []string{
		`  mode dynamic-message-routing defer`, // semp/all.cli:136
		`mode dynamic-message-routing defer`,
	}
	for _, line := range keep {
		if dropRoutingInterfaceLine(line) {
			t.Errorf("dropRoutingInterfaceLine(%q) = true, want false", line)
		}
	}
}

// TestSectionsDropMessageSpoolLines pins dropMessageSpoolLines against the
// real "Message Spool" block text (semp/all.cli:171-200). RISK: too greedy
// (e.g. matching any line containing "spool-sync" or "spool-usage") drops the
// mirrored spool-sync ack-timeout settings or the cache-usage setting the
// operator chose to keep, silently changing broker tuning nobody asked to
// change; too narrow and it lets `max-spool-usage`, `spool-sync mode` or the
// shutdown/no-shutdown pair back through, which bounces the
// guaranteed-messaging data path and reapplies a spool size Kubernetes
// already sizes from the CR and the PVC.
func TestSectionsDropMessageSpoolLines(t *testing.T) {
	drop := []string{
		"hardware message-spool max-spool-usage 10000",     // semp/all.cli:192
		`hardware message-spool spool-sync mode "mirrored"`, // semp/all.cli:187
		"hardware message-spool shutdown",                  // semp/all.cli:186
		"no hardware message-spool shutdown",                // semp/all.cli:200
	}
	for _, line := range drop {
		if !dropMessageSpoolLines(line) {
			t.Errorf("dropMessageSpoolLines(%q) = false, want true", line)
		}
	}
	keep := []string{
		// an event threshold, kept ordinary broker setting
		"hardware message-spool event egress-flows thresholds set-percentage 80 clear-percentage 60",
		// the defragment schedule, kept ordinary broker settings (semp/all.cli:194, 197)
		"hardware message-spool defragment-spool-files schedule shutdown",
		"hardware message-spool defragment-spool-files threshold shutdown",
		// close to "spool-sync mode" but is "spool-sync mirrored ...": must not match
		"hardware message-spool spool-sync mirrored message-ack-timeout 10000",
		// close to "max-spool-usage" but is "max-cache-usage": must not match
		"hardware message-spool max-cache-usage 10",
	}
	for _, line := range keep {
		if dropMessageSpoolLines(line) {
			t.Errorf("dropMessageSpoolLines(%q) = true, want false", line)
		}
	}
}

// TestSectionsCreateHardwareSharesMessageSpoolPredicate pins the sections.go
// comment that "Create hardware" is "classified with Message Spool for the
// case it is not [empty]": both rules must filter with the IDENTICAL
// predicate, not two separately hand-written copies that could drift apart
// the next time one is edited without the other.
func TestSectionsCreateHardwareSharesMessageSpoolPredicate(t *testing.T) {
	hw, ok := RuleFor("Create hardware")
	if !ok || hw.DropLines == nil {
		t.Fatalf(`RuleFor("Create hardware") = %+v, ok=%v; want ApplyFiltered with DropLines set`, hw, ok)
	}
	spool, ok := RuleFor("Message Spool")
	if !ok || spool.DropLines == nil {
		t.Fatalf(`RuleFor("Message Spool") = %+v, ok=%v; want ApplyFiltered with DropLines set`, spool, ok)
	}
	if reflect.ValueOf(hw.DropLines).Pointer() != reflect.ValueOf(spool.DropLines).Pointer() {
		t.Error(`"Create hardware" and "Message Spool" must share the exact same DropLines predicate`)
	}
}

// TestSectionsFirstIsCreateLoggingOnly strengthens importdoc_test.go's
// TestOnlyOneSectionIsAppliedFirst (which only refuses MORE than one) into an
// exact assertion: exactly one section is First, and it is "Create logging".
// RISK: a second section wrongly marked First would fight over "runs alone,
// in its own RunCLI invocation"; "Create logging" losing First would let a
// later command sharing its script silently never execute, since logging
// changes end the CLI session it needed (semp/all.cli:251-260).
func TestSectionsFirstIsCreateLoggingOnly(t *testing.T) {
	var first []string
	for _, r := range Rules() {
		if r.First {
			first = append(first, r.Section)
		}
	}
	if len(first) != 1 || first[0] != "Create logging" {
		t.Errorf("sections with First=true = %v, want exactly [%q]", first, "Create logging")
	}
}

// TestSectionsRulesOrderStable: docs/import.md is generated from Rules() and
// its row order IS this order (sections.go's own doc comment on sectionRules)
// -- so two calls returning different content or order would make the
// generated docs nondeterministic and the golden comparison in
// TestImportDocs flap independent of any real change. Non-emptiness is
// checked too, since an accidentally emptied table would otherwise pass every
// other test in this file by vacuous truth.
func TestSectionsRulesOrderStable(t *testing.T) {
	a := Rules()
	b := Rules()
	if len(a) == 0 {
		t.Fatal("Rules() returned no sections")
	}
	if len(a) != len(b) {
		t.Fatalf("Rules() length changed between calls: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if !ruleFieldsEqual(a[i], b[i]) {
			t.Errorf("Rules()[%d] differs between calls: %+v vs %+v", i, a[i], b[i])
		}
	}
	// Pin the documented file order at both ends: semp/all.cli's broker-level
	// region runs "Configure System" (line 31) to "Configure Management
	// Message Vpn" (line 377).
	if a[0].Section != "Configure System" {
		t.Errorf("Rules()[0].Section = %q, want %q", a[0].Section, "Configure System")
	}
	if last := a[len(a)-1].Section; last != "Configure Management Message Vpn" {
		t.Errorf("Rules()[len-1].Section = %q, want %q", last, "Configure Management Message Vpn")
	}
}

// TestSectionsRulesReturnsCopy: Rules() hands out sectionRules through
// append([]SectionRule(nil), sectionRules...), a defensive copy of the live
// backing array. RISK: without the copy, one caller (the docs generator, or a
// future report that sorts its own view) mutating its slice would corrupt the
// baked-in classification for every OTHER caller in the same process,
// including a concurrent import run reading the same table.
func TestSectionsRulesReturnsCopy(t *testing.T) {
	a := Rules()
	original := a[0].Section
	a[0].Section = "mutated by test"
	b := Rules()
	if b[0].Section != original {
		t.Errorf("mutating a Rules() result changed a later call: got %q, want %q", b[0].Section, original)
	}
}

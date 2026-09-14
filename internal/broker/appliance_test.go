package broker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// appliance_test.go covers what only a real APPLIANCE capture can settle.
//
// Every classification in sections.go was originally reasoned from software
// captures, where the appliance-only sections are empty -- so "skip-appliance" was
// an assertion about a section nobody had seen carry anything. semp/show-currentconfig-all.out
// is a real appliance (soltr/10_4_1, 9MB, 10 message-VPNs) and turns that reasoning
// into evidence. It also found one section that had no rule at all.
//
// The capture is gitignored (it carries a production configuration), so every test
// here SKIPS when it is absent rather than failing a fresh clone.

// loadApplianceCapture parses the appliance capture, or skips.
func loadApplianceCapture(t *testing.T) *Capture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "semp", "show-currentconfig-all.out"))
	if err != nil {
		t.Skipf("appliance capture not available (%v)", err)
	}
	c, err := ParseBlocks(raw)
	if err != nil {
		t.Fatalf("parse the appliance capture: %v", err)
	}
	return c
}

// TestApplianceCaptureIsReadAsAnAppliance pins the type discrimination on a real
// appliance header.
//
// The pre-10.25.4 generation writes `soltr/<version>` with a `VMR` suffix for
// software and none for an appliance, so these two differ by three characters --
// `soltr/10_4_1` here versus `soltr/10_4_1VMR` in semp/all-cfg-tps-org.cli. Getting
// it wrong is not a cosmetic error: the whole cross-type refusal rests on it, and
// reading an appliance as software is exactly the import that refusal exists to stop.
func TestApplianceCaptureIsReadAsAnAppliance(t *testing.T) {
	c := loadApplianceCapture(t)
	if c.Type != BrokerAppliance {
		t.Errorf("Type = %q, want %q (schema %q)", c.Type, BrokerAppliance, c.SEMPSchema)
	}
	if !strings.HasPrefix(c.SEMPSchema, "soltr/") || strings.HasSuffix(c.SEMPSchema, "VMR") {
		t.Errorf("schema = %q; this fixture is meant to be the no-VMR appliance form", c.SEMPSchema)
	}
}

// TestEveryApplianceBrokerSectionIsClassified is the gap-finder.
//
// An unclassified section with content is never applied and its lines stay in the
// verification diff, so the import exits non-zero -- loud, but it means the
// configuration does not import at all. This capture found exactly one:
// `Configure Hostname`, empty in both software captures and carrying
// `hostname "<name>" defer` here, which is node identity like Configure Router Name.
//
// It runs over the BROKER region only: VPN-region blocks are rendered per VPN and
// are deliberately never classified (planSections only considers RegionBroker), so
// including them would report dozens of false gaps.
func TestEveryApplianceBrokerSectionIsClassified(t *testing.T) {
	c := loadApplianceCapture(t)

	seen := map[string]bool{}
	var unclassified []string
	for _, b := range c.Blocks {
		if b.Region != RegionBroker || b.Section == "" || seen[b.Section] {
			continue
		}
		seen[b.Section] = true
		if _, ok := RuleFor(b.Section); !ok {
			unclassified = append(unclassified, b.Section)
		}
	}
	if len(seen) == 0 {
		t.Fatal("no broker-region sections found; the capture did not parse as expected")
	}
	if len(unclassified) > 0 {
		t.Errorf("%d appliance section(s) have content but no rule, so they would never be "+
			"applied and would fail verification: %v", len(unclassified), unclassified)
	}
}

// TestApplianceOnlySectionsCarryContent is the other half: it proves the
// skip-appliance rules are load bearing rather than speculative.
//
// Each was classified from a SOFTWARE capture, where the section is empty -- so the
// rule rested on the claim that an appliance would put host and interface identity
// there. It does, and this is where that claim stops being a claim. `ip vrf` appears
// twice (management and msg-backbone), which is also what validates matching that
// section by prefix rather than by equality.
func TestApplianceOnlySectionsCarryContent(t *testing.T) {
	c := loadApplianceCapture(t)

	content := map[string]int{}
	vrfs := 0
	for _, b := range c.Blocks {
		if b.Region != RegionBroker {
			continue
		}
		rule, ok := RuleFor(b.Section)
		if !ok || rule.Disposition != SkipAppliance {
			continue
		}
		content[rule.Section] += len(b.Lines)
		if rule.Section == "Configure ip vrf" {
			vrfs++
		}
	}

	for _, want := range []string{
		"Configure Ethernet Interfaces",
		"Configure Lag Interfaces",
		"Configure ip vrf",
		"Configure DNS",
		"Configure SNMP",
	} {
		if content[want] == 0 {
			t.Errorf("skip-appliance section %q carries nothing even on a real appliance; "+
				"either the rule is wrong or this capture does not exercise it", want)
		}
	}
	if vrfs < 2 {
		t.Errorf("found %d `ip vrf` block(s), want at least 2 -- the appliance names each VRF in "+
			"its own section comment, which is why the rule is matched by prefix", vrfs)
	}
}

// TestApplianceReplicationGrammarDiffersFromSoftware records a platform difference
// found in this capture, and the reason it does not affect import today.
//
// At the SAME broker version (10.4.1) the two platforms spell the replication mate
// differently: software writes `replication mate connect-via "host:port" "ssl"` as
// its own repeatable line with the router-name standing alone, while an appliance
// writes explicit `replication mate connect-port <n>` lines and hangs `connect-via`
// off the router-name line as an attribute. Import is unaffected because a
// cross-type artifact is refused outright -- but anything that GENERATES replication
// configuration has to render per type, so this is pinned rather than left to be
// rediscovered.
func TestApplianceReplicationGrammarDiffersFromSoftware(t *testing.T) {
	app := loadApplianceCapture(t)

	var appLines []string
	for _, b := range app.Blocks {
		if b.Section == "Configure Replication" {
			appLines = append(appLines, b.Lines...)
		}
	}
	if len(appLines) == 0 {
		t.Skip("this appliance capture configures no replication")
	}
	joined := strings.Join(appLines, "\n")

	if !strings.Contains(joined, "replication mate connect-port ") {
		t.Errorf("appliance replication has no `mate connect-port` line; the grammar this test "+
			"records may have changed:\n%s", joined)
	}
	if !strings.Contains(joined, "connect-via") {
		t.Errorf("appliance replication has no connect-via:\n%s", joined)
	}
	// The whole section is classified, so it is never the silent-drop case.
	if _, ok := RuleFor("Configure Replication"); !ok {
		t.Error("Configure Replication is unclassified")
	}
}

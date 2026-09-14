package broker

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"solace/internal/config"
)

// The tests here pin where `skip` is enforced: at EXPORT. The operator's rule is
// that export removes a skipped section and import merely ignores one if it meets
// it -- which, after export has done its job, should never happen. So the artifact
// is exactly what an import applies, and the import report mentions a skipped
// section only when an older or hand-edited artifact still carries one.

// TestOmitAtExportRemovesSkipSectionsAndRecordsThem is the headline property. RISK:
// a Skip section left in the artifact is dead weight at best and, to a reader, a
// broker-level setting that looks like it will be applied and never is.
func TestOmitAtExportRemovesSkipSectionsAndRecordsThem(t *testing.T) {
	blocks := []Block{
		{Section: "Configure SEMP Service", Region: RegionBroker, Kind: "service",
			Lines: []string{"service semp listen-port 8080"}},
		{Section: "Configure SEMP Service", Region: RegionBroker, Kind: "service",
			Lines: []string{"service semp listen-port 1943 ssl"}},
		{Section: "Configure Memory Event", Region: RegionBroker, Kind: "memory-event",
			Lines: []string{"memory-event nab-buffer-load-factor thresholds set-percentage 80 clear-percentage 60"}},
		{Section: `Create Client Username: "acme"`, Region: RegionVPN, Kind: "client-username",
			Name: "app1", VPN: "acme", Lines: []string{`client-username "app1" message-vpn "acme"`, "  exit"}},
	}

	kept, omitted := omitAtExport(blocks)

	for _, b := range kept {
		if b.Section == "Configure SEMP Service" {
			t.Errorf("a Skip section survived export: %q", b.Lines[0])
		}
	}
	if len(kept) != 2 {
		t.Errorf("kept %d blocks, want 2 (the Apply section and the VPN block)", len(kept))
	}

	// Recorded once per section, not once per block, with the rule's own reason.
	if len(omitted) != 1 || omitted[0].Section != "Configure SEMP Service" {
		t.Fatalf("omitted = %+v, want exactly one entry for Configure SEMP Service", omitted)
	}
	rule, _ := RuleFor("Configure SEMP Service")
	if omitted[0].Reason != rule.Reason || omitted[0].Reason == "" {
		t.Errorf("omission reason = %q, want the rule's own reason %q", omitted[0].Reason, rule.Reason)
	}
}

// TestOmitAtExportDropsFilteredLinesAndEmptyBlocks covers the apply-filtered
// half. RISK, two ways: the `routing interface "intf0"` line names host hardware
// and must go while the rest of its block stays; and Message Spool's
// `max-spool-usage` is a ONE-LINER, so dropping its only line must drop the block
// -- an opener-less husk would re-render as nothing but still count as a block.
func TestOmitAtExportDropsFilteredLinesAndEmptyBlocks(t *testing.T) {
	blocks := []Block{
		{Section: "Configure Routing", Region: RegionBroker, Kind: "routing",
			Lines: []string{"routing", `  interface "intf0"`, "  mode dynamic-message-routing defer", "  exit"}},
		{Section: "Message Spool", Region: RegionBroker, Kind: "hardware",
			Lines: []string{"hardware message-spool max-spool-usage 10000"}},
		{Section: "Message Spool", Region: RegionBroker, Kind: "hardware",
			Lines: []string{"hardware message-spool event spool-usage thresholds set-percentage 80 clear-percentage 60"}},
	}

	kept, omitted := omitAtExport(blocks)

	if len(omitted) != 0 {
		t.Errorf("apply-filtered sections must not be recorded as omissions, got %+v", omitted)
	}
	if len(kept) != 2 {
		t.Fatalf("kept %d blocks, want 2: routing (filtered) and the spool threshold", len(kept))
	}
	routing := kept[0]
	for _, l := range routing.Lines {
		if strings.Contains(l, `interface "intf0"`) {
			t.Errorf("host-hardware line survived export in %q", routing.Lines)
		}
	}
	if len(routing.Lines) != 3 {
		t.Errorf("routing block = %q, want opener + mode + exit", routing.Lines)
	}
	if strings.Contains(kept[1].Lines[0], "max-spool-usage") {
		t.Errorf("the one-liner max-spool-usage block should have been dropped whole, got %q", kept[1].Lines)
	}
}

// TestOmitAtExportLeavesTargetDecisionsToImport pins what export must NOT remove.
// SkipAppliance is a decision about the TARGET's type, which only import knows;
// an unclassified section is something import refuses loudly, and removing it here
// would hide a broker release this build does not know about.
func TestOmitAtExportLeavesTargetDecisionsToImport(t *testing.T) {
	blocks := []Block{
		{Section: "Configure SNMP", Region: RegionBroker, Kind: "snmp",
			Lines: []string{"snmp community public"}},
		{Section: "Configure Something New", Region: RegionBroker, Kind: "something",
			Lines: []string{"something new 1"}},
	}
	rule, ok := RuleFor("Configure SNMP")
	if !ok || rule.Disposition != SkipAppliance {
		t.Fatalf("sanity: Configure SNMP should be SkipAppliance, got %v %v", rule.Disposition, ok)
	}

	kept, omitted := omitAtExport(blocks)
	if len(kept) != 2 || len(omitted) != 0 {
		t.Errorf("kept=%d omitted=%d, want both blocks kept and nothing recorded", len(kept), len(omitted))
	}
}

// TestOmitAnnotateRoundTripsOmissions pins that the omission markers survive a
// write and a parse. RISK: the whole point of recording them is that an operator
// reading the artifact -- or import, in its plan report -- can see what is not
// there; a marker that does not read back is a note nobody receives.
func TestOmitAnnotateRoundTripsOmissions(t *testing.T) {
	c := loadSample(t)
	c.Omitted = []Omission{
		{Section: "Configure SEMP Service", Reason: "the channel import runs over"},
		{Section: "Configure Redundancy", Reason: "would leave HA down"},
	}
	out := Annotate(c, "total", "20260911-000000")

	if !strings.Contains(string(out), `! solace-util-omitted: reason="the channel import runs over" section="Configure SEMP Service"`) {
		t.Errorf("omission marker missing or misrendered:\n%s", out)
	}
	back, err := ParseBlocks(out)
	if err != nil {
		t.Fatalf("ParseBlocks(Annotate): %v", err)
	}
	if len(back.Omitted) != 2 || back.Omitted[1].Section != "Configure Redundancy" ||
		back.Omitted[1].Reason != "would leave HA down" {
		t.Errorf("Omitted did not round-trip: %+v", back.Omitted)
	}
}

// TestImportPlanRefusesAFileItDidNotExport pins the provenance gate: import applies
// its input to a live broker, so it accepts only what export-config produced.
//
// RISK: a raw `show current-config` capture looks importable and is not -- its Skip
// sections were never removed and its CLI transcript was never cut, so importing one
// would apply `Configure SEMP Service` (severing the channel mid-run) and feed echoed
// prompts to the CLI as commands. Both invisibly, since cli -Apes exits 0.
//
// This is provenance, NOT tamper-proofing: the marker is a comment and anyone editing
// an artifact can keep it. It stops the wrong file, not a deliberate edit -- and could
// not do more while `broker perform cli-script` runs any script at all.
func TestImportPlanRefusesAFileItDidNotExport(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "currentconfig_sample.cli"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		return testBannerSoftware, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)

	cases := []struct {
		name     string
		artifact []byte
		wantSub  string
	}{
		{
			// The real thing an operator would reach for: the broker's own output,
			// saved straight to a file.
			name:     "raw show current-config capture",
			artifact: raw,
			wantSub:  "was not produced by `broker perform export-config`",
		},
		{
			name:     "hand-written CLI script",
			artifact: []byte(" ! BEGIN\nhome\nenable\nconfigure\nmessage-vpn \"x\"\n  exit\n! END\n"),
			wantSub:  "was not produced by `broker perform export-config`",
		},
		{
			// A newer build's artifact: the marker is there, the version is not one
			// this build agreed to.
			name:     "future export format",
			artifact: []byte("! solace-util-export: 99\n ! BEGIN\nhome\nenable\nconfigure\n! END\n"),
			wantSub:  "export format version",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := o.ImportPlan(context.Background(), config.Primary, c.artifact)
			if err == nil {
				t.Fatal("ImportPlan accepted a file it did not export")
			}
			if !strings.Contains(err.Error(), c.wantSub) {
				t.Errorf("err = %v, want it to mention %q", err, c.wantSub)
			}
			// The refusal must name the way forward, not just say no.
			if !strings.Contains(err.Error(), "export-config") && !strings.Contains(err.Error(), "re-export") {
				t.Errorf("err = %v, want it to name the command that produces an importable artifact", err)
			}
		})
	}
}

// TestImportPlanAcceptsWhatExportProduced is the other half: whatever export writes
// must pass its own provenance gate. RISK: a marker written one way and read another
// would refuse every real artifact, and only an end-to-end test catches that.
func TestImportPlanAcceptsWhatExportProduced(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "currentconfig_sample.cli"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		switch {
		case matchCLI(argv, "export-config"):
			return raw, nil
		// ImportPlan reads the TARGET's own full capture (readTarget) for its VPNs,
		// under a script name of its own so the two reads are distinguishable. The
		// banner leads that transcript because a real `cli -Apes` echoes it ahead of
		// any output, and ImportPlan now reads the TARGET's broker type off exactly
		// this response instead of paying a round trip of its own for it -- so a
		// fixture without one would read back as BrokerUnknown and be refused as a
		// cross-type import.
		case matchCLI(argv, "plan-vpns"):
			return append(append([]byte{}, testBannerSoftware...), raw...), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)

	art, err := o.ExportConfig(context.Background(), config.Primary, nil, false, "20260911-000000")
	if err != nil {
		t.Fatalf("ExportConfig: %v", err)
	}
	if _, err := o.ImportPlan(context.Background(), config.Primary, art); err != nil {
		t.Fatalf("ImportPlan refused this build's own export: %v", err)
	}
}

// TestOmitExportConfigEndToEnd runs the real export path over the sample fixture,
// whose broker region carries `Configure System` (a Skip section), and asserts the
// artifact neither contains it nor pretends it was never there.
func TestOmitExportConfigEndToEnd(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "currentconfig_sample.cli"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "export-config") {
			return raw, nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)

	art, err := o.ExportConfig(context.Background(), config.Primary, nil, false, "20260911-000000")
	if err != nil {
		t.Fatalf("ExportConfig: %v", err)
	}
	s := string(art)
	if strings.Contains(s, "! Configure System:") || strings.Contains(s, "system scaling") {
		t.Errorf("a Skip section is still in the exported artifact:\n%s", s)
	}
	if !strings.Contains(s, `! solace-util-omitted:`) || !strings.Contains(s, `section="Configure System"`) {
		t.Errorf("the artifact does not record that Configure System was omitted:\n%s", s)
	}
	// And nothing else was lost: the VPN and its objects are still there.
	if !strings.Contains(s, `create message-vpn "simple-vpn"`) || !strings.Contains(s, `client-username "app1"`) {
		t.Errorf("export dropped more than the Skip section:\n%s", s)
	}
}

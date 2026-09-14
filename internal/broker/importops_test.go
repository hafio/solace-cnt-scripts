package broker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"solace/internal/config"
	"solace/internal/output"
)

// importops_test.go covers internal/broker/importops.go: ExportConfig,
// ImportPlan, ImportApply, ImportVerify and their pure helpers. It reuses
// newTestOps/fakeTransport/matchCLI (broker_test.go) and loadSample's fixture
// (testdata/currentconfig_sample.cli, blocks_test.go) rather than inventing a
// second set of doubles.
//
// `cli -Apes` exits 0 even when the broker rejects a line, so there are two error
// detectors and the tests here cover both: the apply driver's per-chunk transcript
// check (driver.go) and the re-export diff (diff.go). Everything else here is aimed
// at what is checkable before either -- what gets uploaded, in what order, what the
// generated driver actually says, and what refuses before touching the broker.

// minimalCapture builds the smallest byte slice ParseBlocks accepts: a header
// naming schema/redaction and an empty body. It exists so tests that only care
// about ExportConfig's/ImportPlan's OWN logic (not about specific captured
// content) do not have to hand-roll a full `show current-config` transcript.
func minimalCapture(schema string, redacted string) []byte {
	// Two headers, both required and for different reasons. The export marker is what
	// ImportPlan's provenance check demands -- an artifact without it is refused as
	// not having come from export-config. The broker's own ` ! BEGIN` (leading space
	// and all) is where runCapture cuts the CLI transcript off.
	return []byte(fmt.Sprintf(
		"! solace-util-export: 1\n"+
			" ! BEGIN\n! SEMP Schema Version: %q\n! Redacted: %q\n\nhome\nenable\nconfigure\n\n! END\n",
		schema, redacted))
}

// --- ExportConfig ------------------------------------------------------------

// TestImportOpsExportConfigScopeSelectsCLICommand covers the scope->command
// mapping: no VPNs must run the single whole-broker capture, and --vpn must run
// one capture PER NAME rather than one filtered whole-broker capture. Getting
// this backwards would either miss VPNs a caller asked for by name or capture
// (and expose) far more of the broker than a scoped export is supposed to.
func TestImportOpsExportConfigScopeSelectsCLICommand(t *testing.T) {
	capture := minimalCapture("broker/10.26.0.8827", "False")

	// No VPNs: exactly the whole-broker capture, and the uploaded script is
	// removed afterwards -- it is not credential-bearing here, but the cleanup
	// contract must hold regardless of scope.
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "export-config") {
			return capture, nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if _, err := o.ExportConfig(context.Background(), config.Primary, nil, false, "2026-01-01"); err != nil {
		t.Fatalf("ExportConfig (no vpns): %v", err)
	}
	if body := ft.uploadBody(t, cliScriptPath("export-config")); body != currentConfigScript("", false) {
		t.Errorf("ExportConfig with no VPNs uploaded %q, want the whole-broker capture %q",
			body, currentConfigScript("", false))
	}
	if !ft.removed(cliScriptPath("export-config")) {
		t.Error("ExportConfig must remove the uploaded capture script after running it")
	}

	// --vpn acme,beta: one invocation per name, each scoped to that VPN alone,
	// in the order given.
	ft2 := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "export-config") {
			return capture, nil
		}
		return nil, nil
	}}
	o2, _ := newTestOps(t, &config.Config{}, ft2)
	vpns := []string{"acme", "beta"}
	if _, err := o2.ExportConfig(context.Background(), config.Primary, vpns, false, "2026-01-01"); err != nil {
		t.Fatalf("ExportConfig (vpns): %v", err)
	}
	var got []string
	for _, u := range ft2.uploads {
		if u.dest == cliScriptPath("export-config") {
			got = append(got, u.data)
		}
	}
	if len(got) != len(vpns) {
		t.Fatalf("ExportConfig uploaded %d capture script(s), want exactly one per VPN name (%d)",
			len(got), len(vpns))
	}
	for i, vpn := range vpns {
		if want := currentConfigScript(vpn, false); got[i] != want {
			t.Errorf("upload[%d] (VPN %q) = %q, want %q", i, vpn, got[i], want)
		}
	}
	if !ft2.removed(cliScriptPath("export-config")) {
		t.Error("ExportConfig must remove each uploaded per-VPN capture script")
	}
}

// TestImportOpsExportConfigRejectsUnsafeVPNNameBeforeTransport covers the
// injection boundary: a VPN name is written into a quoted CLI operand
// (`show current-config message-vpn "<name>"`), so a double quote would close
// that operand early and a control character would split the line into extra
// commands run in an already-open CLI session. Both must be refused before any
// capture is attempted, including when an earlier name in the list is valid --
// otherwise the valid VPNs would already have been captured (and their scripts
// uploaded to the broker) by the time the bad one is noticed.
func TestImportOpsExportConfigRejectsUnsafeVPNNameBeforeTransport(t *testing.T) {
	for _, bad := range []string{`evil"vpn`, "evil\nvpn", "evil\x01vpn"} {
		ft := &fakeTransport{}
		o, _ := newTestOps(t, &config.Config{}, ft)
		if _, err := o.ExportConfig(context.Background(), config.Primary, []string{"good", bad}, false, "x"); err == nil {
			t.Errorf("ExportConfig should reject VPN name %q", bad)
		}
		if len(ft.uploads) != 0 || len(ft.outputs) != 0 {
			t.Errorf("ExportConfig must validate every VPN name before any transport call (bad=%q), got uploads=%v outputs=%v",
				bad, ft.uploads, ft.outputs)
		}
	}
}

// TestImportOpsExportConfigBrokerOnlyDropsVPNRegions covers --broker-only: it
// must drop both the create-VPN pass (RegionVPNCreate) and every VPN's own
// objects (RegionVPN). Missing the create-VPN pass is the dangerous half of
// this: a "broker-only" artifact that still carried it would silently create
// every message-VPN the source broker had on whatever target it was imported
// into.
func TestImportOpsExportConfigBrokerOnlyDropsVPNRegions(t *testing.T) {
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
	out, err := o.ExportConfig(context.Background(), config.Primary, nil, true, "2026-01-01")
	if err != nil {
		t.Fatalf("ExportConfig (broker-only): %v", err)
	}
	if strings.Contains(string(out), "message-vpn") {
		t.Errorf("a --broker-only export must carry no message-vpn block (create-VPN pass or "+
			"VPN-scoped object), got:\n%s", out)
	}
}

// TestImportOpsExportConfigRefusesEmptyCapture covers the empty-capture guard:
// an empty result must be a hard failure, not an empty-but-valid artifact. An
// empty artifact would still parse and import as "delete nothing, apply
// nothing" and report success, silently hiding whatever actually went wrong
// with the capture.
func TestImportOpsExportConfigRefusesEmptyCapture(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "export-config") {
			return []byte("   \n"), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if _, err := o.ExportConfig(context.Background(), config.Primary, nil, false, "x"); err == nil {
		t.Error("ExportConfig should fail on an empty capture rather than return an empty artifact")
	}
}

// --- ImportPlan ----------------------------------------------------------------

// TestImportOpsImportPlanRefusesRedactedArtifact covers the redaction guard: a
// redacted artifact has had exactly the credential material an import needs to
// put back stripped out of it, so importing it would apply a broken
// configuration rather than failing loudly up front.
func TestImportOpsImportPlanRefusesRedactedArtifact(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	_, err := o.ImportPlan(context.Background(), config.Primary, minimalCapture("broker/10.26.0.8827", "True"))
	if err == nil {
		t.Fatal("ImportPlan should refuse a redacted artifact")
	}
	if !strings.Contains(err.Error(), "redact") {
		t.Errorf("error %q should name redaction as the reason this artifact cannot be imported", err)
	}
	if len(ft.outputs) != 0 {
		t.Error("ImportPlan must refuse a redacted artifact before it ever reads the target")
	}
}

// TestImportOpsImportPlanRefusesCrossTypeAndUnknownType covers checkSameType's
// whole truth table: a software/appliance mismatch in either direction, and
// either side coming back unknown. RISK: an appliance capture carries
// interface, VRF, DNS, clock and SNMP sections a software broker has none of
// (and the reverse), so letting a cross-type artifact through would apply
// configuration the target cannot make sense of.
func TestImportOpsImportPlanRefusesCrossTypeAndUnknownType(t *testing.T) {
	// The target's type is read off the CLI login banner (bannerType) -- but now
	// off the SAME "plan-vpns" round trip readTarget makes for its VPN listing,
	// not a dedicated "banner" call (the removed Ops.brokerType). So the canned
	// target response is the real banner PREPENDED to a valid target capture, the
	// way a real `cli -Apes` transcript always carries its banner ahead of the
	// script's own output.
	unrecognisable := []byte("no banner in this transcript\n")

	cases := []struct {
		name     string
		artifact []byte
		verOut   []byte
		wantSubs []string
	}{
		{
			name:     "software artifact, appliance target",
			artifact: minimalCapture("broker/10.26.0.8827", "False"),
			verOut:   testBannerAppliance,
			wantSubs: []string{"software", "appliance"},
		},
		{
			name:     "appliance artifact, software target",
			artifact: minimalCapture("app/10.25.4.1", "False"),
			verOut:   testBannerSoftware,
			wantSubs: []string{"appliance", "software"},
		},
		{
			name:     "artifact type unknown",
			artifact: minimalCapture("garbage-schema", "False"),
			verOut:   testBannerSoftware,
			wantSubs: []string{"cannot tell which broker type this artifact"},
		},
		{
			name:     "target type unknown",
			artifact: minimalCapture("broker/10.26.0.8827", "False"),
			verOut:   unrecognisable,
			wantSubs: []string{"cannot tell which broker type the target"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			target := append(append([]byte{}, tc.verOut...), targetVPNCapture()...)
			ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
				if matchCLI(argv, "plan-vpns") {
					return target, nil
				}
				return nil, nil
			}}
			o, _ := newTestOps(t, &config.Config{}, ft)
			_, err := o.ImportPlan(context.Background(), config.Primary, tc.artifact)
			if err == nil {
				t.Fatalf("%s: ImportPlan should refuse", tc.name)
			}
			for _, want := range tc.wantSubs {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("%s: error %q should contain %q", tc.name, err, want)
				}
			}
		})
	}
}

// targetVPNCapture builds the minimal `show current-config all` a fake target can
// answer the plan's VPN listing with: the create-all-VPNs pass names every VPN the
// broker has, and that is all readTarget needs for the VPN split.
func targetVPNCapture(vpns ...string) []byte {
	lines := []string{
		"! BEGIN",
		`! Command: "show current-config all"`,
		"",
		"home",
		"enable",
		"configure",
		"",
		`! Create All: "Message Vpns"`,
	}
	for _, v := range vpns {
		lines = append(lines, `create message-vpn "`+v+`"`, "  exit")
	}
	return []byte(strings.Join(append(lines, "", "! END", ""), "\n"))
}

// TestImportOpsImportPlanSplitsExistingAndNewVPNs covers the Existing/NewVPNs
// split against the target's own configuration.
//
// RISK, and it is the worst one this feature has: a VPN wrongly classed as "new"
// is never torn down, so the artifact's `create` blocks are applied on top of a
// live VPN. A Solace CLI block cannot be applied twice, so every line is rejected
// -- invisibly, because `cli -Apes` exits 0 -- and ImportVerify then PASSES,
// because the untouched VPN still carries every line the artifact asked for. A
// misclassification here does not miss a step, it defeats the only error detection
// that exists while the other VPNs' spooled messages are destroyed.
//
// The second subtest is the regression for how that actually happened. The split
// used to read `show message-vpn *` through a parser that split the name column on
// whitespace, so this broker's own `A VPN WITH LONG NAME AND SPACES` came back as
// seven fragments and matched nothing -- an existing VPN, filed as new. readTarget
// now reads a full capture through the same quote-aware parser that reads the
// artifact, so the two sides cannot disagree about what a VPN is called.
func TestImportOpsImportPlanSplitsExistingAndNewVPNs(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "currentconfig_sample.cli"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	const spaced = "A VPN WITH LONG NAME AND SPACES"

	// The fixture's own VPNs, in first-seen order, are spaced and "simple-vpn"
	// (see blocks_test.go).
	cases := []struct {
		name         string
		targetHas    []string
		wantExisting []string
		wantNew      []string
	}{
		{
			name:         "only the plain VPN is on the target",
			targetHas:    []string{"simple-vpn"},
			wantExisting: []string{"simple-vpn"},
			wantNew:      []string{spaced},
		},
		{
			// A name with spaces must be recognised as EXISTING, or it is applied
			// on top of itself and the import silently does nothing while claiming
			// success.
			name:         "a VPN whose name contains spaces is matched, not shredded",
			targetHas:    []string{spaced},
			wantExisting: []string{spaced},
			wantNew:      []string{"simple-vpn"},
		},
		{
			name:         "both are on the target, so both are torn down",
			targetHas:    []string{spaced, "simple-vpn"},
			wantExisting: []string{spaced, "simple-vpn"},
			wantNew:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// The banner precedes the plan-vpns capture in the same round trip now
			// (readTarget reads the broker type off it, per runCapture) -- there is
			// no more separate dedicated "banner" call.
			target := append(append([]byte{}, testBannerSoftware...), targetVPNCapture(tc.targetHas...)...)
			ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
				switch {
				case matchCLI(argv, "plan-vpns"):
					return target, nil
				case matchCLI(argv, "export-config"):
					return raw, nil
				}
				return nil, nil
			}}
			o, _ := newTestOps(t, &config.Config{}, ft)

			// Export first: ImportPlan accepts only artifacts this tool produced, so
			// the raw fixture has to go through the real export path rather than be
			// handed over directly. Doing it this way also means the input cannot
			// drift from what export actually writes.
			art, err := o.ExportConfig(context.Background(), config.Primary, nil, false, "20260911-000000")
			if err != nil {
				t.Fatalf("ExportConfig: %v", err)
			}
			plan, err := o.ImportPlan(context.Background(), config.Primary, art)
			if err != nil {
				t.Fatalf("ImportPlan: %v", err)
			}
			if !reflect.DeepEqual(plan.Existing, tc.wantExisting) {
				t.Errorf("Existing = %v, want %v", plan.Existing, tc.wantExisting)
			}
			if !reflect.DeepEqual(plan.NewVPNs, tc.wantNew) {
				t.Errorf("NewVPNs = %v, want %v", plan.NewVPNs, tc.wantNew)
			}
		})
	}
}

// TestImportOpsListVPNsFailsLoudly covers both ways the target's VPN listing can
// fail, and it matters more than a normal error-path test.
//
// "I could not tell what VPNs the target has" must never reach the caller as an
// EMPTY list, because empty means "no VPN needs tearing down" -- which silently
// downgrades the import from teardown-and-rebuild to apply-on-top, where a Solace
// CLI block lands on a live VPN, is rejected invisibly (`cli -Apes` exits 0), and
// ImportVerify then passes because the untouched VPN still satisfies the artifact.
// Both failures are therefore errors, each naming its own step.
func TestImportOpsReadTargetFailsLoudly(t *testing.T) {
	cases := []struct {
		name    string
		reply   []byte
		wantErr string
	}{
		{"capture came back empty", []byte("   \n"), "read the target's configuration"},
		{"capture does not parse", []byte(" ! BEGIN\nno terminator here\n"), "parse the target's configuration"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
				if matchCLI(argv, "plan-vpns") {
					return tc.reply, nil
				}
				return nil, nil
			}}
			o, _ := newTestOps(t, &config.Config{}, ft)
			got, typ, err := o.readTarget(context.Background(), config.Primary)
			if err == nil {
				t.Fatalf("readTarget = %v, nil -- an unreadable target must be an error, never an empty capture", got)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("err = %v, want it to name the %q step", err, tc.wantErr)
			}
			if got != nil {
				t.Errorf("readTarget returned %v alongside its error; the caller must get nothing to act on", got)
			}
			if typ != BrokerUnknown {
				t.Errorf("readTarget returned type %q alongside its error, want BrokerUnknown", typ)
			}
		})
	}
}

// TestImportPlanReportPrintsWarningsBeforeTheConfirmation pins WHERE the warnings
// are printed, which is the whole point of them.
//
// They used to be rendered only by ImportResult, after ImportApply had already
// destroyed every existing VPN. An unclassified section this build will not apply,
// or a service InjectShutdown cannot safely bounce, is exactly the sort of thing an
// operator might refuse -- and they were being told about it once refusing was no
// longer possible. The plan report is what they read before the exact-"yes".
func TestImportPlanReportPrintsWarningsBeforeTheConfirmation(t *testing.T) {
	var buf bytes.Buffer
	p := &ImportPlan{
		Capture:  &Capture{Type: BrokerSoftware, Version: "10.26.0.8827"},
		Warnings: []string{`section "Configure Routing DMR" is not classified`},
	}
	p.Report(output.New(&buf))
	if got := buf.String(); !strings.Contains(got, "Configure Routing DMR") {
		t.Errorf("the plan report does not carry the warning the operator has to see before confirming:\n%s", got)
	}
}

// TestImportIgnoreKeepsUnclassifiedSectionsInTheDiff separates a decision from a
// gap, which importIgnore used to conflate.
//
// A section the table says to skip is a decision, so its lines are excluded from
// verification -- otherwise every deliberately-skipped section would read as a
// failure. An UNCLASSIFIED section is dispositioned Skip too, but it is a gap: this
// build does not know the section, so it does not apply it. Excluding those lines
// as well meant a broker carrying configuration this build has never seen got it
// silently dropped and was then reported as a clean import. They now stay in the
// diff, so the command exits non-zero and names them.
func TestImportIgnoreKeepsUnclassifiedSectionsInTheDiff(t *testing.T) {
	p := &ImportPlan{Sections: []PlannedSection{
		{Name: "Configure Redundancy", Disposition: Skip, Reason: "ends shut down"},
		{Name: "Configure Routing DMR", Disposition: Skip, Unclassified: true,
			Reason: "not classified by this build"},
	}}
	ignore := importIgnore(p)

	if !ignore(Block{Section: "Configure Redundancy"}, "redundancy shutdown") {
		t.Error("a rule-based skip must be ignored by the diff: it was never meant to land")
	}
	if ignore(Block{Section: "Configure Routing DMR"}, `create dmr-cluster "mesh"`) {
		t.Error("an UNCLASSIFIED section was excused from the diff, so configuration this build does not know is dropped and the import still reports success")
	}
}

// TestImportResultReportsTeardownAndRebuildSeparately pins the report on the path
// where it is the ONLY statement of what the broker is holding: runImport prints
// the result before returning an apply error, and ImportVerify does not run.
//
// The row used to read "VPNs torn down and rebuilt" straight off TornDown, which
// asserted the rebuild had happened in exactly the case where it had not -- the
// teardown loop completes before any rebuild begins, so an apply failure leaves
// those VPNs emptied, their spooled messages destroyed, and nothing put back. That
// is the single most important fact about the broker's state and the one line
// describing it said the opposite.
func TestImportResultReportsTeardownAndRebuildSeparately(t *testing.T) {
	var buf bytes.Buffer
	res := &ImportResult{TornDown: []string{"acme"}} // Rebuilt stays false: apply failed
	res.Report(output.New(&buf))
	got := buf.String()

	if strings.Contains(got, "torn down and rebuilt") {
		t.Errorf("the report still claims the VPN was rebuilt:\n%s", got)
	}
	if !strings.Contains(got, "NOT rebuilt") {
		t.Errorf("the report does not say the VPN was left empty:\n%s", got)
	}
	if !strings.Contains(got, "Re-run the same import") {
		t.Errorf("the report does not say how to recover:\n%s", got)
	}

	buf.Reset()
	res.Rebuilt = true
	res.Report(output.New(&buf))
	if got := buf.String(); strings.Contains(got, "NOT rebuilt") {
		t.Errorf("a successful apply must not carry the data-loss warning:\n%s", got)
	}
}

// --- planSections --------------------------------------------------------------

// TestImportOpsPlanSectionsUnclassifiedSectionWarns covers the unclassified-gap
// path: a section this build's table has never seen must Skip AND warn by
// name, never default to Apply. Silently applying an unclassified section is
// exactly the hazard sections.go's whole design exists to prevent, so an
// unclassified section that applied with no warning would defeat the point of
// having the table at all.
func TestImportOpsPlanSectionsUnclassifiedSectionWarns(t *testing.T) {
	blocks := []Block{{
		Section: "Configure Something This Build Has Never Seen",
		Region:  RegionBroker,
		Kind:    "widget",
		Lines:   []string{"widget frobnicate"},
	}}
	sections, warnings := planSections(blocks, BrokerSoftware)
	if len(sections) != 1 || sections[0].Disposition != Skip {
		t.Fatalf("planSections = %+v, want a single Skip section", sections)
	}
	found := false
	for _, w := range warnings {
		if strings.Contains(w, "Configure Something This Build Has Never Seen") {
			found = true
		}
	}
	if !found {
		t.Errorf("planSections should warn naming the unclassified section, got %v", warnings)
	}
}

// TestImportOpsPlanSectionsSkipApplianceBecomesApplyOnApplianceTarget covers
// the one disposition that depends on the TARGET rather than being fixed at
// build time: a section classified skip-appliance is genuinely appliance
// config, so it must actually apply when the target IS an appliance -- a
// software-only skip here would silently drop legitimate configuration on the
// one target it was written for.
func TestImportOpsPlanSectionsSkipApplianceBecomesApplyOnApplianceTarget(t *testing.T) {
	blocks := []Block{{Section: "Configure DNS", Region: RegionBroker, Kind: "dns", Lines: []string{"dns something"}}}

	software, _ := planSections(blocks, BrokerSoftware)
	if len(software) != 1 || software[0].Disposition != SkipAppliance {
		t.Fatalf("Configure DNS on a software target = %+v, want SkipAppliance", software)
	}

	appliance, _ := planSections(blocks, BrokerAppliance)
	if len(appliance) != 1 || appliance[0].Disposition != Apply {
		t.Errorf("Configure DNS on an appliance target = %+v, want Apply", appliance)
	}
}

// --- ImportApply -----------------------------------------------------------

// newTeardownApplyFixture builds an ImportPlan with one Existing VPN and two
// broker-level sections classified Apply -- "Create console" (ordinary) and
// "Create logging" (First) -- so the teardown-ordering and first-section-
// isolation tests below can each assert their own property against one
// realistic shape without duplicating it.
// importDriverDest is where runDriver uploads the generated apply script.
var importDriverDest = shellScriptPath("solace-util-import")

// driverChunkRE matches the header renderDriver writes above each chunk.
var driverChunkRE = regexp.MustCompile(`(?m)^# ---- chunk ([0-9]+): (.*) ----$`)

// driverChunkNames lists the chunk names a rendered driver carries, in order.
//
// Tests read the chunking out of the DRIVER TEXT rather than out of a list of
// uploads, because that text is now the whole apply: chunk bodies live inside it as
// heredocs and never become uploads of their own. Asserting on the artifact that
// actually runs is also what keeps these tests honest -- there is no second
// representation for them to agree with while production disagrees.
func driverChunkNames(driver string) []string {
	var out []string
	for _, m := range driverChunkRE.FindAllStringSubmatch(driver, -1) {
		out = append(out, m[2])
	}
	return out
}

// driverChunkBody returns the heredoc body of the named chunk.
func driverChunkBody(t *testing.T, driver, name string) string {
	t.Helper()
	for _, m := range driverChunkRE.FindAllStringSubmatchIndex(driver, -1) {
		if driver[m[4]:m[5]] != name {
			continue
		}
		rest := driver[m[1]:]
		open := "<<'" + heredocDelim + "'\n"
		i := strings.Index(rest, open)
		if i < 0 {
			t.Fatalf("chunk %q has no heredoc", name)
		}
		rest = rest[i+len(open):]
		j := strings.Index(rest, "\n"+heredocDelim+"\n")
		if j < 0 {
			t.Fatalf("chunk %q heredoc is unterminated", name)
		}
		return rest[:j]
	}
	t.Fatalf("no chunk named %q in driver; have %v", name, driverChunkNames(driver))
	return ""
}

// driverAllOK is the marker output a driver prints when every chunk applied.
func driverAllOK(driver string) []byte {
	var b strings.Builder
	for i := range driverChunkNames(driver) {
		fmt.Fprintf(&b, "%s %d\n", markerChunkOK, i)
	}
	b.WriteString(markerDone + "\n")
	return []byte(b.String())
}

// driverFailAt is the marker output for a driver that hit a rejection at chunk idx:
// OK up to it, the echoed detail, the FAIL marker, and NO completion marker --
// exactly what the generated script does, since it exits there.
func driverFailAt(driver string, idx int, detail string) []byte {
	var b strings.Builder
	for i := range driverChunkNames(driver) {
		if i == idx {
			fmt.Fprintf(&b, "%s\n%s\n%s\n", markerDetailOpen, detail, markerDetailShut)
			fmt.Fprintf(&b, "%s %d\n", markerChunkFail, i)
			break
		}
		fmt.Fprintf(&b, "%s %d\n", markerChunkOK, i)
	}
	return []byte(b.String())
}

func newTeardownApplyFixture(t *testing.T) (*Ops, *fakeTransport, *ImportPlan) {
	t.Helper()
	blocks := []Block{
		{Section: "Create console", Region: RegionBroker, Kind: "console",
			Lines: []string{"create console", "  idle-timeout 100", "  exit"}},
		{Section: "Create logging", Region: RegionBroker, Kind: "logging",
			Lines: []string{"logging debug enable"}},
	}
	sections, _ := planSections(blocks, BrokerSoftware)
	plan := &ImportPlan{
		Capture: &Capture{
			Preamble: []string{"home", "enable", "configure"},
			Blocks:   blocks,
			Type:     BrokerSoftware,
		},
		Existing: []string{"acme"},
		Sections: sections,
	}
	ft := &fakeTransport{}
	ft.responder = func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "export-teardown") {
			// The teardown CAPTURE the target hands back: the broker's own
			// generated removal commands for this VPN, opening with its ` ! BEGIN`
			// header the way every real capture does (runCapture cuts at it).
			return []byte(" ! BEGIN\n" + `message-vpn "acme" remove` + "\n"), nil
		}
		// The one `sh <driver>` exec the apply makes. The marker output is derived
		// from the uploaded driver rather than assumed, because parseDriverOutput
		// refuses a marker naming a chunk that does not exist -- a double that
		// guessed the count would decide these tests on the guess.
		if len(argv) > 0 && argv[0] == "sh" {
			return driverAllOK(ft.uploadBody(t, importDriverDest)), nil
		}
		return []byte("ok\n"), nil
	}
	o, _ := newTestOps(t, &config.Config{}, ft)
	return o, ft, plan
}

// TestImportOpsImportApplyTearsDownExistingVPNBeforeApplyingArtifact covers the
// teardown ordering: for each Existing VPN, ImportApply must (1) ask the
// TARGET to generate its own removal commands via
// `show current-config message-vpn "<name>" remove` and (2) apply that
// teardown before any of the artifact's own blocks. RISK: applying the
// artifact's create-VPN pass onto a VPN that still exists cannot work (a CLI
// block cannot be reapplied) -- reversing this order is a broken import, not
// just a slower one.
func TestImportOpsImportApplyTearsDownExistingVPNBeforeApplyingArtifact(t *testing.T) {
	o, ft, plan := newTeardownApplyFixture(t)
	if _, err := o.ImportApply(context.Background(), config.Primary, plan); err != nil {
		t.Fatalf("ImportApply: %v", err)
	}

	wantTeardownScript := currentConfigScript("acme", true)
	if body := ft.uploadBody(t, cliScriptPath("export-teardown")); body != wantTeardownScript {
		t.Errorf("export-teardown script = %q, want %q (the CLI command that asks the target "+
			"to generate its own removal commands)", body, wantTeardownScript)
	}
	if !strings.Contains(wantTeardownScript, `show current-config message-vpn "acme" remove`) {
		t.Fatalf("sanity: currentConfigScript(remove=true) = %q, missing the remove form", wantTeardownScript)
	}

	// The whole apply is now ONE upload and ONE exec: the teardown CAPTURE is a read
	// the broker has to answer before its content can be known, and everything after
	// it lives inside the generated driver as chunks.
	var destOrder []string
	for _, u := range ft.uploads {
		destOrder = append(destOrder, u.dest)
	}
	wantOrder := []string{cliScriptPath("export-teardown"), importDriverDest}
	if !reflect.DeepEqual(destOrder, wantOrder) {
		t.Fatalf("upload order = %v, want %v (the teardown capture, then the one apply driver)",
			destOrder, wantOrder)
	}
	var execs int
	for _, o := range ft.outputs {
		if len(o.argv) > 0 && o.argv[0] == "sh" {
			execs++
		}
	}
	if execs != 1 {
		t.Errorf("the apply made %d `sh` execs, want exactly 1 -- pushing the per-chunk loop "+
			"into the broker is the whole reason the driver exists", execs)
	}

	// Chunk order inside the driver: the teardown before anything is rebuilt, then
	// the artifact's own sections.
	driver := ft.uploadBody(t, importDriverDest)
	names := driverChunkNames(driver)
	if len(names) == 0 || names[0] != "teardown acme" {
		t.Fatalf("chunk order = %v, want the teardown first -- rebuilding onto a VPN that still "+
			"exists cannot work, and the diff cannot see that it did not", names)
	}
}

// TestImportOpsImportApplySeparatesFirstSectionFromMain covers the isolation
// contract for a First section ("Create logging" ends the CLI session, per
// sections.go): it must run in its OWN RunCLI invocation, never sharing one
// with the rest. RISK: sharing an invocation means whatever follows the
// logging change in that same script never runs -- `cli -Apes` is one process,
// and this session is gone once logging is reconfigured.
func TestImportOpsImportApplySeparatesFirstSectionFromMain(t *testing.T) {
	o, ft, plan := newTeardownApplyFixture(t)
	res, err := o.ImportApply(context.Background(), config.Primary, plan)
	if err != nil {
		t.Fatalf("ImportApply: %v", err)
	}

	driver := ft.uploadBody(t, importDriverDest)
	first := driverChunkBody(t, driver, chunkSectionPrefix+"logging")
	main := driverChunkBody(t, driver, chunkSectionPrefix+"broker")

	if !strings.Contains(first, "logging debug enable") {
		t.Errorf("the logging chunk should carry the logging section, got %q", first)
	}
	if strings.Contains(first, "create console") {
		t.Errorf("the logging chunk must carry ONLY the First section, got %q", first)
	}
	if !strings.Contains(main, "create console") {
		t.Errorf("the broker chunk should carry the console section, got %q", main)
	}
	if strings.Contains(main, "logging debug enable") {
		t.Errorf("the logging line must not appear in the broker chunk -- sharing an invocation "+
			"with it means nothing after it would run, got %q", main)
	}

	if len(res.Applied) != 2 || res.Applied[0] != "Create logging" || res.Applied[1] != "Create console" {
		t.Errorf("Applied = %v, want [Create logging Create console] (the First section reported "+
			"before the rest, matching the invocation order)", res.Applied)
	}
	if len(res.TornDown) != 1 || res.TornDown[0] != "acme" {
		t.Errorf("TornDown = %v, want [acme]", res.TornDown)
	}
}

// TestBuildChunksCreatesVPNsBeforeBrokerSections is the regression for an ordering
// regression the chunking introduced.
//
// In the artifact, `! Create All: "Message Vpns"` is the FIRST thing after
// `configure`, ahead of the whole broker-level region -- and that is load bearing,
// because a broker-level line naming a message-VPN is rejected if the VPN does not
// exist yet. Chunking originally folded those blocks into each VPN's own chunk,
// which runs LAST, so every such reference applied before its VPN was created. The
// single-script apply this replaced had it right only by rendering in artifact order.
func TestBuildChunksCreatesVPNsBeforeBrokerSections(t *testing.T) {
	blocks := []Block{
		{Section: `Create All: "Message Vpns"`, Region: RegionVPNCreate, Kind: "message-vpn",
			Name: "acme", VPN: "acme", Create: true,
			Lines: []string{`create message-vpn "acme"`, "  exit"}},
		{Section: "Create console", Region: RegionBroker, Kind: "console",
			Lines: []string{"create console"}},
		{Section: `Create Client Username: "acme"`, Region: RegionVPN, Kind: "client-username",
			Name: "app1", VPN: "acme",
			Lines: []string{`client-username "app1" message-vpn "acme"`, "  exit"}},
	}
	sections, _ := planSections(blocks, BrokerSoftware)
	p := &ImportPlan{
		Capture:  &Capture{Preamble: []string{"home", "enable", "configure"}, Blocks: blocks, Type: BrokerSoftware},
		Sections: sections,
	}
	o, _ := newTestOps(t, &config.Config{}, &fakeTransport{})

	chunks, err := o.buildChunks(context.Background(), config.Primary, p, blocks, &ImportResult{})
	if err != nil {
		t.Fatalf("buildChunks: %v", err)
	}

	idx := func(want string) int {
		for i, c := range chunks {
			if strings.Contains(c.Name, want) {
				return i
			}
		}
		return -1
	}
	create, broker, vpn := idx("create message-vpns"), idx(chunkSectionPrefix+"broker"), idx("message-vpn acme")
	if create < 0 {
		t.Fatalf("no create-VPN chunk; chunks = %v", chunkNames(chunks))
	}
	if broker >= 0 && create > broker {
		t.Errorf("create-VPN chunk is at %d, after the broker chunk at %d -- a broker-level line "+
			"naming a VPN would apply before that VPN exists; chunks = %v", create, broker, chunkNames(chunks))
	}
	if vpn >= 0 && create > vpn {
		t.Errorf("create-VPN chunk is at %d, after the per-VPN chunk at %d", create, vpn)
	}
	// And the create pass must NOT be duplicated into the VPN's own chunk.
	if vpn >= 0 && strings.Contains(chunks[vpn].Body, `create message-vpn "acme"`) {
		t.Errorf("the per-VPN chunk re-creates the VPN:\n%s", chunks[vpn].Body)
	}
}

// chunkNames lists chunk names for a failure message.
func chunkNames(chunks []chunk) []string {
	out := make([]string, len(chunks))
	for i, c := range chunks {
		out[i] = c.Name
	}
	return out
}

// TestTornDownRecordsOnlyAppliedTeardowns pins what the most consequential line in
// the failure report means.
//
// res.TornDown used to be filled when a teardown was CAPTURED from the target, which
// happens before the driver has been rendered, uploaded or run. So every failure
// before or during the driver told the operator their VPN was emptied and its
// spooled messages destroyed when nothing had been applied to it -- the opposite of
// the truth, in the one report they read to decide what to do next.
func TestTornDownRecordsOnlyAppliedTeardowns(t *testing.T) {
	o, ft, plan := newTeardownApplyFixture(t)

	// The driver reports the teardown chunk as FAILED, so nothing was torn down.
	ft.responder = func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "export-teardown") {
			return []byte(" ! BEGIN\n" + `message-vpn "acme" remove` + "\n"), nil
		}
		if len(argv) > 0 && argv[0] == "sh" {
			return driverFailAt(ft.uploadBody(t, importDriverDest), 0, "error: rejected"), nil
		}
		return []byte("ok\n"), nil
	}

	res, err := o.ImportApply(context.Background(), config.Primary, plan)
	if err == nil {
		t.Fatal("a failed teardown chunk must fail the apply")
	}
	if len(res.TornDown) != 0 {
		t.Errorf("TornDown = %v, want empty -- the teardown chunk FAILED, so nothing was destroyed, "+
			"and telling the operator otherwise sends them to recover data that is still there",
			res.TornDown)
	}
	if res.Rebuilt {
		t.Error("Rebuilt must be false after a failed chunk")
	}
}

// TestImportOpsImportApplyTransportFailureIsNotAConfigRejection covers the
// distinction runDriver draws: a transport failure means the apply driver never
// RAN, which is a different failure from the broker rejecting a configuration line.
// RISK: conflating the two would tell an operator their configuration was refused
// when in fact nothing was even attempted -- the wrong next step follows from each.
// It matters more under the driver than it did before, because a rejection now
// arrives as the driver's OWN marker: an exec that produced no markers at all must
// not be described in the language of a rejection.
func TestImportOpsImportApplyTransportFailureIsNotAConfigRejection(t *testing.T) {
	blocks := []Block{{Section: "Create console", Region: RegionBroker, Kind: "console",
		Lines: []string{"create console"}}}
	sections, _ := planSections(blocks, BrokerSoftware)
	plan := &ImportPlan{
		Capture: &Capture{Preamble: []string{"home", "enable", "configure"}, Blocks: blocks, Type: BrokerSoftware},
		Sections: sections,
	}
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		// The apply is now ONE `sh <driver>` exec, not a cli invocation per script.
		if len(argv) > 0 && argv[0] == "sh" {
			return nil, errors.New("exec failed")
		}
		return []byte("ok\n"), nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if _, err := o.ImportApply(context.Background(), config.Primary, plan); err == nil {
		t.Fatal("ImportApply should surface the transport failure")
	} else if !strings.Contains(err.Error(), "did not run") ||
		!strings.Contains(err.Error(), "not a rejected configuration line") {
		t.Errorf("error %q should say the CLI did not run rather than that configuration was rejected", err)
	}
}

// --- ImportVerify ------------------------------------------------------------

// TestImportOpsImportVerifyReExportsAndDiffs covers ImportVerify's whole
// contract: it is the ONLY error detection this feature has (`cli -Apes` exits
// 0 even when the broker rejects a line, per this file's own doc comment), so
// it must re-export the target with NO scope filter (teardown already ran
// per-VPN; verification has to see the whole broker back), strip this tool's
// own markers before comparing, and route the comparison through importIgnore
// so a section the plan deliberately skipped does not read as a failure.
func TestImportOpsImportVerifyReExportsAndDiffs(t *testing.T) {
	applied := Block{Section: "Create console", Region: RegionBroker, Kind: "console", Lines: []string{"create console"}}
	skipped := Block{Section: "Configure System", Region: RegionBroker, Kind: "system",
		Lines: []string{"system scaling  scale  max-connections 1000"}}
	plan := &ImportPlan{
		Capture: &Capture{Blocks: []Block{applied, skipped}},
		Sections: []PlannedSection{
			{Name: "Create console", Disposition: Apply},
			{Name: "Configure System", Disposition: Skip},
		},
	}

	// A re-export carrying this tool's own marker line, which must be stripped
	// before the diff -- the fresh capture and the artifact do not both carry
	// markers at the moment of comparison (see StripMarkers' own doc comment).
	const reExport = "! solace-util-export: 1\n ! BEGIN\n" +
		"! SEMP Schema Version: \"broker/10.26.0.8827\"\n! Redacted: \"False\"\n\n" +
		"home\nenable\nconfigure\n\n! Create console:\ncreate console\n\n! END\n"

	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "verify-export") {
			return []byte(reExport), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)

	got, err := o.ImportVerify(context.Background(), config.Primary, plan)
	if err != nil {
		t.Fatalf("ImportVerify: %v", err)
	}
	if !got.OK() {
		t.Errorf("ImportVerify = %+v, want OK: the applied section matched and the skipped one "+
			"is excused by importIgnore", got)
	}
	if body := ft.uploadBody(t, cliScriptPath("verify-export")); body != currentConfigScript("", false) {
		t.Errorf("ImportVerify capture script = %q, want the unscoped whole-broker capture %q",
			body, currentConfigScript("", false))
	}
}

// TestImportOpsImportVerifyWrapsFailureContext covers the two ways ImportVerify
// can fail before DiffBlocks ever runs, each wrapped naming which STEP failed:
// an operator has to be able to tell "the broker capture failed" from "the
// re-export did not parse" -- the fix for each is completely different.
func TestImportOpsImportVerifyWrapsFailureContext(t *testing.T) {
	plan := &ImportPlan{Capture: &Capture{}}

	t.Run("capture fails", func(t *testing.T) {
		ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
			if matchCLI(argv, "verify-export") {
				return []byte("   \n"), nil // empty capture -> runCapture itself refuses
			}
			return nil, nil
		}}
		o, _ := newTestOps(t, &config.Config{}, ft)
		_, err := o.ImportVerify(context.Background(), config.Primary, plan)
		if err == nil || !strings.Contains(err.Error(), "re-export the target for verification") {
			t.Errorf("err = %v, want it wrapped as a re-export failure", err)
		}
	})

	t.Run("re-export does not parse", func(t *testing.T) {
		ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
			if matchCLI(argv, "verify-export") {
				// Has its BEGIN, so it gets past stripTranscript and fails in
				// ParseBlocks on the missing terminator -- the parse path this
				// subtest is about.
				return []byte(" ! BEGIN\nnot a real capture, missing its terminator\n"), nil
			}
			return nil, nil
		}}
		o, _ := newTestOps(t, &config.Config{}, ft)
		_, err := o.ImportVerify(context.Background(), config.Primary, plan)
		if err == nil || !strings.Contains(err.Error(), "parse the verification re-export") {
			t.Errorf("err = %v, want it wrapped as a parse failure", err)
		}
	})

	// A re-export with NO ` ! BEGIN` is not show current-config output at all -- the
	// CLI printed something else -- and is refused by the capture step, before any
	// parse. RISK: without the cut, the banner and echoed prompts would be diffed as
	// if they were configuration and every verification would report them missing.
	t.Run("re-export is not a capture", func(t *testing.T) {
		ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
			if matchCLI(argv, "verify-export") {
				return testBannerSoftware, nil
			}
			return nil, nil
		}}
		o, _ := newTestOps(t, &config.Config{}, ft)
		_, err := o.ImportVerify(context.Background(), config.Primary, plan)
		if err == nil || !strings.Contains(err.Error(), "re-export the target for verification") ||
			!strings.Contains(err.Error(), "! BEGIN") {
			t.Errorf("err = %v, want the capture step to refuse output with no `! BEGIN`", err)
		}
	})
}

// --- renderScript / importIgnore -----------------------------------------

// TestImportOpsRenderSectionsAppliesPreambleAndDropsFilteredLines covers what
// ends up in an applied script: the mode preamble (home/enable/configure) is
// included, a header comment is not, and an apply-filtered rule's excluded
// line ("Configure Routing" drops its `interface "intf0"` operand, per
// sections.go, because that interface belongs to a skip-appliance section) is
// dropped from the rendered body even though the rest of the block is kept.
func TestImportOpsRenderScriptAppliesPreambleAndDropsFilteredLines(t *testing.T) {
	blocks := []Block{{
		Section: "Configure Routing",
		Region:  RegionBroker,
		Kind:    "routing",
		Lines:   []string{"routing", `  interface "intf0"`, "  exit"},
	}}
	p := &ImportPlan{Capture: &Capture{
		Preamble: []string{"! a broker header comment", "", "home", "enable", "configure"},
		Blocks:   blocks,
	}}
	body, applied := renderScript(blocks, p, []string{"Configure Routing"})

	if len(applied) != 1 || applied[0] != "Configure Routing" {
		t.Fatalf("renderScript applied = %v, want [Configure Routing]", applied)
	}
	for _, want := range []string{"home\n", "enable\n", "configure\n", "routing\n", "  exit\n"} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered body missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, `interface "intf0"`) {
		t.Errorf("renderScript must drop the apply-filtered line, got:\n%s", body)
	}
	if strings.Contains(body, "a broker header comment") {
		t.Errorf("renderScript must not carry a comment line from the preamble, got:\n%s", body)
	}
}

// TestImportOpsRenderScriptEmptySelectionRendersNothing covers the no-op
// return: a caller (ImportApply, for the "first" invocation when nothing is
// First) must be able to tell "nothing to apply" from "an empty script", since
// the two only differ here on whether a RunCLI is skipped entirely.
func TestImportOpsRenderScriptEmptySelectionRendersNothing(t *testing.T) {
	p := &ImportPlan{Capture: &Capture{Blocks: []Block{
		{Section: "Configure Routing", Region: RegionBroker, Lines: []string{"routing"}},
	}}}
	body, applied := renderScript(p.Capture.Blocks, p, nil)
	if body != "" || applied != nil {
		t.Errorf("renderScript(nil names) = (%q, %v), want (\"\", nil)", body, applied)
	}
}

// TestImportOpsVPNChunkCarriesVPNRegionsAndBrokerChunkDoesNot is the regression test
// for the worst bug this feature had: renderScript's predecessor filtered every
// block with `Region != RegionBroker`, so ImportApply tore a conflicting message-VPN
// down via the target's own `remove` script and then applied NOTHING back. The
// command was purely destructive and reported success.
//
// The section classification is a broker-level regime only -- no VPN section is in
// the table, so gating VPN blocks on it can only ever drop them. The two renderers
// are now separate functions rather than one with a flag, which is what makes that
// mistake unavailable: renderVPNScript selects by VPN and never consults the section
// names, and renderScript is broker-level and cannot emit a VPN block at all.
func TestImportOpsVPNChunkCarriesVPNRegionsAndBrokerChunkDoesNot(t *testing.T) {
	blocks := []Block{
		{Section: "Create All: \"Message Vpns\"", Region: RegionVPNCreate, Kind: "message-vpn",
			Name: "acme", VPN: "acme", Create: true,
			Lines: []string{`create message-vpn "acme"`, "  exit"}},
		{Section: "Create logging", Region: RegionBroker, Kind: "logging",
			Lines: []string{`logging command "CLI" mode "config-cmds"`}},
		{Section: "Create Client Username: \"acme\"", Region: RegionVPN, Kind: "client-username",
			Name: "app1", VPN: "acme",
			Lines: []string{`client-username "app1" message-vpn "acme"`, "  no shutdown", "  exit"}},
	}
	p := &ImportPlan{Capture: &Capture{
		Preamble: []string{"home", "enable", "configure"},
		Blocks:   blocks,
	}}

	// Both VPN regions must still be rendered SOMEWHERE, or a torn-down VPN would not
	// be rebuilt -- but they are deliberately in two different chunks now. The create
	// pass has to precede the broker-level sections (a broker line naming a VPN is
	// rejected if the VPN does not exist), while the VPN's own objects come after, so
	// one renderer cannot serve both. What matters is that nothing falls between them.
	create := renderRegionScript(blocks, p, RegionVPNCreate, "")
	if !strings.Contains(create, `create message-vpn "acme"`) {
		t.Errorf("the create-VPN chunk is missing the create -- a torn-down VPN would not be rebuilt:\n%s", create)
	}
	body := renderVPNScript(blocks, p, "acme")
	if !strings.Contains(body, `client-username "app1" message-vpn "acme"`) {
		t.Errorf("the VPN chunk is missing the VPN's own objects:\n%s", body)
	}
	// And the create must not ALSO be in the VPN chunk: applied twice, the second is
	// a rejected duplicate.
	if strings.Contains(body, `create message-vpn "acme"`) {
		t.Errorf("the VPN chunk re-creates the VPN the create-VPN chunk already made:\n%s", body)
	}

	// The isolated "first" chunk must carry ONLY its broker section: sharing it
	// with the VPN work would put the VPN rebuild in the same CLI session as the
	// logging change that ends that session.
	first, _ := renderScript(blocks, p, []string{"Create logging"})
	if !strings.Contains(first, `logging command "CLI"`) {
		t.Errorf("the first invocation lost its own section:\n%s", first)
	}
	for _, unwanted := range []string{`create message-vpn "acme"`, `client-username "app1"`} {
		if strings.Contains(first, unwanted) {
			t.Errorf("the isolated first invocation must not carry VPN work, found %q:\n%s", unwanted, first)
		}
	}
}

// TestImportOpsRenderScriptDoesNotFilterVPNLines pins the other half of the same
// rule: DropLines belongs to a broker section's rule, so it must never be applied
// to a VPN block. A VPN block whose text happens to match a broker predicate --
// a queue's own `max-spool-usage`, which the Message Spool rule drops at broker
// level -- must survive intact, or importing a queue would silently lose its
// quota.
func TestImportOpsRenderScriptDoesNotFilterVPNLines(t *testing.T) {
	blocks := []Block{{
		Section: "Create Endpoints: \"acme\"", Region: RegionVPN,
		Kind: "message-spool", VPN: "acme",
		Lines: []string{
			`message-spool message-vpn "acme"`,
			`  create queue "orders"`,
			"    max-spool-usage 5000",
			"    exit",
			"  exit",
		},
	}}
	p := &ImportPlan{Capture: &Capture{Preamble: []string{"home", "enable", "configure"}, Blocks: blocks}}

	body := renderVPNScript(blocks, p, "acme")
	if !strings.Contains(body, "max-spool-usage 5000") {
		t.Errorf("a queue's own max-spool-usage was dropped by a BROKER-level filter:\n%s", body)
	}
}

// TestImportOpsImportIgnore covers the three outcomes importIgnore feeds to
// the verification diff: every line of a Skip/SkipAppliance section is
// ignored, an ApplyFiltered rule's own dropped line is ignored, and any other
// line is not. RISK: a line the import never intended to land, if not
// ignored, reads as a verification FAILURE on every single import -- the
// opposite mistake (ignoring a line that should have landed) would hide a real
// one.
func TestImportOpsImportIgnore(t *testing.T) {
	p := &ImportPlan{Sections: []PlannedSection{
		{Name: "Configure System", Disposition: Skip},
		{Name: "Configure Routing", Disposition: ApplyFiltered},
		{Name: "Create console", Disposition: Apply},
	}}
	ignore := importIgnore(p)

	if !ignore(Block{Section: "Configure System"}, "system scaling  scale  max-connections 1000") {
		t.Error("importIgnore should ignore every line of a Skip section")
	}

	filtered := Block{Section: "Configure Routing"}
	if !ignore(filtered, `  interface "intf0"`) {
		t.Error("importIgnore should ignore the line an apply-filtered rule drops")
	}
	if ignore(filtered, "  routing mode enable") {
		t.Error("importIgnore must not ignore a KEPT line of an apply-filtered section")
	}

	if ignore(Block{Section: "Create console"}, "create console") {
		t.Error("importIgnore must not ignore a line of a fully-applied section")
	}
}

// --- small pure helpers -------------------------------------------------------

// TestImportOpsDescribeScope is table-driven over describeScope's three
// mutually exclusive forms, since every one of them ends up baked into the
// artifact's own metadata (Annotate's "scope" field) and read back by a human
// deciding whether to trust it.
func TestImportOpsDescribeScope(t *testing.T) {
	cases := []struct {
		vpns       []string
		brokerOnly bool
		want       string
	}{
		{nil, false, "total"},
		{nil, true, "broker"},
		{[]string{"b", "a"}, false, "vpn:a,b"}, // sorted, regardless of input order
		{[]string{"a"}, true, "broker"},        // brokerOnly wins even if vpns is also set
	}
	for _, c := range cases {
		if got := describeScope(c.vpns, c.brokerOnly); got != c.want {
			t.Errorf("describeScope(%v, %v) = %q, want %q", c.vpns, c.brokerOnly, got, c.want)
		}
	}
}

// --- Report methods ----------------------------------------------------------

// TestImportOpsReportsDoNotPanicOnEmptyAndPopulated covers every Report method
// this file defines against both an empty and a populated value, over a real
// output.Sink. A KVBlock/TaggedTable call that assumes a non-empty slice
// somewhere would only panic on the empty case, which a test that only tries
// the populated shape would never catch. The populated ImportPlan case also
// pins that the VPN being torn down is actually named in the rendered report,
// not merely counted -- an operator confirming a destructive import needs to
// see WHICH VPN loses its queues.
func TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(t *testing.T) {
	var buf bytes.Buffer
	sink := output.New(&buf)
	o := &Ops{}

	// ExportReport: whole-broker and broker-only scopes, through the surface an
	// operator actually reads.
	o.ExportReport(sink, nil, false, 123)
	o.ExportReport(sink, []string{"acme"}, true, 456)

	// Empty plan / empty result: nothing to report, must not panic.
	(&ImportPlan{Capture: &Capture{Type: BrokerSoftware, Version: "1.0"}}).Report(sink)
	(&ImportResult{}).Report(sink)

	// Populated plan: must name the VPN it plans to tear down.
	plan := &ImportPlan{
		Capture:  &Capture{Type: BrokerSoftware, Version: "10.26.0.8827"},
		Existing: []string{"acme"},
		NewVPNs:  []string{"beta"},
		Sections: []PlannedSection{{Name: "Create console", Disposition: Apply, Reason: "reason"}},
	}
	plan.Report(sink)
	if !strings.Contains(buf.String(), `"acme"`) {
		t.Errorf("a populated ImportPlan report should name the VPN being torn down, got %q", buf.String())
	}

	// Populated result: must not panic either, including a warning line.
	(&ImportResult{
		TornDown: []string{"acme"},
		Applied:  []string{"Create console"},
		Warnings: []string{"a warning to render"},
	}).Report(sink)
}

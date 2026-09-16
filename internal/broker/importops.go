package broker

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"solace/internal/config"
	"solace/internal/output"
)

// The ops behind `broker perform export-config` and `import-config`.
//
// Everything here goes over the broker CLI: upload a script, run `cli -Apes`, take
// the captured stdout. Every capture (runCapture) is a READ and goes through
// runCLIRead, RunCLI's unwrapped sibling -- RunCLI's own stop-on-error wrapper has
// nothing to protect in a `show current-config` and would change the transcript
// shape stripTranscript and bannerType are tuned against. The one WRITE this file
// issues, the generated import driver script (runDriver), goes through
// Transport.Output directly rather than RunCLI, because it is not a `cli -Apes`
// invocation at all -- see runDriver's own doc comment. Three properties of the
// CLI channel shape the whole design:
//
//   - It needs NO broker credentials. Authorisation is the transport's -- kubectl
//     or docker exec into the container -- so an import that changes the CLI admin
//     password cannot lock the tool out of its own run, and verification still
//     works afterwards.
//   - Every RunCLI execs a NEW cli process, so each invocation is a fresh session.
//     That is what makes it safe to apply a section that ends the session (logging)
//     by giving it an invocation of its own.
//   - `cli -Apes` exits 0 even when the broker rejects a line, so no exit status
//     anywhere here means the configuration landed. There are two detectors and they
//     catch different things. The apply driver (driver.go) runs each chunk through
//     the broker's own `source script ... stop-on-error` and reads the transcript
//     tail, which stops a bad line poisoning the rest of its chunk and is the ONLY
//     way a failed TEARDOWN is ever noticed -- the diff cannot see one, because a VPN
//     that was never removed still satisfies every line the artifact asks for.
//     ImportVerify -- a re-export and a diff -- remains the verdict on whether the
//     configuration is actually THERE. Neither is optional.

// applianceModelRE matches an appliance product line: "Solace PubSub+ 3560 Version".
// A software broker's line reads "Solace Event Broker Software Enterprise Version"
// and never carries a model number in that position.
var applianceModelRE = regexp.MustCompile(`Solace PubSub\+ [0-9]{3,5} Version`)

// bannerType reads the broker type off the CLI login banner.
//
// Every `cli -Apes` invocation echoes the banner before the script's own output,
// so the type is available from ANY CLI round trip and no dedicated command is
// needed. Verified against real output: a software broker opens
// "Solace Event Broker Software Enterprise Version 10.26.0.8827" and cites
// http://www.solace.com/license-software; an appliance opens
// "Solace PubSub+ 3560 Version 10.25.0.200" and cites license-hardware.
//
// The licence URL is the primary marker because it is the least likely to be
// reworded; the product line is the fallback. Neither found -> BrokerUnknown, and
// import REFUSES rather than assuming the common case -- assuming "software" for an
// appliance is precisely the cross-type import the refusal exists to stop. (An
// earlier version scanned `show version` for a SEMP schema string that its output
// does not contain, which is what made a first real import refuse a valid target.)
func bannerType(out []byte) BrokerType {
	s := string(out)
	switch {
	case strings.Contains(s, "license-software"), strings.Contains(s, "Event Broker Software"):
		return BrokerSoftware
	case strings.Contains(s, "license-hardware"), applianceModelRE.MatchString(s):
		return BrokerAppliance
	}
	return BrokerUnknown
}

// stripTranscript drops the CLI transcript that precedes a capture's own header.
//
// `cli -Apes` echoes the login banner, the licence text and every prompt+command
// (`xps-ps-01> home`) into stdout before the script's output. None of that is
// configuration, and fed back to the CLI as a script it would be rejected line by
// line -- invisibly, since `cli -Apes` exits 0 regardless. The capture proper begins
// at the broker's own ` ! BEGIN` line (leading space and all), so everything before
// it goes. Output with no BEGIN is not `show current-config` output and is refused.
//
// This runs on every capture -- export, teardown and the verification re-export --
// so the artifact, the applied teardown and the diffed blocks all start clean.
func stripTranscript(out []byte) ([]byte, error) {
	lines := splitLines(out)
	for i, line := range lines {
		if strings.TrimSpace(line) == "! BEGIN" {
			return []byte(strings.Join(lines[i:], "\n")), nil
		}
	}
	return nil, fmt.Errorf("the broker's output has no `! BEGIN` header, so it is not "+
		"`show current-config` output; the transcript ended:\n%s", lastLines(string(out), 6))
}

// preambleForApply keeps only what an apply script needs from a captured preamble:
// the mode commands and the interpreter pragmas, in order.
//
// A captured preamble also carries the broker's `! Command:` / `! Generated by:`
// header, which is inert. The pragma is NOT inert and must survive:
// `! pragma:interpreter:relaxed-db-restore` is what makes the interpreter tolerate
// a replayed configuration, and an earlier filter that dropped every `!` line to
// remove the header dropped the pragma with it. Whitelisting is also what keeps a
// hand-edited artifact that still carries a transcript from feeding banner text to
// the CLI as commands.
// `no paging` is INSERTED after the captured `home` rather than taken from the artifact.
// A capture's own preamble carries whatever the exporting session had set, and an apply
// that paginates stalls waiting for a keypress nobody is there to give -- so the setting
// is this tool's to make, not the artifact's to supply. cliHome is the same pair every
// generated script opens with.
func preambleForApply(preamble []string) []string {
	var out []string
	for _, line := range preamble {
		t := strings.TrimSpace(line)
		switch {
		case t == "home":
			out = append(out, line, "no paging")
		case t == "no paging":
			// Already added beside the `home` above; a second one is inert but would
			// make the emitted script differ depending on how the capture was taken.
		case t == "enable" || t == "configure" || isPragma(t):
			out = append(out, line)
		}
	}
	return out
}

// ExportConfig captures the broker's configuration and returns the marked artifact.
//
// stamp is passed in rather than read from a clock here, so this package stays
// free of one: the CLI already has nowStamp(), and a caller-supplied stamp is what
// lets a golden test pin the whole artifact byte-for-byte.
func (o *Ops) ExportConfig(ctx context.Context, role config.Role, vpns []string, brokerOnly bool, stamp string) ([]byte, error) {
	for _, v := range vpns {
		if err := validVPNName(v); err != nil {
			return nil, err
		}
	}

	c, err := o.captureScope(ctx, role, vpns)
	if err != nil {
		return nil, err
	}
	if brokerOnly {
		// RegionBroker only: the create-VPN pass must go too, or a broker-only artifact
		// would still create every message-VPN on the target.
		c.Blocks = keepRegion(c.Blocks, RegionBroker)
	}
	c.Blocks, c.Omitted = omitAtExport(c.Blocks)
	return Annotate(c, describeScope(vpns, brokerOnly), stamp), nil
}

// omitAtExport removes from a capture what import would never apply, and returns
// what it removed.
//
// This is where `skip` is enforced. The artifact is then exactly what an import
// applies: nothing in it is dead weight, and a reader is not misled by a section
// that looks like configuration but is never configuration on the target. Import
// still ignores these if it meets them (an older or hand-edited artifact), but
// that is the safety net -- removal here is the primary path, and the import-side
// report mentions a skipped section only when one is actually present, which after
// this should not happen.
//
// Two things are deliberately NOT removed:
//   - SkipAppliance sections. That is a decision about the TARGET's type, and only
//     import knows the target; on a software capture they are empty anyway.
//   - Unclassified sections. Import refuses to apply those and says so; removing
//     them silently here would hide a broker release this build does not know.
//
// ApplyFiltered sections lose their dropped LINES here for the same reason, and a
// block that ends up with no lines at all is dropped with them -- Message Spool's
// `max-spool-usage` is a one-liner, and an opener-less husk would parse as nothing.
func omitAtExport(blocks []Block) ([]Block, []Omission) {
	kept := make([]Block, 0, len(blocks))
	var omitted []Omission
	seen := map[string]bool{}

	for _, b := range blocks {
		if b.Region != RegionBroker {
			kept = append(kept, b)
			continue
		}
		rule, ok := RuleFor(b.Section)
		if !ok {
			kept = append(kept, b)
			continue
		}
		switch rule.Disposition {
		case Skip:
			if !seen[b.Section] {
				seen[b.Section] = true
				omitted = append(omitted, Omission{Section: b.Section, Reason: rule.Reason})
			}
			continue
		case ApplyFiltered:
			if rule.DropLines != nil {
				lines := make([]string, 0, len(b.Lines))
				for _, l := range b.Lines {
					if !rule.DropLines(l) {
						lines = append(lines, l)
					}
				}
				if len(lines) == 0 {
					continue
				}
				b.Lines = lines
			}
		}
		kept = append(kept, b)
	}
	return kept, omitted
}

// captureScope runs the capture the scope needs and returns it parsed.
//
// A VPN list is captured one `show current-config message-vpn` at a time rather
// than by taking the whole broker and filtering, because the broker already offers
// exactly that scope and asking it for less is cheaper than asking for everything
// and throwing most of it away -- `show current-config all` on a real broker is
// tens of thousands of lines.
func (o *Ops) captureScope(ctx context.Context, role config.Role, vpns []string) (*Capture, error) {
	if len(vpns) == 0 {
		// The broker type runCapture also reads off this same round trip is not
		// needed here -- ExportConfig never has to compare types, only ImportPlan
		// and ConfigureReplication do -- so it is discarded.
		out, _, err := o.runCapture(ctx, role, "export-config", "", false)
		if err != nil {
			return nil, err
		}
		return ParseBlocks(out)
	}

	var combined *Capture
	for _, v := range vpns {
		out, _, err := o.runCapture(ctx, role, "export-config", v, false)
		if err != nil {
			return nil, err
		}
		c, err := ParseBlocks(out)
		if err != nil {
			return nil, fmt.Errorf("parse the capture of message-VPN %q: %w", v, err)
		}
		if combined == nil {
			combined = c
			continue
		}
		combined.Blocks = append(combined.Blocks, c.Blocks...)
	}
	return combined, nil
}

// runCapture runs one `show current-config` and returns its stdout, plus the
// broker type read off the same round trip (see bannerType). It goes through
// runCLIRead, not RunCLI: a capture is a read with nothing for RunCLI's
// stop-on-error wrapper to protect, and stripTranscript's `! BEGIN` search and
// bannerType's own banner scan are both tuned against the unwrapped shape a plain
// `cli -Apes` invocation produces.
//
// name is the uploaded script's own name, passed in rather than derived, because
// the four reads this feature makes are four different operations against the
// broker and a shared name would make them indistinguishable in a log and in a
// test's transport double. Two of them read the SOURCE (export) and two read the
// TARGET (the plan's VPN listing, and the verification re-export); a test that
// cannot tell those apart cannot express "the artifact says X and the target says
// Y", which is the only interesting case the plan has.
func (o *Ops) runCapture(ctx context.Context, role config.Role, name, vpn string, remove bool) ([]byte, BrokerType, error) {
	out, err := o.readCLI(ctx, role, name, currentConfigScript(vpn, remove))
	if err != nil {
		return nil, BrokerUnknown, err
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		return nil, BrokerUnknown, fmt.Errorf("the broker returned no output for `show current-config`; " +
			"the capture cannot be empty, so treat this as a failed capture rather than an " +
			"empty configuration")
	}
	// Read before stripTranscript discards everything ahead of the capture's own
	// `! BEGIN` -- the banner it reads is exactly what stripTranscript throws away.
	typ := bannerType(out)
	stripped, err := stripTranscript(out)
	if err != nil {
		return nil, typ, err
	}
	return stripped, typ, nil
}

// ImportPlan is what the preflight decided, and what the confirmation quotes back.
type ImportPlan struct {
	Capture  *Capture
	Existing []string // VPNs in the artifact that already exist on the target: these are TORN DOWN
	NewVPNs  []string // VPNs in the artifact the target does not have
	Sections []PlannedSection
	Warnings []string
	// Target is what the target broker already has, read at plan time from its own
	// capture. Everything in it is something the apply cannot decide from the
	// artifact -- see TargetState.
	Target TargetState
}

// PlannedSection is one section's resolved disposition.
type PlannedSection struct {
	Name        string
	Disposition Disposition
	Reason      string
	Interrupts  string
	First       bool
	Blocks      int
	// Unclassified marks a section this build's table has no rule for. It is
	// dispositioned Skip, because applying configuration nobody classified is the
	// hazard sections.go exists to prevent -- but it is NOT the same as a rule that
	// says skip, and importIgnore has to tell them apart. A rule-based skip is a
	// decision, so its lines are excluded from verification; an unclassified section
	// is a GAP, so its lines stay in the diff and the command exits non-zero naming
	// them. Otherwise a broker carrying configuration this build does not know gets
	// it silently dropped and is then reported as a clean import.
	Unclassified bool
}

// ExistingList renders the conflicting VPNs for the confirmation prompt.
func (p *ImportPlan) ExistingList() string { return quoteList(p.Existing) }

// NewList renders the VPNs that will be created.
func (p *ImportPlan) NewList() string { return quoteList(p.NewVPNs) }

// quoteList is replVPNList with no empty-case wording: the callers here wrap an empty
// result in orNone themselves, so there is one quoting rule rather than two.
func quoteList(names []string) string { return replVPNList(names, "") }

// omittedList renders the sections export removed, for the import plan report.
func omittedList(oms []Omission) string {
	names := make([]string, len(oms))
	for i, om := range oms {
		names[i] = om.Section
	}
	return strings.Join(names, ", ")
}

// ImportPlan reads the artifact and the target, and decides what an import would
// do -- without writing anything.
//
// It is separate from ImportApply because every confirmation lives in internal/cli
// (Ops has no Confirm seam, deliberately), so the question has to be asked between
// reading the target and writing to it.
func (o *Ops) ImportPlan(ctx context.Context, role config.Role, artifact []byte) (*ImportPlan, error) {
	c, err := ParseBlocks(artifact)
	if err != nil {
		return nil, err
	}
	if err := checkProvenance(c); err != nil {
		return nil, err
	}
	// Export never asks the broker to redact, because redaction strips exactly the
	// credential material an import has to put back. This guard is for the file that
	// says it was redacted anyway -- a hand-edited header, or a capture taken by hand
	// -- so import says why rather than applying a configuration whose passwords are
	// all gone.
	if c.Redacted {
		return nil, fmt.Errorf("this artifact says it was captured with `redact`, so the broker " +
			"stripped the credential material an import has to put back; re-export it with " +
			"`broker perform export-config`")
	}

	// The target's broker type comes off the SAME round trip readTarget already
	// makes to list its VPNs (runCapture reads it off the banner every `cli -Apes`
	// echoes ahead of its output) -- there is no longer a dedicated banner-only
	// call spent just to learn it.
	target, targetType, err := o.readTarget(ctx, role)
	if err != nil {
		return nil, err
	}
	if err := checkSameType(c.Type, targetType); err != nil {
		return nil, err
	}

	// The target's own objects decide which removals are valid clears and which are
	// errors, so they are read here, once, from the same capture as its VPNs.
	plan := &ImportPlan{Capture: c, Target: readTargetState(target)}
	have := map[string]bool{}
	for _, v := range target.VPNs() {
		have[v] = true
	}
	for _, v := range c.VPNs() {
		if have[v] {
			plan.Existing = append(plan.Existing, v)
		} else {
			plan.NewVPNs = append(plan.NewVPNs, v)
		}
	}
	plan.Sections, plan.Warnings = planSections(c.Blocks, targetType)
	return plan, nil
}

// planSections resolves every section in the artifact against the baked
// classification.
//
// The disposition is resolved HERE, from this build's table, and never read from
// the artifact's advisory marker -- see annotate.go for why a file that chose what
// executes would be the hazard execguard.go exists to prevent. A marker that
// disagrees produces a warning, not a different outcome.
func planSections(blocks []Block, targetType BrokerType) ([]PlannedSection, []string) {
	var (
		out      []PlannedSection
		warnings []string
		seen     = map[string]int{}
		order    []string
	)
	for _, b := range blocks {
		if b.Region != RegionBroker {
			continue
		}
		if _, ok := seen[b.Section]; !ok {
			order = append(order, b.Section)
		}
		seen[b.Section]++
	}

	for _, name := range order {
		rule, ok := RuleFor(name)
		if !ok {
			// An unclassified non-empty section is a loud gap, not a default-apply: the
			// classification is the whole safety story, so a section it does not cover is
			// something a human has to look at.
			warnings = append(warnings, fmt.Sprintf("section %q is not classified, so it is NOT "+
				"applied; this artifact is from a broker whose sections this build does not know", name))
			out = append(out, PlannedSection{Name: name, Disposition: Skip, Unclassified: true,
				Reason: "not classified by this build", Blocks: seen[name]})
			continue
		}
		d := rule.Disposition
		if d == SkipAppliance && targetType == BrokerAppliance {
			// The section is appliance-only and the target IS an appliance, so it applies.
			d = Apply
		}
		out = append(out, PlannedSection{
			Name: name, Disposition: d, Reason: rule.Reason,
			Interrupts: rule.Interrupts, First: rule.First, Blocks: seen[name],
		})
	}
	return out, warnings
}

// checkProvenance refuses anything this utility did not export.
//
// Import applies its input to a live broker through the CLI, so what it accepts is
// the shape of everything it will run. A raw `show current-config` capture, a
// hand-written script or a file that merely ends in .cli is not an artifact: it has
// not been through the export path, so its Skip sections were never removed, its
// transcript was never cut, and nothing has established which broker it came from.
// Requiring the export marker keeps `import-config` to the one input it was
// designed around, and leaves running an arbitrary script to the command whose job
// that is -- `broker perform cli-script`, which says as much in its own help.
//
// This is a PROVENANCE check, not a tamper check. The marker is a comment: anyone
// editing an artifact can keep it, and nothing here would notice. It stops the
// wrong file being imported, not a determined edit -- and it could not do more
// while cli-script exists, since the same operator can run any script they like
// through that instead.
func checkProvenance(c *Capture) error {
	if !c.Marked {
		return fmt.Errorf("this file was not produced by `broker perform export-config`: it "+
			"carries no `! %s:` marker.\n"+
			"  Import only applies artifacts from that command, because an export removes the "+
			"sections import must never apply and records which broker the configuration came "+
			"from -- a raw `show current-config` capture has had neither done to it.\n"+
			"  Re-export from the source broker, or run an arbitrary script with "+
			"`broker perform cli-script` instead", markerExport)
	}

	// A newer build's artifact may carry markers this one does not understand, and
	// guessing at them would apply an artifact whose shape was never agreed. The
	// version is the one thing an older build can always read, so it is what says no.
	got := c.Meta["export-version"]
	if got != fmt.Sprintf("%d", markerVersion) {
		return fmt.Errorf("this artifact is export format version %q and this build reads "+
			"version %d; re-export it with this build, or use the one that wrote it",
			got, markerVersion)
	}
	return nil
}

// checkSameType refuses a cross-type import.
func checkSameType(artifact, target BrokerType) error {
	if artifact == BrokerUnknown {
		return fmt.Errorf("cannot tell which broker type this artifact came from: its SEMP schema " +
			"version is missing or unrecognised, and importing it could apply appliance-only " +
			"configuration to a software broker or the reverse")
	}
	if target == BrokerUnknown {
		return fmt.Errorf("cannot tell which broker type the target is, so the import is refused " +
			"rather than risking a cross-type apply")
	}
	if artifact != target {
		return fmt.Errorf("this artifact was captured from a %s broker and the target is a %s "+
			"broker; the two do not share a configuration surface (interfaces, VRF, DNS, clock and "+
			"SNMP exist on one and not the other), so the import is refused rather than partly applied",
			artifact, target)
	}
	return nil
}

// readTarget reads the target broker's own configuration.
//
// It reads them from a full `show current-config all` capture through ParseBlocks,
// NOT from the `show message-vpn *` table. That is a deliberate second full read on
// the destructive path, and it buys the one property that matters here: the SAME
// quote-aware parser reads the target and the artifact, so the two sides cannot
// disagree about what a VPN is called. `create message-vpn "X"` sets both Name and
// VPN (newBlock, blocks.go), and the create-all-VPNs pass emits one for every VPN,
// so a full capture names them all.
//
// The table parser it replaced split that name column on whitespace, so this
// broker's own `A VPN WITH LONG NAME AND SPACES` came back as seven names and
// matched nothing. Classifying an EXISTING VPN as new is the worst outcome this
// feature has: it is not torn down, the artifact's `create` blocks are applied on
// top of a live VPN and rejected invisibly (`cli -Apes` exits 0), and ImportVerify
// then PASSES -- the untouched VPN still carries every line the artifact asked for.
// A wrong answer here does not merely misreport, it defeats the only error
// detection that exists, so it does not get a heuristic. parseVPNNames survives for
// broker.Config's far less consequential use of the same output.
//
// Every failure is loud, and that is the other half. "I could not tell" must never
// reach the caller as an empty list: empty means "no VPN needs tearing down", which
// silently downgrades the import to apply-on-top -- the same catastrophe by a
// quieter route. runCapture already refuses an empty capture for the same reason.
// It returns the whole parsed capture rather than just the names, because the plan
// needs more than one fact about the target from it -- the message-VPNs to tear
// down, and the syslog entries that decide whether a `no syslog` is valid or an
// error (ClearExistingSyslogs). One read answers both; asking twice would be two
// more full captures on the destructive path.
//
// It also returns the target's BrokerType, which runCapture reads off this same
// capture's own login banner. ImportPlan used to spend a separate dedicated
// `RunCLI(ctx, role, "banner", "home\n")` round trip on that alone (the removed
// Ops.brokerType); now it comes free off the round trip this already makes.
func (o *Ops) readTarget(ctx context.Context, role config.Role) (*Capture, BrokerType, error) {
	out, typ, err := o.runCapture(ctx, role, "plan-vpns", "", false)
	if err != nil {
		return nil, BrokerUnknown, fmt.Errorf("read the target's configuration to list its message-VPNs: %w", err)
	}
	c, err := ParseBlocks(StripMarkers(out))
	if err != nil {
		return nil, BrokerUnknown, fmt.Errorf("parse the target's configuration to list its message-VPNs: %w", err)
	}
	return c, typ, nil
}

// ImportResult is what an apply did.
type ImportResult struct {
	TornDown []string
	// Rebuilt is set only once the artifact's own blocks have been applied. The
	// teardown loop runs to completion BEFORE any rebuild starts, so on every
	// failure path TornDown is full and this is false -- those VPNs are emptied and
	// nothing has been put back, and the report has to say so rather than describe
	// them as "torn down and rebuilt".
	Rebuilt bool
	Applied []string
	// Failed names the chunk the broker rejected a line in, and FailedDetail is the
	// transcript tail the driver echoed back for it. Empty on a clean run. This is
	// the signal import never had: `cli -Apes` exits 0 on a rejected line, so before
	// the driver there was nothing to report between "the CLI would not run" and the
	// diff's verdict much later.
	Failed       string
	FailedDetail string
	Skipped      []PlannedSection
	Warnings     []string
}

// ImportApply tears down each conflicting VPN and applies the artifact.
//
// The teardown script is not written here: `show current-config message-vpn <n>
// remove` is captured FROM THE TARGET, so the broker orders the removal of its own
// queues, ACL profiles, client profiles, usernames and bridges. Nothing in this
// package needs to know that a queue goes before its ACL profile.
//
// That is also what makes the `default` message-VPN safe with no special case, and
// it is the strongest argument for capturing the teardown rather than synthesising
// one. `default` cannot be deleted, and neither can the `default` client-profile,
// acl-profile or client-username inside any VPN -- which is why this repo's
// `broker configure default-vpn` SHUTS THE VPN DOWN rather than removing it.
// VERIFIED ON A LIVE BROKER: the generated `remove` script simply does
// not emit lines for the VPN itself or for its undeletable objects, so it empties
// `default` and leaves it present. A synthesised teardown would have had to carry a
// list of reserved names and keep it correct forever; the broker already knows.
//
// The artifact then applies onto an emptied-but-present VPN, and its
// `create message-vpn "default"` line is covered by the broker's own
// `! pragma:interpreter:ignore-already-exists` wrapper.
//
// Two things are still NEEDS VERIFICATION ON A LIVE BROKER rather than asserted. That a
// `remove` capture for a NON-DEFAULT VPN removes the VPN itself and leaves no residue --
// the default case above is confirmed, the general one is not, and they are not the same
// question because `default` is the VPN that cannot be deleted. And that a VPN-scoped
// import does not intersect what `render.BrokerCR` owns on Kubernetes, where the operator
// reconciles the CR and would put back whatever it believes it owns.
//
// The apply itself is CHUNKED and runs entirely inside the broker. Everything the
// artifact needs is rendered into one generated shell script (driver.go) that is
// uploaded once and executed once, and that script applies each chunk through the
// broker's own `source script ... stop-on-error`. Three things follow, and none of
// them was true of the single-script apply this replaces:
//
//   - The BROKER stops at the first rejected line in a chunk, instead of running on
//     through everything after it.
//   - Each chunk's transcript tail is checked for a rejection there and then, so a
//     failure is detected during the apply rather than inferred from the diff -- and
//     a failed TEARDOWN is detected at all, which the diff structurally cannot do (a
//     VPN that was never removed still satisfies every artifact line).
//   - A real import costs one exec instead of one per chunk.
//
// The teardown CAPTURES above still cost an exec each: their content is what the
// broker decides, so it cannot be known before asking.
func (o *Ops) ImportApply(ctx context.Context, role config.Role, p *ImportPlan) (*ImportResult, error) {
	res := &ImportResult{Warnings: p.Warnings}

	// Make any bare `listen-port` change safe, then apply the replay transformations
	// ported from the operator's own migration GUI (transform.go). targetState is nil
	// deliberately: reading the target's per-service enablement would be another
	// output parser this tool has not verified, and InjectShutdown's contract for an
	// unknown state is to inject NOTHING and warn -- which is the right direction to
	// be wrong in.
	//
	// The result is a LOCAL slice and is deliberately not written back into the plan.
	// p.Capture.Blocks is the artifact, and it is what ImportVerify diffs the target
	// against; folding synthesised lines into it would make verification demand that
	// the target contain lines the artifact never had.
	blocks, warn, err := InjectShutdown(p.Capture.Blocks, nil)
	if err != nil {
		return res, err
	}
	res.Warnings = append(res.Warnings, warn...)
	blocks = PrepareForApply(blocks, p.Target)

	chunks, err := o.buildChunks(ctx, role, p, blocks, res)
	if err != nil {
		return res, err
	}
	if len(chunks) == 0 {
		res.Rebuilt = true
		res.Skipped = skippedSections(p)
		return res, nil
	}

	results, err := o.runDriver(ctx, role, chunks)
	// Whatever happened, every chunk the driver DID report is what actually ran, so
	// the report is filled in before any error is returned. An import that failed
	// half way is exactly when an operator needs to know which half.
	for _, r := range results {
		if !r.OK {
			res.Failed, res.FailedDetail = r.Name, r.Detail
			break
		}
		// Only a teardown the driver reported OK actually ran.
		if vpn := chunks[r.Index].TearsDown; vpn != "" {
			res.TornDown = append(res.TornDown, vpn)
		}
		res.Applied = append(res.Applied, chunks[r.Index].Sections...)
	}
	if err != nil {
		return res, err
	}
	if res.Failed != "" {
		return res, fmt.Errorf("the broker rejected a line in chunk %q and stopped there; "+
			"nothing after it was applied -- see the detail in the report above", res.Failed)
	}

	// Only now is it true that a torn-down VPN was put back. Recorded as a fact
	// rather than inferred by the report, because every early return above leaves
	// res.TornDown populated and this false -- which is exactly the state the
	// operator most needs named.
	res.Rebuilt = true
	res.Skipped = skippedSections(p)
	return res, nil
}

// chunkSectionPrefix marks a chunk whose name is a broker-level section list, so the
// report can tell an applied SECTION from a teardown or a VPN rebuild.
const chunkSectionPrefix = "sections: "

// buildChunks assembles the ordered work an import does, capturing each existing
// VPN's teardown from the target as it goes.
//
// Order is the whole point of chunking, and it is not the artifact's:
//
//  1. every teardown, before anything is rebuilt;
//  2. `Create logging` alone -- it ends the CLI session, so anything sharing its
//     chunk would never run;
//  3. `Create Usernames` alone -- it changes the CLI admin password, and isolating
//     it means a failure there cannot take the rest of the broker config with it;
//  4. the remaining broker-level sections;
//  5. one chunk per message-VPN, `default` FIRST, so a VPN that fails leaves the
//     others applied and names itself in the report.
func (o *Ops) buildChunks(ctx context.Context, role config.Role, p *ImportPlan, blocks []Block, res *ImportResult) ([]chunk, error) {
	var chunks []chunk

	for _, vpn := range p.Existing {
		o.logf("Capturing the teardown for message-VPN %q (its queues and their spooled messages are lost)...", vpn)
		teardown, _, err := o.runCapture(ctx, role, "export-teardown", vpn, true)
		if err != nil {
			return nil, fmt.Errorf("capture the teardown for message-VPN %q: %w", vpn, err)
		}
		// TearsDown, not res.TornDown: a teardown that has been CAPTURED has not been
		// applied. Recording it here meant every failure before or during the driver
		// reported the VPN as emptied and its spooled messages destroyed when nothing
		// had been run against it -- the exact opposite of the truth, in the report
		// the operator reads to decide what to do next.
		chunks = append(chunks, chunk{Name: "teardown " + vpn, Body: string(teardown), TearsDown: vpn})
	}

	// Objects the TARGET carries that the artifact does not re-create, removed before
	// anything is applied. They have no create line to sit beside, so they cannot go
	// through the block-level clear -- see TargetOnlyRemovals.
	if lines := TargetOnlyRemovals(blocks, p.Target); len(lines) > 0 {
		var b strings.Builder
		for _, line := range preambleForApply(p.Capture.Preamble) {
			b.WriteString(line)
			b.WriteString("\n")
		}
		for _, line := range lines {
			b.WriteString(line)
			b.WriteString("\n")
		}
		chunks = append(chunks, chunk{Name: "clear target-only objects", Body: b.String()})
	}

	// The create-VPN pass, BEFORE any broker-level section.
	//
	// This is the artifact's own order -- `! Create All: "Message Vpns"` is the first
	// thing after `configure`, ahead of the broker-level region -- and it is load
	// bearing: a broker-level line that names a message-VPN is rejected if that VPN
	// does not exist yet. Chunking originally folded these blocks into each VPN's own
	// chunk, which runs LAST, so every such reference applied before its VPN was
	// created. The single-script apply this replaced had it right by rendering in
	// artifact order.
	if body := renderRegionScript(blocks, p, RegionVPNCreate, ""); body != "" {
		chunks = append(chunks, chunk{Name: "create message-vpns", Body: body})
	}

	add := func(name string, names []string) {
		body, order := renderScript(blocks, p, names)
		if body == "" {
			return
		}
		chunks = append(chunks, chunk{Name: chunkSectionPrefix + name, Body: body, Sections: order})
	}

	first, rest := splitFirstSections(p)
	add("logging", first)

	var auth, others []string
	for _, n := range rest {
		if n == authSectionName {
			auth = append(auth, n)
			continue
		}
		others = append(others, n)
	}
	add(authSectionName, auth)
	add("broker", others)

	for _, vpn := range defaultVPNFirst(p.Capture.VPNs()) {
		body := renderVPNScript(blocks, p, vpn)
		if body == "" {
			continue
		}
		chunks = append(chunks, chunk{Name: "message-vpn " + vpn, Body: body})
	}
	return chunks, nil
}

// defaultVPNName is the one message-VPN every broker has and none can delete.
const defaultVPNName = "default"

// defaultVPNFirst orders `default` ahead of every other message-VPN, leaving the
// rest in the artifact's own order.
//
// `default` is not an ordinary VPN. It is the one that cannot be deleted, so it is
// EDITED in place rather than recreated (its teardown empties it and leaves it
// standing), and it is where the port changes that need a service taken down first
// actually live -- the case the operator's migration GUI was written around. Doing it
// before the others means the riskiest chunk either succeeds or stops the run while
// every other VPN is still untouched, rather than failing at the end with the broker
// half rebuilt.
func defaultVPNFirst(vpns []string) []string {
	out := make([]string, 0, len(vpns))
	for _, v := range vpns {
		if v == defaultVPNName {
			out = append(out, v)
		}
	}
	for _, v := range vpns {
		if v != defaultVPNName {
			out = append(out, v)
		}
	}
	return out
}

// renderVPNScript renders one message-VPN's own chunk: its entry in the create-VPN
// pass plus every block scoped to it, in artifact order.
//
// Splitting the VPN region per VPN is what makes a failure attributable. Applied as
// one script, a rejected line in the third VPN told an operator only that "the
// import" failed; as chunks, the report names the VPN and every VPN before it is
// known to have landed.
func renderVPNScript(blocks []Block, p *ImportPlan, vpn string) string {
	return renderRegionScript(blocks, p, RegionVPN, vpn)
}

// renderRegionScript renders every block of one region, optionally narrowed to a
// single message-VPN, with the apply preamble.
//
// The two regions are rendered SEPARATELY and in different chunks on purpose:
// RegionVPNCreate creates every VPN empty and has to run before the broker-level
// sections that may reference one, while RegionVPN carries each VPN's own objects
// and runs after. Folding the create pass into the per-VPN chunk put it after
// everything.
func renderRegionScript(blocks []Block, p *ImportPlan, region Region, vpn string) string {
	var b strings.Builder
	for _, line := range preambleForApply(p.Capture.Preamble) {
		b.WriteString(line)
		b.WriteString("\n")
	}
	wrote := false
	for _, blk := range blocks {
		if blk.Region != region || (vpn != "" && blk.VPN != vpn) {
			continue
		}
		for _, line := range blk.Lines {
			b.WriteString(line)
			b.WriteString("\n")
			wrote = true
		}
	}
	if !wrote {
		return ""
	}
	return b.String()
}

// authSectionName is the section isolated into its own chunk because it rewrites the
// CLI admin password (see sections.go's own note on the consequence).
const authSectionName = "Create Usernames"

// skippedSections is the plan's skipped set, for the result report.
func skippedSections(p *ImportPlan) []PlannedSection {
	var out []PlannedSection
	for _, s := range p.Sections {
		if s.Disposition == Skip || s.Disposition == SkipAppliance {
			out = append(out, s)
		}
	}
	return out
}

// runDriver uploads the generated apply script and runs it, in ONE round trip each.
//
// The output is parsed, never shown. A CLI transcript repeats every credential in
// the configuration it applied, which is why success transcripts never leave the
// broker at all -- the driver prints one marker line per chunk, and only a FAILING
// chunk's tail comes back with it.
func (o *Ops) runDriver(ctx context.Context, role config.Role, chunks []chunk) ([]chunkResult, error) {
	body, err := renderDriver(chunks, CLIScriptsDir, CLIBinary)
	if err != nil {
		return nil, err
	}
	dest := shellScriptPath("solace-util-import")
	// Upload, not UploadFile: the body carries the whole configuration, credentials
	// included, and Upload feeds it on stdin so it never appears in argv or a log.
	if err := o.T.Upload(ctx, role, []byte(body), dest); err != nil {
		return nil, fmt.Errorf("upload the import driver: %w", err)
	}
	defer o.removeFiles(ctx, role, dest)

	out, runErr := o.T.Output(ctx, role, "sh", dest)
	results, parseErr := parseDriverOutput(out, chunks)
	if len(results) == 0 && runErr != nil {
		// Nothing reported AND the exec failed: the exec is the cause, and saying so
		// keeps a distinction that matters more than it looks. "The CLI never ran" and
		// "the broker refused a line" lead to completely different next steps, and the
		// parse error below would describe the symptom while hiding the reason.
		return nil, fmt.Errorf("the import driver did not run: %w -- this is a failure to "+
			"execute, not a rejected configuration line", runErr)
	}
	if parseErr != nil {
		// Otherwise the parse failure is the more informative of the two: a non-zero
		// status is the EXPECTED shape of a rejected chunk, so it explains nothing.
		return results, parseErr
	}
	if runErr != nil && !hasFailedChunk(results) {
		// Non-zero with every chunk reported OK means the script itself broke after
		// the last chunk, not that the broker refused anything.
		return results, fmt.Errorf("the import driver exited non-zero after reporting every chunk "+
			"as applied, so its own run failed rather than the configuration: %w", runErr)
	}
	return results, nil
}

func hasFailedChunk(results []chunkResult) bool {
	for _, r := range results {
		if !r.OK {
			return true
		}
	}
	return false
}

// splitFirstSections separates the sections that must be applied alone and first.
func splitFirstSections(p *ImportPlan) (first, rest []string) {
	for _, s := range p.Sections {
		if s.Disposition != Apply && s.Disposition != ApplyFiltered {
			continue
		}
		if s.First {
			first = append(first, s.Name)
		} else {
			rest = append(rest, s.Name)
		}
	}
	return first, rest
}

// renderScript builds one BROKER-LEVEL CLI script from blocks, in artifact order.
//
// Blocks are included only when their section is in names, and are filtered by that
// section's DropLines rule. VPN work is deliberately not here: each message-VPN gets
// its own chunk from renderVPNScript, so a failure names the VPN. The section
// classification is a broker-level regime only -- gating VPN blocks on it once
// dropped every one of them, so an import tore a VPN down and then reapplied
// nothing, which is why the two renderers are now separate functions rather than one
// with a flag.
//
// Order is the artifact's own throughout, which is the broker's dependency order.
// That is why this channel needs none of the object-ordering machinery a SEMP v2
// restore would: the export already solved it.
func renderScript(blocks []Block, p *ImportPlan, names []string) (string, []string) {
	want := map[string]bool{}
	for _, n := range names {
		want[n] = true
	}

	var b strings.Builder
	for _, line := range preambleForApply(p.Capture.Preamble) {
		b.WriteString(line)
		b.WriteString("\n")
	}

	applied := map[string]bool{}
	var order []string
	wrote := false
	for _, blk := range blocks {
		if blk.Region != RegionBroker || !want[blk.Section] {
			continue
		}

		rule, _ := RuleFor(blk.Section)
		for _, line := range blk.Lines {
			if rule.DropLines != nil && rule.DropLines(line) {
				continue
			}
			b.WriteString(line)
			b.WriteString("\n")
			wrote = true
		}
		if !applied[blk.Section] {
			applied[blk.Section] = true
			order = append(order, blk.Section)
		}
	}
	if !wrote {
		return "", nil
	}
	return b.String(), order
}

// ImportVerify re-exports the target and diffs it against the artifact.
//
// Because each conflicting VPN was torn down first, every artifact block must now
// be present and matching: there is no legitimately-differing set to excuse a
// mismatch, which is what makes this a verdict rather than a heuristic.
func (o *Ops) ImportVerify(ctx context.Context, role config.Role, p *ImportPlan) (DiffResult, error) {
	out, _, err := o.runCapture(ctx, role, "verify-export", "", false)
	if err != nil {
		return DiffResult{}, fmt.Errorf("re-export the target for verification: %w", err)
	}
	got, err := ParseBlocks(StripMarkers(out))
	if err != nil {
		return DiffResult{}, fmt.Errorf("parse the verification re-export: %w", err)
	}
	return DiffBlocks(p.Capture.Blocks, got.Blocks, importIgnore(p)), nil
}

// importIgnore tells the diff which lines the import deliberately did not apply,
// so a skipped section does not read as a failure.
func importIgnore(p *ImportPlan) func(Block, string) bool {
	disp := map[string]Disposition{}
	gap := map[string]bool{}
	for _, s := range p.Sections {
		disp[s.Name] = s.Disposition
		gap[s.Name] = s.Unclassified
	}
	return func(b Block, line string) bool {
		// A line the apply deliberately applied the OPPOSITE of was never intended to
		// land, so demanding it on the target is a guaranteed false failure. Only
		// ShutdownBridges does this -- every other transform adds lines, which a
		// DIRECTIONAL diff tolerates. Before this, any artifact containing a bridge
		// failed verification after a perfectly clean apply.
		if BridgeEnablementInverted(b, line) {
			return true
		}
		// An unclassified section is a gap, not a decision: its lines stay in the
		// diff so the import fails loudly instead of dropping configuration this
		// build does not know and calling the result clean. See PlannedSection.
		if gap[b.Section] {
			return false
		}
		switch disp[b.Section] {
		case Skip, SkipAppliance:
			return true
		case ApplyFiltered:
			if rule, ok := RuleFor(b.Section); ok && rule.DropLines != nil {
				return rule.DropLines(line)
			}
		}
		return false
	}
}

// keepRegion filters blocks to one region. It was variadic, and built a set to test
// membership of, for a single call site passing a single region.
func keepRegion(blocks []Block, keep Region) []Block {
	out := make([]Block, 0, len(blocks))
	for _, b := range blocks {
		if b.Region == keep {
			out = append(out, b)
		}
	}
	return out
}

// describeScope renders the scope for the artifact's metadata.
func describeScope(vpns []string, brokerOnly bool) string {
	switch {
	case brokerOnly:
		return "broker"
	case len(vpns) > 0:
		sorted := append([]string(nil), vpns...)
		sort.Strings(sorted)
		return "vpn:" + strings.Join(sorted, ",")
	default:
		return "total"
	}
}

// ExportReport renders the export summary. It is called only when --out freed stdout.
func (o *Ops) ExportReport(r *output.Sink, vpns []string, brokerOnly bool, size int) {
	r.Section("Configuration export")
	r.KVBlock([]output.KV{
		{Key: "scope", Value: describeScope(vpns, brokerOnly)},
		// Stated on every export, because it is why the file needs the same care as the
		// env file: the broker's encrypted forms are what make it replayable.
		{Key: "secrets", Value: "carried in the broker's encrypted form"},
		{Key: "bytes", Value: fmt.Sprintf("%d", size)},
	})
}

// Report renders the import plan, before the confirmation.
func (p *ImportPlan) Report(r *output.Sink) {
	r.Section("Import plan")
	r.KVBlock([]output.KV{
		{Key: "broker type", Value: string(p.Capture.Type)},
		{Key: "captured from", Value: p.Capture.Version},
		{Key: "VPNs torn down", Value: orNone(p.ExistingList())},
		{Key: "VPNs created", Value: orNone(p.NewList())},
		// What the artifact says export left out. Import never sees these sections,
		// so this row is the only place an operator learns they were dropped.
		{Key: "omitted at export", Value: orNone(omittedList(p.Capture.Omitted))},
	})

	rows := make([]output.Row, 0, len(p.Sections))
	for _, s := range p.Sections {
		rows = append(rows, output.Row{Level: sectionLevel(s), Cells: []string{
			s.Name, string(s.Disposition), s.Interrupts, s.Reason,
		}})
	}
	if len(rows) > 0 {
		r.TaggedTable([]string{"SECTION", "IMPORT DOES", "INTERRUPTS", "WHY"}, rows)
	}

	// The warnings belong HERE, in the plan, because this report is what the
	// operator reads before answering the tear-down confirmation. Printing them only
	// from ImportResult put them after every existing VPN had already been
	// destroyed -- an unclassified section this build will not apply, or a service
	// InjectShutdown cannot safely bounce, is exactly the kind of thing someone
	// might say no to, and they were being told once saying no was no longer
	// possible.
	for _, w := range p.Warnings {
		r.Warn("%s", w)
	}
}

func sectionLevel(s PlannedSection) output.Level {
	switch {
	case s.Disposition == Skip || s.Disposition == SkipAppliance:
		return output.LevelSkip
	case s.Interrupts != "":
		return output.LevelWarn
	default:
		return output.LevelOK
	}
}

// Report renders what an apply did.
//
// This is printed on the FAILURE path too (runImport reports the result before
// returning the error), and on that path it is the only statement of what the
// broker is now holding -- so the teardown and the rebuild are reported as two
// separate facts. The row used to read "VPNs torn down and rebuilt" off TornDown
// alone, which asserted the rebuild happened in precisely the case where it had
// not: the teardown loop completes before any rebuild begins, so an apply that
// fails leaves those VPNs emptied, their spooled messages destroyed, and nothing
// put back. That is the one thing the operator cannot be allowed to misread.
func (r *ImportResult) Report(s *output.Sink) {
	s.Section("Import result")
	s.KVBlock([]output.KV{
		{Key: "VPNs torn down", Value: orNone(quoteList(r.TornDown))},
		{Key: "rebuilt from the artifact", Value: yesNoWord(r.Rebuilt)},
		{Key: "sections applied", Value: fmt.Sprintf("%d", len(r.Applied))},
		{Key: "sections skipped", Value: fmt.Sprintf("%d", len(r.Skipped))},
	})
	if r.Failed != "" {
		s.Fail("the broker rejected a line in chunk %q; `stop-on-error` stopped that chunk "+
			"there and no later chunk ran.", r.Failed)
		if r.FailedDetail != "" {
			// The tail the driver echoed back, indented as a block. It is the only
			// part of any transcript that leaves the broker, and it exists because a
			// chunk that failed with no detail is not actionable.
			s.Line("  the broker's last words on that chunk:")
			for _, line := range strings.Split(r.FailedDetail, "\n") {
				s.Line("    %s", line)
			}
		}
	}
	if len(r.TornDown) > 0 && !r.Rebuilt {
		s.Fail("%s was torn down and NOT rebuilt: it is empty, its spooled messages are "+
			"gone, and the artifact was not applied. Re-run the same import to rebuild it.",
			quoteList(r.TornDown))
	}
	for _, w := range r.Warnings {
		s.Warn("%s", w)
	}
}

func yesNoWord(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func orNone(s string) string {
	if s == "" {
		return "none"
	}
	return s
}

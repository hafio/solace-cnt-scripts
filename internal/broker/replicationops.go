package broker

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"solace/internal/config"
)

// replicationops.go is `broker configure data-replication`: it converges THIS broker to
// the env file's replication block.
//
// It is local-only by design (operator-confirmed). It never reads and never writes the
// mate, so it works on every platform, needs no site access block, and cannot be blocked
// by a WAN outage. The accepted consequence, which its help text and docs/operations.md
// both have to state: this is the ONE path in the feature that can produce two actives --
// run it at the new-active site while the old site still holds that VPN active and both
// are active until someone notices. Moving a role safely is the switch's job.
//
// THE APPLY IS TWO PHASES, IN TWO SEPARATE RunCLI CALLS, because the broker enforces a
// precondition the single-script apply this replaced did not respect: mate configuration
// (the address lines and the virtual-router-name) can only be changed while EVERY VPN on
// the broker has replication disabled. CONFIRMED by the operator.
//
// Worth recording that it was NOT derivable from the documentation: neither
// `enable configure replication mate connect-port` nor `... virtual-router-name` states a
// precondition in the CLI reference, even though that reference does spell one out where it
// exists for other services ("AMQP must be disabled to change the port", and likewise for
// the health check and SEMP). The only evidence in the repo was circumstantial -- the
// broker's own replayable dump orders `! Configure Replication` some 32,000 lines ahead of
// the first per-VPN `no shutdown`, so mate-first, VPN-enable-after is the order the broker
// itself writes.
//
// Phase 1 is what makes that order true when converging an already-running broker, rather
// than replaying a fresh capture onto an empty one. It is therefore NECESSARY, and the
// outage window it opens is the real cost of a mate change rather than something a cleverer
// implementation could avoid. Which is also why phase 1 runs ONLY when the mate actually
// differs: that check is the only thing standing between an ordinary re-run and a
// site-wide replication stop.
//
//	PHASE 1 (mate convergence, only when the mate actually differs): disable REPLICATION
//	on every VPN this broker currently reports admin-ENABLED for it, then the removals,
//	then the new mate lines -- one RunCLI call.
//
//	PHASE 2 (per-VPN, always): for each LISTED VPN, in file order, `state <role>` -- and
//	`no shutdown` ONLY where one is needed -- then `shutdown` for every unlisted VPN this
//	broker still has enabled. One RunCLI call, run whether or not phase 1 ran.
//
// A ROLE IS SET IN PLACE, with no shutdown around it, against a VPN that is up and
// replicating (operator-confirmed): `message-vpn <n>` -> `replication` ->
// `state <active|standby>` is the whole of it. An earlier draft of phase 2 cycled every
// listed VPN down and back up to change its role, on the assumption that a role could
// only move while replication was disabled. That assumption was wrong, and the cost was
// real: it interrupted replication on every listed VPN on every run, including the runs
// where nothing about the mate had changed and nothing needed to stop at all.
//
// The evidence was in the tree the whole time. `setReplicationRoleScript` (scripts.go),
// which `perform dr`'s switchover calls to move a role, sets `state` alone against a VPN
// its own preflight requires to be ENABLED -- so the two paths contradicted each other,
// and the switchover was the one telling the truth. They now agree.
//
// Phase 1 is therefore the ONLY thing that stops replication, and it runs only when the
// mate configuration actually differs. Phase 2 re-enables what phase 1 took down.
//
// A REJECTION IS NEVER RETROACTIVE. `RunCLI` wraps each phase in the broker's own
// `source script ... stop-on-error no-prompt`, so a rejected line stops THAT call; it does
// not undo a call that already succeeded. A phase-1 failure leaves every VPN it shut down
// still down, with phase 2 never sent. A phase-2 failure after a successful phase 1 leaves
// the mate already converged and some prefix of the listed VPNs re-enabled, with the rest
// (and every unlisted-but-replicating VPN) still down from phase 1. Neither failure is
// rolled back, because there is nothing here that could safely reconstruct "what the broker
// held a moment ago" -- the error text says which phase stopped and what state that leaves
// replication in, and re-running is SAFE because the whole script is recomputed from a
// fresh read of the broker rather than replayed -- CONVERGENT, which is the accurate word;
// the re-enable keyword in particular is state-dependent and so is not idempotent in the
// literal sense. It is SUFFICIENT only when the cause was transient: a line the broker
// refuses for a reason in the environment is refused again at the same place.

// ReplicationConfigResult is what one convergence did, for the report.
type ReplicationConfigResult struct {
	Removed []string // mate addresses cleared (rendered; see MateApplied for whether sent)
	Mate    []string // mate address lines rendered (see MateApplied for whether sent)

	// MateApplied is whether phase 1 actually ran. It is false when this broker's mate
	// configuration already matched the env file, in which case Removed and Mate above
	// are what WOULD have been sent, not what was: nothing was written and no VPN was
	// stopped for it.
	MateApplied bool
	// StoppedForMate names every VPN phase 1 disabled REPLICATION on to satisfy the
	// broker's mate-change precondition, sorted. The VPNs themselves are untouched and
	// their clients stay connected -- every line either phase sends is inside the VPN's
	// own `replication` node. Empty when MateApplied is false. Phase 2 turns
	// the listed ones back on and leaves the rest down -- which is the same set its own
	// unlisted pass would have shut down anyway, because both phases key on the same
	// predicate (admin-enabled) precisely so no VPN falls between them.
	StoppedForMate []string

	Enabled  []string            // listed VPNs that were not already admin-enabled
	Disabled []string            // replicating VPNs the file does not list
	Roles    map[string]ReplRole // per listed VPN, the role applied
	Status   string              // the closing `show` output, for display
}

// ConfigureReplication applies the block to this broker in the two phases described in
// this file's own doc comment.
//
// self and mate are the two site entries, already resolved by the caller -- which one
// this broker IS comes from its own router name, and resolving it is the CLI's job so
// that this function stays testable without a broker.
func (o *Ops) ConfigureReplication(ctx context.Context, role config.Role,
	self, mate config.ReplSite, vpns []config.ReplVPN) (*ReplicationConfigResult, error) {
	// readMateConfig makes the round trip anyway (`show replication`), and its
	// response carries the login banner every `cli -Apes` invocation echoes ahead
	// of its own output -- so the broker type comes off THAT read rather than a
	// separate dedicated banner call (the removed Ops.brokerType). That is why this
	// read now runs before RenderMate needs the type, not after.
	have, typ, err := o.readMateConfig(ctx, role)
	if err != nil {
		return nil, err
	}

	want := MateConfig{VirtualRouterName: mate.VirtualRouterName, Endpoints: mate.Endpoints}
	// Rendering FIRST means an unrenderable set -- a software HA pair's two hostnames
	// against an appliance, say -- is refused before anything is written.
	mateLines, err := RenderMate(typ, want)
	if err != nil {
		return nil, err
	}
	// Removals make the apply CONVERGENT rather than additive. Without them, pointing a
	// broker at a new mate leaves the old addresses in place: a second run is not a
	// no-op, and the broker keeps dialling somewhere nobody named.
	removals, err := RenderMateRemovals(typ, have, want)
	if err != nil {
		return nil, err
	}

	state, err := o.readVPNReplication(ctx, role)
	if err != nil {
		return nil, err
	}

	// Refuse before any write, the same principle RenderMate/RenderMateRemovals already
	// follow above. Phase 2 writes `message-vpn <name>` for a listed VPN, which the
	// broker's own grammar treats as a MODIFY (`[create | no] message-vpn <vpn-name>` --
	// bare `message-vpn` is neither), so a name this broker does not have is rejected --
	// and under RunCLI's stop-on-error wrapper that abandons the whole phase-2 chunk,
	// leaving every listed VPN AFTER it in file order shut down by phase 1 and never
	// re-enabled. Checking here costs nothing: readVPNReplication already read every
	// VPN the broker has.
	// Every listed name reaches the broker CLI as a quoted operand inside an already
	// elevated session, so it goes through the same boundary guard the import path
	// applies (validVPNName, broker.go) rather than trusting the env file. `%q` is NOT
	// that guard: it escapes a double quote as \" , which the Solace CLI does not read
	// as an escape, so a name carrying one would close the operand and run the rest as
	// commands. Config validation checks these names for emptiness and duplication only.
	//
	// `#config-sync` is refused by name here too. It is in the broker's own report -- so
	// the presence check below finds it and waves it through -- but it is the broker's
	// internal VPN, not a configurable object, and phase 2 would emit a `message-vpn`
	// line for a node `configure` cannot enter. Under stop-on-error that abandons the
	// whole chunk, leaving every listed VPN after it shut down by phase 1.
	for _, v := range vpns {
		if err := validVPNName(v.Name); err != nil {
			return nil, fmt.Errorf("replication.vpns: %w; nothing was written", err)
		}
		if v.Name == internalConfigSyncVPN {
			return nil, fmt.Errorf("replication.vpns names %q, which is the broker's own internal "+
				"config-sync VPN rather than a message-VPN you configure; it cannot be given a "+
				"replication role and the broker would refuse the line. Remove it from the env "+
				"file. Nothing was written", v.Name)
		}
	}

	if missing := missingListedVPNs(vpns, state); len(missing) > 0 {
		return nil, fmt.Errorf("replication.vpns names %s, which this broker's own "+
			"`show message-vpn * replication` does not report; nothing was written. "+
			"Fix the name (or create the VPN on this broker) and run this command again",
			strings.Join(missing, ", "))
	}

	// The mate decision comes FIRST now, because the per-VPN body depends on it: a
	// listed VPN needs re-enabling only when phase 1 is about to take it down.
	//
	// The mate only needs converging when it actually differs. Built from SameEndpoints
	// and the router name rather than from "removals is empty": a broker holding a
	// SUBSET of the wanted endpoints renders zero removals (RenderMateRemovals only ever
	// clears what `have` holds and `want` does not) and would otherwise be read as
	// already matching, silently dropping the missing address forever. SameEndpoints'
	// own plain-text normalisation is why this is not a raw line comparison either: this
	// broker omits the `plain-text` keyword on render (renderMateSoftware) but a real
	// broker may echo it back, and a keyword-sensitive compare would call that drift on
	// every plain-text endpoint, on every run, forever.
	mateApplied := have.VirtualRouterName != want.VirtualRouterName || !SameEndpoints(have.Endpoints, want.Endpoints)

	vpnLines, res := replicationVPNLines(vpns, self, state, mateApplied)
	res.Removed, res.Mate, res.MateApplied = removals, mateLines, mateApplied

	if res.MateApplied {
		// Iterated in sorted order for the same reason replicationVPNLines' own map
		// range is: a Go map range is randomised, and a generated script that differed
		// run to run would make a diff of two applies unreadable.
		res.StoppedForMate = mateConvergenceShutdowns(state)
		if err := o.runMatePhase(ctx, role, res.StoppedForMate, removals, mateLines); err != nil {
			return nil, err
		}
	}

	// Phase 2 always runs UNLESS there is nothing for it to say: no listed VPN and
	// nothing replicating that the file does not name.
	if len(vpnLines) > 0 {
		listedNames := make([]string, 0, len(vpns))
		for _, v := range vpns {
			listedNames = append(listedNames, v.Name)
		}
		if err := o.runVPNPhase(ctx, role, vpnLines, res.MateApplied, res.StoppedForMate, listedNames); err != nil {
			return nil, err
		}
	}

	// The closing shows run only after a CLEAN apply, so a failure report is not buried
	// under a page of broker output.
	//
	// Their failure is NOT this command's failure. They are display only -- the two write
	// phases have both already succeeded by here -- and res is the only record of what
	// those phases did. Returning the error instead would throw that record away and tell
	// an operator whose configuration landed perfectly that the command failed, which is
	// the worst possible reading: the honest answer is that everything was applied and
	// only the closing picture could not be fetched.
	if res.Status, err = o.replicationStatus(ctx, role); err != nil {
		o.progress().Warn("both phases applied cleanly, but the closing %s could not be read (%v) -- "+
			"the report below is what this command did, not what the broker now says",
			showReplicationHint, err)
	}
	return res, nil
}

// runMatePhase is phase 1: the shutdown prefix for every currently-enabled VPN, then the
// removals, then the new mate lines, all in ONE RunCLI call. Split out of
// ConfigureReplication so the two phases and their two failure shapes are each readable
// on their own.
func (o *Ops) runMatePhase(ctx context.Context, role config.Role, stopped, removals, mateLines []string) error {
	var lines []string
	for _, name := range stopped {
		lines = append(lines, vpnReplicationBlock(name, "shutdown")...)
	}
	lines = append(lines, removals...)
	lines = append(lines, mateLines...)
	script := cliHome + "enable\nconfigure\n" + strings.Join(lines, "\n") + newlineIf(lines)

	// The transcript is discarded here: RunCLI shows it itself when it detects a
	// rejection, which is what the error's "see the output above" refers to.
	_, err := o.RunCLI(ctx, role, scriptConfigureReplMate, script)
	if err != nil {
		if isRunCLIRejection(err) {
			return replPhase1Rejected(err, stopped)
		}
		return replTransportFailure(1, "mate convergence", err)
	}
	return nil
}

// runVPNPhase is phase 2: the per-VPN body replicationVPNLines rendered, in ONE RunCLI
// call. mateApplied, stopped and listed are threaded through only to word a rejection's
// error correctly -- an operator reading it has to be told WHICH VPNs are left down, and
// those names are known here and nowhere after (see replPhase2Rejected).
func (o *Ops) runVPNPhase(ctx context.Context, role config.Role, vpnLines []string,
	mateApplied bool, stopped, listed []string) error {
	script := cliHome + "enable\nconfigure\n" + strings.Join(vpnLines, "\n") + newlineIf(vpnLines)
	_, err := o.RunCLI(ctx, role, scriptConfigureReplVPNs, script)
	if err != nil {
		if isRunCLIRejection(err) {
			return replPhase2Rejected(err, mateApplied, stopped, listed)
		}
		return replTransportFailure(2, "per-VPN replication state", err)
	}
	return nil
}

const (
	scriptConfigureReplMate = "repl-configure-mate"
	scriptConfigureReplVPNs = "repl-configure-vpns"
	scriptReplStatus        = "repl-status"
	scriptRouterName        = "repl-router-name"
)

// showReplicationHint is what every phase failure below points the operator at: the two
// reports that, together, show exactly what landed -- the mate half and the per-VPN half.
const showReplicationHint = "`show replication` and `show message-vpn * replication`"

// isRunCLIRejection tells a rejected configuration line apart from a failure to reach the
// broker at all. The two need different words, because the honest statement differs -- a
// rejection means the broker stopped at that line, so what landed is exactly the lines
// before it; a transport failure means the sourced script kept running INSIDE the broker
// after this process lost the connection, so how much of the phase landed is not known at
// all. This is the same split the import driver's own wording draws (importops.go: "this
// is a failure to execute, not a rejected configuration line").
//
// It asks broker.go's ErrCLIRejected sentinel rather than matching a phrase in RunCLI's
// message, which is what it did first: rewording that sentence would have silently turned
// every rejection into an "unreachable" report, downgrading an error that names what
// landed into one that says it cannot be known, with no test to catch it.
func isRunCLIRejection(err error) bool {
	return errors.Is(err, ErrCLIRejected)
}

// Neither builder below embeds a transcript tail: RunCLI shows the transcript itself
// where it detects a rejection, so a tail here would print the same lines twice.

// replPhase1Rejected reports a rejected mate-convergence chunk. Phase 2 is NEVER sent in
// this case (ConfigureReplication returns before reaching it), and nothing from phase 1
// is rolled back: whatever it shut down to satisfy the broker's precondition stays down.
func replPhase1Rejected(err error, stopped []string) error {
	return fmt.Errorf("phase 1 (mate convergence) was rejected by the broker: %w\n\n"+
		"Nothing is rolled back, and phase 2 (the per-VPN state) was never sent. Phase 1 "+
		"sends, in order: a replication shutdown for each listed VPN, the mate removals, "+
		"then the new mate address lines -- so WHERE it stopped decides what is true now, "+
		"and only the broker can say. If it stopped after the removals, this broker holds "+
		"NO mate configuration at all.\n\n"+
		"These have replication disabled and will NOT have it turned back on by this run: "+
		"%s. The VPNs themselves are untouched and their clients are still connected.\n\n"+
		"Run %s to see what landed, fix what the broker refused, then run this again.",
		err,
		replVPNList(stopped, "(none -- this broker had replication enabled on no VPN)"),
		showReplicationHint)
}

// replVPNList renders a VPN name list for an error. Quoted, because a real name is
// `A VPN WITH LONG NAME AND SPACES` and an unquoted run of those cannot be read apart;
// and never rendered as an empty string, which would leave a sentence dangling at the
// exact moment an operator is reading it carefully. The caller supplies the words for the
// empty case, because "none" means something different in each sentence that uses this.
func replVPNList(names []string, ifEmpty string) string {
	if len(names) == 0 {
		return ifEmpty
	}
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = fmt.Sprintf("%q", n)
	}
	return strings.Join(out, ", ")
}

// replPhase2Rejected reports a rejected per-VPN chunk. mateApplied says whether phase 1
// ran first, which changes what is true about the broker's current state: if it ran, the
// mate is already converged and every VPN it shut down for that is NOT rolled back; if it
// was skipped (the mate already matched), nothing has touched the mate at all and only
// the per-VPN chunk itself is in an unknown state.
func replPhase2Rejected(err error, mateApplied bool, stopped, listed []string) error {
	lead := "Phase 1 (mate convergence) was skipped because the mate configuration already " +
		"matched the env file, so it shut nothing down."
	if mateApplied {
		lead = fmt.Sprintf("Phase 1 (mate convergence) had already been applied and is NOT "+
			"rolled back: to write the mate lines it shut down replication on %s, and phase 2 "+
			"had not turned all of them back on when it stopped.",
			replVPNList(stopped, "no VPN, because none had replication enabled"))
	}
	// Says REPLICATION throughout, never "the VPN": everything either phase sends is
	// inside the VPN's `replication` node, so a VPN is never itself shut down and no
	// client connection is dropped. An operator reading this is deciding how urgent the
	// situation is, and "your VPNs are down" is a much worse thing to believe than "the
	// mate feed stopped".
	//
	// With no listed VPNs, phase 2 only disables replication for unlisted ones, so saying
	// it "applies the listed VPNs" would describe work it never had.
	body := fmt.Sprintf("Phase 2 sets the replication role of each listed VPN in place -- %s -- "+
		"enabling replication only where it was off; after them it disables replication for "+
		"every VPN the file does not list.", replVPNList(listed, ""))
	if len(listed) == 0 {
		body = "Phase 2 had no listed VPNs to apply, so all it sends is a replication " +
			"shutdown for each VPN the env file does not name."
	}
	return fmt.Errorf("phase 2 (per-VPN replication state) was rejected by the broker: %w\n\n"+
		"%s\n\n%s The VPN it stopped on and every one after it are left with replication "+
		"shut down. Nothing is rolled back. Run %s to see which.\n\n"+
		"Fix what the broker refused, then run this again. The mate now matches, so the "+
		"re-run skips phase 1 and shuts nothing down for it.",
		err, lead, body, showReplicationHint)
}

// replTransportFailure reports a failure to reach the broker at all -- a killed
// connection, a deleted pod, a cancelled context -- as distinct from a rejected line. The
// sourced script runs INSIDE the broker once started, so it may have finished, partly
// applied, or not started at all; nothing here can tell which, because nothing read the
// broker back after the connection died.
func replTransportFailure(phase int, name string, err error) error {
	return fmt.Errorf("phase %d (%s) could not be run to completion: %w. This is a failure "+
		"to REACH the broker, not a rejected configuration line, so how much of the phase "+
		"applied is NOT known: the uploaded script keeps running inside the broker after "+
		"this process loses the connection. Replication may be shut down on any of this "+
		"broker's VPNs right now. Read %s.",
		phase, name, err, showReplicationHint)
}

// routerNameRE reads the router name out of a capture header.
//
// Every `show current-config` prints it, on both platforms and in both generations:
// `!   Router: "xps-ps-01"` (semp/all.cli:10, software) and `!   Router: "tuas9csol1"`
// (semp/show-currentconfig-all.out:10, appliance). The quoting is what makes it
// unambiguous -- a router name is operator-chosen and the capture quotes it, so nothing
// has to guess where it ends.
var routerNameRE = regexp.MustCompile(`(?m)^!\s*Router:\s*"([^"]*)"`)

// RouterName reads this broker's own router name, which is how it finds itself in
// replication.sites.
//
// It reads a capture HEADER rather than `show router-name`, because the header's format
// is proven on both platforms and in both header generations, and inventing a parser for
// output nobody here has seen is exactly the guess that made the `show replication`
// layout and the VPN flag table wrong twice. The scope is the `default` VPN -- present on
// every broker and far smaller than `all` -- so the cost is one bounded read.
//
// If a `show router-name` capture turns up later, this is the one function to change: the
// callers want a name, not a mechanism.
func (o *Ops) RouterName(ctx context.Context, role config.Role) (string, error) {
	out, err := o.readCLI(ctx, role, scriptRouterName, currentConfigScript("default", false))
	if err != nil {
		return "", err
	}
	m := routerNameRE.FindSubmatch(out)
	if m == nil {
		return "", fmt.Errorf("could not read this broker's router name: no `! Router:` line in the " +
			"capture header. That line is what identifies which replication site this broker is")
	}
	name := strings.TrimSpace(string(m[1]))
	if name == "" {
		return "", fmt.Errorf("this broker reports an empty router name, so it cannot be matched " +
			"against replication.sites[].routerNames")
	}
	return name, nil
}

// newlineIf terminates a joined block, or contributes nothing when it is empty.
func newlineIf(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return "\n"
}

// readMateConfig reads what this broker currently believes its mate is, and the
// broker type read off the same round trip's own login banner (see bannerType) --
// ConfigureReplication used to spend a separate dedicated banner call on that
// (the removed Ops.brokerType); now it is free.
func (o *Ops) readMateConfig(ctx context.Context, role config.Role) (MateConfig, BrokerType, error) {
	out, err := o.readCLI(ctx, role, scriptShowReplication, showReplicationScript())
	if err != nil {
		return MateConfig{}, BrokerUnknown, err
	}
	typ := bannerType(out)
	if typ == BrokerUnknown {
		return MateConfig{}, BrokerUnknown, fmt.Errorf("cannot tell whether this broker is a software " +
			"broker or an appliance: its CLI login banner named neither, and the mate-address grammar " +
			"and the `show replication` layout both differ between the two")
	}
	m, err := ParseShowReplication(typ, out)
	if err != nil {
		return MateConfig{}, typ, err
	}
	return m, typ, nil
}

// readVPNReplication reads EVERY VPN's replication state, not only the listed ones:
// shutting replication down on a VPN the file does not list requires knowing which
// unlisted VPNs are replicating in the first place, and phase 1's own mate-convergence
// shutdown needs to know every VPN that is admin-enabled, listed or not.
func (o *Ops) readVPNReplication(ctx context.Context, role config.Role) (map[string]VPNRepl, error) {
	out, err := o.readCLI(ctx, role, scriptShowVPN, showVPNReplicationScript())
	if err != nil {
		return nil, err
	}
	return ParseVPNReplication(out)
}

// replicationStatus runs the two closing reports and returns them verbatim.
func (o *Ops) replicationStatus(ctx context.Context, role config.Role) (string, error) {
	mate, err := o.readCLI(ctx, role, scriptReplStatus, showReplicationScript())
	if err != nil {
		return "", err
	}
	vpns, err := o.readCLI(ctx, role, scriptShowVPN, showVPNReplicationScript())
	if err != nil {
		return "", err
	}
	return string(mate) + "\n" + string(vpns), nil
}

// RoleAtSite is the role one VPN takes at one site, derived from the file alone.
//
// This is the whole of why the block can be byte-identical at both ends: the file names
// which site owns active, and each broker computes the COMPLEMENT for itself. Neither
// file states a role, so there is nothing to reverse when copying it across -- and two
// actives is not expressible, because one sentence cannot name two sites.
func RoleAtSite(v config.ReplVPN, site config.ReplSite) ReplRole {
	if v.ActiveAt == site.VirtualRouterName {
		return RoleActive
	}
	return RoleStandby
}

// PlannedRoles is what the file asks of THIS site, which the caller needs before the
// apply so the confirmation can name each VPN and the move it will make.
func PlannedRoles(vpns []config.ReplVPN, self config.ReplSite) map[string]ReplRole {
	out := make(map[string]ReplRole, len(vpns))
	for _, v := range vpns {
		out[v.Name] = RoleAtSite(v, self)
	}
	return out
}

// missingListedVPNs names every replication.vpns entry that state -- this broker's own
// `show message-vpn * replication`, covering every VPN it has -- does not report. Order
// follows the file, since this is for an error message rather than a generated script.
func missingListedVPNs(vpns []config.ReplVPN, state map[string]VPNRepl) []string {
	var missing []string
	for _, v := range vpns {
		if _, ok := state[v.Name]; !ok {
			missing = append(missing, v.Name)
		}
	}
	return missing
}

// mateConvergenceShutdowns names, sorted, every VPN this broker currently reports
// admin-ENABLED for replication -- the set phase 1 must shut down before it can touch the
// mate configuration.
//
// Keyed on ADMIN STATE ALONE, not Replicating() (which also requires a resolved role):
// the broker's precondition is about replication being enabled, not about which role a
// VPN currently holds, and a VPN reading enabled with role "-" would otherwise be left up
// by this predicate and refuse the mate lines that follow in the SAME chunk -- the exact
// failure this phase exists to prevent. `#config-sync`, the internal VPN every broker
// carries, is excluded either way: it reports "-" in the admin column too (real capture,
// semp/show-message-vpn-replication.out), so `Admin == AdminEnabled` is false for it, and
// it is not a configurable object at all -- `configure` cannot enter its node.
//
// Sorted for the same reason replicationVPNLines' own map range is: a Go map range is
// randomised, and a generated script that differed run to run would make a diff of two
// applies unreadable.
func mateConvergenceShutdowns(state map[string]VPNRepl) []string {
	var out []string
	for name, st := range state {
		if st.Admin == AdminEnabled {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

// replicationVPNLines renders phase 2's body and records what it decided.
//
// A ROLE IS SET IN PLACE. `message-vpn <n>` -> `replication` -> `state <role>` is all it
// takes, against a VPN that is up and replicating, with no shutdown around it
// (operator-confirmed). `setReplicationRoleScript` does the same for `perform dr`'s
// switchover, so the two paths agree. Cycling a VPN down and back up to change its
// role would interrupt replication on every run where the mate had not changed.
//
// So a listed VPN is NEVER shut down here. It gets `state <role>`, and `no shutdown` only
// when it actually needs enabling, which is either of two cases:
//
//   - it was not admin-enabled when this run read the broker, or
//   - phase 1 ran and shut it down to converge the mate (mateStopped), which is the one
//     situation where this command is responsible for it being down.
//
// The role is written BEFORE the enable in that case, deliberately: a VPN that comes up
// holding the wrong role, even briefly, is worse than one that comes up a moment later.
//
// UNLISTED-BUT-ENABLED means shut down -- the file is authoritative, so a VPN replicating
// on the broker that the block does not name stops. That is destructive to message flow,
// which is why the command takes an exact-`yes` gate naming those VPNs before any of it
// runs. When phase 1 already stopped it, the line is not repeated: it is recorded in the
// report as disabled, because it is, but nothing re-sends a shutdown to a VPN that is
// already down.
//
// The enable line's KEYWORD comes from the QUEUE column, never from enablement -- see
// reenableLine, and the real capture that settles why the two are different facts.
func replicationVPNLines(vpns []config.ReplVPN, self config.ReplSite,
	state map[string]VPNRepl, mateStopped bool) ([]string, *ReplicationConfigResult) {
	res := &ReplicationConfigResult{Roles: map[string]ReplRole{}}
	listed := make(map[string]bool, len(vpns))
	var lines []string

	for _, v := range vpns {
		listed[v.Name] = true
		role := RoleAtSite(v, self)
		res.Roles[v.Name] = role
		cur := state[v.Name]

		// The role always. No shutdown around it -- see this function's doc comment.
		cmds := []string{"state " + string(role)}

		// The enable only when it is actually needed. mateStopped is the case this
		// command created itself: phase 1 took every enabled VPN down to converge the
		// mate, so a listed one has to be brought back even though it was up when the
		// run started.
		if wasDown := cur.Admin != AdminEnabled; wasDown || mateStopped {
			if wasDown {
				res.Enabled = append(res.Enabled, v.Name)
			}
			cmds = append(cmds, reenableLine(cur))
		}
		lines = append(lines, vpnReplicationBlock(v.Name, cmds...)...)
	}

	// Iterated in sorted order: a map range is randomised, and a generated script that
	// differed run to run would make a diff of two applies unreadable.
	//
	// The predicate is admin state ALONE, deliberately the same one phase 1 shuts down
	// on (mateConvergenceShutdowns) and NOT Replicating(), which additionally demands a
	// resolved role. They have to be the same set or a VPN falls between them: an
	// unlisted VPN that is admin-enabled with an unresolved role was shut down by phase
	// 1, and under the narrower predicate phase 2 would neither re-enable it nor record
	// it -- so it ended up stopped or running depending on whether the MATE happened to
	// differ, which has nothing to do with that VPN.
	for _, name := range sortedVPNs(state) {
		if listed[name] || state[name].Admin != AdminEnabled {
			continue
		}
		// Reported as disabled either way, because it is. The LINE is skipped when
		// phase 1 already stopped it: re-sending a shutdown to a VPN that is already
		// down adds nothing and gives the broker a second chance to refuse it.
		res.Disabled = append(res.Disabled, name)
		if !mateStopped {
			lines = append(lines, vpnReplicationBlock(name, "shutdown")...)
		}
	}
	sort.Strings(res.Enabled)
	return lines, res
}

// reenableLine renders phase 2's `no shutdown`, choosing the queue-guard keyword from
// the VPN's QUEUE column rather than from whether it happened to be enabled.
//
// The broker offers three forms and the default is `fail-on-existing-queue`, which is
// refused when a data replication queue is already there; `force-use-existing-queue` is
// its inverse and is refused when there is NOT one (semp/appliance_cli_reference.html).
// So this is a two-sided choice with no safe constant answer, and the queue column is the
// only direct evidence of which side we are on.
//
// Reading enablement instead was wrong in BOTH directions, which is why this reads Q.
// semp/show-message-vpn-replication.out has `vpn-01  U S U - - - Y A`: admin-UP with no
// queue, so keying on enablement asks a VPN with nothing to reuse to reuse its queue.
// And phase 1 shuts VPNs down, so on the re-run this command's own errors tell the
// operator to make, a VPN that still HAS its queue reads back admin-DOWN and would get
// the bare form the queue makes illegal -- the documented recovery failing on the
// mechanism the command itself created.
//
// `force-recreate-queue` is never emitted. It discards whatever the queue holds, and
// silently dropping messages a broker was still holding is not a call this tool makes on
// an operator's behalf.
//
// NEEDS VERIFICATION ON A LIVE BROKER: that the queue's existence survives phase 2's own
// `shutdown` (assumed, and the reason this reads state captured BEFORE the script runs),
// and what the column reports for a VPN whose role phase 2 is about to flip -- on the one
// capture here the queue tracks the role, so a promotion may need the form the pre-state
// does not predict. QueueUnknown falls back to the broker's own default rather than
// guessing.
func reenableLine(cur VPNRepl) string {
	if cur.Queue.Exists() {
		return "no shutdown force-use-existing-queue"
	}
	return "no shutdown"
}

func sortedVPNs(m map[string]VPNRepl) []string { return sortedKeys(m) }

// vpnReplicationBlock wraps commands in one VPN's replication node. The name is quoted
// because real VPN names contain spaces; the indentation mirrors a capture's own, which
// is what makes a generated script readable beside one.
func vpnReplicationBlock(vpn string, cmds ...string) []string {
	lines := []string{fmt.Sprintf("message-vpn %q", vpn), "  replication"}
	for _, c := range cmds {
		lines = append(lines, "    "+c)
	}
	return append(lines, "    exit", "  exit")
}

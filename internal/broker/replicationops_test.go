package broker

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"

	"solace/internal/config"
)

// replicationops_test.go covers the local convergence's decision half -- which lines it
// renders, and what it records -- without a broker.

func siteAEntry() config.ReplSite {
	return config.ReplSite{VirtualRouterName: siteA, RouterNames: []string{"sol-a1"}}
}

// TestRoleAtSiteIsTheComplement proves the one sentence that lets a single block be
// correct at both ends: the file names which site owns active, and each broker computes
// the complement for itself.
//
// Neither file states a role, so there is nothing to reverse when the block is copied
// across -- and two actives is not expressible, because one `activeAt` cannot name two
// sites.
func TestRoleAtSiteIsTheComplement(t *testing.T) {
	v := config.ReplVPN{Name: "ORDERS", ActiveAt: siteA}
	if got := RoleAtSite(v, config.ReplSite{VirtualRouterName: siteA}); got != RoleActive {
		t.Errorf("at the named site = %q, want active", got)
	}
	if got := RoleAtSite(v, config.ReplSite{VirtualRouterName: siteB}); got != RoleStandby {
		t.Errorf("at the other site = %q, want standby", got)
	}
	// The SAME block read at both sites yields exactly one active.
	actives := 0
	for _, s := range []string{siteA, siteB} {
		if RoleAtSite(v, config.ReplSite{VirtualRouterName: s}) == RoleActive {
			actives++
		}
	}
	if actives != 1 {
		t.Errorf("one block produced %d actives across the pair, want exactly 1", actives)
	}
}

// TestReplicationVPNLinesEnablesListedAndDisablesTherest is the authoritative-file rule.
//
// Listed means enabled with the derived role. Unlisted BUT REPLICATING means shut down:
// the block is the whole intent, so a VPN replicating on the broker that it does not name
// stops. A VPN that is merely present and not replicating is left alone -- shutting down
// something already down would be noise, and this command is destructive enough already.
func TestReplicationVPNLinesEnablesListedAndDisablesTheRest(t *testing.T) {
	state := map[string]VPNRepl{
		"ORDERS":   {Admin: AdminShutdown, Role: RoleStandby}, // listed, needs enabling
		"PAYMENTS": {Admin: AdminEnabled, Role: RoleStandby},  // listed, already enabled
		"LEGACY":   {Admin: AdminEnabled, Role: RoleActive},   // NOT listed and replicating -> shut down
		"QUIET":    {Admin: AdminShutdown, Role: RoleStandby}, // NOT listed, not replicating -> untouched
		"OTHER":    {Admin: AdminNA, Role: RoleNA},            // NOT listed, does not replicate -> untouched
	}
	vpns := []config.ReplVPN{
		{Name: "ORDERS", ActiveAt: siteA},
		{Name: "PAYMENTS", ActiveAt: siteB},
	}
	lines, res := replicationVPNLines(vpns, siteAEntry(), state, false)
	body := strings.Join(lines, "\n")

	if res.Roles["ORDERS"] != RoleActive || res.Roles["PAYMENTS"] != RoleStandby {
		t.Errorf("roles = %+v, want ORDERS active here and PAYMENTS standby", res.Roles)
	}
	if !strings.Contains(body, "state active") || !strings.Contains(body, "state standby") {
		t.Errorf("both roles must be written:\n%s", body)
	}
	if strings.Join(res.Enabled, ",") != "ORDERS" {
		t.Errorf("Enabled = %v, want only the one that was shut down", res.Enabled)
	}
	if strings.Join(res.Disabled, ",") != "LEGACY" {
		t.Errorf("Disabled = %v, want only the unlisted VPN that was replicating", res.Disabled)
	}
	for _, untouched := range []string{"QUIET", "OTHER"} {
		if strings.Contains(body, untouched) {
			t.Errorf("%s is not replicating and must be left alone:\n%s", untouched, body)
		}
	}
}

// TestReplicationVPNLinesQuotesNamesWithSpaces proves a real VPN name survives. The live
// capture carries `A VPN WITH LONG NAME AND SPACES`, and an unquoted operand would name
// a different VPN -- or none.
func TestReplicationVPNLinesQuotesNamesWithSpaces(t *testing.T) {
	lines, _ := replicationVPNLines(
		[]config.ReplVPN{{Name: "A VPN WITH SPACES", ActiveAt: siteA}},
		siteAEntry(), map[string]VPNRepl{}, false)
	body := strings.Join(lines, "\n")
	if !strings.Contains(body, `message-vpn "A VPN WITH SPACES"`) {
		t.Errorf("the VPN name must be quoted:\n%s", body)
	}
}

// TestReplicationVPNLinesIsDeterministic proves two runs against the same broker state
// render byte-identical scripts.
//
// The shutdown set comes from a MAP, and a Go map range is randomised -- so without an
// explicit sort a diff of two applies would be unreadable, and a golden would flap.
func TestReplicationVPNLinesIsDeterministic(t *testing.T) {
	state := map[string]VPNRepl{
		"AAA": {Admin: AdminEnabled, Role: RoleActive},
		"BBB": {Admin: AdminEnabled, Role: RoleActive},
		"CCC": {Admin: AdminEnabled, Role: RoleStandby},
		"DDD": {Admin: AdminEnabled, Role: RoleStandby},
	}
	first, res := replicationVPNLines(nil, siteAEntry(), state, false)
	for i := 0; i < 20; i++ {
		again, _ := replicationVPNLines(nil, siteAEntry(), state, false)
		if strings.Join(again, "\n") != strings.Join(first, "\n") {
			t.Fatalf("run %d rendered a different script:\n%s\n---\n%s",
				i, strings.Join(first, "\n"), strings.Join(again, "\n"))
		}
	}
	want := []string{"AAA", "BBB", "CCC", "DDD"}
	got := append([]string(nil), res.Disabled...)
	sort.Strings(got)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("Disabled = %v, want every replicating VPN when none is listed", res.Disabled)
	}
}

// TestReplicationVPNLinesSetsRoleWithoutCyclingReplication pins the rule that replaced a
// wrong one: a role is set IN PLACE, with no shutdown around it, against a VPN that is up
// and replicating (operator, 2026-09-14).
//
// The earlier draft cycled every listed VPN down and back up to change its role. That
// interrupted replication on each one on EVERY run, including the ones where nothing about
// the mate had changed and nothing needed to stop -- which is every re-run after the first.
// The contradiction was visible in the tree the whole time: setReplicationRoleScript, which
// perform dr's switchover uses, already set `state` alone against a VPN its own preflight
// requires to be ENABLED.
//
// The enable is emitted only where one is actually needed, and its KEYWORD comes from the
// QUEUE column rather than from enablement. The broker's default is fail-on-existing-queue,
// refused when a queue already exists, and force-use-existing-queue is its inverse, refused
// when none does -- two-sided, so a wrong guess is a refused enable either way. Enablement
// is the wrong signal in both directions: the real capture has `vpn-01  U S U - - - Y A`,
// admin-UP with the queue n/a, so an enabled VPN need not have one.
func TestReplicationVPNLinesSetsRoleWithoutCyclingReplication(t *testing.T) {
	state := map[string]VPNRepl{
		// Already up: the role changes under it, and nothing else happens.
		"LIVE": {Admin: AdminEnabled, Role: RoleStandby, Queue: QueueDown},
		// Down with no queue: needs enabling, bare default-guarded form.
		"NEW": {Admin: AdminShutdown, Role: RoleStandby, Queue: QueueNA},
		// Down but the queue survived: needs enabling, and the bare form would be refused.
		"DOWN_WITH_QUEUE": {Admin: AdminShutdown, Role: RoleActive, Queue: QueueDown},
		// Enabled with NO queue -- the vpn-01 row, the case reading enablement got wrong.
		"ENABLED_NO_QUEUE": {Admin: AdminEnabled, Role: RoleStandby, Queue: QueueNA},
	}
	vpns := []config.ReplVPN{
		{Name: "LIVE", ActiveAt: siteA},
		{Name: "NEW", ActiveAt: siteA},
		{Name: "DOWN_WITH_QUEUE", ActiveAt: siteA},
		{Name: "ENABLED_NO_QUEUE", ActiveAt: siteA},
	}
	lines, _ := replicationVPNLines(vpns, siteAEntry(), state, false)
	body := strings.Join(lines, "\n")

	for _, c := range []struct {
		vpn  string
		want []string
		why  string
	}{
		{"LIVE", []string{"state active"},
			"already up and replicating: the role changes in place, with NOTHING else sent"},
		{"ENABLED_NO_QUEUE", []string{"state active"},
			"also already up: still no cycling, whatever its queue column says"},
		{"NEW", []string{"state active", "no shutdown"},
			"was down, so it needs enabling -- role FIRST, so it never comes up in the wrong one"},
		{"DOWN_WITH_QUEUE", []string{"state active", "no shutdown force-use-existing-queue"},
			"was down but kept its queue, so the default guard would refuse the bare form"},
	} {
		want := strings.Join(vpnReplicationBlock(c.vpn, c.want...), "\n")
		if !strings.Contains(body, want) {
			t.Errorf("%s: %s\ngot:\n%s\nwant substring:\n%s", c.vpn, c.why, body, want)
		}
	}
	// The heart of it: no listed VPN is ever taken down to change its role.
	for _, vpn := range []string{"LIVE", "ENABLED_NO_QUEUE"} {
		block := strings.Join(vpnReplicationBlock(vpn, "shutdown"), "\n")
		if strings.Contains(body, block) {
			t.Errorf("%s was shut down to change its role; a role is set in place:\n%s", vpn, body)
		}
	}
	// force-recreate-queue discards whatever the queue holds. Nothing here may emit it.
	if strings.Contains(body, "force-recreate-queue") {
		t.Errorf("force-recreate-queue discards spooled messages and must never be emitted:\n%s", body)
	}
}

// TestNothingEverShutsDownTheVPNItself is the blast-radius guarantee, and it is the one
// property here that separates an interruption from an outage.
//
// `configure dr` disables REPLICATION. It never shuts down a message-VPN. Every command
// either phase sends is nested inside that VPN's own `replication` node, so clients stay
// connected and only the feed to the mate stops. A bare `shutdown` one level higher --
// directly under `message-vpn "X"` -- would disconnect every client of that VPN, and on
// phase 1's path it would do so to EVERY replicating VPN on the broker at once.
//
// The two forms differ by one level of indentation and nothing else, which is exactly why
// this is worth a test rather than a comment: the mistake is invisible in review, produces
// a script that still looks right, and is only discovered in production.
func TestNothingEverShutsDownTheVPNItself(t *testing.T) {
	state := map[string]VPNRepl{
		"LISTED":   {Admin: AdminEnabled, Role: RoleStandby, Queue: QueueDown},
		"UNLISTED": {Admin: AdminEnabled, Role: RoleActive, Queue: QueueDown},
		"OFF":      {Admin: AdminShutdown, Role: RoleStandby, Queue: QueueNA},
	}
	vpns := []config.ReplVPN{{Name: "LISTED", ActiveAt: siteA}, {Name: "OFF", ActiveAt: siteB}}

	// Both phases, and both the mate-changed and mate-unchanged shapes of phase 2.
	var all []string
	for _, stopped := range []bool{false, true} {
		lines, _ := replicationVPNLines(vpns, siteAEntry(), state, stopped)
		all = append(all, lines...)
	}
	for _, name := range mateConvergenceShutdowns(state) {
		all = append(all, vpnReplicationBlock(name, "shutdown")...)
	}

	// Walk the generated script the way the broker reads it: a command belongs to the
	// node it is indented under. Anything at the `message-vpn` child level that is not
	// the `replication` node itself is a command against the VPN.
	const vpnChildIndent = "  "
	inVPN := false
	for _, l := range all {
		switch {
		case strings.HasPrefix(l, "message-vpn "):
			inVPN = true
		case !strings.HasPrefix(l, " "):
			inVPN = false
		case inVPN && strings.HasPrefix(l, vpnChildIndent) && !strings.HasPrefix(l, vpnChildIndent+" "):
			body := strings.TrimSpace(l)
			if body != "replication" && body != "exit" {
				t.Errorf("%q sits directly under message-vpn, so it acts on the VPN rather than "+
					"its replication node; every command here must be nested one level deeper:\n%s",
					body, strings.Join(all, "\n"))
			}
		}
	}

	// And the belt-and-braces version of the same thing, stated as the shape that would
	// actually cause the outage.
	joined := strings.Join(all, "\n")
	if strings.Contains(joined, "\n"+vpnChildIndent+"shutdown") {
		t.Errorf("a bare `shutdown` under message-vpn disconnects every client of that VPN:\n%s", joined)
	}
}

// TestReplicationVPNLinesReEnablesWhatPhase1StoppedOnly pins the one case where a listed
// VPN that was already up still gets an enable: phase 1 took it down to converge the mate,
// so this command is responsible for bringing it back.
//
// When phase 1 did NOT run, the same VPN in the same state must produce no enable at all.
// That difference is the whole point of the change, because the no-phase-1 run is the
// common one -- every re-run after the mate is settled.
func TestReplicationVPNLinesReEnablesWhatPhase1StoppedOnly(t *testing.T) {
	state := map[string]VPNRepl{
		"LIVE":     {Admin: AdminEnabled, Role: RoleStandby, Queue: QueueDown},
		"UNLISTED": {Admin: AdminEnabled, Role: RoleActive, Queue: QueueDown},
	}
	vpns := []config.ReplVPN{{Name: "LIVE", ActiveAt: siteA}}

	stopped, res := replicationVPNLines(vpns, siteAEntry(), state, true)
	stoppedBody := strings.Join(stopped, "\n")
	if !strings.Contains(stoppedBody, "no shutdown") {
		t.Errorf("phase 1 stopped this VPN, so phase 2 must turn it back on:\n%s", stoppedBody)
	}
	// And the unlisted one is NOT shut down again -- phase 1 already did it.
	if strings.Contains(stoppedBody, strings.Join(vpnReplicationBlock("UNLISTED", "shutdown"), "\n")) {
		t.Error("phase 1 already stopped the unlisted VPN; re-sending a shutdown adds nothing " +
			"and gives the broker a second chance to refuse it")
	}
	// It is still REPORTED as disabled, because it is.
	if strings.Join(res.Disabled, ",") != "UNLISTED" {
		t.Errorf("Disabled = %v, want the unlisted VPN recorded even though phase 1 stopped it", res.Disabled)
	}

	running, _ := replicationVPNLines(vpns, siteAEntry(), state, false)
	runningBody := strings.Join(running, "\n")
	if strings.Contains(runningBody, "no shutdown") {
		t.Errorf("nothing stopped this VPN, so nothing may re-enable it -- that is the "+
			"interruption this change exists to remove:\n%s", runningBody)
	}
}

// TestMateConvergenceShutdownsIsDeterministic proves the phase-1 shutdown set -- a second
// map iteration alongside replicationVPNLines' own -- is sorted the same way: a Go map
// range is randomised, and an unsorted second range would make a generated script differ
// run to run even though replicationVPNLines' own set stayed identical.
//
// It also pins the PREDICATE: admin state alone, not Replicating(). "BBB" is enabled with
// no resolved role yet (a state the flag table can represent, see VPNRepl's own doc
// comment) and must still be shut down -- the broker's precondition is about replication
// being enabled, not about which role a VPN currently holds. "#config-sync" -- the
// internal VPN every broker carries, reporting "-" for both facts -- must never appear.
func TestMateConvergenceShutdownsIsDeterministic(t *testing.T) {
	state := map[string]VPNRepl{
		"AAA":          {Admin: AdminEnabled, Role: RoleActive},
		"BBB":          {Admin: AdminEnabled, Role: RoleNA},
		"CCC":          {Admin: AdminShutdown, Role: RoleStandby},
		"#config-sync": {Admin: AdminNA, Role: RoleNA},
	}
	first := mateConvergenceShutdowns(state)
	for i := 0; i < 20; i++ {
		again := mateConvergenceShutdowns(state)
		if strings.Join(again, ",") != strings.Join(first, ",") {
			t.Fatalf("run %d = %v, want %v", i, again, first)
		}
	}
	want := []string{"AAA", "BBB"}
	if strings.Join(first, ",") != strings.Join(want, ",") {
		t.Errorf("got %v, want %v -- keyed on ADMIN STATE alone, not Replicating()", first, want)
	}
}

// TestPlannedRolesNamesEveryListedVPN proves the caller can describe the whole move
// before any of it runs, which is what the exact-`yes` gate needs to show.
func TestPlannedRolesNamesEveryListedVPN(t *testing.T) {
	got := PlannedRoles(replVPNs(), siteAEntry())
	if len(got) != 2 {
		t.Fatalf("got %d roles, want one per listed VPN: %+v", len(got), got)
	}
	if got["ORDERS"] != RoleActive || got["PAYMENTS"] != RoleStandby {
		t.Errorf("roles = %+v", got)
	}
}

// TestRouterNameReadsTheCaptureHeader proves the self-identification source.
//
// The header is used rather than `show router-name` because its format is PROVEN on both
// platforms and in both header generations, and inventing a parser for output nobody here
// has seen is the guess that made the `show replication` layout and the VPN flag table
// wrong twice. The quoting is what makes it unambiguous: a router name is operator-chosen
// and may contain almost anything, so nothing has to work out where it ends.
func TestRouterNameReadsTheCaptureHeader(t *testing.T) {
	cases := map[string]string{
		// software, semp/all.cli:10
		"software": "!   Router: \"xps-ps-01\"\n",
		// appliance, semp/show-currentconfig-all.out:10
		"appliance": "!   Router: \"tuas9csol1\"\n",
		// no padding, and surrounded by the rest of a real header
		"in a full header": "! Command: \"show current-config all\"\n! Generated by:\n" +
			"!   Username: \"appuser\"\n!Router: \"sol-a1\"\n!   Version: \"10.26.0.8827\"\n",
	}
	want := map[string]string{"software": "xps-ps-01", "appliance": "tuas9csol1", "in a full header": "sol-a1"}
	for name, header := range cases {
		t.Run(name, func(t *testing.T) {
			m := routerNameRE.FindStringSubmatch(header)
			if m == nil {
				t.Fatalf("no match in:\n%s", header)
			}
			if m[1] != want[name] {
				t.Errorf("router name = %q, want %q", m[1], want[name])
			}
		})
	}

	t.Run("a name with spaces survives", func(t *testing.T) {
		m := routerNameRE.FindStringSubmatch("!   Router: \"a name\"\n")
		if m == nil || m[1] != "a name" {
			t.Errorf("quoting is what bounds the name; got %v", m)
		}
	})

	t.Run("no header line", func(t *testing.T) {
		if routerNameRE.FindStringSubmatch("! Command: \"show current-config all\"\n") != nil {
			t.Error("a header without a Router line must not match")
		}
	})
}

// TestVPNReplicationBlockShape pins the emitted block against the shape a capture uses, so
// a generated script reads beside one rather than against it: the VPN named and quoted, the
// replication node entered, the commands indented inside it, and both levels closed.
//
// It takes whatever commands the caller has, in the caller's order. Phase 2 usually sends
// one line -- `state <role>` alone, since a role is set in place -- and adds an enable only
// where one is needed, so this must hold for a block of any length rather than for a fixed
// trio, which is what it used to assert.
func TestVPNReplicationBlockShape(t *testing.T) {
	one := strings.Join(vpnReplicationBlock("X", "state active"), "\n")
	wantOne := "message-vpn \"X\"\n  replication\n    state active\n    exit\n  exit"
	if one != wantOne {
		t.Errorf("single-command block =\n%s\nwant\n%s", one, wantOne)
	}
	two := strings.Join(vpnReplicationBlock("X", "state active", "no shutdown"), "\n")
	wantTwo := "message-vpn \"X\"\n  replication\n    state active\n    no shutdown\n    exit\n  exit"
	if two != wantTwo {
		t.Errorf("two-command block =\n%s\nwant\n%s", two, wantTwo)
	}
}

// --- ConfigureReplication over a transport (A4 / RunCLI convergence) ----------

// replTestMate is the mate site most ConfigureReplication tests render against. It
// DIFFERS from what showReplicationSoftware reports: the capture holds TWO ssl addresses
// (broker-b:55443 and broker-b.cluster_default:55443) and this site names only the first,
// so RenderMateRemovals emits one removal line and the mate is judged to differ -- which
// is exactly what these tests need, since they exercise phase 1 actually running.
// replTestMateMatching, below, is the true already-matches fixture.
func replTestMate() config.ReplSite {
	return config.ReplSite{
		VirtualRouterName: "v:broker-b",
		Endpoints:         []config.ReplEndpoint{{Host: "broker-b", Port: 55443, Transport: config.TransportSSL}},
	}
}

// replTestMateMatching is a mate whose router name and BOTH endpoints are exactly what
// showReplicationSoftware reports. Unlike replTestMate, this is the genuinely-already-
// matches case, and it is what a test asserting the phase-1 SKIP has to build on: differing
// by even one endpoint -- as replTestMate deliberately does -- makes the mate differ and
// phase 1 run.
func replTestMateMatching() config.ReplSite {
	return config.ReplSite{
		VirtualRouterName: "v:broker-b",
		Endpoints: []config.ReplEndpoint{
			{Host: "broker-b", Port: 55443, Transport: config.TransportSSL},
			{Host: "broker-b.cluster_default", Port: 55443, Transport: config.TransportSSL},
		},
	}
}

// replReadOnlyResponder answers the two phase-0 reads every ConfigureReplication call
// makes, from the real-shaped fixtures in replication_test.go. Tests compose it with
// whatever the two write phases need to answer.
func replReadOnlyResponder(argv []string) ([]byte, bool) {
	switch {
	case matchCLI(argv, scriptShowReplication):
		return []byte(softwareBanner + showReplicationSoftware), true
	case matchCLI(argv, scriptShowVPN):
		return []byte(softwareBanner + showVPNReplication), true
	}
	return nil, false
}

// TestConfigureReplicationReadsTypeOffReadMateConfig is the A4 property: the broker type
// comes off readMateConfig's own `show replication` round trip (its login banner), not a
// separate dedicated banner call -- Ops.brokerType is gone. A responder that answers ONLY
// the scripts ConfigureReplication actually needs (no "banner" case at all) proves nothing
// else is asked for.
func TestConfigureReplicationReadsTypeOffReadMateConfig(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if out, ok := replReadOnlyResponder(argv); ok {
			return out, nil
		}
		switch {
		case matchCLI(argv, scriptConfigureReplMate):
			return []byte(softwareBanner + "xps-ps-01> \n"), nil
		case matchCLI(argv, scriptConfigureReplVPNs):
			return []byte(softwareBanner + "xps-ps-01> \n"), nil
		case matchCLI(argv, scriptReplStatus):
			return []byte(softwareBanner + showReplicationSoftware), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	self := config.ReplSite{VirtualRouterName: siteA, RouterNames: []string{"sol-a1"}}
	if _, err := o.ConfigureReplication(context.Background(), config.Primary, self,
		replTestMate(), replListed()); err != nil {
		t.Fatalf("ConfigureReplication error: %v", err)
	}
	if hasCall(ft, "banner") {
		t.Error("ConfigureReplication must not make a separate dedicated banner call")
	}
	if !hasCall(ft, scriptConfigureReplMate) {
		t.Error("replTestMate differs from the capture by one endpoint, so phase 1 must run")
	}
	// LISTED VPNs, not nil: phase 1 has stopped everything here, so its own shutdowns now
	// stand in for phase 2's unlisted pass and a run with nothing listed would correctly
	// send no phase-2 chunk. The roles are what gives phase 2 work.
	if !hasCall(ft, scriptConfigureReplVPNs) {
		t.Error("VPNs are listed, so phase 2 must run to give each its role")
	}
}

// TestConfigureReplicationSkipsPhase2WhenPhase1DidItAll pins the other half of that, which
// is a real saving rather than an accident: when the mate differs and the env file lists no
// VPN, phase 1 stops every replicating VPN and there is nothing left for phase 2 to say --
// so no second chunk is sent at all.
//
// Before a role could be set in place, phase 2 always had work here, because it re-sent a
// shutdown to each unlisted VPN that phase 1 had just stopped. That was a second broker
// round trip, and a second chance for the broker to refuse a line, for no change in state.
func TestConfigureReplicationSkipsPhase2WhenPhase1DidItAll(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if out, ok := replReadOnlyResponder(argv); ok {
			return out, nil
		}
		switch {
		case matchCLI(argv, scriptConfigureReplMate):
			return []byte(softwareBanner + "xps-ps-01> \n"), nil
		case matchCLI(argv, scriptReplStatus):
			return []byte(softwareBanner + showReplicationSoftware), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	self := config.ReplSite{VirtualRouterName: siteA, RouterNames: []string{"sol-a1"}}
	res, err := o.ConfigureReplication(context.Background(), config.Primary, self, replTestMate(), nil)
	if err != nil {
		t.Fatalf("ConfigureReplication error: %v", err)
	}
	if !hasCall(ft, scriptConfigureReplMate) {
		t.Fatal("the mate differs, so phase 1 must run")
	}
	if hasCall(ft, scriptConfigureReplVPNs) {
		t.Error("phase 1 already stopped every replicating VPN and nothing is listed, so phase 2 " +
			"has nothing to send -- an empty chunk is a round trip and a rejection risk for no change")
	}
	// Still REPORTED as disabled, because they are: the report describes the end state,
	// not which phase happened to produce it.
	if strings.Join(res.Disabled, ",") == "" {
		t.Error("the VPNs phase 1 stopped and the file does not list must still be reported as disabled")
	}
}

// replListed names VPNs the shared capture actually reports, which is what phase 2 needs
// to have work of its own.
//
// It matters more since a role stopped being a shutdown/state/enable cycle. Phase 2's
// unlisted pass is now SKIPPED when phase 1 already stopped those VPNs, so a run with no
// listed VPNs and a differing mate correctly sends no phase-2 chunk at all -- which means
// a test that wants to exercise phase 2 has to list something.
func replListed() []config.ReplVPN {
	return []config.ReplVPN{
		{Name: "default", ActiveAt: siteA},
		{Name: "vpn-01", ActiveAt: siteB},
	}
}

// TestConfigureReplicationSkipsMateWhenAlreadyMatching is the safety property the whole
// phased design rests on: phase 1 stops replication on every VPN this broker is
// replicating, so getting "differs" wrong in the eager direction turns an ordinary re-run
// into a site-wide outage. Built on replTestMateMatching, which really does match the
// capture in both endpoints and router name.
func TestConfigureReplicationSkipsMateWhenAlreadyMatching(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if out, ok := replReadOnlyResponder(argv); ok {
			return out, nil
		}
		switch {
		case matchCLI(argv, scriptConfigureReplVPNs):
			return []byte(softwareBanner + "xps-ps-01> \n"), nil
		case matchCLI(argv, scriptReplStatus):
			return []byte(softwareBanner + showReplicationSoftware), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	self := config.ReplSite{VirtualRouterName: siteA, RouterNames: []string{"sol-a1"}}
	// Listed VPNs on purpose, not nil. With nil there is no phase-2 body, so the test
	// could not tell "phase 1 was skipped" from "the whole apply returned early once the
	// mate matched" -- and the second is the regression that would silently stop applying
	// every role while this test still passed.
	// Names the broker's own fixture actually reports, because the pre-write check
	// refuses a listed VPN this broker does not have.
	listed := []config.ReplVPN{{Name: "default", ActiveAt: siteA}, {Name: "vpn-01", ActiveAt: siteB}}
	res, err := o.ConfigureReplication(context.Background(), config.Primary, self,
		replTestMateMatching(), listed)
	if err != nil {
		t.Fatalf("ConfigureReplication error: %v", err)
	}
	if hasCall(ft, scriptConfigureReplMate) {
		t.Error("the mate already matches; phase 1 must write NOTHING and shut down NO VPN")
	}
	if res.MateApplied {
		t.Error("MateApplied must be false when phase 1 was skipped")
	}
	if len(res.StoppedForMate) != 0 {
		t.Errorf("StoppedForMate = %v, want none: phase 1 never ran", res.StoppedForMate)
	}
	// The half that makes the rest of this test mean anything: skipping phase 1 must not
	// skip phase 2. A matching mate is the COMMON case -- every re-run after the first --
	// so an early return here would leave every listed VPN's role unapplied forever.
	if !hasCall(ft, scriptConfigureReplVPNs) {
		t.Error("phase 2 must still run when phase 1 is skipped: the roles are its job, not phase 1's")
	}
	if len(res.Roles) == 0 {
		t.Error("a run with listed VPNs must report the roles it applied")
	}
}

// TestConfigureReplicationPhase1BodyOrderIsShutdownsThenRemovalsThenMate pins the owner's
// specified body order for phase 1: every currently-enabled VPN shut down first (sorted,
// so two applies against the same broker state diff cleanly), THEN the removals, THEN the
// new mate lines -- removals before the new lines so a changed virtual-router-name never
// collides with the one already set.
func TestConfigureReplicationPhase1BodyOrderIsShutdownsThenRemovalsThenMate(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if out, ok := replReadOnlyResponder(argv); ok {
			return out, nil
		}
		switch {
		case matchCLI(argv, scriptConfigureReplMate):
			return []byte(softwareBanner + "xps-ps-01> \n"), nil
		case matchCLI(argv, scriptConfigureReplVPNs):
			return []byte(softwareBanner + "xps-ps-01> \n"), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	self := config.ReplSite{VirtualRouterName: siteA, RouterNames: []string{"sol-a1"}}
	if _, err := o.ConfigureReplication(context.Background(), config.Primary, self, replTestMate(), nil); err != nil {
		t.Fatalf("ConfigureReplication error: %v", err)
	}
	body := ft.wrappedCall(t, scriptConfigureReplMate).stdin

	// "default" and "vpn-01" are the two admin-enabled VPNs in showVPNReplication.
	shutdownDefault := strings.Index(body, `message-vpn "default"`)
	shutdownVPN01 := strings.Index(body, `message-vpn "vpn-01"`)
	removal := strings.Index(body, "no replication mate connect-via")
	mateLine := strings.Index(body, "replication mate virtual-router-name")
	if shutdownDefault < 0 || shutdownVPN01 < 0 || removal < 0 || mateLine < 0 {
		t.Fatalf("phase 1 body missing an expected line:\n%s", body)
	}
	if !(shutdownDefault < removal && shutdownVPN01 < removal && removal < mateLine) {
		t.Errorf("phase 1 order must be shutdowns, then removals, then mate lines:\n%s", body)
	}
	if shutdownDefault > shutdownVPN01 {
		t.Errorf("the shutdown set must be sorted (\"default\" before \"vpn-01\"):\n%s", body)
	}
}

// TestConfigureReplicationRefusesAMissingListedVPN proves the refuse-before-write check:
// a listed VPN this broker's own table does not report would abort phase 2 partway
// through under stop-on-error, shutting down every VPN after it in file order and never
// re-enabling them -- so it is refused before anything is written.
func TestConfigureReplicationRefusesAMissingListedVPN(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if out, ok := replReadOnlyResponder(argv); ok {
			return out, nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	self := config.ReplSite{VirtualRouterName: siteA, RouterNames: []string{"sol-a1"}}
	vpns := []config.ReplVPN{{Name: "NOSUCHVPN", ActiveAt: siteA}}
	_, err := o.ConfigureReplication(context.Background(), config.Primary, self, replTestMateMatching(), vpns)
	if err == nil {
		t.Fatal("a listed VPN absent from the broker's own table must be refused")
	}
	if !strings.Contains(err.Error(), "NOSUCHVPN") {
		t.Errorf("err = %v, want it to name the missing VPN", err)
	}
	if hasCall(ft, scriptConfigureReplMate) || hasCall(ft, scriptConfigureReplVPNs) {
		t.Error("nothing must be written when a listed VPN does not exist")
	}
}

// TestConfigureReplicationPhase1RejectionStopsBeforePhase2 pins the honesty requirement
// for a phase-1 failure: the broker stopped at the rejected line, replication is left
// shut down on VPNs this run will not turn back on, and phase 2 must never be sent --
// checking it never even ran is what proves that, rather than merely asserting the text.
//
// The error must NAME the VPNs. Describing them as a category ("every VPN this broker was
// replicating") reads as informative and tells an operator nothing they can act on: the
// names are known at the call site and thrown away everywhere else, so if this error does
// not carry them, nothing does.
//
// It must also not overclaim. Phase 1 sends the shutdowns FIRST, then the removals, then
// the mate lines, and stop-on-error abandons the chunk wherever it stopped -- so a
// rejection inside the shutdown sequence means only a PREFIX of those VPNs was stopped.
// An error asserting all of them were is the same unguaranteeable claim the single-script
// version was rewritten to stop making.
func TestConfigureReplicationPhase1RejectionStopsBeforePhase2(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if out, ok := replReadOnlyResponder(argv); ok {
			return out, nil
		}
		if matchCLI(argv, scriptConfigureReplMate) {
			return []byte(softwareBanner + "Error: invalid mate address\n"), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	self := config.ReplSite{VirtualRouterName: siteA, RouterNames: []string{"sol-a1"}}
	_, err := o.ConfigureReplication(context.Background(), config.Primary, self, replTestMate(), nil)
	if err == nil {
		t.Fatal("ConfigureReplication should fail when the broker rejects phase 1")
	}
	// Each entry is the shortest phrase that still carries the FACT, not a sentence. An
	// earlier version pinned "NOT be turned back on" and broke the moment the sentence
	// around it was reworded to say replication rather than the VPN -- which was a
	// correction to the error, not a regression, and a test that fails on those is a test
	// nobody trusts.
	for _, want := range []string{
		"phase 1",                // which phase stopped
		"rejected",               // a refused line, not an unreachable broker
		"replication disabled",   // what is actually off
		"turned back on",         // and that this run will not restore it
		"Nothing is rolled back", // nothing undoes it
		"phase 2",                // the phase that was never sent
		"show replication",       // where to read the truth
		// The blast radius, which is the difference between an interruption and an
		// outage: only replication stopped, so nobody's clients were dropped. An
		// operator reading this error is deciding how urgent it is.
		"clients are still connected",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, missing %q", err, want)
		}
	}
	// The names, not a category. replReadOnlyResponder's fixture reports `default` and
	// `vpn-01` admin-enabled, so those are the two phase 1 shut down.
	for _, vpn := range []string{`"default"`, `"vpn-01"`} {
		if !strings.Contains(err.Error(), vpn) {
			t.Errorf("err = %v, must NAME the VPN %s it left with replication down", err, vpn)
		}
	}
	// Not "every VPN was stopped": phase 1 sends the shutdowns first and may have been
	// rejected partway through them.
	if strings.Contains(err.Error(), "was shut down on every VPN") {
		t.Errorf("err = %v, must not assert every VPN was stopped -- the chunk may have been "+
			"rejected inside the shutdown sequence itself", err)
	}
	if hasCall(ft, scriptConfigureReplVPNs) {
		t.Error("phase 2 must never be sent once phase 1 is rejected")
	}
}

// TestConfigureReplicationClosingShowFailureStillReportsTheApply pins that a display-only
// read failing is NOT the apply failing.
//
// Both write phases have already succeeded by the time the closing `show` runs, and the
// result is the ONLY record of what they did -- built in memory and never read back from
// the broker. Returning the read's error would throw that record away and tell an operator
// whose configuration landed perfectly that the command failed, which is the worst reading
// available: they would go looking for damage that is not there, or re-run a destructive
// command they did not need to. It warns and reports instead.
func TestConfigureReplicationClosingShowFailureStillReportsTheApply(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		// The closing status read reuses scriptShowVPN, so the failure has to be pinned on
		// the one script only the closing read uses -- otherwise it would break the
		// pre-write read instead and nothing would ever be applied.
		if matchCLI(argv, scriptReplStatus) {
			return nil, errors.New("connection closed")
		}
		if out, ok := replReadOnlyResponder(argv); ok {
			return out, nil
		}
		return []byte(softwareBanner + "xps-ps-01> \n"), nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	self := config.ReplSite{VirtualRouterName: siteA, RouterNames: []string{"sol-a1"}}
	res, err := o.ConfigureReplication(context.Background(), config.Primary, self, replTestMate(), nil)
	if err != nil {
		t.Fatalf("a failed closing read must not fail an apply that succeeded: %v", err)
	}
	if res == nil {
		t.Fatal("the result is the only record of what the two write phases did; it must survive")
	}
	if !res.MateApplied {
		t.Error("the mate differs in this fixture, so phase 1 ran and the result must say so")
	}
	if res.Status != "" {
		t.Errorf("Status = %q, want empty: the closing picture could not be read", res.Status)
	}
}

// TestConfigureReplicationPhase2RejectionAfterPhase1Success pins the worse of the two
// failures: phase 1 already succeeded and is NOT rolled back, and the error must say so,
// because converging the mate meant shutting replication down on VPNs that phase 2 had
// not finished turning back on. It must NAME them and name the listed ones it was working
// through, for the same reason the phase-1 error must: the categories are known here and
// nowhere the operator can reach.
func TestConfigureReplicationPhase2RejectionAfterPhase1Success(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if out, ok := replReadOnlyResponder(argv); ok {
			return out, nil
		}
		switch {
		case matchCLI(argv, scriptConfigureReplMate):
			return []byte(softwareBanner + "xps-ps-01> \n"), nil
		case matchCLI(argv, scriptConfigureReplVPNs):
			return []byte(softwareBanner + "Error: invalid VPN state\n"), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	self := config.ReplSite{VirtualRouterName: siteA, RouterNames: []string{"sol-a1"}}
	// Listed VPNs, so phase 2 has work: with nothing listed, phase 1's own shutdowns cover
	// every unlisted VPN and no phase-2 chunk is sent for this test to have rejected.
	_, err := o.ConfigureReplication(context.Background(), config.Primary, self,
		replTestMate(), replListed())
	if err == nil {
		t.Fatal("ConfigureReplication should fail when the broker rejects phase 2")
	}
	for _, want := range []string{"phase 2", "rejected", "show replication", "NOT rolled back", "phase 1"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, missing %q", err, want)
		}
	}
	// The VPNs phase 1 stopped for the mate, by name: the fixture reports `default` and
	// `vpn-01` admin-enabled, and phase 2 did not get them all back up.
	for _, vpn := range []string{`"default"`, `"vpn-01"`} {
		if !strings.Contains(err.Error(), vpn) {
			t.Errorf("err = %v, must NAME the VPN %s phase 1 shut down and phase 2 left down", err, vpn)
		}
	}
	if !hasCall(ft, scriptConfigureReplMate) {
		t.Error("phase 1 must have run (and succeeded) before phase 2 is attempted")
	}
}

// TestConfigureReplicationTransportFailureIsNotReportedAsARejection proves the other
// failure shape: a connection that dies mid-call is NOT the same as a rejected
// configuration line, and the two must not be worded the same way -- the broker may have
// kept applying the script after this process lost the connection to it, so how much
// landed is genuinely unknown, unlike a rejection where the broker's own stop-on-error
// wrapper guarantees where it stopped.
func TestConfigureReplicationTransportFailureIsNotReportedAsARejection(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if out, ok := replReadOnlyResponder(argv); ok {
			return out, nil
		}
		if matchCLI(argv, scriptConfigureReplMate) {
			return nil, errors.New("connection reset by peer")
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	self := config.ReplSite{VirtualRouterName: siteA, RouterNames: []string{"sol-a1"}}
	_, err := o.ConfigureReplication(context.Background(), config.Primary, self, replTestMate(), nil)
	if err == nil {
		t.Fatal("a transport failure must surface as an error")
	}
	if strings.Contains(err.Error(), "rejected by the broker") {
		t.Errorf("err = %v, must not claim the broker rejected a line -- the connection just failed", err)
	}
	if !strings.Contains(err.Error(), "NOT known") {
		t.Errorf("err = %v, want it to say the outcome is not known", err)
	}
	if !strings.Contains(err.Error(), "phase 1") {
		t.Errorf("err = %v, want it to name the phase", err)
	}
}

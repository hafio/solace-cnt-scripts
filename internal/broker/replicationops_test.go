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
	lines, res := replicationVPNLines(vpns, siteAEntry(), state)
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
		siteAEntry(), map[string]VPNRepl{})
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
	first, res := replicationVPNLines(nil, siteAEntry(), state)
	for i := 0; i < 20; i++ {
		again, _ := replicationVPNLines(nil, siteAEntry(), state)
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

// TestReplicationVPNLinesOrderIsShutdownStateEnable pins DEFECT 2's fix: a role is only
// ever touched while replication is disabled (the owner's stated broker constraint; see
// replicationVPNLines' own doc comment for the counter-evidence this is built on
// regardless), so every listed VPN must be SHUT DOWN, given its role, and only then
// re-enabled -- never enabled first, which is what this used to emit.
//
// It also pins the re-enable KEYWORD, which is read from the QUEUE column and never from
// enablement. The broker's default is fail-on-existing-queue, refused when a replication
// queue already exists, and force-use-existing-queue is its inverse, refused when there is
// none -- a two-sided choice, so a wrong guess is a refused re-enable either way.
//
// Enablement is the wrong signal in BOTH directions, which is why this reads Q, and the
// real capture settles it: semp/show-message-vpn-replication.out carries
// `vpn-01  U S U - - - Y A` -- admin-UP with the queue n/a, so an enabled VPN need not
// have one. And phase 1 shuts VPNs down, so on the re-run this command's own errors ask
// for, a VPN that still holds its queue reads back admin-DOWN.
func TestReplicationVPNLinesOrderIsShutdownStateEnable(t *testing.T) {
	state := map[string]VPNRepl{
		// No queue: the bare, default-guarded form.
		"NEW": {Admin: AdminShutdown, Role: RoleStandby, Queue: QueueNA},
		// Enabled with NO queue -- the vpn-01 row from the real capture, and exactly the
		// case reading enablement instead of the queue column got wrong.
		"ENABLED_NO_QUEUE": {Admin: AdminEnabled, Role: RoleStandby, Queue: QueueNA},
		// Shut down but the queue is still there: what phase 1 leaves behind, so this is
		// the re-run path the failure errors send an operator down.
		"DOWN_WITH_QUEUE": {Admin: AdminShutdown, Role: RoleActive, Queue: QueueDown},
		// A report carrying no Q column: fall back to the broker's own default rather
		// than assert something that was never read.
		"UNREADABLE": {Admin: AdminEnabled, Role: RoleActive, Queue: QueueUnknown},
	}
	vpns := []config.ReplVPN{
		{Name: "NEW", ActiveAt: siteA},
		{Name: "ENABLED_NO_QUEUE", ActiveAt: siteA},
		{Name: "DOWN_WITH_QUEUE", ActiveAt: siteA},
		{Name: "UNREADABLE", ActiveAt: siteA},
	}
	lines, _ := replicationVPNLines(vpns, siteAEntry(), state)
	body := strings.Join(lines, "\n")

	for _, c := range []struct{ vpn, want, why string }{
		{"NEW", "no shutdown", "no queue means the bare, default-guarded form"},
		{"ENABLED_NO_QUEUE", "no shutdown",
			"enabled is NOT the same fact as having a queue, and asking to reuse one that does not exist is refused"},
		{"DOWN_WITH_QUEUE", "no shutdown force-use-existing-queue",
			"the queue outlived the shutdown, so the default guard would refuse this re-enable"},
		{"UNREADABLE", "no shutdown",
			"an unread queue column falls back to the broker's default, never to a claim the report did not make"},
	} {
		want := strings.Join(vpnReplicationBlock(c.vpn, "shutdown", "state active", c.want), "\n")
		if !strings.Contains(body, want) {
			t.Errorf("%s: %s\ngot:\n%s\nwant substring:\n%s", c.vpn, c.why, body, want)
		}
	}
	// force-recreate-queue discards whatever the queue holds. Nothing here may emit it.
	if strings.Contains(body, "force-recreate-queue") {
		t.Errorf("force-recreate-queue discards spooled messages and must never be emitted:\n%s", body)
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

// TestVPNReplicationBlockShape pins the emitted block against the shape a capture uses,
// so a generated script reads beside one rather than against it. Arguments are given in
// phase 2's own order -- shutdown, state, re-enable -- rather than the reversed order this
// used to take.
func TestVPNReplicationBlockShape(t *testing.T) {
	got := strings.Join(vpnReplicationBlock("X", "shutdown", "state active", "no shutdown"), "\n")
	want := "message-vpn \"X\"\n  replication\n    shutdown\n    state active\n    no shutdown\n    exit\n  exit"
	if got != want {
		t.Errorf("block =\n%s\nwant\n%s", got, want)
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
	if _, err := o.ConfigureReplication(context.Background(), config.Primary, self, replTestMate(), nil); err != nil {
		t.Fatalf("ConfigureReplication error: %v", err)
	}
	if hasCall(ft, "banner") {
		t.Error("ConfigureReplication must not make a separate dedicated banner call")
	}
	if !hasCall(ft, scriptConfigureReplMate) {
		t.Error("replTestMate differs from the capture by one endpoint, so phase 1 must run")
	}
	if !hasCall(ft, scriptConfigureReplVPNs) {
		t.Error("the capture reports default and vpn-01 replicating and unlisted, so phase 2 must run")
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
	for _, want := range []string{
		"phase 1",                // which phase stopped
		"rejected",               // a refused line, not an unreachable broker
		"NOT be turned back on",  // what that leaves stopped
		"Nothing is rolled back", // and that nothing undoes it
		"phase 2",                // the phase that was never sent
		"show replication",       // where to read the truth
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
	_, err := o.ConfigureReplication(context.Background(), config.Primary, self, replTestMate(), nil)
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

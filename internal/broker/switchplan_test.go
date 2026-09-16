package broker

import (
	"context"
	"errors"
	"sort"
	"strings"
	"testing"

	"solace/internal/config"
)

// switchplan_test.go covers the safety-critical half of the feature. Everything here is
// pure except ExecuteSwitchPlan, which drives recording fakes -- so the ordering rules
// are provable without a broker.

const (
	siteA = "v:sol-a"
	siteB = "v:sol-b"
)

// replVPNs is the file's intent: ORDERS active at A, PAYMENTS active at B.
func replVPNs() []config.ReplVPN {
	return []config.ReplVPN{
		{Name: "ORDERS", ActiveAt: siteA},
		{Name: "PAYMENTS", ActiveAt: siteB},
	}
}

// st builds one site's reported state from pairs of vpn -> role, all enabled.
func st(site string, roles map[string]ReplRole) SiteState {
	m := map[string]VPNRepl{}
	for v, r := range roles {
		m[v] = VPNRepl{Admin: AdminEnabled, Role: r}
	}
	return SiteState{Site: site, State: m}
}

func phaseActions(p SwitchPlan, kind PhaseKind) []SwitchAction {
	for _, ph := range p.Phases {
		if ph.Kind == kind {
			return ph.Actions
		}
	}
	return nil
}

func actionStrings(as []SwitchAction) []string {
	out := make([]string, 0, len(as))
	for _, a := range as {
		out = append(out, a.Site+"/"+a.VPN+"="+string(a.Role))
	}
	sort.Strings(out)
	return out
}

// TestBuildSwitchPlanConverged proves a pair already matching the file produces an EMPTY
// plan, and reports what it skipped.
//
// An empty plan is what stops a no-op run demanding an exact-`yes`: asking a destructive
// question about a run that destroys nothing teaches an operator to answer without
// reading, which is the opposite of what the gate is for.
func TestBuildSwitchPlanConverged(t *testing.T) {
	a := st(siteA, map[string]ReplRole{"ORDERS": RoleActive, "PAYMENTS": RoleStandby})
	b := st(siteB, map[string]ReplRole{"ORDERS": RoleStandby, "PAYMENTS": RoleActive})
	p, err := BuildSwitchPlan(replVPNs(), a, b)
	if err != nil {
		t.Fatalf("BuildSwitchPlan: %v", err)
	}
	if !p.Empty() {
		t.Errorf("a converged pair must produce an empty plan, got %+v", p.Phases)
	}
	if strings.Join(p.Skipped, ",") != "ORDERS,PAYMENTS" {
		t.Errorf("Skipped = %v, want both VPNs named", p.Skipped)
	}
	if len(p.Moves()) != 0 {
		t.Errorf("Moves = %v, want none", p.Moves())
	}
}

// TestBuildSwitchPlanMovesOneVPN is the single-VPN failover, and the reason the plan
// skips rather than sweeping: the VPN that is NOT moving is never demoted, so its
// clients stay connected.
func TestBuildSwitchPlanMovesOneVPN(t *testing.T) {
	// ORDERS is active at B but the file wants it at A. PAYMENTS is already right.
	a := st(siteA, map[string]ReplRole{"ORDERS": RoleStandby, "PAYMENTS": RoleStandby})
	b := st(siteB, map[string]ReplRole{"ORDERS": RoleActive, "PAYMENTS": RoleActive})
	p, err := BuildSwitchPlan(replVPNs(), a, b)
	if err != nil {
		t.Fatalf("BuildSwitchPlan: %v", err)
	}
	if got := strings.Join(p.Skipped, ","); got != "PAYMENTS" {
		t.Errorf("Skipped = %v, want only PAYMENTS", p.Skipped)
	}
	if got := actionStrings(phaseActions(p, PhaseDemote)); len(got) != 1 || got[0] != siteB+"/ORDERS=standby" {
		t.Errorf("demote = %v, want ORDERS demoted at B only", got)
	}
	if got := actionStrings(phaseActions(p, PhasePromote)); len(got) != 1 || got[0] != siteA+"/ORDERS=active" {
		t.Errorf("promote = %v, want ORDERS promoted at A only", got)
	}
	for _, a := range append(phaseActions(p, PhaseDemote), phaseActions(p, PhasePromote)...) {
		if a.VPN == "PAYMENTS" {
			t.Error("PAYMENTS is already correct and must not be touched")
		}
	}
}

// TestBuildSwitchPlanNeverDemotesTheTarget is the property that makes the demotion safe
// to issue unconditionally: it lands on the site the file does NOT name, so a VPN is
// never taken down at the site it is about to be promoted at.
func TestBuildSwitchPlanNeverDemotesTheTarget(t *testing.T) {
	a := st(siteA, map[string]ReplRole{"ORDERS": RoleStandby, "PAYMENTS": RoleActive})
	b := st(siteB, map[string]ReplRole{"ORDERS": RoleActive, "PAYMENTS": RoleStandby})
	p, err := BuildSwitchPlan(replVPNs(), a, b)
	if err != nil {
		t.Fatalf("BuildSwitchPlan: %v", err)
	}
	target := map[string]string{}
	for _, act := range phaseActions(p, PhasePromote) {
		target[act.VPN] = act.Site
	}
	for _, act := range phaseActions(p, PhaseDemote) {
		if target[act.VPN] == act.Site {
			t.Errorf("message-vpn %q is demoted at %s and promoted at the same site", act.VPN, act.Site)
		}
	}
	// Both VPNs move here, so both appear exactly once in each phase.
	if len(phaseActions(p, PhaseDemote)) != 2 || len(phaseActions(p, PhasePromote)) != 2 {
		t.Errorf("want one demotion and one promotion per moving VPN, got %+v", p.Phases)
	}
}

// TestBuildSwitchPlanPromoteNeverPrecedesDemote is a property over the phase ORDER
// itself, not an example: it is the invariant the whole design rests on.
func TestBuildSwitchPlanPromoteNeverPrecedesDemote(t *testing.T) {
	roles := []ReplRole{RoleActive, RoleStandby}
	for _, ra := range roles {
		for _, rb := range roles {
			if ra == RoleActive && rb == RoleActive {
				continue // split brain, refused; its own test
			}
			a := st(siteA, map[string]ReplRole{"ORDERS": ra})
			b := st(siteB, map[string]ReplRole{"ORDERS": rb})
			p, err := BuildSwitchPlan([]config.ReplVPN{{Name: "ORDERS", ActiveAt: siteA}}, a, b)
			if err != nil {
				t.Fatalf("a=%s b=%s: %v", ra, rb, err)
			}
			demoteAt, promoteAt := -1, -1
			for i, ph := range p.Phases {
				switch ph.Kind {
				case PhaseDemote:
					demoteAt = i
				case PhasePromote:
					promoteAt = i
				}
			}
			if promoteAt < demoteAt {
				t.Errorf("a=%s b=%s: promote phase precedes demote", ra, rb)
			}
			if p.Phases[demoteAt+1].Kind != PhaseConfirm {
				t.Errorf("a=%s b=%s: a confirm must sit between demote and promote, got %s",
					ra, rb, p.Phases[demoteAt+1].Kind)
			}
		}
	}
}

// TestBuildSwitchPlanRefusesSplitBrain proves an arriving two-actives state stops the
// run instead of being silently repaired.
//
// Demoting one side picks a winner, and which site keeps its spooled messages is not a
// decision to make on the operator's behalf. The message names the VPN, both sites, and
// which one the file favours, so the operator has what they need to choose.
func TestBuildSwitchPlanRefusesSplitBrain(t *testing.T) {
	a := st(siteA, map[string]ReplRole{"ORDERS": RoleActive})
	b := st(siteB, map[string]ReplRole{"ORDERS": RoleActive})
	_, err := BuildSwitchPlan([]config.ReplVPN{{Name: "ORDERS", ActiveAt: siteA}}, a, b)
	if err == nil {
		t.Fatal("a VPN active at both sites must refuse the run")
	}
	for _, want := range []string{"ORDERS", siteA, siteB} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q must name %q", err, want)
		}
	}
}

// TestBuildSwitchPlanRefusesUnusableVPNs proves the switch does not quietly become a
// configuration change. It only moves roles; creating or enabling a VPN is
// `configure dr`'s job, and the message says so.
func TestBuildSwitchPlanRefusesUnusableVPNs(t *testing.T) {
	t.Run("missing at one site", func(t *testing.T) {
		a := st(siteA, map[string]ReplRole{"ORDERS": RoleActive})
		b := st(siteB, map[string]ReplRole{})
		_, err := BuildSwitchPlan([]config.ReplVPN{{Name: "ORDERS", ActiveAt: siteA}}, a, b)
		if err == nil || !strings.Contains(err.Error(), "do not exist") {
			t.Fatalf("want a missing-VPN refusal, got: %v", err)
		}
		if !strings.Contains(err.Error(), "configure data-replication") {
			t.Errorf("the error must point at the command that fixes it, got: %v", err)
		}
	})

	t.Run("replication shut down", func(t *testing.T) {
		a := SiteState{Site: siteA, State: map[string]VPNRepl{"ORDERS": {Admin: AdminShutdown, Role: RoleStandby}}}
		b := st(siteB, map[string]ReplRole{"ORDERS": RoleActive})
		_, err := BuildSwitchPlan([]config.ReplVPN{{Name: "ORDERS", ActiveAt: siteA}}, a, b)
		if err == nil || !strings.Contains(err.Error(), "shut down") {
			t.Fatalf("want a disabled-replication refusal, got: %v", err)
		}
	})

	t.Run("n/a is not enabled", func(t *testing.T) {
		a := SiteState{Site: siteA, State: map[string]VPNRepl{"ORDERS": {Admin: AdminNA, Role: RoleNA}}}
		b := st(siteB, map[string]ReplRole{"ORDERS": RoleActive})
		if _, err := BuildSwitchPlan([]config.ReplVPN{{Name: "ORDERS", ActiveAt: siteA}}, a, b); err == nil {
			t.Error("an n/a VPN must not be treated as usable")
		}
	})
}

// TestBuildSwitchPlanRefusesAnUnknownActiveAt proves an activeAt naming neither site is
// caught here as well as at config load -- the plan is built from two site states that
// the caller supplies, and they could disagree with the file.
func TestBuildSwitchPlanRefusesAnUnknownActiveAt(t *testing.T) {
	a := st(siteA, map[string]ReplRole{"ORDERS": RoleActive})
	b := st(siteB, map[string]ReplRole{"ORDERS": RoleStandby})
	_, err := BuildSwitchPlan([]config.ReplVPN{{Name: "ORDERS", ActiveAt: "v:elsewhere"}}, a, b)
	if err == nil || !strings.Contains(err.Error(), "neither site") {
		t.Fatalf("want a refusal naming the unknown site, got: %v", err)
	}
}

// TestBuildSwitchPlanRefusesOneSiteTwice guards against a caller handing the same
// channel twice, which would make every "the other site" decision meaningless.
func TestBuildSwitchPlanRefusesOneSiteTwice(t *testing.T) {
	a := st(siteA, map[string]ReplRole{"ORDERS": RoleActive})
	if _, err := BuildSwitchPlan(replVPNs(), a, a); err == nil {
		t.Error("both sides reporting the same site must be refused")
	}
}

// --- execution ------------------------------------------------------------------

// scriptedMate is a MateChannel whose per-VPN state the test controls, and which records
// every write. ReadRoles returns the CURRENT state, so a test can mutate it between
// phases to model something else moving a role.
type scriptedMate struct {
	site   string
	state  map[string]VPNRepl
	writes []string
	setErr error
	// onRead runs before each ReadRoles, so a test can move state at exactly the
	// moment the code re-reads.
	onRead func()
	// preflight knobs. Zero values are the healthy case, so a test that is not about
	// the preflight does not have to say so.
	notPrimaryActive bool
	mateName         string // what this broker says its replication mate is
}

func (m *scriptedMate) Preflight(context.Context) error { return nil }
func (m *scriptedMate) Status(context.Context) (string, error) { return "", nil }
func (m *scriptedMate) Describe() string                       { return m.site }

func (m *scriptedMate) PrimaryActive(context.Context) (bool, error) {
	return !m.notPrimaryActive, nil
}

func (m *scriptedMate) ShowReplication(context.Context) (MateConfig, error) {
	return MateConfig{VirtualRouterName: m.mateName}, nil
}

func (m *scriptedMate) ReadRoles(_ context.Context, _ []string) (map[string]VPNRepl, error) {
	if m.onRead != nil {
		m.onRead()
	}
	out := map[string]VPNRepl{}
	for k, v := range m.state {
		out[k] = v
	}
	return out, nil
}

func (m *scriptedMate) SetRole(_ context.Context, vpn string, role ReplRole) error {
	m.writes = append(m.writes, vpn+"="+string(role))
	if m.setErr != nil {
		return m.setErr
	}
	m.state[vpn] = VPNRepl{Admin: AdminEnabled, Role: role}
	return nil
}

func twoMates() (map[string]MateChannel, *scriptedMate, *scriptedMate) {
	// Each broker names the OTHER site as its mate, which is the healthy pairing the
	// preflight requires.
	a := &scriptedMate{site: siteA, mateName: siteB, state: map[string]VPNRepl{
		"ORDERS": {Admin: AdminEnabled, Role: RoleStandby},
	}}
	b := &scriptedMate{site: siteB, mateName: siteA, state: map[string]VPNRepl{
		"ORDERS": {Admin: AdminEnabled, Role: RoleActive},
	}}
	return map[string]MateChannel{siteA: a, siteB: b}, a, b
}

// replSites is the file's view of the pair, keyed the way SwitchPreflight takes it.
func replSites() map[string]config.ReplSite {
	return map[string]config.ReplSite{
		siteA: {VirtualRouterName: siteA},
		siteB: {VirtualRouterName: siteB},
	}
}

// TestSwitchPreflightHealthyPair proves the gate passes on a pair where both primaries
// are active and each broker names the other.
func TestSwitchPreflightHealthyPair(t *testing.T) {
	chans, _, _ := twoMates()
	if err := SwitchPreflight(context.Background(), chans, replSites()); err != nil {
		t.Errorf("a healthy pair must pass the preflight: %v", err)
	}
}

// TestSwitchPreflightRequiresThePrimaryActive is the operator's gate: a
// switch runs only when each site is on its own PRIMARY HA node.
//
// The replication role is config-synced, so this is not about where a write lands -- it
// is about not stacking a cross-site role change on top of a local failover, which makes
// both harder to reason about and to undo. It is also the honest failure for the case
// that actually bites: when the backup is active BECAUSE the primary is down, the primary
// is the node this tool addresses, so the run would fail anyway -- and this says why.
func TestSwitchPreflightRequiresThePrimaryActive(t *testing.T) {
	chans, _, b := twoMates()
	b.notPrimaryActive = true
	err := SwitchPreflight(context.Background(), chans, replSites())
	if err == nil {
		t.Fatal("a site running on its backup must refuse the switch")
	}
	if !strings.Contains(err.Error(), siteB) {
		t.Errorf("the error must name the site, got: %v", err)
	}
	if !strings.Contains(err.Error(), "revert-activity") {
		t.Errorf("the error must say how to fix it, got: %v", err)
	}
}

// TestSwitchPreflightStandalonePasses proves a broker with no HA group is not refused for
// failing to hold activity it cannot hold: it is its own only node.
func TestSwitchPreflightStandalonePasses(t *testing.T) {
	chans, a, b := twoMates()
	// PrimaryActive returns true for a standalone, which is what cliMate reports when
	// `show redundancy` says Configuration Status is not Enabled.
	a.notPrimaryActive, b.notPrimaryActive = false, false
	if err := SwitchPreflight(context.Background(), chans, replSites()); err != nil {
		t.Errorf("a standalone pair must pass: %v", err)
	}
}

// TestSwitchPreflightRequiresMatchingVirtualRouterNames proves the brokers and the file
// must agree about which pair this is.
//
// A mismatch means one of three things, none safe to switch through: the file names a
// site the broker has never heard of, `configure data-replication` has not been run since
// the block changed, or the tool is talking to a node whose virtual router is not the one
// the pair was built around.
func TestSwitchPreflightRequiresMatchingVirtualRouterNames(t *testing.T) {
	chans, a, _ := twoMates()
	a.mateName = "v:somewhere-else"
	err := SwitchPreflight(context.Background(), chans, replSites())
	if err == nil {
		t.Fatal("a broker naming a different mate must refuse the switch")
	}
	for _, want := range []string{"v:somewhere-else", siteB, "configure data-replication"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q must mention %q", err, want)
		}
	}
}

// TestSwitchPreflightRunsBeforeAnyWrite proves every refusal happens with the pair
// untouched -- which is the whole reason the checks are a preflight rather than a step.
func TestSwitchPreflightRunsBeforeAnyWrite(t *testing.T) {
	for _, tc := range []struct {
		name  string
		breakIt func(a, b *scriptedMate)
	}{
		{"backup active", func(_, b *scriptedMate) { b.notPrimaryActive = true }},
		{"mate name mismatch", func(a, _ *scriptedMate) { a.mateName = "v:wrong" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chans, a, b := twoMates()
			tc.breakIt(a, b)
			if err := SwitchPreflight(context.Background(), chans, replSites()); err == nil {
				t.Fatal("expected a refusal")
			}
			if len(a.writes)+len(b.writes) != 0 {
				t.Errorf("the preflight must write nothing, got %v and %v", a.writes, b.writes)
			}
		})
	}
}

func ordersToA() []config.ReplVPN { return []config.ReplVPN{{Name: "ORDERS", ActiveAt: siteA}} }

// TestExecuteSwitchPlanOrder proves the writes land in the only order that is safe:
// the demotion first, then the promotion, with the read-backs in between.
func TestExecuteSwitchPlanOrder(t *testing.T) {
	chans, a, b := twoMates()
	plan, err := BuildSwitchPlan(ordersToA(),
		SiteState{siteA, a.state}, SiteState{siteB, b.state})
	if err != nil {
		t.Fatalf("BuildSwitchPlan: %v", err)
	}
	if err := ExecuteSwitchPlan(context.Background(), chans, plan, []string{"ORDERS"}, nil); err != nil {
		t.Fatalf("ExecuteSwitchPlan: %v", err)
	}
	if strings.Join(b.writes, ",") != "ORDERS=standby" {
		t.Errorf("site B writes = %v, want the demotion only", b.writes)
	}
	if strings.Join(a.writes, ",") != "ORDERS=active" {
		t.Errorf("site A writes = %v, want the promotion only", a.writes)
	}
}

// TestExecuteSwitchPlanStopsOnUnconfirmedDemotion proves a demotion that does not take
// stops the run BEFORE any promotion.
//
// The failure direction is deliberate: the VPN is left standby at both sites --
// unavailable, not corrupted -- and the message says so, because an operator who reads
// "unavailable" must not go looking for data loss.
func TestExecuteSwitchPlanStopsOnUnconfirmedDemotion(t *testing.T) {
	chans, a, b := twoMates()
	plan, err := BuildSwitchPlan(ordersToA(), SiteState{siteA, a.state}, SiteState{siteB, b.state})
	if err != nil {
		t.Fatalf("BuildSwitchPlan: %v", err)
	}
	// The write is ACCEPTED but the broker does not apply it -- exactly what `cli -Apes`
	// exiting 0 on a rejected line looks like from here, and the reason a demotion is
	// confirmed by reading it back rather than by the command having returned success.
	b.onRead = func() { b.state["ORDERS"] = VPNRepl{Admin: AdminEnabled, Role: RoleActive} }

	err = ExecuteSwitchPlan(context.Background(), chans, plan, []string{"ORDERS"}, nil)
	if err == nil {
		t.Fatal("an unconfirmed demotion must stop the run")
	}
	if !strings.Contains(err.Error(), "standby at both sites") {
		t.Errorf("the error must say which way it failed, got: %v", err)
	}
	if len(a.writes) != 0 {
		t.Errorf("nothing may be promoted after a failed demotion, got %v", a.writes)
	}
}

// TestExecuteSwitchPlanStopsOnSetRoleError proves a rejected write stops the run with
// nothing promoted.
func TestExecuteSwitchPlanStopsOnSetRoleError(t *testing.T) {
	chans, a, b := twoMates()
	plan, err := BuildSwitchPlan(ordersToA(), SiteState{siteA, a.state}, SiteState{siteB, b.state})
	if err != nil {
		t.Fatalf("BuildSwitchPlan: %v", err)
	}
	b.setErr = errors.New("invalid command")
	if err := ExecuteSwitchPlan(context.Background(), chans, plan, []string{"ORDERS"}, nil); err == nil {
		t.Fatal("a rejected demotion must stop the run")
	}
	if len(a.writes) != 0 {
		t.Errorf("nothing may be promoted after a rejected demotion, got %v", a.writes)
	}
}

// TestExecuteSwitchPlanRefusesToPromoteAgainstStaleState is the window-closing test.
//
// A concurrent `configure dr`, or an operator at a broker CLI, can move a role between
// the confirmation and the promotion -- and by then the confirmation's reading is stale.
// Promoting anyway is precisely how two actives happen, so the run stops instead.
func TestExecuteSwitchPlanRefusesToPromoteAgainstStaleState(t *testing.T) {
	chans, a, b := twoMates()
	plan, err := BuildSwitchPlan(ordersToA(), SiteState{siteA, a.state}, SiteState{siteB, b.state})
	if err != nil {
		t.Fatalf("BuildSwitchPlan: %v", err)
	}
	// After the confirmation passes, something else promotes ORDERS at B again. The
	// re-read immediately before promoting is what sees it.
	confirmed := false
	b.onRead = func() {
		if confirmed {
			b.state["ORDERS"] = VPNRepl{Admin: AdminEnabled, Role: RoleActive}
		}
		confirmed = true
	}

	err = ExecuteSwitchPlan(context.Background(), chans, plan, []string{"ORDERS"}, nil)
	if err == nil {
		t.Fatal("promoting against state that moved since the confirmation must be refused")
	}
	if !strings.Contains(err.Error(), "since the demotion was confirmed") {
		t.Errorf("the error must explain what changed, got: %v", err)
	}
	if len(a.writes) != 0 {
		t.Errorf("nothing may be promoted, got %v", a.writes)
	}
}

// TestExecuteSwitchPlanEmptyDoesNothing proves a converged pair produces no writes and
// no reads that could fail.
func TestExecuteSwitchPlanEmptyDoesNothing(t *testing.T) {
	a := &scriptedMate{site: siteA, state: map[string]VPNRepl{"ORDERS": {Admin: AdminEnabled, Role: RoleActive}}}
	b := &scriptedMate{site: siteB, state: map[string]VPNRepl{"ORDERS": {Admin: AdminEnabled, Role: RoleStandby}}}
	chans := map[string]MateChannel{siteA: a, siteB: b}
	plan, err := BuildSwitchPlan(ordersToA(), SiteState{siteA, a.state}, SiteState{siteB, b.state})
	if err != nil {
		t.Fatalf("BuildSwitchPlan: %v", err)
	}
	if !plan.Empty() {
		t.Fatalf("plan should be empty, got %+v", plan.Phases)
	}
	if err := ExecuteSwitchPlan(context.Background(), chans, plan, []string{"ORDERS"}, nil); err != nil {
		t.Fatalf("ExecuteSwitchPlan: %v", err)
	}
	if len(a.writes)+len(b.writes) != 0 {
		t.Errorf("an empty plan must write nothing, got %v and %v", a.writes, b.writes)
	}
}

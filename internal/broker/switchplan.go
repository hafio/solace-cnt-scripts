package broker

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"solace/internal/config"
)

// switchplan.go decides and performs a replication role switch.
//
// It is a PLAN rather than a straight sequence of MateChannel calls, and that costs one
// struct. The reason is the parked third mechanism: a pair of generated, self-contained
// scripts that an operator runs by hand where this tool cannot reach. Those are rendered
// OFFLINE, with no broker readable, so they cannot be produced by a procedure that reads
// as it goes -- but they can be rendered from a plan. Building this as a procedure would
// mean rewriting the op to add them later.
//
// The safety property, stated precisely because an over-strong claim is worse than none:
// WITHIN ONE RUN, no promotion begins until every demotion in the working set has been
// confirmed by a read-back AND re-read immediately before promoting. It is a guarantee
// about this run, not about the system -- nothing locks the brokers, so a concurrent
// `configure dr` or an operator at a CLI can still move a role. The re-read narrows that
// window to one read and one write; the verify afterwards catches an arrival that lands
// after it.
//
// `config-sync assert-leader message-vpn <n>` is NOT issued after a switchover, and that
// is settled rather than an oversight (operator-confirmed). It forcibly overwrites the
// other leader's content, so adding it speculatively to "tidy up" after a role change is
// worse than omitting it. `broker perform assert-leader` remains a separate command an
// operator invokes deliberately (scripts.go, assertLeaderScript).

// PhaseKind names one step of a switch. The order they appear in a plan IS the order
// they run; nothing reorders them.
type PhaseKind string

const (
	PhaseDemote  PhaseKind = "demote"
	PhaseConfirm PhaseKind = "confirm"
	PhasePromote PhaseKind = "promote"
	PhaseVerify  PhaseKind = "verify"
)

// SwitchAction is one role write at one site.
type SwitchAction struct {
	Site string // the site's virtualRouterName
	VPN  string
	Role ReplRole
}

// SwitchPhase is a set of actions that may run in any order among themselves, but only
// after every earlier phase has completed.
type SwitchPhase struct {
	Kind    PhaseKind
	Actions []SwitchAction
}

// SwitchPlan is the whole ordered work of a switch.
type SwitchPlan struct {
	Phases []SwitchPhase
	// Skipped names the VPNs already correct at BOTH sites. They are reported rather
	// than silently omitted: an operator who asked for a switchover and got a no-op
	// should be told which VPNs were already where the file wants them.
	Skipped []string
}

// Empty reports whether the plan would change nothing.
//
// A no-op run must not demand an exact-`yes` confirmation -- asking a destructive
// question about a run that destroys nothing teaches an operator to answer it without
// reading, which is the opposite of what the gate is for.
func (p SwitchPlan) Empty() bool {
	for _, ph := range p.Phases {
		if len(ph.Actions) > 0 {
			return false
		}
	}
	return true
}

// Moves returns the VPNs the plan actually moves, sorted.
func (p SwitchPlan) Moves() []string {
	seen := map[string]bool{}
	for _, ph := range p.Phases {
		if ph.Kind != PhasePromote {
			continue
		}
		for _, a := range ph.Actions {
			seen[a.VPN] = true
		}
	}
	out := make([]string, 0, len(seen))
	for v := range seen {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

// SiteState is what one site reports about the VPNs a switch is about.
type SiteState struct {
	Site  string // virtualRouterName
	State map[string]VPNRepl
}

// SwitchPreflight is every check a switch makes BEFORE the first write: both channels
// reachable, the primary HA node active at both sites, and the two brokers naming each
// other as the file says they should.
//
// All of it runs before any mutation, so a pair that fails is a pair nobody has touched.
func SwitchPreflight(ctx context.Context, chans map[string]MateChannel,
	sites map[string]config.ReplSite) error {
	for _, name := range sortedSiteNames(chans) {
		if err := chans[name].Preflight(ctx); err != nil {
			return err
		}
	}
	if err := requirePrimaryActive(ctx, chans); err != nil {
		return err
	}
	return requireVirtualRouterNamesMatch(ctx, chans, sites)
}

// requirePrimaryActive refuses a switch unless the PRIMARY HA node holds activity at
// every site (operator-confirmed).
//
// The replication role itself is config-synced, so a role written to a standby HA node
// still reaches the active one -- this gate is not about where the write lands. It is
// about not moving a DR role while a site is already in an abnormal HA state: a site
// running on its backup has had something go wrong, and stacking a cross-site role change
// on top of a local failover makes both harder to reason about and to undo.
//
// It is also the honest gate for the case that actually bites. When the backup is active
// BECAUSE the primary is down, the primary is the node this tool addresses -- so the run
// would fail anyway, and failing here says why instead of surfacing a transport error.
//
// A standalone broker has no HA group and passes: there is no other node for activity to
// be on.
func requirePrimaryActive(ctx context.Context, chans map[string]MateChannel) error {
	var notActive []string
	for _, name := range sortedSiteNames(chans) {
		ok, err := chans[name].PrimaryActive(ctx)
		if err != nil {
			return err
		}
		if !ok {
			notActive = append(notActive, name)
		}
	}
	if len(notActive) == 0 {
		return nil
	}
	return fmt.Errorf("refusing to switch: the primary HA node is not active at %s. A replication "+
		"switch runs only when each site is on its own primary -- a site running on its backup has "+
		"already had something go wrong, and stacking a DR role change on top of a local failover "+
		"makes both harder to undo. Revert activity to the primary (`broker perform assert-leader`, "+
		"or `redundancy revert-activity` on the backup), then run this again",
		strings.Join(notActive, " and "))
}

// requireVirtualRouterNamesMatch refuses a switch unless each broker names the OTHER
// site's declared virtual-router-name as its replication mate.
//
// This is the file and the brokers agreeing on what the pair IS. A mismatch means one of
// three things, and none of them is safe to switch through: the env file names a site the
// broker has never heard of, `configure data-replication` has not been run since the block
// changed, or -- the one this exists for -- the tool is talking to a node whose virtual
// router is not the one the pair was built around.
func requireVirtualRouterNamesMatch(ctx context.Context, chans map[string]MateChannel,
	sites map[string]config.ReplSite) error {
	names := sortedSiteNames(chans)
	if len(names) != 2 {
		return fmt.Errorf("a replication switch needs exactly 2 sites, got %d", len(names))
	}
	var bad []string
	for i, name := range names {
		other := names[1-i]
		m, err := chans[name].ShowReplication(ctx)
		if err != nil {
			return err
		}
		want := sites[other].VirtualRouterName
		if m.VirtualRouterName != want {
			bad = append(bad, fmt.Sprintf("%s names its mate %q, but the env file says the other "+
				"site is %q", chans[name].Describe(), m.VirtualRouterName, want))
		}
	}
	if len(bad) == 0 {
		return nil
	}
	return fmt.Errorf("refusing to switch: the brokers and the env file disagree about which pair "+
		"this is -- %s. Run `broker configure data-replication` at each site, which is what writes "+
		"the mate's virtual-router-name", strings.Join(bad, "; "))
}

func sortedSiteNames(chans map[string]MateChannel) []string { return sortedKeys(chans) }

// BuildSwitchPlan turns the file's intent plus both sites' current state into an ordered
// plan, or refuses.
//
// Three refusals, each a state this tool did not create and must not quietly repair:
//
//   - a listed VPN missing at a site, or present with replication shut down. The switch
//     only moves roles; enabling replication is `configure dr`'s job, and doing it here
//     would turn a switchover into a configuration change nobody asked for.
//   - a VPN ACTIVE AT BOTH sites. Demoting one picks a winner, and which side keeps its
//     spooled messages is not a decision to make on the operator's behalf.
//   - a site in the state map that the file does not declare, which means the caller and
//     the file disagree about what the pair is.
//
// What it does NOT refuse is a VPN already correct: that is skipped, so a single-VPN
// failover leaves the other VPNs' clients connected.
func BuildSwitchPlan(vpns []config.ReplVPN, a, b SiteState) (SwitchPlan, error) {
	if a.Site == b.Site {
		return SwitchPlan{}, fmt.Errorf("both sides of the switch report the same site %q", a.Site)
	}

	var missing, disabled, splitBrain []string
	var demote, promote []SwitchAction
	var skipped []string

	for _, v := range vpns {
		if v.ActiveAt != a.Site && v.ActiveAt != b.Site {
			return SwitchPlan{}, fmt.Errorf("message-vpn %q is meant to be active at %q, which is "+
				"neither site of this pair (%q and %q)", v.Name, v.ActiveAt, a.Site, b.Site)
		}
		sa, oka := a.State[v.Name]
		sb, okb := b.State[v.Name]
		if !oka {
			missing = append(missing, fmt.Sprintf("%s at %s", v.Name, a.Site))
		}
		if !okb {
			missing = append(missing, fmt.Sprintf("%s at %s", v.Name, b.Site))
		}
		if !oka || !okb {
			continue
		}
		if sa.Admin != AdminEnabled {
			disabled = append(disabled, fmt.Sprintf("%s at %s", v.Name, a.Site))
		}
		if sb.Admin != AdminEnabled {
			disabled = append(disabled, fmt.Sprintf("%s at %s", v.Name, b.Site))
		}
		if sa.Admin != AdminEnabled || sb.Admin != AdminEnabled {
			continue
		}
		if sa.Role == RoleActive && sb.Role == RoleActive {
			splitBrain = append(splitBrain, fmt.Sprintf("%s (active at BOTH %s and %s; the file wants "+
				"it active at %s)", v.Name, a.Site, b.Site, v.ActiveAt))
			continue
		}

		target, other := a, b
		if v.ActiveAt == b.Site {
			target, other = b, a
		}
		if target.State[v.Name].Role == RoleActive && other.State[v.Name].Role == RoleStandby {
			skipped = append(skipped, v.Name)
			continue
		}
		// Demote at the site the file does NOT name -- never at the one about to be
		// promoted, so the target is not taken down on its way up. With two sites that
		// is exactly one demotion per moving VPN. It is issued whether or not that side
		// currently reads active, which keeps the step idempotent: demoting an
		// already-standby VPN is a no-op on the broker.
		demote = append(demote, SwitchAction{Site: other.Site, VPN: v.Name, Role: RoleStandby})
		promote = append(promote, SwitchAction{Site: target.Site, VPN: v.Name, Role: RoleActive})
	}

	if len(missing) > 0 || len(disabled) > 0 {
		return SwitchPlan{}, replPrecheckError(missing, disabled)
	}
	if len(splitBrain) > 0 {
		return SwitchPlan{}, fmt.Errorf("refusing to switch: %s.\n"+
			"This tool did not create that state and will not silently repair it -- demoting one side "+
			"picks a winner, and which site keeps its spooled messages is not a decision to make for "+
			"you. Demote the side you do not want, then run this again",
			strings.Join(splitBrain, "; "))
	}

	sort.Strings(skipped)
	return SwitchPlan{
		Phases: []SwitchPhase{
			{Kind: PhaseDemote, Actions: demote},
			{Kind: PhaseConfirm},
			{Kind: PhasePromote, Actions: promote},
			{Kind: PhaseVerify},
		},
		Skipped: skipped,
	}, nil
}

// replPrecheckError says what is wrong and where to fix it, rather than failing with a
// bare mismatch. Both lists are sorted so the same broker state reads the same way twice.
func replPrecheckError(missing, disabled []string) error {
	sort.Strings(missing)
	sort.Strings(disabled)
	var b strings.Builder
	b.WriteString("refusing to switch: ")
	if len(missing) > 0 {
		fmt.Fprintf(&b, "these message-VPNs are listed in replication.vpns but do not exist: %s",
			strings.Join(missing, ", "))
	}
	if len(missing) > 0 && len(disabled) > 0 {
		b.WriteString("; ")
	}
	if len(disabled) > 0 {
		fmt.Fprintf(&b, "these have replication shut down: %s", strings.Join(disabled, ", "))
	}
	b.WriteString(". A switch only moves roles -- run `broker configure data-replication` at each " +
		"site first, which is what creates and enables them")
	return fmt.Errorf("%s", b.String())
}

// ExecuteSwitchPlan performs a plan against the two channels, keyed by site name.
//
// The ordering rules are enforced HERE rather than trusted to the plan's shape, because
// this is the code that actually writes: a demotion that cannot be confirmed stops the
// run, and a VPN whose state moved between the confirmation and the promotion is left
// alone rather than promoted on stale evidence.
func ExecuteSwitchPlan(ctx context.Context, chans map[string]MateChannel, plan SwitchPlan,
	vpns []string, log func(string, ...any)) error {
	if log == nil {
		log = func(string, ...any) {}
	}
	for _, ph := range plan.Phases {
		switch ph.Kind {
		case PhaseDemote:
			for _, act := range ph.Actions {
				c, ok := chans[act.Site]
				if !ok {
					return fmt.Errorf("no channel for site %q", act.Site)
				}
				log("demoting message-vpn %q at %s", act.VPN, act.Site)
				if err := c.SetRole(ctx, act.VPN, act.Role); err != nil {
					return err
				}
			}
		case PhaseConfirm:
			if err := confirmDemotions(ctx, chans, plan, vpns); err != nil {
				return err
			}
		case PhasePromote:
			if err := promoteConfirmed(ctx, chans, ph.Actions, vpns, log); err != nil {
				return err
			}
		case PhaseVerify:
			// The caller owns the verify-and-report step: it has the output sink and
			// the skipped set, and a mismatch there is a report rather than a retry.
		}
	}
	return nil
}

// confirmDemotions re-reads every demoted VPN and refuses to go on unless it is standby.
//
// Reading back is the point: a command that returned success is not evidence that the
// broker applied it, and `cli -Apes` exits 0 even on a rejected line. Failing here leaves
// the VPN standby at BOTH sites -- unavailable, not corrupted, which is the intended
// direction to fail.
func confirmDemotions(ctx context.Context, chans map[string]MateChannel, plan SwitchPlan,
	vpns []string) error {
	for _, ph := range plan.Phases {
		if ph.Kind != PhaseDemote {
			continue
		}
		bySite := map[string][]string{}
		for _, act := range ph.Actions {
			bySite[act.Site] = append(bySite[act.Site], act.VPN)
		}
		for site, want := range bySite {
			state, err := chans[site].ReadRoles(ctx, vpns)
			if err != nil {
				return fmt.Errorf("confirming demotions at %s: %w", site, err)
			}
			var bad []string
			for _, v := range want {
				if state[v].Role != RoleStandby {
					bad = append(bad, fmt.Sprintf("%s reads %s", v, state[v].Role))
				}
			}
			if len(bad) > 0 {
				sort.Strings(bad)
				return fmt.Errorf("demotion not confirmed at %s: %s. Nothing has been promoted, so "+
					"these message-VPNs are standby at both sites -- unavailable, not corrupted. "+
					"Fix the demotion and run this again", site, strings.Join(bad, ", "))
			}
		}
	}
	return nil
}

// promoteConfirmed re-reads immediately before each promotion and promotes only what is
// still standby everywhere.
//
// This is what closes the gap between the confirmation and the write. A concurrent
// `configure dr` at either site, or an operator at a broker CLI, can move a role in
// between -- and by then the confirmation's reading is stale.
//
// A VPN that moved ABORTS the promotion phase -- it is not promoted, and neither is
// anything after it. Skipping the one VPN and promoting the rest was the other option and
// is worse: something outside this run is writing to the pair, so every remaining reading
// is suspect, and promoting against an unread state is precisely how two actives happen.
// Stopping leaves the run's own error as the only thing to read, and nothing further
// changed -- which is what the message says.
func promoteConfirmed(ctx context.Context, chans map[string]MateChannel,
	actions []SwitchAction, vpns []string, log func(string, ...any)) error {
	fresh := map[string]map[string]VPNRepl{}
	for site, c := range chans {
		st, err := c.ReadRoles(ctx, vpns)
		if err != nil {
			return fmt.Errorf("re-reading %s before promoting: %w", site, err)
		}
		fresh[site] = st
	}
	for _, act := range actions {
		var moved []string
		for site, st := range fresh {
			if st[act.VPN].Role == RoleActive {
				moved = append(moved, site)
			}
		}
		if len(moved) > 0 {
			sort.Strings(moved)
			return fmt.Errorf("message-vpn %q is active at %s again since the demotion was confirmed: "+
				"something outside this run moved it. Refusing to promote it at %s -- that is how two "+
				"actives happen. Nothing further has been changed",
				act.VPN, strings.Join(moved, " and "), act.Site)
		}
		c, ok := chans[act.Site]
		if !ok {
			return fmt.Errorf("no channel for site %q", act.Site)
		}
		log("promoting message-vpn %q at %s", act.VPN, act.Site)
		if err := c.SetRole(ctx, act.VPN, act.Role); err != nil {
			return err
		}
	}
	return nil
}

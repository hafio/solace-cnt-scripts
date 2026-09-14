package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"solace/internal/broker"
	"solace/internal/config"
	"solace/internal/k8s"
	"solace/internal/output"
)

// replication.go is the CLI half of data replication: the two leaves, the gates they
// take, and the site resolution both need.
//
// Which site a broker IS cannot come from the file. The `replication:` block is
// byte-identical at both ends by design -- that is what stops it having to be reversed
// when copied -- so nothing in it says "this one". The BROKER is asked instead, and its
// answer is matched against replication.sites[].routerNames.

// resolveSites asks the broker its router name and returns this site and the mate.
//
// The three non-happy paths are errors rather than guesses (config.Replication.Locate):
// no match names what was read beside every declared name, more than one match is a
// config error naming both sites, and there is no "assume the first" -- a wrong guess
// here points a switchover at the wrong broker.
func resolveSites(a *App, o *broker.Ops, role config.Role) (self, mate config.ReplSite, err error) {
	if !a.Cfg.Replication.Configured() {
		return config.ReplSite{}, config.ReplSite{}, usagef(
			"this env file declares no replication: section, so there is no DR pair to act on")
	}
	name, err := o.RouterName(bg(), role)
	if err != nil {
		return config.ReplSite{}, config.ReplSite{}, err
	}
	step("this broker reports router name %q", name)
	return a.Cfg.Replication.Locate(name)
}

// opConfigureReplication converges THIS broker to the replication block.
//
// Local-only: it never reads and never writes the mate, so it works on every platform
// and cannot be blocked by a WAN outage. The accepted consequence is in its Long text --
// this is the one path in the feature that can produce two actives.
func opConfigureReplication(a *App, o *broker.Ops, role config.Role) error {
	self, mate, err := resolveSites(a, o, role)
	if err != nil {
		return err
	}
	step("this broker is site %s; its mate is %s", self.VirtualRouterName, mate.VirtualRouterName)

	// The gate is asked BEFORE the broker is touched, and names what will change. Three
	// things this applies can interrupt message flow: if the mate configuration differs
	// from the file, EVERY VPN this broker is replicating is stopped first (phase 1,
	// internal/broker/replicationops.go) and only the listed ones are turned back on; a
	// listed VPN itself cycles through a shutdown while its role is applied; and a VPN
	// the file does not list has its replication shut down for good.
	if !confirmReplicationConfig(a, self, mate) {
		return nil
	}

	res, err := o.ConfigureReplication(bg(), role, self, mate, a.Cfg.Replication.VPNs)
	if err != nil {
		return err
	}
	reportReplicationConfig(res)
	return nil
}

// confirmReplicationConfig asks the one question this command gets.
//
// It takes the exact-`yes` gate rather than the lenient one, for the same reason
// import-config's VPN scope does: shutting replication down on a VPN that is carrying
// traffic stops delivery, and a habitual `y` is not the answer that should reach it.
//
// This is asked BEFORE ConfigureReplication reads the broker, so it cannot say whether
// phase 1 (mate convergence) will actually run -- only PlannedRoles, derived from the
// file alone, is known yet. The wording below is therefore conditional rather than a
// promise, and it is the bigger of the two hammers this command can swing: phase 1, when
// it runs, stops replication on EVERY VPN this broker currently replicates, including
// ones the file does not even mention, not only the listed ones named below.
func confirmReplicationConfig(a *App, self, mate config.ReplSite) bool {
	roles := broker.PlannedRoles(a.Cfg.Replication.VPNs, self)
	names := make([]string, 0, len(roles))
	for v := range roles {
		names = append(names, v)
	}
	sort.Strings(names)
	step("if this broker's mate configuration does not already match the env file, " +
		"replication is stopped on EVERY VPN this broker is replicating -- including ones " +
		"this file does not list -- and restored only for the VPNs below")
	for _, v := range names {
		step("  message-vpn %q -> %s", v, roles[v])
	}
	return confirmActionStrict(a, "Apply replication configuration to",
		"apply replication configuration to",
		fmt.Sprintf("this broker (site %s, mate %s)", self.VirtualRouterName, mate.VirtualRouterName))
}

// reportReplicationConfig prints what changed and then the broker's own two reports.
func reportReplicationConfig(res *broker.ReplicationConfigResult) {
	out := output.New(os.Stdout)
	out.Section("Replication configuration applied")
	rows := [][]string{}
	if res.MateApplied {
		if len(res.Removed) > 0 {
			rows = append(rows, []string{"mate addresses cleared", strings.Join(res.Removed, "; ")})
		}
		rows = append(rows, []string{"mate addresses applied", strings.Join(res.Mate, "; ")})
		if len(res.StoppedForMate) > 0 {
			rows = append(rows, []string{"replication stopped to change the mate (see below for what was restored)",
				strings.Join(res.StoppedForMate, ", ")})
		}
	} else {
		// res.Mate/res.Removed are what WOULD have been sent -- printing them here
		// would read as an apply that did not happen. Say plainly that nothing was
		// written instead.
		rows = append(rows, []string{"mate configuration",
			"already matched the env file -- nothing written, no VPN stopped for it"})
	}
	if len(res.Enabled) > 0 {
		rows = append(rows, []string{"replication newly enabled", strings.Join(res.Enabled, ", ")})
	}
	if len(res.Disabled) > 0 {
		rows = append(rows, []string{"replication SHUT DOWN (not in the file)", strings.Join(res.Disabled, ", ")})
	}
	out.Table([]string{"WHAT", "DETAIL"}, rows)
	// The broker's own reports follow a clean apply, and only a clean one -- a failure
	// returns before this, so the error is not buried under a page of output.
	fmt.Fprintln(os.Stdout, res.Status)
}

// opPerformReplication moves roles across BOTH brokers.
//
// It configures nothing: the mate addresses and enablement are `configure dr`'s, and
// this only verifies them. Everything that could refuse does so before the first write.
func opPerformReplication(a *App, o *broker.Ops, role config.Role, mateFor mateChannelFunc) error {
	self, mate, err := resolveSites(a, o, role)
	if err != nil {
		return err
	}
	mateChan, err := mateFor(a, o, mate)
	if err != nil {
		return err
	}
	chans := map[string]broker.MateChannel{
		self.VirtualRouterName: o.NewLocalMate(role),
		mate.VirtualRouterName: mateChan,
	}
	sites := map[string]config.ReplSite{
		self.VirtualRouterName: self,
		mate.VirtualRouterName: mate,
	}

	step("checking both sites before changing anything")
	if err := broker.SwitchPreflight(bg(), chans, sites); err != nil {
		return err
	}

	vpns := replVPNNames(a.Cfg.Replication.VPNs)
	states := map[string]broker.SiteState{}
	for name, c := range chans {
		st, readErr := c.ReadRoles(bg(), vpns)
		if readErr != nil {
			return readErr
		}
		states[name] = broker.SiteState{Site: name, State: st}
	}

	plan, err := broker.BuildSwitchPlan(a.Cfg.Replication.VPNs,
		states[self.VirtualRouterName], states[mate.VirtualRouterName])
	if err != nil {
		return err
	}
	if len(plan.Skipped) > 0 {
		step("already correct at both sites: %s", strings.Join(plan.Skipped, ", "))
	}
	// A no-op asks nothing. Putting a destructive question in front of a run that
	// changes nothing teaches an operator to answer it without reading.
	if plan.Empty() {
		step("nothing to move; the pair already matches the env file")
		return reportReplicationStatus(chans)
	}
	if !confirmReplicationSwitch(a, plan) {
		return nil
	}

	if err := broker.ExecuteSwitchPlan(bg(), chans, plan, vpns, lineSink()); err != nil {
		return err
	}
	return verifyReplicationSwitch(a, chans, plan, vpns)
}

// verifyReplicationSwitch re-reads both sites and makes the end state the exit code.
//
// It checks the SKIPPED set too, not only what moved: a skipped VPN is the run's own
// premise, and a change there means something else is writing while this ran.
func verifyReplicationSwitch(a *App, chans map[string]broker.MateChannel, plan broker.SwitchPlan,
	vpns []string) error {
	final := map[string]map[string]broker.VPNRepl{}
	for name, c := range chans {
		st, err := c.ReadRoles(bg(), vpns)
		if err != nil {
			return err
		}
		final[name] = st
	}
	var wrong []string
	for _, v := range a.Cfg.Replication.VPNs {
		for site, st := range final {
			want := broker.RoleStandby
			if v.ActiveAt == site {
				want = broker.RoleActive
			}
			if st[v.Name].Role != want {
				wrong = append(wrong, fmt.Sprintf("%s at %s reads %s, want %s",
					v.Name, site, st[v.Name].Role, want))
			}
		}
	}
	if err := reportReplicationStatus(chans); err != nil {
		return err
	}
	if len(wrong) > 0 {
		sort.Strings(wrong)
		return fmt.Errorf("the switch did not leave the pair where the env file says it should be: %s",
			strings.Join(wrong, "; "))
	}
	return nil
}

// reportReplicationStatus prints both sites' own `show` output, for display only.
func reportReplicationStatus(chans map[string]broker.MateChannel) error {
	out := output.New(os.Stdout)
	names := make([]string, 0, len(chans))
	for n := range chans {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		text, err := chans[n].Status(bg())
		if err != nil {
			return err
		}
		out.Section("Replication status: " + n)
		fmt.Fprintln(os.Stdout, text)
	}
	return nil
}

// confirmReplicationSwitch names every move before asking. A switch interrupts message
// flow for the VPNs it moves, so it takes the exact-`yes` gate.
func confirmReplicationSwitch(a *App, plan broker.SwitchPlan) bool {
	moves := plan.Moves()
	for _, v := range moves {
		step("  message-vpn %q will be demoted, confirmed, then promoted", v)
	}
	return confirmActionStrict(a, "Move the replication role for", "move the replication role for",
		fmt.Sprintf("%d message-VPN(s): %s", len(moves), strings.Join(moves, ", ")))
}

func replVPNNames(vpns []config.ReplVPN) []string {
	out := make([]string, 0, len(vpns))
	for _, v := range vpns {
		out = append(out, v.Name)
	}
	return out
}

// mateChannelFunc builds the channel to the mate site. It is a parameter so the op is
// testable without a cluster, and so internal/cli owns the platform wiring -- the same
// reason the transports are constructed here rather than inside internal/broker.
type mateChannelFunc func(*App, *broker.Ops, config.ReplSite) (broker.MateChannel, error)

// mateChannel is the production builder. `via` names the mechanism and the block beside
// it carries the settings, so an unset one is refused here rather than at load: a file
// that configures replication perfectly well for the local-only command must still load.
func mateChannel(a *App, o *broker.Ops, site config.ReplSite) (broker.MateChannel, error) {
	switch {
	case site.Via.Kubernetes != nil:
		return k8s.NewMateChannel(a.Runner, a.Cfg, site, lineSink())
	case site.Via.SEMP != nil:
		pass, err := mateSEMPPassword(a, site)
		if err != nil {
			return nil, err
		}
		return o.NewSEMPMate(site, pass)
	default:
		return nil, usagef("replication site %q has no via: block, so this tool cannot reach it. "+
			"A switch has to change the role at BOTH sites, so the mate must be reachable",
			site.VirtualRouterName)
	}
}

// mateSEMPPassword resolves the mate's own admin password.
//
// `pass` and `passEnv` are already resolved by the time a Config is loaded --
// secretref.go's one list covers them, which is why passEnv is not read here. Only
// passSecret needs work, and it is read with the site's OWN cluster CLI when it has one
// so a site reached over SEMP can still keep its password in a cluster this machine can
// see.
//
// It is the MATE's password, never this deployment's: a DR site is a different broker,
// and sending semp.adminPass to it would fail at best and, if the two happened to match,
// hide the misconfiguration until the day they diverged.
func mateSEMPPassword(a *App, site config.ReplSite) (string, error) {
	s := site.Via.SEMP
	if s.Pass != "" {
		return s.Pass, nil
	}
	if s.PassSecret == nil {
		return "", usagef("replication site %q declares via.semp but no password: set pass, point "+
			"passEnv at an environment variable, or name a Kubernetes Secret in passSecret",
			site.VirtualRouterName)
	}
	return k8s.ReadSecretKey(a.Runner, a.Cfg, site, s.PassSecret)
}

func opK8sConfigureReplication(a *App) error {
	role, err := podRole(a)
	if err != nil {
		return err
	}
	return opConfigureReplication(a, k8sOps(a), role)
}

func opCtrConfigureReplication(a *App) error {
	return opConfigureReplication(a, ctrOps(a), config.Primary)
}

func opK8sPerformReplication(a *App) error {
	role, err := podRole(a)
	if err != nil {
		return err
	}
	return opPerformReplication(a, k8sOps(a), role, mateChannel)
}

func opCtrPerformReplication(a *App) error {
	return opPerformReplication(a, ctrOps(a), config.Primary, mateChannel)
}

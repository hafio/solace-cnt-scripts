package broker

import (
	"context"
	"fmt"
	"os"
	"strings"

	"solace/internal/config"
)

// This file holds the HA operations the container platforms drive from the
// primary host, where the transport talks only to the single broker on THIS
// host. What k8s does by addressing either pod from one kubectl context
// (verify_ops.go's Leader/Redundancy, showRDPair), these do from one host: every
// read is the local `show redundancy` -- whose Mate Active / ADB fields already
// report the backup as this node sees it -- and the one command that must land
// on the backup, `redundancy revert-activity`, rides the SEMP control channel
// (semp.go). They reuse the package's unexported helpers (showRD, activity,
// field, rdEnabledUp, primaryRedundancyUp, poll, RunCLI) and script builders
// unchanged.

// LocalRole resolves which redundancy role THIS host plays. An explicit roleArg
// (primary|backup|monitor or p|b|m) wins; otherwise it detects the role by
// matching the host's name against the configured node table (nodes.primary/
// backup/monitor .name), tolerating both an FQDN-vs-short-name mismatch and a
// case difference (DNS names are case-insensitive, so "Broker1-Primary" in the
// env file matches a host reporting "broker1-primary"). It fails loud when
// detection matches nothing, and fails loud -- naming the host and every role
// it matched -- when detection matches more than one, so a mis-targeted HA
// operation never runs silently against the wrong node.
func (o *Ops) LocalRole(roleArg string) (config.Role, error) {
	if roleArg != "" {
		return config.ParseRole(roleArg)
	}
	host, err := o.hostname()
	if err != nil {
		return "", fmt.Errorf("detect node role: read hostname: %w", err)
	}
	var matched []config.Role
	for _, m := range []struct {
		name string
		role config.Role
	}{
		{o.Cfg.Nodes.Primary.Name, config.Primary},
		{o.Cfg.Nodes.Backup.Name, config.Backup},
		{o.Cfg.Nodes.Monitor.Name, config.Monitor},
	} {
		if m.name != "" && hostMatches(host, m.name) {
			matched = append(matched, m.role)
		}
	}
	switch len(matched) {
	case 0:
		return "", fmt.Errorf("cannot determine node role from hostname %q; pass primary|backup|monitor explicitly", host)
	case 1:
		return matched[0], nil
	default:
		names := make([]string, len(matched))
		for i, r := range matched {
			names[i] = roleName(r)
		}
		return "", fmt.Errorf("hostname %q matches more than one configured node role (%s); pass primary|backup|monitor explicitly to disambiguate", host, strings.Join(names, ", "))
	}
}

// LeaderLocal asserts the config-sync leader from THIS host, which must be the
// primary (the user's spec: "assert leader should always be executed in the
// primary node"). HA-only. It fails loud on the backup/monitor rather than
// running, waits for local redundancy to be healthy, then runs assert-leader.
func (o *Ops) LeaderLocal(ctx context.Context, roleArg string) error {
	if o.skipIfStandalone("config leader") {
		return nil
	}
	role, err := o.LocalRole(roleArg)
	if err != nil {
		return err
	}
	if role != config.Primary {
		return fmt.Errorf("config leader must run on the primary node; this host is the %s node", roleName(role))
	}

	// Parity with the k8s Leader: revert any released activity on the mate FIRST
	// (050 lines 23-31), now possible over the SEMP channel. An unreachable mate
	// downgrades to a warning -- the leader assertion itself is local, and a
	// backup still holding activity surfaces in the poll below -- but a reachable
	// mate refusing the RPC is a real error, not a skip.
	if err := o.MateSEMPPreflight(ctx); err != nil {
		o.progress().Warn("cannot reach the mate's SEMP service; skipping the revert-activity step: %v", err)
	} else if err := o.MateRevertActivity(ctx); err != nil {
		return err
	}

	o.logf("Waiting for redundancy state to be restored fully...")
	if err := o.poll(ctx, "redundancy to be restored on Primary", func(ctx context.Context) (bool, error) {
		out, err := o.showRD(ctx, role)
		if err != nil {
			return false, err
		}
		return primaryRedundancyUp(out), nil
	}); err != nil {
		if detail, dErr := o.RunCLI(ctx, role, "show-redundancy-detail", showRedundancyDetailScript()); dErr == nil {
			o.show(detail)
		}
		return err
	}

	out, err := o.RunCLI(ctx, role, "assert-leader", assertLeaderScript())
	if err != nil {
		return err
	}
	o.show([]byte(lastLines(string(out), 12)))
	return nil
}

// RedundancyCoordinated exercises a real failover and fail-back for the whole
// redundancy group from ONE invocation on the primary host, mirroring the k8s
// Redundancy op (verify_ops.go) step for step: confirm the primary healthy,
// release activity so the backup takes over, un-release, then revert the backup
// so activity returns. The backup is never exec'd -- the container transport is
// node-local -- so its takeover and return to standby are observed through the
// primary's own "Mate Active" report (a deliberate weakening of k8s's
// independent backup reads, in all three confirm polls alike), and its one
// mutation, revert-activity, goes over the SEMP channel (semp.go). HA-only;
// backup and monitor hosts are rejected loud. roleArg (empty -> detect from
// hostname) is self-identification only, as in LeaderLocal.
func (o *Ops) RedundancyCoordinated(ctx context.Context, roleArg string) error {
	if o.skipIfStandalone("verify redundancy") {
		return nil
	}
	role, err := o.LocalRole(roleArg)
	if err != nil {
		return err
	}
	if role != config.Primary {
		return fmt.Errorf("verify redundancy drives the whole redundancy group from the primary node; "+
			"this host is the %s node -- run it on the primary host", roleName(role))
	}

	local, err := o.showRD(ctx, role)
	if err != nil {
		return err
	}
	if !primaryRedundancyUp(local) {
		o.show([]byte(local))
		return fmt.Errorf("redundancy configuration/status is not healthy on the Primary")
	}

	// Reach the mate BEFORE the first mutation: failing here leaves the group
	// undisturbed, failing after a release would strand the Backup active.
	if err := o.MateSEMPPreflight(ctx); err != nil {
		return err
	}

	if activity(local, activityLocalActive) == 1 {
		o.progress().Info("Detected Primary node is active.")
		if err := o.releaseLocalActivity(ctx, role); err != nil {
			return err
		}
	} else {
		o.progress().Info("Primary node is standby.")
	}

	// k8s reads the Backup directly here; the primary's mate-activity line is
	// the same fact as this node sees it.
	local, err = o.showRD(ctx, role)
	if err != nil {
		return err
	}
	if activity(local, activityMateActive) != 1 {
		o.show([]byte(local))
		return fmt.Errorf("neither the Primary nor its mate appears to be active")
	}
	o.progress().Info("Detected Backup node is active (Mate Active on the Primary).")

	if err := o.MateRevertActivity(ctx); err != nil {
		return err
	}
	if err := o.poll(ctx, "Primary to become active", func(ctx context.Context) (bool, error) {
		out, err := o.showRD(ctx, role)
		if err != nil {
			return false, err
		}
		return rdEnabledUp(out) &&
			activity(out, activityLocalActive) == 1 &&
			activity(out, activityMateActive) == 0, nil
	}); err != nil {
		return err
	}
	o.progress().Info("Reverted back to Primary node successfully.")
	return nil
}

// releaseLocalActivity walks the local primary through release -> un-release so
// the Backup takes over and the Primary stays eligible to reclaim -- the local
// form of k8s releaseToBackup, with the un-release confirm reading only this
// node (its Mate Active line standing in for the Backup's own Local Active).
func (o *Ops) releaseLocalActivity(ctx context.Context, role config.Role) error {
	if _, err := o.RunCLI(ctx, role, "release", releaseActivityScript()); err != nil {
		return err
	}
	if err := o.poll(ctx, "Primary to be released to the Backup", func(ctx context.Context) (bool, error) {
		out, err := o.showRD(ctx, role)
		if err != nil {
			return false, err
		}
		return field(out, labelConfigStatus) == "Enabled-Released" &&
			field(out, labelRedundancyStatus) == "Down" &&
			activity(out, activityMateActive) == 1, nil
	}); err != nil {
		return err
	}
	o.progress().Info("Primary node is released. Backup node is active.")

	if _, err := o.RunCLI(ctx, role, "no-release", noReleaseActivityScript()); err != nil {
		return err
	}
	if err := o.poll(ctx, "Primary to be un-released", func(ctx context.Context) (bool, error) {
		out, err := o.showRD(ctx, role)
		if err != nil {
			return false, err
		}
		return rdEnabledUp(out) && activity(out, activityMateActive) == 1, nil
	}); err != nil {
		return err
	}
	o.progress().Info("Primary node is un-released. Backup node is active.")
	return nil
}

// hostname reads this host's name via the injected Hostname func, defaulting to
// os.Hostname when unset (New sets it; a directly-constructed Ops in tests injects
// a fixed value). The result is trimmed so a trailing newline never defeats a match.
func (o *Ops) hostname() (string, error) {
	fn := o.Hostname
	if fn == nil {
		fn = os.Hostname
	}
	h, err := fn()
	return strings.TrimSpace(h), err
}

// hostMatches reports whether host names the configured node, tolerating an FQDN
// on either side (pri.example.com matches pri and vice versa) and comparing
// case-insensitively throughout -- DNS names are case-insensitive, so a
// case-sensitive compare here would be a spurious failure rather than a
// meaningful distinction. Both the full-string and short-name compares fold
// case, since either side of either compare can carry an FQDN.
func hostMatches(host, name string) bool {
	return strings.EqualFold(host, name) || strings.EqualFold(shortHost(host), shortHost(name))
}

// shortHost is the label before the first dot of an FQDN (the host itself if none).
func shortHost(h string) string {
	if i := strings.Index(h, "."); i >= 0 {
		return h[:i]
	}
	return h
}

// roleName is the long-form role name for actionable HA guard errors.
func roleName(r config.Role) string {
	switch r {
	case config.Backup:
		return "backup"
	case config.Monitor:
		return "monitor"
	default:
		return "primary"
	}
}

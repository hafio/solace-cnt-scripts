package broker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"solace/internal/config"
)

// Field labels and activity states parsed out of `show redundancy` output. They
// are named once here so assert-leader's check and the redundancy state machines
// compare against a single source of truth -- a typo in a copied literal would
// silently break a state check.
const (
	labelConfigStatus      = "Configuration Status"
	labelRedundancyStatus  = "Redundancy Status"
	labelActiveStandbyRole = "Active-Standby Role"
	labelADBLink           = "ADB Link To Mate"
	labelADBHello          = "ADB Hello To Mate"
	labelActivityStatus    = "Activity Status"

	activityLocalActive = "Local Active"
	activityMateActive  = "Mate Active"

	cliShowRD        = "show-rd"
	cliGatherConfigs = "gather-configs"
)

// Login tests a SEMP login against the node, porting 060. The credentials ride a
// curl config on stdin (curl -K -), so the password never appears in an argv or
// an echoed command. It reports success and writes an outcome line to Out.
func (o *Ops) Login(ctx context.Context, role config.Role, user, pass string) (bool, error) {
	// sempCurl's guard, for the same reason: a line break in a curl -K value ends the
	// directive and starts another. config refuses one in every password it loads, but
	// on Kubernetes the password may instead come from a cluster Secret.
	if strings.ContainsAny(user+pass, "\r\n") {
		return false, fmt.Errorf("the SEMP login credential contains a line break, which a curl config cannot carry")
	}
	// curlConfigLine and defaultSEMPPort, not a hand-written line and a literal: the
	// escaping rule for a curl config value has one definition (semp.go), and so does
	// the broker's own plaintext SEMP port.
	cfg := curlConfigLine("user", user+":"+pass)
	out, err := o.T.OutputInput(ctx, role, []byte(cfg), "curl", "-is", "-K", "-",
		fmt.Sprintf("http://localhost:%d/SEMP/v2/monitor", defaultSEMPPort))
	if err != nil {
		return false, fmt.Errorf("SEMP request failed: %w", err)
	}
	lines := httpStatusLines(string(out))
	for _, l := range lines {
		if isHTTP2xx(l) {
			o.report().OK("Login")
			return true, nil
		}
	}
	status := "<no HTTP response from broker>"
	if len(lines) > 0 {
		status = lines[len(lines)-1]
	}
	o.report().Fail("Login: %s", status)
	return false, nil
}

// Leader asserts role's node as the config-sync leader for the router and every
// message-VPN, porting 050's assert step: LeaderCheck, then LeaderAssert. It is
// for a caller that asks nothing first -- Kubernetes, which always passes the
// primary pod, and a container host that resolved itself as the primary.
//
// It checks ONCE and never moves activity. 050 sent `redundancy revert-activity`
// to the backup and polled until the primary came back; the port of that made a
// rejected revert, or a backup pod that was down, abort the assert, and on
// docker/podman it reached the mate over SEMP rather than the CLI. Now nothing
// but the node being asserted from is touched: a node that does not hold activity
// fails at once, with the transcript and the way to put activity right.
func (o *Ops) Leader(ctx context.Context, role config.Role) error {
	ok, err := o.LeaderCheck(ctx, role)
	if err != nil || !ok {
		return err
	}
	return o.LeaderAssert(ctx, role)
}

// LeaderCheck is assert-leader's one read: it refuses unless role's node reports
// Local Active. It returns false with no error when there is nothing to do -- a
// standalone deployment, which it reports as a skip.
//
// It is separate from LeaderAssert for the reason ImportPlan is separate from
// ImportApply: every confirmation lives in internal/cli (Ops has no Confirm seam),
// so a container host that turns out to be the BACKUP is asked between this read
// and the write, and is not asked at all when it could not assert anyway.
//
// The monitor is refused before anything is read. The role is otherwise the node
// this reads on Kubernetes, and only a label on docker/podman, whose transport
// talks to this host's one broker whatever the role says.
func (o *Ops) LeaderCheck(ctx context.Context, role config.Role) (bool, error) {
	if o.skipIfStandalone("assert-leader") {
		return false, nil
	}
	if role == config.Monitor {
		return false, fmt.Errorf("`broker perform assert-leader` is an invalid operation on the monitor node; " +
			"run it on the primary host")
	}
	// showRD, not RunCLI: a read, for the reason showRD's own comment gives.
	out, err := o.showRD(ctx, role)
	if err != nil {
		return false, err
	}
	// >= 1 rather than Redundancy's == 1: the question is whether this node holds
	// activity at all, as cliMate.PrimaryActive asks it.
	if activity(out, activityLocalActive) == 0 {
		o.show([]byte(out))
		return false, leaderNotActive(role)
	}
	return true, nil
}

// LeaderAssert runs the assert-leader script on role's node and shows the tail of
// what the broker printed (050's `| tail -12`). The caller has run LeaderCheck:
// this does not read the node again.
func (o *Ops) LeaderAssert(ctx context.Context, role config.Role) error {
	o.logf("Asserting the config-sync leader from the %s node: the router, then every message-VPN...", role.Word())
	out, err := o.RunCLI(ctx, role, "assert-leader", assertLeaderScript())
	if err != nil {
		return err
	}
	o.show([]byte(lastLines(string(out), 12)))
	return nil
}

// leaderNotActive is LeaderCheck's refusal, worded for the node it read. Nothing
// waited, so on a primary the likeliest cause right after a deploy is a group that
// is still forming; the other is a backup that took activity, and reverting it is
// the operator's call, typed in the backup's own CLI -- this command no longer
// does it for them.
func leaderNotActive(role config.Role) error {
	if role == config.Backup {
		return fmt.Errorf("this node does not hold activity (its Activity Status is not Local Active), " +
			"so it will not assert the config-sync leader; run `broker perform assert-leader` on the node " +
			"that does, normally the primary host")
	}
	return fmt.Errorf("the %s node does not hold activity (its Activity Status is not Local Active), so the "+
		"config-sync leader was not asserted. If the redundancy group is still forming, wait and run this "+
		"again; if the backup holds activity, revert it first -- `enable`, `admin`, `redundancy revert-activity` "+
		"in the backup's Solace CLI (`broker cli --pod backup` on Kubernetes, `broker cli` on the backup host)",
		role.Word())
}

// Redundancy exercises failover, porting 061: confirm the Primary is active,
// release activity to the Backup, un-release, then revert back to the Primary.
// HA-only: it no-ops for standalone.
func (o *Ops) Redundancy(ctx context.Context) error {
	if o.skipIfStandalone("redundancy-test") {
		return nil
	}

	pri, err := o.showRD(ctx, config.Primary)
	if err != nil {
		return err
	}
	if !primaryRedundancyUp(pri) {
		o.show([]byte(pri))
		return fmt.Errorf("redundancy configuration/status is not healthy on the Primary")
	}

	// If the Primary is active, walk it through release -> un-release so the
	// Backup takes over and hands back cleanly.
	if activity(pri, activityLocalActive) == 1 {
		o.progress().Info("Detected Primary node is active.")
		if err := o.releaseToBackup(ctx); err != nil {
			return err
		}
	}

	// Revert activity from the Backup back to the Primary.
	bk, err := o.showRD(ctx, config.Backup)
	if err != nil {
		return err
	}
	if activity(bk, activityLocalActive) != 1 {
		o.show([]byte(bk))
		return fmt.Errorf("neither Primary nor Backup appears to be active")
	}
	o.progress().Info("Detected Backup node is active.")
	if err := o.revertToPrimary(ctx); err != nil {
		return err
	}
	o.progress().Info("Reverted back to Primary node successfully.")
	return nil
}

// releaseToBackup releases activity from the Primary and un-releases it, leaving
// the Backup active (061 release / no-release steps).
func (o *Ops) releaseToBackup(ctx context.Context) error {
	if _, err := o.RunCLI(ctx, config.Primary, "release", releaseActivityScript()); err != nil {
		return err
	}
	if err := o.poll(ctx, "Primary to be released", func(ctx context.Context) (bool, error) {
		out, err := o.showRD(ctx, config.Primary)
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

	if _, err := o.RunCLI(ctx, config.Primary, "no-release", noReleaseActivityScript()); err != nil {
		return err
	}
	if err := o.poll(ctx, "Primary to be un-released", func(ctx context.Context) (bool, error) {
		p, b, err := o.showRDPair(ctx)
		if err != nil {
			return false, err
		}
		return rdEnabledUp(p) &&
			activity(p, activityMateActive) == 1 &&
			activity(b, activityLocalActive) == 1, nil
	}); err != nil {
		return err
	}
	o.progress().Info("Primary node is un-released. Backup node is active.")
	return nil
}

// revertToPrimary reverts activity from the Backup and waits for the Primary to
// become the sole active node (061 revert step).
func (o *Ops) revertToPrimary(ctx context.Context) error {
	if _, err := o.RunCLI(ctx, config.Backup, "revert-activity", revertActivityConfigureScript()); err != nil {
		return err
	}
	return o.poll(ctx, "Primary to become active", func(ctx context.Context) (bool, error) {
		p, b, err := o.showRDPair(ctx)
		if err != nil {
			return false, err
		}
		return rdEnabledUp(p) &&
			activity(p, activityMateActive) == 0 &&
			activity(p, activityLocalActive) == 1 &&
			activity(b, activityLocalActive) == 0, nil
	})
}

// showRD runs `show redundancy` on role and returns its output as a string.
//
// runCLIRead, not RunCLI: this is a READ -- the poll condition of every
// redundancy wait in this file and verify_local.go, and the one read assert-leader
// makes (LeaderCheck). RunCLI's stop-on-error wrapper would add nothing (there is
// no later line a rejection could poison) and its rejectionIn scan would turn a
// runtime-state word into a hard error -- failKeywords carries the bare word
// "busy", vetted against configuration-capture text and never against `show
// redundancy`, so a mate reported busy mid-restore would abort the whole
// verification instead of polling on. The field() and countContains() scans below
// are also tuned against real unwrapped transcripts.
func (o *Ops) showRD(ctx context.Context, role config.Role) (string, error) {
	out, err := o.readCLI(ctx, role, cliShowRD, showRedundancyScript())
	return string(out), err
}

// backupActivityState reports whether `show redundancy` output from the PRIMARY
// describes a backup that currently holds activity. It is the same reading the
// coordinated failover flows do (activity + activityMateActive), so a caller decides
// "is it safe to send revert-activity to the backup?" with the
// tool's own parser rather than a second, drifting copy of it.
func backupActivityState(showRedundancyOutput string) bool {
	return activity(showRedundancyOutput, activityMateActive) == 1
}

// showRDPair fetches `show redundancy` from both the Primary and Backup, used by
// the cross-node poll conditions.
func (o *Ops) showRDPair(ctx context.Context) (primary, backup string, err error) {
	if primary, err = o.showRD(ctx, config.Primary); err != nil {
		return "", "", err
	}
	backup, err = o.showRD(ctx, config.Backup)
	return primary, backup, err
}

// Diagnostics runs the show-command sweep and gather-diagnostics on each of
// roles, then downloads the zipped output and the diagnostics bundle into
// destDir, porting 069. ts is the timestamp stamped into the local zip name.
func (o *Ops) Diagnostics(ctx context.Context, destDir, ts string, days int, roles ...config.Role) error {
	// Owner-only: the bundles hold the broker's configuration.
	if err := os.MkdirAll(destDir, 0o700); err != nil {
		return fmt.Errorf("create diagnostics dir %q: %w", destDir, err)
	}
	for _, role := range roles {
		if err := o.gatherNode(ctx, role, destDir, ts, days); err != nil {
			return err
		}
	}
	o.logf("Diagnostics written to %s", destDir)
	return nil
}

// gatherNode runs the diagnostics collection for a single node and pulls the
// resulting archives back to destDir.
func (o *Ops) gatherNode(ctx context.Context, role config.Role, destDir, ts string, days int) error {
	o.logf("Gathering diagnostics for %q node...", role)
	zipPath := JailRoot + "/zip-configs.sh"
	if err := o.T.Run(ctx, role, "mkdir", "-p", JailRoot+"/configs/cliout"); err != nil {
		return err
	}
	if err := o.T.Upload(ctx, role, []byte(gatherConfigsScript(days)), cliScriptPath(cliGatherConfigs)); err != nil {
		return err
	}
	if err := o.T.Upload(ctx, role, []byte(zipConfigsScript()), zipPath); err != nil {
		return err
	}
	// Deferred the moment they exist, on every path out. These two were removed inside
	// the "Diagnostics saved" branch below, so a run whose CLI output did not carry that
	// field -- or that failed anywhere after the upload -- left both scripts behind.
	defer func() {
		if err := o.rmPaths(ctx, role, cliScriptPath(cliGatherConfigs), zipPath); err != nil {
			o.progress().Warn("cleanup of the gather scripts on %q failed: %v", role, err)
		}
	}()

	out, err := o.T.Output(ctx, role, CLIBinary, "-Apes", cliArg(cliGatherConfigs))
	if err != nil {
		return err
	}
	if err := o.T.Run(ctx, role, "bash", zipPath); err != nil {
		return err
	}

	hostBytes, err := o.T.Output(ctx, role, "hostname")
	if err != nil {
		return err
	}
	host := strings.TrimSpace(string(hostBytes))
	zipName := fmt.Sprintf("gather-configs-%s-%s.zip", host, ts)
	if err := o.T.Download(ctx, role, JailRoot+"/gather-configs.zip", filepath.Join(destDir, zipName)); err != nil {
		return err
	}
	if err := o.T.Run(ctx, role, "rm", "-rf", JailRoot+"/cli-out", JailRoot+"/gather-configs.zip"); err != nil {
		return err
	}

	if diag := field(string(out), "Diagnostics saved"); diag != "" {
		local := strings.TrimPrefix(diag, "logs/")
		remote := JailRoot + "/" + diag
		// A FAILED download must not be followed by the delete. The bundle on the broker
		// is the only copy, and this command exists to retrieve it -- removing it after
		// the retrieval failed destroys exactly what was asked for, and the old code did
		// that unconditionally while warning, then returned nil so nothing downstream
		// noticed. Now the bundle stays put and the error names where to find it.
		if err := o.T.Download(ctx, role, remote, filepath.Join(destDir, local)); err != nil {
			return fmt.Errorf("download the diagnostics bundle from %q: %w\n"+
				"(it is kept on the broker at %s -- retrieve it by hand before re-running)", role, err, remote)
		}
		if err := o.T.Run(ctx, role, "rm", "-rf", remote); err != nil {
			o.progress().Warn("failed to remove the downloaded diagnostics bundle on %q: %v", role, err)
		}
	}
	return nil
}

// poll runs cond up to PollAttempts times, sleeping PollInterval between tries,
// and returns a timeout error if cond never reports true. It replaces the bash
// scripts' unbounded busy-wait loops with a bounded ceiling.
func (o *Ops) poll(ctx context.Context, desc string, cond func(context.Context) (bool, error)) error {
	for i := 0; i < o.PollAttempts; i++ {
		ok, err := cond(ctx)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		if err := o.sleep(ctx); err != nil {
			return err
		}
	}
	return fmt.Errorf("timeout waiting for %s", desc)
}

// rdEnabledUp reports whether `show redundancy` output shows an enabled, up node.
func rdEnabledUp(out string) bool {
	return field(out, labelConfigStatus) == "Enabled" && field(out, labelRedundancyStatus) == "Up"
}

// activity counts `Activity Status` lines matching a given state (Local/Mate Active).
func activity(out, state string) int { return countContains(out, labelActivityStatus, state) }

// primaryRedundancyUp reports whether `show redundancy` output describes a
// healthy, active Primary, porting the field checks of 050/061.
func primaryRedundancyUp(out string) bool {
	return rdEnabledUp(out) &&
		field(out, labelActiveStandbyRole) == "Primary" &&
		field(out, labelADBLink) == "Up" &&
		field(out, labelADBHello) == "Up"
}

// httpStatusLines returns every "HTTP/..." status line (CR stripped) from a raw
// HTTP response. Login treats the request as successful if any of them is 2xx,
// matching the bash grep-then-regex check in 060, which tolerates a preceding
// "100 Continue" or a redirect line.
func httpStatusLines(out string) []string {
	var lines []string
	for _, raw := range strings.Split(out, "\n") {
		line := strings.TrimRight(raw, "\r")
		if strings.HasPrefix(strings.ToUpper(line), "HTTP/") {
			lines = append(lines, line)
		}
	}
	return lines
}

// isHTTP2xx reports whether an HTTP status line carries a 2xx code, matching the
// `\ 2[0-9][0-9]( |$)` check in 060.
func isHTTP2xx(status string) bool {
	fields := strings.Fields(status)
	if len(fields) < 2 {
		return false
	}
	code := fields[1]
	return len(code) == 3 && code[0] == '2' &&
		code[1] >= '0' && code[1] <= '9' && code[2] >= '0' && code[2] <= '9'
}

// lastLines returns the last n lines of s (its trailing newline preserved),
// porting the `| tail -12` display trim of 050.
func lastLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n") + "\n"
}

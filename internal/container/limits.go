package container

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"solace/internal/config"
)

// Where the two host limits that bind a container actually live. NOT
// /etc/security/limits.conf: that is read by pam_limits, for LOGIN SESSIONS. A
// quadlet container is a systemd service and a docker container is a child of
// dockerd, so neither ever reads it.
const (
	// nrOpenPath is the kernel ceiling on any hard nofile limit. No process is
	// given one above it, however privileged, so it binds on every platform.
	nrOpenPath = "/proc/sys/fs/nr_open"
	// defaultSysctlDropIn is where raising nr_open is made to survive a reboot.
	// Manager.SysctlDropIn carries it so a test can redirect the write.
	defaultSysctlDropIn = "/etc/sysctl.d/99-solace.conf"
	// userUnitDropInFmt caps a ROOTLESS podman: its engine runs under
	// user@<uid>.service with no CAP_SYS_RESOURCE, so that unit's hard limits are
	// the most the container can be given.
	//
	// Scoped to the uid rather than the template unit: a drop-in under
	// user@.service.d would raise the limits of EVERY user on the host, which is
	// more than the broker's account needs and more than an administrator is being
	// asked to agree to.
	userUnitDropInFmt = "/etc/systemd/system/user@%d.service.d/99-solace.conf"
	// pamLimitsDropIn shapes LOGIN SESSIONS. It is not what bounds a quadlet
	// container -- that is the unit above -- but it is what bounds everything an
	// operator runs from their own shell against that container, so the remedy
	// names both rather than leaving the second to be discovered.
	pamLimitsDropIn = "/etc/security/limits.d/99-solace.conf"
)

// delegatedControllers are the cgroup controllers user@<uid>.service must hand to
// the user slice. Without cpuset the rootless quadlet's --cpuset-cpus fails the
// container at start; the rest are what a delegated slice normally carries, and
// asking for them together keeps the drop-in one line rather than a moving target.
var delegatedControllers = []string{"cpu", "cpuset", "io", "memory", "pids"}

// delegateSetting is that list as the drop-in states it.
func delegateSetting() string { return "Delegate=" + strings.Join(delegatedControllers, " ") }

// limitUnbounded is what systemd's "infinity" and the shell's "unlimited" mean.
// Making it the largest int keeps every comparison plain integer arithmetic.
const limitUnbounded = math.MaxInt

// errLimitUnreadable marks a limit this host cannot be ASKED about, as opposed to
// one that answered in a shape this cannot read. Docker Desktop runs the engine in
// a VM, so /proc here belongs to a kernel no container will run under -- refusing
// a healthy host over that would be a defect.
//
// NO answer and a GARBLED answer are deliberately different: an empty stdout is a
// probe that told us nothing, which is the same position as one that could not run,
// while a value that will not parse is an anomaly the operator has to see.
var errLimitUnreadable = errors.New("this host exposes no such limit")

// rootlessPodman is the condition that decides which ceilings bind, which caps the
// quadlet carries, and how the data directory is prepared. Spelled once because
// those answers must never disagree.
func (m *Manager) rootlessPodman() bool {
	return m.P == config.Podman && m.Cfg.Podman.Rootless
}

// checkLimits proves this host can give the broker container the limits both
// artifacts ask for. The ask is not configurable (config.ContainerNoFile* and
// friends are constants the renderers emit verbatim), so there is no env-file
// value to read and no "nothing configured" state to skip on.
//
// Two facts, because two things bind:
//   - fs.nr_open, on every platform. Docker and rootful podman run a privileged
//     engine that raises a container's hard limit past its own, so this kernel
//     ceiling is the only thing left.
//   - user@<uid>.service's hard limits, on rootless podman only, where the engine
//     holds no CAP_SYS_RESOURCE and cannot raise anything.
//
// fix comes from PrepHost. Only nr_open is repairable, and only as root; `validate`
// passes false and writes nothing, as its help promises.
func (m *Manager) checkLimits(ctx context.Context, fix bool) error {
	r := m.report()
	if m.isEcho() {
		r.KVRow(reportKeyWidth, "limits", "skipped (preview)")
		return nil
	}
	if m.Geteuid() < 0 {
		// Windows: no POSIX rlimits, no shell to read them with, and an engine whose
		// kernel is inside a VM this process cannot see.
		r.KVRow(reportKeyWidth, "limits", "skipped (no POSIX rlimits here)")
		return nil
	}
	// The euid invariant checkPodmanHost states: probed as the wrong account the
	// user-manager row answers about a user the deploy will never use. Re-derived
	// rather than remembered -- checkPodmanEUID reads only the config and the euid.
	// checkPodmanHost is what FAILS on the mismatch; this row only stays quiet.
	if m.P == config.Podman && m.checkPodmanEUID() != nil {
		r.Skip("nr_open: skipped (euid mismatch)")
		r.Skip("user manager: skipped (euid mismatch)")
		return nil
	}
	// Resolved once, and only where a row can actually print it: both rootless
	// remedies name the account literally so the drop-in can be forwarded to an
	// administrator as-is. Safe to resolve here rather than leave a placeholder,
	// because the euid guard above has already established that this process is
	// the rootless user rather than root standing in for them.
	user := ""
	if m.rootlessPodman() {
		user = m.hostUserName(ctx, m.Geteuid())
	}
	nrOpen := m.checkNrOpen(ctx, fix)
	userMgr := m.checkUserManagerLimits(ctx, user)
	session := m.checkSessionNoFile(ctx, user)
	// Both rootless rows are fixed by a drop-in, and ONE session restart picks up
	// whichever of them were written -- so it is said once, at the end, rather than
	// twice by two rows that commonly fail together on a stock host. nr_open needs
	// no restart: sysctl is live immediately.
	if err := errors.Join(nrOpen, userMgr, session); err != nil {
		if userMgr != nil || session != nil {
			return fmt.Errorf("%w\n  Then run: sudo loginctl terminate-user %d", err, m.Geteuid())
		}
		return err
	}
	return nil
}

// checkNrOpen asserts the kernel ceiling, and raises it when prep runs as root.
// Unlike a pam_limits change this one is live immediately, so a successful raise
// continues rather than asking for a new session.
func (m *Manager) checkNrOpen(ctx context.Context, fix bool) error {
	r := m.report()
	got, err := m.hostLimit(ctx, "cat "+nrOpenPath)
	if errors.Is(err, errLimitUnreadable) {
		r.Skip("nr_open: skipped (%s is not readable here)", nrOpenPath)
		return nil
	}
	if err != nil {
		r.Fail("nr_open: cannot read %s", nrOpenPath)
		return err
	}
	if got >= config.ContainerNoFileHard {
		r.OK("nr_open: %s (need %d)", limitText(got), config.ContainerNoFileHard)
		return nil
	}
	if fix && m.Geteuid() == 0 {
		if err := m.raiseNrOpen(ctx); err != nil {
			r.Fail("nr_open: %d, and raising it failed", got)
			return fmt.Errorf("%w\n%s", err, m.nrOpenRemedy())
		}
		r.OK("nr_open: raised to %d (was %d)", config.ContainerNoFileHard, got)
		return nil
	}
	r.Fail("nr_open: %d (need %d)", got, config.ContainerNoFileHard)
	return fmt.Errorf("fs.nr_open is %d, needs %d.\n%s",
		got, config.ContainerNoFileHard, m.nrOpenRemedy())
}

// raiseNrOpen applies the value now and writes the drop-in that survives a reboot.
func (m *Manager) raiseNrOpen(ctx context.Context) error {
	setting := fmt.Sprintf("fs.nr_open=%d", config.ContainerNoFileHard)
	if err := m.R.Run(ctx, "sysctl", "-w", setting); err != nil {
		return fmt.Errorf("sysctl -w %s: %w", setting, err)
	}
	body := fmt.Sprintf("# Written by solace-util: the broker container needs nofile %d.\nfs.nr_open = %d\n",
		config.ContainerNoFileHard, config.ContainerNoFileHard)
	if _, err := m.writeArtifact(m.SysctlDropIn, []byte(body), "sysctl drop-in", 0o755); err != nil {
		return err
	}
	return nil
}

func (m *Manager) nrOpenRemedy() string {
	return fmt.Sprintf("  Run: sudo sysctl -w fs.nr_open=%d\n"+
		"  Run: echo 'fs.nr_open = %d' | sudo tee %s",
		config.ContainerNoFileHard, config.ContainerNoFileHard, m.SysctlDropIn)
}

// checkUserManagerLimits is the rootless ceiling. It is never repaired here: the
// drop-in needs root, and a rootless deploy running as root is already refused
// (checkPodmanEUID), so this row can only ever hand over instructions.
func (m *Manager) checkUserManagerLimits(ctx context.Context, user string) error {
	if !m.rootlessPodman() {
		return nil
	}
	r := m.report()
	unit := fmt.Sprintf("user@%d.service", m.Geteuid())
	// m.R.Output directly, NOT m.systemctl: that helper prepends --user when
	// rootless, and this is a question for the SYSTEM manager about a system unit.
	out, err := m.R.Output(ctx, "systemctl", "show", unit,
		"-p", "LimitNOFILE", "-p", "LimitNOFILESoft", "-p", "LimitMEMLOCK", "-p", "LimitCORE",
		"-p", "DelegateControllers")
	if err != nil {
		r.Fail("user manager: cannot read %s limits", unit)
		return fmt.Errorf("cannot read %s limits (`systemctl show %s`): %w", unit, unit, err)
	}
	if len(strings.TrimSpace(string(out))) == 0 {
		// Nothing to read rather than a short limit: same position as a probe that
		// could not run, so it skips rather than refusing a host it cannot judge.
		r.Skip("user manager: skipped (%s reported nothing)", unit)
		return nil
	}
	props := parseUnitProps(out)
	var short []string
	for _, want := range []struct {
		prop string
		min  int
	}{
		{"LimitNOFILE", config.ContainerNoFileHard},
		{"LimitMEMLOCK", limitUnbounded},
		{"LimitCORE", limitUnbounded},
	} {
		raw, ok := props[want.prop]
		if !ok {
			r.Fail("user manager: %s did not report %s", unit, want.prop)
			return fmt.Errorf("`systemctl show %s` did not report %s", unit, want.prop)
		}
		got, parseErr := parseLimit(raw, "systemctl show "+unit)
		if parseErr != nil {
			r.Fail("user manager: %s reported %s=%q", unit, want.prop, raw)
			return parseErr
		}
		if got < want.min {
			short = append(short, fmt.Sprintf("%s is %s, needs %s", want.prop, limitText(got), limitText(want.min)))
		}
	}
	// The cgroup controllers, in the same row and the same drop-in: without cpuset
	// delegated, the rootless quadlet's --cpuset-cpus fails the container at start.
	if missing := missingControllers(props["DelegateControllers"]); len(missing) > 0 {
		short = append(short, fmt.Sprintf("DelegateControllers is missing %s", strings.Join(missing, " ")))
	}
	// The soft limit is reported, never gated: any process raises its own soft limit
	// up to its hard one without privilege, and podman sets the container's from the
	// artifact.
	softText := "unknown"
	if soft, ok := props["LimitNOFILESoft"]; ok {
		softText = soft
	}
	if len(short) == 0 {
		r.OK("user manager: %s limits are sufficient (nofile soft %s)", unit, softText)
		return nil
	}
	r.Fail("user manager: %s", strings.Join(short, "; "))
	return fmt.Errorf("%s: %s.\n"+
		"  Create %s as root:\n"+
		"    [Service]\n"+
		"    %s\n"+
		"    LimitNOFILE=%s\n"+
		"    LimitMEMLOCK=%s\n"+
		"    LimitCORE=%s\n"+
		"  Create %s as root:\n"+
		"    %s hard nofile %d\n"+
		"    %s soft nofile %d\n"+
		"    %s hard memlock unlimited\n"+
		"    %s soft memlock unlimited\n"+
		"  Run: sudo systemctl daemon-reload",
		unit, strings.Join(short, "; "), fmt.Sprintf(userUnitDropInFmt, m.Geteuid()), delegateSetting(),
		config.ContainerNoFile(), config.ContainerLimitMemLock, config.ContainerLimitCore,
		pamLimitsDropIn, user, config.ContainerNoFileHard, user, config.ContainerNoFileSoft,
		user, user)
}

// checkSessionNoFile is the third rootless ceiling, and the one that bounds YOU
// rather than the container: `podman exec`, an interactive `podman run` and the
// admin commands all inherit this login session's rlimits, which pam_limits sets
// from /etc/security/limits.d.
//
// Rootless podman only. Docker and rootful podman start the container from a
// privileged daemon, so the shell that issued the command bounds nothing it does.
//
// Never repaired: the drop-in needs root, and a rootless deploy running as root is
// already refused.
func (m *Manager) checkSessionNoFile(ctx context.Context, user string) error {
	if !m.rootlessPodman() {
		return nil
	}
	r := m.report()
	got, err := m.hostLimit(ctx, "ulimit -Hn")
	if errors.Is(err, errLimitUnreadable) {
		r.Skip("session nofile: skipped (`ulimit -Hn` answered nothing)")
		return nil
	}
	if err != nil {
		r.Fail("session nofile: cannot read this session's hard limit")
		return err
	}
	if got >= config.ContainerNoFileHard {
		r.OK("session nofile: %s (need %d)", limitText(got), config.ContainerNoFileHard)
		return nil
	}
	r.Fail("session nofile: %s (need %d)", limitText(got), config.ContainerNoFileHard)
	return fmt.Errorf("session nofile is %s, needs %d.\n"+
		"  Create %s as root:\n"+
		"    %s hard nofile %d\n"+
		"    %s soft nofile %d",
		limitText(got), config.ContainerNoFileHard, pamLimitsDropIn,
		user, config.ContainerNoFileHard, user, config.ContainerNoFileSoft)
}

// hostLimit reads one number out of a shell probe. Through a shell for the reason
// nearestWritableScript uses one: /proc is a path on the LINUX host, and this tool
// is routinely driven from Windows.
func (m *Manager) hostLimit(ctx context.Context, script string) (int, error) {
	out, err := m.R.Output(ctx, "sh", "-c", script)
	if err != nil {
		return 0, fmt.Errorf("%w (`sh -c %q`): %w", errLimitUnreadable, script, err)
	}
	got := strings.TrimSpace(string(out))
	if got == "" {
		return 0, fmt.Errorf("%w (`sh -c %q` printed nothing)", errLimitUnreadable, script)
	}
	return parseLimit(got, script)
}

// parseUnitProps reads systemd's KEY=value output. A property systemd cannot
// answer is omitted rather than guessed at, so the caller fails loud on it.
func parseUnitProps(out []byte) map[string]string {
	props := map[string]string{}
	for _, line := range strings.Split(string(out), "\n") {
		k, v, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok || v == "" {
			continue
		}
		props[k] = v
	}
	return props
}

// missingControllers reports which of delegatedControllers a DelegateControllers
// value does not name. An empty value means nothing is delegated, so all of them.
func missingControllers(got string) []string {
	have := map[string]bool{}
	for _, c := range strings.Fields(got) {
		have[c] = true
	}
	var missing []string
	for _, c := range delegatedControllers {
		if !have[c] {
			missing = append(missing, c)
		}
	}
	return missing
}

// parseLimit reads one limit value: a decimal, or systemd's "infinity" and the
// shell's "unlimited" for an absent ceiling.
func parseLimit(v, source string) (int, error) {
	switch v {
	case "infinity", "unlimited":
		return limitUnbounded, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("cannot parse the limit %q from %s: %w", v, source, err)
	}
	return n, nil
}

// limitText spells the sentinel back out as the word, so a message never carries
// an implausible number.
func limitText(v int) string {
	if v == limitUnbounded {
		return "infinity"
	}
	return strconv.Itoa(v)
}

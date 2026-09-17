package container

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"solace/internal/config"
	"solace/internal/render"
)

// The healthy answers every other test's fixture leans on. nr_open at the kernel
// default satisfies the ask exactly, which is not a coincidence: fs.nr_open and
// config.ContainerNoFileHard are both 1048576.
//
// The user manager reports a LOW SOFT limit on purpose. It is healthy: any process
// raises its own soft limit up to its hard one without privilege, so only the hard
// values can bind, and this fixture is what keeps that true in the tests.
const (
	healthyNrOpen           = "1048576\n"
	healthySessionNoFile    = "1048576\n"
	healthyUserManagerProps = "LimitNOFILE=1048576\nLimitNOFILESoft=1024\n" +
		"LimitMEMLOCK=infinity\nLimitCORE=infinity\n" +
		"DelegateControllers=cpu cpuset io memory pids\n"
)

// limitsMgr builds a Manager for platform p whose limit probes answer nrOpen and
// whose every other probe answers healthily, so these tests fail on the host
// ceilings alone. euid follows the platform's own invariant: rootful podman
// requires root, rootless refuses it.
//
// SysctlDropIn is redirected into a temp directory. The default is /etc, and a
// test that raised nr_open would otherwise write to the machine running it.
func limitsMgr(t *testing.T, p config.Platform, rootless bool, nrOpen string) (*Manager, *capRunner, *bytes.Buffer) {
	t.Helper()
	cfg := ctrCfg(p, "false") // standalone -> the PSK step is skipped
	if p == config.Podman {
		cfg.Podman.Rootless = rootless
		// Both knobs ApplyDefaults derives from Rootless; ctrCfg skips ApplyDefaults.
		if rootless {
			cfg.Podman.SystemctlUser = "--user"
			cfg.Podman.WantedBy = "default.target"
		}
	}
	m, rr, buf := newCapMgr(cfg, p)
	euid := 1000
	if p == config.Podman && !rootless {
		euid = 0
	}
	m.Geteuid = func() int { return euid }
	m.SysctlDropIn = filepath.Join(t.TempDir(), "99-solace.conf")
	rr.outFor = healthyRootlessOut(nrOpen)
	fakeEnv(m, nil)
	return m, rr, buf
}

func nrOpenProbe() (string, []string) { return "sh", []string{"-c", "cat " + nrOpenPath} }

// --- fs.nr_open, every target -----------------------------------------------

// TestPrepHostChecksNrOpenOnEveryTarget is the behaviour change: the kernel
// ceiling binds whichever engine starts the container, so it is checked on all
// three targets rather than on rootless podman alone. A stock host passes, which
// is the outcome that matters -- fs.nr_open defaults to exactly the ask.
func TestPrepHostChecksNrOpenOnEveryTarget(t *testing.T) {
	for _, tc := range []struct {
		name     string
		p        config.Platform
		rootless bool
	}{
		{"docker", config.Docker, false},
		{"rootful podman", config.Podman, false},
		{"rootless podman", config.Podman, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, rr, buf := limitsMgr(t, tc.p, tc.rootless, healthyNrOpen)
			if err := m.PrepHost(context.Background()); err != nil {
				t.Fatalf("a stock host must pass: %v", err)
			}
			name, args := nrOpenProbe()
			if !hasCall(rr, name, args) {
				t.Errorf("prep should read the kernel nofile ceiling:\n%+v", rr.calls)
			}
			if !strings.Contains(buf.String(), "nr_open: 1048576") {
				t.Errorf("prep should report the ceiling it found:\n%s", buf)
			}
		})
	}
}

// TestPrepHostNrOpenTooLowRefusesAsNonRoot: a non-root caller cannot sysctl, so
// the row hands over both commands -- the live one and the one that survives a
// reboot -- and changes nothing.
func TestPrepHostNrOpenTooLowRefusesAsNonRoot(t *testing.T) {
	m, rr, _ := limitsMgr(t, config.Docker, false, "65536\n")
	err := m.PrepHost(context.Background())
	if err == nil {
		t.Fatal("a kernel ceiling below the ask must fail prep")
	}
	for _, want := range []string{"65536", "1048576", "Run: sudo sysctl -w fs.nr_open", m.SysctlDropIn} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal should name %q, got: %v", want, err)
		}
	}
	// pam_limits shapes login sessions; a container is not one. Naming that file
	// would send an operator to change something that cannot affect the container.
	if strings.Contains(err.Error(), "limits.conf") || strings.Contains(err.Error(), "limits.d") {
		t.Errorf("the refusal must not point at pam_limits: %v", err)
	}
	if hasCall(rr, "sysctl", []string{"-w", "fs.nr_open=1048576"}) {
		t.Errorf("a non-root caller must not try to sysctl:\n%+v", rr.calls)
	}
	if _, statErr := os.Stat(m.SysctlDropIn); statErr == nil {
		t.Error("a non-root caller must not write the drop-in")
	}
}

// TestPrepHostNrOpenRaisedAsRoot: raising nr_open is live immediately, unlike a
// pam_limits change, so prep applies it, persists it, and CONTINUES.
func TestPrepHostNrOpenRaisedAsRoot(t *testing.T) {
	m, rr, buf := limitsMgr(t, config.Podman, false, "65536\n") // rootful podman -> euid 0
	if err := m.PrepHost(context.Background()); err != nil {
		t.Fatalf("root can fix this, so prep must continue: %v", err)
	}
	if !hasCall(rr, "sysctl", []string{"-w", "fs.nr_open=1048576"}) {
		t.Errorf("root should raise the ceiling:\n%+v", rr.calls)
	}
	body, err := os.ReadFile(m.SysctlDropIn)
	if err != nil {
		t.Fatalf("the change must survive a reboot: %v", err)
	}
	if !strings.Contains(string(body), "fs.nr_open = 1048576") {
		t.Errorf("drop-in = %q", body)
	}
	if !strings.Contains(buf.String(), "raised to 1048576") {
		t.Errorf("prep should say what it changed:\n%s", buf)
	}
}

// TestPrepHostNrOpenRaiseFailureRefuses: a failed repair falls back to the
// instructions rather than reporting a success the kernel did not grant.
func TestPrepHostNrOpenRaiseFailureRefuses(t *testing.T) {
	m, rr, _ := limitsMgr(t, config.Podman, false, "65536\n")
	rr.fail = func(name string, args []string) error {
		if name == "sysctl" {
			return errors.New("read-only kernel")
		}
		return nil
	}
	err := m.PrepHost(context.Background())
	if err == nil {
		t.Fatal("a failed raise must not pass")
	}
	for _, want := range []string{"read-only kernel", "sysctl -w fs.nr_open"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal should carry %q, got: %v", want, err)
		}
	}
}

// TestPrepHostNrOpenDropInWriteFailureRefuses: the sysctl took effect but the
// change would not survive a reboot, which is not a success to report. The
// drop-in path is pointed inside a regular FILE, so MkdirAll fails the way an
// unwritable /etc would.
func TestPrepHostNrOpenDropInWriteFailureRefuses(t *testing.T) {
	m, _, _ := limitsMgr(t, config.Podman, false, "65536\n") // rootful podman -> euid 0
	blocker := filepath.Join(filepath.Dir(m.SysctlDropIn), "not-a-dir")
	if err := os.WriteFile(blocker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	m.SysctlDropIn = filepath.Join(blocker, "99-solace.conf")
	err := m.PrepHost(context.Background())
	if err == nil {
		t.Fatal("a drop-in that cannot be written must not report success")
	}
	if !strings.Contains(err.Error(), "sysctl -w fs.nr_open") {
		t.Errorf("the refusal should still hand over the remedy, got: %v", err)
	}
}

// TestCheckLimitsRootlessUserManagerUnreadableFailsLoud: a systemctl that cannot
// be reached at all is an anomaly, unlike one that answers with nothing.
func TestCheckLimitsRootlessUserManagerUnreadableFailsLoud(t *testing.T) {
	m, rr, _ := limitsMgr(t, config.Podman, true, healthyNrOpen)
	failOnCall(rr, func(name string, args []string) bool {
		return name == "systemctl" && slices.Contains(args, "LimitMEMLOCK")
	})
	err := m.checkLimits(context.Background(), false)
	if err == nil || !strings.Contains(err.Error(), "cannot read user@1000.service limits") {
		t.Errorf("an unreachable user manager must fail loud naming the unit, got: %v", err)
	}
}

// TestPrepHostNrOpenUnreadableSkips guards the false refusal: Docker Desktop runs
// the engine in a VM, so /proc here belongs to a kernel no container will run
// under. There is genuinely nothing to assert, and refusing a healthy host over a
// missing file would be the defect.
func TestPrepHostNrOpenUnreadableSkips(t *testing.T) {
	m, rr, buf := limitsMgr(t, config.Docker, false, healthyNrOpen)
	failOnCall(rr, func(name string, args []string) bool {
		return name == "sh" && slices.Contains(args, "cat "+nrOpenPath)
	})
	if err := m.PrepHost(context.Background()); err != nil {
		t.Fatalf("an unreadable ceiling must skip, not refuse: %v", err)
	}
	if !strings.Contains(buf.String(), "nr_open: skipped") {
		t.Errorf("the skip must be said out loud:\n%s", buf)
	}
}

// TestPrepHostNrOpenUnparseableFailsLoud: a /proc that answers in a shape this
// cannot read is an anomaly, not a ceiling to be assumed adequate.
func TestPrepHostNrOpenUnparseableFailsLoud(t *testing.T) {
	m, _, _ := limitsMgr(t, config.Docker, false, "not-a-number\n")
	if err := m.PrepHost(context.Background()); err == nil || !strings.Contains(err.Error(), "cannot parse") {
		t.Errorf("an unreadable value must fail loud, got: %v", err)
	}
}

// TestPrepHostNrOpenEmptyAnswerSkips is the other half of that: NO answer is not a
// garbled one. A probe that exits 0 and prints nothing has told us nothing, which
// is the same position as one that could not run, so it skips. This is also what
// keeps every test in this package that is about something else from having to
// seed a limits answer.
func TestPrepHostNrOpenEmptyAnswerSkips(t *testing.T) {
	m, _, buf := limitsMgr(t, config.Docker, false, "")
	if err := m.PrepHost(context.Background()); err != nil {
		t.Fatalf("an empty answer must skip, not refuse: %v", err)
	}
	if !strings.Contains(buf.String(), "nr_open: skipped") {
		t.Errorf("the skip must be said out loud:\n%s", buf)
	}
}

// TestCheckLimitsRootlessEmptyUserManagerAnswerSkips: same rule for the other
// probe. A systemctl that reports nothing cannot be judged; a systemctl that
// reports SOME properties but not the ones asked for still fails loud.
func TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(t *testing.T) {
	m, rr, buf := limitsMgr(t, config.Podman, true, healthyNrOpen)
	healthy := rr.outFor
	rr.outFor = func(name string, args []string) []byte {
		if name == "systemctl" && slices.Contains(args, "LimitMEMLOCK") {
			return nil
		}
		return healthy(name, args)
	}
	if err := m.checkLimits(context.Background(), false); err != nil {
		t.Fatalf("an empty answer must skip, not refuse: %v", err)
	}
	if !strings.Contains(buf.String(), "user manager: skipped") {
		t.Errorf("the skip must be said out loud:\n%s", buf)
	}

	m2, rr2, _ := limitsMgr(t, config.Podman, true, healthyNrOpen)
	healthy2 := rr2.outFor
	rr2.outFor = func(name string, args []string) []byte {
		if name == "systemctl" && slices.Contains(args, "LimitMEMLOCK") {
			return []byte("LimitNOFILESoft=1024\n")
		}
		return healthy2(name, args)
	}
	err := m2.checkLimits(context.Background(), false)
	if err == nil || !strings.Contains(err.Error(), "LimitNOFILE") {
		t.Errorf("a partial answer must fail loud naming the property, got: %v", err)
	}
}

// TestValidateNeverWritesTheSysctlDropIn: `validate` promises a run that disturbs
// nothing, so it reports the same shortfall and repairs none of it -- even as root.
func TestValidateNeverWritesTheSysctlDropIn(t *testing.T) {
	m, rr, _ := limitsMgr(t, config.Podman, false, "65536\n") // euid 0
	if err := m.Check(context.Background()); err == nil {
		t.Fatal("Check must still report the shortfall")
	}
	if hasCall(rr, "sysctl", []string{"-w", "fs.nr_open=1048576"}) {
		t.Errorf("validate must not repair anything:\n%+v", rr.calls)
	}
	if _, err := os.Stat(m.SysctlDropIn); err == nil {
		t.Error("validate must not write the drop-in")
	}
}

// --- user@<uid>.service, rootless only ---------------------------------------

// TestCheckLimitsRootlessRefusesAShortUserManager is the rootless ceiling: the
// engine holds no CAP_SYS_RESOURCE, so user@<uid>.service's HARD limits are the
// most the container can be given. 8 MB is the stock LimitMEMLOCK, so this is the
// refusal a default host actually gets.
func TestCheckLimitsRootlessRefusesAShortUserManager(t *testing.T) {
	m, rr, buf := limitsMgr(t, config.Podman, true, healthyNrOpen)
	healthy := rr.outFor
	rr.outFor = func(name string, args []string) []byte {
		if name == "systemctl" && slices.Contains(args, "LimitMEMLOCK") {
			return []byte("LimitNOFILE=1048576\nLimitNOFILESoft=1024\n" +
				"LimitMEMLOCK=8388608\nLimitCORE=infinity\n" +
				"DelegateControllers=cpu cpuset io memory pids\n")
		}
		return healthy(name, args)
	}
	err := m.checkLimits(context.Background(), true)
	if err == nil {
		t.Fatal("a user manager that cannot grant the limit must fail")
	}
	// BOTH drop-ins: the unit bounds the container, and pam_limits bounds what the
	// operator runs against it from a login shell. Naming only the first is how an
	// operator ends up discovering the second the hard way.
	for _, want := range []string{
		"LimitMEMLOCK", "8388608", "[Service]",
		"LimitMEMLOCK=infinity", pamLimitsDropIn,
		"/etc/systemd/system/user@1000.service.d/99-solace.conf",
		"solace hard memlock unlimited", "solace hard nofile 1048576",
		"systemctl daemon-reload", "loginctl terminate-user 1000",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal should name %q, got: %v", want, err)
		}
	}
	if !strings.Contains(buf.String(), "user manager:") {
		t.Errorf("the row must report:\n%s", buf)
	}
	// Never repaired, whatever fix says: the drop-in needs root, and a rootless
	// deploy running as root is already refused.
	if hasCall(rr, "systemctl", []string{"daemon-reload"}) {
		t.Errorf("this row must only ever hand over instructions:\n%+v", rr.calls)
	}
}

// TestCheckLimitsRootlessRefusesUndelegatedControllers: the rootless quadlet now
// carries a cpuset, and the cpuset controller is not delegated to a user slice by
// default -- so an undelegated host would fail the container at start. It is the
// same row and the same drop-in as the rlimits, because it is the same file.
func TestCheckLimitsRootlessRefusesUndelegatedControllers(t *testing.T) {
	m, rr, buf := limitsMgr(t, config.Podman, true, healthyNrOpen)
	healthy := rr.outFor
	rr.outFor = func(name string, args []string) []byte {
		if name == "systemctl" && slices.Contains(args, "LimitMEMLOCK") {
			return []byte("LimitNOFILE=1048576\nLimitNOFILESoft=1024\n" +
				"LimitMEMLOCK=infinity\nLimitCORE=infinity\n" +
				"DelegateControllers=memory pids\n")
		}
		return healthy(name, args)
	}
	err := m.checkLimits(context.Background(), false)
	if err == nil {
		t.Fatal("a slice without the cpuset controller must refuse")
	}
	for _, want := range []string{"DelegateControllers is missing", "cpu", "cpuset", "io",
		"Delegate=cpu cpuset io memory pids"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal should name %q, got: %v", want, err)
		}
	}
	if !strings.Contains(buf.String(), "DelegateControllers is missing") {
		t.Errorf("the row must report it:\n%s", buf)
	}
}

// TestMissingControllers is the set arithmetic on its own: an empty value means
// nothing is delegated, and order does not matter.
func TestMissingControllers(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want int
	}{
		{"cpu cpuset io memory pids", 0},
		{"pids memory io cpuset cpu", 0},
		{"cpu cpuset io memory pids extra", 0},
		{"memory pids", 3},
		{"", 5},
	} {
		if got := len(missingControllers(tc.in)); got != tc.want {
			t.Errorf("missingControllers(%q) = %d missing, want %d", tc.in, got, tc.want)
		}
	}
}

// TestCheckLimitsRootlessReportsTheSoftLimitWithoutGating: only a hard limit can
// bind. Any process raises its own soft limit up to its hard one without
// privilege, and podman sets the container's from the artifact.
func TestCheckLimitsRootlessReportsTheSoftLimitWithoutGating(t *testing.T) {
	m, _, buf := limitsMgr(t, config.Podman, true, healthyNrOpen)
	if err := m.checkLimits(context.Background(), false); err != nil {
		t.Fatalf("a low soft limit must not refuse: %v", err)
	}
	if !strings.Contains(buf.String(), "nofile soft 1024") {
		t.Errorf("the soft limit must still be reported:\n%s", buf)
	}
}

// TestCheckLimitsPrivilegedSkipsTheUserManager: docker and rootful podman run a
// privileged engine, so there is no user manager in the path to ask about.
func TestCheckLimitsPrivilegedSkipsTheUserManager(t *testing.T) {
	for _, tc := range []struct {
		name string
		p    config.Platform
	}{{"docker", config.Docker}, {"rootful podman", config.Podman}} {
		t.Run(tc.name, func(t *testing.T) {
			m, rr, _ := limitsMgr(t, tc.p, false, healthyNrOpen)
			if err := m.checkLimits(context.Background(), false); err != nil {
				t.Fatalf("checkLimits: %v", err)
			}
			for _, c := range rr.calls {
				if c.name == "systemctl" && slices.Contains(c.args, "LimitMEMLOCK") {
					t.Errorf("a privileged engine has no user manager to ask about:\n%+v", rr.calls)
				}
			}
		})
	}
}

// TestCheckLimitsRootlessSessionNoFile is the third rootless ceiling, and the one
// that bounds the OPERATOR rather than the container: `podman exec` and the admin
// commands inherit this session's rlimits, which pam_limits sets. A stock host is
// at 524288, so this is the row a fresh install actually fails.
func TestCheckLimitsRootlessSessionNoFile(t *testing.T) {
	m, rr, buf := limitsMgr(t, config.Podman, true, healthyNrOpen)
	healthy := rr.outFor
	rr.outFor = func(name string, args []string) []byte {
		if name == "sh" && slices.Contains(args, "ulimit -Hn") {
			return []byte("524288\n")
		}
		return healthy(name, args)
	}
	err := m.checkLimits(context.Background(), true)
	if err == nil {
		t.Fatal("a session limit below the ask must refuse")
	}
	for _, want := range []string{
		"524288", "1048576", pamLimitsDropIn,
		"solace hard nofile 1048576", "solace soft nofile 2448",
		"loginctl terminate-user 1000",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal should name %q, got: %v", want, err)
		}
	}
	if !strings.Contains(buf.String(), "session nofile: 524288") {
		t.Errorf("the row must report what it found:\n%s", buf)
	}
	// Never repaired: the drop-in needs root, and a rootless deploy as root is
	// already refused, so nothing may touch that path.
	for _, c := range rr.calls {
		if slices.Contains(c.args, pamLimitsDropIn) {
			t.Errorf("this row only ever hands over instructions: %s %v", c.name, c.args)
		}
	}
}

// TestCheckLimitsRootlessUnparseableLimitFailsLoud: a property that IS reported
// but in a shape this cannot read is an anomaly, not a limit to assume adequate.
func TestCheckLimitsRootlessUnparseableLimitFailsLoud(t *testing.T) {
	m, rr, _ := limitsMgr(t, config.Podman, true, healthyNrOpen)
	healthy := rr.outFor
	rr.outFor = func(name string, args []string) []byte {
		if name == "systemctl" && slices.Contains(args, "LimitMEMLOCK") {
			return []byte("LimitNOFILE=bogus\nLimitNOFILESoft=1024\n" +
				"LimitMEMLOCK=infinity\nLimitCORE=infinity\n" +
				"DelegateControllers=cpu cpuset io memory pids\n")
		}
		return healthy(name, args)
	}
	err := m.checkLimits(context.Background(), false)
	if err == nil || !strings.Contains(err.Error(), `cannot parse the limit "bogus"`) {
		t.Errorf("a garbled property must fail loud naming the value, got: %v", err)
	}
}

// TestCheckLimitsSessionNoFileUnreadable covers the session row's two non-answers,
// which split the same way every other probe does: nothing at all is a skip, a
// value that will not parse is a refusal.
func TestCheckLimitsSessionNoFileUnreadable(t *testing.T) {
	for _, tc := range []struct {
		name, answer string
		wantErr      string
	}{
		{"empty", "", ""},
		{"garbled", "not-a-number\n", "cannot parse the limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, rr, buf := limitsMgr(t, config.Podman, true, healthyNrOpen)
			healthy := rr.outFor
			rr.outFor = func(name string, args []string) []byte {
				if name == "sh" && slices.Contains(args, "ulimit -Hn") {
					return []byte(tc.answer)
				}
				return healthy(name, args)
			}
			err := m.checkLimits(context.Background(), false)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("an empty answer must skip, not refuse: %v", err)
				}
				if !strings.Contains(buf.String(), "session nofile: skipped") {
					t.Errorf("the skip must be said out loud:\n%s", buf)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("a garbled answer must fail loud, got: %v", err)
			}
		})
	}
}

// TestCheckLimitsSaysRestartTheSessionOnce: a stock rootless host fails BOTH
// drop-in rows, and one session restart picks up whichever were written -- so the
// instruction belongs at the end of the pass, not on each row that needs it.
func TestCheckLimitsSaysRestartTheSessionOnce(t *testing.T) {
	m, rr, _ := limitsMgr(t, config.Podman, true, healthyNrOpen)
	healthy := rr.outFor
	rr.outFor = func(name string, args []string) []byte {
		if name == "systemctl" && slices.Contains(args, "LimitMEMLOCK") {
			return []byte("LimitNOFILE=1048576\nLimitNOFILESoft=1024\n" +
				"LimitMEMLOCK=8388608\nLimitCORE=infinity\n" +
				"DelegateControllers=cpu cpuset io memory pids\n")
		}
		if name == "sh" && slices.Contains(args, "ulimit -Hn") {
			return []byte("524288\n")
		}
		return healthy(name, args)
	}
	err := m.checkLimits(context.Background(), true)
	if err == nil {
		t.Fatal("both drop-in rows are short, so this must refuse")
	}
	// Both causes survive the join, and the restart is stated exactly once.
	for _, want := range []string{"LimitMEMLOCK is 8388608", "session nofile is 524288"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the joined error should carry %q, got: %v", want, err)
		}
	}
	if n := strings.Count(err.Error(), "loginctl terminate-user"); n != 1 {
		t.Errorf("the session restart must be said once, not %d times:\n%v", n, err)
	}
}

// TestCheckLimitsNrOpenAloneNeedsNoRestart: sysctl is live immediately, so the
// nr_open row must not drag the session-restart line in behind it.
func TestCheckLimitsNrOpenAloneNeedsNoRestart(t *testing.T) {
	m, _, _ := limitsMgr(t, config.Docker, false, "65536\n")
	err := m.checkLimits(context.Background(), false)
	if err == nil {
		t.Fatal("a short nr_open must refuse")
	}
	if strings.Contains(err.Error(), "terminate-user") {
		t.Errorf("raising fs.nr_open needs no session restart: %v", err)
	}
}

// TestCheckLimitsSessionNoFileIsRootlessOnly: docker and rootful podman start the
// container from a privileged daemon, so the shell that issued the command bounds
// nothing it does and the probe must not run.
func TestCheckLimitsSessionNoFileIsRootlessOnly(t *testing.T) {
	for _, tc := range []struct {
		name string
		p    config.Platform
	}{{"docker", config.Docker}, {"rootful podman", config.Podman}} {
		t.Run(tc.name, func(t *testing.T) {
			m, rr, _ := limitsMgr(t, tc.p, false, healthyNrOpen)
			if err := m.checkLimits(context.Background(), false); err != nil {
				t.Fatalf("checkLimits: %v", err)
			}
			if hasCall(rr, "sh", []string{"-c", "ulimit -Hn"}) {
				t.Errorf("a privileged daemon is not bounded by this shell:\n%+v", rr.calls)
			}
		})
	}
}

// --- gating ------------------------------------------------------------------

// TestCheckLimitsSkipsWithoutPosixRlimits: Windows has no POSIX rlimits, no shell
// to read them with, and an engine whose kernel is inside a VM.
func TestCheckLimitsSkipsWithoutPosixRlimits(t *testing.T) {
	m, rr, buf := limitsMgr(t, config.Docker, false, healthyNrOpen)
	m.Geteuid = func() int { return -1 }
	if err := m.checkLimits(context.Background(), true); err != nil {
		t.Fatalf("checkLimits: %v", err)
	}
	if !strings.Contains(buf.String(), "no POSIX rlimits") {
		t.Errorf("the skip must be said out loud:\n%s", buf)
	}
	if len(rr.calls) != 0 {
		t.Errorf("nothing may be probed:\n%+v", rr.calls)
	}
}

// TestCheckLimitsSkipsOnAnEUIDMismatch: probed as root, the rootless rows would
// answer about an account the deploy will never use. checkPodmanHost is what fails
// on the mismatch; this block only declines to answer.
func TestCheckLimitsSkipsOnAnEUIDMismatch(t *testing.T) {
	m, rr, buf := limitsMgr(t, config.Podman, true, healthyNrOpen)
	m.Geteuid = func() int { return 0 } // rootless, but running as root
	if err := m.checkLimits(context.Background(), true); err != nil {
		t.Fatalf("this block reports the skip and leaves the failing to the euid row: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "nr_open: skipped (euid mismatch)") ||
		!strings.Contains(out, "user manager: skipped (euid mismatch)") {
		t.Errorf("both rows must skip with the reason:\n%s", out)
	}
	if len(rr.calls) != 0 {
		t.Errorf("nothing may be probed:\n%+v", rr.calls)
	}
}

// --- parsing and the shared definition ---------------------------------------

func TestParseLimit(t *testing.T) {
	for _, tc := range []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"1048576", 1048576, false},
		{"0", 0, false},
		{"infinity", limitUnbounded, false},  // systemd's spelling
		{"unlimited", limitUnbounded, false}, // the shell's
		{"", 0, true},
		{"not-a-number", 0, true},
	} {
		got, err := parseLimit(tc.in, "fixture")
		if (err != nil) != tc.wantErr {
			t.Errorf("parseLimit(%q) err = %v, wantErr %v", tc.in, err, tc.wantErr)
			continue
		}
		if err == nil && got != tc.want {
			t.Errorf("parseLimit(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

// TestLimitsCheckAssertsWhatTheArtifactAsks is why the thresholds are constants in
// internal/config rather than literals here: a check that asserted something the
// artifact does not ask for would pass while the broker started under-provisioned.
func TestLimitsCheckAssertsWhatTheArtifactAsks(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	unit := string(render.Quadlet(cfg, cfg.ResolveNode(config.Primary)))
	for _, want := range []string{
		"Ulimit=nofile=" + config.ContainerNoFile(),
		"LimitNOFILE=" + config.ContainerNoFile(),
	} {
		if !strings.Contains(unit, want) {
			t.Errorf("the quadlet must ask for %q, or the check is asserting the wrong number:\n%s", want, unit)
		}
	}
	dcfg := ctrCfg(config.Docker, "false")
	compose := string(render.Compose(dcfg, dcfg.ResolveNode(config.Primary)))
	if !strings.Contains(compose, "hard: 1048576") {
		t.Errorf("compose must ask for the same hard limit the check asserts:\n%s", compose)
	}
}

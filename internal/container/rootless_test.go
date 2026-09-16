package container

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"solace/internal/config"
)

// --- fixtures ---------------------------------------------------------------

// healthyRootlessOut answers every probe in the podman host-readiness block the
// way a correctly prepared host would. It is keyed by the shape of each call
// rather than by index, so a test can reorder or add rows without re-teaching it.
//
// It is shared with limits_test.go's limitsMgr: those tests are about the host
// ceilings, so every other row has to pass for the failure they assert to be the
// one they mean. nrOpen is the fs.nr_open answer, and the user-manager row always
// answers healthily -- shortLimitsOut is what breaks it.
func healthyRootlessOut(nrOpen string) func(name string, args []string) []byte {
	return func(name string, args []string) []byte {
		switch {
		case name == "id":
			return []byte("solace\n")
		case name == "loginctl":
			return []byte("Linger=yes\n")
		// Before the Version row: both are `systemctl show`, and only this one asks
		// the SYSTEM manager about user@<uid>.service's limits (limits.go).
		case name == "systemctl" && slices.Contains(args, "-p") && slices.Contains(args, "LimitNOFILE"):
			return []byte(healthyUserManagerProps)
		case name == "systemctl" && slices.Contains(args, "show"):
			return []byte("Version=255\n")
		case name == "sh" && slices.Contains(args, nearestWritableScript):
			return []byte("/opt")
		case name == "sh" && slices.Contains(args, "cat "+nrOpenPath):
			return []byte(nrOpen)
		case slices.Contains(args, "--format") && strings.Contains(strings.Join(args, " "), "IDMappings"):
			// 0:1 is this user mapped to container root; 1:65536 is the subuid
			// range, which is what covers ctrCfg's runUser of 1000:0.
			return []byte("0:1 1:65536 \n")
		}
		return nil
	}
}

// lingerOff is healthyRootlessOut with the one row that can be REPAIRED turned off,
// which is what separates the read-only caller from the mutating one.
func lingerOff(nrOpen string) func(name string, args []string) []byte {
	healthy := healthyRootlessOut(nrOpen)
	return func(name string, args []string) []byte {
		if name == "loginctl" {
			return []byte("Linger=no\n")
		}
		return healthy(name, args)
	}
}

// fakeEnv replaces the Manager's environment seams with a map, so a test can seed
// XDG_RUNTIME_DIR / DBUS_SESSION_BUS_ADDRESS and read back what ensureUserSession
// wrote -- without touching the test binary's own environment, which os.Setenv
// would leak into every test after it.
func fakeEnv(m *Manager, initial map[string]string) map[string]string {
	env := map[string]string{}
	for k, v := range initial {
		env[k] = v
	}
	m.Getenv = func(k string) string { return env[k] }
	m.Setenv = func(k, v string) error { env[k] = v; return nil }
	return env
}

// rootlessMgr is the healthy rootless podman Manager every case below starts from
// and then breaks in exactly one place.
func rootlessMgr() (*Manager, *capRunner, *bytes.Buffer) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.Rootless = true
	// Both knobs ApplyDefaults derives from Rootless, set by hand for the same
	// reason ctrCfg sets Scaling.CPU: the executors must work without config.Load,
	// so the fixture skips ApplyDefaults. Leaving SystemctlUser empty would send
	// every rootless probe to the SYSTEM bus, which is the one thing these rows
	// exist to prove does not happen.
	cfg.Podman.SystemctlUser = "--user"
	cfg.Podman.WantedBy = "default.target"
	m, rr, buf := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return 1000 }
	rr.outFor = healthyRootlessOut(healthyNrOpen)
	return m, rr, buf
}

// failOnCall makes exactly one captured Output call fail, leaving every other
// probe healthy -- which is what makes each row's failure attributable.
func failOnCall(rr *capRunner, match func(name string, args []string) bool) {
	rr.outFail = func(name string, args []string) error {
		if match(name, args) {
			return errors.New("boom")
		}
		return nil
	}
}

// --- the healthy path -------------------------------------------------------

func TestCheckPodmanHostHealthyReportsEveryRow(t *testing.T) {
	m, rr, buf := rootlessMgr()
	fakeEnv(m, nil)
	if err := m.checkPodmanHost(context.Background(), false); err != nil {
		t.Fatalf("a healthy rootless host must pass: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"euid: 1000", "user session:", "id mapping:", "linger: enabled",
		"user systemd: answering", "data dir:",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("report is missing %q:\n%s", want, out)
		}
	}
	// The exact argv of each probe, so a rename or a dropped flag is caught here
	// rather than on a host nobody can reproduce.
	for _, want := range []struct {
		name string
		args []string
	}{
		{"loginctl", []string{"show-user", "1000", "--property=Linger"}},
		// withUser, not a hardcoded --user: the token comes from the config knob
		// ApplyDefaults derives, so the assertion tracks the wiring rather than
		// restating it.
		{"systemctl", withUser(m.Cfg, "show", "--property=Version")},
		{"sh", []string{"-c", nearestWritableScript, "/opt/solace/data"}},
	} {
		if !hasCall(rr, want.name, want.args) {
			t.Errorf("missing probe %s %v:\n%+v", want.name, want.args, rr.calls)
		}
	}
}

// TestCheckPodmanHostIsReadOnly is the property that lets this block run from
// `validate` and from the top of `deploy` alike: nothing it does changes the host,
// so a refusal leaves "nothing happened" true.
// It is asserted with linger OFF, which is the one thing the block CAN repair: the
// read-only caller must report it and leave it alone, or `validate`'s own promise
// that it is safe against a system you don't want to disturb stops being true.
func TestCheckPodmanHostIsReadOnly(t *testing.T) {
	m, rr, _ := rootlessMgr()
	fakeEnv(m, nil)
	rr.outFor = lingerOff(healthyNrOpen)
	if err := m.checkPodmanHost(context.Background(), false); err == nil {
		t.Fatal("linger off must be reported by the read-only caller")
	}
	for _, c := range rr.calls {
		if c.method != "Output" {
			t.Errorf("host-readiness probes must only capture output, got %s %s %v", c.method, c.name, c.args)
		}
		if slices.Contains(c.args, "enable-linger") {
			t.Errorf("validate must never enable linger: %s %v", c.name, c.args)
		}
		if c.name == "mkdir" || c.name == "chown" || slices.Contains(c.args, "unshare") {
			t.Errorf("a read-only check must not mutate the host: %s %v", c.name, c.args)
		}
	}
}

// --- gating -----------------------------------------------------------------

func TestCheckPodmanHostSkipsDockerAndPreview(t *testing.T) {
	m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	if err := m.checkPodmanHost(context.Background(), false); err != nil {
		t.Fatalf("docker has no podman host block: %v", err)
	}
	if len(rr.calls) != 0 {
		t.Errorf("docker must probe nothing here:\n%+v", rr.calls)
	}

	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.Rootless = true
	em, buf := newEchoMgr(cfg, config.Podman)
	if err := em.checkPodmanHost(context.Background(), false); err != nil {
		t.Fatalf("preview must not assert: %v", err)
	}
	if !strings.Contains(buf.String(), "skipped (preview)") {
		t.Errorf("the Echo runner answers nothing, so the block must skip:\n%s", buf)
	}
}

// TestCheckPodmanHostRootfulStopsAtEUID: a privileged engine owns the id mapping
// and installs its units under the system systemd instance, so none of the
// rootless rows apply. Its nofile ceiling is checked one level up, where it
// applies to every engine (limits.go).
func TestCheckPodmanHostRootfulStopsAtEUID(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.Rootless = false
	m, rr, buf := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return 0 }
	if err := m.checkPodmanHost(context.Background(), false); err != nil {
		t.Fatalf("rootful podman: %v", err)
	}
	if !strings.Contains(buf.String(), "euid: 0") {
		t.Errorf("the euid row runs on rootful too:\n%s", buf)
	}
	if len(rr.calls) != 0 {
		t.Errorf("no rootless probe may run on rootful podman:\n%+v", rr.calls)
	}
}

// TestCheckPodmanHostEUIDMismatchSkipsTheRest is the reachability-first rule: a
// rootless block probed as root would read the WRONG user's linger and runtime
// directory, so those rows are honest skips rather than answers about an account
// the deploy will never use.
func TestCheckPodmanHostEUIDMismatchSkipsTheRest(t *testing.T) {
	m, rr, buf := rootlessMgr()
	m.Geteuid = func() int { return 0 } // rootless, but running as root
	err := m.checkPodmanHost(context.Background(), false)
	if err == nil {
		t.Fatal("rootless podman running as root must fail")
	}
	out := buf.String()
	if !strings.Contains(out, "euid:") || !strings.Contains(out, "skipped (euid mismatch)") {
		t.Errorf("the later rows must be skipped, not answered:\n%s", out)
	}
	if len(rr.calls) != 0 {
		t.Errorf("nothing may be probed once the euid is wrong:\n%+v", rr.calls)
	}
}

// --- the user session -------------------------------------------------------

func TestEnsureUserSessionDerivesBothWhenUnset(t *testing.T) {
	m, rr, buf := rootlessMgr()
	env := fakeEnv(m, nil)
	if err := m.checkPodmanHost(context.Background(), false); err != nil {
		t.Fatalf("checkPodmanHost: %v", err)
	}
	if got := env[xdgRuntimeDir]; got != "/run/user/1000" {
		t.Errorf("%s = %q, want it derived from the euid", xdgRuntimeDir, got)
	}
	if got := env[dbusSessionBus]; got != "unix:path=/run/user/1000/bus" {
		t.Errorf("%s = %q, want it derived from the runtime dir", dbusSessionBus, got)
	}
	if !hasCall(rr, "sh", []string{"-c", `test -d "$0"`, "/run/user/1000"}) {
		t.Errorf("the runtime dir must be proven to exist before it is exported:\n%+v", rr.calls)
	}
	if !strings.Contains(buf.String(), "(derived)") {
		t.Errorf("the report must say a value was invented here:\n%s", buf)
	}
}

// TestEnsureUserSessionNeverOverwritesInherited: an operator who set one meant it.
func TestEnsureUserSessionNeverOverwritesInherited(t *testing.T) {
	m, _, buf := rootlessMgr()
	env := fakeEnv(m, map[string]string{
		xdgRuntimeDir:  "/custom/run",
		dbusSessionBus: "unix:path=/custom/bus",
	})
	if err := m.checkPodmanHost(context.Background(), false); err != nil {
		t.Fatalf("checkPodmanHost: %v", err)
	}
	if env[xdgRuntimeDir] != "/custom/run" || env[dbusSessionBus] != "unix:path=/custom/bus" {
		t.Errorf("an inherited value must survive untouched, got %v", env)
	}
	if !strings.Contains(buf.String(), "(inherited)") {
		t.Errorf("the report must distinguish an inherited value from a derived one:\n%s", buf)
	}
}

// TestEnsureUserSessionDerivesEachHalfIndependently covers the sudo-ish middle
// ground where one variable survives and the other does not.
func TestEnsureUserSessionDerivesEachHalfIndependently(t *testing.T) {
	m, _, buf := rootlessMgr()
	env := fakeEnv(m, map[string]string{xdgRuntimeDir: "/custom/run"})
	if err := m.checkPodmanHost(context.Background(), false); err != nil {
		t.Fatalf("checkPodmanHost: %v", err)
	}
	if env[dbusSessionBus] != "unix:path=/custom/run/bus" {
		t.Errorf("the derived bus must be built from the INHERITED runtime dir, got %q", env[dbusSessionBus])
	}
	out := buf.String()
	if !strings.Contains(out, "(inherited)") || !strings.Contains(out, "(derived)") {
		t.Errorf("the report must name both origins:\n%s", out)
	}
}

// TestEnsureUserSessionExportsNothingWhenTheDirIsMissing is the ordering this
// whole function exists to get right: a derived path to a directory that is not
// there turns "you are under sudo, or this user has no session" into an obscure
// systemctl failure several rows later.
func TestEnsureUserSessionExportsNothingWhenTheDirIsMissing(t *testing.T) {
	m, rr, buf := rootlessMgr()
	env := fakeEnv(m, nil)
	failOnCall(rr, func(name string, args []string) bool {
		return name == "sh" && slices.Contains(args, `test -d "$0"`)
	})
	err := m.checkPodmanHost(context.Background(), false)
	if err == nil {
		t.Fatal("a missing user runtime directory must fail the block")
	}
	if len(env) != 0 {
		t.Errorf("nothing may be exported once the directory is known absent, got %v", env)
	}
	for _, want := range []string{"sudo", "enable-linger", "/run/user/1000"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error must mention %q, got: %v", want, err)
		}
	}
	// The bus-dependent rows cannot answer for their own reasons any more.
	out := buf.String()
	if !strings.Contains(out, "linger: skipped (no user session)") ||
		!strings.Contains(out, "user systemd: skipped (no user session)") {
		t.Errorf("rows that need the bus must be skipped, not failed:\n%s", out)
	}
	// The rows that do NOT need the bus still run, so one pass still names them.
	if !strings.Contains(out, "id mapping:") || !strings.Contains(out, "data dir:") {
		t.Errorf("bus-independent rows must still report:\n%s", out)
	}
}

// TestEnsureUserSessionProbesOncePerRun: every systemctl call funnels through it,
// so it must cost one probe per run rather than one per command.
func TestEnsureUserSessionProbesOncePerRun(t *testing.T) {
	m, rr, _ := rootlessMgr()
	fakeEnv(m, nil)
	ctx := context.Background()
	for range 3 {
		if _, err := m.systemctlOutput(ctx, "is-active", "sol-pod.service"); err != nil {
			t.Fatalf("systemctlOutput: %v", err)
		}
	}
	n := 0
	for _, c := range rr.calls {
		if c.name == "sh" && slices.Contains(c.args, `test -d "$0"`) {
			n++
		}
	}
	if n != 1 {
		t.Errorf("the session probe ran %d times, want exactly 1:\n%+v", n, rr.calls)
	}
}

// TestSystemctlCarriesTheSessionWithoutPrepHost is why the memoised call lives on
// systemctl rather than only in the readiness block: `broker start`/`stop`/
// `restart`/`remove`/`status` drive systemd without ever calling PrepHost.
func TestSystemctlCarriesTheSessionWithoutPrepHost(t *testing.T) {
	m, rr, _ := rootlessMgr()
	env := fakeEnv(m, nil)
	if err := m.systemctl(context.Background(), "start", "sol-pod.service"); err != nil {
		t.Fatalf("systemctl: %v", err)
	}
	if env[xdgRuntimeDir] == "" {
		t.Errorf("a systemctl path with no PrepHost must still get a user session:\n%+v", rr.calls)
	}
}

// --- the individual rows ----------------------------------------------------

func TestCheckIDMappingRefusesAnUnmappedRunUser(t *testing.T) {
	m, rr, buf := rootlessMgr()
	fakeEnv(m, nil)
	// A range that stops well below the configured runUser of 1000:0.
	rr.outFor = func(name string, args []string) []byte {
		if slices.Contains(args, "--format") && strings.Contains(strings.Join(args, " "), "IDMappings") {
			return []byte("0:1 1:10 ")
		}
		return healthyRootlessOut(healthyNrOpen)(name, args)
	}
	err := m.checkPodmanHost(context.Background(), false)
	if err == nil {
		t.Fatal("a runUser outside this user's subuid range must be refused")
	}
	// The diagnosis and the command, not the mechanism: why a rootless container
	// cannot use an unmapped id belongs in docs/operations.md, not in an error.
	for _, want := range []string{"1000", "subuid", "runUser", "usermod --add-subuids", "podman system migrate"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error must mention %q, got: %v", want, err)
		}
	}
	if !strings.Contains(buf.String(), "id mapping:") {
		t.Errorf("the failing row must be reported:\n%s", buf)
	}
}

// TestCheckIDMappingSkipsNamedRunUser: config accepts account names here, and they
// name accounts inside the container image, which cannot be resolved from the host.
func TestCheckIDMappingSkipsNamedRunUser(t *testing.T) {
	m, _, buf := rootlessMgr()
	fakeEnv(m, nil)
	m.Cfg.Podman.Container.RunUser = "solace:solace"
	if err := m.checkPodmanHost(context.Background(), false); err != nil {
		t.Fatalf("a named runUser must not fail the block: %v", err)
	}
	if !strings.Contains(buf.String(), "names accounts rather than numeric ids") {
		t.Errorf("the skip must say why it skipped:\n%s", buf)
	}
}

// TestCheckIDMappingRefusesWhenNothingIsAllocated is the case a `useradd --system`
// or a directory (LDAP/AD) account arrives in. This process ALREADY runs as the
// account the container will use, so an empty list is an answer about that account
// rather than a gap in what was asked -- podman then maps only container id 0, and
// every non-zero runUser is unusable with nothing on screen to say why.
func TestCheckIDMappingRefusesWhenNothingIsAllocated(t *testing.T) {
	m, rr, buf := rootlessMgr()
	fakeEnv(m, nil)
	rr.outFor = func(name string, args []string) []byte {
		if slices.Contains(args, "--format") && strings.Contains(strings.Join(args, " "), "IDMappings") {
			return nil
		}
		return healthyRootlessOut(healthyNrOpen)(name, args)
	}
	err := m.checkPodmanHost(context.Background(), false)
	if err == nil {
		t.Fatal("no subuid allocation at all must be refused, not skipped")
	}
	for _, want := range []string{"no subuid allocation", "usermod --add-subuids", "podman system migrate"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error must mention %q, got: %v", want, err)
		}
	}
	if !strings.Contains(buf.String(), "no subuid allocation at all") {
		t.Errorf("the failing row must say the allocation is missing entirely:\n%s", buf)
	}
}

// TestIDMappingCommandsNameTheAccountLiterally: the privileged line is meant to be
// forwarded to an administrator, so a `$(id -un)` in it would resolve to whoever
// ends up running it rather than to the account that needs the range.
func TestIDMappingCommandsNameTheAccountLiterally(t *testing.T) {
	m, rr, _ := rootlessMgr()
	fakeEnv(m, nil)
	rr.outFor = func(name string, args []string) []byte {
		if name == "id" {
			return []byte("solace-svc\n")
		}
		if slices.Contains(args, "--format") && strings.Contains(strings.Join(args, " "), "IDMappings") {
			return []byte("0:1 1:10 ") // too small for the 1000:0 default
		}
		return healthyRootlessOut(healthyNrOpen)(name, args)
	}
	err := m.checkPodmanHost(context.Background(), false)
	if err == nil {
		t.Fatal("a range that does not reach runUser must be refused")
	}
	if !strings.Contains(err.Error(), "--add-subgids 100000-165535 solace-svc") {
		t.Errorf("the handed-over command must name the account literally, got: %v", err)
	}
	if strings.Contains(err.Error(), "$(id -un)") {
		t.Errorf("a command substitution would resolve to whoever runs it: %v", err)
	}
}

// TestSuggestSubIDRangeAvoidsExistingAllocations is why the range is computed rather
// than hardcoded: two accounts sharing a subuid range map to the same host uids,
// which is a containment hole, and only newer shadow-utils refuse to create one.
func TestSuggestSubIDRangeAvoidsExistingAllocations(t *testing.T) {
	for _, tc := range []struct {
		name       string
		etc        string
		need       int
		start, end int
	}{
		{"empty host starts at the floor", "", 1001, 100000, 165535},
		{"next block after one account", "alice:100000:65536\nalice:100000:65536", 1001, 165536, 231071},
		{
			"highest wins regardless of order",
			"bob:200000:65536\nalice:100000:65536\ncarol:165536:65536",
			1001, 265536, 331071,
		},
		{"junk lines are skipped", "not-a-range\nalice:abc:65536\nalice:100000:65536", 1001, 165536, 231071},
		// A need larger than the conventional block still gets a range that covers
		// it -- which is what the image's own uid would ask for.
		{"a large need widens the block", "", 1000002, 100000, 1100001},
	} {
		m, rr, _ := rootlessMgr()
		rr.outFor = func(name string, args []string) []byte {
			if name == "sh" && slices.Contains(args, "cat /etc/subuid /etc/subgid 2>/dev/null") {
				return []byte(tc.etc)
			}
			return nil
		}
		start, end, computed := m.suggestSubIDRange(context.Background(), tc.need)
		if !computed {
			t.Errorf("%s: a readable /etc/subuid must yield a computed range", tc.name)
		}
		if start != tc.start || end != tc.end {
			t.Errorf("%s: range = %d-%d, want %d-%d", tc.name, start, end, tc.start, tc.end)
		}
		if end-start+1 < tc.need {
			t.Errorf("%s: range covers %d ids, need %d", tc.name, end-start+1, tc.need)
		}
	}
}

// TestSuggestSubIDRangeSaysWhenItIsGuessing: an unreadable /etc/subuid means the
// proposal is a convention rather than an answer, and the error has to say so --
// otherwise an administrator applies an overlapping range believing it was checked.
func TestSuggestSubIDRangeSaysWhenItIsGuessing(t *testing.T) {
	m, rr, _ := rootlessMgr()
	fakeEnv(m, nil)
	rr.outFor = func(name string, args []string) []byte {
		if slices.Contains(args, "--format") && strings.Contains(strings.Join(args, " "), "IDMappings") {
			return nil // no allocation at all
		}
		return healthyRootlessOut(healthyNrOpen)(name, args)
	}
	failOnCall(rr, func(name string, args []string) bool {
		return name == "sh" && slices.Contains(args, "cat /etc/subuid /etc/subgid 2>/dev/null")
	})
	err := m.checkPodmanHost(context.Background(), false)
	if err == nil {
		t.Fatal("no allocation must still fail when the suggestion cannot be computed")
	}
	if !strings.Contains(err.Error(), "unreadable") || !strings.Contains(err.Error(), "overlap") {
		t.Errorf("an uncomputed range must be flagged as one to check, got: %v", err)
	}
}

// TestIDMappingErrProposesTheNextFreeBlock ties the computation to the message an
// operator actually forwards.
func TestIDMappingErrProposesTheNextFreeBlock(t *testing.T) {
	m, rr, _ := rootlessMgr()
	fakeEnv(m, nil)
	rr.outFor = func(name string, args []string) []byte {
		switch {
		case name == "id":
			return []byte("solace-svc\n")
		case name == "sh" && slices.Contains(args, "cat /etc/subuid /etc/subgid 2>/dev/null"):
			return []byte("alice:100000:65536\n")
		case slices.Contains(args, "--format") && strings.Contains(strings.Join(args, " "), "IDMappings"):
			return nil
		}
		return healthyRootlessOut(healthyNrOpen)(name, args)
	}
	err := m.checkPodmanHost(context.Background(), false)
	if err == nil {
		t.Fatal("no allocation must fail")
	}
	if !strings.Contains(err.Error(), "--add-subuids 165536-231071 --add-subgids 165536-231071 solace-svc") {
		t.Errorf("the command must propose the next free block for this account, got: %v", err)
	}
	// A computed range carries no caveat: the comment is only for the unreadable case.
	if strings.Contains(err.Error(), "overlap") {
		t.Errorf("a computed range needs no overlap caveat, got: %v", err)
	}
}

// TestHostUserNameFallsBackToTheUID: `id -un` is the authoritative answer (it sees
// directory accounts that a /etc/passwd read would miss), but usermod takes a
// numeric uid too, so a failed lookup still yields a usable command.
func TestHostUserNameFallsBackToTheUID(t *testing.T) {
	m, rr, _ := rootlessMgr()
	failOnCall(rr, func(name string, _ []string) bool { return name == "id" })
	if got := m.hostUserName(context.Background(), 1000); got != "1000" {
		t.Errorf("hostUserName = %q, want the uid as a fallback", got)
	}
}

func TestCheckLingerRefusesWhenDisabled(t *testing.T) {
	m, rr, buf := rootlessMgr()
	fakeEnv(m, nil)
	rr.outFor = lingerOff(healthyNrOpen)
	err := m.checkPodmanHost(context.Background(), false)
	if err == nil {
		t.Fatal("linger off must refuse: systemd stops the broker at logout")
	}
	// The consequence in one clause, and the command. Why systemd behaves this way
	// is documented, not restated here.
	for _, want := range []string{"linger is not enabled", "logout", "loginctl enable-linger"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error must mention %q, got: %v", want, err)
		}
	}
	if !strings.Contains(buf.String(), "linger: not enabled") {
		t.Errorf("the failing row must be reported:\n%s", buf)
	}
}

// TestCheckLingerRefusesWhenLoginctlCannotRun: quadlet is a systemd generator, so
// a host whose logind does not answer cannot run this deployment at all. It is
// reached only after the session row passed, so it can no longer fail merely for a
// missing bus address.
func TestCheckLingerRefusesWhenLoginctlCannotRun(t *testing.T) {
	m, rr, _ := rootlessMgr()
	fakeEnv(m, nil)
	failOnCall(rr, func(name string, _ []string) bool { return name == "loginctl" })
	err := m.checkPodmanHost(context.Background(), false)
	if err == nil || !strings.Contains(err.Error(), "linger state") {
		t.Errorf("an unrunnable loginctl must fail loud, got: %v", err)
	}
}

// TestCheckUserSystemdRunsBeforeAnythingIsWritten is the promise Preflight makes
// and deploy's own daemon-reload cannot keep: the unit is on disk by then.
func TestCheckUserSystemdRefusesWhenTheBusIsDead(t *testing.T) {
	m, rr, _ := rootlessMgr()
	fakeEnv(m, nil)
	failOnCall(rr, func(name string, args []string) bool {
		return name == "systemctl" && slices.Contains(args, "show")
	})
	err := m.checkPodmanHost(context.Background(), false)
	if err == nil {
		t.Fatal("a user systemd that does not answer must refuse")
	}
	for _, want := range []string{"podman.socket", "enable-linger", "sudo"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error must mention %q, got: %v", want, err)
		}
	}
}

// TestCheckDataDirRefusesAnUnwritableParent is the row that replaces a bare
// "permission denied" from mkdir: the defaults are rootful-shaped, so this is the
// ordinary first run as a non-root user.
func TestCheckDataDirRefusesAnUnwritableParent(t *testing.T) {
	m, rr, buf := rootlessMgr()
	fakeEnv(m, nil)
	failOnCall(rr, func(name string, args []string) bool {
		return name == "sh" && slices.Contains(args, nearestWritableScript)
	})
	err := m.checkPodmanHost(context.Background(), false)
	if err == nil {
		t.Fatal("an unwritable data dir parent must refuse before mkdir runs")
	}
	for _, want := range []string{"/opt/solace/data", "dataDir", "chown"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error must mention %q, got: %v", want, err)
		}
	}
	// The directory the shell settled on, not just the one that was configured.
	if !strings.Contains(buf.String(), "/opt is not writable") {
		t.Errorf("the row must name the existing directory that is at fault:\n%s", buf)
	}
}

// TestCheckPodmanHostReportsEveryFailureInOnePass is what `validate`'s own help
// promises, and the reason the errors are joined rather than returned at the first.
func TestCheckPodmanHostReportsEveryFailureInOnePass(t *testing.T) {
	m, rr, _ := rootlessMgr()
	fakeEnv(m, nil)
	// Linger off AND an unwritable data dir: two rows this block still owns, since
	// the host ceilings moved out of it (limits.go).
	rr.outFor = lingerOff(healthyNrOpen)
	failOnCall(rr, func(name string, args []string) bool {
		return name == "sh" && slices.Contains(args, nearestWritableScript)
	})
	err := m.checkPodmanHost(context.Background(), false)
	if err == nil {
		t.Fatal("two broken rows must fail")
	}
	if !strings.Contains(err.Error(), "enable-linger") || !strings.Contains(err.Error(), "dataDir") {
		t.Errorf("both failures must survive into the joined error, got: %v", err)
	}
}

// --- PrepHost integration ---------------------------------------------------

// TestPrepHostRefusesBeforeTouchingTheHost: every row is read-only, so a refusal
// still leaves "nothing happened" true -- no mkdir, no chown, no unshare.
// The unready row here is the data directory, deliberately NOT linger: linger is
// the one thing prep now repairs, so using it would assert the opposite of what
// this test is for.
func TestPrepHostRefusesBeforeTouchingTheHost(t *testing.T) {
	m, rr, _ := rootlessMgr()
	fakeEnv(m, nil)
	failOnCall(rr, func(name string, args []string) bool {
		return name == "sh" && slices.Contains(args, nearestWritableScript)
	})
	if err := m.PrepHost(context.Background()); err == nil {
		t.Fatal("PrepHost must refuse when the host is not ready")
	}
	for _, c := range rr.calls {
		if c.name == "mkdir" || c.name == "chown" || slices.Contains(c.args, "unshare") {
			t.Errorf("nothing may be created once prep has refused: %s %v", c.name, c.args)
		}
	}
}

// --- linger is the one row prep repairs ------------------------------------

// TestPrepHostEnablesLinger is the whole point of the fix flag: enabling linger for
// the INVOKING user needs no privilege (polkit's set-self-linger), so it belongs
// with the other unprivileged preparations rather than in a refusal.
//
// The argv is asserted exactly, and the absence of a username argument is asserted
// separately: `loginctl enable-linger <user>` is the variant that needs admin
// authentication, and this must never be able to issue it.
func TestPrepHostEnablesLinger(t *testing.T) {
	m, rr, buf := rootlessMgr()
	fakeEnv(m, nil)
	rr.outFor = lingerOff(healthyNrOpen)
	if err := m.PrepHost(context.Background()); err != nil {
		t.Fatalf("prep must enable linger rather than refusing: %v", err)
	}
	if !hasCall(rr, "loginctl", []string{"enable-linger"}) {
		t.Errorf("prep should enable linger:\n%+v", rr.calls)
	}
	for _, c := range rr.calls {
		if c.name == "loginctl" && slices.Contains(c.args, "enable-linger") && len(c.args) != 1 {
			t.Errorf("enable-linger must take no username -- naming one is the variant that needs admin auth: %v", c.args)
		}
	}
	if !strings.Contains(buf.String(), "linger: enabled for uid 1000 (was no)") {
		t.Errorf("the report should say it changed the state, and from what:\n%s", buf)
	}
	// And prep carries on: the repair is not a reason to stop.
	if !hasCall(rr, "mkdir", []string{"-p", "/opt/solace/data"}) {
		t.Errorf("prep should continue after enabling linger:\n%+v", rr.calls)
	}
}

// TestPrepHostSkipsLingerWhenAlreadyEnabled: the common case costs one read and no
// write, so re-running deploy does not keep poking logind.
func TestPrepHostSkipsLingerWhenAlreadyEnabled(t *testing.T) {
	m, rr, _ := rootlessMgr()
	fakeEnv(m, nil)
	if err := m.PrepHost(context.Background()); err != nil {
		t.Fatalf("PrepHost: %v", err)
	}
	for _, c := range rr.calls {
		if slices.Contains(c.args, "enable-linger") {
			t.Errorf("linger was already on, so nothing should have been enabled: %v", c.args)
		}
	}
}

// TestPrepHostLingerEnableFailureIsActionable covers the case the privilege
// argument does NOT cover: polkit can still decline, on a session that is not
// active or a distribution that tightened set-self-linger. The error has to say so,
// because "you don't need root for this" would otherwise read as a contradiction.
func TestPrepHostLingerEnableFailureIsActionable(t *testing.T) {
	m, rr, buf := rootlessMgr()
	fakeEnv(m, nil)
	rr.outFor = lingerOff(healthyNrOpen)
	rr.fail = func(name string, args []string) error {
		if name == "loginctl" && slices.Contains(args, "enable-linger") {
			return errors.New("Access denied")
		}
		return nil
	}
	err := m.PrepHost(context.Background())
	if err == nil {
		t.Fatal("a refused enable-linger must fail prep")
	}
	// polkit refusing the self form leaves exactly one thing to do, so that is all
	// the message carries: the administrator's form, naming the account.
	for _, want := range []string{"cannot enable linger", "sudo loginctl enable-linger solace"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error must mention %q, got: %v", want, err)
		}
	}
	if !strings.Contains(buf.String(), "linger: could not enable") {
		t.Errorf("the failing row must be reported:\n%s", buf)
	}
	// Nothing may be created once the repair itself failed.
	for _, c := range rr.calls {
		if c.name == "mkdir" || slices.Contains(c.args, "unshare") {
			t.Errorf("prep must stop when it cannot make the host ready: %s %v", c.name, c.args)
		}
	}
}

// TestValidateReportsLingerAndPrepFixesIt pins the two callers reading the SAME
// block differently -- one definition of what a ready host is, two dispositions.
func TestValidateReportsLingerAndPrepFixesIt(t *testing.T) {
	for _, tc := range []struct {
		name   string
		fix    bool
		wantOK bool
	}{
		{"validate reports", false, false},
		{"prep repairs", true, true},
	} {
		m, rr, _ := rootlessMgr()
		fakeEnv(m, nil)
		rr.outFor = lingerOff(healthyNrOpen)
		err := m.checkPodmanHost(context.Background(), tc.fix)
		if tc.wantOK && err != nil {
			t.Errorf("%s: %v", tc.name, err)
		}
		if !tc.wantOK && err == nil {
			t.Errorf("%s: linger off must be reported as a failure", tc.name)
		}
		enabled := hasCall(rr, "loginctl", []string{"enable-linger"})
		if enabled != tc.fix {
			t.Errorf("%s: enable-linger issued = %v, want %v:\n%+v", tc.name, enabled, tc.fix, rr.calls)
		}
	}
}

// TestPrepHostMkdirFailureExplainsRootless ports the guidance the legacy bash
// carried and the Go port dropped. The readiness row above should make it
// unreachable in practice; this is the race where ownership changes in between.
func TestPrepHostMkdirFailureExplainsRootless(t *testing.T) {
	m, rr, _ := rootlessMgr()
	fakeEnv(m, nil)
	rr.fail = func(name string, _ []string) error {
		if name == "mkdir" {
			return errors.New("permission denied")
		}
		return nil
	}
	err := m.PrepHost(context.Background())
	if err == nil {
		t.Fatal("a failed mkdir must fail prep")
	}
	for _, want := range []string{"/opt/solace/data", "podman.container.dataDir", "chown"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the mkdir error must mention %q, got: %v", want, err)
		}
	}
}

// TestPrepHostMkdirHintIsRootlessOnly: rootful podman and docker create the
// directory as root, so the rootless advice would be wrong there.
func TestPrepHostMkdirHintIsRootlessOnly(t *testing.T) {
	for _, p := range []config.Platform{config.Podman, config.Docker} {
		cfg := ctrCfg(p, "false")
		cfg.Podman.Rootless = false
		m, rr, _ := newCapMgr(cfg, p)
		if p == config.Podman {
			m.Geteuid = func() int { return 0 }
		}
		rr.fail = func(name string, _ []string) error {
			if name == "mkdir" {
				return errors.New("permission denied")
			}
			return nil
		}
		err := m.PrepHost(context.Background())
		if err == nil {
			t.Fatalf("%s: a failed mkdir must fail prep", p)
		}
		if strings.Contains(err.Error(), "rootless deploy runs as this user") {
			t.Errorf("%s: the rootless hint must not appear on a privileged create: %v", p, err)
		}
	}
}

// --- Check integration ------------------------------------------------------

// TestCheckReportsDNSAndPodmanHostTogether is the one-pass promise at the Check
// level: a DNS failure must not hide a host-readiness failure or the reverse.
func TestCheckReportsDNSAndPodmanHostTogether(t *testing.T) {
	cfg := ctrCfg(config.Podman, "true") // HA -> DNS failures are loud
	cfg.Podman.Rootless = true
	cfg.Podman.SystemctlUser = "--user" // derived by ApplyDefaults, which ctrCfg skips
	cfg.Podman.WantedBy = "default.target"
	m, rr, buf := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return 1000 }
	fakeEnv(m, nil)
	m.Resolve = func(host string) bool { return host != "bkp-host" }
	// Three blocks broken at once: DNS, the podman readiness rows, and the host
	// ceiling. The one-pass promise now spans all three.
	rr.outFor = lingerOff("65536\n")
	err := m.Check(context.Background())
	if err == nil {
		t.Fatal("every half is broken, so Check must fail")
	}
	out := buf.String()
	if !strings.Contains(out, "does NOT resolve: bkp-host") {
		t.Errorf("the DNS row must still report:\n%s", out)
	}
	if !strings.Contains(out, "linger: not enabled") {
		t.Errorf("the podman host row must report in the same pass:\n%s", out)
	}
	if !strings.Contains(out, "nr_open: 65536") {
		t.Errorf("the host-ceiling row must report in the same pass:\n%s", out)
	}
	for _, want := range []string{"do not resolve", "enable-linger", "fs.nr_open"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the joined error should carry %q, got: %v", want, err)
		}
	}
}

package container

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"solace/internal/config"
)

// dockerInspect is a docker `inspect` reply, trimmed to the fields this tool reads.
// Note the LEADING SLASH on .Name and the health block spelled "Health" -- the two
// things that differ from podman.
const dockerInspect = `[{
  "Name": "/solbroker",
  "RestartCount": 3,
  "State": {
    "Status": "running",
    "Running": true,
    "StartedAt": "2026-09-08T10:11:12.13Z",
    "FinishedAt": "0001-01-01T00:00:00Z",
    "ExitCode": 0,
    "Health": {"Status": "healthy", "FailingStreak": 0}
  },
  "Config": {"Image": "solace/solace-pubsub-standard:10.10.1.128"},
  "Mounts": [
    {"Source": "/opt/solace/data", "Destination": "/var/lib/solace/jail", "RW": true},
    {"Source": "/opt/solace/certs/tls.pem", "Destination": "/mnt/certs/server/tls.pem", "RW": false}
  ]
}]`

// podmanInspect is podman's reply for the same container: no leading slash, and the
// health block spelled "Healthcheck".
const podmanInspect = `[{
  "Name": "solbroker",
  "RestartCount": 0,
  "State": {
    "Status": "running",
    "Running": true,
    "StartedAt": "2026-09-08T10:11:12.13Z",
    "Healthcheck": {"Status": "unhealthy", "FailingStreak": 4}
  },
  "Config": {"Image": "solace/solace-pubsub-standard:10.10.1.128"},
  "Mounts": [{"Source": "/opt/solace/data", "Destination": "/var/lib/solace/jail", "RW": true}]
}]`

// TestInspectNameLosesDockersLeadingSlash. `/solbroker` is not a name either engine
// accepts back, and printing it beside podman's `solbroker` would make one engine's
// report disagree with the other about what the container is called.
func TestInspectNameLosesDockersLeadingSlash(t *testing.T) {
	for _, tc := range []struct{ name, raw string }{
		{"docker", dockerInspect},
		{"podman", podmanInspect},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, err := decodeInspect([]byte(tc.raw))
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got := s.containerName(); got != "solbroker" {
				t.Errorf("containerName() = %q, want %q", got, "solbroker")
			}
		})
	}
}

// TestInspectHealthReadsBothSpellings is the one that matters most: a report naming
// only one engine's spelling renders EMPTY on the other, which reads exactly like
// "no healthcheck configured" -- the most misleading thing this report could say
// about a broker.
func TestInspectHealthReadsBothSpellings(t *testing.T) {
	docker, err := decodeInspect([]byte(dockerInspect))
	if err != nil {
		t.Fatalf("decode docker: %v", err)
	}
	status, configured := docker.health()
	if !configured || status != "healthy" {
		t.Errorf("docker health: got (%q, %v), want (\"healthy\", true)", status, configured)
	}

	podman, err := decodeInspect([]byte(podmanInspect))
	if err != nil {
		t.Fatalf("decode podman: %v", err)
	}
	status, configured = podman.health()
	if !configured {
		t.Fatal("podman health: reported as not configured, but Healthcheck.Status is set")
	}
	// The failing streak rides along: "unhealthy" alone does not say whether this is
	// the first failed probe or the twentieth.
	if !strings.Contains(status, "unhealthy") || !strings.Contains(status, "4") {
		t.Errorf("podman health: got %q, want it to name both the status and the failing streak", status)
	}
}

// TestInspectDistinguishesNoHealthcheckFromUnknown. An engine reports an
// unconfigured healthcheck as an empty status, so "not configured" and "configured
// but unknown" must not print the same -- one is a deployment choice, the other is a
// missing answer.
func TestInspectDistinguishesNoHealthcheckFromUnknown(t *testing.T) {
	const noHealth = `[{"Name": "b", "State": {"Status": "running", "Running": true},
	  "Config": {"Image": "i"}}]`
	s, err := decodeInspect([]byte(noHealth))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if status, configured := s.health(); configured || status != "" {
		t.Errorf("health() on a container with no healthcheck: got (%q, %v), want (\"\", false)",
			status, configured)
	}
}

// TestInspectRejectsAnEmptyReply. An engine answers a name it does not know with an
// empty ARRAY and exit 0 on some versions, so the absence has to be turned into an
// error here or the caller reports a container with no name, no image and no state
// as though that were its condition.
func TestInspectRejectsAnEmptyReply(t *testing.T) {
	for _, raw := range []string{"[]", "", "not json"} {
		if _, err := decodeInspect([]byte(raw)); err == nil {
			t.Errorf("decodeInspect(%q) returned no error", raw)
		}
	}
}

// TestInspectDecodesPartialOutput: an engine version that stops emitting a field
// must leave that field unknown, not fail the whole command. The report is the
// point, and three quarters of it is worth more than none.
func TestInspectDecodesPartialOutput(t *testing.T) {
	const partial = `[{"Name": "/b", "State": {"Status": "exited", "ExitCode": 137, "Error": "OOMKilled"}}]`
	s, err := decodeInspect([]byte(partial))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if s.Config.Image != "" {
		t.Errorf("Image = %q, want empty", s.Config.Image)
	}
	if s.State.ExitCode != 137 || s.State.Error != "OOMKilled" {
		t.Errorf("exit state = (%d, %q), want (137, \"OOMKilled\")", s.State.ExitCode, s.State.Error)
	}
	if orUnknown(s.Config.Image) != "(unknown)" {
		t.Error("a missing field must render as (unknown), not as an empty column")
	}
}

// inspectReply builds an inspect array for a named container reporting the given
// engine-side RestartCount.
func inspectReply(name string, engineRestarts int) []byte {
	return []byte(`[{"Name": "` + name + `", "RestartCount": ` + strconv.Itoa(engineRestarts) + `,
	  "State": {"Status": "running", "Running": true}, "Config": {"Image": "i"}}]`)
}

// TestRestartCountComesFromSystemdOnPodman is the fact this rewrite exists for. A
// quadlet unit is restarted by systemd, which stops and REPLACES the container, so
// the engine's own RestartCount stays 0 through any number of restarts -- reporting
// it would tell an operator their crash-looping broker has never restarted.
func TestRestartCountComesFromSystemdOnPodman(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	m, rr, _ := newCapMgr(cfg, config.Podman)
	rr.outFor = func(name string, args []string) []byte {
		if name == "systemctl" {
			return []byte("12\n")
		}
		return inspectReply("sol-pod", 0) // the engine's counter: untouched by systemd
	}

	s, err := m.inspectState(context.Background(), m.name())
	if err != nil {
		t.Fatalf("inspectState: %v", err)
	}
	got, ok := m.restartCount(context.Background(), s)
	if !ok || got != 12 {
		t.Fatalf("podman restart count = (%d, %v), want (12, true) -- systemd's NRestarts, "+
			"not the engine's %d", got, ok, s.RestartCount)
	}
	if !hasCall(rr, "systemctl", []string{"show", "-p", "NRestarts", "--value", "sol-pod.service"}) {
		t.Errorf("expected systemd to be asked for NRestarts on the unit:\n%+v", rr.calls)
	}
}

// TestRestartCountUsesTheEngineOnDocker: docker restarts the container in place, so
// its own counter is the answer and there is no unit to ask.
func TestRestartCountUsesTheEngineOnDocker(t *testing.T) {
	m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	rr.outFor = func(string, []string) []byte { return inspectReply("/solace", 3) }

	s, err := m.inspectState(context.Background(), m.name())
	if err != nil {
		t.Fatalf("inspectState: %v", err)
	}
	got, ok := m.restartCount(context.Background(), s)
	if !ok || got != 3 {
		t.Fatalf("docker restart count = (%d, %v), want (3, true)", got, ok)
	}
	for _, c := range rr.calls {
		if c.name == "systemctl" {
			t.Errorf("docker must not consult systemd: %+v", c)
		}
	}
}

// TestRestartCountDoesNotAttributeOurUnitToAnotherContainer. `status broker --all`
// discovers containers BY IMAGE, so it reaches brokers this env file knows nothing
// about -- and this env file's unit restart count is not their number. Reporting it
// would put a confident wrong figure next to a container it does not describe.
func TestRestartCountDoesNotAttributeOurUnitToAnotherContainer(t *testing.T) {
	m, rr, _ := newCapMgr(ctrCfg(config.Podman, "false"), config.Podman)
	rr.outFor = func(name string, args []string) []byte {
		if name == "systemctl" {
			return []byte("12\n")
		}
		return inspectReply("someone-elses-broker", 5)
	}

	s, err := m.inspectState(context.Background(), "someone-elses-broker")
	if err != nil {
		t.Fatalf("inspectState: %v", err)
	}
	got, ok := m.restartCount(context.Background(), s)
	if !ok || got != 5 {
		t.Fatalf("restart count for a foreign container = (%d, %v), want (5, true) -- the engine's own "+
			"counter, never this env file's unit", got, ok)
	}
	for _, c := range rr.calls {
		if c.name == "systemctl" {
			t.Errorf("systemd must not be asked about a container this config does not name: %+v", c)
		}
	}
}

// TestRestartCountUnknownRatherThanZero: systemd not answering (the unit is not
// installed, or a rootless/rootful mismatch) must read as unknown. A zero there
// says "this broker has never restarted", which is a claim this tool has no basis
// for making.
func TestRestartCountUnknownRatherThanZero(t *testing.T) {
	for _, tc := range []struct{ name, reply string }{
		{"not a number", "n/a\n"},
		{"empty", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, rr, buf := newCapMgr(ctrCfg(config.Podman, "false"), config.Podman)
			rr.outFor = func(name string, args []string) []byte {
				if name == "systemctl" {
					return []byte(tc.reply)
				}
				return inspectReply("sol-pod", 0)
			}
			s, err := m.inspectState(context.Background(), m.name())
			if err != nil {
				t.Fatalf("inspectState: %v", err)
			}
			if _, ok := m.restartCount(context.Background(), s); ok {
				t.Error("an unreadable NRestarts must report unknown, not a number")
			}
			m.reportState(context.Background(), s)
			if !strings.Contains(buf.String(), "unknown") {
				t.Errorf("the report should say the count is unknown:\n%s", buf)
			}
		})
	}
}

// TestReportStateNeverPrintsTheEnvironment. On docker the compose secrets are
// environment-sourced, so a report that dumped the container's environment would put
// the admin password on a terminal, into scrollback and into whatever ticket the
// output is pasted into. Secrets appear here only as mount PATHS.
func TestReportStateNeverPrintsTheEnvironment(t *testing.T) {
	const withSecrets = `[{"Name": "/solace", "State": {"Status": "running", "Running": true},
	  "Config": {"Image": "i", "Env": ["SOLACE_ADMIN_PASSWORD=super-secret-pw"]},
	  "Mounts": [{"Source": "/opt/solace/secrets/admin", "Destination": "/mnt/secrets/admin_password", "RW": false}]}]`
	m, rr, buf := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	rr.outFor = func(string, []string) []byte { return []byte(withSecrets) }

	if err := m.inspectAndReport(context.Background(), m.name()); err != nil {
		t.Fatalf("inspectAndReport: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "super-secret-pw") || strings.Contains(out, "SOLACE_ADMIN_PASSWORD") {
		t.Fatalf("the container environment reached the report:\n%s", out)
	}
	// The mount is still named, which is how a secret is meant to show up.
	if !strings.Contains(out, "/mnt/secrets/admin_password") {
		t.Errorf("the secret's mount path should be listed:\n%s", out)
	}
}

// TestReportStateOnAStoppedContainer covers the half of the report an operator only
// ever reads when something has gone wrong. A stopped container has no start time
// worth printing but does have the three facts that say why it stopped -- when it
// finished, what it exited with, and the engine's own reason -- and those are exactly
// what a running-container report has no room for.
//
// It also pins that a writable mount is labelled `rw` and a read-only one `ro`: the
// secret and certificate mounts are read-only by design, so a secret that turned up
// writable here is a real finding rather than a cosmetic one.
func TestReportStateOnAStoppedContainer(t *testing.T) {
	const stopped = `[{
	  "Name": "/solace",
	  "RestartCount": 2,
	  "State": {
	    "Status": "exited", "Running": false,
	    "FinishedAt": "2026-09-08T09:10:11.12Z",
	    "ExitCode": 137, "Error": "OOMKilled",
	    "Health": {"Status": "unhealthy", "FailingStreak": 3}
	  },
	  "Config": {"Image": "solace/solace-pubsub-standard:10.10.1.128"},
	  "Mounts": [
	    {"Source": "/opt/solace/data", "Destination": "/var/lib/solace/jail", "RW": true},
	    {"Source": "/opt/solace/certs/tls.pem", "Destination": "/mnt/certs/server/tls.pem", "RW": false}
	  ]
	}]`
	m, rr, buf := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	rr.outFor = func(string, []string) []byte { return []byte(stopped) }

	if err := m.inspectAndReport(context.Background(), m.name()); err != nil {
		t.Fatalf("inspectAndReport: %v", err)
	}
	got := buf.String()
	for _, want := range []string{
		"exited",                  // state
		"unhealthy",               // health, under docker's spelling
		"2026-09-08T09:10:11.12Z", // finished, not started
		"137",                     // exit code
		"OOMKilled",               // the engine's own reason
		"/var/lib/solace/jail (rw)",
		"/mnt/certs/server/tls.pem (ro)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("a stopped container's report should carry %q:\n%s", want, got)
		}
	}
	// A stopped container has no meaningful start time, so the row is omitted rather
	// than printed with the engine's zero value.
	if strings.Contains(got, "started") {
		t.Errorf("a stopped container should report `finished`, not `started`:\n%s", got)
	}
}

// TestReportStateOnAHealthyRunningContainer is the ordinary case, and the mirror of
// the test above: the health status is reported under docker's spelling with no
// failing-streak noise, the start time is shown, and no exit code is invented for a
// container that has not exited.
func TestReportStateOnAHealthyRunningContainer(t *testing.T) {
	m, rr, buf := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	rr.outFor = func(string, []string) []byte { return []byte(dockerInspect) }

	if err := m.inspectAndReport(context.Background(), m.name()); err != nil {
		t.Fatalf("inspectAndReport: %v", err)
	}
	got := buf.String()
	for key, want := range map[string]string{
		"state":    "running",
		"health":   "healthy",
		"restarts": "3", // docker's own counter, which is the answer on this platform
		"name":     "solbroker",
	} {
		if v := kvValue(got, key); v != want {
			t.Errorf("report %s = %q, want %q:\n%s", key, v, want, got)
		}
	}
	if kvValue(got, "started") == "" {
		t.Errorf("a running container should report when it started:\n%s", got)
	}
	if strings.Contains(got, "exit code") || strings.Contains(got, "finished") {
		t.Errorf("a running container has not exited, so neither row belongs:\n%s", got)
	}
	// "healthy" with no streak: a passing probe must not carry a "(failing streak 0)"
	// tail, which would read as a fault.
	if strings.Contains(got, "failing streak") {
		t.Errorf("a passing healthcheck should report no failing streak:\n%s", got)
	}
}

// TestInspectStateSurfacesTheEngineError: an engine that cannot be asked -- daemon
// down, socket gone -- must surface its own error rather than a parse failure about
// the empty output that follows from it. The parse error would send whoever reads it
// looking for a malformed reply that was never sent.
func TestInspectStateSurfacesTheEngineError(t *testing.T) {
	m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	rr.outErr = errListFailed

	_, err := m.inspectState(context.Background(), m.name())
	if err == nil {
		t.Fatal("an engine failure must not be reported as a successful inspect")
	}
	if !errors.Is(err, errListFailed) {
		t.Errorf("inspectState should pass the engine's own error through, got %v", err)
	}
	if strings.Contains(err.Error(), "parsing inspect output") {
		t.Errorf("an engine failure was reported as a parse failure: %v", err)
	}
}

// TestRestartCountUnknownWhenSystemctlFails is the other way systemd can decline to
// answer: the command itself fails (the unit is not installed, or a rootless manager
// is being asked about a rootful unit) rather than answering unreadably. Both must
// read as unknown, never as zero.
func TestRestartCountUnknownWhenSystemctlFails(t *testing.T) {
	m, rr, _ := newCapMgr(ctrCfg(config.Podman, "false"), config.Podman)
	rr.outFor = func(string, []string) []byte { return inspectReply("sol-pod", 0) }
	rr.outFail = failOn("systemctl")

	s, err := m.inspectState(context.Background(), m.name())
	if err != nil {
		t.Fatalf("inspectState: %v", err)
	}
	if n, ok := m.restartCount(context.Background(), s); ok {
		t.Errorf("a failed systemctl must report unknown, got %d", n)
	}
}

// kvValue reads one value out of an output.Sink KVBlock, whose rows are
// `  <key padded> : <value>`. Matching the key and taking what follows the
// separator keeps these assertions independent of the key column's width, which
// KVBlock computes from the longest key in the block -- so adding a longer field
// name later cannot break a test that is not about that field.
func kvValue(report, key string) string {
	for _, line := range strings.Split(report, "\n") {
		k, v, ok := strings.Cut(line, " : ")
		if ok && strings.TrimSpace(k) == key {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

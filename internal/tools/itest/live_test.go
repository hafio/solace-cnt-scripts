package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"solace/internal/broker"
	"solace/internal/config"
	"solace/internal/container"
	"solace/internal/engine"
	"solace/internal/output"
	"solace/internal/render"
)

// These tests cover the probe paths that talk to a live system, using a fake
// engine.Runner instead of one. What they assert is the part a live run cannot:
// the ORDER of the steps, and that cleanup still happens when a step in the
// middle fails. A probe that skipped its restore on the error path would leave a
// throwaway password on a real broker, and that is not something to discover by
// running it against one.

// fakeRunner records every command and answers from a scripted responder.
type fakeRunner struct {
	calls     []fakeCall
	responder func(name string, args []string, stdin []byte) ([]byte, error)
}

type fakeCall struct {
	method string
	name   string
	args   []string
	stdin  string
}

func (f *fakeRunner) record(method, name string, args []string, stdin []byte) ([]byte, error) {
	f.calls = append(f.calls, fakeCall{method: method, name: name, args: args, stdin: string(stdin)})
	if f.responder == nil {
		return nil, nil
	}
	return f.responder(name, args, stdin)
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) error {
	_, err := f.record("Run", name, args, nil)
	return err
}
func (f *fakeRunner) RunInput(_ context.Context, in []byte, name string, args ...string) error {
	_, err := f.record("RunInput", name, args, in)
	return err
}
func (f *fakeRunner) RunEnv(_ context.Context, _ []string, name string, args ...string) error {
	_, err := f.record("RunEnv", name, args, nil)
	return err
}
func (f *fakeRunner) RunInteractive(_ context.Context, name string, args ...string) error {
	_, err := f.record("RunInteractive", name, args, nil)
	return err
}
func (f *fakeRunner) Output(_ context.Context, name string, args ...string) ([]byte, error) {
	return f.record("Output", name, args, nil)
}
func (f *fakeRunner) OutputInput(_ context.Context, in []byte, name string, args ...string) ([]byte, error) {
	return f.record("OutputInput", name, args, in)
}

// argvOf joins a recorded call for substring matching.
func (c fakeCall) argv() string { return c.name + " " + strings.Join(c.args, " ") }

// sawInOrder reports whether the recorded calls contain each needle, in order.
func (f *fakeRunner) sawInOrder(needles ...string) bool {
	i := 0
	for _, c := range f.calls {
		if i < len(needles) && strings.Contains(c.argv(), needles[i]) {
			i++
		}
	}
	return i == len(needles)
}

func (f *fakeRunner) transcript() string {
	var b strings.Builder
	for _, c := range f.calls {
		b.WriteString(c.method + ": " + c.argv() + "\n")
	}
	return b.String()
}

// podmanEnv builds an Env wired to a fake runner, for the container probes.
// PollInterval is 0 so the waits do not sleep, exactly as the broker tests set
// Ops.PollInterval to 0.
func podmanEnv(t *testing.T, f *fakeRunner) (*Env, *bytes.Buffer) {
	t.Helper()
	cfg := &config.Config{Redundancy: "yes"}
	cfg.Podman.Runtime = config.Command{"podman"}
	cfg.Podman.Container.Name = "solace-probe"
	cfg.Podman.Container.DataDir = "/opt/solace/data"
	cfg.Admin.User, cfg.Admin.Pass = "admin", "adminpass"
	cfg.Nodes.Primary.Name, cfg.Nodes.Primary.IP = "pri-host", "10.0.0.1"
	cfg.Nodes.Backup.Name, cfg.Nodes.Backup.IP = "bkp-host", "10.0.0.2"
	cfg.Nodes.Monitor.Name, cfg.Nodes.Monitor.IP = "mon-host", "10.0.0.3"
	cfg.Nodes.PSK = "psk"

	buf := &bytes.Buffer{}
	sink := output.New(buf)
	e := &Env{
		Cfg:          cfg,
		Platform:     config.Podman,
		Runner:       f,
		Manager:      container.NewManager(f, cfg, config.Podman, sink.Line, buf),
		Ops:          broker.New(container.NewTransport(f, cfg, config.Podman), cfg, sink.Line),
		Out:          sink,
		Consent:      Consenter{Yes: true},
		PollInterval: 0,
		PollAttempts: 3,
	}
	e.Ops.Platform = config.Podman
	e.Ops.PollInterval = 0
	return e, buf
}

// --- P1 sequencing and cleanup -------------------------------------------------

// TestP1RunsStopRotateStartReadInOrder pins the experiment's shape. The order is
// the whole design: rotating before the stop, or reading before the start, would
// answer a different question than the one deployPodman's comment asks.
func TestP1RunsStopRotateStartReadInOrder(t *testing.T) {
	f := &fakeRunner{responder: func(name string, args []string, _ []byte) ([]byte, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "is-active"):
			return []byte("active\n"), nil
		case strings.Contains(joined, "cat"):
			return []byte("rotated-value\n"), nil
		}
		return nil, nil
	}}
	e, _ := podmanEnv(t, f)
	// Force the classifier's rotated branch by making the throwaway predictable:
	// the probe reads back whatever the fake says, so assert on the sequence and
	// accept whichever verdict that yields.
	res := runP1(context.Background(), e)

	if !f.sawInOrder("info", "stop", "secret create", "start", "is-active", "cat") {
		t.Errorf("P1 did not run stop -> rotate -> start -> wait -> read in order:\n%s", f.transcript())
	}
	// Whatever the verdict, the original value must have been put back.
	if !f.sawInOrder("cat", "secret create") {
		t.Errorf("P1 must restore the original secret AFTER reading:\n%s", f.transcript())
	}
	if res.Detail == "" {
		t.Error("every result must carry evidence")
	}
}

// TestP1RestoresAfterAFailureMidSequence is the important one: if starting the
// unit fails, the probe must still put the original secret value back. Otherwise
// a failed probe leaves a live broker holding a throwaway password.
func TestP1RestoresAfterAFailureMidSequence(t *testing.T) {
	f := &fakeRunner{responder: func(name string, args []string, _ []byte) ([]byte, error) {
		joined := strings.Join(args, " ")
		if strings.Contains(joined, "systemctl") && strings.Contains(joined, "start") {
			return nil, errors.New("unit failed to start")
		}
		return nil, nil
	}}
	e, buf := podmanEnv(t, f)
	res := runP1(context.Background(), e)

	if res.Verdict != Inconclusive {
		t.Errorf("a failed start is Inconclusive, not a refuted claim; got %v (%s)", res.Verdict, res.Detail)
	}
	// Two secret-create calls: the rotation, then the restore.
	var creates int
	for _, c := range f.calls {
		if strings.Contains(c.argv(), "secret create") {
			creates++
		}
	}
	if creates < 2 {
		t.Errorf("expected the rotation AND the restore (%d secret creates):\n%s", creates, f.transcript())
	}
	if !strings.Contains(buf.String(), "restoring the original value") {
		t.Errorf("the restore must be narrated so an operator can see it happened; got %q", buf.String())
	}
}

// TestP1DryRunTouchesNothing proves --dry-run stops before the first mutation:
// the preflight may run, but no stop, no secret write, no start.
func TestP1DryRunTouchesNothing(t *testing.T) {
	f := &fakeRunner{}
	e, buf := podmanEnv(t, f)
	e.DryRun = true
	res := runP1(context.Background(), e)

	if res.Verdict != Inconclusive {
		t.Errorf("a dry run settles nothing, so the verdict must be Inconclusive; got %v", res.Verdict)
	}
	for _, c := range f.calls {
		if strings.Contains(c.argv(), "stop") || strings.Contains(c.argv(), "secret create") ||
			strings.Contains(c.argv(), "systemctl start") {
			t.Errorf("--dry-run must not mutate; it ran %q", c.argv())
		}
	}
	if !strings.Contains(buf.String(), "would ") {
		t.Errorf("--dry-run must say what it would have done; got %q", buf.String())
	}
}

// TestP1DeclinedConsentTouchesNothing covers the other gate: an operator saying
// no is a SKIP, not a failure, and nothing is written.
func TestP1DeclinedConsentTouchesNothing(t *testing.T) {
	f := &fakeRunner{}
	e, _ := podmanEnv(t, f)
	e.Consent = Consenter{Interactive: func() bool { return true }, In: strings.NewReader("no\n"), Err: &bytes.Buffer{}}
	res := runP1(context.Background(), e)

	if res.Verdict != SkippedNoConsent {
		t.Errorf("verdict = %v, want SkippedNoConsent", res.Verdict)
	}
	for _, c := range f.calls {
		if strings.Contains(c.argv(), "secret create") {
			t.Errorf("a declined probe must write nothing; it ran %q", c.argv())
		}
	}
}

// TestP1WaitActiveIsBounded proves the wait gives up rather than hanging, and
// that the rootless case asks systemd the right question (--user).
func TestP1WaitActiveIsBounded(t *testing.T) {
	t.Run("gives up after PollAttempts", func(t *testing.T) {
		f := &fakeRunner{responder: func(_ string, _ []string, _ []byte) ([]byte, error) {
			return []byte("activating\n"), nil
		}}
		e, _ := podmanEnv(t, f)
		err := p1WaitActive(context.Background(), e)
		if err == nil {
			t.Fatal("a unit that never activates must produce an error, not an infinite wait")
		}
		if len(f.calls) != e.PollAttempts {
			t.Errorf("polled %d times, want exactly PollAttempts (%d)", len(f.calls), e.PollAttempts)
		}
	})
	t.Run("rootless asks the user manager", func(t *testing.T) {
		f := &fakeRunner{responder: func(_ string, _ []string, _ []byte) ([]byte, error) {
			return []byte("active\n"), nil
		}}
		e, _ := podmanEnv(t, f)
		e.Cfg.Podman.Rootless = true
		if err := p1WaitActive(context.Background(), e); err != nil {
			t.Fatalf("p1WaitActive: %v", err)
		}
		if len(f.calls) == 0 || !strings.Contains(f.calls[0].argv(), "--user") {
			t.Errorf("rootless must poll the user manager; got %q", f.transcript())
		}
	})
}

// TestP1RestoreWarnsLoudlyWhenItCannotRestore covers the worst case: the restore
// itself fails. The message has to name the secret, because a human now has to
// fix a live broker by hand.
func TestP1RestoreWarnsLoudlyWhenItCannotRestore(t *testing.T) {
	f := &fakeRunner{responder: func(_ string, args []string, _ []byte) ([]byte, error) {
		if strings.Contains(strings.Join(args, " "), "secret create") {
			return nil, errors.New("store unavailable")
		}
		return nil, nil
	}}
	e, buf := podmanEnv(t, f)
	p1Restore(context.Background(), e, render.ContainerSecret{Name: "solace-probe-admin", EnvKey: "username_admin_password", Value: "v"})
	got := buf.String()
	if !strings.Contains(got, "RESTORE FAILED") || !strings.Contains(got, "solace-probe-admin") {
		t.Errorf("a failed restore must be loud and name the secret; got %q", got)
	}
	if !strings.Contains(got, "[WARN]") {
		t.Errorf("a failed restore must carry the WARN tag; got %q", got)
	}
}

// --- P2 safety gates -----------------------------------------------------------

// TestP2RefusesWhenTheMateIsActive is the safety property the whole probe rests
// on: with the mate active, revert-activity would fail a real HA group over, so
// the probe must stop and must NOT send it.
func TestP2RefusesWhenTheMateIsActive(t *testing.T) {
	f := &fakeRunner{responder: func(_ string, args []string, _ []byte) ([]byte, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "curl"):
			return []byte("HTTP/1.1 200 OK\r\n\r\n{}"), nil
		case strings.Contains(joined, "cli"):
			return []byte("Activity Status : Mate Active\n"), nil
		}
		return nil, nil
	}}
	e, _ := podmanEnv(t, f)
	e.Ops.Hostname = func() (string, error) { return "pri-host", nil }

	res := runP2(context.Background(), e)
	if res.Verdict != Inconclusive {
		t.Fatalf("verdict = %v (%s), want Inconclusive", res.Verdict, res.Detail)
	}
	if !strings.Contains(res.Detail, "ACTIVE") {
		t.Errorf("the refusal must say why; got %q", res.Detail)
	}
	for _, c := range f.calls {
		if strings.Contains(c.stdin, "revert-activity") {
			t.Fatalf("revert-activity must NOT be sent to an active mate:\n%s", f.transcript())
		}
	}
}

// TestP2SkipsOffPrimaryAndStandalone covers the two cases where the probe has no
// business running: a non-primary host, and a standalone deployment.
func TestP2SkipsOffPrimaryAndStandalone(t *testing.T) {
	t.Run("backup host", func(t *testing.T) {
		f := &fakeRunner{}
		e, _ := podmanEnv(t, f)
		e.Ops.Hostname = func() (string, error) { return "bkp-host", nil }
		res := runP2(context.Background(), e)
		if res.Verdict != SkippedPlatform {
			t.Errorf("verdict = %v (%s), want a skip on a backup host", res.Verdict, res.Detail)
		}
		if len(f.calls) != 0 {
			t.Errorf("a skipped probe must issue no commands:\n%s", f.transcript())
		}
	})
	t.Run("standalone", func(t *testing.T) {
		f := &fakeRunner{}
		e, _ := podmanEnv(t, f)
		e.Cfg.Redundancy = "no"
		res := runP2(context.Background(), e)
		if res.Verdict != SkippedPlatform {
			t.Errorf("verdict = %v (%s), want a skip for standalone", res.Verdict, res.Detail)
		}
	})
}

// TestP2RefutesAnUnreachableMate covers the preflight failure path: the mate's
// SEMP endpoint not answering is exactly the finding this probe exists to make.
func TestP2RefutesAnUnreachableMate(t *testing.T) {
	f := &fakeRunner{responder: func(_ string, _ []string, _ []byte) ([]byte, error) {
		return nil, errors.New("curl: (7) connection refused")
	}}
	e, _ := podmanEnv(t, f)
	e.Ops.Hostname = func() (string, error) { return "pri-host", nil }
	res := runP2(context.Background(), e)
	if res.Verdict != Refuted {
		t.Errorf("verdict = %v (%s), want Refuted for an unreachable mate", res.Verdict, res.Detail)
	}
}

// TestP2DryRunStopsBeforeTheRPC proves the read-only half runs for real and the
// RPC does not.
func TestP2DryRunStopsBeforeTheRPC(t *testing.T) {
	f := &fakeRunner{responder: func(_ string, args []string, _ []byte) ([]byte, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "curl"):
			return []byte("HTTP/1.1 200 OK\r\n\r\n{}"), nil
		case strings.Contains(joined, "cli"):
			return []byte("Activity Status : Local Active\n"), nil
		}
		return nil, nil
	}}
	e, _ := podmanEnv(t, f)
	e.Ops.Hostname = func() (string, error) { return "pri-host", nil }
	e.DryRun = true

	res := runP2(context.Background(), e)
	if res.Verdict != Inconclusive {
		t.Errorf("verdict = %v, want Inconclusive under --dry-run", res.Verdict)
	}
	for _, c := range f.calls {
		if strings.Contains(c.stdin, "revert-activity") {
			t.Errorf("--dry-run must not send the RPC:\n%s", f.transcript())
		}
	}
	// but the preflight and the redundancy read DID happen
	if len(f.calls) == 0 {
		t.Error("--dry-run should still have run the read-only steps")
	}
}

// --- P3 polling and cleanup ----------------------------------------------------

// --- P4 over a fake runner -----------------------------------------------------

// TestRunP4ReadsInspectAndChangesNothing covers the read-only probe end to end:
// it must issue an inspect and nothing else that could alter the container.
func TestRunP4ReadsInspectAndChangesNothing(t *testing.T) {
	f := &fakeRunner{responder: func(_ string, args []string, _ []byte) ([]byte, error) {
		if strings.Contains(strings.Join(args, " "), "inspect") {
			return []byte(`[{"HostConfig":{"NanoCpus":2000000000,"Memory":1073741824}}]`), nil
		}
		return nil, nil
	}}
	e, _ := podmanEnv(t, f)
	e.Cfg.Scaling.CPU = "2"
	e.Cfg.Podman.Container.Mem = "1g"

	res := runP4(context.Background(), e)
	if res.Verdict != Verified {
		t.Fatalf("verdict = %v (%s), want Verified", res.Verdict, res.Detail)
	}
	for _, c := range f.calls {
		for _, forbidden := range []string{"stop", "start", "rm", "secret create", "run"} {
			if strings.Contains(c.argv(), forbidden) {
				t.Errorf("P4 is read-only but ran %q", c.argv())
			}
		}
	}
}

// TestRunP4SkipsWithNoLimitsConfigured covers the nothing-to-verify case: it must
// skip rather than claim a verdict about limits the env file never renders.
func TestRunP4SkipsWithNoLimitsConfigured(t *testing.T) {
	f := &fakeRunner{}
	e, _ := podmanEnv(t, f)
	e.Cfg.Scaling.CPU = ""
	e.Cfg.Podman.Container.Mem = ""
	if res := runP4(context.Background(), e); res.Verdict != SkippedPlatform {
		t.Errorf("verdict = %v (%s), want a skip", res.Verdict, res.Detail)
	}
}

// TestRunP4InconclusiveWhenInspectFails covers the not-deployed case: an absent
// container is not evidence that the directives are wrong.
func TestRunP4InconclusiveWhenInspectFails(t *testing.T) {
	f := &fakeRunner{responder: func(_ string, args []string, _ []byte) ([]byte, error) {
		if strings.Contains(strings.Join(args, " "), "inspect") {
			return nil, errors.New("no such container")
		}
		return nil, nil
	}}
	e, _ := podmanEnv(t, f)
	e.Cfg.Scaling.CPU = "2"
	res := runP4(context.Background(), e)
	if res.Verdict != Inconclusive {
		t.Fatalf("verdict = %v (%s), want Inconclusive", res.Verdict, res.Detail)
	}
	if !strings.Contains(res.Detail, "deployed") {
		t.Errorf("detail should suggest the likely cause; got %q", res.Detail)
	}
}

// --- resolveEnv ----------------------------------------------------------------

// minimalPodmanEnv / minimalK8sEnv are the smallest env files config.Load
// accepts for each platform -- enough to construct the entry types, no more.
const minimalPodmanEnv = `image:
  repo: solace/solace-pubsub-standard
  tag: latest
admin:
  pass: adminpass
nodes:
  primary:
    name: pri-host
podman:
  container:
    dataDir: /opt/solace/data
`

const minimalK8sEnv = `image:
  repo: solace/solace-pubsub-standard
  tag: latest
admin:
  pass: adminpass
kubernetes:
  name: solace-probe
  namespace: solace
  adminSecret: solace-admin
  storage:
    msgNode: 30Gi
`

// TestResolveEnvWiresThePlatformsEntryType proves resolveEnv builds the right
// entry type per platform and leaves the other nil -- a probe reads exactly one
// of them, so a mis-wired Env would nil-panic on a live host.
func TestResolveEnvWiresThePlatformsEntryType(t *testing.T) {
	t.Run("podman gets a Manager and no Cluster", func(t *testing.T) {
		dir, name := writeEnv(t, minimalPodmanEnv)
		var errBuf bytes.Buffer
		e, err := resolveEnv(flags{env: name, baseDir: dir}, &errBuf)
		if err != nil {
			t.Fatalf("resolveEnv: %v", err)
		}
		if e.Manager == nil || e.Cluster != nil || e.Ops == nil {
			t.Errorf("podman wiring: manager=%v cluster=%v ops=%v", e.Manager != nil, e.Cluster != nil, e.Ops != nil)
		}
		if e.Ops.Platform != config.Podman {
			t.Errorf("Ops.Platform = %q, want podman so the SEMP port resolution works", e.Ops.Platform)
		}
		if e.PollAttempts == 0 || e.PollInterval == 0 {
			t.Error("resolveEnv must set the production poll defaults")
		}
		if !strings.Contains(errBuf.String(), "==> env file:") {
			t.Errorf("the preamble belongs on stderr; got %q", errBuf.String())
		}
	})
	t.Run("kubernetes gets a Cluster and no Manager", func(t *testing.T) {
		dir, name := writeEnv(t, minimalK8sEnv)
		var errBuf bytes.Buffer
		e, err := resolveEnv(flags{env: name, baseDir: dir}, &errBuf)
		if err != nil {
			t.Fatalf("resolveEnv: %v", err)
		}
		if e.Cluster == nil || e.Manager != nil {
			t.Errorf("k8s wiring: cluster=%v manager=%v", e.Cluster != nil, e.Manager != nil)
		}
	})
	t.Run("a missing env file is an error, not a panic", func(t *testing.T) {
		if _, err := resolveEnv(flags{env: "nope.yaml", baseDir: t.TempDir()}, &bytes.Buffer{}); err == nil {
			t.Error("a missing env file must error")
		}
	})
}

// TestCtrNameReadsTheConfiguredContainer pins the helper the unit polling uses.
func TestCtrNameReadsTheConfiguredContainer(t *testing.T) {
	f := &fakeRunner{}
	e, _ := podmanEnv(t, f)
	if got := ctrName(e); got != "solace-probe" {
		t.Errorf("ctrName = %q, want the configured container name", got)
	}
}

// engineRunnerCompliance keeps fakeRunner honest: if engine.Runner grows a
// method, this fails to compile here rather than somewhere confusing.
var _ engine.Runner = (*fakeRunner)(nil)

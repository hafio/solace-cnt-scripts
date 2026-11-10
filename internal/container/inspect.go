package container

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hafio/go-deploy-utility/containercli/inspect"
	"github.com/hafio/go-deploy-utility/systemd"

	"solace/internal/config"
	"solace/internal/engine"
	"solace/internal/output"
)

// Container status is read as JSON and decoded by the shared inspect package
// (github.com/hafio/go-deploy-utility/containercli/inspect), rather than handed to the
// engine as a `--format` Go template, and the reason is that the two engines do not
// agree on the shape a template has to name. Decode smooths over each disagreement, so a
// template would get them silently wrong rather than loudly:
//
//   - `.Name` carries a LEADING SLASH on docker (`/solbroker`) and none on podman.
//   - The health block is `State.Health` on docker and `State.Healthcheck` on podman. A
//     template naming either one renders EMPTY on the other engine, which reads exactly
//     like "no healthcheck configured" -- the single most misleading thing this report
//     could say about a broker.
//   - `RestartCount` counts restarts the ENGINE performed. On podman the unit is a
//     quadlet, and systemd restarts it by replacing the container, so the engine's own
//     counter stays 0 through any number of restarts. The truthful number is systemd's
//     NRestarts (restartCount below). ASSUMED, NOT VERIFIED: podman was not testable
//     here.
//
// Decoding also lets a missing or unparseable field degrade to "(unknown)" beside the
// fields that did decode, where a template failure loses the whole line.
//
// What is NOT read is as deliberate: the environment, which Decode never decodes. On
// docker the compose secrets are environment-sourced, so a report that dumped it would
// put the admin password on a terminal, into scrollback, and into whatever ticket the
// output is pasted into. Secrets appear here only as MOUNT PATHS -- named, never read.

// health returns the health status of c, and whether a healthcheck is configured at all.
// An engine reports an unconfigured healthcheck as an empty status, which is why the two
// are separate answers: "no healthcheck" and "health unknown" must not print the same.
func health(c inspect.Container) (string, bool) {
	h := c.State.Health
	switch {
	case h.Status == "":
		return "", false
	case h.FailingStreak > 0:
		return fmt.Sprintf("%s (failing streak %d)", h.Status, h.FailingStreak), true
	}
	return h.Status, true
}

// inspectContainer captures and decodes one container's inspect output. An empty array
// is not an error from the engine -- it is what a name it does not know produces -- so
// it is reported as one here, since every caller asked about a container it believed
// existed. Under the Echo runner Output returns nothing, so the decode fails and the
// caller reports the argv it would have run -- which is exactly what the wiring tests
// assert.
func (m *Manager) inspectContainer(ctx context.Context, name string) (inspect.Container, error) {
	base, err := m.base()
	if err != nil {
		return inspect.Container{}, err
	}
	c, _ := inspect.Containers(base, name)
	out, err := engine.Output(ctx, m.R, c)
	if err != nil {
		return inspect.Container{}, err
	}
	cs, err := inspect.Decode(string(out))
	if err != nil {
		return inspect.Container{}, err
	}
	if len(cs) == 0 {
		return inspect.Container{}, fmt.Errorf("inspect returned no container named %q", name)
	}
	return cs[0], nil
}

// restartCount answers "how many times has this broker restarted" per platform.
//
// On docker the engine's own counter is the answer. On podman it is not: a quadlet
// unit is restarted by systemd, which stops and REPLACES the container, so the
// engine's counter is 0 on a unit that has restarted twenty times. systemd's
// NRestarts is the number that means what the operator is asking. A systemctl that
// cannot answer (the unit is not installed, or this is a rootful/rootless mismatch)
// yields ok=false rather than a wrong zero.
//
// systemd is consulted ONLY for the container this env file names. `status broker
// --all` discovers containers by image, so it can reach a broker this config knows
// nothing about -- and this env file's unit restart count is not that container's
// number. There the engine's own counter is reported, understating a systemd
// restart rather than attributing the wrong unit's.
func (m *Manager) restartCount(ctx context.Context, s inspect.Container) (int, bool) {
	if m.P != config.Podman || s.Name != m.name() {
		return s.RestartCount, true
	}
	out, err := m.systemctlOutput(ctx, systemd.NRestarts(m.systemctlUser(), m.name()+".service"))
	if err != nil {
		return 0, false
	}
	// ParseNRestarts refuses a negative or non-numeric answer, and systemd answers a
	// unit it does not know with an empty value: either way the count is unknown, and
	// its text is not printed.
	n, err := systemd.ParseNRestarts(string(out))
	if err != nil {
		return 0, false
	}
	return n, true
}

// stateRows renders the decoded state as the report block both `status broker` and
// `status broker --all --detail` print. Fields that do not apply are omitted rather
// than shown empty: a stopped container has no health to report, and a running one
// has no exit code worth printing.
func (m *Manager) stateRows(ctx context.Context, s inspect.Container) []output.KV {
	rows := []output.KV{
		{Key: "name", Value: inspect.Printable(s.Name)},
		{Key: "image", Value: inspect.Printable(s.Image)},
		{Key: "state", Value: orUnknown(inspect.Printable(s.State.Status))},
	}
	if status, configured := health(s); configured {
		rows = append(rows, output.KV{Key: "health", Value: status})
	} else {
		rows = append(rows, output.KV{Key: "health", Value: "(no healthcheck configured)"})
	}
	if n, ok := m.restartCount(ctx, s); ok {
		rows = append(rows, output.KV{Key: "restarts", Value: strconv.Itoa(n)})
	} else {
		// Podman only, and worth saying why rather than printing a zero this tool
		// has no basis for: systemd owns the count and systemd did not answer.
		rows = append(rows, output.KV{Key: "restarts", Value: "(unknown: systemd did not report NRestarts)"})
	}
	if s.State.Running {
		rows = append(rows, output.KV{Key: "started", Value: orUnknown(inspect.Printable(s.State.StartedAt))})
	} else {
		rows = append(rows,
			output.KV{Key: "finished", Value: orUnknown(inspect.Printable(s.State.FinishedAt))},
			output.KV{Key: "exit code", Value: strconv.Itoa(s.State.ExitCode)})
		if s.State.Error != "" {
			rows = append(rows, output.KV{Key: "error", Value: inspect.Printable(s.State.Error)})
		}
	}
	return rows
}

// reportState prints the state block plus the mount list. Mounts are how a
// broker's secrets show up and deliberately the only way they do -- they live in
// files under /mnt/secrets/<setting>, so listing the mount names every one of them
// without reading any.
//
// The mount COUNT is a row of the key/value block and the mounts themselves are
// indented lines beneath it, rather than a bare `mounts :` header. Two reasons, both
// internal/output conventions: KVBlock computes its key column from the rows it is
// given, so a header printed separately would have to hand-count the padding and
// would go ragged the moment a field name changes; and an empty value would leave a
// trailing space on a line that gets captured into dev-script logs and grepped.
func (m *Manager) reportState(ctx context.Context, s inspect.Container) {
	r := m.report()
	rows := m.stateRows(ctx, s)
	if len(s.Mounts) == 0 {
		r.KVBlock(append(rows, output.KV{Key: "mounts", Value: "(none)"}))
		return
	}
	r.KVBlock(append(rows, output.KV{Key: "mounts", Value: strconv.Itoa(len(s.Mounts))}))
	for _, mt := range s.Mounts {
		mode := "ro"
		if mt.RW {
			mode = "rw"
		}
		r.Line("    - %s -> %s (%s)", inspect.Printable(mt.Source), inspect.Printable(mt.Destination), mode)
	}
}

// inspectAndReport is the whole sequence for one container: capture, decode,
// print. A failure is reported against the name it was asked about and returned, so
// a caller looping over several containers can carry on with the rest.
func (m *Manager) inspectAndReport(ctx context.Context, name string) error {
	s, err := m.inspectContainer(ctx, name)
	if err != nil {
		return err
	}
	m.reportState(ctx, s)
	return nil
}

func orUnknown(s string) string { return orValue(s, "(unknown)") }

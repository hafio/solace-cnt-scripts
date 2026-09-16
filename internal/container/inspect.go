package container

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"solace/internal/config"
	"solace/internal/output"
)

// Container status is read as JSON and decoded HERE, rather than handed to the
// engine as a `--format` Go template, and the reason is that the two engines do not
// agree on the shape a template has to name. Three disagreements, each of which a
// template gets silently wrong rather than loudly:
//
//   - `.Name` carries a LEADING SLASH on docker (`/solbroker`) and none on podman.
//     A template printing it renders a name that is not the container's name on one
//     of the two engines, and `/solbroker` is not a name either engine accepts back.
//   - The health block is `State.Health` on docker and `State.Healthcheck` on
//     podman. A template naming either one renders EMPTY on the other engine, which
//     reads exactly like "no healthcheck configured" -- the single most misleading
//     thing this report could say about a broker. NEEDS VERIFICATION ON A LIVE
//     PODMAN HOST: both known spellings are decoded, so this is only load-bearing
//     if podman uses a third name. Unverified.
//   - `RestartCount` counts restarts the ENGINE performed. On podman the unit is a
//     quadlet, and systemd restarts it by replacing the container, so the engine's
//     own counter stays 0 through any number of restarts. The truthful number is
//     systemd's NRestarts. ASSUMED, NOT VERIFIED: podman was not testable here, and
//     if this is wrong in either direction the report gives a confident number for
//     the wrong thing. Unverified.
//
// Decoding also lets a missing or unparseable field degrade to "(unknown)" beside
// the fields that did decode, where a template failure loses the whole line.
//
// What is NOT read is as deliberate: the environment. On docker the compose secrets
// are environment-sourced, so a report that dumped it would put the admin password
// on a terminal, into scrollback, and into whatever ticket the output is pasted
// into. Secrets appear here only as MOUNT PATHS -- named, never read.

// containerState is the subset of `<runtime> inspect` this tool reports. Every
// field is optional as far as the decoder is concerned: an engine version that
// stops emitting one leaves it zero rather than failing the command.
type containerState struct {
	Name         string
	RestartCount int
	State        struct {
		Status     string
		Running    bool
		StartedAt  string
		FinishedAt string
		ExitCode   int
		Error      string
		// Both spellings, decoded together: docker emits Health, podman emits
		// Healthcheck, and health() takes whichever answered.
		Health      healthState
		Healthcheck healthState
	}
	Config struct {
		Image string
	}
	Mounts []struct {
		Source      string
		Destination string
		RW          bool
	}
}

// healthState is the health block both engines nest under their own spelling.
type healthState struct {
	Status        string
	FailingStreak int
}

// containerName strips docker's leading slash. Podman's name has none, so this is
// idempotent on it -- one function rather than a per-engine branch, because the
// value is being normalised, not interpreted.
func (s containerState) containerName() string { return strings.TrimPrefix(s.Name, "/") }

// health returns the health status under whichever spelling this engine used, and
// whether a healthcheck is configured at all. An engine reports an unconfigured
// healthcheck as an empty status, which is why the two are separate answers: "no
// healthcheck" and "health unknown" must not print the same.
func (s containerState) health() (string, bool) {
	for _, h := range []healthState{s.State.Health, s.State.Healthcheck} {
		if h.Status == "" {
			continue
		}
		if h.FailingStreak > 0 {
			return fmt.Sprintf("%s (failing streak %d)", h.Status, h.FailingStreak), true
		}
		return h.Status, true
	}
	return "", false
}

// decodeInspect reads the array `<runtime> inspect` prints and returns its first
// element. An empty array is not an error from the engine -- it is what a name it
// does not know produces -- so it is reported as one here, since every caller asked
// about a container it believed existed.
func decodeInspect(raw []byte) (containerState, error) {
	var states []containerState
	if err := json.Unmarshal(raw, &states); err != nil {
		return containerState{}, fmt.Errorf("parsing inspect output: %w", err)
	}
	if len(states) == 0 {
		return containerState{}, fmt.Errorf("inspect returned no container")
	}
	return states[0], nil
}

// inspectState captures and decodes one container's inspect output. Under the Echo
// runner Output returns nothing, so the decode fails and the caller reports the
// argv it would have run -- which is exactly what the wiring tests assert.
func (m *Manager) inspectState(ctx context.Context, name string) (containerState, error) {
	raw, err := m.output(ctx, "inspect", name)
	if err != nil {
		return containerState{}, err
	}
	return decodeInspect(raw)
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
func (m *Manager) restartCount(ctx context.Context, s containerState) (int, bool) {
	if m.P != config.Podman || s.containerName() != m.name() {
		return s.RestartCount, true
	}
	out, err := m.systemctlOutput(ctx, "show", "-p", "NRestarts", "--value", m.name()+".service")
	if err != nil {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, false
	}
	return n, true
}

// stateRows renders the decoded state as the report block both `status broker` and
// `status broker --all --detail` print. Fields that do not apply are omitted rather
// than shown empty: a stopped container has no health to report, and a running one
// has no exit code worth printing.
func (m *Manager) stateRows(ctx context.Context, s containerState) []output.KV {
	rows := []output.KV{
		{Key: "name", Value: printable(s.containerName())},
		{Key: "image", Value: printable(s.Config.Image)},
		{Key: "state", Value: orUnknown(printable(s.State.Status))},
	}
	if status, configured := s.health(); configured {
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
		rows = append(rows, output.KV{Key: "started", Value: orUnknown(s.State.StartedAt)})
	} else {
		rows = append(rows,
			output.KV{Key: "finished", Value: orUnknown(s.State.FinishedAt)},
			output.KV{Key: "exit code", Value: strconv.Itoa(s.State.ExitCode)})
		if s.State.Error != "" {
			rows = append(rows, output.KV{Key: "error", Value: printable(s.State.Error)})
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
func (m *Manager) reportState(ctx context.Context, s containerState) {
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
		r.Line("    - %s -> %s (%s)", printable(mt.Source), printable(mt.Destination), mode)
	}
}

// inspectAndReport is the whole sequence for one container: capture, decode,
// print. A failure is reported against the name it was asked about and returned, so
// a caller looping over several containers can carry on with the rest.
func (m *Manager) inspectAndReport(ctx context.Context, name string) error {
	s, err := m.inspectState(ctx, name)
	if err != nil {
		return err
	}
	m.reportState(ctx, s)
	return nil
}

// printable strips control characters from a string the ENGINE reported before it
// reaches a terminal: an image name, a status, an error text or a mount path is data
// about a container this tool did not necessarily create, and a terminal escape inside
// one would be executed by the terminal reading the report.
func printable(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, s)
}

func orUnknown(s string) string { return orValue(s, "(unknown)") }

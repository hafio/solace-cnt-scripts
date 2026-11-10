package container

import (
	"context"

	"github.com/hafio/go-deploy-utility/containercli"

	"solace/internal/engine"
)

// Preflight is the read-only probe every mutating operation runs first. `<runtime>
// info` is the daemon-style equivalent of the cluster's `auth can-i`: it answers
// with one round-trip whether this process can actually talk to the engine, which
// on docker means the daemon is up and this user is in the right group, and on
// rootless podman means the user session and its subuid mapping exist.
//
// It runs before the first artifact is written, not after. A deploy that renders a
// compose file or installs a quadlet unit and only then discovers a stopped daemon
// leaves host state the operator now has to reason about; failing first keeps
// "nothing happened" true.
//
// There is deliberately no skip flag. The one caller that has no engine to probe is the
// Echo runner, which is a TEST seam rather than anything an operator can reach -- there is
// no --dry-run, and `generate` is how you look before leaping -- so it takes the branch
// below and nothing user-facing needs a flag.
// It never starts the daemon or logs anyone in: both are the operator's decisions,
// made with privileges this tool should not be exercising on their behalf.
func (m *Manager) Preflight(ctx context.Context) error {
	rt, err := m.runtime()
	if err != nil {
		return err
	}
	probe := containercli.Info(rt.Argv())
	if m.isEcho() {
		// Echo the probe so the preview shows it, then skip the assertion: the
		// Echo runner answers nothing, and there is no engine to answer.
		if _, err := engine.Output(ctx, m.R, probe); err != nil {
			return err
		}
		m.report().KVRow(reportKeyWidth, "engine", "skipped (preview)")
		return nil
	}
	if _, err := engine.Output(ctx, m.R, probe); err != nil {
		// The engine's own message already went to stderr (out is empty for that
		// reason); add the one line that says what to do about it, which differs by
		// engine and rootless mode and is shared with solmq-conn-util.
		return containercli.InfoError(containercli.Engine(m.P), rt, !m.Cfg.Podman.Rootful, "", err)
	}
	return nil
}

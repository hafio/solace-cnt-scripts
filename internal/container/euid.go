package container

import (
	"github.com/hafio/go-deploy-utility/rootless"
	"github.com/hafio/go-deploy-utility/runner"

	"solace/internal/config"
	"solace/internal/engine"
)

// GuardPodmanEUID checks the DECLARED podman.rootful against the account actually
// running this command, through the shared rootless.Decide: rootless is the default
// and runs as an ordinary user, rootful is deployed only when asked for and needs
// root, and root without podman.rootful is refused.
//
// podman.rootful is declared rather than detected from the euid, and this is the
// other half of that decision. The value decides where the unit is installed, which
// systemd instance loads it, which uid:gid the container runs as and where the
// server private key lands -- so one env file has to render one artifact from any
// account and any OS, which is what `broker generate` promises. Detecting the mode
// instead would make the artifact a property of whichever shell rendered it, and
// would silently turn `sudo` into a second, rootful deployment rather than a
// refusal. So the file states the intent and this compares reality against it.
//
// It runs at the CLI boundary (cli.prepare), before ANY podman command -- version,
// info, exec, cp -- so a mismatch never probes, deploys into, or reads the wrong
// account's engine. `broker generate` is exempt: it is render-only and reaches no
// host at all.
//
// Three cases pass without asking: docker, which has no rootless mode in this
// schema; the Echo runner, whose preview reaches no host; and a platform with no
// POSIX euid (Windows returns -1), where there is nothing to compare against and
// the artifact still has to render (Decide keeps the setting for it).
func GuardPodmanEUID(cfg *config.Config, p config.Platform, r runner.Runner, euid int) error {
	if p != config.Podman {
		return nil
	}
	if engine.IsEcho(r) {
		return nil
	}
	_, err := rootless.Decide(euid, cfg.Podman.Rootful, podmanRootfulKey)
	return err
}

// podmanRootfulKey is how this tool names the setting in its env file, which is what
// rootless.Decide's refusals quote.
const podmanRootfulKey = "podman.rootful"

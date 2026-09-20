package container

import (
	"fmt"

	"solace/internal/config"
	"solace/internal/engine"
)

// GuardPodmanEUID checks the DECLARED podman.rootless against the account actually
// running this command: rootless must not be root, rootful must be.
//
// podman.rootless is declared rather than detected from the euid, and this is the
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
// the artifact still has to render.
func GuardPodmanEUID(cfg *config.Config, p config.Platform, r engine.Runner, euid int) error {
	if p != config.Podman {
		return nil
	}
	if _, echo := r.(engine.Echo); echo {
		return nil
	}
	if euid < 0 {
		return nil
	}
	if cfg.Podman.Rootless && euid == 0 {
		return fmt.Errorf("podman.rootless=true but running as root (euid 0); run as the rootless user itself, " +
			"without sudo -- or set podman.rootless=false to deploy the rootful engine")
	}
	if !cfg.Podman.Rootless && euid != 0 {
		return fmt.Errorf("podman.rootless=false requires root, but running as euid %d; re-run with sudo -- "+
			"or set podman.rootless=true to deploy as this user", euid)
	}
	return nil
}

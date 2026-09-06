package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"solace/internal/broker"
	"solace/internal/config"
	"solace/internal/container"
	"solace/internal/engine"
	"solace/internal/k8s"
	"solace/internal/output"
)

// resolveEnv turns the operator's flags into the live environment the probes
// drive. It mirrors internal/cli's resolvePlatform/prepare rather than importing
// them: those are cobra PreRunE hooks entangled with flag annotations this binary
// has no use for, and the whole of what is needed here is twenty lines over
// config's exported API.
//
// The one deliberate divergence from the CLI: an env file declaring several
// platforms is a hard error naming --platform, never an interactive prompt. This
// binary is built to be copied to a host and run, often unattended; a tool that
// blocked on a question there would hang a probe run rather than fail it.
func resolveEnv(f flags, errOut io.Writer) (*Env, error) {
	path, err := config.ResolveEnvPath(f.baseDir, f.env)
	if err != nil {
		return nil, err
	}
	platform, err := resolvePlatform(path, f.platform)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(path, platform, f.allowCommand...)
	if err != nil {
		return nil, err
	}

	sink := output.New(errOut)
	runner := engine.NewExec(errOut, f.verbose)
	e := &Env{
		Cfg:      cfg,
		Platform: platform,
		Runner:   runner,
		Out:      sink,
		DryRun:   f.dryRun,
		Consent:  Consenter{Yes: f.yes},
		// 2s x 30 = a minute of patience for a unit restart or a config-sync
		// propagation, matching broker.New's own bounded-wait defaults.
		PollInterval: 2 * time.Second,
		PollAttempts: 30,
	}
	// One of the two entry types, matching the platform -- the same wiring
	// internal/cli does in k8sCluster/k8sOps and ctrManager/ctrOps, minus the
	// *App. Progress goes to stderr; a probe's own report lines go through Out.
	switch platform {
	case config.K8s:
		e.Cluster = k8s.NewCluster(runner, cfg, sink.Line, errOut)
		e.Ops = broker.New(k8s.NewTransport(runner, cfg), cfg, sink.Line)
	default:
		e.Manager = container.NewManager(runner, cfg, platform, sink.Line, errOut)
		e.Manager.EnvPath = path
		e.Ops = broker.New(container.NewTransport(runner, cfg, platform), cfg, sink.Line)
		e.Ops.Platform = platform
	}
	sink.Step("env file: %s", path)
	sink.Step("platform: %s", platform)
	return e, nil
}

// resolvePlatform picks the platform: --platform when given (and it must name a
// section the file declares, exactly as the CLI insists), otherwise the sole
// declared section.
func resolvePlatform(path, flag string) (config.Platform, error) {
	declared, err := config.DetectPlatforms(path)
	if err != nil {
		return "", err
	}
	if flag != "" {
		p, err := config.ParsePlatform(flag)
		if err != nil {
			return "", err
		}
		if !containsPlatform(declared, p) {
			return "", fmt.Errorf("--platform %s, but %s declares no %s: section (it declares %s)",
				flag, path, p, platformNames(declared))
		}
		return p, nil
	}
	switch len(declared) {
	case 1:
		return declared[0], nil
	case 0:
		return "", fmt.Errorf("%s declares no platform section; add one of kubernetes:, docker: or podman:", path)
	default:
		return "", fmt.Errorf("%s declares more than one platform (%s); pass --platform to choose "+
			"(this tool never prompts: it is meant to run unattended on a target host)",
			path, platformNames(declared))
	}
}

func containsPlatform(ps []config.Platform, want config.Platform) bool {
	for _, p := range ps {
		if p == want {
			return true
		}
	}
	return false
}

func platformNames(ps []config.Platform) string {
	names := make([]string, 0, len(ps))
	for _, p := range ps {
		names = append(names, string(p))
	}
	return strings.Join(names, ", ")
}

// wouldMutate is the --dry-run gate every probe calls immediately before its
// first mutating step. It reports true when the run must stop there.
//
// A dry run deliberately does NOT swap in engine.Echo and carry on: half the
// probe would then run for real and half would be echoed, and the verdict at the
// end would describe neither. Everything read-only up to this point HAS already
// run -- the preflights, the role detection, the redundancy read -- so the
// preview is backed by the real state of the environment, and the probe stops
// honestly without a verdict rather than claiming one it did not earn.
func (e *Env) wouldMutate(what string) bool {
	if !e.DryRun {
		return false
	}
	e.Out.Info("dry-run: would %s", what)
	return true
}

// ctrName is this host's broker container/unit name. Read straight from config
// rather than through Manager, whose own accessor is unexported -- the harness
// needs the name for the systemd unit it polls, which is not reason enough to
// widen the Manager's public surface.
func ctrName(e *Env) string { return e.Cfg.ContainerBlock(e.Platform).Name }

// roleLabel is the long role name, for report lines an operator reads. config.Role
// is the single letter (p|b|m) that pod names use, and k8s's own long-name helper
// is unexported -- a four-line mapping here is cheaper than widening that package
// for a formatting concern.
func roleLabel(r config.Role) string {
	switch r {
	case config.Backup:
		return "backup"
	case config.Monitor:
		return "monitor"
	default:
		return "primary"
	}
}

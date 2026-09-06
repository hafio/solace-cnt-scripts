package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"solace/internal/config"
	"solace/internal/container"
	"solace/internal/render"
)

// probeP1 settles the ASSUMED, NOT VERIFIED branch in container.deployPodman:
// that a quadlet replaces its container on each service start, so starting a
// stopped unit re-reads podman's secret store and picks up a rotated value.
//
// It matters because docker provably does NOT behave that way -- deployDocker
// force-recreates a stopped container precisely because a plain start replays the
// credentials it was created with. If podman shares that behaviour, then
// `deploy broker --restart` silently fails to apply a rotated password on
// podman, and an operator would believe a rotation took effect when it did not.
//
// The probe is the smallest experiment that answers it: stop the unit, rotate ONE
// secret through the real CreatePodmanSecrets path, start the unit, then read the
// mounted file from inside the container. Whatever the file says is the answer.
func probeP1() Probe {
	return Probe{
		ID:        "p1",
		Name:      "podman quadlet re-reads a rotated secret on unit start",
		Claim:     "internal/container/manager.go deployPodman -- ASSUMED, NOT VERIFIED",
		Platforms: []config.Platform{config.Podman},
		Mutates:   true,
		Touches: "stops the broker unit, rotates the admin-password secret to a throwaway value, " +
			"starts the unit, reads the mounted secret, then restores the original value and bounces again",
		Run: runP1,
	}
}

func runP1(ctx context.Context, e *Env) Result {
	m := e.Manager
	if err := m.Preflight(ctx); err != nil {
		// Loud, not skipped: the probe was selected for this platform, so an
		// unreachable engine is a failure of the run, not an inapplicable probe.
		return inconclusive("podman preflight failed, nothing was touched: %v", err)
	}

	secrets := render.ContainerSecrets(e.Cfg, config.Podman)
	if len(secrets) == 0 {
		return skipped("this env file defines no container secrets to rotate")
	}
	target := secrets[0] // the admin password; ContainerSecrets' order is fixed
	if target.Value == "" {
		return skipped("secret %q has no value in this env file (set %s)", target.Name, target.ConfigKey)
	}

	what := fmt.Sprintf("stop %s, rotate secret %q, start it again, then restore the original value",
		ctrName(e)+".service", target.Name)
	if e.wouldMutate(what) {
		return inconclusive("dry-run: stopped before the first mutation")
	}
	if !e.Consent.Confirm("About to " + what + ".") {
		return Result{SkippedNoConsent, "operator declined the rotation"}
	}

	throwaway, err := p1Throwaway()
	if err != nil {
		return inconclusive("could not generate a throwaway secret value: %v", err)
	}

	// Restore runs whatever happens next. A probe that left a throwaway password
	// on a live broker would be a worse bug than the one it is investigating.
	defer p1Restore(ctx, e, target)

	if err := m.Stop(ctx); err != nil {
		return inconclusive("stopping the unit failed, nothing was rotated: %v", err)
	}
	rotated := target
	rotated.Value = throwaway
	if err := m.CreatePodmanSecrets(ctx, []render.ContainerSecret{rotated}); err != nil {
		return inconclusive("rotating secret %q failed: %v", target.Name, err)
	}
	if err := m.Start(ctx); err != nil {
		return inconclusive("starting the unit after the rotation failed: %v", err)
	}
	if err := p1WaitActive(ctx, e); err != nil {
		return inconclusive("the unit did not become active after the rotation: %v", err)
	}

	got, err := p1ReadMountedSecret(ctx, e, target)
	if err != nil {
		return inconclusive("could not read %s inside the container: %v", target.MountPath(), err)
	}
	return p1Classify(got, throwaway, target.Value)
}

// p1Classify is the verdict rule, kept a pure function of the three strings so it
// is testable without podman.
func p1Classify(mounted, rotated, original string) Result {
	switch strings.TrimSpace(mounted) {
	case strings.TrimSpace(rotated):
		return verified("the mounted secret carried the rotated value: quadlet re-reads the store on unit start")
	case strings.TrimSpace(original):
		return refuted("the mounted secret still carried the PRE-rotation value: a plain systemctl start does " +
			"NOT pick up a rotated secret, so deployPodman's not-running branch needs docker's force-recreate " +
			"equivalent (restart when secrets changed, not start)")
	default:
		return inconclusive("the mounted secret matched neither the rotated nor the original value; " +
			"cannot conclude either way (is another secret mounted at this path?)")
	}
}

// p1Throwaway generates the temporary value. Base64 of 24 random bytes: a
// password-shaped string with no character the broker CLI treats specially, so a
// refuted verdict cannot be blamed on the value itself.
func p1Throwaway() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "itest" + base64.RawURLEncoding.EncodeToString(b), nil
}

// p1WaitActive polls `systemctl is-active` until the unit reports active.
func p1WaitActive(ctx context.Context, e *Env) error {
	unit := ctrName(e) + ".service"
	args := []string{"is-active", unit}
	if e.Cfg.Podman.Rootless {
		args = append([]string{"--user"}, args...)
	}
	var last error
	for i := 0; i < e.PollAttempts; i++ {
		out, err := e.Runner.Output(ctx, "systemctl", args...)
		if err == nil && strings.TrimSpace(string(out)) == "active" {
			return nil
		}
		last = err
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(e.PollInterval):
		}
	}
	return fmt.Errorf("%s never reported active after %d attempts (last error: %v)", unit, e.PollAttempts, last)
}

// p1ReadMountedSecret cats the secret's mount path from inside the container,
// through the same node-local transport the config/verify ops use.
func p1ReadMountedSecret(ctx context.Context, e *Env, s render.ContainerSecret) (string, error) {
	t := container.NewTransport(e.Runner, e.Cfg, config.Podman)
	out, err := t.Output(ctx, config.Primary, "cat", s.MountPath())
	return string(out), err
}

// p1Restore puts the original secret value back and bounces the unit so the
// running broker ends the probe on the credential it started with. Failures are
// warned about loudly and name the secret, because the operator has to fix them
// by hand.
func p1Restore(ctx context.Context, e *Env, original render.ContainerSecret) {
	e.Out.Info("restoring the original value of secret %q", original.Name)
	if err := e.Manager.CreatePodmanSecrets(ctx, []render.ContainerSecret{original}); err != nil {
		e.Out.Warn("RESTORE FAILED: secret %q may still hold the throwaway value -- "+
			"re-run `deploy broker --restart`, or set it back by hand: %v", original.Name, err)
		return
	}
	if err := e.Manager.Restart(ctx); err != nil {
		e.Out.Warn("the original secret value was restored but restarting %s failed; "+
			"the running container may still hold the throwaway value: %v", ctrName(e)+".service", err)
		return
	}
	e.Out.OK("original secret value restored and the unit bounced")
}

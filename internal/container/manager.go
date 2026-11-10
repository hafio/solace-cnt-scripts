package container

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/hafio/go-deploy-utility/containercli"
	"github.com/hafio/go-deploy-utility/containercli/inspect"
	"github.com/hafio/go-deploy-utility/podman"
	"github.com/hafio/go-deploy-utility/runner"
	"github.com/hafio/go-deploy-utility/systemd"

	"solace/internal/config"
	"solace/internal/engine"
	"solace/internal/output"
	"solace/internal/render"
)

// reportKeyWidth is the key column shared by the whole `check` report, matching
// k8s.reportKeyWidth for the same reason: CheckEnv writes what is known locally,
// then the engine, nofile and DNS probes each add a line of their own, so no one
// call site sees every key.
const reportKeyWidth = 15

// Manager deploys and operates the single broker container on THIS host -- the
// container analog of k8s.Cluster. Where k8s.Cluster drives every pod in a
// namespace from one control point, a container host runs one broker, so the
// Manager's operations are node-local and the HA coordination (the redundancy
// test) is a per-host handshake handled in package broker. Every mutating
// command routes through the runner.Runner, so the Echo runner records without running.
// Manager is the one holder that may pass environment to a child -- the compose deploy
// -- so R is the runner as given; every other holder wraps its runner in engine.NoEnv.
type Manager struct {
	R   runner.Runner
	Cfg *config.Config
	P   config.Platform
	// Log is the RAW line sink for progress: one already-formatted line, emitted
	// verbatim to stderr by the caller. The `==> ` and `[TAG ] ` prefixes come
	// from the internal/output Sink built over it (progress below), never from a
	// call site. nil discards.
	Log func(format string, args ...any)
	Out io.Writer // user-facing output; nil -> os.Stdout
	In  io.Reader // reserved for prompts; nil -> os.Stdin

	// Resolve reports whether a hostname resolves; a seam over net.LookupHost so
	// Check/PrepHost DNS probes are testable. NewManager sets the default.
	Resolve func(host string) bool
	// Geteuid reports this process's effective uid; a seam over os.Geteuid so the
	// podman rootless/rootful guard is testable off a POSIX host (Windows returns
	// -1). NewManager sets the default.
	Geteuid func() int
	// Getenv and Setenv are this process's own environment; seams over os.Getenv
	// and os.Setenv so ensureUserSession's XDG_RUNTIME_DIR / DBUS_SESSION_BUS_ADDRESS
	// derivation is testable without mutating the test binary's environment.
	// NewManager sets the defaults.
	Getenv func(key string) string
	Setenv func(key, value string) error

	// The memoised result of ensureUserSession (rootless.go). Every systemctl call
	// funnels through it, so it must probe once per run rather than once per
	// command; sessionDesc is what the `user session` report row prints.
	sessionOnce bool
	sessionErr  error
	sessionDesc string
	// RestartApproved is --restart on `deploy broker`: it pre-approves restarting or
	// recreating a running broker whose artifact or secrets changed, so nothing is
	// asked, and it also recreates one that did not change. It is an approval, not
	// an instruction for any other command: Restart() is the one that bounces a
	// broker on purpose.
	RestartApproved bool
	// Confirm asks the operator a yes/no question before a deploy restarts or
	// recreates a running broker. An error is a question that could not be asked (no
	// terminal and no --no-prompt) and fails the deploy; (false, nil) is a "no", which
	// leaves the host as it was. nil declines, which is what a Manager built without
	// the CLI must do: re-deploying should never bounce a running broker unattended.
	// The CLI wires it to the shared confirmation gate, where --no-prompt approves.
	Confirm func(question string) (bool, error)
	// SysctlDropIn is where a raised fs.nr_open is persisted (limits.go). A field
	// rather than a constant so a test can point it at a temp directory instead of
	// writing to /etc. NewManager sets the default.
	SysctlDropIn string
}

// NewManager builds a container host Manager over the given runner, config and
// platform. It defaults Resolve to a net.LookupHost seam and Geteuid to os.Geteuid.
// In is left for the caller to set.
//
// There is deliberately no PSK generator here any more. This tool used to invent the
// redundancy pre-shared key on a first HA deploy and rewrite the env file it was handed;
// the key is the operator's to make now (config.validateContainerPSK refuses an empty one
// and prints the openssl command), and nothing in this package writes to an env file.
func NewManager(r runner.Runner, cfg *config.Config, p config.Platform, log func(string, ...any), out io.Writer) *Manager {
	return &Manager{
		R:       r,
		Cfg:     cfg,
		P:       p,
		Log:     log,
		Out:     out,
		Resolve: func(host string) bool { _, err := net.LookupHost(host); return err == nil },
		Geteuid: os.Geteuid,
		Getenv:  os.Getenv,
		Setenv:  os.Setenv,

		SysctlDropIn: defaultSysctlDropIn,
	}
}

// progress is the stderr Sink for this package's narration: phases through Step,
// leveled status through OK/Warn/Fail/Info. A nil Log discards.
func (m *Manager) progress() *output.Sink { return output.NewFunc(m.Log) }

// report is the stdout Sink for report bodies -- sections, key/value blocks and
// the per-item outcome lines that belong to a report rather than the narration.
func (m *Manager) report() *output.Sink { return output.New(m.out()) }

// logf announces one phase of work (`==> ...`) via the injected line sink.
func (m *Manager) logf(format string, args ...any) { m.progress().Step(format, args...) }

// out returns the user-facing sink, defaulting to os.Stdout.
func (m *Manager) out() io.Writer {
	if m.Out != nil {
		return m.Out
	}
	return os.Stdout
}

// isEcho reports whether the runner only echoes (runner.Echo), so file writes
// and DNS/euid probes are previewed rather than performed.
func (m *Manager) isEcho() bool { return engine.IsEcho(m.R) }

// runtime is the guarded runtime command (docker.command / podman.command): argv[0]
// plus any leading arguments that precede every call's own. It re-runs
// execguard.Check on every call -- the Manager is built straight from a
// *config.Config, so it is the executor half of the guard's two enforcement points
// and must not assume Validate ever ran (execguard.go).
func (m *Manager) runtime() (config.Command, error) { return m.Cfg.RuntimeCommand(m.P) }
func (m *Manager) name() string                     { return m.Cfg.ContainerBlock(m.P).Name }

// run executes a runtime subcommand (`<runtime> args...`) through the Runner.
func (m *Manager) run(ctx context.Context, args ...string) error {
	r, err := m.runtime()
	if err != nil {
		return err
	}
	return engine.Run(ctx, m.R, runner.Cmd{Argv: r.Argv(args...)})
}

// base is the guarded runtime argv the containercli and podman builders extend.
func (m *Manager) base() ([]string, error) {
	r, err := m.runtime()
	if err != nil {
		return nil, err
	}
	return r.Argv(), nil
}

// runCmd runs the command build makes from the guarded runtime argv, for the containercli
// builders that take it.
func (m *Manager) runCmd(ctx context.Context, build func(base []string) runner.Cmd) error {
	base, err := m.base()
	if err != nil {
		return err
	}
	return engine.Run(ctx, m.R, build(base))
}

// composeCmd is the guarded compose invocation (docker only): whatever
// docker.compose names -- a host with only the standalone v1 binary sets it to
// `docker-compose` -- defaulting to the runtime's own `compose` subcommand. Both the
// derivation and the check live in config.ComposeCommand, so what runs here is
// exactly what Validate approved.
func (m *Manager) composeCmd() (config.Command, error) { return m.Cfg.ComposeCommand() }

// projectDir is the directory compose resolves the file's relative paths against, and
// reads a `.env` from: the directory of the env file this deployment was loaded from.
// A Config built in code has none, and gets the working directory.
func (m *Manager) projectDir() string {
	if d := m.Cfg.BaseDir(); d != "" {
		return d
	}
	return "."
}

// composeSecretEnv is the "VAR=value" list backing the compose file's
// environment-sourced secrets. It is the only place a secret value enters a child
// process, and it never reaches an argv.
//
// preview means "do not read the certificate from disk". It is true under the Echo
// runner, and ALSO true for `down`, which needs each declared secret to be DEFINED
// but never reads it, so a moved certificate must not block a teardown. Only `up`
// consumes the bytes.
func (m *Manager) composeSecretEnv(preview bool) ([]string, error) {
	secrets, err := ResolveSecretValues(m.Cfg, m.P, preview)
	if err != nil {
		return nil, err
	}
	env := make([]string, 0, len(secrets)+1)
	for _, s := range secrets {
		env = append(env, s.EnvVar()+"="+s.Value)
	}
	return env, nil
}

// --- Check ------------------------------------------------------------------

// Check prints the resolved container configuration, probes the runtime, and
// resolves the broker hostname(s). It never mutates host or container state.
func (m *Manager) Check(ctx context.Context) error {
	m.CheckEnv()
	// Reachability is the one hard stop: with no engine every probe after it would
	// fail for that same reason, and three identical failures are less useful than
	// one. Everything after it reports in one pass, because `validate` promises to
	// name every problem it finds rather than stopping at the first.
	if err := m.Reachable(ctx); err != nil {
		return err
	}
	// The same `<runtime> info` probe every mutating op runs first, so validate
	// reports what deploy would refuse on rather than passing and leaving deploy to
	// fail at its own first step. `version` is not a substitute: it answers from a
	// binary that has not talked to anything, while `info` is what exercises the
	// rootless user session and its subuid mapping (preflight.go).
	if err := m.Preflight(ctx); err != nil {
		return err
	}
	// fix=false: `validate` promises a run that disturbs nothing, so the linger row
	// reports rather than enabling (rootless.go).
	return errors.Join(m.checkLimits(ctx, false), m.checkDNS(ctx), m.checkPodmanHost(ctx, false))
}

// CheckEnv writes the effective container configuration, mirroring k8s.CheckEnv.
func (m *Manager) CheckEnv() {
	r := m.report()
	cfg := m.Cfg
	cb := cfg.ContainerBlock(m.P)
	nw := cfg.NetworkBlock(m.P)

	mode := "standalone (single broker)"
	if cfg.RedundancyEnabled() {
		mode = "HA redundancy group (primary + backup + monitor)"
	}

	network := "host"
	if nw.Mode != "host" {
		network = fmt.Sprintf("bridge ports=%d", len(nw.Ports))
	}

	tls := "(not configured)"
	if cfg.TLS.Cert != "" || cfg.TLS.CertKey != "" {
		tls = fmt.Sprintf("cert=%s key=%s cas=%d",
			// Both are host file PATHS, not secret values, so both are shown. CertKey used
		// to be masked as set/MISSING beside its sibling on the same line, which read as
		// though one of the two carried something the other did not.
		orNone(cfg.TLS.Cert), orNone(cfg.TLS.CertKey), len(cfg.TLS.CAs))
	}

	r.Section("Broker deployment (" + string(m.P) + ")")
	rows := []output.KV{
		{Key: "container", Value: fmt.Sprintf("name=%s runtime=%s", cb.Name, cfg.ContainerRuntime(m.P))},
		{Key: "redundancy", Value: mode},
		{Key: "image", Value: orNone(cfg.Image.Ref())},
	}
	if cfg.Image.User != "" || cfg.Image.Pass != "" {
		rows = append(rows, output.KV{Key: "registry login", Value: fmt.Sprintf(
			"user=%s password=%s (prep logs in with `%s login`)",
			orNone(cfg.Image.User), setOrMissing(cfg.Image.Pass), cfg.ContainerRuntime(m.P).Name())})
	}
	rows = append(rows,
		output.KV{Key: "data dir", Value: cb.DataDir},
		output.KV{Key: "run user", Value: cb.RunUser},
		output.KV{Key: "network", Value: network},
		output.KV{Key: "admin", Value: fmt.Sprintf("user=%s password=%s", config.AdminUser, setOrMissing(cfg.SEMP.AdminPass))},
		output.KV{Key: "tls", Value: tls},
	)
	if m.P == config.Podman {
		// baseDir is reported only when set, and labelled for what it still does: the
		// server certificate now rides podman's secret store, and the directory is
		// consulted only to delete the bundle file an earlier build wrote there
		// (removeLegacyCertBundle).
		podmanRow := fmt.Sprintf("rootful=%t quadletDir=%s", cfg.Podman.Rootful, cfg.Podman.QuadletDir)
		if cfg.Podman.BaseDir != "" {
			podmanRow += fmt.Sprintf(" baseDir=%s (legacy bundle clean-up only)", cfg.Podman.BaseDir)
		}
		rows = append(rows, output.KV{Key: "podman", Value: podmanRow})
	} else {
		// The configured value, like the runtime line above -- CheckEnv reports what
		// the env file says, and Reachable (next in Check) is what fails loud if the
		// execution guard rejects it.
		rows = append(rows, output.KV{Key: "docker", Value: fmt.Sprintf(
			"compose=%s", cfg.Docker.Compose)})
	}
	rows = append(rows, output.KV{Key: "secrets", Value: secretSummary(m.P, render.ContainerSecrets(cfg, m.P))})
	if cfg.RedundancyEnabled() {
		n := cfg.Redundancy
		rows = append(rows,
			output.KV{Key: "primary", Value: fmt.Sprintf("%s (%s)", n.Primary.Name, orNone(n.Primary.Addr))},
			output.KV{Key: "backup", Value: fmt.Sprintf("%s (%s)", n.Backup.Name, orNone(n.Backup.Addr))},
			output.KV{Key: "monitor", Value: fmt.Sprintf("%s (%s)", n.Monitor.Name, orNone(n.Monitor.Addr))},
			output.KV{Key: "psk", Value: setOrMissing(n.PSK)},
		)
	} else {
		rows = append(rows, output.KV{Key: "node", Value: orNone(cfg.Redundancy.Primary.Name)})
	}
	// KVRow at the report's shared width rather than KVBlock, for the same reason
	// k8s.CheckEnv does it: the engine/nofile/dns lines are printed by their own
	// methods after this block and have to land in the same column.
	for _, row := range rows {
		r.KVRow(reportKeyWidth, row.Key, row.Value)
	}
}

// Reachable probes `<runtime> version` so a missing/stopped engine fails with an
// actionable error before any deploy step runs. On docker it also probes the
// configured compose command, since every deploy goes through it and the plugin
// is a separate install from the engine. Under the Echo runner it only echoes.
func (m *Manager) Reachable(ctx context.Context) error {
	// Resolve through the guard first so a refused command reports why it is
	// refused, rather than being reported as an unreachable runtime.
	rt, err := m.runtime()
	if err != nil {
		return err
	}
	if _, err := engine.Output(ctx, m.R, runner.Cmd{Argv: rt.Argv("version")}); err != nil {
		return fmt.Errorf("cannot reach the %s runtime %q (is it installed and running?): %w", platformTitle(m.P), rt, err)
	}
	if m.P == config.Docker {
		c, err := m.composeCmd()
		if err != nil {
			return err
		}
		if _, err := engine.Output(ctx, m.R, runner.Cmd{Argv: c.Argv("version")}); err != nil {
			return fmt.Errorf("cannot run the compose command %q (install the docker compose plugin, "+
				"or set docker.compose to this host's standalone 'docker-compose' binary): %w", c, err)
		}
	}
	return nil
}

// checkNoFile has moved to limits.go, and it is no longer rootless-only: the ask
// is a constant both artifacts carry (config.ContainerNoFile), so what has to be
// checked is what the HOST will allow -- fs.nr_open on every platform, plus
// user@<uid>.service's hard limits when rootless. /etc/security/limits.d is not
// involved in either: it is read by pam_limits, for login sessions.
//
// splitLimit went with it. It parsed the old ulimits.nofile key's "soft:hard";
// the pair is two int constants now, so there is nothing left to split.

// checkDNS resolves the broker hostname(s): in HA every node name must resolve
// (fail loud on any miss, matching 002-host-prep.sh); standalone warns only on
// the single name (it is used just as the routername). Skipped under the Echo runner,
// where no real lookups make sense.
func (m *Manager) checkDNS(ctx context.Context) error {
	_ = ctx
	if m.isEcho() {
		m.report().KVRow(reportKeyWidth, "dns", "skipped (preview)")
		return nil
	}
	r := m.report()
	if !m.Cfg.RedundancyEnabled() {
		name := m.Cfg.Redundancy.Primary.Name
		if name == "" || m.Resolve(name) {
			r.OK("broker hostname resolves: %s", orNone(name))
		} else {
			m.progress().Info("broker hostname does not resolve: %s (standalone -- used only as the routername, usually fine)", name)
		}
		return nil
	}
	failed := 0
	for _, n := range []struct{ role, name string }{
		{"primary", m.Cfg.Redundancy.Primary.Name},
		{"backup", m.Cfg.Redundancy.Backup.Name},
		{"monitor", m.Cfg.Redundancy.Monitor.Name},
	} {
		if m.Resolve(n.name) {
			r.OK("%s hostname resolves: %s", n.role, n.name)
			continue
		}
		r.Fail("%s hostname does NOT resolve: %s", n.role, n.name)
		failed++
	}
	if failed > 0 {
		return fmt.Errorf("%d redundancy-group hostname(s) do not resolve; add them to DNS or /etc/hosts on this host", failed)
	}
	return nil
}

// --- PrepHost ---------------------------------------------------------------

// PrepHost prepares this host for the broker container: it creates and chowns
// the data directory, resolves the broker hostname(s), and (HA only) ensures a
// redundancy PSK exists -- generating one and writing it back into the env file
// when absent. It ports bash/docker-podman/002-host-prep.sh.
func (m *Manager) PrepHost(ctx context.Context) error {
	// Before mkdir/chown and before a generated PSK is written back into the env
	// file: prep changes host state, so it takes the same probe deploy does.
	if err := m.Preflight(ctx); err != nil {
		return err
	}
	// The host ceilings and the podman host-readiness block, in one pass, BEFORE the
	// first mkdir: a rootful mkdir/chown here would leave a data directory the later
	// rootless deploy cannot use, a host missing linger or a user bus would get a
	// quadlet unit nothing can load, and a host below the nofile ceiling would get a
	// container the kernel refuses. Joined so an operator short on two things learns
	// both now.
	// fix=true: prep may repair what it finds -- the linger row, and fs.nr_open when
	// it is running as root. Both need no privilege it does not already hold.
	if err := errors.Join(m.checkLimits(ctx, true), m.checkPodmanHost(ctx, true)); err != nil {
		return err
	}

	cb := m.Cfg.ContainerBlock(m.P)
	m.logf("preparing data directory %s", cb.DataDir)
	if err := engine.Run(ctx, m.R, runner.Cmd{Argv: []string{"mkdir", "-p", cb.DataDir}}); err != nil {
		return fmt.Errorf("create data dir %q: %w%s", cb.DataDir, err, m.dataDirHint(cb.DataDir))
	}
	if m.rootlessPodman() {
		// 775 so the directory stays writable from outside the container once the
		// chown below moves it to a subuid nobody can become: container gid 0 maps
		// to this user's own group, so the group bit is what prep and the data-dir
		// row rely on afterwards (checkDataDir, rootless.go).
		//
		// IN THE NAMESPACE, like the chown. On a re-deploy the directory is already
		// owned by that subuid, so a host-side chmod is refused outright -- and
		// inside the namespace this user is root either way, which covers the fresh
		// directory and the already-prepared one with one call.
		if err := m.run(ctx, "unshare", "chmod", "775", cb.DataDir); err != nil {
			return fmt.Errorf("make data dir %q group-writable (rootless): %w%s",
				cb.DataDir, err, m.dataDirHint(cb.DataDir))
		}
		// Rootless podman: the container's uid:gid maps through subuid/subgid, so
		// enter the user namespace to apply the ownership the container will see.
		if err := m.run(ctx, "unshare", "chown", cb.RunUser, cb.DataDir); err != nil {
			return fmt.Errorf("chown data dir %q (rootless): %w%s", cb.DataDir, err, m.dataDirHint(cb.DataDir))
		}
	} else if err := engine.Run(ctx, m.R, runner.Cmd{Argv: []string{"chown", cb.RunUser, cb.DataDir}}); err != nil {
		return fmt.Errorf("chown data dir %q to %q: %w", cb.DataDir, cb.RunUser, err)
	}

	if err := m.checkDNS(ctx); err != nil {
		return err
	}
	return m.registryLogin(ctx)
}

// registryLogin authenticates this host to the image registry when credentials are
// configured. Containers have no analog of the k8s image-pull Secret, so without
// this the credentials in the env file did nothing here and a private-registry pull
// simply failed unauthenticated. The password is fed on stdin, so it never reaches
// an argv or an echoed command.
func (m *Manager) registryLogin(ctx context.Context) error {
	user, pass := m.Cfg.Image.User, m.Cfg.Image.Pass
	if user == "" && pass == "" {
		return nil
	}
	if user == "" || pass == "" {
		return fmt.Errorf("image.user and image.pass must both be set to log in to the registry (one of them is empty)")
	}
	registry := m.Cfg.Image.Registry
	args := []string{"login", "--username", user, "--password-stdin"}
	if registry != "" {
		args = append(args, registry)
	}
	m.logf("logging in to registry %s as %s", orNone(registry), user)
	r, err := m.runtime()
	if err != nil {
		return err
	}
	if err := engine.Run(ctx, m.R, runner.Cmd{Argv: r.Argv(args...), Stdin: []byte(pass)}); err != nil {
		return fmt.Errorf("%s login to registry %s failed: %w", platformTitle(m.P), orNone(registry), err)
	}
	return nil
}

// --- Deploy -----------------------------------------------------------------

// Deploy renders and starts the broker container for role's identity. Podman
// uses a systemd quadlet unit; docker uses a compose file, fed to compose on stdin.
// The broker's secret settings are externalized first, since the rendered artifact
// only references them by name.
//
// The order is compare, ask, write. Nothing is written to the host, and nothing is
// restarted, until the operator has answered for a running service whose artifact or
// secrets changed; a "no" leaves the host exactly as it was.
func (m *Manager) Deploy(ctx context.Context, role config.Role) error {
	// Ahead of every write and every mutating call: deployPodman loads values into
	// podman's store and drops a unit on disk, so an engine that cannot be reached
	// must stop the deploy before either.
	if err := m.Preflight(ctx); err != nil {
		return err
	}
	id := m.Cfg.ResolveNode(role)
	if m.P == config.Podman {
		if err := m.checkPodmanEUID(); err != nil {
			return err
		}
	}
	secrets, err := m.prepareSecrets()
	if err != nil {
		return err
	}
	if m.P == config.Podman {
		return m.deployPodman(ctx, id, secrets)
	}
	return m.deployDocker(ctx, id)
}

// prepareSecrets resolves the broker's secret settings so the deploy artifact can
// reference them instead of carrying them: podman loads each into its own secret
// store, while docker needs nothing prepared -- its compose secrets read this
// deployment's values from the environment `compose` is given (see deployDocker), so no
// secret is written beside anything. That is not the same as never reaching this
// host's disk: compose tars an environment-sourced secret into the container's own
// filesystem (CopyToContainer, mode 0444), where it rests in the writable layer for the
// container's lifetime -- docs/operations.md says so under Re-deploying. An empty value
// fails loud here rather than deploying a broker with no password -- except under the
// Echo runner, which must stay previewable before `broker deploy` has generated the HA
// pre-shared key.
//
// It only reads: the podman store is written later, after the operator has answered.
func (m *Manager) prepareSecrets() ([]render.ContainerSecret, error) {
	// Skipped under the Echo runner: a preview must not need secret values on disk, and
	// `broker generate` renders name-level references only, so there is nothing to check.
	if !m.isEcho() {
		// Preflight FIRST, then resolve. The order matters for the message: an env
		// file that is both pre-`broker deploy` (empty PSK) and missing its cert files
		// should report the PSK and its `broker deploy` hint, not the certificate.
		if err := render.SecretPreflight(m.Cfg, m.P); err != nil {
			return nil, err
		}
	}
	// Resolve on BOTH platforms. Docker re-reads the values for its child environment,
	// but reading the certificate HERE is what makes an unreadable one fail before
	// anything is asked or written, so "nothing happened" stays true.
	return ResolveSecretValues(m.Cfg, m.P, m.isEcho())
}

// storedMatches reports whether podman's store already holds exactly s's value, by
// reading it back (podman.SecretData, which needs podman 4.7: replace says the host has
// it, as asked once per deploy) and comparing in memory (podman.SecretMatches). The value
// arrives on stdout alone and is never printed, logged or put in an error: the runner's
// trace shows argv only. A probe that fails, a secret that is absent, a value that
// differs, a podman below 4.7 and the Echo runner (which probes nothing) all read as
// "changed", which costs at most a consent-gated restart.
func (m *Manager) storedMatches(ctx context.Context, base []string, s render.ContainerSecret, replace bool) bool {
	if m.isEcho() || !replace {
		return false
	}
	c, err := podman.SecretData(base, s.Name)
	if err != nil {
		return false
	}
	out, err := engine.Output(ctx, m.R, c)
	if err != nil {
		return false
	}
	return podman.SecretMatches(string(out), []byte(s.Value))
}

// podmanReplace asks once per deploy whether this host's podman has `secret create
// --replace` and `secret inspect --showsecret` (both 4.7), or false when there are no
// secrets to ask about.
func (m *Manager) podmanReplace(ctx context.Context, base []string, secrets []render.ContainerSecret) bool {
	return len(secrets) > 0 && m.secretReplace(ctx, base)
}

// changedSecrets is the secrets whose stored value differs from the one configured
// now. A secret that matches is reported as left in place. A first deploy, a secret an
// earlier build stored, an unreadable one or a podman below 4.7 counts as changed.
// Nothing is written here.
func (m *Manager) changedSecrets(ctx context.Context, base []string, secrets []render.ContainerSecret, replace bool) []render.ContainerSecret {
	var changed []render.ContainerSecret
	for _, s := range secrets {
		if m.storedMatches(ctx, base, s, replace) {
			m.progress().Info("podman secret %s already holds this value, left in place.", s.Name)
			continue
		}
		changed = append(changed, s)
	}
	return changed
}

// writePodmanSecrets loads each given secret into podman's secret store, feeding the
// value on stdin so it never reaches an argv or the echoed command. No label is written:
// a label rides `secret create`'s argv, which any local user can read, so the next deploy
// compares by reading the stored value back instead (storedMatches).
//
// `create --replace` where replace holds (podman 4.7 and later), and remove-then-create
// otherwise: the pair is idempotent the same way -- `rm --ignore` makes a missing secret
// a success, so a redeploy with a rotated value works, and a first deploy with nothing in
// the store works. The Echo runner takes the older pair, so a preview shows the argv that
// works on every supported podman.
//
// A failure of the rm half is NOT warned away. --ignore already absorbs the only
// benign case, so anything left is real -- a store this user cannot write, say --
// and continuing would create a secret next to one that could not be removed.
func (m *Manager) writePodmanSecrets(ctx context.Context, base []string, secrets []render.ContainerSecret, replace bool) error {
	for _, s := range secrets {
		if !replace {
			if rm, ok := podman.RemoveSecrets(base, s.Name); ok {
				if err := engine.Run(ctx, m.R, rm); err != nil {
					return fmt.Errorf("remove the existing podman secret %q before recreating it (from %s): %w",
						s.Name, s.ConfigKey, err)
				}
			}
		}
		if err := engine.Run(ctx, m.R, podman.CreateSecret(base, s.Name, []byte(s.Value), replace)); err != nil {
			return fmt.Errorf("create podman secret %q from %s: %w", s.Name, s.ConfigKey, err)
		}
	}
	return nil
}

// secretReplace reports whether this host's podman has `secret create --replace`, from
// its client version (stdout alone: podman can warn on stderr first). Under the Echo
// runner nothing is probed and the answer is no. A probe that fails is no as well, which
// is the safe side: remove-then-create works on every podman this tool supports.
func (m *Manager) secretReplace(ctx context.Context, base []string) bool {
	if m.isEcho() {
		return false
	}
	out, err := engine.Output(ctx, m.R, podman.Version(base))
	if err != nil {
		return false
	}
	return podman.SupportsReplace(string(out))
}

// UpdateServerCertSecret is the persistence half of `broker configure server-certs` on
// a container host; the CLI load that follows it is what the running broker serves. It
// never restarts anything.
//
// Podman: the store secret is overwritten under the deploy rule -- a stored value
// that matches leaves it, a different or unreadable one replaces it -- so the next start of the
// unit mounts the new certificate. Docker keeps no secret object: the container holds
// the certificate it was created with, so there is nothing to update, and the next
// `broker deploy` offers to recreate it because the deploy digest no longer matches. With
// no certificate configured it does nothing, and the CLI load reports why.
func (m *Manager) UpdateServerCertSecret(ctx context.Context) error {
	if m.Cfg.TLS.Cert == "" || m.Cfg.TLS.CertKey == "" {
		return nil
	}
	if m.P != config.Podman {
		m.progress().Info("docker keeps no secret store: container %s keeps the certificate it was created with "+
			"until `broker deploy` recreates it, which that deploy offers once the certificate differs.", m.name())
		return nil
	}
	if err := m.Preflight(ctx); err != nil {
		return err
	}
	secrets, err := ResolveSecretValues(m.Cfg, m.P, m.isEcho())
	if err != nil {
		return err
	}
	var cert []render.ContainerSecret
	for _, s := range secrets {
		if len(s.SourceFiles) > 0 {
			cert = append(cert, s)
		}
	}
	base, err := m.base()
	if err != nil {
		return err
	}
	replace := m.podmanReplace(ctx, base, cert)
	return m.writePodmanSecrets(ctx, base, m.changedSecrets(ctx, base, cert, replace), replace)
}

// removeLegacyCertBundle deletes the server-certificate bundle an earlier build wrote
// to <podman.baseDir>/<name>-tls-servercertificate.pem (render.ServerCertBundlePath),
// so upgrading does not leave a private key on the host. It runs on every podman
// deploy, after the new unit is written (so nothing references the file any more),
// and on every podman remove. The path does not depend on TLS being configured: a
// host may have dropped tls.cert since the file was written.
//
// A missing file, or an unset baseDir, is the normal case and says nothing. A failure
// to remove an existing one is FATAL, for the same reason a leftover certificate
// secret is: it holds the broker's private key. Only the FILE goes; the directory may
// hold other brokers' files.
func (m *Manager) removeLegacyCertBundle() error {
	if m.P != config.Podman || m.Cfg.Podman.BaseDir == "" {
		return nil
	}
	path := filepath.FromSlash(render.ServerCertBundlePath(m.Cfg))
	if m.isEcho() {
		if fileExists(path) {
			m.progress().Info("would remove legacy server certificate bundle %s (skipped in preview).", path)
		}
		return nil
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("remove the legacy server certificate bundle %q: an earlier build wrote the broker's "+
			"PRIVATE KEY there, and the certificate now rides podman's secret store, so the file must not be "+
			"left on this host -- remove it by hand and re-run: %w", path, err)
	}
	m.progress().OK("removed legacy server certificate bundle %s (the certificate now rides podman's secret store)", path)
	return nil
}

// unitChanged reports whether the quadlet unit at path is missing or differs from body,
// without writing anything: the deploy compares first so it can ask before it writes.
// Under the Echo runner nothing is read and the unit counts as changed, so the preview
// shows the work a real run would do.
func (m *Manager) unitChanged(path string, body []byte) bool {
	if m.isEcho() {
		return true
	}
	existing, err := os.ReadFile(path)
	return err != nil || !bytes.Equal(existing, body)
}

// writeArtifact writes a rendered artifact (the quadlet unit, the sysctl drop-in),
// reporting whether it differs from what was already on disk. An unchanged artifact is
// not rewritten. Under the Echo runner nothing is written and the artifact counts as
// changed, so the preview shows the work a real run would do. deployPodman has already
// compared the unit (unitChanged) so that it can ask before this writes it.
func (m *Manager) writeArtifact(path string, body []byte, what string, dirMode os.FileMode) (bool, error) {
	if m.isEcho() {
		m.progress().Info("would write %s %s (skipped in preview).", what, path)
		return true, nil
	}
	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, body) {
		m.progress().Info("%s %s is already up to date.", what, path)
		return false, nil
	}
	if dirMode != 0 {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, dirMode); err != nil {
			return false, fmt.Errorf("create %s dir %q: %w", what, dir, err)
		}
	}
	if err := os.WriteFile(path, body, 0o600); err != nil {
		return false, fmt.Errorf("write %s %q: %w", what, path, err)
	}
	m.progress().OK("wrote %s %s", what, path)
	return true, nil
}

// approve decides whether a running broker may be restarted or recreated to pick up a
// change. --restart pre-approves it; otherwise the Confirm seam asks, and an error from
// it (no terminal and no --no-prompt) is the deploy's. A nil seam declines, so a Manager
// built without the CLI never bounces a running broker. Deliberately a decision of its
// own: bouncing a live broker is not something --restart's absence leaves to chance, and
// --no-prompt is what answers it for a script.
func (m *Manager) approve(question string) (bool, error) {
	if m.RestartApproved {
		return true, nil
	}
	if m.Confirm == nil {
		return false, nil
	}
	return m.Confirm(question)
}

// cancelled says a declined deploy changed nothing, and that the next one asks again.
func (m *Manager) cancelled() {
	m.progress().Info("cancelled; nothing was changed. The next deploy asks again.")
}

// deployPodman compares, asks, then writes: the unit rendered now against the file on
// disk, and each secret's value against the one read back from the store (podman 4.7+; older counts as changed). Only a change
// restarts the service, and only after the operator has agreed to restart a running one;
// an unchanged deploy writes nothing and restarts nothing. secrets are the values
// prepareSecrets resolved, so a certificate that could not be read has already aborted
// the deploy before anything was touched.
//
// `daemon-reload` runs only when something was written. `systemctl start` on an active
// unit is a no-op and a reload never restarts, so a running service picks a change up only
// through the restart below.
//
// ASSUMED, NOT VERIFIED: quadlet replaces the container on each service start, so
// starting a stopped unit re-reads the store and picks up a rotated secret. Docker's
// equivalent branch does NOT behave that way -- compose keeps a stopped container and
// starts it with the credentials it was created with, which is what the deploy digest
// label is for. To check on podman: stop the unit, rotate a value, `systemctl start`,
// then read /mnt/secrets/username_admin_password inside the container.
//
// serviceActive is read in its LOSSY form on purpose, unlike deployDocker, which aborts
// when its probe cannot answer (containerRunningKnown). An unanswered probe here reads as
// "not active" and takes the plain start path, which systemd makes a no-op on a running
// unit; the only cost is that the change is written but the running service keeps its old
// unit until it is restarted (`broker restart`).
func (m *Manager) deployPodman(ctx context.Context, id config.NodeIdentity, secrets []render.ContainerSecret) error {
	unit := filepath.ToSlash(filepath.Join(m.Cfg.Podman.QuadletDir, m.name()+".container"))
	svc := m.name() + ".service"
	body := render.Quadlet(m.Cfg, id)

	unitChanged := m.unitChanged(unit, body)
	base, err := m.base()
	if err != nil {
		return err
	}
	replace := m.podmanReplace(ctx, base, secrets)
	changedSecrets := m.changedSecrets(ctx, base, secrets, replace)
	changed := unitChanged || len(changedSecrets) > 0
	active := m.serviceActive(ctx)

	if changed && active {
		what := "its quadlet unit changed"
		switch {
		case unitChanged && len(changedSecrets) > 0:
			what = "its quadlet unit and secrets changed"
		case !unitChanged:
			what = "its secrets changed"
		}
		ok, err := m.approve(fmt.Sprintf("Service %s is running and %s. Restart it now to apply the change?", svc, what))
		if err != nil {
			return err
		}
		if !ok {
			m.cancelled()
			return nil
		}
	}

	if err := m.writePodmanSecrets(ctx, base, changedSecrets, replace); err != nil {
		return err
	}
	if unitChanged {
		if _, err := m.writeArtifact(unit, body, "quadlet unit", 0o755); err != nil {
			return err
		}
	}
	// Now that no unit on disk bind-mounts it. A running container started from the
	// old unit keeps its own mount of the file until it is restarted.
	if err := m.removeLegacyCertBundle(); err != nil {
		return err
	}
	user := m.systemctlUser()
	if changed {
		if err := m.systemctl(ctx, systemd.DaemonReload(user)); err != nil {
			return err
		}
	}
	if !active {
		return m.systemctl(ctx, systemd.Start(user, svc))
	}
	if changed {
		return m.systemctl(ctx, systemd.Restart(user, svc))
	}
	if m.RestartApproved {
		m.progress().Info("unit and secrets unchanged; restarting %s because --restart was given.", svc)
		return m.systemctl(ctx, systemd.Restart(user, svc))
	}
	m.progress().Info("%s is already active on this unit and these secrets -- unchanged, left running.", svc)
	m.progress().Info("`broker restart` (or deploy with --restart) restarts it anyway.")
	return nil
}

// deployDocker feeds the compose file to compose on stdin (nothing is written to disk)
// with the secret values in the child's environment, and lets compose decide what to
// recreate. Compose compares a service's configuration, but a secret's CONTENT is not
// part of it, so the file carries one label (render.DeployDigestLabel) whose value is
// interpolated from render.DeployDigestEnvVar, set here to the digest of the file and of
// every value the child receives (containercli.ComposeDigest, taken before the variable
// is added). A changed file, a rotated password or a renewed certificate therefore all
// change the configuration, and compose recreates the container; an unchanged deploy
// changes nothing, whether the container is running (left alone) or stopped (started).
//
// The digest is also what predicts the recreate, so the operator is asked first: a
// running container whose label differs from the digest -- or has none, or cannot be
// inspected -- is about to be recreated by `up -d`. --restart pre-approves it, and also
// forces the recreate of an unchanged running container. A "no" runs nothing.
//
// "Running" has to be ANSWERED, not assumed (M4): a `ps` that cannot run says nothing
// about whether the container exists, and reading that silence as "not running" would let
// a transient engine failure skip the question for a LIVE broker. containerRunningKnown is
// read directly so an unanswered probe stops here instead of guessing either way.
func (m *Manager) deployDocker(ctx context.Context, id config.NodeIdentity) error {
	c, err := m.composeCmd()
	if err != nil {
		return err
	}
	base, err := m.base()
	if err != nil {
		return err
	}
	file := render.Compose(m.Cfg, id)
	env, err := m.composeSecretEnv(m.isEcho())
	if err != nil {
		return err
	}
	digest := containercli.ComposeDigest(file, env)
	env = append(env, render.DeployDigestEnvVar+"="+digest)

	running, answered, probeErr := m.containerRunningKnown(ctx, base)
	if !answered {
		return fmt.Errorf("cannot tell whether container %s is running -- the `ps` probe itself "+
			"failed, so it is unsafe to guess whether the deploy would recreate it; check `%s ps` by "+
			"hand and re-run: %w", m.name(), m.Cfg.ContainerRuntime(m.P).Name(), probeErr)
	}
	changed := running && m.deployedDigest(ctx, base) != digest
	if changed {
		m.progress().Info("container %s was created from a different compose file or different secrets.", m.name())
		ok, err := m.approve(fmt.Sprintf("Container %s is running and its compose file or a secret changed. "+
			"Recreate it now to apply the change?", m.name()))
		if err != nil {
			return err
		}
		if !ok {
			m.cancelled()
			return nil
		}
	}
	recreate := running && !changed && m.RestartApproved
	if recreate {
		m.progress().Info("deploy unchanged; recreating container %s because --restart was given.", m.name())
	}
	if err := engine.Run(ctx, m.R, containercli.ComposeUp(c.Argv(), file, m.projectDir(), env, recreate)); err != nil {
		return err
	}
	if running && !changed && !recreate {
		m.progress().Info("container %s is already running this deploy -- unchanged, left running.", m.name())
		m.progress().Info("`broker restart` (or deploy with --restart) restarts it anyway.")
	}
	return nil
}

// deployedDigest is the render.DeployDigestLabel value of the running container, or
// empty when it carries none, cannot be inspected or the answer cannot be read -- all of
// which the caller reads as "changed", which costs at most a consent-gated recreate.
func (m *Manager) deployedDigest(ctx context.Context, base []string) string {
	c, _ := inspect.Containers(base, m.name())
	out, err := engine.Output(ctx, m.R, c)
	if err != nil {
		return ""
	}
	cs, err := inspect.Decode(string(out))
	if err != nil || len(cs) == 0 {
		return ""
	}
	return cs[0].Labels[render.DeployDigestLabel]
}

// serviceActive reports whether this host's broker unit is already active, so
// Deploy can restart it instead of issuing a no-op start. "reloading" counts
// (systemd.Active): the unit is running, so a changed one needs the restart. Under the
// Echo runner no host state is probed (matching checkPodmanEUID/checkDNS) and the
// answer is "not active", so the preview shows the plain start path.
func (m *Manager) serviceActive(ctx context.Context) bool {
	if m.isEcho() {
		return false
	}
	state, _ := m.serviceState(ctx)
	return systemd.Active(state)
}

// serviceState returns systemd's own word for this unit ("active", "inactive",
// "failed", "unknown", ...) and whether the probe answered at all.
//
// The exit code cannot answer either question: `systemctl is-active` exits
// non-zero for every state but "active", so a failed command is the NORMAL reply
// for a stopped unit and is indistinguishable -- by exit code alone -- from a
// systemd that could not be reached. Under rootless podman those are genuinely
// different situations: the user session is a separate IPC path from the engine
// socket, so `podman info` can succeed while `systemctl --user` cannot be
// reached at all. The state TEXT is what separates them; an empty reply is what
// says the probe itself failed.
func (m *Manager) serviceState(ctx context.Context) (state string, answered bool) {
	out, _ := m.systemctlOutput(ctx, systemd.IsActive(m.systemctlUser(), m.name()+".service"))
	return systemd.State(string(out))
}

// containerRunningKnown reports whether this host's broker container is up, plus the
// two things a bare bool cannot express: whether the probe answered at all, and its own
// error when it didn't. base is the guarded runtime argv.
//
// The name is matched in Go (containercli.HasName) rather than by `--filter name=`,
// which both engines treat as an unanchored REGEX: a bare name is a substring match, so
// a `solace` deployment would see a running `solace-edge` as its own, and a name
// carrying '.' (which the schema allows) would match any character there. Comparing the
// listed names exactly depends on nothing.
// A `ps` that could not run says nothing about the container, and a caller that reads
// that silence as "not running" would act on a guess -- deployDocker refuses to guess
// whether to ask before recreating a live broker.
func (m *Manager) containerRunningKnown(ctx context.Context, base []string) (running, answered bool, err error) {
	if m.isEcho() {
		return false, true, nil
	}
	out, err := engine.Output(ctx, m.R, containercli.List(base, true))
	if err != nil {
		return false, false, err
	}
	return containercli.HasName(containercli.ParseList(string(out)), m.name()), true, nil
}

// --- Delete -----------------------------------------------------------------

// Delete stops and removes the broker container (and its unit/compose artifact),
// then optionally removes the data directory when purge is set.
func (m *Manager) Delete(ctx context.Context, purge bool) error {
	if err := m.Preflight(ctx); err != nil {
		return err
	}
	var err error
	if m.P == config.Podman {
		err = m.deletePodman(ctx)
	} else {
		err = m.deleteDocker(ctx)
	}
	if err != nil {
		return err
	}
	if purge {
		if err := m.purgeData(ctx); err != nil {
			return err
		}
		m.progress().OK("data directory %s emptied; the directory itself is kept.",
			m.Cfg.ContainerBlock(m.P).DataDir)
		return nil
	}
	m.progress().Info("data directory %s and its contents kept (pass --delete-data to empty it).",
		m.Cfg.ContainerBlock(m.P).DataDir)
	return nil
}

func (m *Manager) deletePodman(ctx context.Context) error {
	svc := m.name() + ".service"
	user := m.systemctlUser()
	if err := m.systemctl(ctx, systemd.Stop(user, svc)); err != nil {
		// "podman info" (Preflight) proves the engine is reachable, not that this
		// unit actually stopped, so a failed stop needs systemd's own confirmation
		// before the unit is removed out from under a broker that is still serving
		// traffic -- and before Delete goes on to rm -rf the data directory.
		// Only a state that CONFIRMS the unit is stopped continues; "still active"
		// and "no answer at all" both stop here, because the case this exists for
		// (an unreachable rootless user session) produces the second one.
		state, answered := m.serviceState(ctx)
		// systemd.Stopped: inactive, failed and "unknown" (systemd saying it has no such
		// unit, which is equally safe to remove) confirm it. Every other answer, no
		// answer included, leaves the broker possibly serving, so a teardown refuses
		// rather than guessing.
		if !answered || !systemd.Stopped(state) {
			reported := "state " + state
			if !answered {
				reported = "nothing -- the state probe itself failed"
			}
			return fmt.Errorf("stop %s failed and it is NOT confirmed stopped (systemd reports %s); "+
				"the quadlet unit and the data directory are left in place. Stop it by hand "+
				"(`%s`) and re-run: %w",
				svc, reported, strings.Join(systemd.Systemctl(user, "status", svc).Argv, " "), err)
		}
		m.progress().Warn("stopping %s failed but systemd reports it %s; continuing: %v", svc, state, err)
	}
	unit := filepath.ToSlash(filepath.Join(m.Cfg.Podman.QuadletDir, m.name()+".container"))
	if m.isEcho() {
		m.progress().Info("would remove quadlet unit %s (skipped in preview).", unit)
	} else if err := os.Remove(unit); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove quadlet unit %q: %w", unit, err)
	} else {
		m.progress().OK("removed quadlet unit %s", unit)
	}
	if err := m.systemctl(ctx, systemd.DaemonReload(user)); err != nil {
		return err
	}
	// With the other things the container consumed, now that the unit is gone. Both
	// run even when one fails, and both failures are reported.
	return errors.Join(m.removeLegacyCertBundle(), m.removePodmanSecrets(ctx))
}

// removePodmanSecrets removes every secret this deployment loaded into podman's
// secret store, using the same render.ContainerSecrets(m.Cfg, m.P) that
// createPodmanSecrets reads them from -- so the create and remove lists cannot
// drift apart. It runs only after the container/unit is gone (deletePodman's
// last step): a leftover credential must not fail a teardown that otherwise
// succeeded, so its rm failure is a warning, not an error, and every secret is
// still attempted.
//
// The server certificate is the exception: it holds the broker's PRIVATE KEY, so
// failing to remove it fails the teardown -- after every other secret has been
// attempted -- rather than leaving the key in the store behind a warning.
func (m *Manager) removePodmanSecrets(ctx context.Context) error {
	secrets := render.ContainerSecrets(m.Cfg, m.P)
	if len(secrets) == 0 {
		return nil
	}
	r, err := m.runtime()
	if err != nil {
		return err
	}
	base := r.Argv()
	var keyErr error
	for _, s := range secrets {
		if m.isEcho() {
			m.progress().Info("would remove podman secret %s (skipped in preview).", s.Name)
			continue
		}
		// podman.RemoveSecrets is --ignore, the same argv createPodmanSecrets' rm half
		// runs. It also makes the warning below honest: "already removed" stops
		// being a plausible cause once a missing secret is a success, so what is
		// left really is a failure worth naming. One secret per call, so one failure
		// is attributable to its name.
		rm, _ := podman.RemoveSecrets(base, s.Name)
		if err := engine.Run(ctx, m.R, rm); err != nil {
			if len(s.SourceFiles) > 0 {
				keyErr = fmt.Errorf("remove podman secret %q: it holds the broker's PRIVATE KEY and must not be "+
					"left in this host's secret store -- remove it by hand (`%s`) and re-run: %w",
					s.Name, strings.Join(append([]string{r.Name()}, r.Args("secret", "rm", s.Name)...), " "), err)
				continue
			}
			m.progress().Warn("removing podman secret %s failed: %v", s.Name, err)
			continue
		}
		m.progress().OK("removed podman secret %s", s.Name)
	}
	return keyErr
}

// deleteDocker takes the compose project down, with the file rendered now fed on stdin
// like the deploy's, so no compose file has to be on this host. compose finds the project
// by the file's own name: key, so a deployment an earlier build created from a file on
// disk comes down the same way, and a project that is not there is not an error. Down
// interpolates the whole model, so every secret is DEFINED (as a placeholder: nothing
// reads a value, which is why a moved certificate cannot block a teardown) and the
// deploy-digest variable falls back to its empty default. Primary is the identity
// because a down needs the service and project names, which no role changes.
func (m *Manager) deleteDocker(ctx context.Context) error {
	c, err := m.composeCmd()
	if err != nil {
		return err
	}
	env, err := m.composeSecretEnv(true)
	if err != nil {
		return err
	}
	file := render.Compose(m.Cfg, m.Cfg.ResolveNode(config.Primary))
	return engine.Run(ctx, m.R, containercli.ComposeDown(c.Argv(), file, m.projectDir(), env))
}

// purgeData empties container.dataDir and LEAVES THE DIRECTORY. Removing it would
// throw away the ownership and mode prep established -- on a rootless host that is
// a subuid mapping only `podman unshare` can restore -- so the next deploy would
// have to re-create it, and could not where the parent is not writable by this
// user. What --delete-data promises is that the broker's data is gone, not that
// the mount point is.
//
// find rather than a shell glob: `rm -rf dir/*` misses dotfiles, and the
// alternatives that do not need two more patterns need a shell. -delete implies
// -depth, so a non-empty subdirectory goes with its contents.
func (m *Manager) purgeData(ctx context.Context) error {
	dir := m.Cfg.ContainerBlock(m.P).DataDir
	m.logf("clearing data directory %s", dir)
	args := []string{dir, "-mindepth", "1", "-delete"}
	if m.rootlessPodman() {
		return m.run(ctx, append([]string{"unshare", "find"}, args...)...)
	}
	return engine.Run(ctx, m.R, runner.Cmd{Argv: append([]string{"find"}, args...)})
}

// --- Lifecycle: Start / Stop / Restart ---------------------------------------
//
// These act on a broker that is already DEPLOYED: the compose file or quadlet unit
// stays on disk and the data directory is untouched, so a stopped broker is one
// `start broker` away from running again. That is the same distinction Kubernetes
// draws by scaling a StatefulSet to 0 rather than deleting it, which is why the CLI
// presents one pair of verbs over both platforms.
//
// Podman drives systemd rather than the engine directly: the quadlet unit owns the
// container's lifecycle, so stopping the container behind systemd's back would just
// invite it to be restarted.

// Start brings the deployed broker back up.
func (m *Manager) Start(ctx context.Context) error {
	if err := m.Preflight(ctx); err != nil {
		return err
	}
	if m.P == config.Podman {
		return m.systemctl(ctx, systemd.Start(m.systemctlUser(), m.name()+".service"))
	}
	return m.runCmd(ctx, func(base []string) runner.Cmd { return containercli.Start(base, m.name()) })
}

// Stop shuts the broker down without removing it. The container and its data
// survive -- `remove broker` is what deletes them.
func (m *Manager) Stop(ctx context.Context) error {
	if err := m.Preflight(ctx); err != nil {
		return err
	}
	if m.P == config.Podman {
		return m.systemctl(ctx, systemd.Stop(m.systemctlUser(), m.name()+".service"))
	}
	return m.runCmd(ctx, func(base []string) runner.Cmd { return containercli.Stop(base, m.name()) })
}

// Restart bounces the broker in place. It applies nothing new: a changed deploy
// artifact needs `deploy broker`, which asks before it restarts.
func (m *Manager) Restart(ctx context.Context) error {
	if err := m.Preflight(ctx); err != nil {
		return err
	}
	if m.P == config.Podman {
		return m.systemctl(ctx, systemd.Restart(m.systemctlUser(), m.name()+".service"))
	}
	return m.runCmd(ctx, func(base []string) runner.Cmd { return containercli.Restart(base, m.name()) })
}

// --- Status / Logs / CLI / Shell --------------------------------------------

// solaceImageMarker identifies a Solace broker container by the image it was
// created from. One substring covers every spelling in circulation --
// solace-pubsub-standard, solace-pubsub-enterprise, and the pubsubplus- variants --
// and it is matched against the image column rather than the container name because
// the name is the operator's to choose (container.name) while the image is not.
const solaceImageMarker = "solace-pubsub"

// listContainers lists every container on this host through containercli.List and
// reads the answer. Matching is done in Go on the rows (HasName, ByImage), because both
// engines treat `--filter name=` as an unanchored regular expression.
func (m *Manager) listContainers(ctx context.Context) ([]containercli.Row, error) {
	base, err := m.base()
	if err != nil {
		return nil, err
	}
	out, err := engine.Output(ctx, m.R, containercli.List(base, false))
	if err != nil {
		return nil, err
	}
	return containercli.ParseList(string(out)), nil
}

// printableRow is a listing row as it reaches a table: every field is engine-supplied
// text about a container this tool did not necessarily create, so each passes through
// inspect.Printable first.
func printableRow(r containercli.Row) []string {
	return []string{inspect.Printable(r.Name), inspect.Printable(r.Image), inspect.Printable(r.Status)}
}

// StatusAll lists every Solace broker container on this host, not just the one this
// env file names -- the container answer to `--all`, which on Kubernetes surveys the
// cluster. Discovery is by image, so a broker someone deployed by hand, or under a
// name this config knows nothing about, still shows up. That is the point: the flag
// exists to answer "what is actually running here", which a config-scoped listing
// cannot.
//
// detail adds each container's mounts and any secrets it carries -- the static
// makeup behind the running row.
func (m *Manager) StatusAll(ctx context.Context, detail bool) error {
	r := m.report()
	rows, err := m.listContainers(ctx)
	if err != nil {
		return fmt.Errorf("listing containers: %w", err)
	}
	names := solaceRows(r, containercli.ByImage(rows, solaceImageMarker))
	if !detail {
		return nil
	}
	for _, name := range names {
		r.Line("")
		r.Section(name)
		// inspect.go decodes the JSON rather than handing the engine a --format
		// template, because the field names differ between docker and podman in
		// ways a template gets silently wrong (see the header comment there).
		if err := m.inspectAndReport(ctx, name); err != nil {
			r.Line("  (could not inspect %s: %v)", name, err)
		}
	}
	return nil
}

// solaceRows prints the header plus every row it is given (the ones whose image names
// a Solace broker), and returns the container names it kept so a caller can go deeper
// on each.
//
// The names did not come from the env file: they came out of `<runtime> ps` on this
// host, which makes them untrusted input on the way back in. A row whose name fails
// containercli.ValidName is still SHOWN -- it is a real container on this host and
// hiding it would be worse -- but its name is not returned, so nothing further is run
// against it. Skipped out loud rather than silently dropped: a missing row reads as
// "no such container". Validating at the boundary is cheaper than trusting every
// consumer downstream.
func solaceRows(s *output.Sink, listed []containercli.Row) []string {
	var names []string
	var rows [][]string
	for _, row := range listed {
		rows = append(rows, printableRow(row))
		if !containercli.ValidName(row.Name) {
			// s.Line, not s.Warn: this is a report body, and Warn is the stderr voice
			// (internal/output). The note rides with the row it is about.
			s.Line("  (not inspected: %q is not a name this tool will pass back to the engine)", row.Name)
			continue
		}
		names = append(names, row.Name)
	}
	// Table computes each column from what is actually in it, so a long image ref
	// widens the column instead of running into the next one.
	s.Table([]string{"NAME", "IMAGE", "STATUS"}, rows)
	// len(rows), not len(names): names holds only the containers whose names passed the
	// check above, so a host running one Solace container under a name this tool will
	// not hand back to the engine printed "none found" directly beneath a table listing it.
	if len(rows) == 0 {
		s.Line("  (no Solace broker containers on this host)")
	}
	return names
}

// Status reports the container's state. Podman also shows its systemd unit.
//
// The listing is what the engine says about the container; the block after it
// is what this tool reads out of `inspect` (inspect.go) -- health under whichever
// spelling this engine uses, and a restart count that on podman comes from systemd
// rather than from the engine's own counter, which a quadlet restart leaves at zero.
// Those are the two facts an operator asking "is my broker healthy" needs and
// neither engine's `ps` line carries.
//
// The listing is `ps --all --no-trunc` with an explicit template, so the columns are
// the same on docker and podman and the PORTS column is never in it: a broker publishes
// a dozen or more ports, and the engine's default table spends most of a terminal line
// on them. Nothing is lost: the ports are in the env file that chose them, in the
// artifact `broker generate` prints, and in the inspect block this command appends. The
// rows are filtered to this container's exact name in Go and printed as this tool's own
// table.
//
// A failed inspect is a WARNING, not a failure: the listing above it already
// answered whether the container exists, so the command has reported something
// useful and must not exit non-zero because the extra detail was unavailable.
func (m *Manager) Status(ctx context.Context) error {
	if m.P == config.Podman {
		// The unit is podman's half of "is it up": a quadlet container that systemd
		// never started has no engine-side row to find.
		svc := m.name() + ".service"
		if err := m.systemctl(ctx, systemd.Systemctl(m.systemctlUser(), "status", svc, "--no-pager")); err != nil {
			m.progress().Warn("systemctl status %s reported non-zero (unit not active?): %v", svc, err)
		}
	}
	listed, err := m.listContainers(ctx)
	if err != nil {
		return err
	}
	var rows [][]string
	for _, row := range listed {
		if row.Name == m.name() {
			rows = append(rows, printableRow(row))
		}
	}
	r := m.report()
	r.Table([]string{"NAME", "IMAGE", "STATUS"}, rows)
	// Not under the Echo runner, which answers nothing: the note would say a deployed
	// broker is missing when no listing was ever made.
	if len(rows) == 0 && !m.isEcho() {
		r.Line("  (no container named %s on this host)", m.name())
	}
	return m.statusDetail(ctx)
}

// statusDetail appends the decoded inspect block to Status, warning rather than
// failing when it cannot be read.
//
// The call is made under the Echo runner too, so a preview still shows the `inspect`
// it would issue -- but the warning is suppressed there. Echo returns no output by
// construction, so the decode cannot succeed, and a preview complaining about a
// parse that was never going to happen reads as a real problem with the deployment.
func (m *Manager) statusDetail(ctx context.Context) error {
	if err := m.inspectAndReport(ctx, m.name()); err != nil && !m.isEcho() {
		m.progress().Warn("could not read %s state in detail: %v", m.name(), err)
	}
	return nil
}

// Describe prints detailed inspection output for this host's broker, the container
// analog of `kubectl describe pod`: `<runtime> container inspect` carries the health
// state, restart count, exit reason, mounts and resource limits that Status's one-line
// listing does not. Podman additionally shows the installed unit definition, which
// answers "what did we actually deploy" -- the summary is already in Status.
func (m *Manager) Describe(ctx context.Context) error {
	if m.P == config.Podman {
		svc := m.name() + ".service"
		if err := m.systemctl(ctx, systemd.Systemctl(m.systemctlUser(), "cat", svc)); err != nil {
			m.progress().Warn("systemctl cat %s failed (unit not installed?): %v", svc, err)
		}
	}
	return m.runCmd(ctx, func(base []string) runner.Cmd {
		c, _ := inspect.Containers(base, m.name())
		return c
	})
}

// Logs reads the broker container's logs, with the caller's log selection (follow,
// tail, since, timestamps -- containercli.LogOptions, whose values the caller has already
// validated). It prints what is buffered and exits unless the caller asked to follow.
//
// Following is the CALLER's choice, never hard-coded here: a hard-coded `-f` blocks
// until interrupted, and no way to ask for one cannot follow at all.
//
// previous is refused here as well as at pre-run, and that is not belt-and-braces.
// checkFlagPlatforms walks only the flags the operator actually SET, so a defaulted value
// never reaches it -- and this method can be called by something that is not this CLI.
// Neither engine keeps a previous container's log, so passing it through would ask for
// something that cannot exist.
func (m *Manager) Logs(ctx context.Context, o containercli.LogOptions, previous bool) error {
	if previous {
		return containercli.NoPrevious(containercli.Engine(m.P))
	}
	return m.runCmd(ctx, func(base []string) runner.Cmd { return containercli.Logs(base, m.name(), o) })
}

// CopyFrom copies files out of the broker container into the working directory,
// mirroring the k8s verb: each file is attempted, failures are reported per file,
// and the command exits non-zero if any failed.
func (m *Manager) CopyFrom(ctx context.Context, files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files specified to copy from the broker")
	}
	t := NewTransport(m.R, m.Cfg, m.P)
	return m.copyEach(files, "from", func(f string) error {
		// config.BaseName, not path.Base -- see the same call in k8s.Cluster.CopyFrom.
		local := config.BaseName(f)
		m.logf("copying %s from container %s", f, m.name())
		if err := t.Download(ctx, config.Primary, f, local); err != nil {
			return err
		}
		m.report().OK("%s -> %s", f, local)
		return nil
	})
}

// copyEach is the per-file loop both copy verbs share: every file is attempted, each
// outcome gets its own report line, and the command fails if ANY file did. Stopping at
// the first failure would hide the files that did copy, which is the opposite of what
// an operator retrieving diagnostics needs to know.
func (m *Manager) copyEach(files []string, direction string, one func(f string) error) error {
	var failed int
	for _, f := range files {
		if err := one(f); err != nil {
			m.report().Fail("%s: %v", f, err)
			failed++
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d file(s) failed to copy %s the broker", failed, len(files), direction)
	}
	return nil
}

// CopyInto copies local files into destDir inside the broker container.
func (m *Manager) CopyInto(ctx context.Context, files []string, destDir string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files specified to copy into the broker")
	}
	if destDir == "" {
		destDir = "."
	}
	t := NewTransport(m.R, m.Cfg, m.P)
	return m.copyEach(files, "into", func(f string) error {
		m.logf("copying %s into container %s:%s", f, m.name(), destDir)
		if err := t.UploadFile(ctx, config.Primary, f, destDir); err != nil {
			return err
		}
		m.report().OK("%s -> %s:%s", f, m.name(), destDir)
		return nil
	})
}

// CLI opens an interactive Solace CLI session inside the container.
func (m *Manager) CLI(ctx context.Context) error { return m.session("cli", "-A") }

// Shell opens an interactive shell inside the container.
func (m *Manager) Shell(ctx context.Context) error { return m.session("bash") }

// session attaches this terminal to argv inside the container: `exec -i -t <name> argv`.
func (m *Manager) session(argv ...string) error {
	base, err := m.base()
	if err != nil {
		return err
	}
	c, err := containercli.Exec(base, m.name(), containercli.ExecOptions{Stdin: true, TTY: true}, argv...)
	if err != nil {
		return err
	}
	return engine.Interactive(m.R, c)
}

// --- helpers ----------------------------------------------------------------

// systemctlUser says whether systemctl talks to the user manager (`--user`): rootless
// podman, which config derives into SystemctlUser from podman.rootful. It is the one
// source for the flag every systemd.* builder takes.
func (m *Manager) systemctlUser() bool { return m.Cfg.Podman.SystemctlUser != "" }

// systemctl runs c, a systemctl command built by the systemd package, through the
// Runner.
func (m *Manager) systemctl(ctx context.Context, c runner.Cmd) error {
	// Every podman systemd path goes through here, not just validate and deploy:
	// `broker start`/`stop`/`restart`/`remove`/`status` drive systemctl without ever
	// calling PrepHost, and on rootless they need the same user bus. ensureUserSession
	// is memoised, so this costs one probe per run (rootless.go).
	if err := m.ensureUserSession(ctx); err != nil {
		return err
	}
	return engine.Run(ctx, m.R, c)
}

// systemctlOutput captures c's stdout, for the state probes (`is-active`,
// `show -p NRestarts`) whose answer picks the next step rather than being shown.
func (m *Manager) systemctlOutput(ctx context.Context, c runner.Cmd) ([]byte, error) {
	// Same reason as systemctl above: the state probes run from commands that never
	// reach PrepHost, and a missing XDG_RUNTIME_DIR would read as "unit not found".
	if err := m.ensureUserSession(ctx); err != nil {
		return nil, err
	}
	return engine.Output(ctx, m.R, c)
}

// checkPodmanEUID is the executor-side call of the one guard definition
// (GuardPodmanEUID, euid.go). cli.prepare has already refused a mismatch before
// any command ran, so in production this never fires -- it is kept at the two
// sites where a wrong-account run does LASTING damage, Deploy's writes into the
// secret store and PrepHost's mkdir/chown, because an executor must not assume
// its caller checked. That is the same stance runtime() takes on execguard.Check.
func (m *Manager) checkPodmanEUID() error {
	return GuardPodmanEUID(m.Cfg, m.P, m.R, m.Geteuid())
}

// fileExists reports whether p exists (used to name a legacy bundle in a preview).
func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// platformTitle is the display name for the platform in Check output/errors.
func platformTitle(p config.Platform) string {
	if p == config.Podman {
		return "Podman"
	}
	return "Docker"
}

// orValue renders an empty string as fallback -- a package-local copy of the k8s
// report formatter (one small copy per package, not over-DRY'd across packages).
func orValue(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func orNone(s string) string { return orValue(s, "(none)") }

// setOrMissing renders a secret's presence without echoing it.
func setOrMissing(s string) string {
	if s == "" {
		return "MISSING"
	}
	return "set"
}

// secretSummary reports the externalized secrets by name, whether each value is
// present, and the mechanism this platform supplies them through. Values never
// appear.
func secretSummary(p config.Platform, secrets []render.ContainerSecret) string {
	store := "podman secret store"
	if p == config.Docker {
		store = "compose environment secrets"
	}
	parts := make([]string, 0, len(secrets))
	for _, s := range secrets {
		// The certificate's value is read from its files at deploy, never at report
		// time, so set/MISSING would always say MISSING. The source keys are the
		// honest answer; whether the files are readable is SecretPreflight's to say.
		if len(s.SourceFiles) > 0 {
			from := "tls.certKey + tls.cert"
			if len(s.SourceFiles) > 2 {
				from += " + tls.cas"
			}
			parts = append(parts, s.Name+"=(from "+from+")")
			continue
		}
		parts = append(parts, s.Name+"="+setOrMissing(s.Value))
	}
	return fmt.Sprintf("%s (%s)", strings.Join(parts, " "), store)
}

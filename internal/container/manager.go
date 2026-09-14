package container

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"solace/internal/broker"
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
// Manager's operations are node-local and the HA coordination (leader,
// redundancy) is a per-host handshake handled in package broker. Every mutating
// command routes through the engine.Runner, so the Echo runner records without running.
type Manager struct {
	R   engine.EnvRunner
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
	// RestartApproved pre-approves bouncing an already-running broker when the
	// deploy artifact changed (the --restart flag on `deploy broker`). It is an
	// approval, not an instruction: Restart() is the command that bounces one on
	// purpose.
	RestartApproved bool
	// Confirm asks the operator a yes/no question. nil declines, which is what a
	// non-interactive run must do: re-deploying should never bounce a running
	// broker unattended. The CLI wires it to the same prompt style as delete.
	Confirm func(question string) bool
}

// NewManager builds a container host Manager over the given runner, config and
// platform. It defaults Resolve to a net.LookupHost seam and Geteuid to os.Geteuid.
// In is left for the caller to set.
//
// There is deliberately no PSK generator here any more. This tool used to invent the
// redundancy pre-shared key on a first HA deploy and rewrite the env file it was handed;
// the key is the operator's to make now (config.validateContainerPSK refuses an empty one
// and prints the openssl command), and nothing in this package writes to an env file.
func NewManager(r engine.EnvRunner, cfg *config.Config, p config.Platform, log func(string, ...any), out io.Writer) *Manager {
	return &Manager{
		R:       r,
		Cfg:     cfg,
		P:       p,
		Log:     log,
		Out:     out,
		Resolve: func(host string) bool { _, err := net.LookupHost(host); return err == nil },
		Geteuid: os.Geteuid,
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

// isEcho reports whether the runner only echoes (engine.Echo), so file writes
// and DNS/euid probes are previewed rather than performed.
func (m *Manager) isEcho() bool { _, ok := m.R.(engine.Echo); return ok }

// runtime is the guarded runtime command (docker.command / podman.command): argv[0]
// plus any leading arguments that precede every call's own. It re-runs
// config.CheckCommand on every call -- the Manager is built straight from a
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
	return m.R.Run(ctx, r.Name(), r.Args(args...)...)
}

// output captures a runtime subcommand's stdout through the Runner.
func (m *Manager) output(ctx context.Context, args ...string) ([]byte, error) {
	r, err := m.runtime()
	if err != nil {
		return nil, err
	}
	return m.R.Output(ctx, r.Name(), r.Args(args...)...)
}

// composeCmd is the guarded compose invocation (docker only): whatever
// docker.compose names -- a host with only the standalone v1 binary sets it to
// `docker-compose` -- defaulting to the runtime's own `compose` subcommand. Both the
// derivation and the check live in config.ComposeCommand, so what runs here is
// exactly what Validate approved.
func (m *Manager) composeCmd() (config.Command, error) { return m.Cfg.ComposeCommand() }

// compose runs a compose subcommand (`<compose> args...`) through the Runner with
// this deployment's secret values in the child's environment, which is where the
// rendered compose file reads them from. Every compose call gets them, not just
// `up`: compose interpolates the whole model on `down` and `ps` too, and the
// values are always available from the config, so there is no case where leaving
// them out would be right.
func (m *Manager) compose(ctx context.Context, args ...string) error {
	c, err := m.composeCmd()
	if err != nil {
		return err
	}
	// preview here means "do not read the certificate from disk". It is true under
	// the Echo runner, and ALSO true for every verb that does not need a secret's
	// value: compose requires each declared secret to be DEFINED for `down`, `stop`,
	// `restart` and `ps`, but never reads it, so a moved certificate must not block
	// a teardown. Only `up` actually consumes the bytes.
	env, err := m.composeSecretEnv(m.isEcho() || !composeNeedsSecretValues(args))
	if err != nil {
		return err
	}
	return m.R.RunEnv(ctx, env, c.Name(), c.Args(args...)...)
}

// composeNeedsSecretValues reports whether this compose invocation will actually
// read a secret's contents. Only `up` creates containers and therefore materializes
// secrets; `down`, `stop`, `start`, `restart` and `ps` operate on what already
// exists. Keeping the list positive rather than negative means a new verb defaults
// to NOT reading a private key from disk.
func composeNeedsSecretValues(args []string) bool {
	for _, a := range args {
		if a == "up" {
			return true
		}
	}
	return false
}

// composeSecretEnv is the "VAR=value" list backing the compose file's
// environment-sourced secrets. It is the only place a secret value enters a child
// process, and it never reaches an argv.
func (m *Manager) composeSecretEnv(preview bool) ([]string, error) {
	secrets, err := ResolveSecretValues(m.Cfg, m.P, preview)
	if err != nil {
		return nil, err
	}
	env := make([]string, 0, len(secrets))
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
	if err := m.Reachable(ctx); err != nil {
		return err
	}
	return m.checkDNS(ctx)
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
			orNone(cfg.TLS.Cert), setOrMissing(cfg.TLS.CertKey), len(cfg.TLS.CAs))
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
		// baseDir is reported because it is where the broker's PRIVATE KEY lands on
		// this host, which is the most consequential new fact about a podman deploy
		// and the one thing an operator cannot infer from anywhere else.
		rows = append(rows, output.KV{Key: "podman", Value: fmt.Sprintf(
			"rootless=%t quadletDir=%s baseDir=%s", cfg.Podman.Rootless, cfg.Podman.QuadletDir, cfg.Podman.BaseDir)})
	} else {
		// The configured value, like the runtime line above -- CheckEnv reports what
		// the env file says, and Reachable (next in Check) is what fails loud if the
		// execution guard rejects it.
		rows = append(rows, output.KV{Key: "docker", Value: fmt.Sprintf(
			"compose=%s composeFile=%s", cfg.Docker.Compose, m.composeFile())})
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
	if _, err := m.R.Output(ctx, rt.Name(), rt.Args("version")...); err != nil {
		return fmt.Errorf("cannot reach the %s runtime %q (is it installed and running?): %w", platformTitle(m.P), rt, err)
	}
	if m.P == config.Docker {
		c, err := m.composeCmd()
		if err != nil {
			return err
		}
		if _, err := m.R.Output(ctx, c.Name(), c.Args("version")...); err != nil {
			return fmt.Errorf("cannot run the compose command %q (install the docker compose plugin, "+
				"or set docker.compose to this host's standalone 'docker-compose' binary): %w", c, err)
		}
	}
	return nil
}

// checkNoFile is the rootless-only half of the nofile limit. Both artifacts ask
// the engine for the configured limit (compose `ulimits.nofile`, quadlet
// `Ulimit=nofile=`), but a ROOTLESS container cannot raise nofile above the hard
// limit of the user invoking podman -- the kernel refuses, and the broker then
// starts against a limit far below what it needs and fails obscurely later.
// Rootful podman and docker are unaffected: their daemon/engine runs privileged
// and can raise the limit itself, which is why this runs only in that branch.
//
// It reports and refuses rather than fixing: raising a hard limit means editing
// host-wide security configuration as root, which is precisely what a rootless
// deployment exists to avoid. The message carries the exact drop-in to add.
func (m *Manager) checkNoFile(ctx context.Context) error {
	want := m.Cfg.ContainerBlock(m.P).Ulimits.NoFile
	_, hardWant := splitLimit(want)
	if hardWant == 0 {
		return nil // unlimited or unparseable: nothing to assert against
	}
	// `ulimit` is a shell builtin, so it needs a shell; this is the user's own
	// limit because prep runs as the user that will own the rootless container.
	out, err := m.R.Output(ctx, "sh", "-c", "ulimit -Hn")
	if m.isEcho() {
		m.report().KVRow(reportKeyWidth, "nofile", "skipped (preview)")
		return err
	}
	if err != nil {
		return fmt.Errorf("cannot read this user's hard nofile limit (`sh -c 'ulimit -Hn'`): %w", err)
	}
	got := strings.TrimSpace(string(out))
	if got == "unlimited" {
		m.report().OK("hard nofile limit: unlimited (need %d)", hardWant)
		return nil
	}
	hardGot, convErr := strconv.Atoi(got)
	if convErr != nil {
		return fmt.Errorf("cannot parse this user's hard nofile limit %q from `ulimit -Hn`: %w", got, convErr)
	}
	if hardGot >= hardWant {
		m.report().OK("hard nofile limit: %d (need %d)", hardGot, hardWant)
		return nil
	}
	// The account is named as a placeholder rather than resolved: prep may be
	// running as root against a rootless deployment (the warning above), in which
	// case this process's own user is the wrong one to write into the drop-in.
	soft, _ := splitLimit(want)
	return fmt.Errorf("rootless podman: this user's hard nofile limit is %d, but "+
		"podman.container.ulimits.nofile needs %d -- a rootless container cannot raise it above the "+
		"user's own hard limit, so the broker would start under-provisioned.\n"+
		"  Add this as root to /etc/security/limits.d/99-solace.conf, replacing <user> with the account "+
		"that runs the container, then log out and back in (a new login session is what re-reads it):\n"+
		"    <user> hard nofile %d\n"+
		"    <user> soft nofile %d",
		hardGot, hardWant, hardWant, soft)
}

// splitLimit parses a `soft:hard` ulimit pair, or a single value meaning both.
// A non-numeric half (including "-1", meaning unlimited) reports 0, which
// checkNoFile reads as "nothing to assert".
func splitLimit(v string) (soft, hard int) {
	s, h := v, v
	if i := strings.Index(v, ":"); i >= 0 {
		s, h = v[:i], v[i+1:]
	}
	soft, _ = strconv.Atoi(strings.TrimSpace(s))
	hard, _ = strconv.Atoi(strings.TrimSpace(h))
	if soft < 0 {
		soft = 0
	}
	if hard < 0 {
		hard = 0
	}
	return soft, hard
}

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
	// Same invariant Deploy enforces via checkPodmanEUID, and the same call: a
	// rootful mkdir/chown here would leave a data directory the later rootless
	// deploy cannot use, so this is a hard stop rather than a warning the
	// operator could miss.
	if m.P == config.Podman {
		if err := m.checkPodmanEUID(); err != nil {
			return err
		}
	}

	cb := m.Cfg.ContainerBlock(m.P)
	m.logf("preparing data directory %s", cb.DataDir)
	if err := m.R.Run(ctx, "mkdir", "-p", cb.DataDir); err != nil {
		return fmt.Errorf("create data dir %q: %w", cb.DataDir, err)
	}
	if m.P == config.Podman && m.Cfg.Podman.Rootless {
		// Rootless podman: the container's uid:gid maps through subuid/subgid, so
		// enter the user namespace to apply the ownership the container will see.
		if err := m.run(ctx, "unshare", "chown", cb.RunUser, cb.DataDir); err != nil {
			return fmt.Errorf("chown data dir %q (rootless): %w", cb.DataDir, err)
		}
		if err := m.checkNoFile(ctx); err != nil {
			return err
		}
	} else if err := m.R.Run(ctx, "chown", cb.RunUser, cb.DataDir); err != nil {
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
// an argv or an echoed command (§3).
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
	if err := m.R.RunInput(ctx, []byte(pass), r.Name(), r.Args(args...)...); err != nil {
		return fmt.Errorf("%s login to registry %s failed: %w", platformTitle(m.P), orNone(registry), err)
	}
	return nil
}

// --- Deploy -----------------------------------------------------------------

// Deploy renders and starts the broker container for role's identity. Podman
// uses a systemd quadlet unit; docker uses a compose file. The broker's secret
// settings are externalized first, since the rendered artifact only references
// them by name.
func (m *Manager) Deploy(ctx context.Context, role config.Role) error {
	// Ahead of every write and every mutating call: prepareSecrets loads values
	// into podman's store and writeArtifact drops a unit/compose file on disk, so
	// an engine that cannot be reached must stop the deploy before either.
	if err := m.Preflight(ctx); err != nil {
		return err
	}
	id := m.Cfg.ResolveNode(role)
	if m.P == config.Podman {
		if err := m.checkPodmanEUID(); err != nil {
			return err
		}
	}
	if err := m.prepareSecrets(ctx); err != nil {
		return err
	}
	if m.P == config.Podman {
		return m.deployPodman(ctx, id)
	}
	return m.deployDocker(ctx, id)
}

// prepareSecrets externalizes the broker's secret settings so the deploy artifact
// can reference them instead of carrying them: podman loads each into its own
// secret store, while docker needs nothing prepared -- its compose secrets read
// this deployment's values from the environment `compose` is given (see compose),
// so no secret is ever written to this host's disk. An empty value fails loud here
// rather than deploying a broker with no password -- except under the Echo runner, which
// must stay previewable before `broker deploy` has generated the HA pre-shared key.
func (m *Manager) prepareSecrets(ctx context.Context) error {
	// Skipped under the Echo runner: a preview must not need secret values on disk, and
	// `broker generate` renders name-level references only, so there is nothing to check.
	if !m.isEcho() {
		// Preflight FIRST, then resolve. The order matters for the message: an env
		// file that is both pre-`broker deploy` (empty PSK) and missing its cert files
		// should report the PSK and its `broker deploy` hint, not the certificate.
		if err := render.SecretPreflight(m.Cfg, m.P); err != nil {
			return err
		}
	}
	// Resolve on BOTH platforms, and discard the result on docker. Docker needs no
	// secret prepared -- its values ride the compose child's environment -- but
	// reading the certificate HERE is what makes an unreadable one fail before
	// deployDocker rewrites the compose file, so "nothing happened" stays true. That
	// is the same invariant the podman side gets by writing the bundle before the unit.
	secrets, err := ResolveSecretValues(m.Cfg, m.P, m.isEcho())
	if err != nil {
		return err
	}
	if m.P == config.Podman {
		return m.CreatePodmanSecrets(ctx, secrets)
	}
	return nil
}

// CreatePodmanSecrets loads each secret into podman's secret store, feeding the
// value on stdin so it never reaches an argv or the dry-run echo (§3).
//
// Remove-then-create, rather than `create --replace`, and the reason is the version
// floor. `--replace` needs podman 4.7, while the rest of this wiring needs only 4.5
// -- so using it would raise the floor of the whole tool for one flag whose effect
// two commands reproduce exactly. `rm --ignore` makes a missing secret a success, so
// the pair is idempotent the same way --replace was: a redeploy with a rotated value
// works, and a first deploy with nothing in the store works.
//
// A failure of the rm half is NOT warned away. --ignore already absorbs the only
// benign case, so anything left is real -- a store this user cannot write, say --
// and continuing would create a secret next to one that could not be removed.
//
// Exported so a caller can rotate ONE secret through this exact
// path and then restarts the unit, to settle whether a quadlet re-reads the store
// on start (deployPodman's ASSUMED, NOT VERIFIED branch). Going through Deploy
// instead would exercise the branching rather than the question.
func (m *Manager) CreatePodmanSecrets(ctx context.Context, secrets []render.ContainerSecret) error {
	r, err := m.runtime()
	if err != nil {
		return err
	}
	for _, s := range secrets {
		if err := m.R.Run(ctx, r.Name(), r.Args("secret", "rm", "--ignore", s.Name)...); err != nil {
			return fmt.Errorf("remove the existing podman secret %q before recreating it (from %s): %w",
				s.Name, s.ConfigKey, err)
		}
		if err := m.R.RunInput(ctx, []byte(s.Value), r.Name(),
			r.Args("secret", "create", s.Name, "-")...); err != nil {
			return fmt.Errorf("create podman secret %q from %s: %w", s.Name, s.ConfigKey, err)
		}
	}
	return nil
}

// writeCertBundle writes podman's copy of the server-certificate bundle to the
// host, returning whether it changed so deployPodman can fold that into the restart
// decision. A deployment with no TLS configured writes nothing and reports no change.
//
// The Echo guard sits HERE rather than being inherited from writeArtifact, and that
// is load-bearing: writeArtifact's body is an ordinary argument, so the caller has
// already evaluated it: broker.ServerCertBundle would have read both files before
// writeArtifact could decline. A dry-run must stay possible before the certificate
// exists on this host.
//
// Mode 0600 in a 0700 directory, and the chmod is explicit because os.WriteFile
// applies its permission only when CREATING a file -- an existing bundle left at
// 0644 by anything else would be truncated and rewritten still 0644, which for a
// file containing a private key is not a mode to inherit.
func (m *Manager) writeCertBundle() (bool, error) {
	if m.Cfg.TLS.Cert == "" || m.Cfg.TLS.CertKey == "" {
		return false, nil
	}
	path := filepath.FromSlash(render.ServerCertBundlePath(m.Cfg))
	if m.isEcho() {
		m.progress().Info("would write server certificate bundle %s (skipped under dry-run).", path)
		return true, nil
	}
	bundle, err := broker.ServerCertBundle(m.Cfg)
	if err != nil {
		return false, err
	}
	changed, err := m.writeArtifact(path, bundle, "server certificate bundle", 0o700)
	if err != nil {
		return false, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return false, fmt.Errorf("restrict the server certificate bundle %q to 0600 (it holds the private key): %w",
			path, err)
	}
	return changed, nil
}

// removeCertBundle deletes podman's copy of the bundle on teardown. It runs with the
// other things the container consumes, after the unit is confirmed gone.
//
// The failure is FATAL, unlike a leftover store secret's warning, and the difference
// is the contents: silently leaving a private key on the host is the outcome least
// like the rest of this teardown's posture. A missing file is not a failure -- a
// deployment that never had TLS, or whose tls.cert was unset since deploy, has
// nothing to remove. Only the FILE goes; the directory may hold other brokers'
// bundles.
func (m *Manager) removeCertBundle() error {
	if m.Cfg.TLS.Cert == "" || m.Cfg.TLS.CertKey == "" {
		return nil
	}
	path := filepath.FromSlash(render.ServerCertBundlePath(m.Cfg))
	if m.isEcho() {
		m.progress().Info("would remove server certificate bundle %s (skipped under dry-run).", path)
		return nil
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("remove the server certificate bundle %q: it holds the broker's PRIVATE KEY and must "+
			"not be left on this host -- remove it by hand and re-run: %w", path, err)
	}
	m.progress().OK("removed server certificate bundle %s", path)
	return nil
}

// writeArtifact writes a rendered deploy artifact, reporting whether it differs
// from what was already on disk. An unchanged artifact is not rewritten, which is
// what lets Deploy tell "nothing to do" apart from "the running broker is now
// stale and needs a bounce". Under the Echo runner nothing is written and the artifact
// counts as changed, so the preview shows the work a real run would do.
func (m *Manager) writeArtifact(path string, body []byte, what string, dirMode os.FileMode) (bool, error) {
	if m.isEcho() {
		m.progress().Info("would write %s %s (skipped under dry-run).", what, path)
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

// approveRestart decides whether an already-running broker may be bounced to pick
// up a changed artifact. --restart pre-approves it; otherwise the Confirm seam
// asks. A non-interactive run without --restart declines, so a scripted deploy
// leaves the new artifact in place and says so rather than dropping messaging
// traffic on its own. Deliberately its own flag: bouncing a live broker is its own
// decision, the same separation --delete-data has from --no-prompt.
func (m *Manager) approveRestart(what string) bool {
	if m.RestartApproved {
		return true
	}
	if m.Confirm == nil {
		return false
	}
	return m.Confirm(fmt.Sprintf("Restart %s now to apply the change?", what))
}

// staleWarning tells the operator the artifact is applied but the running broker
// is still on the old one, and how to finish the job.
func (m *Manager) staleWarning(what string) {
	m.progress().Warn("%s is applied, but the running broker still uses the previous one.", what)
	m.progress().Warn("re-run with --restart (or restart it yourself) to pick up the change.")
}

func (m *Manager) deployPodman(ctx context.Context, id config.NodeIdentity) error {
	// The bundle FIRST, before the unit that bind-mounts it: the file has to exist
	// before any `systemctl start`, and a failure to build it must abort before any
	// host state changes at all.
	certChanged, err := m.writeCertBundle()
	if err != nil {
		return err
	}
	unit := filepath.ToSlash(filepath.Join(m.Cfg.Podman.QuadletDir, m.name()+".container"))
	svc := m.name() + ".service"
	changed, err := m.writeArtifact(unit, render.Quadlet(m.Cfg, id), "quadlet unit", 0o755)
	if err != nil {
		return err
	}
	// A rotated certificate under a byte-identical unit must still bounce the broker.
	// `changed` is the ONLY input to the restart decision below, so without this a
	// renewed certificate would print "nothing to do" and the broker would keep
	// serving the old one.
	changed = changed || certChanged
	// daemon-reload runs either way: a unit that was already correct may still be
	// unknown to systemd (a fresh host, or a manual removal).
	if err := m.systemctl(ctx, "daemon-reload"); err != nil {
		return err
	}
	// ASSUMED, NOT VERIFIED: quadlet replaces the container on each service start, so
	// starting a stopped unit re-reads the store and picks up a rotated secret. Docker's
	// equivalent branch does NOT behave that way -- a stopped container is started with
	// the credentials it was created with, which is why deployDocker force-recreates. If
	// podman turns out to share that behaviour, this branch needs the same fix. To check:
	// stop the unit, rotate a value, `systemctl start`, then read
	// /mnt/secrets/username_admin_password inside the container.
	// serviceActive is read in its LOSSY form here on purpose, unlike deployDocker,
	// which aborts when its probe cannot answer (containerRunningKnown). The
	// consequence of guessing differs entirely: an unanswered probe here sends an
	// already-running unit a `systemctl start`, which systemd makes a no-op, while
	// the docker branch below would force-recreate and destroy a live container.
	// Guessing wrong is free on this side, so it is not worth failing a deploy over.
	// (If the ASSUMED note above turns out to be false and a start DOES replace the
	// container, that changes -- this branch would then need the same guard.)
	if !m.serviceActive(ctx) {
		return m.systemctl(ctx, "start", svc)
	}
	// `systemctl start` on an active unit is a no-op, so an already-running broker
	// would silently keep the old image. Restart is the only way to apply a change.
	if !changed {
		// The store secrets were just replaced, but a running container holds the
		// values it started with, and the unit is byte-identical -- so a rotated
		// secret needs the same explicit restart docker's does.
		if m.RestartApproved {
			m.progress().Info("unit unchanged; restarting %s to apply any rotated secret.", svc)
			return m.systemctl(ctx, "restart", svc)
		}
		m.progress().Info("%s is already active on this unit -- nothing to do.", svc)
		m.progress().Info("if you rotated a secret, re-run with --restart to restart the service with it.")
		return nil
	}
	if !m.approveRestart(svc) {
		m.staleWarning("the quadlet unit")
		return nil
	}
	return m.systemctl(ctx, "restart", svc)
}

func (m *Manager) deployDocker(ctx context.Context, id config.NodeIdentity) error {
	file := m.composeFile()
	changed, err := m.writeArtifact(file, render.Compose(m.Cfg, id), "compose file", 0)
	if err != nil {
		return err
	}
	// --force-recreate rather than a plain up: a container that exists but is
	// stopped would otherwise be *started*, and a compose secret's value is baked in
	// at creation -- so a start replays the credentials it was created with and a
	// rotated password would be silently ignored. Recreating costs nothing here (a
	// stopped broker carries no traffic, and the data dir is a host bind mount), and
	// it makes the postcondition honest: after a deploy, the running container
	// reflects the current config AND the current secrets. With no container at all
	// the flag changes nothing -- compose creates one either way.
	//
	// "no container at all" has to be ANSWERED, not assumed (M4): a `ps` that
	// cannot run says nothing about whether the container exists, and reading that
	// silence as "not running" -- containerRunning's own lossy shape -- would take
	// this branch and force-recreate a LIVE broker on a transient engine failure,
	// with none of the consent the artifact-changed branch below requires for the
	// identical action. containerRunningKnown is read directly so an unanswered
	// probe stops here instead of guessing either way; both branches below assume
	// the state IS known.
	running, answered, probeErr := m.containerRunningKnown(ctx)
	if !answered {
		return fmt.Errorf("cannot tell whether container %s is running -- the `ps` probe itself "+
			"failed, so it is unsafe to guess whether to create or recreate it; check `%s ps` by "+
			"hand and re-run: %w", m.name(), m.Cfg.ContainerRuntime(m.P).Name(), probeErr)
	}
	if !running {
		return m.compose(ctx, "-f", file, "up", "-d", "--force-recreate")
	}
	// `compose up -d` recreates the container when the file changed, which bounces
	// a running broker -- the same hazard podman has, so it takes the same consent.
	if changed {
		if !m.approveRestart("container " + m.name()) {
			m.staleWarning("the compose file")
			return nil
		}
		return m.compose(ctx, "-f", file, "up", "-d")
	}
	// Nothing changed in the artifact -- but a rotated secret is invisible here by
	// design: the value lives only in the config and this host's environment, so
	// there is no on-disk copy to diff. --restart is therefore the explicit way to
	// push a rotated password or key into the running broker, and it has to force
	// the recreate that an unchanged compose file would otherwise skip.
	if m.RestartApproved {
		m.progress().Info("compose file unchanged; recreating container %s to apply any rotated secret.", m.name())
		return m.compose(ctx, "-f", file, "up", "-d", "--force-recreate")
	}
	m.progress().Info("container %s is already running on this compose file -- nothing to do.", m.name())
	m.progress().Info("if you rotated a secret, re-run with --restart to recreate the container with it.")
	return nil
}

// serviceActive reports whether this host's broker unit is already active, so
// Deploy can restart it instead of issuing a no-op start. Under the Echo runner no host
// state is probed (matching checkPodmanEUID/checkDNS) and the answer is "not
// active", so the preview shows the plain start path.
func (m *Manager) serviceActive(ctx context.Context) bool {
	if m.isEcho() {
		return false
	}
	state, _ := m.serviceState(ctx)
	return state == "active"
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
	out, _ := m.systemctlOutput(ctx, "is-active", m.name()+".service")
	state = strings.TrimSpace(string(out))
	return state, state != ""
}

// stoppedStates are the systemd states that CONFIRM the unit is not running.
// "unknown" belongs here because it is systemd saying it has no such unit, which
// is equally safe to remove. Every other answer -- including no answer -- leaves
// the broker possibly serving, so a teardown refuses rather than guessing.
var stoppedStates = map[string]bool{"inactive": true, "failed": true, "unknown": true}

// containerRunning reports whether this host's broker container is up,
// collapsing a failed probe into the same "not running" answer as a genuine
// absence. Nothing in this package reads it any more (M4): its one caller,
// deployDocker, now reads containerRunningKnown directly, because a false
// negative here used to be expensive -- Deploy would take the not-running
// branch and force-recreate a live broker on a transient probe failure without
// asking. Kept (not deleted) because TestContainerRunningMatchesNameExactly
// still exercises it directly.
//
// The name is matched in Go rather than by `--filter name=`, which both engines
// treat as an unanchored REGEX: a bare name is a substring match, so a `solace`
// deployment would see a running `solace-edge` as its own, and a name carrying '.'
// (which the schema allows) would match any character there. Anchoring the pattern
// fixes the first but leaves this decision resting on regex semantics that vary by
// engine. Comparing the listed names exactly depends on nothing.
func (m *Manager) containerRunning(ctx context.Context) bool {
	running, _, _ := m.containerRunningKnown(ctx)
	return running
}

// containerRunningKnown is containerRunning plus the two things containerRunning
// cannot express: whether the probe answered at all, and its own error when it
// didn't. A `ps` that could not run says nothing about the container, and a
// caller that reads that silence as "not running" would act on a guess --
// stopAndRemove refuses a teardown on it (the docker half of the systemd case
// serviceState documents) and deployDocker refuses a force-recreate on it, for
// the same reason.
func (m *Manager) containerRunningKnown(ctx context.Context) (running, answered bool, err error) {
	if m.isEcho() {
		return false, true, nil
	}
	out, err := m.output(ctx, "ps", "--filter", "status=running", "--format", "{{.Names}}")
	if err != nil {
		return false, false, err
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == m.name() {
			return true, true, nil
		}
	}
	return false, true, nil
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
		m.progress().OK("data directory %s deleted.", m.Cfg.ContainerBlock(m.P).DataDir)
		return nil
	}
	m.progress().Info("data directory %s kept (pass --delete-data to remove it).", m.Cfg.ContainerBlock(m.P).DataDir)
	return nil
}

func (m *Manager) deletePodman(ctx context.Context) error {
	svc := m.name() + ".service"
	if err := m.systemctl(ctx, "stop", svc); err != nil {
		// "podman info" (Preflight) proves the engine is reachable, not that this
		// unit actually stopped, so a failed stop needs systemd's own confirmation
		// before the unit is removed out from under a broker that is still serving
		// traffic -- and before Delete goes on to rm -rf the data directory.
		// Only a state that CONFIRMS the unit is stopped continues; "still active"
		// and "no answer at all" both stop here, because the case this exists for
		// (an unreachable rootless user session) produces the second one.
		state, answered := m.serviceState(ctx)
		if !answered || !stoppedStates[state] {
			reported := "state " + state
			if !answered {
				reported = "nothing -- the state probe itself failed"
			}
			return fmt.Errorf("stop %s failed and it is NOT confirmed stopped (systemd reports %s); "+
				"the quadlet unit and the data directory are left in place. Stop it by hand "+
				"(`systemctl %s`) and re-run: %w",
				svc, reported, strings.Join(m.systemctlArgs([]string{"status", svc}), " "), err)
		}
		m.progress().Warn("stopping %s failed but systemd reports it %s; continuing: %v", svc, state, err)
	}
	unit := filepath.ToSlash(filepath.Join(m.Cfg.Podman.QuadletDir, m.name()+".container"))
	if m.isEcho() {
		m.progress().Info("would remove quadlet unit %s (skipped under dry-run).", unit)
	} else if err := os.Remove(unit); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove quadlet unit %q: %w", unit, err)
	} else {
		m.progress().OK("removed quadlet unit %s", unit)
	}
	if err := m.systemctl(ctx, "daemon-reload"); err != nil {
		return err
	}
	// With the other things the container consumed, now that the unit is gone.
	if err := m.removeCertBundle(); err != nil {
		return err
	}
	return m.removePodmanSecrets(ctx)
}

// removePodmanSecrets removes every secret this deployment loaded into podman's
// secret store, using the same render.ContainerSecrets(m.Cfg, m.P) that
// CreatePodmanSecrets reads them from -- so the create and remove lists cannot
// drift apart. It runs only after the container/unit is gone (deletePodman's
// last step): a leftover secret must not fail a teardown that otherwise
// succeeded, so a "not found" (or any other) rm failure is a warning, not an
// error, and every secret is still attempted.
func (m *Manager) removePodmanSecrets(ctx context.Context) error {
	secrets := render.ContainerSecrets(m.Cfg, m.P)
	if len(secrets) == 0 {
		return nil
	}
	r, err := m.runtime()
	if err != nil {
		return err
	}
	for _, s := range secrets {
		if m.isEcho() {
			m.progress().Info("would remove podman secret %s (skipped under dry-run).", s.Name)
			continue
		}
		// --ignore, matching CreatePodmanSecrets so one argv shape serves both call
		// sites. It also makes the warning below honest: "already removed" stops
		// being a plausible cause once a missing secret is a success, so what is
		// left really is a failure worth naming.
		if err := m.R.Run(ctx, r.Name(), r.Args("secret", "rm", "--ignore", s.Name)...); err != nil {
			m.progress().Warn("removing podman secret %s failed: %v", s.Name, err)
			continue
		}
		m.progress().OK("removed podman secret %s", s.Name)
	}
	return nil
}

func (m *Manager) deleteDocker(ctx context.Context) error {
	file := m.composeFile()
	if m.isEcho() || fileExists(file) {
		return m.compose(ctx, "-f", file, "down")
	}
	// No compose file on disk: fall back to stop/rm by container name.
	return m.stopAndRemove(ctx)
}

func (m *Manager) stopAndRemove(ctx context.Context) error {
	if !m.containerExists(ctx) {
		// `docker rm` on a name that was never deployed exits non-zero, which would
		// otherwise turn "reset after a failed deploy" into an error -- every other
		// removal path in the tool (--ignore-not-found) no-ops instead, and this is
		// that same case for the no-compose-file fallback.
		m.progress().Info("no container named %s on this host -- nothing to remove.", m.name())
		return nil
	}
	if err := m.run(ctx, "stop", m.name()); err != nil {
		// Mirrors deletePodman's stop check: "docker info" (Preflight) proves the
		// engine is reachable, not that this container actually stopped. A probe
		// that cannot answer counts as "still running" for the same reason it does
		// there -- silence is not confirmation.
		running, answered, _ := m.containerRunningKnown(ctx)
		if running || !answered {
			why := "it is still running"
			if !answered {
				why = "it could not be confirmed stopped (the `ps` probe itself failed)"
			}
			return fmt.Errorf("stop container %s failed and %s; stop it by hand (check `%s ps`) "+
				"and re-run: %w", m.name(), why, m.Cfg.ContainerRuntime(m.P).Name(), err)
		}
		m.progress().Warn("stopping container %s failed (already stopped?): %v", m.name(), err)
	}
	return m.run(ctx, "rm", m.name())
}

// containerExists reports whether this host has ANY container -- running or
// stopped -- with this deployment's name, so stopAndRemove can no-op when
// nothing was ever deployed instead of running stop/rm into a guaranteed
// docker failure.
//
// Matched in Go against a `ps --all` listing rather than `--filter name=`,
// for the same reason containerRunning is: that filter is an unanchored
// regex on both engines, so a bare name substring-matches a sibling
// deployment and an unescaped '.' (which the schema allows in a container
// name) matches any character.
func (m *Manager) containerExists(ctx context.Context) bool {
	if m.isEcho() {
		return false // unreachable in practice: deleteDocker takes the compose branch under Echo
	}
	out, err := m.output(ctx, "ps", "--all", "--format", "{{.Names}}")
	if err != nil {
		return true // let stop/rm run and surface the real engine error rather than hiding it
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == m.name() {
			return true
		}
	}
	return false
}

func (m *Manager) purgeData(ctx context.Context) error {
	dir := m.Cfg.ContainerBlock(m.P).DataDir
	m.logf("removing data directory %s", dir)
	if m.P == config.Podman && m.Cfg.Podman.Rootless {
		return m.run(ctx, "unshare", "rm", "-rf", dir)
	}
	return m.R.Run(ctx, "rm", "-rf", dir)
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
		return m.systemctl(ctx, "start", m.name()+".service")
	}
	if file := m.composeFile(); m.isEcho() || fileExists(file) {
		return m.compose(ctx, "-f", file, "start")
	}
	return m.run(ctx, "start", m.name())
}

// Stop shuts the broker down without removing it. The container and its data
// survive -- `remove broker` is what deletes them.
func (m *Manager) Stop(ctx context.Context) error {
	if err := m.Preflight(ctx); err != nil {
		return err
	}
	if m.P == config.Podman {
		return m.systemctl(ctx, "stop", m.name()+".service")
	}
	if file := m.composeFile(); m.isEcho() || fileExists(file) {
		return m.compose(ctx, "-f", file, "stop")
	}
	return m.run(ctx, "stop", m.name())
}

// Restart bounces the broker in place. It applies nothing new: a changed deploy
// artifact needs `deploy broker --restart`, which rewrites the artifact first.
func (m *Manager) Restart(ctx context.Context) error {
	if err := m.Preflight(ctx); err != nil {
		return err
	}
	if m.P == config.Podman {
		return m.systemctl(ctx, "restart", m.name()+".service")
	}
	if file := m.composeFile(); m.isEcho() || fileExists(file) {
		return m.compose(ctx, "-f", file, "restart")
	}
	return m.run(ctx, "restart", m.name())
}

// --- Status / Logs / CLI / Shell --------------------------------------------

// solaceImageMarker identifies a Solace broker container by the image it was
// created from. One substring covers every spelling in circulation --
// solace-pubsub-standard, solace-pubsub-enterprise, and the pubsubplus- variants --
// and it is matched against the image column rather than the container name because
// the name is the operator's to choose (container.name) while the image is not.
const solaceImageMarker = "solace-pubsub"

// psFormat is the row this tool prints for a discovered container. Explicit rather
// than the engine's default table, so the columns are the same on docker and podman.
const psFormat = "{{.Names}}\t{{.Image}}\t{{.Status}}"

// psTableFormat is psFormat with the ENGINE rendering the header and column widths, for
// Status, which streams `ps` straight through rather than parsing it the way StatusAll
// does.
//
// It exists to drop PORTS. A broker publishes a dozen or more ports, so the engine's
// default table spends most of a terminal line on
// "0.0.0.0:55555->55555/tcp, [::]:55555->55555/tcp, 0.0.0.0:8080->8080/tcp, ..." and wraps
// every other column into illegibility -- on the one report whose whole job is to answer
// "is it up". Nothing is lost: the ports are in the env file that chose them, in the
// artifact `broker generate` prints, and in the inspect block this same command appends
// below. NAMES, IMAGE and STATUS are what this report is for, and naming them explicitly
// also keeps docker and podman showing the same columns in the same order.
const psTableFormat = "table " + psFormat

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
	raw, err := m.output(ctx, "ps", "--all", "--format", psFormat)
	if err != nil {
		return fmt.Errorf("listing containers: %w", err)
	}
	names := solaceRows(r, string(raw))
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

// engineNameRE is what a container name returned BY THE ENGINE must look like before
// this tool will put it back into an argument vector. It is docker's and podman's own
// grammar for a container name, deliberately no wider.
//
// The names solaceRows returns did not come from the env file: they came out of
// `<runtime> ps` on this host, which makes them untrusted input on the way back in.
// Nothing about that is hypothetical to guard -- a name is created by whoever ran the
// engine, not necessarily by this tool -- and while argv exec means a metacharacter is
// inert, these names also reach report tables, logs and whatever a ticket quotes.
// Validating at the boundary is cheaper than trusting every consumer downstream.
var engineNameRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)

// solaceRows prints the header plus every row whose image names a Solace broker,
// and returns the container names it kept so a caller can go deeper on each.
//
// A row whose name does not match the engine's own grammar is still SHOWN -- it is a
// real container on this host and hiding it would be worse -- but its name is not
// returned, so nothing further is run against it. Skipped out loud rather than
// silently dropped: a missing row reads as "no such container".
func solaceRows(s *output.Sink, raw string) []string {
	lines := strings.Split(strings.TrimRight(raw, "\n"), "\n")
	var names []string
	var rows [][]string
	for _, ln := range lines {
		cols := strings.Split(ln, "\t")
		if len(cols) < 3 || !strings.Contains(cols[1], solaceImageMarker) {
			continue
		}
		rows = append(rows, []string{cols[0], cols[1], cols[2]})
		if !engineNameRE.MatchString(cols[0]) {
			// s.Line, not s.Warn: this is a report body, and Warn is the stderr voice
			// (internal/output). The note rides with the row it is about.
			s.Line("  (not inspected: %q is not a name this tool will pass back to the engine)", cols[0])
			continue
		}
		names = append(names, cols[0])
	}
	// Table computes each column from what is actually in it, so a long image ref
	// widens the column instead of running into the next one -- the fixed %-24s/%-48s
	// this replaces did both, depending on the name.
	s.Table([]string{"NAME", "IMAGE", "STATUS"}, rows)
	if len(names) == 0 {
		s.Line("  (no Solace broker containers on this host)")
	}
	return names
}

// Status reports the container's state. Podman also shows its systemd unit.
//
// The `ps` listing is what the engine says about the container; the block after it
// is what this tool reads out of `inspect` (inspect.go) -- health under whichever
// spelling this engine uses, and a restart count that on podman comes from systemd
// rather than from the engine's own counter, which a quadlet restart leaves at zero.
// Those are the two facts an operator asking "is my broker healthy" needs and
// neither engine's `ps` line carries.
//
// A failed inspect is a WARNING, not a failure: the listing above it already
// answered whether the container exists, so the command has reported something
// useful and must not exit non-zero because the extra detail was unavailable.
//
// Docker used to run `compose ps` first as well. It listed the same single container the
// filtered `ps` does, so the only thing it added was a second row -- carrying the widest
// PORTS column in the output, because compose prints every published port with both host
// bindings. It could not be narrowed in place either: compose's own --format takes only
// `table` or `json`, never a column template, so there is no compose-side equivalent of
// psTableFormat. Dropping it costs nothing a reader of this report wanted: the listing
// below reports the container whether it is running or stopped, and the compose FILE's
// own existence and path are already in `broker check`.
func (m *Manager) Status(ctx context.Context) error {
	if m.P == config.Podman {
		// The unit is podman's half of "is it up": a quadlet container that systemd
		// never started has no engine-side row to find.
		svc := m.name() + ".service"
		if err := m.systemctl(ctx, "status", svc, "--no-pager"); err != nil {
			m.progress().Warn("systemctl status %s reported non-zero (unit not active?): %v", svc, err)
		}
	}
	if err := m.run(ctx, "ps", "--all", "--filter", "name="+exactName(m.name()),
		"--format", psTableFormat); err != nil {
		return err
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
// analog of `kubectl describe pod`: `<runtime> inspect` carries the health state,
// restart count, exit reason, mounts and resource limits that Status's one-line
// listing does not. Podman additionally shows the installed unit definition, which
// answers "what did we actually deploy" -- the summary is already in Status.
func (m *Manager) Describe(ctx context.Context) error {
	if m.P == config.Podman {
		svc := m.name() + ".service"
		if err := m.systemctl(ctx, "cat", svc); err != nil {
			m.progress().Warn("systemctl cat %s failed (unit not installed?): %v", svc, err)
		}
	}
	return m.run(ctx, "inspect", m.name())
}

// Logs reads the broker container's logs, appending the caller's log-selection
// tokens (-f, --tail, --since, --timestamps -- the spellings kubectl and both engines
// share). It prints what is buffered and exits unless the caller asked to follow.
//
// It used to hard-code `-f` with no way to switch it off, which made
// `solace-util logs broker` block until interrupted on docker and podman while the
// kubernetes side could not follow at all: the two platforms were broken in opposite
// directions.
//
// `-p`/`--previous` is refused here as well as at pre-run, and that is not
// belt-and-braces. checkFlagPlatforms walks only the flags the operator actually SET,
// so a defaulted value never reaches it -- and this method can be called by something
// that is not this CLI. Neither engine keeps a previous container's log, so passing it
// through would ask for something that cannot exist.
func (m *Manager) Logs(ctx context.Context, extra ...string) error {
	for _, a := range extra {
		if a == "-p" || a == "--previous" {
			return fmt.Errorf("--previous is not supported on %s: neither engine keeps a previous "+
				"container's log, so there is nothing to read", m.P)
		}
	}
	return m.run(ctx, append([]string{"logs"}, append(extra, m.name())...)...)
}

// CopyFrom copies files out of the broker container into the working directory,
// mirroring the k8s verb: each file is attempted, failures are reported per file,
// and the command exits non-zero if any failed.
func (m *Manager) CopyFrom(ctx context.Context, files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files specified to copy from the broker")
	}
	t := NewTransport(m.R, m.Cfg, m.P)
	var failed int
	for _, f := range files {
		// config.BaseName, not path.Base -- see the same call in k8s.Cluster.CopyFrom.
		local := config.BaseName(f)
		m.logf("copying %s from container %s", f, m.name())
		if err := t.Download(ctx, config.Primary, f, local); err != nil {
			m.report().Fail("%s: %v", f, err)
			failed++
			continue
		}
		m.report().OK("%s -> %s", f, local)
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d file(s) failed to copy from the broker", failed, len(files))
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
	var failed int
	for _, f := range files {
		m.logf("copying %s into container %s:%s", f, m.name(), destDir)
		if err := t.UploadFile(ctx, config.Primary, f, destDir); err != nil {
			m.report().Fail("%s: %v", f, err)
			failed++
			continue
		}
		m.report().OK("%s -> %s:%s", f, m.name(), destDir)
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d file(s) failed to copy into the broker", failed, len(files))
	}
	return nil
}

// CLI opens an interactive Solace CLI session inside the container.
func (m *Manager) CLI(ctx context.Context) error {
	r, err := m.runtime()
	if err != nil {
		return err
	}
	return m.R.RunInteractive(ctx, r.Name(), r.Args("exec", "-it", m.name(), "cli", "-A")...)
}

// Shell opens an interactive shell inside the container.
func (m *Manager) Shell(ctx context.Context) error {
	r, err := m.runtime()
	if err != nil {
		return err
	}
	return m.R.RunInteractive(ctx, r.Name(), r.Args("exec", "-it", m.name(), "bash")...)
}

// --- helpers ----------------------------------------------------------------

// systemctl runs `systemctl [--user] args...` through the Runner, honoring the
// rootless (`--user`) vs rootful mode derived in config.
func (m *Manager) systemctl(ctx context.Context, args ...string) error {
	return m.R.Run(ctx, "systemctl", m.systemctlArgs(args)...)
}

// systemctlOutput captures a systemctl subcommand's stdout, for the state probes
// (`is-active`) whose answer picks the next step rather than being shown.
func (m *Manager) systemctlOutput(ctx context.Context, args ...string) ([]byte, error) {
	return m.R.Output(ctx, "systemctl", m.systemctlArgs(args)...)
}

// systemctlArgs prepends the `--user` token when the config is rootless, honoring
// the rootless vs rootful mode derived in config.
func (m *Manager) systemctlArgs(args []string) []string {
	if u := m.Cfg.Podman.SystemctlUser; u != "" {
		return append([]string{u}, args...)
	}
	return args
}

// checkPodmanEUID enforces the rootless/rootful invariant before a real deploy:
// rootless must not run as root and rootful must. It is skipped under the Echo runner
// and on platforms without a meaningful euid (Windows returns -1), so the deploy
// stays previewable everywhere.
func (m *Manager) checkPodmanEUID() error {
	if m.isEcho() {
		return nil
	}
	euid := m.Geteuid()
	if euid < 0 {
		return nil
	}
	if m.Cfg.Podman.Rootless && euid == 0 {
		return fmt.Errorf("podman.rootless=true but running as root; run as the target rootless user (no sudo)")
	}
	if !m.Cfg.Podman.Rootless && euid != 0 {
		return fmt.Errorf("rootful podman requires root; re-run with sudo or set podman.rootless=true")
	}
	return nil
}

// composeFile is the docker compose path, defaulting an empty config value to
// docker-compose.yml (config leaves it empty; the manager owns the default).
func (m *Manager) composeFile() string {
	if f := m.Cfg.Docker.ComposeFile; f != "" {
		return f
	}
	return "docker-compose.yml"
}

// fileExists reports whether p exists (used to pick compose-down vs stop/rm).
func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// exactName turns a container name into a `ps --filter name=` value that matches it
// and nothing else. The filter value is a regex, so a bare name is a substring match
// (a `solace` deployment would list a sibling `solace-edge`) and an unescaped '.' --
// which the schema allows in a container name -- would match any character. Used for
// the display listings only; containerRunning compares names in Go instead, because
// a decision must not rest on an engine's regex handling.
func exactName(name string) string { return "^" + regexp.QuoteMeta(name) + "$" }

// platformTitle is the display name for the platform in Check output/errors.
func platformTitle(p config.Platform) string {
	if p == config.Podman {
		return "Podman"
	}
	return "Docker"
}

// orNone renders an empty string as "(none)" -- a package-local copy of the k8s
// report formatter (one small copy per package, not over-DRY'd across packages).
func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

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
		parts = append(parts, s.Name+"="+setOrMissing(s.Value))
	}
	return fmt.Sprintf("%s (%s)", strings.Join(parts, " "), store)
}

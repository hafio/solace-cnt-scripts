package container

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/containercli"
	"github.com/hafio/go-deploy-utility/runner"

	"solace/internal/config"
	"solace/internal/render"
)

// These tests exercise the container host Manager over two runners: runner.Echo
// (dry-run) to assert the echoed command sequence portably (no euid/file/DNS side
// effects), and the capturing capRunner from transport_test.go to inspect exact
// argv and the real file-write paths that dry-run skips. capRunner/capCall/eqArgs
// live in transport_test.go (same package).

// ctrCfg builds a container config for platform p with a full node table. TLS is
// left unset (Check's "(not configured)" branch); tests that need it set it.
func ctrCfg(p config.Platform, redundancy string) *config.Config {
	c := &config.Config{}
	c.Image.Repo = "solace/solace-pubsub-standard"
	c.Image.Tag = "latest"
	c.SEMP.AdminPass = "secret-pass"
	c.Redundancy = config.Redundancy{
		Enabled: redundancy,
		Primary: config.Node{Name: "pri-host", Addr: "10.0.0.1"},
		Backup:  config.Node{Name: "bkp-host", Addr: "10.0.0.2"},
		Monitor: config.Node{Name: "mon-host", Addr: "10.0.0.3"},
	}
	// The container default scaling tier and the CPU it fixes. Scaling.CPU is
	// derived by ApplyDefaults, which this hand-built config deliberately skips
	// (executors must work without config.Load), so it is set alongside the tier
	// it comes from -- otherwise the rendered artifacts would carry no resource
	// caps and these tests would exercise the renderers' fail-safe branch instead
	// of the deploy path a real run takes.
	c.Scaling.MaxConnections = 1000
	c.Scaling.CPU = "2"
	switch p {
	case config.Podman:
		c.Podman.Command = config.Command{"podman"}
		// Rootful, as the tests below were written against: the schema default is
		// rootless now, so a test that wants it sets Rootful = false.
		c.Podman.Rootful = true
		c.Podman.Container.Name = "sol-pod"
		c.Podman.Container.RunUser = "1000:0" // config.RootlessRunUser, which ctrCfg skips deriving
		c.Podman.Container.Mem = "6898m"
		c.Podman.Container.DataDir = "/opt/solace/data"
		c.Podman.QuadletDir = "/etc/containers/systemd"
		// No podman.baseDir: it is optional, and deploy and remove DELETE the legacy
		// bundle file under it (removeLegacyCertBundle) -- so a fixed host path here
		// would aim every capRunner deploy at a real file. Tests of that clean-up set
		// it to a t.TempDir().
		c.Podman.Network.Mode = "host"
	default:
		c.Docker.Command = config.Command{"docker"}
		c.Docker.Container.Name = "solace"
		c.Docker.Container.RunUser = "0:0"
		c.Docker.Container.Mem = "6898m"
		c.Docker.Container.DataDir = "/opt/solace/data"
		c.Docker.Network.Mode = "host"
	}
	return c
}

// newEchoMgr builds a Manager over runner.Echo, routing echoed commands, report
// output, and progress to one buffer. Resolve is stubbed (dry-run skips it).
func newEchoMgr(cfg *config.Config, p config.Platform) (*Manager, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	m := NewManager(runner.Echo{W: buf}, cfg, p, func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }, buf)
	m.Resolve = func(string) bool { return true }
	return m, buf
}

// newCapMgr builds a Manager over the capturing capRunner (non-dry): file writes,
// euid checks, and DNS run for real, so Resolve defaults to always-resolves.
func newCapMgr(cfg *config.Config, p config.Platform) (*Manager, *capRunner, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	rr := &capRunner{}
	m := NewManager(rr, cfg, p, func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }, buf)
	m.Resolve = func(string) bool { return true }
	// Never the real process environment: ensureUserSession WRITES XDG_RUNTIME_DIR
	// and DBUS_SESSION_BUS_ADDRESS, and a test that mutated the test binary's own
	// environment would leak into every test after it. fakeEnv (rootless_test.go)
	// replaces this with a map a test can seed and inspect.
	m.Getenv = func(string) string { return "" }
	m.Setenv = func(string, string) error { return nil }
	return m, rr, buf
}

// callIndex is hasCall's answer with the position, for the cases where the ORDER
// of two calls is the property under test. -1 when the call is absent.
func callIndex(rr *capRunner, name string, args []string) int {
	for i, c := range rr.calls {
		if c.name == name && eqArgs(c.args, args) {
			return i
		}
	}
	return -1
}

// hasCall reports whether rr captured a Run/RunInput call to name with exactly args.
func hasCall(rr *capRunner, name string, args []string) bool {
	for _, c := range rr.calls {
		if c.name == name && eqArgs(c.args, args) {
			return true
		}
	}
	return false
}

// assertMode checks the permission bits of a file the manager wrote. Skipped on
// Windows, whose filesystem carries no POSIX mode -- the check matters on the
// hosts that actually run a broker.
func assertMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	if runtime.GOOS == "windows" {
		return
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Errorf("%s mode = %#o, want %#o", path, got, want)
	}
}

// withUser prepends the systemctl --user token when the config is rootless.
func withUser(cfg *config.Config, args ...string) []string {
	if cfg.Podman.SystemctlUser != "" {
		return append([]string{cfg.Podman.SystemctlUser}, args...)
	}
	return args
}

// --- Check ------------------------------------------------------------------

// --- redeploy fixtures --------------------------------------------------------

// psRow is a `ps` listing line for a running container of this deployment's image,
// in the three-field shape containercli.List asks for.
func psRow(name string) string {
	return name + "\tsolace/solace-pubsub-standard:latest\tUp 1 hour\n"
}

// dockerHost makes a capRunner answer a docker deploy's two probes: `ps` lists the
// container when running is set, and `container inspect` reports it with label as its
// deploy digest (none when label is empty). Everything else answers nothing.
func dockerHost(rr *capRunner, running bool, label string) {
	rr.outFor = func(_ string, args []string) []byte {
		switch {
		case slices.Contains(args, "inspect"):
			labels := ""
			if label != "" {
				labels = `"` + render.DeployDigestLabel + `": "` + label + `"`
			}
			return []byte(`[{"Name": "/solace", "State": {"Status": "running", "Running": true},` +
				` "Config": {"Image": "solace/solace-pubsub-standard:latest", "Labels": {` + labels + `}}}]`)
		case slices.Contains(args, "ps") && running:
			return []byte(psRow("solace"))
		}
		return nil
	}
}

// deployDigest is the digest a docker deploy of cfg sets in the compose child: the
// rendered file and every secret value, which is what the running container's label is
// compared with.
func deployDigest(t *testing.T, cfg *config.Config) string {
	t.Helper()
	m, _, _ := newCapMgr(cfg, config.Docker)
	env, err := m.composeSecretEnv(false)
	if err != nil {
		t.Fatalf("composeSecretEnv: %v", err)
	}
	return containercli.ComposeDigest(render.Compose(cfg, cfg.ResolveNode(config.Primary)), env)
}

// composeUp is the captured `compose ... up` call, failing the test when there is none.
func composeUp(t *testing.T, rr *capRunner) capCall {
	t.Helper()
	for _, c := range rr.calls {
		if slices.Contains(c.args, "up") {
			return c
		}
	}
	t.Fatalf("no `compose up` was run:\n%+v", rr.calls)
	return capCall{}
}

// composeUpArgs is the argv of a compose deploy: the file on stdin, no project
// directory of its own (a Config built in code has none, so it is the working directory),
// and extra after up -d.
func composeUpArgs(extra ...string) []string {
	return append([]string{"compose", "-f", "-", "--project-directory", ".", "up", "-d"}, extra...)
}

// podmanHost makes a capRunner answer a podman deploy's probes: `systemctl is-active`
// says state, `podman version` says 5.2.0 (so `secret create --replace` and `secret inspect
// --showsecret` are both available), and `secret inspect --showsecret` answers the value
// held in the store for the secret it names, newline-terminated as podman prints it.
// A secret not in values answers an empty line, which is no match.
func podmanHost(rr *capRunner, state string, values map[string]string) {
	podmanHostVersion(rr, state, "5.2.0\n", values)
}

// podmanHostVersion is podmanHost for a chosen `podman version` answer.
func podmanHostVersion(rr *capRunner, state, version string, values map[string]string) {
	rr.outFor = func(name string, args []string) []byte {
		switch {
		case name == "systemctl" && slices.Contains(args, "is-active"):
			return []byte(state + "\n")
		case name == "podman" && slices.Contains(args, "version"):
			return []byte(version)
		case name == "podman" && slices.Contains(args, "--showsecret"):
			return []byte(values[args[len(args)-1]] + "\n")
		}
		return nil
	}
}

// secretCreateArgs is the `secret create` argv for one secret: no label, since a label
// would put something derived from the value on the command line.
func secretCreateArgs(name string, replace bool) []string {
	args := []string{"secret", "create"}
	if replace {
		args = append(args, "--replace")
	}
	return append(args, name, "-")
}

// secretDataArgs is the read-back probe's argv for one secret.
func secretDataArgs(name string) []string {
	return []string{"secret", "inspect", "--showsecret", "--format", "{{.SecretData}}", name}
}

// wroteSecret reports whether any `secret create` for name was run.
func wroteSecret(rr *capRunner, name string) bool {
	for _, c := range rr.calls {
		if c.name == "podman" && slices.Contains(c.args, "create") && slices.Contains(c.args, name) {
			return true
		}
	}
	return false
}

// neverAsk is a Confirm that fails the test when it is asked.
func neverAsk(t *testing.T) func(string) (bool, error) {
	return func(q string) (bool, error) {
		t.Helper()
		t.Errorf("the deploy asked %q, want no question", q)
		return false, nil
	}
}

// answer is a Confirm that records the question and answers yes or no.
func answer(asked *string, yes bool) func(string) (bool, error) {
	return func(q string) (bool, error) { *asked = q; return yes, nil }
}

// errNoTerminal stands in for the gate's refusal when there is no terminal.
var errNoTerminal = errors.New("confirming the redeploy interactively: stdin is not a terminal; pass --no-prompt instead")

func TestManagerCheckDryRun(t *testing.T) {
	cases := []struct {
		p          config.Platform
		redundancy string
		mode       string
	}{
		{config.Docker, "true", "HA redundancy group"},
		{config.Docker, "false", "standalone (single broker)"},
		{config.Podman, "true", "HA redundancy group"},
		{config.Podman, "false", "standalone (single broker)"},
	}
	for _, tc := range cases {
		cfg := ctrCfg(tc.p, tc.redundancy)
		if tc.redundancy == "true" {
			cfg.TLS.Cert, cfg.TLS.CertKey = "server.pem", "server.key" // exercise the tls-configured branch
		}
		m, buf := newEchoMgr(cfg, tc.p)
		if err := m.Check(context.Background()); err != nil {
			t.Fatalf("%s/%s Check: %v", tc.p, tc.redundancy, err)
		}
		out := buf.String()
		for _, want := range []string{
			"=== Broker deployment (" + string(tc.p) + ")",
			tc.mode,
			"+ " + cfg.ContainerRuntime(tc.p).String() + " version",
			"skipped (preview)",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("%s/%s Check missing %q:\n%s", tc.p, tc.redundancy, want, out)
			}
		}
	}
}

// TestCheckEnvReportsBaseDirOnlyWhenSet: podman.baseDir is optional and consulted
// only to delete a legacy key file, so the report names it when set, says that is all
// it does, and leaves it out when unset rather than printing an empty value.
func TestCheckEnvReportsBaseDirOnlyWhenSet(t *testing.T) {
	for _, tc := range []struct{ base, want, unwanted string }{
		{"/srv/solace-base", "baseDir=/srv/solace-base (legacy bundle clean-up only)", ""},
		{"", "rootful=true quadletDir=/etc/containers/systemd", "baseDir="},
	} {
		cfg := ctrCfg(config.Podman, "false")
		cfg.Podman.BaseDir = tc.base
		m, buf := newEchoMgr(cfg, config.Podman)
		m.CheckEnv()
		if !strings.Contains(buf.String(), tc.want) {
			t.Errorf("baseDir=%q: report missing %q:\n%s", tc.base, tc.want, buf.String())
		}
		if tc.unwanted != "" && strings.Contains(buf.String(), tc.unwanted) {
			t.Errorf("baseDir=%q: report should not carry %q:\n%s", tc.base, tc.unwanted, buf.String())
		}
	}
}

func TestManagerCheckDNSFailsLoudInHA(t *testing.T) {
	m, _, buf := newCapMgr(ctrCfg(config.Docker, "true"), config.Docker)
	m.Resolve = func(host string) bool { return host != "bkp-host" }
	if err := m.Check(context.Background()); err == nil {
		t.Fatal("Check must fail when a redundancy hostname does not resolve")
	}
	if !strings.Contains(buf.String(), "does NOT resolve: bkp-host") {
		t.Errorf("Check should name the failing host:\n%s", buf.String())
	}
}

func TestManagerCheckStandaloneDNSWarnsOnly(t *testing.T) {
	m, _, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	m.Resolve = func(string) bool { return false } // unresolved
	if err := m.Check(context.Background()); err != nil {
		t.Fatalf("standalone Check must not fail on an unresolved name: %v", err)
	}
}

// --- PrepHost ---------------------------------------------------------------

func TestManagerPrepHostRootlessUsesUnshareChown(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false") // standalone -> PSK step is skipped
	cfg.Podman.Rootful = false
	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return 1000 } // rootless as non-root: euid guard passes
	// PrepHost runs the whole podman host-readiness block first (rootless.go), so
	// every one of its probes has to answer before the chown is ever reached.
	rr.outFor = healthyRootlessOut(healthyNrOpen)
	if err := m.PrepHost(context.Background()); err != nil {
		t.Fatalf("PrepHost: %v", err)
	}
	if !hasCall(rr, "podman", []string{"unshare", "chown", "1000:0", "/opt/solace/data"}) {
		t.Errorf("rootless PrepHost should chown via `podman unshare`:\n%+v", rr.calls)
	}
	// Both in the namespace, chmod first: on a re-deploy the directory already
	// belongs to a subuid, so a host-side chmod would be refused outright.
	chmod, chown := callIndex(rr, "podman", []string{"unshare", "chmod", "775", "/opt/solace/data"}),
		callIndex(rr, "podman", []string{"unshare", "chown", "1000:0", "/opt/solace/data"})
	// A bare host-side chmod is the bug this replaced: on a re-deploy the directory
	// already belongs to the subuid and the caller is refused outright.
	if hasCall(rr, "chmod", []string{"775", "/opt/solace/data"}) {
		t.Errorf("the data-dir chmod must run in the namespace, not on the host:\n%+v", rr.calls)
	}
	if chmod < 0 {
		t.Errorf("rootless PrepHost should make the data dir group-writable:\n%+v", rr.calls)
	} else if chmod > chown {
		t.Errorf("chmod 775 must come before the chown, got chmod at %d and chown at %d:\n%+v",
			chmod, chown, rr.calls)
	}
}

// TestManagerPrepHostLeavesBaseDirAlone: prep used to create podman.baseDir 0700
// for the server-certificate bundle. The certificate rides podman's secret store
// now, so a configured baseDir must not make prep create, chmod or chown anything
// there -- a directory made for a private key that never arrives is just clutter.
func TestManagerPrepHostLeavesBaseDirAlone(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.Rootful = false
	cfg.Podman.BaseDir = "/srv/solace-base"
	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return 1000 }
	rr.outFor = healthyRootlessOut(healthyNrOpen)
	if err := m.PrepHost(context.Background()); err != nil {
		t.Fatalf("PrepHost: %v", err)
	}
	for _, c := range rr.calls {
		if slices.Contains(c.args, cfg.Podman.BaseDir) {
			t.Errorf("prep must not touch podman.baseDir any more: %s %v", c.name, c.args)
		}
	}
}

// TestManagerPrepHostDirectoryErrors covers the data directory's mode failing under
// rootless: it names the directory and what it was trying to do, which an engine's
// bare "permission denied" says neither of.
func TestManagerPrepHostDirectoryErrors(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.Rootful = false
	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return 1000 }
	rr.outFor = healthyRootlessOut(healthyNrOpen)
	rr.fail = failOn("775")
	err := m.PrepHost(context.Background())
	if err == nil {
		t.Fatal("a failing data dir chmod must stop prep")
	}
	if !strings.Contains(err.Error(), "group-writable") {
		t.Errorf("error should say what failed, got: %v", err)
	}
}

// TestPrepHostRootlessDryRunSkipsTheReadinessBlock: the nofile probe is one row of
// the podman host-readiness block now (rootless.go), and the block skips as a whole
// under the Echo runner rather than row by row. Nothing it asserts can be answered
// without a real host, so one honest skip beats seven echoed probes and seven
// skipped assertions -- the same shape Preflight uses, one level up.
func TestPrepHostRootlessDryRunSkipsTheReadinessBlock(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.Rootful = false
	m, buf := newEchoMgr(cfg, config.Podman)
	if err := m.PrepHost(context.Background()); err != nil {
		t.Fatalf("PrepHost: %v", err)
	}
	out := buf.String()
	// Two blocks skip now, the readiness rows and the host limits (limits.go).
	for _, want := range []string{"podman host", "limits", "skipped (preview)"} {
		if !strings.Contains(out, want) {
			t.Errorf("dry-run should say %q was skipped:\n%s", want, out)
		}
	}
	// And prep still previews the work it would do after the block passes.
	if !strings.Contains(out, "mkdir") {
		t.Errorf("dry-run should still echo the prep commands:\n%s", out)
	}
}

// --- Deploy -----------------------------------------------------------------

// TestManagerDeployDockerFeedsComposeOnStdin: the compose file goes to compose on stdin
// (`-f -`), anchored by --project-directory, and is written nowhere. The values and the
// deploy digest ride the child's environment, and nothing but the file is on stdin.
func TestManagerDeployDockerFeedsComposeOnStdin(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	up := composeUp(t, rr)
	if up.name != "docker" || !eqArgs(up.args, composeUpArgs()) {
		t.Errorf("compose argv = %s %v, want docker %v", up.name, up.args, composeUpArgs())
	}
	if want := string(render.Compose(cfg, cfg.ResolveNode(config.Primary))); up.stdin != want {
		t.Errorf("compose stdin is not the rendered compose file:\n%s", up.stdin)
	}
	if up.method != "RunEnv" {
		t.Errorf("compose must run with an environment so the values never reach an argv, got %s", up.method)
	}
	if want := render.DeployDigestEnvVar + "=" + deployDigest(t, cfg); !containsStr(up.env, want) {
		t.Errorf("the child environment must carry the deploy digest %s, got %v", want, maskedKeys(up.env))
	}
}

// TestManagerDockerComposeCommandOverride covers the standalone-binary form: a
// host without the compose plugin sets docker.compose, and every compose call has
// to go through it rather than the runtime.
func TestManagerDockerComposeCommandOverride(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.Compose = config.Command{"docker-compose"}
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	want := composeUpArgs()[1:] // docker-compose is its own program: no leading `compose`
	if !hasCall(rr, "docker-compose", want) {
		t.Errorf("Deploy should use the configured compose command %v:\n%+v", want, rr.calls)
	}
}

// TestManagerDockerCheckProbesCompose pins the compose probe: the plugin is a
// separate install from the engine, so a reachable docker with no compose must
// fail at check time rather than at deploy time.
func TestManagerDockerCheckProbesCompose(t *testing.T) {
	m, buf := newEchoMgr(ctrCfg(config.Docker, "false"), config.Docker)
	if err := m.Check(context.Background()); err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !strings.Contains(buf.String(), "+ docker compose version") {
		t.Errorf("docker Check should probe the compose command:\n%s", buf.String())
	}
}

func TestManagerDockerCheckFailsWhenComposeMissing(t *testing.T) {
	m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	// Fail only the compose probe, so the engine still looks reachable and the
	// error has to be the compose-specific one.
	rr.outFail = failOn("compose")
	err := m.Check(context.Background())
	if err == nil {
		t.Fatal("Check must fail when the compose command cannot run")
	}
	if !strings.Contains(err.Error(), "docker.compose") {
		t.Errorf("error should point at the docker.compose override, got: %v", err)
	}
}

// --- describe / copy parity -------------------------------------------------

// TestManagerDescribe covers the container analog of `kubectl describe pod`:
// `<runtime> inspect` on both platforms, plus the installed unit on podman (the
// unit's summary is already in Status, so `cat` answers a different question).
func TestManagerDescribe(t *testing.T) {
	t.Run("docker", func(t *testing.T) {
		m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
		if err := m.Describe(context.Background()); err != nil {
			t.Fatalf("Describe: %v", err)
		}
		if !hasCall(rr, "docker", []string{"container", "inspect", "solace"}) {
			t.Errorf("Describe should inspect the container:\n%+v", rr.calls)
		}
	})
	t.Run("podman also shows the unit", func(t *testing.T) {
		cfg := ctrCfg(config.Podman, "false")
		m, rr, _ := newCapMgr(cfg, config.Podman)
		if err := m.Describe(context.Background()); err != nil {
			t.Fatalf("Describe: %v", err)
		}
		if !hasCall(rr, "systemctl", withUser(cfg, "cat", "sol-pod.service")) {
			t.Errorf("podman Describe should show the installed unit:\n%+v", rr.calls)
		}
		if !hasCall(rr, "podman", []string{"container", "inspect", "sol-pod"}) {
			t.Errorf("podman Describe should inspect the container:\n%+v", rr.calls)
		}
	})
	t.Run("a missing unit is tolerated", func(t *testing.T) {
		m, rr, buf := newCapMgr(ctrCfg(config.Podman, "false"), config.Podman)
		rr.fail = failOn("cat")
		if err := m.Describe(context.Background()); err != nil {
			t.Fatalf("Describe should tolerate an uninstalled unit: %v", err)
		}
		if !strings.Contains(buf.String(), "systemctl cat") {
			t.Errorf("Describe should warn about the missing unit:\n%s", buf.String())
		}
	})
}

// TestManagerCopy covers the copy verbs the container tree previously lacked. The
// per-file reporting and the non-zero exit on any failure mirror the k8s verbs.
func TestManagerCopy(t *testing.T) {
	t.Run("from", func(t *testing.T) {
		m, rr, buf := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
		if err := m.CopyFrom(context.Background(), []string{"/var/lib/solace/jail/logs/debug.log"}); err != nil {
			t.Fatalf("CopyFrom: %v", err)
		}
		if !hasCall(rr, "docker", []string{"cp", "solace:/var/lib/solace/jail/logs/debug.log", "debug.log"}) {
			t.Errorf("CopyFrom should cp out of the container:\n%+v", rr.calls)
		}
		if !strings.Contains(buf.String(), "[ OK ]") {
			t.Errorf("CopyFrom should report each file:\n%s", buf.String())
		}
	})
	t.Run("into", func(t *testing.T) {
		m, rr, _ := newCapMgr(ctrCfg(config.Podman, "false"), config.Podman)
		if err := m.CopyInto(context.Background(), []string{"setup.cli"}, "/tmp"); err != nil {
			t.Fatalf("CopyInto: %v", err)
		}
		if !hasCall(rr, "podman", []string{"cp", "setup.cli", "sol-pod:/tmp"}) {
			t.Errorf("CopyInto should cp into the container:\n%+v", rr.calls)
		}
	})
	t.Run("no files is an error", func(t *testing.T) {
		m, _, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
		if err := m.CopyFrom(context.Background(), nil); err == nil {
			t.Error("CopyFrom with no files should error")
		}
		if err := m.CopyInto(context.Background(), nil, ""); err == nil {
			t.Error("CopyInto with no files should error")
		}
	})
	t.Run("a failed file makes the command fail", func(t *testing.T) {
		m, rr, buf := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
		rr.fail = failOn("cp")
		if err := m.CopyFrom(context.Background(), []string{"a.log", "b.log"}); err == nil {
			t.Error("CopyFrom should fail when a file could not be copied")
		}
		if !strings.Contains(buf.String(), "[FAIL]") {
			t.Errorf("CopyFrom should report the failing file:\n%s", buf.String())
		}
	})
}

// --- registry login ---------------------------------------------------------

// TestManagerPrepHostRegistryLogin pins the credentials that used to be silently
// ignored on containers: prep now logs in, with the password on stdin so it never
// reaches an argv or the dry-run echo.
func TestManagerPrepHostRegistryLogin(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	cfg.Image.Registry = "registry.example.com"
	cfg.Image.User = "repo-user"
	cfg.Image.Pass = "repo-pass"
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.PrepHost(context.Background()); err != nil {
		t.Fatalf("PrepHost: %v", err)
	}
	want := []string{"login", "--username", "repo-user", "--password-stdin", "registry.example.com"}
	if !hasCall(rr, "docker", want) {
		t.Errorf("PrepHost should log in to the registry:\n%+v", rr.calls)
	}
	for _, c := range rr.calls {
		for _, a := range c.args {
			if strings.Contains(a, "repo-pass") {
				t.Errorf("the registry password must not reach an argv: %s %v", c.name, c.args)
			}
		}
	}
}

func TestManagerPrepHostNoLoginWithoutCreds(t *testing.T) {
	m, rr, _ := newCapMgr(ctrCfg(config.Podman, "false"), config.Podman)
	m.Geteuid = func() int { return 0 } // rootful (the ctrCfg default) requires root
	if err := m.PrepHost(context.Background()); err != nil {
		t.Fatalf("PrepHost: %v", err)
	}
	for _, c := range rr.calls {
		if len(c.args) > 0 && c.args[0] == "login" {
			t.Errorf("no credentials means no login attempt:\n%+v", rr.calls)
		}
	}
}

func TestManagerPrepHostRejectsHalfCredentials(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	cfg.Image.User = "repo-user" // pass left empty
	m, _, _ := newCapMgr(cfg, config.Docker)
	err := m.PrepHost(context.Background())
	if err == nil || !strings.Contains(err.Error(), "image.user and image.pass") {
		t.Fatalf("PrepHost err = %v, want the both-or-neither credentials error", err)
	}
}

// --- redeploy: compare, then restart only with consent ----------------------

// --- redeploy: compare, ask, then write -------------------------------------

// TestManagerRedeployUnchangedIsNoOp covers the "nothing changed" arm on both platforms,
// and what it no longer is: docker runs `up -d` and leaves the decision to compose (it
// asks nothing, and with the digest label unchanged compose recreates nothing), while
// podman writes and restarts nothing at all and asks nothing. A re-deploy that renders the
// same artifact against a running broker must not touch it.
func TestManagerRedeployUnchangedIsNoOp(t *testing.T) {
	t.Run("podman", func(t *testing.T) {
		dir := t.TempDir()
		cfg := ctrCfg(config.Podman, "false")
		cfg.Podman.QuadletDir = dir
		unit := filepath.Join(dir, "sol-pod.container")
		if err := os.WriteFile(unit, render.Quadlet(cfg, cfg.ResolveNode(config.Primary)), 0o600); err != nil {
			t.Fatal(err)
		}
		m, rr, buf := newCapMgr(cfg, config.Podman)
		m.Geteuid = func() int { return -1 }
		m.Confirm = neverAsk(t)
		// The unit is running, and the store already holds this password.
		podmanHost(rr, "active", map[string]string{"sol-pod-admin-password": "secret-pass"})
		before, _ := os.Stat(unit)
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		for _, c := range rr.calls {
			if c.name == "systemctl" && (slices.Contains(c.args, "restart") || slices.Contains(c.args, "start") ||
				slices.Contains(c.args, "daemon-reload")) {
				t.Errorf("an unchanged deploy must not reload, start or bounce the broker: %s %v", c.name, c.args)
			}
			if c.name == "podman" && (slices.Contains(c.args, "create") || slices.Contains(c.args, "rm")) {
				t.Errorf("an unchanged secret must not be rewritten: %s %v", c.name, c.args)
			}
		}
		if after, _ := os.Stat(unit); !after.ModTime().Equal(before.ModTime()) {
			t.Error("an unchanged unit must not be rewritten")
		}
		out := buf.String()
		if !strings.Contains(out, "unchanged, left running") || !strings.Contains(out, "--restart") {
			t.Errorf("Deploy should say it left the service alone and how to restart it anyway:\n%s", out)
		}
	})
	t.Run("docker", func(t *testing.T) {
		cfg := ctrCfg(config.Docker, "false")
		m, rr, buf := newCapMgr(cfg, config.Docker)
		m.Confirm = neverAsk(t)
		dockerHost(rr, true, deployDigest(t, cfg)) // running, created from this very deploy
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		// Compose decides, so `up -d` still runs -- and never with --force-recreate.
		if up := composeUp(t, rr); !eqArgs(up.args, composeUpArgs()) {
			t.Errorf("an unchanged running container is left to a plain `up -d`, got %v", up.args)
		}
		if !strings.Contains(buf.String(), "unchanged, left running") {
			t.Errorf("Deploy should say it left the container alone:\n%s", buf.String())
		}
	})
}

// TestManagerRedeployChangedNeedsConsent is the core of the redeploy rule: a changed
// artifact or secret against a RUNNING broker is not applied until the operator agrees,
// and a "no" leaves the host exactly as it was. Podman's old behaviour was worse than
// silent -- `systemctl start` on an active unit is a no-op, so the broker kept the old
// image while the command reported success.
func TestManagerRedeployChangedNeedsConsent(t *testing.T) {
	// podman: the unit on disk is from an older image tag, so the unit changed.
	podman := func(t *testing.T) (cfg *config.Config, unit string, old []byte) {
		t.Helper()
		dir := t.TempDir()
		cfg = ctrCfg(config.Podman, "false")
		cfg.Podman.QuadletDir = dir
		older := ctrCfg(config.Podman, "false")
		older.Image.Tag = "previous"
		old = render.Quadlet(older, older.ResolveNode(config.Primary))
		unit = filepath.Join(dir, "sol-pod.container")
		if err := os.WriteFile(unit, old, 0o600); err != nil {
			t.Fatal(err)
		}
		return cfg, unit, old
	}
	podmanMgr := func(t *testing.T, cfg *config.Config) (*Manager, *capRunner, *bytes.Buffer) {
		m, rr, buf := newCapMgr(cfg, config.Podman)
		m.Geteuid = func() int { return -1 }
		podmanHost(rr, "active", map[string]string{"sol-pod-admin-password": "secret-pass"})
		return m, rr, buf
	}
	// nothingDone asserts a declined podman deploy changed nothing.
	nothingDone := func(t *testing.T, rr *capRunner, unit string, old []byte) {
		t.Helper()
		if got, _ := os.ReadFile(unit); !bytes.Equal(got, old) {
			t.Error("a declined deploy must leave the unit on disk as it was")
		}
		for _, c := range rr.calls {
			if c.name == "systemctl" && (slices.Contains(c.args, "restart") || slices.Contains(c.args, "daemon-reload")) {
				t.Errorf("a declined deploy must not reload or restart: %s %v", c.name, c.args)
			}
			if c.name == "podman" && (slices.Contains(c.args, "create") || slices.Contains(c.args, "rm")) {
				t.Errorf("a declined deploy must not touch the secret store: %s %v", c.name, c.args)
			}
		}
	}

	t.Run("podman declines when the answer is no", func(t *testing.T) {
		cfg, unit, old := podman(t)
		m, rr, buf := podmanMgr(t, cfg)
		var asked string
		m.Confirm = answer(&asked, false)
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v (a \"no\" is a decision, not a failure)", err)
		}
		nothingDone(t, rr, unit, old)
		if !strings.Contains(buf.String(), "cancelled; nothing was changed") {
			t.Errorf("Deploy should say nothing was changed:\n%s", buf.String())
		}
		if !strings.Contains(asked, "sol-pod.service") || !strings.Contains(asked, "quadlet unit") {
			t.Errorf("the question should name the service and what changed, got %q", asked)
		}
	})
	t.Run("podman with no Confirm seam declines", func(t *testing.T) {
		cfg, unit, old := podman(t)
		m, rr, buf := podmanMgr(t, cfg)
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		nothingDone(t, rr, unit, old)
		if !strings.Contains(buf.String(), "cancelled") {
			t.Errorf("a Manager with no seam declines, and says so:\n%s", buf.String())
		}
	})
	t.Run("podman fails, writing nothing, when the question cannot be asked", func(t *testing.T) {
		cfg, unit, old := podman(t)
		m, rr, _ := podmanMgr(t, cfg)
		m.Confirm = func(string) (bool, error) { return false, errNoTerminal }
		err := m.Deploy(context.Background(), config.Primary)
		if !errors.Is(err, errNoTerminal) {
			t.Fatalf("Deploy err = %v, want the gate's refusal", err)
		}
		nothingDone(t, rr, unit, old)
	})
	t.Run("podman restarts with --restart and asks nothing", func(t *testing.T) {
		cfg, unit, old := podman(t)
		m, rr, _ := podmanMgr(t, cfg)
		m.RestartApproved = true
		m.Confirm = neverAsk(t)
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		if !hasCall(rr, "systemctl", withUser(cfg, "restart", "sol-pod.service")) {
			t.Errorf("--restart should restart the active unit:\n%+v", rr.calls)
		}
		if got, _ := os.ReadFile(unit); bytes.Equal(got, old) {
			t.Error("--restart applies the change, so the unit must be rewritten")
		}
	})
	t.Run("podman writes, reloads and restarts when the prompt is accepted", func(t *testing.T) {
		cfg, unit, old := podman(t)
		m, rr, _ := podmanMgr(t, cfg)
		var asked string
		m.Confirm = answer(&asked, true)
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		if got, _ := os.ReadFile(unit); bytes.Equal(got, old) {
			t.Error("an accepted prompt should rewrite the unit")
		}
		reload := callIndex(rr, "systemctl", withUser(cfg, "daemon-reload"))
		restart := callIndex(rr, "systemctl", withUser(cfg, "restart", "sol-pod.service"))
		if reload < 0 || restart < reload {
			t.Errorf("want daemon-reload (%d) then restart (%d):\n%+v", reload, restart, rr.calls)
		}
		if wroteSecret(rr, "sol-pod-admin-password") {
			t.Errorf("the password did not change, so its secret is not rewritten:\n%+v", rr.calls)
		}
	})
	t.Run("podman starts a stopped unit without asking", func(t *testing.T) {
		cfg, _, _ := podman(t)
		m, rr, _ := podmanMgr(t, cfg)
		podmanHost(rr, "inactive", nil)
		m.Confirm = neverAsk(t)
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		if !hasCall(rr, "systemctl", withUser(cfg, "start", "sol-pod.service")) {
			t.Errorf("a stopped unit is started, with nothing to ask:\n%+v", rr.calls)
		}
	})

	// docker: the running container's label is from some other deploy.
	dockerMgr := func(t *testing.T, label string) (*config.Config, *Manager, *capRunner, *bytes.Buffer) {
		cfg := ctrCfg(config.Docker, "false")
		m, rr, buf := newCapMgr(cfg, config.Docker)
		dockerHost(rr, true, label)
		return cfg, m, rr, buf
	}
	noUp := func(t *testing.T, rr *capRunner) {
		t.Helper()
		for _, c := range rr.calls {
			if slices.Contains(c.args, "up") || slices.Contains(c.args, "down") {
				t.Errorf("a declined deploy must run no compose command: %s %v", c.name, c.args)
			}
		}
	}

	t.Run("docker declines when the answer is no", func(t *testing.T) {
		_, m, rr, buf := dockerMgr(t, "an-older-deploy")
		var asked string
		m.Confirm = answer(&asked, false)
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v (a \"no\" is a decision, not a failure)", err)
		}
		noUp(t, rr)
		if !strings.Contains(buf.String(), "cancelled; nothing was changed") {
			t.Errorf("Deploy should say nothing was changed:\n%s", buf.String())
		}
		if !strings.Contains(asked, "solace") {
			t.Errorf("the question should name the container, got %q", asked)
		}
	})
	t.Run("docker with no Confirm seam declines", func(t *testing.T) {
		_, m, rr, _ := dockerMgr(t, "an-older-deploy")
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		noUp(t, rr)
	})
	t.Run("docker fails, running nothing, when the question cannot be asked", func(t *testing.T) {
		_, m, rr, _ := dockerMgr(t, "an-older-deploy")
		m.Confirm = func(string) (bool, error) { return false, errNoTerminal }
		if err := m.Deploy(context.Background(), config.Primary); !errors.Is(err, errNoTerminal) {
			t.Fatalf("Deploy err = %v, want the gate's refusal", err)
		}
		noUp(t, rr)
	})
	t.Run("docker applies the change with --restart and asks nothing", func(t *testing.T) {
		_, m, rr, _ := dockerMgr(t, "an-older-deploy")
		m.RestartApproved = true
		m.Confirm = neverAsk(t)
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		// The changed label already makes compose recreate; --force-recreate is only for
		// an UNCHANGED container (TestManagerRedeployUnchangedRestartsForRotation).
		if up := composeUp(t, rr); !eqArgs(up.args, composeUpArgs()) {
			t.Errorf("a changed deploy is a plain `up -d`, got %v", up.args)
		}
	})
	t.Run("docker applies the change when the prompt is accepted", func(t *testing.T) {
		_, m, rr, _ := dockerMgr(t, "an-older-deploy")
		m.Confirm = answer(new(string), true)
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		if up := composeUp(t, rr); !eqArgs(up.args, composeUpArgs()) {
			t.Errorf("an accepted prompt runs `up -d`, got %v", up.args)
		}
	})
	t.Run("docker asks when the container has no digest label", func(t *testing.T) {
		// A container an earlier build created: it carries no deploy label, so it
		// cannot be shown unchanged, and costs at most a consent-gated recreate.
		_, m, rr, _ := dockerMgr(t, "")
		var asked string
		m.Confirm = answer(&asked, false)
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		if asked == "" {
			t.Error("a running container with no label must be asked about")
		}
		noUp(t, rr)
	})
	t.Run("docker asks when the container cannot be inspected", func(t *testing.T) {
		_, m, rr, _ := dockerMgr(t, "irrelevant")
		rr.outFail = func(_ string, args []string) error {
			if slices.Contains(args, "inspect") {
				return errors.New("no such container")
			}
			return nil
		}
		var asked string
		m.Confirm = answer(&asked, false)
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		if asked == "" {
			t.Error("an unreadable label counts as changed, so the deploy must ask")
		}
		noUp(t, rr)
	})
	t.Run("docker asks when the inspect answer is not a container list", func(t *testing.T) {
		_, m, rr, _ := dockerMgr(t, "irrelevant")
		rr.outFor = func(_ string, args []string) []byte {
			switch {
			case slices.Contains(args, "inspect"):
				return []byte("[]") // an empty array: the engine knows no such container
			case slices.Contains(args, "ps"):
				return []byte(psRow("solace"))
			}
			return nil
		}
		var asked string
		m.Confirm = answer(&asked, false)
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		if asked == "" {
			t.Error("no container in the inspect answer counts as changed, so the deploy must ask")
		}
	})
	t.Run("docker does not ask about a container that is not running", func(t *testing.T) {
		cfg := ctrCfg(config.Docker, "false")
		m, rr, _ := newCapMgr(cfg, config.Docker)
		dockerHost(rr, false, "an-older-deploy")
		m.Confirm = neverAsk(t)
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		if up := composeUp(t, rr); !eqArgs(up.args, composeUpArgs()) {
			t.Errorf("a stopped container is brought up with a plain `up -d`, got %v", up.args)
		}
	})
}

// TestManagerRedeployRotatedPasswordAsks is what the digest buys. Compose compares a
// service's configuration, not a secret's content, and the quadlet unit names its secrets
// rather than their values, so a rotated password leaves both artifacts byte-identical.
// The deploy must still notice it and ask, on both platforms: --restart used to be the
// only way to apply one.
func TestManagerRedeployRotatedPasswordAsks(t *testing.T) {
	t.Run("docker", func(t *testing.T) {
		old := ctrCfg(config.Docker, "false")
		old.SEMP.AdminPass = "old-pass"
		cfg := ctrCfg(config.Docker, "false") // the file is the same; only the password moved
		m, rr, _ := newCapMgr(cfg, config.Docker)
		dockerHost(rr, true, deployDigest(t, old))
		var asked string
		m.Confirm = answer(&asked, true)
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		if asked == "" {
			t.Fatal("a rotated password must be asked about")
		}
		up := composeUp(t, rr)
		if !containsStr(up.env, "SOLACE_ADMIN_PASSWORD=secret-pass") {
			t.Errorf("the new password must reach compose, got %v", maskedKeys(up.env))
		}
		// And the digest it carries differs from the one the container holds, which is
		// what makes compose recreate it.
		if containsStr(up.env, render.DeployDigestEnvVar+"="+deployDigest(t, old)) {
			t.Error("the deploy digest must change with the password")
		}
	})
	t.Run("podman", func(t *testing.T) {
		dir := t.TempDir()
		cfg := ctrCfg(config.Podman, "true")
		cfg.Redundancy.PSK = "test-psk"
		cfg.Podman.QuadletDir = dir
		if err := os.WriteFile(filepath.Join(dir, "sol-pod.container"),
			render.Quadlet(cfg, cfg.ResolveNode(config.Primary)), 0o600); err != nil {
			t.Fatal(err)
		}
		m, rr, _ := newCapMgr(cfg, config.Podman)
		m.Geteuid = func() int { return -1 }
		// The store holds the OLD admin password and the current PSK.
		podmanHost(rr, "active", map[string]string{
			"sol-pod-admin-password": "old-pass",
			"sol-pod-redundancy-psk": "test-psk",
		})
		var asked string
		m.Confirm = answer(&asked, true)
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		if !strings.Contains(asked, "secrets changed") {
			t.Errorf("the question should say the secrets changed, got %q", asked)
		}
		if !wroteSecret(rr, "sol-pod-admin-password") {
			t.Errorf("the rotated secret must be rewritten:\n%+v", rr.calls)
		}
		if wroteSecret(rr, "sol-pod-redundancy-psk") {
			t.Errorf("the unchanged secret must be skipped:\n%+v", rr.calls)
		}
		if !hasCall(rr, "systemctl", withUser(cfg, "restart", "sol-pod.service")) {
			t.Errorf("an accepted rotation restarts the service:\n%+v", rr.calls)
		}
		if !hasCall(rr, "systemctl", withUser(cfg, "daemon-reload")) {
			t.Errorf("a write is followed by a daemon-reload:\n%+v", rr.calls)
		}
	})
}

// TestManagerRedeployPodmanDeclinedRotationWritesNothing: a "no" to a rotated password
// leaves the store as it was, so the next deploy asks again.
func TestManagerRedeployPodmanDeclinedRotationWritesNothing(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = dir
	if err := os.WriteFile(filepath.Join(dir, "sol-pod.container"),
		render.Quadlet(cfg, cfg.ResolveNode(config.Primary)), 0o600); err != nil {
		t.Fatal(err)
	}
	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	podmanHost(rr, "active", map[string]string{"sol-pod-admin-password": "old-pass"})
	m.Confirm = answer(new(string), false)
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if wroteSecret(rr, "sol-pod-admin-password") {
		t.Errorf("a declined rotation must not write the secret:\n%+v", rr.calls)
	}
}

// --- secret externalization -------------------------------------------------

// TestManagerDeployDockerPassesSecretsAsEnv is the docker half of the secret
// model: nothing is written to this host, the values reach compose only through
// the child process environment, and the artifact (which is on stdin, never on disk)
// carries variable names.
func TestManagerDeployDockerPassesSecretsAsEnv(t *testing.T) {
	cfg := ctrCfg(config.Docker, "true")
	cfg.Redundancy.PSK = "test-psk"
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	up := composeUp(t, rr)
	if up.method != "RunEnv" {
		t.Errorf("compose must run through RunEnv so the values never reach an argv, got %s", up.method)
	}
	for _, want := range []string{"SOLACE_ADMIN_PASSWORD=secret-pass", "SOLACE_REDUNDANCY_PSK=test-psk"} {
		if !containsStr(up.env, want) {
			t.Errorf("compose environment should carry %q, got %v", want, maskedKeys(up.env))
		}
	}
	// The values may reach the environment and nothing else: not an argv, and not the
	// file on stdin.
	for _, c := range rr.calls {
		for _, a := range c.args {
			for _, leak := range []string{"secret-pass", "test-psk"} {
				if strings.Contains(a, leak) {
					t.Errorf("secret value reached an argv: %s %v", c.name, c.args)
				}
			}
		}
	}
	for _, leak := range []string{"secret-pass", "test-psk"} {
		if strings.Contains(up.stdin, leak) {
			t.Errorf("compose file must reference secrets, not carry %q:\n%s", leak, up.stdin)
		}
	}
	for _, want := range []string{
		// Quoted: the secret name derives from container.name, and a name YAML 1.1
		// reads as a boolean or number (`yes`, `off`, `0123`) is legal to both
		// engines, so every identifier scalar in the compose document is quoted.
		`source: "solace-admin-password"`,
		// The target is the ABSOLUTE in-container path, not a bare filename: compose
		// resolves a bare one under its own /run/secrets, where the broker never
		// looks (render.ContainerSecret.Target). It must match the *filepath value
		// on the line below, which is the whole point of the two being one
		// expression.
		"target: /mnt/secrets/username_admin_password",
		"environment: SOLACE_ADMIN_PASSWORD",
		"username_admin_passwordfilepath: \"/mnt/secrets/username_admin_password\"",
		// The label compose recreates on, filled from the child's environment.
		render.DeployDigestLabel + `: "${` + render.DeployDigestEnvVar + `:-}"`,
	} {
		if !strings.Contains(up.stdin, want) {
			t.Errorf("compose file should contain %q:\n%s", want, up.stdin)
		}
	}
}

// containsStr reports whether list holds want exactly.
func containsStr(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// maskedKeys renders an environment list for a failure message without its
// values -- a test diagnostic must not print a secret either.
func maskedKeys(env []string) string { return runner.MaskEnv(env) }

// versionAnswer makes the capRunner answer `podman version` with v and nothing else,
// which is all the secrets path asks of it.
func versionAnswer(rr *capRunner, v string) {
	rr.outFor = func(name string, args []string) []byte {
		if name == "podman" && len(args) > 0 && args[0] == "version" {
			return []byte(v)
		}
		return nil
	}
}

// TestManagerDeployPodmanCreatesSecrets covers both create paths, chosen by the
// podman client version (asked once per deploy): 4.7 and later replaces in one
// `secret create --replace`, anything older removes the secret and then creates it. Every
// secret is labelled with the digest of its value, which is how the next deploy tells a
// rotated one from an unchanged one.
func TestManagerDeployPodmanCreatesSecrets(t *testing.T) {
	values := map[string]string{
		"sol-pod-admin-password": "secret-pass",
		"sol-pod-redundancy-psk": "test-psk",
	}
	for _, tc := range []struct {
		name, version string
		replace       bool
	}{
		{"podman 4.6.1 removes then creates", "4.6.1\n", false},
		{"podman 5.2.0 replaces in one create", "5.2.0\n", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ctrCfg(config.Podman, "true")
			cfg.Redundancy.PSK = "test-psk"
			cfg.Podman.QuadletDir = t.TempDir()
			m, rr, _ := newCapMgr(cfg, config.Podman)
			m.Geteuid = func() int { return -1 }
			versionAnswer(rr, tc.version)
			if err := m.Deploy(context.Background(), config.Primary); err != nil {
				t.Fatalf("Deploy: %v", err)
			}
			if n := countCalls(rr, "podman", []string{"version", "--format", "{{.Client.Version}}"}); n != 1 {
				t.Errorf("the client version is asked once per deploy, was asked %d times:\n%+v", n, rr.calls)
			}
			// Names carry the container name, so two brokers on one host cannot overwrite
			// each other's entries in the shared podman store.
			for name := range values {
				rm := hasCall(rr, "podman", []string{"secret", "rm", "--ignore", name})
				plain := hasCall(rr, "podman", secretCreateArgs(name, false))
				replaced := hasCall(rr, "podman", secretCreateArgs(name, true))
				if tc.replace && (rm || plain || !replaced) {
					t.Errorf("%s: want one `secret create --replace` alone (rm=%v plain=%v replace=%v):\n%+v",
						name, rm, plain, replaced, rr.calls)
				}
				if !tc.replace && (!rm || !plain || replaced) {
					t.Errorf("%s: want `secret rm --ignore` then a plain create (rm=%v plain=%v replace=%v):\n%+v",
						name, rm, plain, replaced, rr.calls)
				}
				if !tc.replace {
					rmAt := callIndex(rr, "podman", []string{"secret", "rm", "--ignore", name})
					createAt := callIndex(rr, "podman", secretCreateArgs(name, false))
					if rmAt > createAt {
						t.Errorf("%s: the rm must come before the create, got rm at %d and create at %d", name, rmAt, createAt)
					}
				}
			}
			// The values ride stdin, so they must never appear in an argv.
			for _, c := range rr.calls {
				for _, a := range c.args {
					for _, leak := range []string{"secret-pass", "test-psk"} {
						if strings.Contains(a, leak) {
							t.Errorf("secret value reached an argv: %s %v", c.name, c.args)
						}
					}
				}
			}
			assertMode(t, filepath.Join(cfg.Podman.QuadletDir, "sol-pod.container"), 0o600)
		})
	}
}

// TestManagerDeployPodmanSkipsUnchangedSecrets: every podman secret is compared by reading
// its stored value back (podman 4.7+): one that equals the configured value is left in the
// store and not rewritten, a different one is rewritten, and neither value is printed.
func TestManagerDeployPodmanSkipsUnchangedSecrets(t *testing.T) {
	cfg := ctrCfg(config.Podman, "true")
	cfg.Redundancy.PSK = "test-psk"
	cfg.Podman.QuadletDir = t.TempDir()
	m, rr, buf := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	podmanHost(rr, "inactive", map[string]string{
		"sol-pod-admin-password": "secret-pass", // unchanged
		"sol-pod-redundancy-psk": "an-older-psk",
	})
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if wroteSecret(rr, "sol-pod-admin-password") {
		t.Errorf("an unchanged secret must be left in the store:\n%+v", rr.calls)
	}
	if !hasCall(rr, "podman", secretCreateArgs("sol-pod-redundancy-psk", true)) {
		t.Errorf("a changed secret must be rewritten with its new digest label:\n%+v", rr.calls)
	}
	if strings.Contains(buf.String(), "an-older-psk") || strings.Contains(buf.String(), "test-psk") {
		t.Errorf("no secret value may reach the output:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "podman secret sol-pod-admin-password already holds this value") {
		t.Errorf("Deploy should say which secret it left in place:\n%s", buf.String())
	}
}

// TestManagerDeployPodmanSecretReadBackProbe pins the comparison: the podman version is asked
// once, then one `secret inspect --showsecret` per secret before anything is written, and the
// stored value never reaches an echoed command, the report or an error.
func TestManagerDeployPodmanSecretReadBackProbe(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = t.TempDir()
	m, rr, buf := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	podmanHost(rr, "inactive", map[string]string{"sol-pod-admin-password": "an-older-password"})
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	at := callIndex(rr, "podman", secretDataArgs("sol-pod-admin-password"))
	create := callIndex(rr, "podman", secretCreateArgs("sol-pod-admin-password", true))
	if at < 0 || create < 0 || at > create {
		t.Errorf("want the value read back (%d) before the secret is written (%d):\n%+v", at, create, rr.calls)
	}
	if n := countCalls(rr, "podman", []string{"version", "--format", "{{.Client.Version}}"}); n != 1 {
		t.Errorf("the version is asked once, was asked %d times", n)
	}
	for _, c := range rr.calls {
		for _, a := range c.args {
			if strings.Contains(a, "an-older-password") || strings.Contains(a, "secret-pass") {
				t.Errorf("a secret value reached an argv: %s %v", c.name, c.args)
			}
		}
	}
	if strings.Contains(buf.String(), "an-older-password") || strings.Contains(buf.String(), "secret-pass") {
		t.Errorf("a secret value reached the output:\n%s", buf.String())
	}
}

// TestManagerDeployPodmanBelow47NeverReadsBack: a podman below 4.7 has no --showsecret, so
// every secret counts as changed and no SecretData call is made; an unreadable version reads
// the same way.
func TestManagerDeployPodmanBelow47NeverReadsBack(t *testing.T) {
	for _, version := range []string{"4.6.1\n", "podman\n", ""} {
		cfg := ctrCfg(config.Podman, "false")
		cfg.Podman.QuadletDir = t.TempDir()
		m, rr, _ := newCapMgr(cfg, config.Podman)
		m.Geteuid = func() int { return -1 }
		// The store holds the very value configured: on 4.7 it would be skipped.
		podmanHostVersion(rr, "inactive", version, map[string]string{"sol-pod-admin-password": "secret-pass"})
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("version %q: Deploy: %v", version, err)
		}
		if hasCall(rr, "podman", secretDataArgs("sol-pod-admin-password")) {
			t.Errorf("version %q: no SecretData call may be made below 4.7:\n%+v", version, rr.calls)
		}
		if !wroteSecret(rr, "sol-pod-admin-password") {
			t.Errorf("version %q: every secret counts as changed:\n%+v", version, rr.calls)
		}
	}
}

// TestManagerDeployPodmanTreatsAFailedReadBackAsChanged: a `secret inspect` that fails (no
// such secret on a first deploy) is a mismatch, so the secret is written.
func TestManagerDeployPodmanTreatsAFailedReadBackAsChanged(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = t.TempDir()
	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	podmanHost(rr, "inactive", map[string]string{"sol-pod-admin-password": "secret-pass"})
	rr.outFail = func(_ string, args []string) error {
		if slices.Contains(args, "--showsecret") {
			return errors.New("no such secret")
		}
		return nil
	}
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if !wroteSecret(rr, "sol-pod-admin-password") {
		t.Errorf("an unreadable stored value counts as changed, so the secret is written:\n%+v", rr.calls)
	}
}

// TestManagerDeployPodmanReplaceOnlyOnPodman47 is the floor, from the other side: the
// pin that no argv may carry --replace is gone, since podman 4.7 and later has it and the
// deploy uses it there. What stays pinned is that it is NEVER sent below 4.7, nor when the
// version cannot be read or parsed, which is what keeps the older floor working.
func TestManagerDeployPodmanReplaceOnlyOnPodman47(t *testing.T) {
	for _, tc := range []struct {
		name, version string
		fail          bool
		want          bool
	}{
		{"4.6.9 is below the floor", "4.6.9\n", false, false},
		{"4.7.0 is the floor", "4.7.0\n", false, true},
		{"surrounding space is ignored", "  5.2.0 \n", false, true},
		{"an unparseable version", "podman\n", false, false},
		{"an empty answer", "", false, false},
		{"a probe that fails", "5.2.0\n", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ctrCfg(config.Podman, "false")
			cfg.Podman.QuadletDir = t.TempDir()
			m, rr, _ := newCapMgr(cfg, config.Podman)
			m.Geteuid = func() int { return -1 }
			versionAnswer(rr, tc.version)
			if tc.fail {
				rr.outFail = func(name string, args []string) error {
					if len(args) > 0 && args[0] == "version" {
						return errors.New("boom")
					}
					return nil
				}
			}
			if err := m.Deploy(context.Background(), config.Primary); err != nil {
				t.Fatalf("Deploy: %v", err)
			}
			got := false
			for _, c := range rr.calls {
				if slices.Contains(c.args, "--replace") {
					got = true
				}
			}
			if got != tc.want {
				t.Errorf("--replace sent = %v, want %v:\n%+v", got, tc.want, rr.calls)
			}
		})
	}
}

// TestManagerDeployPreviewNeverReplaces: the Echo runner probes nothing, so a preview
// shows the remove-then-create pair that works on every supported podman, each create
// labelled with the placeholder rather than a digest of it.
func TestManagerDeployPreviewNeverReplaces(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = t.TempDir()
	m, buf := newEchoMgr(cfg, config.Podman)
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "--replace") || strings.Contains(out, "{{.Client.Version}}") {
		t.Errorf("a preview must not probe the version or echo --replace:\n%s", out)
	}
	if !strings.Contains(out, "secret rm --ignore") {
		t.Errorf("a preview shows the remove-then-create pair:\n%s", out)
	}
	if strings.Contains(out, "--label") || strings.Contains(out, "--showsecret") {
		t.Errorf("a preview probes nothing and labels nothing:\n%s", out)
	}
}

// TestManagerDeployDockerDryRunMasksSecretEnv covers the dry-run contract for the
// one path that carries values: the echo names the variables compose would be
// given and prints none of their values, and the file on stdin shows as a byte count.
// A preview asks nothing and writes nothing.
func TestManagerDeployDockerDryRunMasksSecretEnv(t *testing.T) {
	cfg := ctrCfg(config.Docker, "true")
	cfg.Redundancy.PSK = "" // empty is fine here: prep host has not generated it yet
	m, buf := newEchoMgr(cfg, config.Docker)
	m.Confirm = neverAsk(t)
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("dry-run Deploy must stay previewable before prep host: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "SOLACE_ADMIN_PASSWORD=***") || !strings.Contains(out, render.DeployDigestEnvVar+"=***") {
		t.Errorf("dry-run should name the secret variables and the digest with masked values:\n%s", out)
	}
	if strings.Contains(out, "secret-pass") {
		t.Errorf("dry-run echo leaked the admin password:\n%s", out)
	}
	if !strings.Contains(out, "+ docker compose -f - --project-directory . up -d") ||
		!strings.Contains(out, "bytes on stdin") {
		t.Errorf("dry-run should echo the compose command with its file as stdin bytes:\n%s", out)
	}
	if strings.Contains(out, "services:") {
		t.Errorf("the file on stdin must show as a size, not as text:\n%s", out)
	}
}

func TestManagerDeployPodmanDryRunHidesSecretBytes(t *testing.T) {
	cfg := ctrCfg(config.Podman, "true")
	cfg.Redundancy.PSK = "test-psk"
	cfg.Podman.QuadletDir = t.TempDir()
	m, buf := newEchoMgr(cfg, config.Podman)
	m.Confirm = neverAsk(t)
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "secret create sol-pod-admin-password -") {
		t.Errorf("dry-run should echo the secret-create command:\n%s", out)
	}
	if !strings.Contains(out, "secret rm --ignore sol-pod-admin-password") {
		t.Errorf("dry-run should echo the remove that precedes each create:\n%s", out)
	}
	if !strings.Contains(out, "bytes on stdin") {
		t.Errorf("dry-run should report the value as stdin bytes, not print it:\n%s", out)
	}
	for _, leak := range []string{"secret-pass", "test-psk"} {
		if strings.Contains(out, leak) {
			t.Errorf("dry-run echo leaked %q:\n%s", leak, out)
		}
	}
}

// TestManagerDeployRestartsAReloadingUnit: `systemctl is-active` exits 0 for
// "reloading" as well as "active", and a unit in either state is running, so a changed
// unit is restarted rather than handed a start that does nothing.
func TestManagerDeployRestartsAReloadingUnit(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = t.TempDir()
	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	m.RestartApproved = true
	rr.out = []byte("reloading\n")
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if !hasCall(rr, "systemctl", withUser(cfg, "restart", "sol-pod.service")) {
		t.Errorf("a reloading unit must be restarted:\n%+v", rr.calls)
	}
	if hasCall(rr, "systemctl", withUser(cfg, "start", "sol-pod.service")) {
		t.Errorf("a start on a running unit does nothing, so it must not be the answer:\n%+v", rr.calls)
	}
}

// countCalls is hasCall's counting form, for the probes that must run exactly once.
func countCalls(rr *capRunner, name string, args []string) int {
	n := 0
	for _, c := range rr.calls {
		if c.name == name && eqArgs(c.args, args) {
			n++
		}
	}
	return n
}

func TestManagerDeployRejectsEmptySecret(t *testing.T) {
	cfg := ctrCfg(config.Docker, "true") // HA -> the PSK secret is required too
	cfg.Redundancy.PSK = ""
	m, _, _ := newCapMgr(cfg, config.Docker)
	err := m.Deploy(context.Background(), config.Primary)
	if err == nil {
		t.Fatal("Deploy must fail loud when a required secret is empty")
	}
	if !strings.Contains(err.Error(), "redundancy.psk") || !strings.Contains(err.Error(), "env file") {
		t.Errorf("error should name the field and the fix, got: %v", err)
	}
}

func TestManagerDeployPodmanWritesUnit(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = dir
	// Match this process's euid so the rootless/rootful guard passes on any host
	// (on Windows Geteuid()<0 and the guard is skipped entirely).
	if os.Geteuid() != 0 {
		cfg.Podman.Rootful = false
		cfg.Podman.SystemctlUser = "--user"
	}
	m, rr, _ := newCapMgr(cfg, config.Podman)
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if !fileExists(filepath.Join(dir, "sol-pod.container")) {
		t.Error("Deploy should write the quadlet unit")
	}
	if !hasCall(rr, "systemctl", withUser(cfg, "daemon-reload")) {
		t.Errorf("Deploy should daemon-reload:\n%+v", rr.calls)
	}
	if !hasCall(rr, "systemctl", withUser(cfg, "start", "sol-pod.service")) {
		t.Errorf("Deploy should start the service:\n%+v", rr.calls)
	}
}

func TestManagerDeployPodmanDryRunSkipsWrite(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = dir
	m, buf := newEchoMgr(cfg, config.Podman)
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if fileExists(filepath.Join(dir, "sol-pod.container")) {
		t.Error("Deploy must not write the quadlet unit under the Echo runner")
	}
	out := buf.String()
	for _, want := range []string{"+ systemctl daemon-reload", "+ systemctl start sol-pod.service"} {
		if !strings.Contains(out, want) {
			t.Errorf("dry-run Deploy should echo %q:\n%s", want, out)
		}
	}
}

// --- Delete -----------------------------------------------------------------

func TestManagerDeletePodmanRemovesUnit(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = dir
	unit := filepath.Join(dir, "sol-pod.container")
	if err := os.WriteFile(unit, []byte("[Unit]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m, rr, _ := newCapMgr(cfg, config.Podman)
	if err := m.Delete(context.Background(), false); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if fileExists(unit) {
		t.Error("Delete should remove the quadlet unit")
	}
	if !hasCall(rr, "systemctl", []string{"stop", "sol-pod.service"}) {
		t.Errorf("Delete should stop the service:\n%+v", rr.calls)
	}
	if !hasCall(rr, "systemctl", []string{"daemon-reload"}) {
		t.Errorf("Delete should daemon-reload:\n%+v", rr.calls)
	}
}

// TestManagerDeletePodmanStopFailsServiceActiveBlocksRemoval covers B2: a
// stop failure alone proves nothing -- "podman info" (Preflight) only shows the
// engine is reachable -- so a unit that serviceActive still reports active must
// block the unit removal, the daemon-reload, and (via Delete's purge gate) the
// data-directory rm, rather than reporting success over a broker that is still
// serving traffic.
func TestManagerDeletePodmanStopFailsServiceActiveBlocksRemoval(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = dir
	unit := filepath.Join(dir, "sol-pod.container")
	if err := os.WriteFile(unit, []byte("[Unit]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m, rr, _ := newCapMgr(cfg, config.Podman)
	rr.fail = failOn("stop")
	rr.out = []byte("active\n") // serviceActive's `is-active` probe
	err := m.Delete(context.Background(), true)
	if err == nil {
		t.Fatal("Delete must fail when stop fails and the unit is still active")
	}
	if !strings.Contains(err.Error(), "sol-pod.service") {
		t.Errorf("error should name the unit, got: %v", err)
	}
	if !fileExists(unit) {
		t.Error("a still-active unit must not have its quadlet unit removed")
	}
	if hasCall(rr, "systemctl", []string{"daemon-reload"}) {
		t.Errorf("daemon-reload must not run when removal was blocked:\n%+v", rr.calls)
	}
	if hasCall(rr, "rm", []string{"-rf", "/opt/solace/data"}) {
		t.Errorf("the data-dir purge must never be reached when removal was blocked:\n%+v", rr.calls)
	}
}

// TestManagerDeletePodmanStopFailsServiceInactiveProceeds is the same failed
// stop, but serviceActive confirms the unit is already down -- the benign
// "already stopped" case, which still proceeds exactly as before.
func TestManagerDeletePodmanStopFailsServiceInactiveProceeds(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = dir
	unit := filepath.Join(dir, "sol-pod.container")
	if err := os.WriteFile(unit, []byte("[Unit]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m, rr, buf := newCapMgr(cfg, config.Podman)
	rr.fail = failOn("stop")
	rr.out = []byte("inactive\n")
	if err := m.Delete(context.Background(), false); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if fileExists(unit) {
		t.Error("an inactive unit's quadlet file should still be removed")
	}
	if !hasCall(rr, "systemctl", []string{"daemon-reload"}) {
		t.Errorf("Delete should still daemon-reload:\n%+v", rr.calls)
	}
	if !strings.Contains(buf.String(), "inactive") {
		t.Errorf("Delete should warn about the failed stop and name the confirmed state:\n%s", buf.String())
	}
}

// TestManagerDeletePodmanStopFailsStateUnknownBlocksRemoval is the case the
// whole guard exists for: `systemctl stop` failed AND `systemctl is-active`
// could not answer either -- which is what an unreachable rootless user session
// looks like, while `podman info` (Preflight) succeeds on the engine socket.
// Reading that silence as "already stopped" is what would delete a live broker's
// message spool, so it must refuse exactly as a confirmed-active unit does.
func TestManagerDeletePodmanStopFailsStateUnknownBlocksRemoval(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = dir
	unit := filepath.Join(dir, "sol-pod.container")
	if err := os.WriteFile(unit, []byte("[Unit]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m, rr, _ := newCapMgr(cfg, config.Podman)
	rr.fail = failOn("stop")
	// rr.out deliberately unset: the is-active probe answers nothing at all.
	err := m.Delete(context.Background(), true)
	if err == nil {
		t.Fatal("Delete must fail when the stop failed and the state probe answered nothing")
	}
	if !strings.Contains(err.Error(), "NOT confirmed stopped") {
		t.Errorf("error should say the unit is not confirmed stopped, got: %v", err)
	}
	if !fileExists(unit) {
		t.Error("an unconfirmed unit must not have its quadlet unit removed")
	}
	if hasCall(rr, "rm", []string{"-rf", "/opt/solace/data"}) {
		t.Errorf("the data-dir purge must never be reached on an unconfirmed stop:\n%+v", rr.calls)
	}
}

// TestManagerDeletePodmanRemovesSecrets covers M5: nothing ever removed the
// secrets createPodmanSecrets loaded into podman's own store, so they outlived
// `remove all --delete-data`. deletePodman must remove every one of them (by
// the same render.ContainerSecrets list createPodmanSecrets uses), and a
// failing removal must warn rather than fail a teardown that otherwise
// succeeded.
func TestManagerDeletePodmanRemovesSecrets(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = t.TempDir()
	m, rr, buf := newCapMgr(cfg, config.Podman)
	rr.fail = failOn("secret")
	if err := m.Delete(context.Background(), false); err != nil {
		t.Fatalf("Delete should tolerate a failing secret rm: %v", err)
	}
	for _, s := range render.ContainerSecrets(cfg, config.Podman) {
		// --ignore matches the create path, so one argv shape serves both.
		if !hasCall(rr, "podman", []string{"secret", "rm", "--ignore", s.Name}) {
			t.Errorf("Delete should remove podman secret %s:\n%+v", s.Name, rr.calls)
		}
	}
	// The warning no longer guesses "already removed?": --ignore makes a missing
	// secret a success, so what reaches this branch is a real failure.
	if !strings.Contains(buf.String(), "removing podman secret") {
		t.Errorf("a failing secret rm should warn, not fail the removal:\n%s", buf.String())
	}
	if strings.Contains(buf.String(), "already removed") {
		t.Errorf("the warning must not offer 'already removed' as a cause once --ignore absorbs it:\n%s",
			buf.String())
	}
}

func TestManagerDeletePodmanPurgeRootless(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.Rootful = false
	cfg.Podman.QuadletDir = t.TempDir()
	m, buf := newEchoMgr(cfg, config.Podman)
	if err := m.Delete(context.Background(), true); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	// The contents belong to a subuid, so clearing them needs the same namespace
	// the chown used -- and the directory itself stays, keeping the ownership and
	// mode prep established.
	if !strings.Contains(buf.String(), "+ podman unshare find /opt/solace/data -mindepth 1 -delete") {
		t.Errorf("rootless purge should clear through `podman unshare`:\n%s", buf.String())
	}
	if strings.Contains(buf.String(), "rm -rf /opt/solace/data") {
		t.Errorf("purge must not remove the directory itself:\n%s", buf.String())
	}
}

// TestManagerDeleteDockerComposeDown: a removal feeds the file rendered now to
// `compose down` on stdin, so no compose file has to be on the host. Down interpolates the
// whole model, so every secret is defined, and the deploy-digest variable falls back to the
// empty default the label gives it.
func TestManagerDeleteDockerComposeDown(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.Delete(context.Background(), false); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	want := []string{"compose", "-f", "-", "--project-directory", ".", "down"}
	var down capCall
	for _, c := range rr.calls {
		if slices.Contains(c.args, "down") {
			down = c
		}
	}
	if down.name != "docker" || !eqArgs(down.args, want) {
		t.Fatalf("Delete should compose down with the file on stdin, got %s %v:\n%+v", down.name, down.args, rr.calls)
	}
	if down.stdin != string(render.Compose(cfg, cfg.ResolveNode(config.Primary))) {
		t.Errorf("down's stdin is not the rendered compose file:\n%s", down.stdin)
	}
	// Defined, so compose can interpolate; the certificate is the one value down never
	// reads from disk (TestTeardownVerbsDoNotReadTheCertificate).
	if !containsStr(down.env, "SOLACE_ADMIN_PASSWORD=secret-pass") {
		t.Errorf("down needs the secrets defined, got %v", maskedKeys(down.env))
	}
	if down.method != "RunEnv" {
		t.Errorf("down carries its environment the same way up does, got %s", down.method)
	}
}

func TestManagerDeleteDockerPurgeClearsDataDir(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.Delete(context.Background(), true); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !hasCall(rr, "docker", []string{"compose", "-f", "-", "--project-directory", ".", "down"}) {
		t.Errorf("Delete should compose down:\n%+v", rr.calls)
	}
	// The CONTENTS go, the directory stays: removing it would throw away the
	// ownership and mode prep established, which on a rootless host only
	// `podman unshare` can put back.
	if !hasCall(rr, "find", []string{"/opt/solace/data", "-mindepth", "1", "-delete"}) {
		t.Errorf("purge should clear the data dir (rootful/docker):\n%+v", rr.calls)
	}
	for _, c := range rr.calls {
		if c.name == "rm" {
			t.Errorf("purge must not remove the directory itself: %s %v", c.name, c.args)
		}
	}
}

// TestManagerDeleteDockerNeedsNoComposeFile: a teardown no longer looks for a file on
// disk or falls back to stop and rm by name. A host that never had a compose file (or
// whose earlier build's file was deleted) comes down the same way, by the project name in
// the file rendered now.
func TestManagerDeleteDockerNeedsNoComposeFile(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(t.TempDir(), "absent.yml") // retired: never read
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.Delete(context.Background(), false); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	for _, c := range rr.calls {
		if len(c.args) > 0 && (c.args[0] == "stop" || c.args[0] == "rm") {
			t.Errorf("Delete must not fall back to stop/rm: %s %v", c.name, c.args)
		}
		if slices.Contains(c.args, cfg.Docker.ComposeFile) {
			t.Errorf("the retired composeFile must not reach any command: %s %v", c.name, c.args)
		}
	}
}

// --- Lifecycle: Start / Stop / Restart ---------------------------------------

// TestManagerLifecyclePodmanSystemctl covers Start/Stop/Restart on podman: all
// three drive systemd rather than the engine directly, since the quadlet unit
// owns the container's lifecycle. Both rootful and rootless are covered because
// the user flag (`--user`) is set only in the rootless case.
func TestManagerLifecyclePodmanSystemctl(t *testing.T) {
	for _, rootless := range []bool{false, true} {
		label := "rootful"
		if rootless {
			label = "rootless"
		}
		t.Run(label, func(t *testing.T) {
			cfg := ctrCfg(config.Podman, "false")
			if rootless {
				cfg.Podman.SystemctlUser = "--user"
			}
			cases := []struct {
				name, verb string
				call       func(*Manager) error
			}{
				{"start", "start", func(m *Manager) error { return m.Start(context.Background()) }},
				{"stop", "stop", func(m *Manager) error { return m.Stop(context.Background()) }},
				{"restart", "restart", func(m *Manager) error { return m.Restart(context.Background()) }},
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					m, rr, _ := newCapMgr(cfg, config.Podman)
					if err := tc.call(m); err != nil {
						t.Fatalf("%s: %v", tc.name, err)
					}
					if !hasCall(rr, "systemctl", withUser(cfg, tc.verb, "sol-pod.service")) {
						t.Errorf("%s should run systemctl %s on the unit:\n%+v", tc.name, tc.verb, rr.calls)
					}
				})
			}
		})
	}
}

// TestStatusAllFindsBrokersByImage covers the container answer to `--all`:
// discovery is by IMAGE, not by the name this env file configured, so a broker
// someone deployed by hand still shows up and an unrelated container never does.
// That is the whole reason the flag exists -- a config-scoped listing cannot answer
// "what is actually running on this host".
func TestStatusAllFindsBrokersByImage(t *testing.T) {
	ps := "solace\tsolace/solace-pubsub-standard:10.10.1.128\tUp 3 days\n" +
		"legacy-broker\tsolace/solace-pubsubplus-enterprise:10.9\tUp 1 hour\n" +
		"nginx\tnginx:latest\tUp 2 days\n"
	m, rr, out := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	rr.out = []byte(ps)
	if err := m.StatusAll(context.Background(), false); err != nil {
		t.Fatalf("StatusAll: %v", err)
	}
	got := out.String()
	for _, want := range []string{"solace", "legacy-broker"} {
		if !strings.Contains(got, want) {
			t.Errorf("StatusAll output missing the Solace container %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "nginx") {
		t.Errorf("StatusAll listed a non-Solace container:\n%s", got)
	}
}

// TestStatusAllReportsNothingFound: an empty listing says so in words. This is the
// ordinary case on a host that has not been deployed to yet, and printing a bare
// header there reads as though the command failed to look.
func TestStatusAllReportsNothingFound(t *testing.T) {
	m, rr, out := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	rr.out = []byte("nginx\tnginx:latest\tUp 2 days\n")
	if err := m.StatusAll(context.Background(), false); err != nil {
		t.Fatalf("StatusAll: %v", err)
	}
	if !strings.Contains(out.String(), "no Solace broker containers") {
		t.Errorf("StatusAll with nothing to show should say so:\n%s", out.String())
	}
}

// errListFailed stands in for an engine that cannot be asked.
var errListFailed = errors.New("cannot connect to the docker daemon")

// TestStatusAllWrapsListError: an engine that cannot be asked -- docker not
// running, podman socket down -- fails loud and names the cause. Reporting an empty
// host instead would be a lie in exactly the situation an operator is trying to
// diagnose.
func TestStatusAllWrapsListError(t *testing.T) {
	m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	rr.outErr = errListFailed
	err := m.StatusAll(context.Background(), false)
	if err == nil || !strings.Contains(err.Error(), "listing containers") {
		t.Fatalf("StatusAll err = %v, want it to name the listing failure", err)
	}
	if !errors.Is(err, errListFailed) {
		t.Errorf("StatusAll should preserve the cause: %v", err)
	}
}

// TestStatusAllDetailInspectsEachAndKeepsSecretsOut: --detail goes deeper on every
// container it found, and the deeper view names mounts and never the environment. A
// broker's secrets are files under /mnt/secrets, so mounts name them without reading
// them -- while docker's compose secrets are environment-sourced, so printing the
// environment would put passwords on the terminal and into scrollback.
//
// The selection moved from the argv to the DECODER (inspect.go): `inspect` is now
// asked for everything and this tool picks, because the field names differ between
// docker and podman in ways a --format template gets silently wrong. So the argv
// assertion here is that inspect is called plainly, and what is and is not printed is
// asserted against the report.
func TestStatusAllDetailInspectsEachAndKeepsSecretsOut(t *testing.T) {
	const inspected = `[{"Name": "/solace", "State": {"Status": "running", "Running": true},
	  "Config": {"Image": "solace/solace-pubsub-standard:10.10.1.128",
	             "Env": ["SOLACE_ADMIN_PASSWORD=must-not-be-printed"]},
	  "Mounts": [{"Source": "/opt/solace/secrets/admin", "Destination": "/mnt/secrets/admin_password", "RW": false}]}]`
	m, rr, out := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	rr.outFor = func(_ string, args []string) []byte {
		if slices.Contains(args, "inspect") {
			return []byte(inspected)
		}
		return []byte("solace\tsolace/solace-pubsub-standard:10.10.1.128\tUp 3 days\n")
	}
	if err := m.StatusAll(context.Background(), true); err != nil {
		t.Fatalf("StatusAll --detail: %v", err)
	}
	if !hasCall(rr, "docker", []string{"container", "inspect", "solace"}) {
		t.Errorf("StatusAll --detail should inspect each container it found:\n%+v", rr.calls)
	}
	got := out.String()
	if strings.Contains(got, "must-not-be-printed") || strings.Contains(got, "SOLACE_ADMIN_PASSWORD") {
		t.Errorf("the container environment reached the report:\n%s", got)
	}
	if !strings.Contains(got, "/mnt/secrets/admin_password") {
		t.Errorf("--detail should name each mounted path:\n%s", got)
	}
}

// TestManagerLifecycleDocker: Start/Stop/Restart act on the container itself with the
// plain engine verb, which works on every compose implementation and needs no compose
// file -- on disk or on stdin. They are the same verbs on a host that has a leftover
// compose file as on one that never had one.
func TestManagerLifecycleDocker(t *testing.T) {
	cases := []struct {
		name, verb string
		call       func(*Manager) error
	}{
		{"start", "start", func(m *Manager) error { return m.Start(context.Background()) }},
		{"stop", "stop", func(m *Manager) error { return m.Stop(context.Background()) }},
		{"restart", "restart", func(m *Manager) error { return m.Restart(context.Background()) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
			if err := tc.call(m); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if !hasCall(rr, "docker", []string{tc.verb, "solace"}) {
				t.Errorf("%s should run the plain runtime verb on the container:\n%+v", tc.name, rr.calls)
			}
			for _, c := range rr.calls {
				if slices.Contains(c.args, "compose") {
					t.Errorf("%s must not go through compose: %s %v", tc.name, c.name, c.args)
				}
			}
		})
	}
}

// TestManagerLifecycleDockerDryRun: the preview shows the same plain verb.
func TestManagerLifecycleDockerDryRun(t *testing.T) {
	m, buf := newEchoMgr(ctrCfg(config.Docker, "false"), config.Docker)
	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !strings.Contains(buf.String(), "+ docker start solace") {
		t.Errorf("dry-run Start should preview the plain verb:\n%s", buf.String())
	}
}

// --- Status / Logs / CLI / Shell --------------------------------------------

// listArgs is containercli.List's argv after the program, for the engine named by base.
func listArgs(running bool) []string {
	return containercli.List([]string{"x"}, running).Argv[1:]
}

// statusListing answers `ps` with these rows and every other probe (the inspect that
// follows) with nothing, which is a tolerated failure to decode.
func statusListing(rr *capRunner, rows string) {
	rr.outFor = func(_ string, args []string) []byte {
		if slices.Contains(args, "inspect") {
			return nil
		}
		return []byte(rows)
	}
}

func TestManagerStatusPodman(t *testing.T) {
	m, rr, _ := newCapMgr(ctrCfg(config.Podman, "false"), config.Podman)
	if err := m.Status(context.Background()); err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !hasCall(rr, "systemctl", []string{"status", "sol-pod.service", "--no-pager"}) {
		t.Errorf("podman Status should show the unit:\n%+v", rr.calls)
	}
	if !hasCall(rr, "podman", listArgs(false)) {
		t.Errorf("podman Status should list the containers:\n%+v", rr.calls)
	}
}

// TestManagerStatusDocker covers docker Status: ONE listing, `ps --all --no-trunc` with an
// explicit three-column template, filtered to this container's exact name in Go and
// printed as this tool's own table. A sibling whose name merely starts with it, and an
// unrelated container, are never in it -- which `--filter name=` (an unanchored regex on
// both engines) could not promise.
func TestManagerStatusDocker(t *testing.T) {
	m, rr, out := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	const img = "solace/solace-pubsub-standard:10.10.1.128"
	statusListing(rr, "solace\t"+img+"\tUp 3 days\nsolace-edge\t"+img+"\tUp 1 hour\nnginx\tnginx:latest\tUp 2 days\n")
	if err := m.Status(context.Background()); err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !hasCall(rr, "docker", listArgs(false)) {
		t.Errorf("docker Status should list the containers:\n%+v", rr.calls)
	}
	got := out.String()
	for _, want := range []string{"NAME", "IMAGE", "STATUS", "Up 3 days"} {
		if !strings.Contains(got, want) {
			t.Errorf("Status table should contain %q:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"solace-edge", "nginx", "compose", "Ports"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("Status table should not contain %q:\n%s", unwanted, got)
		}
	}
	for _, c := range rr.calls {
		if slices.Contains(c.args, "compose") {
			t.Errorf("Status must not run compose: %s %v", c.name, c.args)
		}
	}
}

// TestManagerStatusDockerDryRun: the preview shows the listing it would make, and not a
// claim that the container is missing, since an Echo listing answers nothing.
func TestManagerStatusDockerDryRun(t *testing.T) {
	m, buf := newEchoMgr(ctrCfg(config.Docker, "false"), config.Docker)
	if err := m.Status(context.Background()); err != nil {
		t.Fatalf("Status: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"+ docker ps --all --no-trunc", "--format", "{{.Status}}", "+ docker container inspect solace"} {
		if !strings.Contains(out, want) {
			t.Errorf("docker Status missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "no container named") {
		t.Errorf("a preview must not say the container is missing:\n%s", out)
	}
}

// TestManagerStatusSaysWhenTheContainerIsNotThere: an empty listing is stated in words,
// the way the --all listing states it.
func TestManagerStatusSaysWhenTheContainerIsNotThere(t *testing.T) {
	m, rr, out := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	statusListing(rr, "nginx\tnginx:latest\tUp 2 days\n")
	if err := m.Status(context.Background()); err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !strings.Contains(out.String(), "no container named solace on this host") {
		t.Errorf("Status should say the container is not there:\n%s", out.String())
	}
}

// TestManagerStatusFailsWhenTheListingFails: unlike the inspect detail, the listing is
// what Status is for, so its failure is the command's.
func TestManagerStatusFailsWhenTheListingFails(t *testing.T) {
	m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	rr.outErr = errListFailed
	if err := m.Status(context.Background()); !errors.Is(err, errListFailed) {
		t.Fatalf("Status err = %v, want the listing's own error", err)
	}
}

// TestStatusPrintsEngineTextThroughPrintable: the name, image and status come from the
// engine about a container this tool did not necessarily create, and a terminal escape in
// one would be acted on by the terminal that prints it. C1 controls go too.
func TestStatusPrintsEngineTextThroughPrintable(t *testing.T) {
	m, rr, out := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	statusListing(rr, "solace\tsolace/solace-pubsub-standard\x1b]0;pwned\x07\tUp\u009b31m 3 days\n")
	if err := m.Status(context.Background()); err != nil {
		t.Fatalf("Status: %v", err)
	}
	got := out.String()
	for _, bad := range []string{"\x1b", "\x07", "\u009b"} {
		if strings.Contains(got, bad) {
			t.Errorf("a control character %q reached the table:\n%q", bad, got)
		}
	}
	if !strings.Contains(got, "solace/solace-pubsub-standard]0;pwned") {
		t.Errorf("the printable text must survive, minus the controls:\n%q", got)
	}
}

func TestManagerStatusPodmanUnitInactiveTolerated(t *testing.T) {
	m, rr, buf := newCapMgr(ctrCfg(config.Podman, "false"), config.Podman)
	rr.fail = failOn("status")
	if err := m.Status(context.Background()); err != nil {
		t.Fatalf("Status should tolerate an inactive unit: %v", err)
	}
	if !strings.Contains(buf.String(), "non-zero") {
		t.Errorf("Status should warn when the unit is not active:\n%s", buf.String())
	}
	if !hasCall(rr, "podman", listArgs(false)) {
		t.Errorf("Status should still list the containers:\n%+v", rr.calls)
	}
}

// TestStatusListingCarriesNoPortsColumn is the property the template exists for, asserted
// on the argv: a broker publishes a dozen or more ports, so the engine's default table
// wraps every other column into illegibility -- on the one report whose job is to answer
// "is it up". The listing names its three columns, and --no-trunc keeps the image intact.
func TestStatusListingCarriesNoPortsColumn(t *testing.T) {
	args := strings.Join(listArgs(false), " ")
	if strings.Contains(args, "Ports") {
		t.Errorf("the listing must not ask for the ports column: %q", args)
	}
	for _, want := range []string{"--all", "--no-trunc", "{{.Names}}", "{{.Image}}", "{{.Status}}"} {
		if !strings.Contains(args, want) {
			t.Errorf("listing argv = %q, want it to contain %q", args, want)
		}
	}
}

func TestManagerLogsCLIShell(t *testing.T) {
	t.Run("logs", func(t *testing.T) {
		m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
		if err := m.Logs(context.Background(), containercli.LogOptions{}, false); err != nil {
			t.Fatalf("Logs: %v", err)
		}
		got := rr.last()
		if got.method != "Run" || got.name != "docker" || !eqArgs(got.args, []string{"logs", "solace"}) {
			t.Errorf("Logs argv: %+v", got)
		}
	})
	t.Run("logs carry every option, before the name", func(t *testing.T) {
		m, rr, _ := newCapMgr(ctrCfg(config.Podman, "false"), config.Podman)
		o := containercli.LogOptions{Follow: true, Timestamps: true, Tail: "50", Since: "5m0s"}
		if err := m.Logs(context.Background(), o, false); err != nil {
			t.Fatalf("Logs: %v", err)
		}
		want := []string{"logs", "-f", "--timestamps", "--tail", "50", "--since", "5m0s", "sol-pod"}
		if got := rr.last(); !eqArgs(got.args, want) {
			t.Errorf("Logs argv = %v, want %v", got.args, want)
		}
	})
	t.Run("cli", func(t *testing.T) {
		m, rr, _ := newCapMgr(ctrCfg(config.Podman, "false"), config.Podman)
		if err := m.CLI(context.Background()); err != nil {
			t.Fatalf("CLI: %v", err)
		}
		got := rr.last()
		if got.method != "RunInteractive" || got.name != "podman" || !eqArgs(got.args, []string{"exec", "-i", "-t", "sol-pod", "cli", "-A"}) {
			t.Errorf("CLI argv: %+v", got)
		}
	})
	t.Run("shell", func(t *testing.T) {
		m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
		if err := m.Shell(context.Background()); err != nil {
			t.Fatalf("Shell: %v", err)
		}
		got := rr.last()
		if got.method != "RunInteractive" || got.name != "docker" || !eqArgs(got.args, []string{"exec", "-i", "-t", "solace", "bash"}) {
			t.Errorf("Shell argv: %+v", got)
		}
	})
}

// TestManagerLogsRefusesPrevious: neither engine keeps a previous container's log, so the
// refusal is the shared one, names the engine, and reaches the runner as nothing.
func TestManagerLogsRefusesPrevious(t *testing.T) {
	for _, p := range []config.Platform{config.Docker, config.Podman} {
		m, rr, _ := newCapMgr(ctrCfg(p, "false"), p)
		err := m.Logs(context.Background(), containercli.LogOptions{}, true)
		if err == nil || !strings.Contains(err.Error(), "--previous is a kubernetes option") ||
			!strings.Contains(strings.ToLower(err.Error()), string(p)) {
			t.Errorf("%s Logs(previous) err = %v, want the shared refusal naming the engine", p, err)
		}
		if len(rr.calls) != 0 {
			t.Errorf("%s: a refused option must not reach the runner:\n%+v", p, rr.calls)
		}
	}
}

// TestManagerSessionNeedsAContainerName: containercli.Exec refuses an empty name, so a
// Manager over a config with no container name fails before it attaches anything.
func TestManagerSessionNeedsAContainerName(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.Container.Name = ""
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.Shell(context.Background()); err == nil || !strings.Contains(err.Error(), "container name") {
		t.Fatalf("Shell err = %v, want the missing-name refusal", err)
	}
	if len(rr.calls) != 0 {
		t.Errorf("nothing may be attached without a name:\n%+v", rr.calls)
	}
}

// --- pure helpers -----------------------------------------------------------

// --- error paths: the capRunner.fail hook drives each `err != nil` wrap branch ---

func TestManagerCheckReachableError(t *testing.T) {
	m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	rr.outErr = fmt.Errorf("no engine") // the `version` probe is the only Output call
	if err := m.Check(context.Background()); err == nil {
		t.Fatal("Check should fail when the runtime version probe errors")
	}
}

func TestManagerPrepHostMkdirError(t *testing.T) {
	m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	rr.fail = failOn("mkdir")
	if err := m.PrepHost(context.Background()); err == nil {
		t.Fatal("PrepHost should propagate a mkdir failure")
	}
}

func TestManagerPrepHostChownError(t *testing.T) {
	m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	rr.fail = failOn("chown")
	if err := m.PrepHost(context.Background()); err == nil {
		t.Fatal("PrepHost should propagate a chown failure")
	}
}

func TestManagerPrepHostRootlessUnshareChownError(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.Rootful = false
	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return 1000 } // avoid the rootless-as-root WARN noise
	// The whole readiness block has to PASS for the chown to be reached at all --
	// without this the id-mapping row refuses first and the test proves nothing.
	rr.outFor = healthyRootlessOut(healthyNrOpen)
	// "chown", not "unshare": the data-dir row's own read-only probe is an unshare
	// call too, so failing every unshare would refuse at the CHECK instead.
	rr.fail = failOn("chown")
	err := m.PrepHost(context.Background())
	if err == nil {
		t.Fatal("rootless PrepHost should propagate an unshare chown failure")
	}
	if !strings.Contains(err.Error(), "chown data dir") {
		t.Errorf("the failure must be the chown, not an earlier row: %v", err)
	}
}

func TestManagerDeployPodmanMkdirError(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(filePath, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = filePath // MkdirAll on a file -> error
	m, _, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 } // skip the euid guard portably
	if err := m.Deploy(context.Background(), config.Primary); err == nil {
		t.Fatal("Deploy should fail when the quadlet dir cannot be created")
	}
}

func TestManagerDeployPodmanWriteUnitError(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = dir
	if err := os.Mkdir(filepath.Join(dir, "sol-pod.container"), 0o755); err != nil {
		t.Fatal(err) // unit path is a directory -> WriteFile fails
	}
	m, _, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	if err := m.Deploy(context.Background(), config.Primary); err == nil {
		t.Fatal("Deploy should fail when the quadlet unit cannot be written")
	}
}

func TestManagerDeployPodmanDaemonReloadError(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = t.TempDir()
	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	rr.fail = failOn("daemon-reload")
	if err := m.Deploy(context.Background(), config.Primary); err == nil {
		t.Fatal("Deploy should propagate a daemon-reload failure")
	}
}

func TestManagerDeployPodmanStartError(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = t.TempDir()
	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	rr.fail = failOn("start")
	if err := m.Deploy(context.Background(), config.Primary); err == nil {
		t.Fatal("Deploy should propagate a service-start failure")
	}
}

func TestManagerDeployPodmanEUIDGuardFails(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.Rootful = false
	cfg.Podman.QuadletDir = t.TempDir()
	m, _, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return 0 } // rootless as root -> guard rejects
	if err := m.Deploy(context.Background(), config.Primary); err == nil {
		t.Fatal("Deploy should fail when the euid guard rejects the host")
	}
}

func TestManagerDeployDockerComposeUpError(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	m, rr, _ := newCapMgr(cfg, config.Docker)
	rr.fail = failOn("compose")
	if err := m.Deploy(context.Background(), config.Primary); err == nil {
		t.Fatal("Deploy should propagate a compose up failure")
	}
}

// TestManagerDeployPodmanSecretError covers BOTH halves of the store write, because
// loading a secret is now two commands rather than one: `secret rm --ignore` then
// `secret create`. Either failing must stop the deploy and name the config key
// behind the secret, so the operator learns which env-file field to look at rather
// than which podman verb failed.
//
// The rm half is deliberately fatal. --ignore already absorbs the only benign case
// (nothing in the store yet), so a failure that survives it is real -- an unwritable
// store, say -- and creating a secret beside one that could not be removed would
// leave the store in a state nobody chose.
func TestManagerDeployPodmanSecretError(t *testing.T) {
	// "--ignore" appears only in the remove argv, "create" only in the create argv,
	// so each subtest fails exactly one half.
	for _, tc := range []struct{ name, failOnToken, version string }{
		{"remove half fails", "--ignore", "4.6.1\n"},
		{"create half fails", "create", "4.6.1\n"},
		{"create --replace fails", "create", "5.2.0\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ctrCfg(config.Podman, "false")
			cfg.Podman.QuadletDir = t.TempDir()
			m, rr, _ := newCapMgr(cfg, config.Podman)
			m.Geteuid = func() int { return -1 }
			versionAnswer(rr, tc.version)
			rr.fail = failOn(tc.failOnToken)
			err := m.Deploy(context.Background(), config.Primary)
			if err == nil {
				t.Fatalf("Deploy should propagate a failure of the %s", tc.name)
			}
			if !strings.Contains(err.Error(), "semp.adminPass") {
				t.Errorf("error should name the config key behind the secret, got: %v", err)
			}
		})
	}
}

// TestManagerRedeployUnchangedRestartsForRotation is --restart's remaining job on docker:
// with nothing changed, it forces the recreate that `up -d` would otherwise skip, for an
// operator who wants the container rebuilt anyway. A rotated password no longer needs it
// (TestManagerRedeployRotatedPasswordAsks), since the deploy digest notices that.
func TestManagerRedeployUnchangedRestartsForRotation(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	m, rr, buf := newCapMgr(cfg, config.Docker)
	m.RestartApproved = true
	m.Confirm = neverAsk(t)
	dockerHost(rr, true, deployDigest(t, cfg)) // running, from this very deploy
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if up := composeUp(t, rr); !eqArgs(up.args, composeUpArgs("--force-recreate")) {
		t.Errorf("--restart on an unchanged running container should force a recreate, got %v", up.args)
	}
	if !strings.Contains(buf.String(), "because --restart was given") {
		t.Errorf("Deploy should say why it recreated the container:\n%s", buf.String())
	}
}

// TestManagerRedeployPodmanUnchangedRestartsForRotation is the podman half: with the unit
// and every secret unchanged, --restart still restarts the running service, and says why.
func TestManagerRedeployPodmanUnchangedRestartsForRotation(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = dir
	if err := os.WriteFile(filepath.Join(dir, "sol-pod.container"),
		render.Quadlet(cfg, cfg.ResolveNode(config.Primary)), 0o600); err != nil {
		t.Fatal(err)
	}
	m, rr, buf := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	m.RestartApproved = true
	m.Confirm = neverAsk(t)
	podmanHost(rr, "active", map[string]string{"sol-pod-admin-password": "secret-pass"})
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if !hasCall(rr, "systemctl", withUser(cfg, "restart", "sol-pod.service")) {
		t.Errorf("--restart on an unchanged unit should restart the service:\n%+v", rr.calls)
	}
	if hasCall(rr, "systemctl", withUser(cfg, "daemon-reload")) {
		t.Errorf("nothing was written, so there is nothing to reload:\n%+v", rr.calls)
	}
	if !strings.Contains(buf.String(), "because --restart was given") {
		t.Errorf("Deploy should say why it restarted:\n%s", buf.String())
	}
}

// TestContainerRunningMatchesNameExactly guards containerRunningKnown's name matching:
// `ps --filter name=` is an unanchored regex on both engines, so a sibling deployment on
// the same host would otherwise be mistaken for this one. Getting it wrong in either
// direction is expensive -- a false positive skips the question for a live broker, a false
// negative asks nothing and lets compose recreate it.
//
// The listing is containercli.List's three-field rows, matched by containercli.HasName.
func TestContainerRunningMatchesNameExactly(t *testing.T) {
	row := func(name string) string { return name + "\tsolace/solace-pubsub-standard\tUp 1 hour\n" }
	cases := []struct {
		name    string
		listing string
		want    bool
	}{
		{"exact match", row("solace"), true},
		{"among others", row("other") + row("solace") + row("solace-edge"), true},
		{"sibling only", row("solace-edge"), false},
		{"prefix only", row("sol"), false},
		{"nothing running", "", false},
		{"padded name", "  solace  \tsolace/solace-pubsub-standard\tUp 1 hour\n", true},
		{"a short row is skipped", "solace\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ctrCfg(config.Docker, "false") // container.name is "solace"
			m, rr, _ := newCapMgr(cfg, config.Docker)
			rr.out = []byte(tc.listing)
			got, answered, err := m.containerRunningKnown(context.Background(), []string{"docker"})
			if err != nil || !answered {
				t.Fatalf("probe should have answered: answered=%v err=%v", answered, err)
			}
			if got != tc.want {
				t.Errorf("containerRunningKnown(%q) = %v, want %v", tc.listing, got, tc.want)
			}
			if !hasCall(rr, "docker", listArgs(true)) {
				t.Errorf("the probe should list running containers:\n%+v", rr.calls)
			}
		})
	}

	// A failed probe answers NEITHER way, and says so. The collapsed bool this used to
	// assert against is what made "not running" look like a fact.
	t.Run("probe fails", func(t *testing.T) {
		m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
		rr.out = []byte(row("solace"))
		rr.outErr = fmt.Errorf("engine unreachable")
		running, answered, err := m.containerRunningKnown(context.Background(), []string{"docker"})
		if running || answered || err == nil {
			t.Errorf("failed probe = (running=%v answered=%v err=%v), want (false, false, an error)",
				running, answered, err)
		}
	})
}

// TestManagerDeployDockerProbeUnansweredErrors is M4: a `ps` probe that fails
// answers neither "running" nor "not running", so deployDocker must not guess.
// Guessing "not running" would skip the question for a live broker that `up -d` is about
// to recreate, with none of the consent a changed deploy requires; guessing "running"
// would ask about a container that is not there. Erroring out is the only safe answer,
// and nothing may reach compose.
func TestManagerDeployDockerProbeUnansweredErrors(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	m, rr, _ := newCapMgr(cfg, config.Docker)
	rr.outFail = failOn("status=running")
	err := m.Deploy(context.Background(), config.Primary)
	if err == nil {
		t.Fatal("Deploy must fail when the running-probe cannot answer")
	}
	if !strings.Contains(err.Error(), "solace") {
		t.Errorf("error should name the container, got: %v", err)
	}
	if !strings.Contains(err.Error(), "cannot tell whether") {
		t.Errorf("error should say the probe could not answer, got: %v", err)
	}
	for _, c := range rr.calls {
		if containsStr(c.args, "up") {
			t.Errorf("an unanswered probe must not reach compose:\n%+v", rr.calls)
		}
	}
}

// TestManagerRedeployStoppedContainerIsLeftToCompose is what replaced "a stopped container
// is always recreated". The recreate was there because a plain `up -d` would START it with
// the credentials it was created with; the deploy digest label now makes compose recreate
// a stopped container whose file or secrets changed (its configuration differs), and an
// unchanged one is simply started. So the deploy runs a plain `up -d`, with no
// --force-recreate and no question: a stopped broker carries no traffic to drop.
func TestManagerRedeployStoppedContainerIsLeftToCompose(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	m, rr, _ := newCapMgr(cfg, config.Docker)
	m.Confirm = neverAsk(t)
	dockerHost(rr, false, "an-older-deploy") // ps lists nothing running
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if up := composeUp(t, rr); !eqArgs(up.args, composeUpArgs()) {
		t.Errorf("a stopped container is left to a plain `up -d`, got %v", up.args)
	}
}

// TestManagerDeployDockerRestartFlagOnAStoppedContainer: --restart on a stopped
// container is still a plain `up -d`: there is no running container to force.
func TestManagerDeployDockerRestartFlagOnAStoppedContainer(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	m, rr, _ := newCapMgr(cfg, config.Docker)
	m.RestartApproved = true
	dockerHost(rr, false, "")
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if up := composeUp(t, rr); !eqArgs(up.args, composeUpArgs()) {
		t.Errorf("--restart forces a recreate only of a running, unchanged container, got %v", up.args)
	}
}

// The podman "a failed stop is tolerated" case is now
// TestManagerDeletePodmanStopFailsServiceInactiveProceeds, which says the same
// thing with the fact B2 added: tolerated only once systemd has CONFIRMED the
// unit down. The unconditional version this replaced could not tell "already
// stopped" from "could not ask", which is the whole hazard.

func TestManagerDeletePodmanDaemonReloadError(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = t.TempDir()
	m, rr, _ := newCapMgr(cfg, config.Podman)
	rr.fail = failOn("daemon-reload")
	if err := m.Delete(context.Background(), false); err == nil {
		t.Fatal("Delete should propagate a daemon-reload failure")
	}
}

func TestManagerDeletePodmanRemoveUnitError(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = dir
	unit := filepath.Join(dir, "sol-pod.container")
	if err := os.Mkdir(unit, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(unit, "keep"), []byte("x"), 0o600); err != nil {
		t.Fatal(err) // non-empty dir -> os.Remove fails with a non-IsNotExist error
	}
	m, _, _ := newCapMgr(cfg, config.Podman)
	if err := m.Delete(context.Background(), false); err == nil {
		t.Fatal("Delete should fail when the quadlet unit cannot be removed")
	}
}

func TestManagerDeleteDockerComposeDownError(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	m, rr, _ := newCapMgr(cfg, config.Docker)
	rr.fail = failOn("down")
	if err := m.Delete(context.Background(), false); err == nil {
		t.Fatal("Delete should propagate a compose down failure")
	}
}

func TestManagerDeletePurgeError(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	m, rr, _ := newCapMgr(cfg, config.Docker)
	rr.fail = failOn("-mindepth") // targets only the purge find, not compose down
	if err := m.Delete(context.Background(), true); err == nil {
		t.Fatal("Delete --purge should propagate a data-dir rm failure")
	}
}

// --- Status permutations / tolerated non-zero -------------------------------

// --- euid guard (via the Geteuid seam) --------------------------------------

// TestGuardPodmanEUID is the ONE definition of the rootless/rootful invariant, the
// same function cli.prepare runs before any podman command and Deploy/PrepHost run
// again at the two sites where a wrong-account run does lasting damage. Rootless is
// the default (rootful false), and rootful is only ever deployed by root, on request.
// The passes are as load-bearing as the refusals: docker has no rootless mode in this
// schema, a preview reaches no host, and a platform with no POSIX euid still has to
// render.
func TestGuardPodmanEUID(t *testing.T) {
	cases := []struct {
		name     string
		platform config.Platform
		echo     bool
		rootful  bool
		euid     int
		wantErr  bool
	}{
		{"root without rootful is refused", config.Podman, false, false, 0, true},
		{"rootful as non-root is refused", config.Podman, false, true, 1000, true},
		{"rootless as non-root passes", config.Podman, false, false, 1000, false},
		{"rootful as root passes", config.Podman, false, true, 0, false},
		{"non-posix euid keeps the setting (rootless)", config.Podman, false, false, -1, false},
		{"non-posix euid keeps the setting (rootful)", config.Podman, false, true, -1, false},
		{"docker is never guarded", config.Docker, false, false, 0, false},
		{"docker as non-root too", config.Docker, false, true, 1000, false},
		{"echo runner reaches no host", config.Podman, true, true, 1000, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ctrCfg(tc.platform, "false")
			cfg.Podman.Rootful = tc.rootful
			var r runner.Runner = &capRunner{}
			if tc.echo {
				r = runner.Echo{W: io.Discard}
			}
			err := GuardPodmanEUID(cfg, tc.platform, r, tc.euid)
			if tc.wantErr != (err != nil) {
				t.Errorf("GuardPodmanEUID %s rootful=%t euid=%d: err=%v, wantErr=%t",
					tc.platform, tc.rootful, tc.euid, err, tc.wantErr)
			}
		})
	}
}

// TestGuardPodmanEUIDMessagesNameTheKeyAndTheFix: both refusals are read by an
// operator who has to decide between re-running differently and editing the file,
// so each names the key (podman.rootful, passed to the shared rule), the account it
// found, and the ways out.
func TestGuardPodmanEUIDMessagesNameTheKeyAndTheFix(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.Rootful = false
	err := GuardPodmanEUID(cfg, config.Podman, &capRunner{}, 0)
	for _, want := range []string{"running as root", "podman.rootful=true", "rootful podman is deployed only when asked for",
		"without sudo"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("the root-without-rootful refusal should name %q, got: %v", want, err)
		}
	}
	cfg.Podman.Rootful = true
	err = GuardPodmanEUID(cfg, config.Podman, &capRunner{}, 1000)
	for _, want := range []string{"podman.rootful=true", "requires root", "euid 1000", "sudo",
		"remove podman.rootful to deploy rootless"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("the rootful-as-non-root refusal should name %q, got: %v", want, err)
		}
	}
}

// TestManagerPrepHostRootlessAsRootFailsHard covers B6/M6: PrepHost used to only
// WARN when deploying rootless but the process is root, then went on to
// mkdir/chown under root's namespace mapping -- leaving a data directory the
// later rootless deploy cannot use. It now shares Deploy's checkPodmanEUID and
// must stop before either, so the invariant has one definition and one message.
func TestManagerPrepHostRootlessAsRootFailsHard(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.Rootful = false
	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return 0 }
	err := m.PrepHost(context.Background())
	if err == nil {
		t.Fatal("PrepHost must fail when podman.rootful is unset (rootless) but running as root")
	}
	if !strings.Contains(err.Error(), "running as root") {
		t.Errorf("error should name the invariant, got: %v", err)
	}
	if hasCall(rr, "mkdir", []string{"-p", "/opt/solace/data"}) {
		t.Errorf("PrepHost must not mkdir before the euid guard, got:\n%+v", rr.calls)
	}
	if hasCall(rr, "podman", []string{"unshare", "chown", "1000:0", "/opt/solace/data"}) {
		t.Errorf("PrepHost must not chown before the euid guard, got:\n%+v", rr.calls)
	}
}

// --- nil Log/Out sinks fall back to discard / os.Stdout ----------------------

func TestManagerNilSinks(t *testing.T) {
	m := NewManager(runner.Echo{}, ctrCfg(config.Docker, "false"), config.Docker, nil, nil)
	m.Resolve = func(string) bool { return true }
	if err := m.Check(context.Background()); err != nil {
		t.Fatalf("Check with nil Log/Out should not error: %v", err)
	}
}

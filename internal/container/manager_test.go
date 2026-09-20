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
	"strings"
	"testing"

	"solace/internal/config"
	"solace/internal/engine"
	"solace/internal/render"
)

// These tests exercise the container host Manager over two runners: engine.Echo
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
		c.Podman.Container.Name = "sol-pod"
		c.Podman.Container.RunUser = "1000:0" // config.RootlessRunUser, which ctrCfg skips deriving
		c.Podman.Container.Mem = "6898m"
		c.Podman.Container.DataDir = "/opt/solace/data"
		c.Podman.QuadletDir = "/etc/containers/systemd"
		// Set by hand for the same reason QuadletDir is: this fixture deliberately
		// skips ApplyDefaults, and baseDir has no default anyway (it is mandatory).
		// Leaving it empty would make ServerCertBundlePath resolve to the filesystem
		// root, and a test that wrote there would be worse than one that failed.
		c.Podman.BaseDir = "/opt/solace"
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

// newEchoMgr builds a Manager over engine.Echo, routing echoed commands, report
// output, and progress to one buffer. Resolve is stubbed (dry-run skips it).
func newEchoMgr(cfg *config.Config, p config.Platform) (*Manager, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	m := NewManager(engine.Echo{W: buf}, cfg, p, func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }, buf)
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
	cfg.Podman.Rootless = true
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

// TestManagerPrepHostCreatesBaseDir: podman.baseDir is created up front rather
// than by the first writeArtifact, and is deliberately NOT part of the 775/chown
// pair -- it holds the server-certificate bundle, which contains the private key.
func TestManagerPrepHostCreatesBaseDir(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.Rootless = true
	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return 1000 }
	rr.outFor = healthyRootlessOut(healthyNrOpen)
	if err := m.PrepHost(context.Background()); err != nil {
		t.Fatalf("PrepHost: %v", err)
	}
	base := cfg.Podman.BaseDir
	if !hasCall(rr, "mkdir", []string{"-p", base}) {
		t.Errorf("prep should create podman.baseDir:\n%+v", rr.calls)
	}
	if !hasCall(rr, "chmod", []string{"700", base}) {
		t.Errorf("baseDir must be 0700 explicitly -- mkdir -p leaves an existing mode alone:\n%+v", rr.calls)
	}
	if hasCall(rr, "chmod", []string{"775", base}) {
		t.Errorf("baseDir must not be group-writable, it holds the private key:\n%+v", rr.calls)
	}
	if hasCall(rr, "podman", []string{"unshare", "chown", "1000:0", base}) {
		t.Errorf("baseDir is read by the container through a bind mount, not owned by it:\n%+v", rr.calls)
	}
}

// TestManagerPrepHostDirectoryErrors covers the three ways preparing the two
// directories can fail. Each is its own return, and each names the directory and
// what it was trying to do -- an engine's bare "permission denied" says neither.
//
// baseDir is a path of its own here rather than the default, so failing on it
// cannot also match the dataDir underneath /opt/solace.
func TestManagerPrepHostDirectoryErrors(t *testing.T) {
	for _, tc := range []struct {
		name, on, want string
	}{
		{"baseDir mkdir", "/srv/solace-base", "create podman.baseDir"},
		{"baseDir mode", "700", "restrict podman.baseDir"},
		{"data dir mode", "775", "group-writable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ctrCfg(config.Podman, "false")
			cfg.Podman.Rootless = true
			cfg.Podman.BaseDir = "/srv/solace-base"
			m, rr, _ := newCapMgr(cfg, config.Podman)
			m.Geteuid = func() int { return 1000 }
			rr.outFor = healthyRootlessOut(healthyNrOpen)
			rr.fail = failOn(tc.on)
			err := m.PrepHost(context.Background())
			if err == nil {
				t.Fatalf("a failing %s must stop prep", tc.name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error should say what failed (%q), got: %v", tc.want, err)
			}
		})
	}
}

// TestManagerPrepHostDockerHasNoBaseDir: baseDir is a podman key, so the docker
// path must not invent one.
func TestManagerPrepHostDockerHasNoBaseDir(t *testing.T) {
	m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
	if err := m.PrepHost(context.Background()); err != nil {
		t.Fatalf("PrepHost: %v", err)
	}
	for _, c := range rr.calls {
		if c.name == "chmod" {
			t.Errorf("docker prep has no directory to chmod: %s %v", c.name, c.args)
		}
	}
}

// TestPrepHostRootlessDryRunSkipsTheReadinessBlock: the nofile probe is one row of
// the podman host-readiness block now (rootless.go), and the block skips as a whole
// under the Echo runner rather than row by row. Nothing it asserts can be answered
// without a real host, so one honest skip beats seven echoed probes and seven
// skipped assertions -- the same shape Preflight uses, one level up.
func TestPrepHostRootlessDryRunSkipsTheReadinessBlock(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.Rootless = true
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

func TestManagerDeployDockerComposeWritesFile(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if !fileExists(cfg.Docker.ComposeFile) {
		t.Error("Deploy should write the compose file")
	}
	// --force-recreate is part of the create path: a stopped container would
	// otherwise be started with the credentials it was created with.
	if !hasCall(rr, "docker", []string{"compose", "-f", cfg.Docker.ComposeFile, "up", "-d", "--force-recreate"}) {
		t.Errorf("Deploy should compose up:\n%+v", rr.calls)
	}
}

// TestManagerDockerComposeCommandOverride covers the standalone-binary form: a
// host without the compose plugin sets docker.compose, and every compose call has
// to go through it rather than the runtime.
func TestManagerDockerComposeCommandOverride(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
	cfg.Docker.Compose = config.Command{"docker-compose"}
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if !hasCall(rr, "docker-compose", []string{"-f", cfg.Docker.ComposeFile, "up", "-d", "--force-recreate"}) {
		t.Errorf("Deploy should use the configured compose command:\n%+v", rr.calls)
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
		if !hasCall(rr, "docker", []string{"inspect", "solace"}) {
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
		if !hasCall(rr, "podman", []string{"inspect", "sol-pod"}) {
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

// TestManagerRedeployUnchangedIsNoOp covers the both-platforms "nothing to do"
// arm: a re-deploy that renders the same artifact against a running broker must
// not touch it at all.
func TestManagerRedeployUnchangedIsNoOp(t *testing.T) {
	t.Run("podman", func(t *testing.T) {
		dir := t.TempDir()
		cfg := ctrCfg(config.Podman, "false")
		cfg.Podman.QuadletDir = dir
		m, rr, buf := newCapMgr(cfg, config.Podman)
		m.Geteuid = func() int { return -1 }
		rr.out = []byte("active\n") // the unit is already running
		if err := os.WriteFile(filepath.Join(dir, "sol-pod.container"),
			render.Quadlet(cfg, cfg.ResolveNode(config.Primary)), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		if hasCall(rr, "systemctl", withUser(cfg, "restart", "sol-pod.service")) {
			t.Errorf("an unchanged unit must not bounce the broker:\n%+v", rr.calls)
		}
		if hasCall(rr, "systemctl", withUser(cfg, "start", "sol-pod.service")) {
			t.Errorf("an active unit needs no start:\n%+v", rr.calls)
		}
		if !strings.Contains(buf.String(), "nothing to do") {
			t.Errorf("Deploy should report there was nothing to do:\n%s", buf.String())
		}
	})
	t.Run("docker", func(t *testing.T) {
		dir := t.TempDir()
		cfg := ctrCfg(config.Docker, "false")
		cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
		if err := os.WriteFile(cfg.Docker.ComposeFile,
			render.Compose(cfg, cfg.ResolveNode(config.Primary)), 0o600); err != nil {
			t.Fatal(err)
		}
		m, rr, buf := newCapMgr(cfg, config.Docker)
		rr.out = []byte("solace\n") // ps lists this container by name -> running
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		for _, c := range rr.calls {
			if containsStr(c.args, "up") {
				t.Errorf("an unchanged compose file must not recreate the container:\n%+v", rr.calls)
			}
		}
		if !strings.Contains(buf.String(), "nothing to do") {
			t.Errorf("Deploy should report there was nothing to do:\n%s", buf.String())
		}
	})
}

// TestManagerRedeployChangedNeedsConsent is the core of the upgrade fix: a
// changed artifact against a running broker must not bounce it silently, and
// --restart is what applies it. Podman's old behaviour was worse than silent --
// `systemctl start` on an active unit is a no-op, so the broker kept the old
// image while the command reported success.
func TestManagerRedeployChangedNeedsConsent(t *testing.T) {
	setup := func(t *testing.T, p config.Platform) (*config.Config, string) {
		t.Helper()
		dir := t.TempDir()
		cfg := ctrCfg(p, "false")
		if p == config.Podman {
			cfg.Podman.QuadletDir = dir
			// An artifact from an older image tag: the on-disk one differs from what
			// this config now renders.
			old := ctrCfg(p, "false")
			old.Image.Tag = "previous"
			if err := os.WriteFile(filepath.Join(dir, "sol-pod.container"),
				render.Quadlet(old, old.ResolveNode(config.Primary)), 0o600); err != nil {
				t.Fatal(err)
			}
			return cfg, "sol-pod.service"
		}
		cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
		old := ctrCfg(p, "false")
		old.Image.Tag = "previous"
		old.Docker.ComposeFile = cfg.Docker.ComposeFile
		if err := os.WriteFile(cfg.Docker.ComposeFile,
			render.Compose(old, old.ResolveNode(config.Primary)), 0o600); err != nil {
			t.Fatal(err)
		}
		return cfg, ""
	}

	t.Run("podman declines without consent", func(t *testing.T) {
		cfg, svc := setup(t, config.Podman)
		m, rr, buf := newCapMgr(cfg, config.Podman)
		m.Geteuid = func() int { return -1 }
		rr.out = []byte("active\n")
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		if hasCall(rr, "systemctl", withUser(cfg, "restart", svc)) {
			t.Errorf("no consent means no bounce:\n%+v", rr.calls)
		}
		if !strings.Contains(buf.String(), "still uses the previous one") {
			t.Errorf("Deploy should warn the running broker is now stale:\n%s", buf.String())
		}
		if !strings.Contains(buf.String(), "--restart") {
			t.Errorf("the warning should name the flag that finishes the job:\n%s", buf.String())
		}
	})
	t.Run("podman restarts with --restart", func(t *testing.T) {
		cfg, svc := setup(t, config.Podman)
		m, rr, _ := newCapMgr(cfg, config.Podman)
		m.Geteuid = func() int { return -1 }
		m.RestartApproved = true
		rr.out = []byte("active\n")
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		if !hasCall(rr, "systemctl", withUser(cfg, "restart", svc)) {
			t.Errorf("--restart should restart the active unit:\n%+v", rr.calls)
		}
	})
	t.Run("podman restarts when the prompt is accepted", func(t *testing.T) {
		cfg, svc := setup(t, config.Podman)
		m, rr, _ := newCapMgr(cfg, config.Podman)
		m.Geteuid = func() int { return -1 }
		asked := ""
		m.Confirm = func(q string) bool { asked = q; return true }
		rr.out = []byte("active\n")
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		if !strings.Contains(asked, svc) {
			t.Errorf("the prompt should name what gets restarted, got %q", asked)
		}
		if !hasCall(rr, "systemctl", withUser(cfg, "restart", svc)) {
			t.Errorf("an accepted prompt should restart:\n%+v", rr.calls)
		}
	})
	t.Run("docker declines without consent", func(t *testing.T) {
		cfg, _ := setup(t, config.Docker)
		m, rr, buf := newCapMgr(cfg, config.Docker)
		rr.out = []byte("solace\n")
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		for _, c := range rr.calls {
			if containsStr(c.args, "up") {
				t.Errorf("no consent means no recreate:\n%+v", rr.calls)
			}
		}
		if !strings.Contains(buf.String(), "still uses the previous one") {
			t.Errorf("Deploy should warn the running broker is now stale:\n%s", buf.String())
		}
	})
	t.Run("docker recreates with --restart", func(t *testing.T) {
		cfg, _ := setup(t, config.Docker)
		m, rr, _ := newCapMgr(cfg, config.Docker)
		m.RestartApproved = true
		rr.out = []byte("solace\n")
		if err := m.Deploy(context.Background(), config.Primary); err != nil {
			t.Fatalf("Deploy: %v", err)
		}
		if !hasCall(rr, "docker", []string{"compose", "-f", cfg.Docker.ComposeFile, "up", "-d"}) {
			t.Errorf("--restart should recreate through compose:\n%+v", rr.calls)
		}
	})
}

// --- secret externalization -------------------------------------------------

// TestManagerDeployDockerPassesSecretsAsEnv is the docker half of the secret
// model: nothing is written to this host, the values reach compose only through
// the child process environment, and the artifact carries variable names.
func TestManagerDeployDockerPassesSecretsAsEnv(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "true")
	cfg.Redundancy.PSK = "test-psk"
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if fileExists(filepath.Join(dir, "solace-secrets")) {
		t.Error("no secret file may be written any more; the values ride the compose process environment")
	}
	var up *capCall
	for i := range rr.calls {
		c := &rr.calls[i]
		if containsStr(c.args, "up") {
			up = c
		}
	}
	if up == nil {
		t.Fatalf("Deploy should run `compose up -d`:\n%+v", rr.calls)
	}
	if up.method != "RunEnv" {
		t.Errorf("compose must run through RunEnv so the values never reach an argv, got %s", up.method)
	}
	for _, want := range []string{"SOLACE_ADMIN_PASSWORD=secret-pass", "SOLACE_REDUNDANCY_PSK=test-psk"} {
		if !containsStr(up.env, want) {
			t.Errorf("compose environment should carry %q, got %v", want, maskedKeys(up.env))
		}
	}
	// The values may reach the environment and nothing else.
	for _, c := range rr.calls {
		for _, a := range c.args {
			for _, leak := range []string{"secret-pass", "test-psk"} {
				if strings.Contains(a, leak) {
					t.Errorf("secret value reached an argv: %s %v", c.name, c.args)
				}
			}
		}
	}
	body, err := os.ReadFile(cfg.Docker.ComposeFile)
	if err != nil {
		t.Fatal(err)
	}
	assertMode(t, cfg.Docker.ComposeFile, 0o600)
	for _, leak := range []string{"secret-pass", "test-psk"} {
		if strings.Contains(string(body), leak) {
			t.Errorf("compose file must reference secrets, not carry %q:\n%s", leak, body)
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
	} {
		if !strings.Contains(string(body), want) {
			t.Errorf("compose file should contain %q:\n%s", want, body)
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
func maskedKeys(env []string) string { return engine.MaskEnv(env) }

func TestManagerDeployPodmanCreatesSecrets(t *testing.T) {
	cfg := ctrCfg(config.Podman, "true")
	cfg.Redundancy.PSK = "test-psk"
	cfg.Podman.QuadletDir = t.TempDir()
	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	// Names carry the container name, so two brokers on one host cannot overwrite
	// each other's entries in the shared podman store.
	for _, name := range []string{"sol-pod-admin-password", "sol-pod-redundancy-psk"} {
		// Remove-then-create, not `create --replace`: --replace needs podman 4.7
		// while the rest of the wiring needs 4.5, and `rm --ignore` makes the pair
		// idempotent the same way.
		if !hasCall(rr, "podman", []string{"secret", "rm", "--ignore", name}) {
			t.Errorf("Deploy should remove any existing podman secret %s first:\n%+v", name, rr.calls)
		}
		if !hasCall(rr, "podman", []string{"secret", "create", name, "-"}) {
			t.Errorf("Deploy should create podman secret %s:\n%+v", name, rr.calls)
		}
	}
	// The floor is the point, so pin the flag's absence rather than only the new
	// shape's presence: this is what stops --replace creeping back in.
	for _, c := range rr.calls {
		if strings.Contains(strings.Join(c.args, " "), "--replace") {
			t.Errorf("no argv may carry --replace, which would raise the podman floor to 4.7: %v", c.args)
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
}

func TestManagerDeployRejectsEmptySecret(t *testing.T) {
	cfg := ctrCfg(config.Docker, "true") // HA -> the PSK secret is required too
	cfg.Redundancy.PSK = ""
	cfg.Docker.ComposeFile = filepath.Join(t.TempDir(), "compose.yml")
	m, _, _ := newCapMgr(cfg, config.Docker)
	err := m.Deploy(context.Background(), config.Primary)
	if err == nil {
		t.Fatal("Deploy must fail loud when a required secret is empty")
	}
	if !strings.Contains(err.Error(), "redundancy.psk") || !strings.Contains(err.Error(), "env file") {
		t.Errorf("error should name the field and the fix, got: %v", err)
	}
}

// TestManagerDeployDockerDryRunMasksSecretEnv covers the dry-run contract for the
// one path that carries values: the echo names the variables compose would be
// given and prints none of their values.
func TestManagerDeployDockerDryRunMasksSecretEnv(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "true")
	cfg.Redundancy.PSK = "" // empty is fine here: prep host has not generated it yet
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
	m, buf := newEchoMgr(cfg, config.Docker)
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("dry-run Deploy must stay previewable before prep host: %v", err)
	}
	if fileExists(filepath.Join(dir, "solace-secrets")) {
		t.Error("dry-run must not create anything, least of all a secrets dir")
	}
	out := buf.String()
	if !strings.Contains(out, "SOLACE_ADMIN_PASSWORD=***") {
		t.Errorf("dry-run should name the secret variables with masked values:\n%s", out)
	}
	if strings.Contains(out, "secret-pass") {
		t.Errorf("dry-run echo leaked the admin password:\n%s", out)
	}
}

func TestManagerDeployPodmanDryRunHidesSecretBytes(t *testing.T) {
	cfg := ctrCfg(config.Podman, "true")
	cfg.Redundancy.PSK = "test-psk"
	cfg.Podman.QuadletDir = t.TempDir()
	m, buf := newEchoMgr(cfg, config.Podman)
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

func TestManagerDeployPodmanWritesUnit(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = dir
	// Match this process's euid so the rootless/rootful guard passes on any host
	// (on Windows Geteuid()<0 and the guard is skipped entirely).
	if os.Geteuid() != 0 {
		cfg.Podman.Rootless = true
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
	cfg.Podman.Rootless = true
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

func TestManagerDeleteDockerComposeDownWhenFileExists(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
	if err := os.WriteFile(cfg.Docker.ComposeFile, []byte("services:\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.Delete(context.Background(), false); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !hasCall(rr, "docker", []string{"compose", "-f", cfg.Docker.ComposeFile, "down"}) {
		t.Errorf("Delete should compose down when the file exists:\n%+v", rr.calls)
	}
}

func TestManagerDeleteDockerPurgeClearsDataDir(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
	if err := os.WriteFile(cfg.Docker.ComposeFile, []byte("services:\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.Delete(context.Background(), true); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !hasCall(rr, "docker", []string{"compose", "-f", cfg.Docker.ComposeFile, "down"}) {
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

func TestManagerDeleteDockerComposeNoFileFallsBackToStopRm(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "absent.yml") // does not exist
	m, rr, _ := newCapMgr(cfg, config.Docker)
	rr.out = []byte("solace\n") // `ps -a` lists this container -> containerExists is true
	if err := m.Delete(context.Background(), false); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !hasCall(rr, "docker", []string{"stop", "solace"}) || !hasCall(rr, "docker", []string{"rm", "solace"}) {
		t.Errorf("compose Delete with no file should fall back to stop+rm:\n%+v", rr.calls)
	}
}

// TestManagerStopAndRemoveContainerAbsentNoOp covers M10: `docker rm` on a name
// that was never deployed exits non-zero, which used to turn "reset after a
// failed deploy" into an error -- every other removal path in the tool
// (--ignore-not-found) no-ops instead, and containerExists is what makes this
// fallback match them.
func TestManagerStopAndRemoveContainerAbsentNoOp(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "absent.yml") // no compose file -> stopAndRemove fallback
	m, rr, buf := newCapMgr(cfg, config.Docker)
	rr.out = []byte("nginx\n") // `ps -a` lists an unrelated container, not this one
	if err := m.Delete(context.Background(), false); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if hasCall(rr, "docker", []string{"stop", "solace"}) || hasCall(rr, "docker", []string{"rm", "solace"}) {
		t.Errorf("an absent container must not be stopped or removed:\n%+v", rr.calls)
	}
	if !strings.Contains(buf.String(), "nothing to remove") {
		t.Errorf("Delete should say there is nothing to remove:\n%s", buf.String())
	}
}

// TestManagerStopAndRemoveStopFailsContainerRunningBlocks covers B2's docker
// half: "docker info" (Preflight) proves the engine is reachable, not that the
// container actually stopped, so a failed stop must not fall through to `rm`
// while the broker is still up.
func TestManagerStopAndRemoveStopFailsContainerRunningBlocks(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "absent.yml")
	m, rr, _ := newCapMgr(cfg, config.Docker)
	rr.fail = failOn("stop")
	rr.out = []byte("solace\n") // seen by both containerExists (ps -a) and containerRunningKnown (ps --filter running)
	err := m.Delete(context.Background(), false)
	if err == nil {
		t.Fatal("Delete must fail when stop fails and the container is still running")
	}
	if !strings.Contains(err.Error(), "solace") {
		t.Errorf("error should name the container, got: %v", err)
	}
	if hasCall(rr, "docker", []string{"rm", "solace"}) {
		t.Errorf("rm must not run when stop failed and the container is still running:\n%+v", rr.calls)
	}
}

// TestManagerStopAndRemoveStopFailsProbeUnansweredBlocks is the docker twin of
// TestManagerDeletePodmanStopFailsStateUnknownBlocksRemoval: the stop failed and
// the running-probe could not answer either. Silence is not confirmation, so
// this must refuse rather than fall through to `rm` -- a bare running bool alone
// answers "false" here, which is the deferred M4 hazard and the reason
// stopAndRemove reads containerRunningKnown's second value instead.
func TestManagerStopAndRemoveStopFailsProbeUnansweredBlocks(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "absent.yml")
	m, rr, _ := newCapMgr(cfg, config.Docker)
	rr.fail = failOn("stop")
	rr.out = []byte("solace\n")           // `ps -a` still sees it, so the no-op branch is skipped
	rr.outFail = failOn("status=running") // ...but the running-probe itself fails
	err := m.Delete(context.Background(), false)
	if err == nil {
		t.Fatal("Delete must fail when stop fails and the running-probe cannot answer")
	}
	if !strings.Contains(err.Error(), "could not be confirmed stopped") {
		t.Errorf("error should say the container could not be confirmed stopped, got: %v", err)
	}
	if hasCall(rr, "docker", []string{"rm", "solace"}) {
		t.Errorf("rm must not run on an unconfirmed stop:\n%+v", rr.calls)
	}
}

// --- Lifecycle: Start / Stop / Restart ---------------------------------------

// TestManagerLifecyclePodmanSystemctl covers Start/Stop/Restart on podman: all
// three drive systemd rather than the engine directly, since the quadlet unit
// owns the container's lifecycle. Both rootful and rootless are covered because
// systemctlArgs prepends `--user` only in the rootless case.
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
		if len(args) > 0 && args[0] == "inspect" {
			return []byte(inspected)
		}
		return []byte("solace\tsolace/solace-pubsub-standard:10.10.1.128\tUp 3 days\n")
	}
	if err := m.StatusAll(context.Background(), true); err != nil {
		t.Fatalf("StatusAll --detail: %v", err)
	}
	if !hasCall(rr, "docker", []string{"inspect", "solace"}) {
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

// TestManagerLifecycleDockerComposeFile covers the docker half when a compose
// file is on disk: Start/Stop/Restart drive it through compose, the same
// artifact Deploy wrote, rather than the plain runtime verb.
func TestManagerLifecycleDockerComposeFile(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
	if err := os.WriteFile(cfg.Docker.ComposeFile, []byte("services:\n"), 0o600); err != nil {
		t.Fatal(err)
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
			m, rr, _ := newCapMgr(cfg, config.Docker)
			if err := tc.call(m); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if !hasCall(rr, "docker", []string{"compose", "-f", cfg.Docker.ComposeFile, tc.verb}) {
				t.Errorf("%s with a compose file present should run compose %s:\n%+v", tc.name, tc.verb, rr.calls)
			}
		})
	}
}

// TestManagerLifecycleDockerNoComposeFile covers the fallback when no compose
// file is on disk: Start/Stop/Restart drive the plain runtime verb against the
// container name instead, mirroring Delete's stop/rm fallback.
func TestManagerLifecycleDockerNoComposeFile(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "absent.yml") // does not exist
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
			m, rr, _ := newCapMgr(cfg, config.Docker)
			if err := tc.call(m); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if !hasCall(rr, "docker", []string{tc.verb, "solace"}) {
				t.Errorf("%s with no compose file should fall back to the plain runtime verb:\n%+v", tc.name, rr.calls)
			}
		})
	}
}

// TestManagerLifecycleDockerDryRunUsesCompose: under the Echo runner there is no file
// on disk to probe (the Echo runner never wrote one), so the preview always takes
// the compose branch rather than guessing from a real deploy's artifact.
func TestManagerLifecycleDockerDryRunUsesCompose(t *testing.T) {
	m, buf := newEchoMgr(ctrCfg(config.Docker, "false"), config.Docker)
	if err := m.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !strings.Contains(buf.String(), "+ docker compose -f docker-compose.yml start") {
		t.Errorf("dry-run Start should preview the compose path even with no file on disk:\n%s", buf.String())
	}
}

// --- Status / Logs / CLI / Shell --------------------------------------------

func TestManagerStatusPodman(t *testing.T) {
	m, rr, _ := newCapMgr(ctrCfg(config.Podman, "false"), config.Podman)
	if err := m.Status(context.Background()); err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !hasCall(rr, "systemctl", []string{"status", "sol-pod.service", "--no-pager"}) {
		t.Errorf("podman Status should show the unit:\n%+v", rr.calls)
	}
	if !hasCall(rr, "podman", []string{"ps", "--all", "--filter", "name=^sol-pod$", "--format", psTableFormat}) {
		t.Errorf("podman Status should ps the container:\n%+v", rr.calls)
	}
}

// TestManagerStatusDocker covers docker Status over the echo seam, which is now ONE
// listing: the narrowed `ps`. The format string is not spelled out here -- it carries real
// tab characters, and Echo shell-quotes what it prints, so a literal would be asserting
// Echo's quoting rather than this command's argv. The pieces that matter are.
func TestManagerStatusDocker(t *testing.T) {
	m, buf := newEchoMgr(ctrCfg(config.Docker, "false"), config.Docker)
	if err := m.Status(context.Background()); err != nil {
		t.Fatalf("Status: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"+ docker ps --all --filter 'name=^solace$'", "--format", "{{.Status}}"} {
		if !strings.Contains(out, want) {
			t.Errorf("docker Status missing %q:\n%s", want, out)
		}
	}
	// compose is gone, and so is the ports column it and the engine default both carried.
	for _, unwanted := range []string{"compose", "Ports"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("docker Status should no longer mention %q:\n%s", unwanted, out)
		}
	}
}

func TestManagerLogsCLIShell(t *testing.T) {
	t.Run("logs", func(t *testing.T) {
		m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
		if err := m.Logs(context.Background()); err != nil {
			t.Fatalf("Logs: %v", err)
		}
		got := rr.last()
		if got.method != "Run" || got.name != "docker" || !eqArgs(got.args, []string{"logs", "solace"}) {
			t.Errorf("Logs argv: %+v", got)
		}
	})
	t.Run("cli", func(t *testing.T) {
		m, rr, _ := newCapMgr(ctrCfg(config.Podman, "false"), config.Podman)
		if err := m.CLI(context.Background()); err != nil {
			t.Fatalf("CLI: %v", err)
		}
		got := rr.last()
		if got.method != "RunInteractive" || got.name != "podman" || !eqArgs(got.args, []string{"exec", "-it", "sol-pod", "cli", "-A"}) {
			t.Errorf("CLI argv: %+v", got)
		}
	})
	t.Run("shell", func(t *testing.T) {
		m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
		if err := m.Shell(context.Background()); err != nil {
			t.Fatalf("Shell: %v", err)
		}
		got := rr.last()
		if got.method != "RunInteractive" || got.name != "docker" || !eqArgs(got.args, []string{"exec", "-it", "solace", "bash"}) {
			t.Errorf("Shell argv: %+v", got)
		}
	})
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
	cfg.Podman.Rootless = true
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
	cfg.Podman.Rootless = true
	cfg.Podman.QuadletDir = t.TempDir()
	m, _, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return 0 } // rootless as root -> guard rejects
	if err := m.Deploy(context.Background(), config.Primary); err == nil {
		t.Fatal("Deploy should fail when the euid guard rejects the host")
	}
}

func TestManagerDeployDockerComposeWriteError(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	// A directory in the compose file's place makes WriteFile fail.
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
	if err := os.Mkdir(cfg.Docker.ComposeFile, 0o755); err != nil {
		t.Fatal(err)
	}
	m, _, _ := newCapMgr(cfg, config.Docker)
	if err := m.Deploy(context.Background(), config.Primary); err == nil {
		t.Fatal("Deploy should fail when the compose file cannot be written")
	}
}

func TestManagerDeployDockerComposeUpError(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
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
	for _, tc := range []struct{ name, failOnToken string }{
		{"remove half fails", "--ignore"},
		{"create half fails", "create"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ctrCfg(config.Podman, "false")
			cfg.Podman.QuadletDir = t.TempDir()
			m, rr, _ := newCapMgr(cfg, config.Podman)
			m.Geteuid = func() int { return -1 }
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

// TestManagerRedeployUnchangedRestartsForRotation covers the one way a rotated
// secret can reach a running broker: the artifact is unchanged (the value lives in
// the config and the environment, so there is nothing to diff), and --restart
// forces the recreate that `compose up -d` would otherwise skip.
func TestManagerRedeployUnchangedRestartsForRotation(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
	if err := os.WriteFile(cfg.Docker.ComposeFile,
		render.Compose(cfg, cfg.ResolveNode(config.Primary)), 0o600); err != nil {
		t.Fatal(err)
	}
	m, rr, buf := newCapMgr(cfg, config.Docker)
	m.RestartApproved = true
	rr.out = []byte("solace\n") // ps lists this container by name -> running
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	want := []string{"compose", "-f", cfg.Docker.ComposeFile, "up", "-d", "--force-recreate"}
	if !hasCall(rr, "docker", want) {
		t.Errorf("--restart on an unchanged compose file should force a recreate:\n%+v", rr.calls)
	}
	if !strings.Contains(buf.String(), "rotated secret") {
		t.Errorf("Deploy should say why it recreated the container:\n%s", buf.String())
	}
}

// TestManagerRedeployPodmanUnchangedRestartsForRotation is the podman half: the
// store secrets were just replaced, but a running container still holds the old
// values and the unit is byte-identical, so --restart is what applies them.
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
	rr.out = []byte("active\n")
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	if !hasCall(rr, "systemctl", withUser(cfg, "restart", "sol-pod.service")) {
		t.Errorf("--restart on an unchanged unit should restart the service:\n%+v", rr.calls)
	}
	if !strings.Contains(buf.String(), "rotated secret") {
		t.Errorf("Deploy should say why it restarted:\n%s", buf.String())
	}
}

// TestContainerRunningMatchesNameExactly guards containerRunningKnown's name matching:
// `ps --filter name=` is an unanchored regex on both engines, so a sibling deployment on
// the same host would otherwise be mistaken for this one. Getting it wrong in either
// direction is expensive -- a false positive skips the deploy, a false negative
// force-recreates a live broker without asking.
//
// It drove a containerRunning wrapper that collapsed the probe's three answers into one
// bool. That wrapper lost its last production caller when deployDocker began reading the
// answered flag (TestManagerDeployDockerProbeUnansweredErrors), so it has been deleted and
// this reads the same returns directly.
func TestContainerRunningMatchesNameExactly(t *testing.T) {
	cases := []struct {
		name    string
		listing string
		want    bool
	}{
		{"exact match", "solace\n", true},
		{"among others", "other\nsolace\nsolace-edge\n", true},
		{"sibling only", "solace-edge\n", false},
		{"prefix only", "sol\n", false},
		{"nothing running", "", false},
		{"padded listing", "  solace  \n", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ctrCfg(config.Docker, "false") // container.name is "solace"
			m, rr, _ := newCapMgr(cfg, config.Docker)
			rr.out = []byte(tc.listing)
			got, answered, err := m.containerRunningKnown(context.Background())
			if err != nil || !answered {
				t.Fatalf("probe should have answered: answered=%v err=%v", answered, err)
			}
			if got != tc.want {
				t.Errorf("containerRunningKnown(%q) = %v, want %v", tc.listing, got, tc.want)
			}
		})
	}

	// A failed probe answers NEITHER way, and says so. The collapsed bool this used to
	// assert against is what made "not running" look like a fact.
	t.Run("probe fails", func(t *testing.T) {
		m, rr, _ := newCapMgr(ctrCfg(config.Docker, "false"), config.Docker)
		rr.out = []byte("solace\n")
		rr.outErr = fmt.Errorf("engine unreachable")
		running, answered, err := m.containerRunningKnown(context.Background())
		if running || answered || err == nil {
			t.Errorf("failed probe = (running=%v answered=%v err=%v), want (false, false, an error)",
				running, answered, err)
		}
	})
}

// TestManagerDeployDockerProbeUnansweredErrors is M4: a `ps` probe that fails
// answers neither "running" nor "not running", so deployDocker must not guess.
// Guessing "not running" would force-recreate a live broker on a transient
// engine hiccup with none of the consent the artifact-changed branch requires
// for the identical action; guessing "running" would silently skip a create
// that should have run. Erroring out is the only safe answer, and neither
// create nor recreate may reach the runner.
func TestManagerDeployDockerProbeUnansweredErrors(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
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
			t.Errorf("an unanswered probe must not create or recreate the container:\n%+v", rr.calls)
		}
	}
}

// TestManagerRedeployStoppedContainerRecreates covers the arm no consent prompt
// guards: the compose file is unchanged and the container exists but is stopped.
// A plain `up -d` would START it, and a compose secret's value is baked in at
// creation -- so the broker would come back on the credentials it was created with
// and nothing would say so. Recreating is safe here (no traffic to drop) and is
// what makes a deploy's result honest.
func TestManagerRedeployStoppedContainerRecreates(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
	if err := os.WriteFile(cfg.Docker.ComposeFile,
		render.Compose(cfg, cfg.ResolveNode(config.Primary)), 0o600); err != nil {
		t.Fatal(err)
	}
	m, rr, _ := newCapMgr(cfg, config.Docker)
	rr.out = []byte("") // ps lists nothing running
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	want := []string{"compose", "-f", cfg.Docker.ComposeFile, "up", "-d", "--force-recreate"}
	if !hasCall(rr, "docker", want) {
		t.Errorf("a stopped container must be recreated, not started, so a rotated secret applies:\n%+v", rr.calls)
	}
}

// TestManagerRedeployUnchangedHintsRotation is the same state without --restart:
// nothing happens, and the operator is told how to apply a rotation.
func TestManagerRedeployUnchangedHintsRotation(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
	if err := os.WriteFile(cfg.Docker.ComposeFile,
		render.Compose(cfg, cfg.ResolveNode(config.Primary)), 0o600); err != nil {
		t.Fatal(err)
	}
	m, rr, buf := newCapMgr(cfg, config.Docker)
	rr.out = []byte("solace\n")
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	for _, c := range rr.calls {
		if containsStr(c.args, "--force-recreate") {
			t.Errorf("without --restart nothing may be recreated:\n%+v", rr.calls)
		}
	}
	if !strings.Contains(buf.String(), "re-run with --restart") {
		t.Errorf("Deploy should name the way to apply a rotated secret:\n%s", buf.String())
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
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
	if err := os.WriteFile(cfg.Docker.ComposeFile, []byte("services:\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m, rr, _ := newCapMgr(cfg, config.Docker)
	rr.fail = failOn("down")
	if err := m.Delete(context.Background(), false); err == nil {
		t.Fatal("Delete should propagate a compose down failure")
	}
}

// TestManagerDeleteDockerStopTolerated is the docker tolerate branch B2 kept: a
// failed `stop` that the engine then CONFIRMS is not running (the benign
// "already stopped" case) still warns and goes on to `rm`. The confirmation is
// what earns the tolerance -- an unanswered probe does not, which is
// TestManagerStopAndRemoveStopFailsProbeUnansweredBlocks. The two probes have to
// answer differently here, hence outFor: `ps --all` still lists the container so
// stop is attempted at all, `ps --filter status=running` does not so it counts
// as confirmed down.
func TestManagerDeleteDockerStopTolerated(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "absent.yml") // no file -> stop/rm fallback
	m, rr, buf := newCapMgr(cfg, config.Docker)
	rr.fail = failOn("stop")
	rr.outFor = func(_ string, args []string) []byte {
		if containsStr(args, "status=running") {
			return []byte("nginx\n") // answered, and this container is NOT in it
		}
		return []byte("solace\n") // `ps --all`: it exists, so stop is attempted
	}
	if err := m.Delete(context.Background(), false); err != nil {
		t.Fatalf("Delete should tolerate a stop failure once the container is confirmed down: %v", err)
	}
	if !strings.Contains(buf.String(), "stopping container") {
		t.Errorf("Delete should warn about the failed stop:\n%s", buf.String())
	}
	if !hasCall(rr, "docker", []string{"rm", "solace"}) {
		t.Errorf("a confirmed-down container should still be removed:\n%+v", rr.calls)
	}
}

func TestManagerDeletePurgeError(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
	if err := os.WriteFile(cfg.Docker.ComposeFile, []byte("services:\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m, rr, _ := newCapMgr(cfg, config.Docker)
	rr.fail = failOn("-mindepth") // targets only the purge find, not compose down
	if err := m.Delete(context.Background(), true); err == nil {
		t.Fatal("Delete --purge should propagate a data-dir rm failure")
	}
}

// --- Status permutations / tolerated non-zero -------------------------------

// TestManagerStatusDockerNoComposeFile covers the other branch of Status: with no
// compose file on disk there is nothing for `compose ps` to read, so only the
// container listing runs.
func TestManagerStatusDockerNoComposeFile(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "absent.yml")
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.Status(context.Background()); err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !hasCall(rr, "docker", []string{"ps", "--all", "--filter", "name=^solace$", "--format", psTableFormat}) {
		t.Errorf("Status should ps the container:\n%+v", rr.calls)
	}
	for _, c := range rr.calls {
		if len(c.args) > 0 && c.args[0] == "compose" {
			t.Errorf("Status must not call compose with no compose file:\n%+v", rr.calls)
		}
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
	if !hasCall(rr, "podman", []string{"ps", "--all", "--filter", "name=^sol-pod$", "--format", psTableFormat}) {
		t.Errorf("Status should still ps the container:\n%+v", rr.calls)
	}
}

// TestManagerStatusDockerRunsNoComposePs replaces the test that used to assert a
// tolerated `compose ps` failure. Status no longer runs compose at all, even WITH a
// compose file on disk: compose listed the same single container and its PORTS column --
// every published port with both host bindings -- was the widest thing in the report,
// with no way to narrow it, since compose's --format takes only `table` or `json`.
func TestManagerStatusDockerRunsNoComposePs(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
	if err := os.WriteFile(cfg.Docker.ComposeFile, []byte("services:\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.Status(context.Background()); err != nil {
		t.Fatalf("Status: %v", err)
	}
	for _, c := range rr.calls {
		if len(c.args) > 0 && c.args[0] == "compose" {
			t.Errorf("Status must not run compose, even with a compose file present:\n%+v", rr.calls)
		}
	}
	if !hasCall(rr, "docker", []string{"ps", "--all", "--filter", "name=^solace$", "--format", psTableFormat}) {
		t.Errorf("Status should ps the container:\n%+v", rr.calls)
	}
}

// TestStatusListingCarriesNoPortsColumn is the property the format exists for, asserted
// on the format itself rather than on engine output nothing here can produce. A broker
// publishes a dozen or more ports, so the engine's default table wraps every other column
// into illegibility -- on the one report whose job is to answer "is it up".
func TestStatusListingCarriesNoPortsColumn(t *testing.T) {
	if strings.Contains(psTableFormat, "Ports") {
		t.Errorf("psTableFormat must not ask for the ports column: %q", psTableFormat)
	}
	for _, want := range []string{"table ", "{{.Names}}", "{{.Image}}", "{{.Status}}"} {
		if !strings.Contains(psTableFormat, want) {
			t.Errorf("psTableFormat = %q, want it to contain %q", psTableFormat, want)
		}
	}
}

// --- euid guard (via the Geteuid seam) --------------------------------------

// TestGuardPodmanEUID is the ONE definition of the rootless/rootful invariant, the
// same function cli.prepare runs before any podman command and Deploy/PrepHost run
// again at the two sites where a wrong-account run does lasting damage. The three
// passes are as load-bearing as the two failures: docker has no rootless mode in
// this schema, a preview reaches no host, and a platform with no POSIX euid still
// has to render.
func TestGuardPodmanEUID(t *testing.T) {
	cases := []struct {
		name     string
		platform config.Platform
		echo     bool
		rootless bool
		euid     int
		wantErr  bool
	}{
		{"rootless as root fails", config.Podman, false, true, 0, true},
		{"rootful as non-root fails", config.Podman, false, false, 1000, true},
		{"rootless as non-root passes", config.Podman, false, true, 1000, false},
		{"rootful as root passes", config.Podman, false, false, 0, false},
		{"non-posix euid skips guard", config.Podman, false, false, -1, false},
		{"docker is never guarded", config.Docker, false, true, 0, false},
		{"docker as non-root too", config.Docker, false, false, 1000, false},
		{"echo runner reaches no host", config.Podman, true, false, 1000, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ctrCfg(tc.platform, "false")
			cfg.Podman.Rootless = tc.rootless
			var r engine.Runner = &capRunner{}
			if tc.echo {
				r = engine.Echo{W: io.Discard}
			}
			err := GuardPodmanEUID(cfg, tc.platform, r, tc.euid)
			if tc.wantErr != (err != nil) {
				t.Errorf("GuardPodmanEUID %s rootless=%t euid=%d: err=%v, wantErr=%t",
					tc.platform, tc.rootless, tc.euid, err, tc.wantErr)
			}
		})
	}
}

// TestGuardPodmanEUIDMessagesNameTheKeyAndTheFix: both refusals are read by an
// operator who has to decide between re-running differently and editing the file,
// so each names the key, the account it found, and both ways out.
func TestGuardPodmanEUIDMessagesNameTheKeyAndTheFix(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.Rootless = true
	err := GuardPodmanEUID(cfg, config.Podman, &capRunner{}, 0)
	for _, want := range []string{"podman.rootless=true", "running as root", "without sudo", "podman.rootless=false"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("the rootless-as-root refusal should name %q, got: %v", want, err)
		}
	}
	cfg.Podman.Rootless = false
	err = GuardPodmanEUID(cfg, config.Podman, &capRunner{}, 1000)
	for _, want := range []string{"podman.rootless=false", "requires root", "euid 1000", "sudo", "podman.rootless=true"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("the rootful-as-non-root refusal should name %q, got: %v", want, err)
		}
	}
}

// TestManagerPrepHostRootlessAsRootFailsHard covers B6/M6: PrepHost used to only
// WARN when podman.rootless=true but the process is root, then went on to
// mkdir/chown under root's namespace mapping -- leaving a data directory the
// later rootless deploy cannot use. It now shares Deploy's checkPodmanEUID and
// must stop before either, so the invariant has one definition and one message.
func TestManagerPrepHostRootlessAsRootFailsHard(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.Rootless = true
	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return 0 }
	err := m.PrepHost(context.Background())
	if err == nil {
		t.Fatal("PrepHost must fail when podman.rootless=true but running as root")
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
	m := NewManager(engine.Echo{}, ctrCfg(config.Docker, "false"), config.Docker, nil, nil)
	m.Resolve = func(string) bool { return true }
	if err := m.Check(context.Background()); err != nil {
		t.Fatalf("Check with nil Log/Out should not error: %v", err)
	}
}

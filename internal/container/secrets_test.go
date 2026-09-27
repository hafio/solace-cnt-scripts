package container

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"solace/internal/config"
	"solace/internal/output"
	"solace/internal/render"
)

// The server certificate is the one secret whose bytes are not in the env file, so
// it is the only one this package has to READ. These tests cover that read, the
// preview that keeps a dry-run possible before the files exist, podman's store entry
// and its change label, and the clean-up of the host file earlier builds wrote.

// certFixture points cfg at a real key and certificate on disk and returns the bundle
// the tool should produce: the key first, then the certificate.
func certFixture(t *testing.T, cfg *config.Config) string {
	t.Helper()
	dir := t.TempDir()
	key := filepath.Join(dir, "tls.key")
	crt := filepath.Join(dir, "tls.crt")
	const keyPEM = "-----BEGIN PRIVATE KEY-----\nKEYBYTES\n-----END PRIVATE KEY-----\n"
	const crtPEM = "-----BEGIN CERTIFICATE-----\nCERTBYTES\n-----END CERTIFICATE-----\n"
	if err := os.WriteFile(key, []byte(keyPEM), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(crt, []byte(crtPEM), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg.TLS.CertKey, cfg.TLS.Cert = key, crt
	return keyPEM + crtPEM
}

// TestResolveSecretValuesBuildsTheBundle covers the read itself: the value the
// resolver produces must be exactly what broker.ServerCertBundle would build, key
// before certificate, and every credential secret must be left alone.
func TestResolveSecretValuesBuildsTheBundle(t *testing.T) {
	cfg := ctrCfg(config.Docker, "true")
	want := certFixture(t, cfg)

	secrets, err := ResolveSecretValues(cfg, config.Docker, false)
	if err != nil {
		t.Fatalf("ResolveSecretValues: %v", err)
	}
	var found bool
	for _, s := range secrets {
		if len(s.SourceFiles) == 0 {
			continue
		}
		found = true
		if s.Value != want {
			t.Errorf("bundle = %q, want the key then the certificate (%q)", s.Value, want)
		}
	}
	if !found {
		t.Fatal("no file-backed secret was resolved; the certificate should be one on docker")
	}
	// Credential values must be exactly what the config holds -- unchanged, not
	// merely non-empty. An empty pre-shared key is legitimate before
	// `prepare host` has generated one, so "non-empty" would be the wrong assertion
	// and would fail for a reason that has nothing to do with the certificate.
	before := render.ContainerSecrets(cfg, config.Docker)
	for i, s := range secrets {
		if len(s.SourceFiles) > 0 {
			continue
		}
		if s.Value != before[i].Value {
			t.Errorf("credential secret %q value changed from %q to %q", s.Name, before[i].Value, s.Value)
		}
	}
}

// TestResolveSecretValuesPreviewReadsNothing is the property that keeps a dry-run
// usable before the certificate exists on this host. compose() runs on six verbs, so
// without it a preview of any of them would fail on a missing file -- exactly the
// previewability prepareSecrets already protects for the pre-shared key.
func TestResolveSecretValuesPreviewReadsNothing(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	// Paths that do not exist: a read would fail, a preview must not attempt one.
	cfg.TLS.CertKey = filepath.Join(t.TempDir(), "absent.key")
	cfg.TLS.Cert = filepath.Join(t.TempDir(), "absent.crt")

	secrets, err := ResolveSecretValues(cfg, config.Docker, true)
	if err != nil {
		t.Fatalf("a preview must not read the certificate: %v", err)
	}
	var checked bool
	for _, s := range secrets {
		if len(s.SourceFiles) == 0 {
			continue
		}
		checked = true
		// Non-empty on purpose: the entry must survive so the child environment keeps
		// one variable per secret, which is a deliberate invariant elsewhere.
		if s.Value == "" {
			t.Error("a previewed file-backed secret must keep a placeholder value, not an empty one")
		}
		if strings.Contains(s.Value, "PRIVATE KEY") {
			t.Error("a preview must not contain real key material")
		}
	}
	if !checked {
		t.Fatal("no file-backed secret present to preview")
	}
}

// TestResolveSecretValuesNamesTheUnreadableFile pins the error's usefulness: it has
// to name the file, the secret and the env-file key, because this failure now reaches
// the operator from six different verbs.
func TestResolveSecretValuesNamesTheUnreadableFile(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	certFixture(t, cfg)
	missing := filepath.Join(t.TempDir(), "gone.crt")
	cfg.TLS.Cert = missing

	_, err := ResolveSecretValues(cfg, config.Docker, false)
	if err == nil {
		t.Fatal("an unreadable certificate must fail loud")
	}
	for _, want := range []string{"gone.crt", "tls.cert"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q must mention %q", err, want)
		}
	}
}

// TestResolveSecretValuesRefusesAKeyBearingCert covers the already-built refusal from
// its NEW call sites. A tls.cert that already contains its private key would produce
// a bundle carrying the key twice, and after this change that refusal fires from
// inside compose() on docker and from the bundle write on podman.
func TestResolveSecretValuesRefusesAKeyBearingCert(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	certFixture(t, cfg)
	chained := filepath.Join(t.TempDir(), "chained.pem")
	body := "-----BEGIN PRIVATE KEY-----\nK\n-----END PRIVATE KEY-----\n" +
		"-----BEGIN CERTIFICATE-----\nC\n-----END CERTIFICATE-----\n"
	if err := os.WriteFile(chained, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg.TLS.Cert = chained
	if _, err := ResolveSecretValues(cfg, config.Docker, false); err == nil {
		t.Error("a cert file that already carries its key must be refused: the bundle would hold the key twice")
	}
}

// TestPodmanResolvesTheBundle: podman resolves the certificate exactly like docker
// does -- one file-backed secret whose value is the key then the certificate -- since
// both engines now deliver it as a secret. A podman list without it would leave the
// broker with no certificate.
func TestPodmanResolvesTheBundle(t *testing.T) {
	cfg := ctrCfg(config.Podman, "true")
	want := certFixture(t, cfg)
	secrets, err := ResolveSecretValues(cfg, config.Podman, false)
	if err != nil {
		t.Fatalf("ResolveSecretValues: %v", err)
	}
	var certs int
	for _, s := range secrets {
		if len(s.SourceFiles) == 0 {
			continue
		}
		certs++
		if s.Name != "sol-pod-tls-servercertificate" {
			t.Errorf("cert secret name = %q, want <container.name>-tls-servercertificate", s.Name)
		}
		if s.Value != want {
			t.Errorf("bundle = %q, want the key then the certificate (%q)", s.Value, want)
		}
	}
	if certs != 1 {
		t.Errorf("podman resolved %d file-backed secrets, want exactly 1", certs)
	}
}

// TestDeployDockerPassesTheBundleAsEnvNotArgv covers the docker delivery end to end:
// the bundle reaches the compose child through its environment, and never through an
// argument vector.
func TestDeployDockerPassesTheBundleAsEnvNotArgv(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	want := certFixture(t, cfg)
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")

	m, rr, _ := newCapMgr(cfg, config.Docker)
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	var sawBundle bool
	for _, c := range rr.calls {
		for _, e := range c.env {
			if strings.Contains(e, want) {
				sawBundle = true
			}
		}
		for _, a := range c.args {
			if strings.Contains(a, "PRIVATE KEY") {
				t.Errorf("the bundle reached an argv: %v", c.args)
			}
		}
	}
	if !sawBundle {
		t.Error("the compose child's environment must carry the bundle")
	}
	// And its digest, which fills the container label the compose file interpolates.
	digest := "SOLACE_TLS_SERVERCERTIFICATE_SHA256=" + bundleHash(want)
	var sawDigest bool
	for _, c := range rr.calls {
		sawDigest = sawDigest || slices.Contains(c.env, digest)
	}
	if !sawDigest {
		t.Errorf("the compose child's environment must carry %s:\n%+v", digest, rr.calls)
	}
	// Nothing of it is written beside the compose file, which is the property the
	// secret channel buys on docker.
	body, err := os.ReadFile(cfg.Docker.ComposeFile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "PRIVATE KEY") {
		t.Error("the compose file must reference the secret, not carry the key")
	}
}

// TestTeardownVerbsDoNotReadTheCertificate covers the operator's decision that a
// moved certificate must not block recovery. compose needs every declared secret to
// be DEFINED for down/stop/restart, but never reads it -- only `up` does -- so those
// verbs must keep working when the file is gone.
func TestTeardownVerbsDoNotReadTheCertificate(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	certFixture(t, cfg)
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
	if err := os.WriteFile(cfg.Docker.ComposeFile, []byte("services: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// The certificate moves away after deploy, which is the real-world case.
	cfg.TLS.Cert = filepath.Join(t.TempDir(), "moved.crt")

	for _, tc := range []struct {
		name string
		run  func(*Manager) error
	}{
		{"delete", func(m *Manager) error { return m.Delete(context.Background(), false) }},
		{"stop", func(m *Manager) error { return m.Stop(context.Background()) }},
		{"restart", func(m *Manager) error { return m.Restart(context.Background()) }},
	} {
		m, _, _ := newCapMgr(cfg, config.Docker)
		if err := tc.run(m); err != nil {
			t.Errorf("%s must not need the certificate's contents (only `up` does), got: %v", tc.name, err)
		}
	}

	// Deploy, which does need them, must still fail loud and name the file.
	m, _, _ := newCapMgr(cfg, config.Docker)
	err := m.Deploy(context.Background(), config.Primary)
	if err == nil {
		t.Fatal("deploy must fail on an unreadable certificate")
	}
	if !strings.Contains(err.Error(), "moved.crt") {
		t.Errorf("deploy error %q must name the file", err)
	}
}

// TestDeployDockerFailsBeforeWritingTheComposeFile pins "nothing happened" on the
// docker side. The certificate is read in prepareSecrets, before deployDocker rewrites
// the compose file, so a bad certificate leaves no artifact behind.
func TestDeployDockerFailsBeforeWritingTheComposeFile(t *testing.T) {
	dir := t.TempDir()
	cfg := ctrCfg(config.Docker, "false")
	certFixture(t, cfg)
	cfg.TLS.Cert = filepath.Join(t.TempDir(), "absent.crt")
	cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")

	m, _, _ := newCapMgr(cfg, config.Docker)
	if err := m.Deploy(context.Background(), config.Primary); err == nil {
		t.Fatal("Deploy must fail on an unreadable certificate")
	}
	if _, err := os.Stat(cfg.Docker.ComposeFile); !os.IsNotExist(err) {
		t.Error("no compose file may be written when the certificate cannot be read")
	}
}

// bundleHash is the render.CertDigestLabel value a deploy of bundle writes.
func bundleHash(bundle string) string {
	sum := sha256.Sum256([]byte(bundle))
	return hex.EncodeToString(sum[:])
}

// certCreate returns the captured `secret create` call for the certificate secret.
func certCreate(rr *capRunner, name string) (capCall, bool) {
	for _, c := range rr.calls {
		if c.method == "RunInput" && slices.Contains(c.args, "create") && slices.Contains(c.args, name) {
			return c, true
		}
	}
	return capCall{}, false
}

// TestDeployPodmanLoadsTheBundleIntoTheStore covers podman's whole delivery: the
// bundle goes into podman's secret store on stdin, labelled with its digest, the
// unit mounts it from there, and nothing lands on the host -- not even with a
// baseDir configured, which is where earlier builds wrote it.
func TestDeployPodmanLoadsTheBundleIntoTheStore(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	want := certFixture(t, cfg)
	cfg.Podman.QuadletDir = t.TempDir()
	cfg.Podman.BaseDir = t.TempDir()

	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}

	const name = "sol-pod-tls-servercertificate"
	c, ok := certCreate(rr, name)
	if !ok {
		t.Fatalf("no `secret create` for %s:\n%+v", name, rr.calls)
	}
	if c.stdin != want {
		t.Errorf("the secret's stdin = %q, want the key then the certificate", c.stdin)
	}
	wantArgs := []string{"secret", "create", "--label", "solace-util.sha256=" + bundleHash(want), name, "-"}
	if !eqArgs(c.args, wantArgs) {
		t.Errorf("create argv = %v, want %v", c.args, wantArgs)
	}
	// The rm that makes the pair idempotent comes before it, and the label is read
	// before the rm discards it.
	inspect := callIndex(rr, "podman", []string{"secret", "inspect", "--format",
		`{{index .Spec.Labels "solace-util.sha256"}}`, name})
	rm := callIndex(rr, "podman", []string{"secret", "rm", "--ignore", name})
	if inspect < 0 || rm < 0 || inspect > rm {
		t.Errorf("want the label inspected (%d) before the rm (%d):\n%+v", inspect, rm, rr.calls)
	}
	for _, call := range rr.calls {
		for _, a := range call.args {
			if strings.Contains(a, "PRIVATE KEY") {
				t.Errorf("the bundle reached an argv: %v", call.args)
			}
		}
	}

	unit, err := os.ReadFile(filepath.Join(cfg.Podman.QuadletDir, cfg.Podman.Container.Name+".container"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(unit), "Secret="+name+",type=mount,target=/mnt/certs/server/tls.pem\n") {
		t.Errorf("the unit must mount the certificate from the store:\n%s", unit)
	}
	if strings.Contains(string(unit), "PRIVATE KEY") {
		t.Error("the unit must reference the secret, not carry it")
	}
	if entries, err := os.ReadDir(cfg.Podman.BaseDir); err != nil || len(entries) != 0 {
		t.Errorf("nothing may be written to podman.baseDir any more: %v %v", entries, err)
	}
}

// TestDeployPodmanFailsBeforeWritingAnythingOnABadCert is podman's half of the
// "nothing happened" property: the bundle is read before the store is touched, so a
// failure loads no secret, leaves no unit and issues no systemctl call.
func TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	certFixture(t, cfg)
	cfg.TLS.Cert = filepath.Join(t.TempDir(), "absent.crt")
	cfg.Podman.QuadletDir = t.TempDir()

	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	if err := m.Deploy(context.Background(), config.Primary); err == nil {
		t.Fatal("Deploy must fail when the certificate cannot be read")
	}
	unit := filepath.Join(cfg.Podman.QuadletDir, cfg.Podman.Container.Name+".container")
	if _, err := os.Stat(unit); !os.IsNotExist(err) {
		t.Error("no quadlet unit may be written when the certificate cannot be read")
	}
	for _, c := range rr.calls {
		joined := strings.Join(c.args, " ")
		if c.name == "systemctl" || strings.Contains(joined, "systemctl") || strings.Contains(joined, "start") {
			t.Errorf("nothing should have been started: %v", c.args)
		}
		if strings.Contains(joined, "secret create") || strings.Contains(joined, "secret rm") {
			t.Errorf("the store must not be touched before the certificate is read: %v", c.args)
		}
	}
}

// TestCertSecretLabelDrivesTheRestart is why the label exists. The unit names the
// secret rather than its value, so a renewed certificate leaves the unit
// byte-identical; the digest label is the only thing that tells a redeploy the
// running broker is serving a stale certificate. The secret's name is derived, so the
// label decides alone: an equal one leaves the secret in place and is "nothing to do";
// a different or missing one -- a first deploy through this path, or a host an
// earlier build deployed -- overwrites it and asks to restart.
func TestCertSecretLabelDrivesTheRestart(t *testing.T) {
	for _, tc := range []struct {
		name      string
		label     func(want string) string
		fail      bool
		wantAsked bool
	}{
		{"same certificate", func(want string) string { return bundleHash(want) }, false, false},
		{"renewed certificate", func(string) string { return bundleHash("an older bundle") }, false, true},
		{"no label yet", func(string) string { return "" }, false, true},
		{"label unreadable", func(want string) string { return bundleHash(want) }, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ctrCfg(config.Podman, "false")
			want := certFixture(t, cfg)
			cfg.Podman.QuadletDir = t.TempDir()
			// The unit is already current, so the label is the only possible trigger.
			unit := filepath.Join(cfg.Podman.QuadletDir, cfg.Podman.Container.Name+".container")
			if err := os.WriteFile(unit, render.Quadlet(cfg, cfg.ResolveNode(config.Primary)), 0o600); err != nil {
				t.Fatal(err)
			}
			m, rr, _ := newCapMgr(cfg, config.Podman)
			m.Geteuid = func() int { return -1 }
			rr.outFor = func(_ string, args []string) []byte {
				switch {
				case slices.Contains(args, "is-active"):
					return []byte("active\n")
				case slices.Contains(args, "inspect"):
					return []byte(tc.label(want) + "\n")
				}
				return nil
			}
			if tc.fail {
				failOnCall(rr, func(_ string, args []string) bool {
					return slices.Contains(args, "secret") && slices.Contains(args, "inspect")
				})
			}
			var asked bool
			m.Confirm = func(string) bool { asked = true; return true }
			if err := m.Deploy(context.Background(), config.Primary); err != nil {
				t.Fatalf("Deploy: %v", err)
			}
			restarted := hasCall(rr, "systemctl", withUser(cfg, "restart", cfg.Podman.Container.Name+".service"))
			if asked != tc.wantAsked || restarted != tc.wantAsked {
				t.Errorf("asked=%t restarted=%t, want both %t:\n%+v", asked, restarted, tc.wantAsked, rr.calls)
			}
			const name = "sol-pod-tls-servercertificate"
			_, created := certCreate(rr, name)
			removed := hasCall(rr, "podman", []string{"secret", "rm", "--ignore", name})
			if created != tc.wantAsked || removed != tc.wantAsked {
				t.Errorf("overwrote the certificate secret: rm=%t create=%t, want both %t:\n%+v",
					removed, created, tc.wantAsked, rr.calls)
			}
			// The credentials are overwritten either way: they carry no label.
			if !hasCall(rr, "podman", []string{"secret", "rm", "--ignore", "sol-pod-admin-password"}) {
				t.Errorf("the admin password must still be reloaded:\n%+v", rr.calls)
			}
		})
	}
}

// TestDockerCertLabelDrivesTheRecreate is the docker counterpart. Compose keeps no
// secret object to label, so the digest rides the container as a label, filled from
// the compose child's environment; a redeploy reads it back from the running
// container. Under a compose file that is already current, an equal label is "nothing
// to do"; a different or missing one -- or an inspect that cannot answer -- asks
// before recreating, and recreates with --force-recreate on consent.
func TestDockerCertLabelDrivesTheRecreate(t *testing.T) {
	for _, tc := range []struct {
		name      string
		label     func(want string) string
		fail      bool
		wantAsked bool
	}{
		{"same certificate", func(want string) string { return bundleHash(want) }, false, false},
		{"renewed certificate", func(string) string { return bundleHash("an older bundle") }, false, true},
		{"no label yet", func(string) string { return "" }, false, true},
		{"inspect fails", func(string) string { return "" }, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			cfg := ctrCfg(config.Docker, "false")
			want := certFixture(t, cfg)
			cfg.Docker.ComposeFile = filepath.Join(dir, "compose.yml")
			if err := os.WriteFile(cfg.Docker.ComposeFile,
				render.Compose(cfg, cfg.ResolveNode(config.Primary)), 0o600); err != nil {
				t.Fatal(err)
			}
			m, rr, buf := newCapMgr(cfg, config.Docker)
			rr.outFor = func(_ string, args []string) []byte {
				switch {
				case slices.Contains(args, "ps"):
					return []byte("solace\n")
				case slices.Contains(args, "inspect"):
					return []byte(tc.label(want) + "\n")
				}
				return nil
			}
			if tc.fail {
				failOnCall(rr, func(_ string, args []string) bool { return slices.Contains(args, "inspect") })
			}
			var asked bool
			m.Confirm = func(string) bool { asked = true; return true }
			if err := m.Deploy(context.Background(), config.Primary); err != nil {
				t.Fatalf("Deploy: %v", err)
			}
			recreated := hasCall(rr, "docker", []string{"compose", "-f", cfg.Docker.ComposeFile, "up", "-d", "--force-recreate"})
			if asked != tc.wantAsked || recreated != tc.wantAsked {
				t.Errorf("asked=%t recreated=%t, want both %t:\n%+v", asked, recreated, tc.wantAsked, rr.calls)
			}
			if tc.wantAsked && !strings.Contains(buf.String(), "server certificate differs") {
				t.Errorf("the reason for the recreate should be reported:\n%s", buf.String())
			}
		})
	}
}

// TestUpdateServerCertSecretPersistsWithoutRestarting is the store half of `broker
// configure server-certs`, which never restarts anything. Podman overwrites only the
// certificate secret, under the deploy rule (a matching label leaves it, a different
// or missing one replaces it), so the next start mounts it; docker keeps no secret
// object and says so; no certificate configured touches nothing.
func TestUpdateServerCertSecretPersistsWithoutRestarting(t *testing.T) {
	const name = "sol-pod-tls-servercertificate"
	for _, tc := range []struct {
		name      string
		label     func(want string) string
		wantWrite bool
	}{
		{"same certificate", func(want string) string { return bundleHash(want) }, false},
		{"renewed certificate", func(string) string { return bundleHash("an older bundle") }, true},
		{"no label yet", func(string) string { return "" }, true},
	} {
		t.Run("podman/"+tc.name, func(t *testing.T) {
			cfg := ctrCfg(config.Podman, "false")
			want := certFixture(t, cfg)
			m, rr, _ := newCapMgr(cfg, config.Podman)
			rr.outFor = func(_ string, args []string) []byte {
				if slices.Contains(args, "inspect") {
					return []byte(tc.label(want) + "\n")
				}
				return nil
			}
			if err := m.UpdateServerCertSecret(context.Background()); err != nil {
				t.Fatalf("UpdateServerCertSecret: %v", err)
			}
			if _, created := certCreate(rr, name); created != tc.wantWrite {
				t.Errorf("certificate secret written = %t, want %t:\n%+v", created, tc.wantWrite, rr.calls)
			}
			for _, c := range rr.calls {
				joined := strings.Join(c.args, " ")
				if strings.Contains(joined, "admin-password") {
					t.Errorf("only the certificate secret may be touched: %v", c.args)
				}
				if strings.Contains(joined, "restart") || strings.Contains(joined, "start ") {
					t.Errorf("a hot-swap must never restart the broker: %v", c.args)
				}
			}
		})
	}
	t.Run("docker keeps no store", func(t *testing.T) {
		cfg := ctrCfg(config.Docker, "false")
		certFixture(t, cfg)
		m, rr, buf := newCapMgr(cfg, config.Docker)
		if err := m.UpdateServerCertSecret(context.Background()); err != nil {
			t.Fatalf("UpdateServerCertSecret: %v", err)
		}
		if len(rr.calls) != 0 {
			t.Errorf("docker has nothing to update, got %+v", rr.calls)
		}
		if !strings.Contains(buf.String(), "docker keeps no secret store") {
			t.Errorf("docker should say where its copy lives:\n%s", buf.String())
		}
	})
	t.Run("no certificate configured", func(t *testing.T) {
		m, rr, _ := newCapMgr(ctrCfg(config.Podman, "false"), config.Podman)
		if err := m.UpdateServerCertSecret(context.Background()); err != nil || len(rr.calls) != 0 {
			t.Errorf("err=%v calls=%+v, want nothing done", err, rr.calls)
		}
	})
	// Both failures stop before the store is touched, so the running broker and the
	// stored copy still agree.
	for _, tc := range []struct {
		name, want string
		setup      func(*config.Config, *capRunner)
	}{
		{"engine unreachable", "info", func(_ *config.Config, rr *capRunner) {
			failOnCall(rr, func(_ string, args []string) bool { return slices.Contains(args, "info") })
		}},
		{"certificate unreadable", "gone.crt", func(cfg *config.Config, _ *capRunner) {
			cfg.TLS.Cert = filepath.Join(filepath.Dir(cfg.TLS.Cert), "gone.crt")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ctrCfg(config.Podman, "false")
			certFixture(t, cfg)
			m, rr, _ := newCapMgr(cfg, config.Podman)
			tc.setup(cfg, rr)
			err := m.UpdateServerCertSecret(context.Background())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want a failure naming %q", err, tc.want)
			}
			for _, c := range rr.calls {
				if slices.Contains(c.args, "secret") {
					t.Errorf("the store must not be touched after a failure: %v", c.args)
				}
			}
		})
	}
}

// TestDockerCertChangedEdges covers the two answers TestDockerCertLabelDrivesTheRecreate
// cannot reach through Deploy, where prepareSecrets reads the certificate first: a
// preview compares nothing (no container exists to ask), and an unreadable certificate
// is an error naming the file rather than a guess either way.
func TestDockerCertChangedEdges(t *testing.T) {
	cfg := ctrCfg(config.Docker, "false")
	certFixture(t, cfg)
	echo, buf := newEchoMgr(cfg, config.Docker)
	if changed, err := echo.dockerCertChanged(context.Background()); changed || err != nil {
		t.Errorf("preview: changed=%t err=%v, want an unchanged answer", changed, err)
	}
	if buf.Len() != 0 {
		t.Errorf("a preview must not probe the container:\n%s", buf.String())
	}

	cfg.TLS.Cert = filepath.Join(filepath.Dir(cfg.TLS.Cert), "gone.crt")
	m, rr, _ := newCapMgr(cfg, config.Docker)
	if _, err := m.dockerCertChanged(context.Background()); err == nil || !strings.Contains(err.Error(), "gone.crt") {
		t.Errorf("err = %v, want a failure naming the unreadable file", err)
	}
	if len(rr.calls) != 0 {
		t.Errorf("nothing may be inspected once the certificate cannot be read: %+v", rr.calls)
	}
}

// TestCertSecretRemovalFailureIsFatal: the certificate secret holds the private key,
// so failing to remove it fails the teardown instead of warning like any other
// leftover secret -- and every other secret is still attempted first.
func TestCertSecretRemovalFailureIsFatal(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	certFixture(t, cfg)
	cfg.Podman.QuadletDir = t.TempDir()
	const name = "sol-pod-tls-servercertificate"

	m, rr, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	rr.fail = failOn(name)
	err := m.Delete(context.Background(), false)
	if err == nil {
		t.Fatal("a certificate secret that could not be removed must fail the teardown")
	}
	for _, want := range []string{name, "PRIVATE KEY", "podman secret rm " + name} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q must say %q", err, want)
		}
	}
	if !hasCall(rr, "podman", []string{"secret", "rm", "--ignore", "sol-pod-admin-password"}) {
		t.Errorf("the other secrets must still be removed:\n%+v", rr.calls)
	}
}

// writeLegacyBundle puts a file where earlier builds wrote podman's bundle.
func writeLegacyBundle(t *testing.T, cfg *config.Config) string {
	t.Helper()
	cfg.Podman.BaseDir = t.TempDir()
	path := filepath.FromSlash(render.ServerCertBundlePath(cfg))
	if err := os.WriteFile(path, []byte("-----BEGIN PRIVATE KEY-----\nOLD\n-----END PRIVATE KEY-----\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestLegacyBundleIsRemovedOnDeployAndRemove: a host an earlier build deployed holds
// the private key as <baseDir>/<name>-tls-servercertificate.pem. Upgrading must not
// leave it there, so deploy (once the new unit no longer names it) and remove both
// delete it -- with or without TLS still configured, since the file outlives that
// setting -- and leave the directory itself, which may hold other brokers' files.
func TestLegacyBundleIsRemovedOnDeployAndRemove(t *testing.T) {
	for _, tc := range []struct {
		name string
		tls  bool
		run  func(*Manager) error
	}{
		{"deploy", true, func(m *Manager) error { return m.Deploy(context.Background(), config.Primary) }},
		{"remove", true, func(m *Manager) error { return m.Delete(context.Background(), false) }},
		{"remove without tls", false, func(m *Manager) error { return m.Delete(context.Background(), false) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ctrCfg(config.Podman, "false")
			if tc.tls {
				certFixture(t, cfg)
			}
			cfg.Podman.QuadletDir = t.TempDir()
			path := writeLegacyBundle(t, cfg)
			m, _, buf := newCapMgr(cfg, config.Podman)
			m.Geteuid = func() int { return -1 }
			if err := tc.run(m); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Errorf("the legacy bundle holds a private key and must not survive %s", tc.name)
			}
			if _, err := os.Stat(cfg.Podman.BaseDir); err != nil {
				t.Errorf("podman.baseDir itself must be left in place: %v", err)
			}
			if !strings.Contains(buf.String(), "removed legacy server certificate bundle") {
				t.Errorf("the removal should be reported:\n%s", buf.String())
			}
		})
	}
}

// TestDeletePodmanToleratesAMissingBundle covers the normal case after the upgrade:
// no legacy file under a configured baseDir, or no baseDir at all.
func TestDeletePodmanToleratesAMissingBundle(t *testing.T) {
	for _, base := range []string{t.TempDir(), ""} {
		cfg := ctrCfg(config.Podman, "false")
		cfg.Podman.QuadletDir = t.TempDir()
		cfg.Podman.BaseDir = base
		certFixture(t, cfg)

		m, _, buf := newCapMgr(cfg, config.Podman)
		m.Geteuid = func() int { return -1 }
		if err := m.Delete(context.Background(), false); err != nil {
			t.Errorf("baseDir=%q: a missing legacy bundle must not fail the teardown: %v", base, err)
		}
		if strings.Contains(buf.String(), "legacy") {
			t.Errorf("baseDir=%q: nothing to remove, so nothing to say:\n%s", base, buf.String())
		}
	}
}

// TestLegacyBundleRemovalFailureIsFatal covers the delete's error path on both
// commands. A non-empty directory standing where the file belongs is the cheapest
// way to make os.Remove fail without depending on permissions, which differ across
// the OSes this suite runs on. Remove still attempts the store secrets.
func TestLegacyBundleRemovalFailureIsFatal(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(*Manager) error
	}{
		{"deploy", func(m *Manager) error { return m.Deploy(context.Background(), config.Primary) }},
		{"remove", func(m *Manager) error { return m.Delete(context.Background(), false) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ctrCfg(config.Podman, "false")
			certFixture(t, cfg)
			cfg.Podman.QuadletDir = t.TempDir()
			cfg.Podman.BaseDir = t.TempDir()
			path := filepath.FromSlash(render.ServerCertBundlePath(cfg))
			if err := os.MkdirAll(filepath.Join(path, "occupied"), 0o700); err != nil {
				t.Fatal(err)
			}
			m, rr, _ := newCapMgr(cfg, config.Podman)
			m.Geteuid = func() int { return -1 }
			err := tc.run(m)
			if err == nil {
				t.Fatalf("%s must fail when the legacy bundle cannot be removed", tc.name)
			}
			for _, want := range []string{"legacy server certificate bundle", "PRIVATE KEY", path} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q must say %q", err, want)
				}
			}
			if tc.name == "remove" &&
				!hasCall(rr, "podman", []string{"secret", "rm", "--ignore", "sol-pod-tls-servercertificate"}) {
				t.Errorf("remove must still clear the store secrets:\n%+v", rr.calls)
			}
		})
	}
}

// TestDryRunNeedsNoCertificateOnDisk is the previewability guarantee, one level up
// from the pre-shared key's: a preview must work before any certificate exists, show
// the store write with a placeholder rather than a digest, and write nothing.
func TestDryRunNeedsNoCertificateOnDisk(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = t.TempDir()
	cfg.Podman.BaseDir = t.TempDir()
	cfg.TLS.CertKey = filepath.Join(t.TempDir(), "absent.key")
	cfg.TLS.Cert = filepath.Join(t.TempDir(), "absent.crt")

	m, buf := newEchoMgr(cfg, config.Podman)
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("a dry-run must not need the certificate on disk: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "secret create --label 'solace-util.sha256=(preview)' sol-pod-tls-servercertificate -") {
		t.Errorf("the preview should show the certificate's store write, labelled with the placeholder:\n%s", out)
	}
	if strings.Contains(out, "PRIVATE KEY") || strings.Contains(out, "secret inspect") {
		t.Errorf("a preview must not print key material or probe the store:\n%s", out)
	}
	if entries, err := os.ReadDir(cfg.Podman.BaseDir); err != nil || len(entries) != 0 {
		t.Errorf("a dry-run must write nothing: %v %v", entries, err)
	}
}

// TestEchoPreviewsTheLegacyBundleRemoval covers the dry-run half of the clean-up on
// both commands: a preview says what it would remove and removes nothing. A dry-run
// that actually deleted a private key is the mistake worth a test of its own.
func TestEchoPreviewsTheLegacyBundleRemoval(t *testing.T) {
	for _, tc := range []struct {
		name string
		run  func(*Manager) error
	}{
		{"deploy", func(m *Manager) error { return m.Deploy(context.Background(), config.Primary) }},
		{"remove", func(m *Manager) error { return m.Delete(context.Background(), false) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ctrCfg(config.Podman, "false")
			certFixture(t, cfg)
			cfg.Podman.QuadletDir = t.TempDir()
			path := writeLegacyBundle(t, cfg)

			m, buf := newEchoMgr(cfg, config.Podman)
			if err := tc.run(m); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if !strings.Contains(buf.String(), "would remove legacy server certificate bundle") {
				t.Errorf("the preview should say it would remove the legacy bundle:\n%s", buf.String())
			}
			if _, err := os.Stat(path); err != nil {
				t.Errorf("a dry-run must not delete the legacy bundle: %v", err)
			}
		})
	}
}

// TestSecretSummaryNamesTheCertificateSource: the report reads no file, so the
// certificate's value is always empty there and set/MISSING would say MISSING for a
// perfectly good certificate. It names the two source keys instead, on both engines.
func TestSecretSummaryNamesTheCertificateSource(t *testing.T) {
	for _, p := range []config.Platform{config.Docker, config.Podman} {
		cfg := ctrCfg(p, "false")
		certFixture(t, cfg)
		got := secretSummary(p, render.ContainerSecrets(cfg, p))
		name := cfg.ContainerBlock(p).Name + "-tls-servercertificate"
		if !strings.Contains(got, name+"=(from tls.certKey + tls.cert)") {
			t.Errorf("%s: summary %q must name the certificate's sources", p, got)
		}
		if strings.Contains(got, name+"=MISSING") {
			t.Errorf("%s: summary %q reports a configured certificate as MISSING", p, got)
		}
		if !strings.Contains(got, cfg.ContainerBlock(p).Name+"-admin-password=set") {
			t.Errorf("%s: credentials keep set/MISSING: %q", p, got)
		}
	}
}

// TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv closes the round trip that had
// no check: the names come out of `<runtime> ps` on this host, not from the env file,
// and StatusAll puts each one straight back into an `inspect` argument vector.
//
// The row is still SHOWN -- it is a real container and hiding it would be worse -- but
// its name is not returned, so nothing further runs against it, and the skip is stated
// rather than silent. A missing row would read as "no such container".
func TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(t *testing.T) {
	buf := &bytes.Buffer{}
	sink := output.New(buf)

	const img = "solace/solace-pubsub-standard:latest"
	raw := strings.Join([]string{
		"good-broker\t" + img + "\tUp 2 hours",
		"also.good_1\t" + img + "\tUp 1 hour",
		"bad;name\t" + img + "\tUp 3 hours",
		"-leading-dash\t" + img + "\tUp 4 hours",
		"has space\t" + img + "\tUp 5 hours",
		"$(whoami)\t" + img + "\tUp 6 hours",
	}, "\n")

	got := solaceRows(sink, raw)

	want := []string{"good-broker", "also.good_1"}
	if len(got) != len(want) {
		t.Fatalf("returned names = %v, want only the engine-grammar ones %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("returned name[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	out := buf.String()
	// Every refused row is still visible, and each refusal is stated.
	for _, bad := range []string{"bad;name", "-leading-dash", "has space", "$(whoami)"} {
		if !strings.Contains(out, bad) {
			t.Errorf("the row for %q must still be shown; hiding a real container is worse:\n%s", bad, out)
		}
		if !strings.Contains(out, "not inspected") {
			t.Errorf("the skip must be stated, not silent:\n%s", out)
		}
	}
}

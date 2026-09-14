package container

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"solace/internal/config"
	"solace/internal/output"
	"solace/internal/render"
)

// The server certificate is the one secret whose bytes are not in the env file, so
// it is the only one this package has to READ. These tests cover that read, the
// preview that keeps a dry-run possible before the files exist, and the podman host
// file's whole lifecycle.

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

// TestPodmanResolvesNoFileBackedSecret is the proof of the split: podman's list has
// no file-backed entry, so no private key can reach podman's secret store.
func TestPodmanResolvesNoFileBackedSecret(t *testing.T) {
	cfg := ctrCfg(config.Podman, "true")
	certFixture(t, cfg)
	secrets, err := ResolveSecretValues(cfg, config.Podman, false)
	if err != nil {
		t.Fatalf("ResolveSecretValues: %v", err)
	}
	for _, s := range secrets {
		if len(s.SourceFiles) > 0 {
			t.Errorf("podman resolved a file-backed secret %q; it must reach podman as a bind-mounted file, "+
				"never as a store entry", s.Name)
		}
		if strings.Contains(s.Value, "PRIVATE KEY") {
			t.Errorf("secret %q carries key material into podman's store", s.Name)
		}
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

// TestDeployPodmanWritesTheBundleBeforeTheUnit covers podman's whole delivery: the
// file exists before anything starts, at a mode fit for a private key, and the unit
// bind-mounts exactly that path.
func TestDeployPodmanWritesTheBundleBeforeTheUnit(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	want := certFixture(t, cfg)
	cfg.Podman.QuadletDir = t.TempDir()
	cfg.Podman.BaseDir = t.TempDir()

	m, _, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}

	bundle := filepath.FromSlash(render.ServerCertBundlePath(cfg))
	got, err := os.ReadFile(bundle)
	if err != nil {
		t.Fatalf("the bundle must exist before the unit starts: %v", err)
	}
	if string(got) != want {
		t.Errorf("bundle = %q, want the key then the certificate", got)
	}
	// 0600 explicitly, because os.WriteFile only applies its mode when CREATING a
	// file: an existing bundle left at 0644 would otherwise be rewritten still 0644.
	if runtime.GOOS != "windows" {
		info, err := os.Stat(bundle)
		if err != nil {
			t.Fatal(err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("bundle mode = %#o, want 0600: it holds the broker's private key", perm)
		}
	}
	// The unit must name that same path, or podman creates a directory at a source
	// that does not exist and the broker starts with no certificate.
	unit, err := os.ReadFile(filepath.Join(cfg.Podman.QuadletDir, cfg.Podman.Container.Name+".container"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(unit), render.ServerCertBundlePath(cfg)) {
		t.Errorf("the quadlet unit must bind-mount %q:\n%s", render.ServerCertBundlePath(cfg), unit)
	}
	if strings.Contains(string(unit), "PRIVATE KEY") {
		t.Error("the unit must reference the bundle, not carry it")
	}
}

// TestDeployPodmanFailsBeforeWritingAnythingOnABadCert is podman's half of the
// "nothing happened" property: the bundle is built first, so a failure leaves no unit
// and issues no systemctl call.
func TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	certFixture(t, cfg)
	cfg.TLS.Cert = filepath.Join(t.TempDir(), "absent.crt")
	cfg.Podman.QuadletDir = t.TempDir()
	cfg.Podman.BaseDir = t.TempDir()

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
		if strings.Contains(strings.Join(c.args, " "), "systemctl") ||
			strings.Contains(strings.Join(c.args, " "), "start") {
			t.Errorf("nothing should have been started: %v", c.args)
		}
	}
}

// TestDeletePodmanRemovesTheBundleAfterTheUnit covers the teardown ordering and the
// severity choice. The bundle is something the running container consumed, so it goes
// after the unit is confirmed gone -- and its removal is FATAL rather than a warning,
// because silently leaving a private key on the host is the outcome least like the
// rest of this teardown.
func TestDeletePodmanRemovesTheBundleAfterTheUnit(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	certFixture(t, cfg)
	cfg.Podman.QuadletDir = t.TempDir()
	cfg.Podman.BaseDir = t.TempDir()

	m, _, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	if err := m.Deploy(context.Background(), config.Primary); err != nil {
		t.Fatalf("Deploy: %v", err)
	}
	bundle := filepath.FromSlash(render.ServerCertBundlePath(cfg))
	if _, err := os.Stat(bundle); err != nil {
		t.Fatalf("fixture problem: the bundle should exist after deploy: %v", err)
	}

	m2, _, _ := newCapMgr(cfg, config.Podman)
	m2.Geteuid = func() int { return -1 }
	if err := m2.Delete(context.Background(), false); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := os.Stat(bundle); !os.IsNotExist(err) {
		t.Error("the bundle holds a private key and must not survive a teardown")
	}
	// The directory may hold other brokers' bundles, so only the file goes.
	if _, err := os.Stat(cfg.Podman.BaseDir); err != nil {
		t.Errorf("podman.baseDir must be left in place: %v", err)
	}
}

// TestDeletePodmanToleratesAMissingBundle covers a teardown of a deployment that
// never had TLS, or whose tls.cert was unset since deploy.
func TestDeletePodmanToleratesAMissingBundle(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	cfg.Podman.QuadletDir = t.TempDir()
	cfg.Podman.BaseDir = t.TempDir()
	certFixture(t, cfg) // configured, but nothing was ever deployed

	m, _, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	if err := m.Delete(context.Background(), false); err != nil {
		t.Errorf("a missing bundle must not fail the teardown: %v", err)
	}
}

// TestDryRunNeedsNoCertificateOnDisk is the previewability guarantee, one level up
// from the pre-shared key's: a preview must work before any certificate exists.
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
	if !strings.Contains(out, "would write server certificate bundle") {
		t.Errorf("the preview should say it would write the bundle:\n%s", out)
	}
	if strings.Contains(out, "PRIVATE KEY") {
		t.Errorf("a preview must not print key material:\n%s", out)
	}
	bundle := filepath.FromSlash(render.ServerCertBundlePath(cfg))
	if _, err := os.Stat(bundle); !os.IsNotExist(err) {
		t.Error("a dry-run must write nothing")
	}
}

// TestEchoDeletePreviewsTheBundleRemoval covers the dry-run half of teardown: a
// preview must say what it would remove and remove nothing. Without this the only
// coverage of that branch would be the deploy-side preview, and a dry-run that
// actually deleted a private key is the mistake worth a test of its own.
func TestEchoDeletePreviewsTheBundleRemoval(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	certFixture(t, cfg)
	cfg.Podman.QuadletDir = t.TempDir()
	cfg.Podman.BaseDir = t.TempDir()

	// A real bundle on disk, so "removed nothing" is a meaningful claim.
	bundle := filepath.FromSlash(render.ServerCertBundlePath(cfg))
	if err := os.WriteFile(bundle, []byte("KEEP ME\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	m, buf := newEchoMgr(cfg, config.Podman)
	if err := m.Delete(context.Background(), false); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if !strings.Contains(buf.String(), "would remove server certificate bundle") {
		t.Errorf("the preview should say it would remove the bundle:\n%s", buf.String())
	}
	if _, err := os.Stat(bundle); err != nil {
		t.Errorf("a dry-run must not delete the bundle: %v", err)
	}
}

// TestDeployPodmanFailsWhenTheBundleCannotBeWritten covers the write's error path and
// pins that the message names the artifact rather than surfacing a bare filesystem
// error. A directory standing where the file belongs is the cheapest way to make the
// write fail without depending on permissions, which differ across the OSes this
// suite runs on.
func TestDeployPodmanFailsWhenTheBundleCannotBeWritten(t *testing.T) {
	cfg := ctrCfg(config.Podman, "false")
	certFixture(t, cfg)
	cfg.Podman.QuadletDir = t.TempDir()
	cfg.Podman.BaseDir = t.TempDir()

	// Occupy the bundle's own path with a directory.
	bundle := filepath.FromSlash(render.ServerCertBundlePath(cfg))
	if err := os.MkdirAll(bundle, 0o700); err != nil {
		t.Fatal(err)
	}

	m, _, _ := newCapMgr(cfg, config.Podman)
	m.Geteuid = func() int { return -1 }
	err := m.Deploy(context.Background(), config.Primary)
	if err == nil {
		t.Fatal("Deploy must fail when the bundle cannot be written")
	}
	if !strings.Contains(err.Error(), "server certificate bundle") {
		t.Errorf("error %q must name the artifact, not just the syscall", err)
	}
	// And nothing downstream ran: the unit is not written when the bundle is not.
	unit := filepath.Join(cfg.Podman.QuadletDir, cfg.Podman.Container.Name+".container")
	if _, err := os.Stat(unit); !os.IsNotExist(err) {
		t.Error("no quadlet unit may be written when the bundle write failed")
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

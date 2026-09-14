package render

import (
	"strconv"
	"strings"
	"testing"

	"solace/internal/config"
)

// The server certificate is the one secret whose delivery differs by engine, and the
// only one whose bytes are not in the env file. These tests cover that split; the
// goldens cover its rendered shape.

// TestServerCertFilePathKeyKeepsItsUnderscore guards a spelling that neither engine
// nor broker would complain about getting wrong.
//
// Every credential setting appends a BARE "filepath" (username_admin_password ->
// username_admin_passwordfilepath). The server certificate's does not: it carries an
// underscore. Deriving it from the generic suffix would produce
// tls_servercertificatefilepath, a setting the broker does not read -- and nothing
// errors on an unknown environment key, so TLS would simply be off with nothing
// pointing at the cause.
func TestServerCertFilePathKeyKeepsItsUnderscore(t *testing.T) {
	if certFilePathKey != "tls_servercertificate_filepath" {
		t.Errorf("certFilePathKey = %q, want the broker's exact spelling", certFilePathKey)
	}
	if certFilePathKey == "tls_servercertificate"+filePathSuffix {
		t.Error("certFilePathKey must NOT be derivable from filePathSuffix: that is the trap this constant " +
			"exists to close, and the resulting setting is one the broker silently ignores")
	}

	c := load(t, config.Docker)
	var seen int
	for _, p := range EnvPairs(c, c.ResolveNode(config.Primary)) {
		if p.Key == certFilePathKey {
			seen++
			if p.Value != certMount {
				t.Errorf("%s = %q, want the in-container mount %q", p.Key, p.Value, certMount)
			}
		}
		if p.Key == "tls_servercertificate"+filePathSuffix {
			t.Errorf("the misspelled setting %q reached the artifact", p.Key)
		}
	}
	if seen != 1 {
		t.Errorf("EnvPairs emitted the certificate setting %d times, want exactly 1", seen)
	}
}

// TestServerCertIsADockerOnlySecret pins the one place the secret list differs by
// platform, and why it has to.
//
// Docker can source a secret from the compose child's environment, so the bundle
// reaches the container with nothing written to this host. A quadlet unit cannot
// inline content, so podman gets the same bytes as a bind-mounted host file instead.
// If the certificate appeared in podman's list, CreatePodmanSecrets would load a
// PRIVATE KEY into podman's secret store.
func TestServerCertIsADockerOnlySecret(t *testing.T) {
	c := load(t, config.Docker)
	c.Redundancy.Enabled = "false"

	for _, s := range ContainerSecrets(c, config.Podman) {
		if len(s.SourceFiles) > 0 {
			t.Errorf("podman's secret list contains the file-backed secret %q: it would put the private key "+
				"into podman's secret store", s.Name)
		}
	}

	docker := ContainerSecrets(c, config.Docker)
	var certs []ContainerSecret
	for _, s := range docker {
		if len(s.SourceFiles) > 0 {
			certs = append(certs, s)
		}
	}
	if len(certs) != 1 {
		t.Fatalf("docker's secret list has %d file-backed secrets, want exactly 1", len(certs))
	}
	cert := certs[0]
	if docker[len(docker)-1].Name != cert.Name {
		t.Errorf("the certificate must be appended LAST (got %q last); prepending it would shift every "+
			"existing index and golden", docker[len(docker)-1].Name)
	}
	if cert.Target() != certMount || cert.MountPath() != certMount {
		t.Errorf("cert target/mount = %q/%q, want %q", cert.Target(), cert.MountPath(), certMount)
	}
	if cert.FilePathKey() != certFilePathKey {
		t.Errorf("cert setting = %q, want %q", cert.FilePathKey(), certFilePathKey)
	}
	if cert.Value != "" {
		t.Error("render must not populate the value: it is pure and may not read a file")
	}
	// The reported source must be the bundle that is actually built, in the same
	// order, or a report would name a different file from the one read.
	if len(cert.SourceFiles) != 2 || cert.SourceFiles[0] != c.TLS.CertKey || cert.SourceFiles[1] != c.TLS.Cert {
		t.Errorf("SourceFiles = %v, want {certKey, cert} in that order", cert.SourceFiles)
	}

	// With no certificate configured there is no such secret on either platform.
	c.TLS.Cert, c.TLS.CertKey = "", ""
	for _, p := range []config.Platform{config.Docker, config.Podman} {
		for _, s := range ContainerSecrets(c, p) {
			if len(s.SourceFiles) > 0 {
				t.Errorf("%s: no certificate is configured, so there must be no file-backed secret", p)
			}
		}
	}
}

// TestServerCertificateReachesTheContainerOnBothEngines is the named regression test
// for a property that previously had only goldens behind it. A golden break is
// routinely answered with -update, which would silently bless the certificate
// disappearing from an artifact altogether.
func TestServerCertificateReachesTheContainerOnBothEngines(t *testing.T) {
	c := load(t, config.Docker)
	id := c.ResolveNode(config.Primary)

	compose := string(Compose(c, id))
	if !strings.Contains(compose, "target: "+certMount) {
		t.Errorf("compose must mount the certificate at %q:\n%s", certMount, compose)
	}
	if !strings.Contains(compose, c.ContainerBlock(config.Docker).Name+certSuffix) {
		t.Error("compose must define the certificate secret at the top level")
	}
	if strings.Contains(compose, c.TLS.Cert) {
		t.Errorf("compose must NOT name the host certificate path any more: it rides the secret channel\n%s",
			compose)
	}

	quadlet := string(Quadlet(c, id))
	wantVol := "Volume=" + ServerCertBundlePath(c) + ":" + certMount + ":ro,Z"
	if !strings.Contains(quadlet, wantVol) {
		t.Errorf("quadlet must bind-mount the tool-written bundle (%q):\n%s", wantVol, quadlet)
	}
	if strings.Contains(quadlet, c.TLS.Cert) {
		t.Errorf("quadlet must NOT name the host certificate path: podman reads the bundle instead\n%s", quadlet)
	}

	// Both engines point the broker at the same in-container path.
	pairs := string(envLines(EnvPairs(c, id)))
	if !strings.Contains(pairs, certFilePathKey+"="+certMount) {
		t.Errorf("the broker setting must name %q:\n%s", certMount, pairs)
	}

	// The negative case: nothing mentions the mount when no certificate is set.
	c.TLS.Cert, c.TLS.CertKey = "", ""
	for name, body := range map[string]string{
		"compose": string(Compose(c, id)),
		"quadlet": string(Quadlet(c, id)),
	} {
		if strings.Contains(body, certMount) {
			t.Errorf("%s mentions %q with no certificate configured:\n%s", name, certMount, body)
		}
	}
}

// TestServerCertBundlePathIsPosixAndUnderBaseDir pins the one expression the quadlet
// renderer and the Manager's writer must share. If they drifted, podman would find
// no file at the source and create a DIRECTORY there instead, and the broker would
// start with no certificate and no error anywhere.
func TestServerCertBundlePathIsPosixAndUnderBaseDir(t *testing.T) {
	c := load(t, config.Podman)
	got := ServerCertBundlePath(c)
	if strings.ContainsRune(got, '\\') {
		t.Errorf("path %q must use forward slashes: it names a file on the Linux host podman runs on, "+
			"even when this tool is driven from Windows", got)
	}
	if !strings.HasPrefix(got, c.Podman.BaseDir+"/") {
		t.Errorf("path %q must sit directly under podman.baseDir %q", got, c.Podman.BaseDir)
	}
	if !strings.HasSuffix(got, ".pem") {
		t.Errorf("path %q should end in .pem", got)
	}
	// The container name is in the FILENAME, not just the directory, so two brokers
	// sharing one baseDir cannot overwrite each other's key.
	if !strings.Contains(got, c.Podman.Container.Name) {
		t.Errorf("path %q must carry the container name", got)
	}
	c2 := load(t, config.Podman)
	c2.Podman.Container.Name = "other-broker"
	if ServerCertBundlePath(c2) == got {
		t.Error("two brokers sharing a baseDir must not share a bundle path")
	}
	// A trailing separator on baseDir must not double up.
	c3 := load(t, config.Podman)
	c3.Podman.BaseDir = "/opt/solace/"
	if strings.Contains(ServerCertBundlePath(c3), "//") {
		t.Errorf("a trailing separator on baseDir produced %q", ServerCertBundlePath(c3))
	}
}

// TestFileBackedSecretIsExemptFromSecretPreflight covers a message that would
// otherwise be a lie. SecretPreflight blames an empty value with "set it in the env
// file", which for the certificate is wrong: tls.cert IS set, the bytes just live on
// the host. Readability is enforced where the files are read instead.
func TestFileBackedSecretIsExemptFromSecretPreflight(t *testing.T) {
	c := load(t, config.Docker)
	c.Redundancy.Enabled = "false"
	if err := SecretPreflight(c, config.Docker); err != nil {
		t.Errorf("a configured certificate must not trip the preflight (its value is not in the env file): %v", err)
	}
	// The exemption must not weaken the checks it exists beside.
	c.SEMP.AdminPass = ""
	if err := SecretPreflight(c, config.Docker); err == nil {
		t.Error("an empty admin.pass must still be refused")
	}
}

// TestPodmanNeverGetsAFileBackedSecret pins the structural guard the SecretScript
// refusal used to provide.
//
// SecretScript took its secret slice from the caller, so a docker-resolved list handed to
// the podman branch would have printed `podman secret create` for the server certificate
// and put a private key into podman's store. That renderer is gone, and the guarantee
// moved one level up and got stronger: ContainerSecrets never OFFERS a file-backed entry
// for podman at all. A refusal can be bypassed by a new caller; an absent entry cannot.
//
// The certificate still reaches a podman broker -- written to baseDir at 0600 and
// bind-mounted by the unit -- which is the whole reason it must not also be in the store,
// where teardown would try to delete it.
func TestPodmanNeverGetsAFileBackedSecret(t *testing.T) {
	c := load(t, config.Podman)
	// The docker list is the contrast: there the certificate IS a secret, so this proves
	// the fixture actually configures one and the podman check below is meaningful.
	var dockerHasFileBacked bool
	for _, s := range ContainerSecrets(c, config.Docker) {
		if len(s.SourceFiles) > 0 {
			dockerHasFileBacked = true
		}
	}
	if !dockerHasFileBacked {
		t.Fatal("fixture problem: the docker list should contain the file-backed certificate")
	}
	for _, s := range ContainerSecrets(c, config.Podman) {
		if len(s.SourceFiles) > 0 {
			t.Errorf("podman secret %q is file-backed; nothing that reads a host file may become "+
				"a podman store entry", s.Name)
		}
		if strings.HasSuffix(s.Name, certSuffix) {
			t.Errorf("podman secret %q is the server certificate; it is bind-mounted from baseDir, "+
				"never stored", s.Name)
		}
	}
}

// TestComposeEscapesTheDollarSign covers a value compose would otherwise consume
// before the container ever saw it.
//
// Compose interpolates `$VAR` and `${VAR}` across the whole document from its own
// environment, and turns `$$` back into one `$`. The case that makes this necessary
// rather than defensive is the health-check command: it is a shell command the ENGINE
// runs inside the broker, so `$(...)` and `$VAR` are legitimate there and the config
// gate permits them on purpose. Unescaped, an undefined variable would become empty
// and a defined one would be silently substituted.
func TestComposeEscapesTheDollarSign(t *testing.T) {
	if got := composeEscape("a$b${c}d"); got != "a$$b$${c}d" {
		t.Errorf("composeEscape = %q, want every '$' doubled", got)
	}
	if got := composeEscape("no dollars"); got != "no dollars" {
		t.Errorf("composeEscape must not disturb a value with no '$': %q", got)
	}

	c := load(t, config.Docker)
	c.Docker.Container.HealthCheck = config.HealthCheck{
		Enabled: true, Interval: "5s", Timeout: "5s", Retries: 3, StartPeriod: "60s",
		Cmd: []string{"sh", "-c", "curl -sf http://localhost:$PORT/health || exit 1"},
	}
	c.Image.Tag = modernTag
	c.Timezone = "Etc/GMT$0"

	body := string(Compose(c, c.ResolveNode(config.Primary)))
	if strings.Contains(body, "$PORT") && !strings.Contains(body, "$$PORT") {
		t.Errorf("the health-check command's '$' must be doubled, or compose eats it:\n%s", body)
	}
	if !strings.Contains(body, "$$PORT") {
		t.Errorf("expected the escaped form in the compose file:\n%s", body)
	}
	if !strings.Contains(body, "Etc/GMT$$0") {
		t.Errorf("an environment value's '$' must be doubled too:\n%s", body)
	}

	// The quadlet sink must NOT be escaped this way: systemd expands '%', not '$',
	// and quadletEscape already handles its own syntax.
	quadlet := string(Quadlet(c, c.ResolveNode(config.Primary)))
	if strings.Contains(quadlet, "$$PORT") {
		t.Errorf("the quadlet must not carry compose's escaping:\n%s", quadlet)
	}
}

// TestComposeQuotesIdentifierScalars covers a YAML 1.1 misread the name grammar cannot
// prevent. `yes`, `no`, `on`, `off`, `true` and `false` are all perfectly legal
// container names to both engines, and `0123` and `1.5` are too -- but bare in a compose
// document they are read as a boolean or a number, so the service key, container_name
// and hostname stop being the string the operator wrote.
func TestComposeQuotesIdentifierScalars(t *testing.T) {
	for _, name := range []string{"yes", "no", "on", "off", "true", "false", "0123", "1.5", "y"} {
		c := load(t, config.Docker)
		c.Docker.Container.Name = name
		body := string(Compose(c, c.ResolveNode(config.Primary)))

		for _, want := range []string{
			"  " + strconv.Quote(name) + ":\n",              // the service key
			"container_name: " + strconv.Quote(name) + "\n", // and the container name
		} {
			if !strings.Contains(body, want) {
				t.Errorf("container.name %q: compose must carry %q, or YAML reads it as a bool/number:\n%s",
					name, want, body)
			}
		}
		// The bare form must be gone from those positions.
		if strings.Contains(body, "container_name: "+name+"\n") {
			t.Errorf("container.name %q is still written bare, so compose will not read it as a string:\n%s",
				name, body)
		}
	}

	// The hostname comes from the node table and gets the same treatment.
	c := load(t, config.Docker)
	c.Redundancy.Primary.Name = "no"
	body := string(Compose(c, c.ResolveNode(config.Primary)))
	if !strings.Contains(body, `hostname: "no"`) {
		t.Errorf("hostname must be quoted for the same reason:\n%s", body)
	}
}

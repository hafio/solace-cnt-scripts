package render

import (
	"strconv"
	"strings"
	"testing"

	"solace/internal/config"
)

// The server certificate is the only secret whose bytes are not in the env file, and
// the only one mounted outside secretMount. These tests cover that on both engines; the
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

// TestServerCertIsASecretOnBothEngines pins the certificate as an ordinary engine
// secret on docker AND podman: one entry, appended last, mounted at certMount, read
// from {certKey, cert} in that order. Podman used to get a 0600 host file instead,
// which the default non-root broker could not read; an entry missing from podman's
// list would bring that back, or leave the broker with no certificate at all.
func TestServerCertIsASecretOnBothEngines(t *testing.T) {
	for _, p := range []config.Platform{config.Docker, config.Podman} {
		c := load(t, p)
		c.Redundancy.Enabled = "false"
		list := ContainerSecrets(c, p)
		var certs []ContainerSecret
		for _, s := range list {
			if len(s.SourceFiles) > 0 {
				certs = append(certs, s)
			}
		}
		if len(certs) != 1 {
			t.Fatalf("%s: the secret list has %d file-backed secrets, want exactly 1", p, len(certs))
		}
		cert := certs[0]
		if want := c.ContainerBlock(p).Name + certSuffix; cert.Name != want {
			t.Errorf("%s: cert secret name = %q, want %q (derived from container.name)", p, cert.Name, want)
		}
		if list[len(list)-1].Name != cert.Name {
			t.Errorf("%s: the certificate must be appended LAST (got %q last); prepending it would shift every "+
				"existing index and golden", p, list[len(list)-1].Name)
		}
		if cert.Target() != certMount || cert.MountPath() != certMount {
			t.Errorf("%s: cert target/mount = %q/%q, want %q", p, cert.Target(), cert.MountPath(), certMount)
		}
		if cert.FilePathKey() != certFilePathKey {
			t.Errorf("%s: cert setting = %q, want %q", p, cert.FilePathKey(), certFilePathKey)
		}
		if cert.Value != "" {
			t.Errorf("%s: render must not populate the value: it is pure and may not read a file", p)
		}
		// The reported source must be the bundle that is actually built, in the same
		// order, or a report would name a different file from the one read.
		if len(cert.SourceFiles) != 2 || cert.SourceFiles[0] != c.TLS.CertKey || cert.SourceFiles[1] != c.TLS.Cert {
			t.Errorf("%s: SourceFiles = %v, want {certKey, cert} in that order", p, cert.SourceFiles)
		}

		// With no certificate configured there is no such secret.
		c.TLS.Cert, c.TLS.CertKey = "", ""
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

	pc := load(t, config.Podman)
	quadlet := string(Quadlet(pc, pc.ResolveNode(config.Primary)))
	wantSecret := "Secret=" + pc.ContainerBlock(config.Podman).Name + certSuffix + ",type=mount,target=" + certMount + "\n"
	if !strings.Contains(quadlet, wantSecret) {
		t.Errorf("quadlet must mount the certificate from podman's store (%q):\n%s", wantSecret, quadlet)
	}
	// No host file at all: not the operator's certificate, and not the bundle an
	// earlier build wrote under baseDir -- which is 0600 and owned by container
	// root, so the default non-root broker could not read it.
	for _, host := range []string{pc.TLS.Cert, pc.TLS.CertKey, ServerCertBundlePath(pc)} {
		if strings.Contains(quadlet, host) {
			t.Errorf("quadlet must NOT name the host file %q: the certificate rides the secret store\n%s", host, quadlet)
		}
	}
	for _, line := range strings.Split(quadlet, "\n") {
		if strings.HasPrefix(line, "Volume=") && strings.Contains(line, certMount) {
			t.Errorf("quadlet still bind-mounts something at %q: %s", certMount, line)
		}
	}

	// Both engines point the broker at the same in-container path.
	pairs := string(envLines(EnvPairs(c, id)))
	if !strings.Contains(pairs, certFilePathKey+"="+certMount) {
		t.Errorf("the broker setting must name %q:\n%s", certMount, pairs)
	}

	// The negative case: nothing mentions the mount when no certificate is set.
	c.TLS.Cert, c.TLS.CertKey = "", ""
	pc.TLS.Cert, pc.TLS.CertKey = "", ""
	for name, body := range map[string]string{
		"compose": string(Compose(c, id)),
		"quadlet": string(Quadlet(pc, pc.ResolveNode(config.Primary))),
	} {
		if strings.Contains(body, certMount) {
			t.Errorf("%s mentions %q with no certificate configured:\n%s", name, certMount, body)
		}
	}
}

// TestServerCertBundlePathIsPosixAndUnderBaseDir pins the LEGACY path: where an
// earlier build wrote podman's bundle, and so the one file
// container.removeLegacyCertBundle deletes. If it drifted from what those builds
// wrote, an upgraded host would keep a private key on disk with nothing to say so.
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

// TestComposeLabelsTheCertificateDigest pins docker's half of the certificate's
// change detection. Compose keeps no secret object to label, so the digest is a
// container label, interpolated from the compose child's environment: the file must
// name the variable -- unescaped, since interpolation is the point -- and never a
// digest, so it stays byte-identical across a renewal. No certificate, no label.
func TestComposeLabelsTheCertificateDigest(t *testing.T) {
	c := load(t, config.Docker)
	var cert ContainerSecret
	for _, s := range ContainerSecrets(c, config.Docker) {
		if len(s.SourceFiles) > 0 {
			cert = s
		}
	}
	if want := cert.EnvVar() + "_SHA256"; cert.DigestEnvVar() != want {
		t.Errorf("DigestEnvVar = %q, want %q (EnvVar's name and fixed suffix, plus _SHA256)", cert.DigestEnvVar(), want)
	}
	compose := string(Compose(c, c.ResolveNode(config.Primary)))
	want := "    labels:\n      " + CertDigestLabel + ": \"${" + cert.DigestEnvVar() + "}\"\n"
	if !strings.Contains(compose, want) {
		t.Errorf("compose must label the container with the interpolated digest (%q):\n%s", want, compose)
	}
	if strings.Contains(compose, "$${"+cert.DigestEnvVar()) {
		t.Error("the digest variable must not be composeEscape'd, or compose would never fill it")
	}

	c.TLS.Cert, c.TLS.CertKey = "", ""
	compose = string(Compose(c, c.ResolveNode(config.Primary)))
	if strings.Contains(compose, "labels:") || strings.Contains(compose, "_SHA256") {
		t.Errorf("with no certificate there is no digest to label:\n%s", compose)
	}
}

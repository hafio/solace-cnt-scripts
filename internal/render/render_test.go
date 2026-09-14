package render

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"solace/internal/config"
)

// -update regenerates the testdata goldens from the current renderers. Run it
// only after eyeballing the diff -- the committed goldens are the reviewed
// expected output, so `go test ./internal/render` guards against regressions.
var update = flag.Bool("update", false, "regenerate golden files in testdata/")

// sampleFixture is the shared env file every golden renders from: the user
// template doubles as the golden fixture (one fixture, all three platforms).
const sampleFixture = "../../env/sample.yaml"

// goldenCertPath and goldenCertKeyPath replace whatever config.Load resolved
// tls.cert and tls.certKey to.
//
// They exist because Load now rebases a relative host path onto the env file's own
// directory, so the real values are <checkout>/env/certs/tls.crt -- machine-specific
// AND OS-specific, which a byte-compared golden can never hold. Without this the
// same test writes C:\Users\...\env\certs\tls.crt here and /home/runner/... in CI.
//
// Absolute and posix-shaped on purpose: absolute is what the rebase produces for
// every real deployment, and posix is the only shape that means anything in these
// artifacts, which are read by a container engine on Linux.
//
// Set UNCONDITIONALLY, like goldenPodmanBaseDir below. They used to be applied only when
// the fixture already set them, which stopped being safe once the sample shipped with TLS
// commented out: the goldens would then have quietly stopped covering the server
// certificate at all -- the quadlet `Volume=` mount, the compose secret, the CR tls block --
// and a golden that covers less still passes.
//
// The fixture supplying TLS is the right way round anyway. What these goldens exist to pin
// is what the RENDERERS emit for a TLS-configured broker, which should not depend on
// whether the shipped starter file happens to opt into it.
const (
	goldenCertPath    = "/etc/solace/certs/tls.crt"
	goldenCertKeyPath = "/etc/solace/certs/tls.key"

	// goldenPodmanBaseDir pins podman.baseDir, which reaches the quadlet as the
	// `Volume=` source of the server-certificate bundle. Set UNCONDITIONALLY, unlike
	// the two above: the sample declares its own value, and pinning it here is what
	// keeps the golden from depending on whatever an env file happens to say.
	goldenPodmanBaseDir = "/etc/solace"

	// goldenTLSSecret pins kubernetes.tlsServerSecret, whose presence is what puts the
	// tls block in the rendered CR.
	goldenTLSSecret = "solace-tls-secret"

	// goldenCertBundle stands in for the bytes broker.ServerCertBundle would read.
	// render is pure and cannot open a file, so the secret-script golden needs a
	// fixed fake -- and a fixed one is also what keeps that golden machine-independent
	// and free of a real private key.
	goldenCertBundle = "-----BEGIN PRIVATE KEY-----\nGOLDENKEY\n-----END PRIVATE KEY-----\n" +
		"-----BEGIN CERTIFICATE-----\nGOLDENCERT\n-----END CERTIFICATE-----\n"
)

// assertNoCheckoutPath fails when a rendered artifact carries the directory this
// repository is checked out into, under either separator spelling.
//
// It is the guard for a whole class of mistake rather than one field: a golden that
// held such a path would pass for whoever generated it and fail for everyone else,
// including CI, and the failure would name a line rather than the cause.
func assertNoCheckoutPath(t *testing.T, baseDir string, got []byte) {
	t.Helper()
	if baseDir == "" {
		t.Fatal("the render fixture has no base directory: config.Load should have derived one, so either " +
			"the rebase stopped running or the fixture stopped going through Load")
	}
	body := string(got)
	for _, spelling := range []string{baseDir, filepath.ToSlash(baseDir)} {
		if strings.Contains(body, spelling) {
			t.Errorf("the rendered artifact carries this checkout's own directory %q.\n"+
				"A golden cannot hold a machine-specific path. Pin the offending fixture field to a fixed "+
				"absolute value, the way goldenCertPath pins tls.cert.", spelling)
			return
		}
	}
}

func load(t *testing.T, p config.Platform) *config.Config {
	t.Helper()
	c, err := config.Load(sampleFixture, p)
	if err != nil {
		t.Fatalf("load %s under %s: %v", sampleFixture, p, err)
	}
	c.TLS.Cert = goldenCertPath
	c.TLS.CertKey = goldenCertKeyPath
	// The k8s CR emits its tls block off this name rather than off the file pair, so it
	// has to be set here too or the CR golden loses that section.
	c.K8s.TLSServerSecret = goldenTLSSecret
	c.Podman.BaseDir = goldenPodmanBaseDir
	return c
}

// healthCheckFixture enables the health check with no cmd, so the goldens show the
// built-in readiness probe, and with the timings ApplyDefaults fills. Shared by the
// quadlet and compose cases so the two goldens differ only in framing.
func healthCheckFixture() config.HealthCheck {
	return config.HealthCheck{
		Enabled:     true,
		Interval:    "5s",
		Timeout:     "5s",
		Retries:     3,
		StartPeriod: "60s",
	}
}

// modernTag is a broker release new enough for the built-in readiness probe, which
// config.Validate requires before the probe can be enabled. The health-check cases
// pin it rather than relying on the sample's tag: env/sample.yaml is a template
// whose tag is meant to be edited, and these two cases would render a config
// Validate rejects if it were ever set below 10.26. It is why those goldens carry a
// different image tag from every other one.
const modernTag = "10.26.0.5"

func envLines(pairs []EnvPair) []byte {
	var b strings.Builder
	for _, p := range pairs {
		b.WriteString(p.Assignment())
		b.WriteByte('\n')
	}
	return []byte(b.String())
}

func TestGolden(t *testing.T) {
	cases := []struct {
		name string
		file string
		gen  func(t *testing.T) []byte
	}{
		{
			name: "k8s broker CR",
			file: "k8s_broker_cr.golden",
			gen:  func(t *testing.T) []byte { return BrokerCR(load(t, config.K8s)) },
		},
		{
			// The sample leaves kubernetes.ports commented, so the case above renders the
			// 16 defaults. An explicit list here covers the other branch of
			// ApplyDefaults' port handling, and with it the two forms only a custom
			// list uses: a container port differing from the service port, and an
			// explicit protocol.
			name: "k8s broker CR with custom ports",
			file: "k8s_broker_cr_ports.golden",
			gen: func(t *testing.T) []byte {
				c := load(t, config.K8s)
				c.K8s.Ports = []string{"tcp-semp=8080", "tcp-smf=55555:55556", "tls-smf=55443/TCP"}
				return BrokerCR(c)
			},
		},
		{
			// Timezone and both security blocks are optional and omitted by
			// default, so the cases above prove they stay out of the CR. This one
			// proves they land correctly when set.
			name: "k8s broker CR with timezone and security context",
			file: "k8s_broker_cr_security.golden",
			gen: func(t *testing.T) []byte {
				c := load(t, config.K8s)
				c.Timezone = "Asia/Singapore"
				c.K8s.SecurityContext = config.PodSecurity{RunAsUser: "1000001", FSGroup: "1000002"}
				readOnly := false
				c.K8s.ContainerSecurity = config.ContainerSecurity{
					RunAsUser:              "1000001",
					RunAsGroup:             "1000002",
					ReadOnlyRootFilesystem: &readOnly,
				}
				return BrokerCR(c)
			},
		},
		{
			// The CR knobs that used to be hardcoded or inexpressible. pullPolicy
			// replaces a literal; podAnnotations/podLabels are new optional blocks,
			// so the cases above prove they stay out when unset. The values here
			// carry a colon and a quote to exercise the escaping.
			name: "k8s broker CR with pull policy and pod metadata",
			file: "k8s_broker_cr_podmeta.golden",
			gen: func(t *testing.T) []byte {
				c := load(t, config.K8s)
				c.K8s.ImagePullPolicy = "Always"
				c.K8s.PodAnnotations = map[string]string{
					"prometheus.io/scrape": "true",
					"example.com/note":     `a: "quoted" value`,
				}
				c.K8s.PodLabels = map[string]string{"example.com/tier": "messaging"}
				return BrokerCR(c)
			},
		},
		{
			// The additive affinity blocks alongside the legacy anti-affinity term:
			// the fixed broker-spread term must still come first, unchanged, with the
			// configured terms after it.
			name: "k8s broker CR with node and pod affinity",
			file: "k8s_broker_cr_affinity.golden",
			gen: func(t *testing.T) []byte {
				c := load(t, config.K8s)
				c.K8s.Placement.NodeAffinity = config.NodeAffinity{
					Preferred: []config.WeightedNodeTerm{{
						Weight: 80,
						Match: []config.NodeMatchExpr{
							{Key: "topology.kubernetes.io/zone", Operator: "In", Values: []string{"az-1", "az-2"}},
						},
					}},
					Required: []config.NodeMatchExpr{
						{Key: "solace.com/broker", Operator: "Exists"},
						{Key: "kubernetes.io/arch", Operator: "NotIn", Values: []string{"arm64"}},
					},
				}
				c.K8s.Placement.PodAffinity = []config.PodAffinityTerm{{
					Weight:      20,
					TopologyKey: "topology.kubernetes.io/zone",
					MatchLabels: map[string]string{"app": "gateway"},
					Namespaces:  []string{"edge"},
				}}
				c.K8s.Placement.PodAntiAffinity = []config.PodAffinityTerm{{
					TopologyKey: "topology.kubernetes.io/zone",
					MatchLabels: map[string]string{"app.kubernetes.io/name": "pubsubpluseventbroker"},
				}}
				return BrokerCR(c)
			},
		},
		{
			// loadBalancer.annotations and placement.labels* are user-supplied
			// "key: value" fragments that used to be pasted into the manifest
			// verbatim. The values here carry a colon, a slash and a URL, which only
			// stay intact because both halves are quoted. This is also the only case
			// covering nodeSelector and tolerations at all.
			name: "k8s broker CR with annotations and node labels",
			file: "k8s_broker_cr_labels.golden",
			gen: func(t *testing.T) []byte {
				c := load(t, config.K8s)
				c.K8s.LoadBalancer.Annotations = []string{
					"external-dns.alpha.kubernetes.io/hostname: broker.example.com",
					"service.beta.kubernetes.io/target: https://lb.example.com:8443",
				}
				c.K8s.Placement.LabelsPrimary = []string{"nodetype: solace"}
				c.K8s.Placement.LabelsBackup = []string{"nodetype: solace"}
				c.K8s.Placement.LabelsMonitor = []string{"nodetype: solace"}
				c.K8s.Placement.TolerationsPrimary = []string{"dedicated=solace:NoSchedule"}
				c.K8s.Placement.TolerationsMonitor = []string{"dedicated:NoExecute"}
				return BrokerCR(c)
			},
		},
		{
			name: "podman quadlet primary",
			file: "podman_quadlet_primary.golden",
			gen: func(t *testing.T) []byte {
				c := load(t, config.Podman)
				return Quadlet(c, c.ResolveNode(config.Primary))
			},
		},
		{
			name: "docker compose primary",
			file: "docker_compose_primary.golden",
			gen: func(t *testing.T) []byte {
				c := load(t, config.Docker)
				return Compose(c, c.ResolveNode(config.Primary))
			},
		},
		{
			// The health check is opt-in, so every other container case proves the
			// artifacts stay unchanged without it; this one proves both framings.
			name: "podman quadlet with health check",
			file: "podman_quadlet_healthcheck.golden",
			gen: func(t *testing.T) []byte {
				c := load(t, config.Podman)
				c.Image.Tag = modernTag
				c.Podman.Container.HealthCheck = healthCheckFixture()
				return Quadlet(c, c.ResolveNode(config.Primary))
			},
		},
		{
			name: "docker compose with health check",
			file: "docker_compose_healthcheck.golden",
			gen: func(t *testing.T) []byte {
				c := load(t, config.Docker)
				c.Image.Tag = modernTag
				c.Docker.Container.HealthCheck = healthCheckFixture()
				return Compose(c, c.ResolveNode(config.Primary))
			},
		},
		{
			// Standalone drops the whole redundancy block from the compose file,
			// including its secret reference -- one secret instead of two.
			name: "docker compose standalone",
			file: "docker_compose_standalone.golden",
			gen: func(t *testing.T) []byte {
				c := load(t, config.Docker)
				c.Redundancy.Enabled = "false"
				return Compose(c, c.ResolveNode(config.Primary))
			},
		},
		{
			// The sample sets no tz, so this covers the omitted-TZ branch; the
			// standalone case below sets one and covers the other.
			name: "container env-pairs primary HA",
			file: "container_envpairs_primary.golden",
			gen: func(t *testing.T) []byte {
				c := load(t, config.Podman)
				return envLines(EnvPairs(c, c.ResolveNode(config.Primary)))
			},
		},
		{
			name: "container env-pairs standalone",
			file: "container_envpairs_standalone.golden",
			gen: func(t *testing.T) []byte {
				c := load(t, config.Podman)
				// Exercise the standalone and TZ-present branches without a second fixture.
				c.Redundancy.Enabled = "false"
				c.Timezone = "Asia/Singapore"
				return envLines(EnvPairs(c, c.ResolveNode(config.Primary)))
			},
		},
		// The two secret-script goldens are gone with SecretScript. No container
		// artifact carries a secret VALUE now: `broker generate` renders the compose
		// file or the quadlet and stops there, and `broker deploy` is the only thing
		// that materialises the values -- podman into its secret store, docker into
		// the compose child's environment. There is no longer a rendering to pin.
	}

	// The directory config.Load rebased this fixture's relative paths onto, resolved
	// once rather than per case.
	fixtureBase := load(t, config.Docker).BaseDir()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.gen(t)
			// Before it can be written OR compared: no artifact may carry the
			// directory this checkout happens to live in. config.Load rebases every
			// relative host path onto the env file's own directory, so any fixture
			// field that reaches an artifact and is not pinned to a fixed absolute
			// value (goldenCertPath and its sibling) leaks a machine-specific path
			// into a byte-compared file. Checked here rather than in one case,
			// because the next such field will arrive in a different one.
			assertNoCheckoutPath(t, fixtureBase, got)
			golden := filepath.Join("testdata", tc.file)
			if *update {
				if err := os.WriteFile(golden, got, 0o644); err != nil {
					t.Fatalf("write golden %s: %v", golden, err)
				}
				return
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden %s (regenerate: go test ./internal/render -update): %v", golden, err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("%s mismatch\n--- got ---\n%s\n--- want ---\n%s", tc.file, got, want)
			}
		})
	}
}

// TestArtifactsCarryNoSecrets is the regression guard behind the secret
// externalization: every deployment artifact must reference the admin password
// and the redundancy pre-shared key by name and never carry their values, so
// `--gen-only` output is safe to share. Distinctive values make a leak
// unmistakable -- the goldens alone would not catch one reintroduced together
// with a regenerated golden.
func TestArtifactsCarryNoSecrets(t *testing.T) {
	const (
		pass     = "UNIQUE-ADMIN-PASSWORD-VALUE"
		psk      = "UNIQUE-PRESHARED-KEY-VALUE"
		userPass = "UNIQUE-EXTRA-USER-PASSWORD"
	)
	secrets := []string{pass, psk, userPass}

	for _, p := range []config.Platform{config.K8s, config.Docker, config.Podman} {
		c := load(t, p)
		c.SEMP.AdminPass = pass
		c.Redundancy.PSK = psk
		c.SEMP.AdditionalUsers = []config.AdditionalUser{
			{Username: "appuser", AccessLevel: "read-only", Password: userPass},
		}
		id := c.ResolveNode(config.Primary)

		artifacts := map[string][]byte{}
		switch p {
		case config.K8s:
			artifacts["broker CR"] = BrokerCR(c)
		case config.Podman:
			artifacts["quadlet"] = Quadlet(c, id)
		default:
			artifacts["compose"] = Compose(c, id)
		}
		for name, body := range artifacts {
			for _, secret := range secrets {
				if bytes.Contains(body, []byte(secret)) {
					t.Errorf("%s %s artifact carries the secret %q; it must reference it by name instead", p, name, secret)
				}
			}
		}
	}
}

// TestContainerSecretsRedundancy pins which secrets exist per mode: standalone has
// no mate link, so the PSK secret must not be referenced at all there.
func TestContainerSecretsRedundancy(t *testing.T) {
	c := load(t, config.Podman)

	ha := ContainerSecrets(c, config.Podman)
	if len(ha) != 2 {
		t.Fatalf("HA secrets = %d, want 2 (admin password + PSK)", len(ha))
	}
	if ha[0].EnvKey != "username_admin_password" || ha[1].EnvKey != "redundancy_authentication_presharedkey_key" {
		t.Errorf("HA secret env keys = %q, %q", ha[0].EnvKey, ha[1].EnvKey)
	}
	if got := ha[0].FilePathKey(); got != "username_admin_passwordfilepath" {
		t.Errorf("FilePathKey = %q", got)
	}
	// The mount is named after the setting, not the host-side secret, so the layout
	// inside the container matches the k8s Secret's data keys.
	if got := ha[0].MountPath(); got != "/mnt/secrets/username_admin_password" {
		t.Errorf("MountPath = %q", got)
	}

	c.Redundancy.Enabled = "false"
	if standalone := ContainerSecrets(c, config.Podman); len(standalone) != 1 {
		t.Errorf("standalone secrets = %d, want 1 (admin password only)", len(standalone))
	}

	// An encrypted server-certificate key adds a third secret, and only then: an
	// empty passphrase must not produce one the broker would use to unlock a plain
	// key. The broker reads it from the mounted file via the *filepath variant.
	c.TLS.CertPassphrase = "cert-pass"
	withPass := ContainerSecrets(c, config.Podman)
	if len(withPass) != 2 {
		t.Fatalf("standalone + passphrase = %d secrets, want 2", len(withPass))
	}
	pass := withPass[1]
	if pass.EnvKey != "tls_servercertificate_passphrase" {
		t.Errorf("passphrase env key = %q", pass.EnvKey)
	}
	if got := pass.FilePathKey(); got != "tls_servercertificate_passphrasefilepath" {
		t.Errorf("passphrase file-path key = %q", got)
	}
}

// TestContainerSecretNamesAreHostScoped pins the de-confliction: the engine-side
// name carries the container name, so two brokers on one host never share a podman
// store entry or a compose variable -- while the in-container filename stays the
// same everywhere. The default name keeps the historical names byte-for-byte.
func TestContainerSecretNamesAreHostScoped(t *testing.T) {
	c := load(t, config.Docker)
	if got := ContainerSecrets(c, config.Docker)[0].Name; got != "solace-admin-password" {
		t.Errorf("with the default container name the secret must stay %q, got %q", "solace-admin-password", got)
	}

	c.Docker.Container.Name = "edge-2.broker"
	s := ContainerSecrets(c, config.Docker)[0]
	if s.Name != "edge-2.broker-admin-password" {
		t.Errorf("secret name = %q, want it prefixed with the container name", s.Name)
	}
	// The in-container path must not carry the host-side prefix, and target must be
	// the absolute path rather than a bare filename -- a bare one is resolved under
	// the engine's own /run/secrets, where the broker never looks.
	if s.Target() != "/mnt/secrets/username_admin_password" ||
		s.MountPath() != "/mnt/secrets/username_admin_password" {
		t.Errorf("the in-container name must not carry the host prefix, and target must be absolute: "+
			"target=%q path=%q", s.Target(), s.MountPath())
	}
	// '.' and '-' cannot appear in a variable name; a leading digit cannot start one.
	if got := s.EnvVar(); got != "EDGE_2_BROKER_ADMIN_PASSWORD" {
		t.Errorf("EnvVar = %q", got)
	}
	c.Docker.Container.Name = "9lives"
	if got := ContainerSecrets(c, config.Docker)[0].EnvVar(); got != "_9LIVES_ADMIN_PASSWORD" {
		t.Errorf("a name starting with a digit must be prefixed to stay exportable, got %q", got)
	}
}

// TestAdditionalUsersReachBothHalves pins the extra-user wiring: the password is a
// secret named per host and mounted under the setting it feeds, while the access
// level is not a secret and rides the artifact as an ordinary pair.
func TestAdditionalUsersReachBothHalves(t *testing.T) {
	c := load(t, config.Docker)
	c.Redundancy.Enabled = "false"
	c.SEMP.AdditionalUsers = []config.AdditionalUser{
		{Username: "appuser", AccessLevel: "read-write", Password: "app-secret"},
	}

	// Three on docker, not two: the sample sets tls.cert/certKey, so the
	// server-certificate bundle is appended LAST. That ordering is what keeps
	// secrets[1] the additional user rather than shifting every index.
	secrets := ContainerSecrets(c, config.Docker)
	if len(secrets) != 3 {
		t.Fatalf("secrets = %d, want 3 (admin + appuser + server certificate)", len(secrets))
	}
	u := secrets[1]
	if u.Name != "solace-user-appuser-password" || u.EnvKey != "username_appuser_password" {
		t.Errorf("additional-user secret = %+v", u)
	}
	if u.ConfigKey != "semp.additionalUsers.appuser.password" {
		t.Errorf("ConfigKey = %q; it must name the env-file key for an actionable error", u.ConfigKey)
	}

	pairs := envLines(EnvPairs(c, c.ResolveNode(config.Primary)))
	for _, want := range []string{
		"username_appuser_globalaccesslevel=read-write",
		"username_appuser_passwordfilepath=/mnt/secrets/username_appuser_password",
	} {
		if !strings.Contains(string(pairs), want) {
			t.Errorf("env pairs should contain %q:\n%s", want, pairs)
		}
	}
	if strings.Contains(string(pairs), "app-secret") {
		t.Errorf("env pairs must not carry the password:\n%s", pairs)
	}
}

// TestQuadletHealthCmdEscapesPercent covers the specifier trap: systemd expands
// %-specifiers in every assignment in a unit file, not only the quoted
// Environment= ones, so a percent-encoded character in a probe URL has to be
// doubled or systemd fails to resolve it and drops the line -- silently disabling
// the health check. Quotes and backslashes must NOT be escaped here: the value is
// unquoted and podman splits the command line itself.
func TestQuadletHealthCmdEscapesPercent(t *testing.T) {
	c := load(t, config.Podman)
	hc := healthCheckFixture()
	// An explicit probe, which is also what an older broker would have to use.
	hc.Cmd = []string{"curl", "-fs", `http://localhost:5550/health?q=a%20b&s="x"`}
	c.Podman.Container.HealthCheck = hc
	got := string(Quadlet(c, c.ResolveNode(config.Primary)))

	if !strings.Contains(got, `HealthCmd=curl -fs http://localhost:5550/health?q=a%%20b&s="x"`) {
		t.Errorf("HealthCmd must double %% and leave quotes alone:\n%s", got)
	}
	// Compose has no specifier expansion, so the same probe must stay literal there:
	// doubling the percent would change the URL the broker is actually polled with.
	d := load(t, config.Docker)
	d.Docker.Container.HealthCheck = hc
	compose := string(Compose(d, d.ResolveNode(config.Primary)))
	if !strings.Contains(compose, `q=a%20b`) {
		t.Errorf("compose must keep the percent single:\n%s", compose)
	}
	if strings.Contains(compose, `%%`) {
		t.Errorf("compose must not apply systemd's percent doubling:\n%s", compose)
	}
}

// TestHealthCmdDefaultsToReadiness pins the built-in probe: an enabled block with
// no cmd polls the broker's own readiness endpoint, and an explicit cmd wins.
func TestHealthCmdDefaultsToReadiness(t *testing.T) {
	got := healthCmd(config.HealthCheck{Enabled: true})
	want := []string{"curl", "-fs", "http://localhost:5550/health-check/readiness"}
	if len(got) != len(want) {
		t.Fatalf("default probe = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("default probe = %v, want %v", got, want)
			break
		}
	}
	custom := []string{"/opt/probe.sh"}
	if only := healthCmd(config.HealthCheck{Enabled: true, Cmd: custom}); only[0] != custom[0] || len(only) != 1 {
		t.Errorf("an explicit cmd must win, got %v", only)
	}
}

// TestCustomVolumeMountRendersTheCRArray pins the translation at the boundary: the env
// file keys on this tool's lowercase role word, the CRD constrains
// customVolumeMount[].name to a capitalised enum, and the order is fixed rather than map
// order -- a Go map iterates randomly, and this renders into a CR that is diffed and
// re-applied, so an unstable order would look like a change on every deploy.
func TestCustomVolumeMountRendersTheCRArray(t *testing.T) {
	c := load(t, config.K8s)
	c.K8s.Storage.Class = ""
	c.K8s.Storage.CustomVolumeMount = map[string]string{
		"monitor": "pvc-m", "primary": "pvc-p", "backup": "pvc-b",
	}
	got := string(BrokerCR(c))

	want := "    customVolumeMount:\n" +
		"    - name: Primary\n      persistentVolumeClaim:\n        claimName: pvc-p\n" +
		"    - name: Backup\n      persistentVolumeClaim:\n        claimName: pvc-b\n" +
		"    - name: Monitor\n      persistentVolumeClaim:\n        claimName: pvc-m\n"
	if !strings.Contains(got, want) {
		t.Errorf("CR storage block missing the custom mounts in redundancy order:\n%s", got)
	}
	if strings.Contains(got, "useStorageClass") {
		t.Error("a custom-mounted deployment must not also name a StorageClass")
	}
}

// TestNoCustomVolumeMountEmitsNothing: the block is absent, not empty. An empty array is a
// different statement from an unset field, and the operator reads them differently.
func TestNoCustomVolumeMountEmitsNothing(t *testing.T) {
	c := load(t, config.K8s)
	if got := string(BrokerCR(c)); strings.Contains(got, "customVolumeMount") {
		t.Errorf("no custom mounts configured, so the key must not appear:\n%s", got)
	}
}

// TestPreSharedAuthKeySecretFollowsTheKey pins the k8s half of the PSK asymmetry. The
// field names a Secret, not a value, so emitting it unconditionally would point the
// operator at a Secret carrying no `preshared_auth_key` entry -- breaking a deployment
// the operator would otherwise have keyed itself. Absent means "generate your own".
func TestPreSharedAuthKeySecretFollowsTheKey(t *testing.T) {
	const field = "preSharedAuthKeySecret: solace-admin-secret"

	c := load(t, config.K8s) // the sample carries a literal psk
	if got := string(BrokerCR(c)); !strings.Contains(got, field) {
		t.Errorf("a configured psk must point the CR at the admin Secret:\n%s", got)
	}

	// The reference form is equally "a key was configured": Load resolves pskEnv into
	// PSK, but a config built by hand may carry only the variable name.
	c.Redundancy.PSK = ""
	c.Redundancy.PSKEnv = "SOLACE_REDUNDANCY_PSK"
	if got := string(BrokerCR(c)); !strings.Contains(got, field) {
		t.Errorf("pskEnv names a key too, so the field must still be emitted:\n%s", got)
	}

	c.Redundancy.PSKEnv = ""
	if got := string(BrokerCR(c)); strings.Contains(got, "preSharedAuthKeySecret") {
		t.Errorf("with no key configured the field must be absent, not empty:\n%s", got)
	}
}

// TestExtraEnvVarsSecretFollowsTheUsers pins the CR half of admin.additionalUsers on
// Kubernetes. The field names a Secret, so emitting it with no users would point the
// operator at an object that does not exist and fail the pod on a mount the deployment
// never needed -- the same rule preSharedAuthKeySecret follows.
func TestExtraEnvVarsSecretFollowsTheUsers(t *testing.T) {
	c := load(t, config.K8s)
	if got := string(BrokerCR(c)); strings.Contains(got, "extraEnvVarsSecret") {
		t.Errorf("no additional users configured, so the field must be absent:\n%s", got)
	}

	c.SEMP.AdditionalUsers = []config.AdditionalUser{
		{Username: "appuser", AccessLevel: "read-only", Password: "app-pass"},
	}
	got := string(BrokerCR(c))
	if !strings.Contains(got, "extraEnvVarsSecret: dev-broker-additional-users") {
		t.Errorf("the CR must name the additional-users Secret:\n%s", got)
	}
	// The CR is not where a secret value belongs, whatever else changes.
	if strings.Contains(got, "app-pass") {
		t.Errorf("the CR carries a password:\n%s", got)
	}
}

// TestSecretPreflight pins the precondition `broker deploy` and `broker generate`
// share: creating a secret with an empty value leaves the broker with a blank
// password or mate-link key that only fails later, so it is refused up front.
func TestSecretPreflight(t *testing.T) {
	c := load(t, config.Podman) // HA sample: admin password and PSK both set
	if err := SecretPreflight(c, config.Podman); err != nil {
		t.Fatalf("a fully configured deployment must pass preflight: %v", err)
	}

	c.Redundancy.PSK = ""
	err := SecretPreflight(c, config.Podman)
	if err == nil {
		t.Fatal("an empty PSK must be refused before a secret is created from it")
	}
	if !strings.Contains(err.Error(), "redundancy.psk") || !strings.Contains(err.Error(), "env file") {
		t.Errorf("the error should name the field and the fix, got: %v", err)
	}

	// Standalone has no mate link, so an empty PSK is not a secret at all there.
	c.Redundancy.Enabled = "false"
	if err := SecretPreflight(c, config.Podman); err != nil {
		t.Errorf("standalone does not need a PSK: %v", err)
	}

	c.SEMP.AdminPass = ""
	if err := SecretPreflight(c, config.Podman); err == nil || !strings.Contains(err.Error(), "semp.adminPass") {
		t.Errorf("an empty admin password must be refused, got: %v", err)
	}
}

// TestParsePort covers the port-entry parsing branches independently of the CR.
func TestParsePort(t *testing.T) {
	cases := []struct {
		in                              string
		name, container, service, proto string
	}{
		{"tcp-semp=8080", "tcp-semp", "8080", "8080", "TCP"},
		{"tcp-smf=55555:55556", "tcp-smf", "55555", "55556", "TCP"},
		{"tls-smf=55443/TCP", "tls-smf", "55443", "55443", "TCP"},
		{"udp-x=1234:5678/UDP", "udp-x", "1234", "5678", "UDP"},
	}
	for _, tc := range cases {
		got := parsePort(tc.in)
		if got.name != tc.name || got.container != tc.container || got.service != tc.service || got.proto != tc.proto {
			t.Errorf("parsePort(%q) = %+v, want {%s %s %s %s}",
				tc.in, got, tc.name, tc.container, tc.service, tc.proto)
		}
	}
}

// TestParseToleration covers Equal (key=value) vs Exists (bare key) forms.
func TestParseToleration(t *testing.T) {
	key, val, eff, equal := parseToleration("dedicated=solace:NoSchedule")
	if key != "dedicated" || val != "solace" || eff != "NoSchedule" || !equal {
		t.Errorf("Equal form: got (%q,%q,%q,%v)", key, val, eff, equal)
	}
	key, val, eff, equal = parseToleration("dedicated:NoExecute")
	if key != "dedicated" || val != "" || eff != "NoExecute" || equal {
		t.Errorf("Exists form: got (%q,%q,%q,%v)", key, val, eff, equal)
	}
}

// TestQuadletEscape guards the systemd Environment=" " escaping.
func TestQuadletEscape(t *testing.T) {
	if got := quadletEscape(`a%b"c\d`); got != `a%%b\"c\\d` {
		t.Errorf("quadletEscape = %q", got)
	}
}

// TestScalingTierReachesEveryArtifact pins the point of the change: one tier
// value in the env file decides the CPU cap in all three artifacts, and the
// goldens only ever show the default tier. 100000 is not the default on any
// platform, so this also proves the value is read rather than hardcoded.
func TestScalingTierReachesEveryArtifact(t *testing.T) {
	k := load(t, config.K8s)
	k.Scaling.MaxConnections = 100000
	k.K8s.MsgNode.Mem = "" // clear the tier-100 default so the new tier can fill it
	k.ApplyDefaults(config.K8s)
	cr := string(BrokerCR(k))
	if !strings.Contains(cr, `messagingNodeCpu: "8"`) {
		t.Errorf("broker CR missing the 100000-tier CPU:\n%s", cr)
	}
	if !strings.Contains(cr, "messagingNodeMemory: 30925Mi") {
		t.Errorf("broker CR missing the 100000-tier memory:\n%s", cr)
	}

	d := load(t, config.Docker)
	d.Scaling.MaxConnections = 100000
	d.Docker.Container.Mem = ""
	d.ApplyDefaults(config.Docker)
	compose := string(Compose(d, d.ResolveNode(config.Primary)))
	if !strings.Contains(compose, `    cpus: "8"`) || !strings.Contains(compose, "    mem_limit: 30925m") {
		t.Errorf("compose missing the 100000-tier limits:\n%s", compose)
	}

	p := load(t, config.Podman)
	p.Scaling.MaxConnections = 100000
	p.Podman.Container.Mem = ""
	p.ApplyDefaults(config.Podman)
	unit := string(Quadlet(p, p.ResolveNode(config.Primary)))
	if !strings.Contains(unit, "PodmanArgs=--cpus=8") || !strings.Contains(unit, "Memory=30925m") {
		t.Errorf("quadlet missing the 100000-tier limits:\n%s", unit)
	}
}

// scalingSettings is every broker scaling setting whose destination name is
// IDENTICAL on every platform: k8s writes them into spec.systemScaling and the
// containers pass them as environment variables, under the same name.
// // TestScalingReachesContainersAsEnv each check it under its own
// platform-specific name instead of through this shared list.
var scalingSettings = []string{
	"system_scaling_maxconnectioncount",
	"system_scaling_maxqueuemessagecount",
	"system_scaling_maxkafkabridgecount",
	"system_scaling_maxkafkabrokerconnectioncount",
	"system_scaling_maxbridgecount",
	"system_scaling_maxsubscriptioncount",
	"system_scaling_maxguaranteedmessagesize",
	"messagespool_maxspoolusage",
}

// TestScalingReachesContainersAsEnv pins the delivery split. Every scaling knob
// must reach docker and podman as a container environment variable -- five of
// them used to be rendered on k8s only -- with the values the env file gave,
// including an explicit 0, which is a real setting and not an absent one.
func TestScalingReachesContainersAsEnv(t *testing.T) {
	for _, p := range []config.Platform{config.Docker, config.Podman} {
		c := load(t, p)
		c.Scaling.MaxQueueMessages = 240
		c.Scaling.MaxKafkaBridge = 0
		c.Scaling.MaxKafkaConnections = 0
		c.Scaling.MaxBridges = 500
		c.Scaling.MaxSubscriptions = 5000000
		c.Scaling.MaxGuaranteedMsgMB = 30
		c.Scaling.MaxSpoolUsageMB = 1500

		got := map[string]string{}
		for _, pair := range EnvPairs(c, c.ResolveNode(config.Primary)) {
			got[pair.Key] = pair.Value
		}
		for _, want := range scalingSettings {
			if _, ok := got[want]; !ok {
				t.Errorf("%s: %s is not passed to the container", p, want)
			}
		}
		for k, want := range map[string]string{
			"system_scaling_maxqueuemessagecount":          "240",
			"system_scaling_maxkafkabridgecount":           "0",
			"system_scaling_maxkafkabrokerconnectioncount": "0",
			"system_scaling_maxbridgecount":                "500",
			"system_scaling_maxsubscriptioncount":          "5000000",
			"system_scaling_maxguaranteedmessagesize":      "30",
			"messagespool_maxspoolusage":                   "1500",
		} {
			if got[k] != want {
				t.Errorf("%s: %s = %q, want %q", p, k, got[k], want)
			}
		}
	}
}

// TestScalingReachesK8sAsSpecOnly is the other half: on k8s these are CR fields
// under spec.systemScaling, never container environment variables. The operator
// owns the pod, so an env-var delivery there would be both wrong and invisible.
func TestScalingReachesK8sAsSpecOnly(t *testing.T) {
	c := load(t, config.K8s)
	// A value no default or tier would produce, so the spool assertion below proves
	// the renderer READ this field rather than happening to match the fixture.
	c.Scaling.MaxSpoolUsageMB = 1507
	cr := string(BrokerCR(c))
	if !strings.Contains(cr, "  systemScaling:\n") {
		t.Fatalf("broker CR has no systemScaling block:\n%s", cr)
	}
	for _, want := range scalingSettings {
		if !strings.Contains(cr, "    "+want+": ") {
			t.Errorf("broker CR does not carry %s under systemScaling:\n%s", want, cr)
		}
	}
	// The spool key is the RENAME this change makes, so it gets a value check of
	// its own rather than resting on the presence check above: a renderer that
	// emitted the new name against the wrong field would pass presence and be
	// wrong in exactly the way a rename goes wrong.
	if !strings.Contains(cr, "    messagespool_maxspoolusage: 1507\n") {
		t.Errorf("broker CR should carry the spool size under its new name, from "+
			"scaling.maxSpoolUsageMB:\n%s", cr)
	}
	// maxSpoolUsage was the CR's OWN spelling until this change; now the CR uses
	// the same messagespool_maxspoolusage name the containers always have, so the
	// old camelCase form must no longer appear anywhere.
	if strings.Contains(cr, "maxSpoolUsage:") {
		t.Errorf("the retired camelCase spool spelling must not appear in the CR:\n%s", cr)
	}
	// env: is how a pod would take variables; the CR must not grow one for these.
	if strings.Contains(cr, "\n  env:\n") {
		t.Errorf("scaling must reach k8s as spec fields, not pod env vars:\n%s", cr)
	}
}

// TestContainerMemOverrideReachesArtifact proves the asymmetry survives to the
// artifact: memory is the operator's to override, CPU is not.
func TestContainerMemOverrideReachesArtifact(t *testing.T) {
	d := load(t, config.Docker)
	d.Docker.Container.Mem = "24g"
	compose := string(Compose(d, d.ResolveNode(config.Primary)))
	if !strings.Contains(compose, "    mem_limit: 24g") {
		t.Errorf("compose did not carry the mem override:\n%s", compose)
	}
	if !strings.Contains(compose, `    cpus: "2"`) {
		t.Errorf("compose CPU should stay the tier's, got:\n%s", compose)
	}
}

// TestUnresolvedTierOmitsLimits covers the renderers' fail-safe branch. A Config
// built in code -- which is what the executors are handed, and what several
// container tests construct -- carries maxConnections 0, which is no tier. The
// artifacts must then omit the limits rather than emit an empty cpus:/--cpus=,
// which the engines would reject outright.
func TestUnresolvedTierOmitsLimits(t *testing.T) {
	d := load(t, config.Docker)
	d.Scaling.CPU = ""
	d.Docker.Container.Mem = ""
	compose := string(Compose(d, d.ResolveNode(config.Primary)))
	for _, unwanted := range []string{"cpus:", "mem_limit:"} {
		if strings.Contains(compose, unwanted) {
			t.Errorf("compose emitted %q with no tier resolved:\n%s", unwanted, compose)
		}
	}

	p := load(t, config.Podman)
	p.Scaling.CPU = ""
	p.Podman.Container.Mem = ""
	unit := string(Quadlet(p, p.ResolveNode(config.Primary)))
	for _, unwanted := range []string{"PodmanArgs=", "Memory="} {
		if strings.Contains(unit, unwanted) {
			t.Errorf("quadlet emitted %q with no tier resolved:\n%s", unwanted, unit)
		}
	}

	k := load(t, config.K8s)
	k.Scaling.CPU = ""
	if cr := string(BrokerCR(k)); strings.Contains(cr, "messagingNodeCpu:") {
		t.Errorf("broker CR emitted messagingNodeCpu with no tier resolved:\n%s", cr)
	}
}

// TestSecretTargetsAreAbsolutePaths is the guard for the silent half of the
// secrets-directory move.
//
// Both engines resolve a BARE `target=`/`target:` under their own /run/secrets. So
// if Target ever went back to returning just the setting name, the file would be
// created, the container would start, and the broker would look for it under
// secretMount and find nothing -- with no error from the engine, the tool, or the
// broker's own startup. Authentication would simply fail later, somewhere else.
// Nothing about that failure points here, which is why it gets its own test rather
// than relying on the goldens.
func TestSecretTargetsAreAbsolutePaths(t *testing.T) {
	c := load(t, config.Podman)
	c.Redundancy.Enabled = "true"
	c.TLS.CertPassphrase = "cert-pass"
	c.SEMP.AdditionalUsers = []config.AdditionalUser{
		{Username: "appuser", AccessLevel: "read-only", Password: "pw"},
	}
	secrets := ContainerSecrets(c, config.Podman)
	if len(secrets) < 4 {
		t.Fatalf("expected the admin password, an extra user, the PSK and the passphrase; got %d", len(secrets))
	}
	for _, s := range secrets {
		if !strings.HasPrefix(s.Target(), "/") {
			t.Errorf("secret %q target %q must be an absolute path: a bare name is resolved under the "+
				"engine's own /run/secrets, where the broker never looks", s.Name, s.Target())
		}
		if s.Target() != s.MountPath() {
			t.Errorf("secret %q target %q and mount path %q must be the same expression, or the engine "+
				"mounts one path while the broker reads another", s.Name, s.Target(), s.MountPath())
		}
		if !strings.HasPrefix(s.MountPath(), secretMount+"/") {
			t.Errorf("secret %q mount path %q must sit under %q", s.Name, s.MountPath(), secretMount)
		}
		// The *filepath setting the broker reads has to name that same file.
		if want := s.MountPath(); !strings.HasSuffix(s.FilePathKey(), "filepath") {
			t.Errorf("secret %q setting %q must be the *filepath variant naming %q", s.Name, s.FilePathKey(), want)
		}
	}
}

// TestSecretsAndCertDoNotNest pins that no secret's file can collide with the
// server certificate's mount, which is the hazard the old layout carried: the cert
// lived INSIDE the secrets directory (/run/secrets/tls.crt beside
// /run/secrets/<setting>), so a setting named tls.crt would have been the same path.
// They are now separate trees, and this asserts it rather than trusting it.
func TestSecretsAndCertDoNotNest(t *testing.T) {
	c := load(t, config.Podman)
	c.Redundancy.Enabled = "true"
	c.TLS.CertPassphrase = "cert-pass"
	seen := map[string]string{}
	for _, s := range ContainerSecrets(c, config.Podman) {
		if other, dup := seen[s.MountPath()]; dup {
			t.Errorf("secrets %q and %q mount at the same path %q", other, s.Name, s.MountPath())
		}
		seen[s.MountPath()] = s.Name
		if s.MountPath() == certMount {
			t.Errorf("secret %q mounts at the server certificate's path %q", s.Name, certMount)
		}
		if strings.HasPrefix(certMount, s.MountPath()+"/") || strings.HasPrefix(s.MountPath(), certMount+"/") {
			t.Errorf("secret %q at %q nests with the certificate mount %q; a directory mount shadows "+
				"anything under it", s.Name, s.MountPath(), certMount)
		}
	}
}

// TestComposeProjectIsDeclaredNotDerived. With no top-level `name:`, compose takes
// the project name from the basename of the directory holding the file -- so it
// changes when the artifact is generated somewhere else or the directory is renamed,
// and the previous project's containers, network and volumes become orphans that
// `down` no longer finds. Declaring it ties the project to the container name, which
// is already unique on this host and already validated.
func TestComposeProjectIsDeclaredNotDerived(t *testing.T) {
	cfg := load(t, config.Docker)
	got := Compose(cfg, cfg.ResolveNode(config.Primary))

	want := "name: " + strconv.Quote(ComposeProject(cfg.Docker.Container.Name))
	if !strings.Contains(string(got), want) {
		t.Fatalf("the compose file should declare its project name (%s):\n%s", want, got)
	}
	// First line, so it is the first thing read in `generate broker` and cannot be
	// mistaken for a field of the service below it.
	if first := strings.SplitN(string(got), "\n", 2)[0]; first != want {
		t.Errorf("project name should be the first line, got %q", first)
	}
}

// TestComposeProjectFoldsToComposesGrammar. Compose's project-name grammar is
// narrower than the container-name grammar config enforces: it lowercases and admits
// only '_' and '-' as punctuation, while `My.Broker` is a container name both engines
// accept. Folding is the right trade -- refusing a perfectly good container name over
// a compose spelling rule would not be.
func TestComposeProjectFoldsToComposesGrammar(t *testing.T) {
	cases := map[string]string{
		"solace":       "solace",
		"My.Broker":    "my-broker",
		"sol_pod-1":    "sol_pod-1",
		"UPPER":        "upper",
		"a.b.c":        "a-b-c",
		"broker.prod2": "broker-prod2",
		"":             "",
	}
	for in, want := range cases {
		if got := ComposeProject(in); got != want {
			t.Errorf("ComposeProject(%q) = %q, want %q", in, got, want)
		}
	}
	// Every fold result must be a legal compose project name: lowercase
	// alphanumerics, '_' and '-', starting with an alphanumeric. config already
	// guarantees the first character is a letter or a digit, and lowercasing keeps
	// it one, so the fold cannot produce a leading '-'.
	legal := regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)
	for in := range cases {
		got := ComposeProject(in)
		if got == "" {
			continue
		}
		if !legal.MatchString(got) {
			t.Errorf("ComposeProject(%q) = %q, which compose would refuse", in, got)
		}
	}
}

package k8s

import (
	"bytes"
	"encoding/base64"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"solace/internal/config"
)

// update regenerates the testdata goldens for the whole k8s package (secrets,
// operator, ...). Run `go test ./internal/k8s -update` only after eyeballing the
// diff; the committed goldens are the reviewed expected output.
var update = flag.Bool("update", false, "regenerate golden files in testdata/")

// sampleFixture is the shared env template, reused as the golden fixture so one file
// drives every renderer (matches internal/render's approach).
const sampleFixture = "../../env/sample.yaml"

func loadK8s(t *testing.T) *config.Config {
	t.Helper()
	c, err := config.Load(sampleFixture, config.K8s)
	if err != nil {
		t.Fatalf("load %s under k8s: %v", sampleFixture, err)
	}
	return c
}

// checkGolden compares got against testdata/<file>, or rewrites it under -update.
func checkGolden(t *testing.T, file string, got []byte) {
	t.Helper()
	golden := filepath.Join("testdata", file)
	if *update {
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatalf("write golden %s: %v", golden, err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden %s (regenerate: go test ./internal/k8s -update): %v", golden, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("%s mismatch\n--- got ---\n%s\n--- want ---\n%s", file, got, want)
	}
}

func TestSecretGoldens(t *testing.T) {
	cases := []struct {
		name string
		file string
		gen  func(t *testing.T) []byte
	}{
		{
			name: "admin secret",
			file: "admin_secret.golden",
			gen: func(t *testing.T) []byte {
				cfg := loadK8s(t)
				// additionalUsers are configured but must NOT reach this Secret: the
				// operator only reads the admin and monitor keys, so extra ones were
				// passwords written where nothing would ever read them. The golden is
				// what pins that -- see TestAdminSecretExcludesAdditionalUsers.
				cfg.SEMP.AdditionalUsers = []config.AdditionalUser{
					{Username: "appuser", AccessLevel: "read-only", Password: "apppass"},
				}
				b, err := AdminSecret(cfg)
				if err != nil {
					t.Fatalf("AdminSecret: %v", err)
				}
				return b
			},
		},
		{
			name: "tls secret",
			file: "tls_secret.golden",
			gen: func(t *testing.T) []byte {
				cfg := loadK8s(t)
				dir := t.TempDir()
				crt := filepath.Join(dir, "tls.crt")
				ca := filepath.Join(dir, "ca.crt")
				key := filepath.Join(dir, "tls.key")
				writeFile(t, crt, "CERTDATA\n")
				writeFile(t, ca, "CADATA\n")
				writeFile(t, key, "KEYDATA\n")
				cfg.TLS.Cert = crt
				cfg.TLS.CAs = []string{ca}
				cfg.TLS.CertKey = key
				// The sample no longer opts into TLS, so the name comes from here too.
				cfg.K8s.TLSServerSecret = "solace-tls-secret"
				b, err := TLSSecret(cfg)
				if err != nil {
					t.Fatalf("TLSSecret: %v", err)
				}
				return b
			},
		},
		{
			name: "docker registry secret",
			file: "docker_registry_secret.golden",
			gen: func(t *testing.T) []byte {
				b, err := DockerRegistrySecret(loadK8s(t))
				if err != nil {
					t.Fatalf("DockerRegistrySecret: %v", err)
				}
				return b
			},
		},
		{
			name: "operator regcred",
			file: "operator_regcred.golden",
			gen: func(t *testing.T) []byte {
				b, err := operatorRegcred(loadK8s(t), "pubsubplus-operator-system")
				if err != nil {
					t.Fatalf("operatorRegcred: %v", err)
				}
				return b
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			checkGolden(t, tc.file, tc.gen(t))
		})
	}
}

// TestAdminSecretDecodes proves the base64 data round-trips to the plaintext
// secrets (the golden guards format; this guards semantics).
func TestAdminSecretDecodes(t *testing.T) {
	cfg := loadK8s(t)
	got, err := AdminSecret(cfg)
	if err != nil {
		t.Fatalf("AdminSecret: %v", err)
	}
	if dec := decodeDataValue(t, got, "username_admin_password"); dec != "CHANGE-ME-admin" {
		t.Errorf("admin password decodes to %q", dec)
	}
	if dec := decodeDataValue(t, got, "username_monitor_password"); dec != "CHANGE-ME-monitor" {
		t.Errorf("monitor password decodes to %q", dec)
	}
}

// TestAdminSecretExcludesAdditionalUsers pins the finding that shaped the k8s user
// path: the operator reads only the admin and monitor keys out of this Secret, so
// writing a third user's password into it achieved nothing except putting a secret
// somewhere nobody reads. Those users are created over the CLI instead
// (`config additional-users`).
func TestAdminSecretExcludesAdditionalUsers(t *testing.T) {
	cfg := loadK8s(t)
	cfg.SEMP.AdditionalUsers = []config.AdditionalUser{
		{Username: "appuser", AccessLevel: "read-only", Password: "UNIQUE-APPUSER-PASSWORD"},
	}
	got, err := AdminSecret(cfg)
	if err != nil {
		t.Fatalf("AdminSecret: %v", err)
	}
	for _, absent := range []string{"username_appuser_password", "UNIQUE-APPUSER-PASSWORD"} {
		if bytes.Contains(got, []byte(absent)) {
			t.Errorf("the admin Secret must not carry %q; the operator ignores it:\n%s", absent, got)
		}
	}
	// The base64 of the password must not appear either -- the data values are encoded.
	enc := base64.StdEncoding.EncodeToString([]byte("UNIQUE-APPUSER-PASSWORD"))
	if bytes.Contains(got, []byte(enc)) {
		t.Errorf("the admin Secret carries the additional user's password base64-encoded:\n%s", got)
	}
}

func TestAdminSecretErrors(t *testing.T) {
	base := loadK8s(t)
	cases := []struct {
		name   string
		mutate func(c *config.Config)
	}{
		{"empty admin pass", func(c *config.Config) { c.SEMP.AdminPass = "" }},
		{"empty secret name", func(c *config.Config) { c.K8s.AdminSecret = "" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := *base
			tc.mutate(&cfg)
			if _, err := AdminSecret(&cfg); err == nil {
				t.Errorf("AdminSecret(%s) expected error", tc.name)
			}
		})
	}
}

func TestTLSSecretErrors(t *testing.T) {
	t.Run("missing cert fields", func(t *testing.T) {
		cfg := loadK8s(t)
		cfg.TLS.Cert = ""
		if _, err := TLSSecret(cfg); err == nil {
			t.Error("TLSSecret should fail when tls.cert is unset")
		}
	})
	t.Run("empty secret name", func(t *testing.T) {
		cfg := loadK8s(t)
		cfg.K8s.TLSServerSecret = ""
		if _, err := TLSSecret(cfg); err == nil {
			t.Error("TLSSecret should fail when kubernetes.tlsServerSecret is unset")
		}
	})
	t.Run("cert file missing", func(t *testing.T) {
		cfg := loadK8s(t)
		cfg.TLS.Cert = filepath.Join(t.TempDir(), "nope.crt")
		if _, err := TLSSecret(cfg); err == nil {
			t.Error("TLSSecret should fail when the cert file does not exist")
		}
	})
	t.Run("CA file missing", func(t *testing.T) {
		cfg := loadK8s(t)
		dir := t.TempDir()
		crt := filepath.Join(dir, "tls.crt")
		key := filepath.Join(dir, "tls.key")
		writeFile(t, crt, "C\n")
		writeFile(t, key, "K\n")
		cfg.TLS.Cert = crt
		cfg.TLS.CertKey = key
		cfg.TLS.CAs = []string{filepath.Join(dir, "missing-ca.crt")}
		if _, err := TLSSecret(cfg); err == nil {
			t.Error("TLSSecret should fail when a CA file does not exist")
		}
	})
	t.Run("key file missing", func(t *testing.T) {
		cfg := loadK8s(t)
		dir := t.TempDir()
		crt := filepath.Join(dir, "tls.crt")
		writeFile(t, crt, "C\n")
		cfg.TLS.Cert = crt
		cfg.TLS.CertKey = filepath.Join(dir, "missing.key")
		cfg.TLS.CAs = nil
		if _, err := TLSSecret(cfg); err == nil {
			t.Error("TLSSecret should fail when the key file does not exist")
		}
	})
}

func TestDockerRegistrySecretEmptyName(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.ImagePullSecret = ""
	if _, err := DockerRegistrySecret(cfg); err == nil {
		t.Error("DockerRegistrySecret should fail with an empty pull-secret name")
	}
}

// --- test helpers ----------------------------------------------------------

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// decodeDataValue finds `  <key>: <base64>` in a rendered secret and returns the
// decoded plaintext. Fails the test if the key is absent or not valid base64.
func decodeDataValue(t *testing.T, manifest []byte, key string) string {
	t.Helper()
	prefix := []byte("  " + key + ": ")
	for _, line := range bytes.Split(manifest, []byte("\n")) {
		if bytes.HasPrefix(line, prefix) {
			raw := bytes.TrimPrefix(line, prefix)
			dec, err := base64.StdEncoding.DecodeString(string(raw))
			if err != nil {
				t.Fatalf("data[%s] is not valid base64: %v", key, err)
			}
			return string(dec)
		}
	}
	t.Fatalf("data key %q not found in manifest:\n%s", key, manifest)
	return ""
}

// TestAdminSecretCarriesThePSKOnlyWhenSet pins the k8s half of the PSK asymmetry from the
// Secret's side. The key rides in the SAME Secret as the credentials -- the CR points at it
// through a separate field, so one object serves both -- and under the CRD's own spelling,
// not ours. An unset key must add no entry at all: an empty preshared_auth_key would be a
// key the operator honours and the group then fails to form on.
func TestAdminSecretCarriesThePSKOnlyWhenSet(t *testing.T) {
	cfg := loadK8s(t)
	cfg.Redundancy.PSK = "UNIQUE-PSK-VALUE"
	got, err := AdminSecret(cfg)
	if err != nil {
		t.Fatalf("AdminSecret: %v", err)
	}
	if dec := decodeDataValue(t, got, pskSecretKey); dec != "UNIQUE-PSK-VALUE" {
		t.Errorf("%s decodes to %q, want the configured key", pskSecretKey, dec)
	}

	cfg.Redundancy.PSK = ""
	got, err = AdminSecret(cfg)
	if err != nil {
		t.Fatalf("AdminSecret without a psk: %v", err)
	}
	if bytes.Contains(got, []byte(pskSecretKey)) {
		t.Errorf("no psk configured, so the entry must be absent rather than empty: %s", got)
	}
	// The credentials are unaffected either way -- the psk is additive.
	if dec := decodeDataValue(t, got, "username_admin_password"); dec != "CHANGE-ME-admin" {
		t.Errorf("admin password decodes to %q", dec)
	}
}

// Naming a TLS Secret and supplying the files it is built from are separate decisions.
// These pin what follows from the second one being absent: the CR still references the
// Secret by name, but nothing here builds it, reads a certificate, or removes it.

// TestGenSecretsSkipsATLSSecretItDoesNotOwn is the fix for a real failure: an env file
// naming an existing Secret and supplying no cert/key made `broker generate` read
// certs/tls.crt -- an invented path -- and fail on a file the operator never mentioned.
func TestGenSecretsSkipsATLSSecretItDoesNotOwn(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.TLSServerSecret = "byo-tls-secret"
	cfg.TLS.Cert, cfg.TLS.CertKey = "", ""

	got, err := GenSecrets(cfg)
	if err != nil {
		t.Fatalf("an existing Secret needs no files, so this must not fail: %v", err)
	}
	if bytes.Contains(got, []byte("kubernetes.io/tls")) {
		t.Errorf("no TLS Secret may be built when the env file supplies no cert/key:\n%s", got)
	}
	// The credentials Secret is unaffected -- this narrows one document, not the stream.
	if !bytes.Contains(got, []byte("username_admin_password")) {
		t.Errorf("the admin Secret must still be rendered:\n%s", got)
	}
}

// TestGenSecretsBuildsATLSSecretItOwns is the other arm: with the pair supplied the
// Secret is rendered as before, under the configured name.
func TestGenSecretsBuildsATLSSecretItOwns(t *testing.T) {
	cfg := loadK8s(t)
	dir := t.TempDir()
	cfg.K8s.TLSServerSecret = "solace-tls-secret"
	cfg.TLS.Cert = writeTempPEM(t, dir, "tls.crt", "CERTIFICATE")
	cfg.TLS.CertKey = writeTempPEM(t, dir, "tls.key", "PRIVATE KEY")

	got, err := GenSecrets(cfg)
	if err != nil {
		t.Fatalf("GenSecrets: %v", err)
	}
	for _, want := range []string{"kubernetes.io/tls", "solace-tls-secret"} {
		if !bytes.Contains(got, []byte(want)) {
			t.Errorf("the built TLS Secret is missing %q:\n%s", want, got)
		}
	}
}

// writeTempPEM writes a stub PEM file and returns its path -- enough for the renderers,
// which copy the bytes rather than parsing them.
func writeTempPEM(t *testing.T, dir, name, label string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	body := "-----BEGIN " + label + "-----\nc3R1Yg==\n-----END " + label + "-----\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

// admin.additionalUsers is applied on Kubernetes through a Secret of its own that the CR
// names in spec.extraEnvVarsSecret. These pin the three things that makes true: the Secret
// carries both halves of each user, the credentials Secret does NOT, and the CR references
// it only when there is something to reference.

// TestAdditionalUsersSecretCarriesBothHalves: the access level has to ride the Secret too,
// even though it is not sensitive. extraEnvVarsSecret is the only channel the CRD gives us
// for an environment variable, so a setting that must reach the broker as one has nowhere
// else to go -- unlike the container platforms, which put it in the artifact.
func TestAdditionalUsersSecretCarriesBothHalves(t *testing.T) {
	cfg := loadK8s(t)
	cfg.SEMP.AdditionalUsers = []config.AdditionalUser{
		{Username: "appuser", AccessLevel: "read-only", Password: "UNIQUE-APP-PASSWORD"},
		{Username: "ops_user", AccessLevel: "admin", Password: "UNIQUE-OPS-PASSWORD"},
	}
	got, err := AdditionalUsersSecret(cfg)
	if err != nil {
		t.Fatalf("AdditionalUsersSecret: %v", err)
	}
	for key, want := range map[string]string{
		"username_appuser_globalaccesslevel":  "read-only",
		"username_appuser_password":           "UNIQUE-APP-PASSWORD",
		"username_ops_user_globalaccesslevel": "admin",
		"username_ops_user_password":          "UNIQUE-OPS-PASSWORD",
	} {
		if dec := decodeDataValue(t, got, key); dec != want {
			t.Errorf("%s decodes to %q, want %q", key, dec, want)
		}
	}
	if !bytes.Contains(got, []byte("name: dev-broker-additional-users")) {
		t.Errorf("the Secret must carry the derived name:\n%s", got)
	}
	// Values are base64 in the data block, so the plaintext must not appear anywhere.
	for _, secret := range []string{"UNIQUE-APP-PASSWORD", "UNIQUE-OPS-PASSWORD"} {
		if bytes.Contains(got, []byte(secret)) {
			t.Errorf("the Secret carries %q in plaintext:\n%s", secret, got)
		}
	}
}

// TestAdditionalUsersSecretIsAbsentWithNoUsers: nil, not an empty Secret. An empty one
// would be an object to create, name and clean up for nothing, and the CR omits
// extraEnvVarsSecret in the same case -- naming a Secret that does not exist fails the pod.
func TestAdditionalUsersSecretIsAbsentWithNoUsers(t *testing.T) {
	got, err := AdditionalUsersSecret(loadK8s(t))
	if err != nil {
		t.Fatalf("AdditionalUsersSecret: %v", err)
	}
	if got != nil {
		t.Errorf("no users configured, so there is no Secret to build:\n%s", got)
	}
}

// TestAdditionalUsersStayOutOfTheCredentialsSecret is the property the split exists for,
// and it is not tidiness. extraEnvVarsSecret is projected with envFrom, which turns EVERY
// key of the named Secret into an environment variable: if these users lived in the
// credentials Secret, pointing the CR at it would put the ADMIN and MONITOR passwords in the
// pod environment to get the extra users in.
func TestAdditionalUsersStayOutOfTheCredentialsSecret(t *testing.T) {
	cfg := loadK8s(t)
	cfg.SEMP.AdditionalUsers = []config.AdditionalUser{
		{Username: "appuser", AccessLevel: "read-only", Password: "UNIQUE-APP-PASSWORD"},
	}
	admin, err := AdminSecret(cfg)
	if err != nil {
		t.Fatalf("AdminSecret: %v", err)
	}
	for _, absent := range []string{"username_appuser_password", "username_appuser_globalaccesslevel", "UNIQUE-APP-PASSWORD"} {
		if bytes.Contains(admin, []byte(absent)) {
			t.Errorf("the credentials Secret must not carry %q -- envFrom would export it:\n%s", absent, admin)
		}
	}
	users, err := AdditionalUsersSecret(cfg)
	if err != nil {
		t.Fatalf("AdditionalUsersSecret: %v", err)
	}
	for _, absent := range []string{"username_admin_password", "username_monitor_password"} {
		if bytes.Contains(users, []byte(absent)) {
			t.Errorf("the additional-users Secret must not carry %q -- envFrom would export it:\n%s", absent, users)
		}
	}
}

// TestGenSecretsIncludesTheAdditionalUsers: the stream `broker deploy` applies and
// `broker generate` prints carries the Secret, and omits it when there are no users.
func TestGenSecretsIncludesTheAdditionalUsers(t *testing.T) {
	cfg := loadK8s(t)
	cfg.SEMP.AdditionalUsers = []config.AdditionalUser{
		{Username: "appuser", AccessLevel: "read-only", Password: "app-pass"},
	}
	with, err := GenSecrets(cfg)
	if err != nil {
		t.Fatalf("GenSecrets: %v", err)
	}
	if !bytes.Contains(with, []byte("dev-broker-additional-users")) {
		t.Errorf("the apply stream is missing the additional-users Secret:\n%s", with)
	}

	cfg.SEMP.AdditionalUsers = nil
	without, err := GenSecrets(cfg)
	if err != nil {
		t.Fatalf("GenSecrets: %v", err)
	}
	if bytes.Contains(without, []byte("additional-users")) {
		t.Errorf("no users configured, so no Secret in the stream:\n%s", without)
	}
}

// TestAdditionalUsersSecretRefusesAnEmptyPassword covers the guard's own branch. It is
// defensive rather than everyday: config.Validate already refuses an empty
// additionalUsers password at load, so reaching this needs a Config built in code -- which
// is exactly what the executors are handed, and the same reason AdminSecret keeps its own
// empty-password check.
func TestAdditionalUsersSecretRefusesAnEmptyPassword(t *testing.T) {
	cfg := loadK8s(t)
	cfg.SEMP.AdditionalUsers = []config.AdditionalUser{
		{Username: "appuser", AccessLevel: "read-only", Password: ""},
	}
	_, err := AdditionalUsersSecret(cfg)
	if err == nil {
		t.Fatal("a user with no password must not become a Secret entry")
	}
	if !strings.Contains(err.Error(), "appuser") {
		t.Errorf("error %q should name the user, so the fix is readable off it", err)
	}
}

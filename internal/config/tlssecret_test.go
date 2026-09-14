package config

import (
	"strings"
	"testing"
)

// A TLS Secret can come from either side: this tool builds it from files the env file
// names, or the operator created it themselves and the env file only says which one the
// broker should use. `kubernetes.tlsServerSecret` names it in both cases -- what tells the
// two apart is whether tls.cert/tls.certKey are supplied. These pin that split, and the
// two half-states it makes possible.

// tlsCfg is a valid Kubernetes config with the TLS fields under the caller's control.
func tlsCfg(secret, cert, key string) *Config {
	c := validK8sConfig()
	c.K8s.TLSServerSecret = secret
	c.TLS.Cert = cert
	c.TLS.CertKey = key
	return c
}

// TestNamingTheTLSSecretDoesNotInventCertPaths is the regression this split exists for.
// `applyK8sDefaults` used to fill tls.cert/tls.certKey with `certs/tls.crt` and
// `certs/tls.key` whenever the Secret was named, so "the Secret already exists" turned
// into a read of a file the operator had never mentioned -- and `broker generate` failed
// on a path that appears nowhere in their env file.
func TestNamingTheTLSSecretDoesNotInventCertPaths(t *testing.T) {
	c := tlsCfg("solace-tls-secret", "", "")
	c.ApplyDefaults(K8s)
	if c.TLS.Cert != "" || c.TLS.CertKey != "" {
		t.Errorf("defaulting invented cert=%q key=%q; naming a Secret is not supplying one",
			c.TLS.Cert, c.TLS.CertKey)
	}
	if c.ManagesTLSSecret() {
		t.Error("with no files supplied the Secret is not this tool's to build")
	}
	if err := c.Validate(K8s); err != nil {
		t.Errorf("pointing at an existing Secret is a valid deployment: %v", err)
	}
}

// TestSuppliedCertsMakeTheSecretOurs is the other half: files present means this tool
// builds it, which is what `broker generate` renders and `broker remove` cleans up.
func TestSuppliedCertsMakeTheSecretOurs(t *testing.T) {
	c := tlsCfg("solace-tls-secret", "certs/tls.crt", "certs/tls.key")
	if !c.ManagesTLSSecret() {
		t.Error("a supplied cert/key pair is what makes the Secret ours to build")
	}
	if err := c.Validate(K8s); err != nil {
		t.Errorf("a fully supplied pair must validate: %v", err)
	}
}

// TestTLSCertAndKeyMustBeSetTogetherOnKubernetes: the pairing used to be enforced only on
// docker and podman, because defaulting filled both fields at once and one could not
// arrive alone. Without the defaulting it can, and it would build a Secret carrying a
// certificate and no key -- which the operator mounts and the broker cannot start a
// listener over.
func TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(t *testing.T) {
	for _, tc := range []struct{ name, cert, key string }{
		{"cert alone", "certs/tls.crt", ""},
		{"key alone", "", "certs/tls.key"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tlsCfg("solace-tls-secret", tc.cert, tc.key).Validate(K8s)
			if err == nil {
				t.Fatal("half a pair must be refused")
			}
			for _, want := range []string{"tls.cert", "tls.certKey", "tlsServerSecret"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q should name %q -- including the way out", err, want)
				}
			}
		})
	}
}

// TestSuppliedCertsNeedASecretName: files with no name give the Secret nowhere to be
// created and the CR no tls block, so the certificate would be silently unused -- the
// quietest of the failure modes here, and the one worth a loud error.
func TestSuppliedCertsNeedASecretName(t *testing.T) {
	err := tlsCfg("", "certs/tls.crt", "certs/tls.key").Validate(K8s)
	if err == nil {
		t.Fatal("a cert/key pair with no Secret name must be refused")
	}
	if !strings.Contains(err.Error(), "kubernetes.tlsServerSecret") {
		t.Errorf("error %q should name the missing key", err)
	}
}

// TestNoTLSAtAllStaysValid: TLS is opt-in on every platform, and the checks above must not
// have made the plainest deployment fail.
func TestNoTLSAtAllStaysValid(t *testing.T) {
	c := tlsCfg("", "", "")
	c.ApplyDefaults(K8s)
	if c.ManagesTLSSecret() {
		t.Error("no TLS configured, so there is no Secret to manage")
	}
	if err := c.Validate(K8s); err != nil {
		t.Errorf("a broker with no TLS must validate: %v", err)
	}
}

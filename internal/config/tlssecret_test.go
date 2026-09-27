package config

import (
	"strings"
	"testing"
)

// A TLS Secret can come from either side: this tool builds it from files the env file
// names, or the operator created it themselves and the env file only says which one the
// broker should use. What tells the two apart is whether tls.cert/tls.certKey are
// supplied; `kubernetes.tlsServerSecret` names the Secret in both cases, and may be left
// unset when the files are supplied, since Config.TLSServerSecretName then derives
// <kubernetes.name>-tls -- the three states imagepullsecret_test.go pins for the
// image-pull Secret. These pin that split and the half-states it makes possible.

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

// TestSuppliedCertsDeriveASecretName: files with no name used to be refused, because the
// Secret had nowhere to be created and the CR no tls block. They now get the derived
// default instead, so the certificate is used rather than silently dropped -- and the
// configuration that used to be an error must validate.
func TestSuppliedCertsDeriveASecretName(t *testing.T) {
	c := tlsCfg("", "certs/tls.crt", "certs/tls.key")
	if err := c.Validate(K8s); err != nil {
		t.Fatalf("a cert/key pair with no Secret name must validate now that the name is derived: %v", err)
	}
	if got := c.TLSServerSecretName(); got != "mybroker-tls" {
		t.Errorf("TLSServerSecretName() = %q, want the derived default %q", got, "mybroker-tls")
	}
}

// TestTLSServerSecretNameFollowsTheStates pins every state of the resolver the CR, the
// Secret builder, the delete set and the server-certs routing all read. The configured
// name wins whether or not this tool builds the Secret behind it; a default is derived
// only when there are files to build it from, never invented for an unnamed Secret with
// nothing behind it.
func TestTLSServerSecretNameFollowsTheStates(t *testing.T) {
	cases := []struct {
		name, secret, cert, key, want string
		noK8sName                     bool
	}{
		{name: "named and built from files", secret: "custom-tls", cert: "certs/tls.crt", key: "certs/tls.key", want: "custom-tls"},
		{name: "named, brought by the operator", secret: "byo-tls", want: "byo-tls"},
		{name: "files but no name derives", cert: "certs/tls.crt", key: "certs/tls.key", want: "mybroker-tls"},
		{name: "neither names nothing"},
		// Only a Config built in code gets here -- Load requires kubernetes.name -- but
		// "-tls" alone is no Secret name, so there is nothing to derive.
		{name: "files but no kubernetes.name to derive from", cert: "certs/tls.crt", key: "certs/tls.key", noK8sName: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := tlsCfg(tc.secret, tc.cert, tc.key)
			if tc.noK8sName {
				c.K8s.Name = ""
			}
			if got := c.TLSServerSecretName(); got != tc.want {
				t.Errorf("TLSServerSecretName() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestNoTLSAtAllStaysValid: TLS is opt-in on every platform, and the pairing check above
// must not have made the plainest deployment fail.
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

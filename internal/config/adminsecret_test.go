package config

import (
	"strings"
	"testing"
)

// The broker's admin credentials Secret has the three states the TLS and image-pull
// Secrets have (tlssecret_test.go, imagepullsecret_test.go), keyed on semp.adminPass: with
// it this tool builds the Secret (kubernetes.adminSecret, or <kubernetes.name>-admin);
// with only a name the CR references a Secret someone else made; with neither the CR
// names none and the Solace operator generates one. These pin the states, the load-time
// refusals that keep them honest, and that none of it leaks into docker or podman.

// adminCfg is a valid Kubernetes config with the admin-Secret inputs under the caller's
// control.
func adminCfg(name, pass string) *Config {
	c := validK8sConfig()
	c.K8s.AdminSecret = name
	c.SEMP.AdminPass = pass
	return c
}

// TestAdminSecretNameFollowsTheStates pins every state of the resolver the CR, the
// Secret builder, the delete set, the report and the login read-back all go through.
func TestAdminSecretNameFollowsTheStates(t *testing.T) {
	cases := []struct {
		name, secret, pass, want string
		manages, noK8sName       bool
	}{
		{name: "named and built from the password", secret: "custom-admin", pass: "pw", want: "custom-admin", manages: true},
		{name: "password but no name derives", pass: "pw", want: "mybroker-admin", manages: true},
		{name: "named, brought by the operator", secret: "byo-admin", want: "byo-admin"},
		{name: "neither names nothing -- the operator generates one"},
		// Only a Config built in code gets here -- Load requires kubernetes.name -- but
		// "-admin" alone is no Secret name, so nothing is derived.
		{name: "password but no kubernetes.name to derive from", pass: "pw", manages: true, noK8sName: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := adminCfg(tc.secret, tc.pass)
			if tc.noK8sName {
				c.K8s.Name = ""
			}
			if got := c.AdminSecretName(); got != tc.want {
				t.Errorf("AdminSecretName() = %q, want %q", got, tc.want)
			}
			if got := c.ManagesAdminSecret(); got != tc.manages {
				t.Errorf("ManagesAdminSecret() = %v, want %v -- ownership follows the password, never the name",
					got, tc.manages)
			}
		})
	}
}

// TestKubernetesValidatesWithoutAnAdminPassword: the requirement it replaces guarded
// against a hardcoded default password, and neither remaining state has one -- a Secret
// someone else made, or a random password the operator generates.
func TestKubernetesValidatesWithoutAnAdminPassword(t *testing.T) {
	for _, secret := range []string{"byo-admin", ""} {
		if err := adminCfg(secret, "").Validate(K8s); err != nil {
			t.Errorf("kubernetes.adminSecret=%q with no semp.adminPass must validate: %v", secret, err)
		}
	}
}

// TestMonitorPasswordAndKeyNeedTheAdminPassword: both are entries of the Secret this tool
// builds from semp.adminPass. Without it this tool builds none, and a referenced or
// operator-generated Secret is never written to -- so either would be a key the env file
// sets and the broker never receives. Refused in both remaining states, and in both
// spellings (a Config built in code may carry only the *Env name).
func TestMonitorPasswordAndKeyNeedTheAdminPassword(t *testing.T) {
	cases := []struct {
		name   string
		set    func(*Config)
		naming []string
	}{
		{"monitor password", func(c *Config) { c.SEMP.MonitorPass = "mon" }, []string{"semp.monitorPass", "semp.adminPass"}},
		{"monitor password variable", func(c *Config) { c.SEMP.MonitorPassEnv = "MON" }, []string{"semp.monitorPass", "semp.adminPass"}},
		{"pre-shared key", func(c *Config) { c.Redundancy.PSK = "k" }, []string{"redundancy.psk", "semp.adminPass"}},
		{"pre-shared key variable", func(c *Config) { c.Redundancy.PSKEnv = "PSK" }, []string{"redundancy.psk", "semp.adminPass"}},
	}
	for _, tc := range cases {
		for _, secret := range []string{"byo-admin", ""} {
			t.Run(tc.name+" with adminSecret="+secret, func(t *testing.T) {
				c := adminCfg(secret, "")
				tc.set(c)
				err := c.Validate(K8s)
				if err == nil {
					t.Fatal("must be refused: nothing would carry it to the broker")
				}
				for _, want := range append(tc.naming, "operator generates") {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error %q should mention %q -- the key and both ways out", err, want)
					}
				}
			})
		}
	}
	// With the password both are fine: they ride in the Secret it builds.
	c := adminCfg("", "pw")
	c.SEMP.MonitorPass, c.Redundancy.PSK = "mon", "k"
	if err := c.Validate(K8s); err != nil {
		t.Errorf("a monitor password and key beside semp.adminPass must validate: %v", err)
	}
}

// TestAdminSecretRefusesTheOperatorsOwnNames: the operator generates, owns and deletes
// <name>-pubsubplus-* Secrets with the broker, so pointing the CR at one -- or building
// over one -- leaves the next deploy after `broker remove` without it.
func TestAdminSecretRefusesTheOperatorsOwnNames(t *testing.T) {
	for _, pass := range []string{"pw", ""} {
		c := adminCfg("mybroker-pubsubplus-admin-creds", pass)
		err := c.Validate(K8s)
		if err == nil {
			t.Fatalf("semp.adminPass=%q: an operator-generated name must be refused", pass)
		}
		for _, want := range []string{"mybroker-pubsubplus-admin-creds", "Copy it"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q should mention %q", err, want)
			}
		}
	}
	// Another broker's operator names are just names here.
	if err := adminCfg("other-pubsubplus-admin-creds", "pw").Validate(K8s); err != nil {
		t.Errorf("a different broker's generated name is not this operator's to own here: %v", err)
	}
}

// TestSecretNamesMustNotCollide: the derived defaults make it possible for two Secrets the
// CR names to resolve to one name without either being written twice, and one Secret
// cannot carry two sets of keys.
func TestSecretNamesMustNotCollide(t *testing.T) {
	t.Run("a derived admin name against an explicit TLS name", func(t *testing.T) {
		c := adminCfg("", "pw")
		c.K8s.TLSServerSecret = "mybroker-admin"
		err := c.Validate(K8s)
		if err == nil {
			t.Fatal("two Secrets resolving to one name must be refused")
		}
		for _, want := range []string{"kubernetes.adminSecret (derived from kubernetes.name)", "kubernetes.tlsServerSecret", "mybroker-admin"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q should mention %q", err, want)
			}
		}
	})
	t.Run("the admin name against the additional users' Secret", func(t *testing.T) {
		c := adminCfg("mybroker-additional-users", "pw")
		c.SEMP.AdditionalUsers = []AdditionalUser{{Username: "app", AccessLevel: "read-only", Password: "p"}}
		if err := c.Validate(K8s); err == nil || !strings.Contains(err.Error(), "semp.additionalUsers") {
			t.Errorf("err = %v, want a collision naming semp.additionalUsers", err)
		}
	})
	t.Run("distinct names pass", func(t *testing.T) {
		c := adminCfg("", "pw")
		c.K8s.TLSServerSecret = "tls-secret"
		if err := c.Validate(K8s); err != nil {
			t.Errorf("distinct names must validate: %v", err)
		}
	})
}

// TestContainersStillRequireTheAdminPassword: none of the Kubernetes states reaches
// docker or podman. There is no operator there to generate a password and no Secret to
// reference, so semp.adminPass stays mandatory.
func TestContainersStillRequireTheAdminPassword(t *testing.T) {
	for _, p := range []Platform{Docker, Podman} {
		c := validContainerConfig(p, "false")
		c.SEMP.AdminPass = ""
		err := c.Validate(p)
		if err == nil || !strings.Contains(err.Error(), "semp.adminPass") {
			t.Errorf("%s: err = %v, want semp.adminPass still required", p, err)
		}
	}
}

// TestOperatorAdminSecretName pins the name the operator generates, which the report,
// the remove warning, the deploy guard and the login read-back all name.
func TestOperatorAdminSecretName(t *testing.T) {
	if got := validK8sConfig().OperatorAdminSecretName(); got != "mybroker-pubsubplus-admin-creds" {
		t.Errorf("OperatorAdminSecretName() = %q, want %q", got, "mybroker-pubsubplus-admin-creds")
	}
}

// TestPSKConfiguredCountsBothSpellings: the CR, the report and the refusal above all ask
// this one question, so they cannot disagree about a key supplied only by variable name.
func TestPSKConfiguredCountsBothSpellings(t *testing.T) {
	c := validK8sConfig()
	if c.PSKConfigured() {
		t.Error("no key configured")
	}
	c.Redundancy.PSKEnv = "PSK"
	if !c.PSKConfigured() {
		t.Error("a variable name is a configured key")
	}
	c.Redundancy.PSKEnv, c.Redundancy.PSK = "", "k"
	if !c.PSKConfigured() {
		t.Error("a literal is a configured key")
	}
}

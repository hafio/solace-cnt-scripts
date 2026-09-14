package config

import (
	"strings"
	"testing"
)

// The pre-shared key is the one credential whose requirement differs by platform, and the
// asymmetry is the operator's rather than a preference: the Solace operator invents a key
// and distributes it to the pods it owns when the CR names no Secret. Nothing does that for
// three container hosts -- each reads its own env file -- so the group would simply fail to
// form, with no obvious cause. These pin both halves.

// pskCfg is a valid HA config for p, with the pre-shared key under the caller's control.
func pskCfg(p Platform, psk, pskEnv string) *Config {
	c := validContainerConfig(p, "true")
	c.Redundancy.PSK = psk
	c.Redundancy.PSKEnv = pskEnv
	return c
}

// TestPSKIsMandatoryOnContainers. This tool used to generate the key on a first HA deploy
// and rewrite the env file; that is gone, so the operator makes it -- and the error is the
// only place they will be told how. It must therefore carry the whole instruction, not just
// the field name: the command, the same-value-on-all-three-hosts rule, and the fact that
// Kubernetes does not need it. An error that only says "must not be empty" leaves someone
// to guess a key format for a credential that has to match on two other machines.
func TestPSKIsMandatoryOnContainers(t *testing.T) {
	for _, p := range []Platform{Docker, Podman} {
		c := pskCfg(p, "", "")
		err := c.Validate(p)
		if err == nil {
			t.Fatalf("%s: an HA group with no pre-shared key must be refused", p)
		}
		for _, want := range []string{
			"redundancy.psk",          // the field, as the env file spells it
			"openssl rand -base64 32", // the command, copy-pasteable
			"SAME value",              // why one host generating its own is not enough
			"all three hosts",         //
			string(p),                 // which platform is refusing
			"Kubernetes",              // and where the rule does not apply
		} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s: error must contain %q, got:\n%s", p, want, err)
			}
		}
	}
}

// TestPSKWhitespaceIsNotAKey: a psk of blanks is the shape a half-finished edit leaves
// behind, and it would otherwise satisfy a bare non-empty test while giving the group a key
// that cannot match anything typed on another host.
func TestPSKWhitespaceIsNotAKey(t *testing.T) {
	c := pskCfg(Docker, "   ", "")
	if err := c.Validate(Docker); err == nil {
		t.Error("a whitespace-only psk must be refused like an empty one")
	}
}

// TestNothingGeneratesThePSK is the deletion's own guard. The generator was unreachable
// once the key became mandatory at load -- load rejected the file before the deploy that
// would have filled it in could run -- so it was removed rather than left as dead code that
// the docs kept promising. If a future change reintroduces "we will make one for you", this
// test is where the contradiction surfaces: the two cannot both be true.
func TestNothingGeneratesThePSK(t *testing.T) {
	c := pskCfg(Podman, "", "")
	err := c.Validate(Podman)
	if err == nil {
		t.Fatal("an empty psk must be refused, not filled in")
	}
	for _, forbidden := range []string{"will generate", "generates one", "writes it back", "first HA run"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Errorf("the error must not promise generation (%q):\n%s", forbidden, err)
		}
	}
}

// TestPSKEnvSatisfiesTheContainerRequirement: the reference form is a configured key. Load
// resolves pskEnv into PSK before validation, but a config that never went through Load
// carries only the variable name -- and "a key was configured" is what the rule asks.
func TestPSKEnvSatisfiesTheContainerRequirement(t *testing.T) {
	c := pskCfg(Docker, "", "SOLACE_REDUNDANCY_PSK")
	if err := c.Validate(Docker); err != nil {
		t.Errorf("pskEnv names a key, so the requirement is met: %v", err)
	}
}

// TestPSKIsNotRequiredStandalone: with no group there is nothing to authenticate to.
func TestPSKIsNotRequiredStandalone(t *testing.T) {
	c := validContainerConfig(Docker, "false")
	c.Redundancy.PSK = ""
	if err := c.Validate(Docker); err != nil {
		t.Errorf("a standalone broker needs no pre-shared key: %v", err)
	}
}

// --- the standalone routername ---------------------------------------------
//
// redundancy.<role>.name is the broker's routername and the container's hostname. In HA
// all three are mandatory and cannot be defaulted: each is the KEY of that node's entry
// in the group table EVERY host renders, and a host knows its own name and no other
// machine's, so a guess would build a table the other two hosts disagree with. Standalone
// has no table and exactly one node, which is always this host -- so the host's own name
// is not a guess there, it is the answer.

func TestFillStandaloneNodeNameUsesTheHost(t *testing.T) {
	c := validContainerConfig(Docker, "false")
	c.Redundancy.Primary.Name = ""
	if !c.FillStandaloneNodeName("solace-host-7") {
		t.Fatal("an empty standalone name must be filled from the host")
	}
	if c.Redundancy.Primary.Name != "solace-host-7" {
		t.Errorf("primary.name = %q, want the host's own name", c.Redundancy.Primary.Name)
	}
}

func TestFillStandaloneNodeNameKeepsAConfiguredName(t *testing.T) {
	c := validContainerConfig(Docker, "false")
	c.Redundancy.Primary.Name = "chosen-name"
	if c.FillStandaloneNodeName("solace-host-7") {
		t.Error("a configured routername must win over the hostname")
	}
	if c.Redundancy.Primary.Name != "chosen-name" {
		t.Errorf("primary.name = %q, want the configured name untouched", c.Redundancy.Primary.Name)
	}
}

// TestFillStandaloneNodeNameNeverFillsHA is the guard that matters. Filling one in here
// would produce a group table naming this host where the other two expect the configured
// node -- three hosts each rendering a different table, and a group that never forms.
func TestFillStandaloneNodeNameNeverFillsHA(t *testing.T) {
	c := validContainerConfig(Docker, "true")
	c.Redundancy.Primary.Name = ""
	if c.FillStandaloneNodeName("solace-host-7") {
		t.Fatal("HA names must never be defaulted from the local hostname")
	}
	if c.Redundancy.Primary.Name != "" {
		t.Errorf("primary.name = %q, want it left empty so validation refuses the file", c.Redundancy.Primary.Name)
	}
}

// TestFillStandaloneNodeNameIgnoresAnEmptyHostname: os.Hostname can return an empty
// string, and writing that in would swap a missing routername for a blank one.
func TestFillStandaloneNodeNameIgnoresAnEmptyHostname(t *testing.T) {
	c := validContainerConfig(Docker, "false")
	c.Redundancy.Primary.Name = ""
	if c.FillStandaloneNodeName("") {
		t.Error("an empty hostname is not a name")
	}
}

// TestPSKIsOptionalOnKubernetes is the other half. Refusing it there would demand a key the
// operator is about to generate anyway, which is why the requirement lives in
// validateContainer rather than in the shared set.
func TestPSKIsOptionalOnKubernetes(t *testing.T) {
	c := validK8sConfig()
	c.Redundancy.Enabled = "true"
	if err := c.Validate(K8s); err != nil {
		t.Fatalf("an HA Kubernetes deployment with no psk must validate: %v", err)
	}

	// Set, it is honoured rather than ignored -- it becomes the preshared_auth_key entry
	// of the admin Secret and the CR points at it.
	c.Redundancy.PSK = "explicit-key"
	if err := c.Validate(K8s); err != nil {
		t.Errorf("an explicit psk is equally valid on Kubernetes: %v", err)
	}
}

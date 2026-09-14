package config

import (
	"strings"
	"testing"
)

// A broker node gets its data volume one of two ways: provisioned by the operator from a
// StorageClass, or mounted from a PersistentVolumeClaim that already exists. These pin the
// rules that keep those two stories from being told at once, and the reason each rule errs
// the way it does.

func storageCfg(class string, redundancy string, mounts map[string]string) *Config {
	c := &Config{}
	c.Redundancy.Enabled = redundancy
	c.K8s.Name = "dev-broker"
	c.K8s.Namespace = "solace"
	c.K8s.Runtime = Command{"kubectl"}
	c.K8s.UpdateStrategy = "automatedRolling" // validateK8s checks it beyond the required set
	c.K8s.Storage.Class = class
	c.K8s.Storage.CustomVolumeMount = mounts
	c.Image.Repo = "solace-pubsub-standard"
	c.Image.Tag = "10.10.1.128"
	c.SEMP.AdminPass = "CHANGE-ME-admin"
	return c
}

// TestStorageClassAndCustomMountAreMutuallyExclusive is the headline rule. Both fields
// describe where the data volume comes from, the CRD does not say which one wins, and a
// tool that picked for you would be guessing about persistent data.
func TestStorageClassAndCustomMountAreMutuallyExclusive(t *testing.T) {
	c := storageCfg("standard", "false", map[string]string{"primary": "my-pvc"})
	err := c.validateStorage()
	if err == nil {
		t.Fatal("class + customVolumeMount must be refused")
	}
	for _, want := range []string{"kubernetes.storage.class", "customVolumeMount", "standard"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should name %q, so the fix is readable off it", err, want)
		}
	}
}

// TestCustomMountMustCoverEveryNode: a half-covered HA group is far more likely to be a
// mistyped key than a deliberate mix, and the failure mode is silent -- the uncovered node
// provisions from the cluster default and nothing says so.
func TestCustomMountMustCoverEveryNode(t *testing.T) {
	c := storageCfg("", "true", map[string]string{"primary": "pvc-p", "backup": "pvc-b"})
	err := c.validateStorage()
	if err == nil {
		t.Fatal("an HA group with only two of three nodes covered must be refused")
	}
	if !strings.Contains(err.Error(), "customVolumeMount.monitor") {
		t.Errorf("error %q should name the missing role, not just say the set is wrong", err)
	}

	full := storageCfg("", "true", map[string]string{"primary": "pvc-p", "backup": "pvc-b", "monitor": "pvc-m"})
	if err := full.validateStorage(); err != nil {
		t.Errorf("a fully covered HA group must be accepted: %v", err)
	}
}

// TestCustomMountIgnoresRolesOutsideTheGroup: a backup entry in a standalone file names a
// node that does not exist. That is harmless, and it is exactly what a file switched from
// HA to standalone looks like -- refusing it would make the switch a two-step edit for no
// gain.
func TestCustomMountIgnoresRolesOutsideTheGroup(t *testing.T) {
	c := storageCfg("", "false", map[string]string{
		"primary": "pvc-p", "backup": "pvc-b", "monitor": "pvc-m",
	})
	if err := c.validateStorage(); err != nil {
		t.Errorf("backup/monitor entries in a standalone file are ignored, not refused: %v", err)
	}
	if _, ok := c.K8s.Storage.CustomMountFor(Primary); !ok {
		t.Error("the primary's claim must still be found")
	}
}

// TestCustomMountRejectsAnUnknownRoleKey: the dangerous typo. A key that is not a role
// would leave the node it was meant for on default provisioning, quietly, on a broker
// whose storage someone deliberately took control of.
func TestCustomMountRejectsAnUnknownRoleKey(t *testing.T) {
	for _, key := range []string{"primry", "p", "Primary"} {
		c := storageCfg("", "false", map[string]string{key: "pvc-p"})
		err := c.validateStorage()
		if err == nil {
			t.Errorf("customVolumeMount.%s must be refused: it is not a role", key)
			continue
		}
		if !strings.Contains(err.Error(), "primary") {
			t.Errorf("error %q should list the roles it does accept", err)
		}
	}
}

// TestCustomMountRejectsAnEmptyClaim: a key with no value is a half-finished edit, and
// rendering it would emit a claimName the CRD requires and Kubernetes would reject.
func TestCustomMountRejectsAnEmptyClaim(t *testing.T) {
	c := storageCfg("", "false", map[string]string{"primary": "  "})
	if err := c.validateStorage(); err == nil {
		t.Fatal("an empty claim name must be refused")
	}
}

// TestMsgNodeSizeIsOptionalOnlyWithCustomMounts: a size is what the operator asks a
// StorageClass to provision. With custom mounts the volumes already exist, so demanding one
// would be demanding a number that changes nothing -- but without them it is still the one
// storage field that cannot be defaulted.
func TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(t *testing.T) {
	bare := storageCfg("standard", "false", nil)
	err := bare.validateK8s()
	if err == nil || !strings.Contains(err.Error(), "kubernetes.storage.msgNodeSize") {
		t.Fatalf("without custom mounts msgNodeSize is mandatory, got %v", err)
	}

	custom := storageCfg("", "false", map[string]string{"primary": "pvc-p"})
	if err := custom.validateK8s(); err != nil {
		t.Errorf("with every node custom-mounted there is nothing to size: %v", err)
	}
}

// TestRoleWordAndTitleAgree pins the two spellings of one concept against each other. The
// env file keys on the lowercase word; the CRD constrains customVolumeMount[].name to a
// capitalised enum. They are derived rather than typed twice, and this is what says so.
func TestRoleWordAndTitleAgree(t *testing.T) {
	for _, tc := range []struct {
		role        Role
		word, title string
	}{
		{Primary, "primary", "Primary"},
		{Backup, "backup", "Backup"},
		{Monitor, "monitor", "Monitor"},
	} {
		if got := tc.role.Word(); got != tc.word {
			t.Errorf("Word() = %q, want %q", got, tc.word)
		}
		if got := tc.role.Title(); got != tc.title {
			t.Errorf("Title() = %q, want %q -- the CRD enum is capitalised", got, tc.title)
		}
	}
}

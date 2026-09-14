package k8s

import (
	"testing"

	"solace/internal/config"
)

// These build a Config literally rather than through config.Load, so Runtime is
// set here to what ApplyDefaults would have filled in -- without it every
// command would run with an empty argv[0].
func haCfg() *config.Config {
	return &config.Config{Redundancy: config.Redundancy{Enabled: "true"}, K8s: config.K8sConfig{
		Runtime: config.Command{"kubectl"}, Name: "dev-broker", Namespace: "solace",
	}}
}

func saCfg() *config.Config {
	return &config.Config{Redundancy: config.Redundancy{Enabled: "false"}, K8s: config.K8sConfig{
		Runtime: config.Command{"kubectl"}, Name: "dev-broker", Namespace: "solace",
	}}
}

func TestResourceNames(t *testing.T) {
	cfg := haCfg()
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"pod primary", podName(cfg, config.Primary), "dev-broker-pubsubplus-p-0"},
		{"pod backup", podName(cfg, config.Backup), "dev-broker-pubsubplus-b-0"},
		{"pod monitor", podName(cfg, config.Monitor), "dev-broker-pubsubplus-m-0"},
		{"pvc primary", pvcName(cfg, config.Primary), "data-dev-broker-pubsubplus-p-0"},
		{"pvc monitor", pvcName(cfg, config.Monitor), "data-dev-broker-pubsubplus-m-0"},
		{"sts primary", stsName(cfg, config.Primary), "dev-broker-pubsubplus-p"},
		{"sts backup", stsName(cfg, config.Backup), "dev-broker-pubsubplus-b"},
		{"lb service", lbServiceName(cfg), "dev-broker-pubsubplus"},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q", tc.name, tc.got, tc.want)
		}
	}
}

func TestHARoles(t *testing.T) {
	if got := HARoles(haCfg()); len(got) != 3 || got[0] != config.Primary || got[1] != config.Backup || got[2] != config.Monitor {
		t.Errorf("HARoles(HA) = %v, want [p b m]", got)
	}
	if got := HARoles(saCfg()); len(got) != 1 || got[0] != config.Primary {
		t.Errorf("HARoles(standalone) = %v, want [p]", got)
	}
}

// TestRestartOrder pins the safe bounce order for a manual pod restart: the
// monitor carries no messaging traffic, the backup is standby, and the node most
// likely to be serving clients goes last.
func TestRestartOrder(t *testing.T) {
	if got := RestartOrder(haCfg()); len(got) != 3 || got[0] != config.Monitor || got[1] != config.Backup || got[2] != config.Primary {
		t.Errorf("RestartOrder(HA) = %v, want [m b p]", got)
	}
	if got := RestartOrder(saCfg()); len(got) != 1 || got[0] != config.Primary {
		t.Errorf("RestartOrder(standalone) = %v, want [p]", got)
	}
}

func TestProductKeyRoles(t *testing.T) {
	if got := ProductKeyRoles(haCfg()); len(got) != 2 || got[0] != config.Primary || got[1] != config.Backup {
		t.Errorf("ProductKeyRoles(HA) = %v, want [p b]", got)
	}
	if got := ProductKeyRoles(saCfg()); len(got) != 1 || got[0] != config.Primary {
		t.Errorf("ProductKeyRoles(standalone) = %v, want [p]", got)
	}
}

// TestPodNameSuffixMatchesTheConfigBound is the drift guard between two packages that
// deliberately do not import each other.
//
// internal/config bounds kubernetes.name by the length of the operator's longest
// suffix, and it spells that suffix out rather than importing it -- config is the leaf
// every platform package reads, so the dependency cannot go the other way. This derives
// a real pod name and checks the length config assumes, so a change to brokerSuffix or
// to the role letters fails here instead of silently loosening a bound.
func TestPodNameSuffixMatchesTheConfigBound(t *testing.T) {
	cfg := haCfg()
	const name = "b" // a one-character name isolates the suffix's own length
	cfg.K8s.Name = name

	// The pod is the longest derived name, and the one whose length actually binds:
	// Kubernetes copies a pod's name into the statefulset.kubernetes.io/pod-name
	// label, and a label value stops at 63.
	longest := 0
	for _, role := range []config.Role{config.Primary, config.Backup, config.Monitor} {
		if got := len(podName(cfg, role)) - len(name); got > longest {
			longest = got
		}
		// Every derived name must be no longer than the pod's, or config is bounding
		// the wrong one.
		for label, derived := range map[string]string{
			"statefulset": stsName(cfg, role),
			"lb service":  lbServiceName(cfg),
		} {
			if len(derived) > len(podName(cfg, role)) {
				t.Errorf("the %s name %q is longer than the pod name %q, so the bound in internal/config is "+
					"computed from the wrong derivation", label, derived, podName(cfg, role))
			}
		}
	}
	if longest != len(config.BrokerPodSuffixShape()) {
		t.Errorf("the operator's pod suffix is %d characters but internal/config assumes %d (%q); the bound on "+
			"kubernetes.name is now wrong in one direction or the other",
			longest, len(config.BrokerPodSuffixShape()), config.BrokerPodSuffixShape())
	}
}

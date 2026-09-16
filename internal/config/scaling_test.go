package config

import (
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// tierOrder is the ascending key order scalingTierList claims. Keeping it here
// rather than sorting the map lets TestScalingTierListMatchesTable catch both a
// list that drifts from the table and a tier added to one but not the other.
var tierOrder = []int{100, 1000, 10000, 100000, 200000}

func TestScalingTiers(t *testing.T) {
	cases := []struct {
		maxConnections int
		cpu, mem       string
	}{
		{100, "2", "3410Mi"},
		{1000, "2", "6898Mi"},
		{10000, "4", "12435Mi"},
		{100000, "8", "30925Mi"},
		{200000, "12", "52581Mi"},
	}
	if len(cases) != len(scalingTiers) {
		t.Fatalf("scalingTiers has %d entries, this test pins %d", len(scalingTiers), len(cases))
	}
	for _, tc := range cases {
		got, ok := tierFor(tc.maxConnections)
		if !ok {
			t.Errorf("tierFor(%d) reported no tier", tc.maxConnections)
			continue
		}
		if got.cpu != tc.cpu || got.mem != tc.mem {
			t.Errorf("tierFor(%d) = {cpu:%q mem:%q}, want {cpu:%q mem:%q}",
				tc.maxConnections, got.cpu, got.mem, tc.cpu, tc.mem)
		}
	}
}

// TestTierForRejectsOffTierValues covers the deliberate absence of rounding: a
// value between tiers has no published sizing, so it resolves to nothing rather
// than to the neighbour above or below it.
func TestTierForRejectsOffTierValues(t *testing.T) {
	for _, v := range []int{0, 1, 99, 101, 500, 999, 5000, 50000, 150000, 200001, -100} {
		if _, ok := tierFor(v); ok {
			t.Errorf("tierFor(%d) resolved a tier; only %s are tiers", v, scalingTierList)
		}
	}
}

func TestScalingTierListMatchesTable(t *testing.T) {
	if len(tierOrder) != len(scalingTiers) {
		t.Fatalf("scalingTiers has %d entries, tierOrder has %d", len(scalingTiers), len(tierOrder))
	}
	parts := make([]string, len(tierOrder))
	for i, v := range tierOrder {
		if i > 0 && v <= tierOrder[i-1] {
			t.Fatalf("tierOrder is not ascending at index %d", i)
		}
		if _, ok := scalingTiers[v]; !ok {
			t.Errorf("tierOrder names %d, which is not in scalingTiers", v)
		}
		parts[i] = strconv.Itoa(v)
	}
	if want := strings.Join(parts, ", "); scalingTierList != want {
		t.Errorf("scalingTierList = %q, want %q", scalingTierList, want)
	}
}

// TestContainerMem pins the one rewrite between the two memory spellings this
// schema carries: Kubernetes' Mi against the bare m docker and podman accept.
func TestContainerMem(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"3410Mi", "3410m"},
		{"6898Mi", "6898m"},
		{"12435Mi", "12435m"},
		{"30925Mi", "30925m"},
		{"52581Mi", "52581m"},
		{"4Gi", "4g"},
		{"512m", "512m"}, // already a container value: untouched
		{"", ""},
	} {
		if got := containerMem(tc.in); got != tc.want {
			t.Errorf("containerMem(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
	// Every tier's rewrite must satisfy the validator the same value would face
	// coming from an env file -- otherwise the default itself would be rejected.
	for _, v := range tierOrder {
		mem := containerMem(scalingTiers[v].mem)
		if !containerMemRE.MatchString(mem) {
			t.Errorf("tier %d default mem %q is not accepted by containerMemRE", v, mem)
		}
	}
}

func TestApplyScalingTierDefaultsK8s(t *testing.T) {
	c := &Config{}
	c.Scaling.MaxConnections = 10000
	c.ApplyDefaults(K8s)

	if c.Scaling.CPU != "4" {
		t.Errorf("Scaling.CPU = %q, want 4", c.Scaling.CPU)
	}
	if c.K8s.MsgNode.Mem != "12435Mi" {
		t.Errorf("MsgNode.Mem = %q, want 12435Mi", c.K8s.MsgNode.Mem)
	}
	// The removed key stays empty so validateK8s can treat any value as user-set.
	if c.K8s.MsgNode.CPU != "" {
		t.Errorf("MsgNode.CPU = %q, want empty", c.K8s.MsgNode.CPU)
	}
}

// TestApplyScalingTierDefaultsMemOverride pins the asymmetry the whole change
// rests on: memory is a default, CPU is not.
func TestApplyScalingTierDefaultsMemOverride(t *testing.T) {
	c := &Config{}
	c.Scaling.MaxConnections = 200000
	c.K8s.MsgNode.Mem = "64Gi"
	c.ApplyDefaults(K8s)

	if c.K8s.MsgNode.Mem != "64Gi" {
		t.Errorf("MsgNode.Mem = %q, want the explicit 64Gi to survive", c.K8s.MsgNode.Mem)
	}
	if c.Scaling.CPU != "12" {
		t.Errorf("Scaling.CPU = %q, want the tier's 12 regardless of the mem override", c.Scaling.CPU)
	}

	// Same for a container block.
	d := &Config{}
	d.Scaling.MaxConnections = 100000
	d.Docker.Container.Mem = "24g"
	d.Docker.Container.CPUSet = "8-15"
	d.ApplyDefaults(Docker)
	if d.Docker.Container.Mem != "24g" {
		t.Errorf("Docker.Container.Mem = %q, want the explicit 24g to survive", d.Docker.Container.Mem)
	}
	// Which cpus is the operator's too: the tier says how many, not which.
	if d.Docker.Container.CPUSet != "8-15" {
		t.Errorf("Docker.Container.CPUSet = %q, want the explicit 8-15 to survive", d.Docker.Container.CPUSet)
	}
	if d.Scaling.CPU != "8" {
		t.Errorf("Scaling.CPU = %q, want 8", d.Scaling.CPU)
	}
}

// TestApplyScalingTierDefaultsContainerBlocks mirrors applyContainerDefaults'
// existing parity: both blocks are filled whichever container platform is active.
func TestApplyScalingTierDefaultsContainerBlocks(t *testing.T) {
	for _, p := range []Platform{Docker, Podman} {
		c := &Config{}
		c.Scaling.MaxConnections = 10000
		c.ApplyDefaults(p)
		if c.Docker.Container.Mem != "12435m" || c.Podman.Container.Mem != "12435m" {
			t.Errorf("%s: container mem = docker %q / podman %q, want 12435m for both",
				p, c.Docker.Container.Mem, c.Podman.Container.Mem)
		}
		if c.Docker.Container.CPUSet != "0-3" || c.Podman.Container.CPUSet != "0-3" {
			t.Errorf("%s: container cpuset = docker %q / podman %q, want 0-3 for both",
				p, c.Docker.Container.CPUSet, c.Podman.Container.CPUSet)
		}
		// Still derived, and now read by the Kubernetes CR alone.
		if c.Scaling.CPU != "4" {
			t.Errorf("%s: Scaling.CPU = %q, want 4", p, c.Scaling.CPU)
		}
	}
}

// TestApplyScalingTierDefaultsOffTier pins the fail-safe: an unresolvable tier
// derives nothing rather than inventing a footprint, so validateScalingTier's
// error is what the operator sees instead of a plausible-looking artifact.
func TestApplyScalingTierDefaultsOffTier(t *testing.T) {
	c := &Config{}
	c.Scaling.MaxConnections = 12345
	c.ApplyDefaults(K8s)
	if c.Scaling.CPU != "" || c.K8s.MsgNode.Mem != "" {
		t.Errorf("off-tier derived cpu=%q mem=%q, want both empty", c.Scaling.CPU, c.K8s.MsgNode.Mem)
	}
	if err := c.Validate(K8s); err == nil || !strings.Contains(err.Error(), "scaling.maxConnections must be one of") {
		t.Errorf("expected the tier error, got: %v", err)
	}

	// Same on a container platform: no cpuset rather than an invalid one.
	d := &Config{}
	d.Scaling.MaxConnections = 12345
	d.ApplyDefaults(Docker)
	if d.Docker.Container.Mem != "" || d.Docker.Container.CPUSet != "" {
		t.Errorf("off-tier derived mem=%q cpuset=%q, want both empty",
			d.Docker.Container.Mem, d.Docker.Container.CPUSet)
	}
}

func TestValidateScalingTierRejectsOffTier(t *testing.T) {
	for _, v := range []int{0, 500, 5000, 200001} {
		c := validK8sConfig()
		c.Scaling.MaxConnections = v
		err := c.Validate(K8s)
		if err == nil || !strings.Contains(err.Error(), "scaling.maxConnections must be one of") {
			t.Errorf("maxConnections %d: expected the tier error, got: %v", v, err)
		}
		if err != nil && !strings.Contains(err.Error(), scalingTierList) {
			t.Errorf("maxConnections %d: error should list the tiers, got: %v", v, err)
		}
		// Every platform reads this value, so every platform rejects it.
		for _, p := range []Platform{Docker, Podman} {
			cc := validContainerConfig(p, "true")
			cc.Scaling.MaxConnections = v
			if err := cc.Validate(p); err == nil || !strings.Contains(err.Error(), "scaling.maxConnections must be one of") {
				t.Errorf("%s maxConnections %d: expected the tier error, got: %v", p, v, err)
			}
		}
	}
}

func TestValidateScalingTierAcceptsEveryTier(t *testing.T) {
	for _, v := range tierOrder {
		c := validK8sConfig()
		c.Scaling.MaxConnections = v
		c.ApplyDefaults(K8s) // re-derive: the fixture defaulted at 100
		if err := c.Validate(K8s); err != nil {
			t.Errorf("k8s tier %d rejected: %v", v, err)
		}
		for _, p := range []Platform{Docker, Podman} {
			cc := validContainerConfig(p, "true")
			cc.Scaling.MaxConnections = v
			cc.ApplyDefaults(p)
			if err := cc.Validate(p); err != nil {
				t.Errorf("%s tier %d rejected: %v", p, v, err)
			}
		}
	}
}

// TestValidateK8sMsgNodeCPURemoved mirrors TestValidateDockerRunModeRemoved: the
// key still decodes, so the operator gets a reason rather than "field not found".
func TestValidateK8sMsgNodeCPURemoved(t *testing.T) {
	c := validK8sConfig()
	c.K8s.MsgNode.CPU = "4"
	err := c.Validate(K8s)
	if err == nil || !strings.Contains(err.Error(), "was removed") {
		t.Fatalf("expected the msgNode.cpu removal error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "scaling.maxConnections") {
		t.Errorf("removal error should name scaling.maxConnections, got: %v", err)
	}
	if !strings.Contains(err.Error(), "kubernetes.msgNode.mem") {
		t.Errorf("removal error should say mem is unaffected, got: %v", err)
	}
}

// TestValidateMaxPoolRemoved covers the second key this change folded away:
// maxPool and maxSpoolUsageMB named one broker setting under two platform-
// specific names, which is exactly what the scaling block no longer has.
func TestValidateMaxPoolRemoved(t *testing.T) {
	for _, tc := range []struct {
		p   Platform
		cfg func() *Config
	}{
		{K8s, validK8sConfig},
		{Docker, func() *Config { return validContainerConfig(Docker, "true") }},
		{Podman, func() *Config { return validContainerConfig(Podman, "true") }},
	} {
		c := tc.cfg()
		c.Scaling.MaxPool = 10000
		err := c.Validate(tc.p)
		if err == nil || !strings.Contains(err.Error(), "was removed") {
			t.Fatalf("%s: expected the maxPool removal error, got: %v", tc.p, err)
		}
		if !strings.Contains(err.Error(), "scaling.maxSpoolUsageMB") {
			t.Errorf("%s: removal error should name the replacement, got: %v", tc.p, err)
		}
	}
	// Zero is unset, not a value: the sentinel must not fire on every config.
	if err := validK8sConfig().Validate(K8s); err != nil {
		t.Errorf("an unset maxPool must not trip the removal error: %v", err)
	}
}

func TestValidateContainerMem(t *testing.T) {
	for _, p := range []Platform{Docker, Podman} {
		// The likely mistake: the Kubernetes spelling copied across.
		c := validContainerConfig(p, "true")
		setContainerMem(c, p, "3410Mi")
		err := c.Validate(p)
		if err == nil || !strings.Contains(err.Error(), ".container.mem") {
			t.Fatalf("%s: expected a container.mem format error, got: %v", p, err)
		}
		if !strings.Contains(err.Error(), "Mi/Gi") {
			t.Errorf("%s: the error should name the Mi/Gi trap, got: %v", p, err)
		}

		for _, bad := range []string{"6898", "6898mb", "6g b", "-1g", "1.5g", "lots"} {
			c := validContainerConfig(p, "true")
			setContainerMem(c, p, bad)
			if err := c.Validate(p); err == nil {
				t.Errorf("%s: container.mem %q was accepted", p, bad)
			}
		}
		for _, good := range []string{"6898m", "512M", "2g", "4G", "1024k", "536870912b", ""} {
			c := validContainerConfig(p, "true")
			setContainerMem(c, p, good)
			if err := c.Validate(p); err != nil {
				t.Errorf("%s: container.mem %q was rejected: %v", p, good, err)
			}
		}
	}
}

func setContainerMem(c *Config, p Platform, mem string) {
	if p == Podman {
		c.Podman.Container.Mem = mem
		return
	}
	c.Docker.Container.Mem = mem
}

func setContainerCPUSet(c *Config, p Platform, set string) {
	if p == Podman {
		c.Podman.Container.CPUSet = set
		return
	}
	c.Docker.Container.CPUSet = set
}

// TestCPUSetRange pins the tier-count -> cpuset rewrite, including the fail-safe:
// anything that is not a positive integer yields "", which setDefault reads as no
// default at all rather than as an invalid cpuset the engine would reject.
func TestCPUSetRange(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"2", "0-1"}, {"4", "0-3"}, {"8", "0-7"}, {"12", "0-11"},
		{"1", "0"}, {" 4 ", "0-3"},
		{"0", ""}, {"-1", ""}, {"", ""}, {"500m", ""}, {"two", ""},
	} {
		if got := cpuSetRange(tc.in); got != tc.want {
			t.Errorf("cpuSetRange(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
	// Every tier's own default must satisfy the validator an env file value faces,
	// or the default itself would be refused. This is what catches a tier added to
	// the table with a cpu this cannot turn into a cpuset.
	for _, v := range []int{100, 1000, 10000, 100000, 200000} {
		set := cpuSetRange(scalingTiers[v].cpu)
		if !cpuSetRE.MatchString(set) {
			t.Errorf("tier %d derives cpuset %q, which validateContainer would reject", v, set)
		}
	}
}

// TestApplyScalingTierDefaultsCPUSetPerTier drives the published table through the
// real ApplyDefaults, on both container platforms, and checks it disturbs neither
// Scaling.CPU (the Kubernetes CR still renders from it) nor the k8s memory default.
func TestApplyScalingTierDefaultsCPUSetPerTier(t *testing.T) {
	for _, tc := range []struct {
		conns  int
		cpu    string
		cpuset string
	}{
		{100, "2", "0-1"}, {1000, "2", "0-1"}, {10000, "4", "0-3"},
		{100000, "8", "0-7"}, {200000, "12", "0-11"},
	} {
		for _, p := range []Platform{Docker, Podman} {
			c := &Config{}
			c.Scaling.MaxConnections = tc.conns
			c.ApplyDefaults(p)
			if c.ContainerBlock(p).CPUSet != tc.cpuset {
				t.Errorf("%s tier %d: cpuset = %q, want %q", p, tc.conns, c.ContainerBlock(p).CPUSet, tc.cpuset)
			}
			if c.Scaling.CPU != tc.cpu {
				t.Errorf("%s tier %d: Scaling.CPU = %q, want %q", p, tc.conns, c.Scaling.CPU, tc.cpu)
			}
			if c.K8s.MsgNode.Mem != "" {
				t.Errorf("%s tier %d: a container platform filled K8s.MsgNode.Mem = %q",
					p, tc.conns, c.K8s.MsgNode.Mem)
			}
		}
	}
}

func TestValidateContainerCPUSet(t *testing.T) {
	for _, p := range []Platform{Docker, Podman} {
		// The two likely mistakes: a core count, and a memlock-style -1.
		for _, bad := range []string{"-1", "4 cores", "0-", ",0", "0,,1", "0 - 3", "0-3,", "all", "1.5", "0x1", "$(id -u)"} {
			c := validContainerConfig(p, "true")
			setContainerCPUSet(c, p, bad)
			if err := c.Validate(p); err == nil {
				t.Errorf("%s: container.cpuset %q was accepted", p, bad)
			}
		}
		// A backwards range passes the charset and is refused on its own terms.
		c := validContainerConfig(p, "true")
		setContainerCPUSet(c, p, "3-0")
		err := c.Validate(p)
		if err == nil || !strings.Contains(err.Error(), "backwards") {
			t.Errorf("%s: expected the backwards-range error, got: %v", p, err)
		}

		c = validContainerConfig(p, "true")
		setContainerCPUSet(c, p, "4 cores")
		if err := c.Validate(p); err == nil || !strings.Contains(err.Error(), ".container.cpuset") {
			t.Fatalf("%s: expected a container.cpuset error naming the key, got: %v", p, err)
		}

		// "8-11" is four cores that are not the tier's first four, and that is legal:
		// which cpus are free is a host fact the tier cannot know.
		for _, good := range []string{"0", "0-1", "0-11", "0,2,4", "0-3,8", "8-11", "2-2", ""} {
			c := validContainerConfig(p, "true")
			setContainerCPUSet(c, p, good)
			if err := c.Validate(p); err != nil {
				t.Errorf("%s: container.cpuset %q was rejected: %v", p, good, err)
			}
		}
	}
}

// --- Scaling.UnmarshalYAML: the dual-spelling allowlist ----------------------

// decodeScaling runs a document through the same strict decoder Load uses, up
// to but NOT including ApplyDefaults/Validate -- so an explicit 0 is still
// visible here, the same reason decodeCommand (command_test.go) stops short
// of the full Load pipeline. Every case below exercises the real schema path,
// including Scaling.UnmarshalYAML, rather than a bare Scaling built in Go.
func decodeScaling(t *testing.T, doc string) (*Config, error) {
	t.Helper()
	var c Config
	dec := yaml.NewDecoder(strings.NewReader(doc))
	dec.KnownFields(true)
	err := dec.Decode(&c)
	return &c, err
}

// TestScalingDualSpellingAliasesTheSameField pins rule B: a friendly name and
// its destination broker setting both write the one typed field -- and an
// explicit 0 survives decode (defaulting is ApplyDefaults' job, not the
// decoder's), the same property TestScalingReachesContainersAsEnv pins at the
// render layer.
func TestScalingDualSpellingAliasesTheSameField(t *testing.T) {
	for _, tc := range []struct {
		name string
		doc  string
		get  func(*Scaling) int
		want int
	}{
		{"friendly maxConnections", "scaling:\n  maxConnections: 1000\n",
			func(s *Scaling) int { return s.MaxConnections }, 1000},
		{"destination maxConnections", "scaling:\n  system_scaling_maxconnectioncount: 1000\n",
			func(s *Scaling) int { return s.MaxConnections }, 1000},
		{"friendly maxBridges", "scaling:\n  maxBridges: 25\n",
			func(s *Scaling) int { return s.MaxBridges }, 25},
		{"destination maxBridges", "scaling:\n  system_scaling_maxbridgecount: 25\n",
			func(s *Scaling) int { return s.MaxBridges }, 25},
		{"explicit zero survives decode", "scaling:\n  maxKafkaBridge: 0\n",
			func(s *Scaling) int { return s.MaxKafkaBridge }, 0},
		{"omitted key decodes to zero", "scaling:\n  maxConnections: 1000\n",
			func(s *Scaling) int { return s.MaxBridges }, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, err := decodeScaling(t, tc.doc)
			if err != nil {
				t.Fatalf("decode %q: %v", tc.doc, err)
			}
			if got := tc.get(&c.Scaling); got != tc.want {
				t.Errorf("got %d, want %d", got, tc.want)
			}
		})
	}
}

// TestScalingUnknownKeyFailsAtLoad is the property the custom decoder is most
// likely to have silently destroyed: a custom UnmarshalYAML takes over
// decoding for the whole struct, and yaml.v3's KnownFields(true) never
// reaches inside it, so without Scaling.UnmarshalYAML doing its own policing,
// BOTH cases here would decode clean, leave the real field at zero, and
// ApplyDefaults would silently size the broker at the default tier. The
// second case is a REAL broker setting this tool does not map
// (system_scaling_maxtransactedsessioncount), not a typo -- proving the
// allowlist is closed rather than open passthrough for anything
// system_scaling_*-shaped.
func TestScalingUnknownKeyFailsAtLoad(t *testing.T) {
	for _, tc := range []struct{ name, doc, want string }{
		{"typo", "scaling:\n  maxConections: 1000\n", "scaling.maxConections is not a scaling setting"},
		{"unmapped real broker setting", "scaling:\n  system_scaling_maxtransactedsessioncount: 500\n",
			"scaling.system_scaling_maxtransactedsessioncount is not a scaling setting"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTempYAML(t, tc.doc)
			_, err := Load(path, K8s)
			if err == nil {
				t.Fatalf("%s: expected an error, got none", tc.name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("%s: error = %v, want it to contain %q", tc.name, err, tc.want)
			}
			// Stays a schema error, the same shape TestLoadUnknownFieldHasNoConvertHint
			// pins for a top-level typo -- this is the same failure, one level deeper.
			if !strings.Contains(err.Error(), "parse env file") {
				t.Errorf("%s: error should keep the schema-error shape, got: %v", tc.name, err)
			}
		})
	}
}

// TestScalingBothSpellingsAtOnceFails pins rule D: a file naming a setting
// under both its friendly and destination spelling is refused rather than
// silently letting the second win, and the error names both keys.
func TestScalingBothSpellingsAtOnceFails(t *testing.T) {
	path := writeTempYAML(t, "scaling:\n  maxConnections: 1000\n  system_scaling_maxconnectioncount: 1000\n")
	_, err := Load(path, K8s)
	if err == nil {
		t.Fatal("setting both spellings of one setting should fail to load")
	}
	for _, want := range []string{"scaling.system_scaling_maxconnectioncount", "scaling.maxConnections"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should name %q, got: %v", want, err)
		}
	}
}

// TestScalingSameKeyTwiceFails covers the collateral cost of walking the
// mapping by hand: yaml.v3's own duplicate-key rejection (uniqueKeys) never
// runs on content a custom UnmarshalYAML reads itself, so this is now this
// method's job, with its own wording distinct from the two-spellings case.
func TestScalingSameKeyTwiceFails(t *testing.T) {
	_, err := decodeScaling(t, "scaling:\n  maxConnections: 1000\n  maxConnections: 2000\n")
	if err == nil {
		t.Fatal("the same scaling key set twice should fail to decode")
	}
	if !strings.Contains(err.Error(), "scaling.maxConnections is set twice") {
		t.Errorf("error should say the key is set twice, got: %v", err)
	}
}

// TestScalingDeniesDerivedFields pins the denylist: scaling.cpu is fixed by
// the maxConnections tier, so neither spelling belongs in this schema, and the
// operator gets a reason rather than the generic unknown-key message.
func TestScalingDeniesDerivedFields(t *testing.T) {
	for _, tc := range []struct{ name, doc, want string }{
		{"cpu", "scaling:\n  cpu: \"4\"\n", "scaling.cpu is fixed by the maxConnections tier"},
		{"messagingNodeCpu", "scaling:\n  messagingNodeCpu: \"4\"\n", "messagingNodeCpu is fixed by the maxConnections tier"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTempYAML(t, tc.doc)
			_, err := Load(path, K8s)
			if err == nil {
				t.Fatalf("%s: expected an error, got none", tc.name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("%s: error = %v, want it to contain %q", tc.name, err, tc.want)
			}
		})
	}
}

// TestScalingMustBeAMapping covers the one non-mapping-kind branch:
// Scaling.UnmarshalYAML must refuse a scalar or sequence the same way
// Command.UnmarshalYAML refuses a kind it does not accept, rather than
// panicking on value.Content.
func TestScalingMustBeAMapping(t *testing.T) {
	for _, tc := range []struct{ name, doc string }{
		{"scalar", "scaling: oops\n"},
		{"sequence", "scaling:\n  - 1\n  - 2\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTempYAML(t, tc.doc)
			_, err := Load(path, K8s)
			if err == nil || !strings.Contains(err.Error(), "scaling must be a mapping") {
				t.Errorf("%s: expected the mapping error, got: %v", tc.name, err)
			}
		})
	}
}

// TestScalingValueTypeErrorIsActionable: a value of the wrong type still
// surfaces yaml's own decode error, naming the offending key, rather than
// falling through to a zero value with nothing said about it.
func TestScalingValueTypeErrorIsActionable(t *testing.T) {
	path := writeTempYAML(t, "scaling:\n  maxConnections: not-a-number\n")
	_, err := Load(path, K8s)
	if err == nil {
		t.Fatal("a non-numeric scaling value should fail to load")
	}
	if !strings.Contains(err.Error(), "scaling.maxConnections") {
		t.Errorf("error should name the offending key, got: %v", err)
	}
}

// TestValidateMaxPoolRemovedThroughLoad is TestValidateMaxPoolRemoved's
// decode-path sibling. That test sets Scaling.MaxPool directly in Go and
// never decodes YAML, so it would keep passing even if maxPool had been
// dropped from scalingKeys and an env file carrying it started failing with a
// generic unknown-key message instead of validateScaling's explanation --
// exactly the regression the hazard this change was warned about. This drives
// the same key through the real decoder and Validate.
func TestValidateMaxPoolRemovedThroughLoad(t *testing.T) {
	path := writeTempYAML(t, "scaling:\n  maxPool: 10000\n")
	_, err := Load(path, K8s)
	if err == nil || !strings.Contains(err.Error(), "scaling.maxPool was removed") {
		t.Fatalf("expected the maxPool removal error through Load, got: %v", err)
	}
	if !strings.Contains(err.Error(), "scaling.maxSpoolUsageMB") {
		t.Errorf("removal error should name the replacement, got: %v", err)
	}
}

// TestRetiredContainerLimitKeysFailLoud pins the retired-key rule for the four
// container limits: the old key must still DECODE (that is what keeps the error
// about the key rather than a bare unknown field) and must then fail Validate
// naming the key and the value it is now fixed at.
func TestRetiredContainerLimitKeysFailLoud(t *testing.T) {
	cases := []struct {
		name, doc string
		p         Platform
		want      string
		fixed     string
	}{
		{"docker.shmSize", "docker:\n  container:\n    shmSize: 1g\n", Docker,
			"docker.container.shmSize was removed", "2g"},
		{"docker.nofile", "docker:\n  container:\n    ulimits:\n      nofile: \"1024:1024\"\n", Docker,
			"docker.container.ulimits.nofile was removed", "2448:1048576"},
		{"docker.memlock", "docker:\n  container:\n    ulimits:\n      memlock: \"-1\"\n", Docker,
			"docker.container.ulimits.memlock was removed", "-1"},
		{"docker.core", "docker:\n  container:\n    ulimits:\n      core: \"0\"\n", Docker,
			"docker.container.ulimits.core was removed", "-1"},
		{"podman.shmSize", "podman:\n  container:\n    shmSize: 1g\n", Podman,
			"podman.container.shmSize was removed", "2g"},
		{"podman.nofile", "podman:\n  container:\n    ulimits:\n      nofile: \"1024:1024\"\n", Podman,
			"podman.container.ulimits.nofile was removed", "2448:1048576"},
		{"podman.memlock", "podman:\n  container:\n    ulimits:\n      memlock: \"-1\"\n", Podman,
			"podman.container.ulimits.memlock was removed", "-1"},
		{"podman.core", "podman:\n  container:\n    ulimits:\n      core: \"0\"\n", Podman,
			"podman.container.ulimits.core was removed", "-1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c Config
			if err := decodeStrict(tc.doc, &c); err != nil {
				t.Fatalf("the removed key must still decode (that is the point): %v", err)
			}
			c.ApplyDefaults(tc.p)
			err := c.Validate(tc.p)
			if err == nil {
				t.Fatal("the removed key must fail validation")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to contain %q", err, tc.want)
			}
			if !strings.Contains(err.Error(), tc.fixed) {
				t.Errorf("error = %v, want it to name the fixed value %q", err, tc.fixed)
			}
			if !strings.Contains(err.Error(), "cpuset") {
				t.Errorf("error = %v, want it to name the keys that still work", err)
			}
		})
	}

	// The sentinel must not fire on every config: an empty ulimits block sets
	// nothing, and a valid fixture still has to pass. This is what would catch a
	// resurrected setDefault in applyContainerBlockDefaults.
	var empty Config
	if err := decodeStrict("docker:\n  container:\n    ulimits: {}\n", &empty); err != nil {
		t.Fatalf("an empty ulimits block must decode: %v", err)
	}
	for _, p := range []Platform{Docker, Podman} {
		if err := validContainerConfig(p, "true").Validate(p); err != nil {
			t.Errorf("%s: a valid config must not trip the removal error: %v", p, err)
		}
	}
}

// TestRetiredContainerKeysAreRefusedOnEveryPlatform pins the placement: the check
// is platform-independent, like the renamed docker.runtime, so a shared env file
// resolved to kubernetes still hears about a dead key in its docker block.
func TestRetiredContainerKeysAreRefusedOnEveryPlatform(t *testing.T) {
	var c Config
	if err := decodeStrict("docker:\n  container:\n    shmSize: 1g\n", &c); err != nil {
		t.Fatalf("decode: %v", err)
	}
	c.ApplyDefaults(K8s)
	if err := c.Validate(K8s); err == nil || !strings.Contains(err.Error(), "docker.container.shmSize was removed") {
		t.Errorf("expected the removal error on kubernetes too, got: %v", err)
	}
}

// TestRetiredContainerKeyFailsThroughLoad proves the sentinel survives the whole
// pipeline -- strict decode, ApplyDefaults, home expansion, Validate -- and not
// merely a hand-built Config. This is the test that fails if the retired defaults
// are ever repointed at their constants instead of deleted.
func TestRetiredContainerKeyFailsThroughLoad(t *testing.T) {
	path := writeTempYAML(t, `
image:
  repo: solace/solace-pubsub-standard
  tag: 10.10.1.35
semp:
  adminPass: s3cret
redundancy:
  enabled: "false"
docker:
  container:
    dataDir: /opt/solace/data
    shmSize: 1g
`)
	_, err := Load(path, Docker)
	if err == nil || !strings.Contains(err.Error(), "docker.container.shmSize was removed") {
		t.Errorf("Load should refuse the retired key by name, got: %v", err)
	}
}

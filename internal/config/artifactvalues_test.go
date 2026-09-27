package config

import (
	"strconv"
	"strings"
	"testing"
)

// artifactInjections are values that would restructure the artifact they land in if
// written as-is: a new quadlet key, a new quadlet section with a command systemd
// runs, a new YAML key, a new YAML document, and a compose interpolation that
// expands one of the secrets in the compose child's environment.
var artifactInjections = []string{
	"8080:8080\nPodmanArgs=--privileged",
	"8080:8080\n[Service]\nExecStartPre=/bin/sh -c id",
	"5s\n    cap_add: [ALL]",
	"latest\n---\nkind: Secret",
	"${SOLACE_ADMIN_PASSWORD}",
}

// TestValidateImageReference holds image.registry, image.repo and image.tag to the
// reference grammar, which is what keeps them from restructuring the quadlet unit,
// the compose file or the broker CR they are written into. Each refusal names the
// field and the value; the check is platform-neutral, so it holds on all three.
func TestValidateImageReference(t *testing.T) {
	fields := []struct {
		name string
		set  func(*Config, string)
		good []string
		bad  []string
	}{
		{"image.registry", func(c *Config, v string) { c.Image.Registry = v },
			[]string{"registry.example.com", "localhost:5000", "10.0.0.5:5000", "ghcr.io/solace", "Registry.Example.com"},
			[]string{"https://registry.example.com", "registry.example.com/", "-registry", "reg istry", "reg:port"}},
		{"image.repo", func(c *Config, v string) { c.Image.Repo = v },
			[]string{"solace-pubsub-standard", "solace/solace-pubsub-standard", "docker.io/solace/solace-pubsub-standard",
				"localhost:5000/solace/broker", "a__b", "a--b", "a.b_c"},
			[]string{"Solace/Broker", "solace//broker", "/solace", "solace/", "-solace", "solace broker", "solace:tag"}},
		{"image.tag", func(c *Config, v string) { c.Image.Tag = v },
			[]string{"latest", "10.26.1.5", "v1_2-rc.3", "_x", strings.Repeat("a", 128)},
			[]string{"-rc", ".hidden", "10:26", "sha256@abc", "1 0", strings.Repeat("a", 129)}},
	}
	for _, p := range []Platform{K8s, Docker, Podman} {
		for _, f := range fields {
			for _, v := range f.good {
				c := validPlatformConfig(p)
				f.set(c, v)
				if err := c.Validate(p); err != nil {
					t.Errorf("%s: %s = %q must be accepted: %v", p, f.name, v, err)
				}
			}
			for _, v := range append(append([]string{}, f.bad...), artifactInjections...) {
				c := validPlatformConfig(p)
				f.set(c, v)
				err := c.Validate(p)
				if err == nil {
					t.Errorf("%s: %s = %q must be refused", p, f.name, v)
					continue
				}
				for _, want := range []string{f.name, strconv.Quote(v)} {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("%s: error %q should name %s", p, err, want)
					}
				}
			}
		}
	}
}

// validPlatformConfig is the valid fixture for p from the per-platform builders.
func validPlatformConfig(p Platform) *Config {
	if p == K8s {
		return validK8sConfig()
	}
	return validContainerConfig(p, "false")
}

// TestValidContainerPort pins the one publish form this schema accepts,
// [ip:]host:container[/tcp|/udp] with optional ranges, and that everything outside
// it -- including every injection -- is refused with an error naming the entry.
func TestValidContainerPort(t *testing.T) {
	for _, good := range []string{
		"8080:8080", "55555:55555/tcp", "8000:8000/udp", "2222:22",
		"127.0.0.1:8080:8080", "[::1]:8080:8080", "[2001:db8::5]:1943:1943/tcp",
		"55000-55010:55000-55010", "8000-8010:8080",
	} {
		if err := validContainerPort("docker.network.ports[0]", good); err != nil {
			t.Errorf("%q must be accepted: %v", good, err)
		}
	}
	for _, bad := range append([]string{
		"8080", ":8080", "8080:", "0:8080", "8080:65536", "010:8080", "+80:80",
		"8080:8080/sctp", "8080:8080/TCP", "localhost:8080:8080", "::1:8080:8080", "[1.2.3.4]:80:80",
		"[::1]8080:8080", "8010-8000:8010-8000", "8000-8001:8000-8002", "8080:8000-8001",
		" 8080:8080", "8080:8080 ", "1.2.3.4:5:6:7",
	}, artifactInjections...) {
		err := validContainerPort("docker.network.ports[0]", bad)
		if err == nil {
			t.Errorf("%q must be refused", bad)
			continue
		}
		for _, want := range []string{"docker.network.ports[0]", strconv.Quote(bad), "host:container"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q should say %s", err, want)
			}
		}
	}
}

// TestValidateContainerArtifactValues covers both checks through Validate on both
// engines: a bad port entry is refused by index even in host mode (a file switched
// to bridge later carries the same list), and each health-check timing must be a
// positive Go duration -- zero allowed only for startPeriod -- with injections
// refused like any other malformed value. The defaults ApplyDefaults fills pass.
func TestValidateContainerArtifactValues(t *testing.T) {
	for _, p := range []Platform{Docker, Podman} {
		key := platformKey(p)
		if err := validContainerConfig(p, "false").Validate(p); err != nil {
			t.Fatalf("%s: the defaults must validate: %v", p, err)
		}
		for _, mode := range []string{"host", "bridge"} {
			c := validContainerConfig(p, "false")
			net := &c.Docker.Network
			if p == Podman {
				net = &c.Podman.Network
			}
			net.Mode = mode
			net.Ports = []string{"8080:8080", "8080:8080\n[Service]\nExecStartPre=/bin/true"}
			err := c.Validate(p)
			if err == nil || !strings.Contains(err.Error(), key+".network.ports[1]") {
				t.Errorf("%s %s mode: err = %v, want a refusal naming %s.network.ports[1]", p, mode, err, key)
			}
		}
		for _, tc := range []struct {
			name, value string
			ok          bool
		}{
			{"interval", "1m30s", true}, {"interval", "0s", false}, {"interval", "-5s", false},
			{"interval", "5", false}, {"timeout", "0s", false}, {"timeout", "250ms", true},
			{"startPeriod", "0s", true}, {"startPeriod", "-1s", false},
			{"startPeriod", "5s\n    cap_add: [ALL]", false},
		} {
			c := validContainerConfig(p, "false")
			hc := &c.Docker.Container.HealthCheck
			if p == Podman {
				hc = &c.Podman.Container.HealthCheck
			}
			switch tc.name {
			case "interval":
				hc.Interval = tc.value
			case "timeout":
				hc.Timeout = tc.value
			default:
				hc.StartPeriod = tc.value
			}
			field := key + ".container.healthCheck." + tc.name
			err := c.Validate(p)
			switch {
			case tc.ok && err != nil:
				t.Errorf("%s = %q must be accepted: %v", field, tc.value, err)
			case !tc.ok && (err == nil || !strings.Contains(err.Error(), field)):
				t.Errorf("%s = %q: err = %v, want a refusal naming the field", field, tc.value, err)
			}
		}
	}
}

// TestValidateK8sArtifactValues covers the Kubernetes values text/template and the CR
// renderer substitute raw: the operator image (repository, optional tag and digest),
// the operator's cpu and mem and the two storage sizes (Kubernetes quantities, no
// sign or exponent), and every watchNamespaces entry (a DNS-1123 label, after the
// same trimming the watch list gets). Each refusal names the field and the value.
func TestValidateK8sArtifactValues(t *testing.T) {
	digest := "@sha256:" + strings.Repeat("a", 64)
	for _, f := range []struct {
		name string
		set  func(*Config, string)
		good []string
		bad  []string
	}{
		{"kubernetes.operator.image", func(c *Config, v string) { c.K8s.Operator.Image = v },
			[]string{"solace/pubsubplus-eventbroker-operator:1.4.2", "solace/op", "solace/op" + digest,
				"solace/op:1.4.2" + digest, "localhost:5000/solace/op:1.4.2"},
			[]string{"Solace/Op:1", "solace/op:", "solace/op:-x", "solace/op@sha256:abc", "solace/op 1.4.2"}},
		{"kubernetes.operator.cpu", func(c *Config, v string) { c.K8s.Operator.CPU = v },
			[]string{"500m", "1", "1.5", ".5"}, []string{"-1", "1e3", "500 m", "m", "1.", "0x10"}},
		{"kubernetes.operator.mem", func(c *Config, v string) { c.K8s.Operator.Mem = v },
			[]string{"512Mi", "1Gi", "1G", "268435456"}, []string{"512MB", "512mi", "Mi", "+512Mi"}},
		{"kubernetes.storage.msgNodeSize", func(c *Config, v string) { c.K8s.Storage.MsgNodeSize = v },
			[]string{"30Gi", "1Ti", "500G"}, []string{"30 Gi", "30GiB", "-30Gi"}},
		{"kubernetes.storage.monNodeSize", func(c *Config, v string) { c.K8s.Storage.MonNodeSize = v },
			[]string{"5Gi"}, []string{"5gb"}},
	} {
		for _, v := range f.good {
			c := validK8sConfig()
			f.set(c, v)
			if err := c.Validate(K8s); err != nil {
				t.Errorf("%s = %q must be accepted: %v", f.name, v, err)
			}
		}
		for _, v := range append(append([]string{}, f.bad...), artifactInjections...) {
			c := validK8sConfig()
			f.set(c, v)
			err := c.Validate(K8s)
			if err == nil {
				t.Errorf("%s = %q must be refused", f.name, v)
				continue
			}
			for _, want := range []string{f.name, strconv.Quote(v)} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q should name %s", err, want)
				}
			}
		}
	}

	for _, good := range []string{"", "team-a", "team-a, team-b", "team-a,,team-b,"} {
		c := validK8sConfig()
		c.K8s.Operator.WatchNamespaces = good
		if err := c.Validate(K8s); err != nil {
			t.Errorf("watchNamespaces %q must be accepted: %v", good, err)
		}
	}
	for _, bad := range []string{"team-a,Team-B", "team_a", "team-a\"\n  - name: X", "team-a\n---\nkind: Secret"} {
		c := validK8sConfig()
		c.K8s.Operator.WatchNamespaces = bad
		err := c.Validate(K8s)
		if err == nil || !strings.Contains(err.Error(), "kubernetes.operator.watchNamespaces entry") {
			t.Errorf("watchNamespaces %q: err = %v, want a refusal naming the entry", bad, err)
		}
	}
}

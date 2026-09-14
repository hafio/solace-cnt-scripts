package config

import (
	"strings"
	"testing"
)

// validReplication is a complete, well-formed DR pair: one site reached over its own
// cluster CLI, the other over SEMP, and two VPNs active at opposite ends. Tests mutate
// one field of it, the way validK8sConfig is used for the rest of the schema.
//
// The two sites deliberately differ in mechanism. A pair where both ends use the same
// one would let a bug that only reads via.kubernetes pass every test.
func validReplication() Replication {
	return Replication{
		Sites: []ReplSite{
			{
				VirtualRouterName: "v:sol-a",
				RouterNames:       []string{"sol-a1", "sol-a2"},
				Endpoints: []ReplEndpoint{
					{Host: "10.0.0.1", Port: 55555},
					{Host: "10.0.0.1", Port: 55443, Transport: TransportSSL},
				},
				Via: ReplVia{Kubernetes: &ReplViaKube{
					Command:   Command{"kubectl", "--context", "site-a"},
					Namespace: "solace-a",
					Name:      "solace",
				}},
			},
			{
				VirtualRouterName: "v:sol-b",
				RouterNames:       []string{"sol-b1"},
				Endpoints: []ReplEndpoint{
					{Host: "10.0.1.1", Port: 55443, Transport: TransportSSL},
				},
				Via: ReplVia{SEMP: &ReplViaSEMP{
					Host: "10.0.1.1", Port: 1943, TLS: true, PassEnv: "MATE_PASS",
				}},
			},
		},
		VPNs: []ReplVPN{
			{Name: "ORDERS", ActiveAt: "v:sol-a"},
			{Name: "PAYMENTS", ActiveAt: "v:sol-b"},
		},
	}
}

// replConfig is validK8sConfig with a replication block attached.
func replConfig(r Replication) *Config {
	c := validK8sConfig()
	c.Replication = r
	return c
}

// TestValidateReplicationValid proves the fixture itself passes, so every rejection
// below fails for the reason it names rather than because the baseline was broken.
func TestValidateReplicationValid(t *testing.T) {
	if err := replConfig(validReplication()).Validate(K8s); err != nil {
		t.Errorf("a valid replication block must be accepted: %v", err)
	}
}

// TestValidateReplicationAbsent proves the whole section is optional. Most deployments
// have no DR pair, and an omitted block must not make every other command fail to load.
func TestValidateReplicationAbsent(t *testing.T) {
	c := validK8sConfig()
	if c.Replication.Configured() {
		t.Fatal("the base fixture must not declare replication")
	}
	if err := c.Validate(K8s); err != nil {
		t.Errorf("an absent replication block must be accepted: %v", err)
	}
}

// TestValidateReplicationRejects covers every way the block can be wrong. Each case
// names the field it is about, because an operator meets these messages while holding
// the file open at that line.
func TestValidateReplicationRejects(t *testing.T) {
	cases := []struct {
		name  string
		apply func(*Replication)
		want  string
	}{
		{"one site", func(r *Replication) {
			r.Sites = r.Sites[:1]
			r.VPNs = r.VPNs[:1]
		}, "must hold exactly 2 entries"},
		{"three sites", func(r *Replication) {
			r.Sites = append(r.Sites, r.Sites[0])
		}, "must hold exactly 2 entries"},
		{"missing virtualRouterName", func(r *Replication) {
			r.Sites[0].VirtualRouterName = ""
		}, "replication.sites[0].virtualRouterName must be set"},
		{"duplicate virtualRouterName", func(r *Replication) {
			r.Sites[1].VirtualRouterName = r.Sites[0].VirtualRouterName
		}, "the two\nsites must be distinguishable"},
		{"no routerNames", func(r *Replication) {
			r.Sites[1].RouterNames = nil
		}, "replication.sites[1].routerNames must list at least one name"},
		{"empty routerName", func(r *Replication) {
			r.Sites[0].RouterNames = []string{"sol-a1", "  "}
		}, "replication.sites[0].routerNames[1] must not be empty"},
		{"overlapping routerNames", func(r *Replication) {
			r.Sites[1].RouterNames = []string{"sol-a2"}
		}, "already declared by replication.sites[0]"},
		{"no endpoints", func(r *Replication) {
			r.Sites[1].Endpoints = nil
		}, "replication.sites[1].endpoints must list at least one address"},
		{"empty endpoint host", func(r *Replication) {
			r.Sites[0].Endpoints[0].Host = ""
		}, "replication.sites[0].endpoints[0].host must be set"},
		{"port zero", func(r *Replication) {
			r.Sites[0].Endpoints[0].Port = 0
		}, "endpoints[0].port must be between 1 and 65535"},
		{"port too high", func(r *Replication) {
			r.Sites[0].Endpoints[0].Port = 65536
		}, "endpoints[0].port must be between 1 and 65535"},
		{"transport encrypted", func(r *Replication) {
			r.Sites[0].Endpoints[0].Transport = "encrypted"
		}, "use ssl"},
		{"transport unknown", func(r *Replication) {
			r.Sites[0].Endpoints[0].Transport = "tls"
		}, "must be plainText, compressed or ssl"},
		{"three on one transport", func(r *Replication) {
			r.Sites[0].Endpoints = []ReplEndpoint{
				{Host: "a", Port: 55443, Transport: TransportSSL},
				{Host: "b", Port: 55443, Transport: TransportSSL},
				{Host: "c", Port: 55443, Transport: TransportSSL},
			}
		}, "at most 2 addresses per transport"},
		// A repeated address is a typo with a lasting cost, not a harmless one.
		// `broker configure data-replication` decides whether the mate needs converging
		// by comparing this set against what the broker reports, and the broker reports
		// each address ONCE -- so a duplicate never compares equal, and every run would
		// shut replication down on every VPN to rewrite the identical mate lines.
		{"duplicate endpoint", func(r *Replication) {
			r.Sites[0].Endpoints = []ReplEndpoint{
				{Host: "a", Port: 55443, Transport: TransportSSL},
				{Host: "a", Port: 55443, Transport: TransportSSL},
			}
		}, "repeats the address already given as replication.sites[0].endpoints[0]"},
		// The two spellings of plain text are the SAME address: the keyword is optional
		// and an absent one means plain text, so comparing the written forms would let
		// this pair through.
		{"duplicate endpoint, one spelled plain-text", func(r *Replication) {
			r.Sites[0].Endpoints = []ReplEndpoint{
				{Host: "a", Port: 55555},
				{Host: "a", Port: 55555, Transport: TransportPlainText},
			}
		}, "repeats the address already given"},
		{"via with both mechanisms", func(r *Replication) {
			r.Sites[1].Via.Kubernetes = &ReplViaKube{Command: Command{"kubectl"}, Namespace: "n", Name: "b"}
		}, "declares both kubernetes and semp"},
		{"kube without command", func(r *Replication) {
			r.Sites[0].Via.Kubernetes.Command = nil
		}, "via.kubernetes.command must be set"},
		{"kube without namespace", func(r *Replication) {
			r.Sites[0].Via.Kubernetes.Namespace = ""
		}, "via.kubernetes.namespace must be set"},
		{"kube without name", func(r *Replication) {
			r.Sites[0].Via.Kubernetes.Name = ""
		}, "via.kubernetes.name must be set"},
		{"semp without host", func(r *Replication) {
			r.Sites[1].Via.SEMP.Host = ""
		}, "via.semp.host must be set"},
		{"semp bad port", func(r *Replication) {
			r.Sites[1].Via.SEMP.Port = 0
		}, "via.semp.port must be between 1 and 65535"},
		{"semp without credential", func(r *Replication) {
			r.Sites[1].Via.SEMP.PassEnv = ""
		}, "needs the mate's admin password"},
		{"semp with two credentials", func(r *Replication) {
			r.Sites[1].Via.SEMP.Pass = "p"
		}, "declares more than one of pass, passEnv and passSecret"},
		{"passSecret missing key", func(r *Replication) {
			r.Sites[1].Via.SEMP.PassEnv = ""
			r.Sites[1].Via.SEMP.PassSecret = &ReplPassSecret{Namespace: "n", Name: "s"}
		}, "via.semp.passSecret.key must be set"},
		{"vpn without name", func(r *Replication) {
			r.VPNs[0].Name = ""
		}, "replication.vpns[0].name must be set"},
		{"duplicate vpn", func(r *Replication) {
			r.VPNs[1].Name = r.VPNs[0].Name
		}, "is listed twice"},
		{"activeAt names nothing", func(r *Replication) {
			r.VPNs[0].ActiveAt = "v:nowhere"
		}, "names no site"},
		{"activeAt is a router name", func(r *Replication) {
			r.VPNs[0].ActiveAt = "sol-a1"
		}, "not a router name"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := validReplication()
			tc.apply(&r)
			err := replConfig(r).Validate(K8s)
			if err == nil {
				t.Fatalf("expected an error containing %q, got nil", tc.want)
			}
			// The wanted text is matched with newlines folded, so a message can be
			// wrapped in the source without the test caring where the break lands.
			got := strings.Join(strings.Fields(err.Error()), " ")
			want := strings.Join(strings.Fields(tc.want), " ")
			if !strings.Contains(got, want) {
				t.Errorf("error = %q, want it to mention %q", got, want)
			}
		})
	}
}

// TestValidateReplicationEmptyViaAccepted pins the deliberate gap: `via` says how to
// reach a site when it is the MATE, and only the switch command needs it. Requiring it
// at load would refuse a file that configures replication perfectly well for the
// local-only command, so a site with no via loads and the switch refuses instead.
func TestValidateReplicationEmptyViaAccepted(t *testing.T) {
	r := validReplication()
	r.Sites[0].Via = ReplVia{}
	r.Sites[1].Via = ReplVia{}
	if err := replConfig(r).Validate(K8s); err != nil {
		t.Errorf("a site with no via must still load: %v", err)
	}
	if r.Sites[0].Via.Set() {
		t.Error("an empty ReplVia must report Set() == false")
	}
}

// TestReplicationLocate proves a broker finds itself by ANY of its site's router names
// -- the HA backup node reports its own name, not the primary's -- and that the two
// non-happy paths are errors rather than guesses. A wrong guess here would point a
// switchover at the wrong broker.
func TestReplicationLocate(t *testing.T) {
	r := validReplication()

	for _, name := range []string{"sol-a1", "sol-a2"} {
		t.Run("finds/"+name, func(t *testing.T) {
			self, mate, err := r.Locate(name)
			if err != nil {
				t.Fatalf("Locate(%q): %v", name, err)
			}
			if self.VirtualRouterName != "v:sol-a" {
				t.Errorf("self = %q, want v:sol-a", self.VirtualRouterName)
			}
			if mate.VirtualRouterName != "v:sol-b" {
				t.Errorf("mate = %q, want v:sol-b", mate.VirtualRouterName)
			}
		})
	}

	t.Run("no match names what it read and what is declared", func(t *testing.T) {
		_, _, err := r.Locate("sol-c1")
		if err == nil {
			t.Fatal("expected an error for a router name in neither site")
		}
		for _, want := range []string{"sol-c1", "sol-a1", "sol-a2", "sol-b1"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q must mention %q so the typo is visible", err, want)
			}
		}
	})

	t.Run("overlap is an error, never the first match", func(t *testing.T) {
		dup := validReplication()
		dup.Sites[1].RouterNames = []string{"sol-a1"}
		if _, _, err := dup.Locate("sol-a1"); err == nil {
			t.Fatal("an overlapping router name must not resolve to whichever site came first")
		}
	})
}

// TestSiteCommandGuarded proves the second enforcement point: a site's cluster CLI is
// re-checked when the command is fetched, not only at load, because the mate channel is
// built straight from a *Config and must not assume Validate ran.
func TestSiteCommandGuarded(t *testing.T) {
	c := replConfig(validReplication())

	got, err := c.SiteCommand("v:sol-a")
	if err != nil {
		t.Fatalf("SiteCommand: %v", err)
	}
	if got.String() != (Command{"kubectl", "--context", "site-a"}).String() {
		t.Errorf("SiteCommand = %v", got)
	}

	t.Run("rejects a binary off the allowlist", func(t *testing.T) {
		bad := validReplication()
		bad.Sites[0].Via.Kubernetes.Command = Command{"curl"}
		if _, err := replConfig(bad).SiteCommand("v:sol-a"); err == nil {
			t.Error("a site command outside the kubernetes allowlist must be refused")
		}
	})

	t.Run("rejects a shell metacharacter", func(t *testing.T) {
		bad := validReplication()
		bad.Sites[0].Via.Kubernetes.Command = Command{"kubectl;", "rm"}
		if _, err := replConfig(bad).SiteCommand("v:sol-a"); err == nil {
			t.Error("a site command carrying a metacharacter must be refused")
		}
	})

	t.Run("names the site's own field, not kubernetes.runtime", func(t *testing.T) {
		bad := validReplication()
		bad.Sites[0].Via.Kubernetes.Command = Command{"curl"}
		err := replConfig(bad).Validate(K8s)
		if err == nil || !strings.Contains(err.Error(), "replication.sites[0].via.kubernetes.command") {
			t.Errorf("Validate must name the site's own field, got: %v", err)
		}
	})

	t.Run("wrong mechanism says which one the site declares", func(t *testing.T) {
		_, err := c.SiteCommand("v:sol-b")
		if err == nil || !strings.Contains(err.Error(), "via.semp") {
			t.Errorf("expected an error naming via.semp, got: %v", err)
		}
	})

	t.Run("unknown site", func(t *testing.T) {
		if _, err := c.SiteCommand("v:nowhere"); err == nil {
			t.Error("an unknown virtualRouterName must be refused")
		}
	})
}

// TestReplicationPassEnvResolves proves the mate's password is referable from the
// environment like every other secret. Without an entry on the one secretRefs list, a
// passEnv would silently stay an unexpanded variable NAME and be sent to the mate as
// the password.
func TestReplicationPassEnvResolves(t *testing.T) {
	t.Setenv("MATE_PASS", "from-env")
	c := replConfig(validReplication())
	if err := c.resolveSecretRefs(); err != nil {
		t.Fatalf("resolveSecretRefs: %v", err)
	}
	if got := c.Replication.Sites[1].Via.SEMP.Pass; got != "from-env" {
		t.Errorf("via.semp.pass = %q, want the value of MATE_PASS", got)
	}
}

// TestReplicationCredentialChars proves the mate password gets the same charset check as
// every other credential. It reaches a curl config on stdin, where curlConfigLine does
// not escape a newline -- so a value carrying one breaks out of its line before the
// request is sent.
func TestReplicationCredentialChars(t *testing.T) {
	for _, bc := range controlCharCases {
		t.Run(bc.name, func(t *testing.T) {
			r := validReplication()
			r.Sites[1].Via.SEMP.PassEnv = ""
			r.Sites[1].Via.SEMP.Pass = bc.value
			err := replConfig(r).Validate(K8s)
			if err == nil {
				t.Fatal("a credential carrying a control character must be refused")
			}
			if strings.Contains(err.Error(), bc.value) {
				t.Errorf("the error must not echo the secret value: %v", err)
			}
		})
	}
}

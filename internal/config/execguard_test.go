package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/execguard"
	"gopkg.in/yaml.v3"
)

// decodeStrict runs a document through the same strict decoder Load uses, so the
// schema tests below exercise the real load path. The sibling of command_test.go's
// decodeCommand, which returns one field rather than the error.
func decodeStrict(doc string, c *Config) error {
	dec := yaml.NewDecoder(strings.NewReader(doc))
	dec.KnownFields(true)
	return dec.Decode(c)
}

// guardConfig builds a config that passes Validate on p with everything EXCEPT the
// command fields already correct, so a Validate failure in these tests can only
// have come from the execution guard. Defaults are applied first, exactly as
// config.Load does, so the runtime fields hold their real defaults unless a case
// overrides them.
func guardConfig(p Platform) *Config {
	c := &Config{Redundancy: Redundancy{Enabled: "false"}}
	c.Image.Repo = "solace/solace-pubsub-standard"
	c.Image.Tag = "10.10.1.35"
	c.SEMP.AdminPass = "s3cret-not-a-real-password"
	c.K8s.Name = "broker"
	c.K8s.Namespace = "solace"
	c.K8s.Storage.MsgNodeSize = "30Gi"
	c.Redundancy.Primary.Name = "primary-host"
	c.ApplyDefaults(p)
	return c
}

// TestGuardConfigIsValid guards the fixture itself: every case below reads a
// Validate failure as "the guard rejected the command", which is only sound while
// the untouched fixture validates cleanly on every platform. It walks Platforms()
// rather than a literal, so a fourth platform is covered the day it is declared
// instead of the day someone remembers this loop.
func TestGuardConfigIsValid(t *testing.T) {
	for _, p := range Platforms() {
		if err := guardConfig(p).Validate(p); err != nil {
			t.Errorf("guardConfig(%s) must validate cleanly before any command is overridden: %v", p, err)
		}
	}
}

// TestExecBinariesCoversEveryPlatform pins the allowlist against Platforms(). A
// missing key is not a compile error and not an obvious runtime one either:
// rulesFor would hand execguard an empty Allowed, so Check would refuse every
// command that platform names with a wiring error, and the "allowed" list in the
// error would be empty -- a confusing refusal rather than a loud wiring failure.
// This is the test that turns that into a build-time answer.
func TestExecBinariesCoversEveryPlatform(t *testing.T) {
	for _, p := range Platforms() {
		names, ok := execBinaries[p]
		if !ok {
			t.Errorf("execBinaries has no entry for platform %s: every command it names would be refused", p)
			continue
		}
		if len(names) == 0 {
			t.Errorf("execBinaries[%s] is empty: no command could ever be allowed on that platform", p)
		}
	}
	if len(execBinaries) != len(Platforms()) {
		t.Errorf("execBinaries has %d entries but there are %d platforms: an entry for a platform that does not exist is dead allowlist",
			len(execBinaries), len(Platforms()))
	}
}

// --- layer 5: the validator and the executor share one definition -------------

// TestValidatorAndExecutorAgree is the shared-definition test. For every case in
// the matrix it drives BOTH enforcement points -- Validate (which config.Load runs)
// and the accessor every executor calls immediately before building argv -- and
// requires the same verdict. If someone ever forks the check, this fails: that is
// the whole point, since a hostile env file must be inert on a path that never
// reached Validate.
func TestValidatorAndExecutorAgree(t *testing.T) {
	cases := []struct {
		name    string
		p       Platform
		cmd     Command
		allowed []string
		wantOK  bool
	}{
		{"kubectl accepted", K8s, Command{"kubectl"}, nil, true},
		{"oc accepted", K8s, Command{"oc"}, nil, true},
		{"kubectl.exe accepted", K8s, Command{"kubectl.exe"}, nil, true},
		{"kubectl with flags accepted", K8s, Command{"kubectl", "--context", "prod", "-n", "ns"}, nil, true},
		{"docker accepted", Docker, Command{"docker"}, nil, true},
		{"podman accepted", Podman, Command{"podman"}, nil, true},
		{"chained with hatch accepted", Podman, Command{"lima", "podman"}, []string{"lima"}, true},

		{"curl rejected", K8s, Command{"curl"}, nil, false},
		{"relative path rejected", K8s, Command{"./kubectl"}, nil, false},
		{"absolute path rejected", K8s, Command{"/usr/bin/kubectl"}, nil, false},
		{"bare subcommand rejected", K8s, Command{"kubectl", "delete", "ns", "prod"}, nil, false},
		{"chained without hatch rejected", Podman, Command{"lima", "podman"}, nil, false},
		{"escalation rejected even with the hatch", Podman, Command{"sudo", "podman"}, []string{"lima"}, false},
		{"double dash rejected", K8s, Command{"kubectl", "--"}, nil, false},
		{"empty rejected", K8s, Command{""}, nil, false},
		{"metachar rejected", K8s, Command{"kubectl;rm"}, nil, false},
		{"docker unlisted rejected", Docker, Command{"curl"}, nil, false},
		{"podman unlisted rejected", Podman, Command{"curl"}, nil, false},

		// The shared guard's additions over the old local one: C1 control characters
		// and invalid UTF-8 in a token, and the eight escalators it gained.
		{"C1 character rejected", K8s, Command{"kubectl", "--context", "pr\u009bod"}, nil, false},
		{"invalid UTF-8 rejected", K8s, Command{"kubectl", "--context", "pr\x9bod"}, nil, false},
		{"runuser rejected", Podman, Command{"runuser", "podman"}, nil, false},
		{"chroot rejected even with the hatch", Docker, Command{"chroot", "docker"}, []string{"lima"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// The validator's verdict.
			vc := guardConfig(tc.p)
			if err := vc.AllowCommands(tc.allowed); err != nil {
				t.Fatalf("AllowCommands(%q) = %v", tc.allowed, err)
			}
			setGuardCommand(vc, tc.p, tc.cmd)
			validatorErr := vc.Validate(tc.p)

			// The executor's verdict, from a config that never saw Validate.
			ec := guardConfig(tc.p)
			if err := ec.AllowCommands(tc.allowed); err != nil {
				t.Fatalf("AllowCommands(%q) = %v", tc.allowed, err)
			}
			setGuardCommand(ec, tc.p, tc.cmd)
			_, executorErr := guardCommandOf(ec, tc.p)

			if (validatorErr == nil) != (executorErr == nil) {
				t.Fatalf("validator and executor disagree on %q: validator=%v executor=%v",
					tc.cmd, validatorErr, executorErr)
			}
			if gotOK := executorErr == nil; gotOK != tc.wantOK {
				t.Errorf("executor check(%q) accepted = %v, want %v (err: %v)", tc.cmd, gotOK, tc.wantOK, executorErr)
			}
		})
	}
}

// setGuardCommand writes cmd into the field the platform p executes.
func setGuardCommand(c *Config, p Platform, cmd Command) {
	switch p {
	case Docker:
		c.Docker.Command = cmd
		// Keep compose independently valid, so a docker case tests the command
		// field alone rather than tripping over a compose derived from it.
		c.Docker.Compose = Command{"docker", "compose"}
	case Podman:
		c.Podman.Command = cmd
	default:
		c.K8s.Command = cmd
	}
}

// guardCommandOf reads the same field back through the executor's accessor.
func guardCommandOf(c *Config, p Platform) (Command, error) {
	if p.IsContainer() {
		return c.RuntimeCommand(p)
	}
	return c.ClusterCommand()
}

// TestExecutorRejectsWithoutValidate pins the reason the check runs twice: a
// Config built in code -- which is what every executor is handed, and what a test
// or a future caller may construct without config.Load -- must still be refused.
func TestExecutorRejectsWithoutValidate(t *testing.T) {
	c := &Config{}
	c.K8s.Command = Command{"./evil"}
	c.Docker.Command = Command{"curl"}
	c.Podman.Command = Command{"lima", "podman"}
	c.Docker.Compose = Command{"docker", "rm"}

	if _, err := c.ClusterCommand(); err == nil {
		t.Error("ClusterCommand accepted ./evil on a config that never ran Validate")
	}
	if _, err := c.RuntimeCommand(Docker); err == nil {
		t.Error("RuntimeCommand(Docker) accepted curl on a config that never ran Validate")
	}
	if _, err := c.RuntimeCommand(Podman); err == nil {
		t.Error("RuntimeCommand(Podman) accepted an unapproved lima chain")
	}
	if _, err := c.ComposeCommand(); err == nil {
		t.Error("ComposeCommand accepted a smuggled `docker rm`")
	}
}

// TestNewEscalatorsAreNeverApproved: the eight wrappers the shared guard added to
// the escalation ban are refused in a command and as an --allow-command value, with
// the message that names running solace-util itself under the wrapper. A sample of
// the old list is not repeated here -- the shared module's own tests carry it.
func TestNewEscalatorsAreNeverApproved(t *testing.T) {
	for _, name := range []string{"runuser", "sg", "chroot", "gosu", "su-exec", "setuidgid", "s6-setuidgid", "chpst"} {
		t.Run(name, func(t *testing.T) {
			c := guardConfig(Podman)
			c.Podman.Command = Command{name, "podman"}
			err := c.Validate(Podman)
			if err == nil || !strings.Contains(err.Error(), "privilege-escalation wrapper") {
				t.Errorf("Validate with command %q = %v, want the privilege-escalation refusal", name, err)
			}

			err = (&Config{}).AllowCommands([]string{name})
			if err == nil {
				t.Fatalf("AllowCommands(%q) approved a privilege-escalation wrapper", name)
			}
			for _, want := range []string{"--allow-command " + name + " is never permitted", name + " solace-util ..."} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("AllowCommands(%q) = %v\n  want it to contain %q", name, err, want)
				}
			}
		})
	}
}

// --- the compose derivation ---------------------------------------------------

// TestComposeCommandDerivation: docker.compose defaults to the runtime's own
// `compose` subcommand, and ComposeCommand owns that derivation so what Validate
// approved and what Manager.compose runs are the same expression.
func TestComposeCommandDerivation(t *testing.T) {
	c := &Config{}
	c.Docker.Command = Command{"docker"}
	got, err := c.ComposeCommand()
	if err != nil {
		t.Fatalf("ComposeCommand = %v", err)
	}
	if got.String() != "docker compose" {
		t.Errorf("derived compose = %q, want %q", got, "docker compose")
	}

	// ApplyDefaults must store the same value the accessor derives.
	d := &Config{}
	d.ApplyDefaults(Docker)
	stored := d.Docker.Compose.String()
	derived, err := d.ComposeCommand()
	if err != nil {
		t.Fatalf("ComposeCommand after defaults = %v", err)
	}
	if stored != derived.String() {
		t.Errorf("ApplyDefaults stored %q but ComposeCommand derives %q -- the two definitions have drifted",
			stored, derived)
	}
}

// TestComposeDerivationInheritsRejection: an unapproved runtime cannot become an
// approved compose command by way of the derivation.
func TestComposeDerivationInheritsRejection(t *testing.T) {
	c := &Config{}
	c.Docker.Command = Command{"curl"}
	if _, err := c.ComposeCommand(); err == nil {
		t.Error("a compose command derived from an unlisted command was accepted")
	}
}

// --- the schema cannot reach the hatch ----------------------------------------

// TestAllowCommandIsNotASchemaKey is the structural half of "the config author has
// no say": strict decoding (KnownFields) means any key attempting to name the
// escape hatch is a load error, not a silently ignored one. The field backing it is
// unexported, so yaml.v3 could not write it even without strict mode.
func TestAllowCommandIsNotASchemaKey(t *testing.T) {
	for _, key := range []string{"allowCommand", "allow-command", "allowCommands", "approvals"} {
		doc := "redundancy:\n  enabled: false\n" + key + ": [lima]\n"
		var c Config
		if err := decodeStrict(doc, &c); err == nil {
			t.Errorf("an env file key %q was accepted; the allowlist must not be settable from config", key)
		}
		if err := execguard.Check(runtimeRules(Podman), Command{"lima", "podman"}, c.approvals); err == nil {
			t.Errorf("an env file key %q reached the allowlist: `lima podman` was accepted", key)
		}
	}
}

// --- home-directory expansion in a command ------------------------------------

// TestCommandTildeExpansion covers expandCommandHomes field by field. It is named
// in expandCommandHomes' own doc comment as the reason its two field enumerations
// (this one and validateExecCommands') cannot silently drift: a field missing from
// the pass keeps its literal '~' and the row here fails.
//
// Every case runs the whole pass and then execguard.Check, because the two halves are
// one feature: expansion has to produce something the guard accepts, which is what
// makes `--kubeconfig ~/solace/kubecontext` work end to end.
func TestCommandTildeExpansion(t *testing.T) {
	const home = "/home/op"
	seam := func(c *Config) { c.homeDir = func() (string, error) { return home, nil } }

	cases := []struct {
		name     string
		platform Platform
		set      func(*Config)
		get      func(*Config) Command
		rules    execguard.Rules
	}{
		{
			"kubernetes.command", K8s,
			func(c *Config) { c.K8s.Command = Command{"oc", "--kubeconfig", "~/solace/kubecontext"} },
			func(c *Config) Command { return c.K8s.Command },
			clusterRules(),
		},
		{
			"docker.command", Docker,
			func(c *Config) { c.Docker.Command = Command{"docker", "--config", "~/dk"} },
			func(c *Config) Command { return c.Docker.Command },
			runtimeRules(Docker),
		},
		{
			// The STANDALONE binary, not `docker compose`, and the guard is what
			// decides that: the flag-shape rule admits the `compose` subword only as the
			// LAST token and only directly after an allowed binary, so there is no
			// position left for a flag value in that shape -- `docker --config ~/dk
			// compose` is refused for the subword, not for the tilde. A tilde in
			// this field is therefore only expressible through docker-compose.
			"docker.compose", Docker,
			func(c *Config) { c.Docker.Compose = Command{"docker-compose", "--project-directory", "~/dk"} },
			func(c *Config) Command { return c.Docker.Compose },
			composeRules(),
		},
		{
			"podman.command", Podman,
			func(c *Config) { c.Podman.Command = Command{"podman", "--root", "~/pm"} },
			func(c *Config) Command { return c.Podman.Command },
			runtimeRules(Podman),
		},
		{
			"replication.sites[0].via.kubernetes.command", K8s,
			func(c *Config) {
				c.Replication.Sites = []ReplSite{{
					VirtualRouterName: "v:dr",
					Via:               ReplVia{Kubernetes: &ReplViaKube{Command: Command{"kubectl", "--kubeconfig", "~/dr"}}},
				}}
			},
			func(c *Config) Command { return c.Replication.Sites[0].Via.Kubernetes.Command },
			siteRules("0"),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := guardConfig(tc.platform)
			seam(c)
			tc.set(c)
			if err := c.expandCommandHomes(tc.platform); err != nil {
				t.Fatalf("expandCommandHomes = %v, want the leading '~' expanded", err)
			}
			got := tc.get(c)
			last := got[len(got)-1]
			if !strings.HasPrefix(last, home+"/") {
				t.Errorf("%s expanded to %q, want it to start with %q -- the field is missing from "+
					"expandCommandHomes' list", tc.name, last, home+"/")
			}
			// The point of expanding at all: the guard must then accept it.
			if err := execguard.Check(tc.rules, got, execguard.Approvals{}); err != nil {
				t.Errorf("execguard.Check after expansion = %v, want accepted", err)
			}
		})
	}

	// argv[0] is deliberately NOT expanded: it names a bare binary resolved through
	// the operator's PATH, so a home directory cannot be part of it. The pass leaves
	// it alone and the shared guard's own message is what the operator sees.
	t.Run("argv[0] is left alone and then refused", func(t *testing.T) {
		c := guardConfig(K8s)
		seam(c)
		c.K8s.Command = Command{"~/bin/kubectl"}
		if err := c.expandCommandHomes(K8s); err != nil {
			t.Fatalf("expandCommandHomes = %v, want argv[0] skipped rather than an error", err)
		}
		if c.K8s.Command[0] != "~/bin/kubectl" {
			t.Errorf("argv[0] = %q, want it untouched", c.K8s.Command[0])
		}
		if err := execguard.Check(clusterRules(), c.K8s.Command, execguard.Approvals{}); err == nil {
			t.Error("an unexpanded argv[0] was accepted; the guard must refuse a path there")
		}
	})

	// An 8.3 short name survives the pass unchanged, which is the case that makes
	// the leading-vs-embedded distinction load-bearing rather than pedantic.
	t.Run("an embedded tilde survives", func(t *testing.T) {
		c := guardConfig(K8s)
		seam(c)
		const short = "C:/Users/RUNNER~1/.kube/config"
		c.K8s.Command = Command{"kubectl", "--kubeconfig", short}
		if err := c.expandCommandHomes(K8s); err != nil {
			t.Fatalf("expandCommandHomes = %v, want an 8.3 short name left alone", err)
		}
		if c.K8s.Command[2] != short {
			t.Errorf("token = %q, want %q unchanged", c.K8s.Command[2], short)
		}
	})

	// A home this tool cannot resolve fails naming that, rather than letting a
	// literal '~' reach the guard and be refused for a reason the operator cannot act
	// on.
	t.Run("an unresolvable home is reported", func(t *testing.T) {
		c := guardConfig(K8s)
		c.homeDir = func() (string, error) { return "", fmt.Errorf("no home for this user") }
		c.K8s.Command = Command{"kubectl", "--kubeconfig", "~/kc"}
		err := c.expandCommandHomes(K8s)
		if err == nil {
			t.Fatal("an unresolvable home was accepted")
		}
		for _, want := range []string{"kubernetes.command[2]", "no home for this user"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error = %v\n  want it to contain %q", err, want)
			}
		}
	})

	// A home with a space in it is refused HERE, in the words that fit a command --
	// the shared guard's own whitespace message tells the operator to write each argument
	// as its own list entry, which is wrong advice for a space they did not type.
	t.Run("a home with whitespace is refused in command words", func(t *testing.T) {
		c := guardConfig(K8s)
		c.homeDir = func() (string, error) { return `C:\Users\John Smith`, nil }
		c.K8s.Command = Command{"kubectl", "--kubeconfig", "~/kc"}
		err := c.expandCommandHomes(K8s)
		if err == nil {
			t.Fatal("a home directory containing a space was accepted")
		}
		if !strings.Contains(err.Error(), "split into arguments on whitespace") {
			t.Errorf("error = %v\n  want the command-shaped whitespace explanation", err)
		}
	})
}

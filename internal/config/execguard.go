package config

import (
	"fmt"
	"strconv"
	"unicode"

	"github.com/hafio/go-deploy-utility/execguard"
)

// An env file is executable content. kubernetes.command, docker.command, podman.command
// and docker.compose each name a binary this process runs on the operator's own
// machine, and env files travel -- repos, pull requests, shared archives -- so the
// person who wrote one is routinely not the person who runs it.
//
// The token rules, the escalation ban and the --allow-command approvals live in the
// shared execguard package, which solmq-conn-util uses too, so one definition and one
// set of messages cover both tools; its package doc carries the defence, the limits
// and the flag-value caveat. What stays here is the wiring: which fields are guarded,
// which binaries each may run, and the accessors every executor resolves argv[0]
// through, so the check the validator ran is re-run immediately before argv is built
// (k8s.Cluster.clusterCmd, container.Manager.runtime and composeCmd, and the two
// transports). A hostile env file is therefore inert even if nobody ran validation
// first. TestFlagValuePositionIsNotGuaranteed, in the shared module, pins the one
// known limit.

// execBinaries is the per-platform allowlist: the CLIs this tool actually drives.
// Nothing else may be argv[0] from config alone. `oc` is OpenShift's kubectl and
// `nerdctl` is a docker-CLI drop-in, so both speak the exact command surface this
// tool issues; `docker-compose` is here for docker.compose on a host carrying only
// the standalone v1 binary. Ordered for the error message, which names the list.
//
// Deliberately absent: wrappers such as microk8s, lima, minikube or a site shim.
// They are legitimate ways to reach a broker CLI, but approving one is the operator's
// call, not the env file's -- that is what --allow-command is for. Privilege-escalation
// wrappers are absent for a different and stronger reason: --allow-command cannot
// approve them either (see the execguard package).
var execBinaries = map[Platform][]string{
	K8s:    {"kubectl", "oc"},
	Docker: {"docker", "docker-compose", "nerdctl"},
	Podman: {"podman"},
}

// clusterRules guard kubernetes.command -- the cluster CLI, checked on every platform
// because ApplyDefaults fills it everywhere and only k8s reads it.
func clusterRules() execguard.Rules { return rulesFor("kubernetes.command", K8s, "") }

// runtimeRules guard docker.command / podman.command.
func runtimeRules(p Platform) execguard.Rules { return rulesFor(platformKey(p)+".command", p, "") }

// siteRules guard one replication site's via.kubernetes.command -- the cluster CLI this
// tool runs to reach a DR mate. The allowlist is the KUBERNETES one whatever platform
// the LOCAL broker runs on: a docker-hosted broker whose mate sits in a cluster still
// needs kubectl allowed, and the binary being run is a cluster CLI regardless of what
// this end is.
//
// site is the schema path's index. Validate passes the numeric position, since a file
// with a missing virtualRouterName still has to say WHERE to look; SiteCommand passes
// the name, which identifies the site better once the file is known good.
func siteRules(site string) execguard.Rules {
	return rulesFor("replication.sites["+site+"].via.kubernetes.command", K8s, "")
}

// composeRules guard docker.compose. It is the one field allowed a bare token, and
// only the exact word `compose` as the LAST token, directly after an allowed
// binary: the compose plugin really is a subcommand of the runtime
// (`docker compose`, or `lima nerdctl compose` when the runtime is wrapped), so
// refusing it would refuse the field's own derived default. Constraining it to that
// literal in that position is the point -- a `compose: docker rm` would otherwise
// smuggle a destructive verb ahead of the `-f <file> up -d` this tool appends.
func composeRules() execguard.Rules { return rulesFor("docker.compose", Docker, "compose") }

// rulesFor builds one field's rules. Scope is the platform name, which is what keeps
// the message reading "allowed on kubernetes are ...", and Allowed is a copy so a
// caller of the shared package can never edit the allowlist through the rules.
func rulesFor(field string, p Platform, subword string) execguard.Rules {
	return execguard.Rules{
		Field:   field,
		Scope:   string(p),
		Allowed: append([]string(nil), execBinaries[p]...),
		Subword: subword,
	}
}

// isSpace reports the whitespace a single argument may not contain, over the whole
// Unicode White_Space property rather than just ASCII space, the same rule the
// shared token check applies. The host-path gate uses it for its own whitespace
// refusal.
func isSpace(r rune) bool { return unicode.IsSpace(r) }

// AllowCommands records the operator's --allow-command approvals for this
// invocation and is the ONLY way to widen the allowlist. It is deliberately not
// reachable from the env file: the authority to run something unusual belongs to
// the person at the keyboard, who can see what they are approving, and never to
// the config author, who may be a stranger. There is no config key for it, the
// field holding it is unexported so the YAML decoder cannot reach it, and the flag
// lives on the CLI's platform commands rather than anywhere config is read.
//
// The values are checked and recorded by execguard.Approve, which refuses a path, a
// token the charset refuses, and any privilege-escalation wrapper; a bad value
// records nothing.
func (c *Config) AllowCommands(names []string) error {
	a, err := execguard.Approve("solace-util", names)
	if err != nil {
		return err
	}
	c.approvals = a
	return nil
}

// guardedCmd pairs one command VALUE with the rules it is checked against, so
// validateExecCommands can build its per-platform list in one place and check it in
// one loop. cmdField below is the expansion pass's counterpart and deliberately
// carries a POINTER and a field name instead: one gate reads a value, the other
// rewrites a field and has to name it in an error.
type guardedCmd struct {
	rules execguard.Rules
	cmd   Command
}

// cmdField names one command field and points at it, for the pass that rewrites
// tokens in place.
type cmdField struct {
	field string
	cmd   *Command
}

// validateExecCommands runs the guard over every command field this tool executes
// on the platform being validated. It is the validator half of layer 5.
//
// An UNSET field is skipped rather than refused, which is the one place the two
// enforcement points deliberately differ. In this schema an omitted key, an empty
// string and an empty list all mean "unset" (setDefaultCmd), ApplyDefaults runs
// before Validate on every path config.Load takes, and reporting "kubernetes.command is
// empty" for a file that simply never mentioned it would be a worse error than the
// mandatory-fields list it would displace. The executor has no such context -- by
// the time it is asked, an empty command means an empty argv -- so execguard.Check
// itself still refuses one, and that is what actually protects exec.
func (c *Config) validateExecCommands(p Platform) error {
	fields := []guardedCmd{
		// kubernetes.command is checked on every platform: ApplyDefaults fills it
		// everywhere, and it is printable from any code path.
		{clusterRules(), c.K8s.Command},
	}
	if p.IsContainer() {
		fields = append(fields, guardedCmd{runtimeRules(p), c.ContainerRuntime(p)})
	}
	if p == Docker {
		// ComposeCommand owns the "unset -> <runtime> compose" derivation, so what
		// is checked here is exactly what Manager.compose will run.
		fields = append(fields, guardedCmd{composeRules(), c.composeOrDerived()})
	}
	// A replication site's own cluster CLI is config text that reaches os/exec exactly
	// like kubernetes.command, so it goes through the same guard on every platform --
	// the mate may be in a cluster whatever this end runs on. Indexed by position here
	// because a file whose virtualRouterName is missing must still be told where to
	// look.
	for i, s := range c.Replication.Sites {
		if s.Via.Kubernetes == nil {
			continue
		}
		fields = append(fields, guardedCmd{siteRules(strconv.Itoa(i)), s.Via.Kubernetes.Command})
	}
	for _, f := range fields {
		if len(f.cmd) == 0 {
			continue
		}
		if err := execguard.Check(f.rules, f.cmd, c.approvals); err != nil {
			return err
		}
	}
	return nil
}

// expandCommandHomes resolves a leading `~` in every command token EXCEPT argv[0],
// for the same fields validateExecCommands guards on this platform. `oc --kubeconfig
// ~/solace/kubecontext` is the case it exists for: exec never involves a shell, so
// without this the tilde reaches the cluster CLI verbatim and names a directory
// called `~`.
//
// argv[0] is skipped because it is not a path at all -- it is a bare allowlisted
// name resolved through the operator's PATH (the shared binary check), so there is nothing there
// for a home directory to be part of, and expanding it would turn a refusable
// mistake into an absolute path that then fails the bare-name rule with a longer
// value in the message.
//
// It runs BEFORE Validate, which is the INVERSE of expandHomePaths' placement, and
// the reason is what each gate polices. CheckHostPath polices what the FILE says.
// The command guard polices what will EXECUTE: an expanded token is what reaches
// os/exec, so it is the expanded token that has to pass the charset, the flag shape
// and every other layer -- and the executors re-run that same check on a Config that
// may never have seen Load, where the shared leading-tilde rule then fails closed.
//
// The two field enumerations are separate lists rather than one shared one, because
// validateExecCommands checks a derived VALUE for docker.compose while expansion
// needs the stored field's address. They cannot silently drift apart: a field this
// misses keeps its literal `~`, and the guard that runs next refuses one in any
// token, so the failure is loud. TestCommandTildeExpansion covers each field by name.
func (c *Config) expandCommandHomes(p Platform) error {
	home := c.home()
	fields := []cmdField{
		// Every platform, matching validateExecCommands: ApplyDefaults fills
		// kubernetes.command everywhere and only k8s reads it.
		{"kubernetes.command", &c.K8s.Command},
	}
	switch p {
	case Docker:
		fields = append(fields,
			cmdField{"docker.command", &c.Docker.Command},
			// The stored field only. An UNSET compose is derived from
			// docker.command by composeOrDerived, which reads the expanded value
			// above and appends the bare word `compose`.
			cmdField{"docker.compose", &c.Docker.Compose},
		)
	case Podman:
		fields = append(fields, cmdField{"podman.command", &c.Podman.Command})
	}
	for i := range c.Replication.Sites {
		via := c.Replication.Sites[i].Via.Kubernetes
		if via == nil {
			continue
		}
		fields = append(fields, cmdField{fmt.Sprintf("replication.sites[%d].via.kubernetes.command", i), &via.Command})
	}
	for _, f := range fields {
		cmd := *f.cmd
		for i := 1; i < len(cmd); i++ {
			v, err := expandTildeToken(cmd[i], home)
			if err != nil {
				return fmt.Errorf("%s[%d] starts with '~': %w", f.field, i, err)
			}
			cmd[i] = v
		}
	}
	return nil
}

// SiteCommand returns the guarded cluster command for one replication site, named by its
// virtualRouterName. It is the sibling of ClusterCommand and exists for the same reason:
// the mate channel is built straight from a *Config and must not assume Validate ever
// ran, so the check the validator performed is re-run immediately before argv is built.
func (c *Config) SiteCommand(virtualRouterName string) (Command, error) {
	for _, s := range c.Replication.Sites {
		if s.VirtualRouterName != virtualRouterName {
			continue
		}
		if s.Via.Kubernetes == nil {
			return nil, fmt.Errorf("replication site %q is not reached over kubernetes: it declares %s",
				virtualRouterName, viaDescription(s.Via))
		}
		cmd := s.Via.Kubernetes.Command
		if err := execguard.Check(siteRules(virtualRouterName), cmd, c.approvals); err != nil {
			return nil, err
		}
		return cmd, nil
	}
	return nil, fmt.Errorf("replication.sites declares no site with virtualRouterName %q", virtualRouterName)
}

// viaDescription names what a site DOES declare, so a wrong-mechanism error says which
// one to reach for rather than only which one is missing.
func viaDescription(v ReplVia) string {
	switch {
	case v.SEMP != nil:
		return "via.semp"
	case v.Kubernetes != nil:
		return "via.kubernetes"
	default:
		return "no via block at all"
	}
}

// ClusterCommand returns the guarded kubernetes.command command. Every k8s executor
// resolves argv[0] through this rather than reading the field, so the check the
// validator ran is re-run immediately before argv is built.
func (c *Config) ClusterCommand() (Command, error) {
	cmd := c.K8s.Command
	if err := execguard.Check(clusterRules(), cmd, c.approvals); err != nil {
		return nil, err
	}
	return cmd, nil
}

// RuntimeCommand returns the guarded container runtime command for p (the checked
// counterpart of ContainerRuntime, which every executor now goes through).
func (c *Config) RuntimeCommand(p Platform) (Command, error) {
	cmd := c.ContainerRuntime(p)
	if err := execguard.Check(runtimeRules(p), cmd, c.approvals); err != nil {
		return nil, err
	}
	return cmd, nil
}

// ComposeCommand returns the guarded docker.compose command, defaulting an unset
// value to the runtime's own `compose` subcommand. ApplyDefaults fills the field
// too, so the fallback here only matters for a hand-built config -- but it is the
// one definition of the derivation, so the manager cannot compute a different
// compose command from the one Validate checked.
func (c *Config) ComposeCommand() (Command, error) {
	cmd := c.composeOrDerived()
	if err := execguard.Check(composeRules(), cmd, c.approvals); err != nil {
		return nil, err
	}
	return cmd, nil
}

// composeOrDerived is the one definition of the compose command: docker.compose if
// set, otherwise the runtime's own `compose` subcommand. ApplyDefaults stores the
// result and ComposeCommand checks it, so the value Validate approved and the value
// Manager.compose runs are the same expression rather than two copies of it.
func (c *Config) composeOrDerived() Command {
	if len(c.Docker.Compose) > 0 {
		return c.Docker.Compose
	}
	derived := make(Command, 0, len(c.Docker.Command)+1)
	derived = append(derived, c.Docker.Command...)
	return append(derived, "compose")
}

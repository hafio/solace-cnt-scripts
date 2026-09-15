package config

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// An env file is executable content. kubernetes.command, docker.command, podman.command
// and docker.compose each name a binary this process runs on the operator's own
// machine, and env files travel -- repos, pull requests, shared archives -- so the
// person who wrote one is routinely not the person who runs it. Everything in this
// file exists to keep the file from choosing what executes.
//
// What it defends: an unlisted binary; a path form pointing at a file shipped
// beside the config; a subcommand smuggled ahead of the one this tool appends.
//
// What it does not defend, by design: a trojan already installed on the operator's
// PATH (the host is compromised before this code runs -- no in-process check can
// help), and config that is malicious but entirely legitimate in form, such as a
// real kubectl aimed at the wrong cluster. Both are review problems, covered in
// docs/configuration.md's trust-model note rather than here.
//
// The whole check is one function, CheckCommand, called from BOTH Validate and
// every executor immediately before argv is built (k8s.Cluster.clusterCmd,
// container.Manager.runtimeCmd, and the two transports). A hostile env file is
// therefore inert even if nobody ran validation first, and the two enforcement
// points cannot drift because there is only one definition to drift from.
//
// The honest limit: a flag's VALUE is an unvalidatable position. Argument arity is
// unknowable without modelling every flag of every allowed binary, so the token
// after `--kubeconfig` is accepted as that flag's value whatever it says. The hard
// guarantee therefore covers argv[0] and every bare token -- what runs, and what
// could act as a subcommand -- not the contents of a flag value.
//
// That limit is wider than it first reads, and the width is ACCEPTED rather than
// closed. Because arity is unknowable, the check cannot tell a flag that takes a
// value from a ZERO-arity boolean flag -- so a bare word after a boolean reaches
// subcommand position, which is the one thing the paragraph above says is covered:
//
//	kubernetes.command: kubectl --insecure-skip-tls-verify delete
//
// passes (at i=1 the token starts with `-`; at i=2 `delete` is taken for its
// value), and the real argv becomes `kubectl --insecure-skip-tls-verify delete
// apply -f -`, where kubectl resolves its subcommand as `delete`. The container
// equivalent is `docker-compose --verbose down`.
//
// Closing it would mean either requiring `--flag=value` everywhere -- which breaks
// the ordinary `--context foo` form operators already have in working env files --
// or carrying a per-binary table of which flags take values, which is a
// compatibility burden that silently rots as those CLIs add flags. Neither is worth
// it at this trust level: reaching this needs write access to the env file, and
// anyone with that can also set kubernetes.command to any allowlisted binary and
// aim it at any cluster, which the trust-model note in docs/configuration.md
// already tells reviewers to read env files as executable content. So the rule to
// carry away is the one stated above -- argv[0] and bare tokens are guaranteed,
// flag values are not -- and this is what that costs in practice.
// TestFlagValuePositionIsNotGuaranteed pins it so it stays a known limit rather
// than becoming a surprise.

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
// approve them either (see neverAllowed).
var execBinaries = map[Platform][]string{
	K8s:    {"kubectl", "oc"},
	Docker: {"docker", "docker-compose", "nerdctl"},
	Podman: {"podman"},
}

// neverAllowed are binaries --allow-command may not approve, at any time, by anyone.
// They are all privilege-escalation wrappers, and the reason is not that escalating
// is wrong -- rootful podman genuinely needs root -- but that escalating HERE is the
// wrong place for it. `sudo solace-util deploy` elevates one process the operator
// chose, visibly, at the moment they typed it. A `command: sudo podman` elevates
// every command this tool issues for the lifetime of an env file, decided by whoever
// wrote that file, and the operator who approves it once on the command line cannot
// see what it will be used for. Escalate before invoking this tool, never through it.
//
// Refusing the category rather than the word: a list that blocked only `sudo` while
// allowing `doas` or `pkexec` would be a control in name only.
var neverAllowed = map[string]string{
	"sudo":   "sudo",
	"doas":   "doas",
	"su":     "su",
	"pkexec": "pkexec",
	"run0":   "run0",
	"runas":  "runas", // Windows
	"gsudo":  "gsudo", // Windows
}

// unsafeTokenChars are the characters no command token may carry. Under argv exec
// none of them is an injection -- exec never involves a shell, so ';' is an
// ordinary filename character -- but a token holding one is inert here only as
// long as it stays in an argv. These same tokens reach log lines, echoed output
// pasted into tickets, and the rendered compose/quadlet artifacts, so they are
// refused at the boundary instead of being escaped correctly by every consumer
// forever (S3: validate at the boundary AND sanitize at the shell layer).
//
// Backslash is included, which means a Windows path cannot appear in any token.
// That is not a loss for argv[0], which may not be a path at all, and a flag value
// takes forward slashes -- kubectl, docker and podman all accept
// `C:/Users/you/.kube/config`. The error message says so.
//
// Tilde is deliberately NOT included, and the reason is positional rather than a
// relaxation. A LEADING tilde is refused by checkToken's own rule below, because
// nothing downstream expands one: exec involves no shell, so `--kubeconfig ~/kc`
// would hand the cluster CLI a directory literally named `~`. An EMBEDDED tilde has
// to be ALLOWED, because a real Windows home directory is routinely an 8.3 short
// name (`C:\Users\RUNNER~1\...`) -- so once expandCommandHomes resolves a leading
// tilde, the resulting token legitimately carries one in the middle, and a blanket
// refusal would reject this tool's own expansion on exactly the machines that need
// it. That is the same leading-vs-everywhere-else line hostpath.go draws for host
// paths, which is why hostPathAdmits no longer has to remove this character.
const unsafeTokenChars = "\"'`$;|&<>()*?[]{}#!\\"

// pathSeparators are the characters that make a token a path rather than a bare
// name. Both are rejected on every platform: an env file that ships alongside a
// `./kubectl` must not be able to point at it, and a Windows-only check would let
// the same file do exactly that when carried to Linux.
const pathSeparators = `/\`

// commandRules describe what one Command-typed field may contain. Every field this
// tool executes has exactly one rules value, built by the helpers below, so the
// per-field differences live in one place rather than at the call sites.
type commandRules struct {
	field    string   // schema path, for error messages ("kubernetes.command")
	platform Platform // which allowlist applies
	subword  string   // the single bare subcommand this field may carry at index 1
}

// clusterRules guard kubernetes.command -- the cluster CLI, checked on every platform
// because ApplyDefaults fills it everywhere and only k8s reads it.
func clusterRules() commandRules {
	return commandRules{field: "kubernetes.command", platform: K8s}
}

// runtimeRules guard docker.command / podman.command.
func runtimeRules(p Platform) commandRules {
	return commandRules{field: platformKey(p) + ".command", platform: p}
}

// siteRules guard one replication site's via.kubernetes.command -- the cluster CLI this
// tool runs to reach a DR mate. The allowlist is the KUBERNETES one whatever platform
// the LOCAL broker runs on: a docker-hosted broker whose mate sits in a cluster still
// needs kubectl allowed, and the binary being run is a cluster CLI regardless of what
// this end is.
//
// site is the schema path's index. Validate passes the numeric position, since a file
// with a missing virtualRouterName still has to say WHERE to look; SiteCommand passes
// the name, which identifies the site better once the file is known good.
func siteRules(site string) commandRules {
	return commandRules{field: "replication.sites[" + site + "].via.kubernetes.command", platform: K8s}
}

// composeRules guard docker.compose. It is the one field allowed a bare token, and
// only the exact word `compose` as the LAST token, directly after an allowed
// binary: the compose plugin really is a subcommand of the runtime
// (`docker compose`, or `lima nerdctl compose` when the runtime is wrapped), so
// refusing it would refuse the field's own derived default. Constraining it to that
// literal in that position is the point -- a `compose: docker rm` would otherwise
// smuggle a destructive verb ahead of the `-f <file> up -d` this tool appends.
func composeRules() commandRules {
	return commandRules{field: "docker.compose", platform: Docker, subword: "compose"}
}

// allowed builds the lookup for one rules value: the platform allowlist plus the
// operator's --allow-command additions for this invocation. Rebuilt per call
// rather than cached -- the lists hold a handful of entries, and a cache would be
// one more thing that could disagree between validator and executor.
func (r commandRules) allowed(extra map[string]bool) map[string]bool {
	set := make(map[string]bool, len(execBinaries[r.platform])+len(extra))
	for _, name := range execBinaries[r.platform] {
		set[name] = true
	}
	for name := range extra {
		set[name] = true
	}
	// Belt and braces on the escalation rule. AllowCommands already refuses these
	// with a message that explains the alternative, which is where an operator
	// actually meets the rule -- this second pass makes the outcome structural, so a
	// future edit that adds one to execBinaries, or a caller that populates
	// extraAllowed some other way, still cannot put a privilege-escalation wrapper
	// in front of the broker CLI.
	//
	// It sweeps the SET rather than iterating neverAllowed and deleting by name:
	// the two are only the same while every key is already lowercase, so deleting
	// by name left `Sudo` (or `SUDO.exe`) sitting in the set untouched -- the
	// identical blind spot the other enforcement point had. Asking escalator about
	// each key instead means both belts fold case the same way, by construction.
	for name := range set {
		if escalator(name) != "" {
			delete(set, name)
		}
	}
	return set
}

// list renders this platform's allowlist for an error message, in declaration
// order, without the operator's additions -- naming those back would suggest the
// env file could have asked for them.
func (r commandRules) list() string {
	return strings.Join(execBinaries[r.platform], ", ")
}

// CheckCommand is the single definition of what an executable command may be. It
// runs layers 1-3 in order: every token passes the charset; argv[0] is a bare,
// allowlisted binary name; every later token is a flag, a flag's value, or another
// allowlisted binary. extra carries the operator's --allow-command approvals for
// this invocation, and is nil on the config-only path.
//
// It is exported so the executors can re-run the exact check the validator ran,
// immediately before they build argv (layer 5). Callers use the Command accessors
// below rather than calling this directly.
func CheckCommand(r commandRules, cmd Command, extra map[string]bool) error {
	if len(cmd) == 0 {
		return fmt.Errorf("%s is empty: it must name the binary to run (one of: %s)", r.field, r.list())
	}
	for i, tok := range cmd {
		if err := checkToken(r.field, i, tok); err != nil {
			return err
		}
	}
	ok := r.allowed(extra)
	if err := checkBinary(r, cmd[0], ok); err != nil {
		return err
	}
	return checkFlagShape(r, cmd, ok)
}

// checkToken is layer 1: the token charset. Emptiness and control characters keep
// their own messages, since an empty argument and a converted-from-bash newline
// are the two mistakes a well-meaning env file actually makes.
func checkToken(field string, i int, tok string) error {
	if tok == "" {
		return fmt.Errorf("%s[%d] is an empty argument; remove it or quote the intended value", field, i)
	}
	if j := strings.IndexFunc(tok, isCtrl); j >= 0 {
		return fmt.Errorf("%s[%d] contains a control character (0x%02x) at offset %d: %q",
			field, i, tok[j], j, tok)
	}
	if strings.IndexFunc(tok, isSpace) >= 0 {
		return fmt.Errorf("%s[%d] = %q contains whitespace inside one argument; write each argument as its own "+
			"list entry, or use the scalar form, which splits on whitespace", field, i, tok)
	}
	if j := strings.IndexFunc(tok, isInvisible); j >= 0 {
		r, _ := utf8.DecodeRuneInString(tok[j:])
		return fmt.Errorf("%s[%d] = %q contains an invisible formatting character (U+%04X) at offset %d; "+
			"it cannot be seen in a review or a log line, so it is not allowed in a command token", field, i, tok, r, j)
	}
	if j := strings.IndexAny(tok, unsafeTokenChars); j >= 0 {
		return fmt.Errorf("%s[%d] = %q contains %q, which is not allowed in a command token; "+
			"a Windows path works with forward slashes (C:/Users/you/.kube/config)",
			field, i, tok, string(tok[j]))
	}
	// A LEADING tilde, refused at every position but for two different reasons --
	// which is why the message differs. Only a leading one is refused: an embedded
	// tilde is an 8.3 short name (`C:/Users/RUNNER~1/...`) and is exactly what
	// expandCommandHomes itself produces on such a machine (unsafeTokenChars' comment
	// has the argument), so refusing those would refuse this tool's own expansion.
	//
	// Reaching this branch at index 0 is an operator's mistake and says so. Reaching
	// it anywhere else means the command never went through config.Load, because
	// expandCommandHomes resolves those tokens BEFORE Validate runs -- so the layer-5
	// re-check the executors make on a hand-built Config still fails closed rather
	// than handing exec a directory named `~`.
	if strings.HasPrefix(tok, "~") {
		if i == 0 {
			return fmt.Errorf("%s[0] = %q starts with '~': argv[0] names the binary to run and is resolved "+
				"through the operator's own PATH, so it must be a bare name (one of the allowed binaries) "+
				"rather than a path of any kind -- a home directory cannot be part of it", field, tok)
		}
		return fmt.Errorf("%s[%d] = %q still starts with '~', which nothing downstream expands: exec never "+
			"involves a shell, so this would name a directory literally called '~'. A command read from an "+
			"env file has its leading '~' expanded before this check runs, so a '~' surviving here means the "+
			"command was built without going through config.Load", field, i, tok)
	}
	return nil
}

// checkBinary is layer 2: argv[0] is a bare name from the allowlist. The bare-name
// rule is what stops `command: ./kubectl` -- a relative or absolute path would run
// a file the env file chose, which is the whole attack; a bare name can only be
// resolved through the operator's own PATH (engine.Exec does that resolution
// explicitly, and refuses to resolve from the current directory).
func checkBinary(r commandRules, tok string, allowed map[string]bool) error {
	if j := strings.IndexAny(tok, pathSeparators); j >= 0 {
		return fmt.Errorf("%s[0] = %q must be a bare binary name, not a path: %q would run a file named by the "+
			"env file rather than the one on your PATH -- write it as %q and let PATH resolve it",
			r.field, tok, string(tok[j]), execBase(tok))
	}
	if !allowed[execBase(tok)] {
		return fmt.Errorf("%s[0] = %q is not a binary this tool runs: allowed on %s are %s -- "+
			"correct the env file, or approve this one for a single run with --allow-command %s",
			r.field, tok, r.platform, r.list(), execBase(tok))
	}
	return nil
}

// checkFlagShape is layer 3: everything after argv[0]. A token passes if it is a
// flag, another allowlisted binary (a chained runner such as an approved
// `lima nerdctl`), this field's one permitted subword, or the value of the flag
// before it. Anything else is a bare word sitting exactly where this tool appends
// its own subcommand, so it is refused.
func checkFlagShape(r commandRules, cmd Command, allowed map[string]bool) error {
	for i := 1; i < len(cmd); i++ {
		tok := cmd[i]
		switch {
		case tok == "--":
			// End-of-flags would reopen everything the bare-word rule closes: the
			// next token stops being a subcommand to the shell-free parser and
			// starts being a positional the allowed binary happily accepts.
			return fmt.Errorf("%s[%d] = \"--\" is not allowed: end-of-flags would let the env file smuggle a "+
				"positional argument ahead of the subcommand this tool appends -- remove it", r.field, i)
		case strings.HasPrefix(tok, "-"):
		case allowed[execBase(tok)]:
		case r.subword != "" && tok == r.subword && i == len(cmd)-1 && allowed[execBase(cmd[i-1])]:
			// The field's one permitted subcommand, and only in the one position
			// where it means what the field says: last, directly after an allowed
			// binary. That covers `docker compose` and, with lima approved,
			// `lima nerdctl compose` -- while `docker rm` is still a bare word
			// and `docker compose up` still has a token this tool did not append.
		case strings.HasPrefix(cmd[i-1], "-") && !strings.Contains(cmd[i-1], "="):
			// The value of the preceding flag. Unvalidatable beyond the charset:
			// arity is unknowable without modelling every flag of every allowed
			// binary, so this position is deliberately trusted (see the file header).
		default:
			return fmt.Errorf("%s[%d] = %q is not allowed in subcommand position: this tool appends its own "+
				"subcommand, so a bare word here would run ahead of it -- only flags, their values, and "+
				"allowed binaries (%s) may follow; approve a chained runner for a single run with "+
				"--allow-command %s", r.field, i, tok, r.list(), execBase(tok))
		}
	}
	return nil
}

// execBase strips one optional .exe/.EXE suffix so a Windows operator may write
// `kubectl.exe` and match the same allowlist entry. Only the suffix is folded --
// the name itself is compared exactly, because on a case-sensitive filesystem
// `KUBECTL` and `kubectl` are different files.
func execBase(tok string) string {
	if len(tok) > 4 && strings.EqualFold(tok[len(tok)-4:], ".exe") {
		return tok[:len(tok)-4]
	}
	return tok
}

// escalator reports the privilege-escalation wrapper tok names, or "" for anything
// else. It is the ONE case-insensitive comparison in this file, and the asymmetry
// is deliberate:
//
//   - The ALLOWLIST is matched exactly (execBase), because on a case-sensitive
//     filesystem `KUBECTL` and `kubectl` are different files and folding case there
//     would let an env file naming `KUBECTL` be approved by `kubectl`'s entry and
//     then execute something else entirely. Loosening a positive match adds
//     binaries nobody approved.
//
//   - This DENY list must fold, because the filesystems that decide whether the
//     name resolves do. macOS is case-insensitive by default and Windows always
//     is, so `Sudo` there resolves to exactly the binary `sudo` names -- and
//     Windows has shipped sudo.exe since 2024. Matching it exactly meant
//     `--allow-command Sudo` was accepted, stored, and passed both enforcement
//     points, so a floor that "can be approved by nobody" was bypassed by a
//     capital letter.
//
// Tightening a deny list can only refuse more, which is the safe direction to be
// wrong in: the worst case is refusing a legitimately-named binary that happens to
// case-fold onto a wrapper, and no such binary exists on any platform this tool
// supports.
func escalator(tok string) string { return neverAllowed[strings.ToLower(execBase(tok))] }

// isSpace reports the whitespace a single argument may not contain, over the whole
// Unicode White_Space property rather than just ASCII space. That matters because
// the two YAML forms do not agree: the scalar form is split with strings.Fields,
// which is Unicode-aware, so `runtime: kubectl -n a<U+3000>b` becomes three tokens
// and never reaches here with the character embedded -- but the explicit sequence
// form preserves it. An ASCII-only check would therefore accept through one form
// exactly what it rejects through the other. Tab and the vertical whitespace are
// caught earlier as control characters, which gives them a better message.
func isSpace(r rune) bool { return unicode.IsSpace(r) }

// isInvisible reports the Unicode format characters (category Cf): zero-width
// spaces and joiners, bidirectional overrides, and their relatives. They are not
// whitespace and carry no argv-splitting risk under argv exec, so they get their
// own check and their own message -- what makes them unacceptable is that they are
// invisible. A token that renders identically to a legitimate one, in a review, a
// log line, or a an echoed transcript pasted into a ticket, defeats the reading
// this whole file asks an operator to do. Nothing legitimate needs one.
func isInvisible(r rune) bool { return unicode.Is(unicode.Cf, r) }

// AllowCommands records the operator's --allow-command approvals for this
// invocation and is the ONLY way to widen the allowlist. It is deliberately not
// reachable from the env file: the authority to run something unusual belongs to
// the person at the keyboard, who can see what they are approving, and never to
// the config author, who may be a stranger. There is no config key for it, the
// field holding it is unexported so the YAML decoder cannot reach it, and the flag
// lives on the CLI's platform commands rather than anywhere config is read.
//
// Each value gets the same charset check as a command token and must be a bare
// name, so an approval cannot itself become the path form layer 2 refuses.
func (c *Config) AllowCommands(names []string) error {
	for _, name := range names {
		if err := checkToken("--allow-command", 0, name); err != nil {
			// Re-phrase for a flag rather than a schema field: the index means
			// nothing to someone who typed a value on the command line.
			return fmt.Errorf("invalid --allow-command value: %s",
				strings.TrimPrefix(strings.TrimPrefix(err.Error(), "--allow-command[0] "), "= "))
		}
		if j := strings.IndexAny(name, pathSeparators); j >= 0 {
			return fmt.Errorf("invalid --allow-command value %q: it must be a bare binary name, not a path "+
				"(found %q); approve it as %q and let PATH resolve it", name, string(name[j]), execBase(name))
		}
		if esc := escalator(name); esc != "" {
			return fmt.Errorf("--allow-command %s is never permitted: %s would elevate every command this tool "+
				"issues, for the whole life of an env file -- elevate the tool instead, at the moment you run it "+
				"(%s solace ...), so the privilege belongs to one invocation you chose", name, esc, esc)
		}
		if c.extraAllowed == nil {
			c.extraAllowed = make(map[string]bool, len(names))
		}
		c.extraAllowed[execBase(name)] = true
	}
	return nil
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
// the time it is asked, an empty command means an empty argv -- so CheckCommand
// itself still refuses one, and that is what actually protects exec.
func (c *Config) validateExecCommands(p Platform) error {
	fields := []struct {
		rules commandRules
		cmd   Command
	}{
		// kubernetes.command is checked on every platform: ApplyDefaults fills it
		// everywhere, and it is printable from any code path.
		{clusterRules(), c.K8s.Command},
	}
	if p.IsContainer() {
		fields = append(fields, struct {
			rules commandRules
			cmd   Command
		}{runtimeRules(p), c.ContainerRuntime(p)})
	}
	if p == Docker {
		// ComposeCommand owns the "unset -> <runtime> compose" derivation, so what
		// is checked here is exactly what Manager.compose will run.
		fields = append(fields, struct {
			rules commandRules
			cmd   Command
		}{composeRules(), c.composeOrDerived()})
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
		fields = append(fields, struct {
			rules commandRules
			cmd   Command
		}{siteRules(strconv.Itoa(i)), s.Via.Kubernetes.Command})
	}
	for _, f := range fields {
		if len(f.cmd) == 0 {
			continue
		}
		if err := CheckCommand(f.rules, f.cmd, c.extraAllowed); err != nil {
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
// name resolved through the operator's PATH (checkBinary), so there is nothing there
// for a home directory to be part of, and expanding it would turn a refusable
// mistake into an absolute path that then fails the bare-name rule with a longer
// value in the message.
//
// It runs BEFORE Validate, which is the INVERSE of expandHomePaths' placement, and
// the reason is what each gate polices. CheckHostPath polices what the FILE says.
// The command guard polices what will EXECUTE: an expanded token is what reaches
// os/exec, so it is the expanded token that has to pass the charset, the flag shape
// and every other layer -- and the executors re-run that same check on a Config that
// may never have seen Load, where checkToken's leading-tilde rule then fails closed.
//
// The two field enumerations are separate lists rather than one shared one, because
// validateExecCommands checks a derived VALUE for docker.compose while expansion
// needs the stored field's address. They cannot silently drift apart: a field this
// misses keeps its literal `~`, and the guard that runs next refuses one in any
// token, so the failure is loud. TestCommandTildeExpansion covers each field by name.
func (c *Config) expandCommandHomes(p Platform) error {
	home := c.home()
	fields := []struct {
		field string
		cmd   *Command
	}{
		// Every platform, matching validateExecCommands: ApplyDefaults fills
		// kubernetes.command everywhere and only k8s reads it.
		{"kubernetes.command", &c.K8s.Command},
	}
	switch p {
	case Docker:
		fields = append(fields,
			struct {
				field string
				cmd   *Command
			}{"docker.command", &c.Docker.Command},
			// The stored field only. An UNSET compose is derived from
			// docker.command by composeOrDerived, which reads the expanded value
			// above and appends the bare word `compose`.
			struct {
				field string
				cmd   *Command
			}{"docker.compose", &c.Docker.Compose},
		)
	case Podman:
		fields = append(fields, struct {
			field string
			cmd   *Command
		}{"podman.command", &c.Podman.Command})
	}
	for i := range c.Replication.Sites {
		via := c.Replication.Sites[i].Via.Kubernetes
		if via == nil {
			continue
		}
		fields = append(fields, struct {
			field string
			cmd   *Command
		}{fmt.Sprintf("replication.sites[%d].via.kubernetes.command", i), &via.Command})
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
		if err := CheckCommand(siteRules(virtualRouterName), cmd, c.extraAllowed); err != nil {
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
	if err := CheckCommand(clusterRules(), cmd, c.extraAllowed); err != nil {
		return nil, err
	}
	return cmd, nil
}

// RuntimeCommand returns the guarded container runtime command for p (the checked
// counterpart of ContainerRuntime, which every executor now goes through).
func (c *Config) RuntimeCommand(p Platform) (Command, error) {
	cmd := c.ContainerRuntime(p)
	if err := CheckCommand(runtimeRules(p), cmd, c.extraAllowed); err != nil {
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
	if err := CheckCommand(composeRules(), cmd, c.extraAllowed); err != nil {
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

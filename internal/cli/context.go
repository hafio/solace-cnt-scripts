package cli

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"solace/internal/config"
	"solace/internal/engine"
	"solace/internal/output"
)

// App is the shared context threaded through every command: the parsed global
// flags, the resolved config, and the Runner used for all external commands.
// It replaces the bash "source 000-env.sh" bootstrap -- one load, reused by the
// whole command tree. Explicit context, never a global.
type App struct {
	EnvName string // -e/--env value: an env file name, or a path
	BaseDir string // dir searched for the env file, and holding env/ (defaults to CWD)
	Verbose bool   // -v/--verbose: announce every external command as it runs

	// AllowCommand is the repeatable --allow-command escape hatch: binaries the
	// OPERATOR approved for the config's platform command, for this invocation
	// only. It reaches config through Load's argument list, never through the
	// schema, so an env file has no way to extend its own allowlist.
	AllowCommand []string

	// PlatformFlag is the raw --platform value, still in whatever spelling was
	// typed (canonical or the kube/dk/pm abbreviations). Platform below is the
	// resolved one, settled by resolvePlatform from this flag or, when it is
	// empty, from the platform sections the env file declares.
	PlatformFlag string

	Platform config.Platform
	Cfg      *config.Config
	Runner   engine.EnvRunner

	// Prompt seams, in the spirit of Manager.Resolve/Geteuid: the confirm
	// helpers gate destructive actions on an interactive terminal, and a test cannot
	// supply one. Interactive nil means stdinCanAnswer and PromptIn nil means
	// os.Stdin. Nothing user-facing sets either field; only tests do.
	Interactive func() bool
	PromptIn    io.Reader

	// The run's single buffered prompt reader, built on first use by promptSource.
	// It must be shared across questions: a per-call reader discards whatever it
	// buffered past the line it returned, which loses the answer to the NEXT
	// question in the same run.
	promptReader *bufio.Reader

	// NewRunner builds the Runner every command executes through. It exists as a
	// seam for the same reason Interactive/PromptIn do: a test cannot supply a
	// cluster or a container engine, and the property worth asserting is which
	// argv a command would issue. Nil means the production runner, so nothing but
	// a test ever substitutes one -- there is no user-facing way to make this tool
	// print commands instead of running them, because `generate` is how you look
	// at an artifact before applying it.
	NewRunner func(a *App) engine.EnvRunner

	// Command-local flag scratch space. Only one command runs per invocation,
	// so sharing these on the app context is safe.
	deleteLayer bool     // remove --delete-data / --delete-crd (take the retained layer too)
	noPrompt    bool     // --no-prompt: ask nothing, and take the safe answer to each question
	all         bool     // status broker --all (every broker in the cluster)
	detail      bool     // status --detail (static artifacts, not just running ones)
	pod         string   // --pod role selector for cli/shell/copy/cli-script
	destDir     string   // copy into --dir
	days        int      // diagnostics --days
	restart     bool     // deploy broker --restart (bounce a running broker)
	out         string   // --out/-o: write the rendered artifact here instead of stdout
	follow      bool     // logs --follow/-f: keep streaming instead of printing a snapshot
	tail        string   // logs --tail: how many trailing lines (a count, or "all")
	since       string   // logs --since: how far back to read, canonicalised as a duration
	timestamps  bool     // logs --timestamps: prefix each line with its time
	previous    bool     // logs broker --previous: the PREVIOUS container's logs (kubernetes only)
	vpns        []string // export-config --vpn: capture only these message-VPNs (repeatable)
	brokerOnly  bool     // export-config --broker-only: broker-level configuration only

	// The `broker configure` direction flags. Each leaf has a DEFAULT direction and a
	// flag for the other one, so `configure domain-certs` applies and
	// `configure domain-certs --remove` removes -- no sub-verb, and no flag whose
	// absence has to be memorised. The default flag is registered too and is
	// redundant by design: a script may prefer to say which way it is going, and
	// refusing the obvious spelling would be a papercut for nothing.
	//
	// Both flags of a pair set at once is a usage error, checked by the pair's own
	// helper rather than left to whichever the handler happened to read first.
	flagApply   bool // configure ... --apply (the default for certs and keys)
	flagRemove  bool // configure ... --remove
	flagDisable bool // configure ... --disable (the default for the hardening pair)
	flagEnable  bool // configure ... --enable

	// Hostname resolves this host's name for the container node-role detection
	// containerRole does (broker.Ops.LocalRole). Unset -- the production case --
	// it stays broker.New's os.Hostname; a test injects a fixed name, which is
	// the only way to exercise a deploy that decides its own role. Same shape as
	// NewRunner/Interactive/PromptIn.
	Hostname func() (string, error)

	// LocalAddrs is the second half of that detection: this machine's own interface
	// addresses, matched against redundancy.*.addr when no configured name matches
	// the hostname. Same seam, same reason -- a test can no more renumber the machine
	// running the suite than rename it, and detection that read the real interfaces
	// would pass or fail on what the CI host happens to be addressed as.
	LocalAddrs func() (map[string]bool, error)

	// kubeContext is the kubeconfig context announceKubeContext resolved at load,
	// repeated by every destructive Kubernetes prompt (k8sWhat). Empty when it
	// could not be resolved, when the platform is not Kubernetes, or under a
	// test-supplied runner -- the prompts drop the clause rather than guess.
	kubeContext string
}

// load resolves the env file for the app's platform and builds the config +
// runner. Called last in every runnable command's PreRunE (prepare, platform.go)
// so config errors surface before any work, and help still works without a valid
// env (cobra skips PreRun for --help). By then a.Platform is settled, which is
// what config.Load needs to scope its defaults and validation. cmd is the command
// about to run: announceCommands needs it to stay quiet where nothing executes.
func (a *App) load(cmd *cobra.Command) error {
	path, err := config.ResolveEnvPath(a.BaseDir, a.EnvName)
	if err != nil {
		return err
	}
	// Echo the winner: a file in the base dir shadows the env/ copy of the same
	// name, and that has to be visible rather than silent.
	step("env file: %s", path)
	// The operator's --allow-command approvals go in here, before Validate runs
	// inside Load, so the execution guard sees the same allowlist the executors
	// will. Passing them as an argument is what keeps them out of the schema.
	cfg, err := config.Load(path, a.Platform, a.AllowCommand...)
	if err != nil {
		return err
	}
	a.Cfg = cfg
	a.fillStandaloneNodeName()
	if a.NewRunner != nil {
		// A test-supplied runner: it does not execute, so there are no binaries
		// worth resolving and announcing.
		a.Runner = a.NewRunner(a)
		return nil
	}
	a.Runner = engine.NewExec(os.Stderr, a.Verbose)
	if a.willExecute(cmd) {
		a.announceCommands()
		a.announceKubeContext()
	}
	return nil
}

// fillStandaloneNodeName names a STANDALONE container broker after the host it runs
// on when the env file left redundancy.primary.name empty. The name is the broker's
// routername and the container's hostname, and with one node it is always THIS host --
// which is what the broker would have called itself anyway.
//
// Here, at load, rather than where the role is decided: the routername is read by the
// check report, the DNS check, the rendered artifact and `validate` alike, and a fill
// that happened on only the deploying paths would have those four disagree about the
// name of the same broker.
//
// Container-only and standalone-only. Kubernetes never reads redundancy.*.name (the
// operator names the pods off kubernetes.name), and HA keeps all three names mandatory
// at load, because each one keys a group-table entry every host renders and no host
// knows another machine's name -- see config.FillStandaloneNodeName.
func (a *App) fillStandaloneNodeName() {
	if !a.Platform.IsContainer() {
		return
	}
	name, err := a.hostname()
	if err != nil {
		// Not fatal on its own: validate does not require the name in standalone, and
		// the render simply carries an empty routername. Say so rather than failing a
		// command that may not need it.
		warn("cannot read this host's name for the broker's routername: %v", err)
		return
	}
	if a.Cfg.FillStandaloneNodeName(name) {
		step("routername not configured; using this host's name: %s", name)
	}
}

// hostname reads this host's name through the App's seam, defaulting to os.Hostname.
// Same shape as broker.Ops.hostname, and the same reason: a test cannot rename the
// machine running the suite.
func (a *App) hostname() (string, error) {
	fn := a.Hostname
	if fn == nil {
		fn = os.Hostname
	}
	h, err := fn()
	return strings.TrimSpace(h), err
}

// announceKubeContext resolves and reports the kubeconfig context every kubectl
// call in this run will land in, beside the `using kubectl:` line.
//
// It is the one fact the preamble was missing that an operator cannot recover
// from anywhere else. The env file names a namespace, never a cluster: which
// cluster is decided entirely by the kubeconfig's current context, so an env
// file that says "dev" and a context that drifted to prod look identical right
// up until the work lands. Naming it here puts the answer in front of the
// operator BEFORE any prompt, which is also why a.kubeContext is kept for the
// destructive prompts to repeat (k8sWhat) rather than looked up again there.
//
// `config current-context` reads the kubeconfig file and contacts nothing, so
// this costs no round trip and works against an unreachable cluster.
//
// Like announceCommands it never fails: a kubeconfig with no current context is
// a real state (a fresh install, or --kubeconfig pointing somewhere bare) and
// the first real call reports it far better than a preamble could. kubectl's own
// stderr still reaches the operator when that happens -- engine.Exec wires the
// child's stderr straight through -- which is left alone deliberately: this tool
// has no --context of its own, so a kubeconfig with no current context is about
// to fail every call, and saying so once here is a head start, not noise.
func (a *App) announceKubeContext() {
	if a.Platform != config.K8s {
		return
	}
	kc, err := a.Cfg.ClusterCommand()
	if err != nil {
		return
	}
	out, err := a.Runner.Output(bg(), kc.Name(), kc.Args("config", "current-context")...)
	if ctx := strings.TrimSpace(string(out)); err == nil && ctx != "" {
		a.kubeContext = ctx
		step("kube-context: %q", ctx) // engine output, so quoted before it reaches a terminal
	}
}

// announceCommands names the binaries this env file chose, resolved, before any of them
// runs -- so the information sits with `==> env file:` in the preamble instead of
// repeating itself between report lines on every call.
//
// The set is exactly the four fields the execution guard exists for (kubernetes.command,
// docker.command, podman.command, docker.compose), read through their guarded accessors:
// those are the binaries CONFIG TEXT chose, which is the whole reason their location is
// worth showing. The tool's own fixed helpers (mkdir, chown, rm, sh, systemctl) were
// chosen here, not by an env file, so they are announced only under --verbose.
//
// It never fails. A name that resolves nowhere is skipped silently: this is a report,
// and the first real execution already fails with engine.Resolve's own actionable
// message (not found on PATH, or the current-directory refusal).
func (a *App) announceCommands() {
	var cmds []config.Command
	switch a.Platform {
	case config.K8s:
		if c, err := a.Cfg.ClusterCommand(); err == nil {
			cmds = append(cmds, c)
		}
	case config.Docker, config.Podman:
		if c, err := a.Cfg.RuntimeCommand(a.Platform); err == nil {
			cmds = append(cmds, c)
		}
		if a.Platform == config.Docker {
			if c, err := a.Cfg.ComposeCommand(); err == nil {
				cmds = append(cmds, c)
			}
		}
	}
	// Keyed by resolved path, not by name: `docker` and a derived `docker compose` are
	// one binary and deserve one line, while a standalone docker-compose gets its own.
	seen := make(map[string]bool, len(cmds))
	for _, c := range cmds {
		path, err := engine.Resolve(c.Name())
		if err != nil || seen[path] {
			continue
		}
		seen[path] = true
		step("using %s: %s", c.Name(), path)
	}
}

// progress is the stderr sink every line this package narrates goes through, so
// the `==> ` and `[TAG ] ` shapes are defined once, in internal/output, rather
// than at each call site. It is a function rather than a package variable
// because a Sink holds nothing worth caching and nothing here is global, and the
// helpers below need one from call sites that have no *App in scope.
func progress() *output.Sink { return output.New(os.Stderr) }

// lineSink is what the k8s/container/broker entry types are handed as their Log:
// a RAW line emitter. They build their own output.Sink over it and add the
// prefixes there, which is why this deliberately is not `step` -- passing a
// prefixing function would prefix twice, once here and once in the callee.
func lineSink() func(string, ...any) { return progress().Line }

// warn prints a non-fatal warning to stderr in the house [WARN] style.
func warn(format string, args ...any) { progress().Warn(format, args...) }

// step prints a progress line to stderr so it never pollutes rendered stdout.
func step(format string, args ...any) { progress().Step(format, args...) }

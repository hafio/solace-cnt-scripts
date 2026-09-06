# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

A single Go binary, `solace`, that deploys and operates Solace PubSub+ Event Brokers on Kubernetes (via the Solace EventBroker Operator / PubSubPlusEventBroker CRD), Docker, and Podman. You describe the broker once in a YAML env file and drive the whole lifecycle from one standardized command tree. **Unsupported** -- not a Solace product. [README.md](README.md) is the quick start; [docs/configuration.md](docs/configuration.md) is the full env-file reference and [docs/operations.md](docs/operations.md) the day-2 lifecycle.

## Go implementation (`solace-util` binary)

The `solace-util` binary presents one standardized lifecycle command tree across Kubernetes, Docker, and Podman. Build and the dev-script tasks are in [docs/developer.md](docs/developer.md), the lifecycle in [README.md](README.md) and [docs/operations.md](docs/operations.md); package layout is `internal/{abbrev,config,engine,render,broker,k8s,container,convert,cli}` + `main.go`,
plus `internal/tools/vulnjudge` -- a dev-only command the `scan` task pipes govulncheck's
JSON through, so a fixable vulnerability fails the gate and one with no released fix warns --
and `internal/tools/itest`, a dev-only, never-shipped harness (not in `BUILD_TARGETS`, built
by the `itest` task) that drives the real production code paths against a LIVE environment to
settle the three `ASSUMED, NOT VERIFIED` / `NEEDS VERIFICATION` comments in the code. It
produces evidence rather than gating: each probe reports VERIFIED / REFUTED / INCONCLUSIVE so
the comment can become a fact, and names the follow-up fix when the assumption was wrong. It
is **operator-run only** -- the probes mutate live broker state, and the dev scripts
deliberately only build it. See [docs/itest.md](docs/itest.md).

**Platform naming: one word, two places.** A platform's `config.Platform` constant value
is simultaneously the env file's section key (`yaml:"kubernetes"`) and the `--platform`
value. The flag reads the constant through `config.ParsePlatform`; the yaml tag is a
separate literal, so `TestPlatformConstantsMatchSchemaSections` is what stops the two
drifting apart. The word is the product's own -- `kubernetes`, `docker`, `podman`; the Go
identifiers stay short (`internal/k8s`, `K8sConfig`, `opK8s*`) and are not user-facing.
`--platform` also takes the abbreviations `kube`/`dk`/`pm`, which are shorthand for a current
name rather than a second name: they are refused in the env file, never printed back, and
deliberately not offered by completion (a completion is already the thing that saves you
the typing). Neither `k8s` nor `k8` is accepted as a `--platform` value.

**Verb then noun, and nothing implicit (`internal/cli/commands.go`).** Every verb that acts
on more than one kind of thing names the thing -- `deploy broker`, `remove operator`,
`status broker` -- and a verb that owns objects has NO `RunE`, so running it bare prints
what it can act on instead of acting. `group()` is the helper that builds those; it
deliberately does not go through `leaf()`/`wireExec()`, because a group runs nothing and so
needs neither the pre-run nor `--allow-command`. `TestDangerousVerbsHaveNoBareAlias` is what
stops a `RunE` being added to one later: it is also what makes `rm` safe to offer as an
abbreviation, since `rm` alone removes nothing. `group()` takes its `Long` as a parameter;
the leaf builders (`leaf`/`roleOnK8sLeaf`/`roleOnContainerLeaf`/`dispatchLeaf`) take only a
`Short`, so `withLong(c, text)` (helpers.go) attaches one, wrapping inline the way `onlyOn`
and `renderOnly` do rather than forcing a named variable per command. A command built from a
`cobra.Command` literal sets the `Long:` field directly. Those are the only two shapes -- a
`c.Long = ...` assignment is drift. The text is for what the `Short` had no room for (what
actually runs, what must already be true, what is left behind, the constraint); the generator
skips a `Long` equal to the `Short`, and the flag table below it already carries every flag's
usage string, so restating either is noise.

**Removals keep the expensive layer (`helpers.go`: `layer`, `addLayerFlags`, `confirmLayer`).**
`remove broker` keeps persistent data (PVCs / the container data dir); `remove operator`
keeps the CRDs, whose deletion would cascade to every PubSubPlusEventBroker in the cluster.
Both ask the same way and report the outcome either way: `--delete-data`/`--delete-crd`
delete without asking, `--no-prompt` keeps without asking, an interactive run needs an exact
`yes`, and a non-interactive one keeps. `--no-prompt` is the ONE silencer -- it confirms the
removal and takes the safe answer to the layer question, so `--delete-data --no-prompt` is
what a fully unattended removal-with-data looks like; the two flags answer different
questions and deliberately compose rather than conflict. There is no `--yes`: a single
silencer per command beats two flags whose overlap has to be memorised. `remove all` also
leaves the operator installed, since it is cluster-scoped. Every destructive command
confirms, `remove secrets`/`remove namespace` and `restart broker` included.

**There is no `--dry-run` and no `--gen-*` flags.** `generate <target>` is the only
look-before-you-leap surface: rendering is a command with a named target rather than a flag
on the commands that deploy, so an artifact you meant to read and a cluster you meant to
change are not one typo apart. `engine.Echo` is a TEST-ONLY runner, installed through the
`App.NewRunner` seam (context.go) -- the same shape as the `Interactive`/`PromptIn` prompt
seams, and the reason ~200 wiring tests can still assert exact argv without a cluster.
Nothing user-facing installs it.

That surface is also where secret values are reviewed, and it is the ONLY place they are
shown: reports keep saying `set`/`MISSING`, `engine.Echo` keeps masking child env as
`NAME=***`, and `broker.AdditionalUsers` keeps withholding its CLI transcript. So each
deploy artifact has a secret-free render and a secret-bearing one beside it --
`generate broker` + `generate secrets broker`, and `generate operator` +
`generate secrets operator`. That second pair is why `Cluster.OperatorApply` issues three
applies (namespace, regcred, bundle) instead of splicing the regcred into the bundle:
everything it applies is byte-for-byte what those two commands print, and the namespace
still precedes the namespaced secret, which is what a first install needs.

**Abbreviations: one set of rules, four sets of words (`internal/abbrev`).** Nothing
rewrites argv -- a command name is expanded by cobra's `Find`, a role by
`config.ParseRole`, a `--platform` value by `config.ParsePlatform`, a flag shorthand by
pflag, each at the moment it reads its own input. `internal/abbrev` is what they share:
a stdlib-only leaf (the `internal/output` shape, which is what lets both `config` and
`cli` import it) owning the RULES, not the vocabulary. A short form is lowercase letters
or digits, strictly shorter than its word, claimed once, and never shadows a canonical
name; `abbrev.New` panics on a declaration that breaks one, so every test that builds a
set is a validation test. Each package still declares its own set, because the canonical
words are already owned -- a platform's is the env file's section key, a command's is the
cobra tree's -- and there is no registry or init-time side effect: a caller that needs
several names them itself. Sets carry a `Note`, and a word with NO short form is declared
rather than omitted (`start`/`stop`: any two-letter form is ambiguous between them and
`status`, the one slip that costs an outage), so a deliberate absence is data that prints
and can be pinned instead of something a reader has to notice is missing. Errors render
from the set (`Set.List()`), never from a list typed beside it. `commandAbbrev`
(`internal/cli/aliases.go`) is keyed by command NAME and applied by a tree walk, so a word
means the same thing wherever it appears (`br`, `op` ride under every verb that takes
them); `applyAliases` keeps its own panic on a collision with a sibling's name or alias,
since that is the only thing that also catches a command's hand-written `Aliases`. One
rule CLI-wide: abbreviations work everywhere, completion teaches the canonical spelling.
[docs/abbreviation.md](docs/abbreviation.md) is the glossary and is **generated** from the
three sets plus pflag's shorthands by `internal/cli/abbrevdoc_test.go`, sharing
`commanddoc_test.go`'s `-update` flag -- so one `regen` rewrites both goldens, and never
hand-edit either.

**Platform resolution (`internal/cli/platform.go`).** There is ONE command tree, not one per
platform. The platform is a property of the deployment the env file already describes, so
`resolvePlatform` reads it from there: `--platform` wins and must name a section the file
declares; otherwise the declared sections decide -- exactly one is used silently, none is a
loud error, several prompt (`interactive(a)`) or fail loudly when nothing can be asked.
`config.DetectPlatforms` must probe the RAW YAML, because a decoded `Config` cannot answer
the question: the platform sections are value structs, so omitted and empty are the same
zero value, and `ApplyDefaults` writes into `kubernetes.runtime` on every platform. A file
must therefore declare its section even when empty (`docker: {}`) -- the container schema
has no mandatory field that would otherwise force the section to exist.

The tree is the union of every platform's commands and is the SAME shape on all of them,
because help and completion render it without an env file. Applicability is therefore said
in the help text and enforced at pre-run (`onlyOn` / `flagOnlyOn` / `rejectRole`), never
expressed by hiding a command. For a command whose implementation differs, the
`platformOps` map is the single source of truth: its keys set the annotation that help and
the refusal read, its values are what runs, so the two cannot disagree.

The env-loading hook lives on each runnable command's `PreRunE` (`wireExec`), NOT on root's
`PersistentPreRunE`. Cobra runs the nearest ancestor's persistent hook and `__complete` is
root's own child, so a hook on root would parse an untrusted env file on every TAB press.
`TestCompletionNeverReadsTheEnvFile` pins it. `--allow-command` is likewise a non-persistent
flag on each command that executes, rather than a root flag, which is also what keeps it a
usage error on `convert`.

**The `broker:` section** holds the post-deployment configuration every platform applies over
the broker CLI -- `cliScriptsFolder`, `diagDir`, `productKeys`, `domainCerts`. It is
platform-neutral on purpose: platform detection (`config.DetectPlatforms`) works by which
top-level section an env file declares, so keeping these fields under `kubernetes.*` would
force a container env file to carry a `kubernetes:` section too and make its own platform
undetectable.

`internal/convert` is the one-way migration aid behind `solace-util convert`: it parses a legacy
bash env file (the pre-Go `bash/env/<name>` format), maps the `SOLBK_*`/`SOLOP_*`/
`IMAGEREPO_*`/`REPL_*` variables onto the YAML schema, and emits only what the source
actually set. It depends on `internal/config` (schema + validation) and nothing else, so
`config` must never import it -- the invalid-YAML hint in `config.Load` therefore carries its
own bash-file sniff rather than calling into `convert`.

**`examples` is the only top-level noun, and its templates are embedded
(`internal/examples`).** `solace-util examples [name]` writes a starting env file:
`kubernetes`/`docker`/`podman` are minimal standalone files declaring only their own
section (so the file they write needs no `--platform`), and `full` is the complete
annotated schema. It is a bare `cobra.Command` literal beside
`convert`/`version`/`auto-complete`, NOT a `leaf`: it emits an env file, so it must not
load one, and not calling `wireExec` is the whole opt-out -- which is also why it never
offers `--allow-command`. `TestExamplesNeedsNoEnvFile` pins that, and the two
hand-maintained exemption lists (`TestEveryRunnableCommandIsWired`,
`TestAllowCommandIsRegisteredWhereItExecutes`) both have to name it. A platform
abbreviation (`eg dk`) is expanded by `config.ParsePlatform` inside the command, at the
moment it reads its own input -- the same rule as every other abbreviation site;
completion still teaches only the canonical names.

`//go:embed` cannot reach outside its own package, so `internal/examples/assets/full.yaml`
is the AUTHORITATIVE copy and **`env/sample.yaml` is generated from it** by
`internal/examples/examples_test.go`'s `-update` flag. Never hand-edit the sample: it is
the shared fixture for `internal/{cli,render,k8s}` and the file every doc points at, so
`regen` runs `./internal/examples` FIRST -- the render and k8s goldens are rendered from
the sample it rewrites, and the reverse order would leave them a pass behind.

Every test in the repo is catalogued in [docs/test.md](docs/test.md) -- what each one proves,
the per-package fixtures and doubles to reuse, and the injectable seams. Update it in the
same change when you add or remove a test.

### Execution guard (`internal/config/execguard.go`)

`kubernetes.runtime`, `docker.runtime`, `podman.runtime` and `docker.compose` name binaries this
process runs, and env files travel, so config text must not be able to choose what executes.
One function, `config.CheckCommand`, is the whole rule: every token passes a charset (control
characters, Unicode whitespace via `unicode.IsSpace`, invisible `Cf` formatting characters,
and the shell metacharacter set -- the Unicode halves matter because the scalar YAML form
splits on `strings.Fields`, so an ASCII-only check would accept through the sequence form
what the scalar form rejects; pinned by `TestCharsetAgreesAcrossBothYAMLForms`); argv[0]
is a bare name (no `/` or `\`, one optional `.exe` stripped) from a per-platform allowlist;
every later token is a flag, a flag's value, or another allowlisted binary -- never a bare
word, which would smuggle a subcommand ahead of the one the tool appends. The acknowledged
limit is the flag-value position, which arity makes unvalidatable (pinned by
`TestFlagValuePositionIsNotGuaranteed`).

It is enforced **twice from that one definition**: in `Validate` (via `validateExecCommands`)
and again in every executor immediately before argv is built -- `Cluster.cmd`,
`kubectlTransport.cmd`, `Manager.runtime`/`composeCmd`, `containerTransport.runtime`, all of
which now return `(Command, error)`. The executors are handed a `*config.Config` directly and
must not assume `config.Load` ran. The one deliberate asymmetry: `Validate` skips an *unset*
field, since in this schema unset means "will be defaulted"; `CheckCommand` itself still
refuses an empty command, which is what protects exec.

The only way to widen the allowlist is the operator's `--allow-command` flag, threaded
through `config.Load`'s variadic tail into an **unexported** `Config.extraAllowed`. It has a
floor: `neverAllowed` (sudo, doas, su, pkexec, run0, runas, gsudo) can be approved by nobody,
because escalating *here* hands every command the tool issues to whoever wrote the env file
-- `sudo solace-util ...` elevates one invocation the operator chose instead. `AllowCommands`
refuses them with that explanation, and `commandRules.allowed` strips the category again so
the outcome does not depend on that being the only door. There is
deliberately no schema key, no env var, and no binding layer -- an env file that could approve
its own binary would make the allowlist decorative. `internal/cli` registers the flag on each
command that executes (`wireExec`) and rejects it where nothing does (`renderOnly` annotation,
the mirror of `genCapable`) -- and it is absent from root, so `convert` never offers it.

`healthCheck.cmd` is the one command field that keeps the loose validation rules
(`validateProbeCommand`) instead of the execution guard: it is rendered into the
compose/quadlet artifact and run by the container engine *inside* the broker, so it never
becomes argv here.

Two supporting layers: `engine.Resolve` resolves argv[0] with `exec.LookPath` and treats
`exec.ErrDot` as an error (never the current directory). It is shared by `engine.Exec`, which
resolves immediately before running, and by the CLI, which resolves the binaries the env file
names (`kubernetes.runtime`, `docker.runtime`, `podman.runtime`, `docker.compose`) once at load and
prints them as `==> using <name>: <path>` -- the location the allowlist cannot guarantee,
reported with the rest of the preamble rather than repeated on every call. `Exec` itself is
silent unless `-v/--verbose` installs its `Announce` hook, which traces every command as
`==> exec: <resolved path> <args>`. And every mutating operation runs a
read-only preflight first -- `Cluster.Preflight` (`auth can-i <verb> <resource>`) and
`Manager.Preflight` (`<runtime> info`) -- which stops before the first write, passes the CLI's
own error through, adds one actionable hint, never authenticates on the operator's behalf, and
has no skip flag.

[docs/commands.md](docs/commands.md) is the full CLI reference and is **generated** from the
cobra tree by `internal/cli/commanddoc_test.go`. It is a golden: `test` fails while it is
stale, so any command, flag, or `Short` change means regenerating it in the same change with
the `regen` task (deliberately outside `all`/`full`, since a gate must not rewrite what it
compares against). Never hand-edit it. The "Applies to:" line comes from the
`platformAnnotation` marker, so a command scoped to one platform must be wrapped in
`onlyOn`; the "Also available as:" line comes from cobra's `Aliases`, which `applyAliases`
fills from the abbreviation table.

Shell completion is owned rather than inherited: `newCompletionCmd` replaces the one cobra
would add during `Execute`, which never reached the golden because that renders the tree
without executing it. Two conventions follow, both pinned by `completion_test.go`. A command
taking a `[role]` sets `ValidArgs: config.RoleNames()` -- `leaf` and `roleLeaf` already do it,
the inline ones must say so. A flag taking a value that is not a plain file path registers a
completer next to where it is declared (`registerFlagCompletion`), or it silently falls back
to filename completion. Completion never loads the env file: the hook lives on each command's
own `PreRunE`, which is never in `__complete`'s ancestry, and keeping it that way is what
stops a TAB press from parsing untrusted YAML or printing into the shell.

### Output conventions (`internal/output`)

**One leaf package owns every stdout/stderr shape**, so a convention cannot drift per call
site the way five of them once did (`[ OK ]`, `[ERROR]`, `[WARN]`, `[Info]` and a bracket-free
`Login OK` all coexisted, and the same tag rendered as `==> [WARN] ...` or `  [WARN] ...`
depending on which function printed it). `output` imports nothing but the stdlib, which is
what lets BOTH `engine` and `cli` use it and is why the `==> ` prefix is now defined once
instead of being spelled out in each.

Two voices, split by stream: **stderr narrates, stdout reports.** A stderr line is
`==> <phase>` (`Sink.Step`) or `[TAG ] <msg>` (`OK`/`Warn`/`Fail`/`Info`) and NEVER both --
the tag already says what the line is. Every tag is the same width, which is why `[ERROR]`
became `[FAIL]`: a 7-character tag makes a column of mixed outcomes jog sideways. stdout
carries report bodies -- `Section` (a `=== title ===...` header padded to `SectionWidth`),
`KVBlock` (key column computed from the longest key), `KVRow` (an explicit shared width, for
the `check` reports whose rows come from four different methods so no call site sees every
key) and `Table` (per-column widths, last column unpadded). Plain ASCII, no ANSI colour:
these lines are read in PowerShell 5.1, captured into dev-script logs, and grepped.

`Cluster`/`Manager`/`Ops` reach it through the field they already had: `Log` is now the RAW
line sink -- it takes a formatted line and emits it verbatim, and each package wraps it in a
Sink (`progress()`) that adds the prefixes. That is why the CLI hands them `lineSink()` and
deliberately not `step`: passing a prefixing function would prefix twice. `report()` is the
stdout counterpart over the existing `Out`. `Cluster.Err` is new and is where the node
picker's prompt goes -- prompts are not report content, so `prepare labels > file` captures
results rather than questions (`TestPromptsGoToErrNotOut`). `engine.Echo`'s `+ <argv>` form
is untouched: it is the test-only runner's own shape, not this tool's voice.

### Container platform (`internal/container`)

Docker and Podman are one **host-local** platform: one container per host, so there is no operator and no cross-node control point (contrast `internal/k8s`, which drives the whole redundancy group from a single `kubectl` context). The moving parts:

- **Node-local transport** ([internal/container/transport.go](internal/container/transport.go)): a `broker.Transport` over `<runtime> exec`/`cp`. It **ignores the role arg** -- every op targets this host's single container -- so the CLI wires the shared `broker.Ops` config/verify methods with a nominal `config.Primary`. No `--` separator (docker `exec` rejects it); uploads ride stdin via `sh -c 'cat > <dest>'` (secret-safe, body never in argv).
- **Host Manager** ([internal/container/manager.go](internal/container/manager.go)): the container analog of `k8s.Cluster` -- `Check`/`PrepHost`/`Deploy`/`Delete`/`Status`/`Logs`/`CLI`/`Shell`. Podman renders a systemd quadlet unit; Docker a compose file. `Resolve`/`GenPSK`/`Geteuid` are injectable seams (defaults `net.LookupHost`, crypto/rand, `os.Geteuid`) so DNS, PSK generation, and the rootless/rootful euid guard are testable off a Linux host.
- **Secrets are files on every platform**, read through the broker setting's `*filepath` variant and mounted at `/run/secrets/<setting>` -- the same naming the k8s credentials Secret uses for its data keys. Host-side names carry `container.name` so two brokers on one host cannot collide. Podman mounts from its own store (`Secret=...,type=mount`); Docker's compose secrets are **environment-sourced** (`environment: <VAR>`), and `Deploy` passes the values to the compose child through `engine.Runner.RunEnv` -- nothing secret is written *beside the artifact* (docker materializes each one into the container's own filesystem as a 0444 root-owned file, verified with `docker diff`, so it is on disk exactly as long as the container is and survives a restart with no variable in the environment). `--restart` is what applies a rotated value, since no artifact changes when a password does; on docker, redeploy also force-recreates a *stopped* container rather than starting it, since a plain start would replay the credentials it was created with and silently miss a rotation. Podman's not-running branch runs a plain `systemctl start`, on the assumption that quadlet replaces the container at each start and so needs no equivalent fix -- marked `ASSUMED, NOT VERIFIED` in the code, since podman was not testable here. `Echo.RunEnv` masks values as `NAME=***`.
- **Primary-driven HA verification** ([internal/broker/verify_local.go](internal/broker/verify_local.go), [internal/broker/semp.go](internal/broker/semp.go)): the k8s `Leader`/`Redundancy` ops drive both pods from one kubectl context; containers get the same single-invocation shape from the primary host instead. Two facts make that possible: the primary's own `show redundancy` already reports the mate (`Mate Active`, ADB link fields), and the ONE command that must land on the backup -- admin `redundancy revert-activity` -- rides SEMP v1 over HTTP to `nodes.backup.ip` (curl exec'd in the local container, creds and body on stdin via `-K -`, mirroring `Login`; port 8080, or the bridge `network.ports` mapping of container port 8080, resolved by `sempPort` off `Ops.Platform`). `LocalRole(arg)` detects the role from the host name against the `nodes.*` table when `arg` is empty (loud error on no match); backup and monitor hosts are rejected loud on both ops. `RedundancyCoordinated` asserts the primary healthy, preflights the mate's SEMP (`/SEMP/v2/monitor`, BEFORE any mutation, so a firewalled mate aborts while the group is undisturbed -- HA working only proves ports 8300-8302/8741/55555), releases/un-releases, confirms the takeover via the primary's `Mate Active` count (the deliberate weakening vs k8s's independent backup reads), reverts the mate over SEMP, and polls activity home. `LeaderLocal` reverts the mate the same way first (k8s parity) but downgrades an unreachable mate to a warning -- its own job is local. The SEMP RPC body/reply shapes are marked NEEDS VERIFICATION ON A LIVE BROKER in semp.go. There is no fallback per-host handshake anymore; a failed run's recovery commands are in [docs/operations.md](docs/operations.md).
- **Config reuse + one divergence**: container `config`/`verify` read the shared `broker.*` fields (`domainCerts`, `productKeys`, `diagDir`, `cliScriptsFolder`) -- there is no separate container config namespace. Keeping them under `kubernetes.*` would force a container env file to declare a `kubernetes:` section too, which would make the file's own platform undetectable. `admin.additionalUsers` is shared too, but delivered differently: containers create the users at boot from the mounted password file plus a `username_<u>_globalaccesslevel` setting, while k8s creates them post-deployment over the broker CLI (`config apply additional-users` -> `broker.AdditionalUsers`). Verified against a live cluster: extra `username_<u>_password` keys in the credentials Secret are ignored by the operator, and `extraEnvVars`/`extraEnvVarsSecret` -- the only declarative alternative -- would expose the passwords in the pod environment. The CLI op therefore never shows its output (the transcript repeats the passwords), deletes the uploaded script via `defer`, and fails rather than reconciling an existing user, so it is deliberately not re-runnable -- which is the reason there is no run-everything config step on either platform: a command that stopped halfway through on its second run would be worse than no command. `config`'s Long text carries the order that works on a fresh broker instead. `config leader` remains cross-host + primary-only on containers, so it fails loud on a backup or monitor host.

## Knowledge graph

`graphify-out/` (tracked in git -- `.graphifyignore` scopes what graphify *reads*, a
separate mechanism from `.gitignore`, which has no graphify entry) contains a persistent graph of the repo: [graph.html](graphify-out/graph.html) (interactive viz), [GRAPH_REPORT.md](graphify-out/GRAPH_REPORT.md) (god nodes, surprising cross-document links, suggested questions), [graph.json](graphify-out/graph.json), plus an Obsidian vault and per-community wiki. Use it to answer "what calls X?" / "which files touch concept Y?" without re-reading dozens of files. Rebuild with `/graphify .` after notable changes; incremental updates use `/graphify --update .`. Files listed in [.graphifyignore](.graphifyignore) are excluded from the corpus (graphify does **not** read `.gitignore`, so the two files are maintained independently).

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).

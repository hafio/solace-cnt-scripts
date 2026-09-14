# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

A single Go binary, `solace`, that deploys and operates Solace PubSub+ Event Brokers on Kubernetes (via the Solace EventBroker Operator / PubSubPlusEventBroker CRD), Docker, and Podman. You describe the broker once in a YAML env file and drive the whole lifecycle from one standardized command tree. **Unsupported** -- not a Solace product. [README.md](README.md) is the quick start; [docs/configuration.md](docs/configuration.md) is the full env-file reference and [docs/operations.md](docs/operations.md) the day-2 lifecycle.

## Go implementation (`solace-util` binary)

The `solace-util` binary presents one standardized lifecycle command tree across Kubernetes, Docker, and Podman. Build and the dev-script tasks are in [docs/developer.md](docs/developer.md), the lifecycle in [README.md](README.md) and [docs/operations.md](docs/operations.md); package layout is `internal/{abbrev,config,engine,render,broker,k8s,container,convert,cli}` + `main.go`,
plus `internal/tools/vulnjudge` -- a dev-only command the `scan` task pipes govulncheck's
JSON through, so a fixable vulnerability fails the gate and one with no released fix warns.
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

**Noun then verb, and nothing implicit (`internal/cli/commands.go`).** You name the thing,
then what to do to it -- `broker deploy`, `operator remove`, `broker status`. The two nouns
own their verbs, and `group()`'s `RunE` never dispatches to a platform op: run one bare and
it prints what it can do, give it a word it does not own and it fails loud instead of cobra's
default of printing help and exiting 0. `group()` is the helper that builds those; it
deliberately does not go through `leaf()`/`wireExec()`, because a group never reaches an
external command and so needs neither the pre-run nor `--allow-command`. `broker copy`,
`broker configure` and `broker perform` are groups for the same reason.

`TestNounGroupsRunNothing` is what pins that a group's `RunE` still dispatches to nothing,
and it is what makes the short forms safe: `br` and `op` on their own act on nothing. The invariant
inverted with the tree -- it used to be the VERB that had to run nothing, because `rm` was
reachable at the top level; now `rm` is only reachable under a noun, and the noun is what
must be inert.

`group()` takes its `Long` as a parameter; the leaf builders (`leaf`/`dispatchLeaf`) take
only a `Short`, so `withLong(c, text)` (helpers.go) attaches one, wrapping inline the way
`onlyOn`, `renderOnly` and `rolePositionalTeachesPod` do rather than forcing a named
variable per command. A command built from a
`cobra.Command` literal sets the `Long:` field directly. Those are the only two shapes -- a
`c.Long = ...` assignment is drift. The text is a BRIEF summary of what the command does and
how to invoke it -- one to four lines; mechanics, constraints and history belong in
[docs/operations.md](docs/operations.md), not in help. A destructive command carries exactly
one sentence saying what is destroyed. The generator still skips a `Long` equal to the
`Short`, so a `Long` that would only restate it is not written at all.

**Removals keep the expensive layer (`flags.go`: `addRemoveFlags`; `confirm.go`: `layer`, `confirmLayer`).**
`broker remove` keeps persistent data (PVCs / the container data dir); `operator remove`
keeps the CRDs, whose deletion would cascade to every PubSubPlusEventBroker in the cluster.

The layer flag RAISES the question rather than answering it. Without `--delete-data` /
`--delete-crd` the layer is kept and nothing is asked; with it, an interactive run needs an
exact `yes`, `--no-prompt` answers yes, and a non-interactive run without `--no-prompt`
keeps the layer and says so loudly. So `--delete-data --no-prompt` is what a fully
unattended removal-with-data looks like, and the two flags compose rather than conflict.
There is no `--yes`: one silencer per command beats two flags whose overlap has to be
memorised. `TestConfirmLayerMatrix` pins all of it.

`broker remove` also owns what used to be `remove secrets`/`remove namespace`. The namespace
is never in the delete set: after everything this env file owns is gone, what is left is
listed, and only an otherwise-empty namespace is OFFERED for removal -- an occupied one is
kept on every path, `--no-prompt` included, because deleting it would cascade to whatever
someone else put there. `operator remove` leaves the operator running when other namespaces
still need it, narrowing its watch list instead. Every destructive command confirms,
`broker restart` included, as do `broker configure default-vpn` and
`broker configure default-users` when disabling (never when enabling) and
`broker configure domain-certs --remove`; all of them take the same `--no-prompt` silencer.

**Whether a human is there is decided by READING, never by the kind of stream stdin is**
(`confirm.go`: `stdinCanAnswer`, `promptLine`). The rule cost two bugs to learn. The
first test was `isTTY(os.Stdin)`, and in a shell whose stdin was a pipe it made every
confirmation in the tree a silent refusal -- which is how `export-config -o <existing
file>` came to decline an overwrite without asking, with the operator sitting at the
terminal. Reading the console device instead (`/dev/tty`, then `CONIN$`, what git and
sudo reach for) fixed that and was worse: in a mintty-style shell the keystrokes go into
the stdin pipe and never reach the console input buffer, so the process HUNG on a handle
nothing would ever write to.

So `interactive()` now asks only whether stdin is usable at all, the question goes to
stderr, the answer is read from stdin, and `promptLine` returns `answered=false` when the
stream ends before a reply arrives. Unanswered is NOT a "no": a reply of `n` is the
operator's decision, EOF means nobody could make one, and only the second earns the
refusal line naming `--no-prompt`. The trade is that the question is printed before it is
known whether a reply can arrive, so an unattended run shows its question above its
refusal -- strictly more informative than the silence it replaced. The one residual cost
is a run whose stdin is an open pipe nobody writes to or closes, which waits instead of
refusing; `--no-prompt` is the answer. `internal/cli`'s `TestMain` points stdin at an
already-closed pipe so no test can block on the console the suite was launched from.

**There is no `--dry-run` and no `--gen-*` flags.** `generate <target>` is the only
look-before-you-leap surface: rendering is a command with a named target rather than a flag
on the commands that deploy, so an artifact you meant to read and a cluster you meant to
change are not one typo apart. `engine.Echo` is a TEST-ONLY runner, installed through the
`App.NewRunner` seam (context.go) -- the same shape as the `Interactive`/`PromptIn` prompt
seams, and the reason ~200 wiring tests can still assert exact argv without a cluster.
Nothing user-facing installs it.

`generate` has no `secrets` sub-tree: one command per noun prints everything that noun's
deploy applies. On Kubernetes that means the Namespace, the Secret manifests and the CR --
each emitted before whatever references it, so the stream applies in order to an EMPTY
CLUSTER, not merely to a namespace someone already made. `k8s.GenBroker` and
`k8s.GenOperator` are the two definitions of that stream, and both are pinned against what
their deploy actually applies (`TestGenBrokerMatchesWhatDeployApplies`) rather than
described alongside it. On
docker and podman no secret value is rendered at all, because neither engine can carry one
in its artifact -- a quadlet references the podman store by name and a compose file names an
environment variable, so the values are the deploy's business and nothing else's.

That surface is also where secret values are reviewed, and it is the ONLY place they are
shown: reports keep saying `set`/`MISSING`, `engine.Echo` keeps masking child env as
`NAME=***`. So each
noun has exactly ONE generate, and on Kubernetes it carries the values -- a Secret
manifest IS the artifact there, so there is no secret-free variant to reach for and the
output is as sensitive as the env file. On docker and podman the same command carries no
secret at all, because neither engine can put one in its artifact.

`Cluster.OperatorApply` issues three applies (namespace, regcred, bundle) instead of
splicing the regcred into the bundle so that everything it applies is byte-for-byte what
`operator generate` prints, with the namespace still preceding the namespaced secret --
which is what a first install needs. `k8s.GenBroker` is the same contract for the broker
half, and `TestGenBrokerMatchesWhatDeployApplies` is what holds the two together.

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
in the help text and enforced at pre-run (`onlyOn` / `flagOnlyOn`), never
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
(`internal/examples`).** `solace-util examples` writes a starting env file, the template chosen by `--platform`:
`kubernetes`/`docker`/`podman` are minimal standalone files declaring only their own
section (so the file they write needs no `--platform`), and `full` is the complete
annotated schema. It is a bare `cobra.Command` literal beside
`convert`/`version`/`auto-complete`, NOT a `leaf`: it emits an env file, so it must not
load one, and not calling `wireExec` is the whole opt-out -- which is also why it never
offers `--allow-command`. `TestExamplesNeedsNoEnvFile` pins that, and the two
hand-maintained exemption lists (`TestEveryRunnableCommandIsWired`,
`TestAllowCommandIsRegisteredWhereItExecutes`) both have to name it. A platform
abbreviation (`eg --platform dk`) is expanded by `config.ParsePlatform` inside the command, at the
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
command that executes (`wireExec`) and rejects it where nothing does (the `renderOnly`
annotation) -- and it is absent from root, so `convert` never offers it.

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

Cobra's WINDOWS-ONLY Explorer check is switched off in `newRootCmd`
(`cobra.MousetrapHelpText = ""`). Left on, it decides whether the binary was
double-clicked and, if it thinks so, prints "this is a command line application" and
DECLINES to run -- a failure mode a tool driven from terminals, scripts and CI should not
have. It is also expensive: the hook walks the OS process table on every `Execute`
(`Getppid` plus a snapshot scan), which a CPU profile put at 37% of `internal/cli`'s whole
test run -- 10.6s of 13.8s, and 85% of everything cobra did, against 1.8s for the commands
under test. Operators pay it once per invocation. `TestMousetrapIsDisabled` pins it,
because nothing else would notice it coming back: every test would still pass, just slower.

Shell completion is owned rather than inherited: `newCompletionCmd` replaces the one cobra
would add during `Execute`, which never reached the golden because that renders the tree
without executing it. Two conventions follow, both pinned by `completion_test.go`. A command
that names a role does it with `--pod`, whose completer `addPodFlag` registers for it -- no
command takes a `[role]` positional any more, and `TestNoCommandTakesARolePositional` keeps
it that way. A flag taking a value that is not a plain file path registers a
completer next to where it is declared (`registerFlagCompletion`), or it silently falls back
to filename completion. Completion never loads the env file: the hook lives on each command's
own `PreRunE`, which is never in `__complete`'s ancestry, and keeping it that way is what
stops a TAB press from parsing untrusted YAML or printing into the shell.

### Config export and import (`internal/broker`)

`broker perform export-config [--vpn NAME]... [--broker-only] [-o FILE]`
and `broker perform import-config <file>` are named that way rather than
`backup`/`restore` for two reasons: `backup` is already a redundancy role in this
CLI (`config.Backup`, `--pod backup`, `docs/abbreviation.md` declares `backup` ->
`b`), and the artifact provably cannot rebuild a broker -- it carries no product
key, no TLS private key, no Kerberos keytab and no message data -- so calling it a
backup would promise something the tool cannot keep.

**The channel.** Both commands go entirely over `Ops.RunCLI`: `<runtime> exec
<pod|container> -- sh -c '<skeleton>'`, the skeleton writing the script, wrapping
it in the broker's own `source script <name> stop-on-error no-prompt`, exec'ing
`.../bin/cli -Apes <wrapper>` and cleaning up after itself. Three properties of
that channel shape the whole design. No broker credentials are involved --
authorisation is kubectl/docker access to the container, not SEMP login -- so an
import that overwrites the CLI admin password (`Create Usernames`, below) cannot
lock the tool out of its own run. Every `RunCLI` execs a NEW `cli` process, so each
invocation is a fresh session; that is what makes it safe to apply a
session-ending section in an invocation of its own. And `cli -Apes` ALWAYS EXITS
0, even when the broker rejects a line, so the transcript is the only evidence
there is -- which is what `RunCLI`'s wrapper and its `rejectionIn` scan exist to
make readable, and why `internal/broker/diff.go`'s re-export-and-compare is a
second verdict rather than the only one.

**`RunCLI` writes, `runCLIRead` reads, and the split is not a style choice**
(broker.go). A write goes through the stop-on-error wrapper and is scanned by
`rejectionIn` over the transcript tail: the broker stops at the first rejected
line, so the rejection IS the tail and the scan is sound BY CONSTRUCTION rather
than by luck. A script that ends in a `show` (serverCertScript,
assertLeaderScript) is still a write and still sound -- if the mutation ahead of
it is rejected, the `show` never runs. A READ goes through `runCLIRead`, which is
the old unwrapped shape byte for byte, for two reasons. The wrapper would add
nothing (there is no later line a rejection could poison), and the scan would be
actively wrong: `failKeywords` carries the bare word `busy`, vetted against
configuration-capture text and never against runtime state, so a `show redundancy`
poll during an HA restoration could abort the whole verification instead of
polling on. Every parser here -- `bannerType`, `stripTranscript`'s `! BEGIN`
search, the replication parsers, `verify_ops.go`/`verify_local.go`'s `field()` and
`countContains()` scans -- was tuned against real UNWRAPPED transcripts, and the
wrapper's echoed `source script ...` line is one no golden has seen. `RunCLI`
self-cleans (the skeleton's own `trap`); `runCLIRead` does not, so its callers own
their `removeCLI` exactly as before.

A rejection is wrapped with the `ErrCLIRejected` sentinel, and a caller that must
tell it from an unreachable broker asks `errors.Is` rather than matching a phrase
in the message. The distinction is not cosmetic: a rejection means the broker
stopped at a known line, so what landed is exactly the lines before it, while a
lost connection means the sourced script kept running INSIDE the broker and how
much of it applied cannot be known from here. `configure dr`'s two phases are the
one caller that needs it, and they said it with `strings.Contains` on `RunCLI`'s own
wording first -- so rewording that one sentence would have silently downgraded every
rejection report into "cannot be known", with no test to notice.

**Why the CLI and neither SEMP API**, stated as schema facts so it is not
re-litigated: `show current-config` has zero hits in either SEMP v1 XSD
(`semp/semp-rpc-soltr.xsd`, `semp-rpc-reply-soltr.xsd`) -- it is CLI-only. The v1
reply schema models only the `show` verb, interleaves configuration with live
runtime state with no marker separating them, and its request root allows one
verb per `<rpc>`. SEMP v2 config's whole writable surface is 11 paths (`/` plus
`msgVpns`, `certAuthorities`, `clientCertAuthorities`, `domainCertAuthorities`,
`dmrClusters`, `oauthProfiles`, `proxies`, `virtualHostnames`) with NO endpoint
for product keys, CLI users, redundancy, interfaces, syslog, spool sizing,
router-name or system scaling -- strictly less complete than the CLI capture, and
its credentials surface is write-only besides.

**The artifact (`internal/broker/blocks.go`)** parses a `show current-config`
capture into `Block`s -- one column-0 opener plus its indented children -- inside
four `Region`s the broker emits in a fixed order: the mode preamble, the pass
that creates every message-VPN empty, the broker-level configuration, then each
VPN's own objects. `--broker-only` has to exclude the create-VPN pass and not
merely the per-VPN region, which is why that boundary is its own `Region` rather
than folded into VPN scope. Every broker-level `Block` carries the broker's own
`! Configure X` / `! Create X` comment as `Section`, which is the classification
key `sections.go` matches against. Block boundaries come from INDENTATION, never
from matching `exit`: a block's own `exit` sits one level deeper than its opener
(`authentication` at indent 2 closes at indent 4; `message-vpn` at indent 0
closes at indent 2), so hunting for the matching `exit` would need a full nesting
parse and a miscount would silently merge two objects. The opener is tokenised
quote-aware rather than split on whitespace, because a real capture creates the
VPN `"A VPN WITH LONG NAME AND SPACES"` -- a whitespace split renames it to `"A"`
with no error anywhere. A `! pragma:` line sits at column 0 but is never a
boundary: it is held and attached to the block that follows, because it can
appear between a section comment and its opener as well as inside a block, and
losing a held `ignore-already-exists` changes whether a `create` line can be
replayed.

**The marker layer (`internal/broker/annotate.go`)** re-renders a parsed capture
with this tool's own `! solace-util` fences -- a metadata header, then region and
section BEGIN/END markers -- so an operator can read an import's plan straight
out of the file instead of import inferring it. BOUNDARIES ARE AUTHORITATIVE,
DISPOSITION IS ADVISORY: a section fence records what THIS BUILD would do with
that section, but `ImportPlan` always re-resolves the disposition from
`sections.go` and warns when the two disagree, never trusting the artifact's own
say-so. That is the same argument `internal/config/execguard.go` makes for
`kubernetes.runtime`/`docker.runtime`/etc. -- an artifact that chose what to
execute would be untrusted config text doing exactly what the execution guard
exists to prevent, just carried in a different file than the env file.

**The classification (`internal/broker/sections.go`, 50 rules)** is BAKED IN --
no config key, no flag, no runtime decision file -- for the same reason the
marker's disposition is advisory rather than authoritative. `docs/import.md` is
GENERATED from `Rules()` by `internal/broker/importdoc_test.go`, so what an
import does cannot drift from what the docs say; `regen` now runs
`./internal/broker` too. The provenance check is deliberately NOT tamper-proofing: the marker is a comment,
and an artifact can be edited with it left in place. It stops the wrong file --
a raw `show current-config` capture still carries the sections import must never
apply and the CLI transcript export cuts off, so importing one would sever the SEMP
channel and feed echoed prompts to the CLI, both invisibly. It could not do more
while `broker perform cli-script` runs any script it is handed, which is that
command's job and what its help says.

Removed at EXPORT (`omitAtExport`, recorded in the artifact as
`! solace-util-omitted:` lines so the file states its own absences), never applied, and
not opt-in-able -- import only ignores such a section if an older or hand-edited
artifact still carries one, and reports it only then: `Configure SEMP
Service` (it shuts SEMP down to change its own port, and SEMP is the channel
import runs over), `Configure Router Name`, `Configure Redundancy`, `Create
Redundancy PSK`, `Configure Matelink Service`, `Configure Redundancy Service`,
`Configure Config Sync`, `Configure System`. Three sections are applied against
the classification's own recommendation, on the operator's explicit decision,
and their `Reason` records the consequence rather than an objection: `Create
Usernames` overwrites the target's CLI admin password (harmless to this tool,
but `semp.adminPass` then has to be updated to match or Kubernetes pods go out of
readiness against the old `adminCredentialsSecret`); `Configure Replication`
carries a config-sync PSK and mate router-name, so the mate needs re-importing to
match; `Create Domain Certificate Authority` deletes every domain CA on the
target first and overlaps `broker configure domain-certs`.

**Import is three phases (`internal/broker/importops.go`).** `ImportPlan` parses
the artifact, refuses anything this tool did not export (`checkProvenance`: no
`! solace-util-export:` marker, or a format version this build does not read), refuses
a redacted one (redaction strips exactly the credential material import has to put
back -- export never asks the broker to redact, so this guard is for a hand-edited
header rather than anything this tool writes; there is no `--redact` flag, because it
could only produce a file import refuses), reads the ARTIFACT's broker type from its SEMP
schema version prefix (`broker/` software, `app/` appliance) and the TARGET's from
the CLI login banner that every `cli -Apes` invocation echoes ahead of its output
(`bannerType`: the `license-software`/`license-hardware` URL, then the product
line -- a `show version` scan came first and refused a valid target, because its
output never carried the schema string), refuses a cross-type artifact, then
splits the artifact's VPNs into Existing and New against `show message-vpn *`.
The type is read off THAT read's own raw transcript, before `stripTranscript`
cuts it, rather than from a round trip of its own: a dedicated `home` script cost
an upload, an exec and a remove to learn something the very next call already
carried -- three avoidable `kubectl exec`/`docker exec` per plan, and per
`configure dr` against a mate that may sit across a WAN. `readMateConfig` reads
its type the same way.
That same echoed banner is why every capture is cut at the broker's own
`! BEGIN` line before it is parsed or replayed (`stripTranscript`), and why an
apply script takes only the mode commands and pragmas from a preamble
(`preambleForApply`) -- fed back verbatim, `xps-ps-01> home` is a rejected
command, invisibly. `ImportApply` tears down and rebuilds every
EXISTING VPN rather than reconciling it in place, because a Solace CLI
configuration block cannot be applied a second time -- and the teardown script is
captured FROM THE TARGET with `show current-config message-vpn <n> remove` and
applied as-is, so the broker orders the removal of its own queues, ACL profiles,
client profiles, usernames and bridges and this tool never encodes that order
itself. That destroys the messages spooled in the VPN's queues, which is why VPN
scope takes the exact-`yes` gate `--delete-data` takes (a new
`confirmActionStrict` in `internal/cli/confirm.go`), asked ONCE for the whole
run; broker scope never asks at all, because which sections it applies is the
fixed classification above, not a per-run decision. `ImportVerify` re-exports the
target and diffs it against the artifact (`internal/broker/diff.go`); the
comparison is DIRECTIONAL -- every artifact line must be present on the target, not
set equality -- because the target's own export also emits defaults the source
capture never carried.

**The apply is CHUNKED and the loop runs inside the broker
(`internal/broker/driver.go`).** `ImportApply` renders one generated `/bin/sh`
script carrying every chunk as a quoted heredoc, uploads it ONCE and executes it
ONCE. Each chunk is applied through the broker's own
`source script <name> stop-on-error no-prompt` -- so the BROKER stops at the first
rejected line instead of running on through everything after it -- and the driver
then greps the last 10 lines of that chunk's transcript for a rejection
(`failKeywords`). Three things follow. A failure is caught DURING the apply
rather than inferred from the diff afterwards; a failed TEARDOWN is caught at all,
which the diff structurally cannot do, because a VPN that was never removed still
satisfies every line the artifact asks for; and a real import costs one exec
instead of dozens. Chunk order is teardowns, `Create logging` (alone -- it ends the
CLI session), `Create Usernames` (alone -- it rewrites the CLI admin password), the
remaining broker sections, then one chunk per message-VPN with **`default` first**,
since `default` is the VPN that is edited rather than recreated and carries the
port changes most likely to fail.

**The two detectors are SEQUENTIAL, not both-always** (`runImport`,
`internal/cli/exportconfig.go`). A failed apply reports and returns; `ImportVerify`
runs only after a fully successful one. Diffing after a failure would report every
chunk the run never reached as missing -- true, and useless, burying the one line
naming the actual failure. `TestImportConfigFailedApplySkipsVerification` pins it.

Only marker lines cross back: `SOLACE-UTIL-CHUNK-OK|FAIL <index>`, and on failure
the transcript tail. A successful chunk's transcript never leaves the broker,
because applying a configuration repeats every credential in it, and a `trap`
deletes every chunk and temp file on EXIT/INT/TERM so an interrupted import leaves
nothing behind. Chunks are addressed by INDEX in the shell -- a chunk is named
after a message-VPN, and a real one is called `A VPN WITH LONG NAME AND SPACES`.
`failKeywords` is the operator's list and is COMPLETE as it stands (confirmed
2026-09-13; it began as a placeholder the day before). A new phrase is added when real
rejection output turns one up, as its own change -- it is not provisional, and it is
now the ONLY list: every CLI write in the tree scans with it through the one
`rejectionIn`, whether it comes from the import driver's per-chunk grep, `RunCLI`,
`broker configure data-replication` or the DR CLI leg. The two whole-transcript
scanners that used to sit beside it are gone -- `ProductKeys`'s
`containsAnyFold("error","fail")` and `ExecCLI`'s `countAnyFold("invalid","error",
"busy")` had the OPPOSITE failure mode from the tail scan, false-positiving on an
object legitimately named `error-events`, and four `config_ops` operations
(`ServerCert`, `DomainCerts`, `defaultUsers`, `Remove*Certs`) had no detection at
all and returned nil over a rejected line.
The four in it match nothing in 33k lines of real capture, including all 917
`reject-msg-*` lines.
A phrase MISSING from the list fails quietly rather than loudly -- the chunk reports
OK and only the diff afterwards notices -- which is the designed fallback, and the
reason the list wants to be generous rather than minimal.

**`InjectShutdown` (`internal/broker/inject.go`)** makes a captured
`service X listen-port ...` line safe to import into a broker that refuses a
port change while the service is up: it injects the LEADING shutdown only and
NEVER synthesises the trailing enablement state, because a capture can
legitimately end with a service shut down (nothing forces the un-shutdown to
exist). When the artifact carries no enablement line at all for that service, it
reads the TARGET's current state and appends a restore line reproducing exactly
that reading, so the round trip is a no-op on enablement; when even that read
comes back unknown, it injects nothing and warns rather than guessing in either
direction. It is BLOCK-level, and that is what left a hole for a long time: a
broker-level `service semp listen-port 8080` is its own column-0 block, but the
VPN-level equivalents are indented CHILDREN of one `message-vpn "X"` block, so no
block-level rule reached them. There were 48 of them in the real capture, all
unhandled.

**`internal/broker/transform.go` is that hole closed, and the rest of
`semp/CurrentConfigParser.html` with it.** That file is the operator's hand-built
GUI for preparing these captures for migration by hand, and it is the reference for
what a replay actually needs. Ported, as pure functions over `[]Block`:
VPN-level port shutdowns per service AND TRANSPORT (`service mqtt` alone has
plain-text, ssl, websocket and websocket-secure in one capture, so a blanket
`service mqtt shutdown` would drop three of them to change one);
`shutdown` for each VPN's `default` client-username, which cannot be deleted and so
is edited in place; bridges forced to `shutdown`, because a bridge that comes up
starts moving messages between two brokers before anyone has read the result;
client-profile/acl-profile/client-username moved to the END of their VPN, since a
username naming a profile the broker does not yet have is refused;
`create username "admin"` rewritten to `username admin`, since admin always exists
and `create` is a duplicate; consecutive duplicate lines collapsed. NOT ported: the
GUI's hardcoded replication PSK (the artifact's own captured auth applies verbatim,
and a credential does not belong in source), its `ssl-*`/`auth-scheme`
comment-out, and its config-sync line filter (that section is already removed at
export). `PrepareForApply` runs them in the one order that works and is pinned
IDEMPOTENT against the real capture -- re-running a failed import is the documented
recovery, so a transform that grew the script each pass would apply something
different the second time.

**`ClearExistingSyslogs` needs the TARGET, not the artifact.** A `create syslog` for
an entry that already exists is rejected, but `no syslog` for one that does not
exist is itself an error -- so the removal can only be emitted for names the target
actually has. `readTarget` (renamed from `listVPNs`) returns the whole parsed target
capture for exactly this reason: one read answers both the VPN split and the syslog
set, so comparing against the target costs NO extra round trip.

The alternative -- emit `no syslog` unconditionally and ignore the error when the
entry is absent -- was considered and rejected (2026-09-12). It is not free here:
the driver applies every chunk with `stop-on-error` and then greps the tail for
`error:`, which is exactly what that removal produces, so tolerating it needs a
chunk whose failures are deliberately unchecked -- a hole in the only mechanism that
detects a rejected line. Comparing costs nothing and keeps the guarantee absolute.

Two shapes are confirmed against a real capture (semp/all-cfg-tps-org.cli:252-253,
and semp/all.cli before the local copy of it was replaced with a capture from a
broker that configures no syslog destination at all): each
destination gets its OWN `! Create Syslog: "<name>"` section, so the rule is matched
by prefix and there is one per entry; and the operand is quoted, which is what the
emitted removal mirrors. Broker scope stays ADDITIVE: a syslog destination the
target has and the artifact does not is left alone, because dropping one the
target's own operator configured is a log-delivery outage nobody asked for.

**Undeletable objects need no special case, and that is the point of capturing the
teardown.** The `default` message-VPN cannot be deleted, and neither can the
`default` client-profile, acl-profile or client-username inside any VPN -- which is
why `broker configure default-vpn` SHUTS THE VPN DOWN rather than removing it.
Verified on a live broker (2026-09-11): a `remove` capture of the default VPN simply
does not emit lines for the VPN itself or for its undeletable
objects, so it empties `default` and leaves it present, and the artifact then
applies onto it. A synthesised teardown would have needed a reserved-name list kept
correct forever; the broker already knows, so it is asked. There is no filter and no
`default` branch anywhere in the import path.

**A VPN-scoped capture carries its own `create message-vpn` line.** VERIFIED ON A
LIVE BROKER (2026-09-11): `show current-config message-vpn <n>` emits the create.
That is what makes VPN scope survivable at all -- import tears the VPN down first,
so a capture with no create line would leave nothing able to rebuild it, and the
one VPN-scoped capture in the repo (the `remove` variant, which legitimately has
no create) made that look like the likely shape. It is not.

Still marked NEEDS VERIFICATION ON A LIVE BROKER rather than asserted: that a
`remove` for a NON-default VPN removes the VPN itself and leaves no residue (the
default case above is confirmed; the general one is not), and that a VPN-scoped
import does not intersect what `render.BrokerCR` owns on Kubernetes.

New: `internal/broker/{blocks,annotate,sections,inject,diff,importops}.go`,
`internal/cli/exportconfig.go`. `docs/import.md` is generated; never hand-edit it.

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
stdout counterpart over the existing `Out`. This package no longer prompts at all -- the node
picker went with the labelling surface, and its one remaining question (the operator
downgrade) goes through the `Confirm` func seam -- so every confirmation now lives in
`internal/cli`, which keeps prompts on stderr for the same reason: redirecting stdout should
capture results, not questions. `engine.Echo`'s `+ <argv>` form is untouched: it is the
test-only runner's own shape, not this tool's voice.

### Container platform (`internal/container`)

Docker and Podman are one **host-local** platform: one container per host, so there is no operator and no cross-node control point (contrast `internal/k8s`, which drives the whole redundancy group from a single `kubectl` context). The moving parts:

- **Node-local transport** ([internal/container/transport.go](internal/container/transport.go)): a `broker.Transport` over `<runtime> exec`/`cp`. It **ignores the role arg** -- every op targets this host's single container -- so the CLI wires the shared `broker.Ops` config/verify methods with a nominal `config.Primary`. No `--` separator (docker `exec` rejects it); uploads ride stdin via `sh -c 'cat > <dest>'` (secret-safe, body never in argv).
- **Host Manager** ([internal/container/manager.go](internal/container/manager.go)): the container analog of `k8s.Cluster` -- `Check`/`PrepHost`/`Deploy`/`Delete`/`Status`/`Logs`/`CLI`/`Shell`. Podman renders a systemd quadlet unit; Docker a compose file. `Resolve`/`Geteuid` are injectable seams (defaults `net.LookupHost`, `os.Geteuid`) so DNS and the rootless/rootful euid guard are testable off a Linux host. There is deliberately no PSK generator seam any more: `redundancy.psk` is the operator's to supply (mandatory in an HA container group, refused at load when empty), and nothing in this package writes to an env file.
- **Secrets are files on every platform**, read through the broker setting's `*filepath` variant and mounted at `/mnt/secrets/<setting>` -- the same naming the k8s credentials Secret uses for its data keys. That directory is deliberately NOT the engines' own `/run/secrets`, and the move has two halves that must stay together: `render.secretMount` is what the `*filepath` settings point at, and `ContainerSecret.Target()` must return the ABSOLUTE path, because a bare `target=`/`target:` is resolved by each engine under its own `/run/secrets` and the broker would then find nothing with no error anywhere. A path-valued target needs podman 4.5 and compose 2.23.1 (`docs/operations.md` version floors, none enforced at runtime). Host-side names carry `container.name` so two brokers on one host cannot collide. Podman mounts from its own store (`Secret=...,type=mount`); Docker's compose secrets are **environment-sourced** (`environment: <VAR>`), and `Deploy` passes the values to the compose child through `engine.EnvRunner.RunEnv` -- a SEPARATE interface from `engine.Runner`, which no longer carries that method, because this is the only caller in the tool that needs it and an interface offering the others the ability to put a value in a child's environment would be inviting a secret somewhere none of them should be putting one -- nothing secret is written *beside the artifact* (docker materializes each one into the container's own filesystem as a 0444 root-owned file, verified with `docker diff`, so it is on disk exactly as long as the container is and survives a restart with no variable in the environment). `--restart` is what applies a rotated value, since no artifact changes when a password does; on docker, redeploy also force-recreates a *stopped* container rather than starting it, since a plain start would replay the credentials it was created with and silently miss a rotation. Podman's not-running branch runs a plain `systemctl start`, on the assumption that quadlet replaces the container at each start and so needs no equivalent fix -- marked `ASSUMED, NOT VERIFIED` in the code, since podman was not testable here. `Echo.RunEnv` masks values as `NAME=***`.
- **Primary-driven HA verification** ([internal/broker/verify_local.go](internal/broker/verify_local.go), [internal/broker/semp.go](internal/broker/semp.go)): the k8s `Leader`/`Redundancy` ops drive both pods from one kubectl context; containers get the same single-invocation shape from the primary host instead. Two facts make that possible: the primary's own `show redundancy` already reports the mate (`Mate Active`, ADB link fields), and the ONE command that must land on the backup -- admin `redundancy revert-activity` -- rides SEMP v1 over HTTP to `redundancy.backup.addr` (curl exec'd in the local container, creds and body on stdin via `-K -`, mirroring `Login`; port 8080, or the bridge `network.ports` mapping of container port 8080, resolved by `sempPort` off `Ops.Platform`). `LocalRole(arg)` detects the role when `arg` is empty -- `DetectRole` matches this host's name against `redundancy.*.name`, then, only if nothing matched, this machine's own interface addresses (the `LocalAddrs` seam) against `redundancy.*.addr`, which is the cloud case where the routername and the OS hostname are legitimately unrelated; loud error on no match or an ambiguous one. An explicit `--pod` wins but is checked against the same detection and WARNS on a disagreement rather than prompting. A STANDALONE container deployment needs no node entry at all: `Config.FillStandaloneNodeName` names the broker after its host at load when `redundancy.primary.name` is empty, which HA deliberately cannot do because every name there keys a group-table entry all three hosts render; backup and monitor hosts are rejected loud on both ops. `RedundancyCoordinated` asserts the primary healthy, preflights the mate's SEMP (`/SEMP/v2/monitor`, BEFORE any mutation, so a firewalled mate aborts while the group is undisturbed -- HA working only proves ports 8300-8302/8741/55555), releases/un-releases, confirms the takeover via the primary's `Mate Active` count (the deliberate weakening vs k8s's independent backup reads), reverts the mate over SEMP, and polls activity home. `LeaderLocal` reverts the mate the same way first (k8s parity) but downgrades an unreachable mate to a warning -- its own job is local. The SEMP RPC body and reply shapes are now confirmed against the broker's own schemas (`semp/semp-rpc-soltr.xsd`, `semp-rpc-reply-soltr.xsd`): the request path is `rpc > admin > redundancy > revert-activity`, `semp-version` is optional on a request, and a reply carries `<execute-result code="ok|fail">`. What is still marked NEEDS VERIFICATION ON A LIVE BROKER in semp.go is the `/SEMP` endpoint path itself and whether reverting an already-standby mate is idempotent. There is no fallback per-host handshake anymore; a failed run's recovery commands are in [docs/operations.md](docs/operations.md).
- **Config reuse + one divergence**: container `config`/`verify` read the shared `broker.*` fields (`domainCerts`, `productKeys`, `diagDir`, `cliScriptsFolder`) -- there is no separate container config namespace. Keeping them under `kubernetes.*` would force a container env file to declare a `kubernetes:` section too, which would make the file's own platform undetectable. `semp.additionalUsers` is shared too, and every platform now applies it declaratively -- the users exist from the broker's first boot, with nothing to run afterwards. Only the delivery differs, and the difference is the CRD's rather than a preference. Containers mount the password as a FILE (`/mnt/secrets/username_<u>_password`, reached through `username_<u>_passwordfilepath`) with the access level in the artifact. Kubernetes puts BOTH halves in the pod environment, from a Secret of its own (`<kubernetes.name>-additional-users`, `k8s.AdditionalUsersSecret`) that the CR names in `spec.extraEnvVarsSecret`: the CRD offers `extraEnvVars`/`extraEnvVarsCM`/`extraEnvVarsSecret` and NO volume passthrough, so there is no way to mount an arbitrary Secret and `username_<u>_passwordfilepath` would name a path nothing creates.

That Secret is separate from `kubernetes.adminSecret` and must be: `extraEnvVarsSecret` is projected with `envFrom`, which exports EVERY key of the Secret it names, so pointing it at the credentials Secret would publish the admin and monitor passwords to get the extra users in. It is also why `config.validateAdditionalUsers` holds these usernames to a stricter rule on Kubernetes -- the kubelet silently DROPS variables whose names are not identifiers, so a `.` or `-` would yield a user with no password and nothing to say so. `broker.AdditionalUsers`, the retired broker-CLI op, is gone: this is the replacement its marker described.

The CLI route it used to take was retired rather than renamed. Verified against a live cluster: extra `username_<u>_password` keys in the credentials Secret are ignored by the operator, and `extraEnvVars`/`extraEnvVarsSecret` -- the only declarative alternative -- would expose the passwords in the pod environment. Driving the broker CLI avoided both, but the op could not show its output (the transcript repeats the passwords), had to delete its uploaded script via `defer`, and failed rather than reconciling an existing user, so it was never re-runnable. That is still the reason there is no run-everything configure step on either platform: a command that stopped halfway through on its second run would be worse than no command. `broker configure`'s Long text carries the order that works on a fresh broker instead. `broker perform assert-leader` remains cross-host + primary-only on containers, so it fails loud on a backup or monitor host.

### Data replication (`internal/broker`, `internal/k8s`, `internal/config`)

`broker configure data-replication` converges ONE broker to the env file's `replication:`
block; `broker perform data-replication` moves per-VPN roles across BOTH brokers of the pair.
Both are registered as `dr` -- an initialism of a hyphenated two-word leaf, the shape
`gather-diagnostics` -> `gd` already has, rather than a synonym for `replication`, which
`internal/abbrev` refuses. Replication is NOT redundancy: redundancy is the three nodes of
one HA group, replication is two HA groups in different places, and a deployment routinely
has both. [docs/operations.md](docs/operations.md) is the procedure and
[docs/configuration.md](docs/configuration.md) the block.

**The `replication:` block is BYTE-IDENTICAL at both sites, and that is the constraint the
whole schema is shaped around.** There is no `mate:` key and no "my role" -- both would have
to be reversed in the other site's file, which is the pairing an operator would have to keep
in step by hand. Three of the four things a site entry holds are what the OTHER side needs to
know about it (`virtualRouterName`, `endpoints`, `via`); only `routerNames` is who I am. A
broker reads its own router name, finds itself, and whichever entry is not itself is its
mate, so a failover is ONE edit to ONE `activeAt` line. `Replication.Locate` errors on zero
matches and on more than one -- there is deliberately no "assume the first", because a wrong
guess there points a switchover at the wrong broker.

`virtualRouterName` is the site's KEY (operator, 2026-09-13), mandatory and never derived. An
earlier draft defaulted it to `v:` + the first router name and carried an open question about
whether that always holds; making it explicit deleted the derivation and the question
together, and the file now states the exact string the CLI will be handed. `routerNames` is
separate and is a LIST because a site is an HA group: run against the backup node and
`show router-name` answers the backup's own name, which a single-name key would not match.
`via` is a one-of OBJECT rather than a discriminator string beside a settings block -- the key
present IS the mechanism, so a `via` that names one thing and configures another stops being
expressible instead of needing a rule to catch it.

**One format, two grammars, at the same broker version.** Software writes
`replication mate connect-via "host:port" ["ssl"]` per endpoint; an appliance writes
`replication mate connect-port <n> [compressed|ssl]` plus a single
`virtual-router-name "<v>" connect-via "<host>"` where a port is ILLEGAL. It is an endpoint
LIST, not a host x port cross product -- proven by the SEMP v1 reply schema, which models the
two as sibling structures under `<mate>` (an appliance `connect-via` plus a `connect-ports`
xs:all trio; a software `mate-list` capped at 6 = 3 transports x 2 addresses). So
`renderMateAppliance` REFUSES a second distinct host or two endpoints sharing a transport
rather than dropping either silently. The transport vocabulary is exactly three words --
`plain-text`, `compressed`, `ssl`; `encrypted` appears in the request XSD and is refused by
name (operator, 2026-09-13).

**`show replication` is NOT normalised across platforms**, and an earlier draft that claimed
it was cited that as the evidence one model serves both. The two real captures disprove it:
the report mirrors each platform's CLI grammar. The parsers are therefore split by
`BrokerType`, and reading with the wrong one FAILS LOUDLY rather than returning an empty
`MateConfig` -- `TestParseShowReplicationWrongPlatformFails` is what pins that, and
`TestRealApplianceCaptureMateRoundTrips` pins the round trip against the real capture.
Enablement and role are TWO facts, not one (`admin-state` and `config-state`), both
tri-state: `n/a` is what a VPN that does not replicate returns, and folding it into `standby`
would let a switch promote against a VPN it cannot see.

**`MateChannel` (`internal/broker/matechannel.go`) is the seam, and it is deliberately not
`Transport`.** `Transport` takes `(ctx, role)` and has no notion of a second site. One
`MateChannel` stands for a whole HA GROUP, because every replication setting is `HA: yes` in
the broker's own config-sync table -- so a write that reaches the site's primary reaches its
backup, and there is no per-node loop. Config-sync is ASSUMED operational and never checked
(operator, 2026-09-13): it is off by default and can be oper-down, but an HA group in service
has it running, and checking would be this tool second-guessing an invariant the operator
maintains. That fact is appliance-proven from the reference's per-command CONFIG-SYNC blocks
and VERIFIED ON LIVE SOFTWARE BROKERS by the operator, so the feature treats both platforms
identically.

The kubectl leg lives in `internal/k8s/matechannel.go` and NOT in `internal/broker`, because
`internal/k8s` already imports `internal/broker` and the reverse would be a cycle;
`broker.CLIRunner` is a func type for exactly that reason. It deletes each script it
uploads, on every path out including a failed run -- the LOCAL leg always did
(`Ops.removeCLI`) and this one did not, so every DR command over kubectl left a file
inside a broker belonging to the OTHER site. That is the one place this tool writes into
somebody else's cluster, so it is the one place a leak is least defensible, and the
comment justifying the upload says the script carries no secret TODAY. Cleanup is
best-effort and warns rather than failing a role change that already succeeded, which is
why the constructor takes the same raw line sink `Cluster`, `Manager` and `Ops` do. It addresses the site by the
SITE's own `via.kubernetes` block -- a DR mate is usually a different cluster --
and `TestNewMateChannelUsesTheSiteNotTheLocalDeployment` is what stops it quietly driving the
local broker while reporting on the remote one. That site's `command` goes through
`config.SiteCommand`, the same execution guard as `kubernetes.runtime` and always against the
KUBERNETES allowlist whatever this end runs on (`TestSiteCommandGuarded`). The SEMP leg
(`sempmate.go`) needs no broker-type discovery at all -- which sibling structure the reply
populates says which platform it is -- and its `<execute-result code="ok|fail">` is a real
verdict, where the CLI leg has to scan a transcript because `cli -Apes` ALWAYS EXITS 0. The CLI
leg shares `failKeywords` with the import driver rather than carrying a second list; the SEMP
leg needs none.

**`perform dr` is a PLAN, not a procedure** (`switchplan.go`), and that costs one struct now
so the parked third mechanism -- two self-contained bash scripts for a site this tool cannot
reach -- can be added later without rewriting the op. The executor builds the plan AFTER
reading both sites, so its phases carry only what must move; a renderer would build it from
the file alone. The phases are fixed: Demote, Confirm, Promote, Verify. The demote lands on
the site `activeAt` does NOT name, so a VPN is never demoted at the site it is about to be
promoted at; every demotion is confirmed by READ-BACK, and the state is re-read AGAIN
immediately before promoting, because a concurrent `configure dr` or an operator at a CLI can
move a role in the gap and by then the confirmation is stale
(`TestBuildSwitchPlanPromoteNeverPrecedesDemote`,
`TestExecuteSwitchPlanRefusesToPromoteAgainstStaleState`).

The guarantee is stated narrowly on purpose: **WITHIN ONE RUN, no promotion begins until
every demotion has been confirmed and re-read.** It is not a claim about the system --
nothing locks the brokers -- and "structurally unreachable" would be false. Failing between
demote and promote leaves the VPNs standby at BOTH sites: unavailable, NOT corrupted, which
is the intended direction to fail and is what the report must say, so nobody goes looking for
data loss. `SwitchPreflight` refuses before any write (`TestSwitchPreflightRunsBeforeAnyWrite`)
on an unreachable channel, on a site not running on its primary HA node (operator, 2026-09-13
-- not about where the config-synced write lands, but about not stacking a cross-site role
change on an unresolved local failover), and on brokers whose `show replication` does not name
the other site's `virtualRouterName`. A VPN already active at BOTH sites stops the run rather
than being repaired: demoting one side picks a winner, and which site keeps its spooled
messages is not this tool's decision.

**`configure dr` is deliberately outside all of that** and can create two actives on its own.
It is local-only -- never reads and never writes the mate -- so it works on every platform,
needs no `via:` anywhere, and cannot be blocked by a WAN outage. That is the operator's
accepted trade (2026-09-13), stated in its `Long` text and in docs/operations.md rather than
guarded against. It also takes the exact-`yes` gate, because shutting replication down on an
unlisted VPN and demoting a listed one both interrupt delivery.

**`configure dr`'s own apply is TWO PHASES in two separate broker calls, not one script**
(`internal/broker/replicationops.go`), because mate configuration -- the address lines and the
virtual-router-name -- can only change while EVERY VPN on the broker has replication disabled
(CONFIRMED by the operator 2026-09-14, carried as an assumption from 2026-09-13). It is NOT
in the CLI reference, which does state such preconditions for other services, so the
documentation would never have settled it; the only repo-side evidence was circumstantial,
the broker's own replayable dump ordering `! Configure Replication` ~32,000 lines ahead of the
first per-VPN `no shutdown`. Phase 1's outage window is therefore the true cost of a mate
change, not an artefact of this implementation -- which is what makes the "only when the mate
differs" check load-bearing rather than an optimisation. Phase 1 shuts
down every VPN this broker reports admin-ENABLED -- not `Replicating()`, which also requires a
resolved role and would leave an enabled-but-roleless VPN up to refuse the very mate lines this
phase exists to protect -- then the removals, then the new mate lines, and ONLY when the mate
actually differs (`SameEndpoints` plus the router name, never "removals is empty", which a
broker holding a subset of the wanted addresses would pass through unconverged forever). Phase
2 runs whenever it has anything to say: per listed VPN, `shutdown`, `state <role>`,
`no shutdown` -- never enable before the role is set, which is backwards from what this used to
emit -- then `shutdown` for every unlisted VPN this broker has replication ENABLED for. Both
phases key on admin state alone and deliberately on the SAME predicate: `Replicating()` also
demands a resolved role, and using it in one phase but not the other dropped an
enabled-but-roleless VPN between them, shut down by phase 1 and neither re-enabled nor recorded
by phase 2.

The re-enable keyword is read from the report's `Q` (queue state) column, not from enablement.
The broker's default is `fail-on-existing-queue`, refused when a data replication queue already
exists, and `force-use-existing-queue` is its inverse, refused when there is NOT one -- a
two-sided choice with no safe constant answer. Enablement is the wrong signal for it in BOTH
directions, which the real capture settles: `vpn-01  U S U - - - Y A` is admin-UP with the
queue n/a, so an enabled VPN need not have one, and phase 1 shuts VPNs down, so on the re-run
these errors tell an operator to make, a VPN that still holds its queue reads back admin-DOWN.
A report with no `Q` column yields `QueueUnknown`, which falls back to the broker's default
rather than asserting what the report did not say. `force-recreate-queue` is never emitted: it
discards whatever the queue holds. Still NEEDS VERIFICATION ON A LIVE BROKER: that the queue
survives phase 2's own `shutdown`, and what the column reports for a VPN whose role phase 2 is
about to flip -- on the one capture here the queue tracks the ROLE, so a promotion may need a
form the pre-state does not predict.
Neither phase rolls the other back: `RunCLI`'s `stop-on-error` stops the ONE call it wraps, so
a phase-1 rejection leaves whatever it shut down still down with phase 2 never sent, and a
phase-2 rejection after a successful phase 1 leaves the mate converged with only a prefix of
the listed VPNs re-enabled. The error names which phase stopped, names the VPNs left
with replication down, and states plainly that a phase-1 rejection may have stopped anywhere in
its own sequence -- so what is true depends on where, and only the broker can say. Re-running is
SAFE because the script is recomputed from a fresh read rather than replayed (CONVERGENT, not
idempotent: the queue-guard keyword is state-dependent), but converges further only when the
cause was transient.

A ROLE IS SET IN PLACE (operator, 2026-09-14): `message-vpn <n>` -> `replication` ->
`state <active|standby>`, against a VPN that is up and replicating, with no shutdown around
it. An earlier draft of phase 2 cycled every listed VPN down and back up to change its role,
on the assumption that a role could only move while replication was disabled. The assumption
was wrong and it cost a real outage: replication stopped on every listed VPN on every run,
including the runs where the mate had not changed and nothing needed to stop at all.

The evidence was already in the tree. `setReplicationRoleScript`, which `perform dr`'s
switchover calls, has always set `state` alone against a VPN its own preflight requires to be
ENABLED -- so the two paths contradicted each other and the switchover was the one telling
the truth. They now agree, and phase 1 is the ONLY thing that stops replication.

A site's `via` is validated ON DEMAND, not at load: a file with none is valid for `configure
dr` and is refused by `perform dr`'s own preflight. An EMPTY `via:` is accepted for a reason
that is not a preference -- a block with neither child decodes to the same zero value as an
absent key, so the two are indistinguishable and only the switch command can report a site it
cannot reach (`TestValidateReplicationEmptyViaAccepted`). `internal/convert` carries NOTHING
over: `REPL_MATE`/`REPL_CONN_SSL`/`REPL_PSK` describe one mate, the block describes a pair,
and all three are named in a warning rather than dropped silently.

Still open and NOT in scope until asked for: a read-only replication status command. Both
commands end with `show replication` + `show message-vpn * replication`, so the AFTER picture
is covered; what is missing is reading the pair without mutating it. `MateChannel.Status`
makes it nearly free. Also settled and not to be re-litigated:
`config-sync assert-leader message-vpn <n>` is NOT issued after a switchover
(operator-confirmed, 2026-09-13) -- it forcibly overwrites the other leader's content, so
adding it speculatively is worse than omitting it.

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

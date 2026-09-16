# CLAUDE.md

Guidance for Claude Code (claude.ai/code) working in this repository.

This file holds **cross-task conventions only** -- the rules that bind you before you touch
any part of the tree. Per-feature mechanics live in `docs/` and in the packages' own comments;
the map at the bottom says which is which. If a rule here and a doc disagree, the doc is
describing behaviour and this file is describing a constraint: fix the drift, do not pick one.

## What this repo is

A single Go binary, `solace-util`, that deploys and operates Solace PubSub+ Event Brokers on
Kubernetes (via the Solace EventBroker Operator and its PubSubPlusEventBroker CRD), Docker and
Podman. You describe the broker once in a YAML env file and drive the whole lifecycle from one
standardized command tree. **Unsupported** -- not a Solace product.

[README.md](README.md) is the quick start. [docs/configuration.md](docs/configuration.md) is the
env-file reference, [docs/operations.md](docs/operations.md) the day-2 lifecycle,
[docs/developer.md](docs/developer.md) the build and dev-script guide, and
[docs/test.md](docs/test.md) the catalogue of every test, fixture and injectable seam.

## Packages, and which way they may import

`internal/{abbrev,config,engine,render,broker,k8s,container,convert,cli,output,examples}` plus
`main.go`, and `internal/tools/vulnjudge` -- a dev-only command the `scan` task pipes
govulncheck's JSON through, so a fixable vulnerability fails the gate and one with no released
fix warns.

The layering is enforced by convention, not by a compiler, so it is stated here:

- **`abbrev`, `output` and `examples` import nothing but the stdlib.** That is what lets any
  package use them, and it is why `output` can define the `==> ` prefix once for both `engine`
  and `cli`.
- **`config` imports only `abbrev`.** It must never import `convert` (which is why
  `config.Load`'s invalid-YAML hint carries its own bash-file sniff) and never `render` (which
  is why `validate.go` keeps a small copy of the environment-variable folding `render` owns).
- **`k8s` and `container` import `broker`; the reverse would be a cycle.** `broker.CLIRunner`
  is a func type for exactly that reason, and it is how `internal/k8s/matechannel.go` supplies
  the kubectl leg to a `broker` seam.
- **`cli` imports everything and nothing imports `cli`.**

## Generated files -- never hand-edit one

[docs/commands.md](docs/commands.md), [docs/abbreviation.md](docs/abbreviation.md),
[docs/import.md](docs/import.md), `env/sample.yaml` and every `testdata/*.golden` are
**generated**. `test` fails while one is stale, so any change to a command, flag, `Short`,
abbreviation, import rule or the schema means regenerating in the same change with the `regen`
task. `regen` is deliberately outside `all`/`full`: a gate must not rewrite what it compares
against.

`regen` runs `./internal/examples` **first**. `//go:embed` cannot reach outside its own package,
so `internal/examples/assets/full.yaml` is the authoritative copy and `env/sample.yaml` is
generated from it -- and the render and k8s goldens are rendered from the sample, so the reverse
order would leave them a pass behind.

## The command tree

**Noun then verb, and nothing implicit** (`internal/cli/commands.go`). You name the thing, then
what to do to it: `broker deploy`, `operator remove`. `group()`'s `RunE` never dispatches to a
platform op -- run a noun bare and it prints what it can do, give it a word it does not own and
it fails loud rather than printing help and exiting 0. `TestNounGroupsRunNothing` pins it, and
that is what makes the short forms safe: `br` and `op` alone act on nothing.

**Help text is brief.** `group()` takes its `Long` as a parameter and the leaf builders take
only a `Short`, so a leaf gets one through `withLong(c, text)`; a command built from a
`cobra.Command` literal sets the field directly. Those are the only two shapes -- a
`c.Long = ...` assignment is drift. One to four lines saying what the command does and how to
invoke it; mechanics and history belong in [docs/operations.md](docs/operations.md). A
destructive command carries exactly one sentence naming what is destroyed.

**One word per platform, in two places.** A `config.Platform` constant's value is simultaneously
the env file's section key and the `--platform` value. The yaml tag is a separate literal, so
`TestPlatformConstantsMatchSchemaSections` is what stops the two drifting.

**There is no `--dry-run` and no `--gen-*` flags.** `generate <target>` is the only
look-before-you-leap surface, so an artifact you meant to read and a cluster you meant to change
are not one typo apart. `engine.Echo` is a TEST-ONLY runner installed through the
`App.NewRunner` seam; nothing user-facing installs it.

**Secret values are shown in exactly one place** -- `generate` on Kubernetes, where the Secret
manifest *is* the artifact. Everywhere else reports say `set`/`MISSING` and `engine.Echo` masks
child environment as `NAME=***`.

**The env-loading hook lives on each runnable command's `PreRunE`, never on root's
`PersistentPreRunE`.** Cobra runs the nearest ancestor's persistent hook and `__complete` is
root's own child, so a hook on root would parse an untrusted env file on every TAB press
(`TestCompletionNeverReadsTheEnvFile`). `--allow-command` is a non-persistent flag on each
command that executes, for the same reason.

**Applicability is said in help and enforced at pre-run, never expressed by hiding a command.**
There is ONE tree, the same shape on every platform, because help and completion render it
without an env file. For a command whose implementation differs, the `platformOps` map is the
single source of truth: its keys set the annotation help reads, its values are what runs.

**`examples` is the one command that must not load an env file**, since it emits one. It is a
bare `cobra.Command` literal rather than a `leaf`, and not calling `wireExec` is the whole
opt-out -- which is also why it offers no `--allow-command`. Both hand-maintained exemption
lists have to name it.

**Abbreviations: nothing rewrites argv.** Each site expands its own input at the moment it reads
it -- cobra's `Find` for a command, `config.ParseRole`, `config.ParsePlatform`, pflag for a
shorthand. `internal/abbrev` owns the RULES and each package declares its own words. One rule
CLI-wide: abbreviations work everywhere, completion teaches the canonical spelling.

**Cobra's Windows Explorer check stays off** (`cobra.MousetrapHelpText = ""` in `newRootCmd`).
Left on it declines to run when it thinks the binary was double-clicked, and it walks the OS
process table on every `Execute`. `TestMousetrapIsDisabled` pins it, because nothing else would
notice it coming back -- every test would still pass, just slower.

## Security invariants

**Config text must not be able to choose what executes.** `config.CheckCommand` is the whole
rule for `kubernetes.command`, `docker.command`, `podman.command`, `docker.compose` and each
`replication.sites[].via.kubernetes.command`, and it is enforced **twice from that one
definition**: in `Validate`, and again in every executor immediately before argv is built.
Executors are handed a `*config.Config` directly and **must not assume `config.Load` ran** --
if you add one, re-run the check there too.

**`--allow-command` is the only way to widen the allowlist**, and it has a floor: `neverAllowed`
(the escalators -- sudo, doas, pkexec, run0, nsenter and the rest) can be approved by nobody.
There is deliberately no schema key, no env var and no binding layer: an env file that could
approve its own binary would make the allowlist decorative.

**A retired key is retained and refused by name.** `.runtime` became `.command`; the old field
still parses so a file carrying it fails loud naming its replacement instead of hitting a bare
unknown-key error. Do the same for any future rename.

**Home expansion is one rule: a LEADING `~` resolves everywhere, an EMBEDDED one is an ordinary
character everywhere.** Both passes run inside `Load`, after `ApplyDefaults` and **before**
`Validate`, because a guard polices the value that will actually be used. A surviving leading
`~` reaching either guard means a field is missing from a pass' list.

**A custom `UnmarshalYAML` takes over decoding for the whole struct it is declared on, and
`dec.KnownFields(true)` never reaches inside it.** Any block that declares one must
re-implement that strictness by hand or it becomes open passthrough, where a typo decodes clean
and leaves the real field at zero. `internal/config/scaling.go` is the worked example.

**Secrets ride stdin, never argv, and are never echoed back.** Reports say `set`/`MISSING`.

## Output and prompts

**One leaf package owns every stdout and stderr shape** (`internal/output`), so a convention
cannot drift per call site. Two voices, split by stream: **stderr narrates, stdout reports.** A
stderr line is `==> <phase>` or `[TAG ] <msg>` and never both. stdout carries report bodies --
`Section`, `KVBlock`, `KVRow`, `Table`. Plain ASCII, no ANSI colour: these lines are read in
PowerShell 5.1, captured into dev-script logs, and grepped.

`Log` on `Cluster`/`Manager`/`Ops` is the **raw** line sink -- it emits a formatted line
verbatim and the callee's own Sink adds the prefixes, so the CLI hands it `lineSink()` and
deliberately not `step`.

**No driver package prompts.** Every confirmation lives in `internal/cli` and goes to stderr, so
redirecting stdout captures results rather than questions. `k8s` and `container` take a
`Confirm` func seam that declines when nil.

**Whether a human is present is decided by READING stdin, never by what kind of stream it is.**
The question goes to stderr, the answer is read from stdin, and an unanswered question is NOT a
"no": a reply of `n` is a decision, EOF means nobody could make one, and only the second earns
the refusal line naming `--no-prompt`. See [docs/operations.md](docs/operations.md) for the
operator-facing contract and `internal/cli/confirm.go` for why every other approach failed.

**Removals keep the expensive layer.** `broker remove` keeps persistent data, `operator remove`
keeps the CRDs. The `--delete-*` flag RAISES the question rather than silencing it, `--no-prompt`
is the one silencer, and there is no global `--yes`. The full truth table is in
[docs/operations.md](docs/operations.md#the-layer-flag-raises-the-question);
`TestConfirmLayerMatrix` pins it.

## Tests and gates

Every test is catalogued in [docs/test.md](docs/test.md) -- what it proves, the per-package
fixtures and doubles to reuse, and the injectable seams. **Update it in the same change** as the
test. Branch added means test added; branch removed means test removed or repurposed.

The coverage floor is the last recorded total. An unexplained drop is a failed gate, not a
rounding error; deleting dead code or a weak test can legitimately lower it, and that has to be
said in the change.

You do not run the gates. Hand over the invocation and continue from the output reported back.

## Where the mechanics live

| Feature | Operator-facing | Engineering rationale |
| --- | --- | --- |
| Env-file schema, command fields, scaling, replication block | [docs/configuration.md](docs/configuration.md) | `internal/config` |
| Lifecycle, removals, preflight, version floors | [docs/operations.md](docs/operations.md) | `internal/k8s`, `internal/container` |
| Config export and import | [docs/operations.md](docs/operations.md), [docs/import.md](docs/import.md) | `internal/broker/{blocks,annotate,sections,inject,transform,diff,importops,driver}.go` |
| Data replication and DR switchover | [docs/operations.md](docs/operations.md) | `internal/broker/{replication,replicationops,matechannel,switchplan,sempmate}.go` |
| Docker and Podman internals | [docs/operations.md](docs/operations.md) | `internal/container`, `internal/render` |
| Bash env-file migration | [docs/configuration.md](docs/configuration.md) | `internal/convert` |
| Build, dev-script tasks, releases | [docs/developer.md](docs/developer.md) | `scripts/dev.sh`, `scripts/dev.ps1` |

Those packages carry their reasoning in their own file-level comments, which is where a design
decision belongs once it is implemented. Record a new one there, not here.

## Knowledge graph

`graphify-out/` is tracked in git and holds a persistent graph of the repo:
[graph.html](graphify-out/graph.html) (interactive), [GRAPH_REPORT.md](graphify-out/GRAPH_REPORT.md)
(god nodes, cross-document links, suggested questions), `graph.json`, plus an Obsidian vault and
a per-community wiki.

- For a codebase question, run `graphify query "<question>"` first -- it returns a scoped
  subgraph, usually far smaller than grep output. `graphify path "<A>" "<B>"` answers
  relationships and `graphify explain "<concept>"` focuses one concept.
- `graphify-out/wiki/index.md` is for broad navigation; read `GRAPH_REPORT.md` only for
  architecture review or when the queries do not surface enough.
- The graph is refreshed by the `graphify` leg of the `full` task, so a green `full` leaves it
  current. It is local-only and skipped in CI.
- [.graphifyignore](.graphifyignore) scopes what graphify **reads**. Graphify does not read
  `.gitignore`, so the two files are maintained independently.

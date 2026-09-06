# Abbreviations

Every short form `solace-util` accepts, and the word it stands for.

**Generated from the abbreviation sets and the command tree -- do not edit by hand.**
Regenerate after adding or removing one:

```
go test ./internal/cli -update
```

The `test` task fails while this file is stale, so it cannot drift from the code.

## How a short form is resolved

Nothing rewrites the command line. There is no expansion pass in front of
execution: each word is expanded by whatever already owns that input.

| Kind | Expanded by | When |
| --- | --- | --- |
| Command name | cobra's own `Find`, from the aliases `applyAliases` attaches at construction | dispatch |
| Node role | `config.ParseRole` | while the command runs |
| `--platform` value | `config.ParsePlatform` | flag handling |
| Flag shorthand | pflag | flag parsing |

What the four share is `internal/abbrev`, which owns the rules rather than the
words: a short form is lowercase letters or digits, is strictly shorter than the
word it stands for, is claimed by one word only, and never shadows a canonical
name. A declaration that breaks one of those fails at construction, so it cannot
reach a release.

Three consequences worth knowing:

- **Only these exact words work.** Cobra's prefix matching is off, so `dep` is not
  `deploy`.
- **A word means the same thing at every depth.** `br` is `broker` under whichever
  verb it sits below, because the table is keyed by the command's name rather than
  by its path.
- **Completion teaches the canonical spelling.** The short forms work everywhere but
  are deliberately not offered by TAB completion -- a completion already saves the
  typing an abbreviation is for.

## Commands

| Command | Short form | Notes |
| --- | --- | --- |
| `check` | `ck` |  |
| `config` | `cfg` |  |
| `convert` | `cv` |  |
| `copy` | `cp` |  |
| `deploy` | `dp` |  |
| `diagnostics` | `diag` |  |
| `examples` | `eg` | the one noun at the top level: it emits an env file rather than acting on a deployment |
| `generate` | `gen` |  |
| `logs` | `lg` |  |
| `prepare` | `pre` |  |
| `remove` | `rm` | safe to hand out because `remove` takes a noun before it does anything: `rm` alone prints help |
| `restart` | `rs` |  |
| `shell` | `sh` |  |
| `start` | none | no short form: any two-letter form is ambiguous with `stop` and `status`, and that is the one slip that costs an outage |
| `status` | `sts` |  |
| `stop` | none | no short form, for the same reason as `start` |
| `version` | `ver` |  |
| `broker` | `br` | works under every verb that takes a broker |
| `operator` | `op` | works under every verb that takes the operator |

Verbs with no short form and no recorded reason: `auto-complete`, `cli`, `smoke`.

## Node roles

| Role | Short form | Notes |
| --- | --- | --- |
| `primary` | `p` | the default when the argument is omitted |
| `backup` | `b` |  |
| `monitor` | `m` |  |

The letter is not a second spelling: it is the role's internal form, the one that
appears in Kubernetes pod-name suffixes. Both spellings work anywhere a `[role]`
positional or `--pod` is taken.

## Platforms

| `--platform` | Short form | Notes |
| --- | --- | --- |
| `kubernetes` | `kube` | `k8s` and `k8` are refused: the section key is the product's own word |
| `docker` | `dk` |  |
| `podman` | `pm` |  |

The canonical word is the env file's own section key, so it is the only spelling
the file, the error messages and the completions use. A short form is accepted on
`--platform` alone and is refused in the env file.

## Flag shorthands

| Shorthand | Flag | Where | Meaning |
| --- | --- | --- | --- |
| `-e` | `--env` | every command | env file name, searched in the base dir then &lt;base-dir&gt;/env; a value with a directory is used as-is |
| `-i` | `--input` | `solace-util cli` | run this Solace CLI script instead of opening an interactive session |
| `-o` | `--out` | `solace-util convert` | write the YAML here instead of stdout |
| `-o` | `--out` | `solace-util examples` | write the env file here instead of stdout |
| `-v` | `--verbose` | every command | announce every external command as it runs; by default the binaries this env file names are resolved and listed once, up front |


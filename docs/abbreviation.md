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
| `broker` | `br` | the noun runs nothing on its own, so `br` alone cannot act |
| `operator` | `op` | same: `op` alone prints what the operator can be asked to do |
| `examples` | `eg` | emits an env file rather than acting on a deployment, so it sits at the top level |
| `convert` | `cv` |  |
| `version` | `ver` |  |
| `validate` | `vld` | works at the top level and under both nouns: the whole env file, or either half |
| `deploy` | `dp` |  |
| `remove` | `rm` | safe because the noun comes first: `rm` alone is not a command |
| `generate` | `gen` |  |
| `restart` | `rs` |  |
| `status` | `sts` |  |
| `logs` | `lg` |  |
| `shell` | `sh` |  |
| `copy` | `cp` |  |
| `configure` | `cfg` |  |
| `perform` | `pf` |  |
| `start` | none | no short form: any two-letter form is ambiguous with `stop` and `status`, and that is the one slip that costs an outage |
| `stop` | none | no short form, for the same reason as `start` |
| `cli` | none | no short form: three letters already, and `cl` would not read as anything |
| `gather-diagnostics` | `gd` |  |
| `data-replication` | `dr` |  |
| `server-certs` | none | no short form: completion types it, and `sc` reads as nothing |
| `domain-certs` | none | no short form, for the same reason as server-certs |
| `product-keys` | none | no short form, for the same reason as server-certs |
| `default-vpn` | none | no short form: `dv` would not distinguish it from default-users |
| `default-users` | none | no short form, for the same reason as default-vpn |
| `assert-leader` | none | no short form: rare, and run deliberately |
| `redundancy-test` | none | no short form: invasive, so spelling it out is a feature |
| `semp-login-check` | none | no short form: rare, and run deliberately |
| `cli-script` | none | no short form: `cs` would not distinguish it from shell-script |
| `shell-script` | none | no short form: `sh` is already claimed by `shell` |
| `export-config` | none | no short form: completion types it, and `ec` reads as nothing |
| `import-config` | none | no short form: it tears down and rebuilds a message-VPN, losing the messages spooled in its queues, so spelling it out is a feature -- the same reason redundancy-test has none |
| `from` | none | no short form: it is already one syllable and reads as a direction |
| `into` | none | no short form, for the same reason as from |

Verbs with no short form and no recorded reason: `auto-complete`.

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
| `-f` | `--follow` | `solace-util broker logs` | keep streaming new lines instead of exiting |
| `-f` | `--follow` | `solace-util operator logs` | keep streaming new lines instead of exiting |
| `-o` | `--out` | `solace-util broker generate` | write the artifact to this file instead of stdout |
| `-o` | `--out` | `solace-util broker perform export-config` | write the artifact to this file instead of stdout |
| `-o` | `--out` | `solace-util convert` | write the artifact to this file instead of stdout |
| `-o` | `--out` | `solace-util examples` | write the artifact to this file instead of stdout |
| `-o` | `--out` | `solace-util operator generate` | write the artifact to this file instead of stdout |
| `-p` | `--previous` | `solace-util broker logs` | read the PREVIOUS container's logs instead of the current one (kubernetes only) |
| `-v` | `--verbose` | every command | announce every external command as it runs; by default the binaries this env file names are resolved and listed once, up front |


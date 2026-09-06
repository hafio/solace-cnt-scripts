# solace-util

A single Go binary that deploys and operates Solace PubSub+ Event Brokers on
Kubernetes, Docker, or Podman. You describe the broker once in a YAML env file and drive
the whole lifecycle through one standardized command tree:

```
check deploy -> prepare all -> deploy all     build it
config ...                                    post-deployment, over the broker CLI
check semp-login / smoke redundancy           prove it works
stop broker / start broker                    pause it without removing it
remove all                                    tear it down
```

Every verb that owns more than one kind of object names the object it acts on --
`deploy broker`, `remove operator`, `status broker` -- and the bare verb prints what it
can act on rather than doing something implicit; `remove` on its own removes nothing.

**Platform status**

| Platform | State |
| --- | --- |
| Kubernetes (via the Solace EventBroker Operator) | Fully supported |
| Docker / Podman (host-local containers, no operator) | Fully supported |

> Unsupported -- this is not a Solace product. Use at your own risk.

## Documentation

| Document | What it covers |
| --- | --- |
| This file | Install, quick start, the commands you type most |
| [docs/commands.md](docs/commands.md) | Every command and every flag, generated from the command tree |
| [docs/abbreviation.md](docs/abbreviation.md) | Every short form -- commands, roles, `--platform`, flag shorthands -- and how each is resolved |
| [docs/configuration.md](docs/configuration.md) | The env file: keys, platform detection, scaling, secrets, the command allowlist, `convert` |
| [docs/operations.md](docs/operations.md) | Day-2: preflight, config order, HA mechanics, re-deploying, removal, upgrades, troubleshooting |
| [docs/developer.md](docs/developer.md) | Build, dev-script tasks, gates, goldens, releases, repo layout |
| [env/sample.yaml](env/sample.yaml) | The annotated schema -- `solace-util examples full`, committed for reference |

## Requirements

- **Kubernetes:** `kubectl` on your `PATH` and a reachable cluster/context. The binary
  shells out to `kubectl`; it does not embed a Kubernetes client. Set `kubernetes.runtime`
  to use `oc` instead, or to carry a whole profile such as
  `kubectl --kubeconfig /path/.kubeconfig-dev`.
- **Docker / Podman:** the `docker` or `podman` binary on your `PATH`, on the host that runs
  the broker. Podman deploys a systemd **quadlet** unit and needs systemd; Docker deploys
  through **compose**, so that host needs the compose plugin (`docker compose`) or the
  standalone `docker-compose` binary -- set `docker.compose` when it is the latter.
- **Building from source:** Go 1.26+.

Version-specific caveats (podman secret flags, the compose version the generated file
needs, wrapper runtimes such as `microk8s kubectl`) are in
[Troubleshooting](docs/operations.md#troubleshooting).

## Get the binary

Download a release binary for your platform, or build it:

```
go build -o solace-util .
```

`solace-util version` prints the stamped version (or `dev` for a plain `go build`), the Go
toolchain, and the OS/arch it was built for:

```
solace-util v1.2.3 go1.26.5 linux/amd64
```

Release builds and cross-compilation are in [docs/developer.md](docs/developer.md#build).

## Shell completion

`solace-util auto-complete <shell>` prints a completion script on stdout:

```
source <(solace-util auto-complete bash)                               # bash
source <(solace-util auto-complete zsh)                                # zsh
solace-util auto-complete fish | source                                # fish
solace-util auto-complete powershell | Out-String | Invoke-Expression  # PowerShell
```

For every session, write it where the shell looks -- `/etc/bash_completion.d/solace-util`,
`"${fpath[1]}/_solace-util"`, `~/.config/fish/completions/solace-util.fish`, or a `.ps1`
sourced from your `$PROFILE`.

Beyond command and flag names it completes the values they take: the env files `-e`/`--env`
would actually resolve, `primary`/`backup`/`monitor` for the `[role]` positionals and
`--pod`, directories for `--base-dir` and `--dir`, and the platform names for `--platform`.
Completion never reads your env file, so a TAB press cannot parse config, run a command, or
print anything into the shell. Add `--no-descriptions` to drop the help text shown beside
each suggestion.

## Quick start

Write yourself an env file to edit. `examples` (`eg`) carries the templates inside the
binary, so this works from a bare download:

```
solace-util examples                             # list what it can write
solace-util examples kubernetes -o env/dev.yaml  # or docker, or podman
```

Each of those is a minimal standalone file: only the keys that platform cannot default,
and only its own section, so nothing else has to be pruned and no `--platform` is needed.
Replace every `CHANGE-ME` before running anything against it.

`solace-util examples full` prints the complete annotated schema instead -- every key the
loader accepts, the default each omitted one takes, and all three platform sections at
once. It is the same text as [env/sample.yaml](env/sample.yaml), so a checkout can also
just `cp env/sample.yaml env/dev.yaml`; delete the platform sections you are not using,
because the CLI picks the platform from whichever of `kubernetes:`, `docker:` and
`podman:` your file declares. See
[Which platform runs](docs/configuration.md#which-platform-runs).

`-e` takes a file name, searched in the current directory then `./env`, so `-e dev.yaml`
finds `env/dev.yaml` here.

### Kubernetes

At minimum set `image.repo`, `image.tag`, `admin.pass`, `kubernetes.name`,
`kubernetes.namespace`, and `kubernetes.storage.msgNode`.

1. Render the broker manifest to see exactly what would be applied. This needs no cluster
   at all -- `generate` never contacts one and runs nothing:

   ```
   solace-util generate broker -e dev.yaml
   ```

2. A fresh cluster needs the EventBroker operator once. It is cluster-scoped and shared
   between brokers, so it is installed on its own rather than as part of bringing up any one
   broker -- and `remove all` never takes it away again:

   ```
   solace-util deploy operator -e dev.yaml
   ```

3. Check prerequisites, then bring the broker up. `deploy all` runs
   check -> prepare -> deploy -> assert the config-sync leader if HA:

   ```
   solace-util check deploy -e dev.yaml
   solace-util deploy all -e dev.yaml
   ```

   If you pin the HA roles to particular machines, run `solace-util prepare labels` once
   when provisioning the cluster. It is interactive -- only you can say which machine
   carries which role -- which is why it sits outside `prepare all` and `deploy all`.

4. Prove it works, and inspect:

   ```
   solace-util check semp-login -e dev.yaml
   solace-util status broker -e dev.yaml
   ```

5. Configure the running broker (TLS, hardening, product keys). These drive the Solace CLI
   inside a broker that is already up, so none of them are part of `deploy` -- run them in
   the order under
   [Post-deployment configuration order](docs/operations.md#post-deployment-configuration-order).

6. Tear it down. `remove all` keeps persistent data by default, asks before deleting it,
   and leaves the operator installed:

   ```
   solace-util remove all -e dev.yaml                   # asks about the PVCs; keeps them if you decline
   solace-util remove all -e dev.yaml --delete-data     # deletes them without asking (irreversible)
   ```

### Docker / Podman

One container per host, no operator. At minimum set `image.repo`, `image.tag`, `admin.pass`,
and a `docker: {}` or `podman: {}` section (write the empty section even when every setting
under it defaults -- it is what marks the file's platform).

1. Render the artifact this host would get -- a compose file on Docker, a systemd quadlet
   unit on Podman:

   ```
   solace-util generate broker -e prod.yaml
   ```

2. Bring it up. On a standalone broker that is one command per host; in an HA group each
   host names its own role, and `prepare host` generates the shared redundancy PSK:

   ```
   solace-util deploy all -e prod.yaml               # standalone
   solace-util deploy all primary -e prod.yaml      # HA: on the primary host
   solace-util deploy all backup  -e prod.yaml      # HA: on the backup host
   solace-util deploy all monitor -e prod.yaml      # HA: on the monitor host
   ```

3. In an HA group, assert the config-sync leader and exercise a real failover -- both from
   the **primary host only**, which drives the whole group:

   ```
   solace-util config leader -e prod.yaml
   solace-util smoke redundancy -e prod.yaml
   ```

4. Inspect, then tear down:

   ```
   solace-util status broker -e prod.yaml
   solace-util remove broker -e prod.yaml           # keeps the data directory by default
   ```

Re-running `deploy broker` is safe: it compares the rendered artifact with the one on disk
and only bounces a running broker with `--restart` or your consent. That is also how an
image-tag bump and a rotated password are applied -- see
[Docker and Podman mechanics](docs/operations.md#docker-and-podman-mechanics).

## Common commands

The full surface -- every command, argument, and flag with its default -- is
[docs/commands.md](docs/commands.md). These are the ones you type.

| Command | Platform | What it does |
| --- | --- | --- |
| `check deploy` | all | Validate config and platform prerequisites before deploying |
| `prepare all` | all | Namespace + secrets on Kubernetes; the data dir, DNS and PSK on docker/podman |
| `prepare labels` | kubernetes | Label nodes for primary/backup/monitor placement. Interactive, once per cluster |
| `deploy all [role]` | all | The whole bring-up: check -> prepare -> deploy -> leader (HA) |
| `deploy operator` | kubernetes | Install the cluster-scoped operator, once per cluster |
| `deploy broker [role]` | all | Deploy just the broker (`--restart` on docker/podman) |
| `config leader [role]` | all | Assert the config-sync leader (HA only) |
| `config apply server-cert` | all | Load or update the TLS server certificate |
| `config disable default-vpn` | all | Shut down the default message-VPN (one-way) |
| `status broker` | all | Pods/services/statefulset, or the local container. `--all`, `--detail`, `--pod` (kubernetes) |
| `logs broker` | all | Tail the broker's logs (`--pod` picks the pod on kubernetes) |
| `cli` | all | Open a Solace CLI in the broker (`-i <file>` runs a script instead; `--pod` picks the pod on kubernetes) |
| `shell` | all | Open a shell in the broker (`--pod` picks the pod on kubernetes) |
| `check semp-login` | all | Prove an authenticated SEMP request works (`--pod` picks the pod on kubernetes) |
| `smoke redundancy [role]` | all | Exercise a real failover and fail back (HA only; **disturbs the broker**) |
| `restart broker` | all | Bounce every pod, one at a time; `--pod <role>` restarts just one |
| `stop broker` / `start broker` | all | Pause and resume without removing anything |
| `diagnostics --days <n>` | all | Gather a support bundle into `broker.diagDir` |
| `generate broker [role]` | all | Print the artifact `deploy broker` would apply, without applying it |
| `remove all` | all | Remove the broker, its secrets and its namespace. Keeps the operator |
| `remove broker` | all | Remove just the broker. Keeps persistent data unless `--delete-data` |
| `examples [name]` | all | Write a starting env file (`-o <file>`). Bare, it lists them |
| `convert <bash-env-file>` | all | Turn a legacy bash env file into YAML |

Two rules apply everywhere:

- **No implicit actions.** A verb that owns more than one kind of object (`config`,
  `status`, `remove`, ...) never acts when run bare -- it prints what it can act on, and you
  name one. `remove` alone removes nothing.
- **Abbreviations.** `check`=`ck`, `config`=`cfg`, `convert`=`cv`, `copy`=`cp`,
  `deploy`=`dp`, `diagnostics`=`diag`, `examples`=`eg`, `generate`=`gen`, `logs`=`lg`,
  `prepare`=`pre`,
  `remove`=`rm`, `restart`=`rs`, `shell`=`sh`, `status`=`sts`, `version`=`ver`, and
  `broker`/`operator` are `br`/`op` under whichever verb takes them. So `dp op`,
  `sts br --all` and `rm all` all work. `start` and `stop` deliberately have none -- any
  short form is ambiguous between them and `status`.

A role is `primary`, `backup` or `monitor` (or `p`/`b`/`m`), and appears two ways. `[role]`
is a positional on `deploy broker`, `deploy all`, `generate broker`, `config leader` and
`smoke redundancy` -- on docker/podman it says which node identity *this* container host
deploys/renders/acts as (omit it in an HA group and the hostname is matched against
`nodes.*` to detect it -- a host matching none is an error on the deploying commands, while
`generate broker` renders the primary's artifact with a warning, since it changes nothing and
names the node in its own output); on Kubernetes it means nothing there and is refused loud, since the
cluster already knows every node. `--pod` is the only way to pick a Kubernetes **pod**:
`cli`, `shell`, `logs broker`, `check semp-login`, `status broker` and `restart broker` all
take it, and it is refused on docker/podman, where there is one container per host to target.
Passing a role where it means nothing is refused with a named error rather than ignored.

## Global flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `-e`, `--env <file>` | `env.yaml` | Env file to load: a file name searched in the base dir then `<base-dir>/env`, or a path used as-is |
| `--base-dir <dir>` | current dir | Directory searched for the env file, and holding `env/` |
| `--platform <name>` | -- | Platform to drive: `kubernetes` (`kube`), `docker` (`dk`) or `podman` (`pm`). Default: the one the env file declares, or a prompt if it declares several |
| `-v`, `--verbose` | `false` | Announce every external command as it runs (`==> exec: <resolved path> <args>`). By default the binaries this env file names are resolved and listed once, up front |

Every command that executes something also takes `--allow-command <name>`, which approves
one extra binary for this env file's platform command for that run only. It is rejected on
the `generate` tree, where nothing executes. See
[The command fields are executable content](docs/configuration.md#the-command-fields-are-executable-content).

There is no global `--yes` and no `--dry-run`. Confirmation is per-command: only the
commands that destroy something ask, and `--no-prompt` silences that one command's
questions -- see
[Removing a broker](docs/operations.md#removing-a-broker-what-stays-what-goes). To preview
an artifact instead of applying it, use `generate <target>`.

Before anything changes, every mutating command runs one cheap read-only probe -- see
[The preflight](docs/operations.md#the-preflight).

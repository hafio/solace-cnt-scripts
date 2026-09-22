# solace-util

A single Go binary that deploys and operates Solace PubSub+ Event Brokers on
Kubernetes, Docker, or Podman. You describe the broker once in a YAML env file and drive
the whole lifecycle through one standardized command tree:

```
validate                        check the env file and the platform
operator deploy                 install the cluster-scoped operator (kubernetes)
broker deploy                   prerequisites and the broker, in one idempotent step
broker configure ...            post-deployment settings, over the broker CLI
broker perform ...              one-shot actions: leader, failover test, diagnostics,
                                config export/import, DR role switch
broker stop / broker start      pause it without removing it
broker remove                   tear it down
```

The tree is **noun first**: you name the thing, then what to do to it. `broker` and
`operator` own their verbs and run nothing themselves, so a bare `broker` prints what it
can do rather than doing something implicit -- which is what makes `br` and `op` safe to
type. Prerequisites are not a separate step: `broker deploy` creates whatever the broker
needs and is safe to re-run.

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
| [docs/configuration.md](docs/configuration.md) | The env file: keys, platform detection, scaling, secrets, replication, the command allowlist, `convert` |
| [docs/operations.md](docs/operations.md) | Day-2: preflight, config order, HA mechanics, re-deploying, removal, data replication, upgrades, troubleshooting |
| [docs/developer.md](docs/developer.md) | Build, dev-script tasks, gates, goldens, releases, repo layout |
| [docs/container-security.md](docs/container-security.md) | The container security rules every platform must satisfy, and how each is expressed -- written to be reused in other projects |
| [env/sample.yaml](env/sample.yaml) | The annotated schema -- what a bare `solace-util examples` prints, committed for reference |

## Requirements

- **Kubernetes:** `kubectl` on your `PATH` and a reachable cluster/context. The binary
  shells out to `kubectl`; it does not embed a Kubernetes client. Set `kubernetes.command`
  to use `oc` instead, or to carry a whole profile such as
  `kubectl --kubeconfig /path/.kubeconfig-dev`.
- **Docker / Podman:** the `docker` or `podman` binary on your `PATH`, on the host that runs
  the broker. Podman deploys a systemd **quadlet** unit and needs systemd; Docker deploys
  through **compose**, so that host needs the compose plugin (`docker compose`) or the
  standalone `docker-compose` binary -- set `docker.compose` when it is the latter.
- **Building from source:** Go 1.27+.

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

To load it for every new shell, write it where that shell looks:

```
solace-util auto-complete bash > /etc/bash_completion.d/solace-util      # bash
solace-util auto-complete zsh > "${fpath[1]}/_solace-util"               # zsh
solace-util auto-complete fish > ~/.config/fish/completions/solace-util.fish
```

`/etc/bash_completion.d/` is read by the `bash-completion` package, not by bash itself. On
a host without that package there is no directory to write to, so source the script from
your `~/.bashrc` instead -- it needs nothing but bash:

```
echo 'source <(solace-util auto-complete bash)' >> ~/.bashrc
```

PowerShell has no such directory, so write the script once and dot-source it from your
profile -- generating it once is what keeps shell start-up fast, since the alternative
runs this binary on every new shell:

```powershell
solace-util auto-complete powershell > $HOME\solace-util.ps1
Add-Content $PROFILE '. $HOME\solace-util.ps1'
```

If `$PROFILE` does not exist yet, `New-Item -ItemType File -Force $PROFILE` creates it.

`solace-util auto-complete` on its own prints all of this, and each shell's own `--help`
adds whatever that shell needs first (zsh wants `compinit` enabled; bash needs nothing).

Beyond command and flag names it completes the values they take: the env files `-e`/`--env`
would actually resolve, `primary`/`backup`/`monitor` for `--pod`, directories for
`--base-dir` and `--dir`, and the platform names for `--platform`.
Completion never reads your env file, so a TAB press cannot parse config, run a command, or
print anything into the shell. Add `--no-descriptions` to drop the help text shown beside
each suggestion.

## Quick start

Write yourself an env file to edit. `examples` (`eg`) carries the templates inside the
binary, so this works from a bare download:

```
solace-util examples --platform kubernetes -o env/dev.yaml   # or docker, or podman
```

Each of those is a minimal standalone file: only the keys that platform cannot default,
and only its own section, so nothing else has to be pruned and no `--platform` is needed
when you later run against it. Replace every `CHANGE-ME` before you do.

Bare `solace-util examples` prints the complete annotated schema instead -- every key the
loader accepts, the default each omitted one takes, and all three platform sections at
once. It is the same text as [env/sample.yaml](env/sample.yaml), so a checkout can also
just `cp env/sample.yaml env/dev.yaml`; delete the platform sections you are not using,
because the CLI picks the platform from whichever of `kubernetes:`, `docker:` and
`podman:` your file declares. See
[Which platform runs](docs/configuration.md#which-platform-runs).

`-e` takes a file name, searched in the current directory then `./env`, so `-e dev.yaml`
finds `env/dev.yaml` here.

### Kubernetes

At minimum set `image.repo`, `image.tag`, `semp.adminPass`, `kubernetes.name`,
`kubernetes.namespace`, and `kubernetes.storage.msgNodeSize`.

1. Render the broker manifest to see exactly what would be applied. This needs no cluster
   at all -- `generate` never contacts one and runs nothing:

   ```
   solace-util broker generate -e dev.yaml
   ```

2. A fresh cluster needs the EventBroker operator once. It is cluster-scoped and shared
   between brokers, so it is installed on its own rather than as part of bringing up any one
   broker -- and `broker remove` never takes it away again:

   ```
   solace-util operator deploy -e dev.yaml
   ```

   Run it again from a second env file and the operator's watch list GROWS to cover both
   namespaces rather than being replaced, so one operator can serve several brokers.

3. Check the env file and the cluster, then bring the broker up. `broker deploy` creates
   the namespace and the secrets as well, and is safe to re-run:

   ```
   solace-util validate -e dev.yaml
   solace-util broker deploy -e dev.yaml
   ```

   If you pin the HA roles to particular machines, label those nodes yourself and name the
   labels under `kubernetes.placement`. This tool never labels your nodes -- it only writes
   the selectors into the broker resource.

4. Prove it works, and inspect. On HA, assert the config-sync leader once the pods are up:

   ```
   solace-util broker perform semp-login-check -e dev.yaml
   solace-util broker status -e dev.yaml
   solace-util broker perform assert-leader -e dev.yaml     # HA only
   ```

5. Configure the running broker (TLS, hardening, product keys). These drive the Solace CLI
   inside a broker that is already up, so none of them are part of `deploy` -- run them in
   the order under
   [Post-deployment configuration order](docs/operations.md#post-deployment-configuration-order).

6. Tear it down. `broker remove` keeps persistent data unless you ask for it, prompts for
   every step, and leaves the operator installed:

   ```
   solace-util broker remove -e dev.yaml                   # PVCs kept; nothing is asked about them
   solace-util broker remove -e dev.yaml --delete-data     # asks before deleting them (irreversible)
   ```

   The namespace is only offered for deletion once nothing else is left in it. If anything
   you did not deploy is still there, it is listed and the namespace is kept -- on every
   path, `--no-prompt` included.

### Docker / Podman

One container per host, no operator. At minimum set `image.repo`, `image.tag`, `semp.adminPass`,
and a `docker: {}` or `podman: {}` section (write the empty section even when every setting
under it defaults -- it is what marks the file's platform).

1. Render the artifact this host would get -- a compose file on Docker, a systemd quadlet
   unit on Podman:

   ```
   solace-util broker generate -e prod.yaml
   ```

2. Bring it up. `broker deploy` prepares the host as well -- the data directory, DNS and
   the registry login. It does NOT invent the shared redundancy PSK: generate one yourself
   with `openssl rand -base64 32` and put the SAME value in the env file on all three
   hosts. Nothing here edits the file you hand it.

   In an HA group each host says which node it is, or leave `--pod` off and the host is
   matched against your `redundancy.*` entries -- first by hostname against `name`, then
   by this machine's own addresses against `addr`. A standalone broker needs no node
   entry at all: it is named after the host it runs on unless
   `redundancy.primary.name` says otherwise.

   ```
   solace-util broker deploy -e prod.yaml                    # standalone, or role by hostname
   solace-util broker deploy --pod primary -e prod.yaml      # HA: on the primary host
   solace-util broker deploy --pod backup  -e prod.yaml      # HA: on the backup host
   solace-util broker deploy --pod monitor -e prod.yaml      # HA: on the monitor host
   ```

3. In an HA group, assert the config-sync leader and exercise a real failover -- both from
   the **primary host only**, which drives the whole group:

   ```
   solace-util broker perform assert-leader -e prod.yaml
   solace-util broker perform redundancy-test -e prod.yaml
   ```

4. Inspect, then tear down:

   ```
   solace-util broker status -e prod.yaml
   solace-util broker remove -e prod.yaml           # keeps the data directory by default
   ```

Re-running `broker deploy` is safe: it compares the rendered artifact with the one on disk
and only bounces a running broker with `--restart` or your consent. That is also how an
image-tag bump and a rotated password are applied -- see
[Docker and Podman mechanics](docs/operations.md#docker-and-podman-mechanics).

## Common commands

The full surface -- every command, argument, and flag with its default -- is
[docs/commands.md](docs/commands.md). These are the ones you type.

| Command | Platform | What it does |
| --- | --- | --- |
| `validate` | all | Check the whole env file and the platform. `broker validate` / `operator validate` are the scoped halves |
| `operator deploy` | kubernetes | Install the cluster-scoped operator. Re-run from another env file and its watch list grows |
| `operator remove` | kubernetes | Remove it, or narrow its watch list if other namespaces still need it |
| `broker deploy` | all | Prerequisites and the broker, idempotent (`--restart` on docker/podman, `--pod` names this host) |
| `broker remove` | all | Remove the broker. Data kept unless `--delete-data`; the namespace only if nothing else is in it |
| `broker generate` | all | Print what `broker deploy` would apply, without applying it |
| `broker status` | all | Pods/services/statefulset, or the local container. `--all`, `--detail`, `--pod` |
| `broker logs` | all | Read the broker's logs. `--follow`, `--tail`, `--since`, `--pod` |
| `broker cli` / `broker shell` | all | Open an interactive Solace CLI or shell (`--pod` picks the pod) |
| `broker copy from` / `into` | all | Copy files out of or into the broker |
| `broker restart` | all | Bounce every pod, one at a time; `--pod <role>` restarts just one |
| `broker stop` / `broker start` | all | Pause and resume without removing anything |
| `broker configure server-certs` | all | Load or update the TLS server certificate |
| `broker configure domain-certs` | all | Load the domain CAs (`--remove` deletes them) |
| `broker configure product-keys` | all | Apply the configured product keys |
| `broker configure default-vpn` | all | Shut the default message-VPN down (`--enable` brings it back) |
| `broker configure default-users` | all | Shut the default client-usernames down (`--enable` brings them back) |
| `broker configure data-replication` | all | Converge this broker to the `replication:` block -- mate addresses, which VPNs replicate, each one's role. Never contacts the mate |
| `broker perform assert-leader` | all | Assert the config-sync leader (HA only) |
| `broker perform redundancy-test` | all | Exercise a real failover and fail back (HA only; **disturbs the broker**) |
| `broker perform gather-diagnostics` | all | Gather a support bundle into `broker.hostDiagnosticDir` (`--days`, `--pod` narrows to one node) |
| `broker perform semp-login-check` | all | Prove an authenticated SEMP request works |
| `broker perform export-config` | all | Capture the broker's configuration as one artifact (`--vpn`, `--broker-only`, `-o`) |
| `broker perform import-config <file>` | all | Apply a captured configuration back; **tears down and rebuilds an existing VPN** |
| `broker perform cli-script <file>` | all | Run a Solace CLI script in the broker |
| `broker perform shell-script <file>` | all | Run a host shell script inside the broker |
| `broker perform data-replication` | all | Move each message-VPN's replication role across the DR pair; **interrupts message flow for every VPN it moves** |
| `examples` | all | Write a starting env file (`--platform <p>`, `-o <file>`). Bare, it writes the full schema |
| `convert <bash-env-file>` | all | Turn a legacy bash env file into YAML |

Two rules apply everywhere:

- **No implicit actions.** `broker` and `operator` own their verbs and run nothing
  themselves: run one bare and it prints what it can do. That is what makes the short
  forms safe -- `br` and `op` on their own act on nothing.
- **Abbreviations.** `broker`=`br`, `operator`=`op`, `configure`=`cfg`, `convert`=`cv`,
  `copy`=`cp`, `data-replication`=`dr`, `deploy`=`dp`, `examples`=`eg`,
  `gather-diagnostics`=`gd`, `generate`=`gen`, `logs`=`lg`, `perform`=`pf`, `remove`=`rm`,
  `restart`=`rs`, `shell`=`sh`, `status`=`sts`,
  `validate`=`vld`, `version`=`ver`. A word means the same thing at any depth, so `br dp`,
  `op rm`, `br sts --all` and `br pf gd` all work. `start` and `stop` deliberately have
  none -- any short form is ambiguous between them and `status`, and that is the one slip
  that costs an outage.

The role is always `--pod`, never a positional, and it reads two ways depending on the
platform. On Kubernetes it picks the **pod** to act on and defaults to the primary:
`broker cli`, `shell`, `logs`, `status`, `restart`, `copy`, the `configure` commands and
most of `perform` take it. On docker and podman it says which node **this host** is, on
`broker deploy` and `broker generate` -- omit it in an HA group and the host is matched
against your `redundancy.*` entries, by hostname first and then by address (a host matching
none is an error on deploy, while `generate` renders the primary's artifact with a warning,
since it changes nothing). Passed, it is obeyed but still checked: disagreeing with what
the host looks like warns and proceeds rather than prompting.

Each command scopes the flag to the platform where it means something, so `--pod` on a
command that cannot use it is refused by name rather than quietly ignored.

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

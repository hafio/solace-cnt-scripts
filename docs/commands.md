# Command reference

Every command `solace-util` exposes, with its arguments and flags.

**Generated from the command tree -- do not edit by hand.** Regenerate after any
command, flag, or description change:

```
go test ./internal/cli -update
```

The `test` task fails while this file is stale, so it cannot drift from the code.

## Tree

```
solace-util
  auto-complete
    bash
    fish
    powershell
    zsh
  check
    deploy
    semp-login
  cli
  config
    apply
      additional-users
      domain-certs
      product-keys
      server-cert
    delete
      domain-certs
    disable
      default-users
      default-vpn
    leader [role]
  convert <bash-env-file>
  copy
    from files...
    into files...
  deploy
    all [role]
    broker [role]
    operator
  diagnostics
  examples [name]
  generate
    broker [role]
    operator
    secrets
      broker
      operator
  logs
    broker
    operator
  prepare
    all
    host
    labels
    namespace
    secrets
  remove
    all
    broker
    namespace
    operator
    secrets
  restart
    broker
    operator
  shell
  smoke
    redundancy [role]
  start
    broker
  status
    broker
    operator
  stop
    broker
  version
```

## Index

- [auto-complete](#solace-util-auto-complete)
  - [auto-complete bash](#solace-util-auto-complete-bash)
  - [auto-complete fish](#solace-util-auto-complete-fish)
  - [auto-complete powershell](#solace-util-auto-complete-powershell)
  - [auto-complete zsh](#solace-util-auto-complete-zsh)
- [check](#solace-util-check)
  - [check deploy](#solace-util-check-deploy)
  - [check semp-login](#solace-util-check-semp-login)
- [cli](#solace-util-cli)
- [config](#solace-util-config)
  - [config apply](#solace-util-config-apply)
    - [config apply additional-users](#solace-util-config-apply-additional-users)
    - [config apply domain-certs](#solace-util-config-apply-domain-certs)
    - [config apply product-keys](#solace-util-config-apply-product-keys)
    - [config apply server-cert](#solace-util-config-apply-server-cert)
  - [config delete](#solace-util-config-delete)
    - [config delete domain-certs](#solace-util-config-delete-domain-certs)
  - [config disable](#solace-util-config-disable)
    - [config disable default-users](#solace-util-config-disable-default-users)
    - [config disable default-vpn](#solace-util-config-disable-default-vpn)
  - [config leader](#solace-util-config-leader)
- [convert](#solace-util-convert)
- [copy](#solace-util-copy)
  - [copy from](#solace-util-copy-from)
  - [copy into](#solace-util-copy-into)
- [deploy](#solace-util-deploy)
  - [deploy all](#solace-util-deploy-all)
  - [deploy broker](#solace-util-deploy-broker)
  - [deploy operator](#solace-util-deploy-operator)
- [diagnostics](#solace-util-diagnostics)
- [examples](#solace-util-examples)
- [generate](#solace-util-generate)
  - [generate broker](#solace-util-generate-broker)
  - [generate operator](#solace-util-generate-operator)
  - [generate secrets](#solace-util-generate-secrets)
    - [generate secrets broker](#solace-util-generate-secrets-broker)
    - [generate secrets operator](#solace-util-generate-secrets-operator)
- [logs](#solace-util-logs)
  - [logs broker](#solace-util-logs-broker)
  - [logs operator](#solace-util-logs-operator)
- [prepare](#solace-util-prepare)
  - [prepare all](#solace-util-prepare-all)
  - [prepare host](#solace-util-prepare-host)
  - [prepare labels](#solace-util-prepare-labels)
  - [prepare namespace](#solace-util-prepare-namespace)
  - [prepare secrets](#solace-util-prepare-secrets)
- [remove](#solace-util-remove)
  - [remove all](#solace-util-remove-all)
  - [remove broker](#solace-util-remove-broker)
  - [remove namespace](#solace-util-remove-namespace)
  - [remove operator](#solace-util-remove-operator)
  - [remove secrets](#solace-util-remove-secrets)
- [restart](#solace-util-restart)
  - [restart broker](#solace-util-restart-broker)
  - [restart operator](#solace-util-restart-operator)
- [shell](#solace-util-shell)
- [smoke](#solace-util-smoke)
  - [smoke redundancy](#solace-util-smoke-redundancy)
- [start](#solace-util-start)
  - [start broker](#solace-util-start-broker)
- [status](#solace-util-status)
  - [status broker](#solace-util-status-broker)
  - [status operator](#solace-util-status-operator)
- [stop](#solace-util-stop)
  - [stop broker](#solace-util-stop-broker)
- [version](#solace-util-version)

## Reading this reference

- **Global flags are inherited by every command** and are listed once, under
  [Global flags](#global-flags), rather than repeated in each command's flag table.
- **`--allow-command` approves one extra binary** for the env file's platform command,
  for that run only; it is repeatable and takes a bare name, never a path. It is listed
  on every command that executes something, and refused with a named error by the ones
  that only render.
- **A verb that owns more than one kind of object never acts when run bare.** Its entry
  below says "Runs nothing on its own" and names the objects instead; you pick one to
  make it do anything. `remove` alone removes nothing.
- **`[role]` is `primary`, `backup` or `monitor`**, and the letters `p`, `b`, `m` work
  everywhere the long names do. On Kubernetes it picks which pod a command targets,
  defaulting to the primary. On docker and podman there is one container per host, so it
  instead names which host in the redundancy group this invocation runs on -- required
  where the artifact is per-host, and detected from the host name where it may be
  omitted. Passing one where it means nothing is refused with a named error.
- **Abbreviations mean the same thing at every depth.** A command's short form is on its
  "Also available as:" line, and it works under whichever verb the command sits below --
  `br` and `op` ride under every verb that takes `broker` and `operator`. Every short
  form the tool accepts, including the role letters and the `--platform` spellings, is
  in [abbreviation.md](abbreviation.md).
- **"Applies to:" is enforced, not advisory.** A command scoped to some platforms refuses
  on the others with a named error rather than silently doing nothing; a command with no
  "Applies to:" line works on all of them. The tree is the same shape everywhere, because
  help and completion render it without reading an env file.

## Global flags

Inherited by every command.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--base-dir` | (none) | directory searched for the env file, and holding env/ (default: current directory) |
| `-e`, `--env` | `env.yaml` | env file name, searched in the base dir then &lt;base-dir&gt;/env; a value with a directory is used as-is |
| `--platform` | (none) | platform to drive: kubernetes (kube), docker (dk) or podman (pm). Default: the one the env file declares, or a prompt if it declares several |
| `-v`, `--verbose` | `false` | announce every external command as it runs; by default the binaries this env file names are resolved and listed once, up front |

## Commands

### solace-util

Deploy and operate Solace PubSub+ brokers on Kubernetes, Docker, or Podman

solace-util is a single CLI for deploying and operating Solace PubSub+ Event Brokers.
It presents the same lifecycle verbs on every platform, and every verb names
what it acts on -- run a verb on its own to see what it can act on:

  check deploy -> prepare all -> deploy all     build it
  config ...                                    POST-DEPLOYMENT, over the broker CLI
  check semp-login / smoke redundancy           prove it works
  stop broker / start broker                    pause it without removing it
  remove all                                    tear it down

The operator is cluster-scoped and shared, so it is installed and removed on
its own: `deploy operator`, `remove operator`.

`generate` renders any artifact to stdout without applying it -- that is how
you see what a command would send before you send it.

Every command takes -e/--env <file>, searched in the current directory then
./env. The platform comes from that file: whichever of kubernetes:, docker:
or podman: it declares is the one driven. A file declaring more than one asks
which to use, and --platform kubernetes|docker|podman (kube|dk|pm) answers that
up front. A few commands apply to only one platform; their help says so.

No env file yet? `examples <platform>` writes one to start from, and
`examples full` prints the whole annotated schema.

Coming from the bash scripts? 'solace-util convert <bash-env-file>' turns an old
env file into the YAML this reads.

```
solace-util
```

Subcommands: `auto-complete`, `check`, `cli`, `config`, `convert`, `copy`, `deploy`, `diagnostics`, `examples`, `generate`, `logs`, `prepare`, `remove`, `restart`, `shell`, `smoke`, `start`, `status`, `stop`, `version`


### solace-util auto-complete

Print the shell auto-completion script for solace-util

Print a shell's completion script on stdout. Load it to complete commands and
flags, plus the values they take: env files for -e/--env, primary|backup|monitor
for the [role] positionals and --pod, and directories for --base-dir and --dir.

Completion never reads the env file, so it stays inert -- a TAB press cannot
parse config or run anything. See each shell's help for how to load it.

```
solace-util auto-complete
```

Subcommands: `bash`, `fish`, `powershell`, `zsh`


### solace-util auto-complete bash

Print the bash completion script

Load into the current shell:

  source <(solace-util auto-complete bash)

Load for every session (needs the bash-completion package):

  solace-util auto-complete bash > /etc/bash_completion.d/solace-util

```
solace-util auto-complete bash [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--no-descriptions` | `false` | omit the descriptions shown beside each completion |


### solace-util auto-complete fish

Print the fish completion script

Load into the current shell:

  solace-util auto-complete fish | source

Load for every session:

  solace-util auto-complete fish > ~/.config/fish/completions/solace-util.fish

```
solace-util auto-complete fish [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--no-descriptions` | `false` | omit the descriptions shown beside each completion |


### solace-util auto-complete powershell

Print the powershell completion script

Load into the current shell:

  solace-util auto-complete powershell | Out-String | Invoke-Expression

Load for every session, by writing the script once and sourcing it from
your profile:

  solace-util auto-complete powershell > solace-util.ps1

```
solace-util auto-complete powershell [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--no-descriptions` | `false` | omit the descriptions shown beside each completion |


### solace-util auto-complete zsh

Print the zsh completion script

Load into the current shell:

  source <(solace-util auto-complete zsh)

Load for every session (compinit must be enabled in ~/.zshrc):

  solace-util auto-complete zsh > "${fpath[1]}/_solace-util"

```
solace-util auto-complete zsh [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--no-descriptions` | `false` | omit the descriptions shown beside each completion |


### solace-util check

Run read-only checks

Every check here is read-only: it reports and changes nothing.

  check deploy      before deploying -- config, cluster/engine reachability,
                    storage or DNS, and whether the operator is installed
  check semp-login  after deploying -- the broker answers an authenticated
                    SEMP request

The failover exercise is deliberately not here: it moves live traffic, so it
lives under `smoke` with the other invasive checks.

```
solace-util check
```

Runs nothing on its own. It names what `solace-util check` can act on -- `deploy`, `semp-login` -- and given a word it does not know it fails rather than reporting success.

Also available as: ck


### solace-util check deploy

Validate config and platform prerequisites before deploying

On Kubernetes this probes cluster reachability first; an unreachable
cluster stops the check there, reporting permission, the operator, and
the StorageClass as skipped rather than run. Once the cluster answers,
permission to create the broker resource, whether the operator is
installed, and the StorageClass are each checked regardless of whether
another one failed -- a missing operator only warns, since `deploy
operator` installs it.

On docker and podman it probes the runtime (docker also checks
compose), then the broker hostname(s) -- failing in HA on a miss but
only noting it standalone; an unreachable runtime stops the check
before DNS is tried.

```
solace-util check deploy [flags]
```

Also available as: dp

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util check semp-login

Test an authenticated SEMP request against a running broker

You run curl against http://localhost:8080/SEMP/v2/monitor from inside
the broker itself -- kubectl exec on Kubernetes, docker/podman exec on a
container host -- with the credential on stdin, never in argv or a log.

Kubernetes always logs in as the operator's fixed `admin` user, whatever
--pod you pick; docker and podman log in as whatever `admin.user` the
env file configures.

A failed login reports the response's failing HTTP status line (e.g.
`401 Unauthorized`), and the command still exits non-zero.

```
solace-util check semp-login [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--pod` | (none) | pod role to target (p\|b\|m) (kubernetes only) |


### solace-util cli

Open an interactive Solace CLI in the broker (Kubernetes: --pod picks the pod)

With no flags this opens an interactive Solace CLI session.

--input runs a script through that CLI instead of opening a session: a bare
filename is resolved under broker.cliScriptsFolder, a path is used as typed,
and the file is uploaded to the broker and run there. A CLI script is a
sequence of independent commands, so every line still runs even when the
broker rejects one -- but the command now fails at the end when any line
was rejected, instead of only warning.

```
solace-util cli [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `-i`, `--input` | (none) | run this Solace CLI script instead of opening an interactive session |
| `--pod` | (none) | pod role to target (p\|b\|m) (kubernetes only) |


### solace-util config

Configure a DEPLOYED broker (certs, hardening, product keys)

Post-deployment configuration: every step here talks to a broker that is already
deployed and running, over the Solace CLI. None of it is part of `deploy`.

There is no run-everything command, because these steps are not uniformly
re-runnable. The order that works on a fresh broker is:

  1. config leader                        (HA only; on containers, the primary)
  2. config apply server-cert             (when TLS is configured)
  3. config apply domain-certs            (when any are listed)
  4. config disable default-vpn
  5. config disable default-users
  6. config apply additional-users        (Kubernetes; after the hardening, so
                                           the sequence reads harden-then-provision.
                                           NOT re-runnable: the broker refuses to
                                           create a user that already exists)
  7. config apply product-keys            (when any are listed)

Only domain-certs can be undone from here (`config delete domain-certs`).
There is no un-harden, and no way to withdraw a server certificate or a
product key through this tool.

```
solace-util config
```

Runs nothing on its own. It names what `solace-util config` can act on -- `apply`, `delete`, `disable`, `leader` -- and given a word it does not know it fails rather than reporting success.

Also available as: cfg


### solace-util config apply

Apply configuration to the running broker

`domain-certs` and `product-keys` run identically on every platform;
`additional-users` is Kubernetes only, because docker and podman create
those CLI users at container boot from the mounted password file
instead of over the CLI here. `server-cert` does too, unless
`kubernetes.tlsServerSecret` names a Secret, in which case Kubernetes
rewrites that Secret directly with kubectl instead of running the
broker-CLI path containers always use.

The certificate bodies, product keys and passwords each of these
applies ride in on an upload or over stdin -- never as an argument on
this command's own command line.

```
solace-util config apply
```

Runs nothing on its own. It names what `solace-util config apply` can act on -- `additional-users`, `domain-certs`, `product-keys`, `server-cert` -- and given a word it does not know it fails rather than reporting success.


### solace-util config apply additional-users

Create the admin.additionalUsers CLI users (not re-runnable) (kubernetes only)

Uploads a generated CLI script and runs it on the primary node only:
management users replicate to the mates via config-sync, so one run is
enough for the whole deployment -- every node in the group when HA is
enabled, or the one primary on a standalone broker.

A username that already exists makes the whole run fail rather than
reconcile, and the transcript is withheld either way since it echoes
the passwords. Delete the user on the broker, or drop it from the env
file, then re-run this command.

```
solace-util config apply additional-users [flags]
```

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util config apply domain-certs

Load the configured domain CA certificates

Reads broker.domainCerts.folder and broker.domainCerts.files from the env
file, uploads each listed certificate file, then loads all of them into
the broker with one CLI script. It targets the primary node only, even in
an HA deployment.

With no domain CAs configured this is a safe no-op: it logs that nothing
is set and returns without touching the broker, rather than failing.

```
solace-util config apply domain-certs [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util config apply product-keys

Apply the configured product keys

Fails loud with "no product keys configured" if broker.productKeys is empty
in the env file -- there is nothing to self-skip here, unlike domain-certs.

Applies to the primary node, plus the backup on a Kubernetes HA deployment
(the monitor is skipped -- it carries no message spool); docker and podman
have only the primary. Each key is validated before anything is uploaded,
and the CLI script that applied it is removed from the node afterward.

```
solace-util config apply product-keys [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util config apply server-cert

Load/update the TLS server certificate

On Kubernetes, when kubernetes.tlsServerSecret is set, this rebuilds and
re-applies that Secret in place -- the broker already has it mounted and
re-reads it, so no pod restart runs here. Otherwise, and always on
docker/podman, it uploads the certificate over the broker CLI: on
Kubernetes to every role in the deployment (primary alone, or primary,
backup and monitor together when redundancy is enabled); on docker/podman,
only to this host's one container.

Either path needs tls.cert and tls.certKey set.

```
solace-util config apply server-cert [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util config delete

Remove configuration from the running broker

Only domain certificates can be withdrawn this way. A server certificate, the
default-VPN hardening and an applied product key all stay applied.

```
solace-util config delete
```

Runs nothing on its own. It names what `solace-util config delete` can act on -- `domain-certs` -- and given a word it does not know it fails rather than reporting success.


### solace-util config delete domain-certs

Remove the configured domain CA certificates

Runs `no ssl domain-certificate-authority <name>` for every CA still listed
under `broker.domainCerts` in the env file, then shows the broker's resulting
CA list -- drop a CA from the file first and this command no longer knows to
remove it; with none configured it makes no CLI call at all.

The certificate files `config apply domain-certs` uploaded stay in place --
only the CA configuration entries come out. On Kubernetes this runs against
the primary pod only, same as `config apply domain-certs`.

```
solace-util config delete domain-certs [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util config disable

Shut down the broker's built-in defaults (hardening)

Both steps are one-way: this tool has no command to re-enable what they shut down.

```
solace-util config disable
```

Runs nothing on its own. It names what `solace-util config disable` can act on -- `default-users`, `default-vpn` -- and given a word it does not know it fails rather than reporting success.


### solace-util config disable default-users

Shut down the default client-usernames in all VPNs

You query every message-VPN on the primary node (`show message-vpn *`) and
shut down the `default` client-username inside each one found -- not just
a VPN literally named `default`. If no VPN names parse from that output,
you get a warning and the broker is left untouched rather than the command
failing.

On docker and podman, "the primary node" is simply the host's one
container.

```
solace-util config disable default-users [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util config disable default-vpn

Shut down the default message-VPN

Uploads and runs two CLI scripts on the primary node (or the local
container): one shuts down the VPN's authentication (basic and
client-certificate), its default client-username, and every service under
it (SMF, REST incoming, MQTT and its websocket variants, AMQP, and
web-transport), also turning off SSL-downgrade-to-plain-text; the other
lists the resulting VPNs. Both scripts are removed once that listing runs.

It never reaches beyond that node -- on a Kubernetes deployment with
redundancy enabled, unlike `apply server-cert` (backup and monitor) and
`apply product-keys` (backup). Elsewhere (docker, podman, or standalone
Kubernetes) those two commands are primary-only as well.

```
solace-util config disable default-vpn [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util config leader

Assert the config-sync leader (HA only) (containers: run on the primary; [role] is this host, detected from its name when omitted)

No-ops on a standalone deployment rather than erroring. In HA it first
reverts any released activity back home -- straight to the backup pod on
Kubernetes, over SEMP to the mate on containers (a warning, not a failure,
if the mate is unreachable) -- then waits for redundancy to report restored
on the primary and asserts the primary as config-sync leader for the router
and every message-VPN.

In HA, containers refuse this on the backup or monitor host; on a
standalone deployment every host no-ops instead, since that check runs
first.

```
solace-util config leader [role] [flags]
```

Arguments: `[role]` is one of `primary` (`p`), `backup` (`b`), `monitor` (`m`) -- see [Reading this reference](#reading-this-reference).

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util convert

Convert a legacy bash env file into a YAML env file

Convert a legacy bash env file -- the pre-Go format sourced by bash/000-env.sh --
into the YAML env file this CLI reads.

The target platform section is detected from the variables present; pass
--platform to choose it yourself. Variables with no YAML equivalent are
reported on stderr rather than dropped silently.

The output carries every secret from the source file verbatim, so treat it
like the source: write it with -o rather than through a shared terminal, and
never commit it.

  solace-util convert bash/env/prod -o prod.yaml
  solace-util convert bash/env/prod --platform podman -o prod.yaml
  solace-util check deploy -e prod.yaml

```
solace-util convert <bash-env-file> [flags]
```

Also available as: cv

| Flag | Default | Meaning |
| --- | --- | --- |
| `--force` | `false` | overwrite the --out file if it already exists |
| `-o`, `--out` | (none) | write the YAML here instead of stdout |


### solace-util copy

Copy files to/from the broker

On Kubernetes, `--pod` (p|b|m, default primary) on `from`/`into` selects
the pod; docker and podman route to this host's single broker container
regardless, so the verbs exist only so a script need not know which
platform it is driving.

`from` lands each file under its basename in the current directory; `into`
defaults to `.` inside the pod or container -- the pod's login directory
on Kubernetes -- unless `--dir` names another one. Both attempt every
file and report per-file failures, exiting non-zero with a count instead
of stopping at the first bad path.

```
solace-util copy
```

Runs nothing on its own. It names what `solace-util copy` can act on -- `from`, `into` -- and given a word it does not know it fails rather than reporting success.

Also available as: cp


### solace-util copy from

Copy files from the broker to the host

For example, `copy from /var/lib/solace/logs/debug.log` writes
`./debug.log` in your current directory -- `kubectl cp` on Kubernetes,
`<runtime> cp` on docker/podman.

```
solace-util copy from files... [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--pod` | (none) | pod role to target (p\|b\|m) (kubernetes only) |


### solace-util copy into

Copy files from the host into the broker

`into` never checks first that the pod or container exists, so copying
into an undeployed broker surfaces as a plain `cp` error.

```
solace-util copy into files... [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--dir` | (none) | destination directory inside the broker |
| `--pod` | (none) | pod role to target (p\|b\|m) (kubernetes only) |


### solace-util deploy

Deploy the broker, the operator, or the whole broker stack

`deploy broker` applies just the broker. `deploy all` runs the whole bring-up
for it: check -> prepare -> deploy -> assert the config-sync leader (HA).

Neither installs the operator. It is cluster-scoped and may already be serving
other brokers, so `deploy operator` is its own command -- run it once per
cluster. `check deploy` reports when it is missing.

```
solace-util deploy
```

Runs nothing on its own. It names what `solace-util deploy` can act on -- `all`, `broker`, `operator` -- and given a word it does not know it fails rather than reporting success.

Also available as: dp


### solace-util deploy all

Orchestrate the whole bring-up for this broker

Kubernetes runs check -> create namespace -> create secrets -> apply the
broker CR, then asserts the config-sync leader when HA is enabled. Docker
and podman stop after check -> prepare host -> deploy -- the leader is
asserted separately, with `config leader` run on the primary once every
host is up.

[role] (docker/podman only) tells this host which identity -- primary,
backup, or monitor -- to deploy its own container as; run it once per
host, each time with that host's own role. Omitting it silently
defaults to primary rather than detecting it from the hostname, unlike
`config leader`.

```
solace-util deploy all [role] [flags]
```

Arguments: `[role]` is one of `primary` (`p`), `backup` (`b`), `monitor` (`m`) -- see [Reading this reference](#reading-this-reference).

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--restart` | `false` | restart an already-running broker when the deploy artifact changed (otherwise you are asked, and a non-interactive run leaves it running) (docker/podman only) |


### solace-util deploy broker

Deploy the broker (containers: this host's container; in HA the role is detected from the hostname unless given)

Kubernetes: applies the PubSubPlusEventBroker CR via the configured
`kubernetes.runtime` (`kubectl` by default). It creates neither the namespace
nor the secrets -- run `prepare all` first, or use `deploy all` for both plus
the broker in one step.

Docker and podman: `prepare host` must have generated the PSK first. Podman
loads every secret into its own store before writing the quadlet unit;
docker's compose file only names an env var, filled in at run time. The
artifact is then started, or -- if already running and changed -- restarted
with `--restart` or your confirmation. `[role]` picks which host this is and
is refused on Kubernetes.

```
solace-util deploy broker [role] [flags]
```

Arguments: `[role]` is one of `primary` (`p`), `backup` (`b`), `monitor` (`m`) -- see [Reading this reference](#reading-this-reference).

Also available as: br

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--restart` | `false` | restart an already-running broker when the deploy artifact changed (otherwise you are asked, and a non-interactive run leaves it running) (docker/podman only) |


### solace-util deploy operator

Install the cluster-scoped EventBroker Operator (kubernetes only)

After checking permission to create customresourcedefinitions, this
installs the operator in three applies: the bundle's own Namespace
document, then -- as a separate artifact, not part of the bundle -- the
image-pull secret when `kubernetes.imagePullSecret` is set, then the
rest of the bundle (CRDs, RBAC and the controller Deployment). That is
exactly what `generate operator` and `generate secrets operator` print.

Re-running it rolls the operator to a newer `kubernetes.operator.image`;
pointing it at an OLDER image warns you and asks before applying
anything, and running it unattended refuses outright.

```
solace-util deploy operator [flags]
```

Also available as: op

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util diagnostics

Gather a support bundle from the broker into broker.diagDir

Runs the broker's full `show` command sweep plus `gather-diagnostics`, then
downloads the zipped output plus the diagnostics bundle into broker.diagDir
-- a failed download of the bundle is a warning, not a failed run.

Kubernetes gathers one archive per HA node (primary, backup, monitor when
redundancy is enabled); docker and podman only ever have this host's one
broker. Each run deletes the remote artifacts after downloading them.

```
solace-util diagnostics [flags]
```

Also available as: diag

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--days` | `1` | days of logs/diagnostics to gather |


### solace-util examples

Print a sample env file to start from

[name] is kubernetes, docker, podman or full. The three platform names give a minimal
standalone env file carrying only the keys that platform cannot default, and
declaring only its own section -- so the file it writes needs no --platform.
`full` is the complete annotated schema: every key the loader accepts, the
default each omitted one takes, and all three platform sections at once.
Omit [name] to list them.

Nothing here reads an env file or contacts anything, so -e/--env is ignored
and no cluster or container runtime has to exist yet.

Every secret in the output is a CHANGE-ME placeholder: an unedited file is
refused by the broker rather than deployed with a password that is public in
this repo. Each one has a sibling *Env key naming a variable to read instead,
which is what keeps a committed env file secret-free.

  solace-util examples docker -o env/dev.yaml
  solace-util examples full | less
  solace-util eg kubernetes -o env/prod.yaml && solace-util check deploy -e env/prod.yaml

```
solace-util examples [name] [flags]
```

Also available as: eg

| Flag | Default | Meaning |
| --- | --- | --- |
| `--force` | `false` | overwrite the --out file if it already exists |
| `-o`, `--out` | (none) | write the env file here instead of stdout |


### solace-util generate

Render a deployment artifact to stdout without applying it

Nothing here contacts the cluster or the container engine, so it is safe to run
against an env file you have not vetted.

The nouns are the same ones the acting verbs use: `generate broker` renders what
`deploy broker` would apply, whichever platform that is -- a custom resource on
Kubernetes, a compose file or systemd quadlet on a container host (which is
per-host, so it takes a [role] there).

Only the operator targets are platform-scoped, and because the thing does not
exist elsewhere rather than because it goes by another name: there is no
container operator to install.

```
solace-util generate
```

Runs nothing on its own. It names what `solace-util generate` can act on -- `broker`, `operator`, `secrets` -- and given a word it does not know it fails rather than reporting success.

Also available as: gen


### solace-util generate broker

Render what `deploy broker` would apply

Kubernetes: the PubSubPlusEventBroker custom resource. Docker and podman: this
host's deploy artifact -- a compose file or a systemd quadlet unit -- which is
per-host, so [role] selects which node's artifact to render.

```
solace-util generate broker [role] [flags]
```

Arguments: `[role]` is one of `primary` (`p`), `backup` (`b`), `monitor` (`m`) -- see [Reading this reference](#reading-this-reference).

Also available as: br

Renders to stdout and changes nothing: it runs no external command, so it needs no cluster or runtime, runs no preflight, and refuses `--allow-command` -- there is nothing here for it to approve.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util generate operator

Render the operator install bundle (kubernetes only)

`deploy operator` applies this bundle's namespace first, then the
image-pull secret `generate secrets operator` renders when one is
configured, then the rest -- applying that secret before its
namespace exists is what once failed a first install.

```
solace-util generate operator [flags]
```

Also available as: op

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

Renders to stdout and changes nothing: it runs no external command, so it needs no cluster or runtime, runs no preflight, and refuses `--allow-command` -- there is nothing here for it to approve.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util generate secrets

Render a secret artifact

`generate secrets broker` renders what the broker deployment needs -- Secret
manifests on Kubernetes, a shell script on a container host.

`generate secrets operator` renders the operator's image-pull secret, the one
part of the operator install that carries a credential. `generate operator`
renders the rest of that install and carries no secret value, so the two
together are exactly what `deploy operator` applies.

```
solace-util generate secrets
```

Runs nothing on its own. It names what `solace-util generate secrets` can act on -- `broker`, `operator` -- and given a word it does not know it fails rather than reporting success.


### solace-util generate secrets broker

Render the broker's secret-creation artifact (Kubernetes: Secret manifests; containers: a shell script)

Kubernetes prints the same manifest `prepare secrets` applies -- admin secret
always, TLS and image-pull secrets only when kubernetes.tlsServerSecret /
kubernetes.imagePullSecret name one. Docker and podman print a shell script
instead: `secret create --replace` lines for podman, `export` lines for
docker to source before a manual `docker compose up` -- `deploy broker` sets
those variables itself, so the script is for a hand-run compose only.

Refuses to print an artifact for a secret whose value is still unset; on
containers, a missing nodes.psk points you at `prepare host` to generate it.

```
solace-util generate secrets broker [flags]
```

Also available as: br

Renders to stdout and changes nothing: it runs no external command, so it needs no cluster or runtime, runs no preflight, and refuses `--allow-command` -- there is nothing here for it to approve.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util generate secrets operator

Render the operator's image-pull secret (kubernetes only)

The rendered Secret is always named `regcred`, not the value of
`kubernetes.imagePullSecret` -- that field only gates whether this renders
at all (unset, it fails naming the setting), since the operator's own
manifests reference that literal name.

It resolves `kubernetes.operator.namespace`, falling back to the default
operator namespace when that is unset -- there is no live cluster here to
discover which one is actually running.

```
solace-util generate secrets operator [flags]
```

Also available as: op

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

Renders to stdout and changes nothing: it runs no external command, so it needs no cluster or runtime, runs no preflight, and refuses `--allow-command` -- there is nothing here for it to approve.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util logs

Tail broker or operator logs

On Kubernetes, `broker` and `operator` both run a single `kubectl logs` and
print whatever is already buffered there -- neither one follows the stream,
so the command exits as soon as that snapshot is printed.

```
solace-util logs
```

Runs nothing on its own. It names what `solace-util logs` can act on -- `broker`, `operator` -- and given a word it does not know it fails rather than reporting success.

Also available as: lg


### solace-util logs broker

Tail the broker's logs

On Kubernetes this runs a single `kubectl logs` against the picked pod
(default: primary) and returns; it does not follow new lines. On docker and
podman it runs `<runtime> logs -f` against this host's one container
instead, which streams continuously until you interrupt it.

--pod only applies on Kubernetes -- a container host has one broker per
machine, so --pod there is refused rather than accepted and silently
dropped.

```
solace-util logs broker [flags]
```

Also available as: br

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--pod` | (none) | pod role to target (p\|b\|m) (kubernetes only) |


### solace-util logs operator

Tail the operator's controller logs (kubernetes only)

Runs `kubectl logs` against the operator's controller Deployment: one
snapshot of what it has already written, not a live follow, so watching a
rollout means calling this again rather than leaving it open.

The namespace comes from kubernetes.operator.namespace when set, otherwise
the built-in default `deploy operator` installs to -- a missing or
unreachable operator just surfaces as kubectl's own error.

```
solace-util logs operator [flags]
```

Also available as: op

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util prepare

Prepare the prerequisites a broker deployment needs

Everything a broker needs to exist before it is deployed.

`prepare all` runs the steps that are needed every time and need no input --
the namespace and its secrets on Kubernetes, the host on docker and podman --
so it is safe to script. `deploy all` runs the same steps for you.

Two things are deliberately outside it. The operator is cluster-scoped and
shared between brokers, so it is installed and removed on its own
(`deploy operator`). And `prepare labels` cannot be scripted at all: the env
file names the label each broker role wants, but only you can say which
machine should carry it, so it prompts -- run it once when provisioning the
cluster, not on every deployment.

```
solace-util prepare
```

Runs nothing on its own. It names what `solace-util prepare` can act on -- `all`, `host`, `labels`, `namespace`, `secrets` -- and given a word it does not know it fails rather than reporting success.

Also available as: pre


### solace-util prepare all

Run every applicable prepare step, in order

On Kubernetes this creates the namespace, then applies the admin/monitor
credentials secret plus -- if configured -- the TLS server and image-pull
secrets, via `kubectl apply`; idempotent, so a namespace failure stops
before any secret is touched.

Docker and podman: identical to `prepare host` -- data directory, DNS
check, registry login when both image.user and image.pass are set (one
alone aborts the run), and the redundancy PSK (HA only) generated once;
once nodes.psk is set, later runs just remind you to keep it identical.

```
solace-util prepare all [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util prepare host

Create/own the data dir, verify DNS, generate the redundancy PSK (docker/podman only)

Resolves the redundancy hostnames first (a miss fails loud in HA, but
standalone just continues), then logs in to the image registry when image.user
and image.pass are both set (a loud error if only one is). Outside standalone
mode it also generates nodes.psk when empty and writes it back to the env file
-- either way, copy the same value onto the other two hosts by hand.

On rootless podman it first warns if this process is running as root (prep
must run as the target rootless user), then chowns through `podman unshare`
and checks this user's hard nofile limit against ulimits.nofile, stopping with
the exact fix when it falls short.

```
solace-util prepare host [flags]
```

Applies to: docker, podman. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util prepare labels

Label cluster nodes for primary/backup/monitor placement (interactive, one-off) (kubernetes only)

With no placement labels configured (kubernetes.placement.labels*) this is
a no-op that says so and exits cleanly. Otherwise it needs a real
terminal: there is no flag for the node choice, so a non-interactive run
refuses outright instead of failing deep inside the prompt.

Per configured role it lists the cluster's nodes, asks you to pick one,
then runs `kubectl label node --overwrite` for each configured key=value.
Kubernetes-managed prefixes (kubernetes.io/, k8s.io/, and similar) are
silently skipped, and a failed label is reported and skipped rather than
aborting the rest.

```
solace-util prepare labels [flags]
```

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util prepare namespace

Create the broker namespace (kubernetes only)

Applies a bare Namespace object named for `kubernetes.namespace` -- no
labels, quotas, or other content. `kubectl apply` makes this safe to
re-run: an existing namespace is left as it is.

Run this before `prepare secrets`, since a Secret cannot be applied into
a namespace that does not exist yet; `prepare all` and `deploy all`
already run the two in that order.

```
solace-util prepare namespace [flags]
```

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util prepare secrets

Create admin/monitor, TLS, and image-pull secrets (kubernetes only)

Always creates the admin/monitor secret. The TLS secret joins it only
when kubernetes.tlsServerSecret is set; the image-pull secret only when
kubernetes.imagePullSecret is set.

Applied with `kubectl apply` on stdin -- no secret value reaches an argv
or echoed command -- and the run is idempotent, safe to re-run after a
change.

The manifest names the namespace directly, so `prepare namespace` (or
`prepare all`) must run first. `generate secrets broker` renders the
same manifest without applying it, for review.

```
solace-util prepare secrets [flags]
```

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util remove

Remove the broker, the operator, or the whole broker stack

Every command here asks before it removes anything, and --no-prompt is the one
flag that makes it silent -- a script switches off one thing, not one per
question.

Nothing here removes the layer that is expensive to get back unless you say so:
the broker's persistent data and the operator's CRDs are kept by default, you
are asked about them separately, and what happened is printed either way. The
two flags compose, so an unattended removal that also drops the data is
`--delete-data --no-prompt`: naming the data you are willing to lose is not the
same as confirming the removal, so neither flag implies the other.

`remove all` takes this broker and its namespace. It leaves the operator, which
is cluster-scoped and may be serving brokers this env file does not describe.

```
solace-util remove
```

Runs nothing on its own. It names what `solace-util remove` can act on -- `all`, `broker`, `namespace`, `operator`, `secrets` -- and given a word it does not know it fails rather than reporting success.

Also available as: rm


### solace-util remove all

Remove the broker, its secrets and its namespace (the operator is kept)

On Kubernetes this deletes the broker, then its secrets, then the
namespace -- exactly reversing the order `prepare all` creates them
in. The namespace delete also takes anything else that happens to
live in it, not just what this env file created.

On docker and podman there is no separate namespace layer, and podman's
secret store is removed with the container rather than as its own step,
so this is exactly `remove broker`.

```
solace-util remove all [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--delete-data` | `false` | delete the broker's persistent data too (Kubernetes PVCs / the container data directory). Without it the data is kept |
| `--no-prompt` | `false` | do not ask anything: proceed with the removal, and keep whatever is kept by default unless a --delete-* flag says otherwise |


### solace-util remove broker

Remove the deployed broker

On Kubernetes this deletes the PubSubPlusEventBroker custom resource with
`--ignore-not-found`, so running it again after a successful removal is a
no-op. On docker and podman it stops the broker container: docker via
`compose down` when a compose file is on disk (a plain stop and rm by
name otherwise) -- it never deletes the compose file; podman via
`systemctl stop` followed by deleting the `.container` quadlet unit
file and a daemon-reload.

It keeps the broker's persistent data by default -- Kubernetes PVCs, or
the container's data directory -- and reports that either way; pass
`--delete-data` to remove it too. Beyond the broker and, if requested,
its data, this command touches nothing else: secrets, namespace, and
the operator all survive.

```
solace-util remove broker [flags]
```

Also available as: br

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--delete-data` | `false` | delete the broker's persistent data too (Kubernetes PVCs / the container data directory). Without it the data is kept |
| `--no-prompt` | `false` | do not ask anything: proceed with the removal, and keep whatever is kept by default unless a --delete-* flag says otherwise |


### solace-util remove namespace

Delete the broker's namespace (kubernetes only)

Deletes the whole namespace, not just what your env file describes -- every
secret and PVC Kubernetes considers namespaced to it goes in the same
cascade.

A repeat run is a no-op rather than an error: the underlying delete already
tolerates a missing namespace. You get the namespace back cheaply (`prepare
namespace` recreates it from the env file), but whatever you stored on its
volumes is gone for good.

```
solace-util remove namespace [flags]
```

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--no-prompt` | `false` | do not ask anything: proceed with the removal, and keep whatever is kept by default unless a --delete-* flag says otherwise |


### solace-util remove operator

Remove the cluster-scoped EventBroker Operator (kubernetes only)

Once confirmed, it deletes the operator's namespace, RBAC and controller
Deployment outright. The CustomResourceDefinitions are the one piece kept
by default, since deleting them cascades to every PubSubPlusEventBroker in
the cluster, not just the broker this env file describes.

This never runs as part of `remove broker` or `remove all`: run it on its
own, once nothing else in the cluster still depends on it; `deploy
operator` reinstalls it afterward.

```
solace-util remove operator [flags]
```

Also available as: op

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--delete-crd` | `false` | delete the operator's CustomResourceDefinitions too. Without it they are kept, so existing brokers survive |
| `--no-prompt` | `false` | do not ask anything: proceed with the removal, and keep whatever is kept by default unless a --delete-* flag says otherwise |


### solace-util remove secrets

Delete the broker's secrets (kubernetes only)

Deletes the admin/monitor credentials secret, and -- if configured -- the
TLS server secret and the image-pull secret: the same three `prepare
secrets` creates. --ignore-not-found means an already-missing one is not
an error, but a real failure stops before the rest are touched.

Removes only the Secret objects -- the namespace, the broker's other
resources, and the operator's CRDs are untouched, and `prepare secrets`
(or `prepare all`) recreates every one from the env file, which is why
there is no --delete-* flag here to keep one back.

```
solace-util remove secrets [flags]
```

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--no-prompt` | `false` | do not ask anything: proceed with the removal, and keep whatever is kept by default unless a --delete-* flag says otherwise |


### solace-util restart

Bounce a running broker or the operator

Restarting applies nothing new. A changed deploy artifact needs
`deploy broker` (containers: with --restart), which rewrites it first.

```
solace-util restart
```

Runs nothing on its own. It names what `solace-util restart` can act on -- `broker`, `operator` -- and given a word it does not know it fails rather than reporting success.

Also available as: rs


### solace-util restart broker

Restart the broker (Kubernetes: delete pods so the statefulset recreates them)

For kubernetes.updateStrategy=manualPodRestart: `deploy broker` updates the
statefulset's pod template but the operator waits for a pod to be deleted before
applying it.

With no --pod, every pod is restarted in the safe order (monitor, backup, primary;
standalone: just the primary), waiting for each to become ready before the next.
The order is by configured role, not by which node is currently active -- after a
failover they differ. Check `solace-util smoke redundancy` first, or pass --pod
and restart them one at a time.

On docker and podman there is one broker per host and no pod to pick: the
container is restarted in place.

```
solace-util restart broker [flags]
```

Also available as: br

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--no-prompt` | `false` | do not ask anything: proceed with the removal, and keep whatever is kept by default unless a --delete-* flag says otherwise |
| `--pod` | (none) | pod role to target (p\|b\|m) (kubernetes only) |


### solace-util restart operator

Restart the operator's controller deployment (kubernetes only)

Runs `kubectl rollout restart deployment` on the operator's own
controller deployment -- not the broker's statefulset -- in whatever
namespace it resolves to, which need not be the broker's own namespace.

Unlike `restart broker`, this asks nothing first: the controller carries
no messaging traffic, so there is no traffic impact to confirm.

It bounces the controller already installed; a changed bundle still
needs `deploy operator`, which re-applies the manifest.

```
solace-util restart operator [flags]
```

Also available as: op

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util shell

Open an interactive shell in the broker

Kubernetes execs `bash` into the --pod pod (`kubectl exec -it`, default
primary); docker and podman exec into this host's one broker container
instead (`<runtime> exec -it`). An explicit --pod there is refused
rather than accepted -- the same flag scoping `cli` uses -- since there
is only one container on the host to target.

This is a plain OS shell inside the broker image, not the Solace CLI --
use `cli` for the broker's own CLI, or its `--input` to run a script
instead. Either way it targets the pod or container by name, so it must
already exist and be running.

```
solace-util shell [flags]
```

Also available as: sh

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--pod` | (none) | pod role to target (p\|b\|m) (kubernetes only) |


### solace-util smoke

Run invasive checks that exercise the broker

These checks prove the broker works by making it work, so they disturb it.
Read-only questions live under `check`.

```
solace-util smoke
```

Runs nothing on its own. It names what `solace-util smoke` can act on -- `redundancy` -- and given a word it does not know it fails rather than reporting success.


### solace-util smoke redundancy

Exercise a real failover and fail back (HA only) (containers: run on the primary; [role] is this host, detected from its name when omitted)

On a standalone deployment this is a no-op: a [SKIP] line, not a failure --
there is no redundancy pair to exercise.

It normally walks the group through release, un-release, then revert,
confirming `show redundancy` after each step -- skipping straight to the
revert if the Primary is not already active. On containers, the backup's
one mutation rides SEMP to nodes.backup.ip, preflighted before release.

A run that dies partway (Ctrl-C included) can leave the group released or
failed over; re-run once you have fixed the cause.

```
solace-util smoke redundancy [role] [flags]
```

Arguments: `[role]` is one of `primary` (`p`), `backup` (`b`), `monitor` (`m`) -- see [Reading this reference](#reading-this-reference).

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util start

Start a broker that is deployed but not running

The counterpart to `stop broker`: the deploy artifact and the persistent data
are already on disk, so this only brings the broker back up. It applies
nothing new -- a changed artifact needs `deploy broker` first.

There is nothing to start until a broker has been deployed; on a host or
cluster that has never had one, `deploy broker` is the command you want.

```
solace-util start
```

Runs nothing on its own. It names what `solace-util start` can act on -- `broker` -- and given a word it does not know it fails rather than reporting success.


### solace-util start broker

Start the broker (Kubernetes: scale the statefulset(s) to 1; containers: start the container)

On kubernetes, a redundancy group comes up primary, backup, monitor -- the
reverse of `restart broker`'s order -- waiting for each statefulset to report
ready before scaling the next. Standalone scales just the primary.

Docker runs `compose start` against the deployed compose file when one
exists, falling back to a plain `docker start` on the container itself
when it does not. Podman always starts the quadlet's systemd unit instead,
since quadlet owns the container once deployed. Neither path creates a
broker that was never deployed.

```
solace-util start broker [flags]
```

Also available as: br

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util status

Report on the broker or the operator

By default this reports the RUNNING artifacts. --detail adds the static ones --
the full description of what is deployed, load balancer included.

```
solace-util status
```

Runs nothing on its own. It names what `solace-util status` can act on -- `broker`, `operator` -- and given a word it does not know it fails rather than reporting success.

Also available as: sts


### solace-util status broker

Show the broker's deployment status

Kubernetes reports the operator's CR conditions, then pods, services, and
statefulsets -- never the broker, so one unreachable over SEMP still reports
up. Docker and podman check the systemd unit or compose state, then `ps`.

`--detail` and `--all` compose instead of replacing each other. Alone, it
describes the --pod pod (primary by default; --pod is Kubernetes-only) and
the load balancer Service, or runs `inspect` on containers (podman: plus the
installed unit). Under `--all` it instead adds an image column to the
survey, or inspects every container found.

```
solace-util status broker [flags]
```

Also available as: br

| Flag | Default | Meaning |
| --- | --- | --- |
| `--all` | `false` | report every Solace broker found, not just the one this env file describes (Kubernetes: across all namespaces; docker/podman: every Solace container on this host) |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--detail` | `false` | include the static artifacts, not just the running ones (Kubernetes: secrets, configmaps and PVCs; docker/podman: mounts, which is also where secrets appear) |
| `--pod` | (none) | pod role to target (p\|b\|m) (kubernetes only) |


### solace-util status operator

Show the operator's controller status (kubernetes only)

Reads the operator's controller Deployment (`pubsubplus-eventbroker-operator`)
in whichever namespace `kubernetes.operator.namespace` names, or else the
built-in default `pubsubplus-operator-system` that `deploy operator` installs
to -- the same two rules everywhere, so this reads the operator this env file
would deploy and never one belonging to someone else.

Without --detail, a missing Deployment is reported as not installed
rather than failing the command; with --detail, the follow-up `kubectl
describe` fails on it instead. This checks the controller's own health,
not any broker's -- see `status broker` for that.

```
solace-util status operator [flags]
```

Also available as: op

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--detail` | `false` | include the full description of the operator deployment |


### solace-util stop

Stop a running broker without removing it

The deployment, its persistent data and its configuration all survive --
`start broker` brings it back. Use `remove broker` to delete it.

```
solace-util stop
```

Runs nothing on its own. It names what `solace-util stop` can act on -- `broker` -- and given a word it does not know it fails rather than reporting success.


### solace-util stop broker

Stop the broker (Kubernetes: scale the statefulset(s) to 0; containers: stop the container)

On Kubernetes every role's StatefulSet scales to zero together in one
command (all three in HA; just the primary standalone) -- there is no
role argument. It returns as soon as the scale succeeds; unlike
`start broker`, it does not wait for the pods to actually terminate.

Unlike `restart broker` and every removal, this does not ask for
confirmation before acting.

```
solace-util stop broker [flags]
```

Also available as: br

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util version

Print the solace-util version

Print the version this binary was built at, plus the Go toolchain and
platform that built it -- useful to paste alongside a support request.

A release binary (built by scripts/dev.sh or dev.ps1) reports the git tag
it shipped as, e.g. v1.2.3 -- matching the GitHub release exactly. A plain
`go build .` with no version stamped reports "dev".

```
solace-util version
```

Also available as: ver


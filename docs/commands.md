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
  broker
    cli
    configure
      data-replication
      default-users
      default-vpn
      domain-certs
      product-keys
      server-certs
    copy
      from files...
      into files...
    deploy
    generate
    logs
    perform
      assert-leader
      cli-script <file>
      data-replication
      export-config
      gather-diagnostics
      import-config <file>
      redundancy-test
      semp-login-check
      shell-script <file>
    remove
    restart
    shell
    start
    status
    stop
    validate
  convert <bash-env-file>
  examples
  operator
    deploy
    generate
    logs
    remove
    restart
    start
    status
    stop
    validate
  validate
  version
```

## Index

- [auto-complete](#solace-util-auto-complete)
  - [auto-complete bash](#solace-util-auto-complete-bash)
  - [auto-complete fish](#solace-util-auto-complete-fish)
  - [auto-complete powershell](#solace-util-auto-complete-powershell)
  - [auto-complete zsh](#solace-util-auto-complete-zsh)
- [broker](#solace-util-broker)
  - [broker cli](#solace-util-broker-cli)
  - [broker configure](#solace-util-broker-configure)
    - [broker configure data-replication](#solace-util-broker-configure-data-replication)
    - [broker configure default-users](#solace-util-broker-configure-default-users)
    - [broker configure default-vpn](#solace-util-broker-configure-default-vpn)
    - [broker configure domain-certs](#solace-util-broker-configure-domain-certs)
    - [broker configure product-keys](#solace-util-broker-configure-product-keys)
    - [broker configure server-certs](#solace-util-broker-configure-server-certs)
  - [broker copy](#solace-util-broker-copy)
    - [broker copy from](#solace-util-broker-copy-from)
    - [broker copy into](#solace-util-broker-copy-into)
  - [broker deploy](#solace-util-broker-deploy)
  - [broker generate](#solace-util-broker-generate)
  - [broker logs](#solace-util-broker-logs)
  - [broker perform](#solace-util-broker-perform)
    - [broker perform assert-leader](#solace-util-broker-perform-assert-leader)
    - [broker perform cli-script](#solace-util-broker-perform-cli-script)
    - [broker perform data-replication](#solace-util-broker-perform-data-replication)
    - [broker perform export-config](#solace-util-broker-perform-export-config)
    - [broker perform gather-diagnostics](#solace-util-broker-perform-gather-diagnostics)
    - [broker perform import-config](#solace-util-broker-perform-import-config)
    - [broker perform redundancy-test](#solace-util-broker-perform-redundancy-test)
    - [broker perform semp-login-check](#solace-util-broker-perform-semp-login-check)
    - [broker perform shell-script](#solace-util-broker-perform-shell-script)
  - [broker remove](#solace-util-broker-remove)
  - [broker restart](#solace-util-broker-restart)
  - [broker shell](#solace-util-broker-shell)
  - [broker start](#solace-util-broker-start)
  - [broker status](#solace-util-broker-status)
  - [broker stop](#solace-util-broker-stop)
  - [broker validate](#solace-util-broker-validate)
- [convert](#solace-util-convert)
- [examples](#solace-util-examples)
- [operator](#solace-util-operator)
  - [operator deploy](#solace-util-operator-deploy)
  - [operator generate](#solace-util-operator-generate)
  - [operator logs](#solace-util-operator-logs)
  - [operator remove](#solace-util-operator-remove)
  - [operator restart](#solace-util-operator-restart)
  - [operator start](#solace-util-operator-start)
  - [operator status](#solace-util-operator-status)
  - [operator stop](#solace-util-operator-stop)
  - [operator validate](#solace-util-operator-validate)
- [validate](#solace-util-validate)
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

Deploy and operate Solace PubSub+ Event Brokers from one YAML env file, with the
same commands on every platform. Name the thing, then the verb:

  broker <verb>      the broker this env file describes
  operator <verb>    the cluster-scoped EventBroker Operator (kubernetes only)
  validate           check the whole env file
  examples           write a starting env file

Every command takes -e/--env <file>; the platform comes from that file, or from
--platform kubernetes|docker|podman when it declares more than one.

Exit status: 0 worked, 2 bad command line or env file, 1 anything else.

```
solace-util
```

Subcommands: `auto-complete`, `broker`, `convert`, `examples`, `operator`, `validate`, `version`


### solace-util auto-complete

Print the shell auto-completion script for solace-util

Print a shell's completion script on stdout. Load it to complete commands and
flags, plus the values they take: env files for -e/--env, primary|backup|monitor
for --pod, platform names for --platform, and directories for --base-dir and
--dir.

To load it for every new shell:

  bash        solace-util auto-complete bash > /etc/bash_completion.d/solace-util
  zsh         solace-util auto-complete zsh > "${fpath[1]}/_solace-util"
  fish        solace-util auto-complete fish > ~/.config/fish/completions/solace-util.fish
  powershell  solace-util auto-complete powershell > $HOME\solace-util.ps1
              then add  . $HOME\solace-util.ps1  to $PROFILE

Each shell's own help has the one-liner for loading into the CURRENT shell
instead, and the prerequisites where a shell has any.

Completion never reads the env file, so it stays inert -- a TAB press cannot
parse config or run anything.

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

Load for every session. Write the script once, then dot-source it from your
profile -- generating it once is what keeps shell start-up fast, since the
alternative runs this binary on every new shell:

  solace-util auto-complete powershell > $HOME\solace-util.ps1
  Add-Content $PROFILE '. $HOME\solace-util.ps1'

If $PROFILE does not exist yet, create it first:

  New-Item -ItemType File -Force $PROFILE

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


### solace-util broker

Deploy and operate the broker

Everything that acts on the broker this env file describes.

  broker validate         check the env file and the cluster or host
  broker generate         see exactly what would be applied
  broker deploy           prerequisites + the broker, idempotent
  broker status           watch it come up
  broker configure ...    settings the env file describes (certs, keys, hardening)
  broker perform ...      one-shot actions (diagnostics, failover test)
  broker stop / start     pause it without removing it
  broker remove           tear it down, prompting for each layer


```
solace-util broker
```

Runs nothing on its own. It names what `solace-util broker` can act on -- `cli`, `configure`, `copy`, `deploy`, `generate`, `logs`, `perform`, `remove`, `restart`, `shell`, `start`, `status`, `stop`, `validate` -- and given a word it does not know it fails rather than reporting success.

Also available as: br


### solace-util broker cli

Open an interactive Solace CLI in the broker

Interactive only; execs `cli -A` into the picked pod (Kubernetes) or this
host's container (docker/podman).

To run a script instead of typing at a prompt, use
`broker perform cli-script <file>`.

```
solace-util broker cli [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |


### solace-util broker configure

Configure a deployed broker (certs, keys, hardening)

Settings the env file describes, applied to a deployed broker over its own
CLI.

On a fresh HA broker, run in this order:

  broker perform assert-leader     config-sync leader FIRST
  broker configure server-certs
  broker configure domain-certs
  broker configure default-vpn     hardening
  broker configure default-users
  broker configure product-keys    last

  broker configure data-replication   a DR pair, when there is one


```
solace-util broker configure
```

Runs nothing on its own. It names what `solace-util broker configure` can act on -- `data-replication`, `default-users`, `default-vpn`, `domain-certs`, `product-keys`, `server-certs` -- and given a word it does not know it fails rather than reporting success.

Also available as: cfg


### solace-util broker configure data-replication

Converge this broker to the replication: block

Converges THIS broker to the env file's replication: block -- the mate
addresses, which message-VPNs replicate, and each one's active/standby role.

It never contacts the mate. Run it at BOTH sites with the same file.

Shuts down replication on any replicating VPN the file does not list, and can
leave a VPN active at both sites if the mate still holds it -- move a role with
`broker perform data-replication` instead.

```
solace-util broker configure data-replication [flags]
```

Also available as: dr

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--no-prompt` | `false` | answer yes to the confirmation |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |


### solace-util broker configure default-users

Shut down the default client-usernames in all VPNs

Shuts down the default client-username in every message-VPN found on the
broker.

This blocks any client still relying on the default username to connect.

```
solace-util broker configure default-users [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--disable` | `false` | shut it down (the default; accepted explicitly so a script can say so) |
| `--enable` | `false` | start it back up instead of shutting it down |
| `--no-prompt` | `false` | do not ask: answer yes to every question this command would ask. It does not RAISE a question -- without a --delete-* flag the expensive layer is still kept |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |


### solace-util broker configure default-vpn

Shut down the default message-VPN

Shuts down the broker's default message-VPN.

This stops every client connection using it, and every service it fronts.

```
solace-util broker configure default-vpn [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--disable` | `false` | shut it down (the default; accepted explicitly so a script can say so) |
| `--enable` | `false` | start it back up instead of shutting it down |
| `--no-prompt` | `false` | do not ask: answer yes to every question this command would ask. It does not RAISE a question -- without a --delete-* flag the expensive layer is still kept |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |


### solace-util broker configure domain-certs

Load the configured domain CA certificates

Uploads the domain CA certificates broker.domainCerts describes and creates
a domain certificate authority for each.

--remove deletes those authorities from the broker; the certificate files
themselves stay.

```
solace-util broker configure domain-certs [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--apply` | `false` | apply what the env file configures (the default; accepted explicitly so a script can say so) |
| `--no-prompt` | `false` | do not ask: answer yes to every question this command would ask. It does not RAISE a question -- without a --delete-* flag the expensive layer is still kept |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |
| `--remove` | `false` | remove it from the broker instead of applying it |


### solace-util broker configure product-keys

Apply the configured product keys

Applies broker.productKeys to the primary node (and backup, in HA); fails
loud rather than silently succeeding when none are configured.

--remove revokes those keys and can leave the broker unlicensed.

```
solace-util broker configure product-keys [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--apply` | `false` | apply what the env file configures (the default; accepted explicitly so a script can say so) |
| `--no-prompt` | `false` | do not ask: answer yes to every question this command would ask. It does not RAISE a question -- without a --delete-* flag the expensive layer is still kept |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |
| `--remove` | `false` | remove it from the broker instead of applying it |


### solace-util broker configure server-certs

Load or update the TLS server certificate

Loads or updates the broker's TLS server certificate, over its own CLI or,
on Kubernetes with kubernetes.tlsServerSecret set, via the operator's Secret.

--remove takes TLS down immediately on every listener configured to present
a certificate.

```
solace-util broker configure server-certs [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--apply` | `false` | apply what the env file configures (the default; accepted explicitly so a script can say so) |
| `--no-prompt` | `false` | do not ask: answer yes to every question this command would ask. It does not RAISE a question -- without a --delete-* flag the expensive layer is still kept |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |
| `--remove` | `false` | remove it from the broker instead of applying it |


### solace-util broker copy

Copy files to or from the broker

Attempts every file in both directions and reports each one, rather than
stopping at the first failure; a partial failure still exits non-zero.

```
solace-util broker copy
```

Runs nothing on its own. It names what `solace-util broker copy` can act on -- `from`, `into` -- and given a word it does not know it fails rather than reporting success.

Also available as: cp


### solace-util broker copy from

Copy files from the broker to the host

Each file lands in the current directory under its base name. The base name is
split on both separators, so a remote path written with backslashes cannot
produce a local filename containing one.

```
solace-util broker copy from files... [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |


### solace-util broker copy into

Copy files from the host into the broker

--dir is the destination inside the broker, defaulting to the exec's working
directory. The destination is not checked first: a copy into a path that does
not exist fails for that file and says so.

```
solace-util broker copy into files... [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--dir` | (none) | destination directory inside the broker |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |


### solace-util broker deploy

Deploy the broker and its prerequisites

Applies every prerequisite -- namespace, secrets, TLS -- before creating the
broker itself, and is safe to re-run.

Kubernetes readiness is not waited on; use `broker status` to watch it come
up.

```
solace-util broker deploy [flags]
```

Also available as: dp

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (docker/podman only) |
| `--restart` | `false` | restart an already-running broker when the deploy artifact changed (otherwise you are asked, and a non-interactive run leaves it running) (docker/podman only) |


### solace-util broker generate

Render what deploy would apply, without applying it

Renders exactly what `broker deploy` would apply, without applying it.

Kubernetes output can be piped straight to `kubectl apply -f -`; prefer
-o/--out to `>` redirection, which PowerShell corrupts.

```
solace-util broker generate [flags]
```

Also available as: gen

Renders to stdout and changes nothing: it runs no external command, so it needs no cluster or runtime, runs no preflight, and refuses `--allow-command` -- there is nothing here for it to approve.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--no-prompt` | `false` | answer yes to the --out overwrite question (no effect without --out) |
| `-o`, `--out` | (none) | write the artifact to this file instead of stdout |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (docker/podman only) |


### solace-util broker logs

Read the broker's logs

Reads the picked pod's log (Kubernetes) or this host's container's log
(docker/podman).

```
solace-util broker logs [flags]
```

Also available as: lg

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `-f`, `--follow` | `false` | keep streaming new lines instead of exiting |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |
| `-p`, `--previous` | `false` | read the PREVIOUS container's logs instead of the current one (kubernetes only) |
| `--since` | (none) | read only lines newer than this duration (e.g. 30s, 5m, 2h) |
| `--tail` | (none) | print only this many trailing lines (a count, or `all`) |
| `--timestamps` | `false` | prefix each line with its timestamp |


### solace-util broker perform

Run a one-shot action against the broker

One-shot actions against a running broker; not settings the env file
describes.

  assert-leader        make this node the config-sync leader (HA)
  redundancy-test      exercise a real failover and fail back (HA, INVASIVE)
  gather-diagnostics   collect a support bundle into broker.diagDir
  semp-login-check     prove the admin credentials work over SEMP
  export-config        capture the broker's configuration as one artifact
  import-config <file> apply a captured configuration back (DESTRUCTIVE)
  cli-script <file>    run a Solace CLI script
  shell-script <file>  run a host shell script inside the broker
  data-replication     move replication roles across a DR pair (DESTRUCTIVE)


```
solace-util broker perform
```

Runs nothing on its own. It names what `solace-util broker perform` can act on -- `assert-leader`, `cli-script`, `data-replication`, `export-config`, `gather-diagnostics`, `import-config`, `redundancy-test`, `semp-login-check`, `shell-script` -- and given a word it does not know it fails rather than reporting success.

Also available as: pf


### solace-util broker perform assert-leader

Assert the config-sync leader (HA only)

Asserts this node as the config-sync leader for the router and every
message-VPN; a no-op, not an error, on a standalone deployment.

Run this first on a fresh HA broker, before any `broker configure` step.

```
solace-util broker perform assert-leader [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (docker/podman only) |


### solace-util broker perform cli-script

Run a Solace CLI script in the broker

Uploads a local Solace CLI script and runs it in the broker; a bare
filename resolves under broker.cliScriptsFolder.

A rejected line does not stop the rest of the script, but the run is then
reported as a failure.

```
solace-util broker perform cli-script <file> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |


### solace-util broker perform data-replication

Move replication roles across the DR pair

Moves each message-VPN's replication role to the site the env file names,
across BOTH brokers: demote at the losing site, confirm it, then promote.

It changes no configuration -- run `broker configure data-replication` first.
Refuses unless both sites are on their primary HA node and name each other.

Interrupts message flow for every VPN it moves.

```
solace-util broker perform data-replication [flags]
```

Also available as: dr

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--no-prompt` | `false` | answer yes to the role-change confirmation |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |


### solace-util broker perform export-config

Capture the broker's configuration as one artifact

Captures the broker's configuration as a replayable Solace CLI script; it
changes nothing on the broker.

--vpn NAME (repeatable) narrows to those VPNs; --broker-only excludes them
instead -- naming both is refused.

```
solace-util broker perform export-config [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--broker-only` | `false` | capture only broker-level configuration, and no message-VPN at all |
| `--no-prompt` | `false` | answer yes to the --out overwrite question (no effect without --out) |
| `-o`, `--out` | (none) | write the artifact to this file instead of stdout |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |
| `--vpn` | `[]` | capture only this message-VPN, repeatable for several; omit it and every VPN is captured |


### solace-util broker perform gather-diagnostics

Gather a support bundle into broker.diagDir

Runs the broker's full diagnostic sweep and downloads the resulting bundle
to broker.diagDir.

Kubernetes collects one bundle per role by default; --pod narrows to just
one.

```
solace-util broker perform gather-diagnostics [flags]
```

Also available as: gd

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--days` | `1` | days of logs/diagnostics to gather |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |


### solace-util broker perform import-config

Apply a captured configuration back to the broker

Applies a `broker perform export-config` artifact back to a running
broker. Only that command's own output is accepted; to run any other script,
use `broker perform cli-script`.

Any message-VPN in the artifact that already exists on the target is torn
down and rebuilt, destroying every message spooled in it.

```
solace-util broker perform import-config <file> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--no-prompt` | `false` | answer yes to the tear-down confirmation |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |


### solace-util broker perform redundancy-test

Exercise a real failover and fail back (HA only)

Fails the broker over to its mate and back for real; a no-op on a
standalone deployment.

A Ctrl-C partway through can leave the group failed over; see
docs/operations.md to recover.

```
solace-util broker perform redundancy-test [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (docker/podman only) |


### solace-util broker perform semp-login-check

Test an authenticated SEMP request against the broker

Runs an authenticated SEMP request from inside the broker and reports
whether the credentials were accepted.

Run this after rotating the admin password.

```
solace-util broker perform semp-login-check [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |


### solace-util broker perform shell-script

Run a host shell script inside the broker

Uploads a local shell script and runs it with bash inside the broker, as
the broker's own user.

Bash reports one exit status for the whole run, and the full output is
shown.

```
solace-util broker perform shell-script <file> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |


### solace-util broker remove

Remove the broker, prompting for each layer

Removes the deployed broker and its secrets on every platform.

This permanently deletes the broker; add --delete-data (with its own
confirmation) to also destroy its persistent data.

```
solace-util broker remove [flags]
```

Also available as: rm

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--delete-data` | `false` | delete the broker's persistent data too (Kubernetes PVCs / the container data directory). Without it the data is kept |
| `--no-prompt` | `false` | do not ask: answer yes to every question this command would ask. It does not RAISE a question -- without a --delete-* flag the expensive layer is still kept |


### solace-util broker restart

Bounce the broker's pod(s) or container

Restarts the broker's pod(s) gracefully (Kubernetes), or the container
(docker/podman).

With no --pod every pod is bounced in turn: monitor, then backup, then
primary.

This drops in-flight messaging on whatever it bounces.

```
solace-util broker restart [flags]
```

Also available as: rs

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--no-prompt` | `false` | do not ask: answer yes to every question this command would ask. It does not RAISE a question -- without a --delete-* flag the expensive layer is still kept |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |


### solace-util broker shell

Open an interactive shell in the broker

Interactive only; execs `bash` into the picked pod (Kubernetes) or this
host's container (docker/podman).

To run a script instead of typing at a prompt, use
`broker perform shell-script <file>`.

```
solace-util broker shell [flags]
```

Also available as: sh

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |


### solace-util broker start

Start a broker that is deployed but not running

Starts what `broker deploy` already created; it does not create anything
itself.

Kubernetes starts primary, then backup, then monitor, waiting for each
rollout.

```
solace-util broker start [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util broker status

Report the broker's deployment status

Reports pod/container state, plus the operator's CR conditions on
Kubernetes.

--all is how to find a broker whose env file you have lost.

```
solace-util broker status [flags]
```

Also available as: sts

| Flag | Default | Meaning |
| --- | --- | --- |
| `--all` | `false` | report every Solace broker found, not just the one this env file describes (Kubernetes: across all namespaces; docker/podman: every Solace container on this host) |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--detail` | `false` | include the static artifacts, not just the running ones (Kubernetes: secrets, configmaps and PVCs; docker/podman: mounts, which is also where secrets appear) |
| `--pod` | (none) | role to act on: primary (p), backup (b) or monitor (m). Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted (kubernetes only) |


### solace-util broker stop

Stop a running broker without removing it

Scales the broker to zero replicas, or stops the container, without
deleting anything.

This takes messaging down until `broker start` brings it back.

```
solace-util broker stop [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util broker validate

Check the env file and what this broker needs

Read-only, and safe to run against a system you don't want to disturb.

Reports every problem it finds in one pass, rather than stopping at the
first.

```
solace-util broker validate [flags]
```

Also available as: vld

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
| `--no-prompt` | `false` | answer yes to the --out overwrite question (no effect without --out) |
| `-o`, `--out` | (none) | write the artifact to this file instead of stdout |


### solace-util examples

Print a sample env file to start from

`examples` on its own prints the FULL annotated schema: every key the loader accepts,
the default each omitted one takes, and all three platform sections at once. That is
the file to read when you want to know what is configurable.

--platform writes a minimal STARTER instead: a standalone env file carrying only the
keys that platform cannot default, and declaring only its own section -- so the file it
writes needs no --platform of its own afterwards.

This is the one command that takes --platform without an env file to resolve it
against, because it is what produces the env file. The abbreviations work here too
(kube, dk, pm), expanded by the same parser every other --platform value goes through.

Nothing here reads an env file or contacts anything, so -e/--env is ignored and no
cluster or container runtime has to exist yet.

Every secret in the output is a CHANGE-ME placeholder: an unedited file is refused by
the broker rather than deployed with a password that is public in this repo. Each one
has a sibling *Env key naming a variable to read instead, which is what keeps a
committed env file secret-free.

  solace-util examples --platform docker -o env/dev.yaml
  solace-util examples | less
  solace-util eg --platform kube -o env/prod.yaml && solace-util validate -e env/prod.yaml

```
solace-util examples [flags]
```

Also available as: eg

| Flag | Default | Meaning |
| --- | --- | --- |
| `--no-prompt` | `false` | answer yes to the --out overwrite question (no effect without --out) |
| `-o`, `--out` | (none) | write the artifact to this file instead of stdout |


### solace-util operator

Install and operate the cluster-scoped EventBroker Operator (kubernetes only)

Installs and operates the operator, which is shared across every broker it
watches.

`operator deploy` adds this env file's namespaces to the watch list;
`operator remove` removes them, deleting the install only when nothing
else needs it.

```
solace-util operator
```

Runs nothing on its own. It names what `solace-util operator` can act on -- `deploy`, `generate`, `logs`, `remove`, `restart`, `start`, `status`, `stop`, `validate` -- and given a word it does not know it fails rather than reporting success.

Also available as: op

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.


### solace-util operator deploy

Install the operator, or add this env file's namespaces to it (kubernetes only)

Installs the operator, or unions this env file's namespaces into an
already-running one's watch list.

Widening the watch to all namespaces, or installing an older version,
needs an interactive confirmation with no unattended override.

```
solace-util operator deploy [flags]
```

Also available as: dp

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util operator generate

Render the operator install bundle (kubernetes only)

Renders everything `operator deploy` would apply, in apply order, without
applying it.

The watch list shown is only this env file's; a real deploy unions it
with whatever the running operator already watches.

```
solace-util operator generate [flags]
```

Also available as: gen

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

Renders to stdout and changes nothing: it runs no external command, so it needs no cluster or runtime, runs no preflight, and refuses `--allow-command` -- there is nothing here for it to approve.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--no-prompt` | `false` | answer yes to the --out overwrite question (no effect without --out) |
| `-o`, `--out` | (none) | write the artifact to this file instead of stdout |


### solace-util operator logs

Read the operator's controller logs (kubernetes only)

Runs `kubectl logs` against the controller Deployment, showing
reconciliation for every broker it watches, not just this env file's.

```
solace-util operator logs [flags]
```

Also available as: lg

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `-f`, `--follow` | `false` | keep streaming new lines instead of exiting |
| `--since` | (none) | read only lines newer than this duration (e.g. 30s, 5m, 2h) |
| `--tail` | (none) | print only this many trailing lines (a count, or `all`) |
| `--timestamps` | `false` | prefix each line with its timestamp |


### solace-util operator remove

Release this env file's claim on the operator (kubernetes only)

Removes this env file's namespaces from the operator's watch list,
deleting the operator itself only when nothing else is left watching.

--delete-crd additionally deletes every PubSubPlusEventBroker in the
cluster, and is refused while any broker resource still exists.

```
solace-util operator remove [flags]
```

Also available as: rm

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--delete-crd` | `false` | delete the operator's CustomResourceDefinitions too. Without it they are kept, so existing brokers survive |
| `--no-prompt` | `false` | do not ask: answer yes to every question this command would ask. It does not RAISE a question -- without a --delete-* flag the expensive layer is still kept |


### solace-util operator restart

Bounce the operator's controller (kubernetes only)

Runs a rollout restart on the operator's controller Deployment; it does
not change what is installed.

This briefly interrupts reconciliation for every broker the operator
watches.

```
solace-util operator restart [flags]
```

Also available as: rs

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util operator start

Start the operator's controller (kubernetes only)

Scales the controller Deployment back to one replica and waits for the
rollout.

Reconciliation resumes for every namespace the operator watches.

```
solace-util operator start [flags]
```

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util operator status

Report the operator's controller status (kubernetes only)

Reads the controller Deployment and its pods to report the running
version and the namespaces it watches.

```
solace-util operator status [flags]
```

Also available as: sts

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--detail` | `false` | include the full description of the operator deployment |


### solace-util operator stop

Stop the operator's controller, freezing reconciliation (kubernetes only)

Scales the controller Deployment to zero.

This freezes reconciliation for every broker in every watched namespace,
not just this env file's, until `operator start`.

```
solace-util operator stop [flags]
```

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |
| `--no-prompt` | `false` | do not ask: answer yes to every question this command would ask. It does not RAISE a question -- without a --delete-* flag the expensive layer is still kept |


### solace-util operator validate

Check the operator's install state and watch scope (kubernetes only)

Reports the operator's install state, running version versus this env
file's, and the namespaces it watches.

Every value is read live from the cluster, since the operator is shared
and may already watch namespaces this env file never named.

```
solace-util operator validate [flags]
```

Also available as: vld

Applies to: kubernetes. On any other platform this command fails rather than doing nothing.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--allow-command` | `[]` | approve one extra binary for the config's platform command, for this run only (repeatable; a bare name, never a path). The env file cannot grant this |


### solace-util validate

Check the whole env file and what it needs

Reports both the broker and operator halves of the env file together;
`broker validate` and `operator validate` are the same rows, scoped.

Read-only; a [WARN] (such as an operator not installed yet) does not
fail it, only a [FAIL] does.

```
solace-util validate [flags]
```

Also available as: vld

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


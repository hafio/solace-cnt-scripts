# Configuration

Every `solace-util` command reads one YAML env file. This document explains how that file is
found, how it decides which platform runs, and what every commonly-used key means.

**Do not assemble a file from the tables below -- start from one the binary writes:**

```
solace-util examples kubernetes -o env/dev.yaml   # or docker, or podman
solace-util examples full                         # every key, annotated
```

The three platform names give a minimal standalone file: only the keys that platform
cannot default, declaring only its own section. `full` is the fully annotated schema --
every key the loader accepts and the default each omitted one takes -- and is the same
text as [env/sample.yaml](../env/sample.yaml), which is generated from it. This document
explains the keys; those are the things you start from.

- [Choosing the env file](#choosing-the-env-file)
- [Which platform runs](#which-platform-runs)
- [The keys](#the-keys)
- [Scaling](#scaling)
- [Secrets](#secrets)
- [The command fields are executable content](#the-command-fields-are-executable-content)
- [Migrating from the bash env files](#migrating-from-the-bash-env-files-solace-util-convert)

## Choosing the env file

Every command reads one YAML env file, selected with `-e`/`--env`. The value is an actual
**file name**, taken literally -- no extension is ever inferred, so `-e dev` and
`-e dev.yaml` name different files:

- A **bare file name** is searched in the base directory, then in `<base-dir>/env` -- so
  `-e dev.yaml` finds `./dev.yaml` if it exists, otherwise `./env/dev.yaml`. The first hit
  wins, which means a copy in the base directory **shadows** the `env/` copy of the same
  name. Every run echoes the file it resolved to (`==> env file: ...`, on stderr), so the
  winner is never a surprise.
- `--base-dir` replaces the current directory for both lookups (default: current directory).
- A value carrying a **directory component** is used exactly as typed and is *not* retried
  under `env/` or joined with `--base-dir` -- e.g. `-e ./configs/prod.yaml`,
  `-e ../shared/prod.yaml`, or an absolute path.
- The default name is `env.yaml`, so a bare `solace-util check deploy` looks for
  `./env.yaml` then `./env/env.yaml`. Neither is shipped; write one with
  `solace-util examples <platform> -o env.yaml`.

When no candidate exists the error names every path that was tried.

Decoding is **strict**: an unknown or misspelled key is a hard error, so typos fail loud
instead of being silently ignored. A file that is not YAML at all is reported as such --
and if it looks like a legacy bash env file, the error points at `solace-util convert`
([below](#migrating-from-the-bash-env-files-solace-util-convert)).

## Which platform runs

The command tree is flat and identical on every platform -- there is no `kubernetes`,
`docker` or `podman` subtree to type. The platform is a property of the deployment the env
file already describes, so it is resolved from that file rather than repeated on the
command line. Resolution happens in this order:

1. **`--platform <name>`**, if given. It accepts the canonical names `kubernetes`, `docker`,
   `podman`, or the abbreviations `kube`, `dk`, `pm`. Neither `k8s` nor `k8` is accepted, as
   an abbreviation or as a section name -- the platform goes by the product's own word. It
   must name a platform section the env file actually declares -- passing `--platform docker`
   against a file with no `docker:` section fails loudly rather than running against a
   platform the file never described.
2. **Otherwise the env file decides**, from whichever of the top-level `kubernetes:`,
   `docker:`, and `podman:` sections it declares:
   - **exactly one** -> used silently, and named in the preamble
     (`==> platform: docker (from ./dev.yaml)`)
   - **none** -> a loud error telling you to add one
   - **more than one** -> an interactive prompt listing them; a non-interactive run
     (piped, CI) fails loudly instead and tells you to pass `--platform`

An env file **must** declare its platform section even when every setting under it
defaults -- write `docker: {}` rather than leaving the section out. This matters because
docker and podman have no mandatory field of their own, so without the marker the file
would be indistinguishable from a kubernetes one that simply hasn't set any docker keys.

The tree itself is the union of every platform's commands, and it renders the same way
regardless of which platform an env file names -- `--help` and shell completion never load
one, so a command cannot appear or disappear depending on a file they have not read. A
command that does not apply to the platform a file resolves to says so in its help text
(for example "(kubernetes only)") and refuses at run time with a named error rather than
silently doing nothing.

## The keys

Minimum required (Kubernetes):

| Key | Purpose |
| --- | --- |
| `image.repo` | Broker image repository |
| `image.tag` | Image tag |
| `admin.pass` | Broker admin password (never defaulted) |
| `kubernetes.name` | Broker / custom-resource name |
| `kubernetes.namespace` | Target namespace |
| `kubernetes.storage.msgNode` | Message-node PVC size (e.g. `30Gi`) |

Common optional knobs:

| Key | Default | Purpose |
| --- | --- | --- |
| `redundancy` | `no` | `yes` = HA group (primary+backup+monitor); `no` = single standalone broker. HA provisions three brokers, so it must be asked for explicitly |
| `image.registry` | docker.io | Registry prefix for the image reference |
| `kubernetes.storage.class` | cluster default | StorageClass for the broker PVCs |
| `kubernetes.updateStrategy` | `automatedRolling` | `automatedRolling` or `manualPodRestart` |
| `kubernetes.runtime` | `kubectl` | Cluster CLI (legacy `KUBE`). A scalar is split on whitespace, so it can be a drop-in (`oc`) or a profile (`kubectl --kubeconfig <file>`). **Restricted** -- see [The command fields are executable content](#the-command-fields-are-executable-content) |
| `docker.runtime` / `podman.runtime` | `docker` / `podman` | Container CLI (legacy `CONTAINER_RUNTIME`), same forms and the same restrictions as `kubernetes.runtime` |
| `docker.compose` | `<runtime> compose` | The compose invocation. Set it to `docker-compose` on a host carrying only the standalone v1 binary; same forms and restrictions as `runtime`, plus the one permitted `compose` subcommand |
| `<docker\|podman>.container.healthCheck.enabled` | `false` | Adds an engine health check polling the broker's own `/health-check/readiness` on port 5550 every 5s, so `docker ps` and podman's auto-restart see readiness rather than liveness. Needs broker **10.26 or later** and a version-numbered `image.tag`; set `healthCheck.cmd` to supply your own probe instead (which skips the version check). Container-only by design -- on Kubernetes the operator already probes the pods |
| `kubernetes.tlsServerSecret` | -- | Name of the TLS Secret built from `tls.cert`/`tls.certKey`; its presence enables the CR's TLS block. Lives under `kubernetes.*` because it names a Kubernetes Secret object -- the cert/key files themselves stay platform-neutral under `tls.*` |
| `kubernetes.imagePullSecret` | -- | Name of the image-pull Secret built from `image.user`/`image.pass`; its presence enables the CR's `pullSecrets` block and the operator's `regcred`. The registry credentials themselves stay under `image.*` (docker/podman use them for `<runtime> login`) |
| `kubernetes.imagePullPolicy` | -- | `Always` \| `IfNotPresent` \| `Never`; unset keeps the CR's own `IfNotPresent` |
| `kubernetes.adminSecret` | `solace-admin-secret` | Name of the Kubernetes Secret holding the admin/monitor credentials. |
| `kubernetes.operator.namespace` | `pubsubplus-operator-system` | Namespace the cluster-scoped EventBroker Operator is installed to and addressed in. Two rules only, and neither one asks the cluster: use this when set, otherwise the fixed default that `deploy operator` installs to -- so `deploy operator`, `remove operator` and every other operator command always resolve the SAME namespace. It used to be discovered by listing every namespace's Deployments and taking the first one whose name merely CONTAINED the operator's, an unanchored match with no uniqueness check that could resolve to another team's operator on a cluster running two installs; that search is gone. Stays optional -- most deployments never set it |
| `admin.additionalUsers` | -- | Extra CLI (management) users, each `{username, accessLevel, password\|passwordEnv}` with `accessLevel` one of `none`, `read-only`, `mesh-manager`, `read-write`, `admin`. Created at boot on containers, and by `config apply additional-users` on Kubernetes -- see [Extra CLI users differ by platform](operations.md#extra-cli-users-differ-by-platform) |
| `admin.user` | `admin` | Broker admin username. **docker/podman only** -- it names the container's `username_<user>_globalaccesslevel` setting, its mounted password file and the SEMP login. On Kubernetes the operator reads the fixed `username_admin_password` key out of `kubernetes.adminSecret` and creates the user itself, so the admin user is always `admin` there and any other value is a load-time error rather than a silently ignored key |
| `admin.passEnv` (and every other `*Env`) | -- | Name of an environment variable holding the secret, instead of the value itself. See [Secrets](#secrets) |
| `timezone` | -- | Broker timezone, all platforms (the CR's `timezone` and the containers' `TZ`). Omitted keeps the image default |
| `broker.cliScriptsFolder` / `broker.diagDir` / `broker.productKeys` / `broker.domainCerts` | -- | Platform-neutral: local folder for `cli --input` scripts, local folder for `diagnostics` output, the list `config apply product-keys` applies, and the CA files `config apply domain-certs` loads. Every platform runs these same post-deployment steps identically, which is why the section sits at the top level rather than under `kubernetes.*` |
| `kubernetes.securityContext` | -- | `runAsUser`/`fsGroup` for the pod. Omitted entirely when unset |
| `kubernetes.containerSecurity` | -- | `runAsUser`/`runAsGroup`/`readOnlyRootFilesystem` for the broker container |
| `scaling.*` | see [Scaling](#scaling) | Broker sizing, applied on every platform -- the CR's `spec.systemScaling` on Kubernetes, container environment variables on docker and podman |
| `scaling.maxConnections` | `100` (Kubernetes) / `1000` (container) | The Solace scaling tier. Fixes the broker's CPU and defaults its memory on every platform -- see [Scaling tiers](#scaling-tiers) |
| `<docker\|podman>.container.mem` | the tier's memory | Container memory limit, in docker's and podman's own `b\|k\|m\|g` suffix (not Kubernetes' `Mi`/`Gi`). There is no matching cpu key: CPU is fixed by the tier |

## Scaling

Every key under `scaling` applies to every platform. Only the delivery differs: Kubernetes
writes them into the broker CR's `spec.systemScaling`, while docker and podman pass them to
the container as environment variables under the broker's own setting names. One env file
therefore sizes the same broker whichever platform runs it.

| `scaling` key | Broker setting (container env var / CR field) |
| --- | --- |
| `maxConnections` | `system_scaling_maxconnectioncount` |
| `maxQueueMessages` | `system_scaling_maxqueuemessagecount` |
| `maxKafkaBridge` | `system_scaling_maxkafkabridgecount` |
| `maxKafkaConnections` | `system_scaling_maxkafkabrokerconnectioncount` |
| `maxBridges` | `system_scaling_maxbridgecount` |
| `maxSubscriptions` | `system_scaling_maxsubscriptioncount` |
| `maxGuaranteedMsgMB` | `system_scaling_maxguaranteedmessagesize` |
| `maxSpoolUsageMB` | `messagespool_maxspoolusage` (the CR spells it `maxSpoolUsage`) |

Defaults are identical across platforms except `maxConnections` (100 on Kubernetes, 1000 on
containers) and `maxSpoolUsageMB` (10000 on Kubernetes, 100000 on containers).

### Scaling tiers

`scaling.maxConnections` is the Solace scaling tier, and it decides the broker's CPU on all
three platforms. CPU is **not** configurable: sizing a broker by connection count and then
sizing its CPU independently is how a 200k-connection broker ends up on two cores. Memory is
the tier's default and stays yours to override; storage is untouched by the tier.

| `scaling.maxConnections` | CPU cores (fixed) | Memory default (Kubernetes / container) |
| --- | --- | --- |
| `100` (Kubernetes default) | 2 | `3410Mi` / `3410m` |
| `1000` (container default) | 2 | `6898Mi` / `6898m` |
| `10000` | 4 | `12435Mi` / `12435m` |
| `100000` | 8 | `30925Mi` / `30925m` |
| `200000` | 12 | `52581Mi` / `52581m` |

The value must be **exactly** one of those five. A value between tiers is rejected rather
than rounded, because Solace publishes no sizing for it. Override memory with
`kubernetes.msgNode.mem` (a Kubernetes quantity, `Mi`/`Gi`) or `<docker|podman>.container.mem`
(docker's and podman's own `b|k|m|g` suffix -- the engines reject `Mi`, so the two spellings
are not interchangeable and the loader says so).

Docker and podman carry the tier's CPU and memory caps in the generated compose file
(`cpus:`, `mem_limit:`) and quadlet unit (`PodmanArgs=--cpus=`, `Memory=`), so an existing
container deployment needs a full **redeploy** -- not just a restart -- to pick up a changed
cap. In an HA group the monitor host gets the same caps as the messaging hosts; these are
ceilings rather than reservations, so an oversized monitor limit costs nothing.

A rootless podman host also has a file-descriptor ceiling the tier cannot raise on its own --
see [File descriptors on rootless podman](operations.md#file-descriptors-on-rootless-podman).

## Secrets

Every secret field takes either the value itself or -- through a sibling `*Env` key --
the **name of an environment variable** to read it from. Setting both is an error, and so
is naming a variable that is unset or empty: the load fails naming the key and the
variable rather than deploying a broker with a blank password.

| Value key | Reference key |
| --- | --- |
| `admin.pass` | `admin.passEnv` |
| `admin.monitorPass` | `admin.monitorPassEnv` |
| `admin.additionalUsers[].password` | `admin.additionalUsers[].passwordEnv` |
| `tls.certPassphrase` | `tls.certPassphraseEnv` |
| `image.pass` | `image.passEnv` |
| `nodes.psk` | `nodes.pskEnv` |

```yaml
admin:
  passEnv: SOLACE_ADMIN_PASS     # export SOLACE_ADMIN_PASS before any command
```

With the `*Env` form the env file carries no secret and is safe to commit and share. A
value is otherwise used **verbatim** on every platform -- a `$VAR` or `${VAR}` inside one
is a literal password, never expanded. `nodes.pskEnv` also opts out of PSK generation:
`prepare host` only generates a key when the literal `nodes.psk` is empty, so with the
reference form create it yourself (`openssl rand -base64 60`) and export the same value on
all three hosts.

The tool never echoes a secret. Values piped to a command on stdin show as
`<<< (N bytes on stdin)` under `-v/--verbose`, values passed to a child process's environment
as `NAME=***`, and `check deploy`/`status broker` report only whether each one is set. The
one exception is explicit: `generate secrets` prints the values themselves, because printing
them is what that command is for -- `generate secrets broker` for the broker's, `generate
secrets operator` for the operator's image-pull credential. See
[Rendering without applying](operations.md#rendering-without-applying) for where each secret
lands at rest.

## The command fields are executable content

`kubernetes.runtime`, `docker.runtime`, `podman.runtime` and `docker.compose` each name a binary
this tool runs **on your machine**. Env files travel -- repositories, pull requests, shared
archives -- so the person who wrote one is routinely not the person who runs it. Treat an
env file the way you would treat a script someone sent you: **read the command fields before
running anything with it.**

To make that review short, the fields are restricted. A command is accepted only when:

1. **Every token is inert and visible.** No control characters, no whitespace inside a single
   argument (any Unicode whitespace, not just the ASCII space), no invisible formatting
   characters (zero-width spaces and joiners, bidirectional overrides), no quotes, no
   backslash, no backtick, and none of `$ ; | & < > ( ) * ? [ ] { } ~ # !`. Nothing is ever
   passed through a shell, so these are not injections -- but tokens end up in logs and in
   the `-v/--verbose` exec trace, and a token you cannot see is one you cannot
   review. A Windows path in a flag value therefore needs forward slashes:
   `--kubeconfig C:/Users/you/.kube/config`.
2. **The binary is a bare name from the allowlist.** No `/` or `\` anywhere in it: a path
   would run a file the env file chose -- such as a `./kubectl` unpacked beside it -- rather
   than the one on your `PATH`. One optional `.exe` is stripped, then the name must be:

   | Platform | Allowed |
   | --- | --- |
   | Kubernetes | `kubectl`, `oc` |
   | Docker | `docker`, `docker-compose`, `nerdctl` |
   | Podman | `podman` |

3. **Nothing after it is a bare word.** Flags and their values are fine
   (`kubectl --context prod -n solace`); a bare word is not, because this tool appends its
   own subcommand and a word in that position would run ahead of it. `kubectl delete` in a
   config is exactly the attack. The literal `--` is refused for the same reason.

   **One acknowledged gap, in that third rule.** The check cannot know how many values a
   flag takes -- that would mean carrying a table of every flag of every allowed CLI, which
   would rot as those CLIs change -- so the token after any flag is accepted as that flag's
   value. After a flag that takes *no* value, that token is not a value at all, and it lands
   in subcommand position after all:

   ```yaml
   kubernetes.runtime: kubectl --insecure-skip-tls-verify delete   # accepted; runs `delete`
   docker.compose: docker-compose --verbose down                   # the container equivalent
   ```

   This is a known limit, not an oversight, and it is why the guarantee is stated narrowly:
   **argv[0] and every bare token are checked; the contents of a flag value are not.** It
   costs an attacker nothing more than the access they already need -- anyone who can edit
   the env file can also point a perfectly legitimate `kubectl` at your production cluster,
   which no check in this file can detect. Both are review problems: read an env file's
   command fields the way you would read a script it ships.

Anything else -- a wrapper such as `microk8s kubectl` or `lima nerdctl`, a site-specific
shim -- runs only when **you** approve it, per invocation:

```sh
solace-util deploy broker --allow-command microk8s   # kubernetes env file wrapping kubectl in microk8s
solace-util deploy all --allow-command lima          # docker/podman env file wrapping the runtime in lima
```

`--allow-command` is repeatable, takes a bare name (never a path), and exists **only** as a
command-line flag. There is deliberately no env-file key, environment variable, or any other
way for a config to widen its own allowlist: the authority to run something unusual belongs
to the person who can see what they are approving. It is rejected on any `generate` command,
where nothing executes.

**Privilege escalation is never approvable**, by the config or by you: `sudo`, `doas`, `su`,
`pkexec`, `run0`, `runas` and `gsudo` are refused as `--allow-command` values, in any casing
(`Sudo` and `SUDO.exe` too -- Windows and a default macOS filesystem resolve those to the same
binary, so matching them exactly would have let a capital letter through a floor that is
supposed to stop everyone). The allowlist in rule 2 above is the opposite: it stays
case-sensitive on purpose, because `KUBECTL` is a genuinely different file where filesystems
say it is, and a positive match that folded case would approve a binary nobody listed. This is
not a
ban on running as root -- rootful podman needs it. It is about *where* you elevate. A
`runtime: sudo podman` elevates every command this tool issues, for the whole life of an env
file, decided by whoever wrote that file. Elevate the tool instead, at the moment you run it,
so the privilege belongs to one invocation you chose:

```sh
sudo solace-util deploy all -e prod.yaml   # yes (prod.yaml is a podman env file)
# runtime: sudo podman  in the env file      # never
```

The same check runs twice -- once when the env file is loaded, and again immediately before
any command line is built -- from a single implementation, so a hostile file is inert even
on a path that skipped validation.

**What this does not protect against.** Two things are out of scope, and no amount of
parsing would fix either:

- **A compromised machine.** If an attacker has already put a trojan `kubectl` on your
  `PATH`, they own the host; nothing this tool checks can help. What it does do is make the
  binary's real location visible -- before any work starts, each binary this env file names
  (`kubernetes.runtime`, `docker.runtime`, `podman.runtime`, `docker.compose`) is resolved and
  printed as `==> using <name>: <resolved path>`, and `-v/--verbose` prints every command as
  it runs -- and refuse to resolve a bare name from the current directory.
- **Config that is malicious but perfectly legitimate in form.** `kubernetes.namespace: production`,
  or a valid `kubectl --context` aimed at the wrong cluster, is a review problem. So is a
  flag's *value*: this tool cannot know how many arguments a flag takes, so the token after
  `--kubeconfig` is accepted as that flag's value whatever it says. The hard guarantee covers
  the binary and every bare word -- not flag values.

Rendering executes nothing: every `generate` command only ever calls the templating package,
never an external command, so pointing one at an env file you did not write cannot run
anything. Note that it still *loads* the file, so one whose command field breaks the rules
above fails there rather than printing an artifact -- which is itself the answer you wanted
about that file. To read a command field without loading anything at all, open the file.

## Migrating from the bash env files (`solace-util convert`)

The pre-Go scripts kept their configuration in shell files under `bash/env/`, sourced by
`000-env.sh`. `solace-util convert` turns one into the YAML this CLI reads:

```
solace-util convert bash/env/prod -o prod.yaml                 # kubernetes flavour
solace-util convert bash/docker-podman/env/prod -o prod.yaml   # docker/podman flavour
solace-util check deploy -e prod.yaml
```

- The **platform section** is detected from the variables present (`SOLBK_NS`/`SOLOP_*` ->
  `kubernetes`, `SOLBK_NODE_*`/`DOCKER_MODE`/`PODMAN_ROOTLESS` -> `docker`/`podman`). Pass
  `--platform kubernetes|docker|podman` to choose it yourself; the choice is echoed either way.
- The source is read as an **env file, not a shell script**: one assignment per line
  (scalars, `( ... )` arrays, `declare -A` maps, `export`/`declare` prefixes, `${VAR}`
  references, and trailing comments are all understood). Shell constructs beyond
  assignments are skipped.
- Only what the env file **actually set** is written. Values the bash bootstrap defaulted
  are left out, so the Go defaults apply instead.
- A variable with no YAML equivalent is **named on stderr**, never dropped silently. So are
  a non-numeric value for a numeric field and an unrecognised `SOLBK_REDUNDANCY`.
- The converted file is re-read and validated, so a source env that was already missing
  mandatory values says so at conversion time.
- Without `-o` the YAML goes to stdout (warnings stay on stderr). With `-o` the file is
  written `0600` and an existing file is **not** overwritten unless you pass `--force`.

The output carries every secret from the source file verbatim -- treat it like the source,
and never commit it. (Switch the values to their `*Env` reference keys afterwards and it
becomes safe to commit; see [Secrets](#secrets).) `SOLBK_USR_SECRET` converts to
`kubernetes.adminSecret` (`SOLBK_ADM_SECRET` is accepted as an alias for the same key;
when both are set and disagree, the canonical `SOLBK_USR_SECRET` wins with a warning),
`SOLBK_SVR_SECRET` to `kubernetes.tlsServerSecret`, `IMAGEREPO_SECRET` to
`kubernetes.imagePullSecret` (on a docker/podman conversion those two are dropped with a
warning naming that kubernetes-only home -- they name Kubernetes Secret objects, which have
no container equivalent), and each `SOLBK_USR_PASS` entry to an `admin.additionalUsers` entry with
`accessLevel: none` -- the bash flow set no level, so the converter picks the least
privileged one and says so; raise it per user as needed.

**`kubernetes.msgNode.cpu` and `scaling.maxPool` fail to load, each naming its
replacement.** Broker CPU is fixed by the scaling tier rather than set by hand, and
`maxPool` would name the same broker setting as `scaling.maxSpoolUsageMB` -- one concept
under two platform-specific keys, which the scaling block keeps under a single name.
`kubernetes.msgNode.mem` is unaffected. Docker always deploys through compose, so there is
no `docker.mode` key to choose a mode with; a file still carrying it fails strict decoding
as an unknown field. See [Scaling](#scaling).

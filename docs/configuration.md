# Configuration

Every `solace-util` command reads one YAML env file. This document explains how that file is
found, how it decides which platform runs, and what every commonly-used key means.

**Do not assemble a file from the tables below -- start from one the binary writes:**

```
solace-util examples --platform kubernetes -o env/dev.yaml   # or docker, or podman
solace-util examples                              # every key, annotated
```

The three platform names give a minimal standalone file: only the keys that platform
cannot default, declaring only its own section. `full` is the fully annotated schema --
every key the loader accepts and the default each omitted one takes -- and is the same
text as [env/sample.yaml](../env/sample.yaml), which is generated from it. This document
explains the keys; those are the things you start from.

- [Choosing the env file](#choosing-the-env-file)
  - [Relative paths resolve against the env file, not the current directory](#relative-paths-resolve-against-the-env-file-not-the-current-directory)
- [Which platform runs](#which-platform-runs)
- [The keys](#the-keys)
- [Scaling](#scaling)
- [Secrets](#secrets)
- [Replication](#replication)
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
- The default name is `env.yaml`, so a bare `solace-util validate` looks for
  `./env.yaml` then `./env/env.yaml`. Neither is shipped; write one with
  `solace-util examples --platform <platform> -o env.yaml`.

When no candidate exists the error names every path that was tried.

Decoding is **strict**: an unknown or misspelled key is a hard error, so typos fail loud
instead of being silently ignored. A file that is not YAML at all is reported as such --
and if it looks like a legacy bash env file, the error points at `solace-util convert`
([below](#migrating-from-the-bash-env-files-solace-util-convert)).

### Relative paths resolve against the env file, not the current directory

Every **host path** in the file -- `tls.cert`, `tls.certKey`, `tls.cas`,
`broker.cliScriptsDir`, `broker.hostDiagnosticDir`, `broker.domainCerts.dirs[].path`,
`broker.domainCerts.files{}`, `docker.composeFile` -- is resolved against **the directory
the env file itself was found in**, never against the directory the command was run from.
So `/srv/solace/env/prod.yaml` declaring `tls.cert: certs/tls.crt` means
`/srv/solace/env/certs/tls.crt`, whichever directory you drive it from. An absolute value
is left exactly as written.

The reason is not tidiness. A **podman** `Volume=` source with no leading separator is read
by podman as the name of a *named volume*, so a relative certificate path made podman
create an empty volume and mount it over the real certificate -- no error, unit starts,
broker starts, and TLS is not what you configured. systemd gives a unit no useful working
directory, so there was never a current directory that would have made the relative form
work.

Three fields deliberately do **not** follow the rule:

| Field | Why |
| --- | --- |
| `<platform>.container.dataDir` | **Required absolute** instead. It is the host side of a bind mount *and* what `broker remove --delete-data` deletes recursively, so quietly changing which directory that points at is not a fix. The default `/opt/solace/data` already satisfies it |
| `podman.quadletDir` | Checked, never resolved. The unit must live where systemd scans, so resolving a relative value would invent a location systemd never reads. Both defaults are already absolute |
| `podman.baseDir` | **Mandatory and required absolute.** It is a `Volume=` source too, and it receives a file containing a private key, so resolving a relative value would invent a location for that key which you never named |

Host paths are also **character-checked**, and the rule is deliberately looser than the one
for the [`command` fields](#the-command-fields-are-executable-content): a path may carry a
backslash and a colon, because `C:\certs\tls.crt` is a real path and a command token has no
business carrying either. (A tilde is not part of that difference: both sides admit an
embedded one -- an 8.3 short name such as `C:\Users\RUNNER~1\...` -- and both expand a leading
one, see below.) What
a path may **not** carry is a `$` or a shell metacharacter (both container artifacts write
these values verbatim, and compose interpolates `$` across the whole document), whitespace
(a mount is written `source:target:options` on one line, so a space cannot be delimited), or
a control or invisible character. The check runs on the value as you wrote it, before any
resolution.

**A leading `~` expands to your own home directory, in every path key and in every `command`
field.** `~`, or `~/...`, or `~\...`, expands to `os.UserHomeDir()` -- the home directory of
whoever is **running this tool** -- before the relative-path rule above ever sees it, so the
expanded value is already absolute and is never joined onto the env file's directory. A tilde
anywhere else in the value is left alone, which is what keeps an 8.3 short name like
`C:\Users\RUNNER~1\...` working. `~someoneelse/...` (another user's home) is not supported and
is refused by name rather than guessed at. An unresolvable home directory is an error naming
the field, and so is a home directory whose own path carries whitespace or a character the
value's destination cannot carry.

There are no exceptions among the path keys -- the three in the table above included. Those
three are still never *resolved* against the env file's directory, so a relative value is
still refused; a `~` one is not relative by the time the check runs. What each is checked for
does not change:

```yaml
podman:
  quadletDir: ~/.config/containers/systemd
  baseDir: ~/solace
  container:
    dataDir: ~/solace/data
```

In a `command` field every token expands **except the first**, which names the binary and must
stay a bare allowlisted name (`kubectl`, `oc`, `docker`, `podman`, ...) with no path in it at
all -- see [the `command` fields](#the-command-fields-are-executable-content). So this works:

```yaml
kubernetes:
  command: oc --kubeconfig ~/solace/kubecontext
```

and `command: ~/bin/oc ...` is refused, naming argv[0] as the reason.

A `~` value is the one thing in this schema that does **not travel**. Every other host path
either is absolute or resolves against the env file's own directory, so the same file
describes the same deployment on any machine; a `~` resolves against whoever runs the
tool, so an env file carrying one means something different for each operator -- and if their
home cannot be resolved, or carries a space, it means a failed load. Write it out
in full in any file you share.

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
| `semp.adminPass` | Broker admin password (never defaulted). The username is always `admin` -- the broker's own name for the built-in account, and there is no key to change it |
| `kubernetes.name` | Broker / custom-resource name |
| `kubernetes.namespace` | Target namespace |
| `kubernetes.storage.msgNodeSize` | Message-node PVC size (e.g. `30Gi`). Mandatory unless `customVolumeMount` covers every node, which leaves nothing to provision |
| `kubernetes.storage.customVolumeMount.<primary\|backup\|monitor>` | Mount an EXISTING PersistentVolumeClaim for that node instead of provisioning one. Mutually exclusive with `kubernetes.storage.class` -- naming both is refused, since the CRD does not say which wins. All nodes in the redundancy group or none. **`broker remove --delete-data` never deletes these**: the volume may hold data that predates this broker, so it is reported and left for you |

Common optional knobs:

| Key | Default | Purpose |
| --- | --- | --- |
| `redundancy.<primary\|backup\|monitor>.name` | this host's hostname (standalone containers only) | The broker's **routername**, and the container's hostname. On docker/podman in HA all three are **mandatory**: each is the key of that node's entry in the group table every host renders (`redundancy_group_node_<name>_connectvia`), and a host knows its own name and no other machine's, so one it filled in itself would build a table the other two disagree with. In standalone it is optional -- there is one node, it is always this host, so the host's own OS hostname is used when the key is omitted, and the run says so. Not read on Kubernetes, where the operator names the pods after `kubernetes.name` |
| `redundancy.enabled` | `false` | `true` = HA group (primary+backup+monitor); `false` = single standalone broker. HA provisions three brokers, so it must be asked for explicitly. The group's members live under `redundancy.<primary\|backup\|monitor>.<name\|addr>` and the key they authenticate with under `redundancy.psk` |
| `replication.*` | -- | A DR pair: two SEPARATE brokers, each message-VPN active at one site and standby at the other. Not `redundancy`, which is the three nodes of one HA group -- a replicated deployment usually has both, an HA group at each site. Omit the section entirely unless this broker replicates; see [Replication](#replication) |
| `image.registry` | docker.io | Registry prefix for the image reference |
| `kubernetes.storage.class` | cluster default | StorageClass for the broker PVCs |
| `kubernetes.updateStrategy` | `automatedRolling` | `automatedRolling` or `manualPodRestart` |
| `kubernetes.command` | `kubectl` | Cluster CLI (legacy `KUBE`). A scalar is split on whitespace, so it can be a drop-in (`oc`) or a profile (`kubectl --kubeconfig <file>`). **Restricted** -- see [The command fields are executable content](#the-command-fields-are-executable-content) |
| `docker.command` / `podman.command` | `docker` / `podman` | Container CLI (legacy `CONTAINER_RUNTIME`), same forms and the same restrictions as `kubernetes.command` |
| `docker.compose` | `<command> compose` | The compose invocation. Set it to `docker-compose` on a host carrying only the standalone v1 binary; same forms and restrictions as `docker.command`, plus the one permitted `compose` subcommand |
| `<docker\|podman>.container.healthCheck.enabled` | `false` | Adds an engine health check polling the broker's own `/health-check/readiness` on port 5550 every 5s, so `docker ps` and podman's auto-restart see readiness rather than liveness. Needs broker **10.26 or later** and a version-numbered `image.tag`; set `healthCheck.cmd` to supply your own probe instead (which skips the version check). Container-only by design -- on Kubernetes the operator already probes the pods |
| `kubernetes.tlsServerSecret` | -- | Name of the TLS Secret the broker uses; its presence enables the CR's TLS block. Lives under `kubernetes.*` because it names a Kubernetes Secret object -- the cert/key files themselves stay platform-neutral under `tls.*`. **Naming it does not mean building it**: with `tls.cert`/`tls.certKey` set, this tool builds the Secret and removes it on teardown; without them the Secret must already exist and is only referenced -- see [Bring your own TLS Secret](#bring-your-own-tls-secret) |
| `tls.cert` / `tls.certKey` | -- | The server certificate and its private key, as two separate host files. **Inseparable on every platform**: setting one without the other is refused at load, in both directions. On docker and podman the broker reads the certificate as ONE file containing the key followed by the certificate, and this tool builds that file from the two halves. Kubernetes takes them as two keys in a Secret and lets the operator assemble them -- a Secret carrying only `tls.crt` is one the broker cannot start a listener over |
| `tls.cas` | -- | Trusted CA files, applied by `broker configure domain-certs` into the broker's own trust store. They are **not** part of the server certificate and are not mounted into the container or the pod |
| `<docker\|podman>.container.name` | `solace` | The container's name, and the stem of every derived name (the podman unit and service, the host-side secret names). Held to the engines' own grammar: it must start with a letter or digit, then letters, digits, `.`, `_` or `-`. A name that YAML would read as a boolean or number (`yes`, `off`, `0123`) is legal here and quoted in the generated compose file, so it stays the string you wrote. On docker it is also the compose **project** name, lowercased with anything outside `[a-z0-9_-]` folded to `-`, since compose's grammar is narrower than the engines' -- override with `COMPOSE_PROJECT_NAME` ([operations.md](operations.md#docker-and-podman-mechanics)) |
| `podman.baseDir` | -- | **Mandatory on podman**, absolute. Host directory for files this tool writes for podman: today the server-certificate bundle, which contains the private key, written `0600` in a `0700` directory. Mandatory rather than defaulted because where a private key lands on your host is your decision. Kept separate from `quadletDir`, since the unit must live where systemd scans. Removed by `broker remove`, not by `--delete-data`. Docker needs no equivalent: a compose file can inline what a quadlet unit cannot, so docker's bundle never touches the host |
| `kubernetes.imagePullSecret` | -- | Name of the image-pull Secret the CR references. **Optional**: unset derives `<kubernetes.name>-image-pull` when registry credentials are configured, and names nothing (no `pullSecrets` block) when they are not. **Naming it does not mean building it** -- the same rule `kubernetes.tlsServerSecret` follows just above: with `image.user`/`image.pass` set (or their `*Env` equivalents, resolved into them at load), this tool builds the Secret under this name or the derived default, applies it alongside the operator's own fixed-name `regcred` Secret, and removes both on teardown; without those credentials, a configured name points at a Secret that must already exist -- created by hand or by a cluster admin -- and is only referenced, never built or deleted. The registry credentials themselves stay under `image.*` (docker/podman use them for `<command> login`, which has no operator and no `regcred` of its own) |
| `kubernetes.imagePullPolicy` | -- | `Always` \| `IfNotPresent` \| `Never`; unset keeps the CR's own `IfNotPresent` |
| `kubernetes.adminSecret` | `solace-admin-secret` | Name of the Kubernetes Secret holding the admin/monitor credentials. |
| `kubernetes.operator.namespace` | `pubsubplus-operator-system` | Namespace the cluster-scoped EventBroker Operator is installed to and addressed in. Two rules only, and neither one asks the cluster: use this when set, otherwise the fixed default that `operator deploy` installs to -- so `operator deploy`, `operator remove` and every other operator command always resolve the SAME namespace. It used to be discovered by listing every namespace's Deployments and taking the first one whose name merely CONTAINED the operator's, an unanchored match with no uniqueness check that could resolve to another team's operator on a cluster running two installs; that search is gone. Stays optional -- most deployments never set it |
| `semp.additionalUsers` | -- | Extra CLI (management) users, each `{username, accessLevel, password\|passwordEnv}` with `accessLevel` one of `none`, `read-only`, `mesh-manager`, `read-write`, `admin`. Created at boot on every platform. The username must start with a letter or `_` and be 1-32 characters (the broker's own rule); on Kubernetes it may not contain `.` or `-` either, because the credentials ride the pod environment there and the kubelet drops variables whose names are not identifiers. See [Extra CLI users differ by platform](operations.md#extra-cli-users-differ-by-platform) |
| `semp.adminPassEnv` (and every other `*Env`) | -- | Name of an environment variable holding the secret, instead of the value itself. See [Secrets](#secrets) |
| `timezone` | -- | Broker timezone, all platforms (the CR's `timezone` and the containers' `TZ`). Omitted keeps the image default |
| `broker.cliScriptsDir` / `broker.hostDiagnosticDir` / `broker.productKeys` / `broker.domainCerts` | `cli` / `diag-configs` / -- / -- | Platform-neutral: host folder for `broker perform cli-script` scripts, host folder for `diagnostics` output, the list `broker configure product-keys` applies, and the CA certificates `broker configure domain-certs` loads (`dirs`, a list of directories walked one level deep, plus `files`, explicit `CA-NAME: full host path` entries -- see [operations.md](operations.md#post-deployment-configuration-order)). Every platform runs these same post-deployment steps identically, which is why the section sits at the top level rather than under `kubernetes.*`. The first two are defaulted on every platform; `domainCerts.dirs`/`files` are **not** defaulted, so an env file that configures none is a no-op. A `files` key is checked at LOAD against the same charset and 64-character cap a directory-derived name satisfies by construction, and an entry with no path is refused there too -- the directory WALK still waits for `broker configure domain-certs`, since a directory that does not exist on this machine must not fail a `deploy` that never touches certificates. All of them are **host** paths and resolve against the env file's directory ([above](#relative-paths-resolve-against-the-env-file-not-the-current-directory)) |
| `kubernetes.securityContext` | -- | `runAsUser`/`fsGroup` for the pod. Omitted entirely when unset |
| `kubernetes.containerSecurity` | -- | `runAsUser`/`runAsGroup`/`readOnlyRootFilesystem` for the broker container |
| `scaling.*` | see [Scaling](#scaling) | Broker sizing, applied on every platform -- the CR's `spec.systemScaling` on Kubernetes, container environment variables on docker and podman |
| `scaling.maxConnections` | `100` (Kubernetes) / `1000` (container) | The Solace scaling tier. Fixes the broker's CPU and defaults its memory on every platform -- see [Scaling tiers](#scaling-tiers) |
| `<docker\|podman>.container.mem` | the tier's memory | Container memory limit, in docker's and podman's own `b\|k\|m\|g` suffix (not Kubernetes' `Mi`/`Gi`). There is no matching cpu key: CPU is fixed by the tier |


## Bring your own TLS Secret

`kubernetes.tlsServerSecret` and `tls.cert`/`tls.certKey` answer different questions --
what the Secret is CALLED, and what it is built FROM -- and only the first is always this
tool's business. Which of them you set decides who owns the Secret:

| `tls.cert` + `tls.certKey` | What happens |
| --- | --- |
| set | This tool builds the `kubernetes.io/tls` Secret from those files, `broker generate` prints it ahead of the CR, `broker deploy` applies it, `broker configure server-certs` rotates it, and `broker remove` deletes it |
| unset | The Secret must already exist -- created by hand, by cert-manager, or by anything else. The CR references it by name and nothing here reads, applies, rotates or deletes it. `broker remove` leaves it alone, and the namespace gate counts it as someone else's |

Set the pair or neither: one without the other is refused, because the Secret carries both
keys and a Secret with only `tls.crt` in it is one the broker cannot start a listener over.
Supplying the files without naming the Secret is refused too -- the Secret would have no
name and the CR no `tls` block, so the certificate would be silently unused.

`broker validate` states which of the two it is, so a missing Secret is not first
discovered by a pod that will not mount.

**`tls.certPassphrase` is docker/podman only.** The CRD's `spec.tls` carries only
`serverTlsConfigSecret`, `certFilename`, `certKeyFilename` and `enabled` -- there is no
passphrase field and no Secret key the operator reads one from. On Kubernetes, supply an
unencrypted key or decrypt it into the Secret yourself; `broker validate` warns when the
key is set.


## Scaling

Every key under `scaling` applies to every platform. Only the delivery differs: Kubernetes
writes them into the broker CR's `spec.systemScaling`, while docker and podman pass them to
the container as environment variables under the broker's own setting names. One env file
therefore sizes the same broker whichever platform runs it.

**Every setting is settable under either spelling.** The `scaling` key column below is the
schema's own friendly name; the destination column is the broker setting this tool emits for
it, which is *also* a legal key in the env file -- `system_scaling_maxconnectioncount: 1000`
works exactly like `maxConnections: 1000`. Both write the same value. Setting the same
setting under both spellings at once fails to load, naming both keys -- a file that could
hold a contradiction is worse than one that refuses it.

| `scaling` key | Broker setting (container env var / CR field) |
| --- | --- |
| `maxConnections` | `system_scaling_maxconnectioncount` |
| `maxQueueMessages` | `system_scaling_maxqueuemessagecount` |
| `maxKafkaBridge` | `system_scaling_maxkafkabridgecount` |
| `maxKafkaConnections` | `system_scaling_maxkafkabrokerconnectioncount` |
| `maxBridges` | `system_scaling_maxbridgecount` |
| `maxSubscriptions` | `system_scaling_maxsubscriptioncount` |
| `maxGuaranteedMsgMB` | `system_scaling_maxguaranteedmessagesize` |
| `maxSpoolUsageMB` | `messagespool_maxspoolusage` (now the same name on every platform) |

Every destination name above is the name BOTH platforms use: the key inside the CR's
`spec.systemScaling`, and the environment variable on docker and podman. That is a constraint
rather than a coincidence, and it has a limit worth knowing -- a systemd `Environment=` name
may contain only letters, digits and underscores, so a broker setting spelled with a hyphen
cannot reach a container this way and could not be added to the table above as it stands.

A key that is neither a friendly name nor a destination name -- a typo, or a real broker
setting this tool does not map (`system_scaling_maxtransactedsessioncount`, say) -- fails to
load exactly as an unknown key does anywhere else in this schema. `cpu` and `messagingNodeCpu`
are refused by name for a different reason: `scaling.cpu` is fixed by the `maxConnections`
tier and derived, so there is no key for it under either spelling (see Scaling tiers, below).

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
| `semp.adminPass` | `semp.adminPassEnv` |
| `semp.monitorPass` | `semp.monitorPassEnv` |
| `semp.additionalUsers[].password` | `semp.additionalUsers[].passwordEnv` |
| `tls.certPassphrase` | `tls.certPassphraseEnv` |
| `image.pass` | `image.passEnv` |
| `redundancy.psk` | `redundancy.pskEnv` |
| `replication.sites[].via.semp.pass` | `replication.sites[].via.semp.passEnv` |

```yaml
semp:
  adminPassEnv: SOLACE_ADMIN_PASS   # export SOLACE_ADMIN_PASS before any command
```

With the `*Env` form the env file carries no secret and is safe to commit and share. A
value is otherwise used **verbatim** on every platform -- a `$VAR` or `${VAR}` inside one
is a literal password, never expanded.

The pre-shared key is **mandatory on docker and podman** and **optional on Kubernetes**, and
it is refused at load when it is missing where it is required -- the error carries the
`openssl rand -base64 32` command. **Nothing in this tool generates it.** An earlier version
made one on a first HA deploy and rewrote the env file; that is gone, because it only ever
ran on one host, the value still had to be copied to the other two by hand, and a deploy
that edits the file it was handed is a surprise on a file that may be version-controlled. Nothing distributes a key across three container hosts, so each host's
env file must carry the same value or the group cannot form. On Kubernetes the operator
generates and distributes one itself when the key is empty, and the CR's
`spec.preSharedAuthKeySecret` is then omitted entirely; set it and the value is written as the
`preshared_auth_key` entry of `kubernetes.adminSecret` -- the same Secret the admin credentials
live in -- and the CR points at it.

The tool never echoes a secret. Values piped to a command on stdin show as
`<<< (N bytes on stdin)` under `-v/--verbose`, values passed to a child process's environment
as `NAME=***`, and `validate`/`broker status` report only whether each one is set. The
one exception is explicit and Kubernetes-only: there a Secret manifest IS the artifact, so
`broker generate` and `operator generate` print the values they would apply. On docker and
podman neither prints a secret at all -- the artifact references them by name and only
`broker deploy` handles the values. See
[Rendering without applying](operations.md#rendering-without-applying) for where each secret
lands at rest.

## Replication

`replication:` describes a **DR pair**: two separate brokers, each message-VPN active at one
site and standby at the other. It is not `redundancy:`, which is the three nodes of one HA
group -- a replicated deployment usually has both, an HA group at each site. Omit the whole
section unless this broker replicates.

Two commands read it. `broker configure data-replication` converges THIS broker to it -- the
mate's addresses, which VPNs replicate, and each one's role -- and never contacts the mate.
`broker perform data-replication` verifies both brokers and moves roles across the pair. See
[Data replication](operations.md#data-replication) for what each does and when to run it.

**The block is byte-identical at both sites.** Nothing in it is written from one broker's
point of view: there is no `mate:` key and no "my role", both of which would have to be
reversed in the other site's file. Each broker reads its own `show router-name` instead,
finds itself among `sites[].routerNames`, and whichever entry is not itself is its mate. So
the same text is pasted into both env files, and a failover is one edit -- change one
`activeAt` -- rather than two files kept in step.

```yaml
replication:
  sites:                                   # exactly 2
    - virtualRouterName: "v:sg1"           # quote it: the colon is a YAML indicator
      routerNames: [sg1, sg1b]             # every node of this site's HA group
      endpoints:
        - { host: 10.160.132.1, port: 55443, transport: ssl }
      via:                                 # optional; read only from the OTHER site
        kubernetes: { command: kubectl --context sg, namespace: solace-sg, name: solace }
    - virtualRouterName: "v:dr1"
      routerNames: [dr1]
      endpoints:
        - { host: 10.150.132.1, port: 55443, transport: ssl }
      via:
        semp: { host: 10.150.132.1, port: 1943, tls: true, passEnv: SOLACE_DR_ADMIN_PASS }
  vpns:
    - { name: ORDERS,   activeAt: "v:sg1" }
    - { name: PAYMENTS, activeAt: "v:dr1" }
```

| Key | Purpose |
| --- | --- |
| `replication.sites` | Exactly **2** entries. Replication is a pair; one site or three is refused at load |
| `sites[].virtualRouterName` | The site's **key**: what `vpns[].activeAt` references, and the literal operand the mate is given as `replication mate virtual-router-name`. Mandatory and never derived -- the file states the exact string the broker CLI will be handed, so nothing in the mate-address path is inferred. The two sites' values must differ. **Quote it**: a bare `v:sg1` does parse -- a colon ends a plain scalar only when a blank follows it -- but quoting a value whose whole point is a literal colon leaves nothing to reason about, and it is what the annotated sample teaches |
| `sites[].routerNames` | What this site's brokers **answer to**, matched against their own `show router-name` so a broker can find itself in this file. A list, because a site is usually an HA group and the backup node reports its own name -- list every node. Non-empty, and no name may appear under both sites |
| `sites[].endpoints` | How the **other broker** dials this one. At least one, at most 2 per transport. This tool never dials them: they are rendered into the mate's own CLI lines |
| `endpoints[].host` / `.port` | The address, port 1-65535. Against an **appliance** mate every endpoint must carry the same host and a distinct transport: that grammar has one `connect-via` address and one non-repeatable `connect-port` per transport, so a second host or a second port of the same transport is refused when the lines are rendered rather than silently dropped |
| `endpoints[].transport` | `plainText` (what an omitted transport means), `compressed` or `ssl`. `encrypted` is the routing name for the same thing and is refused by name, pointing at `ssl` |
| `sites[].via` | How **this tool** reaches that site when it is the mate. Exactly one child, `kubernetes:` or `semp:` -- the key present IS the mechanism, so a `via` cannot name one thing and configure another. Read only from the OTHER site's entry: a broker takes its own access from this file's `kubernetes:`/`docker:`/`podman:` section like every other command |
| `via.kubernetes.command` | The cluster CLI this tool runs to reach that site. Carry the cluster in it (`kubectl --context dr`) rather than beside it. **Restricted** -- see [The command fields are executable content](#the-command-fields-are-executable-content) |
| `via.kubernetes.namespace` / `.name` | The mate's namespace, and its PubSubPlusEventBroker name, which its pod is named after |
| `via.semp.host` / `.port` | The mate's SEMP address. **Not derived from `endpoints`**: replication runs over the message backbone, so a reachable replication endpoint proves nothing about SEMP reachability |
| `via.semp.tls` / `.insecure` | `https` rather than `http`, and whether to skip certificate verification for a self-signed mate. Declared, never inferred from this broker's own posture -- this hop is a WAN rather than a rack |
| `via.semp.pass` / `.passEnv` / `.passSecret` | Exactly one of the three supplies the **mate's** admin password, never this deployment's `semp.adminPass` -- a DR site is a different broker, and reusing this one's password would fail at best and hide the mistake if the two happened to match. `passSecret` is `{namespace, name, key}` of a Kubernetes Secret, read with that site's own `via.kubernetes.command` when it has one and this file's `kubernetes.command` otherwise. The username is always `admin`, and there is no prompt for the password |
| `replication.vpns[].name` | Listing a VPN **enables** replication for it at both sites. A VPN replicating on the broker but absent from this list has its replication **shut down** by `broker configure data-replication` -- the file is authoritative. Each name appears once |
| `replication.vpns[].activeAt` | Which site holds the **active** role for that VPN; the other is standby. It names a `virtualRouterName`, not a router name |

`via` is **optional**, and a file whose sites declare none is valid. Only
`broker perform data-replication` needs it, and it checks both sites' blocks in its own
preflight before anything is written; requiring it at load would refuse a file that
configures replication perfectly well with the local-only command. A `via:` key with nothing
under it decodes to the same value as an absent one, so it is accepted here too and the
switch command is what reports a site it cannot reach.

Everything else is checked at load: two sites, a `virtualRouterName` on each and no
duplicate, non-empty non-overlapping `routerNames`, at least one endpoint with a valid port
and transport and no more than two per transport, VPN names that appear once, and an
`activeAt` naming a declared site. A partially written block is an error rather than a
half-configured switchover waiting to happen.

## The command fields are executable content

`kubernetes.command`, `docker.command`, `podman.command`, `docker.compose` and each
`replication.sites[].via.kubernetes.command` name a binary this tool runs **on your
machine**. Env files travel -- repositories, pull requests, shared
archives -- so the person who wrote one is routinely not the person who runs it. Treat an
env file the way you would treat a script someone sent you: **read the command fields before
running anything with it.**

To make that review short, the fields are restricted. A command is accepted only when:

1. **Every token is inert and visible.** No control characters, no whitespace inside a single
   argument (any Unicode whitespace, not just the ASCII space), no invisible formatting
   characters (zero-width spaces and joiners, bidirectional overrides), no quotes, no
   backslash, no backtick, and none of `$ ; | & < > ( ) * ? [ ] { } # !`. Nothing is ever
   passed through a shell, so these are not injections -- but tokens end up in logs and in
   the `-v/--verbose` exec trace, and a token you cannot see is one you cannot
   review. A Windows path in a flag value therefore needs forward slashes:
   `--kubeconfig C:/Users/you/.kube/config`.
2. **A leading `~` is expanded, not refused** -- in every token but the first, to the home
   directory of whoever runs the tool, exactly as in a [path key](#relative-paths-resolve-against-the-env-file-not-the-current-directory).
   So `command: oc --kubeconfig ~/solace/kubecontext` works. A tilde anywhere else in a token
   is an ordinary character (an 8.3 short name such as `C:/Users/RUNNER~1/.kube/config` is
   what expansion itself produces on Windows), and a `~` that is still leading when the guard
   runs is refused -- nothing downstream would expand it.
3. **The binary is a bare name from the allowlist.** No `/` or `\` anywhere in it, and no
   leading `~` either -- it is the one token that is not expanded, because a path
   would run a file the env file chose -- such as a `./kubectl` unpacked beside it -- rather
   than the one on your `PATH`. One optional `.exe` is stripped, then the name must be:

   | Platform | Allowed |
   | --- | --- |
   | Kubernetes | `kubectl`, `oc` |
   | Docker | `docker`, `docker-compose`, `nerdctl` |
   | Podman | `podman` |

   A replication site's `via.kubernetes.command` is always held to the **Kubernetes** row
   whatever platform this end runs on: the mate may sit in a cluster while the local broker
   runs on docker, and the binary being run is a cluster CLI either way.

4. **Nothing after it is a bare word.** Flags and their values are fine
   (`kubectl --context prod -n solace`); a bare word is not, because this tool appends its
   own subcommand and a word in that position would run ahead of it. `kubectl delete` in a
   config is exactly the attack. The literal `--` is refused for the same reason.

   **One acknowledged gap, in that last rule.** The check cannot know how many values a
   flag takes -- that would mean carrying a table of every flag of every allowed CLI, which
   would rot as those CLIs change -- so the token after any flag is accepted as that flag's
   value. After a flag that takes *no* value, that token is not a value at all, and it lands
   in subcommand position after all:

   ```yaml
   kubernetes.command: kubectl --insecure-skip-tls-verify delete   # accepted; runs `delete`
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
solace-util broker deploy --allow-command microk8s   # kubernetes env file wrapping kubectl in microk8s
solace-util broker deploy --allow-command lima       # docker/podman env file wrapping the runtime in lima
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
`command: sudo podman` elevates every command this tool issues, for the whole life of an env
file, decided by whoever wrote that file. Elevate the tool instead, at the moment you run it,
so the privilege belongs to one invocation you chose:

```sh
sudo solace-util broker deploy -e prod.yaml   # yes (prod.yaml is a podman env file)
# command: sudo podman  in the env file      # never
```

The same check runs twice -- once when the env file is loaded, and again immediately before
any command line is built -- from a single implementation, so a hostile file is inert even
on a path that skipped validation.

**What this does not protect against.** Two things are out of scope, and no amount of
parsing would fix either:

- **A compromised machine.** If an attacker has already put a trojan `kubectl` on your
  `PATH`, they own the host; nothing this tool checks can help. What it does do is make the
  binary's real location visible -- before any work starts, each binary this env file names
  (`kubernetes.command`, `docker.command`, `podman.command`, `docker.compose`) is resolved and
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
solace-util validate -e prod.yaml
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
- `REPL_MATE`, `REPL_CONN_SSL` and `REPL_PSK` are **read and deliberately not carried
  over**, with a warning naming all three. They describe one mate; the `replication:` block
  describes both sites of the pair and needs values a bash env file does not hold, so a
  converted block could not validate. Write it by hand -- see [Replication](#replication).
- The converted file is re-read and validated, so a source env that was already missing
  mandatory values says so at conversion time.
- Without `-o` the YAML goes to stdout (warnings stay on stderr). With `-o` the file is
  written `0600`, and an existing file is confirmed before it is replaced -- `--no-prompt`
  answers yes, and a run with no terminal keeps the file and says so.

The output carries every secret from the source file verbatim -- treat it like the source,
and never commit it. (Switch the values to their `*Env` reference keys afterwards and it
becomes safe to commit; see [Secrets](#secrets).) `SOLBK_USR_SECRET` converts to
`kubernetes.adminSecret` (`SOLBK_ADM_SECRET` is accepted as an alias for the same key;
when both are set and disagree, the canonical `SOLBK_USR_SECRET` wins with a warning),
`SOLBK_SVR_SECRET` to `kubernetes.tlsServerSecret`, `IMAGEREPO_SECRET` to
`kubernetes.imagePullSecret` (on a docker/podman conversion those two are dropped with a
warning naming that kubernetes-only home -- they name Kubernetes Secret objects, which have
no container equivalent), and each `SOLBK_USR_PASS` entry to an `semp.additionalUsers` entry with
`accessLevel: none` -- the bash flow set no level, so the converter picks the least
privileged one and says so; raise it per user as needed.

**`kubernetes.msgNode.cpu` and `scaling.maxPool` fail to load, each naming its
replacement.** Broker CPU is fixed by the scaling tier rather than set by hand, and
`maxPool` would name the same broker setting as `scaling.maxSpoolUsageMB` -- one concept
under two platform-specific keys, which the scaling block keeps under a single name today.
That is not the same thing as the dual-spelling alias every other scaling setting now
gets (see [Scaling](#scaling)): `maxPool` and `maxSpoolUsageMB` had no defined winner if a
file set both, where an alias pair does -- setting both spellings of one setting fails to
load, naming both. `kubernetes.msgNode.mem` is unaffected. Docker always deploys through
compose, so there is no `docker.mode` key to choose a mode with; a file still carrying it
fails strict decoding as an unknown field. See [Scaling](#scaling).

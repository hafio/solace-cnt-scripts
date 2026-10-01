# Container security guidelines

> [!WARNING]
> **Not a supported Solace product.** `solace-util` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this tool, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this tool only, not the Solace PubSub+ Event Broker
> or the EventBroker Operator that it deploys and operates.

The rules a deployment tool must apply when it creates a Solace PubSub+ Event Broker
container, and how each one is expressed on docker, podman and Kubernetes.

Rules 1 to 5 restate Solace's own guidance -- [Security Considerations](https://docs.solace.com/Security/Security-Solace.htm)
and [Configuring Secrets](https://docs.solace.com/Software-Broker/Container-Tasks/Config-Secrets.htm).
Rules 6 to 8 are this document's own engineering rules. The operator facts come from the
[PubSub+ Kubernetes operator](https://github.com/SolaceProducts/pubsubplus-kubernetes-quickstart)
(repository `SolaceProducts/pubsubplus-kubernetes-quickstart`: `controllers/statefulset.go` and
the CRD, read at tag v1.4.2 -- which is also the operator this tool bundles and installs, so
they describe what actually runs); the image uid and rootless facts
from Solace's [Rootless Containers](https://docs.solace.com/Software-Broker/Container-Tasks/rootless-containers.htm)
page; the pre-shared-key `*filepath` key from the broker's configuration-keys reference.
Re-read those sources, and this document against them, whenever the bundled operator, the
podman floor or the broker image's major version changes. The
document is written to be lifted into any project that creates broker containers, not just this
one. The broker is shipped as a Linux container, so two layers have to hold: the container's
own posture, and the host it lands on (a third, the image's own provenance -- pinning it by
digest rather than a movable tag, verifying its signature -- is out of scope here). A tool can
only own the first, which is what this document is about.

## The eight rules

| # | Rule | Docker | Podman | Kubernetes | Passes when |
| --- | --- | --- | --- | --- | --- |
| 1 | Deny extended privileges | `privileged: false`, `cap_drop: [ALL]`, bridge networking unless host is asked for | `DropCapability=all`, the engine's private network (bridge rootful, pasta or slirp4netns rootless) unless host is asked for; there is no `Privileged=` key, so never widen | operator's job | audit greps 2 and 3 find the negatives, 4 finds nothing, 5 finds only asked-for host networking |
| 2 | Deny privilege escalation | `security_opt: no-new-privileges=true` | `NoNewPrivileges=true` | operator's job (`allowPrivilegeEscalation: false` in a raw pod spec) | grep 1 finds it in every container artifact |
| 3 | Run as a non-root identity | `user:` always emitted | `User=` always emitted, `Group=` whenever the run user carries a gid | non-zero `runAsUser`/`runAsGroup`/`fsGroup`, or unset where the operator or an SCC assigns them | grep 7 finds an identity in every artifact, never uid 0 on a privileged engine, and grep 10 finds SCC handling where OpenShift is a target |
| 4 | Bind only non-root ports | published ports are the operator's own list, passed through | same | Service ports name 8008/1443/1943 | grep 8 finds none of 80, 443 or 843 |
| 5 | Hand secrets over as files | engine secret, absolute target | engine secret, absolute target | Kubernetes Secrets | grep 9 finds a `*filepath` for every sensitive setting |
| 6 | Narrow the filesystem | private key as a secret, never a bind mount; read-only root not supported | same | read-only root not supported (`readOnlyRootFilesystem` is refused) | grep 6 finds nothing |
| 7 | State the setting, do not inherit it | emit it, pin it with a golden | same | emit what the CR carries and pin it; document the operator's hardcodes and pin the operator version | a golden per emitted setting, and a test forbidding every widening token |
| 8 | Document the version floor each setting needs | compose release per key | podman release per quadlet key | bundled operator version, and broker release per CR field | every floor has a row saying whether anything enforces it |

### 1. Deny extended privileges

Extended privileges let a container modify host files, host devices and the host's network
configuration. The broker needs none of that.

Docker's default is already off, so the work is to **not** widen it: no `privileged: true`,
no `cap_add`, no `devices`, no `sysctls`, no `security_opt` naming an unconfined seccomp or
AppArmor profile, no `pid: host` and no `ipc: host`. Emit `privileged: false` anyway, for the
reason rule 7 gives.

A podman quadlet unit -- the systemd `.container` file podman generates a container service
from -- has no `Privileged=` key at all, so the negative cannot be stated
there; what matters is that nothing reaches `PodmanArgs=` carrying `--privileged`, and that
`AddCapability=`, `SecurityLabelDisable=` and `SeccompProfile=` stay absent. An SELinux
relabel suffix on a volume (`:Z`, `:ro,Z`) is not a widening -- it is the opposite.

**Not adding a capability is not the same as dropping one.** Both engines start a container
with a default capability set -- docker's includes `NET_RAW`, `DAC_OVERRIDE`, `SETUID` and
`SETGID` among others, podman's a smaller one -- and a broker that runs as a non-root user and
binds only unprivileged ports (rules 3 and 4) uses none of them. Drop the whole set:
`cap_drop: [ALL]` in compose, `DropCapability=all` in the quadlet unit, `--cap-drop=ALL` on a
bare `create`/`run`. Neither engine drops anything unless told to, so like `no_new_privs` this
is silently unmet in any project that never mentions it. The operator's own `drop: [ALL]` for
the same image on Kubernetes, below, is the evidence that nothing the broker does needs one.

**Host networking is a widening too -- the one this document permits, and only on an explicit
opt-in.** `network_mode: host`, `Network=host` or `hostNetwork: true` put the container in the
host's network namespace: it sees every host interface, loopback included, binds its listeners
on all of them, and what can reach it is left to the host firewall. Default to the engine's
own private network -- a bridge on docker and rootful podman; rootless podman's default is
pasta, or slirp4netns before podman 5.0 -- publish only the ports the operator lists, and
render host networking only when the operator asks for it, so the choice is visible in the
artifact (rule 7). This tool
defaults `network.mode` to `bridge` and passes `network.ports` through as written, so an empty
list publishes nothing -- and an HA member on bridge must list its redundancy and SEMP ports
itself.

The Solace `PubSubPlusEventBroker` CRD exposes no `privileged` field, so a tool rendering
that CR cannot state this rule either way. The operator hardcodes `privileged: false` plus
`capabilities: drop: [ALL]` on the broker container, so the rule is met there by the operator
version you install, which makes the operator version part of your security posture.

The widening tokens, in every spelling this document checks for -- rule 7's negative test and
the audit greps at the end use this list, the compose keys anchored to their colon
(`devices:`, `sysctls:`) so a prose mention does not count: `privileged: true`, `--privileged`,
`cap_add`, `AddCapability=`, `devices`, `AddDevice=`, `sysctls`, `Sysctl=`,
`SecurityLabelDisable=`, `SeccompProfile=`, `unconfined`, `pid: host`, `ipc: host`. Host
networking is deliberately not in that list, since it is permitted on request; the fifth audit
grep finds it instead, in the spellings `network_mode: host`, `Network=host` and
`hostNetwork: true`.

### 2. Deny privilege escalation

This is the `no_new_privs` bit: once set, no process in the container can gain permissions
it did not start with, which neuters setuid binaries inside the image.

**No runtime turns it on for you.** Docker leaves it off unless the daemon is started with
`--no-new-privileges` or `daemon.json` sets it, and a podman quadlet's `NoNewPrivileges=`
defaults to false -- some versions of the podman-systemd.unit man page say otherwise, and the
generator source is what to believe. So this rule, with rule 1's capability drop, is silently
unmet in any project that simply never mentions it.

| Platform | The line to emit |
| --- | --- |
| docker compose | `security_opt:` then `- no-new-privileges=true` (the `:` separator also works, but dockerd logs a deprecation warning for it on every create) |
| podman quadlet | `NoNewPrivileges=true` in `[Container]` |
| docker/podman CLI | `--security-opt no-new-privileges` on `create`/`run` |
| Kubernetes | `allowPrivilegeEscalation: false` in the container security context |

On the Solace CR there is no such field; the operator hardcodes it, along with
`runAsNonRoot: true` and a `RuntimeDefault` seccomp profile.

### 3. Run as a non-root identity

Emit the identity explicitly on every platform and every deploy. Inheriting the image's user
means a future image tag silently changes who your broker runs as. The one place to leave it
unset is Kubernetes when the operator or an OpenShift SecurityContextConstraints object (an
SCC, OpenShift's admission policy for pod identities and privileges) assigns the ids for you -- then
document the default you inherit (the Solace operator's 1000001/1000002 with
`runAsNonRoot: true`), because that default is now part of your posture.

Which uid depends on whether the engine is privileged, and this is the part that is easy to
get wrong:

| Engine | Identity | Why |
| --- | --- | --- |
| docker, rootful podman | the image's own uid (`1000001`), group `0` | the engine maps container uids straight onto host uids, so container uid 0 **is** host root |
| rootless podman | a low uid (`1000`), group `0` | it has to fall inside a stock 65536-entry subuid range; a uid near 1000001 would need an allocation no `useradd` default provides |
| Kubernetes | non-zero `runAsUser`, `runAsGroup`, `fsGroup` | the pod security context and the container security context are separate fields and both matter |

Group 0 is not root-the-user: the Solace image expects gid 0 for its own files, and under
rootless podman container gid 0 maps to the invoking account's primary group.

Two corollaries worth enforcing rather than documenting. First, **refuse uid 0 where the
engine is privileged** -- accepting it means accepting host root, and a format check on
`uid[:gid]` will not catch it. Under rootless podman `0:0` is safe, because container uid 0
is only the invoking user. Second, on OpenShift leave the Kubernetes ids **unset** (or `0`,
which asks for auto-assignment): the restricted SCC assigns ids from the project's range, and
a concrete uid outside that range fails admission. A tool that suggests a literal uid in its
sample configuration should say this next to the suggestion.

The same identity governs the data directory: chown it to the run user on every deploy, and
under rootless podman do it inside the namespace (`podman unshare chown`), since the host-side
owner is a subuid nobody can become.

### 4. Bind only non-root ports

Binding below 1024 needs a root capability, so the broker's defaults moved up and a
non-root container is only possible if nothing asks for the old numbers:

| Service | Port | Was |
| --- | --- | --- |
| Web transport | 8008 | 80 |
| Web transport over TLS | 1443 | 443 |
| SEMP over TLS | 1943 | 843 |

Plaintext SEMP is 8080 and needs no change. A tool does not usually choose these -- the
broker image does -- so the tool's job is to make sure every port list it emits (a published
list in bridge mode, a Kubernetes Service, a health check, its own SEMP client) names the
same set, and that 80, 443 and 843 appear nowhere. Under host networking nothing is published
at all and the image's listeners bind directly on the host, which is worth saying out loud in
the documentation, because it means the port rule is satisfied by the image and not by you --
one more reason host networking is an opt-in (rule 1) rather than a default. Under bridge
networking the published list is exactly what the operator wrote: pass it through rather than
defaulting one, since a default list decides for every deployment which listeners face the
network.

### 5. Hand secrets over as files

Every sensitive broker setting has a `*filepath` variant, and that is the only form a tool
should use: redundancy group pre-shared key, server certificate, certificate passphrase, and
each CLI user password.

```
username_admin_passwordfilepath
username_<user>_passwordfilepath
redundancy_authentication_presharedkey_keyfilepath
tls_servercertificate_filepath
tls_servercertificate_passphrasefilepath
```

Six properties make the difference between using that mechanism and only appearing to:

- **Absolute targets.** Solace's example mounts a host directory at `/run/secrets`, but that
  path is each engine's own default secret mount, so a bare target name silently lands there
  and a `*filepath` pointer elsewhere breaks with no error. On RHEL and OpenShift, CRI-O also
  bind-mounts host subscription data over `/run/secrets`. Name the full path you point at.
- **Values never on argv.** Feed the engine's secret store on stdin (`podman secret create
  <name> -`), or hand the value to the child process environment, or apply a manifest with
  `kubectl apply -f -`. A command line is visible to every process on the host.
- **Mode and owner on the mount.** Both docker compose and podman default a mounted secret to
  uid 0, mode 0444 -- world-readable inside the container, and not owned by the non-root user
  from rule 3. Set them. On compose that only works for a secret sourced from `environment:`;
  a `file:` secret is a bind-mount and the three attributes are silently ignored, so pre-set
  the host file's owner and mode instead.
- **Know where the bytes rest.** This is where "kept in RAM" claims usually fail. A compose
  secret sourced from `environment:` is copied into the container's writable layer on the host
  disk, not a tmpfs. Podman's default file driver keeps every secret, unencrypted (base64
  values, mode 0600), in `<graphroot>/secrets/filedriver/secretsdata.json` -- the graph root
  is podman's storage directory, `podman info --format '{{.Store.GraphRoot}}'` -- and at
  `podman create` copies each one into the container's own storage directory under the graph
  root, applying uid/gid/mode there; at start those files are bind-mounted into the container,
  and nothing is placed in a tmpfs. Kubernetes Secret volumes are kubelet tmpfs, while
  encryption at rest in etcd is cluster configuration nobody's deploy tool sets. Write down
  which of these applies rather than asserting the strongest one.
- **A core dump is one more resting place.** It is a copy of the broker's memory, credentials
  included. Solace recommends an unlimited core limit so a crash can be diagnosed, and this
  tool defaults to it; where that trade is wrong, lower the limit -- `ulimits.core: 0` here
  asks for none. The limit only says what the container may produce: whether a dump is kept,
  and where, is the host's `kernel.core_pattern` and the handler it names, which is host
  hardening and outside what a deploy tool owns.
- **Never echo a secret back.** Reports say `set` or `MISSING`; a preview masks child
  environment values.

Where a platform genuinely cannot carry a secret -- the CRD has no field for a certificate
passphrase, and arbitrary extra users ride an env-var Secret -- fail loud and name the
limitation. A silent drop turns into a broker with a blank credential.

### 6. Narrow the filesystem

A read-only root filesystem is the strongest form of this rule: only the data directory, and
whatever scratch paths the broker writes, stay writable. The broker accepts one from 10.9 (the
floor the operator's CRD states for `readOnlyRootFilesystem`), but **this tool does not support
it on any platform yet**. Nothing it renders asks for one, and
`kubernetes.containerSecurity.readOnlyRootFilesystem` is refused by name, so no env file can
believe it is in effect. Turning it on needs the paths the broker writes outside its data
directory established per engine -- docker gives a read-only container no writable scratch
space of its own -- and proven against a running broker. Until then the root stays writable,
and the sixth audit grep proves nothing asks otherwise.

What does apply: mount anything the broker only reads -- a CA directory, say -- read-only, and
hand a private key over as an engine secret like any other credential (rule 5) rather than
bind-mounting a host file. A key file restricted to 0600 on the host is owned by whoever wrote
it, which inside the container is root, so the non-root identity from rule 3 cannot read it --
the restriction that protects the key on the host is the one that breaks the broker.

### 7. State the setting, do not inherit it

A security setting that comes from a runtime default is invisible in the artifact and
unprotected against a future edit. Nobody auditing your generated compose file or quadlet
unit can see that extended privileges are denied, and nothing fails if someone adds a
widening key next quarter.

So: emit the setting even when it matches the default, and keep two tests on the rendered
artifact: one that each emitted negative is present (`privileged: false` and `cap_drop: [ALL]`
in a compose file, `NoNewPrivileges=true` and `DropCapability=all` in a quadlet unit), and one
that it contains none of the widening tokens in rule 1's list, in every spelling that list
gives. The negative test is the part that survives a refactor; the emitted line is what an
operator can read. The one permitted widening, host networking, gets a test of its own: absent
from every artifact rendered without the opt-in, present exactly once in the one rendered with
it.

Neither test sees a widening key that arrives *inside a value*. A quadlet unit and a YAML file
are line-structured, so any configurable value written into one as-is -- a published port, a
health-check interval, an image tag -- can carry a newline and start a key of its own:
`AddCapability=ALL`, a `[Service]` section with an `ExecStartPre=` that systemd runs as root, a
`cap_add:` in compose, or a second `---` document in a manifest. Hold every such value to the
grammar of the line it lands on at load, and test that the injections are refused, not just
that the rendered defaults are clean.

### 8. Document the version floor each setting needs

Security settings have floors like any other feature, and a floor failure at deploy time is
better than a setting that silently does nothing. `NoNewPrivileges=` and `DropCapability=`
arrived with Quadlet in podman 4.4. The quadlet `Secret=` key needs 4.5 (an absolute `target=` on `--secret` has been
accepted since 4.4, so the key, not the path, is the 4.5 dependency). Quadlet refuses a unit
carrying a key it does not know, so every key is a floor: `ShmSize=` and `Ulimit=` need 4.7,
and `Memory=` needs 5.5 -- a memory cap written as `PodmanArgs=--memory=` needs only the
podman run flag, which every supported podman has. Check each key against the release that
introduced it, and pin that in a test rather than a comment. This tool documents
Compose 2.23.1 as its floor for the `environment:` secret source; the public record (compose
PR #10084, merged 2022-12, already applying uid/gid to environment-sourced secrets) shows the
source accepted well before that, so treat such a number as the floor a project chose, not the
one the feature needs, and say which. Put them in the same table as your other floors and say
whether anything enforces them -- this tool's table is in
[operations.md](operations.md#version-floors), and it says plainly that nothing enforces the
compose floor at runtime.

On Kubernetes the floor is the operator you install. Its CRD decides which security-relevant
fields the CR can carry at all -- the two security contexts, and the Secret references
`adminCredentialsSecret`, `preSharedAuthKeySecret` and `tls.serverTlsConfigSecret` -- and its
controller decides what is hardcoded on the broker pod whatever the CR says: the
`privileged: false`, `drop: [ALL]`, `allowPrivilegeEscalation: false`, `runAsNonRoot: true` and
`RuntimeDefault` seccomp profile that rules 1 to 3 rely on. A different operator version can
change either list without the CR changing at all, so pin the operator version, name it (this
tool bundles 1.4.2), and re-check rules 1 to 3 against the new controller before bumping it. A
CR field that needs a broker release -- `readOnlyRootFilesystem` needs 10.9 -- is a floor on
the image tag as well.

## Auditing a project against this

Read-only, in a checkout, from the repository root. What a result means differs per grep, so
the table below says, for each numbered line, which rule it checks and what a finding looks
like. For greps 1 to 6, which ask what is emitted, a hit in a `.go` file counts for nothing
either way -- only a golden or a yaml shows that; greps 8 to 10 also look for handling in the
code, where a `.go` hit is the evidence. The first five share one `--include` set so a token
cannot hide in the file type one grep skips.

```
grep -rn -iE 'no-new-privileges|NoNewPrivileges|allowPrivilegeEscalation' --include='*.golden' --include='*.yaml' --include='*.go' .   # 1
grep -rn -E 'privileged: *false' --include='*.golden' --include='*.yaml' --include='*.go' .   # 2
grep -rn -E 'cap_drop|DropCapability=' --include='*.golden' --include='*.yaml' --include='*.go' .   # 3
grep -rn -E 'privileged: *true|--privileged|cap_add|AddCapability=|devices:|AddDevice=|sysctls:|Sysctl=|SecurityLabelDisable=|SeccompProfile=|unconfined|pid: *host|ipc: *host' --include='*.golden' --include='*.yaml' --include='*.go' .   # 4
grep -rn -E 'network_mode: *host|Network=host|hostNetwork: *true' --include='*.golden' --include='*.yaml' --include='*.go' .   # 5
grep -rn -E 'read_only:|ReadOnly=|--read-only|readOnlyRootFilesystem' --include='*.golden' --include='*.yaml' .   # 6
grep -rn -E 'runAsUser|runAsGroup|fsGroup|^ *user:|^User=|^Group=' --include='*.golden' .   # 7
grep -rn -E '(:|=|(containerPort|servicePort|targetPort|hostPort|port): )(80|443|843)\b' --include='*.go' --include='*.golden' --include='*.yaml' .   # 8
grep -rn -iE 'passwordfilepath|presharedkey_keyfilepath|servercertificate_filepath|passphrasefilepath' --include='*.go' --include='*.golden' .   # 9
grep -rn -iE 'openshift|securitycontextconstraints|\bscc\b|anyuid' --include='*.go' --include='*.md' --exclude=container-security.md .   # 10
```

| Grep | Rule | A finding is |
| --- | --- | --- |
| 1 no-new-privileges | 2 | an EMPTY result, or a container artifact missing from it |
| 2 `privileged: false` | 1, 7 | an EMPTY result, or a compose file missing from it |
| 3 capability drop | 1, 7 | an EMPTY result, or a compose file or quadlet unit missing from it |
| 4 widening tokens | 1 | any hit in a golden or a yaml |
| 5 host networking | 1 | a hit in an artifact rendered without the operator asking for it -- here the two goldens pinning `network.mode: host` are the expected hits, and a default carries none |
| 6 read-only root | 6 | for a tool that does not support it, as this one does not, any hit in a golden or a yaml; a tool that does support it reads a hit as its evidence instead |
| 7 identity | 3 | an artifact with no identity, or uid 0 on a privileged engine |
| 8 ports | 4 | 80, 443 or 843 in any port position |
| 9 `*filepath` | 5 | a sensitive setting reaching the broker by value rather than by file |
| 10 OpenShift | 3 | no handling of SCC-assigned ids where OpenShift is a target |

The port grep matches the three shapes a port takes in a rendered artifact -- `host:80` (a
published pair), `containerPort: 80` (a Kubernetes Service, anchored to the port keys so a
pod-affinity `weight: 80` does not match) and `tcp-web=80` (a named list); a `<file>:80` line
reference in a comment is its expected false positive. The identity grep's expected false
positive here is a legacy env file's `image.user` under `internal/convert/testdata`, this
project's own fixture directory -- a lifted copy substitutes its own. The last grep
excludes this document, which would otherwise match itself; a real hit there is an OpenShift
or SCC check in the Kubernetes package, not a mention of `oc` or of `"0"` auto-assignment in a
comment or hint.

Then read the rendered artifacts rather than the renderer: a golden file is what the operator
actually gets. For each of the eight rules, classify what you find as emitted explicitly,
inherited from a runtime default, left to user configuration with no default and no
validation, or absent -- and record which. The four are not the same finding, and only the
first is durable.

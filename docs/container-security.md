# Container security guidelines

The rules a deployment tool must apply when it creates a Solace PubSub+ Event Broker
container, and how each one is expressed on docker, podman and Kubernetes.

Sourced from Solace's own guidance -- [Security Considerations](https://docs.solace.com/Security/Security-Solace.htm)
and [Configuring Secrets](https://docs.solace.com/Software-Broker/Container-Tasks/Config-Secrets.htm)
-- and written to be lifted into any project that creates broker containers, not just this
one. The broker is shipped as a Linux container, so two layers have to hold: the container's
own posture, and the host it lands on. A tool can only own the first, which is what this
document is about.

[docs/container-security-remediation.md](container-security-remediation.md) records where
**this** repository currently stands against these rules.

## The eight rules

| # | Rule | Docker | Podman | Kubernetes |
| --- | --- | --- | --- | --- |
| 1 | Deny extended privileges | `privileged: false` | no key exists; never widen | operator's job |
| 2 | Deny privilege escalation | `security_opt: no-new-privileges:true` | `NoNewPrivileges=true` | `allowPrivilegeEscalation: false` |
| 3 | Run as a non-root identity | `user:` always emitted | `User=`/`Group=` always emitted | non-zero `runAsUser`/`runAsGroup`/`fsGroup` |
| 4 | Bind only non-root ports | published list names 8008/1443/1943 | same | Service ports name the same set |
| 5 | Hand secrets over as files | engine secret, absolute target | engine secret, absolute target | Kubernetes Secrets |
| 6 | Narrow the filesystem | read-only root where the image allows | same | `readOnlyRootFilesystem` |
| 7 | State the setting, do not inherit it | emit it, pin it with a golden | same | same |
| 8 | Document the version floor each setting needs | -- | -- | -- |

### 1. Deny extended privileges

Extended privileges let a container modify host files, host devices and the host's network
configuration. The broker needs none of that.

Docker's default is already off, so the work is to **not** widen it: no `privileged: true`,
no `cap_add`, no `devices`, no `sysctls`, no `security_opt` naming an unconfined seccomp or
AppArmor profile, no `pid: host` and no `ipc: host`. Emit `privileged: false` anyway, for the
reason rule 7 gives.

A podman quadlet unit has no `Privileged=` key at all, so the negative cannot be stated
there; what matters is that nothing reaches `PodmanArgs=` carrying `--privileged`, and that
`AddCapability=`, `SecurityLabelDisable=` and `SeccompProfile=` stay absent. An SELinux
relabel suffix on a volume (`:Z`, `:ro,Z`) is not a widening -- it is the opposite.

The Solace `PubSubPlusEventBroker` CRD exposes no `privileged` field, so a tool rendering
that CR cannot state this rule either way. The operator hardcodes `privileged: false` plus
`capabilities: drop: [ALL]` on the broker container, so the rule is met there by the operator
version you install, which makes the operator version part of your security posture.

### 2. Deny privilege escalation

This is the `no_new_privs` bit: once set, no process in the container can gain permissions
it did not start with, which neuters setuid binaries inside the image.

**No runtime turns it on for you.** Docker leaves it off unless the daemon is started with
`--no-new-privileges` or `daemon.json` sets it, and a podman quadlet's `NoNewPrivileges=`
defaults to false -- some versions of the podman-systemd.unit man page say otherwise, and the
generator source is what to believe. So this is the one rule that is silently unmet in any
project that simply never mentions it.

| Platform | The line to emit |
| --- | --- |
| docker compose | `security_opt:` then `- no-new-privileges:true` |
| podman quadlet | `NoNewPrivileges=true` in `[Container]` |
| docker/podman CLI | `--security-opt no-new-privileges` on `create`/`run` |
| Kubernetes | `allowPrivilegeEscalation: false` in the container security context |

On the Solace CR there is no such field; the operator hardcodes it, along with
`runAsNonRoot: true` and a `RuntimeDefault` seccomp profile.

### 3. Run as a non-root identity

Emit the identity explicitly on every platform and every deploy. Inheriting the image's user
means a future image tag silently changes who your broker runs as.

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
the documentation, because it means the port rule is satisfied by the image and not by you.

### 5. Hand secrets over as files

Every sensitive broker setting has a `*filepath` variant, and that is the only form a tool
should use: redundancy group pre-shared key, server certificate, certificate passphrase, and
each CLI user password.

```
username_admin_passwordfilepath
redundancy_authentication_presharedkey_keyfilepath
tls_servercertificate_filepath
tls_servercertificate_passphrasefilepath
```

Five properties make the difference between using that mechanism and only appearing to:

- **Absolute targets.** Solace's example mounts a host directory at `/run/secrets`, but that
  path is each engine's own default secret mount, so a bare target name silently lands there
  and a `*filepath` pointer elsewhere breaks with no error. On RHEL and OpenShift, CRI-O also
  bind-mounts host subscription data over `/run/secrets`. Name the full path you point at.
- **Values never on argv.** Feed the engine's secret store on stdin (`podman secret create
  <name> -`), or hand the value to the child process environment, or apply a manifest with
  `kubectl apply -f -`. A command line is visible to every process on the host.
- **Mode and owner on the mount.** Both docker compose and podman default a mounted secret to
  uid 0, mode 0444 -- world-readable inside the container, and not owned by the non-root user
  from rule 3. Set them.
- **Know where the bytes rest.** This is where "kept in RAM" claims usually fail. A compose
  secret sourced from `environment:` is copied into the container's writable layer on the host
  disk, not a tmpfs. Podman's default file driver keeps secret data in a plaintext JSON file
  under the graph root and copies each secret into the run directory at start, so it is tmpfs
  only insofar as that directory is. Kubernetes Secret volumes are kubelet tmpfs, while
  encryption at rest in etcd is cluster configuration nobody's deploy tool sets. Write down
  which of these applies rather than asserting the strongest one.
- **Never echo a secret back.** Reports say `set` or `MISSING`; a preview masks child
  environment values.

Where a platform genuinely cannot carry a secret -- the CRD has no field for a certificate
passphrase, and arbitrary extra users ride an env-var Secret -- fail loud and name the
limitation. A silent drop turns into a broker with a blank credential.

### 6. Narrow the filesystem

Ask for a read-only root filesystem where the image supports it (broker 10.9 and later), keep
the writable data directory as the one exception, and mount anything the broker only reads --
a certificate bundle, a CA directory -- read-only. Restrict any host-side file holding a
private key to 0600 inside a 0700 directory, and delete it on teardown.

### 7. State the setting, do not inherit it

A security setting that comes from a runtime default is invisible in the artifact and
unprotected against a future edit. Nobody auditing your generated compose file or quadlet
unit can see that extended privileges are denied, and nothing fails if someone adds a
widening key next quarter.

So: emit the setting even when it matches the default, and keep a test that the rendered
artifact contains none of the widening tokens -- `--privileged`, `AddCapability=`,
`cap_add`, `SecurityLabelDisable=`, `unconfined`, `pid: host`. The negative test is the part
that survives a refactor; the emitted line is what an operator can read.

### 8. Document the version floor each setting needs

Security settings have floors like any other feature, and a floor failure at deploy time is
better than a setting that silently does nothing. `NoNewPrivileges=` needs podman 4.4. An
absolute `target=` on a mounted podman secret needs 4.5. The `environment:` secret source in
a compose file needs Compose 2.23.1. Put them in the same table as your other floors and say
whether anything enforces them.

## Auditing a project against this

Read-only, in a checkout, from the repository root. Any empty result on the first two is a
finding, not a pass.

```
grep -rn -iE 'no-new-privileges|NoNewPrivileges|allowPrivilegeEscalation' --include='*.go' --include='*.yaml' --include='*.golden' .
grep -rn -iE 'privileged|cap_add|AddCapability|SecurityLabelDisable|unconfined|pid: *host|ipc: *host' --include='*.go' --include='*.golden' .
grep -rn -iE 'runAsUser|runAsGroup|fsGroup|user:|User=' --include='*.golden' .
grep -rn -E ':80\b|:443\b|:843\b' --include='*.go' --include='*.golden' --include='*.yaml' .
grep -rn -iE 'passwordfilepath|presharedkey_keyfilepath|servercertificate_filepath|passphrasefilepath' --include='*.go' --include='*.golden' .
grep -rn -iE 'openshift|securitycontextconstraints|\bscc\b|anyuid' --include='*.go' --include='*.md' .
```

Then read the rendered artifacts rather than the renderer: a golden file is what the operator
actually gets. For each of the eight rules, classify what you find as emitted explicitly,
inherited from a runtime default, left to user configuration with no default and no
validation, or absent -- and record which. The four are not the same finding, and only the
first is durable.

# Operations

Running a broker after the [quick start](../README.md#quick-start): what each phase does,
what re-deploying and removing actually change, and how to upgrade. For the complete command
surface see [commands.md](commands.md); for the env file see
[configuration.md](configuration.md).

- [The preflight](#the-preflight)
- [Bringing up a fresh cluster](#bringing-up-a-fresh-cluster)
- [Post-deployment configuration order](#post-deployment-configuration-order)
- [Docker and Podman mechanics](#docker-and-podman-mechanics)
- [Rendering without applying](#rendering-without-applying)
- [Removing a broker: what stays, what goes](#removing-a-broker-what-stays-what-goes)
- [Upgrading a running broker](#upgrading-a-running-broker)
- [Troubleshooting](#troubleshooting)

## The preflight

Every command that writes a file or changes remote state runs one cheap **read-only** probe
first, using the same command the real work will use:

| Platform | Probe | Answers |
| --- | --- | --- |
| Kubernetes | `<runtime> auth can-i <verb> <resource> -n <namespace>` | Is the context live, and may this identity do the thing? |
| Docker / Podman | `<runtime> info` | Is the daemon up and reachable as this user? |

If it fails, the command stops nonzero **before the first byte is written**, passes the
CLI's own error through, and adds one line saying what to do -- log in, ask for a role
binding, start the daemon. It never logs you in or starts anything on your behalf, and
there is no flag to skip it: previewing a command's effect without touching a cluster is
what `generate` is for, and a render-only command never runs the preflight because it never
runs anything.

## Bringing up a fresh cluster

The EventBroker operator is cluster-scoped and shared between brokers, so it is installed
and removed on its own rather than as a side effect of any one broker's `deploy all` or
`remove all`. A cluster that has never run this tool (or any other operator install) needs
it once:

```
solace-util deploy operator -e dev.yaml
```

After that, any number of env files can each `deploy all` their own broker against the same
cluster without touching the operator again. `check deploy` warns rather than fails when
the operator or its CRD looks missing, so that warning is what tells you this step was
skipped -- the actual `deploy broker` (or `deploy all`) then fails once it tries to apply a
custom resource the cluster does not know how to reconcile.

`check deploy`'s report opens with a **Config** section, ahead of Deployment: it states that
the env file was already accepted -- `config.Load` validates before this report can even
start building, so reaching the report means it passed; this is not a fresh check running --
then lists the resolved `kubernetes.ports` as `name=port` pairs behind a leading count (e.g.
`ports (17)` for the default set), so what will actually render into the broker CR is visible
without deploying or running `generate broker`. Every row in this section is informational
and can never fail the check.

End to end, a first run against a brand-new cluster looks like:

```
solace-util deploy operator -e dev.yaml     # once per cluster
solace-util check deploy -e dev.yaml        # prerequisites: cluster, StorageClass, operator
solace-util deploy all -e dev.yaml          # check -> namespace -> secrets -> CR -> leader (HA)
solace-util check semp-login -e dev.yaml    # prove it answers
```

Removing a broker (`remove broker` / `remove all`) never removes the operator either -- see
[Removing a broker: what stays, what goes](#removing-a-broker-what-stays-what-goes) below.
Uninstall it explicitly, and only once nothing else in the cluster still depends on it:

```
solace-util remove operator -e dev.yaml
```

## Post-deployment configuration order

Everything under **`config` is post-deployment**: each step drives the Solace CLI inside a
broker that is already running, so none of it is part of `deploy` and none of it is run by
`deploy all`. Wait for the broker to be ready (the pods, or the container/service), then run
these in order.

They are not uniformly re-runnable -- `config apply additional-users` fails outright on a
user that already exists -- so the order lives here rather than in a run-everything command
that would stop partway through a second run. On a fresh broker:

1. `config leader` (HA only; on containers, run it on the primary -- it also reverts the
   backup over SEMP first, so `nodes.backup.ip` should be reachable)
2. `config apply server-cert` (when TLS is configured)
3. `config apply domain-certs` (when any are listed)
4. `config disable default-vpn`
5. `config disable default-users`
6. `config apply additional-users` (Kubernetes only; run after the hardening steps above,
   and **not re-runnable** -- the broker refuses to create a user that already exists)
7. `config apply product-keys` (when any are listed)

Only step 3 can be undone from here (`config delete domain-certs`). There is no un-harden,
and no way to withdraw a server certificate or a product key through this tool.

**The exit code now reflects what the broker said.** Every `config apply` step (and
`cli --input`) still runs the whole script it uploads -- a Solace CLI script is a sequence of
independent commands, so one rejected line never stops the rest, and the output is always
shown and the uploaded script always removed. But the command itself now fails once the
script has finished if any line of that output contains `invalid`, `error` or `busy`
(case-insensitively): the error names how many lines were rejected, without quoting them,
since a CLI transcript can carry passwords. This used to only warn and exit 0.
**This is a breaking change for automation**: a CI job that runs a `config apply` step or
`cli --input <script>` and gates on the exit code -- and passes today -- can start failing,
because a partly-applied script is now caught instead of silently accepted. Check any
pipeline that runs these commands before relying on the new exit code.

### Extra CLI users differ by platform

`admin.additionalUsers` reaches the broker two different ways, because the operator has no
declarative route for it:

- **Docker / Podman** -- created at container boot, from the mounted password file plus a
  `username_<username>_globalaccesslevel` setting in the artifact. Nothing to run
  afterwards.
- **Kubernetes** -- created post-deployment by **`solace-util config apply additional-users`**,
  which builds a Solace CLI script and runs it on the primary. Verified against a live
  cluster: extra `username_<user>_password` keys in the credentials Secret are **ignored by
  the operator**, and the only declarative alternative (`extraEnvVars` /
  `extraEnvVarsSecret`) would publish the passwords in the pod's environment, where
  `kubectl describe` and every process in the container can read them. So the CLI is the
  route, and the Secret carries no extra users at all.

Two consequences on Kubernetes worth knowing:

- **It is not re-runnable.** The broker's `create username` fails if the user exists, and
  that is reported rather than reconciled -- re-setting a password an operator rotated on
  the broker would be worse. So a repeated `config apply additional-users` fails once the
  users exist; run the other `config` steps above individually instead, or drop the
  already-created users from the env file.
- **The password charset is restricted.** The value goes onto a CLI line, and the broker
  rejects ``:()";'<>,`\*&|`` inside it. An env file using one of those fails to load *for
  Kubernetes* with the offending character named (never the password). The same file stays valid
  for docker and podman, which write the password to a file instead.

## Docker and Podman mechanics

The `docker` and `podman` halves of the command tree share one implementation for a
**host-local** broker: one container per host, driven over `<runtime> exec`/`cp` (no
operator, no cluster). Only the deploy artifact differs -- Docker renders a compose file
and brings it up with `docker compose`, Podman a systemd **quadlet** `.container` unit.
(Docker always deploys through compose -- there is no `docker.mode` key to choose a mode
with. A bare `docker run` cannot recreate an existing container, so re-deploying after an
image-tag bump would fail on a name conflict where compose recreates cleanly. An env file
carrying a `docker.mode` key fails strict decoding as an unknown field.)

**HA verification runs from the primary.** The transport is node-local, so exec reaches
only this host's broker -- but the broker itself is a control channel: the primary's own
`show redundancy` already reports the mate's activity, and the one command that must land
on the backup (`redundancy revert-activity`) is sent over SEMP to `nodes.backup.ip`. Bring
the group up by running `deploy all <role>` on each host with its own role -- or omit
`<role>` and it is detected by matching this host's hostname against `nodes.*`
(case-insensitively, tolerating an FQDN on either side), announcing the role it chose on
stderr; a hostname matching none of the three names fails loud, and one matching more than
one fails loud too, naming every role it matched. An explicit role always wins over
detection, which is the escape hatch for a host whose name does not match the env file.
`generate broker` detects the same way but does NOT fail on an unrecognised host: it renders
the primary's artifact with a warning, because it changes nothing and the artifact names the
node it is for, so the fallback is visible in the output you are about to read. Refusing there
would make reviewing another node's artifact from a laptop impossible.
Then run `smoke redundancy` **once, on the primary host**: it confirms the primary healthy, checks
the backup's SEMP service is reachable (before anything is disturbed), releases and
un-releases activity so the backup takes over, then reverts the backup and waits for
activity to come home. Backup and monitor hosts are rejected loud, and `config leader`
runs only on the primary. Prerequisite: the backup's SEMP port (8080, or the mapped host
port under `network.mode: bridge`) must be reachable from the primary host -- a working HA
group only proves the redundancy ports (8300-8302, 8741, 55555) are open, so the
preflight fails loud with the address it tried when SEMP is firewalled.

**If a `smoke redundancy` run dies partway** (including Ctrl-C, which runs no cleanup), the
group can be left released or failed over. Restore it with `no redundancy release-activity`
on the primary's CLI, or `redundancy revert-activity` on the backup's -- or simply re-run
`smoke redundancy` once the cause is fixed.

Example (HA -- run each line on the matching host; `prod.yaml` is a podman env file):

```
solace-util deploy all primary -e prod.yaml   # on the primary host
solace-util deploy all backup  -e prod.yaml   # on the backup host
solace-util deploy all monitor -e prod.yaml   # on the monitor host
solace-util config leader -e prod.yaml        # on the primary only
solace-util smoke redundancy -e prod.yaml     # on the primary only -- drives the whole group
```

Naming the role explicitly is what to do when a host's name will not match `nodes.*` (or
matches more than one); when it does match, `solace-util deploy all -e prod.yaml` alone on
each host detects and announces the same role.

### Re-deploying is safe and explicit

`deploy broker` renders the artifact and compares it with
what is already on disk, so the three outcomes are distinguishable:

- **Unchanged, broker running** -- reported as nothing to do; the broker is not touched.
  With `--restart` it is recreated/restarted anyway, which is how a rotated secret is
  applied (nothing in the artifact changes when a password does).
- **Broker not running** -- recreated from the current config, so a rotated secret takes
  effect without `--restart`. A stopped container would otherwise be *started* with the
  credentials it was created with; there is no traffic to protect here, so no consent is
  asked. `--restart` is only needed for a running broker, which is the case where applying
  a change costs a bounce. Confirmed on Docker (`--force-recreate`); on Podman this relies
  on quadlet's own container replacement at unit start, which is assumed but has not been
  independently verified.
  On Docker, "not running" has to be an answered probe, not an assumption: if the `ps` check
  itself fails (a transient engine hiccup), `deploy broker` now aborts naming the container
  rather than guessing "not running" and force-recreating a broker that might still be live --
  by hand, check `<runtime> ps` and re-run. A probe that answers and simply finds no match
  (genuinely absent, or stopped) is unaffected. Podman's equivalent path issues a plain
  `systemctl start`, a no-op on an already-active unit, so it carries no such risk and still
  treats an unanswered probe as "not active".
- **Changed, broker not running** -- written and started.
- **Changed, broker running** -- written, then you are asked before it is bounced.
  `--restart` pre-approves; a non-interactive run declines, leaving the new artifact in
  place and warning that the running broker is still on the previous one. `--restart` is
  deliberately its own flag: dropping messaging traffic is its own decision.

This is what makes an image-tag bump a one-command upgrade: edit `image.tag`, then
`solace-util deploy broker <role> --restart -e prod.yaml` (podman) or
`solace-util deploy broker --restart -e prod.yaml` (docker). Podman needs `--restart`
because `systemctl start` on an already-active unit is a no-op, so without it the unit file
is rewritten but the running container keeps the previous image.

**Secrets.** `deploy broker` externalizes every secret before applying the artifact, so no
value is ever written into the compose file or quadlet unit (see the table under
[Rendering without applying](#rendering-without-applying) for the names and
paths). Podman loads them into its own secret store (`podman secret create --replace`,
value on stdin) and the unit mounts them; Docker's compose file names a host environment
variable per secret, and `deploy broker` sets those variables for its own `docker compose`
process, so no value ever reaches an argv or a file beside the compose file. A missing
value (notably `nodes.psk` before `prepare host` has run) fails the deploy loudly rather
than starting a broker without a password.

What that does and does not buy you: **nothing is written next to the artifact** (no
plaintext file a project-directory backup, `tar`, or non-root user would pick up, and
nothing to clean up on teardown), but the value still ends up at rest -- Docker
materializes each secret into the container's own filesystem as a `0444` root-owned file
under `/run/secrets`, which is the same at-rest exposure class as podman's store. It is
not in the container's environment, so `docker inspect` does not show it.

`generate secrets broker` prints the equivalent shell, one line per secret, for running
compose yourself: `podman secret create` commands to run once on Podman, and `export` lines to
**source** in the shell you run `docker compose` from on Docker. A manual
`docker compose up` needs those variables exported -- unset, compose refuses.

**Rotating a secret** takes `--restart`. A new password or PSK changes no artifact (its
value lives in the config and, for Docker, only in the environment), so the ordinary
"unchanged, nothing to do" path cannot see it; `deploy broker --restart` recreates the
container (Docker) or restarts the service (Podman) to pick it up, and the no-op message
names that as the way to apply one.

**Config source.** The container platform has no separate config namespace: its post-deploy
`config` steps read the platform-neutral `broker.*` fields -- `broker.domainCerts`,
`broker.productKeys`, `broker.diagDir`, `broker.cliScriptsFolder` -- shared verbatim with
Kubernetes, plus `tls.cert`/`tls.certKey` (server certificate) and `admin.user`/`admin.pass`
(SEMP login); the `nodes.*` names drive role detection for `config leader` /
`smoke redundancy`, and `nodes.backup.ip` is also read at verify time -- it is where those
two commands reach the backup's SEMP service from the primary. The rest of the `nodes.*`
table and `nodes.psk` are consumed earlier, at `prepare`/`deploy` (`prepare host` generates
the PSK and writes it back to the env file; `deploy broker` externalizes it as a secret). Container-only knobs live under `docker.*` /
`podman.*` (runtime, compose invocation, container name, data dir, network mode, rootless).
See `solace-util examples full`, or the same text at [env/sample.yaml](../env/sample.yaml).

## Rendering without applying

`generate <target>` is the one, dedicated way to review the exact artifact a command would
apply before it touches a cluster or a host. There is no `--dry-run` and no per-command
render flag. It prints to stdout and changes nothing: it runs no external command at all,
which is what makes it the safe way to inspect an env file you did not write (see
[The command fields are executable content](configuration.md#the-command-fields-are-executable-content)):

```
solace-util generate broker -e dev.yaml                          # the PubSubPlusEventBroker CR (kubernetes)
solace-util generate operator -e dev.yaml                        # the operator bundle (kubernetes)
solace-util generate secrets broker -e dev.yaml                  # the Secret manifests (kubernetes; secret values!)
solace-util generate secrets operator -e dev.yaml                # the operator's pull secret (kubernetes; secret values!)
solace-util generate broker primary -e dev.yaml --platform docker    # the compose file
solace-util generate broker primary -e dev.yaml --platform podman    # the quadlet unit
solace-util generate secrets broker -e dev.yaml --platform docker    # commands that supply the secrets
```

`generate` is a command with a named target rather than a flag on `deploy broker`, on
purpose: an artifact you meant to inspect and a cluster you meant to change should not be
one typo apart.

**Secrets are never part of a deployment artifact.** Each one lives in podman's secret
store, in a host environment variable the compose file names, or in a Kubernetes Secret --
and the quadlet unit, compose file, CR, and operator bundle reference it by name only. So
`generate broker` and `generate operator` output is safe to review, diff, and share, while
**`generate secrets broker` and `generate secrets operator` print the values themselves**
and must be handled exactly like the env file.

That split is why `deploy operator` issues three applies rather than one: the operator
namespace, then the image-pull secret into it, then the rest of the bundle. Everything it
applies is byte-for-byte what those two `generate` commands print, and the ordering is what
makes a first install work -- the secret is namespaced, and the namespace only exists in
the bundle.

The broker's own settings -- routername, redundancy, the scaling knobs -- are inlined into
whatever `generate broker` renders: `Environment=` lines in a quadlet unit, an
`environment:` block in a compose file, `spec.systemScaling` in the CR. There is no
separate command for them, because there is no separate artifact.

**Every platform hands the broker its secrets as files**, read through the setting's
`*filepath` variant, and named after the setting they feed -- so the layout inside a
container matches the data keys of the equivalent Kubernetes Secret:

| Secret | In-container path | Host-side name |
| --- | --- | --- |
| `admin.pass` | `/run/secrets/username_<admin.user>_password` | `<container.name>-admin-password` |
| `admin.additionalUsers[].password` | `/run/secrets/username_<username>_password` | `<container.name>-user-<username>-password` |
| `nodes.psk` (HA) | `/run/secrets/redundancy_authentication_presharedkey_key` | `<container.name>-redundancy-psk` |
| `tls.certPassphrase` | `/run/secrets/tls_servercertificate_passphrase` | `<container.name>-tls-passphrase` |

The host-side name carries `container.name` (default `solace`) so two brokers on one host
never share a podman store entry or a compose variable. On Kubernetes the operator mounts
the credentials Secret itself, so the only data keys that matter are
`username_admin_password` and `username_monitor_password`.

## Removing a broker: what stays, what goes

`remove broker`, `remove operator` and `remove all` are the destructive commands, and each
one keeps the layer that is expensive or impossible to get back **by default**: `remove
broker` / `remove all` keep the broker's persistent data (Kubernetes PVCs, or the
container's data directory); `remove operator` keeps the operator's CRDs. Deleting the CRDs
is the sharper of the two, because they are cluster-wide -- it cascade-deletes **every**
PubSubPlusEventBroker in the cluster, including ones this env file has never heard of, not
just the one it describes. So `--delete-crd` refuses outright while any PubSubPlusEventBroker
custom resource still exists anywhere in the cluster -- each is named `<namespace>/<name>` in
the refusal, and a failure to even list them refuses too. Only the CRD layer is refused: the
operator's own controller Deployment is removed either way. `remove all` also leaves the operator itself installed, for the
same reason: it is cluster-scoped and may be serving other brokers, so removing it is always
its own explicit command (see [Bringing up a fresh cluster](#bringing-up-a-fresh-cluster)).

Both removals ask about their retained layer the same way, so learning the contract on one
teaches the other:

- **`--delete-data`** (on `remove broker` / `remove all`) or **`--delete-crd`** (on `remove
  operator`) deletes the layer without asking.
- **`--no-prompt`** asks nothing at all: it confirms the removal and takes the safe answer
  to the layer question, so the layer is kept unless a `--delete-*` flag says otherwise.
  The two flags answer different questions and deliberately **compose** -- a fully
  unattended removal that also drops the data is `--delete-data --no-prompt`.
- **Interactively, with neither flag**, you are prompted and told what the layer is and
  what deleting it costs; only an exact, case-insensitive `yes` deletes it, and anything
  else -- including the lenient `y` that answers the removal prompt below -- keeps it.
- **Non-interactively with neither flag**, the layer is kept. A scripted or piped removal
  can never lose data by omission.
- **Either way, the outcome is printed** -- what was kept, or what was deleted -- so it is
  never left to be inferred from silence.
- **On `remove operator`, the CRD question is not even asked when a broker still exists.**
  Its answer is already fixed to "keep" -- deleting the CRDs would cascade-delete that
  broker along with every other one in the cluster -- so prompting would invite a "yes" this
  tool will not honour. A warning names what was found instead. An explicit `--delete-crd`
  still runs straight into the refusal above rather than being silently downgraded to
  "kept": naming the flag earns a loud failure listing the brokers in the way, not a silent
  no-op.
- **On Kubernetes, a PVC delete that fails is reported as a failure.** `--delete-data` (or
  `remove all`) still deletes every role's PVC and keeps going even if one delete fails, but
  if any did fail after `--ignore-not-found` already absorbed the benign "already gone" case
  -- an RBAC denial, a stuck finalizer -- `remove broker` now returns an error naming every
  PVC that survived, instead of reporting the data as gone.

**Docker and Podman have their own teardown wrinkles**, on top of the `--delete-data` layer
above:

- Podman's secret store is a real, separate persistence layer too -- `remove broker` /
  `remove all` now remove every secret `deploy broker` loaded into it as part of removing
  the container, so they no longer survive a teardown the way they used to. A secret that
  is already gone (or fails to remove) only warns; it does not stop the removal.
- If `systemctl stop` (Podman) or `stop` (Docker) fails, the removal now aborts before
  touching the unit, the container, or the data directory -- it used to warn and carry on
  regardless, which could delete a running broker's data out from under it. It continues
  only when the broker is *confirmed* down: systemd reporting the unit `inactive`, `failed`
  or `unknown`, or the engine listing no running container. "Still running" and "could not
  be confirmed either way" both abort, because the case this guard exists for -- a rootless
  Podman whose systemd user session is unreachable while `podman info` still succeeds --
  produces the second one, and silence is not confirmation. A stop that fails only because
  there was nothing running to stop is unaffected.
- `remove all` (which is exactly `remove broker` on these platforms) against a host with
  nothing deployed -- no compose file, no container by that name -- is now a no-op rather
  than an error, matching `--ignore-not-found` everywhere else in the tool.

**`remove namespace` (and the namespace half of `remove all`) refuses rather than asks when
the namespace holds work this tool did not create.** Four Kubernetes system namespaces --
`default`, `kube-system`, `kube-public`, `kube-node-lease` -- are refused outright and
unconditionally, with no flag to override it: there is no interpretation of "delete
kube-system" this tool should ever carry out. For every other namespace, its contents are
enumerated and classified OURS or FOREIGN before the delete is even considered:

- OURS is anything named `<kubernetes.name>-pubsubplus` (every object the operator creates
  off the broker), the broker custom resource itself, a Secret matching one of the three
  configured secret names, and the two objects Kubernetes puts in every namespace on its own
  (the `kube-root-ca.crt` ConfigMap, the default ServiceAccount).
- Anything else is FOREIGN -- most notably **a second Solace broker sharing the namespace**:
  its objects carry the same operator-assigned `-pubsubplus` suffix as ours do, but the
  match includes the broker's own name, so a different broker stays foreign rather than
  being swept up as ours.
- A single foreign object refuses the **whole** delete, listing what was found. This is a
  hard refusal, not a prompt -- `--no-prompt` cannot pass it, because there is no question
  left to silence. When it is refused, `remove broker` and `remove secrets` still remove
  this deployment's own objects; deleting the namespace itself is then your own `kubectl
  delete namespace` call to make, once you have confirmed what else is in there.
- A failure to even list the namespace's contents refuses too -- being unable to see what is
  in there is not permission to proceed -- but a failure to list the broker custom resource
  only warns, and the teardown proceeds by what it could see: an absent CRD is exactly the
  state this tool's own documented removal order (brokers, then `remove operator
  --delete-crd`, then `remove namespace`) produces, and refusing there would make a
  legitimate teardown impossible.

This is a second, independent decision from *whether to remove the broker (or operator) at
all*. **Every command that destroys something confirms first** -- `remove broker`,
`remove operator`, `remove secrets`, `remove namespace`, `remove all`, and `restart broker`,
which drops every in-flight connection (deletes pods on Kubernetes, bounces the container on
Docker/Podman). This is now the same on all three platforms; `restart broker` used to run
unconditionally on Docker and Podman, and on Kubernetes it now asks with the verb "Restart"
rather than "Delete", matching the container side -- the mechanism is unchanged (it deletes
the pod so the StatefulSet recreates it), only the prompt's wording. An interactive terminal
is asked `[y/N]`; a non-interactive session without `--no-prompt` refuses loudly rather than
destroying anything unattended.

**On Kubernetes, every one of those prompts also says where it will act.** The env file names
a namespace but never a cluster -- which cluster a `kubectl` call actually reaches is decided
entirely by the kubeconfig's current context, so a `dev.yaml` run against a context that has
drifted to `prod` looks identical right up until the delete lands. That fact is now surfaced
twice. First, the startup preamble prints `==> kube-context: <name>` beside the existing
`==> using kubectl: <path>` line, read from `kubectl config current-context` -- a kubeconfig
read, not a cluster round trip, so it costs nothing and still prints against an unreachable
cluster. Second, the prompt itself repeats both facts: `remove broker`, `remove secrets` and
`restart broker` read "... in namespace `<ns>` (context `<ctx>`)"; `remove namespace` and
`remove all` already name their own namespace as part of the sentence, so they append only
the `(context <ctx>)` clause. `remove operator` is the one exception -- it names the
**operator's own** namespace rather than repeating the broker's, since the two can differ and
naming the wrong one would be worse than naming none. The context clause is dropped entirely,
never rendered as an empty `(context )`, on a kubeconfig with no current context or on a
non-Kubernetes platform.

There is no global `--yes`. `--no-prompt` is the one flag that silences all of it, so a
script switches off one thing rather than one per question. It still cannot lose you data on
its own: the layer stays unless `--delete-data`/`--delete-crd` names it. Two flags, two
questions, on purpose -- dropping messaging data and cascading a CRD deletion across the
cluster are each too costly to answer as a side effect of the other.

## Upgrading a running broker

Changing the image tag (or any other setting) is the same edit on every platform --
bump `image.tag` in the env file -- but applying it differs:

**Kubernetes, `updateStrategy: automatedRolling` (the default)**

```
solace-util deploy broker -e dev.yaml
```

`deploy broker` re-applies the custom resource; the operator sees the new tag and rolls the
pods itself (monitor, then backup, then the active node).

**Kubernetes, `updateStrategy: manualPodRestart`**

```
solace-util deploy broker -e dev.yaml     # updates the statefulset template; no pod is touched
solace-util restart broker -e dev.yaml    # bounces monitor -> backup -> primary, waiting for each
```

The operator deliberately waits for you here, so `deploy broker` alone changes nothing
visible. `restart broker --pod <role>` bounces one pod if you would rather drive the order
yourself -- `--pod` is the one way to name a Kubernetes pod, here and on `cli`, `shell`,
`logs broker`, `check semp-login` and `status broker` too; worth doing after a failover,
since the order above is by configured role and the active node may not be the configured
primary. Check with `solace-util smoke redundancy` first.

**Docker / Podman** (on each host, with its own role)

```
solace-util deploy broker primary -e prod.yaml --restart    # prod.yaml: a podman env file
solace-util deploy broker -e prod.yaml --restart             # prod.yaml: a docker env file
```

`deploy broker` compares the rendered artifact with the one on disk: unchanged is a no-op,
changed is written and then applied to the running broker -- with `--restart`, or
after being asked. Without consent the new artifact is left in place and the command
says the broker is still on the previous one. In an HA group, upgrade the monitor and
backup before the primary.

## Troubleshooting

### File descriptors on rootless podman

Both container artifacts ask the engine for `<docker|podman>.container.ulimits.nofile`
(default `2448:1048576`). A **rootless** container cannot raise `nofile` above the hard limit
of the user invoking podman -- the kernel refuses -- so `prepare host` checks it on a podman
env file and stops with the exact drop-in to add when it is too low:

```
solace-util prepare host -e env/prod.yaml
...
error: rootless podman: this user's hard nofile limit is 1024, but
podman.container.ulimits.nofile needs 1048576 -- a rootless container cannot raise it
above the user's own hard limit, so the broker would start under-provisioned.
  Add this as root to /etc/security/limits.d/99-solace.conf, replacing <user> with the
  account that runs the container, then log out and back in:
    <user> hard nofile 1048576
    <user> soft nofile 2448
```

Prep reports and refuses rather than fixing it: raising a hard limit means editing host-wide
security configuration as root, which is what a rootless deployment exists to avoid. Rootful
podman and docker are unaffected -- their privileged engine raises the limit itself -- so the
check runs only for `podman.rootless: true`.

### Podman secret flags

`deploy broker` stores the broker's secrets with `podman secret create --replace` and mounts
them into the container (`type=mount`), neither of which the oldest podman builds support.
Confirm yours does (`podman secret create --help | grep -- --replace`) -- `check deploy` only
proves the runtime answers `version`, so an unsupported flag surfaces at deploy time.

### Docker compose secrets need compose 2.23.1+

The generated compose file sources each secret from a host environment variable, which needs
**compose v2.23.1 or later** (`docker compose version`). On an older compose the
`environment:` secret source is not understood and `deploy broker` fails. On a host carrying
only the standalone v1 binary, set `docker.compose: docker-compose`.

### Wrong cluster (kubeconfig drift)

The env file names a namespace, never a cluster: which cluster every `kubectl` call reaches
is decided entirely by the kubeconfig's current context, so a `dev.yaml` run against a
context that has drifted to a prod cluster looks identical to a real dev run right up until a
destructive command lands. Before answering `yes` to any removal or restart prompt, check its
`(context <name>)` clause -- or the `==> kube-context: <name>` line printed at startup -- against
`kubectl config current-context` (list all of them with `kubectl config get-contexts`), and
switch with `kubectl config use-context <name>` first if it does not match what you expect.
`--kubeconfig`/`KUBECONFIG` pointed at the wrong file produces the same symptom: the tool has
no `--context` flag of its own, so whatever the kubeconfig currently resolves to is where the
command runs.

### A wrapper runtime is refused

`kubernetes.runtime`, `docker.runtime`, `podman.runtime` and `docker.compose` accept only an
allowlisted bare binary name. A wrapper such as `microk8s kubectl` or `lima nerdctl` needs
`--allow-command <name>` on the invocation -- see
[The command fields are executable content](configuration.md#the-command-fields-are-executable-content).

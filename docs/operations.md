# Operations

Running a broker after the [quick start](../README.md#quick-start): what each phase does,
what re-deploying and removing actually change, and how to upgrade. For the complete command
surface see [commands.md](commands.md); for the env file see
[configuration.md](configuration.md).

- [Exit codes](#exit-codes)
- [The preflight](#the-preflight)
- [Version floors](#version-floors)
- [Bringing up a fresh cluster](#bringing-up-a-fresh-cluster)
- [Post-deployment configuration order](#post-deployment-configuration-order)
- [Docker and Podman mechanics](#docker-and-podman-mechanics)
- [Rendering without applying](#rendering-without-applying)
- [Removing a broker: what stays, what goes](#removing-a-broker-what-stays-what-goes)
- [Exporting and importing configuration](#exporting-and-importing-configuration)
- [Data replication](#data-replication)
- [Upgrading a running broker](#upgrading-a-running-broker)
- [Troubleshooting](#troubleshooting)

## Exit codes

| Code | Meaning |
| --- | --- |
| 0 | It did what it was asked. |
| 1 | It was asked for something sensible and could not do it -- an unreachable cluster, a refused permission, a broker that would not come up. |
| 2 | The request itself was wrong: the command line, or the env file it named. |

**A script may retry a 1. It must never retry a 2.** That is the whole point of the split:
nothing about the world is going to change the outcome of a mistyped flag or an env file
that fails validation, so a retry loop around one spins forever. The line between the two
is "would a different invocation have helped?".

An env file that cannot be found, cannot be parsed, fails validation, declares no platform
section or declares several without `--platform` is therefore **2**, not 1 -- you chose that
file, and no amount of waiting fixes it. So is a command or flag that does not apply to the
platform the file selected, a bad `--pod`/`--platform`/`--since` value, and
`--allow-command` on a command that renders without executing.

**One documented exception:** a mistyped *top-level* command exits **1**, not 2
(`solace-util depoy broker`). Cobra produces that error before this tool can classify it,
and the available fix is worse than the gap -- it would make a mistyped command print help
and exit **0**. A mistyped *sub*-command (`remove bogus`, `broker logs bogus`) is 2 as
expected.

**`cli` and `shell` pass their session's status through.** These two hand your terminal to
a session inside the broker, so the exit status of whatever you ran last in there becomes
this tool's exit status -- collapsing it to 1 would throw away the answer you asked for. No
other command does this: everything else runs kubectl or a container engine as an
implementation detail, and leaking those codes would make the table above meaningless.

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

## Version floors

The preflight above proves the runtime *answers*. It does not prove the runtime is new enough
to understand what this tool will ask of it. These are the floors, and **nothing enforces any
of them at runtime**:

| Component | Floor | What needs it | If older |
| --- | --- | --- | --- |
| Docker Compose | 2.23.1 | The `environment:` secret source in the generated compose file | `broker deploy` fails on the secret source. Loud. On a host with only the standalone v1 binary, set `docker.compose: docker-compose` |
| podman | 4.5 | `secret rm --ignore`, `secret create` reading the value from stdin, and `Secret=...,type=mount` in the quadlet unit | The unknown flag or directive surfaces at deploy time. Loud. This tool deliberately avoids `secret create --replace`, which would work but needs **4.7**, so one flag would raise the floor of the whole tool; remove-then-create is idempotent the same way |
| Solace broker image | 10.26 | The built-in readiness endpoint the opt-in health check uses with no `cmd` of its own | Refused at load: an older tag and an unidentifiable one are both rejected. This is the one floor that IS checked, because the tag is in the env file rather than on the host |
| EventBroker operator | bundled 1.4.2 | The `PubSubPlusEventBroker` schema this tool renders | `operator deploy` installs the bundled version. An older operator already in the cluster prompts before a downgrade |
| kubectl / oc | none | Namespace, Secret and the custom resource are all core API shapes plus the operator's own CRD | n/a |
| Go | 1.27 | Building the binary, pinned so builds use a patched standard library | n/a |

Why nothing enforces them: `version` reports this tool's own build, the preflight probes
reachability and one permission, and `validate` proves only that the runtime answers
`version`. Parsing an engine's version string is its own class of problem, so the floors are
documented rather than gated. The consequence worth knowing is that a floor failure is loud on
every row above, so an under-versioned engine tells you at deploy time rather than running a
broker that quietly cannot read its own credentials.

## Bringing up a fresh cluster

The EventBroker operator is cluster-scoped and shared between brokers, so it is installed
and removed on its own rather than as a side effect of any one broker's `broker deploy`
or `broker remove`. A cluster that has never run this tool (or any other operator install)
needs it once:

```
solace-util operator deploy -e dev.yaml
```

After that, any number of env files can each `broker deploy` their own broker against the
same cluster. Running `operator deploy` again from a second env file does not replace the
watch list -- it UNIONS the namespaces, so the operator keeps reconciling the brokers that
were already there. `validate` warns rather than fails when the operator or its CRD looks
missing, so that warning is what tells you this step was skipped -- `broker deploy` then
fails once it tries to apply a custom resource the cluster does not know how to reconcile.

`validate`'s report opens with a **Config** section, ahead of Deployment: it states that
the env file was already accepted -- `config.Load` validates before this report can even
start building, so reaching the report means it passed; this is not a fresh check running --
then lists the resolved `kubernetes.ports` as `name=port` pairs behind a leading count (e.g.
`ports (17)` for the default set), so what will actually render into the broker CR is visible
without deploying or running `broker generate`. Every row in this section is informational
and can never fail the check.

The rest of the report only runs once the cluster answers: an unreachable cluster marks
everything after it `[SKIP]` rather than reporting checks it never made. From there it
checks permission to create the broker resource, reads the installed operator to see
whether it actually watches this broker's namespace, and checks the target StorageClass's
volume-binding mode and whether it allows volume expansion. On docker and podman there is
no operator, so that section of the report is omitted rather than failed.

End to end, a first run against a brand-new cluster looks like:

```
solace-util operator deploy -e dev.yaml                  # once per cluster
solace-util validate -e dev.yaml                         # cluster, StorageClass, operator
solace-util broker deploy -e dev.yaml                    # namespace -> secrets -> CR
solace-util broker perform semp-login-check -e dev.yaml  # prove it answers
solace-util broker perform assert-leader -e dev.yaml     # HA only, once the pods are up
```

On Kubernetes, `broker deploy` reads the CR back after applying it, since `kubectl apply`
exiting 0 does not prove the object actually exists. Before it applies anything,
`operator deploy` also checks permission for every resource kind the bundle needs -- not
just the CRD -- so a missing permission is caught before the apply is half done rather than
partway through it; an operator already watching every namespace is left alone rather than
narrowed, and when the watch list comes out unchanged the rendered controller Deployment is
byte-identical, so it is not rolled for nothing.

Removing a broker never removes the operator either -- see
[Removing a broker: what stays, what goes](#removing-a-broker-what-stays-what-goes) below.
Uninstall it explicitly, and only once nothing else in the cluster still depends on it:

```
solace-util operator remove -e dev.yaml
```

`operator stop` is occasionally exactly what you want: editing a StatefulSet the operator
owns is otherwise a fight with the controller reconciling it back. `operator restart` is for
an operator that is wedged rather than out of date -- `operator deploy` is what re-applies a
changed bundle. `operator logs` does not offer `--previous`: a Deployment's pod name changes
on restart, so there is no prior container left to read, unlike `broker logs` on Kubernetes.

## Post-deployment configuration order

Everything under **`broker configure` and `broker perform` is post-deployment**: each step
drives the Solace CLI inside a broker that is already running, so none of it is part of
`broker deploy`. Wait for the broker to be ready (the pods, or the container/service), then
run these in order.

They are not uniformly re-runnable, so the order lives here rather than in a
run-everything command that would stop partway through a second run. On a fresh broker:

1. `broker perform assert-leader` (HA only; on containers, run it on the primary -- it also reverts the
   backup over SEMP first, so `redundancy.backup.addr` should be reachable)
2. `broker configure server-certs` (when TLS is configured)
3. `broker configure domain-certs` (when any `broker.domainCerts.dirs` or `.files` are listed)
4. `broker configure default-vpn`
5. `broker configure default-users`
6. `broker configure product-keys` (a no-op when none are listed, like `domain-certs`)

`broker configure data-replication` belongs to that phase too, but only when this broker is
half of a DR pair, and it is run at both sites -- see
[Data replication](#data-replication).

Steps 3, 4 and 5 can be undone from here: `broker configure domain-certs --remove`,
`broker configure server-certs --remove`, `broker configure product-keys --remove`, and
`--enable` on either hardening step -- `no ssl server-certificate` and `no product-key
<key>` are confirmed CLI forms, so both removals now run for real instead of refusing.

Each step's own scope and failure mode is worth knowing. `server-certs` applies to every
node in the redundancy group on Kubernetes, or this host's container on docker/podman; if
the env file names a Secret it supplies no certificate files for, that Secret is assumed
managed elsewhere (kubectl, cert-manager) and the command refuses rather than guessing, and
with `kubernetes.tlsServerSecret` set, `--remove` refuses too -- the operator would put the
certificate straight back -- so clear the key instead. `domain-certs` with neither `dirs` nor
`files` configured is a logged no-op, not an error; a configured `dirs` entry that cannot be
read (missing, or unreadable) IS an error, before anything is uploaded -- it is not skipped.
Every directory is walked one level deep (no subdirectories) for its matching files
(`.cer`/`.crt`/`.pem` by default, or the entry's own `fileExt`), each named
`<last-directory-element>_<filename>`; two certificates that would resolve to the same name
-- across directories, or against an explicit `files` key -- fail the command naming both
source paths. Override a name you do not want by listing that certificate explicitly under
`files` instead. `--remove` walks the same `dirs`/`files` to learn the CA names to withdraw,
so it needs the same directories to still be readable; if a directory has been removed since
`domain-certs` last ran, point `dirs` at wherever the certificates live now (or list the
remaining ones under `files`) before removing. `product-keys` is never applied to the monitor node, which
carries no message spool -- `--pod` narrows it to one node and warns, since a partly-licensed
redundancy group is usually a mistake -- and the broker's CLI output is scanned for errors so
a rejected or nonexistent key is reported as a failure rather than silently accepted. With no
keys configured it says so and does nothing, on both directions, exactly as `domain-certs`
does; the script never echoes the keys back, since the transcript is shown.
`default-vpn` does not touch the default client-username -- that is `default-users`, which
covers every VPN rather than only this one -- and in HA, config-sync replicates both, so
either only needs to run on one node. `default-users` reads the VPN list live from the
broker (`show message-vpn *`), not from the env file, since which VPNs exist is broker
state; parsing zero VPNs is a warning, not a failure.

**The broker stops at the first rejected line, and the exit code says so.** Every
`broker configure` step and `broker perform cli-script` runs its script through the
broker's own `source script ... stop-on-error no-prompt`, the same wrapper
`import-config` applies each chunk with. A Solace CLI script is a sequence of
independent commands, so without that wrapper one rejected line let every later line
run on top of it; with it the broker stops there. The output is always shown and the
uploaded script always removed, whatever happened.

The command then fails if the tail of that output carries one of the rejection keywords
-- one list for the whole tool, shared with `import-config` rather than kept correct
twice. Because the broker stopped, the rejection is the last thing it printed, which is
what makes scanning the tail sound. The error names the keyword it found and never the
line, since a CLI transcript can carry passwords; read the printed output for the
detail.

**This is a breaking change for automation**: a CI job that runs a `broker configure`
step or `broker perform cli-script <file>` and gates on the exit code -- and passed
before -- can start failing, because a rejected line is now caught instead of silently
accepted, and less of the script runs after one. Check any pipeline that runs these
commands before relying on the new exit code.

A few of the other `broker perform` steps have their own operational details. `assert-leader`
in HA first reverts activity to the primary and waits for redundancy to report Up with the
primary active before asserting leadership; on docker/podman it runs only on the primary host
and fails loud on a backup or monitor host, and it reverts the mate over SEMP first but
downgrades an unreachable mate to a warning, since its own job is local. `semp-login-check`
passes credentials on stdin as a curl config file so the password never reaches an argv,
process list or log, and a failed login is reported as a failure of the login itself -- the
request was made and answered, and the answer was no. `gather-diagnostics` deletes the
helper scripts it uploads on every path out, and the in-broker archive once it is safely
downloaded; a cleanup failure only warns rather than failing the collection, but a failed
DOWNLOAD fails the command and leaves the bundle on the broker, naming the path to fetch it
from -- deleting it there would destroy the only copy of what was asked for. `--pod` narrows
collection to one node. `cli-script`'s in-broker name is the file's own base name
(split on both path separators), so one env file cannot name two different files depending on
which host drove it. `shell-script` deletes the script it uploaded once the run finishes,
even when it failed, and a script that echoes a secret prints it in the output; neither
`shell-script` nor `broker shell` validates or reports on what runs beyond that -- they are
the escape hatch for what this tool does not model.

### Extra CLI users differ by platform

`semp.additionalUsers` reaches the broker on every platform now, declaratively -- the
users exist from its first boot, with nothing to run afterwards. What differs is how the
password gets there:

- **Docker / Podman** -- the password is a mounted FILE
  (`/mnt/secrets/username_<username>_password`, pointed at by
  `username_<username>_passwordfilepath`), and the access level rides the artifact as a
  plain setting. Each user gets its own engine secret.
- **Kubernetes** -- both halves ride the pod ENVIRONMENT, from a Secret of its own named
  `<kubernetes.name>-additional-users` that the CR references in
  `spec.extraEnvVarsSecret`.

**Kubernetes is the one place a password reaches the broker as an environment variable
rather than a file, and that is the CRD's constraint, not a choice.** `spec` offers
`extraEnvVars`, `extraEnvVarsCM` and `extraEnvVarsSecret` and no volume passthrough, so
there is no way to mount an arbitrary Secret as files -- `username_<u>_passwordfilepath`
would name a path nothing creates. Anyone who can exec into the pod can read these
passwords from its environment; if that is unacceptable for a given user, do not list them
here. `broker validate` states where they land rather than leaving it to be discovered.

Two consequences worth knowing:

- **The Secret is separate from `kubernetes.adminSecret`, and must be.** `extraEnvVarsSecret`
  is projected with `envFrom`, which turns EVERY key of the Secret it names into an
  environment variable -- pointing it at the credentials Secret would publish the admin and
  monitor passwords too, just to get the extra users in. (Extra `username_<user>_password`
  keys in the credentials Secret are ignored by the operator anyway; verified against a live
  cluster.)
- **Usernames are stricter here.** The kubelet silently DROPS environment variables whose
  names are not letters, digits and underscores, so a username carrying `.` or `-` would
  produce a user with no password, or no user at all. That is refused at load on Kubernetes
  and still allowed on the container platforms, which mount a file.

The broker reads these settings at boot, so changing a password takes effect when the pod
next restarts.

That left the CLI, which brought two problems the replacement has to solve:

- **It was not re-runnable.** The broker's `create username` fails if the user exists, and
  that was reported rather than reconciled -- re-setting a password an operator rotated on
  the broker would be worse.
- **The password charset was restricted.** The value went onto a CLI line, and the broker
  rejects ``:()";'<>,`\*&|`` inside it. That constraint still applies on docker and podman
  only through the file they write, so those platforms accept the full charset.

## Docker and Podman mechanics

The `docker` and `podman` halves of the command tree share one implementation for a
**host-local** broker: one container per host, driven over `<runtime> exec`/`cp` (no
operator, no cluster). Only the deploy artifact differs -- Docker renders a compose file
and brings it up with `docker compose`, Podman a systemd **quadlet** `.container` unit.
(Docker always deploys through compose -- there is no `docker.mode` key to choose a mode
with. A bare `docker run` cannot recreate an existing container, so re-deploying after an
image-tag bump would fail on a name conflict where compose recreates cleanly. An env file
carrying a `docker.mode` key fails strict decoding as an unknown field.)

On a fresh host, `broker deploy` also creates and takes ownership of the data directory,
confirms the redundancy hostnames resolve, and logs in to the configured registry, before
creating the engine secrets and writing and starting the artifact.

**The compose project name is declared, not derived.** The generated compose file carries a
top-level `name:` taken from `docker.container.name`, lowercased with anything outside
`[a-z0-9_-]` folded to `-`. Without it, compose names the project after the *directory* the
file happens to sit in, so generating the artifact somewhere else -- or renaming that
directory -- silently starts a new project and leaves the old containers, network and
volumes as orphans `docker compose down` can no longer find. Declaring it in the file (rather
than passing `docker compose -p`) also keeps it visible in `broker generate`, and leaves
`COMPOSE_PROJECT_NAME` working as an override: compose's own precedence puts that environment
variable above a file's `name:`. Export it before `broker deploy` if you need a specific
project name -- for instance to adopt containers an earlier deployment created under the
directory-derived name.

**`broker status` reports health and restarts.** (On Kubernetes the same command instead
reports the operator's CR conditions, then pods, Services and StatefulSets.) The `ps` line above it is deliberately
narrowed to NAMES, IMAGE and STATUS: a broker publishes a dozen or more ports, and the
engine's default PORTS column spends most of a terminal line listing both host bindings of
each one, wrapping every other column into illegibility on the one report whose job is to
answer "is it up". The ports are in the env file that chose them and in the artifact
`broker generate` prints. Docker no longer runs `compose ps` first either -- it listed the
same single container, carried the same ports, and compose's own `--format` takes only
`table` or `json`, so it could not be narrowed the way the engine's `ps` can.

After that line it prints a
decoded block: name, image, state, health, restart count, and every path mounted in. Two
fields there are not what the engine's own `ps` or a `--format` template would give you.
*Health* is read under whichever spelling this engine uses -- docker nests it at
`State.Health`, podman at `State.Healthcheck` -- and "no healthcheck configured" is printed
distinctly from "unknown", because those are a deployment choice and a missing answer. On
podman the *restart count* comes from `systemctl show -p NRestarts`, not from the engine:
systemd restarts a quadlet unit by replacing the container, so the engine's own counter
reads 0 on a broker that has restarted twenty times. `broker status --all` discovers
containers by image, and any it finds that this env file does not name gets the engine's
counter instead -- this deployment's unit is not their unit. `--detail` adds the full
`inspect` dump on top. The container's **environment is never printed** on either platform:
on docker the compose secrets are environment-sourced, so secrets appear only as the mount
paths that name them.

`broker logs --previous` is Kubernetes only, because a container engine keeps no prior-run
log to read; `--since` is re-serialised through Go's duration parser before it reaches
argv, so e.g. `90m` becomes `1h30m0s` in the command a `-v` trace shows -- expected, not an
error.

**HA verification runs from the primary.** The transport is node-local, so exec reaches
only this host's broker -- but the broker itself is a control channel: the primary's own
`show redundancy` already reports the mate's activity, and the one command that must land
on the backup (`redundancy revert-activity`) is sent over SEMP to `redundancy.backup.addr`. Bring
the group up by running `broker deploy --pod <role>` on each host with its own role -- or
omit `--pod` and the role is detected in two passes, announced on stderr: this host's
hostname against `redundancy.*.name` (case-insensitively, tolerating an FQDN on either
side), then -- only if that matched nothing -- this machine's own interface addresses
against `redundancy.*.addr`. The second pass is the cloud case: an instance reports
`ip-10-0-0-12` while the env file names the broker `bkp-host`, so the routername and the
OS hostname are legitimately unrelated and the address is the only thing both ends agree
on. Matching neither fails loud, and matching more than one role fails loud too, naming
every role it matched and how. An explicit role always wins over detection, which is the
escape hatch for a host that matches nothing -- but it is CHECKED against what the host
looks like, and a disagreement WARNS and proceeds. It never prompts: the operator said
which node this is, and a deploy scripted across three hosts must not stop to ask.
`broker generate` detects the same way but does NOT fail on an unrecognised host: it renders
the primary's artifact with a warning, because it changes nothing and the artifact names the
node it is for, so the fallback is visible in the output you are about to read. Refusing there
would make reviewing another node's artifact from a laptop impossible.
Then run `broker perform redundancy-test` **once, on the primary host**: it confirms the primary healthy, checks
the backup's SEMP service is reachable (before anything is disturbed), releases and
un-releases activity so the backup takes over, then reverts the backup and waits for
activity to come home. Backup and monitor hosts are rejected loud, and `broker perform assert-leader`
runs only on the primary. Prerequisite: the backup's SEMP port (8080, or the mapped host
port under `network.mode: bridge`) must be reachable from the primary host -- a working HA
group only proves the redundancy ports (8300-8302, 8741, 55555) are open, so the
preflight fails loud with the address it tried when SEMP is firewalled.

**If a `broker perform redundancy-test` run dies partway** (including Ctrl-C, which runs no cleanup), the
group can be left released or failed over. Restore it with `no redundancy release-activity`
on the primary's CLI, or `redundancy revert-activity` on the backup's -- or simply re-run
`broker perform redundancy-test` once the cause is fixed.

Example (HA -- run each line on the matching host; `prod.yaml` is a podman env file):

```
solace-util broker deploy --pod primary -e prod.yaml   # on the primary host
solace-util broker deploy --pod backup  -e prod.yaml   # on the backup host
solace-util broker deploy --pod monitor -e prod.yaml   # on the monitor host
solace-util broker perform assert-leader -e prod.yaml        # on the primary only
solace-util broker perform redundancy-test -e prod.yaml     # on the primary only -- drives the whole group
```

Naming the role explicitly is what to do when a host matches no `redundancy.*` entry (or
matches more than one); when it does match, `solace-util broker deploy -e prod.yaml` alone
on each host detects and announces the same role.

**A standalone broker names itself after its host.** `redundancy.<role>.name` is the
routername and the container's hostname, and with one node it is always this machine --
so leaving `redundancy.primary.name` out of a standalone docker/podman env file is
supported, and the host's own OS hostname is used, reported as `==> routername not
configured; using this host's name: <host>`. It is settled once, at load, so the check
report, the DNS check, the rendered artifact and `validate` all name the same broker. HA
is the opposite and stays mandatory in all three: each name keys that node's entry in a
group table every host renders, and no host can fill in another machine's.

### Which user the broker runs as

Both artifacts always carry a `-u` equivalent -- quadlet `User=`/`Group=`, compose `user:` --
so `container.runUser` overrides whatever USER the image declares. The default is chosen by
platform, and for podman by `rootless`:

| Deployment | Default `runUser` | Why |
| --- | --- | --- |
| docker | `1000001:0` | the image's own user and group. The engine is privileged, so container uid 0 would be **host root** |
| podman, rootful | `1000001:0` | same reason |
| podman, rootless | `1000:0` | the image's uid is unreachable here (below), so a low id the stock subuid allocation covers |

The rootless split is arithmetic, not preference. A rootless container's uids are drawn from
the invoking user's subuid range, and a range has to reach the id itself -- so uid `1000001`
would need **1000002** entries against the 65536 `useradd` allocates by default. `1000` needs
1001, which a stock range covers.

**The gid is 0 everywhere**, which is the group the broker image expects and also the one gid
that costs nothing on either side: rootful maps it to the root group, and rootless maps
container gid 0 to the invoking user's own primary group. So a rootless deployment needs a
**subuid** allocation but no subgid one.

It is worth being explicit about what uid 0 means in each mode, because it is the usual
source of confusion:

| Container id | Rootful host id | Rootless host id |
| --- | --- | --- |
| uid `0` | **root** | the invoking user -- unprivileged |
| uid `1000` | 1000 | `subuid_start + 999`, e.g. 100999 |
| uid `1000001` | 1000001 | needs a hand-allocated range |
| gid `0` | the root group | the invoking user's own primary group -- no subgid needed |

So `runUser: "0:0"` is genuinely safe on rootless (it is just you) and is the setting to
reach for if you want the bind-mounted data directory readable from the host without
`podman unshare`. It needs no subuid allocation at all. The cost is that the broker runs as
container-root, with no in-container isolation.

Changing `runUser` on an existing deployment changes who owns the data directory. `broker
deploy` re-chowns it on every run, so a redeploy heals it -- but the broker must be restarted
to pick the new identity up, and on rootless the host-side ownership moves into the subuid
range.

### Rootless podman prerequisites

`podman.rootless: true` deploys as an ordinary user, and a rootless host needs preparation
most of which this tool deliberately does **not** perform. `validate` and the first step of
`broker deploy` read the same block, so there is one definition of a ready host; they differ
only in what they are allowed to do about it. `validate` never changes anything, as its help
promises:

| Row | What it asserts | If it fails |
| --- | --- | --- |
| `euid` | not root when `rootless: true`, root when `false` | every row below is skipped -- probed as root they would answer about the wrong account |
| `user session` | `XDG_RUNTIME_DIR` names a directory that exists | you are under `sudo`/`su`/cron, or the user has no session |
| `id mapping` | this user has a subuid/subgid allocation **and** it reaches `container.runUser` | an administrator allocates or widens it (below), then you run `podman system migrate` |
| `linger` | `loginctl show-user <uid> --property=Linger` reports `yes` | **`broker deploy` enables it for you**; `validate` reports it |
| `user systemd` | `systemctl --user` answers | `systemctl --user start podman.socket` |
| `data dir` | `container.dataDir`'s nearest existing parent is writable | move `dataDir`, or pre-create and chown it |
| `nofile` | the hard limit covers `container.ulimits.nofile` | [File descriptors on rootless podman](#file-descriptors-on-rootless-podman) |

**What decides whether a row is repaired: the privilege it needs, not how easy it is.** The
tool already performs several unprivileged host changes during prep -- `mkdir -p`, `chown`,
`podman unshare chown`, `podman login` -- and enabling linger for the invoking user is in
exactly that class. The rows it will not touch are the ones that would need root, because a
tool that escalated on your behalf would defeat the point of running rootless at all. Note
which way round the privileges actually go:

| Operation | Privilege |
| --- | --- |
| `podman unshare chown` (what prep runs) | must **not** be root -- it enters the user namespace rootless podman already runs in |
| `loginctl show-user ... --property=Linger` (what the check runs) | none; a D-Bus read of your own user |
| `loginctl enable-linger`, no username (**what prep runs**) | none in the ordinary case: it targets the caller's own account, which polkit's `set-self-linger` grants to an active session |
| `loginctl enable-linger <other-user>` | admin authentication -- and prep never issues this form, because it passes no username at all |
| widening `/etc/subuid`, `/etc/security/limits.d` | root, so these are reported and refused |

polkit can still decline the self form -- a session that is not active under some SSH and cron
setups, or a distribution that has tightened `org.freedesktop.login1.set-self-linger`. Prep
says so when that happens, and names the administrator's `loginctl enable-linger <user>` as
the way round it.

**Where subuid ranges come from.** Usually nowhere you have to think about: `useradd`
allocates one at account-creation time from `/etc/login.defs` (`SUB_UID_MIN` 100000,
`SUB_UID_COUNT` 65536 by default), so an ordinary interactive account already has
`you:100000:65536` and covers the `1000:1000` default with room to spare. The accounts that
arrive with **nothing** are worth knowing before you go asking: those made with `useradd
--system` (a plausible way to create a broker service account), directory accounts from
LDAP/AD/SSSD, which are not in `/etc/subuid` at all, and accounts predating the distribution
enabling auto-allocation. The `id mapping` row reports that case as a failure rather than
skipping it, because podman then maps only container id 0 and every non-zero `runUser`
becomes unusable with nothing on screen to explain it.

**What to hand an administrator.** The rows that need root print these, naming your account
literally so the command can be forwarded as-is; they are collected here so you can send them
on without re-running anything. First, as yourself, see what you have:

```
grep "^$(id -un):" /etc/subuid /etc/subgid
ulimit -Hn
```

Then the privileged steps, if those came up short:

```
# subuid/subgid range -- the failing row prints this line with the range and account
# already filled in; it reads /etc/subuid and /etc/subgid and proposes the next FREE
# block, so it does not overlap another account
sudo usermod --add-subuids <start>-<end> --add-subgids <start>-<end> <user>

# file-descriptor ceiling, in /etc/security/limits.d/99-solace.conf; needs a fresh login to take effect
<user> hard nofile 1048576
<user> soft nofile 2448

# only if container.dataDir is somewhere this user cannot create
sudo mkdir -p /opt/solace/data && sudo chown <user> /opt/solace/data
```

And finally, back as the rootless user, with no root at all:

```
podman system migrate
```

`podman system migrate` is what makes a changed subuid range take effect for containers and
storage that already exist. Setting `podman.container.dataDir` to something under your own
home (`~/solace/data`) removes the third step entirely.

**The session variables are set for you too.** `systemctl --user` and
`loginctl` reach the user bus through `XDG_RUNTIME_DIR` and `DBUS_SESSION_BUS_ADDRESS`, and
both are unset under `sudo`, `su` and bare cron. Those describe *this invocation* rather than
the host, so when either is empty it is derived -- `XDG_RUNTIME_DIR=/run/user/<uid>` and
`DBUS_SESSION_BUS_ADDRESS=unix:path=$XDG_RUNTIME_DIR/bus`. A value you already set is never
overwritten, and the report says which half came from where:

```
[ OK ] user session: XDG_RUNTIME_DIR=/run/user/1000 (derived) DBUS_SESSION_BUS_ADDRESS=unix:path=/run/user/1000/bus (derived)
```

The runtime directory is proven to exist **before** either variable is exported, because a
derived path to a directory that is not there turns a clear "you are under sudo, or this user
has no session" into an obscure systemctl failure several rows later. Every podman command
that drives systemd gets the same treatment, not just `validate` and `deploy` -- `broker
start`, `stop`, `restart`, `remove` and `status` never call the prep step but need the same
bus.

Boot-start is where linger bites hardest, and why prep enables it rather than asking: the
quadlet carries `WantedBy=default.target`, but without lingering systemd tears the user's
instance down at logout and starts nothing at boot, so the broker simply disappears when you
log out. It is worth knowing that linger is scoped to the **account**, not to this
deployment -- it makes every one of that user's services survive logout, and `broker remove`
does not turn it off again.

### Re-deploying is safe and explicit

`broker deploy` renders the artifact and compares it with
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
  itself fails (a transient engine hiccup), `broker deploy` now aborts naming the container
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
`solace-util broker deploy --pod <role> --restart -e prod.yaml` (podman) or
`solace-util broker deploy --restart -e prod.yaml` (docker). Podman needs `--restart`
because `systemctl start` on an already-active unit is a no-op, so without it the unit file
is rewritten but the running container keeps the previous image.

**Secrets.** `broker deploy` externalizes every secret before applying the artifact, so no
value is ever written into the compose file or quadlet unit (see the table under
[Rendering without applying](#rendering-without-applying) for the names and
paths). Podman loads them into its own secret store (`podman secret rm --ignore` then
`podman secret create`, value on stdin) and the unit mounts them; Docker's compose file names a host environment
variable per secret, and `broker deploy` sets those variables for its own `docker compose`
process, so no value ever reaches an argv or a file beside the compose file. A missing
value fails the deploy loudly rather than starting a broker without a password.
`redundancy.psk` never gets that far: it is mandatory in an HA container group and an empty
one is refused at LOAD, with the `openssl rand -base64 32` command in the message. Nothing
here generates it -- a key one host invented is a key the other two never see.

What that does and does not buy you: **nothing is written next to the artifact** (no
plaintext file a project-directory backup, `tar`, or non-root user would pick up, and
nothing to clean up on teardown), but the value still ends up at rest -- Docker
materializes each secret into the container's own filesystem as a `0444` root-owned file
at the absolute `target:` the compose file names (`/mnt/secrets/<setting>`), which is the
same at-rest exposure class as podman's store. It is not in the container's environment, so
`docker inspect` does not show it. That now includes the server certificate's **private
key**, which arrives on the same channel and lands at the same `0444` inside the container --
the same treatment the admin password already gets there, in a container that runs only the
broker.

**Podman is the one platform where this tool writes a secret to the host.** A quadlet unit
cannot inline file content the way a compose file can, so the server-certificate bundle is
written to `<podman.baseDir>/<container.name>-tls-servercertificate.pem` at mode `0600` in a
`0700` directory, and the unit bind-mounts it read-only with an SELinux relabel. That makes
it a second podman-side persistence layer beside the secret store, and it is why
`podman.baseDir` is mandatory rather than defaulted: where a private key lands on your host
is your decision, not this tool's. `broker remove` deletes the file as an artifact, like the
quadlet unit, and unlike a leftover store secret a failure to delete it is **fatal** --
silently leaving a private key behind is the outcome least like the rest of that teardown.
`--delete-data` is not what removes it. A teardown that refuses because the unit could not be
confirmed stopped deliberately leaves the file in place, since the container may still be
reading it.

**Nothing prints the secret values on a container platform any more.** There used to be a
renderer that emitted `podman secret create` commands and `export` lines for running compose
by hand; it went with the `generate secrets` sub-tree, because the only thing it could print
WAS the values and `broker deploy` creates them itself -- podman through its secret store,
docker by passing them to the compose child's environment. If you run `docker compose up`
by hand against the generated file, you must export the variables it names yourself; unset,
compose refuses.

**Rotating a secret** takes `--restart`. A new password or PSK changes no artifact (its
value lives in the config and, for Docker, only in the environment), so the ordinary
"unchanged, nothing to do" path cannot see it; `broker deploy --restart` recreates the
container (Docker) or restarts the service (Podman) to pick it up, and the no-op message
names that as the way to apply one.

**Config source.** The container platform has no separate config namespace: its post-deploy
`config` steps read the platform-neutral `broker.*` fields -- `broker.domainCerts`,
`broker.productKeys`, `broker.hostDiagnosticDir`, `broker.cliScriptsDir` -- shared verbatim with
Kubernetes, plus `tls.cert`/`tls.certKey` (server certificate) and `semp.adminPass`
(SEMP login); the `redundancy.*` names drive role detection for `broker perform assert-leader` /
`broker perform redundancy-test`, and `redundancy.backup.addr` is also read at verify time -- it is where those
two commands reach the backup's SEMP service from the primary. The rest of the `redundancy.*`
table and `redundancy.psk` are consumed earlier, by `broker deploy`, which externalizes the
key as a secret -- it does not create it, and no command here writes to an env file.
Container-only knobs live under `docker.*` /
`podman.*` (runtime, compose invocation, container name, data dir, network mode, rootless).
See a bare `solace-util examples`, or the same text at [env/sample.yaml](../env/sample.yaml).

## Rendering without applying

`generate <target>` is the one, dedicated way to review the exact artifact a command would
apply before it touches a cluster or a host. There is no `--dry-run` and no per-command
render flag. It prints to stdout and changes nothing: it runs no external command at all,
which is what makes it the safe way to inspect an env file you did not write (see
[The command fields are executable content](configuration.md#the-command-fields-are-executable-content)):

```
solace-util broker generate -e dev.yaml                            # namespace + Secrets + CR (kubernetes)
solace-util operator generate -e dev.yaml                          # namespace + regcred + bundle (kubernetes)
solace-util broker generate -e dev.yaml --platform docker          # the compose file
solace-util broker generate -e dev.yaml --platform podman          # the quadlet unit
```

`generate` is a command with a named target rather than a flag on `broker deploy`, on
purpose: an artifact you meant to inspect and a cluster you meant to change should not be
one typo apart.

**Use `-o/--out` rather than `>` to keep the artifact.** Shell redirection is not portable:
Windows PowerShell 5.1 re-encodes this tool's plain ASCII output as UTF-16LE with a BOM, and
`kubectl apply` then rejects the file with an error that points at the YAML rather than at
the shell. The re-encoding happens after the command has exited, so nothing inside it can
prevent it -- `-o` sidesteps the shell entirely:

```
solace-util broker generate -e dev.yaml -o solace.yaml
solace-util operator generate -e dev.yaml -o operator.yaml --no-prompt
```

It buys three more things over a redirect: the file appears only if the render SUCCEEDED
(a shell truncates the target before the command even runs, so a failure leaves a
half-written file that looks like output), the secret-bearing Kubernetes stream stays out of
terminal scrollback, and the file is created `0600` instead of inheriting the shell's
default. An existing path is confirmed before it is replaced -- `--no-prompt` answers yes,
and a run with no terminal keeps the file and says which flag would have proceeded.

The same `-o`/`--no-prompt` pair is on `examples` and `convert`: every command whose output
is an artifact you keep answers the overwrite question the same way. There is no `--force`.

**On docker and podman, secrets are never part of a deployment artifact.** Each one lives
in podman's secret store or in a host environment variable the compose file names, and the
quadlet unit and compose file reference it by name only -- so their `broker generate`
output is safe to review, diff and share.

**On Kubernetes the opposite is true, and there is only one command.** A Secret manifest IS
the artifact, so `broker generate` and `operator generate` carry the admin password, the
pre-shared key, the TLS private key and the registry credential in base64. There is no
secret-free variant to reach for -- the `generate secrets` sub-tree was removed precisely
because splitting the stream in two made the ordering the operator's problem. Treat that
output exactly like the env file it came from.

`operator deploy` issues three applies when registry credentials are configured
(`image.user`/`image.pass`, or their `*Env` equivalents) -- the operator namespace, then the
fixed-name `regcred` image-pull secret into it, then the rest of the bundle -- and two
otherwise (namespace, then bundle; there is no `regcred` to apply without credentials to
build it from). Everything it applies is byte-for-byte what `operator generate` prints, and
the ordering is what makes a first install work -- the secret is namespaced, and the
namespace only exists in the bundle.

The broker's own settings -- routername, redundancy, the scaling knobs -- are inlined into
whatever `broker generate` renders: `Environment=` lines in a quadlet unit, an
`environment:` block in a compose file, `spec.systemScaling` in the CR. There is no
separate command for them, because there is no separate artifact.

**Every platform hands the broker its secrets as files**, read through the setting's
`*filepath` variant, and named after the setting they feed -- so the layout inside a
container matches the data keys of the equivalent Kubernetes Secret:

| Secret | In-container path | Host-side name |
| --- | --- | --- |
| `semp.adminPass` | `/mnt/secrets/username_admin_password` | `<container.name>-admin-password` |
| `semp.additionalUsers[].password` | `/mnt/secrets/username_<username>_password` | `<container.name>-user-<username>-password` |
| `redundancy.psk` (HA) | `/mnt/secrets/redundancy_authentication_presharedkey_key` | `<container.name>-redundancy-psk` |
| `tls.certPassphrase` | `/mnt/secrets/tls_servercertificate_passphrase` | `<container.name>-tls-passphrase` |
| `tls.cert` + `tls.certKey` | `/mnt/certs/server/tls.pem` | `<container.name>-tls-servercertificate` (docker); a host FILE on podman, see below |

The host-side name carries `container.name` (default `solace`) so two brokers on one host
never share a podman store entry or a compose variable. On Kubernetes the operator mounts
the credentials Secret itself, so the only data keys that matter are
`username_admin_password` and `username_monitor_password`.

## Removing a broker: what stays, what goes

There is one removal per noun: **`broker remove`** and **`operator remove`**. Each keeps the
layer that is expensive or impossible to get back **by default** -- the broker's persistent
data (Kubernetes PVCs, or the container's data directory), and the operator's CRDs.

`broker remove` also owns what used to be separate `remove secrets` and `remove namespace`
commands: it deletes the broker resource, then this deployment's secrets, then considers the
namespace. `operator remove` never removes the operator's own namespace, and prints the
`kubectl delete namespace` line to finish by hand, so the outcome is stated rather than
inferred. The operator's image-pull secret is removed either way.

Deleting the CRDs is the sharper of the two layers, because they are cluster-wide -- it
cascade-deletes **every** PubSubPlusEventBroker in the cluster, including ones this env file
has never heard of. So `--delete-crd` refuses outright while any PubSubPlusEventBroker still
exists anywhere -- each is named `<namespace>/<name>` in the refusal, and a failure to even
list them refuses too. Only the CRD layer is refused; the operator's controller Deployment is
removed either way.

### The layer flag raises the question

The `--delete-*` flag does not answer the question, it **asks** it. Learning the contract on
one removal teaches the other:

| `--delete-data` / `--delete-crd` | `--no-prompt` | Terminal | Outcome |
| --- | --- | --- | --- |
| no | either | either | Layer **kept**. Nothing is asked about it. |
| yes | no | yes | Prompted; only an exact, case-insensitive `yes` deletes it. |
| yes | yes | either | **Deleted**, unattended. |
| yes | no | no | Layer **kept**, with a loud warning naming `--no-prompt`. The removal itself still proceeds. |

So a fully unattended removal that also drops the data is `--delete-data --no-prompt`: the
two flags answer different questions and deliberately **compose**. There is no global
`--yes` -- one silencer per command beats two flags whose overlap has to be memorised.

The last row is the one worth reading twice. A scripted removal cannot lose data by
omission, and it cannot lose it by *asking* either: without a terminal to answer the
question, the layer survives and the run says so rather than proceeding in silence.

On docker and podman the directory `--delete-data` deletes recursively is exactly
`<platform>.container.dataDir`, which is why that key is **required to be an absolute path**
and is the one host path not resolved against the env file's directory: what a recursive
delete points at should never move because of where the command was run from
([configuration.md](configuration.md#relative-paths-resolve-against-the-env-file-not-the-current-directory)).

Two more properties hold on every path:

- **On `operator remove`, the CRD question is not even asked when a broker still exists.**
  Its answer is already fixed to "keep", so prompting would invite a "yes" this tool will not
  honour. A warning names what was found. An explicit `--delete-crd` still runs into the
  refusal above rather than being downgraded to "kept": naming the flag earns a loud failure
  listing the brokers in the way, not a silent no-op.
- **A PVC delete that fails is reported as a failure.** `--delete-data` deletes every role's
  PVC and keeps going even if one fails, but if any did fail after `--ignore-not-found`
  already absorbed the benign "already gone" case -- an RBAC denial, a stuck finalizer --
  `broker remove` returns an error naming every PVC that survived, instead of reporting the
  data as gone.

### The namespace is only offered when it is empty

The namespace is **never in the delete set**. After everything this env file owns is gone,
what remains is enumerated, and only an otherwise-empty namespace is *offered* for removal --
as its own question, which `--no-prompt` answers yes.

An occupied namespace is **listed and kept, on every path, `--no-prompt` included**. There is
no flag that deletes it: deleting a namespace takes everything in it, including objects
another team put there, and this tool may not have created it in the first place.

- Four Kubernetes system namespaces -- `default`, `kube-system`, `kube-public`,
  `kube-node-lease` -- are refused outright and unconditionally. There is no interpretation
  of "delete kube-system" this tool should ever carry out.
- Occupancy is checked ownership-BLIND, by one classifier (`Cluster.NamespaceContents`): it
  lists a fixed set of kinds (`all`, PersistentVolumeClaims, Secrets, ConfigMaps) and
  discounts only what Kubernetes itself puts in every namespace -- the `kube-root-ca.crt`
  ConfigMap and the default ServiceAccount (without that, no namespace would ever read as
  empty) -- plus a LimitRange, ResourceQuota or Event, which are cluster policy rather than
  occupancy. Anything else still there blocks removal, with no further judgement of who put
  it there: **a second Solace broker sharing the namespace** blocks it exactly like anything
  else would.
- **Retained data keeps the namespace too.** Without `--delete-data` the PVCs are still
  there, so the namespace is not empty and is kept -- which is what makes "delete the
  namespace and take the retained data with it" impossible to reach by accident rather than
  merely discouraged.
- A failure to even list the contents keeps the namespace: being unable to see what is in
  there is not permission to proceed. That direction is the whole safety property -- a
  namespace wrongly reported empty gets cascade-deleted, while one wrongly reported occupied
  merely stays.

**Docker and Podman have their own teardown wrinkles**, on top of the `--delete-data` layer.
The sequence is: stop the container, remove it, remove the compose file or quadlet unit,
remove the engine secrets and the server-certificate bundle, then ask about the data
directory.

- Podman's secret store is a real, separate persistence layer -- `broker remove` removes
  every secret `broker deploy` loaded into it as part of removing the container. A secret
  that is already gone (or fails to remove) only warns. The server-certificate bundle under
  `podman.baseDir` is different: failing to remove that one is **fatal**, because leaving a
  private key on the host is the worst outcome available.
- If `systemctl stop` (Podman) or `stop` (Docker) fails, the removal aborts before touching
  the unit, the container, or the data directory -- it continues only when the broker is
  *confirmed* down: systemd reporting `inactive`, `failed` or `unknown`, or the engine
  listing no running container. "Still running" and "could not be confirmed either way" both
  abort, because the case this guard exists for -- a rootless Podman whose systemd user
  session is unreachable while `podman info` still succeeds -- produces the second one, and
  silence is not confirmation.
- `broker remove` against a host with nothing deployed -- no compose file, no container by
  that name -- is a no-op rather than an error, matching `--ignore-not-found` elsewhere.

### Removing the operator does not always remove it

`operator remove` reads the installed operator's watch list before doing anything, because
the operator is cluster-scoped and may be serving brokers this env file knows nothing about:

- If this env file's namespaces cover **everything** the operator watches, it is removed.
- Otherwise the watch list is **narrowed** -- this env file's namespaces are dropped from it
  and the operator keeps running for the rest. The narrowing edits only `WATCH_NAMESPACE`,
  so a removal can never change the operator's image out from under whoever else is using it.
- An operator watching **all** namespaces (an empty `WATCH_NAMESPACE`) cannot be narrowed to
  exclude one, so it is kept with a warning rather than being silently left watching
  everything minus nothing.

### Every destructive command confirms

`broker remove`, `operator remove`, `operator stop` and `broker restart` all ask before
acting -- `restart` included, because it drops every in-flight connection (it deletes pods on Kubernetes so the
StatefulSet recreates them, and bounces the container on Docker/Podman). So do
`broker configure default-vpn` and `broker configure default-users` when disabling (never
when enabling: bringing something back up needs no gate), since each drops client
connections that depend on it, and `broker configure domain-certs --remove`, since it
deletes certificate authorities already configured on the broker. An interactive terminal is
asked `[y/N]`; a session that does not answer, and was not given `--no-prompt`, refuses
loudly rather than destroying anything unattended.

**What counts as "answered" is decided by reading, not by what kind of stream stdin is.**
The question always goes to stderr and the answer is always read from stdin, so a shell
that hands this tool a pipe rather than a console -- Git Bash, and anything that wraps the
binary -- is asked and answered normally. An input that ends without a reply is the
unattended case, and the refusal names `--no-prompt`. The one case with no good answer is
a scripted run whose stdin is an open pipe that nothing ever writes to and nothing ever
closes: it waits at the question instead of refusing it. **Pass `--no-prompt` in any
unattended run** -- CI, cron, a systemd unit -- and none of this arises.

On Kubernetes, deleting a pod is graceful, not abrupt: it sends SIGTERM and honours
`terminationGracePeriodSeconds`, and only `--force --grace-period=0` would kill outright --
this tool never passes it. `broker stop` is not in the list above: it asks nothing, since
nothing is deleted.

**On Kubernetes, every one of those prompts also says where it will act.** The env file names
a namespace but never a cluster -- which cluster a `kubectl` call reaches is decided entirely
by the kubeconfig's current context, so a `dev.yaml` run against a context that has drifted to
`prod` looks identical right up until the delete lands. That fact is surfaced twice. First,
the startup preamble prints `==> kube-context: <name>` beside the `==> using kubectl: <path>`
line, read from `kubectl config current-context` -- a kubeconfig read, not a cluster round
trip, so it costs nothing and still prints against an unreachable cluster. Second, the prompt
repeats both facts: `broker remove` and `broker restart` read "... in namespace `<ns>`
(context `"<ctx>"`)". The context is quoted because it comes from `kubectl config
current-context` rather than from the env file. The namespace question names its own
namespace as part of the sentence and appends only the context clause. `operator remove` names the **operator's own**
namespace rather than repeating the broker's, since the two can differ and naming the wrong
one would be worse than naming none. The context clause is dropped entirely, never rendered as
an empty `(context )`, on a kubeconfig with no current context or on a non-Kubernetes platform.

## Exporting and importing configuration

`broker perform export-config` and `broker perform import-config <file>` capture a
running broker's configuration as one artifact and apply it back to a broker --
possibly a different one. Both drive the Solace CLI over the same `<runtime> exec`
channel every `broker configure`/`broker perform` command already uses, so neither
needs a broker credential: authorisation is `kubectl`/`docker`/`podman` access to the
pod or container, not `semp.adminPass`. That is also what keeps `import-config` safe
to run even when the artifact itself changes the CLI admin password -- the channel it
runs over never depended on that password to begin with.

### What export captures, and why it is not a backup

Scope is the presence of a flag, not an enum: with neither flag, export captures the
whole broker including every message-VPN; `--vpn NAME` (repeatable) captures only the
named VPNs; `--broker-only` captures the broker level with no VPN at all. The artifact
is the broker's own `show current-config` output -- the only capture that covers
product keys, CLI users, redundancy, interfaces, syslog and spool sizing, none of
which either SEMP API exposes.

**This is not a backup, and treating it as one is the mistake this section exists to
prevent.** The artifact carries no product key, no TLS private key, no Kerberos
keytab and no message data -- a broker rebuilt from it alone does not come up. An
operator who believes an `export-config` capture is a disaster-recovery backup has
stopped planning for the disaster it does not cover: recovering a lost node is still
"deploy it from its own env file and let config-sync populate it"
([Bringing up a fresh cluster](#bringing-up-a-fresh-cluster)), never "import this
artifact into an empty broker". What this tool does offer for disaster recovery is
[Data replication](#data-replication) -- a second broker already carrying the messages, which
is a different thing from a configuration capture and is configured separately.

### Handling the artifact

The artifact carries every configured secret in the broker's encrypted form -- that
is what makes it replayable by `import-config` -- so it is as sensitive as the env
file itself, and `-o`/`--out` writes it `0600`.

**Write it outside a git working tree.** `0600` protects it from other users on the
host and does nothing at all about `git add -A`. This repo ignores `/*.cli` and
`/*.backup` at its own root as a backstop, but that only covers the two names an
export happens to land on here -- `-o` takes any path, and a capture written into a
subdirectory, or into another repo, is one commit away from being published. Keep
artifacts in a directory no repository tracks.

The broker also offers a `redact` form of `show current-config`, and this tool does
not use it. Redaction strips exactly the credential material an import has to put
back, so a redacted capture is an artifact `import-config` would refuse: the flag
could only ever produce a file this tool declines to read. For a copy safe to
circulate, redact a copy of the artifact yourself, and keep the real one `0600`.

The broker runs `show current-config` **in series** with configuration commands, so
while an export is running, other configuration changes on that broker wait. Do not
run `export-config` at the same time as `broker configure`, an `import-config`, or an
operator-driven upgrade against the same broker -- queue them instead of overlapping.

### The import lifecycle

`import-config <file>` runs three phases against the target: **plan** (parse the
artifact, read the target's broker type and message-VPNs, and classify every
broker-level section against the fixed table below), **confirm** (see below), then
**apply and verify**.

**The destructive part, stated plainly:** a message-VPN in the artifact that already
exists on the target is torn down first, which destroys every message spooled in
every one of its queues and removes any object the artifact does not contain -- ACL
profiles, client profiles, usernames, bridges, all of it. On an HA pair, config-sync
propagates those deletions to the mate. A VPN the target does not yet have is created
fresh instead and loses nothing.

This is why VPN scope asks once for the whole run, and needs an exact `yes` rather
than the ordinary `[y/N]` -- the same bar `--delete-data` sets on `broker remove`
([The layer flag raises the question](#the-layer-flag-raises-the-question)).
`--no-prompt` answers it, so `import-config <file> --no-prompt` is what a fully
unattended overwrite looks like. Creating VPNs the target does not have destroys
nothing and only asks `[y/N]`; broker scope never asks anything at all, because which
broker-level sections apply is a fixed classification rather than a per-run decision
(see below).

**The apply runs as chunks, and the broker stops itself at a bad line.** The whole
apply is one generated shell script, uploaded once and run once, that applies each
chunk through the broker's own `source script <name> stop-on-error no-prompt` and
checks the last ten lines of that chunk's transcript for a rejection. That matters
because `cli -Apes` exits 0 even when the broker refuses a line: without
`stop-on-error` one bad line let the rest of the script run on top of it, and
without the transcript check a failed **teardown** was invisible -- the verification
diff cannot see one, since a VPN that was never removed still satisfies every line
the artifact asks for.

Chunks run in this order, and a failure stops the run there:

1. every existing VPN's teardown;
2. removing any target-only objects the artifact does not recreate (today, stale
   virtual hostnames), when there are any;
3. creating every message-VPN the artifact defines, existing or new -- this runs
   before any broker-level section because a broker-level line naming a
   message-VPN is rejected if that VPN does not exist yet;
4. `Create logging`, alone, because it ends the CLI session;
5. `Create Usernames`, alone, because it rewrites the CLI admin password;
6. the remaining broker-level sections;
7. one chunk per message-VPN, **`default` first** -- it is the VPN that is edited
   rather than recreated, and it carries the port changes most likely to be refused.

**A failed apply stops there and the verification diff does not run.** The report
names the chunk that failed and quotes the broker on it, and that is the whole
answer -- diffing at that point would list every chunk after the failure as
"missing", which is true and useless, burying the one line that matters under work
that was never attempted. The diff runs only after a fully successful apply, and
then reports whatever did not land despite the broker accepting it.

The report names the chunk that failed and quotes the broker's last words on it.
Nothing else from the apply comes back: a transcript of applying a configuration
repeats every credential in it, so a successful chunk's output never leaves the
broker, and the script deletes every file it wrote even if it is interrupted.

Because a failure stops the run, **re-running the same import is the recovery** --
the chunks that already applied are applied again, which is safe, and the one that
failed is retried. Nothing resumes from a high-water mark.

**Why teardown rather than reconcile:** a Solace CLI configuration block cannot be
applied a second time, so there is no in-place update to fall back to. The teardown
script is not written by this tool -- it is captured FROM THE TARGET with
`show current-config message-vpn <name> remove`, so the broker orders the removal of
its own queues, ACL profiles, client profiles, usernames and bridges in whatever
sequence its own dependencies require.

That also settles the `default` message-VPN, which cannot be deleted at all -- and
neither can the `default` client-profile, acl-profile or client-username inside any
VPN, which is why `broker configure default-vpn` shuts the default VPN down rather
than removing it. The broker's `remove` output simply omits the lines it would
refuse: importing over `default` empties it and leaves the VPN itself in place, then
applies the artifact onto it. There is no reserved-name filter and no `default`
branch in the import path, because a tool-written teardown would have needed a list
of undeletable objects kept correct across broker versions, and asking the broker
avoids owning that list at all.

### Two detectors, in order

`cli -Apes` exits 0 even when the broker rejects a configuration line, so the exit
code tells you nothing and the transcript is the only evidence there is. Two things
read it, and they run in sequence rather than both always.

The first is the apply itself. Every chunk goes through the broker's own
`source script ... stop-on-error no-prompt`, so the broker STOPS at the first rejected
line, and the driver then checks that chunk's transcript tail for a rejection keyword.
Because the broker stopped, the rejection is the last thing it printed -- which is what
makes a ten-line tail scan sound rather than lucky. A rejection there names the chunk
and reports the broker's own words.

The second is verification. After a fully successful apply -- and only then --
`import-config` re-exports the target and diffs it block by block against the artifact.
A non-zero exit from that phase means the report found a block that differs or is
missing. Read the printed report, which names them, rather than the exit code alone.

A failed apply reports and returns without diffing. Diffing after a failure would list
every chunk the run never reached as missing: true, and useless, since it buries the one
line naming the actual failure.

**Recovery is re-running the same command.** A failed or interrupted import leaves
whatever it had already applied in place, half-built VPN included; running
`import-config` again finds that VPN already exists on the target, tears it down, and
rebuilds it from the artifact. There is no resume and no partial-apply flag -- the
whole three-phase run repeats.

### Broker scope is a fixed classification

Which broker-level sections `import-config` applies, skips, or applies with lines
dropped is baked into this build -- there is no config key, no flag and no per-run
decision, because an artifact that chose what to execute would be exactly the
untrusted-input hazard `internal/config/execguard.go` exists to prevent elsewhere.
[docs/import.md](import.md) is the section-by-section table, generated from that
classification so the published docs cannot drift from what the code does -- read it
rather than this paragraph for the full list.

**`import-config` accepts only what `export-config` produced.** An artifact carries a
`! solace-util-export:` marker, and a file without one -- a raw `show current-config`
capture, a hand-written script, anything else ending in `.cli` -- is refused, naming
the command that produces an importable one. That is not fussiness: a raw capture
still contains the sections import must never apply and the CLI transcript
(`xps-ps-01> home`) that export cuts off, so importing one would sever the SEMP
channel mid-run and feed echoed prompts to the CLI as commands, both invisibly
because `cli -Apes` exits 0. An artifact from a newer export format is refused the
same way rather than guessed at.

This is a **provenance** check, not a tamper check, and the distinction matters. The
marker is a comment; anyone editing an artifact can keep it and nothing will notice.
It stops the wrong file being imported, not a deliberate edit -- and it could not do
more while `broker perform cli-script` runs any script you hand it. That command is
the acknowledged way to run unchecked CLI, and its help says so.

A section marked **skip** is removed by `export-config`, not merely ignored by
`import-config`: the artifact never contains it, and records the omission as a
`! solace-util-omitted:` line naming the section and the reason, so what is absent
is visible rather than something to notice. Lines an apply-filtered section drops
(`routing interface "intf0"`, the spool sizing) are removed the same way. Import
still ignores such a section if an older or hand-edited artifact carries one, and
its plan report mentions skipping only then -- after a current export, it has
nothing to mention.

Several of the applied sections interrupt something while they run, and broker scope
does not prompt before doing any of them: `Configure SMF Service` bounces all
messaging, `Configure Service` bounces the msg-backbone data path, `Create logging`
ends the CLI session it runs in (so it is always applied alone, in its own
invocation), `Configure SSL` can drop live TLS sessions, and
`Configure HealthCheck Service` bounces the endpoint the kubelet readiness probe
depends on.

### Kubernetes-specific hazards

Three sections collide with fields the EventBroker operator's CR already manages, and
each is worth knowing before running `import-config` against a Kubernetes broker:

- **`Create Usernames` vs `adminCredentialsSecret`.** This section IS applied (an
  explicit override of the classification's own recommendation), and it overwrites
  the target's CLI admin password with the artifact's. The CLI channel itself does
  not care -- it needs no broker credential -- but `adminCredentialsSecret` still
  holds the old value, so every pod fails its readiness probe and the operator has no
  way to repair that on its own. Update `semp.adminPass` to match and redeploy.
- **`Create Redundancy PSK` vs `preSharedAuthKeySecret`.** This section is never
  applied, precisely because of this collision: the operator ignores
  `preSharedAuthKeySecret` updates once an HA group already exists, so a CLI-pushed
  PSK would permanently disagree with what the Secret says, with no automatic repair.
- **`Configure System` vs `systemScaling`.** Also never applied: the CR sets
  connection/subscription scaling at every pod boot, so a value pushed over the CLI
  would only survive until the next pod recreate anyway.

A cross-type artifact -- a software-broker capture applied to an appliance, or the
reverse -- is refused outright rather than partly applied. The check reads the SEMP
schema version each side reports; the two broker families do not share a
configuration surface (interfaces, VRF, DNS, clock and SNMP exist on one and not the
other), so there is no partial overlap that is safe to apply.

`Create Domain Certificate Authority` **is** applied, and applying it deletes every
domain CA already on the target before loading the artifact's list, with no conflict
check. That overlaps `broker configure domain-certs`
([Post-deployment configuration order](#post-deployment-configuration-order)), which
already owns this surface and is the narrower tool: it manages exactly the CAs the env
file lists, in place, without touching anything else an `import-config` run would also
rewrite. Prefer `domain-certs` for ordinary certificate rotation; reach for
`import-config` only when the intent is to replace the target's whole configuration.

## Data replication

**Replication is a DR pair; redundancy is one HA group.** Two separate brokers, each
message-VPN active at one site and standby at the other. A replicated deployment usually has
both -- an HA group at each site -- and nothing here changes how either group runs locally.
The intended steady state is declared in the env file's `replication:` block
([Replication](configuration.md#replication)), the same text in both sites' files, and two
commands act on it:

```
solace-util broker configure data-replication    # converge THIS broker to the block
solace-util broker perform   data-replication    # move roles across BOTH brokers
```

Both are registered as `dr`, run on every platform, and take `--no-prompt` (which answers
the confirmation and nothing else) plus `--pod` on Kubernetes.

**Both write to ONE node of the local HA group -- the primary unless `--pod` names another
on Kubernetes -- and config-sync carries the change to the rest.** Every replication setting
is `HA: yes` in the broker's own config-sync table, so there is no per-node loop to run and
no second host to visit. Config-sync is
**assumed operational** and is never checked: it is off by default and can be oper-down, but
an HA group in service has it running, and a stale backup after a run that reported success
is the first place to look.

**Which site a broker is comes from the broker, not the file.** The block is byte-identical
at both ends, so nothing in it says "this one". The tool asks the broker for its own router
name and matches it against `sites[].routerNames`. No match names the router name it read
beside every name the file declares; a name matching both sites is a config error naming
both. There is no "assume the first" -- a wrong guess here points a switchover at the wrong
broker.

### Configuring a site

`broker configure data-replication` converges the broker it is run against, and **never
contacts the mate**. It therefore works on every platform, needs no `via:` block anywhere,
and cannot be blocked by a WAN outage. Run it at both sites with the same file.

**The apply is TWO PHASES, in two separate broker calls**, because mate configuration (the
address lines and the virtual-router-name) can only be changed while every VPN on the broker
has replication disabled. That is a real broker rule, confirmed by the operator; it is not
stated in the CLI reference, which does spell such preconditions out for other services, so
it is worth knowing that the documentation will not tell you.

**This is why a mate change is disruptive and a role change is not.** Changing which addresses
this broker dials its mate on requires stopping replication everywhere on it first, so there
is no version of that operation that avoids the interruption. Changing a VPN's role requires
nothing of the sort. The command keeps the two apart, and only does the disruptive half when
the mate configuration genuinely differs from the env file:

1. **Mate convergence -- only when the mate actually differs** from what this broker already
   holds (compared by address set and virtual-router-name, not by whether a removal happens to
   be non-empty, so a broker missing one of several wanted addresses is still recognised as
   differing). When it differs: every VPN this broker currently reports admin-enabled for
   replication is shut down first, then the `no` forms for any mate address or
   virtual-router-name the broker holds and the file does not carry, then the new address
   lines rendered for THIS broker's type from the OTHER site's `endpoints` and
   `virtualRouterName`. Rendering happens, and is checked for a broker-type mismatch, BEFORE
   this or the per-VPN step ever writes anything. When the mate already matches, this phase
   writes NOTHING and stops no VPN -- the report says so.
2. **Per message-VPN, always**: for each one in `replication.vpns`, in file order,
   `state active|standby` -- the role derived from `activeAt`, set IN PLACE against a VPN
   that is up and replicating, with no shutdown around it. A VPN is only enabled
   (`no shutdown`) when it actually needs it: either replication was off for it when the
   run started, or phase 1 just stopped it to change the mate. The role is written before
   the enable in that case, so a VPN never comes up holding the wrong one, even briefly.
   Then `shutdown` on any VPN the file does not list that this broker currently has
   replication enabled for -- skipped when phase 1 already stopped it.

   This is why phase 1 is the only thing that interrupts replication. An earlier version
   cycled every listed VPN down and back up to change its role, which stopped replication on
   each of them on every run, including runs where the mate had not changed and nothing
   needed to stop. **If the mate configuration already matches your env file, this command
   now interrupts nothing.**

Phase 1's shutdown is broader than the listed VPNs: it stops replication on **every** VPN this
broker currently has replication enabled for, including ones the file does not mention,
whenever the mate needs to change. Enabled is a wider set than replicating -- it also catches a
VPN enabled without a resolved role, which is exactly the VPN that would otherwise refuse the
mate lines this phase exists to make writable. Phase 2 then turns the listed ones back on. The command takes the
**exact-`yes`** gate rather than the ordinary `[y/N]`, the same bar `--delete-data` sets on
`broker remove`, and states this conditionally before asking (it cannot yet know whether phase
1 will run) alongside the role it will give each listed VPN; `--no-prompt` answers it once for
the whole run.

**A rejection is never retroactive.** Each phase runs inside the broker's own
`source script ... stop-on-error`, so a rejected line stops THAT phase; an EARLIER phase that
already succeeded is not rolled back. A phase-1 rejection leaves replication disabled on every VPN
it reached, and phase 2 is never sent. A phase-2 rejection after a successful phase 1 leaves
the mate already converged, replication back on for some prefix of the listed VPNs, and the
rest -- including every unlisted one -- still disabled from phase 1.

Throughout, what stops is REPLICATION, never the message-VPN. Every line either phase sends
is inside the VPN's own `replication` node, so the VPNs keep running and their clients stay
connected; what a failure interrupts is the feed to the mate. The error names which
phase stopped and states plainly what that leaves running or stopped, and points at
`show replication` and `show message-vpn * replication` to read the rest back. **Re-running is
SAFE** -- the whole script is recomputed from a fresh read of the broker rather than replayed,
which also re-picks the queue-guard keyword from what the broker now reports -- but it is
**sufficient only when the cause was transient**: a line the broker refuses for a reason in the environment is
refused again at the same place. A re-run after a successful phase 1 is cheaper than the
first attempt, because the mate now matches and that phase is skipped.

The exit code reflects what the broker said: `cli -Apes` exits 0 even on a rejected line, so
each phase's transcript is scanned for the same rejection keywords `import-config` uses -- one
list, not a second one to keep correct. After a clean apply the command prints what changed
(including every VPN phase 1 stopped, when it ran) and then the broker's own `show replication`
and `show message-vpn * replication`, so the result is read back from the broker rather than
asserted by the tool.

**This is the one path in the feature that can produce two actives.** It is local-only by
design, so it cannot see what the mate is doing: run it at the new-active site while the old
site still holds that VPN active and both are active until somebody notices. That is the
accepted trade for a command that works with the mate unreachable. The safe way to MOVE a
role is `broker perform data-replication`; use `configure` to stand a pair up and to change
everything that is not a live role handover.

### Switching roles

`broker perform data-replication` changes **nothing but per-VPN replication state**. Mate
addresses and enablement are `configure`'s; this command only verifies them. A failover is
therefore an edit and a run:

```
# edit replication.vpns[].activeAt in the env file (both sites' copies), then:
solace-util broker perform data-replication -e env/prod.yaml
```

Everything that can refuse does so before the first write:

- **Both channels are preflighted**, proving each site is reachable and mutating nothing. A
  missing, malformed or unreachable `via:` block stops the run here. A site with no `via:` at
  all is refused by name -- a switch has to change the role at both ends, so the mate must be
  reachable.
- **Both sites must be on their primary HA node.** A site running on its backup has already
  had something go wrong, and stacking a DR role change on a local failover makes both harder
  to undo. The refusal names the site and points at `broker perform assert-leader`. A
  standalone broker has no HA group and always passes.
- **The brokers and the file must agree about which pair this is.** Each broker's own
  `show replication` must report the OTHER site's declared `virtualRouterName` as its mate. A
  mismatch means the file names a site the broker never heard of, or `configure` has not been
  run since the block changed; the refusal says which side disagrees and with what.
- **Every listed VPN must exist and have replication enabled at both sites.** The check is
  scoped to the list -- a VPN the file does not manage is ignored, not reported. Missing or
  shut-down VPNs are named, and the refusal points at `configure`, which is what creates and
  enables them.
- **A VPN already active at BOTH sites stops the run.** This tool did not create that state
  and will not silently repair it: demoting one side picks a winner, and which site keeps its
  spooled messages is not a decision to make for you. Demote the side you do not want, then
  run again.

**A VPN already correct at both sites is skipped entirely** -- no demote, no promote, no
interruption -- so a single-VPN failover never touches the VPNs that are not moving. When
nothing is left to move the run says so, prints both sites' status and exits **without
asking anything**: a no-op must not demand an exact-`yes`, or the question stops being read.
Otherwise it names every move and takes the exact-`yes` gate.

**The safety ordering is demote everywhere, confirm, re-read, then promote.** The demote
lands on the site `activeAt` does NOT name, so a VPN is never demoted at the site it is about
to be promoted at; it is issued whether or not that site currently reads active, which makes
the step idempotent. Every demotion is then confirmed by reading it back -- a successful call
is not evidence the broker applied it -- and the state is read **again** immediately before
promoting, because a concurrent `configure` run or an operator at a CLI can move a role in
between and by then the confirmation is stale. A VPN that has gone active again aborts the
promotion phase rather than being promoted over.

Stated precisely, because an over-strong claim is worse than none: **within one run, no
promotion begins until every demotion has been confirmed and re-read.** It is a guarantee
about the run, not about the system -- nothing locks the brokers, and a `configure` run at
either site can still promote something. What catches an arrival that lands after the last
read is the verification below.

**Verification decides the exit code.** Both sites are re-read afterwards and printed
(`show replication` and `show message-vpn * replication` at each), and every VPN in the list
is checked -- the ones that moved and the ones that were skipped, since a skipped VPN is the
run's own premise and a change there means something else is writing. Any VPN not where the
file says it should be is a **failure**, not a warning: the whole point of the command is
that the end state is known.

### Reaching the mate

`perform` needs to write the role at the far site, so the mate's entry must carry a `via:`
block; `configure` needs none. The key that is present is the mechanism
([Replication](configuration.md#replication)):

| `via:` | What it does |
| --- | --- |
| `kubernetes` | Runs that site's own cluster CLI (`kubectl --context dr ...`) and drives the broker CLI in its primary pod, exactly as this tool drives a local one. The command goes through the same allowlist as `kubernetes.command` |
| `semp` | Posts SEMP v1 requests, with `curl` exec'd inside THIS broker's own container -- so it needs no second kubeconfig and no new binary on your machine. The password is the MATE's admin password, never this deployment's |

The SEMP leg has one advantage worth knowing when a run is being debugged: a SEMP reply
carries an explicit `ok`/`fail` verdict, while the CLI leg has to scan a transcript because
`cli -Apes` always exits 0.

### When a run fails partway

**Failing between the demote and the promote leaves those VPNs standby at BOTH sites --
unavailable, not corrupted.** That is the intended direction to fail, and the report says so
plainly: an operator who reads "unavailable" must not go looking for data loss.

**Recovery is re-running the same command** once the cause is fixed. A both-standby VPN is
not a special case: it fails the already-correct test, moves through a demote that is a
no-op, and is promoted. `--no-prompt` silences the confirmation and none of the checks.

## Upgrading a running broker

Changing the image tag (or any other setting) is the same edit on every platform --
bump `image.tag` in the env file -- but applying it differs:

**Kubernetes, `updateStrategy: automatedRolling` (the default)**

```
solace-util broker deploy -e dev.yaml
```

`broker deploy` re-applies the custom resource; the operator sees the new tag and rolls the
pods itself (monitor, then backup, then the active node).

**Kubernetes, `updateStrategy: manualPodRestart`**

```
solace-util broker deploy -e dev.yaml     # updates the statefulset template; no pod is touched
solace-util broker restart -e dev.yaml    # bounces monitor -> backup -> primary, waiting for each
```

The operator deliberately waits for you here, so `broker deploy` alone changes nothing
visible. `broker restart --pod <role>` bounces one pod if you would rather drive the order
yourself -- `--pod` is the one way to name a Kubernetes pod (defaulting to the primary when
omitted), here and on `cli`, `shell`,
`broker logs`, `broker perform semp-login-check` and `broker status` too; worth doing after a failover,
since the order above is by configured role and the active node may not be the configured
primary. Check with `solace-util broker perform redundancy-test` first.

**Docker / Podman** (on each host, with its own role)

```
solace-util broker deploy --pod primary -e prod.yaml --restart    # prod.yaml: a podman env file
solace-util broker deploy -e prod.yaml --restart             # prod.yaml: a docker env file
```

`broker deploy` compares the rendered artifact with the one on disk: unchanged is a no-op,
changed is written and then applied to the running broker -- with `--restart`, or
after being asked. Without consent the new artifact is left in place and the command
says the broker is still on the previous one. In an HA group, upgrade the monitor and
backup before the primary.

## Troubleshooting

### File descriptors on rootless podman

Both container artifacts ask the engine for `<docker|podman>.container.ulimits.nofile`
(default `2448:1048576`). A **rootless** container cannot raise `nofile` above the hard limit
of the user invoking podman -- the kernel refuses -- so `broker deploy` checks it on a
podman env file and stops with the exact drop-in to add when it is too low:

```
solace-util broker deploy -e env/prod.yaml
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

`broker deploy` stores the broker's secrets with `podman secret rm --ignore` followed by
`podman secret create` (value on stdin) and mounts them into the container (`type=mount`),
none of which the oldest podman builds support. The symptom is an unknown-flag or unknown-
directive error at deploy time, because `validate` proves only that the runtime answers
`version`. See [Version floors](#version-floors) for the floor and why `--replace` is
deliberately not used.

### Docker compose secrets need compose 2.23.1+

The generated compose file sources each secret from a host environment variable, so an older
compose does not understand the `environment:` secret source and `broker deploy` fails. See
[Version floors](#version-floors) for the version and the v1 fallback.

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

### Import reported failure

`import-config` always re-exports the target and diffs it against the artifact after
applying, and that diff -- never the applied script's own output -- is what decides
the exit code (see
[Two detectors, in order](#two-detectors-in-order)).
A non-zero exit means the printed report named at least one block that differs from
the artifact or is missing from the target entirely; it does not mean a particular CLI
line was rejected, since `cli -Apes` exits 0 either way. Read the report for which
section and which VPN disagree, fix the cause if one is obvious (a value the target
genuinely cannot accept, a section this build classifies differently than expected),
and re-run `import-config` with the same file: a half-built VPN from the failed
attempt is found, torn down, and rebuilt from scratch, so re-running is the whole
recovery procedure -- there is no separate resume step.

### A replication switch refuses before changing anything

`broker perform data-replication` runs a series of checks before its first write, and each
refusal names what to do. **Not on the primary HA node** means that site is running on its
backup: revert activity first (`broker perform assert-leader`, or `redundancy revert-activity`
on the backup). **The brokers and the env file disagree about which pair this is** means a
broker's `show replication` does not name the other site's `virtualRouterName` as its mate:
run `broker configure data-replication` at each site, which is what writes it. **Listed VPNs
do not exist or have replication shut down** is the same answer -- `configure` is what creates
and enables them. **A VPN active at both sites** is refused deliberately and is not repaired
automatically; demote the side you do not want, then run again. Nothing has been written in
any of these cases. See [Data replication](#data-replication).

### A wrapper runtime is refused

`kubernetes.command`, `docker.command`, `podman.command` and `docker.compose` accept only an
allowlisted bare binary name. A wrapper such as `microk8s kubectl` or `lima nerdctl` needs
`--allow-command <name>` on the invocation -- see
[The command fields are executable content](configuration.md#the-command-fields-are-executable-content).

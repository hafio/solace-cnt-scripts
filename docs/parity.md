# Platform parity: solmq-conn-util and solace-util

> [!WARNING]
> **Not a supported Solace product.** `solace-util` was created by Solace
> Professional Services and is supported only by Solace Professional Services --
> not by Solace Support. For help with this tool, contact your Solace
> Professional Services representative rather than opening a Solace Support
> case. This notice covers this tool only, not the Solace PubSub+ Event Broker
> or the EventBroker Operator that it deploys and operates.

`solmq-conn-util` deploys the Solace PubSub+ Connector for IBM MQ. `solace-util` deploys the
Solace PubSub+ Event Broker. Both drive Docker, Podman and Kubernetes with the same external
commands. This page is the contract the two tools share, the places where they differ, and
the decisions taken for those differences. The same text is kept in both repositories.

Each tool documents its own platform behaviour elsewhere. For `solmq-conn-util` read
`docs/DESIGN.md` sections 4 to 7 and `docs/operations.md`. For `solace-util` read
`docs/operations.md` and `docs/container-security.md`. File names below are relative to each
repository root.

## Contract

Both tools follow every rule in this section today.

1. **Kubernetes manifests travel on stdin.** `kubectl apply -f -` creates and updates,
   `kubectl delete -f - --ignore-not-found` removes. No manifest is written to disk.
   `solmq-conn-util`: `internal/runner/runner.go`. `solace-util`: `internal/k8s/cluster.go`.

2. **Every kind a command writes is probed before its first write.** The probe is
   `kubectl auth can-i <verb> <resource>`, with `-n <namespace>` for a namespaced resource and
   `--all-namespaces` for a cluster-scoped one. Reads are never probed: not status, logs or
   exec, and not the read-back after a write. The answer's last line is the verdict; `no`,
   `no - ...` and `no (...)` refuse, and anything but `yes` refuses. A refusal stops the run
   and names the permission, where it applies and the binding to ask for; no answer at all
   is reported apart from a refusal, with a hint to log in. Both tools use the shared
   `kube/rbac`. `solmq-conn-util`: `internal/runner/runner.go` (`Preflight`). `solace-util`:
   `internal/k8s/preflight.go`.

3. **The engine is checked first.** `<engine> info` runs before any other docker or podman
   command. A failure reports a platform hint, such as a stopped daemon or a missing rootless
   socket, in one message both tools take from the shared `containercli`. `solmq-conn-util`:
   `internal/runner/runner.go` (`Preflight`). `solace-util`: `internal/container/preflight.go`.

4. **Podman units are driven through systemd.** After the unit file is written the tool runs
   `systemctl [--user] daemon-reload`, then `start`, `stop` or `restart <service>`. The
   restart count is read with `systemctl [--user] show <unit> -p NRestarts --value`. Both
   tools build these commands with the shared `podman`, `systemd` and `rootless` packages.
   `solmq-conn-util`: `internal/runner/runner.go` (`PodmanDeploy`). `solace-util`:
   `internal/container/manager.go`, `internal/container/inspect.go`.

5. **Podman secrets arrive on stdin.** `podman secret create <name> -` reads the value from
   stdin. A value never appears on argv. On podman 4.7 and later a secret is replaced with
   `secret create --replace`; below that it is removed with `podman secret rm --ignore`
   first. On 4.7 and later the stored value is read back with `secret inspect --showsecret`
   and compared in memory, so an unchanged secret is not replaced. Stale secrets go with
   `podman secret rm --ignore`. `solmq-conn-util`:
   `internal/runner/runner.go` (`PodmanSecretCreate`). `solace-util`:
   `internal/container/manager.go` (`createPodmanSecrets`).

6. **Exec, logs and upload share one argv shape.** Kubernetes uses
   `kubectl exec [-i] [-t] <pod> -n <namespace> -- <command>`. Docker and podman put their
   own flags first and use no `--`. A file is uploaded with `exec -i ... sh -c 'cat > <path>'`,
   the path single-quoted, and the body on stdin. `kubectl logs` flags pass through;
   `--previous` is refused on docker and podman. `solmq-conn-util`:
   `internal/runner/runner.go`, `cmd/solmq-conn-util/shell.go`, `cmd/solmq-conn-util/logs.go`.
   `solace-util`: `internal/k8s/transport.go`, `internal/k8s/ops.go`.

7. **Inspect output is decoded the same way.** Both tools ask `container inspect` and decode
   the answer with the shared `inspect` package. The leading slash is stripped from `Name`.
   Health comes from `State.Health` or, on older podman, `State.Healthcheck`. On podman the
   restart count comes from systemd, not from the engine. Engine text is printed without its
   control characters. `solmq-conn-util`: `internal/statusreport/parse.go`. `solace-util`:
   `internal/container/inspect.go`.

8. **Containers are listed, then filtered in Go.** `<engine> ps --all --no-trunc --format` with
   a tab-separated name, image and status template lists everything. Matching by exact name or
   by image substring happens in the tool, and a listed name that is not a valid container name
   is reported, never passed back to the engine. `solmq-conn-util`:
   `internal/runner/runner.go`. `solace-util`: `internal/container/manager.go`.

9. **An emptied namespace is offered for deletion, never assumed empty.** After a successful
   remove the tool lists `all,persistentvolumeclaims,secrets,configmaps` in the namespace as
   JSON. These do not count: an object being deleted; the deployment's own objects, by kind
   and name, and everything they own through ownerReferences; events, limit ranges and
   resource quotas; `kube-root-ca.crt`, the `default` service account and its token secret.
   Anything else keeps the namespace and is named. `default`, `kube-system`, `kube-public` and
   `kube-node-lease` are never deleted. `delete namespaces` is probed before the question,
   `--no-prompt` approves, and without a terminal the namespace is kept. A listing that fails
   or cannot be read, a refusal, a decline and a failed delete each keep the namespace with a
   warning, and none of them fails the remove. The namespace is deleted by name. Both tools
   use the shared `kube/nsgate`. `solmq-conn-util`: `cmd/solmq-conn-util/namespace.go`.
   `solace-util`: `internal/k8s/namespace.go`, `internal/cli/ops_k8s.go`.

10. **Compose secrets come from the environment.** Each `secrets:` entry has an
    `environment:` source and an absolute `target:`. Values live only in the compose child's
    environment. `solmq-conn-util`: `internal/dockergen`. `solace-util`:
    `internal/render/render.go` (`Compose`).

11. **Hardening is written into every artifact.** Compose files carry `privileged: false`,
    `cap_drop: [ALL]`, `security_opt: no-new-privileges=true` and an explicit `user:`.
    Quadlet units carry `DropCapability=all`, `NoNewPrivileges=true`, `User=` and `Group=`.
    A compose value taken from the operator or from a name is quoted, and every quadlet
    `Environment=` is quoted whole. Both tools write these lines with the shared
    `containerart` packages and check every rendered artifact with `containerart/artifacttest`.
    `solmq-conn-util`: `internal/dockergen`, `internal/podmangen`. `solace-util`:
    `internal/render/render.go`.

12. **Exit codes are 0, 1 and 2.** Success, runtime failure, usage error. A shell or CLI
    session inside the instance returns the child's own code. `solmq-conn-util`:
    `cmd/solmq-conn-util/shell.go`. `solace-util`: `internal/cli/exit.go`.

13. **Watch loops redraw in place and Ctrl-C exits 0.** Both tools run the shared `watch`
    loop. On a terminal each refresh clears the screen and prints the header
    `Every <interval>: <title>   <time>   Ctrl-C to stop`; anywhere else (a pipe, a file, a
    console without escape processing) each refresh is appended under a timestamped section
    line and no escape byte is written. A failed refresh is reported and retried, the loop
    pauses for the interval after each refresh, and the interval is 1s to 1h.
    `solmq-conn-util`: `cmd/solmq-conn-util/status.go`. `solace-util`:
    `internal/cli/watch.go`.

14. **Command binaries are bare names from an allowlist.** `command:` must name `kubectl`,
    `oc`, `docker` or `podman` without a path. The name is resolved through `PATH`, and a
    match in the current directory is refused. `--allow-command <name>` approves another
    name for one run. A privilege-escalation wrapper such as `sudo` is never accepted in a
    command, not even with `--allow-command`; run the tool itself under `sudo` instead. Both
    tools check commands with the shared `execguard` package. `solmq-conn-util`:
    `internal/validate`, `internal/runner/runner.go`. `solace-util`:
    `internal/config/execguard.go`, `internal/engine/runner.go`.

15. **A namespace is created only when absent.** Deploy reads the namespace by name, and an
    existing one needs nothing more. An absent one is created on its own, after a
    `create namespaces` probe; a refusal stops the deploy and names who can create it. When
    the read itself is refused, the probe decides: allowed, the namespace is created; refused,
    the tool warns and carries on, and the next write's own probe stops a deploy into a
    missing namespace. `solace-util`'s `operator deploy` stops there instead, since its
    cluster-scoped documents land first. `solmq-conn-util`: `cmd/solmq-conn-util/main.go`.
    `solace-util`: `internal/k8s/prep.go`.

16. **A rollout is waited on, and a refused wait is a warning.** The wait is
    `kubectl rollout status <kind>/<name> -n <namespace> --timeout=300s`. `solmq-conn-util`
    waits after deploy, start and restart; `solace-util` after start and restart, and after
    scaling the operator up. A wait refused with Forbidden warns and the command succeeds,
    since the change is already applied; a rollout that fails or times out is an error.
    `solace-util`'s deploy hands the StatefulSets to the operator and reads the custom resource
    back instead, and a refused read-back warns. `solmq-conn-util`:
    `cmd/solmq-conn-util/main.go`, `cmd/solmq-conn-util/lifecycle.go`. `solace-util`:
    `internal/k8s/ops.go`, `internal/k8s/operator.go`.

## Divergences

| Topic | solmq-conn-util | solace-util | Status |
| --- | --- | --- | --- |
| Compose file | handed to compose on stdin, never written | the same, with `--project-directory` set to the env file's directory | aligned |
| Confirmation without a terminal | fails at once with exit 1 and names `--no-prompt` | fails at once with exit 1 and names `--no-prompt` | aligned |
| Rootless podman | rootless unless `podman.rootful: true`, which needs root; root without it is refused; `XDG_RUNTIME_DIR` and the DBus address are repaired | the same | aligned |
| Quadlet directory | `$XDG_CONFIG_HOME/containers/systemd`, else `$HOME/.config/containers/systemd`; a relative `XDG_CONFIG_HOME` is refused | the same, and `podman.quadletDir` overrides it | aligned |
| Host checks for rootless podman | linger, which deploy enables for its own account; session repair | the same, plus the limits under Inherent differences | aligned |
| Namespace on deploy | read by name and created only when absent | the same | aligned |
| Waiting after a change | `rollout status --timeout=300s` after deploy, start and restart; a refused wait warns | the same after start and restart; deploy reads the custom resource back, since the operator owns the StatefulSets | aligned |
| Permission probes | every kind before the first write; no read is probed | the same | aligned |
| `can-i` answer parsing | last line; `no`, `no - ...` and `no (...)`; anything but `yes` refuses; unanswered is distinct from refused | the same | aligned |
| Podman secret replace | `secret create --replace` on podman 4.7 and later, remove then create below; on 4.7 and later a secret whose stored value is unchanged is skipped | the same | aligned |
| Redeploy over a running service | left to the platform; a change to a running service asks before anything is written, and `--no-prompt` approves | the same; `--restart` also approves, and recreates an unchanged service | aligned |
| subuid range too small | asked of `podman info` and refused, with a `usermod` fix naming both ranges, in deploy and validate | the same | aligned |
| Namespace gate details | discounts its own objects and what they own, terminating objects and the default token secret; probes `delete namespaces` before asking; never fails the remove | the same, with the custom resource as the root | aligned |
| Output | buffered; one `ok:` or `error:` line at the end | streamed as it happens | open |

## Decisions

**Compose file.** The rendered compose file is handed to compose on stdin with `-f -`.
`--project-directory` is the directory that holds the env file. Nothing is written to disk.

**Confirmation without a terminal.** A destructive command whose stdin is not a terminal
fails at once with exit 1. The message names `--no-prompt`. No answer is read from a pipe.
Under a mintty-style shell stdin is a pipe even at a terminal, so pass `--no-prompt` there.

**Rootless podman.** Rootless is the default: a non-root invoker deploys rootless as itself.
Rootful podman is deployed only when `podman.rootful: true` asks for it, and only by root; a
non-root invoker asking for it is refused, naming the uid. Root without it is refused too,
since rootful is never assumed. Root deploying on behalf of a rootless user is a later step.

**Redeploy over a running service.** The platform decides, and an unchanged deploy changes
nothing. A compose file carries a label set from a digest of the file and every value the
compose child receives, so compose recreates exactly when one of them changed. On podman each
secret's stored value and the unit on disk are compared with the new ones; only a change is
written, and only a change restarts an active unit. When a running service would be recreated
or restarted, the tool asks before writing anything. A "no" changes nothing. `--no-prompt`
approves, and without a terminal the deploy fails with exit 1 naming it.

**Namespace gate.** The gate is the union of both tools' checks: what neither tool counts is
not counted, and what either tool asked before deleting is asked. It is a convenience, so
every way out short of a confirmed delete keeps the namespace with a warning, and none fails
the remove that ran it.

**Permission probes.** Probes cover writes only, and ask for the lowest permission that does
the job. A read the step depends on fails at the read with kubectl's own refusal. A read taken
only for convenience, such as a rollout wait or a read-back, warns when it is refused and
never blocks the command.

## Inherent differences

These follow from what each tool deploys and stay as they are.

- **Workload.** `solmq-conn-util` renders a Deployment and a Service. `solace-util` applies
  the EventBroker Operator's custom resource, and the operator owns the StatefulSets.
- **Exec target.** `solmq-conn-util` names the `connector` container and starts `sh`.
  `solace-util` starts `bash` or `cli -A` in the broker pod.
- **Registry.** `solmq-conn-util` pulls the image to read its entrypoint. `solace-util` logs
  in to the registry and never pulls.
- **Host limits.** The `fs.nr_open`, nofile, memlock, shm and session core checks match the
  broker's needs and stay in `solace-util`. `solmq-conn-util`'s unit asks for no core limit.

## Runner seams

Every external command in both tools crosses the shared `runner.Runner`. `Run` executes a
`runner.Cmd`, which holds the argv, the stdin bytes and any extra environment, and writes the
two output streams to the writers it is given. `Combined`, `Split` and `Output` are capture
helpers over `Run`. `Attach` hands the terminal to an interactive session and returns the
session's exit code.

`runner.Exec` resolves the program through `PATH` and refuses a match in the current
directory. A cancelled command is interrupted, then killed when a grace period ends; an
attached session is never interrupted. Errors read `<program>: <cause>` and never carry an
argument, the stdin bytes or an environment value. `runner.Echo` prints the command instead
of running it, with stdin as a byte count and every environment value masked.

What each tool keeps on top of the shared runner:

| Policy | solmq-conn-util | solace-util |
| --- | --- | --- |
| Verbose trail | none | `--verbose` announces each resolved command |
| `runner.Echo` | not used | the runner its tests install |
| Environment for the child | any command | compose only; every other holder refuses it |
| A cancelled log follow | a clean end, exit 0 | no follow is cancelled; Ctrl-C ends the process |

Output policy is the Output row under Divergences.

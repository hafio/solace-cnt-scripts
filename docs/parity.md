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

2. **Permissions are probed before the command that needs them.** The probe is
   `kubectl auth can-i <verb> <resource>`, with `--namespace` or `-n` when the resource is
   namespaced. A refusal stops the run and names the permission and the namespace.
   `solmq-conn-util`: `internal/runner/runner.go` (`Preflight`). `solace-util`:
   `internal/k8s/preflight.go`.

3. **The engine is checked first.** `<engine> info` runs before any other docker or podman
   command. A failure reports a platform hint, such as a stopped daemon or a missing rootless
   socket. `solmq-conn-util`: `internal/runner/runner.go` (`Preflight`). `solace-util`:
   `internal/container/preflight.go`.

4. **Podman units are driven through systemd.** After the unit file is written the tool runs
   `systemctl [--user] daemon-reload`, then `start`, `stop` or `restart <service>`. The
   restart count is read with `systemctl [--user] show <unit> -p NRestarts --value`.
   `solmq-conn-util`: `internal/runner/runner.go` (`PodmanDeploy`). `solace-util`:
   `internal/container/manager.go`, `internal/container/inspect.go`.

5. **Podman secrets arrive on stdin.** `podman secret create <name> -` reads the value from
   stdin. A value never appears on argv. Stale secrets go with `podman secret rm --ignore`.
   `solmq-conn-util`: `internal/runner/runner.go` (`PodmanSecretCreate`). `solace-util`:
   `internal/container/manager.go` (`createPodmanSecrets`).

6. **Exec, logs and upload share one argv shape.** Kubernetes uses
   `kubectl exec [-i] [-t] <pod> -n <namespace> -- <command>`. Docker and podman put their
   own flags first and use no `--`. A file is uploaded with `exec -i ... sh -c 'cat > <path>'`
   and the body on stdin. `kubectl logs` flags pass through; `--previous` is refused on docker
   and podman. `solmq-conn-util`: `internal/runner/runner.go`, `cmd/solmq-conn-util/shell.go`,
   `cmd/solmq-conn-util/logs.go`. `solace-util`: `internal/k8s/transport.go`,
   `internal/k8s/ops.go`.

7. **Inspect output is decoded the same way.** The leading slash is stripped from `Name`.
   Health comes from `State.Health` or, on older podman, `State.Healthcheck`. On podman the
   restart count comes from systemd, not from the engine. `solmq-conn-util`:
   `internal/statusreport/parse.go`. `solace-util`: `internal/container/inspect.go`.

8. **Containers are listed, then filtered in Go.** `<engine> ps --all --format` with a
   tab-separated template lists everything. Matching by exact name or by image substring
   happens in the tool. `solmq-conn-util`: `internal/runner/runner.go`. `solace-util`:
   `internal/container/manager.go`.

9. **An emptied namespace is offered for deletion, never assumed empty.** After the workload
   is gone the tool lists `all,persistentvolumeclaims,secrets,configmaps` in the namespace.
   `kube-root-ca.crt`, the `default` service account and its token secrets do not count.
   `default`, `kube-system`, `kube-public` and `kube-node-lease` are never deleted. A listing
   that fails keeps the namespace. `solmq-conn-util`: `cmd/solmq-conn-util/namespace.go`.
   `solace-util`: `internal/k8s/namespace.go`.

10. **Compose secrets come from the environment.** Each `secrets:` entry has an
    `environment:` source and an absolute `target:`. Values live only in the compose child's
    environment. `solmq-conn-util`: `internal/dockergen`. `solace-util`:
    `internal/render/render.go` (`Compose`).

11. **Hardening is written into every artifact.** Compose files carry `privileged: false`,
    `cap_drop: [ALL]`, `security_opt: no-new-privileges=true` and an explicit `user:`.
    Quadlet units carry `DropCapability=all`, `NoNewPrivileges=true`, `User=` and `Group=`.
    `solmq-conn-util`: `internal/dockergen`, `internal/podmangen`. `solace-util`:
    `internal/render/render.go`.

12. **Exit codes are 0, 1 and 2.** Success, runtime failure, usage error. A shell or CLI
    session inside the instance returns the child's own code. `solmq-conn-util`:
    `cmd/solmq-conn-util/shell.go`. `solace-util`: `internal/cli/exit.go`.

13. **Watch loops redraw in place and Ctrl-C exits 0.** `solmq-conn-util`:
    `cmd/solmq-conn-util/status.go`. `solace-util`: `internal/cli/watch.go`.

14. **Command binaries are bare names from an allowlist.** `command:` must name `kubectl`,
    `oc`, `docker` or `podman` without a path. The name is resolved through `PATH`, and a
    match in the current directory is refused. `--allow-command <name>` approves another
    name for one run. A privilege-escalation wrapper such as `sudo` is never accepted in a
    command, not even with `--allow-command`; run the tool itself under `sudo` instead. Both
    tools check commands with the shared `execguard` package. `solmq-conn-util`:
    `internal/validate`, `internal/runner/runner.go`. `solace-util`:
    `internal/config/execguard.go`, `internal/engine/runner.go`.

## Divergences

| Topic | solmq-conn-util | solace-util | Status |
| --- | --- | --- | --- |
| Compose file | handed to compose on stdin, never written | written beside the env file and compared on redeploy | decided, not yet applied: stdin |
| Confirmation without a terminal | fails at once with exit 1 and names `--no-prompt` | fails at once with exit 1 and names `--no-prompt` | aligned |
| Rootless podman | derived from the effective uid | declared as `podman.rootless`; a uid mismatch is refused; `XDG_RUNTIME_DIR` and the DBus address are repaired | decided, not yet applied |
| Quadlet directory | `$HOME/.config/containers/systemd`; `XDG_CONFIG_HOME` is not read | `$XDG_CONFIG_HOME/containers/systemd`, configurable | adopt: solmq-conn-util honours `XDG_CONFIG_HOME` |
| Host checks for rootless podman | none; a short subuid range warns | linger, session core limit, `fs.nr_open`, user manager limits, session nofile; a short subuid range refuses | adopt: solmq-conn-util takes linger, session repair and the core limit |
| Namespace on deploy | the Namespace document is applied on every deploy | read by name and created only when absent | adopt: solmq-conn-util |
| Waiting after a change | none | `rollout status --timeout=300s`, then the resource is read back | adopt: solmq-conn-util |
| Permission probes | one per verb, including status, logs and cli | every kind before the first write; none for status, logs and exec | adopt: solmq-conn-util probes every kind it writes; read probes in solace-util are open |
| `can-i` answer parsing | `no` and `no - ...` | last line; `no`, `no - ...` and `no (...)`; anything but `yes` refuses; unanswered is distinct from refused | adopt: solmq-conn-util takes the wider parser |
| Podman secret replace | `secret create --replace` on podman 4.7 and later, remove then create below | remove then create; skipped when the digest label is unchanged | adopt: solace-util uses `--replace` on 4.7 and later |
| Redeploy over a running service | restarts or recreates without asking | compares the artifact and the certificate digest, then asks; `--restart` pre-approves | open |
| subuid range too small | warning | error | open |
| Namespace gate details | discounts objects owned by the deployment; does not probe `delete namespaces` | counts every object; probes `delete namespaces` first | open |
| Output | buffered; one `ok:` or `error:` line at the end | streamed as it happens | open |

## Decisions

**Compose file.** The rendered compose file is handed to compose on stdin with `-f -`.
`--project-directory` is the directory that holds the env file. Nothing is written to disk.

**Confirmation without a terminal.** A destructive command whose stdin is not a terminal
fails at once with exit 1. The message names `--no-prompt`. No answer is read from a pipe.
Under a mintty-style shell stdin is a pipe even at a terminal, so pass `--no-prompt` there.

**Rootless podman.** A non-root invoker deploys rootless. A declared `rootless: false` under a
non-root uid is an error that names the uid. Root follows the declared value. A declared
rootless deploy as root means deploying for the declared non-root user. Until that is built,
both tools refuse that combination with the same message.

**Redeploy over a running service.** The platform decides what changed. Compose recreates a
container whose configuration changed. systemd restarts a unit whose file changed. Whether a
confirmation precedes the restart is still open.

## Inherent differences

These follow from what each tool deploys and stay as they are.

- **Workload.** `solmq-conn-util` renders a Deployment and a Service. `solace-util` applies
  the EventBroker Operator's custom resource, and the operator owns the StatefulSets.
- **Exec target.** `solmq-conn-util` names the `connector` container and starts `sh`.
  `solace-util` starts `bash` or `cli -A` in the broker pod.
- **Registry.** `solmq-conn-util` pulls the image to read its entrypoint. `solace-util` logs
  in to the registry and never pulls.
- **Host limits.** The `fs.nr_open`, nofile, memlock and shm checks match the broker's needs
  and stay in `solace-util`.

## Runner seams

Every external command goes through one runner interface per tool. The two interfaces differ,
so the platform code above is written twice rather than shared.

| Capability | solmq-conn-util `runner.Runner` | solace-util `engine.Runner` |
| --- | --- | --- |
| Combined stdout and stderr capture | `Run` | none |
| Separate stdout and stderr capture | `RunSplit` (optional) | `Output`: stdout captured, stderr live |
| Stream to any writer | `Stream(ctx, cmd, out, err)` | `Run` to the process's stdout and stderr only |
| Context | `Stream` only | every call |
| Environment for the child | on any call | `EnvRunner` only |
| Stdin and environment together | yes | no |
| Interactive terminal | `Attach(cmd, files)` returns the exit code | `RunInteractive(ctx, ...)` returns an error |
| Dry run, masked environment, announce | none | `Echo`, `MaskEnv`, `Announce` |

Neither interface is a superset of the other. Each call site is written against its own
runner's output and error model.

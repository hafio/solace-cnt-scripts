# Decisions on the 2026-11-11 critical review

Owner decisions on every finding in
[2026-11-11-critical-review.md](2026-11-11-critical-review.md), taken 2026-11-13. Nothing here
is built yet. Evidence and suggested diffs stay in the report.

Home says where the work lands:

- `here`: this repository.
- `gdu N`: item N of `docs/todo.md` in go-deploy-utility, for the part both tools share.
- `init-devops`: the skill's templates or skeletons first, then a re-deploy here.
- `-`: no work.

## Rules the owner stated

- Questions come first and actions after. Every question a command asks is answered before
  its first write or delete. A step that needs no question just runs. Questions are for the
  destructive steps already documented as gated, and no new ones are added.
- The tool acts on what the env file names. It does not go looking (CR-0009).
- Data replication and DR switchover are not finished. Findings in that code are parked
  until that work resumes.
- Behaviour both tools share is built once in go-deploy-utility and imported.

## Decisions

| ID | Severity | Decision | Home |
|----|----------|----------|------|
| CR-0001 | blocker | An import limited to chosen VPNs leaves virtual hostnames alone. A whole-broker import still clears stale ones, and names them in the plan and in the confirmation. | here |
| CR-0002 | blocker | The seven `kubernetes.*` values are held to a grammar at load (report patch). Nothing is quoted at the sink, so the goldens stay. | here |
| CR-0003 | blocker | `kubernetes.storage.class` is checked at load and again right before kubectl is called. | here, gdu 4 |
| CR-0004 | blocker | The watch namespace read back from a running operator is validated before it is written into the operator bundle (report patch). | here |
| CR-0005 | major | After the down on docker, and after the unit stop on podman, the engine is asked whether the container is still there. If it is, remove fails before any data is emptied. | here, gdu 9 |
| CR-0006 | major | Parked on the shared module: its namespace listing must cover more kinds. No change here until then. | gdu 1 |
| CR-0007 | major | Parked with the DR work. | here |
| CR-0008 | major | Parked with the DR work. | here |
| CR-0009 | major | Only the operator namespace the env file resolves to is read. Discovery of an operator in other namespaces is removed. | here |
| CR-0010 | major | The import learns which CLI users the target already has, and turns a create for one of them into an edit that keeps its settings. Done together with CR-0081. | here |
| CR-0011 | major | The shutdown warnings appear in the plan, before anything is torn down (report patch). | here |
| CR-0012 | major | `operator remove --delete-crd` fails loudly when the operator is kept. That is decided before the CRD question is asked. | here |
| CR-0013 | major | `dataDir` has a floor: `/`, the home directory itself and system directories are refused, and `/data` stays valid. The `--delete-data` question shows the resolved path. | here |
| CR-0014 | major | A podman unit carries a deploy-digest label, and a redeploy compares it with the running container's label, as on docker. | here, gdu 10 |
| CR-0015 | major | `convert` copies values as written. It expands nothing and warns about nothing. | here, gdu 2 |
| CR-0016 | major | Parked with the DR work. | here |
| CR-0017 | major | Sanitised copies of the real broker captures are committed as test data, and the tests fail instead of skipping without them. The owner supplies the originals. | here |
| CR-0018 | minor | All twelve platform-specific commands are built by the one builder the others use. Behaviour does not change. | here |
| CR-0019 | minor | The echo-runner branches stay and are relabelled as a test seam. The three dry-run passages leave operations.md. | here |
| CR-0020 | minor | The protocol in `kubernetes.ports` is upper-cased when rendered (report patch). The generator page gets the same line and a manual cross-check. | here |
| CR-0021 | minor | Parked with the DR work. The shared builder's name check covers the kubectl side. | here, gdu 4 |
| CR-0022 | minor | A host name is validated before it fills a missing router name. The error says to set `redundancy.primary.name`. | here |
| CR-0023 | minor | The broker type is read from the banner only (report patch). | here |
| CR-0024 | minor | Parked with the DR work. | here |
| CR-0025 | minor | Parked with the DR work. | here |
| CR-0026 | minor | The kube context the prompt names is pinned on every kubectl call of the run. The helper comes from the shared module. | here, gdu 11 |
| CR-0027 | minor | An empty answer to the broker listing refuses the CRD delete (report patch). | here |
| CR-0028 | minor | A readiness condition that is missing or Unknown is shown as not reported, on an `[INFO]` line that names it (report patch). | here |
| CR-0029 | minor | `runUser` uid 0 on docker and rootful podman loads, and deploy warns that the broker runs as host root. container-security.md softens "enforce". | here |
| CR-0030 | minor | A `%` in `dataDir` is escaped in the `Volume=` line with the shared `quadlet.EscapePercent`. A `:` in it is refused at load. | here, gdu 4 |
| CR-0031 | minor | Container deploy asks the redeploy question before any host change: sysctl, linger, registry login and the data directory's owner. | here, gdu 3 |
| CR-0032 | minor | Bare `auto-complete` exits 2 (report patch). | here |
| CR-0033 | minor | `convert` keeps printing the converted file with its secrets. CLAUDE.md and operations.md name it as the second place secrets are shown. | here |
| CR-0034 | minor | `image.user` without a password, or a password without a user, fails at load on every platform. A user with `image.passEnv` is valid. | here |
| CR-0035 | minor | Each retired v0.2.x key is refused by name with its replacement: `nodes`, `nodes.*.ip`, `storage.msgNode`, `storage.monNode` and the scalar `redundancy`. | here |
| CR-0036 | minor | Load and `convert` share one prepare step, so `convert` no longer warns about `~` paths that load fine. | here, gdu 2 |
| CR-0037 | minor | A second YAML document in an env file is an error (report patch). | here |
| CR-0038 | minor | The import driver deletes itself from the broker when it ends. Ctrl-C during an import is caught and says the broker may still be applying. There is no run lock. | here |
| CR-0039 | minor | The driver checks each chunk and wrapper write and refuses an empty transcript (report change). The CLI skeleton run gets the same check. | here |
| CR-0040 | minor | Parked with the DR work. | here |
| CR-0041 | minor | Deletes during `broker remove` are bounded, and a timeout says what is stuck. The shared delete gains the timeout, and the PVC delete is bounded here. | here, gdu 12 |
| CR-0042 | minor | Stale podman secrets are removed by ownership labels. Every secret is created with the labels, and a deploy or a remove deletes the labelled ones the env file no longer names. Nothing is asked, and the tool names what it removed. The shared module owns the labels and the removal. | gdu 16, here |
| CR-0043 | minor | `--pod` is checked before a watch starts (report patch). The shared loop stops on a usage error. | here, gdu 13 |
| CR-0044 | minor | `vulnjudge` requires govulncheck's opening config message, so an empty or cut stream fails (report patch). | here |
| CR-0045 | minor | Parked with the DR work: all six unused functions are handled then. | here |
| CR-0046 | minor | The set/MISSING and none helpers move into `internal/output`, with one spelling, `(none)`. | here |
| CR-0047 | minor | Every listed error-path test asserts the error text and that no later step ran. | here |
| CR-0048 | minor | The three secret-leak tests check stdout, stderr and the returned error, for every secret in the fixture. | here |
| CR-0049 | minor | Both workflows are re-copied from the init-devops templates as they are, and CI runs `all cov scan`. The dev scripts first learn to run the suite once for `all cov`. | here |
| CR-0050 | minor | The template's Windows leg runs Windows PowerShell 5.1. Then the workflow is re-copied here. | init-devops |
| CR-0051 | minor | The dev scripts gain `--local`. Without it a run builds the pinned shared module, as CI does. With it the run builds against the local checkout and says so. `go.work` stays uncommitted. | here, gdu 15 |
| CR-0052 | minor | `vet` gains the `gofmt -l` check in both scripts. | here |
| CR-0053 | minor | A test here ties the version stamp to the module path. The template's release job compares the binary's version with the tag. | here, init-devops |
| CR-0054 | minor | One-line fix for the empty array on old bash (report patch). | here |
| CR-0055 | minor | `broker stop` and `perform redundancy-test` ask before acting, as restart does. `redundancy-test` asks after its checks. | here |
| CR-0056 | minor | `redundancy-test` uses the common `--pod` check (report patch). | here |
| CR-0057 | minor | Answering n to the `--out` overwrite question exits 0 and says the file was kept. | here |
| CR-0058 | minor | `--tail` accepts a count only, and leaving it out means every line. Help follows. | here, gdu 14 |
| CR-0059 | minor | `validate` reports every env-file problem in one pass. | here |
| CR-0060 | minor | The four help and README texts are corrected (report patch), then `regen`. | here |
| CR-0061 | minor | Both messages name the current commands (report patch). | here |
| CR-0062 | minor | The final error line is printed as `[FAIL]` through `internal/output`. operations.md follows. | here |
| CR-0063 | minor | Accepted. | - |
| CR-0064 | minor | Both scripts follow the contract: the race detector runs when gcc is on PATH, and the skip is logged otherwise. `SOLACE_RACE` is read the same way in both. | here |
| CR-0065 | minor | `gen-cert.sh` is removed from the repository. | here |
| CR-0066 | minor | The three doc passages are corrected: the CI row, the regen package list and the shared-package list in CLAUDE.md. | here |
| CR-0067 | minor | The `typescript` capture is deleted and the name is git-ignored. | here |
| CR-0068 | nit | Apply the report's patch, and correct the stale comment beside it. | here |
| CR-0069 | nit | `broker status --pod` narrows the report to that pod, with or without `--detail`. | here |
| CR-0070 | nit | Apply the report's patch. | here |
| CR-0071 | nit | Errors over map-typed fields report every bad entry, sorted so tests stay stable. Part of CR-0059. | here |
| CR-0072 | nit | Apply the report's patch. | here |
| CR-0073 | nit | Parked with the DR work. | here |
| CR-0074 | nit | Parked with the DR work. | here |
| CR-0075 | nit | The stale comments and catalogue rows are corrected. | here |
| CR-0076 | nit | The misnamed test is renamed and its stdout swap is restored with `t.Cleanup`. Nothing else changes. | here |
| CR-0077 | nit | Old review ids in tests and code become plain statements. | here |
| CR-0078 | nit | The non-ASCII bytes in the two files become ASCII. | here |
| CR-0079 | nit | Accepted. | - |
| CR-0080 | nit | `docs/review/` joins `.graphifyignore`, the dead ignore entries go, and CLAUDE.md says `docs/review/` is exempt from the notice rule. | here |
| CR-0081 | suspected | Checked against the real capture first. If confirmed, fixed together with CR-0010. | here |
| CR-0082 | suspected | Fixed by the context helper of CR-0026: the context the calls really use is the one named and pinned. | here, gdu 11 |
| CR-0083 | suspected | Waits on a cluster check. `OnDelete` means the wait changes to pod readiness. `RollingUpdate` closes it. | here, gdu 17 |
| CR-0084 | suspected | A CLI rejection is detected from the last non-blank line, trimmed, when it equals one of the six captured rows below. A separate check catches a CLI that did not run. `busy` and `command not found` leave the keyword list. | here |
| CR-0085 | suspected | Left open. No fix is planned, and it is not accepted. | - |
| CR-0086 | suspected | A repeated `--vpn` is dropped silently. | here |
| CR-0087 | suspected | Parked with the DR work. | here |
| CR-0088 | suspected | Left open. No fix is planned, and it is not accepted. | - |
| CR-0089 | suspected | Rootless podman on a cgroup v1 host is refused up front, with a remedy that works. The docs gain the prerequisite. | here |
| CR-0090 | suspected | developer.md says `BUILD_TARGETS` is the release set and gives its value. The release plan job prints the resolved targets. | here, init-devops |

Parked with the DR work: CR-0007, CR-0008, CR-0016, CR-0021, CR-0024, CR-0025, CR-0040,
CR-0045, CR-0073, CR-0074 and CR-0087. They stay open in the ledger, and a later review
re-verifies them. HA redundancy is a different feature and is not parked.

CR-0085 and CR-0088: the owner said "leave it". They stay open in the ledger and will be
listed again by the next review.

CR-0084, the six rows the owner captured from the Solace CLI. Each is a whole line. The
comparison is case-insensitive, which is an assumption stated to the owner and not yet
confirmed:

```
Invalid command input
missing command
-> unknown command
ERROR:
Command Failed
-> missing parameter
```

## Planned work the report does not hold

### Certificate renewal on Kubernetes without a restart

Raised 2026-11-13 under CR-0083. After a renewed server certificate on Kubernetes the tool no
longer restarts the pods. The owner's statement is that the operator watches the TLS Secret.
The tool prints a line that tells the operator to check the pods are restarting. The restart
step, its question, its help text, its docs and its tests go for Kubernetes. The next review
run gives this an id.

### Stale objects after a config change

Raised 2026-11-13 from CR-0042. A secret or an object the tool created, and the env file no
longer names, is removed by ownership labels. The labels and the removal are one standard
behaviour in go-deploy-utility that this tool and solmq-conn-util import, and the design is
worked through there (gdu 1, gdu 16, gdu 18). For this tool:

- Podman: CR-0042 above. A secret the env file still names and that carries no label is
  created again once with the same value, so it gains the labels.
- Kubernetes: the admin, TLS, image-pull and additional-users Secrets are created with the
  labels. A deploy and a remove list the instance's labelled Secrets and delete the ones the
  env file no longer supplies. Today such a Secret stays through every later deploy and
  through `broker remove`. The next review run gives this an id.
- An unlabelled secret or Secret is never deleted. That protects one the operator made and
  the ones the Solace operator generates.
- Docker: nothing to do. A compose secret is copied into the container and nothing is stored
  outside it.
- Known limit: a secret dropped before the labels exist stays unlabelled and is never found,
  so it is removed by hand.

### Shared module

`docs/todo.md` in go-deploy-utility holds the shared part of these decisions. The shared
change and a pin bump come before the work here that uses it.

- gdu 1: ownership labels and a wider namespace listing (CR-0006, the stale objects above).
- gdu 2: expansion in config values (CR-0015, CR-0036).
- gdu 3: one step order for deploy and remove (CR-0031).
- gdu 4: builders refuse a name outside its grammar, and quadlet lines escape `%` (CR-0003,
  CR-0021, CR-0030).
- gdu 9: remove is idempotent, and the container is checked after the down (CR-0005).
- gdu 10: a deploy digest for podman units (CR-0014).
- gdu 11: resolve and pin the kube context (CR-0026, CR-0082).
- gdu 12: a bound on every delete (CR-0041).
- gdu 13: a watch loop stops on an error no retry can fix (CR-0043).
- gdu 14: one check for the log tail count (CR-0058).
- gdu 15: `--local`, and what `dev-all` and its consuming guide need for it (CR-0051).
- gdu 16: stale podman secrets by labels (CR-0042).
- gdu 17: a readiness wait for OnDelete StatefulSets (CR-0083, if confirmed).
- gdu 18: stale Kubernetes objects by labels.

## Still undecided

- CR-0085 and CR-0088: accepted, or parked for later.
- CR-0069: what `--pod` together with `--all` does.
- gdu 1: the label key names.
- gdu 16 and gdu 18: the shape of the shared code, and what a deploy does when the label
  listing is refused. For this tool only Secrets are in play on Kubernetes.

## Consequences to check before the work starts

- CR-0002, CR-0013, CR-0030, CR-0034, CR-0035 and CR-0037 turn env files that load today into
  load errors.
- CR-0009: an operator installed in another namespace is no longer detected. The docs say so.
- CR-0015: a bash file that relied on `${VAR}` being expanded converts to the literal text.
- CR-0014 and CR-0042 change every quadlet golden, since the unit gains label lines. The
  first deploy after CR-0042 writes each podman secret once more with the same value.
- CR-0029 softens a guarantee container-security.md states today.
- CR-0032 and CR-0057 change exit codes: 1 to 2, and 1 to 0.
- CR-0049: the dev scripts must run the suite once for `all cov` before CI calls it.
- CR-0051: a plain local run builds the pinned module. `dev-all` in go-deploy-utility must
  pass `--local`, or its steps before a tag fail here (gdu 15).
- CR-0055: a script that calls `broker stop` or `perform redundancy-test` without a terminal
  needs `--no-prompt`.
- CR-0059: tests that expect the first error only change.
- CR-0020 adds an edit to the generator page, on top of the two cross-checks still owed.

## Steps the owner runs

These can wait until the work is under way.

- CR-0083: `kubectl get sts <name>-pubsubplus-p -n <ns> -o jsonpath='{.spec.updateStrategy.type}'`.
- CR-0017 and CR-0081: supply the real broker captures.
- CR-0090: `gh variable get BUILD_TARGETS`.
- CR-0064: whether the Windows runner has gcc.
- CR-0052: one `gofmt -l` run over the tree before the check is switched on.
- The generator page cross-checks in developer.md: `kubernetes.ports` (2026-10-02), joinPEM
  (2026-10-23) and the CR-0020 edit.
- CR-0042: on the oldest supported podman, that `podman secret create --label` and
  `podman secret inspect` behave as documented.
- CR-0084: the capture is at `tmp/solace-cli-err.out`, which git ignores. Its rows are copied
  above.
- Every commit, and the gates after each change.

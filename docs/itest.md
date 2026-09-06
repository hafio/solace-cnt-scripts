# Live-environment probe harness (`solace-itest`)

`internal/tools/itest` is a **dev-only** binary that checks, against real
infrastructure, the handful of claims this repository's source marks as assumed
but never verified. It is never shipped: it is absent from `BUILD_TARGETS` and
from the release workflow, and the `itest` dev-script task builds it on demand.

## Why it exists

The unit suite deliberately cannot answer these questions. Every external effect
in this codebase is seamed so that tests run with no cluster, no container engine
and no broker (see [test.md](test.md)) -- which is what makes the suite fast and
hermetic, and also what leaves a handful of comments in the code saying, in
effect, *this should be true, but nobody has checked*.

This tool is how one gets checked. It produces **evidence**, not a gate: a run's
job is to let the comment be replaced by a fact, and where the assumption turns
out to be wrong, to name the follow-up fix.

It drives the real production code paths -- `config.Load`, `engine.Exec`,
`k8s.Cluster`, `container.Manager`, `broker.Ops` -- because a harness that
reimplemented them would only prove things about itself.

## Operator-run only

**An AI assistant must never invoke this binary.** Its probes mutate live broker
state; deciding to do that to a given environment belongs to a person. The dev
scripts only ever *build* it, never run it, for the same reason.

## The probes

| ID | Claim it settles | Platform | Effect | What it touches |
| --- | --- | --- | --- | --- |
| `p1` | `internal/container/manager.go`, `deployPodman` -- ASSUMED, NOT VERIFIED: a quadlet re-reads the secret store on unit start | podman | **mutates** | Stops the broker unit, rotates the admin-password secret to a throwaway value, starts the unit, reads the mounted secret, then restores the original and bounces again |
| `p2` | `internal/broker/semp.go`, `revertActivityMateBody` -- NEEDS VERIFICATION ON A LIVE BROKER: the `/SEMP` path, the ok-reply shape, and idempotence on an already-standby mate | docker, podman | **mutates** | Reads `show redundancy` locally, GETs the mate's SEMP monitor endpoint, and -- only if the mate is provably standby -- POSTs one `redundancy revert-activity`, which should be a no-op |
| `p4` | `internal/render/render.go` quadlet cpu/mem -- ASSUMED, NOT VERIFIED: `PodmanArgs=--cpus=` and `Memory=` actually constrain the container | podman | read-only | Runs `podman inspect` against the already-deployed container |

### What each verdict means

- **VERIFIED** -- the assumption held here. Replace the source comment with
  `VERIFIED on <date> against <version>` in the same change.
- **REFUTED** -- the assumption is wrong. This is the valuable outcome. Each
  probe's detail names the follow-up fix:
  - `p1` -> `deployPodman`'s not-running branch needs docker's force-recreate
    equivalent: restart, not start, when secrets were just written.
  - `p2` -> fix `sempV1OKRe`, or add the benign-error carve-out, from the reply
    the probe printed. Keep that reply as a fixture under
    `internal/broker/testdata/`.
  - `p4` -> fold `Memory=` into the `PodmanArgs=` line, which `render.go` already
    names as the fallback.
- **INCONCLUSIVE** -- the probe ran but the evidence does not settle the claim
  (a preflight failed, the mate was active, the broker answered something
  unfamiliar). Not a defect, so it exits 0. Downgrade the comment to name the
  specific unknown rather than deleting it.
- **SKIPPED** -- the probe does not apply to this platform, or consent was
  declined. Reported, never silently omitted: a missing row would read as a pass.

Exit codes: `0` nothing refuted, `1` at least one claim REFUTED, `2` usage or
setup error.

## Safety

- Every mutating probe asks first, and the answer must be the exact word `yes`
  (the same bar `remove broker --delete-data` sets). `--yes` pre-approves for an
  unattended run; a non-interactive run without it **declines**, so a forgotten
  flag cannot silently rotate a production secret.
- Consent is asked **per probe**, not once per run.
- `p2` refuses to send `revert-activity` unless the primary's own
  `show redundancy` says the mate is standby -- sending it to an active mate is a
  real failover. Afterwards it re-reads redundancy and escalates to REFUTED if
  activity moved, so the probe checks its own safety premise rather than trusting
  it.
- `p1` restores what it changed through a `defer`, so cleanup runs even
  when a step in the middle fails. If a restore itself fails, the warning names
  the secret, because a human then has to finish the job by hand.
- Every wait is bounded. A probe reports INCONCLUSIVE rather than hanging.

## Build and run

Build for the host, or cross-compile for the target:

```
./scripts/dev.sh itest                                    # host
TARGET_OS=linux TARGET_ARCH=amd64 ./scripts/dev.sh itest   # for a Linux host
```

Windows: `.\scripts\dev.ps1 itest`. Output lands in
`dist/solace-itest-<os>-<arch>[.exe]`, version-stamped from the same
`git describe` the `build` task uses, so a result can be pinned to the code that
produced it (`--version`).

Then, **on the target host**:

```
./solace-itest --list                        # what each probe touches; needs no env file
./solace-itest -e prod.yaml --dry-run p1     # read-only steps for real, mutations previewed
./solace-itest -e prod.yaml p1               # run it, asking before anything mutating
./solace-itest -e prod.yaml --yes p1 p4      # unattended
```

All three probes run on the broker host itself (`p2` on the **primary**). No probe
currently targets Kubernetes: the one that did, on config-sync replicating
additional CLI users, was settled on 2026-08-21 without a test -- config-sync
replicates both `username` and `client-username`, so the comment in
`internal/cli/ops_k8s.go` now records that rather than asking. Pointing this
harness at a `kubernetes:` env file is therefore valid but reports three SKIPs.

`--dry-run` runs every read-only step for real -- the preflights, the role
detection, the redundancy read -- then stops at the first mutation and reports
what it would have done. It deliberately does not swap in an echo runner and
carry on: half the probe would run for real and half would be echoed, and the
verdict would describe neither.

### Flags

| Flag | Meaning |
| --- | --- |
| `-e`, `--env` | env file name or path (default `env.yaml`) |
| `--base-dir` | directory searched for the env file |
| `--platform` | `kubernetes\|docker\|podman`; needed only when the file declares several. This tool never prompts -- an ambiguous file is a hard error, because it is built to run unattended on a target host |
| `--probe` | comma-separated probe ids; same as positional args |
| `--list` | print the catalogue and exit; needs no env file |
| `--version` | print the build version and exit |
| `--yes` | pre-approve every mutating probe |
| `--dry-run` | real read-only steps, then stop before mutating |
| `-v`, `--verbose` | announce every external command as it runs |
| `--allow-command` | approve a binary the env file names (repeatable), as on the CLI |

## After a run

Feed the result back into the code in the same change:

1. Update the source comment the probe names -- VERIFIED with the date and the
   broker/podman version, or the specific unknown that remains.
2. For a REFUTED claim, make the fix its own scoped unit of work.
3. Replace the hand-authored fixtures with what the live run actually returned.
   `sempOKReply` in the harness tests and the `show redundancy` shapes in
   `internal/broker` were written from the documentation, not from a broker.

## Testing the harness itself

The harness's own logic is unit-tested like any other package -- over a fake
`engine.Runner`, with no live environment (see the `internal/tools/itest` section
of [test.md](test.md)). Those tests cover what a live run cannot: the **order** of
each probe's steps, and that cleanup still happens when a step in the middle
fails. A probe that skipped its restore on the error path would leave a throwaway
password on a real broker, and that is not something to discover by running it
against one.

# Test catalogue

Every Go test in this repository, grouped by package and file, with a one-line statement of
what each one proves. Use it to find existing coverage before adding a test, and to spot
what is *not* covered.

This file is maintained by hand. When you add, rename, or delete a test, update the matching
row in the same change (CLAUDE.md S6).

## Running the tests

Tests are run through the mirrored dev scripts, never with a bare `go test` in CI:

| Want | Windows | Linux/macOS |
| --- | --- | --- |
| Run everything | `scripts\dev.ps1 test` | `./scripts/dev.sh test` |
| Coverage profile + total | `scripts\dev.ps1 cov` | `./scripts/dev.sh cov` |
| Build + vet + test (CI's gate) | `scripts\dev.ps1 all` | `./scripts/dev.sh all` |

- The `test` task runs `go test -count=1 ./...`. Race detection is on by default in `dev.sh`;
  on `dev.ps1` it is opt-in with `SOLACE_RACE=1`.
- `cov` writes `coverage/coverage.out` and `coverage/coverage.html` and prints the total. The
  previous total in `scripts/logs/cov.log` is the local floor -- an unexplained drop is a
  failed gate. CI is a fresh checkout with no prior log, so it cannot catch a coverage
  regression; that check is local only.
- Per-task logs land in `scripts/logs/<task>.log`.

Narrowing a run during development (not a substitute for the gate):

```
go test ./internal/config -run TestResolveEnvPath -v
```

Four packages carry golden files and accept `-update` to regenerate them. The `regen` task
does all four; only run it after eyeballing the diff, since the committed goldens ARE the
reviewed expected output:

```
./scripts/dev.sh regen        # or: .\scripts\dev.ps1 regen
```

`regen` is deliberately absent from `all` and `full`: those gate, and a gate that rewrote
the thing it compares against could never fail. `test` is what reports a stale golden.
The individual invocations behind it, when only one package needs rewriting:

```
go test ./internal/render -update
go test ./internal/k8s -update
go test ./internal/convert -update
go test ./internal/cli -update      # rewrites docs/commands.md
```

Every fixture a test reads must be committed. `bash/` is gitignored in its entirety, so no
test may point at it -- a fresh CI checkout has no such files.

## Summary

77 test files, 1248 test functions. Three of those are not tests. Two are os/exec
helper-process shims, each a no-op unless its own environment variable is set:
`TestHelperProcess` in `internal/engine` (`GO_WANT_HELPER_PROCESS=1`) and
`TestHelperExitProcess` in `internal/cli` (`SOLACE_TEST_CHILD_EXIT_CODE`), which exists
because a real `*exec.ExitError` cannot be constructed by hand. The third is `TestMain`
in `internal/cli`, which points `os.Stdin` at an already-closed pipe for the whole package
so a confirm helper under test reads EOF instead of blocking on the console the suite was
launched from.

| Package | Files | Tests |
| --- | --- | --- |
| internal/k8s | 18 | 200 |
| internal/broker | 22 | 426 |
| internal/cli | 11 | 188 |
| internal/config | 11 | 164 |
| internal/container | 6 | 132 |
| internal/convert | 1 | 37 |
| internal/render | 2 | 32 |
| internal/engine | 2 | 26 |
| internal/output | 1 | 16 |
| internal/tools/vulnjudge | 1 | 11 |
| internal/abbrev | 1 | 8 |
| internal/examples | 1 | 8 |
| **Total** | **77** | **1248** |


## Coverage

Last recorded run, from `scripts/logs/cov.log` (2026-09-11), total **94.5%**. Re-run `cov`
after any change; these figures go stale the moment tests move, and the previous total is
the floor the next run has to hold.

**94.1%/94.2% -> 94.5%, and the move is a coverage-import-then-close, not a regression.**
The config-export/import feature added roughly 2300 lines of new code to `internal/broker`
(`blocks.go`, `annotate.go`, `sections.go`, `inject.go`, `diff.go`, `importops.go`) with no
tests yet behind them, which briefly took that package to 52.0% and the total to 84.7% --
adding uncovered code lowers the ratio the same way deleting covered code does (see the
96.1% note below). Seven new test files (`blocks_test.go`, `sections_test.go`,
`inject_test.go`, `diff_test.go`, `annotate_test.go`, `importops_test.go`,
`importdoc_test.go`) then closed it, taking `internal/broker` to 95.0% and the total to
94.5%, above the previous floor either doc had recorded. `internal/cli` moved the same way
on a smaller scale: `exportconfig.go`'s two new leaves (`broker perform
export-config`/`import-config`) plus their wiring in `commands.go`/`helpers.go`/
`ops_k8s.go`/`ops_container.go`/`aliases.go` took it to 84.9% before
`exportconfig_test.go`'s 13 tests brought it to 90.1%. docs/developer.md last recorded
94.1% and this file 94.2% -- 94.5% supersedes both.

**One dip along the way, and what it caught.** A run during the hostname-fallback work came
in at 94.1%, 0.1pp under the 94.2% floor, and `internal/config` was 0.4pp of that: a
`Config.node(Role)` helper an earlier design of that work read the node table through, which
the shipped version does not call at all. `go vet` does not flag an unused method, so nothing
but the coverage number noticed. Deleting it put `internal/config` back to 97.8%, which is
the whole argument for treating an unexplained drop as a failed gate rather than a rounding
error -- the number was the only thing reporting dead code.

The other 0.1pp was two defensive branches in `defaultLocalAddrs` (a failed
`net.InterfaceAddrs` read, and the non-`*net.IPNet` address form) and stays uncovered
deliberately: faking either means faking the stdlib, and `TestDefaultLocalAddrs` covers the
path that actually runs.

| Package | Coverage |
| --- | --- |
| internal/abbrev | 100.0% |
| internal/output | 100.0% |
| internal/tools/vulnjudge | 98.9% |
| internal/config | 97.8% |
| internal/render | 97.4% |
| internal/convert | 97.5% |
| internal/broker | 95.0% |
| internal/container | 94.8% |
| internal/k8s | 92.6% |
| internal/cli | 90.1% |
| internal/examples | 90.0% |
| internal/engine | see below |

**The total is down from 96.1%, and the reason is recorded rather than absorbed.** The
noun-first command-tree overhaul added two substantial mechanics -- the operator watch-list
reconciliation (`internal/k8s/watch.go`) and the namespace occupancy gate
(`internal/k8s/namespace.go`) -- and deleted a lot of well-covered code with it: the
node-labelling surface, `render.SecretScript`, and the `deploy all` / `remove all`
orchestration. Deleting covered code lowers the ratio on its own; adding uncovered code
lowers it twice.

`watch_test.go` and `namespace_test.go` closed the worst of it -- `namespace.go` went from
0% to covered and `internal/k8s` from 84.2% to 90.5%. The redundancy-block reshape then added
the deploy read-back (`ConfirmBrokerApplied`) and the custom-volume-mount skip in
`DeleteBroker`, both of which shipped uncovered; `deploy_test.go`'s seven new cases closed
them and took `internal/k8s` to 92.1% and the total back over its floor.

**94.3% -> 94.2%, and the 0.1pp is a deletion rather than a regression.** Applying
`semp.additionalUsers` declaratively on Kubernetes retired the broker-CLI route it
replaced: `broker.AdditionalUsers`, `additionalUsersScript`, `validCLIPassword` and both
copies of `cliForbiddenPassword` are gone, along with their six tests. That code was
well covered, so removing it lowers the ratio on its own -- `internal/broker` 93.3% ->
92.9% is almost all of the move. The code that replaced it is smaller and fully covered:
`AdditionalUsersSecret` and its four cases, plus the two username rules.

What remains uncovered is concentrated in `internal/cli/ops_k8s.go`, the thin handlers
wiring these mechanics to commands, and in the error arms of `internal/k8s/operator.go`,
`internal/k8s/checkreport.go` and `internal/k8s/prep.go`. **That is the gap to close before
the next feature**, and 94.2% is the floor until it is.

**`internal/engine` is not currently measurable.** It reports `0.0%` in this run and
`100.0%` in others, with every test passing both times -- so the figure is a measurement
artifact, not coverage that vanished. The package's `Exec` tests re-execute the test binary
as a child process (`TestHelperProcess`/`helperCommand`), and a coverage-instrumented child
can clobber the parent's profile; the run also slows noticeably when it happens. Until that
is fixed, treat the total above as understated by roughly engine's share.

`main.go` reports `0.0%` and always will: it is three statements -- call `Execute`, print
what came back, exit with `ExitCode`'s answer -- and no test can drive them, which is
exactly why the classification lives in `internal/cli` (exit.go) rather than in main.

---


## internal/config

Config loading, defaults, validation, and env-file resolution, plus the `Command`
type behind the platform CLI overrides and the execution guard that decides what a
`Command` may be, the scaling block that sizes the broker on every platform, and the
platform vocabulary the CLI resolves against, the host-path rules every file-valued
key is held to, and the two storage stories a Kubernetes deployment may tell.
164 tests across 11 files.

### command_test.go

| Test | What it covers |
| --- | --- |
| `TestCommandUnmarshal` | Both accepted forms of a `Command`: a scalar split on whitespace (reproducing the bash bootstraps' unquoted expansion, so a quoted scalar still splits and whitespace runs collapse), and a sequence kept token-for-token -- the only way to express a token containing a space. Empty scalar, empty list, and an omitted key all decode to nothing, so the default applies |
| `TestCommandUnmarshalRejectsOtherKinds` | A mapping is neither a command line nor an argv, so it fails loud at decode naming the accepted forms |
| `TestCommandUnmarshalPropagatesDecodeErrors` | A node of an accepted kind whose contents still will not decode (`!!binary` with invalid base64; a sequence element that is not a scalar) surfaces yaml's error instead of falling through to an empty command |
| `TestCommandNameAndArgs` | `Name`/`Args` split a command into argv[0] and the leading arguments that precede each call's own, including the unset and bare-binary cases |
| `TestCommandArgsDoesNotAliasCommand` | `Args` allocates: with spare capacity in the backing array, a naive `append(cmd[1:], ...)` would corrupt the previous call's argv. Two successive calls must stay independent and the `Command` itself unchanged |
| `TestCommandString` | Display rendering, used by the check reports and error messages |
| `TestValidateProbeCommandAccepts` | The container health-check probe keeps the loose rules, because it runs inside the broker rather than here: a path, and a shell pipeline with metacharacters, both pass, as does an unset command. The field-by-field opposite of `TestCheckCommandRejects` |
| `TestValidateProbeCommandRejects` | Even the probe rejects empty arguments and control characters (newline, NUL), naming the field and the offending index |
| `TestValidateRejectsBadRuntime` | A malformed runtime fails `Validate` for all three platforms, ahead of the mandatory-field checks, so the message names the runtime rather than the fields also missing |
| `TestRuntimeDefaults` | Defaults resolve to exactly one token with no leading args (`kubectl`/`docker`/`podman`), so existing argv is byte-identical; `kubernetes.runtime` is defaulted on every platform |
| `TestRuntimeExplicitValueSurvivesDefaults` | A configured override is never overwritten by defaulting |

### execguard_test.go

The execution guard: what a config-declared command may be, and the proof that the
validator and every executor enforce it from one definition.

| Test | What it covers |
| --- | --- |
| `TestGuardConfigIsValid` | Guards the fixture the rest of the file rests on: the untouched `guardConfig` validates cleanly on all three platforms, so a `Validate` failure below can only have come from the guard |
| `TestCheckCommandAccepts` | The shapes an operator legitimately writes: every allowlisted binary bare, a flags-and-values profile, `--flag=value` followed by another flag, a lone `-`, a stripped `.exe`/`.EXE` suffix, a chained runner WITH the escape hatch (`lima podman`, `microk8s kubectl`, `lima --tty=false nerdctl`), and both compose forms including one derived behind a wrapper. Every wrapper here is non-escalating on purpose -- `--allow-command` can never approve sudo and its relatives |
| `TestCheckCommandRejects` | The full reject matrix, each case also asserting the message names the offending token and a way out: an empty or nil command, an unlisted binary (`curl`, `bash`), the right binary on the wrong platform, every path form, a bare word in subcommand position, a bare word after `--flag=value`, `compose` outside the one field and position that permits it, a token after `compose`, a chained runner WITHOUT the hatch, the wrong name in the hatch, an escalation wrapper even when one is forced into the allow-set, the literal `--`, and one case per charset class -- all 21 metacharacters, quotes, backslash, backtick, `$`, control characters, NUL, DEL, seven Unicode space characters beyond ASCII, and four invisible formatting characters (zero-width space and joiner, RTL override, soft hyphen) whose message names the code point since the character cannot be seen -- in argv[0] and in a later token |
| `TestFlagValuePositionIsNotGuaranteed` | Documents the one acknowledged limit as a property rather than a surprise, and how far it actually reaches: `docker --tls rm` and `kubectl --insecure-skip-tls-verify delete` both pass. Neither flag takes a value, so the word after it is not a flag value at all -- it reaches SUBCOMMAND position, the one thing the guard otherwise guarantees. Arity is what cannot be known without a per-binary flag table, so this is accepted rather than closed. The whole charset -- metacharacters, Unicode whitespace, invisible characters -- still applies there, which is the half an adversarial review found incomplete once. Fails deliberately if a future change narrows the limit, so it gets rewritten as a rejection |
| `TestEscalationFloorFoldsCaseBothWays` | The escalation floor is matched case-INSENSITIVELY at both enforcement points, and the allowlist deliberately is not. `--allow-command Sudo` used to be accepted and stored, and `allowed()`'s purge deleted only the lowercase key, so both belts had the same hole -- which bites exactly where filesystems are case-insensitive (Windows, which ships sudo.exe, and macOS by default). The third subtest pins the asymmetry: `KUBECTL` must still be REFUSED by the allowlist, because a deny list may only ever refuse more, while a positive match that folded case would approve a file nobody listed |
| `TestCharsetAgreesAcrossBothYAMLForms` | The regression guard for how that gap arose: a `Command` may be a scalar (split with the Unicode-aware `strings.Fields`) or an explicit sequence (preserved token for token). For ten Unicode space characters, the scalar form must split and the sequence form must refuse -- otherwise two spellings of one config would get two verdicts |
| `TestGuardErrorsAreActionable` | Every guard message is one line, names the field, and names a remedy (the allowlist, the escape hatch, or the specific fix) -- a message that only said "invalid" would leave the operator guessing which of four fields to edit |
| `TestValidatorAndExecutorAgree` | The shared-definition test. Every accept and reject case is driven through BOTH enforcement points -- `Validate`, and the accessor each executor calls before building argv -- and the two must return the same verdict. Fails if anyone ever forks the check |
| `TestExecutorRejectsWithoutValidate` | The reason the check runs twice: a `Config` built in code, which never went through `config.Load`, is still refused by `ClusterCommand`, `RuntimeCommand` and `ComposeCommand` |
| `TestAllowCommandsAccepts` | `--allow-command` is repeatable, each value extends the same set, and `.exe` folds to one entry |
| `TestAllowCommandsRejects` | A bad hatch value is a usage error naming the flag: paths (so the hatch cannot reintroduce the path form layer 2 refuses), metacharacters, whitespace, control characters, empty |
| `TestAllowCommandsRejectsEscalation` | The escape hatch has a floor: `sudo`, `doas`, `su`, `pkexec`, `run0`, `runas`, `gsudo` and their `.exe` spellings can be approved by nobody. Granting one elevates every command the tool issues for the life of an env file, where `sudo solace-util ...` elevates one invocation the operator chose -- so the message must name that alternative, and the name must not be recorded despite the failure |
| `TestEscalationCannotBeAllowedByAnyRoute` | The structural backstop: even with an escalation wrapper forced into the allow-set (a future edit to `execBinaries`, or a caller populating `extraAllowed` directly), `allowed()` strips the category back out -- while a legitimate wrapper in the same forced set still works, proving it is a deny-list and not a broken allow-set |
| `TestAllowedBinaryIsNotGloballyAllowed` | An approval is per-`Config`, so it cannot leak into a second config in the same process |
| `TestComposeCommandDerivation` | `docker.compose` defaults to the runtime's own `compose` subcommand, and `ApplyDefaults` stores exactly what `ComposeCommand` derives -- the two definitions cannot drift |
| `TestComposeDerivationInheritsRejection` | An unlisted runtime cannot become an approved compose command by way of the derivation |
| `TestExecBinariesCoversEveryPlatform` | The allowlist is pinned against `Platforms()`. A missing key is neither a compile error nor an obvious runtime one: `commandRules.allowed` would return an empty set, so every command that platform names would be refused as un-allowlisted and the error would offer an empty "allowed" list -- a confusing refusal rather than a loud wiring failure |
| `TestAllowCommandIsNotASchemaKey` | The structural half of "the config author has no say": every plausible spelling of an allowlist key fails strict decoding, and none reaches the unexported field backing it |

### config_test.go

| Test | What it covers |
| --- | --- |
| `TestPlatformIsContainer` | `Platform.IsContainer` is true for docker/podman, false for k8s |
| `TestPlatformConstantsMatchSchemaSections` | Each `Platform` constant value is also a top-level section key in `Config`, reflected off the yaml tags. The subtree name and `--platform` read the constant, but the tag is a separate literal, so this is the only thing tying the word an operator types to the block they edit |
| `TestRedundancyEnabled` | Only the literal `true` enables HA; `false`, empty, junk, and the old `yes` spelling do not -- the last one matters, since a file carrying it must fail validation rather than quietly deploy a standalone broker |
| `TestImageRef` | `Image.Ref` joins repo:tag, prefixing the registry only when set |
| `TestParseRole` | Long and short role spellings parse, empty defaults to primary, junk errors |
| `TestRoleNames` | Pins the shell-completion suggestion list to the parser: every name `RoleNames` offers parses, no two name the same role, and all three roles are covered. Both now read the same abbrev set, so what is left to catch is a `Role` constant that never reached `roleTable` and is therefore unreachable from the CLI |
| `TestRoleAbbrevIsTheRoleValue` | Why roles need no reverse table: the approved short form IS the `Role` constant, so `p` cannot come to mean anything but `Primary`, and an entry carrying some other letter would hand the CLI a spelling the pod-name suffixes do not use |
| `TestRoleErrorTeachesBothSpellings` | The rejection is rendered from the set rather than typed beside it, and names every letter and every long name |
| `TestRoleLetter` | Role -> `p`/`b`/`m` letter used in resource names |
| `TestResolveNodeStandalone` | Standalone ignores the role and always resolves the primary as a message-routing node |
| `TestResolveNodeHA` | HA resolves each role to its host name, with the monitor typed `monitoring` |
| `TestContainerRuntime` | Runtime command comes from the platform's block, leading args included; k8s has none |
| `TestContainerBlock` | Podman reads its own container block; everything else falls through to docker's |
| `TestNetworkBlock` | Network block is selected per platform |
| `TestApplyDefaultsK8s` | Every k8s default lands: redundancy, update strategy, admin secret, diag dir, CLI folder, storage, operator image/resources, scaling, ports, anti-affinity. Broker resources now come from the scaling tier instead: `msgNode.cpu` stays empty (it is the removal sentinel, not a value), `msgNode.mem` is the tier-100 default and `Scaling.CPU` its cores |
| `TestApplyDefaultsDocker` | Docker defaults (runtime, compose mode, the compose command derived from the runtime, host network, admin user, container name) plus the shared `kubernetes.*` fields containers reuse |
| `TestApplyDefaultsPodmanRootful` | Rootful podman gets the system quadlet dir, no `--user`, `multi-user.target` |
| `TestApplyDefaultsPodmanRootlessXDG` | Rootless quadlet dir derives from `XDG_CONFIG_HOME`, with `--user` and `default.target` |
| `TestApplyDefaultsPodmanRootlessHomeDir` | Empty `XDG_CONFIG_HOME` falls back to the user home dir branch |
| `TestValidateK8sValid` | A fully populated k8s config validates clean |
| `TestValidateK8sMissingMandatory` | Every missing mandatory k8s field is named in one message, exact wording pinned |
| `TestValidateK8sBadUpdateStrategy` | `kubernetes.updateStrategy` enum is rejected loud |
| `TestValidateContainerHA` | A valid HA container config validates for both docker and podman |
| `TestValidateContainerStandalone` | Standalone requires NO node field at all: one node, always this host, and an omitted `redundancy.primary.name` is answered by the host's own hostname. A configured name still wins |
| `TestValidateContainerMissingMandatory` | The BULK missing-fields message: one run names every empty mandatory field (image, admin, all three node name/addr pairs) rather than making the operator fix them one at a time. `redundancy.psk` is deliberately NOT in that list and is set in the fixture so it does not fire -- it gets its own error ahead of this one, because it is the single field nothing else will fill in and it needs the `openssl` command rather than a place in a list (see `TestPSKIsMandatoryOnContainers`) |
| `TestValidateContainerBridge` | `network.mode=bridge` without ports errors; with ports it passes |
| `TestValidateContainerIdentifiers` | Container and node names that reach the compose/quadlet artifact in structural positions are format-checked, so a colon, '=', space or newline is an error instead of a broken artifact. Empty backup/monitor names stay legal in standalone |
| `TestValidateContainerRunUser` | runUser keeps its own `uid[:gid]` pattern -- the default "0:0" carries a colon the identifier check would reject |
| `TestValidateK8sKeyValueEntries` | The "key: value" fragments (loadBalancer.annotations, placement.labels*) must carry a key; a value holding a colon is fine because the renderer quotes both halves |
| `TestValidatePullPolicy` | `kubernetes.imagePullPolicy` enum, including the empty case that keeps the renderer's own IfNotPresent |
| `TestValidateK8sDNSLabels` | The two DNS-1123 shapes Kubernetes really enforces: `kubernetes.namespace`/`name`/`operator.namespace` must be a single LABEL (no dot, 63-char cap), while the three secret-name fields (`adminSecret`/`tlsServerSecret`/`imagePullSecret`) get the looser SUBDOMAIN rule (dots allowed, 253-char cap) that lets `prod.solace-admin-secret` -- accepted by kubectl today -- keep validating; a shared bad-case set (uppercase, leading hyphen, an embedded YAML document separator, colon, space) is rejected by both |
| `TestValidateK8sOptionalSecretNamesStayOptional` | The regression the DNS-1123 hardening could have introduced: the three secret-name fields and `operator.namespace` were never mandatory, so leaving one blank must still validate exactly as before |
| `TestValidatePlacementAffinity` | The additive affinity blocks: unknown operator, missing key, In without values, and a pod term with no topologyKey each fail naming the field; a full valid set passes |
| `TestDefaultK8sPortsMatchesOperator` | The built-in port list is the operator's own 17 entries, including the tcp-ssh entry, with no duplicate names |
| `TestValidateK8sPorts` | `kubernetes.ports` entry shapes `parsePort` accepts (name=container, an explicit service port, an explicit protocol) all validate; the "name:port" typo with no `=`, port 0, port 65536, a non-numeric port, a bad protocol, an over-length port name, a duplicate name, and a duplicate container port each fail naming the offending entry |
| `TestApplyBridgePortDefaults` | Bridge mode with no ports defaults to the k8s set as host:container pairs on both platforms; host mode and an explicit list are untouched |
| `TestImageTagVersion` | Tag parsing behind the health-check gate: dotted versions, a `-rc1` suffix, and a two-part tag parse; `latest`, empty, a bare major and a non-numeric tag report *unknown* rather than guessing; `AtLeast` compares major before minor |
| `TestValidateHealthCheck` | The opt-in probe: with no cmd it uses the built-in readiness endpoint, so 10.26+ is accepted while an older tag and an unidentifiable one are both refused (naming the explicit-cmd escape hatch); an explicit cmd skips the version gate but keeps the exec-boundary check; disabled stays legal on any tag |
| `TestValidateContainerBadNetworkMode` | Unknown network mode is rejected loud |
| `TestValidateDockerComposeCommand` | `docker.compose` gets the same exec-boundary check as the runtimes: an empty argument is rejected |
| `TestValidateUnknownPlatform` | An unrecognised platform fails rather than validating nothing |
| `TestLegacyAdminSectionIsRejectedByName` | The rename's own migration path. An env file still carrying `admin:` would otherwise fail with yaml's bare "field admin not found in type config.Config", which names the problem and nothing else -- so the section is decoded into a map purely to be refused with the new name (`semp:`), the two keys renamed with it (`pass` -> `adminPass`, `passEnv` -> `adminPassEnv`) and the one REMOVED (`user`). A half-done mechanical rename is the likeliest way this arrives |
| `TestValidateBadRedundancy` | The `redundancy.enabled` enum is rejected loud, naming the two values it takes |
| `TestResolveEnvPath` | Env-file lookup over a real temp tree: base dir first, `env/` fallback, base dir shadows `env/`, default name, no extension inference, a path used verbatim with no `env/` retry, a directory is not a match, both candidates named in the not-found error, control characters rejected |
| `TestResolveEnvPathEmptyBaseDir` | An empty base dir means the current directory, for both candidates |
| `TestResolveEnvPathDefaultInBaseDir` | The default name resolves in the base dir before the `env/` fallback is tried |
| `TestLoadSuccess` | A valid file loads, and defaults are applied during `Load` |
| `TestLoadReadError` | A missing file errors with `read env file` |
| `TestLoadParseError` | Malformed YAML errors with `parse env file` |
| `TestLoadRejectsTheOldK8sSection` | A `k8s:` section is not accepted as an alias for `kubernetes:`: it fails as an ordinary unknown key, so a stale env file cannot half-work |
| `TestLoadUnknownField` | Strict decoding turns a typo'd key into a hard error |
| `TestLoadBashEnvFileHint` | A legacy bash env file is reported as not-YAML and points at `solace-util convert` |
| `TestLoadNotYAMLHint` | Any other non-YAML file says the env file must be YAML and names the schema and the converter |
| `TestLoadUnknownFieldHasNoConvertHint` | A valid-YAML file with an unknown key stays a schema error, without the convert hint |
| `TestLoadValidationError` | A file that parses but fails validation surfaces the missing-fields message |
| `TestLoadResolvesSecretRefs` | An env file carrying no secret at all -- `passEnv` plus a per-user `passwordEnv` -- loads into a fully populated config |
| `TestLoadSecretRefErrors` | Every way a reference fails: unset variable, exported-but-empty variable, both keys set, a `${...}` value where a NAME belongs, and the same on a per-user entry. No message echoes a value |
| `TestSecretRefsLeaveLiteralsAlone` | A literal `${VAR}` password stays exactly that: only the dedicated `*Env` key resolves anything |
| `TestAdditionalUserNameFollowsTheBrokerRule` | The broker's OWN grammar, checked on every platform: start with a letter or `_`, 1-32 characters. Both boundaries are pinned (a leading underscore and exactly 32 characters are legal; 33 is not). Checked at load because the broker would otherwise reject it at CREATE time -- on a running deployment, after everything else has been applied |
| `TestAdditionalUserNameIsStricterOnKubernetes` | The one rule that differs by platform. The users reach a Kubernetes broker through `spec.extraEnvVarsSecret`, projected with `envFrom` -- and the kubelet SILENTLY DROPS keys that are not valid environment variable names, so a username carrying `.` or `-` yields a user with no password, or none at all, with nothing in the deploy to say so. The same name stays legal on docker/podman, which mount the password as a file, and the error has to say "DROPS" so the silence is the thing being explained |
| `TestAdditionalUserPasswordCharsAreFreeOnKubernetes` | A restriction this change LIFTED. Kubernetes used to create these users over the broker CLI, so their passwords could not carry the characters the CLI rejects in a quoted value. They go into a Secret now, base64-encoded and never interpolated into a CLI line, so a password that was refused before must load |
| `TestValidateAdditionalUsers` | On k8s and docker alike: a valid entry passes, and missing/invalid/duplicate usernames, the built-in `admin`/`monitor` names, a missing or invalid access level, and an empty password all fail |
| `TestAdditionalUserNamesCollideOnDocker` | Two usernames differing only in `.`/`_`/`-` fold to ONE docker host variable name, which would feed one user's password to both. Docker-specific for a reason: on Kubernetes the stricter username rule rejects a `-` or `.` outright, so the pair can never be formed there. It used to live in the shared table and stopped meaning anything on the k8s half |
| `TestValidateAdditionalUserClashesWithABuiltIn` | `admin` and `monitor` are the broker's own accounts, with their own keys (`semp.adminPass`, `semp.monitorPass`), so listing one under additionalUsers would produce two secrets feeding a single broker setting. It used to test a name matching a CONFIGURED `admin.user`; that key is gone, so the clash is against a fixed pair -- and it also pins that a name like `operator`, which only collided because the key could be set to it, is now an ordinary username |
| `TestValidateCredentialControlChars` | `semp.adminPass`/`semp.monitorPass`/`redundancy.psk`/`tls.certPassphrase` each reject a control character (naming the field, never echoing the value), driven off the shared `controlCharCases` table |
| `TestValidateAdditionalUserControlChars` | The same control-character rule for `semp.additionalUsers[].password`. A control character in a credential is a malformed value on every platform -- it reaches a mounted file on containers and a Secret data value on Kubernetes -- so this one is not platform-scoped |

### platform_test.go

The platform vocabulary and the env-file probe behind it. These matter more than their
size suggests: the platform is never typed as a subcommand, so `DetectPlatforms` is what
decides which system every later step talks to.

| Test | What it covers |
| --- | --- |
| `TestPlatforms` | The enumerator's contents and order -- the order is what a prompt lists and what an error names, so it is contract, not an accident of the literal |
| `TestParsePlatform` | Every accepted spelling (canonical plus `kube`/`dk`/`pm`), empty meaning "not specified", and the rejections -- `k8s` and `k8` are both listed explicitly among them, so reinstating either silently would fail |
| `TestParsePlatformErrorTeachesCanonical` | A rejection names the canonical words the env file's own sections use, not only the abbreviations -- and is the set's own `List()` rendering, so the sentence cannot outlive the table it describes the way the hand-written one beside the map could |
| `TestEveryPlatformHasExactlyOneShortForm` | The set is built by walking `Platforms()`, so a platform added there arrives with no short form rather than missing from `--platform`'s vocabulary; this is what says the entry is still incomplete |
| `TestDetectPlatforms` | The whole resolution input space: none, one, several sections; `docker: {}` and a null section both count as declared; results come back in `Platforms()` order regardless of the order in the file |
| `TestDetectPlatformsMissingFile` | A bad path fails here rather than reaching `Load` with an empty platform list |
| `TestDetectPlatformsBashFileHint` | The legacy-bash-env mistake keeps its `solace-util convert` hint. `DetectPlatforms` is now the first thing to read the file, so without this the message would degrade to a bare decode error |

### scaling_test.go

The scaling-tier table: `scaling.maxConnections` fixes the broker's CPU on every
platform and defaults its memory, so these cover the table itself, the derivation,
and the two keys the change removed or added.

| Test | What it covers |
| --- | --- |
| `TestScalingTiers` | All five tiers resolve to the published cores and memory, and the case count is asserted against the table so a tier added to one and not the other fails |
| `TestTierForRejectsOffTierValues` | The deliberate absence of rounding: values between, below and above the tiers (including 0 and a negative) resolve to nothing rather than to a neighbour |
| `TestScalingTierListMatchesTable` | The error message's tier list cannot drift from the table -- every listed value is a key, the list is ascending, and its rendering is exact. The package avoids `sort`, so the order is a literal that needs pinning |
| `TestContainerMem` | The one rewrite between the schema's two memory spellings: Kubernetes' `Mi`/`Gi` to the bare `m`/`g` docker and podman accept, leaving an already-container value untouched. Every tier's rewritten default is checked against the validator it would face from an env file, so a default cannot be one the loader rejects |
| `TestApplyScalingTierDefaultsK8s` | A non-default tier derives its cores into `Scaling.CPU` and its memory into `kubernetes.msgNode.mem`, while `msgNode.cpu` stays empty so `validateK8s` can read any value there as user-set |
| `TestApplyScalingTierDefaultsMemOverride` | The asymmetry the change rests on: an explicit memory survives defaulting on both k8s and container, while CPU is the tier's regardless |
| `TestApplyScalingTierDefaultsContainerBlocks` | Both container blocks are filled whichever container platform is active, matching `applyContainerDefaults`' existing parity |
| `TestApplyScalingTierDefaultsOffTier` | The fail-safe: an unresolvable tier derives nothing rather than inventing a footprint, and `Validate` is what the operator hears from |
| `TestValidateScalingTierRejectsOffTier` | An off-tier value fails on all three platforms with a message listing the five tiers -- the check sits ahead of the platform switch because every platform now renders a CPU limit from it |
| `TestValidateScalingTierAcceptsEveryTier` | Every tier validates cleanly on every platform, so the enum cannot be narrower than the table |
| `TestValidateK8sMsgNodeCPURemoved` | `kubernetes.msgNode.cpu` still decodes but fails validation, so the operator gets a reason naming `scaling.maxConnections` and noting `mem` is unaffected, rather than a bare unknown-field error |
| `TestValidateMaxPoolRemoved` | `maxPool` would name the same broker setting as `maxSpoolUsageMB` under a platform-specific name; it is rejected on all three platforms naming the replacement, and an unset (zero) value does not trip the sentinel |
| `TestValidateContainerMem` | `container.mem` takes docker's and podman's own `b\|k\|m\|g` suffix: the likely mistake (a `Mi` quantity copied from `kubernetes.msgNode.mem`) is refused naming that trap, alongside bare numbers, decimals and unknown suffixes, while every legal form and the unset case pass |

### duration_test.go

| Test | What it covers |
| --- | --- |
| `TestCanonicalDuration` | The round-trip that keeps a free-form string out of an argv: what comes back is produced by `time.Duration.String()`, so it is provably digits plus unit letters and no consumer downstream has to police it |

### hostpath_test.go

Every file-valued key an env file carries becomes a host path this process reads, joins
or hands to a container engine. These cover the charset those paths are held to, the
cross-platform answers (`IsAbsHostPath`, `BaseName`) that must not depend on which OS
loaded the file, the rebasing of relative paths against the env file's own directory,
and the keys deliberately exempt from it.

| Test | What it covers |
| --- | --- |
| `TestHostPathCharsetIsDerivedFromTheTokenCharset` | Why `hostPathUnsafe` is built by `strings.Map` rather than typed out: two hand-written charsets drift, and the drift is invisible until an operator hits the one character in one list and not the other. It asserts the relationship, not the contents -- widening `unsafeTokenChars` widens the path set automatically |
| `TestCheckHostPathAccepts` | The shapes an operator legitimately writes, including the two Windows forms a command token may not carry |
| `TestCheckHostPathRejects` | Each refusal with its own reason, so a message that stops naming the cause fails here rather than confusing an operator |
| `TestLeadingTildeIsRefusedButAnEmbeddedOneIsNot` | The carve-out pinned on its own, because the two cases differ only by the tilde's position and the reasons are unrelated: nothing here expands a home directory, while an 8.3 short name is a real path a Windows runner hands us for its own temp directory |
| `TestIsAbsHostPath` | The question asked for both operating systems at once, which is the point -- `filepath.IsAbs` alone answers only for the machine running the test, so one env file would be judged differently depending on where it was loaded |
| `TestBaseNameSplitsOnBothSeparators` | The cross-platform agreement test. The value becomes an in-broker filename AND a Secret data key, so two answers for one env-file name two different objects rather than differing cosmetically |
| `TestHasPathSeparator` | The name-or-path question `broker perform cli-script` and the domain-certificate filenames both rest on |
| `TestValidateHostPathsChecksEveryPlatformsFields` | The gate fires for every field it covers, on every platform. It uses a character the path charset refuses (`$`) so a failure can only have come from `CheckHostPath`, and sets one field at a time so a message naming the wrong field is caught rather than masked by a neighbour |
| `TestValidateHostPathsMatchesTheRebaseList` | The drift test between the two halves of the host-path story: a rebased field must also be gated. The reverse is allowed and deliberate -- `podman.quadletDir` is gated but never rebased, because the unit has to live where systemd scans |
| `TestRebaseResolvesAgainstTheEnvFileDirectory` | The join itself: relative values move, absolute ones do not, and an empty value stays empty rather than becoming the base directory |
| `TestRebaseIsANoOpWithoutABaseDir` | The property `internal/convert` and every hand-built `Config` depend on -- `ApplyDefaults` and `Validate` are called directly, with no path to derive a base from, and those paths behave exactly as before |
| `TestQuadletDirAndDataDirAreNotRebased` | The two deliberate exclusions, each for its own reason: the unit must live where systemd scans, and the data dir is what a recursive delete points at |
| `TestPodmanBaseDirIsRequiredAndAbsolute` | Both halves of the one mandatory path key. Mandatory rather than defaulted because it receives the server-certificate bundle, which contains a PRIVATE KEY; required absolute rather than rebased because it is a quadlet `Volume=` source, and podman reads a relative source as a NAMED VOLUME -- mounting an empty volume over the certificate with no error at all |
| `TestDataDirMustBeAbsolute` | The one host path REQUIRED absolute rather than resolved, with a message explaining both halves of why: it is a bind-mount source (podman reads a bare one as a named volume) and it is what `broker remove --delete-data` deletes recursively |
| `TestDomainCertsFolderIsDefaulted` | A promise the sample env file made that the code did not keep: an omitted folder made `filepath.Join("", file)` collapse to a bare filename resolved against whatever directory the command ran from |
| `TestContainerCertRequiresKey` | A deliberate tightening: `tls.cert` without `tls.certKey` now fails to load on docker and podman. Until the server certificate got a real delivery path the container renderers mounted `tls.cert` ALONE and never read the key, so broker-owned TLS could only have worked if that one file already carried it. Failing at load names the missing field; the old behaviour produced a running broker whose TLS quietly did not work. Kubernetes is unaffected -- the operator assembles both halves from a Secret |
| `TestKubernetesNameIsBoundedByItsDerivedNames` | A name that passes on its own and still produces an object the cluster refuses: `kubernetes.name` is a DNS-1123 label, so 63 characters is legal for the name itself, but the operator suffixes it into every object it creates and Kubernetes copies a pod's name into the `statefulset.kubernetes.io/pod-name` label, whose values stop at 63 |
| `TestContainerNameMustBeAnEngineName` | The half of `container.name`'s grammar the charset check did not cover: the FIRST character must be alphanumeric. `-broker` passes the charset and both engines refuse it -- but not before this tool has written it into a compose file and a quadlet unit, and not before it has reached argument positions where a leading dash reads as a flag |

### storage_test.go

A broker node gets its data volume one of two ways: provisioned by the operator from a
StorageClass, or mounted from a PersistentVolumeClaim that already exists. These pin the
rules that keep those two stories from being told at once, and the reason each errs the
way it does. `storageCfg(class, redundancy, mounts)` is the shared fixture.

| Test | What it covers |
| --- | --- |
| `TestStorageClassAndCustomMountAreMutuallyExclusive` | The headline rule. Both fields describe where the data volume comes from, the CRD does not say which wins, and a tool that picked for you would be guessing about persistent data. The error names both keys and the offending class |
| `TestCustomMountMustCoverEveryNode` | A half-covered HA group is far likelier to be a mistyped key than a deliberate mix, and the failure mode is silent -- the uncovered node provisions from the cluster default and nothing says so. The error names the MISSING role rather than saying the set is wrong |
| `TestCustomMountIgnoresRolesOutsideTheGroup` | A `backup` entry in a standalone file names a node that does not exist. Harmless, and exactly what a file switched from HA to standalone looks like -- refusing it would make the switch a two-step edit for no gain |
| `TestCustomMountRejectsAnUnknownRoleKey` | The dangerous typo. A key that is not a role (`primry`, `p`, `Primary`) would leave the node it was meant for on default provisioning, quietly, on a broker whose storage someone deliberately took control of -- so it is refused, listing the roles it does accept |
| `TestCustomMountRejectsAnEmptyClaim` | A key with no value is a half-finished edit, and rendering it would emit a `claimName` the CRD requires and Kubernetes would reject |
| `TestMsgNodeSizeIsOptionalOnlyWithCustomMounts` | A size is what the operator asks a StorageClass to provision. With every node custom-mounted the volumes already exist, so demanding one would be demanding a number that changes nothing -- but without them it is still the one storage field that cannot be defaulted |
| `TestRoleWordAndTitleAgree` | The two spellings of one concept pinned against each other: the env file keys on the lowercase role word, the CRD constrains `customVolumeMount[].name` to a capitalised enum. They are derived rather than typed twice, and this is what says so |

### tlssecret_test.go

A TLS Secret can come from either side: this tool builds it from files the env file names,
or the operator created it and the env file only says which one the broker should use.
`kubernetes.tlsServerSecret` names it in both cases -- what tells them apart is whether
`tls.cert`/`tls.certKey` are supplied. `tlsCfg(secret, cert, key)` is the shared fixture.

| Test | What it covers |
| --- | --- |
| `TestNamingTheTLSSecretDoesNotInventCertPaths` | The regression the split exists for. `applyK8sDefaults` used to fill `tls.cert`/`tls.certKey` with `certs/tls.crt`/`certs/tls.key` whenever the Secret was named, so "the Secret already exists" turned into a read of a file the operator had never mentioned -- and `broker generate` failed on a path that appears nowhere in their env file |
| `TestSuppliedCertsMakeTheSecretOurs` | The other half: files present means this tool builds it, which is what `broker generate` renders and `broker remove` cleans up |
| `TestTLSCertAndKeyMustBeSetTogetherOnKubernetes` | The pairing was enforced only on docker and podman while defaulting filled both fields at once, so one could not arrive alone. Without the defaulting it can, and it would build a Secret carrying a certificate and no key -- which the operator mounts and the broker cannot start a listener over. The error names the way out as well as the fault |
| `TestSuppliedCertsNeedASecretName` | Files with no name give the Secret nowhere to be created and the CR no `tls` block, so the certificate would be silently unused -- the quietest failure mode here, and the one worth a loud error |
| `TestNoTLSAtAllStaysValid` | TLS is opt-in on every platform, and the two checks above must not have made the plainest deployment fail |

### replication_test.go

| Test | What it covers |
| --- | --- |
| `TestValidateReplicationValid` | The complete DR-pair fixture passes, so every rejection below fails for the reason it names rather than because the baseline was broken |
| `TestValidateReplicationAbsent` | The whole `replication:` section is optional: most deployments have no DR pair, and an omitted block must not make every other command fail to load |
| `TestValidateReplicationRejects` | Every way the block can be wrong, each message naming the field: not exactly 2 sites, a missing or duplicated `virtualRouterName`, missing/empty/overlapping `routerNames`, no endpoints, an empty host, a port outside 1-65535, `transport: encrypted` (refused by name, pointing at `ssl`), an unknown transport, three endpoints on one transport, a `via` declaring both mechanisms, each missing `via.kubernetes` field, `via.semp` without a host or with a bad port, zero or two credential sources, an incomplete `passSecret`, a nameless or duplicated VPN, and an `activeAt` naming nothing or naming a ROUTER name instead of a virtual-router name |
| `TestValidateReplicationEmptyViaAccepted` | A site with no `via` still loads. The block says how to reach a site when it is the MATE and only the switch command needs it, so requiring it at load would refuse a file that configures replication perfectly well for the local-only command |
| `TestReplicationLocate` | A broker finds itself by ANY of its site's router names -- the HA backup node reports its own name, not the primary's -- and both non-happy paths are errors: no match prints what was read beside every declared name, and an overlap refuses rather than resolving to whichever site came first |
| `TestSiteCommandGuarded` | A replication site's cluster CLI is re-checked when the command is fetched, not only at load: a binary off the kubernetes allowlist and one carrying a metacharacter are both refused, `Validate` names `replication.sites[0].via.kubernetes.command` rather than `kubernetes.runtime`, a site reached over SEMP says so, and an unknown `virtualRouterName` is refused |
| `TestReplicationPassEnvResolves` | The mate's admin password resolves through the one `secretRefs` list. Without an entry there a `passEnv` would stay an unexpanded variable NAME and be sent to the mate as the password |
| `TestReplicationCredentialChars` | The mate password gets the same charset check as every other credential -- it reaches a curl config on stdin where a newline breaks out of its line -- and the error never echoes the value |

### redundancy_test.go

The pre-shared key is the one credential whose requirement differs by platform, and the
asymmetry is the operator's rather than a preference. `pskCfg(p, psk, pskEnv)` builds a
valid HA config with the key under the caller's control.

| Test | What it covers |
| --- | --- |
| `TestPSKIsMandatoryOnContainers` | An HA group with no key is refused on docker and podman -- and the error carries the WHOLE instruction, asserted piece by piece: the field, `openssl rand -base64 32`, the same-value-on-all-three-hosts rule, the platform refusing, and that Kubernetes does not need it. This tool used to generate the key and rewrite the env file; with that gone the error is the only place an operator is told how, and "must not be empty" would leave them guessing a format for a credential that has to match on two other machines |
| `TestPSKWhitespaceIsNotAKey` | A psk of blanks is what a half-finished edit leaves behind, and it would satisfy a bare non-empty check while giving the group a key that cannot match anything typed on another host |
| `TestNothingGeneratesThePSK` | The deletion's own guard: the error must not promise generation. The generator was unreachable once the key became mandatory at load -- load rejected the file before the deploy that would have filled it in could run -- so it was removed rather than left as dead code the docs kept promising. If "we will make one for you" ever comes back, the contradiction surfaces here |
| `TestPSKEnvSatisfiesTheContainerRequirement` | The reference form is a configured key: `Load` resolves `pskEnv` into `PSK` before validation, but a config that never went through `Load` carries only the variable name, and "a key was configured" is what the rule asks |
| `TestPSKIsNotRequiredStandalone` | With no group there is nothing to authenticate to |
| `TestFillStandaloneNodeNameUsesTheHost` | An empty standalone `redundancy.primary.name` is filled from the host: one node, always this host, so the host's own name is the answer rather than a guess |
| `TestFillStandaloneNodeNameKeepsAConfiguredName` | A configured routername wins over the hostname |
| `TestFillStandaloneNodeNameNeverFillsHA` | The guard that matters: an HA name is never defaulted. Each one keys that node's entry in the group table EVERY host renders, and a host knows its own name and no other machine's -- filling one in would have three hosts render three different tables and the group would never form, so the field is left empty for validation to refuse |
| `TestFillStandaloneNodeNameIgnoresAnEmptyHostname` | `os.Hostname` can return `""`; writing that in would swap a missing routername for a blank one |
| `TestPSKIsOptionalOnKubernetes` | The other half: the operator invents and distributes a key when the CR names no Secret, so refusing an empty one would demand a value that changes nothing -- which is why the requirement lives in `validateContainer` rather than the shared set. An explicit key is equally valid, and is then honoured rather than ignored |

---

## internal/cli

Command-tree wiring, global flags, confirm prompts, and end-to-end passes over the
sample env, plus command aliases, platform resolution, the generated command and
abbreviation references, shell completion, and the end-to-end behaviour of the
execution guard.

`echoRunner`, a test-only helper (`cli_test.go`) that installs `engine.Echo` via the
`App.NewRunner` seam (`internal/cli/context.go`), is how a test asserts the exact argv a
command would issue. Reach for `runRootWith(t, args, echoRunner)` (or the
`runStandalone`/`runCtr` wrappers built on it) for that.

`exportconfig_test.go` adds its own doubles for `broker perform export-config`/
`import-config`: `exportconfigTransportOutput` cans `opRunner`'s (`cli_test.go`) output
hook to answer the three probes the CLI-over-exec channel makes (the login banner, which
carries the broker type; `show message-vpn *`; and the broker-generated teardown script)
without a live broker;
`writeExportconfigArtifact` and the `exportconfigNoVPNArtifact`/`exportconfigOneVPNArtifact`
fixtures stand in for a captured `show current-config` transcript; and
`exportconfigReadCounter` wraps `App.PromptIn` to prove a confirmation prompt was never
actually read, not merely that the command did not block.
188 tests across 11 files.

Because the platform is a flag rather than the first word of a command, the
invocations here name it explicitly (`--platform docker`) rather than relying on
detection: `env/sample.yaml` declares all three sections, so an invocation without it
would hit the ambiguity error instead of the behaviour under test. Detection itself is
covered in `platform_test.go`, against fixtures written for that purpose.

### cli_test.go

| Test | What it covers |
| --- | --- |
| `TestEnvFileLookup` | `-e`/`--env` as the CLI wires it: `env/` fallback, base dir shadowing, no extension inference, the `==> env file:` echo, and long/short flag parity |
| `TestEmit` | `emit` writes bytes to stdout unchanged |
| `TestWarnAndStep` | `warn` and `step` write `[WARN]` / `==>` lines to stderr, now by delegating to the shared `internal/output` Sink rather than formatting the prefixes themselves |
| `TestAnnounceCommandsNamesResolvedBinaries` | The preamble that replaced the per-call `exec:` line: each binary the env file names is resolved and printed once as `==> using <name>: <path>`. k8s announces the cluster CLI; docker announces one line when `compose` is the runtime's own subcommand and two when a standalone `docker-compose` is configured; a name that resolves nowhere is skipped in silence, since a report must not invent a failure the first real execution already reports. Hermetic -- a stub binary is written into a temp dir put at the front of `PATH`, so the expected path is exact and no test host needs kubectl or docker |
| `TestBinaryAnnouncementWiring` | The same through the real command tree, with the stub named `kubectl` so it is the schema default and needs no `--allow-command` (which is itself refused where nothing executes). A real run (`validate`) announces before it works; the render-only `broker generate` and `operator generate` announce nothing, since neither one ever runs a binary |
| `TestVerboseFlagTracesEveryCommand` | `-v` prints `==> exec: <path> <args>` per call, a run without it prints none, and pairing `-v` with the test-only echo seam still works rather than fighting it, since Echo already echoes every command |
| `TestTreeStructure` | A representative set of command paths exists in the one unified tree, covering a top-level leaf, a group's child, and each applicability class |
| `TestEveryRunnableCommandIsWired` | The wiring that replaced the two `PersistentPreRunE` hooks: every command with a `RunE` (bar `convert`, `examples`, `version`, `completion`) carries the shared pre-run and the `--allow-command` flag. Missing either is invisible until that one command is run, so it is checked structurally |
| `TestGroupCommandsPrintHelpAndDoNothing` | The no-implicit-actions rule: a verb that owns objects (`check`, `smoke`, `prepare`, `deploy`, `config`, `start`, `stop`, `restart`, `status`, `logs`, `copy`, `generate`, `remove`, and `config`'s own `apply`/`delete`/`disable`) carries no `RunE`, so running it bare prints its own help and touches nothing -- proven by succeeding with no `--env` at all, since a runnable leaf would instead fail resolving the missing default env file |
| `TestFlagsRegistered` | Per-command flags are registered where expected: `--restart` and `--pod` on `broker deploy`; `--delete-data`/`--no-prompt` on `broker remove`; `--delete-crd`/`--no-prompt` on `operator remove`; `--days` on `broker perform gather-diagnostics`; `--pod` on `broker cli`, `broker perform cli-script` and `broker perform shell-script`; `--pod` on `broker copy from`; `--pod`/`--dir` on `broker copy into`; `--all`/`--detail`/`--pod` on `broker status`; `--detail` on `operator status`; `--out`/`--no-prompt` on `examples`, `broker generate`, `operator generate` and `convert` -- every command whose output is an artifact you keep takes the same pair; `--apply`/`--remove`/`--no-prompt` on `broker configure server-certs` and `broker configure product-keys`, the two `configure` leaves with an outage on the other side of their removal (TLS down, and an unlicensed broker) and therefore the two that ask |
| `TestHelpNoConfig` | `--help` short-circuits before config load, so no env file is needed |
| `TestGenerateWired` | `generate`'s leaves, none of which contact the cluster or the container engine: `broker generate` renders what `broker deploy` would apply on EITHER family (the CR's `apiVersion:` on kubernetes, `services:`/`[Unit]` on docker/podman), plus `broker generate` on both and `operator generate` on kubernetes. A plain `runRoot` with no echo seam is enough since nothing here executes |
| `TestCtrWiredDryRun` | Every container command safe to run against the HA sample drives clean over the echo seam, echoing the expected runtime/systemctl/mkdir command |
| `TestCtrRoleGuards` | Primary-only HA guards: leader and redundancy both reject backup and monitor hosts, bad roles error, and standalone self-skips -- all before any poll or SEMP call |
| `TestCtrConfigDryRun` | Container config steps run clean on a standalone env; cert/product-key-gated steps self-skip |
| `TestCtrExecCLIPathSeparator` | `opCtrExecCLI`'s used-as-is branch: a `cli --input` file argument containing a path separator is used as-is, not joined under the CLI scripts folder |
| `TestRemoveServerCertsOverTheCLI` | The direction that was a loud placeholder until `no ssl server-certificate` was confirmed. Three properties, each one a reason it was held back: `--no-prompt` runs the confirmed form; an interactive `n` changes nothing; and a run with no terminal and no `--no-prompt` keeps the certificate rather than failing. It asks at all because removing the certificate a broker is presenting takes TLS down immediately |
| `TestRemoveProductKeysOverTheCLI` | The last direction that was a placeholder. `--no-prompt` revokes every configured key; an interactive `n` issues nothing; and an env file with no `broker.productKeys` is refused rather than reported as done. It asks at all because revoking every key can leave the broker UNLICENSED -- an outage whose cause points nowhere near the command |
| `TestRemoveServerCertsRefusedOnASecretManagedDeployment` | With `kubernetes.tlsServerSecret` set the operator mounts the certificate and would reconcile it straight back, so a CLI removal would report success over a broker that still presents it. The refusal names the Secret, the key to clear and the reconcile that would undo it -- "not supported here" would leave an operator with no next move |
| `TestConfigureServerCertsRefusesASecretItDoesNotOwn` | The three-way routing. Keying the Secret branch on `ManagesTLSSecret` sent a bring-your-own deployment down the pod-exec CLI path, where it died on a generic "tls.cert and tls.certKey must both be set" -- pointing the operator at fields that, if they set them, would make this tool overwrite a cert-manager Secret. The named Secret is what selects the Secret route; whether we can rebuild it is `UpdateServerCertSecret`'s question, and its tailored refusal was unreachable until this |
| `TestCtrErrorPaths` | Container `config apply` failures are actionable: no TLS configured for server-cert, no product keys configured, and a failed `broker perform semp-login-check` |
| `TestCtrDiagnosticsDryRun` | Container `diagnostics` echoes its node-local gather/download sequence over the echo seam (isolated because it creates a diag dir) |
| `TestCtrRoleArgCount` | Role-taking commands reject a second positional argument |
| `TestCtrRoleHelp` | Role-taking commands expose `--help` without loading an env |
| `TestK8sWiredDryRun` | Every k8s command safe to run against the HA sample drives clean over the echo seam, with `+ kubectl` echoed on the acting paths and absent on the skip paths |
| `TestK8sStandaloneDryRun` | Redundancy-branching commands on a standalone env: the HA-only `broker perform assert-leader`/`broker perform redundancy-test` self-skip, while `broker deploy` runs clean |
| `TestDeployBrokerDoesNotApplyOperator` | `broker deploy` never installs the operator: it is cluster-scoped and shared, so one env file reaching for it would silently re-apply a controller other namespaces depend on |
| `TestDeployDoesNotAssertLeader` | The inverse of a property `deploy all` had. Asserting the config-sync leader reverts activity on the mate, which changes redundancy state, and a deploy has no business making that change unasked -- it is `broker perform assert-leader` now |
| `TestDeployNeverLabelsNodes` | A deploy touches no node. It used to be true because `deploy all` deliberately left the interactive picker out; it is now true because nothing in the tool labels nodes at all. The terminal is made interactive on purpose here, since that is what the old bug looked like |
| `TestCheckDeployWarnsWhenOperatorAbsent` | `opK8sCheck`'s operator probe: `validate` is read-only, so a missing operator is reported as a stderr warning rather than failing the check itself |
| `TestStartStopRestartBroker` | The day-2 `start`/`stop`/`broker restart` verbs on both platform families: Kubernetes scales the statefulset(s) via a kubectl echo, containers start/stop/restart the container in place via a docker echo |
| `TestCLICommand` | `cli` has two shapes distinguished by a flag: bare, it opens an interactive session; with `--input`, it uploads and runs a script instead. Two H2 regression subtests pin `--pod` steering EITHER shape onto the named pod: `--input --pod backup` must not fall back to the primary (the old bug, where `[role]` and `--pod` were two disconnected branches), and bare `--pod backup` opens the session against that pod |
| `TestStatusBrokerFlags` | How `--all` and `--detail` compose on `broker status`: they widen the report along independent axes (every broker in the cluster vs. this env file's one; the static description vs. the running inspection) rather than one replacing the other, on both kubernetes and a container platform |
| `TestRemoveBrokerLayerContract` | The retained-layer contract on `broker remove`: persistent data is kept by default, `--no-prompt` proceeds while still keeping it, `--delete-data` deletes it, and a non-interactive run keeps it. Losing data always takes an explicit `--delete-data`, whatever else is on the command line |
| `TestRemoveOperatorLayerContract` | Mirrors `TestRemoveBrokerLayerContract` for the operator's CRDs: kept by default, since deleting them cascades to every broker in the cluster, and deleted only when `--delete-crd` names them |
| `TestRemoveFlagsCompose` | `--delete-data` and `--no-prompt` answer DIFFERENT questions -- what to do with the data, and whether to ask about anything -- so they must combine, not conflict. A fully unattended removal that also drops the data needs both |
| `TestSecretsNeverEchoed` | `broker perform semp-login-check` over the echo seam shows stdin as a byte count and never the admin password |
| `TestConfirmLayerMatrix` | The whole removal-confirmation contract, which INVERTED with the tree. The `--delete-*` flag no longer answers the question, it RAISES it: without it the layer is kept and nothing is asked; with it, an interactive run needs an exact `yes`, `--no-prompt` answers yes, and a non-interactive run without `--no-prompt` keeps the layer and says so loudly. All four cells, because the dangerous one is only dangerous in combination and a matrix with a hole in it is how the wrong cell ships |
| `TestConfirmDeleteShortcut` | The half that did NOT change: `--no-prompt` confirms the removal itself, the question every destructive command asks before it does anything |
| `TestConfirmNonTTY` | Without a TTY and without `--no-prompt`, `confirmDelete` refuses, and `confirmLayer` with no flags set keeps the layer -- both with no prompt read |
| `TestConfirmDowngradeNeverUnattended` | M9: `confirmDowngrade`'s own gate -- non-interactive declines and tells the operator to re-run interactively (never advising `--no-prompt`, which this command does not register), `noPrompt` is not a side door (still declines non-interactively even when set), an interactive `yes` proceeds and an interactive `no` refuses |
| `TestDeployOperatorNoPromptStaysUnknownFlag` | The other half of M9: fixing `confirmDowngrade`'s dead `a.noPrompt` branch must not resurrect `--no-prompt` on `operator deploy` -- the flag was never registered there, so it must still fail as cobra's own unknown-flag error |
| `TestPromptYesNo` | Lenient delete prompt: `y`/`yes` in any case accept, everything else declines |
| `TestPromptYes` | Strict layer-deletion prompt: only an exact trimmed `yes` accepts; a bare `y` does not |
| `TestErrorPaths` | Global rejections: unresolvable env file, an invalid role via `--pod` on both container and k8s leaves (`broker deploy`, `broker generate`, `broker logs`), and an unknown verb under a noun |
| `TestK8sGenSecretsWired` | The Secret manifests render through `broker generate`, on the standalone env rather than the HA sample whose kubernetes.tlsServerSecret points at cert files absent from a checkout |
| `TestGenNeverLeaksSecrets` | End-to-end: `broker generate` output on both container platforms omits the admin password, while `broker generate` carries it (it is what creates the secret) |
| `TestGenerateWritesTheArtifactItself` | The `-o` flag's whole reason for existing. `broker generate > x.yaml` is not portable: Windows PowerShell 5.1 re-encodes this tool's plain ASCII as UTF-16LE with a BOM and kubectl then rejects the file, blaming the YAML rather than the shell. The corruption happens after the process exits, so writing the file from inside is the only fix -- and the property that matters is byte-for-byte identity between what `-o` writes and what stdout carries, asserted alongside no-BOM and no-NUL |
| `TestGenerateToFileConfirmsAnOverwrite` | The artifact is a file you keep, so replacing one is confirmed like every other destructive step. A test run has no terminal, which is the non-interactive branch: it must keep the file and name the flag that would have proceeded |
| `TestGenerateNoPromptWithoutOutIsHarmless` | `--no-prompt` answers the overwrite question, and with no `-o` there is no file and no question. Accepted rather than refused -- a script passing it unconditionally is not making a mistake worth an exit 2, and nothing is silently skipped |
| `TestGenerateEmitsSecretsBeforeTheCR` | End to end through the command: the Namespace, then a Secret, then the CR. The ordering half of the generate consolidation on Kubernetes. Two commands used to print these and the operator applied them in the right order by hand; one command prints all three now, so the order is this tool's responsibility -- the Namespace before the namespaced objects, and a Secret before the CR that names it, or applying the stream to an empty cluster leaves the broker referencing something that does not exist yet |
| `TestOperatorGenerateIsTheWholeInstallStream` | `operator generate` emits namespace -> regcred -> bundle, and the ORDER is what is pinned rather than the contents: the Secret is namespaced and its namespace only exists inside the bundle, so a stream that put it first would not apply to an empty cluster. With no pull secret configured there is simply no Secret document -- a deployable stream, not an error |
| `TestConfigStepsDoNotLeakSecrets` | Drives each `config apply`/`config disable` step individually against a container fixture carrying every optional value (server cert/key, a domain CA, a product key) and asserts none prints the private key material to stdout -- `config` does not aggregate these steps into a re-runnable ordering, so each needs its own direct coverage |
| `TestConvertToStdout` | `convert` writes YAML to stdout and its warnings to stderr, so the artifact stays clean |
| `TestConvertToFile` | `-o` writes the file, a second run refuses to clobber it, and `--no-prompt` answers the overwrite question. The converted file carries the source's secrets verbatim, which is why replacing one is worth confirming |
| `TestConvertRoundTrip` | A converted file loads: `-e` against it drives a real command |
| `TestConvertErrorPaths` | Bad `--platform`, a missing source file, and a missing argument all fail loud |
| `TestVersionPrintsStampedValue` | `version` reports whatever the dev scripts' `-X` flag (or a test) set the package var to, verbatim -- the contract that aligns a release binary with its git tag |
| `TestVersionDefaultsToDev` | An unstamped build (plain `go build .` or `go test`) reports "dev" |
| `TestVersionIncludesToolchainAndPlatform` | Output carries `runtime.Version()` and GOOS/GOARCH, for support triage |
| `TestVersionRejectsArgs` | `version` takes no arguments |
| `TestBashEnvGivenToEnvFlag` | Pointing `-e` at a legacy bash file reports not-valid-YAML and names `solace-util convert` |
| `TestExecute` | `Execute()` builds the tree and runs `--help` without error |
| `TestK8sConfirmDeclined` | Every removal declines unattended: `broker remove`/`all`/`secrets`/`namespace`/`operator` without `--no-prompt` make zero cluster calls, using the App.Interactive seam instead of ambient stdin. secrets and namespace confirm like every other removal, and `remove namespace` takes everything living in the namespace, not only what this env file put there |
| `TestK8sPromptsNameNamespaceAndContext` | Table-driven over `broker remove`/`secrets`/`namespace`/`all`/`operator` and `broker restart`, with `a.kubeContext` set directly (the same seam as `App.Interactive`): every prompt now names the namespace and repeats the `(context <name>)` clause, except `operator remove`, which names the OPERATOR's own namespace (`k8s.OperatorNamespace`) instead of the broker's |
| `TestK8sPromptsOmitAnUnknownContext` | The other half: with `kubeContext` left empty, as an unresolved `announceKubeContext` lookup leaves it, `broker remove`'s prompt still names the namespace but drops the context clause entirely rather than rendering an empty `(context )` |
| `TestK8sRestartRollsOrTargetsOnePod` | H2 regression guard: `opK8sRestart` must keep reading `app.pod` RAW rather than through `podRole`, whose empty-defaults-to-primary would silently turn "restart every pod" into "restart the primary pod alone" -- no `--pod` rolls monitor/backup/primary in order, `--pod backup` touches only that one pod |
| `TestK8sRestartConfirmGate` | a non-interactive `broker restart` (rolling, or one `--pod` role) bounces nothing, and a bad `--pod` role is rejected before any prompt |
| `TestCtrRestartConfirmGate` | `opCtrRestartBroker`'s new `confirmDelete` gate, the container-side counterpart of `TestK8sRestartConfirmGate`: a non-interactive run bounces nothing and `--no-prompt` issues the restart, on both docker and podman -- `broker restart` used to run unconditionally on containers |
| `TestCtrConfirmDeclined` | a non-interactive `broker remove` on docker without `--no-prompt` issues zero runtime calls |
| `TestStdinCanAnswerClosedFile` | a stdin that cannot be Stat'd (a closed file) is a stream no reply can arrive on, so `interactive` is false and the question is never asked |
| `TestPromptAsksWhenStdinIsNotATTY` | regression for a live failure -- `export-config -o <existing>` refused to overwrite WITHOUT asking. A pipe on stdin carrying `y` is now as good as a console: the question is asked and the answer honoured. Deliberately opens no terminal; the first attempt at this fix read `/dev/tty`/`CONIN$` and hung |
| `TestUnansweredPromptStillRefuses` | a readable stdin that ends immediately: the question IS asked, then goes unanswered, and the refusal names `--no-prompt`. Also pins `promptSource(nil)` |
| `TestPromptSeamsBeatStdin` | an App carrying `Interactive`/`PromptIn` never touches `os.Stdin`, which is what keeps the ~200 wiring tests independent of what the suite was launched from |
| `TestPromptKeepsBufferedAnswersForLaterQuestions` | A command asking TWO questions in one run (`broker remove --delete-data`) gets both answers. `promptLine` built a fresh `bufio.Reader` per call and discarded it, so a pipe delivering `y\nyes\n` in one read answered the first question and threw the second answer away -- the layer prompt then reported itself unanswered and kept the data the operator had just authorised deleting |
| `TestCtrManagerConfirmWiring` | ctrManager wires Manager.Confirm to confirmRestart, and a non-interactive session (via App.Interactive) declines without reading a prompt |
| `TestK8sLoginOutcomes` | a transport failure propagates as an error and a canned 200 OK response returns nil, the two real SEMP outcomes engine.Echo's fixed (nil,nil) can never produce |
| `TestCtrLoginOutcomes` | same as TestK8sLoginOutcomes for the container login path |
| `TestOpK8sDeployAborts` | `opK8sDeploy`'s error-return arms (CreateNamespace, CreateSecrets, DeployBroker), each failed in its own sub-test with no later step's command issued afterward. Two arms the `deploy all` version had are gone: no Check step, since a deploy runs the per-operation preflight rather than the whole validate report, and no operator-apply step, since the operator has its own noun |
| `TestOpK8sRemoveAborts` | `opK8sRemoveBroker`'s two error-return arms (DeleteBroker, DeleteSecrets): a failed broker- or secrets-deletion stops before the namespace is even inspected. That question is only meaningful once everything this env file owns is gone -- getting it wrong would offer to cascade-delete a namespace still holding a broker whose deletion had failed |
| `TestOpCtrVerifyRedundancyRunsCoordinated` | When this host is the primary, `opCtrVerifyRedundancy` actually calls `RedundancyCoordinated` instead of only ever hitting the skip/reject arms `TestCtrRoleGuards` already covers -- driven over a fake Runner seeded with an active-but-unhealthy `show redundancy` transcript so the health check fails immediately, before any poll loop or mate SEMP call |
| `TestContainerRoleDetectsFromHostname` | `containerRole` (how `broker deploy` and `broker generate` resolve which node THIS host is) over the injected `App.Hostname` seam. An omitted `--pod` in HA is detected from the hostname against `redundancy.*` and announced on stderr; an explicit `--pod` still wins; an unrecognised hostname fails loud naming `primary\|backup\|monitor` as the way out (closing the old silent-second-primary bug, since `config.ParseRole("")` alone always returned Primary); standalone keeps defaulting to primary with no detection announced. An explicit `--pod` is also CHECKED against what the host looks like: disagreement WARNS and proceeds (never prompts -- a three-host scripted deploy must not stall), agreement and an undetectable host both stay silent, since undetectable is the very case `--pod` exists for |
| `TestStandaloneRouternameFallsBackToTheHost` | The load-time fill (`App.fillStandaloneNodeName`) end to end: a standalone container env file naming no routername renders an artifact whose `hostname` AND `routername` are the HOST's name, announced on stderr. Through the real command, because the point of filling at load is that every reader agrees -- the artifact here, the check report and the DNS check elsewhere, all off one value |
| `TestConfiguredRouternameSurvivesTheFallback` | The other half: a configured routername is never overwritten by the host's name |
| `TestK8sSmokeRedundancyUnhealthy` | `broker perform redundancy-test` fails on its first check rather than polling: over the echo seam, `engine.Echo`'s empty `show redundancy` output makes `primaryRedundancyUp` false, so `opK8sVerifyRedundancy` returns the redundancy-unhealthy error before any SEMP login |
| `TestK8sConfigDeleteDomainCertsConfigured` | With a CA actually configured, `broker configure domain-certs --remove` issues a kubectl exec instead of self-skipping -- every other test's `domainCerts.files` map is empty, so `domainCANames`'s map-to-slice conversion was correct only by vacuity until this one configures a CA |
| `TestConvertParseError` | a malformed legacy env file (unterminated array assignment) surfaces the parser's own error through the CLI |
| `TestConvertWriteError` | an -o path whose parent directory is absent fails with a wrapped write error naming the path |
| `TestK8sGenSecretsMissingCertFile` | kubernetes.tlsServerSecret configured with an unreadable tls.cert fails loud naming the read failure through `broker generate`, instead of only being caught at real deploy time |
| `TestRemoveDomainCertsOverTheCLI` | `domain-certs --remove` now confirms like its `server-certs`/`product-keys` siblings: `--no-prompt` runs the removal, an interactive `n` issues nothing, no terminal and no `--no-prompt` keeps the certificates, and `--apply` (the enable direction) still runs unconfirmed |
| `TestDisableDefaultVPNConfirmGate` | `default-vpn`'s newly confirmed disable direction: it stops every client connection using the VPN, so it takes the same four-way gate (`--no-prompt` shuts it down, decline changes nothing, no terminal keeps it up, `--enable` runs unconfirmed) |
| `TestDisableDefaultUsersConfirmGate` | `default-users`' matching disable gate. `DisableDefaultUsers` reads the broker's VPN list before it can build the disable script, so the echo seam's canned empty reply means the script itself never runs either way -- the property distinguishing a confirmed run from a declined one is the upload-and-run of the `show-vpn` probe, which only a confirmed (or `--enable`) run reaches |

### aliases_test.go

Command abbreviations (`internal/cli/aliases.go`): an [abbrev](../internal/abbrev) set
keyed by a command's NAME, applied by a tree walk, panicking at construction on a
sibling collision. What these tests protect is not the individual words but the two
properties that make them safe to hand out -- an alias resolves to exactly the command
its long form does, and no alias can ever be ambiguous with a sibling -- plus the
deliberate absences (`start`/`stop` get none, and no verb that owns objects gets a
bare-alias shortcut to acting). The charset and shorter-than rules are the set's own
and are covered in `internal/abbrev`.

| Test | What it covers |
| --- | --- |
| `TestAliasesResolveToTheCanonicalCommand` | Pins equivalence by resolution rather than by running anything: cobra's own `Find` -- what dispatch itself uses -- resolves both an alias path (e.g. `rm br`) and its canonical form (`broker remove`) to the same `*cobra.Command`. Covers every verb/noun alias plus the standalone ones (`diag`, `cv`, `ver`, `sh`) |
| `TestAliasesDoNotCollide` | Walks the real tree and proves no two siblings answer to the same word (name or alias). `applyAliases` already panics on a collision at construction, so this is the second line of defence -- the one that would also catch a collision introduced by a command's own hand-written `Aliases` rather than by the table |
| `TestEveryAliasEntryIsLive` | Catches the quiet failure mode of a name-keyed table: every canonical name in `commandAbbrev` must be a command actually present in the tree, so a renamed command's alias cannot silently stop applying while the set still claims to provide it |
| `TestNounGroupsRunNothing` | The safety property behind offering short forms at all, inverted with the tree. It used to be the VERB that had to run nothing, because `rm` was reachable at the top level; the tree is noun-first now, so `rm` only exists under a noun and the NOUN is what must be inert. `broker`, `operator`, `broker copy`, `broker configure` and `broker perform` all carry subcommands and no `RunE`, so `br` and `op` act on nothing. A `RunE` added to one later fails here |
| `TestMousetrapIsDisabled` | Cobra's Windows double-click guard stays off. Behaviourally it must: the hook prints "this is a command line application" and DECLINES to run when it decides the binary came from Explorer, which is not a failure mode a tool driven from terminals, scripts and CI should have. It is also the largest single cost in this package's suite -- the check walks the OS process table on every `Execute`, measured at 37% of the whole run, 10.6s of 13.8s, against 1.8s for the commands under test. Nothing else would notice it returning: every test would still pass, only slower |
| `TestGroupsRejectAnUnknownVerb` | Why a group is runnable at all. Cobra answers a NON-runnable command by printing help and exiting 0 whatever arguments it got, so `broker remvoe` would report success having done nothing and a script would never notice. The mistyped word is the VERB now the tree is noun-first. Bare still prints help and succeeds; an unknown verb fails loudly |
| `TestStartStopHaveNoAlias` | Pins the one deliberate omission: `start` and `stop` get no two-letter form at all, in the set and on the actual tree commands, because any short spelling would be ambiguous with the other (and with `status`) at the exact place where guessing wrong costs an outage. Both must be DECLARED with no short form and with a note saying why -- an omission and a decision read the same otherwise, and the note is what a later reader finds when they wonder |

### allowcommand_test.go

The execution guard end to end -- flag parsing, `config.Load`, `Validate`, and the
executors -- driven through the real command tree.

| Test | What it covers |
| --- | --- |
| `TestAllowCommandIsRegisteredWhereItExecutes` | `--allow-command` is declared on every command that runs something -- the unit the operator is approving a binary for -- and is still NOT a root flag, so it stays a usage error on `solace-util convert`, which loads no config and runs no platform CLI |
| `TestAllowCommandIsRepeatable` | Both values of a repeated flag are collected in order, so a chain needing two approvals does not force a choice |
| `TestAllowCommandApprovesAWrappedRuntime` | The accept case end to end: `microk8s kubectl` is refused with a message naming the hatch, and runs -- reaching the echoed command -- once the operator passes `--allow-command microk8s` |
| `TestAllowCommandRejectsBadValues` | A path, a metacharacter, an empty value, or any privilege-escalation wrapper is a usage error, so the hatch cannot reintroduce what the guard refuses |
| `TestEscalationIsRefusedEndToEnd` | An env file naming `sudo kubectl` is refused with and without the flag, and the refusal names the supported alternative -- elevate the tool itself, at the moment you run it |
| `TestAllowCommandRejectedWhereNothingExecutes` | The flag is refused on every leaf under `generate` (broker, operator, secrets, artifact), so it is never learned as harmless boilerplate that later gets pasted into a run that does execute |
| `TestHostileRuntimeIsRefusedByEveryVerb` | An unlisted binary stops `check`, `status`, `deploy`, `remove`, `status --all` and `logs` alike -- `status` running `curl` is the same arbitrary execution `deploy` running it would be |
| `TestSmuggledSubcommandIsRefused` | `kubectl delete`, `kubectl delete ns prod` and a literal `--` in the config are all refused: this tool appends its own subcommand, and a word there would run ahead of it |
| `TestPathRuntimeIsRefused` | The bare-name rule end to end for relative, absolute, and parent-directory forms -- the `./kubectl` shipped beside the env file |
| `TestGenPathNeverExecutes` | Backs the trust-model promise that rendering an untrusted env file is safe: `broker generate` issues no external command at all |

### platform_test.go

Platform resolution and the one-tree consequences of it. This is where the CLI's
central decision is pinned: the platform is inferred from the env file (or `--platform`)
rather than typed as the first word of a command, so what it resolves to, and what happens
when a command does not apply to it, are the behaviours most worth holding still.

Fixtures: `writePlatformEnv(t, platforms...)` writes a minimal but VALID env declaring
exactly the given sections (the kubernetes and container schemas need different
mandatory fields, so it emits both shapes); `runPlatform` runs `broker status` over
the echo seam, the cheapest command that still goes all the way through resolution
and config loading.
The prompt branches are driven through `App.Interactive`/`App.PromptIn`, the same seams
the destructive-confirmation tests use.

| Test | What it covers |
| --- | --- |
| `TestResolvesSinglePlatformSilently` | The everyday case on all three platforms: one declared section is used with nothing said on the command line, proven by which binary the echo seam names |
| `TestNoPlatformSectionIsRefused` | The marker requirement. A container env file needs no `docker:` keys at all -- they all default -- so without the section it would be indistinguishable from a kubernetes one; the error names all three sections |
| `TestMultiPlatformNonInteractiveIsRefused` | With no terminal to ask, guessing is the one thing that must not happen: the error names the declared platforms and points at `--platform` |
| `TestMultiPlatformPromptSelects` | The interactive branch: a numbered answer selects that platform |
| `TestMultiPlatformPromptRejectsBadAnswer` | Empty, out-of-range, zero, and a platform name typed instead of a number all stop the run rather than falling through to a default |
| `TestPlatformFlagSilencesThePrompt` | `--platform` is what makes an ambiguous file usable from a script |
| `TestPlatformFlagAcceptsAbbreviations` | `kube`/`dk`/`pm` survive the whole resolution path, not just the parser |
| `TestPlatformFlagRejectsUndeclaredSection` | Naming a platform the file does not describe is refused, naming both what was asked for and what the file declares -- the alternative is deploying from defaults nobody wrote down |
| `TestPlatformFlagRejectsUnknownValue` | Keeps `k8s` and `k8` out, along with other near-misses (`swarm`) |
| `TestUnsupportedCommandFailsLoud` | The other half of the one-tree decision: since the tree shows every command everywhere, the refusal is what tells an operator a command does not apply -- and it names where it does apply, so the message is not a dead end. Covers both directions (kubernetes-only commands on docker, container-only on kubernetes) |
| `TestScopedFlagFailsLoud` | `--restart`, `--all` and `--pod` are refused where they mean nothing rather than accepted and ignored -- a `--restart` that did nothing would read as "already restarted" |
| `TestUnusableRoleFailsLoud` | A role that was TYPED must never be silently dropped. No command takes a `[role]` positional any more, so every case is the one migration error (`noRolePositional`), on both platforms -- `broker logs backup`, `broker cli backup`, `broker shell monitor`, `broker status backup`, `broker restart backup`, `broker deploy backup`. The two refusal kinds collapsed into one when `rejectRole` went with the positional; the platform half of the guard is now `--pod` being scoped per command, which `TestScopedFlagFailsLoud` covers |
| `TestRolePositionalTeachesPodFlag` | The likelier half of the H2 migration: an operator on KUBERNETES typing the OLD documented spelling (`shell backup`, `broker logs monitor`, `cli primary`, `broker status backup`, `broker restart backup`, `broker perform semp-login-check backup`) gets an error naming both `--pod` and the role typed, rather than cobra's bare "unknown command" hiding the fact that it merely moved. A `shell typo` subtest pins the other side: a word that is not a role keeps cobra's own wording and never offers the `--pod` hint, since a typo is not a migration |
| `TestPlatformIsAnnouncedInThePreamble` | The platform is inferred rather than typed by the operator, so it is stated in the preamble -- otherwise the one fact the operator does not type themselves would also be the one they cannot see |
| `TestCompletionNeverReadsTheEnvFile` | The invariant that decided where the pre-run hook lives. Cobra runs the NEAREST ancestor's `PersistentPreRunE` and `__complete` is root's own child, so a hook on root would parse an untrusted env file on every TAB press; keeping it per-command prevents that, proven by completing with an env file that does not exist |
| `TestPlatformFlagIsOnRoot` | `--platform` is a root persistent flag inherited by every command including `convert`, which is what lets one word mean one thing across the whole CLI |
| `TestScopedCommandsSaySoInHelp` | The tree is one static shape, so help text is the only place to learn a command does not apply before running it; a command that applies everywhere carries no scope tail |
| `TestPlatformOpsCoversEveryPlatform` | The builder pinned against `Platforms()`. Every other consumer of an ops map already walks `Platforms()`, so `platformOps` is the one place a fourth platform would be dropped silently: with both halves non-nil it would still return three entries, `supported()` would omit the new name, `onlyOn` would tag the command for three platforms, and the new one would refuse every command that has a perfectly good implementation -- and the refusal would look deliberate |
| `TestLogArgsBuildsOneSetForBothPlatforms` | The shared log-flag builder. The two argv builders must not disagree about what a flag means, which is why the tokens are produced once rather than assembled per platform |
| `TestPlatformAnnotationsMatchDispatch` | Every applicability annotation in the tree parses, so a malformed one cannot silently widen or empty a command's platform set |

### commanddoc_test.go

| Test | What it covers |
| --- | --- |
| `TestCommandDocs` | Renders the command reference from the live tree and fails while `docs/commands.md` is stale -- the drift gate for every command path, positional, flag, `Short`/`Long` string, alias, and platform/render-only annotation. This file is also the generator: `-update` rewrites the doc |

### abbrevdoc_test.go

The abbreviation glossary. The short forms the tool accepts come from four places that
never meet at run time -- the command set, the role set, the platform set and pflag's
shorthands -- so a hand-written glossary would go stale at whichever of them changed
first. This renders it from all four. The `-update` flag is `commanddoc_test.go`'s, so
one `regen` rewrites both goldens.

| Test | What it covers |
| --- | --- |
| `TestAbbreviationDocs` | Renders `docs/abbreviation.md` from the three abbrev sets plus the live tree's flag shorthands, and fails while the committed file is stale. Also the generator: `-update` rewrites the doc. It renders the verbs that have NO entry too, so a new command shows up as unabbreviated-and-unexplained rather than silently absent |
| `TestFlagShorthandsAreConsistent` | The one rule `abbrev` cannot enforce, because pflag owns shorthand registration: pflag rejects a duplicate within one flag set, but nothing stops `-o` meaning `--out` on one command and `--output` on another. A letter that means two things is the failure the whole scheme exists to avoid, so it is checked across the whole tree |

### completion_test.go

Shell completion end to end. The value tests drive cobra's hidden `__complete`
endpoint through the real tree -- the same request a loaded completion script makes
on every TAB press -- via the `runComplete` helper, which cannot reuse `runRoot`
because that discards cobra's own writer. Nothing here loads an env file: the
per-command `PreRunE` never runs for `__complete`, which is what keeps a TAB press
from parsing config or executing anything. (`TestCompletionNeverReadsTheEnvFile` in
`platform_test.go` pins that invariant directly, by completing against an env file
that does not exist.)

| Test | What it covers |
| --- | --- |
| `TestCompletionScriptsGenerate` | Each of bash/zsh/fish/powershell emits its own script, matched on the line that actually binds the completer to `solace-util`, so a script that generated but wired up nothing still fails |
| `TestCompletionNoDescriptions` | `--no-descriptions` is honoured on every shell: the generated script requests `__completeNoDesc` instead of `__complete`, and does not without the flag |
| `TestCompletionNeedsAShell` | An unsupported shell, or none at all, fails loud with nothing on stdout -- the reason the parent carries a `RunE`, since cobra answers a non-runnable command by printing help to stdout and exiting 0, which would put help text into `solace-util completion tcsh > solace-util.ps1` and call it a success |
| `TestCompletionHelpStillWorks` | `--help` short-circuits ahead of that `RunE`, so asking how to use the command is not itself an error |
| `TestEnvFlagCompletesEnvFiles` | `-e` is completed from the two directories `config.ResolveEnvPath` searches, by bare name: base dir first, the shadowed `env/` copy of the same name offered once, and a non-YAML file not suggested |
| `TestEnvFlagPrefixFilters` | A partial name narrows the suggestions instead of returning every env file |
| `TestEnvFlagWithPathDefersToShell` | A value carrying a directory resolves verbatim, so completion returns the default directive and hands back to the shell rather than offering bare names that would resolve elsewhere |
| `TestPodFlagCompletesRoles` | The role completer, now the only one: `--pod` offers `primary`/`backup`/`monitor` and not filenames, table-driven over every command that takes it -- `copy from`/`copy into` (which always did), plus the six the flag moved onto: `cli`, `shell`, `broker logs`, `broker perform semp-login-check`, `broker status`, `broker restart`. `TestRoleArgsComplete` is gone with the positional it completed: one flag, one completer, registered next to where the flag is declared |
| `TestPlatformFlagCompletes` | The root `--platform` flag, and `convert` inheriting the same one, offer exactly the three canonical platform names. The empty detect value is left out (omitting the flag is how you ask for it); so are `kube`/`dk`/`pm` -- the abbreviations save typing something you already know, which is what a completion does anyway, and offering both would put two names for one platform in front of the user -- and so are `k8s` and `k8` |
| `TestDirFlagCompletesDirectories` | `--dir` asks the shell to filter to directories |
| `TestNoArgsLeafOffersNoFiles` | A command built by `leaf` offers nothing, stopping cobra's filename fallback across the tree. The file-path commands (`broker copy from`/`into`, `broker perform cli-script`/`shell-script`, `convert`) are excluded, since a path is exactly what they take |
| `TestAllowCommandOffersNoFiles` | `--allow-command` offers no files: the value is a bare binary name, and paths are what its own help text warns against |
| `TestFlagCompletionsRegistered` | The drift gate: every flag that should have a completion function still has one, since a renamed flag silently reverts to filename completion at a TAB press and no other test would notice |

### examples_test.go

`examples` is the one command that must work before an env file exists, so these run it
with no `--env` at all (and once with a deliberately missing one). It writes files, so the
`--out` cases all target `t.TempDir()`.

| Test | What it covers |
| --- | --- |
| `TestExamplesBareEmitsTheFullSchema` | With no `--platform`, the complete annotated schema goes to STDOUT. There is no name list to print any more -- the templates are selected by `--platform`, which completion already enumerates, so a listing would be a second vocabulary for the same three words |
| `TestExamplesEmitsToStdout` | The default sink for each name, emitting the package's own bytes and nothing else -- the artifact is this command's only stdout, exactly like `generate` |
| `TestExamplesAcceptsAPlatformAbbreviation` | The CLI-wide rule that an abbreviation works wherever the word does. It is the same `--platform` every other command reads, but it resolves differently here and that is the point: everywhere else it must name a section the env file declares, while `examples` has no env file to check against and expands the word with `config.ParsePlatform` alone |
| `TestExamplesWritesOutFile` | The `--out` branch: the file lands on disk and the confirmation goes to stderr, so stdout stays empty and pipe-safe |
| `TestExamplesRefusesToOverwrite` | The destructive-path guard: an existing `--out` path is CONFIRMED rather than requiring a second flag. `--force` is gone -- the question is asked the way every other destructive question in this tree is, `--no-prompt` is the one silencer, and a run with no terminal keeps the file. It also asserts `--force` is now an UNKNOWN flag rather than a silent no-op |
| `TestExamplesRejectsAnUnknownPlatform` | The error boundary: an unknown word is refused naming every alternative and nothing is written. `k8s` is the interesting value -- the spelling people reach for, deliberately accepted nowhere, so the refusal has to list what is |
| `TestExamplesNeedsNoEnvFile` | Why this command is a bare cobra literal rather than a `leaf`: it emits an env file, so it must work before one exists. A `wireExec` added here later fails this instead of failing a first-time user |

### exportconfig_test.go

Fixtures and doubles for `broker perform export-config`/`import-config`
(`exportconfig.go`, plus the wiring in `commands.go`/`helpers.go`/`ops_k8s.go`/
`ops_container.go`). Both commands talk to the broker over the CLI-over-exec channel
(`broker.Ops`, `internal/broker/importops.go`): upload a script, run `cli -Apes`, read
stdout. `exportconfigTransportOutput` cans `opRunner`'s (`cli_test.go`) output hook to
answer the three probes that channel makes -- the CLI login banner (target broker type,
read by `bannerType`), `show message-vpn *` (target's VPN list), and the broker-generated
teardown script -- so
`ImportPlan`/`ImportApply` run against real business logic with no live broker; anything
uncanned falls through to `(nil, nil)`, the same empty answer `engine.Echo` gives, which is
what makes an uncanned step fail exactly the way an unanswered real probe would.
`writeExportconfigArtifact` and the `exportconfigNoVPNArtifact`/`exportconfigOneVPNArtifact`
fixtures (the latter carrying `exportconfigFixtureSecret`, a fake credential-shaped value)
stand in for a captured `show current-config` transcript, and `exportconfigReadCounter`
wraps `App.PromptIn` to prove a confirmation prompt was never actually READ, not merely that
the command did not block.

| Test | What it covers |
| --- | --- |
| `TestExportConfigScopeConflictIsUsageError` | `--vpn`/`--broker-only` together is a usage error end to end through the command, and the flag conflict never reaches the transport -- S5: bad input rejected AND the downstream call not made |
| `TestExportConfigLeafRunsOnAllPlatforms` | The leaf reaches its real per-platform handler on kubernetes, docker and podman: under `engine.Echo` the capture is necessarily empty, so all three fail identically on `runCapture`'s own "no output" refusal, which is itself the proof the wiring is not platform-specific |
| `TestImportConfigLeafRunsOnAllPlatforms` | `import-config`'s counterpart, failing the same way for the same reason: reading a well-formed artifact succeeds, but the first thing `ImportPlan` asks the TARGET for is its own capture, and under `engine.Echo` that answer is empty -- so the empty-capture refusal comes back identically on all three platforms. It used to be the unknown-target-type refusal, from a dedicated banner probe that ran first; that probe is gone, so the empty capture is caught one refusal earlier |
| `TestExportImportPodFlagPlatformScope` | `--pod` is accepted on Kubernetes (names the pod) and refused as a usage error on docker/podman (one container per host, so an accepted-and-ignored flag would be worse than a refused one), for both leaves |
| `TestExportConfigWithoutOutWritesOnlyArtifactToStdout` | With no `--out` the artifact IS stdout: no report `"==="` section header may appear there (it would corrupt anything piping the output into `import-config` or a file), while narration still reaches stderr unconditionally |
| `TestExportConfigWithOutWritesFile0600AndReportsOnStdout` | With `--out`, the report (Section + KVBlock) goes to stdout instead and the artifact -- as sensitive as the env file -- is written to a file mode 0600 |
| `TestExportConfigOutOverwriteGate` | `--out` over an existing path, all four cases: refused (file left untouched) when there is nobody to ask, naming `--no-prompt`; replaced when the question is ASKED and answered yes; replaced under `--no-prompt`; and the replaced file is still mode 0600 (POSIX only) -- `os.WriteFile` honours its perm only when it CREATES the file, so overwriting an existing 0644 path left the credential-bearing capture world-readable |
| `TestImportConfigMissingFileIsUsageError` | A bad artifact path is the operator's own mistake: reading it fails as a usage error naming the read failure, and nothing is attempted against the broker over a file that was never read |
| `TestImportConfigFailedApplySkipsVerification` | The sequencing rule: a rejected chunk stops the apply, reports, and the verification diff does NOT run -- asserted by the absence of the `.verify-export.cli` call. Diffing after a failure lists every chunk the run never reached as missing, which is true and useless, burying the one line naming the actual failure. It also pins that the broker's own words on the rejection reach the report |
| `TestExportConfigRejectsABadVPNNameAsUsage` | An empty, whitespace or quote-bearing `--vpn` exits 2 (usage), not 1, and reaches the broker not at all. `--vpn "$VPN"` with VPN unset is a typo, and the name was validated only inside `broker.ExportConfig`, whose error came back unwrapped -- so a wrapper retrying on 1 retried a typo forever |
| `TestImportConfigConfirmationMatrix` | `confirmImport`'s whole contract: a lenient `y` proceeds over a NEW VPN but is refused by the strict gate over an EXISTING one (which an import would tear down, destroying every message spooled in its queues); only an exact `yes` proceeds there; `--no-prompt` proceeds either way; and a non-interactive run with no `--no-prompt` warns and touches the broker not at all -- proceeding is judged by whether the teardown/re-capture call was actually attempted, never by whether the whole import finishes cleanly |
| `TestImportConfigBrokerScopeNeverPrompts` | An artifact with no message-VPN at all asks nothing and never even READS the prompt (counted via `exportconfigReadCounter`), because which broker-level sections apply is a fixed classification (docs/import.md), not a per-run decision |
| `TestExportImportSecretsNeverEchoed` | The S3 smoke check for both leaves over the echo seam: export-config's env-carried admin password and import-config's artifact-carried `exportconfigFixtureSecret` never reach stdout or stderr |
| `TestExportScopeFlagCombination` | `exportScope` directly: both flags pass through unchanged when at most one is set, and the conflict is a usage error -- the same pair-reading shape `wantRemove`/`wantEnable` already use |
| `TestPluralVPNWording` | `pluralVPN`'s one piece of variable English in the confirmation sentence: `1 -> "it"`, `2 -> "them"` |

### replication_test.go

| Test | What it covers |
| --- | --- |
| `TestMateChannelPicksTheMechanismTheSiteDeclares` | The dispatch this package owns: `via` holds exactly one child and the key present IS the mechanism, so a kubernetes site is built as a kubernetes channel and a SEMP site as a SEMP one. The fixture's two sites deliberately differ, so a dispatch that always took one branch fails on the other rather than passing on both |
| `TestMateChannelRefusesASiteWithNoVia` | `via` is optional at LOAD and mandatory here: an env file that only ever runs the local-only `configure` must still load, but a switch writes the role at BOTH sites, so an unreachable mate is refused before anything is touched |
| `TestMateSEMPPasswordUsesTheLiteral` | A declared password is used as-is and costs no cluster round trip |
| `TestMateSEMPPasswordIsTheMatesNotThisBrokers` | The misconfiguration the resolution exists to prevent. A DR site is a DIFFERENT broker: sending this deployment's `semp.adminPass` would fail at best and, if the two happened to match, hide the mistake until the day they diverged |
| `TestMateSEMPPasswordReadsTheNamedSecret` | The `passSecret` branch: the value comes back base64-decoded, and the read names the MATE's Secret rather than this deployment's `adminSecret` -- asserted against the configured name, since the two are otherwise easy to confuse |
| `TestMateSEMPPasswordRefusesASiteWithNoPassword` | All three ways to supply one are named in the error. There is deliberately no stdin prompt: `promptLine` echoes, and a no-echo read is a dependency decision this feature does not need to take |
| `TestConfirmReplicationConfigNamesTheBiggerHammer` | The gate states the bigger of its two hazards even though it is asked BEFORE `ConfigureReplication` has read anything: phase 1, when it turns out to run, stops replication on EVERY VPN this broker is replicating, not only the listed ones the per-VPN lines name below the warning |
| `TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply` | The other side of that gate, table-driven: an exact `yes` proceeds, a lenient `y` does NOT (this takes the `--delete-data` bar, because a habitual y should not reach a command whose phase 1 stops replication on every VPN the broker replicates), an explicit `no` refuses, and an unanswerable prompt refuses. Every case also asserts the transport recorded NO call, so a refusal is judged by the broker being untouched rather than by the command merely returning nil |
| `TestReportReplicationConfigMateSkippedSaysNothingWasWritten` | The report never claims the mate was applied when phase 1 was skipped -- `res.Mate` is what WOULD have been sent, not what was, and printing it unconditionally would misdescribe a run that wrote nothing and stopped no VPN |
| `TestReportReplicationConfigMateAppliedNamesWhatPhase1Stopped` | When phase 1 DID run, the report names every VPN it stopped -- the half of the blast radius `res.Enabled` cannot show, since phase 1's shutdown covers unlisted VPNs too |
| `TestResolveSitesRefusesAFileWithNoReplication` | The refusal lands before the broker is asked anything -- with no `replication:` section, a router name matched against an empty list is a round trip whose answer cannot matter |

### exit_test.go

The exit-code contract -- 0 success, 1 operational failure, 2 usage error, and a child's
own status passed through from the interactive commands. `TestHelperExitProcess` (the
shim named in the Summary) is here, because a real `*exec.ExitError` cannot be built by
hand: the test binary re-executes itself and exits with the status the environment names.

| Test | What it covers |
| --- | --- |
| `TestExitCodeContract` | Real invocations through the tree, each asserting the code it produces. Classifying an error handed straight to `ExitCode` would prove only that the switch works; these prove the markers are in the paths an operator reaches. The documented gap is pinned too, deliberately: a mistyped TOP-LEVEL command exits 1 because cobra produces that error before any interception point, so the day someone finds a clean fix, this is the case that says the fix worked |
| `TestEveryArgValidatorIsAUsageError` | The drift guard behind `markUsageArgs`. The `Args` declarations are spread across four files, so this walks the BUILT tree rather than trusting the next one added was remembered. Each validator is offered nothing and then far too much; whichever it refuses must be a usage error, and one that accepts both is reported -- not because the command is wrong, but because the test has stopped saying anything about it |
| `TestRootKeepsCobrasUnknownCommandError` | The reason for the documented gap. Cobra produces its unknown-command error inside `Find` only while `root.Args` is nil; give root a validator and the error is never produced, and a non-runnable root with no error prints help and exits 0 -- so a mistyped command would report success |
| `TestChildExitStatusIsScopedToInteractiveSessions` | The property that matters about the passthrough: one child error carrying one status becomes this tool's exit code from `broker shell` and does NOT from `broker logs`. Without the scoping every command would start reporting whatever kubectl or the engine exited with, which is neither documentable nor scriptable |
| `TestChildExitFallsBackWhenThereIsNoStatus` | A marked error with no exit status behind it (binary not found, child killed by a signal) must not invent one |
| `TestChildExitKeepsItsMessage` | `main()` PRINTS what `Execute` returned before exiting with the code, so a marker that swallowed the message would leave a failed session reporting a bare status and nothing about what went wrong |
| `TestExitCodeIsNeverNegative` | The guard on a signal-killed child -- the one status this process cannot pass through, since `os.Exit` is documented for 0-125 and Go reports -1 for a signalled child. The two operating systems produce different codes for a Kill, so the assertion is the portable property rather than a number: whatever comes back is a status this process can legally exit with |
| `TestUsageErrorKeepsItsMessage` | The classification costs nothing in what the operator reads -- every one of these errors names the flag, word or platform that was wrong |

---

## internal/convert

The legacy bash env -> YAML converter: a shell-assignment parser, the variable
mapping, and the YAML emitter. 37 tests.

### convert_test.go

| Test | What it covers |
| --- | --- |
| `TestConvertLegacyK8sEnv` | `testdata/legacy-k8s.env` converts end to end and matches `testdata/legacy-k8s.yaml.golden`: platform detected as kubernetes and written as a `kubernetes:` section, `true` -> `yes`, every scalar/array/associative value mapped, `${SOLBK_NS}` expanded, a trailing comment stripped, an explicit `0` kept, a multi-word `KUBE` preserved as `kubernetes.runtime` argv, and only the two expected advisories. The fixture sets `SOLBK_MSGNODE_CPU`, as every real legacy file does, so the drop is exercised here: it warns, and no `cpu` reaches the YAML. It also sets `REPL_MATE`/`REPL_CONN_SSL`/`REPL_PSK`, and none of them survive: the `replication:` block now describes BOTH sites of a DR pair, which three variables about one mate cannot populate, so they are dropped with a warning naming all three (B5 itself stays covered by `TestParseSingleQuotedDollarSurvives`, at the parser level where the bug was) |
| `TestParseSingleQuotedDollarSurvives` | The B5 regression. Single quotes are bash's idiom for keeping `$` literal, and the tokenizer used to strip quoting before `expand()` ran -- so `SOLBK_REDUNDANCY_PSK='p$s3cret'` silently converted to `psk: "p"`, truncating the secret with no warning and surfacing much later as an unexplained redundancy failure. The name in the fixture is unassigned, which is exactly the case that vanished into `""` |
| `TestParseMixedQuotingExpandsOnlyTheExpandableHalf` | Why literalness is tracked per SEGMENT rather than per word: in `a'$b'"$B"` the single-quoted half stays literal even though `B` is a real assigned variable, while the double-quoted half still expands. A per-word flag could not express this |
| `TestParseArrayElementsPreserveQuoting` | The array body is a SECOND, separate call site from the scalar path (`tokenizeSegments` + per-word expand in `parse()`), so it gets its own coverage rather than being assumed to follow |
| `TestParseDoubleQuotedBackslashEscapes` | The other route to the same secret loss, which the single-quote fix alone did not close. Inside double quotes bash honours a backslash before exactly `` $ ` " \ `` and leaves it literal before anything else; stripping it unconditionally turned `"\$SECRET"` into an expandable `$SECRET` (truncating it even when the name IS assigned -- one subtest pins that) and quietly ate the separators in `"C:\Users\me"`. An escaped character becomes its own literal segment, so it can never reach the substitution |
| `TestConvertUserPasswordsBecomeAdditionalUsers` | The one legacy variable with no like-for-like successor: `SOLBK_USR_PASS` becomes structured `semp.additionalUsers` entries with the least-privileged `accessLevel: none` plus a warning naming that choice, malformed entries are dropped with a warning naming their POSITION and never their text (a malformed entry is most likely a bare password), and `Convert` re-validating its own output proves the emitted level is a legal one |
| `TestConvertAdminUserIsDroppedOnEveryPlatform` | The one admin field that is not portable: `SOLBK_ADM_USER` is emitted only for docker/podman, and on a k8s target is dropped with a warning naming why (`validateK8s` refuses any non-`admin` value), stays out of the generic unmapped list because it is still read, and leaves a document that validates -- no "will not load as-is". A source that already said `admin` warns about nothing |
| `TestConvertAdminSecretAlias` | `SOLBK_ADM_SECRET` is an accepted alias of `SOLBK_USR_SECRET` for `kubernetes.adminSecret` (hand-maintained env files used it; the repo's bootstraps never defined it): the alias alone maps silently, an agreeing pair maps silently, and a disagreeing pair keeps the canonical `SOLBK_USR_SECRET` with a warning naming the choice |
| `TestConvertK8sSecretNamesAreK8sOnly` | `IMAGEREPO_SECRET` and `SOLBK_SVR_SECRET` name Kubernetes Secret objects, so a container conversion drops each with a warning naming its `kubernetes.*` home (imagePullSecret / tlsServerSecret) -- never the generic unmapped list -- and neither key reaches the container YAML |
| `TestConvertContainer` | A container env file maps the node table, container block, ulimits, network, and spool scaling |
| `TestConvertPlatformDetection` | Podman markers, docker markers, and both-present all resolve to the expected section |
| `TestConvertPodmanSection` | Podman rootless and quadlet dir land in the podman block, and no docker block is written |
| `TestConvertExplicitPlatformWins` | `--platform` overrides detection and suppresses the detection warning |
| `TestConvertUnmappedVariablesWarn` | Variables with no YAML equivalent are named in the warnings, not dropped silently |
| `TestConvertBashPlumbingIsSilent` | Bootstrap-only variables (`EXDIR`, `GENONLY`) are dropped without noise |
| `TestConvertKubeMapsToK8sRuntime` | `KUBE` becomes `kubernetes.runtime` in every shape it carried: a drop-in (`oc`), a wrapper (`microk8s kubectl`), a `--kubeconfig` profile, and an absolute path -- no warning |
| `TestConvertKubeEchoIsDropped` | `KUBE="echo"` was the bash preview trick, so it warns pointing at `generate` and emits no `runtime`, rather than becoming a runtime that no-ops every command |
| `TestConvertKubeSilentOnContainerPlatform` | `KUBE` belongs to the Kubernetes bootstrap: a container conversion consumes it silently and never emits `kubernetes.runtime` |
| `TestConvertRedundancySpellings` | Both legacy spellings normalise onto the schema's own, any case -- `true`/`yes` become `"true"` and `false`/`no` become `"false"`, quoted because both are YAML-ambiguous -- while anything else copies through as a plain scalar with a warning |
| `TestConvertRedundancyOmitted` | An unset SOLBK_REDUNDANCY emits no key either way, but a container source is warned that its bootstrap defaulted to HA while this CLI defaults to standalone; a k8s source stays silent because the defaults already agree |
| `TestConvertDockerRunModeWarns` | DOCKER_MODE=run is dropped with the removal reason rather than carried over to fail validation later |
| `TestConvertBadNumberWarns` | A non-numeric value for a numeric field warns and is not written |
| `TestConvertSpoolVariablesUnify` | Two legacy names for one key: the k8s bootstrap's `SOLBK_SCALING_MAXPOOL` and the container one's `SOLBK_SPOOL_MAXUSAGE` both map to `scaling.maxSpoolUsageMB`, each platform's own name wins when both are set, and the warning says which was used rather than picking in silence |
| `TestConvertOffTierMaxConnWarns` | `SOLBK_SCALING_MAXCONN` was any integer and is now one of five tiers. An off-tier value is still written -- rewriting the operator's declared load would be worse than reporting it -- and `Convert` re-validating its own output is what surfaces it, so this needs no mapping code of its own |
| `TestConvertBadBooleanWarns` | An unparseable boolean (`SOLOP_WATCH_SOLBK_NS`, which the bootstrap never enum-checked) warns and is not written |
| `TestGeneratedHeaderSanitisesSource` | A control character in the source name cannot end the header comment and inject document structure |
| `TestConvertIncompleteEnvWarns` | A source env missing mandatory fields converts, but says the result is incomplete |
| `TestConvertUnterminatedArray` | An array assignment with no closing paren is a hard error |
| `TestConvertInvalidPlatformSection` | An unrecognised platform still writes the shared sections and surfaces the validation warning |
| `TestParseAssignmentForms` | Every assignment form: bare, double/single quoted, empty, trailing comment, `export`/`declare`, inline and multi-line arrays, `declare -A` and bare `[k]=v` maps, `${VAR}`/`$VAR`/unset references, and a function definition that must not parse |
| `TestParseScalarListFallback` | A single-entry list written as a scalar reads as a one-element list; an absent one is nil |
| `TestParseCRLF` | CRLF line endings parse the same as LF, including multi-line arrays |
| `TestParseEscapedQuote` | `\"` inside a double-quoted value survives |
| `TestUnmappedTracksFileOrder` | Unmapped variables are reported in file order, not map order |
| `TestScalarQuoting` | Values a YAML reader could misread (bools, null, numbers, paths, `:`, `#`, quotes, backslashes) are quoted; plain identifiers are not |
| `TestEmptyBlocksOmitted` | Blocks with no content are left out entirely |
| `TestGeneratedHeader` | The output carries the provenance header naming the source file |

---

## internal/broker

Broker CLI operations over an injected transport: script generation, config steps, verify
state machines, the primary-driven container HA variants, the SEMP mate channel, and the
config-export/import feature's block parser, marker layer, section classification, shutdown
injection, verification diff, replay transformations and the generated apply
driver, and the replication mate renderers and readers. 426 tests across 22 files.

`blocks_test.go`, `sections_test.go`, `inject_test.go`, `diff_test.go`, `annotate_test.go`
and `importops_test.go` share one fixture: `testdata/currentconfig_sample.cli`, a
hand-written capture that reproduces the STRUCTURE of a real `show current-config all`
transcript (traced by line number against a real broker's own semp/all.cli, which is never
itself read, copied, or committed since it carries live admin/LDAP/redundancy secrets) with
fake names and no credentials, loaded through `loadSample`/`blockByIndex`
(`blocks_test.go`). `importops_test.go` also reuses `broker_test.go`'s own
`newTestOps`/`fakeTransport`/`matchCLI` doubles rather than a second set.

This package now owns one `-update` golden of its own: `docs/import.md`, generated from
`sections.go`'s classification table by `importdoc_test.go` -- the same
drift-gate-and-generator shape `internal/cli/commanddoc_test.go` established. `regen` now
runs `./internal/broker` alongside the other three packages.

### annotate_test.go

The `! solace-util` marker layer (`annotate.go`): `Annotate`/`StripMarkers` as the round-trip
pair, and the parsing/rendering helpers underneath. The headline property is the round trip --
an export that cannot be re-parsed into the same blocks cannot be imported -- and the rest pin
the design rule annotate.go states: boundaries (the region/section fences) are authoritative,
disposition (the `advisory=` field) is never trusted back from the file. Reuses
`loadSample`/`blockByIndex` from `blocks_test.go` rather than declaring its own fixture.

| Test | What it covers |
| --- | --- |
| `TestAnnotateRoundTripPreservesBlocks` | The headline property: `Annotate` then `ParseBlocks` yields the same block count and the same Kind/Name/VPN/Section/Region per block, over a fixture that exercises every region (vpn-create, broker, vpn) and a section whose own comment embeds a quoted operand |
| `TestAnnotateRoundTripCaptureIsMarked` | The re-parsed `Capture` must itself say it is a marked artifact (`Marked` true, plus the `Meta` fields import reads) -- an importer that cannot tell a marked artifact from a raw capture cannot decide whether to trust the advisory disposition at all |
| `TestAnnotateIsDeterministic` | Two `Annotate` calls with the same input and stamp produce byte-identical output, since verification diffs a freshly captured artifact against a stored one and a non-deterministic renderer would make that diff noise regardless of a real config change |
| `TestAnnotateRegionAndSectionFencesAreBalanced` | Every region and section fence opened also closes, nested correctly, nothing left open at EOF -- checked structurally via `assertFencesBalanced` rather than eyeballing one fixture's tail, since a begin-only scheme would only show up on a file whose last section is trimmed differently than its others |
| `TestWriteBlocksOpensAndClosesOnChange` | `writeBlocks` directly, on a hand-built two-block/two-region input: a region/section change closes the old fences before opening new ones, and the broker's own section comment is re-emitted verbatim alongside the advisory marker rather than replaced by it |
| `TestStripMarkersKeepsPragmaAndSectionComments` | `StripMarkers` removes every `"! solace-util"` line and NOTHING else -- stripping a broker `"! pragma:"` line would detach `ignore-already-exists` from its create line, and stripping a `"! Configure X:"` comment would destroy the classification key sections.go matches against |
| `TestStripMarkersNoMarkersPassthrough` | BUG: a marker-free capture is not returned byte-for-byte unchanged and stripping is not idempotent -- `splitLines` is built on `strings.Split(raw, "\n")`, which yields a trailing `""` element for any input ending in `\n` (every real capture does, the last line is always `"! END\n"`), and the write loop appends `"\n"` after that element too, so `StripMarkers(raw)` is one line longer than `raw`, growing again on a second call -- undermining the function's own purpose, since a verification diff would read that extra blank line as a difference |
| `TestParseKVQuotedValueKeepsSpaces` | The trap named in annotate.go's own comment: a section name is a value here (`name="Configure SEMP Service"`), and a plain whitespace split would truncate it to "Configure" |
| `TestParseKVMultipleTokens` | `parseKV("a=1 b=2")` reads both pairs |
| `TestParseKVEmptyPayload` | An empty payload parses to an empty map |
| `TestParseKVIgnoresTokenWithNoEquals` | A bare word with no `"="` is dropped rather than erroring, so one corrupt field in a meta line does not fail the whole read |
| `TestSplitKVQuoteAware` | The tokenizer `parseKV` depends on: whitespace splits tokens except inside a double-quoted run, and a tab separates tokens exactly like a space |
| `TestMarkerValueSurvivesAnEmbeddedQuote` | A marker value containing a double quote round-trips intact. The writer escapes with `%q` and the splitter used to toggle its in-quote state on the `"` of a `\"`, so a real section name (`Create Redundancy PSK: "#redundancy-group"`) came back truncated and backslash-bearing -- and that is the name the import plan's "omitted at export" row shows the operator |
| `TestIsMarkerDiscriminatesNamespace` | Our own `"! solace-util-*"` lines are recognised, and a broker pragma or section comment -- both also starting with `"!"` -- are not mistaken for one |
| `TestMarkerVerbSplitsVerbAndPayload` | `markerVerb` splits a meta line into its verb and payload |
| `TestMarkerVerbNoPayload` | A marker with no `":"` at all leaves payload empty via `strings.Cut` rather than panicking or misreading the verb |
| `TestMarkerRegionBeginAndEnd` | The region fence reader: a begin returns the named region, an end returns the EMPTY region so termination is explicit rather than implied by whatever begin comes next, and a non-region marker returns `ok=false` |
| `TestMarkerSectionBeginAndEnd` | The section fence reader: a begin returns the quoted name (ignoring the `advisory=` field), an end returns empty |
| `TestReadMarkerExportSetsMarked` | An export marker sets `Marked` and records `Meta[export-version]` |
| `TestReadMarkerMetaAppliesEveryField` | A `solace-util-meta` line applies every field it carries (`broker-type`, `version`, `redacted`, plus an arbitrary key straight into `Meta`) |
| `TestReadMarkerUnknownVerbIgnored` | A newer build's marker verb is ignored rather than refused (forward compatibility, per readMarker's own doc comment) -- it sets neither `Marked` nor touches `Meta` |
| `TestApplyMetaFieldBrokerTypeSetsType` | `broker-type=software` sets `Type` |
| `TestApplyMetaFieldSempSchemaFillsOnlyWhenUnknown` | The broker's own header is parsed first and must win: `semp-schema` fills `Type` only when it started Unknown, and never overrides an already-known `Type`, while `SEMPSchema` itself is always recorded either way |
| `TestApplyMetaFieldVersion` | `version` sets `Version` |
| `TestApplyMetaFieldRedactedCaseInsensitive` | Any casing of `"true"` sets `Redacted`; anything else, including empty, does not |
| `TestSectionBeginMarkerAdvisoryDisposition` | The `advisory=` field an operator reads a section fence for: the classified disposition (with its `interrupts=` detail when it has one), and `"unclassified"` for a section sections.go has no rule for -- never a silent omission of the field |
| `TestRenderKVSortsKeys` | `renderKV` sorts its keys, so two exports of the same configuration are byte-identical rather than reflecting map iteration order |
| `TestRenderKVSkipsEmptyValues` | A field with no value (e.g. an empty scope) is omitted rather than rendered as `key=""` |

### blocks_test.go

`ParseBlocks` and its helpers against `testdata/currentconfig_sample.cli` (see the package
preamble above). `loadSample`/`blockByIndex` build the shared `*Capture` and block lookup
every other config-import test file in this package reuses rather than re-parsing the
fixture itself.

| Test | What it covers |
| --- | --- |
| `TestParseBlocksHeaderAndPreamble` | The header fields (`Version`, `SEMPSchema`, `Type`, `Redacted`) and the trap in a real capture: its first line is `" ! BEGIN"` WITH A LEADING SPACE (semp/all.cli:1), which must not stop the header from parsing; `Preamble` runs to and includes the bare `configure` line that ends the mode preamble |
| `TestParseBlocksPreambleHoldsNoOrphanPragma` | `Preamble` is exactly the header plus `home`/`enable`/`configure`, nothing more -- the regression is a pragma sitting between a "! Section:" comment and its opener (semp/all.cli:43-45, 282-284): the comment has already closed the previous block, so with none open the pragma used to be appended to `Preamble` and relocated ahead of every block, detaching `ignore-already-exists` from the one `create` line it makes replayable |
| `TestBrokerTypeFromSchema` | Every arm of `brokerTypeFromSchema`, since it gates cross-type import refusal: `broker/`/`soltr/..VMR` read as software, `app/`/bare `soltr/` as appliance, anything else (no slash, an unrecognised prefix, empty) as unknown |
| `TestNewBlockOpener` | The opener grammar `[create] <kind> ["<name>"] [message-vpn "<vpn>"] ...` over the shapes that already parse correctly: a bare message-vpn, `create message-vpn`, a VPN-scoped named object, a bare one-liner, and double-spaced fields |
| `TestNewBlockMessageSpoolQualifierHasNoOwnName` | `message-spool message-vpn "default"` (semp/all.cli:32743, also 571/592/905/926) is a per-VPN singleton with no name of its own -- BUG: `firstQuoted` cannot tell a `message-vpn "X"` scope qualifier from the block's own name operand, so with no other quoted field `Name` ends up equal to `VPN` ("default") instead of "" |
| `TestNewBlockVPNNameWithSpaces` | A message-vpn whose own name has spaces (semp/all.cli:20, 379, `"A VPN WITH LONG NAME AND SPACES"`) must round-trip in full -- BUG: `newBlock` splits the opener with `strings.Fields` before hunting for the quoted operand, so a multi-word name fragments and both `Name` and `VPN` come back as just its first word |
| `TestNewBlockVPNQualifierWithSpaces` | The client-username-side twin: any object scoped to that same multi-word VPN is mis-attributed the same way, which is what would actually break the VPN-scope filter export/import relies on |
| `TestParseBlocksSectionAttribution` | The `"! Configure X"`/`"! Create X"` comment maps to `Block.Section` with the `"! "` prefix and trailing `":"` stripped, but a quoted operand inside the comment (which has no trailing `":"`) survives untouched |
| `TestParseBlocksPragmaBetweenOpenerAndChildren` | A `"! pragma:..."` line at column 0 between an opener and its indented children (semp/all.cli:45-48, `create interface "intf0"`) must not end the block or the object silently loses every attribute after it; pins that BOTH the opening and the between pragma land in the same block's `Lines`, in source order, and that only one `interface` block results |
| `TestParseBlocksNestedSubBlockOwnsBothExits` | Block boundaries come from INDENTATION, never from matching `exit`: the fixture's message-vpn block (index 7) nests `authentication`/`authorization`, each closed by its own `exit` one level deeper than the outer message-vpn's own -- an exit-matching parser would stop at the first nested `exit` and produce two or three objects instead of one |
| `TestParseBlocksMissingEndTerminatorRefused` | `"! END"` is the only evidence a capture is complete (`show current-config` writes to stdout, and a truncated transfer is still valid CLI), so its absence is a loud refusal naming `END`, never a partial parse that looks like success |
| `TestParseBlocksIndentedLineWithNoOpenerRefused` | An indented line with no preceding opener means the capture began mid-block (truncated or hand-edited); refusing beats silently dropping the orphaned attributes |
| `TestParseBlocksCRLFMatchesLF` | A capture crossing a Windows host (`\r\n`) parses identically to one that did not -- asserted rather than assumed, since this repo has a standing problem with fixtures flipping line endings |

### broker_test.go

| Test | What it covers |
| --- | --- |
| `TestField` | `field` extracts a labelled value from CLI output, empty for an absent label |
| `TestCountContains` | `countContains` counts labelled lines carrying a substring |
| `TestValidName` | The boundary validator accepts safe names and rejects empty, `..`, separators, and shell metacharacters |
| `TestPathHelpers` | `cliScriptPath`, `cliArg`, and `certPath` build the in-jail paths |
| `TestLastLines` | `lastLines` returns the tail, and the whole input when it is shorter than n |
| `TestHTTPStatusHelpers` | `isHTTP2xx` accepts only 2xx; `httpStatusLines` extracts every status line |
| `TestPrimaryRedundancyUp` | The primary health predicate requires redundancy Up |
| `TestRunCLIWrapsWithStopOnError` | The shape `RunCLI` now issues: ONE `Transport.OutputInput` call, body on stdin so it never reaches argv or a log, a generated `sh -c` skeleton that sources it through the broker's own `stop-on-error no-prompt`, and a cleanup trap -- so no caller needs a separate `removeCLI` |
| `TestRunCLIDetectsRejection` | Why the wrapper exists: `cli -Apes` exits 0 even over a refused line, so `RunCLI`'s own tail scan (`rejectionIn`, the same `failKeywords` the import driver greps for) is the only thing that becomes a Go error -- the error never quotes the rejected line, which can carry a password, and it wraps the `ErrCLIRejected` sentinel so a caller can classify it. The classification was a `strings.Contains` on this error's own wording until this row: rewording the sentence would have silently turned every rejection into an "unreachable" report |
| `TestRunCLITransportErrorIsNotARejection` | The other half of that classification: a call that never reached the broker must NOT satisfy `errors.Is(err, ErrCLIRejected)`. A rejection stopped the broker at a known line; an unreachable broker may have applied all, some or none of the script, because the sourced script keeps running inside it -- so conflating them tells the operator the wrong thing about what is live right now |
| `TestRunCLICleanRunNoError` | The mirror: a transcript with none of `failKeywords` in its tail must not fail the call |
| `TestRunCLIRejectsBadName` | An invalid script name is rejected before any upload happens |
| `TestSkipIfStandalone` | The HA-only guard is false in HA and true in standalone |
| `TestServerCert` | The uploaded bundle is key+cert+CA concatenated, plus the apply script |
| `TestServerCertRequiresCert` | Missing `tls.cert`/`certKey` errors |
| `TestDomainCerts` | Each CA file is uploaded and the load script matches the generated one |
| `TestDomainCertsRejectsBadName` | A CA name with a space is rejected before any upload |
| `TestDomainCertsEmptySkips` | No configured certs means no calls at all |
| `TestDisableDefaultVPN` | The hardening script is uploaded and its scripts are cleaned up with one `rm -f` |
| `TestDisableDefaultUsers` | Every parsed VPN gets a `client-username default` line |
| `TestDisableDefaultUsersNoVPNs` | Unparseable VPN output means the step does not run |
| `TestProductKeys` | The generated product-key script is uploaded verbatim |
| `TestProductKeysDetectsError` | A broker-reported failure in the output fails loud |
| `TestProductKeysEmpty` | No keys configured is an error |
| `TestProductKeysRejectsMultilineKey` | A product key with a newline or CR is rejected before anything is uploaded -- it would append commands to a CLI script already running as admin -- while an opaque vendor key with `+`, `/` or `=` is accepted, since a key's alphabet is not this tool's to constrain |
| `TestRemoveDomainCerts` | The removal half of the domain-CA pair emits a script naming the CA, reachable from both platforms |
| `TestRemoveDomainCertsRejectsBadName` | A CA name with a space is rejected before any upload |
| `TestRemoveDomainCertsEmptySkips` | No configured CAs makes no calls |
| `TestRemoveServerCerts` | The removal runs the CLI form confirmed on a live broker (`home` / `enable` / `configure` / `no ssl server-certificate`), compared WHOLE rather than by substring -- a `no ssl server-certificate` reached without `enable`/`configure` first is rejected by the broker, and a substring match would not notice |
| `TestRemoveServerCertsSpansEveryRole` | Every role it is given gets the script. The apply path loads the certificate onto the whole group, so a removal that reached one node would leave a half state that survives a failover |
| `TestRemoveServerCertsRunCLIError` | A failure on one node aborts the loop and shows nothing -- half a removal reported as success is the outcome worth preventing |
| `TestRemoveProductKeys` | The revocation half, a loud placeholder until `no product-key <key>` was confirmed. `ProductKeys`' mirror image by design, so the same properties are asserted of it |
| `TestRemoveProductKeysRefusesAnEmptyList` | Nothing configured means nothing to revoke -- reporting success for having done nothing is what the apply path already refuses |
| `TestRemoveProductKeysValidatesBeforeUploading` | Each key is checked before anything reaches the broker, since it is interpolated into a CLI line that runs with admin already enabled -- the same order `ProductKeys` and `DomainCerts` use |
| `TestRemoveProductKeysScansTheOutput` | The property that matters most: revoking a key the broker does not hold is the kind of thing a CLI reports in prose and returns zero for, so without the `error`/`fail` scan the command would report success while the entitlement is still there |
| `TestExecCLI` | A local script is uploaded under its base name and cleaned up afterwards |
| `TestExecCLIRejectsBadName` | A base name of `..` is rejected before upload |
| `TestExecCLIReportsRejectedLines` | L2: a script line the broker rejects does not stop the rest -- the whole script still runs, its output is still shown, and the uploaded script is still removed -- but ExecCLI now returns an error naming how many lines were rejected, without quoting the rejected line itself since a CLI transcript can carry passwords |
| `TestLogin` | A 2xx SEMP response succeeds, and the password rides stdin, never the argv |
| `TestLoginFailure` | A 401 reports failure without erroring |
| `TestLoginNoResponse` | Empty output is reported as "no HTTP response" |
| `TestLeaderStandaloneSkips` | Leader makes no calls in standalone mode |
| `TestLeaderSuccess` | Activity is reverted on the backup, leadership asserted on the primary, sync reported |
| `TestLeaderTimeout` | Redundancy that never recovers times out and dumps `show redundancy detail` |
| `TestRedundancySuccess` | The full failover handshake walks its scripted primary and backup sequences to completion |
| `TestRedundancyStandaloneSkips` | Redundancy makes no calls in standalone mode |
| `TestDiagnostics` | The dest dir is created, the gather script matches, and both the configs zip and the diagnostics bundle are downloaded under their expected names |
| `TestServerCertBundleOrder` | What the bundle CONTAINS and in what order, with no transport involved -- so the container platforms and the CLI path are compared against one definition rather than against each other's behaviour. Key before certificate is the order the CLI path has always written; the CAs are deliberately absent, since trusted CAs are installed into the broker's own trust store by `broker configure domain-certs` and are not part of the certificate the broker presents |
| `TestServerCertBundleRequiresBothHalves` | The guard including the asymmetric cases: a certificate with no key cannot produce a usable bundle, and neither can a key with no certificate |
| `TestServerCertBundleReportsAnUnreadableFile` | The error names the path. Once the container platforms build this bundle at deploy time a mistyped path is the likeliest failure, and it has to point at the file rather than at TLS in general |
| `TestServerCertBundleRefusesAKeyBearingCert` | The misconfiguration this concatenation would otherwise turn into a silently wrong file: a `tls.cert` that already carries its private key, alongside a `tls.certKey`, yields a bundle with the key twice. A deployment predating this tool is exactly where a pre-chained file turns up, so the message has to say which of the two fields to change |

### coverage_test.go

Branch coverage for the paths the happy-path tests in `broker_test.go` cannot reach.

| Test | What it covers |
| --- | --- |
| `TestNewDefaults` | `New` sets the transport, config, 2s poll interval, 60 attempts, and stdout |
| `TestOutDefaultsToStdout` | `out()` falls back to stdout when `Out` is nil |
| `TestShowWritesToStdout` | `show` writes through the resolved default sink |
| `TestSleepElapses` | A tiny poll interval sleeps and returns nil |
| `TestSleepZeroIntervalReturnsCtxErr` | A zero interval with a live context returns nil |
| `TestSleepCancelled` | A cancelled context surfaces `context.Canceled` |
| `TestPollCondError` | A condition error is propagated rather than retried away |
| `TestPollContextCancelled` | A cancelled context ends the poll with `context.Canceled` |
| `TestRemoveCLIWarnsOnFailure` | Failed cleanup warns instead of erroring, and still issues `rm -f` for every path |
| `TestFieldLabelWithoutColon` | A label line with no colon-space separator yields empty |
| `TestLastLinesEqualCount` | `lastLines` when n equals the line count |
| `TestExecCLIWarnsAndFailsOnErrorOutput` | L2: error-looking CLI output still logs the `[WARN]`, but ExecCLI now also fails once the script has finished running, rather than warning and reporting success |
| `TestExecCLIRunError` | A failed run errors, and cleanup is still attempted |
| `TestServerCertBundleReadError` | Unreadable cert/key files error |
| `TestServerCertCAReadError` | An unreadable CA file errors |
| `TestDomainCertsBadFilename` | A certificate *filename* with a space is rejected (the CA name being valid) |
| `TestDiagnosticsTwoRolesNoBundle` | Output with no "Diagnostics saved" line pulls only the configs zip, once per role |
| `TestDiagnosticsRunError` | A transport `Run` failure surfaces |
| `TestLeaderPollCondError` | A failing `show redundancy` propagates, and the detail dump still runs |
| `TestRedundancyShowRDError` | The initial `show redundancy` error surfaces |
| `TestRedundancyUnhealthyPrimary` | An unhealthy primary fails and its output is shown |
| `TestRedundancyNeitherActive` | Neither node locally active is a loud failure |
| `TestRunCLIReadUploadError` | `runCLIRead` wraps and returns an Upload failure with the script name and never invokes the CLI binary. The property used to be asserted against `RunCLI`, which had the same upload-then-exec shape; `RunCLI` now sends one call, so the window between writing a script and running it exists only here |
| `TestRunCLIExecError` | What replaced that window: a failure of `RunCLI`'s single call is wrapped with the script name, no separate Upload is made (the body rides that call's stdin), and it really was ONE transport call |
| `TestServerCertUploadError` | ServerCert stops before apply-server-certs when the certificate bundle upload fails |
| `TestServerCertRunCLIError` | ServerCert returns the apply-server-certs error and never shows output |
| `TestDomainCertsUploadFileError` | DomainCerts stops before load-domain-certs when a CA file upload fails |
| `TestDomainCertsRunCLIError` | DomainCerts returns the load-domain-certs error and never shows output |
| `TestDisableDefaultVPNDisableError` | DisableDefaultVPN stops before reading back show-vpn when disabling the VPN fails |
| `TestDisableDefaultVPNShowError` | DisableDefaultVPN surfaces a failed show-vpn readback and skips cli-script cleanup |
| `TestDisableDefaultUsersShowVPNError` | DisableDefaultUsers stops before disabling anything when the VPN listing fails |
| `TestDisableDefaultUsersDisableError` | DisableDefaultUsers surfaces a failed disable-default-usernames call rather than reporting success |
| `TestProductKeysRunCLIErrorStopsLoop` | ProductKeys stops its per-role loop on a transport failure and never touches the remaining role |
| `TestExecCLIUploadFileError` | ExecCLI reports a failed upload by the script's basename and skips exec/cleanup entirely |
| `TestRemoveDomainCertsRunCLIError` | RemoveDomainCerts fails loud on a rejected removal and skips the cleanup rm |
| `TestLoginTransportError` | Login returns a wrapped transport error instead of writing a `[FAIL] Login` outcome as if it got an HTTP response |
| `TestLeaderRevertActivityError` | Leader aborts before polling redundancy when the initial Backup revert-activity fails |
| `TestLeaderAssertLeaderError` | Leader returns the assert-leader error after a healthy poll without showing partial output |
| `TestRedundancyReleaseError` | Redundancy aborts immediately when releasing activity to the Backup fails, never querying the Backup |
| `TestRedundancyBackupShowError` | Redundancy aborts when its own post-release Backup show-rd read fails, after releaseToBackup already succeeded |
| `TestRedundancyRevertToPrimaryError` | Redundancy surfaces a revertToPrimary failure rather than declaring the drill successful |
| `TestReleaseToBackupReleaseError` | releaseToBackup stops before any show-rd poll when the initial release exec fails |
| `TestReleaseToBackupReleasedTimeout` | releaseToBackup times out and never sends no-release when the release never converges |
| `TestReleaseToBackupNoReleaseError` | releaseToBackup surfaces a failed no-release exec instead of proceeding to the un-released poll |
| `TestReleaseToBackupUnreleasedTimeout` | releaseToBackup times out rather than declaring the Backup active when un-release never converges |
| `TestRevertToPrimaryRunCLIError` | revertToPrimary stops before entering its poll when the Backup's revert-activity exec fails |
| `TestShowRDPairPrimaryError` | showRDPair returns empty strings and the Primary's error without ever querying the Backup |
| `TestDiagnosticsMkdirError` | Diagnostics fails loud with the destination path when it cannot create the diagnostics dir |
| `TestGatherNodeDownloadError` | a failed main-archive Download fails the whole node's diagnostics gather with no local file produced |
| `TestGatherNodeBundleDownloadWarnsOnly` | a failed diagnostics-bundle download only WARNs, naming the bundle, and Diagnostics still succeeds |
| `TestGatherNodeBundleCleanupWarnsOnly` | a failed bundle-cleanup rm only WARNs and Diagnostics still succeeds overall |

### diff_test.go

`diff.go` is the ONLY error detection the config-import feature has: `cli -Apes <script>`
exits 0 even when the broker rejects a line, so the sole way an import learns a line was
silently dropped is to read the target back and `DiffBlocks` it against what the artifact
asked for. A false pass here -- want reported satisfied when the target does not actually
carry it, or a block that should have been skipped reported as a failure -- means an operator
is told an import succeeded over a broker that rejected part of it. Every test is chosen to be
a case a naive "compare the line sets" implementation gets wrong, per `DiffBlocks`' own doc
comment.

| Test | What it covers |
| --- | --- |
| `TestDiffBlocksDirectionalGotMaySupersetWant` | A target's own re-export fills in defaults the artifact never mentioned, and those extra lines must not read as a difference -- proven directional (not accidentally symmetric) by swapping want/got and asserting the swap DOES break the match, naming the one line the reversed want side needs |
| `TestDiffBlocksReportsExactMissingLine` | A block that landed short reports the exact missing line, not the whole block -- sending an operator to re-read a block that mostly landed would be the wrong signal |
| `TestDiffBlocksAbsentBlockLandsInMissingNotDiffering` | A block the target never created at all (`Absent`) sorts into `Missing`, never `Differing`, so a caller can tell "landed short" from "never landed" without inspecting every `BlockDiff` |
| `TestDiffBlocksIdentityIgnoresIndex` | `blockKey` excludes `Index`: two blocks sharing (Kind,Name,VPN) but different Index must match (the target orders/numbers its own export independently of the artifact), while two blocks sharing an Index but differing Kind must NOT be treated as the same object |
| `TestDiffBlocksPoolsRepeatedIdentityBeforeComparing` | The failure mode diff.go's own doc comment names: a real capture repeats "authentication" three times (an LDAP-profile pass, a Radius-profile pass, the real attribute-carrying pass) sharing one identity -- a lookup that took "whichever block is found first" would land on a near-empty early pass and read the real attribute as missing, so all matching `got` occurrences are pooled before comparing |
| `TestDiffBlocksMergeKeepsEachLinesOwnSection` | Each line keeps the Section of the block it came from across the identity merge, asserted where it matters: skipping ONE contributing section must still leave the other's missing line visible. The merge used to keep only the last Section, so `importIgnore` applied that one disposition to every line and a skipped or unclassified sibling silenced the others -- dropping them out of the only error detection this feature has |
| `TestDiffBlocksSeesAMismatchInsideOneNestedSubObject` | The second of two criticals found in review, and the worst: a rejected line inside ONE nested sub-object, where a sibling carries the identical text. Flat set membership found it "somewhere in the block" and reported a clean import -- in the real capture all 901 queues of a VPN are children of one column-0 `message-spool` opener, so per-queue failures were invisible. Lines are now qualified by the openers enclosing them |
| `TestDiffBlocksReportsTheHighestMissingAncestorOnly` | When a whole sub-object never landed, only its opener is reported -- its attribute lines are suppressed under it, so one failure reads as one row instead of twenty-five |
| `TestDiffBlocksIgnorePredicateDropsLinesAndWholeIgnoredBlocks` | The ignore contract's two halves: a line an apply-filtered rule deliberately dropped does not read as a broken import, and a whole section the import never intended to apply (nothing on the target under that identity) does not read as an ABSENT block |
| `TestDiffBlocksNormalizesWhitespaceButNotCaseOrQuotes` | A real capture double-spaces some lines and nothing guarantees the target's re-export reproduces the same run length, so whitespace differences must not count -- but `"55003"` and `55003` are different tokens to the broker and must still be reported |
| `TestDiffNormalizeLinePreservesIndentationCaseAndQuoting` | The direct unit test of `normalizeLine`'s rule: internal whitespace runs collapse and trailing whitespace trims, but leading indentation (structural, per blocks.go), case, and quoting are all preserved exactly |
| `TestDiffResultOKOnlyWhenDifferingAndMissingEmpty` | The exact predicate `RenderDiffResult` and the import command's exit-status logic rely on: `OK()` is true only when both `Differing` and `Missing` are empty, regardless of how high `Matched` is |
| `TestDiffBlockLabelFormatsIdentityForTheReport` | `blockLabel`'s branches: a VPN-scoped object gets a `(message-vpn "X")` qualifier, the VPN block itself is not qualified against its own name, a bare scope opener falls back to no name at all, and a Kind-less/Name-less block falls back to its `Section` comment (which can still take the VPN qualifier) |
| `TestRenderDiffResultCleanReportHasNoFailureRows` | A clean verification prints no table, no `[FAIL]` tag, and no per-block detail -- a report that always printed the same scaffolding would bury the one signal an operator is reading for |
| `TestRenderDiffResultDirtyReportNamesOffendingSectionAndLines` | A dirty report's only actionable content: one `[FAIL]` row per offending block, each naming its section label and the exact missing lines, for both an absent block and a partially-landed one |

### importdoc_test.go

The `-update` golden for `docs/import.md`, generated from `sections.go`'s classification by
`renderImportDocs` -- the same drift-gate-and-generator shape `internal/cli/commanddoc_test.go`
established, and the reason an operator has any way to know what `import-config` will touch
without reading Go source: the classification is baked into the binary, so a hand-written page
would be a second copy of it and the two would drift.

| Test | What it covers |
| --- | --- |
| `TestImportDocs` | Renders `docs/import.md` from the live `Rules()` table and fails while the committed file is stale; `-update` rewrites it |
| `TestEverySectionRuleHasAReason` | Every rule's `Reason` is non-empty (an empty one tells an operator nothing about why their configuration was skipped, the whole value of publishing the table), and `DropLines` is set if and only if the disposition is `ApplyFiltered` |
| `TestOnlyOneSectionIsAppliedFirst` | No more than one section carries `First` -- two would share an invocation and silently drop whichever ran second (the exact-one-section assertion is strengthened in `sections_test.go`'s `TestSectionsFirstIsCreateLoggingOnly`) |

### importops_test.go

`internal/broker/importops.go`: `ExportConfig`, `ImportPlan`, `ImportApply`, `ImportVerify`
and their pure helpers. Reuses `newTestOps`/`fakeTransport`/`matchCLI` (`broker_test.go`) and
`loadSample`'s fixture (`testdata/currentconfig_sample.cli`, `blocks_test.go`) rather than a
second set of doubles; `minimalCapture` builds the smallest byte slice `ParseBlocks` accepts,
for tests that only care about this file's own logic and not specific captured content. This
feature's only error detection is the re-export diff (diff.go's file comment): `cli -Apes`
exits 0 even when the broker rejects a line, so every test here is aimed at what IS checkable
before that point -- what gets uploaded, in what order, and what refuses before touching the
broker at all.

| Test | What it covers |
| --- | --- |
| `TestImportOpsExportConfigScopeSelectsCLICommand` | The scope-to-command mapping: no VPNs runs the single whole-broker capture; `--vpn` runs one capture PER NAME, each scoped to that VPN alone, in order, with the uploaded script removed after each run |
| `TestImportOpsExportConfigRejectsUnsafeVPNNameBeforeTransport` | A VPN name rides a quoted CLI operand (`show current-config message-vpn "<name>"`), so a double quote or control character is refused before ANY capture is attempted -- including when an earlier name in the list is valid, so the valid ones are never captured (and uploaded to the broker) before the bad one is noticed |
| `TestImportOpsExportConfigBrokerOnlyDropsVPNRegions` | `--broker-only` drops both the create-VPN pass (`RegionVPNCreate`) and every VPN's own objects (`RegionVPN`) -- missing the create-VPN pass is the dangerous half, since a "broker-only" artifact that still carried it would silently create every message-VPN the source broker had on whatever target it was imported into |
| `TestImportOpsExportConfigRefusesEmptyCapture` | An empty capture is a hard failure, not an empty-but-valid artifact -- it would otherwise parse and import as "delete nothing, apply nothing" and report success, hiding whatever went wrong with the capture itself |
| `TestImportOpsImportPlanRefusesRedactedArtifact` | An artifact whose header claims redaction has had exactly the credential material an import needs to put back stripped out, so it is refused BEFORE the target is ever read. Export never asks the broker to redact (there is no `--redact` flag -- it could only produce a file import declines), so this guards a hand-edited header or a capture taken by hand |
| `TestImportOpsImportPlanRefusesCrossTypeAndUnknownType` | `checkSameType`'s whole truth table: a software/appliance mismatch in either direction, and either side coming back unknown -- an appliance capture carries interface/VRF/DNS/clock/SNMP sections a software broker has none of, so a cross-type import would apply configuration the target cannot make sense of |
| `TestImportOpsImportPlanSplitsExistingAndNewVPNs` | The Existing/NewVPNs split against a full capture of the target, three ways -- including a VPN whose name contains SPACES, the regression that matters: the old `show message-vpn *` table parser split that name on whitespace, so an existing VPN was filed as new, never torn down, its `create` blocks applied on top of it and rejected invisibly, and `ImportVerify` then PASSED because the untouched VPN still satisfied every artifact line |
| `TestImportOpsPlanSectionsUnclassifiedSectionWarns` | A section this build's table has never seen must Skip AND warn by name, never default to Apply -- the exact hazard sections.go's design exists to prevent |
| `TestImportOpsReadTargetFailsLoudly` | Both ways the target's VPN listing can fail are ERRORS naming their step, and neither returns a list. "I could not tell" must never arrive as an EMPTY list: empty means "nothing needs tearing down", which downgrades the import to apply-on-top, where blocks land on a live VPN, are rejected invisibly, and the diff then passes |
| `TestImportPlanReportPrintsWarningsBeforeTheConfirmation` | The plan report carries `p.Warnings`. They used to print only from `ImportResult`, after every existing VPN had been destroyed -- an unclassified section or an unbounceable service is exactly what an operator might refuse, and they were told once refusing was impossible |
| `TestImportIgnoreKeepsUnclassifiedSectionsInTheDiff` | A rule-based skip is a decision and is excluded from verification; an UNCLASSIFIED section is a GAP and its lines stay in the diff. Excluding both meant a broker carrying configuration this build does not know had it silently dropped and the import reported clean |
| `TestImportResultReportsTeardownAndRebuildSeparately` | On the apply-failure path -- the only place this report is the sole statement of broker state -- the teardown and the rebuild are two facts. The row read "VPNs torn down and rebuilt" off `TornDown` alone, asserting the rebuild in exactly the case it had not happened: the VPN emptied, its spooled messages gone, nothing put back |
| `TestImportOpsPlanSectionsSkipApplianceBecomesApplyOnApplianceTarget` | The one disposition that depends on the TARGET rather than being fixed at build time: skip-appliance content actually applies when the target IS an appliance, so a software-only skip would silently drop legitimate configuration on the one target it was written for |
| `TestImportOpsImportApplyTearsDownExistingVPNBeforeApplyingArtifact` | For each Existing VPN, `ImportApply` asks the TARGET to generate its own removal commands (`show current-config message-vpn "<name>" remove`) and applies that teardown before any of the artifact's own blocks -- reversing this order is a broken import, since a CLI block cannot be reapplied onto a VPN that still exists; also pins the whole upload order (teardown-capture, teardown-apply, then the artifact's First then main sections) |
| `TestImportOpsImportApplySeparatesFirstSectionFromMain` | A `First` section ("Create logging", which ends the CLI session) runs in its OWN `RunCLI` invocation, never sharing one with the rest -- sharing would mean whatever follows the logging change in that same script never runs |
| `TestBuildChunksCreatesVPNsBeforeBrokerSections` | The create-VPN pass runs before every broker-level chunk, as it does in the artifact, and is not duplicated into the per-VPN chunk. Chunking had folded it into each VPN's own chunk, which runs LAST -- so a broker-level line naming a message-VPN applied before that VPN existed. The single-script apply it replaced had this right only by rendering in artifact order |
| `TestTornDownRecordsOnlyAppliedTeardowns` | `TornDown` names VPNs the driver reported as actually torn down, not ones whose teardown was merely CAPTURED. It used to record at capture time, so every pre-driver failure told the operator their VPN was emptied and its spooled messages destroyed when nothing had been applied -- sending them to recover data that was still there |
| `TestImportOpsImportApplyTransportFailureIsNotAConfigRejection` | The distinction `applyScript`'s error text draws: a transport failure means the CLI never ran, a different failure from the broker rejecting a configuration line -- conflating the two points an operator at the wrong next step |
| `TestImportOpsImportVerifyReExportsAndDiffs` | `ImportVerify`'s whole contract, the ONLY error detection this feature has: re-export the target with NO scope filter (verification has to see the whole broker back even though teardown ran per-VPN), strip this tool's own markers before comparing, and route the comparison through `importIgnore` so a deliberately-skipped section does not read as a failure |
| `TestImportOpsImportVerifyWrapsFailureContext` | The two ways `ImportVerify` can fail before `DiffBlocks` ever runs -- the broker capture failing, and the re-export not parsing -- each wrapped naming which step failed, since the fix for each is completely different |
| `TestImportOpsRenderScriptAppliesPreambleAndDropsFilteredLines` | An applied script carries the mode preamble (home/enable/configure) but not a header comment, and an apply-filtered rule's excluded line (`"Configure Routing"` dropping `interface "intf0"`) is dropped from the rendered body even though the rest of the block is kept |
| `TestImportOpsRenderScriptEmptySelectionRendersNothing` | An empty name selection renders an empty body and a nil applied list, so a caller (the "first" invocation when nothing is `First`) can tell "nothing to apply" from "an empty script" and skip the `RunCLI` entirely |
| `TestImportOpsVPNChunkCarriesVPNRegionsAndBrokerChunkDoesNot` | The regression test for the worst bug this feature had: `renderScript`'s predecessor filtered every non-broker-region block, so `ImportApply` tore a conflicting VPN down via the target's own `remove` script and applied NOTHING back -- purely destructive, reported as success. Both VPN regions must still be rendered SOMEWHERE, but they now live in two chunks deliberately: the create pass precedes the broker sections (a broker line naming a VPN is rejected before that VPN exists) and the VPN's objects follow. So it asserts the create is in the create chunk, the objects are in the VPN chunk, and the create is NOT duplicated into both -- applied twice, the second is a rejected duplicate |
| `TestImportOpsRenderScriptDoesNotFilterVPNLines` | `DropLines` belongs to a broker section's rule and must never apply to a VPN block -- a queue's own `max-spool-usage`, which the broker-level Message Spool rule drops, must survive intact or importing a queue would silently lose its quota |
| `TestImportOpsImportIgnore` | The three outcomes `importIgnore` feeds the verification diff: every line of a Skip/SkipAppliance section is ignored, an ApplyFiltered rule's own dropped line is ignored, and any other line (including a kept line of an apply-filtered section) is not |
| `TestImportOpsDescribeScope` | `describeScope`'s three mutually exclusive forms (`"total"`, `"broker"`, `"vpn:a,b"` sorted regardless of input order, and `--broker-only` winning even when `--vpn` is also set) -- baked into the artifact's own `scope` metadata and read back by a human deciding whether to trust it |
| `TestImportOpsReportsDoNotPanicOnEmptyAndPopulated` | Every `Report` method this file defines against both an empty and a populated value over a real `output.Sink`, so a `KVBlock`/`TaggedTable` call assuming a non-empty slice cannot hide behind only the populated case being tried; also pins that a populated `ImportPlan.Report` names the specific VPN being torn down, not merely a count |

### inject_test.go

`InjectShutdown` (`inject.go`) against the four-case contract in that file's own doc comment:
a captured `service X listen-port ...` change is wrapped in a shutdown sandwich only when
doing so cannot invent a final enablement state the artifact never attested to. Fixtures are
either transcribed verbatim from a live broker's semp/all.cli (cited by line number,
service-config lines only) or hand-built to match its shape for a scenario the real capture
does not happen to exercise. `parseInjectCapture` wraps a body in the minimal preamble/`"!
END"` `ParseBlocks` requires and parses it for real, rather than hand-building `Block`
literals, so every case stays honest about `Section` assignment and the "no" trap
`newBlock`/`parseServiceLine` both navigate.

| Test | What it covers |
| --- | --- |
| `TestInjectShutdownAlreadySandwichedPassesThroughUnchanged` | Case 1: every port change in a real broker-level service run (semp, smf, mate-link, redundancy, web-transport, rest, mqtt, amqp, health-check) already sits inside a real shutdown/un-shutdown pair or has no broker-level port line at all, so nothing may be injected and no block's content may change |
| `TestInjectShutdownNeverChangesEnablement` | The central guard, stated as a tally rather than a line diff: for every (section, service, transport) the shutdown/un-shutdown counts after the transform must equal what the artifact specified before it -- a second, independent way of proving the property held even against a future change that kept line COUNT but not per-key PAIRING |
| `TestInjectShutdownInjectsLeadingWhenTrailingExists` | Case 2: no leading shutdown but a real trailing `"no ... shutdown"` exists later in the section (semp's ssl transport) -- only the leading line may be synthesised, and the artifact's own trailing line must survive byte-for-byte |
| `TestInjectShutdownCaseThreeRestoresTargetState` | Case 3 in both directions, on smf's four bare listen-port lines with no wrapper at all: the injected pair must match whatever the target reports (enabled -> `no ... shutdown`, disabled -> `shutdown`), never default to "enabled", and the pair is injected once around the whole run (GROUPING), not once per port line |
| `TestInjectShutdownCaseThreeUnknownTargetLeavesPortUnchanged` | Case 3's "unknown" branch, both ways it is reached (nil `targetState`, and one that answers `known=false`): the port line is left completely unchanged and a warning fires naming the service and "no reachable target state" -- the safe failure is to change nothing, never to guess |
| `TestInjectShutdownUnknownServicePassesThroughWithWarning` | Case 4: a service absent from `ServiceShutdownStyles` passes through unchanged with a warning naming it and "not in the shutdown table", rather than being silently dropped or guessed at |
| `TestInjectShutdownTransportGranularitySemp` | Property 6 for semp's transport-after style: injecting ssl's shutdown must not touch a plain-text port change that already has its own real sandwich right next to it -- a key that ignored transport would misfire on the wrong lines or double-wrap them |
| `TestInjectShutdownTransportGranularityRest` | Property 6 for rest's transport-before style: incoming and outgoing each need their own shutdown/restore around their own port line, with no capture-side sandwich to key off, exercising both case-3 target directions per transport |
| `TestInjectShutdownParsesRealQuotingVariants` | `parseServiceLine`'s tokenisation directly against three real quoting/spacing shapes (smf's double-quoted double-spaced operands, semp's unquoted port with a trailing transport keyword, amqp's quoted port with an unquoted trailing token) |
| `TestInjectShutdownServiceShutdownStylesTable` | `ServiceShutdownStyles()`'s documented contents -- the exact service set and how each spells its shutdown command (transport-after, transport-before, bare) -- so a service silently added to or dropped from the map fails here even without its own dedicated test |
| `TestInjectShutdownUntaggedSempTransportNamesPlainText` | The fix for a real defect: semp's untagged `listen-port 8080` line (semp/all.cli:66) IS the plain-text transport, but its shutdown command carries the literal word `"plain-text"` (semp/all.cli:65,67) -- keying the untagged port line on transport `""` put it in a different key from that shutdown line, so an unsandwiched plain-text port line rendered a BARE `"service semp shutdown"` (semp's whole-service shutdown), which would have taken the ssl transport down too; the fixed behaviour names `"plain-text"` explicitly and leaves ssl untouched |

### matechannel_test.go

| Test | What it covers |
| --- | --- |
| `TestMateChannelPreflightLearnsTheBrokerType` | The preflight is a real read, and the broker type comes from the BROKER rather than from configuration -- the mate grammar and the `show replication` layout both differ by platform, and a `type:` key in the env file would be a second source of truth whose only effect when wrong is to emit CLI the target rejects. The answer is cached, so a later read costs no extra round trip |
| `TestMateChannelPreflightRefusesAnUnknownType` | A login banner naming neither platform stops the run, and the error names the site. Guessing would emit `connect-port` lines to a software broker or six `connect-via` lines to an appliance |
| `TestMateChannelPreflightReportsTransportFailure` | An unreachable site fails at the preflight with the transport's own error passed through -- every flow calls it before the first write, so the pair is undisturbed when it stops |
| `TestMateChannelShowReplication` | The mate report is parsed in the grammar the banner selected, end to end through the channel |
| `TestMateChannelReadRolesNarrowsToTheListedVPNs` | The blast-radius guard: the broker reports every VPN it has, and only the ones asked about come back. A VPN the broker does not have is ABSENT rather than an error, since the caller knows what it asked for |
| `TestMateChannelSetRole` | The role command quotes the VPN name (real ones contain spaces), turns paging off after `home`, and writes the operand as a BARE keyword -- `state "active"` is how a capture echoes it, not what the grammar takes |
| `TestMateChannelSetRoleRefusesNA` | `n/a` is a reading, not a state to write, and a refused role never reaches the broker at all |
| `TestMateChannelSetRoleDetectsRejection` | Every `failKeywords` phrase in the transcript fails the command despite `cli -Apes` exiting 0 -- the fact the whole import driver exists to work around. A role change that silently did nothing would leave a switchover believing it had moved |
| `TestMateChannelRejectionScanIsTailOnly` | The scan reads the END of a transcript. Scanning the whole thing would scan the configuration itself, where an object legitimately named `error-events` matches `error:` |
| `TestMateChannelStatusIsBothReports` | `Status` returns the raw text of both `show` commands, for display only |
| `TestMateChannelDescribeNamesTheTarget` | Every channel says what it is, so an error from a two-site flow names WHICH site |
| `TestMateChannelPrimaryActive` | The activity gate reads real `show redundancy` output: the primary holding activity passes, the backup holding it fails, and a STANDALONE broker (redundancy shut down) passes rather than being refused for failing to hold activity it cannot hold |
| `TestLocalMateRunsThroughOps` | The local channel is the same implementation over the transport the ops already hold, and addresses the PRIMARY only -- the replication role config-syncs to that site's HA backup |

### omit_test.go

Where `skip` is enforced: at export. The operator's rule is that export REMOVES a skipped
section and import merely ignores one if it meets it -- which, after export has done its
job, should never happen. So the artifact is exactly what an import applies, records what
it left out as `! solace-util-omitted:` lines, and the import plan mentions a skipped
section only when an older or hand-edited artifact still carries one.

| Test | What it covers |
| --- | --- |
| `TestOmitAtExportRemovesSkipSectionsAndRecordsThem` | every block of a Skip section is dropped, the Apply and VPN blocks survive, and the omission is recorded ONCE per section with the rule's own reason |
| `TestOmitAtExportDropsFilteredLinesAndEmptyBlocks` | an apply-filtered section loses its dropped lines but keeps the rest (`routing interface "intf0"` goes, the block stays); a one-liner whose only line is dropped (`max-spool-usage`) is removed whole rather than left as an opener-less husk; nothing is recorded as an omission |
| `TestOmitAtExportLeavesTargetDecisionsToImport` | SkipAppliance sections and unclassified sections are NOT removed -- the first is a decision about the target's type only import can make, the second is something import refuses loudly, and removing either here would hide it |
| `TestOmitAnnotateRoundTripsOmissions` | `Annotate` writes one `! solace-util-omitted:` marker per omission with section and reason, and `ParseBlocks` reads them back into `Capture.Omitted` |
| `TestOmitExportConfigEndToEnd` | the real export path over the sample fixture: `Configure System` (Skip) is absent from the artifact, the omission marker names it, and the VPN and its objects are untouched |
| `TestImportPlanRefusesAFileItDidNotExport` | the provenance gate: a raw `show current-config` capture, a hand-written CLI script, and an artifact from a future export format are each refused, and every refusal names the command that produces an importable one. RISK: a raw capture looks importable and is not -- its Skip sections were never removed and its transcript never cut, so it would sever the SEMP channel and feed echoed prompts to the CLI, both invisibly |
| `TestImportPlanAcceptsWhatExportProduced` | the other half, end to end: whatever `ExportConfig` writes passes `ImportPlan`'s own gate -- a marker written one way and read another would refuse every real artifact |

### replication_test.go

| Test | What it covers |
| --- | --- |
| `TestRenderMateSoftware` | The software grammar: one `connect-via` per endpoint with the port inside the address, the router name standing alone, and NO keyword on a plain-text entry -- the CLI's own default, which keeps two spellings of one thing out of the emitted script. Two addresses on one transport render as exactly two lines, not a cross product |
| `TestRenderMateAppliance` | The appliance grammar: a `connect-port` per transport, the single address hung off the router-name line, never a `plain-text` keyword (the appliance grammar has none, so a bare `connect-port` IS that case) and never a port on `connect-via`, which is illegal there |
| `TestRenderMateApplianceRefusesUnrenderable` | The two sets an appliance cannot express are refused BY NAME rather than truncated: two distinct hosts (it has one `connect-via`) and two endpoints sharing a transport (`connect-port` is not repeatable). Dropping either would point replication at an address nobody chose |
| `TestRenderMateNeedsABrokerType` | `BrokerUnknown` is an error for both the renderer and the reader, never a default. The grammar and the report layout are not preferences -- guessing emits CLI the target rejects |
| `TestRenderMateRejectsEmpty` | A mate with no virtual-router-name, and one with no endpoints, are both refused rather than rendering a mate that points nowhere |
| `TestParseShowReplicationSoftware` | The transport-grouped layout: a continuation line is a second address for the SAME transport, the empty `Plain Text`/`Compressed` headings are NOT endpoints (a heading prints whether or not the transport is configured), and the address splits on the LAST colon so a host carrying dots and underscores survives |
| `TestParseShowReplicationAppliance` | The one-host layout: `Connect-Via` plus a port per transport under `Connect-Ports`, with the top-level `SSL:` cipher section not mistaken for a transport heading despite carrying the word |
| `TestParseShowReplicationWrongPlatformFails` | Each reader run on the other platform's report FAILS rather than returning an empty mate. A silent empty would read as "replication is unconfigured" and be carried on from |
| `TestParseShowReplicationUnconfigured` | A broker with no mate parses to an empty `MateConfig` on both platforms rather than erroring -- the ordinary state of most brokers, distinguishable from the wrong-platform case only because the mate is unnamed too |
| `TestMateRoundTripsThroughItsOwnGrammar` | What a broker reports re-renders to what a broker would be told, in both grammars |
| `TestSameEndpointsNormalisesPlainText` | An omitted transport and an explicit `plainText` are one endpoint, order is ignored, and a different transport is a different endpoint. Import's diff is directional, so a keyword this tool omits and a broker echoes would otherwise read as a missing line on every plain-text endpoint |
| `TestRenderMateRemovals` | The apply is convergent, not additive: software names the address being cleared, the appliance names only the transport, a converged broker needs no removals, and the router name is cleared only when it actually changes |
| `TestTransportVocabularyIsThreeWords` | `encrypted` is refused by name in both directions rather than folded into `ssl`. The request schema offers it as a fourth keyword, but neither report emits it and the reply schema's six-entry software cap is three transports x two addresses -- which only holds if the vocabulary is three words |
| `TestRealApplianceShowReplicationParses` | The real `semp/show-replication-appl.out` parses to one host and three ports and re-renders, proving the embedded fixture mirrors it. SKIPS when absent -- `semp/` is gitignored |
| `TestRealSoftwareShowReplicationParses` | The real `semp/show-replication-sw.out` parses to exactly two ssl addresses sharing one port, with the printed-but-empty headings yielding nothing, and the same set is refused by the appliance renderer. SKIPS when absent |
| `TestRealApplianceCaptureMateRoundTrips` | Report and configuration describe the same mate: parsing `show-replication-appl.out` and rendering it reproduces the `Configure Replication` lines of the appliance `show current-config` capture verbatim. SKIPS when either is absent |
| `TestParseVPNReplication` | The spaces-in-names test above all else: splitting this fixed-width flag table on whitespace renames `A VPN WITH LONG NAME AND SPACES` to `A` AND shifts every flag one column, so the admin state is read out of the name. Three of the six rows would be wrong. The column rule defines the fields and is read rather than assumed. It also pins the `Q` (queue) column, and that table is the evidence for the rule the re-enable keyword rests on: `default` is admin-UP WITH a queue and `vpn-01` is admin-UP with NONE, so queue existence cannot be derived from enablement |
| `TestVPNReplStateIsTwoFacts` | Admin state and role are two independent facts: a shut-down VPN still reports a role sitting at its default, so reading the role alone misreports every VPN on a broker where replication was never turned on. `n/a` stays `n/a` rather than folding into standby, which would let a switch promote against a VPN it cannot see, and `Replicating()` is true only for enabled-with-a-real-role |
| `TestParseVPNReplicationReadsColumnOrderFromTheHeader` | The flags are located by their header LETTER, not by position. A broker that reordered or inserted a flag column would otherwise have its admin state read as its role -- silently, with both values still looking valid, which is the one misreading that could send a switchover the wrong way |
| `TestParseVPNReplicationRejectsUnreadable` | Five reports that must fail loudly rather than return a partial map: no column rule, no A/C columns, an unknown admin flag, an unknown role flag, and a table with no rows (every broker has a default VPN, so an empty table means the report was not understood) |
| `TestParseVPNReplicationIgnoresTheEchoedPrompt` | A trailing CLI prompt is not a row -- it is too short to reach the flag columns. Matching it by text instead would have dropped a real row, because the capture contains a VPN NAMED after the router |
| `TestParseVPNReplicationIgnoresRepeatedHeaders` | A paginated report repeats the header and rule partway down, and the repeated RULE parses as a row: its name column trims to a run of dashes and its flag columns are literal `-`, decoding to n/a for both. Left in, a phantom VPN named `--------------------------------` joins the map |
| `TestParseVPNReplicationToleratesALongPrompt` | The other half of the prompt story: a broker whose router name is longer than the name column produces a prompt that DOES reach the flag columns, and without the gutter check its characters fail to decode and turn an ordinary successful report into a hard parse error. The gutter -- the space always preceding a flag column in a padded row -- tells the two apart |
| `TestRealVPNReplicationParses` | The real `semp/show-message-vpn-replication.out` parses, and really does carry VPN names with spaces -- the premise the whole fixed-width reader rests on. SKIPS when absent |

### replicationops_test.go

| Test | What it covers |
| --- | --- |
| `TestRoleAtSiteIsTheComplement` | The one sentence that lets a single `replication:` block be correct at BOTH sites: the file names which site owns active and each broker computes the complement for itself. Neither file states a role, so nothing is reversed when the block is copied across -- and the same block read at both sites yields exactly one active, because one `activeAt` cannot name two |
| `TestReplicationVPNLinesEnablesListedAndDisablesTheRest` | The authoritative-file rule. Listed means enabled with the derived role; unlisted BUT REPLICATING means shut down; and a VPN merely present and not replicating is left alone, since shutting down something already down is noise from a command that is destructive enough already |
| `TestReplicationVPNLinesQuotesNamesWithSpaces` | A real VPN name survives -- the live capture carries `A VPN WITH LONG NAME AND SPACES`, and an unquoted operand would name a different VPN or none |
| `TestRouterNameReadsTheCaptureHeader` | Self-identification reads the capture HEADER's quoted `Router:` line rather than `show router-name`: that format is proven on both platforms and both header generations, and the quoting ends the name unambiguously -- a router name is operator-chosen and may contain almost anything. Covers software, appliance, and an unpadded line inside a full header |
| `TestReplicationVPNLinesIsDeterministic` | Two runs against the same broker state render byte-identical scripts. The shutdown set comes from a map, and a Go map range is randomised, so without an explicit sort a diff of two applies would be unreadable |
| `TestPlannedRolesNamesEveryListedVPN` | The caller can describe the whole move before any of it runs, which is what the exact-`yes` gate has to show |
| `TestVPNReplicationBlockShape` | The emitted block matches the shape a capture uses, so a generated script reads beside one rather than against it. Checked for a block of ANY length rather than a fixed trio, because phase 2 usually sends one line -- `state <role>` alone -- and adds an enable only where one is needed |
| `TestReplicationVPNLinesSetsRoleWithoutCyclingReplication` | A role is set IN PLACE -- `state <role>` alone, against a VPN that is up and replicating, with no shutdown around it. The earlier draft cycled every listed VPN down and back up, interrupting replication on each one on EVERY run, including the runs where the mate had not changed and nothing needed to stop. The contradiction was already in the tree: `setReplicationRoleScript`, which the switchover uses, always set `state` alone against a VPN its preflight requires ENABLED. Also pins the enable KEYWORD off the QUEUE column, including the two cases reading enablement got wrong, and that `force-recreate-queue` is never emitted |
| `TestNothingEverShutsDownTheVPNItself` | The blast-radius guarantee, and the property that separates an interruption from an outage: `configure dr` disables REPLICATION and never shuts down a message-VPN. Every command either phase sends is nested inside that VPN's own `replication` node, so clients stay connected and only the feed to the mate stops. A bare `shutdown` one level higher would disconnect every client of that VPN, and on phase 1's path it would do that to every replicating VPN on the broker at once. The two forms differ by one level of indentation and nothing else, which is why this is a test rather than a comment -- the mistake reads as correct and only shows up in production. Walks the generated script the way the broker parses it, across both phases and both shapes of phase 2 |
| `TestReplicationVPNLinesReEnablesWhatPhase1StoppedOnly` | The one case where a listed VPN that was already up still gets an enable: phase 1 stopped it to converge the mate, so this command is responsible for bringing it back. The same VPN in the same state with phase 1 skipped must produce NO enable -- that difference is the point, because the no-phase-1 run is the common one. Also pins that an unlisted VPN phase 1 already stopped is reported as disabled without a second shutdown being sent |
| `TestMateConvergenceShutdownsIsDeterministic` | Phase 1's shutdown set -- a second map range alongside `replicationVPNLines`' own -- is sorted the same way, and is keyed on ADMIN STATE ALONE rather than `Replicating()`: a VPN enabled with no resolved role yet must still be shut down, because the broker's mate-change precondition is about enablement, not role. `#config-sync` (admin `-`) never appears |
| `TestConfigureReplicationReadsTypeOffReadMateConfig` | The broker type comes off `readMateConfig`'s own `show replication` round trip -- its echoed login banner -- not a dedicated call, so `Ops.brokerType` is gone. The responder answers only the scripts the command actually needs, with no `banner` case, which is what proves nothing else was asked for. It drives LISTED VPNs, because phase 1's own shutdowns now cover the unlisted ones and a run with nothing listed would correctly send no phase-2 chunk |
| `TestConfigureReplicationSkipsPhase2WhenPhase1DidItAll` | When the mate differs and the env file lists no VPN, phase 1 stops every replicating VPN and phase 2 has nothing left to say, so no second chunk is sent. Before a role could be set in place, phase 2 always had work here because it re-sent a shutdown to each VPN phase 1 had just stopped -- a second round trip and a second chance for the broker to refuse a line, for no change in state. Those VPNs are still REPORTED as disabled, because the report describes the end state rather than which phase produced it |
| `TestConfigureReplicationSkipsMateWhenAlreadyMatching` | The core safety property: when the mate configuration already matches the env file, phase 1 writes NOTHING and stops no VPN, so the common case (every re-run after the first) causes no outage. It drives LISTED VPNs on purpose, and asserts phase 2 still ran -- without that the test could not tell a skipped phase 1 from a whole apply returning early once the mate matched, and the second would silently stop applying every role |
| `TestConfigureReplicationPhase1BodyOrderIsShutdownsThenRemovalsThenMate` | The owner's specified phase-1 body order: every currently-enabled VPN shut down first (sorted), THEN the removals, THEN the new mate lines -- removals before the new lines so a changed virtual-router-name never collides with the one already set |
| `TestConfigureReplicationRefusesAMissingListedVPN` | The refuse-before-write check: a listed VPN this broker's own `show message-vpn * replication` does not report is refused before anything is written, naming the missing VPN -- a bare `message-vpn <name>` MODIFIES an existing VPN on the broker's own grammar, so an absent one would be rejected mid-phase-2 under stop-on-error and leave every listed VPN after it shut down and never re-enabled |
| `TestConfigureReplicationPhase1RejectionStopsBeforePhase2` | A phase-1 rejection: the error names the phase, says a line was refused rather than the broker being unreachable, says nothing is rolled back, points at the two `show` reports, and NAMES the VPNs left with replication down -- a category like "every VPN this broker was replicating" reads as informative and is not actionable, and the names exist nowhere else. It must also NOT claim all of them were stopped, since phase 1 sends the shutdowns first and may have been refused partway through them. Phase 2 never running is asserted by the transport, not by the text |
| `TestConfigureReplicationClosingShowFailureStillReportsTheApply` | A display-only read failing is NOT the apply failing. Both write phases have already succeeded by the time the closing `show` runs, and the result is the only record of what they did -- built in memory, never read back -- so returning that read's error would discard it and tell an operator whose configuration landed perfectly that the command failed. They would then hunt for damage that is not there, or re-run a destructive command they did not need. It warns, reports, and leaves Status empty |
| `TestConfigureReplicationPhase2RejectionAfterPhase1Success` | The worse of the two failures: phase 1 succeeded, is NOT rolled back, and the error must say so and name both the VPNs it stopped to converge the mate and the listed ones phase 2 was working through. Phase 1 having actually run first is asserted by the transport |
| `TestConfigureReplicationTransportFailureIsNotReportedAsARejection` | A connection that dies mid-call is worded differently from a rejected line: the broker's stop-on-error guarantee only holds for a call that completed, so a transport failure says the outcome is NOT known rather than claiming the broker rejected anything |

### scripts_test.go

Pins the generated broker CLI script text -- these strings are what the broker executes.

| Test | What it covers |
| --- | --- |
| `TestFixedScripts` | The seven fixed scripts (show redundancy, detail, revert, release, no-release, show vpn, show vpn bare) match byte for byte |
| `TestRevertActivityTrailingSpace` | The redundancy-test revert keeps its trailing space; the leader-path revert does not |
| `TestAssertLeaderScript` | Asserts leader for router and all VPNs, ending with the config-sync database show |
| `TestServerCertScript` | The cert filename format plus the script's prefix, load line, and closing show |
| `TestDomainCertsScriptSorted` | Map input emits CAs in sorted order for deterministic output, with the right prefix/suffix and per-CA block |
| `TestDisableDefaultUsersScriptQuoting` | A VPN name containing a space stays one quoted token |
| `TestRemoveProductKeysScript` | The revocation form confirmed on a live broker, whole: `no product-key <key>` under the same preamble |
| `TestProductKeyScriptsShareAPreamble` | Apply and remove open with the same `home`/`enable`/`admin` and close with the same `show product-key`. A removal that reached `no product-key` from a different CLI context than the apply reached `product-key` from would fail in a way no test comparing only its own literal would catch |
| `TestProductKeysScript` | Exact script text for a list of keys |
| `TestDisableDefaultVPNScript` | The script shuts the VPN down and nothing else, asserted in both directions. It used to also shut down the default client-username, all twelve services and the plaintext downgrade -- two of which were wrong to bundle in: the default client-username belongs to `broker configure default-users`, which shuts it down in EVERY VPN rather than only this one, and disabling the plaintext downgrade is a broker-wide TLS decision that also made the operation unreversible |
| `TestEnableDefaultVPNScript` | The inverse, and the reason the disable script was narrowed: an operation that changed six things could not be undone by one that changes one. `no shutdown` restores exactly what `shutdown` took away, and must not re-open the plaintext downgrade |
| `TestParseVPNNames` | Column-based VPN parsing skips the legend, header, separator, comments, and blank lines |
| `TestParseVPNNamesNoSeparator` | Output without a separator row yields no names |
| `TestGatherConfigsScript` | Prefix, first show command, `gather-diagnostics` with days substituted, and one line per configured show |
| `TestZipConfigsScript` | The zip command is present |
| `TestSortedKeys` | `sortedKeys` returns map keys in sorted order |
| `TestEveryScriptTurnsPagingOffAfterHome` | Every script generator emits `no paging` immediately after every `home`. A paginated report re-prints its column header and rule partway down, and the rule then reads as a data row of all dashes -- which is how `show message-vpn * replication` grew a phantom VPN named after its own separator. The parsers tolerate a repeat as a backstop; turning paging off at the source is the fix. It walks every generator rather than the ones that came to mind, because the gap it closes was exactly a set of scripts written with `home` and without `no paging` |

### sections_test.go

The baked-in classification `import-config` applies to each broker-level section of a
`show current-config` capture (`sections.go`). Every property here traces to a real risk: a
misclassified or unmatched section either silently fails to apply configuration the operator
chose to keep, or applies something (a service bounce, a credential overwrite, an HA-breaking
change) the operator deliberately excluded. Real line text is copied by hand from a live
broker's semp/all.cli, cited by line number, and semp/import-decisions.md carries the
operator's own reasoning per section -- docs/import.md (generated from `Rules()` by
`importdoc_test.go`) is the published classification this file must agree with.

| Test | What it covers |
| --- | --- |
| `TestSectionsRuleForExactMatch` | Every rule `Rules()` advertises is found again by `RuleFor` under its own `Section`, as itself -- guards against `RuleFor` drifting from the backing table (a rule appended but unreachable, or an earlier duplicate shadowing a later one) |
| `TestSectionsRuleForPrefixMatch` | Prefix matching against the full quoted section comments the broker actually emits (an LDAP/RADIUS/domain-CA/authentication profile name, or the redundancy PSK's one Solace-internal singleton name): equality alone would need one rule per possible operator-chosen name |
| `TestSectionsIsPrefixSection` | The membership list `RuleFor`'s prefix loop consults: every listed section matches, and near-misses (an exact-rule section that is itself a text-prefix of another, a trailing space) do not |
| `TestSectionsExactRuleNotShadowedByPrefix` | `"Configure Routing"` (ApplyFiltered, drops the host interface line) is a literal text-prefix of `"Configure Routing MNR"` (Apply, unfiltered) -- neither is in `prefixSections`, so `RuleFor` must resolve the MNR section by equality and not inherit the shorter rule's disposition or line filter |
| `TestSectionsRuleForUnknownAndEmpty` | A caller that gets `ok=false` can trust it completely: empty, an unseen section, wrong case, leading whitespace, and one letter off an exact rule all return the zero `SectionRule` alongside `ok=false`, never a non-zero leftover |
| `TestSectionsSkipDispositions` | The exact `Skip` disposition, by name, for every section the operator decided must never be applied (System, SEMP/Matelink/Redundancy Service, Router Name, Redundancy, Config Sync, Redundancy PSK) -- and that the table has EXACTLY that many `Skip` rules, so one added or removed without updating this test still fails |
| `TestSectionsSkipApplianceDispositions` | The same exact-count check for every appliance-only section (Ethernet/Lag/SolOsPhy Interfaces, ip vrf, DNS, Disk, Clock, Clock Synchronization, SNMP) |
| `TestSectionsDropRoutingInterfaceLine` | `dropRoutingInterfaceLine` against the real "Configure Routing" text (semp/all.cli:134-136): drops `interface "intf0"` (host hardware) with or without its 2-space indent, keeps `mode dynamic-message-routing defer` (would silently disable DMR if dropped too) |
| `TestSectionsDropMessageSpoolLines` | `dropMessageSpoolLines` against the real "Message Spool" text (semp/all.cli:171-200): drops `max-spool-usage`, `spool-sync mode`, and the shutdown/no-shutdown pair; keeps event thresholds, the defragment schedule, and near-miss lines (`spool-sync mirrored ...`, `max-cache-usage`) that must not match |
| `TestSectionsCreateHardwareSharesMessageSpoolPredicate` | "Create hardware" and "Message Spool" filter with the IDENTICAL `DropLines` predicate (compared by function pointer), not two hand-written copies that could drift apart |
| `TestSectionsFirstIsCreateLoggingOnly` | Exactly one section is `First`, and it is "Create logging" -- a second `First` section would fight over its own isolated invocation, and "Create logging" losing `First` would let whatever shares its script silently never run |
| `TestSectionsRulesOrderStable` | `Rules()` returns the same content and order on repeated calls (docs/import.md's row order IS this order, so nondeterminism would make the generated-doc comparison flap independent of any real change), and pins the documented file order at both ends: "Configure System" first, "Configure Management Message Vpn" last |
| `TestSectionsRulesReturnsCopy` | `Rules()` hands out a defensive copy of the backing array -- mutating a caller's result must not corrupt the baked-in classification for a later or concurrent caller |

### semp_test.go

The SEMP-over-HTTP mate control channel in isolation: port resolution, curl config
rendering, reply parsing, and the preflight/revert-activity request shapes.

Every TLS-related case below sets `serverCert` (a `config.TLS{Cert, CertKey}` pair) as the
precondition that makes the broker serve its 1943 listener at all -- port resolution and both
mate calls prefer TLS only when it is actually configured, which is what keeps a plaintext HA
group working unchanged.

| Test | What it covers |
| --- | --- |
| `TestSempPortHostMode` | Host/unset network mode with a server certificate configured resolves to the broker's own TLS port 1943 |
| `TestSempPortWithoutCertificateStaysPlaintext` | The guard against breaking a working plaintext group: with no `tls.cert`/`certKey` the 1943 listener never comes up, so host mode and a bridge mapping of both ports both resolve to plaintext 8080, and a bridge mapping of ONLY 1943 with no certificate refuses loud naming `tls.cert` rather than advising a port to map |
| `TestSempPortBridgePrefersTLS` | A bridge mapping exposing 1943 wins even when 8080 is also mapped, tolerating a `/tcp` suffix and a host-IP-qualified publish |
| `TestSempPortBridgeFallsBackToPlaintext` | With no TLS mapping present the plaintext 8080 mapping is used instead, same tolerances as the TLS case |
| `TestSempPortBridgeRefusesNoMapping` | A bridge ports list without a tcp 8080 mapping refuses loud, naming network.ports (a udp mapping does not count) |
| `TestCurlConfigLineEscapesQuotes` | Backslashes and double quotes in a config value are escaped, never corrupting the `-K -` config |
| `TestSempV1OKMatchesExecuteResultOK` | `<execute-result code="ok"/>` is matched in any attribute order |
| `TestSempV1OKRejectsErrorReply` | A code="fail" reply, a reply with no execute-result, and an empty reply are all not ok |
| `TestHTTPBodyStripsHeaders` | The header block is stripped for both CRLF and LF separators; a body with no separator passes through |
| `TestLastHTTPStatus` | The final status line wins (100 Continue tolerated); no status reads as a named placeholder |
| `TestMateSEMPPreflightSuccess` | With a server certificate configured and no `tls.cas`, the mate is reached over https on 1943 with certificate verification disabled and a warning naming `tls.cas` as the fix; exactly one call |
| `TestMateSEMPPreflightPlaintextWarns` | With no server certificate the mate is reached over plain http on 8080 exactly as before this change, but a warning fires first naming that the admin credentials cross the network unencrypted |
| `TestMateSEMPPreflightMissingIP` | An empty redundancy.backup.addr refuses before any transport call, naming the field |
| `TestMateSEMPPreflightNon2xx` | A non-2xx reply fails naming the status and the actionable causes |
| `TestMateSEMPPreflightTransportError` | A transport error is wrapped with the host-to-host reachability hint |
| `TestMateRevertActivitySuccess` | With a server certificate configured and no `tls.cas`: exact argv (`curl -is -K - https://<mate>:1943/SEMP`) with the user, data and insecure config lines on stdin, a warning naming `tls.cas`, and one call -- callers own the preflight |
| `TestMateRevertActivityCredsAndBodyNeverInArgv` | The S3 boundary: no argv token carries the password or the RPC body |
| `TestMateRevertActivityNon2xx` | A non-2xx reply fails on the status even when the body says ok |
| `TestMateRevertActivityRPCNotOK` | A 2xx reply without code="ok" fails, dumping the header-stripped reply |
| `TestMateRevertActivityBridgePlaintextOnly` | The last-resort branch: a bridge mapping only 8080 (1943 unmapped, even with a certificate configured) keeps today's plain http:// on the mapped port, but now warns first that the credentials are crossing the wire unencrypted |
| `TestMateRevertActivityBridgeTLSNoCA` | A bridge mapping exposing 1943 is preferred over a mapped 8080; with no `tls.cas` configured, verification is disabled (self-signed broker cert is the normal case) and a warning names `tls.cas` |
| `TestMateRevertActivityTLSIgnoresTLSCAs` | The opposite of what the row here used to claim. There was a "fully verified" branch passing `tls.cas[0]` to curl's `--cacert`, and it could not have worked: curl is exec'd INSIDE the broker container while `tls.cas` names files on the HOST running this tool, and nothing mounts CA material into the broker -- domain CAs go into the broker's own trust store via `broker configure domain-certs`, not onto a path curl can read. Now that host paths resolve against the env file's directory the value is unambiguously a host path, so the branch is gone rather than silently pointing curl at a file that is not there, and the warning is required even with CAs configured: the channel is encrypted, the mate's identity is not checked |
| `TestSempCurlSendsTheCallersCredential` | The SEMP login is a parameter, not a field read off the local config. It used to be hardcoded to `config.AdminUser` + `semp.adminPass`, which was correct while the only remote was the HA mate -- same deployment, same password -- but a REPLICATION mate is a different broker with its own credential, so leaving it hardcoded would send this deployment's admin password across a WAN. The check is two-sided: the caller's password present, the LOCAL one absent, and exactly one `user` line |
| `TestLocalAdminIsTheDeploymentsOwnLogin` | `LocalAdmin()` is `config.AdminUser` + `semp.adminPass`, pinning what the HA callers pass so the credential lift cannot quietly change the login those paths have always used |

### sempmate_test.go

| Test | What it covers |
| --- | --- |
| `TestSEMPMateURLAndCredential` | The channel dials the SITE's endpoint with the SITE's password, never this deployment's. A DR mate is a different broker: sending `semp.adminPass` would fail at best and, if the two happened to match, hide the misconfiguration until the day they diverged. `insecure: true` reaches curl, and no password reaches argv |
| `TestSEMPMateRequestShapes` | Every request body, pinned against the request XSD: `show replication`, `show message-vpn * replication`, `show redundancy`, and both role writes. The write has NO `<admin>` wrapper (`rpc > admin` has no `message-vpn` child at all), `vpn-name` comes FIRST because the schema puts it in an `xs:sequence` ahead of the sub-command choice, and the role operand is an EMPTY element rather than text |
| `TestSEMPMateEscapesTheVPNName` | A hand-built body is escaped. The bodies are assembled as strings because element ORDER is load-bearing, which makes escaping this code's job -- an operator-chosen name carrying `&` or `<` would otherwise produce a malformed document, or a different one than intended |
| `TestSEMPMateChecksExecuteResult` | A 200 OK is not taken as success. `execute-result` is a SIBLING of `<rpc>` under `<rpc-reply>` and carries its verdict in the `code` ATTRIBUTE, so a reply can be HTTP 200 and still say fail -- exactly the shape a rejected role change takes. The broker's own `reason` survives into the error |
| `TestSEMPMateReadsBothMateShapes` | The advantage this leg has over the CLI one: `show replication` over the CLI prints two different LAYOUTS and needs the broker type to pick a parser, but its SEMP reply models both as sibling structures under one `<mate>` -- an appliance fills `<connect-via>` with a `<connect-ports>` trio, software fills `<mate-list>` -- so whichever is populated says which broker it is and nothing is discovered. Also: a port of zero is an unconfigured transport, not port 0 |
| `TestSEMPMateReadsRolesTriState` | The reply enumerations decode including `n/a`, which must not fold into standby or a switch could promote against a VPN it cannot see. A VPN name with spaces survives (element text here, unlike the CLI's fixed-width table), the list narrows to what was asked about, and a value outside the enumeration fails loudly rather than being guessed at |
| `TestSEMPMatePrimaryActive` | The activity read, which is per VIRTUAL ROUTER rather than per node -- a chassis reports one activity for its primary virtual router and another for its backup. Standalone (`config-status` not Enabled) passes, being its own only node |
| `TestSEMPMateRejectsANonSEMPSite` | A site declaring `via.kubernetes` is refused here rather than silently driven over the wrong mechanism |
| `TestSEMPMateSetRoleRefusesNA` | `n/a` is not writable and never reaches the broker -- the same guard the CLI leg has, so the two channels cannot diverge on it |
| `TestSEMPMatePlainTextURL` | `tls: false` produces `http` on the declared port and sends no TLS options: the declared-not-inferred choice the site block exists to make explicit |

### switchplan_test.go

| Test | What it covers |
| --- | --- |
| `TestBuildSwitchPlanConverged` | A pair already matching the file produces an EMPTY plan and reports what it skipped. Empty is what stops a no-op run demanding an exact-`yes`: asking a destructive question about a run that destroys nothing teaches an operator to answer without reading |
| `TestBuildSwitchPlanMovesOneVPN` | The single-VPN failover, and the reason the plan skips rather than sweeping: the VPN that is not moving is never demoted, so its clients stay connected |
| `TestBuildSwitchPlanNeverDemotesTheTarget` | The demotion lands on the site the file does NOT name, so a VPN is never taken down at the site it is about to be promoted at. That is what makes the demotion safe to issue unconditionally |
| `TestBuildSwitchPlanPromoteNeverPrecedesDemote` | A property over the phase ORDER across every reachable role combination, not an example: demote, then confirm, then promote. The invariant the whole design rests on |
| `TestBuildSwitchPlanRefusesSplitBrain` | A VPN arriving ACTIVE AT BOTH sites stops the run rather than being silently repaired. Demoting one side picks a winner, and which site keeps its spooled messages is not a decision to make for the operator; the message names the VPN, both sites and which the file favours |
| `TestBuildSwitchPlanRefusesUnusableVPNs` | A switch only moves roles: a listed VPN missing at a site, one with replication shut down, and one reporting `n/a` each refuse, and the message points at `configure data-replication`, which is what creates and enables them |
| `TestBuildSwitchPlanRefusesAnUnknownActiveAt` | An `activeAt` naming neither site is caught here as well as at config load -- the plan is built from two site states a caller supplies, which could disagree with the file |
| `TestBuildSwitchPlanRefusesOneSiteTwice` | Both sides reporting the same site is refused, since every "the other site" decision would otherwise be meaningless |
| `TestSwitchPreflightHealthyPair` | The gate passes on a pair where both primaries hold activity and each broker names the other |
| `TestSwitchPreflightRequiresThePrimaryActive` | A switch runs only when each site is on its own PRIMARY HA node. The replication role is config-synced, so this is not about where a write lands -- it is about not stacking a cross-site role change on top of a local failover, and it is the honest failure for the case that actually bites: when the backup is active BECAUSE the primary is down, the primary is the node this tool addresses, so the run would fail anyway |
| `TestSwitchPreflightStandalonePasses` | A broker with no HA group is not refused for failing to hold activity it cannot hold -- it is its own only node |
| `TestSwitchPreflightRequiresMatchingVirtualRouterNames` | Each broker must name the OTHER site's declared `virtualRouterName` as its replication mate. A mismatch means the file names a site the broker has never heard of, `configure data-replication` has not been run since the block changed, or the tool is talking to a node whose virtual router is not the one the pair was built around |
| `TestSwitchPreflightRunsBeforeAnyWrite` | Every refusal happens with the pair untouched, which is the whole reason these are a preflight rather than a step |
| `TestExecuteSwitchPlanOrder` | The writes land in the only safe order: the demotion at one site, then the promotion at the other, with the read-backs between |
| `TestExecuteSwitchPlanStopsOnUnconfirmedDemotion` | A demotion that is ACCEPTED but not applied -- what `cli -Apes` exiting 0 on a rejected line looks like -- stops the run before any promotion, and the error says the VPN is standby at both sites: unavailable, not corrupted, so an operator does not go looking for data loss |
| `TestExecuteSwitchPlanStopsOnSetRoleError` | A rejected write stops the run with nothing promoted |
| `TestExecuteSwitchPlanRefusesToPromoteAgainstStaleState` | The window-closing test: something outside the run moves a role between the confirmation and the promotion, and the re-read immediately before promoting sees it and refuses. Promoting against an unread state is precisely how two actives happen |
| `TestExecuteSwitchPlanEmptyDoesNothing` | An empty plan writes nothing at all |

### transcript_test.go

What a `cli -Apes` capture has that a script does not: the CLI transcript around it. A
real export (2026-09-11) showed the login banner, the licence text and every echoed
prompt+command landing in stdout ahead of the `! BEGIN` header, and these tests pin the
three consequences. Two real banners live here as fixtures (`testBannerSoftware`,
`testBannerAppliance`, credential-free) and are what the importops tests can as the
target's banner; `testdata/currentconfig_remove_default.cli` is a real, credential-free
`show current-config message-vpn default remove`.

| Test | What it covers |
| --- | --- |
| `TestTranscriptBannerType` | `bannerType` reads software vs appliance off the login banner -- both real banners, plus each with its licence URL removed to prove the product-line fallback, plus no-banner and empty as `BrokerUnknown`. This replaced a `show version` scan whose expected string a real target never printed, so a valid first import was refused |
| `TestTranscriptStripCutsAtBegin` | `stripTranscript` cuts a capture at the broker's own `! BEGIN` line (which the broker emits with a leading space), is idempotent on already-clean input, and refuses output with no BEGIN -- the guard that keeps `xps-ps-01> home` from being replayed as a command, invisibly, since `cli -Apes` exits 0 |
| `TestTranscriptPreambleForApply` | an apply script keeps only the mode commands and the interpreter pragmas from a captured preamble: banner and echoed prompts are dropped, `! pragma:interpreter:relaxed-db-restore` is kept -- an earlier filter that dropped every `!` line to remove the header dropped that pragma with it |
| `TestTranscriptTeardownFixtureResetsNotDeletes` | the real `remove` capture parses, never negates a `message-vpn` opener or a `"default"`-named object, and DOES emit reset blocks for the default VPN, client-username, client-profile and acl-profile -- the property that lets import carry no reserved-name filter and no `default` branch |

### verify_local_test.go

The primary-driven HA operations the container platforms use: LocalRole detection,
LeaderLocal (with its mate revert step), and RedundancyCoordinated -- the one-invocation
failover exercise whose only cross-host touch is the SEMP revert-activity.

| Test | What it covers |
| --- | --- |
| `TestLocalRole` | Role detection: an explicit arg wins, host name matches the node table case-insensitively (including FQDN vs short name and a mixed-case host against a lower-case config), no match and bad args fail loud |
| `TestLocalRoleCaseInsensitiveConfigUpper` | The other half of the fold: an upper-case configured node name still matches a lower-case hostname |
| `TestLocalRoleAmbiguousHostname` | Two node-table entries sharing one configured name fail loud, naming every role the hostname matched, instead of silently returning the first |
| `TestLocalRoleExplicitArgSkipsAmbiguityCheck` | An explicit role argument wins outright and skips detection, even against a node table ambiguous enough that detection alone would fail |
| `TestDetectRoleByAddress` | Detection's SECOND pass: a host whose name matches nothing is identified by matching its own interface addresses against `redundancy.*.addr`, and `how` says which address. This is the cloud case -- an instance reporting `ip-10-0-0-12` while the env file names the broker `bkp-host`, so the routername and the OS hostname are legitimately unrelated |
| `TestDetectRoleNameBeatsAddress` | Order, not preference: a host named after the primary but numbered as the backup resolves to primary. The name is what the broker is called; an address can be shared by more than one interface |
| `TestDetectRoleAmbiguousAddress` | The address half of the ambiguity guard: one address configured for two roles fails loud naming both, and says the match was by address |
| `TestDetectRoleNoMatchNamesBothPasses` | The no-match error names the hostname, BOTH lookups (`redundancy.*.name` and `redundancy.*.addr`) and the roles that can be passed instead -- otherwise someone adds an `addr` to a file that already has one and cannot see why nothing changed |
| `TestDetectRoleAddrsError` | A failed interface read surfaces as its own error carrying the cause and the hostname, not as a bare "no match" that sends the operator to the node table |
| `TestDefaultLocalAddrs` | The unseamed path returns BARE addresses: `net.InterfaceAddrs` yields CIDRs, and leaving the `/24` on would make every `addr` match fail silently |
| `TestLocalAddrsFallsBackToTheDefault` | The seam's nil branch -- an `Ops` built without one reads the real machine, which is the production path |
| `TestLeaderLocalStandaloneSkips` | No calls in standalone mode |
| `TestLeaderLocalRejectsNonPrimary` | Backup and monitor hosts, and an explicit backup arg, are rejected before any upload |
| `TestLeaderLocalSuccess` | On the primary, the mate is preflighted and reverted (two SEMP calls), leadership is asserted and sync reported |
| `TestLeaderLocalRevertsMateFirst` | k8s Leader order: the mate revert POST precedes the first show-rd poll read |
| `TestLeaderLocalMateUnreachableWarnsAndContinues` | An unreachable mate SEMP downgrades to a warning and skips the POST; assert-leader still runs |
| `TestLeaderLocalMateRPCErrorFails` | A reachable mate rejecting the RPC is a real error, stopping before any poll |
| `TestLeaderLocalTimeoutDumpsDetail` | Unrecovered redundancy times out and dumps the detail |
| `TestRedundancyCoordinatedStandaloneSkips` | No calls in standalone mode |
| `TestRedundancyCoordinatedRejectsNonPrimary` | Backup and monitor hosts, and an explicit backup arg, are rejected before any call |
| `TestRedundancyCoordinatedBadRoleArg` | A bad explicit role argument propagates |
| `TestRedundancyCoordinatedInitialShowError` | An initial `show redundancy` error propagates |
| `TestRedundancyCoordinatedUnhealthyPrimary` | An unhealthy primary fails loud without touching the mate or releasing activity |
| `TestRedundancyCoordinatedMissingMateIP` | A missing redundancy.backup.addr refuses, named in the error, before anything is released |
| `TestRedundancyCoordinatedPreflightUnreachable` | The mate is reached BEFORE the first mutation: a failed preflight aborts with no release sent |
| `TestRedundancyCoordinatedSuccess` | Full run consumes the scripted sequence: release + un-release uploaded, exactly two SEMP calls (monitor GET, then the RPC as a data line on stdin), credentials and body never in argv |
| `TestRedundancyCoordinatedStandbyPrimaryStillReverts` | A standby-start skips the release walk but still reverts the mate and waits for activity to come home |
| `TestRedundancyCoordinatedNeverTargetsBackupRole` | The node-local-transport invariant: every transport call targets config.Primary -- config.Backup would silently re-read the local broker |
| `TestRedundancyCoordinatedNeitherActive` | No active mate observed after the release walk fails loud with only the preflight sent (no revert POST) |
| `TestRedundancyCoordinatedMateRPCNotOK` | A 2xx SEMP reply without code="ok" surfaces the mate's reply |
| `TestRedundancyCoordinatedRevertTimeout` | Activity never returning to the primary reports the bounded timeout (the revert was sent) |
| `TestRedundancyCoordinatedReleaseError` | A failed release exec stops before any show-rd poll (only the initial read happened) |
| `TestRedundancyCoordinatedReleasedTimeout` | A release that never converges times out and never sends no-release |
| `TestRedundancyCoordinatedNoReleaseError` | A failed no-release exec surfaces instead of proceeding to the un-released poll |
| `TestRedundancyCoordinatedUnreleasedTimeout` | An un-release that never converges times out without reverting the mate |
| `TestLocalRoleDefaultHostname` | With the seam unset, the real `os.Hostname` is used and an off-table host fails loud |
| `TestLocalRoleHostnameError` | LocalRole wraps and returns a Hostname read failure instead of matching a garbage host against the node table |
| `TestLeaderLocalBadRoleArg` | LeaderLocal propagates an invalid explicit role arg before the primary-only guard runs, making no transport calls |
| `TestLeaderLocalPollCondError` | LeaderLocal aborts on a mid-poll transport error while still dumping show-redundancy-detail, closing an asymmetry with the k8s Leader |
| `TestLeaderLocalAssertLeaderError` | LeaderLocal returns the assert-leader error after a healthy poll without showing output |
| `TestMateActivityStateReadsTheMateColumn` | `MateActivityState` reads the `Activity Status` line for exactly "Mate Active"; local-active, standby, empty output, and the phrase appearing under an unrelated label all read false |
| `TestShowRedundancyIsReadOnly` | The exported `ShowRedundancy` wrapper issues exactly one `show redundancy` CLI script on the named role and nothing else, since a probe calls it to decide whether a mutation is safe |

### appliance_test.go

What only a real APPLIANCE capture can settle. Every classification in `sections.go` was
originally reasoned from software captures, where the appliance-only sections are empty --
so `skip-appliance` was an assertion about a section nobody had seen carry anything.
`semp/show-currentconfig-all.out` (soltr/10_4_1, 9MB, 10 message-VPNs) turns that reasoning
into evidence, and found one section with no rule at all. Gitignored, so every test here
skips when it is absent.

| Test | Proves |
| --- | --- |
| `TestApplianceCaptureIsReadAsAnAppliance` | The pre-10.25.4 header form is discriminated correctly: `soltr/10_4_1` (appliance) versus `soltr/10_4_1VMR` (software) differ by three characters, and the entire cross-type refusal rests on reading them apart |
| `TestEveryApplianceBrokerSectionIsClassified` | No broker-region section on a real appliance is unclassified. This is the gap-finder that caught `Configure Hostname` -- empty in both software captures, carrying `hostname "<name>" defer` here, which is node identity like Configure Router Name. Broker region only: VPN-region blocks are rendered per VPN and deliberately never classified |
| `TestApplianceOnlySectionsCarryContent` | The `skip-appliance` rules are load bearing, not speculative: Ethernet, Lag, DNS, SNMP and `ip vrf` all carry real content on an appliance. `ip vrf` appears TWICE (management, msg-backbone), which is what validates matching that section by prefix rather than equality |
| `TestApplianceReplicationGrammarDiffersFromSoftware` | Records a platform difference at the SAME broker version: software writes `replication mate connect-via "host:port" "ssl"` as its own repeatable line, an appliance writes explicit `mate connect-port <n>` lines and hangs `connect-via` off the router-name line. Import is unaffected (cross-type is refused), but anything GENERATING replication config must render per type |

### driver_test.go

The generated apply script and the marker protocol it reports back through. Both halves
are pure, which matters: this script is what actually applies configuration to a
production broker, so there is no second code path a test could exercise instead.

| Test | Proves |
| --- | --- |
| `TestRenderDriverExactOutput` | The whole generated script for one chunk, matched EXACTLY rather than by Contains. It is shell that runs inside a production broker with the operator's configuration in it -- a stray quote, a lost `trap` or a heredoc that stops quoting are not things to find out about in production, and each changes this string |
| `TestRenderDriverQuotesChunkBodiesLiterally` | A body full of `$`, backticks and quotes reaches the broker byte for byte, and the heredoc delimiter is QUOTED -- an unquoted one would expand the configuration and silently apply something else |
| `TestRenderDriverRefusesABodyCarryingTheDelimiter` | A chunk containing the delimiter is refused: it would close its own heredoc and the remainder would execute as shell |
| `TestRenderDriverCleansUpEveryChunk` | The `trap` removes every chunk and the temp transcript on EXIT/INT/TERM, so an interrupted import leaves no credential-bearing configuration in `cliscripts` |
| `TestRenderDriverSanitisesChunkNamesInComments` | A newline in a chunk name (which is a message-VPN name, so operator-controlled) cannot escape its `#` comment and become code |
| `TestParseDriverOutputOutcomes` | The marker protocol and every way it can be wrong: all-applied, a failure with detail (legitimately no completion marker, since the driver exits there), output cut short, no markers at all, and a marker naming a chunk that was never sent. The last three are ERRORS -- read as success they would certify an import that did not happen |
| `TestParseDriverOutputDetectsGapsAndDuplicates` | What the terminator alone cannot catch: `SOLACE-UTIL-DONE` is not proof every chunk was reported, since a marker dropped or reordered in the MIDDLE still leaves the last result intact and the terminator present. Fewer results than chunks and a gap between reported chunks both error on the count mismatch; a chunk 0 marker reported twice (count matches, 2 results for 2 chunks) is the case that independently exercises the per-position check rather than the count check -- missing either would let a dropped OK for a TEARDOWN chunk omit a destroyed VPN from the report |
| `TestParseDriverOutputSuccessCarriesNoDetail` | A successful chunk's result carries no transcript text; only a failure may, because applying a configuration repeats every credential in it |
| `TestKeywordPatternEscapesEachPhrase` | Each failure phrase is regex-escaped before it reaches `grep -E`, so a keyword with a metacharacter cannot match the wrong thing |
| `TestShQuoteHandlesASingleQuote` | POSIX single-quoting of an embedded quote (close, escape, reopen) -- there is no escape character inside single quotes |

### transform_test.go

The replay transformations ported from `semp/CurrentConfigParser.html`, the operator's
hand-built migration GUI. Where a rule exists because of something in the REAL capture the
test uses the real capture (`semp/all.cli`, skipped when absent since it is gitignored),
because a hand-made input is exactly what missed these rules for so long.

| Test | Proves |
| --- | --- |
| `TestInjectVPNServiceShutdownCoversEveryRealPortLine` | Every VPN-level `service ... listen-port` line in the real capture is preceded by a shutdown. There were 48 and none was handled: they are indented CHILDREN of one `message-vpn` block, not blocks, so no block-level rule reached them -- and the broker refuses a port change while the transport is up while `cli -Apes` exits 0 when it does |
| `TestInjectVPNServiceShutdownFormPerTransport` | The exact shutdown per shape, including `web` -> websocket and `ssl web` -> websocket-secure. Granularity is the difficulty: mqtt carries four transports at once, so a blanket `service mqtt shutdown` would drop three to change one |
| `TestInjectVPNServiceShutdownLeavesColumnZeroAlone` | Broker-level port lines stay `InjectShutdown`'s, so no line gets two shutdowns from two rules |
| `TestInjectVPNServiceShutdownSkipsAnExistingGuard` | A capture that already shuts the transport down immediately before its port change gets nothing added |
| `TestShutdownDefaultClientUsername` | The `default` client-username gets a `shutdown` after its opener and keeps the artifact's own trailing state. It cannot be deleted, so it is edited in place and the broker refuses edits while it is enabled |
| `TestShutdownBridges` | Every `no shutdown` inside a bridge becomes `shutdown`, at both depths, indentation preserved. A bridge that comes up starts moving messages between two brokers before anyone has read the result |
| `TestRewriteAdminUsername` | `create username "admin"` becomes `username admin` -- admin always exists, so `create` is a duplicate and every attribute under it is lost with the block |
| `TestRewriteAdminUsernameFindsTheOpenerPastAPragma` | The rewrite targets `Block.Opener()`, not `Lines[0]`: a held `! pragma:` line is prepended, and rewriting index 0 blindly would corrupt the pragma and leave the `create` |
| `TestReorderVPNProfilesPutsProfilesLast` | client-profile/acl-profile/client-username move to the end of their own VPN and never cross a VPN boundary -- a username naming a profile the broker does not yet have is refused |
| `TestCollapseDuplicateLines` | Consecutive duplicates collapse; blank lines are exempt, since collapsing them only reflows the artifact |
| `TestClearExistingSyslogsOnlyForEntriesTheTargetHas` | `no syslog` is emitted only for a name the TARGET has. Both halves matter and pull opposite ways: `create syslog` for an existing entry is rejected, but `no syslog` for a missing one is itself an error |
| `TestClearExistingSyslogsAgainstTheRealCapture` | The three facts the syslog rule rests on, all guesses until a capture carried one: the section header carries the destination name (so the rule matches by PREFIX, not equality), the section is CLASSIFIED rather than falling through as unclassified (an unclassified one is never applied, so the removal would be emitted and the create would not), and the operand is quoted -- which is what the emitted removal mirrors |
| `TestClearExistingSyslogsNoTargetEntriesIsANoOp` | A target with no syslog configured -- the common case -- gets nothing added |
| `TestSyslogNamesReadsThroughTheQuoteAwareParser` | The target's syslog names come from parsed blocks, so a name with spaces survives -- the same failure that once filed an existing message-VPN as new |
| `TestNestedObjectsAreFoundDespiteTheParentKind` | The discovery the nested-object rules rest on: a client CA is not a column-0 block -- the section opens `authentication` and creates inside it, so `Block.Kind` is the PARENT and every Kind-keyed rule misses it, the same shape that hid the VPN-level service lines. Domain CAs (`ssl`) and virtual hostnames (`service`) are identical. Also asserts both sections are CLASSIFIED, since an unclassified one is never applied |
| `TestClearExistingNestedCoversAllThreeVerbs` | Three sections shipped applied-but-UNGUARDED: a `create` for a name the target already has is rejected, no capture emits a removal, and two carried a Reason claiming a delete-then-create that existed nowhere. Every such create is now immediately preceded by its own removal at the same indent -- the only placement correct at every depth, which the indent-6 LDAP group forced. The `no <verb> <name>` form is VERIFIED against Solace's CLI reference, which documents each as `[create \| no] <verb-name>` |
| `TestClearExistingClientCAsIntersects` | A client CA is cleared only when the target HAS it AND the artifact re-creates it. Without the target half the removal errors on an absent CA; without the artifact half the import deletes trust material the target's own operator added -- and a client CA decides which clients the broker accepts |
| `TestClearTargetVirtualHostnamesCoversBothHalves` | The OPPOSITE rule to the client CAs on an identical shape -- every virtual hostname the target has is cleared, artifact or not, because a stale DNS name pointing at this broker keeps answering for a name the imported configuration never claimed -- and the division of labour that delivers it. A name the artifact re-creates is cleared immediately before that create (the only placement correct at every nesting depth); a name with no create is rendered by `TargetOnlyRemovals` opening the `service` scope itself. Covers both together, since neither half alone is the property, and asserts an in-artifact name is not ALSO cleared as an orphan |
| `TestShutdownBridgesLinesAreExcusedByTheDiff` | Regression for the worst bug the chunked apply shipped with: EVERY artifact containing a bridge failed verification after a perfectly CLEAN apply. `ShutdownBridges` is the one transform that changes a line rather than adding one, and `ImportVerify` diffs the untransformed artifact directionally -- so it demanded a `no shutdown` the import had deliberately applied the opposite of. Also pins that the exemption is narrow and does not leak to non-bridge blocks |
| `TestTargetOnlyRemovalsWhenTheArtifactCreatesNone` | `clearNestedObjects` can only place a removal inside a block that CREATES that verb, so an artifact with no virtual hostnames had no such block and the clear silently did nothing -- precisely when a target carrying names the artifact does not has to be cleared. These removals open their own `service` scope, and a name the artifact re-creates is left to the block-level clear so it is not removed twice |
| `TestPrepareForApplyIsIdempotentOnTheRealCapture` | Running the whole pipeline twice adds no lines, per clear as well as in total. Several rules INSERT, and re-running a failed import is the documented recovery -- a transform that grew the script each pass would apply something different the second time. A populated TARGET is not enough to exercise the clears and believing it was left them dead for a while: every clear attaches its removal to a CREATE in the ARTIFACT, so both halves are taken from the capture that carries the object. The guard that each clear actually INSERTED is what stops the setup going quietly vacuous again; a total line count cannot stand in for it, because the transforms that ignore the target make the total grow either way |

---

## internal/k8s

Everything driven through `kubectl`: the read-only permission preflight, prep, deploy,
operator, day-2 ops, secrets, and the pod transport, plus the operator's watch-list
algebra and the namespace occupancy gate, plus the mate channel that reaches a
replication site in another cluster, and the Secret read that supplies a mate's
password. 200 tests across 18 files.

### matechannel_test.go

| Test | What it covers |
| --- | --- |
| `TestNewMateChannelArgv` | The exact kubectl argv, which is why this lives in `internal/k8s` at all (`internal/k8s` already imports `internal/broker`, so the reverse would be an import cycle): the SITE's own command, namespace and broker name; the pod named with the role LETTER (`-p-0`, not `-primary-0`, which names nothing); the script uploaded on stdin and then run by its RELATIVE name because the CLI's working directory is the scripts folder; and `no paging` after `home` |
| `TestNewMateChannelUsesTheSiteNotTheLocalDeployment` | The mistake the design exists to prevent: reaching a DR mate with THIS env file's namespace and broker name would quietly drive the local broker while reporting on the remote one |
| `TestNewMateChannelRemovesTheScriptItUploaded` | The leak this leg shipped with: the LOCAL leg has always deleted its uploaded scripts, this one did not, so every DR command over kubectl left a `.cli` file inside a broker belonging to the OTHER site -- the one place this tool writes into somebody else's cluster. Both subtests check the cleanup is last, names the script's absolute in-broker path, and goes to the SITE's namespace; the failed-run case is the one that matters, since a rejected run is exactly the one an operator re-runs |
| `TestPodNameForIsTheOneSpelling` | The mate channel names its pod through the shared rule rather than rebuilding it. It could not call `podName`, which takes a whole config, because a DR site is described by its own `via.kubernetes` block -- so it spelled the naming rule out by hand, which is how two spellings of one rule drift apart. Also pins that `podName` stays a wrapper rather than becoming a second spelling |
| `TestNewMateChannelGuardsTheSiteCommand` | The execution guard runs when the channel is BUILT, not only in `Validate` -- it is constructed straight from a `*config.Config` and must not assume `Validate` ever ran, the same reason `Cluster.cmd` re-checks its own command on every call |
| `TestNewMateChannelRefusesANonKubeSite` | A site declaring `via.semp` is refused here rather than silently driven over the wrong mechanism |
| `TestReadSecretKeyArgvAndDecoding` | The mate-password read: the SITE's own cluster CLI, one key fetched by jsonpath rather than `-o yaml` (which would pull every other key of the Secret through this process), and the value base64-decoded with the trailing newline dropped -- the one `kubectl create secret --from-file` keeps and `echo` adds, which would otherwise be sent as part of the password |
| `TestReadSecretKeyEscapesADottedKey` | The silent-failure case: an unescaped dot in a jsonpath member is read as a PATH STEP, so `{.data.tls.key}` looks inside a `tls` object that does not exist and kubectl exits 0 printing nothing |
| `TestReadSecretKeyFallsBackToThisClusterCLI` | A SEMP-reached site can still keep its password in a cluster: with no `via.kubernetes` to borrow a command from, the read uses THIS env file's `kubernetes.runtime` |
| `TestReadSecretKeyRefusesAnEmptyReading` | kubectl exits 0 and prints nothing for a key the Secret does not carry, so without this the mate would be dialled with an empty password and the failure would surface as a login refusal against a perfectly healthy broker |
| `TestReadSecretKeyRefusesNonBase64` | `.data` is always base64 and `.stringData` is write-only, so a value that does not decode means the wrong field was read -- refused rather than decoded to garbage |
| `TestReadSecretKeyGuardsTheCommand` | The execution guard runs on this path too, and stops BEFORE exec: the read is built straight from a `*config.Config` and must not assume `Validate` ever ran |

### names_test.go

| Test | What it covers |
| --- | --- |
| `TestResourceNames` | Pod, PVC, StatefulSet, and load-balancer service names for every role |
| `TestRestartOrder` | The safe manual-bounce order (monitor, backup, primary; standalone just the primary) |
| `TestHARoles` | HA yields all three roles; standalone yields only the primary |
| `TestProductKeyRoles` | Product keys target primary+backup in HA, primary only in standalone |
| `TestPodNameSuffixMatchesTheConfigBound` | The drift guard between two packages that deliberately do not import each other. `internal/config` bounds `kubernetes.name` by the length of the operator's longest suffix and spells that suffix out rather than importing it -- config is the leaf every platform package reads, so the dependency cannot go the other way. This derives a real pod name and checks the length config assumes, so a change to `brokerSuffix` or the role letters fails here instead of silently loosening a bound |

### watch_test.go

The operator's watch scope is the one field of the install that does not belong to any single
env file, so these are about a shared resource edited by several callers rather than about
string handling. The bug they exist to prevent had no error and no log line: `operator deploy`
from a second env file overwrote `WATCH_NAMESPACE`, the first broker's namespace fell out of
it, and that broker simply stopped reconciling.

| Test | What it covers |
| --- | --- |
| `TestWatchFromContainersReadsEveryAllNamespacesSpelling` | The three ways of saying "watch everything" -- no `WATCH_NAMESPACE` entry, an empty value, a value that is only separators -- are ONE answer. Every caller either leaves an all-namespaces operator alone or refuses to narrow it, so collapsing them is what makes those callers correct |
| `TestDesiredWatchAppendsTheBrokerNamespace` | The config half: the broker namespace is appended unless `watchBrokerNs: false`, entries are trimmed and de-duplicated, and the only route to "watch everything" from config takes saying so twice (no `watchNamespaces` AND `watchBrokerNs: false`) |
| `TestUnionWatchPreservesInstalledOrder` | Not cosmetic: the list becomes one env-var value on a pod template, so a reordered but equivalent list is a CHANGED template and rolls the operator pod for nothing. An unchanged set must render byte-identically |
| `TestUnionWatchKeepsAnotherEnvFilesNamespace` | The point of the union as a scenario rather than set algebra: two env files, two namespaces, one operator -- deploying the second broker must not stop the first being reconciled |
| `TestSubtractWatchEmptyRemainingMeansDelete` | The case the caller MUST NOT apply. An empty remainder is not "watch nothing"; an empty `WATCH_NAMESPACE` means watch EVERY namespace, so applying it as a removal would be a privilege escalation dressed as a teardown. The caller has to read empty as "delete the operator" |
| `TestSubtractWatchPreservesOrder` | The same rolling-the-pod-for-nothing concern, on the removal path |
| `TestReconcileWatchDecidesWhatToApply` | Every branch of the deploy-side decision: first install, an operator already watching everything, a union that adds nothing, and one that adds our namespace |
| `TestReconcileWatchFlagsAWideningAsAQuestion` | Going from a named list to every namespace hands a shared operator the whole cluster, so the plan marks it for confirmation rather than applying it as a side effect of a deploy |
| `TestOperatorReleaseNarrowsInsteadOfRemoving` | Why `operator remove` is not just a delete: with another namespace still in the watch list the operator is KEPT and the list narrowed, via `set env` rather than a bundle re-apply -- a removal has no business changing the operator image another env file depends on |
| `TestSetWatchRefusesAnEmptyList` | The trap the design exists around, guarded at the writing function as well as the call site: writing an empty `WATCH_NAMESPACE` would leave the operator with WIDER scope than it started with |

### namespace_test.go

Deleting a Namespace cascades to everything in it, including whatever another team put there
-- and this tool very often did not create it. Every test here is about the DIRECTION of a
wrong answer: a namespace wrongly reported empty is cascade-deleted and unrecoverable, while
one wrongly reported occupied merely stays. The gate is built to fail toward "occupied".

| Test | What it covers |
| --- | --- |
| `TestNamespaceContentsDiscountsWhatKubernetesPutsThere` | Without this the prompt could never appear at all: Kubernetes puts a `default` ServiceAccount and a `kube-root-ca.crt` ConfigMap in every namespace, so counting either as occupancy means no namespace is ever empty -- and it would look like it worked, because "kept" is the safe outcome |
| `TestNamespaceContentsIgnoresClusterPolicyObjects` | A LimitRange or ResourceQuota is stamped on by cluster policy without anyone asking, so refusing on one would mean never removing a namespace on a cluster that sets quotas |
| `TestNamespaceContentsReportsRealOccupants` | The keep branch, including the group-qualified spelling `-o name` sometimes returns (`secrets.v1./x`) -- the suffix must not defeat the match |
| `TestNamespaceContentsErrorMeansOccupied` | Being unable to see what is in a namespace is not permission to delete it; the error names the namespace and says it survived |
| `TestRetainedPVCsKeepTheNamespace` | The hazard made structurally impossible rather than documented: without `--delete-data` the PVCs remain, so the namespace is not empty, so it is not offered. "Delete the namespace and take the retained data with it" cannot be reached by answering one prompt wrong |
| `TestNamespaceContentsAsksOneQuestion` | A fixed kind list in ONE call with `--ignore-not-found`. The `api-resources` alternative would see CRD kinds too, but needs `list` on every kind in the cluster, so a namespace-scoped identity fails the probe rather than answering it |
| `TestNamespaceIsProtectedSharesDeleteNamespacesList` | The four cluster namespaces are refused before the question is asked, from the SAME list `DeleteNamespace` uses -- two lists would be two things to keep in step |
| `TestReportNamespaceOccupiedNamesTheObjects` | The branch an operator sees most often lists what is in the way; "3 objects remain" gives nobody anything to act on |


### runtime_test.go

| Test | What it covers |
| --- | --- |
| `TestClusterHonoursRuntime` | Every `Cluster` helper (`kubectl`, `apply`, `deleteStdin`, `output`, `interactiveExec`) runs argv[0] from `kubernetes.runtime` and places its leading arguments ahead of the subcommand |
| `TestTransportHonoursRuntime` | The pod transport does the same for `exec`, `exec -i`, the stdin `Upload`, and both `cp` directions |
| `TestExecutorRefusesUnapprovedRuntime` | The executor half of enforce-twice: a `Cluster` and a transport built straight from a `*config.Config` that never went through `config.Load` still refuse an unapproved `microk8s kubectl`, on all eleven paths that could reach exec -- and hand the runner nothing at all, since refusing after the call would mean the binary already ran |
| `TestRuntimeDefaultArgvUnchanged` | With the default runtime the argv is exactly `kubectl ...` with no extra tokens -- the regression guard for every existing `+ kubectl ...` assertion |

### cluster_test.go

| Test | What it covers |
| --- | --- |
| `TestOperatorNSExplicit` | A configured operator namespace is used without probing the cluster |
| `TestOperatorNSNeverProbesTheCluster` | H1: table-driven over an operator running elsewhere, an empty cluster, and an unreachable one -- all three resolve to the fixed default and make ZERO cluster calls. Replaces the deleted three-way discovery tests (derived / default-when-absent / default-on-error): resolution used to list Deployments in every namespace and take the first line CONTAINING the operator's name, an unanchored substring match with no uniqueness check that could resolve to another team's operator on a cluster running two installs. Now there are two local rules only (`kubernetes.operator.namespace`, else the fixed default), so nothing outside the env file can steer it |
| `TestApplyOnStdin` | `apply` pipes the manifest on stdin via `kubectl apply -f -` |
| `TestDeleteStdin` | `deleteStdin` pipes the manifest with `--ignore-not-found` |

### check_test.go

`check.go` itself now owns only the reachability probe and the storage-class resolution;
the grouped report (secret-free config echo, the sparse-config wording, the skip notes
under a preview) moved to `checkreport.go` and is covered in `checkreport_test.go` below --
`TestCheckEnvNoSecretLeak`, `TestCheckStorageClass`, `TestCheckDryRun` and
`TestCheckEnvSparseConfig` were removed from here for exactly that reason, and their
coverage is named again on the successor row that replaced each one.

| Test | What it covers |
| --- | --- |
| `TestReachable` | The API-server probe argv, and failure when it errors |
| `TestCheckStopsProbingWhenUnreachable` | Successor to `TestCheckAbortsWhenUnreachable`: Check still fails when the API server is unreachable, but now as a `[FAIL] api server` ROW in the report (naming "cannot reach") rather than only a returned error, every check that needs the cluster reports `[SKIP]` instead of being silently omitted or repeating the same failure, and no further probe runs after the one Reachable call |
| `TestResolveStorageClass` | A configured class short-circuits; a single default resolves; multiple defaults error; no default returns empty |

### checkreport_test.go

The grouped `validate` report (`checkreport.go`) that replaced the old flat
`CheckEnv`/`CheckStorageClass` passthrough: one `checkReport` with a fixed section order,
built and rendered even when an earlier section failed, so one run names every problem
instead of the first.

| Test | What it covers |
| --- | --- |
| `TestValidateNeverPrintsASecret` | Successor to `TestCheckEnvNoSecretLeak`: the admin, monitor and registry passwords never reach the report, which shows only `admin=set`/`monitor=set` |
| `TestValidateReadsDeploymentsOnce` | The fetch-once rule: the watch row and the operator rows want two different facts about the SAME operator Deployment, and each used to issue its own cluster-wide `get deployment --all-namespaces`, so every healthy `validate` paid for a full list twice. It asserts EXACTLY one, not at most one -- at most would also pass if the fetch stopped happening, so a report that silently skipped both rows would read as a successful optimisation. The verdict is ignored on purpose: against a fake that answers nothing some rows legitimately come back failed, and the call count is the same either way |
| `TestValidateGroupsAndOrdersSections` | Pins the layout: sections render in the agreed order (Deployment, Operator, Broker, Credentials, Placement), the four leading rows (namespace, name, image, image pull) lead in order, and cpu/mem are never crammed into one compound row |
| `TestValidateTagsEchoedConfigAsINFO` | Successor to `TestCheckDryRun`: config read back from the env file is tagged `[INFO]`, never `[ OK ]` (a report that tagged unverified config as OK would claim verification nobody did), and a preview reports every cluster-dependent check as `[SKIP]` |
| `TestValidateSparseConfigExplainsItself` | Successor to `TestCheckEnvSparseConfig`: an empty watch list explains itself as "watches ALL namespaces" rather than the reassuring opposite, and unset TLS/admin password read `(not configured)`/`MISSING` |
| `TestValidateReportsEveryFailureInOneRun` | The behaviour change worth having: a refused permission is the first failure, but the run continues and still renders every section after it, and the returned error counts the failures rather than wrapping only the first |
| `TestCheckReportFailedCounts` | `checkReport.failed()` sums the FAIL rows across every section; an empty report counts zero |
| `TestStorageRows` | Successor to `TestCheckStorageClass`, exercising the `storageRows` section builder directly: a suitable configured class is OK with no default lookup; Immediate binding or no expansion is FAIL; missing attributes report `<none>` and FAIL; the actionable message names the fix without the old bash script number; and every read failure along the way (default resolution, the first attribute column, the second after the first succeeds) surfaces its own FAIL row rather than being swallowed |
| `TestValidateReportsResolvedPorts` | M11: an explicit `kubernetes.ports` resolves into the new leading Config section as a `ports (N)` count plus the `name=port` pairs, readable straight off the report with no need to deploy or run `broker generate` |
| `TestValidateReportsDefaultPorts` | M11: leaving `kubernetes.ports` unset is not "no ports configured" -- `ApplyDefaults`'s 17-entry default is what actually renders into the broker CR, so the Config section shows `ports (17)` and both the first and last default port; a precondition assertion fails loud if defaulting itself ever stopped happening, rather than passing vacuously on an empty list |
| `TestPortRowsNeverFail` | M11: `portRows` produces only `[ OK ]` rows at every wrap boundary (0, 1, a full chunk, one over, and the 17-port default) -- `validate` is read-only and must never stop a deploy over how many ports there are to print |
| `TestValidateConfigSectionNeverFails` | M11: the report-level companion to `TestPortRowsNeverFail` -- `validationRows` alone must never move `checkReport.failed()` off zero |
| `TestValidateReportsThePreSharedKeyChoice` | On Kubernetes an empty pre-shared key is a legitimate deployment rather than a gap -- the operator generates and distributes one -- so the report has to say WHICH of the two keys the group will end up using; nothing else shows it, since neither the CR nor the Secret exists until deploy. A configured key is reported by naming the Secret entry it becomes and never the value itself, and a standalone broker gets no row at all |
| `TestValidateSaysWhoOwnsTheTLSSecret` | Naming a Secret and supplying the files it is built from are separate decisions, and the report is the only place the difference shows before a deploy: read as "this tool will create it" in the case where it will not, a missing Secret is first discovered by a pod that will not mount. Covers both origins, plus the `tls.certPassphrase` warning -- the CRD's `spec.tls` has no passphrase field, so an encrypted key cannot be used on this platform -- and that the passphrase itself never reaches the output |
| `TestValidateSkipsTheStorageClassWhenEveryNodeIsCustomMounted` | Closes a gap the shared name hid: there are two `storageRows`, and only the config-side one branched on custom mounts. It drives a recording runner rather than `engine.Echo` on purpose -- the Echo path short-circuits every cluster-backed row to "skipped (preview)" before `storageRows` is reached, so a preview cannot see this either way, and a test written over Echo asserts nothing. The cluster-side namesake ran the class check regardless -- and since `validateStorage` refuses class and `customVolumeMount` together, it always fell through to the cluster default. On a cluster with no default (bare metal, static provisioning -- exactly why someone pre-creates PVCs) that FAILED the whole report over a class the deploy never touches, and advised setting a key `config.Load` then refuses. Also guards that dead-end advice against coming back |
| `TestValidateStillChecksTheClassWhenOnlySomeNodesAreMounted` | The guard asks per role rather than off `UsesCustomMounts`, so a partly covered group -- which still has claims to bind -- keeps being checked. `validateStorage` forbids that state today; this is what stops the check silently disappearing if the rule is ever relaxed |
| `TestCheckReportSkipsEmptySections` | A section with no rows (e.g. Placement on a standalone config) prints no header at all, rather than an empty block |

### prep_test.go

| Test | What it covers |
| --- | --- |
| `TestCreateNamespace` | The namespace manifest is applied on stdin |
| `TestDeleteNamespace` | The ordinary path: the permission probe passes and the delete is issued directly, with no enumeration of the namespace's contents here -- that is `NamespaceContents`' job, upstream of this call (`internal/cli`'s `removeNamespaceIfEmpty`) |
| `TestDeleteNamespaceProtected` | Pins `protectedNamespaces`' floor: none of the four Kubernetes system namespaces (`default`, `kube-system`, `kube-public`, `kube-node-lease`) is ever deleted, and -- unlike every other guard in `DeleteNamespace` -- the refusal happens before the cluster is asked anything at all, RBAC probe included (zero calls made) |
| `TestDeleteNamespaceStopsOnPreflightFailure` | A refused `auth can-i delete namespaces` stops DeleteNamespace before any `kubectl delete` is issued -- the same shape as `TestCreateSecretsStopsOnPreflightFailure`, but for the namespace teardown |
| `TestCreateSecretsAdminOnly` | With no TLS or pull secret only the admin secret is applied, as a single document |
| `TestCreateSecretsAllThree` | Admin + TLS + pull secret join into one multi-doc apply, and the registry password reaches neither argv nor plaintext stdin |
| `TestCreateSecretsPreflight` | Missing TLS inputs fail before any apply runs |
| `TestDeleteSecrets` | All configured secrets are deleted; admin-only config deletes one |
| `TestUpdateServerCertSecret` | The TLS secret is applied on stdin; an unset secret name errors |
| `TestCreateSecretsFailsWithoutAdminFields` | CreateSecrets can pass secretPreflight (TLS-only) and still fail loud inside GenSecrets when semp.adminPass/kubernetes.adminSecret are unset, with zero applies made |
| `TestCreateSecretsStopsOnPreflightFailure` | A refused `auth can-i create secrets` stops CreateSecrets before GenSecrets reads the TLS private key off disk -- loading key material for a cluster that will not accept it is work worth not doing |
| `TestGenBrokerLeadsWithTheNamespace` | The gap this closed: the stream used to start at the Secrets, so piping it at an empty cluster failed -- the Secrets and the CR are namespaced and the namespace was not there yet. `operator generate` had carried its own Namespace document from the start; this half had not. Order is asserted as offsets, since that is the whole property |
| `TestGenBrokerMatchesWhatDeployApplies` | Walks the deploy for real over the echo seam and asserts every manifest it applies is a document of the generated stream, and that there are exactly three. This is what catches the NEXT divergence: a step added to the deploy and not to the stream fails here rather than in someone's `kubectl apply` |
| `TestGenSecretsTLSError` | GenSecrets itself (not just via CreateSecrets' preflight) fails when kubernetes.tlsServerSecret is set but the cert files are unreadable, guarding `broker generate` |
| `TestDeleteSecretsStopsOnPreflightFailure` | A refused `auth can-i delete secrets` stops DeleteSecrets before any `kubectl delete secret` is issued -- the same shape as `TestCreateSecretsStopsOnPreflightFailure`, but for the secret teardown |
| `TestDeleteSecretsSkipsUnconfiguredAdminSecret` | DeleteSecrets never issues `kubectl delete secret ""` when kubernetes.adminSecret was never configured |
| `TestDeleteSecretsStopsOnError` | A genuine delete failure stops the teardown loop and surfaces instead of silently continuing to the remaining secrets |
| `TestCreateNamespaceApplyFails` | A failing apply (RBAC denial) surfaces from CreateNamespace instead of being silently swallowed |

### preflight_test.go

| Test | What it covers |
| --- | --- |
| `TestCanIAnswerReadsTheLastLine` | The verdict is the LAST non-empty line, not the whole output: `kubectl auth can-i` prints advisory lines above it on stdout ("Warning: resource 'x' is not namespace scoped"), and comparing the whole body would turn every such cluster into the unreadable-answer branch -- a preflight failing safe in the wrong direction. Covers plain yes/no, no trailing newline, one and several warnings, blank lines, CRLF, empty, and whitespace-only |
| `TestPreflightAcceptsAWarnedYes` | The end-to-end of the above: a cluster that warns and then permits lets the deploy proceed |
| `TestPreflightRefusesAnUnreadableAnswer` | Exit 0 with neither yes nor no -- a wrapper that swallowed stdout -- is refused rather than assumed permitted, since proceeding would act on a permission nobody confirmed |
| `TestCanIArgsScopesByResourceKind` | A cluster-scoped probe carries `--all-namespaces`, a namespaced one carries `-n`. Regression for a warning every real `broker remove` printed: omitting `-n` was already correct and not enough, because kubectl falls back to the KUBECONFIG context's namespace -- which put a namespace in the access review, where RBAC also evaluates namespaced Roles and could answer YES for a cluster-scoped action the identity cannot perform |
| `TestPreflightIsPreviewableUnderEcho` | The Echo runner records the probe and skips its assertion, so previewing needs no cluster -- which is why there is no skip flag |

### deploy_test.go

| Test | What it covers |
| --- | --- |
| `TestDeployBrokerApply` | One apply on stdin carrying the rendered CR |
| `TestDeployBrokerKeepYAML` | `--keep-yaml` writes `.broker.yaml` byte-identical to what was applied |
| `TestDeleteBrokerNoPurge` | Without purge only the CR is deleted, no PVCs |
| `TestDeleteBrokerPurgeHA` | Purge deletes the CR plus all three role PVCs |
| `TestDeleteBrokerPurgeStandalone` | Purge on standalone deletes the CR plus the single PVC |
| `TestDeleteBrokerPurgeSwallowsPVCError` | A failing PVC delete is real (an RBAC denial or a stuck finalizer survives `--ignore-not-found`) and is now reported as an error naming every PVC that survived, never swallowed; every role is still attempted before the error is returned, and the success line is not logged |
| `TestDeleteBrokerPurgeAllSucceed` | When every PVC delete succeeds, DeleteBroker issues one delete per HA role, returns nil, and logs the "PVCs deleted" success line |
| `TestDeleteBrokerNoPurgeIssuesNoPVCDeletes` | `purge=false` touches no PVCs at all -- only the CR delete runs, and the "PVCs kept" line is logged |
| `TestDeleteBrokerLogsPVCOutcome` | DeleteBroker states which layer survived rather than leaving it to be inferred from the argv: `purge=false` logs "PVCs kept" naming `--delete-data`, with no per-PVC line; `purge=true` logs "deleting PVC \<name\>" for every role plus a final "PVCs deleted" |
| `TestDeployBrokerKeepYAMLWriteError` | DeployBroker fails loud and never applies the manifest when writing .broker.yaml fails, instead of silently proceeding as if the file were saved |
| `TestDeployBrokerStopsOnPreflightFailure` | The preflight ordering guarantee: when `auth can-i` answers no, no manifest is written and no call follows the probe -- without this the probe would be decoration |
| `TestDeleteBrokerProbesClaimsOnlyWhenPurging` | Both halves of the claim probe: a removal that will delete PersistentVolumeClaims asks permission to, and one that keeps them does not demand a permission it never uses |
| `TestConfirmBrokerAppliedAcceptsTheObjectItApplied` | The read-back's happy path: the cluster returns the CR under the configured name and the deploy reports it. `brokerJSON(name)` is the shared reply fixture |
| `TestConfirmBrokerAppliedTreatsSilenceAsSkipped` | The distinction the whole function rests on. A runner that answers with NOTHING has not said the broker is absent -- it has said nothing, which is the preview case and every test seam. Turning that into the alarming answer would fail ~200 wiring tests and every dry run, so it reports `skipped` rather than passing silently |
| `TestConfirmBrokerAppliedFailsWhenTheObjectIsNotThere` | The case the read-back exists for: the apply was accepted and the object is not in the cluster. It is a `[FAIL]`, not a warning -- otherwise "deployed" stops meaning anything -- and the message names both likely causes (an admission webhook, an unestablished CRD), neither of which is visible from the apply's own output |
| `TestConfirmBrokerAppliedFailsOnADifferentObject` | A reply carrying something else is not the broker being confirmed, so it cannot stand in for it |
| `TestConfirmBrokerAppliedSurfacesReadFailures` | Both ways the read itself can fail -- the cluster refusing it, and an answer that will not decode. Each names the next step (`broker validate`), because the deploy has already happened and the operator needs to know what to check rather than what broke in here |
| `TestDeleteDataKeepsEveryCustomMountedClaim` | With every node custom-mounted there is nothing for `--delete-data` to delete: no `pvc` call is issued at all, and the report says so outright rather than printing the ordinary "data is gone" line over a full set of surviving volumes. Each surviving claim is named |
| `TestDeleteDataDeletesOnlyTheProvisionedClaims` | The mixed case: the operator's own claims go, the custom-mounted one stays, and the closing line says which. Someone who passed `--delete-data` and got a surviving volume would otherwise assume the flag failed |
| `TestPreflightUnreachableClusterHints` | An unreachable API server is a different failure from an RBAC refusal and gets the hint that helps (`log in first`, `oc login`), carrying kubectl's own error rather than replacing it; still nothing runs after the probe |

### operator_test.go

| Test | What it covers |
| --- | --- |
| `TestWatchNamespace` | `WATCH_NAMESPACE` joins: broker namespace appended by default, onto a configured list, or omitted when disabled -- plus the dedupe half, since a list that already named the broker namespace (the common case, `watchBrokerNs` defaults on) listed it twice in the report and the applied Deployment. Entries are trimmed, empties and trailing commas dropped, repeats inside the list collapsed, first occurrence winning. controller-runtime's map-keyed cache hid the repeat at runtime, so only these cases can catch a regression |
| `TestOperatorImage` | The registry-prefix rule now shared by `RenderOperator` and `CheckEnv`: prefixed when `image.registry` is set, raw when it is not. Its own test rather than only being reached through the 119 KB bundle render, because the report and the apply drifted for exactly as long as each owned a copy |
| `TestRenderOperatorSubstitutions` | Every substitution point lands (namespace, watch list, image with/without registry prefix, resources, pull-secret reference) and no template marker survives -- and the bundle carries no `.dockerconfigjson`, since the reference is a name and the Secret is its own artifact |
| `TestRenderOperatorHonoursThePassedWatchList` | Why the watch list is a PARAMETER rather than something RenderOperator derives. `operator deploy` reconciles this env file against what is installed and applies the UNION, which cannot be recovered from config -- a renderer that recomputed it would discard the union and re-narrow the operator on every deploy. The list passed is deliberately unrelated to the config, so the test fails if the parameter is ignored |
| `TestGenOperator` | Render-only uses the configured operator namespace, or the fixed default when unset |
| `TestOperatorApply` | The three applies in dependency order: the bundle's own Namespace document alone, then the regcred into it, then the rest of the bundle -- each `apply -f -` on stdin. The bundle half carries no `.dockerconfigjson`, so everything applied is byte-for-byte a `generate` output |
| `TestOperatorApplyNoPullSecret` | With no pull secret the regcred apply is skipped entirely: namespace then bundle, and nothing applied mentions regcred |
| `TestOperatorDelete` | With `deleteCRDs=false`, teardown deletes only the non-CRD documents (Deployment, RBAC, ...) with `--ignore-not-found`, and the piped manifest carries no `kind: CustomResourceDefinition` document at all |
| `TestOperatorDeleteWithCRDs` | With `deleteCRDs=true` and no broker CRs found (M3's guard, satisfied here), OperatorDelete lists PubSubPlusEventBroker CRs across all namespaces, then deletes the CRD document in a SECOND, separate `kubectl delete` call rather than folded into the first -- the split that stops a routine operator teardown from cascade-deleting every PubSubPlusEventBroker resource in the cluster |
| `TestOperatorDeleteRefusesCRDWhenBrokersExist` | M3: any PubSubPlusEventBroker CR still visible anywhere in the cluster refuses the CRD deletion outright -- deleting the CRDs would cascade-delete it along with every broker this env file never mentioned. The operator Deployment/RBAC removal still runs; only the CRD layer is refused, and each `<namespace>/<name>` is named in both the returned error and the warning on stderr |
| `TestOperatorDeleteSkipsBrokerCheckWithoutDeleteCRDs` | M3: `deleteCRDs=false` never even asks the cluster whether broker CRs exist -- there is nothing to refuse when the CRDs are already being kept, so no listing call belongs on that path, broker CRs present or not |
| `TestOperatorDeleteRefusesCRDWhenListingFails` | M3: the RBAC/API-error branch -- being unable to see what the CRD deletion would destroy is not permission to destroy it, so a failed broker-CR listing refuses the CRD layer too, preserving the cause via `%w` |
| `TestSplitOperatorBundle` | The column-0 anchoring behind that split: a ConfigMap with an indented `kind: CustomResourceDefinition` line buried in an example snippet is not misclassified as a CRD, while a real unindented CRD document still is; plus empty input, no-CRD input, and CRD-only input all hand back the nil/populated halves `OperatorDelete` expects |
| `TestOperatorRestart` | `rollout restart deployment` targets the operator deployment in its resolved namespace, behind the `patch deployments` permission probe |
| `TestOperatorInstalled` | Reports true only when both the CRD and the controller Deployment gets succeed, stops after the first failing get without probing the second, and its bool-only signature means an absent operator and an unreachable cluster are indistinguishable to the caller |
| `TestOperatorLogsArgs` | Log passthrough targets the operator deployment |
| `TestOperatorStatus` | OperatorStatus issues the deployment-wide get then the controller-pods get in order, and stops after the first if it fails |
| `TestOperatorDescribe` | OperatorDescribe issues `describe deployment/<name> -n <opNS>` against the resolved operator namespace |
| `TestOperatorNamespaceIsNotInTheDeleteStream` | The safety property the document split exists for. The Namespace document used to sit in `rest`, so `operator remove` deleted it in the same `delete -f -` stream as the Deployment -- cascading to everything else in that namespace, with none of the occupancy checks the broker's own namespace gets |
| `TestOperatorDeleteKeepsTheNamespaceAndRemovesTheRegcred` | The consequence of keeping the namespace: the image-pull Secret used to be reaped as namespace content, so it now needs deleting BY NAME or a registry credential outlives every teardown |
| `TestOperatorProbesEveryKindItTouches` | The fix for a check that passed and then failed halfway through the work: probing only the CRD meant an identity allowed to create CustomResourceDefinitions but not ClusterRoleBindings got past the preflight and died mid-apply -- the exact state `Preflight` exists to prevent |
| `TestOperatorRestartProbesTheOperatorNamespace` | A silent defect pinned: the restart happens in the operator's namespace while its permission check asked about the broker's, so an identity permitted in one and not the other passed and then failed |

### operatorversion_test.go

The operator-downgrade gate `OperatorApply` runs before its bundle apply: reads the
installed operator's own image, compares versions, and refuses to downgrade it without
an explicit yes -- the same shape as the container platforms' rotation-consent prompt.

| Test | What it covers |
| --- | --- |
| `TestImageTag` | Splits a registry/repo/tag reference without mistaking a registry port for a tag; a digest pin or a bare reference reports no tag |
| `TestCompareVersions` | Dotted-version comparison, including a `v` prefix and a missing trailing component defaulting to zero (numeric, not lexical: `1.10.0` > `1.9.0`); `latest`, empty, and non-numeric values report not-orderable rather than guessing |
| `TestOperatorVersionWarningFiresOnlyOnADowngrade` | The warning fires only when the cluster's existing version is newer than the one about to be applied, naming both versions and nothing else; it stays silent on an upgrade, a same-version reapply, a first install (nothing installed yet), and an unorderable or digest-pinned tag |
| `TestConfirmNoDowngradeIsSilentInAPreview` | Under Echo, `confirmNoDowngrade` neither blocks the preview nor warns about an installed version it has no cluster to read |
| `TestConfirmNoDowngradeRefusesByDefault` | The fail-safe default: with no `Confirm` callback, a genuine downgrade is refused and `OperatorApply` issues nothing after the version read (no namespace/regcred/bundle apply) |
| `TestConfirmNoDowngradeRefusedWhenConfirmDeclines` | An interactive operator answering no is refused the same way, still with no apply issued |
| `TestConfirmNoDowngradeProceedsWhenConfirmAccepts` | Answering yes lets a deliberate rollback through to the same three applies `TestOperatorApply` pins (namespace, regcred, bundle) |
| `TestConfirmNoDowngradeAsksNothingWhenNotADowngrade` | A same-version reapply and a genuine upgrade both proceed to all three applies without asking the question at all |
| `TestConfirmNoDowngradeProceedsWhenVersionReadFails` | A failed installed-version read (RBAC denial, unreachable API server) folds to `""` -- the same value a genuine first install produces -- so there is nothing to compare against and the apply proceeds without asking |

### ops_test.go

`Status`/`DescribeBroker`/`DescribeLB`/day-2 ops (`ops.go`) keep their direct kubectl-argv
coverage here. The read-only survey these tests used to cover --
`TestShowAll`/`TestShowAllDetailAddsStaticArtifacts`/`TestSurveyScopesToTheBrokerNamespace`/
`TestShowAllReportsAndContinuesOnGetError` -- was replaced by the rendering `statusreport.go`
now owns; their successors (`TestBrokerReportRunningPicture` and neighbours) are in
`statusreport_test.go` below.

| Test | What it covers |
| --- | --- |
| `TestStatus` | Status queries pods, services, and statefulsets in the broker namespace |
| `TestStatusFailureStopsEarly` | Status stops at whichever get fails first (pods, or pods+svc) instead of continuing to the remaining queries |
| `TestDescribeBroker` | Describe targets the pod for the requested role |
| `TestDescribeLB` | Describe targets the load-balancer service |
| `TestLogsPassthrough` | Extra log flags pass through to the role's pod |
| `TestCLIAndShellAreInteractive` | `cli` and `shell` run interactively with the right in-pod command |
| `TestCopyFrom` | Each file downloads under its base name; an empty list errors; failures are aggregated after every file is attempted |
| `TestCopyInto` | Uploads into the target dir, defaults it to `.`, errors on an empty list, aggregates failures |
| `TestRestartPod` | The manualPodRestart step: delete the role's pod with --ignore-not-found, then wait for the statefulset within the bounded rollout timeout |
| `TestRestartPodDeleteFails` | A failing pod delete surfaces its own actionable message and never reaches the rollout-status wait |
| `TestRestartRolling` | HA bounces monitor -> backup -> primary in that order, standalone only the primary, and a pod that does not come back stops the sequence before the next role |
| `TestReplicasStart` | HA scales and waits for all three roles, standalone only the primary; a stuck rollout fails loud at the first role, and the scale command itself failing outright is its own distinct error |
| `TestReplicasStop` | HA scales all three to zero, standalone only the primary, and a failing scale-down itself is its own distinct error |

### inspect_test.go

Decodes the raw `kubectl get ... -o json` objects the status reports read, against
fixtures under `testdata/` captured from a LIVE minikube deployment (operator 1.4.2,
broker 10.26.0.8755) rather than hand-guessed: the broker CR publishes
`ServiceReady`/`NoWarnings`/`HAReady` conditions rather than a conventional "Ready", and
its booleans are strings, so decoding against real bytes catches what a schema-shaped
guess would not.

| Test | What it covers |
| --- | --- |
| `TestDecodeBrokerCRFromLiveCapture` | Pins the three condition types the operator actually publishes (and that a plain "Ready" is NOT one of them), that `haDeployment`/`tlsSupport` are strings not bools, and the statefulSets/podsList counts |
| `TestBrokerConditionRowsFromLiveCapture` | A healthy captured broker's conditions all render OK rows carrying a message; standalone claims no HA row either way |
| `TestConditionLevelDegradesSafely` | `conditionLevel`'s degrade paths: an unpublished condition is INFO (not a false pass), a genuine False is FAIL carrying the operator's own message, `NoWarnings=False` inverts to WARN rather than FAIL, an unrecognised status value is INFO, and Reason stands in when Message is empty |
| `TestDecodePodsFromLiveCapture` | Pins the pod shape: a broker pod is single-container (READY 1/1, not a multi-container guess's 3/3), fully ready and unrestarted reads OK, and the role is recognised from the pod name |
| `TestPodHealthRules` | `podHealth`'s table across phase/ready/restarts: running+ready OK, ready-but-restarted WARN, pending or not-all-ready WARN, failed/unknown FAIL, `Succeeded` is INFO (not a broker problem), no container statuses yet WARN |
| `TestDecodeServicesFromLiveCapture` | The two real service shapes from the capture: the unassigned LoadBalancer address warns, and the headless discovery ClusterIP is INFO |
| `TestServiceAddressAssigned` | `serviceAddress` resolves an assigned LB IP or hostname to OK; a plain ClusterIP address is INFO -- a fact, not a verdict |
| `TestDecodePVCsFromLiveCapture` | Pins the claim shape including two distinct sizes (messaging vs monitor nodes); all Bound reads OK |
| `TestPVCLevels` | `pvcLevel` table: Bound OK, Pending WARN, Lost/Failed FAIL, an empty phase WARN |
| `TestDecodeOperatorFromLiveCapture` | Reads the installed operator's own image and replica counts from a live capture (1.4.2), proving drift from the 1.4.0 this repo renders is visible in the report |
| `TestReplicaLevels` | `replicaLevel` table: all ready OK, none ready FAIL, partially ready WARN, scaled to zero INFO, a nil spec defaults to one (OK) |
| `TestAgeMatchesKubectlShape` | The `age()` formatter matches kubectl's own AGE bands, including the two the live capture actually showed (6m40s, 40m); a zero timestamp reads "-" and a future one reads "0s" |
| `TestRoleRankOrdersTheRedundancyGroup` | `roleRank` sorts primary/backup/monitor by name suffix rather than kubectl's alphabetical listing, with an unrecognised name sorting last |
| `TestNormalizeToListHandlesBothKubectlShapes` | The regression guard for a live bug: `kubectl get deployment NAME -o json` returns a single object with no `items`, which `normalizeToList` wraps into a one-item list; an existing list or an empty list passes through untouched, and unrecognised input is left for the real decode to report |

### statusreport_test.go

The rendered `broker status`/`status cluster`/`operator status` reports (`statusreport.go`)
that replaced the old kubectl-passthrough survey `ops_test.go` used to cover: this tool
now decodes and grades the objects itself rather than reprinting kubectl's own columns,
driven by the `fixtureRunner` double over the same live-captured `testdata/` fixtures
`inspect_test.go` decodes directly, with the clock fixed via the `Cluster.Now` seam so the
AGE column is reproducible.

| Test | What it covers |
| --- | --- |
| `TestBrokerReportRunningPicture` | Successor to `TestShowAll`: the running picture for a healthy broker (Broker/Pods/Services/StatefulSets sections, the operator's own OK verdict, pods in redundancy order primary/backup/monitor rather than kubectl's alphabetical listing); the pending LoadBalancer still warns; Storage and a ROLE column are absent from the default report |
| `TestBrokerReportDetailAddsStorageAndPlacement` | Successor to `TestShowAllDetailAddsStaticArtifacts`: `--detail` adds the Storage section (claims, sizes) and the IP/NODE placement columns showing the three pods landed on three different nodes |
| `TestBrokerReportWithNoBrokerDeployed` | An absent broker reports a WARN naming it, rather than an empty section that reads as a pass |
| `TestBrokerReportSurfacesAReadFailure` | Successor to `TestShowAllReportsAndContinuesOnGetError`, with the behaviour deliberately changed: a failed pods read now fails the whole report loud (naming the read and the namespace) instead of continuing past it, since a health report that quietly omitted the pods would read as "no problems found" |
| `TestBrokerReportUsesTheCRsOwnPodList` | Broker-pod membership comes from the operator's own `status.podsList` rather than a name-substring filter, so a co-tenant workload sharing the infix cannot be counted as a broker pod |
| `TestOperatorReportNamesTheRunningVersion` | The operator report reads the installed Deployment's own image/replica counts rather than assuming the configured one, so cluster drift (the capture's 1.4.2 vs this repo's rendered 1.4.0) is visible |
| `TestOperatorReportWhenNotInstalled` | A missing operator is a WARN naming `operator deploy`, not an error |
| `TestClusterReportListsEveryBroker` | Successor to `TestShowAll`'s `--all` half: one row per broker across every namespace, carrying its HA state and worst condition |
| `TestClusterReportUnreportedBrokerIsInfo` | M8: a broker with no conditions published yet (freshly deployed, or an operator wedged before writing status) renders INFO ("not yet reported by the operator"), never a false OK |
| `TestClusterReportStillSurfacesAFailingCondition` | Guards the M8 fix from overcorrecting: a broker that DOES have a failing condition still shows FAIL rather than falling back to the unreported state or an OK seed |
| `TestClusterReportWithNoBrokers` | An empty cluster says so in words |
| `TestStatusReportsArePreviewableUnderEcho` | Under Echo, all three reports (broker/cluster/operator) echo their reads and claim no `[ OK ]`, since there is no cluster to decode |
| `TestReportsDegradeRatherThanFalselyAlarm` | The degrade-path table: missing pods/services/statefulsets each WARN; missing claims is INFO, not a warning (running without persistent storage is legitimate); a failed svc/statefulsets/pvc read surfaces naming the read; a failed broker-CR read surfaces on both BrokerReport and ClusterReport; malformed JSON from the cluster names the `kubectl get` command to run by hand |
| `TestOwnsPodFallsBackToTheNameInfix` | The fallback used only when the operator publishes no podsList: excludes the operator's own pod (no leading dash) and, once a list IS present, counts only what it claims |
| `TestOperatorRunningImageWithNoContainers` | A Deployment spec with no containers reports "no container" rather than a blank image string |
| `TestNowDefaultsToTheWallClock` | The injectable `Cluster.Now` clock seam defaults to the real wall clock, not the zero time |
| `TestBrokerReportReadsEachKindOnce` | Pins the read budget: each kind (broker CR, pods, svc, statefulsets, pvc) is read exactly once per report even though the broker CR feeds both conditions and podsList, and the pvc read is skipped entirely without `--detail` |

### secrets_test.go

| Test | What it covers |
| --- | --- |
| `TestSecretGoldens` | Rendered admin, TLS, docker-registry, and operator-regcred secrets match their committed goldens |
| `TestAdminSecretDecodes` | The base64 data round-trips to the expected plaintext passwords |
| `TestAdminSecretExcludesAdditionalUsers` | The finding that shaped the k8s user path: the operator reads only the admin and monitor keys, so an additional user's name and password (plain and base64) must be absent from this Secret entirely. They live in `<kubernetes.name>-additional-users` instead -- see `TestAdditionalUsersStayOutOfTheCredentialsSecret` for the envFrom reason that split is mandatory |
| `TestAdminSecretErrors` | Empty password, empty `kubernetes.adminSecret`, and an additional user with no name, a bad name, or no password all error |
| `TestTLSSecretErrors` | Unset cert, unset secret name, and missing cert/CA/key files all error |
| `TestAdminSecretCarriesThePSKOnlyWhenSet` | The Kubernetes half of the PSK asymmetry, from the Secret's side. The key rides in the SAME Secret as the credentials -- the CR points at it through a separate field, so one object serves both -- and under the CRD's own spelling (`preshared_auth_key`), not ours. An unset key adds no entry at all: an empty one is a key the operator would honour, and the group then fails to form on it |
| `TestGenSecretsSkipsATLSSecretItDoesNotOwn` | The fix for a real failure: an env file naming an existing Secret and supplying no cert/key made `broker generate` read `certs/tls.crt` -- an invented path -- and fail on a file the operator never mentioned. The stream still carries the credentials Secret, so this narrows one document rather than the whole thing |
| `TestGenSecretsBuildsATLSSecretItOwns` | The other arm: with the pair supplied the Secret is rendered as before, under the configured name. `writeTempPEM` is the stub-file helper -- the renderers copy the bytes rather than parsing them |
| `TestAdditionalUsersSecretCarriesBothHalves` | Each user's access level AND password ride the Secret, even though the level is not sensitive: `extraEnvVarsSecret` is the only channel the CRD gives us for an environment variable, so a setting that must reach the broker as one has nowhere else to go -- unlike the container platforms, which put it in the artifact. The values are base64 in the data block, so neither plaintext may appear anywhere |
| `TestAdditionalUsersSecretRefusesAnEmptyPassword` | The guard's own branch. Defensive rather than everyday: `config.Validate` refuses an empty additionalUsers password at load, so reaching this needs a `Config` built in code -- which is exactly what the executors are handed, and the same reason `AdminSecret` keeps its own empty-password check |
| `TestAdditionalUsersSecretIsAbsentWithNoUsers` | nil, not an empty Secret. An empty one would be an object to create, name and clean up for nothing, and the CR omits `extraEnvVarsSecret` in the same case -- naming a Secret that does not exist fails the pod |
| `TestAdditionalUsersStayOutOfTheCredentialsSecret` | The property the two-Secret split exists for, and it is not tidiness: `extraEnvVarsSecret` is projected with `envFrom`, which exports EVERY key of the Secret it names. If these users lived in the credentials Secret, pointing the CR at it would publish the ADMIN and MONITOR passwords to get the extra users in. Asserted in both directions |
| `TestGenSecretsIncludesTheAdditionalUsers` | The stream `broker deploy` applies and `broker generate` prints carries the Secret, and omits it entirely when no users are configured |
| `TestDockerRegistrySecretEmptyName` | An empty pull-secret name errors |

### transport_test.go

| Test | What it covers |
| --- | --- |
| `TestTransportExecArgs` | Exec argv for Run/Output/OutputInput, `-i` only where stdin is used, and `-c` never present (broker pods are single-container) |
| `TestTransportUpload` | The body rides stdin through `sh -c 'cat > <dest>'` and never appears in the argv |
| `TestTransportUploadQuotesDest` | Single-quote escaping stops a metacharacter in a path breaking out of the redirect |
| `TestTransportCopy` | `kubectl cp` argv in both directions with the namespace flag |
| `TestTransportEchoHidesUploadBody` | End to end over Echo: the uploaded body shows as a byte count, never in the traced line, and the exec is still echoed against the primary pod and names this script's own broker-side files. Anchored on the pod and `sh -c` rather than a fragment of the generated skeleton, which is multi-line and shell-quoted |

---

## internal/container

The host-local Docker/Podman manager, its node-local transport, and the engine
preflight that precedes every mutating operation, plus the engine `inspect` decode
behind `broker status` and the server-certificate delivery each engine needs.
132 tests across 6 files.

### runtime_test.go

| Test | What it covers |
| --- | --- |
| `TestManagerHonoursRuntime` | The manager's shell-outs (`run`, `output`, `CLI`, `Shell`) run argv[0] from `docker.runtime` with its leading arguments ahead of the subcommand -- the bash bootstrap expanded `${CONTAINER_RUNTIME}` unquoted, so a wrapper like `sudo -n docker` has to reach exec as argv |
| `TestManagerReachableProbesRuntimeThenCompose` | Docker `Reachable` probes both the engine and compose, and the derived compose default keeps the runtime wrapper (`sudo -n docker compose`, not a bare `docker compose`) |
| `TestCtrTransportHonoursRuntime` | The node-local transport does the same for `exec`, the stdin `Upload`, and `cp` |
| `TestCtrRuntimeDefaultArgvUnchanged` | A single-token runtime produces exactly the argv it did before, for both docker and podman |
| `TestCtrExecutorRefusesUnapprovedRuntime` | The container half of enforce-twice: a `Manager` and a transport built from a `*config.Config` that never saw `config.Load` still refuse an unapproved `sudo -n docker` on every path that could reach exec, handing the runner nothing |

### preflight_test.go

The read-only engine probe, and the child-environment hygiene it shares with
`internal/engine`.

| Test | What it covers |
| --- | --- |
| `TestPreflightRunsBeforeAnything` | The layer-7 ordering guarantee: `<runtime> info` is the FIRST call `Deploy`, `Delete`, `PrepHost`, `Start`, `Stop` and `Restart` make, on both platforms -- anything before it would be host state left behind by an operation that then failed on a stopped daemon |
| `TestPreflightFailureStopsTheDeploy` | An unreachable engine stops `Deploy` nonzero, carries the engine's own error, adds the actionable hint, writes no compose file, and issues no call after the probe |
| `TestPreflightFailureStopsLifecycle` | The same ordering guarantee for `Start`, `Stop` and `Restart`: an unreachable engine fails each one with no call beyond the probe -- these act on an already-deployed broker, so there is no artifact write to check, only that nothing else ran |
| `TestPreflightHintIsPlatformShaped` | Docker gets the daemon/group hint; rootful podman gets `sudo systemctl start podman.socket`; rootless podman gets the user-session hint and explicitly NOT a sudo suggestion, which would start the engine its deploy is not using. None of them offers to act on the operator's behalf |
| `TestPreflightIsPreviewableUnderDryRun` | `--dry-run` echoes the probe and skips its assertion, so previewing needs no engine -- which is why there is no skip flag |
| `TestComposeSecretEnvNamesCannotBeSystemVars` | The config-side half of the child-environment rule: even with `container.name` set to `PATH`, `LD_PRELOAD`, `ld.preload` or `IFS`, every variable name keeps its fixed literal suffix, so no config value can produce a name the child's loader reads |
| `TestComposeSecretEnvIsTheOnlyChildEnvironment` | `composeSecretEnv` passes through exactly the secrets `render` declares and invents none, which is the assumption the test above rests on; values stay masked in any display path |

### manager_test.go

| Test | What it covers |
| --- | --- |
| `TestManagerCheckDryRun` | Preflight report for docker/podman x HA/standalone: title, mode line, runtime version probe, dry-run skip note |
| `TestManagerCheckDNSFailsLoudInHA` | An unresolvable redundancy host fails the check and is named |
| `TestManagerCheckStandaloneDNSWarnsOnly` | Standalone tolerates an unresolved name |
| `TestManagerPrepHostRootlessUsesUnshareChown` | Rootless podman chowns via `podman unshare` |
| `TestPrepHostRootlessNoFileSufficient` | Rootless prep probes this user's hard `nofile` limit with `sh -c 'ulimit -Hn'` and reports the value when it covers `container.ulimits.nofile` |
| `TestPrepHostRootlessNoFileTooLow` | The point of the check: a rootless container cannot raise `nofile` past the user's hard limit, so prep stops rather than deploying a broker that would run under-provisioned. The message carries both numbers and the exact `limits.d` drop-in, including the re-login that re-reads it |
| `TestPrepHostRootlessNoFileUnlimited` | An unlimited hard limit satisfies any configured value |
| `TestPrepHostRootlessNoFileUnreadable` | A limit that will not parse fails loud rather than being assumed adequate |
| `TestPrepHostRootlessNoFileUnsetSkips` | With no configured `nofile` there is nothing to assert against, so the probe never runs -- the hand-built config the executors are handed |
| `TestPrepHostRootfulSkipsNoFile` | Docker and rootful podman never probe: their privileged engine raises the limit itself, so the invoking user's hard limit does not bound the container |
| `TestPrepHostRootlessNoFileDryRun` | `--dry-run` echoes the probe and skips the assertion, the same shape `Preflight` uses, since the Echo runner answers nothing |
| `TestSplitLimit` | The `soft:hard` ulimit parser: a pair, a single value meaning both, surrounding whitespace, and the values that mean "nothing to assert" (`-1`, empty, non-numeric) |
| `TestManagerDeployDockerComposeWritesFile` | Deploy writes the compose file and runs `compose up -d --force-recreate` |
| `TestManagerDockerComposeCommandOverride` | A `docker.compose` override (the standalone `docker-compose` binary) is what every compose call goes through |
| `TestManagerDockerCheckProbesCompose` | Docker `check` probes the compose command, so a missing plugin fails at check time rather than at deploy time |
| `TestManagerDockerCheckFailsWhenComposeMissing` | With only the compose probe failing, the error names the `docker.compose` override |
| `TestManagerDeployDockerPassesSecretsAsEnv` | Deploy writes no secret file at all: `compose up -d` goes through `RunEnv` carrying `SOLACE_ADMIN_PASSWORD`/`SOLACE_REDUNDANCY_PSK`, no value reaches an argv, and the compose file holds the source/target/variable references and the `*filepath` pointer instead of either value |
| `TestManagerDeployPodmanCreatesSecrets` | Deploy loads both secrets into podman's store under container-scoped names (`sol-pod-*`) with `secret rm --ignore` then `secret create` -- remove-then-create rather than `create --replace`, since `--replace` needs podman 4.7 while the rest of the wiring needs only 4.5, and `rm --ignore` makes the pair just as idempotent; the test pins `--replace`'s absence from every call, not only the new shape's presence. Values ride stdin and never reach an argv; the quadlet unit is 0600 |
| `TestManagerDeployRejectsEmptySecret` | An empty required secret fails the deploy naming the field rather than starting a broker without one. It is a backstop rather than the everyday gate: on a loaded env file `config.Validate` has already refused an empty psk, so reaching this needs a `Config` built in code |
| `TestManagerDeployDockerDryRunMasksSecretEnv` | Dry-run creates nothing, needs no secret value on disk, and echoes the compose environment as `NAME=***` without the password |
| `TestManagerDeployPodmanDryRunHidesSecretBytes` | The dry-run echo shows the secret-create command and a stdin byte count, never the values |
| `TestManagerRedeployUnchangedRestartsForRotation` | Docker: `--restart` against an unchanged compose file forces `up -d --force-recreate`, which is the only way a rotated secret reaches the running broker |
| `TestManagerRedeployPodmanUnchangedRestartsForRotation` | Podman: the same state restarts the service (the store was refreshed, but the running container holds the old values) |
| `TestContainerRunningMatchesNameExactly` | The branch selector matches the container name exactly, so a sibling deployment on the same host (`solace-edge` next to `solace`) is never mistaken for this one: exact, among-others, sibling-only, prefix-only, empty and whitespace-padded listings, plus a failed probe reading as not-running |
| `TestManagerDeployDockerProbeUnansweredErrors` | M4: when the `ps --filter status=running` probe itself fails, `deployDocker` no longer reads that silence as "not running" and force-recreates without consent -- it errors out naming the container and that the probe could not answer, and neither create nor recreate reaches the runner |
| `TestManagerRedeployStoppedContainerRecreates` | The arm with no consent prompt: an unchanged compose file with the container stopped recreates rather than starts, because a start would replay the credentials the container was created with |
| `TestManagerRedeployUnchangedHintsRotation` | Without `--restart` nothing is recreated and the log names `--restart` as the way to apply a rotation |
| `TestManagerDeployPodmanWritesUnit` | The quadlet unit is written, then daemon-reload and service start |
| `TestManagerDeployPodmanDryRunSkipsWrite` | Dry-run echoes the systemctl steps without writing the unit |
| `TestManagerPodmanEUIDGuardSkippedOnDryRun` | The rootless/rootful euid guard does not run under dry-run |
| `TestManagerDeletePodmanRemovesUnit` | Delete stops the service, removes the unit, and daemon-reloads |
| `TestManagerDeletePodmanStopFailsServiceActiveBlocksRemoval` | A failed `systemctl stop` proves nothing by itself (`podman info`/Preflight only shows the engine is reachable), so when `serviceState` still reports the unit `active`, Delete blocks the unit removal, the daemon-reload, and (via the purge gate) the data-directory rm, instead of reporting success over a broker still serving traffic |
| `TestManagerDeletePodmanStopFailsServiceInactiveProceeds` | The same failed stop, but `serviceState` confirms the unit is already `inactive` -- the benign "already stopped" case still proceeds exactly as before |
| `TestManagerDeletePodmanStopFailsStateUnknownBlocksRemoval` | The case the guard exists for: the stop failed AND `systemctl is-active` answered nothing, which is what an unreachable rootless systemd user session looks like while `podman info` still succeeds on the engine socket. `is-active` exits non-zero for every state but `active`, so the exit code cannot tell "stopped" from "could not ask" -- only the state text can, and no text means refuse. Pins that silence is never read as "already stopped" |
| `TestManagerDeletePodmanRemovesSecrets` | Delete now removes every secret `CreatePodmanSecrets` loaded into podman's own store (the same `render.ContainerSecrets` list), so they no longer outlive a `broker remove --delete-data`; a failing removal warns rather than failing a teardown that otherwise succeeded |
| `TestManagerDeletePodmanPurgeRootless` | Rootless purge removes the data dir via `podman unshare` |
| `TestManagerDeleteDockerComposeDownWhenFileExists` | With a compose file present, delete runs `compose down` |
| `TestManagerDeleteDockerPurgeRemovesDataDir` | Delete runs `compose down` and, with purge, removes the data dir |
| `TestManagerDeleteDockerComposeNoFileFallsBackToStopRm` | A missing compose file falls back to stop+rm |
| `TestManagerStopAndRemoveContainerAbsentNoOp` | `docker rm` on a name never deployed exits non-zero, which used to turn "reset after a failed deploy" into an error; `containerExists` (matched against a `ps --all` listing, not the unanchored `--filter name=` regex) makes the no-compose-file fallback no-op like every other removal path in the tool |
| `TestManagerStopAndRemoveStopFailsContainerRunningBlocks` | The docker half of the same stop-failure check as podman's: `docker info`/Preflight proves the engine is reachable, not that the container stopped, so a failed stop with the container still running blocks `rm` instead of falling through to it |
| `TestManagerStopAndRemoveStopFailsProbeUnansweredBlocks` | The docker twin of `...StateUnknownBlocksRemoval`: the stop failed and the `ps` running-probe could not answer either. `containerRunning` alone answers "false" here -- the deferred M4 hazard -- which is why `stopAndRemove` reads `containerRunningKnown`'s second return value and refuses on an unanswered probe rather than falling through to `rm` |
| `TestManagerLifecyclePodmanSystemctl` | `Start`/`Stop`/`Restart` on podman drive `systemctl {start\|stop\|restart} <name>.service` rather than the engine directly, since the quadlet unit owns the container's lifecycle -- covered rootful and rootless, since only rootless prepends `--user` |
| `TestStatusAllFindsBrokersByImage` | The container answer to `--all`: discovery is by IMAGE, not by the configured container name, so a broker deployed by hand still appears and an unrelated container never does. A config-scoped listing cannot answer "what is actually running on this host", which is the whole reason the flag exists |
| `TestStatusAllReportsNothingFound` | An empty result says so in words. This is the ordinary case on a host not yet deployed to, where a bare header would read as though the command failed to look |
| `TestStatusAllWrapsListError` | An engine that cannot be asked (docker not running, podman socket down) fails loud naming the cause, rather than reporting an empty host -- which would be a lie in exactly the situation being diagnosed |
| `TestStatusAllDetailInspectsEachAndKeepsSecretsOut` | `--detail` inspects every container it found, and the inspect format is deliberately mounts-only. Secrets are files under `/run/secrets`, so mounts name them without reading them -- while docker's compose secrets are environment-sourced, so reading the environment here would put passwords on the terminal and into scrollback (S3) |
| `TestManagerLifecycleDockerComposeFile` | `Start`/`Stop`/`Restart` on docker drive `compose -f <file> {start\|stop\|restart}` when the compose file is on disk |
| `TestManagerLifecycleDockerNoComposeFile` | With no compose file on disk, the same three verbs fall back to the plain `<runtime> {start\|stop\|restart} <name>`, mirroring Delete's stop/rm fallback |
| `TestManagerLifecycleDockerDryRunUsesCompose` | Under `--dry-run` there is no file to probe (Echo never wrote one), so the preview always takes the compose branch |
| `TestManagerRedeployUnchangedIsNoOp` | Re-deploying an unchanged artifact against a running broker touches nothing on either platform and says there was nothing to do |
| `TestManagerRedeployChangedNeedsConsent` | A changed artifact against a running broker is written but not applied without consent (warning that the broker is still on the previous one); `--restart` and an accepted prompt both apply it. This is what stops a silent no-op on podman, where `systemctl start` on an already-active unit would otherwise leave the old image running while reporting success |
| `TestManagerDescribe` | `<runtime> inspect` on both platforms, plus the installed unit on podman, with a missing unit tolerated |
| `TestManagerCopy` | The copy verbs: cp out of and into the container, per-file reporting, an error with no files, and a non-zero exit when any file fails |
| `TestManagerPrepHostRegistryLogin` | prep logs in to the registry with the password on stdin, never in an argv |
| `TestManagerPrepHostNoLoginWithoutCreds` | No credentials means no login attempt |
| `TestManagerPrepHostRejectsHalfCredentials` | A user with no password (or the reverse) fails loud rather than attempting a broken login |
| `TestManagerStatusPodman` | Status shows the systemd unit and lists the container |
| `TestManagerStatusDocker` | Docker status issues ONE listing, the narrowed `ps`, and mentions neither compose nor a ports column. The format string is not spelled out: it carries real tabs and `Echo` shell-quotes what it prints, so a literal would assert Echo's quoting rather than this command's argv |
| `TestManagerLogsCLIShell` | `logs` follows, `cli` and `shell` exec interactively into the container |
| `TestManagerCheckReachableError` | A failing runtime version probe fails the check |
| `TestManagerPrepHostMkdirError` | A mkdir failure propagates |
| `TestManagerPrepHostChownError` | A chown failure propagates |
| `TestManagerPrepHostRootlessUnshareChownError` | A rootless `unshare chown` failure propagates |
| `TestManagerDeployPodmanMkdirError` | An uncreatable quadlet dir fails deploy |
| `TestManagerDeployPodmanWriteUnitError` | An unwritable unit path fails deploy |
| `TestManagerDeployPodmanDaemonReloadError` | A daemon-reload failure propagates |
| `TestManagerDeployPodmanStartError` | A service-start failure propagates |
| `TestManagerDeployPodmanEUIDGuardFails` | Rootless-as-root is rejected by the euid guard |
| `TestManagerDeployDockerComposeWriteError` | An unwritable compose path fails deploy |
| `TestManagerDeployDockerComposeUpError` | A `compose up` failure propagates |
| `TestManagerDeletePodmanDaemonReloadError` | A daemon-reload failure during delete propagates |
| `TestManagerDeletePodmanRemoveUnitError` | An unremovable unit path fails delete |
| `TestManagerDeleteDockerComposeDownError` | A `compose down` failure propagates |
| `TestManagerDeleteDockerStopTolerated` | The tolerate branch B2 kept: a failed stop that the engine then CONFIRMS is not running still warns and goes on to `rm`. Uses `capRunner.outFor` so the two probes answer differently -- `ps --all` lists the container (so stop is attempted at all), `ps --filter status=running` does not (so it is confirmed down). Confirmation is what earns the tolerance; silence does not |
| `TestManagerDeletePurgeError` | A failing data-dir removal under `--purge` propagates |
| `TestManagerStatusDockerNoComposeFile` | With no compose file on disk, status lists the container and never calls compose |
| `TestManagerStatusPodmanUnitInactiveTolerated` | An inactive unit warns but status still lists the container |
| `TestManagerStatusDockerRunsNoComposePs` | Status runs no compose at all, even WITH a compose file on disk. It replaced the tolerated-failure test: compose listed the same single container, and its PORTS column -- every published port with both host bindings -- was the widest thing in the report with no way to narrow it, since compose's `--format` takes only `table` or `json` |
| `TestStatusListingCarriesNoPortsColumn` | The property the format exists for, asserted on the format rather than on engine output no test can produce: `psTableFormat` asks for NAMES/IMAGE/STATUS and never Ports |
| `TestManagerCheckPodmanEUID` | The euid guard across rootless/rootful x root/non-root, and skipped on a non-POSIX euid |
| `TestManagerPrepHostRootlessAsRootFailsHard` | PrepHost now shares Deploy's `checkPodmanEUID` and fails hard (one definition, one message) when `podman.rootless=true` but the process is root, instead of only warning and going on to mkdir/chown under the wrong namespace mapping -- neither call is reached once the guard rejects |
| `TestManagerDeployPodmanSecretError` | BOTH halves of the store write, because loading a secret is two commands rather than one (`secret rm --ignore` then `secret create`). Either failing must stop the deploy and name the CONFIG KEY behind the secret, so the operator learns which env-file field to look at rather than which podman verb failed. The `rm` half is deliberately fatal: `--ignore` already absorbs the only benign case (nothing in the store yet), so a failure that survives it is real, and creating a secret beside one that could not be removed would leave the store in a state nobody chose |
| `TestManagerNilSinks` | Nil log and output sinks fall back to discard and stdout without erroring |

### transport_test.go

| Test | What it covers |
| --- | --- |
| `TestTransportExecArgs` | Exec argv is `<runtime> exec [-i] <name> ...` with no `--` (docker rejects it), and the role argument is ignored because the transport is node-local |
| `TestTransportUpload` | The body rides stdin through `sh -c 'cat > <dest>'` and never appears in the argv |
| `TestTransportUploadQuotesDest` | Single-quote escaping stops a metacharacter in a path breaking out of the redirect |
| `TestTransportCopy` | `<runtime> cp` argv in both directions, container-name prefixed |
| `TestTransportEchoHidesUploadBody` | End to end over Echo: the body shows as a byte count, never in the traced line, and the exec carries no `--` and names this script's own broker-side files. Not reached by `regen` (which runs six packages, not this one), so its k8s twin failed alone when the argv shape changed |
| `TestTransportEchoHidesSEMPConfig` | End to end over Echo via LeaderLocal: the mate SEMP curl is echoed with its stdin as a byte count, never the admin password |

### inspect_test.go

The single `<runtime> inspect` decode behind `broker status` on both engines. Docker and
podman disagree about spelling in three places -- the container name, the health status
key, and what a restart even counts -- so these pin the reconciliation and the difference
between an unknown answer and a confident wrong one.

| Test | What it covers |
| --- | --- |
| `TestInspectNameLosesDockersLeadingSlash` | `/solbroker` is not a name either engine accepts back, and printing it beside podman's `solbroker` would make one engine's report disagree with the other about what the container is called |
| `TestInspectHealthReadsBothSpellings` | The one that matters most: a report naming only one engine's spelling renders EMPTY on the other, which reads exactly like "no healthcheck configured" -- the most misleading thing this report could say about a broker |
| `TestInspectDistinguishesNoHealthcheckFromUnknown` | An engine reports an unconfigured healthcheck as an empty status, so "not configured" and "configured but unknown" must not print the same -- one is a deployment choice, the other a missing answer |
| `TestInspectRejectsAnEmptyReply` | Some engine versions answer a name they do not know with an empty ARRAY and exit 0, so the absence is turned into an error here rather than reported as a container with no name, no image and no state |
| `TestInspectDecodesPartialOutput` | An engine version that stops emitting a field leaves that field unknown rather than failing the whole command -- the report is the point, and three quarters of it is worth more than none |
| `TestInspectStateSurfacesTheEngineError` | An engine that cannot be asked (daemon down, socket gone) surfaces its own error rather than a parse failure about the empty output that follows, which would send whoever reads it looking for a malformed reply that was never sent |
| `TestRestartCountComesFromSystemdOnPodman` | The fact this decode exists for: a quadlet unit is restarted by systemd, which stops and REPLACES the container, so the engine's own `RestartCount` stays 0 through any number of restarts -- reporting it would tell an operator their crash-looping broker has never restarted |
| `TestRestartCountUsesTheEngineOnDocker` | Docker restarts the container in place, so its own counter is the answer and there is no unit to ask |
| `TestRestartCountDoesNotAttributeOurUnitToAnotherContainer` | `broker status --all` discovers containers BY IMAGE, so it reaches brokers this env file knows nothing about -- and this env file's unit restart count is not their number |
| `TestRestartCountUnknownRatherThanZero` | systemd not answering (unit not installed, rootless/rootful mismatch) reads as unknown. A zero there claims "this broker has never restarted", which this tool has no basis for |
| `TestRestartCountUnknownWhenSystemctlFails` | The other way systemd declines to answer -- the command itself fails rather than answering unreadably. Both must read as unknown, never as zero |
| `TestReportStateNeverPrintsTheEnvironment` | On docker the compose secrets are environment-sourced, so a report that dumped the container's environment would put the admin password on a terminal, into scrollback, and into whatever ticket the output is pasted into. Secrets appear only as mount PATHS |
| `TestReportStateOnAStoppedContainer` | The half of the report an operator only reads when something has gone wrong: no start time worth printing, but the three facts that say why it stopped -- finish time, exit code, and the engine's own reason. It also pins that a writable mount is labelled `rw` and a read-only one `ro`, so a secret mount that turned up writable is a real finding |
| `TestReportStateOnAHealthyRunningContainer` | The ordinary case and the mirror of the above: health under docker's spelling with no failing-streak noise, a start time, and no exit code invented for a container that has not exited |

### secrets_test.go

The server certificate is the one secret the two engines deliver differently -- docker
inlines it through the compose child's environment, podman cannot (a quadlet unit inlines
nothing), so it lands as a `0600` host file the unit bind-mounts. These pin both routes,
the "nothing happened" property on each failure path, and previewability.

| Test | What it covers |
| --- | --- |
| `TestResolveSecretValuesBuildsTheBundle` | The read itself: the resolved value is exactly what `broker.ServerCertBundle` builds, key before certificate, and every credential secret is left alone |
| `TestResolveSecretValuesPreviewReadsNothing` | What keeps a dry run usable before the certificate exists on this host. `compose()` runs on six verbs, so without it a preview of any of them would fail on a missing file -- the same previewability `prepareSecrets` already protects for the pre-shared key |
| `TestResolveSecretValuesNamesTheUnreadableFile` | The error names the file, the secret AND the env-file key, because this failure reaches the operator from six different verbs |
| `TestResolveSecretValuesRefusesAKeyBearingCert` | The existing refusal from its new call sites: a `tls.cert` that already contains its private key would produce a bundle carrying the key twice, and the refusal now fires from inside `compose()` on docker and from the bundle write on podman |
| `TestPodmanResolvesNoFileBackedSecret` | The proof of the split: podman's list has no file-backed entry, so no private key can reach podman's secret store |
| `TestDeployDockerPassesTheBundleAsEnvNotArgv` | The docker delivery end to end -- the bundle reaches the compose child through its environment and never through an argument vector |
| `TestTeardownVerbsDoNotReadTheCertificate` | A moved certificate must not block recovery. Compose needs every declared secret DEFINED for down/stop/restart but never reads it -- only `up` does -- so those verbs keep working when the file is gone |
| `TestDeployDockerFailsBeforeWritingTheComposeFile` | "Nothing happened" on the docker side: the certificate is read in `prepareSecrets`, before `deployDocker` rewrites the compose file, so a bad certificate leaves no artifact behind |
| `TestDeployPodmanWritesTheBundleBeforeTheUnit` | Podman's whole delivery: the file exists before anything starts, at a mode fit for a private key, and the unit bind-mounts exactly that path |
| `TestDeployPodmanFailsBeforeWritingAnythingOnABadCert` | Podman's half of the same property -- the bundle is built first, so a failure leaves no unit and issues no systemctl call |
| `TestDeployPodmanFailsWhenTheBundleCannotBeWritten` | The write's error path, and that the message names the artifact rather than surfacing a bare filesystem error. A directory standing where the file belongs is the cheapest way to fail the write without depending on permissions, which differ across the OSes this suite runs on |
| `TestDeletePodmanRemovesTheBundleAfterTheUnit` | Teardown ordering and the severity choice: the bundle is something the running container consumed, so it goes after the unit is confirmed gone -- and its removal is FATAL rather than a warning, because silently leaving a private key on the host is the outcome least like the rest of this teardown |
| `TestDeletePodmanToleratesAMissingBundle` | A teardown of a deployment that never had TLS, or whose `tls.cert` was unset since deploy |
| `TestDryRunNeedsNoCertificateOnDisk` | The previewability guarantee one level up from the pre-shared key's: a preview must work before any certificate exists |
| `TestEchoDeletePreviewsTheBundleRemoval` | The dry-run half of teardown -- it says what it would remove and removes nothing. Without it the only coverage of that branch would be the deploy-side preview, and a dry run that actually deleted a private key is the mistake worth its own test |
| `TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv` | Closes a round trip that had no check: the names come out of `<runtime> ps` on this host, not from the env file, and `StatusAll` puts each straight back into an `inspect` argv. The row is still SHOWN -- it is a real container and hiding it would be worse -- but its name is not returned, so nothing further runs against it, and the skip is stated rather than silent |

---

## internal/engine

The command runner seam: `Echo` (dry-run) and `Exec` (real subprocess), plus display
quoting, PATH resolution, and the pre-exec announcement. 26 tests across 2 files.

### runner_test.go

| Test | What it covers |
| --- | --- |
| `TestHelperProcess` | Not a test -- the os/exec helper-process shim used as a fake external command by the `Exec` tests below |
| `TestQuoteTok` | Display quoting per token: empty, plain, and every shell-significant character |
| `TestQuote` | Whole command lines, including quoted and empty arguments |
| `TestEchoRun` | `Run` echoes `+ <cmd>` |
| `TestEchoRunInteractive` | `RunInteractive` echoes with quoting applied |
| `TestEchoRunInput` | `RunInput` shows stdin as a byte count, never its contents |
| `TestEchoRunEnv` | `RunEnv` echoes the command first and annotates the variables it would set after it (`<<< (env: NAME=***)`), so `+ <cmd>` stays greppable and no value is printed |
| `TestEchoRunEnvNoEnv` | With nothing to annotate the line is exactly what `Run` prints, not a dangling `(env: )` |
| `TestMaskEnv` | The masking helper keeps names (quoting an odd one) and drops values, including a value holding `=` and an empty one |
| `TestEchoOutput` | `Output` echoes and returns nil bytes |
| `TestEchoDefaultWriter` | A zero-value `Echo` writes to stdout |
| `TestExecOutput` | `Output` captures a child process's stdout |
| `TestExecOutputFail` | A non-zero exit errors and the message names the binary |
| `TestExecRun` | `Run` streams on success and errors (naming the binary) on failure |
| `TestExecRunInput` | `RunInput` feeds stdin to the child and streams its output |
| `TestExecRunEnv` | `RunEnv` gives the child the extra variable *and* still inherits this process's environment; a non-zero exit errors naming the binary |
| `TestExecRunInteractive` | `RunInteractive` runs a child to a clean exit |
| `TestExecOutputInput` | OutputInput wires stdin from `in` into the child and captures stdout into the returned buffer rather than leaking it to the real terminal -- the curl -K - path this backs has no other way to get the response body back |
| `TestExecOutputInputFail` | OutputInput wraps a child failure the same way Output does ("name: err") |
| `TestEchoOutputInput` | the dry-run echo for the credential-bearing curl -K - path prints a byte count and never the stdin body, mirroring TestEchoRunInput but with an explicit assertion that the fake credential never appears in the echoed line |

### resolve_test.go

PATH resolution and the command announcement -- the transparency half of the
execution guard, which lives here because this is where a binary is actually run.
The announcement is injected (`Exec.Announce`) rather than written to a package
variable, so what a run prints is decided by the CLI, in one place.

| Test | What it covers |
| --- | --- |
| `TestExecVerboseAnnouncesEveryCommand` | Under `--verbose`, `Exec` writes `==> exec: <absolute resolved path> <args>` before each command -- the path, not the name as typed, with the arguments alongside -- and on *every* call, since a trail with one entry per binary would not answer "what did this run issue?" |
| `TestExecIsSilentWithoutVerbose` | The default runner announces nothing, and neither does a bare `Exec{}`: nil `Announce` is the quiet default rather than a hole that falls back to stderr. The CLI names the binaries once in its preamble instead, which is what stopped the same resolved path landing between report lines on every command |
| `TestExecEchoesOnEveryMethod` | All six `Runner` methods announce, not just `Run` -- and `Output`/`OutputInput`, which read cluster state, are the least visible to begin with |
| `TestResolveMissingBinaryIsActionable` | A name that resolves nowhere fails before any process starts, naming what was not found rather than a path the operator never typed. Asserted on `Resolve` (which the CLI preamble now shares) and again through `Exec.Run` |
| `TestResolveRefusesCurrentDirectory` | The pair to config's bare-name rule: a bare name must never resolve to a file in the working directory -- the binary unpacked beside a shared env file. Go reports it as `exec.ErrDot`; hosts that do not offer the cwd copy at all are logged and still asserted not to run it |
| `TestChildEnvNamesAreNotSystemVariables` | The variable names this tool passes to a child are never `PATH`, `LD_PRELOAD` or their relatives, and their values stay masked in display paths. The upstream half is container's `TestComposeSecretEnvNamesCannotBeSystemVars` |

---

## internal/render

Manifest and unit-file rendering, guarded by committed goldens, plus the server
certificate's two delivery routes. 32 tests across 2 files.

### render_test.go

| Test | What it covers |
| --- | --- |
| `TestGolden` | Thirteen renderings from the sample env match their goldens: the k8s broker CR (the sample omits `kubernetes.ports`, `timezone` and both security blocks, so this covers the default ports and the omitted branches), the same CR with an explicit port list (a container port differing from the service port, and an explicit protocol), the same CR with timezone and both security blocks set, the podman quadlet, docker compose in HA and standalone (standalone drops the redundancy block and its PSK secret reference), container env pairs for HA (no `timezone`, so no TZ pair) and standalone (`timezone` set, so the TZ pair is present), the quadlet and compose forms of the opt-in health check, the CR with an explicit pullPolicy plus podAnnotations/podLabels, the CR with node and pod affinity alongside the legacy anti-affinity term, and the CR with loadBalancer annotations, node labels and tolerations (values carrying a colon and a URL, which survive only because both halves are quoted). The two secret-script goldens went with `render.SecretScript`: no container artifact carries a secret value any more |
| `TestArtifactsCarryNoSecrets` | The externalization guard: with distinctive values in `semp.adminPass`, `redundancy.psk` and an additional user's password, no deployment artifact on any platform (broker CR, quadlet, compose file) contains any of them -- each references the secret by name and `broker deploy` supplies the value |
| `TestContainerSecretsRedundancy` | HA lists both secrets in a fixed order with the expected broker settings, `FilePathKey`/`MountPath` derive the file form both engines use (the mount is named after the setting, not the host-side secret), and standalone lists the admin password only (no mate link, so no PSK secret). An encrypted server-certificate key adds a third secret reaching the broker as `tls_servercertificate_passphrasefilepath`, and only when the passphrase is actually set |
| `TestContainerSecretNamesAreHostScoped` | The de-confliction: the host-side name is `<container.name>-<suffix>` (the default name keeps the historical `solace-admin-password`), the in-container target and path never carry that prefix, and `EnvVar` maps `.`/`-` to `_` and prefixes a leading digit so the name stays exportable |
| `TestAdditionalUsersReachBothHalves` | An extra user's password becomes a per-host secret named after it (`ConfigKey` naming the env-file key), while its access level and `*filepath` pointer ride the env pairs and the password does not |
| `TestQuadletHealthCmdEscapesPercent` | systemd expands %-specifiers in every unit assignment, not just the quoted `Environment=` ones, so a percent-encoded character in a probe URL is doubled or the line is dropped and the health check silently disabled. Quotes and backslashes stay untouched there (the value is unquoted and podman splits it itself), and compose keeps the percent literal since it has no specifier expansion |
| `TestHealthCmdDefaultsToReadiness` | An enabled block with no cmd polls `/health-check/readiness` on 5550, and an explicit cmd wins |
| `TestSecretPreflight` | The precondition `broker deploy` and `broker generate` share: an empty secret value is refused up front, naming the field. Standalone needs no PSK, so an empty one is fine there |
| `TestParsePort` | Port entries across the `name=container`, `container:service`, and `/PROTO` forms |
| `TestParseToleration` | Toleration Equal (`key=value:effect`) and Exists (`key:effect`) forms |
| `TestQuadletEscape` | systemd `Environment=` escaping of `%`, `"`, and `\` |
| `TestScalingReachesContainersAsEnv` | Every scaling knob reaches docker and podman as a container environment variable, carrying the env file's values, including an explicit `0`, which is a real setting rather than an absent one |
| `TestScalingReachesK8sAsSpecOnly` | The other half of the delivery split: on k8s the same settings are CR fields under `spec.systemScaling` and never pod environment variables, the spool size is spelled `maxSpoolUsage` there, and the container spelling appears nowhere in the CR |
| `TestScalingTierReachesEveryArtifact` | One tier value decides the CPU cap in all three artifacts: the broker CR's `messagingNodeCpu`/`messagingNodeMemory`, compose's `cpus:`/`mem_limit:`, and the quadlet's `PodmanArgs=--cpus=`/`Memory=`. It uses 100000, which is no platform's default, so the value is proven read rather than hardcoded -- the goldens only ever show the default tier |
| `TestContainerMemOverrideReachesArtifact` | The asymmetry survives to the artifact: an overridden `container.mem` reaches compose while the CPU stays the tier's |
| `TestCustomVolumeMountRendersTheCRArray` | The translation at the boundary: the env file keys on this tool's lowercase role word, the CRD constrains `customVolumeMount[].name` to a capitalised enum, and the order is fixed rather than map order -- a Go map iterates randomly, and this renders into a CR that is diffed and re-applied, so an unstable order would look like a change on every deploy |
| `TestNoCustomVolumeMountEmitsNothing` | The block is absent, not empty: an empty array is a different statement from an unset field, and the operator reads them differently |
| `TestExtraEnvVarsSecretFollowsTheUsers` | The CR half of `semp.additionalUsers`. The field names a Secret, so emitting it with no users would point the operator at an object that does not exist and fail the pod on a mount the deployment never needed -- the same rule `preSharedAuthKeySecret` follows. Also pins that no password reaches the CR |
| `TestPreSharedAuthKeySecretFollowsTheKey` | The Kubernetes half of the PSK asymmetry. The field names a Secret, not a value, so emitting it unconditionally would point the operator at a Secret carrying no `preshared_auth_key` entry -- breaking a deployment the operator would otherwise have keyed itself. Absent means "generate your own", and `pskEnv` counts as a configured key |
| `TestSecretTargetsAreAbsolutePaths` | The guard for the silent half of the secrets-directory move. Both engines resolve a BARE `target=`/`target:` under their own `/run/secrets`, so if `Target` ever went back to returning just the setting name the file would be created, the container would start, and the broker would look under `secretMount` and find nothing -- with no error from the engine, the tool, or the broker's own startup |
| `TestSecretsAndCertDoNotNest` | No secret's file can collide with the server certificate's mount, which is the hazard the old layout carried: the cert lived INSIDE the secrets directory, so a setting named `tls.crt` would have been the same path. They are separate trees now, and this asserts it rather than trusting it |
| `TestComposeProjectIsDeclaredNotDerived` | With no top-level `name:`, compose takes the project name from the basename of the directory holding the file -- so it changes when the artifact is generated elsewhere or the directory is renamed, and the previous project's containers, network and volumes become orphans `down` no longer finds |
| `TestComposeProjectFoldsToComposesGrammar` | Compose's project-name grammar is narrower than the container-name grammar config enforces: it lowercases and admits only `_` and `-`, while `My.Broker` is a container name both engines accept. Folding is the right trade -- refusing a perfectly good container name over a compose spelling rule would not be |
| `TestUnresolvedTierOmitsLimits` | The renderers' fail-safe branch. A `Config` built in code -- what the executors are handed -- carries no tier, and all three artifacts must then omit the limits rather than emit an empty `cpus:`/`--cpus=`/`messagingNodeCpu:`, which the engines and the CRD would reject |

### servercert_test.go

The server certificate is delivered differently by each engine and named differently by
the broker, and every one of those differences fails silently when it is wrong -- an
unknown environment key is not an error, and a missing bind-mount source becomes an empty
directory. These pin each one by name rather than leaving it to a golden, which `-update`
would bless away.

| Test | What it covers |
| --- | --- |
| `TestServerCertFilePathKeyKeepsItsUnderscore` | A spelling neither engine nor broker would complain about getting wrong. Every credential setting appends a BARE `filepath`; the server certificate's carries an underscore, so deriving it from the generic suffix would produce a setting the broker does not read -- and nothing errors on an unknown environment key, so TLS would simply be off with nothing pointing at the cause |
| `TestServerCertIsADockerOnlySecret` | The one place the secret list differs by platform, and why it must. Docker can source a secret from the compose child's environment, so the bundle reaches the container with nothing written to this host; a quadlet unit cannot inline content, so podman gets the same bytes as a bind-mounted host file. If the certificate appeared in podman's list, `CreatePodmanSecrets` would load a PRIVATE KEY into podman's secret store |
| `TestServerCertificateReachesTheContainerOnBothEngines` | The named regression test for a property that previously had only goldens behind it -- a golden break is routinely answered with `-update`, which would silently bless the certificate disappearing from an artifact altogether |
| `TestServerCertBundlePathIsPosixAndUnderBaseDir` | The one expression the quadlet renderer and the Manager's writer must share. If they drifted, podman would find no file at the source and create a DIRECTORY there instead, and the broker would start with no certificate and no error anywhere |
| `TestFileBackedSecretIsExemptFromSecretPreflight` | A message that would otherwise be a lie: `SecretPreflight` blames an empty value with "set it in the env file", which for the certificate is wrong -- `tls.cert` IS set, the bytes just live on the host. Readability is enforced where the files are read instead |
| `TestPodmanNeverGetsAFileBackedSecret` | The structural guard that replaced the `SecretScript` refusal, and is stronger than it: `ContainerSecrets` never OFFERS a file-backed entry for podman at all. A refusal can be bypassed by a new caller; an absent entry cannot. The certificate still reaches a podman broker -- written to `baseDir` at `0600` and bind-mounted by the unit -- which is exactly why it must not also be in the store, where teardown would try to delete it |
| `TestComposeEscapesTheDollarSign` | A value compose would otherwise consume before the container saw it. Compose interpolates `$VAR` and `${VAR}` across the whole document from its own environment and turns `$$` back into one `$`. The case that makes this necessary rather than defensive is the health-check command -- a shell command the ENGINE runs inside the broker, where `$(...)` and `$VAR` are legitimate and the config gate permits them on purpose |
| `TestComposeQuotesIdentifierScalars` | A YAML 1.1 misread the name grammar cannot prevent: `yes`, `no`, `on`, `off`, `true`, `false`, `0123` and `1.5` are all legal container names to both engines, but bare in a compose document they read as a boolean or a number -- so the service key, `container_name` and hostname stop being the string the operator wrote |

---

## internal/tools/vulnjudge

The dev-only judge the `scan` task pipes govulncheck JSON through. 11 tests.

### main_test.go

| Test | What it covers |
| --- | --- |
| `TestJudge` | The core policy: a called vulnerability with a fix exits 1, one with no released fix warns and exits 0, and uncalled findings are ignored |
| `TestJudgeModuleFixHint` | A fixable module vulnerability advises `go get <module>@<version>` rather than a toolchain bump |
| `TestJudgeNoFindings` | A clean scan exits 0 with the no-vulnerabilities message |
| `TestJudgeTolerantOfBOM` | A UTF-8 BOM (what PowerShell's encoders emit) does not break decoding |
| `TestJudgeMalformedInput` | Truncated or non-JSON input exits 2 with a malformed-JSON message rather than reporting success |
| `TestJudgeEmptyTrace` | a finding with zero trace frames is folded into the 'uncalled' bucket instead of panicking on Trace[0] a few lines later |
| `TestJudgeSortsFixableByID` | report ordering across two fixable vulns given out of ID order is deterministic (map iteration order can't leak through) and also reaches plural()'s multi-count branch ("2 vulnerabilities") |
| `TestJudgeUnknownModule` | a called finding whose trace carries a function but no module still prints an actionable placeholder instead of a blank field |
| `TestRunUsageError` | run prints the usage message to errOut (never out) and returns code 2 for any argument count other than exactly one |
| `TestRunUnreadableFile` | an unreadable path returns code 2 and names the file in the errOut message rather than handing a zero-value byte slice to judge |
| `TestRunHappyPath` | run wires judge's report to out and judge's exit code to its return value -- the run/judge integration, not just judge in isolation |

---

## internal/abbrev

The leaf package that owns the RULES for every approved short form -- not the words.
Sets are declared by whoever owns the vocabulary (`config` declares roles and platforms,
`cli` declares command names), and this is where what makes a short form legal is
decided once: lowercase letters or digits, strictly shorter than the word it stands for,
claimed by one word only, and never shadowing a canonical name. Every rule is checked at
construction and fails as a panic, so a bad declaration cannot reach a release -- which
also means every test that builds a set is a validation test. 8 tests.

### abbrev_test.go

| Test | What it covers |
| --- | --- |
| `TestExpandResolvesBothSpellings` | The one lookup callers get: a canonical word resolves to itself, so `ParseRole` and `ParsePlatform` each stay a single branch instead of a membership test plus a lookup. Case, leading space and an unknown word all miss |
| `TestOrderIsDeclarationOrder` | Nothing sorts the entries. The order is meaningful in every set the tool declares -- redundancy order for roles, the order platforms are offered in -- and it is the order errors and the reference print, so an alphabetical `Names()` would silently reword both. Also pins that an entry may carry no short form at all |
| `TestAccessorsReturnCopies` | The sets are package-level and shared by the parsers, the CLI wiring and the docs generator, so a caller mutating a returned slice would be mutating a global. `Entries`, `Short` and `Names` all hand back copies |
| `TestShortOfUnknownWord` | What `applyAliases` relies on: it asks the set about every command in the tree, most of which have no entry, and must get nil rather than a panic |
| `TestList` | The error sentence exactly, because `ParsePlatform`'s message IS this rendering and `TestParsePlatformErrorTeachesCanonical` reads it back. Covers three entries, a word with no short form, one entry (no conjunction) and an empty set |
| `TestKind` | The label errors and doc headings use |
| `TestNewPanicsOnAWiringBug` | Every rule, one case each: empty kind, empty canonical, a canonical declared twice, one short form claimed by two words, a short form shadowing a canonical declared LATER (which is why canonicals are claimed in a pass of their own), an empty short form, a short form that is not shorter, and the charset -- uppercase, a shell metacharacter, a space |
| `TestDigitsAreAllowed` | Keeps the charset from being letters-only by accident: a product word ending in a number is exactly the case an abbreviation is worth having for |

---

## internal/output

The leaf package every stdout/stderr convention lives in: the `==> ` progress voice, the
fixed-width `[ OK ]`/`[WARN]`/`[FAIL]`/`[INFO]` tags, and the report primitives (bordered
sections, aligned key/value blocks, tables), the `Level` type the k8s reports use to grade a
row, and the loud `Warning` block a scrollback-worthy alarm (like an operator downgrade) needs.
16 tests.

### output_test.go

| Test | What it covers |
| --- | --- |
| `TestTagsShareOneWidth` | The invariant the tag set exists for: every tag is bracketed and the same width, so a column of mixed outcomes cannot jog sideways. It is what stops a 7-character `[ERROR]` (which `[FAIL]` replaced) coming back |
| `TestStepKeepsTheArrowPrefix` | `Step` emits `==> <msg>` verbatim -- the one string README.md and the `-v/--verbose` help text both promise |
| `TestLeveledLinesCarryTheTagAndNoArrow` | The house rule made testable, per level: a stderr line is `==> <phase>` OR `[TAG ] <msg>`, never both, so the old `==> [WARN] ...` double-tag cannot reappear |
| `TestLineAddsNoPrefix` | `Line` stays raw -- it is what the CLI hands Cluster/Manager/Ops as their line sink |
| `TestNewFuncRoutesThroughTheGivenLineSink` | The seam the three entry types use: a Sink built over another Sink's `Line` still adds its own prefix exactly once |
| `TestNilSinksDiscard` | A nil `*Sink`, a Sink over a nil writer and a Sink over a nil func all stay silent across every method, matching the nil tolerance of the `logf` helpers they replaced |
| `TestSectionPadsToWidth` | The report header shape: padded to `SectionWidth`, and a title too long for the border is printed whole rather than truncated |
| `TestKVBlockAlignsOnTheLongestKey` | The point of the helper -- the column is computed from the longest key in the block, so no caller hand-counts spaces -- plus no stray line for no rows |
| `TestKVRowHonorsAnExplicitWidth` | The multi-method report case (the `check` reports, whose rows come from four methods): rows line up because they name one width, and an over-wide key still gets its separating space |
| `TestTableComputesColumnWidths` | Per-column widths from header and rows together, no trailing whitespace on the last column, and header-only/ragged-row inputs neither panic nor mis-pad |
| `TestLevelTagsMapToTheHouseTags` | `Level.tag()` maps every level to its bracketed tag, and an unknown/zero `Level` reads as `[INFO]` rather than defaulting to a false `[ OK ]` pass |
| `TestLevelOrderKeepsSkipBelowFail` | The ordering the "worst of" idiom relies on across the k8s reports (`if row.Level > lvl`): `Info < OK < Warn < Skip < Fail`, so a row nobody checked (Skip) can never outrank and mask a real Fail in a folded summary |
| `TestKVRowAtLeadsWithTheTag` | `KVRowAt` leads each row with its Level's tag, ahead of the aligned key column and the house `" : "` separator |
| `TestTaggedTableAlignsUnderTheTagColumn` | `TaggedTable`'s header sits over the cell columns rather than the tag column, widths are computed from header and rows together, a header-only table still prints its header, and a headerless table prints rows with no stray blank line |
| `TestWarningIsFindableInAScrollback` | The loud alarm block (used for an operator downgrade, say): surrounded by blank lines, a `!!!` border distinct from a report section's `===` (an alarm must not look like a heading), the title in the header padded to `SectionWidth`, prose indented two spaces, a pre-indented body line nested one level deeper, and an empty body entry rendering as a real blank line |
| `TestWarningTitleLongerThanTheBorder` | A title longer than `SectionWidth` prints whole rather than being truncated, matching `Section`'s own non-truncation rule |

---

## internal/examples

The embedded env-file templates `solace-util examples` prints, and the generator for
`env/sample.yaml`. The package itself is stdlib-only; the test is what imports
`internal/config`, because the invariant worth pinning is not "the bytes are there" but
"the bytes still load".

Each case writes the template to a temp file and drives the real loader, since
`config.DetectPlatforms` reads the RAW file -- a decoded `Config` cannot say which
sections were present.

### examples_test.go

| Test | What it covers |
| --- | --- |
| `TestSampleYAMLMatchesTheFullExample` | The drift gate between the template the binary prints and the copy committed at `env/sample.yaml`: byte-identical, or `test` fails naming the first differing line. Also the generator -- `-update` rewrites the sample from `assets/full.yaml`, which is why `regen` runs this package before `internal/render` and `internal/k8s` |
| `TestEveryExampleLoads` | Every template, under every platform it declares, survives `config.Load` -- so a schema change that invalidates a shipped example fails here instead of at a user's first deploy |
| `TestStartersDeclareExactlyOnePlatform` | What makes a starter usable with no flags: one declared section, so `resolvePlatform` picks it silently and the written file needs no `--platform` |
| `TestFullDeclaresEveryPlatform` | The other half of that contract -- the schema example populates all three sections so every key has one place to look it up, which is exactly why it is the one example that has to be told which platform to use |
| `TestExamplesCarryNoRealSecret` | The S3 check on shipped text: every template's `semp.adminPass` is a CHANGE-ME placeholder, so an unedited copy is refused by the broker rather than deployed with a password that is public in this repo |
| `TestGetRejectsAnUnknownName` | The error boundary: a typo is answered with every name it could have meant, and a platform abbreviation is NOT accepted here -- expanding it is the CLI layer's job, at the point it reads its own input |
| `TestEveryNameResolves` | The declared table against the embedded files: a name added without its asset, or with a path the `//go:embed` pattern misses, fails here rather than panicking in front of a user. Also that every entry has a description for the listing |
| `TestListNamesEveryExample` | The error and help phrasing comes from the table, never from a list typed beside it -- the rule `internal/abbrev`'s `Set.List` follows |

---

## Fixtures and doubles

There is no shared `testutil` package by design -- each package keeps its own small doubles
next to the tests that use them. Reuse the one in your package rather than hand-rolling a
new fake.

### Shared env fixtures

- **`env/sample.yaml`** is the one fixture for goldens and CLI end-to-end runs. It is a valid
  config for all three platforms, so `internal/render`, `internal/k8s`, and `internal/cli`
  all load it (as `../../env/sample.yaml`) instead of maintaining separate fixtures. It is HA
  with TLS configured. It is also **generated** -- from
  `internal/examples/assets/full.yaml`, the template a bare `solace-util examples` prints --
  so a fixture change is an edit to that asset followed by `regen`, never an edit here.
- Tests needing a clean single-broker pass write their own minimal env to a temp file:
  `writeStandaloneEnv` (k8s-shaped) and `writeCtrStandaloneEnv` (container-shaped, needs a
  `redundancy:` block) in `internal/cli/cli_test.go`. `writeRuntimeEnv`
  (`internal/cli/allowcommand_test.go`) is the same idea parameterized by `kubernetes.runtime`, for
  driving one hostile or wrapped command through the whole CLI.
- **`guardConfig`** (`internal/config/execguard_test.go`) is a config that validates cleanly
  on every platform with only the command fields left to vary, so a `Validate` failure in the
  execution-guard tests can only have come from the guard.
- **`wrappedCfg`** (`internal/k8s/runtime_test.go`) and **`wrappedCtrCfg`**
  (`internal/container/runtime_test.go`) carry a chained runner (`microk8s kubectl`,
  `sudo -n docker`) WITH the operator approval the guard requires; `unapprovedCfg` /
  `unapprovedCtrCfg` are the same values without it, for the refusal tests.
- **`internal/convert/testdata/legacy-k8s.env`** is the legacy-format fixture, with
  `legacy-k8s.yaml.golden` as its expected output. It deliberately does not point at
  `bash/env/sample`: the whole `bash/` tree is gitignored, so that path is absent on a fresh
  checkout and CI could never run the test. The container flavour uses an inline fixture
  (`ctrEnv`) in the same file, and `k8sEnv` is an inline minimal-but-complete k8s source
  for tests that need a conversion free of "incomplete" warnings.

### Per-package doubles

| Package | Double | Purpose |
| --- | --- | --- |
| internal/broker | `fakeTransport` (broker_test.go) | Records every upload/run/output and answers via a `responder` func |
| internal/broker | `newTestOps` (broker_test.go) | Builds an `Ops` with a buffer sink and zero poll interval so nothing sleeps |
| internal/broker | `runErrTransport` (coverage_test.go) | Embeds `fakeTransport` but fails `Run`, reaching the best-effort cleanup branches |
| internal/broker | `uploadErrTransport`, `downloadErrTransport`, `runErrMatchTransport` (coverage_test.go) | The same embed-and-override shape for the transport methods `fakeTransport` always succeeds at: failing `Upload`/`UploadFile`, failing `Download` (optionally only for one remote path, so a bundle fetch can fail while the archive succeeds), and failing `Run` only when an argv predicate matches (isolating a best-effort cleanup failure from an earlier fail-loud one). Each still records the call |
| internal/broker | `removed` (broker_test.go) | Whether an uploaded script was deleted afterwards -- `removeCLI` issues `rm -f` through `Run`, so removals land in `runs`, not `outputs`. Matters for any script whose body carries a secret |
| internal/broker | `seqTransport`, `newLocalOps`, `rd`, `sempOK`, `isCurl`, `curlCalls` (verify_local_test.go) | Scripts a sequence of `show redundancy` readings for the primary-driven HA ops; the responder's curl branch answers the mate SEMP channel with the canned `sempOK` reply, and `curlCalls` extracts the recorded SEMP execs for URL/stdin assertions. `localCfg` carries the mate IPs and admin creds the channel needs |
| internal/broker | `serverCert` (semp_test.go) | The `config.TLS{Cert, CertKey}` pair that makes `sempPort`/`MateSEMPPreflight`/`MateRevertActivity` prefer the mate's TLS listener; every TLS-expecting case sets it explicitly since a plaintext broker is what every other test already covers |
| internal/broker | `assertNoPasswordInArgv` (semp_test.go) | The shared S3 boundary check reused by every mate-SEMP test (plaintext and TLS alike): no argv token may carry the admin password or the RPC body |
| internal/k8s | `recRunner` / `rrCall` (transport_test.go) | Capturing `engine.Runner` with `outQueue` and `runErrQueue` for scripting multi-step ops, plus `outErrQueue` (per-`Output` errors, so one read in an op can fail while an earlier one succeeds) and `runInputErr` (fails `apply -f -` / `delete -f -`). Both queues fall back to the blanket `outErr`/`runErr` once drained, so a test that sets only those behaves as before. `canI`/`canIErr` answer `Cluster.Preflight`'s `auth can-i` probe out of band (default: permitted) so it never consumes a queued read written for a different call, and `afterPreflight` asserts the probe came first and returns the calls after it |
| internal/k8s | `haCfg`, `saCfg` (names_test.go), `adminCfg` (prep_test.go), `loadK8s` (secrets_test.go) | Config builders |
| internal/k8s | `checkGolden`, `-update` flag (secrets_test.go) | Golden comparison for the whole package |
| internal/k8s | `loadFixture`, `testdata/*.json` (inspect_test.go) | Decodes a REAL object captured from a live minikube deployment (operator 1.4.2, broker 10.26.0.8755), trimmed to the fields the decoders read -- the shapes were written against these bytes rather than the schema documentation, which is what caught the broker CR's actual condition types and string-typed booleans |
| internal/k8s | `fixtureRunner`, `reportCluster` (statusreport_test.go) | Answers every `get ... -o json` from the same captured `testdata/` fixtures `inspect_test.go` decodes directly, optionally failing one kind (`failOn`) or answering an empty list (`missing`) to drive the report's degrade paths; `reportCluster` also fixes `Cluster.Now` 6m40s after the captured broker's creation timestamp, matching the live capture's own AGE |
| internal/k8s | `operatorDeployJSON` (operatorversion_test.go) | Builds the `kubectl get deployment --all-namespaces -o json` reply `installedOperatorImage` decodes, naming only the image field it reads -- scripted through `recRunner`'s `outQueue` so `confirmNoDowngrade` is driven by the exact query shape production issues |
| internal/cli | `renderCommandDocs`, `-update` flag (commanddoc_test.go) | Renders `docs/commands.md` from the cobra tree; the doc is the golden |
| internal/cli | `renderAbbrevDocs`, `treeShorthands` (abbrevdoc_test.go) | Renders `docs/abbreviation.md` from the command/role/platform sets plus every single-letter flag in the tree. Shares commanddoc_test.go's `-update` flag, so one `regen` rewrites both goldens |
| internal/abbrev | `sample` (abbrev_test.go) | A set with the shape every real one has -- ordered canonicals, one short form each, one word deliberately carrying none. Build ad-hoc sets with `New` only for the malformed cases |
| internal/container | `capRunner` / `failOn` (transport_test.go) | Capturing runner whose `fail` (Run family) and `outFail` (Output family) hooks error on a targeted command, driving each error-wrap branch -- `failOn("info")` is how a failed engine preflight is injected. `outFail` exists because the blanket `outErr` cannot single out one of two probes in the same call. `capCall.env` records the extra environment of a `RunEnv` call, which is how the docker secret path is asserted |
| internal/container | `containsStr`, `maskedKeys` (manager_test.go) | Exact-match lookup in a captured environment, and `engine.MaskEnv` for failure messages -- a test diagnostic must not print a secret either |
| internal/container | `newEchoMgr`, `newCapMgr`, `ctrCfg` (manager_test.go) | Manager over dry-run Echo, or over the capturing runner for real file writes. `Manager.Confirm` is the injectable restart prompt (nil declines, which is what a non-interactive run must do) and `Manager.Restart` is the `--restart` pre-approval |
| internal/container | `assertMode` (manager_test.go) | Permission-bit assertion for the artifacts the manager writes; skipped on Windows, which carries no POSIX mode |
| internal/cli | `runRoot`, `capture`/`captureStdout`/`captureStderr` (cli_test.go) | Builds a fresh command tree per call and captures a standard stream through a pipe |
| internal/cli | `runRootWith` (cli_test.go) | `runRoot` with a hook to configure the `App` before `Execute` -- how the confirm-prompt branches are driven deterministically instead of depending on the test process's own stdin. `runRoot` delegates to it with a nil hook, so it is unchanged for every existing test |
| internal/cli | `opRunner` / `opCall` / `opFailOn` / `opFailOnCount` (cli_test.go) | A fake `engine.Runner` whose failure is targeted by argv substring (or by the Nth matching occurrence, for the repeated identical `apply`/`delete` calls in `broker deploy` and `broker remove`). Ported from internal/container's `capRunner`/`failOn`; this is what makes the orchestration-abort tests possible -- assert step N fails and step N+1 never ran. `Output` answers the k8s `auth can-i` preflight "yes" unless a test supplies its own `output`/`fail` for that call, so the op-level tests stay about the work they were written for |
| internal/cli | `loadDirect`, `healthyShowRD` (cli_test.go) | Loads a config from an inline YAML body for tests that build an `App` directly with a non-Echo runner, and a canned `show redundancy` transcript that satisfies `broker.primaryRedundancyUp` so a poll succeeds on the first read |
| internal/cli | `bashEnv`, `writeBashEnv` (cli_test.go) | Minimal legacy env file for the convert command tests |
| internal/cli | `replEnv`, `replApp`, `siteNamed` (replication_test.go) | A DR pair whose two sites are reached DIFFERENTLY -- one over its own cluster CLI, one over SEMP -- so a dispatch that ignored `via` fails on one site rather than passing on both |
| internal/convert | `strictDecode` (convert_test.go) | Re-reads generated YAML with `KnownFields(true)`, so an emitted key that is not in the schema fails the test |
| internal/convert | `ctrEnv`, `convertOK`, `hasWarning` (convert_test.go) | Container-flavoured legacy fixture and warning assertions |
| internal/config | `envTree`, `writeTempYAML`, `minimalK8s` (config_test.go) | Real temp-dir fixture trees for path resolution and loading, plus the smallest valid k8s document the secret-reference tests append a body to |
| internal/broker | `scriptedMate`, `twoMates`, `st` (switchplan_test.go) | A MateChannel whose per-VPN state the test controls and which records every write, with an `onRead` hook so a test can move a role at exactly the moment the code re-reads -- which is how the stale-state refusal is provable without a broker |
| internal/broker | `fakeRun`, `softwareBanner`, `mateReplies` (matechannel_test.go) | A CLI runner that records every script name and body and replies from a table keyed by script name, so a MateChannel test needs no transport and no broker |
| internal/k8s | `replSite`, `sempSite`, `replCfg`, `unconfiguredReplication` (matechannel_test.go) | A kubernetes-reached DR site whose cluster is carried in its own `command`, its SEMP-reached counterpart for the password-read path, both attached to `loadK8s`, plus the smallest banner-plus-empty-mate reply a preflight accepts |
| internal/broker | `showReplicationSoftware`, `showReplicationAppliance`, `mateSoftware`, `mateAppliance` (replication_test.go) | Embedded `show replication` reports in both platform layouts, transcribed from the real captures with synthetic addresses. Embedded rather than read from `semp/`, which is gitignored -- a test that only read the real files would skip in CI, exactly where it matters most |
| internal/config | `validReplication`, `replConfig` (replication_test.go) | A complete, well-formed DR pair -- one site reached over its own cluster CLI, the other over SEMP -- attached to `validK8sConfig`. The two sites deliberately differ in mechanism, so a bug that only reads `via.kubernetes` cannot pass |
| internal/config | `controlCharCases` (config_test.go) | The shared table of malformed control-character values (newline, CR, NUL, DEL, ...) every credential and additionalUsers-password rejection test drives against its own field |
| internal/engine | `helperCommand` + `TestHelperProcess` (runner_test.go) | Re-invokes the test binary as a fake external command |

### Injectable seams

Small external effects are seamed as function fields so they are testable off a Linux host.
Override them on the struct after construction:

| Seam | Default | Where |
| --- | --- | --- |
| `Manager.Resolve` | `net.LookupHost` | internal/container -- DNS probes in `Check`/`PrepHost` |
| `Manager.Geteuid` | `os.Geteuid` | internal/container -- the rootless/rootful guard (returns -1 on Windows, which skips it) |
| `Ops.Hostname` | `os.Hostname` | internal/broker -- node-role detection in `LocalRole` |
| `Ops.Platform` | `""` (zero value) | internal/broker -- resolves the mate's SEMP port from the bridge network.ports mapping (`sempPort`); set by `ctrOps`, left zero on k8s, which never uses the SEMP channel |
| `App.Interactive` | `stdinCanAnswer()` -- stdin is usable at all | internal/cli -- whether a run may ASK. Gates `confirmDelete`/`confirmLayer`/`confirmRestart`, so every prompt branch guarding a destructive action is testable -- including the one that KEEPS a layer when nobody answers. Whether the question was ANSWERED is a separate fact, reported by `promptLine`'s second return |
| `App.PromptIn` | `os.Stdin` | internal/cli -- where a confirmation answer is read from. Always stdin in production: asking on one stream and reading another is what made the console-handle attempt hang |
| `os.Stdin` (package `TestMain`) | an already-closed pipe | internal/cli -- set for the whole package so a test that reaches a prompt without `App.PromptIn` gets EOF instead of blocking on the console the suite was launched from. Tests needing a real answer swap `os.Stdin` themselves |
| `App.kubeContext` | `""` (unresolved) | internal/cli -- the kubeconfig context `announceKubeContext` resolved at load, repeated by every destructive Kubernetes prompt (`k8sWhat`/`k8sContext`). A test sets it directly the same way it sets `Interactive`/`PromptIn`, since a test-supplied runner skips the real `config current-context` lookup |
| `Ops.LocalAddrs` | `defaultLocalAddrs` (`net.InterfaceAddrs`) | internal/broker -- the address pass of `DetectRole`. `newTestOps` defaults it to an EMPTY set, so no test's "this host matches nothing" case depends on how the machine running the suite is numbered |
| `App.LocalAddrs` | `nil` (`ctrOps` then leaves `Ops.LocalAddrs` at its own default) | internal/cli -- the same injection one level up, for the container role detection behind `broker deploy` and `broker generate` |
| `App.Hostname` | `nil` (`ctrOps` then leaves `Ops.Hostname` at its own default, `os.Hostname`) | internal/cli -- same shape as `NewRunner`/`Interactive`/`PromptIn`: a test injects a fixed hostname so `containerRole` (the role detection behind `broker deploy` and `broker generate` on docker/podman) is testable without depending on the suite's own host being named after a broker node |
| `engine.Runner` | `engine.Exec` | Everywhere -- swapped for `engine.Echo` (dry-run) or a capturing fake |
| `Cluster.Log` / `Manager.Log` / `Ops.Log` | nil (discards) | The RAW progress line sink. It takes an already-formatted line and emits it verbatim: the `==> ` and `[TAG ] ` prefixes are added by the `internal/output` Sink the callee builds over it (`progress()`), so a test that captures this field sees exactly the line a terminal would. The CLI passes `lineSink()`; nil discards |
| `Cluster.Out` / `Manager.Out` / `Ops.Out` | `os.Stdout` | The report sink -- sections, key/value blocks, tables and per-item outcome lines. Wrapped by `report()` |
| `Cluster.Now` | `time.Now` | internal/k8s -- the clock the status reports' AGE column reads; fixed in tests (`reportCluster`) so it is reproducible against a captured fixture's timestamp |
| `Cluster.Confirm` | `nil` (refuses) | internal/k8s -- the operator-downgrade confirmation `OperatorApply` asks before applying an image older than the one already installed; the same nil-declines shape as `Manager.Confirm` |

Filesystem access is *not* seamed: tests use real `t.TempDir()` trees.

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

48 test files, 852 test functions. `TestHelperProcess` in `internal/engine` is not a real
test -- it is the os/exec helper-process shim, a no-op unless `GO_WANT_HELPER_PROCESS=1`.

| Package | Files | Tests |
| --- | --- | --- |
| internal/broker | 5 | 166 |
| internal/k8s | 15 | 162 |
| internal/cli | 8 | 139 |
| internal/config | 5 | 112 |
| internal/container | 4 | 110 |
| internal/tools/itest | 3 | 40 |
| internal/convert | 1 | 37 |
| internal/engine | 2 | 26 |
| internal/render | 1 | 17 |
| internal/output | 1 | 16 |
| internal/tools/vulnjudge | 1 | 11 |
| internal/abbrev | 1 | 8 |
| internal/examples | 1 | 8 |
| **Total** | **48** | **852** |

## Coverage

Last recorded run, from `scripts/logs/cov.log` (2026-08-21 10:30), total **96.1%**. Re-run
`cov` after any change; these figures go stale the moment tests move, and the previous
total is the floor the next run has to hold.

| Package | Coverage |
| --- | --- |
| internal/output | 100.0% |
| internal/tools/vulnjudge | 98.9% |
| internal/config | 98.2% |
| internal/render | 97.6% |
| internal/convert | 97.4% |
| internal/broker | 97.1% |
| internal/k8s | 95.8% |
| internal/container | 95.5% |
| internal/cli | 95.2% |
| internal/tools/itest | 91.4% |
| internal/engine | see below |

`internal/tools/itest` is the lowest by design: `main()` and the `crypto/rand` failure paths
are not reachable from a test, and the probe bodies' live halves are covered over a fake
`engine.Runner` rather than a real one.

**`internal/engine` is not currently measurable.** It reports `0.0%` in this run and
`100.0%` in the one before it, with every test passing both times -- so the figure is a
measurement artifact, not coverage that vanished. The package's `Exec` tests re-execute the
test binary as a child process (`TestHelperProcess`/`helperCommand`), and a coverage-
instrumented child can clobber the parent's profile; the run also slows to 17s when it
happens. Until that is fixed, treat the total above as understated by roughly engine's
share, and do not read `96.9%` as a drop from `97.2%` -- the difference is almost entirely
this artifact.

---

### examples_test.go

`examples` is the one command that must work before an env file exists, so these
run it with no `--env` at all (and once with a deliberately missing one). It writes
files, so the `--out` cases all target `t.TempDir()`.

| Test | What it covers |
| --- | --- |
| `TestExamplesListsWhatItHas` | Bare `examples` names every template on STDOUT, so `examples > list.txt` captures the list. It acts on nothing, which is why the bare form prints rather than failing the way a verb group does |
| `TestExamplesEmitsToStdout` | Each name emits the package's bytes verbatim and nothing else, so the template is the command's only stdout -- the same contract `generate` keeps |
| `TestExamplesAcceptsAPlatformAbbreviation` | The CLI-wide abbreviation rule at this site: `examples dk` and `examples docker` emit the same template, because the word is expanded by the same `config.ParsePlatform` that reads `--platform` |
| `TestExamplesWritesOutFile` | `--out` writes the template verbatim, keeps stdout empty, names the file on stderr, and -- the point of the whole command -- what it wrote loads through `config.Load` as-is |
| `TestExamplesRefusesToOverwrite` | The destructive path: an existing `--out` file is kept and the error names `--force`, which then does replace it |
| `TestExamplesRejectsAnUnknownName` | `k8s` is neither a template name nor an accepted platform word: it is refused naming every alternative, and no `--out` file is created |
| `TestExamplesNeedsNoEnvFile` | Why the command is a bare cobra literal rather than a `leaf`: it runs clean against an env path that does not exist. A `wireExec` added here later fails this instead of failing a first-time user |

## internal/config

Config loading, defaults, validation, and env-file resolution, plus the `Command`
type behind the platform CLI overrides and the execution guard that decides what a
`Command` may be, the scaling block that sizes the broker on every platform, and the
platform vocabulary the CLI resolves against. 112 tests across 5 files.

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
| `TestAllowCommandIsNotASchemaKey` | The structural half of "the config author has no say": every plausible spelling of an allowlist key fails strict decoding, and none reaches the unexported field backing it |

### config_test.go

| Test | What it covers |
| --- | --- |
| `TestPlatformIsContainer` | `Platform.IsContainer` is true for docker/podman, false for k8s |
| `TestPlatformConstantsMatchSchemaSections` | Each `Platform` constant value is also a top-level section key in `Config`, reflected off the yaml tags. The subtree name and `--platform` read the constant, but the tag is a separate literal, so this is the only thing tying the word an operator types to the block they edit |
| `TestRedundancyEnabled` | Only the literal `yes` enables HA; `no`, empty, and junk do not |
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
| `TestApplyDefaultsK8sTLS` | TLS cert/key default only when `kubernetes.tlsServerSecret` is set |
| `TestApplyDefaultsDocker` | Docker defaults (runtime, compose mode, the compose command derived from the runtime, host network, admin user, container name) plus the shared `kubernetes.*` fields containers reuse |
| `TestApplyDefaultsPodmanRootful` | Rootful podman gets the system quadlet dir, no `--user`, `multi-user.target` |
| `TestApplyDefaultsPodmanRootlessXDG` | Rootless quadlet dir derives from `XDG_CONFIG_HOME`, with `--user` and `default.target` |
| `TestApplyDefaultsPodmanRootlessHomeDir` | Empty `XDG_CONFIG_HOME` falls back to the user home dir branch |
| `TestValidateK8sValid` | A fully populated k8s config validates clean |
| `TestValidateK8sMissingMandatory` | Every missing mandatory k8s field is named in one message, exact wording pinned |
| `TestValidateK8sBadUpdateStrategy` | `kubernetes.updateStrategy` enum is rejected loud |
| `TestValidateK8sAdminUserFixed` | Mirrors `TestValidateK8sMsgNodeCPURemoved` for a credential: the operator reads the fixed `username_admin_password` key, so a non-`admin` `admin.user` was silently ignored on Kubernetes and is now refused naming that key. `"admin"` and unset both stay legal (unset means `ApplyDefaults` fills it), and docker still accepts any name -- there the username drives the access-level setting, the password file and the SEMP login |
| `TestValidateContainerHA` | A valid HA container config validates for both docker and podman |
| `TestValidateContainerStandalone` | Standalone only requires `nodes.primary.name` among the node fields |
| `TestValidateContainerMissingMandatory` | Missing container fields (image, admin, all three node name/ip pairs) are all named |
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
| `TestValidateBadRedundancy` | `redundancy` enum is rejected loud |
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
| `TestValidateAdditionalUserPasswordCharsetIsK8sOnly` | The one platform-specific rule: k8s puts the password on a CLI line, so the characters the broker rejects there fail validation (naming the character, never the password), while the same env file stays valid for docker and podman, which mount the value as a file |
| `TestValidateAdditionalUsers` | On k8s and docker alike: a valid entry passes, and missing/invalid/duplicate usernames, the built-in `admin`/`monitor` names, a missing or invalid access level, and an empty password all fail. Two usernames differing only in `.`/`_`/`-` are refused too: they fold to one docker host variable name, which would feed one user's password to both |
| `TestValidateAdditionalUserClashesWithAdminUser` | The container-only clash: a listed user matching a configured `admin.user` is refused (two secrets would feed one broker setting) |
| `TestValidateCredentialControlChars` | `admin.pass`/`admin.monitorPass`/`nodes.psk`/`tls.certPassphrase` each reject a control character (naming the field, never echoing the value), driven off the shared `controlCharCases` table |
| `TestValidateAdditionalUserControlChars` | The same control-character rule for `admin.additionalUsers[].password`, which the k8s path also puts on a broker CLI line (`create username ... password ...`) |

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
| `TestValidateK8sMsgNodeCPURemoved` | Mirrors `TestValidateDockerRunModeRemoved`: `kubernetes.msgNode.cpu` still decodes but fails validation, so the operator gets a reason naming `scaling.maxConnections` and noting `mem` is unaffected, rather than a bare unknown-field error |
| `TestValidateMaxPoolRemoved` | `maxPool` would name the same broker setting as `maxSpoolUsageMB` under a platform-specific name; it is rejected on all three platforms naming the replacement, and an unset (zero) value does not trip the sentinel |
| `TestValidateContainerMem` | `container.mem` takes docker's and podman's own `b\|k\|m\|g` suffix: the likely mistake (a `Mi` quantity copied from `kubernetes.msgNode.mem`) is refused naming that trap, alongside bare numbers, decimals and unknown suffixes, while every legal form and the unset case pass |

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
139 tests across eight files.

Because the platform is a flag rather than the first word of a command, the
invocations here name it explicitly (`--platform docker`) rather than relying on
detection: `env/sample.yaml` declares all three sections, so an invocation without it
would hit the ambiguity error instead of the behaviour under test. Detection itself is
covered in `platform_test.go`, against fixtures written for that purpose.

### cli_test.go

| Test | What it covers |
| --- | --- |
| `TestEnvFileLookup` | `-e`/`--env` as the CLI wires it: `env/` fallback, base dir shadowing, no extension inference, the `==> env file:` echo, and long/short flag parity |
| `TestFirstArg` | `firstArg` on nil and populated slices |
| `TestFirstArgOr` | `firstArgOr` falls back on a missing or empty first argument |
| `TestNotImplemented` | The placeholder error names the command and says "not implemented yet" |
| `TestEmit` | `emit` writes bytes to stdout unchanged |
| `TestWarnAndStep` | `warn` and `step` write `[WARN]` / `==>` lines to stderr, now by delegating to the shared `internal/output` Sink rather than formatting the prefixes themselves |
| `TestAnnounceCommandsNamesResolvedBinaries` | The preamble that replaced the per-call `exec:` line: each binary the env file names is resolved and printed once as `==> using <name>: <path>`. k8s announces the cluster CLI; docker announces one line when `compose` is the runtime's own subcommand and two when a standalone `docker-compose` is configured; a name that resolves nowhere is skipped in silence, since a report must not invent a failure the first real execution already reports. Hermetic -- a stub binary is written into a temp dir put at the front of `PATH`, so the expected path is exact and no test host needs kubectl or docker |
| `TestBinaryAnnouncementWiring` | The same through the real command tree, with the stub named `kubectl` so it is the schema default and needs no `--allow-command` (which is itself refused where nothing executes). A real run (`check deploy`) announces before it works; the render-only `generate broker` and `generate operator` announce nothing, since neither one ever runs a binary |
| `TestVerboseFlagTracesEveryCommand` | `-v` prints `==> exec: <path> <args>` per call, a run without it prints none, and pairing `-v` with the test-only echo seam still works rather than fighting it, since Echo already echoes every command |
| `TestTreeStructure` | A representative set of command paths exists in the one unified tree, covering a top-level leaf, a group's child, and each applicability class |
| `TestEveryRunnableCommandIsWired` | The wiring that replaced the two `PersistentPreRunE` hooks: every command with a `RunE` (bar `convert`, `examples`, `version`, `completion`) carries the shared pre-run and the `--allow-command` flag. Missing either is invisible until that one command is run, so it is checked structurally |
| `TestGroupCommandsPrintHelpAndDoNothing` | The no-implicit-actions rule: a verb that owns objects (`check`, `smoke`, `prepare`, `deploy`, `config`, `start`, `stop`, `restart`, `status`, `logs`, `copy`, `generate`, `remove`, and `config`'s own `apply`/`delete`/`disable`) carries no `RunE`, so running it bare prints its own help and touches nothing -- proven by succeeding with no `--env` at all, since a runnable leaf would instead fail resolving the missing default env file |
| `TestFlagsRegistered` | Per-command flags are registered where expected: `--restart` on `deploy broker`/`deploy all`; `--delete-data`/`--no-prompt` on `remove broker`/`remove all`; `--delete-crd`/`--no-prompt` on `remove operator`; `--days` on `diagnostics`; `--input`/`--pod` on `cli`; `--pod` on `copy from`; `--pod`/`--dir` on `copy into`; `--all`/`--detail`/`--pod` on `status broker`; `--detail` on `status operator`; `--out`/`--force` on `examples` |
| `TestHelpNoConfig` | `--help` short-circuits before config load, so no env file is needed |
| `TestGenerateWired` | `generate`'s leaves, none of which contact the cluster or the container engine: `generate broker` renders what `deploy broker` would apply on EITHER family (the CR's `apiVersion:` on kubernetes, `services:`/`[Unit]` on docker/podman), plus `generate secrets broker` on both and `generate operator` on kubernetes. A plain `runRoot` with no echo seam is enough since nothing here executes |
| `TestCtrWiredDryRun` | Every container command safe to run against the HA sample drives clean over the echo seam, echoing the expected runtime/systemctl/mkdir command |
| `TestCtrRoleGuards` | Primary-only HA guards: leader and redundancy both reject backup and monitor hosts, bad roles error, and standalone self-skips -- all before any poll or SEMP call |
| `TestCtrConfigDryRun` | Container config steps run clean on a standalone env; cert/product-key-gated steps self-skip |
| `TestCtrExecCLIPathSeparator` | `opCtrExecCLI`'s used-as-is branch: a `cli --input` file argument containing a path separator is used as-is, not joined under the CLI scripts folder |
| `TestCtrErrorPaths` | Container `config apply` failures are actionable: no TLS configured for server-cert, no product keys configured, and a failed `check semp-login` |
| `TestCtrDiagnosticsDryRun` | Container `diagnostics` echoes its node-local gather/download sequence over the echo seam (isolated because it creates a diag dir) |
| `TestCtrRoleArgCount` | Role-taking commands reject a second positional argument |
| `TestCtrRoleHelp` | Role-taking commands expose `--help` without loading an env |
| `TestK8sWiredDryRun` | Every k8s command safe to run against the HA sample drives clean over the echo seam, with `+ kubectl` echoed on the acting paths and absent on the skip paths |
| `TestK8sStandaloneDryRun` | Redundancy-branching commands on a standalone env: the HA-only `config leader`/`smoke redundancy` self-skip, while `prepare secrets`/`prepare all`/`deploy all` run clean |
| `TestDeployAllDoesNotApplyOperator` | Pins the property that `deploy all` never installs the operator: it is cluster-scoped and shared between brokers, so only `deploy operator` does. Run on a standalone env to stay off the HA leader-assertion poll, which the echo seam cannot satisfy |
| `TestCheckDeployWarnsWhenOperatorAbsent` | `opK8sCheck`'s operator probe: `check deploy` is read-only, so a missing operator is reported as a stderr warning rather than failing the check itself |
| `TestStartStopRestartBroker` | The day-2 `start`/`stop`/`restart broker` verbs on both platform families: Kubernetes scales the statefulset(s) via a kubectl echo, containers start/stop/restart the container in place via a docker echo |
| `TestCLICommand` | `cli` has two shapes distinguished by a flag: bare, it opens an interactive session; with `--input`, it uploads and runs a script instead. Two H2 regression subtests pin `--pod` steering EITHER shape onto the named pod: `--input --pod backup` must not fall back to the primary (the old bug, where `[role]` and `--pod` were two disconnected branches), and bare `--pod backup` opens the session against that pod |
| `TestStatusBrokerFlags` | How `--all` and `--detail` compose on `status broker`: they widen the report along independent axes (every broker in the cluster vs. this env file's one; the static description vs. the running inspection) rather than one replacing the other, on both kubernetes and a container platform |
| `TestRemoveBrokerLayerContract` | The retained-layer contract on `remove broker`: persistent data is kept by default, `--no-prompt` proceeds while still keeping it, `--delete-data` deletes it, and a non-interactive run keeps it. Losing data always takes an explicit `--delete-data`, whatever else is on the command line |
| `TestRemoveOperatorLayerContract` | Mirrors `TestRemoveBrokerLayerContract` for the operator's CRDs: kept by default, since deleting them cascades to every broker in the cluster, and deleted only when `--delete-crd` names them |
| `TestRemoveFlagsCompose` | `--delete-data` and `--no-prompt` answer DIFFERENT questions -- what to do with the data, and whether to ask about anything -- so they must combine, not conflict. A fully unattended removal that also drops the data needs both |
| `TestSecretsNeverEchoed` | `check semp-login` over the echo seam shows stdin as a byte count and never the admin password |
| `TestConfirmFlagShortcuts` | `--no-prompt` confirms a removal; `confirmLayer` still keeps the layer under it, and `--delete-data` is what deletes -- all without reading stdin |
| `TestConfirmNonTTY` | Without a TTY and without `--no-prompt`, `confirmDelete` refuses, and `confirmLayer` with no flags set keeps the layer -- both with no prompt read |
| `TestConfirmDowngradeNeverUnattended` | M9: `confirmDowngrade`'s own gate -- non-interactive declines and tells the operator to re-run interactively (never advising `--no-prompt`, which this command does not register), `noPrompt` is not a side door (still declines non-interactively even when set), an interactive `yes` proceeds and an interactive `no` refuses |
| `TestDeployOperatorNoPromptStaysUnknownFlag` | The other half of M9: fixing `confirmDowngrade`'s dead `a.noPrompt` branch must not resurrect `--no-prompt` on `deploy operator` -- the flag was never registered there, so it must still fail as cobra's own unknown-flag error |
| `TestPromptYesNo` | Lenient delete prompt: `y`/`yes` in any case accept, everything else declines |
| `TestPromptYes` | Strict layer-deletion prompt: only an exact trimmed `yes` accepts; a bare `y` does not |
| `TestErrorPaths` | Global rejections: unresolvable env file, invalid node roles across container and k8s leaves (`deploy broker`, `generate broker`, `deploy all`, `logs broker --pod`), unknown `generate` target |
| `TestK8sGenSecretsWired` | The Secret manifests render through `generate secrets broker`, on the standalone env rather than the HA sample whose kubernetes.tlsServerSecret points at cert files absent from a checkout |
| `TestK8sGenOperatorSecretsWired` | `generate secrets operator` renders the operator's regcred alone; `generate operator` carries the `imagePullSecrets` reference but no `.dockerconfigjson`, which is the split that makes the bundle shareable; and an env file without kubernetes.imagePullSecret is refused loud naming the field |
| `TestGenSecretsRefusesEmptyValue` | The printed script invites execution, so `generate secrets broker` is refused when running it would create an empty secret -- while `generate broker` stays renderable, since the deploy artifact only references secrets by name |
| `TestGenNeverLeaksSecrets` | End-to-end: `generate broker` output on both container platforms omits the admin password, while `generate secrets broker` carries it (it is what creates the secret) |
| `TestConfigStepsDoNotLeakSecrets` | Drives each `config apply`/`config disable` step individually against a container fixture carrying every optional value (server cert/key, a domain CA, a product key) and asserts none prints the private key material to stdout -- `config` does not aggregate these steps into a re-runnable ordering, so each needs its own direct coverage |
| `TestConvertToStdout` | `convert` writes YAML to stdout and its warnings to stderr, so the artifact stays clean |
| `TestConvertToFile` | `-o` writes the file, a second run refuses to clobber it, and `--force` overrides |
| `TestConvertRoundTrip` | A converted file loads: `-e` against it drives a real command |
| `TestConvertErrorPaths` | Bad `--platform`, a missing source file, and a missing argument all fail loud |
| `TestVersionPrintsStampedValue` | `version` reports whatever the dev scripts' `-X` flag (or a test) set the package var to, verbatim -- the contract that aligns a release binary with its git tag |
| `TestVersionDefaultsToDev` | An unstamped build (plain `go build .` or `go test`) reports "dev" |
| `TestVersionIncludesToolchainAndPlatform` | Output carries `runtime.Version()` and GOOS/GOARCH, for support triage |
| `TestVersionRejectsArgs` | `version` takes no arguments |
| `TestBashEnvGivenToEnvFlag` | Pointing `-e` at a legacy bash file reports not-valid-YAML and names `solace-util convert` |
| `TestExecute` | `Execute()` builds the tree and runs `--help` without error |
| `TestK8sConfirmDeclined` | Every removal declines unattended: `remove broker`/`all`/`secrets`/`namespace`/`operator` without `--no-prompt` make zero cluster calls, using the App.Interactive seam instead of ambient stdin. secrets and namespace confirm like every other removal, and `remove namespace` takes everything living in the namespace, not only what this env file put there |
| `TestK8sPromptsNameNamespaceAndContext` | Table-driven over `remove broker`/`secrets`/`namespace`/`all`/`operator` and `restart broker`, with `a.kubeContext` set directly (the same seam as `App.Interactive`): every prompt now names the namespace and repeats the `(context <name>)` clause, except `remove operator`, which names the OPERATOR's own namespace (`k8s.OperatorNamespace`) instead of the broker's |
| `TestK8sPromptsOmitAnUnknownContext` | The other half: with `kubeContext` left empty, as an unresolved `announceKubeContext` lookup leaves it, `remove broker`'s prompt still names the namespace but drops the context clause entirely rather than rendering an empty `(context )` |
| `TestK8sRestartRollsOrTargetsOnePod` | H2 regression guard: `opK8sRestart` must keep reading `app.pod` RAW rather than through `podRole`, whose empty-defaults-to-primary would silently turn "restart every pod" into "restart the primary pod alone" -- no `--pod` rolls monitor/backup/primary in order, `--pod backup` touches only that one pod |
| `TestK8sRestartConfirmGate` | a non-interactive `restart broker` (rolling, or one `--pod` role) bounces nothing, and a bad `--pod` role is rejected before any prompt |
| `TestCtrRestartConfirmGate` | `opCtrRestartBroker`'s new `confirmDelete` gate, the container-side counterpart of `TestK8sRestartConfirmGate`: a non-interactive run bounces nothing and `--no-prompt` issues the restart, on both docker and podman -- `restart broker` used to run unconditionally on containers |
| `TestCtrConfirmDeclined` | a non-interactive `remove broker` on docker without `--no-prompt` issues zero runtime calls |
| `TestIsTTYClosedFile` | isTTY treats a stream it cannot Stat (a closed file) as non-interactive rather than risking a blocked prompt |
| `TestCtrManagerConfirmWiring` | ctrManager wires Manager.Confirm to confirmRestart, and a non-interactive session (via App.Interactive) declines without reading a prompt |
| `TestK8sLoginOutcomes` | a transport failure propagates as an error and a canned 200 OK response returns nil, the two real SEMP outcomes engine.Echo's fixed (nil,nil) can never produce |
| `TestCtrLoginOutcomes` | same as TestK8sLoginOutcomes for the container login path |
| `TestOpK8sDeployAllAssertsLeaderOnHA` | on an HA config, `deploy all` asserts the config-sync leader as its last step rather than stopping after `DeployBroker`. Unreachable via `runRoot`/`engine.Echo` (Echo's fixed empty output never satisfies Leader's poll), so it drives `opK8sDeployAll` directly over a fake Runner seeded with a healthy `show redundancy` transcript |
| `TestOpK8sDeployAllAborts` | `opK8sDeployAll`'s four error-return arms (Check, CreateNamespace, CreateSecrets, DeployBroker), each failed in its own sub-test with no later step's command issued afterward. There is no operator-apply step to abort before: the operator is cluster-scoped and installed on its own via `deploy operator` |
| `TestOpK8sPrepAllAborts` | same abort-ordering property as `opK8sDeployAll`, on `opK8sPrepAll`'s two error-return arms (CreateNamespace, CreateSecrets). Those two are the whole sequence: the operator is installed by its own command, and node labelling is interactive so it is not in `all` at all |
| `TestPrepLabelsIsInteractiveOnly` | The one command in the tree that cannot be scripted, and why: the env file names the label each broker role wants, but which MACHINE carries it comes from a prompt with no flag to express it. A non-interactive run is refused up front, touching nothing, rather than failing deep in the node picker on an unreadable stdin -- while an env file with no placement labels stays a no-op even without a terminal, since there is no question to ask |
| `TestDeployAllNeverLabelsNodes` | Labelling is out of the scripted path entirely, so `deploy all` never runs it -- not even when placement is configured and stdin happens to be a terminal, which would otherwise make the same command interactive or not depending on where it ran |
| `TestOpK8sRemoveAllAborts` | `opK8sRemoveAll`'s two error-return arms (DeleteBroker, DeleteSecrets): a failed broker- or secrets-deletion stops before `delete-namespace` ever runs, so a real failure cannot leave the namespace removed out from under still-live resources |
| `TestOpCtrVerifyRedundancyRunsCoordinated` | When this host is the primary, `opCtrVerifyRedundancy` actually calls `RedundancyCoordinated` instead of only ever hitting the skip/reject arms `TestCtrRoleGuards` already covers -- driven over a fake Runner seeded with an active-but-unhealthy `show redundancy` transcript so the health check fails immediately, before any poll loop or mate SEMP call |
| `TestContainerRoleDetectsFromHostname` | M1: `containerRole` (`deploy broker`/`deploy all`/`generate broker`'s role resolution) over the injected `App.Hostname` seam. An omitted role in HA is detected from the hostname against `nodes.*` and announced on stderr; an explicit role still wins over the hostname; an unrecognised hostname fails loud naming `primary\|backup\|monitor` as the way out (closing the old silent-second-primary bug, since `config.ParseRole("")` alone always returned Primary); standalone keeps defaulting to primary with no detection announced |
| `TestK8sSmokeRedundancyUnhealthy` | `smoke redundancy` fails on its first check rather than polling: over the echo seam, `engine.Echo`'s empty `show redundancy` output makes `primaryRedundancyUp` false, so `opK8sVerifyRedundancy` returns the redundancy-unhealthy error before any SEMP login |
| `TestK8sConfigDeleteDomainCertsConfigured` | With a CA actually configured, `config delete domain-certs` issues a kubectl exec instead of self-skipping -- every other test's `domainCerts.files` map is empty, so `domainCANames`'s map-to-slice conversion was correct only by vacuity until this one configures a CA |
| `TestConvertParseError` | a malformed legacy env file (unterminated array assignment) surfaces the parser's own error through the CLI |
| `TestConvertWriteError` | an -o path whose parent directory is absent fails with a wrapped write error naming the path |
| `TestK8sGenSecretsMissingCertFile` | kubernetes.tlsServerSecret configured with an unreadable tls.cert fails loud naming the read failure through `generate secrets broker`, instead of only being caught at real deploy time |

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
| `TestAliasesResolveToTheCanonicalCommand` | Pins equivalence by resolution rather than by running anything: cobra's own `Find` -- what dispatch itself uses -- resolves both an alias path (e.g. `rm br`) and its canonical form (`remove broker`) to the same `*cobra.Command`. Covers every verb/noun alias plus the standalone ones (`diag`, `cv`, `ver`, `sh`) |
| `TestAliasesDoNotCollide` | Walks the real tree and proves no two siblings answer to the same word (name or alias). `applyAliases` already panics on a collision at construction, so this is the second line of defence -- the one that would also catch a collision introduced by a command's own hand-written `Aliases` rather than by the table |
| `TestEveryAliasEntryIsLive` | Catches the quiet failure mode of a name-keyed table: every canonical name in `commandAbbrev` must be a command actually present in the tree, so a renamed command's alias cannot silently stop applying while the set still claims to provide it |
| `TestDangerousVerbsHaveNoBareAlias` | The safety property behind giving the most destructive verb a short form at all: `remove`, `deploy`, `config`, `start`, `stop` and `restart` are all verb GROUPS carrying subcommands, so `rm`/`dp`/`cfg`/`rs` reach something that acts on nothing until it is given a noun -- if one ever became a real command, this fails, which is the moment to reconsider the abbreviation rather than after a typo removes a broker |
| `TestGroupsRejectAnUnknownNoun` | Why a verb group is runnable at all. Cobra answers a NON-runnable command by printing help and exiting 0 whatever arguments it got, so `remove bogus` would report success having removed nothing and a script would never notice -- the same trap `completion` documents. Bare still prints help and succeeds; an unknown noun fails loudly |
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
| `TestGenPathNeverExecutes` | Backs the trust-model promise that rendering an untrusted env file is safe: `generate broker` issues no external command at all |

### platform_test.go

Platform resolution and the one-tree consequences of it. This is where the CLI's
central decision is pinned: the platform is inferred from the env file (or `--platform`)
rather than typed as the first word of a command, so what it resolves to, and what happens
when a command does not apply to it, are the behaviours most worth holding still.

Fixtures: `writePlatformEnv(t, platforms...)` writes a minimal but VALID env declaring
exactly the given sections (the kubernetes and container schemas need different
mandatory fields, so it emits both shapes); `runPlatform` runs `status broker` over
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
| `TestUnusableRoleFailsLoud` | A role that was TYPED must never be silently dropped, whichever way it is unusable. On a container platform, the six commands whose `[role]` moved to `--pod` (H2) hit the migration error (`noRolePositional`) -- `logs broker backup`, `cli backup`, `shell monitor`, `check semp-login backup`, `status broker backup`, `restart broker backup` -- before any platform-specific branch runs. On Kubernetes, the four commands that keep a real, container-only positional (`deploy broker`/`deploy all`, `config leader`, `smoke redundancy`) are refused by `rejectRole` instead, since it means nothing there. Both refusals must still name the role |
| `TestRolePositionalTeachesPodFlag` | The likelier half of the H2 migration: an operator on KUBERNETES typing the OLD documented spelling (`shell backup`, `logs broker monitor`, `cli primary`, `status broker backup`, `restart broker backup`, `check semp-login backup`) gets an error naming both `--pod` and the role typed, rather than cobra's bare "unknown command" hiding the fact that it merely moved. A `shell typo` subtest pins the other side: a word that is not a role keeps cobra's own wording and never offers the `--pod` hint, since a typo is not a migration |
| `TestPlatformIsAnnouncedInThePreamble` | The platform is inferred rather than typed by the operator, so it is stated in the preamble -- otherwise the one fact the operator does not type themselves would also be the one they cannot see |
| `TestCompletionNeverReadsTheEnvFile` | The invariant that decided where the pre-run hook lives. Cobra runs the NEAREST ancestor's `PersistentPreRunE` and `__complete` is root's own child, so a hook on root would parse an untrusted env file on every TAB press; keeping it per-command prevents that, proven by completing with an env file that does not exist |
| `TestPlatformFlagIsOnRoot` | `--platform` is a root persistent flag inherited by every command including `convert`, which is what lets one word mean one thing across the whole CLI |
| `TestScopedCommandsSaySoInHelp` | The tree is one static shape, so help text is the only place to learn a command does not apply before running it; a command that applies everywhere carries no scope tail |
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
| `TestRoleArgsComplete` | Every command that still takes a genuine `[role]` positional -- `deploy broker`/`deploy all`, `generate broker`, and the `roleOnContainerLeaf` pair `config leader`/`smoke redundancy` -- offers `primary`/`backup`/`monitor`. In every case the role is THIS host's own node identity, never a Kubernetes pod selector; the six commands where `[role]` moved to `--pod` (H2) are `TestPodFlagCompletesRoles`'s job now |
| `TestPodFlagCompletesRoles` | `--pod` completes to the same role set as the surviving positionals, not to filenames, table-driven over every command that takes it: `copy from`/`copy into` (always did), plus the six H2 moved it onto -- `cli`, `shell`, `logs broker`, `check semp-login`, `status broker`, `restart broker` |
| `TestPlatformFlagCompletes` | The root `--platform` flag, and `convert` inheriting the same one, offer exactly the three canonical platform names. The empty detect value is left out (omitting the flag is how you ask for it); so are `kube`/`dk`/`pm` -- the abbreviations save typing something you already know, which is what a completion does anyway, and offering both would put two names for one platform in front of the user -- and so are `k8s` and `k8` |
| `TestDirFlagCompletesDirectories` | `--dir` asks the shell to filter to directories |
| `TestNoArgsLeafOffersNoFiles` | A command built by `leaf` offers nothing, stopping cobra's filename fallback on the majority of commands in the tree. `deploy broker`/`deploy all` are not in the list: both take a `[role]`, so they complete to role names rather than nothing, which `TestRoleArgsComplete` already covers |
| `TestAllowCommandOffersNoFiles` | `--allow-command` offers no files: the value is a bare binary name, and paths are what its own help text warns against |
| `TestExampleArgsComplete` | `examples` completes to the template names it knows, with no file completion, and offers only the canonical spellings -- it accepts `kube`/`dk`/`pm` too, but a completion is already what saves the typing, so it teaches the full word |
| `TestFlagCompletionsRegistered` | The drift gate: every flag that should have a completion function still has one, since a renamed flag silently reverts to filename completion at a TAB press and no other test would notice |

---

## internal/convert

The legacy bash env -> YAML converter: a shell-assignment parser, the variable
mapping, and the YAML emitter. 37 tests.

### convert_test.go

| Test | What it covers |
| --- | --- |
| `TestConvertLegacyK8sEnv` | `testdata/legacy-k8s.env` converts end to end and matches `testdata/legacy-k8s.yaml.golden`: platform detected as kubernetes and written as a `kubernetes:` section, `true` -> `yes`, every scalar/array/associative value mapped, `${SOLBK_NS}` expanded, a trailing comment stripped, an explicit `0` kept, a single-quoted `REPL_PSK` containing `$` carried through verbatim (the B5 regression, end to end), a multi-word `KUBE` preserved as `kubernetes.runtime` argv, and only the two expected advisories. The fixture sets `SOLBK_MSGNODE_CPU`, as every real legacy file does, so the drop is exercised here: it warns, and no `cpu` reaches the YAML |
| `TestParseSingleQuotedDollarSurvives` | The B5 regression. Single quotes are bash's idiom for keeping `$` literal, and the tokenizer used to strip quoting before `expand()` ran -- so `SOLBK_REDUNDANCY_PSK='p$s3cret'` silently converted to `psk: "p"`, truncating the secret with no warning and surfacing much later as an unexplained redundancy failure. The name in the fixture is unassigned, which is exactly the case that vanished into `""` |
| `TestParseMixedQuotingExpandsOnlyTheExpandableHalf` | Why literalness is tracked per SEGMENT rather than per word: in `a'$b'"$B"` the single-quoted half stays literal even though `B` is a real assigned variable, while the double-quoted half still expands. A per-word flag could not express this |
| `TestParseArrayElementsPreserveQuoting` | The array body is a SECOND, separate call site from the scalar path (`tokenizeSegments` + per-word expand in `parse()`), so it gets its own coverage rather than being assumed to follow |
| `TestParseDoubleQuotedBackslashEscapes` | The other route to the same secret loss, which the single-quote fix alone did not close. Inside double quotes bash honours a backslash before exactly `` $ ` " \ `` and leaves it literal before anything else; stripping it unconditionally turned `"\$SECRET"` into an expandable `$SECRET` (truncating it even when the name IS assigned -- one subtest pins that) and quietly ate the separators in `"C:\Users\me"`. An escaped character becomes its own literal segment, so it can never reach the substitution |
| `TestConvertUserPasswordsBecomeAdditionalUsers` | The one legacy variable with no like-for-like successor: `SOLBK_USR_PASS` becomes structured `admin.additionalUsers` entries with the least-privileged `accessLevel: none` plus a warning naming that choice, malformed entries are dropped with a warning naming their POSITION and never their text (a malformed entry is most likely a bare password), and `Convert` re-validating its own output proves the emitted level is a legal one |
| `TestConvertAdminUserIsContainerOnly` | The one admin field that is not portable: `SOLBK_ADM_USER` is emitted only for docker/podman, and on a k8s target is dropped with a warning naming why (`validateK8s` refuses any non-`admin` value), stays out of the generic unmapped list because it is still read, and leaves a document that validates -- no "will not load as-is". A source that already said `admin` warns about nothing |
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
| `TestConvertRedundancySpellings` | `true`/`yes` and `false`/`no` normalise (any case); anything else copies through with a warning |
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
state machines, the primary-driven container HA variants, and the SEMP mate channel.
166 tests across 5 files.

### broker_test.go

| Test | What it covers |
| --- | --- |
| `TestField` | `field` extracts a labelled value from CLI output, empty for an absent label |
| `TestCountContains` | `countContains` counts labelled lines carrying a substring |
| `TestContainsAnyFold` | Case-insensitive substring match used for error detection |
| `TestCountAnyFold` | The per-line counting form of `containsAnyFold` -- counts CLI output lines containing `invalid`/`error`/`busy` case-insensitively, which is how ExecCLI sizes its failure without quoting the matching lines |
| `TestValidName` | The boundary validator accepts safe names and rejects empty, `..`, separators, and shell metacharacters |
| `TestPathHelpers` | `cliScriptPath`, `cliArg`, and `certPath` build the in-jail paths |
| `TestLastLines` | `lastLines` returns the tail, and the whole input when it is shorter than n |
| `TestHTTPStatusHelpers` | `isHTTP2xx` accepts only 2xx; `httpStatusLines` extracts every status line |
| `TestPrimaryRedundancyUp` | The primary health predicate requires redundancy Up |
| `TestRunCLIUploadsThenExecs` | `RunCLI` uploads the script body, then execs it, and returns its output |
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
| `TestAdditionalUsers` | The k8s-only user-creation op: the generated script is uploaded verbatim, the password never reaches an argv or the shown output (the CLI transcript repeats it, so it is withheld), and the uploaded script is deleted afterwards because its body carries every password |
| `TestAdditionalUsersReportsExistingUser` | The deliberate non-idempotency: a broker-reported `already exists` fails loud naming that as the likeliest cause, the error carries no transcript, and the script is still removed on the failure path |
| `TestAdditionalUsersEmpty` | No users configured is an error |
| `TestAdditionalUsersRejectsBadValues` | Every injection and CLI-quoting path is refused before anything is uploaded -- a username with a space, a multiline access level, an empty password, and a password containing a newline or any of the characters the broker rejects inside a quoted value -- and no error echoes the password. Punctuation the CLI does accept still works |
| `TestRemoveDomainCerts` | The removal half of the domain-CA pair emits a script naming the CA, reachable from both platforms |
| `TestRemoveDomainCertsRejectsBadName` | A CA name with a space is rejected before any upload |
| `TestRemoveDomainCertsEmptySkips` | No configured CAs makes no calls |
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
| `TestRunCLIUploadError` | RunCLI wraps and returns an Upload failure with the script name and never invokes the CLI binary |
| `TestServerCertUploadError` | ServerCert stops before apply-server-certs when the certificate bundle upload fails |
| `TestServerCertRunCLIError` | ServerCert returns the apply-server-certs error and never shows output |
| `TestDomainCertsUploadFileError` | DomainCerts stops before load-domain-certs when a CA file upload fails |
| `TestDomainCertsRunCLIError` | DomainCerts returns the load-domain-certs error and never shows output |
| `TestDisableDefaultVPNDisableError` | DisableDefaultVPN stops before reading back show-vpn when disabling the VPN fails |
| `TestDisableDefaultVPNShowError` | DisableDefaultVPN surfaces a failed show-vpn readback and skips cli-script cleanup |
| `TestDisableDefaultUsersShowVPNError` | DisableDefaultUsers stops before disabling anything when the VPN listing fails |
| `TestDisableDefaultUsersDisableError` | DisableDefaultUsers surfaces a failed disable-default-usernames call rather than reporting success |
| `TestProductKeysRunCLIErrorStopsLoop` | ProductKeys stops its per-role loop on a transport failure and never touches the remaining role |
| `TestAdditionalUsersRunCLITransportError` | a hard RunCLI transport failure in AdditionalUsers still removes the uploaded password-bearing script |
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
| `TestAdditionalUsersScript` | Exact script text for the k8s user-creation run: both values quoted so a password with a space survives, one `create username ... password ...` plus `global-access-level` per user, and no trailing `show` whose output would be discarded anyway |
| `TestProductKeysScript` | Exact script text for a list of keys |
| `TestDisableDefaultVPNScript` | The hardening script disables the VPN, its default user, plain-text downgrade, and plain-text SMF |
| `TestParseVPNNames` | Column-based VPN parsing skips the legend, header, separator, comments, and blank lines |
| `TestParseVPNNamesNoSeparator` | Output without a separator row yields no names |
| `TestGatherConfigsScript` | Prefix, first show command, `gather-diagnostics` with days substituted, and one line per configured show |
| `TestZipConfigsScript` | The zip command is present |
| `TestSortedKeys` | `sortedKeys` returns map keys in sorted order |

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
| `TestMateSEMPPreflightMissingIP` | An empty nodes.backup.ip refuses before any transport call, naming the field |
| `TestMateSEMPPreflightNon2xx` | A non-2xx reply fails naming the status and the actionable causes |
| `TestMateSEMPPreflightTransportError` | A transport error is wrapped with the host-to-host reachability hint |
| `TestMateRevertActivitySuccess` | With a server certificate configured and no `tls.cas`: exact argv (`curl -is -K - https://<mate>:1943/SEMP`) with the user, data and insecure config lines on stdin, a warning naming `tls.cas`, and one call -- callers own the preflight |
| `TestMateRevertActivityCredsAndBodyNeverInArgv` | The S3 boundary: no argv token carries the password or the RPC body |
| `TestMateRevertActivityNon2xx` | A non-2xx reply fails on the status even when the body says ok |
| `TestMateRevertActivityRPCNotOK` | A 2xx reply without code="ok" fails, dumping the header-stripped reply |
| `TestMateRevertActivityBridgePlaintextOnly` | The last-resort branch: a bridge mapping only 8080 (1943 unmapped, even with a certificate configured) keeps today's plain http:// on the mapped port, but now warns first that the credentials are crossing the wire unencrypted |
| `TestMateRevertActivityBridgeTLSNoCA` | A bridge mapping exposing 1943 is preferred over a mapped 8080; with no `tls.cas` configured, verification is disabled (self-signed broker cert is the normal case) and a warning names `tls.cas` |
| `TestMateRevertActivityBridgeTLSWithCA` | The fully-verified branch: a mapped 1943 plus `tls.cas` configured verifies the mate's certificate via `--cacert` instead of disabling verification, and no warning is needed |

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
| `TestRedundancyCoordinatedMissingMateIP` | A missing nodes.backup.ip refuses, named in the error, before anything is released |
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
| `TestRoleName` | Role -> display name |
| `TestLocalRoleDefaultHostname` | With the seam unset, the real `os.Hostname` is used and an off-table host fails loud |
| `TestLocalRoleHostnameError` | LocalRole wraps and returns a Hostname read failure instead of matching a garbage host against the node table |
| `TestLeaderLocalBadRoleArg` | LeaderLocal propagates an invalid explicit role arg before the primary-only guard runs, making no transport calls |
| `TestLeaderLocalPollCondError` | LeaderLocal aborts on a mid-poll transport error while still dumping show-redundancy-detail, closing an asymmetry with the k8s Leader |
| `TestLeaderLocalAssertLeaderError` | LeaderLocal returns the assert-leader error after a healthy poll without showing output |
| `TestMateActivityStateReadsTheMateColumn` | `MateActivityState` reads the `Activity Status` line for exactly "Mate Active"; local-active, standby, empty output, and the phrase appearing under an unrelated label all read false |
| `TestShowRedundancyIsReadOnly` | The exported `ShowRedundancy` wrapper issues exactly one `show redundancy` CLI script on the named role and nothing else, since a probe calls it to decide whether a mutation is safe |

---

## internal/k8s

Everything driven through `kubectl`: the read-only permission preflight, prep, deploy,
operator, day-2 ops, secrets, and the pod transport. 162 tests across 15 files.

### names_test.go

| Test | What it covers |
| --- | --- |
| `TestResourceNames` | Pod, PVC, StatefulSet, and load-balancer service names for every role |
| `TestRestartOrder` | The safe manual-bounce order (monitor, backup, primary; standalone just the primary) |
| `TestHARoles` | HA yields all three roles; standalone yields only the primary |
| `TestProductKeyRoles` | Product keys target primary+backup in HA, primary only in standalone |

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

The grouped `check deploy` report (`checkreport.go`) that replaced the old flat
`CheckEnv`/`CheckStorageClass` passthrough: one `checkReport` with a fixed section order,
built and rendered even when an earlier section failed, so one run names every problem
instead of the first.

| Test | What it covers |
| --- | --- |
| `TestCheckDeployNeverPrintsASecret` | Successor to `TestCheckEnvNoSecretLeak`: the admin, monitor and registry passwords never reach the report, which shows only `admin=set`/`monitor=set` |
| `TestCheckDeployGroupsAndOrdersSections` | Pins the layout: sections render in the agreed order (Deployment, Operator, Broker, Credentials, Placement), the four leading rows (namespace, name, image, image pull) lead in order, and cpu/mem are never crammed into one compound row |
| `TestCheckDeployTagsEchoedConfigAsINFO` | Successor to `TestCheckDryRun`: config read back from the env file is tagged `[INFO]`, never `[ OK ]` (a report that tagged unverified config as OK would claim verification nobody did), and a preview reports every cluster-dependent check as `[SKIP]` |
| `TestCheckDeploySparseConfigExplainsItself` | Successor to `TestCheckEnvSparseConfig`: an empty watch list explains itself as "watches ALL namespaces" rather than the reassuring opposite, and unset TLS/admin password read `(not configured)`/`MISSING` |
| `TestCheckDeployReportsEveryFailureInOneRun` | The behaviour change worth having: a refused permission is the first failure, but the run continues and still renders every section after it, and the returned error counts the failures rather than wrapping only the first |
| `TestCheckReportFailedCounts` | `checkReport.failed()` sums the FAIL rows across every section; an empty report counts zero |
| `TestStorageRows` | Successor to `TestCheckStorageClass`, exercising the `storageRows` section builder directly: a suitable configured class is OK with no default lookup; Immediate binding or no expansion is FAIL; missing attributes report `<none>` and FAIL; the actionable message names the fix without the old bash script number; and every read failure along the way (default resolution, the first attribute column, the second after the first succeeds) surfaces its own FAIL row rather than being swallowed |
| `TestCheckDeployReportsResolvedPorts` | M11: an explicit `kubernetes.ports` resolves into the new leading Config section as a `ports (N)` count plus the `name=port` pairs, readable straight off the report with no need to deploy or run `generate broker` |
| `TestCheckDeployReportsDefaultPorts` | M11: leaving `kubernetes.ports` unset is not "no ports configured" -- `ApplyDefaults`'s 17-entry default is what actually renders into the broker CR, so the Config section shows `ports (17)` and both the first and last default port; a precondition assertion fails loud if defaulting itself ever stopped happening, rather than passing vacuously on an empty list |
| `TestPortRowsNeverFail` | M11: `portRows` produces only `[ OK ]` rows at every wrap boundary (0, 1, a full chunk, one over, and the 17-port default) -- `check deploy` is read-only and must never stop a deploy over how many ports there are to print |
| `TestCheckDeployConfigSectionNeverFails` | M11: the report-level companion to `TestPortRowsNeverFail` -- `validationRows` alone must never move `checkReport.failed()` off zero |
| `TestCheckReportSkipsEmptySections` | A section with no rows (e.g. Placement on a standalone config) prints no header at all, rather than an empty block |

### prep_test.go

| Test | What it covers |
| --- | --- |
| `TestCreateNamespace` | The namespace manifest is applied on stdin |
| `TestDeleteNamespace` | B6: the empty-namespace path through the enumeration guard -- a bare `recRunner` answers both enumeration reads (built-ins, then the broker CR) with zero items, so nothing is FOREIGN and the delete proceeds to `kubectl delete namespace ... --ignore-not-found` exactly as before the guard existed |
| `TestDeleteNamespaceProtected` | B6: the four Kubernetes system namespaces (`default`, `kube-system`, `kube-public`, `kube-node-lease`) are refused outright and unconditionally -- no override, no `--no-prompt` bypass -- and, unlike every other guard in DeleteNamespace, the refusal happens before the cluster is asked anything at all (zero calls made) |
| `TestDeleteNamespaceOnlyOurs` | B6: every OURS classification rule at once -- the `<kubernetes.name>-pubsubplus` name-suffix match (pod, PVC), the configured admin secret, and the two Kubernetes-generated exceptions (`kube-root-ca.crt` ConfigMap, default ServiceAccount) -- none of it is FOREIGN, so the delete proceeds |
| `TestDeleteNamespaceForeignRefuses` | B6: a Deployment this tool did not create means the delete never runs; the foreign object is named in a `Table` report on `Out`, and the error states the count and points at `remove broker`/`remove secrets` |
| `TestDeleteNamespaceForeignCannotBeSilenced` | B6: DeleteNamespace takes no "skip the check" parameter, so calling it exactly as the CLI's `--no-prompt` path does -- straight through, no question asked first -- still refuses; there is nothing upstream left to silence |
| `TestDeleteNamespaceForeignBrokerRefuses` | B6: the case the guard was built for -- a SECOND Solace broker sharing the namespace. Its objects carry the operator's `-pubsubplus` suffix exactly as ours do, so `isOurs` matches the full `<kubernetes.name>-pubsubplus` string rather than the bare suffix, keeping the other broker's pod and LB service foreign while this deployment's own pod is not listed |
| `TestDeleteNamespaceBrokerListingFailsStillProceeds` | B6: the deliberate asymmetry in `namespaceObjects` -- the broker-CR listing failing only warns, and the teardown proceeds by what it could see, because an absent CRD is a state this tool's own documented removal order (brokers, then `remove operator --delete-crd`, then `remove namespace`) produces, and refusing there would block a legitimate teardown forever |
| `TestDeleteNamespaceEnumerationFails` | B6: the built-in listing failing (RBAC denial, unreachable API server) refuses the delete too and wraps the cause -- being unable to see what the namespace holds is not permission to proceed |
| `TestDeleteNamespaceEmpty` | B6: a namespace with nothing in it at all -- what `remove all` leaves behind once the broker and secrets are already gone -- proceeds to the same delete as before the guard |
| `TestDeleteNamespaceStopsOnPreflightFailure` | A refused `auth can-i delete namespaces` stops DeleteNamespace before any `kubectl delete` is issued -- the same shape as `TestCreateSecretsStopsOnPreflightFailure`, but for the namespace teardown |
| `TestCreateSecretsAdminOnly` | With no TLS or pull secret only the admin secret is applied, as a single document |
| `TestCreateSecretsAllThree` | Admin + TLS + pull secret join into one multi-doc apply, and the registry password reaches neither argv nor plaintext stdin |
| `TestCreateSecretsPreflight` | Missing TLS inputs fail before any apply runs |
| `TestDeleteSecrets` | All configured secrets are deleted; admin-only config deletes one |
| `TestUpdateServerCertSecret` | The TLS secret is applied on stdin; an unset secret name errors |
| `TestSplitLabel` | Label parsing across `=` and `:` forms, with whitespace, and rejection of malformed entries |
| `TestIsBuiltinLabel` | Kubernetes-owned label keys are recognised and custom ones are not |
| `TestLabelNodesNoCustomLabels` | No labels means no cluster calls and an early-exit message |
| `TestLabelNodesBuiltinOnly` | Built-in labels are never applied |
| `TestLabelNodesMalformedAndUnsafe` | Malformed and unsafe-character labels are dropped with warnings and never reach the cluster |
| `TestLabelNodesHappyPath` | RBAC precheck (now the shared `Preflight`, which tells "not allowed" from "nobody answered"), node list, then the label call for the selected node |
| `TestLabelNodesReprompt` | Out-of-range and non-numeric selections re-prompt before the correct node is labelled |
| `TestLabelNodesRBACDenied` | A failed RBAC precheck aborts and labels nothing |
| `TestLabelNodesEOFNoSelection` | EOF with no selection errors |
| `TestCreateSecretsFailsWithoutAdminFields` | CreateSecrets can pass secretPreflight (TLS-only) and still fail loud inside GenSecrets when admin.pass/kubernetes.adminSecret are unset, with zero applies made |
| `TestCreateSecretsStopsOnPreflightFailure` | A refused `auth can-i create secrets` stops CreateSecrets before GenSecrets reads the TLS private key off disk -- loading key material for a cluster that will not accept it is work worth not doing |
| `TestGenSecretsTLSError` | GenSecrets itself (not just via CreateSecrets' preflight) fails when kubernetes.tlsServerSecret is set but the cert files are unreadable, guarding `generate secrets broker` |
| `TestDeleteSecretsStopsOnPreflightFailure` | A refused `auth can-i delete secrets` stops DeleteSecrets before any `kubectl delete secret` is issued -- the same shape as `TestCreateSecretsStopsOnPreflightFailure`, but for the secret teardown |
| `TestDeleteSecretsSkipsUnconfiguredAdminSecret` | DeleteSecrets never issues `kubectl delete secret ""` when kubernetes.adminSecret was never configured |
| `TestDeleteSecretsStopsOnError` | A genuine delete failure stops the teardown loop and surfaces instead of silently continuing to the remaining secrets |
| `TestLabelNodesHAOnlyPrimaryConfigured` | In an HA config with only LabelsPrimary set, backup and monitor are read from their own config fields (not Primary's) and are silently skipped when empty, never prompted |
| `TestNodeNamesError` | nodeNames fails loud with its own wrap on a genuine query failure instead of returning a misleadingly empty list |
| `TestLabelNodesNoNodesFound` | LabelNodes fails loud when RBAC passes but the cluster reports zero nodes, instead of misbehaving in promptNode with an empty list |
| `TestLabelNodesLabelFailureIsNonFatal` | A single failed label application is reported and skipped, not fatal, matching the doc comment's stated contract |
| `TestCreateNamespaceApplyFails` | A failing apply (RBAC denial) surfaces from CreateNamespace instead of being silently swallowed |
| `TestPromptsGoToErrNotOut` | The stream split the node picker used to get wrong: its banner, list and `> ` go to `Err` (stderr) like every other prompt in the tool, while the per-label `[ OK ]` outcome stays on `Out`, so `prepare labels > file` captures results and not questions |

### preflight_test.go

| Test | What it covers |
| --- | --- |
| `TestCanIAnswerReadsTheLastLine` | The verdict is the LAST non-empty line, not the whole output: `kubectl auth can-i` prints advisory lines above it on stdout ("Warning: resource 'x' is not namespace scoped"), and comparing the whole body would turn every such cluster into the unreadable-answer branch -- a preflight failing safe in the wrong direction. Covers plain yes/no, no trailing newline, one and several warnings, blank lines, CRLF, empty, and whitespace-only |
| `TestPreflightAcceptsAWarnedYes` | The end-to-end of the above: a cluster that warns and then permits lets the deploy proceed |
| `TestPreflightRefusesAnUnreadableAnswer` | Exit 0 with neither yes nor no -- a wrapper that swallowed stdout -- is refused rather than assumed permitted, since proceeding would act on a permission nobody confirmed |
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
| `TestPreflightUnreachableClusterHints` | An unreachable API server is a different failure from an RBAC refusal and gets the hint that helps (`log in first`, `oc login`), carrying kubectl's own error rather than replacing it; still nothing runs after the probe |

### operator_test.go

| Test | What it covers |
| --- | --- |
| `TestWatchNamespace` | `WATCH_NAMESPACE` joins: broker namespace appended by default, onto a configured list, or omitted when disabled -- plus the dedupe half, since a list that already named the broker namespace (the common case, `watchBrokerNs` defaults on) listed it twice in the report and the applied Deployment. Entries are trimmed, empties and trailing commas dropped, repeats inside the list collapsed, first occurrence winning. controller-runtime's map-keyed cache hid the repeat at runtime, so only these cases can catch a regression |
| `TestOperatorImage` | The registry-prefix rule now shared by `RenderOperator` and `CheckEnv`: prefixed when `image.registry` is set, raw when it is not. Its own test rather than only being reached through the 119 KB bundle render, because the report and the apply drifted for exactly as long as each owned a copy |
| `TestRenderOperatorSubstitutions` | Every substitution point lands (namespace, watch list, image with/without registry prefix, resources, pull-secret reference) and no template marker survives -- and the bundle carries no `.dockerconfigjson`, since the reference is a name and the Secret is its own artifact |
| `TestGenOperator` | Render-only uses the configured operator namespace, or the fixed default when unset |
| `TestGenOperatorSecrets` | `GenOperatorSecrets` renders the regcred alone in the configured operator namespace, falls back to the default namespace, and refuses loud (naming kubernetes.imagePullSecret) when there is no pull secret to render -- printing an empty artifact would read as "no credentials needed" |
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

The rendered `status broker`/`status cluster`/`status operator` reports (`statusreport.go`)
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
| `TestOperatorReportWhenNotInstalled` | A missing operator is a WARN naming `deploy operator`, not an error |
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
| `TestAdminSecretExcludesAdditionalUsers` | The finding that shaped the k8s user path: the operator reads only the admin and monitor keys, so an additional user's name and password (plain and base64) must be absent from this Secret entirely |
| `TestAdminSecretErrors` | Empty password, empty `kubernetes.adminSecret`, and an additional user with no name, a bad name, or no password all error |
| `TestTLSSecretErrors` | Unset cert, unset secret name, and missing cert/CA/key files all error |
| `TestDockerRegistrySecretEmptyName` | An empty pull-secret name errors |

### transport_test.go

| Test | What it covers |
| --- | --- |
| `TestTransportExecArgs` | Exec argv for Run/Output/OutputInput, `-i` only where stdin is used, and `-c` never present (broker pods are single-container) |
| `TestTransportUpload` | The body rides stdin through `sh -c 'cat > <dest>'` and never appears in the argv |
| `TestTransportUploadQuotesDest` | Single-quote escaping stops a metacharacter in a path breaking out of the redirect |
| `TestTransportCopy` | `kubectl cp` argv in both directions with the namespace flag |
| `TestTransportEchoHidesUploadBody` | End to end over Echo: the uploaded body shows as a byte count, and the CLI exec line is still echoed |

---

## internal/container

The host-local Docker/Podman manager, its node-local transport, and the engine
preflight that precedes every mutating operation. 110 tests across 4 files.

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
| `TestManagerPrepHostDryRunDoesNotWritePSK` | Dry-run leaves the env file untouched, never generates a PSK, and still echoes mkdir/chown |
| `TestManagerPrepHostWritesPSK` | The generated PSK is written into `nodes.psk`, the replication PSK is untouched, and the data dir is created and chowned |
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
| `TestManagerDeployPodmanCreatesSecrets` | Deploy loads both secrets into podman's store with `secret create --replace` under container-scoped names (`sol-pod-*`), values on stdin and never in an argv; the quadlet unit is 0600 |
| `TestManagerDeployRejectsEmptySecret` | An empty required secret fails the deploy naming the field and the fix (`prepare host` for the PSK) rather than starting a broker without one |
| `TestManagerDeployDockerDryRunMasksSecretEnv` | Dry-run creates nothing, stays previewable before `prepare host` has generated the PSK, and echoes the compose environment as `NAME=***` without the password |
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
| `TestManagerDeletePodmanRemovesSecrets` | Delete now removes every secret `CreatePodmanSecrets` loaded into podman's own store (the same `render.ContainerSecrets` list), so they no longer outlive a `remove all --delete-data`; a failing removal warns rather than failing a teardown that otherwise succeeded |
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
| `TestManagerStatusDockerCompose` | Compose-mode status runs `compose ps` and a filtered `docker ps` |
| `TestManagerLogsCLIShell` | `logs` follows, `cli` and `shell` exec interactively into the container |
| `TestReplacePSKLine` | Only the `nodes.psk` line is replaced, never the replication one; absence of the line is reported |
| `TestDefaultGenPSK` | The default generator produces 60 base64-encoded random bytes |
| `TestManagerCheckReachableError` | A failing runtime version probe fails the check |
| `TestManagerPrepHostMkdirError` | A mkdir failure propagates |
| `TestManagerPrepHostChownError` | A chown failure propagates |
| `TestManagerPrepHostRootlessUnshareChownError` | A rootless `unshare chown` failure propagates |
| `TestManagerPrepHostGenPSKError` | A PSK generation failure propagates |
| `TestManagerPrepHostWritePSKReadError` | An unreadable env file fails the PSK write-back |
| `TestManagerPrepHostWritePSKWriteError` | A read-only env file fails the PSK write-back (skipped when running as root) |
| `TestManagerPrepHostPSKAlreadySet` | An existing `nodes.psk` skips generation and says so |
| `TestManagerPrepHostNoPSKLinePrintsValue` | With no psk line to replace, the value is printed for the user and the file is not modified |
| `TestManagerDeployPodmanMkdirError` | An uncreatable quadlet dir fails deploy |
| `TestManagerDeployPodmanWriteUnitError` | An unwritable unit path fails deploy |
| `TestManagerDeployPodmanDaemonReloadError` | A daemon-reload failure propagates |
| `TestManagerDeployPodmanStartError` | A service-start failure propagates |
| `TestManagerDeployPodmanEUIDGuardFails` | Rootless-as-root is rejected by the euid guard |
| `TestManagerDeployDockerComposeWriteError` | An unwritable compose path fails deploy |
| `TestManagerDeployDockerComposeUpError` | A `compose up` failure propagates |
| `TestManagerDeployPodmanSecretCreateError` | A failed `secret create` aborts the deploy, naming the config key behind the secret |
| `TestManagerDeletePodmanDaemonReloadError` | A daemon-reload failure during delete propagates |
| `TestManagerDeletePodmanRemoveUnitError` | An unremovable unit path fails delete |
| `TestManagerDeleteDockerComposeDownError` | A `compose down` failure propagates |
| `TestManagerDeleteDockerStopTolerated` | The tolerate branch B2 kept: a failed stop that the engine then CONFIRMS is not running still warns and goes on to `rm`. Uses `capRunner.outFor` so the two probes answer differently -- `ps --all` lists the container (so stop is attempted at all), `ps --filter status=running` does not (so it is confirmed down). Confirmation is what earns the tolerance; silence does not |
| `TestManagerDeletePurgeError` | A failing data-dir removal under `--purge` propagates |
| `TestManagerStatusDockerNoComposeFile` | With no compose file on disk, status lists the container and never calls compose |
| `TestManagerStatusPodmanUnitInactiveTolerated` | An inactive unit warns but status still lists the container |
| `TestManagerStatusDockerComposePsTolerated` | A failing `compose ps` warns but the plain `ps` still runs |
| `TestManagerCheckPodmanEUID` | The euid guard across rootless/rootful x root/non-root, and skipped on a non-POSIX euid |
| `TestManagerPrepHostRootlessAsRootFailsHard` | PrepHost now shares Deploy's `checkPodmanEUID` and fails hard (one definition, one message) when `podman.rootless=true` but the process is root, instead of only warning and going on to mkdir/chown under the wrong namespace mapping -- neither call is reached once the guard rejects |
| `TestManagerNilSinks` | Nil log and output sinks fall back to discard and stdout without erroring |

### transport_test.go

| Test | What it covers |
| --- | --- |
| `TestTransportExecArgs` | Exec argv is `<runtime> exec [-i] <name> ...` with no `--` (docker rejects it), and the role argument is ignored because the transport is node-local |
| `TestTransportUpload` | The body rides stdin through `sh -c 'cat > <dest>'` and never appears in the argv |
| `TestTransportUploadQuotesDest` | Single-quote escaping stops a metacharacter in a path breaking out of the redirect |
| `TestTransportCopy` | `<runtime> cp` argv in both directions, container-name prefixed |
| `TestTransportEchoHidesUploadBody` | End to end over Echo: the body shows as a byte count and the CLI exec line carries no `--` |
| `TestTransportEchoHidesSEMPConfig` | End to end over Echo via LeaderLocal: the mate SEMP curl is echoed with its stdin as a byte count, never the admin password |

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

Manifest and unit-file rendering, guarded by committed goldens. 17 tests.

### render_test.go

| Test | What it covers |
| --- | --- |
| `TestGolden` | Fourteen renderings from the sample env match their goldens: k8s broker CR (the sample omits `kubernetes.ports`, `timezone` and both security blocks, so this covers the default ports and the omitted branches), the same CR with an explicit port list (a container port differing from the service port, and an explicit protocol), the same CR with timezone and both security blocks set, podman quadlet, docker compose in HA and standalone (standalone drops the redundancy block and its PSK secret reference), container env pairs for HA (no `timezone`, so no TZ pair) and standalone (`timezone` set, so the TZ pair is present), the container env file, the podman and docker secret scripts (the docker case uses a password holding a quote, a space and a `$`), the quadlet and compose forms of the opt-in health check, the CR with an explicit pullPolicy plus podAnnotations/podLabels, the CR with node and pod affinity alongside the legacy anti-affinity term, and the CR with loadBalancer annotations, node labels and tolerations (values carrying a colon and a URL, which survive only because both halves are quoted) |
| `TestArtifactsCarryNoSecrets` | The externalization guard: with distinctive values in `admin.pass`, `nodes.psk` and an additional user's password, no deployment artifact on any platform (broker CR, quadlet, compose file, env file) contains any of them, while `SecretScript` -- the renderer that supplies them -- contains all three |
| `TestContainerSecretsRedundancy` | HA lists both secrets in a fixed order with the expected broker settings, `FilePathKey`/`MountPath` derive the file form both engines use (the mount is named after the setting, not the host-side secret), and standalone lists the admin password only (no mate link, so no PSK secret). An encrypted server-certificate key adds a third secret reaching the broker as `tls_servercertificate_passphrasefilepath`, and only when the passphrase is actually set |
| `TestContainerSecretNamesAreHostScoped` | The de-confliction: the host-side name is `<container.name>-<suffix>` (the default name keeps the historical `solace-admin-password`), the in-container target and path never carry that prefix, and `EnvVar` maps `.`/`-` to `_` and prefixes a leading digit so the name stays exportable |
| `TestAdditionalUsersReachBothHalves` | An extra user's password becomes a per-host secret named after it (`ConfigKey` naming the env-file key), while its access level and `*filepath` pointer ride the env pairs and the password does not |
| `TestQuadletHealthCmdEscapesPercent` | systemd expands %-specifiers in every unit assignment, not just the quoted `Environment=` ones, so a percent-encoded character in a probe URL is doubled or the line is dropped and the health check silently disabled. Quotes and backslashes stay untouched there (the value is unquoted and podman splits it itself), and compose keeps the percent literal since it has no specifier expansion |
| `TestHealthCmdDefaultsToReadiness` | An enabled block with no cmd polls `/health-check/readiness` on 5550, and an explicit cmd wins |
| `TestSecretPreflight` | The precondition `deploy` and `generate secrets broker` share: an empty secret value is refused up front, naming the field and `prepare host` for the PSK. Standalone needs no PSK, so an empty one is fine there |
| `TestShQuote` | Secret-script quoting: a value holding a single quote survives as itself instead of ending the shell string |
| `TestParsePort` | Port entries across the `name=container`, `container:service`, and `/PROTO` forms |
| `TestParseToleration` | Toleration Equal (`key=value:effect`) and Exists (`key:effect`) forms |
| `TestQuadletEscape` | systemd `Environment=` escaping of `%`, `"`, and `\` |
| `TestScalingReachesContainersAsEnv` | Every scaling knob reaches docker and podman as a container environment variable, carrying the env file's values, including an explicit `0`, which is a real setting rather than an absent one |
| `TestScalingReachesK8sAsSpecOnly` | The other half of the delivery split: on k8s the same settings are CR fields under `spec.systemScaling` and never pod environment variables, the spool size is spelled `maxSpoolUsage` there, and the container spelling appears nowhere in the CR |
| `TestScalingTierReachesEveryArtifact` | One tier value decides the CPU cap in all three artifacts: the broker CR's `messagingNodeCpu`/`messagingNodeMemory`, compose's `cpus:`/`mem_limit:`, and the quadlet's `PodmanArgs=--cpus=`/`Memory=`. It uses 100000, which is no platform's default, so the value is proven read rather than hardcoded -- the goldens only ever show the default tier |
| `TestContainerMemOverrideReachesArtifact` | The asymmetry survives to the artifact: an overridden `container.mem` reaches compose while the CPU stays the tier's |
| `TestUnresolvedTierOmitsLimits` | The renderers' fail-safe branch. A `Config` built in code -- what the executors are handed -- carries no tier, and all three artifacts must then omit the limits rather than emit an empty `cpus:`/`--cpus=`/`messagingNodeCpu:`, which the engines and the CRD would reject |

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

## internal/tools/itest

The dev-only live-environment probe harness ([itest.md](itest.md)), three probes. Its own tests run with no
cluster, engine or broker, like the rest of the suite -- what they cover is precisely what a
live run cannot: the ORDER of each probe's steps, that cleanup still fires when a step in the
middle fails, and every verdict branch as a pure function. A probe that skipped its restore on
the error path would leave a throwaway password on a real broker, which is not something to
discover by running it against one. 40 tests.

Fixtures here are hand-authored from documentation, not captured from a live system -- that is
the whole point of the harness. `sempOKReply` in particular is flagged in code to be replaced
with a real capture after the first live run.

### itest_test.go

Registry, selection, consent and the pure classifiers.

| Test | What it covers |
| --- | --- |
| `TestRegistryCoversEveryClaim` | Every registered probe names the source comment it settles, declares platforms and a Touches line, and has a unique id -- a probe nobody can trace back to a question, or one that would SKIP everywhere, is a bug in the registry |
| `TestSelectProbes` | Selection: no selection runs everything, positional and `--probe` comma forms agree, case/spacing tolerated, asking twice runs once, and an unknown id errors while listing the known ones |
| `TestListProbesNeedsNoEnvFile` | `--list` renders the whole catalogue (including what each probe touches) on stdout with nothing on stderr, and without resolving an env file -- the property that makes it the safe first command |
| `TestVersionFlag` | `--version` prints the linker-stamped build version, so a probe result can be pinned to the code that produced it |
| `TestUnknownProbeIsAUsageError` | A mistyped selector fails with exit 2 BEFORE the env file is resolved, so a typo is never reported as "cannot reach the cluster" |
| `TestResolvePlatform` | Platform resolution: a sole declared section is used silently; several demand `--platform` and say why this tool never prompts; `--platform` must name a declared section and accepts the `pm`-style abbreviations; no section at all is a loud error |
| `TestConsent` | The gate on every mutation: `--yes` approves without asking, an exact `yes` approves, everything looser (`y`, `YES`, `yes please`, empty) declines, a non-interactive run without `--yes` declines, and the question itself names the risk and the exact word required |
| `TestWouldMutateStopsOnlyUnderDryRun` | The `--dry-run` gate is inert in a normal run and stops the probe -- saying what it would have done -- under `--dry-run` |
| `TestVerdictTagsMatchTheHouseLevels` | Verdict -> output tag: VERIFIED `[ OK ]`, REFUTED `[FAIL]`, INCONCLUSIVE `[WARN]`, both skips `[SKIP]`. A refuted claim must read as a failure and an inconclusive one must not -- they call for different work |
| `TestVerdictNames` | The summary-table spelling of each verdict, including `SKIPPED(no consent)` |
| `TestRunProbesSkipsOffPlatformAndExitsOnRefuted` | An off-platform probe does not run but IS still listed as skipped (an omitted row would read as a pass); only REFUTED moves the exit code to 1; INCONCLUSIVE exits 0 |
| `TestP1Classify` | The podman-rotation verdict rule over all four outcomes, including that a refuted verdict names the force-recreate fix and that "neither value" is Inconclusive rather than Refuted |
| `TestP1ThrowawayIsUsable` | The throwaway secret is unique per call and carries no character a shell or the broker CLI treats specially, so a refuted verdict cannot be an artefact of the value chosen |
| `TestP2Classify` | The three SEMP outcomes: `nil` verifies path+regex+idempotence together; a rejected reply is INCONCLUSIVE with the raw body preserved for a human to adjudicate; a transport failure is REFUTED |
| `TestP4Parse` | `podman inspect` parsing: the docker-compatible fields are read, an ABSENT field stays nil (distinct from present-and-zero, which is the whole question), and an empty array or malformed JSON errors |
| `TestP4Classify` | Every cpu/memory branch: both matching, millicore conversion, absent memory naming the PodmanArgs fallback, absent cpu pointing at the manual cgroup check, mismatch showing both numbers, the quota/period form accepted, the worse half winning, and an unparseable configured value being Inconclusive |
| `TestCPUAndMemoryConversion` | Unit conversion both ways -- cores/millicores to nanocpus, `b\|k\|m\|g` to bytes -- and rejection of non-numeric input |

### live_test.go

The probe paths that talk to a live system, over a fake `engine.Runner`.

| Test | What it covers |
| --- | --- |
| `TestP1RunsStopRotateStartReadInOrder` | P1's experiment shape: preflight, stop, rotate, start, wait-active, read, then restore. The order IS the design -- rotating before the stop, or reading before the start, would answer a different question than deployPodman's comment asks |
| `TestP1RestoresAfterAFailureMidSequence` | The important one: a failed `systemctl start` still restores the original secret (two secret-creates) and narrates it, so a failed probe never leaves a live broker on a throwaway password |
| `TestP1DryRunTouchesNothing` | `--dry-run` issues no stop, no secret write and no start, and says what it would have done |
| `TestP1DeclinedConsentTouchesNothing` | A declined prompt is SkippedNoConsent and writes nothing |
| `TestP1WaitActiveIsBounded` | The unit wait gives up after exactly PollAttempts rather than hanging, and rootless polls the user manager (`--user`) |
| `TestP1RestoreWarnsLoudlyWhenItCannotRestore` | When the restore itself fails, the WARN names the secret, because a human now has to fix a live broker by hand |
| `TestP2RefusesWhenTheMateIsActive` | The safety property the whole probe rests on: with the mate active, revert-activity is never sent, and the refusal says why |
| `TestP2SkipsOffPrimaryAndStandalone` | A backup host and a standalone deployment both skip, issuing no commands at all |
| `TestP2RefutesAnUnreachableMate` | An unreachable SEMP endpoint is itself the finding |
| `TestP2DryRunStopsBeforeTheRPC` | The read-only half runs for real under `--dry-run`; the RPC does not |
| `TestRunP4ReadsInspectAndChangesNothing` | P4 end to end: it issues an inspect and nothing that could alter the container -- the one read-only probe stays read-only |
| `TestRunP4SkipsWithNoLimitsConfigured` | With neither limit rendered there is nothing to verify, so it skips rather than claiming a verdict |
| `TestRunP4InconclusiveWhenInspectFails` | An absent container is not evidence the directives are wrong; the detail suggests the likely cause |
| `TestResolveEnvWiresThePlatformsEntryType` | Per platform, exactly one of Cluster/Manager is built and the other stays nil (a probe reads one, so a mis-wire would nil-panic on a live host); Ops.Platform is set for the SEMP port resolution; poll defaults are applied; the preamble goes to stderr; a missing env file errors |
| `TestCtrNameReadsTheConfiguredContainer` | The unit-name helper the systemd polling uses |

### probes_live_test.go

The happy paths through the full probe bodies, plus main's plumbing.

| Test | What it covers |
| --- | --- |
| `TestRunP2VerifiesAnIdempotentRevert` | P2's success path, including the post-RPC re-read that confirms activity did not move |
| `TestRunP2RefutesWhenActivityMoves` | The alarm case: if a standby mate became active, the probe's own premise was wrong, and that outranks whatever the RPC replied |
| `TestRunP2SkipsOnDeclinedConsent` | The gate between the read-only half and the RPC |
| `TestRunEndToEndDryRun` | `run` over a real env file on disk: probe selection, the summary table on stdout, the dry-run notice on stderr, and no `==>` narration leaking into the report a script would capture |
| `TestRunReportsSetupFailureAsUsage` | An env file that will not load exits 2 naming the tool, rather than running probes against a half-built Env |
| `TestParseFlagsCollectsRepeatableAllowCommand` | `--allow-command` accumulates in order, positional probe ids survive parsing, and the stringList flag round-trips |
| `TestParseFlagsRejectsAnUnknownFlag` | An unknown flag errors and is explained on stderr |
| `TestIsTTYIsFalseForANonTerminal` | nil, a regular file and a closed file are all "not a terminal", so a non-interactive run declines rather than panicking |

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
| `TestExamplesCarryNoRealSecret` | The S3 check on shipped text: every template's `admin.pass` is a CHANGE-ME placeholder, so an unedited copy is refused by the broker rather than deployed with a password that is public in this repo |
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
  `internal/examples/assets/full.yaml`, the template `solace-util examples full` prints --
  so a fixture change is an edit to that asset followed by `regen`, never an edit here.
- Tests needing a clean single-broker pass write their own minimal env to a temp file:
  `writeStandaloneEnv` (k8s-shaped) and `writeCtrStandaloneEnv` (container-shaped, needs a
  `nodes:` block) in `internal/cli/cli_test.go`. `writeRuntimeEnv`
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
| internal/k8s | `nsCluster`, `nsBuiltinJSON`, `nsBrokerJSON`, `nsItem` (prep_test.go) | Build the two enumeration replies `DeleteNamespace`'s B6 guard reads -- `nsBuiltinJSON` in `nsRawList`'s per-item-kind shape (`nsResourceKinds`' heterogeneous `get`), `nsBrokerJSON` in `nsCRList`'s homogeneous shape (the broker CR's own separate `get`), `nsItem` builds one entry for either -- and `nsCluster` wires a `Cluster` to a captured `Out` buffer for asserting the foreign-object report, the same shape `labelCluster` gives `LabelNodes` |
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
| internal/cli | `opRunner` / `opCall` / `opFailOn` / `opFailOnCount` (cli_test.go) | A fake `engine.Runner` whose failure is targeted by argv substring (or by the Nth matching occurrence, for the repeated identical `apply`/`delete` calls in `deploy all`/`remove all`/`prepare all`). Ported from internal/container's `capRunner`/`failOn`; this is what makes the orchestration-abort tests possible -- assert step N fails and step N+1 never ran. `Output` answers the k8s `auth can-i` preflight "yes" unless a test supplies its own `output`/`fail` for that call, so the op-level tests stay about the work they were written for |
| internal/cli | `loadDirect`, `healthyShowRD` (cli_test.go) | Loads a config from an inline YAML body for tests that build an `App` directly with a non-Echo runner, and a canned `show redundancy` transcript that satisfies `broker.primaryRedundancyUp` so a poll succeeds on the first read |
| internal/cli | `bashEnv`, `writeBashEnv` (cli_test.go) | Minimal legacy env file for the convert command tests |
| internal/convert | `strictDecode` (convert_test.go) | Re-reads generated YAML with `KnownFields(true)`, so an emitted key that is not in the schema fails the test |
| internal/convert | `ctrEnv`, `convertOK`, `hasWarning` (convert_test.go) | Container-flavoured legacy fixture and warning assertions |
| internal/config | `envTree`, `writeTempYAML`, `minimalK8s` (config_test.go) | Real temp-dir fixture trees for path resolution and loading, plus the smallest valid k8s document the secret-reference tests append a body to |
| internal/config | `controlCharCases` (config_test.go) | The shared table of malformed control-character values (newline, CR, NUL, DEL, ...) every credential and additionalUsers-password rejection test drives against its own field |
| internal/engine | `helperCommand` + `TestHelperProcess` (runner_test.go) | Re-invokes the test binary as a fake external command |

### Injectable seams

Small external effects are seamed as function fields so they are testable off a Linux host.
Override them on the struct after construction:

| Seam | Default | Where |
| --- | --- | --- |
| `Manager.Resolve` | `net.LookupHost` | internal/container -- DNS probes in `Check`/`PrepHost` |
| `Manager.GenPSK` | crypto/rand + base64 | internal/container -- redundancy PSK generation |
| `Manager.Geteuid` | `os.Geteuid` | internal/container -- the rootless/rootful guard (returns -1 on Windows, which skips it) |
| `Ops.Hostname` | `os.Hostname` | internal/broker -- node-role detection in `LocalRole` |
| `Ops.Platform` | `""` (zero value) | internal/broker -- resolves the mate's SEMP port from the bridge network.ports mapping (`sempPort`); set by `ctrOps`, left zero on k8s, which never uses the SEMP channel |
| `App.Interactive` | `isTTY(os.Stdin)` | internal/cli -- whether a run may prompt. Gates `confirmDelete`/`confirmLayer`/`confirmRestart` and the placement-labelling step of `prepare all` (LabelNodes), so the prompt branches guarding destructive actions are testable |
| `App.PromptIn` | `os.Stdin` | internal/cli -- where a confirmation answer is read from |
| `App.kubeContext` | `""` (unresolved) | internal/cli -- the kubeconfig context `announceKubeContext` resolved at load, repeated by every destructive Kubernetes prompt (`k8sWhat`/`k8sContext`). A test sets it directly the same way it sets `Interactive`/`PromptIn`, since a test-supplied runner skips the real `config current-context` lookup |
| `App.Hostname` | `nil` (`ctrOps` then leaves `Ops.Hostname` at its own default, `os.Hostname`) | internal/cli -- same shape as `NewRunner`/`Interactive`/`PromptIn`: a test injects a fixed hostname so `containerRole` (the M1 role detection behind `deploy broker`/`deploy all`/`generate broker` on docker/podman) is testable without depending on the suite's own host being named after a broker node |
| `engine.Runner` | `engine.Exec` | Everywhere -- swapped for `engine.Echo` (dry-run) or a capturing fake |
| `Cluster.Log` / `Manager.Log` / `Ops.Log` | nil (discards) | The RAW progress line sink. It takes an already-formatted line and emits it verbatim: the `==> ` and `[TAG ] ` prefixes are added by the `internal/output` Sink the callee builds over it (`progress()`), so a test that captures this field sees exactly the line a terminal would. The CLI passes `lineSink()`; nil discards |
| `Cluster.Out` / `Manager.Out` / `Ops.Out` | `os.Stdout` | The report sink -- sections, key/value blocks, tables and per-item outcome lines. Wrapped by `report()` |
| `Cluster.Err` | `os.Stderr` | internal/k8s -- where the node picker's prompt (banner, list, `> `) is written. Separate from `Out` so piping stdout captures the report, not the questions; `TestPromptsGoToErrNotOut` pins the split |
| `Cluster.Now` | `time.Now` | internal/k8s -- the clock the status reports' AGE column reads; fixed in tests (`reportCluster`) so it is reproducible against a captured fixture's timestamp |
| `Cluster.Confirm` | `nil` (refuses) | internal/k8s -- the operator-downgrade confirmation `OperatorApply` asks before applying an image older than the one already installed; the same nil-declines shape as `Manager.Confirm` |
| `Consenter.Interactive` | `isTTY(os.Stdin)` | internal/tools/itest -- whether a live probe may ask before mutating. A non-interactive run without `--yes` declines, so the branch guarding every real mutation is testable |
| `Consenter.In` / `Consenter.Err` | `os.Stdin` / `os.Stderr` | internal/tools/itest -- where the exact-`yes` answer is read and the question asked |
| `Env.PollInterval` / `Env.PollAttempts` | 2s / 30 (set by `resolveEnv`) | internal/tools/itest -- the bounded waits for a unit to become active and for config-sync to propagate. Fields for the same reason `Ops.PollInterval` is one: a test sets the interval to 0 so it does not sleep |
| `Env.DryRun` | false | internal/tools/itest -- `--dry-run`; every probe calls `wouldMutate` before its first mutating step |

Filesystem access is *not* seamed: tests use real `t.TempDir()` trees.

# Graph Report - solace-k8-scripts  (2026-09-06)

## Corpus Check
- 135 files · ~301,183 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2392 nodes · 7395 edges · 93 communities (82 shown, 7 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 1197 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `72a3a798`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- cli_test.go
- Sink
- context.Context
- captureStdout
- config_test.go
- Commands
- Manager
- Config
- load.go
- Command
- NewTransport
- NewCluster
- eqArgs
- convert_test.go
- internal/k8s
- manager_test.go
- cli/platform_test.go
- newTestOps
- dev.sh
- prep_test.go
- dev.ps1
- strings.Builder
- judge
- verify_local_test.go
- Role
- Cluster
- Platform
- completion_test.go
- solace
- NodeAffinity
- CLAUDE.md
- .Preflight
- scaling_test.go
- runRootWith
- .resolveSecretRefs
- recRunner
- statusreport_test.go
- Compose
- runner_test.go
- haCfg
- secrets_test.go
- Set
- io.Writer
- .Run
- testing.T
- command_test.go
- README.md
- manager.go
- Probe
- newRootCmd
- podmanEnv
- newEchoMgr
- Operations
- Test catalogue
- container/preflight_test.go
- internal/cli
- internal/broker
- internal/config
- internal/container
- operator.go
- Troubleshooting
- TestManagerDeployDockerPassesSecretsAsEnv
- Configuration
- runP1
- time.Time
- Command reference
- Developer guide
- semp_test.go
- Live-environment probe harness (`solace-itest`)
- runP2
- scripts.go
- resolveEnv
- New
- Fixtures and doubles
- internal/tools/itest
- itest_test.go
- age
- inspect_test.go
- Ops
- inspect.go
- kubectlTransport
- TestDefaultGenPSK
- render.go
- .releaseToBackup
- Abbreviations
- internal/engine
- .MateRevertActivity
- NewManager

## God Nodes (most connected - your core abstractions)
1. `ctrCfg()` - 98 edges
2. `newTestOps()` - 90 edges
3. `newCapMgr()` - 89 edges
4. `Config` - 86 edges
5. `Role` - 85 edges
6. `NewCluster()` - 76 edges
7. `Platform` - 75 edges
8. `bg()` - 69 edges
9. `Commands` - 68 edges
10. `Manager` - 65 edges

## Surprising Connections (you probably didn't know these)
- `TestNewDefaults()` --calls--> `New()`  [INFERRED]
  internal/broker/coverage_test.go → internal/broker/broker.go
- `TestField()` --calls--> `field()`  [INFERRED]
  internal/broker/broker_test.go → internal/broker/broker.go
- `TestFieldLabelWithoutColon()` --calls--> `field()`  [INFERRED]
  internal/broker/coverage_test.go → internal/broker/broker.go
- `TestCountContains()` --calls--> `countContains()`  [INFERRED]
  internal/broker/broker_test.go → internal/broker/broker.go
- `activity()` --calls--> `countContains()`  [INFERRED]
  internal/broker/verify_ops.go → internal/broker/broker.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (93 total, 7 thin omitted)

### Community 0 - "bg"
Cohesion: 0.06
Nodes (114): TestConfirmFlagShortcuts(), TestConfirmNonTTY(), TestContainerRoleDetectsFromHostname(), TestK8sLoginOutcomes(), TestWarnAndStep(), App, lineSink(), progress() (+106 more)

### Community 1 - "cli_test.go"
Cohesion: 0.09
Nodes (38): collectPaths(), firstLine(), runRoot(), TestBashEnvGivenToEnvFlag(), TestConvertErrorPaths(), TestConvertParseError(), TestConvertRoundTrip(), TestConvertToFile() (+30 more)

### Community 2 - "Sink"
Cohesion: 0.17
Nodes (6): Level, Sink, pad(), TestLevelTagsMapToTheHouseTags(), KV, Row

### Community 3 - "context.Context"
Cohesion: 0.10
Nodes (13): context.Context, Cluster, Cluster, HARoles(), lbServiceName(), podName(), ProductKeyRoles(), pvcName() (+5 more)

### Community 4 - "captureStdout"
Cohesion: 0.14
Nodes (20): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch(), opCanI() (+12 more)

### Community 5 - "config_test.go"
Cohesion: 0.05
Nodes (70): assertContainerBlockDefaults(), assertContainerScaling(), envTree(), Config, haNodesConfig(), minimalK8s(), TestApplyDefaultsDocker(), TestApplyDefaultsK8s() (+62 more)

### Community 6 - "Commands"
Cohesion: 0.03
Nodes (68): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util check (+60 more)

### Community 8 - "Config"
Cohesion: 0.11
Nodes (26): Admin, Broker, Container, ContainerSecurity, DockerConfig, DomainCerts, Image, Network (+18 more)

### Community 9 - "load.go"
Cohesion: 0.12
Nodes (19): scalingTier, TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config (+11 more)

### Community 10 - "Command"
Cohesion: 0.06
Nodes (49): commandRules, keyValueEntries, Command, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules() (+41 more)

### Community 11 - "NewTransport"
Cohesion: 0.18
Nodes (15): capCall, capRunner, New(), Transport, TestManagerLogsCLIShell(), NewTransport(), dockerCfg(), eqArgs() (+7 more)

### Community 12 - "NewCluster"
Cohesion: 0.10
Nodes (46): TestCheckStopsProbingWhenUnreachable(), TestResolveStorageClass(), TestCheckDeployConfigSectionNeverFails(), TestCheckDeployGroupsAndOrdersSections(), TestCheckDeployNeverPrintsASecret(), TestCheckDeployReportsDefaultPorts(), TestCheckDeployReportsEveryFailureInOneRun(), TestCheckDeployReportsResolvedPorts() (+38 more)

### Community 13 - "eqArgs"
Cohesion: 0.17
Nodes (23): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), TestCLIAndShellAreInteractive() (+15 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (63): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+55 more)

### Community 15 - "internal/k8s"
Cohesion: 0.12
Nodes (16): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, names_test.go, operator_test.go (+8 more)

### Community 16 - "manager_test.go"
Cohesion: 0.10
Nodes (76): fileExists(), containsStr(), ctrCfg(), hasCall(), newCapMgr(), TestContainerRunningMatchesNameExactly(), TestManagerCheckDNSFailsLoudInHA(), TestManagerCheckPodmanEUID() (+68 more)

### Community 17 - "cli/platform_test.go"
Cohesion: 0.26
Nodes (17): App, runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection() (+9 more)

### Community 18 - "newTestOps"
Cohesion: 0.08
Nodes (52): appUsers(), Ops, newTestOps(), ranContains(), TestAdditionalUsers(), TestAdditionalUsersEmpty(), TestAdditionalUsersRejectsBadValues(), TestAdditionalUsersReportsExistingUser() (+44 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (22): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+14 more)

### Community 20 - "prep_test.go"
Cohesion: 0.13
Nodes (35): saCfg(), adminCfg(), Cluster, labelCluster(), nsBrokerJSON(), nsBuiltinJSON(), nsCluster(), nsItem() (+27 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (18): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+10 more)

### Community 22 - "strings.Builder"
Cohesion: 0.18
Nodes (24): shorthand, github.com/spf13/pflag.FlagSet, strings.Builder, mdRow(), renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands() (+16 more)

### Community 23 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.17
Nodes (36): uploadedForRole(), curlCalls(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestLeaderLocalAssertLeaderError() (+28 more)

### Community 25 - "Role"
Cohesion: 0.11
Nodes (14): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+6 more)

### Community 27 - "Platform"
Cohesion: 0.09
Nodes (77): layer, opFunc, roleOpFunc, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, applyAliases(), addCommands(), addPodFlag() (+69 more)

### Community 28 - "completion_test.go"
Cohesion: 0.10
Nodes (40): Example, runComplete(), TestAllowCommandOffersNoFiles(), TestDirFlagCompletesDirectories(), TestEnvFlagCompletesEnvFiles(), TestEnvFlagPrefixFilters(), TestEnvFlagWithPathDefersToShell(), TestExampleArgsComplete() (+32 more)

### Community 31 - "NodeAffinity"
Cohesion: 0.31
Nodes (8): WeightedNodeTerm, NodeAffinity, NodeMatchExpr, Placement, validateMatchExprs(), validatePlacementAffinity(), writeMatchExprs(), writeNodeAffinity()

### Community 32 - "CLAUDE.md"
Cohesion: 0.25
Nodes (5): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh

### Community 34 - "scaling_test.go"
Cohesion: 0.17
Nodes (12): Config, setContainerMem(), TestApplyScalingTierDefaultsContainerBlocks(), TestApplyScalingTierDefaultsK8s(), TestApplyScalingTierDefaultsMemOverride(), TestApplyScalingTierDefaultsOffTier(), TestScalingTierListMatchesTable(), TestValidateContainerMem() (+4 more)

### Community 35 - "runRootWith"
Cohesion: 0.17
Nodes (23): TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), captureStderr(), echoRunner(), App, runRootWith(), runStatusStderr(), TestCLICommand() (+15 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.16
Nodes (11): Runner, TestCanIAnswerReadsTheLastLine(), NewTransport(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs(), TestTransportUpload() (+3 more)

### Community 38 - "statusreport_test.go"
Cohesion: 0.10
Nodes (23): operatorRunningImage(), ownsPod(), Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure() (+15 more)

### Community 39 - "Compose"
Cohesion: 0.12
Nodes (32): opCtrGenArtifact(), opCtrGenSecrets(), Config, NodeIdentity, Compose(), ContainerSecrets(), EnvPairs(), ContainerSecret (+24 more)

### Community 40 - "runner_test.go"
Cohesion: 0.08
Nodes (37): Echo, Exec, os/exec.Cmd, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveMissingBinaryIsActionable() (+29 more)

### Community 41 - "haCfg"
Cohesion: 0.13
Nodes (22): haCfg(), TestHARoles(), TestProductKeyRoles(), TestRestartOrder(), TestOperatorImage(), operatorDeployJSON(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsSilentInAPreview() (+14 more)

### Community 42 - "secrets_test.go"
Cohesion: 0.14
Nodes (23): GenOperatorSecrets(), TestGenOperatorSecrets(), GenSecrets(), TestCreateSecretsAllThree(), TestGenSecretsTLSError(), TestUpdateServerCertSecret(), AdminSecret(), DockerRegistrySecret() (+15 more)

### Community 43 - "Set"
Cohesion: 0.10
Nodes (20): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+12 more)

### Community 44 - "io.Writer"
Cohesion: 0.06
Nodes (31): bufio.Reader, io.Reader, io.Writer, os.File, capture(), TestIsTTYClosedFile(), TestPromptYes(), TestPromptYesNo() (+23 more)

### Community 45 - ".Run"
Cohesion: 0.12
Nodes (27): allowRuntime(), fakeBinaryOnPath(), runCtr(), runStandalone(), TestAnnounceCommandsNamesResolvedBinaries(), TestBinaryAnnouncementWiring(), TestConfigStepsDoNotLeakSecrets(), TestConfirmDowngradeNeverUnattended() (+19 more)

### Community 46 - "testing.T"
Cohesion: 0.09
Nodes (52): testing.T, matchCLI(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultUsersShowVPNError() (+44 more)

### Community 47 - "command_test.go"
Cohesion: 0.16
Nodes (13): decodeRuntime(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRuntimeDefaults() (+5 more)

### Community 48 - "README.md"
Cohesion: 0.22
Nodes (6): internal/broker, internal/cli, internal/config, internal/engine, internal/k8s, internal/render

### Community 49 - "manager.go"
Cohesion: 0.17
Nodes (11): exactName(), orNone(), platformTitle(), replacePSKLine(), secretSummary(), setOrMissing(), solaceRows(), splitLimit() (+3 more)

### Community 50 - "Probe"
Cohesion: 0.17
Nodes (19): ids(), TestRegistryCoversEveryClaim(), TestRunProbesSkipsOffPlatformAndExitsOnRefuted(), TestSelectProbes(), probeP1(), probeP2(), probeP4(), Env (+11 more)

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (29): TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestDangerousVerbsHaveNoBareAlias(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownNoun(), TestStartStopHaveNoAlias(), TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRegisteredWhereItExecutes() (+21 more)

### Community 52 - "podmanEnv"
Cohesion: 0.16
Nodes (19): Env, podmanEnv(), TestCtrNameReadsTheConfiguredContainer(), TestP1DeclinedConsentTouchesNothing(), TestP1DryRunTouchesNothing(), TestP1RestoresAfterAFailureMidSequence(), TestP1RestoreWarnsLoudlyWhenItCannotRestore(), TestP1RunsStopRotateStartReadInOrder() (+11 more)

### Community 53 - "newEchoMgr"
Cohesion: 0.10
Nodes (21): bytes.Buffer, Manager, newEchoMgr(), rootlessNoFileMgr(), setNoFile(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployDockerDryRunMasksSecretEnv() (+13 more)

### Community 54 - "Operations"
Cohesion: 0.20
Nodes (10): Bringing up a fresh cluster, Docker and Podman mechanics, Extra CLI users differ by platform, Operations, Post-deployment configuration order, Re-deploying is safe and explicit, Removing a broker: what stays, what goes, Rendering without applying (+2 more)

### Community 55 - "Test catalogue"
Cohesion: 0.12
Nodes (17): abbrev_test.go, convert_test.go, Coverage, examples_test.go, examples_test.go, internal/abbrev, internal/convert, internal/examples (+9 more)

### Community 56 - "container/preflight_test.go"
Cohesion: 0.25
Nodes (7): TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestComposeSecretEnvNamesCannotBeSystemVars(), TestPreflightFailureStopsLifecycle(), TestPreflightFailureStopsTheDeploy(), TestPreflightHintIsPlatformShaped(), TestPreflightIsPreviewableUnderDryRun(), TestPreflightRunsBeforeAnything()

### Community 57 - "internal/cli"
Cohesion: 0.25
Nodes (8): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, internal/cli, platform_test.go

### Community 58 - "internal/broker"
Cohesion: 0.33
Nodes (6): broker_test.go, coverage_test.go, internal/broker, scripts_test.go, semp_test.go, verify_local_test.go

### Community 59 - "internal/config"
Cohesion: 0.33
Nodes (6): command_test.go, config_test.go, execguard_test.go, internal/config, platform_test.go, scaling_test.go

### Community 60 - "internal/container"
Cohesion: 0.40
Nodes (5): internal/container, manager_test.go, preflight_test.go, runtime_test.go, transport_test.go

### Community 61 - "operator.go"
Cohesion: 0.09
Nodes (34): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), failRow(), Cluster, info(), okRow() (+26 more)

### Community 62 - "Troubleshooting"
Cohesion: 0.33
Nodes (6): A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, File descriptors on rootless podman, Podman secret flags, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 63 - "TestManagerDeployDockerPassesSecretsAsEnv"
Cohesion: 0.50
Nodes (4): assertMode(), maskedKeys(), TestManagerDeployDockerPassesSecretsAsEnv(), TestManagerDeployPodmanCreatesSecrets()

### Community 65 - "Configuration"
Cohesion: 0.22
Nodes (9): Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Scaling, Scaling tiers, Secrets, The command fields are executable content, The keys (+1 more)

### Community 66 - "runP1"
Cohesion: 0.33
Nodes (10): ctrName(), Env, Env, Result, p1Classify(), p1ReadMountedSecret(), p1Restore(), p1Throwaway() (+2 more)

### Community 67 - "time.Time"
Cohesion: 0.33
Nodes (4): time.Time, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 71 - "semp_test.go"
Cohesion: 0.11
Nodes (31): sempPort(), sempV1OK(), assertNoPasswordInArgv(), TestCurlConfigLineEscapesQuotes(), TestHTTPBodyStripsHeaders(), TestLastHTTPStatus(), TestMateRevertActivityBridgePlaintextOnly(), TestMateRevertActivityBridgeTLSNoCA() (+23 more)

### Community 72 - "Live-environment probe harness (`solace-itest`)"
Cohesion: 0.20
Nodes (10): After a run, Build and run, Flags, Live-environment probe harness (`solace-itest`), Operator-run only, Safety, Testing the harness itself, The probes (+2 more)

### Community 73 - "runP2"
Cohesion: 0.15
Nodes (16): containsPlatform(), platformNames(), resolvePlatform(), roleLabel(), TestP2Classify(), TestResolvePlatform(), writeEnv(), Env (+8 more)

### Community 74 - "scripts.go"
Cohesion: 0.06
Nodes (45): showCmd, containsAnyFold(), countContains(), validCLILine(), validCLIPassword(), validName(), concatFiles(), countAnyFold() (+37 more)

### Community 77 - "resolveEnv"
Cohesion: 0.14
Nodes (17): Env, resolveEnv(), TestListProbesNeedsNoEnvFile(), TestUnknownProbeIsAUsageError(), TestVerdictNames(), TestVersionFlag(), Env, main() (+9 more)

### Community 78 - "New"
Cohesion: 0.16
Nodes (19): TestCheckReportSkipsEmptySections(), New(), NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow() (+11 more)

### Community 79 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 80 - "internal/tools/itest"
Cohesion: 0.50
Nodes (4): internal/tools/itest, itest_test.go, live_test.go, probes_live_test.go

### Community 81 - "itest_test.go"
Cohesion: 0.16
Nodes (19): Env, sinkOver(), TestConsent(), TestCPUAndMemoryConversion(), testEnv(), TestP1Classify(), TestP1ThrowawayIsUsable(), TestP4Classify() (+11 more)

### Community 82 - "age"
Cohesion: 0.37
Nodes (4): orNone(), age(), roleRank(), Cluster

### Community 83 - "inspect_test.go"
Cohesion: 0.18
Nodes (17): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+9 more)

### Community 86 - "inspect.go"
Cohesion: 0.26
Nodes (14): brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), ownedPods(), brokerList, brokerStatus, condition (+6 more)

### Community 89 - "render.go"
Cohesion: 0.13
Nodes (27): PodAffinityTerm, boolStr(), BrokerCR(), containerSecretSpecs(), cut(), escapePercent(), groupKey(), itoa() (+19 more)

### Community 91 - ".releaseToBackup"
Cohesion: 0.15
Nodes (13): field(), Ops, hostMatches(), roleName(), shortHost(), TestRoleName(), activity(), Ops (+5 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 100 - "internal/engine"
Cohesion: 0.67
Nodes (3): internal/engine, resolve_test.go, runner_test.go

### Community 103 - ".MateRevertActivity"
Cohesion: 0.28
Nodes (10): mateTarget, anyHTTP2xx(), bridgeHostPort(), curlConfigFlag(), curlConfigLine(), Ops, httpBody(), lastHTTPStatus() (+2 more)

### Community 106 - "NewManager"
Cohesion: 0.26
Nodes (12): NewManager(), TestManagerNilSinks(), Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime() (+4 more)

## Knowledge Gaps
- **179 isolated node(s):** `solace`, `showCmd`, `Config`, `operatorTmplVars`, `NO_COLOR` (+174 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 249 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **7 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `bg`, `captureStdout`, `config_test.go`, `Manager`, `Config`, `load.go`, `Command`, `NewTransport`, `convert_test.go`, `manager_test.go`, `Role`, `completion_test.go`, `scaling_test.go`, `Compose`, `manager.go`, `Probe`, `newEchoMgr`, `semp_test.go`, `runP2`, `Ops`, `NewManager`?**
  _High betweenness centrality (0.071) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `bg`, `context.Context`, `captureStdout`, `Manager`, `NewTransport`, `NewCluster`, `convert_test.go`, `manager_test.go`, `newTestOps`, `prep_test.go`, `Role`, `Cluster`, `recRunner`, `Compose`, `haCfg`, `secrets_test.go`, `io.Writer`, `.Run`, `Probe`, `newEchoMgr`, `operator.go`, `semp_test.go`, `Ops`, `kubectlTransport`, `render.go`, `NewManager`?**
  _High betweenness centrality (0.059) - this node is a cross-community bridge._
- **Why does `Manager` connect `Manager` to `bg`, `recRunner`, `Config`, `NewManager`, `io.Writer`, `manager.go`, `Probe`, `Platform`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **Are the 10 inferred relationships involving `ctrCfg()` (e.g. with `TestComposeSecretEnvIsTheOnlyChildEnvironment()` and `TestComposeSecretEnvNamesCannotBeSystemVars()`) actually correct?**
  _`ctrCfg()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **Are the 50 inferred relationships involving `newTestOps()` (e.g. with `TestAdditionalUsersRunCLITransportError()` and `TestDiagnosticsMkdirError()`) actually correct?**
  _`newTestOps()` has 50 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `newCapMgr()` (e.g. with `NewManager()` and `TestComposeSecretEnvIsTheOnlyChildEnvironment()`) actually correct?**
  _`newCapMgr()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `showCmd`, `Config` to the rest of the system?**
  _179 weakly-connected nodes found - possible documentation gaps or missing edges._
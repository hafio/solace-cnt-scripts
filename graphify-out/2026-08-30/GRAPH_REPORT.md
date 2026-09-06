# Graph Report - solace-k8-scripts  (2026-08-29)

## Corpus Check
- 125 files · ~248,789 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2196 nodes · 6702 edges · 91 communities (85 shown, 6 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 1064 edges (avg confidence: 0.8)
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
- newTestOps
- NewTransport
- NewCluster
- eqArgs
- convert_test.go
- internal/k8s
- manager_test.go
- cli/platform_test.go
- secrets_test.go
- dev.sh
- prep_test.go
- dev.ps1
- CLAUDE.md
- judge
- verify_local_test.go
- Role
- .Run
- github.com/spf13/cobra.Command
- strings.Builder
- solace
- CheckCommand
- Command reference
- .Preflight
- completion_test.go
- .Run
- .resolveSecretRefs
- recRunner
- tierFor
- Compose
- Cluster
- k8s/runtime_test.go
- semp_test.go
- validate.go
- Cluster
- captureStderr
- coverage_test.go
- testing.T
- resolveEnv
- manager.go
- Probe
- newRootCmd
- podmanEnv
- execguard_test.go
- Operations
- Test catalogue
- scripts.go
- internal/cli
- internal/broker
- internal/config
- internal/container
- .configRows
- operator.go
- eqArgs
- Configuration
- runP1
- containerTransport
- allowcommand_test.go
- Developer guide
- Platform
- .AdditionalUsers
- Live-environment probe harness (`solace-itest`)
- itest_test.go
- time.Time
- Image
- TestCheckStorageClass
- TestShowRedundancyIsReadOnly
- New
- Fixtures and doubles
- internal/tools/itest
- render.go
- age
- inspect_test.go
- io.Writer
- statusreport_test.go
- inspect.go
- runP2
- Ops

## God Nodes (most connected - your core abstractions)
1. `ctrCfg()` - 91 edges
2. `newTestOps()` - 83 edges
3. `newCapMgr()` - 82 edges
4. `Config` - 81 edges
5. `Role` - 81 edges
6. `Platform` - 75 edges
7. `bg()` - 68 edges
8. `Commands` - 67 edges
9. `Manager` - 61 edges
10. `NewCluster()` - 58 edges

## Surprising Connections (you probably didn't know these)
- `TestFieldLabelWithoutColon()` --calls--> `field()`  [INFERRED]
  internal/broker/coverage_test.go → internal/broker/broker.go
- `activity()` --calls--> `countContains()`  [INFERRED]
  internal/broker/verify_ops.go → internal/broker/broker.go
- `matchCLI()` --calls--> `cliArg()`  [INFERRED]
  internal/broker/broker_test.go → internal/broker/transport.go
- `TestAdditionalUsersRunCLITransportError()` --calls--> `matchCLI()`  [INFERRED]
  internal/broker/coverage_test.go → internal/broker/broker_test.go
- `TestDisableDefaultUsersShowVPNError()` --calls--> `matchCLI()`  [INFERRED]
  internal/broker/coverage_test.go → internal/broker/broker_test.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (91 total, 6 thin omitted)

### Community 0 - "bg"
Cohesion: 0.06
Nodes (100): TestCtrManagerConfirmWiring(), TestWarnAndStep(), lineSink(), progress(), step(), warn(), App, newConvertCmd() (+92 more)

### Community 1 - "cli_test.go"
Cohesion: 0.14
Nodes (32): firstLine(), runCtr(), runRoot(), TestBashEnvGivenToEnvFlag(), TestConfigStepsDoNotLeakSecrets(), TestConvertErrorPaths(), TestConvertParseError(), TestConvertRoundTrip() (+24 more)

### Community 2 - "Sink"
Cohesion: 0.19
Nodes (6): Level, Sink, pad(), TestLevelTagsMapToTheHouseTags(), KV, Row

### Community 3 - "context.Context"
Cohesion: 0.10
Nodes (7): context.Context, Cluster, Cluster, Cluster, Cluster, badJSONRunner, fixtureRunner

### Community 4 - "captureStdout"
Cohesion: 0.14
Nodes (20): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch(), opCanI() (+12 more)

### Community 5 - "config_test.go"
Cohesion: 0.05
Nodes (77): assertContainerBlockDefaults(), assertContainerScaling(), envTree(), Config, haNodesConfig(), minimalK8s(), TestApplyDefaultsDocker(), TestApplyDefaultsK8s() (+69 more)

### Community 6 - "Commands"
Cohesion: 0.03
Nodes (67): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util check (+59 more)

### Community 8 - "Config"
Cohesion: 0.11
Nodes (24): Admin, Broker, Container, ContainerSecurity, DockerConfig, DomainCerts, Network, Node (+16 more)

### Community 9 - "load.go"
Cohesion: 0.30
Nodes (10): TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), setDefault(), setDefaultCmd() (+2 more)

### Community 10 - "newTestOps"
Cohesion: 0.09
Nodes (45): appUsers(), Ops, newTestOps(), TestAdditionalUsers(), TestAdditionalUsersEmpty(), TestAdditionalUsersRejectsBadValues(), TestAdditionalUsersReportsExistingUser(), TestDiagnostics() (+37 more)

### Community 11 - "NewTransport"
Cohesion: 0.19
Nodes (14): capCall, capRunner, New(), TestNewDefaults(), Transport, NewTransport(), dockerCfg(), podmanCfg() (+6 more)

### Community 12 - "NewCluster"
Cohesion: 0.12
Nodes (37): TestCheckDeployGroupsAndOrdersSections(), TestCheckDeployNeverPrintsASecret(), TestCheckDeployReportsEveryFailureInOneRun(), TestCheckDeploySparseConfigExplainsItself(), TestCheckDeployTagsEchoedConfigAsINFO(), NewCluster(), logBuf(), TestDeleteBrokerLogsPVCOutcome() (+29 more)

### Community 13 - "eqArgs"
Cohesion: 0.15
Nodes (26): Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSDefaultOnError(), TestOperatorNSDefaultWhenAbsent(), TestOperatorNSDerived(), TestOperatorNSExplicit() (+18 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (58): doc, Result, vars, boolOf(), commentSafe(), Convert(), countMarkers(), emitYAML() (+50 more)

### Community 15 - "internal/k8s"
Cohesion: 0.17
Nodes (12): check_test.go, cluster_test.go, deploy_test.go, internal/k8s, names_test.go, operator_test.go, ops_test.go, preflight_test.go (+4 more)

### Community 16 - "manager_test.go"
Cohesion: 0.07
Nodes (94): fileExists(), NewManager(), assertMode(), containsStr(), ctrCfg(), hasCall(), maskedKeys(), newCapMgr() (+86 more)

### Community 17 - "cli/platform_test.go"
Cohesion: 0.35
Nodes (13): App, runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection() (+5 more)

### Community 18 - "secrets_test.go"
Cohesion: 0.16
Nodes (21): GenSecrets(), TestCreateSecretsAllThree(), TestGenSecretsTLSError(), TestUpdateServerCertSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), operatorRegcred() (+13 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (22): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+14 more)

### Community 20 - "prep_test.go"
Cohesion: 0.16
Nodes (26): haCfg(), saCfg(), TestHARoles(), TestProductKeyRoles(), TestRestartOrder(), adminCfg(), Cluster, labelCluster() (+18 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (18): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+10 more)

### Community 22 - "CLAUDE.md"
Cohesion: 0.12
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 23 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.19
Nodes (34): curlCalls(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestLeaderLocalAssertLeaderError(), TestLeaderLocalBadRoleArg() (+26 more)

### Community 25 - "Role"
Cohesion: 0.11
Nodes (14): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+6 more)

### Community 26 - ".Run"
Cohesion: 0.12
Nodes (17): Echo, Exec, os/exec.Cmd, App, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand() (+9 more)

### Community 27 - "github.com/spf13/cobra.Command"
Cohesion: 0.06
Nodes (110): layer, opFunc, roleOpFunc, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, github.com/spf13/pflag.FlagSet, os.File, applyAliases() (+102 more)

### Community 28 - "strings.Builder"
Cohesion: 0.20
Nodes (16): WeightedNodeTerm, strings.Builder, NodeAffinity, NodeMatchExpr, Placement, PodAffinityTerm, parseToleration(), sortedKeys() (+8 more)

### Community 31 - "CheckCommand"
Cohesion: 0.18
Nodes (15): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), execBase() (+7 more)

### Community 32 - "Command reference"
Cohesion: 0.24
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 34 - "completion_test.go"
Cohesion: 0.24
Nodes (16): runComplete(), TestAllowCommandOffersNoFiles(), TestCompletionHelpStillWorks(), TestCompletionNeedsAShell(), TestCompletionNoDescriptions(), TestCompletionScriptsGenerate(), TestDirFlagCompletesDirectories(), TestEnvFlagCompletesEnvFiles() (+8 more)

### Community 35 - ".Run"
Cohesion: 0.19
Nodes (21): echoRunner(), App, runRootWith(), TestCLICommand(), TestCtrConfirmDeclined(), TestErrorPaths(), TestFirstArgOr(), TestGenerateWired() (+13 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.18
Nodes (9): NewTransport(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest(), recRunner (+1 more)

### Community 38 - "tierFor"
Cohesion: 0.27
Nodes (7): scalingTier, containerMem(), Config, TestContainerMem(), TestScalingTiers(), TestTierForRejectsOffTierValues(), tierFor()

### Community 39 - "Compose"
Cohesion: 0.17
Nodes (25): NodeIdentity, Compose(), ContainerSecrets(), healthCmd(), Quadlet(), SecretPreflight(), SecretScript(), shQuote() (+17 more)

### Community 40 - "Cluster"
Cohesion: 0.14
Nodes (10): Cluster, HARoles(), lbServiceName(), podName(), ProductKeyRoles(), pvcName(), RestartOrder(), stsName() (+2 more)

### Community 41 - "k8s/runtime_test.go"
Cohesion: 0.43
Nodes (7): TestClusterHonoursRuntime(), TestExecutorRefusesUnapprovedRuntime(), TestRuntimeDefaultArgvUnchanged(), TestTransportHonoursRuntime(), unapprovedCfg(), withLeading(), wrappedCfg()

### Community 42 - "semp_test.go"
Cohesion: 0.10
Nodes (31): TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigLine(), Ops, httpBody(), lastHTTPStatus(), revertActivityMateBody(), sempPort() (+23 more)

### Community 43 - "validate.go"
Cohesion: 0.24
Nodes (10): keyValueEntries, foldToEnvVar(), Config, missingErr(), requireAll(), requireKeyValue(), sortStrings(), validateMatchExprs() (+2 more)

### Community 44 - "Cluster"
Cohesion: 0.11
Nodes (12): bufio.Reader, Cluster, isBuiltinLabel(), joinManifests(), namespaceManifest(), roleName(), rolePlacementLabels(), splitLabel() (+4 more)

### Community 45 - "captureStderr"
Cohesion: 0.18
Nodes (17): allowRuntime(), capture(), captureStderr(), fakeBinaryOnPath(), runStandalone(), runStatusStderr(), TestAnnounceCommandsNamesResolvedBinaries(), TestBinaryAnnouncementWiring() (+9 more)

### Community 46 - "coverage_test.go"
Cohesion: 0.06
Nodes (50): matchCLI(), ranContains(), TestExecCLI(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError() (+42 more)

### Community 47 - "testing.T"
Cohesion: 0.10
Nodes (40): testing.T, decodeRuntime(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds() (+32 more)

### Community 48 - "resolveEnv"
Cohesion: 0.14
Nodes (17): Env, resolveEnv(), TestListProbesNeedsNoEnvFile(), TestUnknownProbeIsAUsageError(), TestVerdictNames(), TestVersionFlag(), Env, main() (+9 more)

### Community 49 - "manager.go"
Cohesion: 0.19
Nodes (10): defaultGenPSK(), exactName(), orNone(), platformTitle(), replacePSKLine(), secretSummary(), setOrMissing(), solaceRows() (+2 more)

### Community 50 - "Probe"
Cohesion: 0.17
Nodes (19): ids(), TestRegistryCoversEveryClaim(), TestRunProbesSkipsOffPlatformAndExitsOnRefuted(), TestSelectProbes(), probeP1(), probeP2(), probeP4(), Env (+11 more)

### Community 51 - "newRootCmd"
Cohesion: 0.11
Nodes (23): TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestDangerousVerbsHaveNoBareAlias(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownNoun(), TestStartStopHaveNoAlias(), TestAllowCommandIsRegisteredWhereItExecutes(), TestAllowCommandIsRepeatable() (+15 more)

### Community 52 - "podmanEnv"
Cohesion: 0.16
Nodes (19): Env, podmanEnv(), TestCtrNameReadsTheConfiguredContainer(), TestP1DeclinedConsentTouchesNothing(), TestP1DryRunTouchesNothing(), TestP1RestoresAfterAFailureMidSequence(), TestP1RestoreWarnsLoudlyWhenItCannotRestore(), TestP1RunsStopRotateStartReadInOrder() (+11 more)

### Community 53 - "execguard_test.go"
Cohesion: 0.19
Nodes (15): decodeStrict(), Config, guardCommandOf(), guardConfig(), setGuardCommand(), TestAllowCommandIsNotASchemaKey(), TestAllowCommandsAccepts(), TestAllowCommandsRejects() (+7 more)

### Community 54 - "Operations"
Cohesion: 0.13
Nodes (15): A wrapper runtime is refused, Bringing up a fresh cluster, Docker and Podman mechanics, Docker compose secrets need compose 2.23.1+, Extra CLI users differ by platform, File descriptors on rootless podman, Operations, Podman secret flags (+7 more)

### Community 55 - "Test catalogue"
Cohesion: 0.13
Nodes (15): convert_test.go, Coverage, internal/convert, internal/engine, internal/output, internal/render, internal/tools/vulnjudge, main_test.go (+7 more)

### Community 56 - "scripts.go"
Cohesion: 0.05
Nodes (51): showCmd, field(), TestField(), TestLastLines(), TestPrimaryRedundancyUp(), concatFiles(), Ops, TestLastLinesEqualCount() (+43 more)

### Community 57 - "internal/cli"
Cohesion: 0.29
Nodes (7): aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, internal/cli, platform_test.go

### Community 58 - "internal/broker"
Cohesion: 0.33
Nodes (6): broker_test.go, coverage_test.go, internal/broker, scripts_test.go, semp_test.go, verify_local_test.go

### Community 59 - "internal/config"
Cohesion: 0.33
Nodes (6): command_test.go, config_test.go, execguard_test.go, internal/config, platform_test.go, scaling_test.go

### Community 60 - "internal/container"
Cohesion: 0.40
Nodes (5): internal/container, manager_test.go, preflight_test.go, runtime_test.go, transport_test.go

### Community 61 - ".configRows"
Cohesion: 0.14
Nodes (21): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), failRow(), Cluster, info(), okRow() (+13 more)

### Community 62 - "operator.go"
Cohesion: 0.23
Nodes (13): GenOperator(), GenOperatorSecrets(), joinYAMLDocs(), nonEmpty(), operatorImage(), RenderOperator(), renderOperatorNS(), splitAfterNamespace() (+5 more)

### Community 63 - "eqArgs"
Cohesion: 0.25
Nodes (13): TestManagerLogsCLIShell(), TestPreflightRunsBeforeAnything(), Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime() (+5 more)

### Community 65 - "Configuration"
Cohesion: 0.22
Nodes (9): Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Scaling, Scaling tiers, Secrets, The command fields are executable content, The keys (+1 more)

### Community 66 - "runP1"
Cohesion: 0.33
Nodes (10): ctrName(), Env, Env, Result, p1Classify(), p1ReadMountedSecret(), p1Restore(), p1Throwaway() (+2 more)

### Community 68 - "allowcommand_test.go"
Cohesion: 0.33
Nodes (9): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused(), TestSmuggledSubcommandIsRefused() (+1 more)

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "Platform"
Cohesion: 0.36
Nodes (3): Platform, Config, platformKey()

### Community 71 - ".AdditionalUsers"
Cohesion: 0.23
Nodes (9): containsAnyFold(), countContains(), TestContainsAnyFold(), TestCountContains(), validCLILine(), validCLIPassword(), additionalUsersScript(), TestAdditionalUsersScript() (+1 more)

### Community 72 - "Live-environment probe harness (`solace-itest`)"
Cohesion: 0.20
Nodes (10): After a run, Build and run, Flags, Live-environment probe harness (`solace-itest`), Operator-run only, Safety, Testing the harness itself, The probes (+2 more)

### Community 73 - "itest_test.go"
Cohesion: 0.16
Nodes (19): Env, sinkOver(), TestConsent(), TestCPUAndMemoryConversion(), testEnv(), TestP1Classify(), TestP1ThrowawayIsUsable(), TestP4Classify() (+11 more)

### Community 74 - "time.Time"
Cohesion: 0.33
Nodes (4): time.Time, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

### Community 76 - "TestCheckStorageClass"
Cohesion: 0.40
Nodes (4): TestCheckStopsProbingWhenUnreachable(), TestCheckStorageClass(), TestReachable(), TestResolveStorageClass()

### Community 77 - "TestShowRedundancyIsReadOnly"
Cohesion: 0.67
Nodes (3): TestMateActivityStateReadsTheMateColumn(), TestShowRedundancyIsReadOnly(), MateActivityState()

### Community 78 - "New"
Cohesion: 0.15
Nodes (18): TestCheckReportSkipsEmptySections(), New(), NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow() (+10 more)

### Community 79 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 80 - "internal/tools/itest"
Cohesion: 0.50
Nodes (4): internal/tools/itest, itest_test.go, live_test.go, probes_live_test.go

### Community 81 - "render.go"
Cohesion: 0.13
Nodes (23): LoadBalancer, boolStr(), BrokerCR(), containerSecretSpecs(), cut(), EnvPairs(), escapePercent(), ContainerSecret (+15 more)

### Community 82 - "age"
Cohesion: 0.22
Nodes (9): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+1 more)

### Community 83 - "inspect_test.go"
Cohesion: 0.18
Nodes (17): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+9 more)

### Community 84 - "io.Writer"
Cohesion: 0.27
Nodes (8): io.Reader, io.Writer, TestPromptYes(), promptLine(), promptYes(), isTTY(), TestIsTTYIsFalseForANonTerminal(), Consenter

### Community 85 - "statusreport_test.go"
Cohesion: 0.12
Nodes (23): bytes.Buffer, Manager, rootlessNoFileMgr(), TestPrepHostRootlessNoFileSufficient(), TestPrepHostRootlessNoFileTooLow(), TestPrepHostRootlessNoFileUnlimited(), TestPrepHostRootlessNoFileUnreadable(), TestPrepHostRootlessNoFileUnsetSkips() (+15 more)

### Community 86 - "inspect.go"
Cohesion: 0.29
Nodes (13): brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), brokerList, brokerStatus, condition, deploymentList (+5 more)

### Community 88 - "runP2"
Cohesion: 0.15
Nodes (16): containsPlatform(), platformNames(), resolvePlatform(), roleLabel(), TestP2Classify(), TestResolvePlatform(), writeEnv(), Env (+8 more)

## Knowledge Gaps
- **165 isolated node(s):** `solace`, `showCmd`, `Config`, `operatorTmplVars`, `NO_COLOR` (+160 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **6 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `captureStdout`, `config_test.go`, `Manager`, `Config`, `load.go`, `NewTransport`, `convert_test.go`, `manager_test.go`, `.Run`, `github.com/spf13/cobra.Command`, `CheckCommand`, `tierFor`, `Compose`, `semp_test.go`, `validate.go`, `manager.go`, `Probe`, `execguard_test.go`, `eqArgs`, `containerTransport`, `runP2`, `Ops`?**
  _High betweenness centrality (0.079) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `bg`, `context.Context`, `captureStdout`, `Manager`, `newTestOps`, `NewTransport`, `NewCluster`, `convert_test.go`, `manager_test.go`, `secrets_test.go`, `prep_test.go`, `Role`, `.Run`, `strings.Builder`, `recRunner`, `Compose`, `Cluster`, `k8s/runtime_test.go`, `semp_test.go`, `Cluster`, `captureStderr`, `Probe`, `operator.go`, `eqArgs`, `containerTransport`, `Image`, `render.go`, `Ops`?**
  _High betweenness centrality (0.058) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `containerTransport`, `.AdditionalUsers`, `Manager`, `Cluster`, `newTestOps`, `semp_test.go`, `Cluster`, `runP2`, `scripts.go`, `Ops`, `github.com/spf13/cobra.Command`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Are the 10 inferred relationships involving `ctrCfg()` (e.g. with `TestComposeSecretEnvIsTheOnlyChildEnvironment()` and `TestComposeSecretEnvNamesCannotBeSystemVars()`) actually correct?**
  _`ctrCfg()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **Are the 44 inferred relationships involving `newTestOps()` (e.g. with `TestAdditionalUsersRunCLITransportError()` and `TestDiagnosticsMkdirError()`) actually correct?**
  _`newTestOps()` has 44 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `newCapMgr()` (e.g. with `NewManager()` and `TestComposeSecretEnvIsTheOnlyChildEnvironment()`) actually correct?**
  _`newCapMgr()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `showCmd`, `Config` to the rest of the system?**
  _165 weakly-connected nodes found - possible documentation gaps or missing edges._
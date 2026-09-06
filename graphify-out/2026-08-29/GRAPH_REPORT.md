# Graph Report - solace-k8-scripts  (2026-08-23)

## Corpus Check
- 121 files · ~248,065 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2154 nodes · 6646 edges · 93 communities (84 shown, 9 thin omitted)
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
- testing.T
- Commands
- Manager
- Config
- load.go
- newTestOps
- NewTransport
- NewCluster
- eqArgs
- Platform
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
- cli/platform.go
- github.com/spf13/cobra.Command
- render.go
- solace
- Command
- Command reference
- .Preflight
- completion_test.go
- .Run
- .resolveSecretRefs
- haCfg
- tierFor
- render_test.go
- .gatherNode
- k8s/runtime_test.go
- semp_test.go
- command_test.go
- Cluster
- captureStderr
- coverage_test.go
- runner_test.go
- resolveEnv
- manager.go
- Probe
- newRootCmd
- podmanEnv
- completion.go
- kubectlTransport
- Test catalogue
- scripts.go
- internal/cli
- internal/broker
- internal/config
- internal/container
- operator.go
- .MateRevertActivity
- internal/engine
- commanddoc_test.go
- runP1
- containerTransport
- allowcommand_test.go
- .releaseLocalActivity
- fakeRunner
- .AdditionalUsers
- Live-environment probe harness (`solace-itest`)
- runP4
- .releaseToBackup
- lastLines
- Cluster
- TestShowRedundancyIsReadOnly
- output_test.go
- quadletEscape
- internal/tools/itest
- EnvPairs
- age
- inspect_test.go
- io.Writer
- statusreport_test.go
- inspect.go
- Compose
- itest_test.go
- applyAliases
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
- `TestNewDefaults()` --calls--> `New()`  [INFERRED]
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

## Communities (93 total, 9 thin omitted)

### Community 0 - "bg"
Cohesion: 0.06
Nodes (99): TestCtrManagerConfirmWiring(), TestWarnAndStep(), lineSink(), progress(), step(), warn(), App, newConvertCmd() (+91 more)

### Community 1 - "cli_test.go"
Cohesion: 0.14
Nodes (32): firstLine(), runCtr(), runRoot(), TestBashEnvGivenToEnvFlag(), TestConfigStepsDoNotLeakSecrets(), TestConvertErrorPaths(), TestConvertParseError(), TestConvertRoundTrip() (+24 more)

### Community 2 - "Sink"
Cohesion: 0.12
Nodes (11): TestCheckReportSkipsEmptySections(), Level, Sink, New(), NewFunc(), pad(), TestLevelTagsMapToTheHouseTags(), TestNewFuncRoutesThroughTheGivenLineSink() (+3 more)

### Community 3 - "context.Context"
Cohesion: 0.12
Nodes (10): Echo, Exec, context.Context, os/exec.Cmd, Quote(), Cluster, storageClassSuitable(), Cluster (+2 more)

### Community 4 - "captureStdout"
Cohesion: 0.14
Nodes (19): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch(), opCanI() (+11 more)

### Community 5 - "testing.T"
Cohesion: 0.06
Nodes (81): testing.T, assertContainerBlockDefaults(), assertContainerScaling(), envTree(), Config, haNodesConfig(), minimalK8s(), TestApplyDefaultsDocker() (+73 more)

### Community 6 - "Commands"
Cohesion: 0.03
Nodes (67): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util check (+59 more)

### Community 7 - "Manager"
Cohesion: 0.13
Nodes (3): os.FileMode, NodeIdentity, Manager

### Community 8 - "Config"
Cohesion: 0.11
Nodes (25): Admin, Broker, Container, ContainerSecurity, DockerConfig, DomainCerts, Image, Network (+17 more)

### Community 9 - "load.go"
Cohesion: 0.26
Nodes (11): TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config, setDefault() (+3 more)

### Community 10 - "newTestOps"
Cohesion: 0.11
Nodes (39): appUsers(), Ops, newTestOps(), TestAdditionalUsers(), TestAdditionalUsersEmpty(), TestAdditionalUsersRejectsBadValues(), TestAdditionalUsersReportsExistingUser(), TestDiagnostics() (+31 more)

### Community 11 - "NewTransport"
Cohesion: 0.14
Nodes (23): capCall, capRunner, New(), Transport, Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrTransportHonoursRuntime() (+15 more)

### Community 12 - "NewCluster"
Cohesion: 0.11
Nodes (39): TestCheckStopsProbingWhenUnreachable(), TestCheckStorageClass(), TestResolveStorageClass(), TestCheckDeployGroupsAndOrdersSections(), TestCheckDeployNeverPrintsASecret(), TestCheckDeployReportsEveryFailureInOneRun(), TestCheckDeploySparseConfigExplainsItself(), TestCheckDeployTagsEchoedConfigAsINFO() (+31 more)

### Community 13 - "eqArgs"
Cohesion: 0.16
Nodes (25): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSDefaultOnError(), TestOperatorNSDefaultWhenAbsent(), TestOperatorNSDerived() (+17 more)

### Community 14 - "Platform"
Cohesion: 0.05
Nodes (68): keyValueEntries, doc, Result, vars, Platform, foldToEnvVar(), Config, missingErr() (+60 more)

### Community 15 - "internal/k8s"
Cohesion: 0.17
Nodes (12): check_test.go, cluster_test.go, deploy_test.go, internal/k8s, names_test.go, operator_test.go, ops_test.go, preflight_test.go (+4 more)

### Community 16 - "manager_test.go"
Cohesion: 0.07
Nodes (97): fileExists(), assertMode(), ctrCfg(), Manager, hasCall(), newCapMgr(), newEchoMgr(), rootlessNoFileMgr() (+89 more)

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
Cohesion: 0.15
Nodes (25): ProductKeyRoles(), saCfg(), TestHARoles(), TestProductKeyRoles(), adminCfg(), Cluster, labelCluster(), TestCreateNamespaceApplyFails() (+17 more)

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
Cohesion: 0.08
Nodes (22): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+14 more)

### Community 26 - "cli/platform.go"
Cohesion: 0.16
Nodes (25): checkFlagPlatforms(), commandPlatforms(), declaredList(), App, parsePlatformList(), platformSuffix(), prepare(), promptPlatform() (+17 more)

### Community 27 - "github.com/spf13/cobra.Command"
Cohesion: 0.23
Nodes (39): opFunc, github.com/spf13/cobra.Command, TestFirstArg(), addCommands(), deployBroker(), App, group(), newCheckCmd() (+31 more)

### Community 28 - "render.go"
Cohesion: 0.12
Nodes (30): WeightedNodeTerm, strings.Builder, NodeAffinity, NodeMatchExpr, Placement, PodAffinityTerm, validateMatchExprs(), validatePlacementAffinity() (+22 more)

### Community 31 - "Command"
Cohesion: 0.09
Nodes (33): commandRules, Command, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules() (+25 more)

### Community 32 - "Command reference"
Cohesion: 0.50
Nodes (3): Command reference, Global flags, Tree

### Community 34 - "completion_test.go"
Cohesion: 0.24
Nodes (16): runComplete(), TestAllowCommandOffersNoFiles(), TestCompletionHelpStillWorks(), TestCompletionNeedsAShell(), TestCompletionNoDescriptions(), TestCompletionScriptsGenerate(), TestDirFlagCompletesDirectories(), TestEnvFlagCompletesEnvFiles() (+8 more)

### Community 35 - ".Run"
Cohesion: 0.19
Nodes (21): echoRunner(), App, runRootWith(), TestCLICommand(), TestCtrConfirmDeclined(), TestErrorPaths(), TestFirstArgOr(), TestGenerateWired() (+13 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "haCfg"
Cohesion: 0.15
Nodes (14): haCfg(), TestOperatorImage(), TestCanIAnswerReadsTheLastLine(), TestCreateSecretsFailsWithoutAdminFields(), TestDeleteSecretsSkipsUnconfiguredAdminSecret(), NewTransport(), isCanI(), TestTransportCopy() (+6 more)

### Community 38 - "tierFor"
Cohesion: 0.27
Nodes (7): scalingTier, containerMem(), Config, TestContainerMem(), TestScalingTiers(), TestTierForRejectsOffTierValues(), tierFor()

### Community 39 - "render_test.go"
Cohesion: 0.22
Nodes (18): opCtrGenSecrets(), ContainerSecrets(), SecretPreflight(), SecretScript(), shQuote(), envLines(), healthCheckFixture(), load() (+10 more)

### Community 40 - ".gatherNode"
Cohesion: 0.33
Nodes (4): gatherConfigsScript(), TestGatherConfigsScript(), TestZipConfigsScript(), zipConfigsScript()

### Community 41 - "k8s/runtime_test.go"
Cohesion: 0.43
Nodes (7): TestClusterHonoursRuntime(), TestExecutorRefusesUnapprovedRuntime(), TestRuntimeDefaultArgvUnchanged(), TestTransportHonoursRuntime(), unapprovedCfg(), withLeading(), wrappedCfg()

### Community 42 - "semp_test.go"
Cohesion: 0.15
Nodes (19): sempPort(), TestCurlConfigLineEscapesQuotes(), TestHTTPBodyStripsHeaders(), TestLastHTTPStatus(), TestMateRevertActivityBridgePort(), TestMateRevertActivityCredsAndBodyNeverInArgv(), TestMateRevertActivityNon2xx(), TestMateRevertActivityRPCNotOK() (+11 more)

### Community 43 - "command_test.go"
Cohesion: 0.16
Nodes (13): decodeRuntime(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRuntimeDefaults() (+5 more)

### Community 44 - "Cluster"
Cohesion: 0.11
Nodes (12): bufio.Reader, Cluster, isBuiltinLabel(), joinManifests(), namespaceManifest(), roleName(), rolePlacementLabels(), splitLabel() (+4 more)

### Community 45 - "captureStderr"
Cohesion: 0.16
Nodes (18): allowRuntime(), capture(), captureStderr(), fakeBinaryOnPath(), runStandalone(), runStatusStderr(), TestAnnounceCommandsNamesResolvedBinaries(), TestBinaryAnnouncementWiring() (+10 more)

### Community 46 - "coverage_test.go"
Cohesion: 0.06
Nodes (49): matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultVPNDisableError() (+41 more)

### Community 47 - "runner_test.go"
Cohesion: 0.08
Nodes (38): App, NewManager(), maskedKeys(), TestManagerNilSinks(), TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand() (+30 more)

### Community 48 - "resolveEnv"
Cohesion: 0.13
Nodes (20): containsPlatform(), Env, platformNames(), resolveEnv(), resolvePlatform(), roleLabel(), TestResolvePlatform(), writeEnv() (+12 more)

### Community 49 - "manager.go"
Cohesion: 0.18
Nodes (10): defaultGenPSK(), exactName(), orNone(), platformTitle(), replacePSKLine(), secretSummary(), setOrMissing(), solaceRows() (+2 more)

### Community 50 - "Probe"
Cohesion: 0.20
Nodes (17): TestRunProbesSkipsOffPlatformAndExitsOnRefuted(), probeP1(), p2Classify(), probeP2(), probeP4(), Env, inconclusive(), listProbes() (+9 more)

### Community 51 - "newRootCmd"
Cohesion: 0.11
Nodes (23): TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestDangerousVerbsHaveNoBareAlias(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownNoun(), TestStartStopHaveNoAlias(), TestAllowCommandIsRegisteredWhereItExecutes(), TestAllowCommandIsRepeatable() (+15 more)

### Community 52 - "podmanEnv"
Cohesion: 0.16
Nodes (22): Env, podmanEnv(), TestCtrNameReadsTheConfiguredContainer(), TestP1DeclinedConsentTouchesNothing(), TestP1DryRunTouchesNothing(), TestP1RestoresAfterAFailureMidSequence(), TestP1RestoreWarnsLoudlyWhenItCannotRestore(), TestP1RunsStopRotateStartReadInOrder() (+14 more)

### Community 53 - "completion.go"
Cohesion: 0.29
Nodes (11): github.com/spf13/cobra.ShellCompDirective, completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), completionShell(), App, isEnvFileName() (+3 more)

### Community 55 - "Test catalogue"
Cohesion: 0.12
Nodes (16): convert_test.go, Coverage, Fixtures and doubles, Injectable seams, internal/convert, internal/output, internal/render, internal/tools/vulnjudge (+8 more)

### Community 56 - "scripts.go"
Cohesion: 0.11
Nodes (26): showCmd, assertLeaderScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript(), noReleaseActivityScript(), parseVPNNames(), productKeysScript() (+18 more)

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

### Community 61 - "operator.go"
Cohesion: 0.08
Nodes (33): orValue(), setOrMissing(), setOrNone(), failRow(), Cluster, info(), okRow(), skipRow() (+25 more)

### Community 62 - ".MateRevertActivity"
Cohesion: 0.20
Nodes (12): TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigLine(), Ops, httpBody(), lastHTTPStatus(), revertActivityMateBody(), sempV1OK() (+4 more)

### Community 63 - "internal/engine"
Cohesion: 0.67
Nodes (3): internal/engine, resolve_test.go, runner_test.go

### Community 65 - "commanddoc_test.go"
Cohesion: 0.38
Nodes (9): github.com/spf13/pflag.FlagSet, availableSubs(), firstDiff(), mdCell(), renderCommandDocs(), TestCommandDocs(), walkCommands(), writeCommand() (+1 more)

### Community 66 - "runP1"
Cohesion: 0.33
Nodes (10): ctrName(), Env, Env, Result, p1Classify(), p1ReadMountedSecret(), p1Restore(), p1Throwaway() (+2 more)

### Community 68 - "allowcommand_test.go"
Cohesion: 0.33
Nodes (9): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused(), TestSmuggledSubcommandIsRefused() (+1 more)

### Community 69 - ".releaseLocalActivity"
Cohesion: 0.21
Nodes (11): field(), TestField(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), Ops, hostMatches(), roleName(), shortHost() (+3 more)

### Community 71 - ".AdditionalUsers"
Cohesion: 0.10
Nodes (21): containsAnyFold(), countContains(), TestContainsAnyFold(), TestCountContains(), TestPathHelpers(), TestServerCert(), TestValidName(), validCLILine() (+13 more)

### Community 72 - "Live-environment probe harness (`solace-itest`)"
Cohesion: 0.17
Nodes (10): After a run, Build and run, Flags, Live-environment probe harness (`solace-itest`), Operator-run only, Safety, Testing the harness itself, The probes (+2 more)

### Community 73 - "runP4"
Cohesion: 0.21
Nodes (14): TestCPUAndMemoryConversion(), TestP4Classify(), TestP4Parse(), TestRunP4InconclusiveWhenInspectFails(), TestRunP4ReadsInspectAndChangesNothing(), TestRunP4SkipsWithNoLimitsConfigured(), cpuToNano(), Env (+6 more)

### Community 75 - "lastLines"
Cohesion: 0.67
Nodes (3): TestLastLines(), TestLastLinesEqualCount(), lastLines()

### Community 77 - "TestShowRedundancyIsReadOnly"
Cohesion: 0.67
Nodes (3): TestMateActivityStateReadsTheMateColumn(), TestShowRedundancyIsReadOnly(), MateActivityState()

### Community 78 - "output_test.go"
Cohesion: 0.26
Nodes (13): sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLineAddsNoPrefix(), TestSectionPadsToWidth(), TestStepKeepsTheArrowPrefix() (+5 more)

### Community 79 - "quadletEscape"
Cohesion: 0.67
Nodes (3): escapePercent(), quadletEscape(), TestQuadletEscape()

### Community 80 - "internal/tools/itest"
Cohesion: 0.50
Nodes (4): internal/tools/itest, itest_test.go, live_test.go, probes_live_test.go

### Community 81 - "EnvPairs"
Cohesion: 0.17
Nodes (8): containerSecretSpecs(), EnvPairs(), ContainerSecret, groupKey(), itoa(), secretFilePath(), EnvPair, secretSpec

### Community 82 - "age"
Cohesion: 0.22
Nodes (9): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+1 more)

### Community 83 - "inspect_test.go"
Cohesion: 0.16
Nodes (18): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestConditionLevelDegradesSafely() (+10 more)

### Community 84 - "io.Writer"
Cohesion: 0.12
Nodes (29): layer, roleOpFunc, io.Reader, io.Writer, os.File, TestConfirmFlagShortcuts(), TestConfirmNonTTY(), TestIsTTYClosedFile() (+21 more)

### Community 85 - "statusreport_test.go"
Cohesion: 0.22
Nodes (15): Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList(), TestBrokerReportWithNoBrokerDeployed() (+7 more)

### Community 86 - "inspect.go"
Cohesion: 0.19
Nodes (16): time.Time, brokerConditionRows(), conditionLevel(), findCondition(), Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes(), brokerList (+8 more)

### Community 87 - "Compose"
Cohesion: 0.29
Nodes (11): containsStr(), TestManagerRedeployChangedNeedsConsent(), TestManagerRedeployUnchangedHintsRotation(), TestManagerRedeployUnchangedIsNoOp(), Compose(), healthCmd(), Quadlet(), TestArtifactsCarryNoSecrets() (+3 more)

### Community 88 - "itest_test.go"
Cohesion: 0.14
Nodes (17): bytes.Buffer, Env, ids(), sinkOver(), TestConsent(), testEnv(), TestListProbesNeedsNoEnvFile(), TestP1Classify() (+9 more)

## Knowledge Gaps
- **137 isolated node(s):** `solace`, `showCmd`, `Config`, `operatorTmplVars`, `NO_COLOR` (+132 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **9 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `captureStdout`, `testing.T`, `Manager`, `Config`, `load.go`, `NewTransport`, `manager_test.go`, `cli/platform.go`, `github.com/spf13/cobra.Command`, `Command`, `tierFor`, `render_test.go`, `semp_test.go`, `runner_test.go`, `resolveEnv`, `manager.go`, `Probe`, `containerTransport`, `Ops`?**
  _High betweenness centrality (0.072) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `bg`, `captureStdout`, `Manager`, `newTestOps`, `NewTransport`, `NewCluster`, `Platform`, `manager_test.go`, `secrets_test.go`, `prep_test.go`, `Role`, `render.go`, `haCfg`, `render_test.go`, `k8s/runtime_test.go`, `semp_test.go`, `Cluster`, `captureStderr`, `runner_test.go`, `Probe`, `kubectlTransport`, `operator.go`, `containerTransport`, `Cluster`, `EnvPairs`, `Compose`, `Ops`?**
  _High betweenness centrality (0.070) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `containerTransport`, `.releaseLocalActivity`, `.AdditionalUsers`, `.gatherNode`, `Manager`, `.releaseToBackup`, `Cluster`, `resolveEnv`, `prep_test.go`, `kubectlTransport`, `scripts.go`, `Ops`, `github.com/spf13/cobra.Command`, `.MateRevertActivity`?**
  _High betweenness centrality (0.033) - this node is a cross-community bridge._
- **Are the 10 inferred relationships involving `ctrCfg()` (e.g. with `TestComposeSecretEnvIsTheOnlyChildEnvironment()` and `TestComposeSecretEnvNamesCannotBeSystemVars()`) actually correct?**
  _`ctrCfg()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **Are the 44 inferred relationships involving `newTestOps()` (e.g. with `TestAdditionalUsersRunCLITransportError()` and `TestDiagnosticsMkdirError()`) actually correct?**
  _`newTestOps()` has 44 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `newCapMgr()` (e.g. with `NewManager()` and `TestComposeSecretEnvIsTheOnlyChildEnvironment()`) actually correct?**
  _`newCapMgr()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `showCmd`, `Config` to the rest of the system?**
  _137 weakly-connected nodes found - possible documentation gaps or missing edges._
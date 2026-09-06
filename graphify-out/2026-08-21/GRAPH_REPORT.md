# Graph Report - solace-k8-scripts  (2026-08-21)

## Corpus Check
- 121 files · ~245,656 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2144 nodes · 6616 edges · 93 communities (85 shown, 8 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 1060 edges (avg confidence: 0.8)
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
- newEchoMgr
- Platform
- render.go
- solace
- Command
- Command reference
- .Preflight
- completion_test.go
- runRootWith
- .resolveSecretRefs
- haCfg
- kubectlTransport
- Compose
- scripts_test.go
- k8s/runtime_test.go
- semp_test.go
- command_test.go
- Cluster
- hasCall
- testing.T
- io.Writer
- run
- manager.go
- runP2
- newRootCmd
- podmanEnv
- Cluster
- .Run
- Test catalogue
- scripts.go
- internal/cli
- internal/broker
- internal/config
- internal/container
- .configRows
- .MateRevertActivity
- internal/engine
- RenderOperator
- runP1
- containerTransport
- .gatherNode
- .LeaderLocal
- NewManager
- .AdditionalUsers
- Live-environment probe harness (`solace-itest`)
- runP4
- .releaseToBackup
- resolveEnv
- fakeRunner
- Image
- New
- .ServerCert
- internal/tools/itest
- EnvPairs
- age
- inspect_test.go
- container/preflight_test.go
- statusreport_test.go
- inspect.go
- itest_test.go
- Ops
- TestCheckStorageClass
- time.Time

## God Nodes (most connected - your core abstractions)
1. `ctrCfg()` - 91 edges
2. `newTestOps()` - 83 edges
3. `newCapMgr()` - 82 edges
4. `Role` - 81 edges
5. `Config` - 79 edges
6. `Platform` - 75 edges
7. `bg()` - 68 edges
8. `Commands` - 65 edges
9. `Manager` - 61 edges
10. `NewCluster()` - 58 edges

## Surprising Connections (you probably didn't know these)
- `TestNewDefaults()` --calls--> `New()`  [INFERRED]
  internal/broker/coverage_test.go → internal/broker/broker.go
- `TestField()` --calls--> `field()`  [INFERRED]
  internal/broker/broker_test.go → internal/broker/broker.go
- `TestFieldLabelWithoutColon()` --calls--> `field()`  [INFERRED]
  internal/broker/coverage_test.go → internal/broker/broker.go
- `primaryRedundancyUp()` --calls--> `field()`  [INFERRED]
  internal/broker/verify_ops.go → internal/broker/broker.go
- `TestCountContains()` --calls--> `countContains()`  [INFERRED]
  internal/broker/broker_test.go → internal/broker/broker.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (93 total, 8 thin omitted)

### Community 0 - "bg"
Cohesion: 0.06
Nodes (99): TestCtrManagerConfirmWiring(), TestWarnAndStep(), App, lineSink(), progress(), step(), warn(), App (+91 more)

### Community 1 - "cli_test.go"
Cohesion: 0.11
Nodes (41): allowRuntime(), capture(), captureStderr(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone(), runStatusStderr() (+33 more)

### Community 2 - "Sink"
Cohesion: 0.16
Nodes (6): Level, Sink, pad(), TestLevelTagsMapToTheHouseTags(), KV, Row

### Community 3 - "context.Context"
Cohesion: 0.10
Nodes (7): context.Context, Cluster, Cluster, Cluster, Cluster, badJSONRunner, fixtureRunner

### Community 4 - "captureStdout"
Cohesion: 0.14
Nodes (19): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch(), opCanI() (+11 more)

### Community 5 - "config_test.go"
Cohesion: 0.05
Nodes (78): assertContainerBlockDefaults(), assertContainerScaling(), envTree(), Config, haNodesConfig(), minimalK8s(), TestApplyDefaultsDocker(), TestApplyDefaultsK8s() (+70 more)

### Community 6 - "Commands"
Cohesion: 0.03
Nodes (65): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util check (+57 more)

### Community 7 - "Manager"
Cohesion: 0.13
Nodes (3): os.FileMode, Manager, solaceRows()

### Community 8 - "Config"
Cohesion: 0.11
Nodes (27): Admin, Broker, Container, ContainerSecurity, DockerConfig, DomainCerts, Network, Node (+19 more)

### Community 9 - "load.go"
Cohesion: 0.12
Nodes (19): scalingTier, TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config (+11 more)

### Community 10 - "newTestOps"
Cohesion: 0.08
Nodes (52): containsAnyFold(), appUsers(), Ops, newTestOps(), ranContains(), TestAdditionalUsers(), TestAdditionalUsersEmpty(), TestAdditionalUsersRejectsBadValues() (+44 more)

### Community 11 - "NewTransport"
Cohesion: 0.20
Nodes (14): capCall, capRunner, New(), Transport, NewTransport(), dockerCfg(), eqArgs(), podmanCfg() (+6 more)

### Community 12 - "NewCluster"
Cohesion: 0.12
Nodes (36): TestCheckDeployGroupsAndOrdersSections(), TestCheckDeployNeverPrintsASecret(), TestCheckDeployReportsEveryFailureInOneRun(), TestCheckDeploySparseConfigExplainsItself(), TestCheckDeployTagsEchoedConfigAsINFO(), NewCluster(), logBuf(), TestDeleteBrokerLogsPVCOutcome() (+28 more)

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
Cohesion: 0.12
Nodes (54): ctrCfg(), newCapMgr(), TestContainerRunningMatchesNameExactly(), TestManagerCheckDNSFailsLoudInHA(), TestManagerCheckPodmanEUID(), TestManagerCheckReachableError(), TestManagerCheckStandaloneDNSWarnsOnly(), TestManagerCopy() (+46 more)

### Community 17 - "cli/platform_test.go"
Cohesion: 0.35
Nodes (13): App, runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection() (+5 more)

### Community 18 - "secrets_test.go"
Cohesion: 0.17
Nodes (18): TestUpdateServerCertSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), operatorRegcred(), checkGolden(), decodeDataValue(), TestAdminSecretDecodes() (+10 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (22): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+14 more)

### Community 20 - "prep_test.go"
Cohesion: 0.15
Nodes (25): ProductKeyRoles(), saCfg(), TestHARoles(), TestProductKeyRoles(), adminCfg(), Cluster, labelCluster(), TestCreateSecretsAdminOnly() (+17 more)

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
Cohesion: 0.14
Nodes (12): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+4 more)

### Community 26 - "newEchoMgr"
Cohesion: 0.11
Nodes (20): bytes.Buffer, Manager, newEchoMgr(), rootlessNoFileMgr(), setNoFile(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployPodmanDryRunHidesSecretBytes() (+12 more)

### Community 27 - "Platform"
Cohesion: 0.06
Nodes (106): layer, opFunc, roleOpFunc, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, github.com/spf13/pflag.FlagSet, io.Reader, TestConfirmFlagShortcuts() (+98 more)

### Community 28 - "render.go"
Cohesion: 0.19
Nodes (22): strings.Builder, PodAffinityTerm, Cluster, boolStr(), BrokerCR(), cut(), parsePort(), parseToleration() (+14 more)

### Community 31 - "Command"
Cohesion: 0.06
Nodes (45): commandRules, keyValueEntries, TestCommandArgsDoesNotAliasCommand(), Command, checkBinary(), CheckCommand(), checkFlagShape(), checkToken() (+37 more)

### Community 32 - "Command reference"
Cohesion: 0.50
Nodes (3): Command reference, Global flags, Tree

### Community 34 - "completion_test.go"
Cohesion: 0.37
Nodes (12): runComplete(), TestAllowCommandOffersNoFiles(), TestDirFlagCompletesDirectories(), TestEnvFlagCompletesEnvFiles(), TestEnvFlagPrefixFilters(), TestEnvFlagWithPathDefersToShell(), TestNoArgsLeafOffersNoFiles(), TestPlatformFlagCompletes() (+4 more)

### Community 35 - "runRootWith"
Cohesion: 0.14
Nodes (25): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+17 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "haCfg"
Cohesion: 0.16
Nodes (12): haCfg(), TestCreateSecretsFailsWithoutAdminFields(), TestDeleteSecretsSkipsUnconfiguredAdminSecret(), NewTransport(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs() (+4 more)

### Community 39 - "Compose"
Cohesion: 0.13
Nodes (31): opCtrGenSecrets(), Compose(), ContainerSecrets(), escapePercent(), ContainerSecret, healthCmd(), Quadlet(), quadletEscape() (+23 more)

### Community 40 - "scripts_test.go"
Cohesion: 0.17
Nodes (12): assertLeaderScript(), disableDefaultUsersScript(), domainCertsScript(), parseVPNNames(), showVPNBareScript(), sortedKeys(), TestAssertLeaderScript(), TestDisableDefaultUsersScriptQuoting() (+4 more)

### Community 41 - "k8s/runtime_test.go"
Cohesion: 0.43
Nodes (7): TestClusterHonoursRuntime(), TestExecutorRefusesUnapprovedRuntime(), TestRuntimeDefaultArgvUnchanged(), TestTransportHonoursRuntime(), unapprovedCfg(), withLeading(), wrappedCfg()

### Community 42 - "semp_test.go"
Cohesion: 0.13
Nodes (21): sempPort(), TestCurlConfigLineEscapesQuotes(), TestHTTPBodyStripsHeaders(), TestLastHTTPStatus(), TestMateRevertActivityBridgePort(), TestMateRevertActivityCredsAndBodyNeverInArgv(), TestMateRevertActivityNon2xx(), TestMateRevertActivityRPCNotOK() (+13 more)

### Community 43 - "command_test.go"
Cohesion: 0.18
Nodes (12): decodeRuntime(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRuntimeDefaults(), TestRuntimeExplicitValueSurvivesDefaults() (+4 more)

### Community 44 - "Cluster"
Cohesion: 0.11
Nodes (13): bufio.Reader, GenSecrets(), Cluster, isBuiltinLabel(), joinManifests(), namespaceManifest(), roleName(), rolePlacementLabels() (+5 more)

### Community 45 - "hasCall"
Cohesion: 0.12
Nodes (24): fileExists(), assertMode(), containsStr(), hasCall(), maskedKeys(), TestManagerDeleteDockerComposeDownWhenFileExists(), TestManagerDeletePodmanRemovesUnit(), TestManagerDeployDockerComposeWritesFile() (+16 more)

### Community 46 - "testing.T"
Cohesion: 0.09
Nodes (54): testing.T, matchCLI(), TestLastLines(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError() (+46 more)

### Community 47 - "io.Writer"
Cohesion: 0.06
Nodes (42): Echo, Exec, io.Writer, os/exec.Cmd, os.File, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose() (+34 more)

### Community 48 - "run"
Cohesion: 0.13
Nodes (18): TestListProbesNeedsNoEnvFile(), writeEnv(), Env, main(), parseFlags(), run(), runProbes(), splitProbe() (+10 more)

### Community 49 - "manager.go"
Cohesion: 0.14
Nodes (12): defaultGenPSK(), exactName(), orNone(), platformTitle(), replacePSKLine(), secretSummary(), setOrMissing(), splitLimit() (+4 more)

### Community 50 - "runP2"
Cohesion: 0.17
Nodes (21): TestRegistryCoversEveryClaim(), TestRunProbesSkipsOffPlatformAndExitsOnRefuted(), probeP1(), Env, Result, p2Classify(), probeP2(), runP2() (+13 more)

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (25): applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestDangerousVerbsHaveNoBareAlias(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownNoun(), TestStartStopHaveNoAlias(), TestAllowCommandIsRegisteredWhereItExecutes() (+17 more)

### Community 52 - "podmanEnv"
Cohesion: 0.22
Nodes (17): Env, podmanEnv(), TestCtrNameReadsTheConfiguredContainer(), TestP1DeclinedConsentTouchesNothing(), TestP1DryRunTouchesNothing(), TestP1RestoresAfterAFailureMidSequence(), TestP1RestoreWarnsLoudlyWhenItCannotRestore(), TestP1RunsStopRotateStartReadInOrder() (+9 more)

### Community 53 - "Cluster"
Cohesion: 0.17
Nodes (9): HARoles(), lbServiceName(), podName(), pvcName(), RestartOrder(), stsName(), TestResourceNames(), TestRestartOrder() (+1 more)

### Community 54 - ".Run"
Cohesion: 0.17
Nodes (18): runCtr(), TestConfigStepsDoNotLeakSecrets(), TestCtrConfigDryRun(), TestCtrDiagnosticsDryRun(), TestCtrErrorPaths(), TestCtrExecCLIPathSeparator(), TestCtrLoginOutcomes(), TestCtrRoleArgCount() (+10 more)

### Community 55 - "Test catalogue"
Cohesion: 0.12
Nodes (16): convert_test.go, Coverage, Fixtures and doubles, Injectable seams, internal/convert, internal/output, internal/render, internal/tools/vulnjudge (+8 more)

### Community 56 - "scripts.go"
Cohesion: 0.18
Nodes (14): showCmd, field(), disableDefaultVPNScript(), noReleaseActivityScript(), releaseActivityScript(), revertActivityConfigureScript(), revertActivityScript(), showRedundancyDetailScript() (+6 more)

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

### Community 62 - ".MateRevertActivity"
Cohesion: 0.21
Nodes (11): anyHTTP2xx(), curlConfigLine(), Ops, httpBody(), lastHTTPStatus(), revertActivityMateBody(), sempV1OK(), TestMateActivityStateReadsTheMateColumn() (+3 more)

### Community 63 - "internal/engine"
Cohesion: 0.67
Nodes (3): internal/engine, resolve_test.go, runner_test.go

### Community 65 - "RenderOperator"
Cohesion: 0.27
Nodes (9): GenOperator(), joinYAMLDocs(), operatorImage(), RenderOperator(), splitOperatorBundle(), TestOperatorImage(), TestSplitOperatorBundle(), watchNamespace() (+1 more)

### Community 66 - "runP1"
Cohesion: 0.33
Nodes (10): ctrName(), Env, Env, Result, p1Classify(), p1ReadMountedSecret(), p1Restore(), p1Throwaway() (+2 more)

### Community 68 - ".gatherNode"
Cohesion: 0.33
Nodes (4): gatherConfigsScript(), TestGatherConfigsScript(), TestZipConfigsScript(), zipConfigsScript()

### Community 69 - ".LeaderLocal"
Cohesion: 0.31
Nodes (6): Ops, hostMatches(), roleName(), shortHost(), TestRoleName(), primaryRedundancyUp()

### Community 70 - "NewManager"
Cohesion: 0.26
Nodes (12): NewManager(), TestManagerNilSinks(), Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime() (+4 more)

### Community 71 - ".AdditionalUsers"
Cohesion: 0.18
Nodes (10): countContains(), validCLILine(), validCLIPassword(), Ops, additionalUsersScript(), productKeysScript(), removeDomainCertsScript(), TestAdditionalUsersScript() (+2 more)

### Community 72 - "Live-environment probe harness (`solace-itest`)"
Cohesion: 0.17
Nodes (10): After a run, Build and run, Flags, Live-environment probe harness (`solace-itest`), Operator-run only, Safety, Testing the harness itself, The probes (+2 more)

### Community 73 - "runP4"
Cohesion: 0.38
Nodes (9): TestCPUAndMemoryConversion(), cpuToNano(), Env, Result, memToBytes(), p4Classify(), p4Parse(), runP4() (+1 more)

### Community 75 - "resolveEnv"
Cohesion: 0.31
Nodes (8): containsPlatform(), Env, platformNames(), resolveEnv(), resolvePlatform(), roleLabel(), TestResolvePlatform(), TestResolveEnvWiresThePlatformsEntryType()

### Community 78 - "New"
Cohesion: 0.15
Nodes (18): TestCheckReportSkipsEmptySections(), New(), NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow() (+10 more)

### Community 79 - ".ServerCert"
Cohesion: 0.47
Nodes (4): concatFiles(), serverCertFile(), serverCertScript(), TestServerCertScript()

### Community 80 - "internal/tools/itest"
Cohesion: 0.50
Nodes (4): internal/tools/itest, itest_test.go, live_test.go, probes_live_test.go

### Community 81 - "EnvPairs"
Cohesion: 0.17
Nodes (9): Config, NodeIdentity, containerSecretSpecs(), EnvPairs(), groupKey(), itoa(), secretFilePath(), EnvPair (+1 more)

### Community 82 - "age"
Cohesion: 0.22
Nodes (9): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+1 more)

### Community 83 - "inspect_test.go"
Cohesion: 0.18
Nodes (17): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+9 more)

### Community 84 - "container/preflight_test.go"
Cohesion: 0.40
Nodes (4): TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestComposeSecretEnvNamesCannotBeSystemVars(), TestPreflightIsPreviewableUnderDryRun(), TestPreflightRunsBeforeAnything()

### Community 85 - "statusreport_test.go"
Cohesion: 0.22
Nodes (15): Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList(), TestBrokerReportWithNoBrokerDeployed() (+7 more)

### Community 86 - "inspect.go"
Cohesion: 0.29
Nodes (13): brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), brokerList, brokerStatus, condition, deploymentList (+5 more)

### Community 88 - "itest_test.go"
Cohesion: 0.15
Nodes (16): Env, ids(), sinkOver(), TestConsent(), testEnv(), TestP1Classify(), TestP1ThrowawayIsUsable(), TestP2Classify() (+8 more)

### Community 92 - "TestCheckStorageClass"
Cohesion: 0.40
Nodes (4): TestCheckStopsProbingWhenUnreachable(), TestCheckStorageClass(), TestReachable(), TestResolveStorageClass()

### Community 93 - "time.Time"
Cohesion: 0.33
Nodes (4): time.Time, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

## Knowledge Gaps
- **136 isolated node(s):** `solace`, `showCmd`, `Config`, `operatorTmplVars`, `Cluster` (+131 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `bg`, `captureStdout`, `config_test.go`, `Manager`, `Config`, `load.go`, `NewTransport`, `convert_test.go`, `manager_test.go`, `newEchoMgr`, `Command`, `Compose`, `semp_test.go`, `manager.go`, `runP2`, `containerTransport`, `NewManager`, `resolveEnv`, `Ops`?**
  _High betweenness centrality (0.083) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `bg`, `cli_test.go`, `context.Context`, `captureStdout`, `Manager`, `newTestOps`, `NewTransport`, `NewCluster`, `convert_test.go`, `manager_test.go`, `secrets_test.go`, `prep_test.go`, `newEchoMgr`, `render.go`, `haCfg`, `kubectlTransport`, `Compose`, `k8s/runtime_test.go`, `semp_test.go`, `Cluster`, `hasCall`, `runP2`, `Cluster`, `RenderOperator`, `containerTransport`, `NewManager`, `Image`, `EnvPairs`, `Ops`?**
  _High betweenness centrality (0.072) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `Manager`, `newTestOps`, `prep_test.go`, `Platform`, `kubectlTransport`, `scripts_test.go`, `Cluster`, `Cluster`, `scripts.go`, `.MateRevertActivity`, `containerTransport`, `.gatherNode`, `.LeaderLocal`, `.AdditionalUsers`, `.releaseToBackup`, `resolveEnv`, `EnvPairs`, `Ops`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Are the 10 inferred relationships involving `ctrCfg()` (e.g. with `TestComposeSecretEnvIsTheOnlyChildEnvironment()` and `TestComposeSecretEnvNamesCannotBeSystemVars()`) actually correct?**
  _`ctrCfg()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **Are the 44 inferred relationships involving `newTestOps()` (e.g. with `TestAdditionalUsersRunCLITransportError()` and `TestDiagnosticsMkdirError()`) actually correct?**
  _`newTestOps()` has 44 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `newCapMgr()` (e.g. with `NewManager()` and `TestComposeSecretEnvIsTheOnlyChildEnvironment()`) actually correct?**
  _`newCapMgr()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `showCmd`, `Config` to the rest of the system?**
  _136 weakly-connected nodes found - possible documentation gaps or missing edges._
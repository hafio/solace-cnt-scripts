# Graph Report - solace-k8-scripts  (2026-08-21)

## Corpus Check
- 121 files · ~246,388 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2146 nodes · 6621 edges · 96 communities (91 shown, 5 thin omitted)
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
- config.go
- load.go
- newTestOps
- NewTransport
- eqArgs
- newCluster
- convert_test.go
- internal/k8s
- manager_test.go
- cli/platform_test.go
- Config
- dev.sh
- prep_test.go
- dev.ps1
- CLAUDE.md
- judge
- verify_local_test.go
- Role
- newEchoMgr
- Platform
- strings.Builder
- solace
- Command
- Command reference
- .Preflight
- completion_test.go
- runRootWith
- .resolveSecretRefs
- haCfg
- scaling_test.go
- Compose
- scripts_test.go
- k8s/runtime_test.go
- semp_test.go
- command_test.go
- NewTransport
- hasCall
- testing.T
- runner_test.go
- io.Writer
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
- RenderOperator
- .MateRevertActivity
- internal/engine
- .Run
- runP1
- containerTransport
- Load
- .releaseLocalActivity
- NewManager
- .AdditionalUsers
- Live-environment probe harness (`solace-itest`)
- runP4
- .releaseToBackup
- resolvePlatform
- Cluster
- Ops
- output_test.go
- .ServerCert
- internal/tools/itest
- render.go
- age
- inspect_test.go
- Consenter
- statusreport_test.go
- inspect.go
- TestManagerRedeployUnchangedIsNoOp
- itest_test.go
- ResolveEnvPath
- Ops
- assertContainerBlockDefaults
- NewCluster
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
- `activity()` --calls--> `countContains()`  [INFERRED]
  internal/broker/verify_ops.go → internal/broker/broker.go
- `matchCLI()` --calls--> `cliArg()`  [INFERRED]
  internal/broker/broker_test.go → internal/broker/transport.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (96 total, 5 thin omitted)

### Community 0 - "bg"
Cohesion: 0.06
Nodes (100): TestCtrManagerConfirmWiring(), TestWarnAndStep(), App, lineSink(), progress(), step(), warn(), App (+92 more)

### Community 1 - "cli_test.go"
Cohesion: 0.11
Nodes (41): allowRuntime(), capture(), captureStderr(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone(), runStatusStderr() (+33 more)

### Community 2 - "Sink"
Cohesion: 0.16
Nodes (6): Level, Sink, pad(), TestLevelTagsMapToTheHouseTags(), KV, Row

### Community 3 - "context.Context"
Cohesion: 0.13
Nodes (6): context.Context, Cluster, Cluster, Cluster, badJSONRunner, fixtureRunner

### Community 4 - "captureStdout"
Cohesion: 0.14
Nodes (19): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch(), opCanI() (+11 more)

### Community 5 - "config_test.go"
Cohesion: 0.08
Nodes (41): Config, haNodesConfig(), TestApplyDefaultsK8s(), TestApplyDefaultsK8sTLS(), TestApplyDefaultsPodmanRootlessHomeDir(), TestApplyDefaultsPodmanRootlessXDG(), TestContainerBlock(), TestContainerRuntime() (+33 more)

### Community 6 - "Commands"
Cohesion: 0.03
Nodes (65): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util check (+57 more)

### Community 7 - "Manager"
Cohesion: 0.13
Nodes (3): os.FileMode, Manager, solaceRows()

### Community 8 - "config.go"
Cohesion: 0.09
Nodes (26): Admin, Broker, Container, ContainerSecurity, DockerConfig, DomainCerts, Image, Network (+18 more)

### Community 9 - "load.go"
Cohesion: 0.20
Nodes (12): TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config, parseError() (+4 more)

### Community 10 - "newTestOps"
Cohesion: 0.10
Nodes (44): appUsers(), Ops, newTestOps(), ranContains(), TestAdditionalUsers(), TestAdditionalUsersEmpty(), TestAdditionalUsersRejectsBadValues(), TestAdditionalUsersReportsExistingUser() (+36 more)

### Community 11 - "NewTransport"
Cohesion: 0.19
Nodes (15): capCall, capRunner, New(), Transport, TestManagerLogsCLIShell(), NewTransport(), dockerCfg(), eqArgs() (+7 more)

### Community 12 - "eqArgs"
Cohesion: 0.15
Nodes (28): TestCheckDeploySparseConfigExplainsItself(), logBuf(), TestDeleteBrokerLogsPVCOutcome(), TestDeleteBrokerNoPurge(), TestDeleteBrokerPurgeHA(), TestDeleteBrokerPurgeStandalone(), TestDeleteBrokerPurgeSwallowsPVCError(), TestDeployBrokerApply() (+20 more)

### Community 13 - "newCluster"
Cohesion: 0.13
Nodes (23): Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSDefaultOnError(), TestOperatorNSDefaultWhenAbsent(), TestOperatorNSDerived(), TestOperatorNSExplicit() (+15 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (58): doc, Result, vars, boolOf(), commentSafe(), Convert(), countMarkers(), emitYAML() (+50 more)

### Community 15 - "internal/k8s"
Cohesion: 0.17
Nodes (12): check_test.go, cluster_test.go, deploy_test.go, internal/k8s, names_test.go, operator_test.go, ops_test.go, preflight_test.go (+4 more)

### Community 16 - "manager_test.go"
Cohesion: 0.12
Nodes (49): ctrCfg(), newCapMgr(), TestContainerRunningMatchesNameExactly(), TestManagerCheckDNSFailsLoudInHA(), TestManagerCheckPodmanEUID(), TestManagerCheckReachableError(), TestManagerCheckStandaloneDNSWarnsOnly(), TestManagerCopy() (+41 more)

### Community 17 - "cli/platform_test.go"
Cohesion: 0.35
Nodes (13): App, runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection() (+5 more)

### Community 18 - "Config"
Cohesion: 0.15
Nodes (21): tlsConfigured(), Config, GenSecrets(), joinManifests(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), operatorRegcred() (+13 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (22): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+14 more)

### Community 20 - "prep_test.go"
Cohesion: 0.26
Nodes (16): saCfg(), Cluster, labelCluster(), TestCreateNamespace(), TestDeleteNamespace(), TestLabelNodesBuiltinOnly(), TestLabelNodesEOFNoSelection(), TestLabelNodesHAOnlyPrimaryConfigured() (+8 more)

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

### Community 26 - "newEchoMgr"
Cohesion: 0.11
Nodes (20): bytes.Buffer, Manager, newEchoMgr(), rootlessNoFileMgr(), setNoFile(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployPodmanDryRunHidesSecretBytes() (+12 more)

### Community 27 - "Platform"
Cohesion: 0.06
Nodes (106): layer, opFunc, roleOpFunc, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, github.com/spf13/pflag.FlagSet, io.Reader, TestConfirmFlagShortcuts() (+98 more)

### Community 28 - "strings.Builder"
Cohesion: 0.24
Nodes (14): strings.Builder, PodAffinityTerm, parseToleration(), sortedKeys(), TestParseToleration(), writeKeyValueEntry(), writeLBAnnotations(), writeMatchExprs() (+6 more)

### Community 31 - "Command"
Cohesion: 0.06
Nodes (44): commandRules, keyValueEntries, Command, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules() (+36 more)

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
Cohesion: 0.12
Nodes (17): haCfg(), TestHARoles(), TestProductKeyRoles(), TestResourceNames(), TestRestartOrder(), TestCanIAnswerReadsTheLastLine(), TestCreateSecretsFailsWithoutAdminFields(), TestDeleteSecretsSkipsUnconfiguredAdminSecret() (+9 more)

### Community 38 - "scaling_test.go"
Cohesion: 0.11
Nodes (19): scalingTier, containerMem(), Config, Config, setContainerMem(), TestApplyScalingTierDefaultsContainerBlocks(), TestApplyScalingTierDefaultsK8s(), TestApplyScalingTierDefaultsMemOverride() (+11 more)

### Community 39 - "Compose"
Cohesion: 0.19
Nodes (24): HealthCheck, NodeIdentity, BrokerCR(), Compose(), EnvPairs(), healthCmd(), Quadlet(), envLines() (+16 more)

### Community 40 - "scripts_test.go"
Cohesion: 0.16
Nodes (11): assertLeaderScript(), disableDefaultUsersScript(), gatherConfigsScript(), parseVPNNames(), TestAssertLeaderScript(), TestDisableDefaultUsersScriptQuoting(), TestGatherConfigsScript(), TestParseVPNNames() (+3 more)

### Community 41 - "k8s/runtime_test.go"
Cohesion: 0.43
Nodes (7): TestClusterHonoursRuntime(), TestExecutorRefusesUnapprovedRuntime(), TestRuntimeDefaultArgvUnchanged(), TestTransportHonoursRuntime(), unapprovedCfg(), withLeading(), wrappedCfg()

### Community 42 - "semp_test.go"
Cohesion: 0.13
Nodes (21): sempPort(), TestCurlConfigLineEscapesQuotes(), TestHTTPBodyStripsHeaders(), TestLastHTTPStatus(), TestMateRevertActivityBridgePort(), TestMateRevertActivityCredsAndBodyNeverInArgv(), TestMateRevertActivityNon2xx(), TestMateRevertActivityRPCNotOK() (+13 more)

### Community 43 - "command_test.go"
Cohesion: 0.16
Nodes (13): decodeRuntime(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRuntimeDefaults() (+5 more)

### Community 44 - "NewTransport"
Cohesion: 0.11
Nodes (13): bufio.Reader, Cluster, isBuiltinLabel(), namespaceManifest(), roleName(), rolePlacementLabels(), splitLabel(), TestIsBuiltinLabel() (+5 more)

### Community 45 - "hasCall"
Cohesion: 0.10
Nodes (25): fileExists(), hasCall(), TestManagerDeleteDockerComposeDownWhenFileExists(), TestManagerDeleteDockerComposeNoFileFallsBackToStopRm(), TestManagerDeleteDockerPurgeRemovesDataDir(), TestManagerDeletePodmanRemovesUnit(), TestManagerDeployDockerComposeWritesFile(), TestManagerDeployDockerDryRunMasksSecretEnv() (+17 more)

### Community 46 - "testing.T"
Cohesion: 0.09
Nodes (54): testing.T, matchCLI(), TestLastLines(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError() (+46 more)

### Community 47 - "runner_test.go"
Cohesion: 0.11
Nodes (28): TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveRefusesCurrentDirectory(), verboseExec(), captureStdout(), helperCommand() (+20 more)

### Community 48 - "io.Writer"
Cohesion: 0.12
Nodes (24): io.Writer, NewExec(), New(), Env, resolveEnv(), TestResolvePlatform(), writeEnv(), TestResolveEnvWiresThePlatformsEntryType() (+16 more)

### Community 49 - "manager.go"
Cohesion: 0.14
Nodes (12): defaultGenPSK(), exactName(), orNone(), platformTitle(), replacePSKLine(), secretSummary(), setOrMissing(), splitLimit() (+4 more)

### Community 50 - "runP2"
Cohesion: 0.16
Nodes (23): ids(), TestRegistryCoversEveryClaim(), TestRunProbesSkipsOffPlatformAndExitsOnRefuted(), TestSelectProbes(), probeP1(), Env, Result, p2Classify() (+15 more)

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (25): applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestDangerousVerbsHaveNoBareAlias(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownNoun(), TestStartStopHaveNoAlias(), TestAllowCommandIsRegisteredWhereItExecutes() (+17 more)

### Community 52 - "podmanEnv"
Cohesion: 0.16
Nodes (18): Env, podmanEnv(), TestCtrNameReadsTheConfiguredContainer(), TestP1DeclinedConsentTouchesNothing(), TestP1DryRunTouchesNothing(), TestP1RestoresAfterAFailureMidSequence(), TestP1RestoreWarnsLoudlyWhenItCannotRestore(), TestP1RunsStopRotateStartReadInOrder() (+10 more)

### Community 53 - "Cluster"
Cohesion: 0.14
Nodes (9): Cluster, HARoles(), lbServiceName(), podName(), ProductKeyRoles(), pvcName(), RestartOrder(), stsName() (+1 more)

### Community 54 - ".Run"
Cohesion: 0.17
Nodes (18): runCtr(), TestConfigStepsDoNotLeakSecrets(), TestCtrConfigDryRun(), TestCtrDiagnosticsDryRun(), TestCtrErrorPaths(), TestCtrExecCLIPathSeparator(), TestCtrLoginOutcomes(), TestCtrRoleArgCount() (+10 more)

### Community 55 - "Test catalogue"
Cohesion: 0.12
Nodes (16): convert_test.go, Coverage, Fixtures and doubles, Injectable seams, internal/convert, internal/output, internal/render, internal/tools/vulnjudge (+8 more)

### Community 56 - "scripts.go"
Cohesion: 0.27
Nodes (11): showCmd, noReleaseActivityScript(), releaseActivityScript(), revertActivityConfigureScript(), revertActivityScript(), showRedundancyDetailScript(), showRedundancyScript(), showVPNBareScript() (+3 more)

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

### Community 61 - "RenderOperator"
Cohesion: 0.09
Nodes (32): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), failRow(), Cluster, info(), okRow() (+24 more)

### Community 62 - ".MateRevertActivity"
Cohesion: 0.27
Nodes (9): anyHTTP2xx(), curlConfigLine(), Ops, httpBody(), lastHTTPStatus(), revertActivityMateBody(), sempV1OK(), httpStatusLines() (+1 more)

### Community 63 - "internal/engine"
Cohesion: 0.67
Nodes (3): internal/engine, resolve_test.go, runner_test.go

### Community 65 - ".Run"
Cohesion: 0.20
Nodes (8): Echo, Exec, os/exec.Cmd, TestResolveMissingBinaryIsActionable(), MaskEnv(), Quote(), quoteTok(), Resolve()

### Community 66 - "runP1"
Cohesion: 0.33
Nodes (10): ctrName(), Env, Env, Result, p1Classify(), p1ReadMountedSecret(), p1Restore(), p1Throwaway() (+2 more)

### Community 68 - "Load"
Cohesion: 0.23
Nodes (16): minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadReadError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors() (+8 more)

### Community 69 - ".releaseLocalActivity"
Cohesion: 0.26
Nodes (8): field(), Ops, hostMatches(), roleName(), shortHost(), TestRoleName(), primaryRedundancyUp(), rdEnabledUp()

### Community 70 - "NewManager"
Cohesion: 0.26
Nodes (12): NewManager(), TestManagerNilSinks(), Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime() (+4 more)

### Community 71 - ".AdditionalUsers"
Cohesion: 0.13
Nodes (16): containsAnyFold(), countContains(), TestContainsAnyFold(), TestCountContains(), TestPathHelpers(), TestValidName(), validCLILine(), validCLIPassword() (+8 more)

### Community 72 - "Live-environment probe harness (`solace-itest`)"
Cohesion: 0.17
Nodes (10): After a run, Build and run, Flags, Live-environment probe harness (`solace-itest`), Operator-run only, Safety, Testing the harness itself, The probes (+2 more)

### Community 73 - "runP4"
Cohesion: 0.38
Nodes (9): TestCPUAndMemoryConversion(), cpuToNano(), Env, Result, memToBytes(), p4Classify(), p4Parse(), runP4() (+1 more)

### Community 74 - ".releaseToBackup"
Cohesion: 0.38
Nodes (4): TestMateActivityStateReadsTheMateColumn(), activity(), Ops, MateActivityState()

### Community 75 - "resolvePlatform"
Cohesion: 0.60
Nodes (4): containsPlatform(), platformNames(), resolvePlatform(), roleLabel()

### Community 77 - "Ops"
Cohesion: 0.20
Nodes (8): Ops, disableDefaultVPNScript(), domainCertsScript(), removeDomainCertsScript(), sortedKeys(), TestDisableDefaultVPNScript(), TestDomainCertsScriptSorted(), TestSortedKeys()

### Community 78 - "output_test.go"
Cohesion: 0.21
Nodes (16): NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLineAddsNoPrefix(), TestNewFuncRoutesThroughTheGivenLineSink() (+8 more)

### Community 79 - ".ServerCert"
Cohesion: 0.47
Nodes (4): concatFiles(), serverCertFile(), serverCertScript(), TestServerCertScript()

### Community 80 - "internal/tools/itest"
Cohesion: 0.50
Nodes (4): internal/tools/itest, itest_test.go, live_test.go, probes_live_test.go

### Community 81 - "render.go"
Cohesion: 0.12
Nodes (22): boolStr(), ContainerSecrets(), containerSecretSpecs(), cut(), escapePercent(), ContainerSecret, groupKey(), itoa() (+14 more)

### Community 82 - "age"
Cohesion: 0.22
Nodes (9): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+1 more)

### Community 83 - "inspect_test.go"
Cohesion: 0.18
Nodes (17): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+9 more)

### Community 84 - "Consenter"
Cohesion: 0.33
Nodes (4): os.File, isTTY(), TestIsTTYIsFalseForANonTerminal(), Consenter

### Community 85 - "statusreport_test.go"
Cohesion: 0.22
Nodes (15): Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList(), TestBrokerReportWithNoBrokerDeployed() (+7 more)

### Community 86 - "inspect.go"
Cohesion: 0.29
Nodes (13): brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), brokerList, brokerStatus, condition, deploymentList (+5 more)

### Community 87 - "TestManagerRedeployUnchangedIsNoOp"
Cohesion: 0.29
Nodes (7): assertMode(), containsStr(), maskedKeys(), TestManagerDeployDockerPassesSecretsAsEnv(), TestManagerDeployPodmanCreatesSecrets(), TestManagerRedeployUnchangedHintsRotation(), TestManagerRedeployUnchangedIsNoOp()

### Community 88 - "itest_test.go"
Cohesion: 0.15
Nodes (15): Env, sinkOver(), TestConsent(), testEnv(), TestListProbesNeedsNoEnvFile(), TestP1Classify(), TestP1ThrowawayIsUsable(), TestP2Classify() (+7 more)

### Community 89 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 91 - "assertContainerBlockDefaults"
Cohesion: 0.67
Nodes (4): assertContainerBlockDefaults(), assertContainerScaling(), TestApplyDefaultsDocker(), TestApplyDefaultsPodmanRootful()

### Community 92 - "NewCluster"
Cohesion: 0.13
Nodes (20): TestCheckStopsProbingWhenUnreachable(), TestCheckStorageClass(), TestReachable(), TestResolveStorageClass(), TestCheckDeployGroupsAndOrdersSections(), TestCheckDeployNeverPrintsASecret(), TestCheckDeployReportsEveryFailureInOneRun(), TestCheckDeployTagsEchoedConfigAsINFO() (+12 more)

### Community 93 - "time.Time"
Cohesion: 0.33
Nodes (4): time.Time, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

## Knowledge Gaps
- **135 isolated node(s):** `solace`, `showCmd`, `Config`, `operatorTmplVars`, `NO_COLOR` (+130 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **5 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `bg`, `captureStdout`, `config_test.go`, `Manager`, `config.go`, `load.go`, `NewTransport`, `convert_test.go`, `manager_test.go`, `newEchoMgr`, `Command`, `scaling_test.go`, `Compose`, `semp_test.go`, `manager.go`, `runP2`, `containerTransport`, `Load`, `NewManager`, `resolvePlatform`, `render.go`, `Ops`?**
  _High betweenness centrality (0.081) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `bg`, `cli_test.go`, `captureStdout`, `Manager`, `config.go`, `newTestOps`, `NewTransport`, `eqArgs`, `convert_test.go`, `manager_test.go`, `prep_test.go`, `Role`, `newEchoMgr`, `strings.Builder`, `haCfg`, `Compose`, `k8s/runtime_test.go`, `semp_test.go`, `NewTransport`, `hasCall`, `runP2`, `Cluster`, `RenderOperator`, `containerTransport`, `NewManager`, `Cluster`, `render.go`, `Ops`, `NewCluster`?**
  _High betweenness centrality (0.078) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `containerTransport`, `.releaseLocalActivity`, `.AdditionalUsers`, `scripts_test.go`, `Manager`, `.releaseToBackup`, `resolvePlatform`, `NewTransport`, `Ops`, `Cluster`, `Ops`, `Platform`, `.MateRevertActivity`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **Are the 10 inferred relationships involving `ctrCfg()` (e.g. with `TestComposeSecretEnvIsTheOnlyChildEnvironment()` and `TestComposeSecretEnvNamesCannotBeSystemVars()`) actually correct?**
  _`ctrCfg()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **Are the 44 inferred relationships involving `newTestOps()` (e.g. with `TestAdditionalUsersRunCLITransportError()` and `TestDiagnosticsMkdirError()`) actually correct?**
  _`newTestOps()` has 44 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `newCapMgr()` (e.g. with `NewManager()` and `TestComposeSecretEnvIsTheOnlyChildEnvironment()`) actually correct?**
  _`newCapMgr()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `showCmd`, `Config` to the rest of the system?**
  _135 weakly-connected nodes found - possible documentation gaps or missing edges._
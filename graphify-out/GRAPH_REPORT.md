# Graph Report - solace-cnt-scripts  (2026-11-05)

## Corpus Check
- 203 files · ~704,812 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .golden 17, (none) 3, .cli 2)

## Summary
- 4050 nodes · 20664 edges · 118 communities (112 shown, 6 thin omitted)
- Extraction: 72% EXTRACTED · 28% INFERRED · 0% AMBIGUOUS · INFERRED: 5715 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `fa288e51`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- App
- newTestOps
- diff_test.go
- Runner
- cli_test.go
- config_test.go
- Command Reference Docs
- Manager
- K8sConfig
- completion_test.go
- CheckCommand
- matechannel_test.go
- NewCluster
- Coverage
- convert_test.go
- transform_test.go
- manager_test.go
- commands.go
- coverage_test.go
- Bash Dev Script
- Load
- PowerShell Dev Script
- scaling_test.go
- switchplan_test.go
- newLocalOps
- runner_test.go
- renderDriver
- statusreport_test.go
- importops_test.go
- k8s/inspect.go
- Solace Module Root
- internal/config
- CLAUDE.md
- allowcommand_test.go
- ResolveDomainCerts
- watchLoop
- newRootCmd
- NewCluster
- annotate_test.go
- output_test.go
- Compose
- inject_test.go
- secrets_test.go
- exportconfig_test.go
- VPNRepl
- localCfg
- config_test.go
- eqArgs
- cli/platform_test.go
- Role
- prep_test.go
- .validate
- .dump
- context.Context
- replication_test.go
- TestGeneratorPageEmbedsTheCLI
- replCfg
- confirm.go
- checkRow
- rootless_test.go
- Shared env fixtures
- Platform
- Coverage
- replication_test.go
- cliTransport
- The eight rules
- EnvPairs
- What `import-config` applies
- MateConfig
- blocks_test.go
- Abbreviations
- internal/broker
- step
- .podSection
- .validate
- cli/watch_test.go
- .checkUserManagerLimits
- recRunner
- replicationops_test.go
- replCfg
- limits_test.go
- Developer guide
- .BackupRevertActivity
- inspect_test.go
- RenderMate
- AdminSecret
- render_test.go
- checkReport
- RuleFor
- namespace_test.go
- Container
- command_test.go
- .rpc
- replConfig
- Container
- containerTransport
- Configuration
- ImportPlan
- DetectPlatforms
- replication_test.go
- stsName
- README.md
- examples_test.go
- operatorversion_test.go
- runLeader
- adminsecret_test.go
- scripts.go
- Execute
- capRunner
- Container State Inspection
- bytes.Buffer
- hostpath_test.go
- Capture
- strings.Builder

## God Nodes (most connected - your core abstractions)
1. `App` - 217 edges
2. `newTestOps()` - 167 edges
3. `cli_test.go` - 148 edges
4. `Config` - 146 edges
5. `NewCluster()` - 141 edges
6. `ctrCfg()` - 133 edges
7. `Role` - 132 edges
8. `Per-package doubles` - 125 edges
9. `loadK8s()` - 119 edges
10. `newCapMgr()` - 114 edges

## Surprising Connections (you probably didn't know these)
- `The preflight` --references--> `info()`  [INFERRED]
  docs/operations.md → internal/k8s/checkreport.go
- `annotate_test.go` --references--> `Annotate()`  [INFERRED]
  docs/test.md → internal/broker/annotate.go
- `omit_test.go` --references--> `Annotate()`  [INFERRED]
  docs/test.md → internal/broker/annotate.go
- `annotate_test.go` --references--> `TestAnnotateRoundTripPreservesBlocks()`  [INFERRED]
  docs/test.md → internal/broker/annotate_test.go
- `annotate_test.go` --references--> `TestAnnotateRoundTripCaptureIsMarked()`  [INFERRED]
  docs/test.md → internal/broker/annotate_test.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (118 total, 6 thin omitted)

### Community 0 - "App"
Cohesion: 0.06
Nodes (106): App, logArgs(), confirmAction(), confirmDelete(), lineSink(), progress(), warn(), confirmImport() (+98 more)

### Community 1 - "newTestOps"
Cohesion: 0.08
Nodes (69): recOutput, broker_test.go, countContains(), leaderCalls(), newTestOps(), TestCountContains(), TestDiagnostics(), TestDisableDefaultUsers() (+61 more)

### Community 2 - "diff_test.go"
Cohesion: 0.11
Nodes (34): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, diff_test.go, ancestorReported(), blockLabel() (+26 more)

### Community 3 - "Runner"
Cohesion: 0.13
Nodes (10): Contract, Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams, EnvRunner, Runner (+2 more)

### Community 4 - "cli_test.go"
Cohesion: 0.07
Nodes (102): cli_test.go, allowRuntime(), captureStderr(), collectPaths(), echoRunner(), fakeBinaryOnPath(), firstLine(), runCtr() (+94 more)

### Community 5 - "config_test.go"
Cohesion: 0.04
Nodes (71): Image, config_test.go, redundancy_test.go, atoiPrefix(), haNodesConfig(), TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes(), TestAdditionalUserNamesCollideOnDocker() (+63 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (50): Commands, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker cli, solace-util broker configure (+42 more)

### Community 7 - "Manager"
Cohesion: 0.09
Nodes (18): preflight_test.go, runtime_test.go, certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle() (+10 more)

### Community 8 - "K8sConfig"
Cohesion: 0.11
Nodes (15): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, LoadBalancer, cut() (+7 more)

### Community 9 - "completion_test.go"
Cohesion: 0.22
Nodes (24): completion_test.go, driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion() (+16 more)

### Community 10 - "CheckCommand"
Cohesion: 0.09
Nodes (36): cmdField, commandRules, guardedCmd, execguard_test.go, decodeCommand(), TestCommandUnmarshal(), Command, checkBinary() (+28 more)

### Community 11 - "matechannel_test.go"
Cohesion: 0.15
Nodes (30): CLIRunner, fakeRun, matechannel_test.go, MateChannel, NewCLIMate(), mateReplies(), newTestMate(), TestLocalMateRunsThroughOps() (+22 more)

### Community 12 - "NewCluster"
Cohesion: 0.08
Nodes (59): operator_test.go, GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator(), splitAfterNamespace() (+51 more)

### Community 13 - "Coverage"
Cohesion: 0.15
Nodes (14): Coverage, internal/convert, internal/examples, internal/render, internal/tools/vulnjudge, Running the tests, Test catalogue, ValidScriptName() (+6 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (69): doc, Result, segment, vars, The keys, convert_test.go, DecodeStrict(), boolOf() (+61 more)

### Community 15 - "transform_test.go"
Cohesion: 0.10
Nodes (61): Block, TargetState, transform_test.go, BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects(), ClearTargetVirtualHostnames() (+53 more)

### Community 16 - "manager_test.go"
Cohesion: 0.09
Nodes (102): manager_test.go, GuardPodmanEUID(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountUnknownWhenSystemctlFails(), fileExists() (+94 more)

### Community 17 - "commands.go"
Cohesion: 0.14
Nodes (69): opFunc, roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd() (+61 more)

### Community 18 - "coverage_test.go"
Cohesion: 0.07
Nodes (74): commanddoc_test.go, coverage_test.go, storage_test.go, verify_ops_test.go, ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError() (+66 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "Load"
Cohesion: 0.16
Nodes (26): minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors(), TestLoadSecurityIDsAsWritten() (+18 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "scaling_test.go"
Cohesion: 0.07
Nodes (30): scalingTier, scaling_test.go, containerMem(), cpuSetCount(), cpuSetRange(), decodeScaling(), setContainerCPUSet(), setContainerMem() (+22 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.12
Nodes (45): PhaseKind, SiteState, SwitchAction, SwitchPhase, switchplan_test.go, BuildSwitchPlan(), confirmDemotions(), ExecuteSwitchPlan() (+37 more)

### Community 24 - "newLocalOps"
Cohesion: 0.15
Nodes (35): verify_local_test.go, cliScriptNameFromDest(), uploadedForRole(), cliScriptPath(), isCurl(), newLocalOps(), rd(), seqTransport() (+27 more)

### Community 25 - "runner_test.go"
Cohesion: 0.08
Nodes (42): interactiveFailRunner, Keys that were renamed, internal/engine, resolve_test.go, runner_test.go, Exec, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod() (+34 more)

### Community 26 - "renderDriver"
Cohesion: 0.15
Nodes (22): chunk, chunkResult, driver_test.go, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce() (+14 more)

### Community 27 - "statusreport_test.go"
Cohesion: 0.10
Nodes (25): statusreport_test.go, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportOmitsWorkloadsThatAreNotTheBrokers(), TestBrokerReportOrdersPodsOfTheSameRoleByName(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportReadsOnlyWhatTheEnvFileNames(), TestBrokerReportRunningPicture() (+17 more)

### Community 28 - "importops_test.go"
Cohesion: 0.08
Nodes (60): importops_test.go, TestRenderDiffResultCleanReportHasNoFailureRows(), TestRenderDiffResultDirtyReportNamesOffendingSectionAndLines(), describeScope(), planSections(), chunkNames(), driverAllOK(), driverChunkBody() (+52 more)

### Community 29 - "k8s/inspect.go"
Cohesion: 0.13
Nodes (21): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), operatorRunningImage(), ownedPods(), TestOperatorRunningImageWithNoContainers() (+13 more)

### Community 31 - "internal/config"
Cohesion: 0.11
Nodes (21): artifactvalues_test.go, duration_test.go, imagepullsecret_test.go, internal/config, portOrRange(), TestValidateContainerArtifactValues(), TestValidateImageReference(), TestValidateK8sArtifactValues() (+13 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.25
Nodes (5): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh

### Community 33 - "allowcommand_test.go"
Cohesion: 0.35
Nodes (12): allowcommand_test.go, TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRegisteredWhereItExecutes(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes() (+4 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (54): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domaincerts_resolve_test.go, domaincerts_test.go, caNameSafe() (+46 more)

### Community 35 - "watchLoop"
Cohesion: 0.32
Nodes (5): watchInterval(), stdoutIsTerminal(), watchLabel(), watchLoop(), withWatch()

### Community 36 - "newRootCmd"
Cohesion: 0.09
Nodes (30): How a short form is resolved, Goldens, abbrevdoc_test.go, aliases_test.go, TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), applyAliases(), TestAliasesDoNotCollide() (+22 more)

### Community 37 - "NewCluster"
Cohesion: 0.09
Nodes (68): checkreport_test.go, deploy_test.go, preflight_test.go, TestCheckStopsProbingWhenUnreachable(), TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestOperatorInstallVerdict(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestOperatorValidateAsksNothingAboutTheBroker() (+60 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.07
Nodes (48): Region, annotate_test.go, applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV() (+40 more)

### Community 40 - "output_test.go"
Cohesion: 0.13
Nodes (24): internal/output, output_test.go, columnWidths(), Level, Row, Sink, NewFunc(), pad() (+16 more)

### Community 41 - "Compose"
Cohesion: 0.17
Nodes (18): servercert_test.go, Compose(), composeEscape(), ComposeProject(), ContainerSecrets(), ContainerSecret, SecretPreflight(), ServerCertBundlePath() (+10 more)

### Community 42 - "inject_test.go"
Cohesion: 0.17
Nodes (28): lineRole, serviceLine, shutdownStyle, svcKey, inject_test.go, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "secrets_test.go"
Cohesion: 0.18
Nodes (25): secrets_test.go, ResolveSecretValues(), bundleHash(), certCreate(), certFixture(), TestCertSecretLabelDrivesTheRestart(), TestContainerBundleCarriesTheChain(), TestDeletePodmanToleratesAMissingBundle() (+17 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.11
Nodes (34): exportconfigReadCounter, opRunner, euid_test.go, exportconfig_test.go, internal/cli, runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand() (+26 more)

### Community 47 - "VPNRepl"
Cohesion: 0.10
Nodes (10): AdminState, QueueState, scriptedMate, replMate, ReplRole, VPNRepl, reenableLine(), setReplicationRoleScript() (+2 more)

### Community 48 - "localCfg"
Cohesion: 0.08
Nodes (54): semp_test.go, sempmate_test.go, bridgeHostPort(), sempPort(), sempV1OK(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA() (+46 more)

### Community 49 - "config_test.go"
Cohesion: 0.17
Nodes (18): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, showCmd, span, AdditionalUser, Node (+10 more)

### Community 50 - "eqArgs"
Cohesion: 0.09
Nodes (48): check_test.go, cluster_test.go, internal/k8s, ops_test.go, runtime_test.go, transport_test.go, TestReachable(), TestResolveStorageClass() (+40 more)

### Community 51 - "cli/platform_test.go"
Cohesion: 0.19
Nodes (22): platform_test.go, findCmd(), TestEveryRunnableCommandIsWired(), TestFlagsRegistered(), runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects() (+14 more)

### Community 52 - "Role"
Cohesion: 0.05
Nodes (32): covUploadFailsTo, downloadErrTransport, fakeTransport, recDownload, recRun, recUpload, recUploadFile, runErrMatchTransport (+24 more)

### Community 53 - "prep_test.go"
Cohesion: 0.10
Nodes (53): adminsecret_test.go, prep_test.go, AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestReadAdminPasswordArgvAndDecoding() (+45 more)

### Community 54 - ".validate"
Cohesion: 0.10
Nodes (21): Bringing up a fresh cluster, Broker scope is a fixed classification, Every destructive command confirms, Exit codes, Exporting and importing configuration, Extra CLI users differ by platform, Handling the artifact, Kubernetes-specific hazards (+13 more)

### Community 55 - ".dump"
Cohesion: 0.11
Nodes (60): opCall, ops_k8s_test.go, captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook(), loadDirect() (+52 more)

### Community 56 - "context.Context"
Cohesion: 0.05
Nodes (25): BaseName(), TestBaseNameSplitsOnBothSeparators(), idMapCovers(), origin(), runUserIDs(), normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes(), TestPodNameForIsTheOneSpelling() (+17 more)

### Community 57 - "replication_test.go"
Cohesion: 0.13
Nodes (27): replRig, Commands, solace-util broker, Configuring a site, Data replication, Permissions, Reaching the mate, Switching roles (+19 more)

### Community 58 - "TestGeneratorPageEmbedsTheCLI"
Cohesion: 0.12
Nodes (17): scalingKey, scalingSpelling, generatorpage_test.go, Scaling, scalingKeyIndex(), scalingKeyList(), generatorDefaults(), gunzipBase64() (+9 more)

### Community 59 - "replCfg"
Cohesion: 0.21
Nodes (26): matechannel_test.go, CLIArg(), CLIScriptPath(), execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf() (+18 more)

### Community 60 - "confirm.go"
Cohesion: 0.10
Nodes (26): layer, Injectable seams, hostMatches(), shortHost(), TestConfirmDowngradeNeverUnattended(), TestConfirmLayerMatrix(), TestConfirmNonTTY(), TestPromptAsksWhenStdinIsNotATTY() (+18 more)

### Community 61 - "checkRow"
Cohesion: 0.21
Nodes (16): orNone(), orValue(), setOrMissing(), setOrNone(), additionalUsersRow(), adminPassState(), adminSecretRow(), containsString() (+8 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.19
Nodes (40): rootless_test.go, failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+32 more)

### Community 63 - "Shared env fixtures"
Cohesion: 0.17
Nodes (12): runFailRunner, exit_test.go, Fixtures and doubles, Shared env fixtures, Summary, childStatusError(), k8sEnv(), TestChildExitStatusIsScopedToInteractiveSessions() (+4 more)

### Community 64 - "Platform"
Cohesion: 0.14
Nodes (29): platform_test.go, checkAllowCommand(), checkFlagPlatforms(), commandPlatforms(), declaredList(), parsePlatformList(), platformSuffix(), prepare() (+21 more)

### Community 65 - "Coverage"
Cohesion: 0.17
Nodes (15): ContainerNoFile(), HealthCheck, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), escapePercent(), healthCmd(), Quadlet(), quadletEscape() (+7 more)

### Community 66 - "replication_test.go"
Cohesion: 0.13
Nodes (31): mateChannelFunc, replication_test.go, ReplicationConfigResult, confirmReplicationSwitch(), mateChannel(), mateSEMPPassword(), opPerformReplication(), replVPNNames() (+23 more)

### Community 67 - "cliTransport"
Cohesion: 0.25
Nodes (8): cliTransport(), renderMateAppliance(), renderMateSoftware(), schemaTransport(), sortedKeys(), sortedSet(), TestTransportVocabularyIsThreeWords(), sortedVPNs()

### Community 68 - "The eight rules"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 69 - "EnvPairs"
Cohesion: 0.20
Nodes (11): containerSecretSpecs(), EnvPairs(), groupKey(), itoa(), secretFilePath(), assertNoCheckoutPath(), envLines(), TestAdditionalUsersReachBothHalves() (+3 more)

### Community 70 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 71 - "MateConfig"
Cohesion: 0.15
Nodes (17): containsEndpoint(), MateConfig, labelValue(), normTransport(), parseHostPort(), ParseShowReplication(), parseShowReplicationAppliance(), parseShowReplicationSoftware() (+9 more)

### Community 72 - "blocks_test.go"
Cohesion: 0.13
Nodes (26): token, blocks_test.go, Annotate(), TestAnnotateIsDeterministic(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), firstQuoted(), newBlock() (+18 more)

### Community 73 - "Abbreviations"
Cohesion: 0.20
Nodes (9): Abbreviations, Flag shorthands, Node roles, Platforms, Command reference, Global flags, Index, Reading this reference (+1 more)

### Community 74 - "internal/broker"
Cohesion: 0.13
Nodes (20): config_ops_test.go, internal/broker, omit_test.go, TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(), TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator(), TestExecShellScriptRefusesAnUnsafeNameBeforeUploading() (+12 more)

### Community 75 - "step"
Cohesion: 0.13
Nodes (16): step(), newConvertCmd(), runConvert(), newExamplesCmd(), runExample(), addOutFlags(), emitOrWrite(), opK8sGenBroker() (+8 more)

### Community 76 - ".podSection"
Cohesion: 0.19
Nodes (6): TestOperatorAdminSecretNameMatchesTheBrokerNames(), age(), discoveryServiceName(), lbServiceName(), ownsPod(), TestOwnsPodFallsBackToTheNameInfix()

### Community 77 - ".validate"
Cohesion: 0.12
Nodes (15): solace-util, solace-util operator, A replication switch refuses before changing anything, A wrapper runtime is refused, Docker and Podman mechanics, Docker compose secrets need compose 2.23.1+, Forbidden: cannot list deployments at the cluster scope, Import reported failure (+7 more)

### Community 78 - "cli/watch_test.go"
Cohesion: 0.35
Nodes (11): watch_test.go, stopAfter(), TestOperatorStatusWatches(), TestStatusWithoutWatchRunsOnce(), TestWatchAppendsWhenNotATerminal(), TestWatchEndsQuietlyWhenInterrupted(), TestWatchIntervalIsValidated(), TestWatchKeepsGoingAfterAFailedRefresh() (+3 more)

### Community 79 - ".checkUserManagerLimits"
Cohesion: 0.22
Nodes (6): delegateSetting(), limitText(), missingControllers(), pamCoreValue(), parseLimit(), parseUnitProps()

### Community 80 - "recRunner"
Cohesion: 0.19
Nodes (5): TestOperatorApplyPlansItsNamespace(), canIResource(), canIVerb(), recRunner, rrCall

### Community 81 - "replicationops_test.go"
Cohesion: 0.12
Nodes (47): replicationops_test.go, hasCall(), matchCLI(), outputForRole(), TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), mateConvergenceShutdowns() (+39 more)

### Community 82 - "replCfg"
Cohesion: 0.05
Nodes (27): keyValueEntries, ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, TestRedundancyEnabled(), viaDescription(), checkCredentialChars() (+19 more)

### Community 83 - "limits_test.go"
Cohesion: 0.09
Nodes (51): limits_test.go, main_test.go, limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessCoreFloorFollowsTheConfiguredLimit(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips() (+43 more)

### Community 84 - "Developer guide"
Cohesion: 0.29
Nodes (7): Build, Dev script tasks, Developer guide, Gates, Releases, Tests, Toolchain pin

### Community 85 - ".BackupRevertActivity"
Cohesion: 0.17
Nodes (11): backupTarget, Credential, anyHTTP2xx(), curlConfigFlag(), curlConfigLine(), lastHTTPStatus(), revertActivityBackupBody(), TestCurlConfigLineEscapesQuotes() (+3 more)

### Community 86 - "inspect_test.go"
Cohesion: 0.15
Nodes (22): inspect_test.go, envTree(), TestResolveEnvPath(), TestResolveEnvPathEmptyBaseDir(), podHealth(), pvcLevel(), replicaLevel(), roleRank() (+14 more)

### Community 87 - "RenderMate"
Cohesion: 0.15
Nodes (21): RenderMate(), RenderMateRemovals(), loadShowReplication(), mateAppliance(), mateSoftware(), TestMateRoundTripsThroughItsOwnGrammar(), TestParseShowReplicationAppliance(), TestRealApplianceCaptureMateRoundTrips() (+13 more)

### Community 88 - "AdminSecret"
Cohesion: 0.07
Nodes (44): TLS, The env-file generator page, Per-package doubles, secrets_test.go, tlsSecretDigest(), replShowMate(), replVPNTable(), validPlatformConfig() (+36 more)

### Community 89 - "render_test.go"
Cohesion: 0.16
Nodes (39): render_test.go, boolStr(), BrokerCR(), containerArtifacts(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets(), TestArtifactsCarryNoWideningTokens() (+31 more)

### Community 90 - "checkReport"
Cohesion: 0.23
Nodes (9): storageClassSuitable(), failRow(), okRow(), portRows(), skipRow(), TestCheckReportFailedCounts(), TestPortRowsNeverFail(), checkReport (+1 more)

### Community 91 - "RuleFor"
Cohesion: 0.11
Nodes (36): Disposition, SectionRule, appliance_test.go, importdoc_test.go, sections_test.go, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent() (+28 more)

### Community 92 - "namespace_test.go"
Cohesion: 0.27
Nodes (11): namespace_test.go, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 93 - "Container"
Cohesion: 0.18
Nodes (14): tlssecret_test.go, applyContainerBlockDefaults(), podmanRunUser(), setDefault(), setDefaultCmd(), setDefaultInt(), xdgConfigHome(), TestNamingTheTLSSecretDoesNotInventCertPaths() (+6 more)

### Community 95 - "command_test.go"
Cohesion: 0.16
Nodes (13): command_test.go, TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRenamedRuntimeKeysFailLoud(), TestRuntimeDefaults() (+5 more)

### Community 96 - ".rpc"
Cohesion: 0.27
Nodes (4): sempMate, httpBody(), TestHTTPBodyStripsHeaders(), xmlEscape()

### Community 97 - "replConfig"
Cohesion: 0.19
Nodes (15): Replication, secretRef, replication_test.go, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded() (+7 more)

### Community 98 - "Container"
Cohesion: 0.19
Nodes (10): Container, DockerConfig, Network, PodmanConfig, Ulimits, assertContainerBlockDefaults(), assertContainerScaling(), TestApplyDefaultsDocker() (+2 more)

### Community 100 - "Configuration"
Cohesion: 0.17
Nodes (12): Bring your own TLS Secret, Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling, Scaling tiers (+4 more)

### Community 101 - "ImportPlan"
Cohesion: 0.14
Nodes (19): ImportResult, PlannedSection, defaultVPNFirst(), ImportPlan, importIgnore(), orNone(), preambleForApply(), quoteList() (+11 more)

### Community 102 - "DetectPlatforms"
Cohesion: 0.22
Nodes (18): examples_test.go, Example, TestExamplesEmitsToStdout(), All(), Get(), List(), Names(), read() (+10 more)

### Community 103 - "replication_test.go"
Cohesion: 0.18
Nodes (20): colSpan, replication_test.go, dashSpans(), flagByte(), gutterClear(), ParseVPNReplication(), sliceSpan(), TestGutterClearAndSliceSpanBoundaries() (+12 more)

### Community 104 - "stsName"
Cohesion: 0.20
Nodes (11): names_test.go, BrokerPodSuffixShape(), ProductKeyRoles(), pvcName(), RestartOrder(), stsName(), TestHARoles(), TestPodNameSuffixMatchesTheConfigBound() (+3 more)

### Community 105 - "README.md"
Cohesion: 0.20
Nodes (6): internal/broker, internal/cli, internal/config, internal/engine, internal/k8s, internal/render

### Community 106 - "examples_test.go"
Cohesion: 0.19
Nodes (8): examples_test.go, capture(), TestExamplesAcceptsAPlatformAbbreviation(), TestExamplesBareEmitsTheFullSchema(), TestExamplesNeedsNoEnvFile(), TestExamplesRefusesToOverwrite(), TestExamplesRejectsAnUnknownPlatform(), TestExamplesWritesOutFile()

### Community 107 - "operatorversion_test.go"
Cohesion: 0.18
Nodes (19): operatorversion_test.go, TestOperatorApply(), compareVersions(), imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), operatorDeployJSON() (+11 more)

### Community 108 - "runLeader"
Cohesion: 0.60
Nodes (4): leaderRun, runLeader(), TestAssertLeaderAsksBeforeActingFromABackup(), TestAssertLeaderPrimaryAndMonitorNeverAsk()

### Community 109 - "adminsecret_test.go"
Cohesion: 0.21
Nodes (10): adminsecret_test.go, adminCfg(), TestAdminSecretNameFollowsTheStates(), TestAdminSecretRefusesTheOperatorsOwnNames(), TestContainersStillRequireTheAdminPassword(), TestKubernetesValidatesWithoutAnAdminPassword(), TestMonitorPasswordAndKeyNeedTheAdminPassword(), TestOperatorAdminSecretName() (+2 more)

### Community 110 - "scripts.go"
Cohesion: 0.06
Nodes (41): scripts_test.go, validCLILine(), archiveConfigsScript(), assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), enableDefaultUsersScript() (+33 more)

### Community 111 - "Execute"
Cohesion: 0.50
Nodes (4): Repository layout, TestExecute(), Execute(), main()

### Community 112 - "capRunner"
Cohesion: 0.13
Nodes (21): capCall, capRunner, New(), NewManager(), callIndex(), TestManagerLogsCLIShell(), mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime() (+13 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.11
Nodes (22): containerState, healthState, inspect_test.go, internal/container, transport_test.go, decodeInspect(), orUnknown(), printable() (+14 more)

### Community 126 - "bytes.Buffer"
Cohesion: 0.13
Nodes (28): watch_test.go, subtractWatch(), deniedCluster(), deployJSON(), namedOperatorJSON(), operatorItem(), TestAmbiguousOperatorAbortsTheDeploy(), TestDesiredWatchAppendsTheBrokerNamespace() (+20 more)

### Community 128 - "hostpath_test.go"
Cohesion: 0.08
Nodes (32): hostpath_test.go, guardConfig(), TestGuardConfigIsValid(), CheckHostPath(), expandTilde(), expandTildeToken(), HasPathSeparator(), IsAbsHostPath() (+24 more)

### Community 139 - "Capture"
Cohesion: 0.11
Nodes (20): BrokerType, Capture, cliMate, Omission, transcript_test.go, brokerTypeFromSchema(), parseHeader(), readHeaderField() (+12 more)

### Community 155 - "strings.Builder"
Cohesion: 0.07
Nodes (40): shorthand, WeightedNodeTerm, mdRow(), renderAbbrevDocs(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable() (+32 more)

## Knowledge Gaps
- **13 isolated node(s):** `solace`, `NO_COLOR`, `bash/000-env.sh`, `bash/010-deploy-operator.sh`, `bash/020-deploy-broker.sh` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 185 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **6 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `App` to `Platform`, `replication_test.go`, `watchLoop`, `Runner`, `cli_test.go`, `newRootCmd`, `step`, `exportconfig_test.go`, `commands.go`, `config_test.go`, `cli/platform_test.go`, `.dump`, `context.Context`, `replication_test.go`, `AdminSecret`, `confirm.go`?**
  _High betweenness centrality (0.046) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `newTestOps()` (e.g. with `importops_test.go` and `Injectable seams`) actually correct?**
  _`newTestOps()` has 117 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `NO_COLOR`, `bash/000-env.sh` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.060483159936519135 - nodes in this community are weakly interconnected._
- **Why does `Config` connect `AdminSecret` to `App`, `hostpath_test.go`, `newTestOps`, `Runner`, `cli_test.go`, `config_test.go`, `Manager`, `K8sConfig`, `CheckCommand`, `NewCluster`, `Coverage`, `convert_test.go`, `manager_test.go`, `coverage_test.go`, `Load`, `scaling_test.go`, `strings.Builder`, `internal/config`, `ResolveDomainCerts`, `NewCluster`, `Compose`, `secrets_test.go`, `localCfg`, `config_test.go`, `eqArgs`, `Role`, `prep_test.go`, `.dump`, `context.Context`, `TestGeneratorPageEmbedsTheCLI`, `replCfg`, `checkRow`, `Coverage`, `replication_test.go`, `EnvPairs`, `.podSection`, `replCfg`, `render_test.go`, `namespace_test.go`, `Container`, `replConfig`, `Container`, `containerTransport`, `stsName`, `adminsecret_test.go`, `capRunner`?**
  _High betweenness centrality (0.036) - this node is a cross-community bridge._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.07785602503912363 - nodes in this community are weakly interconnected._
- **Why does `Per-package doubles` connect `AdminSecret` to `App`, `newTestOps`, `Runner`, `cli_test.go`, `config_test.go`, `Manager`, `matechannel_test.go`, `NewCluster`, `convert_test.go`, `transform_test.go`, `manager_test.go`, `Load`, `switchplan_test.go`, `newLocalOps`, `runner_test.go`, `strings.Builder`, `importops_test.go`, `statusreport_test.go`, `NewCluster`, `secrets_test.go`, `exportconfig_test.go`, `VPNRepl`, `localCfg`, `eqArgs`, `Role`, `prep_test.go`, `.dump`, `context.Context`, `replication_test.go`, `replCfg`, `Shared env fixtures`, `replication_test.go`, `recRunner`, `.BackupRevertActivity`, `inspect_test.go`, `RenderMate`, `replConfig`, `examples_test.go`, `operatorversion_test.go`, `runLeader`, `Execute`, `capRunner`, `bytes.Buffer`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
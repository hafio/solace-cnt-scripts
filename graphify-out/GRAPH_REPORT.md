# Graph Report - solace-cnt-scripts  (2026-10-02)

## Corpus Check
- 197 files · ~629,454 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 3848 nodes · 16230 edges · 150 communities (134 shown, 16 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 2218 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `cbc89b48`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- newTestOps
- Sink
- context.Context
- cli_test.go
- testing.T
- Command Reference Docs
- Manager
- K8sConfig
- runRootWith
- CheckCommand
- captureStdout
- NewCluster
- ParseBlocks
- convert_test.go
- transform_test.go
- manager_test.go
- github.com/spf13/cobra.Command
- GenSecrets
- Bash Dev Script
- parse
- PowerShell Dev Script
- Abbreviation Sets
- BuildSwitchPlan
- newLocalOps
- runner_test.go
- renderDriver
- eqArgs
- ImportPlan
- desiredWatch
- Solace Module Root
- Config
- Docs and Legacy Scripts
- ParseVPNReplication
- ResolveDomainCerts
- Role
- App
- recRunner
- annotate_test.go
- completion_test.go
- diff_test.go
- runPlatform
- inject_test.go
- Ops
- exportconfig_test.go
- Echo
- k8s/inspect.go
- replicationVPNLines
- localCfg
- config_test.go
- BrokerType
- newRootCmd
- ExitCode
- Container Security Guidelines
- Operations
- Test catalogue
- Cluster
- MateChannel
- Ops
- .BackupRevertActivity
- confirm.go
- .configRows
- rootless_test.go
- strings.Builder
- output_test.go
- Configuration Guide
- confirmAction
- Mate Channel Tests
- Command Reference Doc
- Developer guide
- replApp
- TestServerCert
- platformOps
- internal/k8s
- emitYAML
- exit.go
- .ConfigureReplication
- Command
- newTeardownApplyFixture
- age
- ReplSite
- k8s/matechannel_test.go
- limits_test.go
- occCluster
- Troubleshooting Guide
- k8s/inspect_test.go
- MateConfig
- haCfg
- render_test.go
- vulnjudge/main.go
- RuleFor
- step
- Container
- Config
- Secret Reference Resolution
- ParseRole
- replConfig
- Abbreviation Reference
- Capture
- .checkUserManagerLimits
- platformTitle
- Load
- certFixture
- Platform
- internal/cli
- .rpc
- AdminSecret
- Import-Config Apply Rules
- Config
- internal/broker
- GenOperator
- capRunner
- Config Package Tests
- newPerformExportConfigCmd
- operatorversion_test.go
- newBlock
- internal/container
- Container State Inspection
- cliTransport
- cliMate
- scripts.go
- matchCLI
- .validateContainerArtifactValues
- sempPort
- runExport
- Convert
- .applyScalingTierDefaults
- go_pkg_runtime
- Scaling YAML Decoding
- podName
- .preflightOne
- ReplVia
- Namespace Protection Checks
- logArgs
- TestShowRedundancyIsReadOnly
- Env Path Resolution
- Duration and Host Path Checks
- Exporting and importing configuration
- internal/engine

## God Nodes (most connected - your core abstractions)
1. `newTestOps()` - 131 edges
2. `ctrCfg()` - 131 edges
3. `Role` - 128 edges
4. `Config` - 113 edges
5. `newCapMgr()` - 111 edges
6. `NewCluster()` - 101 edges
7. `loadK8s()` - 89 edges
8. `Platform` - 79 edges
9. `matchCLI()` - 75 edges
10. `bg()` - 70 edges

## Surprising Connections (you probably didn't know these)
- `parseHeader()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `parseHeader()` --calls--> `readMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `applyMetaField()` --calls--> `BrokerType`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go
- `applyMetaField()` --calls--> `brokerTypeFromSchema()`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go
- `sectionBeginMarker()` --calls--> `RuleFor()`  [INFERRED]
  internal/broker/annotate.go → internal/broker/sections.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (150 total, 16 thin omitted)

### Community 0 - "bg"
Cohesion: 0.16
Nodes (33): bg(), k8sAdminPassword(), k8sCluster(), k8sContext(), k8sLogin(), opK8sCLI(), opK8sConfigLeader(), opK8sCopyFrom() (+25 more)

### Community 1 - "newTestOps"
Cohesion: 0.07
Nodes (51): newTestOps(), ranContains(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultVPN(), TestDomainCerts(), TestDomainCertsAcceptsAFullHostPath(), TestDomainCertsEmptySkips() (+43 more)

### Community 2 - "Sink"
Cohesion: 0.09
Nodes (25): BlockDiff, DiffResult, blockLabel(), diffRow(), orDash(), RenderDiffResult(), reportBlockLines(), TestDiffBlockLabelFormatsIdentityForTheReport() (+17 more)

### Community 3 - "context.Context"
Cohesion: 0.04
Nodes (17): scriptedMate, Transport, Manager, idMapCovers(), origin(), runUserIDs(), EnvRunner, Runner (+9 more)

### Community 4 - "cli_test.go"
Cohesion: 0.08
Nodes (57): capture(), captureStderr(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone(), runStatusStderr(), TestBashEnvGivenToEnvFlag() (+49 more)

### Community 5 - "testing.T"
Cohesion: 0.02
Nodes (202): TestDiagnosticsRunError(), TestDomainCertsUploadFileError(), TestExecCLIFailsOnErrorOutput(), TestExecCLIUploadFileError(), TestGatherNodeDownloadError(), TestOutDefaultsToStdout(), TestPollCondError(), TestPollContextCancelled() (+194 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (7): certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), secretSummary(), setOrMissing()

### Community 8 - "K8sConfig"
Cohesion: 0.17
Nodes (7): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, writeSecurity()

### Community 9 - "runRootWith"
Cohesion: 0.09
Nodes (51): TestAllowCommandRejectedWhereNothingExecutes(), allowRuntime(), echoRunner(), runCtr(), runRootWith(), TestAnnounceCommandsNamesResolvedBinaries(), TestAssertLeaderStandaloneSkipsBeforeAnyPrompt(), TestConfigStepsDoNotLeakSecrets() (+43 more)

### Community 10 - "CheckCommand"
Cohesion: 0.15
Nodes (21): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+13 more)

### Community 11 - "captureStdout"
Cohesion: 0.12
Nodes (34): leaderRun, opCall, captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook(), loadDirect() (+26 more)

### Community 12 - "NewCluster"
Cohesion: 0.08
Nodes (61): TestCheckStopsProbingWhenUnreachable(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts(), TestValidateReportsTheDerivedImagePullSecretName() (+53 more)

### Community 13 - "ParseBlocks"
Cohesion: 0.12
Nodes (22): Annotate(), sectionBeginMarker(), TestAnnotateIsDeterministic(), TestAnnotateRegionAndSectionFencesAreBalanced(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), TestSectionBeginMarkerAdvisoryDisposition(), TestWriteBlocksOpensAndClosesOnChange() (+14 more)

### Community 14 - "convert_test.go"
Cohesion: 0.19
Nodes (31): checkGolden(), convertOK(), hasWarning(), strictDecode(), TestConvertAdminSecretAlias(), TestConvertAdminUserIsDroppedOnEveryPlatform(), TestConvertBadBooleanWarns(), TestConvertBadNumberWarns() (+23 more)

### Community 15 - "transform_test.go"
Cohesion: 0.08
Nodes (53): Block, Region, TargetState, regionFor(), keepRegion(), injectedBlock(), BridgeEnablementInverted(), ClearExistingNested() (+45 more)

### Community 16 - "manager_test.go"
Cohesion: 0.06
Nodes (116): GuardPodmanEUID(), inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman() (+108 more)

### Community 17 - "github.com/spf13/cobra.Command"
Cohesion: 0.21
Nodes (45): addCommands(), addLogFlags(), group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerConfigureCmd(), newBrokerPerformCmd(), newBrokerRemoveCmd() (+37 more)

### Community 18 - "GenSecrets"
Cohesion: 0.07
Nodes (38): GenBroker(), GenSecrets(), joinManifests(), adminCfg(), secretNamesIn(), TestAdminSecretStatesEmitTheirArtifacts(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree() (+30 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "parse"
Cohesion: 0.11
Nodes (20): segment, vars, countMarkers(), resolvePlatform(), TestParseArrayElementsPreserveQuoting(), TestParseAssignmentForms(), TestParseCRLF(), TestParseDoubleQuotedBackslashEscapes() (+12 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Abbreviation Sets"
Cohesion: 0.13
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 23 - "BuildSwitchPlan"
Cohesion: 0.12
Nodes (29): PhaseKind, SiteState, SwitchAction, SwitchPhase, BuildSwitchPlan(), confirmDemotions(), ExecuteSwitchPlan(), SwitchPlan (+21 more)

### Community 24 - "newLocalOps"
Cohesion: 0.12
Nodes (32): TestLeaderAssertLeaderError(), TestRedundancyBackupShowError(), TestRedundancyReleaseError(), TestRedundancyRevertToPrimaryError(), TestReleaseToBackupNoReleaseError(), TestReleaseToBackupUnreleasedTimeout(), CLIScriptPath(), cliScriptPath() (+24 more)

### Community 25 - "runner_test.go"
Cohesion: 0.07
Nodes (35): TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), MaskEnv(), quoteTok(), captureStdout() (+27 more)

### Community 26 - "renderDriver"
Cohesion: 0.11
Nodes (21): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), shQuote() (+13 more)

### Community 27 - "eqArgs"
Cohesion: 0.11
Nodes (30): TestReachable(), newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), saCfg(), TestHARoles() (+22 more)

### Community 28 - "ImportPlan"
Cohesion: 0.14
Nodes (19): ImportResult, PlannedSection, isPragma(), defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList() (+11 more)

### Community 29 - "desiredWatch"
Cohesion: 0.17
Nodes (11): desiredWatch(), Cluster, splitWatch(), subtractWatch(), TestDesiredWatchAppendsTheBrokerNamespace(), TestSubtractWatchEmptyRemainingMeansDelete(), TestSubtractWatchPreservesOrder(), TestUnionWatchKeepsAnotherEnvFilesNamespace() (+3 more)

### Community 31 - "Config"
Cohesion: 0.11
Nodes (14): keyValueEntries, TestValidateContainerArtifactValues(), foldToEnvVar(), Config, invertedCPUSetRange(), missingErr(), platformKey(), requireAll() (+6 more)

### Community 32 - "Docs and Legacy Scripts"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "ParseVPNReplication"
Cohesion: 0.13
Nodes (16): colSpan, ValidVPNName(), validVPNName(), dashSpans(), flagByte(), gutterClear(), ParseVPNReplication(), sliceSpan() (+8 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (41): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domainCANames(), caNameSafe(), checkNoDotDot() (+33 more)

### Community 35 - "Role"
Cohesion: 0.08
Nodes (20): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+12 more)

### Community 36 - "App"
Cohesion: 0.11
Nodes (34): TestCtrManagerConfirmWiring(), childExit(), TestChildExitKeepsItsMessage(), TestExitCodeIsNeverNegative(), resolveScriptPath(), containerRenderRole(), containerRole(), ctrLogin() (+26 more)

### Community 37 - "recRunner"
Cohesion: 0.19
Nodes (4): TestCanIAnswerReadsTheLastLine(), TestTransportExecArgs(), recRunner, rrCall

### Community 38 - "annotate_test.go"
Cohesion: 0.07
Nodes (44): applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV(), readMarker(), renderKV() (+36 more)

### Community 39 - "completion_test.go"
Cohesion: 0.18
Nodes (21): driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion(), TestCompletionDescriptionsAreOptIn() (+13 more)

### Community 40 - "diff_test.go"
Cohesion: 0.15
Nodes (25): blockKey, mergedBlock, mergedLine, ancestorReported(), DiffBlocks(), indexByIdentity(), keyOf(), keySet() (+17 more)

### Community 41 - "runPlatform"
Cohesion: 0.29
Nodes (12): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+4 more)

### Community 42 - "inject_test.go"
Cohesion: 0.15
Nodes (27): lineRole, serviceLine, shutdownStyle, svcKey, classifyServiceRest(), containsPortCommand(), describeKey(), fieldsUnquoted() (+19 more)

### Community 43 - "Ops"
Cohesion: 0.11
Nodes (17): countContains(), field(), TestCountContains(), TestField(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), showRedundancyLocalScript(), defaultLocalAddrs() (+9 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.16
Nodes (18): opRunner, exportconfigDriverOutput(), exportconfigTransportOutput(), exportconfigVPNListOutput(), newExportconfigRunner(), newExportconfigRunnerRejectingChunk(), runImportConfirmation(), TestExportConfigOutOverwriteGate() (+10 more)

### Community 45 - "Echo"
Cohesion: 0.12
Nodes (8): interactiveFailRunner, runFailRunner, Exec, TestExecIsSilentWithoutVerbose(), Echo, NewExec(), Quote(), TestQuote()

### Community 46 - "k8s/inspect.go"
Cohesion: 0.13
Nodes (25): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), imageFromDeployment(), ownedPods(), operatorItem() (+17 more)

### Community 47 - "replicationVPNLines"
Cohesion: 0.23
Nodes (14): ReplicationConfigResult, replicationVPNLines(), siteAEntry(), TestNothingEverShutsDownTheVPNItself(), TestReplicationVPNLinesEnablesListedAndDisablesTheRest(), TestReplicationVPNLinesIsDeterministic(), TestReplicationVPNLinesQuotesNamesWithSpaces(), TestReplicationVPNLinesReEnablesWhatPhase1StoppedOnly() (+6 more)

### Community 48 - "localCfg"
Cohesion: 0.09
Nodes (36): TestLocalMateRunsThroughOps(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK(), TestBackupRevertActivitySuccess() (+28 more)

### Community 49 - "config_test.go"
Cohesion: 0.19
Nodes (24): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, span, dropMessageSpoolLines(), dropRoutingInterfaceLine(), TestSectionsDropMessageSpoolLines() (+16 more)

### Community 50 - "BrokerType"
Cohesion: 0.17
Nodes (12): BrokerType, bannerType(), checkProvenance(), checkSameType(), describeScope(), Ops, planSections(), TestImportOpsDescribeScope() (+4 more)

### Community 51 - "newRootCmd"
Cohesion: 0.10
Nodes (29): applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias(), TestAllowCommandApprovesAWrappedRuntime() (+21 more)

### Community 52 - "ExitCode"
Cohesion: 0.16
Nodes (16): TestExecute(), TestResolveScriptPath(), TestScriptCommandsRefuseBeforeAnythingRuns(), runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand(), TestPodmanEUIDGuardSkips(), writeRootlessPodmanEnv() (+8 more)

### Community 53 - "Container Security Guidelines"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 54 - "Operations"
Cohesion: 0.08
Nodes (24): Bringing up a fresh cluster, Configuring a site, Data replication, Docker and Podman mechanics, Every destructive command confirms, Exit codes, Extra CLI users differ by platform, Operations (+16 more)

### Community 55 - "Test catalogue"
Cohesion: 0.10
Nodes (21): abbrev_test.go, convert_test.go, Coverage, examples_test.go, Fixtures and doubles, Injectable seams, internal/abbrev, internal/convert (+13 more)

### Community 57 - "MateChannel"
Cohesion: 0.22
Nodes (12): MateChannel, requirePrimaryActive(), requireVirtualRouterNamesMatch(), sortedSiteNames(), SwitchPreflight(), replSites(), TestSwitchPreflightHealthyPair(), TestSwitchPreflightRequiresMatchingVirtualRouterNames() (+4 more)

### Community 58 - "Ops"
Cohesion: 0.11
Nodes (10): cliRunNames(), Ops, runCLISkeleton(), TestValidName(), validName(), ValidScriptName(), rejectionIn(), shellScriptPath() (+2 more)

### Community 59 - ".BackupRevertActivity"
Cohesion: 0.16
Nodes (15): backupTarget, Credential, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigLine(), Ops, lastHTTPStatus(), revertActivityBackupBody() (+7 more)

### Community 60 - "confirm.go"
Cohesion: 0.17
Nodes (21): exportconfigReadCounter, layer, TestConfirmDeleteShortcut(), TestPromptYes(), TestPromptYesNo(), TestStdinCanAnswerClosedFile(), confirmActionStrict(), confirmDelete() (+13 more)

### Community 61 - ".configRows"
Cohesion: 0.13
Nodes (23): setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminSecretRow(), containsString(), failRow(), Cluster (+15 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.17
Nodes (35): failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent(), TestCheckIDMappingRefusesAnUnmappedRunUser() (+27 more)

### Community 63 - "strings.Builder"
Cohesion: 0.07
Nodes (42): shorthand, WeightedNodeTerm, mdRow(), renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeAbbrevTable() (+34 more)

### Community 64 - "output_test.go"
Cohesion: 0.24
Nodes (14): sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLevelOrderKeepsSkipBelowFail(), TestLineAddsNoPrefix(), TestSectionPadsToWidth() (+6 more)

### Community 65 - "Configuration Guide"
Cohesion: 0.14
Nodes (14): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+6 more)

### Community 66 - "confirmAction"
Cohesion: 0.20
Nodes (23): confirmAction(), wantEnable(), wantRemove(), confirmAssertFromBackup(), containerWhat(), opCtrConfigDefaultUsers(), opCtrConfigDefaultVPN(), opCtrConfigDomainCerts() (+15 more)

### Community 67 - "Mate Channel Tests"
Cohesion: 0.16
Nodes (17): CLIRunner, fakeRun, Ops, NewCLIMate(), mateReplies(), newTestMate(), TestMateChannelDescribeNamesTheTarget(), TestMateChannelPreflightLearnsTheBrokerType() (+9 more)

### Community 68 - "Command Reference Doc"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.20
Nodes (10): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+2 more)

### Community 70 - "replApp"
Cohesion: 0.32
Nodes (12): mateChannel(), mateSEMPPassword(), replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer(), TestMateChannelPicksTheMechanismTheSiteDeclares(), TestMateChannelRefusesASiteWithNoVia() (+4 more)

### Community 71 - "TestServerCert"
Cohesion: 0.11
Nodes (19): TestDiagnostics(), TestPathHelpers(), TestServerCert(), TestServerCertBundleOrder(), TestServerCertBundleRefusesAKeyBearingCert(), TestServerCertBundleReportsAnUnreadableFile(), TestServerCertBundleRequiresBothHalves(), writeFile() (+11 more)

### Community 72 - "platformOps"
Cohesion: 0.33
Nodes (18): opFunc, roleOpFunc, addPodFlag(), newBrokerCopyCmd(), newBrokerDeployCmd(), newBrokerExecCmd(), newBrokerGenerateCmd(), newBrokerLogsCmd() (+10 more)

### Community 73 - "internal/k8s"
Cohesion: 0.10
Nodes (21): adminsecret_test.go, check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, generatorpage_test.go, inspect_test.go, internal/k8s (+13 more)

### Community 74 - "emitYAML"
Cohesion: 0.22
Nodes (9): doc, boolOf(), commentSafe(), emitYAML(), joinDomainCertPath(), kubeCommand(), redundancy(), scalar() (+1 more)

### Community 75 - "exit.go"
Cohesion: 0.14
Nodes (12): childExitError, usageError, asUsage(), childStatus(), isUsage(), markUsageArgs(), TestEveryArgValidatorIsAUsageError(), TestUsageErrorKeepsItsMessage() (+4 more)

### Community 76 - ".ConfigureReplication"
Cohesion: 0.13
Nodes (14): QueueState, VPNRepl, Ops, isRunCLIRejection(), mateConvergenceShutdowns(), missingListedVPNs(), newlineIf(), reenableLine() (+6 more)

### Community 77 - "Command"
Cohesion: 0.12
Nodes (13): cmdField, guardedCmd, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandUnmarshal(), TestValidateProbeCommandAccepts(), TestValidateProbeCommandRejects(), Command (+5 more)

### Community 78 - "newTeardownApplyFixture"
Cohesion: 0.13
Nodes (16): driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), minimalCapture(), newTeardownApplyFixture(), targetVPNCapture(), TestImportOpsExportConfigScopeSelectsCLICommand() (+8 more)

### Community 79 - "age"
Cohesion: 0.22
Nodes (9): orNone(), orValue(), age(), roleRank(), Cluster, operatorRunningImage(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+1 more)

### Community 81 - "ReplSite"
Cohesion: 0.16
Nodes (12): ReplRole, Replication, PlannedRoles(), RoleAtSite(), TestPlannedRolesNamesEveryListedVPN(), TestRoleAtSiteIsTheComplement(), setReplicationRoleScript(), curlConfigFlag() (+4 more)

### Community 82 - "k8s/matechannel_test.go"
Cohesion: 0.24
Nodes (24): execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf(), replCfg(), replSite() (+16 more)

### Community 83 - "limits_test.go"
Cohesion: 0.13
Nodes (27): missingControllers(), limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessCoreFloorFollowsTheConfiguredLimit(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager() (+19 more)

### Community 84 - "occCluster"
Cohesion: 0.18
Nodes (10): occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants(), TestNamespaceIsProtectedSharesDeleteNamespacesList() (+2 more)

### Community 85 - "Troubleshooting Guide"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 86 - "k8s/inspect_test.go"
Cohesion: 0.14
Nodes (20): sectionLevel(), podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture() (+12 more)

### Community 87 - "MateConfig"
Cohesion: 0.11
Nodes (35): MateConfig, TestMateChannelShowReplication(), containsEndpoint(), labelValue(), normTransport(), parseHostPort(), ParseShowReplication(), parseShowReplicationAppliance() (+27 more)

### Community 88 - "haCfg"
Cohesion: 0.09
Nodes (32): New(), TestNewDefaults(), AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestReadAdminPasswordArgvAndDecoding() (+24 more)

### Community 89 - "render_test.go"
Cohesion: 0.05
Nodes (85): ContainerNoFile(), HealthCheck, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), boolStr(), BrokerCR(), Compose(), composeEscape() (+77 more)

### Community 90 - "vulnjudge/main.go"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.08
Nodes (36): Disposition, SectionRule, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified(), renderImportDocs() (+28 more)

### Community 92 - "step"
Cohesion: 0.14
Nodes (17): mateChannelFunc, App, lineSink(), step(), confirmReplicationConfig(), opConfigureReplication(), opCtrConfigureReplication(), opCtrPerformReplication() (+9 more)

### Community 93 - "Container"
Cohesion: 0.17
Nodes (12): Container, DockerConfig, Network, PodmanConfig, Ulimits, applyContainerBlockDefaults(), Config, podmanRunUser() (+4 more)

### Community 94 - "Config"
Cohesion: 0.20
Nodes (6): expandTilde(), expandTildeToken(), Config, isPathSep(), TestExpandTilde(), tildeHome()

### Community 95 - "Secret Reference Resolution"
Cohesion: 0.50
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 96 - "ParseRole"
Cohesion: 0.13
Nodes (15): argumentLine(), noRolePositional(), completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), isEnvFileName(), matching() (+7 more)

### Community 97 - "replConfig"
Cohesion: 0.31
Nodes (9): replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationEmptyViaAccepted(), TestValidateReplicationRejects(), TestValidateReplicationValid() (+1 more)

### Community 98 - "Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "Capture"
Cohesion: 0.25
Nodes (7): Omission, brokerTypeFromSchema(), Capture, parseHeader(), readHeaderField(), TestBrokerTypeFromSchema(), omittedList()

### Community 100 - ".checkUserManagerLimits"
Cohesion: 0.29
Nodes (6): delegateSetting(), Manager, limitText(), parseLimit(), parseUnitProps(), TestParseLimit()

### Community 102 - "Load"
Cohesion: 0.07
Nodes (53): Example, TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadReadError() (+45 more)

### Community 103 - "certFixture"
Cohesion: 0.09
Nodes (31): callIndex(), TestPreflightRunsBeforeAnything(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime(), TestManagerReachableProbesRuntimeThenCompose(), withWrapper(), wrappedCtrCfg() (+23 more)

### Community 104 - "Platform"
Cohesion: 0.09
Nodes (34): newConvertCmd(), runConvert(), usagef(), checkAllowCommand(), checkFlagPlatforms(), commandPlatforms(), declaredList(), parsePlatformList() (+26 more)

### Community 105 - "internal/cli"
Cohesion: 0.15
Nodes (13): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, euid_test.go, examples_test.go (+5 more)

### Community 106 - ".rpc"
Cohesion: 0.17
Nodes (7): AdminState, sempMate, httpBody(), TestHTTPBodyStripsHeaders(), sempAdminState(), sempReplRole(), xmlEscape()

### Community 107 - "AdminSecret"
Cohesion: 0.18
Nodes (13): AdditionalUsersSecret(), AdminSecret(), decodeDataValue(), TestAdditionalUsersSecretCarriesBothHalves(), TestAdditionalUsersSecretIsAbsentWithNoUsers(), TestAdditionalUsersSecretRefusesAnEmptyPassword(), TestAdditionalUsersStayOutOfTheCredentialsSecret(), TestAdminSecretCarriesThePSKOnlyWhenSet() (+5 more)

### Community 108 - "Import-Config Apply Rules"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "Config"
Cohesion: 0.08
Nodes (15): AdditionalUser, Image, Node, Redundancy, SEMP, TLS, atoiPrefix(), Config (+7 more)

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "GenOperator"
Cohesion: 0.14
Nodes (22): TestOperatorInstallVerdict(), GenOperator(), joinYAMLDocs(), nonEmpty(), operatorProbes(), RenderOperator(), renderOperatorNS(), splitAfterNamespace() (+14 more)

### Community 112 - "capRunner"
Cohesion: 0.15
Nodes (14): capCall, capRunner, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), unapprovedCtrCfg(), NewTransport(), dockerCfg(), podmanCfg() (+6 more)

### Community 113 - "Config Package Tests"
Cohesion: 0.12
Nodes (17): adminsecret_test.go, artifactvalues_test.go, command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go (+9 more)

### Community 114 - "newPerformExportConfigCmd"
Cohesion: 0.29
Nodes (8): newPerformExportConfigCmd(), registerFlagCompletion(), newExamplesCmd(), runExample(), addAllowCommandFlag(), addExportFlags(), addOutFlags(), addRestartFlag()

### Community 115 - "operatorversion_test.go"
Cohesion: 0.21
Nodes (13): compareVersions(), imageTag(), operatorVersionWarning(), parseVersion(), operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsSilentInAPreview() (+5 more)

### Community 116 - "newBlock"
Cohesion: 0.28
Nodes (9): token, firstQuoted(), newBlock(), TestNewBlockMessageSpoolQualifierHasNoOwnName(), TestNewBlockOpener(), TestNewBlockVPNNameWithSpaces(), TestNewBlockVPNQualifierWithSpaces(), tokenizeOpener() (+1 more)

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.16
Nodes (14): containerState, healthState, decodeInspect(), Manager, orUnknown(), printable(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown() (+6 more)

### Community 119 - "cliTransport"
Cohesion: 0.22
Nodes (8): cliTransport(), renderMateAppliance(), renderMateSoftware(), schemaTransport(), sortedKeys(), sortedSet(), TestTransportVocabularyIsThreeWords(), sortedVPNs()

### Community 121 - "scripts.go"
Cohesion: 0.07
Nodes (41): showCmd, validCLILine(), Ops, assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript() (+33 more)

### Community 122 - "matchCLI"
Cohesion: 0.08
Nodes (41): hasCall(), leaderCalls(), matchCLI(), TestDisableDefaultUsersRefusesAQuotedVPNName(), TestLeaderSuccess(), TestLeaderSuccessOnBackup(), TestDisableDefaultUsersDisableError(), TestDisableDefaultUsersShowVPNError() (+33 more)

### Community 123 - ".validateContainerArtifactValues"
Cohesion: 0.20
Nodes (7): Config, portOrRange(), TestValidContainerPort(), validContainerPort(), validCoreLimit(), validHealthDuration(), validDNSLabel()

### Community 124 - "sempPort"
Cohesion: 0.29
Nodes (7): bridgeHostPort(), sempPort(), TestSempPortBridgeFallsBackToPlaintext(), TestSempPortBridgePrefersTLS(), TestSempPortBridgeRefusesNoMapping(), TestSempPortHostMode(), TestSempPortWithoutCertificateStaysPlaintext()

### Community 125 - "runExport"
Cohesion: 0.13
Nodes (14): confirmImport(), pluralVPN(), runExport(), runImport(), TestPluralVPNWording(), emitOrWrite(), App, opCtrExportConfig() (+6 more)

### Community 126 - "Convert"
Cohesion: 0.40
Nodes (5): Result, Convert(), TestConvertUnterminatedArray(), TestGeneratedHeader(), TestGeneratedHeaderSanitisesSource()

### Community 127 - ".applyScalingTierDefaults"
Cohesion: 0.15
Nodes (11): scalingTier, containerMem(), cpuSetCount(), cpuSetRange(), Config, TestContainerMem(), TestCPUSetCount(), TestCPUSetRange() (+3 more)

### Community 128 - "go_pkg_runtime"
Cohesion: 0.40
Nodes (3): TestEmit(), emit(), newVersionCmd()

### Community 129 - "Scaling YAML Decoding"
Cohesion: 0.31
Nodes (5): scalingKey, scalingSpelling, Scaling, scalingKeyIndex(), scalingKeyList()

### Community 130 - "podName"
Cohesion: 0.09
Nodes (17): BrokerPodSuffixShape(), Cluster, TestOperatorAdminSecretNameMatchesTheBrokerNames(), Cluster, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes(), HARoles() (+9 more)

### Community 131 - ".preflightOne"
Cohesion: 0.50
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "ReplVia"
Cohesion: 0.22
Nodes (8): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, checkCredentialChars(), sortedKeys(), validateReplVia(), validateReplViaSEMP()

### Community 134 - "logArgs"
Cohesion: 0.67
Nodes (3): logArgs(), logArgs2(), TestLogArgsBuildsOneSetForBothPlatforms()

### Community 135 - "TestShowRedundancyIsReadOnly"
Cohesion: 0.67
Nodes (3): TestBackupActivityStateReadsTheMateColumn(), TestShowRedundancyIsReadOnly(), backupActivityState()

### Community 139 - "Env Path Resolution"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 141 - "Duration and Host Path Checks"
Cohesion: 0.40
Nodes (5): CanonicalDuration(), TestCanonicalDuration(), CheckHostPath(), TestCheckHostPathAccepts(), TestCheckHostPathRejects()

### Community 153 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 158 - "internal/engine"
Cohesion: 0.67
Nodes (3): internal/engine, resolve_test.go, runner_test.go

## Knowledge Gaps
- **238 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+233 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 359 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **16 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Sink` connect `Sink` to `output_test.go`, `Manager`, `age`, `config_test.go`, `BrokerType`, `k8s/inspect_test.go`, `Cluster`, `Ops`, `ImportPlan`, `.configRows`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `newTestOps`, `podName`, `context.Context`, `Manager`, `K8sConfig`, `runRootWith`, `captureStdout`, `NewCluster`, `convert_test.go`, `manager_test.go`, `GenSecrets`, `eqArgs`, `desiredWatch`, `ResolveDomainCerts`, `Role`, `localCfg`, `config_test.go`, `Cluster`, `Ops`, `.configRows`, `strings.Builder`, `TestServerCert`, `ReplSite`, `k8s/matechannel_test.go`, `occCluster`, `haCfg`, `render_test.go`, `step`, `Container`, `certFixture`, `AdminSecret`, `GenOperator`, `capRunner`, `sempPort`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Why does `Platform` connect `Platform` to `testing.T`, `Manager`, `CheckCommand`, `captureStdout`, `convert_test.go`, `manager_test.go`, `github.com/spf13/cobra.Command`, `parse`, `Config`, `Role`, `config_test.go`, `Ops`, `platformOps`, `emitYAML`, `Command`, `limits_test.go`, `render_test.go`, `step`, `Container`, `Config`, `platformTitle`, `Load`, `certFixture`, `capRunner`, `.validateContainerArtifactValues`, `sempPort`, `Convert`, `.applyScalingTierDefaults`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 49 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 49 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _238 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.0726764500349406 - nodes in this community are weakly interconnected._
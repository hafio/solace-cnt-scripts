# Graph Report - solace-cnt-scripts  (2026-10-01)

## Corpus Check
- 197 files · ~621,650 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 3837 nodes · 16114 edges · 159 communities (140 shown, 19 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 2214 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `256c4db3`
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
- Quadlet
- CheckCommand
- captureStdout
- NewCluster
- Cluster
- convert_test.go
- transform_test.go
- manager_test.go
- commands.go
- GenSecrets
- Bash Dev Script
- MateConfig
- PowerShell Dev Script
- Abbreviation Sets
- BuildSwitchPlan
- newLocalOps
- runner_test.go
- renderDriver
- prep_test.go
- Import Plan and Apply
- newEchoMgr
- Solace Module Root
- Config
- Docs and Legacy Scripts
- ParseVPNReplication
- ResolveDomainCerts
- Role
- App
- recRunner
- annotate_test.go
- runRoot
- diff.go
- Platform Selection
- inject_test.go
- .releaseToBackup
- hasCall
- Echo
- k8s/inspect.go
- replicationVPNLines
- localCfg
- config_test.go
- ParseBlocks
- newRootCmd
- exportconfig_test.go
- Container Security Guidelines
- Operations
- Test catalogue
- Cluster
- MateChannel
- Ops
- .rpc
- confirm.go
- .configRows
- rootless_test.go
- strings.Builder
- output_test.go
- Configuration Guide
- Ops
- Mate Channel Tests
- Command Reference Doc
- Developer guide
- replApp
- parse
- age
- internal/k8s
- Capture
- usagef
- .ConfigureReplication
- Command
- newTeardownApplyFixture
- Container File Transport
- BrokerType
- k8s/matechannel_test.go
- limits_test.go
- occCluster
- Troubleshooting Guide
- Kubernetes Inspect Tests
- ParseShowReplication
- operatorversion_test.go
- load
- vulnjudge/main.go
- Import Section Rules
- step
- Container
- Config
- Secret Reference Resolution
- github.com/spf13/cobra.Command
- replConfig
- Abbreviation Reference
- newBlock
- .checkUserManagerLimits
- .Preflight
- Load
- container/secrets_test.go
- Platform
- internal/cli
- captureStderr
- Compose
- Import-Config Apply Rules
- Image
- internal/broker
- Config
- capRunner
- Config Package Tests
- TestServerCert
- operatorversion.go
- EnvPairs
- internal/container
- Container State Inspection
- podName
- Get
- scripts.go
- coverage_test.go
- .validateContainerArtifactValues
- emitYAML
- App
- .applyScalingTierDefaults
- .validateContainer
- Manager
- Scaling YAML Decoding
- HARoles
- .preflightOne
- ReplVia
- Namespace Protection Checks
- .LeaderLocal
- scripts_test.go
- ReplSite
- ContainerSecrets
- Convert
- Env Path Resolution
- parsePort
- Duration and Host Path Checks
- storageCfg
- Exporting and importing configuration
- ContainerSecret
- .gatherNode
- TestImportOpsExportConfigScopeSelectsCLICommand
- euid_test.go
- internal/engine

## God Nodes (most connected - your core abstractions)
1. `ctrCfg()` - 131 edges
2. `newTestOps()` - 130 edges
3. `Role` - 122 edges
4. `Config` - 113 edges
5. `newCapMgr()` - 111 edges
6. `NewCluster()` - 100 edges
7. `loadK8s()` - 89 edges
8. `matchCLI()` - 81 edges
9. `Platform` - 78 edges
10. `bg()` - 70 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `ExitCode()`  [EXTRACTED]
  main.go → internal/cli/exit.go
- `parseHeader()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `parseHeader()` --calls--> `readMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `applyMetaField()` --calls--> `BrokerType`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go
- `applyMetaField()` --calls--> `brokerTypeFromSchema()`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (159 total, 19 thin omitted)

### Community 0 - "bg"
Cohesion: 0.12
Nodes (51): TestConfirmDeleteShortcut(), confirmDelete(), warn(), wantEnable(), bg(), k8sAdminPassword(), k8sCluster(), k8sContext() (+43 more)

### Community 1 - "newTestOps"
Cohesion: 0.09
Nodes (45): newTestOps(), outputForRole(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultUsersRefusesAQuotedVPNName(), TestDisableDefaultVPN(), TestDomainCerts(), TestDomainCertsAcceptsAFullHostPath() (+37 more)

### Community 2 - "Sink"
Cohesion: 0.20
Nodes (5): solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), columnWidths(), Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.07
Nodes (11): Exec, EnvRunner, Runner, Cluster, readSecretKey(), Cluster, TestReportsDegradeRatherThanFalselyAlarm(), Cluster (+3 more)

### Community 4 - "cli_test.go"
Cohesion: 0.11
Nodes (51): TestAllowCommandRejectedWhereNothingExecutes(), echoRunner(), runCtr(), runRootWith(), TestCLICommand(), TestConfigStepsDoNotLeakSecrets(), TestConfiguredRouternameSurvivesTheFallback(), TestConvertRoundTrip() (+43 more)

### Community 5 - "testing.T"
Cohesion: 0.02
Nodes (188): TestRouterNameReadsTheCaptureHeader(), TestEveryDocCarriesTheSupportNotice(), TestHelperExitProcess(), adminCfg(), TestAdminSecretNameFollowsTheStates(), TestAdminSecretRefusesTheOperatorsOwnNames(), TestContainersStillRequireTheAdminPassword(), TestKubernetesValidatesWithoutAnAdminPassword() (+180 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (9): NodeIdentity, certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary() (+1 more)

### Community 8 - "K8sConfig"
Cohesion: 0.17
Nodes (7): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, writeSecurity()

### Community 9 - "Quadlet"
Cohesion: 0.12
Nodes (21): ContainerNoFile(), HealthCheck, TestLimitsCheckAssertsWhatTheArtifactAsks(), escapePercent(), healthCmd(), Quadlet(), quadletEscape(), containerArtifacts() (+13 more)

### Community 10 - "CheckCommand"
Cohesion: 0.16
Nodes (21): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+13 more)

### Community 11 - "captureStdout"
Cohesion: 0.12
Nodes (29): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook(), loadDirect() (+21 more)

### Community 12 - "NewCluster"
Cohesion: 0.08
Nodes (64): TestCheckStopsProbingWhenUnreachable(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+56 more)

### Community 13 - "Cluster"
Cohesion: 0.22
Nodes (3): Cluster, namespaceManifest(), ownedSecretNames()

### Community 14 - "convert_test.go"
Cohesion: 0.19
Nodes (31): checkGolden(), convertOK(), hasWarning(), strictDecode(), TestConvertAdminSecretAlias(), TestConvertAdminUserIsDroppedOnEveryPlatform(), TestConvertBadBooleanWarns(), TestConvertBadNumberWarns() (+23 more)

### Community 15 - "transform_test.go"
Cohesion: 0.08
Nodes (53): Block, TargetState, importIgnore(), TestImportIgnoreKeepsUnclassifiedSectionsInTheDiff(), TestImportOpsImportIgnore(), injectedBlock(), BridgeEnablementInverted(), ClearExistingNested() (+45 more)

### Community 16 - "manager_test.go"
Cohesion: 0.07
Nodes (94): inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer() (+86 more)

### Community 17 - "commands.go"
Cohesion: 0.18
Nodes (45): opFunc, roleOpFunc, addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerConfigureCmd(), newBrokerCopyCmd() (+37 more)

### Community 18 - "GenSecrets"
Cohesion: 0.11
Nodes (22): GenSecrets(), TestGenSecretsBuildsNothingWithoutMaterial(), AdditionalUsersSecret(), AdminSecret(), decodeDataValue(), TestAdditionalUsersSecretCarriesBothHalves(), TestAdditionalUsersSecretIsAbsentWithNoUsers(), TestAdditionalUsersSecretRefusesAnEmptyPassword() (+14 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "MateConfig"
Cohesion: 0.16
Nodes (17): MateConfig, cliTransport(), containsEndpoint(), labelValue(), normTransport(), parseHostPort(), parseShowReplicationAppliance(), parseShowReplicationSoftware() (+9 more)

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
Cohesion: 0.10
Nodes (41): TestReleaseToBackupUnreleasedTimeout(), CLIScriptPath(), cliScriptPath(), isCurl(), newLocalOps(), rd(), seqTransport(), TestDetectRoleAddrsError() (+33 more)

### Community 25 - "runner_test.go"
Cohesion: 0.10
Nodes (30): TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), MaskEnv(), NewExec(), quoteTok() (+22 more)

### Community 26 - "renderDriver"
Cohesion: 0.09
Nodes (26): chunk, chunkResult, runCLISkeleton(), chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment() (+18 more)

### Community 27 - "prep_test.go"
Cohesion: 0.05
Nodes (74): AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestReadAdminPasswordArgvAndDecoding(), TestReadAdminPasswordGuardsTheCommand(), TestReadAdminPasswordNamesTheNextStep() (+66 more)

### Community 28 - "Import Plan and Apply"
Cohesion: 0.14
Nodes (20): ImportResult, PlannedSection, isPragma(), defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList() (+12 more)

### Community 29 - "newEchoMgr"
Cohesion: 0.10
Nodes (17): NewManager(), newEchoMgr(), TestCheckEnvReportsBaseDirOnlyWhenSet(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployPodmanDryRunHidesSecretBytes(), TestManagerDockerCheckProbesCompose(), TestManagerLifecycleDockerDryRunUsesCompose() (+9 more)

### Community 31 - "Config"
Cohesion: 0.12
Nodes (13): keyValueEntries, checkCredentialChars(), foldToEnvVar(), Config, missingErr(), requireKeyValue(), sortedKeys(), validateReplVia() (+5 more)

### Community 32 - "Docs and Legacy Scripts"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "ParseVPNReplication"
Cohesion: 0.14
Nodes (15): colSpan, ValidVPNName(), validVPNName(), dashSpans(), flagByte(), gutterClear(), ParseVPNReplication(), sliceSpan() (+7 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (41): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domainCANames(), caNameSafe(), checkNoDotDot() (+33 more)

### Community 35 - "Role"
Cohesion: 0.08
Nodes (18): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+10 more)

### Community 36 - "App"
Cohesion: 0.10
Nodes (41): TestCtrManagerConfirmWiring(), confirmAction(), childExit(), TestChildExitKeepsItsMessage(), TestExitCodeIsNeverNegative(), wantRemove(), containerRole(), containerWhat() (+33 more)

### Community 37 - "recRunner"
Cohesion: 0.09
Nodes (17): TestCanIAnswerReadsTheLastLine(), TestExecutorRefusesUnapprovedRuntime(), unapprovedCfg(), NewTransport(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs(), TestTransportUpload() (+9 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.07
Nodes (43): Region, applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV(), readMarker() (+35 more)

### Community 39 - "runRoot"
Cohesion: 0.08
Nodes (43): firstLine(), runRoot(), TestBashEnvGivenToEnvFlag(), TestConfigureServerCertsRefusesASecretItDoesNotOwn(), TestConvertErrorPaths(), TestConvertParseError(), TestConvertToFile(), TestConvertToStdout() (+35 more)

### Community 40 - "diff.go"
Cohesion: 0.12
Nodes (35): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, ancestorReported(), blockLabel(), DiffBlocks() (+27 more)

### Community 41 - "Platform Selection"
Cohesion: 0.29
Nodes (12): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+4 more)

### Community 42 - "inject_test.go"
Cohesion: 0.15
Nodes (27): lineRole, serviceLine, shutdownStyle, svcKey, classifyServiceRest(), containsPortCommand(), describeKey(), fieldsUnquoted() (+19 more)

### Community 43 - ".releaseToBackup"
Cohesion: 0.18
Nodes (14): countContains(), field(), TestCountContains(), TestField(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), noReleaseActivityScript(), releaseActivityScript() (+6 more)

### Community 44 - "hasCall"
Cohesion: 0.20
Nodes (17): cliScriptNameFromDest(), hasCall(), TestDisableDefaultUsersShowVPNError(), TestReleaseToBackupReleasedTimeout(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching() (+9 more)

### Community 45 - "Echo"
Cohesion: 0.26
Nodes (4): interactiveFailRunner, Echo, Quote(), TestQuote()

### Community 46 - "k8s/inspect.go"
Cohesion: 0.16
Nodes (22): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), ownedPods(), splitWatch(), TestWatchFromContainersReadsEveryAllNamespacesSpelling() (+14 more)

### Community 47 - "replicationVPNLines"
Cohesion: 0.19
Nodes (16): ReplicationConfigResult, mateConvergenceShutdowns(), replicationVPNLines(), siteAEntry(), TestMateConvergenceShutdownsIsDeterministic(), TestNothingEverShutsDownTheVPNItself(), TestReplicationVPNLinesEnablesListedAndDisablesTheRest(), TestReplicationVPNLinesIsDeterministic() (+8 more)

### Community 48 - "localCfg"
Cohesion: 0.09
Nodes (36): TestLocalMateRunsThroughOps(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK(), TestBackupRevertActivitySuccess() (+28 more)

### Community 49 - "config_test.go"
Cohesion: 0.19
Nodes (24): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, span, AdditionalUser, Node, Redundancy (+16 more)

### Community 50 - "ParseBlocks"
Cohesion: 0.09
Nodes (27): Annotate(), StripMarkers(), TestAnnotateIsDeterministic(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), TestStripMarkersKeepsPragmaAndSectionComments(), TestStripMarkersNoMarkersPassthrough(), ParseBlocks() (+19 more)

### Community 51 - "newRootCmd"
Cohesion: 0.06
Nodes (41): TestAbbreviationDocs(), applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias() (+33 more)

### Community 52 - "exportconfig_test.go"
Cohesion: 0.10
Nodes (30): exportconfigReadCounter, runFailRunner, ExitCode(), childStatusError(), k8sEnv(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitStatusIsScopedToInteractiveSessions(), TestExitCodeContract() (+22 more)

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
Cohesion: 0.06
Nodes (25): AdminState, QueueState, ReplRole, scriptedMate, VPNRepl, MateChannel, sortedKeys(), reenableLine() (+17 more)

### Community 58 - "Ops"
Cohesion: 0.15
Nodes (9): TestValidName(), validCLILine(), validName(), Ops, domainCertsScript(), removeDomainCertsScript(), removeServerCertScript(), TestDomainCertsScriptSorted() (+1 more)

### Community 59 - ".rpc"
Cohesion: 0.10
Nodes (19): backupTarget, Credential, sempMate, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigLine(), Ops, httpBody() (+11 more)

### Community 60 - "confirm.go"
Cohesion: 0.23
Nodes (16): layer, TestPromptYes(), TestPromptYesNo(), TestStdinCanAnswerClosedFile(), confirmActionStrict(), confirmDowngrade(), confirmGate(), confirmLayer() (+8 more)

### Community 61 - ".configRows"
Cohesion: 0.13
Nodes (23): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminPassState(), adminSecretRow(), containsString() (+15 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.14
Nodes (38): TestManagerPrepHostLeavesBaseDirAlone(), failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+30 more)

### Community 63 - "strings.Builder"
Cohesion: 0.13
Nodes (22): WeightedNodeTerm, mdRow(), writeAbbrevTable(), writeResolutionNotes(), writeReadingNotes(), NodeAffinity, NodeMatchExpr, Placement (+14 more)

### Community 64 - "output_test.go"
Cohesion: 0.13
Nodes (23): TestRenderDiffResultDirtyReportNamesOffendingSectionAndLines(), TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), TestCheckReportSkipsEmptySections(), New(), NewFunc(), sink() (+15 more)

### Community 65 - "Configuration Guide"
Cohesion: 0.14
Nodes (14): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+6 more)

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

### Community 71 - "parse"
Cohesion: 0.10
Nodes (21): segment, vars, countMarkers(), kubeCommand(), resolvePlatform(), TestParseArrayElementsPreserveQuoting(), TestParseAssignmentForms(), TestParseCRLF() (+13 more)

### Community 72 - "age"
Cohesion: 0.24
Nodes (8): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownsPod(), TestOperatorRunningImageWithNoContainers(), TestOwnsPodFallsBackToTheNameInfix()

### Community 73 - "internal/k8s"
Cohesion: 0.10
Nodes (21): adminsecret_test.go, check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, generatorpage_test.go, inspect_test.go, internal/k8s (+13 more)

### Community 74 - "Capture"
Cohesion: 0.33
Nodes (5): Omission, Capture, parseHeader(), readHeaderField(), omittedList()

### Community 75 - "usagef"
Cohesion: 0.08
Nodes (25): childExitError, usageError, logArgs(), newConvertCmd(), runConvert(), asUsage(), childStatus(), isUsage() (+17 more)

### Community 76 - ".ConfigureReplication"
Cohesion: 0.24
Nodes (8): Ops, isRunCLIRejection(), newlineIf(), replPhase1Rejected(), replPhase2Rejected(), replTransportFailure(), replVPNList(), showVPNReplicationScript()

### Community 77 - "Command"
Cohesion: 0.18
Nodes (9): cmdField, guardedCmd, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandUnmarshal(), Command, guardCommandOf(), setGuardCommand() (+1 more)

### Community 78 - "newTeardownApplyFixture"
Cohesion: 0.31
Nodes (8): driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), newTeardownApplyFixture(), TestImportOpsImportApplySeparatesFirstSectionFromMain(), TestImportOpsImportApplyTearsDownExistingVPNBeforeApplyingArtifact(), TestTornDownRecordsOnlyAppliedTeardowns()

### Community 81 - "BrokerType"
Cohesion: 0.23
Nodes (8): BrokerType, cliMate, brokerTypeFromSchema(), TestBrokerTypeFromSchema(), bannerType(), checkSameType(), showReplicationScript(), TestTranscriptBannerType()

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

### Community 86 - "Kubernetes Inspect Tests"
Cohesion: 0.15
Nodes (19): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+11 more)

### Community 87 - "ParseShowReplication"
Cohesion: 0.15
Nodes (25): TestMateChannelShowReplication(), ParseShowReplication(), RenderMate(), SameEndpoints(), loadShowReplication(), mateAppliance(), mateSoftware(), TestMateRoundTripsThroughItsOwnGrammar() (+17 more)

### Community 88 - "operatorversion_test.go"
Cohesion: 0.29
Nodes (9): operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsSilentInAPreview(), TestConfirmNoDowngradeProceedsWhenConfirmAccepts(), TestConfirmNoDowngradeRefusedWhenConfirmDeclines(), TestConfirmNoDowngradeRefusesByDefault(), TestImageTag() (+1 more)

### Community 89 - "load"
Cohesion: 0.20
Nodes (18): boolStr(), BrokerCR(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets(), TestBrokerCRQuotesTheImageReference(), TestBrokerCRStatesRedundancyEitherWay(), TestCustomVolumeMountRendersTheCRArray() (+10 more)

### Community 90 - "vulnjudge/main.go"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "Import Section Rules"
Cohesion: 0.09
Nodes (33): Disposition, SectionRule, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified(), renderImportDocs() (+25 more)

### Community 92 - "step"
Cohesion: 0.14
Nodes (17): mateChannelFunc, App, lineSink(), progress(), step(), confirmReplicationConfig(), opConfigureReplication(), opCtrConfigureReplication() (+9 more)

### Community 93 - "Container"
Cohesion: 0.14
Nodes (14): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestDefaultK8sPortsMatchesOperator(), applyContainerBlockDefaults(), defaultK8sPorts() (+6 more)

### Community 94 - "Config"
Cohesion: 0.20
Nodes (6): expandTilde(), expandTildeToken(), Config, isPathSep(), TestExpandTilde(), tildeHome()

### Community 95 - "Secret Reference Resolution"
Cohesion: 0.50
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 96 - "github.com/spf13/cobra.Command"
Cohesion: 0.07
Nodes (40): shorthand, renderAbbrevDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeShorthandTable(), writeUnabbreviated(), anchor(), argumentLine() (+32 more)

### Community 97 - "replConfig"
Cohesion: 0.31
Nodes (9): replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationEmptyViaAccepted(), TestValidateReplicationRejects(), TestValidateReplicationValid() (+1 more)

### Community 98 - "Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "newBlock"
Cohesion: 0.28
Nodes (9): token, firstQuoted(), newBlock(), TestNewBlockMessageSpoolQualifierHasNoOwnName(), TestNewBlockOpener(), TestNewBlockVPNNameWithSpaces(), TestNewBlockVPNQualifierWithSpaces(), tokenizeOpener() (+1 more)

### Community 100 - ".checkUserManagerLimits"
Cohesion: 0.29
Nodes (6): delegateSetting(), Manager, limitText(), parseLimit(), parseUnitProps(), TestParseLimit()

### Community 102 - "Load"
Cohesion: 0.15
Nodes (27): minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadReadError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors() (+19 more)

### Community 103 - "container/secrets_test.go"
Cohesion: 0.17
Nodes (26): ResolveSecretValues(), bundleHash(), certCreate(), certFixture(), TestCertSecretLabelDrivesTheRestart(), TestCertSecretRemovalFailureIsFatal(), TestContainerBundleCarriesTheChain(), TestDeletePodmanToleratesAMissingBundle() (+18 more)

### Community 104 - "Platform"
Cohesion: 0.14
Nodes (27): checkFlagPlatforms(), commandPlatforms(), declaredList(), flagOnlyOn(), parsePlatformList(), platformSuffix(), prepare(), promptPlatform() (+19 more)

### Community 105 - "internal/cli"
Cohesion: 0.15
Nodes (13): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, euid_test.go, examples_test.go (+5 more)

### Community 106 - "captureStderr"
Cohesion: 0.09
Nodes (28): allowRuntime(), capture(), captureStderr(), fakeBinaryOnPath(), runStandalone(), runStatusStderr(), TestAnnounceCommandsNamesResolvedBinaries(), TestBinaryAnnouncementWiring() (+20 more)

### Community 107 - "Compose"
Cohesion: 0.20
Nodes (12): Compose(), composeEscape(), ComposeProject(), TestBridgePublishesOnlyTheListedPorts(), TestComposeProjectFoldsToComposesGrammar(), TestComposeProjectIsDeclaredNotDerived(), TestComposeQuotesTheCpuset(), TestContainerOverridesReachArtifact() (+4 more)

### Community 108 - "Import-Config Apply Rules"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "Config"
Cohesion: 0.05
Nodes (51): TLS, bridgeHostPort(), sempPort(), TestSempPortBridgeFallsBackToPlaintext(), TestSempPortBridgePrefersTLS(), TestSempPortBridgeRefusesNoMapping(), TestSempPortHostMode(), TestSempPortWithoutCertificateStaysPlaintext() (+43 more)

### Community 112 - "capRunner"
Cohesion: 0.11
Nodes (24): capCall, capRunner, New(), TestNewDefaults(), callIndex(), TestManagerLogsCLIShell(), TestManagerPrepHostRootlessUsesUnshareChown(), TestPreflightRunsBeforeAnything() (+16 more)

### Community 113 - "Config Package Tests"
Cohesion: 0.12
Nodes (17): adminsecret_test.go, artifactvalues_test.go, command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go (+9 more)

### Community 114 - "TestServerCert"
Cohesion: 0.16
Nodes (14): TestDiagnostics(), TestPathHelpers(), TestServerCert(), TestServerCertBundleOrder(), TestServerCertBundleRefusesAKeyBearingCert(), TestServerCertBundleReportsAnUnreadableFile(), TestServerCertBundleRequiresBothHalves(), ServerCertBundle() (+6 more)

### Community 115 - "operatorversion.go"
Cohesion: 0.24
Nodes (9): compareVersions(), Cluster, imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), operatorItem(), TestAmbiguousOperatorIsNotFoldedIntoNotInstalled() (+1 more)

### Community 116 - "EnvPairs"
Cohesion: 0.18
Nodes (13): EnvPairs(), groupKey(), itoa(), ServerCertBundlePath(), assertNoCheckoutPath(), envLines(), TestAdditionalUsersReachBothHalves(), TestGolden() (+5 more)

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.16
Nodes (14): containerState, healthState, decodeInspect(), Manager, orUnknown(), printable(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown() (+6 more)

### Community 119 - "podName"
Cohesion: 0.13
Nodes (7): rejectionIn(), BaseName(), TestBaseNameSplitsOnBothSeparators(), podName(), Cluster, shSingleQuote(), kubectlTransport

### Community 120 - "Get"
Cohesion: 0.13
Nodes (25): Example, TestEnvFlagCompletesEnvFiles(), TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), parseError(), DetectPlatforms(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint() (+17 more)

### Community 121 - "scripts.go"
Cohesion: 0.13
Nodes (21): showCmd, assertLeaderScript(), currentConfigScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), enableDefaultUsersScript(), enableDefaultVPNScript() (+13 more)

### Community 122 - "coverage_test.go"
Cohesion: 0.05
Nodes (57): cliRunNames(), matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError() (+49 more)

### Community 123 - ".validateContainerArtifactValues"
Cohesion: 0.20
Nodes (7): Config, portOrRange(), TestValidContainerPort(), validContainerPort(), validCoreLimit(), validHealthDuration(), validDNSLabel()

### Community 124 - "emitYAML"
Cohesion: 0.24
Nodes (8): doc, boolOf(), commentSafe(), emitYAML(), joinDomainCertPath(), redundancy(), scalar(), TestScalarQuoting()

### Community 125 - "App"
Cohesion: 0.40
Nodes (16): addCommands(), addLogFlags(), newOperatorCmd(), newOperatorDeployCmd(), newOperatorGenerateCmd(), newOperatorLogsCmd(), newOperatorRemoveCmd(), newOperatorRestartCmd() (+8 more)

### Community 126 - ".applyScalingTierDefaults"
Cohesion: 0.22
Nodes (7): scalingTier, containerMem(), Config, TestContainerMem(), TestScalingTiers(), TestTierForRejectsOffTierValues(), tierFor()

### Community 127 - ".validateContainer"
Cohesion: 0.15
Nodes (12): TestValidateContainerArtifactValues(), TestValidateProbeCommandAccepts(), TestValidateProbeCommandRejects(), cpuSetCount(), cpuSetRange(), TestCPUSetCount(), TestCPUSetRange(), invertedCPUSetRange() (+4 more)

### Community 128 - "Manager"
Cohesion: 0.21
Nodes (4): Manager, idMapCovers(), origin(), runUserIDs()

### Community 129 - "Scaling YAML Decoding"
Cohesion: 0.31
Nodes (5): scalingKey, scalingSpelling, Scaling, scalingKeyIndex(), scalingKeyList()

### Community 130 - "HARoles"
Cohesion: 0.14
Nodes (6): Cluster, Cluster, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes(), HARoles()

### Community 131 - ".preflightOne"
Cohesion: 0.50
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "ReplVia"
Cohesion: 0.33
Nodes (5): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, viaDescription()

### Community 134 - ".LeaderLocal"
Cohesion: 0.21
Nodes (8): TestLastLines(), TestLastLinesEqualCount(), defaultLocalAddrs(), Ops, hostMatches(), shortHost(), TestDefaultLocalAddrs(), lastLines()

### Community 135 - "scripts_test.go"
Cohesion: 0.23
Nodes (12): parseVPNNames(), productKeyScript(), productKeysScript(), removeProductKeysScript(), TestDisableDefaultUsersScriptQuoting(), TestParseVPNNames(), TestParseVPNNamesKeepsAMultiWordName(), TestParseVPNNamesNoSeparator() (+4 more)

### Community 136 - "ReplSite"
Cohesion: 0.23
Nodes (10): Replication, missingListedVPNs(), PlannedRoles(), RoleAtSite(), TestPlannedRolesNamesEveryListedVPN(), TestRoleAtSiteIsTheComplement(), replVPNNames(), ReplSite (+2 more)

### Community 137 - "ContainerSecrets"
Cohesion: 0.17
Nodes (12): ContainerSecrets(), containerSecretSpecs(), SecretPreflight(), TestContainerSecretNamesAreHostScoped(), TestContainerSecretsRedundancy(), TestSecretPreflight(), TestSecretsAndCertDoNotNest(), TestSecretTargetsAreAbsolutePaths() (+4 more)

### Community 138 - "Convert"
Cohesion: 0.25
Nodes (7): Result, DecodeStrict(), Convert(), TestConvertUnterminatedArray(), TestGeneratedHeader(), TestGeneratedHeaderSanitisesSource(), validateOutput()

### Community 139 - "Env Path Resolution"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 140 - "parsePort"
Cohesion: 0.25
Nodes (8): LoadBalancer, cut(), parsePort(), splitUser(), TestParsePort(), writeKeyValueEntry(), writeLBAnnotations(), portSpec

### Community 141 - "Duration and Host Path Checks"
Cohesion: 0.40
Nodes (5): CanonicalDuration(), TestCanonicalDuration(), CheckHostPath(), TestCheckHostPathAccepts(), TestCheckHostPathRejects()

### Community 142 - "storageCfg"
Cohesion: 0.25
Nodes (7): storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestStorageClassAndCustomMountAreMutuallyExclusive()

### Community 153 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 155 - ".gatherNode"
Cohesion: 0.33
Nodes (4): gatherConfigsScript(), TestGatherConfigsScript(), TestZipConfigsScript(), zipConfigsScript()

### Community 156 - "TestImportOpsExportConfigScopeSelectsCLICommand"
Cohesion: 0.33
Nodes (6): minimalCapture(), targetVPNCapture(), TestImportOpsExportConfigScopeSelectsCLICommand(), TestImportOpsImportPlanRefusesCrossTypeAndUnknownType(), TestImportOpsImportPlanRefusesRedactedArtifact(), TestImportOpsImportPlanSplitsExistingAndNewVPNs()

### Community 157 - "euid_test.go"
Cohesion: 0.67
Nodes (5): runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand(), TestPodmanEUIDGuardSkips(), writeRootlessPodmanEnv()

### Community 158 - "internal/engine"
Cohesion: 0.67
Nodes (3): internal/engine, resolve_test.go, runner_test.go

## Knowledge Gaps
- **238 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+233 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 359 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **19 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Manager` connect `Manager` to `context.Context`, `App`, `Platform`, `Config`, `config_test.go`, `confirm.go`, `newEchoMgr`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `newTestOps`, `HARoles`, `.LeaderLocal`, `Manager`, `K8sConfig`, `Abbreviation Sets`, `renderDriver`, `.gatherNode`, `Import Plan and Apply`, `prep_test.go`, `Config`, `App`, `.releaseToBackup`, `config_test.go`, `ParseBlocks`, `Ops`, `.rpc`, `Ops`, `Mate Channel Tests`, `usagef`, `.ConfigureReplication`, `Container File Transport`, `BrokerType`, `k8s/matechannel_test.go`, `step`, `github.com/spf13/cobra.Command`, `Config`, `podName`, `scripts.go`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `bg`, `newTestOps`, `HARoles`, `Manager`, `ReplSite`, `K8sConfig`, `ContainerSecrets`, `captureStdout`, `NewCluster`, `Cluster`, `convert_test.go`, `Quadlet`, `manager_test.go`, `GenSecrets`, `prep_test.go`, `newEchoMgr`, `ResolveDomainCerts`, `recRunner`, `localCfg`, `config_test.go`, `Cluster`, `.configRows`, `strings.Builder`, `Ops`, `Container File Transport`, `k8s/matechannel_test.go`, `occCluster`, `load`, `step`, `Container`, `container/secrets_test.go`, `Platform`, `captureStderr`, `Compose`, `Image`, `capRunner`, `TestServerCert`, `EnvPairs`, `podName`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Are the 49 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 49 INFERRED edges - model-reasoned connections that need verification._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _238 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `bg` be split into smaller, more focused modules?**
  _Cohesion score 0.11689291101055807 - nodes in this community are weakly interconnected._
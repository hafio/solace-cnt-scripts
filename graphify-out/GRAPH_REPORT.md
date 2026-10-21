# Graph Report - solace-cnt-scripts  (2026-10-21)

## Corpus Check
- 200 files · ~688,858 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 4090 nodes · 17488 edges · 157 communities (141 shown, 16 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 2552 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `b58d1902`
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
- runPlatform
- CheckCommand
- opArgvMatch
- NewCluster
- ParseBlocks
- convert_test.go
- transform_test.go
- manager_test.go
- commands.go
- matchCLI
- Bash Dev Script
- parse
- PowerShell Dev Script
- Set
- BuildSwitchPlan
- newLocalOps
- runner_test.go
- renderDriver
- completion_test.go
- ImportPlan
- deploymentItem
- Solace Module Root
- Config
- Docs and Legacy Scripts
- ParseVPNReplication
- ResolveDomainCerts
- Ops
- App
- recRunner
- annotate_test.go
- twoMates
- diff_test.go
- EnvPairs
- inject_test.go
- Ops
- runRootWith
- Echo
- eqArgs
- VPNRepl
- localCfg
- config_test.go
- prep_test.go
- newRootCmd
- .Run
- Container Security Guidelines
- Operations
- Test catalogue
- Cluster
- replNewRig
- runExport
- .rpc
- confirm.go
- .configRows
- rootless_test.go
- ctrManager
- output_test.go
- Configuration Guide
- opPerformReplication
- newTestMate
- Command Reference Doc
- Developer guide
- replApp
- TestServerCert
- .applyScalingTierDefaults
- internal/k8s
- Docker and Podman mechanics
- usagef
- Compose
- Command
- newTeardownApplyFixture
- normalizeToList
- ReplSite
- k8s/matechannel_test.go
- limits_test.go
- Removing a broker: what stays, what goes
- Troubleshooting
- k8s/inspect_test.go
- .ConfigureReplication
- AdminSecret
- render_test.go
- vulnjudge/main.go
- RuleFor
- emitYAML
- Container
- Config
- Secret Reference Resolution
- MateChannel
- replConfig
- Abbreviation Reference
- age
- .checkUserManagerLimits
- .Preflight
- Load
- ParseRole
- Platform
- internal/cli
- k8s/inspect.go
- operatorversion_test.go
- Import-Config Apply Rules
- Config
- internal/broker
- GenOperator
- capRunner
- Config Package Tests
- withLong
- .operatorRows
- certFixture
- internal/container
- Container State Inspection
- RenderDiffResult
- Ops
- scripts.go
- Quadlet
- .validateContainerArtifactValues
- writeNode
- currentConfigScript
- bytes.Buffer
- allowcommand_test.go
- guardConfig
- Scaling YAML Decoding
- Role
- .preflightOne
- ReplVia
- Namespace Protection Checks
- Cluster
- containerTransport
- runLeader
- Fixtures and doubles
- ReadAdminPassword
- newBlock
- sempPort
- CheckHostPath
- Convert
- TestShowRedundancyIsReadOnly
- TestExecute
- github.com/spf13/cobra.Command
- ResolveEnvPath

## God Nodes (most connected - your core abstractions)
1. `newTestOps()` - 163 edges
2. `Role` - 131 edges
3. `ctrCfg()` - 131 edges
4. `NewCluster()` - 128 edges
5. `Config` - 113 edges
6. `newCapMgr()` - 111 edges
7. `loadK8s()` - 110 edges
8. `matchCLI()` - 95 edges
9. `Platform` - 79 edges
10. `bg()` - 70 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `ExitCode()`  [EXTRACTED]
  main.go → internal/cli/exit.go
- `assertFencesBalanced()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/annotate.go
- `parseBody()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `parseHeader()` --calls--> `readMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `applyMetaField()` --calls--> `BrokerType`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (157 total, 16 thin omitted)

### Community 0 - "bg"
Cohesion: 0.12
Nodes (49): bg(), k8sAdminPassword(), k8sCluster(), k8sContext(), k8sLogin(), k8sOps(), k8sRestartForCertificate(), k8sWhat() (+41 more)

### Community 1 - "newTestOps"
Cohesion: 0.05
Nodes (67): countContains(), leaderCalls(), newTestOps(), TestCountContains(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultUsersRefusesAQuotedVPNName(), TestDisableDefaultVPN() (+59 more)

### Community 2 - "Sink"
Cohesion: 0.20
Nodes (5): solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), columnWidths(), Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.06
Nodes (13): Exec, Transport, Manager, idMapCovers(), origin(), runUserIDs(), EnvRunner, Runner (+5 more)

### Community 4 - "cli_test.go"
Cohesion: 0.07
Nodes (64): allowRuntime(), capture(), captureStderr(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone(), runStatusStderr() (+56 more)

### Community 5 - "testing.T"
Cohesion: 0.01
Nodes (207): TestExecShellScriptUploadFailureRunsNothing(), TestDiagnosticsRunError(), TestDomainCertsUploadFileError(), TestExecCLIFailsOnErrorOutput(), TestExecCLIUploadFileError(), TestGatherNodeDownloadError(), TestOutDefaultsToStdout(), TestPollCondError() (+199 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (8): certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing()

### Community 8 - "K8sConfig"
Cohesion: 0.10
Nodes (15): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, LoadBalancer, cut() (+7 more)

### Community 9 - "runPlatform"
Cohesion: 0.23
Nodes (14): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+6 more)

### Community 10 - "CheckCommand"
Cohesion: 0.15
Nodes (22): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+14 more)

### Community 11 - "opArgvMatch"
Cohesion: 0.07
Nodes (65): opCall, opRunner, ReplicationConfigResult, captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook() (+57 more)

### Community 12 - "NewCluster"
Cohesion: 0.07
Nodes (67): TestCheckStopsProbingWhenUnreachable(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+59 more)

### Community 13 - "ParseBlocks"
Cohesion: 0.08
Nodes (29): BrokerType, isMarker(), StripMarkers(), TestIsMarkerDiscriminatesNamespace(), TestStripMarkersKeepsPragmaAndSectionComments(), TestStripMarkersNoMarkersPassthrough(), brokerTypeFromSchema(), Capture (+21 more)

### Community 14 - "convert_test.go"
Cohesion: 0.19
Nodes (31): checkGolden(), convertOK(), hasWarning(), strictDecode(), TestConvertAdminSecretAlias(), TestConvertAdminUserIsDroppedOnEveryPlatform(), TestConvertBadBooleanWarns(), TestConvertBadNumberWarns() (+23 more)

### Community 15 - "transform_test.go"
Cohesion: 0.08
Nodes (62): Block, TargetState, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects(), ClearTargetVirtualHostnames() (+54 more)

### Community 16 - "manager_test.go"
Cohesion: 0.06
Nodes (115): inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer() (+107 more)

### Community 17 - "commands.go"
Cohesion: 0.21
Nodes (44): opFunc, roleOpFunc, addCommands(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerConfigureCmd() (+36 more)

### Community 18 - "matchCLI"
Cohesion: 0.07
Nodes (50): cliScriptNameFromDest(), hasCall(), matchCLI(), TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestDisableDefaultUsersDisableError(), TestDisableDefaultUsersShowVPNError(), TestDisableDefaultVPNDisableError() (+42 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "parse"
Cohesion: 0.11
Nodes (20): segment, vars, countMarkers(), resolvePlatform(), TestParseArrayElementsPreserveQuoting(), TestParseAssignmentForms(), TestParseCRLF(), TestParseDoubleQuotedBackslashEscapes() (+12 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Set"
Cohesion: 0.13
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 23 - "BuildSwitchPlan"
Cohesion: 0.15
Nodes (23): PhaseKind, SiteState, SwitchAction, SwitchPhase, BuildSwitchPlan(), confirmDemotions(), SwitchPlan, promoteConfirmed() (+15 more)

### Community 24 - "newLocalOps"
Cohesion: 0.09
Nodes (41): ranContains(), TestDiagnostics(), TestDisableDefaultVPNShowError(), TestLeaderAssertLeaderError(), TestRedundancyBackupShowError(), TestRedundancyReleaseError(), TestRedundancyRevertToPrimaryError(), TestReleaseToBackupNoReleaseError() (+33 more)

### Community 25 - "runner_test.go"
Cohesion: 0.09
Nodes (31): TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), MaskEnv(), NewExec() (+23 more)

### Community 26 - "renderDriver"
Cohesion: 0.14
Nodes (19): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce(), sanitizeComment() (+11 more)

### Community 27 - "completion_test.go"
Cohesion: 0.18
Nodes (21): driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion(), TestCompletionDescriptionsAreOptIn() (+13 more)

### Community 28 - "ImportPlan"
Cohesion: 0.12
Nodes (22): ImportResult, PlannedSection, isPragma(), defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList() (+14 more)

### Community 29 - "deploymentItem"
Cohesion: 0.13
Nodes (17): containsString(), installSteps(), desiredWatch(), Cluster, splitWatch(), subtractWatch(), TestDesiredWatchAppendsTheBrokerNamespace(), TestSubtractWatchEmptyRemainingMeansDelete() (+9 more)

### Community 31 - "Config"
Cohesion: 0.11
Nodes (15): keyValueEntries, TestValidateContainerArtifactValues(), TestRoleNames(), RoleNames(), foldToEnvVar(), Config, invertedCPUSetRange(), missingErr() (+7 more)

### Community 32 - "Docs and Legacy Scripts"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "ParseVPNReplication"
Cohesion: 0.10
Nodes (23): colSpan, dashSpans(), flagByte(), gutterClear(), ParseVPNReplication(), sliceSpan(), TestGutterClearAndSliceSpanBoundaries(), TestParseVPNReplication() (+15 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (41): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domainCANames(), caNameSafe(), checkNoDotDot() (+33 more)

### Community 35 - "Ops"
Cohesion: 0.09
Nodes (12): cliRunNames(), Ops, runCLISkeleton(), TestValidName(), validName(), ValidScriptName(), shQuote(), TestShQuoteHandlesASingleQuote() (+4 more)

### Community 36 - "App"
Cohesion: 0.17
Nodes (27): confirmAction(), wantEnable(), wantRemove(), confirmAssertFromBackup(), containerRenderRole(), containerRole(), containerWhat(), ctrLogin() (+19 more)

### Community 37 - "recRunner"
Cohesion: 0.11
Nodes (13): TestCanIAnswerReadsTheLastLine(), TestExecutorRefusesUnapprovedRuntime(), unapprovedCfg(), TestBrokerReportPreviewStopsAtAnUnapprovedCommand(), NewTransport(), canIVerb(), TestTransportCopy(), TestTransportExecArgs() (+5 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.06
Nodes (53): Region, Annotate(), applyMetaField(), markerRegion(), markerSection(), markerVerb(), parseKV(), readMarker() (+45 more)

### Community 39 - "twoMates"
Cohesion: 0.23
Nodes (20): ExecuteSwitchPlan(), SwitchPreflight(), ordersToA(), replSites(), TestExecuteSwitchPlanEmptyDoesNothing(), TestExecuteSwitchPlanOrder(), TestExecuteSwitchPlanRefusesASiteItHasNoChannelFor(), TestExecuteSwitchPlanRefusesToPromoteAgainstStaleState() (+12 more)

### Community 40 - "diff_test.go"
Cohesion: 0.15
Nodes (25): blockKey, mergedBlock, mergedLine, ancestorReported(), DiffBlocks(), indexByIdentity(), keyOf(), keySet() (+17 more)

### Community 41 - "EnvPairs"
Cohesion: 0.12
Nodes (16): Config, NodeIdentity, containerSecretSpecs(), EnvPairs(), groupKey(), itoa(), ServerCertBundlePath(), assertNoCheckoutPath() (+8 more)

### Community 42 - "inject_test.go"
Cohesion: 0.15
Nodes (27): lineRole, serviceLine, shutdownStyle, svcKey, classifyServiceRest(), containsPortCommand(), describeKey(), fieldsUnquoted() (+19 more)

### Community 43 - "Ops"
Cohesion: 0.11
Nodes (18): field(), TestField(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), gatherConfigsScript(), showRedundancyLocalScript(), TestGatherConfigsScript(), defaultLocalAddrs() (+10 more)

### Community 44 - "runRootWith"
Cohesion: 0.09
Nodes (42): runFailRunner, runRootWith(), TestDeployOperatorNoPromptStaysUnknownFlag(), TestK8sConfigDeleteDomainCertsConfigured(), TestK8sConfigDeleteDomainCertsFromDirs(), TestK8sConfirmDeclined(), TestK8sPromptsNameNamespaceAndContext(), TestK8sPromptsOmitAnUnknownContext() (+34 more)

### Community 45 - "Echo"
Cohesion: 0.26
Nodes (4): interactiveFailRunner, Echo, Quote(), TestQuote()

### Community 46 - "eqArgs"
Cohesion: 0.08
Nodes (36): TestReachable(), newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), saCfg(), TestHARoles() (+28 more)

### Community 47 - "VPNRepl"
Cohesion: 0.10
Nodes (9): AdminState, QueueState, scriptedMate, replMate, ReplRole, VPNRepl, reenableLine(), sempAdminState() (+1 more)

### Community 48 - "localCfg"
Cohesion: 0.08
Nodes (44): TestLocalMateRunsThroughOps(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK(), TestBackupRevertActivitySuccess() (+36 more)

### Community 49 - "config_test.go"
Cohesion: 0.20
Nodes (20): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, span, dropMessageSpoolLines(), dropRoutingInterfaceLine(), TestSectionsDropMessageSpoolLines() (+12 more)

### Community 50 - "prep_test.go"
Cohesion: 0.06
Nodes (57): TestResolveStorageClass(), TestStorageRows(), TestDeleteBrokerProbesClaimsOnlyWhenPurging(), haCfg(), TestOperatorScale(), TestOperatorScaleStopsAtARefusedPermission(), TestOperatorScaleSurfacesAFailure(), GenBroker() (+49 more)

### Community 51 - "newRootCmd"
Cohesion: 0.10
Nodes (26): TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing() (+18 more)

### Community 52 - ".Run"
Cohesion: 0.11
Nodes (34): echoRunner(), runCtr(), TestAssertLeaderStandaloneSkipsBeforeAnyPrompt(), TestConfigStepsDoNotLeakSecrets(), TestConfiguredRouternameSurvivesTheFallback(), TestCtrConfigDryRun(), TestCtrConfirmDeclined(), TestCtrDiagnosticsDryRun() (+26 more)

### Community 53 - "Container Security Guidelines"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 54 - "Operations"
Cohesion: 0.10
Nodes (21): Bringing up a fresh cluster, Broker scope is a fixed classification, Configuring a site, Data replication, Exit codes, Exporting and importing configuration, Extra CLI users differ by platform, Handling the artifact (+13 more)

### Community 55 - "Test catalogue"
Cohesion: 0.10
Nodes (20): abbrev_test.go, convert_test.go, Coverage, examples_test.go, internal/abbrev, internal/convert, internal/engine, internal/examples (+12 more)

### Community 57 - "replNewRig"
Cohesion: 0.20
Nodes (18): replRig, replNewRig(), replUp(), TestOpConfigureReplicationAppliesTheRoleTheBrokersSiteDerives(), TestOpConfigureReplicationConvergesTheMateAndReportsWhatItStopped(), TestOpConfigureReplicationDeclinedLeavesTheBrokerUntouched(), TestOpConfigureReplicationFailureReportsNothingApplied(), TestOpConfigureReplicationRefusesAnUnknownBrokerBeforeAsking() (+10 more)

### Community 58 - "runExport"
Cohesion: 0.13
Nodes (14): confirmImport(), pluralVPN(), runExport(), runImport(), TestPluralVPNWording(), emitOrWrite(), App, resolveScriptPath() (+6 more)

### Community 59 - ".rpc"
Cohesion: 0.12
Nodes (16): backupTarget, Credential, sempMate, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigLine(), Ops, httpBody() (+8 more)

### Community 60 - "confirm.go"
Cohesion: 0.17
Nodes (21): exportconfigReadCounter, layer, TestConfirmDeleteShortcut(), TestPromptYes(), TestPromptYesNo(), TestStdinCanAnswerClosedFile(), confirmActionStrict(), confirmDelete() (+13 more)

### Community 61 - ".configRows"
Cohesion: 0.11
Nodes (27): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminPassState(), adminSecretRow(), allCustomMounted() (+19 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.17
Nodes (35): failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent(), TestCheckIDMappingRefusesAnUnmappedRunUser() (+27 more)

### Community 63 - "ctrManager"
Cohesion: 0.12
Nodes (19): TestCtrManagerConfirmWiring(), logArgs(), childExit(), TestChildExitKeepsItsMessage(), TestExitCodeIsNeverNegative(), ctrManager(), opCtrCheck(), opCtrCLI() (+11 more)

### Community 64 - "output_test.go"
Cohesion: 0.12
Nodes (24): TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportPlanReportShowsOmissionsAndEachSectionsVerdict(), TestImportResultReportQuotesTheFailedChunkAndItsTail(), TestImportResultReportsTeardownAndRebuildSeparately(), TestCheckReportSkipsEmptySections(), New(), NewFunc() (+16 more)

### Community 65 - "Configuration Guide"
Cohesion: 0.14
Nodes (14): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+6 more)

### Community 66 - "opPerformReplication"
Cohesion: 0.14
Nodes (19): mateChannelFunc, App, lineSink(), progress(), step(), confirmReplicationConfig(), confirmReplicationSwitch(), opConfigureReplication() (+11 more)

### Community 67 - "newTestMate"
Cohesion: 0.16
Nodes (21): CLIRunner, fakeRun, NewCLIMate(), mateReplies(), newTestMate(), TestMateChannelCallsKeepTheCauseWhenTheRunnerFails(), TestMateChannelDescribeNamesTheTarget(), TestMateChannelPreflightLearnsTheBrokerType() (+13 more)

### Community 68 - "Command Reference Doc"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.20
Nodes (10): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+2 more)

### Community 70 - "replApp"
Cohesion: 0.22
Nodes (15): mateChannel(), mateSEMPPassword(), replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer(), TestConfirmReplicationSwitchNamesEveryMoveAndIsStrict(), TestMateChannelPicksTheMechanismTheSiteDeclares() (+7 more)

### Community 71 - "TestServerCert"
Cohesion: 0.14
Nodes (17): TestDomainCerts(), TestPathHelpers(), TestServerCert(), writeFile(), TestDomainCertsLoadsInSortedNameOrder(), TestDiagnosticsMkdirError(), TestServerCertCAReadError(), TestServerCertRunCLIError() (+9 more)

### Community 72 - ".applyScalingTierDefaults"
Cohesion: 0.15
Nodes (11): scalingTier, containerMem(), cpuSetCount(), cpuSetRange(), Config, TestContainerMem(), TestCPUSetCount(), TestCPUSetRange() (+3 more)

### Community 73 - "internal/k8s"
Cohesion: 0.10
Nodes (21): adminsecret_test.go, check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, generatorpage_test.go, inspect_test.go, internal/k8s (+13 more)

### Community 74 - "Docker and Podman mechanics"
Cohesion: 0.40
Nodes (5): Docker and Podman mechanics, Re-deploying is safe and explicit, Rootless podman prerequisites, What the broker cannot gain, Which user the broker runs as

### Community 75 - "usagef"
Cohesion: 0.07
Nodes (32): childExitError, usageError, ValidVPNName(), validVPNName(), addLogFlags(), noRolePositional(), registerFlagCompletion(), newConvertCmd() (+24 more)

### Community 76 - "Compose"
Cohesion: 0.15
Nodes (14): Compose(), composeEscape(), ComposeProject(), ContainerSecrets(), ContainerSecret, secretFilePath(), SecretPreflight(), TestComposeProjectFoldsToComposesGrammar() (+6 more)

### Community 77 - "Command"
Cohesion: 0.13
Nodes (12): cmdField, guardedCmd, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandUnmarshal(), TestValidateProbeCommandAccepts(), TestValidateProbeCommandRejects(), Command (+4 more)

### Community 78 - "newTeardownApplyFixture"
Cohesion: 0.11
Nodes (24): planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), newTeardownApplyFixture(), TestBuildChunksCreatesVPNsBeforeBrokerSections() (+16 more)

### Community 79 - "normalizeToList"
Cohesion: 0.20
Nodes (5): Cluster, Cluster, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

### Community 81 - "ReplSite"
Cohesion: 0.13
Nodes (22): Replication, sortedKeys(), missingListedVPNs(), PlannedRoles(), replicationVPNLines(), RoleAtSite(), sortedVPNs(), siteAEntry() (+14 more)

### Community 82 - "k8s/matechannel_test.go"
Cohesion: 0.23
Nodes (25): execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf(), replCfg(), replSite() (+17 more)

### Community 83 - "limits_test.go"
Cohesion: 0.13
Nodes (27): missingControllers(), limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessCoreFloorFollowsTheConfiguredLimit(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager() (+19 more)

### Community 84 - "Removing a broker: what stays, what goes"
Cohesion: 0.40
Nodes (5): Every destructive command confirms, Removing a broker: what stays, what goes, Removing the operator does not always remove it, The layer flag raises the question, The namespace is only offered when it is empty

### Community 85 - "Troubleshooting"
Cohesion: 0.22
Nodes (9): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Forbidden: cannot list deployments at the cluster scope, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting (+1 more)

### Community 86 - "k8s/inspect_test.go"
Cohesion: 0.15
Nodes (19): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+11 more)

### Community 87 - ".ConfigureReplication"
Cohesion: 0.05
Nodes (59): cliMate, TestMateChannelShowReplication(), cliTransport(), containsEndpoint(), MateConfig, labelValue(), normTransport(), parseHostPort() (+51 more)

### Community 88 - "AdminSecret"
Cohesion: 0.12
Nodes (18): AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), checkGolden(), decodeDataValue(), TestAdditionalUsersSecretCarriesBothHalves(), TestAdditionalUsersSecretIsAbsentWithNoUsers(), TestAdditionalUsersSecretRefusesAnEmptyPassword() (+10 more)

### Community 89 - "render_test.go"
Cohesion: 0.12
Nodes (37): boolStr(), BrokerCR(), containerArtifacts(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets(), TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPrivilegePosture() (+29 more)

### Community 90 - "vulnjudge/main.go"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.07
Nodes (38): Disposition, Omission, SectionRule, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+30 more)

### Community 92 - "emitYAML"
Cohesion: 0.22
Nodes (9): doc, boolOf(), commentSafe(), emitYAML(), joinDomainCertPath(), kubeCommand(), redundancy(), scalar() (+1 more)

### Community 93 - "Container"
Cohesion: 0.17
Nodes (12): Container, DockerConfig, Network, PodmanConfig, Ulimits, applyContainerBlockDefaults(), Config, podmanRunUser() (+4 more)

### Community 94 - "Config"
Cohesion: 0.20
Nodes (6): expandTilde(), expandTildeToken(), Config, isPathSep(), TestExpandTilde(), tildeHome()

### Community 95 - "Secret Reference Resolution"
Cohesion: 0.50
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 96 - "MateChannel"
Cohesion: 0.16
Nodes (7): MateChannel, Ops, curlConfigFlag(), Ops, requirePrimaryActive(), requireVirtualRouterNamesMatch(), sortedSiteNames()

### Community 97 - "replConfig"
Cohesion: 0.29
Nodes (10): replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationEmptyViaAccepted(), TestValidateReplicationPassSecretGrammar(), TestValidateReplicationRejects() (+2 more)

### Community 98 - "Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "age"
Cohesion: 0.30
Nodes (6): orNone(), age(), roleRank(), Cluster, ownsPod(), TestOwnsPodFallsBackToTheNameInfix()

### Community 100 - ".checkUserManagerLimits"
Cohesion: 0.29
Nodes (6): delegateSetting(), Manager, limitText(), parseLimit(), parseUnitProps(), TestParseLimit()

### Community 102 - "Load"
Cohesion: 0.07
Nodes (53): Example, TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadReadError() (+45 more)

### Community 103 - "ParseRole"
Cohesion: 0.50
Nodes (5): TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), ParseRole(), RoleAbbrev()

### Community 104 - "Platform"
Cohesion: 0.12
Nodes (29): checkFlagPlatforms(), commandPlatforms(), declaredList(), flagOnlyOn(), parsePlatformList(), platformSuffix(), prepare(), promptPlatform() (+21 more)

### Community 105 - "internal/cli"
Cohesion: 0.14
Nodes (14): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, euid_test.go, examples_test.go (+6 more)

### Community 106 - "k8s/inspect.go"
Cohesion: 0.17
Nodes (19): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), operatorRunningImage(), ownedPods(), TestOperatorRunningImageWithNoContainers() (+11 more)

### Community 107 - "operatorversion_test.go"
Cohesion: 0.15
Nodes (16): TestValidateUnderADeniedClusterWideListReadsTheOperatorNamespace(), TestOperatorApplyProceedsWhenTheClusterWideListIsDenied(), TestOperatorApplyRefusesAFallbackFirstInstallWithoutEscalateOrBind(), operatorDeployJSON(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsNotSkippedByAFailedRead(), TestConfirmNoDowngradeIsSilentInAPreview(), TestConfirmNoDowngradeProceedsWhenConfirmAccepts() (+8 more)

### Community 108 - "Import-Config Apply Rules"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "Config"
Cohesion: 0.09
Nodes (12): AdditionalUser, Image, Node, Redundancy, SEMP, TLS, atoiPrefix(), Config (+4 more)

### Community 110 - "internal/broker"
Cohesion: 0.08
Nodes (25): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, config_ops_test.go, coverage_test.go, diff_test.go, driver_test.go (+17 more)

### Community 111 - "GenOperator"
Cohesion: 0.08
Nodes (33): replShowMate(), replVPNTable(), TestOperatorInstallVerdict(), TestValidateSparseConfigExplainsItself(), GenOperator(), joinYAMLDocs(), nonEmpty(), operatorEscalationProbes() (+25 more)

### Community 112 - "capRunner"
Cohesion: 0.16
Nodes (14): capCall, capRunner, New(), TestNewDefaults(), NewTransport(), dockerCfg(), podmanCfg(), TestTransportCopy() (+6 more)

### Community 113 - "Config Package Tests"
Cohesion: 0.12
Nodes (17): adminsecret_test.go, artifactvalues_test.go, command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go (+9 more)

### Community 114 - "withLong"
Cohesion: 0.42
Nodes (14): newOperatorCmd(), newOperatorDeployCmd(), newOperatorGenerateCmd(), newOperatorLogsCmd(), newOperatorRemoveCmd(), newOperatorRestartCmd(), newOperatorStartCmd(), newOperatorStatusCmd() (+6 more)

### Community 115 - ".operatorRows"
Cohesion: 0.23
Nodes (10): compareVersions(), Cluster, imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestImageTag() (+2 more)

### Community 116 - "certFixture"
Cohesion: 0.09
Nodes (31): callIndex(), TestPreflightRunsBeforeAnything(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime(), TestManagerReachableProbesRuntimeThenCompose(), withWrapper(), wrappedCtrCfg() (+23 more)

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.16
Nodes (14): containerState, healthState, decodeInspect(), Manager, orUnknown(), printable(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown() (+6 more)

### Community 119 - "RenderDiffResult"
Cohesion: 0.23
Nodes (11): BlockDiff, DiffResult, blockLabel(), diffRow(), orDash(), RenderDiffResult(), reportBlockLines(), TestDiffBlockLabelFormatsIdentityForTheReport() (+3 more)

### Community 120 - "Ops"
Cohesion: 0.22
Nodes (4): validCLILine(), Ops, disableDefaultVPNScript(), TestDisableDefaultVPNScript()

### Community 121 - "scripts.go"
Cohesion: 0.10
Nodes (30): showCmd, TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), archiveConfigsScript(), assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), enableDefaultUsersScript() (+22 more)

### Community 122 - "Quadlet"
Cohesion: 0.20
Nodes (14): ContainerNoFile(), HealthCheck, TestLimitsCheckAssertsWhatTheArtifactAsks(), escapePercent(), healthCmd(), Quadlet(), quadletEscape(), healthCheckFixture() (+6 more)

### Community 123 - ".validateContainerArtifactValues"
Cohesion: 0.14
Nodes (11): Config, portOrRange(), TestValidContainerPort(), validContainerPort(), validCoreLimit(), validHealthDuration(), CheckSecretKeyRef(), validDNSLabel() (+3 more)

### Community 124 - "writeNode"
Cohesion: 0.21
Nodes (12): WeightedNodeTerm, NodeAffinity, NodeMatchExpr, Placement, validateMatchExprs(), validatePlacementAffinity(), parseToleration(), TestParseToleration() (+4 more)

### Community 125 - "currentConfigScript"
Cohesion: 0.18
Nodes (10): minimalCapture(), targetVPNCapture(), TestImportOpsExportConfigScopeSelectsCLICommand(), TestImportOpsImportPlanRefusesCrossTypeAndUnknownType(), TestImportOpsImportPlanRefusesRedactedArtifact(), TestImportOpsImportPlanReportsATargetItCannotRead(), TestImportOpsImportPlanSplitsExistingAndNewVPNs(), TestImportOpsImportVerifyReExportsAndDiffs() (+2 more)

### Community 126 - "bytes.Buffer"
Cohesion: 0.09
Nodes (24): occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants(), TestNamespaceIsProtectedSharesDeleteNamespacesList() (+16 more)

### Community 127 - "allowcommand_test.go"
Cohesion: 0.29
Nodes (10): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+2 more)

### Community 128 - "guardConfig"
Cohesion: 0.18
Nodes (11): guardConfig(), IsAbsHostPath(), TestContainerCertRequiresKey(), TestContainerHostDirsExpandATilde(), TestDataDirMustBeAbsolute(), TestIsAbsHostPath(), TestPodmanBaseDirIsOptionalButAbsolute(), TestQuadletDirAndDataDirAreNotRebased() (+3 more)

### Community 129 - "Scaling YAML Decoding"
Cohesion: 0.31
Nodes (5): scalingKey, scalingSpelling, Scaling, scalingKeyIndex(), scalingKeyList()

### Community 130 - "Role"
Cohesion: 0.06
Nodes (29): covUploadFailsTo, downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile (+21 more)

### Community 131 - ".preflightOne"
Cohesion: 0.47
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "ReplVia"
Cohesion: 0.20
Nodes (9): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, checkCredentialChars(), sortedKeys(), validateReplEndpoints(), validateReplVia() (+1 more)

### Community 134 - "Cluster"
Cohesion: 0.22
Nodes (3): Cluster, namespaceManifest(), ownedSecretNames()

### Community 136 - "runLeader"
Cohesion: 0.60
Nodes (4): leaderRun, runLeader(), TestAssertLeaderAsksBeforeActingFromABackup(), TestAssertLeaderPrimaryAndMonitorNeverAsk()

### Community 137 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 138 - "ReadAdminPassword"
Cohesion: 0.27
Nodes (10): AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestReadAdminPasswordArgvAndDecoding(), TestReadAdminPasswordGuardsTheCommand(), TestReadAdminPasswordNamesTheNextStep() (+2 more)

### Community 139 - "newBlock"
Cohesion: 0.28
Nodes (9): token, firstQuoted(), newBlock(), TestNewBlockMessageSpoolQualifierHasNoOwnName(), TestNewBlockOpener(), TestNewBlockVPNNameWithSpaces(), TestNewBlockVPNQualifierWithSpaces(), tokenizeOpener() (+1 more)

### Community 140 - "sempPort"
Cohesion: 0.29
Nodes (7): bridgeHostPort(), sempPort(), TestSempPortBridgeFallsBackToPlaintext(), TestSempPortBridgePrefersTLS(), TestSempPortBridgeRefusesNoMapping(), TestSempPortHostMode(), TestSempPortWithoutCertificateStaysPlaintext()

### Community 141 - "CheckHostPath"
Cohesion: 0.40
Nodes (5): CanonicalDuration(), TestCanonicalDuration(), CheckHostPath(), TestCheckHostPathAccepts(), TestCheckHostPathRejects()

### Community 142 - "Convert"
Cohesion: 0.40
Nodes (5): Result, Convert(), TestConvertUnterminatedArray(), TestGeneratedHeader(), TestGeneratedHeaderSanitisesSource()

### Community 153 - "TestShowRedundancyIsReadOnly"
Cohesion: 0.67
Nodes (3): TestBackupActivityStateReadsTheMateColumn(), TestShowRedundancyIsReadOnly(), backupActivityState()

### Community 154 - "TestExecute"
Cohesion: 0.67
Nodes (3): TestExecute(), Execute(), main()

### Community 155 - "github.com/spf13/cobra.Command"
Cohesion: 0.10
Nodes (31): shorthand, mdRow(), renderAbbrevDocs(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable(), writeUnabbreviated() (+23 more)

### Community 156 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

## Knowledge Gaps
- **243 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+238 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 363 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **16 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Config` connect `Config` to `newTestOps`, `Role`, `cli_test.go`, `Cluster`, `containerTransport`, `K8sConfig`, `Manager`, `ReadAdminPassword`, `opArgvMatch`, `sempPort`, `NewCluster`, `convert_test.go`, `manager_test.go`, `deploymentItem`, `ResolveDomainCerts`, `Ops`, `recRunner`, `EnvPairs`, `eqArgs`, `localCfg`, `config_test.go`, `prep_test.go`, `Cluster`, `.configRows`, `opPerformReplication`, `Compose`, `ReplSite`, `k8s/matechannel_test.go`, `AdminSecret`, `render_test.go`, `Container`, `Platform`, `GenOperator`, `capRunner`, `certFixture`, `Quadlet`, `writeNode`, `bytes.Buffer`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `newTestOps`, `context.Context`, `containerTransport`, `K8sConfig`, `Manager`, `ParseBlocks`, `Set`, `renderDriver`, `ImportPlan`, `Config`, `Ops`, `App`, `EnvPairs`, `Ops`, `config_test.go`, `runExport`, `.rpc`, `opPerformReplication`, `TestServerCert`, `k8s/matechannel_test.go`, `.ConfigureReplication`, `MateChannel`, `ParseRole`, `Ops`, `currentConfigScript`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Why does `Manager` connect `Manager` to `context.Context`, `Platform`, `Config`, `manager_test.go`, `config_test.go`, `confirm.go`, `ctrManager`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Are the 113 inferred relationships involving `newTestOps()` (e.g. with `TestDomainCertsLoadsInSortedNameOrder()` and `TestEnableDefaultUsersStartsEveryDefaultUserBackUp()`) actually correct?**
  _`newTestOps()` has 113 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _243 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `bg` be split into smaller, more focused modules?**
  _Cohesion score 0.12 - nodes in this community are weakly interconnected._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.05217391304347826 - nodes in this community are weakly interconnected._
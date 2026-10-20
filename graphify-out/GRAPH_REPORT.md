# Graph Report - solace-cnt-scripts  (2026-10-20)

## Corpus Check
- 200 files · ~684,256 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 4076 nodes · 17424 edges · 159 communities (142 shown, 17 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 2539 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `adb0fcf3`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- App
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
- opArgvMatch
- NewCluster
- ParseBlocks
- convert_test.go
- transform_test.go
- manager_test.go
- commands.go
- AdminSecret
- Bash Dev Script
- coverage_test.go
- PowerShell Dev Script
- Abbreviation Sets
- BuildSwitchPlan
- newLocalOps
- runner_test.go
- renderDriver
- Quadlet
- ImportPlan
- deploymentItem
- Solace Module Root
- Config
- Docs and Legacy Scripts
- ParseVPNReplication
- ResolveDomainCerts
- Role
- bg
- recRunner
- annotate_test.go
- completion_test.go
- diff_test.go
- Compose
- inject_test.go
- Ops
- exportconfig_test.go
- Echo
- eqArgs
- MateChannel
- localCfg
- config_test.go
- haCfg
- newRootCmd
- euid_test.go
- Container Security Guidelines
- Operations
- Test catalogue
- Cluster
- replNewRig
- New
- .rpc
- confirmAction
- .configRows
- rootless_test.go
- strings.Builder
- output_test.go
- Configuration Guide
- parse
- newTestMate
- Command Reference Doc
- Developer guide
- opPerformReplication
- .DomainCerts
- fakeTransport
- internal/k8s
- Docker and Podman mechanics
- usagef
- .ConfigureReplication
- Command
- newTeardownApplyFixture
- platformOps
- ReplSite
- replCfg
- limits_test.go
- Removing a broker: what stays, what goes
- Troubleshooting
- k8s/inspect_test.go
- mateSoftware
- GenSecrets
- render_test.go
- vulnjudge/main.go
- RuleFor
- github.com/spf13/cobra.Command
- Container
- Config
- Secret Reference Resolution
- twoMates
- replConfig
- Abbreviation Reference
- ParseShowReplication
- .checkUserManagerLimits
- .Preflight
- Load
- MateConfig
- Platform
- internal/cli
- k8s/inspect.go
- deniedCluster
- Import-Config Apply Rules
- Config
- internal/broker
- GenOperator
- capRunner
- Config Package Tests
- emitYAML
- operatorversion.go
- certFixture
- internal/container
- Container State Inspection
- runPlatform
- Cluster
- scripts.go
- matchCLI
- .validateContainerArtifactValues
- EnvPairs
- BrokerType
- watchCluster
- .applyScalingTierDefaults
- age
- Scaling YAML Decoding
- podName
- .preflightOne
- ReplVia
- Namespace Protection Checks
- allowcommand_test.go
- ContainerSecrets
- runLeader
- Fixtures and doubles
- guardConfig
- newBlock
- containerTransport
- CheckHostPath
- normalizeToList
- currentConfigScript
- sempPort
- ParseRole
- ResolveEnvPath
- NodeIdentity
- Convert

## God Nodes (most connected - your core abstractions)
1. `newTestOps()` - 162 edges
2. `Role` - 131 edges
3. `ctrCfg()` - 131 edges
4. `NewCluster()` - 128 edges
5. `Config` - 113 edges
6. `newCapMgr()` - 111 edges
7. `loadK8s()` - 110 edges
8. `matchCLI()` - 94 edges
9. `Platform` - 79 edges
10. `bg()` - 70 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `ExitCode()`  [EXTRACTED]
  main.go → internal/cli/exit.go
- `applyMetaField()` --calls--> `BrokerType`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go
- `TestAnnotateRegionAndSectionFencesAreBalanced()` --calls--> `Annotate()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/annotate.go
- `sectionBeginMarker()` --calls--> `RuleFor()`  [INFERRED]
  internal/broker/annotate.go → internal/broker/sections.go
- `TestAnnotateRegionAndSectionFencesAreBalanced()` --calls--> `loadSample()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/blocks_test.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (159 total, 17 thin omitted)

### Community 0 - "App"
Cohesion: 0.10
Nodes (44): wantEnable(), wantRemove(), domainCANames(), k8sAdminPassword(), k8sCluster(), k8sContext(), k8sLogin(), k8sOps() (+36 more)

### Community 1 - "newTestOps"
Cohesion: 0.07
Nodes (54): recOutput, newTestOps(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultUsersRefusesAQuotedVPNName(), TestDisableDefaultVPN(), TestDomainCertsAcceptsAFullHostPath() (+46 more)

### Community 2 - "Sink"
Cohesion: 0.20
Nodes (5): solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), columnWidths(), Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.05
Nodes (15): covUploadFailsTo, Exec, Manager, idMapCovers(), origin(), runUserIDs(), TestResolveMissingBinaryIsActionable(), EnvRunner (+7 more)

### Community 4 - "cli_test.go"
Cohesion: 0.08
Nodes (58): allowRuntime(), capture(), captureStderr(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone(), runStatusStderr() (+50 more)

### Community 5 - "testing.T"
Cohesion: 0.02
Nodes (202): TestExecShellScriptUploadFailureRunsNothing(), TestRouterNameReadsTheCaptureHeader(), sempV1OK(), TestSempV1OKMatchesExecuteResultOK(), TestSempV1OKRejectsErrorReply(), TestDetectRoleByAddress(), TestDetectRoleNameBeatsAddress(), TestDetectRoleNoMatchNamesBothPasses() (+194 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (8): certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing()

### Community 8 - "K8sConfig"
Cohesion: 0.10
Nodes (15): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, LoadBalancer, cut() (+7 more)

### Community 9 - "runRootWith"
Cohesion: 0.09
Nodes (50): echoRunner(), runCtr(), runRootWith(), TestAssertLeaderOnKubernetesTouchesOnlyThePrimaryPod(), TestAssertLeaderStandaloneSkipsBeforeAnyPrompt(), TestConfigStepsDoNotLeakSecrets(), TestConfiguredRouternameSurvivesTheFallback(), TestCtrConfigDryRun() (+42 more)

### Community 10 - "CheckCommand"
Cohesion: 0.16
Nodes (21): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+13 more)

### Community 11 - "opArgvMatch"
Cohesion: 0.06
Nodes (69): opCall, ReplicationConfigResult, captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook(), loadDirect() (+61 more)

### Community 12 - "NewCluster"
Cohesion: 0.06
Nodes (90): TestCheckStopsProbingWhenUnreachable(), TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestOperatorInstallVerdict(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce() (+82 more)

### Community 13 - "ParseBlocks"
Cohesion: 0.09
Nodes (27): Annotate(), StripMarkers(), TestAnnotateIsDeterministic(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), TestStripMarkersKeepsPragmaAndSectionComments(), TestStripMarkersNoMarkersPassthrough(), ParseBlocks() (+19 more)

### Community 14 - "convert_test.go"
Cohesion: 0.19
Nodes (31): checkGolden(), convertOK(), hasWarning(), strictDecode(), TestConvertAdminSecretAlias(), TestConvertAdminUserIsDroppedOnEveryPlatform(), TestConvertBadBooleanWarns(), TestConvertBadNumberWarns() (+23 more)

### Community 15 - "transform_test.go"
Cohesion: 0.07
Nodes (67): Block, Region, TargetState, regionFor(), importIgnore(), keepRegion(), TestImportIgnoreKeepsUnclassifiedSectionsInTheDiff(), TestImportOpsImportIgnore() (+59 more)

### Community 16 - "manager_test.go"
Cohesion: 0.05
Nodes (121): inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer() (+113 more)

### Community 17 - "commands.go"
Cohesion: 0.19
Nodes (40): addCommands(), addLogFlags(), group(), newBrokerConfigureCmd(), newBrokerPerformCmd(), newBrokerRemoveCmd(), newBrokerStartCmd(), newBrokerStopCmd() (+32 more)

### Community 18 - "AdminSecret"
Cohesion: 0.12
Nodes (19): AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), checkGolden(), decodeDataValue(), TestAdditionalUsersSecretCarriesBothHalves(), TestAdditionalUsersSecretIsAbsentWithNoUsers() (+11 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "coverage_test.go"
Cohesion: 0.05
Nodes (45): ranContains(), TestLastLines(), writeFile(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError() (+37 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Abbreviation Sets"
Cohesion: 0.13
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 23 - "BuildSwitchPlan"
Cohesion: 0.13
Nodes (25): PhaseKind, SiteState, SwitchAction, SwitchPhase, BuildSwitchPlan(), confirmDemotions(), SwitchPlan, promoteConfirmed() (+17 more)

### Community 24 - "newLocalOps"
Cohesion: 0.11
Nodes (32): TestLeaderAssertLeaderError(), TestRedundancyBackupShowError(), TestRedundancyReleaseError(), TestRedundancyRevertToPrimaryError(), TestReleaseToBackupNoReleaseError(), TestReleaseToBackupUnreleasedTimeout(), TestImportOpsExportConfigFailsLoudAtTheFirstBadCapture(), cliScriptPath() (+24 more)

### Community 25 - "runner_test.go"
Cohesion: 0.10
Nodes (29): TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), MaskEnv(), quoteTok(), captureStdout() (+21 more)

### Community 26 - "renderDriver"
Cohesion: 0.12
Nodes (19): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), TestKeywordPatternEscapesEachPhrase() (+11 more)

### Community 27 - "Quadlet"
Cohesion: 0.17
Nodes (16): ContainerNoFile(), HealthCheck, TestLimitsCheckAssertsWhatTheArtifactAsks(), escapePercent(), healthCmd(), Quadlet(), quadletEscape(), healthCheckFixture() (+8 more)

### Community 28 - "ImportPlan"
Cohesion: 0.12
Nodes (22): ImportResult, PlannedSection, isPragma(), defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList() (+14 more)

### Community 29 - "deploymentItem"
Cohesion: 0.12
Nodes (18): containsString(), installSteps(), desiredWatch(), Cluster, splitWatch(), subtractWatch(), TestDesiredWatchAppendsTheBrokerNamespace(), TestSubtractWatchEmptyRemainingMeansDelete() (+10 more)

### Community 31 - "Config"
Cohesion: 0.10
Nodes (17): keyValueEntries, TestValidateContainerArtifactValues(), TestRoleNames(), RoleNames(), checkCredentialChars(), foldToEnvVar(), Config, invertedCPUSetRange() (+9 more)

### Community 32 - "Docs and Legacy Scripts"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "ParseVPNReplication"
Cohesion: 0.10
Nodes (23): colSpan, dashSpans(), flagByte(), gutterClear(), ParseVPNReplication(), sliceSpan(), TestGutterClearAndSliceSpanBoundaries(), TestParseVPNReplication() (+15 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (38): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, caNameSafe(), checkNoDotDot(), DefaultDirReader() (+30 more)

### Community 35 - "Role"
Cohesion: 0.09
Nodes (15): Ops, New(), runCLISkeleton(), TestValidName(), validName(), TestNewDefaults(), shQuote(), TestShQuoteHandlesASingleQuote() (+7 more)

### Community 36 - "bg"
Cohesion: 0.14
Nodes (39): TestCtrManagerConfirmWiring(), confirmAssertFromBackup(), containerRenderRole(), containerRole(), containerWhat(), ctrLogin(), ctrManager(), ctrOps() (+31 more)

### Community 37 - "recRunner"
Cohesion: 0.13
Nodes (10): TestCanIAnswerReadsTheLastLine(), NewTransport(), canIVerb(), TestTransportCopy(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest(), covApplyRunner (+2 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.07
Nodes (45): applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV(), readMarker(), renderKV() (+37 more)

### Community 39 - "completion_test.go"
Cohesion: 0.18
Nodes (21): driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion(), TestCompletionDescriptionsAreOptIn() (+13 more)

### Community 40 - "diff_test.go"
Cohesion: 0.09
Nodes (36): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, ancestorReported(), blockLabel(), DiffBlocks() (+28 more)

### Community 41 - "Compose"
Cohesion: 0.19
Nodes (10): Compose(), composeEscape(), ComposeProject(), ContainerSecret, secretFilePath(), TestBridgePublishesOnlyTheListedPorts(), TestComposeProjectFoldsToComposesGrammar(), TestComposeProjectIsDeclaredNotDerived() (+2 more)

### Community 42 - "inject_test.go"
Cohesion: 0.15
Nodes (27): lineRole, serviceLine, shutdownStyle, svcKey, classifyServiceRest(), containsPortCommand(), describeKey(), fieldsUnquoted() (+19 more)

### Community 43 - "Ops"
Cohesion: 0.10
Nodes (19): countContains(), field(), TestCountContains(), TestField(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), showRedundancyLocalScript(), defaultLocalAddrs() (+11 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.11
Nodes (25): opRunner, TestResolveScriptPath(), ExitCode(), TestChildExitStatusIsScopedToInteractiveSessions(), exportconfigDriverOutput(), exportconfigTransportOutput(), exportconfigVPNListOutput(), newExportconfigRunner() (+17 more)

### Community 45 - "Echo"
Cohesion: 0.20
Nodes (6): interactiveFailRunner, runFailRunner, childStatusError(), Echo, Quote(), TestQuote()

### Community 46 - "eqArgs"
Cohesion: 0.10
Nodes (32): TestReachable(), newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), saCfg(), TestHARoles() (+24 more)

### Community 47 - "MateChannel"
Cohesion: 0.06
Nodes (18): AdminState, QueueState, scriptedMate, replMate, MateChannel, ReplRole, VPNRepl, sortedKeys() (+10 more)

### Community 48 - "localCfg"
Cohesion: 0.08
Nodes (44): TestLocalMateRunsThroughOps(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK(), TestBackupRevertActivitySuccess() (+36 more)

### Community 49 - "config_test.go"
Cohesion: 0.20
Nodes (20): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, span, dropMessageSpoolLines(), dropRoutingInterfaceLine(), TestSectionsDropMessageSpoolLines() (+12 more)

### Community 50 - "haCfg"
Cohesion: 0.09
Nodes (31): AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestReadAdminPasswordArgvAndDecoding(), TestReadAdminPasswordGuardsTheCommand(), TestReadAdminPasswordNamesTheNextStep() (+23 more)

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (28): TestAbbreviationDocs(), applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias() (+20 more)

### Community 52 - "euid_test.go"
Cohesion: 0.67
Nodes (5): runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand(), TestPodmanEUIDGuardSkips(), writeRootlessPodmanEnv()

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
Cohesion: 0.21
Nodes (17): replRig, replNewRig(), replUp(), TestOpConfigureReplicationAppliesTheRoleTheBrokersSiteDerives(), TestOpConfigureReplicationConvergesTheMateAndReportsWhatItStopped(), TestOpConfigureReplicationDeclinedLeavesTheBrokerUntouched(), TestOpConfigureReplicationFailureReportsNothingApplied(), TestOpConfigureReplicationRefusesAnUnknownBrokerBeforeAsking() (+9 more)

### Community 58 - "New"
Cohesion: 0.09
Nodes (25): TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportPlanReportShowsOmissionsAndEachSectionsVerdict(), TestImportResultReportQuotesTheFailedChunkAndItsTail(), TestImportResultReportsTeardownAndRebuildSeparately(), TestWarnAndStep(), App, lineSink() (+17 more)

### Community 59 - ".rpc"
Cohesion: 0.12
Nodes (16): backupTarget, Credential, sempMate, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigLine(), Ops, httpBody() (+8 more)

### Community 60 - "confirmAction"
Cohesion: 0.17
Nodes (22): exportconfigReadCounter, layer, TestConfirmDeleteShortcut(), TestPromptYes(), TestPromptYesNo(), TestStdinCanAnswerClosedFile(), confirmAction(), confirmActionStrict() (+14 more)

### Community 61 - ".configRows"
Cohesion: 0.14
Nodes (22): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminSecretRow(), failRow(), Cluster (+14 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.17
Nodes (35): failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent(), TestCheckIDMappingRefusesAnUnmappedRunUser() (+27 more)

### Community 63 - "strings.Builder"
Cohesion: 0.08
Nodes (37): shorthand, WeightedNodeTerm, mdRow(), renderAbbrevDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes() (+29 more)

### Community 64 - "output_test.go"
Cohesion: 0.16
Nodes (17): NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLevelOrderKeepsSkipBelowFail(), TestLineAddsNoPrefix() (+9 more)

### Community 65 - "Configuration Guide"
Cohesion: 0.14
Nodes (14): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+6 more)

### Community 66 - "parse"
Cohesion: 0.11
Nodes (20): segment, vars, countMarkers(), resolvePlatform(), TestParseArrayElementsPreserveQuoting(), TestParseAssignmentForms(), TestParseCRLF(), TestParseDoubleQuotedBackslashEscapes() (+12 more)

### Community 67 - "newTestMate"
Cohesion: 0.14
Nodes (22): CLIRunner, fakeRun, Ops, NewCLIMate(), mateReplies(), newTestMate(), TestMateChannelCallsKeepTheCauseWhenTheRunnerFails(), TestMateChannelDescribeNamesTheTarget() (+14 more)

### Community 68 - "Command Reference Doc"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.20
Nodes (10): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+2 more)

### Community 70 - "opPerformReplication"
Cohesion: 0.13
Nodes (28): mateChannelFunc, confirmReplicationConfig(), mateChannel(), mateSEMPPassword(), opConfigureReplication(), opCtrConfigureReplication(), opCtrPerformReplication(), opK8sConfigureReplication() (+20 more)

### Community 71 - ".DomainCerts"
Cohesion: 0.32
Nodes (7): TestDomainCerts(), TestPathHelpers(), TestDomainCertsLoadsInSortedNameOrder(), domainCertsScript(), TestDomainCertsScriptSorted(), certPath(), cliArg()

### Community 72 - "fakeTransport"
Cohesion: 0.10
Nodes (18): downloadErrTransport, fakeTransport, recDownload, recRun, recUpload, recUploadFile, runErrMatchTransport, runErrTransport (+10 more)

### Community 73 - "internal/k8s"
Cohesion: 0.10
Nodes (21): adminsecret_test.go, check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, generatorpage_test.go, inspect_test.go, internal/k8s (+13 more)

### Community 74 - "Docker and Podman mechanics"
Cohesion: 0.40
Nodes (5): Docker and Podman mechanics, Re-deploying is safe and explicit, Rootless podman prerequisites, What the broker cannot gain, Which user the broker runs as

### Community 75 - "usagef"
Cohesion: 0.08
Nodes (27): childExitError, usageError, ValidScriptName(), logArgs(), runConvert(), asUsage(), childExit(), childStatus() (+19 more)

### Community 76 - ".ConfigureReplication"
Cohesion: 0.24
Nodes (7): Ops, isRunCLIRejection(), newlineIf(), replPhase1Rejected(), replPhase2Rejected(), replTransportFailure(), replVPNList()

### Community 77 - "Command"
Cohesion: 0.14
Nodes (12): cmdField, guardedCmd, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandUnmarshal(), TestValidateProbeCommandAccepts(), TestValidateProbeCommandRejects(), Command (+4 more)

### Community 78 - "newTeardownApplyFixture"
Cohesion: 0.12
Nodes (21): planSections(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), newTeardownApplyFixture(), TestBuildChunksOrdersTeardownClearCreateSectionsThenVPNs(), TestImportOpsImportApplyClearsTargetOnlyObjectsBeforeAnythingIsApplied() (+13 more)

### Community 79 - "platformOps"
Cohesion: 0.29
Nodes (21): opFunc, roleOpFunc, addPodFlag(), newBrokerCLICmd(), newBrokerCmd(), newBrokerCopyCmd(), newBrokerDeployCmd(), newBrokerExecCmd() (+13 more)

### Community 81 - "ReplSite"
Cohesion: 0.14
Nodes (22): Replication, mateConvergenceShutdowns(), missingListedVPNs(), PlannedRoles(), replicationVPNLines(), RoleAtSite(), siteAEntry(), TestMateConvergenceShutdownsIsDeterministic() (+14 more)

### Community 82 - "replCfg"
Cohesion: 0.16
Nodes (26): CLIArg(), CLIScriptPath(), execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf() (+18 more)

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

### Community 87 - "mateSoftware"
Cohesion: 0.14
Nodes (22): TestMateChannelShowReplication(), containsEndpoint(), normTransport(), RenderMateRemovals(), SameEndpoints(), mateAppliance(), mateSoftware(), TestMateRoundTripsThroughItsOwnGrammar() (+14 more)

### Community 88 - "GenSecrets"
Cohesion: 0.07
Nodes (35): GenBroker(), GenSecrets(), joinManifests(), adminCfg(), secretNamesIn(), TestAdminSecretStatesEmitTheirArtifacts(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree() (+27 more)

### Community 89 - "render_test.go"
Cohesion: 0.16
Nodes (30): boolStr(), BrokerCR(), containerArtifacts(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets(), TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPrivilegePosture() (+22 more)

### Community 90 - "vulnjudge/main.go"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.08
Nodes (35): Disposition, Omission, SectionRule, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+27 more)

### Community 92 - "github.com/spf13/cobra.Command"
Cohesion: 0.13
Nodes (19): newPerformExportConfigCmd(), noRolePositional(), completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), completionShell(), isEnvFileName() (+11 more)

### Community 93 - "Container"
Cohesion: 0.16
Nodes (12): Container, DockerConfig, Network, PodmanConfig, Ulimits, applyContainerBlockDefaults(), Config, podmanRunUser() (+4 more)

### Community 94 - "Config"
Cohesion: 0.20
Nodes (6): expandTilde(), expandTildeToken(), Config, isPathSep(), TestExpandTilde(), tildeHome()

### Community 95 - "Secret Reference Resolution"
Cohesion: 0.50
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 96 - "twoMates"
Cohesion: 0.23
Nodes (20): ExecuteSwitchPlan(), SwitchPreflight(), ordersToA(), replSites(), TestExecuteSwitchPlanEmptyDoesNothing(), TestExecuteSwitchPlanOrder(), TestExecuteSwitchPlanRefusesASiteItHasNoChannelFor(), TestExecuteSwitchPlanRefusesToPromoteAgainstStaleState() (+12 more)

### Community 97 - "replConfig"
Cohesion: 0.31
Nodes (9): replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationEmptyViaAccepted(), TestValidateReplicationRejects(), TestValidateReplicationValid() (+1 more)

### Community 98 - "Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "ParseShowReplication"
Cohesion: 0.23
Nodes (12): ParseShowReplication(), RenderMate(), loadShowReplication(), TestParseShowReplicationRejectsAnUnreadableAddressOrPort(), TestParseShowReplicationUnconfigured(), TestParseShowReplicationWrongPlatformFails(), TestRealApplianceCaptureMateRoundTrips(), TestRealApplianceShowReplicationParses() (+4 more)

### Community 100 - ".checkUserManagerLimits"
Cohesion: 0.29
Nodes (6): delegateSetting(), Manager, limitText(), parseLimit(), parseUnitProps(), TestParseLimit()

### Community 102 - "Load"
Cohesion: 0.06
Nodes (58): Example, TestServerCertBundleOrder(), TestServerCertBundleRefusesAKeyBearingCert(), TestServerCertBundleReportsAnUnreadableFile(), TestServerCertBundleRequiresBothHalves(), ServerCertBundle(), TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout() (+50 more)

### Community 103 - "MateConfig"
Cohesion: 0.17
Nodes (13): cliTransport(), MateConfig, labelValue(), parseHostPort(), parseShowReplicationAppliance(), parseShowReplicationSoftware(), renderMateAppliance(), renderMateSoftware() (+5 more)

### Community 104 - "Platform"
Cohesion: 0.12
Nodes (29): checkAllowCommand(), checkFlagPlatforms(), commandPlatforms(), declaredList(), parsePlatformList(), platformSuffix(), prepare(), promptPlatform() (+21 more)

### Community 105 - "internal/cli"
Cohesion: 0.14
Nodes (14): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, euid_test.go, examples_test.go (+6 more)

### Community 106 - "k8s/inspect.go"
Cohesion: 0.19
Nodes (16): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), ownedPods(), brokerList, brokerStatus (+8 more)

### Community 107 - "deniedCluster"
Cohesion: 0.18
Nodes (13): TestValidateUnderADeniedClusterWideListReadsTheOperatorNamespace(), operatorImage(), TestOperatorApplyAsksBeforeWideningTheWatchList(), TestOperatorImage(), deniedCluster(), namedOperatorJSON(), operatorItem(), TestAmbiguousOperatorAbortsTheDeploy() (+5 more)

### Community 108 - "Import-Config Apply Rules"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "Config"
Cohesion: 0.08
Nodes (16): AdditionalUser, Image, Node, Redundancy, SEMP, TLS, atoiPrefix(), Config (+8 more)

### Community 110 - "internal/broker"
Cohesion: 0.08
Nodes (25): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, config_ops_test.go, coverage_test.go, diff_test.go, driver_test.go (+17 more)

### Community 111 - "GenOperator"
Cohesion: 0.16
Nodes (20): GenOperator(), joinYAMLDocs(), nonEmpty(), operatorEscalationProbes(), operatorProbes(), RenderOperator(), splitAfterNamespace(), splitOperatorBundle() (+12 more)

### Community 112 - "capRunner"
Cohesion: 0.11
Nodes (23): capCall, capRunner, TestPreflightRunsBeforeAnything(), mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime() (+15 more)

### Community 113 - "Config Package Tests"
Cohesion: 0.12
Nodes (17): adminsecret_test.go, artifactvalues_test.go, command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go (+9 more)

### Community 114 - "emitYAML"
Cohesion: 0.22
Nodes (9): doc, boolOf(), commentSafe(), emitYAML(), joinDomainCertPath(), kubeCommand(), redundancy(), scalar() (+1 more)

### Community 115 - "operatorversion.go"
Cohesion: 0.23
Nodes (10): compareVersions(), Cluster, imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestImageTag() (+2 more)

### Community 116 - "certFixture"
Cohesion: 0.14
Nodes (18): ResolveSecretValues(), certFixture(), TestContainerBundleCarriesTheChain(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeWritingTheComposeFile(), TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(), TestEchoPreviewsTheLegacyBundleRemoval(), TestLegacyBundleIsRemovedOnDeployAndRemove() (+10 more)

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.16
Nodes (14): containerState, healthState, decodeInspect(), Manager, orUnknown(), printable(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown() (+6 more)

### Community 119 - "runPlatform"
Cohesion: 0.29
Nodes (12): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+4 more)

### Community 120 - "Cluster"
Cohesion: 0.22
Nodes (3): Cluster, namespaceManifest(), ownedSecretNames()

### Community 121 - "scripts.go"
Cohesion: 0.07
Nodes (39): showCmd, validCLILine(), ValidVPNName(), validVPNName(), Ops, assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript() (+31 more)

### Community 122 - "matchCLI"
Cohesion: 0.11
Nodes (40): cliRunNames(), hasCall(), matchCLI(), TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestDisableDefaultUsersShowVPNError(), TestLeaderReadError() (+32 more)

### Community 123 - ".validateContainerArtifactValues"
Cohesion: 0.20
Nodes (7): Config, portOrRange(), TestValidContainerPort(), validContainerPort(), validCoreLimit(), validHealthDuration(), validDNSLabel()

### Community 124 - "EnvPairs"
Cohesion: 0.21
Nodes (11): EnvPairs(), groupKey(), itoa(), assertNoCheckoutPath(), envLines(), TestAdditionalUsersReachBothHalves(), TestGolden(), TestScalingReachesContainersAsEnv() (+3 more)

### Community 125 - "BrokerType"
Cohesion: 0.24
Nodes (7): BrokerType, cliMate, bannerType(), checkSameType(), showReplicationScript(), showVPNReplicationScript(), TestTranscriptBannerType()

### Community 126 - "watchCluster"
Cohesion: 0.22
Nodes (10): deployJSON(), TestFindOperatorDeploymentIgnoresOtherDeployments(), TestOperatorReleaseDeletesTheOperatorWhenTheEnvFileCoversEveryNamespace(), TestOperatorReleaseLeavesAnOperatorThatWatchesEverything(), TestOperatorReleaseNarrowingNeedsPermissionToPatch(), TestOperatorReleaseNarrowsInsteadOfRemoving(), TestReconcileWatchDecidesWhatToApply(), TestReconcileWatchFlagsAWideningAsAQuestion() (+2 more)

### Community 127 - ".applyScalingTierDefaults"
Cohesion: 0.15
Nodes (11): scalingTier, containerMem(), cpuSetCount(), cpuSetRange(), Config, TestContainerMem(), TestCPUSetCount(), TestCPUSetRange() (+3 more)

### Community 128 - "age"
Cohesion: 0.24
Nodes (8): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownsPod(), TestOperatorRunningImageWithNoContainers(), TestOwnsPodFallsBackToTheNameInfix()

### Community 129 - "Scaling YAML Decoding"
Cohesion: 0.31
Nodes (5): scalingKey, scalingSpelling, Scaling, scalingKeyIndex(), scalingKeyList()

### Community 130 - "podName"
Cohesion: 0.11
Nodes (13): BrokerPodSuffixShape(), TestOperatorAdminSecretNameMatchesTheBrokerNames(), lbServiceName(), podName(), pvcName(), RestartOrder(), stsName(), TestPodNameSuffixMatchesTheConfigBound() (+5 more)

### Community 131 - ".preflightOne"
Cohesion: 0.47
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "ReplVia"
Cohesion: 0.18
Nodes (9): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, viaDescription(), sortedKeys(), validateReplEndpoints(), validateReplVia() (+1 more)

### Community 134 - "allowcommand_test.go"
Cohesion: 0.29
Nodes (10): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+2 more)

### Community 135 - "ContainerSecrets"
Cohesion: 0.20
Nodes (10): ContainerSecrets(), containerSecretSpecs(), SecretPreflight(), TestContainerSecretNamesAreHostScoped(), TestContainerSecretsRedundancy(), TestSecretPreflight(), TestComposeLabelsTheCertificateDigest(), TestFileBackedSecretIsExemptFromSecretPreflight() (+2 more)

### Community 136 - "runLeader"
Cohesion: 0.60
Nodes (4): leaderRun, runLeader(), TestAssertLeaderAsksBeforeActingFromABackup(), TestAssertLeaderPrimaryAndMonitorNeverAsk()

### Community 137 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 138 - "guardConfig"
Cohesion: 0.18
Nodes (11): guardConfig(), IsAbsHostPath(), TestContainerCertRequiresKey(), TestContainerHostDirsExpandATilde(), TestDataDirMustBeAbsolute(), TestIsAbsHostPath(), TestPodmanBaseDirIsOptionalButAbsolute(), TestQuadletDirAndDataDirAreNotRebased() (+3 more)

### Community 139 - "newBlock"
Cohesion: 0.24
Nodes (10): token, firstQuoted(), newBlock(), TestNewBlockMessageSpoolQualifierHasNoOwnName(), TestNewBlockOpener(), TestNewBlockVPNNameWithSpaces(), TestNewBlockVPNQualifierWithSpaces(), tokenizeOpener() (+2 more)

### Community 141 - "CheckHostPath"
Cohesion: 0.33
Nodes (5): CanonicalDuration(), TestCanonicalDuration(), CheckHostPath(), TestCheckHostPathAccepts(), TestCheckHostPathRejects()

### Community 142 - "normalizeToList"
Cohesion: 0.22
Nodes (5): Cluster, Cluster, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

### Community 153 - "currentConfigScript"
Cohesion: 0.22
Nodes (9): minimalCapture(), targetVPNCapture(), TestImportOpsExportConfigScopeSelectsCLICommand(), TestImportOpsImportPlanRefusesCrossTypeAndUnknownType(), TestImportOpsImportPlanRefusesRedactedArtifact(), TestImportOpsImportPlanReportsATargetItCannotRead(), TestImportOpsImportPlanSplitsExistingAndNewVPNs(), TestImportOpsImportVerifyReExportsAndDiffs() (+1 more)

### Community 154 - "sempPort"
Cohesion: 0.29
Nodes (7): bridgeHostPort(), sempPort(), TestSempPortBridgeFallsBackToPlaintext(), TestSempPortBridgePrefersTLS(), TestSempPortBridgeRefusesNoMapping(), TestSempPortHostMode(), TestSempPortWithoutCertificateStaysPlaintext()

### Community 155 - "ParseRole"
Cohesion: 0.33
Nodes (7): opK8sRestart(), roleWord(), TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), ParseRole(), RoleAbbrev()

### Community 156 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 158 - "Convert"
Cohesion: 0.40
Nodes (5): Result, Convert(), TestConvertUnterminatedArray(), TestGeneratedHeader(), TestGeneratedHeaderSanitisesSource()

## Knowledge Gaps
- **243 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+238 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 363 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **17 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Role` connect `Role` to `App`, `newTestOps`, `podName`, `context.Context`, `Manager`, `K8sConfig`, `containerTransport`, `ParseBlocks`, `Abbreviation Sets`, `renderDriver`, `ParseRole`, `ImportPlan`, `NodeIdentity`, `Config`, `bg`, `Ops`, `config_test.go`, `New`, `.rpc`, `newTestMate`, `opPerformReplication`, `.DomainCerts`, `fakeTransport`, `usagef`, `.ConfigureReplication`, `replCfg`, `scripts.go`?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `App`, `newTestOps`, `podName`, `cli_test.go`, `testing.T`, `Manager`, `K8sConfig`, `ContainerSecrets`, `opArgvMatch`, `containerTransport`, `NewCluster`, `convert_test.go`, `manager_test.go`, `AdminSecret`, `sempPort`, `Quadlet`, `deploymentItem`, `ResolveDomainCerts`, `Role`, `recRunner`, `Compose`, `eqArgs`, `localCfg`, `config_test.go`, `haCfg`, `Cluster`, `New`, `.configRows`, `strings.Builder`, `ReplSite`, `replCfg`, `GenSecrets`, `render_test.go`, `Container`, `Load`, `Platform`, `deniedCluster`, `GenOperator`, `capRunner`, `certFixture`, `Cluster`, `EnvPairs`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Why does `Platform` connect `Platform` to `testing.T`, `Manager`, `runLeader`, `ContainerSecrets`, `CheckCommand`, `opArgvMatch`, `containerTransport`, `CheckHostPath`, `guardConfig`, `convert_test.go`, `manager_test.go`, `commands.go`, `sempPort`, `Convert`, `Config`, `Role`, `config_test.go`, `New`, `parse`, `Command`, `platformOps`, `limits_test.go`, `render_test.go`, `Container`, `Config`, `Load`, `capRunner`, `emitYAML`, `certFixture`, `.validateContainerArtifactValues`, `.applyScalingTierDefaults`?**
  _High betweenness centrality (0.013) - this node is a cross-community bridge._
- **Are the 112 inferred relationships involving `newTestOps()` (e.g. with `TestDomainCertsLoadsInSortedNameOrder()` and `TestEnableDefaultUsersStartsEveryDefaultUserBackUp()`) actually correct?**
  _`newTestOps()` has 112 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _243 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.10202020202020202 - nodes in this community are weakly interconnected._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.06715063520871144 - nodes in this community are weakly interconnected._
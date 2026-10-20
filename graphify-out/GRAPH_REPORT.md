# Graph Report - solace-cnt-scripts  (2026-10-20)

## Corpus Check
- 200 files · ~685,528 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 4080 nodes · 17439 edges · 155 communities (137 shown, 18 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 2541 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `85df2b89`
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
- opArgvMatch
- NewCluster
- ParseBlocks
- convert_test.go
- transform_test.go
- manager_test.go
- commands.go
- writeK8sDeployAllEnv
- Bash Dev Script
- matchCLI
- PowerShell Dev Script
- Abbreviation Sets
- BuildSwitchPlan
- newLocalOps
- runner_test.go
- renderDriver
- Compose
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
- parseBody
- completion_test.go
- diff_test.go
- ContainerSecret
- inject.go
- Ops
- captureStderr
- Echo
- eqArgs
- VPNRepl
- localCfg
- config_test.go
- haCfg
- newRootCmd
- .Run
- Container Security Guidelines
- Operations
- Test catalogue
- Cluster
- replNewRig
- New
- .rpc
- confirm.go
- .configRows
- rootless_test.go
- strings.Builder
- output_test.go
- Configuration Guide
- opPerformReplication
- newTestMate
- Command Reference Doc
- Developer guide
- replApp
- Get
- .validateContainer
- internal/k8s
- Docker and Podman mechanics
- usagef
- parseKV
- Command
- newTeardownApplyFixture
- DetectPlatforms
- ReplSite
- replCfg
- limits_test.go
- Removing a broker: what stays, what goes
- Troubleshooting
- k8s/inspect_test.go
- ParseShowReplication
- GenSecrets
- render_test.go
- main_test.go
- RuleFor
- github.com/spf13/cobra.Command
- .applyContainerDefaults
- Config
- Secret Reference Resolution
- MateChannel
- replConfig
- Abbreviation Reference
- occCluster
- .checkUserManagerLimits
- .Preflight
- Load
- github.com/spf13/cobra.ShellCompDirective
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
- planSections
- operatorversion.go
- certFixture
- internal/container
- Container State Inspection
- runPlatform
- Container
- scripts.go
- hasCall
- .validateContainerArtifactValues
- EnvPairs
- execBase
- watchCluster
- .applyScalingTierDefaults
- importIgnore
- Scaling YAML Decoding
- Role
- .preflightOne
- ReplVia
- Namespace Protection Checks
- shellScriptPath
- logArgs
- runLeader
- Fixtures and doubles
- DecodeStrict
- newBlock
- containerTransport
- CheckHostPath
- allCustomMounted
- ParseRole
- ResolveEnvPath

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
- `assertFencesBalanced()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/annotate.go
- `assertFencesBalanced()` --calls--> `markerVerb()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/annotate.go
- `applyMetaField()` --calls--> `BrokerType`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go
- `loadApplianceCapture()` --calls--> `ParseBlocks()`  [INFERRED]
  internal/broker/appliance_test.go → internal/broker/blocks.go
- `TestRealApplianceCaptureMateRoundTrips()` --calls--> `loadApplianceCapture()`  [INFERRED]
  internal/broker/replication_test.go → internal/broker/appliance_test.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (155 total, 18 thin omitted)

### Community 0 - "bg"
Cohesion: 0.12
Nodes (51): TestConfirmDeleteShortcut(), confirmDelete(), warn(), bg(), k8sAdminPassword(), k8sCluster(), k8sContext(), k8sLogin() (+43 more)

### Community 1 - "newTestOps"
Cohesion: 0.06
Nodes (64): newTestOps(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultUsersRefusesAQuotedVPNName(), TestDisableDefaultVPN(), TestDomainCerts(), TestDomainCertsAcceptsAFullHostPath() (+56 more)

### Community 2 - "Sink"
Cohesion: 0.12
Nodes (12): solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), orNone(), orValue(), age(), roleRank(), Cluster, ownsPod() (+4 more)

### Community 3 - "context.Context"
Cohesion: 0.05
Nodes (15): Exec, Manager, idMapCovers(), origin(), runUserIDs(), EnvRunner, Runner, Cluster (+7 more)

### Community 4 - "cli_test.go"
Cohesion: 0.09
Nodes (49): allowRuntime(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone(), runStatusStderr(), TestAnnounceCommandsNamesResolvedBinaries(), TestBashEnvGivenToEnvFlag() (+41 more)

### Community 5 - "testing.T"
Cohesion: 0.01
Nodes (220): TestExecShellScriptUploadFailureRunsNothing(), TestDiagnosticsRunError(), TestDomainCertsUploadFileError(), TestExecCLIFailsOnErrorOutput(), TestExecCLIUploadFileError(), TestGatherNodeDownloadError(), TestOutDefaultsToStdout(), TestPollCondError() (+212 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (8): certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing()

### Community 8 - "K8sConfig"
Cohesion: 0.17
Nodes (7): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, writeSecurity()

### Community 9 - "runRootWith"
Cohesion: 0.09
Nodes (40): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+32 more)

### Community 10 - "CheckCommand"
Cohesion: 0.19
Nodes (17): commandRules, checkBinary(), CheckCommand(), clusterRules(), composeRules(), Config, runtimeRules(), siteRules() (+9 more)

### Community 11 - "opArgvMatch"
Cohesion: 0.08
Nodes (41): opCall, opRunner, ReplicationConfigResult, captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook() (+33 more)

### Community 12 - "NewCluster"
Cohesion: 0.06
Nodes (90): TestCheckStopsProbingWhenUnreachable(), TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestOperatorInstallVerdict(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce() (+82 more)

### Community 13 - "ParseBlocks"
Cohesion: 0.09
Nodes (29): Annotate(), StripMarkers(), assertFencesBalanced(), TestAnnotateIsDeterministic(), TestAnnotateRegionAndSectionFencesAreBalanced(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), TestStripMarkersKeepsPragmaAndSectionComments() (+21 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (65): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+57 more)

### Community 15 - "transform_test.go"
Cohesion: 0.07
Nodes (63): Block, Region, TargetState, regionFor(), keepRegion(), injectedBlock(), ClearExistingNested(), ClearExistingSyslogs() (+55 more)

### Community 16 - "manager_test.go"
Cohesion: 0.06
Nodes (116): inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer() (+108 more)

### Community 17 - "commands.go"
Cohesion: 0.22
Nodes (45): opFunc, roleOpFunc, addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerConfigureCmd(), newBrokerCopyCmd() (+37 more)

### Community 18 - "writeK8sDeployAllEnv"
Cohesion: 0.15
Nodes (28): opCanI(), operatorManagedK8sCfg(), TestDeployRefusesAnOperatorPasswordOverKeptData(), TestRemoveBrokerSaysWhereAnOperatorPasswordLives(), writeK8sDeployAllEnv(), opsK8sAnswer(), opsK8sPods(), opsK8sRun() (+20 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "matchCLI"
Cohesion: 0.05
Nodes (44): leaderCalls(), matchCLI(), ranContains(), TestLeaderSuccess(), TestLeaderSuccessOnBackup(), TestDisableDefaultUsersDisableError(), TestDisableDefaultVPNDisableError(), TestDisableDefaultVPNShowError() (+36 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Abbreviation Sets"
Cohesion: 0.13
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 23 - "BuildSwitchPlan"
Cohesion: 0.12
Nodes (33): PhaseKind, SiteState, SwitchAction, SwitchPhase, BuildSwitchPlan(), confirmDemotions(), ExecuteSwitchPlan(), SwitchPlan (+25 more)

### Community 24 - "newLocalOps"
Cohesion: 0.10
Nodes (37): TestPathHelpers(), TestLeaderAssertLeaderError(), TestRedundancyBackupShowError(), TestRedundancyReleaseError(), TestRedundancyRevertToPrimaryError(), TestReleaseToBackupNoReleaseError(), TestReleaseToBackupUnreleasedTimeout(), TestImportOpsExportConfigFailsLoudAtTheFirstBadCapture() (+29 more)

### Community 25 - "runner_test.go"
Cohesion: 0.10
Nodes (29): TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), MaskEnv(), quoteTok(), captureStdout() (+21 more)

### Community 26 - "renderDriver"
Cohesion: 0.11
Nodes (22): chunk, chunkResult, runCLISkeleton(), chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce() (+14 more)

### Community 27 - "Compose"
Cohesion: 0.11
Nodes (25): ContainerNoFile(), HealthCheck, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), Compose(), composeEscape(), ComposeProject(), escapePercent() (+17 more)

### Community 28 - "ImportPlan"
Cohesion: 0.12
Nodes (22): ImportResult, PlannedSection, isPragma(), defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList() (+14 more)

### Community 29 - "deploymentItem"
Cohesion: 0.14
Nodes (15): desiredWatch(), Cluster, splitWatch(), subtractWatch(), TestDesiredWatchAppendsTheBrokerNamespace(), TestSubtractWatchEmptyRemainingMeansDelete(), TestSubtractWatchPreservesOrder(), TestUnionWatchKeepsAnotherEnvFilesNamespace() (+7 more)

### Community 31 - "Config"
Cohesion: 0.12
Nodes (12): keyValueEntries, checkCredentialChars(), foldToEnvVar(), Config, missingErr(), requireAll(), requireKeyValue(), sortedKeys() (+4 more)

### Community 32 - "Docs and Legacy Scripts"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "ParseVPNReplication"
Cohesion: 0.10
Nodes (23): colSpan, dashSpans(), flagByte(), gutterClear(), ParseVPNReplication(), sliceSpan(), TestGutterClearAndSliceSpanBoundaries(), TestParseVPNReplication() (+15 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (40): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, caNameSafe(), checkNoDotDot(), DefaultDirReader() (+32 more)

### Community 35 - "Ops"
Cohesion: 0.18
Nodes (4): Ops, New(), TestNewDefaults(), TestTransportEchoHidesUploadBody()

### Community 36 - "App"
Cohesion: 0.10
Nodes (44): TestContainerRoleDetectsFromHostname(), TestCtrManagerConfirmWiring(), confirmAction(), wantEnable(), wantRemove(), resolveScriptPath(), confirmAssertFromBackup(), containerRenderRole() (+36 more)

### Community 37 - "recRunner"
Cohesion: 0.13
Nodes (10): TestCanIAnswerReadsTheLastLine(), NewTransport(), canIVerb(), TestTransportCopy(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest(), covApplyRunner (+2 more)

### Community 38 - "parseBody"
Cohesion: 0.08
Nodes (27): applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), readMarker(), TestApplyMetaFieldBrokerTypeSetsType(), TestApplyMetaFieldRedactedCaseInsensitive() (+19 more)

### Community 39 - "completion_test.go"
Cohesion: 0.18
Nodes (21): driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion(), TestCompletionDescriptionsAreOptIn() (+13 more)

### Community 40 - "diff_test.go"
Cohesion: 0.09
Nodes (36): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, ancestorReported(), blockLabel(), DiffBlocks() (+28 more)

### Community 42 - "inject.go"
Cohesion: 0.17
Nodes (27): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+19 more)

### Community 43 - "Ops"
Cohesion: 0.08
Nodes (25): countContains(), field(), TestCountContains(), TestField(), TestLastLines(), TestPrimaryRedundancyUp(), TestValidName(), validName() (+17 more)

### Community 44 - "captureStderr"
Cohesion: 0.08
Nodes (33): capture(), captureStderr(), TestConfirmDowngradeNeverUnattended(), TestConfirmLayerMatrix(), TestConfirmNonTTY(), TestExecute(), TestPromptAsksWhenStdinIsNotATTY(), TestPromptKeepsBufferedAnswersForLaterQuestions() (+25 more)

### Community 45 - "Echo"
Cohesion: 0.20
Nodes (6): interactiveFailRunner, runFailRunner, childStatusError(), Echo, Quote(), TestQuote()

### Community 46 - "eqArgs"
Cohesion: 0.10
Nodes (32): TestReachable(), newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), saCfg(), TestHARoles() (+24 more)

### Community 47 - "VPNRepl"
Cohesion: 0.05
Nodes (29): AdminState, BrokerType, cliMate, QueueState, scriptedMate, replMate, cliRunNames(), ValidVPNName() (+21 more)

### Community 48 - "localCfg"
Cohesion: 0.08
Nodes (44): TestLocalMateRunsThroughOps(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK(), TestBackupRevertActivitySuccess() (+36 more)

### Community 49 - "config_test.go"
Cohesion: 0.19
Nodes (22): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, dropMessageSpoolLines(), dropRoutingInterfaceLine(), TestSectionsDropMessageSpoolLines(), TestSectionsDropRoutingInterfaceLine() (+14 more)

### Community 50 - "haCfg"
Cohesion: 0.06
Nodes (45): AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestReadAdminPasswordArgvAndDecoding(), TestReadAdminPasswordGuardsTheCommand(), TestReadAdminPasswordNamesTheNextStep() (+37 more)

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (28): TestAbbreviationDocs(), applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias() (+20 more)

### Community 52 - ".Run"
Cohesion: 0.16
Nodes (23): runCtr(), TestAssertLeaderStandaloneSkipsBeforeAnyPrompt(), TestConfigStepsDoNotLeakSecrets(), TestCtrConfigDryRun(), TestCtrDiagnosticsDryRun(), TestCtrErrorPaths(), TestCtrLoginOutcomes(), TestCtrRestartConfirmGate() (+15 more)

### Community 53 - "Container Security Guidelines"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 54 - "Operations"
Cohesion: 0.10
Nodes (21): Bringing up a fresh cluster, Broker scope is a fixed classification, Configuring a site, Data replication, Exit codes, Exporting and importing configuration, Extra CLI users differ by platform, Handling the artifact (+13 more)

### Community 55 - "Test catalogue"
Cohesion: 0.10
Nodes (20): abbrev_test.go, convert_test.go, Coverage, examples_test.go, internal/abbrev, internal/convert, internal/engine, internal/examples (+12 more)

### Community 56 - "Cluster"
Cohesion: 0.11
Nodes (5): Cluster, Cluster, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

### Community 57 - "replNewRig"
Cohesion: 0.18
Nodes (19): replRig, replNewRig(), replShowMate(), replUp(), replVPNTable(), TestOpConfigureReplicationAppliesTheRoleTheBrokersSiteDerives(), TestOpConfigureReplicationConvergesTheMateAndReportsWhatItStopped(), TestOpConfigureReplicationDeclinedLeavesTheBrokerUntouched() (+11 more)

### Community 58 - "New"
Cohesion: 0.10
Nodes (22): TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportPlanReportShowsOmissionsAndEachSectionsVerdict(), TestImportResultReportQuotesTheFailedChunkAndItsTail(), TestImportResultReportsTeardownAndRebuildSeparately(), confirmImport(), pluralVPN(), runExport() (+14 more)

### Community 59 - ".rpc"
Cohesion: 0.10
Nodes (19): backupTarget, Credential, sempMate, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigLine(), Ops, httpBody() (+11 more)

### Community 60 - "confirm.go"
Cohesion: 0.19
Nodes (18): exportconfigReadCounter, layer, TestPromptYes(), TestPromptYesNo(), TestStdinCanAnswerClosedFile(), confirmActionStrict(), confirmDowngrade(), confirmGate() (+10 more)

### Community 61 - ".configRows"
Cohesion: 0.13
Nodes (23): setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminSecretRow(), containsString(), failRow(), Cluster (+15 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.17
Nodes (35): failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent(), TestCheckIDMappingRefusesAnUnmappedRunUser() (+27 more)

### Community 63 - "strings.Builder"
Cohesion: 0.10
Nodes (28): WeightedNodeTerm, mdRow(), writeAbbrevTable(), LoadBalancer, NodeAffinity, NodeMatchExpr, Placement, PodAffinityTerm (+20 more)

### Community 64 - "output_test.go"
Cohesion: 0.24
Nodes (14): sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLevelOrderKeepsSkipBelowFail(), TestLineAddsNoPrefix(), TestSectionPadsToWidth() (+6 more)

### Community 65 - "Configuration Guide"
Cohesion: 0.14
Nodes (14): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+6 more)

### Community 66 - "opPerformReplication"
Cohesion: 0.13
Nodes (19): mateChannelFunc, App, lineSink(), progress(), step(), confirmReplicationConfig(), opConfigureReplication(), opCtrConfigureReplication() (+11 more)

### Community 67 - "newTestMate"
Cohesion: 0.14
Nodes (22): CLIRunner, fakeRun, Ops, NewCLIMate(), mateReplies(), newTestMate(), TestMateChannelCallsKeepTheCauseWhenTheRunnerFails(), TestMateChannelDescribeNamesTheTarget() (+14 more)

### Community 68 - "Command Reference Doc"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.20
Nodes (10): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+2 more)

### Community 70 - "replApp"
Cohesion: 0.24
Nodes (15): mateChannel(), mateSEMPPassword(), replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer(), TestConfirmReplicationSwitchNamesEveryMoveAndIsStrict(), TestMateChannelPicksTheMechanismTheSiteDeclares() (+7 more)

### Community 71 - "Get"
Cohesion: 0.24
Nodes (13): Example, TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), All(), Get(), List(), Names(), read() (+5 more)

### Community 72 - ".validateContainer"
Cohesion: 0.21
Nodes (8): TestValidateContainerArtifactValues(), cpuSetCount(), cpuSetRange(), TestCPUSetCount(), TestCPUSetRange(), invertedCPUSetRange(), platformKey(), validIdent()

### Community 73 - "internal/k8s"
Cohesion: 0.10
Nodes (21): adminsecret_test.go, check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, generatorpage_test.go, inspect_test.go, internal/k8s (+13 more)

### Community 74 - "Docker and Podman mechanics"
Cohesion: 0.40
Nodes (5): Docker and Podman mechanics, Re-deploying is safe and explicit, Rootless podman prerequisites, What the broker cannot gain, Which user the broker runs as

### Community 75 - "usagef"
Cohesion: 0.12
Nodes (18): childExitError, usageError, runConvert(), asUsage(), childExit(), childStatus(), isUsage(), markUsageArgs() (+10 more)

### Community 76 - "parseKV"
Cohesion: 0.17
Nodes (12): parseKV(), renderKV(), splitKV(), TestMarkerValueSurvivesAnEmbeddedQuote(), TestParseKVEmptyPayload(), TestParseKVIgnoresTokenWithNoEquals(), TestParseKVMultipleTokens(), TestParseKVQuotedValueKeepsSpaces() (+4 more)

### Community 77 - "Command"
Cohesion: 0.14
Nodes (12): cmdField, guardedCmd, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandUnmarshal(), TestValidateProbeCommandAccepts(), TestValidateProbeCommandRejects(), Command (+4 more)

### Community 78 - "newTeardownApplyFixture"
Cohesion: 0.20
Nodes (14): driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), newTeardownApplyFixture(), TestImportOpsImportApplyClearsTargetOnlyObjectsBeforeAnythingIsApplied(), TestImportOpsImportApplyDriverExitingNonZeroAfterEveryChunkIsItsOwnFailure(), TestImportOpsImportApplyFailedTeardownCaptureStopsBeforeAnythingRuns() (+6 more)

### Community 79 - "DetectPlatforms"
Cohesion: 0.25
Nodes (11): parseError(), DetectPlatforms(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint(), TestDetectPlatformsMissingFile(), writeEnv(), TestEveryExampleLoads(), TestExamplesCarryNoRealSecret() (+3 more)

### Community 81 - "ReplSite"
Cohesion: 0.17
Nodes (12): Replication, missingListedVPNs(), PlannedRoles(), RoleAtSite(), TestPlannedRolesNamesEveryListedVPN(), TestRoleAtSiteIsTheComplement(), curlConfigFlag(), Ops (+4 more)

### Community 82 - "replCfg"
Cohesion: 0.21
Nodes (22): execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf(), replCfg(), replSite() (+14 more)

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
Cohesion: 0.14
Nodes (20): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestConditionLevelDegradesSafely() (+12 more)

### Community 87 - "ParseShowReplication"
Cohesion: 0.07
Nodes (49): TestMateChannelShowReplication(), cliTransport(), containsEndpoint(), MateConfig, labelValue(), normTransport(), parseHostPort(), ParseShowReplication() (+41 more)

### Community 88 - "GenSecrets"
Cohesion: 0.06
Nodes (40): GenSecrets(), joinManifests(), TestCreateSecretsAllThree(), TestGenSecretsBuildsNothingWithoutMaterial(), TestGenSecretsStopsAtASecretItCannotBuild(), TestGenSecretsTLSError(), TestJoinManifestsDropsBlankDocuments(), TestUpdateServerCertSecret() (+32 more)

### Community 89 - "render_test.go"
Cohesion: 0.13
Nodes (38): boolStr(), BrokerCR(), ContainerSecrets(), SecretPreflight(), containerArtifacts(), healthCheckFixture(), load(), TestAdminCredentialsSecretFollowsTheStates() (+30 more)

### Community 90 - "main_test.go"
Cohesion: 0.17
Nodes (18): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+10 more)

### Community 91 - "RuleFor"
Cohesion: 0.07
Nodes (39): Disposition, Omission, SectionRule, sectionBeginMarker(), TestSectionBeginMarkerAdvisoryDisposition(), TestWriteBlocksOpensAndClosesOnChange(), writeBlocks(), loadApplianceCapture() (+31 more)

### Community 92 - "github.com/spf13/cobra.Command"
Cohesion: 0.15
Nodes (27): addLogFlags(), newOperatorCmd(), newOperatorDeployCmd(), newOperatorGenerateCmd(), newOperatorLogsCmd(), newOperatorRemoveCmd(), newOperatorRestartCmd(), newOperatorStartCmd() (+19 more)

### Community 93 - ".applyContainerDefaults"
Cohesion: 0.27
Nodes (7): applyContainerBlockDefaults(), Config, podmanRunUser(), setDefault(), setDefaultCmd(), setDefaultInt(), xdgConfigHome()

### Community 94 - "Config"
Cohesion: 0.20
Nodes (6): expandTilde(), expandTildeToken(), Config, isPathSep(), TestExpandTilde(), tildeHome()

### Community 95 - "Secret Reference Resolution"
Cohesion: 0.50
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 96 - "MateChannel"
Cohesion: 0.16
Nodes (16): MateChannel, sortedKeys(), sortedVPNs(), requirePrimaryActive(), requireVirtualRouterNamesMatch(), sortedSiteNames(), SwitchPreflight(), replSites() (+8 more)

### Community 97 - "replConfig"
Cohesion: 0.31
Nodes (9): replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationEmptyViaAccepted(), TestValidateReplicationRejects(), TestValidateReplicationValid() (+1 more)

### Community 98 - "Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "occCluster"
Cohesion: 0.18
Nodes (10): occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants(), TestNamespaceIsProtectedSharesDeleteNamespacesList() (+2 more)

### Community 100 - ".checkUserManagerLimits"
Cohesion: 0.29
Nodes (6): delegateSetting(), Manager, limitText(), parseLimit(), parseUnitProps(), TestParseLimit()

### Community 102 - "Load"
Cohesion: 0.15
Nodes (27): minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadReadError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors() (+19 more)

### Community 103 - "github.com/spf13/cobra.ShellCompDirective"
Cohesion: 0.22
Nodes (8): completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), isEnvFileName(), matching(), TestRoleNames(), RoleNames()

### Community 104 - "Platform"
Cohesion: 0.14
Nodes (26): checkFlagPlatforms(), commandPlatforms(), declaredList(), flagOnlyOn(), parsePlatformList(), platformSuffix(), prepare(), promptPlatform() (+18 more)

### Community 105 - "internal/cli"
Cohesion: 0.14
Nodes (14): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, euid_test.go, examples_test.go (+6 more)

### Community 106 - "k8s/inspect.go"
Cohesion: 0.19
Nodes (18): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), operatorRunningImage(), ownedPods(), TestOperatorRunningImageWithNoContainers(), brokerList (+10 more)

### Community 107 - "deniedCluster"
Cohesion: 0.18
Nodes (13): TestValidateUnderADeniedClusterWideListReadsTheOperatorNamespace(), operatorImage(), TestOperatorApplyAsksBeforeWideningTheWatchList(), TestOperatorImage(), deniedCluster(), namedOperatorJSON(), operatorItem(), TestAmbiguousOperatorAbortsTheDeploy() (+5 more)

### Community 108 - "Import-Config Apply Rules"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "Config"
Cohesion: 0.08
Nodes (15): AdditionalUser, Image, Node, Redundancy, SEMP, TLS, atoiPrefix(), Config (+7 more)

### Community 110 - "internal/broker"
Cohesion: 0.08
Nodes (25): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, config_ops_test.go, coverage_test.go, diff_test.go, driver_test.go (+17 more)

### Community 111 - "GenOperator"
Cohesion: 0.16
Nodes (19): GenOperator(), joinYAMLDocs(), nonEmpty(), RenderOperator(), renderOperatorNS(), splitAfterNamespace(), splitOperatorBundle(), mustContain() (+11 more)

### Community 112 - "capRunner"
Cohesion: 0.09
Nodes (28): capCall, capRunner, callIndex(), TestPreflightRunsBeforeAnything(), mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime() (+20 more)

### Community 113 - "Config Package Tests"
Cohesion: 0.12
Nodes (17): adminsecret_test.go, artifactvalues_test.go, command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go (+9 more)

### Community 114 - "planSections"
Cohesion: 0.22
Nodes (10): planSections(), chunkNames(), TestBuildChunksCreatesVPNsBeforeBrokerSections(), TestBuildChunksOrdersTeardownClearCreateSectionsThenVPNs(), TestImportOpsImportApplyRefusesAChunkThatCarriesTheDriverDelimiter(), TestImportOpsImportApplyTransportFailureIsNotAConfigRejection(), TestImportOpsImportApplyUploadFailureRunsNothing(), TestImportOpsImportApplyWithNothingToApplyRunsNoDriver() (+2 more)

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
Cohesion: 0.23
Nodes (14): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+6 more)

### Community 120 - "Container"
Cohesion: 0.32
Nodes (5): Container, DockerConfig, Network, PodmanConfig, Ulimits

### Community 121 - "scripts.go"
Cohesion: 0.07
Nodes (40): showCmd, validCLILine(), Ops, TestDomainCertsLoadsInSortedNameOrder(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript() (+32 more)

### Community 122 - "hasCall"
Cohesion: 0.10
Nodes (38): hasCall(), TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestDisableDefaultUsersShowVPNError(), TestLeaderReadError(), TestReleaseToBackupReleasedTimeout(), reenableLine() (+30 more)

### Community 123 - ".validateContainerArtifactValues"
Cohesion: 0.20
Nodes (7): Config, portOrRange(), TestValidContainerPort(), validContainerPort(), validCoreLimit(), validHealthDuration(), validDNSLabel()

### Community 124 - "EnvPairs"
Cohesion: 0.18
Nodes (13): containerSecretSpecs(), EnvPairs(), groupKey(), itoa(), assertNoCheckoutPath(), envLines(), TestAdditionalUsersReachBothHalves(), TestGolden() (+5 more)

### Community 125 - "execBase"
Cohesion: 0.40
Nodes (5): checkFlagShape(), checkToken(), escalator(), execBase(), TestAllowCommandsRejectsEscalation()

### Community 126 - "watchCluster"
Cohesion: 0.22
Nodes (10): deployJSON(), TestFindOperatorDeploymentIgnoresOtherDeployments(), TestOperatorReleaseDeletesTheOperatorWhenTheEnvFileCoversEveryNamespace(), TestOperatorReleaseLeavesAnOperatorThatWatchesEverything(), TestOperatorReleaseNarrowingNeedsPermissionToPatch(), TestOperatorReleaseNarrowsInsteadOfRemoving(), TestReconcileWatchDecidesWhatToApply(), TestReconcileWatchFlagsAWideningAsAQuestion() (+2 more)

### Community 127 - ".applyScalingTierDefaults"
Cohesion: 0.22
Nodes (7): scalingTier, containerMem(), Config, TestContainerMem(), TestScalingTiers(), TestTierForRejectsOffTierValues(), tierFor()

### Community 128 - "importIgnore"
Cohesion: 0.47
Nodes (5): importIgnore(), TestImportIgnoreKeepsUnclassifiedSectionsInTheDiff(), TestImportOpsImportIgnore(), BridgeEnablementInverted(), TestShutdownBridgesLinesAreExcusedByTheDiff()

### Community 129 - "Scaling YAML Decoding"
Cohesion: 0.31
Nodes (5): scalingKey, scalingSpelling, Scaling, scalingKeyIndex(), scalingKeyList()

### Community 130 - "Role"
Cohesion: 0.05
Nodes (35): covUploadFailsTo, downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile (+27 more)

### Community 131 - ".preflightOne"
Cohesion: 0.32
Nodes (5): operatorEscalationProbes(), operatorProbes(), canIAnswer(), Cluster, probe

### Community 132 - "ReplVia"
Cohesion: 0.33
Nodes (6): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, validateReplVia(), validateReplViaSEMP()

### Community 134 - "shellScriptPath"
Cohesion: 0.40
Nodes (5): TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(), TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator(), TestExecShellScriptUploadsRunsWithBashAndRemovesIt(), shellScriptPath()

### Community 135 - "logArgs"
Cohesion: 0.50
Nodes (4): logArgs(), opCtrLogs(), logArgs2(), TestLogArgsBuildsOneSetForBothPlatforms()

### Community 136 - "runLeader"
Cohesion: 0.60
Nodes (4): leaderRun, runLeader(), TestAssertLeaderAsksBeforeActingFromABackup(), TestAssertLeaderPrimaryAndMonitorNeverAsk()

### Community 137 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 139 - "newBlock"
Cohesion: 0.24
Nodes (10): token, firstQuoted(), newBlock(), TestNewBlockMessageSpoolQualifierHasNoOwnName(), TestNewBlockOpener(), TestNewBlockVPNNameWithSpaces(), TestNewBlockVPNQualifierWithSpaces(), tokenizeOpener() (+2 more)

### Community 141 - "CheckHostPath"
Cohesion: 0.40
Nodes (5): CanonicalDuration(), TestCanonicalDuration(), CheckHostPath(), TestCheckHostPathAccepts(), TestCheckHostPathRejects()

### Community 155 - "ParseRole"
Cohesion: 0.09
Nodes (27): shorthand, renderAbbrevDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeResolutionNotes(), writeShorthandTable(), writeUnabbreviated(), TestAllowCommandIsRegisteredWhereItExecutes() (+19 more)

### Community 156 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

## Knowledge Gaps
- **243 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+238 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 363 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **18 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Role` connect `Role` to `bg`, `Manager`, `K8sConfig`, `containerTransport`, `ParseBlocks`, `matchCLI`, `Abbreviation Sets`, `renderDriver`, `ParseRole`, `ImportPlan`, `Config`, `Ops`, `App`, `Ops`, `VPNRepl`, `config_test.go`, `New`, `.rpc`, `opPerformReplication`, `newTestMate`, `Config`, `scripts.go`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._
- **Why does `Manager` connect `Manager` to `context.Context`, `App`, `Platform`, `Config`, `manager_test.go`, `config_test.go`, `confirm.go`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `newTestOps`, `Role`, `context.Context`, `cli_test.go`, `testing.T`, `Manager`, `K8sConfig`, `opArgvMatch`, `containerTransport`, `NewCluster`, `convert_test.go`, `allCustomMounted`, `manager_test.go`, `writeK8sDeployAllEnv`, `Compose`, `deploymentItem`, `ResolveDomainCerts`, `Ops`, `recRunner`, `eqArgs`, `localCfg`, `config_test.go`, `haCfg`, `Cluster`, `.configRows`, `strings.Builder`, `opPerformReplication`, `ReplSite`, `replCfg`, `GenSecrets`, `render_test.go`, `occCluster`, `Platform`, `deniedCluster`, `GenOperator`, `capRunner`, `certFixture`, `Container`, `EnvPairs`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Are the 112 inferred relationships involving `newTestOps()` (e.g. with `TestDomainCertsLoadsInSortedNameOrder()` and `TestEnableDefaultUsersStartsEveryDefaultUserBackUp()`) actually correct?**
  _`newTestOps()` has 112 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _243 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `bg` be split into smaller, more focused modules?**
  _Cohesion score 0.11764705882352941 - nodes in this community are weakly interconnected._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.05575065847234416 - nodes in this community are weakly interconnected._
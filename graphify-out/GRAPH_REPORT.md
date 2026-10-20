# Graph Report - solace-cnt-scripts  (2026-10-20)

## Corpus Check
- 200 files · ~687,236 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 4085 nodes · 17462 edges · 149 communities (134 shown, 15 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 2543 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `ae31f6ec`
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
- Block
- manager_test.go
- commands.go
- opArgvMatch
- Bash Dev Script
- transform_test.go
- PowerShell Dev Script
- Set
- BuildSwitchPlan
- newLocalOps
- helperCommand
- renderDriver
- .ConfigureReplication
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
- parseKV
- twoMates
- diff_test.go
- parseBody
- inject.go
- Ops
- exportconfig_test.go
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
- runExport
- .rpc
- confirm.go
- .configRows
- rootless_test.go
- strings.Builder
- output_test.go
- Configuration Guide
- New
- newTestMate
- Command Reference Doc
- Developer guide
- replApp
- Get
- cpuSetRange
- internal/k8s
- Docker and Podman mechanics
- usagef
- k8sOps
- Command
- newTeardownApplyFixture
- TestPodNameSuffixMatchesTheConfigBound
- ReplSite
- k8s/matechannel_test.go
- limits_test.go
- Removing a broker: what stays, what goes
- Troubleshooting
- k8s/inspect_test.go
- ParseShowReplication
- AdminSecret
- render_test.go
- main_test.go
- RuleFor
- platformOps
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
- deniedCluster
- Import-Config Apply Rules
- Config
- internal/broker
- GenOperator
- capRunner
- Config Package Tests
- wireExec
- operatorversion_test.go
- container/secrets_test.go
- internal/container
- Container State Inspection
- RenderDiffResult
- abbrev_test.go
- scripts.go
- inspectReply
- .validateContainerArtifactValues
- NodeAffinity
- MaskEnv
- watchCluster
- writeBlocks
- Scaling YAML Decoding
- Role
- .preflightOne
- ReplVia
- Namespace Protection Checks
- runLeader
- Fixtures and doubles
- newBlock
- CheckHostPath
- github.com/spf13/cobra.Command
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
- `sample()` --calls--> `New()`  [INFERRED]
  internal/abbrev/abbrev_test.go → internal/abbrev/abbrev.go
- `TestList()` --calls--> `New()`  [INFERRED]
  internal/abbrev/abbrev_test.go → internal/abbrev/abbrev.go
- `assertFencesBalanced()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/annotate.go
- `assertFencesBalanced()` --calls--> `markerVerb()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/annotate.go
- `parseBody()` --calls--> `readMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (149 total, 15 thin omitted)

### Community 0 - "bg"
Cohesion: 0.13
Nodes (38): TestConfirmDeleteShortcut(), confirmDelete(), warn(), bg(), k8sAdminPassword(), k8sCluster(), k8sContext(), k8sLogin() (+30 more)

### Community 1 - "newTestOps"
Cohesion: 0.03
Nodes (157): countContains(), hasCall(), leaderCalls(), matchCLI(), newTestOps(), ranContains(), TestCountContains(), TestDiagnostics() (+149 more)

### Community 2 - "Sink"
Cohesion: 0.20
Nodes (5): solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), columnWidths(), Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.05
Nodes (16): Exec, Transport, Manager, idMapCovers(), origin(), runUserIDs(), EnvRunner, Runner (+8 more)

### Community 4 - "cli_test.go"
Cohesion: 0.07
Nodes (63): capture(), captureStderr(), collectPaths(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone(), runStatusStderr() (+55 more)

### Community 5 - "testing.T"
Cohesion: 0.01
Nodes (238): TestLastLines(), TestExecShellScriptUploadFailureRunsNothing(), TestDiagnosticsRunError(), TestDomainCertsUploadFileError(), TestExecCLIFailsOnErrorOutput(), TestExecCLIUploadFileError(), TestGatherNodeDownloadError(), TestLastLinesEqualCount() (+230 more)

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
Cohesion: 0.08
Nodes (49): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+41 more)

### Community 10 - "CheckCommand"
Cohesion: 0.16
Nodes (21): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+13 more)

### Community 11 - "captureStdout"
Cohesion: 0.09
Nodes (33): opCall, opRunner, ReplicationConfigResult, captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook() (+25 more)

### Community 12 - "NewCluster"
Cohesion: 0.06
Nodes (90): TestCheckStopsProbingWhenUnreachable(), TestOperatorInstallVerdict(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsEveryFailureInOneRun() (+82 more)

### Community 13 - "ParseBlocks"
Cohesion: 0.09
Nodes (29): Annotate(), StripMarkers(), assertFencesBalanced(), TestAnnotateIsDeterministic(), TestAnnotateRegionAndSectionFencesAreBalanced(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), TestStripMarkersKeepsPragmaAndSectionComments() (+21 more)

### Community 14 - "convert_test.go"
Cohesion: 0.05
Nodes (67): doc, Result, segment, vars, DecodeStrict(), boolOf(), commentSafe(), Convert() (+59 more)

### Community 15 - "Block"
Cohesion: 0.13
Nodes (31): Block, TargetState, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects(), ClearTargetVirtualHostnames() (+23 more)

### Community 16 - "manager_test.go"
Cohesion: 0.07
Nodes (100): TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAStoppedContainer(), fileExists(), assertMode(), containsStr(), ctrCfg(), hasCall() (+92 more)

### Community 17 - "commands.go"
Cohesion: 0.25
Nodes (29): addCommands(), addLogFlags(), group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerCopyCmd(), newBrokerPerformCmd(), newBrokerRemoveCmd() (+21 more)

### Community 18 - "opArgvMatch"
Cohesion: 0.14
Nodes (32): opArgvMatch(), opCanI(), operatorManagedK8sCfg(), opFailOn(), TestDeployRefusesAnOperatorPasswordOverKeptData(), TestRemoveBrokerSaysWhereAnOperatorPasswordLives(), writeK8sDeployAllEnv(), opsK8sAnswer() (+24 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "transform_test.go"
Cohesion: 0.11
Nodes (31): InjectVPNServiceShutdown(), covAllLines(), covBlock(), covNestedBlocks(), covRemovals(), loadRealCapture(), TestClearExistingClientCAsIntersects(), TestClearExistingNestedClearsOnlyWhatBothSidesCarry() (+23 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Set"
Cohesion: 0.16
Nodes (8): Entry, checkShort(), Set, New(), TestDigitsAreAllowed(), TestNewPanicsOnAWiringBug(), newPlatforms(), newRoles()

### Community 23 - "BuildSwitchPlan"
Cohesion: 0.15
Nodes (23): PhaseKind, SiteState, SwitchAction, SwitchPhase, BuildSwitchPlan(), confirmDemotions(), SwitchPlan, promoteConfirmed() (+15 more)

### Community 24 - "newLocalOps"
Cohesion: 0.09
Nodes (39): TestLeaderAssertLeaderError(), TestRedundancyBackupShowError(), TestRedundancyReleaseError(), TestRedundancyRevertToPrimaryError(), TestReleaseToBackupNoReleaseError(), TestReleaseToBackupReleasedTimeout(), TestReleaseToBackupUnreleasedTimeout(), TestImportOpsExportConfigFailsLoudAtTheFirstBadCapture() (+31 more)

### Community 25 - "helperCommand"
Cohesion: 0.19
Nodes (14): TestExecEchoesOnEveryMethod(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), captureStdout(), helperCommand(), TestEchoDefaultWriter(), TestExecOutput(), TestExecOutputFail() (+6 more)

### Community 26 - "renderDriver"
Cohesion: 0.11
Nodes (22): chunk, chunkResult, runCLISkeleton(), chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce() (+14 more)

### Community 27 - ".ConfigureReplication"
Cohesion: 0.14
Nodes (14): BrokerType, cliMate, bannerType(), checkSameType(), Ops, isRunCLIRejection(), newlineIf(), replPhase1Rejected() (+6 more)

### Community 28 - "ImportPlan"
Cohesion: 0.12
Nodes (22): ImportResult, PlannedSection, isPragma(), defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList() (+14 more)

### Community 29 - "deploymentItem"
Cohesion: 0.13
Nodes (17): containsString(), installSteps(), desiredWatch(), Cluster, splitWatch(), subtractWatch(), TestDesiredWatchAppendsTheBrokerNamespace(), TestSubtractWatchEmptyRemainingMeansDelete() (+9 more)

### Community 31 - "Config"
Cohesion: 0.11
Nodes (14): keyValueEntries, checkCredentialChars(), foldToEnvVar(), Config, invertedCPUSetRange(), missingErr(), platformKey(), requireAll() (+6 more)

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
Cohesion: 0.18
Nodes (4): Ops, New(), TestNewDefaults(), TestTransportEchoHidesUploadBody()

### Community 36 - "App"
Cohesion: 0.11
Nodes (36): TestCtrManagerConfirmWiring(), resolveScriptPath(), confirmAssertFromBackup(), containerRenderRole(), containerRole(), containerWhat(), ctrLogin(), ctrManager() (+28 more)

### Community 37 - "recRunner"
Cohesion: 0.13
Nodes (10): TestCanIAnswerReadsTheLastLine(), NewTransport(), canIVerb(), TestTransportCopy(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest(), covApplyRunner (+2 more)

### Community 38 - "parseKV"
Cohesion: 0.07
Nodes (28): applyMetaField(), markerSection(), markerVerb(), parseKV(), readMarker(), renderKV(), splitKV(), TestApplyMetaFieldBrokerTypeSetsType() (+20 more)

### Community 39 - "twoMates"
Cohesion: 0.23
Nodes (20): ExecuteSwitchPlan(), SwitchPreflight(), ordersToA(), replSites(), TestExecuteSwitchPlanEmptyDoesNothing(), TestExecuteSwitchPlanOrder(), TestExecuteSwitchPlanRefusesASiteItHasNoChannelFor(), TestExecuteSwitchPlanRefusesToPromoteAgainstStaleState() (+12 more)

### Community 40 - "diff_test.go"
Cohesion: 0.15
Nodes (25): blockKey, mergedBlock, mergedLine, ancestorReported(), DiffBlocks(), indexByIdentity(), keyOf(), keySet() (+17 more)

### Community 41 - "parseBody"
Cohesion: 0.13
Nodes (16): Omission, Region, isMarker(), markerRegion(), TestIsMarkerDiscriminatesNamespace(), TestMarkerRegionBeginAndEnd(), Capture, isIndented() (+8 more)

### Community 42 - "inject.go"
Cohesion: 0.17
Nodes (27): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+19 more)

### Community 43 - "Ops"
Cohesion: 0.12
Nodes (15): field(), TestField(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), showRedundancyLocalScript(), defaultLocalAddrs(), Ops, hostMatches() (+7 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.10
Nodes (29): runFailRunner, TestExecute(), TestResolveScriptPath(), ExitCode(), childStatusError(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitStatusIsScopedToInteractiveSessions(), exportconfigDriverOutput() (+21 more)

### Community 45 - "Echo"
Cohesion: 0.26
Nodes (4): interactiveFailRunner, Echo, Quote(), TestQuote()

### Community 46 - "eqArgs"
Cohesion: 0.11
Nodes (30): TestReachable(), newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), saCfg(), TestCLIAndShellAreInteractive() (+22 more)

### Community 47 - "VPNRepl"
Cohesion: 0.09
Nodes (12): AdminState, QueueState, scriptedMate, replMate, rejectionIn(), ReplRole, VPNRepl, mateConvergenceShutdowns() (+4 more)

### Community 48 - "localCfg"
Cohesion: 0.08
Nodes (44): TestLocalMateRunsThroughOps(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK(), TestBackupRevertActivitySuccess() (+36 more)

### Community 49 - "config_test.go"
Cohesion: 0.19
Nodes (22): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, dropMessageSpoolLines(), dropRoutingInterfaceLine(), TestSectionsDropMessageSpoolLines(), TestSectionsDropRoutingInterfaceLine() (+14 more)

### Community 50 - "haCfg"
Cohesion: 0.04
Nodes (65): opK8sGenBroker(), AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestReadAdminPasswordArgvAndDecoding(), TestReadAdminPasswordGuardsTheCommand() (+57 more)

### Community 51 - "newRootCmd"
Cohesion: 0.07
Nodes (43): TestAbbreviationDocs(), applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias() (+35 more)

### Community 52 - ".Run"
Cohesion: 0.13
Nodes (28): allowRuntime(), runCtr(), TestAnnounceCommandsNamesResolvedBinaries(), TestAssertLeaderStandaloneSkipsBeforeAnyPrompt(), TestConfigStepsDoNotLeakSecrets(), TestCtrConfigDryRun(), TestCtrDiagnosticsDryRun(), TestCtrErrorPaths() (+20 more)

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
Cohesion: 0.16
Nodes (20): replRig, replNewRig(), replShowMate(), replUp(), replVPNTable(), TestOpConfigureReplicationAppliesTheRoleTheBrokersSiteDerives(), TestOpConfigureReplicationConvergesTheMateAndReportsWhatItStopped(), TestOpConfigureReplicationDeclinedLeavesTheBrokerUntouched() (+12 more)

### Community 58 - "runExport"
Cohesion: 0.14
Nodes (13): confirmImport(), pluralVPN(), runExport(), runImport(), TestPluralVPNWording(), emitOrWrite(), App, opCtrExportConfig() (+5 more)

### Community 59 - ".rpc"
Cohesion: 0.10
Nodes (19): backupTarget, Credential, sempMate, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigLine(), Ops, httpBody() (+11 more)

### Community 60 - "confirm.go"
Cohesion: 0.19
Nodes (18): exportconfigReadCounter, layer, TestPromptYes(), TestPromptYesNo(), TestStdinCanAnswerClosedFile(), confirmActionStrict(), confirmDowngrade(), confirmGate() (+10 more)

### Community 61 - ".configRows"
Cohesion: 0.12
Nodes (26): orNone(), orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminPassState(), adminSecretRow() (+18 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.17
Nodes (35): failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent(), TestCheckIDMappingRefusesAnUnmappedRunUser() (+27 more)

### Community 63 - "strings.Builder"
Cohesion: 0.18
Nodes (17): mdRow(), writeAbbrevTable(), LoadBalancer, PodAffinityTerm, cut(), sortedKeys(), splitUser(), writeCustomVolumeMounts() (+9 more)

### Community 64 - "output_test.go"
Cohesion: 0.19
Nodes (17): NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLevelOrderKeepsSkipBelowFail(), TestLineAddsNoPrefix() (+9 more)

### Community 65 - "Configuration Guide"
Cohesion: 0.14
Nodes (14): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+6 more)

### Community 66 - "New"
Cohesion: 0.09
Nodes (27): mateChannelFunc, TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportPlanReportShowsOmissionsAndEachSectionsVerdict(), TestImportResultReportQuotesTheFailedChunkAndItsTail(), TestImportResultReportsTeardownAndRebuildSeparately(), App, lineSink() (+19 more)

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

### Community 71 - "Get"
Cohesion: 0.24
Nodes (13): Example, TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), All(), Get(), List(), Names(), read() (+5 more)

### Community 72 - "cpuSetRange"
Cohesion: 0.50
Nodes (4): cpuSetCount(), cpuSetRange(), TestCPUSetCount(), TestCPUSetRange()

### Community 73 - "internal/k8s"
Cohesion: 0.10
Nodes (21): adminsecret_test.go, check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, generatorpage_test.go, inspect_test.go, internal/k8s (+13 more)

### Community 74 - "Docker and Podman mechanics"
Cohesion: 0.40
Nodes (5): Docker and Podman mechanics, Re-deploying is safe and explicit, Rootless podman prerequisites, What the broker cannot gain, Which user the broker runs as

### Community 75 - "usagef"
Cohesion: 0.09
Nodes (22): childExitError, usageError, ValidVPNName(), validVPNName(), logArgs(), newConvertCmd(), runConvert(), asUsage() (+14 more)

### Community 76 - "k8sOps"
Cohesion: 0.25
Nodes (18): confirmAction(), wantEnable(), wantRemove(), opCtrConfigDefaultUsers(), opCtrConfigDefaultVPN(), opCtrConfigDomainCerts(), k8sOps(), k8sWhat() (+10 more)

### Community 77 - "Command"
Cohesion: 0.14
Nodes (12): cmdField, guardedCmd, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandUnmarshal(), TestValidateProbeCommandAccepts(), TestValidateProbeCommandRejects(), Command (+4 more)

### Community 78 - "newTeardownApplyFixture"
Cohesion: 0.11
Nodes (24): planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), newTeardownApplyFixture(), TestBuildChunksCreatesVPNsBeforeBrokerSections() (+16 more)

### Community 79 - "TestPodNameSuffixMatchesTheConfigBound"
Cohesion: 0.12
Nodes (11): BrokerPodSuffixShape(), Cluster, TestOperatorAdminSecretNameMatchesTheBrokerNames(), Cluster, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes(), lbServiceName() (+3 more)

### Community 81 - "ReplSite"
Cohesion: 0.23
Nodes (10): Replication, missingListedVPNs(), PlannedRoles(), RoleAtSite(), TestPlannedRolesNamesEveryListedVPN(), TestRoleAtSiteIsTheComplement(), replVPNNames(), ReplSite (+2 more)

### Community 82 - "k8s/matechannel_test.go"
Cohesion: 0.26
Nodes (23): execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf(), replCfg(), replSite() (+15 more)

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

### Community 87 - "ParseShowReplication"
Cohesion: 0.07
Nodes (49): TestMateChannelShowReplication(), cliTransport(), containsEndpoint(), MateConfig, labelValue(), normTransport(), parseHostPort(), ParseShowReplication() (+41 more)

### Community 88 - "AdminSecret"
Cohesion: 0.16
Nodes (14): AdditionalUsersSecret(), AdminSecret(), dockerRegistrySecret(), decodeDataValue(), TestAdditionalUsersSecretCarriesBothHalves(), TestAdditionalUsersSecretIsAbsentWithNoUsers(), TestAdditionalUsersSecretRefusesAnEmptyPassword(), TestAdditionalUsersStayOutOfTheCredentialsSecret() (+6 more)

### Community 89 - "render_test.go"
Cohesion: 0.05
Nodes (85): ContainerNoFile(), HealthCheck, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), boolStr(), BrokerCR(), Compose(), composeEscape() (+77 more)

### Community 90 - "main_test.go"
Cohesion: 0.17
Nodes (18): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+10 more)

### Community 91 - "RuleFor"
Cohesion: 0.08
Nodes (36): Disposition, SectionRule, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified(), renderImportDocs() (+28 more)

### Community 92 - "platformOps"
Cohesion: 0.34
Nodes (18): addPodFlag(), newBrokerConfigureCmd(), newBrokerRestartCmd(), newConfigureDataReplicationCmd(), newConfigureDefaultUsersCmd(), newConfigureDefaultVPNCmd(), newConfigureDomainCertsCmd(), newConfigureProductKeysCmd() (+10 more)

### Community 93 - "Container"
Cohesion: 0.10
Nodes (19): Container, DockerConfig, Network, PodmanConfig, scalingTier, Ulimits, applyContainerBlockDefaults(), Config (+11 more)

### Community 94 - "Config"
Cohesion: 0.20
Nodes (6): expandTilde(), expandTildeToken(), Config, isPathSep(), TestExpandTilde(), tildeHome()

### Community 95 - "Secret Reference Resolution"
Cohesion: 0.50
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 96 - "MateChannel"
Cohesion: 0.13
Nodes (9): MateChannel, Ops, sortedKeys(), sortedVPNs(), curlConfigFlag(), Ops, requirePrimaryActive(), requireVirtualRouterNamesMatch() (+1 more)

### Community 97 - "replConfig"
Cohesion: 0.29
Nodes (10): replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationEmptyViaAccepted(), TestValidateReplicationPassSecretGrammar(), TestValidateReplicationRejects() (+2 more)

### Community 98 - "Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "age"
Cohesion: 0.26
Nodes (7): age(), roleRank(), Cluster, operatorRunningImage(), ownsPod(), TestOperatorRunningImageWithNoContainers(), TestOwnsPodFallsBackToTheNameInfix()

### Community 100 - ".checkUserManagerLimits"
Cohesion: 0.29
Nodes (6): delegateSetting(), Manager, limitText(), parseLimit(), parseUnitProps(), TestParseLimit()

### Community 102 - "Load"
Cohesion: 0.08
Nodes (43): TestServerCertBundleOrder(), TestServerCertBundleRefusesAKeyBearingCert(), TestServerCertBundleReportsAnUnreadableFile(), TestServerCertBundleRequiresBothHalves(), ServerCertBundle(), minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint() (+35 more)

### Community 103 - "ParseRole"
Cohesion: 0.29
Nodes (8): noRolePositional(), TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), TestRoleNames(), ParseRole(), RoleAbbrev(), RoleNames()

### Community 104 - "Platform"
Cohesion: 0.12
Nodes (29): checkFlagPlatforms(), commandPlatforms(), declaredList(), flagOnlyOn(), parsePlatformList(), platformSuffix(), prepare(), promptPlatform() (+21 more)

### Community 105 - "internal/cli"
Cohesion: 0.14
Nodes (14): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, euid_test.go, examples_test.go (+6 more)

### Community 106 - "k8s/inspect.go"
Cohesion: 0.19
Nodes (17): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), ownedPods(), brokerList, brokerStatus (+9 more)

### Community 107 - "deniedCluster"
Cohesion: 0.19
Nodes (12): TestValidateUnderADeniedClusterWideListReadsTheOperatorNamespace(), deniedCluster(), namedOperatorJSON(), operatorItem(), TestAmbiguousOperatorAbortsTheDeploy(), TestFindOperatorDeploymentFallsBackWhenTheClusterWideListFails(), TestFindOperatorDeploymentIgnoresOtherDeployments(), TestFindOperatorDeploymentIsScopedByNamespace() (+4 more)

### Community 108 - "Import-Config Apply Rules"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "Config"
Cohesion: 0.08
Nodes (17): AdditionalUser, Image, Node, Redundancy, SEMP, TLS, atoiPrefix(), Config (+9 more)

### Community 110 - "internal/broker"
Cohesion: 0.08
Nodes (25): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, config_ops_test.go, coverage_test.go, diff_test.go, driver_test.go (+17 more)

### Community 111 - "GenOperator"
Cohesion: 0.15
Nodes (21): GenOperator(), joinYAMLDocs(), nonEmpty(), operatorEscalationProbes(), operatorProbes(), RenderOperator(), renderOperatorNS(), splitAfterNamespace() (+13 more)

### Community 112 - "capRunner"
Cohesion: 0.09
Nodes (27): capCall, capRunner, NewManager(), callIndex(), TestManagerLogsCLIShell(), TestManagerNilSinks(), TestManagerPrepHostRootlessUsesUnshareChown(), TestPreflightRunsBeforeAnything() (+19 more)

### Community 113 - "Config Package Tests"
Cohesion: 0.12
Nodes (17): adminsecret_test.go, artifactvalues_test.go, command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go (+9 more)

### Community 114 - "wireExec"
Cohesion: 0.29
Nodes (13): opFunc, roleOpFunc, newBrokerDeployCmd(), newBrokerExecCmd(), newBrokerGenerateCmd(), newBrokerLogsCmd(), newBrokerStatusCmd(), newPerformCLIScriptCmd() (+5 more)

### Community 115 - "operatorversion_test.go"
Cohesion: 0.15
Nodes (18): operatorImage(), TestOperatorImage(), compareVersions(), Cluster, imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion() (+10 more)

### Community 116 - "container/secrets_test.go"
Cohesion: 0.16
Nodes (26): ResolveSecretValues(), bundleHash(), certCreate(), certFixture(), TestCertSecretLabelDrivesTheRestart(), TestCertSecretRemovalFailureIsFatal(), TestContainerBundleCarriesTheChain(), TestDeletePodmanToleratesAMissingBundle() (+18 more)

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.16
Nodes (14): containerState, healthState, decodeInspect(), Manager, orUnknown(), printable(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown() (+6 more)

### Community 119 - "RenderDiffResult"
Cohesion: 0.23
Nodes (11): BlockDiff, DiffResult, blockLabel(), diffRow(), orDash(), RenderDiffResult(), reportBlockLines(), TestDiffBlockLabelFormatsIdentityForTheReport() (+3 more)

### Community 120 - "abbrev_test.go"
Cohesion: 0.46
Nodes (7): sample(), TestAccessorsReturnCopies(), TestExpandResolvesBothSpellings(), TestKind(), TestList(), TestOrderIsDeclarationOrder(), TestShortOfUnknownWord()

### Community 121 - "scripts.go"
Cohesion: 0.08
Nodes (35): showCmd, validCLILine(), Ops, assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), enableDefaultUsersScript() (+27 more)

### Community 122 - "inspectReply"
Cohesion: 0.21
Nodes (8): inspectReply(), kvValue(), TestReportStateOnAHealthyRunningContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer(), TestRestartCountUnknownRatherThanZero(), TestRestartCountUnknownWhenSystemctlFails(), TestRestartCountUsesTheEngineOnDocker()

### Community 123 - ".validateContainerArtifactValues"
Cohesion: 0.14
Nodes (11): Config, portOrRange(), TestValidContainerPort(), validContainerPort(), validCoreLimit(), validHealthDuration(), CheckSecretKeyRef(), validDNSLabel() (+3 more)

### Community 124 - "NodeAffinity"
Cohesion: 0.38
Nodes (6): WeightedNodeTerm, NodeAffinity, NodeMatchExpr, Placement, validateMatchExprs(), validatePlacementAffinity()

### Community 125 - "MaskEnv"
Cohesion: 0.33
Nodes (6): TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestChildEnvNamesAreNotSystemVariables(), MaskEnv(), quoteTok(), TestMaskEnv(), TestQuoteTok()

### Community 126 - "watchCluster"
Cohesion: 0.25
Nodes (9): deployJSON(), TestOperatorReleaseDeletesTheOperatorWhenTheEnvFileCoversEveryNamespace(), TestOperatorReleaseLeavesAnOperatorThatWatchesEverything(), TestOperatorReleaseNarrowingNeedsPermissionToPatch(), TestOperatorReleaseNarrowsInsteadOfRemoving(), TestReconcileWatchDecidesWhatToApply(), TestReconcileWatchFlagsAWideningAsAQuestion(), TestSetWatchRefusesAnEmptyList() (+1 more)

### Community 127 - "writeBlocks"
Cohesion: 0.50
Nodes (4): sectionBeginMarker(), TestSectionBeginMarkerAdvisoryDisposition(), TestWriteBlocksOpensAndClosesOnChange(), writeBlocks()

### Community 129 - "Scaling YAML Decoding"
Cohesion: 0.31
Nodes (5): scalingKey, scalingSpelling, Scaling, scalingKeyIndex(), scalingKeyList()

### Community 130 - "Role"
Cohesion: 0.05
Nodes (34): covUploadFailsTo, downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile (+26 more)

### Community 131 - ".preflightOne"
Cohesion: 0.47
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "ReplVia"
Cohesion: 0.25
Nodes (7): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, viaDescription(), validateReplVia(), validateReplViaSEMP()

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

### Community 155 - "github.com/spf13/cobra.Command"
Cohesion: 0.08
Nodes (37): shorthand, renderAbbrevDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeResolutionNotes(), writeShorthandTable(), writeUnabbreviated(), TestAllowCommandIsRegisteredWhereItExecutes() (+29 more)

### Community 156 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

## Knowledge Gaps
- **243 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+238 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 363 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **15 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Role` connect `Role` to `bg`, `newTestOps`, `context.Context`, `Manager`, `K8sConfig`, `ParseBlocks`, `Set`, `renderDriver`, `.ConfigureReplication`, `ImportPlan`, `Config`, `Ops`, `App`, `Ops`, `config_test.go`, `runExport`, `.rpc`, `New`, `k8sOps`, `TestPodNameSuffixMatchesTheConfigBound`, `MateChannel`, `ParseRole`, `Config`, `scripts.go`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `newTestOps`, `Role`, `context.Context`, `testing.T`, `Manager`, `K8sConfig`, `captureStdout`, `NewCluster`, `convert_test.go`, `manager_test.go`, `opArgvMatch`, `deploymentItem`, `ResolveDomainCerts`, `Ops`, `recRunner`, `eqArgs`, `localCfg`, `config_test.go`, `haCfg`, `.Run`, `Cluster`, `.configRows`, `strings.Builder`, `New`, `TestPodNameSuffixMatchesTheConfigBound`, `ReplSite`, `k8s/matechannel_test.go`, `AdminSecret`, `render_test.go`, `Container`, `Load`, `Platform`, `GenOperator`, `capRunner`, `operatorversion_test.go`, `container/secrets_test.go`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Why does `Manager` connect `Manager` to `context.Context`, `App`, `Platform`, `Config`, `capRunner`, `config_test.go`, `confirm.go`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Are the 112 inferred relationships involving `newTestOps()` (e.g. with `TestDomainCertsLoadsInSortedNameOrder()` and `TestEnableDefaultUsersStartsEveryDefaultUserBackUp()`) actually correct?**
  _`newTestOps()` has 112 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _243 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `bg` be split into smaller, more focused modules?**
  _Cohesion score 0.1349527665317139 - nodes in this community are weakly interconnected._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.028175414678482162 - nodes in this community are weakly interconnected._
# Graph Report - solace-cnt-scripts  (2026-10-16)

## Corpus Check
- 200 files · ~682,594 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 4071 nodes · 17402 edges · 150 communities (134 shown, 16 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 2532 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `50cf51a3`
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
- GenSecrets
- Bash Dev Script
- coverage_test.go
- PowerShell Dev Script
- Abbreviation Sets
- MateChannel
- newLocalOps
- runner_test.go
- planSections
- Compose
- ImportPlan
- deploymentItem
- Solace Module Root
- Config
- Docs and Legacy Scripts
- ParseVPNReplication
- ResolveDomainCerts
- Role
- App
- recRunner
- parseBody
- completion_test.go
- diff_test.go
- ContainerSecret
- inject_test.go
- Ops
- exportconfig_test.go
- Echo
- eqArgs
- shellScriptPath
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
- Ops
- .rpc
- confirm.go
- .configRows
- rootless_test.go
- strings.Builder
- output_test.go
- Configuration Guide
- parseKV
- newTestMate
- Command Reference Doc
- Developer guide
- opPerformReplication
- TestServerCert
- writeK8sDeployAllEnv
- internal/k8s
- Docker and Podman mechanics
- exit.go
- .ConfigureReplication
- Command
- newTeardownApplyFixture
- writeNode
- ReplSite
- replCfg
- limits_test.go
- Removing a broker: what stays, what goes
- Troubleshooting
- k8s/inspect.go
- mateSoftware
- adminCfg
- render_test.go
- vulnjudge/main.go
- RuleFor
- registerFlagCompletion
- Container
- IsAbsHostPath
- Secret Reference Resolution
- github.com/spf13/cobra.Command
- replConfig
- Abbreviation Reference
- ParseShowReplication
- .checkUserManagerLimits
- platformTitle
- Load
- MateConfig
- Platform
- internal/cli
- .validateContainer
- deniedCluster
- Import-Config Apply Rules
- Config
- internal/broker
- GenOperator
- capRunner
- Config Package Tests
- RenderDiffResult
- operatorversion.go
- ExitCode
- internal/container
- Container State Inspection
- runPlatform
- Cluster
- scripts.go
- hasCall
- .validateContainerArtifactValues
- EnvPairs
- cliMate
- watchCluster
- .applyScalingTierDefaults
- github.com/spf13/cobra.ShellCompDirective
- Scaling YAML Decoding
- podName
- .preflightOne
- ReplVia
- Namespace Protection Checks
- storageCfg
- ContainerSecrets
- runLeader
- Fixtures and doubles
- logArgs
- Duration and Host Path Checks

## God Nodes (most connected - your core abstractions)
1. `newTestOps()` - 159 edges
2. `ctrCfg()` - 131 edges
3. `Role` - 129 edges
4. `NewCluster()` - 128 edges
5. `Config` - 113 edges
6. `newCapMgr()` - 111 edges
7. `loadK8s()` - 110 edges
8. `matchCLI()` - 92 edges
9. `Platform` - 79 edges
10. `bg()` - 70 edges

## Surprising Connections (you probably didn't know these)
- `parseBody()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `assertFencesBalanced()` --calls--> `markerVerb()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/annotate.go
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

## Communities (150 total, 16 thin omitted)

### Community 0 - "bg"
Cohesion: 0.12
Nodes (49): warn(), bg(), k8sAdminPassword(), k8sCluster(), k8sContext(), k8sLogin(), k8sOps(), k8sRestartForCertificate() (+41 more)

### Community 1 - "newTestOps"
Cohesion: 0.06
Nodes (68): leaderCalls(), matchCLI(), newTestOps(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultUsersRefusesAQuotedVPNName(), TestDisableDefaultVPN(), TestDomainCerts() (+60 more)

### Community 2 - "Sink"
Cohesion: 0.20
Nodes (5): solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), columnWidths(), Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.06
Nodes (13): scriptedMate, Exec, Manager, idMapCovers(), origin(), runUserIDs(), EnvRunner, Runner (+5 more)

### Community 4 - "cli_test.go"
Cohesion: 0.08
Nodes (59): capture(), captureStderr(), collectPaths(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone(), runStatusStderr() (+51 more)

### Community 5 - "testing.T"
Cohesion: 0.02
Nodes (204): TestExecShellScriptUploadFailureRunsNothing(), TestRouterNameReadsTheCaptureHeader(), bridgeHostPort(), sempPort(), sempV1OK(), TestSempPortBridgeFallsBackToPlaintext(), TestSempPortBridgePrefersTLS(), TestSempPortBridgeRefusesNoMapping() (+196 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (7): certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), secretSummary(), setOrMissing()

### Community 8 - "K8sConfig"
Cohesion: 0.12
Nodes (14): ContainerSecurity, Operator, PodSecurity, Resources, K8sConfig, LoadBalancer, cut(), parsePort() (+6 more)

### Community 9 - "runRootWith"
Cohesion: 0.09
Nodes (40): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+32 more)

### Community 10 - "CheckCommand"
Cohesion: 0.16
Nodes (21): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+13 more)

### Community 11 - "opArgvMatch"
Cohesion: 0.12
Nodes (34): opCall, captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch() (+26 more)

### Community 12 - "NewCluster"
Cohesion: 0.06
Nodes (83): TestCheckStopsProbingWhenUnreachable(), TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsEveryFailureInOneRun() (+75 more)

### Community 13 - "ParseBlocks"
Cohesion: 0.06
Nodes (42): BrokerType, Omission, Annotate(), isMarker(), StripMarkers(), assertFencesBalanced(), TestAnnotateIsDeterministic(), TestAnnotateRegionAndSectionFencesAreBalanced() (+34 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (65): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+57 more)

### Community 15 - "transform_test.go"
Cohesion: 0.08
Nodes (62): Block, TargetState, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects(), ClearTargetVirtualHostnames() (+54 more)

### Community 16 - "manager_test.go"
Cohesion: 0.05
Nodes (142): inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer() (+134 more)

### Community 17 - "commands.go"
Cohesion: 0.22
Nodes (45): opFunc, roleOpFunc, addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerConfigureCmd(), newBrokerCopyCmd() (+37 more)

### Community 18 - "GenSecrets"
Cohesion: 0.06
Nodes (41): GenSecrets(), TestCreateSecretsAllThree(), TestGenSecretsBuildsNothingWithoutMaterial(), TestGenSecretsStopsAtASecretItCannotBuild(), TestServerCertChanged(), TestUpdateServerCertSecret(), TestUpdateServerCertSecretStopsOnPreflightFailure(), tlsCfg() (+33 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "coverage_test.go"
Cohesion: 0.05
Nodes (44): ranContains(), TestLastLines(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultVPNDisableError() (+36 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Abbreviation Sets"
Cohesion: 0.13
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 23 - "MateChannel"
Cohesion: 0.07
Nodes (49): PhaseKind, SiteState, SwitchAction, SwitchPhase, MateChannel, curlConfigFlag(), Ops, BuildSwitchPlan() (+41 more)

### Community 24 - "newLocalOps"
Cohesion: 0.12
Nodes (33): TestLeaderAssertLeaderError(), TestRedundancyBackupShowError(), TestRedundancyReleaseError(), TestRedundancyRevertToPrimaryError(), TestReleaseToBackupNoReleaseError(), TestReleaseToBackupUnreleasedTimeout(), TestImportOpsExportConfigFailsLoudAtTheFirstBadCapture(), cliScriptPath() (+25 more)

### Community 25 - "runner_test.go"
Cohesion: 0.07
Nodes (38): TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), MaskEnv(), quoteTok(), captureStdout(), helperCommand() (+30 more)

### Community 26 - "planSections"
Cohesion: 0.08
Nodes (30): chunk, chunkResult, runCLISkeleton(), chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment() (+22 more)

### Community 27 - "Compose"
Cohesion: 0.12
Nodes (25): ContainerNoFile(), HealthCheck, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), Compose(), composeEscape(), ComposeProject(), escapePercent() (+17 more)

### Community 28 - "ImportPlan"
Cohesion: 0.12
Nodes (22): ImportResult, PlannedSection, isPragma(), defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList() (+14 more)

### Community 29 - "deploymentItem"
Cohesion: 0.13
Nodes (16): desiredWatch(), Cluster, splitWatch(), subtractWatch(), TestDesiredWatchAppendsTheBrokerNamespace(), TestSubtractWatchEmptyRemainingMeansDelete(), TestSubtractWatchPreservesOrder(), TestUnionWatchKeepsAnotherEnvFilesNamespace() (+8 more)

### Community 31 - "Config"
Cohesion: 0.11
Nodes (14): keyValueEntries, TestRoleNames(), RoleNames(), checkCredentialChars(), foldToEnvVar(), Config, missingErr(), requireKeyValue() (+6 more)

### Community 32 - "Docs and Legacy Scripts"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "ParseVPNReplication"
Cohesion: 0.10
Nodes (23): colSpan, dashSpans(), flagByte(), gutterClear(), ParseVPNReplication(), sliceSpan(), TestGutterClearAndSliceSpanBoundaries(), TestParseVPNReplication() (+15 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (39): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, caNameSafe(), checkNoDotDot(), DefaultDirReader() (+31 more)

### Community 35 - "Role"
Cohesion: 0.05
Nodes (27): covUploadFailsTo, downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile (+19 more)

### Community 36 - "App"
Cohesion: 0.07
Nodes (58): TestCtrManagerConfirmWiring(), confirmAction(), childExit(), TestChildExitKeepsItsMessage(), TestExitCodeIsNeverNegative(), usagef(), confirmImport(), pluralVPN() (+50 more)

### Community 37 - "recRunner"
Cohesion: 0.11
Nodes (13): TestCanIAnswerReadsTheLastLine(), TestExecutorRefusesUnapprovedRuntime(), unapprovedCfg(), TestBrokerReportPreviewStopsAtAnUnapprovedCommand(), NewTransport(), canIVerb(), TestTransportCopy(), TestTransportExecArgs() (+5 more)

### Community 38 - "parseBody"
Cohesion: 0.07
Nodes (32): Region, token, applyMetaField(), markerRegion(), markerSection(), markerVerb(), readMarker(), TestApplyMetaFieldBrokerTypeSetsType() (+24 more)

### Community 39 - "completion_test.go"
Cohesion: 0.18
Nodes (21): driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion(), TestCompletionDescriptionsAreOptIn() (+13 more)

### Community 40 - "diff_test.go"
Cohesion: 0.15
Nodes (25): blockKey, mergedBlock, mergedLine, ancestorReported(), DiffBlocks(), indexByIdentity(), keyOf(), keySet() (+17 more)

### Community 42 - "inject_test.go"
Cohesion: 0.15
Nodes (27): lineRole, serviceLine, shutdownStyle, svcKey, classifyServiceRest(), containsPortCommand(), describeKey(), fieldsUnquoted() (+19 more)

### Community 43 - "Ops"
Cohesion: 0.11
Nodes (18): countContains(), field(), TestCountContains(), TestField(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), TestZipConfigsScript(), zipConfigsScript() (+10 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.16
Nodes (18): opRunner, exportconfigDriverOutput(), exportconfigTransportOutput(), exportconfigVPNListOutput(), newExportconfigRunner(), newExportconfigRunnerRejectingChunk(), runImportConfirmation(), TestExportConfigOutOverwriteGate() (+10 more)

### Community 45 - "Echo"
Cohesion: 0.18
Nodes (7): interactiveFailRunner, runFailRunner, TestExecIsSilentWithoutVerbose(), Echo, NewExec(), Quote(), TestQuote()

### Community 46 - "eqArgs"
Cohesion: 0.10
Nodes (31): TestReachable(), newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), saCfg(), TestHARoles() (+23 more)

### Community 47 - "shellScriptPath"
Cohesion: 0.40
Nodes (5): TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(), TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator(), TestExecShellScriptUploadsRunsWithBashAndRemovesIt(), shellScriptPath()

### Community 48 - "localCfg"
Cohesion: 0.08
Nodes (44): TestLocalMateRunsThroughOps(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK(), TestBackupRevertActivitySuccess() (+36 more)

### Community 49 - "config_test.go"
Cohesion: 0.20
Nodes (20): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, span, dropMessageSpoolLines(), dropRoutingInterfaceLine(), TestSectionsDropMessageSpoolLines() (+12 more)

### Community 50 - "haCfg"
Cohesion: 0.09
Nodes (30): AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestReadAdminPasswordArgvAndDecoding(), TestReadAdminPasswordGuardsTheCommand(), TestReadAdminPasswordNamesTheNextStep() (+22 more)

### Community 51 - "newRootCmd"
Cohesion: 0.11
Nodes (22): TestAbbreviationDocs(), applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias() (+14 more)

### Community 52 - ".Run"
Cohesion: 0.14
Nodes (26): allowRuntime(), runCtr(), TestAnnounceCommandsNamesResolvedBinaries(), TestAssertLeaderStandaloneSkipsBeforeAnyPrompt(), TestConfigStepsDoNotLeakSecrets(), TestCtrConfigDryRun(), TestCtrDiagnosticsDryRun(), TestCtrErrorPaths() (+18 more)

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
Cohesion: 0.14
Nodes (4): Cluster, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

### Community 57 - "replNewRig"
Cohesion: 0.17
Nodes (19): replRig, replNewRig(), replShowMate(), replUp(), replVPNTable(), TestOpConfigureReplicationAppliesTheRoleTheBrokersSiteDerives(), TestOpConfigureReplicationConvergesTheMateAndReportsWhatItStopped(), TestOpConfigureReplicationDeclinedLeavesTheBrokerUntouched() (+11 more)

### Community 59 - ".rpc"
Cohesion: 0.11
Nodes (17): backupTarget, Credential, sempMate, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigLine(), Ops, httpBody() (+9 more)

### Community 60 - "confirm.go"
Cohesion: 0.17
Nodes (20): exportconfigReadCounter, layer, TestConfirmDeleteShortcut(), TestConfirmNonTTY(), TestPromptYes(), TestPromptYesNo(), TestStdinCanAnswerClosedFile(), confirmActionStrict() (+12 more)

### Community 61 - ".configRows"
Cohesion: 0.11
Nodes (29): orNone(), orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminPassState(), adminSecretRow() (+21 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.17
Nodes (35): failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent(), TestCheckIDMappingRefusesAnUnmappedRunUser() (+27 more)

### Community 63 - "strings.Builder"
Cohesion: 0.11
Nodes (26): shorthand, mdRow(), renderAbbrevDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable() (+18 more)

### Community 64 - "output_test.go"
Cohesion: 0.12
Nodes (24): TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportPlanReportShowsOmissionsAndEachSectionsVerdict(), TestImportResultReportQuotesTheFailedChunkAndItsTail(), TestImportResultReportsTeardownAndRebuildSeparately(), TestCheckReportSkipsEmptySections(), New(), NewFunc() (+16 more)

### Community 65 - "Configuration Guide"
Cohesion: 0.14
Nodes (14): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+6 more)

### Community 66 - "parseKV"
Cohesion: 0.17
Nodes (12): parseKV(), renderKV(), splitKV(), TestMarkerValueSurvivesAnEmbeddedQuote(), TestParseKVEmptyPayload(), TestParseKVIgnoresTokenWithNoEquals(), TestParseKVMultipleTokens(), TestParseKVQuotedValueKeepsSpaces() (+4 more)

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
Cohesion: 0.08
Nodes (37): mateChannelFunc, App, lineSink(), progress(), step(), confirmReplicationConfig(), confirmReplicationSwitch(), mateChannel() (+29 more)

### Community 71 - "TestServerCert"
Cohesion: 0.11
Nodes (19): TestDiagnostics(), TestPathHelpers(), TestServerCert(), TestServerCertBundleOrder(), TestServerCertBundleRefusesAKeyBearingCert(), TestServerCertBundleReportsAnUnreadableFile(), TestServerCertBundleRequiresBothHalves(), ServerCertBundle() (+11 more)

### Community 72 - "writeK8sDeployAllEnv"
Cohesion: 0.19
Nodes (24): writeK8sDeployAllEnv(), opsK8sAnswer(), opsK8sPods(), opsK8sRun(), opsK8sSilenced(), opsK8sUnattended(), TestOpsK8sConfigureLeavesRefuseOppositeDirections(), TestOpsK8sDomainCertsResolveBeforeAnythingIsAskedOrIssued() (+16 more)

### Community 73 - "internal/k8s"
Cohesion: 0.10
Nodes (21): adminsecret_test.go, check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, generatorpage_test.go, inspect_test.go, internal/k8s (+13 more)

### Community 74 - "Docker and Podman mechanics"
Cohesion: 0.40
Nodes (5): Docker and Podman mechanics, Re-deploying is safe and explicit, Rootless podman prerequisites, What the broker cannot gain, Which user the broker runs as

### Community 75 - "exit.go"
Cohesion: 0.12
Nodes (15): childExitError, usageError, ValidVPNName(), validVPNName(), asUsage(), childStatus(), isUsage(), markUsageArgs() (+7 more)

### Community 76 - ".ConfigureReplication"
Cohesion: 0.27
Nodes (7): Ops, isRunCLIRejection(), newlineIf(), replPhase1Rejected(), replPhase2Rejected(), replTransportFailure(), replVPNList()

### Community 77 - "Command"
Cohesion: 0.16
Nodes (10): cmdField, guardedCmd, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandUnmarshal(), Command, guardCommandOf(), setGuardCommand() (+2 more)

### Community 78 - "newTeardownApplyFixture"
Cohesion: 0.16
Nodes (16): driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), newTeardownApplyFixture(), TestImportOpsImportApplyClearsTargetOnlyObjectsBeforeAnythingIsApplied(), TestImportOpsImportApplyDriverExitingNonZeroAfterEveryChunkIsItsOwnFailure(), TestImportOpsImportApplyFailedTeardownCaptureStopsBeforeAnythingRuns() (+8 more)

### Community 79 - "writeNode"
Cohesion: 0.15
Nodes (17): WeightedNodeTerm, NodeAffinity, NodeMatchExpr, Placement, PodAffinityTerm, validateMatchExprs(), validatePlacementAffinity(), parseToleration() (+9 more)

### Community 81 - "ReplSite"
Cohesion: 0.07
Nodes (35): AdminState, QueueState, replMate, Replication, ReplRole, VPNRepl, sortedKeys(), ReplicationConfigResult (+27 more)

### Community 82 - "replCfg"
Cohesion: 0.19
Nodes (23): CLIScriptPath(), execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf(), replCfg() (+15 more)

### Community 83 - "limits_test.go"
Cohesion: 0.13
Nodes (27): missingControllers(), limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessCoreFloorFollowsTheConfiguredLimit(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager() (+19 more)

### Community 84 - "Removing a broker: what stays, what goes"
Cohesion: 0.40
Nodes (5): Every destructive command confirms, Removing a broker: what stays, what goes, Removing the operator does not always remove it, The layer flag raises the question, The namespace is only offered when it is empty

### Community 85 - "Troubleshooting"
Cohesion: 0.22
Nodes (9): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Forbidden: cannot list deployments at the cluster scope, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting (+1 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.08
Nodes (42): age(), anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), podHealth(), pvcLevel(), replicaLevel() (+34 more)

### Community 87 - "mateSoftware"
Cohesion: 0.16
Nodes (18): TestMateChannelShowReplication(), containsEndpoint(), normTransport(), RenderMateRemovals(), SameEndpoints(), mateAppliance(), mateSoftware(), TestParseShowReplicationAppliance() (+10 more)

### Community 88 - "adminCfg"
Cohesion: 0.12
Nodes (18): opK8sGenBroker(), GenBroker(), joinManifests(), adminCfg(), secretNamesIn(), TestAdminSecretStatesEmitTheirArtifacts(), TestCreateSecretsAdminOnly(), TestCreateSecretsPreflight() (+10 more)

### Community 89 - "render_test.go"
Cohesion: 0.14
Nodes (34): boolStr(), BrokerCR(), containerArtifacts(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets(), TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPrivilegePosture() (+26 more)

### Community 90 - "vulnjudge/main.go"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.07
Nodes (40): Disposition, SectionRule, sectionBeginMarker(), TestSectionBeginMarkerAdvisoryDisposition(), TestWriteBlocksOpensAndClosesOnChange(), writeBlocks(), loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance() (+32 more)

### Community 92 - "registerFlagCompletion"
Cohesion: 0.19
Nodes (11): registerFlagCompletion(), newConvertCmd(), runConvert(), newExamplesCmd(), runExample(), addAllowCommandFlag(), addDisableEnableFlags(), addExportFlags() (+3 more)

### Community 93 - "Container"
Cohesion: 0.16
Nodes (12): Container, DockerConfig, Network, PodmanConfig, Ulimits, applyContainerBlockDefaults(), Config, podmanRunUser() (+4 more)

### Community 94 - "IsAbsHostPath"
Cohesion: 0.16
Nodes (9): expandTilde(), expandTildeToken(), Config, IsAbsHostPath(), isPathSep(), TestContainerHostDirsExpandATilde(), TestExpandTilde(), TestIsAbsHostPath() (+1 more)

### Community 95 - "Secret Reference Resolution"
Cohesion: 0.50
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 96 - "github.com/spf13/cobra.Command"
Cohesion: 0.27
Nodes (18): addCommands(), addLogFlags(), newOperatorCmd(), newOperatorDeployCmd(), newOperatorGenerateCmd(), newOperatorLogsCmd(), newOperatorRemoveCmd(), newOperatorRestartCmd() (+10 more)

### Community 97 - "replConfig"
Cohesion: 0.31
Nodes (9): replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationEmptyViaAccepted(), TestValidateReplicationRejects(), TestValidateReplicationValid() (+1 more)

### Community 98 - "Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "ParseShowReplication"
Cohesion: 0.18
Nodes (16): ParseShowReplication(), RenderMate(), loadShowReplication(), TestMateRoundTripsThroughItsOwnGrammar(), TestParseShowReplicationRejectsAnUnreadableAddressOrPort(), TestParseShowReplicationUnconfigured(), TestParseShowReplicationWrongPlatformFails(), TestRealApplianceCaptureMateRoundTrips() (+8 more)

### Community 100 - ".checkUserManagerLimits"
Cohesion: 0.29
Nodes (6): delegateSetting(), Manager, limitText(), parseLimit(), parseUnitProps(), TestParseLimit()

### Community 102 - "Load"
Cohesion: 0.07
Nodes (53): Example, TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadReadError() (+45 more)

### Community 103 - "MateConfig"
Cohesion: 0.17
Nodes (13): cliTransport(), MateConfig, labelValue(), parseHostPort(), parseShowReplicationAppliance(), parseShowReplicationSoftware(), renderMateAppliance(), renderMateSoftware() (+5 more)

### Community 104 - "Platform"
Cohesion: 0.12
Nodes (29): checkFlagPlatforms(), commandPlatforms(), declaredList(), flagOnlyOn(), parsePlatformList(), platformSuffix(), prepare(), promptPlatform() (+21 more)

### Community 105 - "internal/cli"
Cohesion: 0.14
Nodes (14): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, euid_test.go, examples_test.go (+6 more)

### Community 106 - ".validateContainer"
Cohesion: 0.16
Nodes (11): TestValidateProbeCommandAccepts(), TestValidateProbeCommandRejects(), cpuSetCount(), cpuSetRange(), TestCPUSetCount(), TestCPUSetRange(), invertedCPUSetRange(), platformKey() (+3 more)

### Community 107 - "deniedCluster"
Cohesion: 0.18
Nodes (14): TestValidateUnderADeniedClusterWideListReadsTheOperatorNamespace(), TestOperatorApplyProceedsWhenTheClusterWideListIsDenied(), TestConfirmNoDowngradeStillAsksWhenTheFallbackFindsANewerOperator(), deniedCluster(), deniedRunner(), namedOperatorJSON(), operatorItem(), TestAmbiguousOperatorAbortsTheDeploy() (+6 more)

### Community 108 - "Import-Config Apply Rules"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "Config"
Cohesion: 0.10
Nodes (10): AdditionalUser, Image, Node, Redundancy, SEMP, TLS, atoiPrefix(), Config (+2 more)

### Community 110 - "internal/broker"
Cohesion: 0.08
Nodes (25): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, config_ops_test.go, coverage_test.go, diff_test.go, driver_test.go (+17 more)

### Community 111 - "GenOperator"
Cohesion: 0.11
Nodes (26): TestOperatorInstallVerdict(), TestValidateSparseConfigExplainsItself(), GenOperator(), joinYAMLDocs(), nonEmpty(), operatorEscalationProbes(), operatorImage(), operatorProbes() (+18 more)

### Community 112 - "capRunner"
Cohesion: 0.11
Nodes (24): capCall, capRunner, New(), TestNewDefaults(), mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime() (+16 more)

### Community 113 - "Config Package Tests"
Cohesion: 0.12
Nodes (17): adminsecret_test.go, artifactvalues_test.go, command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go (+9 more)

### Community 114 - "RenderDiffResult"
Cohesion: 0.23
Nodes (11): BlockDiff, DiffResult, blockLabel(), diffRow(), orDash(), RenderDiffResult(), reportBlockLines(), TestDiffBlockLabelFormatsIdentityForTheReport() (+3 more)

### Community 115 - "operatorversion.go"
Cohesion: 0.23
Nodes (10): compareVersions(), Cluster, imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestImageTag() (+2 more)

### Community 116 - "ExitCode"
Cohesion: 0.18
Nodes (12): TestExecute(), TestResolveScriptPath(), TestScriptCommandsRefuseBeforeAnythingRuns(), ExitCode(), childStatusError(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitStatusIsScopedToInteractiveSessions(), TestExportConfigScopeConflictIsUsageError() (+4 more)

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
Cohesion: 0.08
Nodes (35): showCmd, validCLILine(), Ops, TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript() (+27 more)

### Community 122 - "hasCall"
Cohesion: 0.15
Nodes (27): hasCall(), TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestDisableDefaultUsersShowVPNError(), TestLeaderReadError(), TestReleaseToBackupReleasedTimeout(), replListed(), replReadOnlyResponder() (+19 more)

### Community 123 - ".validateContainerArtifactValues"
Cohesion: 0.20
Nodes (7): Config, portOrRange(), TestValidContainerPort(), validContainerPort(), validCoreLimit(), validHealthDuration(), validDNSLabel()

### Community 124 - "EnvPairs"
Cohesion: 0.24
Nodes (10): EnvPairs(), groupKey(), itoa(), assertNoCheckoutPath(), envLines(), TestAdditionalUsersReachBothHalves(), TestGolden(), TestServerCertFilePathKeyKeepsItsUnderscore() (+2 more)

### Community 125 - "cliMate"
Cohesion: 0.36
Nodes (3): cliMate, showReplicationScript(), showVPNReplicationScript()

### Community 126 - "watchCluster"
Cohesion: 0.25
Nodes (9): deployJSON(), TestOperatorReleaseDeletesTheOperatorWhenTheEnvFileCoversEveryNamespace(), TestOperatorReleaseLeavesAnOperatorThatWatchesEverything(), TestOperatorReleaseNarrowingNeedsPermissionToPatch(), TestOperatorReleaseNarrowsInsteadOfRemoving(), TestReconcileWatchDecidesWhatToApply(), TestReconcileWatchFlagsAWideningAsAQuestion(), TestSetWatchRefusesAnEmptyList() (+1 more)

### Community 127 - ".applyScalingTierDefaults"
Cohesion: 0.22
Nodes (7): scalingTier, containerMem(), Config, TestContainerMem(), TestScalingTiers(), TestTierForRejectsOffTierValues(), tierFor()

### Community 128 - "github.com/spf13/cobra.ShellCompDirective"
Cohesion: 0.29
Nodes (6): completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), isEnvFileName(), matching()

### Community 129 - "Scaling YAML Decoding"
Cohesion: 0.31
Nodes (5): scalingKey, scalingSpelling, Scaling, scalingKeyIndex(), scalingKeyList()

### Community 130 - "podName"
Cohesion: 0.08
Nodes (18): BrokerPodSuffixShape(), Cluster, TestOperatorAdminSecretNameMatchesTheBrokerNames(), Cluster, TestPodNameForIsTheOneSpelling(), HARoles(), lbServiceName(), podName() (+10 more)

### Community 131 - ".preflightOne"
Cohesion: 0.47
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "ReplVia"
Cohesion: 0.29
Nodes (7): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, viaDescription(), validateReplVia(), validateReplViaSEMP()

### Community 134 - "storageCfg"
Cohesion: 0.25
Nodes (7): storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestStorageClassAndCustomMountAreMutuallyExclusive()

### Community 135 - "ContainerSecrets"
Cohesion: 0.29
Nodes (7): ContainerSecrets(), containerSecretSpecs(), SecretPreflight(), TestComposeLabelsTheCertificateDigest(), TestFileBackedSecretIsExemptFromSecretPreflight(), TestServerCertIsASecretOnBothEngines(), secretSpec

### Community 136 - "runLeader"
Cohesion: 0.60
Nodes (4): leaderRun, runLeader(), TestAssertLeaderAsksBeforeActingFromABackup(), TestAssertLeaderPrimaryAndMonitorNeverAsk()

### Community 137 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 138 - "logArgs"
Cohesion: 0.67
Nodes (3): logArgs(), logArgs2(), TestLogArgsBuildsOneSetForBothPlatforms()

### Community 141 - "Duration and Host Path Checks"
Cohesion: 0.40
Nodes (5): CanonicalDuration(), TestCanonicalDuration(), CheckHostPath(), TestCheckHostPathAccepts(), TestCheckHostPathRejects()

## Knowledge Gaps
- **243 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+238 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 363 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **16 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Config` connect `Config` to `newTestOps`, `podName`, `testing.T`, `Manager`, `K8sConfig`, `ContainerSecrets`, `opArgvMatch`, `NewCluster`, `convert_test.go`, `manager_test.go`, `GenSecrets`, `runner_test.go`, `Compose`, `deploymentItem`, `ResolveDomainCerts`, `Role`, `recRunner`, `eqArgs`, `localCfg`, `config_test.go`, `haCfg`, `.Run`, `Cluster`, `Ops`, `.configRows`, `opPerformReplication`, `TestServerCert`, `writeK8sDeployAllEnv`, `Command`, `writeNode`, `ReplSite`, `replCfg`, `adminCfg`, `render_test.go`, `Container`, `Platform`, `GenOperator`, `capRunner`, `Cluster`, `EnvPairs`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `newTestOps`, `podName`, `Manager`, `ParseBlocks`, `Abbreviation Sets`, `planSections`, `ImportPlan`, `Config`, `App`, `Ops`, `config_test.go`, `haCfg`, `Ops`, `.rpc`, `strings.Builder`, `newTestMate`, `opPerformReplication`, `TestServerCert`, `.ConfigureReplication`, `newTeardownApplyFixture`, `scripts.go`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Why does `Platform` connect `Platform` to `testing.T`, `Manager`, `runLeader`, `ContainerSecrets`, `CheckCommand`, `opArgvMatch`, `convert_test.go`, `manager_test.go`, `commands.go`, `Config`, `Role`, `config_test.go`, `Ops`, `opPerformReplication`, `Command`, `limits_test.go`, `render_test.go`, `Container`, `IsAbsHostPath`, `platformTitle`, `Load`, `.validateContainer`, `capRunner`, `.validateContainerArtifactValues`, `.applyScalingTierDefaults`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Are the 109 inferred relationships involving `newTestOps()` (e.g. with `TestDomainCertsLoadsInSortedNameOrder()` and `TestEnableDefaultUsersStartsEveryDefaultUserBackUp()`) actually correct?**
  _`newTestOps()` has 109 INFERRED edges - model-reasoned connections that need verification._
- **Are the 49 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 49 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _243 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `bg` be split into smaller, more focused modules?**
  _Cohesion score 0.12408163265306123 - nodes in this community are weakly interconnected._
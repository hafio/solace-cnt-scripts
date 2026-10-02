# Graph Report - solace-cnt-scripts  (2026-10-02)

## Corpus Check
- 197 files · ~627,305 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 3844 nodes · 16203 edges · 142 communities (126 shown, 16 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 2213 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `4b81fb80`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ops_k8s.go
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
- ParseBlocks
- convert_test.go
- transform_test.go
- manager_test.go
- commands.go
- GenSecrets
- Bash Dev Script
- RenderDiffResult
- PowerShell Dev Script
- Abbreviation Sets
- BuildSwitchPlan
- verify_local_test.go
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
- hasCall
- Echo
- k8s/inspect.go
- VPNRepl
- localCfg
- config_test.go
- BrokerType
- newRootCmd
- exportconfig_test.go
- Container Security Guidelines
- Operations
- Test catalogue
- Cluster
- ReplSite
- validName
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
- opPerformReplication
- watchCluster
- ParseRole
- internal/k8s
- importIgnore
- usagef
- .ConfigureReplication
- Command
- newTeardownApplyFixture
- k8s/matechannel_test.go
- limits_test.go
- occCluster
- Troubleshooting Guide
- k8s/inspect_test.go
- MateConfig
- haCfg
- render_test.go
- main_test.go
- sections_test.go
- step
- Container
- Config
- Secret Reference Resolution
- github.com/spf13/cobra.Command
- replConfig
- Abbreviation Reference
- blocks.go
- .checkUserManagerLimits
- platformTitle
- Load
- ctrCfg
- Platform
- internal/cli
- Compose
- Import-Config Apply Rules
- Config
- internal/broker
- GenOperator
- capRunner
- Config Package Tests
- operatorversion.go
- EnvPairs
- internal/container
- Container State Inspection
- scripts.go
- coverage_test.go
- .validateContainerArtifactValues
- newOperatorCmd
- .validateContainer
- Manager
- Scaling YAML Decoding
- podName
- k8s/preflight.go
- ReplVia
- Namespace Protection Checks
- lastLines
- productKeysScript
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
6. `NewCluster()` - 100 edges
7. `loadK8s()` - 89 edges
8. `Platform` - 79 edges
9. `matchCLI()` - 75 edges
10. `bg()` - 70 edges

## Surprising Connections (you probably didn't know these)
- `assertFencesBalanced()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/annotate.go
- `parseBody()` --calls--> `readMarker()`  [INFERRED]
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

## Communities (142 total, 16 thin omitted)

### Community 0 - "ops_k8s.go"
Cohesion: 0.15
Nodes (50): bg(), domainCANames(), k8sAdminPassword(), k8sCluster(), k8sContext(), k8sLogin(), k8sOps(), k8sRestartForCertificate() (+42 more)

### Community 1 - "newTestOps"
Cohesion: 0.07
Nodes (54): countContains(), leaderCalls(), newTestOps(), TestCountContains(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultUsersRefusesAQuotedVPNName() (+46 more)

### Community 2 - "Sink"
Cohesion: 0.18
Nodes (6): sectionLevel(), columnWidths(), Level, Sink, pad(), TestLevelTagsMapToTheHouseTags()

### Community 3 - "context.Context"
Cohesion: 0.06
Nodes (12): scriptedMate, Exec, EnvRunner, Runner, Cluster, Cluster, Cluster, namespaceManifest() (+4 more)

### Community 4 - "cli_test.go"
Cohesion: 0.05
Nodes (108): TestAllowCommandRejectedWhereNothingExecutes(), allowRuntime(), captureStderr(), echoRunner(), fakeBinaryOnPath(), firstLine(), runCtr(), runRoot() (+100 more)

### Community 5 - "testing.T"
Cohesion: 0.02
Nodes (201): TestRouterNameReadsTheCaptureHeader(), bridgeHostPort(), sempPort(), TestSempPortBridgeFallsBackToPlaintext(), TestSempPortBridgePrefersTLS(), TestSempPortBridgeRefusesNoMapping(), TestSempPortHostMode(), TestSempPortWithoutCertificateStaysPlaintext() (+193 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (9): certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), secretSummary(), setOrMissing(), ResolveSecretValues() (+1 more)

### Community 8 - "K8sConfig"
Cohesion: 0.17
Nodes (7): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, writeSecurity()

### Community 9 - "Quadlet"
Cohesion: 0.17
Nodes (16): ContainerNoFile(), HealthCheck, TestLimitsCheckAssertsWhatTheArtifactAsks(), escapePercent(), healthCmd(), Quadlet(), quadletEscape(), assertNoCheckoutPath() (+8 more)

### Community 10 - "CheckCommand"
Cohesion: 0.17
Nodes (20): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+12 more)

### Community 11 - "captureStdout"
Cohesion: 0.08
Nodes (43): leaderRun, opCall, opRunner, capture(), captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod() (+35 more)

### Community 12 - "NewCluster"
Cohesion: 0.06
Nodes (72): TestCheckStopsProbingWhenUnreachable(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+64 more)

### Community 13 - "ParseBlocks"
Cohesion: 0.16
Nodes (17): Annotate(), TestAnnotateIsDeterministic(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), ParseBlocks(), blockByIndex(), loadSample(), TestParseBlocksCRLFMatchesLF() (+9 more)

### Community 14 - "convert_test.go"
Cohesion: 0.05
Nodes (67): doc, Result, segment, vars, DecodeStrict(), boolOf(), commentSafe(), Convert() (+59 more)

### Community 15 - "transform_test.go"
Cohesion: 0.09
Nodes (50): Block, TargetState, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects(), ClearTargetVirtualHostnames() (+42 more)

### Community 16 - "manager_test.go"
Cohesion: 0.07
Nodes (81): TestRestartCountUnknownWhenSystemctlFails(), fileExists(), assertMode(), containsStr(), hasCall(), maskedKeys(), newCapMgr(), TestContainerRunningMatchesNameExactly() (+73 more)

### Community 17 - "commands.go"
Cohesion: 0.20
Nodes (48): roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerConfigureCmd() (+40 more)

### Community 18 - "GenSecrets"
Cohesion: 0.06
Nodes (39): GenBroker(), GenSecrets(), joinManifests(), adminCfg(), secretNamesIn(), TestAdminSecretStatesEmitTheirArtifacts(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree() (+31 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "RenderDiffResult"
Cohesion: 0.23
Nodes (11): BlockDiff, DiffResult, blockLabel(), diffRow(), orDash(), RenderDiffResult(), reportBlockLines(), TestDiffBlockLabelFormatsIdentityForTheReport() (+3 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Abbreviation Sets"
Cohesion: 0.13
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 23 - "BuildSwitchPlan"
Cohesion: 0.13
Nodes (28): PhaseKind, SiteState, SwitchAction, SwitchPhase, BuildSwitchPlan(), confirmDemotions(), ExecuteSwitchPlan(), SwitchPlan (+20 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.13
Nodes (38): TestPathHelpers(), TestLeaderAssertLeaderError(), TestRedundancyReleaseError(), TestRedundancyRevertToPrimaryError(), CLIArg(), cliArg(), cliScriptPath(), isCurl() (+30 more)

### Community 25 - "runner_test.go"
Cohesion: 0.10
Nodes (30): TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), MaskEnv(), NewExec(), quoteTok() (+22 more)

### Community 26 - "renderDriver"
Cohesion: 0.11
Nodes (21): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), shQuote() (+13 more)

### Community 27 - "eqArgs"
Cohesion: 0.11
Nodes (29): TestReachable(), newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), saCfg(), TestCLIAndShellAreInteractive() (+21 more)

### Community 28 - "ImportPlan"
Cohesion: 0.14
Nodes (19): ImportResult, PlannedSection, isPragma(), defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList() (+11 more)

### Community 29 - "desiredWatch"
Cohesion: 0.31
Nodes (5): desiredWatch(), Cluster, splitWatch(), TestDesiredWatchAppendsTheBrokerNamespace(), watchPlan

### Community 31 - "Config"
Cohesion: 0.12
Nodes (13): keyValueEntries, TestRoleNames(), RoleNames(), checkCredentialChars(), foldToEnvVar(), Config, missingErr(), requireAll() (+5 more)

### Community 32 - "Docs and Legacy Scripts"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "ParseVPNReplication"
Cohesion: 0.12
Nodes (19): colSpan, dashSpans(), flagByte(), gutterClear(), ParseVPNReplication(), sliceSpan(), TestParseVPNReplication(), TestParseVPNReplicationIgnoresRepeatedHeaders() (+11 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (40): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, caNameSafe(), checkNoDotDot(), DefaultDirReader() (+32 more)

### Community 35 - "Role"
Cohesion: 0.07
Nodes (20): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrTransport (+12 more)

### Community 36 - "App"
Cohesion: 0.11
Nodes (43): confirmAction(), warn(), childExit(), TestChildExitKeepsItsMessage(), wantEnable(), wantRemove(), resolveScriptPath(), confirmAssertFromBackup() (+35 more)

### Community 37 - "recRunner"
Cohesion: 0.11
Nodes (13): New(), TestNewDefaults(), TestCanIAnswerReadsTheLastLine(), TestExecutorRefusesUnapprovedRuntime(), unapprovedCfg(), NewTransport(), TestTransportCopy(), TestTransportEchoHidesUploadBody() (+5 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.09
Nodes (34): applyMetaField(), markerRegion(), markerSection(), markerVerb(), parseKV(), readMarker(), renderKV(), sectionBeginMarker() (+26 more)

### Community 39 - "completion_test.go"
Cohesion: 0.18
Nodes (21): driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion(), TestCompletionDescriptionsAreOptIn() (+13 more)

### Community 40 - "diff_test.go"
Cohesion: 0.15
Nodes (25): blockKey, mergedBlock, mergedLine, ancestorReported(), DiffBlocks(), indexByIdentity(), keyOf(), keySet() (+17 more)

### Community 41 - "runPlatform"
Cohesion: 0.21
Nodes (14): logArgs2(), runPlatform(), TestLogArgsBuildsOneSetForBothPlatforms(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations() (+6 more)

### Community 42 - "inject_test.go"
Cohesion: 0.15
Nodes (27): lineRole, serviceLine, shutdownStyle, svcKey, classifyServiceRest(), containsPortCommand(), describeKey(), fieldsUnquoted() (+19 more)

### Community 43 - "Ops"
Cohesion: 0.13
Nodes (13): field(), TestField(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), defaultLocalAddrs(), Ops, hostMatches(), shortHost() (+5 more)

### Community 44 - "hasCall"
Cohesion: 0.21
Nodes (17): hasCall(), TestDisableDefaultUsersShowVPNError(), TestLeaderReadError(), TestReleaseToBackupReleasedTimeout(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching() (+9 more)

### Community 45 - "Echo"
Cohesion: 0.24
Nodes (5): interactiveFailRunner, runFailRunner, Echo, Quote(), TestQuote()

### Community 46 - "k8s/inspect.go"
Cohesion: 0.14
Nodes (23): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), operatorRunningImage(), ownedPods(), TestOperatorRunningImageWithNoContainers() (+15 more)

### Community 47 - "VPNRepl"
Cohesion: 0.09
Nodes (31): AdminState, QueueState, ReplRole, VPNRepl, sortedKeys(), sortedSet(), ReplicationConfigResult, mateConvergenceShutdowns() (+23 more)

### Community 48 - "localCfg"
Cohesion: 0.09
Nodes (36): TestLocalMateRunsThroughOps(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK(), TestBackupRevertActivitySuccess() (+28 more)

### Community 49 - "config_test.go"
Cohesion: 0.19
Nodes (19): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, span, Scaling, generatorDefaults(), scalingOf() (+11 more)

### Community 50 - "BrokerType"
Cohesion: 0.11
Nodes (20): BrokerType, StripMarkers(), TestStripMarkersKeepsPragmaAndSectionComments(), TestStripMarkersNoMarkersPassthrough(), splitLines(), ValidVPNName(), validVPNName(), bannerType() (+12 more)

### Community 51 - "newRootCmd"
Cohesion: 0.07
Nodes (38): TestAbbreviationDocs(), applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias() (+30 more)

### Community 52 - "exportconfig_test.go"
Cohesion: 0.09
Nodes (33): TestExecute(), TestResolveScriptPath(), runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand(), TestPodmanEUIDGuardSkips(), writeRootlessPodmanEnv(), ExitCode() (+25 more)

### Community 53 - "Container Security Guidelines"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 54 - "Operations"
Cohesion: 0.08
Nodes (24): Bringing up a fresh cluster, Configuring a site, Data replication, Docker and Podman mechanics, Every destructive command confirms, Exit codes, Extra CLI users differ by platform, Operations (+16 more)

### Community 55 - "Test catalogue"
Cohesion: 0.10
Nodes (21): abbrev_test.go, convert_test.go, Coverage, examples_test.go, Fixtures and doubles, Injectable seams, internal/abbrev, internal/convert (+13 more)

### Community 57 - "ReplSite"
Cohesion: 0.15
Nodes (16): MateChannel, curlConfigFlag(), Ops, requirePrimaryActive(), requireVirtualRouterNamesMatch(), sortedSiteNames(), SwitchPreflight(), replSites() (+8 more)

### Community 58 - "validName"
Cohesion: 0.17
Nodes (10): cliRunNames(), runCLISkeleton(), TestExecCLI(), TestValidName(), validName(), ValidScriptName(), rejectionIn(), shellScriptPath() (+2 more)

### Community 59 - ".rpc"
Cohesion: 0.10
Nodes (21): backupTarget, Credential, sempMate, TestHTTPStatusHelpers(), schemaTransport(), TestTransportVocabularyIsThreeWords(), anyHTTP2xx(), curlConfigLine() (+13 more)

### Community 60 - "confirm.go"
Cohesion: 0.21
Nodes (17): exportconfigReadCounter, layer, TestStdinCanAnswerClosedFile(), confirmActionStrict(), confirmDelete(), confirmDowngrade(), confirmGate(), confirmLayer() (+9 more)

### Community 61 - ".configRows"
Cohesion: 0.13
Nodes (24): orNone(), orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminPassState(), adminSecretRow() (+16 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.17
Nodes (35): failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent(), TestCheckIDMappingRefusesAnUnmappedRunUser() (+27 more)

### Community 63 - "strings.Builder"
Cohesion: 0.10
Nodes (28): WeightedNodeTerm, mdRow(), writeAbbrevTable(), LoadBalancer, NodeAffinity, NodeMatchExpr, Placement, PodAffinityTerm (+20 more)

### Community 64 - "output_test.go"
Cohesion: 0.11
Nodes (24): TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), TestCheckReportSkipsEmptySections(), New(), NewFunc() (+16 more)

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

### Community 70 - "opPerformReplication"
Cohesion: 0.15
Nodes (25): mateChannelFunc, lineSink(), progress(), confirmReplicationConfig(), mateChannel(), mateSEMPPassword(), opConfigureReplication(), opCtrConfigureReplication() (+17 more)

### Community 71 - "watchCluster"
Cohesion: 0.29
Nodes (6): deployJSON(), TestOperatorReleaseNarrowsInsteadOfRemoving(), TestReconcileWatchDecidesWhatToApply(), TestReconcileWatchFlagsAWideningAsAQuestion(), TestSetWatchRefusesAnEmptyList(), watchCluster()

### Community 72 - "ParseRole"
Cohesion: 0.50
Nodes (5): TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), ParseRole(), RoleAbbrev()

### Community 73 - "internal/k8s"
Cohesion: 0.10
Nodes (21): adminsecret_test.go, check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, generatorpage_test.go, inspect_test.go, internal/k8s (+13 more)

### Community 74 - "importIgnore"
Cohesion: 0.67
Nodes (3): importIgnore(), TestImportIgnoreKeepsUnclassifiedSectionsInTheDiff(), TestImportOpsImportIgnore()

### Community 75 - "usagef"
Cohesion: 0.09
Nodes (24): childExitError, usageError, logArgs(), runConvert(), asUsage(), childStatus(), isUsage(), markUsageArgs() (+16 more)

### Community 76 - ".ConfigureReplication"
Cohesion: 0.16
Nodes (10): cliMate, Ops, isRunCLIRejection(), newlineIf(), replPhase1Rejected(), replPhase2Rejected(), replTransportFailure(), replVPNList() (+2 more)

### Community 77 - "Command"
Cohesion: 0.16
Nodes (10): cmdField, guardedCmd, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandUnmarshal(), Command, guardCommandOf(), setGuardCommand() (+2 more)

### Community 78 - "newTeardownApplyFixture"
Cohesion: 0.13
Nodes (16): driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), minimalCapture(), newTeardownApplyFixture(), targetVPNCapture(), TestImportOpsExportConfigScopeSelectsCLICommand() (+8 more)

### Community 82 - "k8s/matechannel_test.go"
Cohesion: 0.22
Nodes (25): CLIScriptPath(), execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf(), replCfg() (+17 more)

### Community 83 - "limits_test.go"
Cohesion: 0.14
Nodes (25): limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessCoreFloorFollowsTheConfiguredLimit(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager(), TestCheckLimitsRootlessRefusesUndelegatedControllers() (+17 more)

### Community 84 - "occCluster"
Cohesion: 0.18
Nodes (10): occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants(), TestNamespaceIsProtectedSharesDeleteNamespacesList() (+2 more)

### Community 85 - "Troubleshooting Guide"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 86 - "k8s/inspect_test.go"
Cohesion: 0.13
Nodes (22): age(), podHealth(), pvcLevel(), replicaLevel(), roleRank(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape() (+14 more)

### Community 87 - "MateConfig"
Cohesion: 0.11
Nodes (37): MateConfig, TestMateChannelShowReplication(), cliTransport(), containsEndpoint(), labelValue(), normTransport(), parseHostPort(), ParseShowReplication() (+29 more)

### Community 88 - "haCfg"
Cohesion: 0.10
Nodes (29): AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestReadAdminPasswordArgvAndDecoding(), TestReadAdminPasswordGuardsTheCommand(), TestReadAdminPasswordNamesTheNextStep() (+21 more)

### Community 89 - "render_test.go"
Cohesion: 0.13
Nodes (34): boolStr(), BrokerCR(), containerArtifacts(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets(), TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPrivilegePosture() (+26 more)

### Community 90 - "main_test.go"
Cohesion: 0.17
Nodes (18): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+10 more)

### Community 91 - "sections_test.go"
Cohesion: 0.09
Nodes (37): Disposition, SectionRule, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified(), renderImportDocs() (+29 more)

### Community 92 - "step"
Cohesion: 0.31
Nodes (4): App, step(), TestResolveMissingBinaryIsActionable(), Resolve()

### Community 93 - "Container"
Cohesion: 0.09
Nodes (21): Container, DockerConfig, Network, PodmanConfig, scalingTier, Ulimits, TestDefaultK8sPortsMatchesOperator(), applyContainerBlockDefaults() (+13 more)

### Community 94 - "Config"
Cohesion: 0.20
Nodes (6): expandTilde(), expandTildeToken(), Config, isPathSep(), TestExpandTilde(), tildeHome()

### Community 95 - "Secret Reference Resolution"
Cohesion: 0.50
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 96 - "github.com/spf13/cobra.Command"
Cohesion: 0.08
Nodes (36): shorthand, renderAbbrevDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeResolutionNotes(), writeShorthandTable(), writeUnabbreviated(), anchor() (+28 more)

### Community 97 - "replConfig"
Cohesion: 0.23
Nodes (10): Replication, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationEmptyViaAccepted(), TestValidateReplicationRejects() (+2 more)

### Community 98 - "Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "blocks.go"
Cohesion: 0.12
Nodes (25): Omission, Region, token, isMarker(), TestIsMarkerDiscriminatesNamespace(), brokerTypeFromSchema(), firstQuoted(), Capture (+17 more)

### Community 100 - ".checkUserManagerLimits"
Cohesion: 0.23
Nodes (8): delegateSetting(), Manager, limitText(), missingControllers(), parseLimit(), parseUnitProps(), TestMissingControllers(), TestParseLimit()

### Community 102 - "Load"
Cohesion: 0.07
Nodes (52): Example, TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), TestExamplesWritesOutFile(), minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError() (+44 more)

### Community 103 - "ctrCfg"
Cohesion: 0.07
Nodes (52): inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer() (+44 more)

### Community 104 - "Platform"
Cohesion: 0.13
Nodes (27): opFunc, checkFlagPlatforms(), commandPlatforms(), declaredList(), parsePlatformList(), platformSuffix(), prepare(), promptPlatform() (+19 more)

### Community 105 - "internal/cli"
Cohesion: 0.15
Nodes (13): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, euid_test.go, examples_test.go (+5 more)

### Community 107 - "Compose"
Cohesion: 0.13
Nodes (16): Compose(), composeEscape(), ComposeProject(), ContainerSecrets(), ContainerSecret, secretFilePath(), SecretPreflight(), TestBridgePublishesOnlyTheListedPorts() (+8 more)

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
Nodes (22): TestValidateSparseConfigExplainsItself(), GenOperator(), joinYAMLDocs(), nonEmpty(), operatorProbes(), RenderOperator(), renderOperatorNS(), splitAfterNamespace() (+14 more)

### Community 112 - "capRunner"
Cohesion: 0.09
Nodes (27): capCall, capRunner, NewManager(), callIndex(), TestManagerLogsCLIShell(), TestManagerNilSinks(), TestManagerPrepHostRootlessUsesUnshareChown(), TestPreflightRunsBeforeAnything() (+19 more)

### Community 113 - "Config Package Tests"
Cohesion: 0.12
Nodes (17): adminsecret_test.go, artifactvalues_test.go, command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go (+9 more)

### Community 115 - "operatorversion.go"
Cohesion: 0.15
Nodes (14): operatorImage(), TestOperatorImage(), compareVersions(), Cluster, imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion() (+6 more)

### Community 116 - "EnvPairs"
Cohesion: 0.15
Nodes (13): NodeIdentity, containerSecretSpecs(), EnvPairs(), groupKey(), itoa(), ServerCertBundlePath(), envLines(), TestAdditionalUsersReachBothHalves() (+5 more)

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.16
Nodes (14): containerState, healthState, decodeInspect(), Manager, orUnknown(), printable(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown() (+6 more)

### Community 121 - "scripts.go"
Cohesion: 0.07
Nodes (35): showCmd, TestDomainCerts(), TestServerCert(), validCLILine(), Ops, assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript() (+27 more)

### Community 122 - "coverage_test.go"
Cohesion: 0.05
Nodes (52): runErrMatchTransport, matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError() (+44 more)

### Community 123 - ".validateContainerArtifactValues"
Cohesion: 0.20
Nodes (7): Config, portOrRange(), TestValidContainerPort(), validContainerPort(), validCoreLimit(), validHealthDuration(), validDNSLabel()

### Community 125 - "newOperatorCmd"
Cohesion: 0.33
Nodes (13): newOperatorCmd(), newOperatorDeployCmd(), newOperatorGenerateCmd(), newOperatorLogsCmd(), newOperatorRemoveCmd(), newOperatorRestartCmd(), newOperatorStartCmd(), newOperatorStatusCmd() (+5 more)

### Community 127 - ".validateContainer"
Cohesion: 0.16
Nodes (11): TestValidateContainerArtifactValues(), TestValidateProbeCommandAccepts(), TestValidateProbeCommandRejects(), cpuSetCount(), cpuSetRange(), TestCPUSetCount(), TestCPUSetRange(), invertedCPUSetRange() (+3 more)

### Community 128 - "Manager"
Cohesion: 0.21
Nodes (4): Manager, idMapCovers(), origin(), runUserIDs()

### Community 129 - "Scaling YAML Decoding"
Cohesion: 0.31
Nodes (5): scalingKey, scalingSpelling, Scaling, scalingKeyIndex(), scalingKeyList()

### Community 130 - "podName"
Cohesion: 0.10
Nodes (15): BrokerPodSuffixShape(), Cluster, TestOperatorAdminSecretNameMatchesTheBrokerNames(), Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes(), HARoles(), lbServiceName() (+7 more)

### Community 131 - "k8s/preflight.go"
Cohesion: 0.44
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "ReplVia"
Cohesion: 0.18
Nodes (9): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, viaDescription(), sortedKeys(), validateReplEndpoints(), validateReplVia() (+1 more)

### Community 134 - "lastLines"
Cohesion: 0.67
Nodes (3): TestLastLines(), TestLastLinesEqualCount(), lastLines()

### Community 135 - "productKeysScript"
Cohesion: 0.40
Nodes (6): productKeyScript(), productKeysScript(), removeProductKeysScript(), TestProductKeyScriptsShareAPreamble(), TestProductKeysScript(), TestRemoveProductKeysScript()

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

- **Why does `Sink` connect `Sink` to `output_test.go`, `Ops`, `opPerformReplication`, `Manager`, `config_test.go`, `BrokerType`, `RenderDiffResult`, `k8s/inspect_test.go`, `ImportPlan`, `.configRows`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **Why does `Platform` connect `Platform` to `testing.T`, `Manager`, `CheckCommand`, `captureStdout`, `convert_test.go`, `manager_test.go`, `commands.go`, `Config`, `Role`, `config_test.go`, `Ops`, `Command`, `limits_test.go`, `render_test.go`, `step`, `Container`, `Config`, `platformTitle`, `Load`, `ctrCfg`, `Compose`, `capRunner`, `.validateContainerArtifactValues`, `.validateContainer`?**
  _High betweenness centrality (0.015) - this node is a cross-community bridge._
- **Why does `ImportPlan` connect `ImportPlan` to `importIgnore`, `usagef`, `newTeardownApplyFixture`, `transform_test.go`, `config_test.go`, `BrokerType`?**
  _High betweenness centrality (0.014) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 49 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 49 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _238 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.06654567453115548 - nodes in this community are weakly interconnected._
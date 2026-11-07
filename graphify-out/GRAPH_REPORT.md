# Graph Report - solace-cnt-scripts  (2026-11-07)

## Corpus Check
- 202 files · ~707,906 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .golden 17, (none) 3, .cli 2)

## Summary
- 4031 nodes · 20737 edges · 118 communities (110 shown, 8 thin omitted)
- Extraction: 72% EXTRACTED · 28% INFERRED · 0% AMBIGUOUS · INFERRED: 5724 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `7310d556`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- App
- newTestOps
- .kubectl
- Role
- cli_test.go
- config_test.go
- Command Reference Docs
- Manager
- K8sConfig
- completion_test.go
- Command
- output_test.go
- NewCluster
- main_test.go
- convert_test.go
- transform_test.go
- manager_test.go
- commands.go
- coverage_test.go
- Bash Dev Script
- localCfg
- PowerShell Dev Script
- vars
- switchplan_test.go
- newLocalOps
- diff_test.go
- renderDriver
- hostpath_test.go
- importops_test.go
- watch_test.go
- Solace Module Root
- statusreport_test.go
- CLAUDE.md
- k8s/preflight_test.go
- ResolveDomainCerts
- watchLoop
- github.com/spf13/cobra.ShellCompDirective
- .operatorRows
- annotate_test.go
- ReplSite
- Per-package doubles
- inject_test.go
- ctrCfg
- exportconfig_test.go
- io.Writer
- .BackupRevertActivity
- checkRow
- config_test.go
- scripts.go
- eqArgs
- config_ops_test.go
- haCfg
- .ConfigureReplication
- .dump
- MateConfig
- replication_test.go
- .OperatorApply
- replCfg
- MateChannel
- replConfig
- rootless_test.go
- Test catalogue
- Injectable seams
- secrets_test.go
- blocks.go
- step
- RenderMate
- k8s/inspect.go
- checkRow
- NewCluster
- Operations
- tlssecret_test.go
- .RunCLI
- Coverage
- EnvPairs
- cli/watch_test.go
- newRootCmd
- BrokerType
- containerTransport
- replicationops_test.go
- Platform
- limits_test.go
- Developer guide
- VPNRepl
- namespace_test.go
- replication_test.go
- Config
- render_test.go
- New
- RuleFor
- Configuration
- Sink
- gen-cert.sh
- storage_test.go
- Container
- .validate
- RenderDiffResult
- Image
- Platforms
- allowcommand_test.go
- .checkUserManagerLimits
- preflight_test.go
- kubectlTransport
- Exporting and importing configuration
- adminsecret_test.go
- NewExec
- cliTransport
- capRunner
- ParseRole
- solaceRows
- runGuarded
- .inspectContainer
- operatorversion_test.go
- strings.Builder

## God Nodes (most connected - your core abstractions)
1. `App` - 218 edges
2. `newTestOps()` - 167 edges
3. `Config` - 151 edges
4. `cli_test.go` - 151 edges
5. `ctrCfg()` - 145 edges
6. `NewCluster()` - 144 edges
7. `Per-package doubles` - 134 edges
8. `Role` - 132 edges
9. `newCapMgr()` - 126 edges
10. `loadK8s()` - 120 edges

## Surprising Connections (you probably didn't know these)
- `The preflight` --references--> `info()`  [INFERRED]
  docs/operations.md → internal/k8s/checkreport.go
- `annotate_test.go` --references--> `Annotate()`  [INFERRED]
  docs/test.md → internal/broker/annotate.go
- `annotate_test.go` --references--> `TestAnnotateRoundTripPreservesBlocks()`  [INFERRED]
  docs/test.md → internal/broker/annotate_test.go
- `annotate_test.go` --references--> `TestAnnotateRoundTripCaptureIsMarked()`  [INFERRED]
  docs/test.md → internal/broker/annotate_test.go
- `annotate_test.go` --references--> `TestAnnotateIsDeterministic()`  [INFERRED]
  docs/test.md → internal/broker/annotate_test.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (118 total, 8 thin omitted)

### Community 0 - "App"
Cohesion: 0.07
Nodes (100): App, euid_test.go, logArgs(), logSince(), confirmAction(), confirmDelete(), lineSink(), progress() (+92 more)

### Community 1 - "newTestOps"
Cohesion: 0.05
Nodes (67): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, broker_test.go (+59 more)

### Community 2 - ".kubectl"
Cohesion: 0.05
Nodes (27): runtime_test.go, BrokerPodSuffixShape(), Cluster, normalizeToList(), discoveryServiceName(), HARoles(), lbServiceName(), podName() (+19 more)

### Community 3 - "Role"
Cohesion: 0.09
Nodes (20): mateChannelFunc, Ops, New(), Transport, runExport(), exportScope(), opCtrExportConfig(), k8sAdminPassword() (+12 more)

### Community 4 - "cli_test.go"
Cohesion: 0.06
Nodes (113): leaderRun, cli_test.go, examples_test.go, Shared env fixtures, allowRuntime(), capture(), captureStderr(), collectPaths() (+105 more)

### Community 5 - "config_test.go"
Cohesion: 0.05
Nodes (135): config_test.go, redundancy_test.go, scaling_test.go, TestContainersStillRequireTheAdminPassword(), TestOperatorAdminSecretName(), TestPSKConfiguredCountsBothSpellings(), TestValidateK8sArtifactValues(), assertContainerBlockDefaults() (+127 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (50): Commands, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker cli, solace-util broker configure (+42 more)

### Community 7 - "Manager"
Cohesion: 0.09
Nodes (8): Manager, orNone(), platformTitle(), secretSummary(), setOrMissing(), runUserIDs(), Output(), ContainerSecret

### Community 8 - "K8sConfig"
Cohesion: 0.18
Nodes (7): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, writeSecurity()

### Community 9 - "completion_test.go"
Cohesion: 0.19
Nodes (24): completion_test.go, driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion() (+16 more)

### Community 10 - "Command"
Cohesion: 0.04
Nodes (55): cmdField, guardedCmd, ReplVia, ReplViaKube, ReplViaSEMP, command_test.go, duration_test.go, execguard_test.go (+47 more)

### Community 11 - "output_test.go"
Cohesion: 0.16
Nodes (19): internal/output, output_test.go, columnWidths(), Level, pad(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag() (+11 more)

### Community 12 - "NewCluster"
Cohesion: 0.11
Nodes (48): operator_test.go, GenOperator(), operatorImage(), RenderOperator(), boolPtr(), bundledDefaultPorts(), commentedExamplePorts(), covSwapBundle() (+40 more)

### Community 13 - "main_test.go"
Cohesion: 0.19
Nodes (20): internal/tools/vulnjudge, main_test.go, describe(), judge(), main(), plural(), run(), load() (+12 more)

### Community 14 - "convert_test.go"
Cohesion: 0.11
Nodes (57): Result, segment, convert_test.go, Convert(), checkGolden(), convertOK(), hasWarning(), strictDecode() (+49 more)

### Community 15 - "transform_test.go"
Cohesion: 0.10
Nodes (62): Block, TargetState, transform_test.go, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects() (+54 more)

### Community 16 - "manager_test.go"
Cohesion: 0.07
Nodes (92): inspect_test.go, manager_test.go, GuardPodmanEUID(), inspectReply(), kvValue(), TestInspectContainerArgv(), TestInspectContainerRejectsAnUnusableReply(), TestInspectContainerSurfacesTheEngineError() (+84 more)

### Community 17 - "commands.go"
Cohesion: 0.12
Nodes (75): opFunc, roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd() (+67 more)

### Community 18 - "coverage_test.go"
Cohesion: 0.06
Nodes (85): runErrMatchTransport, runErrTransport, coverage_test.go, internal/broker, verify_ops_test.go, matchCLI(), newTestOps(), ranContains() (+77 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "localCfg"
Cohesion: 0.09
Nodes (48): semp_test.go, sempmate_test.go, bridgeHostPort(), sempPort(), sempV1OK(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA() (+40 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "vars"
Cohesion: 0.12
Nodes (14): doc, vars, Node roles, Reading this reference, The keys, boolOf(), commentSafe(), countMarkers() (+6 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.14
Nodes (36): PhaseKind, SiteState, SwitchAction, SwitchPhase, switchplan_test.go, BuildSwitchPlan(), confirmDemotions(), ExecuteSwitchPlan() (+28 more)

### Community 24 - "newLocalOps"
Cohesion: 0.15
Nodes (40): verify_local_test.go, cliScriptPath(), isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn(), TestDefaultLocalAddrs() (+32 more)

### Community 25 - "diff_test.go"
Cohesion: 0.15
Nodes (26): blockKey, mergedBlock, mergedLine, diff_test.go, ancestorReported(), DiffBlocks(), indexByIdentity(), keyOf() (+18 more)

### Community 26 - "renderDriver"
Cohesion: 0.15
Nodes (22): chunk, chunkResult, driver_test.go, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce() (+14 more)

### Community 27 - "hostpath_test.go"
Cohesion: 0.08
Nodes (32): hostpath_test.go, guardConfig(), BaseName(), CheckHostPath(), expandTilde(), expandTildeToken(), HasPathSeparator(), IsAbsHostPath() (+24 more)

### Community 28 - "importops_test.go"
Cohesion: 0.06
Nodes (57): ImportResult, PlannedSection, importops_test.go, isPragma(), defaultVPNFirst(), describeScope(), ImportPlan, importIgnore() (+49 more)

### Community 29 - "watch_test.go"
Cohesion: 0.13
Nodes (27): watch_test.go, subtractWatch(), deniedCluster(), deployJSON(), namedOperatorJSON(), operatorItem(), TestAmbiguousOperatorAbortsTheDeploy(), TestDesiredWatchAppendsTheBrokerNamespace() (+19 more)

### Community 31 - "statusreport_test.go"
Cohesion: 0.15
Nodes (29): statusreport_test.go, ownsPod(), reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportOmitsWorkloadsThatAreNotTheBrokers(), TestBrokerReportOrdersPodsOfTheSameRoleByName(), TestBrokerReportPreviewStopsAtAnUnapprovedCommand(), TestBrokerReportReadsEachKindOnce() (+21 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.05
Nodes (38): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, Command reference, Global flags, Index (+30 more)

### Community 33 - "k8s/preflight_test.go"
Cohesion: 0.23
Nodes (14): preflight_test.go, TestCanIAnswerReadsTheLastLine(), TestCanIAnswersWithoutFailing(), TestCanIArgsScopesByResourceKind(), TestCanIIsPermittedUnderEcho(), TestCanIRefusedReadsEveryRefusalSpelling(), TestCanIUnderEchoStillHonoursTheExecutionGuard(), TestPreflightAcceptsAWarnedYes() (+6 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.08
Nodes (42): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domaincerts_resolve_test.go, domaincerts_test.go, caNameSafe() (+34 more)

### Community 35 - "watchLoop"
Cohesion: 0.31
Nodes (4): watchInterval(), watchLabel(), watchLoop(), withWatch()

### Community 36 - "github.com/spf13/cobra.ShellCompDirective"
Cohesion: 0.33
Nodes (6): completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), isEnvFileName(), matching()

### Community 37 - ".operatorRows"
Cohesion: 0.33
Nodes (6): compareVersions(), imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), TestFindOperatorDeploymentIsScopedByNamespace()

### Community 38 - "annotate_test.go"
Cohesion: 0.06
Nodes (54): Capture, annotate_test.go, applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV() (+46 more)

### Community 40 - "ReplSite"
Cohesion: 0.18
Nodes (20): PlannedRoles(), RoleAtSite(), TestPlannedRolesNamesEveryListedVPN(), TestRoleAtSiteIsTheComplement(), curlConfigFlag(), requirePrimaryActive(), requireVirtualRouterNamesMatch(), sortedSiteNames() (+12 more)

### Community 41 - "Per-package doubles"
Cohesion: 0.17
Nodes (39): Per-package doubles, replShowMate(), replVPNTable(), answer(), assertMode(), callIndex(), composeUp(), composeUpArgs() (+31 more)

### Community 42 - "inject_test.go"
Cohesion: 0.17
Nodes (28): lineRole, serviceLine, shutdownStyle, svcKey, inject_test.go, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "ctrCfg"
Cohesion: 0.15
Nodes (31): secrets_test.go, TestManagerDeployPodmanReplaceOnlyOnPodman47(), TestManagerDeployPodmanSecretError(), versionAnswer(), ResolveSecretValues(), certCreate(), certFixture(), TestCertSecretLabelDrivesTheRestart() (+23 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.10
Nodes (46): exportconfigReadCounter, opRunner, exit_test.go, exportconfig_test.go, internal/cli, runRootWith(), TestDeployOperatorNoPromptApprovesOnlyTheDowngrade(), TestK8sConfirmDeclined() (+38 more)

### Community 46 - "io.Writer"
Cohesion: 0.11
Nodes (18): interactiveFailRunner, runFailRunner, Repository layout, internal/engine, runner_test.go, noEnv, legacyMethod(), legacyMethod() (+10 more)

### Community 47 - ".BackupRevertActivity"
Cohesion: 0.16
Nodes (13): backupTarget, Credential, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigLine(), httpBody(), lastHTTPStatus(), revertActivityBackupBody() (+5 more)

### Community 48 - "checkRow"
Cohesion: 0.29
Nodes (12): orNone(), orValue(), setOrMissing(), setOrNone(), additionalUsersRow(), adminSecretRow(), info(), pskRow() (+4 more)

### Community 49 - "config_test.go"
Cohesion: 0.15
Nodes (31): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, span, AdditionalUser, Node, Redundancy (+23 more)

### Community 50 - "scripts.go"
Cohesion: 0.07
Nodes (48): showCmd, scripts_test.go, TestDiagnostics(), TestDisableDefaultVPN(), TestProductKeys(), TestRemoveProductKeys(), TestServerCert(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList() (+40 more)

### Community 51 - "eqArgs"
Cohesion: 0.10
Nodes (46): check_test.go, cluster_test.go, internal/k8s, names_test.go, ops_test.go, transport_test.go, TestReachable(), TestResolveStorageClass() (+38 more)

### Community 52 - "config_ops_test.go"
Cohesion: 0.19
Nodes (13): config_ops_test.go, TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(), TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator(), TestExecShellScriptRefusesAnUnsafeNameBeforeUploading(), TestExecShellScriptTrustsTheExitStatusNotTheTranscript() (+5 more)

### Community 53 - "haCfg"
Cohesion: 0.12
Nodes (44): prep_test.go, haCfg(), GenBroker(), joinManifests(), namespaceManifest(), adminCfg(), secretNamesIn(), TestAdminSecretStatesEmitTheirArtifacts() (+36 more)

### Community 54 - ".ConfigureReplication"
Cohesion: 0.25
Nodes (8): ValidVPNName(), validVPNName(), isRunCLIRejection(), newlineIf(), replPhase1Rejected(), replPhase2Rejected(), replTransportFailure(), replVPNList()

### Community 55 - ".dump"
Cohesion: 0.10
Nodes (64): opCall, ops_k8s_test.go, failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch() (+56 more)

### Community 56 - "MateConfig"
Cohesion: 0.16
Nodes (18): TestMateChannelShowReplication(), containsEndpoint(), MateConfig, labelValue(), normTransport(), parseHostPort(), ParseShowReplication(), parseShowReplicationAppliance() (+10 more)

### Community 57 - "replication_test.go"
Cohesion: 0.08
Nodes (54): replRig, Commands, solace-util broker, Configuring a site, Data replication, Permissions, Reaching the mate, Switching roles (+46 more)

### Community 58 - ".OperatorApply"
Cohesion: 0.17
Nodes (11): joinYAMLDocs(), nonEmpty(), operatorEscalationProbes(), operatorProbes(), splitAfterNamespace(), splitOperatorBundle(), canIAnswer(), canIRefused() (+3 more)

### Community 59 - "replCfg"
Cohesion: 0.19
Nodes (28): matechannel_test.go, CLIArg(), CLIScriptPath(), execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf() (+20 more)

### Community 60 - "MateChannel"
Cohesion: 0.18
Nodes (25): CLIRunner, fakeRun, matechannel_test.go, MateChannel, NewCLIMate(), mateReplies(), newTestMate(), TestLocalMateRunsThroughOps() (+17 more)

### Community 61 - "replConfig"
Cohesion: 0.13
Nodes (22): Replication, secretRef, imagepullsecret_test.go, replication_test.go, pullSecretCfg(), TestImagePullSecretDerivedSuffixFitsWhereAdditionalUsersAlreadyFits(), TestImagePullSecretNameBringYourOwn(), TestImagePullSecretNameConfiguredNameWins() (+14 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.19
Nodes (41): rootless_test.go, failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+33 more)

### Community 63 - "Test catalogue"
Cohesion: 0.33
Nodes (6): Fixtures and doubles, internal/convert, internal/render, Running the tests, Summary, Test catalogue

### Community 64 - "Injectable seams"
Cohesion: 0.15
Nodes (12): layer, Injectable seams, defaultLocalAddrs(), confirmActionStrict(), confirmDowngrade(), confirmLayer(), confirmRedeploy(), confirmRestart() (+4 more)

### Community 65 - "secrets_test.go"
Cohesion: 0.12
Nodes (33): secrets_test.go, GenSecrets(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), readTLSMaterial(), checkGolden() (+25 more)

### Community 66 - "blocks.go"
Cohesion: 0.09
Nodes (37): Omission, Region, token, blocks_test.go, omit_test.go, Annotate(), TestAnnotateIsDeterministic(), TestAnnotateRoundTripCaptureIsMarked() (+29 more)

### Community 67 - "step"
Cohesion: 0.11
Nodes (23): step(), newConvertCmd(), runConvert(), newExamplesCmd(), runExample(), addOutFlags(), checkAllowCommand(), emitOrWrite() (+15 more)

### Community 68 - "RenderMate"
Cohesion: 0.18
Nodes (18): RenderMate(), RenderMateRemovals(), mateAppliance(), mateSoftware(), TestMateRoundTripsThroughItsOwnGrammar(), TestRenderMateAppliance(), TestRenderMateApplianceRefusesTwoHostsWhoseTransportsDiffer(), TestRenderMateApplianceRefusesUnrenderable() (+10 more)

### Community 69 - "k8s/inspect.go"
Cohesion: 0.11
Nodes (35): inspect_test.go, age(), anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), podHealth(), pvcLevel() (+27 more)

### Community 70 - "checkRow"
Cohesion: 0.24
Nodes (11): storageClassSuitable(), containsString(), failRow(), installSteps(), okRow(), portRows(), skipRow(), TestCheckReportFailedCounts() (+3 more)

### Community 71 - "NewCluster"
Cohesion: 0.09
Nodes (62): checkreport_test.go, deploy_test.go, TestCheckStopsProbingWhenUnreachable(), TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestOperatorInstallVerdict(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestOperatorValidateAsksNothingAboutTheBroker(), TestStorageRows() (+54 more)

### Community 72 - "Operations"
Cohesion: 0.09
Nodes (22): A replication switch refuses before changing anything, A wrapper runtime is refused, Bringing up a fresh cluster, Docker compose secrets need compose 2.23.1+, Every destructive command confirms, Exit codes, Extra CLI users differ by platform, Forbidden: cannot list deployments at the cluster scope (+14 more)

### Community 73 - "tlssecret_test.go"
Cohesion: 0.43
Nodes (8): tlssecret_test.go, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsDeriveASecretName(), TestSuppliedCertsMakeTheSecretOurs(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), TestTLSServerSecretNameFollowsTheStates(), tlsCfg()

### Community 74 - ".RunCLI"
Cohesion: 0.12
Nodes (11): field(), runCLISkeleton(), validCLILine(), validName(), ValidScriptName(), rejectionIn(), activity(), DiagnosticsOptions (+3 more)

### Community 75 - "Coverage"
Cohesion: 0.21
Nodes (16): Coverage, ContainerNoFile(), HealthCheck, TestLimitsCheckAssertsWhatTheArtifactAsks(), healthCmd(), Quadlet(), assertNoCheckoutPath(), containerArtifacts() (+8 more)

### Community 76 - "EnvPairs"
Cohesion: 0.11
Nodes (24): servercert_test.go, NodeIdentity, Compose(), ContainerSecrets(), containerSecretSpecs(), EnvPairs(), groupKey(), itoa() (+16 more)

### Community 77 - "cli/watch_test.go"
Cohesion: 0.38
Nodes (10): watch_test.go, stopAfter(), TestOperatorStatusWatches(), TestStatusWithoutWatchRunsOnce(), TestWatchAppendsWhenNotATerminal(), TestWatchEndsQuietlyWhenInterrupted(), TestWatchIntervalIsValidated(), TestWatchKeepsGoingAfterAFailedRefresh() (+2 more)

### Community 78 - "newRootCmd"
Cohesion: 0.10
Nodes (38): abbrevdoc_test.go, aliases_test.go, platform_test.go, TestFlagShorthandsAreConsistent(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb() (+30 more)

### Community 79 - "BrokerType"
Cohesion: 0.22
Nodes (8): BrokerType, cliMate, transcript_test.go, bannerType(), showReplicationScript(), showVPNReplicationScript(), TestTranscriptBannerType(), TestTranscriptTeardownFixtureResetsNotDeletes()

### Community 81 - "replicationops_test.go"
Cohesion: 0.17
Nodes (34): replicationops_test.go, hasCall(), mateConvergenceShutdowns(), replicationVPNLines(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching() (+26 more)

### Community 82 - "Platform"
Cohesion: 0.04
Nodes (48): keyValueEntries, scalingTier, artifactvalues_test.go, checkTimezone(), portOrRange(), TestCheckTimezone(), TestValidateContainerArtifactValues(), TestValidateImageReference() (+40 more)

### Community 83 - "limits_test.go"
Cohesion: 0.17
Nodes (32): limits_test.go, limitsMgr(), nrOpenProbe(), procLimits(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessCoreFloorFollowsTheConfiguredLimit(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips() (+24 more)

### Community 84 - "Developer guide"
Cohesion: 0.11
Nodes (20): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Tests, The env-file generator page (+12 more)

### Community 85 - "VPNRepl"
Cohesion: 0.07
Nodes (16): AdminState, covUploadFailsTo, QueueState, scriptedMate, sempMate, uploadErrTransport, replMate, ReplRole (+8 more)

### Community 86 - "namespace_test.go"
Cohesion: 0.31
Nodes (11): namespace_test.go, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 87 - "replication_test.go"
Cohesion: 0.15
Nodes (24): colSpan, replication_test.go, dashSpans(), flagByte(), gutterClear(), ParseVPNReplication(), sliceSpan(), loadShowReplication() (+16 more)

### Community 88 - "Config"
Cohesion: 0.10
Nodes (22): TLS, adminsecret_test.go, Config, clearTierDerived(), setContainerCPUSet(), setContainerMem(), setMonitorCPUSet(), AdminSecretInUse() (+14 more)

### Community 89 - "render_test.go"
Cohesion: 0.15
Nodes (39): render_test.go, boolStr(), BrokerCR(), ComposeProject(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets(), TestArtifactsCarryNoWideningTokens() (+31 more)

### Community 90 - "New"
Cohesion: 0.16
Nodes (11): orNone(), quoteList(), sectionLevel(), TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportPlanReportShowsOmissionsAndEachSectionsVerdict(), TestImportResultReportQuotesTheFailedChunkAndItsTail(), TestImportResultReportsTeardownAndRebuildSeparately() (+3 more)

### Community 91 - "RuleFor"
Cohesion: 0.13
Nodes (36): Disposition, SectionRule, appliance_test.go, importdoc_test.go, sections_test.go, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent() (+28 more)

### Community 93 - "Configuration"
Cohesion: 0.15
Nodes (13): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+5 more)

### Community 95 - "Sink"
Cohesion: 0.38
Nodes (4): Sink, NewFunc(), TestLeveledLinesCarryTheTagAndNoArrow(), TestNilSinksDiscard()

### Community 96 - "gen-cert.sh"
Cohesion: 0.27
Nodes (13): ask(), askPass(), askSan(), askYesNo(), buildSanList(), checkExit(), concatCnIfNotBlank(), echoUsage() (+5 more)

### Community 97 - "storage_test.go"
Cohesion: 0.36
Nodes (9): storage_test.go, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 98 - "Container"
Cohesion: 0.27
Nodes (6): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestContainerCoreLimitSpellings()

### Community 99 - ".validate"
Cohesion: 0.16
Nodes (9): solace-util, solace-util operator, Docker and Podman mechanics, Re-deploying is safe and explicit, Rootless podman prerequisites, What the broker cannot gain, Which user the broker runs as, TestCheckReportSkipsEmptySections() (+1 more)

### Community 100 - "RenderDiffResult"
Cohesion: 0.23
Nodes (11): BlockDiff, DiffResult, blockLabel(), diffRow(), orDash(), RenderDiffResult(), reportBlockLines(), TestDiffBlockLabelFormatsIdentityForTheReport() (+3 more)

### Community 102 - "Platforms"
Cohesion: 0.10
Nodes (37): examples_test.go, internal/examples, platform_test.go, Example, TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), TestExecBinariesCoversEveryPlatform(), TestGuardConfigIsValid() (+29 more)

### Community 103 - "allowcommand_test.go"
Cohesion: 0.38
Nodes (11): allowcommand_test.go, TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb() (+3 more)

### Community 104 - ".checkUserManagerLimits"
Cohesion: 0.19
Nodes (4): delegateSetting(), missingControllers(), pamCoreValue(), parseUnitProps()

### Community 105 - "preflight_test.go"
Cohesion: 0.30
Nodes (9): preflight_test.go, TestManagerLifecycleDocker(), TestManagerLifecyclePodmanSystemctl(), TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestComposeSecretEnvNamesCannotBeSystemVars(), TestPreflightFailureStopsLifecycle(), TestPreflightFailureStopsTheDeploy(), TestPreflightHintIsPlatformShaped() (+1 more)

### Community 107 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 108 - "adminsecret_test.go"
Cohesion: 0.48
Nodes (7): adminsecret_test.go, adminCfg(), TestAdminSecretNameFollowsTheStates(), TestAdminSecretRefusesTheOperatorsOwnNames(), TestKubernetesValidatesWithoutAnAdminPassword(), TestMonitorPasswordAndKeyNeedTheAdminPassword(), TestSecretNamesMustNotCollide()

### Community 109 - "NewExec"
Cohesion: 0.38
Nodes (6): resolve_test.go, TestChildEnvNamesAreNotSystemVariables(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), NewExec()

### Community 110 - "cliTransport"
Cohesion: 0.29
Nodes (7): cliTransport(), renderMateAppliance(), renderMateSoftware(), sortedKeys(), sortedSet(), TestTransportVocabularyIsThreeWords(), sortedVPNs()

### Community 112 - "capRunner"
Cohesion: 0.11
Nodes (27): capCall, capRunner, internal/container, runtime_test.go, transport_test.go, TestManagerLogsCLIShell(), TestManagerLogsRefusesPrevious(), TestManagerSessionNeedsAContainerName() (+19 more)

### Community 114 - "ParseRole"
Cohesion: 0.22
Nodes (9): Abbreviations, Flag shorthands, How a short form is resolved, Platforms, applyAliases(), argumentLine(), availableSubs(), writeCommand() (+1 more)

### Community 115 - "solaceRows"
Cohesion: 0.67
Nodes (3): printableRow(), solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv()

### Community 116 - "runGuarded"
Cohesion: 0.67
Nodes (5): runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand(), TestPodmanEUIDGuardSkips(), writeRootlessPodmanEnv()

### Community 118 - ".inspectContainer"
Cohesion: 0.22
Nodes (8): health(), orUnknown(), decodeOne(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings(), orValue(), KV

### Community 126 - "operatorversion_test.go"
Cohesion: 0.16
Nodes (22): operatorversion_test.go, TestOperatorApply(), TestOperatorApplyNamedPullSecretWithoutCredentials(), TestOperatorApplyNoCredentialsSkipsRegcred(), operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsNotSkippedByAFailedRead() (+14 more)

### Community 155 - "strings.Builder"
Cohesion: 0.07
Nodes (41): shorthand, WeightedNodeTerm, mdRow(), renderAbbrevDocs(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable() (+33 more)

## Knowledge Gaps
- **13 isolated node(s):** `solace`, `NO_COLOR`, `internal/broker`, `internal/cli`, `internal/config` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 188 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `App` to `Injectable seams`, `watchLoop`, `step`, `Role`, `cli_test.go`, `Manager`, `github.com/spf13/cobra.ShellCompDirective`, `Per-package doubles`, `exportconfig_test.go`, `io.Writer`, `newRootCmd`, `commands.go`, `Platform`, `config_test.go`, `VPNRepl`, `.dump`, `Config`, `replication_test.go`?**
  _High betweenness centrality (0.042) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `newTestOps()` (e.g. with `importops_test.go` and `Injectable seams`) actually correct?**
  _`newTestOps()` has 117 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `NO_COLOR`, `internal/broker` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.06585858585858585 - nodes in this community are weakly interconnected._
- **Why does `Config` connect `Config` to `App`, `newTestOps`, `.kubectl`, `Role`, `cli_test.go`, `config_test.go`, `Manager`, `K8sConfig`, `Command`, `NewCluster`, `convert_test.go`, `manager_test.go`, `coverage_test.go`, `localCfg`, `hostpath_test.go`, `strings.Builder`, `k8s/preflight_test.go`, `ResolveDomainCerts`, `ReplSite`, `Per-package doubles`, `ctrCfg`, `checkRow`, `config_test.go`, `eqArgs`, `haCfg`, `.dump`, `.OperatorApply`, `replCfg`, `replConfig`, `secrets_test.go`, `NewCluster`, `tlssecret_test.go`, `Coverage`, `EnvPairs`, `containerTransport`, `Platform`, `namespace_test.go`, `render_test.go`, `storage_test.go`, `Container`, `Image`, `Platforms`, `kubectlTransport`, `adminsecret_test.go`, `capRunner`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Are the 11 inferred relationships involving `Config` (e.g. with `config_test.go` and `execguard_test.go`) actually correct?**
  _`Config` has 11 INFERRED edges - model-reasoned connections that need verification._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.05126452494873548 - nodes in this community are weakly interconnected._
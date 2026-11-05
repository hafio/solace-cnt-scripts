# Graph Report - solace-cnt-scripts  (2026-11-05)

## Corpus Check
- 204 files · ~697,488 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .golden 17, (none) 3, .cli 2)

## Summary
- 3997 nodes · 20474 edges · 111 communities (103 shown, 8 thin omitted)
- Extraction: 73% EXTRACTED · 27% INFERRED · 0% AMBIGUOUS · INFERRED: 5629 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `c350035e`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- App
- newTestOps
- .ns
- kubectlTransport
- cli_test.go
- config_test.go
- Command Reference Docs
- Manager
- K8sConfig
- completion_test.go
- Command
- deploy_test.go
- NewCluster
- k8s/inspect.go
- convert_test.go
- transform_test.go
- manager_test.go
- commands.go
- coverage_test.go
- Bash Dev Script
- newTestOps
- PowerShell Dev Script
- Load
- switchplan_test.go
- newLocalOps
- diff_test.go
- renderDriver
- testing.T
- importops_test.go
- watch_test.go
- Solace Module Root
- .ApplyDefaults
- CLAUDE.md
- k8s/preflight_test.go
- ResolveDomainCerts
- watchLoop
- github.com/spf13/cobra.ShellCompDirective
- .logf
- annotate_test.go
- output_test.go
- Compose
- inject_test.go
- secrets_test.go
- exportconfig_test.go
- cli/platform_test.go
- VPNRepl
- checkRow
- config_test.go
- replConfig
- internal/k8s
- scripts.go
- haCfg
- .runMatePhase
- .dump
- Abbreviations
- replication_test.go
- .OperatorApply
- ReplSite
- MateConfig
- ParseBlocks
- rootless_test.go
- Test catalogue
- Platform
- Quadlet
- .readCLI
- Ops
- The eight rules
- .operatorRows
- What `import-config` applies
- validContainerConfig
- Operations
- imagepullsecret_test.go
- Role
- writeNode
- EnvPairs
- cli/watch_test.go
- newRootCmd
- adminsecret_test.go
- containerTransport
- replicationops_test.go
- .validateContainer
- limits_test.go
- Developer guide
- .BackupRevertActivity
- SameEndpoints
- replication_test.go
- AdminSecret
- render_test.go
- .validate
- RuleFor
- validPlatformConfig
- .checkUserManagerLimits
- gen-cert.sh
- platform_test.go
- Container
- .kubectl
- Troubleshooting
- Exporting and importing configuration
- Get
- Platform parity: solmq-conn-util and solace-util
- runLeader
- exit.go
- capRunner
- Container State Inspection
- operatorversion_test.go
- hostpath_test.go
- strings.Builder

## God Nodes (most connected - your core abstractions)
1. `App` - 216 edges
2. `newTestOps()` - 167 edges
3. `Config` - 147 edges
4. `cli_test.go` - 146 edges
5. `NewCluster()` - 144 edges
6. `ctrCfg()` - 133 edges
7. `Role` - 132 edges
8. `loadK8s()` - 120 edges
9. `Per-package doubles` - 120 edges
10. `newCapMgr()` - 114 edges

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

## Communities (111 total, 8 thin omitted)

### Community 0 - "App"
Cohesion: 0.07
Nodes (91): App, logArgs(), confirmAction(), confirmDelete(), confirmRestart(), lineSink(), progress(), warn() (+83 more)

### Community 1 - "newTestOps"
Cohesion: 0.06
Nodes (69): broker_test.go, cliRunNames(), countContains(), runCLISkeleton(), cliScriptNameFromDest(), leaderCalls(), TestCountContains(), TestDisableDefaultUsers() (+61 more)

### Community 2 - ".ns"
Cohesion: 0.09
Nodes (14): names_test.go, BrokerPodSuffixShape(), normalizeToList(), discoveryServiceName(), HARoles(), lbServiceName(), podName(), pvcName() (+6 more)

### Community 4 - "cli_test.go"
Cohesion: 0.07
Nodes (107): cli_test.go, Shared env fixtures, allowRuntime(), captureStderr(), collectPaths(), echoRunner(), fakeBinaryOnPath(), firstLine() (+99 more)

### Community 5 - "config_test.go"
Cohesion: 0.05
Nodes (53): Image, config_test.go, atoiPrefix(), assertContainerBlockDefaults(), assertContainerScaling(), envTree(), haNodesConfig(), TestAdditionalUserNamesDoNotCollideOnKubernetes() (+45 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (50): Commands, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker cli, solace-util broker configure (+42 more)

### Community 7 - "Manager"
Cohesion: 0.07
Nodes (13): scriptedMate, certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), TestPreflightFailureStopsLifecycle() (+5 more)

### Community 8 - "K8sConfig"
Cohesion: 0.11
Nodes (15): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, LoadBalancer, cut() (+7 more)

### Community 9 - "completion_test.go"
Cohesion: 0.22
Nodes (23): completion_test.go, driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion() (+15 more)

### Community 10 - "Command"
Cohesion: 0.06
Nodes (40): cmdField, guardedCmd, command_test.go, duration_test.go, execguard_test.go, internal/config, decodeCommand(), TestCommandArgsDoesNotAliasCommand() (+32 more)

### Community 11 - "deploy_test.go"
Cohesion: 0.09
Nodes (51): checkreport_test.go, deploy_test.go, TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestOperatorInstallVerdict(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestOperatorValidateAsksNothingAboutTheBroker(), TestStorageRows(), TestValidateConfigSectionNeverFails() (+43 more)

### Community 12 - "NewCluster"
Cohesion: 0.09
Nodes (67): operator_test.go, GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), RenderOperator(), splitAfterNamespace(), splitOperatorBundle() (+59 more)

### Community 13 - "k8s/inspect.go"
Cohesion: 0.11
Nodes (37): inspect_test.go, orNone(), age(), anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), podHealth() (+29 more)

### Community 14 - "convert_test.go"
Cohesion: 0.07
Nodes (68): doc, Result, segment, vars, convert_test.go, DecodeStrict(), boolOf(), commentSafe() (+60 more)

### Community 15 - "transform_test.go"
Cohesion: 0.10
Nodes (61): Block, TargetState, transform_test.go, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects() (+53 more)

### Community 16 - "manager_test.go"
Cohesion: 0.09
Nodes (98): manager_test.go, TestRestartCountComesFromSystemdOnPodman(), TestRestartCountUnknownWhenSystemctlFails(), fileExists(), NewManager(), assertMode(), containsStr(), ctrCfg() (+90 more)

### Community 17 - "commands.go"
Cohesion: 0.13
Nodes (74): opFunc, roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd() (+66 more)

### Community 18 - "coverage_test.go"
Cohesion: 0.07
Nodes (69): coverage_test.go, verify_ops_test.go, matchCLI(), ranContains(), TestDiagnostics(), TestDisableDefaultVPN(), writeFile(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList() (+61 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "newTestOps"
Cohesion: 0.07
Nodes (65): config_ops_test.go, internal/broker, semp_test.go, sempmate_test.go, newTestOps(), TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestEnableDefaultVPNFailureSkipsTheReadback() (+57 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Load"
Cohesion: 0.07
Nodes (50): scalingTier, scaling_test.go, minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs() (+42 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.08
Nodes (74): CLIRunner, fakeRun, PhaseKind, SiteState, SwitchAction, SwitchPhase, matechannel_test.go, switchplan_test.go (+66 more)

### Community 24 - "newLocalOps"
Cohesion: 0.14
Nodes (43): verify_local_test.go, TestImportOpsExportConfigFailsLoudAtTheFirstBadCapture(), TestImportOpsExportConfigScopeSelectsCLICommand(), TestImportOpsImportVerifyReExportsAndDiffs(), TestRouterNameReadsTheDefaultVPNCaptureHeader(), currentConfigScript(), cliScriptPath(), isCurl() (+35 more)

### Community 25 - "diff_test.go"
Cohesion: 0.12
Nodes (37): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, diff_test.go, ancestorReported(), blockLabel() (+29 more)

### Community 26 - "renderDriver"
Cohesion: 0.15
Nodes (22): chunk, chunkResult, driver_test.go, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce() (+14 more)

### Community 27 - "testing.T"
Cohesion: 0.07
Nodes (63): examples_test.go, namespace_test.go, statusreport_test.go, storage_test.go, runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand(), TestPodmanEUIDGuardSkips() (+55 more)

### Community 28 - "importops_test.go"
Cohesion: 0.07
Nodes (67): ImportResult, PlannedSection, importops_test.go, defaultVPNFirst(), describeScope(), ImportPlan, importIgnore(), orNone() (+59 more)

### Community 29 - "watch_test.go"
Cohesion: 0.16
Nodes (29): watch_test.go, subtractWatch(), deniedCluster(), deployJSON(), namedOperatorJSON(), operatorItem(), TestDesiredWatchAppendsTheBrokerNamespace(), TestFindOperatorDeploymentFallsBackWhenTheClusterWideListFails() (+21 more)

### Community 31 - ".ApplyDefaults"
Cohesion: 0.18
Nodes (14): tlssecret_test.go, applyContainerBlockDefaults(), podmanRunUser(), setDefault(), setDefaultCmd(), setDefaultInt(), xdgConfigHome(), TestNamingTheTLSSecretDoesNotInventCertPaths() (+6 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "k8s/preflight_test.go"
Cohesion: 0.21
Nodes (15): preflight_test.go, TestCanIAnswerReadsTheLastLine(), TestCanIAnswersWithoutFailing(), TestCanIArgsScopesByResourceKind(), TestCanIIsPermittedUnderEcho(), TestCanIRefusedReadsEveryRefusalSpelling(), TestCanIUnderEchoStillHonoursTheExecutionGuard(), TestPreflightAcceptsAWarnedYes() (+7 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (53): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domaincerts_resolve_test.go, domaincerts_test.go, caNameSafe() (+45 more)

### Community 35 - "watchLoop"
Cohesion: 0.31
Nodes (5): watchInterval(), stdoutIsTerminal(), watchLabel(), watchLoop(), withWatch()

### Community 36 - "github.com/spf13/cobra.ShellCompDirective"
Cohesion: 0.33
Nodes (6): completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), isEnvFileName(), matching()

### Community 37 - ".logf"
Cohesion: 0.09
Nodes (14): imageFromDeployment(), namespaceManifest(), ownedSecretNames(), operatorRunningImage(), TestOperatorRunningImageWithNoContainers(), desiredWatch(), splitWatch(), watchFromContainers() (+6 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.07
Nodes (48): Region, annotate_test.go, applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV() (+40 more)

### Community 40 - "output_test.go"
Cohesion: 0.14
Nodes (23): internal/output, output_test.go, columnWidths(), Sink, NewFunc(), pad(), sink(), TestKVBlockAlignsOnTheLongestKey() (+15 more)

### Community 41 - "Compose"
Cohesion: 0.17
Nodes (18): internal/render, servercert_test.go, Compose(), composeEscape(), ComposeProject(), ContainerSecrets(), ContainerSecret, SecretPreflight() (+10 more)

### Community 42 - "inject_test.go"
Cohesion: 0.17
Nodes (28): lineRole, serviceLine, shutdownStyle, svcKey, inject_test.go, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "secrets_test.go"
Cohesion: 0.13
Nodes (29): capCall, secrets_test.go, secretSummary(), setOrMissing(), solaceRows(), ResolveSecretValues(), bundleHash(), certCreate() (+21 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.16
Nodes (27): exportconfigReadCounter, opRunner, exportconfig_test.go, internal/cli, pluralVPN(), exportconfigDriverOutput(), exportconfigTransportOutput(), exportconfigVPNListOutput() (+19 more)

### Community 46 - "cli/platform_test.go"
Cohesion: 0.25
Nodes (23): platform_test.go, supported(), logArgs2(), runPlatform(), TestLogArgsBuildsOneSetForBothPlatforms(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects() (+15 more)

### Community 47 - "VPNRepl"
Cohesion: 0.09
Nodes (20): AdminState, QueueState, replMate, ValidVPNName(), validVPNName(), ReplRole, VPNRepl, ReplicationConfigResult (+12 more)

### Community 48 - "checkRow"
Cohesion: 0.21
Nodes (16): orValue(), setOrMissing(), setOrNone(), additionalUsersRow(), adminPassState(), adminSecretRow(), containsString(), info() (+8 more)

### Community 49 - "config_test.go"
Cohesion: 0.15
Nodes (31): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, span, AdditionalUser, Node, Redundancy (+23 more)

### Community 50 - "replConfig"
Cohesion: 0.28
Nodes (13): Replication, replication_test.go, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationAbsent() (+5 more)

### Community 51 - "internal/k8s"
Cohesion: 0.22
Nodes (15): adminsecret_test.go, check_test.go, internal/k8s, transport_test.go, AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight() (+7 more)

### Community 52 - "scripts.go"
Cohesion: 0.10
Nodes (43): showCmd, scripts_test.go, TestServerCert(), archiveConfigsScript(), assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript() (+35 more)

### Community 53 - "haCfg"
Cohesion: 0.07
Nodes (78): ops_test.go, prep_test.go, TestReachable(), TestResolveStorageClass(), newCluster(), TestApplyOnStdin(), TestDeleteStdin(), RestartOrder() (+70 more)

### Community 54 - ".runMatePhase"
Cohesion: 0.36
Nodes (6): isRunCLIRejection(), newlineIf(), replPhase1Rejected(), replPhase2Rejected(), replTransportFailure(), replVPNList()

### Community 55 - ".dump"
Cohesion: 0.09
Nodes (75): mateChannelFunc, opCall, ops_k8s_test.go, Per-package doubles, capture(), captureStdout(), failDisableDefaultUsersUpload(), isDeletePod() (+67 more)

### Community 56 - "Abbreviations"
Cohesion: 0.20
Nodes (9): Abbreviations, Flag shorthands, Node roles, Platforms, Command reference, Global flags, Index, Reading this reference (+1 more)

### Community 57 - "replication_test.go"
Cohesion: 0.10
Nodes (44): replRig, Commands, solace-util broker, Configuring a site, Data replication, Permissions, Reaching the mate, Switching roles (+36 more)

### Community 58 - ".OperatorApply"
Cohesion: 0.24
Nodes (6): operatorEscalationProbes(), operatorProbes(), canIAnswer(), canIRefused(), missingNamespace(), probe

### Community 59 - "ReplSite"
Cohesion: 0.06
Nodes (57): interactiveFailRunner, runFailRunner, Repository layout, internal/engine, matechannel_test.go, resolve_test.go, runner_test.go, noEnv (+49 more)

### Community 60 - "MateConfig"
Cohesion: 0.17
Nodes (15): cliTransport(), MateConfig, labelValue(), parseHostPort(), parseShowReplicationAppliance(), parseShowReplicationSoftware(), renderMateAppliance(), renderMateSoftware() (+7 more)

### Community 61 - "ParseBlocks"
Cohesion: 0.08
Nodes (42): Capture, Omission, token, blocks_test.go, omit_test.go, Annotate(), TestAnnotateIsDeterministic(), TestAnnotateRegionAndSectionFencesAreBalanced() (+34 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.19
Nodes (40): rootless_test.go, failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+32 more)

### Community 63 - "Test catalogue"
Cohesion: 0.19
Nodes (14): examples_test.go, Fixtures and doubles, internal/convert, internal/examples, internal/tools/vulnjudge, Running the tests, Summary, Test catalogue (+6 more)

### Community 64 - "Platform"
Cohesion: 0.13
Nodes (21): layer, Injectable seams, defaultLocalAddrs(), confirmDowngrade(), confirmLayer(), confirmScope(), checkAllowCommand(), containerRole() (+13 more)

### Community 65 - "Quadlet"
Cohesion: 0.15
Nodes (17): Coverage, ValidScriptName(), ContainerNoFile(), HealthCheck, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), escapePercent(), healthCmd() (+9 more)

### Community 66 - ".readCLI"
Cohesion: 0.14
Nodes (12): BrokerType, cliMate, transcript_test.go, bannerType(), checkSameType(), stripTranscript(), showReplicationScript(), showVPNReplicationScript() (+4 more)

### Community 67 - "Ops"
Cohesion: 0.08
Nodes (35): euid_test.go, Ops, New(), confirmActionStrict(), step(), newConvertCmd(), runConvert(), newExamplesCmd() (+27 more)

### Community 68 - "The eight rules"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 69 - ".operatorRows"
Cohesion: 0.26
Nodes (9): storageClassSuitable(), allCustomMounted(), failRow(), okRow(), portRows(), skipRow(), TestCheckReportFailedCounts(), TestPortRowsNeverFail() (+1 more)

### Community 70 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 71 - "validContainerConfig"
Cohesion: 0.07
Nodes (35): redundancy_test.go, TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes(), TestAdditionalUserNamesCollideOnDocker(), TestValidateAdditionalUserClashesWithABuiltIn(), TestValidateAdditionalUsers(), TestValidateContainerBadNetworkMode(), TestValidateContainerBridge() (+27 more)

### Community 72 - "Operations"
Cohesion: 0.14
Nodes (14): Bringing up a fresh cluster, Every destructive command confirms, Exit codes, Extra CLI users differ by platform, Operations, Post-deployment configuration order, Removing a broker: what stays, what goes, Removing the operator does not always remove it (+6 more)

### Community 73 - "imagepullsecret_test.go"
Cohesion: 0.23
Nodes (10): secretRef, imagepullsecret_test.go, pullSecretCfg(), TestImagePullSecretDerivedSuffixFitsWhereAdditionalUsersAlreadyFits(), TestImagePullSecretNameBringYourOwn(), TestImagePullSecretNameConfiguredNameWins(), TestImagePullSecretNameDerivedWhenCredentialsButNoName(), TestImagePullSecretNameEmptyWithNeither() (+2 more)

### Community 74 - "Role"
Cohesion: 0.05
Nodes (27): covUploadFailsTo, downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile (+19 more)

### Community 75 - "writeNode"
Cohesion: 0.21
Nodes (12): WeightedNodeTerm, NodeAffinity, NodeMatchExpr, Placement, validateMatchExprs(), validatePlacementAffinity(), parseToleration(), TestParseToleration() (+4 more)

### Community 76 - "EnvPairs"
Cohesion: 0.18
Nodes (12): containerSecretSpecs(), EnvPairs(), groupKey(), itoa(), secretFilePath(), assertNoCheckoutPath(), envLines(), TestAdditionalUsersReachBothHalves() (+4 more)

### Community 77 - "cli/watch_test.go"
Cohesion: 0.35
Nodes (11): watch_test.go, stopAfter(), TestOperatorStatusWatches(), TestStatusWithoutWatchRunsOnce(), TestWatchAppendsWhenNotATerminal(), TestWatchEndsQuietlyWhenInterrupted(), TestWatchIntervalIsValidated(), TestWatchKeepsGoingAfterAFailedRefresh() (+3 more)

### Community 78 - "newRootCmd"
Cohesion: 0.08
Nodes (43): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, commanddoc_test.go, exit_test.go, TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), applyAliases() (+35 more)

### Community 79 - "adminsecret_test.go"
Cohesion: 0.21
Nodes (10): adminsecret_test.go, adminCfg(), TestAdminSecretNameFollowsTheStates(), TestAdminSecretRefusesTheOperatorsOwnNames(), TestContainersStillRequireTheAdminPassword(), TestKubernetesValidatesWithoutAnAdminPassword(), TestMonitorPasswordAndKeyNeedTheAdminPassword(), TestOperatorAdminSecretName() (+2 more)

### Community 81 - "replicationops_test.go"
Cohesion: 0.17
Nodes (34): replicationops_test.go, hasCall(), replicationVPNLines(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), siteAEntry() (+26 more)

### Community 82 - ".validateContainer"
Cohesion: 0.05
Nodes (26): keyValueEntries, ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, TestRedundancyEnabled(), viaDescription(), checkCredentialChars() (+18 more)

### Community 83 - "limits_test.go"
Cohesion: 0.09
Nodes (51): limits_test.go, main_test.go, limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessCoreFloorFollowsTheConfiguredLimit(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips() (+43 more)

### Community 84 - "Developer guide"
Cohesion: 0.13
Nodes (15): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Tests, Toolchain pin (+7 more)

### Community 85 - ".BackupRevertActivity"
Cohesion: 0.11
Nodes (14): backupTarget, Credential, sempMate, transport_test.go, anyHTTP2xx(), curlConfigFlag(), curlConfigLine(), httpBody() (+6 more)

### Community 86 - "SameEndpoints"
Cohesion: 0.26
Nodes (12): containsEndpoint(), normTransport(), RenderMateRemovals(), SameEndpoints(), mateAppliance(), TestParseShowReplicationAppliance(), TestRenderMateAppliance(), TestRenderMateRemovals() (+4 more)

### Community 87 - "replication_test.go"
Cohesion: 0.13
Nodes (40): colSpan, replication_test.go, dashSpans(), flagByte(), gutterClear(), ParseShowReplication(), ParseVPNReplication(), RenderMate() (+32 more)

### Community 88 - "AdminSecret"
Cohesion: 0.11
Nodes (41): TLS, The env-file generator page, secrets_test.go, tlsSecretDigest(), Config, renderOperatorNS(), GenSecrets(), AdditionalUsersSecret() (+33 more)

### Community 89 - "render_test.go"
Cohesion: 0.16
Nodes (38): render_test.go, boolStr(), BrokerCR(), containerArtifacts(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets(), TestArtifactsCarryNoWideningTokens() (+30 more)

### Community 90 - ".validate"
Cohesion: 0.16
Nodes (9): solace-util, solace-util operator, Docker and Podman mechanics, Re-deploying is safe and explicit, Rootless podman prerequisites, What the broker cannot gain, Which user the broker runs as, TestCheckReportSkipsEmptySections() (+1 more)

### Community 91 - "RuleFor"
Cohesion: 0.11
Nodes (36): Disposition, SectionRule, appliance_test.go, importdoc_test.go, sections_test.go, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent() (+28 more)

### Community 93 - "validPlatformConfig"
Cohesion: 0.18
Nodes (10): artifactvalues_test.go, portOrRange(), TestValidateContainerArtifactValues(), TestValidateImageReference(), TestValidateK8sArtifactValues(), TestValidContainerPort(), validPlatformConfig(), validContainerPort() (+2 more)

### Community 95 - ".checkUserManagerLimits"
Cohesion: 0.24
Nodes (6): delegateSetting(), limitText(), missingControllers(), pamCoreValue(), parseLimit(), parseUnitProps()

### Community 96 - "gen-cert.sh"
Cohesion: 0.27
Nodes (13): ask(), askPass(), askSan(), askYesNo(), buildSanList(), checkExit(), concatCnIfNotBlank(), echoUsage() (+5 more)

### Community 97 - "platform_test.go"
Cohesion: 0.24
Nodes (10): platform_test.go, PlatformAbbrev(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint(), TestDetectPlatformsMissingFile(), TestEveryPlatformHasExactlyOneShortForm(), TestParsePlatform(), TestParsePlatformErrorTeachesCanonical() (+2 more)

### Community 98 - "Container"
Cohesion: 0.27
Nodes (6): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestContainerCoreLimitSpellings()

### Community 99 - ".kubectl"
Cohesion: 0.09
Nodes (21): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+13 more)

### Community 100 - "Troubleshooting"
Cohesion: 0.22
Nodes (9): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Forbidden: cannot list deployments at the cluster scope, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting (+1 more)

### Community 101 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 102 - "Get"
Cohesion: 0.36
Nodes (9): Example, All(), Get(), List(), Names(), read(), TestEveryNameResolves(), TestGetRejectsAnUnknownName() (+1 more)

### Community 103 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.29
Nodes (6): Contract, Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 104 - "runLeader"
Cohesion: 0.60
Nodes (4): leaderRun, runLeader(), TestAssertLeaderAsksBeforeActingFromABackup(), TestAssertLeaderPrimaryAndMonitorNeverAsk()

### Community 112 - "capRunner"
Cohesion: 0.10
Nodes (29): capRunner, internal/container, preflight_test.go, runtime_test.go, callIndex(), TestManagerLogsCLIShell(), TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestComposeSecretEnvNamesCannotBeSystemVars() (+21 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.10
Nodes (23): containerState, healthState, inspect_test.go, decodeInspect(), orUnknown(), printable(), inspectReply(), kvValue() (+15 more)

### Community 126 - "operatorversion_test.go"
Cohesion: 0.14
Nodes (22): operatorversion_test.go, compareVersions(), imageTag(), operatorVersionWarning(), parseVersion(), operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade() (+14 more)

### Community 128 - "hostpath_test.go"
Cohesion: 0.08
Nodes (31): hostpath_test.go, guardConfig(), BaseName(), CheckHostPath(), expandTilde(), expandTildeToken(), HasPathSeparator(), IsAbsHostPath() (+23 more)

### Community 155 - "strings.Builder"
Cohesion: 0.09
Nodes (31): shorthand, How a short form is resolved, mdRow(), renderAbbrevDocs(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable() (+23 more)

## Knowledge Gaps
- **13 isolated node(s):** `solace`, `NO_COLOR`, `bash/000-env.sh`, `bash/010-deploy-operator.sh`, `bash/020-deploy-broker.sh` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 187 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `App` to `Platform`, `watchLoop`, `Ops`, `cli_test.go`, `github.com/spf13/cobra.ShellCompDirective`, `Manager`, `exportconfig_test.go`, `cli/platform_test.go`, `newRootCmd`, `commands.go`, `config_test.go`, `.dump`, `AdminSecret`, `replication_test.go`, `ReplSite`?**
  _High betweenness centrality (0.042) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `newTestOps()` (e.g. with `importops_test.go` and `Injectable seams`) actually correct?**
  _`newTestOps()` has 117 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `NO_COLOR`, `bash/000-env.sh` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.07106227106227106 - nodes in this community are weakly interconnected._
- **Why does `Config` connect `AdminSecret` to `App`, `hostpath_test.go`, `newTestOps`, `.ns`, `cli_test.go`, `config_test.go`, `kubectlTransport`, `Manager`, `K8sConfig`, `Command`, `deploy_test.go`, `NewCluster`, `convert_test.go`, `manager_test.go`, `newTestOps`, `Load`, `testing.T`, `strings.Builder`, `.ApplyDefaults`, `k8s/preflight_test.go`, `ResolveDomainCerts`, `.logf`, `Compose`, `secrets_test.go`, `checkRow`, `config_test.go`, `replConfig`, `internal/k8s`, `haCfg`, `.dump`, `ReplSite`, `Test catalogue`, `Platform`, `Quadlet`, `Ops`, `.operatorRows`, `validContainerConfig`, `imagepullsecret_test.go`, `Role`, `writeNode`, `EnvPairs`, `adminsecret_test.go`, `containerTransport`, `.validateContainer`, `render_test.go`, `validPlatformConfig`, `Container`, `.kubectl`, `capRunner`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Are the 9 inferred relationships involving `Config` (e.g. with `config_test.go` and `execguard_test.go`) actually correct?**
  _`Config` has 9 INFERRED edges - model-reasoned connections that need verification._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.05754385964912281 - nodes in this community are weakly interconnected._
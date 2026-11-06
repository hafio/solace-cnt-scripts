# Graph Report - solace-cnt-scripts  (2026-11-06)

## Corpus Check
- 204 files · ~697,975 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .golden 17, (none) 3, .cli 2)

## Summary
- 3998 nodes · 20473 edges · 109 communities (101 shown, 8 thin omitted)
- Extraction: 73% EXTRACTED · 27% INFERRED · 0% AMBIGUOUS · INFERRED: 5630 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `388b78c4`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- App
- newTestOps
- .kubectl
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
- localCfg
- PowerShell Dev Script
- Load
- MateChannel
- newLocalOps
- diff_test.go
- renderDriver
- testing.T
- importops_test.go
- watch_test.go
- Solace Module Root
- .ApplyDefaults
- CLAUDE.md
- NewCluster
- ResolveDomainCerts
- watchLoop
- github.com/spf13/cobra.ShellCompDirective
- .operatorRows
- annotate_test.go
- Role
- Compose
- inject_test.go
- secrets_test.go
- exportconfig_test.go
- switchplan_test.go
- ReplRole
- .validate
- config_test.go
- newSEMPTestMate
- eqArgs
- config_ops_test.go
- haCfg
- .runMatePhase
- .dump
- Abbreviations
- replication_test.go
- .OperatorApply
- ReplSite
- ReplVia
- blocks_test.go
- rootless_test.go
- Test catalogue
- Injectable seams
- Quadlet
- internal/broker
- Ops
- The eight rules
- HARoles
- What `import-config` applies
- twoMates
- Operations
- internal/config
- .RunCLI
- writeNode
- EnvPairs
- cli/watch_test.go
- newRootCmd
- adminsecret_test.go
- context.Context
- replicationops_test.go
- Platform
- limits_test.go
- Developer guide
- .BackupRevertActivity
- namespace_test.go
- broker/replication_test.go
- AdminSecret
- render_test.go
- .decodeScalingEntry
- RuleFor
- validPlatformConfig
- .checkUserManagerLimits
- gen-cert.sh
- .RedundancyEnabled
- Container
- Configuration
- ResolveEnvPath
- Image
- DetectPlatforms
- Platform parity: solmq-conn-util and solace-util
- exit.go
- capRunner
- Container State Inspection
- operatorversion_test.go
- strings.Builder

## God Nodes (most connected - your core abstractions)
1. `App` - 216 edges
2. `newTestOps()` - 167 edges
3. `Config` - 148 edges
4. `cli_test.go` - 146 edges
5. `NewCluster()` - 144 edges
6. `ctrCfg()` - 133 edges
7. `Role` - 132 edges
8. `loadK8s()` - 120 edges
9. `Per-package doubles` - 120 edges
10. `newCapMgr()` - 114 edges

## Surprising Connections (you probably didn't know these)
- `duration_test.go` --references--> `TestCanonicalDuration()`  [INFERRED]
  docs/test.md → internal/config/duration_test.go
- `The preflight` --references--> `info()`  [INFERRED]
  docs/operations.md → internal/k8s/checkreport.go
- `annotate_test.go` --references--> `Annotate()`  [INFERRED]
  docs/test.md → internal/broker/annotate.go
- `annotate_test.go` --references--> `TestAnnotateRoundTripPreservesBlocks()`  [INFERRED]
  docs/test.md → internal/broker/annotate_test.go
- `annotate_test.go` --references--> `TestAnnotateRoundTripCaptureIsMarked()`  [INFERRED]
  docs/test.md → internal/broker/annotate_test.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (109 total, 8 thin omitted)

### Community 0 - "App"
Cohesion: 0.06
Nodes (101): App, logArgs(), confirmAction(), confirmDelete(), warn(), runExport(), diagnosticsOptions(), exportScope() (+93 more)

### Community 1 - "newTestOps"
Cohesion: 0.08
Nodes (69): broker_test.go, countContains(), newTestOps(), TestCountContains(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultUsersRefusesAQuotedVPNName() (+61 more)

### Community 2 - ".kubectl"
Cohesion: 0.07
Nodes (16): internal/k8s, runtime_test.go, transport_test.go, Cluster, normalizeToList(), stsName(), bareKind(), joinYAMLDocs() (+8 more)

### Community 3 - "kubectlTransport"
Cohesion: 0.39
Nodes (3): podName(), shSingleQuote(), kubectlTransport

### Community 4 - "cli_test.go"
Cohesion: 0.07
Nodes (112): cli_test.go, Shared env fixtures, allowRuntime(), captureStderr(), collectPaths(), echoRunner(), fakeBinaryOnPath(), firstLine() (+104 more)

### Community 5 - "config_test.go"
Cohesion: 0.05
Nodes (66): config_test.go, redundancy_test.go, haNodesConfig(), TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes(), TestAdditionalUserNamesCollideOnDocker(), TestAdditionalUserNamesDoNotCollideOnKubernetes(), TestAdditionalUserPasswordCharsAreFreeOnKubernetes() (+58 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (52): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker cli (+44 more)

### Community 7 - "Manager"
Cohesion: 0.07
Nodes (20): preflight_test.go, certDigest(), composeNeedsSecretValues(), exactName(), fileExists(), Manager, orNone(), platformTitle() (+12 more)

### Community 8 - "K8sConfig"
Cohesion: 0.11
Nodes (15): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, LoadBalancer, cut() (+7 more)

### Community 9 - "completion_test.go"
Cohesion: 0.22
Nodes (23): completion_test.go, driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion() (+15 more)

### Community 10 - "Command"
Cohesion: 0.07
Nodes (35): cmdField, guardedCmd, command_test.go, execguard_test.go, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandArgv(), TestCommandNameAndArgs() (+27 more)

### Community 11 - "deploy_test.go"
Cohesion: 0.16
Nodes (26): deploy_test.go, brokerJSON(), logBuf(), TestConfirmBrokerAppliedAcceptsTheObjectItApplied(), TestConfirmBrokerAppliedFailsOnADifferentObject(), TestConfirmBrokerAppliedFailsWhenTheObjectIsNotThere(), TestConfirmBrokerAppliedSurfacesReadFailures(), TestConfirmBrokerAppliedTreatsSilenceAsSkipped() (+18 more)

### Community 12 - "NewCluster"
Cohesion: 0.11
Nodes (61): operator_test.go, GenOperator(), operatorImage(), RenderOperator(), boolPtr(), bundledDefaultPorts(), commentedExamplePorts(), covSwapBundle() (+53 more)

### Community 13 - "k8s/inspect.go"
Cohesion: 0.13
Nodes (34): inspect_test.go, anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), podHealth(), pvcLevel(), replicaLevel() (+26 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (69): doc, Result, segment, vars, The keys, convert_test.go, DecodeStrict(), boolOf() (+61 more)

### Community 15 - "transform_test.go"
Cohesion: 0.10
Nodes (61): Block, TargetState, transform_test.go, BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects(), ClearTargetVirtualHostnames() (+53 more)

### Community 16 - "manager_test.go"
Cohesion: 0.09
Nodes (94): manager_test.go, GuardPodmanEUID(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountUnknownWhenSystemctlFails(), assertMode(), containsStr(), ctrCfg(), hasCall() (+86 more)

### Community 17 - "commands.go"
Cohesion: 0.13
Nodes (73): opFunc, roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd() (+65 more)

### Community 18 - "coverage_test.go"
Cohesion: 0.08
Nodes (49): coverage_test.go, ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultVPNDisableError() (+41 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "localCfg"
Cohesion: 0.15
Nodes (28): semp_test.go, bridgeHostPort(), sempPort(), sempV1OK(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv() (+20 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Load"
Cohesion: 0.08
Nodes (45): scalingTier, scaling_test.go, minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadReadError(), TestLoadRejectsTheOldK8sSection() (+37 more)

### Community 23 - "MateChannel"
Cohesion: 0.16
Nodes (27): CLIRunner, fakeRun, matechannel_test.go, MateChannel, NewCLIMate(), mateReplies(), newTestMate(), TestLocalMateRunsThroughOps() (+19 more)

### Community 24 - "newLocalOps"
Cohesion: 0.12
Nodes (39): verify_local_test.go, cliScriptPath(), isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn(), TestDetectRoleAddrsError() (+31 more)

### Community 25 - "diff_test.go"
Cohesion: 0.06
Nodes (61): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, diff_test.go, internal/output, output_test.go (+53 more)

### Community 26 - "renderDriver"
Cohesion: 0.15
Nodes (22): chunk, chunkResult, driver_test.go, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce() (+14 more)

### Community 27 - "testing.T"
Cohesion: 0.04
Nodes (107): examples_test.go, generatorpage_test.go, hostpath_test.go, platform_test.go, statusreport_test.go, storage_test.go, TestExamplesAcceptsAPlatformAbbreviation(), TestExamplesBareEmitsTheFullSchema() (+99 more)

### Community 28 - "importops_test.go"
Cohesion: 0.05
Nodes (73): ImportResult, PlannedSection, importops_test.go, isPragma(), defaultVPNFirst(), describeScope(), ImportPlan, importIgnore() (+65 more)

### Community 29 - "watch_test.go"
Cohesion: 0.12
Nodes (29): watch_test.go, TestConfirmNoDowngradeStillAsksWhenTheFallbackFindsANewerOperator(), subtractWatch(), deniedCluster(), deniedRunner(), deployJSON(), namedOperatorJSON(), operatorItem() (+21 more)

### Community 31 - ".ApplyDefaults"
Cohesion: 0.15
Nodes (12): applyContainerBlockDefaults(), podmanRunUser(), setDefault(), setDefaultCmd(), setDefaultInt(), xdgConfigHome(), containerMem(), cpuSetCount() (+4 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "NewCluster"
Cohesion: 0.11
Nodes (42): checkreport_test.go, preflight_test.go, TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestCheckReportSkipsEmptySections(), TestOperatorInstallVerdict(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestOperatorValidateAsksNothingAboutTheBroker(), TestPortRowsNeverFail() (+34 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (53): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domaincerts_resolve_test.go, domaincerts_test.go, caNameSafe() (+45 more)

### Community 35 - "watchLoop"
Cohesion: 0.32
Nodes (5): watchInterval(), stdoutIsTerminal(), watchLabel(), watchLoop(), withWatch()

### Community 36 - "github.com/spf13/cobra.ShellCompDirective"
Cohesion: 0.33
Nodes (6): completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), isEnvFileName(), matching()

### Community 37 - ".operatorRows"
Cohesion: 0.10
Nodes (21): containsString(), installSteps(), compareVersions(), imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), operatorRunningImage() (+13 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.07
Nodes (48): Region, annotate_test.go, applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV() (+40 more)

### Community 40 - "Role"
Cohesion: 0.08
Nodes (20): covUploadFailsTo, downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile (+12 more)

### Community 41 - "Compose"
Cohesion: 0.22
Nodes (13): servercert_test.go, Compose(), ComposeProject(), ContainerSecrets(), ContainerSecret, SecretPreflight(), TestComposeProjectFoldsToComposesGrammar(), TestComposeProjectIsDeclaredNotDerived() (+5 more)

### Community 42 - "inject_test.go"
Cohesion: 0.17
Nodes (28): lineRole, serviceLine, shutdownStyle, svcKey, inject_test.go, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "secrets_test.go"
Cohesion: 0.16
Nodes (29): secrets_test.go, solaceRows(), ResolveSecretValues(), bundleHash(), certCreate(), certFixture(), TestCertSecretLabelDrivesTheRestart(), TestContainerBundleCarriesTheChain() (+21 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.14
Nodes (31): exportconfigReadCounter, opRunner, euid_test.go, exportconfig_test.go, internal/cli, runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand() (+23 more)

### Community 46 - "switchplan_test.go"
Cohesion: 0.15
Nodes (34): PhaseKind, SiteState, SwitchAction, SwitchPhase, switchplan_test.go, BuildSwitchPlan(), confirmDemotions(), ExecuteSwitchPlan() (+26 more)

### Community 47 - "ReplRole"
Cohesion: 0.24
Nodes (10): ReplRole, ReplicationConfigResult, PlannedRoles(), RoleAtSite(), TestPlannedRolesNamesEveryListedVPN(), TestRoleAtSiteIsTheComplement(), setReplicationRoleScript(), sempReplRole() (+2 more)

### Community 48 - ".validate"
Cohesion: 0.13
Nodes (26): Docker and Podman mechanics, Re-deploying is safe and explicit, Rootless podman prerequisites, What the broker cannot gain, Which user the broker runs as, setOrMissing(), setOrNone(), storageClassSuitable() (+18 more)

### Community 49 - "config_test.go"
Cohesion: 0.16
Nodes (27): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, showCmd, span, AdditionalUser, Node (+19 more)

### Community 50 - "newSEMPTestMate"
Cohesion: 0.20
Nodes (19): sempmate_test.go, covSEMPMate(), newSEMPTestMate(), sempReply(), sempSite(), TestSEMPMateChecksExecuteResult(), TestSEMPMateDescribeNamesTheSiteAndEndpoint(), TestSEMPMateEscapesTheVPNName() (+11 more)

### Community 51 - "eqArgs"
Cohesion: 0.10
Nodes (43): check_test.go, cluster_test.go, ops_test.go, TestCheckStopsProbingWhenUnreachable(), TestReachable(), TestResolveStorageClass(), newCluster(), TestApplyOnStdin() (+35 more)

### Community 52 - "config_ops_test.go"
Cohesion: 0.18
Nodes (12): config_ops_test.go, TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(), TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator(), TestExecShellScriptRefusesAnUnsafeNameBeforeUploading(), TestExecShellScriptTrustsTheExitStatusNotTheTranscript(), TestExecShellScriptUploadFailureRunsNothing() (+4 more)

### Community 53 - "haCfg"
Cohesion: 0.10
Nodes (53): adminsecret_test.go, prep_test.go, AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestOperatorAdminSecretNameMatchesTheBrokerNames() (+45 more)

### Community 54 - ".runMatePhase"
Cohesion: 0.36
Nodes (6): isRunCLIRejection(), newlineIf(), replPhase1Rejected(), replPhase2Rejected(), replTransportFailure(), replVPNList()

### Community 55 - ".dump"
Cohesion: 0.08
Nodes (79): leaderRun, mateChannelFunc, opCall, ops_k8s_test.go, Per-package doubles, capture(), captureStdout(), failDisableDefaultUsersUpload() (+71 more)

### Community 56 - "Abbreviations"
Cohesion: 0.20
Nodes (9): Abbreviations, Flag shorthands, Node roles, Platforms, Command reference, Global flags, Index, Reading this reference (+1 more)

### Community 57 - "replication_test.go"
Cohesion: 0.10
Nodes (44): replRig, Commands, solace-util broker, Configuring a site, Data replication, Permissions, Reaching the mate, Switching roles (+36 more)

### Community 58 - ".OperatorApply"
Cohesion: 0.54
Nodes (3): canIAnswer(), canIRefused(), probe

### Community 59 - "ReplSite"
Cohesion: 0.06
Nodes (60): interactiveFailRunner, runFailRunner, Replication, secretRef, Repository layout, internal/engine, matechannel_test.go, replication_test.go (+52 more)

### Community 60 - "ReplVia"
Cohesion: 0.12
Nodes (13): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, viaDescription(), CheckSecretKeyRef(), sortedKeys(), validateReplEndpoints() (+5 more)

### Community 61 - "blocks_test.go"
Cohesion: 0.14
Nodes (24): token, blocks_test.go, TestAnnotateIsDeterministic(), brokerTypeFromSchema(), firstQuoted(), newBlock(), blockByIndex(), loadSample() (+16 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.19
Nodes (40): rootless_test.go, failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+32 more)

### Community 63 - "Test catalogue"
Cohesion: 0.33
Nodes (6): Fixtures and doubles, internal/convert, internal/render, Running the tests, Summary, Test catalogue

### Community 64 - "Injectable seams"
Cohesion: 0.15
Nodes (12): layer, Injectable seams, resolve_test.go, confirmDowngrade(), confirmLayer(), confirmRestart(), confirmScope(), TestChildEnvNamesAreNotSystemVariables() (+4 more)

### Community 65 - "Quadlet"
Cohesion: 0.14
Nodes (19): Coverage, ValidScriptName(), defaultLocalAddrs(), TestDefaultLocalAddrs(), ContainerNoFile(), HealthCheck, TestLimitsCheckAssertsWhatTheArtifactAsks(), healthCmd() (+11 more)

### Community 66 - "internal/broker"
Cohesion: 0.11
Nodes (26): BrokerType, Capture, Omission, internal/broker, omit_test.go, transcript_test.go, Annotate(), TestAnnotateRoundTripCaptureIsMarked() (+18 more)

### Community 67 - "Ops"
Cohesion: 0.08
Nodes (31): Ops, New(), confirmActionStrict(), lineSink(), progress(), step(), newConvertCmd(), runConvert() (+23 more)

### Community 68 - "The eight rules"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 69 - "HARoles"
Cohesion: 0.13
Nodes (13): names_test.go, BrokerPodSuffixShape(), orNone(), orValue(), allCustomMounted(), age(), discoveryServiceName(), HARoles() (+5 more)

### Community 70 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 71 - "twoMates"
Cohesion: 0.33
Nodes (13): requirePrimaryActive(), requireVirtualRouterNamesMatch(), sortedSiteNames(), SwitchPreflight(), replSites(), TestSwitchPreflightHealthyPair(), TestSwitchPreflightNeedsExactlyTwoSites(), TestSwitchPreflightRequiresMatchingVirtualRouterNames() (+5 more)

### Community 72 - "Operations"
Cohesion: 0.07
Nodes (29): A replication switch refuses before changing anything, A wrapper runtime is refused, Bringing up a fresh cluster, Broker scope is a fixed classification, Docker compose secrets need compose 2.23.1+, Every destructive command confirms, Exit codes, Exporting and importing configuration (+21 more)

### Community 73 - "internal/config"
Cohesion: 0.18
Nodes (18): duration_test.go, imagepullsecret_test.go, internal/config, tlssecret_test.go, pullSecretCfg(), TestImagePullSecretDerivedSuffixFitsWhereAdditionalUsersAlreadyFits(), TestImagePullSecretNameBringYourOwn(), TestImagePullSecretNameConfiguredNameWins() (+10 more)

### Community 74 - ".RunCLI"
Cohesion: 0.05
Nodes (50): scripts_test.go, cliRunNames(), field(), runCLISkeleton(), validCLILine(), validName(), rejectionIn(), archiveConfigsScript() (+42 more)

### Community 75 - "writeNode"
Cohesion: 0.15
Nodes (17): WeightedNodeTerm, NodeAffinity, NodeMatchExpr, Placement, PodAffinityTerm, validateMatchExprs(), validatePlacementAffinity(), parseToleration() (+9 more)

### Community 76 - "EnvPairs"
Cohesion: 0.13
Nodes (16): NodeIdentity, containerSecretSpecs(), EnvPairs(), groupKey(), itoa(), secretFilePath(), ServerCertBundlePath(), assertNoCheckoutPath() (+8 more)

### Community 77 - "cli/watch_test.go"
Cohesion: 0.35
Nodes (11): watch_test.go, stopAfter(), TestOperatorStatusWatches(), TestStatusWithoutWatchRunsOnce(), TestWatchAppendsWhenNotATerminal(), TestWatchEndsQuietlyWhenInterrupted(), TestWatchIntervalIsValidated(), TestWatchKeepsGoingAfterAFailedRefresh() (+3 more)

### Community 78 - "newRootCmd"
Cohesion: 0.07
Nodes (49): Goldens, abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, commanddoc_test.go, exit_test.go, TestAbbreviationDocs(), TestFlagShorthandsAreConsistent() (+41 more)

### Community 79 - "adminsecret_test.go"
Cohesion: 0.21
Nodes (10): adminsecret_test.go, adminCfg(), TestAdminSecretNameFollowsTheStates(), TestAdminSecretRefusesTheOperatorsOwnNames(), TestContainersStillRequireTheAdminPassword(), TestKubernetesValidatesWithoutAnAdminPassword(), TestMonitorPasswordAndKeyNeedTheAdminPassword(), TestOperatorAdminSecretName() (+2 more)

### Community 80 - "context.Context"
Cohesion: 0.08
Nodes (5): scriptedMate, replMate, containerTransport, shSingleQuote(), ownedSecretNames()

### Community 81 - "replicationops_test.go"
Cohesion: 0.11
Nodes (49): replicationops_test.go, verify_ops_test.go, hasCall(), matchCLI(), TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), TestDisableDefaultUsersShowVPNError() (+41 more)

### Community 82 - "Platform"
Cohesion: 0.06
Nodes (37): keyValueEntries, How a short form is resolved, applyAliases(), checkFlagPlatforms(), commandPlatforms(), declaredList(), parsePlatformList(), promptPlatform() (+29 more)

### Community 83 - "limits_test.go"
Cohesion: 0.09
Nodes (52): internal/tools/vulnjudge, limits_test.go, main_test.go, limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessCoreFloorFollowsTheConfiguredLimit() (+44 more)

### Community 84 - "Developer guide"
Cohesion: 0.25
Nodes (8): Build, Dev script tasks, Developer guide, Gates, Releases, Tests, The env-file generator page, Toolchain pin

### Community 85 - ".BackupRevertActivity"
Cohesion: 0.13
Nodes (13): backupTarget, Credential, sempMate, anyHTTP2xx(), curlConfigFlag(), curlConfigLine(), httpBody(), lastHTTPStatus() (+5 more)

### Community 86 - "namespace_test.go"
Cohesion: 0.31
Nodes (11): namespace_test.go, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 87 - "broker/replication_test.go"
Cohesion: 0.05
Nodes (80): AdminState, cliMate, colSpan, QueueState, replication_test.go, ValidVPNName(), validVPNName(), cliTransport() (+72 more)

### Community 88 - "AdminSecret"
Cohesion: 0.11
Nodes (39): TLS, secrets_test.go, Config, renderOperatorNS(), GenSecrets(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret() (+31 more)

### Community 89 - "render_test.go"
Cohesion: 0.18
Nodes (33): render_test.go, boolStr(), BrokerCR(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets(), TestBridgePublishesOnlyTheListedPorts(), TestBrokerCRQuotesTheImageReference() (+25 more)

### Community 90 - ".decodeScalingEntry"
Cohesion: 0.32
Nodes (5): scalingKey, scalingSpelling, Scaling, scalingKeyIndex(), scalingKeyList()

### Community 91 - "RuleFor"
Cohesion: 0.10
Nodes (40): Disposition, SectionRule, appliance_test.go, importdoc_test.go, sections_test.go, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent() (+32 more)

### Community 93 - "validPlatformConfig"
Cohesion: 0.20
Nodes (10): artifactvalues_test.go, checkTimezone(), portOrRange(), TestCheckTimezone(), TestValidateContainerArtifactValues(), TestValidateImageReference(), TestValidateK8sArtifactValues(), TestValidContainerPort() (+2 more)

### Community 95 - ".checkUserManagerLimits"
Cohesion: 0.22
Nodes (6): delegateSetting(), limitText(), missingControllers(), pamCoreValue(), parseLimit(), parseUnitProps()

### Community 96 - "gen-cert.sh"
Cohesion: 0.27
Nodes (13): ask(), askPass(), askSan(), askYesNo(), buildSanList(), checkExit(), concatCnIfNotBlank(), echoUsage() (+5 more)

### Community 98 - "Container"
Cohesion: 0.19
Nodes (10): Container, DockerConfig, Network, PodmanConfig, Ulimits, assertContainerBlockDefaults(), assertContainerScaling(), TestApplyDefaultsDocker() (+2 more)

### Community 99 - "Configuration"
Cohesion: 0.15
Nodes (13): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+5 more)

### Community 100 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 102 - "DetectPlatforms"
Cohesion: 0.15
Nodes (26): examples_test.go, internal/examples, platform_test.go, Example, parseError(), DetectPlatforms(), PlatformAbbrev(), TestDetectPlatforms() (+18 more)

### Community 103 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.29
Nodes (6): Contract, Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 112 - "capRunner"
Cohesion: 0.11
Nodes (28): capCall, capRunner, internal/container, runtime_test.go, transport_test.go, NewManager(), callIndex(), TestManagerLogsCLIShell() (+20 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.10
Nodes (23): containerState, healthState, inspect_test.go, decodeInspect(), orUnknown(), printable(), inspectReply(), kvValue() (+15 more)

### Community 126 - "operatorversion_test.go"
Cohesion: 0.15
Nodes (19): operatorversion_test.go, operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsNotSkippedByAFailedRead(), TestConfirmNoDowngradeIsSilentInAPreview(), TestConfirmNoDowngradeProceedsWhenConfirmAccepts(), TestConfirmNoDowngradeRefusedWhenConfirmDeclines() (+11 more)

### Community 155 - "strings.Builder"
Cohesion: 0.11
Nodes (25): shorthand, mdRow(), renderAbbrevDocs(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable(), writeUnabbreviated() (+17 more)

## Knowledge Gaps
- **13 isolated node(s):** `solace`, `NO_COLOR`, `bash/000-env.sh`, `bash/010-deploy-operator.sh`, `bash/020-deploy-broker.sh` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 191 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `App` to `Injectable seams`, `watchLoop`, `Ops`, `cli_test.go`, `github.com/spf13/cobra.ShellCompDirective`, `Manager`, `testing.T`, `exportconfig_test.go`, `newRootCmd`, `context.Context`, `commands.go`, `Platform`, `config_test.go`, `.dump`, `AdminSecret`, `replication_test.go`, `ReplSite`?**
  _High betweenness centrality (0.044) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `newTestOps()` (e.g. with `importops_test.go` and `Injectable seams`) actually correct?**
  _`newTestOps()` has 117 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `NO_COLOR`, `bash/000-env.sh` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.06475247524752475 - nodes in this community are weakly interconnected._
- **Why does `Config` connect `AdminSecret` to `App`, `newTestOps`, `.kubectl`, `kubectlTransport`, `cli_test.go`, `config_test.go`, `Manager`, `K8sConfig`, `Command`, `NewCluster`, `convert_test.go`, `manager_test.go`, `localCfg`, `Load`, `testing.T`, `strings.Builder`, `NewCluster`, `ResolveDomainCerts`, `.operatorRows`, `Compose`, `secrets_test.go`, `.validate`, `config_test.go`, `eqArgs`, `haCfg`, `.dump`, `ReplSite`, `Quadlet`, `Ops`, `HARoles`, `internal/config`, `writeNode`, `EnvPairs`, `adminsecret_test.go`, `context.Context`, `Platform`, `namespace_test.go`, `render_test.go`, `.decodeScalingEntry`, `validPlatformConfig`, `.RedundancyEnabled`, `Container`, `Image`, `DetectPlatforms`, `capRunner`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._
- **Are the 10 inferred relationships involving `Config` (e.g. with `config_test.go` and `execguard_test.go`) actually correct?**
  _`Config` has 10 INFERRED edges - model-reasoned connections that need verification._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.07785602503912363 - nodes in this community are weakly interconnected._
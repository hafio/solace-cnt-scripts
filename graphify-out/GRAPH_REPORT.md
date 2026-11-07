# Graph Report - solace-cnt-scripts  (2026-11-07)

## Corpus Check
- 204 files · ~708,008 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .golden 17, (none) 3, .cli 2)

## Summary
- 4038 nodes · 20748 edges · 122 communities (112 shown, 10 thin omitted)
- Extraction: 72% EXTRACTED · 28% INFERRED · 0% AMBIGUOUS · INFERRED: 5721 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `298f0cef`
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
- NewCluster
- ResolveDomainCerts
- watchLoop
- github.com/spf13/cobra.ShellCompDirective
- .operatorRows
- annotate_test.go
- Role
- Per-package doubles
- inject_test.go
- ctrCfg
- exportconfig_test.go
- io.Writer
- Platform
- checkRow
- config_test.go
- scripts.go
- eqArgs
- config_ops_test.go
- haCfg
- ReplSite
- .dump
- Command reference
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
- The eight rules
- k8s/inspect.go
- What `import-config` applies
- deploy_test.go
- Operations
- tlssecret_test.go
- .RunCLI
- platform_test.go
- EnvPairs
- cli/watch_test.go
- newRootCmd
- Capture
- containerTransport
- replicationops_test.go
- Platform
- limits_test.go
- Developer guide
- VPNRepl
- namespace_test.go
- broker/replication_test.go
- Config
- render_test.go
- ImportPlan
- RuleFor
- redundancy_test.go
- Sink
- gen-cert.sh
- storage_test.go
- Container
- .validate
- .ApplyDefaults
- Image
- DetectPlatforms
- Platform parity: solmq-conn-util and solace-util
- .checkUserManagerLimits
- preflight_test.go
- kubectlTransport
- .ExportConfig
- mergedBlock
- Troubleshooting
- README.md
- examples_test.go
- capRunner
- TestGeneratorPageEmbedsTheCLI
- Abbreviations
- solaceRows
- runGuarded
- SEMP
- .inspectContainer
- Redundancy
- operatorversion_test.go
- strings.Builder

## God Nodes (most connected - your core abstractions)
1. `App` - 219 edges
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
- `omit_test.go` --references--> `Annotate()`  [INFERRED]
  docs/test.md → internal/broker/annotate.go
- `annotate_test.go` --references--> `Capture`  [INFERRED]
  docs/test.md → internal/broker/blocks.go
- `annotate_test.go` --references--> `ParseBlocks()`  [INFERRED]
  docs/test.md → internal/broker/blocks.go
- `importops_test.go` --references--> `ParseBlocks()`  [INFERRED]
  docs/test.md → internal/broker/blocks.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (122 total, 10 thin omitted)

### Community 0 - "App"
Cohesion: 0.08
Nodes (78): App, logArgs(), logSince(), confirmAction(), confirmDelete(), confirmRedeploy(), warn(), diagnosticsOptions() (+70 more)

### Community 1 - "newTestOps"
Cohesion: 0.07
Nodes (71): recOutput, broker_test.go, countContains(), leaderCalls(), newTestOps(), TestCountContains(), TestDiagnostics(), TestDisableDefaultUsers() (+63 more)

### Community 2 - ".kubectl"
Cohesion: 0.04
Nodes (35): covUploadFailsTo, uploadErrTransport, replMate, Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`) (+27 more)

### Community 3 - "Role"
Cohesion: 0.06
Nodes (47): mateChannelFunc, How a short form is resolved, euid_test.go, Ops, New(), Transport, applyAliases(), completionShell() (+39 more)

### Community 4 - "cli_test.go"
Cohesion: 0.06
Nodes (117): leaderRun, cli_test.go, Shared env fixtures, allowRuntime(), capture(), captureStderr(), echoRunner(), fakeBinaryOnPath() (+109 more)

### Community 5 - "config_test.go"
Cohesion: 0.06
Nodes (77): config_test.go, validPlatformConfig(), assertContainerBlockDefaults(), assertContainerScaling(), envTree(), haNodesConfig(), TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes() (+69 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (50): Commands, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker cli, solace-util broker configure (+42 more)

### Community 7 - "Manager"
Cohesion: 0.09
Nodes (8): Manager, orNone(), platformTitle(), secretSummary(), setOrMissing(), runUserIDs(), Output(), ContainerSecret

### Community 8 - "K8sConfig"
Cohesion: 0.11
Nodes (15): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, LoadBalancer, cut() (+7 more)

### Community 9 - "completion_test.go"
Cohesion: 0.22
Nodes (23): completion_test.go, driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion() (+15 more)

### Community 10 - "Command"
Cohesion: 0.07
Nodes (35): cmdField, guardedCmd, command_test.go, duration_test.go, execguard_test.go, internal/config, decodeCommand(), TestCommandArgsDoesNotAliasCommand() (+27 more)

### Community 11 - "output_test.go"
Cohesion: 0.23
Nodes (18): internal/output, output_test.go, Level, sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLevelOrderKeepsSkipBelowFail() (+10 more)

### Community 12 - "NewCluster"
Cohesion: 0.10
Nodes (65): operator_test.go, GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator(), splitAfterNamespace() (+57 more)

### Community 13 - "k8s/inspect.go"
Cohesion: 0.21
Nodes (21): inspect_test.go, normalizeToList(), podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape() (+13 more)

### Community 14 - "convert_test.go"
Cohesion: 0.07
Nodes (68): doc, Result, segment, vars, convert_test.go, boolOf(), commentSafe(), Convert() (+60 more)

### Community 15 - "transform_test.go"
Cohesion: 0.10
Nodes (62): Block, TargetState, transform_test.go, importIgnore(), injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs() (+54 more)

### Community 16 - "manager_test.go"
Cohesion: 0.08
Nodes (71): manager_test.go, GuardPodmanEUID(), fileExists(), NewManager(), assertMode(), callIndex(), countCalls(), listArgs() (+63 more)

### Community 17 - "commands.go"
Cohesion: 0.14
Nodes (71): opFunc, roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd() (+63 more)

### Community 18 - "coverage_test.go"
Cohesion: 0.08
Nodes (72): coverage_test.go, internal/broker, verify_ops_test.go, matchCLI(), ranContains(), writeFile(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), TestDiagnosticsMkdirError() (+64 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "localCfg"
Cohesion: 0.07
Nodes (58): backupTarget, Credential, semp_test.go, sempmate_test.go, anyHTTP2xx(), bridgeHostPort(), curlConfigLine(), httpBody() (+50 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Load"
Cohesion: 0.07
Nodes (47): scalingTier, Node roles, Reading this reference, scaling_test.go, minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError() (+39 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.15
Nodes (44): PhaseKind, SiteState, SwitchAction, SwitchPhase, switchplan_test.go, BuildSwitchPlan(), confirmDemotions(), ExecuteSwitchPlan() (+36 more)

### Community 24 - "newLocalOps"
Cohesion: 0.16
Nodes (40): verify_local_test.go, cliScriptPath(), isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn(), TestDefaultLocalAddrs() (+32 more)

### Community 25 - "diff_test.go"
Cohesion: 0.16
Nodes (29): BlockDiff, DiffResult, diff_test.go, blockLabel(), DiffBlocks(), diffRow(), keySet(), normalizeLine() (+21 more)

### Community 26 - "renderDriver"
Cohesion: 0.15
Nodes (22): chunk, chunkResult, driver_test.go, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce() (+14 more)

### Community 27 - "hostpath_test.go"
Cohesion: 0.08
Nodes (33): hostpath_test.go, guardConfig(), BaseName(), CheckHostPath(), expandTilde(), expandTildeToken(), HasPathSeparator(), IsAbsHostPath() (+25 more)

### Community 28 - "importops_test.go"
Cohesion: 0.08
Nodes (55): importops_test.go, describeScope(), planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt() (+47 more)

### Community 29 - "watch_test.go"
Cohesion: 0.15
Nodes (32): watch_test.go, subtractWatch(), deniedCluster(), deniedRunner(), deployJSON(), namedOperatorJSON(), operatorItem(), TestAmbiguousOperatorAbortsTheDeploy() (+24 more)

### Community 31 - "statusreport_test.go"
Cohesion: 0.16
Nodes (29): statusreport_test.go, ownsPod(), reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportOmitsWorkloadsThatAreNotTheBrokers(), TestBrokerReportOrdersPodsOfTheSameRoleByName(), TestBrokerReportPreviewStopsAtAnUnapprovedCommand(), TestBrokerReportReadsEachKindOnce() (+21 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.25
Nodes (5): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh

### Community 33 - "NewCluster"
Cohesion: 0.12
Nodes (41): checkreport_test.go, preflight_test.go, TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestCheckReportSkipsEmptySections(), TestOperatorInstallVerdict(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestOperatorValidateAsksNothingAboutTheBroker(), TestPortRowsNeverFail() (+33 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (54): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domaincerts_resolve_test.go, domaincerts_test.go, caNameSafe() (+46 more)

### Community 35 - "watchLoop"
Cohesion: 0.32
Nodes (5): watchInterval(), stdoutIsTerminal(), watchLabel(), watchLoop(), withWatch()

### Community 36 - "github.com/spf13/cobra.ShellCompDirective"
Cohesion: 0.33
Nodes (6): completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), isEnvFileName(), matching()

### Community 37 - ".operatorRows"
Cohesion: 0.14
Nodes (17): containsString(), installSteps(), compareVersions(), imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), operatorRunningImage() (+9 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.11
Nodes (44): annotate_test.go, Annotate(), applyMetaField(), isMarker(), markerSection(), markerVerb(), parseKV(), readMarker() (+36 more)

### Community 40 - "Role"
Cohesion: 0.10
Nodes (12): downloadErrTransport, fakeTransport, recDownload, recRun, recUpload, recUploadFile, runErrMatchTransport, runErrTransport (+4 more)

### Community 41 - "Per-package doubles"
Cohesion: 0.12
Nodes (42): capCall, capRunner, Per-package doubles, replShowMate(), replVPNTable(), answer(), composeUp(), composeUpArgs() (+34 more)

### Community 42 - "inject_test.go"
Cohesion: 0.17
Nodes (28): lineRole, serviceLine, shutdownStyle, svcKey, inject_test.go, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "ctrCfg"
Cohesion: 0.12
Nodes (50): inspect_test.go, secrets_test.go, inspectReply(), kvValue(), TestInspectContainerArgv(), TestInspectContainerRejectsAnUnusableReply(), TestInspectContainerSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment() (+42 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.14
Nodes (28): exportconfigReadCounter, opRunner, exportconfig_test.go, internal/cli, pluralVPN(), exportconfigDriverOutput(), exportconfigTransportOutput(), exportconfigVPNListOutput() (+20 more)

### Community 46 - "io.Writer"
Cohesion: 0.12
Nodes (15): interactiveFailRunner, runFailRunner, Repository layout, internal/engine, runner_test.go, noEnv, Interactive(), IsEcho() (+7 more)

### Community 47 - "Platform"
Cohesion: 0.09
Nodes (38): platform_test.go, checkFlagPlatforms(), commandPlatforms(), declaredList(), parsePlatformList(), prepare(), promptPlatform(), resolvePlatform() (+30 more)

### Community 48 - "checkRow"
Cohesion: 0.20
Nodes (18): orNone(), orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminPassState(), adminSecretRow() (+10 more)

### Community 49 - "config_test.go"
Cohesion: 0.14
Nodes (29): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, span, keyValueEntries, scalingKey, scalingSpelling (+21 more)

### Community 50 - "scripts.go"
Cohesion: 0.10
Nodes (40): showCmd, scripts_test.go, archiveConfigsScript(), assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript() (+32 more)

### Community 51 - "eqArgs"
Cohesion: 0.09
Nodes (46): check_test.go, cluster_test.go, internal/k8s, names_test.go, ops_test.go, transport_test.go, TestCheckStopsProbingWhenUnreachable(), TestReachable() (+38 more)

### Community 52 - "config_ops_test.go"
Cohesion: 0.21
Nodes (12): config_ops_test.go, TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(), TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator(), TestExecShellScriptRefusesAnUnsafeNameBeforeUploading() (+4 more)

### Community 53 - "haCfg"
Cohesion: 0.09
Nodes (56): adminsecret_test.go, prep_test.go, AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestReadAdminPasswordArgvAndDecoding() (+48 more)

### Community 54 - "ReplSite"
Cohesion: 0.09
Nodes (33): ReplRole, sortedKeys(), sortedSet(), ReplicationConfigResult, isRunCLIRejection(), mateConvergenceShutdowns(), missingListedVPNs(), newlineIf() (+25 more)

### Community 55 - ".dump"
Cohesion: 0.11
Nodes (64): opCall, ops_k8s_test.go, captureStdout(), ctrRedeployApp(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook() (+56 more)

### Community 56 - "Command reference"
Cohesion: 0.50
Nodes (4): Command reference, Global flags, Index, Tree

### Community 57 - "replication_test.go"
Cohesion: 0.11
Nodes (43): replRig, solace-util broker, Configuring a site, Data replication, Reaching the mate, Switching roles, When a run fails partway, replication_test.go (+35 more)

### Community 58 - ".OperatorApply"
Cohesion: 0.33
Nodes (4): operatorEscalationProbes(), canIAnswer(), canIRefused(), probe

### Community 59 - "replCfg"
Cohesion: 0.19
Nodes (28): matechannel_test.go, CLIArg(), CLIScriptPath(), execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf() (+20 more)

### Community 60 - "MateChannel"
Cohesion: 0.13
Nodes (31): CLIRunner, fakeRun, matechannel_test.go, MateChannel, NewCLIMate(), mateReplies(), newTestMate(), TestLocalMateRunsThroughOps() (+23 more)

### Community 61 - "replConfig"
Cohesion: 0.18
Nodes (15): Replication, secretRef, replication_test.go, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded() (+7 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.18
Nodes (42): rootless_test.go, failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+34 more)

### Community 63 - "Test catalogue"
Cohesion: 0.25
Nodes (8): Fixtures and doubles, internal/convert, internal/examples, internal/render, internal/tools/vulnjudge, Running the tests, Summary, Test catalogue

### Community 64 - "Injectable seams"
Cohesion: 0.11
Nodes (18): layer, Injectable seams, resolve_test.go, defaultLocalAddrs(), confirmActionStrict(), confirmDowngrade(), confirmLayer(), confirmRestart() (+10 more)

### Community 65 - "secrets_test.go"
Cohesion: 0.12
Nodes (33): The env-file generator page, secrets_test.go, tlsSecretDigest(), GenSecrets(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret() (+25 more)

### Community 66 - "blocks.go"
Cohesion: 0.13
Nodes (32): Region, token, blocks_test.go, markerRegion(), brokerTypeFromSchema(), firstQuoted(), isIndented(), isPragma() (+24 more)

### Community 67 - "step"
Cohesion: 0.14
Nodes (13): step(), newConvertCmd(), runConvert(), newExamplesCmd(), runExample(), addOutFlags(), checkAllowCommand(), emitOrWrite() (+5 more)

### Community 68 - "The eight rules"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 69 - "k8s/inspect.go"
Cohesion: 0.16
Nodes (15): age(), anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), roleRank(), ownedPods(), brokerList (+7 more)

### Community 70 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 71 - "deploy_test.go"
Cohesion: 0.16
Nodes (26): deploy_test.go, brokerJSON(), logBuf(), TestConfirmBrokerAppliedAcceptsTheObjectItApplied(), TestConfirmBrokerAppliedFailsOnADifferentObject(), TestConfirmBrokerAppliedFailsWhenTheObjectIsNotThere(), TestConfirmBrokerAppliedSurfacesReadFailures(), TestConfirmBrokerAppliedTreatsSilenceAsSkipped() (+18 more)

### Community 72 - "Operations"
Cohesion: 0.07
Nodes (27): Bringing up a fresh cluster, Broker scope is a fixed classification, Docker and Podman mechanics, Every destructive command confirms, Exit codes, Exporting and importing configuration, Extra CLI users differ by platform, Handling the artifact (+19 more)

### Community 73 - "tlssecret_test.go"
Cohesion: 0.43
Nodes (8): tlssecret_test.go, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsDeriveASecretName(), TestSuppliedCertsMakeTheSecretOurs(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), TestTLSServerSecretNameFollowsTheStates(), tlsCfg()

### Community 74 - ".RunCLI"
Cohesion: 0.09
Nodes (16): cliRunNames(), field(), runCLISkeleton(), validCLILine(), validName(), ValidScriptName(), rejectionIn(), noReleaseActivityScript() (+8 more)

### Community 75 - "platform_test.go"
Cohesion: 0.28
Nodes (17): platform_test.go, runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection() (+9 more)

### Community 76 - "EnvPairs"
Cohesion: 0.10
Nodes (31): servercert_test.go, ContainerNoFile(), HealthCheck, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), Compose(), ContainerSecrets(), containerSecretSpecs() (+23 more)

### Community 77 - "cli/watch_test.go"
Cohesion: 0.35
Nodes (11): watch_test.go, stopAfter(), TestOperatorStatusWatches(), TestStatusWithoutWatchRunsOnce(), TestWatchAppendsWhenNotATerminal(), TestWatchEndsQuietlyWhenInterrupted(), TestWatchIntervalIsValidated(), TestWatchKeepsGoingAfterAFailedRefresh() (+3 more)

### Community 78 - "newRootCmd"
Cohesion: 0.07
Nodes (45): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, exit_test.go, TestFlagShorthandsAreConsistent(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive() (+37 more)

### Community 79 - "Capture"
Cohesion: 0.13
Nodes (16): BrokerType, Capture, cliMate, transcript_test.go, parseHeader(), readHeaderField(), bannerType(), checkProvenance() (+8 more)

### Community 81 - "replicationops_test.go"
Cohesion: 0.21
Nodes (25): replicationops_test.go, hasCall(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), TestConfigureReplicationClosingShowFailureStillReportsTheApply(), TestConfigureReplicationClosingVPNReadFailureStillReportsTheApply() (+17 more)

### Community 82 - "Platform"
Cohesion: 0.04
Nodes (34): ReplVia, ReplViaKube, ReplViaSEMP, artifactvalues_test.go, checkTimezone(), portOrRange(), TestCheckTimezone(), TestValidateContainerArtifactValues() (+26 more)

### Community 83 - "limits_test.go"
Cohesion: 0.09
Nodes (51): limits_test.go, main_test.go, limitsMgr(), nrOpenProbe(), procLimits(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessCoreFloorFollowsTheConfiguredLimit() (+43 more)

### Community 84 - "Developer guide"
Cohesion: 0.15
Nodes (14): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Tests, Toolchain pin (+6 more)

### Community 85 - "VPNRepl"
Cohesion: 0.08
Nodes (11): AdminState, QueueState, scriptedMate, sempMate, VPNRepl, schemaTransport(), reenableLine(), showVPNReplicationScript() (+3 more)

### Community 86 - "namespace_test.go"
Cohesion: 0.31
Nodes (11): namespace_test.go, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 87 - "broker/replication_test.go"
Cohesion: 0.08
Nodes (65): colSpan, replication_test.go, ValidVPNName(), validVPNName(), cliTransport(), containsEndpoint(), dashSpans(), flagByte() (+57 more)

### Community 88 - "Config"
Cohesion: 0.10
Nodes (24): TLS, adminsecret_test.go, Coverage, imagepullsecret_test.go, adminCfg(), TestAdminSecretNameFollowsTheStates(), TestAdminSecretRefusesTheOperatorsOwnNames(), TestContainersStillRequireTheAdminPassword() (+16 more)

### Community 89 - "render_test.go"
Cohesion: 0.13
Nodes (47): render_test.go, boolStr(), BrokerCR(), ComposeProject(), assertNoCheckoutPath(), containerArtifacts(), healthCheckFixture(), load() (+39 more)

### Community 90 - "ImportPlan"
Cohesion: 0.19
Nodes (15): ImportResult, PlannedSection, defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList(), renderRegionScript() (+7 more)

### Community 91 - "RuleFor"
Cohesion: 0.13
Nodes (36): Disposition, SectionRule, appliance_test.go, importdoc_test.go, sections_test.go, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent() (+28 more)

### Community 93 - "redundancy_test.go"
Cohesion: 0.21
Nodes (12): redundancy_test.go, pskCfg(), TestFillStandaloneNodeNameIgnoresAnEmptyHostname(), TestFillStandaloneNodeNameKeepsAConfiguredName(), TestFillStandaloneNodeNameNeverFillsHA(), TestFillStandaloneNodeNameUsesTheHost(), TestNothingGeneratesThePSK(), TestPSKEnvSatisfiesTheContainerRequirement() (+4 more)

### Community 95 - "Sink"
Cohesion: 0.24
Nodes (7): columnWidths(), KV, Sink, NewFunc(), pad(), TestLeveledLinesCarryTheTagAndNoArrow(), TestNilSinksDiscard()

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
Cohesion: 0.22
Nodes (7): solace-util, solace-util operator, okRow(), portRows(), TestOperatorInstallVerdictShowsOnceInEveryView(), checkReport, validateScope

### Community 100 - ".ApplyDefaults"
Cohesion: 0.17
Nodes (10): guardCommandOf(), setGuardCommand(), TestValidatorAndExecutorAgree(), applyContainerBlockDefaults(), podmanRunUser(), setDefault(), setDefaultCmd(), setDefaultInt() (+2 more)

### Community 102 - "DetectPlatforms"
Cohesion: 0.22
Nodes (18): examples_test.go, Example, TestExamplesEmitsToStdout(), All(), Get(), List(), Names(), read() (+10 more)

### Community 103 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.29
Nodes (6): Contract, Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 104 - ".checkUserManagerLimits"
Cohesion: 0.19
Nodes (4): delegateSetting(), missingControllers(), pamCoreValue(), parseUnitProps()

### Community 105 - "preflight_test.go"
Cohesion: 0.27
Nodes (10): preflight_test.go, TestManagerLifecycleDocker(), TestManagerLifecyclePodmanSystemctl(), TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestComposeSecretEnvNamesCannotBeSystemVars(), TestPreflightFailureStopsLifecycle(), TestPreflightFailureStopsTheDeploy(), TestPreflightHintIsPlatformShaped() (+2 more)

### Community 107 - ".ExportConfig"
Cohesion: 0.27
Nodes (8): Omission, omit_test.go, keepRegion(), omitAtExport(), omittedList(), TestOmitAtExportDropsFilteredLinesAndEmptyBlocks(), TestOmitAtExportLeavesTargetDecisionsToImport(), TestOmitAtExportRemovesSkipSectionsAndRecordsThem()

### Community 108 - "mergedBlock"
Cohesion: 0.28
Nodes (8): blockKey, mergedBlock, mergedLine, ancestorReported(), indexByIdentity(), keyOf(), mergeByIdentity(), missingUnder()

### Community 109 - "Troubleshooting"
Cohesion: 0.22
Nodes (9): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Forbidden: cannot list deployments at the cluster scope, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting (+1 more)

### Community 110 - "README.md"
Cohesion: 0.22
Nodes (6): internal/broker, internal/cli, internal/config, internal/engine, internal/k8s, internal/render

### Community 111 - "examples_test.go"
Cohesion: 0.29
Nodes (7): examples_test.go, TestExamplesAcceptsAPlatformAbbreviation(), TestExamplesBareEmitsTheFullSchema(), TestExamplesNeedsNoEnvFile(), TestExamplesRefusesToOverwrite(), TestExamplesRejectsAnUnknownPlatform(), TestExamplesWritesOutFile()

### Community 112 - "capRunner"
Cohesion: 0.14
Nodes (25): internal/container, runtime_test.go, transport_test.go, TestManagerLogsCLIShell(), TestManagerLogsRefusesPrevious(), mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged() (+17 more)

### Community 113 - "TestGeneratorPageEmbedsTheCLI"
Cohesion: 0.40
Nodes (5): generatorpage_test.go, gunzipBase64(), gzipBase64(), templateActions(), TestGeneratorPageEmbedsTheCLI()

### Community 114 - "Abbreviations"
Cohesion: 0.50
Nodes (4): Abbreviations, Commands, Flag shorthands, Platforms

### Community 115 - "solaceRows"
Cohesion: 0.67
Nodes (3): printableRow(), solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv()

### Community 116 - "runGuarded"
Cohesion: 0.67
Nodes (4): runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand(), writeRootlessPodmanEnv()

### Community 118 - ".inspectContainer"
Cohesion: 0.24
Nodes (7): health(), orUnknown(), decodeOne(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings(), orValue()

### Community 126 - "operatorversion_test.go"
Cohesion: 0.16
Nodes (19): operatorversion_test.go, operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsNotSkippedByAFailedRead(), TestConfirmNoDowngradeIsSilentInAPreview(), TestConfirmNoDowngradeProceedsWhenConfirmAccepts(), TestConfirmNoDowngradeRefusedWhenConfirmDeclines() (+11 more)

### Community 155 - "strings.Builder"
Cohesion: 0.08
Nodes (36): shorthand, WeightedNodeTerm, mdRow(), renderAbbrevDocs(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable() (+28 more)

## Knowledge Gaps
- **13 isolated node(s):** `solace`, `NO_COLOR`, `bash/000-env.sh`, `bash/010-deploy-operator.sh`, `bash/020-deploy-broker.sh` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 191 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **10 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `App` to `Injectable seams`, `.kubectl`, `watchLoop`, `step`, `Role`, `cli_test.go`, `github.com/spf13/cobra.ShellCompDirective`, `Per-package doubles`, `platform_test.go`, `exportconfig_test.go`, `io.Writer`, `Platform`, `newRootCmd`, `commands.go`, `config_test.go`, `.dump`, `Config`, `replication_test.go`?**
  _High betweenness centrality (0.039) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `newTestOps()` (e.g. with `importops_test.go` and `Injectable seams`) actually correct?**
  _`newTestOps()` has 117 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `NO_COLOR`, `bash/000-env.sh` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.07992007992007992 - nodes in this community are weakly interconnected._
- **Why does `Config` connect `Config` to `App`, `newTestOps`, `.kubectl`, `Role`, `cli_test.go`, `config_test.go`, `Manager`, `K8sConfig`, `Command`, `NewCluster`, `convert_test.go`, `manager_test.go`, `localCfg`, `Load`, `hostpath_test.go`, `strings.Builder`, `NewCluster`, `ResolveDomainCerts`, `.operatorRows`, `Per-package doubles`, `ctrCfg`, `Platform`, `checkRow`, `config_test.go`, `eqArgs`, `haCfg`, `.dump`, `replCfg`, `replConfig`, `Test catalogue`, `secrets_test.go`, `tlssecret_test.go`, `EnvPairs`, `containerTransport`, `Platform`, `namespace_test.go`, `render_test.go`, `redundancy_test.go`, `storage_test.go`, `Container`, `.ApplyDefaults`, `Image`, `kubectlTransport`, `capRunner`, `SEMP`, `Redundancy`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **Are the 11 inferred relationships involving `Config` (e.g. with `config_test.go` and `execguard_test.go`) actually correct?**
  _`Config` has 11 INFERRED edges - model-reasoned connections that need verification._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.0747871158830063 - nodes in this community are weakly interconnected._
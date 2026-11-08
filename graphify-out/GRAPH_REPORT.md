# Graph Report - solace-cnt-scripts  (2026-11-08)

## Corpus Check
- 202 files · ~711,244 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .golden 17, (none) 3, .cli 2)

## Summary
- 4050 nodes · 20873 edges · 109 communities (103 shown, 6 thin omitted)
- Extraction: 72% EXTRACTED · 28% INFERRED · 0% AMBIGUOUS · INFERRED: 5759 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `41850d42`
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
- Ops
- NewCluster
- main_test.go
- convert_test.go
- transform_test.go
- manager_test.go
- commands.go
- coverage_test.go
- Bash Dev Script
- verify_local_test.go
- PowerShell Dev Script
- Abbreviations
- switchplan_test.go
- internal/broker
- Sink
- .RunCLI
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
- .rpc
- annotate_test.go
- Load
- manager_test.go
- inject_test.go
- ctrCfg
- exportconfig_test.go
- io.Writer
- replication_test.go
- podName
- config_test.go
- scripts_test.go
- github.com/hafio/go-deploy-utility/runner.Runner
- config_ops_test.go
- haCfg
- VPNRepl
- .dump
- The eight rules
- replication_test.go
- .OperatorApply
- replCfg
- parseBody
- replConfig
- rootless_test.go
- Platform parity: solmq-conn-util and solace-util
- .prompter
- secrets_test.go
- blocks.go
- step
- internal/k8s
- k8s/inspect.go
- .validate
- deploy_test.go
- Operations
- .validateK8s
- What `import-config` applies
- imagepullsecret_test.go
- EnvPairs
- runRootWith
- platform_test.go
- blocks.go
- preflight_test.go
- replicationops_test.go
- Platform
- limits_test.go
- Goldens
- context.Context
- namespace_test.go
- broker/replication_test.go
- Config
- render_test.go
- README.md
- RuleFor
- NewExec
- Exporting and importing configuration
- gen-cert.sh
- euid_test.go
- Container
- Test catalogue
- Removing a broker: what stays, what goes
- Platforms
- newRootCmd
- internal/config
- capRunner
- solaceRows
- .inspectContainer
- operatorversion_test.go
- strings.Builder

## God Nodes (most connected - your core abstractions)
1. `App` - 217 edges
2. `newTestOps()` - 167 edges
3. `Config` - 151 edges
4. `cli_test.go` - 151 edges
5. `NewCluster()` - 146 edges
6. `ctrCfg()` - 145 edges
7. `Per-package doubles` - 134 edges
8. `Role` - 132 edges
9. `newCapMgr()` - 126 edges
10. `loadK8s()` - 121 edges

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

## Communities (109 total, 6 thin omitted)

### Community 0 - "App"
Cohesion: 0.08
Nodes (65): App, logOptions(), logSince(), confirmDelete(), runExport(), diagnosticsOptions(), resolveScriptPath(), ctrLogin() (+57 more)

### Community 1 - "newTestOps"
Cohesion: 0.08
Nodes (67): broker_test.go, countContains(), newTestOps(), TestCountContains(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultUsersRefusesAQuotedVPNName() (+59 more)

### Community 2 - ".kubectl"
Cohesion: 0.05
Nodes (25): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+17 more)

### Community 3 - "Role"
Cohesion: 0.06
Nodes (25): covUploadFailsTo, downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile (+17 more)

### Community 4 - "cli_test.go"
Cohesion: 0.07
Nodes (76): leaderRun, cli_test.go, examples_test.go, Shared env fixtures, allowRuntime(), capture(), captureStderr(), collectPaths() (+68 more)

### Community 5 - "config_test.go"
Cohesion: 0.06
Nodes (97): config_test.go, redundancy_test.go, TestContainersStillRequireTheAdminPassword(), TestOperatorAdminSecretName(), TestPSKConfiguredCountsBothSpellings(), TestValidateK8sArtifactValues(), assertContainerBlockDefaults(), assertContainerScaling() (+89 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (50): Commands, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker cli, solace-util broker configure (+42 more)

### Community 7 - "Manager"
Cohesion: 0.09
Nodes (8): Manager, orNone(), orValue(), platformTitle(), secretSummary(), setOrMissing(), TestManagerHonoursRuntime(), ContainerSecret

### Community 8 - "K8sConfig"
Cohesion: 0.11
Nodes (15): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, LoadBalancer, cut() (+7 more)

### Community 9 - "completion_test.go"
Cohesion: 0.21
Nodes (24): completion_test.go, driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion() (+16 more)

### Community 10 - "Command"
Cohesion: 0.07
Nodes (37): cmdField, guardedCmd, command_test.go, execguard_test.go, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandArgv(), TestCommandNameAndArgs() (+29 more)

### Community 11 - "Ops"
Cohesion: 0.09
Nodes (47): mateChannelFunc, Injectable seams, Ops, New(), confirmAction(), confirmRestart(), lineSink(), warn() (+39 more)

### Community 12 - "NewCluster"
Cohesion: 0.10
Nodes (65): operator_test.go, GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), RenderOperator(), splitAfterNamespace(), splitOperatorBundle() (+57 more)

### Community 13 - "main_test.go"
Cohesion: 0.19
Nodes (20): internal/tools/vulnjudge, main_test.go, describe(), judge(), main(), plural(), run(), load() (+12 more)

### Community 14 - "convert_test.go"
Cohesion: 0.07
Nodes (68): doc, Result, segment, vars, convert_test.go, boolOf(), commentSafe(), Convert() (+60 more)

### Community 15 - "transform_test.go"
Cohesion: 0.10
Nodes (63): Block, TargetState, transform_test.go, importIgnore(), injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs() (+55 more)

### Community 16 - "manager_test.go"
Cohesion: 0.08
Nodes (63): inspect_test.go, manager_test.go, GuardPodmanEUID(), inspectReply(), kvValue(), TestInspectContainerArgv(), TestInspectContainerRejectsAnUnusableReply(), TestInspectContainerSurfacesTheEngineError() (+55 more)

### Community 17 - "commands.go"
Cohesion: 0.13
Nodes (72): opFunc, roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd() (+64 more)

### Community 18 - "coverage_test.go"
Cohesion: 0.08
Nodes (56): coverage_test.go, matchCLI(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultUsersShowVPNError() (+48 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "verify_local_test.go"
Cohesion: 0.07
Nodes (86): semp_test.go, sempmate_test.go, verify_local_test.go, bridgeHostPort(), sempPort(), sempV1OK(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly() (+78 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Abbreviations"
Cohesion: 0.18
Nodes (10): Abbreviations, Flag shorthands, How a short form is resolved, Node roles, Platforms, Command reference, Global flags, Index (+2 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.06
Nodes (77): CLIRunner, fakeRun, PhaseKind, SiteState, SwitchAction, SwitchPhase, matechannel_test.go, switchplan_test.go (+69 more)

### Community 24 - "internal/broker"
Cohesion: 0.18
Nodes (14): internal/broker, verify_ops_test.go, ranContains(), TestRemoveCLIWarnsOnFailure(), cliScriptPath(), TestDiagnosticsFetchesAnEncryptedBundle(), TestDiagnosticsGathersEveryNodeDespiteAFailure(), TestGatherNodeCleanupFailureOnlyWarns() (+6 more)

### Community 25 - "Sink"
Cohesion: 0.07
Nodes (61): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, diff_test.go, internal/output, output_test.go (+53 more)

### Community 26 - ".RunCLI"
Cohesion: 0.11
Nodes (13): cliRunNames(), field(), runCLISkeleton(), validCLILine(), validName(), ValidScriptName(), rejectionIn(), removeServerCertScript() (+5 more)

### Community 27 - "hostpath_test.go"
Cohesion: 0.07
Nodes (34): hostpath_test.go, guardConfig(), BaseName(), CheckHostPath(), expandTilde(), expandTildeToken(), HasPathSeparator(), IsAbsHostPath() (+26 more)

### Community 28 - "importops_test.go"
Cohesion: 0.06
Nodes (73): ImportResult, PlannedSection, importops_test.go, isPragma(), defaultVPNFirst(), describeScope(), ImportPlan, orNone() (+65 more)

### Community 29 - "watch_test.go"
Cohesion: 0.10
Nodes (39): watch_test.go, desiredWatch(), splitWatch(), subtractWatch(), deniedCluster(), deniedRunner(), deployJSON(), namedOperatorJSON() (+31 more)

### Community 31 - "statusreport_test.go"
Cohesion: 0.17
Nodes (28): statusreport_test.go, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportOmitsWorkloadsThatAreNotTheBrokers(), TestBrokerReportOrdersPodsOfTheSameRoleByName(), TestBrokerReportPreviewStopsAtAnUnapprovedCommand(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportReadsOnlyWhatTheEnvFileNames() (+20 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.25
Nodes (5): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh

### Community 33 - "NewCluster"
Cohesion: 0.12
Nodes (40): checkreport_test.go, preflight_test.go, TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestCheckReportSkipsEmptySections(), TestOperatorInstallVerdict(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestOperatorValidateAsksNothingAboutTheBroker(), TestPortRowsNeverFail() (+32 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (50): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domaincerts_resolve_test.go, domaincerts_test.go, caNameSafe() (+42 more)

### Community 35 - "watchLoop"
Cohesion: 0.36
Nodes (4): watchInterval(), watchLabel(), watchLoop(), withWatch()

### Community 36 - "github.com/spf13/cobra.ShellCompDirective"
Cohesion: 0.33
Nodes (6): completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), isEnvFileName(), matching()

### Community 37 - ".rpc"
Cohesion: 0.11
Nodes (15): backupTarget, Credential, sempMate, schemaTransport(), anyHTTP2xx(), curlConfigFlag(), curlConfigLine(), httpBody() (+7 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.12
Nodes (45): annotate_test.go, Annotate(), applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV() (+37 more)

### Community 40 - "Load"
Cohesion: 0.09
Nodes (41): scalingTier, scaling_test.go, minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs() (+33 more)

### Community 41 - "manager_test.go"
Cohesion: 0.12
Nodes (60): capCall, capRunner, Per-package doubles, answer(), assertMode(), callIndex(), composeUp(), composeUpArgs() (+52 more)

### Community 42 - "inject_test.go"
Cohesion: 0.17
Nodes (28): lineRole, serviceLine, shutdownStyle, svcKey, inject_test.go, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "ctrCfg"
Cohesion: 0.12
Nodes (41): secrets_test.go, NewManager(), ctrCfg(), newEchoMgr(), TestCheckEnvReportsBaseDirOnlyWhenSet(), TestManagerCheckDNSFailsLoudInHA(), TestManagerCheckDryRun(), TestManagerCheckReachableError() (+33 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.15
Nodes (29): exportconfigReadCounter, opRunner, exportconfig_test.go, internal/cli, confirmImport(), pluralVPN(), exportconfigDriverOutput(), exportconfigTransportOutput() (+21 more)

### Community 46 - "io.Writer"
Cohesion: 0.12
Nodes (8): interactiveFailRunner, runFailRunner, noEnv, legacyMethod(), refuseEnv(), legacyMethod(), writeOut(), recRunner

### Community 47 - "replication_test.go"
Cohesion: 0.19
Nodes (23): replication_test.go, ReplicationConfigResult, confirmReplicationConfig(), mateChannel(), mateSEMPPassword(), reportReplicationConfig(), replApp(), siteNamed() (+15 more)

### Community 48 - "podName"
Cohesion: 0.12
Nodes (12): BrokerPodSuffixShape(), allCustomMounted(), discoveryServiceName(), HARoles(), lbServiceName(), podName(), pvcName(), RestartOrder() (+4 more)

### Community 49 - "config_test.go"
Cohesion: 0.15
Nodes (29): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, span, AdditionalUser, Node, Redundancy (+21 more)

### Community 50 - "scripts_test.go"
Cohesion: 0.09
Nodes (42): showCmd, scripts_test.go, archiveConfigsScript(), assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript() (+34 more)

### Community 51 - "github.com/hafio/go-deploy-utility/runner.Runner"
Cohesion: 0.17
Nodes (17): Repository layout, runner_test.go, CheckSecretKeyRef(), validDNSLabel(), validDNSSubdomain(), validSecretKey(), Interactive(), IsEcho() (+9 more)

### Community 52 - "config_ops_test.go"
Cohesion: 0.17
Nodes (14): config_ops_test.go, TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(), TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator() (+6 more)

### Community 53 - "haCfg"
Cohesion: 0.06
Nodes (89): cluster_test.go, names_test.go, ops_test.go, prep_test.go, opK8sGenBroker(), TestAdminPasswordPreflight(), TestReachable(), newCluster() (+81 more)

### Community 54 - "VPNRepl"
Cohesion: 0.11
Nodes (10): AdminState, QueueState, replMate, ReplRole, VPNRepl, reenableLine(), setReplicationRoleScript(), sempAdminState() (+2 more)

### Community 55 - ".dump"
Cohesion: 0.09
Nodes (70): opCall, ops_k8s_test.go, captureStdout(), ctrRedeployApp(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook() (+62 more)

### Community 56 - "The eight rules"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 57 - "replication_test.go"
Cohesion: 0.15
Nodes (24): replRig, Configuring a site, Data replication, Reaching the mate, Switching roles, When a run fails partway, replNewRig(), replUp() (+16 more)

### Community 59 - "replCfg"
Cohesion: 0.20
Nodes (27): matechannel_test.go, CLIArg(), CLIScriptPath(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf() (+19 more)

### Community 60 - "parseBody"
Cohesion: 0.14
Nodes (15): Omission, Region, omit_test.go, isIndented(), parseBody(), regionFor(), sectionName(), trimTrailingBlank() (+7 more)

### Community 61 - "replConfig"
Cohesion: 0.31
Nodes (12): Replication, replication_test.go, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationEmptyViaAccepted() (+4 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.18
Nodes (42): rootless_test.go, failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+34 more)

### Community 63 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.29
Nodes (6): Contract, Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 64 - ".prompter"
Cohesion: 0.27
Nodes (7): layer, confirmActionStrict(), confirmDowngrade(), confirmLayer(), confirmRedeploy(), confirmScope(), confirmReplicationSwitch()

### Community 65 - "secrets_test.go"
Cohesion: 0.13
Nodes (36): secrets_test.go, GenSecrets(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), operatorRegcred(), readTLSMaterial() (+28 more)

### Community 66 - "blocks.go"
Cohesion: 0.17
Nodes (24): token, blocks_test.go, brokerTypeFromSchema(), firstQuoted(), newBlock(), ParseBlocks(), blockByIndex(), loadSample() (+16 more)

### Community 67 - "step"
Cohesion: 0.16
Nodes (11): step(), newConvertCmd(), runConvert(), newExamplesCmd(), runExample(), addOutFlags(), checkAllowCommand(), emitOrWrite() (+3 more)

### Community 68 - "internal/k8s"
Cohesion: 0.13
Nodes (21): The env-file generator page, adminsecret_test.go, check_test.go, generatorpage_test.go, internal/k8s, transport_test.go, AdminSecretInUse(), ReadAdminPassword() (+13 more)

### Community 69 - "k8s/inspect.go"
Cohesion: 0.08
Nodes (44): inspect_test.go, orNone(), orValue(), age(), anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition() (+36 more)

### Community 70 - ".validate"
Cohesion: 0.16
Nodes (21): setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminSecretRow(), containsString(), failRow(), info() (+13 more)

### Community 71 - "deploy_test.go"
Cohesion: 0.17
Nodes (25): deploy_test.go, brokerJSON(), logBuf(), TestConfirmBrokerAppliedAcceptsTheObjectItApplied(), TestConfirmBrokerAppliedFailsOnADifferentObject(), TestConfirmBrokerAppliedFailsWhenTheObjectIsNotThere(), TestConfirmBrokerAppliedSurfacesReadFailures(), TestConfirmBrokerAppliedTreatsSilenceAsSkipped() (+17 more)

### Community 72 - "Operations"
Cohesion: 0.07
Nodes (29): Commands, solace-util, solace-util broker, solace-util operator, Secrets, A replication switch refuses before changing anything, A wrapper runtime is refused, Bringing up a fresh cluster (+21 more)

### Community 73 - ".validateK8s"
Cohesion: 0.14
Nodes (7): keyValueEntries, missingErr(), requireAll(), requireKeyValue(), validK8sID(), validPortName(), validPortNumber()

### Community 74 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 75 - "imagepullsecret_test.go"
Cohesion: 0.23
Nodes (10): secretRef, imagepullsecret_test.go, pullSecretCfg(), TestImagePullSecretDerivedSuffixFitsWhereAdditionalUsersAlreadyFits(), TestImagePullSecretNameBringYourOwn(), TestImagePullSecretNameConfiguredNameWins(), TestImagePullSecretNameDerivedWhenCredentialsButNoName(), TestImagePullSecretNameEmptyWithNeither() (+2 more)

### Community 76 - "EnvPairs"
Cohesion: 0.10
Nodes (31): servercert_test.go, ContainerNoFile(), HealthCheck, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), Compose(), ContainerSecrets(), containerSecretSpecs() (+23 more)

### Community 77 - "runRootWith"
Cohesion: 0.09
Nodes (55): watch_test.go, echoRunner(), legacyMethod(), runCtr(), runRootWith(), TestAssertLeaderStandaloneSkipsBeforeAnyPrompt(), TestConfigStepsDoNotLeakSecrets(), TestConfiguredRouternameSurvivesTheFallback() (+47 more)

### Community 78 - "platform_test.go"
Cohesion: 0.22
Nodes (21): platform_test.go, supported(), runPlatform(), TestLogOptionsBuildsOneSetForBothPlatforms(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused() (+13 more)

### Community 79 - "blocks.go"
Cohesion: 0.13
Nodes (15): BrokerType, Capture, cliMate, transcript_test.go, parseHeader(), readHeaderField(), bannerType(), checkProvenance() (+7 more)

### Community 80 - "preflight_test.go"
Cohesion: 0.40
Nodes (7): preflight_test.go, TestManagerLifecycleDocker(), TestManagerLifecyclePodmanSystemctl(), TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestComposeSecretEnvNamesCannotBeSystemVars(), TestPreflightFailureStopsLifecycle(), TestPreflightRunsBeforeAnything()

### Community 81 - "replicationops_test.go"
Cohesion: 0.12
Nodes (49): replicationops_test.go, hasCall(), isRunCLIRejection(), mateConvergenceShutdowns(), missingListedVPNs(), newlineIf(), PlannedRoles(), replicationVPNLines() (+41 more)

### Community 82 - "Platform"
Cohesion: 0.05
Nodes (45): artifactvalues_test.go, checkFlagPlatforms(), commandPlatforms(), declaredList(), parsePlatformList(), prepare(), promptPlatform(), resolvePlatform() (+37 more)

### Community 83 - "limits_test.go"
Cohesion: 0.17
Nodes (32): limits_test.go, limitsMgr(), nrOpenProbe(), procLimits(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessCoreFloorFollowsTheConfiguredLimit(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips() (+24 more)

### Community 84 - "Goldens"
Cohesion: 0.25
Nodes (9): Goldens, abbrevdoc_test.go, commanddoc_test.go, TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), firstDiff(), TestCommandDocs(), TestEveryDocCarriesTheSupportNotice() (+1 more)

### Community 85 - "context.Context"
Cohesion: 0.07
Nodes (8): scriptedMate, containerTransport, delegateSetting(), missingControllers(), pamCoreValue(), parseUnitProps(), runUserIDs(), Output()

### Community 86 - "namespace_test.go"
Cohesion: 0.35
Nodes (14): namespace_test.go, nsItem(), nsList(), occCfg(), occCluster(), TestNamespaceOccupantsAsksOneQuestion(), TestNamespaceOccupantsDiscountsTheBrokerAndWhatItOwns(), TestNamespaceOccupantsDiscountsWhatKubernetesPutsThere() (+6 more)

### Community 87 - "broker/replication_test.go"
Cohesion: 0.06
Nodes (74): colSpan, ReplVia, ReplViaKube, ReplViaSEMP, replication_test.go, ValidVPNName(), validVPNName(), cliTransport() (+66 more)

### Community 88 - "Config"
Cohesion: 0.12
Nodes (9): Image, TLS, atoiPrefix(), Config, DecodeStrict(), clearTierDerived(), validateOutput(), adminPassState() (+1 more)

### Community 89 - "render_test.go"
Cohesion: 0.12
Nodes (49): Coverage, render_test.go, monitorPassState(), boolStr(), BrokerCR(), ComposeProject(), assertNoCheckoutPath(), containerArtifacts() (+41 more)

### Community 90 - "README.md"
Cohesion: 0.22
Nodes (6): internal/broker, internal/cli, internal/config, internal/engine, internal/k8s, internal/render

### Community 91 - "RuleFor"
Cohesion: 0.07
Nodes (58): chunk, chunkResult, Disposition, SectionRule, appliance_test.go, driver_test.go, importdoc_test.go, sections_test.go (+50 more)

### Community 93 - "NewExec"
Cohesion: 0.32
Nodes (7): internal/engine, resolve_test.go, TestChildEnvNamesAreNotSystemVariables(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), NewExec()

### Community 95 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 96 - "gen-cert.sh"
Cohesion: 0.27
Nodes (13): ask(), askPass(), askSan(), askYesNo(), buildSanList(), checkExit(), concatCnIfNotBlank(), echoUsage() (+5 more)

### Community 97 - "euid_test.go"
Cohesion: 0.67
Nodes (6): euid_test.go, runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand(), TestPodmanEUIDGuardSkips(), writeRootlessPodmanEnv()

### Community 98 - "Container"
Cohesion: 0.27
Nodes (6): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestContainerCoreLimitSpellings()

### Community 99 - "Test catalogue"
Cohesion: 0.29
Nodes (7): Fixtures and doubles, internal/convert, internal/examples, internal/render, Running the tests, Summary, Test catalogue

### Community 100 - "Removing a broker: what stays, what goes"
Cohesion: 0.40
Nodes (5): Every destructive command confirms, Removing a broker: what stays, what goes, Removing the operator does not always remove it, The layer flag raises the question, The namespace is only offered when it is empty

### Community 102 - "Platforms"
Cohesion: 0.13
Nodes (29): examples_test.go, platform_test.go, Example, TestExamplesEmitsToStdout(), parseError(), DetectPlatforms(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint() (+21 more)

### Community 103 - "newRootCmd"
Cohesion: 0.08
Nodes (42): aliases_test.go, allowcommand_test.go, exit_test.go, applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb() (+34 more)

### Community 108 - "internal/config"
Cohesion: 0.11
Nodes (28): adminsecret_test.go, duration_test.go, internal/config, storage_test.go, tlssecret_test.go, adminCfg(), TestAdminSecretNameFollowsTheStates(), TestAdminSecretRefusesTheOperatorsOwnNames() (+20 more)

### Community 112 - "capRunner"
Cohesion: 0.16
Nodes (23): internal/container, runtime_test.go, transport_test.go, TestManagerLogsCLIShell(), mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime() (+15 more)

### Community 115 - "solaceRows"
Cohesion: 0.67
Nodes (3): printableRow(), solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv()

### Community 118 - ".inspectContainer"
Cohesion: 0.29
Nodes (6): health(), orUnknown(), decodeOne(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings()

### Community 126 - "operatorversion_test.go"
Cohesion: 0.15
Nodes (22): operatorversion_test.go, compareVersions(), imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), operatorDeployJSON(), TestCompareVersions() (+14 more)

### Community 155 - "strings.Builder"
Cohesion: 0.08
Nodes (39): shorthand, WeightedNodeTerm, mdRow(), renderAbbrevDocs(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable() (+31 more)

## Knowledge Gaps
- **13 isolated node(s):** `solace`, `NO_COLOR`, `bash/000-env.sh`, `bash/010-deploy-operator.sh`, `bash/020-deploy-broker.sh` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 188 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **6 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `App` to `Role`, `Ops`, `commands.go`, `watchLoop`, `github.com/spf13/cobra.ShellCompDirective`, `manager_test.go`, `exportconfig_test.go`, `replication_test.go`, `config_test.go`, `github.com/hafio/go-deploy-utility/runner.Runner`, `haCfg`, `.dump`, `replication_test.go`, `.prompter`, `step`, `runRootWith`, `platform_test.go`, `Platform`, `context.Context`, `Config`, `newRootCmd`?**
  _High betweenness centrality (0.041) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `newTestOps()` (e.g. with `importops_test.go` and `Injectable seams`) actually correct?**
  _`newTestOps()` has 117 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `NO_COLOR`, `bash/000-env.sh` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.08125 - nodes in this community are weakly interconnected._
- **Why does `Config` connect `Config` to `App`, `newTestOps`, `.kubectl`, `cli_test.go`, `config_test.go`, `Manager`, `K8sConfig`, `Command`, `Ops`, `NewCluster`, `convert_test.go`, `manager_test.go`, `verify_local_test.go`, `hostpath_test.go`, `strings.Builder`, `watch_test.go`, `NewCluster`, `ResolveDomainCerts`, `Load`, `manager_test.go`, `ctrCfg`, `podName`, `config_test.go`, `haCfg`, `.dump`, `replCfg`, `replConfig`, `secrets_test.go`, `internal/k8s`, `.validate`, `.validateK8s`, `imagepullsecret_test.go`, `EnvPairs`, `Platform`, `context.Context`, `namespace_test.go`, `render_test.go`, `Container`, `Test catalogue`, `internal/config`, `capRunner`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Are the 11 inferred relationships involving `Config` (e.g. with `config_test.go` and `execguard_test.go`) actually correct?**
  _`Config` has 11 INFERRED edges - model-reasoned connections that need verification._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.08074534161490683 - nodes in this community are weakly interconnected._
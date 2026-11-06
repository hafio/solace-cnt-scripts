# Graph Report - solace-cnt-scripts  (2026-11-06)

## Corpus Check
- 204 files · ~701,668 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .golden 17, (none) 3, .cli 2)

## Summary
- 4011 nodes · 20571 edges · 109 communities (101 shown, 8 thin omitted)
- Extraction: 73% EXTRACTED · 27% INFERRED · 0% AMBIGUOUS · INFERRED: 5655 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `f938a30a`
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
- ContainerSecrets
- inject_test.go
- secrets_test.go
- exportconfig_test.go
- io.Writer
- Platforms
- .validate
- config_test.go
- execguard_test.go
- internal/k8s
- config_ops_test.go
- haCfg
- .runMatePhase
- .dump
- Abbreviations
- replication_test.go
- .OperatorApply
- replCfg
- ReplVia
- replConfig
- rootless_test.go
- Test catalogue
- github.com/hafio/go-deploy-utility/runner.Runner
- Compose
- blocks_test.go
- Ops
- The eight rules
- .podSection
- What `import-config` applies
- Injectable seams
- Operations
- tlssecret_test.go
- .RunCLI
- platform_test.go
- EnvPairs
- runRootWith
- newRootCmd
- captureStdout
- containerTransport
- replicationops_test.go
- Platform
- limits_test.go
- Developer guide
- .BackupRevertActivity
- namespace_test.go
- broker/replication_test.go
- AdminSecret
- render_test.go
- allowcommand_test.go
- RuleFor
- redundancy_test.go
- Goldens
- gen-cert.sh
- storage_test.go
- Container
- Configuration
- runLeader
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
3. `Config` - 149 edges
4. `cli_test.go` - 146 edges
5. `NewCluster()` - 144 edges
6. `ctrCfg()` - 137 edges
7. `Role` - 132 edges
8. `loadK8s()` - 120 edges
9. `Per-package doubles` - 120 edges
10. `newCapMgr()` - 117 edges

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

## Communities (109 total, 8 thin omitted)

### Community 0 - "App"
Cohesion: 0.06
Nodes (100): App, layer, New(), logArgs(), confirmAction(), confirmDelete(), confirmDowngrade(), confirmLayer() (+92 more)

### Community 1 - "newTestOps"
Cohesion: 0.08
Nodes (69): recOutput, broker_test.go, countContains(), cliScriptNameFromDest(), leaderCalls(), newTestOps(), outputForRole(), TestCountContains() (+61 more)

### Community 2 - ".kubectl"
Cohesion: 0.05
Nodes (19): runtime_test.go, allCustomMounted(), Cluster, HARoles(), podName(), pvcName(), RestartOrder(), stsName() (+11 more)

### Community 3 - "Role"
Cohesion: 0.11
Nodes (14): Ops, Transport, runExport(), exportScope(), opCtrExportConfig(), k8sAdminPassword(), k8sLogin(), opK8sExportConfig() (+6 more)

### Community 4 - "cli_test.go"
Cohesion: 0.07
Nodes (85): cli_test.go, examples_test.go, allowRuntime(), captureStderr(), collectPaths(), fakeBinaryOnPath(), firstLine(), opFailOn() (+77 more)

### Community 5 - "config_test.go"
Cohesion: 0.06
Nodes (75): config_test.go, assertContainerBlockDefaults(), assertContainerScaling(), envTree(), haNodesConfig(), TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes(), TestAdditionalUserNamesCollideOnDocker() (+67 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (52): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker cli (+44 more)

### Community 7 - "Manager"
Cohesion: 0.05
Nodes (22): scriptedMate, replMate, preflight_test.go, delegateSetting(), missingControllers(), pamCoreValue(), parseUnitProps(), certDigest() (+14 more)

### Community 8 - "K8sConfig"
Cohesion: 0.11
Nodes (15): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, LoadBalancer, cut() (+7 more)

### Community 9 - "completion_test.go"
Cohesion: 0.22
Nodes (23): completion_test.go, driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion() (+15 more)

### Community 10 - "Command"
Cohesion: 0.09
Nodes (24): cmdField, guardedCmd, command_test.go, duration_test.go, internal/config, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandArgv() (+16 more)

### Community 11 - "output_test.go"
Cohesion: 0.14
Nodes (25): internal/output, output_test.go, solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), columnWidths(), Sink, NewFunc(), pad() (+17 more)

### Community 12 - "NewCluster"
Cohesion: 0.10
Nodes (64): operator_test.go, GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), RenderOperator(), splitAfterNamespace(), splitOperatorBundle() (+56 more)

### Community 13 - "k8s/inspect.go"
Cohesion: 0.13
Nodes (35): inspect_test.go, anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), normalizeToList(), podHealth(), pvcLevel() (+27 more)

### Community 14 - "convert_test.go"
Cohesion: 0.07
Nodes (70): doc, Result, segment, vars, The keys, convert_test.go, DecodeStrict(), boolOf() (+62 more)

### Community 15 - "transform_test.go"
Cohesion: 0.10
Nodes (62): Block, TargetState, transform_test.go, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects() (+54 more)

### Community 16 - "manager_test.go"
Cohesion: 0.09
Nodes (100): manager_test.go, GuardPodmanEUID(), fileExists(), NewManager(), assertMode(), containsStr(), ctrCfg(), hasCall() (+92 more)

### Community 17 - "commands.go"
Cohesion: 0.13
Nodes (72): opFunc, roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd() (+64 more)

### Community 18 - "coverage_test.go"
Cohesion: 0.07
Nodes (66): coverage_test.go, verify_ops_test.go, matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle() (+58 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "localCfg"
Cohesion: 0.10
Nodes (47): semp_test.go, sempmate_test.go, bridgeHostPort(), sempPort(), sempV1OK(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA() (+39 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Load"
Cohesion: 0.08
Nodes (44): scalingTier, scaling_test.go, minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs() (+36 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.07
Nodes (76): CLIRunner, fakeRun, PhaseKind, SiteState, SwitchAction, SwitchPhase, matechannel_test.go, switchplan_test.go (+68 more)

### Community 24 - "newLocalOps"
Cohesion: 0.16
Nodes (40): verify_local_test.go, cliScriptPath(), isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn(), TestDefaultLocalAddrs() (+32 more)

### Community 25 - "diff_test.go"
Cohesion: 0.12
Nodes (37): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, diff_test.go, ancestorReported(), blockLabel() (+29 more)

### Community 26 - "renderDriver"
Cohesion: 0.15
Nodes (22): chunk, chunkResult, driver_test.go, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce() (+14 more)

### Community 27 - "hostpath_test.go"
Cohesion: 0.07
Nodes (35): hostpath_test.go, guardConfig(), BaseName(), CheckHostPath(), expandTilde(), expandTildeToken(), HasPathSeparator(), IsAbsHostPath() (+27 more)

### Community 28 - "importops_test.go"
Cohesion: 0.07
Nodes (71): ImportResult, PlannedSection, importops_test.go, defaultVPNFirst(), describeScope(), ImportPlan, importIgnore(), orNone() (+63 more)

### Community 29 - "watch_test.go"
Cohesion: 0.15
Nodes (32): watch_test.go, subtractWatch(), deniedCluster(), deniedRunner(), deployJSON(), namedOperatorJSON(), operatorItem(), TestAmbiguousOperatorAbortsTheDeploy() (+24 more)

### Community 31 - "statusreport_test.go"
Cohesion: 0.19
Nodes (26): statusreport_test.go, ownsPod(), reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportOmitsWorkloadsThatAreNotTheBrokers(), TestBrokerReportOrdersPodsOfTheSameRoleByName(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportReadsOnlyWhatTheEnvFileNames() (+18 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "NewCluster"
Cohesion: 0.08
Nodes (72): checkreport_test.go, deploy_test.go, preflight_test.go, captureStdout(), TestRunStreamsToThisProcess(), TestAdminPasswordPreflight(), TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestCheckReportSkipsEmptySections() (+64 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (50): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domaincerts_resolve_test.go, domaincerts_test.go, caNameSafe() (+42 more)

### Community 35 - "watchLoop"
Cohesion: 0.31
Nodes (5): watchInterval(), stdoutIsTerminal(), watchLabel(), watchLoop(), withWatch()

### Community 36 - "github.com/spf13/cobra.ShellCompDirective"
Cohesion: 0.33
Nodes (6): completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), isEnvFileName(), matching()

### Community 37 - ".operatorRows"
Cohesion: 0.15
Nodes (13): compareVersions(), imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), operatorRunningImage(), splitWatch(), watchFromContainers() (+5 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.07
Nodes (49): Region, annotate_test.go, applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV() (+41 more)

### Community 40 - "Role"
Cohesion: 0.09
Nodes (11): covUploadFailsTo, downloadErrTransport, fakeTransport, recDownload, recRun, recUpload, recUploadFile, runErrMatchTransport (+3 more)

### Community 41 - "ContainerSecrets"
Cohesion: 0.27
Nodes (7): ContainerSecrets(), ContainerSecret, TestContainerSecretNamesAreHostScoped(), TestContainerSecretsRedundancy(), TestSecretsAndCertDoNotNest(), TestComposeLabelsTheCertificateDigest(), TestServerCertIsASecretOnBothEngines()

### Community 42 - "inject_test.go"
Cohesion: 0.17
Nodes (28): lineRole, serviceLine, shutdownStyle, svcKey, inject_test.go, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "secrets_test.go"
Cohesion: 0.16
Nodes (27): secrets_test.go, secretSummary(), setOrMissing(), ResolveSecretValues(), bundleHash(), certCreate(), certFixture(), TestCertSecretLabelDrivesTheRestart() (+19 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.13
Nodes (33): exportconfigReadCounter, opRunner, euid_test.go, exportconfig_test.go, internal/cli, runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand() (+25 more)

### Community 46 - "io.Writer"
Cohesion: 0.15
Nodes (10): interactiveFailRunner, runFailRunner, noEnv, legacyMethod(), legacyMethod(), refuseEnv(), Run(), legacyMethod() (+2 more)

### Community 47 - "Platforms"
Cohesion: 0.16
Nodes (20): platform_test.go, parsePlatformList(), supported(), TestPlatformAnnotationsMatchDispatch(), TestPlatformOpsCoversEveryPlatform(), parseError(), DetectPlatforms(), newPlatforms() (+12 more)

### Community 48 - ".validate"
Cohesion: 0.13
Nodes (22): setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminPassState(), adminSecretRow(), containsString(), failRow() (+14 more)

### Community 49 - "config_test.go"
Cohesion: 0.14
Nodes (32): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, showCmd, span, AdditionalUser, Node (+24 more)

### Community 50 - "execguard_test.go"
Cohesion: 0.19
Nodes (14): execguard_test.go, clusterRules(), composeRules(), rulesFor(), runtimeRules(), siteRules(), TestAllowCommandIsNotASchemaKey(), TestCommandTildeExpansion() (+6 more)

### Community 51 - "internal/k8s"
Cohesion: 0.09
Nodes (30): The env-file generator page, adminsecret_test.go, check_test.go, generatorpage_test.go, internal/k8s, names_test.go, transport_test.go, BrokerPodSuffixShape() (+22 more)

### Community 52 - "config_ops_test.go"
Cohesion: 0.17
Nodes (14): config_ops_test.go, TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(), TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator(), TestExecShellScriptRefusesAnUnsafeNameBeforeUploading() (+6 more)

### Community 53 - "haCfg"
Cohesion: 0.08
Nodes (74): cluster_test.go, ops_test.go, prep_test.go, TestReachable(), newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit() (+66 more)

### Community 54 - ".runMatePhase"
Cohesion: 0.36
Nodes (6): isRunCLIRejection(), newlineIf(), replPhase1Rejected(), replPhase2Rejected(), replTransportFailure(), replVPNList()

### Community 55 - ".dump"
Cohesion: 0.11
Nodes (59): opCall, ops_k8s_test.go, Per-package doubles, failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook(), opArgvMatch() (+51 more)

### Community 56 - "Abbreviations"
Cohesion: 0.20
Nodes (9): Abbreviations, Flag shorthands, Node roles, Platforms, Command reference, Global flags, Index, Reading this reference (+1 more)

### Community 57 - "replication_test.go"
Cohesion: 0.10
Nodes (44): replRig, Commands, solace-util broker, Configuring a site, Data replication, Permissions, Reaching the mate, Switching roles (+36 more)

### Community 58 - ".OperatorApply"
Cohesion: 0.54
Nodes (3): canIAnswer(), canIRefused(), probe

### Community 59 - "replCfg"
Cohesion: 0.19
Nodes (28): matechannel_test.go, CLIArg(), CLIScriptPath(), execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf() (+20 more)

### Community 60 - "ReplVia"
Cohesion: 0.12
Nodes (13): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, viaDescription(), CheckSecretKeyRef(), sortedKeys(), validateReplEndpoints() (+5 more)

### Community 61 - "replConfig"
Cohesion: 0.18
Nodes (15): Replication, secretRef, replication_test.go, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded() (+7 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.18
Nodes (42): rootless_test.go, failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+34 more)

### Community 63 - "Test catalogue"
Cohesion: 0.25
Nodes (8): Coverage, internal/convert, internal/render, internal/tools/vulnjudge, Running the tests, Summary, Test catalogue, ValidScriptName()

### Community 64 - "github.com/hafio/go-deploy-utility/runner.Runner"
Cohesion: 0.19
Nodes (15): Repository layout, internal/engine, resolve_test.go, runner_test.go, TestChildEnvNamesAreNotSystemVariables(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), verboseExec() (+7 more)

### Community 65 - "Compose"
Cohesion: 0.18
Nodes (16): ContainerNoFile(), HealthCheck, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), Compose(), ComposeProject(), healthCmd(), Quadlet() (+8 more)

### Community 66 - "blocks_test.go"
Cohesion: 0.08
Nodes (53): BrokerType, Capture, Omission, token, blocks_test.go, internal/broker, omit_test.go, transcript_test.go (+45 more)

### Community 67 - "Ops"
Cohesion: 0.09
Nodes (25): mateChannelFunc, confirmActionStrict(), progress(), step(), checkAllowCommand(), checkFlagPlatforms(), commandPlatforms(), declaredList() (+17 more)

### Community 68 - "The eight rules"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 69 - ".podSection"
Cohesion: 0.28
Nodes (5): orNone(), orValue(), age(), roleRank(), discoveryServiceName()

### Community 70 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 71 - "Injectable seams"
Cohesion: 0.13
Nodes (15): Injectable seams, defaultLocalAddrs(), noRolePositional(), newExamplesCmd(), runExample(), emitOrWrite(), containerRenderRole(), containerRole() (+7 more)

### Community 72 - "Operations"
Cohesion: 0.06
Nodes (34): A replication switch refuses before changing anything, A wrapper runtime is refused, Bringing up a fresh cluster, Broker scope is a fixed classification, Docker and Podman mechanics, Docker compose secrets need compose 2.23.1+, Every destructive command confirms, Exit codes (+26 more)

### Community 73 - "tlssecret_test.go"
Cohesion: 0.43
Nodes (8): tlssecret_test.go, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsDeriveASecretName(), TestSuppliedCertsMakeTheSecretOurs(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), TestTLSServerSecretNameFollowsTheStates(), tlsCfg()

### Community 74 - ".RunCLI"
Cohesion: 0.05
Nodes (51): scripts_test.go, cliRunNames(), field(), runCLISkeleton(), validCLILine(), validName(), ValidVPNName(), validVPNName() (+43 more)

### Community 75 - "platform_test.go"
Cohesion: 0.28
Nodes (17): platform_test.go, runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection() (+9 more)

### Community 76 - "EnvPairs"
Cohesion: 0.12
Nodes (20): servercert_test.go, containerSecretSpecs(), EnvPairs(), groupKey(), itoa(), secretFilePath(), SecretPreflight(), ServerCertBundlePath() (+12 more)

### Community 77 - "runRootWith"
Cohesion: 0.09
Nodes (52): exit_test.go, Fixtures and doubles, Shared env fixtures, watch_test.go, echoRunner(), runCtr(), runRootWith(), TestConfigStepsDoNotLeakSecrets() (+44 more)

### Community 78 - "newRootCmd"
Cohesion: 0.17
Nodes (19): How a short form is resolved, aliases_test.go, applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing() (+11 more)

### Community 79 - "captureStdout"
Cohesion: 0.16
Nodes (16): capture(), captureStdout(), loadDirect(), TestCheckDeployWarnsWhenOperatorAbsent(), TestContainerRoleDetectsFromHostname(), TestDeployNeverLabelsNodes(), TestExecute(), TestOpCtrVerifyRedundancyRunsCoordinated() (+8 more)

### Community 81 - "replicationops_test.go"
Cohesion: 0.08
Nodes (55): cliMate, QueueState, replicationops_test.go, hasCall(), ReplRole, VPNRepl, ReplicationConfigResult, mateConvergenceShutdowns() (+47 more)

### Community 82 - "Platform"
Cohesion: 0.05
Nodes (39): keyValueEntries, artifactvalues_test.go, checkTimezone(), portOrRange(), TestCheckTimezone(), TestValidateContainerArtifactValues(), TestValidateImageReference(), TestValidateK8sArtifactValues() (+31 more)

### Community 83 - "limits_test.go"
Cohesion: 0.09
Nodes (51): limits_test.go, main_test.go, limitsMgr(), nrOpenProbe(), procLimits(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessCoreFloorFollowsTheConfiguredLimit() (+43 more)

### Community 84 - "Developer guide"
Cohesion: 0.29
Nodes (7): Build, Dev script tasks, Developer guide, Gates, Releases, Tests, Toolchain pin

### Community 85 - ".BackupRevertActivity"
Cohesion: 0.08
Nodes (20): AdminState, backupTarget, Credential, sempMate, transport_test.go, TestHTTPStatusHelpers(), schemaTransport(), anyHTTP2xx() (+12 more)

### Community 86 - "namespace_test.go"
Cohesion: 0.27
Nodes (11): namespace_test.go, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 87 - "broker/replication_test.go"
Cohesion: 0.08
Nodes (67): colSpan, replication_test.go, cliTransport(), containsEndpoint(), dashSpans(), flagByte(), MateConfig, gutterClear() (+59 more)

### Community 88 - "AdminSecret"
Cohesion: 0.07
Nodes (58): TLS, adminsecret_test.go, imagepullsecret_test.go, secrets_test.go, adminCfg(), TestAdminSecretNameFollowsTheStates(), TestAdminSecretRefusesTheOperatorsOwnNames(), TestContainersStillRequireTheAdminPassword() (+50 more)

### Community 89 - "render_test.go"
Cohesion: 0.17
Nodes (36): render_test.go, boolStr(), BrokerCR(), assertNoCheckoutPath(), containerArtifacts(), healthCheckFixture(), load(), loadRootful() (+28 more)

### Community 90 - "allowcommand_test.go"
Cohesion: 0.38
Nodes (11): allowcommand_test.go, TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb() (+3 more)

### Community 91 - "RuleFor"
Cohesion: 0.11
Nodes (36): Disposition, SectionRule, appliance_test.go, importdoc_test.go, sections_test.go, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent() (+28 more)

### Community 93 - "redundancy_test.go"
Cohesion: 0.21
Nodes (12): redundancy_test.go, pskCfg(), TestFillStandaloneNodeNameIgnoresAnEmptyHostname(), TestFillStandaloneNodeNameKeepsAConfiguredName(), TestFillStandaloneNodeNameNeverFillsHA(), TestFillStandaloneNodeNameUsesTheHost(), TestNothingGeneratesThePSK(), TestPSKEnvSatisfiesTheContainerRequirement() (+4 more)

### Community 95 - "Goldens"
Cohesion: 0.25
Nodes (9): Goldens, abbrevdoc_test.go, commanddoc_test.go, TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), firstDiff(), TestCommandDocs(), TestEveryDocCarriesTheSupportNotice() (+1 more)

### Community 96 - "gen-cert.sh"
Cohesion: 0.27
Nodes (13): ask(), askPass(), askSan(), askYesNo(), buildSanList(), checkExit(), concatCnIfNotBlank(), echoUsage() (+5 more)

### Community 97 - "storage_test.go"
Cohesion: 0.36
Nodes (9): storage_test.go, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 98 - "Container"
Cohesion: 0.15
Nodes (11): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestContainerCoreLimitSpellings(), applyContainerBlockDefaults(), podmanRunUser() (+3 more)

### Community 99 - "Configuration"
Cohesion: 0.15
Nodes (13): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+5 more)

### Community 100 - "runLeader"
Cohesion: 0.60
Nodes (4): leaderRun, runLeader(), TestAssertLeaderAsksBeforeActingFromABackup(), TestAssertLeaderPrimaryAndMonitorNeverAsk()

### Community 102 - "DetectPlatforms"
Cohesion: 0.21
Nodes (18): examples_test.go, internal/examples, Example, TestExamplesEmitsToStdout(), All(), Get(), List(), Names() (+10 more)

### Community 103 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.29
Nodes (6): Contract, Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 112 - "capRunner"
Cohesion: 0.13
Nodes (25): capCall, capRunner, runtime_test.go, callIndex(), countCalls(), TestManagerLogsCLIShell(), mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime() (+17 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.10
Nodes (26): containerState, healthState, inspect_test.go, internal/container, decodeInspect(), orUnknown(), printable(), inspectReply() (+18 more)

### Community 126 - "operatorversion_test.go"
Cohesion: 0.16
Nodes (19): operatorversion_test.go, operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsNotSkippedByAFailedRead(), TestConfirmNoDowngradeIsSilentInAPreview(), TestConfirmNoDowngradeProceedsWhenConfirmAccepts(), TestConfirmNoDowngradeRefusedWhenConfirmDeclines() (+11 more)

### Community 155 - "strings.Builder"
Cohesion: 0.09
Nodes (35): shorthand, WeightedNodeTerm, mdRow(), renderAbbrevDocs(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable() (+27 more)

## Knowledge Gaps
- **13 isolated node(s):** `solace`, `NO_COLOR`, `bash/000-env.sh`, `bash/010-deploy-operator.sh`, `bash/020-deploy-broker.sh` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 190 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `App` to `github.com/hafio/go-deploy-utility/runner.Runner`, `watchLoop`, `Ops`, `github.com/spf13/cobra.ShellCompDirective`, `Role`, `Manager`, `Injectable seams`, `platform_test.go`, `exportconfig_test.go`, `runRootWith`, `newRootCmd`, `commands.go`, `Platform`, `config_test.go`, `.dump`, `AdminSecret`, `replication_test.go`?**
  _High betweenness centrality (0.043) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `newTestOps()` (e.g. with `importops_test.go` and `Injectable seams`) actually correct?**
  _`newTestOps()` has 117 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `NO_COLOR`, `bash/000-env.sh` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.064065230052417 - nodes in this community are weakly interconnected._
- **Why does `Config` connect `AdminSecret` to `App`, `newTestOps`, `.kubectl`, `Role`, `cli_test.go`, `config_test.go`, `Manager`, `K8sConfig`, `Command`, `NewCluster`, `convert_test.go`, `manager_test.go`, `localCfg`, `Load`, `hostpath_test.go`, `strings.Builder`, `NewCluster`, `ResolveDomainCerts`, `ContainerSecrets`, `secrets_test.go`, `.validate`, `config_test.go`, `execguard_test.go`, `internal/k8s`, `haCfg`, `.dump`, `replCfg`, `replConfig`, `Compose`, `.podSection`, `tlssecret_test.go`, `EnvPairs`, `captureStdout`, `containerTransport`, `Platform`, `namespace_test.go`, `render_test.go`, `redundancy_test.go`, `storage_test.go`, `Container`, `Image`, `DetectPlatforms`, `capRunner`?**
  _High betweenness centrality (0.036) - this node is a cross-community bridge._
- **Are the 10 inferred relationships involving `Config` (e.g. with `config_test.go` and `execguard_test.go`) actually correct?**
  _`Config` has 10 INFERRED edges - model-reasoned connections that need verification._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.07967806841046278 - nodes in this community are weakly interconnected._
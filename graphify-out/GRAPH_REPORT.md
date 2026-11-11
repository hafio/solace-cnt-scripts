# Graph Report - solace-cnt-scripts  (2026-11-11)

## Corpus Check
- 202 files · ~713,135 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .golden 17, (none) 3, .cli 2)

## Summary
- 4065 nodes · 20953 edges · 110 communities (103 shown, 7 thin omitted)
- Extraction: 72% EXTRACTED · 28% INFERRED · 0% AMBIGUOUS · INFERRED: 5786 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `20ec2f70`
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
- eqArgs
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
- Command reference
- switchplan_test.go
- verify_local_test.go
- output_test.go
- diff_test.go
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
- opArgvMatch
- podName
- config_test.go
- scripts_test.go
- Coverage
- internal/broker
- prep_test.go
- .podSection
- .dump
- The eight rules
- replication_test.go
- .OperatorApply
- Platforms
- parseBody
- replConfig
- rootless_test.go
- Platform parity: solmq-conn-util and solace-util
- captureStderr
- secrets_test.go
- blocks.go
- Platform
- TestGeneratorPageEmbedsTheCLI
- k8s/inspect.go
- .validate
- .ApplyDefaults
- Operations
- ParseRole
- What `import-config` applies
- .checkUserManagerLimits
- EnvPairs
- Test catalogue
- echoRunner
- blocks.go
- preflight_test.go
- replicationops_test.go
- .validateContainer
- limits_test.go
- Configuration
- context.Context
- namespace_test.go
- broker/replication_test.go
- Config
- render_test.go
- Quadlet
- RuleFor
- NewExec
- containerTransport
- gen-cert.sh
- euid_test.go
- Container
- deploymentItem
- examples_test.go
- runLeader
- Platforms
- newRootCmd
- TestComposeProjectIsDeclaredNotDerived
- tlssecret_test.go
- capRunner
- .inspectContainer
- operatorversion_test.go
- strings.Builder

## God Nodes (most connected - your core abstractions)
1. `App` - 221 edges
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
- `annotate_test.go` --references--> `TestAnnotateIsDeterministic()`  [INFERRED]
  docs/test.md → internal/broker/annotate_test.go
- `annotate_test.go` --references--> `Capture`  [INFERRED]
  docs/test.md → internal/broker/blocks.go
- `annotate_test.go` --references--> `ParseBlocks()`  [INFERRED]
  docs/test.md → internal/broker/blocks.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (110 total, 7 thin omitted)

### Community 0 - "App"
Cohesion: 0.06
Nodes (94): App, mateChannelFunc, logOptions(), logSince(), newPerformCLIScriptCmd(), runExport(), runImport(), wantEnable() (+86 more)

### Community 1 - "newTestOps"
Cohesion: 0.08
Nodes (70): broker_test.go, countContains(), leaderCalls(), newTestOps(), TestCountContains(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs() (+62 more)

### Community 2 - ".kubectl"
Cohesion: 0.06
Nodes (18): Build, Dev script tasks, Developer guide, Gates, Releases, Tests, Toolchain pin, runtime_test.go (+10 more)

### Community 3 - "Role"
Cohesion: 0.04
Nodes (43): covUploadFailsTo, downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile (+35 more)

### Community 4 - "cli_test.go"
Cohesion: 0.07
Nodes (98): cli_test.go, allowRuntime(), collectPaths(), ctrRedeployApp(), fakeBinaryOnPath(), firstLine(), runCtr(), runRoot() (+90 more)

### Community 5 - "config_test.go"
Cohesion: 0.04
Nodes (81): Image, artifactvalues_test.go, config_test.go, redundancy_test.go, checkTimezone(), TestCheckTimezone(), TestValidateContainerArtifactValues(), TestValidateImageReference() (+73 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (50): Commands, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker cli, solace-util broker configure (+42 more)

### Community 7 - "Manager"
Cohesion: 0.08
Nodes (9): Manager, orNone(), platformTitle(), secretSummary(), setOrMissing(), runUserIDs(), TestManagerHonoursRuntime(), Output() (+1 more)

### Community 8 - "K8sConfig"
Cohesion: 0.11
Nodes (15): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, LoadBalancer, cut() (+7 more)

### Community 9 - "completion_test.go"
Cohesion: 0.17
Nodes (27): completion_test.go, driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion() (+19 more)

### Community 10 - "Command"
Cohesion: 0.07
Nodes (36): cmdField, guardedCmd, command_test.go, duration_test.go, execguard_test.go, internal/config, decodeCommand(), TestCommandArgsDoesNotAliasCommand() (+28 more)

### Community 11 - "eqArgs"
Cohesion: 0.09
Nodes (49): check_test.go, cluster_test.go, internal/k8s, names_test.go, ops_test.go, transport_test.go, BrokerPodSuffixShape(), TestReachable() (+41 more)

### Community 12 - "NewCluster"
Cohesion: 0.09
Nodes (61): operator_test.go, GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), RenderOperator(), splitAfterNamespace(), splitOperatorBundle() (+53 more)

### Community 13 - "main_test.go"
Cohesion: 0.18
Nodes (22): main_test.go, describe(), judge(), main(), plural(), run(), load(), TestJudge() (+14 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (71): doc, Result, segment, vars, The keys, convert_test.go, DecodeStrict(), boolOf() (+63 more)

### Community 15 - "transform_test.go"
Cohesion: 0.10
Nodes (59): Block, TargetState, transform_test.go, importIgnore(), injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs() (+51 more)

### Community 16 - "manager_test.go"
Cohesion: 0.09
Nodes (68): manager_test.go, GuardPodmanEUID(), fileExists(), hasCall(), listArgs(), newEchoMgr(), statusListing(), TestCheckEnvReportsBaseDirOnlyWhenSet() (+60 more)

### Community 17 - "commands.go"
Cohesion: 0.13
Nodes (71): opFunc, roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd() (+63 more)

### Community 18 - "coverage_test.go"
Cohesion: 0.07
Nodes (71): coverage_test.go, storage_test.go, verify_ops_test.go, ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle() (+63 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "verify_local_test.go"
Cohesion: 0.07
Nodes (59): semp_test.go, sempmate_test.go, TestLocalMateRunsThroughOps(), bridgeHostPort(), sempPort(), sempV1OK(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly() (+51 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Command reference"
Cohesion: 0.50
Nodes (4): Command reference, Global flags, Index, Tree

### Community 23 - "switchplan_test.go"
Cohesion: 0.07
Nodes (74): CLIRunner, fakeRun, PhaseKind, SiteState, SwitchAction, SwitchPhase, matechannel_test.go, switchplan_test.go (+66 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.14
Nodes (35): verify_local_test.go, cliScriptPath(), defaultLocalAddrs(), isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn() (+27 more)

### Community 25 - "output_test.go"
Cohesion: 0.13
Nodes (26): internal/output, output_test.go, lineSink(), progress(), columnWidths(), Sink, New(), NewFunc() (+18 more)

### Community 26 - "diff_test.go"
Cohesion: 0.12
Nodes (37): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, diff_test.go, ancestorReported(), blockLabel() (+29 more)

### Community 27 - "hostpath_test.go"
Cohesion: 0.08
Nodes (30): hostpath_test.go, guardConfig(), CheckHostPath(), expandTilde(), expandTildeToken(), HasPathSeparator(), IsAbsHostPath(), isPathSep() (+22 more)

### Community 28 - "importops_test.go"
Cohesion: 0.07
Nodes (71): ImportResult, PlannedSection, importops_test.go, defaultVPNFirst(), describeScope(), ImportPlan, orNone(), planSections() (+63 more)

### Community 29 - "watch_test.go"
Cohesion: 0.10
Nodes (32): watch_test.go, splitWatch(), subtractWatch(), deniedCluster(), deniedRunner(), deployJSON(), namedOperatorJSON(), operatorItem() (+24 more)

### Community 31 - "statusreport_test.go"
Cohesion: 0.10
Nodes (30): statusreport_test.go, operatorRunningImage(), ownsPod(), reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportOmitsWorkloadsThatAreNotTheBrokers(), TestBrokerReportOrdersPodsOfTheSameRoleByName(), TestBrokerReportPreviewStopsAtAnUnapprovedCommand() (+22 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "NewCluster"
Cohesion: 0.09
Nodes (68): checkreport_test.go, deploy_test.go, preflight_test.go, TestCheckStopsProbingWhenUnreachable(), TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestOperatorInstallVerdict(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestOperatorValidateAsksNothingAboutTheBroker() (+60 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (54): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domaincerts_resolve_test.go, domaincerts_test.go, domainCANames() (+46 more)

### Community 35 - "watchLoop"
Cohesion: 0.33
Nodes (4): watchInterval(), watchLabel(), watchLoop(), withWatch()

### Community 36 - "github.com/spf13/cobra.ShellCompDirective"
Cohesion: 0.33
Nodes (6): completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), isEnvFileName(), matching()

### Community 37 - ".rpc"
Cohesion: 0.12
Nodes (14): AdminState, backupTarget, Credential, sempMate, anyHTTP2xx(), curlConfigFlag(), curlConfigLine(), httpBody() (+6 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.08
Nodes (48): annotate_test.go, Annotate(), applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV() (+40 more)

### Community 40 - "Load"
Cohesion: 0.07
Nodes (47): scalingTier, Node roles, Reading this reference, scaling_test.go, minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError() (+39 more)

### Community 41 - "manager_test.go"
Cohesion: 0.12
Nodes (45): capRunner, Per-package doubles, replShowMate(), replVPNTable(), answer(), assertMode(), callIndex(), composeUp() (+37 more)

### Community 42 - "inject_test.go"
Cohesion: 0.17
Nodes (28): lineRole, serviceLine, shutdownStyle, svcKey, inject_test.go, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "ctrCfg"
Cohesion: 0.12
Nodes (50): inspect_test.go, secrets_test.go, inspectReply(), kvValue(), TestInspectContainerArgv(), TestInspectContainerRejectsAnUnusableReply(), TestInspectContainerSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment() (+42 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.15
Nodes (30): exportconfigReadCounter, opRunner, exportconfig_test.go, internal/cli, runRootWith(), confirmImport(), pluralVPN(), exportconfigDriverOutput() (+22 more)

### Community 46 - "io.Writer"
Cohesion: 0.18
Nodes (7): lockedWriter, canIResource(), canIVerb(), legacyMethod(), writeOut(), recRunner, rrCall

### Community 47 - "opArgvMatch"
Cohesion: 0.15
Nodes (29): opCall, captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch() (+21 more)

### Community 49 - "config_test.go"
Cohesion: 0.17
Nodes (13): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, showCmd, span, mustNotice(), hostMatches() (+5 more)

### Community 50 - "scripts_test.go"
Cohesion: 0.07
Nodes (39): scripts_test.go, archiveConfigsScript(), assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), enableDefaultUsersScript(), enableDefaultVPNScript() (+31 more)

### Community 51 - "Coverage"
Cohesion: 0.09
Nodes (31): interactiveFailRunner, runFailRunner, Repository layout, Coverage, internal/engine, runner_test.go, noEnv, redirect (+23 more)

### Community 52 - "internal/broker"
Cohesion: 0.09
Nodes (33): chunk, chunkResult, config_ops_test.go, driver_test.go, internal/broker, TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp() (+25 more)

### Community 53 - "prep_test.go"
Cohesion: 0.10
Nodes (54): adminsecret_test.go, prep_test.go, AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestOperatorAdminSecretNameMatchesTheBrokerNames() (+46 more)

### Community 54 - ".podSection"
Cohesion: 0.20
Nodes (4): age(), discoveryServiceName(), lbServiceName(), getRead

### Community 55 - ".dump"
Cohesion: 0.14
Nodes (47): Injectable seams, ops_k8s_test.go, opCanI(), TestOpK8sRemoveAborts(), TestRemoveBrokerKeepsTheNamespaceWhileItsPVCsRemain(), writeK8sDeployAllEnv(), confirmDelete(), confirmDowngrade() (+39 more)

### Community 56 - "The eight rules"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 57 - "replication_test.go"
Cohesion: 0.06
Nodes (86): replRig, Abbreviations, Commands, Flag shorthands, Platforms, solace-util broker, Configuring a site, Data replication (+78 more)

### Community 58 - ".OperatorApply"
Cohesion: 0.26
Nodes (3): operatorEscalationProbes(), operatorProbes(), operatorRegcred()

### Community 59 - "Platforms"
Cohesion: 0.14
Nodes (16): platform_test.go, TestPlatformConstantsMatchSchemaSections(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), TestExecBinariesCoversEveryPlatform(), TestGuardConfigIsValid(), TestDomainCertsDirsAreNotDefaulted(), newPlatforms() (+8 more)

### Community 60 - "parseBody"
Cohesion: 0.21
Nodes (11): Omission, Region, omit_test.go, regionFor(), keepRegion(), omitAtExport(), omittedList(), TestOmitAnnotateRoundTripsOmissions() (+3 more)

### Community 61 - "replConfig"
Cohesion: 0.18
Nodes (15): Replication, secretRef, replication_test.go, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded() (+7 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.18
Nodes (42): rootless_test.go, failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+34 more)

### Community 63 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.29
Nodes (6): Contract, Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 64 - "captureStderr"
Cohesion: 0.15
Nodes (20): layer, capture(), captureStderr(), TestAPipeIsRefusedAndNeverRead(), TestConfirmDowngradeNoPromptApprovesOnlyTheDowngrade(), TestConfirmLayerMatrix(), TestConfirmNonTTY(), TestPromptKeepsBufferedAnswersForLaterQuestions() (+12 more)

### Community 65 - "secrets_test.go"
Cohesion: 0.11
Nodes (36): The env-file generator page, secrets_test.go, GenSecrets(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), readTLSMaterial() (+28 more)

### Community 66 - "blocks.go"
Cohesion: 0.13
Nodes (23): token, blocks_test.go, TestAnnotateIsDeterministic(), brokerTypeFromSchema(), firstQuoted(), newBlock(), blockByIndex(), loadSample() (+15 more)

### Community 67 - "Platform"
Cohesion: 0.13
Nodes (23): step(), newConvertCmd(), runConvert(), newExamplesCmd(), runExample(), addOutFlags(), checkAllowCommand(), emitOrWrite() (+15 more)

### Community 68 - "TestGeneratorPageEmbedsTheCLI"
Cohesion: 0.12
Nodes (17): scalingKey, scalingSpelling, generatorpage_test.go, Scaling, scalingKeyIndex(), scalingKeyList(), generatorDefaults(), gunzipBase64() (+9 more)

### Community 69 - "k8s/inspect.go"
Cohesion: 0.12
Nodes (38): inspect_test.go, TestResolveEnvPath(), anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), podHealth(), pvcLevel() (+30 more)

### Community 70 - ".validate"
Cohesion: 0.14
Nodes (25): orNone(), orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminSecretRow(), containsString() (+17 more)

### Community 71 - ".ApplyDefaults"
Cohesion: 0.17
Nodes (10): guardCommandOf(), setGuardCommand(), TestValidatorAndExecutorAgree(), applyContainerBlockDefaults(), podmanRunUser(), setDefault(), setDefaultCmd(), setDefaultInt() (+2 more)

### Community 72 - "Operations"
Cohesion: 0.06
Nodes (38): solace-util, solace-util operator, A replication switch refuses before changing anything, A wrapper runtime is refused, Bringing up a fresh cluster, Broker scope is a fixed classification, Docker and Podman mechanics, Docker compose secrets need compose 2.23.1+ (+30 more)

### Community 73 - "ParseRole"
Cohesion: 0.14
Nodes (15): shorthand, How a short form is resolved, treeShorthands(), writeUnabbreviated(), applyAliases(), anchor(), argumentLine(), availableSubs() (+7 more)

### Community 74 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 75 - ".checkUserManagerLimits"
Cohesion: 0.19
Nodes (4): delegateSetting(), pamCoreValue(), parseUnitProps(), TestPamCoreValue()

### Community 76 - "EnvPairs"
Cohesion: 0.12
Nodes (24): internal/render, servercert_test.go, NodeIdentity, Compose(), ContainerSecrets(), containerSecretSpecs(), EnvPairs(), groupKey() (+16 more)

### Community 77 - "Test catalogue"
Cohesion: 0.14
Nodes (12): Fixtures and doubles, internal/convert, internal/tools/vulnjudge, Running the tests, Shared env fixtures, Summary, Test catalogue, legacyMethod() (+4 more)

### Community 78 - "echoRunner"
Cohesion: 0.13
Nodes (32): platform_test.go, watch_test.go, echoRunner(), TestCtrConfirmDeclined(), runPlatform(), TestLogOptionsBuildsOneSetForBothPlatforms(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer() (+24 more)

### Community 79 - "blocks.go"
Cohesion: 0.16
Nodes (18): BrokerType, Capture, transcript_test.go, ParseBlocks(), parseHeader(), readHeaderField(), splitLines(), bannerType() (+10 more)

### Community 80 - "preflight_test.go"
Cohesion: 0.27
Nodes (10): preflight_test.go, TestManagerLifecycleDocker(), TestManagerLifecyclePodmanSystemctl(), TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestComposeSecretEnvNamesCannotBeSystemVars(), TestPreflightFailureStopsLifecycle(), TestPreflightFailureStopsTheDeploy(), TestPreflightHintIsPlatformShaped() (+2 more)

### Community 81 - "replicationops_test.go"
Cohesion: 0.13
Nodes (44): replicationops_test.go, hasCall(), matchCLI(), TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), TestImportPlanAcceptsWhatExportProduced(), TestOmitExportConfigEndToEnd() (+36 more)

### Community 82 - ".validateContainer"
Cohesion: 0.04
Nodes (35): keyValueEntries, ReplVia, ReplViaKube, ReplViaSEMP, portOrRange(), TestValidContainerPort(), validContainerPort(), validCoreLimit() (+27 more)

### Community 83 - "limits_test.go"
Cohesion: 0.12
Nodes (32): limits_test.go, missingControllers(), limitsMgr(), nrOpenProbe(), procLimits(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessCoreFloorFollowsTheConfiguredLimit() (+24 more)

### Community 84 - "Configuration"
Cohesion: 0.15
Nodes (13): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+5 more)

### Community 85 - "context.Context"
Cohesion: 0.07
Nodes (11): cliMate, QueueState, scriptedMate, replMate, ReplRole, VPNRepl, mateConvergenceShutdowns(), reenableLine() (+3 more)

### Community 86 - "namespace_test.go"
Cohesion: 0.24
Nodes (15): namespace_test.go, nsItem(), nsList(), occCfg(), occCluster(), TestNamespaceOccupantsAsksOneQuestion(), TestNamespaceOccupantsDiscountsTheBrokerAndWhatItOwns(), TestNamespaceOccupantsDiscountsWhatKubernetesPutsThere() (+7 more)

### Community 87 - "broker/replication_test.go"
Cohesion: 0.06
Nodes (67): colSpan, replication_test.go, ValidVPNName(), validVPNName(), cliTransport(), containsEndpoint(), dashSpans(), flagByte() (+59 more)

### Community 88 - "Config"
Cohesion: 0.07
Nodes (36): AdditionalUser, Node, Redundancy, SEMP, TLS, adminsecret_test.go, imagepullsecret_test.go, adminCfg() (+28 more)

### Community 89 - "render_test.go"
Cohesion: 0.14
Nodes (43): render_test.go, boolStr(), BrokerCR(), assertNoCheckoutPath(), containerArtifacts(), healthCheckFixture(), load(), loadRootful() (+35 more)

### Community 90 - "Quadlet"
Cohesion: 0.20
Nodes (10): opCtrGenArtifact(), ContainerNoFile(), HealthCheck, TestLimitsCheckAssertsWhatTheArtifactAsks(), healthCmd(), Quadlet(), secretFilePath(), TestHealthCmdDefaultsToReadiness() (+2 more)

### Community 91 - "RuleFor"
Cohesion: 0.13
Nodes (36): Disposition, SectionRule, appliance_test.go, importdoc_test.go, sections_test.go, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent() (+28 more)

### Community 93 - "NewExec"
Cohesion: 0.38
Nodes (6): resolve_test.go, TestChildEnvNamesAreNotSystemVariables(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), NewExec()

### Community 96 - "gen-cert.sh"
Cohesion: 0.27
Nodes (13): ask(), askPass(), askSan(), askYesNo(), buildSanList(), checkExit(), concatCnIfNotBlank(), echoUsage() (+5 more)

### Community 97 - "euid_test.go"
Cohesion: 0.36
Nodes (10): euid_test.go, TestContainerRoleDetectsFromHostname(), runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand(), TestPodmanEUIDGuardSkips(), writeRootlessPodmanEnv(), containerRenderRole() (+2 more)

### Community 98 - "Container"
Cohesion: 0.19
Nodes (10): Container, DockerConfig, Network, PodmanConfig, Ulimits, assertContainerBlockDefaults(), assertContainerScaling(), TestApplyDefaultsDocker() (+2 more)

### Community 99 - "deploymentItem"
Cohesion: 0.31
Nodes (4): imageFromDeployment(), watchFromDeployment(), deploymentItem, deploymentList

### Community 100 - "examples_test.go"
Cohesion: 0.29
Nodes (7): examples_test.go, TestExamplesAcceptsAPlatformAbbreviation(), TestExamplesBareEmitsTheFullSchema(), TestExamplesNeedsNoEnvFile(), TestExamplesRefusesToOverwrite(), TestExamplesRejectsAnUnknownPlatform(), TestExamplesWritesOutFile()

### Community 101 - "runLeader"
Cohesion: 0.60
Nodes (4): leaderRun, runLeader(), TestAssertLeaderAsksBeforeActingFromABackup(), TestAssertLeaderPrimaryAndMonitorNeverAsk()

### Community 102 - "Platforms"
Cohesion: 0.16
Nodes (24): examples_test.go, internal/examples, Example, TestExamplesEmitsToStdout(), DetectPlatforms(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint(), TestDetectPlatformsMissingFile() (+16 more)

### Community 103 - "newRootCmd"
Cohesion: 0.07
Nodes (45): Goldens, abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, commanddoc_test.go, exit_test.go, TestAbbreviationDocs(), TestFlagShorthandsAreConsistent() (+37 more)

### Community 104 - "TestComposeProjectIsDeclaredNotDerived"
Cohesion: 0.67
Nodes (3): ComposeProject(), TestComposeProjectFoldsToComposesGrammar(), TestComposeProjectIsDeclaredNotDerived()

### Community 108 - "tlssecret_test.go"
Cohesion: 0.43
Nodes (8): tlssecret_test.go, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsDeriveASecretName(), TestSuppliedCertsMakeTheSecretOurs(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), TestTLSServerSecretNameFollowsTheStates(), tlsCfg()

### Community 112 - "capRunner"
Cohesion: 0.13
Nodes (25): capCall, internal/container, runtime_test.go, transport_test.go, New(), NewManager(), TestManagerLogsCLIShell(), mgrOver() (+17 more)

### Community 118 - ".inspectContainer"
Cohesion: 0.22
Nodes (8): health(), orUnknown(), decodeOne(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings(), orValue(), KV

### Community 126 - "operatorversion_test.go"
Cohesion: 0.22
Nodes (19): operatorversion_test.go, TestOperatorApply(), compareVersions(), imageTag(), operatorVersionWarning(), parseVersion(), operatorDeployJSON(), TestCompareVersions() (+11 more)

### Community 155 - "strings.Builder"
Cohesion: 0.11
Nodes (25): WeightedNodeTerm, mdRow(), renderAbbrevDocs(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable(), mdCell(), writeFlagTable() (+17 more)

## Knowledge Gaps
- **13 isolated node(s):** `solace`, `NO_COLOR`, `bash/000-env.sh`, `bash/010-deploy-operator.sh`, `bash/020-deploy-broker.sh` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 190 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **7 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `App` to `Role`, `cli_test.go`, `commands.go`, `switchplan_test.go`, `watchLoop`, `github.com/spf13/cobra.ShellCompDirective`, `manager_test.go`, `exportconfig_test.go`, `io.Writer`, `opArgvMatch`, `config_test.go`, `Coverage`, `.dump`, `replication_test.go`, `captureStderr`, `Platform`, `echoRunner`, `context.Context`, `Config`, `Quadlet`, `euid_test.go`, `newRootCmd`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `newTestOps()` (e.g. with `importops_test.go` and `Injectable seams`) actually correct?**
  _`newTestOps()` has 117 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `NO_COLOR`, `bash/000-env.sh` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.060352233676975946 - nodes in this community are weakly interconnected._
- **Why does `Config` connect `Config` to `App`, `newTestOps`, `.kubectl`, `Role`, `cli_test.go`, `config_test.go`, `Manager`, `K8sConfig`, `Command`, `eqArgs`, `NewCluster`, `convert_test.go`, `manager_test.go`, `coverage_test.go`, `verify_local_test.go`, `hostpath_test.go`, `strings.Builder`, `NewCluster`, `ResolveDomainCerts`, `Load`, `manager_test.go`, `ctrCfg`, `opArgvMatch`, `podName`, `config_test.go`, `prep_test.go`, `.podSection`, `.dump`, `replication_test.go`, `.OperatorApply`, `replConfig`, `secrets_test.go`, `TestGeneratorPageEmbedsTheCLI`, `.validate`, `.ApplyDefaults`, `EnvPairs`, `.validateContainer`, `namespace_test.go`, `render_test.go`, `Quadlet`, `containerTransport`, `Container`, `Platforms`, `tlssecret_test.go`, `capRunner`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Are the 11 inferred relationships involving `Config` (e.g. with `config_test.go` and `execguard_test.go`) actually correct?**
  _`Config` has 11 INFERRED edges - model-reasoned connections that need verification._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.076103500761035 - nodes in this community are weakly interconnected._
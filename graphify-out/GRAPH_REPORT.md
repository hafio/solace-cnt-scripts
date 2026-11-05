# Graph Report - solace-cnt-scripts  (2026-11-05)

## Corpus Check
- 203 files · ~705,655 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .golden 17, (none) 3, .cli 2)

## Summary
- 4048 nodes · 20664 edges · 109 communities (102 shown, 7 thin omitted)
- Extraction: 72% EXTRACTED · 28% INFERRED · 0% AMBIGUOUS · INFERRED: 5720 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `04c1b858`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- App
- newTestOps
- diff_test.go
- kubectlTransport
- cli_test.go
- config_test.go
- Command Reference Docs
- Manager
- K8sConfig
- completion_test.go
- CheckCommand
- deploy_test.go
- NewCluster
- Test catalogue
- convert_test.go
- transform_test.go
- manager_test.go
- commands.go
- coverage_test.go
- Bash Dev Script
- Echo
- PowerShell Dev Script
- Load
- switchplan_test.go
- newLocalOps
- runner_test.go
- renderDriver
- statusreport_test.go
- importops_test.go
- Exec
- Solace Module Root
- .validateContainer
- CLAUDE.md
- main_test.go
- ResolveDomainCerts
- watchLoop
- newRootCmd
- NewCluster
- annotate_test.go
- output_test.go
- Compose
- inject_test.go
- secrets_test.go
- exportconfig_test.go
- context.Context
- localCfg
- config_test.go
- internal/k8s
- cli/platform_test.go
- .RunCLI
- prep_test.go
- Operations
- .dump
- context.Context
- replication_test.go
- .OperatorApply
- replCfg
- Injectable seams
- .validateK8s
- rootless_test.go
- Exporting and importing configuration
- Platform
- Coverage
- github.com/spf13/cobra.ShellCompDirective
- NewMateChannel
- The eight rules
- What `import-config` applies
- Removing a broker: what stays, what goes
- internal/broker
- Command reference
- config_ops_test.go
- Role
- .podSection
- Docker and Podman mechanics
- writeCtrStandaloneEnv
- .checkUserManagerLimits
- recRunner
- replicationops_test.go
- .Validate
- limits_test.go
- Developer guide
- .BackupRevertActivity
- k8s/inspect.go
- broker/replication_test.go
- AdminSecret
- render_test.go
- Config
- RuleFor
- namespace_test.go
- tlssecret_test.go
- Platform parity: solmq-conn-util and solace-util
- replConfig
- Container
- Configuration
- Platforms
- operatorversion_test.go
- runLeader
- adminsecret_test.go
- scripts.go
- capRunner
- Container State Inspection
- bytes.Buffer
- testing.T
- strings.Builder

## God Nodes (most connected - your core abstractions)
1. `App` - 216 edges
2. `newTestOps()` - 167 edges
3. `Config` - 146 edges
4. `cli_test.go` - 146 edges
5. `NewCluster()` - 143 edges
6. `ctrCfg()` - 133 edges
7. `Role` - 132 edges
8. `Per-package doubles` - 125 edges
9. `loadK8s()` - 120 edges
10. `newCapMgr()` - 114 edges

## Surprising Connections (you probably didn't know these)
- `Repository layout` --references--> `Execute()`  [INFERRED]
  docs/developer.md → internal/cli/root.go
- `The preflight` --references--> `info()`  [INFERRED]
  docs/operations.md → internal/k8s/checkreport.go
- `omit_test.go` --references--> `Annotate()`  [INFERRED]
  docs/test.md → internal/broker/annotate.go
- `annotate_test.go` --references--> `Capture`  [INFERRED]
  docs/test.md → internal/broker/blocks.go
- `annotate_test.go` --references--> `ParseBlocks()`  [INFERRED]
  docs/test.md → internal/broker/blocks.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (109 total, 7 thin omitted)

### Community 0 - "App"
Cohesion: 0.07
Nodes (95): App, layer, logArgs(), confirmAction(), confirmDelete(), confirmDowngrade(), confirmLayer(), confirmRestart() (+87 more)

### Community 1 - "newTestOps"
Cohesion: 0.07
Nodes (71): fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, broker_test.go, countContains() (+63 more)

### Community 2 - "diff_test.go"
Cohesion: 0.12
Nodes (36): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, diff_test.go, ancestorReported(), blockLabel() (+28 more)

### Community 4 - "cli_test.go"
Cohesion: 0.06
Nodes (102): cli_test.go, examples_test.go, allowRuntime(), captureStderr(), collectPaths(), fakeBinaryOnPath(), firstLine(), runRoot() (+94 more)

### Community 5 - "config_test.go"
Cohesion: 0.06
Nodes (73): Image, artifactvalues_test.go, config_test.go, TestValidateContainerArtifactValues(), TestValidateImageReference(), TestValidateK8sArtifactValues(), TestValidContainerPort(), validPlatformConfig() (+65 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (52): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker cli (+44 more)

### Community 7 - "Manager"
Cohesion: 0.07
Nodes (18): internal/container, preflight_test.go, certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle() (+10 more)

### Community 8 - "K8sConfig"
Cohesion: 0.11
Nodes (15): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, LoadBalancer, cut() (+7 more)

### Community 9 - "completion_test.go"
Cohesion: 0.22
Nodes (24): completion_test.go, driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion() (+16 more)

### Community 10 - "CheckCommand"
Cohesion: 0.10
Nodes (33): cmdField, commandRules, guardedCmd, execguard_test.go, Command, checkBinary(), CheckCommand(), checkFlagShape() (+25 more)

### Community 11 - "deploy_test.go"
Cohesion: 0.16
Nodes (26): deploy_test.go, brokerJSON(), logBuf(), TestConfirmBrokerAppliedAcceptsTheObjectItApplied(), TestConfirmBrokerAppliedFailsOnADifferentObject(), TestConfirmBrokerAppliedFailsWhenTheObjectIsNotThere(), TestConfirmBrokerAppliedSurfacesReadFailures(), TestConfirmBrokerAppliedTreatsSilenceAsSkipped() (+18 more)

### Community 12 - "NewCluster"
Cohesion: 0.10
Nodes (60): operator_test.go, GenOperator(), operatorImage(), RenderOperator(), boolPtr(), bundledDefaultPorts(), commentedExamplePorts(), covSwapBundle() (+52 more)

### Community 13 - "Test catalogue"
Cohesion: 0.33
Nodes (6): internal/convert, internal/examples, internal/render, internal/tools/vulnjudge, Running the tests, Test catalogue

### Community 14 - "convert_test.go"
Cohesion: 0.07
Nodes (69): doc, Result, segment, vars, The keys, convert_test.go, DecodeStrict(), boolOf() (+61 more)

### Community 15 - "transform_test.go"
Cohesion: 0.09
Nodes (60): Block, TargetState, transform_test.go, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects() (+52 more)

### Community 16 - "manager_test.go"
Cohesion: 0.09
Nodes (94): manager_test.go, GuardPodmanEUID(), fileExists(), assertMode(), containsStr(), ctrCfg(), hasCall(), maskedKeys() (+86 more)

### Community 17 - "commands.go"
Cohesion: 0.13
Nodes (72): opFunc, roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd() (+64 more)

### Community 18 - "coverage_test.go"
Cohesion: 0.06
Nodes (69): downloadErrTransport, runErrMatchTransport, runErrTransport, coverage_test.go, verify_ops_test.go, matchCLI(), ranContains(), writeFile() (+61 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "Echo"
Cohesion: 0.14
Nodes (8): interactiveFailRunner, Runner seams, TestChildEnvNamesAreNotSystemVariables(), Echo, EnvRunner, MaskEnv(), Quote(), quoteTok()

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Load"
Cohesion: 0.05
Nodes (65): scalingTier, Node roles, Reading this reference, scaling_test.go, minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError() (+57 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.07
Nodes (74): CLIRunner, fakeRun, PhaseKind, SiteState, SwitchAction, SwitchPhase, matechannel_test.go, switchplan_test.go (+66 more)

### Community 24 - "newLocalOps"
Cohesion: 0.16
Nodes (40): verify_local_test.go, cliScriptPath(), isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn(), TestDefaultLocalAddrs() (+32 more)

### Community 25 - "runner_test.go"
Cohesion: 0.19
Nodes (24): runner_test.go, captureStdout(), helperCommand(), TestEchoDefaultWriter(), TestEchoOutput(), TestEchoOutputInput(), TestEchoRun(), TestEchoRunEnv() (+16 more)

### Community 26 - "renderDriver"
Cohesion: 0.15
Nodes (22): chunk, chunkResult, driver_test.go, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce() (+14 more)

### Community 27 - "statusreport_test.go"
Cohesion: 0.19
Nodes (26): statusreport_test.go, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportOmitsWorkloadsThatAreNotTheBrokers(), TestBrokerReportOrdersPodsOfTheSameRoleByName(), TestBrokerReportPreviewStopsAtAnUnapprovedCommand(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportReadsOnlyWhatTheEnvFileNames() (+18 more)

### Community 28 - "importops_test.go"
Cohesion: 0.07
Nodes (72): ImportResult, PlannedSection, importops_test.go, defaultVPNFirst(), describeScope(), ImportPlan, importIgnore(), orNone() (+64 more)

### Community 29 - "Exec"
Cohesion: 0.20
Nodes (12): Keys that were renamed, internal/engine, resolve_test.go, Exec, TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveMissingBinaryIsActionable() (+4 more)

### Community 31 - ".validateContainer"
Cohesion: 0.18
Nodes (8): portOrRange(), validContainerPort(), validCoreLimit(), validHealthDuration(), invertedCPUSetRange(), missingErr(), platformKey(), validIdent()

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "main_test.go"
Cohesion: 0.20
Nodes (19): main_test.go, describe(), judge(), main(), plural(), run(), load(), TestJudge() (+11 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (50): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domaincerts_resolve_test.go, domaincerts_test.go, caNameSafe() (+42 more)

### Community 35 - "watchLoop"
Cohesion: 0.32
Nodes (5): watchInterval(), stdoutIsTerminal(), watchLabel(), watchLoop(), withWatch()

### Community 36 - "newRootCmd"
Cohesion: 0.07
Nodes (47): runFailRunner, aliases_test.go, allowcommand_test.go, exit_test.go, Fixtures and doubles, Shared env fixtures, Summary, applyAliases() (+39 more)

### Community 37 - "NewCluster"
Cohesion: 0.12
Nodes (42): checkreport_test.go, preflight_test.go, TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestCheckReportSkipsEmptySections(), TestOperatorInstallVerdict(), TestOperatorInstallVerdictShowsOnceInEveryView(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestOperatorValidateAsksNothingAboutTheBroker() (+34 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.12
Nodes (45): annotate_test.go, Annotate(), applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV() (+37 more)

### Community 40 - "output_test.go"
Cohesion: 0.14
Nodes (25): internal/output, output_test.go, columnWidths(), Level, Row, Sink, NewFunc(), pad() (+17 more)

### Community 41 - "Compose"
Cohesion: 0.11
Nodes (27): servercert_test.go, NodeIdentity, Compose(), composeEscape(), ContainerSecrets(), containerSecretSpecs(), EnvPairs(), ContainerSecret (+19 more)

### Community 42 - "inject_test.go"
Cohesion: 0.17
Nodes (28): lineRole, serviceLine, shutdownStyle, svcKey, inject_test.go, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "secrets_test.go"
Cohesion: 0.15
Nodes (29): secrets_test.go, solaceRows(), ResolveSecretValues(), bundleHash(), certCreate(), certFixture(), TestCertSecretLabelDrivesTheRestart(), TestContainerBundleCarriesTheChain() (+21 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.15
Nodes (28): exportconfigReadCounter, opRunner, exportconfig_test.go, internal/cli, runRootWith(), pluralVPN(), exportconfigDriverOutput(), exportconfigTransportOutput() (+20 more)

### Community 47 - "context.Context"
Cohesion: 0.05
Nodes (18): cliMate, covUploadFailsTo, QueueState, scriptedMate, uploadErrTransport, replMate, containerTransport, ReplRole (+10 more)

### Community 48 - "localCfg"
Cohesion: 0.10
Nodes (47): semp_test.go, sempmate_test.go, bridgeHostPort(), sempPort(), sempV1OK(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA() (+39 more)

### Community 49 - "config_test.go"
Cohesion: 0.15
Nodes (33): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, span, AdditionalUser, Node, Redundancy (+25 more)

### Community 50 - "internal/k8s"
Cohesion: 0.20
Nodes (16): adminsecret_test.go, check_test.go, generatorpage_test.go, internal/k8s, transport_test.go, AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg() (+8 more)

### Community 51 - "cli/platform_test.go"
Cohesion: 0.24
Nodes (19): platform_test.go, supported(), runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations() (+11 more)

### Community 52 - ".RunCLI"
Cohesion: 0.10
Nodes (15): cliRunNames(), field(), runCLISkeleton(), validCLILine(), validName(), rejectionIn(), archiveConfigsScript(), gatherConfigsScript() (+7 more)

### Community 53 - "prep_test.go"
Cohesion: 0.06
Nodes (88): cluster_test.go, names_test.go, ops_test.go, prep_test.go, BrokerPodSuffixShape(), TestReachable(), TestResolveStorageClass(), newCluster() (+80 more)

### Community 54 - "Operations"
Cohesion: 0.12
Nodes (17): A replication switch refuses before changing anything, A wrapper runtime is refused, Bringing up a fresh cluster, Docker compose secrets need compose 2.23.1+, Exit codes, Extra CLI users differ by platform, Forbidden: cannot list deployments at the cluster scope, Import reported failure (+9 more)

### Community 55 - ".dump"
Cohesion: 0.10
Nodes (64): opCall, ops_k8s_test.go, Per-package doubles, failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook(), loadDirect() (+56 more)

### Community 56 - "context.Context"
Cohesion: 0.07
Nodes (12): runtime_test.go, Cluster, normalizeToList(), podName(), pvcName(), stsName(), TestResourceNames(), bareKind() (+4 more)

### Community 57 - "replication_test.go"
Cohesion: 0.08
Nodes (55): replRig, Commands, solace-util broker, Configuring a site, Data replication, Permissions, Reaching the mate, Switching roles (+47 more)

### Community 58 - ".OperatorApply"
Cohesion: 0.18
Nodes (10): joinYAMLDocs(), nonEmpty(), operatorEscalationProbes(), operatorProbes(), splitAfterNamespace(), splitOperatorBundle(), canIAnswer(), canIRefused() (+2 more)

### Community 59 - "replCfg"
Cohesion: 0.26
Nodes (22): matechannel_test.go, ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf(), replCfg(), replSite(), sempSite() (+14 more)

### Community 60 - "Injectable seams"
Cohesion: 0.16
Nodes (10): Abbreviations, Flag shorthands, How a short form is resolved, Platforms, Injectable seams, defaultLocalAddrs(), argumentLine(), availableSubs() (+2 more)

### Community 61 - ".validateK8s"
Cohesion: 0.15
Nodes (11): keyValueEntries, CheckSecretKeyRef(), requireAll(), requireKeyValue(), validDNSLabel(), validDNSSubdomain(), validK8sID(), validPortName() (+3 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.19
Nodes (40): rootless_test.go, failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+32 more)

### Community 63 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 64 - "Platform"
Cohesion: 0.10
Nodes (30): confirmActionStrict(), progress(), step(), warn(), newConvertCmd(), runConvert(), newExamplesCmd(), runExample() (+22 more)

### Community 65 - "Coverage"
Cohesion: 0.16
Nodes (16): Coverage, ValidScriptName(), ContainerNoFile(), HealthCheck, TestLimitsCheckAssertsWhatTheArtifactAsks(), escapePercent(), healthCmd(), Quadlet() (+8 more)

### Community 66 - "github.com/spf13/cobra.ShellCompDirective"
Cohesion: 0.33
Nodes (6): completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), isEnvFileName(), matching()

### Community 67 - "NewMateChannel"
Cohesion: 0.33
Nodes (6): CLIArg(), CLIScriptPath(), execArgs(), NewMateChannel(), TestPodNameForIsTheOneSpelling(), podNameFor()

### Community 68 - "The eight rules"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 70 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 71 - "Removing a broker: what stays, what goes"
Cohesion: 0.40
Nodes (5): Every destructive command confirms, Removing a broker: what stays, what goes, Removing the operator does not always remove it, The layer flag raises the question, The namespace is only offered when it is empty

### Community 72 - "internal/broker"
Cohesion: 0.06
Nodes (53): BrokerType, Capture, Omission, Region, token, blocks_test.go, internal/broker, omit_test.go (+45 more)

### Community 73 - "Command reference"
Cohesion: 0.50
Nodes (4): Command reference, Global flags, Index, Tree

### Community 74 - "config_ops_test.go"
Cohesion: 0.19
Nodes (13): config_ops_test.go, TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(), TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator(), TestExecShellScriptRefusesAnUnsafeNameBeforeUploading() (+5 more)

### Community 75 - "Role"
Cohesion: 0.07
Nodes (30): mateChannelFunc, Ops, New(), ValidVPNName(), validVPNName(), Transport, runExport(), exportScope() (+22 more)

### Community 76 - ".podSection"
Cohesion: 0.19
Nodes (7): orNone(), orValue(), age(), discoveryServiceName(), HARoles(), lbServiceName(), ownsPod()

### Community 77 - "Docker and Podman mechanics"
Cohesion: 0.40
Nodes (5): Docker and Podman mechanics, Re-deploying is safe and explicit, Rootless podman prerequisites, What the broker cannot gain, Which user the broker runs as

### Community 78 - "writeCtrStandaloneEnv"
Cohesion: 0.16
Nodes (29): euid_test.go, watch_test.go, echoRunner(), runCtr(), TestConfiguredRouternameSurvivesTheFallback(), TestCtrConfirmDeclined(), TestCtrDiagnosticsDryRun(), TestCtrErrorPaths() (+21 more)

### Community 79 - ".checkUserManagerLimits"
Cohesion: 0.22
Nodes (6): delegateSetting(), limitText(), missingControllers(), pamCoreValue(), parseLimit(), parseUnitProps()

### Community 80 - "recRunner"
Cohesion: 0.23
Nodes (4): canIResource(), canIVerb(), TestTransportExecArgs(), recRunner

### Community 81 - "replicationops_test.go"
Cohesion: 0.11
Nodes (51): replicationops_test.go, hasCall(), ReplicationConfigResult, isRunCLIRejection(), mateConvergenceShutdowns(), missingListedVPNs(), newlineIf(), PlannedRoles() (+43 more)

### Community 82 - ".Validate"
Cohesion: 0.12
Nodes (11): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, viaDescription(), checkCredentialChars(), foldToEnvVar(), sortedKeys() (+3 more)

### Community 83 - "limits_test.go"
Cohesion: 0.16
Nodes (32): limits_test.go, limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessCoreFloorFollowsTheConfiguredLimit(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager() (+24 more)

### Community 84 - "Developer guide"
Cohesion: 0.08
Nodes (26): shorthand, Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout (+18 more)

### Community 85 - ".BackupRevertActivity"
Cohesion: 0.09
Nodes (20): AdminState, backupTarget, Credential, sempMate, transport_test.go, TestHTTPStatusHelpers(), schemaTransport(), anyHTTP2xx() (+12 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.10
Nodes (40): inspect_test.go, anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), podHealth(), pvcLevel(), replicaLevel() (+32 more)

### Community 87 - "broker/replication_test.go"
Cohesion: 0.08
Nodes (66): colSpan, replication_test.go, cliTransport(), containsEndpoint(), dashSpans(), flagByte(), MateConfig, gutterClear() (+58 more)

### Community 88 - "AdminSecret"
Cohesion: 0.14
Nodes (34): secrets_test.go, GenSecrets(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), operatorRegcred(), readTLSMaterial() (+26 more)

### Community 89 - "render_test.go"
Cohesion: 0.15
Nodes (41): render_test.go, boolStr(), BrokerCR(), ComposeProject(), containerArtifacts(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets() (+33 more)

### Community 90 - "Config"
Cohesion: 0.09
Nodes (30): TLS, Config, clearTierDerived(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminPassState() (+22 more)

### Community 91 - "RuleFor"
Cohesion: 0.12
Nodes (39): Disposition, SectionRule, appliance_test.go, importdoc_test.go, sections_test.go, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent() (+31 more)

### Community 92 - "namespace_test.go"
Cohesion: 0.31
Nodes (11): namespace_test.go, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 93 - "tlssecret_test.go"
Cohesion: 0.43
Nodes (8): tlssecret_test.go, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsDeriveASecretName(), TestSuppliedCertsMakeTheSecretOurs(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), TestTLSServerSecretNameFollowsTheStates(), tlsCfg()

### Community 95 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.40
Nodes (5): Contract, Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util

### Community 97 - "replConfig"
Cohesion: 0.10
Nodes (27): Replication, secretRef, duration_test.go, imagepullsecret_test.go, internal/config, replication_test.go, CanonicalDuration(), TestCanonicalDuration() (+19 more)

### Community 98 - "Container"
Cohesion: 0.27
Nodes (6): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestContainerCoreLimitSpellings()

### Community 100 - "Configuration"
Cohesion: 0.17
Nodes (12): Bring your own TLS Secret, Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling, Scaling tiers (+4 more)

### Community 102 - "Platforms"
Cohesion: 0.12
Nodes (32): examples_test.go, platform_test.go, Example, TestExamplesEmitsToStdout(), TestPlatformConstantsMatchSchemaSections(), TestExecBinariesCoversEveryPlatform(), parseError(), DetectPlatforms() (+24 more)

### Community 107 - "operatorversion_test.go"
Cohesion: 0.18
Nodes (19): operatorversion_test.go, compareVersions(), imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), operatorDeployJSON(), TestCompareVersions() (+11 more)

### Community 108 - "runLeader"
Cohesion: 0.60
Nodes (4): leaderRun, runLeader(), TestAssertLeaderAsksBeforeActingFromABackup(), TestAssertLeaderPrimaryAndMonitorNeverAsk()

### Community 109 - "adminsecret_test.go"
Cohesion: 0.21
Nodes (10): adminsecret_test.go, adminCfg(), TestAdminSecretNameFollowsTheStates(), TestAdminSecretRefusesTheOperatorsOwnNames(), TestContainersStillRequireTheAdminPassword(), TestKubernetesValidatesWithoutAnAdminPassword(), TestMonitorPasswordAndKeyNeedTheAdminPassword(), TestOperatorAdminSecretName() (+2 more)

### Community 110 - "scripts.go"
Cohesion: 0.09
Nodes (43): showCmd, scripts_test.go, TestServerCert(), TestDomainCertsLoadsInSortedNameOrder(), TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript() (+35 more)

### Community 112 - "capRunner"
Cohesion: 0.12
Nodes (25): capCall, capRunner, runtime_test.go, NewManager(), callIndex(), TestManagerLogsCLIShell(), mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime() (+17 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.10
Nodes (25): containerState, healthState, inspect_test.go, decodeInspect(), orUnknown(), printable(), inspectReply(), kvValue() (+17 more)

### Community 126 - "bytes.Buffer"
Cohesion: 0.15
Nodes (32): watch_test.go, subtractWatch(), deniedCluster(), deniedRunner(), deployJSON(), namedOperatorJSON(), operatorItem(), TestAmbiguousOperatorAbortsTheDeploy() (+24 more)

### Community 128 - "testing.T"
Cohesion: 0.05
Nodes (75): command_test.go, hostpath_test.go, redundancy_test.go, storage_test.go, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString() (+67 more)

### Community 155 - "strings.Builder"
Cohesion: 0.10
Nodes (31): WeightedNodeTerm, mdRow(), renderAbbrevDocs(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable(), writeUnabbreviated(), anchor() (+23 more)

## Knowledge Gaps
- **13 isolated node(s):** `solace`, `NO_COLOR`, `internal/broker`, `internal/cli`, `internal/config` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 184 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **7 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `App` to `Platform`, `github.com/spf13/cobra.ShellCompDirective`, `watchLoop`, `newRootCmd`, `Manager`, `Role`, `exportconfig_test.go`, `writeCtrStandaloneEnv`, `context.Context`, `commands.go`, `config_test.go`, `cli/platform_test.go`, `Echo`, `.dump`, `replication_test.go`, `Config`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `newTestOps()` (e.g. with `importops_test.go` and `Injectable seams`) actually correct?**
  _`newTestOps()` has 117 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `NO_COLOR`, `internal/broker` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.06765463917525773 - nodes in this community are weakly interconnected._
- **Why does `Config` connect `Config` to `App`, `testing.T`, `newTestOps`, `kubectlTransport`, `cli_test.go`, `config_test.go`, `Manager`, `K8sConfig`, `CheckCommand`, `NewCluster`, `Test catalogue`, `convert_test.go`, `manager_test.go`, `Load`, `strings.Builder`, `ResolveDomainCerts`, `NewCluster`, `Compose`, `secrets_test.go`, `context.Context`, `localCfg`, `config_test.go`, `internal/k8s`, `prep_test.go`, `.dump`, `context.Context`, `replCfg`, `Platform`, `Coverage`, `NewMateChannel`, `.RedundancyEnabled`, `Role`, `.podSection`, `AdminSecret`, `render_test.go`, `namespace_test.go`, `tlssecret_test.go`, `replConfig`, `Container`, `adminsecret_test.go`, `capRunner`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.06751054852320675 - nodes in this community are weakly interconnected._
- **Why does `Per-package doubles` connect `.dump` to `App`, `newTestOps`, `cli_test.go`, `config_test.go`, `Manager`, `NewCluster`, `convert_test.go`, `transform_test.go`, `manager_test.go`, `coverage_test.go`, `Echo`, `Load`, `switchplan_test.go`, `newLocalOps`, `runner_test.go`, `strings.Builder`, `importops_test.go`, `Exec`, `statusreport_test.go`, `newRootCmd`, `secrets_test.go`, `exportconfig_test.go`, `context.Context`, `localCfg`, `.RunCLI`, `prep_test.go`, `context.Context`, `replication_test.go`, `.OperatorApply`, `replCfg`, `Role`, `recRunner`, `Developer guide`, `.BackupRevertActivity`, `k8s/inspect.go`, `broker/replication_test.go`, `AdminSecret`, `replConfig`, `operatorversion_test.go`, `runLeader`, `capRunner`, `bytes.Buffer`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
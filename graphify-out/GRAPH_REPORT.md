# Graph Report - solace-cnt-scripts  (2026-11-05)

## Corpus Check
- 204 files · ~700,472 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .golden 17, (none) 3, .cli 2)

## Summary
- 4046 nodes · 20618 edges · 103 communities (96 shown, 7 thin omitted)
- Extraction: 72% EXTRACTED · 28% INFERRED · 0% AMBIGUOUS · INFERRED: 5687 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `35f98912`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- App
- newTestOps
- context.Context
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
- internal/broker
- PowerShell Dev Script
- Load
- switchplan_test.go
- newLocalOps
- runner_test.go
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
- newRootCmd
- Exec
- annotate_test.go
- output_test.go
- Compose
- inject_test.go
- secrets_test.go
- exportconfig_test.go
- VPNRepl
- Echo
- config_test.go
- containerState
- newEchoMgr
- scripts.go
- haCfg
- .runMatePhase
- .dump
- Abbreviations
- replication_test.go
- .OperatorApply
- ReplSite
- Command reference
- ParseBlocks
- rootless_test.go
- Test catalogue
- Platform
- Quadlet
- transcript_test.go
- opPerformReplication
- The eight rules
- What `import-config` applies
- redundancy_test.go
- .Restart
- Role
- Coverage
- Image
- cli/watch_test.go
- adminsecret_test.go
- containerTransport
- replicationops_test.go
- .validateContainer
- limits_test.go
- Developer guide
- .BackupRevertActivity
- k8s/inspect.go
- replication_test.go
- AdminSecret
- render_test.go
- checkreport_test.go
- RuleFor
- internal/config
- .report
- gen-cert.sh
- Container
- Configuration
- Platforms
- .decodeScalingEntry
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
5. `NewCluster()` - 143 edges
6. `ctrCfg()` - 133 edges
7. `Role` - 132 edges
8. `Per-package doubles` - 125 edges
9. `loadK8s()` - 120 edges
10. `newCapMgr()` - 114 edges

## Surprising Connections (you probably didn't know these)
- `Repository layout` --references--> `Execute()`  [INFERRED]
  docs/developer.md → internal/cli/root.go
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

## Communities (103 total, 7 thin omitted)

### Community 0 - "App"
Cohesion: 0.06
Nodes (107): App, mateChannelFunc, Injectable seams, logArgs(), confirmAction(), confirmDelete(), lineSink(), confirmImport() (+99 more)

### Community 1 - "newTestOps"
Cohesion: 0.08
Nodes (67): broker_test.go, countContains(), leaderCalls(), newTestOps(), TestCountContains(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs() (+59 more)

### Community 2 - "context.Context"
Cohesion: 0.07
Nodes (13): runtime_test.go, Cluster, normalizeToList(), HARoles(), RestartOrder(), stsName(), bareKind(), ownedSecretNames() (+5 more)

### Community 3 - "kubectlTransport"
Cohesion: 0.33
Nodes (5): BrokerPodSuffixShape(), podName(), TestPodNameSuffixMatchesTheConfigBound(), shSingleQuote(), kubectlTransport

### Community 4 - "cli_test.go"
Cohesion: 0.05
Nodes (127): leaderRun, cli_test.go, Fixtures and doubles, Shared env fixtures, allowRuntime(), captureStderr(), collectPaths(), echoRunner() (+119 more)

### Community 5 - "config_test.go"
Cohesion: 0.07
Nodes (65): config_test.go, assertContainerBlockDefaults(), assertContainerScaling(), envTree(), haNodesConfig(), TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes(), TestAdditionalUserNamesCollideOnDocker() (+57 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (50): Commands, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker cli, solace-util broker configure (+42 more)

### Community 7 - "Manager"
Cohesion: 0.09
Nodes (14): runtime_test.go, certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary() (+6 more)

### Community 8 - "K8sConfig"
Cohesion: 0.11
Nodes (15): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, LoadBalancer, cut() (+7 more)

### Community 9 - "completion_test.go"
Cohesion: 0.22
Nodes (24): completion_test.go, driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion() (+16 more)

### Community 10 - "Command"
Cohesion: 0.12
Nodes (20): cmdField, guardedCmd, execguard_test.go, Command, clusterRules(), composeRules(), rulesFor(), runtimeRules() (+12 more)

### Community 11 - "deploy_test.go"
Cohesion: 0.17
Nodes (25): deploy_test.go, brokerJSON(), logBuf(), TestConfirmBrokerAppliedAcceptsTheObjectItApplied(), TestConfirmBrokerAppliedFailsOnADifferentObject(), TestConfirmBrokerAppliedFailsWhenTheObjectIsNotThere(), TestConfirmBrokerAppliedSurfacesReadFailures(), TestConfirmBrokerAppliedTreatsSilenceAsSkipped() (+17 more)

### Community 12 - "NewCluster"
Cohesion: 0.09
Nodes (65): operator_test.go, GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), RenderOperator(), splitAfterNamespace(), splitOperatorBundle() (+57 more)

### Community 13 - "k8s/inspect.go"
Cohesion: 0.08
Nodes (30): orNone(), age(), anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), roleRank(), discoveryServiceName() (+22 more)

### Community 14 - "convert_test.go"
Cohesion: 0.07
Nodes (67): doc, Result, segment, vars, The keys, convert_test.go, boolOf(), commentSafe() (+59 more)

### Community 15 - "transform_test.go"
Cohesion: 0.10
Nodes (61): Block, TargetState, transform_test.go, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects() (+53 more)

### Community 16 - "manager_test.go"
Cohesion: 0.11
Nodes (80): manager_test.go, GuardPodmanEUID(), assertMode(), containsStr(), ctrCfg(), hasCall(), maskedKeys(), newCapMgr() (+72 more)

### Community 17 - "commands.go"
Cohesion: 0.13
Nodes (72): opFunc, roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd() (+64 more)

### Community 18 - "coverage_test.go"
Cohesion: 0.06
Nodes (69): coverage_test.go, internal/broker, verify_ops_test.go, matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError() (+61 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "internal/broker"
Cohesion: 0.16
Nodes (14): config_ops_test.go, TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(), TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator(), TestExecShellScriptRefusesAnUnsafeNameBeforeUploading() (+6 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Load"
Cohesion: 0.08
Nodes (45): scalingTier, scaling_test.go, minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs() (+37 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.05
Nodes (92): CLIRunner, fakeRun, PhaseKind, SiteState, SwitchAction, SwitchPhase, matechannel_test.go, sempmate_test.go (+84 more)

### Community 24 - "newLocalOps"
Cohesion: 0.07
Nodes (68): semp_test.go, verify_local_test.go, bridgeHostPort(), sempPort(), sempV1OK(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA() (+60 more)

### Community 25 - "runner_test.go"
Cohesion: 0.17
Nodes (26): runner_test.go, TestChildEnvNamesAreNotSystemVariables(), MaskEnv(), quoteTok(), captureStdout(), helperCommand(), TestEchoDefaultWriter(), TestEchoOutput() (+18 more)

### Community 26 - "renderDriver"
Cohesion: 0.15
Nodes (22): chunk, chunkResult, driver_test.go, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce() (+14 more)

### Community 27 - "testing.T"
Cohesion: 0.05
Nodes (96): command_test.go, commanddoc_test.go, examples_test.go, exit_test.go, namespace_test.go, platform_test.go, statusreport_test.go, storage_test.go (+88 more)

### Community 28 - "importops_test.go"
Cohesion: 0.07
Nodes (71): ImportResult, PlannedSection, importops_test.go, defaultVPNFirst(), describeScope(), ImportPlan, importIgnore(), orNone() (+63 more)

### Community 29 - "watch_test.go"
Cohesion: 0.08
Nodes (34): watch_test.go, canIResource(), canIVerb(), desiredWatch(), subtractWatch(), deniedCluster(), deniedRunner(), deployJSON() (+26 more)

### Community 31 - ".ApplyDefaults"
Cohesion: 0.14
Nodes (16): tlssecret_test.go, applyContainerBlockDefaults(), podmanRunUser(), setDefault(), setDefaultCmd(), setDefaultInt(), xdgConfigHome(), containerMem() (+8 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.13
Nodes (13): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, Auditing a project against this, Container security guidelines, internal/broker (+5 more)

### Community 33 - "NewCluster"
Cohesion: 0.11
Nodes (42): checkreport_test.go, preflight_test.go, TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestCheckReportSkipsEmptySections(), TestOperatorInstallVerdict(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestOperatorValidateAsksNothingAboutTheBroker(), TestPortRowsNeverFail() (+34 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.05
Nodes (55): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domaincerts_resolve_test.go, domaincerts_test.go, caNameSafe() (+47 more)

### Community 35 - "watchLoop"
Cohesion: 0.32
Nodes (5): watchInterval(), stdoutIsTerminal(), watchLabel(), watchLoop(), withWatch()

### Community 36 - "newRootCmd"
Cohesion: 0.12
Nodes (21): aliases_test.go, applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias() (+13 more)

### Community 37 - "Exec"
Cohesion: 0.22
Nodes (11): Keys that were renamed, resolve_test.go, Exec, TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveMissingBinaryIsActionable(), TestResolveRefusesCurrentDirectory() (+3 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.07
Nodes (49): Region, annotate_test.go, applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV() (+41 more)

### Community 40 - "output_test.go"
Cohesion: 0.06
Nodes (65): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, diff_test.go, internal/output, output_test.go (+57 more)

### Community 41 - "Compose"
Cohesion: 0.12
Nodes (24): internal/render, servercert_test.go, NodeIdentity, Compose(), composeEscape(), ContainerSecrets(), containerSecretSpecs(), EnvPairs() (+16 more)

### Community 42 - "inject_test.go"
Cohesion: 0.17
Nodes (28): lineRole, serviceLine, shutdownStyle, svcKey, inject_test.go, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "secrets_test.go"
Cohesion: 0.18
Nodes (27): secrets_test.go, ResolveSecretValues(), bundleHash(), certCreate(), certFixture(), TestCertSecretLabelDrivesTheRestart(), TestContainerBundleCarriesTheChain(), TestDeletePodmanToleratesAMissingBundle() (+19 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.13
Nodes (30): exportconfigReadCounter, opRunner, euid_test.go, exportconfig_test.go, internal/cli, runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand() (+22 more)

### Community 47 - "VPNRepl"
Cohesion: 0.06
Nodes (20): AdminState, QueueState, scriptedMate, sempMate, replMate, ReplRole, VPNRepl, ReplicationConfigResult (+12 more)

### Community 48 - "Echo"
Cohesion: 0.23
Nodes (4): interactiveFailRunner, runFailRunner, Echo, Quote()

### Community 49 - "config_test.go"
Cohesion: 0.17
Nodes (19): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, showCmd, span, AdditionalUser, keyValueEntries (+11 more)

### Community 50 - "containerState"
Cohesion: 0.27
Nodes (5): containerState, healthState, printable(), TestPrintableStripsControlCharacters(), KV

### Community 51 - "newEchoMgr"
Cohesion: 0.20
Nodes (10): NewManager(), newEchoMgr(), TestManagerCheckDryRun(), TestManagerDeployDockerDryRunMasksSecretEnv(), TestManagerDeployPodmanDryRunHidesSecretBytes(), TestManagerDeployPodmanDryRunSkipsWrite(), TestManagerDockerCheckProbesCompose(), TestManagerLifecycleDockerDryRunUsesCompose() (+2 more)

### Community 52 - "scripts.go"
Cohesion: 0.06
Nodes (39): scripts_test.go, validCLILine(), TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), archiveConfigsScript(), assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript() (+31 more)

### Community 53 - "haCfg"
Cohesion: 0.04
Nodes (108): TLS, adminsecret_test.go, check_test.go, cluster_test.go, internal/k8s, names_test.go, ops_test.go, prep_test.go (+100 more)

### Community 54 - ".runMatePhase"
Cohesion: 0.36
Nodes (6): isRunCLIRejection(), newlineIf(), replPhase1Rejected(), replPhase2Rejected(), replTransportFailure(), replVPNList()

### Community 55 - ".dump"
Cohesion: 0.09
Nodes (71): opCall, ops_k8s_test.go, Per-package doubles, capture(), captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod() (+63 more)

### Community 56 - "Abbreviations"
Cohesion: 0.29
Nodes (6): Abbreviations, Flag shorthands, How a short form is resolved, Node roles, Platforms, Reading this reference

### Community 57 - "replication_test.go"
Cohesion: 0.08
Nodes (50): replRig, Commands, solace-util broker, Configuring a site, Data replication, Permissions, Reaching the mate, Switching roles (+42 more)

### Community 58 - ".OperatorApply"
Cohesion: 0.20
Nodes (6): operatorEscalationProbes(), operatorProbes(), canIAnswer(), canIRefused(), missingNamespace(), probe

### Community 59 - "ReplSite"
Cohesion: 0.07
Nodes (58): Replication, ReplVia, ReplViaKube, ReplViaSEMP, secretRef, matechannel_test.go, replication_test.go, CLIArg() (+50 more)

### Community 60 - "Command reference"
Cohesion: 0.50
Nodes (4): Command reference, Global flags, Index, Tree

### Community 61 - "ParseBlocks"
Cohesion: 0.08
Nodes (41): Capture, Omission, token, blocks_test.go, omit_test.go, Annotate(), TestAnnotateIsDeterministic(), TestAnnotateRoundTripCaptureIsMarked() (+33 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.19
Nodes (40): rootless_test.go, failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+32 more)

### Community 63 - "Test catalogue"
Cohesion: 0.33
Nodes (6): internal/convert, internal/engine, internal/examples, internal/tools/vulnjudge, Running the tests, Test catalogue

### Community 64 - "Platform"
Cohesion: 0.17
Nodes (21): checkFlagPlatforms(), commandPlatforms(), declaredList(), parsePlatformList(), prepare(), promptPlatform(), resolvePlatform(), supportsPlatform() (+13 more)

### Community 65 - "Quadlet"
Cohesion: 0.24
Nodes (11): opCtrGenArtifact(), ContainerNoFile(), HealthCheck, TestLimitsCheckAssertsWhatTheArtifactAsks(), escapePercent(), healthCmd(), Quadlet(), quadletEscape() (+3 more)

### Community 66 - "transcript_test.go"
Cohesion: 0.40
Nodes (6): transcript_test.go, stripTranscript(), TestTranscriptBannerType(), TestTranscriptPreambleForApply(), TestTranscriptStripCutsAtBegin(), TestTranscriptTeardownFixtureResetsNotDeletes()

### Community 67 - "opPerformReplication"
Cohesion: 0.08
Nodes (24): layer, PlannedRoles(), confirmActionStrict(), confirmRestart(), confirmScope(), step(), warn(), newConvertCmd() (+16 more)

### Community 68 - "The eight rules"
Cohesion: 0.22
Nodes (9): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+1 more)

### Community 70 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 71 - "redundancy_test.go"
Cohesion: 0.21
Nodes (12): redundancy_test.go, pskCfg(), TestFillStandaloneNodeNameIgnoresAnEmptyHostname(), TestFillStandaloneNodeNameKeepsAConfiguredName(), TestFillStandaloneNodeNameNeverFillsHA(), TestFillStandaloneNodeNameUsesTheHost(), TestNothingGeneratesThePSK(), TestPSKEnvSatisfiesTheContainerRequirement() (+4 more)

### Community 73 - ".Restart"
Cohesion: 0.26
Nodes (12): internal/container, preflight_test.go, fileExists(), TestManagerLifecycleDockerComposeFile(), TestManagerLifecycleDockerNoComposeFile(), TestManagerLifecyclePodmanSystemctl(), TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestComposeSecretEnvNamesCannotBeSystemVars() (+4 more)

### Community 74 - "Role"
Cohesion: 0.05
Nodes (31): covUploadFailsTo, downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile (+23 more)

### Community 75 - "Coverage"
Cohesion: 0.40
Nodes (4): Coverage, ValidScriptName(), defaultLocalAddrs(), TestDefaultLocalAddrs()

### Community 78 - "cli/watch_test.go"
Cohesion: 0.17
Nodes (23): allowcommand_test.go, watch_test.go, TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRegisteredWhereItExecutes(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd() (+15 more)

### Community 79 - "adminsecret_test.go"
Cohesion: 0.21
Nodes (10): adminsecret_test.go, adminCfg(), TestAdminSecretNameFollowsTheStates(), TestAdminSecretRefusesTheOperatorsOwnNames(), TestContainersStillRequireTheAdminPassword(), TestKubernetesValidatesWithoutAnAdminPassword(), TestMonitorPasswordAndKeyNeedTheAdminPassword(), TestOperatorAdminSecretName() (+2 more)

### Community 80 - "containerTransport"
Cohesion: 0.13
Nodes (10): containerTransport, Contract, Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams, shSingleQuote() (+2 more)

### Community 81 - "replicationops_test.go"
Cohesion: 0.16
Nodes (35): replicationops_test.go, hasCall(), reenableLine(), replicationVPNLines(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching() (+27 more)

### Community 82 - ".validateContainer"
Cohesion: 0.10
Nodes (14): TestRedundancyEnabled(), cpuSetCount(), cpuSetRange(), TestCPUSetCount(), TestCPUSetRange(), foldToEnvVar(), invertedCPUSetRange(), missingErr() (+6 more)

### Community 83 - "limits_test.go"
Cohesion: 0.08
Nodes (55): limits_test.go, main_test.go, missingControllers(), limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessCoreFloorFollowsTheConfiguredLimit() (+47 more)

### Community 84 - "Developer guide"
Cohesion: 0.11
Nodes (20): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+12 more)

### Community 85 - ".BackupRevertActivity"
Cohesion: 0.13
Nodes (15): backupTarget, Credential, transport_test.go, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigFlag(), curlConfigLine(), httpBody() (+7 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.22
Nodes (20): inspect_test.go, podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture() (+12 more)

### Community 87 - "replication_test.go"
Cohesion: 0.06
Nodes (71): BrokerType, cliMate, colSpan, replication_test.go, bannerType(), cliTransport(), containsEndpoint(), dashSpans() (+63 more)

### Community 88 - "AdminSecret"
Cohesion: 0.11
Nodes (36): secrets_test.go, GenSecrets(), TestGenSecretsStopsAtASecretItCannotBuild(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), operatorRegcred() (+28 more)

### Community 89 - "render_test.go"
Cohesion: 0.12
Nodes (49): render_test.go, boolStr(), BrokerCR(), ComposeProject(), assertNoCheckoutPath(), containerArtifacts(), envLines(), healthCheckFixture() (+41 more)

### Community 90 - "checkreport_test.go"
Cohesion: 0.05
Nodes (58): solace-util, solace-util operator, A replication switch refuses before changing anything, A wrapper runtime is refused, Bringing up a fresh cluster, Broker scope is a fixed classification, Docker and Podman mechanics, Docker compose secrets need compose 2.23.1+ (+50 more)

### Community 91 - "RuleFor"
Cohesion: 0.12
Nodes (39): Disposition, SectionRule, appliance_test.go, importdoc_test.go, sections_test.go, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent() (+31 more)

### Community 93 - "internal/config"
Cohesion: 0.10
Nodes (22): artifactvalues_test.go, duration_test.go, imagepullsecret_test.go, internal/config, portOrRange(), TestValidateContainerArtifactValues(), TestValidateImageReference(), TestValidateK8sArtifactValues() (+14 more)

### Community 95 - ".report"
Cohesion: 0.10
Nodes (8): delegateSetting(), limitText(), pamCoreValue(), parseLimit(), parseUnitProps(), idMapCovers(), origin(), runUserIDs()

### Community 96 - "gen-cert.sh"
Cohesion: 0.27
Nodes (13): ask(), askPass(), askSan(), askYesNo(), buildSanList(), checkExit(), concatCnIfNotBlank(), echoUsage() (+5 more)

### Community 98 - "Container"
Cohesion: 0.27
Nodes (6): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestContainerCoreLimitSpellings()

### Community 99 - "Configuration"
Cohesion: 0.17
Nodes (12): Bring your own TLS Secret, Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling, Scaling tiers (+4 more)

### Community 102 - "Platforms"
Cohesion: 0.10
Nodes (35): examples_test.go, platform_test.go, Example, TestPlatformConstantsMatchSchemaSections(), TestGuardConfigIsValid(), TestDomainCertsDirsAreNotDefaulted(), TestValidateHostPathsMatchesTheRebaseList(), parseError() (+27 more)

### Community 103 - ".decodeScalingEntry"
Cohesion: 0.16
Nodes (12): scalingKey, scalingSpelling, Scaling, scalingKeyIndex(), scalingKeyList(), generatorDefaults(), scalingOf(), genBlock (+4 more)

### Community 112 - "capRunner"
Cohesion: 0.14
Nodes (19): capCall, capRunner, New(), callIndex(), TestManagerLogsCLIShell(), mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged() (+11 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.14
Nodes (20): inspect_test.go, decodeInspect(), orUnknown(), inspectReply(), kvValue(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings() (+12 more)

### Community 126 - "operatorversion_test.go"
Cohesion: 0.19
Nodes (19): operatorversion_test.go, compareVersions(), imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), operatorDeployJSON(), TestCompareVersions() (+11 more)

### Community 128 - "hostpath_test.go"
Cohesion: 0.08
Nodes (33): hostpath_test.go, guardConfig(), BaseName(), CheckHostPath(), expandTilde(), expandTildeToken(), HasPathSeparator(), IsAbsHostPath() (+25 more)

### Community 155 - "strings.Builder"
Cohesion: 0.08
Nodes (36): shorthand, WeightedNodeTerm, mdRow(), renderAbbrevDocs(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable() (+28 more)

## Knowledge Gaps
- **13 isolated node(s):** `solace`, `NO_COLOR`, `internal/broker`, `internal/cli`, `internal/config` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 184 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **7 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `App` to `Platform`, `Quadlet`, `context.Context`, `watchLoop`, `opPerformReplication`, `cli_test.go`, `newRootCmd`, `Manager`, `ReplSite`, `exportconfig_test.go`, `containerTransport`, `commands.go`, `config_test.go`, `haCfg`, `.dump`, `replication_test.go`, `testing.T`?**
  _High betweenness centrality (0.037) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `newTestOps()` (e.g. with `importops_test.go` and `Injectable seams`) actually correct?**
  _`newTestOps()` has 117 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `NO_COLOR`, `internal/broker` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.06013048845000882 - nodes in this community are weakly interconnected._
- **Why does `Config` connect `haCfg` to `App`, `hostpath_test.go`, `newTestOps`, `context.Context`, `cli_test.go`, `config_test.go`, `kubectlTransport`, `Manager`, `K8sConfig`, `Command`, `NewCluster`, `k8s/inspect.go`, `convert_test.go`, `manager_test.go`, `Load`, `newLocalOps`, `testing.T`, `strings.Builder`, `watch_test.go`, `.ApplyDefaults`, `NewCluster`, `ResolveDomainCerts`, `Compose`, `secrets_test.go`, `config_test.go`, `newEchoMgr`, `.dump`, `ReplSite`, `Test catalogue`, `Platform`, `Quadlet`, `redundancy_test.go`, `Role`, `Coverage`, `Image`, `adminsecret_test.go`, `containerTransport`, `.validateContainer`, `AdminSecret`, `render_test.go`, `checkreport_test.go`, `internal/config`, `Container`, `.decodeScalingEntry`, `capRunner`?**
  _High betweenness centrality (0.036) - this node is a cross-community bridge._
- **Are the 9 inferred relationships involving `Config` (e.g. with `config_test.go` and `execguard_test.go`) actually correct?**
  _`Config` has 9 INFERRED edges - model-reasoned connections that need verification._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.08074534161490683 - nodes in this community are weakly interconnected._
# Graph Report - solace-cnt-scripts  (2026-10-28)

## Corpus Check
- 204 files · ~707,024 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .golden 17, (none) 3, .cli 2)

## Summary
- 4074 nodes · 20789 edges · 103 communities (97 shown, 6 thin omitted)
- Extraction: 72% EXTRACTED · 28% INFERRED · 0% AMBIGUOUS · INFERRED: 5793 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `d4f4a51b`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- App
- newTestOps
- scaling_test.go
- podName
- cli_test.go
- config_test.go
- Command Reference Docs
- Manager
- K8sConfig
- testing.T
- CheckCommand
- runLeader
- NewCluster
- BrokerType
- convert_test.go
- transform_test.go
- manager_test.go
- commands.go
- coverage_test.go
- Bash Dev Script
- Load
- PowerShell Dev Script
- Set
- switchplan_test.go
- newLocalOps
- runner_test.go
- renderDriver
- statusreport_test.go
- importops_test.go
- deploymentItem
- Solace Module Root
- Platform
- CLAUDE.md
- internal/config
- ResolveDomainCerts
- watchLoop
- github.com/spf13/cobra.ShellCompDirective
- NewCluster
- annotate_test.go
- Sink
- Compose
- inject_test.go
- Block
- exportconfig_test.go
- VPNRepl
- localCfg
- config_test.go
- eqArgs
- cli/platform_test.go
- Ops
- prep_test.go
- .validate
- .dump
- context.Context
- replication_test.go
- .OperatorApply
- adminsecret_test.go
- confirm.go
- .configRows
- rootless_test.go
- exit_test.go
- Platforms
- Coverage
- opPerformReplication
- ParseRole
- The eight rules
- EnvPairs
- What `import-config` applies
- transcript_test.go
- Configuration
- Command reference
- validContainerConfig
- newRootCmd
- Abbreviations
- cli/watch_test.go
- replicationops_test.go
- replCfg
- limits_test.go
- Developer guide
- k8s/inspect.go
- .ConfigureReplication
- AdminSecret
- render_test.go
- Image
- RuleFor
- Container
- replConfig
- DetectPlatforms
- operatorversion_test.go
- Config
- config_ops_test.go
- capRunner
- certFixture
- Container State Inspection
- Role
- bytes.Buffer
- hostpath_test.go
- blocks.go
- strings.Builder

## God Nodes (most connected - your core abstractions)
1. `App` - 217 edges
2. `newTestOps()` - 167 edges
3. `cli_test.go` - 148 edges
4. `Config` - 146 edges
5. `NewCluster()` - 141 edges
6. `ctrCfg()` - 133 edges
7. `Role` - 132 edges
8. `Per-package doubles` - 126 edges
9. `loadK8s()` - 119 edges
10. `newCapMgr()` - 114 edges

## Surprising Connections (you probably didn't know these)
- `Repository layout` --references--> `Execute()`  [INFERRED]
  docs/developer.md → internal/cli/root.go
- `The preflight` --references--> `info()`  [INFERRED]
  docs/operations.md → internal/k8s/checkreport.go
- `Per-package doubles` --references--> `sample()`  [INFERRED]
  docs/test.md → internal/abbrev/abbrev_test.go
- `annotate_test.go` --references--> `Annotate()`  [INFERRED]
  docs/test.md → internal/broker/annotate.go
- `annotate_test.go` --references--> `TestAnnotateRoundTripPreservesBlocks()`  [INFERRED]
  docs/test.md → internal/broker/annotate_test.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (103 total, 6 thin omitted)

### Community 0 - "App"
Cohesion: 0.07
Nodes (94): App, logArgs(), confirmAction(), lineSink(), runExport(), runImport(), diagnosticsOptions(), exportScope() (+86 more)

### Community 1 - "newTestOps"
Cohesion: 0.07
Nodes (72): broker_test.go, cliRunNames(), countContains(), runCLISkeleton(), leaderCalls(), newTestOps(), TestCountContains(), TestDiagnostics() (+64 more)

### Community 2 - "scaling_test.go"
Cohesion: 0.06
Nodes (34): scalingTier, Node roles, Reading this reference, scaling_test.go, containerMem(), cpuSetCount(), cpuSetRange(), clearTierDerived() (+26 more)

### Community 3 - "podName"
Cohesion: 0.16
Nodes (10): allCustomMounted(), discoveryServiceName(), HARoles(), lbServiceName(), podName(), pvcName(), stsName(), TestResourceNames() (+2 more)

### Community 4 - "cli_test.go"
Cohesion: 0.07
Nodes (111): cli_test.go, Shared env fixtures, allowRuntime(), captureStderr(), collectPaths(), echoRunner(), fakeBinaryOnPath(), firstLine() (+103 more)

### Community 5 - "config_test.go"
Cohesion: 0.09
Nodes (50): config_test.go, assertContainerBlockDefaults(), assertContainerScaling(), envTree(), haNodesConfig(), TestAdditionalUserNameIsStricterOnKubernetes(), TestAdditionalUserNamesDoNotCollideOnKubernetes(), TestAdditionalUserPasswordCharsAreFreeOnKubernetes() (+42 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (50): Commands, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker cli, solace-util broker configure (+42 more)

### Community 7 - "Manager"
Cohesion: 0.07
Nodes (18): internal/container, preflight_test.go, certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle() (+10 more)

### Community 8 - "K8sConfig"
Cohesion: 0.12
Nodes (13): ContainerSecurity, Operator, PodSecurity, Resources, Storage, WeightedNodeTerm, K8sConfig, NodeAffinity (+5 more)

### Community 9 - "testing.T"
Cohesion: 0.06
Nodes (79): allowcommand_test.go, command_test.go, completion_test.go, examples_test.go, namespace_test.go, storage_test.go, TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRegisteredWhereItExecutes() (+71 more)

### Community 10 - "CheckCommand"
Cohesion: 0.11
Nodes (31): cmdField, commandRules, guardedCmd, execguard_test.go, decodeCommand(), Command, checkBinary(), CheckCommand() (+23 more)

### Community 11 - "runLeader"
Cohesion: 0.60
Nodes (4): leaderRun, runLeader(), TestAssertLeaderAsksBeforeActingFromABackup(), TestAssertLeaderPrimaryAndMonitorNeverAsk()

### Community 12 - "NewCluster"
Cohesion: 0.11
Nodes (47): operator_test.go, GenOperator(), operatorImage(), RenderOperator(), boolPtr(), bundledDefaultPorts(), commentedExamplePorts(), covSwapBundle() (+39 more)

### Community 13 - "BrokerType"
Cohesion: 0.25
Nodes (5): BrokerType, cliMate, bannerType(), showReplicationScript(), showVPNReplicationScript()

### Community 14 - "convert_test.go"
Cohesion: 0.07
Nodes (69): doc, Result, segment, vars, The keys, convert_test.go, DecodeStrict(), boolOf() (+61 more)

### Community 15 - "transform_test.go"
Cohesion: 0.13
Nodes (37): transform_test.go, InjectVPNServiceShutdown(), nestedCreateRE(), readTargetState(), TargetOnlyRemovals(), covAllLines(), covBlock(), covNestedBlocks() (+29 more)

### Community 16 - "manager_test.go"
Cohesion: 0.09
Nodes (94): manager_test.go, GuardPodmanEUID(), fileExists(), assertMode(), containsStr(), ctrCfg(), hasCall(), maskedKeys() (+86 more)

### Community 17 - "commands.go"
Cohesion: 0.13
Nodes (73): opFunc, roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd() (+65 more)

### Community 18 - "coverage_test.go"
Cohesion: 0.07
Nodes (68): coverage_test.go, verify_ops_test.go, matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle() (+60 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "Load"
Cohesion: 0.16
Nodes (26): minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors(), TestLoadSecurityIDsAsWritten() (+18 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Set"
Cohesion: 0.14
Nodes (17): Entry, abbrev_test.go, commanddoc_test.go, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies() (+9 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.07
Nodes (73): CLIRunner, fakeRun, PhaseKind, SwitchAction, SwitchPhase, matechannel_test.go, switchplan_test.go, MateChannel (+65 more)

### Community 24 - "newLocalOps"
Cohesion: 0.16
Nodes (33): verify_local_test.go, cliScriptPath(), isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn(), TestDetectRoleAddrsError() (+25 more)

### Community 25 - "runner_test.go"
Cohesion: 0.08
Nodes (42): interactiveFailRunner, Keys that were renamed, internal/engine, resolve_test.go, runner_test.go, Exec, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod() (+34 more)

### Community 26 - "renderDriver"
Cohesion: 0.15
Nodes (22): chunk, chunkResult, driver_test.go, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce() (+14 more)

### Community 27 - "statusreport_test.go"
Cohesion: 0.19
Nodes (26): statusreport_test.go, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportOmitsWorkloadsThatAreNotTheBrokers(), TestBrokerReportOrdersPodsOfTheSameRoleByName(), TestBrokerReportPreviewStopsAtAnUnapprovedCommand(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportReadsOnlyWhatTheEnvFileNames() (+18 more)

### Community 28 - "importops_test.go"
Cohesion: 0.06
Nodes (64): ImportResult, PlannedSection, importops_test.go, isPragma(), defaultVPNFirst(), describeScope(), ImportPlan, importIgnore() (+56 more)

### Community 29 - "deploymentItem"
Cohesion: 0.23
Nodes (9): compareVersions(), imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), watchFromDeployment(), deploymentItem, deploymentList (+1 more)

### Community 31 - "Platform"
Cohesion: 0.12
Nodes (24): runConvert(), usagef(), checkAllowCommand(), checkFlagPlatforms(), commandPlatforms(), declaredList(), parsePlatformList(), prepare() (+16 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "internal/config"
Cohesion: 0.14
Nodes (18): duration_test.go, internal/abbrev, internal/config, internal/convert, internal/tools/vulnjudge, Running the tests, Summary, Test catalogue (+10 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (54): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domaincerts_resolve_test.go, domaincerts_test.go, caNameSafe() (+46 more)

### Community 35 - "watchLoop"
Cohesion: 0.32
Nodes (5): watchInterval(), stdoutIsTerminal(), watchLabel(), watchLoop(), withWatch()

### Community 36 - "github.com/spf13/cobra.ShellCompDirective"
Cohesion: 0.33
Nodes (6): completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), isEnvFileName(), matching()

### Community 37 - "NewCluster"
Cohesion: 0.08
Nodes (74): checkreport_test.go, deploy_test.go, preflight_test.go, TestCheckStopsProbingWhenUnreachable(), TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestOperatorInstallVerdict(), TestOperatorInstallVerdictShowsOnceInEveryView(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne() (+66 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.08
Nodes (44): annotate_test.go, applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV(), readMarker() (+36 more)

### Community 40 - "Sink"
Cohesion: 0.05
Nodes (74): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, diff_test.go, internal/output, output_test.go (+66 more)

### Community 41 - "Compose"
Cohesion: 0.20
Nodes (14): internal/render, servercert_test.go, Compose(), composeEscape(), ContainerSecrets(), ContainerSecret, SecretPreflight(), TestComposeEscapesTheDollarSign() (+6 more)

### Community 42 - "inject_test.go"
Cohesion: 0.17
Nodes (28): lineRole, serviceLine, shutdownStyle, svcKey, inject_test.go, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "Block"
Cohesion: 0.17
Nodes (28): Block, Region, TargetState, regionFor(), keepRegion(), injectedBlock(), BridgeEnablementInverted(), ClearExistingNested() (+20 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.11
Nodes (34): exportconfigReadCounter, opRunner, euid_test.go, exportconfig_test.go, internal/cli, runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand() (+26 more)

### Community 47 - "VPNRepl"
Cohesion: 0.08
Nodes (13): AdminState, QueueState, scriptedMate, sempMate, SiteState, replMate, ReplRole, VPNRepl (+5 more)

### Community 48 - "localCfg"
Cohesion: 0.06
Nodes (64): backupTarget, Credential, semp_test.go, sempmate_test.go, transport_test.go, bridgeHostPort(), curlConfigFlag(), curlConfigLine() (+56 more)

### Community 49 - "config_test.go"
Cohesion: 0.17
Nodes (23): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, showCmd, span, scalingKey, scalingSpelling (+15 more)

### Community 50 - "eqArgs"
Cohesion: 0.11
Nodes (41): check_test.go, cluster_test.go, internal/k8s, ops_test.go, transport_test.go, TestReachable(), TestResolveStorageClass(), newCluster() (+33 more)

### Community 51 - "cli/platform_test.go"
Cohesion: 0.21
Nodes (25): platform_test.go, findCmd(), TestEveryRunnableCommandIsWired(), logArgs2(), runPlatform(), TestLogArgsBuildsOneSetForBothPlatforms(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer() (+17 more)

### Community 52 - "Ops"
Cohesion: 0.05
Nodes (50): scripts_test.go, field(), Ops, validCLILine(), validName(), ValidScriptName(), TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), archiveConfigsScript() (+42 more)

### Community 53 - "prep_test.go"
Cohesion: 0.09
Nodes (52): names_test.go, prep_test.go, BrokerPodSuffixShape(), ProductKeyRoles(), RestartOrder(), haCfg(), TestHARoles(), TestPodNameSuffixMatchesTheConfigBound() (+44 more)

### Community 54 - ".validate"
Cohesion: 0.07
Nodes (29): A replication switch refuses before changing anything, A wrapper runtime is refused, Bringing up a fresh cluster, Broker scope is a fixed classification, Docker compose secrets need compose 2.23.1+, Every destructive command confirms, Exit codes, Exporting and importing configuration (+21 more)

### Community 55 - ".dump"
Cohesion: 0.09
Nodes (72): opCall, replRig, Fixtures and doubles, ops_k8s_test.go, Per-package doubles, capture(), captureStdout(), failDisableDefaultUsersUpload() (+64 more)

### Community 56 - "context.Context"
Cohesion: 0.06
Nodes (10): runtime_test.go, Runner, Cluster, normalizeToList(), bareKind(), ownedSecretNames(), TestClusterHonoursRuntime(), badJSONRunner (+2 more)

### Community 57 - "replication_test.go"
Cohesion: 0.09
Nodes (49): Commands, solace-util broker, Configuring a site, Data replication, Permissions, Reaching the mate, Switching roles, The limits the container actually gets (+41 more)

### Community 58 - ".OperatorApply"
Cohesion: 0.16
Nodes (11): joinYAMLDocs(), nonEmpty(), operatorEscalationProbes(), operatorProbes(), splitAfterNamespace(), splitOperatorBundle(), canIAnswer(), canIRefused() (+3 more)

### Community 59 - "adminsecret_test.go"
Cohesion: 0.21
Nodes (10): adminsecret_test.go, adminCfg(), TestAdminSecretNameFollowsTheStates(), TestAdminSecretRefusesTheOperatorsOwnNames(), TestContainersStillRequireTheAdminPassword(), TestKubernetesValidatesWithoutAnAdminPassword(), TestMonitorPasswordAndKeyNeedTheAdminPassword(), TestOperatorAdminSecretName() (+2 more)

### Community 60 - "confirm.go"
Cohesion: 0.23
Nodes (15): layer, Injectable seams, confirmActionStrict(), confirmDowngrade(), confirmGate(), confirmRestart(), interactive(), promptLine() (+7 more)

### Community 61 - ".configRows"
Cohesion: 0.09
Nodes (33): solace-util, solace-util operator, Docker and Podman mechanics, Re-deploying is safe and explicit, Rootless podman prerequisites, What the broker cannot gain, Which user the broker runs as, orNone() (+25 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.19
Nodes (40): rootless_test.go, failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+32 more)

### Community 63 - "exit_test.go"
Cohesion: 0.16
Nodes (23): runFailRunner, exit_test.go, TestExecute(), childExit(), ExitCode(), isUsage(), childStatusError(), k8sEnv() (+15 more)

### Community 64 - "Platforms"
Cohesion: 0.17
Nodes (15): platform_test.go, TestExecBinariesCoversEveryPlatform(), TestGuardConfigIsValid(), TestDomainCertsDirsAreNotDefaulted(), newPlatforms(), PlatformAbbrev(), Platforms(), TestDetectPlatforms() (+7 more)

### Community 65 - "Coverage"
Cohesion: 0.18
Nodes (10): Coverage, defaultLocalAddrs(), TestDefaultLocalAddrs(), HealthCheck, healthCmd(), healthCheckFixture(), TestHealthCmdDefaultsToReadiness(), TestQuadletHealthCmdEscapesPercent() (+2 more)

### Community 66 - "opPerformReplication"
Cohesion: 0.17
Nodes (11): mateChannelFunc, step(), confirmReplicationConfig(), confirmReplicationSwitch(), opConfigureReplication(), opCtrConfigureReplication(), opCtrPerformReplication(), opPerformReplication() (+3 more)

### Community 67 - "ParseRole"
Cohesion: 0.17
Nodes (8): TestParseRole(), TestRedundancyEnabled(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), TestRoleNames(), ParseRole(), RoleAbbrev(), RoleNames()

### Community 68 - "The eight rules"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 69 - "EnvPairs"
Cohesion: 0.13
Nodes (21): opCtrGenArtifact(), ContainerNoFile(), NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), containerSecretSpecs(), EnvPairs(), escapePercent(), groupKey() (+13 more)

### Community 70 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 71 - "transcript_test.go"
Cohesion: 0.50
Nodes (5): transcript_test.go, stripTranscript(), TestTranscriptBannerType(), TestTranscriptStripCutsAtBegin(), TestTranscriptTeardownFixtureResetsNotDeletes()

### Community 72 - "Configuration"
Cohesion: 0.17
Nodes (12): Bring your own TLS Secret, Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling, Scaling tiers (+4 more)

### Community 73 - "Command reference"
Cohesion: 0.50
Nodes (4): Command reference, Global flags, Index, Tree

### Community 74 - "validContainerConfig"
Cohesion: 0.10
Nodes (28): artifactvalues_test.go, redundancy_test.go, TestValidateContainerArtifactValues(), TestValidateImageReference(), TestValidateK8sArtifactValues(), validPlatformConfig(), TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNamesCollideOnDocker() (+20 more)

### Community 75 - "newRootCmd"
Cohesion: 0.09
Nodes (25): childExitError, usageError, How a short form is resolved, aliases_test.go, applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive() (+17 more)

### Community 76 - "Abbreviations"
Cohesion: 0.67
Nodes (3): Abbreviations, Flag shorthands, Platforms

### Community 78 - "cli/watch_test.go"
Cohesion: 0.35
Nodes (11): watch_test.go, stopAfter(), TestOperatorStatusWatches(), TestStatusWithoutWatchRunsOnce(), TestWatchAppendsWhenNotATerminal(), TestWatchEndsQuietlyWhenInterrupted(), TestWatchIntervalIsValidated(), TestWatchKeepsGoingAfterAFailedRefresh() (+3 more)

### Community 81 - "replicationops_test.go"
Cohesion: 0.10
Nodes (49): replicationops_test.go, hasCall(), ReplicationConfigResult, isRunCLIRejection(), mateConvergenceShutdowns(), missingListedVPNs(), newlineIf(), PlannedRoles() (+41 more)

### Community 82 - "replCfg"
Cohesion: 0.06
Nodes (50): keyValueEntries, ReplVia, ReplViaKube, ReplViaSEMP, matechannel_test.go, CLIArg(), CLIScriptPath(), ReplPassSecret (+42 more)

### Community 83 - "limits_test.go"
Cohesion: 0.06
Nodes (60): limits_test.go, main_test.go, delegateSetting(), limitText(), missingControllers(), pamCoreValue(), parseLimit(), parseUnitProps() (+52 more)

### Community 84 - "Developer guide"
Cohesion: 0.12
Nodes (17): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+9 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.09
Nodes (39): inspect_test.go, age(), anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), podHealth(), pvcLevel() (+31 more)

### Community 87 - ".ConfigureReplication"
Cohesion: 0.06
Nodes (69): colSpan, replication_test.go, ValidVPNName(), validVPNName(), cliTransport(), containsEndpoint(), dashSpans(), flagByte() (+61 more)

### Community 88 - "AdminSecret"
Cohesion: 0.12
Nodes (33): secrets_test.go, GenSecrets(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), readTLSMaterial(), checkGolden() (+25 more)

### Community 89 - "render_test.go"
Cohesion: 0.15
Nodes (41): render_test.go, boolStr(), BrokerCR(), ComposeProject(), containerArtifacts(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets() (+33 more)

### Community 91 - "RuleFor"
Cohesion: 0.13
Nodes (36): Disposition, SectionRule, appliance_test.go, importdoc_test.go, sections_test.go, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent() (+28 more)

### Community 93 - "Container"
Cohesion: 0.20
Nodes (9): guardCommandOf(), setGuardCommand(), TestValidatorAndExecutorAgree(), applyContainerBlockDefaults(), podmanRunUser(), setDefault(), setDefaultCmd(), setDefaultInt() (+1 more)

### Community 97 - "replConfig"
Cohesion: 0.19
Nodes (15): Replication, secretRef, replication_test.go, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded() (+7 more)

### Community 102 - "DetectPlatforms"
Cohesion: 0.23
Nodes (18): examples_test.go, internal/examples, Example, parseError(), DetectPlatforms(), All(), Get(), List() (+10 more)

### Community 107 - "operatorversion_test.go"
Cohesion: 0.13
Nodes (18): operatorversion_test.go, TestOperatorApplyPlansItsNamespace(), operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsNotSkippedByAFailedRead(), TestConfirmNoDowngradeIsSilentInAPreview(), TestConfirmNoDowngradeProceedsWhenConfirmAccepts() (+10 more)

### Community 109 - "Config"
Cohesion: 0.08
Nodes (35): AdditionalUser, Container, DockerConfig, Network, Node, PodmanConfig, Redundancy, SEMP (+27 more)

### Community 110 - "config_ops_test.go"
Cohesion: 0.16
Nodes (15): config_ops_test.go, internal/broker, TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(), TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator() (+7 more)

### Community 112 - "capRunner"
Cohesion: 0.11
Nodes (27): capCall, capRunner, runtime_test.go, New(), NewManager(), callIndex(), TestManagerLogsCLIShell(), mgrOver() (+19 more)

### Community 116 - "certFixture"
Cohesion: 0.17
Nodes (28): secrets_test.go, ResolveSecretValues(), bundleHash(), certCreate(), certFixture(), TestCertSecretLabelDrivesTheRestart(), TestContainerBundleCarriesTheChain(), TestDeletePodmanToleratesAMissingBundle() (+20 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.10
Nodes (25): containerState, healthState, inspect_test.go, decodeInspect(), orUnknown(), printable(), inspectReply(), kvValue() (+17 more)

### Community 121 - "Role"
Cohesion: 0.06
Nodes (22): covUploadFailsTo, downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile (+14 more)

### Community 126 - "bytes.Buffer"
Cohesion: 0.13
Nodes (35): watch_test.go, desiredWatch(), splitWatch(), subtractWatch(), deniedCluster(), deniedRunner(), deployJSON(), namedOperatorJSON() (+27 more)

### Community 128 - "hostpath_test.go"
Cohesion: 0.08
Nodes (33): hostpath_test.go, guardConfig(), BaseName(), CheckHostPath(), expandTilde(), expandTildeToken(), HasPathSeparator(), IsAbsHostPath() (+25 more)

### Community 139 - "blocks.go"
Cohesion: 0.09
Nodes (42): Capture, Omission, token, blocks_test.go, omit_test.go, Annotate(), TestAnnotateIsDeterministic(), TestAnnotateRegionAndSectionFencesAreBalanced() (+34 more)

### Community 155 - "strings.Builder"
Cohesion: 0.07
Nodes (43): shorthand, abbrevdoc_test.go, mdRow(), renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeAbbrevTable() (+35 more)

## Knowledge Gaps
- **13 isolated node(s):** `solace`, `NO_COLOR`, `bash/000-env.sh`, `bash/010-deploy-operator.sh`, `bash/020-deploy-broker.sh` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 188 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **6 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `App` to `opPerformReplication`, `watchLoop`, `cli_test.go`, `github.com/spf13/cobra.ShellCompDirective`, `EnvPairs`, `newRootCmd`, `exportconfig_test.go`, `Config`, `replication_test.go`, `capRunner`, `commands.go`, `config_test.go`, `cli/platform_test.go`, `.dump`, `context.Context`, `Role`, `confirm.go`, `Platform`?**
  _High betweenness centrality (0.038) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `newTestOps()` (e.g. with `importops_test.go` and `Injectable seams`) actually correct?**
  _`newTestOps()` has 117 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `NO_COLOR`, `bash/000-env.sh` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.0669652855543113 - nodes in this community are weakly interconnected._
- **Why does `Per-package doubles` connect `.dump` to `App`, `newTestOps`, `cli_test.go`, `config_test.go`, `Manager`, `runLeader`, `NewCluster`, `convert_test.go`, `transform_test.go`, `manager_test.go`, `Load`, `Set`, `switchplan_test.go`, `newLocalOps`, `runner_test.go`, `strings.Builder`, `importops_test.go`, `statusreport_test.go`, `deploymentItem`, `NewCluster`, `exportconfig_test.go`, `VPNRepl`, `localCfg`, `eqArgs`, `Ops`, `prep_test.go`, `context.Context`, `replication_test.go`, `.OperatorApply`, `exit_test.go`, `opPerformReplication`, `validContainerConfig`, `replCfg`, `k8s/inspect.go`, `.ConfigureReplication`, `AdminSecret`, `replConfig`, `operatorversion_test.go`, `capRunner`, `certFixture`, `Role`, `bytes.Buffer`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.0715950715950716 - nodes in this community are weakly interconnected._
- **Why does `Commands` connect `Command Reference Docs` to `Command reference`, `.configRows`, `replication_test.go`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._
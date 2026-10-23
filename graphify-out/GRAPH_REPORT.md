# Graph Report - solace-cnt-scripts  (2026-10-24)

## Corpus Check
- 204 files · ~705,502 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .golden 17, (none) 3, .cli 2)

## Summary
- 4067 nodes · 20760 edges · 110 communities (101 shown, 9 thin omitted)
- Extraction: 72% EXTRACTED · 28% INFERRED · 0% AMBIGUOUS · INFERRED: 5785 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `f239d143`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- App
- newTestOps
- output_test.go
- .ns
- cli_test.go
- config_test.go
- Command Reference Docs
- Manager
- K8sConfig
- completion_test.go
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
- testing.T
- importops_test.go
- deploymentItem
- Solace Module Root
- .validateContainer
- CLAUDE.md
- internal/config
- ResolveDomainCerts
- watchLoop
- newRootCmd
- NewCluster
- annotate_test.go
- diff_test.go
- Compose
- inject_test.go
- Test catalogue
- exportconfig_test.go
- VPNRepl
- localCfg
- config_test.go
- eqArgs
- allowcommand_test.go
- scripts_test.go
- prep_test.go
- .validate
- .dump
- .kubectl
- replication_test.go
- Exec
- .rpc
- confirm.go
- .configRows
- rootless_test.go
- exit_test.go
- .releaseToBackup
- Coverage
- opPerformReplication
- Capture
- Echo
- EnvPairs
- writeNode
- .checkUserManagerLimits
- Configuration
- namespace_test.go
- redundancy_test.go
- usagef
- podName
- command_test.go
- cli/watch_test.go
- Container
- omit_test.go
- replicationops_test.go
- replCfg
- limits_test.go
- Developer guide
- euid_test.go
- k8s/inspect.go
- .ConfigureReplication
- AdminSecret
- render_test.go
- Image
- RuleFor
- capture
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
- fakeTransport
- containerTransport
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
- `The preflight` --references--> `info()`  [INFERRED]
  docs/operations.md → internal/k8s/checkreport.go
- `Per-package doubles` --references--> `sample()`  [INFERRED]
  docs/test.md → internal/abbrev/abbrev_test.go
- `annotate_test.go` --references--> `Annotate()`  [INFERRED]
  docs/test.md → internal/broker/annotate.go
- `omit_test.go` --references--> `Annotate()`  [INFERRED]
  docs/test.md → internal/broker/annotate.go
- `annotate_test.go` --references--> `TestAnnotateRoundTripPreservesBlocks()`  [INFERRED]
  docs/test.md → internal/broker/annotate_test.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (110 total, 9 thin omitted)

### Community 0 - "App"
Cohesion: 0.06
Nodes (101): App, logArgs(), confirmAction(), confirmDelete(), warn(), confirmImport(), runExport(), runImport() (+93 more)

### Community 1 - "newTestOps"
Cohesion: 0.08
Nodes (68): recOutput, broker_test.go, countContains(), leaderCalls(), newTestOps(), TestCountContains(), TestDiagnostics(), TestDisableDefaultUsers() (+60 more)

### Community 2 - "output_test.go"
Cohesion: 0.15
Nodes (23): internal/output, output_test.go, columnWidths(), Sink, NewFunc(), pad(), sink(), TestKVBlockAlignsOnTheLongestKey() (+15 more)

### Community 3 - ".ns"
Cohesion: 0.10
Nodes (10): TestOperatorAdminSecretNameMatchesTheBrokerNames(), age(), discoveryServiceName(), HARoles(), lbServiceName(), podName(), pvcName(), stsName() (+2 more)

### Community 4 - "cli_test.go"
Cohesion: 0.07
Nodes (108): cli_test.go, Shared env fixtures, allowRuntime(), captureStderr(), collectPaths(), echoRunner(), fakeBinaryOnPath(), firstLine() (+100 more)

### Community 5 - "config_test.go"
Cohesion: 0.06
Nodes (74): config_test.go, validPlatformConfig(), assertContainerBlockDefaults(), assertContainerScaling(), envTree(), haNodesConfig(), TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes() (+66 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (50): Commands, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker cli, solace-util broker configure (+42 more)

### Community 7 - "Manager"
Cohesion: 0.06
Nodes (18): preflight_test.go, certDigest(), composeNeedsSecretValues(), exactName(), Manager, platformTitle(), TestManagerLifecycleDockerComposeFile(), TestManagerLifecycleDockerNoComposeFile() (+10 more)

### Community 8 - "K8sConfig"
Cohesion: 0.12
Nodes (14): ContainerSecurity, Operator, PodSecurity, Resources, K8sConfig, LoadBalancer, cut(), parsePort() (+6 more)

### Community 9 - "completion_test.go"
Cohesion: 0.22
Nodes (24): completion_test.go, driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion() (+16 more)

### Community 10 - "CheckCommand"
Cohesion: 0.11
Nodes (33): cmdField, commandRules, guardedCmd, execguard_test.go, decodeCommand(), Command, checkBinary(), CheckCommand() (+25 more)

### Community 11 - "runLeader"
Cohesion: 0.60
Nodes (4): leaderRun, runLeader(), TestAssertLeaderAsksBeforeActingFromABackup(), TestAssertLeaderPrimaryAndMonitorNeverAsk()

### Community 12 - "NewCluster"
Cohesion: 0.08
Nodes (60): operator_test.go, GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), RenderOperator(), splitAfterNamespace(), splitOperatorBundle() (+52 more)

### Community 13 - "BrokerType"
Cohesion: 0.15
Nodes (12): BrokerType, cliMate, transcript_test.go, ValidVPNName(), validVPNName(), bannerType(), stripTranscript(), showReplicationScript() (+4 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (69): doc, Result, segment, vars, The keys, convert_test.go, DecodeStrict(), boolOf() (+61 more)

### Community 15 - "transform_test.go"
Cohesion: 0.10
Nodes (62): Block, TargetState, transform_test.go, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects() (+54 more)

### Community 16 - "manager_test.go"
Cohesion: 0.09
Nodes (96): manager_test.go, GuardPodmanEUID(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountUnknownWhenSystemctlFails(), fileExists(), assertMode(), containsStr(), ctrCfg() (+88 more)

### Community 17 - "commands.go"
Cohesion: 0.13
Nodes (72): opFunc, roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd() (+64 more)

### Community 18 - "coverage_test.go"
Cohesion: 0.06
Nodes (69): runErrMatchTransport, runErrTransport, coverage_test.go, verify_ops_test.go, matchCLI(), ranContains(), writeFile(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList() (+61 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "Load"
Cohesion: 0.07
Nodes (47): scalingTier, scaling_test.go, minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs() (+39 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Set"
Cohesion: 0.16
Nodes (16): Entry, abbrev_test.go, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed() (+8 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.07
Nodes (75): CLIRunner, fakeRun, PhaseKind, SiteState, SwitchAction, SwitchPhase, matechannel_test.go, switchplan_test.go (+67 more)

### Community 24 - "newLocalOps"
Cohesion: 0.16
Nodes (40): verify_local_test.go, cliScriptPath(), isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn(), TestDefaultLocalAddrs() (+32 more)

### Community 25 - "runner_test.go"
Cohesion: 0.19
Nodes (24): runner_test.go, captureStdout(), helperCommand(), TestEchoDefaultWriter(), TestEchoOutput(), TestEchoOutputInput(), TestEchoRun(), TestEchoRunEnv() (+16 more)

### Community 26 - "renderDriver"
Cohesion: 0.17
Nodes (22): chunk, chunkResult, driver_test.go, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce() (+14 more)

### Community 27 - "testing.T"
Cohesion: 0.07
Nodes (71): platform_test.go, platform_test.go, statusreport_test.go, storage_test.go, findCmd(), TestEveryRunnableCommandIsWired(), supported(), runPlatform() (+63 more)

### Community 28 - "importops_test.go"
Cohesion: 0.06
Nodes (74): ImportResult, PlannedSection, importops_test.go, defaultVPNFirst(), describeScope(), ImportPlan, importIgnore(), orNone() (+66 more)

### Community 29 - "deploymentItem"
Cohesion: 0.13
Nodes (15): compareVersions(), imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), operatorRunningImage(), TestOperatorRunningImageWithNoContainers(), splitWatch() (+7 more)

### Community 31 - ".validateContainer"
Cohesion: 0.08
Nodes (17): artifactvalues_test.go, portOrRange(), TestValidateContainerArtifactValues(), TestValidateImageReference(), TestValidateK8sArtifactValues(), TestValidContainerPort(), validContainerPort(), validCoreLimit() (+9 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.05
Nodes (37): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, Abbreviations, Flag shorthands, Node roles (+29 more)

### Community 33 - "internal/config"
Cohesion: 0.14
Nodes (20): adminsecret_test.go, duration_test.go, internal/config, tlssecret_test.go, adminCfg(), TestAdminSecretRefusesTheOperatorsOwnNames(), TestContainersStillRequireTheAdminPassword(), TestKubernetesValidatesWithoutAnAdminPassword() (+12 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (53): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domaincerts_resolve_test.go, domaincerts_test.go, caNameSafe() (+45 more)

### Community 35 - "watchLoop"
Cohesion: 0.32
Nodes (5): watchInterval(), stdoutIsTerminal(), watchLabel(), watchLoop(), withWatch()

### Community 36 - "newRootCmd"
Cohesion: 0.10
Nodes (27): How a short form is resolved, Goldens, abbrevdoc_test.go, aliases_test.go, commanddoc_test.go, TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), applyAliases() (+19 more)

### Community 37 - "NewCluster"
Cohesion: 0.08
Nodes (70): checkreport_test.go, deploy_test.go, preflight_test.go, TestCheckStopsProbingWhenUnreachable(), TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestOperatorInstallVerdict(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestOperatorValidateAsksNothingAboutTheBroker() (+62 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.10
Nodes (48): Region, annotate_test.go, applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV() (+40 more)

### Community 40 - "diff_test.go"
Cohesion: 0.11
Nodes (37): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, diff_test.go, ancestorReported(), blockLabel() (+29 more)

### Community 41 - "Compose"
Cohesion: 0.16
Nodes (19): internal/render, servercert_test.go, Compose(), composeEscape(), ComposeProject(), ContainerSecrets(), ContainerSecret, SecretPreflight() (+11 more)

### Community 42 - "inject_test.go"
Cohesion: 0.17
Nodes (28): lineRole, serviceLine, shutdownStyle, svcKey, inject_test.go, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "Test catalogue"
Cohesion: 0.29
Nodes (7): Fixtures and doubles, internal/abbrev, internal/convert, internal/engine, Running the tests, Summary, Test catalogue

### Community 44 - "exportconfig_test.go"
Cohesion: 0.13
Nodes (27): exportconfigReadCounter, opRunner, exportconfig_test.go, internal/cli, pluralVPN(), exportconfigDriverOutput(), exportconfigTransportOutput(), exportconfigVPNListOutput() (+19 more)

### Community 47 - "VPNRepl"
Cohesion: 0.08
Nodes (19): AdminState, QueueState, scriptedMate, VPNRepl, sortedKeys(), sortedSet(), isRunCLIRejection(), mateConvergenceShutdowns() (+11 more)

### Community 48 - "localCfg"
Cohesion: 0.10
Nodes (47): semp_test.go, sempmate_test.go, bridgeHostPort(), sempPort(), sempV1OK(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA() (+39 more)

### Community 49 - "config_test.go"
Cohesion: 0.16
Nodes (34): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, span, AdditionalUser, Node, Redundancy (+26 more)

### Community 50 - "eqArgs"
Cohesion: 0.11
Nodes (39): check_test.go, cluster_test.go, internal/k8s, ops_test.go, transport_test.go, TestReachable(), TestResolveStorageClass(), newCluster() (+31 more)

### Community 51 - "allowcommand_test.go"
Cohesion: 0.35
Nodes (12): allowcommand_test.go, TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRegisteredWhereItExecutes(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes() (+4 more)

### Community 52 - "scripts_test.go"
Cohesion: 0.10
Nodes (40): showCmd, scripts_test.go, assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript(), enableDefaultUsersScript() (+32 more)

### Community 53 - "prep_test.go"
Cohesion: 0.10
Nodes (49): names_test.go, prep_test.go, BrokerPodSuffixShape(), ProductKeyRoles(), RestartOrder(), haCfg(), TestHARoles(), TestPodNameSuffixMatchesTheConfigBound() (+41 more)

### Community 54 - ".validate"
Cohesion: 0.06
Nodes (37): solace-util, solace-util operator, A replication switch refuses before changing anything, A wrapper runtime is refused, Bringing up a fresh cluster, Broker scope is a fixed classification, Docker and Podman mechanics, Docker compose secrets need compose 2.23.1+ (+29 more)

### Community 55 - ".dump"
Cohesion: 0.10
Nodes (64): opCall, ops_k8s_test.go, captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook(), loadDirect() (+56 more)

### Community 56 - ".kubectl"
Cohesion: 0.07
Nodes (17): runtime_test.go, Cluster, operatorEscalationProbes(), operatorProbes(), canIAnswer(), canIRefused(), missingNamespace(), namespaceManifest() (+9 more)

### Community 57 - "replication_test.go"
Cohesion: 0.07
Nodes (60): replMate, replRig, Commands, solace-util broker, Configuring a site, Data replication, Permissions, Reaching the mate (+52 more)

### Community 58 - "Exec"
Cohesion: 0.20
Nodes (12): Keys that were renamed, resolve_test.go, Exec, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveMissingBinaryIsActionable() (+4 more)

### Community 59 - ".rpc"
Cohesion: 0.11
Nodes (16): backupTarget, Credential, sempMate, TestHTTPStatusHelpers(), schemaTransport(), anyHTTP2xx(), curlConfigLine(), httpBody() (+8 more)

### Community 60 - "confirm.go"
Cohesion: 0.25
Nodes (14): layer, Injectable seams, confirmActionStrict(), confirmDowngrade(), confirmGate(), confirmLayer(), confirmRestart(), interactive() (+6 more)

### Community 61 - ".configRows"
Cohesion: 0.12
Nodes (28): orNone(), orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminPassState(), adminSecretRow() (+20 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.19
Nodes (40): rootless_test.go, failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+32 more)

### Community 63 - "exit_test.go"
Cohesion: 0.15
Nodes (23): runFailRunner, Repository layout, exit_test.go, TestExecute(), childExit(), ExitCode(), isUsage(), childStatusError() (+15 more)

### Community 64 - ".releaseToBackup"
Cohesion: 0.18
Nodes (7): field(), revertActivityConfigureScript(), showRedundancyLocalScript(), defaultLocalAddrs(), activity(), primaryRedundancyUp(), rdEnabledUp()

### Community 65 - "Coverage"
Cohesion: 0.19
Nodes (15): Coverage, ContainerNoFile(), HealthCheck, TestLimitsCheckAssertsWhatTheArtifactAsks(), escapePercent(), healthCmd(), Quadlet(), quadletEscape() (+7 more)

### Community 66 - "opPerformReplication"
Cohesion: 0.18
Nodes (11): mateChannelFunc, step(), confirmReplicationConfig(), confirmReplicationSwitch(), opConfigureReplication(), opCtrPerformReplication(), opPerformReplication(), replVPNNames() (+3 more)

### Community 67 - "Capture"
Cohesion: 0.22
Nodes (11): Capture, appliance_test.go, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified(), parseHeader() (+3 more)

### Community 68 - "Echo"
Cohesion: 0.26
Nodes (4): interactiveFailRunner, Echo, Quote(), quoteTok()

### Community 69 - "EnvPairs"
Cohesion: 0.16
Nodes (12): NodeIdentity, containerSecretSpecs(), EnvPairs(), groupKey(), itoa(), secretFilePath(), assertNoCheckoutPath(), envLines() (+4 more)

### Community 70 - "writeNode"
Cohesion: 0.21
Nodes (12): WeightedNodeTerm, NodeAffinity, NodeMatchExpr, Placement, validateMatchExprs(), validatePlacementAffinity(), parseToleration(), TestParseToleration() (+4 more)

### Community 71 - ".checkUserManagerLimits"
Cohesion: 0.26
Nodes (4): delegateSetting(), limitText(), parseLimit(), parseUnitProps()

### Community 72 - "Configuration"
Cohesion: 0.17
Nodes (12): Bring your own TLS Secret, Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling, Scaling tiers (+4 more)

### Community 73 - "namespace_test.go"
Cohesion: 0.27
Nodes (11): namespace_test.go, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 74 - "redundancy_test.go"
Cohesion: 0.21
Nodes (12): redundancy_test.go, pskCfg(), TestFillStandaloneNodeNameIgnoresAnEmptyHostname(), TestFillStandaloneNodeNameKeepsAConfiguredName(), TestFillStandaloneNodeNameNeverFillsHA(), TestFillStandaloneNodeNameUsesTheHost(), TestNothingGeneratesThePSK(), TestPSKEnvSatisfiesTheContainerRequirement() (+4 more)

### Community 75 - "usagef"
Cohesion: 0.09
Nodes (30): childExitError, usageError, runConvert(), newExamplesCmd(), runExample(), asUsage(), childStatus(), markUsageArgs() (+22 more)

### Community 77 - "command_test.go"
Cohesion: 0.15
Nodes (14): command_test.go, TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRenamedRuntimeKeysFailLoud() (+6 more)

### Community 78 - "cli/watch_test.go"
Cohesion: 0.35
Nodes (11): watch_test.go, stopAfter(), TestOperatorStatusWatches(), TestStatusWithoutWatchRunsOnce(), TestWatchAppendsWhenNotATerminal(), TestWatchEndsQuietlyWhenInterrupted(), TestWatchIntervalIsValidated(), TestWatchKeepsGoingAfterAFailedRefresh() (+3 more)

### Community 79 - "Container"
Cohesion: 0.27
Nodes (6): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestContainerCoreLimitSpellings()

### Community 80 - "omit_test.go"
Cohesion: 0.24
Nodes (10): Omission, omit_test.go, omitAtExport(), omittedList(), TestImportPlanAcceptsWhatExportProduced(), TestImportPlanRefusesAFileItDidNotExport(), TestOmitAtExportDropsFilteredLinesAndEmptyBlocks(), TestOmitAtExportLeavesTargetDecisionsToImport() (+2 more)

### Community 81 - "replicationops_test.go"
Cohesion: 0.14
Nodes (40): replicationops_test.go, hasCall(), PlannedRoles(), replicationVPNLines(), RoleAtSite(), replListed(), replReadOnlyResponder(), replTestMate() (+32 more)

### Community 82 - "replCfg"
Cohesion: 0.06
Nodes (49): keyValueEntries, ReplVia, ReplViaKube, ReplViaSEMP, matechannel_test.go, CLIArg(), CLIScriptPath(), ReplPassSecret (+41 more)

### Community 83 - "limits_test.go"
Cohesion: 0.09
Nodes (49): internal/tools/vulnjudge, limits_test.go, main_test.go, missingControllers(), limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager() (+41 more)

### Community 84 - "Developer guide"
Cohesion: 0.25
Nodes (8): Build, Dev script tasks, Developer guide, Gates, Releases, Tests, The env-file generator page, Toolchain pin

### Community 85 - "euid_test.go"
Cohesion: 0.67
Nodes (6): euid_test.go, runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand(), TestPodmanEUIDGuardSkips(), writeRootlessPodmanEnv()

### Community 86 - "k8s/inspect.go"
Cohesion: 0.13
Nodes (36): inspect_test.go, anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), normalizeToList(), podHealth(), pvcLevel() (+28 more)

### Community 87 - ".ConfigureReplication"
Cohesion: 0.09
Nodes (64): colSpan, replication_test.go, cliTransport(), containsEndpoint(), dashSpans(), flagByte(), MateConfig, gutterClear() (+56 more)

### Community 88 - "AdminSecret"
Cohesion: 0.13
Nodes (36): secrets_test.go, GenSecrets(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), operatorRegcred(), readTLSMaterial() (+28 more)

### Community 89 - "render_test.go"
Cohesion: 0.16
Nodes (38): render_test.go, boolStr(), BrokerCR(), containerArtifacts(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets(), TestArtifactsCarryNoWideningTokens() (+30 more)

### Community 91 - "RuleFor"
Cohesion: 0.16
Nodes (30): Disposition, SectionRule, importdoc_test.go, sections_test.go, renderImportDocs(), TestEverySectionRuleHasAReason(), TestImportDocs(), TestOnlyOneSectionIsAppliedFirst() (+22 more)

### Community 93 - "Container"
Cohesion: 0.19
Nodes (9): guardCommandOf(), applyContainerBlockDefaults(), podmanRunUser(), setDefault(), setDefaultCmd(), setDefaultInt(), xdgConfigHome(), containerMem() (+1 more)

### Community 97 - "replConfig"
Cohesion: 0.19
Nodes (15): Replication, secretRef, replication_test.go, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded() (+7 more)

### Community 102 - "DetectPlatforms"
Cohesion: 0.13
Nodes (27): examples_test.go, examples_test.go, internal/examples, Example, TestExamplesAcceptsAPlatformAbbreviation(), TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), TestExamplesNeedsNoEnvFile() (+19 more)

### Community 107 - "operatorversion_test.go"
Cohesion: 0.13
Nodes (19): operatorversion_test.go, TestOperatorApplyPlansItsNamespace(), operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsNotSkippedByAFailedRead(), TestConfirmNoDowngradeIsSilentInAPreview(), TestConfirmNoDowngradeProceedsWhenConfirmAccepts() (+11 more)

### Community 109 - "Config"
Cohesion: 0.12
Nodes (23): TLS, adminsecret_test.go, imagepullsecret_test.go, TestAdminSecretNameFollowsTheStates(), TestPSKConfiguredCountsBothSpellings(), Config, pullSecretCfg(), TestImagePullSecretDerivedSuffixFitsWhereAdditionalUsersAlreadyFits() (+15 more)

### Community 110 - "config_ops_test.go"
Cohesion: 0.21
Nodes (12): config_ops_test.go, TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(), TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator(), TestExecShellScriptRefusesAnUnsafeNameBeforeUploading() (+4 more)

### Community 112 - "capRunner"
Cohesion: 0.08
Nodes (30): capCall, capRunner, internal/container, runtime_test.go, transport_test.go, New(), NewManager(), callIndex() (+22 more)

### Community 116 - "certFixture"
Cohesion: 0.13
Nodes (29): covUploadFailsTo, uploadErrTransport, Per-package doubles, secrets_test.go, maskedKeys(), ResolveSecretValues(), bundleHash(), certCreate() (+21 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.09
Nodes (26): containerState, healthState, inspect_test.go, decodeInspect(), orUnknown(), printable(), inspectReply(), kvValue() (+18 more)

### Community 121 - "Role"
Cohesion: 0.09
Nodes (14): Storage, cliRunNames(), Ops, runCLISkeleton(), validCLILine(), validName(), ValidScriptName(), rejectionIn() (+6 more)

### Community 126 - "bytes.Buffer"
Cohesion: 0.13
Nodes (27): watch_test.go, subtractWatch(), deniedCluster(), deployJSON(), namedOperatorJSON(), operatorItem(), TestDesiredWatchAppendsTheBrokerNamespace(), TestFindOperatorDeploymentFallsBackWhenTheClusterWideListFails() (+19 more)

### Community 128 - "hostpath_test.go"
Cohesion: 0.08
Nodes (34): hostpath_test.go, guardConfig(), BaseName(), CheckHostPath(), expandTilde(), expandTildeToken(), HasPathSeparator(), IsAbsHostPath() (+26 more)

### Community 130 - "fakeTransport"
Cohesion: 0.12
Nodes (10): downloadErrTransport, fakeTransport, recDownload, recRun, recUpload, recUploadFile, cliScriptNameFromDest(), outputForRole() (+2 more)

### Community 139 - "blocks.go"
Cohesion: 0.14
Nodes (30): token, blocks_test.go, internal/broker, Annotate(), TestAnnotateIsDeterministic(), TestAnnotateRegionAndSectionFencesAreBalanced(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks() (+22 more)

### Community 155 - "strings.Builder"
Cohesion: 0.12
Nodes (24): shorthand, mdRow(), renderAbbrevDocs(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable(), writeUnabbreviated() (+16 more)

## Knowledge Gaps
- **13 isolated node(s):** `solace`, `NO_COLOR`, `bash/000-env.sh`, `bash/010-deploy-operator.sh`, `bash/020-deploy-broker.sh` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 188 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **9 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `App` to `opPerformReplication`, `watchLoop`, `cli_test.go`, `newRootCmd`, `Manager`, `usagef`, `exportconfig_test.go`, `Config`, `capRunner`, `commands.go`, `config_test.go`, `certFixture`, `.dump`, `replication_test.go`, `testing.T`, `confirm.go`?**
  _High betweenness centrality (0.040) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `newTestOps()` (e.g. with `importops_test.go` and `Injectable seams`) actually correct?**
  _`newTestOps()` has 117 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `NO_COLOR`, `bash/000-env.sh` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.06134731120170841 - nodes in this community are weakly interconnected._
- **Why does `Per-package doubles` connect `certFixture` to `App`, `newTestOps`, `fakeTransport`, `cli_test.go`, `config_test.go`, `Manager`, `runLeader`, `NewCluster`, `convert_test.go`, `transform_test.go`, `manager_test.go`, `coverage_test.go`, `Load`, `Set`, `switchplan_test.go`, `newLocalOps`, `runner_test.go`, `strings.Builder`, `importops_test.go`, `testing.T`, `deploymentItem`, `NewCluster`, `Test catalogue`, `exportconfig_test.go`, `VPNRepl`, `localCfg`, `eqArgs`, `prep_test.go`, `.dump`, `.kubectl`, `replication_test.go`, `Exec`, `.rpc`, `exit_test.go`, `.releaseToBackup`, `opPerformReplication`, `replCfg`, `k8s/inspect.go`, `.ConfigureReplication`, `AdminSecret`, `capture`, `replConfig`, `operatorversion_test.go`, `capRunner`, `Role`, `bytes.Buffer`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.08157349896480331 - nodes in this community are weakly interconnected._
- **Why does `Commands` connect `Command Reference Docs` to `CLAUDE.md`, `replication_test.go`, `.validate`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
# Graph Report - solace-k8-scripts  (2026-09-14)

## Corpus Check
- 180 files · ~509,336 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 3331 nodes · 10522 edges · 134 communities (121 shown, 9 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 1939 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `30176770`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- cli_test.go
- Sink
- context.Context
- completion_test.go
- config_test.go
- Commands
- Manager
- Config
- operator.go
- Command
- capRunner
- NewCluster
- eqArgs
- convert_test.go
- Block
- ctrCfg
- importops.go
- k8s/secrets_test.go
- dev.sh
- prep_test.go
- dev.ps1
- newTestOps
- .Run
- verify_local_test.go
- .releaseToBackup
- renderDriver
- Platform
- commanddoc_test.go
- solace
- validate.go
- CLAUDE.md
- importops_test.go
- kubectlTransport
- Role
- .resolveSecretRefs
- haCfg
- MateChannel
- helpers.go
- statusreport_test.go
- cli/platform_test.go
- InjectShutdown
- runRootWith
- NewMateChannel
- opRunner
- broker/replication.go
- .ConfigureReplication
- testing.T
- Ops
- runner_test.go
- newRootCmd
- cli/replication_test.go
- container/inspect_test.go
- Operations
- Test catalogue
- Exporting and importing configuration
- DiffBlocks
- .rpc
- Cluster
- containerTransport
- checkreport.go
- render.go
- replicationops_test.go
- Configuration
- newCapMgr
- broker/replication_test.go
- Command reference
- Developer guide
- ParseBlocks
- localCfg
- output_test.go
- internal/k8s
- annotate_test.go
- Set
- namespace_test.go
- io.Writer
- validContainerConfig
- switchplan_test.go
- scaling_test.go
- blocks.go
- age
- k8s/inspect_test.go
- execguard_test.go
- broker/matechannel_test.go
- k8s/inspect.go
- .preflightOne
- github.com/spf13/cobra.Command
- load
- judge
- RuleFor
- storage_test.go
- applyAliases
- containerState
- strings.Builder
- Abbreviations
- platformOps
- tlsCfg
- scripts.go
- opPerformReplication
- container/secrets_test.go
- watch_test.go
- internal/cli
- Load
- operatorversion.go
- What `import-config` applies
- ReplSite
- internal/broker
- exit_test.go
- Troubleshooting
- internal/config
- newSEMPTestMate
- Fixtures and doubles
- operatorversion_test.go
- internal/container
- Cluster
- load.go
- blocks_test.go
- NewManager
- .Run
- replicationVPNLines
- resolve_test.go
- ResolveEnvPath
- rootlessNoFileMgr
- .hostPath
- runConvert
- cliMate
- ReplRole
- runExample
- internal/engine
- ParseRole

## God Nodes (most connected - your core abstractions)
1. `newTestOps()` - 127 edges
2. `Role` - 116 edges
3. `ctrCfg()` - 115 edges
4. `newCapMgr()` - 100 edges
5. `Config` - 94 edges
6. `NewCluster()` - 89 edges
7. `matchCLI()` - 80 edges
8. `loadK8s()` - 77 edges
9. `Platform` - 72 edges
10. `bg()` - 68 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Execute()`  [EXTRACTED]
  main.go → internal/cli/root.go
- `parseBody()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `parseHeader()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `parseBody()` --calls--> `readMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `parseHeader()` --calls--> `readMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (134 total, 9 thin omitted)

### Community 0 - "bg"
Cohesion: 0.07
Nodes (103): TestConfirmDeleteShortcut(), TestCtrManagerConfirmWiring(), lineSink(), progress(), warn(), confirmImport(), App, pluralVPN() (+95 more)

### Community 1 - "cli_test.go"
Cohesion: 0.07
Nodes (63): os.File, allowRuntime(), capture(), captureStderr(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone() (+55 more)

### Community 2 - "Sink"
Cohesion: 0.14
Nodes (8): TestCheckReportSkipsEmptySections(), KV, Sink, New(), NewFunc(), pad(), TestNewFuncRoutesThroughTheGivenLineSink(), TestNilSinksDiscard()

### Community 3 - "context.Context"
Cohesion: 0.07
Nodes (10): scriptedMate, context.Context, Manager, Cluster, Cluster, pvcName(), Cluster, Cluster (+2 more)

### Community 4 - "completion_test.go"
Cohesion: 0.09
Nodes (41): Example, github.com/spf13/cobra.ShellCompDirective, completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), completionShell(), App (+33 more)

### Community 5 - "config_test.go"
Cohesion: 0.07
Nodes (40): assertContainerBlockDefaults(), assertContainerScaling(), Config, haNodesConfig(), TestAdditionalUserPasswordCharsAreFreeOnKubernetes(), TestApplyDefaultsDocker(), TestApplyDefaultsK8s(), TestApplyDefaultsPodmanRootful() (+32 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.09
Nodes (12): os.FileMode, composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing() (+4 more)

### Community 8 - "Config"
Cohesion: 0.09
Nodes (27): AdditionalUser, Broker, Container, ContainerSecurity, DockerConfig, DomainCerts, Image, Network (+19 more)

### Community 9 - "operator.go"
Cohesion: 0.24
Nodes (14): GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator(), renderOperatorNS(), splitAfterNamespace() (+6 more)

### Community 10 - "Command"
Cohesion: 0.13
Nodes (21): commandRules, Command, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules() (+13 more)

### Community 11 - "capRunner"
Cohesion: 0.18
Nodes (15): capCall, capRunner, New(), Transport, TestManagerLogsCLIShell(), NewTransport(), dockerCfg(), eqArgs() (+7 more)

### Community 12 - "NewCluster"
Cohesion: 0.07
Nodes (67): TestCheckStopsProbingWhenUnreachable(), TestResolveStorageClass(), TestStorageRows(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts() (+59 more)

### Community 13 - "eqArgs"
Cohesion: 0.17
Nodes (23): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), TestCLIAndShellAreInteractive() (+15 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (62): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+54 more)

### Community 15 - "Block"
Cohesion: 0.09
Nodes (54): Block, TargetState, regexp.Regexp, importIgnore(), injectedBlock(), BridgeEnablementInverted(), ClearExistingClientCAs(), ClearExistingNested() (+46 more)

### Community 16 - "ctrCfg"
Cohesion: 0.08
Nodes (68): fileExists(), assertMode(), containsStr(), ctrCfg(), hasCall(), maskedKeys(), newEchoMgr(), setNoFile() (+60 more)

### Community 17 - "importops.go"
Cohesion: 0.22
Nodes (16): ImportResult, PlannedSection, defaultVPNFirst(), ImportPlan, omittedList(), orNone(), preambleForApply(), quoteList() (+8 more)

### Community 18 - "k8s/secrets_test.go"
Cohesion: 0.11
Nodes (31): GenBroker(), GenSecrets(), joinManifests(), namespaceManifest(), TestUpdateServerCertSecret(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret() (+23 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "prep_test.go"
Cohesion: 0.19
Nodes (16): adminCfg(), TestCreateNamespaceApplyFails(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree(), TestCreateSecretsFailsWithoutAdminFields(), TestCreateSecretsPreflight(), TestCreateSecretsStopsOnPreflightFailure(), TestDeleteNamespaceProtected() (+8 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "newTestOps"
Cohesion: 0.06
Nodes (67): Ops, matchCLI(), newTestOps(), outputForRole(), ranContains(), TestCountContains(), TestDiagnostics(), TestDisableDefaultUsers() (+59 more)

### Community 23 - ".Run"
Cohesion: 0.20
Nodes (19): testing.M, runCtr(), TestConfigStepsDoNotLeakSecrets(), TestCtrConfigDryRun(), TestCtrDiagnosticsDryRun(), TestCtrErrorPaths(), TestCtrExecCLIPathSeparator(), TestCtrLoginOutcomes() (+11 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.15
Nodes (41): cliScriptPath(), curlCalls(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestDetectRoleAddrsError() (+33 more)

### Community 25 - ".releaseToBackup"
Cohesion: 0.13
Nodes (14): field(), showRedundancyLocalScript(), defaultLocalAddrs(), Ops, hostMatches(), shortHost(), TestDefaultLocalAddrs(), TestMateActivityStateReadsTheMateColumn() (+6 more)

### Community 26 - "renderDriver"
Cohesion: 0.18
Nodes (19): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), shQuote() (+11 more)

### Community 27 - "Platform"
Cohesion: 0.15
Nodes (30): checkFlagPlatforms(), commandPlatforms(), declaredList(), flagOnlyOn(), App, parsePlatformList(), platformSuffix(), prepare() (+22 more)

### Community 28 - "commanddoc_test.go"
Cohesion: 0.16
Nodes (23): shorthand, github.com/spf13/pflag.FlagSet, mdRow(), renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeAbbrevTable() (+15 more)

### Community 31 - "validate.go"
Cohesion: 0.12
Nodes (20): keyValueEntries, BrokerPodSuffixShape(), checkCredentialChars(), foldToEnvVar(), Config, missingErr(), platformKey(), requireAll() (+12 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "importops_test.go"
Cohesion: 0.08
Nodes (37): planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, minimalCapture() (+29 more)

### Community 35 - "Role"
Cohesion: 0.07
Nodes (23): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+15 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "haCfg"
Cohesion: 0.10
Nodes (24): lbServiceName(), haCfg(), saCfg(), TestHARoles(), TestPodNameSuffixMatchesTheConfigBound(), TestProductKeyRoles(), TestResourceNames(), TestRestartOrder() (+16 more)

### Community 38 - "MateChannel"
Cohesion: 0.19
Nodes (20): PhaseKind, SwitchAction, SwitchPhase, MateChannel, confirmDemotions(), SwitchPlan, promoteConfirmed(), replPrecheckError() (+12 more)

### Community 39 - "helpers.go"
Cohesion: 0.14
Nodes (28): exportconfigReadCounter, layer, io.Reader, TestPromptYesNo(), TestStdinCanAnswerClosedFile(), registerFlagCompletion(), usagef(), addAllowCommandFlag() (+20 more)

### Community 40 - "statusreport_test.go"
Cohesion: 0.20
Nodes (17): Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList(), TestBrokerReportWithNoBrokerDeployed() (+9 more)

### Community 41 - "cli/platform_test.go"
Cohesion: 0.16
Nodes (25): echoRunner(), App, TestConfiguredRouternameSurvivesTheFallback(), TestCtrConfirmDeclined(), logArgs(), App, logArgs2(), runPlatform() (+17 more)

### Community 42 - "InjectShutdown"
Cohesion: 0.17
Nodes (27): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+19 more)

### Community 43 - "runRootWith"
Cohesion: 0.13
Nodes (33): runRootWith(), TestCLICommand(), TestDeployOperatorNoPromptStaysUnknownFlag(), TestK8sConfigDeleteDomainCertsConfigured(), TestK8sConfirmDeclined(), TestK8sPromptsNameNamespaceAndContext(), TestK8sPromptsOmitAnUnknownContext(), TestK8sRestartConfirmGate() (+25 more)

### Community 44 - "NewMateChannel"
Cohesion: 0.24
Nodes (22): Runner, execArgs(), NewMateChannel(), ReadSecretKey(), secretCommand(), hasPrefixArgv(), indexOf(), operandOf() (+14 more)

### Community 45 - "opRunner"
Cohesion: 0.12
Nodes (22): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch(), opCanI() (+14 more)

### Community 46 - "broker/replication.go"
Cohesion: 0.17
Nodes (23): colSpan, MateConfig, cliTransport(), containsEndpoint(), dashSpans(), flagByte(), gutterClear(), labelValue() (+15 more)

### Community 47 - ".ConfigureReplication"
Cohesion: 0.15
Nodes (16): QueueState, VPNRepl, Ops, isRunCLIRejection(), mateConvergenceShutdowns(), missingListedVPNs(), newlineIf(), reenableLine() (+8 more)

### Community 48 - "testing.T"
Cohesion: 0.06
Nodes (64): testing.T, writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultUsersShowVPNError(), TestDisableDefaultVPNDisableError() (+56 more)

### Community 49 - "Ops"
Cohesion: 0.10
Nodes (15): time.Duration, cliRunNames(), countContains(), Ops, runCLISkeleton(), validName(), ValidVPNName(), validVPNName() (+7 more)

### Community 50 - "runner_test.go"
Cohesion: 0.14
Nodes (22): captureStdout(), helperCommand(), TestEchoDefaultWriter(), TestEchoOutput(), TestEchoOutputInput(), TestEchoRun(), TestEchoRunEnv(), TestEchoRunEnvNoEnv() (+14 more)

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (32): TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias(), TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRegisteredWhereItExecutes() (+24 more)

### Community 52 - "cli/replication_test.go"
Cohesion: 0.26
Nodes (17): mateChannel(), mateSEMPPassword(), reportReplicationConfig(), App, replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer() (+9 more)

### Community 53 - "container/inspect_test.go"
Cohesion: 0.26
Nodes (11): inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer() (+3 more)

### Community 54 - "Operations"
Cohesion: 0.10
Nodes (21): Bringing up a fresh cluster, Configuring a site, Data replication, Docker and Podman mechanics, Every destructive command confirms, Exit codes, Extra CLI users differ by platform, Operations (+13 more)

### Community 55 - "Test catalogue"
Cohesion: 0.11
Nodes (17): abbrev_test.go, convert_test.go, Coverage, examples_test.go, internal/abbrev, internal/convert, internal/examples, internal/output (+9 more)

### Community 56 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 57 - "DiffBlocks"
Cohesion: 0.11
Nodes (36): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, ancestorReported(), blockLabel(), DiffBlocks() (+28 more)

### Community 58 - ".rpc"
Cohesion: 0.07
Nodes (34): AdminState, Credential, mateTarget, sempExecuteResult, sempMate, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply (+26 more)

### Community 61 - "checkreport.go"
Cohesion: 0.15
Nodes (21): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), allCustomMounted(), containsString(), failRow() (+13 more)

### Community 62 - "render.go"
Cohesion: 0.11
Nodes (26): LoadBalancer, NodeIdentity, boolStr(), containerSecretSpecs(), cut(), EnvPairs(), escapePercent(), ContainerSecret (+18 more)

### Community 63 - "replicationops_test.go"
Cohesion: 0.27
Nodes (17): hasCall(), TestDisableDefaultUsersNoVPNs(), TestReleaseToBackupReleasedTimeout(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), TestConfigureReplicationClosingShowFailureStillReportsTheApply() (+9 more)

### Community 65 - "Configuration"
Cohesion: 0.17
Nodes (12): Bring your own TLS Secret, Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling, Scaling tiers (+4 more)

### Community 66 - "newCapMgr"
Cohesion: 0.15
Nodes (26): newCapMgr(), TestManagerCopy(), TestManagerDeleteDockerComposeDownError(), TestManagerDeletePodmanDaemonReloadError(), TestManagerDeletePodmanRemovesSecrets(), TestManagerDeletePodmanStopFailsServiceInactiveProceeds(), TestManagerDeletePurgeError(), TestManagerDeployDockerComposeUpError() (+18 more)

### Community 67 - "broker/replication_test.go"
Cohesion: 0.15
Nodes (30): ParseShowReplication(), ParseVPNReplication(), RenderMate(), loadShowReplication(), mateAppliance(), mateSoftware(), TestMateRoundTripsThroughItsOwnGrammar(), TestParseShowReplicationAppliance() (+22 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "ParseBlocks"
Cohesion: 0.15
Nodes (15): BrokerType, ParseBlocks(), splitLines(), bannerType(), checkProvenance(), checkSameType(), describeScope(), Ops (+7 more)

### Community 71 - "localCfg"
Cohesion: 0.14
Nodes (26): TestLocalMateRunsThroughOps(), sempV1OK(), assertNoPasswordInArgv(), TestLocalAdminIsTheDeploymentsOwnLogin(), TestMateRevertActivityBridgePlaintextOnly(), TestMateRevertActivityBridgeTLSNoCA(), TestMateRevertActivityCredsAndBodyNeverInArgv(), TestMateRevertActivityNon2xx() (+18 more)

### Community 72 - "output_test.go"
Cohesion: 0.24
Nodes (14): sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLevelOrderKeepsSkipBelowFail(), TestLineAddsNoPrefix(), TestSectionPadsToWidth() (+6 more)

### Community 73 - "internal/k8s"
Cohesion: 0.11
Nodes (19): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go, names_test.go (+11 more)

### Community 74 - "annotate_test.go"
Cohesion: 0.09
Nodes (43): Annotate(), applyMetaField(), Capture, isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV() (+35 more)

### Community 75 - "Set"
Cohesion: 0.14
Nodes (14): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+6 more)

### Community 76 - "namespace_test.go"
Cohesion: 0.30
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 77 - "io.Writer"
Cohesion: 0.29
Nodes (7): interactiveFailRunner, io.Writer, Echo, MaskEnv(), NewExec(), Quote(), quoteTok()

### Community 78 - "validContainerConfig"
Cohesion: 0.11
Nodes (24): TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes(), TestAdditionalUserNamesCollideOnDocker(), TestValidateAdditionalUserClashesWithABuiltIn(), TestValidateContainerBadNetworkMode(), TestValidateContainerBridge(), TestValidateContainerHA(), TestValidateContainerIdentifiers() (+16 more)

### Community 79 - "switchplan_test.go"
Cohesion: 0.20
Nodes (25): SiteState, PlannedRoles(), RoleAtSite(), TestPlannedRolesNamesEveryListedVPN(), TestRoleAtSiteIsTheComplement(), BuildSwitchPlan(), ExecuteSwitchPlan(), actionStrings() (+17 more)

### Community 80 - "scaling_test.go"
Cohesion: 0.11
Nodes (19): scalingTier, containerMem(), Config, Config, setContainerMem(), TestApplyScalingTierDefaultsContainerBlocks(), TestApplyScalingTierDefaultsK8s(), TestApplyScalingTierDefaultsMemOverride() (+11 more)

### Community 81 - "blocks.go"
Cohesion: 0.14
Nodes (23): Omission, Region, token, brokerTypeFromSchema(), firstQuoted(), Capture, isIndented(), isPragma() (+15 more)

### Community 82 - "age"
Cohesion: 0.37
Nodes (4): orNone(), age(), roleRank(), Cluster

### Community 83 - "k8s/inspect_test.go"
Cohesion: 0.14
Nodes (24): brokerConditionRows(), conditionLevel(), findCondition(), podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture() (+16 more)

### Community 84 - "execguard_test.go"
Cohesion: 0.07
Nodes (38): TestCanonicalDuration(), decodeStrict(), Config, guardCommandOf(), guardConfig(), setGuardCommand(), TestAllowCommandIsNotASchemaKey(), TestAllowCommandsAccepts() (+30 more)

### Community 85 - "broker/matechannel_test.go"
Cohesion: 0.20
Nodes (18): CLIRunner, fakeRun, Ops, NewCLIMate(), mateReplies(), newTestMate(), TestMateChannelDescribeNamesTheTarget(), TestMateChannelPreflightLearnsTheBrokerType() (+10 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.13
Nodes (20): time.Time, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes(), operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+12 more)

### Community 87 - ".preflightOne"
Cohesion: 0.44
Nodes (3): canIAnswer(), Cluster, probe

### Community 88 - "github.com/spf13/cobra.Command"
Cohesion: 0.21
Nodes (46): github.com/spf13/cobra.Command, addCommands(), addLogFlags(), App, group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerConfigureCmd() (+38 more)

### Community 89 - "load"
Cohesion: 0.12
Nodes (38): BrokerCR(), Compose(), composeEscape(), ComposeProject(), ContainerSecrets(), SecretPreflight(), splitPair(), assertNoCheckoutPath() (+30 more)

### Community 90 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.11
Nodes (34): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+26 more)

### Community 92 - "storage_test.go"
Cohesion: 0.33
Nodes (9): Config, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 94 - "containerState"
Cohesion: 0.21
Nodes (10): containerState, healthState, decodeInspect(), Manager, orUnknown(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings() (+2 more)

### Community 95 - "strings.Builder"
Cohesion: 0.18
Nodes (17): WeightedNodeTerm, strings.Builder, NodeAffinity, NodeMatchExpr, Placement, PodAffinityTerm, parseToleration(), sortedKeys() (+9 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "platformOps"
Cohesion: 0.36
Nodes (17): opFunc, roleOpFunc, addPodFlag(), newBrokerCopyCmd(), newBrokerDeployCmd(), newBrokerExecCmd(), newBrokerGenerateCmd(), newBrokerLogsCmd() (+9 more)

### Community 100 - "tlsCfg"
Cohesion: 0.43
Nodes (7): Config, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsMakeTheSecretOurs(), TestSuppliedCertsNeedASecretName(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), tlsCfg()

### Community 101 - "scripts.go"
Cohesion: 0.07
Nodes (46): showCmd, validCLILine(), Ops, assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript() (+38 more)

### Community 102 - "opPerformReplication"
Cohesion: 0.22
Nodes (17): mateChannelFunc, bufio.Reader, App, step(), confirmReplicationConfig(), confirmReplicationSwitch(), App, opConfigureReplication() (+9 more)

### Community 103 - "container/secrets_test.go"
Cohesion: 0.21
Nodes (18): ResolveSecretValues(), certFixture(), TestDeletePodmanRemovesTheBundleAfterTheUnit(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeWritingTheComposeFile(), TestDeployDockerPassesTheBundleAsEnvNotArgv(), TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(), TestDeployPodmanFailsWhenTheBundleCannotBeWritten() (+10 more)

### Community 104 - "watch_test.go"
Cohesion: 0.13
Nodes (21): desiredWatch(), Cluster, splitWatch(), subtractWatch(), deployJSON(), Cluster, TestDesiredWatchAppendsTheBrokerNamespace(), TestOperatorReleaseNarrowsInsteadOfRemoving() (+13 more)

### Community 105 - "internal/cli"
Cohesion: 0.17
Nodes (12): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, examples_test.go, exit_test.go (+4 more)

### Community 106 - "Load"
Cohesion: 0.23
Nodes (16): minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadReadError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors() (+8 more)

### Community 107 - "operatorversion.go"
Cohesion: 0.24
Nodes (9): compareVersions(), Cluster, imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestImageTag() (+1 more)

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "ReplSite"
Cohesion: 0.25
Nodes (13): Replication, ReplSite, Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded() (+5 more)

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "exit_test.go"
Cohesion: 0.10
Nodes (25): childExitError, runFailRunner, usageError, github.com/spf13/cobra.PositionalArgs, asUsage(), childExit(), childStatus(), ExitCode() (+17 more)

### Community 112 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, File descriptors on rootless podman, Import reported failure, Podman secret flags, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 113 - "internal/config"
Cohesion: 0.17
Nodes (12): command_test.go, config_test.go, duration_test.go, execguard_test.go, hostpath_test.go, internal/config, platform_test.go, redundancy_test.go (+4 more)

### Community 114 - "newSEMPTestMate"
Cohesion: 0.29
Nodes (13): newSEMPTestMate(), sempReply(), sempSite(), TestSEMPMateChecksExecuteResult(), TestSEMPMateEscapesTheVPNName(), TestSEMPMatePlainTextURL(), TestSEMPMatePrimaryActive(), TestSEMPMateReadsBothMateShapes() (+5 more)

### Community 115 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 116 - "operatorversion_test.go"
Cohesion: 0.39
Nodes (7): operatorDeployJSON(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsSilentInAPreview(), TestConfirmNoDowngradeProceedsWhenConfirmAccepts(), TestConfirmNoDowngradeProceedsWhenVersionReadFails(), TestConfirmNoDowngradeRefusedWhenConfirmDeclines(), TestConfirmNoDowngradeRefusesByDefault()

### Community 117 - "internal/container"
Cohesion: 0.29
Nodes (7): inspect_test.go, internal/container, manager_test.go, preflight_test.go, runtime_test.go, secrets_test.go, transport_test.go

### Community 119 - "load.go"
Cohesion: 0.20
Nodes (12): TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config, parseError() (+4 more)

### Community 120 - "blocks_test.go"
Cohesion: 0.26
Nodes (12): TestAnnotateRoundTripCaptureIsMarked(), blockByIndex(), Capture, loadSample(), TestParseBlocksCRLFMatchesLF(), TestParseBlocksHeaderAndPreamble(), TestParseBlocksIndentedLineWithNoOpenerRefused(), TestParseBlocksMissingEndTerminatorRefused() (+4 more)

### Community 121 - "NewManager"
Cohesion: 0.26
Nodes (12): NewManager(), TestManagerNilSinks(), Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime() (+4 more)

### Community 123 - "replicationVPNLines"
Cohesion: 0.38
Nodes (10): replicationVPNLines(), siteAEntry(), TestNothingEverShutsDownTheVPNItself(), TestReplicationVPNLinesEnablesListedAndDisablesTheRest(), TestReplicationVPNLinesIsDeterministic(), TestReplicationVPNLinesQuotesNamesWithSpaces(), TestReplicationVPNLinesReEnablesWhatPhase1StoppedOnly(), TestReplicationVPNLinesSetsRoleWithoutCyclingReplication() (+2 more)

### Community 124 - "resolve_test.go"
Cohesion: 0.28
Nodes (8): TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveMissingBinaryIsActionable(), TestResolveRefusesCurrentDirectory(), verboseExec(), Resolve()

### Community 125 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 126 - "rootlessNoFileMgr"
Cohesion: 0.25
Nodes (8): bytes.Buffer, Manager, rootlessNoFileMgr(), TestPrepHostRootlessNoFileSufficient(), TestPrepHostRootlessNoFileTooLow(), TestPrepHostRootlessNoFileUnlimited(), TestPrepHostRootlessNoFileUnreadable(), TestPrepHostRootlessNoFileUnsetSkips()

### Community 128 - "runConvert"
Cohesion: 0.83
Nodes (3): App, newConvertCmd(), runConvert()

### Community 130 - "ReplRole"
Cohesion: 0.50
Nodes (3): ReplRole, ReplicationConfigResult, setReplicationRoleScript()

### Community 131 - "runExample"
Cohesion: 0.83
Nodes (3): App, newExamplesCmd(), runExample()

### Community 132 - "internal/engine"
Cohesion: 0.67
Nodes (3): internal/engine, resolve_test.go, runner_test.go

### Community 135 - "ParseRole"
Cohesion: 0.36
Nodes (7): TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), TestRoleNames(), ParseRole(), RoleAbbrev(), RoleNames()

## Knowledge Gaps
- **213 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+208 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 319 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **9 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Config` connect `Config` to `bg`, `cli_test.go`, `context.Context`, `Manager`, `operator.go`, `capRunner`, `NewCluster`, `convert_test.go`, `ctrCfg`, `k8s/secrets_test.go`, `prep_test.go`, `newTestOps`, `kubectlTransport`, `Role`, `haCfg`, `NewMateChannel`, `opRunner`, `Ops`, `.rpc`, `containerTransport`, `checkreport.go`, `render.go`, `newCapMgr`, `localCfg`, `namespace_test.go`, `load`, `strings.Builder`, `opPerformReplication`, `container/secrets_test.go`, `watch_test.go`, `ReplSite`, `Cluster`, `NewManager`?**
  _High betweenness centrality (0.042) - this node is a cross-community bridge._
- **Why does `Platform` connect `Platform` to `config_test.go`, `Manager`, `Config`, `Command`, `capRunner`, `convert_test.go`, `ctrCfg`, `validate.go`, `opRunner`, `Ops`, `.rpc`, `containerTransport`, `newCapMgr`, `validContainerConfig`, `scaling_test.go`, `execguard_test.go`, `github.com/spf13/cobra.Command`, `load`, `platformOps`, `opPerformReplication`, `container/secrets_test.go`, `Load`, `load.go`, `NewManager`?**
  _High betweenness centrality (0.038) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `context.Context`, `Manager`, `Config`, `ParseRole`, `importops.go`, `newTestOps`, `.releaseToBackup`, `renderDriver`, `validate.go`, `kubectlTransport`, `.ConfigureReplication`, `Ops`, `.rpc`, `containerTransport`, `ParseBlocks`, `Set`, `broker/matechannel_test.go`, `scripts.go`, `opPerformReplication`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 34 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 34 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _213 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `bg` be split into smaller, more focused modules?**
  _Cohesion score 0.07053429730206313 - nodes in this community are weakly interconnected._
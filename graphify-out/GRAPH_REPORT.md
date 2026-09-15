# Graph Report - solace-k8-scripts  (2026-09-15)

## Corpus Check
- 186 files · ~541,821 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 21 file(s) not represented in the graph (top: .golden 15, (none) 2, .cli 2)

## Summary
- 3454 nodes · 10843 edges · 126 communities (117 shown, 5 thin omitted)
- Extraction: 81% EXTRACTED · 19% INFERRED · 0% AMBIGUOUS · INFERRED: 2007 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `0a2063b4`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- runRootWith
- Sink
- context.Context
- commanddoc_test.go
- config_test.go
- Commands
- Manager
- config.go
- GenOperator
- Command
- capRunner
- NewCluster
- eqArgs
- convert_test.go
- Block
- newCapMgr
- importops.go
- k8s/secrets_test.go
- dev.sh
- prep_test.go
- dev.ps1
- .ConfigureReplication
- switchplan_test.go
- verify_local_test.go
- .releaseToBackup
- renderDriver
- exportconfig_test.go
- haCfg
- solace
- validate.go
- CLAUDE.md
- importops_test.go
- domaincerts_resolve_test.go
- Role
- .resolveSecretRefs
- recRunner
- ops_container.go
- replicationops_test.go
- statusreport_test.go
- cli/platform_test.go
- InjectShutdown
- usagef
- captureStdout
- render.go
- Platform
- broker/replication.go
- Get
- Ops
- runner_test.go
- newRootCmd
- confirmAction
- broker/matechannel_test.go
- Operations
- Test catalogue
- Exporting and importing configuration
- DiffBlocks
- .MateRevertActivity
- Cluster
- ReplRole
- checkreport.go
- completion_test.go
- command_test.go
- Configuration
- BrokerType
- broker/replication_test.go
- Command reference
- Developer guide
- blocks.go
- localCfg
- output_test.go
- internal/k8s
- annotate_test.go
- Set
- namespace_test.go
- testing.T
- redundancy_test.go
- kubectlTransport
- scaling.go
- blocks_test.go
- age
- k8s/inspect_test.go
- wireExec
- hostpath_test.go
- k8s/inspect.go
- .preflightOne
- github.com/spf13/cobra.Command
- load
- judge
- RuleFor
- exit_test.go
- cli_test.go
- container/inspect_test.go
- ReplSite
- Abbreviations
- containerTransport
- tlsCfg
- scripts.go
- App
- ctrCfg
- watch_test.go
- internal/cli
- operatorversion.go
- k8s/runtime_test.go
- What `import-config` applies
- logArgs
- internal/broker
- pullSecretCfg
- Troubleshooting
- internal/config
- applyAliases
- completion.go
- internal/container
- ParseBlocks
- NewManager
- Compose
- Fixtures and doubles
- storage_test.go
- ParseRole
- rootlessNoFileMgr
- Config
- broker.go

## God Nodes (most connected - your core abstractions)
1. `newTestOps()` - 128 edges
2. `Role` - 116 edges
3. `ctrCfg()` - 115 edges
4. `newCapMgr()` - 100 edges
5. `Config` - 95 edges
6. `NewCluster()` - 92 edges
7. `loadK8s()` - 85 edges
8. `matchCLI()` - 80 edges
9. `Platform` - 73 edges
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

## Communities (126 total, 5 thin omitted)

### Community 0 - "bg"
Cohesion: 0.15
Nodes (50): bg(), domainCANames(), App, k8sCluster(), k8sContext(), k8sLogin(), k8sOps(), k8sWhat() (+42 more)

### Community 1 - "runRootWith"
Cohesion: 0.11
Nodes (45): testing.M, allowRuntime(), echoRunner(), App, runCtr(), runRootWith(), TestAnnounceCommandsNamesResolvedBinaries(), TestCLICommand() (+37 more)

### Community 2 - "Sink"
Cohesion: 0.21
Nodes (3): KV, Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.06
Nodes (10): scriptedMate, context.Context, Manager, Cluster, Cluster, Cluster, Cluster, Cluster (+2 more)

### Community 4 - "commanddoc_test.go"
Cohesion: 0.16
Nodes (23): shorthand, github.com/spf13/pflag.FlagSet, mdRow(), renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeAbbrevTable() (+15 more)

### Community 5 - "config_test.go"
Cohesion: 0.04
Nodes (98): assertContainerBlockDefaults(), assertContainerScaling(), envTree(), Config, haNodesConfig(), minimalK8s(), TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes() (+90 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (10): composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing(), solaceRows() (+2 more)

### Community 8 - "config.go"
Cohesion: 0.07
Nodes (32): AdditionalUser, Broker, CertDir, Container, ContainerSecurity, DockerConfig, DomainCerts, Network (+24 more)

### Community 9 - "GenOperator"
Cohesion: 0.16
Nodes (22): GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator(), renderOperatorNS(), splitAfterNamespace() (+14 more)

### Community 10 - "Command"
Cohesion: 0.07
Nodes (45): commandRules, Command, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules() (+37 more)

### Community 11 - "capRunner"
Cohesion: 0.18
Nodes (14): capCall, capRunner, New(), TestNewDefaults(), Transport, NewTransport(), dockerCfg(), podmanCfg() (+6 more)

### Community 12 - "NewCluster"
Cohesion: 0.09
Nodes (59): TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts(), TestValidateReportsTheDerivedImagePullSecretName() (+51 more)

### Community 13 - "eqArgs"
Cohesion: 0.17
Nodes (23): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), TestCLIAndShellAreInteractive() (+15 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (63): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+55 more)

### Community 15 - "Block"
Cohesion: 0.09
Nodes (54): Block, TargetState, regexp.Regexp, importIgnore(), injectedBlock(), BridgeEnablementInverted(), ClearExistingClientCAs(), ClearExistingNested() (+46 more)

### Community 16 - "newCapMgr"
Cohesion: 0.07
Nodes (77): fileExists(), assertMode(), containsStr(), hasCall(), maskedKeys(), newCapMgr(), TestContainerRunningMatchesNameExactly(), TestManagerCheckDNSFailsLoudInHA() (+69 more)

### Community 17 - "importops.go"
Cohesion: 0.19
Nodes (18): ImportResult, PlannedSection, defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList(), renderRegionScript() (+10 more)

### Community 18 - "k8s/secrets_test.go"
Cohesion: 0.12
Nodes (31): GenSecrets(), TestUpdateServerCertSecret(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), checkGolden(), decodeDataValue() (+23 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "prep_test.go"
Cohesion: 0.18
Nodes (17): GenBroker(), joinManifests(), namespaceManifest(), adminCfg(), TestCreateNamespaceApplyFails(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree(), TestCreateSecretsPreflight() (+9 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - ".ConfigureReplication"
Cohesion: 0.15
Nodes (17): AdminState, QueueState, VPNRepl, Ops, isRunCLIRejection(), mateConvergenceShutdowns(), missingListedVPNs(), newlineIf() (+9 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.09
Nodes (52): PhaseKind, SiteState, SwitchAction, SwitchPhase, Replication, MateChannel, BuildSwitchPlan(), confirmDemotions() (+44 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.13
Nodes (43): cliScriptPath(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestDetectRoleAddrsError(), TestDetectRoleAmbiguousAddress() (+35 more)

### Community 25 - ".releaseToBackup"
Cohesion: 0.11
Nodes (17): field(), TestField(), TestLastLines(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), TestLastLinesEqualCount(), showRedundancyLocalScript(), defaultLocalAddrs() (+9 more)

### Community 26 - "renderDriver"
Cohesion: 0.18
Nodes (19): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), shQuote() (+11 more)

### Community 27 - "exportconfig_test.go"
Cohesion: 0.15
Nodes (19): exportconfigReadCounter, opRunner, exportconfigDriverOutput(), exportconfigTransportOutput(), exportconfigVPNListOutput(), App, newExportconfigRunner(), newExportconfigRunnerRejectingChunk() (+11 more)

### Community 28 - "haCfg"
Cohesion: 0.14
Nodes (21): TestCheckStopsProbingWhenUnreachable(), TestResolveStorageClass(), TestStorageRows(), lbServiceName(), haCfg(), saCfg(), TestHARoles(), TestPodNameSuffixMatchesTheConfigBound() (+13 more)

### Community 31 - "validate.go"
Cohesion: 0.07
Nodes (27): keyValueEntries, expandTilde(), expandTildeToken(), Config, IsAbsHostPath(), TestExpandTilde(), TestIsAbsHostPath(), tildeHome() (+19 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "importops_test.go"
Cohesion: 0.13
Nodes (26): planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, minimalCapture() (+18 more)

### Community 34 - "domaincerts_resolve_test.go"
Cohesion: 0.07
Nodes (50): DirReader, fakeDirEntry, os.DirEntry, os.FileInfo, os.FileMode, caNameSafe(), checkNoDotDot(), DefaultDirReader() (+42 more)

### Community 35 - "Role"
Cohesion: 0.07
Nodes (26): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+18 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.15
Nodes (10): TestCanIAnswerReadsTheLastLine(), NewTransport(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest() (+2 more)

### Community 38 - "ops_container.go"
Cohesion: 0.14
Nodes (41): TestCtrManagerConfirmWiring(), wantRemove(), containerRenderRole(), containerRole(), containerWhat(), ctrLogin(), ctrManager(), ctrOps() (+33 more)

### Community 39 - "replicationops_test.go"
Cohesion: 0.18
Nodes (27): hasCall(), replicationVPNLines(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), siteAEntry(), TestConfigureReplicationClosingShowFailureStillReportsTheApply() (+19 more)

### Community 40 - "statusreport_test.go"
Cohesion: 0.20
Nodes (17): Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList(), TestBrokerReportWithNoBrokerDeployed() (+9 more)

### Community 41 - "cli/platform_test.go"
Cohesion: 0.29
Nodes (16): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+8 more)

### Community 42 - "InjectShutdown"
Cohesion: 0.17
Nodes (27): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+19 more)

### Community 43 - "usagef"
Cohesion: 0.09
Nodes (29): childExitError, usageError, github.com/spf13/cobra.PositionalArgs, noRolePositional(), registerFlagCompletion(), App, newConvertCmd(), runConvert() (+21 more)

### Community 44 - "captureStdout"
Cohesion: 0.14
Nodes (21): opCall, os.File, capture(), captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch() (+13 more)

### Community 45 - "render.go"
Cohesion: 0.17
Nodes (24): strings.Builder, PodAffinityTerm, boolStr(), containerSecretSpecs(), cut(), parsePort(), parseToleration(), sortedKeys() (+16 more)

### Community 46 - "Platform"
Cohesion: 0.22
Nodes (21): checkFlagPlatforms(), commandPlatforms(), declaredList(), App, parsePlatformList(), platformSuffix(), prepare(), promptPlatform() (+13 more)

### Community 47 - "broker/replication.go"
Cohesion: 0.17
Nodes (22): colSpan, MateConfig, cliTransport(), containsEndpoint(), dashSpans(), flagByte(), gutterClear(), labelValue() (+14 more)

### Community 48 - "Get"
Cohesion: 0.16
Nodes (24): Example, TestExamplesEmitsToStdout(), DetectPlatforms(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint(), TestDetectPlatformsMissingFile(), TestParsePlatform(), TestPlatforms() (+16 more)

### Community 49 - "Ops"
Cohesion: 0.11
Nodes (16): time.Duration, Ops, TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), confirmImport(), App, pluralVPN() (+8 more)

### Community 50 - "runner_test.go"
Cohesion: 0.05
Nodes (63): Exec, os/exec.Cmd, CLIArg(), cliArg(), CLIScriptPath(), TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose() (+55 more)

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (30): TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias(), TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRegisteredWhereItExecutes() (+22 more)

### Community 52 - "confirmAction"
Cohesion: 0.25
Nodes (21): layer, io.Reader, io.Writer, TestConfirmDeleteShortcut(), TestPromptYes(), TestPromptYesNo(), TestStdinCanAnswerClosedFile(), confirmAction() (+13 more)

### Community 53 - "broker/matechannel_test.go"
Cohesion: 0.20
Nodes (18): fakeRun, Ops, NewCLIMate(), mateReplies(), newTestMate(), TestLocalMateRunsThroughOps(), TestMateChannelDescribeNamesTheTarget(), TestMateChannelPreflightLearnsTheBrokerType() (+10 more)

### Community 54 - "Operations"
Cohesion: 0.10
Nodes (21): Bringing up a fresh cluster, Configuring a site, Data replication, Docker and Podman mechanics, Every destructive command confirms, Exit codes, Extra CLI users differ by platform, Operations (+13 more)

### Community 55 - "Test catalogue"
Cohesion: 0.10
Nodes (20): abbrev_test.go, convert_test.go, Coverage, examples_test.go, internal/abbrev, internal/convert, internal/engine, internal/examples (+12 more)

### Community 56 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 57 - "DiffBlocks"
Cohesion: 0.11
Nodes (36): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, ancestorReported(), blockLabel(), DiffBlocks() (+28 more)

### Community 58 - ".MateRevertActivity"
Cohesion: 0.11
Nodes (23): Credential, mateTarget, TestHTTPStatusHelpers(), anyHTTP2xx(), bridgeHostPort(), curlConfigFlag(), curlConfigLine(), Ops (+15 more)

### Community 60 - "ReplRole"
Cohesion: 0.14
Nodes (12): ReplRole, sempExecuteResult, sempMate, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, encoding/xml.Name, rejectionIn() (+4 more)

### Community 61 - "checkreport.go"
Cohesion: 0.15
Nodes (20): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), containsString(), failRow(), Cluster (+12 more)

### Community 62 - "completion_test.go"
Cohesion: 0.25
Nodes (15): runComplete(), TestAllowCommandOffersNoFiles(), TestCompletionHelpStillWorks(), TestCompletionNeedsAShell(), TestCompletionNoDescriptions(), TestCompletionScriptsGenerate(), TestDirFlagCompletesDirectories(), TestEnvFlagCompletesEnvFiles() (+7 more)

### Community 63 - "command_test.go"
Cohesion: 0.15
Nodes (14): decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRenamedRuntimeKeysFailLoud() (+6 more)

### Community 65 - "Configuration"
Cohesion: 0.17
Nodes (12): Bring your own TLS Secret, Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling, Scaling tiers (+4 more)

### Community 66 - "BrokerType"
Cohesion: 0.23
Nodes (7): BrokerType, cliMate, CLIRunner, bannerType(), checkSameType(), showReplicationScript(), showVPNReplicationScript()

### Community 67 - "broker/replication_test.go"
Cohesion: 0.15
Nodes (32): ParseShowReplication(), ParseVPNReplication(), RenderMate(), SameEndpoints(), loadShowReplication(), mateAppliance(), mateSoftware(), TestMateRoundTripsThroughItsOwnGrammar() (+24 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "blocks.go"
Cohesion: 0.12
Nodes (26): Omission, Region, token, markerRegion(), TestMarkerRegionBeginAndEnd(), brokerTypeFromSchema(), firstQuoted(), Capture (+18 more)

### Community 71 - "localCfg"
Cohesion: 0.11
Nodes (38): sempV1OK(), assertNoPasswordInArgv(), TestLocalAdminIsTheDeploymentsOwnLogin(), TestMateRevertActivityBridgePlaintextOnly(), TestMateRevertActivityBridgeTLSNoCA(), TestMateRevertActivityCredsAndBodyNeverInArgv(), TestMateRevertActivityNon2xx(), TestMateRevertActivityRPCNotOK() (+30 more)

### Community 72 - "output_test.go"
Cohesion: 0.19
Nodes (17): NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLevelOrderKeepsSkipBelowFail(), TestLineAddsNoPrefix() (+9 more)

### Community 73 - "internal/k8s"
Cohesion: 0.11
Nodes (19): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go, names_test.go (+11 more)

### Community 74 - "annotate_test.go"
Cohesion: 0.09
Nodes (42): Annotate(), applyMetaField(), Capture, isMarker(), markerSection(), markerVerb(), parseKV(), readMarker() (+34 more)

### Community 75 - "Set"
Cohesion: 0.14
Nodes (14): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+6 more)

### Community 76 - "namespace_test.go"
Cohesion: 0.30
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 77 - "testing.T"
Cohesion: 0.06
Nodes (109): testing.T, Ops, matchCLI(), newTestOps(), ranContains(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs() (+101 more)

### Community 78 - "redundancy_test.go"
Cohesion: 0.21
Nodes (12): Config, pskCfg(), TestFillStandaloneNodeNameIgnoresAnEmptyHostname(), TestFillStandaloneNodeNameKeepsAConfiguredName(), TestFillStandaloneNodeNameNeverFillsHA(), TestFillStandaloneNodeNameUsesTheHost(), TestNothingGeneratesThePSK(), TestPSKEnvSatisfiesTheContainerRequirement() (+4 more)

### Community 80 - "scaling.go"
Cohesion: 0.16
Nodes (14): scalingKey, scalingSpelling, scalingTier, containerMem(), Config, Scaling, Scaling, yaml.Node (+6 more)

### Community 81 - "blocks_test.go"
Cohesion: 0.26
Nodes (12): TestAnnotateIsDeterministic(), blockByIndex(), Capture, loadSample(), TestParseBlocksCRLFMatchesLF(), TestParseBlocksHeaderAndPreamble(), TestParseBlocksIndentedLineWithNoOpenerRefused(), TestParseBlocksMissingEndTerminatorRefused() (+4 more)

### Community 82 - "age"
Cohesion: 0.22
Nodes (9): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+1 more)

### Community 83 - "k8s/inspect_test.go"
Cohesion: 0.15
Nodes (19): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+11 more)

### Community 84 - "wireExec"
Cohesion: 0.23
Nodes (14): opFunc, roleOpFunc, newBrokerDeployCmd(), newBrokerExecCmd(), newBrokerGenerateCmd(), newBrokerLogsCmd(), newBrokerStatusCmd(), withPodRole() (+6 more)

### Community 85 - "hostpath_test.go"
Cohesion: 0.11
Nodes (23): TestCanonicalDuration(), guardConfig(), CheckHostPath(), Config, setContainerName(), TestCheckHostPathAccepts(), TestCheckHostPathRejects(), TestContainerCertRequiresKey() (+15 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.14
Nodes (20): time.Time, brokerConditionRows(), conditionLevel(), findCondition(), Cluster, normalizeToList(), TestConditionLevelDegradesSafely(), TestNormalizeToListHandlesBothKubectlShapes() (+12 more)

### Community 87 - ".preflightOne"
Cohesion: 0.44
Nodes (3): canIAnswer(), Cluster, probe

### Community 88 - "github.com/spf13/cobra.Command"
Cohesion: 0.21
Nodes (51): github.com/spf13/cobra.Command, addCommands(), addLogFlags(), addPodFlag(), App, group(), newBrokerCLICmd(), newBrokerCmd() (+43 more)

### Community 89 - "load"
Cohesion: 0.16
Nodes (29): BrokerCR(), ContainerSecrets(), SecretPreflight(), assertNoCheckoutPath(), envLines(), healthCheckFixture(), load(), TestAdditionalUsersReachBothHalves() (+21 more)

### Community 90 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.10
Nodes (38): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+30 more)

### Community 92 - "exit_test.go"
Cohesion: 0.13
Nodes (20): interactiveFailRunner, runFailRunner, childExit(), ExitCode(), childStatusError(), k8sEnv(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitKeepsItsMessage() (+12 more)

### Community 93 - "cli_test.go"
Cohesion: 0.08
Nodes (56): captureStderr(), collectPaths(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone(), runStatusStderr(), TestBashEnvGivenToEnvFlag() (+48 more)

### Community 94 - "container/inspect_test.go"
Cohesion: 0.12
Nodes (21): containerState, healthState, decodeInspect(), Manager, orUnknown(), inspectReply(), kvValue(), TestInspectDecodesPartialOutput() (+13 more)

### Community 95 - "ReplSite"
Cohesion: 0.13
Nodes (38): mateChannelFunc, ReplicationConfigResult, PlannedRoles(), RoleAtSite(), lineSink(), progress(), step(), confirmReplicationConfig() (+30 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 100 - "tlsCfg"
Cohesion: 0.43
Nodes (7): Config, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsMakeTheSecretOurs(), TestSuppliedCertsNeedASecretName(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), tlsCfg()

### Community 101 - "scripts.go"
Cohesion: 0.07
Nodes (44): showCmd, validCLILine(), Ops, assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript() (+36 more)

### Community 102 - "App"
Cohesion: 0.39
Nodes (3): bufio.Reader, App, EnvRunner

### Community 103 - "ctrCfg"
Cohesion: 0.13
Nodes (33): ctrCfg(), newEchoMgr(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployPodmanDryRunHidesSecretBytes(), TestManagerDockerCheckProbesCompose(), TestManagerLifecycleDockerDryRunUsesCompose(), TestManagerPodmanEUIDGuardSkippedOnDryRun() (+25 more)

### Community 104 - "watch_test.go"
Cohesion: 0.13
Nodes (21): desiredWatch(), Cluster, splitWatch(), subtractWatch(), deployJSON(), Cluster, TestDesiredWatchAppendsTheBrokerNamespace(), TestOperatorReleaseNarrowsInsteadOfRemoving() (+13 more)

### Community 105 - "internal/cli"
Cohesion: 0.17
Nodes (12): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, examples_test.go, exit_test.go (+4 more)

### Community 106 - "operatorversion.go"
Cohesion: 0.24
Nodes (9): compareVersions(), Cluster, imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestImageTag() (+1 more)

### Community 107 - "k8s/runtime_test.go"
Cohesion: 0.43
Nodes (7): TestClusterHonoursRuntime(), TestExecutorRefusesUnapprovedRuntime(), TestRuntimeDefaultArgvUnchanged(), TestTransportHonoursRuntime(), unapprovedCfg(), withLeading(), wrappedCfg()

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "logArgs"
Cohesion: 0.40
Nodes (5): logArgs(), App, logArgs2(), TestLogArgsBuildsOneSetForBothPlatforms(), CanonicalDuration()

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "pullSecretCfg"
Cohesion: 0.33
Nodes (9): Config, pullSecretCfg(), TestImagePullSecretCredentialsResolveFromPassEnv(), TestImagePullSecretDerivedSuffixFitsWhereAdditionalUsersAlreadyFits(), TestImagePullSecretNameBringYourOwn(), TestImagePullSecretNameConfiguredNameWins(), TestImagePullSecretNameDerivedWhenCredentialsButNoName(), TestImagePullSecretNameEmptyWithNeither() (+1 more)

### Community 112 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, File descriptors on rootless podman, Import reported failure, Podman secret flags, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 113 - "internal/config"
Cohesion: 0.13
Nodes (15): command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go, hostpath_test.go, imagepullsecret_test.go (+7 more)

### Community 115 - "completion.go"
Cohesion: 0.31
Nodes (10): github.com/spf13/cobra.ShellCompDirective, completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), completionShell(), App, isEnvFileName() (+2 more)

### Community 117 - "internal/container"
Cohesion: 0.29
Nodes (7): inspect_test.go, internal/container, manager_test.go, preflight_test.go, runtime_test.go, secrets_test.go, transport_test.go

### Community 120 - "ParseBlocks"
Cohesion: 0.16
Nodes (13): ParseBlocks(), splitLines(), checkProvenance(), describeScope(), Ops, Capture, keepRegions(), stripTranscript() (+5 more)

### Community 121 - "NewManager"
Cohesion: 0.23
Nodes (14): NewManager(), TestManagerLogsCLIShell(), TestManagerNilSinks(), Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime() (+6 more)

### Community 124 - "Compose"
Cohesion: 0.11
Nodes (23): NodeIdentity, Compose(), composeEscape(), ComposeProject(), EnvPairs(), escapePercent(), ContainerSecret, groupKey() (+15 more)

### Community 125 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 126 - "storage_test.go"
Cohesion: 0.33
Nodes (9): Config, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 128 - "ParseRole"
Cohesion: 0.36
Nodes (7): TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), TestRoleNames(), ParseRole(), RoleAbbrev(), RoleNames()

### Community 130 - "rootlessNoFileMgr"
Cohesion: 0.22
Nodes (9): bytes.Buffer, Manager, rootlessNoFileMgr(), setNoFile(), TestPrepHostRootfulSkipsNoFile(), TestPrepHostRootlessNoFileDryRun(), TestPrepHostRootlessNoFileTooLow(), TestPrepHostRootlessNoFileUnlimited() (+1 more)

### Community 132 - "Config"
Cohesion: 0.15
Nodes (6): Image, TLS, atoiPrefix(), Config, Scaling, OperatorNamespace()

### Community 133 - "broker.go"
Cohesion: 0.15
Nodes (11): cliRunNames(), countContains(), runCLISkeleton(), TestCountContains(), TestValidName(), validName(), ValidVPNName(), validVPNName() (+3 more)

## Knowledge Gaps
- **217 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+212 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 326 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **5 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `rootlessNoFileMgr`, `config_test.go`, `Manager`, `config.go`, `Command`, `capRunner`, `convert_test.go`, `newCapMgr`, `validate.go`, `captureStdout`, `Get`, `Ops`, `.MateRevertActivity`, `redundancy_test.go`, `scaling.go`, `wireExec`, `hostpath_test.go`, `github.com/spf13/cobra.Command`, `load`, `containerTransport`, `App`, `ctrCfg`, `NewManager`?**
  _High betweenness centrality (0.039) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `runRootWith`, `rootlessNoFileMgr`, `context.Context`, `Manager`, `config.go`, `GenOperator`, `capRunner`, `NewCluster`, `convert_test.go`, `newCapMgr`, `k8s/secrets_test.go`, `prep_test.go`, `switchplan_test.go`, `haCfg`, `Role`, `recRunner`, `captureStdout`, `render.go`, `Ops`, `runner_test.go`, `.MateRevertActivity`, `checkreport.go`, `localCfg`, `namespace_test.go`, `testing.T`, `kubectlTransport`, `load`, `containerTransport`, `App`, `ctrCfg`, `watch_test.go`, `k8s/runtime_test.go`, `NewManager`, `Compose`?**
  _High betweenness centrality (0.036) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `ParseRole`, `broker.go`, `Manager`, `config.go`, `importops.go`, `.ConfigureReplication`, `.releaseToBackup`, `renderDriver`, `validate.go`, `importops_test.go`, `ops_container.go`, `Ops`, `broker/matechannel_test.go`, `.MateRevertActivity`, `BrokerType`, `Set`, `kubectlTransport`, `ReplSite`, `containerTransport`, `scripts.go`, `ParseBlocks`?**
  _High betweenness centrality (0.032) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 34 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 34 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _217 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `bg` be split into smaller, more focused modules?**
  _Cohesion score 0.14901960784313725 - nodes in this community are weakly interconnected._
# Graph Report - solace-cnt-scripts  (2026-09-16)

## Corpus Check
- 187 files · ~550,485 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 21 file(s) not represented in the graph (top: .golden 15, (none) 2, .cli 2)

## Summary
- 3537 nodes · 11122 edges · 138 communities (127 shown, 11 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 2038 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `5fb20901`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- cli_test.go
- Sink
- context.Context
- runRootWith
- config_test.go
- Commands
- Manager
- Config
- GenOperator
- Command
- NewManager
- NewCluster
- eqArgs
- convert_test.go
- Block
- ctrCfg
- ParseBlocks
- k8s/secrets_test.go
- dev.sh
- blocks_test.go
- dev.ps1
- cli/replication_test.go
- switchplan_test.go
- verify_local_test.go
- .releaseLocalActivity
- renderDriver
- ops_container.go
- importops.go
- solace
- validate.go
- CLAUDE.md
- importops_test.go
- domaincerts_resolve_test.go
- Role
- .resolveSecretRefs
- haCfg
- scripts_test.go
- replicationops_test.go
- broker/matechannel_test.go
- cli/platform_test.go
- InjectShutdown
- exit_test.go
- k8sOps
- runner_test.go
- Platform
- NewMateChannel
- io.Writer
- Ops
- Manager
- newRootCmd
- scaling_test.go
- captureStdout
- Operations
- Test catalogue
- Exporting and importing configuration
- DiffBlocks
- ReplRole
- MateChannel
- redundancy_test.go
- checkreport.go
- rootless_test.go
- Cluster
- Configuration
- testing.T
- broker/replication.go
- Command reference
- Developer guide
- blocks.go
- localCfg
- usagef
- internal/k8s
- annotate_test.go
- Set
- namespace_test.go
- newTestOps
- Get
- exportconfig_test.go
- New
- k8s/inspect_test.go
- age
- statusreport_test.go
- config/replication_test.go
- newSEMPTestMate
- k8s/inspect.go
- prep_test.go
- commands.go
- render.go
- judge
- RuleFor
- opPerformReplication
- completion_test.go
- HasPathSeparator
- scaling.go
- Ops
- Abbreviations
- containerTransport
- runExport
- scripts.go
- .preflightOne
- container/secrets_test.go
- watch_test.go
- internal/cli
- operatorversion_test.go
- tlsCfg
- What `import-config` applies
- omitAtExport
- internal/broker
- Cluster
- .releaseToBackup
- internal/config
- completion.go
- logArgs
- ReplSite
- internal/container
- .stateRows
- hostpath_test.go
- load.go
- pullSecretCfg
- broker/transport.go
- container/inspect_test.go
- ResolveEnvPath
- Fixtures and doubles
- storage_test.go
- ParseRole
- broker.go
- .ServerCert
- github.com/spf13/cobra.Command
- warn
- Troubleshooting
- childExit
- cluster.go
- Load
- BrokerType
- Data replication

## God Nodes (most connected - your core abstractions)
1. `newTestOps()` - 129 edges
2. `ctrCfg()` - 120 edges
3. `Role` - 116 edges
4. `newCapMgr()` - 105 edges
5. `Config` - 95 edges
6. `NewCluster()` - 92 edges
7. `loadK8s()` - 85 edges
8. `matchCLI()` - 81 edges
9. `Platform` - 73 edges
10. `bg()` - 68 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Execute()`  [EXTRACTED]
  main.go → internal/cli/root.go
- `assertFencesBalanced()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/annotate.go
- `TestIsMarkerDiscriminatesNamespace()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/annotate.go
- `parseBody()` --calls--> `readMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `parseHeader()` --calls--> `readMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (138 total, 11 thin omitted)

### Community 0 - "bg"
Cohesion: 0.20
Nodes (34): bg(), App, k8sCluster(), k8sContext(), k8sLogin(), opK8sCLI(), opK8sConfigLeader(), opK8sCopyFrom() (+26 more)

### Community 1 - "cli_test.go"
Cohesion: 0.06
Nodes (70): os.File, allowRuntime(), capture(), captureStderr(), collectPaths(), fakeBinaryOnPath(), firstLine(), runCtr() (+62 more)

### Community 2 - "Sink"
Cohesion: 0.18
Nodes (7): columnWidths(), KV, Level, Row, Sink, pad(), TestLevelTagsMapToTheHouseTags()

### Community 3 - "context.Context"
Cohesion: 0.07
Nodes (10): scriptedMate, Exec, context.Context, os/exec.Cmd, Cluster, Cluster, Cluster, Cluster (+2 more)

### Community 4 - "runRootWith"
Cohesion: 0.14
Nodes (35): testing.M, TestAllowCommandRejectedWhereNothingExecutes(), echoRunner(), App, runRootWith(), TestCLICommand(), TestConfiguredRouternameSurvivesTheFallback(), TestCtrConfirmDeclined() (+27 more)

### Community 5 - "config_test.go"
Cohesion: 0.06
Nodes (53): assertContainerBlockDefaults(), assertContainerScaling(), Config, haNodesConfig(), TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes(), TestAdditionalUserNamesCollideOnDocker(), TestAdditionalUserNamesDoNotCollideOnKubernetes() (+45 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.09
Nodes (12): composeNeedsSecretValues(), exactName(), Manager, orNone(), orValue(), platformTitle(), secretSummary(), setOrMissing() (+4 more)

### Community 8 - "Config"
Cohesion: 0.06
Nodes (34): AdditionalUser, Broker, CertDir, Container, ContainerSecurity, DockerConfig, DomainCerts, Image (+26 more)

### Community 9 - "GenOperator"
Cohesion: 0.27
Nodes (13): GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator(), renderOperatorNS(), splitAfterNamespace() (+5 more)

### Community 10 - "Command"
Cohesion: 0.09
Nodes (36): cmdField, commandRules, guardedCmd, Command, checkBinary(), CheckCommand(), checkFlagShape(), checkToken() (+28 more)

### Community 11 - "NewManager"
Cohesion: 0.23
Nodes (14): NewManager(), TestManagerNilSinks(), TestPreflightRunsBeforeAnything(), Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime() (+6 more)

### Community 12 - "NewCluster"
Cohesion: 0.07
Nodes (70): TestStorageRows(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+62 more)

### Community 13 - "eqArgs"
Cohesion: 0.13
Nodes (27): TestCheckStopsProbingWhenUnreachable(), TestReachable(), TestResolveStorageClass(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit() (+19 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (62): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+54 more)

### Community 15 - "Block"
Cohesion: 0.09
Nodes (53): Block, TargetState, regexp.Regexp, importIgnore(), injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs() (+45 more)

### Community 16 - "ctrCfg"
Cohesion: 0.07
Nodes (105): bytes.Buffer, fileExists(), assertMode(), containsStr(), ctrCfg(), Manager, hasCall(), maskedKeys() (+97 more)

### Community 17 - "ParseBlocks"
Cohesion: 0.12
Nodes (18): StripMarkers(), TestStripMarkersKeepsPragmaAndSectionComments(), TestStripMarkersNoMarkersPassthrough(), ParseBlocks(), splitLines(), TestParseBlocksCRLFMatchesLF(), TestParseBlocksIndentedLineWithNoOpenerRefused(), TestParseBlocksMissingEndTerminatorRefused() (+10 more)

### Community 18 - "k8s/secrets_test.go"
Cohesion: 0.12
Nodes (31): GenSecrets(), TestUpdateServerCertSecret(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), checkGolden(), decodeDataValue() (+23 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "blocks_test.go"
Cohesion: 0.22
Nodes (15): newBlock(), blockByIndex(), Capture, loadSample(), TestBrokerTypeFromSchema(), TestNewBlockMessageSpoolQualifierHasNoOwnName(), TestNewBlockOpener(), TestNewBlockVPNNameWithSpaces() (+7 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "cli/replication_test.go"
Cohesion: 0.26
Nodes (17): ReplicationConfigResult, mateChannel(), mateSEMPPassword(), reportReplicationConfig(), App, replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply() (+9 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.19
Nodes (28): SiteState, BuildSwitchPlan(), ExecuteSwitchPlan(), replPrecheckError(), actionStrings(), ordersToA(), replSites(), replVPNs() (+20 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.15
Nodes (41): cliScriptPath(), curlCalls(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestDetectRoleAddrsError() (+33 more)

### Community 25 - ".releaseLocalActivity"
Cohesion: 0.16
Nodes (11): field(), showRedundancyLocalScript(), defaultLocalAddrs(), Ops, hostMatches(), shortHost(), TestBackupActivityStateReadsTheMateColumn(), TestDefaultLocalAddrs() (+3 more)

### Community 26 - "renderDriver"
Cohesion: 0.17
Nodes (20): chunk, chunkResult, runCLISkeleton(), chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment() (+12 more)

### Community 27 - "ops_container.go"
Cohesion: 0.16
Nodes (36): TestCtrManagerConfirmWiring(), containerRenderRole(), containerRole(), containerWhat(), ctrLogin(), ctrManager(), ctrOps(), detectContainerRole() (+28 more)

### Community 28 - "importops.go"
Cohesion: 0.22
Nodes (16): ImportResult, PlannedSection, defaultVPNFirst(), ImportPlan, omittedList(), orNone(), preambleForApply(), quoteList() (+8 more)

### Community 31 - "validate.go"
Cohesion: 0.06
Nodes (35): keyValueEntries, TestRoleNames(), CanonicalDuration(), TestCanonicalDuration(), CheckHostPath(), expandTilde(), expandTildeToken(), Config (+27 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.16
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "importops_test.go"
Cohesion: 0.08
Nodes (37): planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, minimalCapture() (+29 more)

### Community 34 - "domaincerts_resolve_test.go"
Cohesion: 0.08
Nodes (49): DirReader, fakeDirEntry, os.DirEntry, os.FileInfo, os.FileMode, caNameSafe(), checkNoDotDot(), DefaultDirReader() (+41 more)

### Community 35 - "Role"
Cohesion: 0.07
Nodes (24): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+16 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "haCfg"
Cohesion: 0.09
Nodes (27): haCfg(), saCfg(), TestHARoles(), TestPodNameSuffixMatchesTheConfigBound(), TestProductKeyRoles(), TestResourceNames(), TestRestartOrder(), TestOperatorImage() (+19 more)

### Community 38 - "scripts_test.go"
Cohesion: 0.13
Nodes (15): assertLeaderScript(), disableDefaultVPNScript(), domainCertsScript(), parseVPNNames(), TestAssertLeaderScript(), TestDisableDefaultVPNScript(), TestDomainCertsScriptSorted(), TestParseVPNNames() (+7 more)

### Community 39 - "replicationops_test.go"
Cohesion: 0.16
Nodes (29): hasCall(), mateConvergenceShutdowns(), replicationVPNLines(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), siteAEntry() (+21 more)

### Community 40 - "broker/matechannel_test.go"
Cohesion: 0.20
Nodes (18): CLIRunner, fakeRun, Ops, NewCLIMate(), mateReplies(), newTestMate(), TestMateChannelDescribeNamesTheTarget(), TestMateChannelPreflightLearnsTheBrokerType() (+10 more)

### Community 41 - "cli/platform_test.go"
Cohesion: 0.29
Nodes (16): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+8 more)

### Community 42 - "InjectShutdown"
Cohesion: 0.17
Nodes (27): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+19 more)

### Community 43 - "exit_test.go"
Cohesion: 0.19
Nodes (13): runFailRunner, ExitCode(), childStatusError(), k8sEnv(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitStatusIsScopedToInteractiveSessions(), TestExitCodeContract(), TestHelperExitProcess() (+5 more)

### Community 44 - "k8sOps"
Cohesion: 0.26
Nodes (17): confirmAction(), wantRemove(), opCtrConfigDomainCerts(), opCtrConfigServerCerts(), domainCANames(), k8sOps(), k8sWhat(), opK8sConfigDefaultUsers() (+9 more)

### Community 45 - "runner_test.go"
Cohesion: 0.07
Nodes (39): interactiveFailRunner, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveMissingBinaryIsActionable(), TestResolveRefusesCurrentDirectory(), verboseExec() (+31 more)

### Community 46 - "Platform"
Cohesion: 0.17
Nodes (26): checkFlagPlatforms(), commandPlatforms(), declaredList(), flagOnlyOn(), App, parsePlatformList(), platformSuffix(), prepare() (+18 more)

### Community 47 - "NewMateChannel"
Cohesion: 0.23
Nodes (23): execArgs(), NewMateChannel(), ReadSecretKey(), secretCommand(), hasPrefixArgv(), indexOf(), operandOf(), replCfg() (+15 more)

### Community 48 - "io.Writer"
Cohesion: 0.19
Nodes (22): exportconfigReadCounter, layer, io.Reader, io.Writer, TestConfirmDeleteShortcut(), TestPromptYes(), TestPromptYesNo(), TestStdinCanAnswerClosedFile() (+14 more)

### Community 49 - "Ops"
Cohesion: 0.13
Nodes (7): time.Duration, Ops, validName(), rejectionIn(), shellScriptPath(), BaseName(), TestBaseNameSplitsOnBothSeparators()

### Community 50 - "Manager"
Cohesion: 0.22
Nodes (4): Manager, idMapCovers(), origin(), runUserIDs()

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (30): applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias(), TestAllowCommandApprovesAWrappedRuntime() (+22 more)

### Community 52 - "scaling_test.go"
Cohesion: 0.16
Nodes (15): decodeScaling(), Config, setContainerMem(), TestApplyScalingTierDefaultsContainerBlocks(), TestApplyScalingTierDefaultsK8s(), TestApplyScalingTierDefaultsMemOverride(), TestApplyScalingTierDefaultsOffTier(), TestScalingDualSpellingAliasesTheSameField() (+7 more)

### Community 53 - "captureStdout"
Cohesion: 0.16
Nodes (19): opCall, captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch(), opCanI(), opFailOn() (+11 more)

### Community 54 - "Operations"
Cohesion: 0.11
Nodes (18): Bringing up a fresh cluster, Docker and Podman mechanics, Every destructive command confirms, Exit codes, Extra CLI users differ by platform, Operations, Post-deployment configuration order, Re-deploying is safe and explicit (+10 more)

### Community 55 - "Test catalogue"
Cohesion: 0.10
Nodes (20): abbrev_test.go, convert_test.go, Coverage, examples_test.go, internal/abbrev, internal/convert, internal/engine, internal/examples (+12 more)

### Community 56 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 57 - "DiffBlocks"
Cohesion: 0.12
Nodes (35): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, ancestorReported(), blockLabel(), DiffBlocks() (+27 more)

### Community 58 - "ReplRole"
Cohesion: 0.08
Nodes (28): AdminState, backupTarget, Credential, ReplRole, sempExecuteResult, sempMate, sempRedundancyReply, sempReplicationReply (+20 more)

### Community 59 - "MateChannel"
Cohesion: 0.30
Nodes (12): PhaseKind, SwitchAction, SwitchPhase, MateChannel, confirmDemotions(), SwitchPlan, promoteConfirmed(), requirePrimaryActive() (+4 more)

### Community 60 - "redundancy_test.go"
Cohesion: 0.21
Nodes (12): Config, pskCfg(), TestFillStandaloneNodeNameIgnoresAnEmptyHostname(), TestFillStandaloneNodeNameKeepsAConfiguredName(), TestFillStandaloneNodeNameNeverFillsHA(), TestFillStandaloneNodeNameUsesTheHost(), TestNothingGeneratesThePSK(), TestPSKEnvSatisfiesTheContainerRequirement() (+4 more)

### Community 61 - "checkreport.go"
Cohesion: 0.15
Nodes (21): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), allCustomMounted(), containsString(), failRow() (+13 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.09
Nodes (48): capCall, capRunner, New(), Transport, failOnCall(), fakeEnv(), Manager, healthyRootlessOut() (+40 more)

### Community 65 - "Configuration"
Cohesion: 0.15
Nodes (13): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+5 more)

### Community 66 - "testing.T"
Cohesion: 0.06
Nodes (67): testing.T, TestLastLines(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultUsersShowVPNError() (+59 more)

### Community 67 - "broker/replication.go"
Cohesion: 0.06
Nodes (68): colSpan, MateConfig, QueueState, VPNRepl, cliTransport(), containsEndpoint(), dashSpans(), flagByte() (+60 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "blocks.go"
Cohesion: 0.18
Nodes (18): Omission, Region, token, isMarker(), markerRegion(), brokerTypeFromSchema(), firstQuoted(), Capture (+10 more)

### Community 71 - "localCfg"
Cohesion: 0.11
Nodes (33): TestLocalMateRunsThroughOps(), bridgeHostPort(), sempPort(), sempV1OK(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv() (+25 more)

### Community 72 - "usagef"
Cohesion: 0.09
Nodes (31): childExitError, usageError, github.com/spf13/cobra.PositionalArgs, noRolePositional(), registerFlagCompletion(), App, newConvertCmd(), runConvert() (+23 more)

### Community 73 - "internal/k8s"
Cohesion: 0.11
Nodes (19): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go, names_test.go (+11 more)

### Community 74 - "annotate_test.go"
Cohesion: 0.09
Nodes (40): Annotate(), applyMetaField(), Capture, markerSection(), markerVerb(), parseKV(), readMarker(), renderKV() (+32 more)

### Community 75 - "Set"
Cohesion: 0.07
Nodes (38): Entry, shorthand, github.com/spf13/pflag.FlagSet, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies() (+30 more)

### Community 76 - "namespace_test.go"
Cohesion: 0.30
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 77 - "newTestOps"
Cohesion: 0.07
Nodes (60): cliRunNames(), Ops, matchCLI(), newTestOps(), outputForRole(), ranContains(), TestCountContains(), TestDiagnostics() (+52 more)

### Community 78 - "Get"
Cohesion: 0.15
Nodes (25): Example, TestExamplesEmitsToStdout(), parseError(), DetectPlatforms(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint(), TestDetectPlatformsMissingFile(), TestParsePlatform() (+17 more)

### Community 79 - "exportconfig_test.go"
Cohesion: 0.17
Nodes (18): opRunner, exportconfigDriverOutput(), exportconfigTransportOutput(), exportconfigVPNListOutput(), App, newExportconfigRunner(), newExportconfigRunnerRejectingChunk(), runImportConfirmation() (+10 more)

### Community 80 - "New"
Cohesion: 0.15
Nodes (20): TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), TestCheckReportSkipsEmptySections(), New(), NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth() (+12 more)

### Community 81 - "k8s/inspect_test.go"
Cohesion: 0.18
Nodes (17): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+9 more)

### Community 82 - "age"
Cohesion: 0.22
Nodes (9): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+1 more)

### Community 83 - "statusreport_test.go"
Cohesion: 0.20
Nodes (17): Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList(), TestBrokerReportWithNoBrokerDeployed() (+9 more)

### Community 84 - "config/replication_test.go"
Cohesion: 0.36
Nodes (11): Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationAbsent(), TestValidateReplicationEmptyViaAccepted() (+3 more)

### Community 85 - "newSEMPTestMate"
Cohesion: 0.29
Nodes (13): newSEMPTestMate(), sempReply(), sempSite(), TestSEMPMateChecksExecuteResult(), TestSEMPMateEscapesTheVPNName(), TestSEMPMatePlainTextURL(), TestSEMPMatePrimaryActive(), TestSEMPMateReadsBothMateShapes() (+5 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.14
Nodes (21): time.Time, anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), Cluster, normalizeToList(), TestConditionLevelDegradesSafely() (+13 more)

### Community 87 - "prep_test.go"
Cohesion: 0.23
Nodes (14): GenBroker(), joinManifests(), namespaceManifest(), adminCfg(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree(), TestCreateSecretsPreflight(), TestCreateSecretsStopsOnPreflightFailure() (+6 more)

### Community 88 - "commands.go"
Cohesion: 0.21
Nodes (47): opFunc, roleOpFunc, addCommands(), addPodFlag(), App, group(), newBrokerCLICmd(), newBrokerCmd() (+39 more)

### Community 89 - "render.go"
Cohesion: 0.06
Nodes (78): strings.Builder, HealthCheck, PodAffinityTerm, NodeIdentity, boolStr(), BrokerCR(), Compose(), composeEscape() (+70 more)

### Community 90 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.11
Nodes (34): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+26 more)

### Community 92 - "opPerformReplication"
Cohesion: 0.20
Nodes (18): mateChannelFunc, bufio.Reader, App, step(), confirmReplicationConfig(), confirmReplicationSwitch(), App, opConfigureReplication() (+10 more)

### Community 93 - "completion_test.go"
Cohesion: 0.25
Nodes (15): runComplete(), TestAllowCommandOffersNoFiles(), TestCompletionHelpStillWorks(), TestCompletionNeedsAShell(), TestCompletionNoDescriptions(), TestCompletionScriptsGenerate(), TestDirFlagCompletesDirectories(), TestEnvFlagCompletesEnvFiles() (+7 more)

### Community 95 - "scaling.go"
Cohesion: 0.15
Nodes (14): scalingKey, scalingSpelling, scalingTier, containerMem(), Config, Scaling, Scaling, yaml.Node (+6 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 100 - "runExport"
Cohesion: 0.23
Nodes (11): confirmImport(), App, pluralVPN(), runExport(), runImport(), TestPluralVPNWording(), emitOrWrite(), nowStamp() (+3 more)

### Community 101 - "scripts.go"
Cohesion: 0.14
Nodes (26): showCmd, defaultUsersScript(), disableDefaultUsersScript(), enableDefaultUsersScript(), enableDefaultVPNScript(), gatherConfigsScript(), noReleaseActivityScript(), productKeyScript() (+18 more)

### Community 102 - ".preflightOne"
Cohesion: 0.44
Nodes (3): canIAnswer(), Cluster, probe

### Community 103 - "container/secrets_test.go"
Cohesion: 0.21
Nodes (18): ResolveSecretValues(), certFixture(), TestDeletePodmanRemovesTheBundleAfterTheUnit(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeWritingTheComposeFile(), TestDeployDockerPassesTheBundleAsEnvNotArgv(), TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(), TestDeployPodmanFailsWhenTheBundleCannotBeWritten() (+10 more)

### Community 104 - "watch_test.go"
Cohesion: 0.11
Nodes (25): imageFromDeployment(), desiredWatch(), Cluster, splitWatch(), subtractWatch(), deployJSON(), Cluster, operatorItem() (+17 more)

### Community 105 - "internal/cli"
Cohesion: 0.17
Nodes (12): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, examples_test.go, exit_test.go (+4 more)

### Community 106 - "operatorversion_test.go"
Cohesion: 0.16
Nodes (15): compareVersions(), Cluster, imageTag(), operatorVersionWarning(), parseVersion(), operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade() (+7 more)

### Community 107 - "tlsCfg"
Cohesion: 0.43
Nodes (7): Config, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsMakeTheSecretOurs(), TestSuppliedCertsNeedASecretName(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), tlsCfg()

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "omitAtExport"
Cohesion: 0.32
Nodes (7): omitAtExport(), TestImportPlanAcceptsWhatExportProduced(), TestImportPlanRefusesAFileItDidNotExport(), TestOmitAtExportDropsFilteredLinesAndEmptyBlocks(), TestOmitAtExportLeavesTargetDecisionsToImport(), TestOmitAtExportRemovesSkipSectionsAndRecordsThem(), TestOmitExportConfigEndToEnd()

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 113 - "internal/config"
Cohesion: 0.13
Nodes (15): command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go, hostpath_test.go, imagepullsecret_test.go (+7 more)

### Community 114 - "completion.go"
Cohesion: 0.31
Nodes (10): github.com/spf13/cobra.ShellCompDirective, completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), completionShell(), App, isEnvFileName() (+2 more)

### Community 115 - "logArgs"
Cohesion: 0.67
Nodes (4): logArgs(), App, logArgs2(), TestLogArgsBuildsOneSetForBothPlatforms()

### Community 116 - "ReplSite"
Cohesion: 0.33
Nodes (6): Replication, PlannedRoles(), RoleAtSite(), Ops, ReplSite, ReplVPN

### Community 117 - "internal/container"
Cohesion: 0.25
Nodes (8): inspect_test.go, internal/container, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go, transport_test.go

### Community 118 - ".stateRows"
Cohesion: 0.30
Nodes (6): containerState, healthState, Manager, orUnknown(), printable(), TestPrintableStripsControlCharacters()

### Community 119 - "hostpath_test.go"
Cohesion: 0.16
Nodes (18): guardConfig(), Config, setContainerName(), TestContainerCertRequiresKey(), TestContainerHostDirsExpandATilde(), TestContainerNameMustBeAnEngineName(), TestDataDirMustBeAbsolute(), TestExpandHomePathsCoversEveryFieldItClaims() (+10 more)

### Community 120 - "load.go"
Cohesion: 0.24
Nodes (12): TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config, podmanRunUser() (+4 more)

### Community 121 - "pullSecretCfg"
Cohesion: 0.39
Nodes (8): Config, pullSecretCfg(), TestImagePullSecretDerivedSuffixFitsWhereAdditionalUsersAlreadyFits(), TestImagePullSecretNameBringYourOwn(), TestImagePullSecretNameConfiguredNameWins(), TestImagePullSecretNameDerivedWhenCredentialsButNoName(), TestImagePullSecretNameEmptyWithNeither(), TestManagesImagePullSecretNeedsBothCredentials()

### Community 122 - "broker/transport.go"
Cohesion: 0.47
Nodes (5): TestPathHelpers(), certPath(), CLIArg(), cliArg(), CLIScriptPath()

### Community 123 - "container/inspect_test.go"
Cohesion: 0.18
Nodes (17): decodeInspect(), inspectReply(), kvValue(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings(), TestInspectNameLosesDockersLeadingSlash(), TestInspectRejectsAnEmptyReply() (+9 more)

### Community 124 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 125 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 126 - "storage_test.go"
Cohesion: 0.33
Nodes (9): Config, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 127 - "ParseRole"
Cohesion: 0.47
Nodes (5): TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), ParseRole(), RoleAbbrev()

### Community 128 - "broker.go"
Cohesion: 0.50
Nodes (4): countContains(), validCLILine(), ValidVPNName(), validVPNName()

### Community 129 - ".ServerCert"
Cohesion: 0.83
Nodes (3): serverCertFile(), serverCertScript(), TestServerCertScript()

### Community 130 - "github.com/spf13/cobra.Command"
Cohesion: 0.28
Nodes (17): github.com/spf13/cobra.Command, addLogFlags(), newOperatorCmd(), newOperatorDeployCmd(), newOperatorGenerateCmd(), newOperatorLogsCmd(), newOperatorRemoveCmd(), newOperatorRestartCmd() (+9 more)

### Community 131 - "warn"
Cohesion: 0.83
Nodes (3): lineSink(), progress(), warn()

### Community 132 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, File descriptors on rootless podman, Import reported failure, Podman secret flags, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 133 - "childExit"
Cohesion: 0.67
Nodes (3): childExit(), TestChildExitKeepsItsMessage(), TestExitCodeIsNeverNegative()

### Community 138 - "Load"
Cohesion: 0.15
Nodes (25): minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors(), TestLoadSuccess() (+17 more)

### Community 139 - "BrokerType"
Cohesion: 0.33
Nodes (5): BrokerType, cliMate, bannerType(), checkSameType(), showReplicationScript()

### Community 143 - "Data replication"
Cohesion: 0.40
Nodes (5): Configuring a site, Data replication, Reaching the mate, Switching roles, When a run fails partway

## Knowledge Gaps
- **221 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+216 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 329 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **11 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `config_test.go`, `Manager`, `Config`, `Command`, `Load`, `NewManager`, `convert_test.go`, `ctrCfg`, `validate.go`, `Ops`, `scaling_test.go`, `captureStdout`, `redundancy_test.go`, `rootless_test.go`, `localCfg`, `Get`, `commands.go`, `render.go`, `opPerformReplication`, `scaling.go`, `containerTransport`, `container/secrets_test.go`, `hostpath_test.go`, `load.go`?**
  _High betweenness centrality (0.044) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `cli_test.go`, `cluster.go`, `Manager`, `GenOperator`, `NewManager`, `NewCluster`, `convert_test.go`, `ctrCfg`, `k8s/secrets_test.go`, `Role`, `haCfg`, `NewMateChannel`, `Ops`, `captureStdout`, `checkreport.go`, `rootless_test.go`, `Cluster`, `localCfg`, `namespace_test.go`, `newTestOps`, `prep_test.go`, `render.go`, `opPerformReplication`, `containerTransport`, `container/secrets_test.go`, `watch_test.go`, `ReplSite`?**
  _High betweenness centrality (0.039) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `Manager`, `Config`, `BrokerType`, `ParseBlocks`, `.releaseLocalActivity`, `renderDriver`, `ops_container.go`, `importops.go`, `validate.go`, `haCfg`, `scripts_test.go`, `broker/matechannel_test.go`, `k8sOps`, `NewMateChannel`, `Ops`, `ReplRole`, `broker/replication.go`, `Set`, `newTestOps`, `opPerformReplication`, `Ops`, `containerTransport`, `runExport`, `.releaseToBackup`, `ParseRole`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 39 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 39 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _221 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `cli_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.06298904538341157 - nodes in this community are weakly interconnected._
# Graph Report - solace-k8-scripts  (2026-09-16)

## Corpus Check
- 188 files · ~550,836 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 21 file(s) not represented in the graph (top: .golden 15, (none) 2, .cli 2)

## Summary
- 3539 nodes · 11123 edges · 128 communities (120 shown, 4 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 2038 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `3e4d2c81`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- cli_test.go
- Sink
- context.Context
- strings.Builder
- config_test.go
- Commands
- Manager
- config.go
- Config
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
- ParseBlocks
- dev.ps1
- cli/replication_test.go
- switchplan_test.go
- verify_local_test.go
- .releaseToBackup
- renderDriver
- ops_container.go
- haCfg
- solace
- validate.go
- CLAUDE.md
- importops_test.go
- domaincerts_resolve_test.go
- Role
- .resolveSecretRefs
- recRunner
- render.go
- replicationops_test.go
- broker/matechannel_test.go
- cli/platform_test.go
- InjectShutdown
- runRootWith
- Compose
- Echo
- Platform
- ReplSite
- confirmAction
- Ops
- runRoot
- newRootCmd
- k8s/runtime_test.go
- captureStdout
- Operations
- Test catalogue
- Exporting and importing configuration
- DiffBlocks
- .rpc
- platformOps
- runner_test.go
- checkreport.go
- rootless_test.go
- newEchoMgr
- Configuration
- command_test.go
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
- testing.T
- Get
- exportconfig_test.go
- output_test.go
- k8s/inspect_test.go
- age
- statusreport_test.go
- config/replication_test.go
- newSEMPTestMate
- k8s/inspect.go
- prep_test.go
- github.com/spf13/cobra.Command
- load
- judge
- RuleFor
- opPerformReplication
- completion_test.go
- hostpath.go
- scaling.go
- Abbreviations
- containerTransport
- runExport
- scripts.go
- .preflightOne
- container/secrets_test.go
- watch_test.go
- internal/cli
- operatorversion.go
- tlsCfg
- What `import-config` applies
- omitAtExport
- internal/broker
- Cluster
- execguard_test.go
- internal/config
- completion.go
- logArgs
- internal/container
- .stateRows
- hostpath_test.go
- load.go
- container/inspect_test.go
- Fixtures and doubles
- storage_test.go
- emit
- Troubleshooting
- Load
- .ConfigureReplication
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
- `main()` --calls--> `ExitCode()`  [EXTRACTED]
  main.go → internal/cli/exit.go
- `parseBody()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `parseBody()` --calls--> `readMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `applyMetaField()` --calls--> `BrokerType`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go
- `applyMetaField()` --calls--> `brokerTypeFromSchema()`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (128 total, 4 thin omitted)

### Community 0 - "bg"
Cohesion: 0.16
Nodes (48): bg(), domainCANames(), App, k8sCluster(), k8sContext(), k8sLogin(), k8sOps(), k8sWhat() (+40 more)

### Community 1 - "cli_test.go"
Cohesion: 0.09
Nodes (62): testing.M, allowRuntime(), captureStderr(), echoRunner(), fakeBinaryOnPath(), runCtr(), runStandalone(), runStatusStderr() (+54 more)

### Community 2 - "Sink"
Cohesion: 0.12
Nodes (14): solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), TestCheckReportSkipsEmptySections(), columnWidths(), KV, Level, Row, Sink (+6 more)

### Community 3 - "context.Context"
Cohesion: 0.06
Nodes (12): context.Context, Manager, idMapCovers(), origin(), runUserIDs(), Cluster, Cluster, Cluster (+4 more)

### Community 4 - "strings.Builder"
Cohesion: 0.17
Nodes (25): shorthand, github.com/spf13/pflag.FlagSet, strings.Builder, mdRow(), renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands() (+17 more)

### Community 5 - "config_test.go"
Cohesion: 0.04
Nodes (76): assertContainerBlockDefaults(), assertContainerScaling(), envTree(), Config, haNodesConfig(), TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes(), TestAdditionalUserNamesCollideOnDocker() (+68 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.09
Nodes (10): composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing(), splitLimit() (+2 more)

### Community 8 - "config.go"
Cohesion: 0.08
Nodes (27): AdditionalUser, Broker, CertDir, Container, ContainerSecurity, DockerConfig, DomainCerts, Image (+19 more)

### Community 9 - "Config"
Cohesion: 0.15
Nodes (16): Config, Scaling, OperatorNamespace(), GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes() (+8 more)

### Community 10 - "Command"
Cohesion: 0.14
Nodes (23): cmdField, commandRules, guardedCmd, Command, checkBinary(), CheckCommand(), checkFlagShape(), checkToken() (+15 more)

### Community 11 - "capRunner"
Cohesion: 0.11
Nodes (27): capCall, capRunner, New(), Transport, NewManager(), TestManagerLogsCLIShell(), TestManagerNilSinks(), Manager (+19 more)

### Community 12 - "NewCluster"
Cohesion: 0.08
Nodes (69): TestStorageRows(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+61 more)

### Community 13 - "eqArgs"
Cohesion: 0.17
Nodes (23): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), TestCLIAndShellAreInteractive() (+15 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (63): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+55 more)

### Community 15 - "Block"
Cohesion: 0.09
Nodes (53): Block, TargetState, regexp.Regexp, importIgnore(), injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs() (+45 more)

### Community 16 - "ctrCfg"
Cohesion: 0.09
Nodes (84): fileExists(), assertMode(), containsStr(), ctrCfg(), hasCall(), maskedKeys(), newCapMgr(), TestContainerRunningMatchesNameExactly() (+76 more)

### Community 17 - "importops.go"
Cohesion: 0.10
Nodes (32): BrokerType, ImportResult, PlannedSection, StripMarkers(), TestStripMarkersKeepsPragmaAndSectionComments(), TestStripMarkersNoMarkersPassthrough(), splitLines(), bannerType() (+24 more)

### Community 18 - "k8s/secrets_test.go"
Cohesion: 0.11
Nodes (34): GenBroker(), GenSecrets(), joinManifests(), namespaceManifest(), TestUpdateServerCertSecret(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret() (+26 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "ParseBlocks"
Cohesion: 0.19
Nodes (19): Annotate(), TestAnnotateIsDeterministic(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), ParseBlocks(), blockByIndex(), Capture, loadSample() (+11 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "cli/replication_test.go"
Cohesion: 0.34
Nodes (14): mateChannel(), mateSEMPPassword(), App, replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer(), TestMateChannelPicksTheMechanismTheSiteDeclares() (+6 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.13
Nodes (40): PhaseKind, SiteState, SwitchAction, SwitchPhase, MateChannel, BuildSwitchPlan(), confirmDemotions(), ExecuteSwitchPlan() (+32 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.13
Nodes (44): CLIArg(), cliArg(), CLIScriptPath(), cliScriptPath(), curlCalls(), Ops, isCurl(), newLocalOps() (+36 more)

### Community 25 - ".releaseToBackup"
Cohesion: 0.15
Nodes (12): field(), defaultLocalAddrs(), Ops, hostMatches(), shortHost(), TestBackupActivityStateReadsTheMateColumn(), TestDefaultLocalAddrs(), activity() (+4 more)

### Community 26 - "renderDriver"
Cohesion: 0.16
Nodes (20): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), shQuote() (+12 more)

### Community 27 - "ops_container.go"
Cohesion: 0.14
Nodes (42): TestCtrManagerConfirmWiring(), childExit(), TestChildExitKeepsItsMessage(), TestExitCodeIsNeverNegative(), wantRemove(), containerRenderRole(), containerRole(), containerWhat() (+34 more)

### Community 28 - "haCfg"
Cohesion: 0.14
Nodes (20): TestCheckStopsProbingWhenUnreachable(), TestResolveStorageClass(), haCfg(), saCfg(), TestHARoles(), TestPodNameSuffixMatchesTheConfigBound(), TestProductKeyRoles(), TestResourceNames() (+12 more)

### Community 31 - "validate.go"
Cohesion: 0.08
Nodes (27): keyValueEntries, ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, TestRoleNames(), viaDescription(), RoleNames() (+19 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "importops_test.go"
Cohesion: 0.10
Nodes (32): planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, minimalCapture() (+24 more)

### Community 34 - "domaincerts_resolve_test.go"
Cohesion: 0.07
Nodes (50): DirReader, fakeDirEntry, os.DirEntry, os.FileInfo, os.FileMode, caNameSafe(), checkNoDotDot(), DefaultDirReader() (+42 more)

### Community 35 - "Role"
Cohesion: 0.07
Nodes (24): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+16 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.14
Nodes (10): TestCanIAnswerReadsTheLastLine(), NewTransport(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest() (+2 more)

### Community 38 - "render.go"
Cohesion: 0.12
Nodes (28): Placement, PodAffinityTerm, boolStr(), containerSecretSpecs(), cut(), escapePercent(), groupKey(), itoa() (+20 more)

### Community 39 - "replicationops_test.go"
Cohesion: 0.23
Nodes (19): hasCall(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), TestConfigureReplicationClosingShowFailureStillReportsTheApply(), TestConfigureReplicationPhase1BodyOrderIsShutdownsThenRemovalsThenMate(), TestConfigureReplicationPhase1RejectionStopsBeforePhase2() (+11 more)

### Community 40 - "broker/matechannel_test.go"
Cohesion: 0.20
Nodes (18): CLIRunner, fakeRun, Ops, NewCLIMate(), mateReplies(), newTestMate(), TestMateChannelDescribeNamesTheTarget(), TestMateChannelPreflightLearnsTheBrokerType() (+10 more)

### Community 41 - "cli/platform_test.go"
Cohesion: 0.29
Nodes (16): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+8 more)

### Community 42 - "InjectShutdown"
Cohesion: 0.17
Nodes (27): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+19 more)

### Community 43 - "runRootWith"
Cohesion: 0.09
Nodes (40): runFailRunner, TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb() (+32 more)

### Community 44 - "Compose"
Cohesion: 0.14
Nodes (21): HealthCheck, NodeIdentity, Compose(), composeEscape(), EnvPairs(), ContainerSecret, healthCmd(), Quadlet() (+13 more)

### Community 45 - "Echo"
Cohesion: 0.09
Nodes (20): interactiveFailRunner, Exec, os/exec.Cmd, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveMissingBinaryIsActionable() (+12 more)

### Community 46 - "Platform"
Cohesion: 0.19
Nodes (24): checkFlagPlatforms(), commandPlatforms(), declaredList(), flagOnlyOn(), App, parsePlatformList(), platformSuffix(), prepare() (+16 more)

### Community 47 - "ReplSite"
Cohesion: 0.22
Nodes (24): ReplSite, execArgs(), NewMateChannel(), ReadSecretKey(), secretCommand(), hasPrefixArgv(), indexOf(), operandOf() (+16 more)

### Community 48 - "confirmAction"
Cohesion: 0.19
Nodes (23): exportconfigReadCounter, layer, io.Reader, io.Writer, TestConfirmDeleteShortcut(), TestPromptYes(), TestPromptYesNo(), TestStdinCanAnswerClosedFile() (+15 more)

### Community 49 - "Ops"
Cohesion: 0.15
Nodes (9): time.Duration, cliRunNames(), countContains(), Ops, runCLISkeleton(), validName(), ValidVPNName(), validVPNName() (+1 more)

### Community 50 - "runRoot"
Cohesion: 0.10
Nodes (27): firstLine(), runRoot(), TestBashEnvGivenToEnvFlag(), TestConfigureServerCertsRefusesASecretItDoesNotOwn(), TestConvertErrorPaths(), TestConvertParseError(), TestConvertRoundTrip(), TestConvertToFile() (+19 more)

### Community 51 - "newRootCmd"
Cohesion: 0.10
Nodes (24): applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias(), TestAllowCommandIsRegisteredWhereItExecutes() (+16 more)

### Community 52 - "k8s/runtime_test.go"
Cohesion: 0.43
Nodes (7): TestClusterHonoursRuntime(), TestExecutorRefusesUnapprovedRuntime(), TestRuntimeDefaultArgvUnchanged(), TestTransportHonoursRuntime(), unapprovedCfg(), withLeading(), wrappedCfg()

### Community 53 - "captureStdout"
Cohesion: 0.14
Nodes (21): opCall, os.File, capture(), captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch() (+13 more)

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

### Community 58 - ".rpc"
Cohesion: 0.09
Nodes (27): AdminState, backupTarget, Credential, sempExecuteResult, sempMate, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply (+19 more)

### Community 59 - "platformOps"
Cohesion: 0.27
Nodes (22): opFunc, roleOpFunc, addPodFlag(), group(), newBrokerCmd(), newBrokerCopyCmd(), newBrokerDeployCmd(), newBrokerExecCmd() (+14 more)

### Community 60 - "runner_test.go"
Cohesion: 0.15
Nodes (21): captureStdout(), helperCommand(), TestEchoDefaultWriter(), TestEchoOutput(), TestEchoOutputInput(), TestEchoRun(), TestEchoRunEnv(), TestEchoRunEnvNoEnv() (+13 more)

### Community 61 - "checkreport.go"
Cohesion: 0.14
Nodes (24): orNone(), orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), allCustomMounted(), containsString() (+16 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.16
Nodes (37): failOnCall(), fakeEnv(), Manager, healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirRefusesAnUnwritableParent(), TestCheckIDMappingRefusesAnUnmappedRunUser() (+29 more)

### Community 63 - "newEchoMgr"
Cohesion: 0.12
Nodes (19): bytes.Buffer, Manager, newEchoMgr(), rootlessNoFileMgr(), setNoFile(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployPodmanDryRunHidesSecretBytes() (+11 more)

### Community 65 - "Configuration"
Cohesion: 0.15
Nodes (13): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+5 more)

### Community 66 - "command_test.go"
Cohesion: 0.15
Nodes (14): decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRenamedRuntimeKeysFailLoud() (+6 more)

### Community 67 - "broker/replication.go"
Cohesion: 0.08
Nodes (58): colSpan, MateConfig, QueueState, cliTransport(), containsEndpoint(), dashSpans(), flagByte(), gutterClear() (+50 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "blocks.go"
Cohesion: 0.14
Nodes (21): Omission, Region, token, brokerTypeFromSchema(), firstQuoted(), Capture, isIndented(), isPragma() (+13 more)

### Community 71 - "localCfg"
Cohesion: 0.11
Nodes (33): TestLocalMateRunsThroughOps(), bridgeHostPort(), sempPort(), sempV1OK(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv() (+25 more)

### Community 72 - "usagef"
Cohesion: 0.10
Nodes (28): childExitError, usageError, github.com/spf13/cobra.PositionalArgs, newPerformExportConfigCmd(), registerFlagCompletion(), App, newConvertCmd(), runConvert() (+20 more)

### Community 73 - "internal/k8s"
Cohesion: 0.11
Nodes (19): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go, names_test.go (+11 more)

### Community 74 - "annotate_test.go"
Cohesion: 0.10
Nodes (38): applyMetaField(), Capture, isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV(), readMarker() (+30 more)

### Community 75 - "Set"
Cohesion: 0.14
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 76 - "namespace_test.go"
Cohesion: 0.30
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 77 - "testing.T"
Cohesion: 0.05
Nodes (121): testing.T, Ops, matchCLI(), newTestOps(), outputForRole(), ranContains(), TestCountContains(), TestDiagnostics() (+113 more)

### Community 78 - "Get"
Cohesion: 0.16
Nodes (24): Example, TestExamplesEmitsToStdout(), DetectPlatforms(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint(), TestDetectPlatformsMissingFile(), TestParsePlatform(), TestPlatforms() (+16 more)

### Community 79 - "exportconfig_test.go"
Cohesion: 0.19
Nodes (16): opRunner, exportconfigDriverOutput(), exportconfigTransportOutput(), exportconfigVPNListOutput(), App, newExportconfigRunner(), newExportconfigRunnerRejectingChunk(), runImportConfirmation() (+8 more)

### Community 80 - "output_test.go"
Cohesion: 0.24
Nodes (14): sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLevelOrderKeepsSkipBelowFail(), TestLineAddsNoPrefix(), TestSectionPadsToWidth() (+6 more)

### Community 81 - "k8s/inspect_test.go"
Cohesion: 0.16
Nodes (18): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestConditionLevelDegradesSafely() (+10 more)

### Community 82 - "age"
Cohesion: 0.23
Nodes (8): age(), roleRank(), Cluster, operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers(), TestOwnsPodFallsBackToTheNameInfix()

### Community 83 - "statusreport_test.go"
Cohesion: 0.20
Nodes (17): Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList(), TestBrokerReportWithNoBrokerDeployed() (+9 more)

### Community 84 - "config/replication_test.go"
Cohesion: 0.27
Nodes (12): Replication, Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationAbsent() (+4 more)

### Community 85 - "newSEMPTestMate"
Cohesion: 0.31
Nodes (12): newSEMPTestMate(), sempReply(), sempSite(), TestSEMPMateChecksExecuteResult(), TestSEMPMateEscapesTheVPNName(), TestSEMPMatePlainTextURL(), TestSEMPMatePrimaryActive(), TestSEMPMateReadsRolesTriState() (+4 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.16
Nodes (18): time.Time, anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes() (+10 more)

### Community 87 - "prep_test.go"
Cohesion: 0.25
Nodes (13): adminCfg(), TestCreateNamespaceApplyFails(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree(), TestCreateSecretsPreflight(), TestCreateSecretsStopsOnPreflightFailure(), TestDeleteNamespaceStopsOnPreflightFailure(), TestDeleteSecrets() (+5 more)

### Community 88 - "github.com/spf13/cobra.Command"
Cohesion: 0.22
Nodes (41): github.com/spf13/cobra.Command, addCommands(), addLogFlags(), App, newBrokerCLICmd(), newBrokerConfigureCmd(), newBrokerRemoveCmd(), newBrokerShellCmd() (+33 more)

### Community 89 - "load"
Cohesion: 0.15
Nodes (28): BrokerCR(), ComposeProject(), ContainerSecrets(), SecretPreflight(), load(), TestArtifactsCarryNoSecrets(), TestComposeProjectFoldsToComposesGrammar(), TestComposeProjectIsDeclaredNotDerived() (+20 more)

### Community 90 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.11
Nodes (34): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+26 more)

### Community 92 - "opPerformReplication"
Cohesion: 0.19
Nodes (20): mateChannelFunc, bufio.Reader, App, lineSink(), progress(), step(), warn(), confirmReplicationConfig() (+12 more)

### Community 93 - "completion_test.go"
Cohesion: 0.35
Nodes (12): runComplete(), TestAllowCommandOffersNoFiles(), TestCompletionNeedsAShell(), TestDirFlagCompletesDirectories(), TestEnvFlagCompletesEnvFiles(), TestEnvFlagPrefixFilters(), TestEnvFlagWithPathDefersToShell(), TestNoArgsLeafOffersNoFiles() (+4 more)

### Community 94 - "hostpath.go"
Cohesion: 0.14
Nodes (12): BaseName(), expandTilde(), expandTildeToken(), Config, HasPathSeparator(), IsAbsHostPath(), isPathSep(), TestBaseNameSplitsOnBothSeparators() (+4 more)

### Community 95 - "scaling.go"
Cohesion: 0.16
Nodes (14): scalingKey, scalingSpelling, scalingTier, containerMem(), Config, Scaling, Scaling, yaml.Node (+6 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 100 - "runExport"
Cohesion: 0.25
Nodes (10): confirmImport(), App, pluralVPN(), runExport(), runImport(), TestPluralVPNWording(), emitOrWrite(), nowStamp() (+2 more)

### Community 101 - "scripts.go"
Cohesion: 0.06
Nodes (47): showCmd, validCLILine(), Ops, assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript() (+39 more)

### Community 102 - ".preflightOne"
Cohesion: 0.44
Nodes (3): canIAnswer(), Cluster, probe

### Community 103 - "container/secrets_test.go"
Cohesion: 0.21
Nodes (18): ResolveSecretValues(), certFixture(), TestDeletePodmanRemovesTheBundleAfterTheUnit(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeWritingTheComposeFile(), TestDeployDockerPassesTheBundleAsEnvNotArgv(), TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(), TestDeployPodmanFailsWhenTheBundleCannotBeWritten() (+10 more)

### Community 104 - "watch_test.go"
Cohesion: 0.12
Nodes (24): desiredWatch(), Cluster, splitWatch(), subtractWatch(), deployJSON(), Cluster, operatorItem(), TestAmbiguousOperatorIsNotFoldedIntoNotInstalled() (+16 more)

### Community 105 - "internal/cli"
Cohesion: 0.17
Nodes (12): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, examples_test.go, exit_test.go (+4 more)

### Community 106 - "operatorversion.go"
Cohesion: 0.25
Nodes (8): compareVersions(), Cluster, imageTag(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestImageTag(), TestOperatorVersionWarningFiresOnlyOnADowngrade()

### Community 107 - "tlsCfg"
Cohesion: 0.43
Nodes (7): Config, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsMakeTheSecretOurs(), TestSuppliedCertsNeedASecretName(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), tlsCfg()

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "omitAtExport"
Cohesion: 0.47
Nodes (5): omitAtExport(), TestImportPlanRefusesAFileItDidNotExport(), TestOmitAtExportDropsFilteredLinesAndEmptyBlocks(), TestOmitAtExportLeavesTargetDecisionsToImport(), TestOmitAtExportRemovesSkipSectionsAndRecordsThem()

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 112 - "execguard_test.go"
Cohesion: 0.19
Nodes (12): Config, guardCommandOf(), setGuardCommand(), TestAllowCommandsAccepts(), TestAllowCommandsRejects(), TestAllowCommandsRejectsEscalation(), TestAllowedBinaryIsNotGloballyAllowed(), TestComposeCommandDerivation() (+4 more)

### Community 113 - "internal/config"
Cohesion: 0.13
Nodes (15): command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go, hostpath_test.go, imagepullsecret_test.go (+7 more)

### Community 114 - "completion.go"
Cohesion: 0.31
Nodes (10): github.com/spf13/cobra.ShellCompDirective, completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), completionShell(), App, isEnvFileName() (+2 more)

### Community 115 - "logArgs"
Cohesion: 0.67
Nodes (4): logArgs(), App, logArgs2(), TestLogArgsBuildsOneSetForBothPlatforms()

### Community 117 - "internal/container"
Cohesion: 0.25
Nodes (8): inspect_test.go, internal/container, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go, transport_test.go

### Community 118 - ".stateRows"
Cohesion: 0.18
Nodes (13): containerState, healthState, decodeInspect(), Manager, orUnknown(), printable(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown() (+5 more)

### Community 119 - "hostpath_test.go"
Cohesion: 0.11
Nodes (23): CanonicalDuration(), TestCanonicalDuration(), guardConfig(), CheckHostPath(), Config, setContainerName(), TestCheckHostPathAccepts(), TestCheckHostPathRejects() (+15 more)

### Community 120 - "load.go"
Cohesion: 0.18
Nodes (13): TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config, parseError() (+5 more)

### Community 123 - "container/inspect_test.go"
Cohesion: 0.26
Nodes (11): inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer() (+3 more)

### Community 125 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 126 - "storage_test.go"
Cohesion: 0.33
Nodes (9): Config, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 132 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, File descriptors on rootless podman, Import reported failure, Podman secret flags, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 138 - "Load"
Cohesion: 0.07
Nodes (48): minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadReadError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors() (+40 more)

### Community 139 - ".ConfigureReplication"
Cohesion: 0.07
Nodes (33): cliMate, ReplRole, scriptedMate, VPNRepl, Ops, ReplicationConfigResult, isRunCLIRejection(), mateConvergenceShutdowns() (+25 more)

### Community 143 - "Data replication"
Cohesion: 0.40
Nodes (5): Configuring a site, Data replication, Reaching the mate, Switching roles, When a run fails partway

## Knowledge Gaps
- **221 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+216 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 331 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Config` connect `Config` to `cli_test.go`, `context.Context`, `strings.Builder`, `Manager`, `config.go`, `capRunner`, `NewCluster`, `convert_test.go`, `ctrCfg`, `k8s/secrets_test.go`, `haCfg`, `Role`, `recRunner`, `render.go`, `Compose`, `ReplSite`, `Ops`, `k8s/runtime_test.go`, `captureStdout`, `checkreport.go`, `newEchoMgr`, `localCfg`, `namespace_test.go`, `testing.T`, `config/replication_test.go`, `prep_test.go`, `load`, `opPerformReplication`, `containerTransport`, `container/secrets_test.go`, `watch_test.go`?**
  _High betweenness centrality (0.046) - this node is a cross-community bridge._
- **Why does `Platform` connect `Platform` to `config_test.go`, `Manager`, `config.go`, `Command`, `Load`, `capRunner`, `convert_test.go`, `ctrCfg`, `validate.go`, `Ops`, `captureStdout`, `platformOps`, `newEchoMgr`, `localCfg`, `Get`, `github.com/spf13/cobra.Command`, `load`, `opPerformReplication`, `hostpath.go`, `scaling.go`, `containerTransport`, `container/secrets_test.go`, `execguard_test.go`, `hostpath_test.go`, `load.go`?**
  _High betweenness centrality (0.037) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `config_test.go`, `Manager`, `config.go`, `.ConfigureReplication`, `importops.go`, `.releaseToBackup`, `renderDriver`, `ops_container.go`, `validate.go`, `recRunner`, `broker/matechannel_test.go`, `ReplSite`, `Ops`, `.rpc`, `Set`, `testing.T`, `opPerformReplication`, `hostpath.go`, `containerTransport`, `runExport`, `scripts.go`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 39 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 39 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _221 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `cli_test.go` be split into smaller, more focused modules?**
  _Cohesion score 0.08878968253968254 - nodes in this community are weakly interconnected._
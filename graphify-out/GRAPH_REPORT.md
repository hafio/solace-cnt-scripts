# Graph Report - solace-k8-scripts  (2026-09-15)

## Corpus Check
- 186 files · ~537,037 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 3450 nodes · 10828 edges · 134 communities (126 shown, 4 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 2001 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `fb4301ca`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- matchCLI
- Sink
- context.Context
- commanddoc_test.go
- testing.T
- Commands
- Manager
- config.go
- GenOperator
- CheckCommand
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
- VPNRepl
- switchplan_test.go
- verify_local_test.go
- .releaseToBackup
- renderDriver
- exportconfig_test.go
- Config
- solace
- validate.go
- CLAUDE.md
- importops_test.go
- domaincerts_resolve_test.go
- Role
- .resolveSecretRefs
- recRunner
- transform_test.go
- replicationops_test.go
- statusreport_test.go
- cli/platform_test.go
- inject.go
- usagef
- NewMateChannel
- render.go
- Platform
- broker/replication.go
- Get
- Ops
- runner_test.go
- newRootCmd
- opPerformReplication
- broker/matechannel_test.go
- Operations
- Test catalogue
- Exporting and importing configuration
- diff.go
- .MateRevertActivity
- Cluster
- .rpc
- checkreport.go
- Cluster
- container/preflight_test.go
- Configuration
- .ConfigureReplication
- broker/replication_test.go
- Command reference
- Developer guide
- blocks.go
- localCfg
- New
- internal/k8s
- annotate_test.go
- Set
- namespace_test.go
- newTestOps
- MateChannel
- DiffBlocks
- scaling.go
- ParseBlocks
- age
- k8s/inspect_test.go
- github.com/spf13/cobra.Command
- hostpath_test.go
- k8s/inspect.go
- .preflightOne
- commands.go
- load
- judge
- RuleFor
- exit_test.go
- cli_test.go
- containerState
- ReplSite
- Abbreviations
- config/replication_test.go
- tlsCfg
- scripts.go
- InjectShutdown
- container/secrets_test.go
- watch_test.go
- internal/cli
- operatorversion.go
- container/inspect_test.go
- What `import-config` applies
- execguard_test.go
- internal/broker
- pullSecretCfg
- Troubleshooting
- internal/config
- Command
- completion.go
- abbrevdoc_test.go
- internal/container
- allowcommand_test.go
- load.go
- StripMarkers
- NewManager
- newBlock
- ContainerSecret
- Compose
- Fixtures and doubles
- storage_test.go
- cli/examples_test.go
- ParseRole
- EnvPairs
- newEchoMgr
- importIgnore
- Image
- broker/transport.go

## God Nodes (most connected - your core abstractions)
1. `newTestOps()` - 128 edges
2. `Role` - 116 edges
3. `ctrCfg()` - 115 edges
4. `newCapMgr()` - 100 edges
5. `Config` - 95 edges
6. `NewCluster()` - 92 edges
7. `loadK8s()` - 85 edges
8. `matchCLI()` - 80 edges
9. `Platform` - 72 edges
10. `bg()` - 68 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `ExitCode()`  [EXTRACTED]
  main.go → internal/cli/exit.go
- `assertFencesBalanced()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/annotate.go
- `assertFencesBalanced()` --calls--> `markerVerb()`  [INFERRED]
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

## Communities (134 total, 4 thin omitted)

### Community 0 - "bg"
Cohesion: 0.05
Nodes (125): exportconfigReadCounter, layer, bufio.Reader, io.Reader, io.Writer, TestStdinCanAnswerClosedFile(), logArgs(), confirmAction() (+117 more)

### Community 1 - "matchCLI"
Cohesion: 0.06
Nodes (53): matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultVPNDisableError() (+45 more)

### Community 2 - "Sink"
Cohesion: 0.21
Nodes (3): KV, Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.07
Nodes (9): scriptedMate, context.Context, Manager, Cluster, Cluster, Cluster, Cluster, badJSONRunner (+1 more)

### Community 4 - "commanddoc_test.go"
Cohesion: 0.29
Nodes (13): github.com/spf13/pflag.FlagSet, anchor(), argumentLine(), availableSubs(), firstDiff(), mdCell(), renderCommandDocs(), TestCommandDocs() (+5 more)

### Community 5 - "testing.T"
Cohesion: 0.04
Nodes (123): testing.T, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds() (+115 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.09
Nodes (11): composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing(), solaceRows() (+3 more)

### Community 8 - "config.go"
Cohesion: 0.09
Nodes (24): AdditionalUser, Broker, CertDir, ContainerSecurity, DomainCerts, Node, Operator, PodSecurity (+16 more)

### Community 9 - "GenOperator"
Cohesion: 0.27
Nodes (13): GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator(), renderOperatorNS(), splitAfterNamespace() (+5 more)

### Community 10 - "CheckCommand"
Cohesion: 0.19
Nodes (18): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+10 more)

### Community 11 - "capRunner"
Cohesion: 0.17
Nodes (16): capCall, capRunner, New(), TestNewDefaults(), Transport, TestManagerLogsCLIShell(), NewTransport(), dockerCfg() (+8 more)

### Community 12 - "NewCluster"
Cohesion: 0.07
Nodes (77): TestCheckStopsProbingWhenUnreachable(), TestResolveStorageClass(), TestStorageRows(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts() (+69 more)

### Community 13 - "eqArgs"
Cohesion: 0.17
Nodes (23): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), TestCLIAndShellAreInteractive() (+15 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (63): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+55 more)

### Community 15 - "Block"
Cohesion: 0.16
Nodes (28): Block, TargetState, regexp.Regexp, ClearExistingClientCAs(), ClearExistingNested(), clearNestedObjects(), ClearTargetVirtualHostnames(), CollapseDuplicateLines() (+20 more)

### Community 16 - "ctrCfg"
Cohesion: 0.09
Nodes (80): fileExists(), assertMode(), containsStr(), ctrCfg(), hasCall(), maskedKeys(), newCapMgr(), TestContainerRunningMatchesNameExactly() (+72 more)

### Community 17 - "importops.go"
Cohesion: 0.10
Nodes (29): BrokerType, ImportResult, PlannedSection, isPragma(), bannerType(), checkProvenance(), checkSameType(), defaultVPNFirst() (+21 more)

### Community 18 - "k8s/secrets_test.go"
Cohesion: 0.12
Nodes (31): GenSecrets(), TestUpdateServerCertSecret(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), checkGolden(), decodeDataValue() (+23 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "prep_test.go"
Cohesion: 0.17
Nodes (18): GenBroker(), joinManifests(), namespaceManifest(), adminCfg(), TestCreateNamespaceApplyFails(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree(), TestCreateSecretsPreflight() (+10 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "VPNRepl"
Cohesion: 0.12
Nodes (29): ReplRole, VPNRepl, ReplicationConfigResult, mateConvergenceShutdowns(), missingListedVPNs(), PlannedRoles(), reenableLine(), replicationVPNLines() (+21 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.19
Nodes (28): SiteState, BuildSwitchPlan(), ExecuteSwitchPlan(), actionStrings(), ordersToA(), phaseActions(), replSites(), replVPNs() (+20 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.14
Nodes (42): uploadedForRole(), cliScriptPath(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestDetectRoleAddrsError() (+34 more)

### Community 25 - ".releaseToBackup"
Cohesion: 0.12
Nodes (16): field(), TestLastLines(), TestLastLinesEqualCount(), TestZipConfigsScript(), zipConfigsScript(), defaultLocalAddrs(), Ops, hostMatches() (+8 more)

### Community 26 - "renderDriver"
Cohesion: 0.18
Nodes (19): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), shQuote() (+11 more)

### Community 27 - "exportconfig_test.go"
Cohesion: 0.08
Nodes (41): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch(), opCanI() (+33 more)

### Community 28 - "Config"
Cohesion: 0.10
Nodes (27): Config, Scaling, allCustomMounted(), TestValidateStillChecksTheClassWhenOnlySomeNodesAreMounted(), OperatorNamespace(), HARoles(), lbServiceName(), ProductKeyRoles() (+19 more)

### Community 31 - "validate.go"
Cohesion: 0.07
Nodes (31): keyValueEntries, CanonicalDuration(), TestCanonicalDuration(), checkContainerHostPath(), CheckHostPath(), expandTilde(), Config, IsAbsHostPath() (+23 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "importops_test.go"
Cohesion: 0.12
Nodes (28): planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, minimalCapture() (+20 more)

### Community 34 - "domaincerts_resolve_test.go"
Cohesion: 0.07
Nodes (50): DirReader, fakeDirEntry, os.DirEntry, os.FileInfo, os.FileMode, caNameSafe(), checkNoDotDot(), DefaultDirReader() (+42 more)

### Community 35 - "Role"
Cohesion: 0.07
Nodes (17): downloadErrTransport, fakeTransport, recDownload, recRun, recUpload, recUploadFile, runErrMatchTransport, runErrTransport (+9 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.15
Nodes (10): TestCanIAnswerReadsTheLastLine(), NewTransport(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest() (+2 more)

### Community 38 - "transform_test.go"
Cohesion: 0.13
Nodes (23): injectedBlock(), ClearExistingSyslogs(), InjectVPNServiceShutdown(), Capture, loadRealCapture(), TestClearExistingClientCAsIntersects(), TestClearExistingNestedCoversAllThreeVerbs(), TestClearExistingSyslogsAgainstTheRealCapture() (+15 more)

### Community 39 - "replicationops_test.go"
Cohesion: 0.23
Nodes (18): cliScriptNameFromDest(), hasCall(), TestDisableDefaultUsersShowVPNError(), TestReleaseToBackupReleasedTimeout(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching() (+10 more)

### Community 40 - "statusreport_test.go"
Cohesion: 0.20
Nodes (17): Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList(), TestBrokerReportWithNoBrokerDeployed() (+9 more)

### Community 41 - "cli/platform_test.go"
Cohesion: 0.29
Nodes (16): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+8 more)

### Community 42 - "inject.go"
Cohesion: 0.22
Nodes (14): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+6 more)

### Community 43 - "usagef"
Cohesion: 0.09
Nodes (28): childExitError, usageError, github.com/spf13/cobra.PositionalArgs, noRolePositional(), registerFlagCompletion(), App, newConvertCmd(), runConvert() (+20 more)

### Community 44 - "NewMateChannel"
Cohesion: 0.23
Nodes (23): execArgs(), NewMateChannel(), ReadSecretKey(), secretCommand(), hasPrefixArgv(), indexOf(), operandOf(), replCfg() (+15 more)

### Community 45 - "render.go"
Cohesion: 0.19
Nodes (22): strings.Builder, Placement, PodAffinityTerm, boolStr(), cut(), parsePort(), parseToleration(), sortedKeys() (+14 more)

### Community 46 - "Platform"
Cohesion: 0.22
Nodes (22): checkFlagPlatforms(), commandPlatforms(), declaredList(), flagOnlyOn(), App, parsePlatformList(), platformSuffix(), prepare() (+14 more)

### Community 47 - "broker/replication.go"
Cohesion: 0.15
Nodes (24): colSpan, MateConfig, QueueState, cliTransport(), containsEndpoint(), dashSpans(), flagByte(), gutterClear() (+16 more)

### Community 48 - "Get"
Cohesion: 0.17
Nodes (23): Example, DetectPlatforms(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint(), TestDetectPlatformsMissingFile(), TestParsePlatform(), TestPlatforms(), writeEnv() (+15 more)

### Community 49 - "Ops"
Cohesion: 0.14
Nodes (10): time.Duration, cliRunNames(), countContains(), Ops, runCLISkeleton(), validName(), ValidVPNName(), validVPNName() (+2 more)

### Community 50 - "runner_test.go"
Cohesion: 0.07
Nodes (41): interactiveFailRunner, Exec, os/exec.Cmd, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveMissingBinaryIsActionable() (+33 more)

### Community 51 - "newRootCmd"
Cohesion: 0.08
Nodes (38): TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias(), TestAllowCommandIsRegisteredWhereItExecutes(), collectPaths() (+30 more)

### Community 52 - "opPerformReplication"
Cohesion: 0.36
Nodes (13): mateChannelFunc, confirmReplicationSwitch(), App, opConfigureReplication(), opCtrConfigureReplication(), opCtrPerformReplication(), opK8sConfigureReplication(), opK8sPerformReplication() (+5 more)

### Community 53 - "broker/matechannel_test.go"
Cohesion: 0.21
Nodes (18): CLIRunner, fakeRun, NewCLIMate(), mateReplies(), newTestMate(), TestLocalMateRunsThroughOps(), TestMateChannelDescribeNamesTheTarget(), TestMateChannelPreflightLearnsTheBrokerType() (+10 more)

### Community 54 - "Operations"
Cohesion: 0.10
Nodes (21): Bringing up a fresh cluster, Configuring a site, Data replication, Docker and Podman mechanics, Every destructive command confirms, Exit codes, Extra CLI users differ by platform, Operations (+13 more)

### Community 55 - "Test catalogue"
Cohesion: 0.10
Nodes (20): abbrev_test.go, convert_test.go, Coverage, examples_test.go, internal/abbrev, internal/convert, internal/engine, internal/examples (+12 more)

### Community 56 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 57 - "diff.go"
Cohesion: 0.16
Nodes (20): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, ancestorReported(), blockLabel(), diffRow() (+12 more)

### Community 58 - ".MateRevertActivity"
Cohesion: 0.10
Nodes (26): Credential, mateTarget, TestHTTPStatusHelpers(), anyHTTP2xx(), bridgeHostPort(), curlConfigFlag(), curlConfigLine(), Ops (+18 more)

### Community 60 - ".rpc"
Cohesion: 0.18
Nodes (10): AdminState, sempExecuteResult, sempMate, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, encoding/xml.Name, Ops (+2 more)

### Community 61 - "checkreport.go"
Cohesion: 0.15
Nodes (20): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), containsString(), failRow(), Cluster (+12 more)

### Community 62 - "Cluster"
Cohesion: 0.14
Nodes (5): time.Time, Cluster, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

### Community 63 - "container/preflight_test.go"
Cohesion: 0.40
Nodes (4): TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestComposeSecretEnvNamesCannotBeSystemVars(), TestPreflightIsPreviewableUnderDryRun(), TestPreflightRunsBeforeAnything()

### Community 65 - "Configuration"
Cohesion: 0.17
Nodes (12): Bring your own TLS Secret, Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling, Scaling tiers (+4 more)

### Community 66 - ".ConfigureReplication"
Cohesion: 0.19
Nodes (7): cliMate, Ops, isRunCLIRejection(), newlineIf(), replTransportFailure(), showReplicationScript(), showVPNReplicationScript()

### Community 67 - "broker/replication_test.go"
Cohesion: 0.15
Nodes (30): ParseShowReplication(), ParseVPNReplication(), RenderMate(), loadShowReplication(), mateAppliance(), mateSoftware(), TestMateRoundTripsThroughItsOwnGrammar(), TestParseShowReplicationAppliance() (+22 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "blocks.go"
Cohesion: 0.18
Nodes (16): Omission, Region, isMarker(), markerRegion(), TestIsMarkerDiscriminatesNamespace(), brokerTypeFromSchema(), Capture, isIndented() (+8 more)

### Community 71 - "localCfg"
Cohesion: 0.12
Nodes (36): assertNoPasswordInArgv(), TestLocalAdminIsTheDeploymentsOwnLogin(), TestMateRevertActivityBridgePlaintextOnly(), TestMateRevertActivityBridgeTLSNoCA(), TestMateRevertActivityCredsAndBodyNeverInArgv(), TestMateRevertActivityNon2xx(), TestMateRevertActivityRPCNotOK(), TestMateRevertActivitySuccess() (+28 more)

### Community 72 - "New"
Cohesion: 0.14
Nodes (22): TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), TestCheckReportSkipsEmptySections(), New(), NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey() (+14 more)

### Community 73 - "internal/k8s"
Cohesion: 0.11
Nodes (19): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go, names_test.go (+11 more)

### Community 74 - "annotate_test.go"
Cohesion: 0.11
Nodes (32): applyMetaField(), Capture, markerSection(), markerVerb(), parseKV(), readMarker(), renderKV(), sectionBeginMarker() (+24 more)

### Community 75 - "Set"
Cohesion: 0.14
Nodes (14): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+6 more)

### Community 76 - "namespace_test.go"
Cohesion: 0.30
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 77 - "newTestOps"
Cohesion: 0.07
Nodes (56): recOutput, Ops, newTestOps(), outputForRole(), TestCountContains(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs() (+48 more)

### Community 78 - "MateChannel"
Cohesion: 0.23
Nodes (13): PhaseKind, SwitchAction, SwitchPhase, MateChannel, Ops, confirmDemotions(), SwitchPlan, promoteConfirmed() (+5 more)

### Community 79 - "DiffBlocks"
Cohesion: 0.26
Nodes (16): DiffBlocks(), normalizeLine(), qualifiedKeys(), newDiffBlock(), TestDiffBlocksAbsentBlockLandsInMissingNotDiffering(), TestDiffBlocksDirectionalGotMaySupersetWant(), TestDiffBlocksIdentityIgnoresIndex(), TestDiffBlocksIgnorePredicateDropsLinesAndWholeIgnoredBlocks() (+8 more)

### Community 80 - "scaling.go"
Cohesion: 0.16
Nodes (14): scalingKey, scalingSpelling, scalingTier, containerMem(), Config, Scaling, Scaling, yaml.Node (+6 more)

### Community 81 - "ParseBlocks"
Cohesion: 0.22
Nodes (17): Annotate(), TestAnnotateIsDeterministic(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), ParseBlocks(), blockByIndex(), Capture, loadSample() (+9 more)

### Community 82 - "age"
Cohesion: 0.22
Nodes (9): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+1 more)

### Community 83 - "k8s/inspect_test.go"
Cohesion: 0.15
Nodes (19): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+11 more)

### Community 84 - "github.com/spf13/cobra.Command"
Cohesion: 0.21
Nodes (20): github.com/spf13/cobra.Command, applyAliases(), addCommands(), addLogFlags(), newOperatorCmd(), newOperatorDeployCmd(), newOperatorGenerateCmd(), newOperatorLogsCmd() (+12 more)

### Community 85 - "hostpath_test.go"
Cohesion: 0.14
Nodes (20): guardConfig(), Config, setContainerName(), TestBaseNameSplitsOnBothSeparators(), TestContainerCertRequiresKey(), TestContainerHostPathsRefuseATilde(), TestContainerNameMustBeAnEngineName(), TestDataDirMustBeAbsolute() (+12 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.24
Nodes (16): brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), brokerList, brokerStatus, condition, containerSpec (+8 more)

### Community 87 - ".preflightOne"
Cohesion: 0.44
Nodes (3): canIAnswer(), Cluster, probe

### Community 88 - "commands.go"
Cohesion: 0.22
Nodes (45): opFunc, roleOpFunc, addPodFlag(), App, group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerConfigureCmd() (+37 more)

### Community 89 - "load"
Cohesion: 0.15
Nodes (29): BrokerCR(), ContainerSecrets(), SecretPreflight(), assertNoCheckoutPath(), envLines(), healthCheckFixture(), load(), TestAdditionalUsersReachBothHalves() (+21 more)

### Community 90 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.09
Nodes (39): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+31 more)

### Community 92 - "exit_test.go"
Cohesion: 0.22
Nodes (11): runFailRunner, childExit(), childStatusError(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitKeepsItsMessage(), TestEveryArgValidatorIsAUsageError(), TestExitCodeIsNeverNegative(), TestHelperExitProcess() (+3 more)

### Community 93 - "cli_test.go"
Cohesion: 0.06
Nodes (108): os.File, testing.M, allowRuntime(), capture(), captureStderr(), echoRunner(), fakeBinaryOnPath(), firstLine() (+100 more)

### Community 94 - "containerState"
Cohesion: 0.21
Nodes (10): containerState, healthState, decodeInspect(), Manager, orUnknown(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings() (+2 more)

### Community 95 - "ReplSite"
Cohesion: 0.34
Nodes (15): confirmReplicationConfig(), mateChannel(), mateSEMPPassword(), App, replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer() (+7 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "config/replication_test.go"
Cohesion: 0.27
Nodes (12): Replication, Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationAbsent() (+4 more)

### Community 100 - "tlsCfg"
Cohesion: 0.43
Nodes (7): Config, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsMakeTheSecretOurs(), TestSuppliedCertsNeedASecretName(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), tlsCfg()

### Community 101 - "scripts.go"
Cohesion: 0.07
Nodes (43): showCmd, validCLILine(), Ops, assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript() (+35 more)

### Community 102 - "InjectShutdown"
Cohesion: 0.43
Nodes (13): InjectShutdown(), enablementCounts(), linesOf(), parseInjectCapture(), TestInjectShutdownAlreadySandwichedPassesThroughUnchanged(), TestInjectShutdownCaseThreeRestoresTargetState(), TestInjectShutdownCaseThreeUnknownTargetLeavesPortUnchanged(), TestInjectShutdownInjectsLeadingWhenTrailingExists() (+5 more)

### Community 103 - "container/secrets_test.go"
Cohesion: 0.21
Nodes (18): ResolveSecretValues(), certFixture(), TestDeletePodmanRemovesTheBundleAfterTheUnit(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeWritingTheComposeFile(), TestDeployDockerPassesTheBundleAsEnvNotArgv(), TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(), TestDeployPodmanFailsWhenTheBundleCannotBeWritten() (+10 more)

### Community 104 - "watch_test.go"
Cohesion: 0.13
Nodes (21): desiredWatch(), Cluster, splitWatch(), subtractWatch(), deployJSON(), Cluster, TestDesiredWatchAppendsTheBrokerNamespace(), TestOperatorReleaseNarrowsInsteadOfRemoving() (+13 more)

### Community 105 - "internal/cli"
Cohesion: 0.17
Nodes (12): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, examples_test.go, exit_test.go (+4 more)

### Community 106 - "operatorversion.go"
Cohesion: 0.24
Nodes (9): compareVersions(), Cluster, imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestImageTag() (+1 more)

### Community 107 - "container/inspect_test.go"
Cohesion: 0.26
Nodes (11): inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer() (+3 more)

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "execguard_test.go"
Cohesion: 0.18
Nodes (13): Config, guardCommandOf(), setGuardCommand(), TestAllowCommandsAccepts(), TestAllowCommandsRejects(), TestAllowCommandsRejectsEscalation(), TestAllowedBinaryIsNotGloballyAllowed(), TestComposeCommandDerivation() (+5 more)

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "pullSecretCfg"
Cohesion: 0.39
Nodes (8): Config, pullSecretCfg(), TestImagePullSecretDerivedSuffixFitsWhereAdditionalUsersAlreadyFits(), TestImagePullSecretNameBringYourOwn(), TestImagePullSecretNameConfiguredNameWins(), TestImagePullSecretNameDerivedWhenCredentialsButNoName(), TestImagePullSecretNameEmptyWithNeither(), TestManagesImagePullSecretNeedsBothCredentials()

### Community 112 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, File descriptors on rootless podman, Import reported failure, Podman secret flags, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 113 - "internal/config"
Cohesion: 0.13
Nodes (15): command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go, hostpath_test.go, imagepullsecret_test.go (+7 more)

### Community 114 - "Command"
Cohesion: 0.22
Nodes (3): Command, yaml.Node, Config

### Community 115 - "completion.go"
Cohesion: 0.31
Nodes (10): github.com/spf13/cobra.ShellCompDirective, completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), completionShell(), App, isEnvFileName() (+2 more)

### Community 116 - "abbrevdoc_test.go"
Cohesion: 0.35
Nodes (10): shorthand, mdRow(), renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes() (+2 more)

### Community 117 - "internal/container"
Cohesion: 0.29
Nodes (7): inspect_test.go, internal/container, manager_test.go, preflight_test.go, runtime_test.go, secrets_test.go, transport_test.go

### Community 118 - "allowcommand_test.go"
Cohesion: 0.29
Nodes (10): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+2 more)

### Community 119 - "load.go"
Cohesion: 0.16
Nodes (17): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults() (+9 more)

### Community 120 - "StripMarkers"
Cohesion: 0.18
Nodes (10): StripMarkers(), assertFencesBalanced(), TestAnnotateRegionAndSectionFencesAreBalanced(), TestStripMarkersKeepsPragmaAndSectionComments(), TestStripMarkersNoMarkersPassthrough(), splitLines(), stripTranscript(), TestTranscriptBannerType() (+2 more)

### Community 121 - "NewManager"
Cohesion: 0.26
Nodes (12): NewManager(), TestManagerNilSinks(), Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime() (+4 more)

### Community 122 - "newBlock"
Cohesion: 0.24
Nodes (10): token, firstQuoted(), newBlock(), TestNewBlockMessageSpoolQualifierHasNoOwnName(), TestNewBlockOpener(), TestNewBlockVPNNameWithSpaces(), TestNewBlockVPNQualifierWithSpaces(), tokenizeOpener() (+2 more)

### Community 124 - "Compose"
Cohesion: 0.13
Nodes (20): HealthCheck, Config, NodeIdentity, Compose(), composeEscape(), ComposeProject(), escapePercent(), healthCmd() (+12 more)

### Community 125 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 126 - "storage_test.go"
Cohesion: 0.33
Nodes (9): Config, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 127 - "cli/examples_test.go"
Cohesion: 0.25
Nodes (7): TestExamplesAcceptsAPlatformAbbreviation(), TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), TestExamplesNeedsNoEnvFile(), TestExamplesRefusesToOverwrite(), TestExamplesRejectsAnUnknownPlatform(), TestExamplesWritesOutFile()

### Community 128 - "ParseRole"
Cohesion: 0.36
Nodes (7): TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), TestRoleNames(), ParseRole(), RoleAbbrev(), RoleNames()

### Community 129 - "EnvPairs"
Cohesion: 0.25
Nodes (7): containerSecretSpecs(), EnvPairs(), groupKey(), itoa(), TestScalingReachesContainersAsEnv(), EnvPair, secretSpec

### Community 130 - "newEchoMgr"
Cohesion: 0.12
Nodes (18): bytes.Buffer, Manager, newEchoMgr(), rootlessNoFileMgr(), setNoFile(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployPodmanDryRunHidesSecretBytes() (+10 more)

### Community 131 - "importIgnore"
Cohesion: 0.47
Nodes (5): importIgnore(), TestImportIgnoreKeepsUnclassifiedSectionsInTheDiff(), TestImportOpsImportIgnore(), BridgeEnablementInverted(), TestShutdownBridgesLinesAreExcusedByTheDiff()

### Community 133 - "broker/transport.go"
Cohesion: 0.50
Nodes (4): CLIArg(), cliArg(), CLIScriptPath(), shellScriptPath()

## Knowledge Gaps
- **217 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+212 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 326 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Config` connect `Config` to `bg`, `EnvPairs`, `newEchoMgr`, `Image`, `Manager`, `config.go`, `GenOperator`, `capRunner`, `NewCluster`, `convert_test.go`, `ctrCfg`, `k8s/secrets_test.go`, `prep_test.go`, `exportconfig_test.go`, `Role`, `recRunner`, `NewMateChannel`, `render.go`, `Ops`, `.MateRevertActivity`, `checkreport.go`, `Cluster`, `localCfg`, `namespace_test.go`, `newTestOps`, `load`, `cli_test.go`, `config/replication_test.go`, `container/secrets_test.go`, `watch_test.go`, `load.go`, `NewManager`, `Compose`?**
  _High betweenness centrality (0.044) - this node is a cross-community bridge._
- **Why does `Platform` connect `Platform` to `bg`, `newEchoMgr`, `testing.T`, `Manager`, `config.go`, `CheckCommand`, `capRunner`, `convert_test.go`, `ctrCfg`, `exportconfig_test.go`, `validate.go`, `Role`, `Get`, `Ops`, `.MateRevertActivity`, `scaling.go`, `hostpath_test.go`, `commands.go`, `load`, `container/secrets_test.go`, `execguard_test.go`, `Command`, `load.go`, `NewManager`?**
  _High betweenness centrality (0.041) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `ParseRole`, `Manager`, `config.go`, `importops.go`, `verify_local_test.go`, `.releaseToBackup`, `renderDriver`, `Config`, `validate.go`, `NewMateChannel`, `Ops`, `opPerformReplication`, `.ConfigureReplication`, `Set`, `newTestOps`, `MateChannel`, `scripts.go`, `Compose`?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 34 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 34 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _217 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `bg` be split into smaller, more focused modules?**
  _Cohesion score 0.05326797385620915 - nodes in this community are weakly interconnected._
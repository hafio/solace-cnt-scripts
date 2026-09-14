# Graph Report - solace-k8-scripts  (2026-09-14)

## Corpus Check
- 180 files · ~508,548 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 3327 nodes · 10500 edges · 127 communities (117 shown, 6 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 1931 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `57394ed9`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- cli_test.go
- Sink
- context.Context
- Get
- testing.T
- Commands
- Manager
- config.go
- Config
- execguard_test.go
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
- runRootWith
- verify_local_test.go
- ops_container.go
- renderDriver
- .releaseToBackup
- strings.Builder
- solace
- validate.go
- CLAUDE.md
- importops_test.go
- kubectlTransport
- Role
- .resolveSecretRefs
- recRunner
- switchplan_test.go
- eqArgs
- statusreport_test.go
- cli/platform_test.go
- InjectShutdown
- exportconfig_test.go
- NewMateChannel
- opRunner
- broker/replication_test.go
- ReplSite
- matchCLI
- Ops
- runner_test.go
- newRootCmd
- cli/replication_test.go
- container/inspect_test.go
- Operations
- Test catalogue
- Exporting and importing configuration
- DiffBlocks
- .MateRevertActivity
- Cluster
- containerTransport
- checkreport.go
- TestServerCert
- replicationops_test.go
- Configuration
- broker.go
- MateChannel
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
- allowcommand_test.go
- Command
- tierFor
- blocks.go
- age
- k8s/inspect_test.go
- hostpath_test.go
- completion_test.go
- k8s/inspect.go
- .preflightOne
- github.com/spf13/cobra.Command
- render.go
- judge
- RuleFor
- storage_test.go
- broker/matechannel_test.go
- containerState
- ReplRole
- Abbreviations
- manager.go
- tlsCfg
- scripts.go
- opPerformReplication
- container/secrets_test.go
- watch_test.go
- internal/cli
- .LeaderLocal
- operatorversion.go
- What `import-config` applies
- config/replication_test.go
- internal/broker
- exit_test.go
- Troubleshooting
- internal/config
- k8s/runtime_test.go
- Fixtures and doubles
- haCfg
- internal/container
- check.go
- load.go
- internal/engine
- TestMain
- New
- cliMate
- .hostPath
- emitOrWrite
- ParseRole

## God Nodes (most connected - your core abstractions)
1. `newTestOps()` - 126 edges
2. `Role` - 116 edges
3. `ctrCfg()` - 115 edges
4. `newCapMgr()` - 100 edges
5. `Config` - 94 edges
6. `NewCluster()` - 89 edges
7. `matchCLI()` - 79 edges
8. `loadK8s()` - 77 edges
9. `Platform` - 72 edges
10. `bg()` - 68 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `ExitCode()`  [EXTRACTED]
  main.go → internal/cli/exit.go
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

## Communities (127 total, 6 thin omitted)

### Community 0 - "bg"
Cohesion: 0.15
Nodes (50): TestConfirmDeleteShortcut(), confirmDelete(), bg(), domainCANames(), App, k8sCluster(), k8sContext(), k8sLogin() (+42 more)

### Community 1 - "cli_test.go"
Cohesion: 0.08
Nodes (56): os.File, capture(), captureStderr(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone(), runStatusStderr() (+48 more)

### Community 2 - "Sink"
Cohesion: 0.21
Nodes (3): KV, Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.07
Nodes (8): scriptedMate, context.Context, Cluster, Cluster, Cluster, Cluster, badJSONRunner, fixtureRunner

### Community 4 - "Get"
Cohesion: 0.16
Nodes (24): Example, TestExamplesEmitsToStdout(), DetectPlatforms(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint(), TestDetectPlatformsMissingFile(), TestParsePlatform(), TestPlatforms() (+16 more)

### Community 5 - "testing.T"
Cohesion: 0.04
Nodes (110): testing.T, decodeRuntime(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds() (+102 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.12
Nodes (3): os.FileMode, NodeIdentity, Manager

### Community 8 - "config.go"
Cohesion: 0.09
Nodes (25): AdditionalUser, Broker, ContainerSecurity, DomainCerts, Image, Node, Operator, PodSecurity (+17 more)

### Community 9 - "Config"
Cohesion: 0.12
Nodes (26): Config, OperatorNamespace(), GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator() (+18 more)

### Community 10 - "execguard_test.go"
Cohesion: 0.12
Nodes (27): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+19 more)

### Community 11 - "capRunner"
Cohesion: 0.18
Nodes (14): capCall, capRunner, New(), TestNewDefaults(), Transport, NewTransport(), dockerCfg(), podmanCfg() (+6 more)

### Community 12 - "NewCluster"
Cohesion: 0.09
Nodes (59): TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts(), TestValidateReportsThePreSharedKeyChoice() (+51 more)

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
Cohesion: 0.07
Nodes (102): fileExists(), NewManager(), assertMode(), containsStr(), ctrCfg(), Manager, hasCall(), maskedKeys() (+94 more)

### Community 17 - "importops.go"
Cohesion: 0.11
Nodes (27): BrokerType, ImportResult, PlannedSection, bannerType(), checkProvenance(), checkSameType(), defaultVPNFirst(), describeScope() (+19 more)

### Community 18 - "k8s/secrets_test.go"
Cohesion: 0.14
Nodes (25): GenBroker(), GenSecrets(), joinManifests(), namespaceManifest(), TestUpdateServerCertSecret(), AdditionalUsersSecret(), AdminSecret(), checkGolden() (+17 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "prep_test.go"
Cohesion: 0.25
Nodes (13): adminCfg(), TestCreateNamespaceApplyFails(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree(), TestCreateSecretsPreflight(), TestCreateSecretsStopsOnPreflightFailure(), TestDeleteNamespaceStopsOnPreflightFailure(), TestDeleteSecrets() (+5 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "newTestOps"
Cohesion: 0.07
Nodes (50): Ops, newTestOps(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultVPN(), TestDomainCerts(), TestDomainCertsEmptySkips() (+42 more)

### Community 23 - "runRootWith"
Cohesion: 0.12
Nodes (42): allowRuntime(), echoRunner(), App, runCtr(), runRootWith(), TestAnnounceCommandsNamesResolvedBinaries(), TestCLICommand(), TestConfigStepsDoNotLeakSecrets() (+34 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.13
Nodes (43): cliScriptPath(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestDetectRoleAddrsError(), TestDetectRoleAmbiguousAddress() (+35 more)

### Community 25 - "ops_container.go"
Cohesion: 0.15
Nodes (40): TestCtrManagerConfirmWiring(), confirmAction(), wantRemove(), containerRenderRole(), containerRole(), containerWhat(), ctrLogin(), ctrManager() (+32 more)

### Community 26 - "renderDriver"
Cohesion: 0.18
Nodes (19): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), shQuote() (+11 more)

### Community 27 - ".releaseToBackup"
Cohesion: 0.18
Nodes (11): field(), TestField(), TestLastLines(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), TestLastLinesEqualCount(), activity(), Ops (+3 more)

### Community 28 - "strings.Builder"
Cohesion: 0.11
Nodes (36): shorthand, github.com/spf13/pflag.FlagSet, strings.Builder, mdRow(), renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands() (+28 more)

### Community 31 - "validate.go"
Cohesion: 0.12
Nodes (19): keyValueEntries, checkCredentialChars(), foldToEnvVar(), Config, missingErr(), platformKey(), requireAll(), requireKeyValue() (+11 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "importops_test.go"
Cohesion: 0.10
Nodes (30): planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, minimalCapture() (+22 more)

### Community 35 - "Role"
Cohesion: 0.07
Nodes (28): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+20 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.16
Nodes (9): NewTransport(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest(), recRunner (+1 more)

### Community 38 - "switchplan_test.go"
Cohesion: 0.19
Nodes (28): SiteState, BuildSwitchPlan(), SwitchPreflight(), actionStrings(), ordersToA(), replSites(), replVPNs(), st() (+20 more)

### Community 39 - "eqArgs"
Cohesion: 0.28
Nodes (12): TestPreflightRunsBeforeAnything(), Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime(), TestManagerReachableProbesRuntimeThenCompose() (+4 more)

### Community 40 - "statusreport_test.go"
Cohesion: 0.20
Nodes (17): Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList(), TestBrokerReportWithNoBrokerDeployed() (+9 more)

### Community 41 - "cli/platform_test.go"
Cohesion: 0.22
Nodes (20): logArgs(), App, logArgs2(), runPlatform(), TestLogArgsBuildsOneSetForBothPlatforms(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects() (+12 more)

### Community 42 - "InjectShutdown"
Cohesion: 0.17
Nodes (27): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+19 more)

### Community 43 - "exportconfig_test.go"
Cohesion: 0.21
Nodes (18): exportconfigDriverOutput(), exportconfigTransportOutput(), exportconfigVPNListOutput(), App, newExportconfigRunner(), newExportconfigRunnerRejectingChunk(), runImportConfirmation(), TestExportConfigOutOverwriteGate() (+10 more)

### Community 44 - "NewMateChannel"
Cohesion: 0.22
Nodes (24): Runner, execArgs(), NewMateChannel(), ReadSecretKey(), secretCommand(), hasPrefixArgv(), indexOf(), operandOf() (+16 more)

### Community 45 - "opRunner"
Cohesion: 0.13
Nodes (20): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch(), opCanI() (+12 more)

### Community 46 - "broker/replication_test.go"
Cohesion: 0.09
Nodes (54): colSpan, MateConfig, QueueState, cliTransport(), containsEndpoint(), dashSpans(), flagByte(), gutterClear() (+46 more)

### Community 47 - "ReplSite"
Cohesion: 0.11
Nodes (28): VPNRepl, Replication, Ops, isRunCLIRejection(), mateConvergenceShutdowns(), missingListedVPNs(), newlineIf(), PlannedRoles() (+20 more)

### Community 48 - "matchCLI"
Cohesion: 0.06
Nodes (49): matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultVPNDisableError() (+41 more)

### Community 49 - "Ops"
Cohesion: 0.17
Nodes (9): time.Duration, Ops, confirmImport(), App, pluralVPN(), runExport(), runImport(), TestPluralVPNWording() (+1 more)

### Community 50 - "runner_test.go"
Cohesion: 0.10
Nodes (29): bytes.Buffer, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveRefusesCurrentDirectory(), verboseExec(), captureStdout() (+21 more)

### Community 51 - "newRootCmd"
Cohesion: 0.11
Nodes (22): TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias(), TestAllowCommandIsRegisteredWhereItExecutes(), collectPaths() (+14 more)

### Community 52 - "cli/replication_test.go"
Cohesion: 0.24
Nodes (18): ReplicationConfigResult, mateChannel(), mateSEMPPassword(), reportReplicationConfig(), App, replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply() (+10 more)

### Community 53 - "container/inspect_test.go"
Cohesion: 0.18
Nodes (17): decodeInspect(), inspectReply(), kvValue(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings(), TestInspectNameLosesDockersLeadingSlash(), TestInspectRejectsAnEmptyReply() (+9 more)

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

### Community 58 - ".MateRevertActivity"
Cohesion: 0.14
Nodes (18): Credential, mateTarget, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigFlag(), curlConfigLine(), Ops, httpBody() (+10 more)

### Community 61 - "checkreport.go"
Cohesion: 0.19
Nodes (16): additionalUsersRow(), containsString(), failRow(), Cluster, info(), okRow(), portRows(), pskRow() (+8 more)

### Community 62 - "TestServerCert"
Cohesion: 0.23
Nodes (9): TestPathHelpers(), TestServerCert(), serverCertFile(), serverCertScript(), TestServerCertScript(), certPath(), CLIArg(), cliArg() (+1 more)

### Community 63 - "replicationops_test.go"
Cohesion: 0.29
Nodes (15): hasCall(), TestDisableDefaultUsersShowVPNError(), TestReleaseToBackupReleasedTimeout(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), TestConfigureReplicationClosingShowFailureStillReportsTheApply(), TestConfigureReplicationPhase1BodyOrderIsShutdownsThenRemovalsThenMate() (+7 more)

### Community 65 - "Configuration"
Cohesion: 0.17
Nodes (12): Bring your own TLS Secret, Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling, Scaling tiers (+4 more)

### Community 66 - "broker.go"
Cohesion: 0.21
Nodes (10): cliRunNames(), countContains(), runCLISkeleton(), TestCountContains(), TestValidName(), validName(), ValidVPNName(), validVPNName() (+2 more)

### Community 67 - "MateChannel"
Cohesion: 0.22
Nodes (14): PhaseKind, SwitchAction, SwitchPhase, MateChannel, Ops, confirmDemotions(), ExecuteSwitchPlan(), SwitchPlan (+6 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "ParseBlocks"
Cohesion: 0.15
Nodes (20): TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), ParseBlocks(), splitLines(), blockByIndex(), Capture, loadSample(), TestParseBlocksCRLFMatchesLF() (+12 more)

### Community 71 - "localCfg"
Cohesion: 0.09
Nodes (44): bridgeHostPort(), sempPort(), assertNoPasswordInArgv(), TestLastHTTPStatus(), TestLocalAdminIsTheDeploymentsOwnLogin(), TestMateRevertActivityBridgePlaintextOnly(), TestMateRevertActivityBridgeTLSNoCA(), TestMateRevertActivityCredsAndBodyNeverInArgv() (+36 more)

### Community 72 - "output_test.go"
Cohesion: 0.24
Nodes (14): sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLevelOrderKeepsSkipBelowFail(), TestLineAddsNoPrefix(), TestSectionPadsToWidth() (+6 more)

### Community 73 - "internal/k8s"
Cohesion: 0.11
Nodes (19): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go, names_test.go (+11 more)

### Community 74 - "annotate_test.go"
Cohesion: 0.10
Nodes (40): Annotate(), applyMetaField(), Capture, isMarker(), markerSection(), markerVerb(), parseKV(), readMarker() (+32 more)

### Community 75 - "Set"
Cohesion: 0.14
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 76 - "namespace_test.go"
Cohesion: 0.30
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 77 - "io.Writer"
Cohesion: 0.17
Nodes (11): interactiveFailRunner, Exec, io.Writer, os/exec.Cmd, TestResolveMissingBinaryIsActionable(), Echo, MaskEnv(), NewExec() (+3 more)

### Community 78 - "allowcommand_test.go"
Cohesion: 0.29
Nodes (10): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+2 more)

### Community 79 - "Command"
Cohesion: 0.18
Nodes (8): Command, decodeStrict(), Config, guardCommandOf(), setGuardCommand(), TestAllowCommandIsNotASchemaKey(), TestValidatorAndExecutorAgree(), yaml.Node

### Community 80 - "tierFor"
Cohesion: 0.27
Nodes (7): scalingTier, containerMem(), Config, TestContainerMem(), TestScalingTiers(), TestTierForRejectsOffTierValues(), tierFor()

### Community 81 - "blocks.go"
Cohesion: 0.12
Nodes (26): Omission, Region, token, markerRegion(), TestMarkerRegionBeginAndEnd(), brokerTypeFromSchema(), firstQuoted(), Capture (+18 more)

### Community 82 - "age"
Cohesion: 0.22
Nodes (9): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+1 more)

### Community 83 - "k8s/inspect_test.go"
Cohesion: 0.15
Nodes (19): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+11 more)

### Community 84 - "hostpath_test.go"
Cohesion: 0.09
Nodes (26): CanonicalDuration(), TestCanonicalDuration(), guardConfig(), BaseName(), CheckHostPath(), HasPathSeparator(), IsAbsHostPath(), Config (+18 more)

### Community 85 - "completion_test.go"
Cohesion: 0.25
Nodes (15): runComplete(), TestAllowCommandOffersNoFiles(), TestCompletionHelpStillWorks(), TestCompletionNeedsAShell(), TestCompletionNoDescriptions(), TestCompletionScriptsGenerate(), TestDirFlagCompletesDirectories(), TestEnvFlagCompletesEnvFiles() (+7 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.15
Nodes (20): time.Time, brokerConditionRows(), conditionLevel(), findCondition(), Cluster, normalizeToList(), TestConditionLevelDegradesSafely(), TestNormalizeToListHandlesBothKubectlShapes() (+12 more)

### Community 87 - ".preflightOne"
Cohesion: 0.44
Nodes (3): canIAnswer(), Cluster, probe

### Community 88 - "github.com/spf13/cobra.Command"
Cohesion: 0.06
Nodes (132): exportconfigReadCounter, layer, opFunc, roleOpFunc, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, io.Reader, applyAliases() (+124 more)

### Community 89 - "render.go"
Cohesion: 0.07
Nodes (63): HealthCheck, boolStr(), BrokerCR(), Compose(), composeEscape(), ComposeProject(), ContainerSecrets(), containerSecretSpecs() (+55 more)

### Community 90 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.10
Nodes (38): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+30 more)

### Community 92 - "storage_test.go"
Cohesion: 0.33
Nodes (9): Config, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 93 - "broker/matechannel_test.go"
Cohesion: 0.21
Nodes (18): CLIRunner, fakeRun, NewCLIMate(), mateReplies(), newTestMate(), TestLocalMateRunsThroughOps(), TestMateChannelDescribeNamesTheTarget(), TestMateChannelPreflightLearnsTheBrokerType() (+10 more)

### Community 94 - "containerState"
Cohesion: 0.35
Nodes (4): containerState, healthState, Manager, orUnknown()

### Community 95 - "ReplRole"
Cohesion: 0.13
Nodes (13): AdminState, ReplRole, sempExecuteResult, sempMate, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, encoding/xml.Name (+5 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "manager.go"
Cohesion: 0.23
Nodes (8): composeNeedsSecretValues(), exactName(), orNone(), platformTitle(), secretSummary(), setOrMissing(), splitLimit(), Manager

### Community 100 - "tlsCfg"
Cohesion: 0.43
Nodes (7): Config, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsMakeTheSecretOurs(), TestSuppliedCertsNeedASecretName(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), tlsCfg()

### Community 101 - "scripts.go"
Cohesion: 0.07
Nodes (41): showCmd, validCLILine(), Ops, defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript(), enableDefaultUsersScript() (+33 more)

### Community 102 - "opPerformReplication"
Cohesion: 0.19
Nodes (20): mateChannelFunc, bufio.Reader, App, lineSink(), progress(), step(), warn(), confirmReplicationConfig() (+12 more)

### Community 103 - "container/secrets_test.go"
Cohesion: 0.21
Nodes (18): ResolveSecretValues(), certFixture(), TestDeletePodmanRemovesTheBundleAfterTheUnit(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeWritingTheComposeFile(), TestDeployDockerPassesTheBundleAsEnvNotArgv(), TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(), TestDeployPodmanFailsWhenTheBundleCannotBeWritten() (+10 more)

### Community 104 - "watch_test.go"
Cohesion: 0.13
Nodes (21): desiredWatch(), Cluster, splitWatch(), subtractWatch(), deployJSON(), Cluster, TestDesiredWatchAppendsTheBrokerNamespace(), TestOperatorReleaseNarrowsInsteadOfRemoving() (+13 more)

### Community 105 - "internal/cli"
Cohesion: 0.17
Nodes (12): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, examples_test.go, exit_test.go (+4 more)

### Community 106 - ".LeaderLocal"
Cohesion: 0.21
Nodes (8): assertLeaderScript(), showRedundancyDetailScript(), TestAssertLeaderScript(), defaultLocalAddrs(), Ops, hostMatches(), shortHost(), TestDefaultLocalAddrs()

### Community 107 - "operatorversion.go"
Cohesion: 0.24
Nodes (9): compareVersions(), Cluster, imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestImageTag() (+1 more)

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "config/replication_test.go"
Cohesion: 0.36
Nodes (11): Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationAbsent(), TestValidateReplicationEmptyViaAccepted() (+3 more)

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "exit_test.go"
Cohesion: 0.10
Nodes (27): childExitError, runFailRunner, usageError, github.com/spf13/cobra.PositionalArgs, asUsage(), childExit(), childStatus(), ExitCode() (+19 more)

### Community 112 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, File descriptors on rootless podman, Import reported failure, Podman secret flags, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 113 - "internal/config"
Cohesion: 0.17
Nodes (12): command_test.go, config_test.go, duration_test.go, execguard_test.go, hostpath_test.go, internal/config, platform_test.go, redundancy_test.go (+4 more)

### Community 114 - "k8s/runtime_test.go"
Cohesion: 0.43
Nodes (7): TestClusterHonoursRuntime(), TestExecutorRefusesUnapprovedRuntime(), TestRuntimeDefaultArgvUnchanged(), TestTransportHonoursRuntime(), unapprovedCfg(), withLeading(), wrappedCfg()

### Community 115 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 116 - "haCfg"
Cohesion: 0.14
Nodes (19): TestCheckStopsProbingWhenUnreachable(), TestResolveStorageClass(), TestStorageRows(), haCfg(), saCfg(), TestHARoles(), TestProductKeyRoles(), TestRestartOrder() (+11 more)

### Community 117 - "internal/container"
Cohesion: 0.29
Nodes (7): inspect_test.go, internal/container, manager_test.go, preflight_test.go, runtime_test.go, secrets_test.go, transport_test.go

### Community 118 - "check.go"
Cohesion: 0.40
Nodes (4): orValue(), setOrMissing(), setOrNone(), storageClassSuitable()

### Community 119 - "load.go"
Cohesion: 0.17
Nodes (17): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults() (+9 more)

### Community 120 - "internal/engine"
Cohesion: 0.67
Nodes (3): internal/engine, resolve_test.go, runner_test.go

### Community 122 - "New"
Cohesion: 0.12
Nodes (11): TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), TestCheckReportSkipsEmptySections(), Cluster, New() (+3 more)

### Community 124 - "cliMate"
Cohesion: 0.36
Nodes (3): cliMate, showReplicationScript(), showVPNReplicationScript()

### Community 128 - "emitOrWrite"
Cohesion: 0.38
Nodes (6): App, newConvertCmd(), runConvert(), emitOrWrite(), opK8sGenBroker(), opK8sGenOperator()

### Community 135 - "ParseRole"
Cohesion: 0.36
Nodes (7): TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), TestRoleNames(), ParseRole(), RoleAbbrev(), RoleNames()

## Knowledge Gaps
- **213 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+208 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 319 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **6 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Config` connect `Config` to `bg`, `Manager`, `config.go`, `capRunner`, `NewCluster`, `convert_test.go`, `ctrCfg`, `k8s/secrets_test.go`, `prep_test.go`, `newTestOps`, `runRootWith`, `strings.Builder`, `kubectlTransport`, `Role`, `recRunner`, `eqArgs`, `NewMateChannel`, `opRunner`, `ReplSite`, `Ops`, `containerTransport`, `checkreport.go`, `localCfg`, `namespace_test.go`, `render.go`, `opPerformReplication`, `container/secrets_test.go`, `watch_test.go`, `k8s/runtime_test.go`, `haCfg`, `load.go`, `New`?**
  _High betweenness centrality (0.047) - this node is a cross-community bridge._
- **Why does `Platform` connect `github.com/spf13/cobra.Command` to `Get`, `testing.T`, `Manager`, `config.go`, `execguard_test.go`, `capRunner`, `convert_test.go`, `ctrCfg`, `validate.go`, `eqArgs`, `opRunner`, `Ops`, `containerTransport`, `localCfg`, `Command`, `tierFor`, `hostpath_test.go`, `render.go`, `manager.go`, `opPerformReplication`, `container/secrets_test.go`, `load.go`?**
  _High betweenness centrality (0.038) - this node is a cross-community bridge._
- **Why does `Sink` connect `Sink` to `opPerformReplication`, `Manager`, `output_test.go`, `importops.go`, `Ops`, `k8s/inspect_test.go`, `age`, `DiffBlocks`, `New`, `checkreport.go`?**
  _High betweenness centrality (0.032) - this node is a cross-community bridge._
- **Are the 80 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 80 INFERRED edges - model-reasoned connections that need verification._
- **Are the 34 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 34 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _213 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `bg` be split into smaller, more focused modules?**
  _Cohesion score 0.14901960784313725 - nodes in this community are weakly interconnected._
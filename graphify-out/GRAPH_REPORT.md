# Graph Report - solace-k8-scripts  (2026-09-14)

## Corpus Check
- 180 files · ~509,939 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 3330 nodes · 10516 edges · 131 communities (120 shown, 7 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 1936 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `18854f02`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- runRoot
- Sink
- context.Context
- Get
- config_test.go
- Commands
- Manager
- Config
- operator.go
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
- haCfg
- dev.ps1
- newTestOps
- cli_test.go
- verify_local_test.go
- ops_container.go
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
- recRunner
- switchplan_test.go
- helpers.go
- statusreport_test.go
- cli/platform_test.go
- InjectShutdown
- runRootWith
- NewMateChannel
- opRunner
- broker/replication.go
- ReplSite
- matchCLI
- Ops
- testing.T
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
- Ops
- broker/replication_test.go
- Command reference
- Developer guide
- ParseBlocks
- localCfg
- New
- internal/k8s
- annotate_test.go
- Set
- namespace_test.go
- io.Writer
- validContainerConfig
- Command
- scaling_test.go
- blocks.go
- age
- k8s/inspect_test.go
- hostpath_test.go
- completion_test.go
- k8s/inspect.go
- .preflightOne
- commands.go
- render.go
- judge
- RuleFor
- storage_test.go
- github.com/spf13/cobra.Command
- containerState
- .rpc
- Abbreviations
- manager.go
- tlsCfg
- scripts.go
- opPerformReplication
- container/secrets_test.go
- watch_test.go
- internal/cli
- Load
- operatorversion.go
- What `import-config` applies
- config/replication_test.go
- internal/broker
- usagef
- Troubleshooting
- internal/config
- k8s/runtime_test.go
- Fixtures and doubles
- operatorversion_test.go
- internal/container
- exit_test.go
- load.go
- execguard_test.go
- abbrevdoc_test.go
- completion.go
- ReplVia
- omitAtExport
- ResolveEnvPath
- TestCanonicalDuration
- .hostPath
- emitOrWrite
- TestCheckStopsProbingWhenUnreachable
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

## Communities (131 total, 7 thin omitted)

### Community 0 - "bg"
Cohesion: 0.15
Nodes (49): confirmDelete(), bg(), domainCANames(), App, k8sCluster(), k8sContext(), k8sLogin(), k8sOps() (+41 more)

### Community 1 - "runRoot"
Cohesion: 0.08
Nodes (36): runRoot(), runStandalone(), TestBashEnvGivenToEnvFlag(), TestConfigureServerCertsRefusesASecretItDoesNotOwn(), TestConvertErrorPaths(), TestConvertParseError(), TestConvertRoundTrip(), TestConvertToFile() (+28 more)

### Community 2 - "Sink"
Cohesion: 0.18
Nodes (3): KV, Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.08
Nodes (8): scriptedMate, context.Context, Cluster, Cluster, Cluster, Cluster, desiredWatch(), Cluster

### Community 4 - "Get"
Cohesion: 0.15
Nodes (25): Example, App, newExamplesCmd(), runExample(), DetectPlatforms(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint(), TestDetectPlatformsMissingFile() (+17 more)

### Community 5 - "config_test.go"
Cohesion: 0.08
Nodes (38): assertContainerBlockDefaults(), assertContainerScaling(), Config, haNodesConfig(), TestAdditionalUserNameIsStricterOnKubernetes(), TestAdditionalUserPasswordCharsAreFreeOnKubernetes(), TestApplyDefaultsDocker(), TestApplyDefaultsK8s() (+30 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.12
Nodes (3): os.FileMode, NodeIdentity, Manager

### Community 8 - "Config"
Cohesion: 0.08
Nodes (28): AdditionalUser, Broker, Container, ContainerSecurity, DockerConfig, DomainCerts, Image, Network (+20 more)

### Community 9 - "operator.go"
Cohesion: 0.14
Nodes (15): GenOperator(), Cluster, joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator(), renderOperatorNS() (+7 more)

### Community 10 - "CheckCommand"
Cohesion: 0.16
Nodes (19): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+11 more)

### Community 11 - "capRunner"
Cohesion: 0.12
Nodes (25): capCall, capRunner, New(), TestNewDefaults(), Transport, Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime() (+17 more)

### Community 12 - "NewCluster"
Cohesion: 0.08
Nodes (62): TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts(), TestValidateReportsThePreSharedKeyChoice() (+54 more)

### Community 13 - "eqArgs"
Cohesion: 0.17
Nodes (23): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), TestCLIAndShellAreInteractive() (+15 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (62): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+54 more)

### Community 15 - "Block"
Cohesion: 0.08
Nodes (56): Block, TargetState, regexp.Regexp, importIgnore(), TestImportIgnoreKeepsUnclassifiedSectionsInTheDiff(), TestImportOpsImportIgnore(), injectedBlock(), BridgeEnablementInverted() (+48 more)

### Community 16 - "ctrCfg"
Cohesion: 0.06
Nodes (107): bytes.Buffer, TestRestartCountComesFromSystemdOnPodman(), TestRestartCountUnknownWhenSystemctlFails(), fileExists(), NewManager(), assertMode(), containsStr(), ctrCfg() (+99 more)

### Community 17 - "importops.go"
Cohesion: 0.23
Nodes (15): ImportResult, PlannedSection, defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList(), renderRegionScript() (+7 more)

### Community 18 - "k8s/secrets_test.go"
Cohesion: 0.12
Nodes (30): GenBroker(), GenSecrets(), joinManifests(), namespaceManifest(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret() (+22 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "haCfg"
Cohesion: 0.13
Nodes (25): TestStorageRows(), ProductKeyRoles(), haCfg(), saCfg(), TestHARoles(), TestProductKeyRoles(), TestRestartOrder(), TestOperatorImage() (+17 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "newTestOps"
Cohesion: 0.09
Nodes (45): Ops, newTestOps(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultVPN(), TestDomainCerts(), TestDomainCertsEmptySkips(), TestDomainCertsRejectsBadName() (+37 more)

### Community 23 - "cli_test.go"
Cohesion: 0.09
Nodes (60): testing.M, allowRuntime(), captureStderr(), echoRunner(), fakeBinaryOnPath(), firstLine(), runCtr(), runStatusStderr() (+52 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.13
Nodes (43): cliScriptPath(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestDetectRoleAddrsError(), TestDetectRoleAmbiguousAddress() (+35 more)

### Community 25 - "ops_container.go"
Cohesion: 0.16
Nodes (39): confirmAction(), wantRemove(), containerRenderRole(), containerRole(), containerWhat(), ctrLogin(), ctrManager(), ctrOps() (+31 more)

### Community 26 - "renderDriver"
Cohesion: 0.18
Nodes (19): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), shQuote() (+11 more)

### Community 27 - "Platform"
Cohesion: 0.17
Nodes (26): checkFlagPlatforms(), commandPlatforms(), declaredList(), flagOnlyOn(), App, parsePlatformList(), platformSuffix(), prepare() (+18 more)

### Community 28 - "commanddoc_test.go"
Cohesion: 0.29
Nodes (13): github.com/spf13/pflag.FlagSet, anchor(), argumentLine(), availableSubs(), firstDiff(), mdCell(), renderCommandDocs(), TestCommandDocs() (+5 more)

### Community 31 - "validate.go"
Cohesion: 0.13
Nodes (17): keyValueEntries, checkCredentialChars(), foldToEnvVar(), Config, missingErr(), platformKey(), requireAll(), requireKeyValue() (+9 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "importops_test.go"
Cohesion: 0.10
Nodes (31): planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, minimalCapture() (+23 more)

### Community 35 - "Role"
Cohesion: 0.06
Nodes (30): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+22 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.15
Nodes (10): TestCanIAnswerReadsTheLastLine(), NewTransport(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest() (+2 more)

### Community 38 - "switchplan_test.go"
Cohesion: 0.08
Nodes (59): CLIRunner, fakeRun, PhaseKind, SiteState, SwitchAction, SwitchPhase, MateChannel, Ops (+51 more)

### Community 39 - "helpers.go"
Cohesion: 0.17
Nodes (23): exportconfigReadCounter, layer, io.Reader, TestStdinCanAnswerClosedFile(), registerFlagCompletion(), addAllowCommandFlag(), addExportFlags(), addOutFlags() (+15 more)

### Community 40 - "statusreport_test.go"
Cohesion: 0.12
Nodes (19): Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList(), TestBrokerReportWithNoBrokerDeployed() (+11 more)

### Community 41 - "cli/platform_test.go"
Cohesion: 0.29
Nodes (16): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+8 more)

### Community 42 - "InjectShutdown"
Cohesion: 0.17
Nodes (27): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+19 more)

### Community 43 - "runRootWith"
Cohesion: 0.15
Nodes (28): App, runRootWith(), TestK8sConfigDeleteDomainCertsConfigured(), TestRemoveServerCertsRefusedOnASecretManagedDeployment(), ExitCode(), TestChildExitStatusIsScopedToInteractiveSessions(), exportconfigDriverOutput(), exportconfigTransportOutput() (+20 more)

### Community 44 - "NewMateChannel"
Cohesion: 0.24
Nodes (22): Runner, execArgs(), NewMateChannel(), ReadSecretKey(), secretCommand(), hasPrefixArgv(), indexOf(), operandOf() (+14 more)

### Community 45 - "opRunner"
Cohesion: 0.18
Nodes (12): opCall, opRunner, failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), opArgvMatch(), opCanI(), opFailOn(), opFailOnCount() (+4 more)

### Community 46 - "broker/replication.go"
Cohesion: 0.11
Nodes (31): colSpan, MateConfig, QueueState, cliTransport(), containsEndpoint(), dashSpans(), flagByte(), gutterClear() (+23 more)

### Community 47 - "ReplSite"
Cohesion: 0.12
Nodes (31): ReplRole, VPNRepl, Ops, ReplicationConfigResult, isRunCLIRejection(), mateConvergenceShutdowns(), missingListedVPNs(), newlineIf() (+23 more)

### Community 48 - "matchCLI"
Cohesion: 0.06
Nodes (50): matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultVPNDisableError() (+42 more)

### Community 49 - "Ops"
Cohesion: 0.18
Nodes (8): time.Duration, Ops, confirmImport(), App, pluralVPN(), runExport(), runImport(), nowStamp()

### Community 50 - "testing.T"
Cohesion: 0.10
Nodes (42): os.File, testing.T, capture(), captureStdout(), loadDirect(), TestCheckDeployWarnsWhenOperatorAbsent(), TestConfirmDeleteShortcut(), TestContainerRoleDetectsFromHostname() (+34 more)

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (31): TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias(), TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRegisteredWhereItExecutes() (+23 more)

### Community 52 - "cli/replication_test.go"
Cohesion: 0.38
Nodes (13): mateChannel(), mateSEMPPassword(), App, replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer(), TestMateChannelPicksTheMechanismTheSiteDeclares() (+5 more)

### Community 53 - "container/inspect_test.go"
Cohesion: 0.29
Nodes (9): inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer(), TestRestartCountUnknownRatherThanZero() (+1 more)

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
Cohesion: 0.13
Nodes (20): Credential, mateTarget, TestHTTPStatusHelpers(), anyHTTP2xx(), bridgeHostPort(), curlConfigFlag(), curlConfigLine(), Ops (+12 more)

### Community 61 - "checkreport.go"
Cohesion: 0.15
Nodes (21): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), allCustomMounted(), containsString(), failRow() (+13 more)

### Community 62 - "TestServerCert"
Cohesion: 0.23
Nodes (10): TestDiagnostics(), TestPathHelpers(), TestServerCert(), serverCertFile(), serverCertScript(), TestServerCertScript(), certPath(), CLIArg() (+2 more)

### Community 63 - "replicationops_test.go"
Cohesion: 0.25
Nodes (18): hasCall(), TestDisableDefaultUsersShowVPNError(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), TestConfigureReplicationClosingShowFailureStillReportsTheApply(), TestConfigureReplicationPhase1BodyOrderIsShutdownsThenRemovalsThenMate() (+10 more)

### Community 65 - "Configuration"
Cohesion: 0.17
Nodes (12): Bring your own TLS Secret, Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling, Scaling tiers (+4 more)

### Community 66 - "Ops"
Cohesion: 0.12
Nodes (12): cliRunNames(), countContains(), runCLISkeleton(), TestCountContains(), TestValidName(), validCLILine(), validName(), ValidVPNName() (+4 more)

### Community 67 - "broker/replication_test.go"
Cohesion: 0.21
Nodes (24): ParseShowReplication(), RenderMate(), SameEndpoints(), loadShowReplication(), mateAppliance(), mateSoftware(), TestMateRoundTripsThroughItsOwnGrammar(), TestParseShowReplicationAppliance() (+16 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "ParseBlocks"
Cohesion: 0.10
Nodes (21): BrokerType, cliMate, ParseBlocks(), splitLines(), TestParseBlocksCRLFMatchesLF(), TestParseBlocksIndentedLineWithNoOpenerRefused(), TestParseBlocksMissingEndTerminatorRefused(), bannerType() (+13 more)

### Community 71 - "localCfg"
Cohesion: 0.10
Nodes (41): sempPort(), assertNoPasswordInArgv(), TestLocalAdminIsTheDeploymentsOwnLogin(), TestMateRevertActivityBridgePlaintextOnly(), TestMateRevertActivityBridgeTLSNoCA(), TestMateRevertActivityCredsAndBodyNeverInArgv(), TestMateRevertActivityNon2xx(), TestMateRevertActivityRPCNotOK() (+33 more)

### Community 72 - "New"
Cohesion: 0.12
Nodes (24): TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), TestCheckReportSkipsEmptySections(), New(), NewFunc() (+16 more)

### Community 73 - "internal/k8s"
Cohesion: 0.11
Nodes (19): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go, names_test.go (+11 more)

### Community 74 - "annotate_test.go"
Cohesion: 0.09
Nodes (44): Annotate(), applyMetaField(), Capture, isMarker(), markerSection(), markerVerb(), parseKV(), readMarker() (+36 more)

### Community 75 - "Set"
Cohesion: 0.14
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 76 - "namespace_test.go"
Cohesion: 0.30
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 77 - "io.Writer"
Cohesion: 0.13
Nodes (16): interactiveFailRunner, Exec, io.Writer, os/exec.Cmd, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecVerboseAnnouncesEveryCommand(), TestResolveMissingBinaryIsActionable() (+8 more)

### Community 78 - "validContainerConfig"
Cohesion: 0.11
Nodes (24): TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNamesCollideOnDocker(), TestValidateAdditionalUserClashesWithABuiltIn(), TestValidateAdditionalUsers(), TestValidateContainerBadNetworkMode(), TestValidateContainerBridge(), TestValidateContainerHA(), TestValidateContainerIdentifiers() (+16 more)

### Community 79 - "Command"
Cohesion: 0.12
Nodes (15): decodeRuntime(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRuntimeDefaults() (+7 more)

### Community 80 - "scaling_test.go"
Cohesion: 0.11
Nodes (19): scalingTier, containerMem(), Config, Config, setContainerMem(), TestApplyScalingTierDefaultsContainerBlocks(), TestApplyScalingTierDefaultsK8s(), TestApplyScalingTierDefaultsMemOverride() (+11 more)

### Community 81 - "blocks.go"
Cohesion: 0.10
Nodes (33): Omission, Region, token, markerRegion(), TestAnnotateIsDeterministic(), brokerTypeFromSchema(), firstQuoted(), Capture (+25 more)

### Community 82 - "age"
Cohesion: 0.22
Nodes (9): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+1 more)

### Community 83 - "k8s/inspect_test.go"
Cohesion: 0.15
Nodes (19): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+11 more)

### Community 84 - "hostpath_test.go"
Cohesion: 0.13
Nodes (22): guardConfig(), CheckHostPath(), HasPathSeparator(), IsAbsHostPath(), Config, setContainerName(), TestCheckHostPathAccepts(), TestCheckHostPathRejects() (+14 more)

### Community 85 - "completion_test.go"
Cohesion: 0.31
Nodes (13): runComplete(), TestAllowCommandOffersNoFiles(), TestCompletionNeedsAShell(), TestCompletionNoDescriptions(), TestDirFlagCompletesDirectories(), TestEnvFlagCompletesEnvFiles(), TestEnvFlagPrefixFilters(), TestEnvFlagWithPathDefersToShell() (+5 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.15
Nodes (20): time.Time, brokerConditionRows(), conditionLevel(), findCondition(), Cluster, normalizeToList(), TestConditionLevelDegradesSafely(), TestNormalizeToListHandlesBothKubectlShapes() (+12 more)

### Community 87 - ".preflightOne"
Cohesion: 0.44
Nodes (3): canIAnswer(), Cluster, probe

### Community 88 - "commands.go"
Cohesion: 0.20
Nodes (46): opFunc, roleOpFunc, addLogFlags(), addPodFlag(), App, group(), newBrokerCLICmd(), newBrokerCmd() (+38 more)

### Community 89 - "render.go"
Cohesion: 0.06
Nodes (76): strings.Builder, Placement, PodAffinityTerm, boolStr(), BrokerCR(), Compose(), composeEscape(), ComposeProject() (+68 more)

### Community 90 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.11
Nodes (34): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+26 more)

### Community 92 - "storage_test.go"
Cohesion: 0.33
Nodes (9): Config, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 93 - "github.com/spf13/cobra.Command"
Cohesion: 0.32
Nodes (19): github.com/spf13/cobra.Command, applyAliases(), addCommands(), newOperatorCmd(), newOperatorDeployCmd(), newOperatorGenerateCmd(), newOperatorLogsCmd(), newOperatorRemoveCmd() (+11 more)

### Community 94 - "containerState"
Cohesion: 0.21
Nodes (10): containerState, healthState, decodeInspect(), Manager, orUnknown(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings() (+2 more)

### Community 95 - ".rpc"
Cohesion: 0.16
Nodes (11): AdminState, sempExecuteResult, sempMate, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, encoding/xml.Name, Ops (+3 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "manager.go"
Cohesion: 0.21
Nodes (9): composeNeedsSecretValues(), exactName(), orNone(), platformTitle(), secretSummary(), setOrMissing(), splitLimit(), TestSplitLimit() (+1 more)

### Community 100 - "tlsCfg"
Cohesion: 0.43
Nodes (7): Config, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsMakeTheSecretOurs(), TestSuppliedCertsNeedASecretName(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), tlsCfg()

### Community 101 - "scripts.go"
Cohesion: 0.05
Nodes (58): showCmd, field(), TestField(), TestLastLines(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), TestLastLinesEqualCount(), assertLeaderScript() (+50 more)

### Community 102 - "opPerformReplication"
Cohesion: 0.18
Nodes (21): mateChannelFunc, bufio.Reader, App, lineSink(), progress(), step(), warn(), confirmReplicationConfig() (+13 more)

### Community 103 - "container/secrets_test.go"
Cohesion: 0.21
Nodes (18): ResolveSecretValues(), certFixture(), TestDeletePodmanRemovesTheBundleAfterTheUnit(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeWritingTheComposeFile(), TestDeployDockerPassesTheBundleAsEnvNotArgv(), TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(), TestDeployPodmanFailsWhenTheBundleCannotBeWritten() (+10 more)

### Community 104 - "watch_test.go"
Cohesion: 0.15
Nodes (19): splitWatch(), subtractWatch(), deployJSON(), Cluster, TestDesiredWatchAppendsTheBrokerNamespace(), TestOperatorReleaseNarrowsInsteadOfRemoving(), TestReconcileWatchDecidesWhatToApply(), TestReconcileWatchFlagsAWideningAsAQuestion() (+11 more)

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

### Community 109 - "config/replication_test.go"
Cohesion: 0.27
Nodes (12): Replication, Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationAbsent() (+4 more)

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "usagef"
Cohesion: 0.13
Nodes (17): childExitError, usageError, github.com/spf13/cobra.PositionalArgs, logArgs(), asUsage(), childStatus(), isUsage(), markUsageArgs() (+9 more)

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

### Community 116 - "operatorversion_test.go"
Cohesion: 0.39
Nodes (7): operatorDeployJSON(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsSilentInAPreview(), TestConfirmNoDowngradeProceedsWhenConfirmAccepts(), TestConfirmNoDowngradeProceedsWhenVersionReadFails(), TestConfirmNoDowngradeRefusedWhenConfirmDeclines(), TestConfirmNoDowngradeRefusesByDefault()

### Community 117 - "internal/container"
Cohesion: 0.29
Nodes (7): inspect_test.go, internal/container, manager_test.go, preflight_test.go, runtime_test.go, secrets_test.go, transport_test.go

### Community 118 - "exit_test.go"
Cohesion: 0.21
Nodes (13): runFailRunner, childExit(), childStatusError(), k8sEnv(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitKeepsItsMessage(), TestEveryArgValidatorIsAUsageError(), TestExitCodeContract() (+5 more)

### Community 119 - "load.go"
Cohesion: 0.20
Nodes (12): TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config, parseError() (+4 more)

### Community 120 - "execguard_test.go"
Cohesion: 0.20
Nodes (13): decodeStrict(), Config, guardCommandOf(), setGuardCommand(), TestAllowCommandIsNotASchemaKey(), TestAllowCommandsAccepts(), TestAllowCommandsRejects(), TestAllowCommandsRejectsEscalation() (+5 more)

### Community 121 - "abbrevdoc_test.go"
Cohesion: 0.35
Nodes (10): shorthand, mdRow(), renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes() (+2 more)

### Community 122 - "completion.go"
Cohesion: 0.31
Nodes (10): github.com/spf13/cobra.ShellCompDirective, completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), completionShell(), App, isEnvFileName() (+2 more)

### Community 123 - "ReplVia"
Cohesion: 0.29
Nodes (7): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, viaDescription(), validateReplVia(), validateReplViaSEMP()

### Community 124 - "omitAtExport"
Cohesion: 0.32
Nodes (7): omitAtExport(), TestImportPlanAcceptsWhatExportProduced(), TestImportPlanRefusesAFileItDidNotExport(), TestOmitAtExportDropsFilteredLinesAndEmptyBlocks(), TestOmitAtExportLeavesTargetDecisionsToImport(), TestOmitAtExportRemovesSkipSectionsAndRecordsThem(), TestOmitExportConfigEndToEnd()

### Community 125 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 128 - "emitOrWrite"
Cohesion: 0.38
Nodes (6): App, newConvertCmd(), runConvert(), emitOrWrite(), opK8sGenBroker(), opK8sGenOperator()

### Community 135 - "ParseRole"
Cohesion: 0.36
Nodes (7): TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), TestRoleNames(), ParseRole(), RoleAbbrev(), RoleNames()

## Knowledge Gaps
- **213 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+208 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 319 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **7 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `Get`, `config_test.go`, `Manager`, `Config`, `CheckCommand`, `capRunner`, `convert_test.go`, `ctrCfg`, `validate.go`, `Ops`, `testing.T`, `containerTransport`, `localCfg`, `validContainerConfig`, `scaling_test.go`, `hostpath_test.go`, `commands.go`, `render.go`, `manager.go`, `opPerformReplication`, `container/secrets_test.go`, `Load`, `load.go`, `execguard_test.go`?**
  _High betweenness centrality (0.036) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `bg`, `context.Context`, `Manager`, `operator.go`, `capRunner`, `NewCluster`, `convert_test.go`, `ctrCfg`, `k8s/secrets_test.go`, `haCfg`, `newTestOps`, `cli_test.go`, `kubectlTransport`, `Role`, `recRunner`, `NewMateChannel`, `opRunner`, `Ops`, `testing.T`, `containerTransport`, `checkreport.go`, `localCfg`, `namespace_test.go`, `render.go`, `opPerformReplication`, `container/secrets_test.go`, `config/replication_test.go`, `k8s/runtime_test.go`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `Manager`, `Config`, `ParseRole`, `importops.go`, `haCfg`, `ops_container.go`, `renderDriver`, `validate.go`, `importops_test.go`, `kubectlTransport`, `switchplan_test.go`, `ReplSite`, `Ops`, `.MateRevertActivity`, `containerTransport`, `Ops`, `ParseBlocks`, `Set`, `scripts.go`, `opPerformReplication`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 34 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 34 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _213 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `runRoot` be split into smaller, more focused modules?**
  _Cohesion score 0.08108108108108109 - nodes in this community are weakly interconnected._
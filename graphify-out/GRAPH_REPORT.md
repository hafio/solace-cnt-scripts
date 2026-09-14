# Graph Report - solace-k8-scripts  (2026-09-14)

## Corpus Check
- 182 files · ~510,454 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 3337 nodes · 10525 edges · 122 communities (114 shown, 4 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 1940 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `17dc295d`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- cli_test.go
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
- prep_test.go
- dev.ps1
- newTestOps
- ops_container.go
- verify_local_test.go
- .releaseToBackup
- renderDriver
- exportconfig_test.go
- strings.Builder
- solace
- validate.go
- CLAUDE.md
- importops_test.go
- Ops
- Role
- .resolveSecretRefs
- recRunner
- confirmAction
- ReplRole
- statusreport_test.go
- cli/platform_test.go
- InjectShutdown
- newSEMPTestMate
- ReplSite
- opRunner
- broker/replication.go
- .ConfigureReplication
- testing.T
- broker.go
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
- exit_test.go
- checkreport.go
- Cluster
- matchCLI
- Configuration
- kubectlTransport
- broker/replication_test.go
- Command reference
- Developer guide
- broker/transport.go
- localCfg
- New
- internal/k8s
- annotate_test.go
- Set
- namespace_test.go
- .Run
- validContainerConfig
- switchplan_test.go
- scaling_test.go
- blocks.go
- age
- k8s/inspect_test.go
- Command
- Data replication
- k8s/inspect.go
- .preflightOne
- github.com/spf13/cobra.Command
- render.go
- judge
- RuleFor
- storage_test.go
- runRoot
- containerState
- Abbreviations
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
- haCfg
- Fixtures and doubles
- internal/container
- load.go
- ParseBlocks
- NewManager
- ResolveEnvPath
- runExport
- BrokerType
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

## Communities (122 total, 4 thin omitted)

### Community 0 - "bg"
Cohesion: 0.15
Nodes (50): bg(), domainCANames(), App, k8sCluster(), k8sContext(), k8sLogin(), k8sOps(), k8sWhat() (+42 more)

### Community 1 - "cli_test.go"
Cohesion: 0.08
Nodes (74): testing.M, allowRuntime(), captureStderr(), echoRunner(), fakeBinaryOnPath(), firstLine(), App, runCtr() (+66 more)

### Community 2 - "Sink"
Cohesion: 0.21
Nodes (3): KV, Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.08
Nodes (8): scriptedMate, context.Context, Manager, Cluster, Cluster, Cluster, badJSONRunner, fixtureRunner

### Community 4 - "Get"
Cohesion: 0.25
Nodes (17): Example, TestExamplesEmitsToStdout(), All(), Get(), List(), Names(), read(), firstDiff() (+9 more)

### Community 5 - "config_test.go"
Cohesion: 0.08
Nodes (38): assertContainerBlockDefaults(), assertContainerScaling(), Config, haNodesConfig(), TestAdditionalUserNameIsStricterOnKubernetes(), TestAdditionalUserPasswordCharsAreFreeOnKubernetes(), TestApplyDefaultsDocker(), TestApplyDefaultsK8s() (+30 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.09
Nodes (11): os.FileMode, NodeIdentity, composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary() (+3 more)

### Community 8 - "Config"
Cohesion: 0.08
Nodes (27): AdditionalUser, Broker, ContainerSecurity, DomainCerts, Image, Node, Operator, PodSecurity (+19 more)

### Community 9 - "operator.go"
Cohesion: 0.27
Nodes (13): GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator(), renderOperatorNS(), splitAfterNamespace() (+5 more)

### Community 10 - "CheckCommand"
Cohesion: 0.15
Nodes (21): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+13 more)

### Community 11 - "capRunner"
Cohesion: 0.22
Nodes (11): capCall, capRunner, NewTransport(), dockerCfg(), podmanCfg(), TestTransportCopy(), TestTransportEchoHidesSEMPConfig(), TestTransportEchoHidesUploadBody() (+3 more)

### Community 12 - "NewCluster"
Cohesion: 0.08
Nodes (65): TestCheckStopsProbingWhenUnreachable(), TestResolveStorageClass(), TestStorageRows(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts() (+57 more)

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
Nodes (102): bytes.Buffer, fileExists(), assertMode(), containsStr(), ctrCfg(), Manager, hasCall(), maskedKeys() (+94 more)

### Community 17 - "importops.go"
Cohesion: 0.15
Nodes (21): ImportResult, PlannedSection, checkProvenance(), checkSameType(), defaultVPNFirst(), describeScope(), ImportPlan, Ops (+13 more)

### Community 18 - "k8s/secrets_test.go"
Cohesion: 0.11
Nodes (31): GenBroker(), GenSecrets(), joinManifests(), namespaceManifest(), TestUpdateServerCertSecret(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret() (+23 more)

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
Cohesion: 0.06
Nodes (59): Ops, newTestOps(), outputForRole(), TestCountContains(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultVPN() (+51 more)

### Community 23 - "ops_container.go"
Cohesion: 0.15
Nodes (40): TestCtrManagerConfirmWiring(), wantEnable(), wantRemove(), containerRenderRole(), containerRole(), containerWhat(), ctrLogin(), ctrManager() (+32 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.15
Nodes (41): cliScriptPath(), curlCalls(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestDetectRoleAddrsError() (+33 more)

### Community 25 - ".releaseToBackup"
Cohesion: 0.12
Nodes (19): field(), assertLeaderScript(), noReleaseActivityScript(), releaseActivityScript(), showRedundancyDetailScript(), showRedundancyScript(), TestFixedScripts(), defaultLocalAddrs() (+11 more)

### Community 26 - "renderDriver"
Cohesion: 0.19
Nodes (17): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), TestKeywordPatternEscapesEachPhrase() (+9 more)

### Community 27 - "exportconfig_test.go"
Cohesion: 0.17
Nodes (20): exportconfigReadCounter, exportconfigDriverOutput(), exportconfigTransportOutput(), exportconfigVPNListOutput(), App, newExportconfigRunner(), newExportconfigRunnerRejectingChunk(), runImportConfirmation() (+12 more)

### Community 28 - "strings.Builder"
Cohesion: 0.11
Nodes (35): shorthand, github.com/spf13/pflag.FlagSet, strings.Builder, mdRow(), renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands() (+27 more)

### Community 31 - "validate.go"
Cohesion: 0.05
Nodes (45): keyValueEntries, CanonicalDuration(), TestCanonicalDuration(), guardConfig(), TestGuardConfigIsValid(), CheckHostPath(), Config, HasPathSeparator() (+37 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "importops_test.go"
Cohesion: 0.08
Nodes (37): planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, minimalCapture() (+29 more)

### Community 34 - "Ops"
Cohesion: 0.24
Nodes (4): time.Duration, Ops, New(), Transport

### Community 35 - "Role"
Cohesion: 0.06
Nodes (29): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+21 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.15
Nodes (9): TestCanIAnswerReadsTheLastLine(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest(), recRunner (+1 more)

### Community 38 - "confirmAction"
Cohesion: 0.25
Nodes (21): layer, io.Reader, io.Writer, TestConfirmDeleteShortcut(), TestPromptYes(), TestPromptYesNo(), TestStdinCanAnswerClosedFile(), confirmAction() (+13 more)

### Community 39 - "ReplRole"
Cohesion: 0.13
Nodes (15): AdminState, ReplRole, sempExecuteResult, sempMate, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, encoding/xml.Name (+7 more)

### Community 40 - "statusreport_test.go"
Cohesion: 0.20
Nodes (17): Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList(), TestBrokerReportWithNoBrokerDeployed() (+9 more)

### Community 41 - "cli/platform_test.go"
Cohesion: 0.29
Nodes (16): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+8 more)

### Community 42 - "InjectShutdown"
Cohesion: 0.17
Nodes (27): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+19 more)

### Community 43 - "newSEMPTestMate"
Cohesion: 0.29
Nodes (13): newSEMPTestMate(), sempReply(), sempSite(), TestSEMPMateChecksExecuteResult(), TestSEMPMateEscapesTheVPNName(), TestSEMPMatePlainTextURL(), TestSEMPMatePrimaryActive(), TestSEMPMateReadsBothMateShapes() (+5 more)

### Community 44 - "ReplSite"
Cohesion: 0.22
Nodes (24): ReplSite, execArgs(), NewMateChannel(), ReadSecretKey(), secretCommand(), hasPrefixArgv(), indexOf(), operandOf() (+16 more)

### Community 45 - "opRunner"
Cohesion: 0.11
Nodes (22): opCall, opRunner, os.File, capture(), captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect() (+14 more)

### Community 46 - "broker/replication.go"
Cohesion: 0.15
Nodes (24): colSpan, MateConfig, QueueState, cliTransport(), containsEndpoint(), dashSpans(), flagByte(), gutterClear() (+16 more)

### Community 47 - ".ConfigureReplication"
Cohesion: 0.13
Nodes (26): VPNRepl, Ops, ReplicationConfigResult, isRunCLIRejection(), mateConvergenceShutdowns(), missingListedVPNs(), newlineIf(), reenableLine() (+18 more)

### Community 48 - "testing.T"
Cohesion: 0.06
Nodes (61): testing.T, ranContains(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultVPNDisableError(), TestDisableDefaultVPNShowError(), TestDomainCertsBadFilename() (+53 more)

### Community 49 - "broker.go"
Cohesion: 0.17
Nodes (11): countContains(), runCLISkeleton(), validName(), ValidVPNName(), validVPNName(), shQuote(), TestShQuoteHandlesASingleQuote(), rejectionIn() (+3 more)

### Community 50 - "runner_test.go"
Cohesion: 0.15
Nodes (21): captureStdout(), helperCommand(), TestEchoDefaultWriter(), TestEchoOutput(), TestEchoOutputInput(), TestEchoRun(), TestEchoRunEnv(), TestEchoRunEnvNoEnv() (+13 more)

### Community 51 - "newRootCmd"
Cohesion: 0.08
Nodes (32): applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias(), TestAllowCommandApprovesAWrappedRuntime() (+24 more)

### Community 52 - "cli/replication_test.go"
Cohesion: 0.34
Nodes (14): mateChannel(), mateSEMPPassword(), App, replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer(), TestMateChannelPicksTheMechanismTheSiteDeclares() (+6 more)

### Community 53 - "container/inspect_test.go"
Cohesion: 0.26
Nodes (11): inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer() (+3 more)

### Community 54 - "Operations"
Cohesion: 0.12
Nodes (16): Bringing up a fresh cluster, Docker and Podman mechanics, Every destructive command confirms, Exit codes, Extra CLI users differ by platform, Operations, Post-deployment configuration order, Re-deploying is safe and explicit (+8 more)

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
Cohesion: 0.18
Nodes (14): Credential, mateTarget, anyHTTP2xx(), bridgeHostPort(), curlConfigFlag(), curlConfigLine(), Ops, httpBody() (+6 more)

### Community 60 - "exit_test.go"
Cohesion: 0.15
Nodes (18): runFailRunner, childExit(), ExitCode(), childStatusError(), k8sEnv(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitKeepsItsMessage(), TestChildExitStatusIsScopedToInteractiveSessions() (+10 more)

### Community 61 - "checkreport.go"
Cohesion: 0.15
Nodes (21): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), allCustomMounted(), containsString(), failRow() (+13 more)

### Community 63 - "matchCLI"
Cohesion: 0.18
Nodes (26): cliRunNames(), hasCall(), matchCLI(), TestDisableDefaultUsersShowVPNError(), TestReleaseToBackupReleasedTimeout(), TestImportPlanAcceptsWhatExportProduced(), TestOmitExportConfigEndToEnd(), replListed() (+18 more)

### Community 65 - "Configuration"
Cohesion: 0.17
Nodes (12): Bring your own TLS Secret, Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling, Scaling tiers (+4 more)

### Community 67 - "broker/replication_test.go"
Cohesion: 0.15
Nodes (30): ParseShowReplication(), ParseVPNReplication(), RenderMate(), loadShowReplication(), mateAppliance(), mateSoftware(), TestMateRoundTripsThroughItsOwnGrammar(), TestParseShowReplicationAppliance() (+22 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "broker/transport.go"
Cohesion: 0.47
Nodes (5): TestPathHelpers(), certPath(), CLIArg(), cliArg(), CLIScriptPath()

### Community 71 - "localCfg"
Cohesion: 0.10
Nodes (34): TestLocalMateRunsThroughOps(), sempPort(), assertNoPasswordInArgv(), TestCurlConfigLineEscapesQuotes(), TestHTTPBodyStripsHeaders(), TestLastHTTPStatus(), TestLocalAdminIsTheDeploymentsOwnLogin(), TestMateRevertActivityBridgePlaintextOnly() (+26 more)

### Community 72 - "New"
Cohesion: 0.14
Nodes (21): solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), TestCheckReportSkipsEmptySections(), New(), NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag() (+13 more)

### Community 73 - "internal/k8s"
Cohesion: 0.11
Nodes (19): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go, names_test.go (+11 more)

### Community 74 - "annotate_test.go"
Cohesion: 0.09
Nodes (40): applyMetaField(), Capture, isMarker(), markerSection(), markerVerb(), parseKV(), readMarker(), renderKV() (+32 more)

### Community 75 - "Set"
Cohesion: 0.14
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 76 - "namespace_test.go"
Cohesion: 0.30
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 77 - ".Run"
Cohesion: 0.11
Nodes (20): interactiveFailRunner, Exec, os/exec.Cmd, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveMissingBinaryIsActionable() (+12 more)

### Community 78 - "validContainerConfig"
Cohesion: 0.11
Nodes (24): TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNamesCollideOnDocker(), TestValidateAdditionalUserClashesWithABuiltIn(), TestValidateAdditionalUsers(), TestValidateContainerBadNetworkMode(), TestValidateContainerBridge(), TestValidateContainerHA(), TestValidateContainerIdentifiers() (+16 more)

### Community 79 - "switchplan_test.go"
Cohesion: 0.08
Nodes (58): fakeRun, PhaseKind, SiteState, SwitchAction, SwitchPhase, MateChannel, Ops, NewCLIMate() (+50 more)

### Community 80 - "scaling_test.go"
Cohesion: 0.11
Nodes (19): scalingTier, containerMem(), Config, Config, setContainerMem(), TestApplyScalingTierDefaultsContainerBlocks(), TestApplyScalingTierDefaultsK8s(), TestApplyScalingTierDefaultsMemOverride() (+11 more)

### Community 81 - "blocks.go"
Cohesion: 0.13
Nodes (24): Omission, Region, token, markerRegion(), brokerTypeFromSchema(), firstQuoted(), Capture, isIndented() (+16 more)

### Community 82 - "age"
Cohesion: 0.22
Nodes (9): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+1 more)

### Community 83 - "k8s/inspect_test.go"
Cohesion: 0.15
Nodes (19): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+11 more)

### Community 84 - "Command"
Cohesion: 0.13
Nodes (14): Command, decodeStrict(), Config, guardCommandOf(), setGuardCommand(), TestAllowCommandIsNotASchemaKey(), TestAllowCommandsAccepts(), TestAllowCommandsRejects() (+6 more)

### Community 85 - "Data replication"
Cohesion: 0.40
Nodes (5): Configuring a site, Data replication, Reaching the mate, Switching roles, When a run fails partway

### Community 86 - "k8s/inspect.go"
Cohesion: 0.14
Nodes (20): time.Time, brokerConditionRows(), conditionLevel(), findCondition(), Cluster, normalizeToList(), TestConditionLevelDegradesSafely(), TestNormalizeToListHandlesBothKubectlShapes() (+12 more)

### Community 87 - ".preflightOne"
Cohesion: 0.44
Nodes (3): canIAnswer(), Cluster, probe

### Community 88 - "github.com/spf13/cobra.Command"
Cohesion: 0.07
Nodes (119): opFunc, roleOpFunc, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, addCommands(), addLogFlags(), addPodFlag(), App (+111 more)

### Community 89 - "render.go"
Cohesion: 0.07
Nodes (64): Cluster, boolStr(), BrokerCR(), Compose(), composeEscape(), ComposeProject(), ContainerSecrets(), containerSecretSpecs() (+56 more)

### Community 90 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.10
Nodes (38): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+30 more)

### Community 92 - "storage_test.go"
Cohesion: 0.33
Nodes (9): Config, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 93 - "runRoot"
Cohesion: 0.08
Nodes (38): runRoot(), TestBashEnvGivenToEnvFlag(), TestConfigureServerCertsRefusesASecretItDoesNotOwn(), TestConvertErrorPaths(), TestConvertParseError(), TestConvertToFile(), TestConvertToStdout(), TestConvertWriteError() (+30 more)

### Community 94 - "containerState"
Cohesion: 0.21
Nodes (10): containerState, healthState, decodeInspect(), Manager, orUnknown(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings() (+2 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 100 - "tlsCfg"
Cohesion: 0.43
Nodes (7): Config, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsMakeTheSecretOurs(), TestSuppliedCertsNeedASecretName(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), tlsCfg()

### Community 101 - "scripts.go"
Cohesion: 0.07
Nodes (41): showCmd, validCLILine(), Ops, defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript(), enableDefaultUsersScript() (+33 more)

### Community 102 - "opPerformReplication"
Cohesion: 0.19
Nodes (19): mateChannelFunc, bufio.Reader, App, lineSink(), progress(), step(), confirmReplicationConfig(), confirmReplicationSwitch() (+11 more)

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

### Community 109 - "config/replication_test.go"
Cohesion: 0.36
Nodes (11): Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationAbsent(), TestValidateReplicationEmptyViaAccepted() (+3 more)

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "usagef"
Cohesion: 0.13
Nodes (16): childExitError, usageError, github.com/spf13/cobra.PositionalArgs, logArgs(), asUsage(), childStatus(), isUsage(), markUsageArgs() (+8 more)

### Community 112 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, File descriptors on rootless podman, Import reported failure, Podman secret flags, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 113 - "internal/config"
Cohesion: 0.17
Nodes (12): command_test.go, config_test.go, duration_test.go, execguard_test.go, hostpath_test.go, internal/config, platform_test.go, redundancy_test.go (+4 more)

### Community 114 - "haCfg"
Cohesion: 0.13
Nodes (24): TestDeleteBrokerProbesClaimsOnlyWhenPurging(), haCfg(), saCfg(), TestHARoles(), TestProductKeyRoles(), TestRestartOrder(), TestOperatorImage(), operatorDeployJSON() (+16 more)

### Community 115 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 117 - "internal/container"
Cohesion: 0.29
Nodes (7): inspect_test.go, internal/container, manager_test.go, preflight_test.go, runtime_test.go, secrets_test.go, transport_test.go

### Community 119 - "load.go"
Cohesion: 0.17
Nodes (17): Container, DockerConfig, Network, PodmanConfig, Ulimits, HealthCheck, TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator() (+9 more)

### Community 120 - "ParseBlocks"
Cohesion: 0.14
Nodes (23): Annotate(), TestAnnotateIsDeterministic(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), ParseBlocks(), splitLines(), blockByIndex(), Capture (+15 more)

### Community 121 - "NewManager"
Cohesion: 0.23
Nodes (14): NewManager(), TestManagerNilSinks(), TestPreflightRunsBeforeAnything(), Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime() (+6 more)

### Community 125 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 128 - "runExport"
Cohesion: 0.23
Nodes (11): App, newConvertCmd(), runConvert(), confirmImport(), App, pluralVPN(), runExport(), runImport() (+3 more)

### Community 129 - "BrokerType"
Cohesion: 0.21
Nodes (6): BrokerType, cliMate, CLIRunner, bannerType(), showReplicationScript(), showVPNReplicationScript()

### Community 135 - "ParseRole"
Cohesion: 0.36
Nodes (7): TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), TestRoleNames(), ParseRole(), RoleAbbrev(), RoleNames()

## Knowledge Gaps
- **213 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+208 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 319 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **4 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Config` connect `Config` to `bg`, `cli_test.go`, `Manager`, `operator.go`, `capRunner`, `NewCluster`, `convert_test.go`, `ctrCfg`, `k8s/secrets_test.go`, `prep_test.go`, `newTestOps`, `strings.Builder`, `Ops`, `Role`, `ReplSite`, `opRunner`, `checkreport.go`, `Cluster`, `kubectlTransport`, `localCfg`, `namespace_test.go`, `render.go`, `opPerformReplication`, `container/secrets_test.go`, `watch_test.go`, `haCfg`, `load.go`, `NewManager`?**
  _High betweenness centrality (0.049) - this node is a cross-community bridge._
- **Why does `Platform` connect `github.com/spf13/cobra.Command` to `config_test.go`, `Manager`, `Config`, `CheckCommand`, `capRunner`, `convert_test.go`, `ctrCfg`, `validate.go`, `Ops`, `Role`, `opRunner`, `localCfg`, `validContainerConfig`, `scaling_test.go`, `Command`, `render.go`, `opPerformReplication`, `container/secrets_test.go`, `Load`, `load.go`, `NewManager`?**
  _High betweenness centrality (0.043) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `runExport`, `BrokerType`, `bg`, `Manager`, `Config`, `ParseRole`, `importops.go`, `newTestOps`, `ops_container.go`, `.releaseToBackup`, `renderDriver`, `validate.go`, `Ops`, `ReplSite`, `.ConfigureReplication`, `broker.go`, `DiffBlocks`, `.MateRevertActivity`, `kubectlTransport`, `Set`, `switchplan_test.go`, `scripts.go`, `opPerformReplication`?**
  _High betweenness centrality (0.028) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 34 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 34 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _213 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `bg` be split into smaller, more focused modules?**
  _Cohesion score 0.14901960784313725 - nodes in this community are weakly interconnected._
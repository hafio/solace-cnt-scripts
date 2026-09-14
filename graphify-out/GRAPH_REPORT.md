# Graph Report - solace-k8-scripts  (2026-09-14)

## Corpus Check
- 184 files · ~518,771 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 3372 nodes · 10607 edges · 124 communities (112 shown, 8 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 1958 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `70ef401c`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- newTestOps
- Sink
- context.Context
- strings.Builder
- config_test.go
- Commands
- Manager
- config.go
- Config
- CheckCommand
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
- matchCLI
- switchplan_test.go
- verify_local_test.go
- .releaseToBackup
- renderDriver
- runRootWith
- writeNode
- solace
- validate.go
- CLAUDE.md
- importops_test.go
- DeriveCAName
- Role
- .resolveSecretRefs
- recRunner
- k8s/runtime_test.go
- ReplRole
- statusreport_test.go
- cli/platform_test.go
- InjectShutdown
- ContainerSecrets
- NewMateChannel
- opRunner
- broker/replication.go
- .ConfigureReplication
- ReplSite
- Ops
- testing.T
- newRootCmd
- cli/replication_test.go
- omit_test.go
- Operations
- Test catalogue
- Exporting and importing configuration
- DiffBlocks
- .MateRevertActivity
- Cluster
- ContainerSecret
- checkreport.go
- Cluster
- replicationops_test.go
- Configuration
- kubectlTransport
- broker/replication_test.go
- Command reference
- Developer guide
- internal/engine
- localCfg
- internal/k8s
- annotate_test.go
- Set
- namespace_test.go
- .Run
- redundancy_test.go
- MateChannel
- scaling.go
- blocks.go
- age
- k8s/inspect_test.go
- Command
- hostpath_test.go
- k8s/inspect.go
- .preflightOne
- github.com/spf13/cobra.Command
- load
- judge
- RuleFor
- cli_test.go
- container/inspect_test.go
- broker/matechannel_test.go
- Abbreviations
- tlsCfg
- scripts.go
- opPerformReplication
- ctrCfg
- watch_test.go
- internal/cli
- .operatorRows
- What `import-config` applies
- config/replication_test.go
- internal/broker
- Troubleshooting
- internal/config
- haCfg
- internal/container
- completion_test.go
- load.go
- blocks_test.go
- NewManager
- .LeaderLocal
- manager.go
- render.go
- containerTransport
- BrokerType
- rootlessNoFileMgr
- TestServerCert
- .hostPath
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

## Communities (124 total, 8 thin omitted)

### Community 0 - "bg"
Cohesion: 0.05
Nodes (118): exportconfigReadCounter, layer, io.Reader, io.Writer, time.Duration, Ops, TestStdinCanAnswerClosedFile(), confirmAction() (+110 more)

### Community 1 - "newTestOps"
Cohesion: 0.08
Nodes (46): Ops, newTestOps(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultVPN(), TestDomainCerts(), TestDomainCertsEmptySkips() (+38 more)

### Community 2 - "Sink"
Cohesion: 0.21
Nodes (3): KV, Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.08
Nodes (7): scriptedMate, context.Context, Cluster, Cluster, Cluster, badJSONRunner, fixtureRunner

### Community 4 - "strings.Builder"
Cohesion: 0.17
Nodes (25): shorthand, github.com/spf13/pflag.FlagSet, strings.Builder, mdRow(), renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands() (+17 more)

### Community 5 - "config_test.go"
Cohesion: 0.04
Nodes (93): assertContainerBlockDefaults(), assertContainerScaling(), envTree(), Config, haNodesConfig(), minimalK8s(), TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes() (+85 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 8 - "config.go"
Cohesion: 0.09
Nodes (24): AdditionalUser, Broker, ContainerSecurity, DomainCerts, Image, Node, Operator, PodSecurity (+16 more)

### Community 9 - "Config"
Cohesion: 0.14
Nodes (21): Config, Scaling, OperatorNamespace(), GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes() (+13 more)

### Community 10 - "CheckCommand"
Cohesion: 0.16
Nodes (20): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+12 more)

### Community 11 - "capRunner"
Cohesion: 0.17
Nodes (16): capCall, capRunner, New(), TestNewDefaults(), Transport, TestManagerLogsCLIShell(), NewTransport(), dockerCfg() (+8 more)

### Community 12 - "NewCluster"
Cohesion: 0.09
Nodes (58): TestStorageRows(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+50 more)

### Community 13 - "eqArgs"
Cohesion: 0.17
Nodes (23): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), TestCLIAndShellAreInteractive() (+15 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (62): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+54 more)

### Community 15 - "Block"
Cohesion: 0.09
Nodes (54): Block, TargetState, regexp.Regexp, importIgnore(), injectedBlock(), BridgeEnablementInverted(), ClearExistingClientCAs(), ClearExistingNested() (+46 more)

### Community 16 - "newCapMgr"
Cohesion: 0.08
Nodes (75): fileExists(), assertMode(), containsStr(), hasCall(), maskedKeys(), newCapMgr(), TestContainerRunningMatchesNameExactly(), TestManagerCheckDNSFailsLoudInHA() (+67 more)

### Community 17 - "importops.go"
Cohesion: 0.10
Nodes (32): ImportResult, PlannedSection, ParseBlocks(), splitLines(), checkProvenance(), checkSameType(), defaultVPNFirst(), describeScope() (+24 more)

### Community 18 - "k8s/secrets_test.go"
Cohesion: 0.12
Nodes (31): GenBroker(), GenSecrets(), joinManifests(), namespaceManifest(), TestUpdateServerCertSecret(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret() (+23 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "prep_test.go"
Cohesion: 0.23
Nodes (14): adminCfg(), TestCreateNamespaceApplyFails(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree(), TestCreateSecretsPreflight(), TestCreateSecretsStopsOnPreflightFailure(), TestDeleteNamespaceProtected(), TestDeleteNamespaceStopsOnPreflightFailure() (+6 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "matchCLI"
Cohesion: 0.06
Nodes (51): matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultUsersShowVPNError() (+43 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.25
Nodes (21): SiteState, BuildSwitchPlan(), ExecuteSwitchPlan(), actionStrings(), ordersToA(), phaseActions(), replVPNs(), st() (+13 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.13
Nodes (43): cliScriptPath(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestDetectRoleAddrsError(), TestDetectRoleAmbiguousAddress() (+35 more)

### Community 25 - ".releaseToBackup"
Cohesion: 0.18
Nodes (11): field(), TestField(), TestLastLines(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), TestLastLinesEqualCount(), activity(), Ops (+3 more)

### Community 26 - "renderDriver"
Cohesion: 0.18
Nodes (19): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), shQuote() (+11 more)

### Community 27 - "runRootWith"
Cohesion: 0.06
Nodes (57): childExitError, runFailRunner, usageError, github.com/spf13/cobra.PositionalArgs, App, runRootWith(), TestCLICommand(), TestDeployOperatorNoPromptStaysUnknownFlag() (+49 more)

### Community 28 - "writeNode"
Cohesion: 0.20
Nodes (12): Placement, PodAffinityTerm, parseToleration(), sortedKeys(), TestParseToleration(), writeMatchExprs(), writeNode(), writeNodeAffinity() (+4 more)

### Community 31 - "validate.go"
Cohesion: 0.12
Nodes (19): keyValueEntries, BrokerPodSuffixShape(), checkCredentialChars(), foldToEnvVar(), Config, missingErr(), platformKey(), requireAll() (+11 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "importops_test.go"
Cohesion: 0.08
Nodes (39): planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, minimalCapture() (+31 more)

### Community 34 - "DeriveCAName"
Cohesion: 0.27
Nodes (12): caNameSafe(), DeriveCAName(), lastPathElement(), MatchesCertExt(), sanitiseCAPart(), splitCAExt(), assertUsableCAName(), TestDeriveCAName() (+4 more)

### Community 35 - "Role"
Cohesion: 0.06
Nodes (29): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+21 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.15
Nodes (10): TestCanIAnswerReadsTheLastLine(), NewTransport(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest() (+2 more)

### Community 38 - "k8s/runtime_test.go"
Cohesion: 0.43
Nodes (7): TestClusterHonoursRuntime(), TestExecutorRefusesUnapprovedRuntime(), TestRuntimeDefaultArgvUnchanged(), TestTransportHonoursRuntime(), unapprovedCfg(), withLeading(), wrappedCfg()

### Community 39 - "ReplRole"
Cohesion: 0.13
Nodes (14): AdminState, ReplRole, sempExecuteResult, sempMate, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, encoding/xml.Name (+6 more)

### Community 40 - "statusreport_test.go"
Cohesion: 0.20
Nodes (17): Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList(), TestBrokerReportWithNoBrokerDeployed() (+9 more)

### Community 41 - "cli/platform_test.go"
Cohesion: 0.22
Nodes (20): logArgs(), App, logArgs2(), runPlatform(), TestLogArgsBuildsOneSetForBothPlatforms(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects() (+12 more)

### Community 42 - "InjectShutdown"
Cohesion: 0.17
Nodes (27): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+19 more)

### Community 43 - "ContainerSecrets"
Cohesion: 0.32
Nodes (7): ContainerSecrets(), SecretPreflight(), TestFileBackedSecretIsExemptFromSecretPreflight(), TestPodmanNeverGetsAFileBackedSecret(), TestServerCertBundlePathIsPosixAndUnderBaseDir(), TestServerCertFilePathKeyKeepsItsUnderscore(), TestServerCertIsADockerOnlySecret()

### Community 44 - "NewMateChannel"
Cohesion: 0.25
Nodes (21): execArgs(), NewMateChannel(), ReadSecretKey(), secretCommand(), hasPrefixArgv(), indexOf(), operandOf(), replCfg() (+13 more)

### Community 45 - "opRunner"
Cohesion: 0.11
Nodes (22): opCall, opRunner, os.File, capture(), captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect() (+14 more)

### Community 46 - "broker/replication.go"
Cohesion: 0.13
Nodes (25): colSpan, MateConfig, QueueState, cliTransport(), containsEndpoint(), dashSpans(), flagByte(), gutterClear() (+17 more)

### Community 47 - ".ConfigureReplication"
Cohesion: 0.13
Nodes (26): VPNRepl, Ops, isRunCLIRejection(), mateConvergenceShutdowns(), missingListedVPNs(), newlineIf(), reenableLine(), replicationVPNLines() (+18 more)

### Community 48 - "ReplSite"
Cohesion: 0.48
Nodes (5): Replication, PlannedRoles(), RoleAtSite(), ReplSite, ReplVPN

### Community 49 - "Ops"
Cohesion: 0.13
Nodes (13): cliRunNames(), countContains(), runCLISkeleton(), TestCountContains(), TestValidName(), validCLILine(), validName(), ValidVPNName() (+5 more)

### Community 50 - "testing.T"
Cohesion: 0.06
Nodes (69): testing.T, TestConfirmDeleteShortcut(), TestCtrManagerConfirmWiring(), TestPromptYes(), TestPromptYesNo(), TestExamplesAcceptsAPlatformAbbreviation(), TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout() (+61 more)

### Community 51 - "newRootCmd"
Cohesion: 0.08
Nodes (33): TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias(), TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRegisteredWhereItExecutes() (+25 more)

### Community 52 - "cli/replication_test.go"
Cohesion: 0.38
Nodes (13): mateChannel(), mateSEMPPassword(), App, replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer(), TestMateChannelPicksTheMechanismTheSiteDeclares() (+5 more)

### Community 53 - "omit_test.go"
Cohesion: 0.33
Nodes (5): TestImportPlanAcceptsWhatExportProduced(), TestImportPlanRefusesAFileItDidNotExport(), TestOmitAnnotateRoundTripsOmissions(), TestOmitAtExportDropsFilteredLinesAndEmptyBlocks(), TestOmitExportConfigEndToEnd()

### Community 54 - "Operations"
Cohesion: 0.10
Nodes (21): Bringing up a fresh cluster, Configuring a site, Data replication, Docker and Podman mechanics, Every destructive command confirms, Exit codes, Extra CLI users differ by platform, Operations (+13 more)

### Community 55 - "Test catalogue"
Cohesion: 0.10
Nodes (21): abbrev_test.go, convert_test.go, Coverage, examples_test.go, Fixtures and doubles, Injectable seams, internal/abbrev, internal/convert (+13 more)

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

### Community 63 - "replicationops_test.go"
Cohesion: 0.30
Nodes (16): hasCall(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), TestConfigureReplicationClosingShowFailureStillReportsTheApply(), TestConfigureReplicationPhase1BodyOrderIsShutdownsThenRemovalsThenMate(), TestConfigureReplicationPhase1RejectionStopsBeforePhase2() (+8 more)

### Community 65 - "Configuration"
Cohesion: 0.17
Nodes (12): Bring your own TLS Secret, Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling, Scaling tiers (+4 more)

### Community 67 - "broker/replication_test.go"
Cohesion: 0.15
Nodes (30): ParseShowReplication(), ParseVPNReplication(), RenderMate(), SameEndpoints(), loadShowReplication(), mateAppliance(), mateSoftware(), TestMateRoundTripsThroughItsOwnGrammar() (+22 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "internal/engine"
Cohesion: 0.67
Nodes (3): internal/engine, resolve_test.go, runner_test.go

### Community 71 - "localCfg"
Cohesion: 0.10
Nodes (42): sempPort(), assertNoPasswordInArgv(), TestLocalAdminIsTheDeploymentsOwnLogin(), TestMateRevertActivityBridgePlaintextOnly(), TestMateRevertActivityBridgeTLSNoCA(), TestMateRevertActivityCredsAndBodyNeverInArgv(), TestMateRevertActivityNon2xx(), TestMateRevertActivityRPCNotOK() (+34 more)

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

### Community 77 - ".Run"
Cohesion: 0.11
Nodes (20): interactiveFailRunner, Exec, os/exec.Cmd, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveMissingBinaryIsActionable() (+12 more)

### Community 78 - "redundancy_test.go"
Cohesion: 0.21
Nodes (12): Config, pskCfg(), TestFillStandaloneNodeNameIgnoresAnEmptyHostname(), TestFillStandaloneNodeNameKeepsAConfiguredName(), TestFillStandaloneNodeNameNeverFillsHA(), TestFillStandaloneNodeNameUsesTheHost(), TestNothingGeneratesThePSK(), TestPSKEnvSatisfiesTheContainerRequirement() (+4 more)

### Community 79 - "MateChannel"
Cohesion: 0.19
Nodes (19): PhaseKind, SwitchAction, SwitchPhase, MateChannel, confirmDemotions(), SwitchPlan, promoteConfirmed(), replPrecheckError() (+11 more)

### Community 80 - "scaling.go"
Cohesion: 0.16
Nodes (14): scalingKey, scalingSpelling, scalingTier, containerMem(), Config, Scaling, Scaling, yaml.Node (+6 more)

### Community 81 - "blocks.go"
Cohesion: 0.12
Nodes (25): Omission, Region, token, markerRegion(), brokerTypeFromSchema(), firstQuoted(), Capture, isIndented() (+17 more)

### Community 82 - "age"
Cohesion: 0.22
Nodes (9): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+1 more)

### Community 83 - "k8s/inspect_test.go"
Cohesion: 0.15
Nodes (19): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+11 more)

### Community 84 - "Command"
Cohesion: 0.12
Nodes (15): Command, yaml.Node, decodeStrict(), Config, guardCommandOf(), setGuardCommand(), TestAllowCommandIsNotASchemaKey(), TestAllowCommandsAccepts() (+7 more)

### Community 85 - "hostpath_test.go"
Cohesion: 0.10
Nodes (24): CanonicalDuration(), TestCanonicalDuration(), guardConfig(), CheckHostPath(), HasPathSeparator(), IsAbsHostPath(), Config, setContainerName() (+16 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.17
Nodes (17): time.Time, brokerConditionRows(), conditionLevel(), findCondition(), Cluster, normalizeToList(), TestConditionLevelDegradesSafely(), TestNormalizeToListHandlesBothKubectlShapes() (+9 more)

### Community 87 - ".preflightOne"
Cohesion: 0.44
Nodes (3): canIAnswer(), Cluster, probe

### Community 88 - "github.com/spf13/cobra.Command"
Cohesion: 0.05
Nodes (144): opFunc, roleOpFunc, Example, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, applyAliases(), addCommands(), addLogFlags() (+136 more)

### Community 89 - "load"
Cohesion: 0.14
Nodes (32): HealthCheck, BrokerCR(), Compose(), composeEscape(), ComposeProject(), healthCmd(), splitPair(), assertNoCheckoutPath() (+24 more)

### Community 90 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.10
Nodes (37): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+29 more)

### Community 93 - "cli_test.go"
Cohesion: 0.07
Nodes (87): testing.M, allowRuntime(), captureStderr(), echoRunner(), fakeBinaryOnPath(), firstLine(), opFailOn(), runCtr() (+79 more)

### Community 94 - "container/inspect_test.go"
Cohesion: 0.12
Nodes (21): containerState, healthState, decodeInspect(), Manager, orUnknown(), inspectReply(), kvValue(), TestInspectDecodesPartialOutput() (+13 more)

### Community 95 - "broker/matechannel_test.go"
Cohesion: 0.19
Nodes (19): CLIRunner, fakeRun, Ops, NewCLIMate(), mateReplies(), newTestMate(), TestLocalMateRunsThroughOps(), TestMateChannelDescribeNamesTheTarget() (+11 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 100 - "tlsCfg"
Cohesion: 0.43
Nodes (7): Config, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsMakeTheSecretOurs(), TestSuppliedCertsNeedASecretName(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), tlsCfg()

### Community 101 - "scripts.go"
Cohesion: 0.08
Nodes (38): showCmd, defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript(), enableDefaultUsersScript(), enableDefaultVPNScript(), gatherConfigsScript() (+30 more)

### Community 102 - "opPerformReplication"
Cohesion: 0.21
Nodes (18): mateChannelFunc, bufio.Reader, App, step(), confirmReplicationConfig(), confirmReplicationSwitch(), App, opConfigureReplication() (+10 more)

### Community 103 - "ctrCfg"
Cohesion: 0.12
Nodes (37): ctrCfg(), newEchoMgr(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployDockerDryRunMasksSecretEnv(), TestManagerDeployPodmanDryRunHidesSecretBytes(), TestManagerDeployPodmanDryRunSkipsWrite(), TestManagerDockerCheckProbesCompose() (+29 more)

### Community 104 - "watch_test.go"
Cohesion: 0.13
Nodes (21): desiredWatch(), Cluster, splitWatch(), subtractWatch(), deployJSON(), Cluster, TestDesiredWatchAppendsTheBrokerNamespace(), TestOperatorReleaseNarrowsInsteadOfRemoving() (+13 more)

### Community 105 - "internal/cli"
Cohesion: 0.17
Nodes (12): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, examples_test.go, exit_test.go (+4 more)

### Community 107 - ".operatorRows"
Cohesion: 0.21
Nodes (11): compareVersions(), Cluster, imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestImageTag() (+3 more)

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "config/replication_test.go"
Cohesion: 0.36
Nodes (11): Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationAbsent(), TestValidateReplicationEmptyViaAccepted() (+3 more)

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 112 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, File descriptors on rootless podman, Import reported failure, Podman secret flags, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 113 - "internal/config"
Cohesion: 0.15
Nodes (13): command_test.go, config_test.go, domaincerts_test.go, duration_test.go, execguard_test.go, hostpath_test.go, internal/config, platform_test.go (+5 more)

### Community 114 - "haCfg"
Cohesion: 0.14
Nodes (21): TestCheckStopsProbingWhenUnreachable(), TestResolveStorageClass(), TestDeleteBrokerProbesClaimsOnlyWhenPurging(), lbServiceName(), haCfg(), saCfg(), TestHARoles(), TestPodNameSuffixMatchesTheConfigBound() (+13 more)

### Community 117 - "internal/container"
Cohesion: 0.29
Nodes (7): inspect_test.go, internal/container, manager_test.go, preflight_test.go, runtime_test.go, secrets_test.go, transport_test.go

### Community 118 - "completion_test.go"
Cohesion: 0.35
Nodes (12): runComplete(), TestAllowCommandOffersNoFiles(), TestCompletionScriptsGenerate(), TestDirFlagCompletesDirectories(), TestEnvFlagCompletesEnvFiles(), TestEnvFlagPrefixFilters(), TestEnvFlagWithPathDefersToShell(), TestNoArgsLeafOffersNoFiles() (+4 more)

### Community 119 - "load.go"
Cohesion: 0.18
Nodes (16): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults() (+8 more)

### Community 120 - "blocks_test.go"
Cohesion: 0.26
Nodes (12): TestAnnotateIsDeterministic(), blockByIndex(), Capture, loadSample(), TestParseBlocksCRLFMatchesLF(), TestParseBlocksHeaderAndPreamble(), TestParseBlocksIndentedLineWithNoOpenerRefused(), TestParseBlocksMissingEndTerminatorRefused() (+4 more)

### Community 121 - "NewManager"
Cohesion: 0.26
Nodes (12): NewManager(), TestManagerNilSinks(), Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime() (+4 more)

### Community 122 - ".LeaderLocal"
Cohesion: 0.23
Nodes (7): assertLeaderScript(), TestAssertLeaderScript(), defaultLocalAddrs(), Ops, hostMatches(), shortHost(), TestDefaultLocalAddrs()

### Community 123 - "manager.go"
Cohesion: 0.19
Nodes (10): composeNeedsSecretValues(), exactName(), orNone(), platformTitle(), secretSummary(), setOrMissing(), solaceRows(), splitLimit() (+2 more)

### Community 124 - "render.go"
Cohesion: 0.14
Nodes (24): NodeIdentity, boolStr(), containerSecretSpecs(), cut(), EnvPairs(), escapePercent(), groupKey(), itoa() (+16 more)

### Community 129 - "BrokerType"
Cohesion: 0.38
Nodes (4): BrokerType, cliMate, bannerType(), showReplicationScript()

### Community 130 - "rootlessNoFileMgr"
Cohesion: 0.25
Nodes (8): bytes.Buffer, Manager, rootlessNoFileMgr(), setNoFile(), TestPrepHostRootfulSkipsNoFile(), TestPrepHostRootlessNoFileTooLow(), TestPrepHostRootlessNoFileUnlimited(), TestPrepHostRootlessNoFileUnreadable()

### Community 133 - "TestServerCert"
Cohesion: 0.26
Nodes (9): TestPathHelpers(), TestServerCert(), serverCertFile(), serverCertScript(), TestServerCertScript(), certPath(), CLIArg(), cliArg() (+1 more)

### Community 135 - "ParseRole"
Cohesion: 0.36
Nodes (7): TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), TestRoleNames(), ParseRole(), RoleAbbrev(), RoleNames()

## Knowledge Gaps
- **215 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+210 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 322 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `github.com/spf13/cobra.Command` to `bg`, `rootlessNoFileMgr`, `config_test.go`, `Manager`, `config.go`, `CheckCommand`, `capRunner`, `convert_test.go`, `newCapMgr`, `validate.go`, `ContainerSecrets`, `opRunner`, `localCfg`, `redundancy_test.go`, `scaling.go`, `Command`, `hostpath_test.go`, `load`, `opPerformReplication`, `ctrCfg`, `load.go`, `NewManager`, `manager.go`, `containerTransport`?**
  _High betweenness centrality (0.052) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `bg`, `newTestOps`, `rootlessNoFileMgr`, `strings.Builder`, `Manager`, `config.go`, `capRunner`, `NewCluster`, `convert_test.go`, `newCapMgr`, `k8s/secrets_test.go`, `prep_test.go`, `writeNode`, `Role`, `recRunner`, `k8s/runtime_test.go`, `ContainerSecrets`, `NewMateChannel`, `opRunner`, `ReplSite`, `checkreport.go`, `Cluster`, `kubectlTransport`, `localCfg`, `namespace_test.go`, `load`, `cli_test.go`, `opPerformReplication`, `ctrCfg`, `watch_test.go`, `haCfg`, `load.go`, `NewManager`, `render.go`, `containerTransport`?**
  _High betweenness centrality (0.043) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `BrokerType`, `Manager`, `config.go`, `ParseRole`, `importops.go`, `.releaseToBackup`, `renderDriver`, `validate.go`, `importops_test.go`, `.ConfigureReplication`, `Ops`, `.MateRevertActivity`, `kubectlTransport`, `Set`, `broker/matechannel_test.go`, `scripts.go`, `opPerformReplication`, `.LeaderLocal`, `containerTransport`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 34 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 34 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _215 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `bg` be split into smaller, more focused modules?**
  _Cohesion score 0.053841901603095633 - nodes in this community are weakly interconnected._
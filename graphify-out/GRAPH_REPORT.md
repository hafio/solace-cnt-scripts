# Graph Report - solace-cnt-scripts  (2026-10-01)

## Corpus Check
- 196 files · ~612,817 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 3822 nodes · 16043 edges · 142 communities (126 shown, 16 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 2214 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `99eb4623`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- newTestOps
- Sink
- context.Context
- cli_test.go
- testing.T
- Command Reference Docs
- Manager
- K8sConfig
- Quadlet
- CheckCommand
- haCfg
- NewCluster
- GenSecrets
- convert_test.go
- Block
- manager_test.go
- github.com/spf13/cobra.Command
- Admin Secret Rendering
- Bash Dev Script
- MateConfig
- PowerShell Dev Script
- Abbreviation Sets
- BuildSwitchPlan
- verify_local_test.go
- runner_test.go
- renderDriver
- eqArgs
- Import Plan and Apply
- newEchoMgr
- Solace Module Root
- Config
- Docs and Legacy Scripts
- ParseVPNReplication
- Domain Cert Directory Loading
- Role
- App
- recRunner
- annotate_test.go
- completion_test.go
- diff_test.go
- Platform Selection
- inject_test.go
- TestEveryScriptTurnsPagingOffAfterHome
- hasCall
- Echo
- k8s/inspect.go
- replicationVPNLines
- localCfg
- config_test.go
- ParseBlocks
- newRootCmd
- exportconfig_test.go
- Container Security Guidelines
- Operations Guide
- Test catalogue
- Cluster
- MateChannel
- Ops
- VPNRepl
- confirm.go
- .configRows
- rootless_test.go
- strings.Builder
- output_test.go
- Configuration Guide
- Ops
- Mate Channel Tests
- Command Reference Doc
- Developer Guide
- ReplSite
- RenderDiffResult
- age
- internal/k8s
- parseBody
- usagef
- .ConfigureReplication
- Command
- newTeardownApplyFixture
- Container File Transport
- Abbrev Package
- BrokerType
- k8s/matechannel_test.go
- limits_test.go
- occCluster
- Troubleshooting Guide
- Kubernetes Inspect Tests
- ParseShowReplication
- operatorversion_test.go
- render_test.go
- main_test.go
- Import Section Rules
- step
- Container
- Config
- Secret Reference Resolution
- ParseRole
- replConfig
- Abbreviation Reference
- newBlock
- .checkUserManagerLimits
- platformTitle
- Load
- container/secrets_test.go
- watchCluster
- internal/cli
- desiredWatch
- Compose
- Import-Config Apply Rules
- Config
- internal/broker
- GenOperator
- capRunner
- Config Package Tests
- TestServerCert
- imageFromDeployment
- EnvPairs
- internal/container
- Container State Inspection
- reportReplicationConfig
- New
- coverage_test.go
- .validateContainerArtifactValues
- Destructive Removal Confirmations
- .applyScalingTierDefaults
- importIgnore
- Scaling YAML Decoding
- .preflightOne
- ReplVia
- Namespace Protection Checks
- Env Path Resolution
- Duration and Host Path Checks
- Broker Package
- CLI Package
- Config Package
- Container Package
- Convert Package
- Engine Package
- Examples Package
- K8s Package
- Output Package
- Render Package

## God Nodes (most connected - your core abstractions)
1. `ctrCfg()` - 131 edges
2. `newTestOps()` - 130 edges
3. `Role` - 122 edges
4. `Config` - 113 edges
5. `newCapMgr()` - 111 edges
6. `NewCluster()` - 100 edges
7. `loadK8s()` - 89 edges
8. `matchCLI()` - 81 edges
9. `Platform` - 78 edges
10. `bg()` - 70 edges

## Surprising Connections (you probably didn't know these)
- `assertFencesBalanced()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/annotate.go
- `parseBody()` --calls--> `readMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `parseHeader()` --calls--> `readMarker()`  [INFERRED]
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

## Communities (142 total, 16 thin omitted)

### Community 0 - "bg"
Cohesion: 0.13
Nodes (48): warn(), bg(), App, k8sAdminPassword(), k8sCluster(), k8sContext(), k8sLogin(), k8sOps() (+40 more)

### Community 1 - "newTestOps"
Cohesion: 0.08
Nodes (51): Ops, newTestOps(), outputForRole(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultUsersRefusesAQuotedVPNName(), TestDisableDefaultVPN(), TestDomainCerts() (+43 more)

### Community 2 - "Sink"
Cohesion: 0.20
Nodes (5): solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), columnWidths(), Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.05
Nodes (18): Exec, context.Context, os/exec.Cmd, Manager, idMapCovers(), origin(), runUserIDs(), EnvRunner (+10 more)

### Community 4 - "cli_test.go"
Cohesion: 0.05
Nodes (134): opCall, os.File, testing.M, allowRuntime(), capture(), captureStderr(), captureStdout(), echoRunner() (+126 more)

### Community 5 - "testing.T"
Cohesion: 0.01
Nodes (215): testing.T, TestRouterNameReadsTheCaptureHeader(), bridgeHostPort(), sempPort(), sempV1OK(), TestSempPortBridgeFallsBackToPlaintext(), TestSempPortBridgePrefersTLS(), TestSempPortBridgeRefusesNoMapping() (+207 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (7): certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), secretSummary(), setOrMissing()

### Community 8 - "K8sConfig"
Cohesion: 0.10
Nodes (15): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, LoadBalancer, cut() (+7 more)

### Community 9 - "Quadlet"
Cohesion: 0.20
Nodes (14): ContainerNoFile(), HealthCheck, TestLimitsCheckAssertsWhatTheArtifactAsks(), escapePercent(), healthCmd(), Quadlet(), quadletEscape(), healthCheckFixture() (+6 more)

### Community 10 - "CheckCommand"
Cohesion: 0.15
Nodes (22): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+14 more)

### Community 11 - "haCfg"
Cohesion: 0.12
Nodes (22): AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestReadAdminPasswordArgvAndDecoding(), TestReadAdminPasswordGuardsTheCommand(), TestReadAdminPasswordNamesTheNextStep() (+14 more)

### Community 12 - "NewCluster"
Cohesion: 0.08
Nodes (59): TestCheckStopsProbingWhenUnreachable(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+51 more)

### Community 13 - "GenSecrets"
Cohesion: 0.07
Nodes (39): GenBroker(), GenSecrets(), joinManifests(), namespaceManifest(), adminCfg(), secretNamesIn(), TestAdminSecretStatesEmitTheirArtifacts(), TestCreateSecretsAdminOnly() (+31 more)

### Community 14 - "convert_test.go"
Cohesion: 0.05
Nodes (68): doc, Result, segment, vars, DecodeStrict(), Config, boolOf(), commentSafe() (+60 more)

### Community 15 - "Block"
Cohesion: 0.06
Nodes (52): Block, Region, TargetState, regexp.Regexp, regionFor(), keepRegion(), injectedBlock(), ClearExistingNested() (+44 more)

### Community 16 - "manager_test.go"
Cohesion: 0.07
Nodes (94): inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer() (+86 more)

### Community 17 - "github.com/spf13/cobra.Command"
Cohesion: 0.06
Nodes (127): opFunc, roleOpFunc, shorthand, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, TestFlagShorthandsAreConsistent(), treeShorthands(), writeUnabbreviated() (+119 more)

### Community 18 - "Admin Secret Rendering"
Cohesion: 0.18
Nodes (13): AdditionalUsersSecret(), AdminSecret(), decodeDataValue(), TestAdditionalUsersSecretCarriesBothHalves(), TestAdditionalUsersSecretIsAbsentWithNoUsers(), TestAdditionalUsersSecretRefusesAnEmptyPassword(), TestAdditionalUsersStayOutOfTheCredentialsSecret(), TestAdminSecretCarriesThePSKOnlyWhenSet() (+5 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "MateConfig"
Cohesion: 0.13
Nodes (18): MateConfig, cliTransport(), containsEndpoint(), labelValue(), normTransport(), parseHostPort(), parseShowReplicationAppliance(), parseShowReplicationSoftware() (+10 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Abbreviation Sets"
Cohesion: 0.13
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 23 - "BuildSwitchPlan"
Cohesion: 0.12
Nodes (30): PhaseKind, SiteState, SwitchAction, SwitchPhase, BuildSwitchPlan(), confirmDemotions(), ExecuteSwitchPlan(), SwitchPlan (+22 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.13
Nodes (43): cliScriptPath(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn(), TestDetectRoleAddrsError() (+35 more)

### Community 25 - "runner_test.go"
Cohesion: 0.11
Nodes (28): TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), MaskEnv(), quoteTok(), captureStdout(), helperCommand() (+20 more)

### Community 26 - "renderDriver"
Cohesion: 0.09
Nodes (26): chunk, chunkResult, runCLISkeleton(), chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment() (+18 more)

### Community 27 - "eqArgs"
Cohesion: 0.10
Nodes (33): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), saCfg() (+25 more)

### Community 28 - "Import Plan and Apply"
Cohesion: 0.14
Nodes (20): ImportResult, PlannedSection, isPragma(), defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList() (+12 more)

### Community 29 - "newEchoMgr"
Cohesion: 0.10
Nodes (20): bytes.Buffer, NewManager(), Manager, newEchoMgr(), TestCheckEnvReportsBaseDirOnlyWhenSet(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployPodmanDryRunHidesSecretBytes() (+12 more)

### Community 31 - "Config"
Cohesion: 0.09
Nodes (18): keyValueEntries, TestValidateContainerArtifactValues(), TestRoleNames(), RoleNames(), checkCredentialChars(), foldToEnvVar(), Config, invertedCPUSetRange() (+10 more)

### Community 32 - "Docs and Legacy Scripts"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "ParseVPNReplication"
Cohesion: 0.10
Nodes (22): colSpan, ValidVPNName(), validVPNName(), dashSpans(), flagByte(), gutterClear(), ParseVPNReplication(), sliceSpan() (+14 more)

### Community 34 - "Domain Cert Directory Loading"
Cohesion: 0.06
Nodes (44): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, os.DirEntry, os.FileInfo, os.FileMode (+36 more)

### Community 35 - "Role"
Cohesion: 0.06
Nodes (27): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+19 more)

### Community 36 - "App"
Cohesion: 0.10
Nodes (44): confirmAction(), childExit(), TestChildExitKeepsItsMessage(), wantEnable(), wantRemove(), containerRenderRole(), containerRole(), containerWhat() (+36 more)

### Community 37 - "recRunner"
Cohesion: 0.14
Nodes (10): TestCanIAnswerReadsTheLastLine(), TestExecutorRefusesUnapprovedRuntime(), unapprovedCfg(), NewTransport(), TestTransportCopy(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest() (+2 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.07
Nodes (48): Annotate(), applyMetaField(), Capture, markerRegion(), markerSection(), markerVerb(), parseKV(), readMarker() (+40 more)

### Community 39 - "completion_test.go"
Cohesion: 0.17
Nodes (21): driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion(), TestCompletionDescriptionsAreOptIn() (+13 more)

### Community 40 - "diff_test.go"
Cohesion: 0.15
Nodes (25): blockKey, mergedBlock, mergedLine, ancestorReported(), DiffBlocks(), indexByIdentity(), keyOf(), keySet() (+17 more)

### Community 41 - "Platform Selection"
Cohesion: 0.29
Nodes (12): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+4 more)

### Community 42 - "inject_test.go"
Cohesion: 0.15
Nodes (27): lineRole, serviceLine, shutdownStyle, svcKey, classifyServiceRest(), containsPortCommand(), describeKey(), fieldsUnquoted() (+19 more)

### Community 43 - "TestEveryScriptTurnsPagingOffAfterHome"
Cohesion: 0.07
Nodes (39): countContains(), field(), TestCountContains(), TestField(), TestLastLines(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), TestLastLinesEqualCount() (+31 more)

### Community 44 - "hasCall"
Cohesion: 0.20
Nodes (17): cliScriptNameFromDest(), hasCall(), TestDisableDefaultUsersShowVPNError(), TestReleaseToBackupReleasedTimeout(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching() (+9 more)

### Community 45 - "Echo"
Cohesion: 0.20
Nodes (7): interactiveFailRunner, runFailRunner, TestExecIsSilentWithoutVerbose(), Echo, NewExec(), Quote(), TestQuote()

### Community 46 - "k8s/inspect.go"
Cohesion: 0.16
Nodes (21): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), ownedPods(), TestWatchFromContainersReadsEveryAllNamespacesSpelling(), watchFromContainers() (+13 more)

### Community 47 - "replicationVPNLines"
Cohesion: 0.19
Nodes (16): sortedKeys(), mateConvergenceShutdowns(), replicationVPNLines(), sortedVPNs(), siteAEntry(), TestMateConvergenceShutdownsIsDeterministic(), TestNothingEverShutsDownTheVPNItself(), TestPlannedRolesNamesEveryListedVPN() (+8 more)

### Community 48 - "localCfg"
Cohesion: 0.09
Nodes (36): TestLocalMateRunsThroughOps(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK(), TestBackupRevertActivitySuccess() (+28 more)

### Community 49 - "config_test.go"
Cohesion: 0.19
Nodes (48): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, showCmd, span, go_pkg_bytes, go_pkg_context (+40 more)

### Community 50 - "ParseBlocks"
Cohesion: 0.13
Nodes (17): StripMarkers(), TestStripMarkersKeepsPragmaAndSectionComments(), TestStripMarkersNoMarkersPassthrough(), ParseBlocks(), splitLines(), TestParseBlocksCRLFMatchesLF(), TestParseBlocksIndentedLineWithNoOpenerRefused(), TestParseBlocksMissingEndTerminatorRefused() (+9 more)

### Community 51 - "newRootCmd"
Cohesion: 0.07
Nodes (38): TestAbbreviationDocs(), applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias() (+30 more)

### Community 52 - "exportconfig_test.go"
Cohesion: 0.08
Nodes (39): opRunner, go_pkg_runtime, runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand(), TestPodmanEUIDGuardSkips(), writeRootlessPodmanEnv(), ExitCode() (+31 more)

### Community 53 - "Container Security Guidelines"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 54 - "Operations Guide"
Cohesion: 0.08
Nodes (26): Bringing up a fresh cluster, Broker scope is a fixed classification, Configuring a site, Data replication, Docker and Podman mechanics, Exit codes, Exporting and importing configuration, Extra CLI users differ by platform (+18 more)

### Community 55 - "Test catalogue"
Cohesion: 0.08
Nodes (24): abbrev_test.go, convert_test.go, Coverage, examples_test.go, Fixtures and doubles, Injectable seams, internal/abbrev, internal/convert (+16 more)

### Community 56 - "Cluster"
Cohesion: 0.13
Nodes (5): time.Time, Cluster, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

### Community 57 - "MateChannel"
Cohesion: 0.19
Nodes (14): MateChannel, curlConfigFlag(), Ops, requirePrimaryActive(), requireVirtualRouterNamesMatch(), sortedSiteNames(), SwitchPreflight(), replSites() (+6 more)

### Community 58 - "Ops"
Cohesion: 0.07
Nodes (22): TestValidName(), validCLILine(), validName(), Ops, disableDefaultVPNScript(), domainCertsScript(), enableDefaultVPNScript(), productKeyScript() (+14 more)

### Community 59 - "VPNRepl"
Cohesion: 0.06
Nodes (29): AdminState, backupTarget, Credential, QueueState, ReplRole, scriptedMate, sempMate, VPNRepl (+21 more)

### Community 60 - "confirm.go"
Cohesion: 0.22
Nodes (20): exportconfigReadCounter, layer, go_pkg_bufio, io.Reader, io.Writer, TestStdinCanAnswerClosedFile(), confirmActionStrict(), confirmDelete() (+12 more)

### Community 61 - ".configRows"
Cohesion: 0.12
Nodes (26): orNone(), orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminPassState(), adminSecretRow() (+18 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.15
Nodes (39): failOnCall(), fakeEnv(), Manager, healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+31 more)

### Community 63 - "strings.Builder"
Cohesion: 0.11
Nodes (27): WeightedNodeTerm, github.com/spf13/pflag.FlagSet, strings.Builder, mdRow(), renderAbbrevDocs(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable() (+19 more)

### Community 64 - "output_test.go"
Cohesion: 0.19
Nodes (17): NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLevelOrderKeepsSkipBelowFail(), TestLineAddsNoPrefix() (+9 more)

### Community 65 - "Configuration Guide"
Cohesion: 0.14
Nodes (14): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+6 more)

### Community 67 - "Mate Channel Tests"
Cohesion: 0.16
Nodes (17): CLIRunner, fakeRun, Ops, NewCLIMate(), mateReplies(), newTestMate(), TestMateChannelDescribeNamesTheTarget(), TestMateChannelPreflightLearnsTheBrokerType() (+9 more)

### Community 68 - "Command Reference Doc"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer Guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "ReplSite"
Cohesion: 0.24
Nodes (18): PlannedRoles(), RoleAtSite(), TestRoleAtSiteIsTheComplement(), confirmReplicationConfig(), mateChannel(), mateSEMPPassword(), App, replApp() (+10 more)

### Community 71 - "RenderDiffResult"
Cohesion: 0.23
Nodes (11): BlockDiff, DiffResult, blockLabel(), diffRow(), orDash(), RenderDiffResult(), reportBlockLines(), TestDiffBlockLabelFormatsIdentityForTheReport() (+3 more)

### Community 72 - "age"
Cohesion: 0.26
Nodes (7): age(), roleRank(), Cluster, operatorRunningImage(), ownsPod(), TestOperatorRunningImageWithNoContainers(), TestOwnsPodFallsBackToTheNameInfix()

### Community 73 - "internal/k8s"
Cohesion: 0.10
Nodes (20): adminsecret_test.go, check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go (+12 more)

### Community 74 - "parseBody"
Cohesion: 0.15
Nodes (13): Omission, isMarker(), TestIsMarkerDiscriminatesNamespace(), brokerTypeFromSchema(), Capture, isIndented(), parseBody(), parseHeader() (+5 more)

### Community 75 - "usagef"
Cohesion: 0.12
Nodes (19): childExitError, usageError, github.com/spf13/cobra.PositionalArgs, logArgs(), App, newConvertCmd(), runConvert(), asUsage() (+11 more)

### Community 76 - ".ConfigureReplication"
Cohesion: 0.24
Nodes (8): Ops, isRunCLIRejection(), newlineIf(), replPhase1Rejected(), replPhase2Rejected(), replTransportFailure(), replVPNList(), showVPNReplicationScript()

### Community 77 - "Command"
Cohesion: 0.13
Nodes (14): cmdField, guardedCmd, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandUnmarshal(), TestValidateProbeCommandAccepts(), TestValidateProbeCommandRejects(), Command (+6 more)

### Community 78 - "newTeardownApplyFixture"
Cohesion: 0.14
Nodes (17): driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, minimalCapture(), newTeardownApplyFixture(), targetVPNCapture() (+9 more)

### Community 81 - "BrokerType"
Cohesion: 0.33
Nodes (5): BrokerType, cliMate, bannerType(), showReplicationScript(), TestTranscriptBannerType()

### Community 82 - "k8s/matechannel_test.go"
Cohesion: 0.21
Nodes (26): CLIArg(), CLIScriptPath(), execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf() (+18 more)

### Community 83 - "limits_test.go"
Cohesion: 0.12
Nodes (29): go_pkg_slices, missingControllers(), Manager, limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessCoreFloorFollowsTheConfiguredLimit() (+21 more)

### Community 84 - "occCluster"
Cohesion: 0.18
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 85 - "Troubleshooting Guide"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 86 - "Kubernetes Inspect Tests"
Cohesion: 0.15
Nodes (19): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+11 more)

### Community 87 - "ParseShowReplication"
Cohesion: 0.16
Nodes (23): TestMateChannelShowReplication(), ParseShowReplication(), RenderMate(), SameEndpoints(), loadShowReplication(), mateAppliance(), mateSoftware(), TestMateRoundTripsThroughItsOwnGrammar() (+15 more)

### Community 88 - "operatorversion_test.go"
Cohesion: 0.19
Nodes (14): compareVersions(), imageTag(), operatorVersionWarning(), parseVersion(), operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsSilentInAPreview() (+6 more)

### Community 89 - "render_test.go"
Cohesion: 0.13
Nodes (35): boolStr(), BrokerCR(), containerArtifacts(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets(), TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPrivilegePosture() (+27 more)

### Community 90 - "main_test.go"
Cohesion: 0.17
Nodes (18): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+10 more)

### Community 91 - "Import Section Rules"
Cohesion: 0.09
Nodes (34): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+26 more)

### Community 92 - "step"
Cohesion: 0.15
Nodes (19): mateChannelFunc, bufio.Reader, App, lineSink(), progress(), step(), confirmReplicationSwitch(), App (+11 more)

### Community 93 - "Container"
Cohesion: 0.13
Nodes (14): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestDefaultK8sPortsMatchesOperator(), applyContainerBlockDefaults(), defaultK8sPorts() (+6 more)

### Community 94 - "Config"
Cohesion: 0.20
Nodes (6): expandTilde(), expandTildeToken(), Config, isPathSep(), TestExpandTilde(), tildeHome()

### Community 95 - "Secret Reference Resolution"
Cohesion: 0.50
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 96 - "ParseRole"
Cohesion: 0.50
Nodes (5): TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), ParseRole(), RoleAbbrev()

### Community 97 - "replConfig"
Cohesion: 0.23
Nodes (11): Replication, Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationEmptyViaAccepted() (+3 more)

### Community 98 - "Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "newBlock"
Cohesion: 0.28
Nodes (9): token, firstQuoted(), newBlock(), TestNewBlockMessageSpoolQualifierHasNoOwnName(), TestNewBlockOpener(), TestNewBlockVPNNameWithSpaces(), TestNewBlockVPNQualifierWithSpaces(), tokenizeOpener() (+1 more)

### Community 100 - ".checkUserManagerLimits"
Cohesion: 0.29
Nodes (6): delegateSetting(), Manager, limitText(), parseLimit(), parseUnitProps(), TestParseLimit()

### Community 102 - "Load"
Cohesion: 0.07
Nodes (53): Example, go_pkg_embed, TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), TestExamplesWritesOutFile(), minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint() (+45 more)

### Community 103 - "container/secrets_test.go"
Cohesion: 0.17
Nodes (26): ResolveSecretValues(), bundleHash(), certCreate(), certFixture(), TestCertSecretLabelDrivesTheRestart(), TestCertSecretRemovalFailureIsFatal(), TestContainerBundleCarriesTheChain(), TestDeletePodmanToleratesAMissingBundle() (+18 more)

### Community 104 - "watchCluster"
Cohesion: 0.29
Nodes (7): deployJSON(), Cluster, TestOperatorReleaseNarrowsInsteadOfRemoving(), TestReconcileWatchDecidesWhatToApply(), TestReconcileWatchFlagsAWideningAsAQuestion(), TestSetWatchRefusesAnEmptyList(), watchCluster()

### Community 105 - "internal/cli"
Cohesion: 0.15
Nodes (13): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, euid_test.go, examples_test.go (+5 more)

### Community 106 - "desiredWatch"
Cohesion: 0.31
Nodes (5): desiredWatch(), Cluster, splitWatch(), TestDesiredWatchAppendsTheBrokerNamespace(), watchPlan

### Community 107 - "Compose"
Cohesion: 0.14
Nodes (15): Compose(), composeEscape(), ComposeProject(), ContainerSecrets(), ContainerSecret, secretFilePath(), SecretPreflight(), TestBridgePublishesOnlyTheListedPorts() (+7 more)

### Community 108 - "Import-Config Apply Rules"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "Config"
Cohesion: 0.07
Nodes (20): AdditionalUser, Image, Node, Redundancy, SEMP, TLS, New(), TestNewDefaults() (+12 more)

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "GenOperator"
Cohesion: 0.14
Nodes (22): TestValidateSparseConfigExplainsItself(), GenOperator(), joinYAMLDocs(), nonEmpty(), operatorProbes(), RenderOperator(), renderOperatorNS(), splitAfterNamespace() (+14 more)

### Community 112 - "capRunner"
Cohesion: 0.13
Nodes (19): capCall, capRunner, callIndex(), TestManagerLogsCLIShell(), TestManagerPrepHostRootlessUsesUnshareChown(), TestPreflightRunsBeforeAnything(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime() (+11 more)

### Community 113 - "Config Package Tests"
Cohesion: 0.12
Nodes (17): adminsecret_test.go, artifactvalues_test.go, command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go (+9 more)

### Community 114 - "TestServerCert"
Cohesion: 0.29
Nodes (8): TestDiagnostics(), TestPathHelpers(), TestServerCert(), serverCertFile(), serverCertScript(), TestServerCertScript(), certPath(), cliArg()

### Community 115 - "imageFromDeployment"
Cohesion: 0.50
Nodes (4): imageFromDeployment(), operatorItem(), TestAmbiguousOperatorIsNotFoldedIntoNotInstalled(), TestFindOperatorDeploymentIsScopedByNamespace()

### Community 116 - "EnvPairs"
Cohesion: 0.12
Nodes (16): Config, NodeIdentity, containerSecretSpecs(), EnvPairs(), groupKey(), itoa(), ServerCertBundlePath(), assertNoCheckoutPath() (+8 more)

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.16
Nodes (14): containerState, healthState, decodeInspect(), Manager, orUnknown(), printable(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown() (+6 more)

### Community 119 - "reportReplicationConfig"
Cohesion: 0.67
Nodes (3): reportReplicationConfig(), TestReportReplicationConfigMateAppliedNamesWhatPhase1Stopped(), TestReportReplicationConfigMateSkippedSaysNothingWasWritten()

### Community 121 - "New"
Cohesion: 0.13
Nodes (17): TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), confirmImport(), App, pluralVPN(), runExport(), runImport() (+9 more)

### Community 122 - "coverage_test.go"
Cohesion: 0.05
Nodes (56): cliRunNames(), matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError() (+48 more)

### Community 123 - ".validateContainerArtifactValues"
Cohesion: 0.20
Nodes (7): Config, portOrRange(), TestValidContainerPort(), validContainerPort(), validCoreLimit(), validHealthDuration(), validDNSLabel()

### Community 125 - "Destructive Removal Confirmations"
Cohesion: 0.40
Nodes (5): Every destructive command confirms, Removing a broker: what stays, what goes, Removing the operator does not always remove it, The layer flag raises the question, The namespace is only offered when it is empty

### Community 126 - ".applyScalingTierDefaults"
Cohesion: 0.17
Nodes (11): scalingTier, containerMem(), cpuSetCount(), cpuSetRange(), Config, TestContainerMem(), TestCPUSetCount(), TestCPUSetRange() (+3 more)

### Community 128 - "importIgnore"
Cohesion: 0.47
Nodes (5): importIgnore(), TestImportIgnoreKeepsUnclassifiedSectionsInTheDiff(), TestImportOpsImportIgnore(), BridgeEnablementInverted(), TestShutdownBridgesLinesAreExcusedByTheDiff()

### Community 129 - "Scaling YAML Decoding"
Cohesion: 0.31
Nodes (7): scalingKey, scalingSpelling, Scaling, Scaling, yaml.Node, scalingKeyIndex(), scalingKeyList()

### Community 131 - ".preflightOne"
Cohesion: 0.50
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "ReplVia"
Cohesion: 0.33
Nodes (6): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, validateReplVia(), validateReplViaSEMP()

### Community 139 - "Env Path Resolution"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 141 - "Duration and Host Path Checks"
Cohesion: 0.40
Nodes (5): CanonicalDuration(), TestCanonicalDuration(), CheckHostPath(), TestCheckHostPathAccepts(), TestCheckHostPathRejects()

## Knowledge Gaps
- **237 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+232 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 357 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **16 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Manager` connect `Manager` to `context.Context`, `App`, `Config`, `config_test.go`, `github.com/spf13/cobra.Command`, `confirm.go`, `newEchoMgr`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `newTestOps`, `context.Context`, `cli_test.go`, `testing.T`, `Manager`, `K8sConfig`, `Quadlet`, `haCfg`, `NewCluster`, `GenSecrets`, `convert_test.go`, `manager_test.go`, `github.com/spf13/cobra.Command`, `Admin Secret Rendering`, `eqArgs`, `newEchoMgr`, `Domain Cert Directory Loading`, `Role`, `recRunner`, `localCfg`, `config_test.go`, `Cluster`, `.configRows`, `strings.Builder`, `Ops`, `Command`, `Container File Transport`, `k8s/matechannel_test.go`, `occCluster`, `render_test.go`, `step`, `Container`, `replConfig`, `container/secrets_test.go`, `desiredWatch`, `Compose`, `GenOperator`, `capRunner`, `EnvPairs`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `newTestOps`, `Manager`, `K8sConfig`, `Abbreviation Sets`, `renderDriver`, `Import Plan and Apply`, `Config`, `App`, `TestEveryScriptTurnsPagingOffAfterHome`, `config_test.go`, `ParseBlocks`, `Ops`, `VPNRepl`, `Ops`, `Mate Channel Tests`, `.ConfigureReplication`, `Container File Transport`, `BrokerType`, `k8s/matechannel_test.go`, `step`, `ParseRole`, `Config`, `EnvPairs`, `New`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Are the 49 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 49 INFERRED edges - model-reasoned connections that need verification._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _237 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `bg` be split into smaller, more focused modules?**
  _Cohesion score 0.13120567375886524 - nodes in this community are weakly interconnected._
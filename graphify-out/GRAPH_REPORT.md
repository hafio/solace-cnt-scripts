# Graph Report - solace-cnt-scripts  (2026-09-27)

## Corpus Check
- 194 files · ~592,236 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 3758 nodes · 15589 edges · 153 communities (135 shown, 18 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 2148 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `464a9626`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ops_k8s.go
- cli_test.go
- Sink
- context.Context
- runRoot
- testing.T
- Commands
- Manager
- K8sConfig
- GenOperator
- CheckCommand
- ParseBlocks
- NewCluster
- eqArgs
- convert_test.go
- transform.go
- manager_test.go
- strings.Builder
- AdminSecret
- dev.sh
- newTestOps
- dev.ps1
- Set
- BuildSwitchPlan
- verify_local_test.go
- runner_test.go
- renderDriver
- VPNRepl
- ImportPlan
- age
- solace
- Config
- CLAUDE.md
- newTeardownApplyFixture
- ResolveDomainCerts
- Role
- bg
- recRunner
- haCfg
- ReplSite
- diff_test.go
- runPlatform
- inject_test.go
- .releaseToBackup
- github.com/spf13/cobra.Command
- Echo
- k8s/inspect.go
- k8s/matechannel_test.go
- runRootWith
- config_test.go
- Command
- newRootCmd
- exportconfig_test.go
- The eight rules
- Operations
- Test catalogue
- .LeaderLocal
- RenderDiffResult
- scripts.go
- .rpc
- confirmAction
- .configRows
- rootless_test.go
- Cluster
- output_test.go
- Configuration
- Load
- MateConfig
- Command reference
- Developer guide
- loadSample
- localCfg
- .Run
- internal/k8s
- annotate_test.go
- usagef
- occCluster
- scripts_test.go
- Ops
- Quadlet
- go_pkg_solace_internal_abbrev
- eqArgs
- TestServerCert
- limits_test.go
- HARoles
- Troubleshooting
- k8s/inspect_test.go
- containerTransport
- replApp
- render_test.go
- main_test.go
- RuleFor
- step
- .applyContainerDefaults
- IsAbsHostPath
- .resolveSecretRefs
- Exec
- replConfig
- Abbreviations
- transform_test.go
- .checkUserManagerLimits
- BrokerType
- newTestMate
- certFixture
- Compose
- internal/cli
- MateChannel
- hasCall
- What `import-config` applies
- Config
- internal/broker
- ParseVPNReplication
- capRunner
- internal/config
- Removing a broker: what stays, what goes
- .Preflight
- Manager
- internal/container
- .stateRows
- internal/engine
- .validateContainer
- Ops
- coverage_test.go
- .ConfigureReplication
- ReplRole
- parseServiceLine
- .applyScalingTierDefaults
- runExport
- operatorversion.go
- .decodeScalingEntry
- newBlock
- .preflightOne
- ReplVia
- Cluster
- parsePort
- sempPort
- ContainerSecrets
- GenSecrets
- importIgnore
- ResolveEnvPath
- ParseRole
- CheckHostPath
- HasPathSeparator
- go_pkg_solace_internal_broker
- go_pkg_solace_internal_cli
- go_pkg_solace_internal_config
- go_pkg_solace_internal_container
- go_pkg_solace_internal_convert
- go_pkg_solace_internal_engine
- go_pkg_solace_internal_examples
- go_pkg_solace_internal_k8s
- go_pkg_solace_internal_output
- go_pkg_solace_internal_render

## God Nodes (most connected - your core abstractions)
1. `newTestOps()` - 130 edges
2. `ctrCfg()` - 124 edges
3. `Role` - 122 edges
4. `Config` - 108 edges
5. `newCapMgr()` - 107 edges
6. `NewCluster()` - 98 edges
7. `loadK8s()` - 89 edges
8. `matchCLI()` - 81 edges
9. `Platform` - 76 edges
10. `bg()` - 69 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `ExitCode()`  [EXTRACTED]
  main.go → internal/cli/exit.go
- `applyMetaField()` --calls--> `BrokerType`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go
- `applyMetaField()` --calls--> `brokerTypeFromSchema()`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go
- `TestAnnotateRegionAndSectionFencesAreBalanced()` --calls--> `Annotate()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/annotate.go
- `sectionBeginMarker()` --calls--> `RuleFor()`  [INFERRED]
  internal/broker/annotate.go → internal/broker/sections.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (153 total, 18 thin omitted)

### Community 0 - "ops_k8s.go"
Cohesion: 0.13
Nodes (48): domainCANames(), App, k8sAdminPassword(), k8sCluster(), k8sContext(), k8sLogin(), k8sOps(), k8sWhat() (+40 more)

### Community 1 - "cli_test.go"
Cohesion: 0.08
Nodes (57): opCall, os.File, allowRuntime(), capture(), captureStderr(), captureStdout(), collectPaths(), failDisableDefaultUsersUpload() (+49 more)

### Community 2 - "Sink"
Cohesion: 0.17
Nodes (5): solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), columnWidths(), Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.06
Nodes (15): scriptedMate, context.Context, Transport, EnvRunner, Runner, Cluster, readSecretKey(), Cluster (+7 more)

### Community 4 - "runRoot"
Cohesion: 0.08
Nodes (43): firstLine(), runRoot(), TestBashEnvGivenToEnvFlag(), TestConfigureServerCertsRefusesASecretItDoesNotOwn(), TestConvertErrorPaths(), TestConvertParseError(), TestConvertRoundTrip(), TestConvertToFile() (+35 more)

### Community 5 - "testing.T"
Cohesion: 0.02
Nodes (191): testing.T, TestRouterNameReadsTheCaptureHeader(), TestHelperExitProcess(), adminCfg(), Config, TestAdminSecretNameFollowsTheStates(), TestAdminSecretRefusesTheOperatorsOwnNames(), TestContainersStillRequireTheAdminPassword() (+183 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (7): composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing()

### Community 8 - "K8sConfig"
Cohesion: 0.15
Nodes (8): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, boolStr(), writeSecurity()

### Community 9 - "GenOperator"
Cohesion: 0.13
Nodes (24): TestValidateSparseConfigExplainsItself(), GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator(), renderOperatorNS() (+16 more)

### Community 10 - "CheckCommand"
Cohesion: 0.16
Nodes (21): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+13 more)

### Community 11 - "ParseBlocks"
Cohesion: 0.13
Nodes (17): StripMarkers(), TestStripMarkersKeepsPragmaAndSectionComments(), TestStripMarkersNoMarkersPassthrough(), ParseBlocks(), splitLines(), TestParseBlocksCRLFMatchesLF(), TestParseBlocksIndentedLineWithNoOpenerRefused(), TestParseBlocksMissingEndTerminatorRefused() (+9 more)

### Community 12 - "NewCluster"
Cohesion: 0.08
Nodes (60): TestCheckStopsProbingWhenUnreachable(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+52 more)

### Community 13 - "eqArgs"
Cohesion: 0.12
Nodes (28): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), TestCLIAndShellAreInteractive() (+20 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (64): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+56 more)

### Community 15 - "transform.go"
Cohesion: 0.12
Nodes (34): Block, Region, TargetState, regexp.Regexp, regionFor(), keepRegion(), ClearExistingNested(), clearNestedObjects() (+26 more)

### Community 16 - "manager_test.go"
Cohesion: 0.06
Nodes (112): inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer() (+104 more)

### Community 17 - "strings.Builder"
Cohesion: 0.09
Nodes (36): shorthand, WeightedNodeTerm, github.com/spf13/pflag.FlagSet, strings.Builder, mdRow(), renderAbbrevDocs(), TestFlagShorthandsAreConsistent(), treeShorthands() (+28 more)

### Community 18 - "AdminSecret"
Cohesion: 0.18
Nodes (13): AdditionalUsersSecret(), AdminSecret(), decodeDataValue(), TestAdditionalUsersSecretCarriesBothHalves(), TestAdditionalUsersSecretIsAbsentWithNoUsers(), TestAdditionalUsersSecretRefusesAnEmptyPassword(), TestAdditionalUsersStayOutOfTheCredentialsSecret(), TestAdminSecretCarriesThePSKOnlyWhenSet() (+5 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "newTestOps"
Cohesion: 0.08
Nodes (52): Ops, newTestOps(), outputForRole(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultUsersRefusesAQuotedVPNName(), TestDisableDefaultVPN(), TestDomainCerts() (+44 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Set"
Cohesion: 0.13
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 23 - "BuildSwitchPlan"
Cohesion: 0.12
Nodes (29): PhaseKind, SiteState, SwitchAction, SwitchPhase, BuildSwitchPlan(), confirmDemotions(), ExecuteSwitchPlan(), SwitchPlan (+21 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.13
Nodes (43): cliScriptPath(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn(), TestDetectRoleAddrsError() (+35 more)

### Community 25 - "runner_test.go"
Cohesion: 0.07
Nodes (41): bytes.Buffer, TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), MaskEnv() (+33 more)

### Community 26 - "renderDriver"
Cohesion: 0.10
Nodes (23): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), TestKeywordPatternEscapesEachPhrase() (+15 more)

### Community 27 - "VPNRepl"
Cohesion: 0.16
Nodes (15): QueueState, VPNRepl, mateConvergenceShutdowns(), reenableLine(), replicationVPNLines(), siteAEntry(), TestMateConvergenceShutdownsIsDeterministic(), TestNothingEverShutsDownTheVPNItself() (+7 more)

### Community 28 - "ImportPlan"
Cohesion: 0.15
Nodes (18): ImportResult, PlannedSection, defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList(), renderRegionScript() (+10 more)

### Community 29 - "age"
Cohesion: 0.24
Nodes (8): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownsPod(), TestOperatorRunningImageWithNoContainers(), TestOwnsPodFallsBackToTheNameInfix()

### Community 31 - "Config"
Cohesion: 0.11
Nodes (14): keyValueEntries, TestRoleNames(), RoleNames(), checkCredentialChars(), foldToEnvVar(), Config, requireAll(), requireKeyValue() (+6 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "newTeardownApplyFixture"
Cohesion: 0.31
Nodes (9): driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, newTeardownApplyFixture(), TestImportOpsImportApplySeparatesFirstSectionFromMain(), TestImportOpsImportApplyTearsDownExistingVPNBeforeApplyingArtifact() (+1 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (44): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, os.DirEntry, os.FileInfo, os.FileMode (+36 more)

### Community 35 - "Role"
Cohesion: 0.07
Nodes (22): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+14 more)

### Community 36 - "bg"
Cohesion: 0.14
Nodes (40): TestCtrManagerConfirmWiring(), wantRemove(), containerRenderRole(), containerRole(), containerWhat(), ctrLogin(), ctrManager(), ctrOps() (+32 more)

### Community 37 - "recRunner"
Cohesion: 0.19
Nodes (4): TestCanIAnswerReadsTheLastLine(), TestTransportExecArgs(), recRunner, rrCall

### Community 38 - "haCfg"
Cohesion: 0.08
Nodes (35): New(), TestNewDefaults(), AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestOperatorAdminSecretNameMatchesTheBrokerNames() (+27 more)

### Community 39 - "ReplSite"
Cohesion: 0.17
Nodes (12): Replication, missingListedVPNs(), PlannedRoles(), RoleAtSite(), TestPlannedRolesNamesEveryListedVPN(), TestRoleAtSiteIsTheComplement(), curlConfigFlag(), Ops (+4 more)

### Community 40 - "diff_test.go"
Cohesion: 0.15
Nodes (25): blockKey, mergedBlock, mergedLine, ancestorReported(), DiffBlocks(), indexByIdentity(), keyOf(), keySet() (+17 more)

### Community 41 - "runPlatform"
Cohesion: 0.29
Nodes (12): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+4 more)

### Community 42 - "inject_test.go"
Cohesion: 0.29
Nodes (17): svcKey, describeKey(), InjectShutdown(), ServiceShutdownStyles(), enablementCounts(), linesOf(), parseInjectCapture(), TestInjectShutdownAlreadySandwichedPassesThroughUnchanged() (+9 more)

### Community 43 - ".releaseToBackup"
Cohesion: 0.19
Nodes (11): countContains(), field(), TestCountContains(), TestField(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), noReleaseActivityScript(), activity() (+3 more)

### Community 44 - "github.com/spf13/cobra.Command"
Cohesion: 0.07
Nodes (116): opFunc, roleOpFunc, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, argumentLine(), availableSubs(), writeCommand(), addCommands() (+108 more)

### Community 46 - "k8s/inspect.go"
Cohesion: 0.16
Nodes (22): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), ownedPods(), splitWatch(), TestWatchFromContainersReadsEveryAllNamespacesSpelling() (+14 more)

### Community 47 - "k8s/matechannel_test.go"
Cohesion: 0.22
Nodes (25): CLIScriptPath(), execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf(), replCfg() (+17 more)

### Community 48 - "runRootWith"
Cohesion: 0.10
Nodes (35): echoRunner(), App, runRootWith(), TestCLICommand(), TestConfiguredRouternameSurvivesTheFallback(), TestCtrConfirmDeclined(), TestCtrRestartConfirmGate(), TestDeployOperatorNoPromptStaysUnknownFlag() (+27 more)

### Community 49 - "config_test.go"
Cohesion: 0.20
Nodes (45): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, span, go_pkg_bytes, go_pkg_context, go_pkg_encoding_base64 (+37 more)

### Community 50 - "Command"
Cohesion: 0.18
Nodes (10): cmdField, guardedCmd, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandUnmarshal(), Command, Config, guardCommandOf() (+2 more)

### Community 51 - "newRootCmd"
Cohesion: 0.07
Nodes (38): TestAbbreviationDocs(), applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias() (+30 more)

### Community 52 - "exportconfig_test.go"
Cohesion: 0.11
Nodes (27): opRunner, runFailRunner, go_pkg_runtime, ExitCode(), childStatusError(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitStatusIsScopedToInteractiveSessions(), exportconfigDriverOutput() (+19 more)

### Community 53 - "The eight rules"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 54 - "Operations"
Cohesion: 0.08
Nodes (26): Bringing up a fresh cluster, Broker scope is a fixed classification, Configuring a site, Data replication, Docker and Podman mechanics, Exit codes, Exporting and importing configuration, Extra CLI users differ by platform (+18 more)

### Community 55 - "Test catalogue"
Cohesion: 0.10
Nodes (21): abbrev_test.go, convert_test.go, Coverage, examples_test.go, Fixtures and doubles, Injectable seams, internal/abbrev, internal/convert (+13 more)

### Community 56 - ".LeaderLocal"
Cohesion: 0.17
Nodes (10): TestLastLines(), TestLastLinesEqualCount(), assertLeaderScript(), TestAssertLeaderScript(), defaultLocalAddrs(), Ops, hostMatches(), shortHost() (+2 more)

### Community 57 - "RenderDiffResult"
Cohesion: 0.23
Nodes (11): BlockDiff, DiffResult, blockLabel(), diffRow(), orDash(), RenderDiffResult(), reportBlockLines(), TestDiffBlockLabelFormatsIdentityForTheReport() (+3 more)

### Community 58 - "scripts.go"
Cohesion: 0.11
Nodes (25): showCmd, currentConfigScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript(), enableDefaultUsersScript(), enableDefaultVPNScript() (+17 more)

### Community 59 - ".rpc"
Cohesion: 0.09
Nodes (24): AdminState, backupTarget, Credential, sempMate, TestHTTPStatusHelpers(), schemaTransport(), TestTransportVocabularyIsThreeWords(), anyHTTP2xx() (+16 more)

### Community 60 - "confirmAction"
Cohesion: 0.17
Nodes (26): exportconfigReadCounter, layer, go_pkg_bufio, io.Reader, io.Writer, TestConfirmDeleteShortcut(), TestPromptYes(), TestPromptYesNo() (+18 more)

### Community 61 - ".configRows"
Cohesion: 0.12
Nodes (25): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminPassState(), adminSecretRow(), allCustomMounted() (+17 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.17
Nodes (36): failOnCall(), fakeEnv(), Manager, healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+28 more)

### Community 64 - "output_test.go"
Cohesion: 0.13
Nodes (23): TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), progress(), TestCheckReportSkipsEmptySections(), New(), NewFunc(), sink() (+15 more)

### Community 65 - "Configuration"
Cohesion: 0.14
Nodes (14): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+6 more)

### Community 66 - "Load"
Cohesion: 0.07
Nodes (52): Example, go_pkg_embed, TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError() (+44 more)

### Community 67 - "MateConfig"
Cohesion: 0.10
Nodes (38): MateConfig, TestMateChannelShowReplication(), cliTransport(), containsEndpoint(), labelValue(), normTransport(), parseHostPort(), ParseShowReplication() (+30 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "loadSample"
Cohesion: 0.15
Nodes (17): Annotate(), sectionBeginMarker(), TestAnnotateIsDeterministic(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), TestSectionBeginMarkerAdvisoryDisposition(), TestWriteBlocksOpensAndClosesOnChange(), writeBlocks() (+9 more)

### Community 71 - "localCfg"
Cohesion: 0.09
Nodes (36): TestLocalMateRunsThroughOps(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK(), TestBackupRevertActivitySuccess() (+28 more)

### Community 72 - ".Run"
Cohesion: 0.15
Nodes (24): testing.M, runCtr(), TestConfigStepsDoNotLeakSecrets(), TestCtrConfigDryRun(), TestCtrDiagnosticsDryRun(), TestCtrErrorPaths(), TestCtrExecCLIPathSeparator(), TestCtrRoleArgCount() (+16 more)

### Community 73 - "internal/k8s"
Cohesion: 0.10
Nodes (20): adminsecret_test.go, check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go (+12 more)

### Community 74 - "annotate_test.go"
Cohesion: 0.08
Nodes (39): applyMetaField(), Capture, isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV(), readMarker() (+31 more)

### Community 75 - "usagef"
Cohesion: 0.10
Nodes (23): childExitError, usageError, github.com/spf13/cobra.PositionalArgs, logArgs(), App, newConvertCmd(), runConvert(), asUsage() (+15 more)

### Community 76 - "occCluster"
Cohesion: 0.18
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 77 - "scripts_test.go"
Cohesion: 0.19
Nodes (14): parseVPNNames(), productKeyScript(), productKeysScript(), removeProductKeysScript(), TestDisableDefaultUsersScriptQuoting(), TestParseVPNNames(), TestParseVPNNamesKeepsAMultiWordName(), TestParseVPNNamesNoSeparator() (+6 more)

### Community 78 - "Ops"
Cohesion: 0.31
Nodes (3): validCLILine(), Ops, showVPNBareScript()

### Community 79 - "Quadlet"
Cohesion: 0.13
Nodes (19): ContainerNoFile(), Config, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), EnvPairs(), escapePercent(), groupKey(), itoa() (+11 more)

### Community 81 - "eqArgs"
Cohesion: 0.36
Nodes (8): TestPreflightRunsBeforeAnything(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime(), TestManagerReachableProbesRuntimeThenCompose(), withWrapper(), wrappedCtrCfg(), eqArgs()

### Community 82 - "TestServerCert"
Cohesion: 0.16
Nodes (14): TestDiagnostics(), TestPathHelpers(), TestServerCert(), TestServerCertBundleOrder(), TestServerCertBundleRefusesAKeyBearingCert(), TestServerCertBundleReportsAnUnreadableFile(), TestServerCertBundleRequiresBothHalves(), ServerCertBundle() (+6 more)

### Community 83 - "limits_test.go"
Cohesion: 0.14
Nodes (26): go_pkg_slices, Manager, limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager() (+18 more)

### Community 84 - "HARoles"
Cohesion: 0.11
Nodes (13): time.Time, BrokerPodSuffixShape(), Cluster, Cluster, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes(), HARoles() (+5 more)

### Community 85 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 86 - "k8s/inspect_test.go"
Cohesion: 0.14
Nodes (20): sectionLevel(), podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture() (+12 more)

### Community 88 - "replApp"
Cohesion: 0.32
Nodes (13): mateChannel(), mateSEMPPassword(), App, replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer(), TestMateChannelPicksTheMechanismTheSiteDeclares() (+5 more)

### Community 89 - "render_test.go"
Cohesion: 0.15
Nodes (29): BrokerCR(), containerArtifacts(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets(), TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPrivilegePosture(), TestComposeQuotesTheCpuset() (+21 more)

### Community 90 - "main_test.go"
Cohesion: 0.17
Nodes (18): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+10 more)

### Community 91 - "RuleFor"
Cohesion: 0.09
Nodes (34): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+26 more)

### Community 92 - "step"
Cohesion: 0.15
Nodes (18): mateChannelFunc, bufio.Reader, App, lineSink(), step(), confirmReplicationConfig(), App, opConfigureReplication() (+10 more)

### Community 93 - ".applyContainerDefaults"
Cohesion: 0.13
Nodes (17): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults() (+9 more)

### Community 94 - "IsAbsHostPath"
Cohesion: 0.16
Nodes (9): expandTilde(), expandTildeToken(), Config, IsAbsHostPath(), isPathSep(), TestContainerHostDirsExpandATilde(), TestExpandTilde(), TestIsAbsHostPath() (+1 more)

### Community 95 - ".resolveSecretRefs"
Cohesion: 0.50
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 96 - "Exec"
Cohesion: 0.29
Nodes (4): Exec, os/exec.Cmd, TestResolveMissingBinaryIsActionable(), Resolve()

### Community 97 - "replConfig"
Cohesion: 0.31
Nodes (10): Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationEmptyViaAccepted(), TestValidateReplicationRejects() (+2 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "transform_test.go"
Cohesion: 0.16
Nodes (19): injectedBlock(), ClearExistingSyslogs(), InjectVPNServiceShutdown(), Capture, loadRealCapture(), TestClearExistingClientCAsIntersects(), TestClearExistingSyslogsAgainstTheRealCapture(), TestClearExistingSyslogsNoTargetEntriesIsANoOp() (+11 more)

### Community 100 - ".checkUserManagerLimits"
Cohesion: 0.23
Nodes (8): delegateSetting(), Manager, limitText(), missingControllers(), parseLimit(), parseUnitProps(), TestMissingControllers(), TestParseLimit()

### Community 101 - "BrokerType"
Cohesion: 0.16
Nodes (11): BrokerType, cliMate, Omission, brokerTypeFromSchema(), Capture, readHeaderField(), TestBrokerTypeFromSchema(), bannerType() (+3 more)

### Community 102 - "newTestMate"
Cohesion: 0.16
Nodes (17): CLIRunner, fakeRun, Ops, NewCLIMate(), mateReplies(), newTestMate(), TestMateChannelDescribeNamesTheTarget(), TestMateChannelPreflightLearnsTheBrokerType() (+9 more)

### Community 103 - "certFixture"
Cohesion: 0.14
Nodes (19): ResolveSecretValues(), certFixture(), TestDeletePodmanRemovesTheBundleAfterTheUnit(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeWritingTheComposeFile(), TestDeployDockerPassesTheBundleAsEnvNotArgv(), TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(), TestDeployPodmanFailsWhenTheBundleCannotBeWritten() (+11 more)

### Community 104 - "Compose"
Cohesion: 0.14
Nodes (14): HealthCheck, Compose(), composeEscape(), ComposeProject(), ContainerSecret, healthCmd(), secretFilePath(), healthCheckFixture() (+6 more)

### Community 105 - "internal/cli"
Cohesion: 0.15
Nodes (13): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, euid_test.go, examples_test.go (+5 more)

### Community 106 - "MateChannel"
Cohesion: 0.19
Nodes (15): MateChannel, sortedKeys(), sortedVPNs(), requirePrimaryActive(), requireVirtualRouterNamesMatch(), sortedSiteNames(), SwitchPreflight(), replSites() (+7 more)

### Community 107 - "hasCall"
Cohesion: 0.23
Nodes (16): hasCall(), TestDisableDefaultUsersShowVPNError(), TestReleaseToBackupReleasedTimeout(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), TestConfigureReplicationClosingShowFailureStillReportsTheApply() (+8 more)

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "Config"
Cohesion: 0.07
Nodes (21): AdditionalUser, Image, Node, Redundancy, SEMP, TLS, atoiPrefix(), Config (+13 more)

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "ParseVPNReplication"
Cohesion: 0.13
Nodes (17): colSpan, ValidVPNName(), validVPNName(), dashSpans(), flagByte(), gutterClear(), ParseVPNReplication(), sliceSpan() (+9 more)

### Community 112 - "capRunner"
Cohesion: 0.19
Nodes (11): capCall, capRunner, NewTransport(), dockerCfg(), podmanCfg(), TestTransportCopy(), TestTransportEchoHidesSEMPConfig(), TestTransportEchoHidesUploadBody() (+3 more)

### Community 113 - "internal/config"
Cohesion: 0.12
Nodes (16): adminsecret_test.go, command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go, hostpath_test.go (+8 more)

### Community 114 - "Removing a broker: what stays, what goes"
Cohesion: 0.40
Nodes (5): Every destructive command confirms, Removing a broker: what stays, what goes, Removing the operator does not always remove it, The layer flag raises the question, The namespace is only offered when it is empty

### Community 116 - "Manager"
Cohesion: 0.21
Nodes (4): Manager, idMapCovers(), origin(), runUserIDs()

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - ".stateRows"
Cohesion: 0.16
Nodes (14): containerState, healthState, decodeInspect(), Manager, orUnknown(), printable(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown() (+6 more)

### Community 119 - "internal/engine"
Cohesion: 0.67
Nodes (3): internal/engine, resolve_test.go, runner_test.go

### Community 120 - ".validateContainer"
Cohesion: 0.16
Nodes (11): TestValidateProbeCommandAccepts(), TestValidateProbeCommandRejects(), cpuSetCount(), cpuSetRange(), TestCPUSetCount(), TestCPUSetRange(), invertedCPUSetRange(), missingErr() (+3 more)

### Community 121 - "Ops"
Cohesion: 0.18
Nodes (5): time.Duration, Ops, runCLISkeleton(), shQuote(), TestShQuoteHandlesASingleQuote()

### Community 122 - "coverage_test.go"
Cohesion: 0.05
Nodes (57): cliRunNames(), matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError() (+49 more)

### Community 123 - ".ConfigureReplication"
Cohesion: 0.24
Nodes (8): Ops, isRunCLIRejection(), newlineIf(), replPhase1Rejected(), replPhase2Rejected(), replTransportFailure(), replVPNList(), showVPNReplicationScript()

### Community 124 - "ReplRole"
Cohesion: 0.20
Nodes (8): ReplRole, rejectionIn(), ReplicationConfigResult, setReplicationRoleScript(), sempReplRole(), reportReplicationConfig(), TestReportReplicationConfigMateAppliedNamesWhatPhase1Stopped(), TestReportReplicationConfigMateSkippedSaysNothingWasWritten()

### Community 125 - "parseServiceLine"
Cohesion: 0.22
Nodes (10): lineRole, serviceLine, shutdownStyle, classifyServiceRest(), containsPortCommand(), fieldsUnquoted(), isPortCommand(), parseServiceLine() (+2 more)

### Community 126 - ".applyScalingTierDefaults"
Cohesion: 0.22
Nodes (7): scalingTier, containerMem(), Config, TestContainerMem(), TestScalingTiers(), TestTierForRejectsOffTierValues(), tierFor()

### Community 127 - "runExport"
Cohesion: 0.22
Nodes (10): confirmImport(), App, pluralVPN(), runExport(), runImport(), TestPluralVPNWording(), emitOrWrite(), nowStamp() (+2 more)

### Community 128 - "operatorversion.go"
Cohesion: 0.18
Nodes (12): compareVersions(), Cluster, imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestImageTag() (+4 more)

### Community 129 - ".decodeScalingEntry"
Cohesion: 0.31
Nodes (7): scalingKey, scalingSpelling, Scaling, Scaling, yaml.Node, scalingKeyIndex(), scalingKeyList()

### Community 130 - "newBlock"
Cohesion: 0.28
Nodes (9): token, firstQuoted(), newBlock(), TestNewBlockMessageSpoolQualifierHasNoOwnName(), TestNewBlockOpener(), TestNewBlockVPNNameWithSpaces(), TestNewBlockVPNQualifierWithSpaces(), tokenizeOpener() (+1 more)

### Community 131 - ".preflightOne"
Cohesion: 0.50
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "ReplVia"
Cohesion: 0.29
Nodes (7): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, viaDescription(), validateReplVia(), validateReplViaSEMP()

### Community 134 - "parsePort"
Cohesion: 0.25
Nodes (8): LoadBalancer, cut(), parsePort(), splitUser(), TestParsePort(), writeKeyValueEntry(), writeLBAnnotations(), portSpec

### Community 135 - "sempPort"
Cohesion: 0.29
Nodes (7): bridgeHostPort(), sempPort(), TestSempPortBridgeFallsBackToPlaintext(), TestSempPortBridgePrefersTLS(), TestSempPortBridgeRefusesNoMapping(), TestSempPortHostMode(), TestSempPortWithoutCertificateStaysPlaintext()

### Community 136 - "ContainerSecrets"
Cohesion: 0.29
Nodes (7): ContainerSecrets(), containerSecretSpecs(), SecretPreflight(), TestFileBackedSecretIsExemptFromSecretPreflight(), TestPodmanNeverGetsAFileBackedSecret(), TestServerCertIsADockerOnlySecret(), secretSpec

### Community 137 - "GenSecrets"
Cohesion: 0.06
Nodes (37): GenBroker(), GenSecrets(), Cluster, joinManifests(), namespaceManifest(), ownedSecretNames(), adminCfg(), secretNamesIn() (+29 more)

### Community 138 - "importIgnore"
Cohesion: 0.47
Nodes (5): importIgnore(), TestImportIgnoreKeepsUnclassifiedSectionsInTheDiff(), TestImportOpsImportIgnore(), BridgeEnablementInverted(), TestShutdownBridgesLinesAreExcusedByTheDiff()

### Community 139 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 140 - "ParseRole"
Cohesion: 0.50
Nodes (5): TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), ParseRole(), RoleAbbrev()

### Community 141 - "CheckHostPath"
Cohesion: 0.40
Nodes (5): CanonicalDuration(), TestCanonicalDuration(), CheckHostPath(), TestCheckHostPathAccepts(), TestCheckHostPathRejects()

## Knowledge Gaps
- **236 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+231 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 354 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **18 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Config` connect `Config` to `cli_test.go`, `context.Context`, `sempPort`, `K8sConfig`, `Manager`, `GenOperator`, `GenSecrets`, `NewCluster`, `eqArgs`, `convert_test.go`, `ContainerSecrets`, `manager_test.go`, `strings.Builder`, `AdminSecret`, `newTestOps`, `ResolveDomainCerts`, `Role`, `haCfg`, `ReplSite`, `github.com/spf13/cobra.Command`, `k8s/matechannel_test.go`, `config_test.go`, `.configRows`, `Cluster`, `localCfg`, `occCluster`, `Quadlet`, `eqArgs`, `TestServerCert`, `HARoles`, `containerTransport`, `render_test.go`, `step`, `.applyContainerDefaults`, `certFixture`, `Compose`, `capRunner`, `Ops`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `ops_k8s.go`, `context.Context`, `Manager`, `K8sConfig`, `ParseBlocks`, `ParseRole`, `newTestOps`, `Set`, `renderDriver`, `ImportPlan`, `Config`, `bg`, `haCfg`, `.releaseToBackup`, `k8s/matechannel_test.go`, `config_test.go`, `.LeaderLocal`, `scripts.go`, `.rpc`, `Ops`, `Quadlet`, `HARoles`, `containerTransport`, `step`, `BrokerType`, `newTestMate`, `Config`, `Ops`, `.ConfigureReplication`, `runExport`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Why does `Manager` connect `Manager` to `context.Context`, `bg`, `github.com/spf13/cobra.Command`, `Config`, `manager_test.go`, `config_test.go`, `confirmAction`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 42 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 42 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _236 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ops_k8s.go` be split into smaller, more focused modules?**
  _Cohesion score 0.1326530612244898 - nodes in this community are weakly interconnected._
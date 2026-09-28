# Graph Report - solace-cnt-scripts  (2026-09-28)

## Corpus Check
- 196 files · ~606,452 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 3810 nodes · 15999 edges · 145 communities (130 shown, 15 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 2205 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `ca5d9b04`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- captureStdout
- Sink
- context.Context
- cli_test.go
- testing.T
- Commands
- Manager
- K8sConfig
- allowcommand_test.go
- CheckCommand
- .Run
- NewCluster
- GenSecrets
- convert_test.go
- transform_test.go
- manager_test.go
- github.com/spf13/cobra.Command
- AdminSecret
- dev.sh
- parse
- dev.ps1
- Set
- BuildSwitchPlan
- verify_local_test.go
- runner_test.go
- renderDriver
- eqArgs
- ImportPlan
- Platform
- solace
- Config
- CLAUDE.md
- ServerCertBundle
- ResolveDomainCerts
- Role
- App
- recRunner
- parseKV
- completion_test.go
- diff_test.go
- runPlatform
- inject_test.go
- .releaseToBackup
- commands.go
- Echo
- k8s/inspect.go
- ReplSite
- localCfg
- config_test.go
- BrokerType
- newRootCmd
- exportconfig_test.go
- The eight rules
- Operations
- Test catalogue
- newOperatorCmd
- MateChannel
- scripts.go
- MateConfig
- confirm.go
- .configRows
- rootlessMgr
- strings.Builder
- output_test.go
- Configuration
- runRootWith
- newTestMate
- Command reference
- Developer guide
- replApp
- emitYAML
- age
- internal/k8s
- parseBody
- usagef
- .ConfigureReplication
- Command
- newTeardownApplyFixture
- ParseBlocks
- go_pkg_solace_internal_abbrev
- k8sOps
- k8s/matechannel_test.go
- limits_test.go
- New
- Troubleshooting
- k8s/inspect_test.go
- ParseVPNReplication
- operatorversion_test.go
- render_test.go
- vulnjudge/main.go
- RuleFor
- step
- .applyContainerDefaults
- Config
- .resolveSecretRefs
- ParseRole
- replConfig
- Abbreviations
- haCfg
- .checkUserManagerLimits
- .Preflight
- Load
- container/secrets_test.go
- lastLines
- internal/cli
- desiredWatch
- Compose
- What `import-config` applies
- Config
- internal/broker
- GenOperator
- capRunner
- internal/config
- newEchoMgr
- HARoles
- Quadlet
- internal/container
- .stateRows
- replicationVPNLines
- Convert
- Ops
- newTestOps
- .validateContainerArtifactValues
- newBlock
- Exporting and importing configuration
- .applyScalingTierDefaults
- watchCluster
- logArgs
- .decodeScalingEntry
- .preflightOne
- ReplVia
- Cluster
- ResolveEnvPath
- CheckHostPath
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

## Communities (145 total, 15 thin omitted)

### Community 0 - "bg"
Cohesion: 0.16
Nodes (34): bg(), App, k8sAdminPassword(), k8sCluster(), k8sContext(), k8sLogin(), opK8sCLI(), opK8sConfigLeader() (+26 more)

### Community 1 - "captureStdout"
Cohesion: 0.09
Nodes (37): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook(), loadDirect() (+29 more)

### Community 2 - "Sink"
Cohesion: 0.18
Nodes (5): solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), columnWidths(), Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.04
Nodes (18): scriptedMate, context.Context, Manager, idMapCovers(), origin(), runUserIDs(), EnvRunner, Runner (+10 more)

### Community 4 - "cli_test.go"
Cohesion: 0.08
Nodes (55): os.File, capture(), captureStderr(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone(), runStatusStderr() (+47 more)

### Community 5 - "testing.T"
Cohesion: 0.02
Nodes (212): testing.T, TestRouterNameReadsTheCaptureHeader(), bridgeHostPort(), sempPort(), sempV1OK(), TestSempPortBridgeFallsBackToPlaintext(), TestSempPortBridgePrefersTLS(), TestSempPortBridgeRefusesNoMapping() (+204 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (8): certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing()

### Community 8 - "K8sConfig"
Cohesion: 0.15
Nodes (8): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, boolStr(), writeSecurity()

### Community 9 - "allowcommand_test.go"
Cohesion: 0.29
Nodes (10): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+2 more)

### Community 10 - "CheckCommand"
Cohesion: 0.15
Nodes (22): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+14 more)

### Community 11 - ".Run"
Cohesion: 0.14
Nodes (27): testing.M, allowRuntime(), runCtr(), TestAnnounceCommandsNamesResolvedBinaries(), TestConfigStepsDoNotLeakSecrets(), TestCtrConfigDryRun(), TestCtrDiagnosticsDryRun(), TestCtrErrorPaths() (+19 more)

### Community 12 - "NewCluster"
Cohesion: 0.08
Nodes (60): TestCheckStopsProbingWhenUnreachable(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+52 more)

### Community 13 - "GenSecrets"
Cohesion: 0.08
Nodes (33): GenBroker(), GenSecrets(), joinManifests(), adminCfg(), secretNamesIn(), TestAdminSecretStatesEmitTheirArtifacts(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree() (+25 more)

### Community 14 - "convert_test.go"
Cohesion: 0.19
Nodes (30): checkGolden(), convertOK(), hasWarning(), strictDecode(), TestConvertAdminSecretAlias(), TestConvertAdminUserIsDroppedOnEveryPlatform(), TestConvertBadBooleanWarns(), TestConvertBadNumberWarns() (+22 more)

### Community 15 - "transform_test.go"
Cohesion: 0.09
Nodes (51): Block, TargetState, regexp.Regexp, BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects(), ClearTargetVirtualHostnames() (+43 more)

### Community 16 - "manager_test.go"
Cohesion: 0.07
Nodes (97): inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer() (+89 more)

### Community 17 - "github.com/spf13/cobra.Command"
Cohesion: 0.09
Nodes (39): shorthand, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, github.com/spf13/pflag.FlagSet, renderAbbrevDocs(), treeShorthands(), writeResolutionNotes(), writeShorthandTable() (+31 more)

### Community 18 - "AdminSecret"
Cohesion: 0.12
Nodes (19): AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), checkGolden(), decodeDataValue(), TestAdditionalUsersSecretCarriesBothHalves(), TestAdditionalUsersSecretIsAbsentWithNoUsers() (+11 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "parse"
Cohesion: 0.11
Nodes (20): segment, vars, countMarkers(), resolvePlatform(), TestParseArrayElementsPreserveQuoting(), TestParseAssignmentForms(), TestParseCRLF(), TestParseDoubleQuotedBackslashEscapes() (+12 more)

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
Nodes (46): cliScriptPath(), defaultLocalAddrs(), curlCalls(), Ops, isCurl(), newLocalOps(), rd(), seqTransport() (+38 more)

### Community 25 - "runner_test.go"
Cohesion: 0.11
Nodes (28): TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), MaskEnv(), quoteTok(), captureStdout(), helperCommand() (+20 more)

### Community 26 - "renderDriver"
Cohesion: 0.13
Nodes (18): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), TestKeywordPatternEscapesEachPhrase() (+10 more)

### Community 27 - "eqArgs"
Cohesion: 0.11
Nodes (30): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), saCfg() (+22 more)

### Community 28 - "ImportPlan"
Cohesion: 0.14
Nodes (20): ImportResult, PlannedSection, isPragma(), defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList() (+12 more)

### Community 29 - "Platform"
Cohesion: 0.12
Nodes (30): checkFlagPlatforms(), commandPlatforms(), declaredList(), flagOnlyOn(), App, parsePlatformList(), platformSuffix(), prepare() (+22 more)

### Community 31 - "Config"
Cohesion: 0.10
Nodes (17): keyValueEntries, TestValidateContainerArtifactValues(), TestRoleNames(), RoleNames(), checkCredentialChars(), foldToEnvVar(), Config, invertedCPUSetRange() (+9 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "ServerCertBundle"
Cohesion: 0.40
Nodes (5): TestServerCertBundleOrder(), TestServerCertBundleRefusesAKeyBearingCert(), TestServerCertBundleReportsAnUnreadableFile(), TestServerCertBundleRequiresBothHalves(), ServerCertBundle()

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (45): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, os.DirEntry, os.FileInfo, os.FileMode (+37 more)

### Community 35 - "Role"
Cohesion: 0.06
Nodes (24): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+16 more)

### Community 36 - "App"
Cohesion: 0.11
Nodes (32): TestCtrManagerConfirmWiring(), childExit(), TestChildExitKeepsItsMessage(), TestExitCodeIsNeverNegative(), containerRenderRole(), containerRole(), ctrLogin(), ctrManager() (+24 more)

### Community 37 - "recRunner"
Cohesion: 0.14
Nodes (10): TestCanIAnswerReadsTheLastLine(), TestExecutorRefusesUnapprovedRuntime(), unapprovedCfg(), NewTransport(), TestTransportCopy(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest() (+2 more)

### Community 38 - "parseKV"
Cohesion: 0.08
Nodes (27): applyMetaField(), Capture, markerSection(), markerVerb(), parseKV(), readMarker(), renderKV(), splitKV() (+19 more)

### Community 39 - "completion_test.go"
Cohesion: 0.18
Nodes (20): driveBash(), runComplete(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion(), TestCompletionHelpKeepsTheLoadingInstructions(), TestCompletionHelpStillWorks() (+12 more)

### Community 40 - "diff_test.go"
Cohesion: 0.09
Nodes (36): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, ancestorReported(), blockLabel(), DiffBlocks() (+28 more)

### Community 41 - "runPlatform"
Cohesion: 0.29
Nodes (12): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+4 more)

### Community 42 - "inject_test.go"
Cohesion: 0.15
Nodes (27): lineRole, serviceLine, shutdownStyle, svcKey, classifyServiceRest(), containsPortCommand(), describeKey(), fieldsUnquoted() (+19 more)

### Community 43 - ".releaseToBackup"
Cohesion: 0.13
Nodes (15): countContains(), field(), assertLeaderScript(), showRedundancyDetailScript(), showRedundancyLocalScript(), TestAssertLeaderScript(), TestZipConfigsScript(), zipConfigsScript() (+7 more)

### Community 44 - "commands.go"
Cohesion: 0.22
Nodes (46): opFunc, roleOpFunc, addPodFlag(), App, group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerConfigureCmd() (+38 more)

### Community 45 - "Echo"
Cohesion: 0.14
Nodes (8): interactiveFailRunner, Exec, os/exec.Cmd, TestExecIsSilentWithoutVerbose(), Echo, NewExec(), Quote(), TestQuote()

### Community 46 - "k8s/inspect.go"
Cohesion: 0.14
Nodes (19): time.Time, imageFromDeployment(), operatorRunningImage(), ownedPods(), TestOperatorRunningImageWithNoContainers(), operatorItem(), TestAmbiguousOperatorIsNotFoldedIntoNotInstalled(), TestFindOperatorDeploymentIsScopedByNamespace() (+11 more)

### Community 47 - "ReplSite"
Cohesion: 0.15
Nodes (14): ReplRole, Replication, missingListedVPNs(), PlannedRoles(), RoleAtSite(), TestRoleAtSiteIsTheComplement(), setReplicationRoleScript(), curlConfigFlag() (+6 more)

### Community 48 - "localCfg"
Cohesion: 0.08
Nodes (35): TestLocalMateRunsThroughOps(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK(), TestBackupRevertActivitySuccess() (+27 more)

### Community 49 - "config_test.go"
Cohesion: 0.20
Nodes (45): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, span, go_pkg_bytes, go_pkg_context, go_pkg_crypto_sha256 (+37 more)

### Community 50 - "BrokerType"
Cohesion: 0.14
Nodes (15): BrokerType, StripMarkers(), TestStripMarkersKeepsPragmaAndSectionComments(), TestStripMarkersNoMarkersPassthrough(), bannerType(), checkProvenance(), checkSameType(), describeScope() (+7 more)

### Community 51 - "newRootCmd"
Cohesion: 0.10
Nodes (27): TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing() (+19 more)

### Community 52 - "exportconfig_test.go"
Cohesion: 0.14
Nodes (22): runFailRunner, go_pkg_runtime, ExitCode(), childStatusError(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitStatusIsScopedToInteractiveSessions(), App, newExportconfigRunner() (+14 more)

### Community 53 - "The eight rules"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 54 - "Operations"
Cohesion: 0.08
Nodes (24): Bringing up a fresh cluster, Configuring a site, Data replication, Docker and Podman mechanics, Every destructive command confirms, Exit codes, Extra CLI users differ by platform, Operations (+16 more)

### Community 55 - "Test catalogue"
Cohesion: 0.08
Nodes (24): abbrev_test.go, convert_test.go, Coverage, examples_test.go, Fixtures and doubles, Injectable seams, internal/abbrev, internal/convert (+16 more)

### Community 56 - "newOperatorCmd"
Cohesion: 0.30
Nodes (14): newOperatorCmd(), newOperatorDeployCmd(), newOperatorGenerateCmd(), newOperatorLogsCmd(), newOperatorRemoveCmd(), newOperatorRestartCmd(), newOperatorStartCmd(), newOperatorStatusCmd() (+6 more)

### Community 57 - "MateChannel"
Cohesion: 0.22
Nodes (12): MateChannel, requirePrimaryActive(), requireVirtualRouterNamesMatch(), sortedSiteNames(), SwitchPreflight(), replSites(), TestSwitchPreflightHealthyPair(), TestSwitchPreflightRequiresMatchingVirtualRouterNames() (+4 more)

### Community 58 - "scripts.go"
Cohesion: 0.06
Nodes (45): showCmd, runCLISkeleton(), validCLILine(), validName(), Ops, shQuote(), TestShQuoteHandlesASingleQuote(), rejectionIn() (+37 more)

### Community 59 - "MateConfig"
Cohesion: 0.09
Nodes (24): backupTarget, Credential, MateConfig, sempMate, TestHTTPStatusHelpers(), labelValue(), parseHostPort(), parseShowReplicationAppliance() (+16 more)

### Community 60 - "confirm.go"
Cohesion: 0.16
Nodes (26): exportconfigReadCounter, layer, go_pkg_bufio, io.Reader, io.Writer, TestConfirmDeleteShortcut(), TestPromptYes(), TestPromptYesNo() (+18 more)

### Community 61 - ".configRows"
Cohesion: 0.13
Nodes (23): orNone(), orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminSecretRow(), containsString() (+15 more)

### Community 62 - "rootlessMgr"
Cohesion: 0.13
Nodes (36): failOnCall(), fakeEnv(), Manager, healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+28 more)

### Community 63 - "strings.Builder"
Cohesion: 0.10
Nodes (29): WeightedNodeTerm, strings.Builder, mdRow(), writeAbbrevTable(), LoadBalancer, NodeAffinity, NodeMatchExpr, Placement (+21 more)

### Community 64 - "output_test.go"
Cohesion: 0.16
Nodes (17): NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLevelOrderKeepsSkipBelowFail(), TestLineAddsNoPrefix() (+9 more)

### Community 65 - "Configuration"
Cohesion: 0.14
Nodes (14): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+6 more)

### Community 66 - "runRootWith"
Cohesion: 0.11
Nodes (33): echoRunner(), App, runRootWith(), TestCLICommand(), TestConfiguredRouternameSurvivesTheFallback(), TestCtrConfirmDeclined(), TestCtrRestartConfirmGate(), TestDeployOperatorNoPromptStaysUnknownFlag() (+25 more)

### Community 67 - "newTestMate"
Cohesion: 0.16
Nodes (17): CLIRunner, fakeRun, Ops, NewCLIMate(), mateReplies(), newTestMate(), TestMateChannelDescribeNamesTheTarget(), TestMateChannelPreflightLearnsTheBrokerType() (+9 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "replApp"
Cohesion: 0.32
Nodes (13): mateChannel(), mateSEMPPassword(), App, replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer(), TestMateChannelPicksTheMechanismTheSiteDeclares() (+5 more)

### Community 71 - "emitYAML"
Cohesion: 0.22
Nodes (9): doc, boolOf(), commentSafe(), emitYAML(), joinDomainCertPath(), kubeCommand(), redundancy(), scalar() (+1 more)

### Community 72 - "age"
Cohesion: 0.32
Nodes (5): age(), roleRank(), Cluster, ownsPod(), TestOwnsPodFallsBackToTheNameInfix()

### Community 73 - "internal/k8s"
Cohesion: 0.10
Nodes (20): adminsecret_test.go, check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go (+12 more)

### Community 74 - "parseBody"
Cohesion: 0.12
Nodes (18): Omission, Region, isMarker(), markerRegion(), TestIsMarkerDiscriminatesNamespace(), TestMarkerRegionBeginAndEnd(), brokerTypeFromSchema(), Capture (+10 more)

### Community 75 - "usagef"
Cohesion: 0.12
Nodes (19): childExitError, usageError, github.com/spf13/cobra.PositionalArgs, noRolePositional(), App, newConvertCmd(), runConvert(), asUsage() (+11 more)

### Community 76 - ".ConfigureReplication"
Cohesion: 0.11
Nodes (15): AdminState, cliMate, QueueState, VPNRepl, Ops, isRunCLIRejection(), newlineIf(), reenableLine() (+7 more)

### Community 77 - "Command"
Cohesion: 0.13
Nodes (13): cmdField, guardedCmd, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandUnmarshal(), TestValidateProbeCommandAccepts(), TestValidateProbeCommandRejects(), Command (+5 more)

### Community 78 - "newTeardownApplyFixture"
Cohesion: 0.31
Nodes (9): driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, newTeardownApplyFixture(), TestImportOpsImportApplySeparatesFirstSectionFromMain(), TestImportOpsImportApplyTearsDownExistingVPNBeforeApplyingArtifact() (+1 more)

### Community 79 - "ParseBlocks"
Cohesion: 0.10
Nodes (27): Annotate(), sectionBeginMarker(), assertFencesBalanced(), TestAnnotateIsDeterministic(), TestAnnotateRegionAndSectionFencesAreBalanced(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), TestSectionBeginMarkerAdvisoryDisposition() (+19 more)

### Community 81 - "k8sOps"
Cohesion: 0.19
Nodes (25): confirmAction(), wantEnable(), wantRemove(), containerWhat(), ctrOps(), opCtrConfigDefaultUsers(), opCtrConfigDefaultVPN(), opCtrConfigDomainCerts() (+17 more)

### Community 82 - "k8s/matechannel_test.go"
Cohesion: 0.23
Nodes (24): CLIArg(), CLIScriptPath(), execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf() (+16 more)

### Community 83 - "limits_test.go"
Cohesion: 0.15
Nodes (25): Manager, limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager(), TestCheckLimitsRootlessRefusesUndelegatedControllers() (+17 more)

### Community 84 - "New"
Cohesion: 0.10
Nodes (21): TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), ReplicationConfigResult, confirmImport(), App, pluralVPN(), runExport() (+13 more)

### Community 85 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 86 - "k8s/inspect_test.go"
Cohesion: 0.13
Nodes (25): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), podHealth(), pvcLevel(), replicaLevel(), serviceAddress() (+17 more)

### Community 87 - "ParseVPNReplication"
Cohesion: 0.05
Nodes (58): colSpan, ValidVPNName(), validVPNName(), TestMateChannelShowReplication(), cliTransport(), containsEndpoint(), dashSpans(), flagByte() (+50 more)

### Community 88 - "operatorversion_test.go"
Cohesion: 0.19
Nodes (14): compareVersions(), imageTag(), operatorVersionWarning(), parseVersion(), operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsSilentInAPreview() (+6 more)

### Community 89 - "render_test.go"
Cohesion: 0.13
Nodes (34): BrokerCR(), containerArtifacts(), healthCheckFixture(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets(), TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPrivilegePosture() (+26 more)

### Community 90 - "vulnjudge/main.go"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.08
Nodes (37): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+29 more)

### Community 92 - "step"
Cohesion: 0.14
Nodes (19): mateChannelFunc, bufio.Reader, App, lineSink(), progress(), step(), confirmReplicationConfig(), App (+11 more)

### Community 93 - ".applyContainerDefaults"
Cohesion: 0.14
Nodes (17): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults() (+9 more)

### Community 94 - "Config"
Cohesion: 0.20
Nodes (6): expandTilde(), expandTildeToken(), Config, isPathSep(), TestExpandTilde(), tildeHome()

### Community 95 - ".resolveSecretRefs"
Cohesion: 0.50
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 96 - "ParseRole"
Cohesion: 0.50
Nodes (5): TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), ParseRole(), RoleAbbrev()

### Community 97 - "replConfig"
Cohesion: 0.31
Nodes (10): Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationEmptyViaAccepted(), TestValidateReplicationRejects() (+2 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "haCfg"
Cohesion: 0.11
Nodes (23): AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestReadAdminPasswordArgvAndDecoding(), TestReadAdminPasswordGuardsTheCommand(), TestReadAdminPasswordNamesTheNextStep() (+15 more)

### Community 100 - ".checkUserManagerLimits"
Cohesion: 0.23
Nodes (8): delegateSetting(), Manager, limitText(), missingControllers(), parseLimit(), parseUnitProps(), TestMissingControllers(), TestParseLimit()

### Community 102 - "Load"
Cohesion: 0.07
Nodes (53): Example, go_pkg_embed, TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError() (+45 more)

### Community 103 - "container/secrets_test.go"
Cohesion: 0.17
Nodes (26): ResolveSecretValues(), bundleHash(), certCreate(), certFixture(), TestCertSecretLabelDrivesTheRestart(), TestCertSecretRemovalFailureIsFatal(), TestContainerBundleCarriesTheChain(), TestDeletePodmanToleratesAMissingBundle() (+18 more)

### Community 104 - "lastLines"
Cohesion: 0.67
Nodes (3): TestLastLines(), TestLastLinesEqualCount(), lastLines()

### Community 105 - "internal/cli"
Cohesion: 0.15
Nodes (13): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, euid_test.go, examples_test.go (+5 more)

### Community 106 - "desiredWatch"
Cohesion: 0.14
Nodes (14): desiredWatch(), Cluster, splitWatch(), subtractWatch(), TestDesiredWatchAppendsTheBrokerNamespace(), TestSubtractWatchEmptyRemainingMeansDelete(), TestSubtractWatchPreservesOrder(), TestUnionWatchKeepsAnotherEnvFilesNamespace() (+6 more)

### Community 107 - "Compose"
Cohesion: 0.13
Nodes (16): Compose(), composeEscape(), ComposeProject(), ContainerSecrets(), containerSecretSpecs(), ContainerSecret, secretFilePath(), SecretPreflight() (+8 more)

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "Config"
Cohesion: 0.07
Nodes (19): AdditionalUser, Image, Node, Redundancy, SEMP, TLS, atoiPrefix(), Config (+11 more)

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "GenOperator"
Cohesion: 0.16
Nodes (20): TestValidateSparseConfigExplainsItself(), GenOperator(), joinYAMLDocs(), nonEmpty(), operatorProbes(), RenderOperator(), splitAfterNamespace(), splitOperatorBundle() (+12 more)

### Community 112 - "capRunner"
Cohesion: 0.09
Nodes (30): capCall, capRunner, New(), NewManager(), callIndex(), TestManagerLogsCLIShell(), TestManagerNilSinks(), TestManagerPrepHostRootlessUsesUnshareChown() (+22 more)

### Community 113 - "internal/config"
Cohesion: 0.12
Nodes (17): adminsecret_test.go, artifactvalues_test.go, command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go (+9 more)

### Community 114 - "newEchoMgr"
Cohesion: 0.14
Nodes (14): bytes.Buffer, Manager, newEchoMgr(), TestCheckEnvReportsBaseDirOnlyWhenSet(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployPodmanDryRunHidesSecretBytes(), TestManagerDockerCheckProbesCompose() (+6 more)

### Community 115 - "HARoles"
Cohesion: 0.11
Nodes (14): BrokerPodSuffixShape(), Cluster, TestOperatorAdminSecretNameMatchesTheBrokerNames(), Cluster, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes(), HARoles() (+6 more)

### Community 116 - "Quadlet"
Cohesion: 0.11
Nodes (23): ContainerNoFile(), HealthCheck, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), EnvPairs(), escapePercent(), groupKey(), healthCmd() (+15 more)

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - ".stateRows"
Cohesion: 0.16
Nodes (14): containerState, healthState, decodeInspect(), Manager, orUnknown(), printable(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown() (+6 more)

### Community 119 - "replicationVPNLines"
Cohesion: 0.26
Nodes (13): mateConvergenceShutdowns(), replicationVPNLines(), siteAEntry(), TestMateConvergenceShutdownsIsDeterministic(), TestNothingEverShutsDownTheVPNItself(), TestPlannedRolesNamesEveryListedVPN(), TestReplicationVPNLinesEnablesListedAndDisablesTheRest(), TestReplicationVPNLinesIsDeterministic() (+5 more)

### Community 120 - "Convert"
Cohesion: 0.40
Nodes (5): Result, Convert(), TestConvertUnterminatedArray(), TestGeneratedHeader(), TestGeneratedHeaderSanitisesSource()

### Community 122 - "newTestOps"
Cohesion: 0.03
Nodes (138): cliRunNames(), Ops, hasCall(), matchCLI(), newTestOps(), outputForRole(), ranContains(), TestCountContains() (+130 more)

### Community 123 - ".validateContainerArtifactValues"
Cohesion: 0.22
Nodes (6): Config, portOrRange(), TestValidContainerPort(), validContainerPort(), validHealthDuration(), validDNSLabel()

### Community 124 - "newBlock"
Cohesion: 0.24
Nodes (10): token, firstQuoted(), newBlock(), TestNewBlockMessageSpoolQualifierHasNoOwnName(), TestNewBlockOpener(), TestNewBlockVPNNameWithSpaces(), TestNewBlockVPNQualifierWithSpaces(), tokenizeOpener() (+2 more)

### Community 125 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 126 - ".applyScalingTierDefaults"
Cohesion: 0.15
Nodes (11): scalingTier, containerMem(), cpuSetCount(), cpuSetRange(), Config, TestContainerMem(), TestCPUSetCount(), TestCPUSetRange() (+3 more)

### Community 127 - "watchCluster"
Cohesion: 0.29
Nodes (7): deployJSON(), Cluster, TestOperatorReleaseNarrowsInsteadOfRemoving(), TestReconcileWatchDecidesWhatToApply(), TestReconcileWatchFlagsAWideningAsAQuestion(), TestSetWatchRefusesAnEmptyList(), watchCluster()

### Community 128 - "logArgs"
Cohesion: 0.67
Nodes (4): logArgs(), App, logArgs2(), TestLogArgsBuildsOneSetForBothPlatforms()

### Community 129 - ".decodeScalingEntry"
Cohesion: 0.31
Nodes (7): scalingKey, scalingSpelling, Scaling, Scaling, yaml.Node, scalingKeyIndex(), scalingKeyList()

### Community 131 - ".preflightOne"
Cohesion: 0.50
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "ReplVia"
Cohesion: 0.22
Nodes (8): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, sortedKeys(), validateReplEndpoints(), validateReplVia(), validateReplViaSEMP()

### Community 139 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 141 - "CheckHostPath"
Cohesion: 0.33
Nodes (5): CanonicalDuration(), TestCanonicalDuration(), CheckHostPath(), TestCheckHostPathAccepts(), TestCheckHostPathRejects()

## Knowledge Gaps
- **237 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+232 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 357 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **15 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `captureStdout`, `testing.T`, `Manager`, `CheckCommand`, `CheckHostPath`, `convert_test.go`, `manager_test.go`, `parse`, `Config`, `Role`, `commands.go`, `config_test.go`, `emitYAML`, `Command`, `limits_test.go`, `render_test.go`, `step`, `.applyContainerDefaults`, `Config`, `Load`, `container/secrets_test.go`, `Compose`, `capRunner`, `newEchoMgr`, `Convert`, `Ops`, `.validateContainerArtifactValues`, `.applyScalingTierDefaults`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `captureStdout`, `context.Context`, `testing.T`, `Manager`, `K8sConfig`, `.Run`, `NewCluster`, `GenSecrets`, `convert_test.go`, `manager_test.go`, `AdminSecret`, `eqArgs`, `Platform`, `ServerCertBundle`, `ResolveDomainCerts`, `Role`, `recRunner`, `ReplSite`, `localCfg`, `config_test.go`, `.configRows`, `strings.Builder`, `k8s/matechannel_test.go`, `render_test.go`, `step`, `.applyContainerDefaults`, `haCfg`, `container/secrets_test.go`, `desiredWatch`, `Compose`, `GenOperator`, `capRunner`, `newEchoMgr`, `HARoles`, `Quadlet`, `Ops`, `newTestOps`?**
  _High betweenness centrality (0.018) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `Manager`, `K8sConfig`, `Set`, `renderDriver`, `ImportPlan`, `Config`, `App`, `.releaseToBackup`, `config_test.go`, `BrokerType`, `scripts.go`, `MateConfig`, `newTestMate`, `.ConfigureReplication`, `k8sOps`, `New`, `step`, `ParseRole`, `Config`, `HARoles`, `Ops`, `newTestOps`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Are the 49 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 49 INFERRED edges - model-reasoned connections that need verification._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _237 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `captureStdout` be split into smaller, more focused modules?**
  _Cohesion score 0.09371980676328502 - nodes in this community are weakly interconnected._
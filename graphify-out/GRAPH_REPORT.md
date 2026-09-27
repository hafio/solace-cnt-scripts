# Graph Report - solace-cnt-scripts  (2026-09-28)

## Corpus Check
- 196 files · ~604,926 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 3808 nodes · 15991 edges · 144 communities (130 shown, 14 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 2203 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `96ab6f34`
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
- runRootWith
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
- GenOperator
- ResolveDomainCerts
- Role
- App
- haCfg
- ParseBlocks
- completion_test.go
- diff_test.go
- runPlatform
- inject_test.go
- .releaseToBackup
- platformOps
- Echo
- k8s/inspect.go
- ReplSite
- newEchoMgr
- config_test.go
- Platforms
- newRootCmd
- exportconfig_test.go
- The eight rules
- Operations
- Test catalogue
- commands.go
- MateChannel
- scripts.go
- .rpc
- confirm.go
- .configRows
- rootless_test.go
- strings.Builder
- output_test.go
- Configuration
- Load
- newTestMate
- Command reference
- Developer guide
- replApp
- emitYAML
- age
- internal/k8s
- parseBody
- exit.go
- ExitCode
- Command
- newTeardownApplyFixture
- MateConfig
- go_pkg_solace_internal_abbrev
- BrokerType
- k8s/matechannel_test.go
- limits_test.go
- runExport
- Troubleshooting
- k8s/inspect_test.go
- ParseShowReplication
- operatorversion_test.go
- render_test.go
- vulnjudge/main.go
- RuleFor
- New
- .applyContainerDefaults
- Config
- .resolveSecretRefs
- ParseRole
- replConfig
- Abbreviations
- .ConfigureReplication
- .checkUserManagerLimits
- .Preflight
- Get
- certFixture
- reportCluster
- internal/cli
- desiredWatch
- RenderDiffResult
- What `import-config` applies
- Config
- internal/broker
- ParseVPNReplication
- capRunner
- internal/config
- Removing a broker: what stays, what goes
- HARoles
- parseKV
- internal/container
- .stateRows
- newBlock
- Convert
- Ops
- newTestOps
- .validateContainerArtifactValues
- watchCluster
- logArgs
- .applyScalingTierDefaults
- imageFromDeployment
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
1. `newTestOps()` - 130 edges
2. `ctrCfg()` - 130 edges
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
- `assertFencesBalanced()` --calls--> `markerVerb()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/annotate.go
- `applyMetaField()` --calls--> `BrokerType`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go
- `StripMarkers()` --calls--> `splitLines()`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go
- `sectionBeginMarker()` --calls--> `RuleFor()`  [INFERRED]
  internal/broker/annotate.go → internal/broker/sections.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (144 total, 14 thin omitted)

### Community 0 - "bg"
Cohesion: 0.13
Nodes (49): warn(), wantEnable(), bg(), App, k8sAdminPassword(), k8sCluster(), k8sContext(), k8sLogin() (+41 more)

### Community 1 - "captureStdout"
Cohesion: 0.13
Nodes (30): opCall, captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch() (+22 more)

### Community 2 - "Sink"
Cohesion: 0.20
Nodes (5): solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), columnWidths(), Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.05
Nodes (19): Exec, context.Context, os/exec.Cmd, Manager, idMapCovers(), origin(), runUserIDs(), EnvRunner (+11 more)

### Community 4 - "cli_test.go"
Cohesion: 0.08
Nodes (53): os.File, allowRuntime(), capture(), captureStderr(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone() (+45 more)

### Community 5 - "testing.T"
Cohesion: 0.02
Nodes (186): testing.T, TestLastLines(), TestLastLinesEqualCount(), TestRouterNameReadsTheCaptureHeader(), bridgeHostPort(), sempPort(), TestSempPortBridgeFallsBackToPlaintext(), TestSempPortBridgePrefersTLS() (+178 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (8): certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing()

### Community 8 - "K8sConfig"
Cohesion: 0.18
Nodes (6): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig

### Community 9 - "runRootWith"
Cohesion: 0.10
Nodes (39): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+31 more)

### Community 10 - "CheckCommand"
Cohesion: 0.13
Nodes (24): commandRules, guardedCmd, decodeCommand(), TestCommandUnmarshal(), checkBinary(), CheckCommand(), checkFlagShape(), checkToken() (+16 more)

### Community 11 - ".Run"
Cohesion: 0.16
Nodes (24): testing.M, runCtr(), TestAnnounceCommandsNamesResolvedBinaries(), TestCtrConfigDryRun(), TestCtrDiagnosticsDryRun(), TestCtrErrorPaths(), TestCtrExecCLIPathSeparator(), TestCtrLoginOutcomes() (+16 more)

### Community 12 - "NewCluster"
Cohesion: 0.07
Nodes (62): TestCheckStopsProbingWhenUnreachable(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+54 more)

### Community 13 - "GenSecrets"
Cohesion: 0.06
Nodes (40): GenBroker(), GenSecrets(), joinManifests(), adminCfg(), secretNamesIn(), TestAdminSecretStatesEmitTheirArtifacts(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree() (+32 more)

### Community 14 - "convert_test.go"
Cohesion: 0.18
Nodes (31): checkGolden(), convertOK(), hasWarning(), strictDecode(), TestConvertAdminSecretAlias(), TestConvertAdminUserIsDroppedOnEveryPlatform(), TestConvertBadBooleanWarns(), TestConvertBadNumberWarns() (+23 more)

### Community 15 - "transform_test.go"
Cohesion: 0.09
Nodes (52): Block, TargetState, regexp.Regexp, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects() (+44 more)

### Community 16 - "manager_test.go"
Cohesion: 0.07
Nodes (98): GuardPodmanEUID(), inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman() (+90 more)

### Community 17 - "github.com/spf13/cobra.Command"
Cohesion: 0.13
Nodes (27): shorthand, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, github.com/spf13/pflag.FlagSet, renderAbbrevDocs(), treeShorthands(), writeShorthandTable(), writeUnabbreviated() (+19 more)

### Community 18 - "AdminSecret"
Cohesion: 0.18
Nodes (13): AdditionalUsersSecret(), AdminSecret(), decodeDataValue(), TestAdditionalUsersSecretCarriesBothHalves(), TestAdditionalUsersSecretIsAbsentWithNoUsers(), TestAdditionalUsersSecretRefusesAnEmptyPassword(), TestAdditionalUsersStayOutOfTheCredentialsSecret(), TestAdminSecretCarriesThePSKOnlyWhenSet() (+5 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "parse"
Cohesion: 0.10
Nodes (21): segment, vars, countMarkers(), kubeCommand(), resolvePlatform(), TestParseArrayElementsPreserveQuoting(), TestParseAssignmentForms(), TestParseCRLF() (+13 more)

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
Cohesion: 0.06
Nodes (73): TestLocalMateRunsThroughOps(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK(), TestBackupRevertActivitySuccess() (+65 more)

### Community 25 - "runner_test.go"
Cohesion: 0.11
Nodes (28): TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), MaskEnv(), quoteTok(), captureStdout(), helperCommand() (+20 more)

### Community 26 - "renderDriver"
Cohesion: 0.13
Nodes (18): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), TestKeywordPatternEscapesEachPhrase() (+10 more)

### Community 27 - "eqArgs"
Cohesion: 0.11
Nodes (31): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), saCfg() (+23 more)

### Community 28 - "ImportPlan"
Cohesion: 0.09
Nodes (28): ImportResult, PlannedSection, Region, isPragma(), checkProvenance(), checkSameType(), defaultVPNFirst(), describeScope() (+20 more)

### Community 29 - "Platform"
Cohesion: 0.17
Nodes (23): noRolePositional(), App, newConvertCmd(), runConvert(), usagef(), checkAllowCommand(), checkFlagPlatforms(), commandPlatforms() (+15 more)

### Community 31 - "Config"
Cohesion: 0.09
Nodes (18): keyValueEntries, TestValidateContainerArtifactValues(), TestRoleNames(), RoleNames(), checkCredentialChars(), foldToEnvVar(), Config, invertedCPUSetRange() (+10 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "GenOperator"
Cohesion: 0.15
Nodes (21): GenOperator(), joinYAMLDocs(), nonEmpty(), operatorProbes(), RenderOperator(), renderOperatorNS(), splitAfterNamespace(), splitOperatorBundle() (+13 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (45): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, os.DirEntry, os.FileInfo, os.FileMode (+37 more)

### Community 35 - "Role"
Cohesion: 0.06
Nodes (22): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+14 more)

### Community 36 - "App"
Cohesion: 0.10
Nodes (43): TestCtrManagerConfirmWiring(), confirmAction(), childExit(), TestChildExitKeepsItsMessage(), TestExitCodeIsNeverNegative(), wantRemove(), containerRenderRole(), containerRole() (+35 more)

### Community 37 - "haCfg"
Cohesion: 0.07
Nodes (33): AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestReadAdminPasswordArgvAndDecoding(), TestReadAdminPasswordGuardsTheCommand(), TestReadAdminPasswordNamesTheNextStep() (+25 more)

### Community 38 - "ParseBlocks"
Cohesion: 0.10
Nodes (27): Annotate(), sectionBeginMarker(), assertFencesBalanced(), TestAnnotateIsDeterministic(), TestAnnotateRegionAndSectionFencesAreBalanced(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), TestSectionBeginMarkerAdvisoryDisposition() (+19 more)

### Community 39 - "completion_test.go"
Cohesion: 0.18
Nodes (20): driveBash(), runComplete(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion(), TestCompletionHelpKeepsTheLoadingInstructions(), TestCompletionHelpStillWorks() (+12 more)

### Community 40 - "diff_test.go"
Cohesion: 0.15
Nodes (25): blockKey, mergedBlock, mergedLine, ancestorReported(), DiffBlocks(), indexByIdentity(), keyOf(), keySet() (+17 more)

### Community 41 - "runPlatform"
Cohesion: 0.29
Nodes (12): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+4 more)

### Community 42 - "inject_test.go"
Cohesion: 0.15
Nodes (27): lineRole, serviceLine, shutdownStyle, svcKey, classifyServiceRest(), containsPortCommand(), describeKey(), fieldsUnquoted() (+19 more)

### Community 43 - ".releaseToBackup"
Cohesion: 0.14
Nodes (15): countContains(), field(), assertLeaderScript(), noReleaseActivityScript(), releaseActivityScript(), TestAssertLeaderScript(), TestZipConfigsScript(), zipConfigsScript() (+7 more)

### Community 44 - "platformOps"
Cohesion: 0.19
Nodes (31): opFunc, roleOpFunc, addPodFlag(), newBrokerCopyCmd(), newBrokerDeployCmd(), newBrokerExecCmd(), newBrokerGenerateCmd(), newBrokerLogsCmd() (+23 more)

### Community 45 - "Echo"
Cohesion: 0.26
Nodes (4): interactiveFailRunner, Echo, Quote(), TestQuote()

### Community 46 - "k8s/inspect.go"
Cohesion: 0.19
Nodes (18): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), ownedPods(), brokerList, brokerStatus (+10 more)

### Community 47 - "ReplSite"
Cohesion: 0.14
Nodes (22): Replication, mateConvergenceShutdowns(), missingListedVPNs(), PlannedRoles(), replicationVPNLines(), RoleAtSite(), siteAEntry(), TestMateConvergenceShutdownsIsDeterministic() (+14 more)

### Community 48 - "newEchoMgr"
Cohesion: 0.10
Nodes (20): bytes.Buffer, NewManager(), Manager, newEchoMgr(), TestCheckEnvReportsBaseDirOnlyWhenSet(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployPodmanDryRunHidesSecretBytes() (+12 more)

### Community 49 - "config_test.go"
Cohesion: 0.20
Nodes (45): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, span, go_pkg_bytes, go_pkg_context, go_pkg_crypto_sha256 (+37 more)

### Community 50 - "Platforms"
Cohesion: 0.10
Nodes (21): TestPlatformConstantsMatchSchemaSections(), guardConfig(), TestExecBinariesCoversEveryPlatform(), TestGuardConfigIsValid(), IsAbsHostPath(), TestContainerCertRequiresKey(), TestContainerHostDirsExpandATilde(), TestDataDirMustBeAbsolute() (+13 more)

### Community 51 - "newRootCmd"
Cohesion: 0.10
Nodes (26): TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing() (+18 more)

### Community 52 - "exportconfig_test.go"
Cohesion: 0.16
Nodes (19): opRunner, go_pkg_runtime, exportconfigDriverOutput(), exportconfigTransportOutput(), exportconfigVPNListOutput(), App, newExportconfigRunner(), newExportconfigRunnerRejectingChunk() (+11 more)

### Community 53 - "The eight rules"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 54 - "Operations"
Cohesion: 0.08
Nodes (26): Bringing up a fresh cluster, Broker scope is a fixed classification, Configuring a site, Data replication, Docker and Podman mechanics, Exit codes, Exporting and importing configuration, Extra CLI users differ by platform (+18 more)

### Community 55 - "Test catalogue"
Cohesion: 0.08
Nodes (24): abbrev_test.go, convert_test.go, Coverage, examples_test.go, Fixtures and doubles, Injectable seams, internal/abbrev, internal/convert (+16 more)

### Community 56 - "commands.go"
Cohesion: 0.21
Nodes (38): addCommands(), addLogFlags(), App, group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerConfigureCmd(), newBrokerRemoveCmd() (+30 more)

### Community 57 - "MateChannel"
Cohesion: 0.07
Nodes (27): AdminState, QueueState, ReplRole, scriptedMate, VPNRepl, MateChannel, rejectionIn(), sortedKeys() (+19 more)

### Community 58 - "scripts.go"
Cohesion: 0.07
Nodes (37): showCmd, validCLILine(), Ops, currentConfigScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript() (+29 more)

### Community 59 - ".rpc"
Cohesion: 0.10
Nodes (20): backupTarget, Credential, sempMate, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigLine(), Ops, httpBody() (+12 more)

### Community 60 - "confirm.go"
Cohesion: 0.14
Nodes (28): exportconfigReadCounter, layer, go_pkg_bufio, io.Reader, io.Writer, TestConfirmDeleteShortcut(), TestConfirmNonTTY(), TestPromptAsksWhenStdinIsNotATTY() (+20 more)

### Community 61 - ".configRows"
Cohesion: 0.13
Nodes (24): orNone(), orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminSecretRow(), containsString() (+16 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.15
Nodes (39): failOnCall(), fakeEnv(), Manager, healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+31 more)

### Community 63 - "strings.Builder"
Cohesion: 0.11
Nodes (28): WeightedNodeTerm, strings.Builder, mdRow(), writeAbbrevTable(), writeResolutionNotes(), writeReadingNotes(), LoadBalancer, NodeAffinity (+20 more)

### Community 64 - "output_test.go"
Cohesion: 0.16
Nodes (17): NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLevelOrderKeepsSkipBelowFail(), TestLineAddsNoPrefix() (+9 more)

### Community 65 - "Configuration"
Cohesion: 0.14
Nodes (14): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+6 more)

### Community 66 - "Load"
Cohesion: 0.09
Nodes (41): TestServerCertBundleOrder(), TestServerCertBundleRefusesAKeyBearingCert(), TestServerCertBundleReportsAnUnreadableFile(), TestServerCertBundleRequiresBothHalves(), ServerCertBundle(), minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint() (+33 more)

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
Cohesion: 0.24
Nodes (8): doc, boolOf(), commentSafe(), emitYAML(), joinDomainCertPath(), redundancy(), scalar(), TestScalarQuoting()

### Community 72 - "age"
Cohesion: 0.21
Nodes (8): time.Time, age(), roleRank(), Cluster, operatorRunningImage(), ownsPod(), TestOperatorRunningImageWithNoContainers(), TestOwnsPodFallsBackToTheNameInfix()

### Community 73 - "internal/k8s"
Cohesion: 0.10
Nodes (20): adminsecret_test.go, check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go (+12 more)

### Community 74 - "parseBody"
Cohesion: 0.07
Nodes (34): Omission, applyMetaField(), Capture, isMarker(), markerRegion(), markerSection(), markerVerb(), readMarker() (+26 more)

### Community 75 - "exit.go"
Cohesion: 0.12
Nodes (16): childExitError, usageError, github.com/spf13/cobra.PositionalArgs, App, newExamplesCmd(), runExample(), asUsage(), childStatus() (+8 more)

### Community 76 - "ExitCode"
Cohesion: 0.15
Nodes (14): runFailRunner, TestExecute(), ExitCode(), childStatusError(), k8sEnv(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitStatusIsScopedToInteractiveSessions(), TestExitCodeContract() (+6 more)

### Community 77 - "Command"
Cohesion: 0.18
Nodes (10): cmdField, TestCommandArgsDoesNotAliasCommand(), TestValidateProbeCommandAccepts(), TestValidateProbeCommandRejects(), Command, Config, guardCommandOf(), setGuardCommand() (+2 more)

### Community 78 - "newTeardownApplyFixture"
Cohesion: 0.18
Nodes (14): planSections(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, newTeardownApplyFixture(), TestBuildChunksCreatesVPNsBeforeBrokerSections() (+6 more)

### Community 79 - "MateConfig"
Cohesion: 0.16
Nodes (14): MateConfig, cliTransport(), labelValue(), parseHostPort(), parseShowReplicationAppliance(), parseShowReplicationSoftware(), renderMateAppliance(), renderMateSoftware() (+6 more)

### Community 81 - "BrokerType"
Cohesion: 0.33
Nodes (5): BrokerType, cliMate, bannerType(), showReplicationScript(), TestTranscriptBannerType()

### Community 82 - "k8s/matechannel_test.go"
Cohesion: 0.21
Nodes (26): CLIArg(), CLIScriptPath(), execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf() (+18 more)

### Community 83 - "limits_test.go"
Cohesion: 0.13
Nodes (27): missingControllers(), Manager, limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager() (+19 more)

### Community 84 - "runExport"
Cohesion: 0.13
Nodes (15): confirmImport(), App, pluralVPN(), runExport(), runImport(), TestPluralVPNWording(), emitOrWrite(), App (+7 more)

### Community 85 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 86 - "k8s/inspect_test.go"
Cohesion: 0.15
Nodes (19): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+11 more)

### Community 87 - "ParseShowReplication"
Cohesion: 0.13
Nodes (27): TestMateChannelShowReplication(), containsEndpoint(), normTransport(), ParseShowReplication(), RenderMate(), RenderMateRemovals(), SameEndpoints(), loadShowReplication() (+19 more)

### Community 88 - "operatorversion_test.go"
Cohesion: 0.21
Nodes (13): compareVersions(), imageTag(), operatorVersionWarning(), parseVersion(), operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsSilentInAPreview() (+5 more)

### Community 89 - "render_test.go"
Cohesion: 0.06
Nodes (79): opCtrGenArtifact(), ContainerNoFile(), HealthCheck, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), boolStr(), BrokerCR(), Compose() (+71 more)

### Community 90 - "vulnjudge/main.go"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.08
Nodes (37): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+29 more)

### Community 92 - "New"
Cohesion: 0.09
Nodes (30): mateChannelFunc, bufio.Reader, TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), ReplicationConfigResult, App, lineSink() (+22 more)

### Community 93 - ".applyContainerDefaults"
Cohesion: 0.13
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

### Community 99 - ".ConfigureReplication"
Cohesion: 0.24
Nodes (8): Ops, isRunCLIRejection(), newlineIf(), replPhase1Rejected(), replPhase2Rejected(), replTransportFailure(), replVPNList(), showVPNReplicationScript()

### Community 100 - ".checkUserManagerLimits"
Cohesion: 0.29
Nodes (6): delegateSetting(), Manager, limitText(), parseLimit(), parseUnitProps(), TestParseLimit()

### Community 102 - "Get"
Cohesion: 0.24
Nodes (14): Example, go_pkg_embed, TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), All(), Get(), List(), Names() (+6 more)

### Community 103 - "certFixture"
Cohesion: 0.13
Nodes (19): ResolveSecretValues(), bundleHash(), certFixture(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeWritingTheComposeFile(), TestDeployDockerPassesTheBundleAsEnvNotArgv(), TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(), TestDockerCertChangedEdges() (+11 more)

### Community 104 - "reportCluster"
Cohesion: 0.14
Nodes (14): Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList(), TestBrokerReportWithNoBrokerDeployed() (+6 more)

### Community 105 - "internal/cli"
Cohesion: 0.15
Nodes (13): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, euid_test.go, examples_test.go (+5 more)

### Community 106 - "desiredWatch"
Cohesion: 0.14
Nodes (14): desiredWatch(), Cluster, splitWatch(), subtractWatch(), TestDesiredWatchAppendsTheBrokerNamespace(), TestSubtractWatchEmptyRemainingMeansDelete(), TestSubtractWatchPreservesOrder(), TestUnionWatchKeepsAnotherEnvFilesNamespace() (+6 more)

### Community 107 - "RenderDiffResult"
Cohesion: 0.23
Nodes (11): BlockDiff, DiffResult, blockLabel(), diffRow(), orDash(), RenderDiffResult(), reportBlockLines(), TestDiffBlockLabelFormatsIdentityForTheReport() (+3 more)

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "Config"
Cohesion: 0.09
Nodes (14): AdditionalUser, Image, Node, Redundancy, SEMP, TLS, atoiPrefix(), Config (+6 more)

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "ParseVPNReplication"
Cohesion: 0.10
Nodes (22): colSpan, ValidVPNName(), validVPNName(), dashSpans(), flagByte(), gutterClear(), ParseVPNReplication(), sliceSpan() (+14 more)

### Community 112 - "capRunner"
Cohesion: 0.10
Nodes (26): capCall, capRunner, New(), callIndex(), TestManagerLogsCLIShell(), TestManagerPrepHostRootlessUsesUnshareChown(), TestPreflightRunsBeforeAnything(), TestCtrRuntimeDefaultArgvUnchanged() (+18 more)

### Community 113 - "internal/config"
Cohesion: 0.12
Nodes (17): adminsecret_test.go, artifactvalues_test.go, command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go (+9 more)

### Community 114 - "Removing a broker: what stays, what goes"
Cohesion: 0.40
Nodes (5): Every destructive command confirms, Removing a broker: what stays, what goes, Removing the operator does not always remove it, The layer flag raises the question, The namespace is only offered when it is empty

### Community 115 - "HARoles"
Cohesion: 0.11
Nodes (14): BrokerPodSuffixShape(), Cluster, TestOperatorAdminSecretNameMatchesTheBrokerNames(), Cluster, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes(), HARoles() (+6 more)

### Community 116 - "parseKV"
Cohesion: 0.17
Nodes (12): parseKV(), renderKV(), splitKV(), TestMarkerValueSurvivesAnEmbeddedQuote(), TestParseKVEmptyPayload(), TestParseKVIgnoresTokenWithNoEquals(), TestParseKVMultipleTokens(), TestParseKVQuotedValueKeepsSpaces() (+4 more)

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - ".stateRows"
Cohesion: 0.16
Nodes (14): containerState, healthState, decodeInspect(), Manager, orUnknown(), printable(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown() (+6 more)

### Community 119 - "newBlock"
Cohesion: 0.28
Nodes (9): token, firstQuoted(), newBlock(), TestNewBlockMessageSpoolQualifierHasNoOwnName(), TestNewBlockOpener(), TestNewBlockVPNNameWithSpaces(), TestNewBlockVPNQualifierWithSpaces(), tokenizeOpener() (+1 more)

### Community 120 - "Convert"
Cohesion: 0.29
Nodes (7): Result, DecodeStrict(), Config, Convert(), TestConvertUnterminatedArray(), TestGeneratedHeader(), validateOutput()

### Community 121 - "Ops"
Cohesion: 0.13
Nodes (9): time.Duration, Ops, runCLISkeleton(), validName(), shQuote(), TestShQuoteHandlesASingleQuote(), shellScriptPath(), BaseName() (+1 more)

### Community 122 - "newTestOps"
Cohesion: 0.03
Nodes (144): cliRunNames(), Ops, hasCall(), matchCLI(), newTestOps(), outputForRole(), ranContains(), TestCountContains() (+136 more)

### Community 123 - ".validateContainerArtifactValues"
Cohesion: 0.22
Nodes (6): Config, portOrRange(), TestValidContainerPort(), validContainerPort(), validHealthDuration(), validDNSLabel()

### Community 124 - "watchCluster"
Cohesion: 0.29
Nodes (7): deployJSON(), Cluster, TestOperatorReleaseNarrowsInsteadOfRemoving(), TestReconcileWatchDecidesWhatToApply(), TestReconcileWatchFlagsAWideningAsAQuestion(), TestSetWatchRefusesAnEmptyList(), watchCluster()

### Community 125 - "logArgs"
Cohesion: 0.67
Nodes (4): logArgs(), App, logArgs2(), TestLogArgsBuildsOneSetForBothPlatforms()

### Community 126 - ".applyScalingTierDefaults"
Cohesion: 0.15
Nodes (11): scalingTier, containerMem(), cpuSetCount(), cpuSetRange(), Config, TestContainerMem(), TestCPUSetCount(), TestCPUSetRange() (+3 more)

### Community 128 - "imageFromDeployment"
Cohesion: 0.50
Nodes (4): imageFromDeployment(), operatorItem(), TestAmbiguousOperatorIsNotFoldedIntoNotInstalled(), TestFindOperatorDeploymentIsScopedByNamespace()

### Community 129 - ".decodeScalingEntry"
Cohesion: 0.31
Nodes (7): scalingKey, scalingSpelling, Scaling, Scaling, yaml.Node, scalingKeyIndex(), scalingKeyList()

### Community 131 - ".preflightOne"
Cohesion: 0.50
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "ReplVia"
Cohesion: 0.29
Nodes (7): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, viaDescription(), validateReplVia(), validateReplViaSEMP()

### Community 139 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 141 - "CheckHostPath"
Cohesion: 0.40
Nodes (5): CanonicalDuration(), TestCanonicalDuration(), CheckHostPath(), TestCheckHostPathAccepts(), TestCheckHostPathRejects()

## Knowledge Gaps
- **237 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+232 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 357 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **14 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `captureStdout`, `testing.T`, `Manager`, `CheckCommand`, `convert_test.go`, `manager_test.go`, `parse`, `Config`, `Role`, `platformOps`, `newEchoMgr`, `config_test.go`, `Platforms`, `commands.go`, `Load`, `emitYAML`, `Command`, `limits_test.go`, `render_test.go`, `New`, `.applyContainerDefaults`, `Config`, `certFixture`, `capRunner`, `Convert`, `Ops`, `.validateContainerArtifactValues`, `.applyScalingTierDefaults`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `captureStdout`, `context.Context`, `cli_test.go`, `testing.T`, `Manager`, `K8sConfig`, `NewCluster`, `GenSecrets`, `convert_test.go`, `manager_test.go`, `AdminSecret`, `verify_local_test.go`, `eqArgs`, `GenOperator`, `ResolveDomainCerts`, `Role`, `haCfg`, `ReplSite`, `newEchoMgr`, `config_test.go`, `.configRows`, `strings.Builder`, `Load`, `k8s/matechannel_test.go`, `render_test.go`, `New`, `.applyContainerDefaults`, `certFixture`, `desiredWatch`, `capRunner`, `HARoles`, `Ops`, `newTestOps`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `Manager`, `K8sConfig`, `Set`, `renderDriver`, `ImportPlan`, `Config`, `App`, `haCfg`, `.releaseToBackup`, `config_test.go`, `scripts.go`, `.rpc`, `newTestMate`, `BrokerType`, `k8s/matechannel_test.go`, `runExport`, `New`, `ParseRole`, `.ConfigureReplication`, `HARoles`, `Ops`, `newTestOps`?**
  _High betweenness centrality (0.018) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 48 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 48 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _237 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `bg` be split into smaller, more focused modules?**
  _Cohesion score 0.12755102040816327 - nodes in this community are weakly interconnected._
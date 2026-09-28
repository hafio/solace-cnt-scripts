# Graph Report - solace-cnt-scripts  (2026-09-28)

## Corpus Check
- 196 files · ~606,126 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 3809 nodes · 15994 edges · 153 communities (136 shown, 17 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 2203 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `048bd794`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ops_k8s.go
- captureStdout
- Sink
- context.Context
- runRoot
- testing.T
- Commands
- Manager
- K8sConfig
- allowcommand_test.go
- CheckCommand
- cli_test.go
- NewCluster
- prep_test.go
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
- newTestOps
- ResolveDomainCerts
- Role
- bg
- recRunner
- annotate_test.go
- completion_test.go
- diff.go
- echoRunner
- inject.go
- .releaseToBackup
- commands.go
- Echo
- k8s/inspect.go
- ReplSite
- localCfg
- config_test.go
- ParseBlocks
- newRootCmd
- runRootWith
- The eight rules
- Operations
- Test catalogue
- newOperatorCmd
- MateChannel
- scripts.go
- .rpc
- confirmAction
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
- check.go
- internal/k8s
- blocks.go
- usagef
- TestExecute
- Command
- newTeardownApplyFixture
- Ops
- go_pkg_solace_internal_abbrev
- Manager
- k8s/matechannel_test.go
- limits_test.go
- New
- Troubleshooting
- k8s/inspect_test.go
- .ConfigureReplication
- operatorversion.go
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
- hasCall
- .checkUserManagerLimits
- .Preflight
- Get
- ctrCfg
- .LeaderLocal
- internal/cli
- desiredWatch
- Compose
- What `import-config` applies
- Config
- internal/broker
- scripts_test.go
- capRunner
- internal/config
- Removing a broker: what stays, what goes
- HARoles
- EnvPairs
- internal/container
- .stateRows
- Cluster
- Convert
- Ops
- coverage_test.go
- .validateContainerArtifactValues
- containerArtifacts
- eqArgs
- .applyScalingTierDefaults
- Cluster
- Quadlet
- .decodeScalingEntry
- ContainerSecrets
- k8s/preflight.go
- ReplVia
- Cluster
- parsePort
- normalizeToList
- TestImportOpsExportConfigScopeSelectsCLICommand
- euid_test.go
- pskCfg
- ResolveEnvPath
- .ServerCert
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
- `main()` --calls--> `ExitCode()`  [EXTRACTED]
  main.go → internal/cli/exit.go
- `assertFencesBalanced()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/annotate_test.go → internal/broker/annotate.go
- `parseBody()` --calls--> `readMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `parseHeader()` --calls--> `readMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `applyMetaField()` --calls--> `BrokerType`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (153 total, 17 thin omitted)

### Community 0 - "ops_k8s.go"
Cohesion: 0.14
Nodes (47): domainCANames(), App, k8sAdminPassword(), k8sCluster(), k8sContext(), k8sLogin(), k8sOps(), k8sWhat() (+39 more)

### Community 1 - "captureStdout"
Cohesion: 0.10
Nodes (34): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod(), k8sDeployAllOutputHook(), loadDirect() (+26 more)

### Community 2 - "Sink"
Cohesion: 0.20
Nodes (5): solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), columnWidths(), Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.05
Nodes (18): QueueState, ReplRole, scriptedMate, VPNRepl, Exec, context.Context, os/exec.Cmd, reenableLine() (+10 more)

### Community 4 - "runRoot"
Cohesion: 0.07
Nodes (44): os.File, capture(), captureStderr(), fakeBinaryOnPath(), firstLine(), runRoot(), runStatusStderr(), TestBashEnvGivenToEnvFlag() (+36 more)

### Community 5 - "testing.T"
Cohesion: 0.02
Nodes (207): testing.T, flagByte(), ParseVPNReplication(), TestParseVPNReplication(), TestParseVPNReplicationIgnoresRepeatedHeaders(), TestParseVPNReplicationIgnoresTheEchoedPrompt(), TestParseVPNReplicationReadsColumnOrderFromTheHeader(), TestParseVPNReplicationRejectsUnreadable() (+199 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (10): certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing() (+2 more)

### Community 8 - "K8sConfig"
Cohesion: 0.15
Nodes (8): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, boolStr(), writeSecurity()

### Community 9 - "allowcommand_test.go"
Cohesion: 0.29
Nodes (10): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+2 more)

### Community 10 - "CheckCommand"
Cohesion: 0.15
Nodes (22): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), escalator() (+14 more)

### Community 11 - "cli_test.go"
Cohesion: 0.11
Nodes (42): testing.M, allowRuntime(), runCtr(), runStandalone(), TestAnnounceCommandsNamesResolvedBinaries(), TestCLICommand(), TestConfigStepsDoNotLeakSecrets(), TestConfiguredRouternameSurvivesTheFallback() (+34 more)

### Community 12 - "NewCluster"
Cohesion: 0.06
Nodes (83): TestCheckStopsProbingWhenUnreachable(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+75 more)

### Community 13 - "prep_test.go"
Cohesion: 0.06
Nodes (60): BrokerPodSuffixShape(), AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestOperatorAdminSecretNameMatchesTheBrokerNames(), TestReadAdminPasswordArgvAndDecoding() (+52 more)

### Community 14 - "convert_test.go"
Cohesion: 0.18
Nodes (31): checkGolden(), convertOK(), hasWarning(), strictDecode(), TestConvertAdminSecretAlias(), TestConvertAdminUserIsDroppedOnEveryPlatform(), TestConvertBadBooleanWarns(), TestConvertBadNumberWarns() (+23 more)

### Community 15 - "transform_test.go"
Cohesion: 0.09
Nodes (52): Block, TargetState, regexp.Regexp, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects() (+44 more)

### Community 16 - "manager_test.go"
Cohesion: 0.07
Nodes (81): TestRestartCountUnknownWhenSystemctlFails(), fileExists(), assertMode(), containsStr(), hasCall(), maskedKeys(), newCapMgr(), TestContainerRunningMatchesNameExactly() (+73 more)

### Community 17 - "github.com/spf13/cobra.Command"
Cohesion: 0.12
Nodes (30): github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, argumentLine(), availableSubs(), writeCommand(), addLogFlags(), newPerformExportConfigCmd(), noRolePositional() (+22 more)

### Community 18 - "AdminSecret"
Cohesion: 0.13
Nodes (17): AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), checkGolden(), decodeDataValue(), TestAdditionalUsersSecretCarriesBothHalves(), TestAdditionalUsersSecretIsAbsentWithNoUsers(), TestAdditionalUsersSecretRefusesAnEmptyPassword() (+9 more)

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
Cohesion: 0.13
Nodes (43): cliScriptPath(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn(), TestDetectRoleAddrsError() (+35 more)

### Community 25 - "runner_test.go"
Cohesion: 0.11
Nodes (28): TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), MaskEnv(), quoteTok(), captureStdout(), helperCommand() (+20 more)

### Community 26 - "renderDriver"
Cohesion: 0.10
Nodes (23): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), TestKeywordPatternEscapesEachPhrase() (+15 more)

### Community 27 - "eqArgs"
Cohesion: 0.07
Nodes (42): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), saCfg() (+34 more)

### Community 28 - "ImportPlan"
Cohesion: 0.14
Nodes (20): ImportResult, PlannedSection, isPragma(), defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList() (+12 more)

### Community 29 - "Platform"
Cohesion: 0.13
Nodes (29): checkFlagPlatforms(), commandPlatforms(), declaredList(), App, parsePlatformList(), platformSuffix(), prepare(), promptPlatform() (+21 more)

### Community 31 - "Config"
Cohesion: 0.10
Nodes (17): keyValueEntries, TestValidateContainerArtifactValues(), TestRoleNames(), RoleNames(), checkCredentialChars(), foldToEnvVar(), Config, invertedCPUSetRange() (+9 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "newTestOps"
Cohesion: 0.07
Nodes (56): Ops, newTestOps(), outputForRole(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultUsersRefusesAQuotedVPNName(), TestDisableDefaultVPN() (+48 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (44): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, os.DirEntry, os.FileInfo, os.FileMode (+36 more)

### Community 35 - "Role"
Cohesion: 0.06
Nodes (22): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+14 more)

### Community 36 - "bg"
Cohesion: 0.15
Nodes (38): childExit(), TestChildExitKeepsItsMessage(), wantRemove(), containerRole(), containerWhat(), ctrLogin(), ctrManager(), ctrOps() (+30 more)

### Community 37 - "recRunner"
Cohesion: 0.09
Nodes (18): TestCanIAnswerReadsTheLastLine(), TestExecutorRefusesUnapprovedRuntime(), unapprovedCfg(), NewTransport(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs(), TestTransportUpload() (+10 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.07
Nodes (46): Annotate(), applyMetaField(), Capture, markerSection(), markerVerb(), parseKV(), readMarker(), renderKV() (+38 more)

### Community 39 - "completion_test.go"
Cohesion: 0.19
Nodes (19): driveBash(), runComplete(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion(), TestCompletionHelpStillWorks(), TestCompletionNeedsAShell() (+11 more)

### Community 40 - "diff.go"
Cohesion: 0.11
Nodes (36): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, ancestorReported(), blockLabel(), DiffBlocks() (+28 more)

### Community 41 - "echoRunner"
Cohesion: 0.15
Nodes (21): echoRunner(), App, TestK8sRestartConfirmGate(), TestRemoveBrokerLayerContract(), TestRemoveFlagsCompose(), TestStandaloneRouternameFallsBackToTheHost(), runPlatform(), TestMultiPlatformNonInteractiveIsRefused() (+13 more)

### Community 42 - "inject.go"
Cohesion: 0.17
Nodes (27): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+19 more)

### Community 43 - ".releaseToBackup"
Cohesion: 0.19
Nodes (11): countContains(), field(), TestCountContains(), TestField(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), showRedundancyLocalScript(), activity() (+3 more)

### Community 44 - "commands.go"
Cohesion: 0.21
Nodes (46): opFunc, roleOpFunc, addPodFlag(), App, group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerConfigureCmd() (+38 more)

### Community 45 - "Echo"
Cohesion: 0.22
Nodes (6): interactiveFailRunner, TestExecIsSilentWithoutVerbose(), Echo, NewExec(), Quote(), TestQuote()

### Community 46 - "k8s/inspect.go"
Cohesion: 0.18
Nodes (20): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), operatorRunningImage(), ownedPods(), TestOperatorRunningImageWithNoContainers() (+12 more)

### Community 47 - "ReplSite"
Cohesion: 0.10
Nodes (28): Replication, sortedKeys(), sortedSet(), mateConvergenceShutdowns(), missingListedVPNs(), PlannedRoles(), replicationVPNLines(), RoleAtSite() (+20 more)

### Community 48 - "localCfg"
Cohesion: 0.09
Nodes (36): TestLocalMateRunsThroughOps(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK(), TestBackupRevertActivitySuccess() (+28 more)

### Community 49 - "config_test.go"
Cohesion: 0.21
Nodes (44): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, go_pkg_bytes, go_pkg_context, go_pkg_crypto_sha256, go_pkg_encoding_base64 (+36 more)

### Community 50 - "ParseBlocks"
Cohesion: 0.12
Nodes (19): StripMarkers(), TestStripMarkersKeepsPragmaAndSectionComments(), TestStripMarkersNoMarkersPassthrough(), ParseBlocks(), splitLines(), TestParseBlocksCRLFMatchesLF(), TestParseBlocksIndentedLineWithNoOpenerRefused(), TestParseBlocksMissingEndTerminatorRefused() (+11 more)

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (28): TestAbbreviationDocs(), applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias() (+20 more)

### Community 52 - "runRootWith"
Cohesion: 0.09
Nodes (42): exportconfigReadCounter, runFailRunner, go_pkg_runtime, runRootWith(), TestDeployOperatorNoPromptStaysUnknownFlag(), TestK8sConfigDeleteDomainCertsConfigured(), TestK8sConfigDeleteDomainCertsFromDirs(), TestK8sConfirmDeclined() (+34 more)

### Community 53 - "The eight rules"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 54 - "Operations"
Cohesion: 0.08
Nodes (26): Bringing up a fresh cluster, Broker scope is a fixed classification, Configuring a site, Data replication, Docker and Podman mechanics, Exit codes, Exporting and importing configuration, Extra CLI users differ by platform (+18 more)

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
Cohesion: 0.15
Nodes (25): showCmd, currentConfigScript(), defaultUsersScript(), disableDefaultUsersScript(), enableDefaultUsersScript(), gatherConfigsScript(), noReleaseActivityScript(), productKeyScript() (+17 more)

### Community 59 - ".rpc"
Cohesion: 0.11
Nodes (19): AdminState, backupTarget, Credential, sempMate, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigLine(), Ops (+11 more)

### Community 60 - "confirmAction"
Cohesion: 0.19
Nodes (25): layer, go_pkg_bufio, io.Reader, io.Writer, TestPromptYes(), TestPromptYesNo(), TestStdinCanAnswerClosedFile(), confirmAction() (+17 more)

### Community 61 - ".configRows"
Cohesion: 0.17
Nodes (17): additionalUsersRow(), adminSecretRow(), containsString(), failRow(), Cluster, info(), okRow(), portRows() (+9 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.17
Nodes (36): failOnCall(), fakeEnv(), Manager, healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+28 more)

### Community 63 - "strings.Builder"
Cohesion: 0.08
Nodes (37): shorthand, WeightedNodeTerm, github.com/spf13/pflag.FlagSet, strings.Builder, mdRow(), renderAbbrevDocs(), TestFlagShorthandsAreConsistent(), treeShorthands() (+29 more)

### Community 64 - "output_test.go"
Cohesion: 0.19
Nodes (17): NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLevelOrderKeepsSkipBelowFail(), TestLineAddsNoPrefix() (+9 more)

### Community 65 - "Configuration"
Cohesion: 0.14
Nodes (14): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+6 more)

### Community 66 - "Load"
Cohesion: 0.16
Nodes (25): minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadReadError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors() (+17 more)

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

### Community 72 - "check.go"
Cohesion: 0.19
Nodes (10): orNone(), orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), age(), roleRank(), Cluster (+2 more)

### Community 73 - "internal/k8s"
Cohesion: 0.10
Nodes (20): adminsecret_test.go, check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go (+12 more)

### Community 74 - "blocks.go"
Cohesion: 0.11
Nodes (28): Omission, Region, token, isMarker(), markerRegion(), TestIsMarkerDiscriminatesNamespace(), TestMarkerRegionBeginAndEnd(), brokerTypeFromSchema() (+20 more)

### Community 75 - "usagef"
Cohesion: 0.13
Nodes (17): childExitError, usageError, github.com/spf13/cobra.PositionalArgs, logArgs(), runConvert(), asUsage(), isUsage(), markUsageArgs() (+9 more)

### Community 76 - "TestExecute"
Cohesion: 0.67
Nodes (3): TestExecute(), Execute(), main()

### Community 77 - "Command"
Cohesion: 0.13
Nodes (13): cmdField, guardedCmd, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandUnmarshal(), TestValidateProbeCommandAccepts(), TestValidateProbeCommandRejects(), Command (+5 more)

### Community 78 - "newTeardownApplyFixture"
Cohesion: 0.31
Nodes (9): driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, newTeardownApplyFixture(), TestImportOpsImportApplySeparatesFirstSectionFromMain(), TestImportOpsImportApplyTearsDownExistingVPNBeforeApplyingArtifact() (+1 more)

### Community 79 - "Ops"
Cohesion: 0.18
Nodes (6): validCLILine(), Ops, disableDefaultVPNScript(), enableDefaultVPNScript(), TestDisableDefaultVPNScript(), TestEnableDefaultVPNScript()

### Community 81 - "Manager"
Cohesion: 0.21
Nodes (4): Manager, idMapCovers(), origin(), runUserIDs()

### Community 82 - "k8s/matechannel_test.go"
Cohesion: 0.21
Nodes (26): CLIArg(), CLIScriptPath(), execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf() (+18 more)

### Community 83 - "limits_test.go"
Cohesion: 0.15
Nodes (25): Manager, limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager(), TestCheckLimitsRootlessRefusesUndelegatedControllers() (+17 more)

### Community 84 - "New"
Cohesion: 0.11
Nodes (19): TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), ReplicationConfigResult, lineSink(), progress(), confirmImport(), App (+11 more)

### Community 85 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 86 - "k8s/inspect_test.go"
Cohesion: 0.15
Nodes (19): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+11 more)

### Community 87 - ".ConfigureReplication"
Cohesion: 0.06
Nodes (52): BrokerType, cliMate, MateConfig, bannerType(), TestMateChannelShowReplication(), cliTransport(), containsEndpoint(), labelValue() (+44 more)

### Community 88 - "operatorversion.go"
Cohesion: 0.22
Nodes (10): operatorImage(), compareVersions(), Cluster, imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), operatorItem() (+2 more)

### Community 89 - "render_test.go"
Cohesion: 0.20
Nodes (23): BrokerCR(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets(), TestBrokerCRQuotesTheImageReference(), TestComposeQuotesTheCpuset(), TestContainerOverridesReachArtifact(), TestContainerSecretNamesAreHostScoped() (+15 more)

### Community 90 - "vulnjudge/main.go"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.08
Nodes (37): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+29 more)

### Community 92 - "step"
Cohesion: 0.15
Nodes (18): mateChannelFunc, bufio.Reader, App, step(), confirmReplicationConfig(), App, opConfigureReplication(), opCtrConfigureReplication() (+10 more)

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
Cohesion: 0.29
Nodes (8): containerRenderRole(), detectContainerRole(), roleWord(), TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), ParseRole(), RoleAbbrev()

### Community 97 - "replConfig"
Cohesion: 0.31
Nodes (10): Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationEmptyViaAccepted(), TestValidateReplicationRejects() (+2 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "hasCall"
Cohesion: 0.23
Nodes (16): hasCall(), TestDisableDefaultUsersShowVPNError(), TestReleaseToBackupReleasedTimeout(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), TestConfigureReplicationClosingShowFailureStillReportsTheApply() (+8 more)

### Community 100 - ".checkUserManagerLimits"
Cohesion: 0.23
Nodes (8): delegateSetting(), Manager, limitText(), missingControllers(), parseLimit(), parseUnitProps(), TestMissingControllers(), TestParseLimit()

### Community 102 - "Get"
Cohesion: 0.14
Nodes (25): Example, go_pkg_embed, TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), parseError(), DetectPlatforms(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint() (+17 more)

### Community 103 - "ctrCfg"
Cohesion: 0.06
Nodes (55): bytes.Buffer, inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman() (+47 more)

### Community 104 - ".LeaderLocal"
Cohesion: 0.17
Nodes (10): TestLastLines(), TestLastLinesEqualCount(), assertLeaderScript(), TestAssertLeaderScript(), defaultLocalAddrs(), Ops, hostMatches(), shortHost() (+2 more)

### Community 105 - "internal/cli"
Cohesion: 0.15
Nodes (13): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, euid_test.go, examples_test.go (+5 more)

### Community 106 - "desiredWatch"
Cohesion: 0.14
Nodes (14): desiredWatch(), Cluster, splitWatch(), subtractWatch(), TestDesiredWatchAppendsTheBrokerNamespace(), TestSubtractWatchEmptyRemainingMeansDelete(), TestSubtractWatchPreservesOrder(), TestUnionWatchKeepsAnotherEnvFilesNamespace() (+6 more)

### Community 107 - "Compose"
Cohesion: 0.19
Nodes (10): Compose(), composeEscape(), ComposeProject(), ContainerSecret, secretFilePath(), TestComposeProjectFoldsToComposesGrammar(), TestComposeProjectIsDeclaredNotDerived(), TestMonitorIsSizedForQuorumNotForTheTier() (+2 more)

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "Config"
Cohesion: 0.05
Nodes (42): AdditionalUser, Image, Node, Redundancy, SEMP, TLS, New(), TestNewDefaults() (+34 more)

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "scripts_test.go"
Cohesion: 0.15
Nodes (15): colSpan, dashSpans(), gutterClear(), sliceSpan(), domainCertsScript(), parseVPNNames(), TestDisableDefaultUsersScriptQuoting(), TestDomainCertsScriptSorted() (+7 more)

### Community 113 - "internal/config"
Cohesion: 0.12
Nodes (17): adminsecret_test.go, artifactvalues_test.go, command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go (+9 more)

### Community 114 - "Removing a broker: what stays, what goes"
Cohesion: 0.40
Nodes (5): Every destructive command confirms, Removing a broker: what stays, what goes, Removing the operator does not always remove it, The layer flag raises the question, The namespace is only offered when it is empty

### Community 115 - "HARoles"
Cohesion: 0.21
Nodes (6): Cluster, allCustomMounted(), Cluster, HARoles(), pvcName(), stsName()

### Community 116 - "EnvPairs"
Cohesion: 0.19
Nodes (12): EnvPairs(), groupKey(), itoa(), ServerCertBundlePath(), assertNoCheckoutPath(), envLines(), TestAdditionalUsersReachBothHalves(), TestGolden() (+4 more)

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - ".stateRows"
Cohesion: 0.16
Nodes (14): containerState, healthState, decodeInspect(), Manager, orUnknown(), printable(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown() (+6 more)

### Community 120 - "Convert"
Cohesion: 0.29
Nodes (7): Result, DecodeStrict(), Config, Convert(), TestConvertUnterminatedArray(), TestGeneratedHeader(), validateOutput()

### Community 121 - "Ops"
Cohesion: 0.11
Nodes (11): time.Duration, Ops, runCLISkeleton(), TestValidName(), validName(), shQuote(), TestShQuoteHandlesASingleQuote(), rejectionIn() (+3 more)

### Community 122 - "coverage_test.go"
Cohesion: 0.05
Nodes (57): cliRunNames(), matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError() (+49 more)

### Community 123 - ".validateContainerArtifactValues"
Cohesion: 0.22
Nodes (6): Config, portOrRange(), TestValidContainerPort(), validContainerPort(), validHealthDuration(), validDNSLabel()

### Community 124 - "containerArtifacts"
Cohesion: 0.18
Nodes (11): HealthCheck, healthCmd(), containerArtifacts(), healthCheckFixture(), TestArtifactsCarryNoWideningTokens(), TestArtifactsStateTheirPrivilegePosture(), TestHealthCmdDefaultsToReadiness(), TestQuadletHealthCmdEscapesPercent() (+3 more)

### Community 125 - "eqArgs"
Cohesion: 0.24
Nodes (11): callIndex(), TestManagerLogsCLIShell(), TestManagerPrepHostRootlessUsesUnshareChown(), TestPreflightRunsBeforeAnything(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime(), TestManagerReachableProbesRuntimeThenCompose() (+3 more)

### Community 126 - ".applyScalingTierDefaults"
Cohesion: 0.15
Nodes (11): scalingTier, containerMem(), cpuSetCount(), cpuSetRange(), Config, TestContainerMem(), TestCPUSetCount(), TestCPUSetRange() (+3 more)

### Community 127 - "Cluster"
Cohesion: 0.24
Nodes (3): Cluster, namespaceManifest(), ownedSecretNames()

### Community 128 - "Quadlet"
Cohesion: 0.31
Nodes (8): ContainerNoFile(), NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), escapePercent(), Quadlet(), quadletEscape(), TestQuadletAsksTheServiceAndTheContainerForTheSameLimits(), TestQuadletEscape()

### Community 129 - ".decodeScalingEntry"
Cohesion: 0.31
Nodes (7): scalingKey, scalingSpelling, Scaling, Scaling, yaml.Node, scalingKeyIndex(), scalingKeyList()

### Community 130 - "ContainerSecrets"
Cohesion: 0.22
Nodes (9): ContainerSecrets(), containerSecretSpecs(), SecretPreflight(), TestSecretPreflight(), TestSecretsAndCertDoNotNest(), TestComposeLabelsTheCertificateDigest(), TestFileBackedSecretIsExemptFromSecretPreflight(), TestServerCertIsASecretOnBothEngines() (+1 more)

### Community 131 - "k8s/preflight.go"
Cohesion: 0.44
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "ReplVia"
Cohesion: 0.22
Nodes (8): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, sortedKeys(), validateReplEndpoints(), validateReplVia(), validateReplViaSEMP()

### Community 134 - "parsePort"
Cohesion: 0.25
Nodes (8): LoadBalancer, cut(), parsePort(), splitUser(), TestParsePort(), writeKeyValueEntry(), writeLBAnnotations(), portSpec

### Community 135 - "normalizeToList"
Cohesion: 0.29
Nodes (4): time.Time, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

### Community 136 - "TestImportOpsExportConfigScopeSelectsCLICommand"
Cohesion: 0.33
Nodes (6): minimalCapture(), targetVPNCapture(), TestImportOpsExportConfigScopeSelectsCLICommand(), TestImportOpsImportPlanRefusesCrossTypeAndUnknownType(), TestImportOpsImportPlanRefusesRedactedArtifact(), TestImportOpsImportPlanSplitsExistingAndNewVPNs()

### Community 137 - "euid_test.go"
Cohesion: 0.67
Nodes (5): runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand(), TestPodmanEUIDGuardSkips(), writeRootlessPodmanEnv()

### Community 138 - "pskCfg"
Cohesion: 0.33
Nodes (6): Config, pskCfg(), TestNothingGeneratesThePSK(), TestPSKEnvSatisfiesTheContainerRequirement(), TestPSKIsMandatoryOnContainers(), TestPSKWhitespaceIsNotAKey()

### Community 139 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 140 - ".ServerCert"
Cohesion: 0.83
Nodes (3): serverCertFile(), serverCertScript(), TestServerCertScript()

### Community 141 - "CheckHostPath"
Cohesion: 0.33
Nodes (5): CanonicalDuration(), TestCanonicalDuration(), CheckHostPath(), TestCheckHostPathAccepts(), TestCheckHostPathRejects()

## Knowledge Gaps
- **237 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+232 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 357 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **17 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `captureStdout`, `ContainerSecrets`, `testing.T`, `Manager`, `CheckCommand`, `pskCfg`, `CheckHostPath`, `convert_test.go`, `manager_test.go`, `parse`, `Config`, `Role`, `commands.go`, `config_test.go`, `Load`, `emitYAML`, `Command`, `limits_test.go`, `render_test.go`, `step`, `.applyContainerDefaults`, `Config`, `Get`, `ctrCfg`, `Config`, `Convert`, `Ops`, `.validateContainerArtifactValues`, `.applyScalingTierDefaults`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `ops_k8s.go`, `Manager`, `K8sConfig`, `Set`, `renderDriver`, `ImportPlan`, `Config`, `newTestOps`, `bg`, `.releaseToBackup`, `config_test.go`, `ParseBlocks`, `.rpc`, `newTestMate`, `Ops`, `k8s/matechannel_test.go`, `New`, `.ConfigureReplication`, `step`, `ParseRole`, `.LeaderLocal`, `Config`, `scripts_test.go`, `HARoles`, `Ops`?**
  _High betweenness centrality (0.019) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `Quadlet`, `captureStdout`, `ContainerSecrets`, `testing.T`, `Manager`, `K8sConfig`, `cli_test.go`, `NewCluster`, `prep_test.go`, `convert_test.go`, `manager_test.go`, `AdminSecret`, `eqArgs`, `Platform`, `newTestOps`, `ResolveDomainCerts`, `Role`, `recRunner`, `ReplSite`, `localCfg`, `config_test.go`, `.configRows`, `strings.Builder`, `k8s/matechannel_test.go`, `operatorversion.go`, `render_test.go`, `step`, `.applyContainerDefaults`, `ctrCfg`, `desiredWatch`, `Compose`, `HARoles`, `EnvPairs`, `Cluster`, `Ops`, `eqArgs`, `Cluster`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 48 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 48 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _237 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ops_k8s.go` be split into smaller, more focused modules?**
  _Cohesion score 0.1356382978723404 - nodes in this community are weakly interconnected._
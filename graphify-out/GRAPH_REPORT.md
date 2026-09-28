# Graph Report - solace-cnt-scripts  (2026-09-28)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 3811 nodes · 16004 edges · 145 communities (128 shown, 17 thin omitted)
- Extraction: 86% EXTRACTED · 14% INFERRED · 0% AMBIGUOUS · INFERRED: 2205 edges (avg confidence: 0.85)
- Token cost: 95,240 input · 5,185 output

## Graph Freshness
- Built from commit: `8fb56002`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- CLI Kubernetes Ops Wiring
- Ops Call Recording Helpers
- Output Sink Formatting
- Rootless Podman Preflight
- CLI Test Helpers
- VPN Replication Parsing Tests
- Command Reference Docs
- Container Manager Checks
- Kubernetes Config Schema
- Quadlet and Health Checks
- Command Allowlist Guard
- Container CLI Tests
- Kubernetes Validate Report Tests
- Kubernetes Secret Generation
- Annotated Doc Builder
- Config Block Transforms
- Container State Report Tests
- Operator Command Tree
- Admin Secret Rendering
- Bash Dev Script
- Exit Code Tests
- PowerShell Dev Script
- Abbreviation Sets
- DR Switchover Plan
- Local Role Detection Tests
- Exec Runner Tests
- Import Driver Chunks
- Kubernetes Cluster Tests
- Import Plan and Apply
- Platform Resolution
- Solace Module Root
- Config Validation Rules
- Docs and Legacy Scripts
- Broker CLI Execution
- Domain Cert Directory Loading
- Broker Transport Fakes
- CLI Container Ops Wiring
- Kubernetes Admin Secret Preflight
- Capture Annotation Markers
- Shell Completion Tests
- Config Block Diff
- Platform Selection
- Service Shutdown Injection
- Redundancy Role Detection
- Command Tree Builders
- Echo Test Runner
- Kubernetes Broker Inspection
- Replication Site Config
- Backup Revert Activity Tests
- SEMP Reply Decoding
- Config Export Capture
- Command Alias Tests
- Export Config Test Doubles
- Container Security Guidelines
- Operations Guide
- Test Catalogue Docs
- Kubectl Cluster Client
- Switchover Preflight
- Broker Default Config Toggles
- Broker Admin SEMP Operations
- Confirmation Prompts
- Kubernetes Check Report Rows
- Rootless Podman Tests
- Abbreviation Docs Rendering
- Output Formatting Tests
- Configuration Guide
- Allow-Command Security Tests
- Mate Channel Tests
- Command Reference Doc
- Developer Guide
- Replication Mate Selection
- Namespace Occupancy Tests
- Kubernetes Status Reports
- K8s Package Tests
- Capture Block Markers
- CLI Exit Errors
- CLI Mate Replication
- Command Field Decoding
- Import Apply Tests
- Container File Transport
- Abbrev Package
- LoadBalancer Port Parsing
- Kubernetes Mate Channel
- Container Limits Tests
- Storage Mount Validation
- Troubleshooting Guide
- Kubernetes Inspect Tests
- Show Replication Parsing
- Operator Version Checks
- Artifact Render Tests
- Vulnerability Judge Tool
- Import Section Rules
- CLI App Setup
- Container Config Defaults
- Home Path Expansion
- Secret Reference Resolution
- Node Role Parsing
- Replication Config Validation
- Abbreviation Reference
- Container Secret Specs
- Container Resource Limits
- Container Engine Preflight
- Examples and Config Loading
- Container Secret Bundle Tests
- Broker Apply Confirmation
- CLI Package Tests
- Operator Watch Namespaces
- Compose File Rendering
- Import-Config Apply Rules
- Config Secret Name Resolution
- Broker Package Tests
- Operator Bundle Rendering
- Capturing Runner Double
- Config Package Tests
- Container Manager Tests
- Kubernetes Resource Naming
- Container Env and Cert Paths
- Container Package Tests
- Container State Inspection
- Replication VPN Lines
- SEMP Port Resolution
- Broker Ops CLI Helpers
- Broker Ops Test Harness
- Artifact Value Validation
- Node Identity Resolution
- Destructive Removal Confirmations
- Scaling Tier Config
- Operator Watch Reconciliation
- Import Ignore Rules
- Scaling YAML Decoding
- Kubernetes RBAC Preflight
- Replication Via Validation
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
- `wantEnable()` --calls--> `usagef()`  [INFERRED]
  internal/cli/flags.go → internal/cli/exit.go
- `opCtrConfigDefaultUsers()` --calls--> `wantEnable()`  [INFERRED]
  internal/cli/ops_container.go → internal/cli/flags.go
- `opCtrConfigDefaultVPN()` --calls--> `wantEnable()`  [INFERRED]
  internal/cli/ops_container.go → internal/cli/flags.go
- `wantRemove()` --calls--> `usagef()`  [INFERRED]
  internal/cli/flags.go → internal/cli/exit.go
- `opCtrConfigDomainCerts()` --calls--> `wantRemove()`  [INFERRED]
  internal/cli/ops_container.go → internal/cli/flags.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (145 total, 17 thin omitted)

### Community 0 - "CLI Kubernetes Ops Wiring"
Cohesion: 0.14
Nodes (55): wantEnable(), wantRemove(), bg(), domainCANames(), App, k8sAdminPassword(), k8sCluster(), k8sContext() (+47 more)

### Community 1 - "Ops Call Recording Helpers"
Cohesion: 0.10
Nodes (35): opCall, os.File, ReplicationConfigResult, capture(), captureStdout(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod() (+27 more)

### Community 2 - "Output Sink Formatting"
Cohesion: 0.20
Nodes (5): solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), columnWidths(), Sink, pad()

### Community 3 - "Rootless Podman Preflight"
Cohesion: 0.05
Nodes (17): Exec, context.Context, os/exec.Cmd, Manager, idMapCovers(), origin(), runUserIDs(), EnvRunner (+9 more)

### Community 4 - "CLI Test Helpers"
Cohesion: 0.09
Nodes (55): allowRuntime(), captureStderr(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone(), runStatusStderr(), TestAnnounceCommandsNamesResolvedBinaries() (+47 more)

### Community 5 - "VPN Replication Parsing Tests"
Cohesion: 0.02
Nodes (193): testing.T, flagByte(), ParseVPNReplication(), TestParseVPNReplication(), TestParseVPNReplicationIgnoresRepeatedHeaders(), TestParseVPNReplicationIgnoresTheEchoedPrompt(), TestParseVPNReplicationReadsColumnOrderFromTheHeader(), TestParseVPNReplicationRejectsUnreadable() (+185 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Container Manager Checks"
Cohesion: 0.10
Nodes (8): certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing()

### Community 8 - "Kubernetes Config Schema"
Cohesion: 0.15
Nodes (8): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, boolStr(), writeSecurity()

### Community 9 - "Quadlet and Health Checks"
Cohesion: 0.15
Nodes (18): ContainerNoFile(), HealthCheck, TestLimitsCheckAssertsWhatTheArtifactAsks(), escapePercent(), healthCmd(), Quadlet(), quadletEscape(), containerArtifacts() (+10 more)

### Community 10 - "Command Allowlist Guard"
Cohesion: 0.15
Nodes (22): commandRules, guardedCmd, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules() (+14 more)

### Community 11 - "Container CLI Tests"
Cohesion: 0.15
Nodes (26): testing.M, runCtr(), TestConfigStepsDoNotLeakSecrets(), TestCtrConfigDryRun(), TestCtrDiagnosticsDryRun(), TestCtrErrorPaths(), TestCtrExecCLIPathSeparator(), TestCtrLoginOutcomes() (+18 more)

### Community 12 - "Kubernetes Validate Report Tests"
Cohesion: 0.08
Nodes (61): TestCheckStopsProbingWhenUnreachable(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+53 more)

### Community 13 - "Kubernetes Secret Generation"
Cohesion: 0.07
Nodes (38): GenBroker(), GenSecrets(), joinManifests(), adminCfg(), secretNamesIn(), TestAdminSecretStatesEmitTheirArtifacts(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree() (+30 more)

### Community 14 - "Annotated Doc Builder"
Cohesion: 0.06
Nodes (64): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+56 more)

### Community 15 - "Config Block Transforms"
Cohesion: 0.09
Nodes (52): Block, TargetState, regexp.Regexp, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects() (+44 more)

### Community 16 - "Container State Report Tests"
Cohesion: 0.07
Nodes (97): inspectReply(), kvValue(), TestInspectStateSurfacesTheEngineError(), TestReportStateNeverPrintsTheEnvironment(), TestReportStateOnAHealthyRunningContainer(), TestReportStateOnAStoppedContainer(), TestRestartCountComesFromSystemdOnPodman(), TestRestartCountDoesNotAttributeOurUnitToAnotherContainer() (+89 more)

### Community 17 - "Operator Command Tree"
Cohesion: 0.12
Nodes (37): github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, addLogFlags(), newOperatorCmd(), newOperatorDeployCmd(), newOperatorGenerateCmd(), newOperatorLogsCmd(), newOperatorRemoveCmd() (+29 more)

### Community 18 - "Admin Secret Rendering"
Cohesion: 0.18
Nodes (13): AdditionalUsersSecret(), AdminSecret(), decodeDataValue(), TestAdditionalUsersSecretCarriesBothHalves(), TestAdditionalUsersSecretIsAbsentWithNoUsers(), TestAdditionalUsersSecretRefusesAnEmptyPassword(), TestAdditionalUsersStayOutOfTheCredentialsSecret(), TestAdminSecretCarriesThePSKOnlyWhenSet() (+5 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "Exit Code Tests"
Cohesion: 0.16
Nodes (13): runFailRunner, TestExecute(), ExitCode(), childStatusError(), k8sEnv(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitStatusIsScopedToInteractiveSessions(), TestExitCodeContract() (+5 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Abbreviation Sets"
Cohesion: 0.13
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 23 - "DR Switchover Plan"
Cohesion: 0.13
Nodes (28): PhaseKind, SiteState, SwitchAction, SwitchPhase, BuildSwitchPlan(), confirmDemotions(), ExecuteSwitchPlan(), SwitchPlan (+20 more)

### Community 24 - "Local Role Detection Tests"
Cohesion: 0.13
Nodes (46): cliScriptPath(), defaultLocalAddrs(), curlCalls(), Ops, isCurl(), newLocalOps(), rd(), seqTransport() (+38 more)

### Community 25 - "Exec Runner Tests"
Cohesion: 0.11
Nodes (28): TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), MaskEnv(), quoteTok(), captureStdout(), helperCommand() (+20 more)

### Community 26 - "Import Driver Chunks"
Cohesion: 0.16
Nodes (18): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), TestKeywordPatternEscapesEachPhrase() (+10 more)

### Community 27 - "Kubernetes Cluster Tests"
Cohesion: 0.11
Nodes (30): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), saCfg() (+22 more)

### Community 28 - "Import Plan and Apply"
Cohesion: 0.14
Nodes (20): ImportResult, PlannedSection, isPragma(), defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList() (+12 more)

### Community 29 - "Platform Resolution"
Cohesion: 0.12
Nodes (30): checkFlagPlatforms(), commandPlatforms(), declaredList(), flagOnlyOn(), App, parsePlatformList(), platformSuffix(), prepare() (+22 more)

### Community 31 - "Config Validation Rules"
Cohesion: 0.09
Nodes (20): keyValueEntries, TestValidateContainerArtifactValues(), TestValidateProbeCommandAccepts(), TestValidateProbeCommandRejects(), TestRoleNames(), RoleNames(), checkCredentialChars(), foldToEnvVar() (+12 more)

### Community 32 - "Docs and Legacy Scripts"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "Broker CLI Execution"
Cohesion: 0.19
Nodes (8): runCLISkeleton(), validName(), shQuote(), TestShQuoteHandlesASingleQuote(), rejectionIn(), shellScriptPath(), BaseName(), TestBaseNameSplitsOnBothSeparators()

### Community 34 - "Domain Cert Directory Loading"
Cohesion: 0.06
Nodes (44): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, os.DirEntry, os.FileInfo, os.FileMode (+36 more)

### Community 35 - "Broker Transport Fakes"
Cohesion: 0.10
Nodes (14): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+6 more)

### Community 36 - "CLI Container Ops Wiring"
Cohesion: 0.10
Nodes (44): TestCtrManagerConfirmWiring(), confirmAction(), childExit(), TestChildExitKeepsItsMessage(), TestExitCodeIsNeverNegative(), emitOrWrite(), containerRenderRole(), containerRole() (+36 more)

### Community 37 - "Kubernetes Admin Secret Preflight"
Cohesion: 0.07
Nodes (33): AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestReadAdminPasswordArgvAndDecoding(), TestReadAdminPasswordGuardsTheCommand(), TestReadAdminPasswordNamesTheNextStep() (+25 more)

### Community 38 - "Capture Annotation Markers"
Cohesion: 0.07
Nodes (46): Annotate(), applyMetaField(), Capture, markerSection(), markerVerb(), parseKV(), readMarker(), renderKV() (+38 more)

### Community 39 - "Shell Completion Tests"
Cohesion: 0.17
Nodes (21): driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion(), TestCompletionDescriptionsAreOptIn() (+13 more)

### Community 40 - "Config Block Diff"
Cohesion: 0.11
Nodes (36): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, ancestorReported(), blockLabel(), DiffBlocks() (+28 more)

### Community 41 - "Platform Selection"
Cohesion: 0.29
Nodes (12): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+4 more)

### Community 42 - "Service Shutdown Injection"
Cohesion: 0.16
Nodes (28): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "Redundancy Role Detection"
Cohesion: 0.16
Nodes (11): countContains(), field(), showRedundancyDetailScript(), showRedundancyLocalScript(), Ops, hostMatches(), shortHost(), activity() (+3 more)

### Community 44 - "Command Tree Builders"
Cohesion: 0.22
Nodes (46): opFunc, roleOpFunc, addCommands(), addPodFlag(), App, group(), newBrokerCLICmd(), newBrokerCmd() (+38 more)

### Community 45 - "Echo Test Runner"
Cohesion: 0.26
Nodes (4): interactiveFailRunner, Echo, Quote(), TestQuote()

### Community 46 - "Kubernetes Broker Inspection"
Cohesion: 0.20
Nodes (18): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), ownedPods(), brokerList, brokerStatus (+10 more)

### Community 47 - "Replication Site Config"
Cohesion: 0.19
Nodes (11): Replication, missingListedVPNs(), PlannedRoles(), RoleAtSite(), TestRoleAtSiteIsTheComplement(), curlConfigFlag(), Ops, replVPNNames() (+3 more)

### Community 48 - "Backup Revert Activity Tests"
Cohesion: 0.08
Nodes (35): TestLocalMateRunsThroughOps(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK(), TestBackupRevertActivitySuccess() (+27 more)

### Community 49 - "SEMP Reply Decoding"
Cohesion: 0.22
Nodes (43): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, go_pkg_bytes, go_pkg_context, go_pkg_crypto_sha256, go_pkg_encoding_base64 (+35 more)

### Community 50 - "Config Export Capture"
Cohesion: 0.13
Nodes (18): StripMarkers(), TestStripMarkersKeepsPragmaAndSectionComments(), TestStripMarkersNoMarkersPassthrough(), ParseBlocks(), splitLines(), TestParseBlocksCRLFMatchesLF(), TestParseBlocksIndentedLineWithNoOpenerRefused(), TestParseBlocksMissingEndTerminatorRefused() (+10 more)

### Community 51 - "Command Alias Tests"
Cohesion: 0.09
Nodes (28): TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing() (+20 more)

### Community 52 - "Export Config Test Doubles"
Cohesion: 0.14
Nodes (21): exportconfigReadCounter, opRunner, go_pkg_runtime, exportconfigDriverOutput(), exportconfigTransportOutput(), exportconfigVPNListOutput(), App, newExportconfigRunner() (+13 more)

### Community 53 - "Container Security Guidelines"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 54 - "Operations Guide"
Cohesion: 0.08
Nodes (26): Bringing up a fresh cluster, Broker scope is a fixed classification, Configuring a site, Data replication, Docker and Podman mechanics, Exit codes, Exporting and importing configuration, Extra CLI users differ by platform (+18 more)

### Community 55 - "Test Catalogue Docs"
Cohesion: 0.08
Nodes (24): abbrev_test.go, convert_test.go, Coverage, examples_test.go, Fixtures and doubles, Injectable seams, internal/abbrev, internal/convert (+16 more)

### Community 57 - "Switchover Preflight"
Cohesion: 0.17
Nodes (15): MateChannel, sortedKeys(), sortedVPNs(), requirePrimaryActive(), requireVirtualRouterNamesMatch(), sortedSiteNames(), SwitchPreflight(), replSites() (+7 more)

### Community 58 - "Broker Default Config Toggles"
Cohesion: 0.06
Nodes (50): colSpan, showCmd, validCLILine(), Ops, dashSpans(), gutterClear(), sliceSpan(), assertLeaderScript() (+42 more)

### Community 59 - "Broker Admin SEMP Operations"
Cohesion: 0.06
Nodes (28): AdminState, backupTarget, Credential, QueueState, ReplRole, scriptedMate, sempMate, VPNRepl (+20 more)

### Community 60 - "Confirmation Prompts"
Cohesion: 0.19
Nodes (24): layer, go_pkg_bufio, io.Reader, io.Writer, TestConfirmDeleteShortcut(), TestPromptYes(), TestPromptYesNo(), TestStdinCanAnswerClosedFile() (+16 more)

### Community 61 - "Kubernetes Check Report Rows"
Cohesion: 0.14
Nodes (22): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminSecretRow(), containsString(), failRow() (+14 more)

### Community 62 - "Rootless Podman Tests"
Cohesion: 0.17
Nodes (36): failOnCall(), fakeEnv(), Manager, healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+28 more)

### Community 63 - "Abbreviation Docs Rendering"
Cohesion: 0.09
Nodes (37): shorthand, WeightedNodeTerm, github.com/spf13/pflag.FlagSet, strings.Builder, mdRow(), renderAbbrevDocs(), treeShorthands(), writeAbbrevTable() (+29 more)

### Community 64 - "Output Formatting Tests"
Cohesion: 0.24
Nodes (14): sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLevelOrderKeepsSkipBelowFail(), TestLineAddsNoPrefix(), TestSectionPadsToWidth() (+6 more)

### Community 65 - "Configuration Guide"
Cohesion: 0.14
Nodes (14): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+6 more)

### Community 66 - "Allow-Command Security Tests"
Cohesion: 0.10
Nodes (38): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+30 more)

### Community 67 - "Mate Channel Tests"
Cohesion: 0.16
Nodes (17): CLIRunner, fakeRun, Ops, NewCLIMate(), mateReplies(), newTestMate(), TestMateChannelDescribeNamesTheTarget(), TestMateChannelPreflightLearnsTheBrokerType() (+9 more)

### Community 68 - "Command Reference Doc"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer Guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "Replication Mate Selection"
Cohesion: 0.32
Nodes (13): mateChannel(), mateSEMPPassword(), App, replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer(), TestMateChannelPicksTheMechanismTheSiteDeclares() (+5 more)

### Community 71 - "Namespace Occupancy Tests"
Cohesion: 0.18
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 72 - "Kubernetes Status Reports"
Cohesion: 0.24
Nodes (8): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownsPod(), TestOperatorRunningImageWithNoContainers(), TestOwnsPodFallsBackToTheNameInfix()

### Community 73 - "K8s Package Tests"
Cohesion: 0.10
Nodes (20): adminsecret_test.go, check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go (+12 more)

### Community 74 - "Capture Block Markers"
Cohesion: 0.11
Nodes (27): Omission, Region, token, isMarker(), markerRegion(), TestIsMarkerDiscriminatesNamespace(), TestMarkerRegionBeginAndEnd(), brokerTypeFromSchema() (+19 more)

### Community 75 - "CLI Exit Errors"
Cohesion: 0.11
Nodes (21): childExitError, usageError, github.com/spf13/cobra.PositionalArgs, logArgs(), noRolePositional(), App, newConvertCmd(), runConvert() (+13 more)

### Community 76 - "CLI Mate Replication"
Cohesion: 0.13
Nodes (14): BrokerType, cliMate, bannerType(), checkSameType(), Ops, isRunCLIRejection(), newlineIf(), replPhase1Rejected() (+6 more)

### Community 77 - "Command Field Decoding"
Cohesion: 0.20
Nodes (9): cmdField, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandUnmarshal(), Command, Config, guardCommandOf(), setGuardCommand() (+1 more)

### Community 78 - "Import Apply Tests"
Cohesion: 0.18
Nodes (14): planSections(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, newTeardownApplyFixture(), TestBuildChunksCreatesVPNsBeforeBrokerSections() (+6 more)

### Community 81 - "LoadBalancer Port Parsing"
Cohesion: 0.25
Nodes (8): LoadBalancer, cut(), parsePort(), splitUser(), TestParsePort(), writeKeyValueEntry(), writeLBAnnotations(), portSpec

### Community 82 - "Kubernetes Mate Channel"
Cohesion: 0.22
Nodes (25): CLIScriptPath(), execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf(), replCfg() (+17 more)

### Community 83 - "Container Limits Tests"
Cohesion: 0.14
Nodes (26): go_pkg_slices, Manager, limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager() (+18 more)

### Community 84 - "Storage Mount Validation"
Cohesion: 0.25
Nodes (8): Config, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestStorageClassAndCustomMountAreMutuallyExclusive()

### Community 85 - "Troubleshooting Guide"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 86 - "Kubernetes Inspect Tests"
Cohesion: 0.15
Nodes (19): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+11 more)

### Community 87 - "Show Replication Parsing"
Cohesion: 0.09
Nodes (42): MateConfig, TestMateChannelShowReplication(), cliTransport(), containsEndpoint(), labelValue(), normTransport(), parseHostPort(), ParseShowReplication() (+34 more)

### Community 88 - "Operator Version Checks"
Cohesion: 0.13
Nodes (19): compareVersions(), Cluster, imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), operatorDeployJSON(), TestCompareVersions() (+11 more)

### Community 89 - "Artifact Render Tests"
Cohesion: 0.19
Nodes (24): BrokerCR(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets(), TestBrokerCRQuotesTheImageReference(), TestComposeQuotesTheCpuset(), TestContainerOverridesReachArtifact(), TestContainerSecretNamesAreHostScoped() (+16 more)

### Community 90 - "Vulnerability Judge Tool"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "Import Section Rules"
Cohesion: 0.09
Nodes (34): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+26 more)

### Community 92 - "CLI App Setup"
Cohesion: 0.14
Nodes (20): mateChannelFunc, bufio.Reader, App, lineSink(), progress(), step(), confirmReplicationConfig(), confirmReplicationSwitch() (+12 more)

### Community 93 - "Container Config Defaults"
Cohesion: 0.13
Nodes (17): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults() (+9 more)

### Community 94 - "Home Path Expansion"
Cohesion: 0.16
Nodes (9): expandTilde(), expandTildeToken(), Config, IsAbsHostPath(), isPathSep(), TestContainerHostDirsExpandATilde(), TestExpandTilde(), TestIsAbsHostPath() (+1 more)

### Community 95 - "Secret Reference Resolution"
Cohesion: 0.50
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 96 - "Node Role Parsing"
Cohesion: 0.40
Nodes (6): argumentLine(), TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), ParseRole(), RoleAbbrev()

### Community 97 - "Replication Config Validation"
Cohesion: 0.31
Nodes (10): Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationEmptyViaAccepted(), TestValidateReplicationRejects() (+2 more)

### Community 98 - "Abbreviation Reference"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "Container Secret Specs"
Cohesion: 0.25
Nodes (8): ContainerSecrets(), containerSecretSpecs(), SecretPreflight(), TestSecretPreflight(), TestComposeLabelsTheCertificateDigest(), TestFileBackedSecretIsExemptFromSecretPreflight(), TestServerCertIsASecretOnBothEngines(), secretSpec

### Community 100 - "Container Resource Limits"
Cohesion: 0.23
Nodes (8): delegateSetting(), Manager, limitText(), missingControllers(), parseLimit(), parseUnitProps(), TestMissingControllers(), TestParseLimit()

### Community 102 - "Examples and Config Loading"
Cohesion: 0.07
Nodes (53): Example, go_pkg_embed, TestExamplesBareEmitsTheFullSchema(), TestExamplesEmitsToStdout(), minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError() (+45 more)

### Community 103 - "Container Secret Bundle Tests"
Cohesion: 0.17
Nodes (26): ResolveSecretValues(), bundleHash(), certCreate(), certFixture(), TestCertSecretLabelDrivesTheRestart(), TestCertSecretRemovalFailureIsFatal(), TestContainerBundleCarriesTheChain(), TestDeletePodmanToleratesAMissingBundle() (+18 more)

### Community 104 - "Broker Apply Confirmation"
Cohesion: 0.29
Nodes (4): time.Time, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

### Community 105 - "CLI Package Tests"
Cohesion: 0.15
Nodes (13): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, euid_test.go, examples_test.go (+5 more)

### Community 106 - "Operator Watch Namespaces"
Cohesion: 0.14
Nodes (14): desiredWatch(), Cluster, splitWatch(), subtractWatch(), TestDesiredWatchAppendsTheBrokerNamespace(), TestSubtractWatchEmptyRemainingMeansDelete(), TestSubtractWatchPreservesOrder(), TestUnionWatchKeepsAnotherEnvFilesNamespace() (+6 more)

### Community 107 - "Compose File Rendering"
Cohesion: 0.19
Nodes (10): Compose(), composeEscape(), ComposeProject(), ContainerSecret, secretFilePath(), TestComposeProjectFoldsToComposesGrammar(), TestComposeProjectIsDeclaredNotDerived(), TestMonitorIsSizedForQuorumNotForTheTier() (+2 more)

### Community 108 - "Import-Config Apply Rules"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "Config Secret Name Resolution"
Cohesion: 0.06
Nodes (23): AdditionalUser, Image, Node, Redundancy, SEMP, TLS, atoiPrefix(), Config (+15 more)

### Community 110 - "Broker Package Tests"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "Operator Bundle Rendering"
Cohesion: 0.14
Nodes (22): TestValidateSparseConfigExplainsItself(), GenOperator(), joinYAMLDocs(), nonEmpty(), operatorProbes(), RenderOperator(), renderOperatorNS(), splitAfterNamespace() (+14 more)

### Community 112 - "Capturing Runner Double"
Cohesion: 0.11
Nodes (23): capCall, capRunner, New(), callIndex(), TestManagerLogsCLIShell(), TestManagerPrepHostRootlessUsesUnshareChown(), TestPreflightRunsBeforeAnything(), TestCtrRuntimeDefaultArgvUnchanged() (+15 more)

### Community 113 - "Config Package Tests"
Cohesion: 0.12
Nodes (17): adminsecret_test.go, artifactvalues_test.go, command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go (+9 more)

### Community 114 - "Container Manager Tests"
Cohesion: 0.10
Nodes (20): bytes.Buffer, NewManager(), Manager, newEchoMgr(), TestCheckEnvReportsBaseDirOnlyWhenSet(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployPodmanDryRunHidesSecretBytes() (+12 more)

### Community 115 - "Kubernetes Resource Naming"
Cohesion: 0.13
Nodes (10): BrokerPodSuffixShape(), TestOperatorAdminSecretNameMatchesTheBrokerNames(), lbServiceName(), podName(), stsName(), TestPodNameSuffixMatchesTheConfigBound(), TestResourceNames(), Cluster (+2 more)

### Community 116 - "Container Env and Cert Paths"
Cohesion: 0.19
Nodes (12): EnvPairs(), groupKey(), itoa(), ServerCertBundlePath(), assertNoCheckoutPath(), envLines(), TestAdditionalUsersReachBothHalves(), TestGolden() (+4 more)

### Community 117 - "Container Package Tests"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.16
Nodes (14): containerState, healthState, decodeInspect(), Manager, orUnknown(), printable(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown() (+6 more)

### Community 119 - "Replication VPN Lines"
Cohesion: 0.33
Nodes (11): replicationVPNLines(), siteAEntry(), TestNothingEverShutsDownTheVPNItself(), TestPlannedRolesNamesEveryListedVPN(), TestReplicationVPNLinesEnablesListedAndDisablesTheRest(), TestReplicationVPNLinesIsDeterministic(), TestReplicationVPNLinesQuotesNamesWithSpaces(), TestReplicationVPNLinesReEnablesWhatPhase1StoppedOnly() (+3 more)

### Community 120 - "SEMP Port Resolution"
Cohesion: 0.29
Nodes (7): bridgeHostPort(), sempPort(), TestSempPortBridgeFallsBackToPlaintext(), TestSempPortBridgePrefersTLS(), TestSempPortBridgeRefusesNoMapping(), TestSempPortHostMode(), TestSempPortWithoutCertificateStaysPlaintext()

### Community 121 - "Broker Ops CLI Helpers"
Cohesion: 0.09
Nodes (22): time.Duration, Ops, TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), CLIArg(), cliArg(), confirmImport() (+14 more)

### Community 122 - "Broker Ops Test Harness"
Cohesion: 0.03
Nodes (143): cliRunNames(), Ops, hasCall(), matchCLI(), newTestOps(), outputForRole(), ranContains(), TestCountContains() (+135 more)

### Community 123 - "Artifact Value Validation"
Cohesion: 0.22
Nodes (6): Config, portOrRange(), TestValidContainerPort(), validContainerPort(), validHealthDuration(), validDNSLabel()

### Community 125 - "Destructive Removal Confirmations"
Cohesion: 0.40
Nodes (5): Every destructive command confirms, Removing a broker: what stays, what goes, Removing the operator does not always remove it, The layer flag raises the question, The namespace is only offered when it is empty

### Community 126 - "Scaling Tier Config"
Cohesion: 0.15
Nodes (11): scalingTier, containerMem(), cpuSetCount(), cpuSetRange(), Config, TestContainerMem(), TestCPUSetCount(), TestCPUSetRange() (+3 more)

### Community 127 - "Operator Watch Reconciliation"
Cohesion: 0.29
Nodes (7): deployJSON(), Cluster, TestOperatorReleaseNarrowsInsteadOfRemoving(), TestReconcileWatchDecidesWhatToApply(), TestReconcileWatchFlagsAWideningAsAQuestion(), TestSetWatchRefusesAnEmptyList(), watchCluster()

### Community 128 - "Import Ignore Rules"
Cohesion: 0.67
Nodes (3): importIgnore(), TestImportIgnoreKeepsUnclassifiedSectionsInTheDiff(), TestImportOpsImportIgnore()

### Community 129 - "Scaling YAML Decoding"
Cohesion: 0.31
Nodes (7): scalingKey, scalingSpelling, Scaling, Scaling, yaml.Node, scalingKeyIndex(), scalingKeyList()

### Community 131 - "Kubernetes RBAC Preflight"
Cohesion: 0.44
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "Replication Via Validation"
Cohesion: 0.22
Nodes (8): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, viaDescription(), sortedKeys(), validateReplVia(), validateReplViaSEMP()

### Community 139 - "Env Path Resolution"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 141 - "Duration and Host Path Checks"
Cohesion: 0.40
Nodes (5): CanonicalDuration(), TestCanonicalDuration(), CheckHostPath(), TestCheckHostPathAccepts(), TestCheckHostPathRejects()

## Knowledge Gaps
- **237 isolated node(s):** `abbrevdoc_test.go`, `aliases_test.go`, `allowcommand_test.go`, `cli_test.go`, `commanddoc_test.go` (+232 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 357 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **17 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Config` connect `Config Secret Name Resolution` to `Ops Call Recording Helpers`, `Rootless Podman Preflight`, `CLI Test Helpers`, `Container Manager Checks`, `Kubernetes Config Schema`, `Quadlet and Health Checks`, `Kubernetes Validate Report Tests`, `Kubernetes Secret Generation`, `Annotated Doc Builder`, `Container State Report Tests`, `Admin Secret Rendering`, `Kubernetes Cluster Tests`, `Platform Resolution`, `Domain Cert Directory Loading`, `Kubernetes Admin Secret Preflight`, `Replication Site Config`, `Backup Revert Activity Tests`, `SEMP Reply Decoding`, `Kubectl Cluster Client`, `Kubernetes Check Report Rows`, `Abbreviation Docs Rendering`, `Namespace Occupancy Tests`, `Container File Transport`, `Kubernetes Mate Channel`, `Artifact Render Tests`, `CLI App Setup`, `Container Config Defaults`, `Container Secret Specs`, `Container Secret Bundle Tests`, `Operator Watch Namespaces`, `Compose File Rendering`, `Operator Bundle Rendering`, `Capturing Runner Double`, `Container Manager Tests`, `Kubernetes Resource Naming`, `Container Env and Cert Paths`, `SEMP Port Resolution`, `Broker Ops CLI Helpers`, `Broker Ops Test Harness`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Why does `Platform` connect `Platform Resolution` to `Ops Call Recording Helpers`, `VPN Replication Parsing Tests`, `Container Manager Checks`, `Command Allowlist Guard`, `Annotated Doc Builder`, `Container State Report Tests`, `Config Validation Rules`, `Command Tree Builders`, `SEMP Reply Decoding`, `Command Field Decoding`, `Container File Transport`, `Container Limits Tests`, `Artifact Render Tests`, `CLI App Setup`, `Container Config Defaults`, `Home Path Expansion`, `Container Secret Specs`, `Examples and Config Loading`, `Container Secret Bundle Tests`, `Capturing Runner Double`, `Container Manager Tests`, `SEMP Port Resolution`, `Broker Ops CLI Helpers`, `Artifact Value Validation`, `Scaling Tier Config`?**
  _High betweenness centrality (0.022) - this node is a cross-community bridge._
- **Why does `Role` connect `Broker Transport Fakes` to `CLI Kubernetes Ops Wiring`, `Container Manager Checks`, `Kubernetes Config Schema`, `Abbreviation Sets`, `Import Driver Chunks`, `Import Plan and Apply`, `Config Validation Rules`, `Broker CLI Execution`, `CLI Container Ops Wiring`, `Kubernetes Admin Secret Preflight`, `Redundancy Role Detection`, `SEMP Reply Decoding`, `Config Export Capture`, `Broker Default Config Toggles`, `Broker Admin SEMP Operations`, `Mate Channel Tests`, `CLI Mate Replication`, `Container File Transport`, `Kubernetes Mate Channel`, `CLI App Setup`, `Node Role Parsing`, `Config Secret Name Resolution`, `Kubernetes Resource Naming`, `Broker Ops CLI Helpers`, `Broker Ops Test Harness`, `Node Identity Resolution`?**
  _High betweenness centrality (0.018) - this node is a cross-community bridge._
- **Are the 49 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 49 INFERRED edges - model-reasoned connections that need verification._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **What connects `abbrevdoc_test.go`, `aliases_test.go`, `allowcommand_test.go` to the rest of the system?**
  _237 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `CLI Kubernetes Ops Wiring` be split into smaller, more focused modules?**
  _Cohesion score 0.13506493506493505 - nodes in this community are weakly interconnected._
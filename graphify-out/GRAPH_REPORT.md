# Graph Report - solace-cnt-scripts  (2026-09-20)

## Corpus Check
- 191 files · ~571,645 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 3658 nodes · 12263 edges · 136 communities (130 shown, 6 thin omitted)
- Extraction: 83% EXTRACTED · 17% INFERRED · 0% AMBIGUOUS · INFERRED: 2091 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `233eb03c`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- Ops
- Sink
- context.Context
- completion_test.go
- config_test.go
- Commands
- Manager
- config.go
- GenOperator
- Command
- ops_container.go
- NewCluster
- go_pkg_testing
- convert_test.go
- Block
- ctrCfg
- blocks_test.go
- k8s/secrets_test.go
- dev.sh
- testing.T
- dev.ps1
- render.go
- switchplan_test.go
- verify_local_test.go
- runner_test.go
- driver.go
- .ConfigureReplication
- importops.go
- confirmAction
- solace
- validate.go
- CLAUDE.md
- importops_test.go
- domaincerts_resolve_test.go
- Role
- secretref.go
- recRunner
- runRootWith
- ContainerSecret
- Load
- cli/platform_test.go
- inject.go
- .releaseToBackup
- github.com/spf13/cobra.Command
- Echo
- age
- k8s/matechannel_test.go
- Quadlet
- go_pkg_strings
- sempmate.go
- newRootCmd
- opRunner
- exportconfig_test.go
- Operations
- Test catalogue
- Exporting and importing configuration
- diff.go
- switchplan.go
- semp.go
- cli_test.go
- checkreport.go
- rootless_test.go
- Cluster
- New
- Configuration
- scaling_test.go
- broker/replication.go
- Command reference
- Developer guide
- blocks.go
- semp_test.go
- Manager
- internal/k8s
- annotate_test.go
- usagef
- namespace_test.go
- Config
- scripts_test.go
- resolve_test.go
- examples/examples_test.go
- k8s/inspect_test.go
- cli/replication_test.go
- limits_test.go
- haCfg
- ParseBlocks
- k8s/inspect.go
- load.go
- config/replication_test.go
- render_test.go
- vulnjudge/main.go
- RuleFor
- cli/replication.go
- newEchoMgr
- redundancy_test.go
- scaling.go
- newTestOps
- hostpath.go
- Abbreviations
- ReplSite
- .checkUserManagerLimits
- scripts.go
- storage_test.go
- container/secrets_test.go
- watch_test.go
- internal/cli
- prep_test.go
- replicationops_test.go
- What `import-config` applies
- execguard_test.go
- internal/broker
- statusreport_test.go
- container/transport_test.go
- internal/config
- Container
- .Preflight
- containerTransport
- internal/container
- .stateRows
- hostpath_test.go
- .LeaderLocal
- Ops
- App
- container/inspect_test.go
- k8s/runtime_test.go
- Fixtures and doubles
- .Run
- imagepullsecret_test.go
- operatorversion.go
- ResolveEnvPath
- Cluster
- .preflightOne
- Troubleshooting
- Cluster
- Data replication
- tlssecret_test.go

## God Nodes (most connected - your core abstractions)
1. `newTestOps()` - 129 edges
2. `ctrCfg()` - 124 edges
3. `Role` - 116 edges
4. `newCapMgr()` - 107 edges
5. `Config` - 95 edges
6. `NewCluster()` - 92 edges
7. `loadK8s()` - 85 edges
8. `matchCLI()` - 81 edges
9. `Platform` - 76 edges
10. `bg()` - 68 edges

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

## Communities (136 total, 6 thin omitted)

### Community 0 - "bg"
Cohesion: 0.15
Nodes (50): bg(), domainCANames(), App, k8sCluster(), k8sContext(), k8sLogin(), k8sOps(), k8sWhat() (+42 more)

### Community 1 - "Ops"
Cohesion: 0.11
Nodes (11): runCLISkeleton(), validCLILine(), validName(), Ops, domainCertsScript(), enableDefaultVPNScript(), removeDomainCertsScript(), removeServerCertScript() (+3 more)

### Community 2 - "Sink"
Cohesion: 0.20
Nodes (5): columnWidths(), Level, Row, Sink, pad()

### Community 3 - "context.Context"
Cohesion: 0.07
Nodes (12): Exec, context.Context, os/exec.Cmd, BaseName(), TestBaseNameSplitsOnBothSeparators(), Cluster, Cluster, Cluster (+4 more)

### Community 4 - "completion_test.go"
Cohesion: 0.19
Nodes (19): driveBash(), runComplete(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion(), TestCompletionHelpStillWorks(), TestCompletionNeedsAShell() (+11 more)

### Community 5 - "config_test.go"
Cohesion: 0.06
Nodes (54): assertContainerBlockDefaults(), assertContainerScaling(), Config, haNodesConfig(), TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes(), TestAdditionalUserNamesCollideOnDocker(), TestAdditionalUserNamesDoNotCollideOnKubernetes() (+46 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (7): composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing()

### Community 8 - "config.go"
Cohesion: 0.08
Nodes (25): AdditionalUser, Broker, CertDir, ContainerSecurity, DomainCerts, Image, Node, Operator (+17 more)

### Community 9 - "GenOperator"
Cohesion: 0.14
Nodes (21): GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator(), renderOperatorNS(), splitAfterNamespace() (+13 more)

### Community 10 - "Command"
Cohesion: 0.12
Nodes (27): cmdField, commandRules, guardedCmd, go_pkg_unicode, go_pkg_unicode_utf8, Command, checkBinary(), CheckCommand() (+19 more)

### Community 11 - "ops_container.go"
Cohesion: 0.15
Nodes (40): TestCtrManagerConfirmWiring(), lineSink(), wantRemove(), containerRenderRole(), containerRole(), containerWhat(), ctrLogin(), ctrManager() (+32 more)

### Community 12 - "NewCluster"
Cohesion: 0.09
Nodes (59): TestCheckStopsProbingWhenUnreachable(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+51 more)

### Community 13 - "go_pkg_testing"
Cohesion: 0.15
Nodes (25): go_pkg_testing, TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster() (+17 more)

### Community 14 - "convert_test.go"
Cohesion: 0.05
Nodes (64): doc, Result, segment, vars, Config, boolOf(), commentSafe(), Convert() (+56 more)

### Community 15 - "Block"
Cohesion: 0.09
Nodes (53): Block, TargetState, regexp.Regexp, importIgnore(), injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs() (+45 more)

### Community 16 - "ctrCfg"
Cohesion: 0.08
Nodes (88): fileExists(), assertMode(), containsStr(), ctrCfg(), hasCall(), maskedKeys(), newCapMgr(), TestContainerRunningMatchesNameExactly() (+80 more)

### Community 17 - "blocks_test.go"
Cohesion: 0.23
Nodes (13): go_pkg_reflect, TestAnnotateIsDeterministic(), blockByIndex(), Capture, loadSample(), TestParseBlocksCRLFMatchesLF(), TestParseBlocksHeaderAndPreamble(), TestParseBlocksIndentedLineWithNoOpenerRefused() (+5 more)

### Community 18 - "k8s/secrets_test.go"
Cohesion: 0.10
Nodes (34): go_pkg_encoding_base64, GenBroker(), GenSecrets(), joinManifests(), namespaceManifest(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret() (+26 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "testing.T"
Cohesion: 0.06
Nodes (75): testing.T, matchCLI(), ranContains(), TestLastLines(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle() (+67 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "render.go"
Cohesion: 0.18
Nodes (23): strings.Builder, Placement, PodAffinityTerm, boolStr(), cut(), parsePort(), parseToleration(), sortedKeys() (+15 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.19
Nodes (27): SiteState, BuildSwitchPlan(), replPrecheckError(), actionStrings(), ordersToA(), replSites(), replVPNs(), st() (+19 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.12
Nodes (45): CLIArg(), cliArg(), CLIScriptPath(), cliScriptPath(), defaultLocalAddrs(), Ops, isCurl(), newLocalOps() (+37 more)

### Community 25 - "runner_test.go"
Cohesion: 0.15
Nodes (21): captureStdout(), helperCommand(), TestEchoDefaultWriter(), TestEchoOutput(), TestEchoOutputInput(), TestEchoRun(), TestEchoRunEnv(), TestEchoRunEnvNoEnv() (+13 more)

### Community 26 - "driver.go"
Cohesion: 0.18
Nodes (19): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), shQuote() (+11 more)

### Community 27 - ".ConfigureReplication"
Cohesion: 0.08
Nodes (19): AdminState, cliMate, ReplRole, scriptedMate, VPNRepl, rejectionIn(), Ops, ReplicationConfigResult (+11 more)

### Community 28 - "importops.go"
Cohesion: 0.23
Nodes (15): ImportResult, PlannedSection, defaultVPNFirst(), ImportPlan, orNone(), preambleForApply(), quoteList(), renderRegionScript() (+7 more)

### Community 29 - "confirmAction"
Cohesion: 0.20
Nodes (24): layer, go_pkg_bufio, io.Reader, io.Writer, TestConfirmDeleteShortcut(), TestPromptYes(), TestPromptYesNo(), TestStdinCanAnswerClosedFile() (+16 more)

### Community 31 - "validate.go"
Cohesion: 0.11
Nodes (20): keyValueEntries, BrokerPodSuffixShape(), checkCredentialChars(), foldToEnvVar(), Config, invertedCPUSetRange(), missingErr(), platformKey() (+12 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.16
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "importops_test.go"
Cohesion: 0.10
Nodes (32): planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, minimalCapture() (+24 more)

### Community 34 - "domaincerts_resolve_test.go"
Cohesion: 0.07
Nodes (53): DirReader, fakeDirEntry, go_pkg_path, os.DirEntry, os.FileInfo, os.FileMode, caNameSafe(), checkNoDotDot() (+45 more)

### Community 35 - "Role"
Cohesion: 0.06
Nodes (33): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+25 more)

### Community 36 - "secretref.go"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.15
Nodes (9): TestCanIAnswerReadsTheLastLine(), NewTransport(), isCanI(), TestTransportCopy(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest(), recRunner (+1 more)

### Community 38 - "runRootWith"
Cohesion: 0.10
Nodes (36): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+28 more)

### Community 40 - "Load"
Cohesion: 0.14
Nodes (26): minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors(), TestLoadSuccess() (+18 more)

### Community 41 - "cli/platform_test.go"
Cohesion: 0.29
Nodes (16): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+8 more)

### Community 42 - "inject.go"
Cohesion: 0.17
Nodes (27): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+19 more)

### Community 43 - ".releaseToBackup"
Cohesion: 0.22
Nodes (9): countContains(), field(), showRedundancyLocalScript(), TestBackupActivityStateReadsTheMateColumn(), activity(), backupActivityState(), Ops, primaryRedundancyUp() (+1 more)

### Community 44 - "github.com/spf13/cobra.Command"
Cohesion: 0.05
Nodes (143): opFunc, roleOpFunc, shorthand, go_pkg_github_com_spf13_cobra, go_pkg_github_com_spf13_pflag, go_pkg_solace_internal_abbrev, go_pkg_solace_internal_container, go_pkg_solace_internal_convert (+135 more)

### Community 45 - "Echo"
Cohesion: 0.29
Nodes (3): interactiveFailRunner, Echo, Quote()

### Community 46 - "age"
Cohesion: 0.27
Nodes (7): orNone(), orValue(), age(), roleRank(), Cluster, operatorRunningImage(), TestOperatorRunningImageWithNoContainers()

### Community 47 - "k8s/matechannel_test.go"
Cohesion: 0.26
Nodes (21): Runner, execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf(), replCfg() (+13 more)

### Community 48 - "Quadlet"
Cohesion: 0.13
Nodes (17): ContainerNoFile(), Config, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), EnvPairs(), escapePercent(), groupKey(), itoa() (+9 more)

### Community 49 - "go_pkg_strings"
Cohesion: 0.10
Nodes (33): go_pkg_bytes, go_pkg_context, go_pkg_encoding_json, go_pkg_errors, go_pkg_fmt, go_pkg_io, go_pkg_math, go_pkg_net (+25 more)

### Community 50 - "sempmate.go"
Cohesion: 0.36
Nodes (7): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, go_pkg_encoding_xml, encoding/xml.Name, xmlEscape()

### Community 51 - "newRootCmd"
Cohesion: 0.10
Nodes (25): go_pkg_solace_internal_cli, TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias(), TestAllowCommandIsRegisteredWhereItExecutes() (+17 more)

### Community 52 - "opRunner"
Cohesion: 0.14
Nodes (18): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch(), opCanI() (+10 more)

### Community 53 - "exportconfig_test.go"
Cohesion: 0.12
Nodes (27): exportconfigReadCounter, runFailRunner, go_pkg_runtime, ExitCode(), childStatusError(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitStatusIsScopedToInteractiveSessions(), exportconfigDriverOutput() (+19 more)

### Community 54 - "Operations"
Cohesion: 0.11
Nodes (18): Bringing up a fresh cluster, Docker and Podman mechanics, Every destructive command confirms, Exit codes, Extra CLI users differ by platform, Operations, Post-deployment configuration order, Re-deploying is safe and explicit (+10 more)

### Community 55 - "Test catalogue"
Cohesion: 0.10
Nodes (20): abbrev_test.go, convert_test.go, Coverage, examples_test.go, internal/abbrev, internal/convert, internal/engine, internal/examples (+12 more)

### Community 56 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 57 - "diff.go"
Cohesion: 0.12
Nodes (35): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, ancestorReported(), blockLabel(), DiffBlocks() (+27 more)

### Community 58 - "switchplan.go"
Cohesion: 0.30
Nodes (13): PhaseKind, SwitchAction, SwitchPhase, MateChannel, confirmDemotions(), ExecuteSwitchPlan(), SwitchPlan, promoteConfirmed() (+5 more)

### Community 59 - "semp.go"
Cohesion: 0.13
Nodes (16): backupTarget, Credential, sempMate, TestHTTPStatusHelpers(), anyHTTP2xx(), bridgeHostPort(), curlConfigFlag(), curlConfigLine() (+8 more)

### Community 60 - "cli_test.go"
Cohesion: 0.08
Nodes (61): os.File, allowRuntime(), capture(), captureStderr(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone() (+53 more)

### Community 61 - "checkreport.go"
Cohesion: 0.16
Nodes (19): setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), containsString(), failRow(), Cluster, info() (+11 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.15
Nodes (39): go_pkg_slices, failOnCall(), fakeEnv(), Manager, healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir() (+31 more)

### Community 63 - "Cluster"
Cohesion: 0.15
Nodes (5): time.Time, Cluster, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

### Community 64 - "New"
Cohesion: 0.11
Nodes (26): TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), progress(), solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), TestCheckReportSkipsEmptySections(), New() (+18 more)

### Community 65 - "Configuration"
Cohesion: 0.15
Nodes (13): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+5 more)

### Community 66 - "scaling_test.go"
Cohesion: 0.08
Nodes (32): scalingTier, containerMem(), cpuSetCount(), cpuSetRange(), Config, clearTierDerived(), decodeScaling(), Config (+24 more)

### Community 67 - "broker/replication.go"
Cohesion: 0.06
Nodes (76): CLIRunner, colSpan, fakeRun, MateConfig, QueueState, Ops, NewCLIMate(), mateReplies() (+68 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "blocks.go"
Cohesion: 0.11
Nodes (28): Omission, Region, token, isMarker(), markerRegion(), TestIsMarkerDiscriminatesNamespace(), brokerTypeFromSchema(), firstQuoted() (+20 more)

### Community 71 - "semp_test.go"
Cohesion: 0.08
Nodes (48): TestLocalMateRunsThroughOps(), sempPort(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK() (+40 more)

### Community 72 - "Manager"
Cohesion: 0.21
Nodes (4): Manager, idMapCovers(), origin(), runUserIDs()

### Community 73 - "internal/k8s"
Cohesion: 0.11
Nodes (19): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go, names_test.go (+11 more)

### Community 74 - "annotate_test.go"
Cohesion: 0.09
Nodes (38): Annotate(), applyMetaField(), Capture, markerSection(), markerVerb(), parseKV(), readMarker(), renderKV() (+30 more)

### Community 75 - "usagef"
Cohesion: 0.11
Nodes (20): childExitError, usageError, github.com/spf13/cobra.PositionalArgs, logArgs(), asUsage(), childExit(), childStatus(), isUsage() (+12 more)

### Community 76 - "namespace_test.go"
Cohesion: 0.30
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 77 - "Config"
Cohesion: 0.15
Nodes (11): TLS, Config, Scaling, ContainerSecrets(), containerSecretSpecs(), SecretPreflight(), TestFileBackedSecretIsExemptFromSecretPreflight(), TestPodmanNeverGetsAFileBackedSecret() (+3 more)

### Community 78 - "scripts_test.go"
Cohesion: 0.19
Nodes (14): parseVPNNames(), productKeyScript(), productKeysScript(), removeProductKeysScript(), TestDisableDefaultUsersScriptQuoting(), TestParseVPNNames(), TestParseVPNNamesKeepsAMultiWordName(), TestParseVPNNamesNoSeparator() (+6 more)

### Community 79 - "resolve_test.go"
Cohesion: 0.18
Nodes (13): TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveMissingBinaryIsActionable(), TestResolveRefusesCurrentDirectory(), verboseExec(), MaskEnv() (+5 more)

### Community 80 - "examples/examples_test.go"
Cohesion: 0.07
Nodes (42): Entry, Example, go_pkg_embed, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies() (+34 more)

### Community 81 - "k8s/inspect_test.go"
Cohesion: 0.18
Nodes (17): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+9 more)

### Community 82 - "cli/replication_test.go"
Cohesion: 0.26
Nodes (17): mateChannel(), mateSEMPPassword(), reportReplicationConfig(), App, replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer() (+9 more)

### Community 83 - "limits_test.go"
Cohesion: 0.15
Nodes (25): Manager, limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager(), TestCheckLimitsRootlessRefusesUndelegatedControllers() (+17 more)

### Community 84 - "haCfg"
Cohesion: 0.14
Nodes (22): TestResolveStorageClass(), TestStorageRows(), TestDeleteBrokerProbesClaimsOnlyWhenPurging(), lbServiceName(), haCfg(), saCfg(), TestHARoles(), TestPodNameSuffixMatchesTheConfigBound() (+14 more)

### Community 85 - "ParseBlocks"
Cohesion: 0.13
Nodes (18): BrokerType, StripMarkers(), TestStripMarkersKeepsPragmaAndSectionComments(), TestStripMarkersNoMarkersPassthrough(), ParseBlocks(), splitLines(), bannerType(), checkProvenance() (+10 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.20
Nodes (18): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), ownedPods(), brokerList, brokerStatus (+10 more)

### Community 87 - "load.go"
Cohesion: 0.22
Nodes (13): TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config, parseError() (+5 more)

### Community 88 - "config/replication_test.go"
Cohesion: 0.36
Nodes (11): Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationAbsent(), TestValidateReplicationEmptyViaAccepted() (+3 more)

### Community 89 - "render_test.go"
Cohesion: 0.14
Nodes (33): BrokerCR(), Compose(), composeEscape(), ComposeProject(), assertNoCheckoutPath(), healthCheckFixture(), load(), TestArtifactsCarryNoSecrets() (+25 more)

### Community 90 - "vulnjudge/main.go"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.09
Nodes (39): Disposition, SectionRule, go_pkg_flag, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware() (+31 more)

### Community 92 - "cli/replication.go"
Cohesion: 0.31
Nodes (16): mateChannelFunc, go_pkg_solace_internal_k8s, step(), confirmReplicationConfig(), confirmReplicationSwitch(), App, opConfigureReplication(), opCtrConfigureReplication() (+8 more)

### Community 93 - "newEchoMgr"
Cohesion: 0.15
Nodes (13): bytes.Buffer, NewManager(), Manager, newEchoMgr(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployPodmanDryRunHidesSecretBytes(), TestManagerDockerCheckProbesCompose() (+5 more)

### Community 94 - "redundancy_test.go"
Cohesion: 0.21
Nodes (12): Config, pskCfg(), TestFillStandaloneNodeNameIgnoresAnEmptyHostname(), TestFillStandaloneNodeNameKeepsAConfiguredName(), TestFillStandaloneNodeNameNeverFillsHA(), TestFillStandaloneNodeNameUsesTheHost(), TestNothingGeneratesThePSK(), TestPSKEnvSatisfiesTheContainerRequirement() (+4 more)

### Community 95 - "scaling.go"
Cohesion: 0.29
Nodes (8): scalingKey, scalingSpelling, go_pkg_gopkg_in_yaml_v3, Scaling, Scaling, yaml.Node, scalingKeyIndex(), scalingKeyList()

### Community 96 - "newTestOps"
Cohesion: 0.07
Nodes (58): cliRunNames(), Ops, newTestOps(), outputForRole(), TestCountContains(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs() (+50 more)

### Community 97 - "hostpath.go"
Cohesion: 0.25
Nodes (9): expandTilde(), expandTildeToken(), HasPathSeparator(), IsAbsHostPath(), isPathSep(), TestExpandTilde(), TestHasPathSeparator(), TestIsAbsHostPath() (+1 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "ReplSite"
Cohesion: 0.29
Nodes (7): Replication, PlannedRoles(), RoleAtSite(), Ops, ReplSite, ReplVPN, secretCommand()

### Community 100 - ".checkUserManagerLimits"
Cohesion: 0.23
Nodes (8): delegateSetting(), Manager, limitText(), missingControllers(), parseLimit(), parseUnitProps(), TestMissingControllers(), TestParseLimit()

### Community 101 - "scripts.go"
Cohesion: 0.16
Nodes (21): showCmd, defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), enableDefaultUsersScript(), gatherConfigsScript(), noReleaseActivityScript(), releaseActivityScript() (+13 more)

### Community 102 - "storage_test.go"
Cohesion: 0.33
Nodes (9): Config, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 103 - "container/secrets_test.go"
Cohesion: 0.23
Nodes (18): ResolveSecretValues(), certFixture(), TestDeletePodmanRemovesTheBundleAfterTheUnit(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeWritingTheComposeFile(), TestDeployDockerPassesTheBundleAsEnvNotArgv(), TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(), TestDeployPodmanFailsWhenTheBundleCannotBeWritten() (+10 more)

### Community 104 - "watch_test.go"
Cohesion: 0.11
Nodes (25): imageFromDeployment(), desiredWatch(), Cluster, splitWatch(), subtractWatch(), deployJSON(), Cluster, operatorItem() (+17 more)

### Community 105 - "internal/cli"
Cohesion: 0.15
Nodes (13): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, euid_test.go, examples_test.go (+5 more)

### Community 106 - "prep_test.go"
Cohesion: 0.26
Nodes (13): adminCfg(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree(), TestCreateSecretsPreflight(), TestCreateSecretsStopsOnPreflightFailure(), TestDeleteSecrets(), TestDeleteSecretsStopsOnError(), TestDeleteSecretsStopsOnPreflightFailure() (+5 more)

### Community 107 - "replicationops_test.go"
Cohesion: 0.17
Nodes (28): hasCall(), replicationVPNLines(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), siteAEntry(), TestConfigureReplicationClosingShowFailureStillReportsTheApply() (+20 more)

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "execguard_test.go"
Cohesion: 0.24
Nodes (10): Config, guardCommandOf(), setGuardCommand(), TestAllowCommandsAccepts(), TestAllowCommandsRejects(), TestAllowedBinaryIsNotGloballyAllowed(), TestComposeCommandDerivation(), TestComposeDerivationInheritsRejection() (+2 more)

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "statusreport_test.go"
Cohesion: 0.17
Nodes (19): ownsPod(), Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList() (+11 more)

### Community 112 - "container/transport_test.go"
Cohesion: 0.12
Nodes (27): capCall, capRunner, New(), Transport, callIndex(), TestManagerLogsCLIShell(), Manager, mgrOver() (+19 more)

### Community 113 - "internal/config"
Cohesion: 0.13
Nodes (15): command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go, hostpath_test.go, imagepullsecret_test.go (+7 more)

### Community 114 - "Container"
Cohesion: 0.22
Nodes (8): Container, DockerConfig, Network, PodmanConfig, Ulimits, HealthCheck, healthCmd(), TestHealthCmdDefaultsToReadiness()

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - ".stateRows"
Cohesion: 0.29
Nodes (6): containerState, healthState, Manager, printable(), TestPrintableStripsControlCharacters(), KV

### Community 119 - "hostpath_test.go"
Cohesion: 0.11
Nodes (23): CanonicalDuration(), TestCanonicalDuration(), guardConfig(), CheckHostPath(), Config, setContainerName(), TestCheckHostPathAccepts(), TestCheckHostPathRejects() (+15 more)

### Community 120 - ".LeaderLocal"
Cohesion: 0.33
Nodes (4): assertLeaderScript(), showRedundancyDetailScript(), TestAssertLeaderScript(), Ops

### Community 121 - "Ops"
Cohesion: 0.14
Nodes (10): time.Duration, Ops, confirmImport(), App, pluralVPN(), runExport(), runImport(), TestPluralVPNWording() (+2 more)

### Community 122 - "App"
Cohesion: 0.33
Nodes (3): bufio.Reader, App, EnvRunner

### Community 123 - "container/inspect_test.go"
Cohesion: 0.15
Nodes (19): decodeInspect(), orUnknown(), inspectReply(), kvValue(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings(), TestInspectNameLosesDockersLeadingSlash() (+11 more)

### Community 124 - "k8s/runtime_test.go"
Cohesion: 0.43
Nodes (7): TestClusterHonoursRuntime(), TestExecutorRefusesUnapprovedRuntime(), TestRuntimeDefaultArgvUnchanged(), TestTransportHonoursRuntime(), unapprovedCfg(), withLeading(), wrappedCfg()

### Community 125 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 126 - ".Run"
Cohesion: 0.16
Nodes (24): testing.M, runCtr(), TestConfigStepsDoNotLeakSecrets(), TestCtrConfigDryRun(), TestCtrDiagnosticsDryRun(), TestCtrErrorPaths(), TestCtrExecCLIPathSeparator(), TestCtrLoginOutcomes() (+16 more)

### Community 127 - "imagepullsecret_test.go"
Cohesion: 0.39
Nodes (8): Config, pullSecretCfg(), TestImagePullSecretDerivedSuffixFitsWhereAdditionalUsersAlreadyFits(), TestImagePullSecretNameBringYourOwn(), TestImagePullSecretNameConfiguredNameWins(), TestImagePullSecretNameDerivedWhenCredentialsButNoName(), TestImagePullSecretNameEmptyWithNeither(), TestManagesImagePullSecretNeedsBothCredentials()

### Community 128 - "operatorversion.go"
Cohesion: 0.36
Nodes (7): compareVersions(), imageTag(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestImageTag(), TestOperatorVersionWarningFiresOnlyOnADowngrade()

### Community 129 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 131 - ".preflightOne"
Cohesion: 0.50
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 134 - "Data replication"
Cohesion: 0.40
Nodes (5): Configuring a site, Data replication, Reaching the mate, Switching roles, When a run fails partway

### Community 135 - "tlssecret_test.go"
Cohesion: 0.43
Nodes (7): Config, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsMakeTheSecretOurs(), TestSuppliedCertsNeedASecretName(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), tlsCfg()

## Knowledge Gaps
- **223 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+218 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 330 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **6 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `github.com/spf13/cobra.Command` to `config_test.go`, `Manager`, `config.go`, `Command`, `convert_test.go`, `ctrCfg`, `validate.go`, `Load`, `opRunner`, `scaling_test.go`, `semp_test.go`, `Config`, `examples/examples_test.go`, `limits_test.go`, `load.go`, `render_test.go`, `newEchoMgr`, `redundancy_test.go`, `hostpath.go`, `container/secrets_test.go`, `execguard_test.go`, `container/transport_test.go`, `Container`, `containerTransport`, `hostpath_test.go`, `Ops`, `App`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `bg`, `Manager`, `config.go`, `GenOperator`, `NewCluster`, `convert_test.go`, `ctrCfg`, `k8s/secrets_test.go`, `render.go`, `Role`, `recRunner`, `github.com/spf13/cobra.Command`, `k8s/matechannel_test.go`, `Quadlet`, `opRunner`, `cli_test.go`, `checkreport.go`, `Cluster`, `semp_test.go`, `namespace_test.go`, `haCfg`, `render_test.go`, `newEchoMgr`, `newTestOps`, `ReplSite`, `container/secrets_test.go`, `watch_test.go`, `prep_test.go`, `container/transport_test.go`, `Container`, `containerTransport`, `Ops`, `App`, `k8s/runtime_test.go`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `Ops`, `context.Context`, `Manager`, `config.go`, `ops_container.go`, `driver.go`, `.ConfigureReplication`, `importops.go`, `validate.go`, `.releaseToBackup`, `Quadlet`, `semp.go`, `broker/replication.go`, `ParseBlocks`, `cli/replication.go`, `newTestOps`, `containerTransport`, `.LeaderLocal`, `Ops`?**
  _High betweenness centrality (0.014) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 42 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 42 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _223 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `bg` be split into smaller, more focused modules?**
  _Cohesion score 0.14901960784313725 - nodes in this community are weakly interconnected._
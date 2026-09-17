# Graph Report - solace-cnt-scripts  (2026-09-17)

## Corpus Check
- 189 files · ~569,218 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 3649 nodes · 12227 edges · 131 communities (124 shown, 7 thin omitted)
- Extraction: 83% EXTRACTED · 17% INFERRED · 0% AMBIGUOUS · INFERRED: 2082 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `fb77ca0c`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- Set
- Sink
- context.Context
- completion_test.go
- testing.T
- Commands
- Manager
- config.go
- operator.go
- Command
- ops_container.go
- NewCluster
- eqArgs
- convert_test.go
- Block
- ctrCfg
- ParseBlocks
- k8s/secrets_test.go
- dev.sh
- matchCLI
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
- replicationops.go
- recRunner
- runRootWith
- ContainerSecret
- broker/replication.go
- cli/platform_test.go
- inject.go
- .releaseToBackup
- github.com/spf13/cobra.Command
- Echo
- age
- ReplSite
- TestCtrTransportHonoursRuntime
- go_pkg_strings
- go_pkg_strconv
- newRootCmd
- opRunner
- exportconfig_test.go
- Operations
- Test catalogue
- Exporting and importing configuration
- diff.go
- platformOps
- semp.go
- cli_test.go
- checkreport.go
- rootless_test.go
- Cluster
- New
- Configuration
- cpuSetRange
- broker/replication_test.go
- Command reference
- Developer guide
- blocks.go
- semp_test.go
- Platform
- internal/k8s
- annotate_test.go
- exit_test.go
- namespace_test.go
- Config
- VPNRepl
- usagef
- examples/examples_test.go
- k8s/inspect_test.go
- cli/replication_test.go
- limits_test.go
- haCfg
- BrokerType
- k8s/inspect.go
- load.go
- config/replication_test.go
- render_test.go
- vulnjudge/main.go
- RuleFor
- cli/replication.go
- allowcommand_test.go
- runExport
- scaling.go
- newTestOps
- Config
- Abbreviations
- commands.go
- limits.go
- scripts.go
- command_test.go
- container/secrets_test.go
- watch_test.go
- internal/cli
- prep_test.go
- replicationops_test.go
- What `import-config` applies
- execguard_test.go
- internal/broker
- statusreport_test.go
- capRunner
- internal/config
- container/runtime_test.go
- .Preflight
- broker/transport.go
- internal/container
- .stateRows
- hostpath_test.go
- logArgs
- Ops
- .rpc
- container/inspect_test.go
- TestClusterHonoursRuntime
- Fixtures and doubles
- TestMain
- .preflightOne
- Troubleshooting
- Cluster
- tlssecret_test.go

## God Nodes (most connected - your core abstractions)
1. `newTestOps()` - 129 edges
2. `ctrCfg()` - 124 edges
3. `Role` - 116 edges
4. `newCapMgr()` - 108 edges
5. `Config` - 94 edges
6. `NewCluster()` - 92 edges
7. `loadK8s()` - 85 edges
8. `matchCLI()` - 81 edges
9. `Platform` - 75 edges
10. `bg()` - 68 edges

## Surprising Connections (you probably didn't know these)
- `parseBody()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `parseBody()` --calls--> `readMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `applyMetaField()` --calls--> `BrokerType`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go
- `applyMetaField()` --calls--> `brokerTypeFromSchema()`  [INFERRED]
  internal/broker/annotate.go → internal/broker/blocks.go
- `parseBody()` --calls--> `markerRegion()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (131 total, 7 thin omitted)

### Community 0 - "bg"
Cohesion: 0.16
Nodes (48): bg(), domainCANames(), App, k8sCluster(), k8sContext(), k8sLogin(), k8sOps(), k8sWhat() (+40 more)

### Community 1 - "Set"
Cohesion: 0.14
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 2 - "Sink"
Cohesion: 0.17
Nodes (8): solaceRows(), columnWidths(), KV, Level, Row, Sink, pad(), TestLevelTagsMapToTheHouseTags()

### Community 3 - "context.Context"
Cohesion: 0.05
Nodes (17): scriptedMate, Exec, context.Context, os/exec.Cmd, Manager, idMapCovers(), origin(), runUserIDs() (+9 more)

### Community 4 - "completion_test.go"
Cohesion: 0.19
Nodes (19): driveBash(), runComplete(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion(), TestCompletionHelpStillWorks(), TestCompletionNeedsAShell() (+11 more)

### Community 5 - "testing.T"
Cohesion: 0.04
Nodes (128): testing.T, assertContainerBlockDefaults(), assertContainerScaling(), envTree(), Config, haNodesConfig(), minimalK8s(), TestAdditionalUserNameFollowsTheBrokerRule() (+120 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (7): composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing()

### Community 8 - "config.go"
Cohesion: 0.07
Nodes (33): AdditionalUser, Broker, CertDir, Container, ContainerSecurity, DockerConfig, DomainCerts, Image (+25 more)

### Community 9 - "operator.go"
Cohesion: 0.15
Nodes (23): go_pkg_text_template, GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator(), renderOperatorNS() (+15 more)

### Community 10 - "Command"
Cohesion: 0.11
Nodes (29): cmdField, commandRules, guardedCmd, go_pkg_unicode, go_pkg_unicode_utf8, decodeCommand(), TestCommandUnmarshal(), Command (+21 more)

### Community 11 - "ops_container.go"
Cohesion: 0.15
Nodes (40): go_pkg_solace_internal_container, TestCtrManagerConfirmWiring(), wantRemove(), containerRenderRole(), containerRole(), containerWhat(), ctrLogin(), ctrManager() (+32 more)

### Community 12 - "NewCluster"
Cohesion: 0.09
Nodes (59): TestCheckStopsProbingWhenUnreachable(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+51 more)

### Community 13 - "eqArgs"
Cohesion: 0.17
Nodes (23): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), TestCLIAndShellAreInteractive() (+15 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (63): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+55 more)

### Community 15 - "Block"
Cohesion: 0.09
Nodes (53): Block, TargetState, regexp.Regexp, importIgnore(), injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs() (+45 more)

### Community 16 - "ctrCfg"
Cohesion: 0.07
Nodes (105): fileExists(), NewManager(), assertMode(), callIndex(), containsStr(), ctrCfg(), Manager, hasCall() (+97 more)

### Community 17 - "ParseBlocks"
Cohesion: 0.19
Nodes (19): go_pkg_reflect, Annotate(), TestAnnotateIsDeterministic(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), ParseBlocks(), blockByIndex(), Capture (+11 more)

### Community 18 - "k8s/secrets_test.go"
Cohesion: 0.10
Nodes (35): go_pkg_encoding_base64, go_pkg_encoding_json, GenBroker(), GenSecrets(), joinManifests(), namespaceManifest(), AdditionalUsersSecret(), AdminSecret() (+27 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "matchCLI"
Cohesion: 0.06
Nodes (51): matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultVPNDisableError() (+43 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "render.go"
Cohesion: 0.13
Nodes (29): strings.Builder, mdRow(), writeAbbrevTable(), PodAffinityTerm, boolStr(), containerSecretSpecs(), cut(), escapePercent() (+21 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.07
Nodes (66): CLIRunner, fakeRun, PhaseKind, ReplRole, SiteState, SwitchAction, SwitchPhase, MateChannel (+58 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.13
Nodes (43): cliScriptPath(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn(), TestDetectRoleAddrsError() (+35 more)

### Community 25 - "runner_test.go"
Cohesion: 0.12
Nodes (26): bytes.Buffer, TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), NewExec(), captureStdout(), helperCommand() (+18 more)

### Community 26 - "driver.go"
Cohesion: 0.18
Nodes (19): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), shQuote() (+11 more)

### Community 27 - ".ConfigureReplication"
Cohesion: 0.26
Nodes (7): ValidVPNName(), validVPNName(), Ops, isRunCLIRejection(), newlineIf(), replTransportFailure(), showVPNReplicationScript()

### Community 28 - "importops.go"
Cohesion: 0.13
Nodes (23): ImportResult, PlannedSection, checkProvenance(), checkSameType(), defaultVPNFirst(), describeScope(), ImportPlan, Ops (+15 more)

### Community 29 - "confirmAction"
Cohesion: 0.17
Nodes (25): exportconfigReadCounter, layer, go_pkg_bufio, io.Reader, io.Writer, TestConfirmDeleteShortcut(), TestPromptYes(), TestPromptYesNo() (+17 more)

### Community 31 - "validate.go"
Cohesion: 0.11
Nodes (20): keyValueEntries, BrokerPodSuffixShape(), checkCredentialChars(), foldToEnvVar(), Config, invertedCPUSetRange(), missingErr(), platformKey() (+12 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "importops_test.go"
Cohesion: 0.10
Nodes (33): planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, minimalCapture() (+25 more)

### Community 34 - "domaincerts_resolve_test.go"
Cohesion: 0.07
Nodes (53): DirReader, fakeDirEntry, go_pkg_path, os.DirEntry, os.FileInfo, os.FileMode, caNameSafe(), checkNoDotDot() (+45 more)

### Community 35 - "Role"
Cohesion: 0.07
Nodes (20): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+12 more)

### Community 36 - "replicationops.go"
Cohesion: 0.25
Nodes (7): secretRef, go_pkg_regexp, replPhase1Rejected(), replPhase2Rejected(), replVPNList(), Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.12
Nodes (12): TestCanIAnswerReadsTheLastLine(), TestExecutorRefusesUnapprovedRuntime(), unapprovedCfg(), NewTransport(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs() (+4 more)

### Community 38 - "runRootWith"
Cohesion: 0.11
Nodes (45): echoRunner(), App, runCtr(), runRootWith(), TestCLICommand(), TestConfigStepsDoNotLeakSecrets(), TestConfiguredRouternameSurvivesTheFallback(), TestCtrConfigDryRun() (+37 more)

### Community 40 - "broker/replication.go"
Cohesion: 0.19
Nodes (15): colSpan, QueueState, containsEndpoint(), dashSpans(), flagByte(), gutterClear(), labelValue(), normTransport() (+7 more)

### Community 41 - "cli/platform_test.go"
Cohesion: 0.17
Nodes (23): TestAllowCommandIsRegisteredWhereItExecutes(), findCmd(), TestEveryRunnableCommandIsWired(), TestFlagsRegistered(), TestFlagCompletionsRegistered(), runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer() (+15 more)

### Community 42 - "inject.go"
Cohesion: 0.17
Nodes (27): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+19 more)

### Community 43 - ".releaseToBackup"
Cohesion: 0.11
Nodes (19): countContains(), field(), TestCountContains(), TestField(), TestLastLines(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), TestLastLinesEqualCount() (+11 more)

### Community 44 - "github.com/spf13/cobra.Command"
Cohesion: 0.11
Nodes (32): shorthand, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, github.com/spf13/pflag.FlagSet, renderAbbrevDocs(), TestAbbreviationDocs(), treeShorthands(), writeResolutionNotes() (+24 more)

### Community 45 - "Echo"
Cohesion: 0.18
Nodes (9): interactiveFailRunner, TestChildEnvNamesAreNotSystemVariables(), Echo, MaskEnv(), Quote(), quoteTok(), TestMaskEnv(), TestQuote() (+1 more)

### Community 46 - "age"
Cohesion: 0.23
Nodes (8): orNone(), orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), age(), roleRank(), Cluster

### Community 47 - "ReplSite"
Cohesion: 0.20
Nodes (24): Replication, ReplSite, Runner, execArgs(), NewMateChannel(), ReadSecretKey(), secretCommand(), hasPrefixArgv() (+16 more)

### Community 48 - "TestCtrTransportHonoursRuntime"
Cohesion: 0.60
Nodes (5): TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime(), TestManagerReachableProbesRuntimeThenCompose(), withWrapper(), wrappedCtrCfg()

### Community 49 - "go_pkg_strings"
Cohesion: 0.11
Nodes (24): go_pkg_bytes, go_pkg_context, go_pkg_errors, go_pkg_fmt, go_pkg_github_com_spf13_cobra, go_pkg_io, go_pkg_os, go_pkg_path_filepath (+16 more)

### Community 50 - "go_pkg_strconv"
Cohesion: 0.13
Nodes (18): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, go_pkg_encoding_xml, go_pkg_net, go_pkg_strconv, encoding/xml.Name (+10 more)

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (29): go_pkg_solace_internal_examples, github.com/spf13/cobra.PositionalArgs, TestFlagShorthandsAreConsistent(), applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb() (+21 more)

### Community 52 - "opRunner"
Cohesion: 0.12
Nodes (20): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch(), opCanI() (+12 more)

### Community 53 - "exportconfig_test.go"
Cohesion: 0.18
Nodes (20): exportconfigDriverOutput(), exportconfigTransportOutput(), exportconfigVPNListOutput(), App, newExportconfigRunner(), newExportconfigRunnerRejectingChunk(), runImportConfirmation(), TestExportConfigLeafRunsOnAllPlatforms() (+12 more)

### Community 54 - "Operations"
Cohesion: 0.09
Nodes (23): Bringing up a fresh cluster, Configuring a site, Data replication, Docker and Podman mechanics, Every destructive command confirms, Exit codes, Extra CLI users differ by platform, Operations (+15 more)

### Community 55 - "Test catalogue"
Cohesion: 0.10
Nodes (20): abbrev_test.go, convert_test.go, Coverage, examples_test.go, internal/abbrev, internal/convert, internal/engine, internal/examples (+12 more)

### Community 56 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 57 - "diff.go"
Cohesion: 0.12
Nodes (35): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, ancestorReported(), blockLabel(), DiffBlocks() (+27 more)

### Community 58 - "platformOps"
Cohesion: 0.21
Nodes (28): opFunc, roleOpFunc, addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerCopyCmd(), newBrokerDeployCmd() (+20 more)

### Community 59 - "semp.go"
Cohesion: 0.14
Nodes (18): backupTarget, Credential, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigFlag(), curlConfigLine(), Ops, httpBody() (+10 more)

### Community 60 - "cli_test.go"
Cohesion: 0.08
Nodes (60): os.File, allowRuntime(), capture(), captureStderr(), collectPaths(), fakeBinaryOnPath(), firstLine(), runRoot() (+52 more)

### Community 61 - "checkreport.go"
Cohesion: 0.18
Nodes (17): additionalUsersRow(), allCustomMounted(), containsString(), failRow(), Cluster, info(), okRow(), portRows() (+9 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.17
Nodes (36): failOnCall(), fakeEnv(), Manager, healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+28 more)

### Community 64 - "New"
Cohesion: 0.12
Nodes (23): TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), progress(), TestCheckReportSkipsEmptySections(), New(), NewFunc(), sink() (+15 more)

### Community 65 - "Configuration"
Cohesion: 0.15
Nodes (13): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+5 more)

### Community 66 - "cpuSetRange"
Cohesion: 0.50
Nodes (4): cpuSetCount(), cpuSetRange(), TestCPUSetCount(), TestCPUSetRange()

### Community 67 - "broker/replication_test.go"
Cohesion: 0.15
Nodes (32): ParseShowReplication(), ParseVPNReplication(), RenderMate(), SameEndpoints(), loadShowReplication(), mateAppliance(), mateSoftware(), TestMateRoundTripsThroughItsOwnGrammar() (+24 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "blocks.go"
Cohesion: 0.13
Nodes (23): Omission, Region, token, brokerTypeFromSchema(), firstQuoted(), Capture, isIndented(), isPragma() (+15 more)

### Community 71 - "semp_test.go"
Cohesion: 0.10
Nodes (43): bridgeHostPort(), sempPort(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK() (+35 more)

### Community 72 - "Platform"
Cohesion: 0.22
Nodes (22): go_pkg_github_com_spf13_pflag, checkFlagPlatforms(), commandPlatforms(), declaredList(), flagOnlyOn(), App, parsePlatformList(), platformSuffix() (+14 more)

### Community 73 - "internal/k8s"
Cohesion: 0.11
Nodes (19): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go, names_test.go (+11 more)

### Community 74 - "annotate_test.go"
Cohesion: 0.08
Nodes (44): applyMetaField(), Capture, isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV(), readMarker() (+36 more)

### Community 75 - "exit_test.go"
Cohesion: 0.12
Nodes (16): childExitError, runFailRunner, usageError, go_pkg_os_exec, go_pkg_solace_internal_cli, childExit(), childStatus(), ExitCode() (+8 more)

### Community 76 - "namespace_test.go"
Cohesion: 0.30
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 77 - "Config"
Cohesion: 0.14
Nodes (12): TLS, Config, Scaling, OperatorNamespace(), TestPodNameForIsTheOneSpelling(), HARoles(), podName(), podNameFor() (+4 more)

### Community 78 - "VPNRepl"
Cohesion: 0.15
Nodes (19): VPNRepl, sortedKeys(), sortedSet(), mateConvergenceShutdowns(), missingListedVPNs(), reenableLine(), replicationVPNLines(), sortedVPNs() (+11 more)

### Community 79 - "usagef"
Cohesion: 0.18
Nodes (18): go_pkg_solace_internal_convert, newPerformExportConfigCmd(), noRolePositional(), registerFlagCompletion(), App, newConvertCmd(), runConvert(), usagef() (+10 more)

### Community 80 - "examples/examples_test.go"
Cohesion: 0.14
Nodes (27): Example, go_pkg_embed, go_pkg_flag, TestExamplesEmitsToStdout(), parseError(), DetectPlatforms(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint() (+19 more)

### Community 81 - "k8s/inspect_test.go"
Cohesion: 0.18
Nodes (17): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+9 more)

### Community 82 - "cli/replication_test.go"
Cohesion: 0.22
Nodes (19): ReplicationConfigResult, lineSink(), mateChannel(), mateSEMPPassword(), reportReplicationConfig(), App, replApp(), siteNamed() (+11 more)

### Community 83 - "limits_test.go"
Cohesion: 0.13
Nodes (27): go_pkg_slices, Manager, limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager() (+19 more)

### Community 84 - "haCfg"
Cohesion: 0.12
Nodes (24): TestResolveStorageClass(), TestStorageRows(), TestDeleteBrokerProbesClaimsOnlyWhenPurging(), lbServiceName(), haCfg(), saCfg(), TestHARoles(), TestPodNameSuffixMatchesTheConfigBound() (+16 more)

### Community 85 - "BrokerType"
Cohesion: 0.23
Nodes (10): BrokerType, cliMate, MateConfig, bannerType(), cliTransport(), renderMateAppliance(), RenderMateRemovals(), renderMateSoftware() (+2 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.13
Nodes (24): time.Time, anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), Cluster, normalizeToList(), TestConditionLevelDegradesSafely() (+16 more)

### Community 87 - "load.go"
Cohesion: 0.22
Nodes (12): TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config, podmanRunUser() (+4 more)

### Community 88 - "config/replication_test.go"
Cohesion: 0.36
Nodes (11): Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationAbsent(), TestValidateReplicationEmptyViaAccepted() (+3 more)

### Community 89 - "render_test.go"
Cohesion: 0.09
Nodes (54): ContainerNoFile(), NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), BrokerCR(), Compose(), composeEscape(), ComposeProject(), ContainerSecrets() (+46 more)

### Community 90 - "vulnjudge/main.go"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.10
Nodes (38): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+30 more)

### Community 92 - "cli/replication.go"
Cohesion: 0.19
Nodes (19): mateChannelFunc, go_pkg_solace_internal_k8s, bufio.Reader, App, step(), confirmReplicationConfig(), confirmReplicationSwitch(), App (+11 more)

### Community 93 - "allowcommand_test.go"
Cohesion: 0.29
Nodes (10): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+2 more)

### Community 94 - "runExport"
Cohesion: 0.22
Nodes (10): confirmImport(), App, pluralVPN(), runExport(), runImport(), TestPluralVPNWording(), emitOrWrite(), nowStamp() (+2 more)

### Community 95 - "scaling.go"
Cohesion: 0.14
Nodes (15): scalingKey, scalingSpelling, scalingTier, go_pkg_gopkg_in_yaml_v3, containerMem(), Config, Scaling, Scaling (+7 more)

### Community 96 - "newTestOps"
Cohesion: 0.08
Nodes (52): Ops, newTestOps(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultUsersRefusesAQuotedVPNName(), TestDisableDefaultVPN(), TestDomainCerts() (+44 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "commands.go"
Cohesion: 0.27
Nodes (31): addCommands(), addLogFlags(), App, newBrokerConfigureCmd(), newBrokerRemoveCmd(), newBrokerStartCmd(), newBrokerStopCmd(), newBrokerValidateCmd() (+23 more)

### Community 100 - "limits.go"
Cohesion: 0.22
Nodes (9): go_pkg_math, delegateSetting(), Manager, limitText(), missingControllers(), parseLimit(), parseUnitProps(), TestMissingControllers() (+1 more)

### Community 101 - "scripts.go"
Cohesion: 0.07
Nodes (44): showCmd, validCLILine(), Ops, defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript(), enableDefaultUsersScript() (+36 more)

### Community 102 - "command_test.go"
Cohesion: 0.08
Nodes (29): TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRenamedRuntimeKeysFailLoud(), TestRuntimeDefaults(), TestRuntimeExplicitValueSurvivesDefaults() (+21 more)

### Community 103 - "container/secrets_test.go"
Cohesion: 0.21
Nodes (19): ResolveSecretValues(), certFixture(), TestDeletePodmanRemovesTheBundleAfterTheUnit(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeWritingTheComposeFile(), TestDeployDockerPassesTheBundleAsEnvNotArgv(), TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(), TestDeployPodmanFailsWhenTheBundleCannotBeWritten() (+11 more)

### Community 104 - "watch_test.go"
Cohesion: 0.13
Nodes (23): imageFromDeployment(), desiredWatch(), splitWatch(), subtractWatch(), deployJSON(), Cluster, operatorItem(), TestAmbiguousOperatorIsNotFoldedIntoNotInstalled() (+15 more)

### Community 105 - "internal/cli"
Cohesion: 0.17
Nodes (12): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, examples_test.go, exit_test.go (+4 more)

### Community 106 - "prep_test.go"
Cohesion: 0.25
Nodes (13): adminCfg(), TestCreateNamespaceApplyFails(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree(), TestCreateSecretsPreflight(), TestCreateSecretsStopsOnPreflightFailure(), TestDeleteNamespaceStopsOnPreflightFailure(), TestDeleteSecrets() (+5 more)

### Community 107 - "replicationops_test.go"
Cohesion: 0.27
Nodes (17): hasCall(), TestDisableDefaultUsersShowVPNError(), TestReleaseToBackupReleasedTimeout(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), TestConfigureReplicationClosingShowFailureStillReportsTheApply() (+9 more)

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "execguard_test.go"
Cohesion: 0.21
Nodes (11): Config, guardCommandOf(), setGuardCommand(), TestAllowCommandsAccepts(), TestAllowCommandsRejects(), TestAllowedBinaryIsNotGloballyAllowed(), TestComposeCommandDerivation(), TestComposeDerivationInheritsRejection() (+3 more)

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "statusreport_test.go"
Cohesion: 0.17
Nodes (19): ownsPod(), Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList() (+11 more)

### Community 112 - "capRunner"
Cohesion: 0.16
Nodes (14): capCall, capRunner, New(), TestNewDefaults(), Transport, NewTransport(), dockerCfg(), podmanCfg() (+6 more)

### Community 113 - "internal/config"
Cohesion: 0.13
Nodes (15): command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go, hostpath_test.go, imagepullsecret_test.go (+7 more)

### Community 114 - "container/runtime_test.go"
Cohesion: 0.60
Nodes (4): Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), unapprovedCtrCfg()

### Community 116 - "broker/transport.go"
Cohesion: 0.67
Nodes (3): CLIArg(), cliArg(), CLIScriptPath()

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - ".stateRows"
Cohesion: 0.38
Nodes (4): containerState, healthState, Manager, printable()

### Community 119 - "hostpath_test.go"
Cohesion: 0.08
Nodes (34): CanonicalDuration(), TestCanonicalDuration(), guardConfig(), TestGuardConfigIsValid(), CheckHostPath(), expandTilde(), expandTildeToken(), HasPathSeparator() (+26 more)

### Community 120 - "logArgs"
Cohesion: 0.67
Nodes (4): logArgs(), App, logArgs2(), TestLogArgsBuildsOneSetForBothPlatforms()

### Community 121 - "Ops"
Cohesion: 0.11
Nodes (11): time.Duration, cliRunNames(), Ops, runCLISkeleton(), TestValidName(), validName(), rejectionIn(), setReplicationRoleScript() (+3 more)

### Community 122 - ".rpc"
Cohesion: 0.21
Nodes (6): AdminState, sempMate, schemaTransport(), TestTransportVocabularyIsThreeWords(), Ops, sempAdminState()

### Community 123 - "container/inspect_test.go"
Cohesion: 0.16
Nodes (18): decodeInspect(), inspectReply(), kvValue(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings(), TestInspectNameLosesDockersLeadingSlash(), TestInspectRejectsAnEmptyReply() (+10 more)

### Community 124 - "TestClusterHonoursRuntime"
Cohesion: 0.67
Nodes (4): TestClusterHonoursRuntime(), TestTransportHonoursRuntime(), withLeading(), wrappedCfg()

### Community 125 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 131 - ".preflightOne"
Cohesion: 0.50
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 135 - "tlssecret_test.go"
Cohesion: 0.43
Nodes (7): Config, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsMakeTheSecretOurs(), TestSuppliedCertsNeedASecretName(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), tlsCfg()

## Knowledge Gaps
- **222 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+217 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 329 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **7 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `testing.T`, `Manager`, `config.go`, `Command`, `convert_test.go`, `ctrCfg`, `validate.go`, `Role`, `opRunner`, `platformOps`, `semp_test.go`, `examples/examples_test.go`, `limits_test.go`, `load.go`, `render_test.go`, `cli/replication.go`, `scaling.go`, `commands.go`, `container/secrets_test.go`, `execguard_test.go`, `capRunner`, `container/runtime_test.go`, `hostpath_test.go`, `Ops`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `Set`, `testing.T`, `Manager`, `config.go`, `ops_container.go`, `switchplan_test.go`, `driver.go`, `.ConfigureReplication`, `importops.go`, `validate.go`, `.releaseToBackup`, `go_pkg_strings`, `semp.go`, `Config`, `BrokerType`, `cli/replication.go`, `runExport`, `scripts.go`, `Ops`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `Manager`, `config.go`, `operator.go`, `NewCluster`, `convert_test.go`, `ctrCfg`, `k8s/secrets_test.go`, `render.go`, `Role`, `recRunner`, `ReplSite`, `TestCtrTransportHonoursRuntime`, `opRunner`, `cli_test.go`, `checkreport.go`, `Cluster`, `semp_test.go`, `namespace_test.go`, `haCfg`, `render_test.go`, `cli/replication.go`, `newTestOps`, `container/secrets_test.go`, `watch_test.go`, `prep_test.go`, `capRunner`, `container/runtime_test.go`, `Ops`, `TestClusterHonoursRuntime`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 42 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 42 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _222 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Set` be split into smaller, more focused modules?**
  _Cohesion score 0.13538461538461538 - nodes in this community are weakly interconnected._
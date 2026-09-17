# Graph Report - solace-cnt-scripts  (2026-09-17)

## Corpus Check
- 189 files · ~567,731 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 22 file(s) not represented in the graph (top: .golden 16, (none) 2, .cli 2)

## Summary
- 3644 nodes · 12210 edges · 126 communities (121 shown, 5 thin omitted)
- Extraction: 83% EXTRACTED · 17% INFERRED · 0% AMBIGUOUS · INFERRED: 2078 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `2b2a9db7`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- Set
- Sink
- context.Context
- completion_test.go
- config_test.go
- Commands
- Manager
- Config
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
- runRoot
- dev.ps1
- render.go
- switchplan_test.go
- verify_local_test.go
- runner_test.go
- driver.go
- .ConfigureReplication
- importops.go
- confirm.go
- solace
- validate.go
- CLAUDE.md
- importops_test.go
- domaincerts_resolve_test.go
- Role
- .resolveSecretRefs
- recRunner
- Load
- Quadlet
- broker/replication.go
- cli/platform_test.go
- inject.go
- Manager
- commanddoc_test.go
- Echo
- age
- ReplSite
- eqArgs
- go_pkg_strings
- sempmate.go
- newRootCmd
- opRunner
- runRootWith
- Operations
- Test catalogue
- Exporting and importing configuration
- diff.go
- Data replication
- semp.go
- cli_test.go
- .configRows
- rootless_test.go
- Cluster
- output_test.go
- Configuration
- cpuSetRange
- broker/replication_test.go
- Command reference
- Developer guide
- blocks.go
- semp_test.go
- childExit
- internal/k8s
- annotate_test.go
- go_pkg_github_com_spf13_cobra
- namespace_test.go
- Ops
- newEchoMgr
- healthCmd
- Platforms
- k8s/inspect_test.go
- limits_test.go
- haCfg
- compareVersions
- k8s/inspect.go
- .applyContainerDefaults
- render_test.go
- vulnjudge/main.go
- RuleFor
- cli/replication.go
- scaling.go
- testing.T
- hostpath.go
- Abbreviations
- github.com/spf13/cobra.Command
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
- container/transport_test.go
- internal/config
- go_pkg_testing
- .Preflight
- internal/container
- .stateRows
- hostpath_test.go
- Ops
- ReplRole
- container/inspect_test.go
- containerTransport
- Fixtures and doubles
- storage_test.go
- completion.go
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
9. `Platform` - 74 edges
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

## Communities (126 total, 5 thin omitted)

### Community 0 - "bg"
Cohesion: 0.16
Nodes (47): bg(), domainCANames(), App, k8sCluster(), k8sContext(), k8sLogin(), k8sOps(), k8sWhat() (+39 more)

### Community 1 - "Set"
Cohesion: 0.14
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 2 - "Sink"
Cohesion: 0.13
Nodes (12): progress(), solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), TestCheckReportSkipsEmptySections(), columnWidths(), Row, Sink, New() (+4 more)

### Community 3 - "context.Context"
Cohesion: 0.07
Nodes (10): scriptedMate, Exec, context.Context, os/exec.Cmd, Cluster, Cluster, Cluster, Cluster (+2 more)

### Community 4 - "completion_test.go"
Cohesion: 0.19
Nodes (19): driveBash(), runComplete(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion(), TestCompletionHelpStillWorks(), TestCompletionNeedsAShell() (+11 more)

### Community 5 - "config_test.go"
Cohesion: 0.04
Nodes (74): assertContainerBlockDefaults(), assertContainerScaling(), envTree(), Config, haNodesConfig(), TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes(), TestAdditionalUserNamesCollideOnDocker() (+66 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (7): composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing()

### Community 8 - "Config"
Cohesion: 0.07
Nodes (32): AdditionalUser, Broker, CertDir, Container, ContainerSecurity, DockerConfig, DomainCerts, Image (+24 more)

### Community 9 - "operator.go"
Cohesion: 0.14
Nodes (24): go_pkg_text_template, GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator(), renderOperatorNS() (+16 more)

### Community 10 - "Command"
Cohesion: 0.12
Nodes (25): cmdField, commandRules, guardedCmd, go_pkg_unicode_utf8, Command, checkBinary(), CheckCommand(), checkFlagShape() (+17 more)

### Community 11 - "ops_container.go"
Cohesion: 0.14
Nodes (42): go_pkg_solace_internal_container, TestCtrManagerConfirmWiring(), confirmAction(), lineSink(), wantRemove(), containerRenderRole(), containerRole(), containerWhat() (+34 more)

### Community 12 - "NewCluster"
Cohesion: 0.09
Nodes (59): TestCheckStopsProbingWhenUnreachable(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+51 more)

### Community 13 - "eqArgs"
Cohesion: 0.17
Nodes (24): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), saCfg() (+16 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (66): doc, Result, segment, vars, DecodeStrict(), Config, boolOf(), commentSafe() (+58 more)

### Community 15 - "Block"
Cohesion: 0.09
Nodes (53): Block, TargetState, regexp.Regexp, importIgnore(), injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs() (+45 more)

### Community 16 - "ctrCfg"
Cohesion: 0.08
Nodes (86): fileExists(), assertMode(), containsStr(), ctrCfg(), hasCall(), maskedKeys(), newCapMgr(), TestContainerRunningMatchesNameExactly() (+78 more)

### Community 17 - "ParseBlocks"
Cohesion: 0.19
Nodes (19): Annotate(), TestAnnotateIsDeterministic(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), ParseBlocks(), blockByIndex(), Capture, loadSample() (+11 more)

### Community 18 - "k8s/secrets_test.go"
Cohesion: 0.09
Nodes (35): go_pkg_encoding_base64, go_pkg_encoding_json, go_pkg_flag, GenBroker(), GenSecrets(), joinManifests(), namespaceManifest(), AdditionalUsersSecret() (+27 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "runRoot"
Cohesion: 0.12
Nodes (23): firstLine(), runRoot(), TestBashEnvGivenToEnvFlag(), TestConfigureServerCertsRefusesASecretItDoesNotOwn(), TestConvertParseError(), TestConvertRoundTrip(), TestConvertToFile(), TestConvertToStdout() (+15 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "render.go"
Cohesion: 0.14
Nodes (28): WeightedNodeTerm, strings.Builder, NodeAffinity, NodeMatchExpr, Placement, PodAffinityTerm, boolStr(), containerSecretSpecs() (+20 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.08
Nodes (60): CLIRunner, fakeRun, PhaseKind, SiteState, SwitchAction, SwitchPhase, MateChannel, Ops (+52 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.13
Nodes (45): cliScriptPath(), defaultLocalAddrs(), curlCalls(), Ops, isCurl(), newLocalOps(), rd(), seqTransport() (+37 more)

### Community 25 - "runner_test.go"
Cohesion: 0.11
Nodes (26): TestChildEnvNamesAreNotSystemVariables(), MaskEnv(), quoteTok(), captureStdout(), helperCommand(), TestEchoDefaultWriter(), TestEchoOutput(), TestEchoOutputInput() (+18 more)

### Community 26 - "driver.go"
Cohesion: 0.17
Nodes (20): chunk, chunkResult, runCLISkeleton(), chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment() (+12 more)

### Community 27 - ".ConfigureReplication"
Cohesion: 0.11
Nodes (14): cliMate, QueueState, VPNRepl, ValidVPNName(), validVPNName(), Ops, isRunCLIRejection(), missingListedVPNs() (+6 more)

### Community 28 - "importops.go"
Cohesion: 0.12
Nodes (26): BrokerType, ImportResult, Omission, PlannedSection, bannerType(), checkProvenance(), checkSameType(), defaultVPNFirst() (+18 more)

### Community 29 - "confirm.go"
Cohesion: 0.17
Nodes (25): exportconfigReadCounter, layer, go_pkg_bufio, io.Reader, io.Writer, TestConfirmDeleteShortcut(), TestPromptYes(), TestPromptYesNo() (+17 more)

### Community 31 - "validate.go"
Cohesion: 0.11
Nodes (21): keyValueEntries, go_pkg_unicode, checkCredentialChars(), foldToEnvVar(), Config, invertedCPUSetRange(), missingErr(), platformKey() (+13 more)

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
Cohesion: 0.05
Nodes (31): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+23 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.18
Nodes (5): TestCanIAnswerReadsTheLastLine(), isCanI(), TestTransportExecArgs(), recRunner, rrCall

### Community 38 - "Load"
Cohesion: 0.07
Nodes (51): minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadReadError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors() (+43 more)

### Community 39 - "Quadlet"
Cohesion: 0.15
Nodes (14): ContainerNoFile(), NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), EnvPairs(), escapePercent(), ContainerSecret, groupKey(), itoa() (+6 more)

### Community 40 - "broker/replication.go"
Cohesion: 0.19
Nodes (24): MateConfig, cliTransport(), containsEndpoint(), labelValue(), normTransport(), parseHostPort(), ParseShowReplication(), parseShowReplicationAppliance() (+16 more)

### Community 41 - "cli/platform_test.go"
Cohesion: 0.29
Nodes (16): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+8 more)

### Community 42 - "inject.go"
Cohesion: 0.16
Nodes (28): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "Manager"
Cohesion: 0.21
Nodes (4): Manager, idMapCovers(), origin(), runUserIDs()

### Community 44 - "commanddoc_test.go"
Cohesion: 0.16
Nodes (24): shorthand, go_pkg_github_com_spf13_pflag, github.com/spf13/pflag.FlagSet, mdRow(), renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands() (+16 more)

### Community 45 - "Echo"
Cohesion: 0.24
Nodes (4): interactiveFailRunner, runFailRunner, Echo, Quote()

### Community 46 - "age"
Cohesion: 0.30
Nodes (6): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), TestOperatorRunningImageWithNoContainers()

### Community 47 - "ReplSite"
Cohesion: 0.07
Nodes (58): Replication, ReplicationConfigResult, CLIArg(), cliArg(), CLIScriptPath(), Transport, shellScriptPath(), mateChannel() (+50 more)

### Community 48 - "eqArgs"
Cohesion: 0.24
Nodes (11): callIndex(), TestManagerLogsCLIShell(), TestManagerPrepHostRootlessUsesUnshareChown(), TestPreflightRunsBeforeAnything(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime(), TestManagerReachableProbesRuntimeThenCompose() (+3 more)

### Community 49 - "go_pkg_strings"
Cohesion: 0.12
Nodes (25): go_pkg_bytes, go_pkg_context, go_pkg_errors, go_pkg_fmt, go_pkg_io, go_pkg_net, go_pkg_os, go_pkg_path_filepath (+17 more)

### Community 50 - "sempmate.go"
Cohesion: 0.36
Nodes (7): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, go_pkg_encoding_xml, encoding/xml.Name, xmlEscape()

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (33): applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias(), TestAllowCommandApprovesAWrappedRuntime() (+25 more)

### Community 52 - "opRunner"
Cohesion: 0.14
Nodes (19): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch(), opCanI() (+11 more)

### Community 53 - "runRootWith"
Cohesion: 0.09
Nodes (44): go_pkg_runtime, App, runRootWith(), TestCLICommand(), TestDeployOperatorNoPromptStaysUnknownFlag(), TestK8sConfigDeleteDomainCertsConfigured(), TestK8sConfigDeleteDomainCertsFromDirs(), TestK8sPromptsNameNamespaceAndContext() (+36 more)

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

### Community 58 - "Data replication"
Cohesion: 0.40
Nodes (5): Configuring a site, Data replication, Reaching the mate, Switching roles, When a run fails partway

### Community 59 - "semp.go"
Cohesion: 0.16
Nodes (16): backupTarget, Credential, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigFlag(), curlConfigLine(), Ops, httpBody() (+8 more)

### Community 60 - "cli_test.go"
Cohesion: 0.08
Nodes (66): os.File, testing.M, allowRuntime(), capture(), captureStderr(), collectPaths(), echoRunner(), fakeBinaryOnPath() (+58 more)

### Community 61 - ".configRows"
Cohesion: 0.12
Nodes (23): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), allCustomMounted(), containsString(), failRow() (+15 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.17
Nodes (36): failOnCall(), fakeEnv(), Manager, healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+28 more)

### Community 64 - "output_test.go"
Cohesion: 0.24
Nodes (14): sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLevelOrderKeepsSkipBelowFail(), TestLineAddsNoPrefix(), TestSectionPadsToWidth() (+6 more)

### Community 65 - "Configuration"
Cohesion: 0.15
Nodes (13): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+5 more)

### Community 66 - "cpuSetRange"
Cohesion: 0.50
Nodes (4): cpuSetCount(), cpuSetRange(), TestCPUSetCount(), TestCPUSetRange()

### Community 67 - "broker/replication_test.go"
Cohesion: 0.12
Nodes (30): colSpan, dashSpans(), flagByte(), gutterClear(), ParseVPNReplication(), SameEndpoints(), sliceSpan(), mateAppliance() (+22 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "blocks.go"
Cohesion: 0.14
Nodes (23): Region, token, markerRegion(), TestMarkerRegionBeginAndEnd(), brokerTypeFromSchema(), firstQuoted(), Capture, isIndented() (+15 more)

### Community 71 - "semp_test.go"
Cohesion: 0.08
Nodes (45): TestLocalMateRunsThroughOps(), bridgeHostPort(), sempPort(), sempV1OK(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv() (+37 more)

### Community 72 - "childExit"
Cohesion: 0.67
Nodes (3): childExit(), TestChildExitKeepsItsMessage(), TestExitCodeIsNeverNegative()

### Community 73 - "internal/k8s"
Cohesion: 0.11
Nodes (19): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go, names_test.go (+11 more)

### Community 74 - "annotate_test.go"
Cohesion: 0.09
Nodes (41): applyMetaField(), Capture, isMarker(), markerSection(), markerVerb(), parseKV(), readMarker(), renderKV() (+33 more)

### Community 75 - "go_pkg_github_com_spf13_cobra"
Cohesion: 0.11
Nodes (12): childExitError, usageError, go_pkg_github_com_spf13_cobra, go_pkg_os_exec, TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveMissingBinaryIsActionable() (+4 more)

### Community 76 - "namespace_test.go"
Cohesion: 0.30
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 77 - "Ops"
Cohesion: 0.16
Nodes (5): validCLILine(), validName(), Ops, BaseName(), TestBaseNameSplitsOnBothSeparators()

### Community 78 - "newEchoMgr"
Cohesion: 0.13
Nodes (15): bytes.Buffer, NewManager(), Manager, newEchoMgr(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployPodmanDryRunHidesSecretBytes(), TestManagerDockerCheckProbesCompose() (+7 more)

### Community 79 - "healthCmd"
Cohesion: 0.67
Nodes (3): HealthCheck, healthCmd(), TestHealthCmdDefaultsToReadiness()

### Community 80 - "Platforms"
Cohesion: 0.10
Nodes (36): Example, go_pkg_embed, go_pkg_gopkg_in_yaml_v3, go_pkg_solace_internal_abbrev, TestExamplesEmitsToStdout(), TestPlatformConstantsMatchSchemaSections(), TestExecBinariesCoversEveryPlatform(), TestGuardConfigIsValid() (+28 more)

### Community 81 - "k8s/inspect_test.go"
Cohesion: 0.15
Nodes (19): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+11 more)

### Community 83 - "limits_test.go"
Cohesion: 0.13
Nodes (27): go_pkg_slices, Manager, limitsMgr(), nrOpenProbe(), TestCheckLimitsNrOpenAloneNeedsNoRestart(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager() (+19 more)

### Community 84 - "haCfg"
Cohesion: 0.11
Nodes (27): TestResolveStorageClass(), TestStorageRows(), TestDeleteBrokerProbesClaimsOnlyWhenPurging(), ProductKeyRoles(), haCfg(), TestProductKeyRoles(), operatorDeployJSON(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade() (+19 more)

### Community 85 - "compareVersions"
Cohesion: 0.40
Nodes (5): compareVersions(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestOperatorVersionWarningFiresOnlyOnADowngrade()

### Community 86 - "k8s/inspect.go"
Cohesion: 0.14
Nodes (22): time.Time, anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), Cluster, normalizeToList(), TestConditionLevelDegradesSafely() (+14 more)

### Community 87 - ".applyContainerDefaults"
Cohesion: 0.12
Nodes (14): scalingTier, applyContainerBlockDefaults(), Config, podmanRunUser(), setDefault(), setDefaultCmd(), setDefaultInt(), xdgConfigHome() (+6 more)

### Community 89 - "render_test.go"
Cohesion: 0.12
Nodes (41): BrokerCR(), Compose(), composeEscape(), ComposeProject(), ContainerSecrets(), SecretPreflight(), assertNoCheckoutPath(), envLines() (+33 more)

### Community 90 - "vulnjudge/main.go"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.09
Nodes (39): Disposition, SectionRule, go_pkg_reflect, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware() (+31 more)

### Community 92 - "cli/replication.go"
Cohesion: 0.19
Nodes (19): mateChannelFunc, go_pkg_solace_internal_k8s, bufio.Reader, App, step(), confirmReplicationConfig(), confirmReplicationSwitch(), App (+11 more)

### Community 95 - "scaling.go"
Cohesion: 0.33
Nodes (7): scalingKey, scalingSpelling, Scaling, Scaling, yaml.Node, scalingKeyIndex(), scalingKeyList()

### Community 96 - "testing.T"
Cohesion: 0.05
Nodes (125): testing.T, cliRunNames(), Ops, hasCall(), matchCLI(), newTestOps(), outputForRole(), ranContains() (+117 more)

### Community 97 - "hostpath.go"
Cohesion: 0.18
Nodes (10): expandTilde(), expandTildeToken(), Config, HasPathSeparator(), IsAbsHostPath(), isPathSep(), TestExpandTilde(), TestHasPathSeparator() (+2 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "github.com/spf13/cobra.Command"
Cohesion: 0.07
Nodes (117): opFunc, roleOpFunc, go_pkg_solace_internal_convert, go_pkg_solace_internal_examples, github.com/spf13/cobra.Command, github.com/spf13/cobra.PositionalArgs, addCommands(), addLogFlags() (+109 more)

### Community 100 - "limits.go"
Cohesion: 0.22
Nodes (9): go_pkg_math, delegateSetting(), Manager, limitText(), missingControllers(), parseLimit(), parseUnitProps(), TestMissingControllers() (+1 more)

### Community 101 - "scripts.go"
Cohesion: 0.06
Nodes (52): showCmd, countContains(), field(), assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript() (+44 more)

### Community 102 - "command_test.go"
Cohesion: 0.15
Nodes (14): decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRenamedRuntimeKeysFailLoud() (+6 more)

### Community 103 - "container/secrets_test.go"
Cohesion: 0.23
Nodes (18): ResolveSecretValues(), certFixture(), TestDeletePodmanRemovesTheBundleAfterTheUnit(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeWritingTheComposeFile(), TestDeployDockerPassesTheBundleAsEnvNotArgv(), TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(), TestDeployPodmanFailsWhenTheBundleCannotBeWritten() (+10 more)

### Community 104 - "watch_test.go"
Cohesion: 0.10
Nodes (25): imageFromDeployment(), desiredWatch(), Cluster, splitWatch(), subtractWatch(), deployJSON(), Cluster, operatorItem() (+17 more)

### Community 105 - "internal/cli"
Cohesion: 0.17
Nodes (12): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, examples_test.go, exit_test.go (+4 more)

### Community 106 - "prep_test.go"
Cohesion: 0.22
Nodes (15): adminCfg(), TestCreateNamespaceApplyFails(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree(), TestCreateSecretsPreflight(), TestCreateSecretsStopsOnPreflightFailure(), TestDeleteNamespaceStopsOnPreflightFailure(), TestDeleteSecrets() (+7 more)

### Community 107 - "replicationops_test.go"
Cohesion: 0.15
Nodes (30): mateConvergenceShutdowns(), PlannedRoles(), replicationVPNLines(), RoleAtSite(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching() (+22 more)

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "execguard_test.go"
Cohesion: 0.21
Nodes (11): Config, guardCommandOf(), setGuardCommand(), TestAllowCommandsAccepts(), TestAllowCommandsRejects(), TestAllowCommandsRejectsEscalation(), TestAllowedBinaryIsNotGloballyAllowed(), TestComposeCommandDerivation() (+3 more)

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "statusreport_test.go"
Cohesion: 0.17
Nodes (19): ownsPod(), Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList() (+11 more)

### Community 112 - "container/transport_test.go"
Cohesion: 0.19
Nodes (13): capCall, capRunner, New(), NewTransport(), dockerCfg(), podmanCfg(), TestTransportCopy(), TestTransportEchoHidesSEMPConfig() (+5 more)

### Community 113 - "internal/config"
Cohesion: 0.13
Nodes (15): command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go, hostpath_test.go, imagepullsecret_test.go (+7 more)

### Community 114 - "go_pkg_testing"
Cohesion: 0.22
Nodes (7): go_pkg_testing, CanonicalDuration(), TestCanonicalDuration(), Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), unapprovedCtrCfg()

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - ".stateRows"
Cohesion: 0.29
Nodes (6): containerState, healthState, Manager, printable(), TestPrintableStripsControlCharacters(), KV

### Community 119 - "hostpath_test.go"
Cohesion: 0.13
Nodes (21): guardConfig(), CheckHostPath(), Config, setContainerName(), TestCheckHostPathAccepts(), TestCheckHostPathRejects(), TestContainerCertRequiresKey(), TestContainerHostDirsExpandATilde() (+13 more)

### Community 121 - "Ops"
Cohesion: 0.13
Nodes (12): time.Duration, Ops, confirmImport(), App, pluralVPN(), runExport(), runImport(), TestPluralVPNWording() (+4 more)

### Community 122 - "ReplRole"
Cohesion: 0.16
Nodes (8): AdminState, ReplRole, sempMate, rejectionIn(), setReplicationRoleScript(), Ops, sempAdminState(), sempReplRole()

### Community 123 - "container/inspect_test.go"
Cohesion: 0.15
Nodes (19): decodeInspect(), orUnknown(), inspectReply(), kvValue(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings(), TestInspectNameLosesDockersLeadingSlash() (+11 more)

### Community 125 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 126 - "storage_test.go"
Cohesion: 0.33
Nodes (9): Config, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 130 - "completion.go"
Cohesion: 0.18
Nodes (15): github.com/spf13/cobra.ShellCompDirective, completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), App, isEnvFileName(), matching() (+7 more)

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
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 328 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **5 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `github.com/spf13/cobra.Command` to `config_test.go`, `Manager`, `Config`, `Command`, `convert_test.go`, `ctrCfg`, `validate.go`, `Load`, `opRunner`, `semp_test.go`, `newEchoMgr`, `Platforms`, `limits_test.go`, `.applyContainerDefaults`, `render_test.go`, `cli/replication.go`, `container/secrets_test.go`, `execguard_test.go`, `container/transport_test.go`, `go_pkg_testing`, `hostpath_test.go`, `Ops`, `containerTransport`?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `Set`, `completion.go`, `Manager`, `Config`, `ops_container.go`, `switchplan_test.go`, `driver.go`, `.ConfigureReplication`, `importops.go`, `validate.go`, `semp.go`, `Ops`, `haCfg`, `cli/replication.go`, `testing.T`, `scripts.go`, `Ops`, `containerTransport`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `Manager`, `operator.go`, `NewCluster`, `eqArgs`, `convert_test.go`, `ctrCfg`, `k8s/secrets_test.go`, `render.go`, `Role`, `Quadlet`, `ReplSite`, `eqArgs`, `opRunner`, `cli_test.go`, `.configRows`, `Cluster`, `semp_test.go`, `namespace_test.go`, `newEchoMgr`, `haCfg`, `render_test.go`, `cli/replication.go`, `testing.T`, `container/secrets_test.go`, `watch_test.go`, `prep_test.go`, `container/transport_test.go`, `go_pkg_testing`, `Ops`, `containerTransport`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 42 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 42 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _222 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Set` be split into smaller, more focused modules?**
  _Cohesion score 0.13538461538461538 - nodes in this community are weakly interconnected._
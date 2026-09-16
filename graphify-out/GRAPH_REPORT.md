# Graph Report - solace-cnt-scripts  (2026-09-17)

## Corpus Check
- 189 files · ~561,637 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 22 file(s) not represented in the graph (top: .golden 16, (none) 2, .cli 2)

## Summary
- 3619 nodes · 12115 edges · 135 communities (125 shown, 10 thin omitted)
- Extraction: 83% EXTRACTED · 17% INFERRED · 0% AMBIGUOUS · INFERRED: 2060 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `6fa13a03`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- Get
- Sink
- context.Context
- runRoot
- testing.T
- Commands
- Manager
- config.go
- operator.go
- Command
- container/transport_test.go
- NewCluster
- eqArgs
- convert_test.go
- Block
- ctrCfg
- ParseBlocks
- k8s/secrets_test.go
- dev.sh
- Load
- dev.ps1
- render.go
- switchplan_test.go
- verify_local_test.go
- .releaseToBackup
- driver.go
- ops_container.go
- importops.go
- ReplSite
- solace
- validate.go
- CLAUDE.md
- importops_test.go
- domaincerts_resolve_test.go
- Role
- .resolveSecretRefs
- recRunner
- go_pkg_testing
- replicationVPNLines
- broker/matechannel_test.go
- cli/platform_test.go
- inject.go
- ExitCode
- commanddoc_test.go
- runner_test.go
- Config
- k8s/matechannel_test.go
- confirmAction
- go_pkg_strings
- blocks_test.go
- newRootCmd
- execguard_test.go
- exportconfig_test.go
- Operations
- Test catalogue
- Exporting and importing configuration
- DiffBlocks
- sempmate.go
- semp.go
- cli_test.go
- checkreport.go
- rootless_test.go
- Cluster
- output_test.go
- Configuration
- sempmate_test.go
- broker/replication.go
- Command reference
- Developer guide
- blocks.go
- semp_test.go
- usagef
- internal/k8s
- annotate_test.go
- Set
- namespace_test.go
- newEchoMgr
- Platform
- desiredWatch
- tlssecret_test.go
- k8s/inspect_test.go
- age
- limits_test.go
- operatorversion.go
- ContainerSecret
- k8s/inspect.go
- prep_test.go
- github.com/spf13/cobra.Command
- render_test.go
- judge
- RuleFor
- cli/replication.go
- load.go
- hostpath.go
- .decodeScalingEntry
- newTestOps
- Cluster
- Abbreviations
- commands.go
- .checkNrOpen
- scripts.go
- kubectlTransport
- container/secrets_test.go
- watch_test.go
- internal/cli
- haCfg
- replicationops_test.go
- What `import-config` applies
- k8s/runtime_test.go
- internal/broker
- statusreport_test.go
- time.Time
- internal/config
- .ConfigureReplication
- logArgs
- platformTitle
- internal/container
- .stateRows
- hostpath_test.go
- .ResolveNode
- Ops
- VPNRepl
- container/inspect_test.go
- cli/replication_test.go
- Fixtures and doubles
- storage_test.go
- TestCanonicalDuration
- TestServerCert
- config/replication_test.go
- containerTransport
- .preflightOne
- Troubleshooting
- Cluster
- pullSecretCfg

## God Nodes (most connected - your core abstractions)
1. `newTestOps()` - 129 edges
2. `ctrCfg()` - 120 edges
3. `Role` - 116 edges
4. `newCapMgr()` - 104 edges
5. `Config` - 94 edges
6. `NewCluster()` - 92 edges
7. `loadK8s()` - 85 edges
8. `matchCLI()` - 81 edges
9. `Platform` - 74 edges
10. `bg()` - 68 edges

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

## Communities (135 total, 10 thin omitted)

### Community 0 - "bg"
Cohesion: 0.16
Nodes (48): bg(), domainCANames(), App, k8sCluster(), k8sContext(), k8sLogin(), k8sOps(), k8sWhat() (+40 more)

### Community 1 - "Get"
Cohesion: 0.13
Nodes (26): Example, go_pkg_embed, parseError(), DetectPlatforms(), PlatformAbbrev(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint(), TestDetectPlatformsMissingFile() (+18 more)

### Community 2 - "Sink"
Cohesion: 0.16
Nodes (9): solaceRows(), TestSolaceRowsRefusesAnEngineReturnedNameBackIntoArgv(), columnWidths(), KV, Level, Row, Sink, pad() (+1 more)

### Community 3 - "context.Context"
Cohesion: 0.07
Nodes (12): Exec, context.Context, os/exec.Cmd, Manager, idMapCovers(), origin(), runUserIDs(), Cluster (+4 more)

### Community 4 - "runRoot"
Cohesion: 0.09
Nodes (40): runRoot(), TestBashEnvGivenToEnvFlag(), TestConfigureServerCertsRefusesASecretItDoesNotOwn(), TestConvertErrorPaths(), TestConvertParseError(), TestConvertToFile(), TestConvertToStdout(), TestConvertWriteError() (+32 more)

### Community 5 - "testing.T"
Cohesion: 0.05
Nodes (101): testing.T, TestConfirmDeleteShortcut(), TestCtrManagerConfirmWiring(), TestPromptYes(), TestPromptYesNo(), decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs() (+93 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (6): composeNeedsSecretValues(), exactName(), Manager, orNone(), secretSummary(), setOrMissing()

### Community 8 - "config.go"
Cohesion: 0.08
Nodes (27): AdditionalUser, Broker, CertDir, Container, ContainerSecurity, DockerConfig, DomainCerts, Image (+19 more)

### Community 9 - "operator.go"
Cohesion: 0.22
Nodes (15): go_pkg_text_template, GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator(), renderOperatorNS() (+7 more)

### Community 10 - "Command"
Cohesion: 0.11
Nodes (27): cmdField, commandRules, guardedCmd, go_pkg_unicode, go_pkg_unicode_utf8, Command, checkBinary(), CheckCommand() (+19 more)

### Community 11 - "container/transport_test.go"
Cohesion: 0.11
Nodes (29): capCall, capRunner, New(), TestNewDefaults(), Transport, NewManager(), TestManagerLogsCLIShell(), TestManagerNilSinks() (+21 more)

### Community 12 - "NewCluster"
Cohesion: 0.08
Nodes (69): TestCheckStopsProbingWhenUnreachable(), TestStorageRows(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun() (+61 more)

### Community 13 - "eqArgs"
Cohesion: 0.17
Nodes (23): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), TestCLIAndShellAreInteractive() (+15 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (66): doc, Result, segment, vars, DecodeStrict(), Config, boolOf(), commentSafe() (+58 more)

### Community 15 - "Block"
Cohesion: 0.09
Nodes (52): Block, TargetState, regexp.Regexp, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects() (+44 more)

### Community 16 - "ctrCfg"
Cohesion: 0.09
Nodes (83): TestRestartCountComesFromSystemdOnPodman(), TestRestartCountUnknownWhenSystemctlFails(), fileExists(), assertMode(), containsStr(), ctrCfg(), hasCall(), maskedKeys() (+75 more)

### Community 17 - "ParseBlocks"
Cohesion: 0.11
Nodes (20): StripMarkers(), TestStripMarkersKeepsPragmaAndSectionComments(), TestStripMarkersNoMarkersPassthrough(), ParseBlocks(), splitLines(), TestParseBlocksCRLFMatchesLF(), TestParseBlocksIndentedLineWithNoOpenerRefused(), TestParseBlocksMissingEndTerminatorRefused() (+12 more)

### Community 18 - "k8s/secrets_test.go"
Cohesion: 0.11
Nodes (33): go_pkg_encoding_base64, GenBroker(), GenSecrets(), joinManifests(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret() (+25 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "Load"
Cohesion: 0.08
Nodes (45): scalingTier, minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors() (+37 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "render.go"
Cohesion: 0.10
Nodes (36): WeightedNodeTerm, strings.Builder, mdRow(), writeAbbrevTable(), LoadBalancer, NodeAffinity, NodeMatchExpr, Placement (+28 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.19
Nodes (29): SiteState, BuildSwitchPlan(), ExecuteSwitchPlan(), actionStrings(), ordersToA(), phaseActions(), replSites(), replVPNs() (+21 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.14
Nodes (44): cliScriptPath(), curlCalls(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn() (+36 more)

### Community 25 - ".releaseToBackup"
Cohesion: 0.12
Nodes (18): countContains(), field(), TestCountContains(), TestField(), TestLastLines(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), TestLastLinesEqualCount() (+10 more)

### Community 26 - "driver.go"
Cohesion: 0.15
Nodes (22): chunk, chunkResult, runCLISkeleton(), chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment() (+14 more)

### Community 27 - "ops_container.go"
Cohesion: 0.14
Nodes (41): go_pkg_solace_internal_container, childExit(), TestChildExitKeepsItsMessage(), wantRemove(), containerRenderRole(), containerRole(), containerWhat(), ctrLogin() (+33 more)

### Community 28 - "importops.go"
Cohesion: 0.16
Nodes (21): ImportResult, PlannedSection, defaultVPNFirst(), ImportPlan, importIgnore(), orNone(), preambleForApply(), quoteList() (+13 more)

### Community 29 - "ReplSite"
Cohesion: 0.20
Nodes (15): PhaseKind, SwitchAction, SwitchPhase, MateChannel, Ops, confirmDemotions(), SwitchPlan, promoteConfirmed() (+7 more)

### Community 31 - "validate.go"
Cohesion: 0.11
Nodes (20): keyValueEntries, checkCredentialChars(), foldToEnvVar(), Config, invertedCPUSetRange(), missingErr(), platformKey(), requireAll() (+12 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "importops_test.go"
Cohesion: 0.29
Nodes (13): planSections(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, newTeardownApplyFixture(), TestImportOpsImportApplySeparatesFirstSectionFromMain() (+5 more)

### Community 34 - "domaincerts_resolve_test.go"
Cohesion: 0.07
Nodes (53): DirReader, fakeDirEntry, go_pkg_path, os.DirEntry, os.FileInfo, os.FileMode, caNameSafe(), checkNoDotDot() (+45 more)

### Community 35 - "Role"
Cohesion: 0.12
Nodes (14): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+6 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.13
Nodes (10): TestCanIAnswerReadsTheLastLine(), NewTransport(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest() (+2 more)

### Community 38 - "go_pkg_testing"
Cohesion: 0.09
Nodes (21): childExitError, usageError, go_pkg_bytes, go_pkg_flag, go_pkg_github_com_spf13_cobra, go_pkg_gopkg_in_yaml_v3, go_pkg_io, go_pkg_os (+13 more)

### Community 39 - "replicationVPNLines"
Cohesion: 0.22
Nodes (15): PlannedRoles(), reenableLine(), replicationVPNLines(), RoleAtSite(), siteAEntry(), TestNothingEverShutsDownTheVPNItself(), TestPlannedRolesNamesEveryListedVPN(), TestReplicationVPNLinesEnablesListedAndDisablesTheRest() (+7 more)

### Community 40 - "broker/matechannel_test.go"
Cohesion: 0.19
Nodes (19): CLIRunner, fakeRun, Ops, NewCLIMate(), mateReplies(), newTestMate(), TestLocalMateRunsThroughOps(), TestMateChannelDescribeNamesTheTarget() (+11 more)

### Community 41 - "cli/platform_test.go"
Cohesion: 0.29
Nodes (16): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+8 more)

### Community 42 - "inject.go"
Cohesion: 0.16
Nodes (28): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "ExitCode"
Cohesion: 0.16
Nodes (13): TestExecute(), ExitCode(), childStatusError(), k8sEnv(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitStatusIsScopedToInteractiveSessions(), TestExitCodeContract(), TestExitCodeIsNeverNegative() (+5 more)

### Community 44 - "commanddoc_test.go"
Cohesion: 0.16
Nodes (21): shorthand, github.com/spf13/pflag.FlagSet, renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeResolutionNotes(), writeShorthandTable() (+13 more)

### Community 45 - "runner_test.go"
Cohesion: 0.07
Nodes (35): interactiveFailRunner, runFailRunner, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), verboseExec(), Echo (+27 more)

### Community 46 - "Config"
Cohesion: 0.10
Nodes (17): TLS, Config, Scaling, BrokerPodSuffixShape(), allCustomMounted(), OperatorNamespace(), Cluster, HARoles() (+9 more)

### Community 47 - "k8s/matechannel_test.go"
Cohesion: 0.24
Nodes (23): Runner, execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf(), replCfg() (+15 more)

### Community 48 - "confirmAction"
Cohesion: 0.19
Nodes (24): exportconfigReadCounter, layer, go_pkg_bufio, io.Reader, io.Writer, TestStdinCanAnswerClosedFile(), confirmAction(), confirmActionStrict() (+16 more)

### Community 49 - "go_pkg_strings"
Cohesion: 0.12
Nodes (23): go_pkg_context, go_pkg_encoding_json, go_pkg_errors, go_pkg_fmt, go_pkg_math, go_pkg_net, go_pkg_regexp, go_pkg_solace_internal_broker (+15 more)

### Community 50 - "blocks_test.go"
Cohesion: 0.22
Nodes (15): go_pkg_reflect, TestAnnotateRoundTripCaptureIsMarked(), newBlock(), blockByIndex(), Capture, loadSample(), TestNewBlockMessageSpoolQualifierHasNoOwnName(), TestNewBlockOpener() (+7 more)

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (32): applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias(), TestAllowCommandApprovesAWrappedRuntime() (+24 more)

### Community 52 - "execguard_test.go"
Cohesion: 0.19
Nodes (12): Config, guardCommandOf(), setGuardCommand(), TestAllowCommandsAccepts(), TestAllowCommandsRejects(), TestAllowedBinaryIsNotGloballyAllowed(), TestComposeCommandDerivation(), TestComposeDerivationInheritsRejection() (+4 more)

### Community 53 - "exportconfig_test.go"
Cohesion: 0.07
Nodes (43): opCall, opRunner, os.File, capture(), captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect() (+35 more)

### Community 54 - "Operations"
Cohesion: 0.09
Nodes (23): Bringing up a fresh cluster, Configuring a site, Data replication, Docker and Podman mechanics, Every destructive command confirms, Exit codes, Extra CLI users differ by platform, Operations (+15 more)

### Community 55 - "Test catalogue"
Cohesion: 0.10
Nodes (20): abbrev_test.go, convert_test.go, Coverage, examples_test.go, internal/abbrev, internal/convert, internal/engine, internal/examples (+12 more)

### Community 56 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 57 - "DiffBlocks"
Cohesion: 0.12
Nodes (35): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, ancestorReported(), blockLabel(), DiffBlocks() (+27 more)

### Community 58 - "sempmate.go"
Cohesion: 0.36
Nodes (7): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, go_pkg_encoding_xml, encoding/xml.Name, xmlEscape()

### Community 59 - "semp.go"
Cohesion: 0.13
Nodes (22): backupTarget, Credential, TestHTTPStatusHelpers(), anyHTTP2xx(), bridgeHostPort(), curlConfigFlag(), curlConfigLine(), Ops (+14 more)

### Community 60 - "cli_test.go"
Cohesion: 0.08
Nodes (81): testing.M, allowRuntime(), captureStderr(), echoRunner(), fakeBinaryOnPath(), firstLine(), App, runCtr() (+73 more)

### Community 61 - "checkreport.go"
Cohesion: 0.15
Nodes (20): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), containsString(), failRow(), Cluster (+12 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.16
Nodes (37): failOnCall(), fakeEnv(), Manager, healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirRefusesAnUnwritableParent(), TestCheckIDMappingRefusesAnUnmappedRunUser() (+29 more)

### Community 64 - "output_test.go"
Cohesion: 0.24
Nodes (14): sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLevelOrderKeepsSkipBelowFail(), TestLineAddsNoPrefix(), TestSectionPadsToWidth() (+6 more)

### Community 65 - "Configuration"
Cohesion: 0.15
Nodes (13): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+5 more)

### Community 66 - "sempmate_test.go"
Cohesion: 0.29
Nodes (13): newSEMPTestMate(), sempReply(), sempSite(), TestSEMPMateChecksExecuteResult(), TestSEMPMateEscapesTheVPNName(), TestSEMPMatePlainTextURL(), TestSEMPMatePrimaryActive(), TestSEMPMateReadsBothMateShapes() (+5 more)

### Community 67 - "broker/replication.go"
Cohesion: 0.09
Nodes (56): colSpan, MateConfig, cliTransport(), containsEndpoint(), dashSpans(), flagByte(), gutterClear(), labelValue() (+48 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "blocks.go"
Cohesion: 0.14
Nodes (21): Omission, Region, token, isMarker(), markerRegion(), TestIsMarkerDiscriminatesNamespace(), brokerTypeFromSchema(), firstQuoted() (+13 more)

### Community 71 - "semp_test.go"
Cohesion: 0.14
Nodes (25): assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK(), TestBackupRevertActivitySuccess(), TestBackupRevertActivityTLSIgnoresTLSCAs() (+17 more)

### Community 72 - "usagef"
Cohesion: 0.10
Nodes (29): go_pkg_solace_internal_convert, go_pkg_solace_internal_examples, github.com/spf13/cobra.PositionalArgs, noRolePositional(), registerFlagCompletion(), App, newConvertCmd(), runConvert() (+21 more)

### Community 73 - "internal/k8s"
Cohesion: 0.11
Nodes (19): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go, names_test.go (+11 more)

### Community 74 - "annotate_test.go"
Cohesion: 0.10
Nodes (38): Annotate(), applyMetaField(), Capture, markerSection(), markerVerb(), parseKV(), readMarker(), renderKV() (+30 more)

### Community 75 - "Set"
Cohesion: 0.14
Nodes (15): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+7 more)

### Community 76 - "namespace_test.go"
Cohesion: 0.30
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 77 - "newEchoMgr"
Cohesion: 0.15
Nodes (13): bytes.Buffer, Manager, newEchoMgr(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployPodmanDryRunHidesSecretBytes(), TestManagerDockerCheckProbesCompose(), TestManagerLifecycleDockerDryRunUsesCompose() (+5 more)

### Community 78 - "Platform"
Cohesion: 0.28
Nodes (18): go_pkg_github_com_spf13_pflag, checkFlagPlatforms(), commandPlatforms(), declaredList(), App, parsePlatformList(), platformSuffix(), prepare() (+10 more)

### Community 79 - "desiredWatch"
Cohesion: 0.23
Nodes (8): desiredWatch(), Cluster, splitWatch(), TestDesiredWatchAppendsTheBrokerNamespace(), TestWatchFromContainersReadsEveryAllNamespacesSpelling(), watchFromContainers(), watchFromDeployment(), watchPlan

### Community 80 - "tlssecret_test.go"
Cohesion: 0.43
Nodes (7): Config, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsMakeTheSecretOurs(), TestSuppliedCertsNeedASecretName(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), tlsCfg()

### Community 81 - "k8s/inspect_test.go"
Cohesion: 0.18
Nodes (17): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+9 more)

### Community 82 - "age"
Cohesion: 0.30
Nodes (6): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), TestOperatorRunningImageWithNoContainers()

### Community 83 - "limits_test.go"
Cohesion: 0.19
Nodes (18): go_pkg_slices, Manager, limitsMgr(), nrOpenProbe(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager(), TestCheckLimitsRootlessReportsTheSoftLimitWithoutGating() (+10 more)

### Community 84 - "operatorversion.go"
Cohesion: 0.36
Nodes (7): compareVersions(), imageTag(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestImageTag(), TestOperatorVersionWarningFiresOnlyOnADowngrade()

### Community 86 - "k8s/inspect.go"
Cohesion: 0.19
Nodes (18): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), ownedPods(), brokerList, brokerStatus (+10 more)

### Community 87 - "prep_test.go"
Cohesion: 0.25
Nodes (13): adminCfg(), TestCreateNamespaceApplyFails(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree(), TestCreateSecretsPreflight(), TestCreateSecretsStopsOnPreflightFailure(), TestDeleteNamespaceStopsOnPreflightFailure(), TestDeleteSecrets() (+5 more)

### Community 88 - "github.com/spf13/cobra.Command"
Cohesion: 0.16
Nodes (28): github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, addCommands(), addLogFlags(), newOperatorCmd(), newOperatorDeployCmd(), newOperatorGenerateCmd(), newOperatorLogsCmd() (+20 more)

### Community 89 - "render_test.go"
Cohesion: 0.10
Nodes (50): ContainerNoFile(), NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), BrokerCR(), Compose(), composeEscape(), ComposeProject(), ContainerSecrets() (+42 more)

### Community 90 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.10
Nodes (38): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+30 more)

### Community 92 - "cli/replication.go"
Cohesion: 0.16
Nodes (22): mateChannelFunc, go_pkg_solace_internal_k8s, bufio.Reader, App, lineSink(), progress(), step(), confirmReplicationConfig() (+14 more)

### Community 93 - "load.go"
Cohesion: 0.16
Nodes (13): TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config, podmanRunUser() (+5 more)

### Community 94 - "hostpath.go"
Cohesion: 0.15
Nodes (12): BaseName(), expandTilde(), expandTildeToken(), Config, HasPathSeparator(), IsAbsHostPath(), isPathSep(), TestBaseNameSplitsOnBothSeparators() (+4 more)

### Community 95 - ".decodeScalingEntry"
Cohesion: 0.31
Nodes (7): scalingKey, scalingSpelling, Scaling, Scaling, yaml.Node, scalingKeyIndex(), scalingKeyList()

### Community 96 - "newTestOps"
Cohesion: 0.04
Nodes (106): Ops, matchCLI(), newTestOps(), ranContains(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultUsersRefusesAQuotedVPNName() (+98 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "commands.go"
Cohesion: 0.20
Nodes (47): opFunc, roleOpFunc, addPodFlag(), App, group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerConfigureCmd() (+39 more)

### Community 100 - ".checkNrOpen"
Cohesion: 0.31
Nodes (5): Manager, limitText(), parseLimit(), parseUnitProps(), TestParseLimit()

### Community 101 - "scripts.go"
Cohesion: 0.06
Nodes (48): showCmd, cliRunNames(), TestValidName(), validCLILine(), validName(), Ops, TestImportOpsImportVerifyReExportsAndDiffs(), assertLeaderScript() (+40 more)

### Community 103 - "container/secrets_test.go"
Cohesion: 0.23
Nodes (18): ResolveSecretValues(), certFixture(), TestDeletePodmanRemovesTheBundleAfterTheUnit(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeWritingTheComposeFile(), TestDeployDockerPassesTheBundleAsEnvNotArgv(), TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(), TestDeployPodmanFailsWhenTheBundleCannotBeWritten() (+10 more)

### Community 104 - "watch_test.go"
Cohesion: 0.17
Nodes (17): imageFromDeployment(), subtractWatch(), deployJSON(), Cluster, operatorItem(), TestAmbiguousOperatorIsNotFoldedIntoNotInstalled(), TestFindOperatorDeploymentIsScopedByNamespace(), TestOperatorReleaseNarrowsInsteadOfRemoving() (+9 more)

### Community 105 - "internal/cli"
Cohesion: 0.17
Nodes (12): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, examples_test.go, exit_test.go (+4 more)

### Community 106 - "haCfg"
Cohesion: 0.16
Nodes (18): TestResolveStorageClass(), TestDeleteBrokerProbesClaimsOnlyWhenPurging(), haCfg(), saCfg(), TestHARoles(), TestProductKeyRoles(), TestRestartOrder(), operatorDeployJSON() (+10 more)

### Community 107 - "replicationops_test.go"
Cohesion: 0.27
Nodes (17): hasCall(), TestDisableDefaultUsersShowVPNError(), TestReleaseToBackupReleasedTimeout(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), TestConfigureReplicationClosingShowFailureStillReportsTheApply() (+9 more)

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "k8s/runtime_test.go"
Cohesion: 0.43
Nodes (7): TestClusterHonoursRuntime(), TestExecutorRefusesUnapprovedRuntime(), TestRuntimeDefaultArgvUnchanged(), TestTransportHonoursRuntime(), unapprovedCfg(), withLeading(), wrappedCfg()

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "statusreport_test.go"
Cohesion: 0.17
Nodes (19): ownsPod(), Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList() (+11 more)

### Community 112 - "time.Time"
Cohesion: 0.29
Nodes (4): time.Time, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

### Community 113 - "internal/config"
Cohesion: 0.13
Nodes (15): command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go, hostpath_test.go, imagepullsecret_test.go (+7 more)

### Community 114 - ".ConfigureReplication"
Cohesion: 0.14
Nodes (13): BrokerType, cliMate, bannerType(), Ops, ReplicationConfigResult, isRunCLIRejection(), mateConvergenceShutdowns(), missingListedVPNs() (+5 more)

### Community 115 - "logArgs"
Cohesion: 0.67
Nodes (4): logArgs(), App, logArgs2(), TestLogArgsBuildsOneSetForBothPlatforms()

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - ".stateRows"
Cohesion: 0.33
Nodes (5): containerState, healthState, Manager, printable(), TestPrintableStripsControlCharacters()

### Community 119 - "hostpath_test.go"
Cohesion: 0.12
Nodes (23): guardConfig(), CheckHostPath(), Config, setContainerName(), TestCheckHostPathAccepts(), TestCheckHostPathRejects(), TestContainerCertRequiresKey(), TestContainerHostDirsExpandATilde() (+15 more)

### Community 121 - "Ops"
Cohesion: 0.10
Nodes (17): time.Duration, Ops, TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), App, runExport(), runImport() (+9 more)

### Community 122 - "VPNRepl"
Cohesion: 0.10
Nodes (11): AdminState, QueueState, ReplRole, scriptedMate, sempMate, VPNRepl, rejectionIn(), setReplicationRoleScript() (+3 more)

### Community 123 - "container/inspect_test.go"
Cohesion: 0.16
Nodes (17): decodeInspect(), orUnknown(), inspectReply(), kvValue(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings(), TestInspectNameLosesDockersLeadingSlash() (+9 more)

### Community 124 - "cli/replication_test.go"
Cohesion: 0.38
Nodes (13): mateChannel(), mateSEMPPassword(), App, replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer(), TestMateChannelPicksTheMechanismTheSiteDeclares() (+5 more)

### Community 125 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 126 - "storage_test.go"
Cohesion: 0.33
Nodes (9): Config, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 128 - "TestServerCert"
Cohesion: 0.16
Nodes (14): TestPathHelpers(), TestServerCert(), TestServerCertBundleOrder(), TestServerCertBundleRefusesAKeyBearingCert(), TestServerCertBundleReportsAnUnreadableFile(), TestServerCertBundleRequiresBothHalves(), ServerCertBundle(), serverCertFile() (+6 more)

### Community 129 - "config/replication_test.go"
Cohesion: 0.27
Nodes (12): Replication, Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationAbsent() (+4 more)

### Community 131 - ".preflightOne"
Cohesion: 0.50
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 138 - "pullSecretCfg"
Cohesion: 0.39
Nodes (8): Config, pullSecretCfg(), TestImagePullSecretDerivedSuffixFitsWhereAdditionalUsersAlreadyFits(), TestImagePullSecretNameBringYourOwn(), TestImagePullSecretNameConfiguredNameWins(), TestImagePullSecretNameDerivedWhenCredentialsButNoName(), TestImagePullSecretNameEmptyWithNeither(), TestManagesImagePullSecretNeedsBothCredentials()

## Knowledge Gaps
- **222 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+217 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 328 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **10 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Config` connect `Config` to `TestServerCert`, `config/replication_test.go`, `containerTransport`, `Manager`, `config.go`, `operator.go`, `container/transport_test.go`, `NewCluster`, `convert_test.go`, `ctrCfg`, `k8s/secrets_test.go`, `render.go`, `ReplSite`, `recRunner`, `k8s/matechannel_test.go`, `exportconfig_test.go`, `semp.go`, `cli_test.go`, `checkreport.go`, `Cluster`, `semp_test.go`, `namespace_test.go`, `newEchoMgr`, `desiredWatch`, `prep_test.go`, `render_test.go`, `cli/replication.go`, `newTestOps`, `kubectlTransport`, `container/secrets_test.go`, `haCfg`, `k8s/runtime_test.go`, `Ops`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **Why does `Platform` connect `Platform` to `Get`, `containerTransport`, `testing.T`, `Manager`, `config.go`, `Command`, `container/transport_test.go`, `convert_test.go`, `ctrCfg`, `Load`, `validate.go`, `execguard_test.go`, `exportconfig_test.go`, `semp.go`, `newEchoMgr`, `limits_test.go`, `render_test.go`, `cli/replication.go`, `load.go`, `commands.go`, `container/secrets_test.go`, `platformTitle`, `hostpath_test.go`, `Ops`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `containerTransport`, `testing.T`, `Manager`, `config.go`, `ParseBlocks`, `.releaseToBackup`, `driver.go`, `ops_container.go`, `importops.go`, `validate.go`, `recRunner`, `broker/matechannel_test.go`, `Config`, `k8s/matechannel_test.go`, `semp.go`, `Set`, `cli/replication.go`, `hostpath.go`, `scripts.go`, `kubectlTransport`, `.ConfigureReplication`, `.ResolveNode`, `Ops`?**
  _High betweenness centrality (0.024) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 41 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 41 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _222 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Get` be split into smaller, more focused modules?**
  _Cohesion score 0.12962962962962962 - nodes in this community are weakly interconnected._
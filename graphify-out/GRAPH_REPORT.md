# Graph Report - solace-cnt-scripts  (2026-09-17)

## Corpus Check
- 189 files · ~561,341 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 22 file(s) not represented in the graph (top: .golden 16, (none) 2, .cli 2)

## Summary
- 3575 nodes · 11229 edges · 133 communities (127 shown, 6 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 2060 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `8b14e464`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- completion_test.go
- Sink
- context.Context
- cli_test.go
- config_test.go
- Commands
- Manager
- config.go
- GenOperator
- Command
- capRunner
- NewCluster
- eqArgs
- convert_test.go
- Block
- ctrCfg
- runRootWith
- k8s/secrets_test.go
- dev.sh
- Load
- dev.ps1
- render.go
- switchplan_test.go
- verify_local_test.go
- .releaseToBackup
- renderDriver
- ops_container.go
- importops.go
- solace
- validate.go
- CLAUDE.md
- importops_test.go
- domaincerts_resolve_test.go
- Role
- .resolveSecretRefs
- recRunner
- Echo
- .ConfigureReplication
- broker/matechannel_test.go
- cli/platform_test.go
- InjectShutdown
- exit_test.go
- commanddoc_test.go
- testing.T
- haCfg
- ReplSite
- confirmAction
- Ops
- Quadlet
- newRootCmd
- execguard_test.go
- opRunner
- Operations
- Test catalogue
- Exporting and importing configuration
- DiffBlocks
- VPNRepl
- .BackupRevertActivity
- .Run
- checkreport.go
- rootless_test.go
- Cluster
- Configuration
- scaling_test.go
- broker/replication.go
- Command reference
- Developer guide
- blocks.go
- localCfg
- usagef
- internal/k8s
- annotate_test.go
- Set
- namespace_test.go
- newTestOps
- Platform
- exportconfig_test.go
- New
- k8s/inspect_test.go
- age
- limitsMgr
- Config
- ContainerSecrets
- k8s/inspect.go
- prep_test.go
- github.com/spf13/cobra.Command
- load
- judge
- RuleFor
- opPerformReplication
- load.go
- hostpath.go
- scaling.go
- matchCLI
- Abbreviations
- commands.go
- .checkNrOpen
- scripts.go
- kubectlTransport
- container/secrets_test.go
- watch_test.go
- internal/cli
- operatorversion.go
- tlsCfg
- What `import-config` applies
- omit_test.go
- internal/broker
- k8s/runtime_test.go
- validContainerConfig
- internal/config
- BrokerType
- ResolveEnvPath
- .Preflight
- internal/container
- .stateRows
- hostpath_test.go
- completion.go
- runExport
- ParseRole
- container/inspect_test.go
- logArgs
- Fixtures and doubles
- storage_test.go
- TestImportOpsExportConfigScopeSelectsCLICommand
- TestServerCert
- childExit
- Troubleshooting
- pullSecretCfg
- Data replication

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
- `main()` --calls--> `ExitCode()`  [EXTRACTED]
  main.go → internal/cli/exit.go
- `parseBody()` --calls--> `isMarker()`  [INFERRED]
  internal/broker/blocks.go → internal/broker/annotate.go
- `parseBody()` --calls--> `readMarker()`  [INFERRED]
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

## Communities (133 total, 6 thin omitted)

### Community 0 - "bg"
Cohesion: 0.15
Nodes (49): TestK8sLoginOutcomes(), bg(), domainCANames(), App, k8sCluster(), k8sContext(), k8sLogin(), k8sOps() (+41 more)

### Community 1 - "completion_test.go"
Cohesion: 0.19
Nodes (18): runComplete(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion(), TestCompletionHelpStillWorks(), TestCompletionNeedsAShell(), TestCompletionNoDescriptions() (+10 more)

### Community 2 - "Sink"
Cohesion: 0.19
Nodes (6): columnWidths(), Level, Row, Sink, pad(), TestLevelTagsMapToTheHouseTags()

### Community 3 - "context.Context"
Cohesion: 0.06
Nodes (13): Exec, context.Context, os/exec.Cmd, Manager, idMapCovers(), origin(), runUserIDs(), Cluster (+5 more)

### Community 4 - "cli_test.go"
Cohesion: 0.08
Nodes (55): os.File, allowRuntime(), capture(), captureStderr(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone() (+47 more)

### Community 5 - "config_test.go"
Cohesion: 0.07
Nodes (42): assertContainerBlockDefaults(), assertContainerScaling(), Config, haNodesConfig(), TestAdditionalUserNamesDoNotCollideOnKubernetes(), TestAdditionalUserPasswordCharsAreFreeOnKubernetes(), TestApplyDefaultsDocker(), TestApplyDefaultsK8s() (+34 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (8): composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary(), setOrMissing(), solaceRows()

### Community 8 - "config.go"
Cohesion: 0.08
Nodes (29): AdditionalUser, Broker, CertDir, Container, ContainerSecurity, DockerConfig, DomainCerts, Network (+21 more)

### Community 9 - "GenOperator"
Cohesion: 0.16
Nodes (22): GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator(), renderOperatorNS(), splitAfterNamespace() (+14 more)

### Community 10 - "Command"
Cohesion: 0.12
Nodes (26): cmdField, commandRules, guardedCmd, decodeCommand(), TestCommandUnmarshal(), Command, checkBinary(), CheckCommand() (+18 more)

### Community 11 - "capRunner"
Cohesion: 0.13
Nodes (24): capCall, capRunner, New(), TestNewDefaults(), Transport, Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime() (+16 more)

### Community 12 - "NewCluster"
Cohesion: 0.09
Nodes (59): TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts(), TestValidateReportsTheDerivedImagePullSecretName() (+51 more)

### Community 13 - "eqArgs"
Cohesion: 0.17
Nodes (23): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), TestCLIAndShellAreInteractive() (+15 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (63): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+55 more)

### Community 15 - "Block"
Cohesion: 0.09
Nodes (50): Block, TargetState, regexp.Regexp, injectedBlock(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects(), ClearTargetVirtualHostnames() (+42 more)

### Community 16 - "ctrCfg"
Cohesion: 0.07
Nodes (98): fileExists(), NewManager(), assertMode(), containsStr(), ctrCfg(), Manager, hasCall(), newCapMgr() (+90 more)

### Community 17 - "runRootWith"
Cohesion: 0.12
Nodes (33): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused(), TestSmuggledSubcommandIsRefused() (+25 more)

### Community 18 - "k8s/secrets_test.go"
Cohesion: 0.11
Nodes (32): GenBroker(), GenSecrets(), joinManifests(), namespaceManifest(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret() (+24 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "Load"
Cohesion: 0.14
Nodes (26): minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors(), TestLoadSuccess() (+18 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "render.go"
Cohesion: 0.18
Nodes (23): strings.Builder, Placement, PodAffinityTerm, boolStr(), cut(), parsePort(), parseToleration(), sortedKeys() (+15 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.13
Nodes (40): PhaseKind, SiteState, SwitchAction, SwitchPhase, MateChannel, BuildSwitchPlan(), confirmDemotions(), ExecuteSwitchPlan() (+32 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.15
Nodes (40): cliScriptPath(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestDetectRoleAddrsError(), TestDetectRoleAmbiguousAddress() (+32 more)

### Community 25 - ".releaseToBackup"
Cohesion: 0.11
Nodes (19): field(), TestField(), TestLastLines(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), TestLastLinesEqualCount(), showRedundancyLocalScript(), defaultLocalAddrs() (+11 more)

### Community 26 - "renderDriver"
Cohesion: 0.14
Nodes (23): chunk, chunkResult, runCLISkeleton(), chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment() (+15 more)

### Community 27 - "ops_container.go"
Cohesion: 0.16
Nodes (39): TestContainerRoleDetectsFromHostname(), wantRemove(), containerRenderRole(), containerRole(), containerWhat(), ctrLogin(), ctrManager(), ctrOps() (+31 more)

### Community 28 - "importops.go"
Cohesion: 0.13
Nodes (23): ImportResult, PlannedSection, checkProvenance(), checkSameType(), defaultVPNFirst(), describeScope(), ImportPlan, Ops (+15 more)

### Community 31 - "validate.go"
Cohesion: 0.11
Nodes (21): keyValueEntries, BrokerPodSuffixShape(), checkCredentialChars(), foldToEnvVar(), Config, invertedCPUSetRange(), missingErr(), platformKey() (+13 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.16
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "importops_test.go"
Cohesion: 0.10
Nodes (31): importIgnore(), planSections(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops, minimalCapture() (+23 more)

### Community 34 - "domaincerts_resolve_test.go"
Cohesion: 0.07
Nodes (52): DirReader, fakeDirEntry, os.DirEntry, os.FileInfo, os.FileMode, caNameSafe(), checkNoDotDot(), DefaultDirReader() (+44 more)

### Community 35 - "Role"
Cohesion: 0.06
Nodes (25): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+17 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.14
Nodes (10): TestCanIAnswerReadsTheLastLine(), NewTransport(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs(), TestTransportUpload(), TestTransportUploadQuotesDest() (+2 more)

### Community 38 - "Echo"
Cohesion: 0.16
Nodes (11): interactiveFailRunner, TestExecIsSilentWithoutVerbose(), TestResolveMissingBinaryIsActionable(), Echo, EnvRunner, Runner, NewExec(), Quote() (+3 more)

### Community 39 - ".ConfigureReplication"
Cohesion: 0.13
Nodes (28): ReplicationConfigResult, isRunCLIRejection(), mateConvergenceShutdowns(), missingListedVPNs(), newlineIf(), PlannedRoles(), replicationVPNLines(), replPhase1Rejected() (+20 more)

### Community 40 - "broker/matechannel_test.go"
Cohesion: 0.19
Nodes (19): CLIRunner, fakeRun, Ops, NewCLIMate(), mateReplies(), newTestMate(), TestLocalMateRunsThroughOps(), TestMateChannelDescribeNamesTheTarget() (+11 more)

### Community 41 - "cli/platform_test.go"
Cohesion: 0.29
Nodes (16): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+8 more)

### Community 42 - "InjectShutdown"
Cohesion: 0.10
Nodes (44): lineRole, serviceLine, shutdownStyle, span, svcKey, Example, classifyServiceRest(), containsPortCommand() (+36 more)

### Community 43 - "exit_test.go"
Cohesion: 0.24
Nodes (11): runFailRunner, childStatusError(), k8sEnv(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitStatusIsScopedToInteractiveSessions(), TestEveryArgValidatorIsAUsageError(), TestExitCodeContract(), TestHelperExitProcess() (+3 more)

### Community 44 - "commanddoc_test.go"
Cohesion: 0.16
Nodes (23): shorthand, github.com/spf13/pflag.FlagSet, mdRow(), renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeAbbrevTable() (+15 more)

### Community 45 - "testing.T"
Cohesion: 0.06
Nodes (67): bytes.Buffer, testing.T, TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRenamedRuntimeKeysFailLoud() (+59 more)

### Community 46 - "haCfg"
Cohesion: 0.14
Nodes (21): TestCheckStopsProbingWhenUnreachable(), TestResolveStorageClass(), TestStorageRows(), lbServiceName(), haCfg(), saCfg(), TestHARoles(), TestPodNameSuffixMatchesTheConfigBound() (+13 more)

### Community 47 - "ReplSite"
Cohesion: 0.10
Nodes (49): Replication, mateChannel(), mateSEMPPassword(), App, replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer() (+41 more)

### Community 48 - "confirmAction"
Cohesion: 0.17
Nodes (26): exportconfigReadCounter, layer, io.Reader, io.Writer, TestConfirmDowngradeNeverUnattended(), TestConfirmLayerMatrix(), TestConfirmNonTTY(), TestPromptKeepsBufferedAnswersForLaterQuestions() (+18 more)

### Community 49 - "Ops"
Cohesion: 0.09
Nodes (13): time.Duration, cliRunNames(), countContains(), Ops, TestCountContains(), TestValidName(), validCLILine(), validName() (+5 more)

### Community 50 - "Quadlet"
Cohesion: 0.14
Nodes (20): ContainerNoFile(), NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), composeEscape(), EnvPairs(), escapePercent(), groupKey(), itoa() (+12 more)

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (26): TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias(), TestAllowCommandIsRegisteredWhereItExecutes(), TestAllowCommandIsRepeatable() (+18 more)

### Community 52 - "execguard_test.go"
Cohesion: 0.18
Nodes (13): Config, guardCommandOf(), setGuardCommand(), TestAllowCommandsAccepts(), TestAllowCommandsRejects(), TestAllowCommandsRejectsEscalation(), TestAllowedBinaryIsNotGloballyAllowed(), TestComposeCommandDerivation() (+5 more)

### Community 53 - "opRunner"
Cohesion: 0.13
Nodes (19): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch(), opCanI() (+11 more)

### Community 54 - "Operations"
Cohesion: 0.11
Nodes (18): Bringing up a fresh cluster, Docker and Podman mechanics, Every destructive command confirms, Exit codes, Extra CLI users differ by platform, Operations, Post-deployment configuration order, Re-deploying is safe and explicit (+10 more)

### Community 55 - "Test catalogue"
Cohesion: 0.10
Nodes (20): abbrev_test.go, convert_test.go, Coverage, examples_test.go, internal/abbrev, internal/convert, internal/engine, internal/examples (+12 more)

### Community 56 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 57 - "DiffBlocks"
Cohesion: 0.12
Nodes (35): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, ancestorReported(), blockLabel(), DiffBlocks() (+27 more)

### Community 58 - "VPNRepl"
Cohesion: 0.09
Nodes (18): AdminState, QueueState, ReplRole, scriptedMate, sempExecuteResult, sempMate, sempRedundancyReply, sempReplicationReply (+10 more)

### Community 59 - ".BackupRevertActivity"
Cohesion: 0.16
Nodes (16): backupTarget, Credential, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigFlag(), curlConfigLine(), Ops, httpBody() (+8 more)

### Community 60 - ".Run"
Cohesion: 0.20
Nodes (19): testing.M, runCtr(), TestConfigStepsDoNotLeakSecrets(), TestCtrConfigDryRun(), TestCtrDiagnosticsDryRun(), TestCtrErrorPaths(), TestCtrExecCLIPathSeparator(), TestCtrLoginOutcomes() (+11 more)

### Community 61 - "checkreport.go"
Cohesion: 0.19
Nodes (16): additionalUsersRow(), containsString(), failRow(), Cluster, info(), okRow(), portRows(), pskRow() (+8 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.18
Nodes (35): failOnCall(), fakeEnv(), Manager, healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirRefusesAnUnwritableParent(), TestCheckIDMappingRefusesAnUnmappedRunUser() (+27 more)

### Community 63 - "Cluster"
Cohesion: 0.13
Nodes (6): time.Time, Cluster, Cluster, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

### Community 65 - "Configuration"
Cohesion: 0.15
Nodes (13): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+5 more)

### Community 66 - "scaling_test.go"
Cohesion: 0.15
Nodes (17): decodeScaling(), Config, setContainerCPUSet(), setContainerMem(), TestApplyScalingTierDefaultsContainerBlocks(), TestApplyScalingTierDefaultsCPUSetPerTier(), TestApplyScalingTierDefaultsK8s(), TestApplyScalingTierDefaultsMemOverride() (+9 more)

### Community 67 - "broker/replication.go"
Cohesion: 0.08
Nodes (56): colSpan, MateConfig, cliTransport(), containsEndpoint(), dashSpans(), flagByte(), gutterClear(), labelValue() (+48 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "blocks.go"
Cohesion: 0.08
Nodes (43): Omission, Region, token, markerRegion(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks(), brokerTypeFromSchema(), firstQuoted() (+35 more)

### Community 71 - "localCfg"
Cohesion: 0.09
Nodes (46): bridgeHostPort(), sempPort(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK() (+38 more)

### Community 72 - "usagef"
Cohesion: 0.09
Nodes (29): childExitError, usageError, github.com/spf13/cobra.PositionalArgs, noRolePositional(), registerFlagCompletion(), App, newConvertCmd(), runConvert() (+21 more)

### Community 73 - "internal/k8s"
Cohesion: 0.11
Nodes (19): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go, names_test.go (+11 more)

### Community 74 - "annotate_test.go"
Cohesion: 0.09
Nodes (42): Annotate(), applyMetaField(), Capture, isMarker(), markerSection(), markerVerb(), parseKV(), readMarker() (+34 more)

### Community 75 - "Set"
Cohesion: 0.14
Nodes (14): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+6 more)

### Community 76 - "namespace_test.go"
Cohesion: 0.30
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 77 - "newTestOps"
Cohesion: 0.08
Nodes (47): Ops, newTestOps(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultUsersRefusesAQuotedVPNName(), TestDisableDefaultVPN(), TestDomainCerts(), TestDomainCertsAcceptsAFullHostPath() (+39 more)

### Community 78 - "Platform"
Cohesion: 0.14
Nodes (32): checkFlagPlatforms(), commandPlatforms(), declaredList(), flagOnlyOn(), App, parsePlatformList(), platformSuffix(), prepare() (+24 more)

### Community 79 - "exportconfig_test.go"
Cohesion: 0.18
Nodes (21): ExitCode(), exportconfigDriverOutput(), exportconfigTransportOutput(), exportconfigVPNListOutput(), App, newExportconfigRunner(), newExportconfigRunnerRejectingChunk(), runImportConfirmation() (+13 more)

### Community 80 - "New"
Cohesion: 0.12
Nodes (22): TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), TestCheckReportSkipsEmptySections(), New(), NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey() (+14 more)

### Community 81 - "k8s/inspect_test.go"
Cohesion: 0.16
Nodes (18): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestConditionLevelDegradesSafely() (+10 more)

### Community 82 - "age"
Cohesion: 0.23
Nodes (8): orNone(), orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), age(), roleRank(), Cluster

### Community 83 - "limitsMgr"
Cohesion: 0.21
Nodes (17): Manager, limitsMgr(), nrOpenProbe(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager(), TestCheckLimitsRootlessReportsTheSoftLimitWithoutGating(), TestCheckLimitsSkipsOnAnEUIDMismatch() (+9 more)

### Community 84 - "Config"
Cohesion: 0.15
Nodes (6): Image, TLS, atoiPrefix(), Config, Scaling, OperatorNamespace()

### Community 85 - "ContainerSecrets"
Cohesion: 0.15
Nodes (10): ResolveSecretValues(), ContainerSecrets(), containerSecretSpecs(), ContainerSecret, secretFilePath(), SecretPreflight(), TestFileBackedSecretIsExemptFromSecretPreflight(), TestPodmanNeverGetsAFileBackedSecret() (+2 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.24
Nodes (16): anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), brokerList, brokerStatus, condition, containerSpec (+8 more)

### Community 87 - "prep_test.go"
Cohesion: 0.20
Nodes (16): adminCfg(), TestCreateNamespaceApplyFails(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree(), TestCreateSecretsFailsWithoutAdminFields(), TestCreateSecretsPreflight(), TestCreateSecretsStopsOnPreflightFailure(), TestDeleteNamespaceStopsOnPreflightFailure() (+8 more)

### Community 88 - "github.com/spf13/cobra.Command"
Cohesion: 0.26
Nodes (20): github.com/spf13/cobra.Command, applyAliases(), addCommands(), addLogFlags(), newOperatorCmd(), newOperatorDeployCmd(), newOperatorGenerateCmd(), newOperatorLogsCmd() (+12 more)

### Community 89 - "load"
Cohesion: 0.15
Nodes (31): BrokerCR(), Compose(), ComposeProject(), healthCmd(), assertNoCheckoutPath(), healthCheckFixture(), load(), TestArtifactsCarryNoSecrets() (+23 more)

### Community 90 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.10
Nodes (38): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+30 more)

### Community 92 - "opPerformReplication"
Cohesion: 0.18
Nodes (21): mateChannelFunc, bufio.Reader, TestWarnAndStep(), App, lineSink(), progress(), step(), warn() (+13 more)

### Community 93 - "load.go"
Cohesion: 0.20
Nodes (12): TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config, podmanRunUser() (+4 more)

### Community 94 - "hostpath.go"
Cohesion: 0.18
Nodes (10): expandTilde(), expandTildeToken(), Config, HasPathSeparator(), IsAbsHostPath(), isPathSep(), TestExpandTilde(), TestHasPathSeparator() (+2 more)

### Community 95 - "scaling.go"
Cohesion: 0.13
Nodes (16): scalingKey, scalingSpelling, scalingTier, containerMem(), cpuSetRange(), Config, Scaling, Scaling (+8 more)

### Community 96 - "matchCLI"
Cohesion: 0.06
Nodes (62): hasCall(), matchCLI(), ranContains(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultUsersShowVPNError(), TestDisableDefaultVPNDisableError() (+54 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "commands.go"
Cohesion: 0.22
Nodes (43): opFunc, roleOpFunc, addPodFlag(), App, group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerConfigureCmd() (+35 more)

### Community 100 - ".checkNrOpen"
Cohesion: 0.30
Nodes (5): Manager, limitText(), parseLimit(), parseUnitProps(), TestParseLimit()

### Community 101 - "scripts.go"
Cohesion: 0.08
Nodes (41): showCmd, assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript(), enableDefaultUsersScript(), enableDefaultVPNScript() (+33 more)

### Community 103 - "container/secrets_test.go"
Cohesion: 0.21
Nodes (16): certFixture(), TestDeletePodmanRemovesTheBundleAfterTheUnit(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeWritingTheComposeFile(), TestDeployDockerPassesTheBundleAsEnvNotArgv(), TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(), TestDeployPodmanFailsWhenTheBundleCannotBeWritten(), TestDeployPodmanWritesTheBundleBeforeTheUnit() (+8 more)

### Community 104 - "watch_test.go"
Cohesion: 0.11
Nodes (25): imageFromDeployment(), desiredWatch(), Cluster, splitWatch(), subtractWatch(), deployJSON(), Cluster, operatorItem() (+17 more)

### Community 105 - "internal/cli"
Cohesion: 0.17
Nodes (12): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, examples_test.go, exit_test.go (+4 more)

### Community 106 - "operatorversion.go"
Cohesion: 0.36
Nodes (7): compareVersions(), imageTag(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestImageTag(), TestOperatorVersionWarningFiresOnlyOnADowngrade()

### Community 107 - "tlsCfg"
Cohesion: 0.43
Nodes (7): Config, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsMakeTheSecretOurs(), TestSuppliedCertsNeedASecretName(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), tlsCfg()

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "omit_test.go"
Cohesion: 0.40
Nodes (4): TestImportPlanAcceptsWhatExportProduced(), TestImportPlanRefusesAFileItDidNotExport(), TestOmitAnnotateRoundTripsOmissions(), TestOmitExportConfigEndToEnd()

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "k8s/runtime_test.go"
Cohesion: 0.43
Nodes (7): TestClusterHonoursRuntime(), TestExecutorRefusesUnapprovedRuntime(), TestRuntimeDefaultArgvUnchanged(), TestTransportHonoursRuntime(), unapprovedCfg(), withLeading(), wrappedCfg()

### Community 112 - "validContainerConfig"
Cohesion: 0.10
Nodes (25): TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes(), TestAdditionalUserNamesCollideOnDocker(), TestValidateAdditionalUserClashesWithABuiltIn(), TestValidateContainerBadNetworkMode(), TestValidateContainerBridge(), TestValidateContainerHA(), TestValidateContainerIdentifiers() (+17 more)

### Community 113 - "internal/config"
Cohesion: 0.13
Nodes (15): command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go, hostpath_test.go, imagepullsecret_test.go (+7 more)

### Community 114 - "BrokerType"
Cohesion: 0.23
Nodes (6): BrokerType, cliMate, bannerType(), Ops, showReplicationScript(), showVPNReplicationScript()

### Community 115 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - ".stateRows"
Cohesion: 0.11
Nodes (12): containerState, healthState, Manager, orUnknown(), printable(), orValue(), bareKind(), Cluster (+4 more)

### Community 119 - "hostpath_test.go"
Cohesion: 0.11
Nodes (23): CanonicalDuration(), TestCanonicalDuration(), guardConfig(), CheckHostPath(), Config, setContainerName(), TestCheckHostPathAccepts(), TestCheckHostPathRejects() (+15 more)

### Community 120 - "completion.go"
Cohesion: 0.31
Nodes (10): github.com/spf13/cobra.ShellCompDirective, completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), completionShell(), App, isEnvFileName() (+2 more)

### Community 121 - "runExport"
Cohesion: 0.25
Nodes (10): confirmImport(), App, pluralVPN(), runExport(), runImport(), TestPluralVPNWording(), emitOrWrite(), nowStamp() (+2 more)

### Community 122 - "ParseRole"
Cohesion: 0.36
Nodes (7): TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), TestRoleNames(), ParseRole(), RoleAbbrev(), RoleNames()

### Community 123 - "container/inspect_test.go"
Cohesion: 0.16
Nodes (18): decodeInspect(), inspectReply(), kvValue(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings(), TestInspectNameLosesDockersLeadingSlash(), TestInspectRejectsAnEmptyReply() (+10 more)

### Community 124 - "logArgs"
Cohesion: 0.67
Nodes (4): logArgs(), App, logArgs2(), TestLogArgsBuildsOneSetForBothPlatforms()

### Community 125 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 126 - "storage_test.go"
Cohesion: 0.33
Nodes (9): Config, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 127 - "TestImportOpsExportConfigScopeSelectsCLICommand"
Cohesion: 0.67
Nodes (3): TestImportOpsExportConfigScopeSelectsCLICommand(), TestImportOpsImportVerifyReExportsAndDiffs(), currentConfigScript()

### Community 128 - "TestServerCert"
Cohesion: 0.15
Nodes (15): TestDiagnostics(), TestPathHelpers(), TestServerCert(), writeFile(), TestDiagnosticsMkdirError(), TestServerCertCAReadError(), TestServerCertRunCLIError(), TestServerCertUploadError() (+7 more)

### Community 129 - "childExit"
Cohesion: 0.67
Nodes (3): childExit(), TestChildExitKeepsItsMessage(), TestExitCodeIsNeverNegative()

### Community 132 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 138 - "pullSecretCfg"
Cohesion: 0.39
Nodes (8): Config, pullSecretCfg(), TestImagePullSecretDerivedSuffixFitsWhereAdditionalUsersAlreadyFits(), TestImagePullSecretNameBringYourOwn(), TestImagePullSecretNameConfiguredNameWins(), TestImagePullSecretNameDerivedWhenCredentialsButNoName(), TestImagePullSecretNameEmptyWithNeither(), TestManagesImagePullSecretNeedsBothCredentials()

### Community 143 - "Data replication"
Cohesion: 0.40
Nodes (5): Configuring a site, Data replication, Reaching the mate, Switching roles, When a run fails partway

## Knowledge Gaps
- **222 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+217 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 331 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **6 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `config_test.go`, `Manager`, `config.go`, `Command`, `capRunner`, `convert_test.go`, `ctrCfg`, `Load`, `validate.go`, `Role`, `Ops`, `execguard_test.go`, `opRunner`, `scaling_test.go`, `localCfg`, `limitsMgr`, `ContainerSecrets`, `load`, `opPerformReplication`, `load.go`, `scaling.go`, `commands.go`, `validContainerConfig`, `hostpath_test.go`?**
  _High betweenness centrality (0.040) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `cli_test.go`, `Manager`, `config.go`, `GenOperator`, `capRunner`, `NewCluster`, `convert_test.go`, `ctrCfg`, `k8s/secrets_test.go`, `render.go`, `Role`, `recRunner`, `haCfg`, `ReplSite`, `Ops`, `Quadlet`, `opRunner`, `checkreport.go`, `Cluster`, `localCfg`, `namespace_test.go`, `newTestOps`, `ContainerSecrets`, `prep_test.go`, `load`, `opPerformReplication`, `kubectlTransport`, `container/secrets_test.go`, `watch_test.go`, `k8s/runtime_test.go`?**
  _High betweenness centrality (0.037) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `Manager`, `config.go`, `.releaseToBackup`, `renderDriver`, `ops_container.go`, `importops.go`, `validate.go`, `recRunner`, `.ConfigureReplication`, `broker/matechannel_test.go`, `ReplSite`, `Ops`, `.BackupRevertActivity`, `Set`, `opPerformReplication`, `scripts.go`, `kubectlTransport`, `BrokerType`, `runExport`, `ParseRole`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 41 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 41 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _222 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `context.Context` be split into smaller, more focused modules?**
  _Cohesion score 0.06398730830248546 - nodes in this community are weakly interconnected._
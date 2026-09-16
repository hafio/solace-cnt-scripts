# Graph Report - solace-cnt-scripts  (2026-09-17)

## Corpus Check
- 189 files · ~561,637 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 22 file(s) not represented in the graph (top: .golden 16, (none) 2, .cli 2)

## Summary
- 3576 nodes · 11233 edges · 139 communities (128 shown, 11 thin omitted)
- Extraction: 82% EXTRACTED · 18% INFERRED · 0% AMBIGUOUS · INFERRED: 2060 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `c951b765`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- Get
- Sink
- context.Context
- runRoot
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
- captureStderr
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
- haCfg
- Echo
- .ConfigureReplication
- broker/matechannel_test.go
- cli/platform_test.go
- InjectShutdown
- exit_test.go
- commanddoc_test.go
- testing.T
- podName
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
- .rpc
- .BackupRevertActivity
- cli_test.go
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
- asUsage
- internal/k8s
- annotate_test.go
- Set
- namespace_test.go
- newTestOps
- Platform
- runRootWith
- github.com/spf13/cobra.Command
- k8s/inspect_test.go
- age
- limitsMgr
- Config
- ContainerSecret
- k8s/inspect.go
- prep_test.go
- commands.go
- load
- judge
- RuleFor
- opPerformReplication
- load.go
- hostpath.go
- scaling.go
- matchCLI
- Abbreviations
- platformOps
- .checkNrOpen
- scripts.go
- kubectlTransport
- container/secrets_test.go
- watch_test.go
- internal/cli
- operatorversion_test.go
- replicationops_test.go
- What `import-config` applies
- omit_test.go
- internal/broker
- statusreport_test.go
- redundancy_test.go
- internal/config
- BrokerType
- ResolveEnvPath
- manager.go
- internal/container
- .stateRows
- hostpath_test.go
- completion.go
- Ops
- VPNRepl
- container/inspect_test.go
- cli/replication_test.go
- Fixtures and doubles
- storage_test.go
- Manager
- TestServerCert
- config/replication_test.go
- containerTransport
- .preflightOne
- Troubleshooting
- Cluster
- verify_local.go
- helpers.go
- TestCheckStopsProbingWhenUnreachable
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

## Communities (139 total, 11 thin omitted)

### Community 0 - "bg"
Cohesion: 0.16
Nodes (48): bg(), domainCANames(), App, k8sCluster(), k8sContext(), k8sLogin(), k8sOps(), k8sWhat() (+40 more)

### Community 1 - "Get"
Cohesion: 0.15
Nodes (25): Example, TestExamplesEmitsToStdout(), parseError(), DetectPlatforms(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint(), TestDetectPlatformsMissingFile(), TestParsePlatform() (+17 more)

### Community 2 - "Sink"
Cohesion: 0.18
Nodes (7): columnWidths(), KV, Level, Row, Sink, pad(), TestLevelTagsMapToTheHouseTags()

### Community 3 - "context.Context"
Cohesion: 0.08
Nodes (9): Exec, context.Context, os/exec.Cmd, Cluster, Cluster, Cluster, Cluster, badJSONRunner (+1 more)

### Community 4 - "runRoot"
Cohesion: 0.09
Nodes (40): runRoot(), TestBashEnvGivenToEnvFlag(), TestConfigureServerCertsRefusesASecretItDoesNotOwn(), TestConvertErrorPaths(), TestConvertParseError(), TestConvertRoundTrip(), TestConvertToFile(), TestConvertToStdout() (+32 more)

### Community 5 - "config_test.go"
Cohesion: 0.06
Nodes (55): assertContainerBlockDefaults(), assertContainerScaling(), Config, haNodesConfig(), TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes(), TestAdditionalUserNamesCollideOnDocker(), TestAdditionalUserNamesDoNotCollideOnKubernetes() (+47 more)

### Community 6 - "Commands"
Cohesion: 0.04
Nodes (53): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker (+45 more)

### Community 8 - "config.go"
Cohesion: 0.08
Nodes (30): AdditionalUser, Broker, CertDir, Container, ContainerSecurity, DockerConfig, DomainCerts, Network (+22 more)

### Community 9 - "GenOperator"
Cohesion: 0.27
Nodes (13): GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), operatorProbes(), RenderOperator(), renderOperatorNS(), splitAfterNamespace() (+5 more)

### Community 10 - "Command"
Cohesion: 0.13
Nodes (24): cmdField, commandRules, guardedCmd, Command, checkBinary(), CheckCommand(), checkFlagShape(), checkToken() (+16 more)

### Community 11 - "capRunner"
Cohesion: 0.13
Nodes (24): capCall, capRunner, New(), TestNewDefaults(), Transport, Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime() (+16 more)

### Community 12 - "NewCluster"
Cohesion: 0.07
Nodes (70): TestStorageRows(), TestValidateConfigSectionNeverFails(), TestValidateGroupsAndOrdersSections(), TestValidateNeverPrintsASecret(), TestValidateReadsDeploymentsOnce(), TestValidateReportsDefaultPorts(), TestValidateReportsEveryFailureInOneRun(), TestValidateReportsResolvedPorts() (+62 more)

### Community 13 - "eqArgs"
Cohesion: 0.17
Nodes (23): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), TestCLIAndShellAreInteractive() (+15 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (63): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+55 more)

### Community 15 - "Block"
Cohesion: 0.09
Nodes (52): Block, TargetState, regexp.Regexp, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects() (+44 more)

### Community 16 - "ctrCfg"
Cohesion: 0.07
Nodes (102): TestRestartCountUnknownWhenSystemctlFails(), fileExists(), NewManager(), assertMode(), containsStr(), ctrCfg(), Manager, hasCall() (+94 more)

### Community 17 - "captureStderr"
Cohesion: 0.14
Nodes (21): os.File, capture(), captureStderr(), fakeBinaryOnPath(), runStandalone(), runStatusStderr(), TestBinaryAnnouncementWiring(), TestDeployBrokerDoesNotApplyOperator() (+13 more)

### Community 18 - "k8s/secrets_test.go"
Cohesion: 0.12
Nodes (30): GenSecrets(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), checkGolden(), decodeDataValue(), TestAdditionalUsersSecretCarriesBothHalves() (+22 more)

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
Nodes (23): strings.Builder, boolStr(), containerSecretSpecs(), cut(), parsePort(), parseToleration(), sortedKeys(), splitUser() (+15 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.13
Nodes (40): PhaseKind, SiteState, SwitchAction, SwitchPhase, MateChannel, BuildSwitchPlan(), confirmDemotions(), ExecuteSwitchPlan() (+32 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.15
Nodes (40): cliScriptPath(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestDetectRoleAddrsError(), TestDetectRoleAmbiguousAddress() (+32 more)

### Community 25 - ".releaseToBackup"
Cohesion: 0.15
Nodes (13): field(), TestField(), TestLastLines(), TestPrimaryRedundancyUp(), showRedundancyLocalScript(), Ops, TestBackupActivityStateReadsTheMateColumn(), activity() (+5 more)

### Community 26 - "renderDriver"
Cohesion: 0.18
Nodes (19): chunk, chunkResult, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), sanitizeComment(), shQuote() (+11 more)

### Community 27 - "ops_container.go"
Cohesion: 0.15
Nodes (39): TestCtrManagerConfirmWiring(), wantRemove(), containerRenderRole(), containerRole(), containerWhat(), ctrLogin(), ctrManager(), ctrOps() (+31 more)

### Community 28 - "importops.go"
Cohesion: 0.13
Nodes (22): ImportResult, PlannedSection, checkProvenance(), checkSameType(), defaultVPNFirst(), describeScope(), ImportPlan, Ops (+14 more)

### Community 31 - "validate.go"
Cohesion: 0.10
Nodes (22): keyValueEntries, TestRoleNames(), RoleNames(), checkCredentialChars(), foldToEnvVar(), Config, invertedCPUSetRange(), missingErr() (+14 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.16
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "importops_test.go"
Cohesion: 0.10
Nodes (31): importIgnore(), planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt(), Ops (+23 more)

### Community 34 - "domaincerts_resolve_test.go"
Cohesion: 0.07
Nodes (52): DirReader, fakeDirEntry, os.DirEntry, os.FileInfo, os.FileMode, caNameSafe(), checkNoDotDot(), DefaultDirReader() (+44 more)

### Community 35 - "Role"
Cohesion: 0.08
Nodes (21): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+13 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "haCfg"
Cohesion: 0.11
Nodes (20): haCfg(), TestOperatorImage(), TestCreateSecretsFailsWithoutAdminFields(), TestDeleteNamespaceProtected(), TestClusterHonoursRuntime(), TestExecutorRefusesUnapprovedRuntime(), TestRuntimeDefaultArgvUnchanged(), TestTransportHonoursRuntime() (+12 more)

### Community 38 - "Echo"
Cohesion: 0.16
Nodes (10): interactiveFailRunner, TestExecIsSilentWithoutVerbose(), TestResolveMissingBinaryIsActionable(), Echo, EnvRunner, Runner, NewExec(), Quote() (+2 more)

### Community 39 - ".ConfigureReplication"
Cohesion: 0.12
Nodes (29): sortedKeys(), Ops, isRunCLIRejection(), mateConvergenceShutdowns(), missingListedVPNs(), newlineIf(), PlannedRoles(), reenableLine() (+21 more)

### Community 40 - "broker/matechannel_test.go"
Cohesion: 0.19
Nodes (19): CLIRunner, fakeRun, Ops, NewCLIMate(), mateReplies(), newTestMate(), TestLocalMateRunsThroughOps(), TestMateChannelDescribeNamesTheTarget() (+11 more)

### Community 41 - "cli/platform_test.go"
Cohesion: 0.29
Nodes (16): runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection(), TestPlatformFlagRejectsUnknownValue() (+8 more)

### Community 42 - "InjectShutdown"
Cohesion: 0.17
Nodes (27): lineRole, serviceLine, shutdownStyle, span, svcKey, classifyServiceRest(), containsPortCommand(), describeKey() (+19 more)

### Community 43 - "exit_test.go"
Cohesion: 0.20
Nodes (14): runFailRunner, childExit(), childStatusError(), k8sEnv(), TestChildExitFallsBackWhenThereIsNoStatus(), TestChildExitKeepsItsMessage(), TestChildExitStatusIsScopedToInteractiveSessions(), TestEveryArgValidatorIsAUsageError() (+6 more)

### Community 44 - "commanddoc_test.go"
Cohesion: 0.16
Nodes (23): shorthand, github.com/spf13/pflag.FlagSet, mdRow(), renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeAbbrevTable() (+15 more)

### Community 45 - "testing.T"
Cohesion: 0.06
Nodes (66): bytes.Buffer, testing.T, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors() (+58 more)

### Community 46 - "podName"
Cohesion: 0.12
Nodes (16): BrokerPodSuffixShape(), allCustomMounted(), HARoles(), lbServiceName(), podName(), ProductKeyRoles(), pvcName(), RestartOrder() (+8 more)

### Community 47 - "ReplSite"
Cohesion: 0.20
Nodes (25): Ops, ReplSite, execArgs(), NewMateChannel(), ReadSecretKey(), secretCommand(), hasPrefixArgv(), indexOf() (+17 more)

### Community 48 - "confirmAction"
Cohesion: 0.16
Nodes (28): exportconfigReadCounter, layer, io.Reader, io.Writer, TestConfirmDeleteShortcut(), TestConfirmNonTTY(), TestPromptAsksWhenStdinIsNotATTY(), TestPromptKeepsBufferedAnswersForLaterQuestions() (+20 more)

### Community 49 - "Ops"
Cohesion: 0.11
Nodes (16): cliRunNames(), countContains(), runCLISkeleton(), TestCountContains(), TestValidName(), validCLILine(), validName(), ValidVPNName() (+8 more)

### Community 50 - "Quadlet"
Cohesion: 0.16
Nodes (17): ContainerNoFile(), HealthCheck, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), EnvPairs(), escapePercent(), groupKey(), healthCmd() (+9 more)

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (31): TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing(), TestStartStopHaveNoAlias(), TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRegisteredWhereItExecutes() (+23 more)

### Community 52 - "execguard_test.go"
Cohesion: 0.18
Nodes (13): Config, guardCommandOf(), setGuardCommand(), TestAllowCommandsAccepts(), TestAllowCommandsRejects(), TestAllowCommandsRejectsEscalation(), TestAllowedBinaryIsNotGloballyAllowed(), TestComposeCommandDerivation() (+5 more)

### Community 53 - "opRunner"
Cohesion: 0.14
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

### Community 58 - ".rpc"
Cohesion: 0.18
Nodes (9): sempExecuteResult, sempMate, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, encoding/xml.Name, Ops, sempReplRole() (+1 more)

### Community 59 - ".BackupRevertActivity"
Cohesion: 0.19
Nodes (15): backupTarget, Credential, TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigFlag(), curlConfigLine(), Ops, httpBody() (+7 more)

### Community 60 - "cli_test.go"
Cohesion: 0.10
Nodes (51): testing.M, TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), allowRuntime(), echoRunner(), firstLine(), runCtr(), TestAnnounceCommandsNamesResolvedBinaries() (+43 more)

### Community 61 - "checkreport.go"
Cohesion: 0.15
Nodes (20): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), containsString(), failRow(), Cluster (+12 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.18
Nodes (35): failOnCall(), fakeEnv(), Manager, healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirRefusesAnUnwritableParent(), TestCheckIDMappingRefusesAnUnmappedRunUser() (+27 more)

### Community 65 - "Configuration"
Cohesion: 0.15
Nodes (13): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+5 more)

### Community 66 - "scaling_test.go"
Cohesion: 0.14
Nodes (18): decodeScaling(), Config, setContainerCPUSet(), setContainerMem(), TestApplyScalingTierDefaultsContainerBlocks(), TestApplyScalingTierDefaultsCPUSetPerTier(), TestApplyScalingTierDefaultsK8s(), TestApplyScalingTierDefaultsMemOverride() (+10 more)

### Community 67 - "broker/replication.go"
Cohesion: 0.09
Nodes (54): colSpan, MateConfig, cliTransport(), containsEndpoint(), dashSpans(), flagByte(), gutterClear(), labelValue() (+46 more)

### Community 68 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "blocks.go"
Cohesion: 0.07
Nodes (46): Omission, Region, token, Annotate(), markerRegion(), TestAnnotateIsDeterministic(), TestAnnotateRoundTripCaptureIsMarked(), TestAnnotateRoundTripPreservesBlocks() (+38 more)

### Community 71 - "localCfg"
Cohesion: 0.09
Nodes (45): bridgeHostPort(), sempPort(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA(), TestBackupRevertActivityCredsAndBodyNeverInArgv(), TestBackupRevertActivityNon2xx(), TestBackupRevertActivityRPCNotOK() (+37 more)

### Community 72 - "asUsage"
Cohesion: 0.10
Nodes (21): childExitError, usageError, github.com/spf13/cobra.PositionalArgs, logArgs(), App, newConvertCmd(), runConvert(), App (+13 more)

### Community 73 - "internal/k8s"
Cohesion: 0.11
Nodes (19): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, matechannel_test.go, names_test.go (+11 more)

### Community 74 - "annotate_test.go"
Cohesion: 0.09
Nodes (40): applyMetaField(), Capture, isMarker(), markerSection(), markerVerb(), parseKV(), readMarker(), renderKV() (+32 more)

### Community 75 - "Set"
Cohesion: 0.15
Nodes (13): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+5 more)

### Community 76 - "namespace_test.go"
Cohesion: 0.30
Nodes (11): Cluster, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 77 - "newTestOps"
Cohesion: 0.08
Nodes (47): Ops, newTestOps(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs(), TestDisableDefaultUsersRefusesAQuotedVPNName(), TestDisableDefaultVPN(), TestDomainCerts(), TestDomainCertsAcceptsAFullHostPath() (+39 more)

### Community 78 - "Platform"
Cohesion: 0.20
Nodes (23): noRolePositional(), usagef(), checkFlagPlatforms(), commandPlatforms(), declaredList(), App, parsePlatformList(), platformSuffix() (+15 more)

### Community 79 - "runRootWith"
Cohesion: 0.16
Nodes (27): App, runRootWith(), TestK8sConfigDeleteDomainCertsConfigured(), TestK8sConfigDeleteDomainCertsFromDirs(), TestRemoveServerCertsRefusedOnASecretManagedDeployment(), ExitCode(), exportconfigDriverOutput(), exportconfigTransportOutput() (+19 more)

### Community 80 - "github.com/spf13/cobra.Command"
Cohesion: 0.21
Nodes (23): github.com/spf13/cobra.Command, applyAliases(), newBrokerConfigureCmd(), newConfigureDataReplicationCmd(), newConfigureDefaultUsersCmd(), newConfigureDefaultVPNCmd(), newConfigureDomainCertsCmd(), newConfigureProductKeysCmd() (+15 more)

### Community 81 - "k8s/inspect_test.go"
Cohesion: 0.16
Nodes (18): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestConditionLevelDegradesSafely() (+10 more)

### Community 82 - "age"
Cohesion: 0.22
Nodes (9): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+1 more)

### Community 83 - "limitsMgr"
Cohesion: 0.21
Nodes (17): Manager, limitsMgr(), nrOpenProbe(), TestCheckLimitsPrivilegedSkipsTheUserManager(), TestCheckLimitsRootlessEmptyUserManagerAnswerSkips(), TestCheckLimitsRootlessRefusesAShortUserManager(), TestCheckLimitsRootlessReportsTheSoftLimitWithoutGating(), TestCheckLimitsSkipsOnAnEUIDMismatch() (+9 more)

### Community 84 - "Config"
Cohesion: 0.18
Nodes (5): Image, TLS, atoiPrefix(), Config, Scaling

### Community 86 - "k8s/inspect.go"
Cohesion: 0.16
Nodes (20): time.Time, anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes() (+12 more)

### Community 87 - "prep_test.go"
Cohesion: 0.17
Nodes (18): GenBroker(), joinManifests(), namespaceManifest(), adminCfg(), TestCreateNamespaceApplyFails(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree(), TestCreateSecretsPreflight() (+10 more)

### Community 88 - "commands.go"
Cohesion: 0.29
Nodes (25): addCommands(), addLogFlags(), App, group(), newBrokerCLICmd(), newBrokerCmd(), newBrokerRemoveCmd(), newBrokerShellCmd() (+17 more)

### Community 89 - "load"
Cohesion: 0.12
Nodes (40): BrokerCR(), Compose(), composeEscape(), ComposeProject(), ContainerSecrets(), SecretPreflight(), assertNoCheckoutPath(), envLines() (+32 more)

### Community 90 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 91 - "RuleFor"
Cohesion: 0.10
Nodes (38): Disposition, SectionRule, Capture, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+30 more)

### Community 92 - "opPerformReplication"
Cohesion: 0.17
Nodes (22): mateChannelFunc, bufio.Reader, ReplicationConfigResult, App, lineSink(), progress(), step(), warn() (+14 more)

### Community 93 - "load.go"
Cohesion: 0.24
Nodes (12): TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config, podmanRunUser() (+4 more)

### Community 94 - "hostpath.go"
Cohesion: 0.18
Nodes (10): expandTilde(), expandTildeToken(), Config, HasPathSeparator(), IsAbsHostPath(), isPathSep(), TestExpandTilde(), TestHasPathSeparator() (+2 more)

### Community 95 - "scaling.go"
Cohesion: 0.13
Nodes (16): scalingKey, scalingSpelling, scalingTier, containerMem(), cpuSetRange(), Config, Scaling, Scaling (+8 more)

### Community 96 - "matchCLI"
Cohesion: 0.05
Nodes (54): matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultVPNDisableError() (+46 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "platformOps"
Cohesion: 0.25
Nodes (25): opFunc, roleOpFunc, addPodFlag(), newBrokerCopyCmd(), newBrokerDeployCmd(), newBrokerExecCmd(), newBrokerGenerateCmd(), newBrokerLogsCmd() (+17 more)

### Community 100 - ".checkNrOpen"
Cohesion: 0.30
Nodes (5): Manager, limitText(), parseLimit(), parseUnitProps(), TestParseLimit()

### Community 101 - "scripts.go"
Cohesion: 0.08
Nodes (40): showCmd, assertLeaderScript(), currentConfigScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript(), enableDefaultUsersScript() (+32 more)

### Community 103 - "container/secrets_test.go"
Cohesion: 0.20
Nodes (17): ResolveSecretValues(), certFixture(), TestDeletePodmanRemovesTheBundleAfterTheUnit(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeWritingTheComposeFile(), TestDeployDockerPassesTheBundleAsEnvNotArgv(), TestDeployPodmanFailsBeforeWritingAnythingOnABadCert(), TestDeployPodmanFailsWhenTheBundleCannotBeWritten() (+9 more)

### Community 104 - "watch_test.go"
Cohesion: 0.11
Nodes (25): imageFromDeployment(), desiredWatch(), Cluster, splitWatch(), subtractWatch(), deployJSON(), Cluster, operatorItem() (+17 more)

### Community 105 - "internal/cli"
Cohesion: 0.17
Nodes (12): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, examples_test.go, exit_test.go (+4 more)

### Community 106 - "operatorversion_test.go"
Cohesion: 0.16
Nodes (15): compareVersions(), Cluster, imageTag(), operatorVersionWarning(), parseVersion(), operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade() (+7 more)

### Community 107 - "replicationops_test.go"
Cohesion: 0.27
Nodes (17): hasCall(), TestDisableDefaultUsersShowVPNError(), TestReleaseToBackupReleasedTimeout(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), TestConfigureReplicationClosingShowFailureStillReportsTheApply() (+9 more)

### Community 108 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 109 - "omit_test.go"
Cohesion: 0.40
Nodes (4): TestImportPlanAcceptsWhatExportProduced(), TestImportPlanRefusesAFileItDidNotExport(), TestOmitAnnotateRoundTripsOmissions(), TestOmitExportConfigEndToEnd()

### Community 110 - "internal/broker"
Cohesion: 0.09
Nodes (23): annotate_test.go, appliance_test.go, blocks_test.go, broker_test.go, coverage_test.go, diff_test.go, driver_test.go, importdoc_test.go (+15 more)

### Community 111 - "statusreport_test.go"
Cohesion: 0.20
Nodes (17): Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure(), TestBrokerReportUsesTheCRsOwnPodList(), TestBrokerReportWithNoBrokerDeployed() (+9 more)

### Community 112 - "redundancy_test.go"
Cohesion: 0.21
Nodes (12): Config, pskCfg(), TestFillStandaloneNodeNameIgnoresAnEmptyHostname(), TestFillStandaloneNodeNameKeepsAConfiguredName(), TestFillStandaloneNodeNameNeverFillsHA(), TestFillStandaloneNodeNameUsesTheHost(), TestNothingGeneratesThePSK(), TestPSKEnvSatisfiesTheContainerRequirement() (+4 more)

### Community 113 - "internal/config"
Cohesion: 0.13
Nodes (15): command_test.go, config_test.go, domaincerts_resolve_test.go, domaincerts_test.go, duration_test.go, execguard_test.go, hostpath_test.go, imagepullsecret_test.go (+7 more)

### Community 114 - "BrokerType"
Cohesion: 0.28
Nodes (5): BrokerType, cliMate, bannerType(), showReplicationScript(), showVPNReplicationScript()

### Community 115 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 116 - "manager.go"
Cohesion: 0.18
Nodes (10): orUnknown(), composeNeedsSecretValues(), exactName(), orNone(), orValue(), platformTitle(), secretSummary(), setOrMissing() (+2 more)

### Community 117 - "internal/container"
Cohesion: 0.22
Nodes (9): inspect_test.go, internal/container, limits_test.go, manager_test.go, preflight_test.go, rootless_test.go, runtime_test.go, secrets_test.go (+1 more)

### Community 118 - ".stateRows"
Cohesion: 0.36
Nodes (4): containerState, healthState, Manager, printable()

### Community 119 - "hostpath_test.go"
Cohesion: 0.10
Nodes (24): CanonicalDuration(), TestCanonicalDuration(), guardConfig(), CheckHostPath(), Config, setContainerName(), TestCheckHostPathAccepts(), TestCheckHostPathRejects() (+16 more)

### Community 120 - "completion.go"
Cohesion: 0.31
Nodes (10): github.com/spf13/cobra.ShellCompDirective, completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), completionShell(), App, isEnvFileName() (+2 more)

### Community 121 - "Ops"
Cohesion: 0.10
Nodes (20): time.Duration, Ops, TestImportOpsReportsDoNotPanicOnEmptyAndPopulated(), TestImportPlanReportPrintsWarningsBeforeTheConfirmation(), TestImportResultReportsTeardownAndRebuildSeparately(), confirmImport(), App, pluralVPN() (+12 more)

### Community 122 - "VPNRepl"
Cohesion: 0.14
Nodes (7): AdminState, QueueState, ReplRole, scriptedMate, VPNRepl, setReplicationRoleScript(), sempAdminState()

### Community 123 - "container/inspect_test.go"
Cohesion: 0.17
Nodes (17): decodeInspect(), inspectReply(), kvValue(), TestInspectDecodesPartialOutput(), TestInspectDistinguishesNoHealthcheckFromUnknown(), TestInspectHealthReadsBothSpellings(), TestInspectNameLosesDockersLeadingSlash(), TestInspectRejectsAnEmptyReply() (+9 more)

### Community 124 - "cli/replication_test.go"
Cohesion: 0.28
Nodes (16): mateChannel(), mateSEMPPassword(), App, replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer(), TestMateChannelPicksTheMechanismTheSiteDeclares() (+8 more)

### Community 125 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 126 - "storage_test.go"
Cohesion: 0.33
Nodes (9): Config, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 127 - "Manager"
Cohesion: 0.22
Nodes (4): Manager, idMapCovers(), origin(), runUserIDs()

### Community 128 - "TestServerCert"
Cohesion: 0.23
Nodes (10): TestDiagnostics(), TestPathHelpers(), TestServerCert(), serverCertFile(), serverCertScript(), TestServerCertScript(), certPath(), CLIArg() (+2 more)

### Community 129 - "config/replication_test.go"
Cohesion: 0.27
Nodes (12): Replication, Config, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded(), TestValidateReplicationAbsent() (+4 more)

### Community 131 - ".preflightOne"
Cohesion: 0.44
Nodes (3): canIAnswer(), Cluster, probe

### Community 132 - "Troubleshooting"
Cohesion: 0.25
Nodes (8): A replication switch refuses before changing anything, A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, Import reported failure, Podman secret flags, The limits the container actually gets, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 134 - "verify_local.go"
Cohesion: 0.50
Nodes (4): defaultLocalAddrs(), hostMatches(), shortHost(), TestDefaultLocalAddrs()

### Community 135 - "helpers.go"
Cohesion: 0.40
Nodes (3): emit(), renderOnly(), newVersionCmd()

### Community 138 - "pullSecretCfg"
Cohesion: 0.39
Nodes (8): Config, pullSecretCfg(), TestImagePullSecretDerivedSuffixFitsWhereAdditionalUsersAlreadyFits(), TestImagePullSecretNameBringYourOwn(), TestImagePullSecretNameConfiguredNameWins(), TestImagePullSecretNameDerivedWhenCredentialsButNoName(), TestImagePullSecretNameEmptyWithNeither(), TestManagesImagePullSecretNeedsBothCredentials()

### Community 143 - "Data replication"
Cohesion: 0.40
Nodes (5): Configuring a site, Data replication, Reaching the mate, Switching roles, When a run fails partway

## Knowledge Gaps
- **222 isolated node(s):** `solace`, `span`, `Ops`, `showCmd`, `Ops` (+217 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 331 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **11 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `Get`, `containerTransport`, `config_test.go`, `Manager`, `config.go`, `Command`, `capRunner`, `convert_test.go`, `ctrCfg`, `Load`, `validate.go`, `execguard_test.go`, `opRunner`, `scaling_test.go`, `localCfg`, `github.com/spf13/cobra.Command`, `limitsMgr`, `load`, `opPerformReplication`, `load.go`, `scaling.go`, `platformOps`, `container/secrets_test.go`, `redundancy_test.go`, `manager.go`, `hostpath_test.go`, `Ops`?**
  _High betweenness centrality (0.046) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `config/replication_test.go`, `containerTransport`, `Manager`, `config.go`, `GenOperator`, `capRunner`, `NewCluster`, `convert_test.go`, `ctrCfg`, `k8s/secrets_test.go`, `render.go`, `haCfg`, `podName`, `ReplSite`, `Quadlet`, `opRunner`, `cli_test.go`, `checkreport.go`, `Cluster`, `localCfg`, `namespace_test.go`, `newTestOps`, `prep_test.go`, `load`, `opPerformReplication`, `kubectlTransport`, `container/secrets_test.go`, `watch_test.go`, `Ops`?**
  _High betweenness centrality (0.043) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `containerTransport`, `Manager`, `config.go`, `.releaseToBackup`, `renderDriver`, `ops_container.go`, `importops.go`, `validate.go`, `.ConfigureReplication`, `broker/matechannel_test.go`, `podName`, `ReplSite`, `Ops`, `.BackupRevertActivity`, `opPerformReplication`, `scripts.go`, `kubectlTransport`, `BrokerType`, `Ops`?**
  _High betweenness centrality (0.039) - this node is a cross-community bridge._
- **Are the 81 inferred relationships involving `newTestOps()` (e.g. with `TestDiagnosticsMkdirError()` and `TestDiagnosticsTwoRolesNoBundle()`) actually correct?**
  _`newTestOps()` has 81 INFERRED edges - model-reasoned connections that need verification._
- **Are the 41 inferred relationships involving `ctrCfg()` (e.g. with `TestInspectStateSurfacesTheEngineError()` and `TestReportStateNeverPrintsTheEnvironment()`) actually correct?**
  _`ctrCfg()` has 41 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `span`, `Ops` to the rest of the system?**
  _222 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `context.Context` be split into smaller, more focused modules?**
  _Cohesion score 0.0786308973172988 - nodes in this community are weakly interconnected._
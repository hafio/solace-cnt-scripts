# Graph Report - solace-cnt-scripts  (2026-10-22)

## Corpus Check
- 204 files · ~692,080 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 23 file(s) not represented in the graph (top: .golden 17, (none) 2, .cli 2)

## Summary
- 4020 nodes · 20448 edges · 98 communities (92 shown, 6 thin omitted)
- Extraction: 72% EXTRACTED · 28% INFERRED · 0% AMBIGUOUS · INFERRED: 5648 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `e5242b84`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- App
- newTestOps
- .ExportConfig
- context.Context
- cli_test.go
- testing.T
- Command Reference Docs
- Manager
- parsePort
- runPlatform
- CheckCommand
- opArgvMatch
- NewCluster
- Capture
- convert_test.go
- transform_test.go
- manager_test.go
- commands.go
- coverage_test.go
- Bash Dev Script
- .runMatePhase
- PowerShell Dev Script
- Set
- switchplan_test.go
- newLocalOps
- runner_test.go
- renderDriver
- completion_test.go
- importops_test.go
- deploymentItem
- Solace Module Root
- Config
- CLAUDE.md
- tlssecret_test.go
- ResolveDomainCerts
- watchLoop
- github.com/spf13/cobra.ShellCompDirective
- recRunner
- annotate_test.go
- Sink
- Quadlet
- inject_test.go
- Coverage
- exportconfig_test.go
- VPNRepl
- localCfg
- config_test.go
- prep_test.go
- newRootCmd
- runRootWith
- .validate
- .kubectl
- replication_test.go
- .rpc
- confirm.go
- .configRows
- rootless_test.go
- exit_test.go
- opPerformReplication
- Developer guide
- namespace_test.go
- usagef
- Compose
- Command
- replicationops_test.go
- k8s/matechannel_test.go
- limits_test.go
- k8s/inspect.go
- .ConfigureReplication
- AdminSecret
- render_test.go
- RuleFor
- Container
- replConfig
- reportCluster
- DetectPlatforms
- Platforms
- operatorversion_test.go
- Config
- config_ops_test.go
- .OperatorApply
- capRunner
- storage_test.go
- certFixture
- Container State Inspection
- .RunCLI
- .validateContainerArtifactValues
- bytes.Buffer
- hostpath_test.go
- Role
- ReplVia
- containerTransport
- blocks.go
- strings.Builder

## God Nodes (most connected - your core abstractions)
1. `App` - 217 edges
2. `newTestOps()` - 167 edges
3. `cli_test.go` - 148 edges
4. `Config` - 145 edges
5. `ctrCfg()` - 133 edges
6. `Role` - 132 edges
7. `NewCluster()` - 128 edges
8. `Per-package doubles` - 124 edges
9. `newCapMgr()` - 114 edges
10. `loadK8s()` - 111 edges

## Surprising Connections (you probably didn't know these)
- `Repository layout` --references--> `Execute()`  [INFERRED]
  docs/developer.md → internal/cli/root.go
- `The preflight` --references--> `info()`  [INFERRED]
  docs/operations.md → internal/k8s/checkreport.go
- `Per-package doubles` --references--> `sample()`  [INFERRED]
  docs/test.md → internal/abbrev/abbrev_test.go
- `omit_test.go` --references--> `Annotate()`  [INFERRED]
  docs/test.md → internal/broker/annotate.go
- `annotate_test.go` --references--> `TestMarkerRegionBeginAndEnd()`  [INFERRED]
  docs/test.md → internal/broker/annotate_test.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (98 total, 6 thin omitted)

### Community 0 - "App"
Cohesion: 0.06
Nodes (100): App, ValidScriptName(), logArgs(), confirmAction(), confirmDelete(), lineSink(), runExport(), runImport() (+92 more)

### Community 1 - "newTestOps"
Cohesion: 0.08
Nodes (66): broker_test.go, countContains(), leaderCalls(), newTestOps(), TestCountContains(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs() (+58 more)

### Community 2 - ".ExportConfig"
Cohesion: 0.31
Nodes (8): Omission, omit_test.go, keepRegion(), omitAtExport(), TestImportPlanRefusesAFileItDidNotExport(), TestOmitAtExportDropsFilteredLinesAndEmptyBlocks(), TestOmitAtExportLeavesTargetDecisionsToImport(), TestOmitAtExportRemovesSkipSectionsAndRecordsThem()

### Community 3 - "context.Context"
Cohesion: 0.06
Nodes (10): cliMate, scriptedMate, showRedundancyLocalScript(), showReplicationScript(), EnvRunner, Runner, shSingleQuote(), badJSONRunner (+2 more)

### Community 4 - "cli_test.go"
Cohesion: 0.06
Nodes (99): cli_test.go, examples_test.go, allowRuntime(), captureStderr(), collectPaths(), fakeBinaryOnPath(), firstLine(), runCtr() (+91 more)

### Community 5 - "testing.T"
Cohesion: 0.04
Nodes (146): scalingTier, artifactvalues_test.go, config_test.go, redundancy_test.go, scaling_test.go, TestOperatorAdminSecretName(), TestPSKConfiguredCountsBothSpellings(), TestValidateContainerArtifactValues() (+138 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (50): Commands, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker cli, solace-util broker configure (+42 more)

### Community 7 - "Manager"
Cohesion: 0.08
Nodes (19): preflight_test.go, certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), TestManagerCopy() (+11 more)

### Community 8 - "parsePort"
Cohesion: 0.40
Nodes (5): cut(), parsePort(), splitUser(), TestParsePort(), portSpec

### Community 9 - "runPlatform"
Cohesion: 0.28
Nodes (17): platform_test.go, runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection() (+9 more)

### Community 10 - "CheckCommand"
Cohesion: 0.13
Nodes (28): commandRules, execguard_test.go, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules() (+20 more)

### Community 11 - "opArgvMatch"
Cohesion: 0.09
Nodes (61): leaderRun, opCall, replRig, ops_k8s_test.go, Per-package doubles, failDisableDefaultUsersUpload(), isDeletePod(), isGetPod() (+53 more)

### Community 12 - "NewCluster"
Cohesion: 0.06
Nodes (106): checkreport_test.go, deploy_test.go, operator_test.go, TestCheckStopsProbingWhenUnreachable(), TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestOperatorInstallVerdict(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestStorageRows() (+98 more)

### Community 13 - "Capture"
Cohesion: 0.21
Nodes (11): BrokerType, Capture, transcript_test.go, bannerType(), checkProvenance(), checkSameType(), stripTranscript(), TestTranscriptBannerType() (+3 more)

### Community 14 - "convert_test.go"
Cohesion: 0.07
Nodes (69): doc, Result, segment, vars, The keys, convert_test.go, DecodeStrict(), boolOf() (+61 more)

### Community 15 - "transform_test.go"
Cohesion: 0.10
Nodes (61): Block, TargetState, transform_test.go, BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects(), ClearTargetVirtualHostnames() (+53 more)

### Community 16 - "manager_test.go"
Cohesion: 0.09
Nodes (97): manager_test.go, GuardPodmanEUID(), fileExists(), NewManager(), assertMode(), containsStr(), ctrCfg(), hasCall() (+89 more)

### Community 17 - "commands.go"
Cohesion: 0.14
Nodes (71): opFunc, roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd() (+63 more)

### Community 18 - "coverage_test.go"
Cohesion: 0.06
Nodes (69): coverage_test.go, internal/broker, verify_ops_test.go, matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError() (+61 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - ".runMatePhase"
Cohesion: 0.36
Nodes (6): isRunCLIRejection(), newlineIf(), replPhase1Rejected(), replPhase2Rejected(), replTransportFailure(), replVPNList()

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Set"
Cohesion: 0.18
Nodes (15): Entry, abbrev_test.go, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed() (+7 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.06
Nodes (78): CLIRunner, fakeRun, PhaseKind, SiteState, SwitchAction, SwitchPhase, matechannel_test.go, switchplan_test.go (+70 more)

### Community 24 - "newLocalOps"
Cohesion: 0.16
Nodes (33): verify_local_test.go, cliScriptPath(), isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn(), TestDetectRoleAddrsError() (+25 more)

### Community 25 - "runner_test.go"
Cohesion: 0.06
Nodes (54): interactiveFailRunner, Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication (+46 more)

### Community 26 - "renderDriver"
Cohesion: 0.15
Nodes (22): chunk, chunkResult, driver_test.go, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce() (+14 more)

### Community 27 - "completion_test.go"
Cohesion: 0.22
Nodes (24): completion_test.go, driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion() (+16 more)

### Community 28 - "importops_test.go"
Cohesion: 0.06
Nodes (75): ImportResult, PlannedSection, importops_test.go, defaultVPNFirst(), describeScope(), ImportPlan, importIgnore(), omittedList() (+67 more)

### Community 29 - "deploymentItem"
Cohesion: 0.16
Nodes (11): imageFromDeployment(), operatorRunningImage(), TestOperatorRunningImageWithNoContainers(), splitWatch(), TestWatchFromContainersReadsEveryAllNamespacesSpelling(), watchFromContainers(), watchFromDeployment(), containerSpec (+3 more)

### Community 31 - "Config"
Cohesion: 0.07
Nodes (15): keyValueEntries, cpuSetCount(), checkCredentialChars(), foldToEnvVar(), invertedCPUSetRange(), missingErr(), platformKey(), requireAll() (+7 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.05
Nodes (38): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, Abbreviations, Flag shorthands, How a short form is resolved (+30 more)

### Community 33 - "tlssecret_test.go"
Cohesion: 0.43
Nodes (8): tlssecret_test.go, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsDeriveASecretName(), TestSuppliedCertsMakeTheSecretOurs(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), TestTLSServerSecretNameFollowsTheStates(), tlsCfg()

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (52): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domaincerts_resolve_test.go, domaincerts_test.go, domainCANames() (+44 more)

### Community 35 - "watchLoop"
Cohesion: 0.32
Nodes (5): watchInterval(), stdoutIsTerminal(), watchLabel(), watchLoop(), withWatch()

### Community 36 - "github.com/spf13/cobra.ShellCompDirective"
Cohesion: 0.33
Nodes (6): completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), isEnvFileName(), matching()

### Community 37 - "recRunner"
Cohesion: 0.16
Nodes (11): preflight_test.go, TestCanIAnswerReadsTheLastLine(), TestCanIArgsScopesByResourceKind(), TestPreflightAcceptsAWarnedYes(), TestPreflightIsPreviewableUnderEcho(), TestPreflightNamesTheObjectOfANamedProbe(), TestPreflightRefusesAnUnreadableAnswer(), canIVerb() (+3 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.11
Nodes (46): annotate_test.go, Annotate(), applyMetaField(), isMarker(), markerSection(), markerVerb(), parseKV(), readMarker() (+38 more)

### Community 40 - "Sink"
Cohesion: 0.06
Nodes (66): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, diff_test.go, internal/abbrev, internal/convert (+58 more)

### Community 41 - "Quadlet"
Cohesion: 0.10
Nodes (26): opCtrGenArtifact(), ContainerNoFile(), HealthCheck, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), containerSecretSpecs(), EnvPairs(), escapePercent() (+18 more)

### Community 42 - "inject_test.go"
Cohesion: 0.17
Nodes (28): lineRole, serviceLine, shutdownStyle, svcKey, inject_test.go, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "Coverage"
Cohesion: 0.33
Nodes (5): Coverage, defaultLocalAddrs(), TestDefaultLocalAddrs(), TestQuadletKeysAreWithinTheFloor(), versionAtMost()

### Community 44 - "exportconfig_test.go"
Cohesion: 0.14
Nodes (27): exportconfigReadCounter, opRunner, exportconfig_test.go, internal/cli, confirmImport(), pluralVPN(), exportconfigDriverOutput(), exportconfigTransportOutput() (+19 more)

### Community 47 - "VPNRepl"
Cohesion: 0.12
Nodes (11): AdminState, QueueState, replMate, ReplRole, VPNRepl, ReplicationConfigResult, missingListedVPNs(), reenableLine() (+3 more)

### Community 48 - "localCfg"
Cohesion: 0.08
Nodes (53): semp_test.go, sempmate_test.go, bridgeHostPort(), sempPort(), sempV1OK(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA() (+45 more)

### Community 49 - "config_test.go"
Cohesion: 0.16
Nodes (30): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, showCmd, span, AdditionalUser, Node (+22 more)

### Community 50 - "prep_test.go"
Cohesion: 0.06
Nodes (94): adminsecret_test.go, check_test.go, cluster_test.go, internal/k8s, names_test.go, ops_test.go, prep_test.go, transport_test.go (+86 more)

### Community 51 - "newRootCmd"
Cohesion: 0.12
Nodes (28): aliases_test.go, allowcommand_test.go, applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing() (+20 more)

### Community 52 - "runRootWith"
Cohesion: 0.14
Nodes (37): euid_test.go, watch_test.go, echoRunner(), runRootWith(), TestAssertLeaderOnKubernetesTouchesOnlyThePrimaryPod(), TestConfiguredRouternameSurvivesTheFallback(), TestCtrConfirmDeclined(), TestCtrRestartConfirmGate() (+29 more)

### Community 54 - ".validate"
Cohesion: 0.06
Nodes (37): solace-util, solace-util operator, A replication switch refuses before changing anything, A wrapper runtime is refused, Bringing up a fresh cluster, Broker scope is a fixed classification, Docker and Podman mechanics, Docker compose secrets need compose 2.23.1+ (+29 more)

### Community 56 - ".kubectl"
Cohesion: 0.06
Nodes (18): runtime_test.go, BrokerPodSuffixShape(), Cluster, HARoles(), lbServiceName(), podName(), pvcName(), stsName() (+10 more)

### Community 57 - "replication_test.go"
Cohesion: 0.08
Nodes (53): Commands, solace-util broker, Configuring a site, Data replication, Reaching the mate, Switching roles, When a run fails partway, replication_test.go (+45 more)

### Community 59 - ".rpc"
Cohesion: 0.13
Nodes (13): backupTarget, Credential, sempMate, schemaTransport(), anyHTTP2xx(), curlConfigLine(), httpBody(), lastHTTPStatus() (+5 more)

### Community 60 - "confirm.go"
Cohesion: 0.18
Nodes (19): layer, Injectable seams, TestStdinCanAnswerClosedFile(), completionShell(), newCompletionCmd(), confirmActionStrict(), confirmDowngrade(), confirmGate() (+11 more)

### Community 61 - ".configRows"
Cohesion: 0.11
Nodes (28): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminSecretRow(), containsString(), failRow() (+20 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.19
Nodes (40): rootless_test.go, TestManagerPrepHostLeavesBaseDirAlone(), failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir() (+32 more)

### Community 63 - "exit_test.go"
Cohesion: 0.14
Nodes (24): runFailRunner, exit_test.go, Fixtures and doubles, Shared env fixtures, TestExecute(), childExit(), ExitCode(), childStatusError() (+16 more)

### Community 66 - "opPerformReplication"
Cohesion: 0.16
Nodes (10): mateChannelFunc, progress(), step(), emitOrWrite(), opK8sGenBroker(), opK8sGenOperator(), confirmReplicationSwitch(), opCtrPerformReplication() (+2 more)

### Community 69 - "Developer guide"
Cohesion: 0.12
Nodes (17): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+9 more)

### Community 73 - "namespace_test.go"
Cohesion: 0.31
Nodes (11): namespace_test.go, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+3 more)

### Community 75 - "usagef"
Cohesion: 0.08
Nodes (35): childExitError, usageError, ValidVPNName(), noRolePositional(), newConvertCmd(), runConvert(), newExamplesCmd(), runExample() (+27 more)

### Community 76 - "Compose"
Cohesion: 0.20
Nodes (14): internal/render, servercert_test.go, Compose(), composeEscape(), ContainerSecrets(), ContainerSecret, SecretPreflight(), TestComposeEscapesTheDollarSign() (+6 more)

### Community 77 - "Command"
Cohesion: 0.10
Nodes (21): cmdField, guardedCmd, command_test.go, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal() (+13 more)

### Community 81 - "replicationops_test.go"
Cohesion: 0.16
Nodes (42): replicationops_test.go, hasCall(), mateConvergenceShutdowns(), PlannedRoles(), replicationVPNLines(), RoleAtSite(), replListed(), replReadOnlyResponder() (+34 more)

### Community 82 - "k8s/matechannel_test.go"
Cohesion: 0.19
Nodes (28): matechannel_test.go, CLIArg(), CLIScriptPath(), execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf() (+20 more)

### Community 83 - "limits_test.go"
Cohesion: 0.07
Nodes (52): limits_test.go, main_test.go, delegateSetting(), limitText(), missingControllers(), parseLimit(), parseUnitProps(), limitsMgr() (+44 more)

### Community 86 - "k8s/inspect.go"
Cohesion: 0.11
Nodes (38): inspect_test.go, orNone(), age(), anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), normalizeToList() (+30 more)

### Community 87 - ".ConfigureReplication"
Cohesion: 0.08
Nodes (64): colSpan, replication_test.go, validVPNName(), cliTransport(), containsEndpoint(), dashSpans(), flagByte(), MateConfig (+56 more)

### Community 88 - "AdminSecret"
Cohesion: 0.10
Nodes (36): The env-file generator page, generatorpage_test.go, secrets_test.go, gunzipBase64(), gzipBase64(), templateActions(), TestGeneratorPageEmbedsTheCLI(), GenSecrets() (+28 more)

### Community 89 - "render_test.go"
Cohesion: 0.15
Nodes (41): render_test.go, boolStr(), BrokerCR(), ComposeProject(), containerArtifacts(), load(), TestAdminCredentialsSecretFollowsTheStates(), TestArtifactsCarryNoSecrets() (+33 more)

### Community 91 - "RuleFor"
Cohesion: 0.11
Nodes (36): Disposition, SectionRule, appliance_test.go, importdoc_test.go, sections_test.go, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent() (+28 more)

### Community 93 - "Container"
Cohesion: 0.16
Nodes (11): Container, DockerConfig, Network, PodmanConfig, Ulimits, applyContainerBlockDefaults(), podmanRunUser(), setDefault() (+3 more)

### Community 97 - "replConfig"
Cohesion: 0.21
Nodes (14): Replication, secretRef, replication_test.go, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded() (+6 more)

### Community 99 - "reportCluster"
Cohesion: 0.13
Nodes (26): statusreport_test.go, TestExecutorRefusesUnapprovedRuntime(), unapprovedCfg(), ownsPod(), reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportOmitsWorkloadsThatAreNotTheBrokers(), TestBrokerReportOrdersPodsOfTheSameRoleByName() (+18 more)

### Community 102 - "DetectPlatforms"
Cohesion: 0.20
Nodes (19): examples_test.go, internal/examples, Example, TestExamplesEmitsToStdout(), parseError(), DetectPlatforms(), All(), Get() (+11 more)

### Community 104 - "Platforms"
Cohesion: 0.17
Nodes (15): platform_test.go, TestExecBinariesCoversEveryPlatform(), TestGuardConfigIsValid(), TestDomainCertsDirsAreNotDefaulted(), newPlatforms(), PlatformAbbrev(), Platforms(), TestDetectPlatforms() (+7 more)

### Community 107 - "operatorversion_test.go"
Cohesion: 0.32
Nodes (12): operatorversion_test.go, operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsNotSkippedByAFailedRead(), TestConfirmNoDowngradeIsSilentInAPreview(), TestConfirmNoDowngradeProceedsWhenConfirmAccepts(), TestConfirmNoDowngradeRefusedWhenConfirmDeclines() (+4 more)

### Community 109 - "Config"
Cohesion: 0.07
Nodes (33): Image, TLS, adminsecret_test.go, imagepullsecret_test.go, adminCfg(), TestAdminSecretNameFollowsTheStates(), TestAdminSecretRefusesTheOperatorsOwnNames(), TestContainersStillRequireTheAdminPassword() (+25 more)

### Community 110 - "config_ops_test.go"
Cohesion: 0.17
Nodes (14): config_ops_test.go, TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(), TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator(), TestExecShellScriptRefusesAnUnsafeNameBeforeUploading() (+6 more)

### Community 111 - ".OperatorApply"
Cohesion: 0.24
Nodes (9): joinYAMLDocs(), nonEmpty(), operatorEscalationProbes(), operatorProbes(), splitAfterNamespace(), splitOperatorBundle(), canIAnswer(), operatorRegcred() (+1 more)

### Community 112 - "capRunner"
Cohesion: 0.12
Nodes (23): capRunner, internal/container, runtime_test.go, transport_test.go, New(), callIndex(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime() (+15 more)

### Community 113 - "storage_test.go"
Cohesion: 0.36
Nodes (9): storage_test.go, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 116 - "certFixture"
Cohesion: 0.15
Nodes (23): capCall, secrets_test.go, secretSummary(), setOrMissing(), ResolveSecretValues(), bundleHash(), certCreate(), certFixture() (+15 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.10
Nodes (25): containerState, healthState, inspect_test.go, decodeInspect(), orUnknown(), printable(), inspectReply(), kvValue() (+17 more)

### Community 121 - ".RunCLI"
Cohesion: 0.05
Nodes (51): scripts_test.go, cliRunNames(), field(), runCLISkeleton(), validCLILine(), validName(), TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), rejectionIn() (+43 more)

### Community 123 - ".validateContainerArtifactValues"
Cohesion: 0.40
Nodes (4): portOrRange(), validContainerPort(), validCoreLimit(), validHealthDuration()

### Community 126 - "bytes.Buffer"
Cohesion: 0.13
Nodes (27): watch_test.go, subtractWatch(), deniedCluster(), deployJSON(), namedOperatorJSON(), operatorItem(), TestAmbiguousOperatorAbortsTheDeploy(), TestFindOperatorDeploymentFallsBackWhenTheClusterWideListFails() (+19 more)

### Community 128 - "hostpath_test.go"
Cohesion: 0.08
Nodes (35): duration_test.go, hostpath_test.go, internal/config, CanonicalDuration(), TestCanonicalDuration(), guardConfig(), BaseName(), CheckHostPath() (+27 more)

### Community 130 - "Role"
Cohesion: 0.07
Nodes (20): covUploadFailsTo, downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile (+12 more)

### Community 132 - "ReplVia"
Cohesion: 0.14
Nodes (12): ReplVia, ReplViaKube, ReplViaSEMP, ReplPassSecret, viaDescription(), CheckSecretKeyRef(), validateReplVia(), validateReplViaSEMP() (+4 more)

### Community 139 - "blocks.go"
Cohesion: 0.10
Nodes (33): Region, token, blocks_test.go, markerRegion(), TestMarkerRegionBeginAndEnd(), brokerTypeFromSchema(), firstQuoted(), isIndented() (+25 more)

### Community 155 - "strings.Builder"
Cohesion: 0.06
Nodes (47): shorthand, ContainerSecurity, Operator, PodSecurity, Resources, Storage, WeightedNodeTerm, mdRow() (+39 more)

## Knowledge Gaps
- **13 isolated node(s):** `solace`, `NO_COLOR`, `bash/000-env.sh`, `bash/010-deploy-operator.sh`, `bash/020-deploy-broker.sh` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 188 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **6 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `App` to `Role`, `context.Context`, `cli_test.go`, `runPlatform`, `opArgvMatch`, `commands.go`, `watchLoop`, `github.com/spf13/cobra.ShellCompDirective`, `Quadlet`, `exportconfig_test.go`, `config_test.go`, `newRootCmd`, `runRootWith`, `replication_test.go`, `confirm.go`, `exit_test.go`, `opPerformReplication`, `usagef`, `Config`?**
  _High betweenness centrality (0.043) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `newTestOps()` (e.g. with `importops_test.go` and `Injectable seams`) actually correct?**
  _`newTestOps()` has 117 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `NO_COLOR`, `bash/000-env.sh` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.06336633663366337 - nodes in this community are weakly interconnected._
- **Why does `Config` connect `Config` to `App`, `hostpath_test.go`, `Role`, `newTestOps`, `cli_test.go`, `testing.T`, `context.Context`, `containerTransport`, `Manager`, `CheckCommand`, `opArgvMatch`, `NewCluster`, `convert_test.go`, `manager_test.go`, `strings.Builder`, `Config`, `tlssecret_test.go`, `ResolveDomainCerts`, `Quadlet`, `Coverage`, `localCfg`, `config_test.go`, `prep_test.go`, `.kubectl`, `.configRows`, `namespace_test.go`, `usagef`, `Compose`, `Command`, `k8s/matechannel_test.go`, `AdminSecret`, `render_test.go`, `Container`, `replConfig`, `reportCluster`, `DetectPlatforms`, `.OperatorApply`, `capRunner`, `storage_test.go`, `certFixture`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.08184143222506395 - nodes in this community are weakly interconnected._
- **Why does `Per-package doubles` connect `opArgvMatch` to `App`, `newTestOps`, `Role`, `context.Context`, `cli_test.go`, `testing.T`, `Manager`, `NewCluster`, `convert_test.go`, `transform_test.go`, `manager_test.go`, `Set`, `switchplan_test.go`, `newLocalOps`, `runner_test.go`, `strings.Builder`, `importops_test.go`, `deploymentItem`, `recRunner`, `exportconfig_test.go`, `VPNRepl`, `localCfg`, `prep_test.go`, `runRootWith`, `.kubectl`, `replication_test.go`, `.rpc`, `exit_test.go`, `opPerformReplication`, `k8s/matechannel_test.go`, `k8s/inspect.go`, `.ConfigureReplication`, `AdminSecret`, `replConfig`, `reportCluster`, `operatorversion_test.go`, `Config`, `.OperatorApply`, `capRunner`, `certFixture`, `.RunCLI`, `bytes.Buffer`?**
  _High betweenness centrality (0.026) - this node is a cross-community bridge._
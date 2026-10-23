# Graph Report - solace-cnt-scripts  (2026-10-23)

## Corpus Check
- 204 files · ~706,042 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .golden 17, (none) 3, .cli 2)

## Summary
- 4070 nodes · 20778 edges · 116 communities (110 shown, 6 thin omitted)
- Extraction: 72% EXTRACTED · 28% INFERRED · 0% AMBIGUOUS · INFERRED: 5795 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `81bfe9f0`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- App
- newTestOps
- ImportPlan
- context.Context
- cli_test.go
- config_test.go
- Command Reference Docs
- Manager
- K8sConfig
- runPlatform
- CheckCommand
- cli_test.go
- NewCluster
- bannerType
- convert_test.go
- transform_test.go
- manager_test.go
- commands.go
- coverage_test.go
- Bash Dev Script
- Load
- PowerShell Dev Script
- Set
- MateChannel
- newLocalOps
- runner_test.go
- renderDriver
- testing.T
- importops_test.go
- deploymentItem
- Solace Module Root
- Config
- CLAUDE.md
- internal/config
- ResolveDomainCerts
- watchLoop
- github.com/spf13/cobra.ShellCompDirective
- NewCluster
- annotate_test.go
- Sink
- Compose
- inject_test.go
- Coverage
- exportconfig_test.go
- VPNRepl
- localCfg
- config_test.go
- eqArgs
- newRootCmd
- scripts_test.go
- prep_test.go
- .validate
- .dump
- .kubectl
- replication_test.go
- deploy_test.go
- .rpc
- confirm.go
- .configRows
- rootless_test.go
- exit_test.go
- BuildSwitchPlan
- switchplan_test.go
- opPerformReplication
- Capture
- Set
- Goldens
- .failed
- TestAdminPasswordPreflight
- TestGeneratorPageEmbedsTheCLI
- namespace_test.go
- redundancy_test.go
- usagef
- podName
- command_test.go
- The eight rules
- Container
- .decodeScalingEntry
- replicationops_test.go
- k8s/matechannel_test.go
- limits_test.go
- Developer guide
- Exporting and importing configuration
- k8s/inspect.go
- .ConfigureReplication
- AdminSecret
- render_test.go
- importIgnore
- RuleFor
- What `import-config` applies
- Container
- ResolveEnvPath
- Command reference
- replConfig
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
- Role
- bytes.Buffer
- hostpath_test.go
- fakeTransport
- ReplSite
- containerTransport
- blocks.go
- strings.Builder

## God Nodes (most connected - your core abstractions)
1. `App` - 217 edges
2. `newTestOps()` - 167 edges
3. `cli_test.go` - 148 edges
4. `Config` - 146 edges
5. `NewCluster()` - 141 edges
6. `ctrCfg()` - 133 edges
7. `Role` - 132 edges
8. `Per-package doubles` - 126 edges
9. `loadK8s()` - 119 edges
10. `newCapMgr()` - 114 edges

## Surprising Connections (you probably didn't know these)
- `Repository layout` --references--> `Execute()`  [INFERRED]
  docs/developer.md → internal/cli/root.go
- `The preflight` --references--> `info()`  [INFERRED]
  docs/operations.md → internal/k8s/checkreport.go
- `Per-package doubles` --references--> `sample()`  [INFERRED]
  docs/test.md → internal/abbrev/abbrev_test.go
- `annotate_test.go` --references--> `Annotate()`  [INFERRED]
  docs/test.md → internal/broker/annotate.go
- `annotate_test.go` --references--> `TestAnnotateRoundTripPreservesBlocks()`  [INFERRED]
  docs/test.md → internal/broker/annotate_test.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (116 total, 6 thin omitted)

### Community 0 - "App"
Cohesion: 0.07
Nodes (96): App, logArgs(), confirmAction(), childExit(), confirmImport(), runExport(), runImport(), diagnosticsOptions() (+88 more)

### Community 1 - "newTestOps"
Cohesion: 0.07
Nodes (72): recOutput, broker_test.go, countContains(), leaderCalls(), newTestOps(), TestCountContains(), TestDiagnostics(), TestDisableDefaultUsers() (+64 more)

### Community 2 - "ImportPlan"
Cohesion: 0.15
Nodes (20): ImportResult, PlannedSection, Region, isPragma(), defaultVPNFirst(), ImportPlan, keepRegion(), preambleForApply() (+12 more)

### Community 3 - "context.Context"
Cohesion: 0.07
Nodes (6): scriptedMate, uploadErrTransport, EnvRunner, Runner, badJSONRunner, fixtureRunner

### Community 4 - "cli_test.go"
Cohesion: 0.06
Nodes (93): cli_test.go, Shared env fixtures, watch_test.go, allowRuntime(), captureStderr(), echoRunner(), fakeBinaryOnPath(), runCtr() (+85 more)

### Community 5 - "config_test.go"
Cohesion: 0.05
Nodes (65): Image, config_test.go, TestValidateImageReference(), validPlatformConfig(), atoiPrefix(), assertContainerBlockDefaults(), assertContainerScaling(), haNodesConfig() (+57 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (50): Commands, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker cli, solace-util broker configure (+42 more)

### Community 7 - "Manager"
Cohesion: 0.07
Nodes (18): runtime_test.go, certDigest(), composeNeedsSecretValues(), exactName(), Manager, orNone(), platformTitle(), secretSummary() (+10 more)

### Community 8 - "K8sConfig"
Cohesion: 0.11
Nodes (15): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, LoadBalancer, cut() (+7 more)

### Community 9 - "runPlatform"
Cohesion: 0.28
Nodes (17): platform_test.go, runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection() (+9 more)

### Community 10 - "CheckCommand"
Cohesion: 0.10
Nodes (35): cmdField, commandRules, guardedCmd, execguard_test.go, decodeCommand(), TestCommandArgsDoesNotAliasCommand(), TestCommandUnmarshal(), Command (+27 more)

### Community 11 - "cli_test.go"
Cohesion: 0.12
Nodes (41): leaderRun, opCall, capture(), captureStdout(), collectPaths(), failDisableDefaultUsersUpload(), firstLine(), isDeletePod() (+33 more)

### Community 12 - "NewCluster"
Cohesion: 0.09
Nodes (64): operator_test.go, GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), RenderOperator(), splitAfterNamespace(), splitOperatorBundle() (+56 more)

### Community 13 - "bannerType"
Cohesion: 0.47
Nodes (6): transcript_test.go, bannerType(), stripTranscript(), TestTranscriptBannerType(), TestTranscriptStripCutsAtBegin(), TestTranscriptTeardownFixtureResetsNotDeletes()

### Community 14 - "convert_test.go"
Cohesion: 0.07
Nodes (67): doc, Result, segment, vars, The keys, convert_test.go, boolOf(), commentSafe() (+59 more)

### Community 15 - "transform_test.go"
Cohesion: 0.10
Nodes (60): Block, TargetState, transform_test.go, injectedBlock(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects(), ClearTargetVirtualHostnames() (+52 more)

### Community 16 - "manager_test.go"
Cohesion: 0.10
Nodes (95): manager_test.go, GuardPodmanEUID(), fileExists(), assertMode(), containsStr(), ctrCfg(), hasCall(), newCapMgr() (+87 more)

### Community 17 - "commands.go"
Cohesion: 0.14
Nodes (69): opFunc, roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd() (+61 more)

### Community 18 - "coverage_test.go"
Cohesion: 0.07
Nodes (67): coverage_test.go, verify_ops_test.go, matchCLI(), ranContains(), writeFile(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError() (+59 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "Load"
Cohesion: 0.06
Nodes (53): scalingTier, Node roles, Reading this reference, scaling_test.go, minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError() (+45 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Set"
Cohesion: 0.36
Nodes (12): abbrev_test.go, checkShort(), New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings(), TestKind() (+4 more)

### Community 23 - "MateChannel"
Cohesion: 0.14
Nodes (31): CLIRunner, fakeRun, matechannel_test.go, MateChannel, NewCLIMate(), mateReplies(), newTestMate(), TestLocalMateRunsThroughOps() (+23 more)

### Community 24 - "newLocalOps"
Cohesion: 0.16
Nodes (40): verify_local_test.go, cliScriptPath(), isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn(), TestDefaultLocalAddrs() (+32 more)

### Community 25 - "runner_test.go"
Cohesion: 0.06
Nodes (52): interactiveFailRunner, Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication (+44 more)

### Community 26 - "renderDriver"
Cohesion: 0.15
Nodes (22): chunk, chunkResult, driver_test.go, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce() (+14 more)

### Community 27 - "testing.T"
Cohesion: 0.07
Nodes (77): completion_test.go, examples_test.go, statusreport_test.go, runRoot(), TestBashEnvGivenToEnvFlag(), TestConfigureServerCertsRefusesASecretItDoesNotOwn(), TestConvertErrorPaths(), TestConvertParseError() (+69 more)

### Community 28 - "importops_test.go"
Cohesion: 0.09
Nodes (43): importops_test.go, describeScope(), planSections(), chunkNames(), driverAllOK(), driverChunkBody(), driverChunkNames(), driverFailAt() (+35 more)

### Community 29 - "deploymentItem"
Cohesion: 0.18
Nodes (9): imageFromDeployment(), desiredWatch(), splitWatch(), TestWatchFromContainersReadsEveryAllNamespacesSpelling(), watchFromContainers(), watchFromDeployment(), deploymentItem, deploymentList (+1 more)

### Community 31 - "Config"
Cohesion: 0.06
Nodes (26): keyValueEntries, TestRedundancyEnabled(), cpuSetCount(), cpuSetRange(), TestCPUSetCount(), TestCPUSetRange(), checkCredentialChars(), CheckSecretKeyRef() (+18 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.15
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 33 - "internal/config"
Cohesion: 0.12
Nodes (20): artifactvalues_test.go, duration_test.go, internal/config, tlssecret_test.go, portOrRange(), TestValidateContainerArtifactValues(), TestValidateK8sArtifactValues(), TestValidContainerPort() (+12 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.06
Nodes (51): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domaincerts_resolve_test.go, domaincerts_test.go, caNameSafe() (+43 more)

### Community 35 - "watchLoop"
Cohesion: 0.32
Nodes (5): watchInterval(), stdoutIsTerminal(), watchLabel(), watchLoop(), withWatch()

### Community 36 - "github.com/spf13/cobra.ShellCompDirective"
Cohesion: 0.33
Nodes (6): completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), isEnvFileName(), matching()

### Community 37 - "NewCluster"
Cohesion: 0.10
Nodes (41): checkreport_test.go, preflight_test.go, TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestOperatorInstallVerdict(), TestOperatorInstallVerdictShowsOnceInEveryView(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne(), TestOperatorValidateAsksNothingAboutTheBroker(), TestStorageRows() (+33 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.07
Nodes (45): annotate_test.go, applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV(), readMarker() (+37 more)

### Community 40 - "Sink"
Cohesion: 0.05
Nodes (73): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, diff_test.go, internal/output, output_test.go (+65 more)

### Community 41 - "Compose"
Cohesion: 0.09
Nodes (36): servercert_test.go, ContainerNoFile(), NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), Compose(), composeEscape(), ContainerSecrets(), containerSecretSpecs() (+28 more)

### Community 42 - "inject_test.go"
Cohesion: 0.17
Nodes (28): lineRole, serviceLine, shutdownStyle, svcKey, inject_test.go, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "Coverage"
Cohesion: 0.15
Nodes (13): Coverage, Fixtures and doubles, internal/abbrev, internal/convert, internal/engine, internal/examples, internal/render, Running the tests (+5 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.11
Nodes (33): exportconfigReadCounter, opRunner, euid_test.go, exportconfig_test.go, internal/cli, runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand() (+25 more)

### Community 47 - "VPNRepl"
Cohesion: 0.07
Nodes (21): AdminState, cliMate, QueueState, replMate, rejectionIn(), ReplRole, VPNRepl, ReplicationConfigResult (+13 more)

### Community 48 - "localCfg"
Cohesion: 0.10
Nodes (47): semp_test.go, sempmate_test.go, bridgeHostPort(), sempPort(), sempV1OK(), assertNoPasswordInArgv(), TestBackupRevertActivityBridgePlaintextOnly(), TestBackupRevertActivityBridgeTLSNoCA() (+39 more)

### Community 49 - "config_test.go"
Cohesion: 0.19
Nodes (14): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, showCmd, span, hostMatches(), shortHost() (+6 more)

### Community 50 - "eqArgs"
Cohesion: 0.10
Nodes (44): check_test.go, cluster_test.go, internal/k8s, names_test.go, ops_test.go, transport_test.go, TestCheckStopsProbingWhenUnreachable(), TestReachable() (+36 more)

### Community 51 - "newRootCmd"
Cohesion: 0.11
Nodes (32): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, TestFlagShorthandsAreConsistent(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb() (+24 more)

### Community 52 - "scripts_test.go"
Cohesion: 0.06
Nodes (38): scripts_test.go, validCLILine(), archiveConfigsScript(), assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), enableDefaultUsersScript() (+30 more)

### Community 53 - "prep_test.go"
Cohesion: 0.12
Nodes (43): prep_test.go, haCfg(), GenBroker(), joinManifests(), adminCfg(), secretNamesIn(), TestAdminSecretStatesEmitTheirArtifacts(), TestBrokerPodsExist() (+35 more)

### Community 54 - ".validate"
Cohesion: 0.08
Nodes (27): solace-util, solace-util broker, solace-util operator, A replication switch refuses before changing anything, A wrapper runtime is refused, Bringing up a fresh cluster, Docker and Podman mechanics, Docker compose secrets need compose 2.23.1+ (+19 more)

### Community 55 - ".dump"
Cohesion: 0.21
Nodes (33): ops_k8s_test.go, opCanI(), TestRemoveBrokerKeepsTheNamespaceWhileItsPVCsRemain(), writeK8sDeployAllEnv(), opsK8sAnswer(), opsK8sPods(), opsK8sRun(), opsK8sSilenced() (+25 more)

### Community 56 - ".kubectl"
Cohesion: 0.06
Nodes (18): runtime_test.go, BrokerPodSuffixShape(), Cluster, normalizeToList(), discoveryServiceName(), HARoles(), lbServiceName(), pvcName() (+10 more)

### Community 57 - "replication_test.go"
Cohesion: 0.09
Nodes (49): replRig, Abbreviations, Commands, Flag shorthands, Platforms, Configuring a site, Data replication, Reaching the mate (+41 more)

### Community 58 - "deploy_test.go"
Cohesion: 0.16
Nodes (26): deploy_test.go, brokerJSON(), logBuf(), TestConfirmBrokerAppliedAcceptsTheObjectItApplied(), TestConfirmBrokerAppliedFailsOnADifferentObject(), TestConfirmBrokerAppliedFailsWhenTheObjectIsNotThere(), TestConfirmBrokerAppliedSurfacesReadFailures(), TestConfirmBrokerAppliedTreatsSilenceAsSkipped() (+18 more)

### Community 59 - ".rpc"
Cohesion: 0.10
Nodes (14): backupTarget, Credential, sempMate, transport_test.go, schemaTransport(), curlConfigFlag(), curlConfigLine(), httpBody() (+6 more)

### Community 60 - "confirm.go"
Cohesion: 0.20
Nodes (18): layer, Injectable seams, completionShell(), newCompletionCmd(), confirmDelete(), confirmDowngrade(), confirmGate(), confirmLayer() (+10 more)

### Community 61 - ".configRows"
Cohesion: 0.12
Nodes (28): orNone(), orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminPassState(), adminSecretRow() (+20 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.19
Nodes (40): rootless_test.go, failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+32 more)

### Community 63 - "exit_test.go"
Cohesion: 0.15
Nodes (20): runFailRunner, exit_test.go, Summary, TestMain(), TestOpsGatherDiagnosticsRefusesDaysBelowOne(), ExitCode(), childStatusError(), k8sEnv() (+12 more)

### Community 64 - "BuildSwitchPlan"
Cohesion: 0.17
Nodes (22): PhaseKind, SiteState, SwitchAction, SwitchPhase, BuildSwitchPlan(), SwitchPlan, promoteConfirmed(), replPrecheckError() (+14 more)

### Community 65 - "switchplan_test.go"
Cohesion: 0.28
Nodes (21): switchplan_test.go, ExecuteSwitchPlan(), SwitchPreflight(), ordersToA(), replSites(), TestExecuteSwitchPlanEmptyDoesNothing(), TestExecuteSwitchPlanOrder(), TestExecuteSwitchPlanRefusesASiteItHasNoChannelFor() (+13 more)

### Community 66 - "opPerformReplication"
Cohesion: 0.13
Nodes (16): mateChannelFunc, confirmActionStrict(), refuseUnconfirmed(), lineSink(), progress(), step(), confirmReplicationConfig(), confirmReplicationSwitch() (+8 more)

### Community 67 - "Capture"
Cohesion: 0.17
Nodes (13): Capture, Omission, appliance_test.go, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+5 more)

### Community 68 - "Set"
Cohesion: 0.16
Nodes (7): Entry, Set, TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), newPlatforms(), newRoles(), RoleAbbrev()

### Community 69 - "Goldens"
Cohesion: 0.28
Nodes (8): Goldens, commanddoc_test.go, TestAbbreviationDocs(), firstDiff(), TestCommandDocs(), TestEveryDocCarriesTheSupportNotice(), firstDiff(), TestSampleYAMLMatchesTheFullExample()

### Community 70 - ".failed"
Cohesion: 0.23
Nodes (7): Every destructive command confirms, Removing a broker: what stays, what goes, Removing the operator does not always remove it, The layer flag raises the question, The namespace is only offered when it is empty, TestCheckReportSkipsEmptySections(), validateScope

### Community 71 - "TestAdminPasswordPreflight"
Cohesion: 0.30
Nodes (12): adminsecret_test.go, AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminPasswordPreflight(), TestAdminSecretInUseFollowsTheStates(), TestOperatorAdminSecretNameMatchesTheBrokerNames(), TestReadAdminPasswordArgvAndDecoding() (+4 more)

### Community 72 - "TestGeneratorPageEmbedsTheCLI"
Cohesion: 0.18
Nodes (12): generatorpage_test.go, generatorDefaults(), gunzipBase64(), gzipBase64(), scalingOf(), templateActions(), TestGeneratorPageEmbedsTheCLI(), genBlock (+4 more)

### Community 73 - "namespace_test.go"
Cohesion: 0.29
Nodes (12): namespace_test.go, occCfg(), occCluster(), TestNamespaceContentsAsksOneQuestion(), TestNamespaceContentsDiscountsWhatKubernetesPutsThere(), TestNamespaceContentsErrorMeansOccupied(), TestNamespaceContentsIgnoresClusterPolicyObjects(), TestNamespaceContentsReportsRealOccupants() (+4 more)

### Community 74 - "redundancy_test.go"
Cohesion: 0.21
Nodes (12): redundancy_test.go, pskCfg(), TestFillStandaloneNodeNameIgnoresAnEmptyHostname(), TestFillStandaloneNodeNameKeepsAConfiguredName(), TestFillStandaloneNodeNameNeverFillsHA(), TestFillStandaloneNodeNameUsesTheHost(), TestNothingGeneratesThePSK(), TestPSKEnvSatisfiesTheContainerRequirement() (+4 more)

### Community 75 - "usagef"
Cohesion: 0.06
Nodes (52): childExitError, usageError, How a short form is resolved, applyAliases(), argumentLine(), availableSubs(), writeCommand(), noRolePositional() (+44 more)

### Community 76 - "podName"
Cohesion: 0.42
Nodes (3): podName(), shSingleQuote(), kubectlTransport

### Community 77 - "command_test.go"
Cohesion: 0.18
Nodes (12): command_test.go, TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRenamedRuntimeKeysFailLoud(), TestRuntimeDefaults(), TestRuntimeExplicitValueSurvivesDefaults() (+4 more)

### Community 78 - "The eight rules"
Cohesion: 0.18
Nodes (11): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+3 more)

### Community 79 - "Container"
Cohesion: 0.29
Nodes (6): Container, DockerConfig, Network, PodmanConfig, Ulimits, HealthCheck

### Community 80 - ".decodeScalingEntry"
Cohesion: 0.32
Nodes (5): scalingKey, scalingSpelling, Scaling, scalingKeyIndex(), scalingKeyList()

### Community 81 - "replicationops_test.go"
Cohesion: 0.16
Nodes (35): replicationops_test.go, hasCall(), reenableLine(), replicationVPNLines(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching() (+27 more)

### Community 82 - "k8s/matechannel_test.go"
Cohesion: 0.19
Nodes (28): matechannel_test.go, CLIArg(), CLIScriptPath(), execArgs(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf() (+20 more)

### Community 83 - "limits_test.go"
Cohesion: 0.07
Nodes (53): internal/tools/vulnjudge, limits_test.go, main_test.go, delegateSetting(), limitText(), missingControllers(), parseLimit(), parseUnitProps() (+45 more)

### Community 84 - "Developer guide"
Cohesion: 0.25
Nodes (8): Build, Dev script tasks, Developer guide, Gates, Releases, Repository layout, Tests, Toolchain pin

### Community 85 - "Exporting and importing configuration"
Cohesion: 0.29
Nodes (7): Broker scope is a fixed classification, Exporting and importing configuration, Handling the artifact, Kubernetes-specific hazards, The import lifecycle, Two detectors, in order, What export captures, and why it is not a backup

### Community 86 - "k8s/inspect.go"
Cohesion: 0.10
Nodes (37): inspect_test.go, age(), anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), podHealth(), pvcLevel() (+29 more)

### Community 87 - ".ConfigureReplication"
Cohesion: 0.08
Nodes (67): BrokerType, colSpan, replication_test.go, cliTransport(), containsEndpoint(), dashSpans(), flagByte(), MateConfig (+59 more)

### Community 88 - "AdminSecret"
Cohesion: 0.11
Nodes (34): The env-file generator page, secrets_test.go, GenSecrets(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), readTLSMaterial() (+26 more)

### Community 89 - "render_test.go"
Cohesion: 0.13
Nodes (45): render_test.go, boolStr(), BrokerCR(), ComposeProject(), assertNoCheckoutPath(), containerArtifacts(), healthCheckFixture(), load() (+37 more)

### Community 90 - "importIgnore"
Cohesion: 0.47
Nodes (5): importIgnore(), TestImportIgnoreKeepsUnclassifiedSectionsInTheDiff(), TestImportOpsImportIgnore(), BridgeEnablementInverted(), TestShutdownBridgesLinesAreExcusedByTheDiff()

### Community 91 - "RuleFor"
Cohesion: 0.14
Nodes (33): Disposition, SectionRule, importdoc_test.go, sections_test.go, renderImportDocs(), TestEverySectionRuleHasAReason(), TestImportDocs(), TestOnlyOneSectionIsAppliedFirst() (+25 more)

### Community 92 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 93 - "Container"
Cohesion: 0.19
Nodes (9): guardCommandOf(), applyContainerBlockDefaults(), podmanRunUser(), setDefault(), setDefaultCmd(), setDefaultInt(), xdgConfigHome(), containerMem() (+1 more)

### Community 95 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 96 - "Command reference"
Cohesion: 0.50
Nodes (4): Command reference, Global flags, Index, Tree

### Community 97 - "replConfig"
Cohesion: 0.19
Nodes (15): Replication, secretRef, replication_test.go, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded() (+7 more)

### Community 102 - "DetectPlatforms"
Cohesion: 0.24
Nodes (17): examples_test.go, Example, parseError(), DetectPlatforms(), All(), Get(), List(), Names() (+9 more)

### Community 104 - "Platforms"
Cohesion: 0.24
Nodes (10): platform_test.go, PlatformAbbrev(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint(), TestDetectPlatformsMissingFile(), TestEveryPlatformHasExactlyOneShortForm(), TestParsePlatform(), TestParsePlatformErrorTeachesCanonical() (+2 more)

### Community 107 - "operatorversion_test.go"
Cohesion: 0.12
Nodes (24): operatorversion_test.go, operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsNotSkippedByAFailedRead(), TestConfirmNoDowngradeIsSilentInAPreview(), TestConfirmNoDowngradeProceedsWhenConfirmAccepts(), TestConfirmNoDowngradeRefusedWhenConfirmDeclines() (+16 more)

### Community 109 - "Config"
Cohesion: 0.10
Nodes (25): AdditionalUser, Node, Redundancy, SEMP, TLS, adminsecret_test.go, imagepullsecret_test.go, adminCfg() (+17 more)

### Community 110 - "config_ops_test.go"
Cohesion: 0.21
Nodes (12): config_ops_test.go, TestEnableDefaultUsersStartsEveryDefaultUserBackUp(), TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(), TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator(), TestExecShellScriptRefusesAnUnsafeNameBeforeUploading() (+4 more)

### Community 111 - ".OperatorApply"
Cohesion: 0.23
Nodes (7): operatorEscalationProbes(), operatorProbes(), canIAnswer(), canIRefused(), missingNamespace(), operatorRegcred(), probe

### Community 112 - "capRunner"
Cohesion: 0.13
Nodes (21): capCall, capRunner, New(), NewManager(), callIndex(), TestManagerLogsCLIShell(), mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime() (+13 more)

### Community 113 - "storage_test.go"
Cohesion: 0.36
Nodes (9): storage_test.go, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey(), TestMsgNodeSizeIsOptionalOnlyWithCustomMounts(), TestRoleWordAndTitleAgree() (+1 more)

### Community 116 - "certFixture"
Cohesion: 0.11
Nodes (35): internal/container, Per-package doubles, preflight_test.go, secrets_test.go, replShowMate(), replVPNTable(), maskedKeys(), TestComposeSecretEnvIsTheOnlyChildEnvironment() (+27 more)

### Community 118 - "Container State Inspection"
Cohesion: 0.10
Nodes (25): containerState, healthState, inspect_test.go, decodeInspect(), orUnknown(), printable(), inspectReply(), kvValue() (+17 more)

### Community 121 - "Role"
Cohesion: 0.06
Nodes (19): cliRunNames(), field(), Ops, runCLISkeleton(), validName(), validVPNName(), gatherConfigsScript(), noReleaseActivityScript() (+11 more)

### Community 126 - "bytes.Buffer"
Cohesion: 0.14
Nodes (23): watch_test.go, subtractWatch(), deployJSON(), operatorItem(), TestAmbiguousOperatorAbortsTheDeploy(), TestDesiredWatchAppendsTheBrokerNamespace(), TestFindOperatorDeploymentIgnoresOtherDeployments(), TestFindOperatorDeploymentIsScopedByNamespace() (+15 more)

### Community 128 - "hostpath_test.go"
Cohesion: 0.08
Nodes (34): hostpath_test.go, guardConfig(), BaseName(), CheckHostPath(), expandTilde(), expandTildeToken(), HasPathSeparator(), IsAbsHostPath() (+26 more)

### Community 130 - "fakeTransport"
Cohesion: 0.18
Nodes (12): covUploadFailsTo, downloadErrTransport, fakeTransport, recDownload, recRun, recUpload, recUploadFile, runErrMatchTransport (+4 more)

### Community 132 - "ReplSite"
Cohesion: 0.16
Nodes (13): ReplVia, ReplViaKube, ReplViaSEMP, missingListedVPNs(), PlannedRoles(), RoleAtSite(), TestRoleAtSiteIsTheComplement(), replVPNNames() (+5 more)

### Community 139 - "blocks.go"
Cohesion: 0.13
Nodes (35): token, blocks_test.go, internal/broker, omit_test.go, Annotate(), TestAnnotateIsDeterministic(), TestAnnotateRegionAndSectionFencesAreBalanced(), TestAnnotateRoundTripCaptureIsMarked() (+27 more)

### Community 155 - "strings.Builder"
Cohesion: 0.09
Nodes (33): shorthand, WeightedNodeTerm, mdRow(), renderAbbrevDocs(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable() (+25 more)

## Knowledge Gaps
- **13 isolated node(s):** `solace`, `NO_COLOR`, `bash/000-env.sh`, `bash/010-deploy-operator.sh`, `bash/020-deploy-broker.sh` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 188 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **6 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `App` to `opPerformReplication`, `context.Context`, `watchLoop`, `cli_test.go`, `github.com/spf13/cobra.ShellCompDirective`, `runPlatform`, `usagef`, `exportconfig_test.go`, `Config`, `cli_test.go`, `commands.go`, `config_test.go`, `newRootCmd`, `certFixture`, `.dump`, `replication_test.go`, `confirm.go`?**
  _High betweenness centrality (0.041) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `newTestOps()` (e.g. with `importops_test.go` and `Injectable seams`) actually correct?**
  _`newTestOps()` has 117 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `NO_COLOR`, `bash/000-env.sh` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.06572164948453608 - nodes in this community are weakly interconnected._
- **Why does `Per-package doubles` connect `certFixture` to `App`, `newTestOps`, `fakeTransport`, `context.Context`, `cli_test.go`, `config_test.go`, `Manager`, `cli_test.go`, `NewCluster`, `convert_test.go`, `transform_test.go`, `manager_test.go`, `Load`, `Set`, `MateChannel`, `newLocalOps`, `runner_test.go`, `strings.Builder`, `importops_test.go`, `testing.T`, `deploymentItem`, `Coverage`, `exportconfig_test.go`, `VPNRepl`, `localCfg`, `eqArgs`, `prep_test.go`, `.dump`, `.kubectl`, `replication_test.go`, `.rpc`, `exit_test.go`, `BuildSwitchPlan`, `switchplan_test.go`, `opPerformReplication`, `k8s/matechannel_test.go`, `k8s/inspect.go`, `.ConfigureReplication`, `AdminSecret`, `ResolveEnvPath`, `replConfig`, `operatorversion_test.go`, `.OperatorApply`, `capRunner`, `Role`?**
  _High betweenness centrality (0.030) - this node is a cross-community bridge._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.07387387387387387 - nodes in this community are weakly interconnected._
- **Why does `Config` connect `Config` to `App`, `hostpath_test.go`, `newTestOps`, `cli_test.go`, `config_test.go`, `ReplSite`, `containerTransport`, `K8sConfig`, `Manager`, `CheckCommand`, `cli_test.go`, `NewCluster`, `convert_test.go`, `manager_test.go`, `Load`, `strings.Builder`, `deploymentItem`, `Config`, `internal/config`, `ResolveDomainCerts`, `NewCluster`, `Compose`, `Coverage`, `localCfg`, `config_test.go`, `eqArgs`, `prep_test.go`, `.dump`, `.kubectl`, `.configRows`, `TestAdminPasswordPreflight`, `namespace_test.go`, `redundancy_test.go`, `podName`, `Container`, `.decodeScalingEntry`, `k8s/matechannel_test.go`, `AdminSecret`, `render_test.go`, `Container`, `replConfig`, `.OperatorApply`, `capRunner`, `storage_test.go`, `certFixture`, `Role`?**
  _High betweenness centrality (0.023) - this node is a cross-community bridge._
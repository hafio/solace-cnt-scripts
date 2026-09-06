# Graph Report - solace-k8-scripts  (2026-09-05)

## Corpus Check
- 135 files · ~298,470 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2382 nodes · 7356 edges · 97 communities (89 shown, 8 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 1188 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `72a3a798`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- runRoot
- Sink
- context.Context
- captureStdout
- config_test.go
- Commands
- Manager
- config.go
- load.go
- Command
- NewTransport
- NewCluster
- eqArgs
- convert_test.go
- internal/k8s
- manager_test.go
- cli/platform_test.go
- newTestOps
- dev.sh
- prep_test.go
- dev.ps1
- commanddoc_test.go
- judge
- verify_local_test.go
- Role
- Load
- Platform
- Get
- solace
- .Run
- CLAUDE.md
- .Preflight
- scaling_test.go
- runRootWith
- .resolveSecretRefs
- haCfg
- statusreport_test.go
- Compose
- runner_test.go
- k8s/runtime_test.go
- Config
- Set
- prep.go
- cli_test.go
- testing.T
- command_test.go
- completion_test.go
- manager.go
- Probe
- newRootCmd
- podmanEnv
- newEchoMgr
- Operations
- Test catalogue
- .AdditionalUsers
- internal/cli
- internal/broker
- internal/config
- internal/container
- .configRows
- .releaseToBackup
- fileExists
- Configuration
- runP1
- containerTransport
- compareVersions
- Developer guide
- operatorversion_test.go
- semp_test.go
- Live-environment probe harness (`solace-itest`)
- runP2
- scripts.go
- .ServerCert
- ResolveEnvPath
- resolveEnv
- New
- Fixtures and doubles
- internal/tools/itest
- itest_test.go
- age
- inspect_test.go
- ParseRole
- Ops
- inspect.go
- kubectlTransport
- TestDefaultGenPSK
- render.go
- .releaseLocalActivity
- Abbreviations
- internal/engine
- .MateRevertActivity
- eqArgs

## God Nodes (most connected - your core abstractions)
1. `ctrCfg()` - 97 edges
2. `newTestOps()` - 89 edges
3. `newCapMgr()` - 88 edges
4. `Config` - 86 edges
5. `Role` - 85 edges
6. `Platform` - 75 edges
7. `NewCluster()` - 73 edges
8. `bg()` - 69 edges
9. `Commands` - 68 edges
10. `Manager` - 65 edges

## Surprising Connections (you probably didn't know these)
- `TestNewDefaults()` --calls--> `New()`  [INFERRED]
  internal/broker/coverage_test.go → internal/broker/broker.go
- `TestField()` --calls--> `field()`  [INFERRED]
  internal/broker/broker_test.go → internal/broker/broker.go
- `TestFieldLabelWithoutColon()` --calls--> `field()`  [INFERRED]
  internal/broker/coverage_test.go → internal/broker/broker.go
- `TestCountContains()` --calls--> `countContains()`  [INFERRED]
  internal/broker/broker_test.go → internal/broker/broker.go
- `TestContainsAnyFold()` --calls--> `containsAnyFold()`  [INFERRED]
  internal/broker/broker_test.go → internal/broker/broker.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (97 total, 8 thin omitted)

### Community 0 - "bg"
Cohesion: 0.06
Nodes (107): TestConfirmFlagShortcuts(), TestConfirmNonTTY(), TestCtrManagerConfirmWiring(), App, lineSink(), progress(), step(), warn() (+99 more)

### Community 1 - "runRoot"
Cohesion: 0.12
Nodes (25): captureStderr(), fakeBinaryOnPath(), runRoot(), TestBashEnvGivenToEnvFlag(), TestBinaryAnnouncementWiring(), TestConvertErrorPaths(), TestConvertParseError(), TestConvertRoundTrip() (+17 more)

### Community 2 - "Sink"
Cohesion: 0.17
Nodes (6): Level, Sink, pad(), TestLevelTagsMapToTheHouseTags(), KV, Row

### Community 3 - "context.Context"
Cohesion: 0.09
Nodes (8): context.Context, Runner, Cluster, Cluster, Cluster, Cluster, badJSONRunner, fixtureRunner

### Community 4 - "captureStdout"
Cohesion: 0.12
Nodes (22): opCall, opRunner, capture(), captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch() (+14 more)

### Community 5 - "config_test.go"
Cohesion: 0.07
Nodes (48): assertContainerBlockDefaults(), assertContainerScaling(), Config, haNodesConfig(), TestApplyDefaultsDocker(), TestApplyDefaultsK8s(), TestApplyDefaultsK8sTLS(), TestApplyDefaultsPodmanRootful() (+40 more)

### Community 6 - "Commands"
Cohesion: 0.03
Nodes (68): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util check (+60 more)

### Community 8 - "config.go"
Cohesion: 0.10
Nodes (27): Admin, Broker, Container, ContainerSecurity, DockerConfig, DomainCerts, Image, Network (+19 more)

### Community 9 - "load.go"
Cohesion: 0.22
Nodes (11): TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config, setDefault() (+3 more)

### Community 10 - "Command"
Cohesion: 0.06
Nodes (49): commandRules, keyValueEntries, Command, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules() (+41 more)

### Community 11 - "NewTransport"
Cohesion: 0.20
Nodes (13): capCall, capRunner, New(), Transport, NewTransport(), dockerCfg(), podmanCfg(), TestTransportCopy() (+5 more)

### Community 12 - "NewCluster"
Cohesion: 0.10
Nodes (44): TestCheckStopsProbingWhenUnreachable(), TestResolveStorageClass(), TestCheckDeployGroupsAndOrdersSections(), TestCheckDeployNeverPrintsASecret(), TestCheckDeployReportsEveryFailureInOneRun(), TestCheckDeploySparseConfigExplainsItself(), TestCheckDeployTagsEchoedConfigAsINFO(), TestStorageRows() (+36 more)

### Community 13 - "eqArgs"
Cohesion: 0.17
Nodes (23): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), TestCLIAndShellAreInteractive() (+15 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (63): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+55 more)

### Community 15 - "internal/k8s"
Cohesion: 0.12
Nodes (16): check_test.go, checkreport_test.go, cluster_test.go, deploy_test.go, inspect_test.go, internal/k8s, names_test.go, operator_test.go (+8 more)

### Community 16 - "manager_test.go"
Cohesion: 0.10
Nodes (69): containsStr(), ctrCfg(), hasCall(), newCapMgr(), TestContainerRunningMatchesNameExactly(), TestManagerCheckDNSFailsLoudInHA(), TestManagerCheckPodmanEUID(), TestManagerCheckReachableError() (+61 more)

### Community 17 - "cli/platform_test.go"
Cohesion: 0.26
Nodes (17): App, runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection() (+9 more)

### Community 18 - "newTestOps"
Cohesion: 0.09
Nodes (46): appUsers(), Ops, newTestOps(), TestAdditionalUsers(), TestAdditionalUsersEmpty(), TestAdditionalUsersRejectsBadValues(), TestAdditionalUsersReportsExistingUser(), TestContainsAnyFold() (+38 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (22): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+14 more)

### Community 20 - "prep_test.go"
Cohesion: 0.13
Nodes (37): saCfg(), adminCfg(), Cluster, labelCluster(), nsBrokerJSON(), nsBuiltinJSON(), nsCluster(), nsItem() (+29 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (18): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+10 more)

### Community 22 - "commanddoc_test.go"
Cohesion: 0.16
Nodes (23): shorthand, github.com/spf13/pflag.FlagSet, mdRow(), renderAbbrevDocs(), TestAbbreviationDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeAbbrevTable() (+15 more)

### Community 23 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.19
Nodes (34): curlCalls(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestLeaderLocalAssertLeaderError(), TestLeaderLocalBadRoleArg() (+26 more)

### Community 25 - "Role"
Cohesion: 0.07
Nodes (25): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+17 more)

### Community 26 - "Load"
Cohesion: 0.23
Nodes (16): minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadReadError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors() (+8 more)

### Community 27 - "Platform"
Cohesion: 0.06
Nodes (106): layer, opFunc, roleOpFunc, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, io.Reader, io.Writer, os.File (+98 more)

### Community 28 - "Get"
Cohesion: 0.19
Nodes (21): Example, listExamples(), newExamplesCmd(), runExample(), TestExamplesEmitsToStdout(), TestExamplesRejectsAnUnknownName(), All(), Get() (+13 more)

### Community 31 - ".Run"
Cohesion: 0.18
Nodes (10): Echo, Exec, os/exec.Cmd, TestResolveMissingBinaryIsActionable(), MaskEnv(), Quote(), quoteTok(), Resolve() (+2 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.11
Nodes (16): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, Command reference, Global flags, Index (+8 more)

### Community 34 - "scaling_test.go"
Cohesion: 0.11
Nodes (19): scalingTier, containerMem(), Config, Config, setContainerMem(), TestApplyScalingTierDefaultsContainerBlocks(), TestApplyScalingTierDefaultsK8s(), TestApplyScalingTierDefaultsMemOverride() (+11 more)

### Community 35 - "runRootWith"
Cohesion: 0.13
Nodes (29): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+21 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "haCfg"
Cohesion: 0.12
Nodes (18): ProductKeyRoles(), haCfg(), TestHARoles(), TestProductKeyRoles(), TestCanIAnswerReadsTheLastLine(), TestCreateSecretsFailsWithoutAdminFields(), TestDeleteNamespaceProtected(), TestDeleteSecretsSkipsUnconfiguredAdminSecret() (+10 more)

### Community 38 - "statusreport_test.go"
Cohesion: 0.14
Nodes (21): operatorRunningImage(), ownsPod(), Cluster, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportRunningPicture(), TestBrokerReportSurfacesAReadFailure() (+13 more)

### Community 39 - "Compose"
Cohesion: 0.11
Nodes (35): opCtrGenArtifact(), opCtrGenSecrets(), HealthCheck, NodeIdentity, Compose(), ContainerSecrets(), EnvPairs(), escapePercent() (+27 more)

### Community 40 - "runner_test.go"
Cohesion: 0.11
Nodes (27): TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveRefusesCurrentDirectory(), verboseExec(), NewExec(), captureStdout() (+19 more)

### Community 41 - "k8s/runtime_test.go"
Cohesion: 0.43
Nodes (7): TestClusterHonoursRuntime(), TestExecutorRefusesUnapprovedRuntime(), TestRuntimeDefaultArgvUnchanged(), TestTransportHonoursRuntime(), unapprovedCfg(), withLeading(), wrappedCfg()

### Community 42 - "Config"
Cohesion: 0.14
Nodes (20): Replication, Scaling, TLS, tlsConfigured(), Config, OperatorNamespace(), GenOperator(), GenOperatorSecrets() (+12 more)

### Community 43 - "Set"
Cohesion: 0.15
Nodes (13): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+5 more)

### Community 44 - "prep.go"
Cohesion: 0.07
Nodes (34): bufio.Reader, foreignObjects(), GenSecrets(), Cluster, isBuiltinLabel(), isOurs(), joinManifests(), namespaceManifest() (+26 more)

### Community 45 - "cli_test.go"
Cohesion: 0.13
Nodes (36): allowRuntime(), collectPaths(), firstLine(), runCtr(), runStandalone(), runStatusStderr(), TestAnnounceCommandsNamesResolvedBinaries(), TestConfigStepsDoNotLeakSecrets() (+28 more)

### Community 46 - "testing.T"
Cohesion: 0.08
Nodes (57): testing.T, matchCLI(), ranContains(), TestLastLines(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle() (+49 more)

### Community 47 - "command_test.go"
Cohesion: 0.16
Nodes (13): decodeRuntime(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRuntimeDefaults() (+5 more)

### Community 48 - "completion_test.go"
Cohesion: 0.24
Nodes (17): runComplete(), TestAllowCommandOffersNoFiles(), TestCompletionHelpStillWorks(), TestCompletionNeedsAShell(), TestCompletionNoDescriptions(), TestCompletionScriptsGenerate(), TestDirFlagCompletesDirectories(), TestEnvFlagCompletesEnvFiles() (+9 more)

### Community 49 - "manager.go"
Cohesion: 0.17
Nodes (11): exactName(), orNone(), platformTitle(), replacePSKLine(), secretSummary(), setOrMissing(), solaceRows(), splitLimit() (+3 more)

### Community 50 - "Probe"
Cohesion: 0.17
Nodes (19): ids(), TestRegistryCoversEveryClaim(), TestRunProbesSkipsOffPlatformAndExitsOnRefuted(), TestSelectProbes(), probeP1(), probeP2(), probeP4(), Env (+11 more)

### Community 51 - "newRootCmd"
Cohesion: 0.12
Nodes (21): TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestDangerousVerbsHaveNoBareAlias(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownNoun(), TestStartStopHaveNoAlias(), TestAllowCommandIsRegisteredWhereItExecutes(), findCmd() (+13 more)

### Community 52 - "podmanEnv"
Cohesion: 0.16
Nodes (19): Env, podmanEnv(), TestCtrNameReadsTheConfiguredContainer(), TestP1DeclinedConsentTouchesNothing(), TestP1DryRunTouchesNothing(), TestP1RestoresAfterAFailureMidSequence(), TestP1RestoreWarnsLoudlyWhenItCannotRestore(), TestP1RunsStopRotateStartReadInOrder() (+11 more)

### Community 53 - "newEchoMgr"
Cohesion: 0.10
Nodes (22): bytes.Buffer, NewManager(), Manager, newEchoMgr(), rootlessNoFileMgr(), setNoFile(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless() (+14 more)

### Community 54 - "Operations"
Cohesion: 0.12
Nodes (16): A wrapper runtime is refused, Bringing up a fresh cluster, Docker and Podman mechanics, Docker compose secrets need compose 2.23.1+, Extra CLI users differ by platform, File descriptors on rootless podman, Operations, Podman secret flags (+8 more)

### Community 55 - "Test catalogue"
Cohesion: 0.12
Nodes (17): abbrev_test.go, convert_test.go, Coverage, examples_test.go, examples_test.go, internal/abbrev, internal/convert, internal/examples (+9 more)

### Community 56 - ".AdditionalUsers"
Cohesion: 0.24
Nodes (5): containsAnyFold(), validCLILine(), validCLIPassword(), validName(), Ops

### Community 57 - "internal/cli"
Cohesion: 0.25
Nodes (8): abbrevdoc_test.go, aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, internal/cli, platform_test.go

### Community 58 - "internal/broker"
Cohesion: 0.33
Nodes (6): broker_test.go, coverage_test.go, internal/broker, scripts_test.go, semp_test.go, verify_local_test.go

### Community 59 - "internal/config"
Cohesion: 0.33
Nodes (6): command_test.go, config_test.go, execguard_test.go, internal/config, platform_test.go, scaling_test.go

### Community 60 - "internal/container"
Cohesion: 0.40
Nodes (5): internal/container, manager_test.go, preflight_test.go, runtime_test.go, transport_test.go

### Community 61 - ".configRows"
Cohesion: 0.23
Nodes (13): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), failRow(), Cluster, info(), okRow() (+5 more)

### Community 63 - "fileExists"
Cohesion: 0.18
Nodes (11): fileExists(), assertMode(), maskedKeys(), TestManagerDeletePodmanRemovesUnit(), TestManagerDeletePodmanStopFailsServiceActiveBlocksRemoval(), TestManagerDeletePodmanStopFailsServiceInactiveProceeds(), TestManagerDeployDockerComposeWritesFile(), TestManagerDeployDockerDryRunMasksSecretEnv() (+3 more)

### Community 65 - "Configuration"
Cohesion: 0.22
Nodes (9): Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Scaling, Scaling tiers, Secrets, The command fields are executable content, The keys (+1 more)

### Community 66 - "runP1"
Cohesion: 0.33
Nodes (10): ctrName(), Env, Env, Result, p1Classify(), p1ReadMountedSecret(), p1Restore(), p1Throwaway() (+2 more)

### Community 68 - "compareVersions"
Cohesion: 0.36
Nodes (7): compareVersions(), imageTag(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestImageTag(), TestOperatorVersionWarningFiresOnlyOnADowngrade()

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "operatorversion_test.go"
Cohesion: 0.39
Nodes (7): operatorDeployJSON(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsSilentInAPreview(), TestConfirmNoDowngradeProceedsWhenConfirmAccepts(), TestConfirmNoDowngradeProceedsWhenVersionReadFails(), TestConfirmNoDowngradeRefusedWhenConfirmDeclines(), TestConfirmNoDowngradeRefusesByDefault()

### Community 71 - "semp_test.go"
Cohesion: 0.11
Nodes (30): sempPort(), assertNoPasswordInArgv(), TestCurlConfigLineEscapesQuotes(), TestHTTPBodyStripsHeaders(), TestLastHTTPStatus(), TestMateRevertActivityBridgePlaintextOnly(), TestMateRevertActivityBridgeTLSNoCA(), TestMateRevertActivityBridgeTLSWithCA() (+22 more)

### Community 72 - "Live-environment probe harness (`solace-itest`)"
Cohesion: 0.20
Nodes (10): After a run, Build and run, Flags, Live-environment probe harness (`solace-itest`), Operator-run only, Safety, Testing the harness itself, The probes (+2 more)

### Community 73 - "runP2"
Cohesion: 0.15
Nodes (16): containsPlatform(), platformNames(), resolvePlatform(), roleLabel(), TestP2Classify(), TestResolvePlatform(), writeEnv(), Env (+8 more)

### Community 74 - "scripts.go"
Cohesion: 0.09
Nodes (33): showCmd, additionalUsersScript(), assertLeaderScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), domainCertsScript(), gatherConfigsScript(), noReleaseActivityScript() (+25 more)

### Community 75 - ".ServerCert"
Cohesion: 0.47
Nodes (4): concatFiles(), serverCertFile(), serverCertScript(), TestServerCertScript()

### Community 76 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 77 - "resolveEnv"
Cohesion: 0.14
Nodes (17): Env, resolveEnv(), TestListProbesNeedsNoEnvFile(), TestUnknownProbeIsAUsageError(), TestVerdictNames(), TestVersionFlag(), Env, main() (+9 more)

### Community 78 - "New"
Cohesion: 0.14
Nodes (19): TestCheckReportSkipsEmptySections(), New(), NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow() (+11 more)

### Community 79 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 80 - "internal/tools/itest"
Cohesion: 0.50
Nodes (4): internal/tools/itest, itest_test.go, live_test.go, probes_live_test.go

### Community 81 - "itest_test.go"
Cohesion: 0.16
Nodes (19): Env, sinkOver(), TestConsent(), TestCPUAndMemoryConversion(), testEnv(), TestP1Classify(), TestP1ThrowawayIsUsable(), TestP4Classify() (+11 more)

### Community 82 - "age"
Cohesion: 0.37
Nodes (4): orNone(), age(), roleRank(), Cluster

### Community 83 - "inspect_test.go"
Cohesion: 0.18
Nodes (17): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+9 more)

### Community 84 - "ParseRole"
Cohesion: 0.40
Nodes (6): TestParseRole(), TestRoleAbbrevIsTheRoleValue(), TestRoleErrorTeachesBothSpellings(), TestRoleNames(), ParseRole(), RoleAbbrev()

### Community 86 - "inspect.go"
Cohesion: 0.16
Nodes (18): time.Time, brokerConditionRows(), conditionLevel(), findCondition(), Cluster, normalizeToList(), TestConditionLevelDegradesSafely(), TestNormalizeToListHandlesBothKubectlShapes() (+10 more)

### Community 89 - "render.go"
Cohesion: 0.16
Nodes (26): strings.Builder, boolStr(), BrokerCR(), containerSecretSpecs(), cut(), groupKey(), itoa(), parsePort() (+18 more)

### Community 91 - ".releaseLocalActivity"
Cohesion: 0.20
Nodes (12): countContains(), field(), Ops, hostMatches(), roleName(), shortHost(), TestMateActivityStateReadsTheMateColumn(), TestRoleName() (+4 more)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 100 - "internal/engine"
Cohesion: 0.67
Nodes (3): internal/engine, resolve_test.go, runner_test.go

### Community 103 - ".MateRevertActivity"
Cohesion: 0.21
Nodes (13): mateTarget, TestHTTPStatusHelpers(), anyHTTP2xx(), bridgeHostPort(), curlConfigFlag(), curlConfigLine(), Ops, httpBody() (+5 more)

### Community 106 - "eqArgs"
Cohesion: 0.14
Nodes (19): TestManagerLogsCLIShell(), TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestComposeSecretEnvNamesCannotBeSystemVars(), TestPreflightFailureStopsLifecycle(), TestPreflightFailureStopsTheDeploy(), TestPreflightHintIsPlatformShaped(), TestPreflightIsPreviewableUnderDryRun(), TestPreflightRunsBeforeAnything() (+11 more)

## Knowledge Gaps
- **179 isolated node(s):** `solace`, `showCmd`, `Config`, `operatorTmplVars`, `NO_COLOR` (+174 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `bg`, `captureStdout`, `config_test.go`, `Manager`, `config.go`, `load.go`, `Command`, `NewTransport`, `convert_test.go`, `manager_test.go`, `Load`, `scaling_test.go`, `Compose`, `manager.go`, `Probe`, `newEchoMgr`, `containerTransport`, `semp_test.go`, `runP2`, `Ops`, `eqArgs`?**
  _High betweenness centrality (0.084) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `bg`, `context.Context`, `captureStdout`, `Manager`, `config.go`, `NewTransport`, `NewCluster`, `convert_test.go`, `manager_test.go`, `newTestOps`, `prep_test.go`, `Role`, `haCfg`, `Compose`, `k8s/runtime_test.go`, `prep.go`, `cli_test.go`, `Probe`, `newEchoMgr`, `containerTransport`, `semp_test.go`, `Ops`, `kubectlTransport`, `render.go`, `eqArgs`?**
  _High betweenness centrality (0.059) - this node is a cross-community bridge._
- **Why does `Manager` connect `Manager` to `bg`, `context.Context`, `Config`, `manager.go`, `Probe`, `newEchoMgr`, `Platform`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Are the 10 inferred relationships involving `ctrCfg()` (e.g. with `TestComposeSecretEnvIsTheOnlyChildEnvironment()` and `TestComposeSecretEnvNamesCannotBeSystemVars()`) actually correct?**
  _`ctrCfg()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **Are the 50 inferred relationships involving `newTestOps()` (e.g. with `TestAdditionalUsersRunCLITransportError()` and `TestDiagnosticsMkdirError()`) actually correct?**
  _`newTestOps()` has 50 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `newCapMgr()` (e.g. with `NewManager()` and `TestComposeSecretEnvIsTheOnlyChildEnvironment()`) actually correct?**
  _`newCapMgr()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `showCmd`, `Config` to the rest of the system?**
  _179 weakly-connected nodes found - possible documentation gaps or missing edges._
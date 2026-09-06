# Graph Report - solace-k8-scripts  (2026-09-05)

## Corpus Check
- 135 files · ~294,340 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2370 nodes · 7318 edges · 93 communities (85 shown, 8 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 1177 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `72a3a798`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- cli_test.go
- Sink
- context.Context
- captureStdout
- testing.T
- Commands
- Manager
- config.go
- load.go
- validate.go
- NewTransport
- NewCluster
- eqArgs
- convert_test.go
- internal/k8s
- manager_test.go
- cli/platform_test.go
- secrets_test.go
- dev.sh
- prep_test.go
- dev.ps1
- README.md
- judge
- verify_local_test.go
- Role
- Load
- Platform
- completion_test.go
- solace
- Command
- CLAUDE.md
- .Preflight
- execguard_test.go
- .Run
- .resolveSecretRefs
- recRunner
- execguard.go
- Compose
- runner_test.go
- haCfg
- Config
- Set
- prep.go
- writeCtrStandaloneEnv
- newTestOps
- command_test.go
- platformTitle
- Probe
- newRootCmd
- podmanEnv
- newEchoMgr
- Operations
- Test catalogue
- internal/cli
- internal/broker
- internal/config
- internal/container
- .configRows
- Configuration
- runP1
- containerTransport
- Developer guide
- semp_test.go
- Live-environment probe harness (`solace-itest`)
- itest_test.go
- scripts.go
- resolveEnv
- New
- Fixtures and doubles
- internal/tools/itest
- runP4
- age
- inspect_test.go
- fakeRunner
- Ops
- inspect.go
- kubectlTransport
- render.go
- .releaseToBackup
- Troubleshooting
- Abbreviations
- Command reference
- internal/engine
- Cluster
- allowcommand_test.go
- .MateRevertActivity
- time.Time
- container/preflight_test.go

## God Nodes (most connected - your core abstractions)
1. `ctrCfg()` - 97 edges
2. `newTestOps()` - 89 edges
3. `newCapMgr()` - 88 edges
4. `Config` - 86 edges
5. `Role` - 85 edges
6. `Platform` - 75 edges
7. `bg()` - 69 edges
8. `NewCluster()` - 68 edges
9. `Commands` - 68 edges
10. `Manager` - 65 edges

## Surprising Connections (you probably didn't know these)
- `activity()` --calls--> `countContains()`  [INFERRED]
  internal/broker/verify_ops.go → internal/broker/broker.go
- `matchCLI()` --calls--> `cliArg()`  [INFERRED]
  internal/broker/broker_test.go → internal/broker/transport.go
- `seqTransport()` --calls--> `matchCLI()`  [INFERRED]
  internal/broker/verify_local_test.go → internal/broker/broker_test.go
- `TestLeaderLocalAssertLeaderError()` --calls--> `matchCLI()`  [INFERRED]
  internal/broker/verify_local_test.go → internal/broker/broker_test.go
- `TestLeaderLocalMateRPCErrorFails()` --calls--> `matchCLI()`  [INFERRED]
  internal/broker/verify_local_test.go → internal/broker/broker_test.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (93 total, 8 thin omitted)

### Community 0 - "bg"
Cohesion: 0.06
Nodes (106): TestCtrManagerConfirmWiring(), TestWarnAndStep(), App, lineSink(), progress(), step(), warn(), App (+98 more)

### Community 1 - "cli_test.go"
Cohesion: 0.11
Nodes (43): allowRuntime(), captureStderr(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone(), runStatusStderr(), TestAnnounceCommandsNamesResolvedBinaries() (+35 more)

### Community 2 - "Sink"
Cohesion: 0.17
Nodes (6): Level, Sink, pad(), TestLevelTagsMapToTheHouseTags(), KV, Row

### Community 3 - "context.Context"
Cohesion: 0.10
Nodes (10): Echo, Exec, context.Context, os/exec.Cmd, Quote(), Cluster, Cluster, Cluster (+2 more)

### Community 4 - "captureStdout"
Cohesion: 0.12
Nodes (24): opCall, opRunner, capture(), captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch() (+16 more)

### Community 5 - "testing.T"
Cohesion: 0.05
Nodes (93): testing.T, assertContainerBlockDefaults(), assertContainerScaling(), envTree(), Config, haNodesConfig(), TestApplyDefaultsDocker(), TestApplyDefaultsK8s() (+85 more)

### Community 6 - "Commands"
Cohesion: 0.03
Nodes (68): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util check (+60 more)

### Community 7 - "Manager"
Cohesion: 0.10
Nodes (10): os.FileMode, defaultGenPSK(), exactName(), Manager, orNone(), replacePSKLine(), secretSummary(), setOrMissing() (+2 more)

### Community 8 - "config.go"
Cohesion: 0.10
Nodes (25): Broker, Container, ContainerSecurity, DockerConfig, DomainCerts, Image, Network, Node (+17 more)

### Community 9 - "load.go"
Cohesion: 0.12
Nodes (19): scalingTier, TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config (+11 more)

### Community 10 - "validate.go"
Cohesion: 0.16
Nodes (16): keyValueEntries, checkCredentialChars(), foldToEnvVar(), Config, missingErr(), platformKey(), requireAll(), requireKeyValue() (+8 more)

### Community 11 - "NewTransport"
Cohesion: 0.13
Nodes (24): capCall, capRunner, New(), TestNewDefaults(), Transport, Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime() (+16 more)

### Community 12 - "NewCluster"
Cohesion: 0.11
Nodes (43): TestCheckStopsProbingWhenUnreachable(), TestCheckDeployGroupsAndOrdersSections(), TestCheckDeployNeverPrintsASecret(), TestCheckDeployReportsEveryFailureInOneRun(), TestCheckDeploySparseConfigExplainsItself(), TestCheckDeployTagsEchoedConfigAsINFO(), TestStorageRows(), NewCluster() (+35 more)

### Community 13 - "eqArgs"
Cohesion: 0.13
Nodes (27): TestReachable(), TestResolveStorageClass(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster() (+19 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (63): doc, Result, segment, vars, boolOf(), commentSafe(), Convert(), countMarkers() (+55 more)

### Community 15 - "internal/k8s"
Cohesion: 0.17
Nodes (12): check_test.go, cluster_test.go, deploy_test.go, internal/k8s, names_test.go, operator_test.go, ops_test.go, preflight_test.go (+4 more)

### Community 16 - "manager_test.go"
Cohesion: 0.09
Nodes (81): fileExists(), assertMode(), containsStr(), ctrCfg(), hasCall(), newCapMgr(), TestContainerRunningMatchesNameExactly(), TestManagerCheckDNSFailsLoudInHA() (+73 more)

### Community 17 - "cli/platform_test.go"
Cohesion: 0.24
Nodes (19): echoRunner(), App, App, runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused() (+11 more)

### Community 18 - "secrets_test.go"
Cohesion: 0.18
Nodes (19): GenSecrets(), TestUpdateServerCertSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), operatorRegcred(), checkGolden(), decodeDataValue() (+11 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (22): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+14 more)

### Community 20 - "prep_test.go"
Cohesion: 0.14
Nodes (34): saCfg(), TestRestartRolling(), adminCfg(), Cluster, labelCluster(), nsBrokerJSON(), nsBuiltinJSON(), nsCluster() (+26 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (18): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+10 more)

### Community 22 - "README.md"
Cohesion: 0.22
Nodes (6): internal/broker, internal/cli, internal/config, internal/engine, internal/k8s, internal/render

### Community 23 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.19
Nodes (34): curlCalls(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestLeaderLocalAssertLeaderError(), TestLeaderLocalBadRoleArg() (+26 more)

### Community 25 - "Role"
Cohesion: 0.07
Nodes (26): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+18 more)

### Community 26 - "Load"
Cohesion: 0.23
Nodes (16): minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadReadError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors() (+8 more)

### Community 27 - "Platform"
Cohesion: 0.05
Nodes (123): layer, opFunc, roleOpFunc, shorthand, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, github.com/spf13/pflag.FlagSet, io.Reader (+115 more)

### Community 28 - "completion_test.go"
Cohesion: 0.09
Nodes (44): Example, runComplete(), TestAllowCommandOffersNoFiles(), TestCompletionHelpStillWorks(), TestCompletionNeedsAShell(), TestCompletionNoDescriptions(), TestCompletionScriptsGenerate(), TestDirFlagCompletesDirectories() (+36 more)

### Community 31 - "Command"
Cohesion: 0.19
Nodes (14): Command, CheckCommand(), clusterRules(), composeRules(), Config, runtimeRules(), TestCharsetAgreesAcrossBothYAMLForms(), TestCheckCommandAccepts() (+6 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.25
Nodes (5): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh

### Community 34 - "execguard_test.go"
Cohesion: 0.21
Nodes (14): decodeStrict(), Config, guardCommandOf(), guardConfig(), setGuardCommand(), TestAllowCommandIsNotASchemaKey(), TestAllowCommandsAccepts(), TestAllowCommandsRejects() (+6 more)

### Community 35 - ".Run"
Cohesion: 0.22
Nodes (18): runRootWith(), TestCLICommand(), TestCtrRestartConfirmGate(), TestErrorPaths(), TestFirstArgOr(), TestK8sConfigDeleteDomainCertsConfigured(), TestK8sConfirmDeclined(), TestK8sPromptsNameNamespaceAndContext() (+10 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.16
Nodes (11): Runner, TestCanIAnswerReadsTheLastLine(), NewTransport(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs(), TestTransportUpload() (+3 more)

### Community 38 - "execguard.go"
Cohesion: 0.26
Nodes (7): commandRules, checkBinary(), checkFlagShape(), checkToken(), escalator(), execBase(), TestAllowCommandsRejectsEscalation()

### Community 39 - "Compose"
Cohesion: 0.12
Nodes (30): opCtrGenArtifact(), opCtrGenSecrets(), Config, NodeIdentity, Compose(), ContainerSecrets(), escapePercent(), ContainerSecret (+22 more)

### Community 40 - "runner_test.go"
Cohesion: 0.09
Nodes (33): maskedKeys(), TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveRefusesCurrentDirectory(), verboseExec() (+25 more)

### Community 41 - "haCfg"
Cohesion: 0.22
Nodes (13): haCfg(), TestOperatorImage(), TestConfirmNoDowngradeIsSilentInAPreview(), TestCreateSecretsFailsWithoutAdminFields(), TestDeleteNamespaceProtected(), TestDeleteSecretsSkipsUnconfiguredAdminSecret(), TestClusterHonoursRuntime(), TestExecutorRefusesUnapprovedRuntime() (+5 more)

### Community 42 - "Config"
Cohesion: 0.17
Nodes (18): Admin, Replication, Scaling, TLS, tlsConfigured(), Config, GenOperator(), GenOperatorSecrets() (+10 more)

### Community 43 - "Set"
Cohesion: 0.15
Nodes (13): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+5 more)

### Community 44 - "prep.go"
Cohesion: 0.10
Nodes (17): bufio.Reader, foreignObjects(), Cluster, isBuiltinLabel(), isOurs(), joinManifests(), namespaceManifest(), namespaceOursSecrets() (+9 more)

### Community 45 - "writeCtrStandaloneEnv"
Cohesion: 0.25
Nodes (11): runCtr(), TestConfigStepsDoNotLeakSecrets(), TestCtrConfigDryRun(), TestCtrConfirmDeclined(), TestCtrDiagnosticsDryRun(), TestCtrErrorPaths(), TestCtrExecCLIPathSeparator(), TestCtrRoleArgCount() (+3 more)

### Community 46 - "newTestOps"
Cohesion: 0.05
Nodes (88): appUsers(), Ops, matchCLI(), newTestOps(), ranContains(), TestAdditionalUsers(), TestAdditionalUsersEmpty(), TestAdditionalUsersRejectsBadValues() (+80 more)

### Community 47 - "command_test.go"
Cohesion: 0.16
Nodes (13): decodeRuntime(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRuntimeDefaults() (+5 more)

### Community 50 - "Probe"
Cohesion: 0.20
Nodes (17): TestRunProbesSkipsOffPlatformAndExitsOnRefuted(), probeP1(), p2Classify(), probeP2(), probeP4(), Env, inconclusive(), listProbes() (+9 more)

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (25): applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestDangerousVerbsHaveNoBareAlias(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownNoun(), TestStartStopHaveNoAlias(), TestAllowCommandIsRegisteredWhereItExecutes() (+17 more)

### Community 52 - "podmanEnv"
Cohesion: 0.16
Nodes (22): Env, podmanEnv(), TestCtrNameReadsTheConfiguredContainer(), TestP1DeclinedConsentTouchesNothing(), TestP1DryRunTouchesNothing(), TestP1RestoresAfterAFailureMidSequence(), TestP1RestoreWarnsLoudlyWhenItCannotRestore(), TestP1RunsStopRotateStartReadInOrder() (+14 more)

### Community 53 - "newEchoMgr"
Cohesion: 0.10
Nodes (21): NewManager(), Manager, newEchoMgr(), rootlessNoFileMgr(), setNoFile(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployDockerDryRunMasksSecretEnv() (+13 more)

### Community 54 - "Operations"
Cohesion: 0.20
Nodes (10): Bringing up a fresh cluster, Docker and Podman mechanics, Extra CLI users differ by platform, Operations, Post-deployment configuration order, Re-deploying is safe and explicit, Removing a broker: what stays, what goes, Rendering without applying (+2 more)

### Community 55 - "Test catalogue"
Cohesion: 0.12
Nodes (17): abbrev_test.go, convert_test.go, Coverage, examples_test.go, examples_test.go, internal/abbrev, internal/convert, internal/examples (+9 more)

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
Cohesion: 0.14
Nodes (20): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), failRow(), Cluster, info(), okRow() (+12 more)

### Community 65 - "Configuration"
Cohesion: 0.22
Nodes (9): Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Scaling, Scaling tiers, Secrets, The command fields are executable content, The keys (+1 more)

### Community 66 - "runP1"
Cohesion: 0.33
Nodes (10): ctrName(), Env, Env, Result, p1Classify(), p1ReadMountedSecret(), p1Restore(), p1Throwaway() (+2 more)

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 71 - "semp_test.go"
Cohesion: 0.14
Nodes (25): sempPort(), assertNoPasswordInArgv(), TestMateRevertActivityBridgePlaintextOnly(), TestMateRevertActivityBridgeTLSNoCA(), TestMateRevertActivityBridgeTLSWithCA(), TestMateRevertActivityCredsAndBodyNeverInArgv(), TestMateRevertActivityNon2xx(), TestMateRevertActivityRPCNotOK() (+17 more)

### Community 72 - "Live-environment probe harness (`solace-itest`)"
Cohesion: 0.20
Nodes (10): After a run, Build and run, Flags, Live-environment probe harness (`solace-itest`), Operator-run only, Safety, Testing the harness itself, The probes (+2 more)

### Community 73 - "itest_test.go"
Cohesion: 0.14
Nodes (17): bytes.Buffer, Env, ids(), sinkOver(), TestConsent(), testEnv(), TestListProbesNeedsNoEnvFile(), TestP1Classify() (+9 more)

### Community 74 - "scripts.go"
Cohesion: 0.06
Nodes (50): showCmd, containsAnyFold(), countContains(), TestContainsAnyFold(), TestCountContains(), TestPathHelpers(), TestValidName(), validCLILine() (+42 more)

### Community 77 - "resolveEnv"
Cohesion: 0.13
Nodes (20): containsPlatform(), Env, platformNames(), resolveEnv(), resolvePlatform(), roleLabel(), TestResolvePlatform(), writeEnv() (+12 more)

### Community 78 - "New"
Cohesion: 0.16
Nodes (19): TestCheckReportSkipsEmptySections(), New(), NewFunc(), sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow() (+11 more)

### Community 79 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 80 - "internal/tools/itest"
Cohesion: 0.50
Nodes (4): internal/tools/itest, itest_test.go, live_test.go, probes_live_test.go

### Community 81 - "runP4"
Cohesion: 0.21
Nodes (14): TestCPUAndMemoryConversion(), TestP4Classify(), TestP4Parse(), TestRunP4InconclusiveWhenInspectFails(), TestRunP4ReadsInspectAndChangesNothing(), TestRunP4SkipsWithNoLimitsConfigured(), cpuToNano(), Env (+6 more)

### Community 82 - "age"
Cohesion: 0.22
Nodes (9): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+1 more)

### Community 83 - "inspect_test.go"
Cohesion: 0.18
Nodes (17): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+9 more)

### Community 86 - "inspect.go"
Cohesion: 0.29
Nodes (13): brokerConditionRows(), conditionLevel(), findCondition(), TestConditionLevelDegradesSafely(), brokerList, brokerStatus, condition, deploymentList (+5 more)

### Community 89 - "render.go"
Cohesion: 0.13
Nodes (31): strings.Builder, PodAffinityTerm, Cluster, boolStr(), BrokerCR(), containerSecretSpecs(), cut(), EnvPairs() (+23 more)

### Community 91 - ".releaseToBackup"
Cohesion: 0.12
Nodes (19): field(), TestField(), TestLastLines(), TestPrimaryRedundancyUp(), TestFieldLabelWithoutColon(), TestLastLinesEqualCount(), Ops, hostMatches() (+11 more)

### Community 93 - "Troubleshooting"
Cohesion: 0.33
Nodes (6): A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, File descriptors on rootless podman, Podman secret flags, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 100 - "internal/engine"
Cohesion: 0.67
Nodes (3): internal/engine, resolve_test.go, runner_test.go

### Community 102 - "allowcommand_test.go"
Cohesion: 0.33
Nodes (9): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused(), TestSmuggledSubcommandIsRefused() (+1 more)

### Community 103 - ".MateRevertActivity"
Cohesion: 0.15
Nodes (18): mateTarget, TestHTTPStatusHelpers(), anyHTTP2xx(), bridgeHostPort(), curlConfigFlag(), curlConfigLine(), Ops, httpBody() (+10 more)

### Community 104 - "time.Time"
Cohesion: 0.33
Nodes (4): time.Time, Cluster, normalizeToList(), TestNormalizeToListHandlesBothKubectlShapes()

### Community 106 - "container/preflight_test.go"
Cohesion: 0.29
Nodes (6): TestComposeSecretEnvNamesCannotBeSystemVars(), TestPreflightFailureStopsLifecycle(), TestPreflightFailureStopsTheDeploy(), TestPreflightHintIsPlatformShaped(), TestPreflightIsPreviewableUnderDryRun(), TestPreflightRunsBeforeAnything()

## Knowledge Gaps
- **175 isolated node(s):** `solace`, `showCmd`, `Config`, `operatorTmplVars`, `NO_COLOR` (+170 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `bg`, `captureStdout`, `testing.T`, `Manager`, `config.go`, `load.go`, `validate.go`, `NewTransport`, `convert_test.go`, `manager_test.go`, `Load`, `completion_test.go`, `Command`, `execguard_test.go`, `execguard.go`, `Compose`, `platformTitle`, `Probe`, `newEchoMgr`, `containerTransport`, `semp_test.go`, `resolveEnv`, `Ops`?**
  _High betweenness centrality (0.074) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `bg`, `cli_test.go`, `captureStdout`, `Manager`, `config.go`, `NewTransport`, `NewCluster`, `convert_test.go`, `manager_test.go`, `secrets_test.go`, `prep_test.go`, `Role`, `recRunner`, `Compose`, `haCfg`, `prep.go`, `newTestOps`, `Probe`, `newEchoMgr`, `containerTransport`, `semp_test.go`, `Ops`, `kubectlTransport`, `render.go`, `Cluster`?**
  _High betweenness centrality (0.064) - this node is a cross-community bridge._
- **Why does `Manager` connect `Manager` to `bg`, `recRunner`, `Config`, `Probe`, `newEchoMgr`, `Platform`?**
  _High betweenness centrality (0.029) - this node is a cross-community bridge._
- **Are the 10 inferred relationships involving `ctrCfg()` (e.g. with `TestComposeSecretEnvIsTheOnlyChildEnvironment()` and `TestComposeSecretEnvNamesCannotBeSystemVars()`) actually correct?**
  _`ctrCfg()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **Are the 50 inferred relationships involving `newTestOps()` (e.g. with `TestAdditionalUsersRunCLITransportError()` and `TestDiagnosticsMkdirError()`) actually correct?**
  _`newTestOps()` has 50 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `newCapMgr()` (e.g. with `NewManager()` and `TestComposeSecretEnvIsTheOnlyChildEnvironment()`) actually correct?**
  _`newCapMgr()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `showCmd`, `Config` to the rest of the system?**
  _175 weakly-connected nodes found - possible documentation gaps or missing edges._
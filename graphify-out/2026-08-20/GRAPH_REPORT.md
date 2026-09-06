# Graph Report - solace-k8-scripts  (2026-08-20)

## Corpus Check
- 92 files · ~210,629 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1827 nodes · 5749 edges · 67 communities (62 shown, 5 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 915 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `72a3a798`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- bg
- cli_test.go
- SecretScript
- Cluster
- captureStdout
- testing.T
- Commands
- Manager
- Config
- load.go
- newTestOps
- Ops
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
- CLAUDE.md
- judge
- verify_local_test.go
- Role
- newEchoMgr
- Platform
- render.go
- solace
- Command
- Command reference
- context.Context
- completion_test.go
- runRootWith
- .resolveSecretRefs
- haCfg
- kubectlTransport
- Compose
- scripts.go
- k8s/runtime_test.go
- semp_test.go
- .Run
- Cluster
- io.Writer
- coverage_test.go
- runner_test.go
- .Run
- manager.go
- .runtime
- newRootCmd
- quadletEscape
- command_test.go
- HARoles
- Test catalogue
- internal/cli
- internal/broker
- internal/config
- internal/container
- Fixtures and doubles
- internal/engine
- RenderOperator
- containerTransport
- .AdditionalUsers
- .Preflight

## God Nodes (most connected - your core abstractions)
1. `ctrCfg()` - 91 edges
2. `newTestOps()` - 82 edges
3. `newCapMgr()` - 82 edges
4. `Role` - 79 edges
5. `Config` - 78 edges
6. `Platform` - 69 edges
7. `bg()` - 68 edges
8. `Commands` - 65 edges
9. `Manager` - 58 edges
10. `NewCluster()` - 57 edges

## Surprising Connections (you probably didn't know these)
- `activity()` --calls--> `countContains()`  [INFERRED]
  internal/broker/verify_ops.go → internal/broker/broker.go
- `matchCLI()` --calls--> `cliArg()`  [INFERRED]
  internal/broker/broker_test.go → internal/broker/transport.go
- `TestAdditionalUsersRunCLITransportError()` --calls--> `matchCLI()`  [INFERRED]
  internal/broker/coverage_test.go → internal/broker/broker_test.go
- `TestDisableDefaultUsersShowVPNError()` --calls--> `matchCLI()`  [INFERRED]
  internal/broker/coverage_test.go → internal/broker/broker_test.go
- `TestReleaseToBackupReleasedTimeout()` --calls--> `matchCLI()`  [INFERRED]
  internal/broker/coverage_test.go → internal/broker/broker_test.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (67 total, 5 thin omitted)

### Community 0 - "bg"
Cohesion: 0.07
Nodes (97): TestConfirmFlagShortcuts(), TestConfirmNonTTY(), TestCtrManagerConfirmWiring(), warn(), App, newConvertCmd(), runConvert(), confirmDelete() (+89 more)

### Community 1 - "cli_test.go"
Cohesion: 0.10
Nodes (44): os.File, allowRuntime(), capture(), captureStderr(), fakeBinaryOnPath(), firstLine(), runRoot(), runStandalone() (+36 more)

### Community 2 - "SecretScript"
Cohesion: 0.22
Nodes (6): opCtrGenSecrets(), ContainerSecret, secretFilePath(), SecretScript(), shQuote(), TestShQuote()

### Community 3 - "Cluster"
Cohesion: 0.18
Nodes (5): podName(), filterLines(), Cluster, TestFilterLines(), surveySection

### Community 4 - "captureStdout"
Cohesion: 0.13
Nodes (21): opCall, opRunner, captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch(), opCanI() (+13 more)

### Community 5 - "testing.T"
Cohesion: 0.07
Nodes (79): testing.T, assertContainerBlockDefaults(), assertContainerScaling(), envTree(), Config, haNodesConfig(), minimalK8s(), TestApplyDefaultsDocker() (+71 more)

### Community 6 - "Commands"
Cohesion: 0.03
Nodes (65): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util check (+57 more)

### Community 7 - "Manager"
Cohesion: 0.18
Nodes (3): os.FileMode, fileExists(), Manager

### Community 8 - "Config"
Cohesion: 0.10
Nodes (28): Admin, Broker, Container, ContainerSecurity, DockerConfig, DomainCerts, Image, Network (+20 more)

### Community 9 - "load.go"
Cohesion: 0.12
Nodes (19): scalingTier, TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config (+11 more)

### Community 10 - "newTestOps"
Cohesion: 0.11
Nodes (40): appUsers(), Ops, newTestOps(), TestAdditionalUsers(), TestAdditionalUsersEmpty(), TestAdditionalUsersRejectsBadValues(), TestAdditionalUsersReportsExistingUser(), TestDiagnostics() (+32 more)

### Community 11 - "Ops"
Cohesion: 0.10
Nodes (26): capCall, capRunner, time.Duration, Ops, New(), TestNewDefaults(), Transport, Manager (+18 more)

### Community 12 - "NewCluster"
Cohesion: 0.12
Nodes (37): TestCheckAbortsWhenUnreachable(), TestCheckDryRun(), TestCheckEnvNoSecretLeak(), TestCheckEnvSparseConfig(), TestCheckOperatorNS(), TestCheckStorageClass(), TestResolveStorageClass(), NewCluster() (+29 more)

### Community 13 - "eqArgs"
Cohesion: 0.14
Nodes (29): TestReachable(), Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSDefaultOnError(), TestOperatorNSDefaultWhenAbsent(), TestOperatorNSDerived() (+21 more)

### Community 14 - "convert_test.go"
Cohesion: 0.07
Nodes (55): doc, Result, vars, boolOf(), commentSafe(), Convert(), countMarkers(), emitYAML() (+47 more)

### Community 15 - "internal/k8s"
Cohesion: 0.17
Nodes (12): check_test.go, cluster_test.go, deploy_test.go, internal/k8s, names_test.go, operator_test.go, ops_test.go, preflight_test.go (+4 more)

### Community 16 - "manager_test.go"
Cohesion: 0.07
Nodes (88): assertMode(), containsStr(), ctrCfg(), hasCall(), maskedKeys(), newCapMgr(), setNoFile(), TestContainerRunningMatchesNameExactly() (+80 more)

### Community 17 - "cli/platform_test.go"
Cohesion: 0.27
Nodes (17): echoRunner(), App, runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations() (+9 more)

### Community 18 - "secrets_test.go"
Cohesion: 0.16
Nodes (20): GenSecrets(), joinManifests(), TestUpdateServerCertSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), operatorRegcred(), checkGolden() (+12 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "prep_test.go"
Cohesion: 0.15
Nodes (25): saCfg(), adminCfg(), Cluster, labelCluster(), TestCreateNamespaceApplyFails(), TestCreateSecretsAdminOnly(), TestCreateSecretsAllThree(), TestCreateSecretsFailsWithoutAdminFields() (+17 more)

### Community 21 - "dev.ps1"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "CLAUDE.md"
Cohesion: 0.12
Nodes (11): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh, internal/broker, internal/cli, internal/config (+3 more)

### Community 23 - "judge"
Cohesion: 0.16
Nodes (21): describe(), judge(), main(), plural(), run(), load(), TestJudge(), TestJudgeEmptyTrace() (+13 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.19
Nodes (34): curlCalls(), Ops, isCurl(), newLocalOps(), rd(), seqTransport(), TestLeaderLocalAssertLeaderError(), TestLeaderLocalBadRoleArg() (+26 more)

### Community 25 - "Role"
Cohesion: 0.13
Nodes (13): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+5 more)

### Community 26 - "newEchoMgr"
Cohesion: 0.11
Nodes (19): bytes.Buffer, NewManager(), Manager, newEchoMgr(), rootlessNoFileMgr(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployDockerDryRunMasksSecretEnv() (+11 more)

### Community 27 - "Platform"
Cohesion: 0.07
Nodes (93): layer, opFunc, roleOpFunc, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, github.com/spf13/pflag.FlagSet, TestFirstArg(), availableSubs() (+85 more)

### Community 28 - "render.go"
Cohesion: 0.20
Nodes (23): strings.Builder, PodAffinityTerm, boolStr(), BrokerCR(), containerSecretSpecs(), cut(), parsePort(), parseToleration() (+15 more)

### Community 31 - "Command"
Cohesion: 0.06
Nodes (45): commandRules, keyValueEntries, TestCommandArgsDoesNotAliasCommand(), Command, checkBinary(), CheckCommand(), checkFlagShape(), checkToken() (+37 more)

### Community 32 - "Command reference"
Cohesion: 0.50
Nodes (3): Command reference, Global flags, Tree

### Community 33 - "context.Context"
Cohesion: 0.16
Nodes (4): context.Context, Cluster, Cluster, Cluster

### Community 34 - "completion_test.go"
Cohesion: 0.37
Nodes (12): runComplete(), TestAllowCommandOffersNoFiles(), TestDirFlagCompletesDirectories(), TestEnvFlagCompletesEnvFiles(), TestEnvFlagPrefixFilters(), TestEnvFlagWithPathDefersToShell(), TestNoArgsLeafOffersNoFiles(), TestPlatformFlagCompletes() (+4 more)

### Community 35 - "runRootWith"
Cohesion: 0.17
Nodes (21): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused() (+13 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "haCfg"
Cohesion: 0.16
Nodes (12): haCfg(), TestShowAllReportsAndContinuesOnGetError(), TestCanIAnswerReadsTheLastLine(), NewTransport(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs() (+4 more)

### Community 39 - "Compose"
Cohesion: 0.14
Nodes (30): opCtrGenArtifact(), HealthCheck, NodeIdentity, Compose(), ContainerSecrets(), EnvPairs(), groupKey(), healthCmd() (+22 more)

### Community 40 - "scripts.go"
Cohesion: 0.05
Nodes (53): showCmd, field(), TestField(), TestHTTPStatusHelpers(), TestLastLines(), TestPrimaryRedundancyUp(), concatFiles(), Ops (+45 more)

### Community 41 - "k8s/runtime_test.go"
Cohesion: 0.43
Nodes (7): TestClusterHonoursRuntime(), TestExecutorRefusesUnapprovedRuntime(), TestRuntimeDefaultArgvUnchanged(), TestTransportHonoursRuntime(), unapprovedCfg(), withLeading(), wrappedCfg()

### Community 42 - "semp_test.go"
Cohesion: 0.11
Nodes (28): anyHTTP2xx(), curlConfigLine(), Ops, httpBody(), lastHTTPStatus(), revertActivityMateBody(), sempPort(), sempV1OK() (+20 more)

### Community 43 - ".Run"
Cohesion: 0.14
Nodes (15): Echo, Exec, os/exec.Cmd, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand(), TestResolveMissingBinaryIsActionable() (+7 more)

### Community 44 - "Cluster"
Cohesion: 0.12
Nodes (11): bufio.Reader, Cluster, isBuiltinLabel(), namespaceManifest(), roleName(), rolePlacementLabels(), splitLabel(), TestIsBuiltinLabel() (+3 more)

### Community 45 - "io.Writer"
Cohesion: 0.24
Nodes (10): io.Reader, io.Writer, TestPromptYes(), TestPromptYesNo(), App, step(), promptLine(), promptYes() (+2 more)

### Community 46 - "coverage_test.go"
Cohesion: 0.06
Nodes (48): matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle(), TestDisableDefaultUsersDisableError(), TestDisableDefaultVPNDisableError() (+40 more)

### Community 47 - "runner_test.go"
Cohesion: 0.14
Nodes (22): captureStdout(), helperCommand(), TestEchoDefaultWriter(), TestEchoOutput(), TestEchoOutputInput(), TestEchoRun(), TestEchoRunEnv(), TestEchoRunEnvNoEnv() (+14 more)

### Community 48 - ".Run"
Cohesion: 0.16
Nodes (17): runCtr(), TestConfigStepsDoNotLeakSecrets(), TestCtrConfigDryRun(), TestCtrConfirmDeclined(), TestCtrDiagnosticsDryRun(), TestCtrErrorPaths(), TestCtrExecCLIPathSeparator(), TestCtrRoleArgCount() (+9 more)

### Community 49 - "manager.go"
Cohesion: 0.15
Nodes (9): defaultGenPSK(), exactName(), orNone(), platformTitle(), replacePSKLine(), secretSummary(), setOrMissing(), splitLimit() (+1 more)

### Community 51 - "newRootCmd"
Cohesion: 0.09
Nodes (25): applyAliases(), TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestDangerousVerbsHaveNoBareAlias(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownNoun(), TestStartStopHaveNoAlias(), TestAllowCommandIsRegisteredWhereItExecutes() (+17 more)

### Community 52 - "quadletEscape"
Cohesion: 0.67
Nodes (3): escapePercent(), quadletEscape(), TestQuadletEscape()

### Community 53 - "command_test.go"
Cohesion: 0.18
Nodes (12): decodeRuntime(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRuntimeDefaults(), TestRuntimeExplicitValueSurvivesDefaults() (+4 more)

### Community 54 - "HARoles"
Cohesion: 0.16
Nodes (11): Cluster, HARoles(), lbServiceName(), ProductKeyRoles(), pvcName(), RestartOrder(), stsName(), TestHARoles() (+3 more)

### Community 55 - "Test catalogue"
Cohesion: 0.18
Nodes (10): convert_test.go, Coverage, internal/convert, internal/render, internal/tools/vulnjudge, main_test.go, render_test.go, Running the tests (+2 more)

### Community 57 - "internal/cli"
Cohesion: 0.29
Nodes (7): aliases_test.go, allowcommand_test.go, cli_test.go, commanddoc_test.go, completion_test.go, internal/cli, platform_test.go

### Community 58 - "internal/broker"
Cohesion: 0.33
Nodes (6): broker_test.go, coverage_test.go, internal/broker, scripts_test.go, semp_test.go, verify_local_test.go

### Community 59 - "internal/config"
Cohesion: 0.33
Nodes (6): command_test.go, config_test.go, execguard_test.go, internal/config, platform_test.go, scaling_test.go

### Community 60 - "internal/container"
Cohesion: 0.40
Nodes (5): internal/container, manager_test.go, preflight_test.go, runtime_test.go, transport_test.go

### Community 61 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 63 - "internal/engine"
Cohesion: 0.67
Nodes (3): internal/engine, resolve_test.go, runner_test.go

### Community 65 - "RenderOperator"
Cohesion: 0.18
Nodes (13): orNone(), orValue(), setOrMissing(), setOrNone(), GenOperator(), joinYAMLDocs(), operatorImage(), RenderOperator() (+5 more)

### Community 68 - ".AdditionalUsers"
Cohesion: 0.13
Nodes (16): containsAnyFold(), countContains(), TestContainsAnyFold(), TestCountContains(), TestPathHelpers(), TestValidName(), validCLILine(), validCLIPassword() (+8 more)

## Knowledge Gaps
- **125 isolated node(s):** `solace`, `showCmd`, `Config`, `operatorTmplVars`, `Cluster` (+120 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **5 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Config` connect `Config` to `bg`, `cli_test.go`, `SecretScript`, `Cluster`, `captureStdout`, `Manager`, `newTestOps`, `Ops`, `NewCluster`, `convert_test.go`, `manager_test.go`, `secrets_test.go`, `prep_test.go`, `newEchoMgr`, `render.go`, `context.Context`, `haCfg`, `kubectlTransport`, `Compose`, `k8s/runtime_test.go`, `semp_test.go`, `Cluster`, `io.Writer`, `HARoles`, `RenderOperator`, `containerTransport`?**
  _High betweenness centrality (0.093) - this node is a cross-community bridge._
- **Why does `Platform` connect `Platform` to `SecretScript`, `containerTransport`, `captureStdout`, `testing.T`, `Manager`, `Config`, `load.go`, `semp_test.go`, `Ops`, `Compose`, `io.Writer`, `convert_test.go`, `manager_test.go`, `manager.go`, `newEchoMgr`, `Command`?**
  _High betweenness centrality (0.065) - this node is a cross-community bridge._
- **Why does `Role` connect `Role` to `bg`, `containerTransport`, `.AdditionalUsers`, `Cluster`, `kubectlTransport`, `Manager`, `scripts.go`, `Compose`, `Ops`, `Cluster`, `HARoles`, `Platform`?**
  _High betweenness centrality (0.033) - this node is a cross-community bridge._
- **Are the 10 inferred relationships involving `ctrCfg()` (e.g. with `TestComposeSecretEnvIsTheOnlyChildEnvironment()` and `TestComposeSecretEnvNamesCannotBeSystemVars()`) actually correct?**
  _`ctrCfg()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **Are the 43 inferred relationships involving `newTestOps()` (e.g. with `TestAdditionalUsersRunCLITransportError()` and `TestDiagnosticsMkdirError()`) actually correct?**
  _`newTestOps()` has 43 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `newCapMgr()` (e.g. with `NewManager()` and `TestComposeSecretEnvIsTheOnlyChildEnvironment()`) actually correct?**
  _`newCapMgr()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `showCmd`, `Config` to the rest of the system?**
  _125 weakly-connected nodes found - possible documentation gaps or missing edges._
# Graph Report - solace-k8-scripts  (2026-09-04)

## Corpus Check
- 133 files · ~282,316 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2328 nodes · 7177 edges · 101 communities (93 shown, 8 thin omitted)
- Extraction: 84% EXTRACTED · 16% INFERRED · 0% AMBIGUOUS · INFERRED: 1152 edges (avg confidence: 0.85)
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
- config_test.go
- Commands
- Manager
- Config
- load.go
- Command
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
- io.Writer
- Platform
- DetectPlatforms
- solace
- CheckCommand
- CLAUDE.md
- .Preflight
- completion_test.go
- runRootWith
- .resolveSecretRefs
- recRunner
- render.go
- Compose
- runRoot
- haCfg
- semp_test.go
- Set
- prep.go
- operator.go
- newTestOps
- Cluster
- runP2
- manager.go
- Probe
- newRootCmd
- podmanEnv
- newEchoMgr
- Operations
- Test catalogue
- scripts_test.go
- internal/cli
- internal/broker
- internal/config
- internal/container
- .configRows
- scaling_test.go
- NewManager
- Configuration
- runP1
- containerTransport
- command_test.go
- Developer guide
- validate.go
- Load
- Live-environment probe harness (`solace-itest`)
- itest_test.go
- fileExists
- NodeIdentity
- compareVersions
- Consenter
- output_test.go
- Fixtures and doubles
- internal/tools/itest
- runP4
- age
- inspect_test.go
- fakeRunner
- testing.T
- inspect.go
- kubectlTransport
- run
- ContainerSecret
- Ops
- scripts.go
- container/preflight_test.go
- Troubleshooting
- ResolveEnvPath
- Image
- Abbreviations
- Command reference
- internal/engine

## God Nodes (most connected - your core abstractions)
1. `ctrCfg()` - 97 edges
2. `newCapMgr()` - 88 edges
3. `Config` - 86 edges
4. `newTestOps()` - 83 edges
5. `Role` - 82 edges
6. `Platform` - 75 edges
7. `bg()` - 69 edges
8. `NewCluster()` - 68 edges
9. `Commands` - 68 edges
10. `Manager` - 65 edges

## Surprising Connections (you probably didn't know these)
- `TestFieldLabelWithoutColon()` --calls--> `field()`  [INFERRED]
  internal/broker/coverage_test.go → internal/broker/broker.go
- `activity()` --calls--> `countContains()`  [INFERRED]
  internal/broker/verify_ops.go → internal/broker/broker.go
- `matchCLI()` --calls--> `cliArg()`  [INFERRED]
  internal/broker/broker_test.go → internal/broker/transport.go
- `seqTransport()` --calls--> `matchCLI()`  [INFERRED]
  internal/broker/verify_local_test.go → internal/broker/broker_test.go
- `TestLeaderLocalAssertLeaderError()` --calls--> `matchCLI()`  [INFERRED]
  internal/broker/verify_local_test.go → internal/broker/broker_test.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (101 total, 8 thin omitted)

### Community 0 - "bg"
Cohesion: 0.06
Nodes (100): App, lineSink(), progress(), step(), warn(), App, newConvertCmd(), runConvert() (+92 more)

### Community 1 - "cli_test.go"
Cohesion: 0.09
Nodes (55): allowRuntime(), captureStderr(), echoRunner(), fakeBinaryOnPath(), firstLine(), App, runCtr(), runStandalone() (+47 more)

### Community 2 - "Sink"
Cohesion: 0.12
Nodes (11): TestCheckReportSkipsEmptySections(), Level, Sink, New(), NewFunc(), pad(), TestLevelTagsMapToTheHouseTags(), TestNewFuncRoutesThroughTheGivenLineSink() (+3 more)

### Community 3 - "context.Context"
Cohesion: 0.12
Nodes (7): context.Context, Cluster, storageClassSuitable(), Cluster, Cluster, badJSONRunner, fixtureRunner

### Community 4 - "captureStdout"
Cohesion: 0.13
Nodes (21): opCall, opRunner, capture(), captureStdout(), failDisableDefaultUsersUpload(), k8sDeployAllOutputHook(), loadDirect(), opArgvMatch() (+13 more)

### Community 5 - "config_test.go"
Cohesion: 0.07
Nodes (45): assertContainerBlockDefaults(), assertContainerScaling(), Config, haNodesConfig(), TestApplyDefaultsDocker(), TestApplyDefaultsK8s(), TestApplyDefaultsK8sTLS(), TestApplyDefaultsPodmanRootful() (+37 more)

### Community 6 - "Commands"
Cohesion: 0.03
Nodes (68): Commands, solace-util, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util check (+60 more)

### Community 8 - "Config"
Cohesion: 0.10
Nodes (29): Admin, Broker, Container, ContainerSecurity, DockerConfig, DomainCerts, Network, Node (+21 more)

### Community 9 - "load.go"
Cohesion: 0.24
Nodes (11): TestApplyBridgePortDefaults(), TestDefaultK8sPortsMatchesOperator(), applyBridgePortDefaults(), applyContainerBlockDefaults(), defaultContainerPorts(), defaultK8sPorts(), Config, setDefault() (+3 more)

### Community 10 - "Command"
Cohesion: 0.12
Nodes (17): Command, decodeStrict(), Config, guardCommandOf(), guardConfig(), setGuardCommand(), TestAllowCommandIsNotASchemaKey(), TestAllowCommandsAccepts() (+9 more)

### Community 11 - "NewTransport"
Cohesion: 0.18
Nodes (15): capCall, capRunner, New(), TestNewDefaults(), Transport, NewTransport(), dockerCfg(), eqArgs() (+7 more)

### Community 12 - "NewCluster"
Cohesion: 0.11
Nodes (42): TestCheckStopsProbingWhenUnreachable(), TestCheckDeployGroupsAndOrdersSections(), TestCheckDeployNeverPrintsASecret(), TestCheckDeployReportsEveryFailureInOneRun(), TestCheckDeploySparseConfigExplainsItself(), TestCheckDeployTagsEchoedConfigAsINFO(), NewCluster(), logBuf() (+34 more)

### Community 13 - "eqArgs"
Cohesion: 0.15
Nodes (25): Cluster, newCluster(), TestApplyOnStdin(), TestDeleteStdin(), TestOperatorNSExplicit(), TestOperatorNSNeverProbesTheCluster(), TestCLIAndShellAreInteractive(), TestCopyFrom() (+17 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (58): doc, Result, vars, boolOf(), commentSafe(), Convert(), countMarkers(), emitYAML() (+50 more)

### Community 15 - "internal/k8s"
Cohesion: 0.17
Nodes (12): check_test.go, cluster_test.go, deploy_test.go, internal/k8s, names_test.go, operator_test.go, ops_test.go, preflight_test.go (+4 more)

### Community 16 - "manager_test.go"
Cohesion: 0.10
Nodes (72): defaultGenPSK(), containsStr(), ctrCfg(), hasCall(), newCapMgr(), TestContainerRunningMatchesNameExactly(), TestDefaultGenPSK(), TestManagerCheckDNSFailsLoudInHA() (+64 more)

### Community 17 - "cli/platform_test.go"
Cohesion: 0.28
Nodes (16): App, runPlatform(), TestMultiPlatformNonInteractiveIsRefused(), TestMultiPlatformPromptRejectsBadAnswer(), TestMultiPlatformPromptSelects(), TestNoPlatformSectionIsRefused(), TestPlatformFlagAcceptsAbbreviations(), TestPlatformFlagRejectsUndeclaredSection() (+8 more)

### Community 18 - "secrets_test.go"
Cohesion: 0.18
Nodes (18): GenSecrets(), TestUpdateServerCertSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), checkGolden(), decodeDataValue(), TestAdminSecretDecodes() (+10 more)

### Community 19 - "dev.sh"
Cohesion: 0.18
Nodes (22): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+14 more)

### Community 20 - "prep_test.go"
Cohesion: 0.15
Nodes (33): saCfg(), TestRestartRolling(), adminCfg(), Cluster, labelCluster(), nsBrokerJSON(), nsBuiltinJSON(), nsCluster() (+25 more)

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
Cohesion: 0.15
Nodes (12): downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile, runErrMatchTransport (+4 more)

### Community 26 - "io.Writer"
Cohesion: 0.14
Nodes (16): Echo, Exec, io.Writer, os/exec.Cmd, TestChildEnvNamesAreNotSystemVariables(), TestExecEchoesOnEveryMethod(), TestExecIsSilentWithoutVerbose(), TestExecVerboseAnnouncesEveryCommand() (+8 more)

### Community 27 - "Platform"
Cohesion: 0.06
Nodes (115): layer, opFunc, roleOpFunc, shorthand, github.com/spf13/cobra.Command, github.com/spf13/cobra.ShellCompDirective, github.com/spf13/pflag.FlagSet, io.Reader (+107 more)

### Community 28 - "DetectPlatforms"
Cohesion: 0.14
Nodes (27): Example, listExamples(), newExamplesCmd(), runExample(), parseError(), DetectPlatforms(), TestDetectPlatforms(), TestDetectPlatformsBashFileHint() (+19 more)

### Community 31 - "CheckCommand"
Cohesion: 0.18
Nodes (15): commandRules, checkBinary(), CheckCommand(), checkFlagShape(), checkToken(), clusterRules(), composeRules(), execBase() (+7 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.25
Nodes (5): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh

### Community 34 - "completion_test.go"
Cohesion: 0.35
Nodes (13): runComplete(), TestAllowCommandOffersNoFiles(), TestDirFlagCompletesDirectories(), TestEnvFlagCompletesEnvFiles(), TestEnvFlagPrefixFilters(), TestEnvFlagWithPathDefersToShell(), TestExampleArgsComplete(), TestNoArgsLeafOffersNoFiles() (+5 more)

### Community 35 - "runRootWith"
Cohesion: 0.21
Nodes (17): TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues(), TestEscalationIsRefusedEndToEnd(), TestGenPathNeverExecutes(), TestHostileRuntimeIsRefusedByEveryVerb(), TestPathRuntimeIsRefused(), TestSmuggledSubcommandIsRefused() (+9 more)

### Community 36 - ".resolveSecretRefs"
Cohesion: 0.47
Nodes (3): secretRef, Config, unsetOrEmpty()

### Community 37 - "recRunner"
Cohesion: 0.14
Nodes (11): TestCanIAnswerReadsTheLastLine(), roleName(), NewTransport(), isCanI(), TestTransportCopy(), TestTransportEchoHidesUploadBody(), TestTransportExecArgs(), TestTransportUpload() (+3 more)

### Community 38 - "render.go"
Cohesion: 0.14
Nodes (29): strings.Builder, boolStr(), BrokerCR(), containerSecretSpecs(), cut(), EnvPairs(), groupKey(), itoa() (+21 more)

### Community 39 - "Compose"
Cohesion: 0.18
Nodes (25): opCtrGenArtifact(), HealthCheck, Compose(), ContainerSecrets(), escapePercent(), healthCmd(), Quadlet(), quadletEscape() (+17 more)

### Community 40 - "runRoot"
Cohesion: 0.11
Nodes (25): runRoot(), TestBashEnvGivenToEnvFlag(), TestConvertErrorPaths(), TestConvertParseError(), TestConvertToFile(), TestConvertToStdout(), TestConvertWriteError(), TestGenSecretsRefusesEmptyValue() (+17 more)

### Community 41 - "haCfg"
Cohesion: 0.13
Nodes (20): TestCheckStorageClass(), TestReachable(), TestResolveStorageClass(), ProductKeyRoles(), RestartOrder(), haCfg(), TestHARoles(), TestProductKeyRoles() (+12 more)

### Community 42 - "semp_test.go"
Cohesion: 0.10
Nodes (31): TestHTTPStatusHelpers(), anyHTTP2xx(), curlConfigLine(), Ops, httpBody(), lastHTTPStatus(), revertActivityMateBody(), sempPort() (+23 more)

### Community 43 - "Set"
Cohesion: 0.14
Nodes (14): Entry, checkShort(), Set, New(), sample(), TestAccessorsReturnCopies(), TestDigitsAreAllowed(), TestExpandResolvesBothSpellings() (+6 more)

### Community 44 - "prep.go"
Cohesion: 0.10
Nodes (17): bufio.Reader, foreignObjects(), Cluster, isBuiltinLabel(), isOurs(), joinManifests(), namespaceManifest(), namespaceOursSecrets() (+9 more)

### Community 45 - "operator.go"
Cohesion: 0.22
Nodes (14): GenOperator(), GenOperatorSecrets(), joinYAMLDocs(), nonEmpty(), operatorImage(), RenderOperator(), renderOperatorNS(), splitAfterNamespace() (+6 more)

### Community 46 - "newTestOps"
Cohesion: 0.05
Nodes (89): appUsers(), Ops, matchCLI(), newTestOps(), ranContains(), TestAdditionalUsers(), TestAdditionalUsersEmpty(), TestAdditionalUsersRejectsBadValues() (+81 more)

### Community 47 - "Cluster"
Cohesion: 0.16
Nodes (8): Cluster, HARoles(), lbServiceName(), podName(), pvcName(), stsName(), TestResourceNames(), Cluster

### Community 48 - "runP2"
Cohesion: 0.15
Nodes (17): containsPlatform(), Env, platformNames(), resolveEnv(), resolvePlatform(), roleLabel(), TestResolvePlatform(), writeEnv() (+9 more)

### Community 49 - "manager.go"
Cohesion: 0.14
Nodes (11): exactName(), orNone(), platformTitle(), replacePSKLine(), secretSummary(), setOrMissing(), solaceRows(), splitLimit() (+3 more)

### Community 50 - "Probe"
Cohesion: 0.18
Nodes (19): ids(), TestRegistryCoversEveryClaim(), TestRunProbesSkipsOffPlatformAndExitsOnRefuted(), TestSelectProbes(), probeP1(), p2Classify(), probeP2(), probeP4() (+11 more)

### Community 51 - "newRootCmd"
Cohesion: 0.10
Nodes (24): TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestDangerousVerbsHaveNoBareAlias(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownNoun(), TestStartStopHaveNoAlias(), TestAllowCommandIsRegisteredWhereItExecutes(), TestAllowCommandIsRepeatable() (+16 more)

### Community 52 - "podmanEnv"
Cohesion: 0.22
Nodes (17): Env, podmanEnv(), TestCtrNameReadsTheConfiguredContainer(), TestP1DeclinedConsentTouchesNothing(), TestP1DryRunTouchesNothing(), TestP1RestoresAfterAFailureMidSequence(), TestP1RestoreWarnsLoudlyWhenItCannotRestore(), TestP1RunsStopRotateStartReadInOrder() (+9 more)

### Community 53 - "newEchoMgr"
Cohesion: 0.11
Nodes (20): bytes.Buffer, Manager, newEchoMgr(), rootlessNoFileMgr(), setNoFile(), TestManagerCheckDryRun(), TestManagerDeletePodmanPurgeRootless(), TestManagerDeployPodmanDryRunHidesSecretBytes() (+12 more)

### Community 54 - "Operations"
Cohesion: 0.20
Nodes (10): Bringing up a fresh cluster, Docker and Podman mechanics, Extra CLI users differ by platform, Operations, Post-deployment configuration order, Re-deploying is safe and explicit, Removing a broker: what stays, what goes, Rendering without applying (+2 more)

### Community 55 - "Test catalogue"
Cohesion: 0.12
Nodes (17): abbrev_test.go, convert_test.go, Coverage, examples_test.go, examples_test.go, internal/abbrev, internal/convert, internal/examples (+9 more)

### Community 56 - "scripts_test.go"
Cohesion: 0.07
Nodes (31): containsAnyFold(), countContains(), TestContainsAnyFold(), TestCountContains(), TestServerCert(), TestValidName(), validCLILine(), validCLIPassword() (+23 more)

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
Cohesion: 0.24
Nodes (12): orValue(), setOrMissing(), setOrNone(), failRow(), Cluster, info(), okRow(), skipRow() (+4 more)

### Community 62 - "scaling_test.go"
Cohesion: 0.11
Nodes (19): scalingTier, containerMem(), Config, Config, setContainerMem(), TestApplyScalingTierDefaultsContainerBlocks(), TestApplyScalingTierDefaultsK8s(), TestApplyScalingTierDefaultsMemOverride() (+11 more)

### Community 63 - "NewManager"
Cohesion: 0.26
Nodes (12): NewManager(), TestManagerNilSinks(), Manager, mgrOver(), TestCtrExecutorRefusesUnapprovedRuntime(), TestCtrRuntimeDefaultArgvUnchanged(), TestCtrTransportHonoursRuntime(), TestManagerHonoursRuntime() (+4 more)

### Community 65 - "Configuration"
Cohesion: 0.22
Nodes (9): Choosing the env file, Configuration, Migrating from the bash env files (`solace-util convert`), Scaling, Scaling tiers, Secrets, The command fields are executable content, The keys (+1 more)

### Community 66 - "runP1"
Cohesion: 0.33
Nodes (10): ctrName(), Env, Env, Result, p1Classify(), p1ReadMountedSecret(), p1Restore(), p1Throwaway() (+2 more)

### Community 68 - "command_test.go"
Cohesion: 0.15
Nodes (14): decodeRuntime(), TestCommandArgsDoesNotAliasCommand(), TestCommandNameAndArgs(), TestCommandString(), TestCommandUnmarshal(), TestCommandUnmarshalPropagatesDecodeErrors(), TestCommandUnmarshalRejectsOtherKinds(), TestRuntimeDefaults() (+6 more)

### Community 69 - "Developer guide"
Cohesion: 0.22
Nodes (9): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Repository layout, Tests (+1 more)

### Community 70 - "validate.go"
Cohesion: 0.22
Nodes (11): keyValueEntries, foldToEnvVar(), Config, missingErr(), platformKey(), requireAll(), requireKeyValue(), sortStrings() (+3 more)

### Community 71 - "Load"
Cohesion: 0.23
Nodes (16): minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadReadError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs(), TestLoadSecretRefErrors() (+8 more)

### Community 72 - "Live-environment probe harness (`solace-itest`)"
Cohesion: 0.20
Nodes (10): After a run, Build and run, Flags, Live-environment probe harness (`solace-itest`), Operator-run only, Safety, Testing the harness itself, The probes (+2 more)

### Community 73 - "itest_test.go"
Cohesion: 0.16
Nodes (14): Env, sinkOver(), TestConsent(), testEnv(), TestListProbesNeedsNoEnvFile(), TestP1Classify(), TestP1ThrowawayIsUsable(), TestP2Classify() (+6 more)

### Community 74 - "fileExists"
Cohesion: 0.18
Nodes (11): fileExists(), assertMode(), maskedKeys(), TestManagerDeletePodmanStopFailsServiceActiveBlocksRemoval(), TestManagerDeletePodmanStopFailsServiceInactiveProceeds(), TestManagerDeletePodmanStopFailsStateUnknownBlocksRemoval(), TestManagerDeployDockerComposeWritesFile(), TestManagerDeployDockerDryRunMasksSecretEnv() (+3 more)

### Community 76 - "compareVersions"
Cohesion: 0.31
Nodes (8): compareVersions(), imageTag(), operatorVersionWarning(), parseVersion(), TestCompareVersions(), TestConfirmNoDowngradeIsSilentInAPreview(), TestImageTag(), TestOperatorVersionWarningFiresOnlyOnADowngrade()

### Community 77 - "Consenter"
Cohesion: 0.33
Nodes (4): os.File, isTTY(), TestIsTTYIsFalseForANonTerminal(), Consenter

### Community 78 - "output_test.go"
Cohesion: 0.26
Nodes (13): sink(), TestKVBlockAlignsOnTheLongestKey(), TestKVRowAtLeadsWithTheTag(), TestKVRowHonorsAnExplicitWidth(), TestLeveledLinesCarryTheTagAndNoArrow(), TestLineAddsNoPrefix(), TestSectionPadsToWidth(), TestStepKeepsTheArrowPrefix() (+5 more)

### Community 79 - "Fixtures and doubles"
Cohesion: 0.50
Nodes (4): Fixtures and doubles, Injectable seams, Per-package doubles, Shared env fixtures

### Community 80 - "internal/tools/itest"
Cohesion: 0.50
Nodes (4): internal/tools/itest, itest_test.go, live_test.go, probes_live_test.go

### Community 81 - "runP4"
Cohesion: 0.38
Nodes (9): TestCPUAndMemoryConversion(), cpuToNano(), Env, Result, memToBytes(), p4Classify(), p4Parse(), runP4() (+1 more)

### Community 82 - "age"
Cohesion: 0.22
Nodes (9): orNone(), age(), roleRank(), Cluster, operatorRunningImage(), ownedPods(), ownsPod(), TestOperatorRunningImageWithNoContainers() (+1 more)

### Community 83 - "inspect_test.go"
Cohesion: 0.18
Nodes (17): podHealth(), pvcLevel(), replicaLevel(), serviceAddress(), loadFixture(), TestAgeMatchesKubectlShape(), TestBrokerConditionRowsFromLiveCapture(), TestDecodeBrokerCRFromLiveCapture() (+9 more)

### Community 85 - "testing.T"
Cohesion: 0.12
Nodes (38): testing.T, captureStdout(), helperCommand(), TestEchoDefaultWriter(), TestEchoOutput(), TestEchoOutputInput(), TestEchoRun(), TestEchoRunEnv() (+30 more)

### Community 86 - "inspect.go"
Cohesion: 0.17
Nodes (17): time.Time, brokerConditionRows(), conditionLevel(), findCondition(), Cluster, normalizeToList(), TestConditionLevelDegradesSafely(), TestNormalizeToListHandlesBothKubectlShapes() (+9 more)

### Community 88 - "run"
Cohesion: 0.19
Nodes (12): TestUnknownProbeIsAUsageError(), Env, main(), parseFlags(), run(), runProbes(), splitProbe(), verdictName() (+4 more)

### Community 89 - "ContainerSecret"
Cohesion: 0.22
Nodes (6): opCtrGenSecrets(), ContainerSecret, SecretPreflight(), SecretScript(), shQuote(), TestShQuote()

### Community 90 - "Ops"
Cohesion: 0.15
Nodes (5): time.Duration, Ops, Runner, Cluster, Env

### Community 91 - "scripts.go"
Cohesion: 0.08
Nodes (36): showCmd, field(), TestField(), TestLastLines(), TestPrimaryRedundancyUp(), TestLastLinesEqualCount(), assertLeaderScript(), disableDefaultVPNScript() (+28 more)

### Community 92 - "container/preflight_test.go"
Cohesion: 0.25
Nodes (7): TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestComposeSecretEnvNamesCannotBeSystemVars(), TestPreflightFailureStopsLifecycle(), TestPreflightFailureStopsTheDeploy(), TestPreflightHintIsPlatformShaped(), TestPreflightIsPreviewableUnderDryRun(), TestPreflightRunsBeforeAnything()

### Community 93 - "Troubleshooting"
Cohesion: 0.33
Nodes (6): A wrapper runtime is refused, Docker compose secrets need compose 2.23.1+, File descriptors on rootless podman, Podman secret flags, Troubleshooting, Wrong cluster (kubeconfig drift)

### Community 94 - "ResolveEnvPath"
Cohesion: 0.40
Nodes (5): envTree(), TestResolveEnvPath(), TestResolveEnvPathDefaultInBaseDir(), TestResolveEnvPathEmptyBaseDir(), ResolveEnvPath()

### Community 98 - "Abbreviations"
Cohesion: 0.33
Nodes (6): Abbreviations, Commands, Flag shorthands, How a short form is resolved, Node roles, Platforms

### Community 99 - "Command reference"
Cohesion: 0.40
Nodes (5): Command reference, Global flags, Index, Reading this reference, Tree

### Community 100 - "internal/engine"
Cohesion: 0.67
Nodes (3): internal/engine, resolve_test.go, runner_test.go

## Knowledge Gaps
- **175 isolated node(s):** `solace`, `showCmd`, `Config`, `operatorTmplVars`, `NO_COLOR` (+170 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Platform` connect `Platform` to `bg`, `captureStdout`, `config_test.go`, `Manager`, `Config`, `load.go`, `Command`, `NewTransport`, `convert_test.go`, `manager_test.go`, `DetectPlatforms`, `CheckCommand`, `Compose`, `semp_test.go`, `runP2`, `manager.go`, `Probe`, `newEchoMgr`, `scaling_test.go`, `NewManager`, `containerTransport`, `validate.go`, `Load`, `ContainerSecret`, `Ops`?**
  _High betweenness centrality (0.071) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `bg`, `cli_test.go`, `captureStdout`, `Manager`, `NewTransport`, `NewCluster`, `convert_test.go`, `manager_test.go`, `secrets_test.go`, `prep_test.go`, `recRunner`, `render.go`, `Compose`, `haCfg`, `semp_test.go`, `prep.go`, `operator.go`, `newTestOps`, `Cluster`, `newEchoMgr`, `NewManager`, `containerTransport`, `kubectlTransport`, `ContainerSecret`, `Ops`, `Image`?**
  _High betweenness centrality (0.059) - this node is a cross-community bridge._
- **Why does `Manager` connect `Manager` to `bg`, `Ops`, `Config`, `manager.go`, `io.Writer`, `Platform`, `NewManager`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Are the 10 inferred relationships involving `ctrCfg()` (e.g. with `TestComposeSecretEnvIsTheOnlyChildEnvironment()` and `TestComposeSecretEnvNamesCannotBeSystemVars()`) actually correct?**
  _`ctrCfg()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **Are the 10 inferred relationships involving `newCapMgr()` (e.g. with `NewManager()` and `TestComposeSecretEnvIsTheOnlyChildEnvironment()`) actually correct?**
  _`newCapMgr()` has 10 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `showCmd`, `Config` to the rest of the system?**
  _175 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `bg` be split into smaller, more focused modules?**
  _Cohesion score 0.06420906888196608 - nodes in this community are weakly interconnected._
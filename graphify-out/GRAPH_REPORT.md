# Graph Report - solace-cnt-scripts  (2026-11-14)

## Corpus Check
- 205 files · ~750,228 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 24 file(s) not represented in the graph (top: .golden 17, (none) 3, .cli 2)

## Summary
- 4186 nodes · 21085 edges · 117 communities (108 shown, 9 thin omitted)
- Extraction: 73% EXTRACTED · 27% INFERRED · 0% AMBIGUOUS · INFERRED: 5798 edges (avg confidence: 0.89)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `fc8c7e4f`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- App
- newTestOps
- .kubectl
- Role
- cli_test.go
- config_test.go
- Command Reference Docs
- Manager
- K8sConfig
- completion_test.go
- Command
- eqArgs
- NewCluster
- main_test.go
- convert_test.go
- transform_test.go
- manager_test.go
- commands.go
- coverage_test.go
- Bash Dev Script
- verify_local_test.go
- PowerShell Dev Script
- Abbreviations
- switchplan_test.go
- verify_local_test.go
- step
- Sink
- hostpath_test.go
- importops_test.go
- watch_test.go
- Solace Module Root
- statusreport_test.go
- CLAUDE.md
- NewCluster
- ResolveDomainCerts
- watchLoop
- github.com/spf13/cobra.ShellCompDirective
- VPNRepl
- annotate_test.go
- Load
- eqArgs
- inject_test.go
- secrets_test.go
- exportconfig_test.go
- MateChannel
- correctness
- podName
- config_test.go
- scripts_test.go
- Coverage
- internal/broker
- prep_test.go
- replCfg
- Per-package doubles
- The eight rules
- replication_test.go
- .OperatorApply
- Findings by area
- .ExportConfig
- replConfig
- rootless_test.go
- Platform parity: solmq-conn-util and solace-util
- captureStderr
- secrets_test.go
- blocks.go
- Platform
- .validate
- k8s/inspect.go
- .validate
- .ApplyDefaults
- Operations
- ReplSite
- What `import-config` applies
- .checkUserManagerLimits
- EnvPairs
- allowcommand_test.go
- captureStdout
- Capture
- preflight_test.go
- replicationops_test.go
- .validateContainer
- testing.T
- Configuration
- context.Context
- namespace_test.go
- broker/replication_test.go
- Config
- render_test.go
- Quadlet
- RuleFor
- resilience
- Suspected (needs confirmation)
- gen-cert.sh
- euid_test.go
- Container
- .operatorRows
- config_ops_test.go
- writeNode
- Platforms
- newRootCmd
- Decisions on the 2026-11-11 critical review
- DeriveCAName
- redundancy_test.go
- conventions
- tlssecret_test.go
- README.md
- .runMatePhase
- opPerformReplication
- capRunner
- Image
- container-security.md
- operatorversion_test.go
- strings.Builder

## God Nodes (most connected - your core abstractions)
1. `App` - 221 edges
2. `newTestOps()` - 167 edges
3. `Config` - 151 edges
4. `cli_test.go` - 151 edges
5. `NewCluster()` - 146 edges
6. `ctrCfg()` - 145 edges
7. `Per-package doubles` - 134 edges
8. `Role` - 132 edges
9. `newCapMgr()` - 126 edges
10. `loadK8s()` - 121 edges

## Surprising Connections (you probably didn't know these)
- `CR-0018 [minor] [new] platformOps is the single source of truth only for dispatchLeaf; 12 commands build it in RunE` --references--> `platformOps()`  [INFERRED]
  docs/review/2026-11-11-critical-review.md → internal/cli/platform.go
- `The preflight` --references--> `info()`  [INFERRED]
  docs/operations.md → internal/k8s/checkreport.go
- `Build` --references--> `build`  [INFERRED]
  docs/developer.md → internal/k8s/cluster.go
- `Dev script tasks` --references--> `build`  [INFERRED]
  docs/developer.md → internal/k8s/cluster.go
- `CR-0090 [minor] [new] The release target set is defined twice and nothing ties the two together; developer.md calls the local dist list the release targets` --references--> `build`  [INFERRED]
  docs/review/2026-11-11-critical-review.md → internal/k8s/cluster.go

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Legacy Bash Script Family** — bash_000_env_sh, bash_010_deploy_operator_sh, bash_020_deploy_broker_sh, bash_059_execute_cli_sh [EXTRACTED 1.00]
- **Solace Go CLI Architecture** — internal_config, internal_engine, internal_render, internal_broker, internal_k8s, internal_cli [EXTRACTED 1.00]

## Communities (117 total, 9 thin omitted)

### Community 0 - "App"
Cohesion: 0.06
Nodes (94): App, logOptions(), logSince(), warn(), runExport(), runImport(), diagnosticsOptions(), exportScope() (+86 more)

### Community 1 - "newTestOps"
Cohesion: 0.08
Nodes (67): broker_test.go, countContains(), leaderCalls(), newTestOps(), TestCountContains(), TestDiagnostics(), TestDisableDefaultUsers(), TestDisableDefaultUsersNoVPNs() (+59 more)

### Community 2 - ".kubectl"
Cohesion: 0.07
Nodes (17): Coverage, internal/k8s, names_test.go, runtime_test.go, transport_test.go, BrokerPodSuffixShape(), Cluster, pvcName() (+9 more)

### Community 3 - "Role"
Cohesion: 0.05
Nodes (36): covUploadFailsTo, downloadErrTransport, fakeTransport, recDownload, recOutput, recRun, recUpload, recUploadFile (+28 more)

### Community 4 - "cli_test.go"
Cohesion: 0.07
Nodes (99): cli_test.go, examples_test.go, allowRuntime(), capture(), captureStderr(), collectPaths(), fakeBinaryOnPath(), firstLine() (+91 more)

### Community 5 - "config_test.go"
Cohesion: 0.06
Nodes (78): config_test.go, validPlatformConfig(), assertContainerBlockDefaults(), assertContainerScaling(), envTree(), haNodesConfig(), TestAdditionalUserNameFollowsTheBrokerRule(), TestAdditionalUserNameIsStricterOnKubernetes() (+70 more)

### Community 6 - "Command Reference Docs"
Cohesion: 0.04
Nodes (50): Commands, solace-util auto-complete, solace-util auto-complete bash, solace-util auto-complete fish, solace-util auto-complete powershell, solace-util auto-complete zsh, solace-util broker cli, solace-util broker configure (+42 more)

### Community 7 - "Manager"
Cohesion: 0.08
Nodes (12): orUnknown(), Manager, orNone(), orValue(), platformTitle(), printableRow(), secretSummary(), setOrMissing() (+4 more)

### Community 8 - "K8sConfig"
Cohesion: 0.11
Nodes (15): ContainerSecurity, Operator, PodSecurity, Resources, Storage, K8sConfig, LoadBalancer, cut() (+7 more)

### Community 9 - "completion_test.go"
Cohesion: 0.22
Nodes (23): completion_test.go, driveBash(), runComplete(), runCompleteVia(), TestAllowCommandOffersNoFiles(), TestBashFiledirFallbackCompletesDirsOnly(), TestBashInitFallbackRejoinsSplitWords(), TestBashScriptDoesNotNeedBashCompletion() (+15 more)

### Community 10 - "Command"
Cohesion: 0.07
Nodes (36): cmdField, guardedCmd, command_test.go, duration_test.go, execguard_test.go, internal/config, decodeCommand(), TestCommandArgsDoesNotAliasCommand() (+28 more)

### Community 11 - "eqArgs"
Cohesion: 0.09
Nodes (46): check_test.go, cluster_test.go, ops_test.go, TestCheckStopsProbingWhenUnreachable(), TestReachable(), TestResolveStorageClass(), newCluster(), TestApplyOnStdin() (+38 more)

### Community 12 - "NewCluster"
Cohesion: 0.09
Nodes (62): operator_test.go, GenOperator(), joinYAMLDocs(), nonEmpty(), operatorImage(), RenderOperator(), splitAfterNamespace(), splitOperatorBundle() (+54 more)

### Community 13 - "main_test.go"
Cohesion: 0.20
Nodes (19): main_test.go, describe(), judge(), main(), plural(), run(), load(), TestJudge() (+11 more)

### Community 14 - "convert_test.go"
Cohesion: 0.06
Nodes (72): doc, Result, segment, vars, The keys, CR-0035 [minor] [new] v0.2.x schema keys (redundancy scalar, nodes, nodes.*.ip, storage.msgNode/monNode) were dropped, not refused by name, convert_test.go, DecodeStrict() (+64 more)

### Community 15 - "transform_test.go"
Cohesion: 0.10
Nodes (61): Block, TargetState, transform_test.go, injectedBlock(), BridgeEnablementInverted(), ClearExistingNested(), ClearExistingSyslogs(), clearNestedObjects() (+53 more)

### Community 16 - "manager_test.go"
Cohesion: 0.08
Nodes (102): manager_test.go, fileExists(), NewManager(), assertMode(), callIndex(), countCalls(), ctrCfg(), hasCall() (+94 more)

### Community 17 - "commands.go"
Cohesion: 0.14
Nodes (71): opFunc, roleOpFunc, addCommands(), addLogFlags(), addPodFlag(), group(), newBrokerCLICmd(), newBrokerCmd() (+63 more)

### Community 18 - "coverage_test.go"
Cohesion: 0.07
Nodes (66): coverage_test.go, verify_ops_test.go, matchCLI(), ranContains(), writeFile(), TestDiagnosticsMkdirError(), TestDiagnosticsRunError(), TestDiagnosticsTwoRolesNoBundle() (+58 more)

### Community 19 - "Bash Dev Script"
Cohesion: 0.18
Nodes (21): finish(), log_init(), main(), NO_COLOR, dev.sh script, build_one(), cap(), die() (+13 more)

### Community 20 - "verify_local_test.go"
Cohesion: 0.06
Nodes (61): backupTarget, Credential, semp_test.go, sempmate_test.go, TestHTTPStatusHelpers(), anyHTTP2xx(), bridgeHostPort(), curlConfigFlag() (+53 more)

### Community 21 - "PowerShell Dev Script"
Cohesion: 0.18
Nodes (17): Get-Log(), Get-Now(), Build-One(), Cap(), Ok(), Step(), Task-build(), Task-cov() (+9 more)

### Community 22 - "Abbreviations"
Cohesion: 0.20
Nodes (9): Abbreviations, Flag shorthands, Node roles, Platforms, Command reference, Global flags, Index, Reading this reference (+1 more)

### Community 23 - "switchplan_test.go"
Cohesion: 0.12
Nodes (44): PhaseKind, SiteState, SwitchAction, SwitchPhase, switchplan_test.go, BuildSwitchPlan(), confirmDemotions(), ExecuteSwitchPlan() (+36 more)

### Community 24 - "verify_local_test.go"
Cohesion: 0.14
Nodes (41): verify_local_test.go, cliScriptPath(), defaultLocalAddrs(), isCurl(), newLocalOps(), rd(), seqTransport(), TestBackupActivityStateReadsTheMateColumn() (+33 more)

### Community 25 - "step"
Cohesion: 0.10
Nodes (20): CR-0017 [major] [new] 15 real-capture tests skip in every gate because their fixtures are gitignored and absent, CR-0047 [minor] [new] Error-path tests assert only err != nil and cannot tell which step failed or whether later writes ran, CR-0048 [minor] [new] Secret-leak tests scan only stdout; stderr narration and returned errors are never checked, CR-0076 [nit] [new] Coverage-only tests in coverage_test.go assert constructor defaults and trivial edges, CR-0077 [nit] [new] Tests and code cross-reference unresolvable review IDs (M1, M4, M5, M11, M12, H2, H4, H5), tests, confirmActionStrict(), lineSink() (+12 more)

### Community 26 - "Sink"
Cohesion: 0.06
Nodes (61): BlockDiff, blockKey, DiffResult, mergedBlock, mergedLine, diff_test.go, internal/output, output_test.go (+53 more)

### Community 27 - "hostpath_test.go"
Cohesion: 0.08
Nodes (33): hostpath_test.go, guardConfig(), BaseName(), CheckHostPath(), expandTilde(), expandTildeToken(), HasPathSeparator(), IsAbsHostPath() (+25 more)

### Community 28 - "importops_test.go"
Cohesion: 0.07
Nodes (72): ImportResult, PlannedSection, importops_test.go, defaultVPNFirst(), describeScope(), ImportPlan, importIgnore(), orNone() (+64 more)

### Community 29 - "watch_test.go"
Cohesion: 0.15
Nodes (31): watch_test.go, subtractWatch(), deniedCluster(), deployJSON(), namedOperatorJSON(), operatorItem(), TestAmbiguousOperatorAbortsTheDeploy(), TestDesiredWatchAppendsTheBrokerNamespace() (+23 more)

### Community 31 - "statusreport_test.go"
Cohesion: 0.17
Nodes (28): statusreport_test.go, reportCluster(), TestBrokerReportDetailAddsStorageAndPlacement(), TestBrokerReportOmitsWorkloadsThatAreNotTheBrokers(), TestBrokerReportOrdersPodsOfTheSameRoleByName(), TestBrokerReportPreviewStopsAtAnUnapprovedCommand(), TestBrokerReportReadsEachKindOnce(), TestBrokerReportReadsOnlyWhatTheEnvFileNames() (+20 more)

### Community 32 - "CLAUDE.md"
Cohesion: 0.27
Nodes (5): bash/000-env.sh, bash/010-deploy-operator.sh, bash/020-deploy-broker.sh, bash/059-execute-cli.sh, docker-podman/000-env.sh

### Community 33 - "NewCluster"
Cohesion: 0.09
Nodes (69): checkreport_test.go, deploy_test.go, preflight_test.go, TestAdminPasswordPreflight(), TestAdditionalUsersRowSaysWhereThePasswordsLand(), TestOperatorInstallVerdict(), TestOperatorInstallVerdictShowsOnceInEveryView(), TestOperatorRowsCompareTheRunningVersionWithTheConfiguredOne() (+61 more)

### Community 34 - "ResolveDomainCerts"
Cohesion: 0.08
Nodes (41): Broker, CertDir, DirReader, DomainCerts, fakeDirEntry, domaincerts_resolve_test.go, checkNoDotDot(), DefaultDirReader() (+33 more)

### Community 35 - "watchLoop"
Cohesion: 0.33
Nodes (4): watchInterval(), watchLabel(), watchLoop(), withWatch()

### Community 36 - "github.com/spf13/cobra.ShellCompDirective"
Cohesion: 0.33
Nodes (6): completeDirs(), completeEnvFiles(), completePlatforms(), completeRoles(), isEnvFileName(), matching()

### Community 37 - "VPNRepl"
Cohesion: 0.12
Nodes (14): AdminState, QueueState, replMate, ReplRole, VPNRepl, ReplicationConfigResult, mateConvergenceShutdowns(), missingListedVPNs() (+6 more)

### Community 38 - "annotate_test.go"
Cohesion: 0.11
Nodes (45): annotate_test.go, applyMetaField(), isMarker(), markerRegion(), markerSection(), markerVerb(), parseKV(), readMarker() (+37 more)

### Community 40 - "Load"
Cohesion: 0.07
Nodes (47): scalingTier, scaling_test.go, minimalK8s(), TestLoadBashEnvFileHint(), TestLoadNotYAMLHint(), TestLoadParseError(), TestLoadRejectsTheOldK8sSection(), TestLoadResolvesSecretRefs() (+39 more)

### Community 41 - "eqArgs"
Cohesion: 0.26
Nodes (21): answer(), composeUp(), composeUpArgs(), containsStr(), deployDigest(), dockerHost(), maskedKeys(), neverAsk() (+13 more)

### Community 42 - "inject_test.go"
Cohesion: 0.17
Nodes (28): lineRole, serviceLine, shutdownStyle, svcKey, inject_test.go, classifyServiceRest(), containsPortCommand(), describeKey() (+20 more)

### Community 43 - "secrets_test.go"
Cohesion: 0.18
Nodes (23): secrets_test.go, ResolveSecretValues(), certFixture(), TestCertSecretRemovalFailureIsFatal(), TestContainerBundleCarriesTheChain(), TestDeletePodmanToleratesAMissingBundle(), TestDeployDockerFailsBeforeAskingOrRunningAnything(), TestDeployDockerReReadsTheCertificate() (+15 more)

### Community 44 - "exportconfig_test.go"
Cohesion: 0.12
Nodes (40): exportconfigReadCounter, opRunner, exportconfig_test.go, internal/cli, runRootWith(), TestAssertLeaderOnKubernetesTouchesOnlyThePrimaryPod(), TestDeployOperatorNoPromptApprovesOnlyTheDowngrade(), TestK8sConfirmDeclined() (+32 more)

### Community 46 - "MateChannel"
Cohesion: 0.14
Nodes (30): CLIRunner, fakeRun, matechannel_test.go, MateChannel, NewCLIMate(), mateReplies(), newTestMate(), TestLocalMateRunsThroughOps() (+22 more)

### Community 47 - "correctness"
Cohesion: 0.07
Nodes (28): correctness, CR-0001 [blocker] [new] A VPN-scoped import deletes every virtual hostname on the target, silently, CR-0002 [blocker] [new] Seven kubernetes.* values reach the broker CR unquoted with no load-time grammar (YAML/--- injection), CR-0004 [blocker] [new] WATCH_NAMESPACE read back from a running Deployment is rendered raw into the operator bundle an admin applies, CR-0006 [major] [new] Namespace emptiness gate misses RBAC, NetworkPolicy, Ingress and CR objects; --no-prompt then deletes the namespace, CR-0007 [major] [new] Primary-active switch gate is inverted when the local site is addressed through its backup node, CR-0008 [major] [new] PrimaryActive reads Enabled-Released (or an unreadable report) as standalone and passes the gate, CR-0009 [major] [new] An operator found in another namespace drives deploy and remove, which then write and delete in the resolved namespace (+20 more)

### Community 49 - "config_test.go"
Cohesion: 0.14
Nodes (30): sempExecuteResult, sempRedundancyReply, sempReplicationReply, sempVPNReplicationReply, showCmd, span, AdditionalUser, Node (+22 more)

### Community 50 - "scripts_test.go"
Cohesion: 0.08
Nodes (37): scripts_test.go, validCLILine(), assertLeaderScript(), defaultUsersScript(), disableDefaultUsersScript(), disableDefaultVPNScript(), enableDefaultUsersScript(), parseVPNNames() (+29 more)

### Community 51 - "Coverage"
Cohesion: 0.08
Nodes (29): interactiveFailRunner, lockedWriter, runFailRunner, Repository layout, internal/engine, runner_test.go, noEnv, redirect (+21 more)

### Community 52 - "internal/broker"
Cohesion: 0.15
Nodes (22): chunk, chunkResult, driver_test.go, chunkName(), keywordPattern(), parseDriverOutput(), renderDriver(), reportedOnce() (+14 more)

### Community 53 - "prep_test.go"
Cohesion: 0.10
Nodes (52): adminsecret_test.go, prep_test.go, k8sAdminPassword(), AdminSecretInUse(), ReadAdminPassword(), operatorManagedCfg(), TestAdminSecretInUseFollowsTheStates(), TestOperatorAdminSecretNameMatchesTheBrokerNames() (+44 more)

### Community 54 - "replCfg"
Cohesion: 0.20
Nodes (27): matechannel_test.go, CLIArg(), CLIScriptPath(), NewMateChannel(), ReadSecretKey(), hasPrefixArgv(), indexOf(), operandOf() (+19 more)

### Community 55 - "Per-package doubles"
Cohesion: 0.09
Nodes (73): leaderRun, opCall, ops_k8s_test.go, Per-package doubles, ctrRedeployApp(), failDisableDefaultUsersUpload(), isDeletePod(), isGetPod() (+65 more)

### Community 56 - "The eight rules"
Cohesion: 0.22
Nodes (9): 1. Deny extended privileges, 2. Deny privilege escalation, 3. Run as a non-root identity, 4. Bind only non-root ports, 5. Hand secrets over as files, 6. Narrow the filesystem, 7. State the setting, do not inherit it, 8. Document the version floor each setting needs (+1 more)

### Community 57 - "replication_test.go"
Cohesion: 0.12
Nodes (36): replRig, Commands, solace-util broker, Configuring a site, Data replication, Permissions, Reaching the mate, Switching roles (+28 more)

### Community 59 - "Findings by area"
Cohesion: 0.08
Nodes (26): build, cicd, CR-0018 [minor] [new] platformOps is the single source of truth only for dispatchLeaf; 12 commands build it in RunE, CR-0019 [minor] [new] A removed dry-run mode survives as 34 production branches keyed on the test-only Echo runner, CR-0045 [minor] [new] Dead code with no production caller is kept covered by its own tests, CR-0046 [minor] [new] set/MISSING and orNone rendering is duplicated per package and has already drifted, CR-0049 [minor] [new] tag.yml has drifted from the init-devops template: any v* tag publishes a release that GitHub marks Latest, CR-0050 [minor] [new] The Windows CI leg runs dev.ps1 under pwsh 7, so the PowerShell 5.1 compatibility it claims is never tested (+18 more)

### Community 60 - ".ExportConfig"
Cohesion: 0.50
Nodes (3): Region, regionFor(), keepRegion()

### Community 61 - "replConfig"
Cohesion: 0.18
Nodes (15): Replication, secretRef, replication_test.go, replConfig(), TestReplicationCredentialChars(), TestReplicationLocate(), TestReplicationPassEnvResolves(), TestSiteCommandGuarded() (+7 more)

### Community 62 - "rootless_test.go"
Cohesion: 0.18
Nodes (42): rootless_test.go, failOnCall(), fakeEnv(), healthyRootlessOut(), lingerOff(), rootlessMgr(), TestCheckDataDirAcceptsAnAlreadyChownedDir(), TestCheckDataDirRefusesAnUnwritableParent() (+34 more)

### Community 63 - "Platform parity: solmq-conn-util and solace-util"
Cohesion: 0.29
Nodes (6): Contract, Decisions, Divergences, Inherent differences, Platform parity: solmq-conn-util and solace-util, Runner seams

### Community 64 - "captureStderr"
Cohesion: 0.13
Nodes (21): layer, Injectable seams, internal/convert, internal/tools/vulnjudge, Running the tests, Summary, Test catalogue, TestAPipeIsRefusedAndNeverRead() (+13 more)

### Community 65 - "secrets_test.go"
Cohesion: 0.14
Nodes (36): secrets_test.go, GenSecrets(), AdditionalUsersSecret(), AdminSecret(), DockerRegistrySecret(), dockerRegistrySecret(), operatorRegcred(), readTLSMaterial() (+28 more)

### Community 66 - "blocks.go"
Cohesion: 0.12
Nodes (35): token, blocks_test.go, internal/broker, omit_test.go, Annotate(), TestAnnotateIsDeterministic(), TestAnnotateRegionAndSectionFencesAreBalanced(), TestAnnotateRoundTripCaptureIsMarked() (+27 more)

### Community 67 - "Platform"
Cohesion: 0.11
Nodes (30): How a short form is resolved, applyAliases(), argumentLine(), availableSubs(), writeCommand(), checkAllowCommand(), checkFlagPlatforms(), commandPlatforms() (+22 more)

### Community 68 - ".validate"
Cohesion: 0.10
Nodes (17): solace-util, solace-util operator, A replication switch refuses before changing anything, A wrapper runtime is refused, Docker and Podman mechanics, Docker compose secrets need compose 2.23.1+, Forbidden: cannot list deployments at the cluster scope, Import reported failure (+9 more)

### Community 69 - "k8s/inspect.go"
Cohesion: 0.07
Nodes (43): inspect_test.go, orNone(), age(), anyKnownCondition(), brokerConditionRows(), conditionLevel(), findCondition(), normalizeToList() (+35 more)

### Community 70 - ".validate"
Cohesion: 0.13
Nodes (24): orValue(), setOrMissing(), setOrNone(), storageClassSuitable(), additionalUsersRow(), adminPassState(), adminSecretRow(), allCustomMounted() (+16 more)

### Community 71 - ".ApplyDefaults"
Cohesion: 0.17
Nodes (10): guardCommandOf(), setGuardCommand(), TestValidatorAndExecutorAgree(), applyContainerBlockDefaults(), podmanRunUser(), setDefault(), setDefaultCmd(), setDefaultInt() (+2 more)

### Community 72 - "Operations"
Cohesion: 0.10
Nodes (21): Bringing up a fresh cluster, Broker scope is a fixed classification, Every destructive command confirms, Exit codes, Exporting and importing configuration, Extra CLI users differ by platform, Handling the artifact, Kubernetes-specific hazards (+13 more)

### Community 73 - "ReplSite"
Cohesion: 0.21
Nodes (19): confirmReplicationConfig(), mateChannel(), mateSEMPPassword(), resolveSites(), replApp(), siteNamed(), TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(), TestConfirmReplicationConfigNamesTheBiggerHammer() (+11 more)

### Community 74 - "What `import-config` applies"
Cohesion: 0.33
Nodes (6): Appliance only (skipped on a software broker), Applied, Applied, minus some lines, Never applied, Sections that interrupt a service, What `import-config` applies

### Community 75 - ".checkUserManagerLimits"
Cohesion: 0.19
Nodes (4): delegateSetting(), missingControllers(), pamCoreValue(), parseUnitProps()

### Community 76 - "EnvPairs"
Cohesion: 0.11
Nodes (26): internal/render, servercert_test.go, Compose(), ContainerSecrets(), containerSecretSpecs(), EnvPairs(), groupKey(), itoa() (+18 more)

### Community 77 - "allowcommand_test.go"
Cohesion: 0.18
Nodes (18): allowcommand_test.go, Fixtures and doubles, Shared env fixtures, TestAllowCommandApprovesAWrappedRuntime(), TestAllowCommandIsRegisteredWhereItExecutes(), TestAllowCommandIsRepeatable(), TestAllowCommandRejectedWhereNothingExecutes(), TestAllowCommandRejectsBadValues() (+10 more)

### Community 78 - "captureStdout"
Cohesion: 0.21
Nodes (16): watch_test.go, captureStdout(), TestCtrRedeployGate(), TestOpCtrVerifyRedundancyRunsCoordinated(), opCtrVerifyRedundancy(), stopAfter(), TestOperatorStatusWatches(), TestStatusWithoutWatchRunsOnce() (+8 more)

### Community 79 - "Capture"
Cohesion: 0.18
Nodes (13): Capture, Omission, appliance_test.go, loadApplianceCapture(), TestApplianceCaptureIsReadAsAnAppliance(), TestApplianceOnlySectionsCarryContent(), TestApplianceReplicationGrammarDiffersFromSoftware(), TestEveryApplianceBrokerSectionIsClassified() (+5 more)

### Community 80 - "preflight_test.go"
Cohesion: 0.20
Nodes (13): preflight_test.go, resolve_test.go, TestManagerLifecycleDocker(), TestManagerLifecyclePodmanSystemctl(), TestComposeSecretEnvIsTheOnlyChildEnvironment(), TestComposeSecretEnvNamesCannotBeSystemVars(), TestPreflightFailureStopsLifecycle(), TestPreflightRunsBeforeAnything() (+5 more)

### Community 81 - "replicationops_test.go"
Cohesion: 0.18
Nodes (37): replicationops_test.go, hasCall(), replicationVPNLines(), replListed(), replReadOnlyResponder(), replTestMate(), replTestMateMatching(), siteAEntry() (+29 more)

### Community 82 - ".validateContainer"
Cohesion: 0.04
Nodes (41): keyValueEntries, ReplVia, ReplViaKube, ReplViaSEMP, artifactvalues_test.go, checkTimezone(), portOrRange(), TestCheckTimezone() (+33 more)

### Community 83 - "testing.T"
Cohesion: 0.09
Nodes (61): inspect_test.go, limits_test.go, storage_test.go, storageCfg(), TestCustomMountIgnoresRolesOutsideTheGroup(), TestCustomMountMustCoverEveryNode(), TestCustomMountRejectsAnEmptyClaim(), TestCustomMountRejectsAnUnknownRoleKey() (+53 more)

### Community 84 - "Configuration"
Cohesion: 0.15
Nodes (13): Bring your own TLS Secret, Choosing the env file, Configuration, Keys that were renamed, Migrating from the bash env files (`solace-util convert`), Relative paths resolve against the env file, not the current directory, Replication, Scaling (+5 more)

### Community 85 - "context.Context"
Cohesion: 0.07
Nodes (7): scriptedMate, sempMate, containerTransport, schemaTransport(), sempReplRole(), xmlEscape(), KV

### Community 86 - "namespace_test.go"
Cohesion: 0.35
Nodes (14): namespace_test.go, nsItem(), nsList(), occCfg(), occCluster(), TestNamespaceOccupantsAsksOneQuestion(), TestNamespaceOccupantsDiscountsTheBrokerAndWhatItOwns(), TestNamespaceOccupantsDiscountsWhatKubernetesPutsThere() (+6 more)

### Community 87 - "broker/replication_test.go"
Cohesion: 0.05
Nodes (80): BrokerType, cliMate, colSpan, replication_test.go, transcript_test.go, ValidVPNName(), validVPNName(), bannerType() (+72 more)

### Community 88 - "Config"
Cohesion: 0.11
Nodes (22): TLS, adminsecret_test.go, imagepullsecret_test.go, adminCfg(), TestAdminSecretNameFollowsTheStates(), TestAdminSecretRefusesTheOperatorsOwnNames(), TestContainersStillRequireTheAdminPassword(), TestKubernetesValidatesWithoutAnAdminPassword() (+14 more)

### Community 89 - "render_test.go"
Cohesion: 0.14
Nodes (43): render_test.go, boolStr(), BrokerCR(), ComposeProject(), containerArtifacts(), load(), loadRootful(), TestAdminCredentialsSecretFollowsTheStates() (+35 more)

### Community 90 - "Quadlet"
Cohesion: 0.16
Nodes (11): opCtrGenArtifact(), ContainerNoFile(), HealthCheck, NodeIdentity, TestLimitsCheckAssertsWhatTheArtifactAsks(), healthCmd(), Quadlet(), healthCheckFixture() (+3 more)

### Community 91 - "RuleFor"
Cohesion: 0.14
Nodes (33): Disposition, SectionRule, importdoc_test.go, sections_test.go, renderImportDocs(), TestEverySectionRuleHasAReason(), TestImportDocs(), TestOnlyOneSectionIsAppliedFirst() (+25 more)

### Community 93 - "resilience"
Cohesion: 0.11
Nodes (19): CR-0005 [major] [new] Docker broker remove empties the data dir under a broker compose down did not find, CR-0011 [major] [new] InjectShutdown warnings are produced only after the teardown has run, not in the plan, CR-0014 [major] [new] Podman redeploy forgets a written-but-unapplied change and later reports 'unchanged, left running', CR-0016 [major] [new] A switch that fails mid-plan returns a bare error: VPNs left standby at both sites are not named, CR-0036 [minor] [new] convert's validateOutput skips Load's '~' expansion and wrongly warns that '~/...' paths will not load, CR-0037 [minor] [new] DecodeStrict reads only the first YAML document; anything after a second '---' is silently ignored, CR-0038 [minor] [new] Interrupting an import leaves the credential-bearing driver on the broker and the apply running unobserved, CR-0039 [minor] [new] Driver ignores failed chunk/wrapper writes and an empty transcript, so it can report OK for what never ran (+11 more)

### Community 95 - "Suspected (needs confirmation)"
Cohesion: 0.11
Nodes (18): Accepted (not re-litigated), Coverage, CR-0081 [major] [new] RewriteAdminUsername discards every operand on the admin opener, so verification demands what was dropped, CR-0082 [major] [new] Destructive-prompt context clause can name the wrong cluster when kubernetes.command carries --context, CR-0083 [major] [new] broker restart/start wait with `rollout status statefulset`, which kubectl refuses for OnDelete StatefulSets, CR-0084 [minor] [new] Bare 'busy' keyword in the tail scan can fail a clean chunk deterministically, CR-0085 [minor] [new] Credential-bearing artifact lines are printed to stdout by the failure detail and the verification report, CR-0086 [minor] [new] Repeated --vpn duplicates every block of that VPN in the artifact (+10 more)

### Community 96 - "gen-cert.sh"
Cohesion: 0.27
Nodes (13): ask(), askPass(), askSan(), askYesNo(), buildSanList(), checkExit(), concatCnIfNotBlank(), echoUsage() (+5 more)

### Community 97 - "euid_test.go"
Cohesion: 0.67
Nodes (6): euid_test.go, runGuarded(), TestGenerateIgnoresTheEUID(), TestPodmanEUIDGuardRefusesBeforeAnyCommand(), TestPodmanEUIDGuardSkips(), writeRootlessPodmanEnv()

### Community 98 - "Container"
Cohesion: 0.27
Nodes (6): Container, DockerConfig, Network, PodmanConfig, Ulimits, TestContainerCoreLimitSpellings()

### Community 99 - ".operatorRows"
Cohesion: 0.13
Nodes (15): containsString(), installSteps(), compareVersions(), imageFromDeployment(), imageTag(), operatorVersionWarning(), parseVersion(), operatorRunningImage() (+7 more)

### Community 100 - "config_ops_test.go"
Cohesion: 0.17
Nodes (13): config_ops_test.go, TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(), TestEnableDefaultVPNFailureSkipsTheReadback(), TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(), TestExecShellScriptCleanupFailureOnlyWarns(), TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(), TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator(), TestExecShellScriptRefusesAnUnsafeNameBeforeUploading() (+5 more)

### Community 101 - "writeNode"
Cohesion: 0.21
Nodes (12): WeightedNodeTerm, NodeAffinity, NodeMatchExpr, Placement, validateMatchExprs(), validatePlacementAffinity(), parseToleration(), TestParseToleration() (+4 more)

### Community 102 - "Platforms"
Cohesion: 0.07
Nodes (49): Build, Dev script tasks, Developer guide, Gates, Goldens, Releases, Tests, The env-file generator page (+41 more)

### Community 103 - "newRootCmd"
Cohesion: 0.08
Nodes (46): aliases_test.go, exit_test.go, platform_test.go, TestAliasesDoNotCollide(), TestAliasesResolveToTheCanonicalCommand(), TestEveryAliasEntryIsLive(), TestGroupsRejectAnUnknownVerb(), TestNounGroupsRunNothing() (+38 more)

### Community 104 - "Decisions on the 2026-11-11 critical review"
Cohesion: 0.17
Nodes (10): Certificate renewal on Kubernetes without a restart, Consequences to check before the work starts, Decisions, Decisions on the 2026-11-11 critical review, Planned work the report does not hold, Rules the owner stated, Shared module, Stale objects after a config change (+2 more)

### Community 105 - "DeriveCAName"
Cohesion: 0.24
Nodes (12): domaincerts_test.go, caNameSafe(), DeriveCAName(), lastPathElement(), sanitiseCAPart(), splitCAExt(), assertUsableCAName(), TestDeriveCAName() (+4 more)

### Community 106 - "redundancy_test.go"
Cohesion: 0.21
Nodes (12): redundancy_test.go, pskCfg(), TestFillStandaloneNodeNameIgnoresAnEmptyHostname(), TestFillStandaloneNodeNameKeepsAConfiguredName(), TestFillStandaloneNodeNameNeverFillsHA(), TestFillStandaloneNodeNameUsesTheHost(), TestNothingGeneratesThePSK(), TestPSKEnvSatisfiesTheContainerRequirement() (+4 more)

### Community 107 - "conventions"
Cohesion: 0.22
Nodes (9): conventions, CR-0063 [minor] [new] Real env files under env/ are not gitignored; the old protection only covered the removed docker-podman/env layout, CR-0064 [minor] [new] The race flag differs between dev.sh and dev.ps1 and from the contract, so the Windows CI leg never runs -race, CR-0065 [minor] [new] gen-cert.sh hides every openssl error, overwrites an existing CA without asking, and splits a scalar san into characters, CR-0066 [minor] [new] docs/test.md and CLAUDE.md are out of date on the CI gate, the regen package list and the shared packages, CR-0067 [minor] [new] A script(1) terminal capture named 'typescript' is tracked at the repo root, CR-0078 [nit] [new] Non-ASCII bytes in two tracked text files, CR-0079 [nit] [new] README version sample shows the pre-patch toolchain, and the version command's Long exceeds four lines (+1 more)

### Community 108 - "tlssecret_test.go"
Cohesion: 0.43
Nodes (8): tlssecret_test.go, TestNamingTheTLSSecretDoesNotInventCertPaths(), TestNoTLSAtAllStaysValid(), TestSuppliedCertsDeriveASecretName(), TestSuppliedCertsMakeTheSecretOurs(), TestTLSCertAndKeyMustBeSetTogetherOnKubernetes(), TestTLSServerSecretNameFollowsTheStates(), tlsCfg()

### Community 109 - "README.md"
Cohesion: 0.22
Nodes (6): internal/broker, internal/cli, internal/config, internal/engine, internal/k8s, internal/render

### Community 110 - ".runMatePhase"
Cohesion: 0.36
Nodes (6): isRunCLIRejection(), newlineIf(), replPhase1Rejected(), replPhase2Rejected(), replTransportFailure(), replVPNList()

### Community 111 - "opPerformReplication"
Cohesion: 0.33
Nodes (6): mateChannelFunc, opPerformReplication(), replVPNNames(), reportReplicationStatus(), TestReportReplicationStatusPrintsEachSiteInNameOrderAndStopsAtTheFirstFailure(), verifyReplicationSwitch()

### Community 112 - "capRunner"
Cohesion: 0.10
Nodes (28): capCall, capRunner, internal/container, runtime_test.go, transport_test.go, New(), TestManagerLogsCLIShell(), mgrOver() (+20 more)

### Community 126 - "operatorversion_test.go"
Cohesion: 0.21
Nodes (18): operatorversion_test.go, TestOperatorApply(), operatorDeployJSON(), TestCompareVersions(), TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(), TestConfirmNoDowngradeIsNotSkippedByAFailedRead(), TestConfirmNoDowngradeIsSilentInAPreview(), TestConfirmNoDowngradeProceedsWhenConfirmAccepts() (+10 more)

### Community 155 - "strings.Builder"
Cohesion: 0.14
Nodes (21): shorthand, mdRow(), renderAbbrevDocs(), TestFlagShorthandsAreConsistent(), treeShorthands(), writeAbbrevTable(), writeResolutionNotes(), writeShorthandTable() (+13 more)

## Knowledge Gaps
- **13 isolated node(s):** `solace`, `NO_COLOR`, `bash/000-env.sh`, `bash/010-deploy-operator.sh`, `bash/020-deploy-broker.sh` (+8 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 284 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **9 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `App` to `cli_test.go`, `Manager`, `commands.go`, `step`, `watchLoop`, `github.com/spf13/cobra.ShellCompDirective`, `exportconfig_test.go`, `config_test.go`, `Coverage`, `prep_test.go`, `Per-package doubles`, `replication_test.go`, `captureStderr`, `Platform`, `ReplSite`, `captureStdout`, `context.Context`, `Config`, `Quadlet`, `newRootCmd`, `opPerformReplication`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._
- **Are the 117 inferred relationships involving `newTestOps()` (e.g. with `importops_test.go` and `Injectable seams`) actually correct?**
  _`newTestOps()` has 117 INFERRED edges - model-reasoned connections that need verification._
- **What connects `solace`, `NO_COLOR`, `bash/000-env.sh` to the rest of the system?**
  _13 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `App` be split into smaller, more focused modules?**
  _Cohesion score 0.06381578947368421 - nodes in this community are weakly interconnected._
- **Why does `Config` connect `Config` to `App`, `newTestOps`, `.kubectl`, `Role`, `cli_test.go`, `config_test.go`, `Manager`, `K8sConfig`, `Command`, `eqArgs`, `NewCluster`, `convert_test.go`, `manager_test.go`, `verify_local_test.go`, `hostpath_test.go`, `NewCluster`, `ResolveDomainCerts`, `Load`, `eqArgs`, `secrets_test.go`, `podName`, `config_test.go`, `prep_test.go`, `replCfg`, `Per-package doubles`, `replConfig`, `secrets_test.go`, `Platform`, `k8s/inspect.go`, `.validate`, `.ApplyDefaults`, `EnvPairs`, `testing.T`, `context.Context`, `namespace_test.go`, `render_test.go`, `Quadlet`, `Container`, `writeNode`, `Platforms`, `redundancy_test.go`, `tlssecret_test.go`, `capRunner`, `Image`?**
  _High betweenness centrality (0.025) - this node is a cross-community bridge._
- **Are the 11 inferred relationships involving `Config` (e.g. with `config_test.go` and `execguard_test.go`) actually correct?**
  _`Config` has 11 INFERRED edges - model-reasoned connections that need verification._
- **Should `newTestOps` be split into smaller, more focused modules?**
  _Cohesion score 0.08269394714407502 - nodes in this community are weakly interconnected._
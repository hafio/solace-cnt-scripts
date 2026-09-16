package cli

import (
	"github.com/spf13/cobra"

	"solace/internal/config"
)

// The command tree: NOUN first, then verb.
//
// `broker <verb>` and `operator <verb>`. The two nouns are the only things this tool
// acts on, and each is a group that runs NOTHING itself: `solace-util broker` prints
// what can be done to a broker rather than doing something to one. That is the whole
// no-implicit-actions rule, and it is also what keeps the abbreviations safe -- `br` and
// `op` alone destroy nothing, so `br rm` needs the verb spelled before anything happens.
//
// Three rules follow from the noun owning its verbs:
//
//   - One lifecycle per noun. Every prerequisite a deploy needs is applied BY the
//     deploy, idempotently, and removed by the removal. There is no order to learn.
//   - Two nouns, never a third. The operator is cluster-scoped and shared, so it is
//     deployed and removed on its own rather than as a word under someone else's verb.
//   - One home per kind of work. `broker configure` applies settings the env file
//     describes; `broker perform` runs imperative one-shots that act now and are not
//     configuration at all.
//
// Two rules carry over. There is ONE tree, not one per platform: the shape is identical
// everywhere and applicability is enforced at pre-run (onlyOn / flagOnlyOn) rather than
// expressed by hiding commands, so help and completion render without an env file. And
// no command takes a positional argument except a genuine file path -- the platform is
// always --platform, the role is always --pod.

// addCommands hangs the tree off root.
func addCommands(root *cobra.Command, app *App) {
	root.AddCommand(
		newBrokerCmd(app),
		newOperatorCmd(app),
		newValidateCmd(app),
	)
}

// groupAnnotation marks a command that owns others and executes nothing itself. It
// exists so the wiring tests can tell a group apart from a command that runs something:
// a group carries a RunE (see below) but must NOT carry the pre-run or --allow-command,
// because it never reaches an external command.
const groupAnnotation = "solace_group"

// group builds a command that owns others. Run bare it prints its own help -- the list
// of what it owns -- which is the whole no-implicit-actions rule. Given a word it does
// not know, it fails loudly.
//
// That second half is why a group carries a RunE at all. Cobra answers a NON-runnable
// command by printing help to stdout and exiting 0 whatever arguments it was given, so
// `solace-util broker bogus` would report success having done nothing and a script would
// never notice. It is the same trap the completion command documents (completion.go),
// and it matters most on the nouns, since every destructive verb now hangs off one.
func group(use, short, long string) *cobra.Command {
	c := &cobra.Command{
		Use:         use,
		Short:       short,
		Long:        long,
		Annotations: map[string]string{groupAnnotation: "true"},
	}
	c.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return usagef("%q is not something %s can act on; run `%s` to see what it can",
			args[0], cmd.CommandPath(), cmd.CommandPath())
	}
	return c
}

// --- broker -----------------------------------------------------------------

const brokerLong = "Everything that acts on the broker this env file describes.\n\n" +
	"  broker validate         check the env file and the cluster or host\n" +
	"  broker generate         see exactly what would be applied\n" +
	"  broker deploy           prerequisites + the broker, idempotent\n" +
	"  broker status           watch it come up\n" +
	"  broker configure ...    settings the env file describes (certs, keys, hardening)\n" +
	"  broker perform ...      one-shot actions (diagnostics, failover test)\n" +
	"  broker stop / start     pause it without removing it\n" +
	"  broker remove           tear it down, prompting for each layer\n"

func newBrokerCmd(app *App) *cobra.Command {
	c := group("broker", "Deploy and operate the broker", brokerLong)
	c.AddCommand(
		newBrokerValidateCmd(app),
		newBrokerDeployCmd(app),
		newBrokerRemoveCmd(app),
		newBrokerGenerateCmd(app),
		newBrokerStartCmd(app),
		newBrokerStopCmd(app),
		newBrokerRestartCmd(app),
		newBrokerStatusCmd(app),
		newBrokerLogsCmd(app),
		newBrokerCLICmd(app),
		newBrokerShellCmd(app),
		newBrokerCopyCmd(app),
		newBrokerConfigureCmd(app),
		newBrokerPerformCmd(app),
	)
	return c
}

const brokerValidateLong = "Read-only, and safe to run against a system you don't want to disturb.\n\n" +
	"Reports every problem it finds in one pass, rather than stopping at the\n" +
	"first."

func newBrokerValidateCmd(app *App) *cobra.Command {
	return withLong(dispatchLeaf(app, "validate", "Check the env file and what this broker needs",
		platformOps(opK8sValidateBroker, opCtrCheck)), brokerValidateLong)
}

const brokerDeployLong = "Applies every prerequisite -- namespace, secrets, TLS -- before creating the\n" +
	"broker itself, and is safe to re-run.\n\n" +
	"Kubernetes readiness is not waited on; use `broker status` to watch it come\n" +
	"up."

func newBrokerDeployCmd(app *App) *cobra.Command {
	c := wireExec(app, &cobra.Command{
		Use:               "deploy",
		Short:             "Deploy the broker and its prerequisites",
		Long:              brokerDeployLong,
		Args:              noRolePositional,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(_ *cobra.Command, _ []string) error {
			return dispatch(platformOps(opK8sDeploy, opCtrDeployHost), app)
		},
	})
	// --pod here carries the CONTAINER meaning: which node this host is. On kubernetes a
	// deploy applies one CR covering every role, so a single pod is not something it
	// could act on -- hence docker/podman only, refused by name elsewhere.
	addPodFlag(c, app, config.Docker, config.Podman)
	addRestartFlag(c, app)
	flagOnlyOn(c, "restart", config.Docker, config.Podman)
	return c
}

const brokerRemoveLong = "Removes the deployed broker and its secrets on every platform.\n\n" +
	"This permanently deletes the broker; add --delete-data (with its own\n" +
	"confirmation) to also destroy its persistent data."

func newBrokerRemoveCmd(app *App) *cobra.Command {
	c := withLong(dispatchLeaf(app, "remove", "Remove the broker, prompting for each layer",
		platformOps(opK8sRemoveBroker, opCtrRemoveBroker)), brokerRemoveLong)
	addRemoveFlags(c, app, &layerData)
	return c
}

const brokerGenerateLong = "Renders exactly what `broker deploy` would apply, without applying it.\n\n" +
	"Kubernetes output can be piped straight to `kubectl apply -f -`; prefer\n" +
	"-o/--out to `>` redirection, which PowerShell corrupts."

func newBrokerGenerateCmd(app *App) *cobra.Command {
	c := wireExec(app, renderOnly(&cobra.Command{
		Use:               "generate",
		Short:             "Render what deploy would apply, without applying it",
		Long:              brokerGenerateLong,
		Args:              noRolePositional,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(_ *cobra.Command, _ []string) error {
			return dispatch(platformOps(opK8sGenBroker, opCtrGenArtifact), app)
		},
	}))
	addPodFlag(c, app, config.Docker, config.Podman)
	addOutFlags(c, app)
	return c
}

const brokerStartLong = "Starts what `broker deploy` already created; it does not create anything\n" +
	"itself.\n\n" +
	"Kubernetes starts primary, then backup, then monitor, waiting for each\n" +
	"rollout."

const brokerStopLong = "Scales the broker to zero replicas, or stops the container, without\n" +
	"deleting anything.\n\n" +
	"This takes messaging down until `broker start` brings it back."

func newBrokerStartCmd(app *App) *cobra.Command {
	return withLong(dispatchLeaf(app, "start", "Start a broker that is deployed but not running",
		platformOps(opK8sStartBroker, opCtrStartBroker)), brokerStartLong)
}

func newBrokerStopCmd(app *App) *cobra.Command {
	return withLong(dispatchLeaf(app, "stop", "Stop a running broker without removing it",
		platformOps(opK8sStopBroker, opCtrStopBroker)), brokerStopLong)
}

const brokerRestartLong = "Restarts the broker's pod(s) gracefully (Kubernetes), or the container\n" +
	"(docker/podman).\n\n" +
	"With no --pod every pod is bounced in turn: monitor, then backup, then\n" +
	"primary.\n\n" +
	"This drops in-flight messaging on whatever it bounces."

func newBrokerRestartCmd(app *App) *cobra.Command {
	c := wireExec(app, &cobra.Command{
		Use:               "restart",
		Short:             "Bounce the broker's pod(s) or container",
		Long:              brokerRestartLong,
		Args:              noRolePositional,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(_ *cobra.Command, _ []string) error {
			// app.pod RAW, not through podRole: empty here means "every pod, rolling"
			// rather than "the primary", which is the one place the two readings differ.
			return dispatch(platformOps(
				func(a *App) error { return opK8sRestart(a, a.pod) },
				opCtrRestartBroker), app)
		},
	})
	// Nothing is removed, so no layer -- but the confirmation still needs its silencer.
	addRemoveFlags(c, app, nil)
	addPodFlag(c, app, config.K8s)
	return c
}

const brokerStatusLong = "Reports pod/container state, plus the operator's CR conditions on\n" +
	"Kubernetes.\n\n" +
	"--all is how to find a broker whose env file you have lost."

func newBrokerStatusCmd(app *App) *cobra.Command {
	c := wireExec(app, &cobra.Command{
		Use:               "status",
		Short:             "Report the broker's deployment status",
		Long:              brokerStatusLong,
		Args:              noRolePositional,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(_ *cobra.Command, _ []string) error {
			return dispatch(platformOps(withPodRole(opK8sStatusBroker), opCtrStatusBroker), app)
		},
	})
	c.Flags().BoolVar(&app.all, "all", false,
		"report every Solace broker found, not just the one this env file describes "+
			"(Kubernetes: across all namespaces; docker/podman: every Solace container on this host)")
	c.Flags().BoolVar(&app.detail, "detail", false,
		"include the static artifacts, not just the running ones (Kubernetes: secrets, configmaps "+
			"and PVCs; docker/podman: mounts, which is also where secrets appear)")
	addPodFlag(c, app, config.K8s)
	return c
}

const brokerLogsLong = "Reads the picked pod's log (Kubernetes) or this host's container's log\n" +
	"(docker/podman)."

func newBrokerLogsCmd(app *App) *cobra.Command {
	c := wireExec(app, &cobra.Command{
		Use:               "logs",
		Short:             "Read the broker's logs",
		Long:              brokerLogsLong,
		Args:              noRolePositional,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(_ *cobra.Command, _ []string) error {
			return dispatch(platformOps(withPodRole(opK8sLogs), opCtrLogs), app)
		},
	})
	addLogFlags(c, app)
	c.Flags().BoolVarP(&app.previous, "previous", "p", false,
		"read the PREVIOUS container's logs instead of the current one")
	flagOnlyOn(c, "previous", config.K8s)
	addPodFlag(c, app, config.K8s)
	return c
}

const brokerCLILong = "Interactive only; execs `cli -A` into the picked pod (Kubernetes) or this\n" +
	"host's container (docker/podman).\n\n" +
	"To run a script instead of typing at a prompt, use\n" +
	"`broker perform cli-script <file>`."

const brokerShellLong = "Interactive only; execs `bash` into the picked pod (Kubernetes) or this\n" +
	"host's container (docker/podman).\n\n" +
	"To run a script instead of typing at a prompt, use\n" +
	"`broker perform shell-script <file>`."

func newBrokerCLICmd(app *App) *cobra.Command {
	return newBrokerExecCmd(app, "cli", "Open an interactive Solace CLI in the broker",
		brokerCLILong, opK8sCLI, opCtrCLI)
}

func newBrokerShellCmd(app *App) *cobra.Command {
	return newBrokerExecCmd(app, "shell", "Open an interactive shell in the broker",
		brokerShellLong, opK8sShell, opCtrShell)
}

// newBrokerExecCmd builds the two interactive session commands. They differ only in
// which binary they exec, so the wiring -- pod resolution, the --pod flag, the
// no-positional rule -- has one definition; two copies is how the last pair drifted.
func newBrokerExecCmd(app *App, use, short, long string, k8sFn roleOpFunc, ctrFn opFunc) *cobra.Command {
	c := wireExec(app, &cobra.Command{
		Use:               use,
		Short:             short,
		Long:              long,
		Args:              noRolePositional,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(_ *cobra.Command, _ []string) error {
			return dispatch(platformOps(withPodRole(k8sFn), ctrFn), app)
		},
	})
	addPodFlag(c, app, config.K8s)
	return c
}

const brokerCopyLong = "Attempts every file in both directions and reports each one, rather than\n" +
	"stopping at the first failure; a partial failure still exits non-zero."

func newBrokerCopyCmd(app *App) *cobra.Command {
	c := group("copy", "Copy files to or from the broker", brokerCopyLong)

	from := wireExec(app, &cobra.Command{
		Use:   "from files...",
		Short: "Copy files from the broker to the host",
		Long: "Each file lands in the current directory under its base name. The base name is\n" +
			"split on both separators, so a remote path written with backslashes cannot\n" +
			"produce a local filename containing one.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return dispatch(platformOps(
				func(a *App) error { return opK8sCopyFrom(a, args) },
				func(a *App) error { return opCtrCopyFrom(a, args) }), app)
		},
	})
	addPodFlag(from, app, config.K8s)

	into := wireExec(app, &cobra.Command{
		Use:   "into files...",
		Short: "Copy files from the host into the broker",
		Long: "--dir is the destination inside the broker, defaulting to the exec's working\n" +
			"directory. The destination is not checked first: a copy into a path that does\n" +
			"not exist fails for that file and says so.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return dispatch(platformOps(
				func(a *App) error { return opK8sCopyInto(a, args) },
				func(a *App) error { return opCtrCopyInto(a, args) }), app)
		},
	})
	into.Flags().StringVar(&app.destDir, "dir", "", "destination directory inside the broker")
	registerFlagCompletion(into, "dir", completeDirs)
	addPodFlag(into, app, config.K8s)

	c.AddCommand(from, into)
	return c
}

// --- broker configure -------------------------------------------------------

const brokerConfigureLong = "Settings the env file describes, applied to a deployed broker over its own\n" +
	"CLI.\n\n" +
	"On a fresh HA broker, run in this order:\n\n" +
	"  broker perform assert-leader     config-sync leader FIRST\n" +
	"  broker configure server-certs\n" +
	"  broker configure domain-certs\n" +
	"  broker configure default-vpn     hardening\n" +
	"  broker configure default-users\n" +
	"  broker configure product-keys    last\n\n" +
	"  broker configure data-replication   a DR pair, when there is one\n"

func newBrokerConfigureCmd(app *App) *cobra.Command {
	c := group("configure", "Configure a deployed broker (certs, keys, hardening)", brokerConfigureLong)
	c.AddCommand(
		newConfigureServerCertsCmd(app),
		newConfigureDomainCertsCmd(app),
		newConfigureProductKeysCmd(app),
		newConfigureDefaultVPNCmd(app),
		newConfigureDefaultUsersCmd(app),
		newConfigureDataReplicationCmd(app),
	)
	return c
}

// configure data-replication carries the two-actives sentence because it is the ONE
// path in this feature that can cause it: the command is local-only by design, so it
// cannot see what the mate is doing, and run at the new-active site while the old one
// still holds a VPN active both are active until someone notices.
const configureDataReplicationLong = "Converges THIS broker to the env file's replication: block -- the mate\n" +
	"addresses, which message-VPNs replicate, and each one's active/standby role.\n\n" +
	"It never contacts the mate. Run it at BOTH sites with the same file.\n\n" +
	"Shuts down replication on any replicating VPN the file does not list, and can\n" +
	"leave a VPN active at both sites if the mate still holds it -- move a role with\n" +
	"`broker perform data-replication` instead."

func newConfigureDataReplicationCmd(app *App) *cobra.Command {
	c := withLong(dispatchLeaf(app, "data-replication",
		"Converge this broker to the replication: block",
		platformOps(opK8sConfigureReplication, opCtrConfigureReplication)), configureDataReplicationLong)
	addRemoveFlags(c, app, nil)
	addPodFlag(c, app, config.K8s)
	return c
}

const configureServerCertsLong = "Loads or updates the broker's TLS server certificate, over its own CLI or,\n" +
	"on Kubernetes with kubernetes.tlsServerSecret set, via the operator's Secret.\n\n" +
	"--remove takes TLS down immediately on every listener configured to present\n" +
	"a certificate."

func newConfigureServerCertsCmd(app *App) *cobra.Command {
	c := withLong(dispatchLeaf(app, "server-certs", "Load or update the TLS server certificate",
		platformOps(opK8sConfigServerCerts, opCtrConfigServerCerts)), configureServerCertsLong)
	addApplyRemoveFlags(c, app)
	// --remove takes TLS down, so it asks; this is the silencer that answers it. No
	// layer flag: there is no expensive second thing to keep or delete here.
	addRemoveFlags(c, app, nil)
	addPodFlag(c, app, config.K8s)
	return c
}

const configureDomainCertsLong = "Uploads the domain CA certificates broker.domainCerts describes and creates\n" +
	"a domain certificate authority for each.\n\n" +
	"--remove deletes those authorities from the broker; the certificate files\n" +
	"themselves stay."

func newConfigureDomainCertsCmd(app *App) *cobra.Command {
	c := withLong(dispatchLeaf(app, "domain-certs", "Load the configured domain CA certificates",
		platformOps(opK8sConfigDomainCerts, opCtrConfigDomainCerts)), configureDomainCertsLong)
	addApplyRemoveFlags(c, app)
	// --remove deletes the domain CAs from a live broker, so it asks; this is the
	// silencer that answers it. No layer flag: there is no expensive second thing to
	// keep or delete here.
	addRemoveFlags(c, app, nil)
	addPodFlag(c, app, config.K8s)
	return c
}

const configureProductKeysLong = "Applies broker.productKeys to the primary node (and backup, in HA); fails\n" +
	"loud rather than silently succeeding when none are configured.\n\n" +
	"--remove revokes those keys and can leave the broker unlicensed."

func newConfigureProductKeysCmd(app *App) *cobra.Command {
	c := withLong(dispatchLeaf(app, "product-keys", "Apply the configured product keys",
		platformOps(opK8sConfigProductKeys, opCtrConfigProductKeys)), configureProductKeysLong)
	addApplyRemoveFlags(c, app)
	// --remove can unlicense the broker, so it asks; this is the silencer that answers it.
	addRemoveFlags(c, app, nil)
	addPodFlag(c, app, config.K8s)
	return c
}

const configureDefaultVPNLong = "Shuts down the broker's default message-VPN.\n\n" +
	"This stops every client connection using it, and every service it fronts."

func newConfigureDefaultVPNCmd(app *App) *cobra.Command {
	c := withLong(dispatchLeaf(app, "default-vpn", "Shut down the default message-VPN",
		platformOps(opK8sConfigDefaultVPN, opCtrConfigDefaultVPN)), configureDefaultVPNLong)
	addDisableEnableFlags(c, app)
	// The disable direction stops every client connection using the VPN, so it asks;
	// this is the silencer that answers it. No layer flag: there is no expensive
	// second thing to keep or delete here.
	addRemoveFlags(c, app, nil)
	addPodFlag(c, app, config.K8s)
	return c
}

const configureDefaultUsersLong = "Shuts down the default client-username in every message-VPN found on the\n" +
	"broker.\n\n" +
	"This blocks any client still relying on the default username to connect."

func newConfigureDefaultUsersCmd(app *App) *cobra.Command {
	c := withLong(dispatchLeaf(app, "default-users", "Shut down the default client-usernames in all VPNs",
		platformOps(opK8sConfigDefaultUsers, opCtrConfigDefaultUsers)), configureDefaultUsersLong)
	addDisableEnableFlags(c, app)
	// The disable direction blocks any client still relying on the default username,
	// so it asks; this is the silencer that answers it. No layer flag: there is no
	// expensive second thing to keep or delete here.
	addRemoveFlags(c, app, nil)
	addPodFlag(c, app, config.K8s)
	return c
}

// --- broker perform ---------------------------------------------------------

const brokerPerformLong = "One-shot actions against a running broker; not settings the env file\n" +
	"describes.\n\n" +
	"  assert-leader        make this node the config-sync leader (HA)\n" +
	"  redundancy-test      exercise a real failover and fail back (HA, INVASIVE)\n" +
	"  gather-diagnostics   collect a support bundle into broker.hostDiagnosticDir\n" +
	"  semp-login-check     prove the admin credentials work over SEMP\n" +
	"  export-config        capture the broker's configuration as one artifact\n" +
	"  import-config <file> apply a captured configuration back (DESTRUCTIVE)\n" +
	"  cli-script <file>    run a Solace CLI script\n" +
	"  shell-script <file>  run a host shell script inside the broker\n" +
	"  data-replication     move replication roles across a DR pair (DESTRUCTIVE)\n"

func newBrokerPerformCmd(app *App) *cobra.Command {
	c := group("perform", "Run a one-shot action against the broker", brokerPerformLong)
	c.AddCommand(
		newPerformAssertLeaderCmd(app),
		newPerformRedundancyTestCmd(app),
		newPerformGatherDiagnosticsCmd(app),
		newPerformSempLoginCheckCmd(app),
		newPerformExportConfigCmd(app),
		newPerformImportConfigCmd(app),
		newPerformCLIScriptCmd(app),
		newPerformShellScriptCmd(app),
		newPerformDataReplicationCmd(app),
	)
	return c
}

// perform data-replication reaches the MATE, which configure data-replication never
// does -- moving a role is only safe as an ordered pair, and the ordering is what this
// command exists to guarantee.
const performDataReplicationLong = "Moves each message-VPN's replication role to the site the env file names,\n" +
	"across BOTH brokers: demote at the losing site, confirm it, then promote.\n\n" +
	"It changes no configuration -- run `broker configure data-replication` first.\n" +
	"Refuses unless both sites are on their primary HA node and name each other.\n\n" +
	"Interrupts message flow for every VPN it moves."

func newPerformDataReplicationCmd(app *App) *cobra.Command {
	c := withLong(dispatchLeaf(app, "data-replication",
		"Move replication roles across the DR pair",
		platformOps(opK8sPerformReplication, opCtrPerformReplication)), performDataReplicationLong)
	addRemoveFlags(c, app, nil)
	addPodFlag(c, app, config.K8s)
	return c
}

const performAssertLeaderLong = "Asserts this node as the config-sync leader for the router and every\n" +
	"message-VPN; a no-op, not an error, on a standalone deployment.\n\n" +
	"Run this first on a fresh HA broker, before any `broker configure` step."

func newPerformAssertLeaderCmd(app *App) *cobra.Command {
	c := withLong(rolePositionalTeachesPod(dispatchLeaf(app, "assert-leader",
		"Assert the config-sync leader (HA only)",
		platformOps(opK8sConfigLeader, opCtrConfigLeader))), performAssertLeaderLong)
	addPodFlag(c, app, config.Docker, config.Podman)
	return c
}

const performRedundancyTestLong = "Fails the broker over to its mate and back for real; a no-op on a\n" +
	"standalone deployment.\n\n" +
	"A Ctrl-C partway through can leave the group failed over; see\n" +
	"docs/operations.md to recover."

func newPerformRedundancyTestCmd(app *App) *cobra.Command {
	c := withLong(rolePositionalTeachesPod(dispatchLeaf(app, "redundancy-test",
		"Exercise a real failover and fail back (HA only)",
		platformOps(opK8sVerifyRedundancy, opCtrVerifyRedundancy))), performRedundancyTestLong)
	addPodFlag(c, app, config.Docker, config.Podman)
	return c
}

const performGatherDiagnosticsLong = "Runs the broker's full diagnostic sweep and downloads the resulting bundle\n" +
	"to broker.hostDiagnosticDir.\n\n" +
	"Kubernetes collects one bundle per role by default; --pod narrows to just\n" +
	"one."

func newPerformGatherDiagnosticsCmd(app *App) *cobra.Command {
	c := withLong(dispatchLeaf(app, "gather-diagnostics", "Gather a support bundle into broker.hostDiagnosticDir",
		platformOps(opK8sVerifyDiagnostics, opCtrVerifyDiagnostics)), performGatherDiagnosticsLong)
	c.Flags().IntVar(&app.days, "days", 1, "days of logs/diagnostics to gather")
	registerFlagCompletion(c, "days", cobra.NoFileCompletions)
	addPodFlag(c, app, config.K8s)
	return c
}

const performSempLoginCheckLong = "Runs an authenticated SEMP request from inside the broker and reports\n" +
	"whether the credentials were accepted.\n\n" +
	"Run this after rotating the admin password."

func newPerformSempLoginCheckCmd(app *App) *cobra.Command {
	c := withLong(rolePositionalTeachesPod(dispatchLeaf(app, "semp-login-check",
		"Test an authenticated SEMP request against the broker",
		platformOps(withPodRole(opK8sVerifyLogin), opCtrVerifyLogin))), performSempLoginCheckLong)
	addPodFlag(c, app, config.K8s)
	return c
}

const performExportConfigLong = "Captures the broker's configuration as a replayable Solace CLI script; it\n" +
	"changes nothing on the broker.\n\n" +
	"--vpn NAME (repeatable) narrows to those VPNs; --broker-only excludes them\n" +
	"instead -- naming both is refused."

func newPerformExportConfigCmd(app *App) *cobra.Command {
	c := withLong(dispatchLeaf(app, "export-config", "Capture the broker's configuration as one artifact",
		platformOps(withPodRole(opK8sExportConfig), opCtrExportConfig)), performExportConfigLong)
	addExportFlags(c, app)
	addOutFlags(c, app)
	addPodFlag(c, app, config.K8s)
	return c
}

const performImportConfigLong = "Applies a `broker perform export-config` artifact back to a running\n" +
	"broker. Only that command's own output is accepted; to run any other script,\n" +
	"use `broker perform cli-script`.\n\n" +
	"Any message-VPN in the artifact that already exists on the target is torn\n" +
	"down and rebuilt, destroying every message spooled in it."

func newPerformImportConfigCmd(app *App) *cobra.Command {
	c := wireExec(app, &cobra.Command{
		Use:   "import-config <file>",
		Short: "Apply a captured configuration back to the broker",
		Long:  performImportConfigLong,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return dispatch(platformOps(
				func(a *App) error { return opK8sImportConfig(a, args[0]) },
				func(a *App) error { return opCtrImportConfig(a, args[0]) }), app)
		},
	})
	addRemoveFlags(c, app, nil)
	addPodFlag(c, app, config.K8s)
	return c
}

const performCLIScriptLong = "Uploads a local Solace CLI script and runs it in the broker; a bare\n" +
	"filename resolves under broker.cliScriptsDir.\n\n" +
	"A rejected line does not stop the rest of the script, but the run is then\n" +
	"reported as a failure."

func newPerformCLIScriptCmd(app *App) *cobra.Command {
	c := wireExec(app, &cobra.Command{
		Use:   "cli-script <file>",
		Short: "Run a Solace CLI script in the broker",
		Long:  performCLIScriptLong,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return dispatch(platformOps(
				func(a *App) error { return opK8sExecCLI(a, args[0]) },
				func(a *App) error { return opCtrExecCLI(a, args[0]) }), app)
		},
	})
	addPodFlag(c, app, config.K8s)
	return c
}

const performShellScriptLong = "Uploads a local shell script and runs it with bash inside the broker, as\n" +
	"the broker's own user.\n\n" +
	"Bash reports one exit status for the whole run, and the full output is\n" +
	"shown."

func newPerformShellScriptCmd(app *App) *cobra.Command {
	c := wireExec(app, &cobra.Command{
		Use:   "shell-script <file>",
		Short: "Run a host shell script inside the broker",
		Long:  performShellScriptLong,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return dispatch(platformOps(
				func(a *App) error { return opK8sExecShell(a, args[0]) },
				func(a *App) error { return opCtrExecShell(a, args[0]) }), app)
		},
	})
	addPodFlag(c, app, config.K8s)
	return c
}

// --- operator ---------------------------------------------------------------

const operatorLong = "Installs and operates the operator, which is shared across every broker it\n" +
	"watches.\n\n" +
	"`operator deploy` adds this env file's namespaces to the watch list;\n" +
	"`operator remove` removes them, deleting the install only when nothing\n" +
	"else needs it."

func newOperatorCmd(app *App) *cobra.Command {
	c := onlyOn(group("operator", "Install and operate the cluster-scoped EventBroker Operator", operatorLong),
		config.K8s)
	c.AddCommand(
		newOperatorValidateCmd(app),
		newOperatorDeployCmd(app),
		newOperatorRemoveCmd(app),
		newOperatorGenerateCmd(app),
		newOperatorStartCmd(app),
		newOperatorStopCmd(app),
		newOperatorRestartCmd(app),
		newOperatorStatusCmd(app),
		newOperatorLogsCmd(app),
	)
	return c
}

const operatorValidateLong = "Reports the operator's install state, running version versus this env\n" +
	"file's, and the namespaces it watches.\n\n" +
	"Every value is read live from the cluster, since the operator is shared\n" +
	"and may already watch namespaces this env file never named."

func newOperatorValidateCmd(app *App) *cobra.Command {
	return onlyOn(withLong(leaf(app, "validate", "Check the operator's install state and watch scope",
		opK8sValidateOperator), operatorValidateLong), config.K8s)
}

const operatorDeployLong = "Installs the operator, or unions this env file's namespaces into an\n" +
	"already-running one's watch list.\n\n" +
	"Widening the watch to all namespaces, or installing an older version,\n" +
	"needs an interactive confirmation with no unattended override."

func newOperatorDeployCmd(app *App) *cobra.Command {
	return onlyOn(withLong(leaf(app, "deploy", "Install the operator, or add this env file's namespaces to it",
		opK8sOperatorDeploy), operatorDeployLong), config.K8s)
}

const operatorRemoveLong = "Removes this env file's namespaces from the operator's watch list,\n" +
	"deleting the operator itself only when nothing else is left watching.\n\n" +
	"--delete-crd additionally deletes every PubSubPlusEventBroker in the\n" +
	"cluster, and is refused while any broker resource still exists."

func newOperatorRemoveCmd(app *App) *cobra.Command {
	c := onlyOn(withLong(leaf(app, "remove", "Release this env file's claim on the operator",
		opK8sOperatorRemove), operatorRemoveLong), config.K8s)
	addRemoveFlags(c, app, &layerCRD)
	return c
}

const operatorGenerateLong = "Renders everything `operator deploy` would apply, in apply order, without\n" +
	"applying it.\n\n" +
	"The watch list shown is only this env file's; a real deploy unions it\n" +
	"with whatever the running operator already watches."

func newOperatorGenerateCmd(app *App) *cobra.Command {
	c := onlyOn(renderOnly(withLong(leaf(app, "generate", "Render the operator install bundle",
		opK8sGenOperator), operatorGenerateLong)), config.K8s)
	addOutFlags(c, app)
	return c
}

const operatorStartLong = "Scales the controller Deployment back to one replica and waits for the\n" +
	"rollout.\n\n" +
	"Reconciliation resumes for every namespace the operator watches."

const operatorStopLong = "Scales the controller Deployment to zero.\n\n" +
	"This freezes reconciliation for every broker in every watched namespace,\n" +
	"not just this env file's, until `operator start`."

func newOperatorStartCmd(app *App) *cobra.Command {
	return onlyOn(withLong(leaf(app, "start", "Start the operator's controller",
		opK8sOperatorStart), operatorStartLong), config.K8s)
}

func newOperatorStopCmd(app *App) *cobra.Command {
	c := onlyOn(withLong(leaf(app, "stop", "Stop the operator's controller, freezing reconciliation",
		opK8sOperatorStop), operatorStopLong), config.K8s)
	// Not a removal, but cluster-scoped and invisible once done, so it takes the same
	// confirmation contract -- and the same one silencer.
	addRemoveFlags(c, app, nil)
	return c
}

const operatorRestartLong = "Runs a rollout restart on the operator's controller Deployment; it does\n" +
	"not change what is installed.\n\n" +
	"This briefly interrupts reconciliation for every broker the operator\n" +
	"watches."

func newOperatorRestartCmd(app *App) *cobra.Command {
	return onlyOn(withLong(leaf(app, "restart", "Bounce the operator's controller",
		opK8sOperatorRestart), operatorRestartLong), config.K8s)
}

const operatorStatusLong = "Reads the controller Deployment and its pods to report the running\n" +
	"version and the namespaces it watches."

func newOperatorStatusCmd(app *App) *cobra.Command {
	c := onlyOn(withLong(leaf(app, "status", "Report the operator's controller status",
		opK8sStatusOperator), operatorStatusLong), config.K8s)
	c.Flags().BoolVar(&app.detail, "detail", false,
		"include the full description of the operator deployment")
	return c
}

const operatorLogsLong = "Runs `kubectl logs` against the controller Deployment, showing\n" +
	"reconciliation for every broker it watches, not just this env file's."

func newOperatorLogsCmd(app *App) *cobra.Command {
	c := onlyOn(withLong(leaf(app, "logs", "Read the operator's controller logs",
		opK8sOperatorLogs), operatorLogsLong), config.K8s)
	addLogFlags(c, app)
	return c
}

// --- validate ---------------------------------------------------------------

const validateLong = "Reports both the broker and operator halves of the env file together;\n" +
	"`broker validate` and `operator validate` are the same rows, scoped.\n\n" +
	"Read-only; a [WARN] (such as an operator not installed yet) does not\n" +
	"fail it, only a [FAIL] does."

func newValidateCmd(app *App) *cobra.Command {
	return withLong(dispatchLeaf(app, "validate", "Check the whole env file and what it needs",
		platformOps(opK8sValidate, opCtrCheck)), validateLong)
}

// --- shared flag wiring -----------------------------------------------------

// addLogFlags wires the log-reading flags shared by `broker logs` and
// `operator logs`. One definition, so the two cannot drift; --previous is
// deliberately NOT here, because only the broker has a previous container to read.
func addLogFlags(c *cobra.Command, app *App) {
	c.Flags().BoolVarP(&app.follow, "follow", "f", false, "keep streaming new lines instead of exiting")
	c.Flags().StringVar(&app.tail, "tail", "", "print only this many trailing lines (a count, or `all`)")
	c.Flags().StringVar(&app.since, "since", "", "read only lines newer than this duration (e.g. 30s, 5m, 2h)")
	c.Flags().BoolVar(&app.timestamps, "timestamps", false, "prefix each line with its timestamp")
	registerFlagCompletion(c, "tail", cobra.NoFileCompletions)
	registerFlagCompletion(c, "since", cobra.NoFileCompletions)
}

// logArgs turns the log flags into the passthrough tokens kubectl and the container
// engines both accept. One builder for both platforms: the spellings are identical, and
// a second copy is how one of them would quietly lose a flag.
//
// --tail and --since are emitted as two tokens rather than --flag=value because both
// engines accept that form and it keeps the value a separate argv element. --since is
// canonicalised through Go's duration parser first, so a free-form operator string never
// needs charset checking -- round-tripping through a trusted parser beats enumerating
// bad characters.
func logArgs(a *App) ([]string, error) {
	var args []string
	if a.follow {
		args = append(args, "-f")
	}
	if a.tail != "" {
		args = append(args, "--tail", a.tail)
	}
	if a.since != "" {
		d, err := config.CanonicalDuration("--since", a.since)
		if err != nil {
			return nil, asUsage(err)
		}
		args = append(args, "--since", d)
	}
	if a.timestamps {
		args = append(args, "--timestamps")
	}
	return args, nil
}

// podFlagUsage is --pod's help text. One flag, two readings, and the flag has to say
// both: flagOnlyOn appends its own "(kubernetes only)" / "(docker/podman only)" tail, so
// this must not name a platform itself.
const podFlagUsage = "role to act on: primary (p), backup (b) or monitor (m). " +
	"Kubernetes: which pod. Docker/podman: which node THIS host is, detected from this host's name or address when omitted"

// addPodFlag declares --pod and scopes it to the platforms where it means something.
//
// The platforms are a PARAMETER, not a constant, because the flag has two readings and
// each is meaningless on the other side. On kubernetes it picks a pod to act on; on
// docker and podman there is one container per host and the transport ignores the role
// entirely, so `broker cli --pod backup` there would open a session on this host's only
// container regardless. An accepted-and-ignored flag is worse than a refused one, so the
// targeting commands scope it to kubernetes and the four host-identity commands
// (deploy, generate, and the two HA actions) scope it to the container platforms.
//
// This replaced rejectRole, which existed only because those four took a [role]
// positional. One mechanism refuses by flag name on both sides now.
func addPodFlag(c *cobra.Command, app *App, platforms ...config.Platform) {
	c.Flags().StringVar(&app.pod, "pod", "", podFlagUsage)
	flagOnlyOn(c, "pod", platforms...)
	registerFlagCompletion(c, "pod", completeRoles)
}

// withPodRole adapts a role-taking handler to the platform dispatch map by resolving
// --pod first. An unset --pod is the primary (config.ParseRole's own default), and a word
// that is not a role is a usage error, not a runtime one.
func withPodRole(fn roleOpFunc) opFunc {
	return func(a *App) error {
		role, err := podRole(a)
		if err != nil {
			return err
		}
		return fn(a, role)
	}
}

// noRolePositional is the Args validator for every command that used to take a [role]
// positional and now takes --pod.
//
// A typed role gets a message naming the flag, because someone reaching for
// `broker cli backup` has the right idea and the wrong spelling. Anything else gets
// cobra's own unknown-command wording, so a genuine typo is not answered with advice
// about a flag it has nothing to do with.
//
// The message names both readings of --pod rather than the one for this platform: cobra
// validates arguments BEFORE PreRunE, so app.Platform is not settled yet.
// rolePositionalTeachesPod installs noRolePositional on a leaf that USED to take a [role]
// positional, wrapping inline the way onlyOn and renderOnly do.
//
// The leaf builders give a command cobra.NoArgs, which answers a typed role with a bare
// "unknown command" and leaves the reader nothing to act on. Every command that lost a
// positional has to keep teaching the flag that replaced it, or the migration is a dead end
// for exactly the people who knew the old spelling.
func rolePositionalTeachesPod(c *cobra.Command) *cobra.Command {
	c.Args = noRolePositional
	return c
}

func noRolePositional(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	if _, err := config.ParseRole(args[0]); err == nil {
		return usagef("%s does not take a role argument; use --pod %s instead "+
			"(on kubernetes --pod picks the pod to act on, on docker/podman it says which "+
			"node this host is)",
			cmd.CommandPath(), args[0])
	}
	return usagef("unknown command %q for %q", args[0], cmd.CommandPath())
}

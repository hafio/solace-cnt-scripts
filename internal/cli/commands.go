package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"solace/internal/config"
)

// This file builds the whole command tree. Two rules shape it.
//
// ONE TREE, NOT ONE PER PLATFORM. The platform is a property of the deployment the
// env file already describes, so it is resolved from that file (or named with
// --platform) rather than typed again as the first word of every command
// (platform.go). The tree is the union of what the platforms can do and is the SAME
// shape on all of them, because help and completion render it without reading an
// env file. A command that does not apply says so in its help text and refuses at
// pre-run instead of disappearing.
//
// VERB THEN NOUN, AND NOTHING IMPLICIT. Every verb that acts on more than one kind
// of thing names the thing: `deploy broker`, `remove operator`, `status broker`. A
// verb with object children never acts on its own -- running it bare prints what it
// can act on. That costs one word and buys the property that no command does
// something you did not name, which matters most exactly where it is cheapest to
// get wrong: `remove` on its own removes nothing.
//
// The op bodies stay split across ops_k8s.go and ops_container.go, which is where
// the two platforms really are different things.

// addCommands hangs the lifecycle tree off root.
func addCommands(root *cobra.Command, app *App) {
	root.AddCommand(
		newCheckCmd(app),
		newSmokeCmd(app),
		newPrepareCmd(app),
		newDeployCmd(app),
		newConfigCmd(app),
		newStartCmd(app),
		newStopCmd(app),
		newRestartCmd(app),
		newStatusCmd(app),
		newLogsCmd(app),
		newCLICmd(app),
		newShellCmd(app),
		newCopyCmd(app),
		newGenerateCmd(app),
		newDiagnosticsCmd(app),
		newRemoveCmd(app),
	)
}

// groupAnnotation marks a verb that owns objects and executes nothing itself. It
// exists so the wiring tests can tell a group apart from a command that runs
// something: a group carries a RunE (see below) but must NOT carry the pre-run or
// --allow-command, because it never reaches an external command.
const groupAnnotation = "solace_group"

// group builds a verb that owns objects. Run bare it prints its own help -- the
// list of things the verb can act on -- which is the whole no-implicit-actions
// rule. Given a word it does not know, it fails loudly.
//
// That second half is why a group carries a RunE at all. Cobra answers a
// NON-runnable command by printing help to stdout and exiting 0 whatever arguments
// it was given, so `solace-util remove bogus` would report success having removed
// nothing, and a script would never notice. It is the same trap the completion
// command documents (completion.go), and it matters more here: these are the verbs
// that destroy things, so a mistyped noun must not look like it worked.
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
		return fmt.Errorf("%q is not something %s can act on; run `%s` to see what it can",
			args[0], cmd.CommandPath(), cmd.CommandPath())
	}
	return c
}

// --- check / smoke ----------------------------------------------------------

// newCheckCmd owns the read-only questions. Everything here can be run against a
// system you are not willing to disturb -- which is exactly why the failover
// exercise is NOT here but under `smoke`.
func newCheckCmd(app *App) *cobra.Command {
	c := group("check", "Run read-only checks",
		"Every check here is read-only: it reports and changes nothing.\n\n"+
			"  check deploy      before deploying -- config, cluster/engine reachability,\n"+
			"                    storage or DNS, and whether the operator is installed\n"+
			"  check semp-login  after deploying -- the broker answers an authenticated\n"+
			"                    SEMP request\n\n"+
			"The failover exercise is deliberately not here: it moves live traffic, so it\n"+
			"lives under `smoke` with the other invasive checks.")
	c.AddCommand(
		withLong(dispatchLeaf(app, "deploy", "Validate config and platform prerequisites before deploying",
			platformOps(opK8sCheck, opCtrCheck)), checkDeployLong),
		withLong(roleOnK8sLeaf(app, "semp-login", "Test an authenticated SEMP request against a running broker",
			opK8sVerifyLogin, opCtrVerifyLogin), checkSempLoginLong),
	)
	return c
}

const checkDeployLong = "On Kubernetes this probes cluster reachability first; an unreachable\n" +
	"cluster stops the check there, reporting permission, the operator, and\n" +
	"the StorageClass as skipped rather than run. Once the cluster answers,\n" +
	"permission to create the broker resource, whether the operator is\n" +
	"installed, and the StorageClass are each checked regardless of whether\n" +
	"another one failed -- a missing operator only warns, since `deploy\n" +
	"operator` installs it.\n" +
	"\n" +
	"On docker and podman it probes the runtime (docker also checks\n" +
	"compose), then the broker hostname(s) -- failing in HA on a miss but\n" +
	"only noting it standalone; an unreachable runtime stops the check\n" +
	"before DNS is tried."

const checkSempLoginLong = "You run curl against http://localhost:8080/SEMP/v2/monitor from inside\n" +
	"the broker itself -- kubectl exec on Kubernetes, docker/podman exec on a\n" +
	"container host -- with the credential on stdin, never in argv or a log.\n" +
	"\n" +
	"Kubernetes always logs in as the operator's fixed `admin` user, whatever\n" +
	"--pod you pick; docker and podman log in as whatever `admin.user` the\n" +
	"env file configures.\n" +
	"\n" +
	"A failed login reports the response's failing HTTP status line (e.g.\n" +
	"`401 Unauthorized`), and the command still exits non-zero."

// newSmokeCmd is where checks that DISTURB the broker live. The separation is the
// point: an operator scanning `check` should never find something that moves
// messaging traffic, and someone reaching for `smoke` has been told by the word
// itself that this is not a passive question.
func newSmokeCmd(app *App) *cobra.Command {
	c := group("smoke", "Run invasive checks that exercise the broker",
		"These checks prove the broker works by making it work, so they disturb it.\n"+
			"Read-only questions live under `check`.")
	c.AddCommand(withLong(roleOnContainerLeaf(app, "redundancy",
		"Exercise a real failover and fail back (HA only)",
		opK8sVerifyRedundancy, opCtrVerifyRedundancy), smokeRedundancyLong))
	return c
}

const smokeRedundancyLong = "On a standalone deployment this is a no-op: a [SKIP] line, not a failure --\n" +
	"there is no redundancy pair to exercise.\n" +
	"\n" +
	"It normally walks the group through release, un-release, then revert,\n" +
	"confirming `show redundancy` after each step -- skipping straight to the\n" +
	"revert if the Primary is not already active. On containers, the backup's\n" +
	"one mutation rides SEMP to nodes.backup.ip, preflighted before release.\n" +
	"\n" +
	"A run that dies partway (Ctrl-C included) can leave the group released or\n" +
	"failed over; re-run once you have fixed the cause."

// --- prepare ----------------------------------------------------------------

func newPrepareCmd(app *App) *cobra.Command {
	c := group("prepare", "Prepare the prerequisites a broker deployment needs",
		"Everything a broker needs to exist before it is deployed.\n\n"+
			"`prepare all` runs the steps that are needed every time and need no input --\n"+
			"the namespace and its secrets on Kubernetes, the host on docker and podman --\n"+
			"so it is safe to script. `deploy all` runs the same steps for you.\n\n"+
			"Two things are deliberately outside it. The operator is cluster-scoped and\n"+
			"shared between brokers, so it is installed and removed on its own\n"+
			"(`deploy operator`). And `prepare labels` cannot be scripted at all: the env\n"+
			"file names the label each broker role wants, but only you can say which\n"+
			"machine should carry it, so it prompts -- run it once when provisioning the\n"+
			"cluster, not on every deployment.")
	c.AddCommand(
		onlyOn(withLong(leaf(app, "namespace", "Create the broker namespace", opK8sPrepNamespace),
			prepareNamespaceLong), config.K8s),
		onlyOn(withLong(leaf(app, "secrets", "Create admin/monitor, TLS, and image-pull secrets", opK8sPrepSecrets),
			prepareSecretsLong), config.K8s),
		onlyOn(withLong(leaf(app, "labels",
			"Label cluster nodes for primary/backup/monitor placement (interactive, one-off)",
			opK8sPrepLabels), prepareLabelsLong), config.K8s),
		onlyOn(withLong(leaf(app, "host", "Create/own the data dir, verify DNS, generate the redundancy PSK", opCtrPrepHost),
			prepareHostLong), config.Docker, config.Podman),
		withLong(dispatchLeaf(app, "all", "Run every applicable prepare step, in order",
			platformOps(opK8sPrepAll, opCtrPrepAll)), prepareAllLong),
	)
	return c
}

const prepareNamespaceLong = "Applies a bare Namespace object named for `kubernetes.namespace` -- no\n" +
	"labels, quotas, or other content. `kubectl apply` makes this safe to\n" +
	"re-run: an existing namespace is left as it is.\n" +
	"\n" +
	"Run this before `prepare secrets`, since a Secret cannot be applied into\n" +
	"a namespace that does not exist yet; `prepare all` and `deploy all`\n" +
	"already run the two in that order."

const prepareHostLong = "Resolves the redundancy hostnames first (a miss fails loud in HA, but\n" +
	"standalone just continues), then logs in to the image registry when image.user\n" +
	"and image.pass are both set (a loud error if only one is). Outside standalone\n" +
	"mode it also generates nodes.psk when empty and writes it back to the env file\n" +
	"-- either way, copy the same value onto the other two hosts by hand.\n" +
	"\n" +
	"On rootless podman it first warns if this process is running as root (prep\n" +
	"must run as the target rootless user), then chowns through `podman unshare`\n" +
	"and checks this user's hard nofile limit against ulimits.nofile, stopping with\n" +
	"the exact fix when it falls short."

const prepareAllLong = "On Kubernetes this creates the namespace, then applies the admin/monitor\n" +
	"credentials secret plus -- if configured -- the TLS server and image-pull\n" +
	"secrets, via `kubectl apply`; idempotent, so a namespace failure stops\n" +
	"before any secret is touched.\n" +
	"\n" +
	"Docker and podman: identical to `prepare host` -- data directory, DNS\n" +
	"check, registry login when both image.user and image.pass are set (one\n" +
	"alone aborts the run), and the redundancy PSK (HA only) generated once;\n" +
	"once nodes.psk is set, later runs just remind you to keep it identical."

const prepareLabelsLong = "With no placement labels configured (kubernetes.placement.labels*) this is\n" +
	"a no-op that says so and exits cleanly. Otherwise it needs a real\n" +
	"terminal: there is no flag for the node choice, so a non-interactive run\n" +
	"refuses outright instead of failing deep inside the prompt.\n\n" +
	"Per configured role it lists the cluster's nodes, asks you to pick one,\n" +
	"then runs `kubectl label node --overwrite` for each configured key=value.\n" +
	"Kubernetes-managed prefixes (kubernetes.io/, k8s.io/, and similar) are\n" +
	"silently skipped, and a failed label is reported and skipped rather than\n" +
	"aborting the rest."

const prepareSecretsLong = "Always creates the admin/monitor secret. The TLS secret joins it only\n" +
	"when kubernetes.tlsServerSecret is set; the image-pull secret only when\n" +
	"kubernetes.imagePullSecret is set.\n\n" +
	"Applied with `kubectl apply` on stdin -- no secret value reaches an argv\n" +
	"or echoed command -- and the run is idempotent, safe to re-run after a\n" +
	"change.\n\n" +
	"The manifest names the namespace directly, so `prepare namespace` (or\n" +
	"`prepare all`) must run first. `generate secrets broker` renders the\n" +
	"same manifest without applying it, for review."

// --- deploy -----------------------------------------------------------------

func newDeployCmd(app *App) *cobra.Command {
	c := group("deploy", "Deploy the broker, the operator, or the whole broker stack",
		"`deploy broker` applies just the broker. `deploy all` runs the whole bring-up\n"+
			"for it: check -> prepare -> deploy -> assert the config-sync leader (HA).\n\n"+
			"Neither installs the operator. It is cluster-scoped and may already be serving\n"+
			"other brokers, so `deploy operator` is its own command -- run it once per\n"+
			"cluster. `check deploy` reports when it is missing.")

	brokerCmd := wireExec(app, &cobra.Command{
		Use:       "broker [role]",
		Short:     "Deploy the broker (containers: this host's container; in HA the role is detected from the hostname unless given)",
		Long:      deployBrokerLong,
		ValidArgs: config.RoleNames(),
		Args:      cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return deployBroker(app, firstArg(args))
		},
	})
	addRestartFlag(brokerCmd, app)
	flagOnlyOn(brokerCmd, "restart", config.Docker, config.Podman)

	allCmd := wireExec(app, &cobra.Command{
		Use:       "all [role]",
		Short:     "Orchestrate the whole bring-up for this broker",
		Long:      deployAllLong,
		ValidArgs: config.RoleNames(),
		Args:      cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			arg := firstArg(args)
			if err := rejectRole(app.Platform, arg, config.Docker, config.Podman); err != nil {
				return err
			}
			if app.Platform == config.K8s {
				return opK8sDeployAll(app)
			}
			role, err := containerRole(app, arg)
			if err != nil {
				return err
			}
			return opCtrDeployAll(app, role)
		},
	})
	addRestartFlag(allCmd, app)
	flagOnlyOn(allCmd, "restart", config.Docker, config.Podman)

	c.AddCommand(
		brokerCmd,
		onlyOn(withLong(leaf(app, "operator", "Install the cluster-scoped EventBroker Operator", opK8sOperatorDeploy),
			deployOperatorLong), config.K8s),
		allCmd,
	)
	return c
}

const deployBrokerLong = "Kubernetes: applies the PubSubPlusEventBroker CR via the configured\n" +
	"`kubernetes.runtime` (`kubectl` by default). It creates neither the namespace\n" +
	"nor the secrets -- run `prepare all` first, or use `deploy all` for both plus\n" +
	"the broker in one step.\n" +
	"\n" +
	"Docker and podman: `prepare host` must have generated the PSK first. Podman\n" +
	"loads every secret into its own store before writing the quadlet unit;\n" +
	"docker's compose file only names an env var, filled in at run time. The\n" +
	"artifact is then started, or -- if already running and changed -- restarted\n" +
	"with `--restart` or your confirmation. `[role]` picks which host this is and\n" +
	"is refused on Kubernetes."

const deployAllLong = "Kubernetes runs check -> create namespace -> create secrets -> apply the\n" +
	"broker CR, then asserts the config-sync leader when HA is enabled. Docker\n" +
	"and podman stop after check -> prepare host -> deploy -- the leader is\n" +
	"asserted separately, with `config leader` run on the primary once every\n" +
	"host is up.\n" +
	"\n" +
	"[role] (docker/podman only) tells this host which identity -- primary,\n" +
	"backup, or monitor -- to deploy its own container as; run it once per\n" +
	"host, each time with that host's own role. Omitting it silently\n" +
	"defaults to primary rather than detecting it from the hostname, unlike\n" +
	"`config leader`."

const deployOperatorLong = "After checking permission to create customresourcedefinitions, this\n" +
	"installs the operator in three applies: the bundle's own Namespace\n" +
	"document, then -- as a separate artifact, not part of the bundle -- the\n" +
	"image-pull secret when `kubernetes.imagePullSecret` is set, then the\n" +
	"rest of the bundle (CRDs, RBAC and the controller Deployment). That is\n" +
	"exactly what `generate operator` and `generate secrets operator` print.\n" +
	"\n" +
	"Re-running it rolls the operator to a newer `kubernetes.operator.image`;\n" +
	"pointing it at an OLDER image warns you and asks before applying\n" +
	"anything, and running it unattended refuses outright."

// deployBroker is shared by `deploy broker` and the container half of `deploy all`.
func deployBroker(app *App, arg string) error {
	if err := rejectRole(app.Platform, arg, config.Docker, config.Podman); err != nil {
		return err
	}
	if app.Platform == config.K8s {
		return opK8sDeploy(app)
	}
	role, err := containerRole(app, arg)
	if err != nil {
		return err
	}
	return opCtrDeploy(app, role)
}

// --- config -----------------------------------------------------------------

// newConfigCmd owns everything applied to a broker that is already running, over
// its own CLI. There is deliberately no run-everything step: these are not
// uniformly re-runnable -- `apply additional-users` fails outright on a user that
// already exists -- so the order is documented here rather than baked into a
// command that would stop halfway through on a second run.
func newConfigCmd(app *App) *cobra.Command {
	c := group("config", "Configure a DEPLOYED broker (certs, hardening, product keys)",
		"Post-deployment configuration: every step here talks to a broker that is already\n"+
			"deployed and running, over the Solace CLI. None of it is part of `deploy`.\n\n"+
			"There is no run-everything command, because these steps are not uniformly\n"+
			"re-runnable. The order that works on a fresh broker is:\n\n"+
			"  1. config leader                        (HA only; on containers, the primary)\n"+
			"  2. config apply server-cert             (when TLS is configured)\n"+
			"  3. config apply domain-certs            (when any are listed)\n"+
			"  4. config disable default-vpn\n"+
			"  5. config disable default-users\n"+
			"  6. config apply additional-users        (Kubernetes; after the hardening, so\n"+
			"                                           the sequence reads harden-then-provision.\n"+
			"                                           NOT re-runnable: the broker refuses to\n"+
			"                                           create a user that already exists)\n"+
			"  7. config apply product-keys            (when any are listed)\n\n"+
			"Only domain-certs can be undone from here (`config delete domain-certs`).\n"+
			"There is no un-harden, and no way to withdraw a server certificate or a\n"+
			"product key through this tool.")

	apply := group("apply", "Apply configuration to the running broker",
		"`domain-certs` and `product-keys` run identically on every platform;\n"+
			"`additional-users` is Kubernetes only, because docker and podman create\n"+
			"those CLI users at container boot from the mounted password file\n"+
			"instead of over the CLI here. `server-cert` does too, unless\n"+
			"`kubernetes.tlsServerSecret` names a Secret, in which case Kubernetes\n"+
			"rewrites that Secret directly with kubectl instead of running the\n"+
			"broker-CLI path containers always use.\n\n"+
			"The certificate bodies, product keys and passwords each of these\n"+
			"applies ride in on an upload or over stdin -- never as an argument on\n"+
			"this command's own command line.")
	apply.AddCommand(
		withLong(dispatchLeaf(app, "server-cert", "Load/update the TLS server certificate",
			platformOps(opK8sConfigServerCert, opCtrConfigServerCert)), configServerCertLong),
		withLong(dispatchLeaf(app, "domain-certs", "Load the configured domain CA certificates",
			platformOps(opK8sConfigDomainCerts, opCtrConfigDomainCerts)), configDomainCertsLong),
		withLong(dispatchLeaf(app, "product-keys", "Apply the configured product keys",
			platformOps(opK8sConfigProductKeys, opCtrConfigProductKeys)), configProductKeysLong),
		// Containers create these at boot from the mounted password file, so there is
		// nothing to apply afterwards; on Kubernetes the operator ignores extra keys in
		// the credentials Secret, so they are created here over the broker CLI.
		onlyOn(withLong(leaf(app, "additional-users", "Create the admin.additionalUsers CLI users (not re-runnable)",
			opK8sConfigAdditionalUsers), configAdditionalUsersLong), config.K8s),
	)

	del := group("delete", "Remove configuration from the running broker",
		"Only domain certificates can be withdrawn this way. A server certificate, the\n"+
			"default-VPN hardening and an applied product key all stay applied.")
	del.AddCommand(withLong(dispatchLeaf(app, "domain-certs", "Remove the configured domain CA certificates",
		platformOps(opK8sTeardownDomainCerts, opCtrTeardownDomainCerts)), configDeleteDomainCertsLong))

	disable := group("disable", "Shut down the broker's built-in defaults (hardening)",
		"Both steps are one-way: this tool has no command to re-enable what they shut down.")
	disable.AddCommand(
		withLong(dispatchLeaf(app, "default-vpn", "Shut down the default message-VPN",
			platformOps(opK8sConfigDisableVPN, opCtrConfigDisableVPN)), configDisableVPNLong),
		withLong(dispatchLeaf(app, "default-users", "Shut down the default client-usernames in all VPNs",
			platformOps(opK8sConfigDisableUsers, opCtrConfigDisableUsers)), configDisableUsersLong),
	)

	c.AddCommand(apply, del, disable,
		withLong(roleOnContainerLeaf(app, "leader", "Assert the config-sync leader (HA only)",
			opK8sConfigLeader, opCtrConfigLeader), configLeaderLong),
	)
	return c
}

const configServerCertLong = "On Kubernetes, when kubernetes.tlsServerSecret is set, this rebuilds and\n" +
	"re-applies that Secret in place -- the broker already has it mounted and\n" +
	"re-reads it, so no pod restart runs here. Otherwise, and always on\n" +
	"docker/podman, it uploads the certificate over the broker CLI: on\n" +
	"Kubernetes to every role in the deployment (primary alone, or primary,\n" +
	"backup and monitor together when redundancy is enabled); on docker/podman,\n" +
	"only to this host's one container.\n" +
	"\n" +
	"Either path needs tls.cert and tls.certKey set."

const configDomainCertsLong = "Reads broker.domainCerts.folder and broker.domainCerts.files from the env\n" +
	"file, uploads each listed certificate file, then loads all of them into\n" +
	"the broker with one CLI script. It targets the primary node only, even in\n" +
	"an HA deployment.\n" +
	"\n" +
	"With no domain CAs configured this is a safe no-op: it logs that nothing\n" +
	"is set and returns without touching the broker, rather than failing."

const configProductKeysLong = "Fails loud with \"no product keys configured\" if broker.productKeys is empty\n" +
	"in the env file -- there is nothing to self-skip here, unlike domain-certs.\n" +
	"\n" +
	"Applies to the primary node, plus the backup on a Kubernetes HA deployment\n" +
	"(the monitor is skipped -- it carries no message spool); docker and podman\n" +
	"have only the primary. Each key is validated before anything is uploaded,\n" +
	"and the CLI script that applied it is removed from the node afterward."

const configAdditionalUsersLong = "Uploads a generated CLI script and runs it on the primary node only:\n" +
	"management users replicate to the mates via config-sync, so one run is\n" +
	"enough for the whole deployment -- every node in the group when HA is\n" +
	"enabled, or the one primary on a standalone broker.\n" +
	"\n" +
	"A username that already exists makes the whole run fail rather than\n" +
	"reconcile, and the transcript is withheld either way since it echoes\n" +
	"the passwords. Delete the user on the broker, or drop it from the env\n" +
	"file, then re-run this command."

const configDeleteDomainCertsLong = "Runs `no ssl domain-certificate-authority <name>` for every CA still listed\n" +
	"under `broker.domainCerts` in the env file, then shows the broker's resulting\n" +
	"CA list -- drop a CA from the file first and this command no longer knows to\n" +
	"remove it; with none configured it makes no CLI call at all.\n" +
	"\n" +
	"The certificate files `config apply domain-certs` uploaded stay in place --\n" +
	"only the CA configuration entries come out. On Kubernetes this runs against\n" +
	"the primary pod only, same as `config apply domain-certs`."

const configDisableVPNLong = "Uploads and runs two CLI scripts on the primary node (or the local\n" +
	"container): one shuts down the VPN's authentication (basic and\n" +
	"client-certificate), its default client-username, and every service under\n" +
	"it (SMF, REST incoming, MQTT and its websocket variants, AMQP, and\n" +
	"web-transport), also turning off SSL-downgrade-to-plain-text; the other\n" +
	"lists the resulting VPNs. Both scripts are removed once that listing runs.\n" +
	"\n" +
	"It never reaches beyond that node -- on a Kubernetes deployment with\n" +
	"redundancy enabled, unlike `apply server-cert` (backup and monitor) and\n" +
	"`apply product-keys` (backup). Elsewhere (docker, podman, or standalone\n" +
	"Kubernetes) those two commands are primary-only as well."

const configDisableUsersLong = "You query every message-VPN on the primary node (`show message-vpn *`) and\n" +
	"shut down the `default` client-username inside each one found -- not just\n" +
	"a VPN literally named `default`. If no VPN names parse from that output,\n" +
	"you get a warning and the broker is left untouched rather than the command\n" +
	"failing.\n" +
	"\n" +
	"On docker and podman, \"the primary node\" is simply the host's one\n" +
	"container."

const configLeaderLong = "No-ops on a standalone deployment rather than erroring. In HA it first\n" +
	"reverts any released activity back home -- straight to the backup pod on\n" +
	"Kubernetes, over SEMP to the mate on containers (a warning, not a failure,\n" +
	"if the mate is unreachable) -- then waits for redundancy to report restored\n" +
	"on the primary and asserts the primary as config-sync leader for the router\n" +
	"and every message-VPN.\n" +
	"\n" +
	"In HA, containers refuse this on the backup or monitor host; on a\n" +
	"standalone deployment every host no-ops instead, since that check runs\n" +
	"first."

// --- start / stop / restart -------------------------------------------------

// A deployed broker can be stopped without being removed. Kubernetes expresses that
// by scaling the StatefulSet to zero and containers by stopping the container; both
// leave the deployment, its data and its configuration in place, so one pair of
// verbs covers them.

func newStartCmd(app *App) *cobra.Command {
	c := group("start", "Start a broker that is deployed but not running",
		"The counterpart to `stop broker`: the deploy artifact and the persistent data\n"+
			"are already on disk, so this only brings the broker back up. It applies\n"+
			"nothing new -- a changed artifact needs `deploy broker` first.\n"+
			"\n"+
			"There is nothing to start until a broker has been deployed; on a host or\n"+
			"cluster that has never had one, `deploy broker` is the command you want.")
	c.AddCommand(withLong(dispatchLeaf(app, "broker",
		"Start the broker (Kubernetes: scale the statefulset(s) to 1; containers: start the container)",
		platformOps(opK8sStartBroker, opCtrStartBroker)), startBrokerLong))
	return c
}

const startBrokerLong = "On kubernetes, a redundancy group comes up primary, backup, monitor -- the\n" +
	"reverse of `restart broker`'s order -- waiting for each statefulset to report\n" +
	"ready before scaling the next. Standalone scales just the primary.\n" +
	"\n" +
	"Docker runs `compose start` against the deployed compose file when one\n" +
	"exists, falling back to a plain `docker start` on the container itself\n" +
	"when it does not. Podman always starts the quadlet's systemd unit instead,\n" +
	"since quadlet owns the container once deployed. Neither path creates a\n" +
	"broker that was never deployed."

func newStopCmd(app *App) *cobra.Command {
	c := group("stop", "Stop a running broker without removing it",
		"The deployment, its persistent data and its configuration all survive --\n"+
			"`start broker` brings it back. Use `remove broker` to delete it.")
	c.AddCommand(withLong(dispatchLeaf(app, "broker",
		"Stop the broker (Kubernetes: scale the statefulset(s) to 0; containers: stop the container)",
		platformOps(opK8sStopBroker, opCtrStopBroker)), stopBrokerLong))
	return c
}

const stopBrokerLong = "On Kubernetes every role's StatefulSet scales to zero together in one\n" +
	"command (all three in HA; just the primary standalone) -- there is no\n" +
	"role argument. It returns as soon as the scale succeeds; unlike\n" +
	"`start broker`, it does not wait for the pods to actually terminate.\n" +
	"\n" +
	"Unlike `restart broker` and every removal, this does not ask for\n" +
	"confirmation before acting."

func newRestartCmd(app *App) *cobra.Command {
	c := group("restart", "Bounce a running broker or the operator",
		"Restarting applies nothing new. A changed deploy artifact needs\n"+
			"`deploy broker` (containers: with --restart), which rewrites it first.")
	brokerCmd := wireExec(app, &cobra.Command{
		Use:               "broker",
		Short:             "Restart the broker (Kubernetes: delete pods so the statefulset recreates them)",
		Long:              restartBrokerLong,
		Args:              noRolePositional,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(*cobra.Command, []string) error {
			if app.Platform == config.K8s {
				// app.pod is read RAW here, not through podRole: an unset --pod
				// means "every pod, rolling", and podRole's empty-defaults-to-primary
				// would silently turn that into a single-pod restart. opK8sRestart
				// itself branches on the role argument being empty.
				return opK8sRestart(app, app.pod)
			}
			return opCtrRestartBroker(app)
		},
	})
	// Restarting confirms too: on Kubernetes it deletes pods, which drops messaging
	// traffic. It carries no --delete-* because it destroys nothing that survives
	// the bounce.
	addRemoveFlags(brokerCmd, app, nil)
	addPodFlag(brokerCmd, app)

	c.AddCommand(
		brokerCmd,
		onlyOn(withLong(leaf(app, "operator", "Restart the operator's controller deployment", opK8sOperatorRestart),
			restartOperatorLong), config.K8s),
	)
	return c
}

const restartOperatorLong = "Runs `kubectl rollout restart deployment` on the operator's own\n" +
	"controller deployment -- not the broker's statefulset -- in whatever\n" +
	"namespace it resolves to, which need not be the broker's own namespace.\n" +
	"\n" +
	"Unlike `restart broker`, this asks nothing first: the controller carries\n" +
	"no messaging traffic, so there is no traffic impact to confirm.\n" +
	"\n" +
	"It bounces the controller already installed; a changed bundle still\n" +
	"needs `deploy operator`, which re-applies the manifest."

const restartBrokerLong = "For kubernetes.updateStrategy=manualPodRestart: `deploy broker` updates the\n" +
	"statefulset's pod template but the operator waits for a pod to be deleted before\n" +
	"applying it.\n\n" +
	"With no --pod, every pod is restarted in the safe order (monitor, backup, primary;\n" +
	"standalone: just the primary), waiting for each to become ready before the next.\n" +
	"The order is by configured role, not by which node is currently active -- after a\n" +
	"failover they differ. Check `solace-util smoke redundancy` first, or pass --pod\n" +
	"and restart them one at a time.\n\n" +
	"On docker and podman there is one broker per host and no pod to pick: the\n" +
	"container is restarted in place."

// --- status / logs / cli / shell / copy --------------------------------------

func newStatusCmd(app *App) *cobra.Command {
	c := group("status", "Report on the broker or the operator",
		"By default this reports the RUNNING artifacts. --detail adds the static ones --\n"+
			"the full description of what is deployed, load balancer included.")

	brokerCmd := wireExec(app, &cobra.Command{
		Use:               "broker",
		Short:             "Show the broker's deployment status",
		Long:              statusBrokerLong,
		Args:              noRolePositional,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(*cobra.Command, []string) error {
			role, err := podRole(app)
			if err != nil {
				return err
			}
			if app.Platform == config.K8s {
				return opK8sStatusBroker(app, role)
			}
			return opCtrStatusBroker(app)
		},
	})
	brokerCmd.Flags().BoolVar(&app.all, "all", false,
		"report every Solace broker found, not just the one this env file describes "+
			"(Kubernetes: across all namespaces; docker/podman: every Solace container on this host)")
	brokerCmd.Flags().BoolVar(&app.detail, "detail", false,
		"include the static artifacts, not just the running ones (Kubernetes: secrets, "+
			"configmaps and PVCs; docker/podman: mounts, which is also where secrets appear)")
	addPodFlag(brokerCmd, app)

	operatorCmd := onlyOn(withLong(leaf(app, "operator", "Show the operator's controller status", opK8sStatusOperator),
		statusOperatorLong), config.K8s)
	operatorCmd.Flags().BoolVar(&app.detail, "detail", false,
		"include the full description of the operator deployment")

	c.AddCommand(brokerCmd, operatorCmd)
	return c
}

const statusBrokerLong = "Kubernetes reports the operator's CR conditions, then pods, services, and\n" +
	"statefulsets -- never the broker, so one unreachable over SEMP still reports\n" +
	"up. Docker and podman check the systemd unit or compose state, then `ps`.\n" +
	"\n" +
	"`--detail` and `--all` compose instead of replacing each other. Alone, it\n" +
	"describes the --pod pod (primary by default; --pod is Kubernetes-only) and\n" +
	"the load balancer Service, or runs `inspect` on containers (podman: plus the\n" +
	"installed unit). Under `--all` it instead adds an image column to the\n" +
	"survey, or inspects every container found."

const statusOperatorLong = "Reads the operator's controller Deployment (`pubsubplus-eventbroker-operator`)\n" +
	"in whichever namespace `kubernetes.operator.namespace` names, or else the\n" +
	"built-in default `pubsubplus-operator-system` that `deploy operator` installs\n" +
	"to -- the same two rules everywhere, so this reads the operator this env file\n" +
	"would deploy and never one belonging to someone else.\n" +
	"\n" +
	"Without --detail, a missing Deployment is reported as not installed\n" +
	"rather than failing the command; with --detail, the follow-up `kubectl\n" +
	"describe` fails on it instead. This checks the controller's own health,\n" +
	"not any broker's -- see `status broker` for that."

func newLogsCmd(app *App) *cobra.Command {
	c := group("logs", "Tail broker or operator logs",
		"On Kubernetes, `broker` and `operator` both run a single `kubectl logs` and\n"+
			"print whatever is already buffered there -- neither one follows the stream,\n"+
			"so the command exits as soon as that snapshot is printed.")
	c.AddCommand(
		withLong(roleOnK8sLeaf(app, "broker", "Tail the broker's logs", opK8sLogs, opCtrLogs), logsBrokerLong),
		onlyOn(withLong(leaf(app, "operator", "Tail the operator's controller logs", opK8sOperatorLogs),
			logsOperatorLong), config.K8s),
	)
	return c
}

const logsBrokerLong = "On Kubernetes this runs a single `kubectl logs` against the picked pod\n" +
	"(default: primary) and returns; it does not follow new lines. On docker and\n" +
	"podman it runs `<runtime> logs -f` against this host's one container\n" +
	"instead, which streams continuously until you interrupt it.\n" +
	"\n" +
	"--pod only applies on Kubernetes -- a container host has one broker per\n" +
	"machine, so --pod there is refused rather than accepted and silently\n" +
	"dropped."

const logsOperatorLong = "Runs `kubectl logs` against the operator's controller Deployment: one\n" +
	"snapshot of what it has already written, not a live follow, so watching a\n" +
	"rollout means calling this again rather than leaving it open.\n" +
	"\n" +
	"The namespace comes from kubernetes.operator.namespace when set, otherwise\n" +
	"the built-in default `deploy operator` installs to -- a missing or\n" +
	"unreachable operator just surfaces as kubectl's own error."

// newCLICmd opens a Solace CLI session in the broker, or -- with --input -- runs a
// script through one instead. Running a script is a flag rather than its own
// command because it answers the same question ("give me the broker's CLI") with
// the session automated rather than interactive.
func newCLICmd(app *App) *cobra.Command {
	c := wireExec(app, &cobra.Command{
		Use:   "cli",
		Short: "Open an interactive Solace CLI in the broker (Kubernetes: --pod picks the pod)",
		Long: "With no flags this opens an interactive Solace CLI session.\n\n" +
			"--input runs a script through that CLI instead of opening a session: a bare\n" +
			"filename is resolved under broker.cliScriptsFolder, a path is used as typed,\n" +
			"and the file is uploaded to the broker and run there. A CLI script is a\n" +
			"sequence of independent commands, so every line still runs even when the\n" +
			"broker rejects one -- but the command now fails at the end when any line\n" +
			"was rejected, instead of only warning.",
		Args:              noRolePositional,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(*cobra.Command, []string) error {
			if app.inputFile != "" {
				if app.Platform == config.K8s {
					return opK8sExecCLI(app, app.inputFile)
				}
				return opCtrExecCLI(app, app.inputFile)
			}
			if app.Platform == config.K8s {
				role, err := podRole(app)
				if err != nil {
					return err
				}
				return opK8sCLI(app, role)
			}
			return opCtrCLI(app)
		},
	})
	c.Flags().StringVarP(&app.inputFile, "input", "i", "",
		"run this Solace CLI script instead of opening an interactive session")
	addPodFlag(c, app)
	return c
}

func newShellCmd(app *App) *cobra.Command {
	return withLong(roleOnK8sLeaf(app, "shell", "Open an interactive shell in the broker", opK8sShell, opCtrShell),
		shellLong)
}

const shellLong = "Kubernetes execs `bash` into the --pod pod (`kubectl exec -it`, default\n" +
	"primary); docker and podman exec into this host's one broker container\n" +
	"instead (`<runtime> exec -it`). An explicit --pod there is refused\n" +
	"rather than accepted -- the same flag scoping `cli` uses -- since there\n" +
	"is only one container on the host to target.\n" +
	"\n" +
	"This is a plain OS shell inside the broker image, not the Solace CLI --\n" +
	"use `cli` for the broker's own CLI, or its `--input` to run a script\n" +
	"instead. Either way it targets the pod or container by name, so it must\n" +
	"already exist and be running."

// newCopyCmd mirrors the same verbs on every platform. On a container host the
// transport is node-local, so the files are already on this machine -- the verbs
// exist so a script does not have to know which platform it is driving.
func newCopyCmd(app *App) *cobra.Command {
	c := group("copy", "Copy files to/from the broker",
		"On Kubernetes, `--pod` (p|b|m, default primary) on `from`/`into` selects\n"+
			"the pod; docker and podman route to this host's single broker container\n"+
			"regardless, so the verbs exist only so a script need not know which\n"+
			"platform it is driving.\n\n"+
			"`from` lands each file under its basename in the current directory; `into`\n"+
			"defaults to `.` inside the pod or container -- the pod's login directory\n"+
			"on Kubernetes -- unless `--dir` names another one. Both attempt every\n"+
			"file and report per-file failures, exiting non-zero with a count instead\n"+
			"of stopping at the first bad path.")

	from := wireExec(app, &cobra.Command{
		Use:   "from files...",
		Short: "Copy files from the broker to the host",
		Long: "For example, `copy from /var/lib/solace/logs/debug.log` writes\n" +
			"`./debug.log` in your current directory -- `kubectl cp` on Kubernetes,\n" +
			"`<runtime> cp` on docker/podman.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if app.Platform == config.K8s {
				return opK8sCopyFrom(app, args)
			}
			return opCtrCopyFrom(app, args)
		},
	})
	addPodFlag(from, app)

	into := wireExec(app, &cobra.Command{
		Use:   "into files...",
		Short: "Copy files from the host into the broker",
		Long: "`into` never checks first that the pod or container exists, so copying\n" +
			"into an undeployed broker surfaces as a plain `cp` error.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if app.Platform == config.K8s {
				return opK8sCopyInto(app, args)
			}
			return opCtrCopyInto(app, args)
		},
	})
	addPodFlag(into, app)
	into.Flags().StringVar(&app.destDir, "dir", "", "destination directory inside the broker")
	registerFlagCompletion(into, "dir", completeDirs)

	c.AddCommand(from, into)
	return c
}

// podFlagUsage is shared by the commands that can target a specific pod.
const podFlagUsage = "pod role to target (p|b|m)"

// addPodFlag wires --pod identically everywhere it is declared: it is the ONE way
// to name a Kubernetes pod (H2), scoped to Kubernetes so passing it on a container
// platform is refused at pre-run (checkFlagPlatforms) rather than silently accepted
// and ignored -- the same job rejectRole used to do for the [role] positionals this
// flag replaced.
func addPodFlag(c *cobra.Command, app *App) {
	c.Flags().StringVar(&app.pod, "pod", "", podFlagUsage)
	flagOnlyOn(c, "pod", config.K8s)
	registerFlagCompletion(c, "pod", completeRoles)
}

// --- generate ---------------------------------------------------------------

// newGenerateCmd renders artifacts and applies nothing. It is the only way to see
// what a command WOULD send, which is why it is a command with a named target
// rather than a flag on the commands that deploy: an artifact you meant to read
// and a cluster you meant to change should not be one typo apart.
func newGenerateCmd(app *App) *cobra.Command {
	c := group("generate", "Render a deployment artifact to stdout without applying it",
		"Nothing here contacts the cluster or the container engine, so it is safe to run\n"+
			"against an env file you have not vetted.\n\n"+
			"The nouns are the same ones the acting verbs use: `generate broker` renders what\n"+
			"`deploy broker` would apply, whichever platform that is -- a custom resource on\n"+
			"Kubernetes, a compose file or systemd quadlet on a container host (which is\n"+
			"per-host, so it takes a [role] there).\n\n"+
			"Only the operator targets are platform-scoped, and because the thing does not\n"+
			"exist elsewhere rather than because it goes by another name: there is no\n"+
			"container operator to install.")

	brokerCmd := wireExec(app, renderOnly(&cobra.Command{
		Use:   "broker [role]",
		Short: "Render what `deploy broker` would apply",
		Long: "Kubernetes: the PubSubPlusEventBroker custom resource. Docker and podman: this\n" +
			"host's deploy artifact -- a compose file or a systemd quadlet unit -- which is\n" +
			"per-host, so [role] selects which node's artifact to render.",
		ValidArgs: config.RoleNames(),
		Args:      cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			arg := firstArg(args)
			if err := rejectRole(app.Platform, arg, config.Docker, config.Podman); err != nil {
				return err
			}
			if app.Platform == config.K8s {
				return opK8sGenBroker(app)
			}
			role, err := containerRenderRole(app, arg)
			if err != nil {
				return err
			}
			return opCtrGenArtifact(app, role)
		},
	}))

	// `secrets` names the thing it renders for the same reason the acting verbs do:
	// there are two sets of them and they are not interchangeable. The broker's
	// secrets are what `deploy broker` needs; the operator's single image-pull
	// secret is what `deploy operator` needs, and it is the one artifact of that
	// install carrying a credential -- the bundle only references it by name.
	secretsCmd := group("secrets", "Render a secret artifact",
		"`generate secrets broker` renders what the broker deployment needs -- Secret\n"+
			"manifests on Kubernetes, a shell script on a container host.\n\n"+
			"`generate secrets operator` renders the operator's image-pull secret, the one\n"+
			"part of the operator install that carries a credential. `generate operator`\n"+
			"renders the rest of that install and carries no secret value, so the two\n"+
			"together are exactly what `deploy operator` applies.")
	secretsCmd.AddCommand(
		renderOnly(withLong(dispatchLeaf(app, "broker",
			"Render the broker's secret-creation artifact (Kubernetes: Secret manifests; containers: a shell script)",
			platformOps(opK8sGenSecrets, opCtrGenSecrets)), genSecretsBrokerLong)),
		onlyOn(renderOnly(withLong(leaf(app, "operator", "Render the operator's image-pull secret",
			opK8sGenOperatorSecrets), genOperatorSecretsLong)), config.K8s),
	)

	c.AddCommand(
		brokerCmd,
		secretsCmd,
		onlyOn(renderOnly(withLong(leaf(app, "operator", "Render the operator install bundle", opK8sGenOperator),
			genOperatorLong)), config.K8s),
	)
	return c
}

const genOperatorLong = "`deploy operator` applies this bundle's namespace first, then the\n" +
	"image-pull secret `generate secrets operator` renders when one is\n" +
	"configured, then the rest -- applying that secret before its\n" +
	"namespace exists is what once failed a first install."

const genOperatorSecretsLong = "The rendered Secret is always named `regcred`, not the value of\n" +
	"`kubernetes.imagePullSecret` -- that field only gates whether this renders\n" +
	"at all (unset, it fails naming the setting), since the operator's own\n" +
	"manifests reference that literal name.\n\n" +
	"It resolves `kubernetes.operator.namespace`, falling back to the default\n" +
	"operator namespace when that is unset -- there is no live cluster here to\n" +
	"discover which one is actually running."

const genSecretsBrokerLong = "Kubernetes prints the same manifest `prepare secrets` applies -- admin secret\n" +
	"always, TLS and image-pull secrets only when kubernetes.tlsServerSecret /\n" +
	"kubernetes.imagePullSecret name one. Docker and podman print a shell script\n" +
	"instead: `secret create --replace` lines for podman, `export` lines for\n" +
	"docker to source before a manual `docker compose up` -- `deploy broker` sets\n" +
	"those variables itself, so the script is for a hand-run compose only.\n\n" +
	"Refuses to print an artifact for a secret whose value is still unset; on\n" +
	"containers, a missing nodes.psk points you at `prepare host` to generate it."

// --- diagnostics ------------------------------------------------------------

// newDiagnosticsCmd gathers a support bundle. It is not a check: it collects a
// large `show` sweep plus the broker's own diagnostics archive and downloads them,
// which is what Solace support asks for rather than something you read yourself.
func newDiagnosticsCmd(app *App) *cobra.Command {
	c := withLong(dispatchLeaf(app, "diagnostics",
		"Gather a support bundle from the broker into broker.diagDir",
		platformOps(opK8sVerifyDiagnostics, opCtrVerifyDiagnostics)), diagnosticsLong)
	c.Flags().IntVar(&app.days, "days", 1, "days of logs/diagnostics to gather")
	registerFlagCompletion(c, "days", cobra.NoFileCompletions)
	return c
}

const diagnosticsLong = "Runs the broker's full `show` command sweep plus `gather-diagnostics`, then\n" +
	"downloads the zipped output plus the diagnostics bundle into broker.diagDir\n" +
	"-- a failed download of the bundle is a warning, not a failed run.\n\n" +
	"Kubernetes gathers one archive per HA node (primary, backup, monitor when\n" +
	"redundancy is enabled); docker and podman only ever have this host's one\n" +
	"broker. Each run deletes the remote artifacts after downloading them."

// --- remove -----------------------------------------------------------------

// newRemoveCmd owns every deletion. Each object is named: `remove` on its own
// removes nothing, which is the whole reason the verb takes a noun.
//
// Both `remove broker` and `remove operator` keep their expensive layer by default
// -- persistent data and the CRDs respectively -- and ask about it the same way
// (addLayerFlags/confirmLayer in helpers.go), so learning the contract on one
// teaches the other. Either outcome is reported rather than left to be inferred.
func newRemoveCmd(app *App) *cobra.Command {
	c := group("remove", "Remove the broker, the operator, or the whole broker stack",
		"Every command here asks before it removes anything, and --no-prompt is the one\n"+
			"flag that makes it silent -- a script switches off one thing, not one per\n"+
			"question.\n\n"+
			"Nothing here removes the layer that is expensive to get back unless you say so:\n"+
			"the broker's persistent data and the operator's CRDs are kept by default, you\n"+
			"are asked about them separately, and what happened is printed either way. The\n"+
			"two flags compose, so an unattended removal that also drops the data is\n"+
			"`--delete-data --no-prompt`: naming the data you are willing to lose is not the\n"+
			"same as confirming the removal, so neither flag implies the other.\n\n"+
			"`remove all` takes this broker and its namespace. It leaves the operator, which\n"+
			"is cluster-scoped and may be serving brokers this env file does not describe.")

	brokerCmd := withLong(dispatchLeaf(app, "broker", "Remove the deployed broker",
		platformOps(opK8sDelete, opCtrDelete)), removeBrokerLong)
	addRemoveFlags(brokerCmd, app, &layerData)

	operatorCmd := onlyOn(withLong(leaf(app, "operator", "Remove the cluster-scoped EventBroker Operator",
		opK8sOperatorRemove), removeOperatorLong), config.K8s)
	addRemoveFlags(operatorCmd, app, &layerCRD)

	allCmd := withLong(dispatchLeaf(app, "all", "Remove the broker, its secrets and its namespace (the operator is kept)",
		platformOps(opK8sRemoveAll, opCtrRemoveAll)), removeAllLong)
	addRemoveFlags(allCmd, app, &layerData)

	// secrets and namespace have no retained layer -- `prepare` recreates both from
	// the env file -- but they still confirm, because deleting a namespace takes
	// whatever else happens to be in it.
	secretsCmd := onlyOn(withLong(leaf(app, "secrets", "Delete the broker's secrets", opK8sRemoveSecrets),
		removeSecretsLong), config.K8s)
	addRemoveFlags(secretsCmd, app, nil)
	namespaceCmd := onlyOn(withLong(leaf(app, "namespace", "Delete the broker's namespace", opK8sRemoveNamespace),
		removeNamespaceLong), config.K8s)
	addRemoveFlags(namespaceCmd, app, nil)

	c.AddCommand(brokerCmd, operatorCmd, secretsCmd, namespaceCmd, allCmd)
	return c
}

const removeBrokerLong = "On Kubernetes this deletes the PubSubPlusEventBroker custom resource with\n" +
	"`--ignore-not-found`, so running it again after a successful removal is a\n" +
	"no-op. On docker and podman it stops the broker container: docker via\n" +
	"`compose down` when a compose file is on disk (a plain stop and rm by\n" +
	"name otherwise) -- it never deletes the compose file; podman via\n" +
	"`systemctl stop` followed by deleting the `.container` quadlet unit\n" +
	"file and a daemon-reload.\n" +
	"\n" +
	"It keeps the broker's persistent data by default -- Kubernetes PVCs, or\n" +
	"the container's data directory -- and reports that either way; pass\n" +
	"`--delete-data` to remove it too. Beyond the broker and, if requested,\n" +
	"its data, this command touches nothing else: secrets, namespace, and\n" +
	"the operator all survive."

const removeOperatorLong = "Once confirmed, it deletes the operator's namespace, RBAC and controller\n" +
	"Deployment outright. The CustomResourceDefinitions are the one piece kept\n" +
	"by default, since deleting them cascades to every PubSubPlusEventBroker in\n" +
	"the cluster, not just the broker this env file describes.\n" +
	"\n" +
	"This never runs as part of `remove broker` or `remove all`: run it on its\n" +
	"own, once nothing else in the cluster still depends on it; `deploy\n" +
	"operator` reinstalls it afterward."

const removeSecretsLong = "Deletes the admin/monitor credentials secret, and -- if configured -- the\n" +
	"TLS server secret and the image-pull secret: the same three `prepare\n" +
	"secrets` creates. --ignore-not-found means an already-missing one is not\n" +
	"an error, but a real failure stops before the rest are touched.\n" +
	"\n" +
	"Removes only the Secret objects -- the namespace, the broker's other\n" +
	"resources, and the operator's CRDs are untouched, and `prepare secrets`\n" +
	"(or `prepare all`) recreates every one from the env file, which is why\n" +
	"there is no --delete-* flag here to keep one back."

const removeNamespaceLong = "Deletes the whole namespace, not just what your env file describes -- every\n" +
	"secret and PVC Kubernetes considers namespaced to it goes in the same\n" +
	"cascade.\n" +
	"\n" +
	"A repeat run is a no-op rather than an error: the underlying delete already\n" +
	"tolerates a missing namespace. You get the namespace back cheaply (`prepare\n" +
	"namespace` recreates it from the env file), but whatever you stored on its\n" +
	"volumes is gone for good."

const removeAllLong = "On Kubernetes this deletes the broker, then its secrets, then the\n" +
	"namespace -- exactly reversing the order `prepare all` creates them\n" +
	"in. The namespace delete also takes anything else that happens to\n" +
	"live in it, not just what this env file created.\n" +
	"\n" +
	"On docker and podman there is no separate namespace layer, and podman's\n" +
	"secret store is removed with the container rather than as its own step,\n" +
	"so this is exactly `remove broker`."

// --- shared leaf shapes ------------------------------------------------------

// noRolePositional replaces cobra.NoArgs on the commands whose [role] positional
// became --pod. cobra.NoArgs answers `shell backup` with `unknown command
// "backup"`, which is accurate and useless: that spelling was the DOCUMENTED one
// until this change, so the people most likely to type it are exactly the ones
// who learned it from the old help, and telling them the word is unknown hides
// the one thing they need -- that it moved.
//
// Only a word that really is a role gets the migration message; anything else
// keeps cobra's own wording, because a typo is not a migration.
//
// The message names both platforms rather than the one in play. Cobra validates
// args BEFORE PreRunE, and PreRunE is where platform.go resolves the platform
// from the env file, so at this point --platform may be unset and app.Platform
// not yet settled -- a message that guessed would be wrong exactly when the
// operator is already confused about where their command is going.
func noRolePositional(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return nil
	}
	if _, err := config.ParseRole(args[0]); err == nil {
		return fmt.Errorf("%q is a node role, and roles are no longer positional here: on Kubernetes "+
			"pick the pod with `%s --pod %s`; on docker and podman there is one broker per host, so no "+
			"role applies", args[0], cmd.CommandPath(), args[0])
	}
	return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
}

// roleOnK8sLeaf builds a leaf whose --pod picks a pod on Kubernetes and means
// nothing on a container host, where there is one broker per machine and the
// transport ignores the role entirely. --pod is scoped to Kubernetes
// (addPodFlag/flagOnlyOn), which is what refuses it on a container platform now --
// the job rejectRole used to do for the [role] positional this flag replaced.
func roleOnK8sLeaf(app *App, use, short string, k8sFn roleOpFunc, ctrFn opFunc) *cobra.Command {
	c := wireExec(app, &cobra.Command{
		Use:               use,
		Short:             short,
		Args:              noRolePositional,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(*cobra.Command, []string) error {
			if app.Platform == config.K8s {
				role, err := podRole(app)
				if err != nil {
					return err
				}
				return k8sFn(app, role)
			}
			return ctrFn(app)
		},
	})
	addPodFlag(c, app)
	return c
}

// roleOnContainerLeaf is the mirror image: the [role] says which node THIS
// machine is, which only a container host needs -- Kubernetes drives the whole
// redundancy group from one context, so a role there would be answering a
// question the cluster already knows. Both commands built this way (config
// leader, smoke redundancy) are primary-only on containers: they drive the
// whole group from the primary host, reaching the mate over SEMP.
func roleOnContainerLeaf(app *App, use, short string, k8sFn opFunc, ctrFn func(*App, string) error) *cobra.Command {
	return wireExec(app, &cobra.Command{
		Use:       use + " [role]",
		Short:     short + " (containers: run on the primary; [role] is this host, detected from its name when omitted)",
		ValidArgs: config.RoleNames(),
		Args:      cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			arg := firstArg(args)
			if err := rejectRole(app.Platform, arg, config.Docker, config.Podman); err != nil {
				return err
			}
			if app.Platform == config.K8s {
				return k8sFn(app)
			}
			return ctrFn(app, arg)
		},
	})
}

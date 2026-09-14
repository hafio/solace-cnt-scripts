package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"solace/internal/broker"
	"solace/internal/config"
	"solace/internal/k8s"
)

// The k8s handlers wire the cobra tree to the two Kubernetes entry types: k8s.Cluster
// (operations that talk to the cluster/operator, over engine.Runner) and broker.Ops
// (config/verify operations against a running broker, over the kubectl Transport).
// A test installs engine.Echo through App.NewRunner, so every handler is exercisable
// without a cluster. Progress goes to stderr via step; rendered artifacts are the only
// stdout (emit).

// k8sCluster builds a Cluster over the app's runner/config, wiring stdout as the report
// sink. It wires no reader: the package asks exactly one question, through Confirm.
func k8sCluster(a *App) *k8s.Cluster {
	c := k8s.NewCluster(a.Runner, a.Cfg, lineSink(), os.Stdout)
	// Same seam container.Manager carries: the one question this side asks is
	// whether to downgrade the cluster-scoped operator. It is the ONE question
	// --no-prompt cannot answer -- `deploy operator` does not register the flag,
	// and confirmDowngrade has no escape for it, so a non-interactive run always
	// declines. Rolling every broker in a cluster back is a decision made by a
	// human at a terminal or not at all.
	c.Confirm = func(question string) bool { return confirmDowngrade(a, question) }
	return c
}

// k8sOps builds a broker.Ops over the kubectl-exec transport for config/verify steps.
func k8sOps(a *App) *broker.Ops {
	return broker.New(k8s.NewTransport(a.Runner, a.Cfg), a.Cfg, lineSink())
}

// bg is the context for CLI-invoked operations. A plain background context matches the
// bash scripts' run-to-completion behavior; cancellation is a later concern.
func bg() context.Context { return context.Background() }

// today / nowStamp stamp certificate filenames and diagnostics archives, porting the
// `date +%F` / `date +%Y%m%d-%H%M%S` of 051/069.
func today() string    { return time.Now().Format("2006-01-02") }
func nowStamp() string { return time.Now().Format("20060102-150405") }

// lifecycle

// opK8sValidate / opK8sValidateBroker / opK8sValidateOperator are the three views of
// one report (k8s/checkreport.go). One builder behind all three is what makes the
// whole-file view provably equal to the two halves rather than a third thing that has to
// be kept in step with them.
//
// Every cluster-backed row is read from the cluster; the config echo is tagged [INFO] so
// it cannot be mistaken for something that was checked. A missing operator is a [WARN],
// not a failure: validating a cluster before installing anything is a normal thing to do.
func opK8sValidate(a *App) error         { return k8sCluster(a).Validate(bg()) }
func opK8sValidateBroker(a *App) error   { return k8sCluster(a).ValidateBroker(bg()) }
func opK8sValidateOperator(a *App) error { return k8sCluster(a).ValidateOperator(bg()) }

// opK8sDeploy brings the broker up: the prerequisites first, then the CR.
//
// The prerequisites are folded in rather than being their own `prepare` verb, because
// both are idempotent -- `kubectl apply` of a Namespace and of a multi-doc Secret
// manifest -- so running them on every deploy costs nothing and removes a step an
// operator could forget. `deploy operator` stays separate: the operator is
// cluster-scoped and shared between brokers, so installing it is not this broker's
// business.
//
// Node labelling used to be the third prerequisite and is gone entirely; the env file's
// placement labels are selectors the rendered CR carries, not labels this tool stamps
// onto anyone's cluster.
func opK8sDeploy(a *App) error {
	c := k8sCluster(a)
	ctx := bg()
	if err := c.CreateNamespace(ctx); err != nil {
		return err
	}
	if err := c.CreateSecrets(ctx); err != nil {
		return err
	}
	if err := c.DeployBroker(ctx, false); err != nil {
		return err
	}
	// Fire and forget, but not fire and hope: `apply` exiting 0 does not prove the
	// resource exists -- an admission webhook can reject it, and a CRD that is present
	// but not yet established fails differently -- so the CR is read back before this
	// reports success. Readiness is deliberately NOT waited on; `broker status` is for
	// watching it come up.
	return c.ConfirmBrokerApplied(ctx)
}

// The namespace and secrets steps are no longer commands of their own: `broker deploy`
// applies both, idempotently, and `broker remove` removes them. `prepare` existed only
// because they were separate, and a prerequisite you can forget to run is a prerequisite
// that gets forgotten.

// configure steps
//
// There is deliberately no run-everything step. Each of these talks to a live broker over
// its CLI and they are not uniformly re-runnable, so the order that works on a fresh
// broker is documented on the `configure` command rather than baked into a command that
// would stop halfway through on its second run.
//
// Each leaf reads its direction pair (--apply/--remove, or --disable/--enable) through
// wantRemove/wantEnable, which is also where both-at-once becomes a usage error. The
// default direction is the one an operator wants nine times out of ten, so it needs no
// flag at all.

// opK8sConfigServerCerts loads, updates or removes the TLS server certificate.
//
// The secret-managed path (kubernetes.tlsServerSecret set) rewrites the Secret and lets
// the operator mount it -- nothing is exec'd into the broker and no pod is restarted.
// Without it, the certificate is loaded over the broker CLI on every node that needs it.
//
// --remove is CLI-only and spans the same nodes the apply path does. It has no
// secret-managed form, and refuses rather than inventing one -- see the branch below.
func opK8sConfigServerCerts(a *App) error {
	remove, err := wantRemove(a)
	if err != nil {
		return err
	}
	if remove {
		// Secret-managed deployments have no CLI removal, and must not pretend to: the
		// operator mounts kubernetes.tlsServerSecret and would reconcile the certificate
		// straight back, so the command would report success over a broker that still
		// presents it. Say where the certificate actually comes from instead.
		if a.Cfg.K8s.TLSServerSecret != "" {
			return fmt.Errorf("the TLS server certificate comes from the Secret %q, which the operator "+
				"mounts -- removing it over the broker CLI would be undone at the next reconcile.\n"+
				"  To stop presenting it, clear kubernetes.tlsServerSecret (the CR's TLS block goes with it) "+
				"and redeploy; the Secret itself is removed by `broker remove` when this env file owns it",
				a.Cfg.K8s.TLSServerSecret)
		}
		// The whole group, like the apply path: a certificate gone from one node and
		// still loaded on another is a half state nobody asked for. --pod still narrows.
		roles, err := podRoles(a, k8s.HARoles(a.Cfg))
		if err != nil {
			return err
		}
		if !confirmAction(a, "Remove the TLS server certificate from", "remove the TLS server certificate from",
			k8sWhat(a, "broker "+a.Cfg.K8s.Name)) {
			return nil
		}
		return k8sOps(a).RemoveServerCerts(bg(), roles...)
	}
	if a.Cfg.K8s.TLSServerSecret != "" {
		// Whether this env file supplies the files to rebuild it is UpdateServerCertSecret's
		// question, not this one's: a Secret-backed deployment never wants the CLI path,
		// and being told the Secret is managed elsewhere beats being asked for a
		// certificate whose only use would be to overwrite someone else's.
		return k8sCluster(a).UpdateServerCertSecret(bg())
	}
	roles, err := podRoles(a, k8s.HARoles(a.Cfg))
	if err != nil {
		return err
	}
	return k8sOps(a).ServerCert(bg(), today(), roles...)
}

// opK8sConfigDomainCerts loads or removes the configured domain certificate authorities.
// --remove is implemented here, unlike its siblings, because
// `no ssl domain-certificate-authority` is a documented CLI form -- and, like every
// other removal in this tree, it asks before it acts.
//
// The directory walk (config.ResolveDomainCerts) runs here, at the op, rather
// than at Load: it is the first point where "the configured dirs must be
// readable" is actually true, so a deploy, generate or convert that never
// touches domain certificates does not fail on a machine that lacks the
// certificate directory. It runs for BOTH directions -- apply and --remove --
// so a directory-derived CA is not silently left on the broker by a removal
// that used to see only broker.domainCerts.files' explicit names.
func opK8sConfigDomainCerts(a *App) error {
	remove, err := wantRemove(a)
	if err != nil {
		return err
	}
	role, err := podRole(a)
	if err != nil {
		return err
	}
	certs, err := config.ResolveDomainCerts(a.Cfg.Broker.DomainCerts, nil)
	if err != nil {
		return err
	}
	if remove {
		if !confirmAction(a, "Remove", "remove",
			k8sWhat(a, "the configured domain CA certificates from broker "+a.Cfg.K8s.Name)) {
			return nil
		}
		return k8sOps(a).RemoveDomainCerts(bg(), role, domainCANames(certs))
	}
	return k8sOps(a).DomainCerts(bg(), role, certs)
}

// opK8sConfigProductKeys applies or revokes the configured product keys. The roles
// default to primary and, in HA, backup -- never the monitor, which carries no message
// spool -- and BOTH directions use that same set: a key revoked on the primary and left
// on the backup is a licensing state that changes at the next failover.
func opK8sConfigProductKeys(a *App) error {
	remove, err := wantRemove(a)
	if err != nil {
		return err
	}
	roles, err := podRoles(a, k8s.ProductKeyRoles(a.Cfg))
	if err != nil {
		return err
	}
	if remove {
		if !confirmAction(a, "Revoke every configured product key from", "revoke the product keys from",
			k8sWhat(a, "broker "+a.Cfg.K8s.Name)) {
			return nil
		}
		return k8sOps(a).RemoveProductKeys(bg(), a.Cfg.Broker.ProductKeys, roles...)
	}
	return k8sOps(a).ProductKeys(bg(), a.Cfg.Broker.ProductKeys, roles...)
}

// opK8sConfigDefaultVPN shuts the default message-VPN down, or starts it back up.
// One node: config-sync replicates it to the mates. Shutting it down stops every
// client connection using it, so that direction asks first, like every other
// disruptive `broker configure` leaf.
func opK8sConfigDefaultVPN(a *App) error {
	enable, err := wantEnable(a)
	if err != nil {
		return err
	}
	role, err := podRole(a)
	if err != nil {
		return err
	}
	if enable {
		return k8sOps(a).EnableDefaultVPN(bg(), role)
	}
	if !confirmAction(a, "Shut down", "shut down",
		k8sWhat(a, "the default message-VPN on broker "+a.Cfg.K8s.Name)) {
		return nil
	}
	return k8sOps(a).DisableDefaultVPN(bg(), role)
}

// opK8sConfigDefaultUsers shuts the "default" client-username down in every VPN, or
// starts it back up. One node, for the same config-sync reason. Shutting it down
// blocks any client still relying on that username, so that direction asks first.
func opK8sConfigDefaultUsers(a *App) error {
	enable, err := wantEnable(a)
	if err != nil {
		return err
	}
	role, err := podRole(a)
	if err != nil {
		return err
	}
	if enable {
		return k8sOps(a).EnableDefaultUsers(bg(), role)
	}
	if !confirmAction(a, "Shut down", "shut down",
		k8sWhat(a, "the default client-username in every message-VPN on broker "+a.Cfg.K8s.Name)) {
		return nil
	}
	return k8sOps(a).DisableDefaultUsers(bg(), role)
}

// opK8sConfigLeader asserts the config-sync leader. It drives the whole redundancy group,
// so it takes no --pod.
func opK8sConfigLeader(a *App) error { return k8sOps(a).Leader(bg()) }

// `config apply additional-users` is gone, and so is the broker-CLI op behind it.
// admin.additionalUsers is applied declaratively now: k8s.AdditionalUsersSecret builds a
// Secret of its own and the CR names it in spec.extraEnvVarsSecret, so the users exist from
// the broker's first boot rather than being created afterwards by a command that was not
// re-runnable. There is nothing left to wire here.

// opK8sExecCLI uploads and runs a local Solace CLI script in the target pod. A bare
// filename (no path separator) is resolved under the configured cliScripts folder; a
// path is used as-is. The interactive file-picker menu of the bash 059 is not ported.
// opK8sExportConfig captures the broker's configuration from the target pod.
//
// It takes a role rather than resolving one itself because it is wired through
// withPodRole: a capture reads one node, and on Kubernetes which node matters --
// config-sync makes the mates agree on VPN-level configuration but NOT on
// per-node settings, so a capture of the primary does not fully describe the
// backup.
func opK8sExportConfig(a *App, role config.Role) error {
	return runExport(a, k8sOps(a), role)
}

// opK8sImportConfig applies a captured configuration to the target pod.
//
// Unlike opK8sExecCLI the path is used exactly as given, with no
// broker.cliScriptsDir resolution. cliScriptsDir holds scripts an operator
// maintains; this file is an artifact export-config just wrote, so resolving it
// somewhere else would look for it where it is not.
func opK8sImportConfig(a *App, file string) error {
	role, err := podRole(a)
	if err != nil {
		return err
	}
	return runImport(a, k8sOps(a), role, file)
}

func opK8sExecCLI(a *App, file string) error {
	role, localPath, err := resolveScript(a, file, "CLI")
	if err != nil {
		return err
	}
	return k8sOps(a).ExecCLI(bg(), role, localPath)
}

// opK8sExecShell uploads and runs a local shell script inside the target pod.
//
// It resolves its filename exactly like opK8sExecCLI -- a bare name under
// broker.cliScriptsDir, a path as given -- so an operator who keeps both kinds of
// script in one place does not have to remember which command treats the folder
// differently. What it runs is arbitrary code inside the broker container; the command's
// help says so, and says that the output is shown in full.
func opK8sExecShell(a *App, file string) error {
	role, localPath, err := resolveScript(a, file, "shell")
	if err != nil {
		return err
	}
	return k8sOps(a).ExecShellScript(bg(), role, localPath)
}

// resolveScript is the filename-and-role resolution both script runners share: a bare
// filename resolves under broker.cliScriptsFolder, a path is used as given, and the pod
// comes from --pod. One definition, so the two runners cannot resolve the same argument
// to two different files.
func resolveScript(a *App, file, kind string) (config.Role, string, error) {
	if file == "" {
		return config.Primary, "", usagef("a %s script file is required", kind)
	}
	role, err := podRole(a)
	if err != nil {
		return config.Primary, "", err
	}
	localPath := file
	if !config.HasPathSeparator(file) {
		localPath = filepath.Join(a.Cfg.Broker.CLIScriptsDir, file)
	}
	return role, localPath, nil
}

// check / smoke steps

func opK8sVerifyRedundancy(a *App) error { return k8sOps(a).Redundancy(bg()) }

// opK8sVerifyDiagnostics gathers show-command output and a diagnostics bundle from every
// broker node into the configured diagnostics dir.
func opK8sVerifyDiagnostics(a *App) error {
	return k8sOps(a).Diagnostics(bg(), a.Cfg.Broker.HostDiagnosticDir, nowStamp(), a.days, k8s.HARoles(a.Cfg)...)
}

func opK8sVerifyLogin(a *App, role config.Role) error {
	return k8sLogin(a, k8sOps(a), role)
}

// k8sLogin tests a SEMP login as the fixed k8s admin user. Login writes the outcome to
// stdout and reports ok=false (not an error) on a failed login, so the handler turns a
// failed login into a non-zero exit.
func k8sLogin(a *App, o *broker.Ops, role config.Role) error {
	ok, err := o.Login(bg(), role, "admin", a.Cfg.SEMP.AdminPass)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("SEMP login failed on the %s node (see reason above)", role)
	}
	return nil
}

// day-2 ops

// opK8sStatusBroker reports on the broker. The two flags widen it along different
// axes and compose: --all trades this env file's one broker for every broker in the
// cluster, and --detail trades the running-artifact summary for the full
// description, load balancer included.
func opK8sStatusBroker(a *App, role config.Role) error {
	c := k8sCluster(a)
	ctx := bg()
	if a.all {
		return c.ClusterReport(ctx, a.detail)
	}
	if err := c.BrokerReport(ctx, a.detail); err != nil {
		return err
	}
	if !a.detail {
		return nil
	}
	// Scoped to one broker, --detail can afford to go further than the survey's
	// listing: the full description of the pod and the load balancer is what you
	// actually read when one broker is misbehaving.
	if err := c.DescribeBroker(ctx, role); err != nil {
		return err
	}
	return c.DescribeLB(ctx)
}

// opK8sStatusOperator reports the operator's controller state; --detail adds the
// full description of its Deployment.
func opK8sStatusOperator(a *App) error {
	c := k8sCluster(a)
	ctx := bg()
	if err := c.OperatorReport(ctx); err != nil {
		return err
	}
	if !a.detail {
		return nil
	}
	return c.OperatorDescribe(ctx)
}

func opK8sLogs(a *App, role config.Role) error {
	args, err := logArgs(a)
	if err != nil {
		return err
	}
	// --previous is kubernetes-only, so it is appended here rather than in logArgs,
	// which builds only the tokens every platform shares.
	if a.previous {
		args = append(args, "-p")
	}
	return k8sCluster(a).Logs(bg(), role, args)
}

// opK8sCLI / opK8sShell hand the terminal to a session in the pod, so the session's
// own exit status is this tool's (childExit, exit.go). `broker perform cli-script` is
// deliberately NOT marked: it runs a script through the CLI and reports on the
// script, which is this tool's own judgement rather than a status to pass along.
func opK8sCLI(a *App, role config.Role) error   { return childExit(k8sCluster(a).CLI(bg(), role)) }
func opK8sShell(a *App, role config.Role) error { return childExit(k8sCluster(a).Shell(bg(), role)) }

func opK8sCopyFrom(a *App, files []string) error {
	role, err := podRole(a)
	if err != nil {
		return err
	}
	return k8sCluster(a).CopyFrom(bg(), role, files)
}

func opK8sCopyInto(a *App, files []string) error {
	role, err := podRole(a)
	if err != nil {
		return err
	}
	return k8sCluster(a).CopyInto(bg(), role, files, a.destDir)
}

// opK8sStartBroker / opK8sStopBroker scale the broker statefulset(s) between 1 and
// 0. This is Kubernetes' version of stopping a deployed broker without deleting it:
// the StatefulSet, its PVCs and the CR all survive, which is exactly what
// `<runtime> stop` leaves behind on a container host.
func opK8sStartBroker(a *App) error { return k8sCluster(a).ReplicasStart(bg()) }
func opK8sStopBroker(a *App) error  { return k8sCluster(a).ReplicasStop(bg()) }

// opK8sRestart bounces broker pods for a manualPodRestart upgrade. Deleting a pod
// interrupts messaging on that node, so it takes the same confirmation a delete
// does: --no-prompt proceeds, an interactive session is asked, and a non-interactive
// one without it refuses rather than bouncing a production broker unattended.
func opK8sRestart(a *App, roleArg string) error {
	c := k8sCluster(a)
	if roleArg == "" {
		if !confirmAction(a, "Restart", "restart",
			k8sWhat(a, "every broker pod, one at a time (monitor, backup, primary),")) {
			return nil
		}
		return c.RestartRolling(bg())
	}
	role, err := config.ParseRole(roleArg)
	if err != nil {
		return asUsage(err)
	}
	if !confirmAction(a, "Restart", "restart", k8sWhat(a, "the "+roleWord(role)+" broker pod")) {
		return nil
	}
	return c.RestartPod(bg(), role)
}

// roleWord spells a role out for a prompt. It delegates to config.Role.Word, which is
// where the canonical spelling lives now that the env file keys on it too.
func roleWord(role config.Role) string { return role.Word() }

// operator lifecycle

// opK8sOperatorDeploy installs the operator; with a gen flag it prints the
// rendered bundle only. It backs both `operator deploy` and `prep operator`.
func opK8sOperatorDeploy(a *App) error {
	return k8sCluster(a).OperatorApply(bg())
}

// opK8sOperatorRemove releases THIS env file's claim on the operator.
//
// It may not remove anything at all. The operator is cluster-scoped and shared, so
// OperatorRelease first compares the namespaces it WATCHES against the ones this env file
// accounts for: covering all of them means the install is this env file's to delete, and
// covering only some means those namespaces come off the watch list and the operator stays
// up for the rest. An operator watching every namespace is kept with a warning, since
// "all except one" cannot be expressed.
//
// The layer question is the CRDs, whose deletion cascades to every PubSubPlusEventBroker
// in the cluster -- including brokers this env file has never heard of.
func opK8sOperatorRemove(a *App) error {
	if !confirmDelete(a, "the EventBroker operator in namespace "+
		k8s.OperatorNamespace(a.Cfg)+k8sContext(a)) {
		return nil
	}
	c := k8sCluster(a)
	deleteCRDs := a.deleteLayer
	if !deleteCRDs {
		// The CRD question is only worth asking when its answer is not already
		// fixed. A broker still on the cluster fixes it: OperatorDelete refuses
		// the CRD deletion outright there, so prompting would invite a "yes"
		// this tool will not honour -- the worst kind of prompt. Say what was
		// found instead, and keep the CRDs without asking.
		//
		// An explicit --delete-crd deliberately skips this and goes straight
		// through to that refusal: an operator who named the flag has earned a
		// loud failure naming the brokers in the way, not a silent downgrade to
		// "kept".
		// A listing failure is deliberately NOT fatal here. This call runs ahead of
		// OperatorDelete's own Preflight, so failing on it would replace the
		// preflight's actionable "you cannot delete X" with a confusing error about
		// a query the operator never asked for. Falling through to the question
		// costs nothing: refuseCRDDeleteIfBrokersExist re-asks authoritatively,
		// after the preflight, and is what actually stops the cascade.
		refs, err := c.BrokerCRs(bg())
		switch {
		case err != nil:
			deleteCRDs = confirmLayer(a, layerCRD)
		case len(refs) > 0:
			warn("%d broker resource(s) still exist, so the operator CRDs are kept without asking "+
				"(deleting them would cascade-delete every one): %s", len(refs), strings.Join(refs, ", "))
		default:
			deleteCRDs = confirmLayer(a, layerCRD)
		}
	}
	return c.OperatorRelease(bg(), deleteCRDs)
}

// opK8sOperatorStart / opK8sOperatorStop scale the controller Deployment.
//
// Stopping it freezes reconciliation for EVERY namespace the operator watches, not just
// this broker's, and nothing else reports that -- so it confirms first, through the same
// gate a removal uses. Starting it is additive and does not.
func opK8sOperatorStart(a *App) error { return k8sCluster(a).OperatorScale(bg(), 1) }

func opK8sOperatorStop(a *App) error {
	if !confirmAction(a, "Stop", "stop", "the operator, freezing reconciliation for every broker "+
		"it watches"+k8sContext(a)) {
		return nil
	}
	return k8sCluster(a).OperatorScale(bg(), 0)
}

func opK8sOperatorRestart(a *App) error { return k8sCluster(a).OperatorRestart(bg()) }
func opK8sOperatorLogs(a *App) error {
	args, err := logArgs(a)
	if err != nil {
		return err
	}
	return k8sCluster(a).OperatorLogs(bg(), args...)
}

// opK8sGenBroker / opK8sGenOperator print what deploy would apply and change nothing.
// Rendering is a command of its own rather than a flag on the command that would otherwise
// deploy: an artifact you meant to inspect and a cluster you meant to change should not be
// one typo apart.
// opK8sGenBroker renders the Secret manifests followed by the broker CR, joined as one
// multi-document stream in APPLY ORDER.
//
// One command, not the two it used to be. `generate secrets broker` existed so the
// credential-bearing half could be reviewed on its own, but the halves then had to be
// applied in the right order by hand, and the ordering is the part that is easy to get
// wrong -- the CR names Secrets that have to exist first. Emitting both in the order
// `broker deploy` applies them makes this output the artifact rather than a description of
// one.
//
// It therefore carries the admin password, the TLS private key and the registry
// credential in base64. That is what makes Kubernetes the only platform whose generate
// output is secret-bearing: a Secret manifest IS the artifact there, whereas a compose
// file or quadlet unit can only ever reference a secret the engine already holds.
func opK8sGenBroker(a *App) error {
	b, err := k8s.GenBroker(a.Cfg)
	if err != nil {
		return err
	}
	return emitOrWrite(a, b, "broker manifests")
}

// opK8sGenOperator renders the whole operator install stream -- namespace, image-pull
// secret when configured, then the bundle -- in the order OperatorApply applies it. The
// separate `generate secrets operator` is gone for the same reason its broker sibling is.
func opK8sGenOperator(a *App) error {
	b, err := k8s.GenOperator(a.Cfg)
	if err != nil {
		return err
	}
	return emitOrWrite(a, b, "operator install stream")
}

// remove

// opK8sRemoveBroker is the whole prompted teardown: the broker, its secrets, optionally
// its persistent data, and -- only if nothing else is left in it -- its namespace.
//
// The ORDER matters twice over. The CR goes before the PVCs because
// kubernetes.io/pvc-protection holds a claim while a pod still mounts it, so a PVC queued
// ahead of its workload waits forever. And the namespace question comes last, after
// everything this env file owns is gone, because that is the only point at which "is
// anything else in here?" has a meaningful answer.
//
// Keeping the data is what keeps the namespace: retained PVCs are occupancy, so a removal
// that keeps the data cannot then cascade it away by deleting the namespace. That hazard
// is structural rather than documented.
func opK8sRemoveBroker(a *App) error {
	if !confirmDelete(a, k8sWhat(a, "broker "+a.Cfg.K8s.Name)) {
		return nil
	}
	c := k8sCluster(a)
	ctx := bg()
	// The layer question is asked once, before any of the work, so a removal cannot get
	// halfway through and then stop to ask something.
	purge := confirmLayer(a, layerData)
	if err := c.DeleteBroker(ctx, purge); err != nil {
		return err
	}
	if err := c.DeleteSecrets(ctx); err != nil {
		return err
	}
	return removeNamespaceIfEmpty(a, c)
}

// removeNamespaceIfEmpty offers the namespace for deletion ONLY when nothing else is in
// it, and never deletes one Kubernetes owns.
//
// Deleting a namespace cascades to everything inside, including whatever another team put
// there, and this tool very often did not create it -- so the namespace is not in the
// removal's delete set at all. Anything still there is listed and the namespace kept, on
// every path including --no-prompt: a warning an operator can act on beats a prompt whose
// yes would destroy someone else's work.
//
// A check that cannot run keeps the namespace too. That direction is the safety property:
// a namespace wrongly reported empty gets cascade-deleted, while one wrongly reported
// occupied merely stays.
func removeNamespaceIfEmpty(a *App, c *k8s.Cluster) error {
	if c.NamespaceIsProtected() {
		warn("namespace %q is a Kubernetes namespace and is never removed by this tool",
			a.Cfg.K8s.Namespace)
		return nil
	}
	contents, err := c.NamespaceContents(bg())
	if err != nil {
		warn("%v", err)
		return nil
	}
	if len(contents) > 0 {
		c.ReportNamespaceOccupied(contents)
		return nil
	}
	if !confirmDelete(a, "the now-empty namespace "+a.Cfg.K8s.Namespace+k8sContext(a)) {
		return nil
	}
	return c.DeleteNamespace(bg())
}

// k8sWhat labels a destructive Kubernetes target the way containerWhat labels a
// container one: the object, then WHERE it is.
//
// Both facts matter, for different reasons. The namespace is in the env file but not in
// the prompt, so an operator with several env files open had nothing to check the
// question against. The context is not in the env file AT ALL -- the kubeconfig's current
// context decides which cluster every call lands in, so a file that says "dev" against a
// context that has drifted to prod reads identically. It is appended only when known
// (announceKubeContext resolved one), because a prompt that says "context " with nothing
// after it is worse than one that does not mention it.
func k8sWhat(a *App, object string) string {
	return fmt.Sprintf("%s in namespace %s%s", object, a.Cfg.K8s.Namespace, k8sContext(a))
}

// k8sContext is the trailing "(context X)" clause on its own, for the prompts that name
// their own location -- the namespace question says which namespace goes as part of its
// sentence, and the operator prompts are in the OPERATOR's namespace rather than the
// broker's. Appending k8sWhat's namespace to any of those would state a second, wrong or
// duplicate location.
func k8sContext(a *App) string {
	if a.kubeContext == "" {
		return ""
	}
	return fmt.Sprintf(" (context %s)", a.kubeContext)
}

// domainCANames returns the CA names from an already-resolved certificate set
// (config.ResolveDomainCerts, shared by the apply and --remove directions so
// neither sees a different set of names than the other); RemoveDomainCerts
// sorts and validates them.
func domainCANames(certs []config.DomainCert) []string {
	names := make([]string, 0, len(certs))
	for _, c := range certs {
		names = append(names, c.Name)
	}
	return names
}

// podRole resolves --pod to a single role, defaulting to the Primary when unset.
// config.ParseRole owns that default, so the empty string means primary everywhere a role
// is read -- except `broker restart`, which reads a.pod raw because empty there means
// "every pod, rolling".
func podRole(a *App) (config.Role, error) {
	role, err := config.ParseRole(a.pod)
	return role, asUsage(err)
}

// podRoles resolves --pod against the role set an operation would use by default.
//
// Unset --pod keeps the default set, which is what makes the fixed sets meaningful: a
// server certificate belongs on every node, product keys on primary and backup but never
// the monitor. Narrowing a set of more than one is warned about, because a partly-licensed
// or partly-certified redundancy group is almost always a mistake rather than a plan.
func podRoles(a *App, defaults []config.Role) ([]config.Role, error) {
	if a.pod == "" {
		return defaults, nil
	}
	role, err := podRole(a)
	if err != nil {
		return nil, err
	}
	if len(defaults) > 1 {
		warn("--pod %s narrows this to 1 of %d node(s) that would otherwise be covered",
			role, len(defaults))
	}
	return []config.Role{role}, nil
}

// The `deploy all` / `remove all` orchestration is gone with the third noun. `broker
// deploy` and `broker remove` now carry the whole per-broker lifecycle themselves, and the
// operator has its own noun -- so there is no longer a command whose job was to sequence
// two others, and no ambiguity about which of the three a given run touched.

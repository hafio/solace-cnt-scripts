package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"solace/internal/broker"
	"solace/internal/config"
	"solace/internal/container"
	"solace/internal/render"
)

// The container (docker/podman) handlers wire the cobra tree to the two entry
// types, mirroring the k8s handlers: container.Manager (host operations --
// prep/deploy/delete/status, over engine.Runner) and broker.Ops (config/verify
// against the running broker, over the node-local container Transport). a.Platform
// selects docker vs podman downstream. A test installs engine.Echo through
// App.NewRunner, so every handler is exercisable without a real engine.
//
// The container transport is node-local (one broker per host), so every broker.Ops
// call targets a single nominal role (config.Primary); the HA coordination that k8s
// drives cross-pod runs from the primary host instead, reaching the backup over the
// SEMP control channel (LeaderLocal/RedundancyCoordinated).
// Container config/verify reuse the shared kubernetes.* fields (DomainCerts, ProductKeys,
// DiagDir, CLIScriptsFolder) as the broker-ops config source -- no schema change.

// ctrOps builds a broker.Ops over the node-local container exec transport. The
// platform lets the SEMP mate channel resolve a bridge network's port mapping.
func ctrOps(a *App) *broker.Ops {
	o := broker.New(container.NewTransport(a.Runner, a.Cfg, a.Platform), a.Cfg, lineSink())
	o.Platform = a.Platform
	if a.Hostname != nil {
		// Left at broker.New's os.Hostname in production; only a test injects one,
		// so role detection can be exercised without depending on the machine
		// running the suite being named after a broker node.
		o.Hostname = a.Hostname
	}
	if a.LocalAddrs != nil {
		o.LocalAddrs = a.LocalAddrs
	}
	return o
}

// ctrManager builds a container host Manager, wiring stdout as the report sink and
// stdin as the prompt source. The env path is deliberately NOT passed: nothing in this
// tool writes to an env file, now that the redundancy pre-shared key is the operator's
// to supply rather than something a first deploy invented and wrote back.
func ctrManager(a *App) *container.Manager {
	m := container.NewManager(a.Runner, a.Cfg, a.Platform, lineSink(), os.Stdout)
	m.In = os.Stdin
	m.RestartApproved = a.restart
	m.Confirm = func(question string) bool { return confirmRestart(a, question) }
	return m
}

// lifecycle

func opCtrCheck(a *App) error { return ctrManager(a).Check(bg()) }

// opCtrDeployHost brings this host's broker up: the host prerequisites, then the
// container.
//
// PrepHost is folded in rather than being its own `broker deploy` verb, because every step
// it takes is idempotent and none of them has a side effect worth a separate command:
// mkdir -p, chown, a DNS check, a registry login. It used to have one -- generating the
// redundancy pre-shared key and writing it back into the env file on a first HA run -- and
// that is gone: the key is the operator's to make (`openssl rand -base64 32`, the same
// value on all three hosts), config refuses an env file without one, and nothing here
// edits the file it was handed.
func opCtrDeployHost(a *App) error {
	role, err := containerRole(a)
	if err != nil {
		return err
	}
	m := ctrManager(a)
	ctx := bg()
	if err := m.PrepHost(ctx); err != nil {
		return err
	}
	return m.Deploy(ctx, role)
}

// opCtrStartBroker / opCtrStopBroker act on a container that is already deployed:
// the compose file or quadlet unit and the data directory all survive, so a
// stopped broker starts again without redeploying. They are the container half of
// what scaling a StatefulSet to 1 or 0 does on Kubernetes.
func opCtrStartBroker(a *App) error { return ctrManager(a).Start(bg()) }
func opCtrStopBroker(a *App) error  { return ctrManager(a).Stop(bg()) }

// opCtrRestartBroker bounces the running container in place. A restart drops
// every open client connection and any in-flight messaging the same way a delete
// would, so it takes the same gate rather than running unconditionally -- which
// is what makes --no-prompt mean something on containers, matching opK8sRestart
// on Kubernetes. It asks with confirmAction's own verb ("Restart"), since the
// action is not a deletion and a prompt that says otherwise would be read as one.
func opCtrRestartBroker(a *App) error {
	if !confirmAction(a, "Restart", "restart", containerWhat(a)) {
		return nil
	}
	return ctrManager(a).Restart(bg())
}

// configure steps
//
// Node-local: one container per host, so the transport ignores the role and every one of
// these acts on this host's broker. They take the same direction pairs as their Kubernetes
// siblings and read them through the same helpers, so a flag means the same thing on every
// platform.
//
// There is no run-everything step, for the same reason there is none on Kubernetes.

// opCtrConfigServerCerts loads the TLS server certificate over the CLI. There is no
// secret-managed route on containers: the certificate reaches the broker as a mounted
// file, written by `broker deploy`, so re-loading it here is the only in-place update.
func opCtrConfigServerCerts(a *App) error {
	remove, err := wantRemove(a)
	if err != nil {
		return err
	}
	if remove {
		if !confirmAction(a, "Remove the TLS server certificate from", "remove the TLS server certificate from",
			containerWhat(a)) {
			return nil
		}
		return ctrOps(a).RemoveServerCerts(bg(), config.Primary)
	}
	return ctrOps(a).ServerCert(bg(), today(), config.Primary)
}

// opCtrConfigDomainCerts loads or removes the configured domain certificate
// authorities, like its k8s sibling; --remove asks before it acts.
func opCtrConfigDomainCerts(a *App) error {
	remove, err := wantRemove(a)
	if err != nil {
		return err
	}
	if remove {
		if !confirmAction(a, "Remove the configured domain CA certificates from",
			"remove the configured domain CA certificates from", containerWhat(a)) {
			return nil
		}
		return ctrOps(a).RemoveDomainCerts(bg(), config.Primary, domainCANames(a.Cfg))
	}
	return ctrOps(a).DomainCerts(bg(), config.Primary,
		a.Cfg.Broker.DomainCerts.Folder, a.Cfg.Broker.DomainCerts.Files)
}

func opCtrConfigProductKeys(a *App) error {
	remove, err := wantRemove(a)
	if err != nil {
		return err
	}
	if remove {
		if !confirmAction(a, "Revoke every configured product key from", "revoke the product keys from",
			containerWhat(a)) {
			return nil
		}
		return ctrOps(a).RemoveProductKeys(bg(), a.Cfg.Broker.ProductKeys, config.Primary)
	}
	return ctrOps(a).ProductKeys(bg(), a.Cfg.Broker.ProductKeys, config.Primary)
}

// opCtrConfigDefaultVPN shuts the default message-VPN down, or starts it back up,
// like its k8s sibling. Shutting it down stops every client connection using it, so
// that direction asks first.
func opCtrConfigDefaultVPN(a *App) error {
	enable, err := wantEnable(a)
	if err != nil {
		return err
	}
	if enable {
		return ctrOps(a).EnableDefaultVPN(bg(), config.Primary)
	}
	if !confirmAction(a, "Shut down the default message-VPN on",
		"shut down the default message-VPN on", containerWhat(a)) {
		return nil
	}
	return ctrOps(a).DisableDefaultVPN(bg(), config.Primary)
}

// opCtrConfigDefaultUsers shuts the "default" client-username down in every VPN, or
// starts it back up, like its k8s sibling. Shutting it down blocks any client still
// relying on that username, so that direction asks first.
func opCtrConfigDefaultUsers(a *App) error {
	enable, err := wantEnable(a)
	if err != nil {
		return err
	}
	if enable {
		return ctrOps(a).EnableDefaultUsers(bg(), config.Primary)
	}
	if !confirmAction(a, "Shut down the default client-username in every message-VPN on",
		"shut down the default client-username in every message-VPN on", containerWhat(a)) {
		return nil
	}
	return ctrOps(a).DisableDefaultUsers(bg(), config.Primary)
}

// opCtrConfigLeader asserts the config-sync leader from this host. It is primary-only and
// HA-only; --pod (empty -> detect from the hostname) lets an operator override detection.
// LeaderLocal fails loud on a backup or monitor host, and first reverts any released
// activity on the mate over SEMP, matching the k8s Leader order -- an unreachable mate
// downgrades that step to a warning, since this command's own job is local.
func opCtrConfigLeader(a *App) error { return ctrOps(a).LeaderLocal(bg(), a.pod) }

// opCtrExecCLI / opCtrExecShell upload and run a local script in the broker container.
// A bare filename resolves under broker.cliScriptsFolder; a path is used as given -- the
// same rule as the k8s handlers, through the same helper.
// opCtrExportConfig captures this host's broker configuration.
//
// config.Primary is nominal: the container transport ignores the role, because
// there is one broker per host. Which is also why there is no --pod here -- an
// accepted-and-ignored flag is worse than a refused one.
func opCtrExportConfig(a *App) error {
	return runExport(a, ctrOps(a), config.Primary)
}

// opCtrImportConfig applies a captured configuration to this host's broker. The
// path is used as given, for the reason opK8sImportConfig records.
func opCtrImportConfig(a *App, file string) error {
	return runImport(a, ctrOps(a), config.Primary, file)
}

func opCtrExecCLI(a *App, file string) error {
	localPath, err := resolveContainerScript(a, file, "CLI")
	if err != nil {
		return err
	}
	return ctrOps(a).ExecCLI(bg(), config.Primary, localPath)
}

func opCtrExecShell(a *App, file string) error {
	localPath, err := resolveContainerScript(a, file, "shell")
	if err != nil {
		return err
	}
	return ctrOps(a).ExecShellScript(bg(), config.Primary, localPath)
}

// resolveContainerScript is resolveScript without the pod: there is one container per
// host, so there is no role to resolve.
func resolveContainerScript(a *App, file, kind string) (string, error) {
	if file == "" {
		return "", usagef("a %s script file is required", kind)
	}
	if config.HasPathSeparator(file) {
		return file, nil
	}
	return filepath.Join(a.Cfg.Broker.CLIScriptsFolder, file), nil
}

// check / smoke steps

func opCtrVerifyLogin(a *App) error { return ctrLogin(a, ctrOps(a)) }

// opCtrVerifyRedundancy exercises failover for the whole redundancy group from
// this host, which must be the primary (backup/monitor are rejected loud; the
// role arg, empty -> detect, is self-identification only). The coordinating
// broker session the old two-host handshake left as a placeholder: the local
// half runs over the node-local transport, and the backup's one command rides
// SEMP v1 to redundancy.backup.addr, preflighted before anything is released.
// opCtrVerifyRedundancy drives a real failover from THIS host. Primary-only and HA-only;
// --pod (empty -> detect from the hostname) overrides detection. The mate's SEMP is
// preflighted before anything is released, so a firewalled mate aborts while the
// redundancy group is still undisturbed.
func opCtrVerifyRedundancy(a *App) error {
	return ctrOps(a).RedundancyCoordinated(bg(), a.pod)
}

// opCtrVerifyDiagnostics gathers show-command output and a diagnostics bundle from
// this host's broker into the configured diagnostics dir.
func opCtrVerifyDiagnostics(a *App) error {
	return ctrOps(a).Diagnostics(bg(), a.Cfg.Broker.DiagDir, nowStamp(), a.days, config.Primary)
}

// ctrLogin tests a SEMP login as the configured admin user against this host's
// broker. Login reports ok=false (not an error) on a failed login, so the handler
// turns a failed login into a non-zero exit.
func ctrLogin(a *App, o *broker.Ops) error {
	ok, err := o.Login(bg(), config.Primary, config.AdminUser, a.Cfg.SEMP.AdminPass)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("SEMP login failed on this host's broker (see reason above)")
	}
	return nil
}

// day-2 ops (node-local)
func opCtrStatus(a *App) error { return ctrManager(a).Status(bg()) }
func opCtrLogs(a *App) error {
	args, err := logArgs(a)
	if err != nil {
		return err
	}
	return ctrManager(a).Logs(bg(), args...)
}

// opCtrCLI / opCtrShell hand the terminal over, so the session's own exit status is
// this tool's -- the container half of what opK8sCLI/opK8sShell document.
func opCtrCLI(a *App) error      { return childExit(ctrManager(a).CLI(bg())) }
func opCtrShell(a *App) error    { return childExit(ctrManager(a).Shell(bg())) }
func opCtrDescribe(a *App) error { return ctrManager(a).Describe(bg()) }

// opCtrStatusBroker reports on this host's broker: the running-artifact summary,
// plus the full inspection (and, on podman, the installed unit) under --detail.
//
// --all widens it from the container this env file names to every Solace broker
// container on the host, found by image rather than by name -- the host-local
// answer to the cluster survey `--all` gives on Kubernetes. One deployed by hand,
// or under a name this config never mentions, still shows up.
func opCtrStatusBroker(a *App) error {
	if a.all {
		return ctrManager(a).StatusAll(bg(), a.detail)
	}
	if err := opCtrStatus(a); err != nil {
		return err
	}
	if !a.detail {
		return nil
	}
	return opCtrDescribe(a)
}

func opCtrCopyFrom(a *App, files []string) error { return ctrManager(a).CopyFrom(bg(), files) }
func opCtrCopyInto(a *App, files []string) error {
	return ctrManager(a).CopyInto(bg(), files, a.destDir)
}

// opCtrTeardownDomainCerts removes the configured domain CAs from this host's
// broker. The operation is platform-agnostic and already ran over this transport
// for the loading half; domainCANames is shared with the k8s handler.

// opCtrGenArtifact renders this host's deploy artifact (podman quadlet / docker
// compose file) without applying it.
// opCtrGenArtifact renders this host's deploy artifact and nothing else.
//
// No secret value appears, and none can: a compose file references its secrets as
// environment variables and a quadlet references podman's secret store by name, so the
// values are created by `broker deploy`. Kubernetes is the only platform where a secret can
// BE the artifact, which is why only that side of generate is secret-bearing.
//
// The role comes from --pod, falling back to hostname detection -- and unlike deploy, an
// undetectable hostname WARNS and renders the primary rather than failing: a render
// changes nothing, so refusing to show anything would be less useful than showing the
// likeliest answer and saying so.
func opCtrGenArtifact(a *App) error {
	role, err := containerRenderRole(a)
	if err != nil {
		return err
	}
	id := a.Cfg.ResolveNode(role)
	if a.Platform == config.Podman {
		return emitOrWrite(a, render.Quadlet(a.Cfg, id), "quadlet unit")
	}
	return emitOrWrite(a, render.Compose(a.Cfg, id), "compose file")
}

// opCtrGenSecrets renders the shell that supplies this host's secrets (podman:
// create them in its store; docker: export the variables compose reads). It gets
// the same preflight the real deploy does -- printing a script that would create an
// empty secret is worse than refusing.
// `generate secrets broker` is gone, and with it the secret-creation script this rendered.
// On both container platforms a secret cannot be part of the deploy artifact, so the only
// thing such a command could print was the VALUES -- and `broker deploy` creates them
// itself, so nothing needed the script. render.SecretScript went with it;
// container.ResolveSecretValues stays, because deploy uses it.

// opCtrRemoveBroker is the prompted teardown of this host's broker: the container, its
// artifact, its engine secrets and the server-certificate bundle, then -- only if asked
// for -- the data directory.
//
// There is no namespace analog here. The host's data directory is the layer that survives,
// so it is what --delete-data governs; removing the certificate bundle is NOT optional and
// a failure there is fatal, because leaving a private key on the host is the worst outcome
// available.
func opCtrRemoveBroker(a *App) error {
	if !confirmDelete(a, containerWhat(a)) {
		return nil
	}
	return ctrManager(a).Delete(bg(), confirmLayer(a, layerData))
}

// opCtrDeployAll runs the full node-local bring-up: check -> prep host -> deploy.
// The cross-host config-sync leader (HA, primary-only) is a separate explicit step.
// `deploy all` / `remove all` are gone: opCtrDeployHost carries the whole per-host
// lifecycle, and there is no third noun to sequence.

// containerRole resolves which node of a redundancy group THIS host is, for the
// commands whose --pod is a node identity rather than a pod selector (broker deploy,
// broker generate, and the two HA actions under broker perform).
//
// It exists because `config.ParseRole("")` returns Primary regardless of
// redundancy, so omitting the role on an HA backup host silently rendered and
// deployed a SECOND PRIMARY into the group -- while `deploy broker`'s own Short
// said the role was "required in HA", which it was not. Nothing cross-checked the
// hostname, and two primaries is the kind of mistake that shows up as a
// redundancy failure long after the command that caused it.
//
// So an omitted role in HA is now detected against the redundancy.* table
// (broker.Ops.DetectRole) -- by this host's name first, then by this machine's own
// addresses -- announced so the operator can see what was decided for them, and a
// loud error when nothing matches or more than one does.
//
// An explicit role still WINS, which is the escape hatch for a host that matches
// nothing; it is no longer taken unexamined, though. Detection runs anyway and a
// disagreement warns, because naming the wrong role is how a second primary joins a
// group and the mistake is invisible afterwards. It warns rather than prompts: the
// operator said which node this is, and a deploy scripted across three hosts must not
// stall. A host that matches nothing at all is not a disagreement and stays silent.
//
// Standalone keeps ParseRole's default: there is one node, the role
// argument means nothing, and demanding one would be noise.
func containerRole(a *App) (config.Role, error) {
	return detectContainerRole(a, a.pod, true)
}

// containerRenderRole is containerRole for `generate broker`, which RENDERS and
// changes nothing. Detection still helps on a real node, but an undetectable
// hostname must NOT stop it: rendering the backup's artifact to read it on a
// laptop is an ordinary thing to do, and refusing there would make the
// look-before-you-leap surface unusable exactly where it is most wanted.
//
// The asymmetry is safe because the failure modes are not comparable. A rendered
// artifact NAMES the node it is for, so a fallback is visible in the output the
// operator is already reading; the same guess on a deploy is invisible until the
// redundancy group breaks. So this warns and falls back to the primary, and only
// the deploying paths fail loud.
func containerRenderRole(a *App) (config.Role, error) {
	return detectContainerRole(a, a.pod, false)
}

// detectContainerRole is the shared body. mustDetect says whether an
// undetectable hostname is fatal -- see the two wrappers for why that differs.
//
// Standalone detects nothing and always answers primary: there is one node, it is
// always this host, and its routername was settled at load (App.fillStandaloneNodeName).
func detectContainerRole(a *App, arg string, mustDetect bool) (config.Role, error) {
	if !a.Cfg.RedundancyEnabled() {
		role, err := config.ParseRole(arg)
		return role, asUsage(err)
	}
	if arg != "" {
		role, err := config.ParseRole(arg)
		if err != nil {
			return "", asUsage(err)
		}
		// An explicit role wins, but it is still CHECKED against what the host looks
		// like: naming the wrong one is how a second primary joins a group, and the
		// mistake is invisible afterwards. Warn and proceed -- never prompt -- because
		// the operator said which node this is and a scripted deploy must not stall.
		if detected, how, derr := ctrOps(a).DetectRole(); derr == nil && detected != role {
			warn("--pod says %s, but this host looks like the %s node (matched by %s) -- proceeding with %s",
				roleWord(role), roleWord(detected), how, roleWord(role))
		}
		return role, nil
	}
	role, err := ctrOps(a).LocalRole("")
	if err != nil {
		if !mustDetect {
			warn("%v -- rendering the primary's artifact; name a role to render another", err)
			return config.Primary, nil
		}
		return "", fmt.Errorf("%w\n(this host must be one of the redundancy.* node entries, or name the role: "+
			"`broker deploy --pod primary|backup|monitor`)", err)
	}
	step("node role detected: %s", roleWord(role))
	return role, nil
}

// containerWhat labels the delete/down confirmation target.
func containerWhat(a *App) string {
	return fmt.Sprintf("%s broker container %s", a.Platform, a.Cfg.ContainerBlock(a.Platform).Name)
}

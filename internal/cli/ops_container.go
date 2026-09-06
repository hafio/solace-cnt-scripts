package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	return o
}

// ctrManager builds a container host Manager, wiring stdout as the report sink,
// stdin as the prompt source, and the resolved env path so PrepHost can write a
// generated PSK back into it.
func ctrManager(a *App) *container.Manager {
	m := container.NewManager(a.Runner, a.Cfg, a.Platform, lineSink(), os.Stdout)
	m.In = os.Stdin
	m.EnvPath = a.envPath
	m.RestartApproved = a.restart
	m.Confirm = func(question string) bool { return confirmRestart(a, question) }
	return m
}

// lifecycle

func opCtrCheck(a *App) error    { return ctrManager(a).Check(bg()) }
func opCtrPrepAll(a *App) error  { return ctrManager(a).PrepHost(bg()) }
func opCtrPrepHost(a *App) error { return ctrManager(a).PrepHost(bg()) }

// opCtrDeploy renders the deploy artifact and starts the container.
func opCtrDeploy(a *App, role config.Role) error {
	return ctrManager(a).Deploy(bg(), role)
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
// on Kubernetes. It asks with confirmDelete's own verb ("Restart"), since the
// action is not a deletion and a prompt that says otherwise would be read as one.
func opCtrRestartBroker(a *App) error {
	if !confirmAction(a, "Restart", "restart", containerWhat(a)) {
		return nil
	}
	return ctrManager(a).Restart(bg())
}

// config steps
//
// There is no run-everything step: these talk to a live broker over its CLI and are
// not uniformly re-runnable, so `config` documents the order rather than executing
// it. `leader` was always separate anyway -- it is primary-only and part of a
// cross-host handshake, so it fails loud on a backup or monitor host.

// opCtrConfigLeader asserts the config-sync leader from this host. It is
// primary-only and HA-only; the role arg (empty -> detect from hostname) lets an
// operator override detection. LeaderLocal fails loud on a backup/monitor host,
// and first reverts any released activity on the mate over SEMP, matching the
// k8s Leader order -- an unreachable mate downgrades that step to a warning.
func opCtrConfigLeader(a *App, roleArg string) error {
	return ctrOps(a).LeaderLocal(bg(), roleArg)
}

// opCtrConfigServerCert loads the TLS server certificate via the CLI path (there
// is no secret-managed route for containers). Node-local: the single broker.
func opCtrConfigServerCert(a *App) error {
	return ctrOps(a).ServerCert(bg(), today(), config.Primary)
}

func opCtrConfigDomainCerts(a *App) error {
	return ctrOps(a).DomainCerts(bg(), config.Primary, a.Cfg.Broker.DomainCerts.Folder, a.Cfg.Broker.DomainCerts.Files)
}
func opCtrConfigDisableVPN(a *App) error   { return ctrOps(a).DisableDefaultVPN(bg(), config.Primary) }
func opCtrConfigDisableUsers(a *App) error { return ctrOps(a).DisableDefaultUsers(bg(), config.Primary) }
func opCtrConfigProductKeys(a *App) error {
	return ctrOps(a).ProductKeys(bg(), a.Cfg.Broker.ProductKeys, config.Primary)
}

// opCtrExecCLI uploads and runs a local Solace CLI script in the broker container.
// A bare filename (no path separator) resolves under the configured cliScripts
// folder; a path is used as-is -- the same rule as the k8s handler.
func opCtrExecCLI(a *App, file string) error {
	if file == "" {
		return fmt.Errorf("a CLI script file is required (e.g. `solace-util cli --input setup.cli`)")
	}
	localPath := file
	if !strings.ContainsAny(file, `/\`) {
		localPath = filepath.Join(a.Cfg.Broker.CLIScriptsFolder, file)
	}
	return ctrOps(a).ExecCLI(bg(), config.Primary, localPath)
}

// check / smoke steps

func opCtrVerifyLogin(a *App) error { return ctrLogin(a, ctrOps(a)) }

// opCtrVerifyRedundancy exercises failover for the whole redundancy group from
// this host, which must be the primary (backup/monitor are rejected loud; the
// role arg, empty -> detect, is self-identification only). The coordinating
// broker session the old two-host handshake left as a placeholder: the local
// half runs over the node-local transport, and the backup's one command rides
// SEMP v1 to nodes.backup.ip, preflighted before anything is released.
func opCtrVerifyRedundancy(a *App, roleArg string) error {
	return ctrOps(a).RedundancyCoordinated(bg(), roleArg)
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
	ok, err := o.Login(bg(), config.Primary, a.Cfg.Admin.User, a.Cfg.Admin.Pass)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("SEMP login failed on this host's broker (see reason above)")
	}
	return nil
}

// day-2 ops (node-local)
func opCtrStatus(a *App) error   { return ctrManager(a).Status(bg()) }
func opCtrLogs(a *App) error     { return ctrManager(a).Logs(bg()) }
func opCtrCLI(a *App) error      { return ctrManager(a).CLI(bg()) }
func opCtrShell(a *App) error    { return ctrManager(a).Shell(bg()) }
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
func opCtrTeardownDomainCerts(a *App) error {
	return ctrOps(a).RemoveDomainCerts(bg(), config.Primary, domainCANames(a.Cfg))
}
// opCtrGenArtifact renders this host's deploy artifact (podman quadlet / docker
// compose file) without applying it.
func opCtrGenArtifact(a *App, role config.Role) error {
	id := a.Cfg.ResolveNode(role)
	if a.Platform == config.Podman {
		return emit(render.Quadlet(a.Cfg, id))
	}
	return emit(render.Compose(a.Cfg, id))
}

// opCtrGenSecrets renders the shell that supplies this host's secrets (podman:
// create them in its store; docker: export the variables compose reads). It gets
// the same preflight the real deploy does -- printing a script that would create an
// empty secret is worse than refusing.
func opCtrGenSecrets(a *App) error {
	if err := render.SecretPreflight(a.Cfg, a.Platform); err != nil {
		return err
	}
	return emit(render.SecretScript(a.Cfg, a.Platform))
}


// remove + orchestration

// opCtrDelete stops and removes the broker container, guarded by confirmDelete
// (nothing goes without a yes) and confirmLayer (the data directory is kept unless
// asked for by name).
func opCtrDelete(a *App) error {
	if !confirmDelete(a, containerWhat(a)) {
		return nil
	}
	return ctrManager(a).Delete(bg(), confirmLayer(a, layerData))
}

// opCtrDeployAll runs the full node-local bring-up: check -> prep host -> deploy.
// The cross-host config-sync leader (HA, primary-only) is a separate explicit step.
func opCtrDeployAll(a *App, role config.Role) error {
	m := ctrManager(a)
	ctx := bg()
	if err := m.Check(ctx); err != nil {
		return err
	}
	if err := m.PrepHost(ctx); err != nil {
		return err
	}
	return m.Deploy(ctx, role)
}

// opCtrRemoveAll tears down this host's broker. Unlike Kubernetes -- where removing
// everything also takes the secrets and the namespace -- a container host has no
// layer above the broker, so removing all of it is exactly removing the broker.
func opCtrRemoveAll(a *App) error { return opCtrDelete(a) }

// containerRole resolves which node of a redundancy group THIS host is, for the
// three commands whose [role] is a node identity rather than a pod selector
// (deploy broker, deploy all, generate broker).
//
// It exists because `config.ParseRole("")` returns Primary regardless of
// redundancy, so omitting the role on an HA backup host silently rendered and
// deployed a SECOND PRIMARY into the group -- while `deploy broker`'s own Short
// said the role was "required in HA", which it was not. Nothing cross-checked the
// hostname, and two primaries is the kind of mistake that shows up as a
// redundancy failure long after the command that caused it.
//
// So an omitted role in HA is now detected from this host's name against the
// nodes.* table (broker.Ops.LocalRole), announced so the operator can see what
// was decided for them, and a loud error when the hostname matches no configured
// node or more than one. An explicit role always wins -- LocalRole itself returns
// it unexamined -- which is the escape hatch for a host whose name does not match
// the env file. Standalone keeps ParseRole's default: there is one node, the role
// argument means nothing, and demanding one would be noise.
func containerRole(a *App, arg string) (config.Role, error) {
	return detectContainerRole(a, arg, true)
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
func containerRenderRole(a *App, arg string) (config.Role, error) {
	return detectContainerRole(a, arg, false)
}

// detectContainerRole is the shared body. mustDetect says whether an
// undetectable hostname is fatal -- see the two wrappers for why that differs.
func detectContainerRole(a *App, arg string, mustDetect bool) (config.Role, error) {
	if arg != "" || !a.Cfg.RedundancyEnabled() {
		return config.ParseRole(arg)
	}
	role, err := ctrOps(a).LocalRole("")
	if err != nil {
		if !mustDetect {
			warn("%v -- rendering the primary's artifact; name a role to render another", err)
			return config.Primary, nil
		}
		return "", fmt.Errorf("%w\n(this host must be one of the nodes.* entries, or name the role: "+
			"`deploy broker primary|backup|monitor`)", err)
	}
	step("node role detected from hostname: %s", roleWord(role))
	return role, nil
}

// containerWhat labels the delete/down confirmation target.
func containerWhat(a *App) string {
	return fmt.Sprintf("%s broker container %s", a.Platform, a.Cfg.ContainerBlock(a.Platform).Name)
}

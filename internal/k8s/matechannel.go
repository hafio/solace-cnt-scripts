package k8s

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"solace/internal/broker"
	"solace/internal/config"
	"solace/internal/engine"
	"solace/internal/output"
)

// matechannel.go reaches a REPLICATION site in another cluster.
//
// It lives here rather than in internal/broker because internal/k8s already imports
// internal/broker -- the reverse would be an import cycle, and internal/broker execs
// nothing of its own in any case. The interface and every parser stay there; this file
// is only the kubectl leg, handed over as a broker.CLIRunner. It is the same seam
// NewTransport already uses, and internal/cli wires both the same way.
//
// The site is addressed by its OWN kubernetes block, not by this env file's: a DR mate
// is a different deployment, usually in a different cluster, and the cluster is carried
// in that block's `command` (e.g. `kubectl --context dr`) so it passes the execution
// guard exactly as kubernetes.command does.

// NewMateChannel builds the channel to one replication site over its own cluster CLI.
//
// cfg is THIS env file -- it supplies nothing about the mate except the execution
// guard's approved-command set, which is the operator's per-invocation
// --allow-command list and must therefore travel with the process rather than with the
// site. Everything about WHERE to go comes from the site entry.
//
// log is the RAW line sink Cluster, Manager and Ops all take, and for the same reason:
// this leg cleans up after itself and a failed cleanup has to be visible without being
// fatal. A nil log discards, so a caller that has no sink is not forced to invent one.
func NewMateChannel(r engine.Runner, cfg *config.Config, site config.ReplSite,
	log func(string, ...any)) (broker.MateChannel, error) {
	k := site.Via.Kubernetes
	if k == nil {
		return nil, fmt.Errorf("replication site %q is not reached over kubernetes", site.VirtualRouterName)
	}
	// The guard runs HERE, not only in Validate: this channel is built straight from a
	// *config.Config and must not assume Validate ever ran -- the same reason
	// Cluster.cmd and kubectlTransport.cmd re-check their own command on every call.
	cmd, err := cfg.SiteCommand(site.VirtualRouterName)
	if err != nil {
		return nil, err
	}

	// Always the site's PRIMARY pod. The replication role is `HA: yes` in the broker's
	// own config-sync table, so a change written here reaches that site's HA backup
	// without this tool having to find it -- which is what lets one MateChannel stand
	// for a whole HA group.
	pod := podNameFor(k.Name, config.Primary)

	what := fmt.Sprintf("site %s (%s/%s)", site.VirtualRouterName, k.Namespace, pod)
	run := func(ctx context.Context, name, script string) ([]byte, error) {
		// Upload on stdin, never as an argument: a replication script carries no
		// secret today, but the body-on-stdin rule is the transport's and is not
		// worth a local exception.
		shcmd := "cat > " + shSingleQuote(broker.CLIScriptPath(name))
		if _, err := r.OutputInput(ctx, []byte(script), cmd.Name(),
			cmd.Args(execArgs(k.Namespace, pod, true, []string{"sh", "-c", shcmd})...)...); err != nil {
			return nil, fmt.Errorf("upload cli script %q to %s: %w", name, what, err)
		}
		// Deleted on every path out, including a failed run. The LOCAL leg has always
		// done this (Ops.removeCLI); this leg did not, so every `configure dr` and
		// `perform dr` over kubectl accumulated a script inside a broker belonging to
		// the OTHER site -- the one place in this tool that writes to somebody else's
		// cluster, and so the one place a leak is least defensible. The comment above
		// says the script carries no secret TODAY; leaving files behind against that
		// assurance is what makes the word "today" expensive later.
		//
		// Best-effort and never fatal, exactly as removeCLI is: a failed cleanup must
		// not turn a successful role change into a reported failure, and the operator
		// can see it in the warning.
		defer func() {
			if _, err := r.Output(ctx, cmd.Name(),
				cmd.Args(execArgs(k.Namespace, pod, false,
					[]string{"rm", "-f", broker.CLIScriptPath(name)})...)...); err != nil && log != nil {
				output.NewFunc(log).Warn("cleanup of cli script %q on %s failed: %v", name, what, err)
			}
		}()
		// The CLI is handed the RELATIVE name: its working directory is the scripts
		// folder, which is what Ops.RunCLI has always relied on.
		out, err := r.Output(ctx, cmd.Name(),
			cmd.Args(execArgs(k.Namespace, pod, false,
				[]string{broker.CLIBinary, "-Apes", broker.CLIArg(name)})...)...)
		if err != nil {
			return out, fmt.Errorf("run cli script %q on %s: %w", name, what, err)
		}
		return out, nil
	}
	return broker.NewCLIMate(run, what), nil
}

// ReadSecretKey reads one key out of a Kubernetes Secret, for a replication site whose
// mate password lives in a cluster rather than in the env file.
//
// It uses the SITE's own cluster CLI when it has one, so a site reached over SEMP can
// still keep its password in a cluster this machine can see -- and falls back to this
// env file's own runtime otherwise.
//
// `-o jsonpath` + `--decode` rather than `get -o yaml`: the value never lands in a file,
// never reaches argv, and only the one key is fetched. The kubectl invocation itself goes
// through the execution guard, like every other command this tool runs.
func ReadSecretKey(r engine.Runner, cfg *config.Config, site config.ReplSite,
	ref *config.ReplPassSecret) (string, error) {
	cmd, err := secretCommand(cfg, site)
	if err != nil {
		return "", err
	}
	// The key is a JSON object member, so a name carrying a dot has to be escaped or
	// jsonpath reads it as a path step and silently returns nothing.
	jsonpath := fmt.Sprintf(`{.data.%s}`, strings.ReplaceAll(ref.Key, ".", `\.`))
	out, err := r.Output(context.Background(), cmd.Name(),
		cmd.Args("get", "secret", ref.Name, "-n", ref.Namespace, "-o", "jsonpath="+jsonpath)...)
	if err != nil {
		return "", fmt.Errorf("read secret %s/%s for replication site %q: %w",
			ref.Namespace, ref.Name, site.VirtualRouterName, err)
	}
	enc := strings.TrimSpace(string(out))
	if enc == "" {
		return "", fmt.Errorf("secret %s/%s has no key %q, so replication site %q has no password",
			ref.Namespace, ref.Name, ref.Key, site.VirtualRouterName)
	}
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", fmt.Errorf("secret %s/%s key %q is not valid base64", ref.Namespace, ref.Name, ref.Key)
	}
	// A trailing newline is what `echo -n` forgets and `kubectl create secret
	// --from-file` keeps; it would be sent as part of the password.
	return strings.TrimRight(string(raw), "\r\n"), nil
}

// secretCommand picks the cluster CLI for a secret read: the site's own when it declares
// one, this env file's otherwise. Both go through the execution guard.
func secretCommand(cfg *config.Config, site config.ReplSite) (config.Command, error) {
	if site.Via.Kubernetes != nil {
		return cfg.SiteCommand(site.VirtualRouterName)
	}
	return cfg.ClusterCommand()
}

// execArgs builds `exec [-i] -n <ns> <pod> -- argv...`, the same shape
// kubectlTransport.execArgs builds for this deployment's own pods. It is duplicated
// rather than shared because that one derives the namespace and pod from the config,
// which is precisely what a mate in another cluster cannot do.
func execArgs(namespace, pod string, stdin bool, argv []string) []string {
	args := []string{"exec"}
	if stdin {
		args = append(args, "-i")
	}
	args = append(args, "-n", namespace, pod, "--")
	return append(args, argv...)
}

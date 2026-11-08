package k8s

import (
	"context"

	"github.com/hafio/go-deploy-utility/kube"
	"github.com/hafio/go-deploy-utility/runner"

	"solace/internal/broker"
	"solace/internal/config"
	"solace/internal/engine"
)

// kubectlTransport implements broker.Transport by wrapping `kubectl exec <pod> -n <ns>
// --` and `kubectl cp -n <ns>` against the broker pod for a role. Broker pods are
// single-container, so no `-c` flag is used (050:30, 053:56, enter-solace-cli.sh:18).
// Every call routes through the injected runner.Runner, so the Echo runner records the
// command without running it and tests capture the exact argv.
type kubectlTransport struct {
	r   runner.Runner
	cfg *config.Config
}

// NewTransport builds the Kubernetes broker.Transport over the given runner and
// config. The returned transport binds each role to pod <name>-pubsubplus-<role>-0
// in the broker namespace.
func NewTransport(r runner.Runner, cfg *config.Config) broker.Transport {
	return &kubectlTransport{r: engine.NoEnv(r), cfg: cfg}
}

func (k *kubectlTransport) ns() string { return k.cfg.K8s.Namespace }

// cmd is the configured cluster CLI (kubernetes.command, default `kubectl`): argv[0]
// plus any leading arguments that precede every call's own. Like Cluster.cmd it
// re-runs the execution guard on every call -- the transport is constructed
// straight from a *config.Config, so it must not assume Validate ever ran.
func (k *kubectlTransport) cmd() (config.Command, error) { return k.cfg.ClusterCommand() }

// base is the guarded cluster CLI as the argv the kube builders start from.
func (k *kubectlTransport) base() ([]string, error) {
	kc, err := k.cmd()
	if err != nil {
		return nil, err
	}
	return kc.Argv(), nil
}

// execCmd builds `exec [-i] <pod> -n <ns> -- argv...` for a role's pod. stdin adds the
// `-i` flag for commands that read stdin (OutputInput).
func (k *kubectlTransport) execCmd(role config.Role, stdin bool, argv []string) (runner.Cmd, error) {
	base, err := k.base()
	if err != nil {
		return runner.Cmd{}, err
	}
	return kube.Exec(base, podName(k.cfg, role), kube.ExecOptions{Namespace: k.ns(), Stdin: stdin}, argv...)
}

func (k *kubectlTransport) Run(ctx context.Context, role config.Role, argv ...string) error {
	cmd, err := k.execCmd(role, false, argv)
	if err != nil {
		return err
	}
	return engine.Run(ctx, k.r, cmd)
}

func (k *kubectlTransport) Output(ctx context.Context, role config.Role, argv ...string) ([]byte, error) {
	cmd, err := k.execCmd(role, false, argv)
	if err != nil {
		return nil, err
	}
	return engine.Output(ctx, k.r, cmd)
}

func (k *kubectlTransport) OutputInput(ctx context.Context, role config.Role, in []byte, argv ...string) ([]byte, error) {
	cmd, err := k.execCmd(role, true, argv)
	if err != nil {
		return nil, err
	}
	cmd.Stdin = in
	return engine.Output(ctx, k.r, cmd)
}

// Upload writes data to destPath inside the pod by piping it to `sh -c 'cat >
// <dest>'` on stdin (kube.Upload) -- the secret-safe path: the body rides Cmd.Stdin
// (shown as a byte count under the Echo runner), never an argv or a temp file. destPath
// is tool-generated and validName-checked upstream, and kube.Upload quotes it anyway.
func (k *kubectlTransport) Upload(ctx context.Context, role config.Role, data []byte, destPath string) error {
	base, err := k.base()
	if err != nil {
		return err
	}
	cmd, err := kube.Upload(base, podName(k.cfg, role), kube.ExecOptions{Namespace: k.ns()}, destPath, data, false)
	if err != nil {
		return err
	}
	return engine.Run(ctx, k.r, cmd)
}

// UploadFile copies a local file into the pod via
// `kubectl cp -n <ns> <local> <pod>:<dest>` (051:51, 069:196, copy scripts:53).
func (k *kubectlTransport) UploadFile(ctx context.Context, role config.Role, localPath, destPath string) error {
	kc, err := k.cmd()
	if err != nil {
		return err
	}
	return engine.Run(ctx, k.r, runner.Cmd{Argv: kc.Argv("cp", "-n", k.ns(), localPath, podName(k.cfg, role)+":"+destPath)})
}

// Download copies remotePath from the pod to localPath via
// `kubectl cp -n <ns> <pod>:<remote> <local>` (051:53, 069:196, copy scripts:42).
func (k *kubectlTransport) Download(ctx context.Context, role config.Role, remotePath, localPath string) error {
	kc, err := k.cmd()
	if err != nil {
		return err
	}
	return engine.Run(ctx, k.r, runner.Cmd{Argv: kc.Argv("cp", "-n", k.ns(), podName(k.cfg, role)+":"+remotePath, localPath)})
}

// Package container implements the docker/podman ("container") platform: a
// node-local broker.Transport over `<engine> exec/cp`, plus a host Manager that
// deploys and operates one broker container per host (the analog of k8s.Cluster).
// It ports the docker-podman/*.sh script family. Unlike k8s -- where one control
// point drives every pod in a namespace -- a container host runs a single broker,
// so the transport is node-local and ignores the role argument; the one HA
// coordination that needs the backup, the redundancy test, runs from the primary
// host and reaches it over SEMP instead (see broker's RedundancyCoordinated).
package container

import (
	"context"

	"github.com/hafio/go-deploy-utility/containercli"
	"github.com/hafio/go-deploy-utility/runner"

	"solace/internal/broker"
	"solace/internal/config"
	"solace/internal/engine"
)

// containerTransport implements broker.Transport by wrapping `<runtime> exec/cp`
// against the single broker container on this host. Node-local: there is one
// container per host, so the role argument is ignored (the k8s transport maps it
// to a pod; here every call targets the same container). Every call routes through
// the injected runner.Runner, so the Echo runner records the command without running it
// and tests capture the exact argv.
type containerTransport struct {
	r    runner.Runner
	cfg  *config.Config
	p    config.Platform
	name string // container name
}

// NewTransport builds the container broker.Transport over the given runner,
// config, and platform. It captures the container name (cfg.ContainerBlock(p).Name)
// once; the role argument on every method is ignored (node-local, one
// container/host). The runtime command is deliberately NOT captured -- it is
// resolved through the execution guard on every call (runtime below).
func NewTransport(r runner.Runner, cfg *config.Config, p config.Platform) broker.Transport {
	return &containerTransport{
		r:    engine.NoEnv(r),
		cfg:  cfg,
		p:    p,
		name: cfg.ContainerBlock(p).Name,
	}
}

// runtime is the guarded container runtime command (docker.command /
// podman.command). It re-runs execguard.Check on every call rather than
// caching the value at construction: the transport can be built from a Config that
// never went through config.Load, so this is the executor half of the guard's two
// enforcement points and must not trust what it was handed.
func (c *containerTransport) runtime() (config.Command, error) {
	return c.cfg.RuntimeCommand(c.p)
}

// exec builds `exec [-i] <name> argv...` through containercli.Exec, with in as the
// child's stdin when stdin is set (Upload and OutputInput). There is no `--` separator:
// docker `exec` rejects it, and it is unnecessary for both engines since the container
// name is positional.
func (c *containerTransport) exec(stdin bool, in []byte, argv []string) (runner.Cmd, error) {
	rt, err := c.runtime()
	if err != nil {
		return runner.Cmd{}, err
	}
	cmd, err := containercli.Exec(rt.Argv(), c.name, containercli.ExecOptions{Stdin: stdin}, argv...)
	if err != nil {
		return runner.Cmd{}, err
	}
	if stdin {
		cmd.Stdin = in
	}
	return cmd, nil
}

func (c *containerTransport) Run(ctx context.Context, _ config.Role, argv ...string) error {
	cmd, err := c.exec(false, nil, argv)
	if err != nil {
		return err
	}
	return engine.Run(ctx, c.r, cmd)
}

func (c *containerTransport) Output(ctx context.Context, _ config.Role, argv ...string) ([]byte, error) {
	cmd, err := c.exec(false, nil, argv)
	if err != nil {
		return nil, err
	}
	return engine.Output(ctx, c.r, cmd)
}

func (c *containerTransport) OutputInput(ctx context.Context, _ config.Role, in []byte, argv ...string) ([]byte, error) {
	cmd, err := c.exec(true, in, argv)
	if err != nil {
		return nil, err
	}
	return engine.Output(ctx, c.r, cmd)
}

// Upload writes data to destPath inside the container with containercli.Upload, which
// pipes it to `sh -c "cat > '<dest>'"` on stdin -- the secret-safe path: the body rides
// Cmd.Stdin (shown as a byte count under the Echo runner), never an argv or a temp file.
// Upload single-quotes destPath, so it is one word whatever it holds; it is also
// tool-generated and validName-checked upstream.
func (c *containerTransport) Upload(ctx context.Context, _ config.Role, data []byte, destPath string) error {
	rt, err := c.runtime()
	if err != nil {
		return err
	}
	cmd, err := containercli.Upload(rt.Argv(), c.name, destPath, data, false)
	if err != nil {
		return err
	}
	return engine.Run(ctx, c.r, cmd)
}

// UploadFile copies a local file into the container via
// `<runtime> cp <local> <name>:<dest>`.
func (c *containerTransport) UploadFile(ctx context.Context, _ config.Role, localPath, destPath string) error {
	rt, err := c.runtime()
	if err != nil {
		return err
	}
	return engine.Run(ctx, c.r, runner.Cmd{Argv: rt.Argv("cp", localPath, c.name+":"+destPath)})
}

// Download copies remotePath from the container to localPath via
// `<runtime> cp <name>:<remote> <local>`.
func (c *containerTransport) Download(ctx context.Context, _ config.Role, remotePath, localPath string) error {
	rt, err := c.runtime()
	if err != nil {
		return err
	}
	return engine.Run(ctx, c.r, runner.Cmd{Argv: rt.Argv("cp", c.name+":"+remotePath, localPath)})
}

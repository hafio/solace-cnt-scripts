package k8s

import (
	"context"
	"fmt"

	"solace/internal/config"
)

// rolloutTimeout bounds the readiness wait after scaling a broker StatefulSet up.
// The bash replicas-start busy-waited forever (replicas-start-broker.sh:19-21); this
// port waits via `kubectl rollout status --timeout` so a stuck pod fails loud instead
// of hanging.
const rolloutTimeout = "300s"

// Status prints the broker's pods, services and StatefulSets in the broker namespace,
// porting get-broker-status.sh:16-20. Each `get` streams straight through the runner,
// so the Echo runner records the three commands.
func (c *Cluster) Status(ctx context.Context) error {
	if err := c.kubectl(ctx, "get", "pods", "-n", c.ns(), "-o", "wide"); err != nil {
		return err
	}
	if err := c.kubectl(ctx, "get", "svc", "-n", c.ns()); err != nil {
		return err
	}
	return c.kubectl(ctx, "get", "statefulset", "-n", c.ns())
}

// DescribeBroker describes a role's broker pod, porting desc-broker.sh:18.
func (c *Cluster) DescribeBroker(ctx context.Context, role config.Role) error {
	return c.kubectl(ctx, "describe", "pod", "-n", c.ns(), podName(c.Cfg, role))
}

// DescribeLB describes the broker's load-balancer Service, porting desc-lb.sh:16.
func (c *Cluster) DescribeLB(ctx context.Context) error {
	return c.kubectl(ctx, "describe", "service/"+lbServiceName(c.Cfg), "-n", c.ns())
}

// Logs streams a role's pod logs, porting logs-broker.sh:20. passthrough carries any
// extra `kubectl logs` args (-f, --tail=N, ...) straight through.
func (c *Cluster) Logs(ctx context.Context, role config.Role, passthrough []string) error {
	args := append([]string{"logs", "-n", c.ns(), "pod/" + podName(c.Cfg, role)}, passthrough...)
	return c.kubectl(ctx, args...)
}

// interactiveExec runs `kubectl exec -it -n <ns> <pod> -- argv...` with this process's
// stdio wired through, for the interactive sessions (Solace CLI, shell).
func (c *Cluster) interactiveExec(ctx context.Context, role config.Role, argv ...string) error {
	args := append([]string{"exec", "-it", "-n", c.ns(), podName(c.Cfg, role), "--"}, argv...)
	k, err := c.cmd()
	if err != nil {
		return err
	}
	return c.R.RunInteractive(ctx, k.Name(), k.Args(args...)...)
}

// CLI opens an interactive Solace CLI session on a role's pod, porting
// enter-solace-cli.sh:18 (`cli -A`, the bare in-pod launcher, not the full load path).
func (c *Cluster) CLI(ctx context.Context, role config.Role) error {
	return c.interactiveExec(ctx, role, "cli", "-A")
}

// Shell opens an interactive shell on a role's pod (no bash script equivalent; the
// operational analogue of CLI for host-level troubleshooting).
func (c *Cluster) Shell(ctx context.Context, role config.Role) error {
	return c.interactiveExec(ctx, role, "bash")
}

// CopyFrom copies files out of a role's pod into the current directory, porting
// copy-files-from-broker.sh:42 (each lands under its basename). It attempts every
// file and fails loud at the end if any copy failed, rather than aborting on the
// first -- so one bad path does not strand the rest.
func (c *Cluster) CopyFrom(ctx context.Context, role config.Role, files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files specified to copy from the broker")
	}
	t := NewTransport(c.R, c.Cfg)
	var failed int
	for _, f := range files {
		// config.BaseName, not path.Base: the source is an in-broker path the
		// operator typed, and path.Base splits only on '/', so a value written with
		// a backslash would keep its directory component and become a local
		// filename with a separator in it.
		local := config.BaseName(f)
		c.logf("copying %s from %s", f, role.Word())
		if err := t.Download(ctx, role, f, local); err != nil {
			c.report().Fail("%s: %v", f, err)
			failed++
			continue
		}
		c.report().OK("%s -> %s", f, local)
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d file(s) failed to copy from the broker", failed, len(files))
	}
	return nil
}

// CopyInto copies local files into destDir inside a role's pod, porting
// copy-files-into-broker.sh:53 (default destDir "." is the pod's login directory).
// Like CopyFrom it attempts all files and fails loud at the end.
func (c *Cluster) CopyInto(ctx context.Context, role config.Role, files []string, destDir string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files specified to copy into the broker")
	}
	if destDir == "" {
		destDir = "."
	}
	t := NewTransport(c.R, c.Cfg)
	var failed int
	for _, f := range files {
		c.logf("copying %s into %s:%s", f, role.Word(), destDir)
		if err := t.UploadFile(ctx, role, f, destDir); err != nil {
			c.report().Fail("%s: %v", f, err)
			failed++
			continue
		}
		c.report().OK("%s -> %s:%s", f, role.Word(), destDir)
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d file(s) failed to copy into the broker", failed, len(files))
	}
	return nil
}

// RestartPod deletes a role's broker pod so the StatefulSet controller recreates it
// against the pod template the operator has already updated. This is the step
// kubernetes.updateStrategy=manualPodRestart requires after `deploy` changes image.tag:
// the operator updates the template and then waits for a human, so without this
// there was no in-tool way to finish an upgrade. --ignore-not-found makes a repeat
// call harmless, and the readiness wait is bounded like ReplicasStart's.
func (c *Cluster) RestartPod(ctx context.Context, role config.Role) error {
	pod := podName(c.Cfg, role)
	c.logf("deleting pod %s so the statefulset recreates it", pod)
	if err := c.kubectl(ctx, "delete", "pod", "-n", c.ns(), pod, "--ignore-not-found"); err != nil {
		return fmt.Errorf("deleting pod %s: %w", pod, err)
	}
	sts := stsName(c.Cfg, role)
	c.logf("waiting for %s to become ready", sts)
	if err := c.kubectl(ctx, "rollout", "status", "statefulset/"+sts, "-n", c.ns(), "--timeout="+rolloutTimeout); err != nil {
		return fmt.Errorf("%s did not become ready within %s after restarting %s: %w", sts, rolloutTimeout, pod, err)
	}
	return nil
}

// RestartRolling restarts every broker pod in RestartOrder, stopping at the first
// failure so a broken restart never cascades into the next role.
func (c *Cluster) RestartRolling(ctx context.Context) error {
	for _, role := range RestartOrder(c.Cfg) {
		if err := c.RestartPod(ctx, role); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cluster) ReplicasStart(ctx context.Context) error {
	for _, role := range HARoles(c.Cfg) {
		sts := stsName(c.Cfg, role)
		c.logf("scaling %s up to 1 replica", sts)
		if err := c.kubectl(ctx, "scale", "statefulset", sts, "-n", c.ns(), "--replicas=1"); err != nil {
			return fmt.Errorf("scaling %s up: %w", sts, err)
		}
		c.logf("waiting for %s to become ready", sts)
		if err := c.kubectl(ctx, "rollout", "status", "statefulset/"+sts, "-n", c.ns(), "--timeout="+rolloutTimeout); err != nil {
			return fmt.Errorf("%s did not become ready within %s: %w", sts, rolloutTimeout, err)
		}
	}
	return nil
}

// ReplicasStop scales every broker StatefulSet down to zero replicas, porting
// replicas-stop-broker.sh:23-26. The bash script's y/n confirmation lives in the CLI
// layer; here the decision to stop has already been made.
func (c *Cluster) ReplicasStop(ctx context.Context) error {
	for _, role := range HARoles(c.Cfg) {
		sts := stsName(c.Cfg, role)
		c.logf("scaling %s down to 0 replicas", sts)
		if err := c.kubectl(ctx, "scale", "statefulset", sts, "-n", c.ns(), "--replicas=0"); err != nil {
			return fmt.Errorf("scaling %s down: %w", sts, err)
		}
	}
	return nil
}

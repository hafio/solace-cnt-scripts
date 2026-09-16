package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"solace/internal/render"
)

// brokerYAMLFile is where DeployBroker writes the rendered CR when keepYAML is set,
// mirroring the bash `.broker.yaml` the operator deploy leaves in the working
// directory (020:206-210).
const brokerYAMLFile = ".broker.yaml"

// DeployBroker renders the PubSubPlusEventBroker CR and applies it on stdin
// (020:211). When keepYAML is set it also writes the manifest to .broker.yaml in
// the current directory for inspection/version control (020:206-210). The CR
// carries only secret *names*, never secret values, so applying on stdin -- rather
// than the bash apply-from-file -- changes nothing about what lands on disk.
func (c *Cluster) DeployBroker(ctx context.Context, keepYAML bool) error {
	// Before the keepYAML write and before the apply: a deploy that cannot create
	// the CR must not leave a manifest on disk suggesting it got further than it did.
	// `get` is probed alongside `create` because ConfirmBrokerApplied reads the CR
	// back immediately afterwards, and a deploy that could apply but not confirm would
	// fail after the write rather than before it.
	if err := c.PreflightAll(ctx,
		probe{verb: "create", resource: brokerResource},
		probe{verb: "get", resource: brokerResource},
	); err != nil {
		return err
	}
	manifest := render.BrokerCR(c.Cfg)
	if keepYAML {
		if err := os.WriteFile(brokerYAMLFile, manifest, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", brokerYAMLFile, err)
		}
		c.logf("wrote broker manifest to %s", brokerYAMLFile)
	}
	c.logf("deploying broker %s in %s", c.Cfg.K8s.Name, c.ns())
	return c.apply(ctx, manifest)
}

// ConfirmBrokerApplied reads the broker CR back and fails if it is not there.
//
// `kubectl apply` exiting 0 is weaker evidence than it looks: a validating or mutating
// admission webhook can reject or rewrite the object, and a CRD that exists but is not
// yet Established fails in its own way. So "deployed" is made to mean "the object is in
// the cluster" rather than "the write returned without complaining". This is the whole
// of the confirmation -- readiness is NOT waited on, because the operator reconciles
// asynchronously and `broker status` is the command for watching that.
//
// A runner that answers with NOTHING has not told us the broker is absent -- it has told
// us nothing at all, which is a different fact and must not be reported as the alarming
// one. That is the preview case (engine.Echo has no cluster behind it), and it is also
// any seam a test installs.
//
// The distinction is drawn on the OBSERVATION rather than on the runner type. A real
// `kubectl get <resource> <name>` for an object that does not exist exits non-zero and
// is caught as an error above, so a successful call returning no output can only mean
// nobody answered -- which is true of every silent runner, not just engine.Echo.
func (c *Cluster) ConfirmBrokerApplied(ctx context.Context) error {
	name := c.Cfg.K8s.Name
	raw, err := c.kubectlOutput(ctx, "get", brokerResource, name, "-n", c.ns(), "-o", "json")
	if err != nil {
		return fmt.Errorf("applied broker %q but could not read it back in namespace %q: %w\n"+
			"  Check it with `solace-util broker validate`", name, c.ns(), err)
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		c.report().KVRow(reportKeyWidth, "broker", "read-back skipped (no answer from the runner)")
		return nil
	}
	var list brokerList
	if err := json.Unmarshal(normalizeToList(raw), &list); err != nil {
		return fmt.Errorf("applied broker %q but could not read it back in namespace %q: %w\n"+
			"  Check it with `solace-util broker validate`", name, c.ns(), err)
	}
	if len(list.Items) == 0 {
		return fmt.Errorf("applied broker %q but it does not exist in namespace %q.\n"+
			"  The apply was accepted and the object is not there, which usually means an admission "+
			"webhook rejected or rewrote it, or the operator's CRD is installed but not yet established.\n"+
			"  Check the operator with `solace-util operator validate`", name, c.ns())
	}
	for _, item := range list.Items {
		if item.Metadata.Name == name {
			c.progress().OK("broker %q applied in namespace %q.", name, c.ns())
			return nil
		}
	}
	return fmt.Errorf("applied broker %q but the cluster returned a different object in namespace %q",
		name, c.ns())
}

// DeleteBroker removes the broker CR (120:55) via `delete -f - --ignore-not-found`
// of the rendered manifest, so a repeat teardown is a no-op. When purge is set it
// then deletes the per-role data PVCs (data-<name>-pubsubplus-<role>-0; 120:65-69),
// one delete per role rather than aborting on the first failure -- an RBAC gap on
// one PVC must not leave the others behind. `--ignore-not-found` already absorbs
// the benign "already gone" case, so anything that still errors is real (an RBAC
// denial, a stuck finalizer): those are collected and reported, never swallowed,
// because a caller gating on this function's return must be able to trust a nil
// error to mean the data is actually gone.
//
// purge defaults to false at the call site (keep data by default), the deliberately
// safer inverse of legacy 120's purge-by-default. The confirm/flag logic lives in
// the CLI layer; here purge is just the decision already made.
func (c *Cluster) DeleteBroker(ctx context.Context, purge bool) error {
	// The claims are probed only when they will actually be deleted, so a removal
	// that keeps the data does not demand a permission it never uses.
	probes := []probe{{verb: "delete", resource: brokerResource}}
	if purge {
		probes = append(probes, probe{verb: "delete", resource: "persistentvolumeclaims"})
	}
	if err := c.PreflightAll(ctx, probes...); err != nil {
		return err
	}
	c.logf("deleting broker %s in %s", c.Cfg.K8s.Name, c.ns())
	if err := c.deleteStdin(ctx, render.BrokerCR(c.Cfg)); err != nil {
		return err
	}
	if !purge {
		// Stated rather than left to be inferred: which of the two layers survived a
		// removal is the fact an operator most needs from this command's output.
		c.logf("PVCs kept -- the broker's persistent data survives (pass --delete-data to remove it)")
		return nil
	}
	var failed []string
	var errs []error
	var kept []string
	for _, role := range HARoles(c.Cfg) {
		// A custom volume mount names a claim the OPERATOR did not provision and this
		// tool did not create. It can point at a volume that predates this broker
		// entirely, and nothing here can tell the difference -- so --delete-data is a
		// no-op for that role, whatever else it deletes.
		//
		// Reported per role rather than silently skipped: someone who passed
		// --delete-data and got a surviving volume needs to know it survived, and why,
		// or they will assume the flag failed.
		if claim, ok := c.Cfg.K8s.Storage.CustomMountFor(role); ok {
			kept = append(kept, claim)
			c.progress().Info("PVC %s kept: a custom volume mount this tool did not create.", claim)
			continue
		}
		pvc := pvcName(c.Cfg, role)
		c.logf("deleting PVC %s", pvc)
		if err := c.kubectl(ctx, "delete", "pvc", pvc, "-n", c.ns(), "--ignore-not-found"); err != nil {
			// Logged per PVC as it happens, so the operator sees each failure even
			// though the loop keeps going to give every role a chance to delete.
			c.progress().Warn("could not delete PVC %s: %v", pvc, err)
			failed = append(failed, pvc)
			errs = append(errs, fmt.Errorf("%s: %w", pvc, err))
		}
	}
	if len(failed) > 0 {
		// errors.Join, not the last error alone: the roles can fail for DIFFERENT
		// reasons -- one on RBAC, another on a stuck finalizer -- and keeping only the
		// last one hid the first from both the message and errors.Is.
		return fmt.Errorf("PVCs not deleted: %s (persistent data survives; check RBAC or a stuck finalizer): %w",
			strings.Join(failed, ", "), errors.Join(errs...))
	}
	switch {
	case len(kept) > 0 && len(kept) == len(HARoles(c.Cfg)):
		c.logf("no PVC was deleted: every node uses a custom volume mount (%s). "+
			"Those volumes are yours -- remove them by hand if you mean to", strings.Join(kept, ", "))
	case len(kept) > 0:
		c.logf("PVCs deleted, except the custom volume mount(s) kept above: %s", strings.Join(kept, ", "))
	default:
		c.logf("PVCs deleted -- the broker's persistent data is gone")
	}
	return nil
}

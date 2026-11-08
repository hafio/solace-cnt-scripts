package k8s

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/hafio/go-deploy-utility/kube/nsgate"
	"github.com/hafio/go-deploy-utility/runner"
)

// Namespace occupancy, and why a removal never just deletes the namespace.
//
// Deleting a Namespace cascades to EVERYTHING in it, including whatever another team put
// there, and this tool very often did not create it. So `broker remove` never puts the
// Namespace in its delete set. It removes the broker's own objects, then asks what is
// left, and offers the namespace for removal only when the answer is "nothing of anyone
// else's".
//
// The gate is the shared one (kube/nsgate, the same in solmq-conn-util): the protected
// namespaces, one JSON listing of a fixed set of kinds, the occupants left once Kubernetes'
// own objects and this deployment's are discounted, the `delete namespaces` probe, then
// the question. removeNamespaceIfEmpty (internal/cli) runs it in that order, and every
// step short of a confirmed delete keeps the namespace without failing the removal.
//
// It also makes one hazard structurally impossible rather than merely documented. When a
// removal KEEPS the persistent data (no --delete-data), the PVCs are still there, and no
// ownerReference ties them to the broker -- so the namespace is not empty, so it is not
// offered, so "delete the namespace and take the retained data with it" cannot happen by
// answering one prompt wrong.

// namespaceOwned is what this deployment owns, for nsgate.Occupants to discount: the
// broker CR, which is the root -- the operator's StatefulSets, Pods, Services and generated
// Secrets carry ownerReferences to it, which Occupants follows -- and every Secret this
// env file builds (ownedSecretNames), which carry none. A Secret the env file only names is
// not owned, so it still counts.
func (c *Cluster) namespaceOwned() []nsgate.Object {
	owned := []nsgate.Object{{Kind: "PubSubPlusEventBroker", Name: c.Cfg.K8s.Name}}
	for _, name := range ownedSecretNames(c.Cfg) {
		owned = append(owned, nsgate.Object{Kind: "Secret", Name: name})
	}
	return owned
}

// NamespaceOccupants reports what still occupies the broker namespace, as sorted
// "<kind>/<name>" refs (kind in lower case), having discounted what nsgate.Occupants does
// and what namespaceOwned names.
//
// answered is false, with no error, when the runner gave no answer at all (the preview and
// test seams): a row says the check was skipped and the caller keeps the namespace quietly.
// An error means the question could not be answered, which every caller must treat as
// OCCUPIED. That direction is the whole safety property: a namespace wrongly reported empty
// gets cascade-deleted, while one wrongly reported occupied merely stays.
func (c *Cluster) NamespaceOccupants(ctx context.Context) (occupants []string, answered bool, err error) {
	ns := c.ns()
	out, err := c.outKube(ctx, func(b []string) runner.Cmd { return nsgate.List(b, ns) })
	if err != nil {
		return nil, false, fmt.Errorf("could not list what is left in namespace %q: %w\n"+
			"  The namespace is kept. Check it with `kubectl get all,pvc,secrets,configmaps -n %s`",
			ns, err, ns)
	}
	if len(bytes.TrimSpace(out)) == 0 {
		c.report().KVRow(reportKeyWidth, "namespace", "check skipped (no answer from the runner)")
		return nil, false, nil
	}
	occupants, err = nsgate.Occupants(string(out), c.namespaceOwned())
	if err != nil {
		return nil, true, fmt.Errorf("namespace %q: %w", ns, err)
	}
	return occupants, true, nil
}

// ReportNamespaceOccupied prints what is keeping the namespace and why it is being left
// alone. It is the branch an operator sees most often, so it names the objects rather
// than only their count -- "3 objects remain" gives nobody anything to act on.
func (c *Cluster) ReportNamespaceOccupied(contents []string) {
	ns := c.ns()
	c.progress().Warn("namespace %q kept: %d object(s) still in it.", ns, len(contents))
	rows := make([][]string, 0, len(contents))
	for _, ref := range contents {
		kind, name, _ := strings.Cut(ref, "/")
		rows = append(rows, []string{kind, name})
	}
	c.report().Table([]string{"KIND", "NAME"}, rows)
	c.logf("nothing was deleted from %q. Remove it by hand once you are sure nothing else needs it: "+
		"kubectl delete namespace %s", ns, ns)
}

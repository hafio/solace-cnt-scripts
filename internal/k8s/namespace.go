package k8s

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Namespace occupancy, and why a removal never just deletes the namespace.
//
// Deleting a Namespace cascades to EVERYTHING in it, including whatever another team put
// there, and this tool very often did not create it. So `broker remove` never puts the
// Namespace in its delete set. It removes the broker's own objects, then asks what is
// left, and offers the namespace for removal only when the answer is "nothing".
//
// This is a port of the same gate in the sibling project solmq-conn-util, which calls it
// the occupancy-gated namespace prompt: probe a FIXED set of kinds in one call,
// discount the objects Kubernetes puts in every namespace, treat an unreadable answer as
// occupied, and never remove a cluster namespace.
//
// It also makes one hazard structurally impossible rather than merely documented. When a
// removal KEEPS the persistent data (no --delete-data), the PVCs are still there -- so the
// namespace is not empty, so it is not offered, so "delete the namespace and take the
// retained data with it" cannot happen by answering one prompt wrong.

// occupancyKinds is what NamespaceContents asks about: one `kubectl get` of a fixed kind
// list rather than an api-resources sweep.
//
// A fixed list is the deliberate trade. `kubectl api-resources --verbs=list --namespaced`
// followed by a bulk get is more complete -- it would also see CRD-provided kinds -- but
// it demands `list` on every kind in the cluster, so a namespace-scoped identity fails the
// probe rather than the check, and one unreadable kind then decides the whole answer.
// `all` already expands to the workload kinds that matter (pods, services, deployments,
// statefulsets, replicasets, daemonsets, jobs, cronjobs); the three named after it are the
// ones that hold DATA or CREDENTIALS and so must never be cascaded away unnoticed.
var occupancyKinds = []string{"all", "persistentvolumeclaims", "secrets", "configmaps"}

// alwaysPresent are the objects Kubernetes itself puts in every namespace. They are NOT
// occupancy: counting them means no namespace is ever empty and the prompt never appears.
//
//   - the `default` ServiceAccount is created in every namespace by the service-account
//     controller
//   - the `kube-root-ca.crt` ConfigMap is injected into every namespace by the
//     root-CA-configmap publisher (default since Kubernetes 1.21)
//
// Admission controllers commonly add a default LimitRange or ResourceQuota as well. Those
// are not matched by name here because their names are cluster policy rather than a
// Kubernetes constant -- they are handled by kind in ignoredKinds below.
var alwaysPresent = map[string]bool{
	"serviceaccount/default":     true,
	"configmap/kube-root-ca.crt": true,
}

// ignoredKinds are kinds whose presence says nothing about whether a namespace is in use:
// cluster policy stamps them onto namespaces without anyone asking. A LimitRange or
// ResourceQuota left behind is not data and not a workload, and refusing to remove a
// namespace because policy put a quota in it would mean never removing one at all on a
// cluster that sets quotas.
var ignoredKinds = map[string]bool{
	"limitrange":    true,
	"resourcequota": true,
	"event":         true,
}

// NamespaceContents reports what still occupies the broker namespace, as sorted
// "<kind>/<name>" refs, having discounted everything above.
//
// An error means the question could not be answered, which every caller must treat as
// OCCUPIED. That direction is the whole safety property: a namespace wrongly reported
// empty gets cascade-deleted, while one wrongly reported occupied merely stays.
func (c *Cluster) NamespaceContents(ctx context.Context) ([]string, error) {
	ns := c.ns()
	raw, err := c.kubectlOutput(ctx, "get", strings.Join(occupancyKinds, ","), "-n", ns,
		"--ignore-not-found", "-o", "name")
	if err != nil {
		return nil, fmt.Errorf("could not list what is left in namespace %q: %w\n"+
			"  The namespace is kept. Check it with `kubectl get all,pvc,secrets,configmaps -n %s`",
			ns, err, ns)
	}
	var out []string
	for _, line := range strings.Split(string(raw), "\n") {
		ref := strings.TrimSpace(line)
		if ref == "" {
			continue
		}
		// `-o name` yields <resource>/<name>, sometimes group-qualified
		// (persistentvolumeclaim/x, serviceaccounts.v1./default). Compare on the bare
		// kind so the group suffix cannot defeat a match.
		kind, name, ok := strings.Cut(ref, "/")
		if !ok {
			continue
		}
		kind = bareKind(kind)
		if ignoredKinds[kind] || alwaysPresent[kind+"/"+name] {
			continue
		}
		out = append(out, kind+"/"+name)
	}
	sort.Strings(out)
	return out, nil
}

// bareKind strips a group/version suffix and trailing plural from a `-o name` resource
// word, so serviceaccounts.v1. and serviceaccount both compare as serviceaccount.
func bareKind(kind string) string {
	if i := strings.IndexByte(kind, '.'); i >= 0 {
		kind = kind[:i]
	}
	return strings.TrimSuffix(kind, "s")
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

// NamespaceIsProtected reports whether the broker namespace is one nothing should ever
// delete. A misconfigured env file naming `kube-system` must not be one confirmed prompt
// away from taking the cluster down, so this is checked before the question is asked
// rather than trusted to the answer.
//
// It shares DeleteNamespace's own list (protectedNamespaces, prep.go) deliberately. Two
// lists would be two things to keep in step, and the one that mattered would be whichever
// the call path happened to reach.
func (c *Cluster) NamespaceIsProtected() bool { return protectedNamespaces[c.ns()] }

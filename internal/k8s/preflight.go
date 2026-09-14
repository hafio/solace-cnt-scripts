package k8s

import (
	"context"
	"fmt"
	"strings"
)

// Preflight is the read-only probe every mutating operation runs first. It asks the
// cluster one question -- may this identity create this resource here -- using the
// same guarded argv prefix the real work will use, so what it proves is what the
// next command will actually do.
//
// It exists because the alternative failure mode is worse than a slow start: a
// deploy that writes .broker.yaml, applies half a manifest set, and then stops on
// an expired token leaves the operator to work out which half landed. Failing
// before the first byte is written keeps "nothing happened" a true statement.
//
// There is deliberately no skip flag. A probe that can be turned off is a probe
// that is off in exactly the scripted runs that most need it, and the only
// legitimate reason to skip it -- previewing without a cluster -- is already
// the Echo runner, which reaches the branch below.
//
// It never logs anyone in. Authentication is the operator's business and their
// credential store's; a tool that offered to fix an auth failure would be teaching
// people to hand it credentials it has no business holding.
func (c *Cluster) Preflight(ctx context.Context, verb, resource string) error {
	return c.PreflightAll(ctx, probe{verb: verb, resource: resource})
}

// PreflightAll asks several permission questions, stopping at the first refusal.
//
// It exists because probing one resource as a stand-in for a whole apply is a check
// that passes and then fails mid-way: an identity allowed to create custom resource
// definitions but not ClusterRoleBindings passed the operator preflight and died
// half-installed, which is exactly the state Preflight's own doc comment promises to
// prevent. So an operation probes every KIND it will touch, not a representative one.
//
// Sequential, one `auth can-i` each: kubectl has no multi-resource form of the
// question, and `auth can-i --list` answers a different one -- it returns a rule table
// needing wildcard expansion, which would trade a fail-safe verdict for a parser that
// can be wrong in the permissive direction. Each probe is one cheap API call and the
// whole surface is single-digit, so the batching win would be nil.
func (c *Cluster) PreflightAll(ctx context.Context, probes ...probe) error {
	for _, p := range probes {
		if err := c.preflightOne(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cluster) preflightOne(ctx context.Context, p probe) error {
	verb, resource := p.verb, p.resource
	where := c.nsFor(p)
	if where == "" {
		where = "the cluster"
	}
	args := c.canIArgs(p)
	if c.isEcho() {
		// Echo the probe so a preview still shows it, then skip the assertion --
		// the Echo runner answers nothing, and there is no cluster to answer.
		if err := c.kubectl(ctx, args...); err != nil {
			return err
		}
		c.report().KVRow(reportKeyWidth, "permission", "skipped (preview)")
		return nil
	}

	out, err := c.output(ctx, args...)
	answer := canIAnswer(out)
	switch {
	case err == nil && answer == "yes":
		return nil
	case answer == "no":
		// Reached the API server and got a real answer: this is RBAC, not auth.
		// kubectl already printed its own explanation to stderr; add the one line
		// that says what to ask for.
		return fmt.Errorf("not allowed to %s %s in %s: ask a cluster admin for a role binding "+
			"granting %s on %s there, then re-run", verb, resource, where, verb, resource)
	case err != nil:
		// No usable answer: an expired token, no context, an unreachable API
		// server. kubectl's own message is on stderr and wrapped in here too.
		return fmt.Errorf("cannot check permission to %s %s in %s: %w\n"+
			"  log in first (kubectl: `kubectl config use-context <ctx>`; OpenShift: `oc login <server>`), "+
			"or point kubernetes.command at the right profile", verb, resource, where, err)
	default:
		// Exit 0 with something other than "yes" -- a wrapper that swallowed the
		// answer, or a kubectl whose output shape changed. Refusing is the safe
		// direction: proceeding would mean assuming a permission nobody confirmed.
		return fmt.Errorf("could not read the answer to `auth can-i %s %s` (got %q); "+
			"if kubernetes.command wraps kubectl, make sure it passes stdout through unchanged", verb, resource, answer)
	}
}

// canIAnswer extracts the verdict from `auth can-i` output. The answer is the LAST
// non-empty line, not the whole trimmed body: kubectl prints advisory lines above it
// on stdout -- "Warning: resource 'x' is not namespace scoped" is the common one, and
// a cluster with deprecated APIs adds more. Comparing the whole output would turn
// every such cluster into the "could not read the answer" branch and block a deploy
// the operator is perfectly entitled to make.
func canIAnswer(out []byte) string {
	lines := strings.Split(string(out), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if line := strings.TrimSpace(lines[i]); line != "" {
			return line
		}
	}
	return ""
}

// brokerResource is the CRD the operator reconciles -- the thing every broker
// deploy ultimately creates. Named in full (resource.group) so `auth can-i` cannot
// match a same-named resource in another group.
const brokerResource = "pubsubpluseventbrokers.pubsubplus.solace.com"

// clusterScoped names the resources whose permission check must NOT carry a
// namespace. Passing `-n` for one of these makes kubectl print
//
//	Warning: resource 'customresourcedefinitions' is not namespace scoped
//
// and, when the namespace does not exist yet, also fail with a NotFound that has
// nothing to do with the permission being checked -- which is exactly what a
// first `deploy operator` looked like: two scary lines about a namespace, in
// front of a probe that was asking about a cluster-scoped resource.
var clusterScoped = map[string]bool{
	"customresourcedefinitions": true,
	"namespaces":                true,
	"nodes":                     true,
	"clusterroles":              true,
	"clusterrolebindings":       true,
	"storageclasses":            true,
	"persistentvolumes":         true,
}

// probe is one permission question: may this identity do verb to resource, and
// where. An empty ns means the broker namespace, which is what almost every probe
// wants; the operator paths set it explicitly because they act somewhere else.
//
// A namespace field exists because getting this wrong is silent. OperatorRestart
// probed `patch deployments` with no namespace of its own, so the question went to
// the BROKER namespace while the restart it guards happens in the operator's -- an
// identity permitted in one and not the other would pass the check and fail the work.
type probe struct {
	verb     string
	resource string
	ns       string
}

// nsFor is the namespace a probe asks about: its own, else the broker's, and none at
// all for a cluster-scoped resource.
func (c *Cluster) nsFor(p probe) string {
	if clusterScoped[p.resource] {
		return ""
	}
	if p.ns != "" {
		return p.ns
	}
	return c.ns()
}

// canIArgs builds the `auth can-i` argv, scoping it to a namespace only for
// resources that live in one.
func (c *Cluster) canIArgs(p probe) []string {
	args := []string{"auth", "can-i", p.verb, p.resource}
	if ns := c.nsFor(p); ns != "" {
		return append(args, "-n", ns)
	}
	// Omitting -n is NOT enough for a cluster-scoped resource: kubectl falls back to
	// the kubeconfig context's namespace, so the access review still carries one.
	// That is both the source of
	//
	//	Warning: resource 'namespaces' is not namespace scoped
	//
	// on an otherwise clean `broker remove`, and a check that can answer WRONG --
	// with a namespace in the review, RBAC also evaluates namespaced Roles there, so
	// a RoleBinding granting `delete namespaces` in that one namespace would make the
	// probe say yes while the delete it guards still fails. --all-namespaces clears
	// the namespace from the review, which is the question actually being asked.
	return append(args, "--all-namespaces")
}

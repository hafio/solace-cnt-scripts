package k8s

import (
	"context"

	"github.com/hafio/go-deploy-utility/kube/rbac"
	"github.com/hafio/go-deploy-utility/runner"

	"solace/internal/engine"
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
	return c.PreflightAll(ctx, rbac.Probe{Verb: verb, Resource: resource})
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
func (c *Cluster) PreflightAll(ctx context.Context, probes ...rbac.Probe) error {
	for _, p := range probes {
		if err := c.preflightOne(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cluster) preflightOne(ctx context.Context, p rbac.Probe) error {
	p = c.scope(p)
	cmd, err := c.canICmd(p)
	if err != nil {
		return err
	}
	if c.isEcho() {
		// Echo the probe so a preview still shows it, then skip the assertion --
		// the Echo runner answers nothing, and there is no cluster to answer.
		if err := engine.Run(ctx, c.R, cmd); err != nil {
			return err
		}
		c.report().KVRow(reportKeyWidth, "permission", "skipped (preview)")
		return nil
	}
	// stdout alone is the verdict, and stderr already reached the terminal as it
	// arrived, so rbac.Check is given none. A refusal is a *rbac.RefusedError.
	out, err := engine.Output(ctx, c.R, cmd)
	return rbac.Check(p, string(out), "", err)
}

// scope puts the broker namespace on a probe for a namespaced resource that names none,
// so the refusal names the namespace asked about rather than "the current namespace".
// A cluster-scoped resource and an AllNamespaces question stay cluster-wide, which
// rbac.CanI turns into --all-namespaces.
func (c *Cluster) scope(p rbac.Probe) rbac.Probe {
	if p.Namespace == "" && !p.AllNamespaces && !rbac.ClusterScoped(p.Resource) {
		p.Namespace = c.ns()
	}
	return p
}

// canICmd is the `auth can-i` command for a probe already scoped, over the guarded
// cluster CLI.
func (c *Cluster) canICmd(p rbac.Probe) (runner.Cmd, error) {
	base, err := c.base()
	if err != nil {
		return runner.Cmd{}, err
	}
	return rbac.CanI(base, p), nil
}

// canI asks one permission question WITHOUT failing on the answer, for a caller choosing
// between two paths rather than refusing to start: whether to offer a removal at all,
// whether a report row that could not be read was refused or broken. allowed is the
// verdict; answered is false when there was none to read -- an expired login, an
// unreachable server, a reply that was neither yes nor no -- which every caller must treat
// as unknown, never as either answer (rbac.Verdict).
//
// Under the Echo runner it echoes the probe and answers yes, exactly as preflightOne
// passes, so a preview shows the question and then takes the path a permitted run takes.
func (c *Cluster) canI(ctx context.Context, p rbac.Probe) (allowed, answered bool) {
	p = c.scope(p)
	cmd, err := c.canICmd(p)
	if err != nil {
		return false, false
	}
	if c.isEcho() {
		if err := engine.Run(ctx, c.R, cmd); err != nil {
			return false, false
		}
		return true, true
	}
	out, err := engine.Output(ctx, c.R, cmd)
	return rbac.Verdict(string(out), err)
}

// brokerResource is the CRD the operator reconciles -- the thing every broker
// deploy ultimately creates. Named in full (resource.group) so `auth can-i` cannot
// match a same-named resource in another group.
const brokerResource = "pubsubpluseventbrokers.pubsubplus.solace.com"

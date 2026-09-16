package k8s

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"solace/internal/config"
)

// The operator's watch scope, read-modify-write.
//
// WATCH_NAMESPACE is the one field of the operator install that does NOT belong to any
// single env file. The operator is cluster-scoped and shared: two brokers in two
// namespaces, described by two env files, are meant to be reconciled by ONE operator.
// This tool used to render that value from config and apply it blindly, which meant
// `operator deploy` from the second env file silently stopped watching the first
// broker's namespace -- no error anywhere, and the symptom was a broker that had simply
// stopped reconciling.
//
// So the value is now reconciled rather than overwritten: a deploy UNIONS the env file's
// namespaces into whatever is running, and a removal SUBTRACTS them, deleting the
// operator only when the env file accounts for every namespace it watches.
//
// One state dominates every table below: an operator whose WATCH_NAMESPACE is empty (or
// absent) watches EVERY namespace. That is a superset of any list, so a union with it
// changes nothing, and a subtraction from it is not expressible at all -- there is no
// "all namespaces except this one". Both cases are reported and left alone rather than
// approximated.

// watchEnvVar is the operator Deployment's watch-scope environment variable, spelled
// once. It is both what the bundle template renders and what installedWatch reads back,
// so a rename cannot half-apply (the assets/operator-<version>.yaml.tmpl bundle).
const watchEnvVar = "WATCH_NAMESPACE"

// desiredWatch is the watch list this env file describes: the configured
// watchNamespaces with the broker namespace appended when broker-ns watching is enabled
// -- the default (000-env.sh:85-89).
//
// Entries are trimmed and de-duplicated, first occurrence winning. The broker namespace
// is very often already in the configured list, and the repeat reached both the report
// and the applied Deployment's WATCH_NAMESPACE; controller-runtime's cache is map-keyed,
// so it collapsed there harmlessly, which is exactly why it went unnoticed.
//
// An EMPTY result is meaningful, not empty-as-in-unset: it means this env file asks the
// operator to watch every namespace. It is only reachable with an explicit
// `watchBrokerNs: false` and no watchNamespaces, since WatchBrokerNSEnabled defaults to
// true.
func desiredWatch(cfg *config.Config) []string {
	list := cfg.K8s.Operator.WatchNamespaces
	if cfg.K8s.Operator.WatchBrokerNSEnabled() {
		// Appended to the configured list rather than de-duplicated afterwards:
		// splitWatch already trims, drops empties and keeps the first occurrence, so a
		// broker namespace the list already names keeps its position, and the separator
		// left beside an empty list falls away with the empty entry it creates.
		list += "," + cfg.K8s.Namespace
	}
	return splitWatch(list)
}

// watchNamespace is desiredWatch in the form the bundle template substitutes: one
// comma-joined value. Both come from one definition so the report, the rendered
// artifact and the applied Deployment cannot disagree about the list.
func watchNamespace(cfg *config.Config) string {
	return strings.Join(desiredWatch(cfg), ",")
}

// installedWatch reports the watch scope of the operator that is actually running.
//
// It separates three states that mean different things, and conflating any two of them
// is how the original bug worked:
//
//	found=false             no operator Deployment anywhere in the cluster
//	allNS=true, list=nil    an operator watching EVERY namespace -- either
//	                        WATCH_NAMESPACE is absent (the operator-sdk default) or it
//	                        is present and empty. Both mean the same thing to the
//	                        operator, so they mean the same thing here.
//	allNS=false, list=[...]  an operator watching exactly these namespaces
//
// The query is all-namespaces for the same reason installedOperatorImage's is: asking in
// the configured namespace returns NotFound on a first install, and an operator someone
// installed elsewhere is precisely the one whose watch list must not be clobbered.
func (c *Cluster) installedWatch(ctx context.Context) (list []string, allNS, found bool, err error) {
	dep, err := c.findOperatorDeployment(ctx)
	if err != nil {
		return nil, false, false, fmt.Errorf("reading the installed operator's watch scope: %w", err)
	}
	list, allNS, found = watchFromDeployment(dep)
	return list, allNS, found, nil
}

// findOperatorDeployment fetches the operator Deployment, or nil when no operator is
// installed -- which is the ordinary first-install case, not an error.
//
// The query is all-namespaces, for the reason both its callers record: asking in the
// configured namespace returns NotFound on a first install, printing an alarming line
// about a lookup whose failure is expected, and an operator installed somewhere ELSE is
// precisely the one worth noticing before a deploy.
//
// It exists so the fetch can happen ONCE and be read twice. `check deploy` wants two
// facts about this same object -- the watch scope and the running image -- and each
// reader used to issue its own identical cluster-wide list, so every healthy `validate`
// paid for two.
//
// The match is by name AND namespace, in that order of preference. A Deployment of the
// same name in ANY namespace used to answer, and the answer decides whether the watch
// list is widened and whether a deploy is a downgrade -- so anyone able to create a
// Deployment somewhere on the cluster could steer both. The one in the namespace this
// env file resolves wins outright; an install elsewhere counts only when nothing is
// there and it is the sole candidate, and two elsewhere is an error naming them.
func (c *Cluster) findOperatorDeployment(ctx context.Context) (*deploymentItem, error) {
	var deps deploymentList
	if err := c.getJSON(ctx, &deps, "deployment", "--all-namespaces"); err != nil {
		return nil, err
	}
	want := c.operatorNS(ctx)
	var elsewhere []*deploymentItem
	for i := range deps.Items {
		it := &deps.Items[i]
		if it.Metadata.Name != operatorDeployment {
			continue
		}
		if it.Metadata.Namespace == want {
			return it, nil
		}
		elsewhere = append(elsewhere, it)
	}
	switch len(elsewhere) {
	case 0:
		return nil, nil
	case 1:
		return elsewhere[0], nil
	}
	namespaces := make([]string, len(elsewhere))
	for i, it := range elsewhere {
		namespaces[i] = it.Metadata.Namespace
	}
	return nil, fmt.Errorf("%w: %d Deployments named %q outside the resolved operator namespace %s (in %s). "+
		"Set kubernetes.operator.namespace to the one this env file means",
		errAmbiguousOperator, len(elsewhere), operatorDeployment, want, strings.Join(namespaces, ", "))
}

// errAmbiguousOperator marks the ONE findOperatorDeployment failure a caller must never
// fold into "no operator installed".
//
// Every other failure here is a read that did not happen -- an RBAC denial, an
// unreachable API server -- and installedOperatorImage deliberately answers "" for those,
// because a first install has no operator namespace yet and must not alarm. Ambiguity is
// the opposite: the operator IS installed, more than once, and this tool cannot say which
// one a deploy would be downgrading. Answering "" there would disable the downgrade
// confirmation using the very condition it exists to catch.
var errAmbiguousOperator = errors.New("cannot tell which Deployment is the operator")

// watchFromDeployment reads the watch scope off an already-fetched operator Deployment.
// A nil dep is "no operator installed", which is found=false rather than an error.
func watchFromDeployment(dep *deploymentItem) (list []string, allNS, found bool) {
	if dep == nil {
		return nil, false, false
	}
	// Whatever its containers say -- including saying nothing at all, which the operator
	// reads as every namespace -- is the answer.
	list, allNS = watchFromContainers(dep.Spec.Template.Spec.Containers)
	return list, allNS, true
}

// watchFromContainers reads WATCH_NAMESPACE out of a pod template's containers.
//
// All three "watches everything" spellings collapse here, which is the point: no
// WATCH_NAMESPACE entry, an entry with an empty value, and an entry whose value is only
// separators and whitespace all mean the operator watches every namespace. Treating them
// alike is also the safe direction -- every caller either leaves an all-namespaces
// operator alone or refuses to narrow it.
func watchFromContainers(containers []containerSpec) (list []string, allNS bool) {
	for _, ctr := range containers {
		for _, env := range ctr.Env {
			if env.Name != watchEnvVar {
				continue
			}
			if parsed := splitWatch(env.Value); len(parsed) > 0 {
				return parsed, false
			}
			return nil, true
		}
	}
	return nil, true
}

// splitWatch parses a WATCH_NAMESPACE value into its namespaces, trimming and
// de-duplicating exactly the way desiredWatch does so the two lists are comparable.
func splitWatch(value string) []string {
	var out []string
	seen := make(map[string]bool)
	for _, ns := range strings.Split(value, ",") {
		ns = strings.TrimSpace(ns)
		if ns == "" || seen[ns] {
			continue
		}
		seen[ns] = true
		out = append(out, ns)
	}
	return out
}

// unionWatch merges desired into installed, preserving installed's ORDER and appending
// only what is genuinely new, in desired's order.
//
// The ordering rule is not cosmetic. The list becomes one env-var value on a Deployment
// pod template, so a reordered but equivalent list is a changed template and rolls the
// operator pod for nothing. Preserving the installed order means an unchanged set
// renders the byte-identical value and the Deployment is left alone.
//
// added is what the union brought in, for the report; empty means nothing changed.
func unionWatch(installed, desired []string) (merged, added []string) {
	have := make(map[string]bool, len(installed))
	merged = make([]string, 0, len(installed)+len(desired))
	for _, ns := range installed {
		have[ns] = true
		merged = append(merged, ns)
	}
	for _, ns := range desired {
		if have[ns] {
			continue
		}
		have[ns] = true
		merged = append(merged, ns)
		added = append(added, ns)
	}
	return merged, added
}

// subtractWatch removes desired from installed, preserving installed's order.
//
// remaining is what the operator would keep watching; removed is what this env file
// accounted for. A remaining of length zero is the interesting case and the caller MUST
// act on it: it means the env file covers every namespace the operator watches, so the
// operator should be DELETED. It must never be applied as an empty WATCH_NAMESPACE,
// which would silently promote the operator to watching every namespace in the cluster
// -- a privilege escalation dressed as a removal.
func subtractWatch(installed, desired []string) (remaining, removed []string) {
	drop := make(map[string]bool, len(desired))
	for _, ns := range desired {
		drop[ns] = true
	}
	for _, ns := range installed {
		if drop[ns] {
			removed = append(removed, ns)
			continue
		}
		remaining = append(remaining, ns)
	}
	return remaining, removed
}

// watchPlan is what reconcileWatch decided, so OperatorApply can act on it and report
// it without recomputing anything.
type watchPlan struct {
	// value is the WATCH_NAMESPACE to render: "" means every namespace.
	value string
	// widening is true when this plan takes an operator from a specific list to
	// watching every namespace. That is a privilege increase, so it is confirmed.
	widening bool
}

// reconcileWatch decides what WATCH_NAMESPACE `operator deploy` should apply.
//
//	installed        desired      -> applied
//	not installed    anything     -> desired (a first install has nothing to preserve)
//	all namespaces   anything     -> unchanged (already a superset of every list)
//	list L           empty (ALL)  -> all namespaces, CONFIRMED (widening)
//	list L           list D       -> L union D
func (c *Cluster) reconcileWatch(ctx context.Context) (watchPlan, error) {
	desired := desiredWatch(c.Cfg)
	installed, allNS, found, err := c.installedWatch(ctx)
	if err != nil {
		return watchPlan{}, err
	}
	switch {
	case !found:
		if len(desired) == 0 {
			c.logf("installing the operator to watch ALL namespaces")
		} else {
			c.logf("installing the operator to watch %s (%d namespace(s))",
				strings.Join(desired, ","), len(desired))
		}
		return watchPlan{value: strings.Join(desired, ",")}, nil

	case allNS:
		// Nothing to merge into a superset. Said out loud because an operator expecting
		// their namespace to be "added" should learn it was already covered.
		c.progress().OK("the operator already watches ALL namespaces; watch list unchanged.")
		return watchPlan{value: ""}, nil

	case len(desired) == 0:
		// The env file asks for every namespace and the operator currently watches a
		// specific list. Applying that WIDENS what a shared operator reconciles, so it
		// is a question, not a side effect.
		c.progress().Warn("this env file asks the operator to watch ALL namespaces; it currently "+
			"watches only %s.", strings.Join(installed, ","))
		return watchPlan{value: "", widening: true}, nil

	default:
		merged, added := unionWatch(installed, desired)
		if len(added) == 0 {
			c.progress().OK("operator watch list unchanged (%d namespace(s)).", len(merged))
		} else {
			c.logf("widening the operator watch list: + %s", strings.Join(added, ","))
		}
		return watchPlan{value: strings.Join(merged, ",")}, nil
	}
}

// OperatorRelease is what `operator remove` calls: it releases THIS env file's claim on
// the operator, and deletes the install only when that claim covers everything the
// operator watches.
//
//	installed        condition           -> action
//	not installed    --                  -> nothing, reported
//	all namespaces   --                  -> operator KEPT: "all except one" is not
//	                                       expressible, so narrowing is impossible and
//	                                       deleting would take every other broker's
//	                                       operator with it
//	list L           L minus D is empty  -> full delete (this env file accounts for
//	                                       every namespace it watches)
//	list L           L minus D is not    -> narrow to L minus D; operator stays up for
//	                                       the brokers in the namespaces that remain
//
// The empty-remainder branch deletes rather than writing an empty watch list, which is
// the trap this function exists to avoid: an empty WATCH_NAMESPACE means "watch every
// namespace", so a removal that wrote one would leave the operator with WIDER scope than
// it started with.
func (c *Cluster) OperatorRelease(ctx context.Context, deleteCRDs bool) error {
	installed, allNS, found, err := c.installedWatch(ctx)
	if err != nil {
		return err
	}
	opNS := c.operatorNS(ctx)

	if !found && !c.isEcho() {
		c.progress().Warn("no operator is installed in this cluster; nothing to remove.")
		return nil
	}
	if allNS && !c.isEcho() {
		c.progress().Warn("the operator watches ALL namespaces, so it cannot be narrowed to exclude %q.",
			c.Cfg.K8s.Namespace)
		c.logf("operator kept. To remove it anyway, set kubernetes.operator.watchNamespaces to match "+
			"what it watches, or delete it by hand: kubectl delete deployment %s -n %s",
			operatorDeployment, opNS)
		return nil
	}

	// A preview has no cluster to answer, so installedWatch reports nothing found. Show
	// the full-delete path, which is the one an operator previewing a removal means.
	remaining, removed := subtractWatch(installed, desiredWatch(c.Cfg))
	if len(remaining) == 0 {
		if len(removed) > 0 {
			c.logf("this env file accounts for every namespace the operator watches (%s); removing the operator",
				strings.Join(removed, ","))
		}
		return c.OperatorDelete(ctx, deleteCRDs)
	}

	c.progress().Warn("the operator also watches %s, so it is KEPT for those namespaces.",
		strings.Join(remaining, ","))
	return c.setWatch(ctx, opNS, remaining)
}

// setWatch rewrites the installed operator's WATCH_NAMESPACE in place.
//
// A TARGETED edit, deliberately not a bundle re-apply. Re-rendering the bundle would
// also re-apply the operator IMAGE from this env file, and this is the removal path: an
// env file letting go of one namespace has no business changing the operator version
// another env file's broker depends on. `set env` touches one field, cannot carry an
// image with it, and rolls the Deployment exactly once.
//
// It refuses an empty value outright. An empty WATCH_NAMESPACE means "watch every
// namespace", so writing one here would turn a narrowing into a cluster-wide widening;
// the caller is expected to delete the operator in that case instead. The guard is here
// as well as at the call site because this is the function that does the writing.
func (c *Cluster) setWatch(ctx context.Context, opNS string, list []string) error {
	if len(list) == 0 {
		return fmt.Errorf("refusing to set an empty %s: an empty value makes the operator watch EVERY "+
			"namespace, which would widen its scope instead of narrowing it.\n"+
			"  Remove the operator itself if no namespace is left to watch", watchEnvVar)
	}
	if err := c.PreflightAll(ctx, probe{verb: "patch", resource: "deployments", ns: opNS}); err != nil {
		return err
	}
	value := strings.Join(list, ",")
	c.logf("narrowing the operator watch list to %s", value)
	return c.kubectl(ctx, "set", "env", operatorDeployRef, "-n", opNS,
		watchEnvVar+"="+value)
}

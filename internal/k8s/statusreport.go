package k8s

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"solace/internal/output"
)

// This file renders the status report from the decoded objects in inspect.go,
// rather than passing kubectl's own tables through.
//
// Owning the rendering buys four things kubectl's table cannot give:
//
//   - a verdict per row, so "is this healthy" is answered rather than left to be
//     inferred from columns;
//   - redundancy ORDER (primary, backup, monitor) instead of kubectl's
//     alphabetical, which lists backup first and reads as arbitrary;
//   - the columns that matter and not the ones that never do -- a broker's pod
//     table always has NOMINATED NODE and READINESS GATES set to <none>, 34
//     characters of nothing on every line;
//   - one voice shared with the rest of the tool.
//
// What it costs is that this rendering can be incomplete in a way kubectl's is
// not. There is deliberately no `--raw` passthrough flag for that: a second view
// of the same data is a second thing to keep correct, and `kubectl get` is always
// right there for anyone who wants the source. What this file owes in exchange is
// that every health rule degrades to "not reported" rather than to a false alarm
// when the cluster shows it something it does not recognise.

// statusKeyWidth is the key column of the status summary block.
const statusKeyWidth = 16

// BrokerReport renders the broker's own state: the operator's verdict on it from
// the CR's conditions, then the pods, services, StatefulSets, and -- with detail
// -- the static artifacts it is built from.
//
// Every read is BY NAME, of what this env file describes: its broker CR and the objects
// the operator builds for it. A namespace shared with other brokers or other teams'
// workloads contributes none of theirs, and each read needs only `get` on its kind, where
// listing the namespace needed `list` -- so the identity a deploy needs is enough to report
// on it. The CR read ends the report when it fails: without it there is no broker to report
// on, and it is also the read that finds an unreachable server. A section whose own read the
// cluster REFUSES says so in a [SKIP] row and the report goes on -- see unreadSection.
func (c *Cluster) BrokerReport(ctx context.Context, detail bool) error {
	r := c.report()

	if c.isEcho() {
		// Nothing to decode in a preview; echo the reads so the shape is visible.
		return c.echoStatusReads(ctx, detail)
	}

	// The broker CR is read ONCE and threaded down. Both the summary and the pod
	// section need it -- the summary for the operator's conditions, the pods for
	// status.podsList -- and fetching it twice would double a cluster round-trip
	// for data that cannot change meaningfully between two calls in one report.
	var brokers brokerList
	if err := c.getJSON(ctx, &brokers, c.brokerRead()...); err != nil {
		return fmt.Errorf("reading the broker resource in namespace %q: %w\n  Reading broker %s needs get on %s there",
			c.ns(), err, c.Cfg.K8s.Name, brokerResource)
	}

	c.brokerSummary(r, brokers)
	r.Line("")
	if err := c.podSection(ctx, r, detail, ownedPods(brokers)); err != nil {
		return err
	}
	r.Line("")
	if err := c.serviceSection(ctx, r); err != nil {
		return err
	}
	r.Line("")
	if err := c.statefulSetSection(ctx, r); err != nil {
		return err
	}
	if !detail {
		return nil
	}
	r.Line("")
	return c.storageSection(ctx, r)
}

// brokerRead is the `get` arguments of this env file's broker CR, read by name.
func (c *Cluster) brokerRead() []string {
	return []string{"pubsubpluseventbrokers", c.Cfg.K8s.Name, "-n", c.ns(), "--ignore-not-found"}
}

// namedRead is the `get` arguments for the named objects of one kind in the broker
// namespace. --ignore-not-found makes an object that is not there an absence the section
// reports, rather than a failed read.
func (c *Cluster) namedRead(kind string, names []string) []string {
	return append(append([]string{kind}, names...), "-n", c.ns(), "--ignore-not-found")
}

// statusPodNames are the pods the report reads: the one pod per role a broker of this env
// file's shape runs. They come from the env file rather than the CR's podsList on purpose: a
// name the cluster supplied would reach kubectl's argv, where one starting with a dash is a
// flag, and every name this tool puts there is held to a grammar it controls. The podsList
// still decides which of the pods read are the broker's (ownsPod).
func (c *Cluster) statusPodNames() []string {
	var names []string
	for _, role := range HARoles(c.Cfg) {
		names = append(names, podName(c.Cfg, role))
	}
	return names
}

// statusStatefulSetNames and statusClaimNames are the per-role objects the operator builds.
func (c *Cluster) statusStatefulSetNames() []string {
	var names []string
	for _, role := range HARoles(c.Cfg) {
		names = append(names, stsName(c.Cfg, role))
	}
	return names
}

func (c *Cluster) statusClaimNames() []string {
	var names []string
	for _, role := range HARoles(c.Cfg) {
		names = append(names, pvcName(c.Cfg, role))
	}
	return names
}

// unreadSection decides what a section's failed read means. A read the cluster REFUSED -- an
// identity granted the broker but not this kind -- draws the section with one [SKIP] row
// naming the grant that reads it, and the report goes on: visibly missing, never quietly
// absent, since a health report that silently dropped its pods would read as no problems
// found. Anything else -- a fault, or a permission question nobody answered -- ends the report
// with the error as it always did, so a script gating on the exit status still sees it.
func (c *Cluster) unreadSection(ctx context.Context, r *output.Sink, title, key, what, resource string, err error) error {
	if allowed, answered := c.canI(ctx, probe{verb: "get", resource: resource}); answered && !allowed {
		r.Section(title)
		r.KVRowAt(output.LevelSkip, statusKeyWidth, key,
			fmt.Sprintf("not read -- this identity may not get %s in %s", resource, c.ns()))
		return nil
	}
	return fmt.Errorf("reading %s in namespace %q: %w", what, c.ns(), err)
}

// brokerSummary is the top block: what the operator itself says about this
// broker. Its conditions are the most authoritative health signal available --
// the operator reconciles the thing, so its own verdict beats anything inferred
// from pod counts.
func (c *Cluster) brokerSummary(r *output.Sink, list brokerList) {
	r.Section("Broker")
	if len(list.Items) == 0 {
		r.KVRowAt(output.LevelWarn, statusKeyWidth, "broker", fmt.Sprintf(
			"no PubSubPlusEventBroker %s in namespace %s (broker deploy creates it)", c.Cfg.K8s.Name, c.ns()))
		return
	}
	for _, it := range list.Items {
		r.KVRowAt(output.LevelInfo, statusKeyWidth, "name", it.Metadata.Name)
		r.KVRowAt(output.LevelInfo, statusKeyWidth, "namespace", it.Metadata.Namespace)
		r.KVRowAt(output.LevelInfo, statusKeyWidth, "image", orNone(it.Status.Broker.BrokerImage))
		r.KVRowAt(output.LevelInfo, statusKeyWidth, "age", age(it.Metadata.CreationTimestamp, c.now()))

		ha := it.Status.Broker.HADeployment == "true"
		mode := "standalone"
		if ha {
			mode = "HA redundancy group"
		}
		r.KVRowAt(output.LevelInfo, statusKeyWidth, "redundancy", mode)

		// The operator's own conditions, in the order that answers an operator's
		// questions: reachable, whole, complaining.
		for _, row := range brokerConditionRows(it.Status.Conditions, ha) {
			r.KVRowAt(row.Level, statusKeyWidth, row.Key, row.Value)
		}
	}
}

// podSection lists the broker's pods. The CR's own podsList decides membership
// when it is published: the operator knows which pods are its own, so a co-tenant
// workload whose name happens to carry the infix cannot be mistaken for a broker.
func (c *Cluster) podSection(ctx context.Context, r *output.Sink, detail bool, own map[string]bool) error {
	var list podList
	if err := c.getJSON(ctx, &list, c.namedRead("pods", c.statusPodNames())...); err != nil {
		return c.unreadSection(ctx, r, "Pods", "pods", "pods", "pods", err)
	}
	type entry struct {
		name string
		row  output.Row
	}
	var entries []entry
	for i := range list.Items {
		it := list.Items[i]
		if !ownsPod(own, it.Metadata.Name) {
			continue
		}
		ready, total, restarts := list.at(i)
		cells := []string{
			it.Metadata.Name,
			fmt.Sprintf("%d/%d", ready, total),
			it.Status.Phase,
			fmt.Sprintf("%d", restarts),
			age(it.Metadata.CreationTimestamp, c.now()),
		}
		if detail {
			cells = append(cells, orNone(it.Status.PodIP), orNone(it.Spec.NodeName))
		}
		entries = append(entries, entry{it.Metadata.Name, output.Row{
			Level: podHealth(it.Status.Phase, ready, total, restarts),
			Cells: cells,
		}})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		ri, rj := roleRank(entries[i].name), roleRank(entries[j].name)
		if ri != rj {
			return ri < rj
		}
		return entries[i].name < entries[j].name
	})

	header := []string{"NAME", "READY", "STATUS", "RESTARTS", "AGE"}
	if detail {
		header = append(header, "IP", "NODE")
	}
	r.Section("Pods")
	if len(entries) == 0 {
		r.KVRowAt(output.LevelWarn, statusKeyWidth, "pods", "none found for this broker")
		return nil
	}
	rows := make([]output.Row, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, e.row)
	}
	r.TaggedTable(header, rows)
	return nil
}

// ownedPods is the set of pod names the broker CR claims, or nil when the
// operator publishes none -- in which case ownsPod falls back to the name infix.
// A pure function over the list BrokerReport already fetched, so membership costs
// no extra round-trip.
func ownedPods(list brokerList) map[string]bool {
	own := map[string]bool{}
	for _, it := range list.Items {
		for _, p := range it.Status.PodsList {
			own[p] = true
		}
	}
	if len(own) == 0 {
		return nil
	}
	return own
}

// ownsPod reports whether a pod belongs to this broker. The CR's list is
// authoritative when present; without it, the `-pubsubplus-` infix is the same
// filter the passthrough survey used, and excludes the operator's own pod
// (pubsubplus-eventbroker-operator-*, which has no leading dash).
func ownsPod(own map[string]bool, name string) bool {
	if own != nil {
		return own[name]
	}
	return strings.Contains(name, "-pubsubplus-")
}

// serviceSection lists the broker's services with the address a client would
// actually connect to. A LoadBalancer with no address yet is the one genuinely
// useful warning here -- it is the difference between "deployed" and "reachable".
func (c *Cluster) serviceSection(ctx context.Context, r *output.Sink) error {
	var list serviceList
	if err := c.getJSON(ctx, &list, c.namedRead("svc", []string{lbServiceName(c.Cfg), discoveryServiceName(c.Cfg)})...); err != nil {
		return c.unreadSection(ctx, r, "Services", "services", "services", "services", err)
	}
	r.Section("Services")
	var rows []output.Row
	for _, it := range list.Items {
		if !strings.Contains(it.Metadata.Name, "pubsubplus") {
			continue
		}
		lvl, addr := serviceAddress(it.Spec.Type, it.Spec.ClusterIP, it.Status.LoadBalancer.Ingress)
		rows = append(rows, output.Row{Level: lvl, Cells: []string{
			it.Metadata.Name, it.Spec.Type, addr, age(it.Metadata.CreationTimestamp, c.now()),
		}})
	}
	if len(rows) == 0 {
		r.KVRowAt(output.LevelWarn, statusKeyWidth, "services", "none found for this broker")
		return nil
	}
	r.TaggedTable([]string{"NAME", "TYPE", "ADDRESS", "AGE"}, rows)
	return nil
}

// statefulSetSection lists the per-role StatefulSets, in redundancy order.
func (c *Cluster) statefulSetSection(ctx context.Context, r *output.Sink) error {
	var list statefulSetList
	if err := c.getJSON(ctx, &list, c.namedRead("statefulsets", c.statusStatefulSetNames())...); err != nil {
		return c.unreadSection(ctx, r, "StatefulSets", "statefulsets", "statefulsets", "statefulsets", err)
	}
	type entry struct {
		name string
		row  output.Row
	}
	var entries []entry
	for _, it := range list.Items {
		if !strings.Contains(it.Metadata.Name, "-pubsubplus-") {
			continue
		}
		lvl, text := replicaLevel(it.Status.ReadyReplicas, it.Spec.Replicas)
		entries = append(entries, entry{it.Metadata.Name, output.Row{Level: lvl, Cells: []string{
			it.Metadata.Name, text, age(it.Metadata.CreationTimestamp, c.now()),
		}}})
	}
	sort.SliceStable(entries, func(i, j int) bool { return roleRank(entries[i].name) < roleRank(entries[j].name) })

	r.Section("StatefulSets")
	if len(entries) == 0 {
		r.KVRowAt(output.LevelWarn, statusKeyWidth, "statefulsets", "none found for this broker")
		return nil
	}
	rows := make([]output.Row, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, e.row)
	}
	r.TaggedTable([]string{"NAME", "READY", "AGE"}, rows)
	return nil
}

// storageSection is --detail only: the claims a broker's data actually lives on.
// A Pending claim is the usual cause of a broker that never starts, and a claim
// left behind by a removed broker is exactly the thing you only find by looking.
func (c *Cluster) storageSection(ctx context.Context, r *output.Sink) error {
	var list pvcList
	if err := c.getJSON(ctx, &list, c.namedRead("pvc", c.statusClaimNames())...); err != nil {
		return c.unreadSection(ctx, r, "Storage", "claims", "persistent volume claims", "persistentvolumeclaims", err)
	}
	type entry struct {
		name string
		row  output.Row
	}
	var entries []entry
	for _, it := range list.Items {
		if !strings.Contains(it.Metadata.Name, "pubsubplus") {
			continue
		}
		entries = append(entries, entry{it.Metadata.Name, output.Row{
			Level: pvcLevel(it.Status.Phase),
			Cells: []string{
				it.Metadata.Name,
				it.Status.Phase,
				it.Spec.Resources.Requests.Storage,
				orNone(it.Spec.StorageClassName),
				age(it.Metadata.CreationTimestamp, c.now()),
			},
		}})
	}
	sort.SliceStable(entries, func(i, j int) bool { return roleRank(entries[i].name) < roleRank(entries[j].name) })

	r.Section("Storage")
	if len(entries) == 0 {
		r.KVRowAt(output.LevelInfo, statusKeyWidth, "claims", "none found for this broker")
		return nil
	}
	rows := make([]output.Row, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, e.row)
	}
	r.TaggedTable([]string{"NAME", "STATUS", "SIZE", "CLASS", "AGE"}, rows)
	return nil
}

// OperatorReport renders the operator's controller state. The version actually
// running is the point: a cluster can easily carry a newer operator than this
// tool renders, and reconciliation behaviour follows the installed one.
func (c *Cluster) OperatorReport(ctx context.Context) error {
	r := c.report()
	if c.isEcho() {
		return c.kubectl(ctx, "get", "deployment", operatorDeployment, "-n", c.operatorNS(ctx))
	}
	opNS := c.operatorNS(ctx)
	// --ignore-not-found makes an operator that is not there an empty answer, reported below
	// as not found; a read that FAILS is a different fact -- most often an identity that may
	// not read the operator's namespace at all -- and is reported as not read, rather than
	// sending a namespace-scoped user off to install an operator that is already running.
	var list deploymentList
	if err := c.getJSON(ctx, &list, "deployment", operatorDeployment, "-n", opNS, "--ignore-not-found"); err != nil {
		r.Section("Operator")
		r.KVRowAt(output.LevelSkip, statusKeyWidth, "installed",
			fmt.Sprintf("not read (%v) -- reading deployment %s needs get on deployments in %s", err, operatorDeployment, opNS))
		return nil
	}
	r.Section("Operator")
	if len(list.Items) == 0 {
		r.KVRowAt(output.LevelWarn, statusKeyWidth, "installed",
			fmt.Sprintf("not found in namespace %s (operator deploy installs it)", opNS))
		return nil
	}
	for _, it := range list.Items {
		lvl, text := replicaLevel(it.Status.ReadyReplicas, it.Spec.Replicas)
		r.KVRowAt(output.LevelInfo, statusKeyWidth, "namespace", opNS)
		r.KVRowAt(lvl, statusKeyWidth, "controller", text+" ready")
		r.KVRowAt(output.LevelInfo, statusKeyWidth, "image", operatorRunningImage(it.Spec.Template.Spec.Containers))
		r.KVRowAt(output.LevelInfo, statusKeyWidth, "age", age(it.Metadata.CreationTimestamp, c.now()))
	}
	return nil
}

// operatorRunningImage names the image the controller is running, or says so when
// the deployment carries no container this tool can read.
func operatorRunningImage(containers []containerSpec) string {
	if len(containers) == 0 {
		return "(no container reported)"
	}
	return containers[0].Image
}

// echoStatusReads lists the reads the report would make, for the preview path
// where there is no cluster to decode.
func (c *Cluster) echoStatusReads(ctx context.Context, detail bool) error {
	reads := [][]string{
		c.brokerRead(),
		c.namedRead("pods", c.statusPodNames()),
		c.namedRead("svc", []string{lbServiceName(c.Cfg), discoveryServiceName(c.Cfg)}),
		c.namedRead("statefulsets", c.statusStatefulSetNames()),
	}
	if detail {
		reads = append(reads, c.namedRead("pvc", c.statusClaimNames()))
	}
	for _, read := range reads {
		if err := c.kubectl(ctx, append(append([]string{"get"}, read...), "-o", "json")...); err != nil {
			return err
		}
	}
	return nil
}

// ClusterReport is `status broker --all`: every Solace broker in the cluster, not
// just the one this env file names. It answers a different question from
// BrokerReport -- "what is running here" rather than "is my broker healthy" -- so
// it stays one table per broker rather than a section per resource kind, which at
// cluster scale would be unreadable.
//
// Discovery is by resource kind across all namespaces, so a broker deployed by
// somebody else, under a name this config knows nothing about, still appears.
// That is the point of the flag.
func (c *Cluster) ClusterReport(ctx context.Context, detail bool) error {
	r := c.report()
	if c.isEcho() {
		return c.kubectl(ctx, "get", "pubsubpluseventbrokers", "--all-namespaces", "-o", "json")
	}

	// Cluster-wide by request, so a refusal is the answer and the command fails -- there is
	// no narrower view to fall back to that would still be what --all asked for. The error
	// says what the listing needs, and where the env file's own broker can still be seen.
	var list brokerList
	if err := c.getJSON(ctx, &list, "pubsubpluseventbrokers", "--all-namespaces"); err != nil {
		return fmt.Errorf("listing brokers across the cluster: %w\n"+
			"  --all needs list on %s in every namespace (a cluster role binding); "+
			"`broker status` without --all reports this env file's broker with namespace rights alone", err, brokerResource)
	}
	r.Section("Brokers in this cluster")
	if len(list.Items) == 0 {
		r.KVRowAt(output.LevelInfo, statusKeyWidth, "brokers", "none found in any namespace")
		return nil
	}

	rows := make([]output.Row, 0, len(list.Items))
	for _, it := range list.Items {
		ha := it.Status.Broker.HADeployment == "true"
		mode := "standalone"
		if ha {
			mode = "HA"
		}
		// One verdict per broker: the worst of its conditions, so a cluster-wide
		// listing surfaces the one that needs attention rather than averaging it
		// away. But the seed itself must not claim health nobody checked: a
		// broker the operator has published NO conditions for -- freshly
		// deployed, or an operator wedged before it wrote any status -- has to
		// read as unreported rather than as the same OK/"ready" a verified
		// broker gets, since brokerConditionRows only ever RAISES the seed and
		// LevelInfo ("not reported by the operator") sits below LevelOK.
		//
		// "No conditions" is not the only unreported case: a status carrying only
		// conditions this tool does not recognise (an operator newer than this build)
		// raised nothing either, and read as the same confident "ready". Both now
		// resolve to Info.
		lvl := output.LevelOK
		detailText := "ready"
		if !anyKnownCondition(it.Status.Conditions) {
			lvl, detailText = output.LevelInfo, "not yet reported by the operator"
		} else {
			for _, row := range brokerConditionRows(it.Status.Conditions, ha) {
				if row.Level > lvl {
					lvl, detailText = row.Level, row.Key+": "+row.Value
				}
			}
		}
		cells := []string{
			it.Metadata.Namespace, it.Metadata.Name, mode,
			age(it.Metadata.CreationTimestamp, c.now()), detailText,
		}
		if detail {
			cells = append(cells[:4], orNone(it.Status.Broker.BrokerImage), detailText)
		}
		rows = append(rows, output.Row{Level: lvl, Cells: cells})
	}
	header := []string{"NAMESPACE", "NAME", "MODE", "AGE", "STATE"}
	if detail {
		header = []string{"NAMESPACE", "NAME", "MODE", "AGE", "IMAGE", "STATE"}
	}
	r.TaggedTable(header, rows)
	return nil
}

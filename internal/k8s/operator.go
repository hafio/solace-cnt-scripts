package k8s

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"solace/internal/config"
)

// operatorBundle is the full operator install manifest (CRDs, RBAC, controller
// Deployment) as a Go text/template with six substitution points. It is the Go port
// of the ~119 KB heredoc embedded in 010-deploy-operator.sh, applied on stdin.
//
//go:embed assets/operator-1.4.2.yaml.tmpl
var operatorBundle string

// operatorTmplVars are the six substitution points in operatorBundle. See
// RenderOperator for how each is derived from config.
type operatorTmplVars struct {
	Namespace      string // operator namespace (appears 6x in the bundle)
	WatchNamespace string // WATCH_NAMESPACE env value (watch list, broker ns appended)
	Image          string // operator image, registry-prefixed when Image.Registry is set
	CPU            string // manager container cpu limit
	Mem            string // manager container memory limit
	PullSecret     bool   // true -> emit the imagePullSecrets: regcred block
}

// RenderOperator renders the operator bundle for namespace opNS, porting the heredoc
// substitutions of 010-deploy-operator.sh: the operator image is prefixed with
// Image.Registry/ when set (010:2019); the imagePullSecrets block is emitted only when
// an image-pull secret is configured.
//
// watch is the WATCH_NAMESPACE value to substitute, and is a PARAMETER rather than
// something derived here from cfg. That is the whole point: the operator is
// cluster-scoped and shared between env files, so an install must apply the value
// reconciled against what is already running (watch.go, reconcileWatch), while a render
// for `operator generate` applies what this env file alone describes (watchNamespace).
// Deriving it inside this function is what made a second env file's deploy silently
// un-watch the first broker's namespace.
//
// The bundle carries NO secret value. The imagePullSecrets block it emits is a
// reference by name; the regcred Secret itself is a separate artifact, spliced into the
// stream by GenOperator and applied by OperatorApply between the namespace and the
// rest of the bundle.
//
// The one ordering constraint it must respect: the regcred is namespaced, and the
// namespace only exists inside this bundle -- applying the secret before the
// namespace is what once failed a first install with `namespaces "solace-operator"
// not found`. OperatorApply therefore applies the bundle's own Namespace document
// first (splitAfterNamespace), never a separately-invented one.
func RenderOperator(cfg *config.Config, opNS, watch string) ([]byte, error) {
	op := cfg.K8s.Operator
	vars := operatorTmplVars{
		Namespace:      opNS,
		WatchNamespace: watch,
		Image:          operatorImage(cfg),
		CPU:            op.CPU,
		Mem:            op.Mem,
		PullSecret:     cfg.K8s.ImagePullSecret != "",
	}
	t, err := template.New("operator").Parse(operatorBundle)
	if err != nil {
		return nil, fmt.Errorf("parse operator bundle template: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, vars); err != nil {
		return nil, fmt.Errorf("render operator bundle: %w", err)
	}
	return buf.Bytes(), nil
}

// splitAfterNamespace cuts the bundle in two immediately after its Namespace
// document, so a caller can apply the namespace, then something namespaced, then
// the rest. If no Namespace document is found ns is nil and the whole bundle is
// returned as rest -- the caller decides what to do about that rather than having
// an invented namespace applied on its behalf.
func splitAfterNamespace(bundle []byte) (ns, rest []byte) {
	docs := yamlDocSepRE.Split(string(bundle), -1)
	at := 0
	for i, d := range docs {
		if namespaceKindRE.MatchString(d) {
			at = i + 1
			break
		}
	}
	return joinYAMLDocs(nonEmpty(docs[:at])), joinYAMLDocs(nonEmpty(docs[at:]))
}

// nonEmpty drops whitespace-only documents, which the separator split produces at
// the ends of a manifest.
func nonEmpty(docs []string) []string {
	out := make([]string, 0, len(docs))
	for _, d := range docs {
		if strings.TrimSpace(d) != "" {
			out = append(out, d)
		}
	}
	return out
}

// namespaceKindRE matches a document whose own `kind:` is Namespace, anchored to
// column 0 for the same reason crdKindRE is: nested mapping keys are indented.
var namespaceKindRE = regexp.MustCompile(`(?m)^kind:[ \t]*Namespace[ \t]*$`)

// GenOperator renders everything `operator deploy` applies, as one stream in APPLY
// ORDER, without contacting the cluster: the bundle's own Namespace document, then the
// image-pull secret when one is configured, then the rest of the bundle.
//
// One command, not two. There used to be a `generate secrets operator` beside this, so
// the credential-bearing document could be reviewed separately -- but the two outputs
// then had to be applied in the right order by hand, and the ordering is exactly the
// thing that is easy to get wrong (the regcred is namespaced, and its namespace only
// exists inside the bundle). Emitting the same three documents in the same order
// OperatorApply uses makes this output the artifact rather than a description of one:
// `solace-util operator generate | kubectl apply -f -` is equivalent to
// `solace-util operator deploy`, minus the preflight and the watch-list reconciliation.
//
// It therefore CAN carry a credential -- the dockerconfigjson of the image-pull secret --
// which is true of the Kubernetes broker artifact too, and is why kubernetes is the only
// platform whose generate output is secret-bearing. On docker and podman a secret cannot
// be part of the artifact at all.
//
// The WATCH_NAMESPACE rendered here is what THIS env file describes. A real deploy
// applies the union with whatever the running operator already watches, so on a cluster
// that already has an operator this output can differ from what deploy would apply --
// deliberately, because a render has no cluster to ask.
func GenOperator(cfg *config.Config) ([]byte, error) {
	opNS := renderOperatorNS(cfg)
	bundle, err := RenderOperator(cfg, opNS, watchNamespace(cfg))
	if err != nil {
		return nil, err
	}
	if cfg.K8s.ImagePullSecret == "" {
		return bundle, nil
	}
	ns, rest := splitAfterNamespace(bundle)
	if len(ns) == 0 {
		return nil, fmt.Errorf("operator bundle has no Namespace document to apply first")
	}
	regcred, err := operatorRegcred(cfg, opNS)
	if err != nil {
		return nil, fmt.Errorf("build operator regcred: %w", err)
	}
	return joinYAMLDocs([]string{string(ns), string(regcred), string(rest)}), nil
}

// renderOperatorNS resolves the operator namespace for a render: the configured
// Operator.Namespace, falling back to the fixed default when unset. The
// running-deployment discovery of operatorNS needs a live cluster, so render-only
// cannot use it.
func renderOperatorNS(cfg *config.Config) string {
	if ns := cfg.K8s.Operator.Namespace; ns != "" {
		return ns
	}
	return defaultOperatorNS
}

// operatorImage is the operator image reference a deploy will actually pull:
// Operator.Image with Image.Registry/ prefixed when set (010:2019). RenderOperator
// substitutes it into the bundle and CheckEnv reports it, both from this one definition,
// so the report cannot name an image the apply does not use. Image.Ref() is not the
// helper for this: it composes repo:tag, which Operator.Image already carries.
func operatorImage(cfg *config.Config) string {
	if cfg.Image.Registry != "" {
		return cfg.Image.Registry + "/" + cfg.K8s.Operator.Image
	}
	return cfg.K8s.Operator.Image
}

// watchNamespace and the rest of the watch-scope handling live in watch.go: the value is
// reconciled against the running operator rather than rendered from config alone,
// because the operator is shared between env files.

// OperatorApply installs the operator in three applies, in dependency order: the
// bundle's own Namespace document, the image-pull secret (regcred) into that
// namespace when pull creds are configured (010:29), then the rest of the bundle
// -- CRDs, RBAC and the controller Deployment (010:2063).
//
// Three applies rather than one multi-document stream because the secret is not in
// the bundle: everything this installs is byte-for-byte what `generate operator` and
// `generate secrets operator` print, which is the property that makes a rendered
// artifact worth reviewing. The namespace still goes first, and the secret still
// lands before the Deployment that pulls with it, so neither the first-install
// ordering nor the pull is affected.
func (c *Cluster) OperatorApply(ctx context.Context) error {
	opNS := c.operatorNS(ctx)
	// EVERY kind the bundle applies, not one representative. Probing only the CRD
	// meant an identity allowed to create custom resource definitions but not
	// ClusterRoleBindings passed this check and died halfway through the apply --
	// which is the exact state Preflight exists to prevent.
	if err := c.PreflightAll(ctx, operatorProbes("create", opNS, c.Cfg.K8s.ImagePullSecret != "")...); err != nil {
		return err
	}
	// Before anything is applied: the operator is cluster-scoped, and `apply`
	// downgrades an image without comment, so a downgrade is asked about rather
	// than merely narrated. Declining aborts before the first write.
	if err := c.confirmNoDowngrade(ctx, opNS); err != nil {
		return err
	}
	// What the operator should watch is decided against what it ALREADY watches, not
	// rendered from this env file alone -- see watch.go. Also before the first write,
	// because the widening case is a question.
	plan, err := c.reconcileWatch(ctx)
	if err != nil {
		return err
	}
	if plan.widening && !c.confirm("Let the operator watch EVERY namespace in this cluster?") {
		return fmt.Errorf("declined: the operator's watch list was left as it is, and nothing was applied")
	}
	c.logf("deploying operator to namespace %s", opNS)
	manifest, err := RenderOperator(c.Cfg, opNS, plan.value)
	if err != nil {
		return err
	}
	ns, rest := splitAfterNamespace(manifest)
	if len(ns) == 0 {
		// The bundle has always carried its own Namespace document; a build that
		// lost it would otherwise apply the regcred into a namespace that does not
		// exist yet, which is the exact first-install failure this ordering exists
		// to prevent.
		return fmt.Errorf("operator bundle has no Namespace document to apply first")
	}
	if err := c.apply(ctx, ns); err != nil {
		return fmt.Errorf("apply operator namespace: %w", err)
	}
	if c.Cfg.K8s.ImagePullSecret != "" {
		regcred, err := operatorRegcred(c.Cfg, opNS)
		if err != nil {
			return fmt.Errorf("build operator regcred: %w", err)
		}
		if err := c.apply(ctx, regcred); err != nil {
			return fmt.Errorf("apply operator image-pull secret: %w", err)
		}
	}
	if err := c.apply(ctx, rest); err != nil {
		return fmt.Errorf("apply operator bundle: %w", err)
	}
	return nil
}

// crdKindRE matches a document whose own `kind:` is CustomResourceDefinition. It is
// anchored to column 0 on purpose: a CRD carries an OpenAPI schema that contains
// nested `kind:` keys of its own, and only the document's top-level mapping keys sit
// unindented.
var crdKindRE = regexp.MustCompile(`(?m)^kind:[ \t]*CustomResourceDefinition[ \t]*$`)

// yamlDocSepRE matches the `---` line separating documents in the rendered bundle.
var yamlDocSepRE = regexp.MustCompile(`(?m)^---[ \t]*$`)

// operatorProbes is every kind the operator bundle touches, for verb. It is one list
// so the apply and the delete cannot probe different sets, which is how a check drifts
// from the work it guards.
//
// The namespaced kinds carry opNS explicitly: the operator lives in its own namespace,
// not the broker's, and a probe that asked about the broker's would answer a question
// nobody was asking. The Namespace kind itself appears only on the apply side -- the
// delete no longer removes it.
func operatorProbes(verb, opNS string, withRegcred bool) []probe {
	ps := []probe{
		{verb: verb, resource: "customresourcedefinitions"}, // cluster-scoped
		{verb: verb, resource: "clusterroles"},              // cluster-scoped
		{verb: verb, resource: "clusterrolebindings"},       // cluster-scoped
		{verb: verb, resource: "serviceaccounts", ns: opNS},
		{verb: verb, resource: "roles", ns: opNS},
		{verb: verb, resource: "rolebindings", ns: opNS},
		{verb: verb, resource: "deployments", ns: opNS},
	}
	if verb == "create" {
		// Applying the bundle creates the namespace; deleting it does not remove one.
		ps = append(ps, probe{verb: verb, resource: "namespaces"})
	}
	if withRegcred {
		ps = append(ps, probe{verb: verb, resource: "secrets", ns: opNS})
	}
	return ps
}

// splitOperatorBundle separates the rendered bundle into three classes, because all
// three have different blast radii and none of them should be deleted by a decision
// made about another.
//
//   - crds are the cluster-wide type definitions. Removing them cascade-deletes every
//     PubSubPlusEventBroker in the cluster, including brokers this env file has never
//     heard of, so they are a layer the operator asks for by name.
//   - ns is the operator's Namespace. Deleting a namespace takes EVERYTHING in it,
//     including whatever another team put there, and a namespace is a shared scope
//     this tool did not necessarily create. It is split out so it can be left alone.
//   - rest is the Deployment and RBAC that genuinely belong to this operator install.
//
// The Namespace used to sit in rest, which meant `remove operator` deleted it in the
// same `delete -f -` stream as the Deployment -- cascading to anything else in that
// namespace, with none of the occupancy checks `remove namespace` applies to the
// broker's own.
func splitOperatorBundle(manifest []byte) (crds, ns, rest []byte) {
	var crdDocs, nsDocs, restDocs []string
	for _, doc := range nonEmpty(yamlDocSepRE.Split(string(manifest), -1)) {
		switch {
		case crdKindRE.MatchString(doc):
			crdDocs = append(crdDocs, doc)
		case namespaceKindRE.MatchString(doc):
			nsDocs = append(nsDocs, doc)
		default:
			restDocs = append(restDocs, doc)
		}
	}
	return joinYAMLDocs(crdDocs), joinYAMLDocs(nsDocs), joinYAMLDocs(restDocs)
}

// joinYAMLDocs reassembles documents into one multi-document manifest, or returns
// nil when there are none -- so a caller can test the result for emptiness rather
// than piping a lone separator into kubectl.
func joinYAMLDocs(docs []string) []byte {
	if len(docs) == 0 {
		return nil
	}
	var b strings.Builder
	for i, doc := range docs {
		if i > 0 {
			b.WriteString("\n---\n")
		}
		b.WriteString(strings.Trim(doc, "\n"))
		b.WriteString("\n")
	}
	return []byte(b.String())
}

// BrokerCRs lists every PubSubPlusEventBroker custom resource across ALL
// namespaces, as "<namespace>/<name>" refs. This is the same all-namespaces query
// ClusterReport (statusreport.go) already runs for `status broker --all`; it is
// duplicated here rather than factored into one shared helper because
// statusreport.go is out of scope for this change. Converging the two into a
// single shared helper is a follow-up.
//
// It is exported because the CLI needs the same answer BEFORE it asks its
// CRD-layer question: a question whose answer is already fixed should not be
// asked. That is a better prompt, not a safety mechanism -- the refusal in
// refuseCRDDeleteIfBrokersExist stays the thing that actually stops the cascade.
func (c *Cluster) BrokerCRs(ctx context.Context) ([]string, error) {
	var list brokerList
	if err := c.getJSON(ctx, &list, "pubsubpluseventbrokers", "--all-namespaces"); err != nil {
		return nil, fmt.Errorf("listing broker custom resources across the cluster: %w", err)
	}
	refs := make([]string, 0, len(list.Items))
	for _, it := range list.Items {
		refs = append(refs, it.Metadata.Namespace+"/"+it.Metadata.Name)
	}
	return refs, nil
}

// refuseCRDDeleteIfBrokersExist is the guard OperatorDelete runs immediately
// before it touches the CRD documents. Deleting the CRDs cascade-deletes EVERY
// PubSubPlusEventBroker custom resource in EVERY namespace in the cluster -- not
// just the one this env file describes -- and nothing else warns about that with
// anything more than static prose. So this asks the cluster what actually exists
// first: if anything does, each is named on stderr and the deletion is refused
// outright. There is deliberately no question to ask here (the interactive layer
// question --delete-crd/--no-prompt normally decides is answered already, by the
// caller passing deleteCRDs=true) -- the answer to "delete anyway?" is fixed at
// "no" once a broker this tool does not own would be destroyed, so it is refused
// rather than confirmed.
//
// A listing failure (RBAC, an unreachable API server) refuses too, with the cause
// preserved: being unable to see what would be destroyed is not permission to
// destroy it.
func (c *Cluster) refuseCRDDeleteIfBrokersExist(ctx context.Context) error {
	refs, err := c.BrokerCRs(ctx)
	if err != nil {
		return fmt.Errorf("refusing to delete the operator CRDs: %w", err)
	}
	if len(refs) == 0 {
		return nil
	}
	for _, ref := range refs {
		c.progress().Warn("broker custom resource %s still exists and would be cascade-deleted with the CRDs", ref)
	}
	return fmt.Errorf("refusing to delete the operator CRDs: %d PubSubPlusEventBroker custom resource(s) still "+
		"exist in the cluster (%s) and deleting the CRDs would cascade-delete them; remove each one first "+
		"(`remove broker` per deployment), then retry --delete-crd", len(refs), strings.Join(refs, ", "))
}

// OperatorDelete removes the operator by deleting the rendered bundle on stdin with
// --ignore-not-found (110:2057) -- one mirrored path with OperatorApply, replacing the
// separately-maintained delete manifest the legacy 110 shipped.
//
// deleteCRDs is the same retained-layer decision `remove broker` makes about PVCs:
// the expensive-to-recreate, easy-to-regret part is kept unless it is asked for by
// name. Here that is the CRDs, whose removal takes every broker in the cluster with
// them. Either outcome is stated rather than left to be inferred.
//
// The operator NAMESPACE is deliberately kept. Deleting a namespace cascades to
// everything in it, including anything another team put there, and this tool may not
// have created it -- so it is not something to remove as a side effect of removing an
// operator. The command to finish the job by hand is logged instead.
//
// That makes the regcred Secret its own delete: it used to be reaped as namespace
// content, so leaving the namespace in place would otherwise leave a registry
// credential behind after every teardown.
//
// The operator Deployment/RBAC removal (rest) always runs before the CRD layer is
// even considered, and its own error return is unaffected by whatever the CRD
// layer decides -- a refused CRD deletion must never look like the whole removal
// was skipped, and must never undo work the bundle delete already did.
// Before any of that, it decides whether the operator should be removed AT ALL. The
// operator is shared: if it watches namespaces this env file does not account for, then
// another broker depends on it, and removing it would stop that broker reconciling. So
// the watch list is subtracted rather than the install deleted outright -- see
// OperatorRelease, which is what the CLI calls and which delegates here only when the
// whole install is genuinely this env file's to remove.
func (c *Cluster) OperatorDelete(ctx context.Context, deleteCRDs bool) error {
	opNS := c.operatorNS(ctx)
	// The same set the apply probes, with `delete`. The Namespace is absent because
	// this command no longer deletes it.
	if err := c.PreflightAll(ctx, operatorProbes("delete", opNS, c.Cfg.K8s.ImagePullSecret != "")...); err != nil {
		return err
	}
	c.logf("deleting operator from namespace %s", opNS)
	// The watch value is irrelevant to a delete -- `delete -f -` matches documents by
	// name and kind, not by field content -- so this renders what the env file describes
	// purely to get the document SET to delete.
	manifest, err := RenderOperator(c.Cfg, opNS, watchNamespace(c.Cfg))
	if err != nil {
		return err
	}
	crds, _, rest := splitOperatorBundle(manifest)
	// The regcred rides the SAME delete stream rather than a call of its own. It is
	// applied separately by OperatorApply and used to disappear along with the
	// namespace, so now that the namespace stays it has to be named explicitly -- and
	// appending its manifest here keeps the teardown one ordered `delete -f -` instead
	// of growing a second round trip.
	if c.Cfg.K8s.ImagePullSecret != "" {
		regcred, err := operatorRegcred(c.Cfg, opNS)
		if err != nil {
			return err
		}
		// nonEmpty so a bundle with no non-CRD documents does not produce a leading
		// separator with nothing before it.
		rest = joinYAMLDocs(nonEmpty([]string{string(rest), string(regcred)}))
	}
	if len(rest) > 0 {
		if err := c.deleteStdin(ctx, rest); err != nil {
			return fmt.Errorf("delete operator bundle: %w", err)
		}
	}
	c.logf("operator namespace %s kept -- deleting it would take everything else in it too; "+
		"remove it by hand once you are sure nothing else lives there (kubectl delete namespace %s)", opNS, opNS)
	if !deleteCRDs {
		c.logf("operator CRDs kept -- existing broker resources are untouched " +
			"(pass --delete-crd to remove them)")
		return nil
	}
	if len(crds) == 0 {
		c.logf("operator CRDs: none in the bundle, nothing to delete")
		return nil
	}
	if err := c.refuseCRDDeleteIfBrokersExist(ctx); err != nil {
		return err
	}
	if err := c.deleteStdin(ctx, crds); err != nil {
		return fmt.Errorf("delete operator CRDs: %w", err)
	}
	c.logf("operator CRDs deleted -- every PubSubPlusEventBroker resource in the cluster went with them")
	return nil
}

// OperatorRestart bounces the controller without changing what is installed, for the
// case where the operator is wedged rather than out of date -- `deploy operator` is
// what re-applies a changed bundle.
func (c *Cluster) OperatorRestart(ctx context.Context) error {
	opNS := c.operatorNS(ctx)
	// opNS explicitly: the restart happens in the OPERATOR's namespace, and this
	// probe used to ask about the broker's -- so an identity permitted in one and not
	// the other passed the check and then failed the work it was meant to guard.
	if err := c.PreflightAll(ctx, probe{verb: "patch", resource: "deployments", ns: opNS}); err != nil {
		return err
	}
	c.logf("restarting operator deployment %s in %s", operatorDeployment, opNS)
	return c.kubectl(ctx, "rollout", "restart", "deployment", operatorDeployment, "-n", opNS)
}

// OperatorScale scales the controller Deployment to replicas, which is what
// `operator stop` (0) and `operator start` (1) do.
//
// Stopping the operator FREEZES reconciliation cluster-wide: every broker it watches
// stops being reconciled until it is started again, and nothing else notices. That is
// occasionally exactly what is wanted -- editing a StatefulSet the operator owns is
// otherwise a fight with the controller -- but it is a cluster-scoped pause, so the
// command says so and the caller confirms it.
//
// Scaling up waits for the rollout; scaling down does not. There is nothing to wait for
// on the way down, and a stop that blocked until the pod was gone would be slower with
// no more certainty -- the same asymmetry ReplicasStart/ReplicasStop already have for the
// broker.
func (c *Cluster) OperatorScale(ctx context.Context, replicas int) error {
	opNS := c.operatorNS(ctx)
	// `scale` is a patch on the Deployment's scale subresource; `patch deployments` is
	// the probe OperatorRestart already uses, so this adds no new RBAC vocabulary.
	if err := c.PreflightAll(ctx, probe{verb: "patch", resource: "deployments", ns: opNS}); err != nil {
		return err
	}
	c.logf("scaling operator deployment %s in %s to %d", operatorDeployment, opNS, replicas)
	if err := c.kubectl(ctx, "scale", "deployment", operatorDeployment, "-n", opNS,
		"--replicas="+strconv.Itoa(replicas)); err != nil {
		return err
	}
	if replicas == 0 {
		c.progress().Warn("the operator is stopped: no broker in any watched namespace is being " +
			"reconciled until `operator start`.")
		return nil
	}
	return c.kubectl(ctx, "rollout", "status", operatorDeployRef, "-n", opNS,
		"--timeout="+rolloutTimeout)
}

// OperatorInstalled reports whether the operator's CRD and controller Deployment are
// both present. It returns a plain bool rather than an error: every way of failing to
// find them -- absent, wrong namespace, cluster unreachable -- leads to the same
// advice, and the only caller uses this to decide whether to warn before a deploy
// that would otherwise fail confusingly. A false here is never fatal on its own.
func (c *Cluster) OperatorInstalled(ctx context.Context) bool {
	if _, err := c.output(ctx, "get", "crd", brokerResource); err != nil {
		return false
	}
	opNS := c.operatorNS(ctx)
	_, err := c.output(ctx, "get", "deployment", operatorDeployment, "-n", opNS)
	return err == nil
}

// OperatorStatus prints the operator Deployment and its controller pods in the
// operator namespace.
func (c *Cluster) OperatorStatus(ctx context.Context) error {
	opNS := c.operatorNS(ctx)
	if err := c.kubectl(ctx, "get", "deployment", operatorDeployment, "-n", opNS, "-o", "wide"); err != nil {
		return err
	}
	return c.kubectl(ctx, "get", "pods", "-n", opNS, "-l", "control-plane=controller-manager", "-o", "wide")
}

// OperatorLogs streams the operator manager logs; passthrough args (e.g. -f, --tail)
// are forwarded verbatim.
func (c *Cluster) OperatorLogs(ctx context.Context, passthrough ...string) error {
	opNS := c.operatorNS(ctx)
	args := append([]string{"logs", "-n", opNS, operatorDeployRef}, passthrough...)
	return c.kubectl(ctx, args...)
}

// OperatorDescribe describes the operator Deployment in the operator namespace.
func (c *Cluster) OperatorDescribe(ctx context.Context) error {
	opNS := c.operatorNS(ctx)
	return c.kubectl(ctx, "describe", operatorDeployRef, "-n", opNS)
}

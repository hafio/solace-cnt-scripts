package k8s

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"solace/internal/config"
)

// operatorBundle is the full operator install manifest (CRDs, RBAC, controller
// Deployment) as a Go text/template with six substitution points. It is the Go port
// of the ~119 KB heredoc embedded in 010-deploy-operator.sh, applied on stdin.
//
//go:embed assets/operator-1.4.0.yaml.tmpl
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
// Image.Registry/ when set (010:2019); WATCH_NAMESPACE is Operator.WatchNamespaces with
// the broker namespace appended when broker-ns watching is enabled (000-env.sh:85-89);
// the imagePullSecrets block is emitted only when an image-pull secret is configured.
//
// The bundle carries NO secret value. The imagePullSecrets block it emits is a
// reference by name; the regcred Secret itself is a separate artifact, rendered by
// GenOperatorSecrets and applied by OperatorApply between the namespace and the
// rest of the bundle. That split is what makes `generate operator` output safe to
// review, diff and share, and it is the same division the broker lifecycle already
// draws between `generate broker` and `generate secrets broker`.
//
// The one ordering constraint it must respect: the regcred is namespaced, and the
// namespace only exists inside this bundle -- applying the secret before the
// namespace is what once failed a first install with `namespaces "solace-operator"
// not found`. OperatorApply therefore applies the bundle's own Namespace document
// first (splitAfterNamespace), never a separately-invented one.
func RenderOperator(cfg *config.Config, opNS string) ([]byte, error) {
	op := cfg.K8s.Operator
	vars := operatorTmplVars{
		Namespace:      opNS,
		WatchNamespace: watchNamespace(cfg),
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

// GenOperator renders the operator bundle for `generate operator` without contacting
// the cluster. It is the artifact-only counterpart to OperatorApply -- together with
// GenOperatorSecrets it is everything that install applies.
func GenOperator(cfg *config.Config) ([]byte, error) {
	return RenderOperator(cfg, renderOperatorNS(cfg))
}

// GenOperatorSecrets renders the operator's image-pull secret on its own, for
// `generate secrets operator`. It is the one operator artifact that carries a
// credential, which is why it has a target of its own rather than riding inside the
// bundle: the sensitive half can be inspected deliberately, and the bundle stays
// shareable.
func GenOperatorSecrets(cfg *config.Config) ([]byte, error) {
	if cfg.K8s.ImagePullSecret == "" {
		return nil, fmt.Errorf("kubernetes.imagePullSecret is not set, so the operator install " +
			"has no image-pull secret to render -- set it (with image.registry, image.user and " +
			"image.pass) if the operator image needs credentials to pull")
	}
	return operatorRegcred(cfg, renderOperatorNS(cfg))
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

// watchNamespace builds the operator's WATCH_NAMESPACE value: the configured watch
// list with the broker namespace appended (comma-joined) when broker-ns watching is
// enabled -- the default (000-env.sh:85-89).
//
// Entries are trimmed and de-duplicated, first occurrence winning. The broker namespace
// is very often already in the configured list, and the repeat reached both the `check`
// report and the applied Deployment's WATCH_NAMESPACE; controller-runtime's cache is
// map-keyed, so it collapsed there harmlessly, which is exactly why it went unnoticed.
func watchNamespace(cfg *config.Config) string {
	var out []string
	seen := make(map[string]bool)
	add := func(ns string) {
		ns = strings.TrimSpace(ns)
		if ns == "" || seen[ns] {
			return
		}
		seen[ns] = true
		out = append(out, ns)
	}
	for _, ns := range strings.Split(cfg.K8s.Operator.WatchNamespaces, ",") {
		add(ns)
	}
	if cfg.K8s.Operator.WatchBrokerNSEnabled() {
		add(cfg.K8s.Namespace)
	}
	return strings.Join(out, ",")
}

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
	// The bundle is cluster-scoped (CRDs, ClusterRoles), so the permission that
	// matters is the one an under-privileged context most often lacks -- and
	// discovering that halfway through a multi-document apply is the worst case.
	if err := c.Preflight(ctx, "create", "customresourcedefinitions"); err != nil {
		return err
	}
	opNS := c.operatorNS(ctx)
	// Before anything is applied: the operator is cluster-scoped, and `apply`
	// downgrades an image without comment, so a downgrade is asked about rather
	// than merely narrated. Declining aborts before the first write.
	if err := c.confirmNoDowngrade(ctx, opNS); err != nil {
		return err
	}
	c.logf("deploying operator to namespace %s", opNS)
	manifest, err := RenderOperator(c.Cfg, opNS)
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

// splitOperatorBundle separates the CustomResourceDefinition documents from
// everything else in the rendered bundle. The two halves are deleted separately
// because they have very different blast radii: the Deployment and RBAC belong to
// this operator install, while the CRDs are the cluster-wide type definitions --
// removing them cascade-deletes every PubSubPlusEventBroker resource in the cluster,
// including brokers this env file has never heard of.
func splitOperatorBundle(manifest []byte) (crds, rest []byte) {
	var crdDocs, restDocs []string
	for _, doc := range nonEmpty(yamlDocSepRE.Split(string(manifest), -1)) {
		if crdKindRE.MatchString(doc) {
			crdDocs = append(crdDocs, doc)
			continue
		}
		restDocs = append(restDocs, doc)
	}
	return joinYAMLDocs(crdDocs), joinYAMLDocs(restDocs)
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
// The regcred needs no delete of its own even though OperatorApply applies it
// separately: it lives in the operator namespace, which is one of the documents
// deleted here, and deleting a namespace takes its contents with it.
//
// The operator Deployment/RBAC removal (rest) always runs before the CRD layer is
// even considered, and its own error return is unaffected by whatever the CRD
// layer decides -- a refused CRD deletion must never look like the whole removal
// was skipped, and must never undo work the bundle delete already did.
func (c *Cluster) OperatorDelete(ctx context.Context, deleteCRDs bool) error {
	if err := c.Preflight(ctx, "delete", "customresourcedefinitions"); err != nil {
		return err
	}
	opNS := c.operatorNS(ctx)
	c.logf("deleting operator from namespace %s", opNS)
	manifest, err := RenderOperator(c.Cfg, opNS)
	if err != nil {
		return err
	}
	crds, rest := splitOperatorBundle(manifest)
	if len(rest) > 0 {
		if err := c.deleteStdin(ctx, rest); err != nil {
			return fmt.Errorf("delete operator bundle: %w", err)
		}
	}
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
	if err := c.Preflight(ctx, "patch", "deployments"); err != nil {
		return err
	}
	c.logf("restarting operator deployment %s in %s", operatorDeployment, opNS)
	return c.kubectl(ctx, "rollout", "restart", "deployment", operatorDeployment, "-n", opNS)
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
	args := append([]string{"logs", "-n", opNS, "deployment/" + operatorDeployment}, passthrough...)
	return c.kubectl(ctx, args...)
}

// OperatorDescribe describes the operator Deployment in the operator namespace.
func (c *Cluster) OperatorDescribe(ctx context.Context) error {
	opNS := c.operatorNS(ctx)
	return c.kubectl(ctx, "describe", "deployment/"+operatorDeployment, "-n", opNS)
}

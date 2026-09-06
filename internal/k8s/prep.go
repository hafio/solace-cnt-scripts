package k8s

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"solace/internal/config"
)

// out returns the report sink, defaulting to stdout when unset.
func (c *Cluster) out() io.Writer {
	if c.Out != nil {
		return c.Out
	}
	return os.Stdout
}

// errOut returns the prompt sink, defaulting to stderr when unset. Prompts are
// deliberately not on Out: a script capturing this tool's stdout wants the
// report, not the questions.
func (c *Cluster) errOut() io.Writer {
	if c.Err != nil {
		return c.Err
	}
	return os.Stderr
}

// in returns the prompt source, defaulting to stdin when unset. Only the
// interactive operations (LabelNodes) read it.
func (c *Cluster) in() io.Reader {
	if c.In != nil {
		return c.In
	}
	return os.Stdin
}

// namespaceManifest is a minimal core/v1 Namespace. Applying it on stdin is the
// idempotent equivalent of the bash `create ns --dry-run=client -o yaml | apply -f -`
// (011:15): apply creates it if absent and no-ops if it already exists.
func namespaceManifest(ns string) []byte {
	return []byte("apiVersion: v1\nkind: Namespace\nmetadata:\n  name: " + ns + "\n")
}

// CreateNamespace applies the broker namespace (011). Idempotent via `apply`.
func (c *Cluster) CreateNamespace(ctx context.Context) error {
	if err := c.Preflight(ctx, "create", "namespaces"); err != nil {
		return err
	}
	c.logf("creating namespace %s", c.ns())
	return c.apply(ctx, namespaceManifest(c.ns()))
}

// protectedNamespaces are the Kubernetes system namespaces DeleteNamespace
// refuses outright, with no override: deleting any of them breaks the
// cluster, and no Solace deployment ever needs it done. An env file naming one
// of these in kubernetes.namespace (default included -- the easiest namespace
// to point at by accident) is a configuration mistake DeleteNamespace must
// catch before it does anything else, cluster round-trip included.
var protectedNamespaces = map[string]bool{
	"default":         true,
	"kube-system":     true,
	"kube-public":     true,
	"kube-node-lease": true,
}

// DeleteNamespace removes the broker namespace (111). --ignore-not-found makes a
// repeat teardown a no-op rather than an error.
//
// Two guards run before the delete, and neither can be silenced by --no-prompt:
// that flag only decides whether the CLI's own confirmDelete asks a question
// before calling here, so this method is exactly what a fully unattended
// removal reaches, with nothing left upstream to skip its checks.
//
//  1. protectedNamespaces, checked first and without even asking the cluster:
//     there is no interpretation of "delete kube-system" that this tool should
//     ever carry out.
//  2. Everything else the namespace holds is enumerated and classified OURS vs
//     FOREIGN (namespaceObjects/isOurs). A namespace an env file points at by
//     mistake -- "kubernetes.namespace: default" would already be caught above,
//     but nothing stops "kubernetes.namespace: shared-team-ns" -- must not take
//     someone else's pods, services and secrets down with it. Any FOREIGN
//     object refuses the whole delete; only an all-ours or empty namespace
//     proceeds to the same behaviour as before this guard existed.
func (c *Cluster) DeleteNamespace(ctx context.Context) error {
	if protectedNamespaces[c.ns()] {
		return fmt.Errorf("refusing to delete namespace %q: this tool will never delete it; "+
			"use \"remove broker\" / \"remove secrets\" to remove just this deployment's own objects", c.ns())
	}
	if err := c.Preflight(ctx, "delete", "namespaces"); err != nil {
		return err
	}
	objs, err := c.namespaceObjects(ctx)
	if err != nil {
		// Being unable to see what the namespace holds is not permission to
		// proceed -- that is the one thing this guard exists to prevent.
		return err
	}
	if foreign := foreignObjects(c.Cfg, objs); len(foreign) > 0 {
		r := c.report()
		r.Section("Foreign resources in namespace " + c.ns())
		rows := make([][]string, len(foreign))
		for i, o := range foreign {
			rows[i] = []string{o.Kind, o.Name}
		}
		r.Table([]string{"KIND", "NAME"}, rows)
		return fmt.Errorf("namespace %q holds %d resource(s) this tool did not create; refusing to delete it. "+
			"\"remove broker\" and \"remove secrets\" remove just this deployment's own objects; deleting the "+
			"namespace itself is then your own `kubectl delete namespace %s` call to make", c.ns(), len(foreign), c.ns())
	}
	c.logf("deleting namespace %s", c.ns())
	return c.kubectl(ctx, "delete", "namespace", c.ns(), "--ignore-not-found")
}

// nsResourceKinds is every built-in kind DeleteNamespace enumerates in ONE
// `kubectl get` call, joined as kubectl's own comma-separated resource-type
// list. The broker CR rides its own separate call in namespaceObjects rather
// than joining this one: kubectl resolves a multi-kind get's resource types
// before it asks the API server anything, and aborts the WHOLE call if even
// one of them is unrecognised -- which the broker CRD, unlike any of these
// built-ins, may genuinely be (an operator never installed on this cluster).
// Splitting them means an absent CRD only fails the one probe that was asking
// about it, instead of also blinding this guard to every ordinary object below.
const nsResourceKinds = "pods,services,statefulsets,deployments,daemonsets,jobs,cronjobs," +
	"persistentvolumeclaims,secrets,configmaps,serviceaccounts,ingresses"

// brokerCRKind is stamped onto every broker CR namespaceObjects decodes. A
// single-kind `kubectl get <kind> -o json` reply does not carry a per-item
// "kind"/"apiVersion" the way a heterogeneous multi-kind reply does (that is
// the whole reason nsResourceKinds' items decode their own Kind field below),
// so the broker CR's items arrive with no kind of their own and this constant
// fills it in by hand instead.
const brokerCRKind = "PubSubPlusEventBroker"

// nsObject is one object found while enumerating a namespace's contents ahead
// of a delete: only the kind and name are needed to classify ownership.
type nsObject struct {
	Kind string
	Name string
}

// nsRawList decodes nsResourceKinds' heterogeneous reply: each item carries
// its own kind, which is what lets one JSON blob describe objects of several
// different built-in types.
type nsRawList struct {
	Items []struct {
		Kind     string     `json:"kind"`
		Metadata objectMeta `json:"metadata"`
	} `json:"items"`
}

// nsCRList decodes the broker-CR-only reply: a homogeneous list, so only the
// name is read -- brokerCRKind supplies the kind nsRawList would otherwise
// have carried per item.
type nsCRList struct {
	Items []struct {
		Metadata objectMeta `json:"metadata"`
	} `json:"items"`
}

// namespaceObjects lists everything DeleteNamespace must classify before it
// tears a namespace down, across the two calls nsResourceKinds' comment
// explains. The two failures are NOT equivalent, and are deliberately not
// treated alike:
//
//   - The built-in listing failing (RBAC denial, an unreachable API server)
//     blinds the guard completely, and being unable to see is not the same as
//     seeing it is safe. It is wrapped and returned, and refuses the delete.
//
//   - The broker-CR listing failing usually means the CRD is not installed on
//     this cluster, which is a state this tool's own documented order produces:
//     remove the brokers, `remove operator --delete-crd`, then `remove
//     namespace`. Refusing there would make a legitimate teardown impossible
//     forever, so it warns and carries on judging the namespace by what it CAN
//     see. That is safe because the built-in listing is what actually catches a
//     foreign broker: the operator gives every one of its objects a
//     "<name>-pubsubplus" name (isOurs), so another team's broker shows up as
//     foreign pods, a foreign StatefulSet, PVCs and a Service even when its CR
//     is invisible here. Only a CR with no workload at all -- freshly created,
//     or wholly failed -- escapes, and it is named in the warning either way.
//
// kubectl's own message cannot be used to tell the two apart: engine.Exec wires
// the child's stderr straight to this process's (runner.go), so the error
// carries an exit status and nothing else.
func (c *Cluster) namespaceObjects(ctx context.Context) ([]nsObject, error) {
	var built nsRawList
	if err := c.getJSON(ctx, &built, nsResourceKinds, "-n", c.ns()); err != nil {
		return nil, fmt.Errorf("listing namespace %q contents before delete: %w", c.ns(), err)
	}
	objs := make([]nsObject, 0, len(built.Items)+1)
	for _, it := range built.Items {
		objs = append(objs, nsObject{Kind: it.Kind, Name: it.Metadata.Name})
	}

	var brokers nsCRList
	if err := c.getJSON(ctx, &brokers, brokerResource, "-n", c.ns()); err != nil {
		c.progress().Warn("could not list broker resources in namespace %s (the operator's CRD may not be "+
			"installed, or access was denied): %v", c.ns(), err)
		c.progress().Warn("judging namespace %s by its other contents only; a broker custom resource with no "+
			"running workload would not be seen.", c.ns())
		return objs, nil
	}
	for _, it := range brokers.Items {
		objs = append(objs, nsObject{Kind: brokerCRKind, Name: it.Metadata.Name})
	}
	return objs, nil
}

// namespaceOursSecrets is the configured secret names this deployment owns,
// skipping the ones left unset -- TLSServerSecret and ImagePullSecret are both
// optional, and an unset field must not be compared against a Secret's name.
func namespaceOursSecrets(cfg *config.Config) map[string]bool {
	ours := map[string]bool{}
	for _, n := range []string{cfg.K8s.AdminSecret, cfg.K8s.TLSServerSecret, cfg.K8s.ImagePullSecret} {
		if n != "" {
			ours[n] = true
		}
	}
	return ours
}

// isOurs classifies one namespace object as belonging to this deployment.
// Everything it does not recognise is FOREIGN -- work DeleteNamespace must
// never take down with the namespace.
func isOurs(cfg *config.Config, ourSecrets map[string]bool, obj nsObject) bool {
	switch {
	case strings.Contains(obj.Name, cfg.K8s.Name+brokerSuffix):
		// The operator names every resource it creates off the broker
		// "<kubernetes.name>-pubsubplus...": pods, StatefulSets, PVCs and the LB
		// service all do (names.go), and by the same rule so would a
		// PodDisruptionBudget or an auto-created ServiceAccount, for which this
		// repo derives no name helper at all.
		//
		// The broker NAME is part of the test, not just the suffix. Matching the
		// bare suffix would call ANY Solace broker's objects ours, so a second
		// team's broker sharing this namespace -- the exact case this guard
		// exists for -- would be classified as ours and deleted with it. The
		// prefix keeps that broker's objects foreign, which refuses the delete.
		return true
	case obj.Kind == brokerCRKind:
		return obj.Name == cfg.K8s.Name
	case obj.Kind == "Secret":
		return ourSecrets[obj.Name]
	case obj.Kind == "ConfigMap":
		// Kubernetes stamps "kube-root-ca.crt" into EVERY namespace it creates,
		// regardless of who created the namespace -- general Kubernetes
		// behaviour, not recorded anywhere else in this repo. Treating it as
		// foreign would refuse every single teardown.
		return obj.Name == "kube-root-ca.crt"
	case obj.Kind == "ServiceAccount":
		// Same Kubernetes-generated exception, for the "default" ServiceAccount
		// every namespace gets automatically.
		return obj.Name == "default"
	default:
		return false
	}
}

// foreignObjects filters objs down to what isOurs does not recognise.
func foreignObjects(cfg *config.Config, objs []nsObject) []nsObject {
	ours := namespaceOursSecrets(cfg)
	var foreign []nsObject
	for _, o := range objs {
		if !isOurs(cfg, ours, o) {
			foreign = append(foreign, o)
		}
	}
	return foreign
}

// secretPreflight fails loud before any manifest is built when the TLS server
// secret is requested but its cert/key inputs are missing, porting the guard of
// 012:19-24 so the operator does not later fail to mount a half-built secret. The
// admin secret's own guards live in AdminSecret.
func (c *Cluster) secretPreflight() error {
	if c.Cfg.K8s.TLSServerSecret == "" {
		return nil
	}
	if c.Cfg.TLS.Cert == "" || c.Cfg.TLS.CertKey == "" {
		return fmt.Errorf("kubernetes.tlsServerSecret %q is set but tls.cert and tls.certKey are not both configured", c.Cfg.K8s.TLSServerSecret)
	}
	for _, f := range []string{c.Cfg.TLS.Cert, c.Cfg.TLS.CertKey} {
		if _, err := os.Stat(f); err != nil {
			return fmt.Errorf("tls certificate input %q is not readable: %w", f, err)
		}
	}
	return nil
}

// GenSecrets builds every applicable secret (admin always; TLS when
// kubernetes.tlsServerSecret is set; the image-pull secret when
// kubernetes.imagePullSecret is set) and
// joins them into one multi-doc manifest -- porting 012's secret set. It is the
// rendering behind both CreateSecrets and `--gen-secrets-only`, so what a user
// reviews is exactly what gets applied. The manifests carry the base64-encoded
// secret values, so the output is as sensitive as the env file it came from.
func GenSecrets(cfg *config.Config) ([]byte, error) {
	docs := make([][]byte, 0, 3)

	admin, err := AdminSecret(cfg)
	if err != nil {
		return nil, err
	}
	docs = append(docs, admin)

	if cfg.K8s.TLSServerSecret != "" {
		tls, err := TLSSecret(cfg)
		if err != nil {
			return nil, err
		}
		docs = append(docs, tls)
	}

	if cfg.K8s.ImagePullSecret != "" {
		pull, err := DockerRegistrySecret(cfg)
		if err != nil {
			return nil, err
		}
		docs = append(docs, pull)
	}
	return joinManifests(docs), nil
}

// CreateSecrets applies the GenSecrets manifest on stdin, keeping every secret
// value off the argv and out of an echoed command (§3). The whole manifest is
// built before the first apply, so a builder error aborts cleanly without leaving
// a partially-applied secret set.
func (c *Cluster) CreateSecrets(ctx context.Context) error {
	if err := c.secretPreflight(); err != nil {
		return err
	}
	// Ahead of GenSecrets, which reads the cert/key files off disk: no reason to
	// load key material into this process for a cluster that will refuse it.
	if err := c.Preflight(ctx, "create", "secrets"); err != nil {
		return err
	}
	manifest, err := GenSecrets(c.Cfg)
	if err != nil {
		return err
	}
	c.logf("creating secrets in %s", c.ns())
	return c.apply(ctx, manifest)
}

// DeleteSecrets removes the secrets CreateSecrets created (112): the admin secret
// always, the TLS and image-pull secrets only when their names are configured. All
// use --ignore-not-found so a partial or repeat teardown is not an error.
func (c *Cluster) DeleteSecrets(ctx context.Context) error {
	if err := c.Preflight(ctx, "delete", "secrets"); err != nil {
		return err
	}
	names := []string{c.Cfg.K8s.AdminSecret}
	if c.Cfg.K8s.TLSServerSecret != "" {
		names = append(names, c.Cfg.K8s.TLSServerSecret)
	}
	if c.Cfg.K8s.ImagePullSecret != "" {
		names = append(names, c.Cfg.K8s.ImagePullSecret)
	}
	for _, name := range names {
		if name == "" {
			continue
		}
		c.logf("deleting secret %s", name)
		if err := c.kubectl(ctx, "delete", "secret", name, "-n", c.ns(), "--ignore-not-found"); err != nil {
			return err
		}
	}
	return nil
}

// UpdateServerCertSecret rebuilds the kubernetes.io/tls secret from the current
// certificate files and applies it on stdin, porting the secret-managed path of
// 051-load-server-cert.sh (051:28-38). Applying on stdin replaces the bash
// `create secret tls --dry-run|apply`, so the private key never reaches an argv or
// an echoed command (§3). The broker re-reads the secret; no pod restart here.
func (c *Cluster) UpdateServerCertSecret(ctx context.Context) error {
	if c.Cfg.K8s.TLSServerSecret == "" {
		return fmt.Errorf("kubernetes.tlsServerSecret must be set to update the server-certificate secret")
	}
	if err := c.Preflight(ctx, "update", "secrets"); err != nil {
		return err
	}
	manifest, err := TLSSecret(c.Cfg)
	if err != nil {
		return err
	}
	c.logf("updating server-certificate secret %s", c.Cfg.K8s.TLSServerSecret)
	return c.apply(ctx, manifest)
}

// joinManifests concatenates rendered YAML documents with a `---` separator so
// they apply as one multi-doc stream.
func joinManifests(docs [][]byte) []byte {
	parts := make([]string, len(docs))
	for i, d := range docs {
		parts[i] = strings.TrimRight(string(d), "\n")
	}
	return []byte(strings.Join(parts, "\n---\n") + "\n")
}

// --- node labelling (013) --------------------------------------------------

// builtinLabelPrefixes are Kubernetes-managed label namespaces that the operator
// never asks the user to set on a node; entries under them are silently skipped so
// the prompt only offers labels the user actually configured (013).
var builtinLabelPrefixes = []string{
	"kubernetes.io/",
	"k8s.io/",
	"node.kubernetes.io/",
	"beta.kubernetes.io/",
}

// isBuiltinLabel reports whether a label key sits under a Kubernetes-managed
// prefix and should not be applied by hand.
func isBuiltinLabel(key string) bool {
	for _, p := range builtinLabelPrefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}

// splitLabel parses a configured node-label entry into a key and value. It accepts
// both the kubectl form `key=value` and the YAML/bash form `key: value`, trimming
// surrounding space; ok is false when either side is empty or no separator is
// present. The bash port only handled `key: value` (013), so `=` support is a
// deliberate convenience.
func splitLabel(entry string) (key, val string, ok bool) {
	var i int
	if i = strings.IndexByte(entry, '='); i < 0 {
		i = strings.IndexByte(entry, ':')
	}
	if i < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(entry[:i])
	val = strings.TrimSpace(entry[i+1:])
	if key == "" || val == "" {
		return "", "", false
	}
	return key, val, true
}

// labelTokenRE constrains a label key/value to Kubernetes' own charset. It is a
// defensive check on config that reaches the kubectl argv: it rejects whitespace,
// shell metacharacters and a leading '-' (which kubectl would read as a flag). §3.
var labelTokenRE = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9._/-]*[A-Za-z0-9])?$`)

func validLabelToken(s string) bool { return labelTokenRE.MatchString(s) }

// labelKV is one validated custom label destined for a node.
type labelKV struct{ key, val string }

// rolePlacementLabels returns the configured node labels for a role.
func rolePlacementLabels(cfg *config.Config, role config.Role) []string {
	switch role {
	case config.Backup:
		return cfg.K8s.Placement.LabelsBackup
	case config.Monitor:
		return cfg.K8s.Placement.LabelsMonitor
	default:
		return cfg.K8s.Placement.LabelsPrimary
	}
}

// roleName is the human label used in the node-selection prompt.
func roleName(role config.Role) string {
	switch role {
	case config.Backup:
		return "backup"
	case config.Monitor:
		return "monitor"
	default:
		return "primary"
	}
}

// customLabels returns, per broker role present in this deployment, the user labels
// that are safe to apply: malformed entries and Kubernetes-managed prefixes are
// dropped with a warning so they neither prompt nor reach the argv.
func (c *Cluster) customLabels() map[config.Role][]labelKV {
	out := map[config.Role][]labelKV{}
	for _, role := range HARoles(c.Cfg) {
		for _, entry := range rolePlacementLabels(c.Cfg, role) {
			key, val, ok := splitLabel(entry)
			if !ok {
				c.progress().Warn("skipping malformed node label %q for %s", entry, roleName(role))
				continue
			}
			if isBuiltinLabel(key) {
				continue // managed by Kubernetes; not user-applied
			}
			if !validLabelToken(key) || !validLabelToken(val) {
				c.progress().Warn("skipping node label with unsupported characters %q for %s", entry, roleName(role))
				continue
			}
			out[role] = append(out[role], labelKV{key, val})
		}
	}
	return out
}

// nodeNames lists cluster node names via a name-only custom-columns query.
func (c *Cluster) nodeNames(ctx context.Context) ([]string, error) {
	raw, err := c.output(ctx, "get", "nodes", "-o", "custom-columns=NAME:.metadata.name", "--no-headers")
	if err != nil {
		return nil, fmt.Errorf("listing cluster nodes: %w", err)
	}
	var names []string
	for _, line := range strings.Split(string(raw), "\n") {
		if n := strings.TrimSpace(line); n != "" {
			names = append(names, n)
		}
	}
	return names, nil
}

// LabelNodes interactively applies the configured custom node labels, porting 013.
// It early-exits when nothing is configured (so `up` can call it unconditionally),
// prechecks the RBAC to update nodes before prompting, then for each role prompts
// the operator to pick a node from the cluster's node list and runs
// `kubectl label node <node> <key>=<val> --overwrite`. A single label failure is
// reported and skipped, not fatal, matching the bash loop.
func (c *Cluster) LabelNodes(ctx context.Context) error {
	custom := c.customLabels()
	if len(custom) == 0 {
		fmt.Fprintln(c.out(), "No custom node labels configured; nothing to label.")
		return nil
	}
	// This check predates Preflight and was its model; it now shares that one
	// implementation, which tells "you are not allowed" apart from "nobody answered".
	if err := c.Preflight(ctx, "update", "nodes"); err != nil {
		return err
	}
	nodes, err := c.nodeNames(ctx)
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		return fmt.Errorf("no cluster nodes found to label")
	}

	reader := bufio.NewReader(c.in())
	for _, role := range HARoles(c.Cfg) {
		labels := custom[role]
		if len(labels) == 0 {
			continue
		}
		node, err := c.promptNode(reader, role, nodes)
		if err != nil {
			return err
		}
		for _, kv := range labels {
			arg := kv.key + "=" + kv.val
			if err := c.kubectl(ctx, "label", "node", node, arg, "--overwrite"); err != nil {
				c.report().Fail("failed to apply %s to %s: %v", arg, node, err)
				continue
			}
			c.report().OK("labelled %s with %s", node, arg)
		}
	}
	return nil
}

// promptNode asks the operator to choose a node for a role from nodes, re-prompting
// on invalid input. EOF with no valid selection is a hard error rather than a
// silent default, so a mis-piped `up` cannot label the wrong node.
func (c *Cluster) promptNode(r *bufio.Reader, role config.Role, nodes []string) (string, error) {
	w := c.errOut()
	for {
		fmt.Fprintf(w, "Select the node for the %s broker:\n", roleName(role))
		for i, n := range nodes {
			fmt.Fprintf(w, "  %d) %s\n", i+1, n)
		}
		fmt.Fprint(w, "> ")

		line, rerr := r.ReadString('\n')
		choice, cerr := strconv.Atoi(strings.TrimSpace(line))
		if cerr == nil && choice >= 1 && choice <= len(nodes) {
			return nodes[choice-1], nil
		}
		if rerr != nil {
			return "", fmt.Errorf("no valid node selection for the %s role", roleName(role))
		}
		fmt.Fprintln(w, "Invalid selection; enter the number of a listed node.")
	}
}

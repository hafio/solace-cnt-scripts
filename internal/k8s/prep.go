package k8s

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"solace/internal/config"
	"solace/internal/render"
)

// out returns the report sink, defaulting to stdout when unset.
func (c *Cluster) out() io.Writer {
	if c.Out != nil {
		return c.Out
	}
	return os.Stdout
}

// This package no longer prompts. It once carried an errOut/in pair for the
// interactive node-labelling step; that step is gone (see the note on
// kubernetes.placement in the config schema -- worker nodes are never labelled by this
// tool, the labels exist only as the rendered CR's nodeSelector/affinity input), and
// the one remaining question -- the operator downgrade -- is asked through the
// Confirm func seam rather than a writer. Every other confirmation lives in
// internal/cli, which owns the terminal.

// namespaceManifest is a minimal core/v1 Namespace, the equivalent of the bash
// `create ns --dry-run=client -o yaml | apply -f -` (011:15). It is applied only to a
// namespace that is not there: applying it over one that exists is not the no-op it
// looks like -- kubectl PATCHes a namespace that lacks its last-applied annotation, and
// that needs `patch` on namespaces, which the built-in admin and edit roles do not grant.
func namespaceManifest(ns string) []byte {
	return []byte("apiVersion: v1\nkind: Namespace\nmetadata:\n  name: " + ns + "\n")
}

// namespaceState is what a deploy knows about the namespace it is about to write into.
type namespaceState int

const (
	namespaceUnknown namespaceState = iota // the read failed: no answer either way
	namespaceExists
	namespaceAbsent
)

// readNamespace asks whether ns exists, BY NAME. That is a question a namespace-scoped
// identity can answer: a RoleBinding in ns whose role reads namespaces -- the built-in
// view, edit and admin all do -- authorizes `get namespace <ns>`, because a request about
// one namespace is evaluated in that namespace. `create namespaces` never is; only a
// cluster-wide grant covers it, so asking for it when the namespace was already there
// refused exactly the identity this tool most often runs as.
//
// Only the exact `namespace/<ns>` line counts as existing and only an empty answer as
// absent; a failed read, or anything else on stdout, is unknown and leaves the decision
// to the create probe. Under the Echo runner the read is echoed and answers nothing, so a
// preview takes the first-deploy path.
func (c *Cluster) readNamespace(ctx context.Context, ns string) namespaceState {
	out, err := c.kubectlOutput(ctx, "get", "namespace", ns, "--ignore-not-found", "-o", "name")
	if err != nil {
		return namespaceUnknown
	}
	switch strings.TrimSpace(string(out)) {
	case "namespace/" + ns:
		return namespaceExists
	case "":
		return namespaceAbsent
	}
	return namespaceUnknown
}

// missingNamespace words a failed create probe for a namespace known to be absent. The
// namespace has to exist before anything can be deployed into it, so a refusal also says
// who can create it, not only which grant was missing.
func missingNamespace(ns string, err error) error {
	if !errors.Is(err, errNotAllowed) {
		return err
	}
	return fmt.Errorf("namespace %q does not exist, and this identity may not create it (%w).\n"+
		"  Or have a cluster admin create it: kubectl create namespace %s", ns, err, ns)
}

// CreateNamespace makes sure the broker namespace exists before the first namespaced
// write (011), asking for no more permission than that takes:
//
//   - it exists: nothing to create and nothing to ask, which is what lets an identity
//     confined to the namespace deploy into it.
//   - it is absent: creating it is genuinely needed, so a refusal stops the deploy before
//     anything is written and says who can fix it.
//   - its existence could not be read: the create probe decides. Permitted, the manifest
//     is applied. Refused, the deploy goes on with a warning: the namespace most likely
//     exists, and if it does not, the namespaced permission checks of the steps that
//     follow stop the deploy before their first write.
func (c *Cluster) CreateNamespace(ctx context.Context) error {
	ns := c.ns()
	switch c.readNamespace(ctx, ns) {
	case namespaceExists:
		c.logf("namespace %s exists", ns)
		return nil
	case namespaceAbsent:
		if err := c.Preflight(ctx, "create", "namespaces"); err != nil {
			return missingNamespace(ns, err)
		}
	default:
		if err := c.Preflight(ctx, "create", "namespaces"); err != nil {
			if !errors.Is(err, errNotAllowed) {
				return err
			}
			c.progress().Warn("could not confirm namespace %s exists (that read needs get on namespaces in %s), "+
				"and this identity may not create namespaces; continuing -- the next steps' permission checks "+
				"stop the deploy before it writes anything if the namespace is missing", ns, ns)
			return nil
		}
	}
	c.logf("creating namespace %s", ns)
	return c.apply(ctx, namespaceManifest(ns))
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
// Emptiness is not this method's decision -- it is already made by the time this
// runs. Cluster.NamespaceContents (namespace.go) enumerates the namespace
// ownership-BLIND over occupancyKinds: ANY object other than alwaysPresent (the
// Kubernetes-generated kube-root-ca.crt ConfigMap and default ServiceAccount)
// blocks removal. removeNamespaceIfEmpty (internal/cli/ops_k8s.go) is the only
// caller of DeleteNamespace, and it runs NamespaceContents first, reaching here
// only when that reported the namespace empty. This method therefore does not
// re-enumerate or re-classify what the namespace holds: a second guard here once
// did, listing a different twelve-kind set and judging each object OURS/FOREIGN
// (isOurs) rather than "anything remains" -- which is a weaker rule than
// NamespaceContents' own (an object this guard judged "ours" would pass through
// where the design says nothing does), and which cost two lists to keep aligned
// for no benefit. There is one classifier now, and it runs once, upstream.
//
// protectedNamespaces is still checked first here and without even asking the
// cluster: there is no interpretation of "delete kube-system" that this tool
// should ever carry out, and it is not this deployment's to leave to
// NamespaceContents' judgement. Neither guard can be silenced by --no-prompt:
// that flag only decides whether the CLI's own confirmDelete asks a question
// before calling here, so this method is exactly what a fully unattended
// removal reaches, with nothing left upstream to skip its checks.
func (c *Cluster) DeleteNamespace(ctx context.Context) error {
	if protectedNamespaces[c.ns()] {
		return fmt.Errorf("refusing to delete namespace %q: this tool will never delete it; "+
			"use \"remove broker\" / \"remove secrets\" to remove just this deployment's own objects", c.ns())
	}
	if err := c.Preflight(ctx, "delete", "namespaces"); err != nil {
		return err
	}
	c.logf("deleting namespace %s", c.ns())
	return c.kubectl(ctx, "delete", "namespace", c.ns(), "--ignore-not-found")
}

// secretPreflight fails loud before any manifest is built when the certificate inputs
// this tool is about to read are unusable, porting the guard of 012:19-24 so the operator
// does not later fail to mount a half-built secret. The admin secret's own guards live in
// AdminSecret, which runs only when this tool builds that Secret.
//
// It is keyed on ManagesTLSSecret, not on the Secret NAME. A named Secret with no files
// behind it is the bring-your-own case -- the operator created it, the CR references it,
// and there is nothing here to preflight.
func (c *Cluster) secretPreflight() error {
	if !c.Cfg.ManagesTLSSecret() {
		return nil
	}
	if c.Cfg.TLS.Cert == "" || c.Cfg.TLS.CertKey == "" {
		return fmt.Errorf("tls.cert and tls.certKey must be set together to build the TLS Secret (got cert=%q key=%q)",
			c.Cfg.TLS.Cert, c.Cfg.TLS.CertKey)
	}
	for _, f := range []string{c.Cfg.TLS.Cert, c.Cfg.TLS.CertKey} {
		if _, err := os.Stat(f); err != nil {
			return fmt.Errorf("tls certificate input %q is not readable: %w", f, err)
		}
	}
	return nil
}

// ownedSecretNames lists, in apply order, every Secret this env file makes this tool
// build -- and so the ones `broker remove` deletes. Each entry is keyed on the MATERIAL,
// never on a name: a Secret the env file only names is someone else's, and one it
// neither names nor supplies (the admin Secret with no semp.adminPass) is the Solace
// operator's to generate. GenSecrets and DeleteSecrets both read this one list, so what
// is built and what is removed cannot drift apart.
func ownedSecretNames(cfg *config.Config) []string {
	var names []string
	add := func(owned bool, name string) {
		if owned && name != "" {
			names = append(names, name)
		}
	}
	add(cfg.ManagesAdminSecret(), cfg.AdminSecretName())
	add(cfg.ManagesTLSSecret(), cfg.TLSServerSecretName())
	add(cfg.ManagesImagePullSecret(), cfg.ImagePullSecretName())
	add(len(cfg.SEMP.AdditionalUsers) > 0, cfg.AdditionalUsersSecretName())
	return names
}

// GenSecrets builds every Secret this tool owns (ownedSecretNames: the credentials Secret
// under config.Config.ManagesAdminSecret, TLS under ManagesTLSSecret, the image-pull
// Secret under ManagesImagePullSecret, the additional users' when there are any) and
// joins them into one multi-doc manifest -- porting 012's secret set. It is the
// rendering behind both CreateSecrets and `broker generate`, so what a user
// reviews is exactly what gets applied. The manifests carry the base64-encoded
// secret values, so the output is as sensitive as the env file it came from.
//
// nil when there is nothing to build -- no admin password, no TLS files, no registry
// credentials, no additional users -- so neither caller has an empty document to apply.
func GenSecrets(cfg *config.Config) ([]byte, error) {
	docs := make([][]byte, 0, 4)

	// Keyed on the MATERIAL, as every entry below is: with no semp.adminPass the CR
	// references a Secret someone else made, or none, and there is nothing here to build.
	if cfg.ManagesAdminSecret() {
		admin, err := AdminSecret(cfg)
		if err != nil {
			return nil, err
		}
		docs = append(docs, admin)
	}

	// Keyed on the MATERIAL, not on the name. A named Secret that this env file does not
	// supply files for is one the operator created themselves -- the CR still references
	// it, and reading a certificate off disk to rebuild it would be both wrong and, as the
	// path would have to be invented, impossible.
	if cfg.ManagesTLSSecret() {
		tls, err := TLSSecret(cfg)
		if err != nil {
			return nil, err
		}
		docs = append(docs, tls)
	}

	// Keyed on the MATERIAL, not on the name, the identical reason ManagesTLSSecret is
	// just above: a named Secret this env file supplies no credentials for is one the
	// operator created themselves, the CR still references it, and there is nothing here
	// to build from.
	if cfg.ManagesImagePullSecret() {
		pull, err := DockerRegistrySecret(cfg)
		if err != nil {
			return nil, err
		}
		docs = append(docs, pull)
	}
	// nil when no additional users are configured, in which case the CR omits
	// extraEnvVarsSecret too and there is nothing to apply.
	users, err := AdditionalUsersSecret(cfg)
	if err != nil {
		return nil, err
	}
	if users != nil {
		docs = append(docs, users)
	}
	if len(docs) == 0 {
		return nil, nil
	}
	return joinManifests(docs), nil
}

// CreateSecrets applies the GenSecrets manifest on stdin, keeping every secret
// value off the argv and out of an echoed command. The whole manifest is
// built before the first apply, so a builder error aborts cleanly without leaving
// a partially-applied secret set.
//
// With nothing to build it says so and returns, before any probe: `kubectl apply` of an
// empty stream fails ("no objects passed to apply"), and a deploy that creates no Secret
// has no business demanding permission to create one (the deploy.go:112 rule).
func (c *Cluster) CreateSecrets(ctx context.Context) error {
	if len(ownedSecretNames(c.Cfg)) == 0 {
		c.logf("no Secrets to create in %s -- the broker CR references only Secrets that already exist "+
			"or that the operator generates", c.ns())
		return nil
	}
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

// DeleteSecrets removes the secrets CreateSecrets created (112): exactly
// ownedSecretNames, under their resolved names. A Secret this env file only names is not
// ours to remove -- it may be shared with another workload, and nothing about naming it
// in this file made it ours -- and the operator's generated ones go with the broker CR
// that owns them. All use --ignore-not-found so a partial or repeat teardown is not an
// error, and with nothing to delete there is no permission to probe either.
func (c *Cluster) DeleteSecrets(ctx context.Context) error {
	names := ownedSecretNames(c.Cfg)
	if len(names) == 0 {
		return nil
	}
	if err := c.Preflight(ctx, "delete", "secrets"); err != nil {
		return err
	}
	for _, name := range names {
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
// an echoed command.
//
// The Secret is what a pod mounts when it STARTS; a running broker does not re-read
// it. So this is the persistence half of `broker configure server-certs`, and the
// CLI load that follows it there (broker.Ops.ServerCert) is what swaps the
// certificate into the running broker. No pod is restarted here.
//
// The probe asks for what `kubectl apply` actually sends over an existing Secret -- a GET,
// then a PATCH -- rather than `update`, which apply never uses: asking for `update` refused
// a Role granting exactly get and patch, and passed one granting update alone that then
// failed inside the apply.
func (c *Cluster) UpdateServerCertSecret(ctx context.Context) error {
	name := c.Cfg.TLSServerSecretName()
	if name == "" {
		return fmt.Errorf("this env file has no TLS Secret to update: set tls.cert and tls.certKey " +
			"(the Secret is then built as kubernetes.tlsServerSecret, or <kubernetes.name>-tls when that is unset)")
	}
	if !c.Cfg.ManagesTLSSecret() {
		return fmt.Errorf("kubernetes.tlsServerSecret %q names a Secret this env file does not supply the files for, "+
			"so there is nothing here to rebuild it from.\n"+
			"  Rotate it where it is managed (kubectl, cert-manager), or set tls.cert and tls.certKey to hand this tool the pair",
			name)
	}
	if err := c.PreflightAll(ctx, probe{verb: "get", resource: "secrets"}, probe{verb: "patch", resource: "secrets"}); err != nil {
		return err
	}
	manifest, err := TLSSecret(c.Cfg)
	if err != nil {
		return err
	}
	c.logf("updating server-certificate secret %s", name)
	return c.apply(ctx, manifest)
}

// ServerCertChanged reports whether the TLS Secret this env file builds is already in
// the cluster holding a different certificate from the one on disk, read from its
// render.CertDigestLabel annotation -- the rule the container platforms apply to their
// labels. It is asked BEFORE CreateSecrets rewrites the annotation.
//
// false when there is nothing to compare: no Secret this env file builds, or none in
// the cluster yet (a first deploy, or TLS newly added -- where the CR's own tls block
// changes and the operator rolls the pods itself). A Secret with a different or
// missing annotation (one an earlier build applied) is changed, and so is one that
// cannot be read, which costs at most a consent-gated restart.
func (c *Cluster) ServerCertChanged(ctx context.Context) (bool, error) {
	if !c.Cfg.ManagesTLSSecret() {
		return false, nil
	}
	key, crt, err := readTLSMaterial(c.Cfg)
	if err != nil {
		return false, err
	}
	name := c.Cfg.TLSServerSecretName()
	// jsonpath rather than -o json, so the Secret's data -- the private key -- is
	// never read back into this process. The name line says whether it exists.
	path := `{.metadata.name}{"\n"}{.metadata.annotations.` + strings.ReplaceAll(render.CertDigestLabel, ".", `\.`) + `}`
	out, err := c.kubectlOutput(ctx, "get", "secret", name, "-n", c.ns(), "--ignore-not-found", "-o", "jsonpath="+path)
	if err != nil {
		c.progress().Warn("could not read TLS Secret %s to compare certificates (%v); treating the certificate as changed.",
			name, err)
		return true, nil
	}
	found, digest, _ := strings.Cut(strings.TrimSpace(string(out)), "\n")
	if found == "" {
		return false, nil
	}
	return strings.TrimSpace(digest) != tlsDigest(key, crt), nil
}

// BrokerPodsExist reports whether this broker's primary pod exists, which is what
// "a broker is already running" means to a deploy deciding whether a renewed
// certificate needs a restart. A runner that answers nothing answers false, so a
// preview never offers one.
//
// A read that fails answers false too -- no restart is offered for a pod nobody could see
// -- but not silently: the renewed certificate is then in the Secret while a running broker
// keeps serving the old one, so the warning says so and names the way to finish.
func (c *Cluster) BrokerPodsExist(ctx context.Context) bool {
	pod := podName(c.Cfg, config.Primary)
	out, err := c.kubectlOutput(ctx, "get", "pod", pod, "-n", c.ns(), "--ignore-not-found", "-o", "name")
	if err != nil {
		c.progress().Warn("could not read pod %s (that needs get on pods in %s), so no restart is offered for "+
			"the renewed server certificate; if the broker is running it serves the previous one until its pods "+
			"restart -- run `broker restart`, or hot-swap it with `broker configure server-certs`", pod, c.ns())
		return false
	}
	return strings.TrimSpace(string(out)) != ""
}

// GenBroker renders everything `broker deploy` applies, in the order it applies it:
// the Namespace, then the Secrets, then the broker CR. That equivalence is the point of
// the command -- an artifact you can read is only worth reading if it is the artifact
// that would be applied -- and it is what makes `broker generate | kubectl apply -f -`
// work against an empty cluster. The Secrets and the CR are namespaced, so the Namespace
// has to lead for the same reason the Secrets precede the CR that references them.
//
// It mirrors GenOperator, whose stream has carried its own Namespace document from the
// start; this half was the one that did not.
func GenBroker(cfg *config.Config) ([]byte, error) {
	secrets, err := GenSecrets(cfg)
	if err != nil {
		return nil, err
	}
	return joinManifests([][]byte{
		namespaceManifest(cfg.K8s.Namespace),
		secrets,
		render.BrokerCR(cfg),
	}), nil
}

// joinManifests concatenates rendered YAML documents with a `---` separator so
// they apply as one multi-doc stream. A blank document is dropped rather than joined:
// GenBroker passes GenSecrets' nil when there is no Secret to build, and the stream
// must then be the Namespace and the CR with nothing between them.
func joinManifests(docs [][]byte) []byte {
	parts := make([]string, 0, len(docs))
	for _, d := range docs {
		if part := strings.TrimRight(string(d), "\n"); strings.TrimSpace(part) != "" {
			parts = append(parts, part)
		}
	}
	return []byte(strings.Join(parts, "\n---\n") + "\n")
}

// --- node labelling (013): removed ----------------------------------------
//
// Node labelling used to live here: a `kubectl label node` loop fed by an interactive
// picker (`prepare labels`, porting bash 013), the Kubernetes-managed prefix skip list,
// the key=value / key: value parser, and the charset check that kept operator-supplied
// strings out of that argv.
//
// All of it is deliberately gone, and nothing replaced it. This tool never labels
// cluster worker nodes: the kubernetes.placement.labels entries are SELECTORS, consumed
// by internal/render as the rendered CR's nodeSelector and affinity terms, and whoever
// owns the cluster owns which node carries which label. internal/render parses them with
// its own splitPair, so nothing here was left with a caller.
//
// The picker was also the ONE prerequisite that could not be folded into an idempotent
// `broker deploy`: it needed a human at a terminal, and it recorded its choice nowhere,
// so a re-run could label a different node than the first run did.

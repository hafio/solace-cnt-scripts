package k8s

import (
	"context"
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
// AdminSecret.
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

// GenSecrets builds every applicable secret (admin always; TLS when
// kubernetes.tlsServerSecret is set; the image-pull secret when registry
// credentials are present, config.Config.ManagesImagePullSecret) and
// joins them into one multi-doc manifest -- porting 012's secret set. It is the
// rendering behind both CreateSecrets and `broker generate`, so what a user
// reviews is exactly what gets applied. The manifests carry the base64-encoded
// secret values, so the output is as sensitive as the env file it came from.
func GenSecrets(cfg *config.Config) ([]byte, error) {
	docs := make([][]byte, 0, 3)

	admin, err := AdminSecret(cfg)
	if err != nil {
		return nil, err
	}
	docs = append(docs, admin)

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
	return joinManifests(docs), nil
}

// CreateSecrets applies the GenSecrets manifest on stdin, keeping every secret
// value off the argv and out of an echoed command. The whole manifest is
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
	// Only the TLS Secret this tool built. One the operator created and merely pointed the
	// env file at is not ours to remove -- it may be shared with another workload, and
	// nothing about naming it in this file made it ours.
	if c.Cfg.ManagesTLSSecret() && c.Cfg.K8s.TLSServerSecret != "" {
		names = append(names, c.Cfg.K8s.TLSServerSecret)
	}
	// Only the image-pull Secret this tool built, the same rule as the TLS Secret just
	// above: a named Secret with no credentials behind it is the operator's own, not ours
	// to remove.
	if c.Cfg.ManagesImagePullSecret() {
		names = append(names, c.Cfg.ImagePullSecretName())
	}
	if len(c.Cfg.SEMP.AdditionalUsers) > 0 {
		names = append(names, c.Cfg.AdditionalUsersSecretName())
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
// an echoed command. The broker re-reads the secret; no pod restart here.
func (c *Cluster) UpdateServerCertSecret(ctx context.Context) error {
	if c.Cfg.K8s.TLSServerSecret == "" {
		return fmt.Errorf("kubernetes.tlsServerSecret must be set to update the server-certificate secret")
	}
	if !c.Cfg.ManagesTLSSecret() {
		return fmt.Errorf("kubernetes.tlsServerSecret %q names a Secret this env file does not supply the files for, "+
			"so there is nothing here to rebuild it from.\n"+
			"  Rotate it where it is managed (kubectl, cert-manager), or set tls.cert and tls.certKey to hand this tool the pair",
			c.Cfg.K8s.TLSServerSecret)
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
// they apply as one multi-doc stream.
func joinManifests(docs [][]byte) []byte {
	parts := make([]string, len(docs))
	for i, d := range docs {
		parts[i] = strings.TrimRight(string(d), "\n")
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

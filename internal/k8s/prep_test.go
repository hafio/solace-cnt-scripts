package k8s

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"solace/internal/config"
)

// adminCfg is haCfg with the minimum admin fields AdminSecret requires, so
// CreateSecrets/DeleteSecrets tests are hermetic (no dependency on env/sample.yaml).
func adminCfg() *config.Config {
	c := haCfg()
	c.SEMP.AdminPass = "pw"
	c.K8s.AdminSecret = "solace-admin-secret"
	return c
}

func TestCreateNamespace(t *testing.T) {
	rr := &recRunner{}
	c := newCluster(rr)
	if err := c.CreateNamespace(context.Background()); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	got := rr.last()
	if got.method != "RunInput" || got.name != "kubectl" || !eqArgs(got.args, []string{"apply", "-f", "-"}) {
		t.Fatalf("CreateNamespace argv = %+v, want RunInput kubectl [apply -f -]", got)
	}
	if !strings.Contains(got.stdin, "kind: Namespace") || !strings.Contains(got.stdin, "name: solace") {
		t.Errorf("CreateNamespace manifest on stdin =\n%s", got.stdin)
	}
}

// TestCreateNamespaceApplyFails proves a failing apply (RBAC denial, etc.) surfaces
// instead of being silently swallowed. This branch was untestable before recRunner
// grew runInputErr: RunInput unconditionally returned nil, so no RunInput-backed
// call (apply/deleteStdin) anywhere in the suite could ever be made to fail.
func TestCreateNamespaceApplyFails(t *testing.T) {
	rr := &recRunner{runInputErr: errFake}
	c := newCluster(rr)
	if err := c.CreateNamespace(context.Background()); err == nil {
		t.Error("CreateNamespace should fail when the apply fails")
	}
}

// TestDeleteNamespace exercises the ordinary path: the permission probe passes
// and the delete is issued directly, with no enumeration of the namespace's
// contents here -- that is NamespaceContents' job, upstream of this call.
func TestDeleteNamespace(t *testing.T) {
	rr := &recRunner{}
	c := newCluster(rr)
	if err := c.DeleteNamespace(context.Background()); err != nil {
		t.Fatalf("DeleteNamespace: %v", err)
	}
	calls := rr.afterPreflight(t, "delete", "namespaces")
	if len(calls) != 1 {
		t.Fatalf("DeleteNamespace made %d calls after the probe, want 1 delete", len(calls))
	}
	got := calls[0]
	want := []string{"delete", "namespace", "solace", "--ignore-not-found"}
	if got.method != "Run" || got.name != "kubectl" || !eqArgs(got.args, want) {
		t.Errorf("DeleteNamespace argv = %+v, want Run kubectl %v", got, want)
	}
}

// TestDeleteNamespaceProtected pins protectedNamespaces' floor: none of the four
// Kubernetes system namespaces is ever deleted, and -- unlike every other guard
// in DeleteNamespace -- the refusal happens before the cluster is asked
// anything at all, RBAC probe included.
func TestDeleteNamespaceProtected(t *testing.T) {
	for _, ns := range []string{"default", "kube-system", "kube-public", "kube-node-lease"} {
		t.Run(ns, func(t *testing.T) {
			cfg := haCfg()
			cfg.K8s.Namespace = ns
			rr := &recRunner{}
			c := NewCluster(rr, cfg, nil, nil)
			err := c.DeleteNamespace(context.Background())
			if err == nil {
				t.Fatalf("DeleteNamespace(%s) should be refused", ns)
			}
			if !strings.Contains(err.Error(), ns) {
				t.Errorf("error = %v, want it to name the namespace", err)
			}
			if len(rr.calls) != 0 {
				t.Errorf("a protected namespace must not reach the cluster at all; got %d calls: %+v", len(rr.calls), rr.calls)
			}
		})
	}
}

// TestDeleteNamespaceStopsOnPreflightFailure proves a refused permission stops
// DeleteNamespace before any `kubectl delete` is issued -- the same shape as
// TestCreateSecretsStopsOnPreflightFailure, but for the namespace teardown.
func TestDeleteNamespaceStopsOnPreflightFailure(t *testing.T) {
	rr := &recRunner{canI: "no"}
	c := newCluster(rr)
	err := c.DeleteNamespace(context.Background())
	if err == nil {
		t.Fatal("DeleteNamespace must fail when the permission probe answers no")
	}
	if !strings.Contains(err.Error(), "not allowed to delete namespaces") {
		t.Errorf("error = %v, want it to name the refused permission", err)
	}
	for _, call := range rr.calls {
		if len(call.args) > 0 && call.args[0] == "delete" {
			t.Error("no delete namespace call must be issued after an RBAC denial")
		}
	}
}

// TestCreateSecretsAdminOnly: with no TLS and no pull secret, only the admin secret
// is applied, as a single (unseparated) manifest.
func TestCreateSecretsAdminOnly(t *testing.T) {
	rr := &recRunner{}
	c := NewCluster(rr, adminCfg(), nil, nil)
	if err := c.CreateSecrets(context.Background()); err != nil {
		t.Fatalf("CreateSecrets: %v", err)
	}
	if calls := rr.afterPreflight(t, "create", "secrets"); len(calls) != 1 {
		t.Fatalf("CreateSecrets made %d calls after the probe, want 1 apply", len(calls))
	}
	got := rr.last()
	if got.method != "RunInput" || !eqArgs(got.args, []string{"apply", "-f", "-"}) {
		t.Fatalf("CreateSecrets argv = %+v, want RunInput kubectl [apply -f -]", got)
	}
	if !strings.Contains(got.stdin, "name: solace-admin-secret") || !strings.Contains(got.stdin, "type: Opaque") {
		t.Errorf("admin secret missing from stdin:\n%s", got.stdin)
	}
	if strings.Contains(got.stdin, "kubernetes.io/tls") || strings.Contains(got.stdin, ".dockerconfigjson") {
		t.Error("admin-only CreateSecrets should not emit a TLS or pull secret")
	}
	if strings.Contains(got.stdin, "---") {
		t.Error("a single secret must not carry a document separator")
	}
}

// TestCreateSecretsAllThree: admin + TLS + pull secret join into one multi-doc
// manifest applied on stdin, and no secret value ever reaches the argv (§3).
func TestCreateSecretsAllThree(t *testing.T) {
	dir := t.TempDir()
	crt := filepath.Join(dir, "tls.crt")
	key := filepath.Join(dir, "tls.key")
	writeFile(t, crt, "CERT\n")
	writeFile(t, key, "KEY\n")

	cfg := adminCfg()
	cfg.K8s.TLSServerSecret = "solace-tls-secret"
	cfg.TLS.Cert = crt
	cfg.TLS.CertKey = key
	cfg.K8s.ImagePullSecret = "solace-image-pull"
	cfg.Image.User = "u"
	cfg.Image.Pass = "SECRET-REG-PASS"
	cfg.Image.Registry = "registry.example.com"

	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.CreateSecrets(context.Background()); err != nil {
		t.Fatalf("CreateSecrets: %v", err)
	}
	if calls := rr.afterPreflight(t, "create", "secrets"); len(calls) != 1 {
		t.Fatalf("CreateSecrets made %d calls after the probe, want 1 apply", len(calls))
	}
	got := rr.last()
	if got.method != "RunInput" || !eqArgs(got.args, []string{"apply", "-f", "-"}) {
		t.Fatalf("CreateSecrets argv = %+v", got)
	}
	if strings.Count(got.stdin, "---") != 2 {
		t.Errorf("expected 3 joined docs (2 separators):\n%s", got.stdin)
	}
	for _, want := range []string{"name: solace-admin-secret", "name: solace-tls-secret", "name: solace-image-pull"} {
		if !strings.Contains(got.stdin, want) {
			t.Errorf("stdin missing %q", want)
		}
	}
	if strings.Contains(strings.Join(got.args, " "), "SECRET-REG-PASS") {
		t.Error("registry password leaked into the argv")
	}
	if strings.Contains(got.stdin, "SECRET-REG-PASS") {
		t.Error("registry password appears in plaintext on stdin (must be base64 in the secret data)")
	}
}

// TestCreateSecretsPreflight fails loud when this tool is about to build the TLS secret
// and cannot -- before any apply runs (012:19-24).
//
// "Cannot" no longer includes "no cert/key configured at all": that is the bring-your-own
// case, covered by the subtest below, where the Secret already exists and there is nothing
// to preflight. What is still refused is a half-supplied pair and a file that will not read.
func TestCreateSecretsPreflight(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(c *config.Config)
	}{
		{"key unset", func(c *config.Config) { c.TLS.Cert = "certs/tls.crt"; c.TLS.CertKey = "" }},
		{"cert unset", func(c *config.Config) { c.TLS.Cert = ""; c.TLS.CertKey = "certs/tls.key" }},
		{"cert file missing", func(c *config.Config) {
			c.TLS.Cert = filepath.Join(t.TempDir(), "nope.crt")
			c.TLS.CertKey = filepath.Join(t.TempDir(), "nope.key")
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := adminCfg()
			cfg.K8s.TLSServerSecret = "solace-tls-secret"
			tc.mutate(cfg)
			rr := &recRunner{}
			c := NewCluster(rr, cfg, nil, nil)
			if err := c.CreateSecrets(context.Background()); err == nil {
				t.Error("CreateSecrets should fail the preflight")
			}
			if len(rr.calls) != 0 {
				t.Errorf("no apply must run when the preflight fails; got %d calls", len(rr.calls))
			}
		})
	}

	// Naming a Secret with no files behind it is a deployment, not a gap: the operator
	// created it and the CR references it. This used to be the first case above.
	t.Run("a named secret with no files is not preflighted", func(t *testing.T) {
		cfg := adminCfg()
		cfg.K8s.TLSServerSecret = "byo-tls-secret"
		rr := &recRunner{}
		c := NewCluster(rr, cfg, nil, nil)
		if err := c.CreateSecrets(context.Background()); err != nil {
			t.Fatalf("an existing Secret needs no certificate on this host: %v", err)
		}
		if strings.Contains(rr.last().stdin, "kubernetes.io/tls") {
			t.Errorf("no TLS Secret may be applied when none was built:\n%s", rr.last().stdin)
		}
	})
}

// TestCreateSecretsFailsWithoutAdminFields proves CreateSecrets can pass
// secretPreflight (which only validates the TLS inputs) and still fail inside
// GenSecrets when the admin fields are unset -- a real, reachable misconfiguration
// that secretPreflight's TLS-only guard does not catch.
func TestCreateSecretsFailsWithoutAdminFields(t *testing.T) {
	rr := &recRunner{}
	c := NewCluster(rr, haCfg(), nil, nil) // no Admin.Pass, no K8s.AdminSecret
	if err := c.CreateSecrets(context.Background()); err == nil {
		t.Error("CreateSecrets should fail when the admin secret cannot be built")
	}
	// The permission probe runs (it precedes GenSecrets, so key material is never
	// read for a cluster that would refuse it); no apply may follow.
	if calls := rr.afterPreflight(t, "create", "secrets"); len(calls) != 0 {
		t.Errorf("CreateSecrets should abort before any apply; the probe passed, GenSecrets should stop it; got %d calls after it", len(calls))
	}
}

// TestCreateSecretsStopsOnPreflightFailure: a refused permission stops CreateSecrets
// before GenSecrets reads the TLS private key off disk. Loading key material into
// this process for a cluster that will not accept it is work worth not doing.
func TestCreateSecretsStopsOnPreflightFailure(t *testing.T) {
	rr := &recRunner{canI: "no"}
	c := NewCluster(rr, adminCfg(), nil, nil)
	err := c.CreateSecrets(context.Background())
	if err == nil {
		t.Fatal("CreateSecrets must fail when the permission probe answers no")
	}
	if !strings.Contains(err.Error(), "not allowed to create secrets") {
		t.Errorf("error = %v, want it to name the refused permission", err)
	}
	if len(rr.calls) != 1 {
		t.Errorf("%d calls made after a failed probe, want only the probe itself: %+v", len(rr.calls), rr.calls)
	}
}

// TestGenSecretsTLSError covers GenSecrets' own guard on the render-only path
// (`broker generate`), which calls GenSecrets directly and bypasses
// Cluster.secretPreflight entirely: a configured kubernetes.tlsServerSecret with unreadable
// cert files must fail the render rather than emit a broken manifest. Every other
// exercise of GenSecrets goes through CreateSecrets, which pre-empts this via
// preflight, so GenSecrets itself had zero direct tests.
func TestGenSecretsTLSError(t *testing.T) {
	cfg := adminCfg()
	cfg.K8s.TLSServerSecret = "solace-tls-secret"
	cfg.TLS.Cert = filepath.Join(t.TempDir(), "nope.crt")
	cfg.TLS.CertKey = filepath.Join(t.TempDir(), "nope.key")
	if _, err := GenSecrets(cfg); err == nil {
		t.Error("GenSecrets should fail when the TLS cert files are not readable")
	}
}

func TestDeleteSecrets(t *testing.T) {
	t.Run("all three", func(t *testing.T) {
		cfg := adminCfg()
		cfg.K8s.TLSServerSecret = "solace-tls-secret"
		// The cert/key pair is what makes the TLS Secret ours, and so what puts it in
		// the delete set -- see the bring-your-own subtest below.
		cfg.TLS.Cert, cfg.TLS.CertKey = "certs/tls.crt", "certs/tls.key"
		cfg.K8s.ImagePullSecret = "solace-image-pull"
		rr := &recRunner{}
		c := NewCluster(rr, cfg, nil, nil)
		if err := c.DeleteSecrets(context.Background()); err != nil {
			t.Fatalf("DeleteSecrets: %v", err)
		}
		wantNames := []string{"solace-admin-secret", "solace-tls-secret", "solace-image-pull"}
		calls := rr.afterPreflight(t, "delete", "secrets")
		if len(calls) != len(wantNames) {
			t.Fatalf("DeleteSecrets made %d calls after the probe, want %d", len(calls), len(wantNames))
		}
		for i, name := range wantNames {
			want := []string{"delete", "secret", name, "-n", "solace", "--ignore-not-found"}
			if got := calls[i]; got.method != "Run" || !eqArgs(got.args, want) {
				t.Errorf("delete[%d] = %+v, want Run kubectl %v", i, got, want)
			}
		}
	})
	t.Run("admin only", func(t *testing.T) {
		rr := &recRunner{}
		c := NewCluster(rr, adminCfg(), nil, nil)
		if err := c.DeleteSecrets(context.Background()); err != nil {
			t.Fatalf("DeleteSecrets: %v", err)
		}
		if calls := rr.afterPreflight(t, "delete", "secrets"); len(calls) != 1 {
			t.Fatalf("DeleteSecrets (admin only) made %d calls after the probe, want 1", len(calls))
		}
	})
	// A TLS Secret this env file only NAMES is not ours to remove. It may be shared with
	// another workload, and nothing about naming it here made it this deployment's --
	// deleting it would take TLS down for whoever does own it.
	t.Run("a TLS secret we did not create survives", func(t *testing.T) {
		cfg := adminCfg()
		cfg.K8s.TLSServerSecret = "byo-tls-secret"
		cfg.TLS.Cert, cfg.TLS.CertKey = "", ""
		rr := &recRunner{}
		c := NewCluster(rr, cfg, nil, nil)
		if err := c.DeleteSecrets(context.Background()); err != nil {
			t.Fatalf("DeleteSecrets: %v", err)
		}
		for _, call := range rr.afterPreflight(t, "delete", "secrets") {
			for _, a := range call.args {
				if a == "byo-tls-secret" {
					t.Fatalf("a Secret this tool did not build must not be deleted: %+v", call)
				}
			}
		}
	})
}

// TestDeleteSecretsStopsOnPreflightFailure proves a refused permission stops
// DeleteSecrets before any `kubectl delete secret` is issued -- the same shape as
// TestCreateSecretsStopsOnPreflightFailure, but for the secret teardown.
func TestDeleteSecretsStopsOnPreflightFailure(t *testing.T) {
	rr := &recRunner{canI: "no"}
	c := NewCluster(rr, adminCfg(), nil, nil)
	err := c.DeleteSecrets(context.Background())
	if err == nil {
		t.Fatal("DeleteSecrets must fail when the permission probe answers no")
	}
	if !strings.Contains(err.Error(), "not allowed to delete secrets") {
		t.Errorf("error = %v, want it to name the refused permission", err)
	}
	if len(rr.calls) != 1 {
		t.Errorf("%d calls made after a failed probe, want only the probe itself: %+v", len(rr.calls), rr.calls)
	}
}

// TestDeleteSecretsSkipsUnconfiguredAdminSecret proves names' unconditional first
// entry (c.Cfg.K8s.AdminSecret) is silently skipped when blank, rather than issuing
// `kubectl delete secret ""` -- the case of a partial/legacy deployment that never
// configured an admin secret but still calls teardown.
func TestDeleteSecretsSkipsUnconfiguredAdminSecret(t *testing.T) {
	cfg := haCfg() // no K8s.AdminSecret, no TLS/pull secret configured
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.DeleteSecrets(context.Background()); err != nil {
		t.Fatalf("DeleteSecrets: %v", err)
	}
	if calls := rr.afterPreflight(t, "delete", "secrets"); len(calls) != 0 {
		t.Errorf("DeleteSecrets should skip the blank admin-secret entry; got %d calls after the probe", len(calls))
	}
}

// TestDeleteSecretsStopsOnError proves a genuine delete failure (RBAC denial, beyond
// --ignore-not-found's usual no-op) stops the teardown loop and surfaces, instead of
// silently continuing to the remaining secrets. DeleteSecrets had no failure test.
func TestDeleteSecretsStopsOnError(t *testing.T) {
	cfg := adminCfg()
	cfg.K8s.TLSServerSecret = "x"
	// Supplied, so the TLS secret really is in the delete set and "stopped before it"
	// means something -- without the pair it would not be queued at all.
	cfg.TLS.Cert, cfg.TLS.CertKey = "certs/tls.crt", "certs/tls.key"
	rr := &recRunner{runErr: errFake}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.DeleteSecrets(context.Background()); err == nil {
		t.Error("DeleteSecrets should fail loud when a delete errors")
	}
	if calls := rr.afterPreflight(t, "delete", "secrets"); len(calls) != 1 {
		t.Errorf("DeleteSecrets should stop before the TLS secret; got %d calls after the probe", len(calls))
	}
}

func TestUpdateServerCertSecret(t *testing.T) {
	t.Run("applies TLS secret on stdin", func(t *testing.T) {
		dir := t.TempDir()
		crt := filepath.Join(dir, "tls.crt")
		key := filepath.Join(dir, "tls.key")
		writeFile(t, crt, "CERT\n")
		writeFile(t, key, "KEY\n")
		cfg := haCfg()
		cfg.K8s.TLSServerSecret = "solace-tls-secret"
		cfg.TLS.Cert = crt
		cfg.TLS.CertKey = key
		rr := &recRunner{}
		c := NewCluster(rr, cfg, nil, nil)
		if err := c.UpdateServerCertSecret(context.Background()); err != nil {
			t.Fatalf("UpdateServerCertSecret: %v", err)
		}
		got := rr.last()
		if got.method != "RunInput" || !eqArgs(got.args, []string{"apply", "-f", "-"}) {
			t.Fatalf("UpdateServerCertSecret argv = %+v", got)
		}
		if !strings.Contains(got.stdin, "kubernetes.io/tls") || !strings.Contains(got.stdin, "name: solace-tls-secret") {
			t.Errorf("stdin is not the TLS secret:\n%s", got.stdin)
		}
	})
	t.Run("errors without a secret name", func(t *testing.T) {
		rr := &recRunner{}
		c := NewCluster(rr, haCfg(), nil, nil) // no K8s.TLSServerSecret
		if err := c.UpdateServerCertSecret(context.Background()); err == nil {
			t.Error("UpdateServerCertSecret should fail when kubernetes.tlsServerSecret is unset")
		}
	})
	// A named Secret with no files behind it cannot be rebuilt here -- there is nothing
	// to rebuild it FROM. It is a refusal rather than a silent no-op because the operator
	// asked for a rotation and would otherwise believe one happened; the message has to
	// send them where the Secret actually is managed.
	t.Run("a secret we did not build cannot be rotated here", func(t *testing.T) {
		cfg := haCfg()
		cfg.K8s.TLSServerSecret = "byo-tls-secret" // Cert/CertKey left unset
		rr := &recRunner{}
		c := NewCluster(rr, cfg, nil, nil)
		err := c.UpdateServerCertSecret(context.Background())
		if err == nil {
			t.Fatal("rotating a Secret this env file supplies no files for must fail loud")
		}
		for _, want := range []string{"byo-tls-secret", "tls.cert", "cert-manager"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q should mention %q", err, want)
			}
		}
		// It refuses BEFORE the permission probe: nothing about the cluster is asked,
		// because nothing was ever going to be written.
		if len(rr.calls) != 0 {
			t.Errorf("no call may run at all; got %+v", rr.calls)
		}
	})
	// The half-supplied pair still fails, and still after the probe -- that one is a
	// genuine misconfiguration rather than a different deployment shape.
	t.Run("a half-supplied pair fails before any apply", func(t *testing.T) {
		cfg := haCfg()
		cfg.K8s.TLSServerSecret = "solace-tls-secret"
		cfg.TLS.Cert = filepath.Join(t.TempDir(), "tls.crt") // no certKey
		rr := &recRunner{}
		c := NewCluster(rr, cfg, nil, nil)
		if err := c.UpdateServerCertSecret(context.Background()); err == nil {
			t.Error("UpdateServerCertSecret should fail when tls.certKey is not configured")
		}
		if calls := rr.afterPreflight(t, "update", "secrets"); len(calls) != 0 {
			t.Errorf("no apply should run when TLSSecret fails to build; got %d calls after the probe", len(calls))
		}
	})
}

// --- node labelling: removed ----------------------------------------------
//
// The whole labelling surface is gone -- LabelNodes, promptNode, nodeNames, customLabels,
// splitLabel, isBuiltinLabel -- and so are the labelCluster helper and
// TestPromptsGoToErrNotOut, which drove it.
//
// This tool never labels cluster worker nodes. kubernetes.placement.labels are SELECTORS:
// internal/render emits them as the CR nodeSelector and affinity terms, and
// internal/render/render_test.go is where that is pinned. Nothing here parsed or applied
// them any more.
//
// It was also the one prerequisite that could not fold into an idempotent `broker deploy`:
// the picker needed a human at a terminal and recorded its choice nowhere, so a re-run
// could label a different node than the first run did.
//
// TestPromptsGoToErrNotOut pinned a real property -- a prompt belongs on stderr so that
// redirecting stdout captures results and not questions -- but this package no longer
// prompts at all. Its one remaining question, the operator downgrade, goes through the
// Confirm func seam, and every other confirmation lives in internal/cli. The Err/In pair
// this helper wired went with the picker.

// `broker generate` is only worth reading if what it prints is what `broker deploy`
// applies. These pin that equivalence from both ends: the stream carries every document
// the deploy applies, in the order it applies them.

// TestGenBrokerLeadsWithTheNamespace is the gap this closed. The stream used to start at
// the Secrets, so piping it at an empty cluster failed -- the Secrets and the CR are
// namespaced and the namespace was not there yet. `operator generate` had carried its own
// Namespace document from the start; this half had not.
func TestGenBrokerLeadsWithTheNamespace(t *testing.T) {
	cfg := adminCfg()
	got, err := GenBroker(cfg)
	if err != nil {
		t.Fatalf("GenBroker: %v", err)
	}
	docs := strings.Split(string(got), "---")
	if len(docs) < 3 {
		t.Fatalf("want at least namespace + secrets + CR, got %d documents:\n%s", len(docs), got)
	}
	if !strings.Contains(docs[0], "kind: Namespace") || !strings.Contains(docs[0], cfg.K8s.Namespace) {
		t.Errorf("the first document must be the broker namespace:\n%s", docs[0])
	}
	// Order is the whole point: a namespaced object ahead of its namespace does not apply.
	nsAt := strings.Index(string(got), "kind: Namespace")
	secretAt := strings.Index(string(got), "kind: Secret")
	crAt := strings.Index(string(got), "kind: PubSubPlusEventBroker")
	if !(nsAt < secretAt && secretAt < crAt) {
		t.Errorf("documents out of apply order (ns=%d secret=%d cr=%d):\n%s", nsAt, secretAt, crAt, got)
	}
}

// TestGenBrokerMatchesWhatDeployApplies walks the deploy for real over the echo seam and
// asserts every manifest it applies is a document of the generated stream. This is the
// assertion that catches the next divergence: a step added to the deploy and not to the
// stream fails here rather than in someone's `kubectl apply`.
func TestGenBrokerMatchesWhatDeployApplies(t *testing.T) {
	cfg := adminCfg()
	stream, err := GenBroker(cfg)
	if err != nil {
		t.Fatalf("GenBroker: %v", err)
	}

	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	ctx := context.Background()
	if err := c.CreateNamespace(ctx); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	if err := c.CreateSecrets(ctx); err != nil {
		t.Fatalf("CreateSecrets: %v", err)
	}
	if err := c.DeployBroker(ctx, false); err != nil {
		t.Fatalf("DeployBroker: %v", err)
	}

	var applied int
	for _, call := range rr.calls {
		if call.method != "RunInput" || call.stdin == "" {
			continue
		}
		applied++
		if !strings.Contains(string(stream), strings.Trim(call.stdin, "\n")) {
			t.Errorf("the deploy applies a manifest `broker generate` does not print:\n%s", call.stdin)
		}
	}
	if applied != 3 {
		t.Fatalf("expected the deploy to apply 3 manifests (namespace, secrets, CR), got %d", applied)
	}
}

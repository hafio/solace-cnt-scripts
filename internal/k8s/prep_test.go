package k8s

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"sort"
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
// manifest applied on stdin, and no secret value ever reaches the argv.
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

// TestCreateSecretsWithNothingToBuildDoesNothing: with no admin password, no TLS files, no
// registry credentials and no additional users there is no Secret of this tool's to
// create -- the CR references one that exists, or none and the operator generates it. So
// nothing runs at all: `kubectl apply` of an empty stream fails, and a permission probe
// for a create that will not happen demands a right the deploy never uses.
func TestCreateSecretsWithNothingToBuildDoesNothing(t *testing.T) {
	for _, adminSecret := range []string{"", "byo-admin"} {
		cfg := haCfg() // no semp.adminPass
		cfg.K8s.AdminSecret = adminSecret
		rr := &recRunner{}
		if err := NewCluster(rr, cfg, nil, nil).CreateSecrets(context.Background()); err != nil {
			t.Errorf("adminSecret=%q: nothing to build must not be an error: %v", adminSecret, err)
		}
		if len(rr.calls) != 0 {
			t.Errorf("adminSecret=%q: no call may run with nothing to create; got %+v", adminSecret, rr.calls)
		}
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
		// Same rule for the pull secret as the comment above states for TLS: a NAME
		// alone no longer makes it ours to delete (ManagesImagePullSecret). The
		// credentials are what put it in the delete set -- see the bring-your-own
		// subtest below.
		cfg.Image.User, cfg.Image.Pass = "u", "SECRET-REG-PASS"
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
	// The pull-secret analogue of the TLS subtest just above: a Secret this env
	// file only NAMES, with no credentials behind it, is the operator's own --
	// nothing here built it, so nothing here removes it.
	t.Run("an image-pull secret we did not create survives", func(t *testing.T) {
		cfg := adminCfg()
		cfg.K8s.ImagePullSecret = "byo-image-pull"
		rr := &recRunner{}
		c := NewCluster(rr, cfg, nil, nil)
		if err := c.DeleteSecrets(context.Background()); err != nil {
			t.Fatalf("DeleteSecrets: %v", err)
		}
		for _, call := range rr.afterPreflight(t, "delete", "secrets") {
			for _, a := range call.args {
				if a == "byo-image-pull" {
					t.Fatalf("a Secret this tool did not build must not be deleted: %+v", call)
				}
			}
		}
	})
	// The TLS analogue of the derived pull-secret case below: files supplied, no name
	// configured, so the Secret GenSecrets built is <kubernetes.name>-tls and that is the
	// name the teardown must delete.
	t.Run("derived TLS name when files are supplied and no name is configured", func(t *testing.T) {
		cfg := adminCfg()
		cfg.TLS.Cert, cfg.TLS.CertKey = "certs/tls.crt", "certs/tls.key"
		rr := &recRunner{}
		c := NewCluster(rr, cfg, nil, nil)
		if err := c.DeleteSecrets(context.Background()); err != nil {
			t.Fatalf("DeleteSecrets: %v", err)
		}
		want := []string{"delete", "secret", "dev-broker-tls", "-n", "solace", "--ignore-not-found"}
		var saw bool
		for _, call := range rr.afterPreflight(t, "delete", "secrets") {
			if eqArgs(call.args, want) {
				saw = true
			}
		}
		if !saw {
			t.Errorf("DeleteSecrets calls = %+v, want one deleting the derived TLS name %v", rr.calls, want)
		}
	})
	// With credentials present and no name configured, DeleteSecrets must name the
	// same DERIVED default GenSecrets built the Secret under -- not skip it, and not
	// invent a different name.
	t.Run("derived name when credentials are present and no name is configured", func(t *testing.T) {
		cfg := adminCfg()
		cfg.Image.User, cfg.Image.Pass = "u", "SECRET-REG-PASS"
		rr := &recRunner{}
		c := NewCluster(rr, cfg, nil, nil)
		if err := c.DeleteSecrets(context.Background()); err != nil {
			t.Fatalf("DeleteSecrets: %v", err)
		}
		calls := rr.afterPreflight(t, "delete", "secrets")
		want := []string{"delete", "secret", "dev-broker-image-pull", "-n", "solace", "--ignore-not-found"}
		var saw bool
		for _, call := range calls {
			if eqArgs(call.args, want) {
				saw = true
			}
		}
		if !saw {
			t.Errorf("DeleteSecrets calls = %+v, want one deleting the derived name %v", calls, want)
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

// TestDeleteSecretsSkipsUnconfiguredAdminSecret: with no Secret this tool built there is
// nothing to delete and no permission to probe. The operator's generated admin Secret is
// owned by the broker CR and goes with it.
func TestDeleteSecretsSkipsUnconfiguredAdminSecret(t *testing.T) {
	cfg := haCfg() // no semp.adminPass, no kubernetes.adminSecret, no TLS/pull secret
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.DeleteSecrets(context.Background()); err != nil {
		t.Fatalf("DeleteSecrets: %v", err)
	}
	if len(rr.calls) != 0 {
		t.Errorf("DeleteSecrets with nothing to delete must run nothing; got %+v", rr.calls)
	}
}

// TestDeleteSecretsLeavesAReferencedAdminSecret is the data-loss case: a named admin
// Secret with no password behind it is someone else's, and `broker remove` must never
// delete it -- while a Secret the same env file DID build is still removed.
func TestDeleteSecretsLeavesAReferencedAdminSecret(t *testing.T) {
	cfg := haCfg()
	// Referenced only: no semp.adminPass behind the name.
	cfg.K8s.AdminSecret = "byo-admin"
	// Built by this env file: TLS files with no name derive dev-broker-tls.
	cfg.TLS.Cert, cfg.TLS.CertKey = "certs/tls.crt", "certs/tls.key"
	rr := &recRunner{}
	if err := NewCluster(rr, cfg, nil, nil).DeleteSecrets(context.Background()); err != nil {
		t.Fatalf("DeleteSecrets: %v", err)
	}
	var sawTLS bool
	for _, call := range rr.afterPreflight(t, "delete", "secrets") {
		if strings.Contains(strings.Join(call.args, " "), "byo-admin") {
			t.Fatalf("a referenced admin Secret must never be deleted: %+v", call)
		}
		if eqArgs(call.args, []string{"delete", "secret", "dev-broker-tls", "-n", "solace", "--ignore-not-found"}) {
			sawTLS = true
		}
	}
	if !sawTLS {
		t.Errorf("the TLS Secret this env file built must still be deleted: %+v", rr.calls)
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
	t.Run("errors with no TLS Secret at all", func(t *testing.T) {
		rr := &recRunner{}
		c := NewCluster(rr, haCfg(), nil, nil) // no name and no files: nothing to update
		if err := c.UpdateServerCertSecret(context.Background()); err == nil {
			t.Error("UpdateServerCertSecret should fail when there is neither a TLS Secret name nor files")
		}
		if len(rr.calls) != 0 {
			t.Errorf("no call may run with nothing to update; got %+v", rr.calls)
		}
	})
	t.Run("rotates the derived name when only the files are supplied", func(t *testing.T) {
		dir := t.TempDir()
		crt := filepath.Join(dir, "tls.crt")
		key := filepath.Join(dir, "tls.key")
		writeFile(t, crt, "CERT\n")
		writeFile(t, key, "KEY\n")
		cfg := haCfg()
		cfg.TLS.Cert, cfg.TLS.CertKey = crt, key
		rr := &recRunner{}
		c := NewCluster(rr, cfg, nil, nil)
		if err := c.UpdateServerCertSecret(context.Background()); err != nil {
			t.Fatalf("UpdateServerCertSecret: %v", err)
		}
		if got := rr.last(); !strings.Contains(got.stdin, "name: dev-broker-tls") {
			t.Errorf("the rotation must apply the Secret under the derived name the CR references:\n%s", got.stdin)
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

// TestGenBrokerWithNoSecretsIsNamespaceAndCR is the operator-managed shape: no Secret of
// this tool's to build, so `broker generate` prints the Namespace and the CR with nothing
// between them -- no blank document -- and the deploy applies exactly those two, each a
// real manifest. The recRunner-based sibling above could not tell a blank apply from a
// real one, which is how an empty `kubectl apply -f -` would have slipped through.
func TestGenBrokerWithNoSecretsIsNamespaceAndCR(t *testing.T) {
	cfg := haCfg() // no semp.adminPass, no kubernetes.adminSecret, no TLS, pull or users
	stream, err := GenBroker(cfg)
	if err != nil {
		t.Fatalf("GenBroker: %v", err)
	}
	if n := strings.Count(string(stream), "\n---\n"); n != 1 {
		t.Errorf("want exactly one separator (Namespace, CR), got %d:\n%s", n, stream)
	}
	if strings.Contains(string(stream), "kind: Secret") {
		t.Errorf("no Secret is built, so none may be printed:\n%s", stream)
	}

	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	ctx := context.Background()
	for _, step := range []func(context.Context) error{c.CreateNamespace, c.CreateSecrets} {
		if err := step(ctx); err != nil {
			t.Fatalf("deploy step: %v", err)
		}
	}
	if err := c.DeployBroker(ctx, false); err != nil {
		t.Fatalf("DeployBroker: %v", err)
	}
	var applied int
	for _, call := range rr.calls {
		if call.method != "RunInput" {
			continue
		}
		if strings.TrimSpace(call.stdin) == "" {
			t.Errorf("a blank manifest reached `kubectl apply`: %+v", call)
			continue
		}
		applied++
	}
	if applied != 2 {
		t.Errorf("expected the deploy to apply 2 manifests (namespace, CR), got %d", applied)
	}
}

// TestAdminSecretStatesEmitTheirArtifacts pins everything the admin Secret's states emit,
// in one place and for both directions: what `broker generate` prints, what `broker
// deploy` applies, and what `broker remove` deletes. Each state is checked for the Secret
// documents in the stream, the CR's adminCredentialsSecret, the applies (every one a
// real, non-blank part of the printed stream), and the delete set. The two states in
// which this tool builds no admin Secret run beside a TLS Secret it does build, which is
// where a leak would show: the stream and the delete set must carry that Secret and still
// never the admin one.
func TestAdminSecretStatesEmitTheirArtifacts(t *testing.T) {
	withTLS := func(t *testing.T, c *config.Config) {
		t.Helper()
		dir := t.TempDir()
		c.TLS.Cert = writeTempPEM(t, dir, "tls.crt", "CERTIFICATE")
		c.TLS.CertKey = writeTempPEM(t, dir, "tls.key", "PRIVATE KEY")
	}
	cases := []struct {
		name     string
		set      func(*testing.T, *config.Config)
		secrets  []string // Secret documents printed, applied and deleted, sorted
		crAdmin  string   // the CR's adminCredentialsSecret; "" means the field is absent
		neverRef string   // a Secret no emitted artifact may build or delete
	}{
		{"password and a configured name", func(_ *testing.T, c *config.Config) {
			c.SEMP.AdminPass, c.K8s.AdminSecret = "pw", "solace-admin-secret"
		}, []string{"solace-admin-secret"}, "solace-admin-secret", ""},
		{"password and a derived name", func(_ *testing.T, c *config.Config) {
			c.SEMP.AdminPass = "pw"
		}, []string{"dev-broker-admin"}, "dev-broker-admin", ""},
		{"name only, beside a TLS Secret this tool builds", func(t *testing.T, c *config.Config) {
			c.K8s.AdminSecret = "byo-admin"
			withTLS(t, c)
		}, []string{"dev-broker-tls"}, "byo-admin", "byo-admin"},
		{"name only, nothing else", func(_ *testing.T, c *config.Config) {
			c.K8s.AdminSecret = "byo-admin"
		}, nil, "byo-admin", "byo-admin"},
		{"neither, beside a TLS Secret this tool builds", func(t *testing.T, c *config.Config) {
			withTLS(t, c)
		}, []string{"dev-broker-tls"}, "", "dev-broker-pubsubplus-admin-creds"},
		{"neither, nothing else", func(*testing.T, *config.Config) {}, nil, "", "dev-broker-pubsubplus-admin-creds"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := haCfg()
			tc.set(t, cfg)

			// broker generate
			stream, err := GenBroker(cfg)
			if err != nil {
				t.Fatalf("GenBroker: %v", err)
			}
			if got := secretNamesIn(string(stream)); !eqArgs(got, tc.secrets) {
				t.Errorf("generate prints Secrets %v, want %v:\n%s", got, tc.secrets, stream)
			}
			adminLine := "  adminCredentialsSecret: " + tc.crAdmin + "\n"
			switch {
			case tc.crAdmin != "" && !strings.Contains(string(stream), adminLine):
				t.Errorf("CR must reference the admin Secret %q:\n%s", tc.crAdmin, stream)
			case tc.crAdmin == "" && strings.Contains(string(stream), "adminCredentialsSecret"):
				t.Errorf("CR must name no admin Secret, so the operator generates one:\n%s", stream)
			}

			// broker deploy: exactly the printed documents, none blank
			rr := &recRunner{}
			c := NewCluster(rr, cfg, nil, nil)
			ctx := context.Background()
			for _, step := range []func(context.Context) error{c.CreateNamespace, c.CreateSecrets} {
				if err := step(ctx); err != nil {
					t.Fatalf("deploy step: %v", err)
				}
			}
			if err := c.DeployBroker(ctx, false); err != nil {
				t.Fatalf("DeployBroker: %v", err)
			}
			var applied []string
			var appliedSecrets []string
			for _, call := range rr.calls {
				if call.method != "RunInput" {
					continue
				}
				body := strings.Trim(call.stdin, "\n")
				if strings.TrimSpace(body) == "" {
					t.Errorf("a blank manifest reached `kubectl apply`: %+v", call)
					continue
				}
				if !strings.Contains(string(stream), body) {
					t.Errorf("deploy applies a manifest generate does not print:\n%s", call.stdin)
				}
				applied = append(applied, body)
				appliedSecrets = append(appliedSecrets, secretNamesIn(call.stdin)...)
			}
			sort.Strings(appliedSecrets)
			if !eqArgs(appliedSecrets, tc.secrets) {
				t.Errorf("deploy applies Secrets %v, want %v", appliedSecrets, tc.secrets)
			}
			wantApplies := 2 // Namespace, CR
			if len(tc.secrets) > 0 {
				wantApplies++ // the Secrets, as one stream
			}
			if len(applied) != wantApplies {
				t.Errorf("deploy made %d applies, want %d", len(applied), wantApplies)
			}

			// broker remove: exactly the Secrets this tool built
			rr = &recRunner{}
			if err := NewCluster(rr, cfg, nil, nil).DeleteSecrets(ctx); err != nil {
				t.Fatalf("DeleteSecrets: %v", err)
			}
			var deleted []string
			for _, call := range rr.calls {
				if len(call.args) > 2 && call.args[0] == "delete" && call.args[1] == "secret" {
					deleted = append(deleted, call.args[2])
				}
			}
			sort.Strings(deleted)
			if !eqArgs(deleted, tc.secrets) {
				t.Errorf("remove deletes Secrets %v, want %v", deleted, tc.secrets)
			}
			if tc.neverRef != "" {
				for _, name := range append(deleted, appliedSecrets...) {
					if name == tc.neverRef {
						t.Errorf("%s is not this tool's to build or delete", tc.neverRef)
					}
				}
			}
		})
	}
}

// secretNamesIn lists, sorted, the metadata.name of every Secret document in a
// multi-document stream.
func secretNamesIn(stream string) []string {
	var names []string
	for _, doc := range strings.Split(stream, "\n---\n") {
		if !strings.Contains("\n"+doc, "\nkind: Secret\n") {
			continue
		}
		for _, line := range strings.Split(doc, "\n") {
			if name, ok := strings.CutPrefix(line, "  name: "); ok {
				names = append(names, name)
				break
			}
		}
	}
	sort.Strings(names)
	return names
}

// TestGenSecretsBuildsNothingWithoutMaterial: nil, not an empty manifest, when this env
// file supplies nothing to build a Secret from.
func TestGenSecretsBuildsNothingWithoutMaterial(t *testing.T) {
	got, err := GenSecrets(haCfg())
	if err != nil || got != nil {
		t.Errorf("GenSecrets = (%q, %v), want (nil, nil)", got, err)
	}
}

// TestJoinManifestsDropsBlankDocuments: a blank part is left out rather than joined, so a
// stream never carries an empty document between two separators.
func TestJoinManifestsDropsBlankDocuments(t *testing.T) {
	got := string(joinManifests([][]byte{[]byte("a: 1\n"), nil, []byte("\n"), []byte("b: 2\n")}))
	if got != "a: 1\n---\nb: 2\n" {
		t.Errorf("joinManifests = %q, want the two real documents and one separator", got)
	}
}

// tlsCfg is haCfg with a real certificate pair on disk, so the TLS Secret is one this
// env file builds (config.Config.ManagesTLSSecret) under the derived dev-broker-tls.
func tlsCfg(t *testing.T) *config.Config {
	t.Helper()
	dir := t.TempDir()
	crt, key := filepath.Join(dir, "tls.crt"), filepath.Join(dir, "tls.key")
	writeFile(t, crt, "CERT\n")
	writeFile(t, key, "KEY\n")
	cfg := haCfg()
	cfg.TLS.Cert, cfg.TLS.CertKey = crt, key
	return cfg
}

// TestServerCertChanged covers the comparison a Kubernetes deploy makes before it
// rewrites the TLS Secret: the live Secret's digest annotation against the files on
// disk, read by jsonpath so the Secret's data -- the private key -- never comes back.
// Absent Secret, or none this env file builds: unchanged (nothing to compare, and a new
// TLS block rolls the pods through the CR). Equal: unchanged. Different, missing, or
// unreadable: changed.
func TestServerCertChanged(t *testing.T) {
	current := tlsDigest([]byte("KEY\n"), []byte("CERT\n"))
	for _, tc := range []struct {
		name string
		out  string
		err  error
		want bool
	}{
		{"no Secret yet", "", nil, false},
		{"same certificate", "dev-broker-tls\n" + current, nil, false},
		{"renewed certificate", "dev-broker-tls\n" + strings.Repeat("0", 64), nil, true},
		{"no annotation (an earlier build)", "dev-broker-tls\n", nil, true},
		{"unreadable", "", errors.New("forbidden"), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := &recRunner{out: []byte(tc.out), outErr: tc.err}
			c := NewCluster(rr, tlsCfg(t), nil, nil)
			got, err := c.ServerCertChanged(context.Background())
			if err != nil {
				t.Fatalf("ServerCertChanged: %v", err)
			}
			if got != tc.want {
				t.Errorf("changed = %t, want %t", got, tc.want)
			}
			call := rr.last()
			if !slices.Contains(call.args, "dev-broker-tls") || !slices.Contains(call.args, "--ignore-not-found") {
				t.Errorf("argv = %v, want a get of the derived Secret that tolerates its absence", call.args)
			}
			for _, a := range call.args {
				if a == "json" || a == "yaml" || (strings.HasPrefix(a, "jsonpath=") && strings.Contains(a, ".data")) {
					t.Errorf("argv = %v must read only the name and the annotation, never the data", call.args)
				}
			}
		})
	}
	t.Run("a Secret this env file does not build", func(t *testing.T) {
		cfg := haCfg()
		cfg.K8s.TLSServerSecret = "byo-tls-secret" // no files
		rr := &recRunner{}
		got, err := NewCluster(rr, cfg, nil, nil).ServerCertChanged(context.Background())
		if err != nil || got || len(rr.calls) != 0 {
			t.Errorf("changed=%t err=%v calls=%v, want an unchanged answer without asking the cluster", got, err, rr.calls)
		}
	})
}

// TestBrokerPodsExist: what "a broker is already running" means to a deploy deciding
// whether a renewed certificate needs a restart -- the primary pod is there. No answer,
// or a failed read, is false, so a preview never offers a restart.
func TestBrokerPodsExist(t *testing.T) {
	for _, tc := range []struct {
		out  string
		err  error
		want bool
	}{
		{"pod/dev-broker-pubsubplus-p-0\n", nil, true},
		{"", nil, false},
		{"", errors.New("unreachable"), false},
	} {
		rr := &recRunner{out: []byte(tc.out), outErr: tc.err}
		if got := NewCluster(rr, haCfg(), nil, nil).BrokerPodsExist(context.Background()); got != tc.want {
			t.Errorf("out=%q err=%v: BrokerPodsExist = %t, want %t", tc.out, tc.err, got, tc.want)
		}
		if call := rr.last(); !slices.Contains(call.args, "dev-broker-pubsubplus-p-0") {
			t.Errorf("argv = %v, want the primary pod", call.args)
		}
	}
}

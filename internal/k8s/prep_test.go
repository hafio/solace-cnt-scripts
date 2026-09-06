package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"solace/internal/config"
)

// adminCfg is haCfg with the minimum admin fields AdminSecret requires, so
// CreateSecrets/DeleteSecrets tests are hermetic (no dependency on env/sample.yaml).
func adminCfg() *config.Config {
	c := haCfg()
	c.Admin.Pass = "pw"
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

// TestDeleteNamespace exercises the empty-namespace path: the bare recRunner
// answers both enumeration reads with nothing (the Echo-runner shape getJSON
// treats as "zero items"), so nothing is FOREIGN and the delete proceeds
// exactly as it did before the B6 guard existed.
func TestDeleteNamespace(t *testing.T) {
	rr := &recRunner{}
	c := newCluster(rr)
	if err := c.DeleteNamespace(context.Background()); err != nil {
		t.Fatalf("DeleteNamespace: %v", err)
	}
	calls := rr.afterPreflight(t, "delete", "namespaces")
	if len(calls) != 3 {
		t.Fatalf("DeleteNamespace made %d calls after the probe, want 2 enumeration gets + 1 delete", len(calls))
	}
	got := calls[2]
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

// nsCluster wires a Cluster to a captured Out buffer, the same shape
// labelCluster gives LabelNodes, so a test can assert on the foreign-object
// report DeleteNamespace prints ahead of its refusal.
func nsCluster(rr *recRunner, cfg *config.Config) (*Cluster, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return &Cluster{R: rr, Cfg: cfg, Out: buf}, buf
}

// nsBuiltinJSON and nsBrokerJSON build the two enumeration replies
// namespaceObjects reads, in the shape each of its two calls actually decodes:
// nsRawList carries each item's own kind, nsCRList does not (see nsResourceKinds
// and brokerCRKind for why).
func nsBuiltinJSON(items string) []byte {
	return []byte(`{"kind":"List","items":[` + items + `]}`)
}
func nsBrokerJSON(items string) []byte {
	return []byte(`{"kind":"PubSubPlusEventBrokerList","items":[` + items + `]}`)
}
func nsItem(kind, name string) string {
	return fmt.Sprintf(`{"kind":%q,"metadata":{"name":%q}}`, kind, name)
}

// TestDeleteNamespaceOnlyOurs covers every OURS rule at once: the brokerSuffix
// name match (pod, PVC), the configured admin secret, and the two
// Kubernetes-generated exceptions (kube-root-ca.crt, the default
// ServiceAccount). None of it is FOREIGN, so the delete proceeds.
func TestDeleteNamespaceOnlyOurs(t *testing.T) {
	cfg := adminCfg() // dev-broker, admin secret "solace-admin-secret"
	rr := &recRunner{outQueue: [][]byte{
		nsBuiltinJSON(strings.Join([]string{
			nsItem("Pod", "dev-broker-pubsubplus-p-0"),
			nsItem("PersistentVolumeClaim", "data-dev-broker-pubsubplus-p-0"),
			nsItem("Secret", "solace-admin-secret"),
			nsItem("ConfigMap", "kube-root-ca.crt"),
			nsItem("ServiceAccount", "default"),
		}, ",")),
		nsBrokerJSON(""),
	}}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.DeleteNamespace(context.Background()); err != nil {
		t.Fatalf("DeleteNamespace: %v", err)
	}
	calls := rr.afterPreflight(t, "delete", "namespaces")
	if len(calls) != 3 {
		t.Fatalf("DeleteNamespace made %d calls after the probe, want 2 enumeration gets + 1 delete", len(calls))
	}
	got := calls[2]
	want := []string{"delete", "namespace", "solace", "--ignore-not-found"}
	if got.method != "Run" || !eqArgs(got.args, want) {
		t.Errorf("final call = %+v, want Run kubectl %v", got, want)
	}
}

// TestDeleteNamespaceForeignRefuses covers the hard-refusal path: a Deployment
// this tool did not create means the delete never runs, and the object is
// named in the report on Out.
func TestDeleteNamespaceForeignRefuses(t *testing.T) {
	cfg := adminCfg()
	rr := &recRunner{outQueue: [][]byte{
		nsBuiltinJSON(nsItem("Deployment", "unrelated-app")),
		nsBrokerJSON(""),
	}}
	c, buf := nsCluster(rr, cfg)
	err := c.DeleteNamespace(context.Background())
	if err == nil {
		t.Fatal("DeleteNamespace should refuse when a foreign object is present")
	}
	if !strings.Contains(err.Error(), "1 resource") {
		t.Errorf("error = %v, want it to say how many foreign resources were found", err)
	}
	if !strings.Contains(err.Error(), "remove broker") || !strings.Contains(err.Error(), "remove secrets") {
		t.Errorf("error = %v, want it to point at remove broker / remove secrets", err)
	}
	if !strings.Contains(buf.String(), "Deployment") || !strings.Contains(buf.String(), "unrelated-app") {
		t.Errorf("foreign object must be named in the report; got %q", buf.String())
	}
	for _, call := range rr.calls {
		if len(call.args) > 1 && call.args[0] == "delete" && call.args[1] == "namespace" {
			t.Error("no delete namespace call must be issued when foreign objects are present")
		}
	}
}

// TestDeleteNamespaceForeignCannotBeSilenced pins that the guard has no
// override: DeleteNamespace takes no "skip the check" parameter, so calling it
// exactly as the CLI's --no-prompt path does -- straight through, with no
// question asked first -- still refuses. There is nothing upstream of this
// method left to silence.
func TestDeleteNamespaceForeignCannotBeSilenced(t *testing.T) {
	cfg := adminCfg()
	rr := &recRunner{outQueue: [][]byte{
		nsBuiltinJSON(nsItem("Secret", "someone-elses-secret")),
		nsBrokerJSON(""),
	}}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.DeleteNamespace(context.Background()); err == nil {
		t.Fatal("DeleteNamespace must refuse even called directly, with no prompt in front of it to skip")
	}
}

// TestDeleteNamespaceForeignBrokerRefuses is the case the whole guard was asked
// for: a SECOND Solace broker sharing this namespace. Its objects carry the
// operator's "-pubsubplus" suffix exactly as ours do, so a rule matching the
// bare suffix would call them ours and delete another team's broker with the
// namespace. isOurs matches "<kubernetes.name>-pubsubplus" instead, which keeps
// them foreign.
func TestDeleteNamespaceForeignBrokerRefuses(t *testing.T) {
	cfg := adminCfg() // kubernetes.name is "dev-broker"
	rr := &recRunner{outQueue: [][]byte{
		nsBuiltinJSON(strings.Join([]string{
			nsItem("Pod", "dev-broker-pubsubplus-p-0"), // ours
			nsItem("Pod", "other-team-pubsubplus-p-0"), // another broker entirely
			nsItem("Service", "other-team-pubsubplus"), // and its LB service
		}, ",")),
		nsBrokerJSON(""),
	}}
	c, buf := nsCluster(rr, cfg)
	err := c.DeleteNamespace(context.Background())
	if err == nil {
		t.Fatal("DeleteNamespace must refuse when another broker shares the namespace")
	}
	if !strings.Contains(buf.String(), "other-team-pubsubplus-p-0") {
		t.Errorf("the other broker's pod must be reported as foreign; got %q", buf.String())
	}
	if strings.Contains(buf.String(), "dev-broker-pubsubplus-p-0") {
		t.Errorf("this deployment's own pod must NOT be listed as foreign; got %q", buf.String())
	}
	for _, call := range rr.calls {
		if len(call.args) > 1 && call.args[0] == "delete" && call.args[1] == "namespace" {
			t.Error("no delete namespace call must be issued when another broker is present")
		}
	}
}

// TestDeleteNamespaceBrokerListingFailsStillProceeds covers the deliberate
// asymmetry in namespaceObjects: the broker-CR listing failing is NOT treated
// like the built-in listing failing. An absent CRD is a state this tool's own
// documented order produces (remove the brokers, `remove operator
// --delete-crd`, then `remove namespace`), and refusing there would make a
// legitimate teardown impossible forever. It warns and judges the namespace by
// what it could see.
func TestDeleteNamespaceBrokerListingFailsStillProceeds(t *testing.T) {
	cfg := adminCfg()
	rr := &recRunner{
		outQueue:    [][]byte{nsBuiltinJSON(nsItem("ConfigMap", "kube-root-ca.crt"))},
		outErrQueue: []error{nil, errFake}, // the built-in listing answers; the CR listing does not
	}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.DeleteNamespace(context.Background()); err != nil {
		t.Fatalf("an unlistable broker CRD must not block a teardown of an otherwise-empty namespace: %v", err)
	}
	calls := rr.afterPreflight(t, "delete", "namespaces")
	last := calls[len(calls)-1]
	want := []string{"delete", "namespace", "solace", "--ignore-not-found"}
	if last.method != "Run" || !eqArgs(last.args, want) {
		t.Errorf("final call = %+v, want Run kubectl %v", last, want)
	}
}

// TestDeleteNamespaceEnumerationFails proves a failed listing (RBAC denial, an
// unreachable API server) refuses the delete too, wrapping the cause: this
// guard exists to stop an unseen deletion, so being unable to see is not
// permission to proceed.
func TestDeleteNamespaceEnumerationFails(t *testing.T) {
	rr := &recRunner{outErr: errFake}
	c := newCluster(rr)
	err := c.DeleteNamespace(context.Background())
	if err == nil {
		t.Fatal("DeleteNamespace should refuse when it cannot enumerate the namespace")
	}
	if !errors.Is(err, errFake) {
		t.Errorf("error = %v, want it to wrap the underlying failure", err)
	}
	for _, call := range rr.calls {
		if len(call.args) > 1 && call.args[0] == "delete" && call.args[1] == "namespace" {
			t.Error("no delete namespace call must be issued when enumeration fails")
		}
	}
}

// TestDeleteNamespaceEmpty covers a namespace with nothing in it at all --
// exactly what "remove all" leaves behind, since it deletes the broker and
// secrets before reaching DeleteNamespace.
func TestDeleteNamespaceEmpty(t *testing.T) {
	rr := &recRunner{outQueue: [][]byte{
		nsBuiltinJSON(""),
		nsBrokerJSON(""),
	}}
	c := newCluster(rr)
	if err := c.DeleteNamespace(context.Background()); err != nil {
		t.Fatalf("DeleteNamespace: %v", err)
	}
	calls := rr.afterPreflight(t, "delete", "namespaces")
	got := calls[len(calls)-1]
	want := []string{"delete", "namespace", "solace", "--ignore-not-found"}
	if got.method != "Run" || !eqArgs(got.args, want) {
		t.Errorf("final call = %+v, want Run kubectl %v", got, want)
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

// TestCreateSecretsPreflight fails loud when the TLS secret is requested but its
// inputs are missing -- before any apply runs (012:19-24).
func TestCreateSecretsPreflight(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(c *config.Config)
	}{
		{"cert/key fields unset", func(c *config.Config) { c.TLS.Cert = ""; c.TLS.CertKey = "" }},
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
// (`prep secrets --gen-secrets-only`), which calls GenSecrets directly and bypasses
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
	// TestUpdateServerCertSecret/configured secret name but cert files unset proves
	// the more likely real case: UpdateServerCertSecret's own precondition only
	// checks kubernetes.tlsServerSecret is set, not that Cert/CertKey are also configured, so a
	// configured name with missing cert inputs must still fail before any apply.
	t.Run("configured secret name but cert files unset", func(t *testing.T) {
		cfg := haCfg()
		cfg.K8s.TLSServerSecret = "solace-tls-secret" // Cert/CertKey left unset
		rr := &recRunner{}
		c := NewCluster(rr, cfg, nil, nil)
		if err := c.UpdateServerCertSecret(context.Background()); err == nil {
			t.Error("UpdateServerCertSecret should fail when tls.cert/tls.certKey are not configured")
		}
		// The permission probe runs first (it precedes reading the cert files); no
		// apply may follow it.
		if calls := rr.afterPreflight(t, "update", "secrets"); len(calls) != 0 {
			t.Errorf("no apply should run when TLSSecret fails to build; got %d calls after the probe", len(calls))
		}
	})
}

// --- pure label helpers ----------------------------------------------------

func TestSplitLabel(t *testing.T) {
	cases := []struct {
		in       string
		key, val string
		ok       bool
	}{
		{"topology.kubernetes.io/zone=z1", "topology.kubernetes.io/zone", "z1", true},
		{"zone: z1", "zone", "z1", true},
		{"zone:z1", "zone", "z1", true},
		{"  a = b ", "a", "b", true},
		{"nosep", "", "", false},
		{"=z1", "", "", false},
		{"zone=", "", "", false},
		{": v", "", "", false},
	}
	for _, tc := range cases {
		key, val, ok := splitLabel(tc.in)
		if ok != tc.ok || key != tc.key || val != tc.val {
			t.Errorf("splitLabel(%q) = (%q,%q,%v), want (%q,%q,%v)", tc.in, key, val, ok, tc.key, tc.val, tc.ok)
		}
	}
}

func TestIsBuiltinLabel(t *testing.T) {
	builtin := []string{"kubernetes.io/arch", "beta.kubernetes.io/os", "node.kubernetes.io/instance-type", "k8s.io/foo"}
	custom := []string{"topology.kubernetes.io/zone", "acme.com/tier", "zone"}
	for _, k := range builtin {
		if !isBuiltinLabel(k) {
			t.Errorf("isBuiltinLabel(%q) = false, want true", k)
		}
	}
	for _, k := range custom {
		if isBuiltinLabel(k) {
			t.Errorf("isBuiltinLabel(%q) = true, want false", k)
		}
	}
}

// --- LabelNodes ------------------------------------------------------------

// labelCluster builds a Cluster wired for interactive labelling: rr as runner, buf
// as the report sink, and stdin as the selection source.
// labelCluster wires all three narration channels -- the report sink, the prompt
// sink and the progress line sink -- to ONE buffer, so a test that only cares
// "was this said" need not know which stream said it.
// TestPromptsGoToErrNotOut is the test that splits them apart on purpose.
func labelCluster(rr *recRunner, cfg *config.Config, stdin string) (*Cluster, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	c := &Cluster{
		R:   rr,
		Cfg: cfg,
		Out: buf,
		Err: buf,
		In:  strings.NewReader(stdin),
		Log: func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) },
	}
	return c, buf
}

func TestLabelNodesNoCustomLabels(t *testing.T) {
	rr := &recRunner{}
	c, buf := labelCluster(rr, saCfg(), "")
	if err := c.LabelNodes(context.Background()); err != nil {
		t.Fatalf("LabelNodes: %v", err)
	}
	if len(rr.calls) != 0 {
		t.Errorf("no cluster calls expected with no labels; got %d", len(rr.calls))
	}
	if !strings.Contains(buf.String(), "nothing to label") {
		t.Errorf("expected an early-exit message; got %q", buf.String())
	}
}

func TestLabelNodesBuiltinOnly(t *testing.T) {
	cfg := saCfg()
	cfg.K8s.Placement.LabelsPrimary = []string{"kubernetes.io/arch: amd64"}
	rr := &recRunner{}
	c, _ := labelCluster(rr, cfg, "")
	if err := c.LabelNodes(context.Background()); err != nil {
		t.Fatalf("LabelNodes: %v", err)
	}
	if len(rr.calls) != 0 {
		t.Errorf("built-in labels must not be applied; got %d calls", len(rr.calls))
	}
}

func TestLabelNodesMalformedAndUnsafe(t *testing.T) {
	cfg := saCfg()
	cfg.K8s.Placement.LabelsPrimary = []string{"nosep", "bad key=v"}
	rr := &recRunner{}
	c, buf := labelCluster(rr, cfg, "")
	if err := c.LabelNodes(context.Background()); err != nil {
		t.Fatalf("LabelNodes: %v", err)
	}
	if len(rr.calls) != 0 {
		t.Errorf("dropped labels must not reach the cluster; got %d calls", len(rr.calls))
	}
	out := buf.String()
	if !strings.Contains(out, "malformed") || !strings.Contains(out, "unsupported characters") {
		t.Errorf("expected warnings for both dropped labels; got %q", out)
	}
}

func TestLabelNodesHappyPath(t *testing.T) {
	cfg := saCfg()
	cfg.K8s.Placement.LabelsPrimary = []string{"topology.kubernetes.io/zone: z1"}
	rr := &recRunner{outQueue: [][]byte{
		// The `auth can-i` probe is answered out of band by recRunner.canI, so the
		// queue holds only the reads LabelNodes makes itself.
		[]byte("node-a\nnode-b\n"), // node list
	}}
	c, _ := labelCluster(rr, cfg, "1\n") // pick node-a
	if err := c.LabelNodes(context.Background()); err != nil {
		t.Fatalf("LabelNodes: %v", err)
	}
	// calls: Output(auth can-i), Output(get nodes), Run(label ...)
	if len(rr.calls) != 3 {
		t.Fatalf("LabelNodes made %d calls, want 3", len(rr.calls))
	}
	if a := rr.calls[0].args; a[0] != "auth" || a[1] != "can-i" || a[2] != "update" || a[3] != "nodes" {
		t.Errorf("first call should be the RBAC precheck; got %v", a)
	}
	got := rr.calls[2]
	want := []string{"label", "node", "node-a", "topology.kubernetes.io/zone=z1", "--overwrite"}
	if got.method != "Run" || !eqArgs(got.args, want) {
		t.Errorf("label call = %+v, want Run kubectl %v", got, want)
	}
}

func TestLabelNodesReprompt(t *testing.T) {
	cfg := saCfg()
	cfg.K8s.Placement.LabelsPrimary = []string{"zone: z2"}
	rr := &recRunner{outQueue: [][]byte{
		[]byte("node-a\nnode-b\n"),
	}}
	c, buf := labelCluster(rr, cfg, "9\nx\n2\n") // out-of-range, non-numeric, then node-b
	if err := c.LabelNodes(context.Background()); err != nil {
		t.Fatalf("LabelNodes: %v", err)
	}
	got := rr.last()
	want := []string{"label", "node", "node-b", "zone=z2", "--overwrite"}
	if !eqArgs(got.args, want) {
		t.Errorf("after reprompt, label call = %v, want %v", got.args, want)
	}
	if strings.Count(buf.String(), "Invalid selection") != 2 {
		t.Errorf("expected two re-prompts; got %q", buf.String())
	}
}

func TestLabelNodesRBACDenied(t *testing.T) {
	cfg := saCfg()
	cfg.K8s.Placement.LabelsPrimary = []string{"zone: z1"}
	rr := &recRunner{outErr: errFake}
	c, _ := labelCluster(rr, cfg, "")
	if err := c.LabelNodes(context.Background()); err == nil {
		t.Fatal("LabelNodes should fail when the RBAC precheck fails")
	}
	for _, call := range rr.calls {
		if len(call.args) > 0 && call.args[0] == "label" {
			t.Error("no node must be labelled after an RBAC denial")
		}
	}
}

func TestLabelNodesEOFNoSelection(t *testing.T) {
	cfg := saCfg()
	cfg.K8s.Placement.LabelsPrimary = []string{"zone: z1"}
	rr := &recRunner{outQueue: [][]byte{
		[]byte("node-a\n"),
	}}
	c, _ := labelCluster(rr, cfg, "") // EOF before any selection
	if err := c.LabelNodes(context.Background()); err == nil {
		t.Error("LabelNodes should fail when no node selection is provided")
	}
}

// TestLabelNodesHAOnlyPrimaryConfigured covers rolePlacementLabels' Backup/Monitor
// arms and LabelNodes' per-role skip-continue: every other LabelNodes test uses
// saCfg() (standalone), so HARoles(cfg) only ever yielded Primary and these branches
// never executed. This is the "which config field feeds which role" branch that
// could silently break (e.g. Backup reading LabelsPrimary by mistake) with nothing
// to catch it.
func TestLabelNodesHAOnlyPrimaryConfigured(t *testing.T) {
	cfg := haCfg()                                         // HARoles = [primary, backup, monitor]
	cfg.K8s.Placement.LabelsPrimary = []string{"zone: z1"} // Backup/Monitor left empty
	rr := &recRunner{outQueue: [][]byte{
		[]byte("node-a\nnode-b\n"),
	}}
	c, buf := labelCluster(rr, cfg, "1\n")
	if err := c.LabelNodes(context.Background()); err != nil {
		t.Fatalf("LabelNodes: %v", err)
	}
	var labelCalls int
	for _, call := range rr.calls {
		if len(call.args) > 0 && call.args[0] == "label" {
			labelCalls++
		}
	}
	if labelCalls != 1 {
		t.Errorf("only the primary role has labels configured; got %d label calls, want 1", labelCalls)
	}
	out := buf.String()
	if strings.Contains(out, "Select the node for the backup broker") || strings.Contains(out, "Select the node for the monitor broker") {
		t.Errorf("backup/monitor must not be prompted when they have no configured labels:\n%s", out)
	}
}

// TestNodeNamesError proves nodeNames' own error-wrap fires on a genuine query
// failure: LabelNodes relies on this failing loud rather than silently prompting
// with the wrong/empty node list when the cluster is flaky mid-operation.
func TestNodeNamesError(t *testing.T) {
	rr := &recRunner{outErr: errFake}
	c := newCluster(rr)
	_, err := c.nodeNames(context.Background())
	if err == nil || !strings.Contains(err.Error(), "listing cluster nodes") {
		t.Errorf("nodeNames error = %v, want it to wrap \"listing cluster nodes\"", err)
	}
}

// TestLabelNodesNoNodesFound covers the edge condition where labelling is allowed
// (RBAC passes) but the cluster reports zero nodes: LabelNodes must fail loud
// rather than doing nothing or misbehaving in promptNode with an empty node list.
func TestLabelNodesNoNodesFound(t *testing.T) {
	cfg := saCfg()
	cfg.K8s.Placement.LabelsPrimary = []string{"zone: z1"}
	rr := &recRunner{outQueue: [][]byte{
		[]byte(""), // no nodes reported (the `auth can-i` probe passes via recRunner.canI)
	}}
	c, _ := labelCluster(rr, cfg, "")
	err := c.LabelNodes(context.Background())
	if err == nil || !strings.Contains(err.Error(), "no cluster nodes found") {
		t.Errorf("LabelNodes error = %v, want \"no cluster nodes found\"", err)
	}
}

// TestLabelNodesLabelFailureIsNonFatal pins the doc comment's own contract (and
// CLAUDE.md's fail-loud-vs-skip principle): a single `kubectl label` failure is
// reported and skipped, not fatal. Every happy-path test until now succeeded on
// every label call, so this branch had zero coverage.
func TestLabelNodesLabelFailureIsNonFatal(t *testing.T) {
	cfg := saCfg()
	cfg.K8s.Placement.LabelsPrimary = []string{"zone: z1"}
	rr := &recRunner{
		outQueue: [][]byte{[]byte("node-a\n")},
		runErr:   errFake, // fails the kubectl label call itself
	}
	c, buf := labelCluster(rr, cfg, "1\n")
	if err := c.LabelNodes(context.Background()); err != nil {
		t.Fatalf("a single label failure must not abort LabelNodes, got %v", err)
	}
	if !strings.Contains(buf.String(), "[FAIL] failed to apply") {
		t.Errorf("expected a reported label failure; got %q", buf.String())
	}
}

// TestPromptsGoToErrNotOut pins the stream split the node picker used to get
// wrong: it wrote its banner, list and "> " to Out, so `prepare labels > file`
// captured the questions along with the results. The prompt belongs on Err with
// every other prompt in this tool; the per-label outcome stays on Out, which is
// what a script actually wants to read.
func TestPromptsGoToErrNotOut(t *testing.T) {
	cfg := saCfg()
	cfg.K8s.Placement.LabelsPrimary = []string{"zone: z1"}
	rr := &recRunner{outQueue: [][]byte{[]byte("node-a\n")}}
	var out, errBuf bytes.Buffer
	c := &Cluster{R: rr, Cfg: cfg, Out: &out, Err: &errBuf, In: strings.NewReader("1\n")}
	if err := c.LabelNodes(context.Background()); err != nil {
		t.Fatalf("LabelNodes: %v", err)
	}
	if !strings.Contains(errBuf.String(), "Select the node for the primary broker") {
		t.Errorf("prompt must go to Err; Err = %q", errBuf.String())
	}
	if strings.Contains(out.String(), "Select the node") || strings.Contains(out.String(), "> ") {
		t.Errorf("prompt must NOT reach Out; Out = %q", out.String())
	}
	if !strings.Contains(out.String(), "[ OK ] labelled node-a with zone=z1") {
		t.Errorf("the per-label outcome belongs on Out; Out = %q", out.String())
	}
}

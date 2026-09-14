package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

// TestWatchNamespace pins the WATCH_NAMESPACE join logic (000-env.sh:85-89): the
// broker namespace is appended by default, comma-joined onto any configured list, and
// omitted when broker-ns watching is explicitly disabled.
//
// The dedupe cases are the fix for a report and a manifest that both listed the broker
// namespace twice whenever the configured list already named it (the common case, since
// watchBrokerNs defaults on). controller-runtime's map-keyed cache collapsed the repeat,
// so only the printed and applied text was ever wrong -- which is precisely what makes a
// regression here invisible without these cases.
func TestWatchNamespace(t *testing.T) {
	cases := []struct {
		name     string
		list     string
		watchBrk *bool
		want     string
	}{
		{"default appends broker ns", "", nil, "solace"},
		{"appends onto configured list", "ns-a,ns-b", nil, "ns-a,ns-b,solace"},
		{"disabled keeps only the list", "ns-a,ns-b", boolPtr(false), "ns-a,ns-b"},
		{"disabled with empty list is empty", "", boolPtr(false), ""},
		{"broker ns already listed is not repeated", "ns-a,solace", nil, "ns-a,solace"},
		{"repeat inside the list is dropped", "ns-a,ns-b,ns-a", nil, "ns-a,ns-b,solace"},
		{"entries are trimmed, empties dropped", " ns-a , ,ns-b,", nil, "ns-a,ns-b,solace"},
		{"disabled dedupes the list too", "solace,ns-b,solace", boolPtr(false), "solace,ns-b"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := haCfg() // Namespace = solace
			cfg.K8s.Operator.WatchNamespaces = tc.list
			cfg.K8s.Operator.WatchBrokerNS = tc.watchBrk
			if got := watchNamespace(cfg); got != tc.want {
				t.Errorf("watchNamespace = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestOperatorImage pins the registry-prefix rule now that RenderOperator and CheckEnv
// share it: the report drifted from the apply for exactly as long as each had its own
// idea of the reference, so the helper's two branches are worth their own test rather
// than only being reached through the 119 KB bundle render.
func TestOperatorImage(t *testing.T) {
	cfg := haCfg()
	cfg.K8s.Operator.Image = "solace/pubsubplus-eventbroker-operator:1.4.0"
	cfg.Image.Registry = "registry.example.com"
	if got, want := operatorImage(cfg), "registry.example.com/solace/pubsubplus-eventbroker-operator:1.4.0"; got != want {
		t.Errorf("operatorImage with a registry = %q, want %q", got, want)
	}
	cfg.Image.Registry = ""
	if got, want := operatorImage(cfg), "solace/pubsubplus-eventbroker-operator:1.4.0"; got != want {
		t.Errorf("operatorImage without a registry = %q, want %q", got, want)
	}
}

// TestRenderOperatorSubstitutions renders the bundle for known inputs and asserts each
// of the six substitution points landed, and that no template marker survives. The
// bundle is ~119 KB of otherwise-static CRD/RBAC YAML, so pinning the varying points is
// the reviewable check (vs. a full-bundle golden).
func TestRenderOperatorSubstitutions(t *testing.T) {
	t.Run("registry prefix, pull secret, broker-ns watch", func(t *testing.T) {
		cfg := loadK8s(t) // sample: registry set, pullSecret set, watch defaults on
		out, err := RenderOperator(cfg, "op-ns", watchNamespace(cfg))
		if err != nil {
			t.Fatalf("RenderOperator: %v", err)
		}
		s := string(out)
		mustContain(t, s, "  name: op-ns\n")
		mustContain(t, s, "  namespace: op-ns\n")
		mustContain(t, s, `value: "solace"`) // WATCH_NAMESPACE = broker ns
		mustContain(t, s, "image: registry.example.com/solace/pubsubplus-eventbroker-operator:1.4.2")
		mustContain(t, s, "cpu: 500m")
		mustContain(t, s, "memory: 512Mi")
		mustContain(t, s, "imagePullSecrets:")
		mustContain(t, s, "- name: regcred")
		// The reference is a name; the Secret itself is GenOperatorSecrets' artifact,
		// so the bundle stays shareable even with pull credentials configured.
		if strings.Contains(s, ".dockerconfigjson") {
			t.Errorf("the bundle must carry no secret document, only the imagePullSecrets reference:\n%s", s)
		}
		if strings.Contains(s, "{{") {
			t.Error("rendered bundle still contains an unresolved template marker {{")
		}
	})

	t.Run("no registry, no pull secret, watch list without broker ns", func(t *testing.T) {
		cfg := loadK8s(t)
		cfg.Image.Registry = ""
		cfg.K8s.ImagePullSecret = ""
		cfg.K8s.Operator.WatchNamespaces = "team-a"
		cfg.K8s.Operator.WatchBrokerNS = boolPtr(false)
		out, err := RenderOperator(cfg, "op-ns", watchNamespace(cfg))
		if err != nil {
			t.Fatalf("RenderOperator: %v", err)
		}
		s := string(out)
		mustContain(t, s, "image: solace/pubsubplus-eventbroker-operator:1.4.2")
		mustContain(t, s, `value: "team-a"`)
		if strings.Contains(s, "imagePullSecrets:") {
			t.Error("imagePullSecrets block must be omitted when no pull secret is configured")
		}
		if strings.Contains(s, "- name: regcred") {
			t.Error("regcred reference must be omitted when no pull secret is configured")
		}
	})
}

// TestRenderOperatorHonoursThePassedWatchList is why the watch list is a PARAMETER rather
// than something RenderOperator derives from the config itself.
//
// `operator deploy` reconciles what this env file asks for against what the installed
// operator already watches, and applies the UNION -- so a second env file adding its own
// namespace does not silently stop the first broker from being reconciled. That union is
// computed by the caller and cannot be recovered from the config, so a renderer that
// recomputed the list would quietly discard it and re-narrow the operator on every deploy.
//
// The list passed here is deliberately unrelated to what the config would produce, so the
// test fails if the parameter is ignored.
func TestRenderOperatorHonoursThePassedWatchList(t *testing.T) {
	cfg := loadK8s(t)
	fromConfig := watchNamespace(cfg)
	union := fromConfig + ",team-a,team-b"

	out, err := RenderOperator(cfg, "op-ns", union)
	if err != nil {
		t.Fatalf("RenderOperator: %v", err)
	}
	s := string(out)
	mustContain(t, s, `value: "`+union+`"`)
	if fromConfig != "" && strings.Contains(s, `value: "`+fromConfig+`"`) {
		t.Errorf("the rendered WATCH_NAMESPACE is the config's own list, not the one passed: " +
			"a reconciled union would be discarded")
	}
}

// TestGenOperator covers the render-only path (`generate operator`): it uses the
// configured operator namespace when set, and falls back to the fixed default when
// unset (render-only cannot discover the running deployment). It is otherwise
// RenderOperator, so a spot-check of the namespace substitution suffices.
func TestGenOperator(t *testing.T) {
	t.Run("uses configured operator namespace", func(t *testing.T) {
		cfg := loadK8s(t)
		cfg.K8s.Operator.Namespace = "my-op-ns"
		out, err := GenOperator(cfg)
		if err != nil {
			t.Fatalf("GenOperator: %v", err)
		}
		mustContain(t, string(out), "  namespace: my-op-ns\n")
	})
	t.Run("falls back to the default when unset", func(t *testing.T) {
		cfg := loadK8s(t)
		cfg.K8s.Operator.Namespace = ""
		out, err := GenOperator(cfg)
		if err != nil {
			t.Fatalf("GenOperator: %v", err)
		}
		mustContain(t, string(out), "  namespace: "+defaultOperatorNS+"\n")
	})
}

// TestOperatorApply asserts the apply sequence: the bundle's own Namespace document,
// then the regcred into that namespace, then the rest of the bundle -- each on stdin
// via `apply -f -`. The split exists so nothing applied is absent from a `generate`
// output; the ordering is what makes a first install work at all.
func TestOperatorApply(t *testing.T) {
	cfg := loadK8s(t) // pull secret set -> regcred applied
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.OperatorApply(context.Background()); err != nil {
		t.Fatalf("OperatorApply: %v", err)
	}
	calls := rr.afterPreflights(t, probe{verb: "create", resource: "customresourcedefinitions"})
	if len(calls) != 5 {
		t.Fatalf("OperatorApply made %d calls after the probe, want 5 "+
			"(version read + watch-list read + namespace + regcred + bundle)", len(calls))
	}
	// Both reads are cluster-wide `get deployment`: one resolves the installed version
	// for the downgrade check, the other the installed WATCH_NAMESPACE so the applied
	// value can be the union rather than this env file's list alone.
	for i, read := range []rrCall{calls[0], calls[1]} {
		if read.method != "Output" || !eqArgs(read.args,
			[]string{"get", "deployment", "--all-namespaces", "-o", "json"}) {
			t.Errorf("read %d = %+v, want the cluster-wide deployment read", i+1, read)
		}
	}
	ns, regcred, bundle := calls[2], calls[3], calls[4]
	for i, call := range []rrCall{ns, regcred, bundle} {
		if call.method != "RunInput" || call.name != "kubectl" || !eqArgs(call.args, []string{"apply", "-f", "-"}) {
			t.Errorf("apply call %d = %+v, want RunInput kubectl [apply -f -]", i+1, call)
		}
	}
	// The namespace goes alone and first: the regcred lives in it, and on a first
	// install applying the secret before the namespace died with `namespaces "..."
	// not found` before anything had been applied.
	if !strings.Contains(ns.stdin, "kind: Namespace") || !strings.Contains(ns.stdin, "name: op-ns") {
		t.Errorf("first apply must be the operator Namespace:\n%s", ns.stdin)
	}
	if strings.Contains(ns.stdin, "kind: Deployment") || strings.Contains(ns.stdin, "regcred") {
		t.Errorf("the namespace apply must carry nothing else:\n%s", ns.stdin)
	}
	for _, want := range []string{"kind: Secret", "name: regcred", ".dockerconfigjson", "namespace: op-ns"} {
		if !strings.Contains(regcred.stdin, want) {
			t.Errorf("the regcred apply should carry %q:\n%s", want, regcred.stdin)
		}
	}
	for _, want := range []string{"kind: Deployment", "name: pubsubplus-eventbroker-operator", "namespace: op-ns"} {
		if !strings.Contains(bundle.stdin, want) {
			t.Errorf("the bundle apply should carry %q:\n%s", want, bundle.stdin)
		}
	}
	// The bundle half carries no credential of its own -- that is the whole point of
	// the split, and what makes `generate operator` output shareable.
	if strings.Contains(bundle.stdin, ".dockerconfigjson") {
		t.Errorf("the bundle apply must carry no secret document:\n%s", bundle.stdin)
	}
	if strings.Contains(bundle.stdin, "{{") {
		t.Error("applied bundle still contains an unresolved template marker")
	}
}

// TestOperatorApplyNoPullSecret covers the branch where no image-pull secret is
// configured: the regcred apply is skipped entirely, leaving namespace + bundle.
func TestOperatorApplyNoPullSecret(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	cfg.K8s.ImagePullSecret = ""
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.OperatorApply(context.Background()); err != nil {
		t.Fatalf("OperatorApply: %v", err)
	}
	calls := rr.afterPreflights(t, probe{verb: "create", resource: "customresourcedefinitions"})
	if len(calls) != 4 {
		t.Fatalf("OperatorApply without pull secret made %d calls after the probe, want 4 "+
			"(version read + watch-list read + namespace + bundle)", len(calls))
	}
	// calls[0] and calls[1] are the two cluster-wide reads (installed version, installed
	// watch list); the applies start at calls[2].
	applies := calls[2:]
	if !strings.Contains(applies[0].stdin, "kind: Namespace") {
		t.Errorf("the namespace must still be applied first:\n%s", applies[0].stdin)
	}
	for _, call := range applies {
		if strings.Contains(call.stdin, "regcred") {
			t.Errorf("no pull secret configured, so nothing applied may mention regcred:\n%s", call.stdin)
		}
	}
	if !strings.Contains(applies[1].stdin, "kind: Deployment") {
		t.Errorf("the bundle must still carry the controller Deployment:\n%s", applies[1].stdin)
	}
}

// TestOperatorDelete asserts teardown with deleteCRDs=false deletes only the
// non-CRD documents (Deployment, RBAC, ...) on stdin with --ignore-not-found, and
// leaves the CRDs out of the piped manifest entirely -- deleting them would
// cascade-delete every PubSubPlusEventBroker resource in the cluster, including
// brokers this env file has never heard of.
func TestOperatorDelete(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.OperatorDelete(context.Background(), false); err != nil {
		t.Fatalf("OperatorDelete: %v", err)
	}
	calls := rr.afterPreflights(t, probe{verb: "delete", resource: "customresourcedefinitions"})
	if len(calls) != 1 {
		t.Fatalf("OperatorDelete(deleteCRDs=false) made %d call(s) after the probe, want 1 (the non-CRD documents)", len(calls))
	}
	got := calls[0]
	if got.method != "RunInput" || got.name != "kubectl" ||
		!eqArgs(got.args, []string{"delete", "-f", "-", "--ignore-not-found"}) {
		t.Errorf("OperatorDelete\n got: %+v\nwant RunInput kubectl [delete -f - --ignore-not-found]", got)
	}
	if !strings.Contains(got.stdin, "kind: Deployment") {
		t.Error("delete should still pipe the operator Deployment/RBAC documents")
	}
	if strings.Contains(got.stdin, "kind: CustomResourceDefinition") {
		t.Error("deleteCRDs=false must not include any CustomResourceDefinition document in the piped manifest")
	}
}

// TestOperatorDeleteWithCRDs is the deleteCRDs=true half: OperatorDelete first
// checks the cluster for existing broker custom resources (none here -- the
// default recRunner Output returns empty), then deletes the CRD document in a
// SECOND, separate `kubectl delete -f -` call, never folded into the first -- so
// a caller inspecting just the first call (as most teardowns do) cannot mistake a
// CRD-carrying delete for the safe default.
func TestOperatorDeleteWithCRDs(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.OperatorDelete(context.Background(), true); err != nil {
		t.Fatalf("OperatorDelete: %v", err)
	}
	calls := rr.afterPreflights(t, probe{verb: "delete", resource: "customresourcedefinitions"})
	if len(calls) != 3 {
		t.Fatalf("OperatorDelete(deleteCRDs=true) made %d call(s) after the probe, want 3 (non-CRD documents, "+
			"the broker CR listing, then CRDs)", len(calls))
	}
	rest, list, crds := calls[0], calls[1], calls[2]
	for _, call := range []rrCall{rest, crds} {
		if call.method != "RunInput" || call.name != "kubectl" ||
			!eqArgs(call.args, []string{"delete", "-f", "-", "--ignore-not-found"}) {
			t.Errorf("delete call = %+v, want RunInput kubectl [delete -f - --ignore-not-found]", call)
		}
	}
	wantList := []string{"get", "pubsubpluseventbrokers", "--all-namespaces", "-o", "json"}
	if list.method != "Output" || !eqArgs(list.args, wantList) {
		t.Errorf("broker CR listing call = %+v, want Output kubectl %v", list, wantList)
	}
	if !strings.Contains(rest.stdin, "kind: Deployment") || strings.Contains(rest.stdin, "kind: CustomResourceDefinition") {
		t.Errorf("first delete should carry the Deployment/RBAC documents and no CRD:\n%s", rest.stdin)
	}
	if !strings.Contains(crds.stdin, "kind: CustomResourceDefinition") || strings.Contains(crds.stdin, "kind: Deployment") {
		t.Errorf("second delete should carry only the CRD document, on its own:\n%s", crds.stdin)
	}
}

// TestOperatorDeleteRefusesCRDWhenBrokersExist is M3: any PubSubPlusEventBroker
// custom resource still visible anywhere in the cluster must refuse the CRD
// deletion outright -- deleting the CRDs would cascade-delete it along with
// every other broker, including ones this env file never mentioned. The operator
// Deployment/RBAC removal (the first call) must still go through; only the CRD
// layer is refused, and each broker found is named on stderr.
func TestOperatorDeleteRefusesCRDWhenBrokersExist(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{out: []byte(`{"items":[
		{"metadata":{"namespace":"team-a","name":"broker1"}},
		{"metadata":{"namespace":"team-b","name":"broker2"}}
	]}`)}
	buf := &bytes.Buffer{}
	c := NewCluster(rr, cfg, func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }, nil)
	err := c.OperatorDelete(context.Background(), true)
	if err == nil {
		t.Fatal("OperatorDelete must refuse the CRD layer while broker CRs still exist")
	}
	for _, want := range []string{"team-a/broker1", "team-b/broker2"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal error missing %q:\n%v", want, err)
		}
		if !strings.Contains(buf.String(), want) {
			t.Errorf("warning output missing %q:\n%s", want, buf.String())
		}
	}
	calls := rr.afterPreflights(t, probe{verb: "delete", resource: "customresourcedefinitions"})
	if len(calls) != 2 {
		t.Fatalf("OperatorDelete made %d call(s) after the probe, want 2 (the operator bundle delete, then the "+
			"broker CR listing) -- no CRD delete may be issued", len(calls))
	}
	if calls[0].method != "RunInput" || strings.Contains(calls[0].stdin, "kind: CustomResourceDefinition") {
		t.Errorf("the operator Deployment/RBAC removal must still be issued despite the CRD refusal:\n%+v", calls[0])
	}
	for _, call := range calls {
		if strings.Contains(strings.Join(call.args, " "), "CustomResourceDefinition") {
			t.Errorf("no call may touch the CRDs while broker CRs still exist: %+v", call)
		}
	}
}

// TestOperatorDeleteSkipsBrokerCheckWithoutDeleteCRDs covers the safe default:
// deleteCRDs=false must never even ask the cluster whether broker CRs exist --
// there is nothing to refuse when the CRDs are already being kept, and the
// interactive layer question the CLI would otherwise ask is answered already
// (kept), so no listing call belongs on this path either.
func TestOperatorDeleteSkipsBrokerCheckWithoutDeleteCRDs(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	// Broker CRs exist in this fixture; deleteCRDs=false must still never look.
	rr := &recRunner{out: []byte(`{"items":[{"metadata":{"namespace":"team-a","name":"broker1"}}]}`)}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.OperatorDelete(context.Background(), false); err != nil {
		t.Fatalf("OperatorDelete: %v", err)
	}
	calls := rr.afterPreflights(t, probe{verb: "delete", resource: "customresourcedefinitions"})
	if len(calls) != 1 {
		t.Fatalf("OperatorDelete(deleteCRDs=false) made %d call(s) after the probe, want 1 (the operator bundle "+
			"delete only -- no broker CR listing belongs on this path)", len(calls))
	}
	if calls[0].method != "RunInput" || strings.Contains(calls[0].stdin, "kind: CustomResourceDefinition") {
		t.Errorf("the operator Deployment/RBAC removal must still be issued:\n%+v", calls[0])
	}
}

// TestOperatorDeleteRefusesCRDWhenListingFails covers the RBAC/API-error branch:
// being unable to see what the CRD deletion would destroy is not permission to
// destroy it, so the listing failure refuses the CRD layer too, with the cause
// preserved for %w unwrapping.
func TestOperatorDeleteRefusesCRDWhenListingFails(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{outErr: errFake}
	c := NewCluster(rr, cfg, nil, nil)
	err := c.OperatorDelete(context.Background(), true)
	if err == nil {
		t.Fatal("OperatorDelete must refuse the CRD layer when the broker CR listing fails")
	}
	if !errors.Is(err, errFake) {
		t.Errorf("refusal error must preserve the listing failure via %%w: %v", err)
	}
	calls := rr.afterPreflights(t, probe{verb: "delete", resource: "customresourcedefinitions"})
	for _, call := range calls {
		if strings.Contains(strings.Join(call.args, " "), "CustomResourceDefinition") {
			t.Errorf("no call may touch the CRDs when the listing failed: %+v", call)
		}
	}
}

// TestOperatorLogsArgs checks the log passthrough argv against the operator deployment.
func TestOperatorLogsArgs(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.OperatorLogs(context.Background(), "-f", "--tail=10"); err != nil {
		t.Fatalf("OperatorLogs: %v", err)
	}
	got := rr.last()
	want := []string{"logs", "-n", "op-ns", "deployment/pubsubplus-eventbroker-operator", "-f", "--tail=10"}
	if got.method != "Run" || got.name != "kubectl" || !eqArgs(got.args, want) {
		t.Errorf("OperatorLogs\n got: %+v\nwant Run kubectl %v", got, want)
	}
}

// TestOperatorStatus covers a completely untested function (0.0%): the two kubectl
// gets (deployment wide, then controller pods by label) plus the early return when
// the first fails. Same one-liner-wrapper shape the suite already covers for every
// sibling (TestOperatorLogsArgs, TestOperatorDescribe); a typo in the label selector
// or resource name would otherwise ship silently.
func TestOperatorStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := loadK8s(t)
		cfg.K8s.Operator.Namespace = "op-ns"
		rr := &recRunner{}
		c := NewCluster(rr, cfg, nil, nil)
		if err := c.OperatorStatus(context.Background()); err != nil {
			t.Fatalf("OperatorStatus: %v", err)
		}
		if len(rr.calls) != 2 {
			t.Fatalf("OperatorStatus made %d calls, want 2 (deployment + pods)", len(rr.calls))
		}
		wantDeploy := []string{"get", "deployment", "pubsubplus-eventbroker-operator", "-n", "op-ns", "-o", "wide"}
		if got := rr.calls[0]; got.method != "Run" || !eqArgs(got.args, wantDeploy) {
			t.Errorf("deployment get argv = %+v, want Run kubectl %v", got, wantDeploy)
		}
		wantPods := []string{"get", "pods", "-n", "op-ns", "-l", "control-plane=controller-manager", "-o", "wide"}
		if got := rr.calls[1]; got.method != "Run" || !eqArgs(got.args, wantPods) {
			t.Errorf("pods get argv = %+v, want Run kubectl %v", got, wantPods)
		}
	})
	t.Run("stops after the deployment get fails", func(t *testing.T) {
		cfg := loadK8s(t)
		cfg.K8s.Operator.Namespace = "op-ns"
		rr := &recRunner{runErr: errFake}
		c := NewCluster(rr, cfg, nil, nil)
		if err := c.OperatorStatus(context.Background()); err == nil {
			t.Error("OperatorStatus should fail when the deployment get fails")
		}
		if len(rr.calls) != 1 {
			t.Errorf("OperatorStatus should stop after the first failing get; got %d calls", len(rr.calls))
		}
	})
}

// TestOperatorDescribe covers another completely untested function (0.0%);
// identical shape to TestOperatorLogsArgs already in the suite.
func TestOperatorDescribe(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.OperatorDescribe(context.Background()); err != nil {
		t.Fatalf("OperatorDescribe: %v", err)
	}
	got := rr.last()
	want := []string{"describe", "deployment/pubsubplus-eventbroker-operator", "-n", "op-ns"}
	if got.method != "Run" || got.name != "kubectl" || !eqArgs(got.args, want) {
		t.Errorf("OperatorDescribe\n got: %+v\nwant Run kubectl %v", got, want)
	}
}

// TestSplitOperatorBundle pins the column-0 anchoring OperatorDelete's safety
// depends on, plus the edge shapes joinYAMLDocs must hand back as nil rather than
// a lone "---" separator.
func TestSplitOperatorBundle(t *testing.T) {
	t.Run("indented kind inside a nested block is not misclassified", func(t *testing.T) {
		// The ConfigMap embeds an indented "kind: CustomResourceDefinition" line
		// inside a literal block (an example snippet in its data), which must not
		// make this document count as a CRD -- only a document's OWN top-level
		// kind may. The second document is a real, unindented CRD kind, to prove
		// the anchor still catches the case it exists for.
		manifest := []byte(
			"apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cm\n" +
				"data:\n" +
				"  example.yaml: |\n" +
				"    kind: CustomResourceDefinition\n" +
				"    metadata:\n" +
				"      name: widgets.example.com\n" +
				"---\n" +
				"apiVersion: apiextensions.k8s.io/v1\n" +
				"kind: CustomResourceDefinition\n" +
				"metadata:\n  name: real-crd.example.com\n")
		crds, _, rest := splitOperatorBundle(manifest)
		if !strings.Contains(string(rest), "kind: ConfigMap") {
			t.Errorf("the ConfigMap document should land in rest:\n%s", rest)
		}
		if strings.Contains(string(rest), "real-crd.example.com") {
			t.Errorf("the real CRD document must not also land in rest:\n%s", rest)
		}
		if !strings.Contains(string(crds), "real-crd.example.com") {
			t.Errorf("the real (unindented) CRD document should land in crds:\n%s", crds)
		}
		if strings.Contains(string(crds), "example.yaml") {
			t.Errorf("a ConfigMap with an indented kind: CustomResourceDefinition text must not be misclassified as a CRD:\n%s", crds)
		}
	})
	t.Run("empty input", func(t *testing.T) {
		crds, _, rest := splitOperatorBundle([]byte(""))
		if crds != nil || rest != nil {
			t.Errorf("splitOperatorBundle(\"\") = (%q, %q), want (nil, nil)", crds, rest)
		}
	})
	t.Run("no CRD documents", func(t *testing.T) {
		manifest := []byte("kind: ConfigMap\n---\nkind: Secret\n")
		crds, _, rest := splitOperatorBundle(manifest)
		if crds != nil {
			t.Errorf("crds = %q, want nil when the bundle has no CustomResourceDefinition", crds)
		}
		if !strings.Contains(string(rest), "kind: ConfigMap") || !strings.Contains(string(rest), "kind: Secret") {
			t.Errorf("rest should carry both documents:\n%s", rest)
		}
	})
	t.Run("CRD-only input", func(t *testing.T) {
		manifest := []byte("kind: CustomResourceDefinition\nmetadata:\n  name: widgets.example.com\n")
		crds, _, rest := splitOperatorBundle(manifest)
		if rest != nil {
			t.Errorf("rest = %q, want nil when every document is a CRD", rest)
		}
		if !strings.Contains(string(crds), "widgets.example.com") {
			t.Errorf("crds should carry the CRD document:\n%s", crds)
		}
	})
}

// TestOperatorRestart pins the rollout-restart argv against the resolved operator
// namespace, behind the same permission probe every mutating operator command runs.
func TestOperatorRestart(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.OperatorRestart(context.Background()); err != nil {
		t.Fatalf("OperatorRestart: %v", err)
	}
	calls := rr.afterPreflights(t, probe{verb: "patch", resource: "deployments"})
	if len(calls) != 1 {
		t.Fatalf("OperatorRestart made %d call(s) after the probe, want 1", len(calls))
	}
	got := calls[0]
	want := []string{"rollout", "restart", "deployment", operatorDeployment, "-n", "op-ns"}
	if got.method != "Run" || got.name != "kubectl" || !eqArgs(got.args, want) {
		t.Errorf("OperatorRestart\n got: %+v\nwant Run kubectl %v", got, want)
	}
}

// TestOperatorInstalled covers the read-only probe: it reports true only when
// BOTH the CRD and the controller Deployment are visible, and it stops after the
// first failing get rather than probing the second for nothing -- and, per its
// signature, never returns an error: an absent operator and an unreachable
// cluster are both just "false" to the caller.
func TestOperatorInstalled(t *testing.T) {
	t.Run("both gets succeed", func(t *testing.T) {
		cfg := loadK8s(t)
		cfg.K8s.Operator.Namespace = "op-ns"
		rr := &recRunner{}
		c := NewCluster(rr, cfg, nil, nil)
		if !c.OperatorInstalled(context.Background()) {
			t.Error("OperatorInstalled should report true when both gets succeed")
		}
		if len(rr.calls) != 2 {
			t.Fatalf("OperatorInstalled made %d calls, want 2 (crd, then deployment)", len(rr.calls))
		}
		wantCRD := []string{"get", "crd", brokerResource}
		if got := rr.calls[0]; got.method != "Output" || !eqArgs(got.args, wantCRD) {
			t.Errorf("crd get argv = %+v, want Output kubectl %v", got, wantCRD)
		}
		wantDeploy := []string{"get", "deployment", operatorDeployment, "-n", "op-ns"}
		if got := rr.calls[1]; got.method != "Output" || !eqArgs(got.args, wantDeploy) {
			t.Errorf("deployment get argv = %+v, want Output kubectl %v", got, wantDeploy)
		}
	})
	t.Run("CRD missing", func(t *testing.T) {
		cfg := loadK8s(t)
		cfg.K8s.Operator.Namespace = "op-ns"
		rr := &recRunner{outErr: errFake}
		c := NewCluster(rr, cfg, nil, nil)
		if c.OperatorInstalled(context.Background()) {
			t.Error("OperatorInstalled should report false when the crd get fails")
		}
		if len(rr.calls) != 1 {
			t.Errorf("OperatorInstalled should stop after the failing crd get; got %d calls", len(rr.calls))
		}
	})
	t.Run("deployment missing", func(t *testing.T) {
		cfg := loadK8s(t)
		cfg.K8s.Operator.Namespace = "op-ns"
		rr := &recRunner{outErrQueue: []error{nil, errFake}}
		c := NewCluster(rr, cfg, nil, nil)
		if c.OperatorInstalled(context.Background()) {
			t.Error("OperatorInstalled should report false when the deployment get fails")
		}
		if len(rr.calls) != 2 {
			t.Errorf("OperatorInstalled should still probe the deployment after a successful crd get; got %d calls", len(rr.calls))
		}
	})
}

func mustContain(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("rendered output missing %q", needle)
	}
}

// TestOperatorNamespaceIsNotInTheDeleteStream is the safety property this split
// exists for. The Namespace document used to sit in `rest`, so `remove operator`
// deleted it in the same `delete -f -` stream as the Deployment -- cascading to
// everything else in that namespace, with none of the occupancy checks
// `remove namespace` applies to the broker's own namespace.
func TestOperatorNamespaceIsNotInTheDeleteStream(t *testing.T) {
	manifest := []byte("kind: Namespace\nmetadata:\n  name: solace-operator\n" +
		"---\nkind: Deployment\nmetadata:\n  name: pubsubplus-eventbroker-operator\n" +
		"---\nkind: CustomResourceDefinition\nmetadata:\n  name: brokers.example.com\n")

	crds, ns, rest := splitOperatorBundle(manifest)
	if !strings.Contains(string(ns), "kind: Namespace") {
		t.Errorf("the Namespace document must be split out on its own:\n%s", ns)
	}
	if strings.Contains(string(rest), "kind: Namespace") {
		t.Errorf("the Namespace must NOT ride the delete stream: deleting it cascades to everything in it:\n%s", rest)
	}
	if !strings.Contains(string(rest), "kind: Deployment") {
		t.Errorf("the Deployment belongs in rest:\n%s", rest)
	}
	if !strings.Contains(string(crds), "brokers.example.com") {
		t.Errorf("the CRD stays its own layer:\n%s", crds)
	}
}

// TestOperatorDeleteKeepsTheNamespaceAndRemovesTheRegcred covers the consequence of
// keeping the namespace: the image-pull Secret used to be reaped as namespace content,
// so it now needs deleting by name or a registry credential outlives every teardown.
func TestOperatorDeleteKeepsTheNamespaceAndRemovesTheRegcred(t *testing.T) {
	cfg := loadK8s(t) // the sample sets an image-pull secret, so a regcred was applied
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)

	if err := c.OperatorDelete(context.Background(), false); err != nil {
		t.Fatalf("OperatorDelete: %v", err)
	}

	var sawRegcred, sawNamespaceDelete bool
	for _, call := range rr.calls {
		if strings.Contains(strings.Join(call.args, " "), "delete namespace") {
			sawNamespaceDelete = true
		}
		// The namespace must not reach kubectl on stdin either, which is how it used
		// to be deleted -- as one document among the Deployment and RBAC.
		if strings.Contains(call.stdin, "kind: Namespace") {
			t.Errorf("a Namespace document reached `delete -f -`:\n%s", call.stdin)
		}
		if strings.Contains(call.stdin, "name: "+operatorRegcredName) {
			sawRegcred = true
		}
	}
	if !sawRegcred {
		t.Error("the operator's image-pull secret must be in the delete stream: the namespace that used to " +
			"take it with it is now kept, so without this a registry credential outlives every teardown")
	}
	if sawNamespaceDelete {
		t.Error("`remove operator` must not delete the namespace: it may hold objects this tool never created")
	}
}

// TestOperatorProbesEveryKindItTouches is the fix for a check that passed and then
// failed halfway through the work. Probing only the CRD meant an identity allowed to
// create custom resource definitions but not ClusterRoleBindings got past the
// preflight and died mid-apply -- the exact state Preflight exists to prevent.
func TestOperatorProbesEveryKindItTouches(t *testing.T) {
	for _, tc := range []struct {
		name string
		verb string
		run  func(*Cluster) error
	}{
		{"apply", "create", func(c *Cluster) error { return c.OperatorApply(context.Background()) }},
		{"delete", "delete", func(c *Cluster) error { return c.OperatorDelete(context.Background(), false) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := loadK8s(t) // pull secret set -> the regcred is in play
			cfg.K8s.Operator.Namespace = "op-ns"
			rr := &recRunner{}
			c := NewCluster(rr, cfg, nil, nil)
			if err := tc.run(c); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			want := []probe{
				{verb: tc.verb, resource: "customresourcedefinitions"},
				{verb: tc.verb, resource: "clusterroles"},
				{verb: tc.verb, resource: "clusterrolebindings"},
				{verb: tc.verb, resource: "serviceaccounts"},
				{verb: tc.verb, resource: "roles"},
				{verb: tc.verb, resource: "rolebindings"},
				{verb: tc.verb, resource: "deployments"},
				{verb: tc.verb, resource: "secrets"},
			}
			rr.afterPreflights(t, want...)

			// The namespaced kinds must ask about the OPERATOR's namespace, not the
			// broker's. A probe aimed at the wrong namespace still answers -- it just
			// answers a question nobody asked.
			for _, res := range []string{"serviceaccounts", "roles", "rolebindings", "deployments", "secrets"} {
				ns, found := rr.probedNamespace(tc.verb, res)
				if !found {
					continue // reported by afterPreflights above
				}
				if ns != "op-ns" {
					t.Errorf("%s %s probed namespace %q, want the operator's own (op-ns)", tc.verb, res, ns)
				}
			}
			// Cluster-scoped kinds must carry no namespace at all.
			for _, res := range []string{"customresourcedefinitions", "clusterroles", "clusterrolebindings"} {
				if ns, found := rr.probedNamespace(tc.verb, res); found && ns != "" {
					t.Errorf("%s %s is cluster-scoped but was probed with -n %q", tc.verb, res, ns)
				}
			}
		})
	}
}

// TestOperatorRestartProbesTheOperatorNamespace pins a silent defect: the restart
// happens in the operator's namespace while its permission check asked about the
// broker's, so an identity permitted in one and not the other passed and then failed.
func TestOperatorRestartProbesTheOperatorNamespace(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.OperatorRestart(context.Background()); err != nil {
		t.Fatalf("OperatorRestart: %v", err)
	}
	ns, found := rr.probedNamespace("patch", "deployments")
	if !found {
		t.Fatal("OperatorRestart must probe `patch deployments`")
	}
	if ns != "op-ns" {
		t.Errorf("probed namespace = %q, want op-ns -- the restart acts there, so asking about the broker's "+
			"namespace answers the wrong question", ns)
	}
}

// TestGenOperatorSecrets is gone: `generate secrets operator` is gone.
//
// GenOperator now emits the whole install stream in APPLY ORDER -- the namespace
// document, the image-pull Secret when one is configured, then the bundle -- so the
// regcred is tested as part of that stream rather than on its own. Splitting it out
// meant the two halves had to be applied in the right order by hand, and the ordering
// is the part that is easy to get wrong: the Secret is namespaced and its namespace
// only exists inside the bundle.

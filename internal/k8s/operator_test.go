package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"solace/internal/config"
	"solace/internal/engine"
	"solace/internal/examples"
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
		// The gate moved to the CREDENTIALS (config.Config.ManagesImagePullSecret):
		// clearing the name alone no longer suppresses the block, since the sample
		// fixture also supplies image.user/image.pass. Both are cleared here so this
		// subtest still exercises the "nothing configured at all" case its name
		// promises; TestRenderOperatorPullSecretGatesOnCredentials covers the state
		// this used to test by itself -- a name with no credentials behind it.
		cfg.K8s.ImagePullSecret = ""
		cfg.Image.User, cfg.Image.Pass = "", ""
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

// TestRenderOperatorPullSecretGatesOnCredentials pins the state that had no coverage
// before ManagesImagePullSecret existed: kubernetes.imagePullSecret NAMES a Secret, but
// with no image.user/image.pass behind it that is the operator's bring-your-own case
// (config.Config.ManagesImagePullSecret's own doc comment) -- the broker CR can still
// reference such a Secret (render.BrokerCR), but the operator bundle can only ever
// reference its OWN fixed-name regcred, which there are no credentials here to build, so
// the block must stay omitted exactly as if no name had been configured at all.
func TestRenderOperatorPullSecretGatesOnCredentials(t *testing.T) {
	cfg := loadK8s(t) // sample: kubernetes.imagePullSecret AND image.user/image.pass set
	cfg.Image.User, cfg.Image.Pass = "", ""
	out, err := RenderOperator(cfg, "op-ns", watchNamespace(cfg))
	if err != nil {
		t.Fatalf("RenderOperator: %v", err)
	}
	s := string(out)
	if strings.Contains(s, "imagePullSecrets:") || strings.Contains(s, "- name: regcred") {
		t.Errorf("kubernetes.imagePullSecret is set but no credentials are configured, so the operator "+
			"bundle must carry no pull-secret reference:\n%s", s)
	}
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

// TestGenOperatorSplicesRegcredOnCredentialsNotOnName pins the gate GenOperator shares
// with RenderOperator and OperatorApply: whether the operator's own regcred Secret is
// spliced into the stream depends on image.user/image.pass, never on whether
// kubernetes.imagePullSecret happens to be set -- that name belongs to the BROKER's own
// pull secret (DockerRegistrySecret) and has no bearing on the operator's fixed-name
// "regcred" artifact.
func TestGenOperatorSplicesRegcredOnCredentialsNotOnName(t *testing.T) {
	t.Run("credentials present, no broker pull-secret name configured", func(t *testing.T) {
		cfg := loadK8s(t)
		cfg.K8s.ImagePullSecret = ""
		out, err := GenOperator(cfg)
		if err != nil {
			t.Fatalf("GenOperator: %v", err)
		}
		s := string(out)
		for _, want := range []string{"kind: Secret", "name: regcred", ".dockerconfigjson"} {
			if !strings.Contains(s, want) {
				t.Errorf("credentials are configured, so the regcred must be spliced in even with no broker "+
					"pull-secret name set (%q missing):\n%s", want, s)
			}
		}
	})
	t.Run("named broker pull secret, no credentials", func(t *testing.T) {
		cfg := loadK8s(t) // sample: kubernetes.imagePullSecret = solace-image-pull
		cfg.Image.User, cfg.Image.Pass = "", ""
		out, err := GenOperator(cfg)
		if err != nil {
			t.Fatalf("GenOperator: %v", err)
		}
		s := string(out)
		if strings.Contains(s, "regcred") || strings.Contains(s, ".dockerconfigjson") {
			t.Errorf("kubernetes.imagePullSecret names a Secret the operator did not build, so GenOperator "+
				"has no credentials to build ITS OWN regcred from:\n%s", s)
		}
	})
}

// TestOperatorApply asserts the apply sequence: ONE read of the installed operator, then
// the bundle's own Namespace document (its namespace read and found absent -- the stage
// TestOperatorApplyPlansItsNamespace pins), then the regcred into that namespace, then the
// rest of the bundle -- each on stdin via `apply -f -`. The split exists so nothing
// applied is absent from a `generate` output; the ordering is what makes a first install
// work at all.
func TestOperatorApply(t *testing.T) {
	cfg := loadK8s(t) // pull secret set -> regcred applied
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.OperatorApply(context.Background()); err != nil {
		t.Fatalf("OperatorApply: %v", err)
	}
	calls := withoutNamespaceStage(rr.afterPreflights(t, probe{verb: "create", resource: "customresourcedefinitions"}), "op-ns")
	if len(calls) != 4 {
		t.Fatalf("OperatorApply made %d calls after the probe, want 4 "+
			"(operator read + namespace + regcred + bundle)", len(calls))
	}
	// The one read is the cluster-wide `get deployment`, and both the downgrade check
	// (installed version) and the watch reconcile (installed WATCH_NAMESPACE, so the
	// applied value can be the union rather than this env file's list alone) use it.
	// Each used to issue its own.
	if read := calls[0]; read.method != "Output" || !eqArgs(read.args,
		[]string{"get", "deployment", "--all-namespaces", "-o", "json"}) {
		t.Errorf("read = %+v, want the cluster-wide deployment read", read)
	}
	ns, regcred, bundle := calls[1], calls[2], calls[3]
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

// TestOperatorApplyNoCredentialsSkipsRegcred covers the branch where no registry
// credentials are configured: the regcred apply is skipped entirely, leaving namespace +
// bundle. This used to clear kubernetes.imagePullSecret (the name) to reach that branch;
// the gate moved to the CREDENTIALS (config.Config.ManagesImagePullSecret), so clearing
// the name alone no longer suppresses anything -- see
// TestOperatorApplyNamedPullSecretWithoutCredentials for the state that replaces it.
func TestOperatorApplyNoCredentialsSkipsRegcred(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	cfg.Image.User, cfg.Image.Pass = "", ""
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.OperatorApply(context.Background()); err != nil {
		t.Fatalf("OperatorApply: %v", err)
	}
	calls := withoutNamespaceStage(rr.afterPreflights(t, probe{verb: "create", resource: "customresourcedefinitions"}), "op-ns")
	if len(calls) != 3 {
		t.Fatalf("OperatorApply without credentials made %d calls after the probe, want 3 "+
			"(operator read + namespace + bundle)", len(calls))
	}
	// calls[0] is the one cluster-wide read of the installed operator; the applies start
	// at calls[1].
	applies := calls[1:]
	if !strings.Contains(applies[0].stdin, "kind: Namespace") {
		t.Errorf("the namespace must still be applied first:\n%s", applies[0].stdin)
	}
	for _, call := range applies {
		if strings.Contains(call.stdin, "regcred") {
			t.Errorf("no credentials configured, so nothing applied may mention regcred:\n%s", call.stdin)
		}
	}
	if !strings.Contains(applies[1].stdin, "kind: Deployment") {
		t.Errorf("the bundle must still carry the controller Deployment:\n%s", applies[1].stdin)
	}
	// The preflight itself must not ask about a permission this run never needs --
	// operatorProbes only appends the "secrets" probe when withRegcred is true.
	for _, call := range rr.calls {
		if len(call.args) >= 4 && call.args[0] == "auth" && call.args[1] == "can-i" && call.args[3] == "secrets" {
			t.Errorf("no credentials configured, so the regcred permission must never be probed: %+v", call)
		}
	}
}

// TestOperatorApplyNamedPullSecretWithoutCredentials is the state that had no coverage
// before the credentials gate existed: kubernetes.imagePullSecret names a Secret this env
// file supplies no image.user/image.pass for. render.BrokerCR still references it by name
// (the broker CR side of the bring-your-own case), but the operator bundle can only ever
// reference its OWN fixed-name regcred -- and there are no credentials here to build one
// from, so OperatorApply must build, probe for and apply nothing extra, identical to no
// name being configured at all.
func TestOperatorApplyNamedPullSecretWithoutCredentials(t *testing.T) {
	cfg := loadK8s(t) // sample: kubernetes.imagePullSecret = solace-image-pull
	cfg.K8s.Operator.Namespace = "op-ns"
	cfg.Image.User, cfg.Image.Pass = "", ""
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.OperatorApply(context.Background()); err != nil {
		t.Fatalf("OperatorApply: %v", err)
	}
	calls := withoutNamespaceStage(rr.afterPreflights(t, probe{verb: "create", resource: "customresourcedefinitions"}), "op-ns")
	if len(calls) != 3 {
		t.Fatalf("OperatorApply with a named-but-unbuilt pull secret made %d calls after the probe, want 3 "+
			"(operator read + namespace + bundle)", len(calls))
	}
	for _, call := range calls[1:] {
		if strings.Contains(call.stdin, "kind: Secret") || strings.Contains(call.stdin, ".dockerconfigjson") {
			t.Errorf("a named pull secret with no credentials is the operator's bring-your-own case: "+
				"nothing here may build or apply a Secret for it:\n%s", call.stdin)
		}
		if strings.Contains(call.stdin, "imagePullSecrets:") {
			t.Errorf("the operator bundle can only reference its OWN regcred, so it must carry no "+
				"imagePullSecrets block when there are no credentials to build one from:\n%s", call.stdin)
		}
	}
}

// TestOperatorApplyProceedsWhenTheClusterWideListIsDenied is the reported failure: an
// identity that may not list Deployments cluster-wide got kubectl's Forbidden, and the
// deploy stopped there. It now reads the operator namespace by name instead and carries
// on -- asking the escalation questions on a first install, not when the operator is
// already there -- reading once and warning once either way.
func TestOperatorApplyProceedsWhenTheClusterWideListIsDenied(t *testing.T) {
	const role = "clusterroles/" + operatorClusterRole
	for _, tc := range []struct {
		name      string
		named     []byte // the by-name read's reply; nil = nothing there
		wantProbe bool   // escalate and bind asked
		wantWatch string // what the applied bundle's WATCH_NAMESPACE must carry
	}{
		{"first install: nothing in the operator namespace", nil, true, ""},
		{"re-deploy: the operator is in its namespace",
			namedOperatorJSON("op-ns", "solace/pubsubplus-eventbroker-operator:1.4.2", "team-a"), false, `"team-a,`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := loadK8s(t) // pull secret set -> regcred applied
			cfg.K8s.Operator.Namespace = "op-ns"
			rr := deniedRunner(tc.named)
			buf := &bytes.Buffer{}
			c := NewCluster(rr, cfg, func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }, buf)
			if err := c.OperatorApply(context.Background()); err != nil {
				t.Fatalf("OperatorApply: %v\n%s", err, buf.String())
			}
			calls := withoutNamespaceStage(rr.afterPreflights(t, probe{verb: "create", resource: "customresourcedefinitions"}), "op-ns")
			want := [][]string{
				{"get", "deployment", "--all-namespaces", "-o", "json"},
				{"get", "deployment", operatorDeployment, "-n", "op-ns", "--ignore-not-found", "-o", "json"},
			}
			if tc.wantProbe {
				// escalate asks about the kind: a create is authorized before the new role's
				// name is known, so a grant limited to one name cannot cover it. bind is
				// checked against the role bound, so it names it.
				want = append(want,
					[]string{"auth", "can-i", "escalate", "clusterroles", "--all-namespaces"},
					[]string{"auth", "can-i", "bind", role, "--all-namespaces"})
			}
			if len(calls) != len(want)+3 {
				t.Fatalf("made %d calls after the probe, want %d (%d reads and probes + namespace + regcred + "+
					"bundle):\n%+v", len(calls), len(want)+3, len(want), calls)
			}
			for i, w := range want {
				if calls[i].method != "Output" || !eqArgs(calls[i].args, w) {
					t.Errorf("call %d = %+v, want Output %v", i, calls[i], w)
				}
			}
			for _, call := range calls[len(want):] {
				if call.method != "RunInput" || !eqArgs(call.args, []string{"apply", "-f", "-"}) {
					t.Errorf("after the reads, want only the three applies, got %+v", call)
				}
			}
			if bundle := calls[len(calls)-1].stdin; tc.wantWatch != "" && !strings.Contains(bundle, tc.wantWatch) {
				t.Errorf("the bundle must carry the union with the watch list the fallback read (%s):\n%s",
					tc.wantWatch, bundle)
			}
			if n := strings.Count(buf.String(), "could not list Deployments cluster-wide"); n != 1 {
				t.Errorf("the fallback warning printed %d times, want exactly 1:\n%s", n, buf.String())
			}
		})
	}
}

// TestOperatorApplyRefusesAFallbackFirstInstallWithoutEscalateOrBind is the half-install the
// escalation probes exist to prevent. An identity that cannot list Deployments cluster-wide
// does not hold every permission the operator's ClusterRole grants, so creating that role
// needs `escalate` and binding it needs `bind` -- neither of which the `create` probes can
// see. Without them the bundle would fail mid-apply, after the CRDs and the controller
// Deployment had landed; instead nothing is applied and the error says why.
func TestOperatorApplyRefusesAFallbackFirstInstallWithoutEscalateOrBind(t *testing.T) {
	for _, tc := range []struct{ verb, target string }{
		{"escalate", "clusterroles"},
		{"bind", "clusterroles/" + operatorClusterRole},
	} {
		t.Run(tc.verb, func(t *testing.T) {
			cfg := loadK8s(t)
			cfg.K8s.Operator.Namespace = "op-ns"
			rr := deniedRunner(nil)
			rr.canIByVerb = map[string]string{tc.verb: "no"}
			c := NewCluster(rr, cfg, nil, nil)
			asked := false
			c.Confirm = func(string) bool { asked = true; return true }
			err := c.OperatorApply(context.Background())
			if err == nil {
				t.Fatal("a fallback first install without " + tc.verb + " must be refused before anything is applied")
			}
			for _, want := range []string{"not allowed to " + tc.verb + " " + tc.target + " in the cluster",
				operatorClusterRole, "cannot list Deployments cluster-wide", "Nothing was applied",
				"kubernetes.operator.namespace"} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error = %v\nwant it to contain %q", err, want)
				}
			}
			for _, call := range rr.calls {
				if call.method == "RunInput" {
					t.Errorf("refused, yet something was applied: %+v", call)
				}
			}
			if asked {
				t.Error("a refused deploy must not reach any question")
			}
		})
	}
}

// TestOperatorApplyPreviewReadsTheOperatorOnce: under Echo the read answers nothing and
// fails nothing, so a preview shows the one cluster-wide read where it always stood, and
// neither the by-name fallback read of the operator Deployment nor the escalation probes --
// both belong to a refused list, which a preview cannot have.
func TestOperatorApplyPreviewReadsTheOperatorOnce(t *testing.T) {
	var out strings.Builder
	c := NewCluster(engine.Echo{W: &out}, loadK8s(t), nil, &out)
	if err := c.OperatorApply(context.Background()); err != nil {
		t.Fatalf("OperatorApply under Echo: %v", err)
	}
	got := out.String()
	if n := strings.Count(got, "get deployment --all-namespaces -o json"); n != 1 {
		t.Errorf("the preview showed the cluster-wide read %d times, want exactly 1:\n%s", n, got)
	}
	for _, never := range []string{"get deployment " + operatorDeployment, "can-i escalate", "can-i bind"} {
		if strings.Contains(got, never) {
			t.Errorf("a preview has no refused list, so it must not show %q:\n%s", never, got)
		}
	}
}

// TestOperatorClusterRoleNameMatchesTheBundle pins the name the `bind` escalation probe asks
// about to the role the bundle creates and binds. A renamed role would leave the probe
// asking about an object nothing applies -- answering for the wrong grant.
func TestOperatorClusterRoleNameMatchesTheBundle(t *testing.T) {
	manifest, err := RenderOperator(loadK8s(t), "op-ns", "solace")
	if err != nil {
		t.Fatalf("RenderOperator: %v", err)
	}
	for _, want := range []string{
		"kind: ClusterRole\nmetadata:\n  name: " + operatorClusterRole + "\n",
		"roleRef:\n  apiGroup: rbac.authorization.k8s.io\n  kind: ClusterRole\n  name: " + operatorClusterRole + "\n",
	} {
		if !strings.Contains(string(manifest), want) {
			t.Errorf("the rendered bundle has no %q -- operatorClusterRole has drifted from the template", want)
		}
	}
}

// TestOperatorDeleteOmitsRegcredWithoutCredentials is the delete-side counterpart to
// TestOperatorApplyNoCredentialsSkipsRegcred and TestOperatorApplyNamedPullSecretWithoutCredentials.
// Unlike apply, the regcred rides the SAME delete call as the rest of the bundle rather
// than one of its own (OperatorDelete's own comment explains why: it is folded into `rest`
// so the teardown stays one ordered `delete -f -`), so there is no separate call to assert
// absent -- only that the one call never mentions it, whether or not a name is configured.
func TestOperatorDeleteOmitsRegcredWithoutCredentials(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*config.Config)
	}{
		{"no name, no credentials", func(c *config.Config) {
			c.K8s.ImagePullSecret = ""
			c.Image.User, c.Image.Pass = "", ""
		}},
		{"name set, no credentials (bring-your-own)", func(c *config.Config) {
			c.Image.User, c.Image.Pass = "", ""
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := loadK8s(t)
			cfg.K8s.Operator.Namespace = "op-ns"
			tc.mutate(cfg)
			rr := &recRunner{}
			c := NewCluster(rr, cfg, nil, nil)
			if err := c.OperatorDelete(context.Background(), false); err != nil {
				t.Fatalf("OperatorDelete: %v", err)
			}
			calls := rr.afterPreflights(t, probe{verb: "delete", resource: "clusterroles"})
			if len(calls) != 1 {
				t.Fatalf("OperatorDelete(deleteCRDs=false) made %d call(s) after the probe, want 1 "+
					"(the non-CRD documents)", len(calls))
			}
			if strings.Contains(calls[0].stdin, "regcred") {
				t.Errorf("no credentials configured, so the delete stream must not mention regcred:\n%s", calls[0].stdin)
			}
			for _, call := range rr.calls {
				if len(call.args) >= 4 && call.args[0] == "auth" && call.args[1] == "can-i" && call.args[3] == "secrets" {
					t.Errorf("no credentials configured, so the regcred permission must never be probed: %+v", call)
				}
			}
		})
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
	calls := rr.afterPreflights(t, probe{verb: "delete", resource: "clusterroles"})
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

// TestOperatorDeleteWithCRDs is the deleteCRDs=true half: OperatorDelete probes the
// cluster-wide broker list with every delete, checks the cluster for existing broker
// custom resources BEFORE its first delete (none here -- the default recRunner Output
// returns empty), checks again right before the CRDs go, and deletes the CRD document in
// a SECOND, separate `kubectl delete -f -` call, never folded into the first -- so a
// caller inspecting just the first delete (as most teardowns do) cannot mistake a
// CRD-carrying delete for the safe default.
func TestOperatorDeleteWithCRDs(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.OperatorDelete(context.Background(), true); err != nil {
		t.Fatalf("OperatorDelete: %v", err)
	}
	calls := rr.afterPreflights(t, probe{verb: "delete", resource: "customresourcedefinitions"},
		probe{verb: "list", resource: brokerResource})
	if len(calls) != 4 {
		t.Fatalf("OperatorDelete(deleteCRDs=true) made %d call(s) after the probe, want 4 (the broker CR listing, "+
			"non-CRD documents, the listing again, then CRDs)", len(calls))
	}
	first, rest, list, crds := calls[0], calls[1], calls[2], calls[3]
	for _, call := range []rrCall{rest, crds} {
		if call.method != "RunInput" || call.name != "kubectl" ||
			!eqArgs(call.args, []string{"delete", "-f", "-", "--ignore-not-found"}) {
			t.Errorf("delete call = %+v, want RunInput kubectl [delete -f - --ignore-not-found]", call)
		}
	}
	wantList := []string{"get", "pubsubpluseventbrokers", "--all-namespaces", "-o", "json"}
	for _, l := range []rrCall{first, list} {
		if l.method != "Output" || !eqArgs(l.args, wantList) {
			t.Errorf("broker CR listing call = %+v, want Output kubectl %v", l, wantList)
		}
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
// every other broker, including ones this env file never mentioned. --delete-crd
// was asked for, so the refusal comes BEFORE anything is deleted: the operator stays
// installed, each broker found is named on stderr, and the error says nothing was
// deleted -- rather than the Deployment and RBAC going first and the request then
// failing halfway.
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
	if !strings.Contains(err.Error(), "Nothing was deleted: the operator is still installed") {
		t.Errorf("error = %v, want it to say nothing was deleted", err)
	}
	calls := rr.afterPreflights(t, probe{verb: "delete", resource: "customresourcedefinitions"})
	if len(calls) != 1 || calls[0].method != "Output" {
		t.Fatalf("OperatorDelete made %+v after the probes, want only the broker CR listing -- nothing may be "+
			"deleted once the CRD layer is refused", calls)
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
	calls := rr.afterPreflights(t, probe{verb: "delete", resource: "clusterroles"})
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
// preserved for %w unwrapping -- and, --delete-crd having been asked for, before
// anything at all is deleted.
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
		if call.method == "RunInput" {
			t.Errorf("nothing may be deleted when the listing failed: %+v", call)
		}
	}
}

// TestOperatorDeleteProbesCRDsOnlyWhenDeletingThem: a removal that keeps the CRDs demands
// neither the grant to delete them nor the cluster-wide broker list that guards their
// deletion, so an identity holding the operator's own RBAC and Deployment rights but not
// those can still remove the operator. With --delete-crd both are asked, up front.
func TestOperatorDeleteProbesCRDsOnlyWhenDeletingThem(t *testing.T) {
	for _, deleteCRDs := range []bool{false, true} {
		cfg := loadK8s(t)
		cfg.K8s.Operator.Namespace = "op-ns"
		rr := &recRunner{}
		if err := NewCluster(rr, cfg, nil, nil).OperatorDelete(context.Background(), deleteCRDs); err != nil {
			t.Fatalf("OperatorDelete(%v): %v", deleteCRDs, err)
		}
		_, crdProbe := rr.probedNamespace("delete", "customresourcedefinitions")
		_, listProbe := rr.probedNamespace("list", brokerResource)
		if crdProbe != deleteCRDs || listProbe != deleteCRDs {
			t.Errorf("deleteCRDs=%v: probed the CRD delete = %v and the cluster-wide broker list = %v, want both %v",
				deleteCRDs, crdProbe, listProbe, deleteCRDs)
		}
	}
}

// TestOperatorDeleteRefusesAnUnreadableBrokerListBeforeAnyDelete: --delete-crd needs the
// cluster-wide broker list to prove no broker would be cascade-deleted, so an identity that
// may not read it is refused by the preflight -- naming the cluster role binding to ask for
// -- with nothing read and nothing deleted, the operator still installed.
func TestOperatorDeleteRefusesAnUnreadableBrokerListBeforeAnyDelete(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{canIByVerb: map[string]string{"list": "no"}}
	err := NewCluster(rr, cfg, nil, nil).OperatorDelete(context.Background(), true)
	for _, want := range []string{"not allowed to list " + brokerResource + " in the cluster", "cluster role binding"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("OperatorDelete err = %v, want it to contain %q", err, want)
		}
	}
	for _, call := range rr.calls {
		if !isCanI(call.args) {
			t.Errorf("nothing but probes may run before the refusal: %+v", call)
		}
	}
}

// TestOperatorApplyPlansItsNamespace: the bundle's Namespace document is applied only to a
// namespace that is not there. Existing, it is neither probed for nor applied -- applying
// over it would PATCH a namespace this tool may not have created. Absent, `create
// namespaces` is asked and the document applied. Unreadable, the create probe decides, and a
// refusal is an ERROR with nothing applied: the bundle's cluster-scoped documents would
// otherwise land for a controller whose namespace may not exist.
func TestOperatorApplyPlansItsNamespace(t *testing.T) {
	read := []string{"get", "namespace", "op-ns", "--ignore-not-found", "-o", "name"}
	for _, tc := range []struct {
		name        string
		nsOut       []byte
		nsErr       error
		createNS    string // the `create namespaces` answer; "" is yes
		wantErr     string
		wantProbe   bool
		wantNSApply bool
	}{
		{"exists: nothing asked, nothing applied", []byte("namespace/op-ns\n"), nil, "no", "", false, false},
		{"absent: asked, then applied", nil, nil, "", "", true, true},
		{"absent and not creatable", nil, nil, "no", "does not exist", true, false},
		{"unreadable, creatable", nil, errFake, "", "", true, true},
		{"unreadable and not creatable", nil, errFake, "no", "could not confirm the operator namespace", true, false},
		// No readable answer is neither: the probe's own error, nothing applied.
		{"absent, probe unanswered", nil, nil, "maybe", "could not read the answer", true, false},
		{"unreadable, probe unanswered", nil, errFake, "maybe", "could not read the answer", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := loadK8s(t)
			cfg.K8s.Operator.Namespace = "op-ns"
			rr := &recRunner{
				outQueue:       [][]byte{[]byte(`{"items":[]}`), tc.nsOut},
				outErrQueue:    []error{nil, tc.nsErr},
				canIByResource: map[string]string{"namespaces": tc.createNS},
			}
			err := NewCluster(rr, cfg, nil, nil).OperatorApply(context.Background())
			if tc.wantErr == "" && err != nil {
				t.Fatalf("OperatorApply: %v", err)
			}
			if tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)) {
				t.Fatalf("OperatorApply err = %v, want it to contain %q", err, tc.wantErr)
			}
			readAt, probeAt, applies, nsApplied := -1, -1, 0, false
			for i, call := range rr.calls {
				switch {
				case eqArgs(call.args, read):
					readAt = i
				case isCanI(call.args) && canIResource(call.args) == "namespaces":
					probeAt = i
				case call.method == "RunInput":
					applies++
					nsApplied = nsApplied || strings.Contains(call.stdin, "kind: Namespace")
				}
			}
			if readAt < 0 {
				t.Fatalf("the operator namespace must be read by name first:\n%+v", rr.calls)
			}
			if (probeAt >= 0) != tc.wantProbe {
				t.Errorf("asked create namespaces = %v, want %v", probeAt >= 0, tc.wantProbe)
			}
			if probeAt >= 0 && probeAt < readAt {
				t.Error("the namespace must be read before anyone asks to create it")
			}
			if nsApplied != tc.wantNSApply {
				t.Errorf("Namespace document applied = %v, want %v", nsApplied, tc.wantNSApply)
			}
			if tc.wantErr != "" && applies != 0 {
				t.Errorf("refused, yet %d document(s) were applied", applies)
			}
		})
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
		{"delete", "delete", func(c *Cluster) error { return c.OperatorDelete(context.Background(), true) }},
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

// TestFullExampleListsTheOperatorsDefaultPorts keeps the commented kubernetes.ports list
// in the annotated example equal to the bundled operator's own default. The CLI carries
// no copy of that list any more -- an unset kubernetes.ports leaves spec.service.ports to
// the CRD -- so the example is where an operator reads which ports the default exposes,
// and a bundle bump must not leave it describing the previous operator's list.
func TestFullExampleListsTheOperatorsDefaultPorts(t *testing.T) {
	want := bundledDefaultPorts(t)
	if len(want) == 0 {
		t.Fatal("the bundled CRD carries no spec.service.ports default; the example's commented list " +
			"and the docs saying the operator's default applies are both wrong now")
	}
	got := commentedExamplePorts(t)
	if !slices.Equal(got, want) {
		t.Errorf("full.yaml's commented kubernetes.ports does not match the bundled operator's default\n"+
			" example: %v\n     CRD: %v", got, want)
	}
}

// bundledDefaultPorts reads spec.service.ports' default out of the CRD document in
// operatorBundle, rendered as kubernetes.ports entries. The CRD carries no template
// action, so its document is parsed as written.
func bundledDefaultPorts(t *testing.T) []string {
	t.Helper()
	for _, doc := range strings.Split(operatorBundle, "\n---\n") {
		if !strings.Contains(doc, "kind: CustomResourceDefinition") ||
			!strings.Contains(doc, "name: pubsubpluseventbrokers.pubsubplus.solace.com") {
			continue
		}
		var crd struct {
			Spec struct {
				Versions []struct {
					Storage bool `yaml:"storage"`
					Schema  struct {
						OpenAPIV3Schema struct {
							Properties struct {
								Spec struct {
									Properties struct {
										Service struct {
											Properties struct {
												Ports struct {
													Default []struct {
														ContainerPort int    `yaml:"containerPort"`
														Name          string `yaml:"name"`
														Protocol      string `yaml:"protocol"`
														ServicePort   int    `yaml:"servicePort"`
													} `yaml:"default"`
												} `yaml:"ports"`
											} `yaml:"properties"`
										} `yaml:"service"`
									} `yaml:"properties"`
								} `yaml:"spec"`
							} `yaml:"properties"`
						} `yaml:"openAPIV3Schema"`
					} `yaml:"schema"`
				} `yaml:"versions"`
			} `yaml:"spec"`
		}
		if err := yaml.Unmarshal([]byte(doc), &crd); err != nil {
			t.Fatalf("parse the bundled PubSubPlusEventBroker CRD: %v", err)
		}
		for _, v := range crd.Spec.Versions {
			if !v.Storage {
				continue
			}
			var out []string
			for _, p := range v.Schema.OpenAPIV3Schema.Properties.Spec.Properties.Service.Properties.Ports.Default {
				e := fmt.Sprintf("%s=%d", p.Name, p.ContainerPort)
				if p.ServicePort != p.ContainerPort {
					e += fmt.Sprintf(":%d", p.ServicePort)
				}
				if p.Protocol != "" && p.Protocol != "TCP" {
					e += "/" + p.Protocol
				}
				out = append(out, e)
			}
			return out
		}
		t.Fatal("the bundled PubSubPlusEventBroker CRD has no storage version")
	}
	t.Fatal("operatorBundle holds no PubSubPlusEventBroker CRD document")
	return nil
}

// commentedExamplePorts reads the `  #  - name=port` lines under the commented
// kubernetes.ports key of the annotated example, skipping the key's own wrapped
// comment lines.
func commentedExamplePorts(t *testing.T) []string {
	t.Helper()
	full, err := examples.Get(examples.FullName)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(full.Body), "\n")
	start := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "  # ports:") {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatal("the annotated example has no commented kubernetes.ports key")
	}
	var out []string
	for _, l := range lines[start+1:] {
		if item, ok := strings.CutPrefix(l, "  #  - "); ok {
			out = append(out, item)
			continue
		}
		if len(out) == 0 && strings.HasPrefix(strings.TrimSpace(l), "#") && strings.HasPrefix(l, "        ") {
			continue // the key's own comment, wrapped onto the comment column
		}
		break
	}
	return out
}

// covSwapBundle replaces the embedded operator bundle for the rest of one test and restores it
// when the test ends. The bundle is a package variable only because go:embed needs one; the
// states the tests below build with it -- a template that does not render, a bundle with no
// Namespace document, one with no CRD -- cannot be reached from the shipped asset, and each is a
// branch the apply, the delete or the render-only path has to answer for. It writes a package
// variable, so a test that uses it must not call t.Parallel; none in this package does.
func covSwapBundle(t *testing.T, bundle string) {
	t.Helper()
	orig := operatorBundle
	operatorBundle = bundle
	t.Cleanup(func() { operatorBundle = orig })
}

// covApplyRunner is a recRunner whose RunInput fails for the manifests failOn selects and
// succeeds for the rest. recRunner.runInputErr fails every call alike, so it can only ever reach
// the FIRST apply or delete of a multi-step operation; this is what lets a test let the namespace
// through and fail the Secret or the bundle that follows it.
type covApplyRunner struct {
	recRunner
	// failOn is handed each manifest on stdin and says whether its apply or delete fails.
	failOn func(manifest string) bool
	// err is what a selected manifest fails with.
	err error
}

// RunInput records the call like recRunner does, then fails it when failOn selects the manifest.
func (r *covApplyRunner) RunInput(ctx context.Context, in []byte, name string, args ...string) error {
	if err := r.recRunner.RunInput(ctx, in, name, args...); err != nil {
		return err
	}
	if r.failOn(string(in)) {
		return r.err
	}
	return nil
}

// TestRenderOperatorReportsABrokenBundleTemplate covers the two ways the embedded bundle can fail
// to render -- a template that does not parse, and one that names a variable RenderOperator does
// not supply -- and pins what they must share: an error saying which stage failed, and no bytes
// at all. A caller handed a half-rendered manifest next to the error would be one unchecked
// return away from applying it.
func TestRenderOperatorReportsABrokenBundleTemplate(t *testing.T) {
	for _, tc := range []struct {
		name   string
		bundle string
		want   string
	}{
		{"a template that does not parse", "name: {{.Namespace\n", "parse operator bundle template"},
		{"a template naming an unknown variable", "name: {{.NoSuchVariable}}\n", "render operator bundle"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			covSwapBundle(t, tc.bundle)
			out, err := RenderOperator(haCfg(), "op-ns", "solace")
			if err == nil {
				t.Fatalf("RenderOperator accepted a broken bundle and returned:\n%s", out)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to say %q", err, tc.want)
			}
			if out != nil {
				t.Errorf("a failed render must return no manifest, got %d bytes", len(out))
			}
		})
	}
}

// TestGenOperatorRefusesABundleItCannotUse is the render-only twin of the apply guards below: a
// bundle that does not render fails the whole command and prints nothing, and a bundle with no
// Namespace document cannot have the regcred spliced after it. The stream's ordering rests on
// that document, so emitting the Secret with no namespace ahead of it would print an artifact
// that fails on a first install -- the one thing a rendered artifact must not do.
func TestGenOperatorRefusesABundleItCannotUse(t *testing.T) {
	for _, tc := range []struct {
		name   string
		bundle string
		want   string
	}{
		{"a template that does not render", "name: {{.NoSuchVariable}}\n", "render operator bundle"},
		{"credentials to splice and no Namespace document to splice after",
			"kind: Deployment\nmetadata:\n  name: x\n", "no Namespace document to apply first"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			covSwapBundle(t, tc.bundle)
			out, err := GenOperator(loadK8s(t)) // the sample carries image credentials, so a regcred is due
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("GenOperator = (%q, %v), want an error saying %q", out, err, tc.want)
			}
			if out != nil {
				t.Errorf("a refused render must print nothing, got:\n%s", out)
			}
		})
	}
}

// TestOperatorApplyStopsAtARefusedPermission: the bundle creates cluster-scoped kinds, so an
// identity without the grant is turned away by the first probe -- before the installed operator
// is read, before any question, and long before the first apply. The error is the preflight's
// own, which names the verb, the kind and what to ask a cluster admin for; a deploy that got as
// far as `apply` would leave nobody sure which of the bundle's documents had landed.
func TestOperatorApplyStopsAtARefusedPermission(t *testing.T) {
	rr := &recRunner{canI: "no"}
	c := NewCluster(rr, loadK8s(t), nil, nil)
	asked := false
	c.Confirm = func(string) bool { asked = true; return true }
	err := c.OperatorApply(context.Background())
	if err == nil {
		t.Fatal("OperatorApply must fail when the permission probe answers no")
	}
	for _, want := range []string{"not allowed to create customresourcedefinitions", "role binding"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %v\nwant it to contain %q", err, want)
		}
	}
	if len(rr.calls) != 1 {
		t.Errorf("%d calls after a refused first probe, want only the probe itself: %+v", len(rr.calls), rr.calls)
	}
	if asked {
		t.Error("a refused deploy must not reach any question")
	}
}

// TestOperatorApplyAsksBeforeWideningTheWatchList: an env file that asks for EVERY namespace,
// against a shared operator that watches a named list, hands brokers this env file has never
// heard of a wider reconciler -- a privilege increase, so a question and not a side effect. An
// unattended run (no Confirm) and a plain "no" both apply nothing; only an explicit yes lets the
// empty WATCH_NAMESPACE through, and it is then applied empty rather than as the installed list.
func TestOperatorApplyAsksBeforeWideningTheWatchList(t *testing.T) {
	const question = "Let the operator watch EVERY namespace in this cluster?"
	var asked []string
	answer := func(yes bool) func(string) bool {
		return func(q string) bool { asked = append(asked, q); return yes }
	}
	for _, tc := range []struct {
		name      string
		confirm   func(string) bool
		wantErr   bool
		wantAsked bool
	}{
		{"an unattended run declines", nil, true, false},
		{"answered no", answer(false), true, true},
		{"answered yes", answer(true), false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			asked = nil
			cfg := loadK8s(t)
			cfg.K8s.Operator.Namespace = "op-ns"
			cfg.K8s.Operator.WatchNamespaces = ""
			cfg.K8s.Operator.WatchBrokerNS = boolPtr(false) // nothing listed at all: every namespace
			// Installed at the version this env file would apply, so the downgrade gate has nothing
			// to ask and the widening is the only question in play.
			installed := operatorItem("op-ns", operatorImage(cfg), "team-a")
			rr := &recRunner{out: []byte(`{"items":[` + installed + `]}`)}
			c := NewCluster(rr, cfg, nil, nil)
			c.Confirm = tc.confirm
			err := c.OperatorApply(context.Background())

			var applied []rrCall
			for _, call := range rr.calls {
				if call.method == "RunInput" {
					applied = append(applied, call)
				}
			}
			if tc.wantAsked && (len(asked) != 1 || asked[0] != question) {
				t.Errorf("asked %q, want exactly %q", asked, question)
			}
			if tc.wantErr {
				if err == nil {
					t.Fatal("a declined widening must stop the deploy")
				}
				for _, want := range []string{"declined", "nothing was applied"} {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error = %v, want it to contain %q", err, want)
					}
				}
				if len(applied) != 0 {
					t.Errorf("a declined widening still applied %d document(s): %+v", len(applied), applied)
				}
				return
			}
			if err != nil {
				t.Fatalf("OperatorApply: %v", err)
			}
			if len(applied) != 3 {
				t.Fatalf("an accepted widening made %d applies, want 3 (namespace + regcred + bundle)", len(applied))
			}
			bundle := applied[2].stdin
			if !strings.Contains(bundle, `value: ""`) || strings.Contains(bundle, "team-a") {
				t.Error("the applied WATCH_NAMESPACE must be empty (every namespace), not the installed list")
			}
		})
	}
}

// TestOperatorApplyAppliesNothingFromABundleItCannotUse: both refusals come before the first
// write, so "nothing was applied" stays a true statement. A bundle that does not render stops
// there; one with no Namespace document would apply the regcred into a namespace that does not
// exist yet -- the exact first-install failure the namespace-first ordering exists to prevent.
func TestOperatorApplyAppliesNothingFromABundleItCannotUse(t *testing.T) {
	for _, tc := range []struct {
		name   string
		bundle string
		want   string
	}{
		{"a template that does not render", "name: {{.NoSuchVariable}}\n", "render operator bundle"},
		{"no Namespace document to apply first",
			"kind: Deployment\nmetadata:\n  name: x\n", "no Namespace document to apply first"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			covSwapBundle(t, tc.bundle)
			cfg := loadK8s(t) // image credentials set: a regcred would follow the namespace
			cfg.K8s.Operator.Namespace = "op-ns"
			rr := &recRunner{}
			err := NewCluster(rr, cfg, nil, nil).OperatorApply(context.Background())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("OperatorApply err = %v, want it to say %q", err, tc.want)
			}
			for _, call := range rr.calls {
				if call.method == "RunInput" {
					t.Errorf("a bundle that cannot be used still had something applied: %+v", call)
				}
			}
		})
	}
}

// TestOperatorApplyNamesTheApplyThatFailed: the install is three applies, and a failure of any
// one must say which -- so the operator knows what landed -- carry the cause, and stop what would
// have followed. A namespace that failed must not go on to apply a Secret into it, and a Secret
// that failed must not go on to apply the Deployment that pulls with it.
func TestOperatorApplyNamesTheApplyThatFailed(t *testing.T) {
	for _, tc := range []struct {
		name        string
		failOn      func(manifest string) bool
		want        string
		wantApplies int
	}{
		{"the namespace", func(m string) bool { return strings.Contains(m, "kind: Namespace") },
			"apply operator namespace", 1},
		{"the image-pull secret", func(m string) bool { return strings.Contains(m, "kind: Secret") },
			"apply operator image-pull secret", 2},
		{"the rest of the bundle", func(m string) bool { return strings.Contains(m, "kind: Deployment") },
			"apply operator bundle", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := loadK8s(t) // image credentials set: all three applies are made
			cfg.K8s.Operator.Namespace = "op-ns"
			rr := &covApplyRunner{failOn: tc.failOn, err: errFake}
			err := NewCluster(rr, cfg, nil, nil).OperatorApply(context.Background())
			if !errors.Is(err, errFake) {
				t.Fatalf("OperatorApply err = %v, want it to carry the failed apply's cause", err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to name the failed apply (%q)", err, tc.want)
			}
			applies := 0
			for _, call := range rr.calls {
				if call.method == "RunInput" {
					applies++
				}
			}
			if applies != tc.wantApplies {
				t.Errorf("made %d apply(ies), want %d: nothing may follow the one that failed", applies, tc.wantApplies)
			}
		})
	}
}

// TestOperatorDeleteStopsAtARefusedPermission: the teardown asks `delete` on every kind it will
// touch before it touches one. Refused, nothing is deleted and the CRD layer is never even
// considered -- not the bundle, and not the broker listing that guards the CRDs, even with
// --delete-crd -- so a refusal cannot leave an operator half removed.
func TestOperatorDeleteStopsAtARefusedPermission(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{canI: "no"}
	err := NewCluster(rr, cfg, nil, nil).OperatorDelete(context.Background(), true)
	if err == nil || !strings.Contains(err.Error(), "not allowed to delete customresourcedefinitions") {
		t.Fatalf("OperatorDelete err = %v, want the refused `delete` permission", err)
	}
	if len(rr.calls) != 1 {
		t.Errorf("%d calls after a refused first probe, want only the probe itself: %+v", len(rr.calls), rr.calls)
	}
}

// TestOperatorDeleteDeletesNothingFromABundleItCannotRender: the document set to delete IS the
// rendered bundle, so a bundle that does not render leaves nothing to delete, and nothing is
// deleted -- not a partial stream, and not the CRD layer behind it.
func TestOperatorDeleteDeletesNothingFromABundleItCannotRender(t *testing.T) {
	covSwapBundle(t, "name: {{.NoSuchVariable}}\n")
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{}
	err := NewCluster(rr, cfg, nil, nil).OperatorDelete(context.Background(), true)
	if err == nil || !strings.Contains(err.Error(), "render operator bundle") {
		t.Fatalf("OperatorDelete err = %v, want the render failure", err)
	}
	if calls := rr.afterPreflights(t, probe{verb: "delete", resource: "customresourcedefinitions"}); len(calls) != 0 {
		t.Errorf("a bundle that could not be rendered still issued %+v", calls)
	}
}

// TestOperatorDeleteNamesTheDeleteThatFailed: the teardown is a bundle delete and then, when the
// CRDs are asked for, a second delete of them -- the broker listing that guards them running
// before both. A failure of either delete names which, carries the cause, and stops the rest: a
// bundle delete that failed never reaches the CRD delete, and a CRD delete that failed does not
// log the success line that says every broker resource went with them.
func TestOperatorDeleteNamesTheDeleteThatFailed(t *testing.T) {
	for _, tc := range []struct {
		name        string
		failOn      func(manifest string) bool
		want        string
		wantDeletes int
		wantListing bool
	}{
		{"the operator bundle", func(m string) bool { return strings.Contains(m, "kind: Deployment") },
			"delete operator bundle", 1, true},
		{"the CRDs", func(m string) bool { return strings.Contains(m, "kind: CustomResourceDefinition") },
			"delete operator CRDs", 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := loadK8s(t)
			cfg.K8s.Operator.Namespace = "op-ns"
			rr := &covApplyRunner{failOn: tc.failOn, err: errFake}
			buf := &bytes.Buffer{}
			c := NewCluster(rr, cfg, func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }, nil)
			err := c.OperatorDelete(context.Background(), true)
			if !errors.Is(err, errFake) {
				t.Fatalf("OperatorDelete err = %v, want it to carry the failed delete's cause", err)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to name the failed delete (%q)", err, tc.want)
			}
			deletes, listed := 0, false
			for _, call := range rr.calls {
				if call.method == "RunInput" {
					deletes++
				}
				if strings.Contains(strings.Join(call.args, " "), "get pubsubpluseventbrokers") {
					listed = true
				}
			}
			if deletes != tc.wantDeletes {
				t.Errorf("made %d delete(s), want %d: nothing may follow the one that failed", deletes, tc.wantDeletes)
			}
			if listed != tc.wantListing {
				t.Errorf("broker listing issued = %v, want %v", listed, tc.wantListing)
			}
			if strings.Contains(buf.String(), "operator CRDs deleted") {
				t.Errorf("a failed delete must not claim the CRDs were deleted:\n%s", buf.String())
			}
		})
	}
}

// TestOperatorDeleteSaysSoWhenTheBundleCarriesNoCRDs: --delete-crd on an install whose bundle
// holds no CustomResourceDefinition has nothing to remove, and says so. It neither lists broker
// resources to guard a deletion that will not happen nor pipes an empty stream to kubectl, which
// rejects one ("no objects passed to delete") and would turn a no-op into a failed teardown.
func TestOperatorDeleteSaysSoWhenTheBundleCarriesNoCRDs(t *testing.T) {
	covSwapBundle(t, "kind: Namespace\nmetadata:\n  name: {{.Namespace}}\n---\nkind: Deployment\nmetadata:\n  name: x\n")
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{}
	buf := &bytes.Buffer{}
	c := NewCluster(rr, cfg, func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }, nil)
	if err := c.OperatorDelete(context.Background(), true); err != nil {
		t.Fatalf("OperatorDelete: %v", err)
	}
	calls := rr.afterPreflights(t, probe{verb: "delete", resource: "customresourcedefinitions"})
	if len(calls) != 1 || calls[0].method != "RunInput" {
		t.Fatalf("calls after the probes = %+v, want only the delete of the non-CRD documents: "+
			"no broker listing and no CRD delete", calls)
	}
	if !strings.Contains(buf.String(), "none in the bundle") {
		t.Errorf("the output should say there were no CRDs to delete:\n%s", buf.String())
	}
}

// TestOperatorRestartStopsAtARefusedPermission: a restart is a patch on the operator's own
// Deployment, so the refusal names the OPERATOR's namespace -- the broker's would answer a
// question nobody was asking -- and no `rollout restart` is issued.
func TestOperatorRestartStopsAtARefusedPermission(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{canI: "no"}
	err := NewCluster(rr, cfg, nil, nil).OperatorRestart(context.Background())
	if err == nil || !strings.Contains(err.Error(), "not allowed to patch deployments in op-ns") {
		t.Fatalf("OperatorRestart err = %v, want the refused `patch deployments` permission in op-ns", err)
	}
	for _, call := range rr.calls {
		if call.method == "Run" {
			t.Errorf("a refused restart still issued %+v", call)
		}
	}
}

// TestOperatorScale pins both directions of `operator start` and `operator stop`. Starting waits
// for the rollout, because a start that returned before the controller was ready would say done
// about an operator that is still not reconciling. Stopping does not wait -- there is nothing to
// wait for on the way down -- and says what it costs, because the operator is cluster-scoped and
// a stopped one freezes every broker it watches, in every namespace. Both patch the operator's
// own namespace behind the same `patch deployments` probe a restart uses.
func TestOperatorScale(t *testing.T) {
	scale := func(replicas string) []string {
		return []string{"scale", "deployment", operatorDeployment, "-n", "op-ns", "--replicas=" + replicas}
	}
	rollout := []string{"rollout", "status", operatorDeployRef, "-n", "op-ns", "--timeout=" + rolloutTimeout}
	for _, tc := range []struct {
		name     string
		replicas int
		want     [][]string
		wantWarn bool
	}{
		{"start waits for the rollout", 1, [][]string{scale("1"), rollout}, false},
		{"stop does not wait, and warns", 0, [][]string{scale("0")}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := haCfg()
			cfg.K8s.Operator.Namespace = "op-ns"
			rr := &recRunner{}
			buf := &bytes.Buffer{}
			c := NewCluster(rr, cfg, func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }, nil)
			if err := c.OperatorScale(context.Background(), tc.replicas); err != nil {
				t.Fatalf("OperatorScale(%d): %v", tc.replicas, err)
			}
			calls := rr.afterPreflights(t, probe{verb: "patch", resource: "deployments"})
			if len(calls) != len(tc.want) {
				t.Fatalf("OperatorScale(%d) made %d call(s) after the probe, want %d: %+v",
					tc.replicas, len(calls), len(tc.want), calls)
			}
			for i, want := range tc.want {
				if calls[i].method != "Run" || calls[i].name != "kubectl" || !eqArgs(calls[i].args, want) {
					t.Errorf("call %d = %+v, want Run kubectl %v", i, calls[i], want)
				}
			}
			if ns, found := rr.probedNamespace("patch", "deployments"); !found || ns != "op-ns" {
				t.Errorf("probed namespace = %q (found %v), want the operator's own (op-ns)", ns, found)
			}
			if stopped := strings.Contains(buf.String(), "the operator is stopped"); stopped != tc.wantWarn {
				t.Errorf("stop warning printed = %v, want %v:\n%s", stopped, tc.wantWarn, buf.String())
			}
		})
	}
}

// TestOperatorScaleStopsAtARefusedPermission: scaling the controller is a patch on its
// Deployment, so a refused grant ends it at the probe -- the Deployment is neither scaled nor
// waited on, in either direction.
func TestOperatorScaleStopsAtARefusedPermission(t *testing.T) {
	for _, replicas := range []int{0, 1} {
		t.Run(fmt.Sprintf("replicas=%d", replicas), func(t *testing.T) {
			cfg := haCfg()
			cfg.K8s.Operator.Namespace = "op-ns"
			rr := &recRunner{canI: "no"}
			err := NewCluster(rr, cfg, nil, nil).OperatorScale(context.Background(), replicas)
			if err == nil || !strings.Contains(err.Error(), "not allowed to patch deployments in op-ns") {
				t.Fatalf("OperatorScale(%d) err = %v, want the refused `patch deployments` permission", replicas, err)
			}
			if len(rr.calls) != 1 {
				t.Errorf("%d calls after a refused probe, want only the probe itself: %+v", len(rr.calls), rr.calls)
			}
		})
	}
}

// TestOperatorScaleSurfacesAFailure: a scale that kubectl refused is an error, not a state -- and
// a failed STOP must not claim the operator is stopped, which would tell someone it is safe to
// edit a StatefulSet the controller is still reconciling. A start whose rollout never completes
// is an error too, reported after the scale that was already issued.
func TestOperatorScaleSurfacesAFailure(t *testing.T) {
	for _, tc := range []struct {
		name      string
		replicas  int
		runErrs   []error
		wantCalls int
	}{
		{"a stop that kubectl refuses", 0, []error{errFake}, 1},
		{"a start that kubectl refuses", 1, []error{errFake}, 1},
		{"a start whose rollout does not complete", 1, []error{nil, errFake}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := haCfg()
			cfg.K8s.Operator.Namespace = "op-ns"
			rr := &recRunner{runErrQueue: tc.runErrs}
			buf := &bytes.Buffer{}
			c := NewCluster(rr, cfg, func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }, nil)
			err := c.OperatorScale(context.Background(), tc.replicas)
			if !errors.Is(err, errFake) {
				t.Fatalf("OperatorScale(%d) err = %v, want the failed kubectl call's error", tc.replicas, err)
			}
			if calls := rr.afterPreflights(t, probe{verb: "patch", resource: "deployments"}); len(calls) != tc.wantCalls {
				t.Errorf("made %d call(s) after the probe, want %d: %+v", len(calls), tc.wantCalls, calls)
			}
			if strings.Contains(buf.String(), "the operator is stopped") {
				t.Errorf("a failed scale must not claim the operator is stopped:\n%s", buf.String())
			}
		})
	}
}

package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"solace/internal/config"
)

// The operator's watch scope is the one field of the install that does not belong to any
// single env file, so these tests are about a shared resource being edited by several
// callers rather than about string handling.
//
// The bug they exist to prevent had no error and no log line: `operator deploy` from a
// second env file overwrote WATCH_NAMESPACE, the first broker's namespace fell out of it,
// and that broker simply stopped reconciling. Nothing failed -- it just stopped.

// --- parsing what is installed --------------------------------------------------

// TestWatchFromContainersReadsEveryAllNamespacesSpelling pins the decode, and in
// particular that the three ways of saying "watch everything" are ONE answer.
//
// They arrive differently -- an operator installed by hand has no WATCH_NAMESPACE entry at
// all, the vendor bundle ships `value: ""`, and an edited Deployment can end up with
// separators and nothing else -- and every caller either leaves an all-namespaces operator
// alone or refuses to narrow it, so collapsing them is what makes those callers correct.
func TestWatchFromContainersReadsEveryAllNamespacesSpelling(t *testing.T) {
	env := func(pairs ...string) []containerSpec {
		var c containerSpec
		for i := 0; i < len(pairs); i += 2 {
			c.Env = append(c.Env, envVar{Name: pairs[i], Value: pairs[i+1]})
		}
		return []containerSpec{c}
	}
	for _, tc := range []struct {
		name       string
		containers []containerSpec
		wantList   []string
		wantAllNS  bool
	}{
		{"a listed namespace", env(watchEnvVar, "solace"), []string{"solace"}, false},
		{"several", env(watchEnvVar, "a,b,c"), []string{"a", "b", "c"}, false},
		{"trimmed and de-duplicated", env(watchEnvVar, " a , b ,a"), []string{"a", "b"}, false},
		{"empty value means every namespace", env(watchEnvVar, ""), nil, true},
		{"separators only means every namespace", env(watchEnvVar, " , , "), nil, true},
		{"no such variable means every namespace", env("OTHER", "x"), nil, true},
		{"no containers at all", nil, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			list, allNS := watchFromContainers(tc.containers)
			if allNS != tc.wantAllNS {
				t.Errorf("allNS = %v, want %v", allNS, tc.wantAllNS)
			}
			if strings.Join(list, ",") != strings.Join(tc.wantList, ",") {
				t.Errorf("list = %v, want %v", list, tc.wantList)
			}
		})
	}
}

// TestDesiredWatchAppendsTheBrokerNamespace covers the config half, including the one
// combination that asks for every namespace.
func TestDesiredWatchAppendsTheBrokerNamespace(t *testing.T) {
	cfg := func(watch string, brokerNS *bool) *config.Config {
		c := &config.Config{}
		c.K8s.Namespace = "solace"
		c.K8s.Operator.WatchNamespaces = watch
		c.K8s.Operator.WatchBrokerNS = brokerNS
		return c
	}
	no := false
	yes := true
	for _, tc := range []struct {
		name string
		cfg  *config.Config
		want []string
	}{
		{"default: the broker namespace alone", cfg("", nil), []string{"solace"}},
		{"configured, broker ns appended", cfg("team-a", nil), []string{"team-a", "solace"}},
		{"already listed, not repeated", cfg("solace,team-a", nil), []string{"solace", "team-a"}},
		{"explicit true behaves as the default", cfg("team-a", &yes), []string{"team-a", "solace"}},
		{"broker ns opted out", cfg("team-a", &no), []string{"team-a"}},
		// The only route to "watch everything" from config, and it takes saying so twice.
		{"opted out with nothing configured means ALL", cfg("", &no), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := desiredWatch(tc.cfg); strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("desiredWatch = %v, want %v", got, tc.want)
			}
		})
	}
}

// --- the algebra ------------------------------------------------------------------

// TestUnionWatchPreservesInstalledOrder is the ordering rule, and it is not cosmetic.
//
// The list becomes one env-var value on a pod template, so a reordered but equivalent list
// is a CHANGED template: Kubernetes rolls the operator pod, reconciliation pauses, and
// nothing about the deployment actually differs. Preserving the installed order means an
// unchanged set renders byte-identically and the Deployment is left alone.
func TestUnionWatchPreservesInstalledOrder(t *testing.T) {
	for _, tc := range []struct {
		name                string
		installed, desired  []string
		wantMerged, wantAdd []string
	}{
		{"nothing new", []string{"a", "b"}, []string{"b"}, []string{"a", "b"}, nil},
		{"appended in desired's order", []string{"a"}, []string{"c", "b"}, []string{"a", "c", "b"}, []string{"c", "b"}},
		{"a different order is NOT a change", []string{"b", "a"}, []string{"a", "b"}, []string{"b", "a"}, nil},
		{"first install", nil, []string{"a", "b"}, []string{"a", "b"}, []string{"a", "b"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			merged, added := unionWatch(tc.installed, tc.desired)
			if strings.Join(merged, ",") != strings.Join(tc.wantMerged, ",") {
				t.Errorf("merged = %v, want %v", merged, tc.wantMerged)
			}
			if strings.Join(added, ",") != strings.Join(tc.wantAdd, ",") {
				t.Errorf("added = %v, want %v -- an empty added is what tells the caller "+
					"nothing needs applying", added, tc.wantAdd)
			}
		})
	}
}

// TestUnionWatchKeepsAnotherEnvFilesNamespace is the whole point of the union, stated as
// the scenario rather than as set algebra: two env files, two namespaces, one operator.
func TestUnionWatchKeepsAnotherEnvFilesNamespace(t *testing.T) {
	installed := []string{"team-a"} // deployed earlier, from another env file
	desired := []string{"team-b"}   // this env file
	merged, added := unionWatch(installed, desired)
	if strings.Join(merged, ",") != "team-a,team-b" {
		t.Fatalf("merged = %v, want both namespaces: deploying the second broker must not "+
			"stop the first one from being reconciled", merged)
	}
	if strings.Join(added, ",") != "team-b" {
		t.Errorf("added = %v, want [team-b]", added)
	}
}

// TestSubtractWatchEmptyRemainingMeansDelete pins the case the caller MUST NOT apply.
//
// An empty remaining list is not "watch nothing" -- an empty WATCH_NAMESPACE means watch
// EVERY namespace. Applying it as a removal would hand the operator the whole cluster: a
// privilege escalation dressed as a teardown. The caller has to read empty as "delete the
// operator" instead, so this test asserts the shape that decision is made on.
func TestSubtractWatchEmptyRemainingMeansDelete(t *testing.T) {
	for _, tc := range []struct {
		name                    string
		installed, desired      []string
		wantRemain, wantRemoved []string
	}{
		{"exactly ours", []string{"a"}, []string{"a"}, nil, []string{"a"}},
		{"ours plus more than we know", []string{"a"}, []string{"a", "b"}, nil, []string{"a"}},
		{"someone else still needs it", []string{"a", "b"}, []string{"a"}, []string{"b"}, []string{"a"}},
		{"nothing of ours installed", []string{"a"}, []string{"b"}, []string{"a"}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			remaining, removed := subtractWatch(tc.installed, tc.desired)
			if strings.Join(remaining, ",") != strings.Join(tc.wantRemain, ",") {
				t.Errorf("remaining = %v, want %v", remaining, tc.wantRemain)
			}
			if strings.Join(removed, ",") != strings.Join(tc.wantRemoved, ",") {
				t.Errorf("removed = %v, want %v", removed, tc.wantRemoved)
			}
		})
	}
}

// TestSubtractWatchPreservesOrder: the same rolling-the-pod-for-nothing concern as the
// union, on the removal path.
func TestSubtractWatchPreservesOrder(t *testing.T) {
	remaining, _ := subtractWatch([]string{"c", "a", "b"}, []string{"a"})
	if strings.Join(remaining, ",") != "c,b" {
		t.Errorf("remaining = %v, want [c b]: the surviving entries keep the order they "+
			"were installed in, so the value only changes when the SET changes", remaining)
	}
}

// --- the decisions the algebra feeds ----------------------------------------------

// deployJSON is one operator Deployment as `get deployment --all-namespaces -o json`
// returns it, carrying whatever WATCH_NAMESPACE the caller wants installed.
func deployJSON(watch string) []byte {
	env := ""
	if watch != "(absent)" {
		env = `,"env":[{"name":"WATCH_NAMESPACE","value":"` + watch + `"}]`
	}
	return []byte(`{"items":[{"metadata":{"name":"pubsubplus-eventbroker-operator",` +
		`"namespace":"pubsubplus-operator-system"},"spec":{"template":{"spec":{"containers":[` +
		`{"image":"solace/pubsubplus-eventbroker-operator:1.4.2"` + env + `}]}}}}]}`)
}

func watchCluster(watch string, cfgWatch string) (*Cluster, *recRunner, *bytes.Buffer) {
	cfg := &config.Config{}
	cfg.K8s.Command = config.Command{"kubectl"}
	cfg.K8s.Namespace = "solace"
	cfg.K8s.Name = "dev-broker"
	cfg.K8s.Operator.WatchNamespaces = cfgWatch
	buf := &bytes.Buffer{}
	rr := &recRunner{}
	if watch != "(none)" {
		rr.out = deployJSON(watch)
	} else {
		rr.out = []byte(`{"items":[]}`)
	}
	c := &Cluster{R: rr, Cfg: cfg, Out: buf,
		Log: func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }}
	return c, rr, buf
}

// TestReconcileWatchDecidesWhatToApply covers every branch of the deploy-side decision.
//
// The row that matters most is the last: an env file asking for ALL namespaces against an
// operator that watches a specific list is a WIDENING of a shared resource, so it is
// flagged for confirmation rather than applied as a side effect of a deploy.
func TestReconcileWatchDecidesWhatToApply(t *testing.T) {
	for _, tc := range []struct {
		name         string
		installed    string // "(none)" = not installed, "(absent)" = no env var
		cfgWatch     string
		wantValue    string
		wantWidening bool
	}{
		{"first install", "(none)", "team-a", "team-a,solace", false},
		{"already watches everything", "", "team-a", "", false},
		{"no env var reads as everything", "(absent)", "team-a", "", false},
		{"union adds ours", "team-a", "team-b", "team-a,team-b,solace", false},
		{"already covered, unchanged", "team-a,solace", "team-a", "team-a,solace", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _, _ := watchCluster(tc.installed, tc.cfgWatch)
			plan, err := c.reconcileWatch(context.Background())
			if err != nil {
				t.Fatalf("reconcileWatch: %v", err)
			}
			if plan.value != tc.wantValue {
				t.Errorf("value = %q, want %q", plan.value, tc.wantValue)
			}
			if plan.widening != tc.wantWidening {
				t.Errorf("widening = %v, want %v", plan.widening, tc.wantWidening)
			}
		})
	}
}

// operatorItem renders one Deployment item carrying the operator's name, for fixtures that
// need more than the single item deployJSON builds.
func operatorItem(ns, image, watch string) string {
	return `{"metadata":{"name":"pubsubplus-eventbroker-operator","namespace":"` + ns + `"},` +
		`"spec":{"template":{"spec":{"containers":[{"image":"` + image + `",` +
		`"env":[{"name":"WATCH_NAMESPACE","value":"` + watch + `"}]}]}}}}`
}

// TestFindOperatorDeploymentIsScopedByNamespace is the guard on the lookup two decisions
// rest on -- whether the watch list is widened, and whether a deploy is a downgrade. A
// name-only match across every namespace let a same-named Deployment ANYWHERE answer for
// the operator. The resolved namespace wins outright; an install elsewhere counts only
// when nothing is there and it is the sole candidate; two elsewhere is an error.
func TestFindOperatorDeploymentIsScopedByNamespace(t *testing.T) {
	const resolved = "pubsubplus-operator-system" // the default operatorNS resolves to
	for _, tc := range []struct {
		name      string
		items     []string
		wantNS    string // "" = nil result
		wantErr   string
		wantImage string
	}{
		{"resolved wins over an impostor listed first",
			[]string{operatorItem("attacker", "evil/op:9.9.9", ""), operatorItem(resolved, "solace/op:1.4.2", "team-a")},
			resolved, "", "solace/op:1.4.2"},
		{"one elsewhere is still found, for the installed-somewhere-else warning",
			[]string{operatorItem("ops-team", "solace/op:1.4.2", "team-a")},
			"ops-team", "", "solace/op:1.4.2"},
		{"two elsewhere cannot be told apart",
			[]string{operatorItem("ns-a", "solace/op:1.4.2", ""), operatorItem("ns-b", "solace/op:1.5.0", "")},
			"", "ns-a, ns-b", ""},
		{"none installed", nil, "", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.K8s.Command = config.Command{"kubectl"}
			cfg.K8s.Namespace = "solace"
			rr := &recRunner{out: []byte(`{"items":[` + strings.Join(tc.items, ",") + `]}`)}
			buf := &bytes.Buffer{}
			c := &Cluster{R: rr, Cfg: cfg, Out: buf,
				Log: func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }}
			dep, err := c.findOperatorDeployment(context.Background())
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want it to name %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("findOperatorDeployment: %v", err)
			}
			if tc.wantNS == "" {
				if dep != nil {
					t.Fatalf("got a Deployment in %q, want none", dep.Metadata.Namespace)
				}
				return
			}
			if dep == nil || dep.Metadata.Namespace != tc.wantNS {
				t.Fatalf("got %+v, want the Deployment in %q", dep, tc.wantNS)
			}
			if got := imageFromDeployment(dep); got != tc.wantImage {
				t.Errorf("image = %q, want %q", got, tc.wantImage)
			}
		})
	}
}

// TestAmbiguousOperatorIsNotFoldedIntoNotInstalled is the second half of the namespace
// scoping, and the half that matters most. installedOperatorImage deliberately answers ""
// for a read that did not happen -- a first install has no operator namespace and must not
// alarm -- but "" also means "nothing installed", which makes confirmNoDowngrade skip the
// question entirely. Folding the AMBIGUITY error in there would disable the downgrade
// guard using the very condition it was added to catch, so that one error is returned.
func TestAmbiguousOperatorIsNotFoldedIntoNotInstalled(t *testing.T) {
	cfg := &config.Config{}
	cfg.K8s.Command = config.Command{"kubectl"}
	cfg.K8s.Namespace = "solace"
	items := operatorItem("ns-a", "solace/op:1.4.2", "") + "," + operatorItem("ns-b", "solace/op:1.5.0", "")
	rr := &recRunner{out: []byte(`{"items":[` + items + `]}`)}
	c := &Cluster{R: rr, Cfg: cfg, Out: &bytes.Buffer{}}

	if _, err := c.installedOperatorImage(context.Background()); !errors.Is(err, errAmbiguousOperator) {
		t.Fatalf("installedOperatorImage err = %v, want it to carry errAmbiguousOperator", err)
	}
	// A transport failure still folds to "", which is what keeps a first install quiet.
	rr2 := &recRunner{outErrQueue: []error{errFake}}
	c2 := &Cluster{R: rr2, Cfg: cfg, Out: &bytes.Buffer{}}
	img, err := c2.installedOperatorImage(context.Background())
	if err != nil || img != "" {
		t.Errorf("a failed read = (%q, %v), want (\"\", nil) so a first install does not alarm", img, err)
	}
}

// TestReconcileWatchFlagsAWideningAsAQuestion is that last row on its own, because it is
// the one with a consequence: going from a named list to every namespace hands a shared
// operator the whole cluster, so the plan has to say so rather than just applying it.
func TestReconcileWatchFlagsAWideningAsAQuestion(t *testing.T) {
	cfg := &config.Config{}
	cfg.K8s.Command = config.Command{"kubectl"}
	cfg.K8s.Namespace = "solace"
	no := false
	cfg.K8s.Operator.WatchBrokerNS = &no // with no watchNamespaces: "watch everything"
	rr := &recRunner{out: deployJSON("team-a")}
	buf := &bytes.Buffer{}
	c := &Cluster{R: rr, Cfg: cfg, Out: buf,
		Log: func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }}

	plan, err := c.reconcileWatch(context.Background())
	if err != nil {
		t.Fatalf("reconcileWatch: %v", err)
	}
	if !plan.widening {
		t.Error("widening = false: taking an operator from a named list to ALL namespaces is a " +
			"privilege increase and must be confirmed, not applied silently")
	}
	if plan.value != "" {
		t.Errorf("value = %q, want empty (every namespace)", plan.value)
	}
	if !strings.Contains(buf.String(), "team-a") {
		t.Errorf("the warning should name what it currently watches:\n%s", buf.String())
	}
}

// TestOperatorReleaseNarrowsInsteadOfRemoving is the removal-side decision, and the reason
// `operator remove` is not just a delete: the operator is shared, so letting go of one
// namespace must not take it away from the others.
func TestOperatorReleaseNarrowsInsteadOfRemoving(t *testing.T) {
	c, rr, buf := watchCluster("team-a,solace", "")
	if err := c.OperatorRelease(context.Background(), false); err != nil {
		t.Fatalf("OperatorRelease: %v", err)
	}
	var setEnv *rrCall
	for i := range rr.calls {
		if len(rr.calls[i].args) > 1 && rr.calls[i].args[0] == "set" && rr.calls[i].args[1] == "env" {
			setEnv = &rr.calls[i]
		}
		if rr.calls[i].method == "RunInput" {
			t.Fatalf("the operator was DELETED while team-a still needs it: %+v", rr.calls[i])
		}
	}
	if setEnv == nil {
		t.Fatalf("no `set env` call: the watch list should have been narrowed.\ncalls: %+v\n%s",
			rr.calls, buf.String())
	}
	joined := strings.Join(setEnv.args, " ")
	if !strings.Contains(joined, "WATCH_NAMESPACE=team-a") || strings.Contains(joined, "solace") {
		t.Errorf("set env argv = %v, want WATCH_NAMESPACE=team-a with our namespace dropped", setEnv.args)
	}
}

// TestSetWatchRefusesAnEmptyList guards the trap the whole design exists around: an empty
// WATCH_NAMESPACE means watch EVERY namespace, so writing one as the end of a removal
// would leave the operator with WIDER scope than it started with. The caller is meant to
// delete the operator instead, and this refuses in case a future caller forgets.
func TestSetWatchRefusesAnEmptyList(t *testing.T) {
	c, rr, _ := watchCluster("team-a", "")
	err := c.setWatch(context.Background(), "pubsubplus-operator-system", nil)
	if err == nil {
		t.Fatal("setWatch(nil) must refuse: it would widen the operator to every namespace")
	}
	for _, want := range []string{"EVERY", "widen"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should explain the widening, not just refuse", err)
		}
	}
	for _, call := range rr.calls {
		if len(call.args) > 1 && call.args[0] == "set" {
			t.Errorf("a refused setWatch still issued %v", call.args)
		}
	}
}

package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"solace/internal/config"
	"solace/internal/engine"
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
			// Fed the way OperatorApply feeds it: the Deployment is read once, then handed in.
			dep, _, err := c.findOperatorDeployment(context.Background())
			if err != nil {
				t.Fatalf("findOperatorDeployment: %v", err)
			}
			plan := c.reconcileWatch(dep)
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
			dep, onlyNS, err := c.findOperatorDeployment(context.Background())
			// A list that answered is final, whatever it said -- including the ambiguity
			// error: no namespace fallback, no warning, and the whole cluster searched.
			if len(rr.calls) != 1 {
				t.Errorf("made %d reads, want exactly the one cluster-wide list: %+v", len(rr.calls), rr.calls)
			}
			if onlyNS != "" || strings.Contains(buf.String(), "could not list") {
				t.Errorf("onlyNS = %q, output %q: a list that answered must not read as a fallback", onlyNS, buf.String())
			}
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

// TestAmbiguousOperatorAbortsTheDeploy is the second half of the namespace scoping, and the
// half that matters most. Two operators elsewhere is an ANSWER -- the operator is installed,
// more than once -- and reading it as "nothing installed" would skip the downgrade question
// using the very condition it was added to catch. So OperatorApply stops on it after its one
// read: nothing asked, nothing applied, and no namespace fallback, because the list worked.
func TestAmbiguousOperatorAbortsTheDeploy(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	items := operatorItem("ns-a", "solace/op:1.4.2", "") + "," + operatorItem("ns-b", "solace/op:1.5.0", "")
	rr := &recRunner{out: []byte(`{"items":[` + items + `]}`)}
	c := NewCluster(rr, cfg, nil, nil)
	asked := false
	c.Confirm = func(string) bool { asked = true; return true }

	err := c.OperatorApply(context.Background())
	if !errors.Is(err, errAmbiguousOperator) {
		t.Fatalf("OperatorApply err = %v, want it to carry errAmbiguousOperator", err)
	}
	for _, want := range []string{"nothing was applied", "ns-a, ns-b", "kubernetes.operator.namespace"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %v, want it to contain %q", err, want)
		}
	}
	if asked {
		t.Error("an ambiguous install must stop the deploy, not reach a question")
	}
	calls := rr.afterPreflights(t, probe{verb: "create", resource: "customresourcedefinitions"})
	if len(calls) != 1 || !eqArgs(calls[0].args, []string{"get", "deployment", "--all-namespaces", "-o", "json"}) {
		t.Errorf("calls after the probe = %+v, want only the one cluster-wide read", calls)
	}
}

// --- an identity that may not list Deployments cluster-wide -----------------------

// errListDenied stands in for kubectl refusing `get deployment --all-namespaces`. The
// runner only ever sees the exit status: kubectl's Forbidden goes to the terminal.
var errListDenied = errors.New("kubectl: exit status 1")

// namedOperatorJSON is the operator Deployment as a get BY NAME returns it: the bare
// object, with a kind and no items. The List-shaped builders above would still pass if
// normalizeToList stopped wrapping that shape, which is the miss inspect.go records.
func namedOperatorJSON(ns, image, watch string) []byte {
	return []byte(`{"kind":"Deployment",` + strings.TrimPrefix(operatorItem(ns, image, watch), "{"))
}

// deniedRunner is a recRunner for that identity: the cluster-wide list fails with
// errListDenied, and the by-name read in the operator namespace answers named -- nil
// being what --ignore-not-found prints when nothing is there. `auth can-i` probes bypass
// both queues, so a preflight never shifts them.
func deniedRunner(named []byte) *recRunner {
	return &recRunner{outQueue: [][]byte{nil, named}, outErrQueue: []error{errListDenied, nil}}
}

// deniedCluster is a Cluster over deniedRunner whose narration lands in the returned buffer.
// opNS is kubernetes.operator.namespace; "" leaves it to the default.
func deniedCluster(opNS string, named []byte) (*Cluster, *recRunner, *bytes.Buffer) {
	cfg := &config.Config{}
	cfg.K8s.Command = config.Command{"kubectl"}
	cfg.K8s.Namespace = "solace"
	cfg.K8s.Name = "dev-broker"
	cfg.K8s.Operator.Namespace = opNS
	rr := deniedRunner(named)
	buf := &bytes.Buffer{}
	c := &Cluster{R: rr, Cfg: cfg, Out: buf,
		Log: func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }}
	return c, rr, buf
}

// TestFindOperatorDeploymentFallsBackWhenTheClusterWideListFails is the lookup half of the
// reported failure: kubectl refused the cluster-wide list, and the whole command stopped.
// The list failing now sends the lookup to the one namespace this env file resolves, read
// by name, and says once how far the answer reaches.
func TestFindOperatorDeploymentFallsBackWhenTheClusterWideListFails(t *testing.T) {
	captured, err := os.ReadFile(filepath.Join("testdata", "operator-deploy-named.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	for _, tc := range []struct {
		name      string
		opNS      string // kubernetes.operator.namespace
		named     []byte // the by-name read's reply
		wantNS    string // the namespace read; also onlyNS
		wantFound bool
		wantWarn  string
	}{
		{"found: a captured by-name reply", "solace-operator", captured, "solace-operator", true, "found there"},
		{"found: the default namespace", "",
			namedOperatorJSON(defaultOperatorNS, "solace/op:1.4.2", "team-a"), defaultOperatorNS, true, "found there"},
		{"nothing there: --ignore-not-found prints nothing", "op-ns", nil, "op-ns", false,
			"set kubernetes.operator.namespace to it"},
		{"nothing there: an empty List", "op-ns", []byte(`{"apiVersion":"v1","items":[],"kind":"List"}`),
			"op-ns", false, "is not there"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, rr, buf := deniedCluster(tc.opNS, tc.named)
			dep, onlyNS, err := c.findOperatorDeployment(context.Background())
			if err != nil {
				t.Fatalf("a refused list with a readable namespace is not an error: %v", err)
			}
			if onlyNS != tc.wantNS {
				t.Errorf("onlyNS = %q, want %q", onlyNS, tc.wantNS)
			}
			if (dep != nil) != tc.wantFound {
				t.Fatalf("dep = %+v, want found=%v", dep, tc.wantFound)
			}
			if tc.wantFound && dep.Metadata.Namespace != tc.wantNS {
				t.Errorf("found the Deployment in %q, want %q", dep.Metadata.Namespace, tc.wantNS)
			}
			want := [][]string{
				{"get", "deployment", "--all-namespaces", "-o", "json"},
				{"get", "deployment", operatorDeployment, "-n", tc.wantNS, "--ignore-not-found", "-o", "json"},
			}
			if len(rr.calls) != len(want) {
				t.Fatalf("made %d reads, want %d: %+v", len(rr.calls), len(want), rr.calls)
			}
			for i, w := range want {
				if rr.calls[i].method != "Output" || !eqArgs(rr.calls[i].args, w) {
					t.Errorf("read %d = %+v, want Output %v", i, rr.calls[i], w)
				}
			}
			out := buf.String()
			if n := strings.Count(out, "could not list Deployments cluster-wide"); n != 1 {
				t.Errorf("the fallback warning printed %d times, want exactly 1:\n%s", n, out)
			}
			for _, w := range []string{tc.wantNS, tc.wantWarn, "not seen"} {
				if !strings.Contains(out, w) {
					t.Errorf("the warning should say %q:\n%s", w, out)
				}
			}
		})
	}
}

// TestFindOperatorDeploymentNamesBothReadsWhenBothFail: when the namespace read fails too,
// there is no answer at all. The error names both reads and what to ask for, carries both
// causes, prints no fallback warning (there was no fallback answer to qualify), and is
// never the ambiguity error -- that one means the operator IS installed.
func TestFindOperatorDeploymentNamesBothReadsWhenBothFail(t *testing.T) {
	errNamed := errors.New("kubectl: exit status 1 (named read)")
	c, rr, buf := deniedCluster("op-ns", nil)
	rr.outErrQueue = []error{errListDenied, errNamed}
	dep, onlyNS, err := c.findOperatorDeployment(context.Background())
	if dep != nil || onlyNS != "" || err == nil {
		t.Fatalf("got (%+v, %q, %v), want (nil, \"\", an error)", dep, onlyNS, err)
	}
	if !errors.Is(err, errListDenied) || !errors.Is(err, errNamed) {
		t.Errorf("error = %v, want it to wrap both failed reads", err)
	}
	if errors.Is(err, errAmbiguousOperator) {
		t.Error("a read nobody could make is not an ambiguous install")
	}
	for _, want := range []string{"cluster-wide", operatorDeployment, "namespace op-ns", "on deployments in op-ns"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %v, want it to contain %q", err, want)
		}
	}
	if strings.Contains(buf.String(), "could not list") {
		t.Errorf("no fallback answer, so no fallback warning:\n%s", buf.String())
	}
}

// TestReconcileWatchUnionsWithAnOperatorFoundByTheFallback: the watch list read off the
// operator the fallback found is merged exactly as one from the cluster-wide list would
// be. Applying this env file's list alone is the bug watch.go exists to prevent, and a
// restricted identity must not bring it back.
func TestReconcileWatchUnionsWithAnOperatorFoundByTheFallback(t *testing.T) {
	c, _, buf := deniedCluster("", namedOperatorJSON(defaultOperatorNS, "solace/op:1.4.2", "team-a"))
	c.Cfg.K8s.Operator.WatchNamespaces = "team-b"
	dep, _, err := c.findOperatorDeployment(context.Background())
	if err != nil {
		t.Fatalf("findOperatorDeployment: %v", err)
	}
	plan := c.reconcileWatch(dep)
	if plan.value != "team-a,team-b,solace" || plan.widening {
		t.Errorf("plan = %+v, want the union team-a,team-b,solace and no widening", plan)
	}
	if !strings.Contains(buf.String(), "widening the operator watch list: + team-b,solace") {
		t.Errorf("the union should be narrated:\n%s", buf.String())
	}
}

// TestOperatorReleaseUnderADeniedList is the removal side: `operator remove` read the watch
// list with the same cluster-wide list and stopped the same way. It now acts on what the
// operator namespace holds -- the only namespace a removal ever touches -- and, finding
// nothing there, removes nothing.
func TestOperatorReleaseUnderADeniedList(t *testing.T) {
	t.Run("found: narrowed, not deleted", func(t *testing.T) {
		c, rr, buf := deniedCluster("op-ns", namedOperatorJSON("op-ns", "solace/op:1.4.2", "team-a,solace"))
		if err := c.OperatorRelease(context.Background(), false); err != nil {
			t.Fatalf("OperatorRelease: %v\n%s", err, buf.String())
		}
		var setEnv []string
		for _, call := range rr.calls {
			if call.method == "RunInput" {
				t.Fatalf("the operator was DELETED while team-a still needs it: %+v", call)
			}
			if len(call.args) > 1 && call.args[0] == "set" && call.args[1] == "env" {
				setEnv = call.args
			}
		}
		want := []string{"set", "env", operatorDeployRef, "-n", "op-ns", "WATCH_NAMESPACE=team-a"}
		if !eqArgs(setEnv, want) {
			t.Errorf("set env argv = %v, want %v", setEnv, want)
		}
	})
	t.Run("nothing there: nothing removed", func(t *testing.T) {
		c, rr, buf := deniedCluster("op-ns", nil)
		if err := c.OperatorRelease(context.Background(), false); err != nil {
			t.Fatalf("OperatorRelease: %v", err)
		}
		for _, call := range rr.calls {
			if call.method != "Output" {
				t.Errorf("nothing was found, so nothing may be changed: %+v", call)
			}
		}
		for _, want := range []string{"is not there", "nothing to remove"} {
			if !strings.Contains(buf.String(), want) {
				t.Errorf("output should say %q:\n%s", want, buf.String())
			}
		}
	})
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

	dep, _, err := c.findOperatorDeployment(context.Background())
	if err != nil {
		t.Fatalf("findOperatorDeployment: %v", err)
	}
	plan := c.reconcileWatch(dep)
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

// TestFindOperatorDeploymentIgnoresOtherDeployments: `get deployment --all-namespaces` returns
// every Deployment in the cluster and the lookup keys on the operator's NAME, so its neighbours
// must be passed over rather than mistaken for it -- and a cluster holding only neighbours
// answers "no operator installed", which is what lets a first install proceed.
func TestFindOperatorDeploymentIgnoresOtherDeployments(t *testing.T) {
	const resolved = "pubsubplus-operator-system" // the default operatorNS resolves to
	neighbour := func(ns, name string) string {
		return `{"metadata":{"name":"` + name + `","namespace":"` + ns + `"}}`
	}
	coredns, nginx := neighbour("kube-system", "coredns"), neighbour("ingress", "nginx")
	operator := operatorItem(resolved, "solace/op:1.4.2", "team-a")
	for _, tc := range []struct {
		name  string
		items []string
		want  bool
	}{
		{"alongside the operator", []string{coredns, operator, nginx}, true},
		{"neighbours only", []string{coredns, nginx}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, rr, _ := watchCluster("(none)", "")
			rr.out = []byte(`{"items":[` + strings.Join(tc.items, ",") + `]}`)
			dep, onlyNS, err := c.findOperatorDeployment(context.Background())
			if err != nil {
				t.Fatalf("findOperatorDeployment: %v", err)
			}
			if onlyNS != "" {
				t.Errorf("onlyNS = %q, want the whole cluster searched", onlyNS)
			}
			if (dep != nil) != tc.want {
				t.Fatalf("got %+v, want found=%v", dep, tc.want)
			}
			if tc.want && dep.Metadata.Namespace != resolved {
				t.Errorf("found the Deployment in %q, want the operator in %q", dep.Metadata.Namespace, resolved)
			}
		})
	}
}

// TestOperatorApplyDoesNotAskForAFirstInstallThatWatchesEveryNamespace: with nothing installed
// there is no watch list to widen, so an env file that asks for every namespace (watchBrokerNs:
// false and no watchNamespaces) installs exactly that WITHOUT the question -- which exists to stop
// a deploy from widening a SHARED operator, not to second-guess a fresh one. It is still said out
// loud, because every namespace is the widest scope the operator has and the least likely to have
// been chosen on purpose.
func TestOperatorApplyDoesNotAskForAFirstInstallThatWatchesEveryNamespace(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	cfg.K8s.Operator.WatchNamespaces = ""
	cfg.K8s.Operator.WatchBrokerNS = boolPtr(false)
	rr := &recRunner{} // nothing installed
	buf := &bytes.Buffer{}
	c := NewCluster(rr, cfg, func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }, buf)
	asked := false
	c.Confirm = func(string) bool { asked = true; return false }
	if err := c.OperatorApply(context.Background()); err != nil {
		t.Fatalf("OperatorApply: %v\n%s", err, buf.String())
	}
	if asked {
		t.Error("a first install has nothing to widen, so it must not be asked about")
	}
	if !strings.Contains(buf.String(), "installing the operator to watch ALL namespaces") {
		t.Errorf("a cluster-wide first install should be said out loud:\n%s", buf.String())
	}
	var bundle string
	for _, call := range rr.calls {
		if call.method == "RunInput" {
			bundle = call.stdin // the last apply is the rest of the bundle, Deployment included
		}
	}
	if !strings.Contains(bundle, `value: ""`) {
		t.Error("the applied bundle must carry an empty WATCH_NAMESPACE (every namespace)")
	}
}

// TestOperatorReleaseStopsWhenTheInstalledOperatorCannotBeRead: removal acts on what the operator
// watches, so an operator nobody could read is not "nothing installed" -- answering that would
// report a clean removal while a shared operator stays up. Both ways the read can fail stop it
// with the cause: two operators that cannot be told apart, and neither the cluster-wide list nor
// the read in the operator namespace answering. Nothing is changed either way.
func TestOperatorReleaseStopsWhenTheInstalledOperatorCannotBeRead(t *testing.T) {
	ambiguous := func() (*Cluster, *recRunner) {
		c, rr, _ := watchCluster("(none)", "")
		items := operatorItem("ns-a", "solace/op:1.4.2", "") + "," + operatorItem("ns-b", "solace/op:1.5.0", "")
		rr.out = []byte(`{"items":[` + items + `]}`)
		return c, rr
	}
	unreadable := func() (*Cluster, *recRunner) {
		c, rr, _ := deniedCluster("op-ns", nil)
		rr.outErrQueue = []error{errListDenied, errFake}
		return c, rr
	}
	for _, tc := range []struct {
		name    string
		cluster func() (*Cluster, *recRunner)
		cause   error
	}{
		{"two operators that cannot be told apart", ambiguous, errAmbiguousOperator},
		{"neither read answers", unreadable, errFake},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, rr := tc.cluster()
			err := c.OperatorRelease(context.Background(), true)
			if !errors.Is(err, tc.cause) {
				t.Fatalf("OperatorRelease err = %v, want it to carry %v", err, tc.cause)
			}
			if !strings.Contains(err.Error(), "reading the installed operator's watch scope") {
				t.Errorf("error = %v, want it to say what could not be read", err)
			}
			for _, call := range rr.calls {
				if call.method != "Output" {
					t.Errorf("an operator that could not be read must leave the cluster alone, got %+v", call)
				}
			}
		})
	}
}

// TestOperatorReleaseLeavesAnOperatorThatWatchesEverything: "all namespaces except this one" is not
// expressible, so an operator already watching everything cannot be narrowed -- and deleting it
// would take every other broker's operator with it. It is left running, the reason is given, and
// the command that would remove it anyway is spelled out for whoever does mean it. Both spellings
// of everything read the same: an empty value, and no variable at all.
func TestOperatorReleaseLeavesAnOperatorThatWatchesEverything(t *testing.T) {
	for _, tc := range []struct {
		name      string
		installed string
	}{
		{"an empty WATCH_NAMESPACE", ""},
		{"no WATCH_NAMESPACE variable", "(absent)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, rr, buf := watchCluster(tc.installed, "")
			if err := c.OperatorRelease(context.Background(), true); err != nil {
				t.Fatalf("OperatorRelease: %v\n%s", err, buf.String())
			}
			for _, call := range rr.calls {
				if call.method != "Output" {
					t.Errorf("an all-namespaces operator must be left alone, got %+v", call)
				}
			}
			byHand := "kubectl delete deployment " + operatorDeployment + " -n " + defaultOperatorNS
			out := buf.String()
			for _, want := range []string{"watches ALL namespaces", `exclude "solace"`, "operator kept", byHand} {
				if !strings.Contains(out, want) {
					t.Errorf("output should say %q:\n%s", want, out)
				}
			}
		})
	}
}

// TestOperatorReleaseDeletesTheOperatorWhenTheEnvFileCoversEveryNamespace: when this env file
// accounts for every namespace the operator watches, nobody else depends on it, so the removal is
// the full delete -- not a narrowing to an empty list, which would silently promote the operator
// to watching the whole cluster. The delete is OperatorDelete's own, so the CRD layer stays a
// separate step: kept by default, and reached only when asked for by name.
func TestOperatorReleaseDeletesTheOperatorWhenTheEnvFileCoversEveryNamespace(t *testing.T) {
	t.Run("the CRDs are kept", func(t *testing.T) {
		c, rr, buf := watchCluster("solace", "") // installed: exactly the broker namespace this env file names
		if err := c.OperatorRelease(context.Background(), false); err != nil {
			t.Fatalf("OperatorRelease: %v\n%s", err, buf.String())
		}
		var deletes []rrCall
		for _, call := range rr.calls {
			if len(call.args) > 1 && call.args[0] == "set" && call.args[1] == "env" {
				t.Errorf("the watch list was narrowed instead of the operator being removed: %v", call.args)
			}
			if call.method == "RunInput" {
				deletes = append(deletes, call)
			}
		}
		if len(deletes) != 1 || !eqArgs(deletes[0].args, []string{"delete", "-f", "-", "--ignore-not-found"}) {
			t.Fatalf("RunInput calls = %+v, want exactly one `delete -f - --ignore-not-found`", deletes)
		}
		if !strings.Contains(deletes[0].stdin, "kind: Deployment") ||
			strings.Contains(deletes[0].stdin, "kind: CustomResourceDefinition") {
			t.Errorf("the delete should carry the operator's Deployment and no CRD:\n%s", deletes[0].stdin)
		}
		why := "this env file accounts for every namespace the operator watches (solace); removing the operator"
		if !strings.Contains(buf.String(), why) {
			t.Errorf("the output should say why the whole operator is going:\n%s", buf.String())
		}
	})
	t.Run("--delete-crd reaches the delete", func(t *testing.T) {
		c, rr, buf := watchCluster("solace", "")
		// The first read is the installed operator; the second is the broker listing that guards the
		// CRDs before anything is deleted, and the repeat of it right before they go reads empty too.
		rr.out = nil
		rr.outQueue = [][]byte{deployJSON("solace"), []byte(`{"items":[]}`)}
		if err := c.OperatorRelease(context.Background(), true); err != nil {
			t.Fatalf("OperatorRelease: %v\n%s", err, buf.String())
		}
		var deletes []rrCall
		for _, call := range rr.calls {
			if call.method == "RunInput" {
				deletes = append(deletes, call)
			}
		}
		if len(deletes) != 2 || !strings.Contains(deletes[1].stdin, "kind: CustomResourceDefinition") {
			t.Fatalf("RunInput calls = %+v, want the operator delete and then the CRD delete", deletes)
		}
	})
}

// TestOperatorReleasePreviewShowsTheFullDelete: a preview has no cluster to answer, so the
// installed operator reads as nothing found -- and a removal that took that for "nothing to
// remove" would preview as a no-op. It shows the full delete instead, which is the path an
// operator previewing a removal means, and claims nothing about narrowing a watch list it could
// not read.
func TestOperatorReleasePreviewShowsTheFullDelete(t *testing.T) {
	var out bytes.Buffer
	log := func(f string, a ...any) { fmt.Fprintf(&out, f+"\n", a...) }
	c := NewCluster(engine.Echo{W: &out}, loadK8s(t), log, &out)
	if err := c.OperatorRelease(context.Background(), false); err != nil {
		t.Fatalf("OperatorRelease under Echo: %v\n%s", err, out.String())
	}
	got := out.String()
	if !strings.Contains(got, "+ kubectl delete -f - --ignore-not-found  <<< (") {
		t.Errorf("the preview should show the operator delete:\n%s", got)
	}
	for _, never := range []string{"set env", "nothing to remove"} {
		if strings.Contains(got, never) {
			t.Errorf("a preview has no installed operator to read, so it must not show %q:\n%s", never, got)
		}
	}
}

// TestOperatorReleaseNarrowingNeedsPermissionToPatch: narrowing the watch list is a patch on the
// operator's Deployment, asked about before it is attempted. Refused, the error is the
// preflight's own and the operator is left exactly as it was -- the `set env` is never issued, so
// another env file's namespace cannot be dropped by a half-finished removal.
func TestOperatorReleaseNarrowingNeedsPermissionToPatch(t *testing.T) {
	c, rr, _ := watchCluster("team-a,solace", "") // team-a still depends on it, so the list is narrowed
	rr.canI = "no"
	err := c.OperatorRelease(context.Background(), false)
	want := "not allowed to patch deployments in " + defaultOperatorNS
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("OperatorRelease err = %v, want it to say %q", err, want)
	}
	for _, call := range rr.calls {
		if call.method != "Output" {
			t.Errorf("a refused narrowing must change nothing, got %+v", call)
		}
	}
}

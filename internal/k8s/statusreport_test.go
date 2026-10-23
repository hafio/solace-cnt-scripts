package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"solace/internal/config"
	"solace/internal/engine"
)

// These replace the tests that pinned the old kubectl-passthrough survey. They
// assert the same properties -- the running picture, what --detail adds, and that
// one unreadable kind does not cost the rest of the report -- but against the
// rendering this tool now owns, driven by the objects captured from a live
// cluster (testdata/).

// fixtureRunner answers every `get ... -o json` from a captured fixture, and can
// be told to fail one kind so the degrade paths are reachable.
type fixtureRunner struct {
	calls   []string
	failOn  string
	missing map[string]bool // kinds that answer with an empty list
	// brokerFile overrides the pubsubpluseventbrokers fixture, for tests that
	// need a broker CR shape broker-cr.json does not carry (no conditions, a
	// failing one) rather than the healthy capture every other test shares.
	brokerFile string
	// raw answers a kind with this literal JSON instead of a fixture, for the shapes no
	// capture carries: a workload that shares the namespace, or two brokers in it.
	raw map[string]string
	// refuse answers `auth can-i` no for these resources and yes for every other, so a
	// section's failed read can be told apart as a refusal rather than a fault.
	refuse map[string]bool
}

func (f *fixtureRunner) Run(context.Context, string, ...string) error              { return nil }
func (f *fixtureRunner) RunInput(context.Context, []byte, string, ...string) error { return nil }
func (f *fixtureRunner) RunEnv(context.Context, []string, string, ...string) error { return nil }
func (f *fixtureRunner) RunInteractive(context.Context, string, ...string) error   { return nil }
func (f *fixtureRunner) OutputInput(context.Context, []byte, string, ...string) ([]byte, error) {
	return nil, nil
}

func (f *fixtureRunner) Output(_ context.Context, _ string, args ...string) ([]byte, error) {
	joined := strings.Join(args, " ")
	f.calls = append(f.calls, joined)
	for i, a := range args {
		if a == "can-i" && i+2 < len(args) {
			if f.refuse[args[i+2]] {
				return []byte("no\n"), nil
			}
			return []byte("yes\n"), nil
		}
	}
	file := map[string]string{
		"pubsubpluseventbrokers": "broker-cr.json",
		"pods":                   "pods.json",
		"svc":                    "services.json",
		"statefulsets":           "statefulsets.json",
		"pvc":                    "pvc.json",
		// The NAMED shape, because that is what OperatorReport asks for. Using
		// the list fixture here is what let the single-object bug through.
		"deployment": "operator-deploy-named.json",
	}
	if f.brokerFile != "" {
		file["pubsubpluseventbrokers"] = f.brokerFile
	}
	for _, a := range args {
		if f.failOn != "" && a == f.failOn {
			return nil, errFake
		}
		if f.missing[a] {
			return []byte(`{"items":[]}`), nil
		}
		if body, ok := f.raw[a]; ok {
			return []byte(body), nil
		}
		if name, ok := file[a]; ok {
			return os.ReadFile(filepath.Join("testdata", name))
		}
	}
	return []byte(`{"items":[]}`), nil
}

// reportCluster wires a Cluster whose reads come from the fixtures and whose
// clock is fixed 6m40s after the captured broker was created -- the same instant
// the live capture was taken, so the AGE column is reproducible.
func reportCluster(t *testing.T, f *fixtureRunner) (*Cluster, *bytes.Buffer) {
	t.Helper()
	var list brokerList
	raw, err := os.ReadFile(filepath.Join("testdata", "broker-cr.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &list); err != nil {
		t.Fatal(err)
	}
	created := list.Items[0].Metadata.CreationTimestamp

	cfg := &config.Config{}
	cfg.K8s.Namespace = "solace"
	cfg.K8s.Name = "dev-broker"
	cfg.K8s.Operator.Namespace = "solace-operator"
	cfg.K8s.Command = config.Command{"kubectl"}

	buf := &bytes.Buffer{}
	c := NewCluster(f, cfg, nil, buf)
	c.Now = func() time.Time { return created.Add(6*time.Minute + 40*time.Second) }
	return c, buf
}

// TestBrokerReportRunningPicture is the successor to TestShowAll: the sections a
// healthy broker reports, the operator's own verdict at the top, and the
// redundancy ORDER that kubectl's alphabetical listing cannot give.
func TestBrokerReportRunningPicture(t *testing.T) {
	c, buf := reportCluster(t, &fixtureRunner{})
	if err := c.BrokerReport(context.Background(), false); err != nil {
		t.Fatalf("BrokerReport: %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		"=== Broker", "=== Pods", "=== Services", "=== StatefulSets",
		"dev-broker", "HA redundancy group",
		"[ OK ] service", "[ OK ] ha group", "[ OK ] warnings",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing %q:\n%s", want, out)
		}
	}
	// The pending LoadBalancer is the one warning a healthy deployment still
	// carries, and it is the difference between deployed and reachable.
	if !strings.Contains(out, "[WARN] dev-broker-pubsubplus ") {
		t.Errorf("the unassigned LoadBalancer address should warn:\n%s", out)
	}
	// Redundancy order: primary before backup before monitor.
	p := strings.Index(out, "pubsubplus-p-0")
	b := strings.Index(out, "pubsubplus-b-0")
	m := strings.Index(out, "pubsubplus-m-0")
	if !(p >= 0 && p < b && b < m) {
		t.Errorf("pods should read primary, backup, monitor (got offsets %d/%d/%d):\n%s", p, b, m, out)
	}
	// Storage is --detail only, so the default report must not carry it.
	if strings.Contains(out, "=== Storage") {
		t.Errorf("the default report should not include Storage:\n%s", out)
	}
	// No ROLE column: the name already ends in -p-0/-b-0/-m-0.
	if strings.Contains(out, "ROLE") {
		t.Errorf("the pods table should carry no ROLE column:\n%s", out)
	}
}

// TestBrokerReportDetailAddsStorageAndPlacement is the successor to
// TestShowAllDetailAddsStaticArtifacts: --detail adds the claims a broker's data
// lives on, plus the IP/NODE columns that show whether anti-affinity spread it.
func TestBrokerReportDetailAddsStorageAndPlacement(t *testing.T) {
	c, buf := reportCluster(t, &fixtureRunner{})
	if err := c.BrokerReport(context.Background(), true); err != nil {
		t.Fatalf("BrokerReport --detail: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"=== Storage", "Bound", "sol-standard", "30Gi", "5Gi", "IP", "NODE"} {
		if !strings.Contains(out, want) {
			t.Errorf("--detail report missing %q:\n%s", want, out)
		}
	}
	// The three brokers really were on three different nodes; that is the fact
	// NODE exists to show.
	for _, node := range []string{"mk", "mk-m02", "mk-m03"} {
		if !strings.Contains(out, node) {
			t.Errorf("--detail should name the node each pod landed on, missing %q:\n%s", node, out)
		}
	}
}

// TestBrokerReportWithNoBrokerDeployed covers the empty case: it must say so as a
// warning rather than printing an empty section that reads like a pass.
func TestBrokerReportWithNoBrokerDeployed(t *testing.T) {
	f := &fixtureRunner{missing: map[string]bool{
		"pubsubpluseventbrokers": true, "pods": true, "svc": true, "statefulsets": true,
	}}
	c, buf := reportCluster(t, f)
	if err := c.BrokerReport(context.Background(), false); err != nil {
		t.Fatalf("BrokerReport: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"no PubSubPlusEventBroker", "[WARN]"} {
		if !strings.Contains(out, want) {
			t.Errorf("an absent broker should be reported as a warning, missing %q:\n%s", want, out)
		}
	}
}

// TestBrokerReportSurfacesAReadFailure is the successor to
// TestShowAllReportsAndContinuesOnGetError. The old passthrough continued past a
// kind it could not list; this report fails loud instead, and deliberately: a
// health report that quietly omitted the pods would read as "no problems found". The
// one exception is a read the cluster refused -- TestBrokerReportShowsARefusedSection.
func TestBrokerReportSurfacesAReadFailure(t *testing.T) {
	c, _ := reportCluster(t, &fixtureRunner{failOn: "pods"})
	err := c.BrokerReport(context.Background(), false)
	if err == nil {
		t.Fatal("a failed pod read must surface, not be silently skipped")
	}
	if !strings.Contains(err.Error(), "reading pods") || !strings.Contains(err.Error(), "solace") {
		t.Errorf("error = %v, want it to name the read and the namespace", err)
	}
}

// TestBrokerReportShowsARefusedSection: an identity granted the broker but not its pods is
// not a fault, and a report that stopped there would hide the services and StatefulSets it
// could read. So a read the cluster refused draws the section with one [SKIP] row naming the
// grant, every later section still renders, and the report succeeds -- the section visibly
// missing rather than quietly absent.
func TestBrokerReportShowsARefusedSection(t *testing.T) {
	c, buf := reportCluster(t, &fixtureRunner{failOn: "pods", refuse: map[string]bool{"pods": true}})
	if err := c.BrokerReport(context.Background(), false); err != nil {
		t.Fatalf("a refused section must not fail the report: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"=== Pods", "[SKIP] pods", "may not get pods in solace", "=== Services", "=== StatefulSets"} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing %q:\n%s", want, out)
		}
	}
}

// TestBrokerReportUsesTheCRsOwnPodList proves membership comes from the operator's
// podsList rather than a name-substring filter, so a co-tenant workload whose name
// happens to carry the infix cannot be counted as a broker pod.
func TestBrokerReportUsesTheCRsOwnPodList(t *testing.T) {
	c, buf := reportCluster(t, &fixtureRunner{})
	if err := c.BrokerReport(context.Background(), false); err != nil {
		t.Fatalf("BrokerReport: %v", err)
	}
	// Exactly the three pods the CR claims, no more.
	if n := strings.Count(buf.String(), "dev-broker-pubsubplus-"); n < 3 {
		t.Errorf("expected the three pods from status.podsList:\n%s", buf.String())
	}
}

// TestOperatorReportNamesTheRunningVersion is the point of reading the deployment
// rather than assuming: the cluster can run a different operator from the one this
// tool renders, and the report has to show which one actually answered.
func TestOperatorReportNamesTheRunningVersion(t *testing.T) {
	c, buf := reportCluster(t, &fixtureRunner{})
	if err := c.OperatorReport(context.Background()); err != nil {
		t.Fatalf("OperatorReport: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"=== Operator", "solace-operator", "[ OK ] controller", "1/1 ready", "1.4.2"} {
		if !strings.Contains(out, want) {
			t.Errorf("operator report missing %q:\n%s", want, out)
		}
	}
}

// TestOperatorReportWhenTheOperatorCannotBeRead: a read that FAILS is not an operator that
// is missing. An identity confined to the broker namespace most often may not read the
// operator's at all, and telling it to install an operator that is already running would
// send it the wrong way. So the row says not read, names the grant that reads it, and
// advises no install -- still a report line, not an error.
func TestOperatorReportWhenTheOperatorCannotBeRead(t *testing.T) {
	c, buf := reportCluster(t, &fixtureRunner{failOn: "deployment"})
	if err := c.OperatorReport(context.Background()); err != nil {
		t.Fatalf("an unreadable operator is a report line, not an error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"[SKIP] installed", "not read", "needs get on deployments in solace-operator"} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "installs it") {
		t.Errorf("an unreadable operator must not be reported as one to install:\n%s", out)
	}
}

// TestClusterReportListsEveryBroker is `--all`: one row per broker across every
// namespace, carrying the worst of its conditions so a cluster-wide listing
// surfaces the one needing attention instead of averaging it away.
func TestClusterReportListsEveryBroker(t *testing.T) {
	c, buf := reportCluster(t, &fixtureRunner{})
	if err := c.ClusterReport(context.Background(), false); err != nil {
		t.Fatalf("ClusterReport: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"=== Brokers in this cluster", "NAMESPACE", "solace", "dev-broker", "HA", "[ OK ]", "ready"} {
		if !strings.Contains(out, want) {
			t.Errorf("cluster report missing %q:\n%s", want, out)
		}
	}
}

// TestClusterReportUnreportedBrokerIsInfo is M8: a broker the operator has
// published NO conditions for -- freshly deployed, or an operator wedged
// before writing any status -- must not render identically to a verified
// healthy one. OK must mean a condition actually said so.
func TestClusterReportUnreportedBrokerIsInfo(t *testing.T) {
	c, buf := reportCluster(t, &fixtureRunner{brokerFile: "broker-cr-no-conditions.json"})
	if err := c.ClusterReport(context.Background(), false); err != nil {
		t.Fatalf("ClusterReport: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "[ OK ]") {
		t.Errorf("an unreported broker must not render OK:\n%s", out)
	}
	if !strings.Contains(out, "[INFO]") || !strings.Contains(out, "not yet reported by the operator") {
		t.Errorf("an unreported broker should read INFO/not yet reported:\n%s", out)
	}
}

// TestClusterReportStillSurfacesAFailingCondition guards the M8 fix from
// overcorrecting: a broker that DOES have conditions, one of them failing,
// must still show the failure rather than falling back to the unreported
// state or the OK seed.
func TestClusterReportStillSurfacesAFailingCondition(t *testing.T) {
	c, buf := reportCluster(t, &fixtureRunner{brokerFile: "broker-cr-failing-condition.json"})
	if err := c.ClusterReport(context.Background(), false); err != nil {
		t.Fatalf("ClusterReport: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[FAIL]") || !strings.Contains(out, "service: No active broker pod found") {
		t.Errorf("a broker with a failing condition should surface it:\n%s", out)
	}
}

func TestClusterReportWithNoBrokers(t *testing.T) {
	f := &fixtureRunner{missing: map[string]bool{"pubsubpluseventbrokers": true}}
	c, buf := reportCluster(t, f)
	if err := c.ClusterReport(context.Background(), false); err != nil {
		t.Fatalf("ClusterReport: %v", err)
	}
	if !strings.Contains(buf.String(), "none found in any namespace") {
		t.Errorf("an empty cluster should say so:\n%s", buf.String())
	}
}

// TestStatusReportsArePreviewableUnderEcho keeps the property every other command
// has: with the Echo runner there is no cluster to decode, so the reads are echoed
// and nothing is claimed about state.
func TestStatusReportsArePreviewableUnderEcho(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := &config.Config{}
	cfg.K8s.Namespace = "solace"
	cfg.K8s.Operator.Namespace = "solace-operator"
	cfg.K8s.Command = config.Command{"kubectl"}
	c := NewCluster(engine.Echo{W: buf}, cfg, nil, buf)

	for name, fn := range map[string]func() error{
		"broker":  func() error { return c.BrokerReport(context.Background(), true) },
		"cluster": func() error { return c.ClusterReport(context.Background(), false) },
		"operator": func() error {
			return c.OperatorReport(context.Background())
		},
	} {
		buf.Reset()
		if err := fn(); err != nil {
			t.Errorf("%s report under Echo: %v", name, err)
		}
		if !strings.Contains(buf.String(), "+ kubectl get") {
			t.Errorf("%s report should echo its reads:\n%s", name, buf.String())
		}
		if strings.Contains(buf.String(), "[ OK ]") {
			t.Errorf("%s report must not claim health with no cluster to read:\n%s", name, buf.String())
		}
	}
}

// --- degrade paths -------------------------------------------------------------

// TestReportsDegradeRatherThanFalselyAlarm covers what the "own the rendering"
// decision costs: every branch that fires when the cluster shows something this
// code did not expect. None of them may produce a false alarm or a false pass.
func TestReportsDegradeRatherThanFalselyAlarm(t *testing.T) {
	t.Run("no pods for a deployed broker warns", func(t *testing.T) {
		c, buf := reportCluster(t, &fixtureRunner{missing: map[string]bool{"pods": true}})
		if err := c.BrokerReport(context.Background(), false); err != nil {
			t.Fatalf("BrokerReport: %v", err)
		}
		if !strings.Contains(buf.String(), "[WARN] pods") {
			t.Errorf("a broker with no pods should warn:\n%s", buf.String())
		}
	})
	t.Run("no services warns", func(t *testing.T) {
		c, buf := reportCluster(t, &fixtureRunner{missing: map[string]bool{"svc": true}})
		if err := c.BrokerReport(context.Background(), false); err != nil {
			t.Fatalf("BrokerReport: %v", err)
		}
		if !strings.Contains(buf.String(), "[WARN] services") {
			t.Errorf("no services should warn:\n%s", buf.String())
		}
	})
	t.Run("no statefulsets warns", func(t *testing.T) {
		c, buf := reportCluster(t, &fixtureRunner{missing: map[string]bool{"statefulsets": true}})
		if err := c.BrokerReport(context.Background(), false); err != nil {
			t.Fatalf("BrokerReport: %v", err)
		}
		if !strings.Contains(buf.String(), "[WARN] statefulsets") {
			t.Errorf("no statefulsets should warn:\n%s", buf.String())
		}
	})
	t.Run("no claims is INFO, not a warning", func(t *testing.T) {
		c, buf := reportCluster(t, &fixtureRunner{missing: map[string]bool{"pvc": true}})
		if err := c.BrokerReport(context.Background(), true); err != nil {
			t.Fatalf("BrokerReport --detail: %v", err)
		}
		// A broker can legitimately run without persistent claims, so this is a
		// fact rather than a problem.
		if !strings.Contains(buf.String(), "[INFO] claims") {
			t.Errorf("no claims should be INFO:\n%s", buf.String())
		}
	})
	for _, tc := range []struct{ kind, resource, wantErr, wantRow string }{
		{"svc", "services", "reading services", "[SKIP] services"},
		{"statefulsets", "statefulsets", "reading statefulsets", "[SKIP] statefulsets"},
		{"pvc", "persistentvolumeclaims", "reading persistent volume claims", "[SKIP] claims"},
	} {
		t.Run("a failed "+tc.kind+" read surfaces", func(t *testing.T) {
			c, _ := reportCluster(t, &fixtureRunner{failOn: tc.kind})
			err := c.BrokerReport(context.Background(), true)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("err = %v, want it to name %q", err, tc.wantErr)
			}
		})
		t.Run("a refused "+tc.kind+" read is shown, not fatal", func(t *testing.T) {
			c, buf := reportCluster(t, &fixtureRunner{failOn: tc.kind, refuse: map[string]bool{tc.resource: true}})
			if err := c.BrokerReport(context.Background(), true); err != nil {
				t.Fatalf("BrokerReport --detail: %v", err)
			}
			if !strings.Contains(buf.String(), tc.wantRow) || !strings.Contains(buf.String(), "may not get "+tc.resource) {
				t.Errorf("report should carry %q naming the grant:\n%s", tc.wantRow, buf.String())
			}
		})
	}
	t.Run("a failed broker read surfaces", func(t *testing.T) {
		c, _ := reportCluster(t, &fixtureRunner{failOn: "pubsubpluseventbrokers"})
		err := c.BrokerReport(context.Background(), false)
		if err == nil || !strings.Contains(err.Error(), "needs get on "+brokerResource) {
			t.Errorf("err = %v: a failed CR read ends the report, naming what reading it needs", err)
		}
	})
	t.Run("a failed cluster-wide read surfaces", func(t *testing.T) {
		c, _ := reportCluster(t, &fixtureRunner{failOn: "pubsubpluseventbrokers"})
		err := c.ClusterReport(context.Background(), false)
		for _, want := range []string{"listing brokers across the cluster", "cluster role binding", "without --all"} {
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("err = %v: a failed --all read fails the command, and should say %q", err, want)
			}
		}
	})
	t.Run("malformed JSON names the command to run by hand", func(t *testing.T) {
		c, _ := reportCluster(t, &fixtureRunner{})
		var list brokerList
		err := c.getJSON(context.Background(), &list, "garbage-kind")
		_ = err // the fixture runner answers valid JSON; the shape below is the check
		if e := (&Cluster{R: &badJSONRunner{}, Cfg: c.Cfg}).getJSON(
			context.Background(), &list, "pods", "-n", "solace"); e == nil ||
			!strings.Contains(e.Error(), "kubectl get") {
			t.Errorf("err = %v, want it to name the command to run by hand", e)
		}
	})
}

// badJSONRunner answers with bytes that are not JSON, for the decode-failure path.
type badJSONRunner struct{ fixtureRunner }

func (*badJSONRunner) Output(context.Context, string, ...string) ([]byte, error) {
	return []byte("not json at all"), nil
}

// TestOwnsPodFallsBackToTheNameInfix covers the branch taken when the operator
// publishes no podsList: the same filter the passthrough used, which excludes the
// operator's own pod because that name carries no leading dash.
func TestOwnsPodFallsBackToTheNameInfix(t *testing.T) {
	for name, want := range map[string]bool{
		"dev-broker-pubsubplus-p-0":           true,
		"pubsubplus-eventbroker-operator-abc": false,
		"unrelated-workload":                  false,
	} {
		if got := ownsPod(nil, name); got != want {
			t.Errorf("ownsPod(nil, %q) = %v, want %v", name, got, want)
		}
	}
	// With a list, only what the CR claims counts.
	own := map[string]bool{"dev-broker-pubsubplus-p-0": true}
	if ownsPod(own, "dev-broker-pubsubplus-b-0") {
		t.Error("a pod absent from the CR's list must not be counted")
	}
}

func TestOperatorRunningImageWithNoContainers(t *testing.T) {
	if got := operatorRunningImage(nil); !strings.Contains(got, "no container") {
		t.Errorf("operatorRunningImage(nil) = %q, want it to say so rather than be blank", got)
	}
}

// TestNowDefaultsToTheWallClock covers the seam's default branch.
func TestNowDefaultsToTheWallClock(t *testing.T) {
	c := &Cluster{}
	if c.now().IsZero() {
		t.Error("an unset Now must fall back to the wall clock, not the zero time")
	}
}

// TestBrokerReportReadsEachKindOnce pins the read budget. The broker CR is needed
// twice over -- once for the operator's conditions, once for status.podsList -- and
// the obvious way to write that is two fetches. One report should cost one read per
// kind, so this fails if a refactor reintroduces the duplicate.
func TestBrokerReportReadsEachKindOnce(t *testing.T) {
	f := &fixtureRunner{}
	c, _ := reportCluster(t, f)
	if err := c.BrokerReport(context.Background(), true); err != nil {
		t.Fatalf("BrokerReport --detail: %v", err)
	}
	counts := map[string]int{}
	for _, call := range f.calls {
		for _, kind := range []string{"pubsubpluseventbrokers", "pods", "svc", "statefulsets", "pvc"} {
			if strings.Contains(call, "get "+kind+" ") {
				counts[kind]++
			}
		}
	}
	for kind, n := range counts {
		if n != 1 {
			t.Errorf("%s read %d times, want exactly 1:\n%v", kind, n, f.calls)
		}
	}
	if len(counts) != 5 {
		t.Errorf("read %d kinds, want 5 (broker, pods, svc, statefulsets, pvc): %v", len(counts), f.calls)
	}
	// Without --detail the claims are not read at all.
	f2 := &fixtureRunner{}
	c2, _ := reportCluster(t, f2)
	if err := c2.BrokerReport(context.Background(), false); err != nil {
		t.Fatalf("BrokerReport: %v", err)
	}
	for _, call := range f2.calls {
		if strings.Contains(call, "get pvc") {
			t.Errorf("the default report should not read claims: %v", f2.calls)
		}
	}
}

// TestBrokerReportReadsOnlyWhatTheEnvFileNames: every read names its objects -- this env
// file's broker CR, then one pod, StatefulSet and claim per role and the load-balancer and
// discovery services, all from the env file -- so the report needs only `get` on each kind,
// where listing the namespace needed `list`, and a namespace shared with other workloads
// contributes nothing of theirs. --ignore-not-found makes a missing object an absence the
// section reports rather than a failed read.
func TestBrokerReportReadsOnlyWhatTheEnvFileNames(t *testing.T) {
	f := &fixtureRunner{}
	c, _ := reportCluster(t, f)
	if err := c.BrokerReport(context.Background(), true); err != nil {
		t.Fatalf("BrokerReport --detail: %v", err)
	}
	const tail = " -n solace --ignore-not-found -o json"
	want := []string{
		"get pubsubpluseventbrokers dev-broker" + tail,
		"get pods dev-broker-pubsubplus-p-0" + tail,
		"get svc dev-broker-pubsubplus dev-broker-pubsubplus-discovery" + tail,
		"get statefulsets dev-broker-pubsubplus-p" + tail,
		"get pvc data-dev-broker-pubsubplus-p-0" + tail,
	}
	if strings.Join(f.calls, "\n") != strings.Join(want, "\n") {
		t.Errorf("reads =\n%s\nwant\n%s", strings.Join(f.calls, "\n"), strings.Join(want, "\n"))
	}
}

// TestBrokerReportOmitsWorkloadsThatAreNotTheBrokers: a namespace is often shared, so every list
// the report reads holds objects that are not this broker's, and none of them may appear as one.
// A pod is the broker's own only when the operator's podsList claims it -- a workload whose name
// merely carries the `-pubsubplus-` infix is still not a broker pod -- while services,
// StatefulSets and claims are told by name. Each section here is listed with an impostor beside
// the broker's own object.
func TestBrokerReportOmitsWorkloadsThatAreNotTheBrokers(t *testing.T) {
	f := &fixtureRunner{raw: map[string]string{
		"pods": `{"items":[
			{"metadata":{"name":"dev-broker-pubsubplus-p-0"},"status":{"phase":"Running","containerStatuses":[{"ready":true}]}},
			{"metadata":{"name":"cache-pubsubplus-proxy-0"},"status":{"phase":"Running"}},
			{"metadata":{"name":"metrics-exporter-0"},"status":{"phase":"Running"}}
		]}`,
		"svc": `{"items":[
			{"metadata":{"name":"dev-broker-pubsubplus"},"spec":{"type":"ClusterIP","clusterIP":"10.0.0.1"}},
			{"metadata":{"name":"metrics-exporter"},"spec":{"type":"ClusterIP","clusterIP":"10.0.0.2"}}
		]}`,
		"statefulsets": `{"items":[
			{"metadata":{"name":"dev-broker-pubsubplus-p"},"spec":{"replicas":1},"status":{"readyReplicas":1}},
			{"metadata":{"name":"metrics-exporter"},"spec":{"replicas":1},"status":{"readyReplicas":1}}
		]}`,
		"pvc": `{"items":[
			{"metadata":{"name":"data-dev-broker-pubsubplus-p-0"},"status":{"phase":"Bound"}},
			{"metadata":{"name":"data-metrics-exporter-0"},"status":{"phase":"Bound"}}
		]}`,
	}}
	c, buf := reportCluster(t, f)
	if err := c.BrokerReport(context.Background(), true); err != nil {
		t.Fatalf("BrokerReport --detail: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"=== Pods", "=== Services", "=== StatefulSets", "=== Storage"} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing the section %q:\n%s", want, out)
		}
	}
	for _, want := range []string{"dev-broker-pubsubplus-p-0", "data-dev-broker-pubsubplus-p-0"} {
		if !strings.Contains(out, want) {
			t.Errorf("the broker's own object %q is missing from the report:\n%s", want, out)
		}
	}
	for _, never := range []string{"cache-pubsubplus-proxy-0", "metrics-exporter"} {
		if strings.Contains(out, never) {
			t.Errorf("an object that is not this broker's appeared in the report (%q):\n%s", never, out)
		}
	}
}

// TestBrokerReportOrdersPodsOfTheSameRoleByName: two broker resources can share a namespace, each
// with its own primary, and role order alone cannot sort them. The tie is broken by name, so the
// listing is the same on every run whatever order the cluster happened to return the pods in --
// a report that reshuffled between runs would read as something having changed.
func TestBrokerReportOrdersPodsOfTheSameRoleByName(t *testing.T) {
	f := &fixtureRunner{raw: map[string]string{
		"pubsubpluseventbrokers": `{"items":[
			{"metadata":{"name":"qa-broker","namespace":"solace"},"status":{"podsList":["qa-broker-pubsubplus-p-0"]}},
			{"metadata":{"name":"dev-broker","namespace":"solace"},"status":{"podsList":["dev-broker-pubsubplus-p-0"]}}
		]}`,
		"pods": `{"items":[
			{"metadata":{"name":"qa-broker-pubsubplus-p-0"},"status":{"phase":"Running","containerStatuses":[{"ready":true}]}},
			{"metadata":{"name":"dev-broker-pubsubplus-p-0"},"status":{"phase":"Running","containerStatuses":[{"ready":true}]}}
		]}`,
	}}
	c, buf := reportCluster(t, f)
	if err := c.BrokerReport(context.Background(), false); err != nil {
		t.Fatalf("BrokerReport: %v", err)
	}
	out := buf.String()
	dev, qa := strings.Index(out, "dev-broker-pubsubplus-p-0"), strings.Index(out, "qa-broker-pubsubplus-p-0")
	if dev < 0 || qa < 0 || dev > qa {
		t.Errorf("pods of one role should read in name order (dev at %d, qa at %d):\n%s", dev, qa, out)
	}
}

// TestOperatorReportWhenTheDeploymentListIsEmpty: the read succeeding with nothing in it is a
// different answer from the read failing -- the namespace was reachable and holds no operator --
// and it is reported as such, in the namespace that was looked in, rather than as an error or as
// a section with no rows that reads like a pass.
func TestOperatorReportWhenTheDeploymentListIsEmpty(t *testing.T) {
	c, buf := reportCluster(t, &fixtureRunner{missing: map[string]bool{"deployment": true}})
	if err := c.OperatorReport(context.Background()); err != nil {
		t.Fatalf("an operator that is not there is a report line, not an error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"[WARN] installed", "not found in namespace solace-operator", "operator deploy installs it"} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "controller") {
		t.Errorf("no controller row may be shown for an operator that is not there:\n%s", out)
	}
}

// TestBrokerReportPreviewStopsAtAnUnapprovedCommand: a preview echoes the reads it would make, and
// the execution guard still applies to them -- an env file naming a command nobody approved
// fails the preview with the guard's own message and echoes nothing, rather than printing a
// `+ microk8s kubectl get ...` line that reads as something this tool would have run.
func TestBrokerReportPreviewStopsAtAnUnapprovedCommand(t *testing.T) {
	var buf bytes.Buffer
	c := NewCluster(engine.Echo{W: &buf}, unapprovedCfg(), nil, &buf)
	err := c.BrokerReport(context.Background(), true)
	if err == nil || !strings.Contains(err.Error(), "--allow-command microk8s") {
		t.Fatalf("BrokerReport under Echo err = %v, want the guard to name the escape hatch", err)
	}
	if buf.Len() != 0 {
		t.Errorf("a refused command must echo nothing, got:\n%s", buf.String())
	}
}

// TestClusterReportDetailAddsTheBrokerImage: `status broker --all --detail` adds the broker image,
// the column that tells apart brokers an operator sees side by side in different namespaces. It
// sits ahead of STATE so the verdict still ends the line, and the default listing carries neither
// the column nor the value.
func TestClusterReportDetailAddsTheBrokerImage(t *testing.T) {
	const image = "registry.example.com/solace/solace-pubsub-enterprise:10.26.0.8755" // broker-cr.json's brokerImage
	c, buf := reportCluster(t, &fixtureRunner{})
	if err := c.ClusterReport(context.Background(), true); err != nil {
		t.Fatalf("ClusterReport --detail: %v", err)
	}
	var header, row string
	for _, line := range strings.Split(buf.String(), "\n") {
		switch {
		case strings.Contains(line, "NAMESPACE"):
			header = line
		case strings.Contains(line, "dev-broker"):
			row = line
		}
	}
	if !strings.Contains(header, "IMAGE") || strings.Index(header, "IMAGE") > strings.Index(header, "STATE") {
		t.Errorf("--detail header = %q, want an IMAGE column ahead of STATE", header)
	}
	if !strings.Contains(row, image) || strings.Index(row, image) > strings.Index(row, "ready") {
		t.Errorf("--detail row = %q, want the broker image ahead of its state", row)
	}

	plain, plainBuf := reportCluster(t, &fixtureRunner{})
	if err := plain.ClusterReport(context.Background(), false); err != nil {
		t.Fatalf("ClusterReport: %v", err)
	}
	if got := plainBuf.String(); strings.Contains(got, "IMAGE") || strings.Contains(got, image) {
		t.Errorf("the default listing must not carry the image:\n%s", got)
	}
}

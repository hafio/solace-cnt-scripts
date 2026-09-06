package k8s

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"solace/internal/output"
)

// The fixtures under testdata/ are REAL objects, captured from a live minikube
// deployment on 2026-08-21 (operator 1.4.2, broker 10.26.0.8755) and trimmed to
// the fields these decoders read; only the private registry host was generalised.
// That matters: the shapes here are not what the schema documentation suggests --
// the broker CR publishes ServiceReady/NoWarnings/HAReady rather than a
// conventional "Ready", and its booleans are strings. Decoding was written against
// these bytes, so a schema change shows up as a failing test rather than as a
// wrong report.

// loadFixture decodes a captured object into v.
func loadFixture(t *testing.T, name string, v any) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
}

// TestDecodeBrokerCRFromLiveCapture is the test that would have caught the guess:
// the operator publishes no "Ready" condition, so anything looking for one finds
// nothing. It pins the three types this operator really publishes, the string
// booleans, and podsList.
func TestDecodeBrokerCRFromLiveCapture(t *testing.T) {
	var list brokerList
	loadFixture(t, "broker-cr.json", &list)
	if len(list.Items) != 1 {
		t.Fatalf("decoded %d brokers, want 1", len(list.Items))
	}
	got := list.Items[0]

	if got.Metadata.Name != "dev-broker" || got.Metadata.Namespace != "solace" {
		t.Errorf("metadata = %+v", got.Metadata)
	}
	if got.Metadata.CreationTimestamp.IsZero() {
		t.Error("creationTimestamp did not decode, so the AGE column would read '-'")
	}

	// The three types this operator actually publishes.
	for _, want := range []string{condServiceReady, condNoWarnings, condHAReady} {
		if _, ok := findCondition(got.Status.Conditions, want); !ok {
			t.Errorf("condition %q missing from the live capture", want)
		}
	}
	// And the one a reasonable person would have guessed at, which is NOT there.
	if _, ok := findCondition(got.Status.Conditions, "Ready"); ok {
		t.Error("the capture now has a plain Ready condition; conditionLevel should prefer it")
	}

	// haDeployment/tlsSupport are STRINGS in this schema. A bool field would have
	// failed to decode and silently read false.
	if got.Status.Broker.HADeployment != "true" {
		t.Errorf("haDeployment = %q, want the string \"true\"", got.Status.Broker.HADeployment)
	}
	if got.Status.Broker.ServiceName != "dev-broker-pubsubplus" ||
		got.Status.Broker.ServiceType != "LoadBalancer" {
		t.Errorf("status.broker = %+v", got.Status.Broker)
	}
	if len(got.Status.Broker.StatefulSets) != 3 {
		t.Errorf("statefulSets = %v, want the three role sets", got.Status.Broker.StatefulSets)
	}
	if len(got.Status.PodsList) != 3 {
		t.Errorf("podsList = %v, want three pods", got.Status.PodsList)
	}
}

// TestBrokerConditionRowsFromLiveCapture proves a healthy live broker reads as
// healthy -- the regression the original guess would have produced.
func TestBrokerConditionRowsFromLiveCapture(t *testing.T) {
	var list brokerList
	loadFixture(t, "broker-cr.json", &list)
	conds := list.Items[0].Status.Conditions

	rows := brokerConditionRows(conds, true)
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want service/ha group/warnings", len(rows))
	}
	for _, r := range rows {
		if r.Level != output.LevelOK {
			t.Errorf("row %q = level %v (%q), want LevelOK for a healthy broker", r.Key, r.Level, r.Value)
		}
		if r.Value == "" {
			t.Errorf("row %q carries no message; the operator's own text is the useful part", r.Key)
		}
	}

	// Standalone has no redundancy group, so no HA row is claimed either way.
	for _, r := range brokerConditionRows(conds, false) {
		if r.Key == "ha group" {
			t.Error("standalone must not report an HA row")
		}
	}
}

// TestConditionLevelDegradesSafely covers the cases a future operator version can
// produce: a condition it stopped publishing, and a False one. Neither may become
// a false pass, and only a genuine False may become a failure.
func TestConditionLevelDegradesSafely(t *testing.T) {
	t.Run("absent condition claims nothing", func(t *testing.T) {
		lvl, msg := conditionLevel(nil, condHAReady)
		if lvl != output.LevelInfo {
			t.Errorf("level = %v, want LevelInfo -- an unpublished condition is not a pass or a failure", lvl)
		}
		if msg == "" {
			t.Error("the row should say the operator did not report it")
		}
	})
	t.Run("False is a failure and carries the operator message", func(t *testing.T) {
		conds := []condition{{Type: condHAReady, Status: "False", Message: "node b not HA ready"}}
		lvl, msg := conditionLevel(conds, condHAReady)
		if lvl != output.LevelFail || msg != "node b not HA ready" {
			t.Errorf("level=%v msg=%q, want LevelFail with the operator's message", lvl, msg)
		}
	})
	t.Run("NoWarnings inverts to a warning, not a failure", func(t *testing.T) {
		conds := []condition{{Type: condNoWarnings, Status: "False", Message: "disk usage high"}}
		lvl, _ := conditionLevel(conds, condNoWarnings)
		if lvl != output.LevelWarn {
			t.Errorf("level = %v, want LevelWarn: the broker complaining is not the deployment failing", lvl)
		}
	})
	t.Run("an unknown status value claims nothing", func(t *testing.T) {
		conds := []condition{{Type: condHAReady, Status: "Unknown", Message: "still settling"}}
		lvl, _ := conditionLevel(conds, condHAReady)
		if lvl != output.LevelInfo {
			t.Errorf("level = %v, want LevelInfo for a status that is neither True nor False", lvl)
		}
	})
	t.Run("reason stands in when there is no message", func(t *testing.T) {
		conds := []condition{{Type: condServiceReady, Status: "True", Reason: "ActivePodAndServiceExists"}}
		_, msg := conditionLevel(conds, condServiceReady)
		if msg != "ActivePodAndServiceExists" {
			t.Errorf("msg = %q, want the reason as a fallback", msg)
		}
	})
}

// TestDecodePodsFromLiveCapture pins the pod shape, notably that a broker pod runs
// ONE container -- so READY is 1/1, not the 3/3 a multi-container guess assumed.
func TestDecodePodsFromLiveCapture(t *testing.T) {
	var list podList
	loadFixture(t, "pods.json", &list)
	if len(list.Items) != 3 {
		t.Fatalf("decoded %d pods, want 3", len(list.Items))
	}
	for i := range list.Items {
		ready, total, restarts := list.at(i)
		name := list.Items[i].Metadata.Name
		if total != 1 {
			t.Errorf("%s has %d containers, want 1 (a broker pod is single-container)", name, total)
		}
		if ready != total || restarts != 0 {
			t.Errorf("%s: ready=%d/%d restarts=%d, want fully ready and unrestarted", name, ready, total, restarts)
		}
		if lvl := podHealth(list.Items[i].Status.Phase, ready, total, restarts); lvl != output.LevelOK {
			t.Errorf("%s classified %v, want LevelOK", name, lvl)
		}
		if roleRank(name) == 3 {
			t.Errorf("%s: role not recognised from the pod name, so it would sort last", name)
		}
	}
}

func TestPodHealthRules(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		phase                  string
		ready, total, restarts int
		want                   output.Level
	}{
		{"running and ready", "Running", 1, 1, 0, output.LevelOK},
		{"ready but has restarted", "Running", 1, 1, 4, output.LevelWarn},
		{"pending", "Pending", 0, 1, 0, output.LevelWarn},
		{"running, not all ready", "Running", 0, 1, 0, output.LevelWarn},
		{"failed", "Failed", 0, 1, 0, output.LevelFail},
		{"unknown", "Unknown", 0, 1, 0, output.LevelFail},
		{"completed is not a broker problem", "Succeeded", 0, 1, 0, output.LevelInfo},
		{"no container statuses yet", "Running", 0, 0, 0, output.LevelWarn},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := podHealth(tc.phase, tc.ready, tc.total, tc.restarts); got != tc.want {
				t.Errorf("podHealth(%q,%d,%d,%d) = %v, want %v",
					tc.phase, tc.ready, tc.total, tc.restarts, got, tc.want)
			}
		})
	}
}

// TestDecodeServicesFromLiveCapture covers the two service shapes a broker really
// has, and the one genuinely useful warning: the LoadBalancer had no address.
func TestDecodeServicesFromLiveCapture(t *testing.T) {
	var list serviceList
	loadFixture(t, "services.json", &list)
	if len(list.Items) != 2 {
		t.Fatalf("decoded %d services, want 2", len(list.Items))
	}
	var sawPendingLB, sawHeadless bool
	for _, it := range list.Items {
		lvl, addr := serviceAddress(it.Spec.Type, it.Spec.ClusterIP, it.Status.LoadBalancer.Ingress)
		switch it.Spec.Type {
		case "LoadBalancer":
			// The captured cluster showed EXTERNAL-IP <pending>.
			if lvl != output.LevelWarn {
				t.Errorf("%s: level = %v, want LevelWarn for an unassigned address", it.Metadata.Name, lvl)
			}
			sawPendingLB = true
		default:
			if lvl != output.LevelInfo {
				t.Errorf("%s: level = %v, want LevelInfo -- a ClusterIP is not a health signal", it.Metadata.Name, lvl)
			}
			if addr == "headless" {
				sawHeadless = true
			}
		}
	}
	if !sawPendingLB || !sawHeadless {
		t.Errorf("expected both a pending LoadBalancer and the headless discovery service")
	}
}

func TestServiceAddressAssigned(t *testing.T) {
	type ingress = struct {
		IP       string `json:"ip"`
		Hostname string `json:"hostname"`
	}
	if lvl, addr := serviceAddress("LoadBalancer", "10.0.0.1", []ingress{{IP: "203.0.113.7"}}); lvl != output.LevelOK || addr != "203.0.113.7" {
		t.Errorf("assigned IP = %v/%q, want LevelOK", lvl, addr)
	}
	if lvl, addr := serviceAddress("LoadBalancer", "10.0.0.1", []ingress{{Hostname: "lb.example.com"}}); lvl != output.LevelOK || addr != "lb.example.com" {
		t.Errorf("assigned hostname = %v/%q, want LevelOK", lvl, addr)
	}
	// A plain ClusterIP with a real address: a fact, not a verdict.
	if lvl, addr := serviceAddress("ClusterIP", "10.96.0.20", nil); lvl != output.LevelInfo || addr != "10.96.0.20" {
		t.Errorf("ClusterIP = %v/%q, want LevelInfo with the address", lvl, addr)
	}
}

// TestDecodePVCsFromLiveCapture pins the claim shape, including that the monitor
// node's claim is deliberately a different size from the messaging nodes'.
func TestDecodePVCsFromLiveCapture(t *testing.T) {
	var list pvcList
	loadFixture(t, "pvc.json", &list)
	if len(list.Items) != 3 {
		t.Fatalf("decoded %d claims, want 3", len(list.Items))
	}
	sizes := map[string]int{}
	for _, it := range list.Items {
		if lvl := pvcLevel(it.Status.Phase); lvl != output.LevelOK {
			t.Errorf("%s: phase %q classified %v, want LevelOK for Bound", it.Metadata.Name, it.Status.Phase, lvl)
		}
		if it.Spec.StorageClassName == "" {
			t.Errorf("%s: no storage class decoded", it.Metadata.Name)
		}
		sizes[it.Spec.Resources.Requests.Storage]++
	}
	if len(sizes) != 2 {
		t.Errorf("claim sizes = %v, want two distinct sizes (messaging vs monitor)", sizes)
	}
}

func TestPVCLevels(t *testing.T) {
	for phase, want := range map[string]output.Level{
		"Bound": output.LevelOK, "Pending": output.LevelWarn,
		"Lost": output.LevelFail, "Failed": output.LevelFail, "": output.LevelWarn,
	} {
		if got := pvcLevel(phase); got != want {
			t.Errorf("pvcLevel(%q) = %v, want %v", phase, got, want)
		}
	}
}

// TestDecodeOperatorFromLiveCapture reads the installed operator's own image,
// which is what lets a report show the version actually running -- the captured
// cluster ran 1.4.2 while this repo renders 1.4.0, exactly the drift worth saying.
func TestDecodeOperatorFromLiveCapture(t *testing.T) {
	var list deploymentList
	loadFixture(t, "operator-deploy.json", &list)
	if len(list.Items) != 1 {
		t.Fatalf("decoded %d deployments, want 1", len(list.Items))
	}
	it := list.Items[0]
	if it.Metadata.Name != operatorDeployment {
		t.Errorf("name = %q, want %q", it.Metadata.Name, operatorDeployment)
	}
	lvl, text := replicaLevel(it.Status.ReadyReplicas, it.Spec.Replicas)
	if lvl != output.LevelOK || text != "1/1" {
		t.Errorf("replicaLevel = %v/%q, want LevelOK 1/1", lvl, text)
	}
	containers := it.Spec.Template.Spec.Containers
	if len(containers) == 0 || containers[0].Image == "" {
		t.Fatal("no operator image decoded, so the report cannot name the running version")
	}
	if got := containers[0].Image; !strings.Contains(got, "1.4.2") {
		t.Errorf("operator image = %q, want the captured 1.4.2 tag", got)
	}
}

func TestReplicaLevels(t *testing.T) {
	one, zero, three := 1, 0, 3
	for _, tc := range []struct {
		name  string
		ready int
		want  *int
		lvl   output.Level
	}{
		{"all ready", 1, &one, output.LevelOK},
		{"none ready", 0, &one, output.LevelFail},
		{"partially ready", 1, &three, output.LevelWarn},
		{"scaled to zero", 0, &zero, output.LevelInfo},
		{"nil spec defaults to one", 1, nil, output.LevelOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if lvl, _ := replicaLevel(tc.ready, tc.want); lvl != tc.lvl {
				t.Errorf("replicaLevel(%d,%v) = %v, want %v", tc.ready, tc.want, lvl, tc.lvl)
			}
		})
	}
}

// TestAgeMatchesKubectlShape pins the AGE format against the bands kubectl uses,
// including the two the live capture actually showed (6m40s and 40m). This column
// replaces kubectl's, so a difference here is a difference an operator would
// notice while comparing the two.
func TestAgeMatchesKubectlShape(t *testing.T) {
	base := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		d    time.Duration
		want string
	}{
		{0, "<1s"},
		{45 * time.Second, "45s"},
		{119 * time.Second, "119s"},
		{6*time.Minute + 40*time.Second, "6m40s"}, // the live pods
		{7 * time.Minute, "7m"},
		{40 * time.Minute, "40m"}, // the live operator
		{2*time.Hour + 59*time.Minute, "179m"},
		{5*time.Hour + 30*time.Minute, "5h30m"},
		{6 * time.Hour, "6h"},
		{30 * time.Hour, "30h"},
		{50 * time.Hour, "2d2h"},
		{10 * 24 * time.Hour, "10d"},
	} {
		if got := age(base.Add(-tc.d), base); got != tc.want {
			t.Errorf("age(%s) = %q, want %q", tc.d, got, tc.want)
		}
	}
	if got := age(time.Time{}, base); got != "-" {
		t.Errorf("age(zero) = %q, want \"-\" rather than a computed age", got)
	}
	if got := age(base.Add(time.Hour), base); got != "0s" {
		t.Errorf("a future timestamp = %q, want 0s", got)
	}
}

func TestRoleRankOrdersTheRedundancyGroup(t *testing.T) {
	// Ordering only: there is no ROLE column, because the name already ends in
	// -p-0/-b-0/-m-0. What kubectl cannot give is this order -- it sorts
	// alphabetically, which puts backup first.
	for name, want := range map[string]int{
		"dev-broker-pubsubplus-p-0": 0,
		"dev-broker-pubsubplus-b-0": 1,
		"dev-broker-pubsubplus-m-0": 2,
		"dev-broker-pubsubplus-p":   0,
		"something-else":            3,
	} {
		if got := roleRank(name); got != want {
			t.Errorf("roleRank(%q) = %d, want %d", name, got, want)
		}
	}
}

// TestNormalizeToListHandlesBothKubectlShapes is the regression test for a bug
// that reached a live cluster: `kubectl get deployment NAME -o json` returns the
// single object, NOT a list, so a decoder expecting `items` reported a running
// operator as "not found in namespace solace-operator".
//
// The original unit test missed it because its fixture had been captured from an
// UNNAMED get, which does carry `items`. testdata/operator-deploy-named.json is
// the shape production actually asks for.
func TestNormalizeToListHandlesBothKubectlShapes(t *testing.T) {
	t.Run("a named get returns one object and still decodes", func(t *testing.T) {
		raw, err := os.ReadFile(filepath.Join("testdata", "operator-deploy-named.json"))
		if err != nil {
			t.Fatal(err)
		}
		// The fixture must genuinely lack `items`, or this test proves nothing.
		if strings.Contains(string(raw), `"items"`) {
			t.Fatal("the named-get fixture should be a single object, not a list")
		}
		var list deploymentList
		if err := json.Unmarshal(normalizeToList(raw), &list); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(list.Items) != 1 {
			t.Fatalf("decoded %d items from a single object, want 1", len(list.Items))
		}
		if list.Items[0].Metadata.Name != operatorDeployment {
			t.Errorf("name = %q, want %q", list.Items[0].Metadata.Name, operatorDeployment)
		}
		if list.Items[0].Status.ReadyReplicas != 1 {
			t.Errorf("readyReplicas = %d, want the running 1", list.Items[0].Status.ReadyReplicas)
		}
	})
	t.Run("a list passes through untouched", func(t *testing.T) {
		raw := []byte(`{"kind":"DeploymentList","items":[{"metadata":{"name":"a"}}]}`)
		if got := normalizeToList(raw); string(got) != string(raw) {
			t.Errorf("a list must not be rewrapped:\n%s", got)
		}
	})
	t.Run("an empty list passes through", func(t *testing.T) {
		raw := []byte(`{"kind":"DeploymentList","items":[]}`)
		var list deploymentList
		if err := json.Unmarshal(normalizeToList(raw), &list); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(list.Items) != 0 {
			t.Errorf("an empty list must stay empty, got %d items", len(list.Items))
		}
	})
	t.Run("unrecognised input is left for the real decode to report", func(t *testing.T) {
		for _, raw := range []string{`not json`, `[]`, `{}`} {
			if got := normalizeToList([]byte(raw)); string(got) != raw {
				t.Errorf("normalizeToList(%q) rewrapped it as %q", raw, got)
			}
		}
	})
}

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

// The namespace gate exists because deleting a Namespace cascades to everything in it,
// including whatever another team put there -- and this tool very often did not create the
// namespace in the first place.
//
// Every test here is really about the DIRECTION of a wrong answer. A namespace wrongly
// reported empty gets cascade-deleted and the mistake is unrecoverable; one wrongly
// reported occupied merely stays. So the gate is built to fail toward "occupied", and
// these pin that it does.

func occCfg() *config.Config {
	c := &config.Config{}
	// The runtime is what the exec guard checks before argv is built, so a fixture
	// without one fails inside CheckCommand rather than reaching the code under test.
	c.K8s.Command = config.Command{"kubectl"}
	c.K8s.Namespace = "solace"
	c.K8s.Name = "dev-broker"
	return c
}

func occCluster(out string, err error) (*Cluster, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	rr := &recRunner{out: []byte(out), outErr: err}
	return &Cluster{R: rr, Cfg: occCfg(), Out: buf,
		Log: func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }}, buf
}

// TestNamespaceContentsDiscountsWhatKubernetesPutsThere is the test without which the
// prompt could never appear at all.
//
// Kubernetes creates a `default` ServiceAccount in every namespace and injects a
// `kube-root-ca.crt` ConfigMap into every namespace. Counting either as occupancy means no
// namespace is ever empty, so the namespace is never offered, so the feature silently does
// nothing -- and it would look like it worked, because "kept" is the safe outcome.
func TestNamespaceContentsDiscountsWhatKubernetesPutsThere(t *testing.T) {
	c, _ := occCluster("serviceaccount/default\nconfigmap/kube-root-ca.crt\n", nil)
	got, err := c.NamespaceContents(context.Background())
	if err != nil {
		t.Fatalf("NamespaceContents: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("contents = %v, want empty: a namespace holding only what Kubernetes puts "+
			"in every namespace is empty, or the prompt never appears at all", got)
	}
}

// TestNamespaceContentsIgnoresClusterPolicyObjects: a LimitRange or ResourceQuota is
// stamped on by cluster policy without anyone asking, so refusing to remove a namespace
// because one is present would mean never removing one on a cluster that sets quotas.
func TestNamespaceContentsIgnoresClusterPolicyObjects(t *testing.T) {
	c, _ := occCluster("limitrange/mem-default\nresourcequota/compute\nevent/something.17\n", nil)
	got, err := c.NamespaceContents(context.Background())
	if err != nil {
		t.Fatalf("NamespaceContents: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("contents = %v, want empty", got)
	}
}

// TestNamespaceContentsReportsRealOccupants covers the branch that keeps the namespace,
// including the group-qualified spelling `-o name` sometimes returns.
func TestNamespaceContentsReportsRealOccupants(t *testing.T) {
	c, _ := occCluster("deployment.apps/someone-elses-app\nsecrets.v1./their-token\n"+
		"serviceaccount/default\n", nil)
	got, err := c.NamespaceContents(context.Background())
	if err != nil {
		t.Fatalf("NamespaceContents: %v", err)
	}
	want := "deployment/someone-elses-app,secret/their-token"
	if strings.Join(got, ",") != want {
		t.Errorf("contents = %v, want %s -- the group/version suffix must not defeat the "+
			"match, or a foreign object reads as a kind nothing recognises", got, want)
	}
}

// TestNamespaceContentsErrorMeansOccupied is the safety direction stated directly: being
// unable to see what is in a namespace is not permission to delete it.
func TestNamespaceContentsErrorMeansOccupied(t *testing.T) {
	c, _ := occCluster("", errors.New("forbidden: cannot list secrets"))
	_, err := c.NamespaceContents(context.Background())
	if err == nil {
		t.Fatal("a failed enumeration must be an error, so the caller keeps the namespace")
	}
	for _, want := range []string{"solace", "kept"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should name %q -- it has to say the namespace survived", err, want)
		}
	}
}

// TestRetainedPVCsKeepTheNamespace is the hazard the gate makes structurally impossible
// rather than merely documented.
//
// `broker remove` without --delete-data keeps the PVCs. Those PVCs are IN the namespace, so
// the namespace is not empty, so it is not offered -- which means "delete the namespace and
// take the retained data with it" cannot be reached by answering one prompt wrong. There is
// no flag combination that produces it.
func TestRetainedPVCsKeepTheNamespace(t *testing.T) {
	c, _ := occCluster("persistentvolumeclaim/data-dev-broker-pubsubplus-p-0\n"+
		"serviceaccount/default\n", nil)
	got, err := c.NamespaceContents(context.Background())
	if err != nil {
		t.Fatalf("NamespaceContents: %v", err)
	}
	if len(got) != 1 || !strings.HasPrefix(got[0], "persistentvolumeclaim/") {
		t.Fatalf("contents = %v, want the retained PVC to count as occupancy", got)
	}
}

// TestNamespaceContentsAsksOneQuestion pins the probe itself: a fixed kind list in ONE
// call, with --ignore-not-found so a kind the cluster lacks is not an error.
//
// The alternative -- `api-resources` then a bulk get -- would see CRD kinds too, but it
// needs `list` on every kind in the cluster, so a namespace-scoped identity fails the probe
// rather than answering it, and one unreadable kind then decides the whole question.
func TestNamespaceContentsAsksOneQuestion(t *testing.T) {
	rr := &recRunner{out: []byte("")}
	c := &Cluster{R: rr, Cfg: occCfg(), Out: &bytes.Buffer{}}
	if _, err := c.NamespaceContents(context.Background()); err != nil {
		t.Fatalf("NamespaceContents: %v", err)
	}
	if len(rr.calls) != 1 {
		t.Fatalf("made %d calls, want exactly 1", len(rr.calls))
	}
	want := []string{"get", "all,persistentvolumeclaims,secrets,configmaps", "-n", "solace",
		"--ignore-not-found", "-o", "name"}
	if !eqArgs(rr.calls[0].args, want) {
		t.Errorf("argv = %v, want %v", rr.calls[0].args, want)
	}
}

// TestNamespaceIsProtectedSharesDeleteNamespacesList: the four cluster namespaces are
// refused before the question is asked, and from the SAME list DeleteNamespace uses. Two
// lists would be two things to keep in step, and the one that mattered would be whichever
// the call path happened to reach.
func TestNamespaceIsProtectedSharesDeleteNamespacesList(t *testing.T) {
	for _, ns := range []string{"default", "kube-system", "kube-public", "kube-node-lease"} {
		cfg := occCfg()
		cfg.K8s.Namespace = ns
		c := &Cluster{Cfg: cfg}
		if !c.NamespaceIsProtected() {
			t.Errorf("%q must be protected: there is no reading of the env file that should "+
				"let this tool delete it", ns)
		}
		if !protectedNamespaces[ns] {
			t.Errorf("%q is not in protectedNamespaces, so DeleteNamespace and this gate "+
				"disagree", ns)
		}
	}
	c := &Cluster{Cfg: occCfg()}
	if c.NamespaceIsProtected() {
		t.Error("an ordinary broker namespace must not be protected, or nothing is removable")
	}
}

// TestReportNamespaceOccupiedNamesTheObjects: the operator sees this branch most often, so
// it lists what is in the way. "3 objects remain" gives nobody anything to act on.
func TestReportNamespaceOccupiedNamesTheObjects(t *testing.T) {
	buf := &bytes.Buffer{}
	c := &Cluster{Cfg: occCfg(), Out: buf, Log: func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }}
	c.ReportNamespaceOccupied([]string{"deployment/someone-elses-app", "secret/their-token"})
	out := buf.String()
	for _, want := range []string{"someone-elses-app", "their-token", "kubectl delete namespace solace"} {
		if !strings.Contains(out, want) {
			t.Errorf("report should contain %q:\n%s", want, out)
		}
	}
}

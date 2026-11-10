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

// nsItem is one object of kubectl's JSON listing: its kind and name, and the owners it
// names as "Kind/name" ownerReferences.
func nsItem(kind, name string, owners ...string) string {
	refs := make([]string, 0, len(owners))
	for _, o := range owners {
		k, n, _ := strings.Cut(o, "/")
		refs = append(refs, fmt.Sprintf(`{"kind":%q,"name":%q}`, k, n))
	}
	return fmt.Sprintf(`{"kind":%q,"metadata":{"name":%q,"ownerReferences":[%s]}}`, kind, name, strings.Join(refs, ","))
}

// nsList is the listing nsgate.List asks for, holding items.
func nsList(items ...string) string { return `{"items":[` + strings.Join(items, ",") + `]}` }

// TestNamespaceOccupantsDiscountsWhatKubernetesPutsThere is the test without which the
// prompt could never appear at all.
//
// Kubernetes creates a `default` ServiceAccount in every namespace, injects a
// `kube-root-ca.crt` ConfigMap into every namespace, and on a cluster that still mints
// them gives the default ServiceAccount a token Secret. Counting any as occupancy means no
// namespace is ever empty, so the namespace is never offered, so the feature silently does
// nothing -- and it would look like it worked, because "kept" is the safe outcome.
func TestNamespaceOccupantsDiscountsWhatKubernetesPutsThere(t *testing.T) {
	token := `{"kind":"Secret","type":"kubernetes.io/service-account-token","metadata":{"name":"default-token-x7k2p",` +
		`"annotations":{"kubernetes.io/service-account.name":"default"}}}`
	c, _ := occCluster(nsList(nsItem("ServiceAccount", "default"), nsItem("ConfigMap", "kube-root-ca.crt"), token), nil)
	got, answered, err := c.NamespaceOccupants(context.Background())
	if err != nil || !answered {
		t.Fatalf("NamespaceOccupants = %v, answered %v, err %v", got, answered, err)
	}
	if len(got) != 0 {
		t.Errorf("occupants = %v, want empty: a namespace holding only what Kubernetes puts "+
			"in every namespace is empty, or the prompt never appears at all", got)
	}
}

// TestNamespaceOccupantsIgnoresClusterPolicyObjects: a LimitRange or ResourceQuota is
// stamped on by cluster policy without anyone asking, so refusing to remove a namespace
// because one is present would mean never removing one on a cluster that sets quotas.
func TestNamespaceOccupantsIgnoresClusterPolicyObjects(t *testing.T) {
	c, _ := occCluster(nsList(nsItem("LimitRange", "mem-default"), nsItem("ResourceQuota", "compute"),
		nsItem("Event", "something.17")), nil)
	got, _, err := c.NamespaceOccupants(context.Background())
	if err != nil {
		t.Fatalf("NamespaceOccupants: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("occupants = %v, want empty", got)
	}
}

// TestNamespaceOccupantsReportsRealOccupants covers the branch that keeps the namespace:
// another team's objects count, named as lower-case kind/name.
func TestNamespaceOccupantsReportsRealOccupants(t *testing.T) {
	c, _ := occCluster(nsList(nsItem("Deployment", "someone-elses-app"), nsItem("Secret", "their-token"),
		nsItem("ServiceAccount", "default")), nil)
	got, _, err := c.NamespaceOccupants(context.Background())
	if err != nil {
		t.Fatalf("NamespaceOccupants: %v", err)
	}
	want := "deployment/someone-elses-app,secret/their-token"
	if strings.Join(got, ",") != want {
		t.Errorf("occupants = %v, want %s", got, want)
	}
}

// TestNamespaceOccupantsDiscountsTheBrokerAndWhatItOwns: the broker CR is the root of
// what the operator builds, so its StatefulSet, Pods and Service (ownerReferences, followed
// to the end) are the deployment's own leftovers, and so are the Secrets this env file
// builds. Another team's Deployment in the same namespace still keeps it.
func TestNamespaceOccupantsDiscountsTheBrokerAndWhatItOwns(t *testing.T) {
	cfg := occCfg()
	cfg.SEMP.AdminPass = "pw" // the env file builds the admin Secret, <name>-admin
	owned := ownedSecretNames(cfg)
	if len(owned) != 1 {
		t.Fatalf("ownedSecretNames = %v, want the one admin Secret", owned)
	}
	list := nsList(
		nsItem("PubSubPlusEventBroker", "dev-broker"),
		nsItem("StatefulSet", "dev-broker-pubsubplus-p", "PubSubPlusEventBroker/dev-broker"),
		nsItem("Pod", "dev-broker-pubsubplus-p-0", "StatefulSet/dev-broker-pubsubplus-p"),
		nsItem("Service", "dev-broker-pubsubplus", "PubSubPlusEventBroker/dev-broker"),
		nsItem("Secret", owned[0]),
		nsItem("Deployment", "someone-elses-app"),
	)
	c := &Cluster{R: &recRunner{out: []byte(list)}, Cfg: cfg, Out: &bytes.Buffer{}}
	got, _, err := c.NamespaceOccupants(context.Background())
	if err != nil {
		t.Fatalf("NamespaceOccupants: %v", err)
	}
	if strings.Join(got, ",") != "deployment/someone-elses-app" {
		t.Errorf("occupants = %v, want only the foreign Deployment", got)
	}

	// A second broker's objects are not this one's: its Pod names another root.
	other := nsList(nsItem("PubSubPlusEventBroker", "other-broker"),
		nsItem("Pod", "other-broker-pubsubplus-p-0", "StatefulSet/other-broker-pubsubplus-p"))
	c = &Cluster{R: &recRunner{out: []byte(other)}, Cfg: occCfg(), Out: &bytes.Buffer{}}
	got, _, _ = c.NamespaceOccupants(context.Background())
	if len(got) != 2 {
		t.Errorf("occupants = %v, want a second broker's objects to keep the namespace", got)
	}
}

// TestNamespaceOccupantsErrorMeansOccupied is the safety direction stated directly: being
// unable to see what is in a namespace, or to read the answer, is not permission to delete
// it.
func TestNamespaceOccupantsErrorMeansOccupied(t *testing.T) {
	c, _ := occCluster("", errors.New("forbidden: cannot list secrets"))
	_, _, err := c.NamespaceOccupants(context.Background())
	if err == nil {
		t.Fatal("a failed enumeration must be an error, so the caller keeps the namespace")
	}
	for _, want := range []string{"solace", "kept"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should name %q -- it has to say the namespace survived", err, want)
		}
	}
	c, _ = occCluster("NAME  READY\npod/x  1/1\n", nil)
	if _, answered, err := c.NamespaceOccupants(context.Background()); err == nil || !answered {
		t.Errorf("an answer that is not JSON: answered %v, err %v; want an error, so the namespace is kept", answered, err)
	}
}

// TestNamespaceOccupantsSkipsWhenTheRunnerAnswersNothing: Echo and a silent test seam
// return no listing at all. That is not "empty" and not a failure, so the check reports
// itself skipped and the caller keeps the namespace without a warning.
func TestNamespaceOccupantsSkipsWhenTheRunnerAnswersNothing(t *testing.T) {
	c, buf := occCluster("", nil)
	got, answered, err := c.NamespaceOccupants(context.Background())
	if err != nil || answered || len(got) != 0 {
		t.Fatalf("NamespaceOccupants = %v, answered %v, err %v; want no answer and no error", got, answered, err)
	}
	if !strings.Contains(buf.String(), "namespace") || !strings.Contains(buf.String(), "check skipped (no answer from the runner)") {
		t.Errorf("report = %q, want the skipped-check row", buf.String())
	}
}

// TestRetainedPVCsKeepTheNamespace is the hazard the gate makes structurally impossible
// rather than merely documented.
//
// `broker remove` without --delete-data keeps the PVCs. Those PVCs are IN the namespace and
// carry no ownerReference to the broker, so the namespace is not empty, so it is not
// offered -- which means "delete the namespace and take the retained data with it" cannot be
// reached by answering one prompt wrong. There is no flag combination that produces it.
func TestRetainedPVCsKeepTheNamespace(t *testing.T) {
	c, _ := occCluster(nsList(nsItem("PersistentVolumeClaim", "data-dev-broker-pubsubplus-p-0"),
		nsItem("ServiceAccount", "default")), nil)
	got, _, err := c.NamespaceOccupants(context.Background())
	if err != nil {
		t.Fatalf("NamespaceOccupants: %v", err)
	}
	if len(got) != 1 || !strings.HasPrefix(got[0], "persistentvolumeclaim/") {
		t.Fatalf("occupants = %v, want the retained PVC to count as occupancy", got)
	}
}

// TestNamespaceOccupantsAsksOneQuestion pins the probe itself, nsgate.List: a fixed kind
// list in ONE call, as JSON (so ownerReferences and deletion times are readable), and with
// no --ignore-not-found, with which kubectl prints nothing at all for an empty list.
//
// The alternative -- `api-resources` then a bulk get -- would see CRD kinds too, but it
// needs `list` on every kind in the cluster, so a namespace-scoped identity fails the probe
// rather than answering it, and one unreadable kind then decides the whole question.
func TestNamespaceOccupantsAsksOneQuestion(t *testing.T) {
	rr := &recRunner{out: []byte(nsList())}
	c := &Cluster{R: rr, Cfg: occCfg(), Out: &bytes.Buffer{}}
	if _, _, err := c.NamespaceOccupants(context.Background()); err != nil {
		t.Fatalf("NamespaceOccupants: %v", err)
	}
	if len(rr.calls) != 1 {
		t.Fatalf("made %d calls, want exactly 1", len(rr.calls))
	}
	want := []string{"get", "all,persistentvolumeclaims,secrets,configmaps", "-n", "solace", "-o", "json"}
	if !eqArgs(rr.calls[0].args, want) {
		t.Errorf("argv = %v, want %v", rr.calls[0].args, want)
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

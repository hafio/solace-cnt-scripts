package k8s

import (
	"context"
	"errors"
	"testing"
)

// newCluster builds a Cluster over a fresh recRunner for argv/behaviour assertions.
func newCluster(rr *recRunner) *Cluster {
	return NewCluster(rr, haCfg(), nil, nil)
}

func TestOperatorNSExplicit(t *testing.T) {
	rr := &recRunner{}
	c := newCluster(rr)
	c.Cfg.K8s.Operator.Namespace = "chosen-ns"
	if got := c.operatorNS(context.Background()); got != "chosen-ns" {
		t.Errorf("operatorNS = %q, want chosen-ns", got)
	}
	if len(rr.calls) != 0 {
		t.Errorf("explicit namespace must not probe the cluster; got %d calls", len(rr.calls))
	}
}

// TestOperatorNSNeverProbesTheCluster is what replaced the three
// discovery tests (derived / default-when-absent / default-on-error). Resolving
// the operator namespace used to list Deployments in EVERY namespace and take the
// first line CONTAINING the operator's name -- an unanchored substring match with
// no uniqueness check, which on a cluster running two operator installs could
// resolve to another team's and have `remove operator` delete it.
//
// Now there are two local rules and no cluster call at all, which is the property
// worth pinning: whatever the cluster looks like, and whatever a scripted reply
// says, resolution cannot be steered by anything outside the env file.
func TestOperatorNSNeverProbesTheCluster(t *testing.T) {
	operatorRow := []byte(
		"NS            NAME\n" +
			"kube-system   coredns\n" +
			"my-op-ns      pubsubplus-eventbroker-operator\n")
	for _, tc := range []struct {
		name string
		rr   *recRunner
		want string
	}{
		{"an operator running elsewhere is ignored", &recRunner{out: operatorRow}, defaultOperatorNS},
		{"so is an empty cluster", &recRunner{out: []byte("NS   NAME\n")}, defaultOperatorNS},
		{"so is an unreachable one", &recRunner{outErr: errors.New("connection refused")}, defaultOperatorNS},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newCluster(tc.rr) // haCfg leaves Operator.Namespace unset
			if got := c.operatorNS(context.Background()); got != tc.want {
				t.Errorf("operatorNS = %q, want %q", got, tc.want)
			}
			if len(tc.rr.calls) != 0 {
				t.Errorf("resolving the operator namespace must not talk to the cluster; got %d calls: %+v",
					len(tc.rr.calls), tc.rr.calls)
			}
		})
	}
}

func TestApplyOnStdin(t *testing.T) {
	rr := &recRunner{}
	c := newCluster(rr)
	manifest := []byte("kind: Namespace\n")
	if err := c.apply(context.Background(), manifest); err != nil {
		t.Fatalf("apply: %v", err)
	}
	got := rr.last()
	if got.method != "RunInput" || got.name != "kubectl" ||
		!eqArgs(got.args, []string{"apply", "-f", "-"}) || got.stdin != string(manifest) {
		t.Errorf("apply\n got: %+v\nwant RunInput kubectl [apply -f -] with manifest on stdin", got)
	}
}

func TestDeleteStdin(t *testing.T) {
	rr := &recRunner{}
	c := newCluster(rr)
	manifest := []byte("kind: Namespace\n")
	if err := c.deleteStdin(context.Background(), manifest); err != nil {
		t.Fatalf("deleteStdin: %v", err)
	}
	got := rr.last()
	if got.method != "RunInput" || got.name != "kubectl" ||
		!eqArgs(got.args, []string{"delete", "-f", "-", "--ignore-not-found"}) || got.stdin != string(manifest) {
		t.Errorf("deleteStdin\n got: %+v\nwant RunInput kubectl [delete -f - --ignore-not-found]", got)
	}
}

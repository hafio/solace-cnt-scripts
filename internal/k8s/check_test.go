package k8s

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

// errFake is a stand-in cluster error shared by the check/prep tests.
var errFake = errors.New("connection refused")

func TestReachable(t *testing.T) {
	t.Run("reachable", func(t *testing.T) {
		rr := &recRunner{}
		c := newCluster(rr)
		if err := c.Reachable(context.Background()); err != nil {
			t.Fatalf("Reachable: %v", err)
		}
		got := rr.last()
		if got.method != "Output" || !eqArgs(got.args, []string{"version", "-o", "json"}) {
			t.Errorf("Reachable argv = %+v, want Output kubectl [version -o json]", got)
		}
	})
	t.Run("unreachable", func(t *testing.T) {
		c := newCluster(&recRunner{outErr: errFake})
		if err := c.Reachable(context.Background()); err == nil {
			t.Error("Reachable should fail when the API server errors")
		}
	})
}

// TestCheckStopsProbingWhenUnreachable pins what an unreachable API server costs.
// The report still fails, but the reason appears as a [FAIL] ROW rather than in
// the returned error, and -- the part worth pinning -- no further probe runs: with
// no API server every one of them would fail for the same reason, and three
// identical failures are less useful than one plus three honest skips.
func TestCheckStopsProbingWhenUnreachable(t *testing.T) {
	rr := &recRunner{outErr: errFake}
	buf := &bytes.Buffer{}
	c := NewCluster(rr, loadK8s(t), nil, buf)
	err := c.Check(context.Background())
	if err == nil {
		t.Fatal("Check must fail when the cluster cannot be reached")
	}
	out := buf.String()
	if !strings.Contains(out, "[FAIL] api server") || !strings.Contains(out, "cannot reach") {
		t.Errorf("the reachability failure belongs in the report:\n%s", out)
	}
	// The checks that need the cluster say they were skipped, rather than being
	// omitted (which would read as a pass) or repeating the same failure.
	for _, want := range []string{"[SKIP] permission", "[SKIP] installed", "[SKIP] storage class"} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing %q:\n%s", want, out)
		}
	}
	if len(rr.calls) != 1 {
		t.Errorf("Check should stop after the Reachable probe; got %d calls", len(rr.calls))
	}
}

func TestResolveStorageClass(t *testing.T) {
	t.Run("configured", func(t *testing.T) {
		cfg := haCfg()
		cfg.K8s.Storage.Class = "chosen"
		rr := &recRunner{}
		c := NewCluster(rr, cfg, nil, nil)
		got, err := c.resolveStorageClass(context.Background())
		if err != nil || got != "chosen" {
			t.Fatalf("resolveStorageClass = (%q,%v), want (chosen,nil)", got, err)
		}
		if len(rr.calls) != 0 {
			t.Errorf("a configured class must not query the cluster; got %d calls", len(rr.calls))
		}
	})
	t.Run("single default", func(t *testing.T) {
		rr := &recRunner{out: []byte("standard\n")}
		c := NewCluster(rr, haCfg(), nil, nil)
		got, err := c.resolveStorageClass(context.Background())
		if err != nil || got != "standard" {
			t.Fatalf("resolveStorageClass = (%q,%v), want (standard,nil)", got, err)
		}
		if a := rr.last().args; a[0] != "get" || a[1] != "sc" || a[2] != "-o" || !strings.HasPrefix(a[3], "jsonpath=") {
			t.Errorf("default-resolution argv = %v", a)
		}
	})
	t.Run("multiple defaults is an error", func(t *testing.T) {
		c := NewCluster(&recRunner{out: []byte("a b\n")}, haCfg(), nil, nil)
		if _, err := c.resolveStorageClass(context.Background()); err == nil {
			t.Error("resolveStorageClass should reject multiple default classes")
		}
	})
	t.Run("no default returns empty", func(t *testing.T) {
		c := NewCluster(&recRunner{out: []byte("")}, haCfg(), nil, nil)
		got, err := c.resolveStorageClass(context.Background())
		if err != nil || got != "" {
			t.Fatalf("resolveStorageClass = (%q,%v), want (\"\",nil)", got, err)
		}
	})
}

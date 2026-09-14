package k8s

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"solace/internal/engine"
)

func TestImageTag(t *testing.T) {
	for _, tc := range []struct {
		ref, repo, tag string
	}{
		{"registry.example.com/solace/operator:1.4.2", "registry.example.com/solace/operator", "1.4.2"},
		{"solace/operator:1.4.0", "solace/operator", "1.4.0"},
		{"operator:latest", "operator", "latest"},
		// A registry port must not be mistaken for a tag.
		{"registry.example.com:5000/solace/operator", "registry.example.com:5000/solace/operator", ""},
		{"registry.example.com:5000/solace/operator:2.0", "registry.example.com:5000/solace/operator", "2.0"},
		// A digest pin has no version to order.
		{"solace/operator@sha256:abc123", "solace/operator", ""},
		{"solace/operator", "solace/operator", ""},
	} {
		repo, tag := imageTag(tc.ref)
		if repo != tc.repo || tag != tc.tag {
			t.Errorf("imageTag(%q) = %q,%q; want %q,%q", tc.ref, repo, tag, tc.repo, tc.tag)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
		ok   bool
	}{
		{"1.4.2", "1.4.0", 1, true},
		{"1.4.0", "1.4.2", -1, true},
		{"1.4.0", "1.4.0", 0, true},
		{"2.0.0", "1.9.9", 1, true},
		{"1.10.0", "1.9.0", 1, true}, // numeric, not lexical
		{"v1.4.2", "1.4.0", 1, true},
		{"1.4", "1.4.0", 0, true}, // a missing component is zero
		{"1.4.0-rc1", "1.4.0", 0, true},
		// Not orderable: a guess here would produce a downgrade alarm nobody can act on.
		{"latest", "1.4.0", 0, false},
		{"1.4.0", "latest", 0, false},
		{"20260821", "1.4.0", 1, true}, // a date stamp IS dotted-numeric; it just sorts high
		{"", "1.4.0", 0, false},
		{"abc.def", "1.4.0", 0, false},
	} {
		got, ok := compareVersions(tc.a, tc.b)
		if ok != tc.ok || (ok && got != tc.want) {
			t.Errorf("compareVersions(%q,%q) = %d,%v; want %d,%v", tc.a, tc.b, got, ok, tc.want, tc.ok)
		}
	}
}

// TestOperatorVersionWarningFiresOnlyOnADowngrade is the whole point: the block
// appears when the cluster is AHEAD of us and stays silent otherwise. A warning on
// every deploy would be scrolled past, and the one case that matters would go with
// it.
func TestOperatorVersionWarningFiresOnlyOnADowngrade(t *testing.T) {
	const newer = "registry.example.com/solace/pubsubplus-eventbroker-operator:1.4.2"
	const older = "registry.example.com/solace/pubsubplus-eventbroker-operator:1.4.0"

	t.Run("existing newer than new -- warn", func(t *testing.T) {
		body := operatorVersionWarning(newer, older)
		if body == nil {
			t.Fatal("a downgrade must warn")
		}
		joined := strings.Join(body, "\n")
		for _, want := range []string{"existing : 1.4.2", "new      : 1.4.0"} {
			if !strings.Contains(joined, want) {
				t.Errorf("warning body missing %q:\n%s", want, joined)
			}
		}
		// Deliberately just the two versions: the confirmation prompt that follows
		// is what carries the weight, and prose in front of a question is prose
		// somebody scrolls past to reach the question.
		if len(body) != 2 {
			t.Errorf("body has %d lines, want just the two versions:\n%s", len(body), joined)
		}
	})
	t.Run("upgrade -- silent", func(t *testing.T) {
		if body := operatorVersionWarning(older, newer); body != nil {
			t.Errorf("an upgrade must not warn:\n%s", strings.Join(body, "\n"))
		}
	})
	t.Run("same version -- silent", func(t *testing.T) {
		if body := operatorVersionWarning(newer, newer); body != nil {
			t.Errorf("re-applying the same version must not warn:\n%s", strings.Join(body, "\n"))
		}
	})
	t.Run("nothing installed yet -- silent", func(t *testing.T) {
		if body := operatorVersionWarning("", older); body != nil {
			t.Error("a first install has nothing to downgrade")
		}
	})
	t.Run("unorderable tags -- silent", func(t *testing.T) {
		if body := operatorVersionWarning("solace/operator:latest", older); body != nil {
			t.Error("an unorderable tag must not produce an alarm nobody can act on")
		}
		if body := operatorVersionWarning("solace/operator@sha256:abc", older); body != nil {
			t.Error("a digest-pinned operator has no version to compare")
		}
	})
}

// TestConfirmNoDowngradeIsSilentInAPreview covers the guard: a preview has no
// cluster to read, so it must neither claim anything about what is installed nor
// block the preview on a question it has no basis to ask.
func TestConfirmNoDowngradeIsSilentInAPreview(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := haCfg()
	cfg.K8s.Operator.Image = "solace/pubsubplus-eventbroker-operator:1.4.0"
	c := NewCluster(engine.Echo{W: buf}, cfg, func(f string, a ...any) {
		fmt.Fprintf(buf, f+"\n", a...)
	}, buf)
	if err := c.confirmNoDowngrade(context.Background(), "solace-operator"); err != nil {
		t.Errorf("a preview must not block: %v", err)
	}
	if strings.Contains(buf.String(), "WARNING") {
		t.Errorf("a preview must not warn about an installed version it cannot read: %q", buf.String())
	}
}

// operatorDeployJSON builds the `kubectl get deployment --all-namespaces -o json`
// reply installedOperatorImage decodes, naming only the fields it reads. Tests
// script it through recRunner's outQueue so confirmNoDowngrade is driven by the
// exact query shape production issues, not a shortcut around it.
func operatorDeployJSON(image string) []byte {
	return []byte(fmt.Sprintf(
		`{"kind":"DeploymentList","items":[{"metadata":{"name":%q},"spec":{"template":{"spec":{"containers":[{"image":%q}]}}}}]}`,
		operatorDeployment, image))
}

// The tests below drive confirmNoDowngrade through OperatorApply with the
// RECORDING runner rather than Echo, so isEcho's short-circuit (the branch every
// prior test took) is out of the way and the guard's real logic -- including the
// nil-Confirm fail-safe default -- actually runs. Each asserts on the recorded
// argv, not just the returned error: a refusal only counts if no apply call was
// ever issued.

// TestConfirmNoDowngradeRefusesByDefault is the fail-safe default this whole gate
// exists for and the most important case: an unattended run (nil Confirm)
// declines a downgrade rather than silently applying an operator older than the
// one already running, and nothing is applied before the refusal.
func TestConfirmNoDowngradeRefusesByDefault(t *testing.T) {
	cfg := haCfg()
	cfg.K8s.Operator.Image = "solace/pubsubplus-eventbroker-operator:1.4.0"
	rr := &recRunner{outQueue: [][]byte{
		operatorDeployJSON("registry.example.com/solace/pubsubplus-eventbroker-operator:1.4.2"),
	}}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.OperatorApply(context.Background()); err == nil {
		t.Fatal("a downgrade with no Confirm must be refused")
	}
	calls := rr.afterPreflights(t, probe{verb: "create", resource: "customresourcedefinitions"})
	if len(calls) != 1 {
		t.Fatalf("refused downgrade made %d calls after the probe, want 1 (the version read only, no apply)", len(calls))
	}
}

// TestConfirmNoDowngradeRefusedWhenConfirmDeclines covers an interactive operator
// answering no: still refused, still nothing applied.
func TestConfirmNoDowngradeRefusedWhenConfirmDeclines(t *testing.T) {
	cfg := haCfg()
	cfg.K8s.Operator.Image = "solace/pubsubplus-eventbroker-operator:1.4.0"
	rr := &recRunner{outQueue: [][]byte{
		operatorDeployJSON("registry.example.com/solace/pubsubplus-eventbroker-operator:1.4.2"),
	}}
	c := NewCluster(rr, cfg, nil, nil)
	c.Confirm = func(string) bool { return false }
	if err := c.OperatorApply(context.Background()); err == nil {
		t.Fatal("declining the confirm must refuse the downgrade")
	}
	calls := rr.afterPreflights(t, probe{verb: "create", resource: "customresourcedefinitions"})
	if len(calls) != 1 {
		t.Fatalf("declined downgrade made %d calls after the probe, want 1 (the version read only, no apply)", len(calls))
	}
}

// TestConfirmNoDowngradeProceedsWhenConfirmAccepts covers the deliberate-rollback
// path: a downgrade is a legitimate thing to want, so answering yes lets the
// apply through the same three writes TestOperatorApply pins.
func TestConfirmNoDowngradeProceedsWhenConfirmAccepts(t *testing.T) {
	cfg := loadK8s(t) // pull secret set -> regcred applied
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{outQueue: [][]byte{
		operatorDeployJSON("registry.example.com/solace/pubsubplus-eventbroker-operator:1.4.2"),
	}}
	c := NewCluster(rr, cfg, nil, nil)
	c.Confirm = func(string) bool { return true }
	if err := c.OperatorApply(context.Background()); err != nil {
		t.Fatalf("OperatorApply: %v", err)
	}
	calls := rr.afterPreflights(t, probe{verb: "create", resource: "customresourcedefinitions"})
	if len(calls) != 5 {
		t.Fatalf("accepted downgrade made %d calls after the probe, want 5 "+
			"(version read + watch-list read + namespace + regcred + bundle)", len(calls))
	}
}

// TestConfirmNoDowngradeAsksNothingWhenNotADowngrade covers the two cases that
// are not downgrades at all -- the same version re-applied, and a genuine
// upgrade -- neither of which may even ask the question, let alone block on it.
func TestConfirmNoDowngradeAsksNothingWhenNotADowngrade(t *testing.T) {
	for _, tc := range []struct {
		name     string
		existing string
	}{
		{"same version", "registry.example.com/solace/pubsubplus-eventbroker-operator:1.4.0"},
		{"older installed -- an upgrade", "registry.example.com/solace/pubsubplus-eventbroker-operator:1.3.0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := loadK8s(t)
			cfg.K8s.Operator.Namespace = "op-ns"
			rr := &recRunner{outQueue: [][]byte{operatorDeployJSON(tc.existing)}}
			c := NewCluster(rr, cfg, nil, nil)
			asked := false
			c.Confirm = func(string) bool { asked = true; return false }
			if err := c.OperatorApply(context.Background()); err != nil {
				t.Fatalf("OperatorApply: %v", err)
			}
			if asked {
				t.Error("an upgrade or same-version apply must not ask at all")
			}
			calls := rr.afterPreflights(t, probe{verb: "create", resource: "customresourcedefinitions"})
			if len(calls) != 5 {
				t.Fatalf("made %d calls after the probe, want 5 "+
					"(version read + watch-list read + namespace + regcred + bundle)", len(calls))
			}
		})
	}
}

// TestConfirmNoDowngradeProceedsWhenVersionReadFails pins what the code actually
// does when the installed-version read itself fails (RBAC denial, an unreachable
// API server): installedOperatorImage folds any read error into "" -- the same
// value a genuine first install produces -- so confirmNoDowngrade has nothing to
// compare against and proceeds without asking. That is installedOperatorImage's
// existing documented behaviour (a first install must not alarm), not a new
// decision made here; this test exists so a future change to it is deliberate
// rather than an unnoticed regression in the downgrade gate.
func TestConfirmNoDowngradeProceedsWhenVersionReadFails(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.Namespace = "op-ns"
	rr := &recRunner{outErrQueue: []error{errFake}}
	c := NewCluster(rr, cfg, nil, nil)
	asked := false
	c.Confirm = func(string) bool { asked = true; return false }
	if err := c.OperatorApply(context.Background()); err != nil {
		t.Fatalf("OperatorApply: %v", err)
	}
	if asked {
		t.Error("a failed version read must not be treated as a downgrade to ask about")
	}
	calls := rr.afterPreflights(t, probe{verb: "create", resource: "customresourcedefinitions"})
	if len(calls) != 5 {
		t.Fatalf("made %d calls after the probe, want 4 (version read + namespace + regcred + bundle)", len(calls))
	}
}

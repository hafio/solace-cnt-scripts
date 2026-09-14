package k8s

import (
	"context"
	"strings"
	"testing"

	"solace/internal/engine"
)

// TestCanIAnswerReadsTheLastLine: kubectl prints advisory lines above the verdict on
// STDOUT -- "Warning: resource 'x' is not namespace scoped" is routine, and a cluster
// carrying deprecated APIs adds more. Comparing the whole output would turn every
// such cluster into the unreadable-answer branch and block a deploy the operator is
// entitled to make, which is a preflight that fails safe in the wrong direction.
func TestCanIAnswerReadsTheLastLine(t *testing.T) {
	cases := []struct {
		name string
		out  string
		want string
	}{
		{"plain yes", "yes\n", "yes"},
		{"plain no", "no\n", "no"},
		{"no trailing newline", "yes", "yes"},
		{"warning above the answer", "Warning: resource 'pubsubpluseventbrokers' is not namespace scoped\nyes\n", "yes"},
		{"several warnings", "Warning: a\nWarning: b\nno\n", "no"},
		{"blank lines around", "\n\nyes\n\n\n", "yes"},
		{"crlf", "Warning: a\r\nyes\r\n", "yes"},
		{"empty", "", ""},
		{"only whitespace", "  \n\t\n", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := canIAnswer([]byte(tc.out)); got != tc.want {
				t.Errorf("canIAnswer(%q) = %q, want %q", tc.out, got, tc.want)
			}
		})
	}
}

// TestPreflightAcceptsAWarnedYes is the end-to-end of the above: a cluster that warns
// and then permits must let the deploy proceed.
func TestPreflightAcceptsAWarnedYes(t *testing.T) {
	rr := &recRunner{canI: "Warning: resource 'pubsubpluseventbrokers' is not namespace scoped\nyes"}
	c := NewCluster(rr, loadK8s(t), nil, nil)
	if err := c.Preflight(context.Background(), "create", brokerResource); err != nil {
		t.Fatalf("a warned `yes` must pass: %v", err)
	}
}

// TestPreflightRefusesAnUnreadableAnswer: exit 0 with something that is neither yes
// nor no -- a wrapper that swallowed stdout -- is refused rather than assumed
// permitted. Proceeding would mean acting on a permission nobody confirmed.
func TestPreflightRefusesAnUnreadableAnswer(t *testing.T) {
	rr := &recRunner{canI: "maybe"}
	c := NewCluster(rr, loadK8s(t), nil, nil)
	err := c.Preflight(context.Background(), "create", brokerResource)
	if err == nil {
		t.Fatal("an unreadable answer must not be treated as permission")
	}
	for _, want := range []string{"could not read the answer", "passes stdout through unchanged"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %v, want it to contain %q", err, want)
		}
	}
}

// TestPreflightIsPreviewableUnderEcho: the Echo runner records the probe and skips the
// assertion, so a preview needs no cluster. This is why there is no skip flag -- the
// one legitimate reason to skip already has one.
func TestPreflightIsPreviewableUnderEcho(t *testing.T) {
	var out strings.Builder
	c := NewCluster(engine.Echo{W: &out}, loadK8s(t), nil, &out)
	if err := c.Preflight(context.Background(), "create", brokerResource); err != nil {
		t.Fatalf("Preflight under the Echo runner must not fail: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "auth can-i create "+brokerResource) {
		t.Errorf("dry-run should still show the probe it would run:\n%s", got)
	}
	if !strings.Contains(got, "skipped (preview)") {
		t.Errorf("dry-run should say the assertion was skipped:\n%s", got)
	}
}

// TestCanIArgsScopesByResourceKind is the regression for a warning a real
// `broker remove` printed on every run:
//
//	kubectl auth can-i delete namespaces
//	Warning: resource 'namespaces' is not namespace scoped
//
// Omitting -n was already correct and was not enough. kubectl falls back to the
// KUBECONFIG context's namespace when no -n is given, so the access review still
// carried one -- which printed that warning and, worse, asked a subtly different
// question: with a namespace in the review RBAC also evaluates namespaced Roles
// there, so a RoleBinding granting `delete namespaces` in that one namespace would
// answer YES for a cluster-scoped action the identity cannot actually perform.
// --all-namespaces clears the namespace from the review.
//
// RISK if this regresses: the probe passes and the delete it guards fails, which is
// exactly the half-done state Preflight exists to prevent.
func TestCanIArgsScopesByResourceKind(t *testing.T) {
	c := NewCluster(nil, loadK8s(t), nil, nil)

	cases := []struct {
		p    probe
		want []string
	}{
		{probe{verb: "delete", resource: "namespaces"},
			[]string{"auth", "can-i", "delete", "namespaces", "--all-namespaces"}},
		{probe{verb: "create", resource: "customresourcedefinitions"},
			[]string{"auth", "can-i", "create", "customresourcedefinitions", "--all-namespaces"}},
		{probe{verb: "delete", resource: "secrets"},
			[]string{"auth", "can-i", "delete", "secrets", "-n", "solace"}},
		{probe{verb: "patch", resource: "deployments", ns: "solace-system"},
			[]string{"auth", "can-i", "patch", "deployments", "-n", "solace-system"}},
	}
	for _, tc := range cases {
		got := c.canIArgs(tc.p)
		if strings.Join(got, " ") != strings.Join(tc.want, " ") {
			t.Errorf("canIArgs(%+v) = %v, want %v", tc.p, got, tc.want)
		}
		// Belt and braces: a cluster-scoped probe must never carry -n at all.
		if clusterScoped[tc.p.resource] {
			for _, a := range got {
				if a == "-n" {
					t.Errorf("cluster-scoped probe %+v carried -n: %v", tc.p, got)
				}
			}
		}
	}
}

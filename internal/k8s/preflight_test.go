package k8s

import (
	"context"
	"errors"
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

// TestCanIRefusedReadsEveryRefusalSpelling: kubectl prints a refusal as a bare `no`, as
// `no - <reason>` when the authorizer gave one, and as `no (<evaluation error>)`. Only the
// bare form used to count, so a reasoned refusal read as "could not check" and sent the
// operator to log in again. Something that merely starts with "no" is not a verdict.
func TestCanIRefusedReadsEveryRefusalSpelling(t *testing.T) {
	for _, tc := range []struct {
		answer string
		want   bool
	}{
		{"no", true},
		{"no - RBAC: access denied", true},
		{"no (evaluation error: role not found)", true},
		{"yes", false},
		{"nope", false},
		{"no such host", false},
		{"", false},
	} {
		if got := canIRefused(tc.answer); got != tc.want {
			t.Errorf("canIRefused(%q) = %v, want %v", tc.answer, got, tc.want)
		}
	}
}

// TestPreflightRefusalIsNotAllowed: a refusal -- in any of kubectl's spellings, and even
// though kubectl exits non-zero with it -- wraps errNotAllowed, which is what lets a caller
// degrade a guard only when the identity merely lacks a grant. A probe nobody could
// answer does not wrap it and keeps its log-in advice. The advice names the binding that
// can actually grant the permission: a RoleBinding for a namespaced kind, a cluster role
// binding for a cluster-scoped one, since a RoleBinding cannot reach outside its namespace.
func TestPreflightRefusalIsNotAllowed(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name     string
		rr       *recRunner
		p        probe
		refusal  bool
		contains []string
		absent   string
	}{
		{"namespaced refusal", &recRunner{canI: "no"}, probe{verb: "create", resource: "secrets"},
			true, []string{"not allowed to create secrets in solace", "a role binding granting"}, "cluster role binding"},
		{"reasoned refusal with kubectl's exit status", &recRunner{canI: "no - RBAC: access denied", canIErr: errors.New("exit status 1")},
			probe{verb: "create", resource: "secrets"}, true, []string{"not allowed to create secrets"}, "log in first"},
		{"cluster-scoped refusal", &recRunner{canI: "no"}, probe{verb: "delete", resource: "namespaces"},
			true, []string{"not allowed to delete namespaces in the cluster", "a cluster role binding granting"}, ""},
		{"cluster-wide question about a namespaced kind", &recRunner{canI: "no"},
			probe{verb: "list", resource: brokerResource, allNamespaces: true},
			true, []string{"in the cluster", "a cluster role binding granting list"}, ""},
		{"no answer at all", &recRunner{canIErr: errors.New("connection refused")}, probe{verb: "create", resource: "secrets"},
			false, []string{"cannot check permission", "log in first"}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := NewCluster(tc.rr, loadK8s(t), nil, nil).PreflightAll(ctx, tc.p)
			if err == nil {
				t.Fatal("the probe must fail")
			}
			if got := errors.Is(err, errNotAllowed); got != tc.refusal {
				t.Errorf("errors.Is(err, errNotAllowed) = %v, want %v (err: %v)", got, tc.refusal, err)
			}
			for _, w := range tc.contains {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error = %v, want it to contain %q", err, w)
				}
			}
			if tc.absent != "" && strings.Contains(err.Error(), tc.absent) {
				t.Errorf("error = %v, must not contain %q", err, tc.absent)
			}
		})
	}
}

// TestCanIAnswersWithoutFailing: canI is the path chooser, so it never errors -- it says
// what the cluster answered, and that nobody answered when that is the truth. Every reply
// that would refuse in Preflight is a refusal here, and every reply Preflight could not
// read (an error with no verdict, an unreadable body) is unanswered, never a "no".
func TestCanIAnswersWithoutFailing(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name              string
		rr                *recRunner
		allowed, answered bool
	}{
		{"yes", &recRunner{}, true, true},
		{"warned yes", &recRunner{canI: "Warning: a\nyes"}, true, true},
		{"no", &recRunner{canI: "no"}, false, true},
		{"reasoned no", &recRunner{canI: "no - RBAC: access denied", canIErr: errors.New("exit status 1")}, false, true},
		{"error with no verdict", &recRunner{canIErr: errors.New("connection refused")}, false, false},
		{"unreadable", &recRunner{canI: "maybe"}, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := NewCluster(tc.rr, loadK8s(t), nil, nil)
			p := probe{verb: "delete", resource: "namespaces"}
			allowed, answered := c.canI(ctx, p)
			if allowed != tc.allowed || answered != tc.answered {
				t.Errorf("canI = (%v, %v), want (%v, %v)", allowed, answered, tc.allowed, tc.answered)
			}
			// It asks the same question Preflight asks, so the two cannot drift apart.
			if len(tc.rr.calls) != 1 || !eqArgs(tc.rr.calls[0].args, c.canIArgs(p)) {
				t.Errorf("calls = %+v, want exactly the probe %v", tc.rr.calls, c.canIArgs(p))
			}
		})
	}
}

// TestCanIIsPermittedUnderEcho: a preview shows the question and takes the path a
// permitted run takes, exactly as Preflight passes under the Echo runner.
func TestCanIIsPermittedUnderEcho(t *testing.T) {
	var out strings.Builder
	c := NewCluster(engine.Echo{W: &out}, loadK8s(t), nil, &out)
	allowed, answered := c.canI(context.Background(), probe{verb: "delete", resource: "namespaces"})
	if !allowed || !answered {
		t.Errorf("canI under Echo = (%v, %v), want (true, true)", allowed, answered)
	}
	if !strings.Contains(out.String(), "auth can-i delete namespaces --all-namespaces") {
		t.Errorf("the preview must show the probe:\n%s", out.String())
	}
}

// TestCanIUnderEchoStillHonoursTheExecutionGuard: a preview answers yes only for a probe it
// could echo. A cluster command nobody approved is refused before argv is built, so the
// question was never asked -- unanswered, never a yes -- and nothing is echoed.
func TestCanIUnderEchoStillHonoursTheExecutionGuard(t *testing.T) {
	var out strings.Builder
	c := NewCluster(engine.Echo{W: &out}, unapprovedCfg(), nil, &out)
	allowed, answered := c.canI(context.Background(), probe{verb: "delete", resource: "namespaces"})
	if allowed || answered {
		t.Errorf("canI under Echo with an unapproved command = (%v, %v), want (false, false)", allowed, answered)
	}
	if out.Len() != 0 {
		t.Errorf("a refused command must echo nothing, got:\n%s", out.String())
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
		// A named probe asks about the one object, and is scoped by its bare resource.
		{probe{verb: "bind", resource: "clusterroles", name: operatorClusterRole},
			[]string{"auth", "can-i", "bind", "clusterroles/" + operatorClusterRole, "--all-namespaces"}},
		// A namespaced kind asked about across every namespace: the cluster-wide list the
		// CRD cascade gate needs, without making the kind cluster-scoped for every probe.
		{probe{verb: "list", resource: brokerResource, allNamespaces: true},
			[]string{"auth", "can-i", "list", brokerResource, "--all-namespaces"}},
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

// TestPreflightNamesTheObjectOfANamedProbe: a named probe asks about `resource/name`, keeps
// the --all-namespaces its cluster-scoped resource earns, and a refusal names the object,
// so the grant a cluster admin is asked for is the exact one -- bind can be granted for a
// single role through resourceNames.
func TestPreflightNamesTheObjectOfANamedProbe(t *testing.T) {
	rr := &recRunner{canI: "no"}
	c := NewCluster(rr, loadK8s(t), nil, nil)
	err := c.PreflightAll(context.Background(), probe{verb: "bind", resource: "clusterroles", name: "some-role"})
	if err == nil {
		t.Fatal("a refused named probe must fail")
	}
	want := []string{"auth", "can-i", "bind", "clusterroles/some-role", "--all-namespaces"}
	if len(rr.calls) != 1 || !eqArgs(rr.calls[0].args, want) {
		t.Errorf("calls = %+v, want exactly the probe %v", rr.calls, want)
	}
	for _, w := range []string{"not allowed to bind clusterroles/some-role in the cluster",
		"granting bind on clusterroles/some-role"} {
		if !strings.Contains(err.Error(), w) {
			t.Errorf("error = %v, want it to contain %q", err, w)
		}
	}
}

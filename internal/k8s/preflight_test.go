package k8s

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/kube/rbac"
	"github.com/hafio/go-deploy-utility/runner"
)

// TestPreflightRefusalIsARefusedError: a refusal -- in any of kubectl's spellings, and even
// though kubectl exits non-zero with it -- is a *rbac.RefusedError, which is what lets a
// caller degrade a guard only when the identity merely lacks a grant. A probe nobody could
// answer is not one and keeps its log-in advice. The advice names the binding that can
// actually grant the permission: a RoleBinding for a namespaced kind, a cluster role binding
// for a cluster-scoped one, since a RoleBinding cannot reach outside its namespace. (The
// spellings and the wording themselves are pinned by the shared rbac tests.)
func TestPreflightRefusalIsARefusedError(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name     string
		rr       *recRunner
		p        rbac.Probe
		refusal  bool
		contains []string
		absent   string
	}{
		{"namespaced refusal", &recRunner{canI: "no"}, rbac.Probe{Verb: "create", Resource: "secrets"},
			true, []string{`not allowed to create secrets in namespace "solace"`, "a role binding granting"}, "cluster role binding"},
		{"reasoned refusal with kubectl's exit status", &recRunner{canI: "no - RBAC: access denied", canIErr: errors.New("exit status 1")},
			rbac.Probe{Verb: "create", Resource: "secrets"}, true, []string{"not allowed to create secrets"}, "log in or select"},
		{"cluster-scoped refusal", &recRunner{canI: "no"}, rbac.Probe{Verb: "delete", Resource: "namespaces"},
			true, []string{"not allowed to delete namespaces in the cluster", "a cluster role binding granting"}, ""},
		{"cluster-wide question about a namespaced kind", &recRunner{canI: "no"},
			rbac.Probe{Verb: "list", Resource: brokerResource, AllNamespaces: true},
			true, []string{"in the cluster", "a cluster role binding granting list"}, ""},
		{"no answer at all", &recRunner{canIErr: errors.New("connection refused")}, rbac.Probe{Verb: "create", Resource: "secrets"},
			false, []string{"cannot check permission", "log in or select a working context first"}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := NewCluster(tc.rr, loadK8s(t), nil, nil).PreflightAll(ctx, tc.p)
			if err == nil {
				t.Fatal("the probe must fail")
			}
			var refused *rbac.RefusedError
			if got := errors.As(err, &refused); got != tc.refusal {
				t.Errorf("errors.As(err, *rbac.RefusedError) = %v, want %v (err: %v)", got, tc.refusal, err)
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

// canIArgs is the `auth can-i` argv (after the program) the cluster asks for p: scoped to
// the broker namespace when p is namespaced and names none, as every probe is.
func canIArgs(t *testing.T, c *Cluster, p rbac.Probe) []string {
	t.Helper()
	cmd, err := c.canICmd(c.scope(p))
	if err != nil {
		t.Fatalf("canICmd: %v", err)
	}
	return cmd.Argv[1:]
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
			p := rbac.Probe{Verb: "delete", Resource: "namespaces"}
			allowed, answered := c.canI(ctx, p)
			if allowed != tc.allowed || answered != tc.answered {
				t.Errorf("canI = (%v, %v), want (%v, %v)", allowed, answered, tc.allowed, tc.answered)
			}
			// It asks the same question Preflight asks, so the two cannot drift apart.
			want := canIArgs(t, c, p)
			if len(tc.rr.calls) != 1 || !eqArgs(tc.rr.calls[0].args, want) {
				t.Errorf("calls = %+v, want exactly the probe %v", tc.rr.calls, want)
			}
		})
	}
}

// TestCanIIsPermittedUnderEcho: a preview shows the question and takes the path a
// permitted run takes, exactly as Preflight passes under the Echo runner.
func TestCanIIsPermittedUnderEcho(t *testing.T) {
	var out strings.Builder
	c := NewCluster(runner.Echo{W: &out}, loadK8s(t), nil, &out)
	allowed, answered := c.canI(context.Background(), rbac.Probe{Verb: "delete", Resource: "namespaces"})
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
	c := NewCluster(runner.Echo{W: &out}, unapprovedCfg(), nil, &out)
	allowed, answered := c.canI(context.Background(), rbac.Probe{Verb: "delete", Resource: "namespaces"})
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
	c := NewCluster(runner.Echo{W: &out}, loadK8s(t), nil, &out)
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
		p    rbac.Probe
		want []string
	}{
		{rbac.Probe{Verb: "delete", Resource: "namespaces"},
			[]string{"auth", "can-i", "delete", "namespaces", "--all-namespaces"}},
		{rbac.Probe{Verb: "create", Resource: "customresourcedefinitions"},
			[]string{"auth", "can-i", "create", "customresourcedefinitions", "--all-namespaces"}},
		{rbac.Probe{Verb: "delete", Resource: "secrets"},
			[]string{"auth", "can-i", "delete", "secrets", "-n", "solace"}},
		{rbac.Probe{Verb: "patch", Resource: "deployments", Namespace: "solace-system"},
			[]string{"auth", "can-i", "patch", "deployments", "-n", "solace-system"}},
		// A named probe asks about the one object, and is scoped by its bare resource.
		{rbac.Probe{Verb: "bind", Resource: "clusterroles", Name: operatorClusterRole},
			[]string{"auth", "can-i", "bind", "clusterroles/" + operatorClusterRole, "--all-namespaces"}},
		// A namespaced kind asked about across every namespace: the cluster-wide list the
		// CRD cascade gate needs, without making the kind cluster-scoped for every probe.
		{rbac.Probe{Verb: "list", Resource: brokerResource, AllNamespaces: true},
			[]string{"auth", "can-i", "list", brokerResource, "--all-namespaces"}},
	}
	for _, tc := range cases {
		got := canIArgs(t, c, tc.p)
		if strings.Join(got, " ") != strings.Join(tc.want, " ") {
			t.Errorf("can-i argv for %+v = %v, want %v", tc.p, got, tc.want)
		}
		// Belt and braces: a cluster-scoped probe must never carry -n at all.
		if rbac.ClusterScoped(tc.p.Resource) {
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
	err := c.PreflightAll(context.Background(), rbac.Probe{Verb: "bind", Resource: "clusterroles", Name: "some-role"})
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

package k8s

import (
	"context"
	"io"
	"time"

	"solace/internal/config"
	"solace/internal/engine"
	"solace/internal/output"
)

// defaultOperatorNS is where the operator lands when kubernetes.operator.namespace
// is not configured (000-env.sh:83). It is what both `deploy operator` installs to
// and what every other operator command then addresses, so the two cannot disagree.
const defaultOperatorNS = "pubsubplus-operator-system"

// operatorDeployment is the fixed name of the operator's controller Deployment and
// ServiceAccount (assets/operator-1.4.2.yaml.tmpl).
const operatorDeployment = "pubsubplus-eventbroker-operator"

// operatorDeployRef is the same Deployment in the `<kind>/<name>` form kubectl wants for
// rollout, logs, describe and set env. Written once so the four call sites cannot
// disagree about the prefix.
const operatorDeployRef = "deployment/" + operatorDeployment

// Cluster performs Kubernetes operations that talk to the cluster or the operator --
// as opposed to a running broker, which goes through internal/broker over the
// transport. Every command routes through R, so the Echo runner records it and tests capture
// the exact argv. Out is the report sink; Log is the narration sink.
type Cluster struct {
	R   engine.Runner
	Cfg *config.Config
	// Log is the RAW line sink for progress: it receives one already-formatted
	// line and emits it verbatim. The `==> ` and `[TAG ] ` prefixes are added by
	// the internal/output Sink built over it (progress below), so a call site
	// cannot hand-type a prefix that drifts from the rest of the tool. nil
	// discards.
	Log func(string, ...any)
	Out io.Writer
	// Now is the clock, a seam so the AGE column the status commands render is
	// testable against a fixed instant. nil means time.Now.
	Now func() time.Time

	// Confirm asks the operator a yes/no question, the same seam
	// container.Manager carries. nil DECLINES, which is what an unattended run
	// must do when the question is "may I downgrade a cluster-scoped operator".
	//
	// This is the ONLY question this package asks, and it is a func rather than a
	// reader/writer pair on purpose: internal/cli owns the terminal, so every other
	// confirmation -- including the removal prompts and the namespace question -- is
	// asked there and reaches here as a decision already made. The In/Err pair this
	// struct once carried existed for the interactive node picker and went with it.
	Confirm func(question string) bool
}

// NewCluster builds a Cluster over the given runner, config, line sink and output
// sink. Confirm is left nil, which DECLINES; a caller that can ask sets it.
func NewCluster(r engine.Runner, cfg *config.Config, log func(string, ...any), out io.Writer) *Cluster {
	return &Cluster{R: r, Cfg: cfg, Log: log, Out: out}
}

// confirm asks the operator question through the Confirm seam. A nil Confirm DECLINES:
// the questions this package asks are all "may I do something whose blast radius is
// wider than this env file" -- downgrade a shared operator, widen its watch scope -- and
// an unattended run must answer no to those. Routed through one helper so the
// nil-declines rule cannot be spelled differently at a second call site.
func (c *Cluster) confirm(question string) bool {
	return c.Confirm != nil && c.Confirm(question)
}

// progress is the stderr Sink for this package's narration: phases through Step,
// leveled status through OK/Warn/Fail/Info. A nil Log discards.
func (c *Cluster) progress() *output.Sink { return output.NewFunc(c.Log) }

// report is the stdout Sink for report bodies -- sections, key/value blocks,
// tables and the per-item outcome lines that belong to a report rather than to
// the progress narration.
func (c *Cluster) report() *output.Sink { return output.New(c.out()) }

// logf announces one phase of work (`==> ...`) via the injected line sink.
func (c *Cluster) logf(format string, a ...any) { c.progress().Step(format, a...) }

// ns is the broker namespace.
func (c *Cluster) ns() string { return c.Cfg.K8s.Namespace }

// cmd is the configured cluster CLI (kubernetes.command, default `kubectl`): argv[0]
// plus any leading arguments that precede every call's own. Ported from the bash
// KUBE variable, which the scripts expanded unquoted so it could carry a whole
// profile (`kubectl --kubeconfig <file>`), not just a binary name.
//
// It re-runs the execution guard (config.CheckCommand, via ClusterCommand) on every
// call rather than trusting the value Validate already saw. That is the second of
// the guard's two enforcement points: this package can be driven from a Config that
// never went through config.Load, and a hostile command must be inert there too. The
// check is a few string comparisons over a two-token slice, so running it per call
// costs nothing measurable and removes the "did anyone validate this?" question from
// every call site below.
func (c *Cluster) cmd() (config.Command, error) { return c.Cfg.ClusterCommand() }

// kubectl runs `kubectl args...`, streaming stdout/stderr.
func (c *Cluster) kubectl(ctx context.Context, args ...string) error {
	k, err := c.cmd()
	if err != nil {
		return err
	}
	return c.R.Run(ctx, k.Name(), k.Args(args...)...)
}

// apply pipes a rendered manifest to `kubectl apply -f -` on stdin (never a temp
// file, so secret-bearing manifests stay off disk -- §3).
func (c *Cluster) apply(ctx context.Context, manifest []byte) error {
	k, err := c.cmd()
	if err != nil {
		return err
	}
	return c.R.RunInput(ctx, manifest, k.Name(), k.Args("apply", "-f", "-")...)
}

// deleteStdin pipes a rendered manifest to `kubectl delete -f - --ignore-not-found`,
// so teardown mirrors apply through one code path and is idempotent.
func (c *Cluster) deleteStdin(ctx context.Context, manifest []byte) error {
	k, err := c.cmd()
	if err != nil {
		return err
	}
	return c.R.RunInput(ctx, manifest, k.Name(), k.Args("delete", "-f", "-", "--ignore-not-found")...)
}

// output runs `kubectl args...` and returns captured stdout.
func (c *Cluster) output(ctx context.Context, args ...string) ([]byte, error) {
	k, err := c.cmd()
	if err != nil {
		return nil, err
	}
	return c.R.Output(ctx, k.Name(), k.Args(args...)...)
}

// operatorNS resolves the namespace the operator runs in: the configured
// Operator.Namespace if set, otherwise the fixed default it is installed to.
// Mirrors 000-env.sh:73-83.
func (c *Cluster) operatorNS(ctx context.Context) string {
	ns, _ := c.operatorNSOrigin(ctx)
	return ns
}

// operatorNSOrigin is the one definition of that two-branch rule, and also names
// where the value came from; only the `check` report needs the origin.
//
// It used to have a third branch between them, which searched the cluster: it
// listed Deployments in EVERY namespace and took the namespace of the first line
// CONTAINING the operator's deployment name -- an unanchored substring match,
// with no label, owner or uniqueness check, and its errors swallowed so an RBAC
// denial and an absent operator arrived here identically. On a cluster running
// two operator installs it could resolve to another team's, and `remove operator`
// would then delete the one the operator answering the prompt never saw named.
// So the search is gone. deploy and remove now resolve the SAME namespace from
// the SAME two local rules, and this tool never looks at a namespace the env file
// did not name or the default did not imply.
//
// ctx is retained only to keep this and operatorNS interchangeable with the
// ~13 call sites that already hold one; nothing here reaches the cluster any more,
// which is the point.
func (c *Cluster) operatorNSOrigin(_ context.Context) (ns, origin string) {
	if ns := c.Cfg.K8s.Operator.Namespace; ns != "" {
		return ns, "kubernetes.operator.namespace"
	}
	return defaultOperatorNS, "default -- kubernetes.operator.namespace is unset"
}

// OperatorNamespace is that same rule for a caller holding only a config: the CLI
// names the operator's namespace in its removal prompt, which happens before there
// is a Cluster to ask. Exporting it is only honest now that the answer is two
// local rules -- while it still searched the cluster, no config-only caller could
// have produced the same answer.
func OperatorNamespace(cfg *config.Config) string {
	if ns := cfg.K8s.Operator.Namespace; ns != "" {
		return ns
	}
	return defaultOperatorNS
}

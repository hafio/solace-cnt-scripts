package k8s

import (
	"context"
	"io"
	"time"

	"github.com/hafio/go-deploy-utility/kube"
	"github.com/hafio/go-deploy-utility/runner"

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

// operatorClusterRole is the fixed name of the ClusterRole the operator bundle grants its
// ServiceAccount, which the bundle's fixed-name ClusterRoleBinding refers to
// (assets/operator-1.4.2.yaml.tmpl). The `bind` escalation probe names it; a test pins it
// to the rendered bundle so the two cannot drift.
const operatorClusterRole = "pubsubplus-eventbroker-operator-role"

// operatorDeployRef is the same Deployment in the `<kind>/<name>` form kubectl wants for
// rollout, logs, describe and set env. Written once so the four call sites cannot
// disagree about the prefix.
const operatorDeployRef = "deployment/" + operatorDeployment

// Cluster performs Kubernetes operations that talk to the cluster or the operator --
// as opposed to a running broker, which goes through internal/broker over the
// transport. Every command routes through R, so the Echo runner records it and tests capture
// the exact argv. Out is the report sink; Log is the narration sink.
type Cluster struct {
	R   runner.Runner
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

	// Confirm asks the operator whether to downgrade the cluster-scoped operator, the
	// same seam container.Manager carries. nil DECLINES, which is what a run with no
	// way to ask must do. An error is a question that could not be put (no terminal,
	// no --no-prompt) and is returned as the command's failure, not read as a "no".
	//
	// A func rather than a reader/writer pair on purpose: internal/cli owns the
	// terminal, so every other confirmation -- including the removal prompts and the
	// namespace question -- is asked there and reaches here as a decision already made.
	Confirm func(question string) (bool, error)

	// ConfirmScope asks whether the operator may watch EVERY namespace. It is its own
	// seam because --no-prompt answers Confirm and must never answer this: widening
	// the watch needs a terminal. nil DECLINES; an error is returned as the failure.
	ConfirmScope func(question string) (bool, error)
}

// NewCluster builds a Cluster over the given runner, config, line sink and output
// sink. Confirm is left nil, which DECLINES; a caller that can ask sets it.
func NewCluster(r runner.Runner, cfg *config.Config, log func(string, ...any), out io.Writer) *Cluster {
	return &Cluster{R: engine.NoEnv(r), Cfg: cfg, Log: log, Out: out}
}

// confirm asks the operator question through the Confirm seam. A nil Confirm DECLINES:
// the question this seam carries is "may I do something whose blast radius is wider
// than this env file" -- downgrade a shared operator -- and a run with no way to ask
// must answer no. Routed through one helper so the nil-declines rule cannot be spelled
// differently at a second call site.
func (c *Cluster) confirm(question string) (bool, error) {
	if c.Confirm == nil {
		return false, nil
	}
	return c.Confirm(question)
}

// confirmScope is confirm for the watch-widening question, over ConfirmScope.
func (c *Cluster) confirmScope(question string) (bool, error) {
	if c.ConfirmScope == nil {
		return false, nil
	}
	return c.ConfirmScope(question)
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
// It re-runs the execution guard (execguard.Check, via ClusterCommand) on every
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
	return engine.Run(ctx, c.R, runner.Cmd{Argv: k.Argv(args...)})
}

// build makes one command from the guarded cluster CLI argv: a closure over a builder
// in the shared kube package, which takes that argv as its first argument.
type build func(base []string) runner.Cmd

// base is the guarded cluster CLI as the argv every kube builder starts from (kubectl, or
// oc, plus any --context or --kubeconfig kubernetes.command carries). It re-runs the
// execution guard through cmd, for the reason cmd gives.
func (c *Cluster) base() ([]string, error) {
	k, err := c.cmd()
	if err != nil {
		return nil, err
	}
	return k.Argv(), nil
}

// runKube runs what b builds, streaming stdout/stderr.
func (c *Cluster) runKube(ctx context.Context, b build) error {
	base, err := c.base()
	if err != nil {
		return err
	}
	return engine.Run(ctx, c.R, b(base))
}

// outKube runs what b builds and returns captured stdout, stderr streaming.
func (c *Cluster) outKube(ctx context.Context, b build) ([]byte, error) {
	base, err := c.base()
	if err != nil {
		return nil, err
	}
	return engine.Output(ctx, c.R, b(base))
}

// teeKube runs what b builds, streaming both streams as runKube does, and also returns
// what it wrote to stderr, for a step that must tell a refusal (kube.Forbidden) from any
// other failure. Echo writes nothing there.
func (c *Cluster) teeKube(ctx context.Context, b build) (string, error) {
	base, err := c.base()
	if err != nil {
		return "", err
	}
	return engine.RunTee(ctx, c.R, b(base))
}

// outTeeKube is outKube that also returns stderr, which still reaches the terminal.
func (c *Cluster) outTeeKube(ctx context.Context, b build) ([]byte, string, error) {
	base, err := c.base()
	if err != nil {
		return nil, "", err
	}
	return engine.OutputTee(ctx, c.R, b(base))
}

// apply pipes a rendered manifest to `kubectl apply -f -` on stdin (never a temp
// file, so secret-bearing manifests stay off disk).
func (c *Cluster) apply(ctx context.Context, manifest []byte) error {
	return c.runKube(ctx, func(base []string) runner.Cmd { return kube.Apply(base, manifest) })
}

// deleteStdin pipes a rendered manifest to `kubectl delete -f - --ignore-not-found`,
// so teardown mirrors apply through one code path and is idempotent.
func (c *Cluster) deleteStdin(ctx context.Context, manifest []byte) error {
	return c.runKube(ctx, func(base []string) runner.Cmd { return kube.Delete(base, manifest) })
}

// output runs `kubectl args...` and returns captured stdout.
func (c *Cluster) kubectlOutput(ctx context.Context, args ...string) ([]byte, error) {
	k, err := c.cmd()
	if err != nil {
		return nil, err
	}
	return engine.Output(ctx, c.R, runner.Cmd{Argv: k.Argv(args...)})
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
// There is deliberately no third branch that SEARCHES the cluster for the operator.
// Matching a Deployment by name across every namespace is unanchored and unowned: on a
// cluster running two operator installs it can resolve to another team's, and `remove
// operator` would then delete the one the person answering the prompt never saw named.
// deploy and remove resolve the SAME namespace from the SAME two local rules, and this
// tool never looks at a namespace the env file
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

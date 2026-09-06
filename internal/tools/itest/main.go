// Command itest is a dev-only harness that verifies, against a LIVE environment,
// the handful of claims this repository's source marks as assumed but unverified.
// It is never shipped: it is absent from BUILD_TARGETS and from the release
// workflow, and the `itest` dev-script task builds it on demand.
//
// It exists because the unit suite deliberately cannot answer these questions.
// Every external effect in this codebase is seamed so tests run with no cluster,
// no container engine and no broker -- which is what makes the suite fast and
// hermetic, and also what leaves three comments in the code saying, in effect,
// "this should be true, but nobody has checked". Each probe here checks one, on
// real infrastructure, and reports VERIFIED / REFUTED / INCONCLUSIVE so the
// comment can be replaced by a fact.
//
// It drives the REAL production code paths -- config.Load, engine.Exec,
// k8s.Cluster, container.Manager, broker.Ops -- because a harness that
// reimplemented them would only prove things about itself.
//
// Usage (see docs/itest.md; build with the `itest` task):
//
//	solace-itest --list                      # what each probe touches; needs no env file
//	solace-itest -e prod.yaml --dry-run p1   # read-only steps for real, mutations previewed
//	solace-itest -e prod.yaml p2             # run it, asking before anything mutating
//
// This binary is operator-run ONLY. An AI assistant must never invoke it: it
// mutates live broker state, and the decision to do that to a given environment
// belongs to a person.
//
// Exit codes: 0 nothing refuted, 1 at least one claim REFUTED, 2 usage or setup
// error. An INCONCLUSIVE run exits 0 -- it found no defect, it just did not
// settle the question.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"solace/internal/config"
	"solace/internal/output"
)

// version is stamped by the linker (-X main.version=...) from the same
// `git describe` the build task uses, so a probe result can be pinned to the
// exact code that produced it. It deliberately does not read
// internal/cli.version: this binary does not import internal/cli.
var version = "dev"

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

// flags is the parsed command line, kept as a struct so resolveEnv takes one
// argument and a test can construct it directly.
type flags struct {
	env          string
	baseDir      string
	platform     string
	probe        string
	list         bool
	version      bool
	yes          bool
	dryRun       bool
	verbose      bool
	allowCommand []string
	args         []string // positional probe ids
}

// stringList collects a repeatable flag, matching --allow-command on the CLI.
type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

// parseFlags is separated from run so a test can assert the parse without
// driving the whole binary.
func parseFlags(args []string, errOut io.Writer) (flags, error) {
	var f flags
	var allow stringList
	fs := flag.NewFlagSet("solace-itest", flag.ContinueOnError)
	fs.SetOutput(errOut)
	fs.StringVar(&f.env, "env", config.EnvFileDefault, "env file name or path")
	fs.StringVar(&f.env, "e", config.EnvFileDefault, "env file name or path (shorthand)")
	fs.StringVar(&f.baseDir, "base-dir", "", "directory searched for the env file (default: cwd)")
	fs.StringVar(&f.platform, "platform", "", "kubernetes|docker|podman; needed only when the env file declares more than one")
	fs.StringVar(&f.probe, "probe", "", "comma-separated probe ids to run (default: all)")
	fs.BoolVar(&f.list, "list", false, "list every probe and what it touches, then exit")
	fs.BoolVar(&f.version, "version", false, "print the version this harness was built from, then exit")
	fs.BoolVar(&f.yes, "yes", false, "pre-approve every mutating probe (for an unattended run)")
	fs.BoolVar(&f.dryRun, "dry-run", false, "run read-only steps for real, then stop before anything mutating")
	fs.BoolVar(&f.verbose, "verbose", false, "announce every external command as it runs")
	fs.BoolVar(&f.verbose, "v", false, "announce every external command as it runs (shorthand)")
	fs.Var(&allow, "allow-command", "approve a binary the env file names (repeatable)")
	fs.Usage = func() { fmt.Fprint(errOut, usage) }
	if err := fs.Parse(args); err != nil {
		return f, err
	}
	f.allowCommand = allow
	f.args = fs.Args()
	return f, nil
}

// usage is spelled out rather than left to flag's own dump, so the two lines
// that matter -- operator-run only, start with --list -- are the first thing read.
const usage = "usage: solace-itest [flags] [probe...]\n" +
	"\n" +
	"Verifies, against a LIVE environment, the claims this repo's source marks as\n" +
	"assumed. Operator-run only. Start with --list, then --dry-run.\n" +
	"\n" +
	"  -e, --env FILE       env file name or path (default env.yaml)\n" +
	"      --base-dir DIR   directory searched for the env file\n" +
	"      --platform NAME  kubernetes|docker|podman (only if the file declares several)\n" +
	"      --probe IDS      comma-separated probe ids (same as positional args)\n" +
	"      --list           print every probe and what it touches, then exit\n" +
	"      --version        print the build version, then exit\n" +
	"      --yes            pre-approve every mutating probe (unattended runs)\n" +
	"      --dry-run        real read-only steps, then stop before mutating\n" +
	"  -v, --verbose        announce every external command as it runs\n" +
	"      --allow-command  approve a binary the env file names (repeatable)\n" +
	"\n" +
	"Exit codes: 0 nothing refuted, 1 a claim was REFUTED, 2 usage/setup error.\n"

// run is main's body: parse, resolve, run the selected probes, summarise.
func run(args []string, out, errOut io.Writer) int {
	f, err := parseFlags(args, errOut)
	if err != nil {
		return 2 // the flag package already reported it
	}
	report := output.New(out)
	if f.version {
		fmt.Fprintf(out, "solace-itest %s\n", version)
		return 0
	}
	if f.list {
		listProbes(report)
		return 0
	}

	selected, err := selectProbes(append(splitProbe(f.probe), f.args...))
	if err != nil {
		fmt.Fprintf(errOut, "solace-itest: %v\n", err)
		return 2
	}
	e, err := resolveEnv(f, errOut)
	if err != nil {
		fmt.Fprintf(errOut, "solace-itest: %v\n", err)
		return 2
	}
	if f.dryRun {
		e.Out.Info("dry-run: read-only steps run for real; each probe stops before it would mutate")
	}
	return runProbes(context.Background(), e, selected, report)
}

// runProbes runs each probe and prints the summary. It returns the exit code:
// non-zero only for a REFUTED claim, since that is the one outcome naming a
// defect. Inconclusive and skipped runs are reported plainly and exit 0 -- they
// found nothing wrong, they just did not settle the question, and inflating that
// into a failure would train an operator to ignore the exit code.
func runProbes(ctx context.Context, e *Env, selected []Probe, report *output.Sink) int {
	results := make([]Result, len(selected))
	for i, p := range selected {
		e.Out.Step("%s: %s", p.ID, p.Name)
		e.Out.Info("claim: %s", p.Claim)
		if !p.appliesTo(e.Platform) {
			results[i] = Result{SkippedPlatform,
				fmt.Sprintf("not applicable to %s (runs on %s)", e.Platform, p.platformList())}
		} else {
			e.Out.Info("touches: %s", p.Touches)
			results[i] = p.Run(ctx, e)
		}
		results[i].Verdict.tag(e.Out, "%s: %s", p.ID, results[i].Detail)
	}

	report.Line("")
	report.Section("itest results")
	report.KVRow(resultKeyWidth, "platform", string(e.Platform))
	report.KVRow(resultKeyWidth, "harness", version)
	report.Line("")
	rows := make([][]string, 0, len(selected))
	refuted := 0
	for i, p := range selected {
		if results[i].Verdict == Refuted {
			refuted++
		}
		rows = append(rows, []string{p.ID, verdictName(results[i].Verdict), results[i].Detail})
	}
	report.Table([]string{"PROBE", "VERDICT", "EVIDENCE"}, rows)
	if refuted > 0 {
		return 1
	}
	return 0
}

// resultKeyWidth is the key column of the results header.
const resultKeyWidth = 8

func verdictName(v Verdict) string {
	switch v {
	case Verified:
		return "VERIFIED"
	case Refuted:
		return "REFUTED"
	case Inconclusive:
		return "INCONCLUSIVE"
	case SkippedNoConsent:
		return "SKIPPED(no consent)"
	default:
		return "SKIPPED"
	}
}

// splitProbe folds --probe into the same list the positional args feed, so the
// two spellings cannot disagree.
func splitProbe(s string) []string {
	if s == "" {
		return nil
	}
	return []string{s}
}

package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"solace/internal/broker"
	"solace/internal/config"
	"solace/internal/container"
	"solace/internal/engine"
	"solace/internal/k8s"
	"solace/internal/output"
)

// Verdict is what a probe concluded about the claim it exists to settle.
type Verdict int

const (
	// Verified: the code comment's assumption held, against this environment.
	Verified Verdict = iota
	// Refuted: the assumption is wrong here. This is the valuable outcome -- it
	// names a real defect and the follow-up fix.
	Refuted
	// Inconclusive: the probe ran but the evidence does not settle the claim.
	// Deliberately NOT folded into Refuted: "we did not learn" and "the code is
	// wrong" call for different work.
	Inconclusive
	// SkippedPlatform / SkippedNoConsent: the probe did not run. Neither is a
	// pass; a report that showed them as one would be lying by omission.
	SkippedPlatform
	SkippedNoConsent
)

// tag prints the verdict through the house leveled tags.
func (v Verdict) tag(s *output.Sink, format string, args ...any) {
	switch v {
	case Verified:
		s.OK(format, args...)
	case Refuted:
		s.Fail(format, args...)
	case Inconclusive:
		s.Warn(format, args...)
	default:
		s.Skip(format, args...)
	}
}

// Result is a probe's outcome plus the evidence for it. Detail is printed
// verbatim, so a probe that could not settle its claim puts the raw reply there
// -- an operator adjudicating a SEMP body needs the bytes, not a summary.
type Result struct {
	Verdict Verdict
	Detail  string
}

func verified(format string, a ...any) Result {
	return Result{Verified, fmt.Sprintf(format, a...)}
}
func refuted(format string, a ...any) Result {
	return Result{Refuted, fmt.Sprintf(format, a...)}
}
func inconclusive(format string, a ...any) Result {
	return Result{Inconclusive, fmt.Sprintf(format, a...)}
}
func skipped(format string, a ...any) Result {
	return Result{SkippedPlatform, fmt.Sprintf(format, a...)}
}

// Env is the live environment a probe drives: real production types over a real
// engine.Exec, built once from the env file the operator named. Exactly one of
// Cluster/Manager is non-nil, matching Platform.
type Env struct {
	Cfg      *config.Config
	Platform config.Platform
	Runner   engine.Runner
	Cluster  *k8s.Cluster
	Manager  *container.Manager
	Ops      *broker.Ops
	Out      *output.Sink // stderr narration: phases and verdicts
	DryRun   bool         // --dry-run: stop at the first mutating step
	Consent  Consenter

	// Polling knobs for the probes that wait on a live system (a unit becoming
	// active, config-sync propagating). Fields rather than constants for the same
	// reason Ops.PollInterval is one: a test sets the interval to 0 so it does not
	// sleep, and every wait here is bounded so a probe reports INCONCLUSIVE rather
	// than hanging. resolveEnv sets the production defaults.
	PollInterval time.Duration
	PollAttempts int
}

// Probe is one targeted check of one claim the source code marks as assumed.
// There is deliberately no generic "smoke test everything" probe: this binary
// exists to answer named open questions, and a broad pass/fail sweep would
// answer none of them.
type Probe struct {
	ID        string            // p1..p4, also the selector
	Name      string            // one-line title
	Claim     string            // the source comment this settles, file:line
	Platforms []config.Platform // where it applies; elsewhere it SKIPs
	Mutates   bool              // gates the consent prompt
	Touches   string            // what running it does to the environment
	Run       func(ctx context.Context, e *Env) Result
}

// appliesTo reports whether this probe runs on p.
func (pr Probe) appliesTo(p config.Platform) bool {
	for _, x := range pr.Platforms {
		if x == p {
			return true
		}
	}
	return false
}

// platformList renders the probe's platforms for --list.
func (pr Probe) platformList() string {
	names := make([]string, 0, len(pr.Platforms))
	for _, p := range pr.Platforms {
		names = append(names, string(p))
	}
	return strings.Join(names, ",")
}

// probes is the registry, in run order.
func probes() []Probe {
	return []Probe{probeP1(), probeP2(), probeP4()}
}

// selectProbes resolves the operator's selection. No selection means every
// registered probe -- the ones that do not apply to the resolved platform still
// appear, as SKIP, because a report that silently omitted them would read as
// "everything passed".
func selectProbes(ids []string) ([]Probe, error) {
	all := probes()
	if len(ids) == 0 {
		return all, nil
	}
	byID := make(map[string]Probe, len(all))
	for _, p := range all {
		byID[p.ID] = p
	}
	var out []Probe
	seen := map[string]bool{}
	for _, raw := range ids {
		for _, id := range strings.Split(raw, ",") {
			id = strings.ToLower(strings.TrimSpace(id))
			if id == "" {
				continue
			}
			p, ok := byID[id]
			if !ok {
				return nil, fmt.Errorf("unknown probe %q; known probes are %s (use --list to see what each one touches)",
					id, strings.Join(probeIDs(all), ", "))
			}
			if seen[id] {
				continue // asking twice is not an error, but it runs once
			}
			seen[id] = true
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no probes selected")
	}
	return out, nil
}

func probeIDs(all []Probe) []string {
	ids := make([]string, 0, len(all))
	for _, p := range all {
		ids = append(ids, p.ID)
	}
	sort.Strings(ids)
	return ids
}

// listProbes writes the catalogue. It needs no env file and touches nothing, so
// it is the safe way to find out what a probe would do before pointing it at a
// live broker.
func listProbes(s *output.Sink) {
	s.Section("itest probes")
	for _, p := range probes() {
		mut := "read-only"
		if p.Mutates {
			mut = "MUTATES (asks first)"
		}
		s.Line("")
		s.Line("  %s  %s", p.ID, p.Name)
		s.KVRow(listKeyWidth, "claim", p.Claim)
		s.KVRow(listKeyWidth, "platforms", p.platformList())
		s.KVRow(listKeyWidth, "effect", mut)
		s.KVRow(listKeyWidth, "touches", p.Touches)
	}
}

// listKeyWidth is the key column of the --list catalogue.
const listKeyWidth = 9

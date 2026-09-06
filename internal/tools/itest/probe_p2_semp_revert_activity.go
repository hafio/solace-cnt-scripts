package main

import (
	"context"
	"strings"

	"solace/internal/broker"
	"solace/internal/config"
)

// probeP2 settles semp.go's NEEDS VERIFICATION ON A LIVE BROKER: the /SEMP path,
// whether sempV1OK's regex matches a real broker reply, and whether reverting a
// mate that is ALREADY standby returns code="ok" (idempotent) or a benign error
// that needs a carve-out.
//
// This is the riskiest claim in the codebase to leave unverified. Every container
// HA flow ends by reverting activity to the primary over this one RPC, and its
// reply is only pattern-matched -- so if the regex is wrong, RedundancyCoordinated
// reports a failure after a successful revert, or worse, reports success after a
// failed one, leaving the group backwards.
//
// It is also the claim that looks hardest to test safely, and the reason it is
// testable at all is the idempotence question itself: sending revert-activity to a
// mate that is ALREADY standby should be a no-op. So the probe reads the group's
// state first and only fires when the mate is provably standby -- and then checks
// afterwards that activity did not move, which is the probe verifying its own
// safety assumption rather than trusting it.
func probeP2() Probe {
	return Probe{
		ID:        "p2",
		Name:      "SEMP mate channel: preflight reply, revert-activity path and idempotence",
		Claim:     "internal/broker/semp.go revertActivityMateBody -- NEEDS VERIFICATION ON A LIVE BROKER",
		Platforms: []config.Platform{config.Docker, config.Podman},
		Mutates:   true,
		Touches: "reads `show redundancy` on this host, GETs the mate's SEMP monitor endpoint, and -- only if " +
			"the mate is already standby -- POSTs one redundancy revert-activity to it, which should be a no-op",
		Run: runP2,
	}
}

func runP2(ctx context.Context, e *Env) Result {
	if !e.Cfg.RedundancyEnabled() {
		return skipped("standalone deployment: there is no mate to reach over SEMP")
	}
	role, err := e.Ops.LocalRole("")
	if err != nil {
		return inconclusive("could not determine this host's role: %v", err)
	}
	if role != config.Primary {
		return skipped("this host is the %s; the SEMP mate channel is driven from the primary", roleLabel(role))
	}

	// Read-only half, ungated: this is the preflight the coordinated flows run
	// before their own first mutation, and the claim that it works at all (correct
	// path, correct port resolution through a bridge mapping) is half of what P2
	// exists to settle.
	if err := e.Ops.MateSEMPPreflight(ctx); err != nil {
		return refuted("the mate's SEMP monitor endpoint did not answer, so the coordinated flows would abort "+
			"here (which is the intended behaviour -- but it also means this path is unverified): %v", err)
	}
	e.Out.OK("mate SEMP preflight answered: /SEMP/v2/monitor is reachable and the port resolution is right")

	// Safety gate. Sending revert-activity to an ACTIVE mate is a real failover,
	// which this probe must never cause -- so it reads the primary's own view of
	// who holds activity, with the same parser the failover flows use, and refuses
	// rather than guessing.
	rd, err := e.Ops.ShowRedundancy(ctx, config.Primary)
	if err != nil {
		return inconclusive("could not read `show redundancy` on the primary, so the mate's state is unknown; "+
			"refusing to send revert-activity blind: %v", err)
	}
	if broker.MateActivityState(rd) {
		return inconclusive("the mate is currently ACTIVE. Sending revert-activity now would trigger a real " +
			"failover, which this probe must not do. Retry once the group is in its steady state " +
			"(primary active, mate standby)")
	}
	e.Out.OK("the mate is standby, so revert-activity should be a no-op -- which is the claim under test")

	what := "POST one `redundancy revert-activity` to the mate over SEMP v1 (expected to be a no-op, " +
		"since the mate is standby, but it is a real write RPC to a live broker)"
	if e.wouldMutate(what) {
		return inconclusive("dry-run: stopped before the revert-activity RPC")
	}
	if !e.Consent.Confirm("About to " + what + ".") {
		return Result{SkippedNoConsent, "operator declined the revert-activity RPC"}
	}

	revertErr := e.Ops.MateRevertActivity(ctx)
	res := p2Classify(revertErr)

	// The probe's own regression check: whatever the RPC replied, activity must
	// not have moved. If it did, the "safe because standby" assumption this probe
	// is built on is itself wrong, and that outranks the verdict above.
	if after, err := e.Ops.ShowRedundancy(ctx, config.Primary); err != nil {
		e.Out.Warn("could not re-read `show redundancy` to confirm activity did not move: %v", err)
	} else if broker.MateActivityState(after) {
		return refuted("ACTIVITY MOVED: revert-activity sent to an already-standby mate made it active. "+
			"The idempotence assumption is wrong and the group may now be backwards -- check "+
			"`show redundancy` and revert by hand. Original RPC outcome: %s", res.Detail)
	} else {
		e.Out.OK("activity did not move: the primary still holds it")
	}
	return res
}

// p2Classify turns the RPC outcome into a verdict. Pure function of the error so
// every branch is testable without a broker.
//
// The rejected-reply case is deliberately INCONCLUSIVE rather than REFUTED: the
// broker answered, and whether its answer is a benign "already standby" or a real
// failure is a judgement about SEMP semantics that a human has to make from the
// reply text. So the reply is passed through verbatim for exactly that purpose.
func p2Classify(err error) Result {
	if err == nil {
		return verified("the mate accepted revert-activity while already standby and replied code=\"ok\": " +
			"the /SEMP path is right, sempV1OK's regex matches a real reply, and the RPC is idempotent")
	}
	if strings.Contains(err.Error(), "mate rejected the revert-activity RPC") {
		return inconclusive("the mate answered but did not reply code=\"ok\". Read the reply below and decide "+
			"whether it is a benign already-standby response needing a carve-out in MateRevertActivity, or a "+
			"genuine sempV1OK regex mismatch: %v", err)
	}
	return refuted("the revert-activity RPC did not complete: %v", err)
}

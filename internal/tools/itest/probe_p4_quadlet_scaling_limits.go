package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"solace/internal/config"
)

// probeP4 settles render.go's ASSUMED, NOT VERIFIED pair: that a quadlet unit's
// `PodmanArgs=--cpus=` and `Memory=` are the right directives for the CPU and
// memory the scaling tier picked, and that `Memory=` is even understood by the
// podman on the target host.
//
// If either is wrong the broker runs unconstrained -- the tier's sizing silently
// does nothing -- which is the kind of defect that shows up as a production
// incident rather than an error message. render.go's own comment already names
// the fallback: fold Memory= into the PodmanArgs line.
//
// This is the one read-only probe. It inspects an ALREADY-DEPLOYED container
// rather than deploying one, because the question is what podman did with the
// unit we render, and deploying on the probe's behalf would answer a different
// question (what podman does with a unit we rendered just now, on a host in
// whatever state this left it).
func probeP4() Probe {
	return Probe{
		ID:        "p4",
		Name:      "quadlet cpu/memory directives actually constrain the container",
		Claim:     "internal/render/render.go quadlet cpu/mem -- ASSUMED, NOT VERIFIED",
		Platforms: []config.Platform{config.Podman},
		Mutates:   false,
		Touches:   "runs `podman inspect` against the already-deployed container; changes nothing",
		Run:       runP4,
	}
}

func runP4(ctx context.Context, e *Env) Result {
	if err := e.Manager.Preflight(ctx); err != nil {
		return inconclusive("podman preflight failed: %v", err)
	}
	wantCPU := e.Cfg.Scaling.CPU
	wantMem := e.Cfg.ContainerBlock(config.Podman).Mem
	if wantCPU == "" && wantMem == "" {
		return skipped("this env file renders neither a cpu nor a memory limit, so there is nothing to verify")
	}

	rt, err := e.Cfg.RuntimeCommand(config.Podman)
	if err != nil {
		return inconclusive("podman runtime command refused by the execution guard: %v", err)
	}
	name := ctrName(e)
	// Full JSON, not a narrow --format: which field podman populates for cpu and
	// memory is precisely the unknown here, so the probe reads the document and
	// reports what it found rather than asking for a field that may not exist.
	raw, err := e.Runner.Output(ctx, rt.Name(), rt.Args("inspect", name)...)
	if err != nil {
		return inconclusive("`podman inspect %s` failed (is the broker deployed on this host?): %v", name, err)
	}
	got, err := p4Parse(raw)
	if err != nil {
		return inconclusive("could not read podman inspect output: %v", err)
	}
	return p4Classify(wantCPU, wantMem, got)
}

// p4Limits is what podman reported about the container's resource caps. A nil
// field means podman's inspect document did not carry it at all, which is a
// different finding from "carried it, and it was zero".
type p4Limits struct {
	NanoCPUs  *int64
	CPUQuota  *int64
	CPUPeriod *int64
	Memory    *int64
}

// p4Parse pulls the resource caps out of a `podman inspect` array. The field
// names are docker-compatible ones podman has historically mirrored; the probe
// does not assume which of them is populated.
func p4Parse(raw []byte) (p4Limits, error) {
	var docs []struct {
		HostConfig struct {
			NanoCpus  *int64 `json:"NanoCpus"`
			CpuQuota  *int64 `json:"CpuQuota"`
			CpuPeriod *int64 `json:"CpuPeriod"`
			Memory    *int64 `json:"Memory"`
		} `json:"HostConfig"`
	}
	if err := json.Unmarshal(raw, &docs); err != nil {
		return p4Limits{}, err
	}
	if len(docs) == 0 {
		return p4Limits{}, fmt.Errorf("inspect returned no container documents")
	}
	h := docs[0].HostConfig
	return p4Limits{NanoCPUs: h.NanoCpus, CPUQuota: h.CpuQuota, CPUPeriod: h.CpuPeriod, Memory: h.Memory}, nil
}

// p4Classify compares what the env file renders against what podman applied. It
// is a pure function so every branch is testable without podman.
//
// Each half is judged separately and the worse verdict wins: a run where the CPU
// cap took and the memory cap did not is exactly the split render.go's comment
// anticipates, and collapsing it to one answer would hide which directive to fix.
func p4Classify(wantCPU, wantMem string, got p4Limits) Result {
	var findings []string
	worst := Verified

	demote := func(v Verdict) {
		if v > worst {
			worst = v
		}
	}

	if wantCPU != "" {
		switch {
		case got.NanoCPUs != nil && *got.NanoCPUs > 0:
			wantNano, err := cpuToNano(wantCPU)
			if err != nil {
				findings = append(findings, fmt.Sprintf("cpu: cannot parse the configured %q: %v", wantCPU, err))
				demote(Inconclusive)
			} else if *got.NanoCPUs == wantNano {
				findings = append(findings, fmt.Sprintf("cpu: applied as NanoCpus=%d, matching %s", *got.NanoCPUs, wantCPU))
			} else {
				findings = append(findings, fmt.Sprintf("cpu: MISMATCH -- configured %s (%d nanocpus) but podman applied %d",
					wantCPU, wantNano, *got.NanoCPUs))
				demote(Refuted)
			}
		case got.CPUQuota != nil && *got.CPUQuota > 0 && got.CPUPeriod != nil && *got.CPUPeriod > 0:
			findings = append(findings, fmt.Sprintf("cpu: applied as CpuQuota=%d/CpuPeriod=%d (%.3f cpus) against configured %s",
				*got.CPUQuota, *got.CPUPeriod, float64(*got.CPUQuota)/float64(*got.CPUPeriod), wantCPU))
		default:
			findings = append(findings, fmt.Sprintf("cpu: configured %s but podman reports NO cpu cap -- "+
				"PodmanArgs=--cpus= may not be reaching podman run; check the cgroup directly "+
				"(cat /sys/fs/cgroup/.../cpu.max) before changing render.go", wantCPU))
			demote(Refuted)
		}
	}

	if wantMem != "" {
		switch {
		case got.Memory != nil && *got.Memory > 0:
			wantBytes, err := memToBytes(wantMem)
			if err != nil {
				findings = append(findings, fmt.Sprintf("memory: cannot parse the configured %q: %v", wantMem, err))
				demote(Inconclusive)
			} else if *got.Memory == wantBytes {
				findings = append(findings, fmt.Sprintf("memory: applied as Memory=%d, matching %s", *got.Memory, wantMem))
			} else {
				findings = append(findings, fmt.Sprintf("memory: MISMATCH -- configured %s (%d bytes) but podman applied %d",
					wantMem, wantBytes, *got.Memory))
				demote(Refuted)
			}
		default:
			findings = append(findings, fmt.Sprintf("memory: configured %s but podman reports NO memory cap -- "+
				"this podman may not support the quadlet Memory= key, in which case it must fold into "+
				"PodmanArgs=--memory= (render.go names this fallback)", wantMem))
			demote(Refuted)
		}
	}

	return Result{worst, strings.Join(findings, "; ")}
}

// cpuToNano converts a Kubernetes-style CPU quantity to podman's nanocpus. The
// scaling tier writes either a plain core count ("2") or millicores ("500m").
func cpuToNano(s string) (int64, error) {
	const nanosPerCPU = 1_000_000_000
	if m, ok := strings.CutSuffix(strings.TrimSpace(s), "m"); ok {
		milli, err := strconv.ParseFloat(m, 64)
		if err != nil {
			return 0, err
		}
		return int64(milli * nanosPerCPU / 1000), nil
	}
	cpus, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, err
	}
	return int64(cpus * nanosPerCPU), nil
}

// memToBytes converts the engines' own b|k|m|g suffix (what container.mem takes,
// deliberately NOT the Mi/Gi Kubernetes spelling) to bytes.
func memToBytes(s string) (int64, error) {
	t := strings.ToLower(strings.TrimSpace(s))
	mult := int64(1)
	switch {
	case strings.HasSuffix(t, "g"):
		mult, t = 1<<30, strings.TrimSuffix(t, "g")
	case strings.HasSuffix(t, "m"):
		mult, t = 1<<20, strings.TrimSuffix(t, "m")
	case strings.HasSuffix(t, "k"):
		mult, t = 1<<10, strings.TrimSuffix(t, "k")
	case strings.HasSuffix(t, "b"):
		t = strings.TrimSuffix(t, "b")
	}
	n, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		return 0, err
	}
	return n * mult, nil
}

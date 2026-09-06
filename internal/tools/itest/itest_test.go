package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"solace/internal/config"
	"solace/internal/output"
)

// sinkOver is a Sink over a buffer -- the shape these tests assert on.
func sinkOver(buf *bytes.Buffer) *output.Sink { return output.New(buf) }

// testEnv is a minimal Env whose narration lands in the returned buffer. It has
// no Cluster/Manager/Ops: the tests here cover the parts that do not need a live
// environment, which is deliberately all of the decision logic.
func testEnv() (*Env, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return &Env{
		Cfg:      &config.Config{},
		Platform: config.Podman,
		Out:      output.New(buf),
		Consent:  Consenter{Interactive: func() bool { return false }, Err: &bytes.Buffer{}},
	}, buf
}

// --- registry / selection ------------------------------------------------------

// TestRegistryCoversEveryClaim is the guard on this binary's whole reason to
// exist: each registered probe must name the source comment it settles, and no
// two probes may share an id (the selector would then be ambiguous). A probe
// added without a Claim is a probe nobody can trace back to a question.
func TestRegistryCoversEveryClaim(t *testing.T) {
	seen := map[string]bool{}
	for _, p := range probes() {
		if p.ID == "" || seen[p.ID] {
			t.Errorf("probe %q has an empty or duplicate id", p.ID)
		}
		seen[p.ID] = true
		if p.Claim == "" {
			t.Errorf("probe %s has no Claim: a probe must name the comment it settles", p.ID)
		}
		if p.Touches == "" {
			t.Errorf("probe %s has no Touches: an operator must be able to read what it does first", p.ID)
		}
		if len(p.Platforms) == 0 {
			t.Errorf("probe %s declares no platforms, so it would SKIP everywhere", p.ID)
		}
		if p.Run == nil {
			t.Errorf("probe %s has no Run func", p.ID)
		}
	}
	if len(seen) != 3 {
		t.Errorf("registry has %d probes, want the 3 documented in docs/itest.md", len(seen))
	}
}

func TestSelectProbes(t *testing.T) {
	all := probes()
	t.Run("no selection runs everything", func(t *testing.T) {
		got, err := selectProbes(nil)
		if err != nil || len(got) != len(all) {
			t.Fatalf("selectProbes(nil) = %d probes, err %v; want %d", len(got), err, len(all))
		}
	})
	t.Run("positional and comma forms agree", func(t *testing.T) {
		a, err := selectProbes([]string{"p1", "p4"})
		if err != nil {
			t.Fatalf("positional: %v", err)
		}
		b, err := selectProbes([]string{"p1,p4"})
		if err != nil {
			t.Fatalf("comma: %v", err)
		}
		if len(a) != 2 || len(b) != 2 || a[0].ID != b[0].ID || a[1].ID != b[1].ID {
			t.Errorf("the two spellings disagree: %v vs %v", ids(a), ids(b))
		}
	})
	t.Run("case and spacing tolerated", func(t *testing.T) {
		got, err := selectProbes([]string{" P2 , p1 "})
		if err != nil {
			t.Fatalf("err %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %v, want p2 and p1", ids(got))
		}
	})
	t.Run("asking twice runs once", func(t *testing.T) {
		got, err := selectProbes([]string{"p1", "p1"})
		if err != nil || len(got) != 1 {
			t.Errorf("selectProbes(p1,p1) = %v, err %v; want one probe", ids(got), err)
		}
	})
	t.Run("unknown id names the known ones", func(t *testing.T) {
		_, err := selectProbes([]string{"p9"})
		if err == nil || !strings.Contains(err.Error(), "p1") {
			t.Errorf("err = %v, want it to list the known probe ids", err)
		}
	})
}

func ids(ps []Probe) []string {
	out := make([]string, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.ID)
	}
	return out
}

// TestListProbesNeedsNoEnvFile pins the property that makes --list the safe first
// command: it renders the whole catalogue, including what each probe touches,
// without resolving or reading an env file.
func TestListProbesNeedsNoEnvFile(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"--list"}, &out, &errOut); code != 0 {
		t.Fatalf("--list exit = %d, want 0; stderr: %s", code, errOut.String())
	}
	got := out.String()
	for _, want := range []string{"p1", "p2", "p4", "MUTATES", "read-only", "touches"} {
		if !strings.Contains(got, want) {
			t.Errorf("--list output missing %q:\n%s", want, got)
		}
	}
	if errOut.Len() != 0 {
		t.Errorf("--list wrote to stderr: %q", errOut.String())
	}
}

func TestVersionFlag(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"--version"}, &out, &errOut); code != 0 {
		t.Fatalf("--version exit = %d, want 0", code)
	}
	if !strings.Contains(out.String(), version) {
		t.Errorf("--version = %q, want it to name the stamped version %q", out.String(), version)
	}
}

// TestUnknownProbeIsAUsageError proves a mistyped selector fails BEFORE the env
// file is resolved -- so a typo cannot be reported as "cannot reach the cluster".
func TestUnknownProbeIsAUsageError(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"--probe", "nope", "-e", "definitely-not-here.yaml"}, &out, &errOut); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut.String(), "unknown probe") {
		t.Errorf("stderr = %q, want the unknown-probe message rather than an env-file error", errOut.String())
	}
}

// --- env resolution ------------------------------------------------------------

func writeEnv(t *testing.T, body string) (dir, name string) {
	t.Helper()
	dir = t.TempDir()
	name = "probe.yaml"
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir, name
}

func TestResolvePlatform(t *testing.T) {
	const podmanOnly = "podman:\n  rootless: true\n"
	const both = "docker: {}\npodman:\n  rootless: true\n"
	const none = "image:\n  name: x\n"

	t.Run("sole declared section is used silently", func(t *testing.T) {
		dir, name := writeEnv(t, podmanOnly)
		got, err := resolvePlatform(filepath.Join(dir, name), "")
		if err != nil || got != config.Podman {
			t.Errorf("got %q, err %v; want podman", got, err)
		}
	})
	t.Run("several sections demand --platform and never prompt", func(t *testing.T) {
		dir, name := writeEnv(t, both)
		_, err := resolvePlatform(filepath.Join(dir, name), "")
		if err == nil || !strings.Contains(err.Error(), "--platform") {
			t.Fatalf("err = %v, want it to name --platform", err)
		}
		if !strings.Contains(err.Error(), "unattended") {
			t.Errorf("err = %v, want it to say why it does not prompt", err)
		}
	})
	t.Run("--platform must name a declared section", func(t *testing.T) {
		dir, name := writeEnv(t, podmanOnly)
		_, err := resolvePlatform(filepath.Join(dir, name), "docker")
		if err == nil || !strings.Contains(err.Error(), "no docker") {
			t.Errorf("err = %v, want a refusal naming the missing section", err)
		}
	})
	t.Run("--platform accepts the abbreviations", func(t *testing.T) {
		dir, name := writeEnv(t, podmanOnly)
		got, err := resolvePlatform(filepath.Join(dir, name), "pm")
		if err != nil || got != config.Podman {
			t.Errorf("got %q, err %v; want podman from the pm abbreviation", got, err)
		}
	})
	t.Run("no platform section at all is a loud error", func(t *testing.T) {
		dir, name := writeEnv(t, none)
		_, err := resolvePlatform(filepath.Join(dir, name), "")
		if err == nil || !strings.Contains(err.Error(), "no platform section") {
			t.Errorf("err = %v, want it to say no section is declared", err)
		}
	})
}

// --- consent -------------------------------------------------------------------

func TestConsent(t *testing.T) {
	yes := func() bool { return true }
	no := func() bool { return false }

	t.Run("--yes approves without asking", func(t *testing.T) {
		var asked bytes.Buffer
		c := Consenter{Yes: true, Interactive: no, In: strings.NewReader(""), Err: &asked}
		if !c.Confirm("do the thing") {
			t.Error("--yes must approve")
		}
		if asked.Len() != 0 {
			t.Errorf("--yes must not ask; asked %q", asked.String())
		}
	})
	t.Run("exact yes approves", func(t *testing.T) {
		c := Consenter{Interactive: yes, In: strings.NewReader("yes\n"), Err: &bytes.Buffer{}}
		if !c.Confirm("do the thing") {
			t.Error("an exact 'yes' must approve")
		}
	})
	t.Run("anything looser declines", func(t *testing.T) {
		for _, answer := range []string{"y\n", "Y\n", "YES\n", "yes please\n", "\n", "no\n"} {
			c := Consenter{Interactive: yes, In: strings.NewReader(answer), Err: &bytes.Buffer{}}
			if c.Confirm("do the thing") {
				t.Errorf("answer %q must NOT approve a live mutation", answer)
			}
		}
	})
	t.Run("non-interactive without --yes declines", func(t *testing.T) {
		c := Consenter{Interactive: no, In: strings.NewReader("yes\n"), Err: &bytes.Buffer{}}
		if c.Confirm("do the thing") {
			t.Error("with nobody to ask, the answer must be no")
		}
	})
	t.Run("the question says it is live", func(t *testing.T) {
		var asked bytes.Buffer
		c := Consenter{Interactive: yes, In: strings.NewReader("yes\n"), Err: &asked}
		c.Confirm("About to rotate a secret.")
		if !strings.Contains(asked.String(), "LIVE") || !strings.Contains(asked.String(), "'yes'") {
			t.Errorf("prompt = %q, want it to name the risk and the exact word required", asked.String())
		}
	})
}

// --- dry-run gate --------------------------------------------------------------

// TestWouldMutateStopsOnlyUnderDryRun covers the gate every probe leans on: it
// must be inert in a normal run and must stop (and say what it would have done)
// under --dry-run.
func TestWouldMutateStopsOnlyUnderDryRun(t *testing.T) {
	e, buf := testEnv()
	if e.wouldMutate("rotate a secret") {
		t.Error("wouldMutate must be false when --dry-run is off")
	}
	if buf.Len() != 0 {
		t.Errorf("a normal run must say nothing here; got %q", buf.String())
	}

	e.DryRun = true
	if !e.wouldMutate("rotate a secret") {
		t.Error("wouldMutate must stop the probe under --dry-run")
	}
	if !strings.Contains(buf.String(), "would rotate a secret") {
		t.Errorf("dry-run must report what it would do; got %q", buf.String())
	}
}

// --- verdict plumbing ----------------------------------------------------------

// TestVerdictTagsMatchTheHouseLevels pins the mapping from verdict to output tag.
// A refuted claim must read as a failure and an inconclusive one must not: they
// call for different work, and the tag is how an operator tells them apart.
func TestVerdictTagsMatchTheHouseLevels(t *testing.T) {
	for _, tc := range []struct {
		v    Verdict
		want string
	}{
		{Verified, "[ OK ]"},
		{Refuted, "[FAIL]"},
		{Inconclusive, "[WARN]"},
		{SkippedPlatform, "[SKIP]"},
		{SkippedNoConsent, "[SKIP]"},
	} {
		_, buf := testEnv()
		s := sinkOver(buf)
		tc.v.tag(s, "detail")
		if !strings.HasPrefix(buf.String(), tc.want) {
			t.Errorf("verdict %v tagged %q, want prefix %q", tc.v, buf.String(), tc.want)
		}
	}
}

func TestVerdictNames(t *testing.T) {
	for v, want := range map[Verdict]string{
		Verified: "VERIFIED", Refuted: "REFUTED", Inconclusive: "INCONCLUSIVE",
		SkippedNoConsent: "SKIPPED(no consent)", SkippedPlatform: "SKIPPED",
	} {
		if got := verdictName(v); got != want {
			t.Errorf("verdictName(%v) = %q, want %q", v, got, want)
		}
	}
}

// TestRunProbesSkipsOffPlatformAndExitsOnRefuted covers the two properties of the
// summary that matter: a probe that does not apply is REPORTED as skipped rather
// than omitted (an omitted probe reads as a pass), and only a REFUTED claim moves
// the exit code.
func TestRunProbesSkipsOffPlatformAndExitsOnRefuted(t *testing.T) {
	ran := false
	onPodman := Probe{ID: "x", Name: "podman only", Claim: "c", Touches: "t",
		Platforms: []config.Platform{config.Podman},
		Run:       func(_ context.Context, _ *Env) Result { ran = true; return verified("ok") }}

	e, _ := testEnv()
	e.Platform = config.Docker
	var out bytes.Buffer
	code := runProbes(context.Background(), e, []Probe{onPodman}, sinkOver(&out))
	if ran {
		t.Error("an off-platform probe must not run")
	}
	if code != 0 {
		t.Errorf("a skipped probe must not fail the run; exit = %d", code)
	}
	if !strings.Contains(out.String(), "SKIPPED") || !strings.Contains(out.String(), "x") {
		t.Errorf("the summary must still list the skipped probe:\n%s", out.String())
	}

	bad := Probe{ID: "y", Name: "refuter", Claim: "c", Touches: "t",
		Platforms: []config.Platform{config.Docker},
		Run:       func(_ context.Context, _ *Env) Result { return refuted("the assumption is wrong") }}
	out.Reset()
	if code := runProbes(context.Background(), e, []Probe{bad}, sinkOver(&out)); code != 1 {
		t.Errorf("a REFUTED claim must exit 1; got %d", code)
	}
	if !strings.Contains(out.String(), "REFUTED") {
		t.Errorf("summary missing the verdict:\n%s", out.String())
	}

	// Inconclusive is not a defect: it must not fail the run.
	meh := Probe{ID: "z", Name: "unsure", Claim: "c", Touches: "t",
		Platforms: []config.Platform{config.Docker},
		Run:       func(_ context.Context, _ *Env) Result { return inconclusive("could not tell") }}
	out.Reset()
	if code := runProbes(context.Background(), e, []Probe{meh}, sinkOver(&out)); code != 0 {
		t.Errorf("an INCONCLUSIVE run must exit 0; got %d", code)
	}
}

// --- P1 classifier -------------------------------------------------------------

// TestP1Classify is the verdict rule for the podman secret-rotation question,
// exercised as a pure function so every outcome is covered without podman.
func TestP1Classify(t *testing.T) {
	const rotated, original = "newvalue", "oldvalue"
	t.Run("rotated value found is verified", func(t *testing.T) {
		got := p1Classify(rotated+"\n", rotated, original)
		if got.Verdict != Verified {
			t.Errorf("verdict = %v (%s), want Verified", got.Verdict, got.Detail)
		}
	})
	t.Run("original value found is refuted and names the fix", func(t *testing.T) {
		got := p1Classify(original+"\n", rotated, original)
		if got.Verdict != Refuted {
			t.Fatalf("verdict = %v (%s), want Refuted", got.Verdict, got.Detail)
		}
		if !strings.Contains(got.Detail, "force-recreate") {
			t.Errorf("a refuted verdict must name the follow-up fix; got %q", got.Detail)
		}
	})
	t.Run("neither value is inconclusive, not refuted", func(t *testing.T) {
		got := p1Classify("something else", rotated, original)
		if got.Verdict != Inconclusive {
			t.Errorf("verdict = %v (%s), want Inconclusive", got.Verdict, got.Detail)
		}
	})
	t.Run("surrounding whitespace does not change the answer", func(t *testing.T) {
		if got := p1Classify("  "+rotated+" \r\n", rotated, original); got.Verdict != Verified {
			t.Errorf("verdict = %v, want Verified despite whitespace", got.Verdict)
		}
	})
}

func TestP1ThrowawayIsUsable(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		v, err := p1Throwaway()
		if err != nil {
			t.Fatalf("p1Throwaway: %v", err)
		}
		if seen[v] {
			t.Fatalf("p1Throwaway repeated a value: %q", v)
		}
		seen[v] = true
		if strings.ContainsAny(v, " \t\r\n\"'$&|;<>()") {
			t.Errorf("throwaway value %q carries a character the broker CLI or a shell could treat specially", v)
		}
	}
}

// --- P2 classifier -------------------------------------------------------------

// TestP2Classify covers the three SEMP outcomes, and in particular that a
// rejected reply is INCONCLUSIVE with the body passed through: whether an
// already-standby mate's refusal is benign is a judgement a human makes from the
// reply text, so the text has to survive.
func TestP2Classify(t *testing.T) {
	t.Run("nil error verifies path, regex and idempotence", func(t *testing.T) {
		got := p2Classify(nil)
		if got.Verdict != Verified {
			t.Fatalf("verdict = %v (%s), want Verified", got.Verdict, got.Detail)
		}
		for _, want := range []string{"/SEMP", "idempotent"} {
			if !strings.Contains(got.Detail, want) {
				t.Errorf("detail %q should name %q", got.Detail, want)
			}
		}
	})
	t.Run("rejected reply is inconclusive and keeps the body", func(t *testing.T) {
		body := `<rpc-reply><execute-result code="fail" reason="already standby"/></rpc-reply>`
		err := fmt.Errorf("mate rejected the revert-activity RPC; its reply:\n%s", body)
		got := p2Classify(err)
		if got.Verdict != Inconclusive {
			t.Fatalf("verdict = %v, want Inconclusive -- a broker that answered is not a proven defect", got.Verdict)
		}
		if !strings.Contains(got.Detail, "already standby") {
			t.Errorf("the raw reply must survive into the detail for a human to adjudicate; got %q", got.Detail)
		}
		if !strings.Contains(got.Detail, "carve-out") {
			t.Errorf("detail should name both candidate explanations; got %q", got.Detail)
		}
	})
	t.Run("transport failure is refuted", func(t *testing.T) {
		got := p2Classify(errors.New("curl: (7) connection refused"))
		if got.Verdict != Refuted {
			t.Errorf("verdict = %v (%s), want Refuted", got.Verdict, got.Detail)
		}
	})
}

// --- P3 classifier -------------------------------------------------------------

// --- P4 parsing and classification --------------------------------------------

func TestP4Parse(t *testing.T) {
	t.Run("reads the docker-compatible fields", func(t *testing.T) {
		got, err := p4Parse([]byte(`[{"HostConfig":{"NanoCpus":2000000000,"Memory":1073741824}}]`))
		if err != nil {
			t.Fatalf("p4Parse: %v", err)
		}
		if got.NanoCPUs == nil || *got.NanoCPUs != 2000000000 {
			t.Errorf("NanoCPUs = %v, want 2000000000", got.NanoCPUs)
		}
		if got.Memory == nil || *got.Memory != 1073741824 {
			t.Errorf("Memory = %v, want 1073741824", got.Memory)
		}
	})
	t.Run("absent fields stay nil, distinct from zero", func(t *testing.T) {
		got, err := p4Parse([]byte(`[{"HostConfig":{}}]`))
		if err != nil {
			t.Fatalf("p4Parse: %v", err)
		}
		if got.NanoCPUs != nil || got.Memory != nil {
			t.Errorf("absent fields must be nil, got cpu=%v mem=%v", got.NanoCPUs, got.Memory)
		}
	})
	t.Run("empty array and malformed input error", func(t *testing.T) {
		if _, err := p4Parse([]byte(`[]`)); err == nil {
			t.Error("an empty inspect array must error rather than report no limits")
		}
		if _, err := p4Parse([]byte(`not json`)); err == nil {
			t.Error("malformed JSON must error")
		}
	})
}

func TestP4Classify(t *testing.T) {
	i64 := func(v int64) *int64 { return &v }

	t.Run("both applied and matching is verified", func(t *testing.T) {
		got := p4Classify("2", "1g", p4Limits{NanoCPUs: i64(2_000_000_000), Memory: i64(1 << 30)})
		if got.Verdict != Verified {
			t.Errorf("verdict = %v (%s), want Verified", got.Verdict, got.Detail)
		}
	})
	t.Run("millicores convert", func(t *testing.T) {
		got := p4Classify("500m", "", p4Limits{NanoCPUs: i64(500_000_000)})
		if got.Verdict != Verified {
			t.Errorf("verdict = %v (%s), want Verified for 500m", got.Verdict, got.Detail)
		}
	})
	t.Run("memory absent is refuted and names the PodmanArgs fallback", func(t *testing.T) {
		got := p4Classify("", "2g", p4Limits{})
		if got.Verdict != Refuted {
			t.Fatalf("verdict = %v, want Refuted", got.Verdict)
		}
		if !strings.Contains(got.Detail, "PodmanArgs") {
			t.Errorf("detail must name render.go's stated fallback; got %q", got.Detail)
		}
	})
	t.Run("cpu absent is refuted and suggests checking the cgroup first", func(t *testing.T) {
		got := p4Classify("2", "", p4Limits{})
		if got.Verdict != Refuted {
			t.Fatalf("verdict = %v, want Refuted", got.Verdict)
		}
		if !strings.Contains(got.Detail, "cgroup") {
			t.Errorf("detail should point at the manual cgroup check; got %q", got.Detail)
		}
	})
	t.Run("mismatch reports both numbers", func(t *testing.T) {
		got := p4Classify("2", "", p4Limits{NanoCPUs: i64(1_000_000_000)})
		if got.Verdict != Refuted {
			t.Fatalf("verdict = %v, want Refuted", got.Verdict)
		}
		if !strings.Contains(got.Detail, "MISMATCH") {
			t.Errorf("detail = %q, want it to show configured vs applied", got.Detail)
		}
	})
	t.Run("quota/period form is reported, not failed", func(t *testing.T) {
		got := p4Classify("2", "", p4Limits{CPUQuota: i64(200000), CPUPeriod: i64(100000)})
		if got.Verdict != Verified {
			t.Errorf("verdict = %v (%s): a cap expressed as quota/period is still a cap", got.Verdict, got.Detail)
		}
	})
	t.Run("the worse half wins", func(t *testing.T) {
		// cpu fine, memory missing -- the run must not read as verified.
		got := p4Classify("2", "1g", p4Limits{NanoCPUs: i64(2_000_000_000)})
		if got.Verdict != Refuted {
			t.Errorf("verdict = %v (%s), want Refuted because memory was not applied", got.Verdict, got.Detail)
		}
		if !strings.Contains(got.Detail, "cpu:") || !strings.Contains(got.Detail, "memory:") {
			t.Errorf("both halves must be reported; got %q", got.Detail)
		}
	})
	t.Run("unparseable configured values are inconclusive, not refuted", func(t *testing.T) {
		got := p4Classify("what", "", p4Limits{NanoCPUs: i64(1)})
		if got.Verdict != Inconclusive {
			t.Errorf("verdict = %v (%s), want Inconclusive", got.Verdict, got.Detail)
		}
	})
}

func TestCPUAndMemoryConversion(t *testing.T) {
	for in, want := range map[string]int64{"1": 1_000_000_000, "2": 2_000_000_000, "500m": 500_000_000, " 1.5 ": 1_500_000_000} {
		got, err := cpuToNano(in)
		if err != nil || got != want {
			t.Errorf("cpuToNano(%q) = %d, %v; want %d", in, got, err, want)
		}
	}
	if _, err := cpuToNano("abc"); err == nil {
		t.Error("cpuToNano must reject a non-numeric value")
	}
	for in, want := range map[string]int64{"1g": 1 << 30, "512m": 512 << 20, "2048k": 2048 << 10, "1024b": 1024, "4096": 4096} {
		got, err := memToBytes(in)
		if err != nil || got != want {
			t.Errorf("memToBytes(%q) = %d, %v; want %d", in, got, err, want)
		}
	}
	if _, err := memToBytes("lots"); err == nil {
		t.Error("memToBytes must reject a non-numeric value")
	}
}

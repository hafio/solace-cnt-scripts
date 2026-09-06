package main

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
)

// The happy paths through the full probe bodies, plus main's own plumbing. These
// complete the coverage of every decision a live run makes: the earlier files
// cover the classifiers and the refusals, these cover "the assumption held".

// sempOKReply is a SEMP v1 success reply, headers included, in the shape
// sempV1OK's regex expects. It is HAND-AUTHORED: no live broker has been seen
// yet, which is the entire reason probe p2 exists. Replace it with a real capture
// after the first live run, and move it into internal/broker's testdata so the
// regex test uses the same bytes.
const sempOKReply = "HTTP/1.1 200 OK\r\n\r\n" +
	`<rpc-reply semp-version="soltr/10_10"><execute-result code="ok"/></rpc-reply>`

// --- P2 full body --------------------------------------------------------------

// TestRunP2VerifiesAnIdempotentRevert walks P2's success path: the preflight
// answers, the mate is standby, the RPC replies ok, and activity has not moved.
func TestRunP2VerifiesAnIdempotentRevert(t *testing.T) {
	f := &fakeRunner{responder: func(_ string, args []string, stdin []byte) ([]byte, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "-Apes"):
			return []byte("Activity Status : Local Active\n"), nil
		case strings.Contains(joined, "curl") && strings.Contains(string(stdin), "revert-activity"):
			return []byte(sempOKReply), nil
		case strings.Contains(joined, "curl"):
			return []byte("HTTP/1.1 200 OK\r\n\r\n{}"), nil
		}
		return nil, nil
	}}
	e, buf := podmanEnv(t, f)
	e.Ops.Hostname = func() (string, error) { return "pri-host", nil }

	res := runP2(context.Background(), e)
	if res.Verdict != Verified {
		t.Fatalf("verdict = %v (%s), want Verified", res.Verdict, res.Detail)
	}
	if !strings.Contains(buf.String(), "activity did not move") {
		t.Errorf("the probe must confirm its own safety assumption; narration was %q", buf.String())
	}
}

// TestRunP2RefutesWhenActivityMoves is the alarm case: if revert-activity sent to
// a standby mate made it active, the probe's own premise was wrong, and that
// outranks whatever the RPC replied.
func TestRunP2RefutesWhenActivityMoves(t *testing.T) {
	shows := 0
	f := &fakeRunner{responder: func(_ string, args []string, _ []byte) ([]byte, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "-Apes"):
			shows++
			if shows == 1 {
				return []byte("Activity Status : Local Active\n"), nil // standby before
			}
			return []byte("Activity Status : Mate Active\n"), nil // ... active after
		case strings.Contains(joined, "curl"):
			return []byte(sempOKReply), nil
		}
		return nil, nil
	}}
	e, _ := podmanEnv(t, f)
	e.Ops.Hostname = func() (string, error) { return "pri-host", nil }

	res := runP2(context.Background(), e)
	if res.Verdict != Refuted {
		t.Fatalf("verdict = %v (%s), want Refuted", res.Verdict, res.Detail)
	}
	if !strings.Contains(res.Detail, "ACTIVITY MOVED") {
		t.Errorf("the verdict must say activity moved; got %q", res.Detail)
	}
}

// TestRunP2SkipsOnDeclinedConsent covers the gate between the read-only half and
// the RPC.
func TestRunP2SkipsOnDeclinedConsent(t *testing.T) {
	f := &fakeRunner{responder: func(_ string, args []string, _ []byte) ([]byte, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.Contains(joined, "-Apes"):
			return []byte("Activity Status : Local Active\n"), nil
		case strings.Contains(joined, "curl"):
			return []byte("HTTP/1.1 200 OK\r\n\r\n{}"), nil
		}
		return nil, nil
	}}
	e, _ := podmanEnv(t, f)
	e.Ops.Hostname = func() (string, error) { return "pri-host", nil }
	e.Consent = Consenter{Interactive: func() bool { return true }, In: strings.NewReader("nope\n"), Err: &bytes.Buffer{}}

	res := runP2(context.Background(), e)
	if res.Verdict != SkippedNoConsent {
		t.Fatalf("verdict = %v (%s), want SkippedNoConsent", res.Verdict, res.Detail)
	}
	for _, c := range f.calls {
		if strings.Contains(c.stdin, "revert-activity") {
			t.Errorf("a declined probe must not send the RPC:\n%s", f.transcript())
		}
	}
}

// --- run() end to end ----------------------------------------------------------

// TestRunEndToEndDryRun drives main's body against a real env file on disk: probe
// selection, --dry-run, and the summary table. Nothing live is touched because
// every probe stops at its dry-run gate.
func TestRunEndToEndDryRun(t *testing.T) {
	dir, name := writeEnv(t, minimalPodmanEnv)
	var out, errOut bytes.Buffer
	code := run([]string{"-e", name, "--base-dir", dir, "--dry-run", "--yes", "p4"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("exit = %d, want 0; stderr:\n%s", code, errOut.String())
	}
	report := out.String()
	for _, want := range []string{"=== itest results", "PROBE", "VERDICT", "p4"} {
		if !strings.Contains(report, want) {
			t.Errorf("summary missing %q:\n%s", want, report)
		}
	}
	if !strings.Contains(errOut.String(), "dry-run") {
		t.Errorf("a dry run must announce itself on stderr; got %q", errOut.String())
	}
	// The report is stdout and the narration is stderr, so a script capturing
	// stdout gets the verdicts without the commentary.
	if strings.Contains(report, "==>") {
		t.Errorf("progress narration leaked into the report:\n%s", report)
	}
}

// TestRunReportsSetupFailureAsUsage proves an env file that will not load exits 2
// with the tool's own message, rather than running probes against a half-built Env.
func TestRunReportsSetupFailureAsUsage(t *testing.T) {
	dir, name := writeEnv(t, "podman: {}\n") // section declared, validation fails
	var out, errOut bytes.Buffer
	if code := run([]string{"-e", name, "--base-dir", dir}, &out, &errOut); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut.String(), "solace-itest:") {
		t.Errorf("stderr = %q, want the tool to name itself in the failure", errOut.String())
	}
}

// --- flag plumbing -------------------------------------------------------------

func TestParseFlagsCollectsRepeatableAllowCommand(t *testing.T) {
	f, err := parseFlags([]string{"--allow-command", "podman", "--allow-command", "systemctl", "p1"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parseFlags: %v", err)
	}
	if len(f.allowCommand) != 2 || f.allowCommand[0] != "podman" || f.allowCommand[1] != "systemctl" {
		t.Errorf("allowCommand = %v, want both values in order", f.allowCommand)
	}
	if len(f.args) != 1 || f.args[0] != "p1" {
		t.Errorf("positional args = %v, want [p1]", f.args)
	}
	var l stringList
	if err := l.Set("x"); err != nil || l.String() != "x" {
		t.Errorf("stringList round-trip = %q, err %v", l.String(), err)
	}
}

func TestParseFlagsRejectsAnUnknownFlag(t *testing.T) {
	var errOut bytes.Buffer
	if _, err := parseFlags([]string{"--nope"}, &errOut); err == nil {
		t.Error("an unknown flag must be an error")
	}
	if errOut.Len() == 0 {
		t.Error("the failure must be explained on stderr")
	}
}

func TestIsTTYIsFalseForANonTerminal(t *testing.T) {
	if isTTY(nil) {
		t.Error("a nil file is not a terminal")
	}
	f, err := os.CreateTemp(t.TempDir(), "notty")
	if err != nil {
		t.Fatal(err)
	}
	if isTTY(f) {
		t.Error("a regular file is not a terminal")
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if isTTY(f) {
		t.Error("a closed file cannot be stat'd, so it is not a terminal")
	}
}

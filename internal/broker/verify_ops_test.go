package broker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"solace/internal/config"
)

// verify_ops_test.go covers the failure branches of internal/broker/verify_ops.go that
// the happy-path tests in broker_test.go and the branch tests in coverage_test.go leave
// open: a redundancy poll whose `show redundancy` read itself fails, and every step of the
// diagnostics gather that follows its scripts reaching the broker. It adds only what those
// files do not already assert; the release, revert and download branches stay where they are.

// covUploadFailsTo wraps a fakeTransport but makes Upload fail only for one destination,
// so a test can let the first of a node's two script uploads succeed and refuse the second.
// Every call is still recorded, and the other methods are inherited.
type covUploadFailsTo struct {
	*fakeTransport
	dest string
	err  error
}

func (t *covUploadFailsTo) Upload(ctx context.Context, role config.Role, data []byte, dest string) error {
	_ = t.fakeTransport.Upload(ctx, role, data, dest) // still record the call
	if dest == t.dest {
		return t.err
	}
	return nil
}

// TestReleaseToBackupStopsWhenAPollReadFails covers the two polls of releaseToBackup when
// the `show redundancy` read behind them fails, which is a different fault from the poll
// timing out. A read that errors says nothing about whether the Primary released, so the
// error is returned as itself rather than retried into a misleading timeout -- and it
// stops the sequence where it is: nothing past the failed read may be sent, and what was
// already sent stays on record so the failure can be read for how far the drill got.
func TestReleaseToBackupStopsWhenAPollReadFails(t *testing.T) {
	boom := errors.New("show redundancy failed")
	released := rd("Primary", "Enabled-Released", "Down", "Mate Active")

	t.Run("while waiting for the release", func(t *testing.T) {
		ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
			if matchCLI(argv, "show-rd") {
				return nil, boom
			}
			return nil, nil
		}}
		o, _ := newTestOps(t, &config.Config{}, ft)
		err := o.releaseToBackup(context.Background())
		if !errors.Is(err, boom) {
			t.Fatalf("err = %v, want the failed read itself, not a timeout", err)
		}
		if hasCall(ft, "no-release") {
			t.Error("a poll that could not read the Primary must not go on to send no-release")
		}
	})

	t.Run("while waiting for the un-release", func(t *testing.T) {
		reads := 0
		ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
			if matchCLI(argv, "show-rd") {
				reads++
				if reads > 1 {
					return nil, boom
				}
				return []byte(released), nil
			}
			return nil, nil
		}}
		o, _ := newTestOps(t, &config.Config{}, ft)
		err := o.releaseToBackup(context.Background())
		if !errors.Is(err, boom) {
			t.Fatalf("err = %v, want the failed read itself, not a timeout", err)
		}
		if !hasCall(ft, "no-release") {
			t.Error("the un-release was already sent before the read that failed")
		}
	})
}

// TestRevertToPrimaryStopsWhenThePollReadFails is the same distinction for the final poll:
// the revert has been sent, and a `show redundancy` pair that cannot be read is returned as
// the read failure, not waited out into "timeout waiting for Primary to become active" --
// which would tell an operator the revert did not take when nothing is known about it.
func TestRevertToPrimaryStopsWhenThePollReadFails(t *testing.T) {
	boom := errors.New("show redundancy failed")
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "show-rd") {
			return nil, boom
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)

	err := o.revertToPrimary(context.Background())
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want the failed read itself, not a timeout", err)
	}
	if !hasCall(ft, "revert-activity") {
		t.Error("the revert was already sent before the read that failed")
	}
}

// TestGatherNodeUploadFailureReturnsTheErrorAndRunsNothing covers a node whose gather
// scripts cannot be put on the broker. Either upload failing ends that node's collection
// with the upload's own error, before anything has been run there: a half-installed pair of
// scripts is not something to execute.
func TestGatherNodeUploadFailureReturnsTheErrorAndRunsNothing(t *testing.T) {
	boom := errors.New("upload boom")
	zipDest := JailRoot + "/zip-configs.sh"
	for _, tc := range []struct {
		name string
		dest string
	}{
		{"the gather script", cliScriptPath(cliGatherConfigs)},
		{"the zip script", zipDest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ft := &fakeTransport{}
			ut := &covUploadFailsTo{fakeTransport: ft, dest: tc.dest, err: boom}
			o := &Ops{T: ut, Cfg: &config.Config{}, Out: &bytes.Buffer{}}

			err := o.Diagnostics(context.Background(), filepath.Join(t.TempDir(), "diag"), "ts", 3, config.Primary)
			if !errors.Is(err, boom) {
				t.Fatalf("err = %v, want the upload failure itself", err)
			}
			if len(ft.outputs) != 0 {
				t.Errorf("nothing may be run on the broker after a failed upload, got %v", ft.outputs)
			}
		})
	}
}

// TestGatherNodeReturnsTheFailureOfEachStepAndStillCleansUp covers the steps that run once
// the scripts are on the broker: the CLI gather, the zip, the hostname read that names the
// local archive, and the removal of the intermediate files. Each failure is the node's
// failure -- a gather that stopped halfway is not a collection -- and each still removes the
// two uploaded scripts, which is what the deferral exists for: before it, a failure anywhere
// after the upload left both behind on the broker.
func TestGatherNodeReturnsTheFailureOfEachStepAndStillCleansUp(t *testing.T) {
	boom := errors.New("step failed")
	isGather := func(argv []string) bool { return matchCLI(argv, cliGatherConfigs) }
	isHostname := func(argv []string) bool { return len(argv) == 1 && argv[0] == "hostname" }
	isZip := func(argv []string) bool { return len(argv) > 0 && argv[0] == "bash" }
	isArchiveRemoval := func(argv []string) bool {
		return len(argv) > 1 && argv[0] == "rm" && argv[1] == "-rf" &&
			strings.Contains(strings.Join(argv, " "), "gather-configs.zip")
	}
	cases := []struct {
		name    string
		failOut func([]string) bool
		failRun func([]string) bool
	}{
		{name: "the CLI gather", failOut: isGather},
		{name: "the zip", failRun: isZip},
		{name: "the hostname read", failOut: isHostname},
		{name: "the removal of the intermediate files", failRun: isArchiveRemoval},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
				if tc.failOut != nil && tc.failOut(argv) {
					return nil, boom
				}
				if isHostname(argv) {
					return []byte("node1\n"), nil
				}
				return nil, nil
			}}
			rt := &runErrMatchTransport{fakeTransport: ft, err: boom, match: func(argv []string) bool {
				return tc.failRun != nil && tc.failRun(argv)
			}}
			o := &Ops{T: rt, Cfg: &config.Config{}, Out: &bytes.Buffer{}}

			err := o.Diagnostics(context.Background(), filepath.Join(t.TempDir(), "diag"), "ts", 3, config.Primary)
			if !errors.Is(err, boom) {
				t.Fatalf("err = %v, want the failed step itself", err)
			}
			if !ranContains(ft, "rm", "-f", cliScriptPath(cliGatherConfigs), JailRoot+"/zip-configs.sh") {
				t.Errorf("both uploaded scripts must be removed on the way out of a failure, runs = %v", ft.runs)
			}
		})
	}
}

// TestGatherNodeCleanupFailureOnlyWarns proves the removal of the uploaded scripts is
// best-effort. By the time it runs the diagnostics have been collected and downloaded, so
// failing the node over a leftover script would report a failure that did not happen; it is a
// warning that names what was left behind.
func TestGatherNodeCleanupFailureOnlyWarns(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if len(argv) == 1 && argv[0] == "hostname" {
			return []byte("node1\n"), nil
		}
		return nil, nil
	}}
	rt := &runErrMatchTransport{fakeTransport: ft, err: errors.New("rm boom"), match: func(argv []string) bool {
		return len(argv) > 1 && argv[0] == "rm" && argv[1] == "-f"
	}}
	var logs []string
	o := &Ops{T: rt, Cfg: &config.Config{}, Out: &bytes.Buffer{},
		Log: func(f string, a ...any) { logs = append(logs, fmt.Sprintf(f, a...)) }}

	dest := filepath.Join(t.TempDir(), "diag")
	if err := o.Diagnostics(context.Background(), dest, "20260731", 3, config.Primary); err != nil {
		t.Fatalf("a failed removal of the gather scripts must not fail a gather that succeeded: %v", err)
	}
	if !strings.Contains(strings.Join(logs, "\n"), "[WARN] cleanup of the gather scripts") {
		t.Errorf("want a warning that the gather scripts could not be removed, got %v", logs)
	}
}

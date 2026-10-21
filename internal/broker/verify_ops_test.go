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
// open: a redundancy poll whose `show redundancy` read itself fails, every step of the
// diagnostics gather that follows its scripts reaching the broker, a node failing among
// several, and the broker-reported names that gather turns into paths. It adds only what those
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
	archiveDest := JailRoot + "/archive-configs.sh"
	for _, tc := range []struct {
		name string
		dest string
	}{
		{"the gather script", cliScriptPath(cliGatherConfigs)},
		{"the archive script", archiveDest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ft := &fakeTransport{}
			ut := &covUploadFailsTo{fakeTransport: ft, dest: tc.dest, err: boom}
			o := &Ops{T: ut, Cfg: &config.Config{}, Out: &bytes.Buffer{}}

			err := o.Diagnostics(context.Background(), filepath.Join(t.TempDir(), "diag"), "ts", DiagnosticsOptions{Days: 3}, config.Primary)
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
// the scripts are on the broker: the CLI gather, the hostname read that names the archive
// and the directory inside it, the archive, and the removal of the intermediate files. Each
// failure is the node's failure -- a gather that stopped halfway is not a collection -- and
// each still removes the two uploaded scripts, which is what the deferral exists for: before
// it, a failure anywhere after the upload left both behind on the broker. A failure in the
// show-command archive's own steps no longer costs the bundle: it is still downloaded,
// because returning first left it on the broker with nothing naming it. Only a failed CLI
// gather, which produced nothing, fetches no bundle.
func TestGatherNodeReturnsTheFailureOfEachStepAndStillCleansUp(t *testing.T) {
	boom := errors.New("step failed")
	isGather := func(argv []string) bool { return matchCLI(argv, cliGatherConfigs) }
	isHostname := func(argv []string) bool { return len(argv) == 1 && argv[0] == "hostname" }
	isArchive := func(argv []string) bool { return len(argv) > 0 && argv[0] == "bash" }
	isArchiveRemoval := func(argv []string) bool {
		return len(argv) > 1 && argv[0] == "rm" && argv[1] == "-rf" &&
			strings.Contains(strings.Join(argv, " "), "gather-configs.tgz")
	}
	const bundle = JailRoot + "/logs/diag-node1.tgz"
	cases := []struct {
		name       string
		failOut    func([]string) bool
		failRun    func([]string) bool
		wantBundle bool
	}{
		{name: "the CLI gather", failOut: isGather},
		{name: "the hostname read", failOut: isHostname, wantBundle: true},
		{name: "the archive", failRun: isArchive, wantBundle: true},
		{name: "the removal of the intermediate files", failRun: isArchiveRemoval, wantBundle: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
				if tc.failOut != nil && tc.failOut(argv) {
					return nil, boom
				}
				if isGather(argv) {
					return []byte("Diagnostics saved: logs/diag-node1.tgz\n"), nil
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

			err := o.Diagnostics(context.Background(), filepath.Join(t.TempDir(), "diag"), "ts", DiagnosticsOptions{Days: 3}, config.Primary)
			if !errors.Is(err, boom) {
				t.Fatalf("err = %v, want the failed step itself", err)
			}
			if !ranContains(ft, "rm", "-f", cliScriptPath(cliGatherConfigs), JailRoot+"/archive-configs.sh") {
				t.Errorf("both uploaded scripts must be removed on the way out of a failure, runs = %v", ft.runs)
			}
			fetched := false
			for _, d := range ft.downloads {
				fetched = fetched || d.remote == bundle
			}
			if fetched != tc.wantBundle {
				t.Errorf("bundle downloaded = %v, want %v; downloads = %v", fetched, tc.wantBundle, ft.downloads)
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
	if err := o.Diagnostics(context.Background(), dest, "20260731", DiagnosticsOptions{Days: 3}, config.Primary); err != nil {
		t.Fatalf("a failed removal of the gather scripts must not fail a gather that succeeded: %v", err)
	}
	if !strings.Contains(strings.Join(logs, "\n"), "[WARN] cleanup of the gather scripts") {
		t.Errorf("want a warning that the gather scripts could not be removed, got %v", logs)
	}
}

// TestDiagnosticsGathersEveryNodeDespiteAFailure: one node failing no longer stops the
// rest. Returning on the first failure meant a refused primary left the backup and monitor
// ungathered; now each node is attempted, and the error says which failed, how many of how
// many, and where what was collected went.
func TestDiagnosticsGathersEveryNodeDespiteAFailure(t *testing.T) {
	boom := errors.New("cli boom")
	ft := &fakeTransport{responder: func(role config.Role, argv []string, _ []byte) ([]byte, error) {
		switch {
		case matchCLI(argv, cliGatherConfigs) && role == config.Primary:
			return nil, boom
		case matchCLI(argv, cliGatherConfigs):
			return []byte("Diagnostics saved: logs/diag-" + role.Word() + ".tgz\n"), nil
		case len(argv) == 1 && argv[0] == "hostname":
			return []byte("node-" + role.Word() + "\n"), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	dest := filepath.Join(t.TempDir(), "diag")

	err := o.Diagnostics(context.Background(), dest, "ts", DiagnosticsOptions{Days: 3}, config.Primary, config.Backup, config.Monitor)
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want the primary's failure itself", err)
	}
	for _, want := range []string{"primary node", "1 of 3", dest} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, want it to carry %q", err, want)
		}
	}
	for _, role := range []config.Role{config.Backup, config.Monitor} {
		got := map[string]bool{}
		for _, d := range ft.downloads {
			if d.role == role {
				got[d.remote] = true
			}
		}
		if !got[JailRoot+"/gather-configs.tgz"] || !got[JailRoot+"/logs/diag-"+role.Word()+".tgz"] {
			t.Errorf("the %s node must still be gathered after the primary failed; downloads = %v", role.Word(), ft.downloads)
		}
	}
}

// TestGatherNodeRefusesAnUnexpectedBundlePath: the bundle path is broker output, and it
// reaches both a local file name and an `rm -rf` on the broker. Anything but a plain
// logs/<file> is refused by value, neither downloaded nor removed -- `logs/../x` would land
// outside the diagnostics dir and `logs/` would remove the broker's whole logs directory --
// while the show-command archive, which does not depend on it, is still collected.
func TestGatherNodeRefusesAnUnexpectedBundlePath(t *testing.T) {
	for _, diag := range []string{"logs/../x", "logs/", "/etc/x", "other/x", "logs/a b", "logs/sub/x"} {
		t.Run(diag, func(t *testing.T) {
			ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
				switch {
				case matchCLI(argv, cliGatherConfigs):
					return []byte("Diagnostics saved: " + diag + "\n"), nil
				case len(argv) == 1 && argv[0] == "hostname":
					return []byte("node1\n"), nil
				}
				return nil, nil
			}}
			o, _ := newTestOps(t, &config.Config{}, ft)

			err := o.Diagnostics(context.Background(), filepath.Join(t.TempDir(), "diag"), "ts", DiagnosticsOptions{Days: 3}, config.Primary)
			if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("%q", diag)) {
				t.Fatalf("err = %v, want a refusal quoting %q", err, diag)
			}
			if len(ft.downloads) != 1 || ft.downloads[0].remote != JailRoot+"/gather-configs.tgz" {
				t.Errorf("only the show-command archive may be downloaded, got %v", ft.downloads)
			}
			for _, r := range ft.runs {
				if len(r.argv) > 2 && r.argv[0] == "rm" && strings.Contains(strings.Join(r.argv, " "), JailRoot+"/"+diag) {
					t.Errorf("a refused bundle path must not be removed, ran %v", r.argv)
				}
			}
		})
	}
}

// TestGatherNodeRefusesAnUnsafeArchiveName: the broker's host name and the timestamp name
// both the local archive and the directory the in-broker script moves the output into, so a
// name that is not plain is refused before anything is archived, rather than joined onto
// the diagnostics dir or handed to a `mv`. An empty host name stays allowed, as it always
// was -- it only shortens the name and cannot leave the dir.
func TestGatherNodeRefusesAnUnsafeArchiveName(t *testing.T) {
	for _, tc := range []struct{ host, ts, want string }{
		{"../x", "ts", "broker hostname"},
		{"a/b", "ts", "broker hostname"},
		{"a b", "ts", "broker hostname"},
		{"node1", "../x", "show-command archive name"},
	} {
		t.Run(tc.host+" "+tc.ts, func(t *testing.T) {
			ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
				if len(argv) == 1 && argv[0] == "hostname" {
					return []byte(tc.host + "\n"), nil
				}
				return nil, nil
			}}
			o, _ := newTestOps(t, &config.Config{}, ft)

			err := o.Diagnostics(context.Background(), filepath.Join(t.TempDir(), "diag"), tc.ts, DiagnosticsOptions{Days: 3}, config.Primary)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want a refusal naming the %s", err, tc.want)
			}
			if len(ft.downloads) != 0 {
				t.Errorf("nothing may be downloaded under an unsafe name, got %v", ft.downloads)
			}
			for _, r := range ft.runs {
				if len(r.argv) > 0 && r.argv[0] == "bash" {
					t.Errorf("nothing may be archived under an unsafe name, ran %v", r.argv)
				}
			}
		})
	}
}

// TestDiagnosticsFetchesAnEncryptedBundle: --encrypt drops no-encrypt from the broker's
// command, and the bundle it reports -- a .tgz.p7m only Solace support can open -- is
// downloaded under its own name and removed from the broker like a readable one.
func TestDiagnosticsFetchesAnEncryptedBundle(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		switch {
		case matchCLI(argv, cliGatherConfigs):
			return []byte("Diagnostics saved: logs/diag-node1.tgz.p7m\n"), nil
		case len(argv) == 1 && argv[0] == "hostname":
			return []byte("node1\n"), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	dest := filepath.Join(t.TempDir(), "diag")

	if err := o.Diagnostics(context.Background(), dest, "ts", DiagnosticsOptions{Days: 2, Encrypt: true}, config.Primary); err != nil {
		t.Fatalf("Diagnostics: %v", err)
	}
	body := ft.uploadBody(t, cliScriptPath(cliGatherConfigs))
	if !strings.Contains(body, "gather-diagnostics days-of-history 2\n") || strings.Contains(body, "no-encrypt") {
		t.Errorf("an encrypted gather must not carry no-encrypt:\n%s", body)
	}
	remote := JailRoot + "/logs/diag-node1.tgz.p7m"
	fetched := false
	for _, d := range ft.downloads {
		fetched = fetched || (d.remote == remote && d.local == filepath.Join(dest, "diag-node1.tgz.p7m"))
	}
	if !fetched {
		t.Errorf("the encrypted bundle was not downloaded under its own name; downloads = %v", ft.downloads)
	}
	if !ranContains(ft, "rm", "-rf", remote) {
		t.Errorf("the downloaded bundle must be removed from the broker, runs = %v", ft.runs)
	}
}

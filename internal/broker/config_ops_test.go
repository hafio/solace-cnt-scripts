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

// config_ops_test.go covers the ops in internal/broker/config_ops.go that had no test of
// their own: ExecShellScript, the one op there with no bash ancestor; the enable
// directions of the two default-VPN / default-user ops, whose disable twins are tested in
// broker_test.go and coverage_test.go; and the order DomainCerts loads in.
//
// Everything runs over the shared fakeTransport (broker_test.go), so each test asserts
// the argv and the files the op put on the broker, not a mock of its own implementation.

// TestExecShellScriptUploadsRunsWithBashAndRemovesIt pins the whole shape of the op.
//
// The script lands at the jail root, never in cliscripts -- the CLI reads every `.*.cli`
// there relative to its own working directory, so a bash script parked beside them is one
// typo from being fed to the Solace CLI. It runs as `bash <file>` directly, with no
// stop-on-error skeleton, its output is shown in full, and the uploaded file is removed
// afterwards because whatever it carries is as sensitive as whatever the operator put in it.
func TestExecShellScriptUploadsRunsWithBashAndRemovesIt(t *testing.T) {
	local := filepath.Join(t.TempDir(), "job.sh")
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if len(argv) > 0 && argv[0] == "bash" {
			return []byte("script output\n"), nil
		}
		return nil, nil
	}}
	o, out := newTestOps(t, &config.Config{}, ft)
	var logs []string
	o.Log = func(f string, a ...any) { logs = append(logs, fmt.Sprintf(f, a...)) }

	if err := o.ExecShellScript(context.Background(), config.Primary, local); err != nil {
		t.Fatalf("ExecShellScript: %v", err)
	}

	dest := shellScriptPath("job.sh")
	if len(ft.uploadFiles) != 1 {
		t.Fatalf("uploadFiles = %v, want exactly the one script", ft.uploadFiles)
	}
	up := ft.uploadFiles[0]
	if up.role != config.Primary || up.local != local || up.dest != dest {
		t.Errorf("upload = %+v, want %q streamed from disk to %q on the primary", up, local, dest)
	}
	if strings.HasPrefix(dest, CLIScriptsDir+"/") {
		t.Errorf("dest %q is inside cliscripts, where the Solace CLI would read it", dest)
	}

	if len(ft.outputs) != 1 {
		t.Fatalf("outputs = %v, want exactly the one bash exec", ft.outputs)
	}
	wantArgv := "bash " + dest
	if got := strings.Join(ft.outputs[0].argv, " "); got != wantArgv {
		t.Errorf("exec argv = %q, want %q: bash runs the uploaded file directly", got, wantArgv)
	}
	if !strings.Contains(out.String(), "script output") {
		t.Errorf("the script's output is shown in full, got %q", out.String())
	}
	if !ft.removed(dest) {
		t.Errorf("the uploaded script must be removed afterwards, runs = %v", ft.runs)
	}
	if !strings.Contains(strings.Join(logs, "\n"), `Running shell script "job.sh"`) {
		t.Errorf("the run is announced before it starts, got %v", logs)
	}
}

// TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator proves one env file
// names one in-broker file whichever OS drives the run. filepath.Base is OS-dependent, so
// a path written with `\` would have produced two different broker filenames for the same
// input; config.BaseName splits on both. The local path itself travels untouched, because
// that is the file the host really has.
func TestExecShellScriptNamesTheRemoteFileByBaseNameOnEitherSeparator(t *testing.T) {
	for _, local := range []string{`C:\ops\job.sh`, "/opt/ops/job.sh", "job.sh"} {
		ft := &fakeTransport{}
		o, _ := newTestOps(t, &config.Config{}, ft)
		if err := o.ExecShellScript(context.Background(), config.Primary, local); err != nil {
			t.Fatalf("ExecShellScript(%q): %v", local, err)
		}
		if len(ft.uploadFiles) != 1 || ft.uploadFiles[0].dest != shellScriptPath("job.sh") {
			t.Errorf("ExecShellScript(%q) uploaded %v, want the broker file named by the base name alone",
				local, ft.uploadFiles)
			continue
		}
		if ft.uploadFiles[0].local != local {
			t.Errorf("local path = %q, want %q passed through untouched", ft.uploadFiles[0].local, local)
		}
	}
}

// TestExecShellScriptRefusesAnUnsafeNameBeforeUploading is the injection boundary: the
// basename becomes a path inside the broker and an argv operand of `bash`, so a space, a
// shell metacharacter, a `..` or an empty name is refused before the broker is touched --
// not uploaded and then run.
func TestExecShellScriptRefusesAnUnsafeNameBeforeUploading(t *testing.T) {
	for _, local := range []string{
		"/tmp/bad name.sh",
		"/tmp/a;rm.sh",
		"/tmp/$(id).sh",
		"/tmp/..",
		"/tmp/a..b.sh",
		"/tmp/",
	} {
		ft := &fakeTransport{}
		o, _ := newTestOps(t, &config.Config{}, ft)
		err := o.ExecShellScript(context.Background(), config.Primary, local)
		if err == nil || !strings.Contains(err.Error(), "invalid shell script filename") {
			t.Errorf("ExecShellScript(%q) = %v, want a refusal naming the filename", local, err)
		}
		if len(ft.uploadFiles) != 0 || len(ft.outputs) != 0 || len(ft.runs) != 0 {
			t.Errorf("a refused name must not reach the broker at all, got uploads=%v outputs=%v runs=%v",
				ft.uploadFiles, ft.outputs, ft.runs)
		}
	}
}

// TestExecShellScriptUploadFailureRunsNothing proves a failed upload stops there: the
// script never reached the broker intact, so nothing is run, and the error names the script
// and keeps the cause.
func TestExecShellScriptUploadFailureRunsNothing(t *testing.T) {
	local := filepath.Join(t.TempDir(), "job.sh")
	boom := errors.New("upload boom")
	ut := &uploadErrTransport{fakeTransport: &fakeTransport{}, err: boom}
	o := &Ops{T: ut, Cfg: &config.Config{}, Out: &bytes.Buffer{}}

	err := o.ExecShellScript(context.Background(), config.Primary, local)
	if !errors.Is(err, boom) || !strings.Contains(err.Error(), `"job.sh"`) {
		t.Errorf("err = %v, want the cause kept and the script named", err)
	}
	if len(ut.fakeTransport.outputs) != 0 {
		t.Errorf("nothing may be run after a failed upload, got %v", ut.fakeTransport.outputs)
	}
}

// TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp covers a script that
// dies partway. Failure here IS the process's own exit status -- bash reports one for the
// whole run -- so the error wraps it, the output the script produced is still shown, and
// the file is still removed: it must not be left behind in the broker just because it
// failed, since what it carries is as sensitive as what the operator put in it.
func TestExecShellScriptFailureIsTheProcessExitStatusAndStillCleansUp(t *testing.T) {
	local := filepath.Join(t.TempDir(), "job.sh")
	boom := errors.New("exit status 3")
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if len(argv) > 0 && argv[0] == "bash" {
			return []byte("half done\n"), boom
		}
		return nil, nil
	}}
	o, out := newTestOps(t, &config.Config{}, ft)

	err := o.ExecShellScript(context.Background(), config.Primary, local)
	if !errors.Is(err, boom) || !strings.Contains(err.Error(), `run shell script "job.sh"`) {
		t.Errorf("err = %v, want the exit status wrapped and the script named", err)
	}
	if !strings.Contains(out.String(), "half done") {
		t.Errorf("what the script printed before it failed is still shown, got %q", out.String())
	}
	if !ft.removed(shellScriptPath("job.sh")) {
		t.Errorf("a failed script must still be removed from the broker, runs = %v", ft.runs)
	}
}

// TestExecShellScriptTrustsTheExitStatusNotTheTranscript is the difference from ExecCLI.
//
// A CLI script's lines run independently, so a rejected one has to be read out of the
// transcript. bash reports one status for the whole run, and an operator's own script may
// print `error:` or `busy` as ordinary text while succeeding -- scanning for those would
// turn a clean run into a failure. Exit status 0 is therefore success, whatever it said.
func TestExecShellScriptTrustsTheExitStatusNotTheTranscript(t *testing.T) {
	local := filepath.Join(t.TempDir(), "job.sh")
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if len(argv) > 0 && argv[0] == "bash" {
			return []byte("error: this is just a log line\nbusy\ninvalid command is a phrase in a comment\n"), nil
		}
		return nil, nil
	}}
	o, out := newTestOps(t, &config.Config{}, ft)
	if err := o.ExecShellScript(context.Background(), config.Primary, local); err != nil {
		t.Fatalf("a script that exited 0 must not be failed for what it printed: %v", err)
	}
	if !strings.Contains(out.String(), "just a log line") {
		t.Errorf("the output is still shown in full, got %q", out.String())
	}
}

// TestExecShellScriptCleanupFailureOnlyWarns proves the removal is best-effort: the script
// has already run, so failing the command over a leftover file would report a failure that
// did not happen. The warning names the path left behind, because that is how an operator
// removes by hand a file that may carry a secret.
func TestExecShellScriptCleanupFailureOnlyWarns(t *testing.T) {
	local := filepath.Join(t.TempDir(), "job.sh")
	rt := &runErrTransport{fakeTransport: &fakeTransport{}, err: errors.New("rm failed")}
	var logs []string
	o := &Ops{
		T:   rt,
		Cfg: &config.Config{},
		Out: &bytes.Buffer{},
		Log: func(f string, a ...any) { logs = append(logs, fmt.Sprintf(f, a...)) },
	}

	if err := o.ExecShellScript(context.Background(), config.Primary, local); err != nil {
		t.Fatalf("a failed cleanup must not fail a script that ran: %v", err)
	}
	dest := shellScriptPath("job.sh")
	warned := false
	for _, l := range logs {
		if strings.Contains(l, "[WARN] cleanup of uploaded file(s)") && strings.Contains(l, dest) {
			warned = true
		}
	}
	if !warned {
		t.Errorf("want a cleanup warning naming %q, got %v", dest, logs)
	}
}

// TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList is DisableDefaultVPN's inverse: it
// runs the enable script (never the disable one), then reads the VPN list back through
// the same show both directions share, so the confirmation an operator reads comes from
// one place. The read half cleans up after itself.
func TestEnableDefaultVPNStartsTheVPNBackUpAndShowsTheList(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "show-vpn") {
			return []byte("vpn list\n"), nil
		}
		return nil, nil
	}}
	o, out := newTestOps(t, &config.Config{}, ft)
	if err := o.EnableDefaultVPN(context.Background(), config.Primary); err != nil {
		t.Fatalf("EnableDefaultVPN: %v", err)
	}
	if body := ft.wrappedCall(t, "enable-default-vpn").stdin; body != enableDefaultVPNScript() {
		t.Errorf("enable-default-vpn body = %q, want enableDefaultVPNScript", body)
	}
	if hasCall(ft, "disable-default-vpn") {
		t.Error("EnableDefaultVPN must never run the disable script")
	}
	if !strings.Contains(out.String(), "vpn list") {
		t.Errorf("the VPN list is shown as the confirmation, got %q", out.String())
	}
	if !ranContains(ft, "rm", "-f") {
		t.Error("the show-vpn read must clean up the script it uploaded")
	}
}

// TestEnableDefaultVPNFailureSkipsTheReadback proves a refused enable stops there: reading
// the VPN list back after a failed start would present a picture of a change that did not
// happen.
func TestEnableDefaultVPNFailureSkipsTheReadback(t *testing.T) {
	boom := errors.New("enable boom")
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "enable-default-vpn") {
			return nil, boom
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.EnableDefaultVPN(context.Background(), config.Primary); !errors.Is(err, boom) {
		t.Errorf("err = %v, want the enable failure", err)
	}
	if hasCall(ft, "show-vpn") {
		t.Error("EnableDefaultVPN must not read back show-vpn when the enable failed")
	}
}

// TestEnableDefaultUsersStartsEveryDefaultUserBackUp covers the enable direction of the
// list-parse-apply op: the VPN names come from the broker's own listing, and each gets the
// `no shutdown` form -- the inverse of what DisableDefaultUsers applies, so the one script
// must not carry a bare `shutdown`.
func TestEnableDefaultUsersStartsEveryDefaultUserBackUp(t *testing.T) {
	row := func(name string) string { return name + strings.Repeat(" ", 40-len(name)) + "Yes" }
	list := strings.Join([]string{strings.Repeat("-", 40), row("default"), row("myvpn")}, "\r\n") + "\r\n"
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "show-vpn") {
			return []byte(list), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.EnableDefaultUsers(context.Background(), config.Primary); err != nil {
		t.Fatalf("EnableDefaultUsers: %v", err)
	}

	body := ft.wrappedCall(t, "enable-default-usernames").stdin
	if want := enableDefaultUsersScript([]string{"default", "myvpn"}); body != want {
		t.Errorf("enable-default-usernames body = %q, want %q", body, want)
	}
	for _, vpn := range []string{"default", "myvpn"} {
		want := `client-username default message-vpn "` + vpn + "\"\nno shutdown\nexit\n"
		if !strings.Contains(body, want) {
			t.Errorf("body missing the start-up of VPN %q's default user:\n%s", vpn, body)
		}
	}
	if strings.Contains(body, "\nshutdown\n") {
		t.Errorf("the enable script must not shut anything down:\n%s", body)
	}
	if hasCall(ft, "disable-default-usernames") {
		t.Error("EnableDefaultUsers must never run the disable script")
	}
}

// TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse pins the one word that differs
// between the two directions in the "nothing to act on" warning: it must say enable, so an
// operator reading it knows which command found nothing -- and nothing may run.
func TestEnableDefaultUsersWarnsAndRunsNothingWhenNoVPNsParse(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, _ []string, _ []byte) ([]byte, error) {
		return []byte("no separator, nothing to parse\n"), nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	var logs []string
	o.Log = func(f string, a ...any) { logs = append(logs, fmt.Sprintf(f, a...)) }

	if err := o.EnableDefaultUsers(context.Background(), config.Primary); err != nil {
		t.Fatalf("EnableDefaultUsers: %v", err)
	}
	if hasCall(ft, "enable-default-usernames") {
		t.Error("nothing parsed, so no script may run")
	}
	if !strings.Contains(strings.Join(logs, "\n"), "nothing to enable.") {
		t.Errorf("the warning must name the enable direction, got %v", logs)
	}
}

// TestDomainCertsLoadsInSortedNameOrder proves the load order does not follow the order
// the env file happened to list the authorities in: uploads and the generated script are
// both sorted by CA name, so two runs of one configuration produce one script and a diff
// of two applies stays readable. The caller's slice is not reordered to get there.
func TestDomainCertsLoadsInSortedNameOrder(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	certs := []config.DomainCert{
		{Name: "zeta-ca", Path: filepath.Join("certs", "zeta.pem")},
		{Name: "alpha-ca", Path: filepath.Join("certs", "alpha.pem")},
	}
	if err := o.DomainCerts(context.Background(), config.Primary, certs); err != nil {
		t.Fatalf("DomainCerts: %v", err)
	}
	if len(ft.uploadFiles) != 2 ||
		ft.uploadFiles[0].dest != certPath("alpha-ca") ||
		ft.uploadFiles[1].dest != certPath("zeta-ca") {
		t.Errorf("uploads = %v, want alpha-ca then zeta-ca", ft.uploadFiles)
	}
	if body, want := ft.wrappedCall(t, "load-domain-certs").stdin, domainCertsScript([]string{"alpha-ca", "zeta-ca"}); body != want {
		t.Errorf("load-domain-certs body = %q, want the sorted script %q", body, want)
	}
	if certs[0].Name != "zeta-ca" {
		t.Errorf("DomainCerts reordered the caller's slice: %v", certs)
	}
}

package broker

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"solace/internal/config"
)

// --- fake transport -------------------------------------------------------

type recUpload struct {
	role config.Role
	dest string
	data string
}
type recUploadFile struct {
	role  config.Role
	local string
	dest  string
}
type recDownload struct {
	role          config.Role
	remote, local string
}
type recRun struct {
	role config.Role
	argv []string
}
type recOutput struct {
	role  config.Role
	argv  []string
	stdin string
}

// fakeTransport records every call and answers Output/OutputInput via responder.
type fakeTransport struct {
	uploads     []recUpload
	uploadFiles []recUploadFile
	downloads   []recDownload
	runs        []recRun
	outputs     []recOutput
	responder   func(role config.Role, argv []string, stdin []byte) ([]byte, error)
}

func (f *fakeTransport) Run(_ context.Context, role config.Role, argv ...string) error {
	f.runs = append(f.runs, recRun{role, argv})
	return nil
}
func (f *fakeTransport) Output(_ context.Context, role config.Role, argv ...string) ([]byte, error) {
	f.outputs = append(f.outputs, recOutput{role, argv, ""})
	if f.responder != nil {
		return f.responder(role, argv, nil)
	}
	return nil, nil
}
func (f *fakeTransport) OutputInput(_ context.Context, role config.Role, in []byte, argv ...string) ([]byte, error) {
	f.outputs = append(f.outputs, recOutput{role, argv, string(in)})
	if f.responder != nil {
		return f.responder(role, argv, in)
	}
	return nil, nil
}
func (f *fakeTransport) Upload(_ context.Context, role config.Role, data []byte, dest string) error {
	f.uploads = append(f.uploads, recUpload{role, dest, string(data)})
	return nil
}
func (f *fakeTransport) UploadFile(_ context.Context, role config.Role, local, dest string) error {
	f.uploadFiles = append(f.uploadFiles, recUploadFile{role, local, dest})
	return nil
}
func (f *fakeTransport) Download(_ context.Context, role config.Role, remote, local string) error {
	f.downloads = append(f.downloads, recDownload{role, remote, local})
	return nil
}

// matchCLI reports whether argv ran name through the CLI, in either of the two
// shapes this package now issues. The first is runCLIRead's unwrapped
// `cli -Apes .<name>.cli`, used by every `show`-style read. The second is RunCLI's
// stop-on-error wrapper: `sh -c <skeleton>`, recognised by the skeleton containing
// both `-Apes` and name's own baked-in body filename (cliRunNames) -- the skeleton
// text itself, not a literal argv slot, is what carries name for a wrapped call.
func matchCLI(argv []string, name string) bool {
	if len(argv) == 3 && argv[0] == CLIBinary && argv[1] == "-Apes" && argv[2] == cliArg(name) {
		return true
	}
	if len(argv) == 3 && argv[0] == "sh" && argv[1] == "-c" {
		bodyName, _, _ := cliRunNames(name)
		return strings.Contains(argv[2], "-Apes") && strings.Contains(argv[2], bodyName)
	}
	return false
}

// wrappedCall returns the recorded call that ran name through RunCLI's
// stop-on-error wrapper (matchCLI's "sh -c" shape), failing the test if none did.
// Its stdin is the CLI script body for a RunCLI write (ExecCLI's body instead
// travels through UploadFile, so its stdin is empty); its argv[2] is the whole
// generated skeleton, which is what a test checks for the cleanup trap.
func (f *fakeTransport) wrappedCall(t *testing.T, name string) recOutput {
	t.Helper()
	for i := len(f.outputs) - 1; i >= 0; i-- {
		o := f.outputs[i]
		if len(o.argv) == 3 && o.argv[0] == "sh" && matchCLI(o.argv, name) {
			return o
		}
	}
	t.Fatalf("no wrapped cli call for %q; outputs=%v", name, f.outputs)
	return recOutput{}
}

// outputForRole reports whether name was run against role, through either CLI
// shape matchCLI recognises.
func outputForRole(ft *fakeTransport, role config.Role, name string) bool {
	for _, o := range ft.outputs {
		if o.role == role && matchCLI(o.argv, name) {
			return true
		}
	}
	return false
}

// hasCall reports whether name was run at all, through either CLI shape matchCLI
// recognises -- the non-fatal counterpart to wrappedCall, for a test asserting
// something did NOT run.
func hasCall(ft *fakeTransport, name string) bool {
	for _, o := range ft.outputs {
		if matchCLI(o.argv, name) {
			return true
		}
	}
	return false
}

// uploadBody returns the last body uploaded to dest, or "" (and fails) if none.
func (f *fakeTransport) uploadBody(t *testing.T, dest string) string {
	t.Helper()
	for i := len(f.uploads) - 1; i >= 0; i-- {
		if f.uploads[i].dest == dest {
			return f.uploads[i].data
		}
	}
	t.Fatalf("no upload to %q; uploads=%v", dest, f.uploads)
	return ""
}

// removed reports whether an uploaded script at dest was deleted afterwards.
// removeCLI issues `rm -f <paths...>` through Run, so removals land in runs, not
// outputs. It matters for any script whose body carries a secret.
func (f *fakeTransport) removed(dest string) bool {
	for _, r := range f.runs {
		if len(r.argv) >= 3 && r.argv[0] == "rm" && r.argv[1] == "-f" {
			for _, p := range r.argv[2:] {
				if p == dest {
					return true
				}
			}
		}
	}
	return false
}

func (f *fakeTransport) hasUpload(dest string) bool {
	for _, u := range f.uploads {
		if u.dest == dest {
			return true
		}
	}
	// A write once reached here via Transport.Upload to cliScriptPath(name); it now
	// rides RunCLI's wrapped "sh -c" call instead, with no Upload at all. Recognising
	// that shape too is what keeps every existing caller of this helper -- written
	// against the old cliScriptPath(name) dest, including verify_ops.go/verify_local.go's
	// own tests -- meaningful rather than vacuously true or false either way.
	if name := cliScriptNameFromDest(dest); name != "" {
		return hasCall(f, name)
	}
	return false
}

// cliScriptNameFromDest reverses cliScriptPath: given
// "<CLIScriptsDir>/.<name>.cli" it returns name, or "" if dest is not that shape.
func cliScriptNameFromDest(dest string) string {
	const prefix, suffix = CLIScriptsDir + "/.", ".cli"
	if !strings.HasPrefix(dest, prefix) || !strings.HasSuffix(dest, suffix) {
		return ""
	}
	name := dest[len(prefix) : len(dest)-len(suffix)]
	if name == "" {
		return ""
	}
	return name
}

func newTestOps(t *testing.T, cfg *config.Config, ft *fakeTransport) (*Ops, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	o := &Ops{T: ft, Cfg: cfg, Out: buf, PollInterval: 0, PollAttempts: 3}
	// No interface addresses by default. DetectRole falls back to matching
	// redundancy.*.addr when no name matches, and reading the real interfaces here
	// would make every "this host matches nothing" case depend on how the machine
	// running the suite happens to be numbered. A test that wants that pass injects
	// its own set.
	o.LocalAddrs = func() (map[string]bool, error) { return map[string]bool{}, nil }
	return o, buf
}

// --- pure helper unit tests -----------------------------------------------

func TestField(t *testing.T) {
	out := "Configuration Status : Enabled\r\nRedundancy Status : Up\r\nActive-Standby Role : Primary\r\n"
	cases := map[string]string{
		"Configuration Status": "Enabled",
		"Redundancy Status":    "Up",
		"Active-Standby Role":  "Primary",
		"Nonexistent Label":    "",
	}
	for label, want := range cases {
		if got := field(out, label); got != want {
			t.Errorf("field(%q) = %q, want %q", label, got, want)
		}
	}
}

func TestCountContains(t *testing.T) {
	out := "Activity Status : Local Active\r\nActivity Status : Mate Active\r\nActivity Status : Local Active\r\n"
	if got := countContains(out, "Activity Status", "Local Active"); got != 2 {
		t.Errorf("countContains Local Active = %d, want 2", got)
	}
	if got := countContains(out, "Activity Status", "Mate Active"); got != 1 {
		t.Errorf("countContains Mate Active = %d, want 1", got)
	}
}

func TestValidName(t *testing.T) {
	for _, ok := range []string{"foo", "foo-bar", "foo.cli", "a_b.c-1"} {
		if err := validName("x", ok); err != nil {
			t.Errorf("validName(%q) unexpected error: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "..", "a/b", "a;b", "a b", "a$b", "../x"} {
		if err := validName("x", bad); err == nil {
			t.Errorf("validName(%q) should have failed", bad)
		}
	}
}

func TestPathHelpers(t *testing.T) {
	if got := cliScriptPath("leader"); got != "/usr/sw/jail/cliscripts/.leader.cli" {
		t.Errorf("cliScriptPath = %q", got)
	}
	if got := cliArg("leader"); got != ".leader.cli" {
		t.Errorf("cliArg = %q", got)
	}
	if got := certPath("tls.crt.key"); got != "/usr/sw/jail/certs/tls.crt.key" {
		t.Errorf("certPath = %q", got)
	}
}

func TestLastLines(t *testing.T) {
	got := lastLines("a\nb\nc\nd\ne\n", 3)
	if got != "c\nd\ne\n" {
		t.Errorf("lastLines = %q", got)
	}
	if got := lastLines("only\n", 3); got != "only\n" {
		t.Errorf("lastLines fewer-than-n = %q", got)
	}
}

func TestHTTPStatusHelpers(t *testing.T) {
	if !isHTTP2xx("HTTP/1.1 200 OK") || !isHTTP2xx("HTTP/2 204") {
		t.Error("isHTTP2xx should accept 2xx")
	}
	for _, bad := range []string{"HTTP/1.1 401 Unauthorized", "HTTP/1.1 500", "", "garbage"} {
		if isHTTP2xx(bad) {
			t.Errorf("isHTTP2xx(%q) should be false", bad)
		}
	}
	lines := httpStatusLines("HTTP/1.1 100 Continue\r\nHTTP/1.1 200 OK\r\n\r\nbody\r\n")
	if len(lines) != 2 || lines[1] != "HTTP/1.1 200 OK" {
		t.Errorf("httpStatusLines = %v", lines)
	}
}

func TestPrimaryRedundancyUp(t *testing.T) {
	up := "Configuration Status : Enabled\nRedundancy Status : Up\nActive-Standby Role : Primary\nADB Link To Mate : Up\nADB Hello To Mate : Up\n"
	if !primaryRedundancyUp(up) {
		t.Error("primaryRedundancyUp should be true for a healthy Primary")
	}
	down := strings.Replace(up, "Redundancy Status : Up", "Redundancy Status : Down", 1)
	if primaryRedundancyUp(down) {
		t.Error("primaryRedundancyUp should be false when redundancy is Down")
	}
}

// --- RunCLI primitive ------------------------------------------------------

// TestRunCLIWrapsWithStopOnError pins the shape RunCLI now issues: ONE
// Transport.OutputInput call, body on stdin (never Upload, so it never touches
// argv or a log), a generated `sh -c` skeleton that sources it through the
// broker's own `stop-on-error`, and a cleanup trap -- so there is no separate
// removeCLI call.
func TestRunCLIWrapsWithStopOnError(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "probe") {
			return []byte("output\n"), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	out, err := o.RunCLI(context.Background(), config.Primary, "probe", "show version\n")
	if err != nil {
		t.Fatalf("RunCLI error: %v", err)
	}
	if string(out) != "output\n" {
		t.Errorf("RunCLI output = %q", out)
	}
	call := ft.wrappedCall(t, "probe")
	if call.stdin != "show version\n" {
		t.Errorf("RunCLI body on stdin = %q, want the script body", call.stdin)
	}
	if !strings.Contains(call.argv[2], "-Apes") || !strings.Contains(call.argv[2], "stop-on-error") {
		t.Errorf("RunCLI skeleton = %q, want a stop-on-error `-Apes` invocation", call.argv[2])
	}
	if !strings.Contains(call.argv[2], "trap cleanup") {
		t.Error("RunCLI skeleton must clean up its own broker-side files on exit")
	}
	if len(ft.uploads) != 0 {
		t.Error("RunCLI must not use Transport.Upload -- the body rides OutputInput's stdin")
	}
	if len(ft.runs) != 0 {
		t.Errorf("RunCLI must not issue a separate cleanup call, the skeleton's own trap does it: %v", ft.runs)
	}
}

func TestRunCLIRejectsBadName(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if _, err := o.RunCLI(context.Background(), config.Primary, "../evil", "body"); err == nil {
		t.Error("RunCLI should reject an invalid name")
	}
	if len(ft.outputs) != 0 {
		t.Error("RunCLI must not exec when the name is invalid")
	}
}

// TestRunCLIDetectsRejection pins the reason the wrapper exists: `cli -Apes` exits
// 0 even when the broker refuses a line, so RunCLI's own scan of the transcript
// tail (rejectionIn, the same failKeywords list driver.go's chunks use) is the
// only thing that turns that into a Go error. The error must never quote the
// rejected line itself -- a CLI transcript can carry passwords.
func TestRunCLIDetectsRejection(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "probe") {
			return []byte("line one OK\nInvalid command at line 2\n"), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	_, err := o.RunCLI(context.Background(), config.Primary, "probe", "body\n")
	if err == nil {
		t.Fatal("RunCLI should fail when the broker rejected a line")
	}
	if strings.Contains(err.Error(), "Invalid command at line 2") {
		t.Errorf("RunCLI error must not quote the rejected line, got %v", err)
	}
	// The sentinel, not the sentence. A caller that has to tell "the broker refused a
	// line" from "the broker could not be reached" -- ConfigureReplication's two phases
	// are the one that does -- reads this with errors.Is. It matched a phrase in the
	// message first, so rewording the message would have silently reclassified every
	// rejection as unreachable, downgrading a report that names what landed into one
	// that says it cannot be known. Nothing pinned that coupling until this line.
	if !errors.Is(err, ErrCLIRejected) {
		t.Errorf("RunCLI rejection must wrap ErrCLIRejected so callers can classify it, got %v", err)
	}
}

// TestRunCLITransportErrorIsNotARejection is the other half of that classification: a
// call that never reached the broker must NOT satisfy errors.Is(err, ErrCLIRejected).
// The two leave the broker in different states -- a rejection stopped it at a known
// line, an unreachable broker may have applied all, some or none of the script, since
// the sourced script keeps running inside it -- so a caller that conflated them would
// tell the operator the wrong thing about what is live right now.
func TestRunCLITransportErrorIsNotARejection(t *testing.T) {
	ft := &fakeTransport{responder: func(config.Role, []string, []byte) ([]byte, error) {
		return nil, errors.New("connection lost")
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	_, err := o.RunCLI(context.Background(), config.Primary, "probe", "body\n")
	if err == nil {
		t.Fatal("RunCLI should fail when the transport does")
	}
	if errors.Is(err, ErrCLIRejected) {
		t.Errorf("a transport failure must not classify as a broker rejection, got %v", err)
	}
}

// TestRunCLICleanRunNoError is TestRunCLIDetectsRejection's mirror: output with
// none of failKeywords in its tail must not fail the call.
func TestRunCLICleanRunNoError(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "probe") {
			return []byte("line one OK\nline two OK\n"), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if _, err := o.RunCLI(context.Background(), config.Primary, "probe", "body\n"); err != nil {
		t.Errorf("RunCLI should not fail on clean output: %v", err)
	}
}

func TestSkipIfStandalone(t *testing.T) {
	ha, _ := newTestOps(t, &config.Config{Redundancy: config.Redundancy{Enabled: "true"}}, &fakeTransport{})
	if ha.skipIfStandalone("x") {
		t.Error("skipIfStandalone should be false in HA")
	}
	sa, _ := newTestOps(t, &config.Config{Redundancy: config.Redundancy{Enabled: "false"}}, &fakeTransport{})
	if !sa.skipIfStandalone("x") {
		t.Error("skipIfStandalone should be true for standalone")
	}
}

// --- config ops ------------------------------------------------------------

func TestServerCert(t *testing.T) {
	dir := t.TempDir()
	key := filepath.Join(dir, "tls.key")
	crt := filepath.Join(dir, "tls.crt")
	ca := filepath.Join(dir, "ca.pem")
	writeFile(t, key, "KEY\n")
	writeFile(t, crt, "CERT\n")
	writeFile(t, ca, "CA\n")

	cfg := &config.Config{TLS: config.TLS{Cert: crt, CertKey: key, CAs: []string{ca}}}
	ft := &fakeTransport{}
	o, _ := newTestOps(t, cfg, ft)
	if err := o.ServerCert(context.Background(), "2026-07-31", config.Primary); err != nil {
		t.Fatalf("ServerCert error: %v", err)
	}
	dest := certPath(serverCertFile("2026-07-31"))
	if body := ft.uploadBody(t, dest); body != "KEY\nCERT\nCA\n" {
		t.Errorf("cert bundle = %q, want key+cert+ca", body)
	}
	// The CLI script itself now rides RunCLI's wrapped call, so its body is on the
	// OutputInput stdin rather than a Transport.Upload -- ft.uploadBody would find
	// nothing.
	if body := ft.wrappedCall(t, "apply-server-certs").stdin; body != serverCertScript("2026-07-31") {
		t.Errorf("apply-server-certs body = %q", body)
	}
}

func TestServerCertRequiresCert(t *testing.T) {
	o, _ := newTestOps(t, &config.Config{}, &fakeTransport{})
	if err := o.ServerCert(context.Background(), "2026-07-31", config.Primary); err == nil {
		t.Error("ServerCert should error when tls.cert/certKey are unset")
	}
}

func TestDomainCerts(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	certs := []config.DomainCert{{Name: "myca", Path: filepath.Join("certs", "myca.pem")}}
	if err := o.DomainCerts(context.Background(), config.Primary, certs); err != nil {
		t.Fatalf("DomainCerts error: %v", err)
	}
	if len(ft.uploadFiles) != 1 ||
		ft.uploadFiles[0].local != filepath.Join("certs", "myca.pem") ||
		ft.uploadFiles[0].dest != certPath("myca") {
		t.Errorf("DomainCerts uploadFiles = %v", ft.uploadFiles)
	}
	if body := ft.wrappedCall(t, "load-domain-certs").stdin; body != domainCertsScript([]string{"myca"}) {
		t.Errorf("load-domain-certs body = %q", body)
	}
}

func TestDomainCertsRejectsBadName(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	certs := []config.DomainCert{{Name: "bad name", Path: "x.pem"}}
	if err := o.DomainCerts(context.Background(), config.Primary, certs); err == nil {
		t.Error("DomainCerts should reject a CA name with a space")
	}
	if len(ft.uploadFiles) != 0 {
		t.Error("DomainCerts must not upload before validating the name")
	}
}

// TestDomainCertsAcceptsAFullHostPath pins the inverted assertion from before the
// domainCerts shape change: the local source is now a full host path (already
// resolved by config.ResolveDomainCerts), not a bare filename joined onto a
// folder -- and it must upload rather than error, since it is no longer either a
// CLI operand or an in-broker path, only a local argument to UploadFile.
func TestDomainCertsAcceptsAFullHostPath(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	certs := []config.DomainCert{{Name: "myca", Path: filepath.FromSlash("/opt/solace/prod-cas/ca.pem")}}
	if err := o.DomainCerts(context.Background(), config.Primary, certs); err != nil {
		t.Fatalf("DomainCerts with a full host path must be accepted: %v", err)
	}
	if len(ft.uploadFiles) != 1 || ft.uploadFiles[0].local != filepath.FromSlash("/opt/solace/prod-cas/ca.pem") {
		t.Errorf("DomainCerts uploadFiles = %v", ft.uploadFiles)
	}
}

func TestDomainCertsEmptySkips(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.DomainCerts(context.Background(), config.Primary, nil); err != nil {
		t.Fatalf("DomainCerts empty error: %v", err)
	}
	if len(ft.uploadFiles) != 0 || len(ft.outputs) != 0 {
		t.Error("DomainCerts with no files should make no calls")
	}
}

func TestDisableDefaultVPN(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "show-vpn") {
			return []byte("vpn list\n"), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.DisableDefaultVPN(context.Background(), config.Primary); err != nil {
		t.Fatalf("DisableDefaultVPN error: %v", err)
	}
	if body := ft.wrappedCall(t, "disable-default-vpn").stdin; body != disableDefaultVPNScript() {
		t.Errorf("disable-default-vpn body mismatch")
	}
	// The write half cleans itself up via its skeleton's own trap; the read half
	// (show-vpn) still goes through removeCLI's `rm -f`.
	if !ranContains(ft, "rm", "-f") {
		t.Error("DisableDefaultVPN should clean up its show-vpn read")
	}
}

func TestDisableDefaultUsers(t *testing.T) {
	row := func(name string) string { return name + strings.Repeat(" ", 40-len(name)) + "Yes" }
	list := strings.Join([]string{
		strings.Repeat("-", 40),
		row("default"),
		row("myvpn"),
	}, "\r\n") + "\r\n"
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "show-vpn") {
			return []byte(list), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.DisableDefaultUsers(context.Background(), config.Primary); err != nil {
		t.Fatalf("DisableDefaultUsers error: %v", err)
	}
	body := ft.wrappedCall(t, "disable-default-usernames").stdin
	for _, want := range []string{
		`client-username default message-vpn "default"`,
		`client-username default message-vpn "myvpn"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("disable-default-usernames body missing %q", want)
		}
	}
}

func TestDisableDefaultUsersNoVPNs(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, _ []string, _ []byte) ([]byte, error) {
		return []byte("no separator, nothing to parse\n"), nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.DisableDefaultUsers(context.Background(), config.Primary); err != nil {
		t.Fatalf("DisableDefaultUsers error: %v", err)
	}
	if hasCall(ft, "disable-default-usernames") {
		t.Error("DisableDefaultUsers should not run when no VPNs are parsed")
	}
}

func TestProductKeys(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "product-keys") {
			return []byte("Product key applied.\n"), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.ProductKeys(context.Background(), []string{"KEY-1"}, config.Primary); err != nil {
		t.Fatalf("ProductKeys error: %v", err)
	}
	if body := ft.wrappedCall(t, "product-keys").stdin; body != productKeysScript([]string{"KEY-1"}) {
		t.Errorf("product-keys body = %q", body)
	}
}

// TestProductKeysDetectsError used to trigger on a bare "fail" via containsAnyFold;
// that scanner is gone, converged onto RunCLI's own failKeywords tail scan (the
// vetted list, not a bare "error"/"fail" that false-positives on e.g.
// "error-events"), so the canned transcript now carries one of those phrases.
func TestProductKeysDetectsError(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, _ []string, _ []byte) ([]byte, error) {
		return []byte("Error: invalid key\n"), nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.ProductKeys(context.Background(), []string{"BAD"}, config.Primary); err == nil {
		t.Error("ProductKeys should fail loud when the broker reports an error")
	}
}

func TestProductKeysEmpty(t *testing.T) {
	o, _ := newTestOps(t, &config.Config{}, &fakeTransport{})
	if err := o.ProductKeys(context.Background(), nil, config.Primary); err == nil {
		t.Error("ProductKeys should error with no keys")
	}
}

// TestProductKeysRejectsMultilineKey closes the injection path: each key is written
// into one line of a script that runs with admin already enabled, so a newline
// would append extra commands to that elevated session. The charset itself is not
// constrained -- a product key is an opaque vendor string -- so this pins the
// control-character rejection and that nothing is uploaded before it.
func TestProductKeysRejectsMultilineKey(t *testing.T) {
	for _, bad := range []string{
		"KEY-1\nshow running-config",
		"KEY-1\rmore",
		"   ",
	} {
		ft := &fakeTransport{}
		o, _ := newTestOps(t, &config.Config{}, ft)
		if err := o.ProductKeys(context.Background(), []string{bad}, config.Primary); err == nil {
			t.Errorf("ProductKeys should reject %q", bad)
		}
		if len(ft.uploads) != 0 || len(ft.outputs) != 0 {
			t.Errorf("ProductKeys must validate before running anything, got uploads=%v runs=%v", ft.uploads, ft.outputs)
		}
	}
	// A key with vendor punctuation outside the identifier charset must still pass:
	// the alphabet of a product key is not ours to decide.
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.ProductKeys(context.Background(), []string{"AbC+dE/f12=="}, config.Primary); err != nil {
		t.Errorf("an opaque vendor key must be accepted: %v", err)
	}
}

// --- additional CLI users (k8s only) ----------------------------------------

// TestRemoveDomainCerts covers the removal half of the domain-CA pair, which had
// no coverage at all before the container tree gained its own teardown path.
func TestRemoveDomainCerts(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.RemoveDomainCerts(context.Background(), config.Primary, []string{"myca"}); err != nil {
		t.Fatalf("RemoveDomainCerts error: %v", err)
	}
	if body := ft.wrappedCall(t, "remove-domain-certs").stdin; !strings.Contains(body, "myca") {
		t.Errorf("remove-domain-certs body should name the CA: %q", body)
	}
}

func TestRemoveDomainCertsRejectsBadName(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.RemoveDomainCerts(context.Background(), config.Primary, []string{"bad name"}); err == nil {
		t.Error("RemoveDomainCerts should reject a CA name with a space")
	}
	if len(ft.uploads) != 0 {
		t.Error("RemoveDomainCerts must not upload before validating the name")
	}
}

func TestRemoveDomainCertsEmptySkips(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.RemoveDomainCerts(context.Background(), config.Primary, nil); err != nil {
		t.Fatalf("RemoveDomainCerts empty error: %v", err)
	}
	if len(ft.uploads) != 0 || len(ft.outputs) != 0 {
		t.Error("RemoveDomainCerts with no CAs should make no calls")
	}
}

// TestRemoveServerCerts pins the removal against the CLI form confirmed on a live
// broker. The script is compared WHOLE rather than by substring: a `no ssl
// server-certificate` reached without `enable`/`configure` first is rejected by the
// broker, and a substring match would not notice.
func TestRemoveServerCerts(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.RemoveServerCerts(context.Background(), config.Primary); err != nil {
		t.Fatalf("RemoveServerCerts error: %v", err)
	}
	const want = "home\nno paging\nenable\nconfigure\nno ssl server-certificate\n"
	if body := ft.wrappedCall(t, "remove-server-certs").stdin; body != want {
		t.Errorf("remove-server-certs body = %q, want %q", body, want)
	}
}

// TestRemoveServerCertsSpansEveryRole: the apply path loads the certificate onto every
// node of the group, so the removal has to reach every node too. A certificate gone from
// the primary and still loaded on the backup is a half state that survives a failover.
func TestRemoveServerCertsSpansEveryRole(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	roles := []config.Role{config.Primary, config.Backup, config.Monitor}
	if err := o.RemoveServerCerts(context.Background(), roles...); err != nil {
		t.Fatalf("RemoveServerCerts error: %v", err)
	}
	for _, r := range roles {
		if !outputForRole(ft, r, "remove-server-certs") {
			t.Errorf("no remove-server-certs run for the %q node", r)
		}
	}
}

// TestRemoveServerCertsRunCLIError: a failure on one node aborts rather than carrying on
// through the group. Half a removal reported as success is the outcome worth preventing.
func TestRemoveServerCertsRunCLIError(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "remove-server-certs") {
			return nil, errors.New("boom")
		}
		return nil, nil
	}}
	o, buf := newTestOps(t, &config.Config{}, ft)
	err := o.RemoveServerCerts(context.Background(), config.Primary, config.Backup)
	if err == nil {
		t.Fatal("RemoveServerCerts should return the CLI failure")
	}
	if buf.String() != "" {
		t.Errorf("nothing should be shown when the removal fails, got %q", buf)
	}
}

// TestRemoveProductKeys covers the revocation half, which was a loud placeholder until
// `no product-key <key>` was confirmed on a live broker. It is ProductKeys' mirror image
// by design, so the same three properties are asserted of it.
func TestRemoveProductKeys(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "remove-product-keys") {
			return []byte("Product key removed.\n"), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.RemoveProductKeys(context.Background(), []string{"KEY-1"}, config.Primary); err != nil {
		t.Fatalf("RemoveProductKeys error: %v", err)
	}
	if body := ft.wrappedCall(t, "remove-product-keys").stdin; body != removeProductKeysScript([]string{"KEY-1"}) {
		t.Errorf("remove-product-keys body = %q", body)
	}
}

// TestRemoveProductKeysRefusesAnEmptyList: with nothing configured there is nothing to
// revoke, and reporting success for having done nothing is what the apply path already
// refuses to do.
func TestRemoveProductKeysRefusesAnEmptyList(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.RemoveProductKeys(context.Background(), nil, config.Primary); err == nil {
		t.Error("an empty key list must be refused, not reported as a successful removal")
	}
	if len(ft.uploads) != 0 {
		t.Error("nothing should be uploaded when there is nothing to revoke")
	}
}

// TestRemoveProductKeysValidatesBeforeUploading: each key is interpolated into a CLI line
// that runs with admin already enabled, so it is checked before anything reaches the
// broker -- the same order ProductKeys and DomainCerts use.
func TestRemoveProductKeysValidatesBeforeUploading(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.RemoveProductKeys(context.Background(), []string{"KEY-1\nshutdown"}, config.Primary); err == nil {
		t.Error("a product key carrying a second CLI line must be rejected")
	}
	if len(ft.uploads) != 0 {
		t.Error("RemoveProductKeys must validate before it uploads")
	}
}

// TestRemoveProductKeysScansTheOutput is the property that matters most here. Revoking a
// key the broker does not hold is the kind of thing a CLI reports in prose and returns
// zero for -- so without this scan the command would report success and the operator
// would believe an entitlement was gone when it is not. The scan itself is now RunCLI's
// own (failKeywords via rejectionIn), not a local containsAnyFold("error","fail").
func TestRemoveProductKeysScansTheOutput(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "remove-product-keys") {
			return []byte("Error: product key not found.\n"), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	err := o.RemoveProductKeys(context.Background(), []string{"KEY-1"}, config.Primary)
	if err == nil || !strings.Contains(err.Error(), "rejected") {
		t.Errorf("err = %v, want the output scan to fail the command", err)
	}
}

func TestExecCLI(t *testing.T) {
	local := filepath.Join(t.TempDir(), "myscript.cli")
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "myscript.cli") {
			return []byte("OK\n"), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.ExecCLI(context.Background(), config.Primary, local); err != nil {
		t.Fatalf("ExecCLI error: %v", err)
	}
	// ExecCLI now uploads to the BARE path `source script` runs, not the old
	// dot-and-.cli form -- cliRunNames is the same convention RunCLI's own writes use.
	bodyName, _, _ := cliRunNames("myscript.cli")
	dest := CLIScriptsDir + "/" + bodyName
	if len(ft.uploadFiles) != 1 || ft.uploadFiles[0].dest != dest || ft.uploadFiles[0].local != local {
		t.Errorf("ExecCLI uploadFiles = %v, want dest %q", ft.uploadFiles, dest)
	}
	call := ft.wrappedCall(t, "myscript.cli")
	if !strings.Contains(call.argv[2], "trap cleanup") {
		t.Error("ExecCLI should clean up the uploaded script via its skeleton's trap")
	}
}

func TestExecCLIRejectsBadName(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.ExecCLI(context.Background(), config.Primary, "/some/dir/.."); err == nil {
		t.Error("ExecCLI should reject a base name of '..'")
	}
	if len(ft.uploadFiles) != 0 {
		t.Error("ExecCLI must not upload an invalid-named script")
	}
}

// TestExecCLIReportsRejectedLines closes the rejected-line branch: ExecCLI now
// runs through RunCLI's stop-on-error wrapper (see ExecCLI's own doc comment), so
// a rejected line is caught by the transcript tail scan rather than the old
// whole-transcript countAnyFold. The transcript is still shown in full either way,
// and the error must not quote the rejected line itself, since a CLI transcript
// can carry passwords (the same reason the removed additional-users op withheld
// its own).
func TestExecCLIReportsRejectedLines(t *testing.T) {
	local := filepath.Join(t.TempDir(), "myscript.cli")
	const out = "line one OK\nInvalid command at line 2\nline three OK\n"
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "myscript.cli") {
			return []byte(out), nil
		}
		return nil, nil
	}}
	o, buf := newTestOps(t, &config.Config{}, ft)
	err := o.ExecCLI(context.Background(), config.Primary, local)
	if err == nil {
		t.Fatal("ExecCLI should fail when the broker rejected a line")
	}
	if strings.Contains(err.Error(), "Invalid command at line 2") {
		t.Errorf("ExecCLI error must not quote the rejected line, got %v", err)
	}
	if !strings.Contains(ft.wrappedCall(t, "myscript.cli").argv[2], "trap cleanup") {
		t.Error("ExecCLI must still clean up the uploaded script when the run was partly rejected")
	}
	if !strings.Contains(buf.String(), out) {
		t.Error("ExecCLI must still show the full output when the run was partly rejected")
	}
}

// --- verify ops ------------------------------------------------------------

func TestLogin(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, _ []string, _ []byte) ([]byte, error) {
		return []byte("HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{}\r\n"), nil
	}}
	o, buf := newTestOps(t, &config.Config{}, ft)
	ok, err := o.Login(context.Background(), config.Primary, "admin", "s3cret")
	if err != nil || !ok {
		t.Fatalf("Login ok=%v err=%v", ok, err)
	}
	if !strings.Contains(buf.String(), "[ OK ] Login") {
		t.Errorf("Login output = %q, want the house [ OK ] outcome line", buf.String())
	}
	// The password must ride stdin, never the argv.
	if ft.outputs[0].stdin != "user = \"admin:s3cret\"\n" {
		t.Errorf("Login stdin = %q", ft.outputs[0].stdin)
	}
	if strings.Contains(strings.Join(ft.outputs[0].argv, " "), "s3cret") {
		t.Error("password leaked into curl argv")
	}
}

func TestLoginFailure(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, _ []string, _ []byte) ([]byte, error) {
		return []byte("HTTP/1.1 401 Unauthorized\r\n"), nil
	}}
	o, buf := newTestOps(t, &config.Config{}, ft)
	ok, err := o.Login(context.Background(), config.Primary, "admin", "bad")
	if err != nil || ok {
		t.Fatalf("Login ok=%v err=%v, want false/nil", ok, err)
	}
	if !strings.Contains(buf.String(), "401") {
		t.Errorf("Login failure output = %q", buf.String())
	}
}

func TestLoginNoResponse(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, _ []string, _ []byte) ([]byte, error) {
		return []byte(""), nil
	}}
	o, buf := newTestOps(t, &config.Config{}, ft)
	ok, _ := o.Login(context.Background(), config.Primary, "admin", "x")
	if ok || !strings.Contains(buf.String(), "no HTTP response") {
		t.Errorf("Login no-response ok=%v out=%q", ok, buf.String())
	}
}

func TestLeaderStandaloneSkips(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{Redundancy: config.Redundancy{Enabled: "false"}}, ft)
	if err := o.Leader(context.Background()); err != nil {
		t.Fatalf("Leader standalone error: %v", err)
	}
	if len(ft.outputs) != 0 || len(ft.uploads) != 0 || len(ft.runs) != 0 {
		t.Error("Leader must make no calls in standalone mode")
	}
}

func TestLeaderSuccess(t *testing.T) {
	healthy := "Configuration Status : Enabled\nRedundancy Status : Up\nActive-Standby Role : Primary\nADB Link To Mate : Up\nADB Hello To Mate : Up\n"
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		switch {
		case matchCLI(argv, "show-rd"):
			return []byte(healthy), nil
		case matchCLI(argv, "assert-leader"):
			return []byte("l1\nl2\nSync Complete\n"), nil
		}
		return nil, nil
	}}
	o, buf := newTestOps(t, &config.Config{Redundancy: config.Redundancy{Enabled: "true"}}, ft)
	if err := o.Leader(context.Background()); err != nil {
		t.Fatalf("Leader error: %v", err)
	}
	// revert-activity is run against the Backup, assert-leader against the Primary.
	// Both now go through RunCLI's wrapped call rather than a plain Upload, so
	// outputForRole (which recognises either CLI shape) is what checks them.
	if !outputForRole(ft, config.Backup, "revert-activity") {
		t.Error("Leader should revert activity on the Backup")
	}
	if !outputForRole(ft, config.Primary, "assert-leader") {
		t.Error("Leader should assert leadership on the Primary")
	}
	if !strings.Contains(buf.String(), "Sync Complete") {
		t.Errorf("Leader output = %q", buf.String())
	}
}

func TestLeaderTimeout(t *testing.T) {
	down := "Configuration Status : Enabled\nRedundancy Status : Down\nActive-Standby Role : Primary\nADB Link To Mate : Up\nADB Hello To Mate : Up\n"
	sawDetail := false
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "show-redundancy-detail") {
			sawDetail = true
			return []byte("detail dump\n"), nil
		}
		return []byte(down), nil
	}}
	o, _ := newTestOps(t, &config.Config{Redundancy: config.Redundancy{Enabled: "true"}}, ft)
	o.PollAttempts = 2
	if err := o.Leader(context.Background()); err == nil {
		t.Error("Leader should time out when redundancy never recovers")
	}
	if !sawDetail {
		t.Error("Leader should dump show-redundancy-detail on timeout")
	}
}

func TestRedundancySuccess(t *testing.T) {
	adb := "ADB Link To Mate : Up\nADB Hello To Mate : Up\n"
	base := func(role, cfg, rdc, act string) string {
		return "Configuration Status : " + cfg + "\nRedundancy Status : " + rdc +
			"\nActive-Standby Role : " + role + "\n" + adb + "Activity Status : " + act + "\n"
	}
	priSeq := []string{
		base("Primary", "Enabled", "Up", "Local Active"),           // a: initial, primary active
		base("Primary", "Enabled-Released", "Down", "Mate Active"), // c: released
		base("Primary", "Enabled", "Up", "Mate Active"),            // e: un-released (mate active)
		base("Primary", "Enabled", "Up", "Local Active"),           // h: reverted (local active)
	}
	bkSeq := []string{
		base("Backup", "Enabled", "Up", "Local Active"), // e: backup active
		base("Backup", "Enabled", "Up", "Local Active"), // f: revert precheck
		base("Backup", "Enabled", "Up", "Mate Active"),  // h: reverted (not active)
	}
	var pri, bk int
	ft := &fakeTransport{responder: func(role config.Role, argv []string, _ []byte) ([]byte, error) {
		if !matchCLI(argv, "show-rd") {
			return nil, nil // release / no-release / revert-activity execs
		}
		if role == config.Primary {
			out := priSeq[pri]
			pri++
			return []byte(out), nil
		}
		out := bkSeq[bk]
		bk++
		return []byte(out), nil
	}}
	o, _ := newTestOps(t, &config.Config{Redundancy: config.Redundancy{Enabled: "true"}}, ft)
	if err := o.Redundancy(context.Background()); err != nil {
		t.Fatalf("Redundancy error: %v", err)
	}
	if pri != len(priSeq) || bk != len(bkSeq) {
		t.Errorf("Redundancy consumed pri=%d/%d bk=%d/%d", pri, len(priSeq), bk, len(bkSeq))
	}
}

func TestRedundancyStandaloneSkips(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, &config.Config{Redundancy: config.Redundancy{Enabled: "false"}}, ft)
	if err := o.Redundancy(context.Background()); err != nil {
		t.Fatalf("Redundancy standalone error: %v", err)
	}
	if len(ft.outputs) != 0 {
		t.Error("Redundancy must make no calls in standalone mode")
	}
}

func TestDiagnostics(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "diag")
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		switch {
		case matchCLI(argv, "gather-configs"):
			return []byte("Diagnostics saved: logs/diag-node1.tgz\n"), nil
		case len(argv) == 1 && argv[0] == "hostname":
			return []byte("node1\n"), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	if err := o.Diagnostics(context.Background(), dest, "20260731", 3, config.Primary); err != nil {
		t.Fatalf("Diagnostics error: %v", err)
	}
	if _, err := os.Stat(dest); err != nil {
		t.Errorf("Diagnostics did not create dest dir: %v", err)
	}
	if body := ft.uploadBody(t, cliScriptPath("gather-configs")); body != gatherConfigsScript(3) {
		t.Error("gather-configs body mismatch")
	}
	wantDownloads := map[string]string{
		JailRoot + "/gather-configs.zip":  filepath.Join(dest, "gather-configs-node1-20260731.zip"),
		JailRoot + "/logs/diag-node1.tgz": filepath.Join(dest, "diag-node1.tgz"),
	}
	if len(ft.downloads) != len(wantDownloads) {
		t.Fatalf("Diagnostics downloads = %v", ft.downloads)
	}
	for _, d := range ft.downloads {
		if want, ok := wantDownloads[d.remote]; !ok || want != d.local {
			t.Errorf("unexpected download %+v", d)
		}
	}
}

// --- test helpers ----------------------------------------------------------

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func ranContains(ft *fakeTransport, argv ...string) bool {
	for _, r := range ft.runs {
		if len(r.argv) >= len(argv) {
			match := true
			for i := range argv {
				if r.argv[i] != argv[i] {
					match = false
					break
				}
			}
			if match {
				return true
			}
		}
	}
	return false
}

func uploadedForRole(ft *fakeTransport, role config.Role, dest string) bool {
	for _, u := range ft.uploads {
		if u.role == role && u.dest == dest {
			return true
		}
	}
	// See hasUpload: a write against dest may now be a RunCLI wrapped call instead
	// of a plain Upload, so recognise that shape too rather than leave every
	// existing caller of this helper silently meaningless.
	if name := cliScriptNameFromDest(dest); name != "" {
		return outputForRole(ft, role, name)
	}
	return false
}

// TestServerCertBundleOrder pins what the bundle CONTAINS and in what order, with no
// transport involved -- so the container platforms and the CLI path can be compared
// against one definition rather than against each other's behaviour.
//
// Key before certificate is the order the CLI path has always written (hence
// serverCertFile's .crt.key extension and the bash ancestor's `cat CERTKEY CERT`).
// The CAs are deliberately absent: trusted CAs are installed into the broker's own
// trust store by `config apply domain-certs`, and are not part of the certificate
// the broker presents.
func TestServerCertBundleOrder(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	cfg := &config.Config{}
	cfg.TLS.CertKey = write("tls.key", "KEYBYTES\n")
	cfg.TLS.Cert = write("tls.crt", "CERTBYTES\n")
	cfg.TLS.CAs = []string{write("ca.pem", "CABYTES\n")}

	got, err := ServerCertBundle(cfg)
	if err != nil {
		t.Fatalf("ServerCertBundle: %v", err)
	}
	if want := "KEYBYTES\nCERTBYTES\n"; string(got) != want {
		t.Errorf("bundle = %q, want %q (key then certificate)", got, want)
	}
	if strings.Contains(string(got), "CABYTES") {
		t.Error("the bundle must not carry tls.cas: CAs go into the broker's trust store via " +
			"`config apply domain-certs`, not into the certificate it presents")
	}
}

// TestServerCertBundleRequiresBothHalves covers the guard, including the asymmetric
// cases: a certificate with no key cannot produce a usable bundle, and neither can a
// key with no certificate.
func TestServerCertBundleRequiresBothHalves(t *testing.T) {
	for _, tc := range []struct{ name, cert, key string }{
		{"neither", "", ""},
		{"cert without key", "certs/tls.crt", ""},
		{"key without cert", "", "certs/tls.key"},
	} {
		cfg := &config.Config{}
		cfg.TLS.Cert, cfg.TLS.CertKey = tc.cert, tc.key
		if _, err := ServerCertBundle(cfg); err == nil {
			t.Errorf("%s: ServerCertBundle must refuse, since the broker needs both halves in one file", tc.name)
		}
	}
}

// TestServerCertBundleReportsAnUnreadableFile makes sure the error names the path.
// Once the container platforms build this bundle at deploy time, a mistyped path is
// the most likely failure, and it has to point at the file rather than at TLS in
// general.
func TestServerCertBundleReportsAnUnreadableFile(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "present.pem")
	if err := os.WriteFile(real, []byte("BYTES\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// One case per half, so whichever file is missing is the one named. A single
	// case with both missing would only ever prove whichever happens to be read
	// first, and would silently stop covering the other half if that order changed.
	for _, tc := range []struct{ name, cert, key, want string }{
		{"certificate missing", filepath.Join(dir, "missing.crt"), real, "missing.crt"},
		{"key missing", real, filepath.Join(dir, "missing.key"), "missing.key"},
	} {
		cfg := &config.Config{}
		cfg.TLS.Cert, cfg.TLS.CertKey = tc.cert, tc.key
		_, err := ServerCertBundle(cfg)
		if err == nil {
			t.Errorf("%s: ServerCertBundle must fail when a file cannot be read", tc.name)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %q must name the unreadable path %q", tc.name, err, tc.want)
		}
	}
}

// TestServerCertBundleRefusesAKeyBearingCert covers the misconfiguration this
// concatenation would otherwise turn into a silently wrong file: a tls.cert that
// already carries its private key, alongside a tls.certKey, yields a bundle with the
// key twice. A deployment predating this tool is exactly where a pre-chained file
// turns up, so the message has to say which of the two fields to change.
func TestServerCertBundleRefusesAKeyBearingCert(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	key := write("tls.key", "-----BEGIN PRIVATE KEY-----\nk\n-----END PRIVATE KEY-----\n")
	// Every spelling openssl and its relatives emit, since the check exists to catch
	// a real operator's file rather than a canonical one.
	for _, header := range []string{
		"-----BEGIN PRIVATE KEY-----",
		"-----BEGIN RSA PRIVATE KEY-----",
		"-----BEGIN EC PRIVATE KEY-----",
		"-----BEGIN ENCRYPTED PRIVATE KEY-----",
	} {
		cfg := &config.Config{}
		cfg.TLS.CertKey = key
		cfg.TLS.Cert = write("chained.pem",
			header+"\nk\n-----END PRIVATE KEY-----\n-----BEGIN CERTIFICATE-----\nc\n-----END CERTIFICATE-----\n")
		_, err := ServerCertBundle(cfg)
		if err == nil {
			t.Errorf("%s: a cert file carrying a private key must be refused, or the bundle holds the key twice",
				header)
			continue
		}
		for _, want := range []string{"tls.cert", "tls.certKey", "twice"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s: error %q must mention %q so the operator knows which field to change",
					header, err, want)
			}
		}
	}
	// A certificate-only file is the normal case and must still be accepted.
	cfg := &config.Config{}
	cfg.TLS.CertKey = key
	cfg.TLS.Cert = write("plain.crt", "-----BEGIN CERTIFICATE-----\nc\n-----END CERTIFICATE-----\n")
	if _, err := ServerCertBundle(cfg); err != nil {
		t.Errorf("a certificate-only file must be accepted: %v", err)
	}
}

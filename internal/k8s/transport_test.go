package k8s

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"solace/internal/broker"
	"solace/internal/config"
	"solace/internal/engine"
)

// --- capturing fake engine.Runner ------------------------------------------

type rrCall struct {
	method string // Run | RunInput | RunInteractive | Output | OutputInput
	name   string
	args   []string
	stdin  string
}

type recRunner struct {
	calls       []rrCall
	out         []byte   // returned by Output/OutputInput when outQueue is exhausted
	outQueue    [][]byte // consumed in order by successive Output calls (for multi-read ops)
	outErr      error    // error returned by Output/OutputInput (nil = success)
	outErrQueue []error  // popped per Output/OutputInput call; falls back to outErr once drained
	runErr      error    // error returned by Run (nil = success); for best-effort paths
	runErrQueue []error  // popped per Run call; falls back to runErr once drained
	runInputErr error    // error returned by RunInput (nil = success)

	// Cluster.Preflight runs `auth can-i` before every mutating operation, so the
	// double answers it out of band: the reply comes from these two fields rather
	// than from out/outQueue, which keeps every pre-existing scripted read aligned
	// with the call it was written for. Empty canI means "yes" -- the permitted
	// case, which is what almost every test is about. A test that exercises a
	// refusal sets canI="no"; one that exercises an unreachable server sets canIErr.
	canI    string
	canIErr error
}

// isCanI reports whether an argv is the Preflight probe. Matched on the token
// rather than a position so a configured kubernetes.runtime with leading arguments still
// resolves to the same answer.
func isCanI(args []string) bool {
	for _, a := range args {
		if a == "can-i" {
			return true
		}
	}
	return false
}

// canIAnswer is the scripted Preflight reply, defaulting to the permitted case.
func (r *recRunner) canIAnswer() ([]byte, error) {
	if r.canI == "" {
		return []byte("yes\n"), r.canIErr
	}
	return []byte(r.canI + "\n"), r.canIErr
}

// nextOut returns the next queued Output body, falling back to out once the queue
// is drained, so a test can script differing results for sequential kubectl reads.
func (r *recRunner) nextOut() []byte {
	if len(r.outQueue) > 0 {
		o := r.outQueue[0]
		r.outQueue = r.outQueue[1:]
		return o
	}
	return r.out
}

// nextRunErr returns the next queued Run error, falling back to runErr once the
// queue is drained, so a test can fail one specific call in a multi-step op (e.g.
// let a scale succeed but the rollout that follows it fail).
func (r *recRunner) nextRunErr() error {
	if len(r.runErrQueue) > 0 {
		e := r.runErrQueue[0]
		r.runErrQueue = r.runErrQueue[1:]
		return e
	}
	return r.runErr
}

// nextOutErr returns the next queued Output/OutputInput error, falling back to
// outErr once the queue is drained -- the Output-side mirror of nextRunErr, so a
// test can fail one specific read in a multi-step op (e.g. let the RBAC precheck
// succeed but the node listing that follows it fail).
func (r *recRunner) nextOutErr() error {
	if len(r.outErrQueue) > 0 {
		e := r.outErrQueue[0]
		r.outErrQueue = r.outErrQueue[1:]
		return e
	}
	return r.outErr
}

func (r *recRunner) Run(_ context.Context, name string, args ...string) error {
	r.calls = append(r.calls, rrCall{"Run", name, args, ""})
	return r.nextRunErr()
}
func (r *recRunner) RunInput(_ context.Context, in []byte, name string, args ...string) error {
	r.calls = append(r.calls, rrCall{"RunInput", name, args, string(in)})
	return r.runInputErr
}

// RunEnv exists to satisfy engine.Runner: nothing in the k8s package passes
// secrets through a child environment (every secret rides stdin), so it records
// the call like Run and the extra environment is deliberately not modelled.
func (r *recRunner) RunEnv(_ context.Context, _ []string, name string, args ...string) error {
	r.calls = append(r.calls, rrCall{"RunEnv", name, args, ""})
	return r.nextRunErr()
}
func (r *recRunner) RunInteractive(_ context.Context, name string, args ...string) error {
	r.calls = append(r.calls, rrCall{"RunInteractive", name, args, ""})
	return nil
}
func (r *recRunner) Output(_ context.Context, name string, args ...string) ([]byte, error) {
	r.calls = append(r.calls, rrCall{"Output", name, args, ""})
	if isCanI(args) {
		return r.canIAnswer()
	}
	return r.nextOut(), r.nextOutErr()
}
func (r *recRunner) OutputInput(_ context.Context, in []byte, name string, args ...string) ([]byte, error) {
	r.calls = append(r.calls, rrCall{"OutputInput", name, args, string(in)})
	return r.out, r.nextOutErr()
}

// afterPreflight asserts the FIRST recorded call is the read-only permission probe
// for verb/resource, and returns everything recorded after it. Every mutating
// Cluster operation must ask before it acts, so this does double duty: it pins that
// ordering, and it lets the call-shape assertions below keep counting only the work
// they are actually about. A test that expects no probe simply reads r.calls.
func (r *recRunner) afterPreflight(t *testing.T, verb, resource string) []rrCall {
	t.Helper()
	if len(r.calls) == 0 {
		t.Fatalf("no calls recorded: the read-only `auth can-i %s %s` probe must run before anything else", verb, resource)
	}
	first := r.calls[0]
	// A cluster-scoped resource is probed with --all-namespaces, never -n. Passing
	// -n made kubectl fail with an unrelated NotFound when the namespace did not
	// exist yet; but merely OMITTING it is not enough either, because kubectl then
	// falls back to the kubeconfig context's namespace -- which both prints
	// "resource X is not namespace scoped" and puts a namespace in the access
	// review, where RBAC would also evaluate namespaced Roles and could answer yes
	// for a permission the cluster-scoped action does not actually have.
	want := []string{"auth", "can-i", verb, resource}
	if clusterScoped[resource] {
		want = append(want, "--all-namespaces")
	} else {
		want = append(want, "-n", "solace")
	}
	if first.method != "Output" || !eqArgs(first.args, want) {
		t.Fatalf("first call = %+v, want the preflight probe Output %v", first, want)
	}
	return r.calls[1:]
}

func (r *recRunner) last() rrCall {
	if len(r.calls) == 0 {
		return rrCall{}
	}
	return r.calls[len(r.calls)-1]
}

func eqArgs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// --- exec argv (Run / Output / OutputInput), no -c ever --------------------

func TestTransportExecArgs(t *testing.T) {
	cfg := haCfg()
	rr := &recRunner{}
	tr := NewTransport(rr, cfg)
	ctx := context.Background()

	cases := []struct {
		name string
		call func()
		want rrCall
	}{
		{
			"Run primary",
			func() { _ = tr.Run(ctx, config.Primary, "show", "version") },
			rrCall{"Run", "kubectl", []string{"exec", "-n", "solace", "dev-broker-pubsubplus-p-0", "--", "show", "version"}, ""},
		},
		{
			"Output backup",
			func() { _, _ = tr.Output(ctx, config.Backup, "hostname") },
			rrCall{"Output", "kubectl", []string{"exec", "-n", "solace", "dev-broker-pubsubplus-b-0", "--", "hostname"}, ""},
		},
		{
			"OutputInput monitor adds -i and rides stdin",
			func() { _, _ = tr.OutputInput(ctx, config.Monitor, []byte("user = \"a:b\"\n"), "curl", "-K", "-") },
			rrCall{"OutputInput", "kubectl", []string{"exec", "-i", "-n", "solace", "dev-broker-pubsubplus-m-0", "--", "curl", "-K", "-"}, "user = \"a:b\"\n"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.call()
			got := rr.last()
			if got.method != tc.want.method || got.name != tc.want.name || !eqArgs(got.args, tc.want.args) || got.stdin != tc.want.stdin {
				t.Errorf("%s\n got: %+v\nwant: %+v", tc.name, got, tc.want)
			}
			for _, a := range got.args {
				if a == "-c" {
					t.Error("broker pods are single-container: -c must never appear")
				}
			}
		})
	}
}

// --- Upload: body on stdin via `sh -c 'cat > <dest>'`, never in argv --------

func TestTransportUpload(t *testing.T) {
	cfg := haCfg()
	rr := &recRunner{}
	tr := NewTransport(rr, cfg)

	secret := "PRIVATE-KEY-MATERIAL"
	dest := "/usr/sw/jail/certs/tls-2026-07-31.crt.key"
	if err := tr.Upload(context.Background(), config.Primary, []byte(secret), dest); err != nil {
		t.Fatalf("Upload: %v", err)
	}
	got := rr.last()
	want := rrCall{
		"RunInput", "kubectl",
		[]string{"exec", "-i", "-n", "solace", "dev-broker-pubsubplus-p-0", "--", "sh", "-c", "cat > '" + dest + "'"},
		secret,
	}
	if got.method != want.method || got.name != want.name || !eqArgs(got.args, want.args) || got.stdin != want.stdin {
		t.Errorf("Upload\n got: %+v\nwant: %+v", got, want)
	}
	if strings.Contains(strings.Join(got.args, " "), secret) {
		t.Error("secret body leaked into the Upload argv")
	}
}

// TestTransportUploadQuotesDest guards the defensive single-quote escaping so a
// metacharacter in a path cannot break out of the `cat >` redirect.
func TestTransportUploadQuotesDest(t *testing.T) {
	rr := &recRunner{}
	tr := NewTransport(rr, haCfg())
	if err := tr.Upload(context.Background(), config.Primary, []byte("x"), "/tmp/a'b"); err != nil {
		t.Fatalf("Upload: %v", err)
	}
	shArg := rr.last().args[len(rr.last().args)-1]
	if shArg != `cat > '/tmp/a'\''b'` {
		t.Errorf("dest quoting = %q", shArg)
	}
}

// --- kubectl cp, both directions, with -n <ns> ------------------------------

func TestTransportCopy(t *testing.T) {
	cfg := haCfg()
	rr := &recRunner{}
	tr := NewTransport(rr, cfg)
	ctx := context.Background()

	if err := tr.UploadFile(ctx, config.Primary, "local/in.cli", "/usr/sw/jail/cliscripts/.x.cli"); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}
	up := rr.last()
	wantUp := []string{"cp", "-n", "solace", "local/in.cli", "dev-broker-pubsubplus-p-0:/usr/sw/jail/cliscripts/.x.cli"}
	if up.method != "Run" || up.name != "kubectl" || !eqArgs(up.args, wantUp) {
		t.Errorf("UploadFile argv\n got: %+v\nwant cp %v", up, wantUp)
	}

	if err := tr.Download(ctx, config.Monitor, "/usr/sw/jail/logs/diag.tgz", "out/diag.tgz"); err != nil {
		t.Fatalf("Download: %v", err)
	}
	dn := rr.last()
	wantDn := []string{"cp", "-n", "solace", "dev-broker-pubsubplus-m-0:/usr/sw/jail/logs/diag.tgz", "out/diag.tgz"}
	if dn.method != "Run" || dn.name != "kubectl" || !eqArgs(dn.args, wantDn) {
		t.Errorf("Download argv\n got: %+v\nwant cp %v", dn, wantDn)
	}
}

// --- end-to-end: a real broker.Ops over engine.Echo never echoes the body ---

func TestTransportEchoHidesUploadBody(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := haCfg()
	tr := NewTransport(engine.Echo{W: buf}, cfg)
	o := broker.New(tr, cfg, nil)

	body := "SECRET-CLI-BODY\n"
	if _, err := o.RunCLI(context.Background(), config.Primary, "probe", body); err != nil {
		t.Fatalf("RunCLI over Echo: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "SECRET-CLI-BODY") {
		t.Errorf("Echo leaked the uploaded body:\n%s", out)
	}
	if !strings.Contains(out, "bytes on stdin") {
		t.Errorf("Echo should show the upload as a byte count:\n%s", out)
	}
	// The exec itself is still echoed as a normal command against the primary pod.
	// RunCLI now sends one `sh -c <skeleton>` rather than a bare `cli -Apes`, so the
	// anchors are the pod and the shell it is handed -- the skeleton's own body is
	// multi-line and shell-quoted, and matching a fragment of it would break on any
	// whitespace change without protecting anything.
	if !strings.Contains(out, "dev-broker-pubsubplus-p-0 -- sh -c ") {
		t.Errorf("Echo missing the cli exec line:\n%s", out)
	}
	// The shell still runs the CLI, and runs THIS call's script: the broker-side
	// filenames are derived from the script name RunCLI was given, and they are the
	// only part of the traced line that is. Dropping this would leave the whole
	// assertion satisfiable by constant skeleton text, so a RunCLI that stopped
	// threading the name through would pass.
	if !strings.Contains(out, broker.CLIBinary) || !strings.Contains(out, "-Apes") {
		t.Errorf("Echo's exec line does not invoke the CLI:\n%s", out)
	}
	if !strings.Contains(out, "solace-util-cli-probe") {
		t.Errorf("Echo's exec line does not name this script's own broker-side files:\n%s", out)
	}
}

// afterPreflights is afterPreflight's counterpart for an operation that probes SEVERAL
// permissions. It consumes every leading `auth can-i` call, asserts the set contains
// each wanted verb/resource, and returns the calls that follow.
//
// The single-probe helper cannot serve here: it asserts calls[0] is the only probe and
// returns calls[1:], so an operation that legitimately asks about seven kinds would
// look like six stray calls before its real work.
func (r *recRunner) afterPreflights(t *testing.T, want ...probe) []rrCall {
	t.Helper()
	got := map[string]bool{}
	i := 0
	for ; i < len(r.calls); i++ {
		a := r.calls[i].args
		if len(a) < 4 || a[0] != "auth" || a[1] != "can-i" {
			break
		}
		got[a[2]+" "+a[3]] = true
	}
	if i == 0 {
		t.Fatal("no leading `auth can-i` probe: every mutating operation must ask before it acts")
	}
	for _, p := range want {
		if !got[p.verb+" "+p.resource] {
			t.Errorf("no probe for %q %q -- an operation must ask about every kind it touches, or it passes "+
				"the check and fails halfway through the work", p.verb, p.resource)
		}
	}
	return r.calls[i:]
}

// probedNamespace returns the -n value of the leading probe for verb/resource, or ""
// when it carried none. It exists because a probe asking about the wrong namespace is
// silent: it answers, it just answers a question nobody asked.
func (r *recRunner) probedNamespace(verb, resource string) (string, bool) {
	for _, c := range r.calls {
		a := c.args
		if len(a) < 4 || a[0] != "auth" || a[1] != "can-i" || a[2] != verb || a[3] != resource {
			continue
		}
		for j := 4; j+1 < len(a); j++ {
			if a[j] == "-n" {
				return a[j+1], true
			}
		}
		return "", true
	}
	return "", false
}

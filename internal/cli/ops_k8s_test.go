package cli

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"solace/internal/config"
	"solace/internal/engine"
	"solace/internal/k8s"
)

// ops_k8s_test.go drives the Kubernetes handlers of ops_k8s.go directly, over opRunner
// (cli_test.go) rather than engine.Echo, for the branches the echo seam cannot reach: every
// confirmation gate in each of its outcomes, every refusal that has to land BEFORE anything
// is issued, and the failures a real cluster hands back mid-operation. Echo answers every
// read with nothing and never fails, so it can show which argv a command would issue but not
// that a refused or failed step stopped the ones after it. opRunner records every call and
// can fail one on demand, which is what those properties are asserted against.

// opsK8sRun runs fn with stdout and stderr both captured. Every direct-call test below
// needs the pair: reports land on one stream, narration, prompts and refusals on the other,
// and which stream a line reached is part of the contract.
func opsK8sRun(t *testing.T, fn func() error) (stdout, stderr string, err error) {
	t.Helper()
	stderr = captureStderr(t, func() {
		stdout = captureStdout(t, func() { err = fn() })
	})
	return stdout, stderr, err
}

// opsK8sPods lists, by role letter and in the order each first appears, the broker pods the
// recorded calls addressed: "pb" is the primary then the backup, "pbm" the whole group and
// "" no pod at all. Every fixture here names its broker dev-broker, so a pod is
// dev-broker-pubsubplus-<p|b|m>-0.
func opsK8sPods(rr *opRunner) string {
	seen := ""
	for _, c := range rr.calls {
		for _, letter := range []string{"p", "b", "m"} {
			if opArgvMatch(c, "dev-broker-pubsubplus-"+letter+"-0") && !strings.Contains(seen, letter) {
				seen += letter
			}
		}
	}
	return seen
}

// opsK8sAnswer is an App seam that makes the run interactive and types reply at it: "y\n"
// is a yes and "n\n" a no. The reply is read only if a question is actually asked, so a
// test can also prove that nothing asked.
func opsK8sAnswer(reply string) func(*App) {
	return func(a *App) {
		a.Interactive = func() bool { return true }
		a.PromptIn = strings.NewReader(reply)
	}
}

// opsK8sUnattended is an App seam for a run with nobody at the terminal: no question is
// asked, and a gate that needs an answer refuses and names --no-prompt.
func opsK8sUnattended(a *App) { a.Interactive = func() bool { return false } }

// opsK8sSilenced is --no-prompt: every question is answered yes without being asked.
func opsK8sSilenced(a *App) {
	a.noPrompt = true
	a.Interactive = func() bool { return false }
}

// TestOpsK8sScopedValidateViewsRenderOnlyTheirOwnHalf drives `broker validate` and
// `operator validate` through the command tree and checks each is wired to ITS scope of the
// one report: the broker view shows the Config, Deployment and Broker sections and no
// Operator section, the operator view shows only the Operator section, and each carries the
// operator-install verdict exactly once. A swapped pair would compile, run and print a
// plausible report, so only the section structure can tell the two views apart.
func TestOpsK8sScopedValidateViewsRenderOnlyTheirOwnHalf(t *testing.T) {
	path := writeStandaloneEnv(t)
	cases := []struct {
		name    string
		args    []string
		want    []string
		notWant []string
	}{
		{"broker validate", []string{"broker", "validate"},
			[]string{"=== Config ", "=== Deployment ", "=== Broker "}, []string{"=== Operator "}},
		{"operator validate", []string{"operator", "validate"},
			[]string{"=== Operator "}, []string{"=== Config ", "=== Deployment ", "=== Broker "}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := runStandalone(t, path, append(tc.args, "--platform", "kubernetes")...)
			if err != nil {
				t.Fatalf("%s err = %v, want nil: a preview never fails the report", tc.name, err)
			}
			for _, want := range tc.want {
				if !strings.Contains(out, want) {
					t.Errorf("%s stdout lacks the %q section:\n%s", tc.name, want, out)
				}
			}
			for _, notWant := range tc.notWant {
				if strings.Contains(out, notWant) {
					t.Errorf("%s stdout carries the %q section, which is not this view's:\n%s", tc.name, notWant, out)
				}
			}
			if got := strings.Count(out, "[SKIP] operator install"); got != 1 {
				t.Errorf("%s shows the operator-install verdict %d time(s), want exactly once:\n%s", tc.name, got, out)
			}
		})
	}
}

// TestOpsK8sDeployRefusesAnUnreadableCertificateBeforeAnyChange pins the first thing a
// deploy does with a TLS pair it builds the Secret from: read it, to compare its digest with
// the live Secret's. A certificate that cannot be read must stop the deploy THERE -- before
// the namespace, the Secrets or the CR is applied -- because everything after it would be a
// change to a cluster that the same deploy could not finish describing. Nothing reaches the
// cluster at all, not even a permission probe.
func TestOpsK8sDeployRefusesAnUnreadableCertificateBeforeAnyChange(t *testing.T) {
	cfg := writeK8sTLSEnv(t, "false")
	if err := os.Remove(cfg.TLS.Cert); err != nil {
		t.Fatalf("remove the certificate: %v", err)
	}
	rr := &opRunner{output: k8sDeployAllOutputHook}
	a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr}
	_, _, err := opsK8sRun(t, func() error { return opK8sDeploy(a) })
	if err == nil || !strings.Contains(err.Error(), "read tls.cert") || !strings.Contains(err.Error(), cfg.TLS.Cert) {
		t.Fatalf("opK8sDeploy err = %v, want the unreadable certificate named", err)
	}
	if len(rr.calls) != 0 {
		t.Errorf("nothing may reach the cluster once the certificate cannot be read:\n%s", rr.dump())
	}
}

// TestOpsK8sDeployReadBackFailureStopsBeforeACertificateRestart: `apply` exiting 0 does not
// prove the CR exists, so a deploy reads it back before it reports success -- and a read-back
// that fails is the end of the deploy, with the cause preserved and the next step named. The
// fixture is built so that swallowing the failure would be visible: the server certificate
// is renewed, the broker is running and --restart pre-approves the bounce, so a deploy that
// carried on would delete pods. This one must not even look for them.
func TestOpsK8sDeployReadBackFailureStopsBeforeACertificateRestart(t *testing.T) {
	cfg := writeK8sTLSEnv(t, "true")
	stale := strings.Repeat("0", 64)
	cause := errors.New("injected read-back failure")
	rr := &opRunner{
		output: func(c opCall) []byte {
			switch {
			case opArgvMatch(c, "jsonpath={.metadata.name}"):
				return []byte("dev-broker-tls\n" + stale)
			case isGetPod(c):
				return []byte("pod/dev-broker-pubsubplus-p-0\n")
			}
			return k8sDeployAllOutputHook(c)
		},
		fail: func(c opCall) error {
			if c.method == "Output" && opArgvMatch(c, "pubsubpluseventbrokers") && opArgvMatch(c, "json") {
				return cause
			}
			return nil
		},
	}
	a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr, restart: true}
	_, _, err := opsK8sRun(t, func() error { return opK8sDeploy(a) })
	if !errors.Is(err, cause) {
		t.Fatalf("opK8sDeploy err = %v, want the read-back failure carried through", err)
	}
	for _, want := range []string{`applied broker "dev-broker" but could not read it back in namespace "solace"`, "solace-util broker validate"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %q, want it to say %q", err, want)
		}
	}
	if got := rr.callCount("apply"); got != 3 {
		t.Errorf("%d apply call(s), want 3 (namespace, Secrets, CR): the failure is the read-back, not an apply:\n%s", got, rr.dump())
	}
	for _, c := range rr.calls {
		if isGetPod(c) || isDeletePod(c) {
			t.Errorf("a deploy whose read-back failed went on to consider a certificate restart: %v", c.args)
		}
	}
}

// TestOpsK8sConfigureLeavesRefuseOppositeDirections: each `broker configure` leaf has a
// default direction and a flag for the other, and both flags of a pair at once is a usage
// error read by the pair's own helper. The refusal has to come before ANYTHING else in the
// leaf -- no resolution, no question, no call -- or a contradictory command line could
// still act on whichever flag the handler happened to read first.
func TestOpsK8sConfigureLeavesRefuseOppositeDirections(t *testing.T) {
	both := func(a *App) { a.flagApply, a.flagRemove = true, true }
	opposed := func(a *App) { a.flagDisable, a.flagEnable = true, true }
	cases := []struct {
		name string
		op   func(*App) error
		set  func(*App)
		want string
	}{
		{"server-certs", opK8sConfigServerCerts, both, "--apply and --remove"},
		{"domain-certs", opK8sConfigDomainCerts, both, "--apply and --remove"},
		{"product-keys", opK8sConfigProductKeys, both, "--apply and --remove"},
		{"default-vpn", opK8sConfigDefaultVPN, opposed, "--disable and --enable"},
		{"default-users", opK8sConfigDefaultUsers, opposed, "--disable and --enable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := &opRunner{}
			reads := 0
			a := &App{Cfg: writeK8sDeployAllEnv(t, "true"), Platform: config.K8s, Runner: rr, Interactive: func() bool { return true },
				PromptIn: exportconfigReadCounter{Reader: strings.NewReader("yes\n"), n: &reads}}
			tc.set(a)
			_, _, err := opsK8sRun(t, func() error { return tc.op(a) })
			if err == nil || ExitCode(err) != exitUsage || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v (exit %d), want a usage refusal naming %q", err, ExitCode(err), tc.want)
			}
			if len(rr.calls) != 0 || reads != 0 {
				t.Errorf("a contradictory request must act and ask nothing: %d call(s), %d prompt read(s):\n%s", len(rr.calls), reads, rr.dump())
			}
		})
	}
}

// TestOpsK8sABadPodIsRefusedBeforeAnythingRuns: --pod names a role, and a word that is not
// one is a usage error the operator can fix by retyping -- so it has to be refused before
// anything else happens. For every handler that resolves a pod that means no call, no
// question, and no side effect: the server-certificate Secret is not rewritten first, the
// import's file is not read, and the diagnostics directory is not created.
func TestOpsK8sABadPodIsRefusedBeforeAnythingRuns(t *testing.T) {
	tlsHA := func(t *testing.T) *config.Config { return writeK8sTLSEnv(t, "true") }
	ha := func(t *testing.T) *config.Config { return writeK8sDeployAllEnv(t, "true") }
	none := func(*App) {}
	remove := func(a *App) { a.flagRemove = true }
	enable := func(a *App) { a.flagEnable = true }
	// The App is built without cobra, so --days' default of 1 never lands; without it the
	// --days refusal would answer first and the --pod one would go untested.
	days := func(a *App) { a.days = 1 }
	cases := []struct {
		name string
		cfg  func(*testing.T) *config.Config
		set  func(*App)
		op   func(*App) error
	}{
		{"server-certs", tlsHA, none, opK8sConfigServerCerts},
		{"server-certs --remove", ha, remove, opK8sConfigServerCerts},
		{"domain-certs", ha, none, opK8sConfigDomainCerts},
		{"domain-certs --remove", ha, remove, opK8sConfigDomainCerts},
		{"product-keys", ha, none, opK8sConfigProductKeys},
		{"product-keys --remove", ha, remove, opK8sConfigProductKeys},
		{"default-vpn", ha, none, opK8sConfigDefaultVPN},
		{"default-vpn --enable", ha, enable, opK8sConfigDefaultVPN},
		{"default-users", ha, none, opK8sConfigDefaultUsers},
		{"default-users --enable", ha, enable, opK8sConfigDefaultUsers},
		{"import-config", ha, none, func(a *App) error { return opK8sImportConfig(a, "export.cli") }},
		{"copy from", ha, none, func(a *App) error { return opK8sCopyFrom(a, []string{"msgrates.log"}) }},
		{"copy into", ha, none, func(a *App) error { return opK8sCopyInto(a, []string{"notes.txt"}) }},
		{"gather-diagnostics", ha, days, opK8sVerifyDiagnostics},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diag := filepath.Join(t.TempDir(), "diag")
			cfg := tc.cfg(t)
			cfg.Broker.HostDiagnosticDir = diag
			rr := &opRunner{}
			reads := 0
			a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr, pod: "bogus", Interactive: func() bool { return true },
				PromptIn: exportconfigReadCounter{Reader: strings.NewReader("yes\n"), n: &reads}}
			tc.set(a)
			_, _, err := opsK8sRun(t, func() error { return tc.op(a) })
			if err == nil || ExitCode(err) != exitUsage || !strings.Contains(err.Error(), `invalid node role "bogus"`) {
				t.Fatalf("err = %v (exit %d), want a usage error naming the bad role", err, ExitCode(err))
			}
			if len(rr.calls) != 0 || reads != 0 {
				t.Errorf("a refused --pod must act and ask nothing: %d call(s), %d prompt read(s):\n%s", len(rr.calls), reads, rr.dump())
			}
			if _, statErr := os.Stat(diag); !errors.Is(statErr, os.ErrNotExist) {
				t.Errorf("the diagnostics directory was created for a refused request: %v", statErr)
			}
		})
	}
}

// TestOpsK8sPodRolesAnnouncesWhatItNarrows pins podRoles against the role set an operation
// would use by default. An unset --pod keeps the whole set, which is what makes the fixed
// sets meaningful (a certificate on every node, product keys never on the monitor).
// Narrowing a set of several to one node is allowed but said out loud, because a
// half-licensed or half-certified redundancy group is almost always a mistake rather than a
// plan; narrowing a set of one says nothing, there being nothing to leave out. And a word
// that is not a role is a usage error that returns no roles at all.
func TestOpsK8sPodRolesAnnouncesWhatItNarrows(t *testing.T) {
	all := []config.Role{config.Primary, config.Backup, config.Monitor}
	pair := []config.Role{config.Primary, config.Backup}
	single := []config.Role{config.Primary}
	cases := []struct {
		name     string
		pod      string
		defaults []config.Role
		want     []config.Role
		warning  string
	}{
		{"unset keeps the whole default set", "", all, all, ""},
		{"a role word narrows a group of three and says so", "backup", all, []config.Role{config.Backup},
			"--pod b narrows this to 1 of 3 node(s) that would otherwise be covered"},
		{"an abbreviation narrows a pair and says so", "b", pair, []config.Role{config.Backup},
			"--pod b narrows this to 1 of 2 node(s) that would otherwise be covered"},
		{"naming the only node narrows nothing and says nothing", "primary", single, single, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := &App{pod: tc.pod}
			var got []config.Role
			var err error
			errOut := captureStderr(t, func() { got, err = podRoles(a, tc.defaults) })
			if err != nil {
				t.Fatalf("podRoles(%q) err = %v, want nil", tc.pod, err)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("podRoles(%q) = %v, want %v", tc.pod, got, tc.want)
			}
			if tc.warning == "" && errOut != "" {
				t.Errorf("podRoles(%q) said %q, want it silent: nothing was left out", tc.pod, errOut)
			}
			if tc.warning != "" && !strings.Contains(errOut, tc.warning) {
				t.Errorf("podRoles(%q) stderr = %q, want it to say %q", tc.pod, errOut, tc.warning)
			}
		})
	}
	t.Run("a word that is not a role is a usage error and returns no roles", func(t *testing.T) {
		var got []config.Role
		var err error
		errOut := captureStderr(t, func() { got, err = podRoles(&App{pod: "bogus"}, all) })
		if got != nil || err == nil || ExitCode(err) != exitUsage || !strings.Contains(err.Error(), `invalid node role "bogus"`) {
			t.Fatalf("podRoles(bogus) = %v, %v (exit %d), want no roles and a usage error naming the word", got, err, ExitCode(err))
		}
		if errOut != "" {
			t.Errorf("a refused --pod must not announce a narrowing it never made: %q", errOut)
		}
	})
}

// TestOpsK8sRemoveServerCertsConfirmsThenSpansTheGroup covers the --remove direction on a
// deployment whose certificate is loaded over the broker CLI. It takes TLS down on a live
// broker, so it asks first -- naming the broker, the namespace and the kubeconfig context,
// the three facts that decide which cluster is about to lose its certificate -- and --no-prompt
// is the one unattended answer. Confirmed, it spans every node (a certificate gone from one
// node and still loaded on another is a half state nobody asked for), narrows to one only
// when --pod says so, and rewrites no Secret. Declined or unanswerable, nothing reaches a pod.
func TestOpsK8sRemoveServerCertsConfirmsThenSpansTheGroup(t *testing.T) {
	const question = `Remove the TLS server certificate from broker dev-broker in namespace solace (context "prod-cluster")? [y/N] `
	const refusal = `refusing to remove the TLS server certificate from broker dev-broker in namespace solace (context "prod-cluster") without confirmation; pass --no-prompt to proceed`
	narrowed := func(a *App) {
		opsK8sSilenced(a)
		a.pod = "backup"
	}
	cases := []struct {
		name     string
		set      func(*App)
		wantPods string
		asked    bool
		stderr   string
	}{
		{"--no-prompt removes it from every node without asking", opsK8sSilenced, "pbm", false, ""},
		{"a yes at the prompt removes it from every node", opsK8sAnswer("y\n"), "pbm", true, question},
		{"a no at the prompt leaves every node alone", opsK8sAnswer("n\n"), "", true, question},
		{"no terminal and no --no-prompt leaves every node alone and names the flag", opsK8sUnattended, "", false, refusal},
		{"--pod narrows the removal to that node and says so", narrowed, "b", false, "--pod b narrows this to 1 of 3 node(s)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := &opRunner{}
			a := &App{Cfg: writeK8sDeployAllEnv(t, "true"), Platform: config.K8s, Runner: rr, flagRemove: true, kubeContext: "prod-cluster"}
			tc.set(a)
			_, errOut, err := opsK8sRun(t, func() error { return opK8sConfigServerCerts(a) })
			if err != nil {
				t.Fatalf("configure server-certs --remove err = %v, want nil", err)
			}
			if got := opsK8sPods(rr); got != tc.wantPods {
				t.Errorf("the removal reached pods %q, want %q:\n%s", got, tc.wantPods, rr.dump())
			}
			removals := 0
			for _, c := range rr.calls {
				if c.method == "RunInput" {
					t.Errorf("--remove is CLI-only and must not rewrite a Secret, but applied one: %+v", c)
				}
				if c.method == "OutputInput" && strings.Contains(c.stdin, "no ssl server-certificate") {
					removals++
				}
			}
			if removals != len(tc.wantPods) {
				t.Errorf("%d removal script(s) ran, want one per node (%d)", removals, len(tc.wantPods))
			}
			if tc.stderr != "" && !strings.Contains(errOut, tc.stderr) {
				t.Errorf("stderr = %q, want it to carry %q", errOut, tc.stderr)
			}
			if asked := strings.Contains(errOut, "[y/N]"); asked != tc.asked {
				t.Errorf("a question was asked = %v, want %v:\n%s", asked, tc.asked, errOut)
			}
		})
	}
}

// TestOpsK8sDomainCertsResolveBeforeAnythingIsAskedOrIssued: the directory walk runs at the
// op, for BOTH directions, so "the configured dirs must be readable" is checked where it is
// first true. A directory that cannot be read ends the command before the --remove question
// is asked and before any upload -- otherwise a removal would act on a smaller set than the
// apply it undoes, and a directory-derived CA would be silently left on the broker.
func TestOpsK8sDomainCertsResolveBeforeAnythingIsAskedOrIssued(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-certs")
	for _, remove := range []bool{false, true} {
		name := "apply"
		if remove {
			name = "--remove"
		}
		t.Run("an unreadable dir refuses "+name, func(t *testing.T) {
			cfg := writeK8sDeployAllEnv(t, "false")
			cfg.Broker.DomainCerts.Dirs = []config.CertDir{{Path: missing}}
			rr := &opRunner{}
			reads := 0
			a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr, flagRemove: remove, Interactive: func() bool { return true },
				PromptIn: exportconfigReadCounter{Reader: strings.NewReader("y\n"), n: &reads}}
			_, _, err := opsK8sRun(t, func() error { return opK8sConfigDomainCerts(a) })
			if err == nil || !strings.Contains(err.Error(), "broker.domainCerts.dirs[0]") || !strings.Contains(err.Error(), missing) {
				t.Fatalf("opK8sConfigDomainCerts err = %v, want the unreadable directory named", err)
			}
			if len(rr.calls) != 0 || reads != 0 {
				t.Errorf("an unresolvable set must act and ask nothing: %d call(s), %d prompt read(s):\n%s", len(rr.calls), reads, rr.dump())
			}
		})
	}
}

// TestOpsK8sRemoveDomainCertsConfirmsAndTargetsOneNode covers the domain-certs --remove
// gate on Kubernetes. It deletes trust anchors from a live broker, so like every removal in
// the tree it asks (--no-prompt answers it), and a no or an unanswerable question changes
// nothing. Unlike the server certificate it acts on ONE node: the primary by default, and
// the one --pod names otherwise -- never the whole group.
func TestOpsK8sRemoveDomainCertsConfirmsAndTargetsOneNode(t *testing.T) {
	const question = "Remove the configured domain CA certificates from broker dev-broker in namespace solace? [y/N] "
	const refusal = "refusing to remove the configured domain CA certificates from broker dev-broker in namespace solace without confirmation; pass --no-prompt to proceed"
	backup := func(a *App) {
		opsK8sSilenced(a)
		a.pod = "backup"
	}
	cases := []struct {
		name     string
		set      func(*App)
		wantPods string
		asked    bool
		stderr   string
	}{
		{"--no-prompt removes the configured CA from the primary", opsK8sSilenced, "p", false, ""},
		{"a yes at the prompt removes it from the primary", opsK8sAnswer("y\n"), "p", true, question},
		{"a no at the prompt removes nothing", opsK8sAnswer("n\n"), "", true, question},
		{"no terminal and no --no-prompt removes nothing and names the flag", opsK8sUnattended, "", false, refusal},
		{"--pod backup removes it from the backup node only", backup, "b", false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := writeK8sDeployAllEnv(t, "true")
			cfg.Broker.DomainCerts.Files = map[string]string{"myca": "/opt/solace/certs/myca.pem"}
			rr := &opRunner{}
			a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr, flagRemove: true}
			tc.set(a)
			_, errOut, err := opsK8sRun(t, func() error { return opK8sConfigDomainCerts(a) })
			if err != nil {
				t.Fatalf("configure domain-certs --remove err = %v, want nil", err)
			}
			if got := opsK8sPods(rr); got != tc.wantPods {
				t.Errorf("the removal reached pods %q, want %q:\n%s", got, tc.wantPods, rr.dump())
			}
			removals := 0
			for _, c := range rr.calls {
				if c.method == "OutputInput" && strings.Contains(c.stdin, "no ssl domain-certificate-authority myca\n") {
					removals++
				}
			}
			if removals != len(tc.wantPods) {
				t.Errorf("%d removal script(s) for CA myca ran, want %d:\n%s", removals, len(tc.wantPods), rr.dump())
			}
			if tc.stderr != "" && !strings.Contains(errOut, tc.stderr) {
				t.Errorf("stderr = %q, want it to carry %q", errOut, tc.stderr)
			}
			if asked := strings.Contains(errOut, "[y/N]"); asked != tc.asked {
				t.Errorf("a question was asked = %v, want %v:\n%s", asked, tc.asked, errOut)
			}
		})
	}
}

// TestOpsK8sProductKeysGoToTheLicensedNodesOnly covers product-keys in both directions. The
// keys go to the primary and, in HA, the backup -- never the monitor, which carries no
// message spool -- and the revocation uses the same set, because a key revoked on one node
// and left on the other is a licensing state that changes at the next failover. Revoking can
// leave the broker unlicensed, so --remove asks first. A key is a licence string: it rides
// the script on stdin and never appears in an argv, where it would be logged and echoed.
func TestOpsK8sProductKeysGoToTheLicensedNodesOnly(t *testing.T) {
	const question = "Revoke every configured product key from broker dev-broker in namespace solace? [y/N] "
	const refusal = "refusing to revoke the product keys from broker dev-broker in namespace solace without confirmation; pass --no-prompt to proceed"
	const applied = "product-key KEY-ONE\nproduct-key KEY-TWO\n"
	const revoked = "no product-key KEY-ONE\nno product-key KEY-TWO\n"
	narrowed := func(a *App) {
		opsK8sSilenced(a)
		a.pod = "backup"
	}
	cases := []struct {
		name     string
		remove   bool
		set      func(*App)
		wantPods string
		script   string
		stderr   string
	}{
		{"apply licenses the primary and the backup, never the monitor", false, opsK8sUnattended, "pb", applied, ""},
		{"--remove --no-prompt revokes on the same two nodes", true, opsK8sSilenced, "pb", revoked, ""},
		{"--remove answered yes revokes on the same two nodes", true, opsK8sAnswer("y\n"), "pb", revoked, question},
		{"--remove answered no revokes nothing", true, opsK8sAnswer("n\n"), "", "", question},
		{"--remove with no terminal revokes nothing and names the flag", true, opsK8sUnattended, "", "", refusal},
		{"--remove --pod backup revokes on that node only and says so", true, narrowed, "b", revoked, "--pod b narrows this to 1 of 2 node(s)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := &opRunner{}
			a := &App{Cfg: writeK8sDeployAllEnv(t, "true"), Platform: config.K8s, Runner: rr, flagRemove: tc.remove}
			a.Cfg.Broker.ProductKeys = []string{"KEY-ONE", "KEY-TWO"}
			tc.set(a)
			_, errOut, err := opsK8sRun(t, func() error { return opK8sConfigProductKeys(a) })
			if err != nil {
				t.Fatalf("configure product-keys err = %v, want nil", err)
			}
			if got := opsK8sPods(rr); got != tc.wantPods {
				t.Errorf("the keys reached pods %q, want %q:\n%s", got, tc.wantPods, rr.dump())
			}
			scripts := 0
			for _, c := range rr.calls {
				if opArgvMatch(c, "KEY-ONE") || opArgvMatch(c, "KEY-TWO") {
					t.Errorf("a product key reached an argv, where it would be logged and echoed: %v", c.args)
				}
				if c.method == "OutputInput" && strings.Contains(c.stdin, tc.script) {
					scripts++
				}
			}
			if scripts != len(tc.wantPods) {
				t.Errorf("%d script(s) carried the expected lines, want one per licensed node (%d):\n%s", scripts, len(tc.wantPods), rr.dump())
			}
			if tc.stderr != "" && !strings.Contains(errOut, tc.stderr) {
				t.Errorf("stderr = %q, want it to carry %q", errOut, tc.stderr)
			}
		})
	}
}

// TestOpsK8sProductKeysRefuseALineBreakingKeyBeforeUploading is the security boundary under
// the product-key leaf. A key is written verbatim into one line of a Solace CLI script that
// runs with admin already enabled, so a key carrying a line break would add commands of its
// own. The executor must refuse it itself, in both directions, because it is handed a Config
// directly and may never have seen config.Load -- and it must refuse BEFORE the first byte
// is uploaded to any node.
func TestOpsK8sProductKeysRefuseALineBreakingKeyBeforeUploading(t *testing.T) {
	for _, remove := range []bool{false, true} {
		name := "apply"
		if remove {
			name = "--remove"
		}
		t.Run(name, func(t *testing.T) {
			rr := &opRunner{}
			a := &App{Cfg: writeK8sDeployAllEnv(t, "true"), Platform: config.K8s, Runner: rr, flagRemove: remove, noPrompt: true}
			a.Cfg.Broker.ProductKeys = []string{"KEY-ONE", "KEY-TWO\nadmin\nno product-key KEY-ONE"}
			_, _, err := opsK8sRun(t, func() error { return opK8sConfigProductKeys(a) })
			if err == nil || !strings.Contains(err.Error(), "control character") {
				t.Fatalf("opK8sConfigProductKeys err = %v, want the line-breaking key refused", err)
			}
			if len(rr.calls) != 0 {
				t.Errorf("a refused key must not reach any node:\n%s", rr.dump())
			}
		})
	}
}

// TestOpsK8sHardeningLeavesAskOnlyBeforeDisturbingClients covers default-vpn and
// default-users, whose disable direction (the default) stops every client that still relies
// on the VPN or the username, so it asks first and --no-prompt is the one unattended answer.
// Starting either back up is additive and never asks -- not even when the question would
// have been answered no, and not with nobody at the terminal. A declined or unanswerable
// shutdown reaches no node at all, and the script that runs is the one for the direction
// asked: the disable script never carries `no shutdown` and the enable script always does.
func TestOpsK8sHardeningLeavesAskOnlyBeforeDisturbingClients(t *testing.T) {
	// `show message-vpn *` as the broker prints it: the name column's width is the rule's.
	vpns := strings.Join([]string{
		"Message VPN" + strings.Repeat(" ", 29) + "Enabled",
		strings.Repeat("-", 40),
		"default" + strings.Repeat(" ", 33) + "Yes",
		"myvpn" + strings.Repeat(" ", 35) + "Yes",
	}, "\n")
	leaves := []struct {
		name      string
		op        func(*App) error
		what      string
		downName  string
		upName    string
		scriptHas []string
	}{
		{"default-vpn", opK8sConfigDefaultVPN, "the default message-VPN on broker dev-broker in namespace solace",
			"disable-default-vpn", "enable-default-vpn", []string{`message-vpn "default"`}},
		{"default-users", opK8sConfigDefaultUsers, "the default client-username in every message-VPN on broker dev-broker in namespace solace",
			"disable-default-usernames", "enable-default-usernames",
			[]string{`client-username default message-vpn "default"`, `client-username default message-vpn "myvpn"`}},
	}
	scenarios := []struct {
		name      string
		enable    bool
		set       func(*App)
		shouldRun bool
		wantText  string
	}{
		{"--no-prompt shuts it down without asking", false, opsK8sSilenced, true, ""},
		{"a yes at the prompt shuts it down", false, opsK8sAnswer("y\n"), true, "question"},
		{"a no at the prompt leaves it up", false, opsK8sAnswer("n\n"), false, "question"},
		{"no terminal and no --no-prompt leaves it up and names the flag", false, opsK8sUnattended, false, "refusal"},
		{"--enable starts it back up without a question", true, opsK8sAnswer("n\n"), true, ""},
		{"--enable with no terminal still starts it back up", true, opsK8sUnattended, true, ""},
	}
	for _, lf := range leaves {
		for _, sc := range scenarios {
			t.Run(lf.name+"/"+sc.name, func(t *testing.T) {
				rr := &opRunner{output: func(c opCall) []byte {
					if c.method == "Output" && opArgvMatch(c, ".show-vpn.cli") {
						return []byte(vpns)
					}
					return nil
				}}
				a := &App{Cfg: writeK8sDeployAllEnv(t, "true"), Platform: config.K8s, Runner: rr, flagEnable: sc.enable}
				sc.set(a)
				_, errOut, err := opsK8sRun(t, func() error { return lf.op(a) })
				if err != nil {
					t.Fatalf("%s err = %v, want nil", lf.name, err)
				}
				ran, other := lf.downName, lf.upName
				if sc.enable {
					ran, other = lf.upName, lf.downName
				}
				body, found := "", false
				for _, c := range rr.calls {
					if c.method == "OutputInput" && opArgvMatch(c, "solace-util-cli-"+ran) {
						body, found = c.stdin, true
					}
				}
				if found != sc.shouldRun {
					t.Fatalf("script %q ran = %v, want %v:\n%s", ran, found, sc.shouldRun, rr.dump())
				}
				if rr.hasCall("solace-util-cli-" + other) {
					t.Errorf("the script for the opposite direction (%q) ran:\n%s", other, rr.dump())
				}
				if !sc.shouldRun && len(rr.calls) != 0 {
					t.Errorf("a shutdown that was not confirmed must reach no node:\n%s", rr.dump())
				}
				if found {
					for _, want := range lf.scriptHas {
						if !strings.Contains(body, want) {
							t.Errorf("script %q lacks %q:\n%s", ran, want, body)
						}
					}
					if got := strings.Contains(body, "no shutdown"); got != sc.enable {
						t.Errorf("script %q carries `no shutdown` = %v, want %v:\n%s", ran, got, sc.enable, body)
					}
				}
				question := "Shut down " + lf.what + "? [y/N] "
				refusal := "refusing to shut down " + lf.what + " without confirmation; pass --no-prompt to proceed"
				switch sc.wantText {
				case "question":
					if !strings.Contains(errOut, question) {
						t.Errorf("stderr = %q, want the question %q", errOut, question)
					}
				case "refusal":
					if !strings.Contains(errOut, refusal) {
						t.Errorf("stderr = %q, want the refusal %q", errOut, refusal)
					}
				}
				if asked := strings.Contains(errOut, "[y/N]"); asked != (sc.wantText == "question") {
					t.Errorf("a question was asked = %v, want %v:\n%s", asked, sc.wantText == "question", errOut)
				}
				if refused := strings.Contains(errOut, "refusing"); refused != (sc.wantText == "refusal") {
					t.Errorf("a refusal line was printed = %v, want %v:\n%s", refused, sc.wantText == "refusal", errOut)
				}
			})
		}
	}
}

// TestOpsK8sGatherDiagnosticsCoversEveryNodeOrTheOneNamed: --pod used to be registered and
// ignored, so every run gathered the whole group whatever it said. It now narrows exactly as
// the server-certificate and product-key steps do, and announces what it left out. Each
// node's show output is downloaded into broker.hostDiagnosticDir as a .tgz, the dir created on
// the way, and --days and --encrypt reach the script the broker runs -- the day count bare.
func TestOpsK8sGatherDiagnosticsCoversEveryNodeOrTheOneNamed(t *testing.T) {
	cases := []struct {
		name     string
		pod      string
		encrypt  bool
		wantPods string
		stderr   string
	}{
		{"no --pod gathers from every node", "", false, "pbm", ""},
		{"--pod narrows it to that node and says so", "backup", false, "b", "--pod b narrows this to 1 of 3 node(s)"},
		{"--encrypt drops no-encrypt", "", true, "pbm", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diag := filepath.Join(t.TempDir(), "diag")
			cfg := writeK8sDeployAllEnv(t, "true")
			cfg.Broker.HostDiagnosticDir = diag
			rr := &opRunner{output: func(c opCall) []byte {
				if c.method == "Output" && opArgvMatch(c, "hostname") {
					return []byte("pod-host\n")
				}
				return nil
			}}
			a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr, pod: tc.pod, days: 3, encrypt: tc.encrypt}
			_, errOut, err := opsK8sRun(t, func() error { return opK8sVerifyDiagnostics(a) })
			if err != nil {
				t.Fatalf("gather-diagnostics err = %v, want nil", err)
			}
			if got := opsK8sPods(rr); got != tc.wantPods {
				t.Errorf("diagnostics were gathered from pods %q, want %q:\n%s", got, tc.wantPods, rr.dump())
			}
			if fi, statErr := os.Stat(diag); statErr != nil || !fi.IsDir() {
				t.Errorf("the diagnostics directory %s was not created: %v", diag, statErr)
			}
			wantGather := "gather-diagnostics days-of-history 3 no-encrypt\n"
			if tc.encrypt {
				wantGather = "gather-diagnostics days-of-history 3\n"
			}
			gathers, downloads := 0, 0
			for _, c := range rr.calls {
				if c.method == "RunInput" && strings.Contains(c.stdin, wantGather) {
					gathers++
				}
				if c.method == "Run" && len(c.args) > 0 && c.args[0] == "cp" &&
					strings.HasPrefix(c.args[len(c.args)-1], filepath.Join(diag, "gather-configs-pod-host-")) &&
					strings.HasSuffix(c.args[len(c.args)-1], ".tgz") {
					downloads++
				}
			}
			if want := len(tc.wantPods); gathers != want || downloads != want {
				t.Errorf("%d gather script(s) and %d download(s), want %d of each (--days 3, into %s):\n%s",
					gathers, downloads, want, diag, rr.dump())
			}
			if !strings.Contains(errOut, "Diagnostics written to "+diag) {
				t.Errorf("stderr = %q, want it to say where the diagnostics were written", errOut)
			}
			if tc.stderr != "" && !strings.Contains(errOut, tc.stderr) {
				t.Errorf("stderr = %q, want it to carry %q", errOut, tc.stderr)
			}
		})
	}
}

// TestOpsK8sStatusBrokerStopsAtTheFirstFailedRead: `status broker` is a sequence of reads,
// and the two a report cannot go on without end it when they fail. A broker resource that
// cannot be read ends the report before any pod, service or StatefulSet is read, and under
// --detail a pod description that fails ends it before the load balancer is described --
// the report must never show a later section as though the earlier one had worked. (A
// section whose read the cluster REFUSES is drawn with a [SKIP] row instead; internal/k8s's
// TestBrokerReportShowsARefusedSection pins that.)
func TestOpsK8sStatusBrokerStopsAtTheFirstFailedRead(t *testing.T) {
	t.Run("an unreadable broker resource ends the report before any other read", func(t *testing.T) {
		rr := &opRunner{fail: opFailOn("pubsubpluseventbrokers")}
		a := &App{Cfg: writeK8sDeployAllEnv(t, "false"), Platform: config.K8s, Runner: rr}
		_, _, err := opsK8sRun(t, func() error { return opK8sStatusBroker(a, config.Primary) })
		if err == nil || !strings.Contains(err.Error(), `reading the broker resource in namespace "solace"`) {
			t.Fatalf("opK8sStatusBroker err = %v, want the failed broker read named with its namespace", err)
		}
		if len(rr.calls) != 1 {
			t.Errorf("the report went on after its first read failed:\n%s", rr.dump())
		}
	})
	t.Run("--detail does not describe the load balancer after the pod's description failed", func(t *testing.T) {
		rr := &opRunner{fail: opFailOn("describe")}
		a := &App{Cfg: writeK8sDeployAllEnv(t, "true"), Platform: config.K8s, Runner: rr, detail: true}
		_, _, err := opsK8sRun(t, func() error { return opK8sStatusBroker(a, config.Backup) })
		if err == nil || !strings.Contains(err.Error(), `injected failure for "describe"`) {
			t.Fatalf("opK8sStatusBroker err = %v, want the failed description", err)
		}
		last := rr.calls[len(rr.calls)-1]
		if got := strings.Join(last.args, " "); got != "describe pod -n solace dev-broker-pubsubplus-b-0" {
			t.Errorf("the last call = %q, want the failed description of the backup pod", got)
		}
		if n := rr.callCount("describe"); n != 1 {
			t.Errorf("%d describe call(s), want only the one that failed:\n%s", n, rr.dump())
		}
	})
}

// TestOpsK8sStatusOperatorRefusesAClusterCommandTheGuardRejects reaches the executor's own
// copy of the execution guard from the CLI handler. The cluster command is re-checked
// immediately before argv is built, because an executor is handed a Config directly and may
// never have seen config.Load; a privilege-escalation wrapper in kubernetes.command must
// therefore fail the command with the guard's message and issue nothing at all, even over a
// runner that would only have echoed it.
func TestOpsK8sStatusOperatorRefusesAClusterCommandTheGuardRejects(t *testing.T) {
	cfg := writeK8sDeployAllEnv(t, "false")
	cfg.K8s.Command = config.Command{"sudo", "kubectl"}
	var echoed strings.Builder
	a := &App{Cfg: cfg, Platform: config.K8s, Runner: engine.Echo{W: &echoed}}
	_, _, err := opsK8sRun(t, func() error { return opK8sStatusOperator(a) })
	if err == nil || !strings.Contains(err.Error(), "kubernetes.command") || !strings.Contains(err.Error(), "sudo") {
		t.Fatalf("opK8sStatusOperator err = %v, want the guard's refusal naming kubernetes.command and sudo", err)
	}
	if echoed.Len() != 0 {
		t.Errorf("a command the guard refused was issued anyway: %q", echoed.String())
	}
}

// TestOpsK8sLogsPreviousComesAfterTheSharedTokens: --previous is Kubernetes-only, so it is
// appended by the Kubernetes handler rather than by logArgs, which builds only the tokens
// every platform shares. It is added only when asked for, and it lands after the shared
// tokens on the one pod --pod resolved -- the exact argv an operator reading a crashed
// container's last words depends on.
func TestOpsK8sLogsPreviousComesAfterTheSharedTokens(t *testing.T) {
	cases := []struct {
		name     string
		previous bool
		want     []string
	}{
		{"the current container", false, []string{"logs", "-n", "solace", "pod/dev-broker-pubsubplus-b-0", "--tail", "20"}},
		{"--previous adds -p after the shared tokens", true, []string{"logs", "-n", "solace", "pod/dev-broker-pubsubplus-b-0", "--tail", "20", "-p"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rr := &opRunner{}
			a := &App{Cfg: writeK8sDeployAllEnv(t, "true"), Platform: config.K8s, Runner: rr, tail: "20", previous: tc.previous}
			_, _, err := opsK8sRun(t, func() error { return opK8sLogs(a, config.Backup) })
			if err != nil {
				t.Fatalf("opK8sLogs err = %v, want nil", err)
			}
			if len(rr.calls) != 1 || rr.calls[0].method != "Run" || rr.calls[0].name != "kubectl" || !slices.Equal(rr.calls[0].args, tc.want) {
				t.Errorf("calls = \n%s\nwant one kubectl run with %q", rr.dump(), tc.want)
			}
		})
	}
}

// TestOpsK8sOperatorLogsRefusesAnUnparseableSince: the operator's logs take the same shared
// flags as the broker's, and --since is canonicalised through a duration parser so a
// free-form string never becomes an argv token. A value that does not parse is a usage
// error and nothing is run; one that does reaches kubectl in its canonical spelling, aimed
// at the controller Deployment in the OPERATOR's namespace rather than the broker's.
func TestOpsK8sOperatorLogsRefusesAnUnparseableSince(t *testing.T) {
	t.Run("an unparseable --since is refused before anything runs", func(t *testing.T) {
		rr := &opRunner{}
		a := &App{Cfg: writeK8sDeployAllEnv(t, "false"), Platform: config.K8s, Runner: rr, since: "yesterday"}
		_, _, err := opsK8sRun(t, func() error { return opK8sOperatorLogs(a) })
		if err == nil || ExitCode(err) != exitUsage || !strings.Contains(err.Error(), "--since") {
			t.Fatalf("opK8sOperatorLogs err = %v (exit %d), want a usage error naming --since", err, ExitCode(err))
		}
		if len(rr.calls) != 0 {
			t.Errorf("a refused flag must not reach the cluster:\n%s", rr.dump())
		}
	})
	t.Run("a good --since reaches the controller in canonical form", func(t *testing.T) {
		cfg := writeK8sDeployAllEnv(t, "false")
		rr := &opRunner{}
		a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr, since: "5m", tail: "10"}
		_, _, err := opsK8sRun(t, func() error { return opK8sOperatorLogs(a) })
		if err != nil {
			t.Fatalf("opK8sOperatorLogs err = %v, want nil", err)
		}
		want := []string{"logs", "-n", k8s.OperatorNamespace(cfg), "deployment/pubsubplus-eventbroker-operator", "--tail", "10", "--since", "5m0s"}
		if len(rr.calls) != 1 || !slices.Equal(rr.calls[0].args, want) {
			t.Errorf("calls = \n%s\nwant one kubectl run with %q", rr.dump(), want)
		}
	})
}

// TestOpsK8sOperatorStartScalesUpAndWaitsWithoutAsking: starting the operator is additive --
// it only resumes reconciliation -- so unlike stopping it there is nothing to confirm, and a
// run with no terminal and no --no-prompt still goes through. It probes permission to patch
// deployments in the OPERATOR's namespace, scales the controller to one replica, and waits
// for the rollout.
func TestOpsK8sOperatorStartScalesUpAndWaitsWithoutAsking(t *testing.T) {
	cfg := writeK8sDeployAllEnv(t, "true")
	opNS := k8s.OperatorNamespace(cfg)
	rr := &opRunner{}
	a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr}
	opsK8sUnattended(a)
	_, errOut, err := opsK8sRun(t, func() error { return opK8sOperatorStart(a) })
	if err != nil {
		t.Fatalf("opK8sOperatorStart err = %v, want nil", err)
	}
	if strings.Contains(errOut, "[y/N]") || strings.Contains(errOut, "refusing") {
		t.Errorf("starting the operator asked or refused:\n%s", errOut)
	}
	var got []string
	for _, c := range rr.calls {
		got = append(got, strings.Join(c.args, " "))
	}
	if len(got) != 3 {
		t.Fatalf("calls =\n%s\nwant the permission probe, the scale and the rollout wait", rr.dump())
	}
	if want := "auth can-i patch deployments -n " + opNS; got[0] != want {
		t.Errorf("first call = %q, want the permission probe %q", got[0], want)
	}
	if want := "scale deployment pubsubplus-eventbroker-operator -n " + opNS + " --replicas=1"; got[1] != want {
		t.Errorf("second call = %q, want %q", got[1], want)
	}
	if want := "rollout status deployment/pubsubplus-eventbroker-operator -n " + opNS; !strings.HasPrefix(got[2], want) {
		t.Errorf("third call = %q, want it to begin %q", got[2], want)
	}
}

// TestOpsK8sOperatorStopConfirmsBecauseItFreezesEveryNamespace: stopping the controller
// freezes reconciliation for EVERY namespace the operator watches, not just this broker's,
// and nothing else reports it -- so it asks first, through the gate a removal uses, and
// says what it freezes and where. --no-prompt is the one unattended answer. Declined or
// unanswerable, the Deployment is not even probed; confirmed, it is scaled to zero with no
// rollout wait and the operator is told how to start it again.
func TestOpsK8sOperatorStopConfirmsBecauseItFreezesEveryNamespace(t *testing.T) {
	const what = "the operator, freezing reconciliation for every broker it watches"
	const question = "Stop " + what + ` (context "prod-cluster")? [y/N] `
	const refusal = "refusing to stop " + what + ` (context "prod-cluster") without confirmation; pass --no-prompt to proceed`
	cases := []struct {
		name   string
		set    func(*App)
		scaled bool
		asked  bool
		stderr string
	}{
		{"--no-prompt scales it to zero without asking", opsK8sSilenced, true, false, "the operator is stopped"},
		{"a yes at the prompt scales it to zero", opsK8sAnswer("y\n"), true, true, question},
		{"a no at the prompt leaves it running", opsK8sAnswer("n\n"), false, true, question},
		{"no terminal and no --no-prompt leaves it running and names the flag", opsK8sUnattended, false, false, refusal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := writeK8sDeployAllEnv(t, "true")
			opNS := k8s.OperatorNamespace(cfg)
			rr := &opRunner{}
			a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr, kubeContext: "prod-cluster"}
			tc.set(a)
			_, errOut, err := opsK8sRun(t, func() error { return opK8sOperatorStop(a) })
			if err != nil {
				t.Fatalf("opK8sOperatorStop err = %v, want nil", err)
			}
			var acted []string
			for _, c := range rr.calls {
				if !opCanI(c) {
					acted = append(acted, strings.Join(c.args, " "))
				}
			}
			scale := "scale deployment pubsubplus-eventbroker-operator -n " + opNS + " --replicas=0"
			switch {
			case tc.scaled && !slices.Equal(acted, []string{scale}):
				t.Errorf("acted = %q, want exactly %q and no rollout wait", acted, scale)
			case !tc.scaled && len(rr.calls) != 0:
				t.Errorf("a stop that was not confirmed must not reach the cluster:\n%s", rr.dump())
			}
			if !strings.Contains(errOut, tc.stderr) {
				t.Errorf("stderr = %q, want it to carry %q", errOut, tc.stderr)
			}
			if asked := strings.Contains(errOut, "[y/N]"); asked != tc.asked {
				t.Errorf("a question was asked = %v, want %v:\n%s", asked, tc.asked, errOut)
			}
		})
	}
}

// TestOpsK8sOperatorRemoveNeverDeletesCRDsOverLiveBrokers covers the CRD layer of `operator
// remove` when brokers exist in the cluster, which deleting the CRDs would cascade-delete,
// including brokers this env file has never heard of. A broker on the cluster fixes the
// answer, so the question is not asked: with --delete-crd the request goes straight to
// OperatorDelete's refusal, which names every broker and comes before anything is deleted,
// so the operator stays installed; without it the CRDs are kept with a warning that does,
// and the rest of the install is removed. No CRD is ever deleted. A broker listing that
// cannot be read is not permission to delete either, and is said rather than swallowed.
func TestOpsK8sOperatorRemoveNeverDeletesCRDsOverLiveBrokers(t *testing.T) {
	const brokers = `{"items":[` +
		`{"metadata":{"name":"dev-broker","namespace":"solace"}},` +
		`{"metadata":{"name":"billing","namespace":"team-b"}}]}`
	setup := func(t *testing.T, listFails bool) (*App, *opRunner) {
		t.Helper()
		cfg := writeK8sDeployAllEnv(t, "false")
		// The running operator watches only this env file's namespace, so OperatorRelease
		// takes the full-delete path: the install is this env file's to remove.
		deployments := `{"items":[{"metadata":{"name":"pubsubplus-eventbroker-operator","namespace":"` +
			k8s.OperatorNamespace(cfg) + `"},"spec":{"replicas":1,"template":{"spec":{"containers":[` +
			`{"image":"solace/pubsubplus-eventbroker-operator:1.4.2",` +
			`"env":[{"name":"WATCH_NAMESPACE","value":"solace"}]}]}}},"status":{"readyReplicas":1,"replicas":1}}]}`
		rr := &opRunner{output: func(c opCall) []byte {
			if c.method != "Output" || len(c.args) < 2 || c.args[0] != "get" {
				return nil
			}
			switch c.args[1] {
			case "pubsubpluseventbrokers":
				return []byte(brokers)
			case "deployment":
				return []byte(deployments)
			}
			return nil
		}}
		if listFails {
			// The read fails, not the permission probe: this is the listing that answers
			// nothing, which the gate must refuse on its own.
			rr.fail = func(c opCall) error {
				if !opCanI(c) && opArgvMatch(c, "pubsubpluseventbrokers") {
					return errors.New("injected listing failure")
				}
				return nil
			}
		}
		return &App{Cfg: cfg, Platform: config.K8s, Runner: rr}, rr
	}
	crdsDeleted := func(rr *opRunner) bool {
		for _, c := range rr.calls {
			if c.method == "RunInput" && strings.Contains(c.stdin, "kind: CustomResourceDefinition") {
				return true
			}
		}
		return false
	}
	bundleDeleted := func(rr *opRunner) bool {
		for _, c := range rr.calls {
			if c.method == "RunInput" && opArgvMatch(c, "delete") && strings.Contains(c.stdin, "kind: Deployment") {
				return true
			}
		}
		return false
	}

	t.Run("--delete-crd over live brokers goes straight to the refusal and asks nothing", func(t *testing.T) {
		a, rr := setup(t, false)
		a.deleteLayer = true
		opsK8sAnswer("y\n")(a) // enough for the one question that is asked: delete the operator
		_, errOut, err := opsK8sRun(t, func() error { return opK8sOperatorRemove(a) })
		for _, want := range []string{"refusing to delete the operator CRDs", "solace/dev-broker", "team-b/billing"} {
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("opK8sOperatorRemove err = %v, want the refusal naming %q", err, want)
			}
		}
		if strings.Contains(errOut, "Also delete") {
			t.Errorf("the CRD question was asked although a live broker fixes its answer:\n%s", errOut)
		}
		if crdsDeleted(rr) {
			t.Errorf("the CRDs were deleted over live brokers:\n%s", rr.dump())
		}
		if !strings.Contains(err.Error(), "Nothing was deleted: the operator is still installed") {
			t.Errorf("err = %v, want it to say nothing was deleted", err)
		}
		if bundleDeleted(rr) {
			t.Errorf("--delete-crd was refused before the first delete, so the operator must stay installed:\n%s", rr.dump())
		}
	})
	t.Run("without --delete-crd the CRDs are kept and the brokers are named", func(t *testing.T) {
		a, rr := setup(t, false)
		opsK8sAnswer("y\n")(a)
		_, errOut, err := opsK8sRun(t, func() error { return opK8sOperatorRemove(a) })
		if err != nil {
			t.Fatalf("opK8sOperatorRemove err = %v, want nil: keeping the CRDs is the default, not a failure", err)
		}
		for _, want := range []string{"2 broker resource(s) still exist, so the operator CRDs are kept", "solace/dev-broker, team-b/billing"} {
			if !strings.Contains(errOut, want) {
				t.Errorf("stderr = %q, want it to say %q", errOut, want)
			}
		}
		if crdsDeleted(rr) {
			t.Errorf("the CRDs were deleted without --delete-crd:\n%s", rr.dump())
		}
		if !bundleDeleted(rr) {
			t.Errorf("the rest of the operator install must still be removed:\n%s", rr.dump())
		}
	})
	t.Run("--delete-crd with an unreadable broker listing still refuses", func(t *testing.T) {
		a, rr := setup(t, true)
		a.deleteLayer = true
		a.noPrompt = true
		_, errOut, err := opsK8sRun(t, func() error { return opK8sOperatorRemove(a) })
		if err == nil || !strings.Contains(err.Error(), "refusing to delete the operator CRDs") || !strings.Contains(err.Error(), "listing broker custom resources") {
			t.Fatalf("opK8sOperatorRemove err = %v, want the CRD deletion refused for want of a listing", err)
		}
		if crdsDeleted(rr) || bundleDeleted(rr) {
			t.Errorf("nothing may be deleted without being able to see what the CRDs would take:\n%s", rr.dump())
		}
		if !strings.Contains(errOut, "could not list the broker resources across the cluster") {
			t.Errorf("stderr = %q, want the failed pre-check said rather than swallowed", errOut)
		}
	})
}

// TestOpsK8sRemoveBrokerStopsWhenASecretWillNotDelete: `broker remove` deletes the CR, then
// the Secrets this env file built, and only then asks about the namespace -- the one
// question whose answer is meaningful solely once everything this env file owns is gone. A
// Secret that will not delete must therefore end the removal there, with the cause kept: the
// namespace is not listed, not offered and not deleted behind a failure.
func TestOpsK8sRemoveBrokerStopsWhenASecretWillNotDelete(t *testing.T) {
	cause := errors.New("injected secret delete failure")
	rr := &opRunner{fail: func(c opCall) error {
		if c.method == "Run" && len(c.args) > 1 && c.args[0] == "delete" && c.args[1] == "secret" {
			return cause
		}
		return nil
	}}
	a := &App{Cfg: writeK8sDeployAllEnv(t, "false"), Platform: config.K8s, Runner: rr, noPrompt: true}
	_, _, err := opsK8sRun(t, func() error { return opK8sRemoveBroker(a) })
	if !errors.Is(err, cause) {
		t.Fatalf("opK8sRemoveBroker err = %v, want the secret-delete failure carried through", err)
	}
	last := rr.calls[len(rr.calls)-1]
	if len(last.args) < 2 || last.args[0] != "delete" || last.args[1] != "secret" {
		t.Errorf("the last call = %v, want the failed secret delete: nothing may follow it, least of all the namespace question:\n%s", last.args, rr.dump())
	}
}

// TestNamespaceScopedIdentityRunsTheBrokerLifecycle is the goal end to end: an identity bound
// only inside its namespace -- every cluster-wide permission question answered no, every
// cluster-wide or cluster-scoped read Forbidden, its own namespace readable by name --
// deploys, validates, reports on and removes its broker, in the default no-admin-password
// configuration. Each broader read it cannot make becomes a warning or a [SKIP] row and none
// fails the command; it neither applies a Namespace it could not patch nor deletes one it
// may not. Only `broker status --all`, which asks for the whole cluster, is refused.
func TestNamespaceScopedIdentityRunsTheBrokerLifecycle(t *testing.T) {
	newApp := func(t *testing.T) (*App, *opRunner) {
		t.Helper()
		cfg := loadDirect(t, "redundancy:\n  enabled: false\n"+
			"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n"+
			"kubernetes:\n  name: dev-broker\n  namespace: solace\n  updateStrategy: automatedRolling\n"+
			"  storage:\n    class: standard\n    msgNodeSize: 30Gi\n", config.K8s)
		rr := &opRunner{
			fail: func(c opCall) error {
				if !opCanI(c) && (opArgvMatch(c, "--all-namespaces") ||
					(len(c.args) > 1 && c.args[0] == "get" && c.args[1] == "sc")) {
					return errors.New("Error from server (Forbidden)")
				}
				return nil
			},
			output: func(c opCall) []byte {
				switch {
				case opCanI(c) && opArgvMatch(c, "--all-namespaces"):
					return []byte("no\n")
				case len(c.args) > 2 && c.args[0] == "get" && c.args[1] == "namespace":
					return []byte("namespace/solace\n")
				}
				return nil
			},
		}
		a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr}
		opsK8sSilenced(a)
		return a, rr
	}
	for _, tc := range []struct {
		name string
		run  func(*App) error
		set  func(*App)
	}{
		{"broker deploy", opK8sDeploy, func(*App) {}},
		{"broker validate", opK8sValidateBroker, func(*App) {}},
		{"broker status", func(a *App) error { return opK8sStatusBroker(a, config.Primary) }, func(*App) {}},
		{"broker remove --delete-data", opK8sRemoveBroker, func(a *App) { a.deleteLayer = true }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, rr := newApp(t)
			tc.set(a)
			_, _, err := opsK8sRun(t, func() error { return tc.run(a) })
			if err != nil {
				t.Fatalf("%s err = %v, want it to proceed on namespace rights alone:\n%s", tc.name, err, rr.dump())
			}
			for _, c := range rr.calls {
				if c.method == "RunInput" && strings.Contains(c.stdin, "kind: Namespace") {
					t.Errorf("%s applied a Namespace it may not patch:\n%s", tc.name, rr.dump())
				}
				if len(c.args) > 1 && c.args[0] == "delete" && c.args[1] == "namespace" {
					t.Errorf("%s deleted a namespace it may not delete:\n%s", tc.name, rr.dump())
				}
			}
		})
	}
	t.Run("broker status --all", func(t *testing.T) {
		a, _ := newApp(t)
		a.all = true
		_, _, err := opsK8sRun(t, func() error { return opK8sStatusBroker(a, config.Primary) })
		if err == nil || !strings.Contains(err.Error(), "without --all") {
			t.Errorf("err = %v: --all asks for the whole cluster, so it is refused, pointing at the namespace view", err)
		}
	})
}

// TestOpsK8sDeployRefusesBeforeTheFirstWrite: every permission the deploy's writes use is
// asked about before any write, so an identity that may create Secrets but not the broker
// resource -- the built-in edit role alone -- is refused with nothing applied, rather than
// after its Secrets were written.
func TestOpsK8sDeployRefusesBeforeTheFirstWrite(t *testing.T) {
	rr := &opRunner{output: func(c opCall) []byte {
		if opCanI(c) && opArgvMatch(c, "pubsubpluseventbrokers") {
			return []byte("no\n")
		}
		return nil
	}}
	a := &App{Cfg: writeK8sDeployAllEnv(t, "false"), Platform: config.K8s, Runner: rr, noPrompt: true}
	_, _, err := opsK8sRun(t, func() error { return opK8sDeploy(a) })
	if err == nil || !strings.Contains(err.Error(), "not allowed to create pubsubpluseventbrokers") {
		t.Fatalf("opK8sDeploy err = %v, want the refused broker resource named", err)
	}
	for _, c := range rr.calls {
		if !opCanI(c) {
			t.Errorf("nothing but probes may run before the refusal, got %v:\n%s", c.args, rr.dump())
		}
	}
}

// TestOpsK8sRemoveBrokerRefusesBeforeTheFirstDelete: every permission the removal uses is
// asked about before anything is deleted, so an identity that may delete the broker and its
// data but not its Secrets is refused with all of them still in place -- not after the CR
// and, with --delete-data, the data are gone and the TLS key and admin password are left
// behind for a half-removed broker.
func TestOpsK8sRemoveBrokerRefusesBeforeTheFirstDelete(t *testing.T) {
	rr := &opRunner{output: func(c opCall) []byte {
		if opCanI(c) && opArgvMatch(c, "secrets") {
			return []byte("no\n")
		}
		return nil
	}}
	a := &App{Cfg: writeK8sDeployAllEnv(t, "false"), Platform: config.K8s, Runner: rr, noPrompt: true, deleteLayer: true}
	_, _, err := opsK8sRun(t, func() error { return opK8sRemoveBroker(a) })
	if err == nil || !strings.Contains(err.Error(), "not allowed to delete secrets") {
		t.Fatalf("opK8sRemoveBroker err = %v, want the refused Secret delete named", err)
	}
	for _, c := range rr.calls {
		if !opCanI(c) {
			t.Errorf("nothing but probes may run before the refusal, got %v:\n%s", c.args, rr.dump())
		}
	}
}

// TestOpsK8sRemoveNamespaceIfEmptyKeepsANamespaceItMayNotDelete: an empty namespace is
// offered only to an identity that may delete it. The namespace-scoped identity never may,
// so even under --no-prompt it is kept with a warning naming the command a cluster admin
// can run, nothing is asked, and the removal it ends still succeeds.
func TestOpsK8sRemoveNamespaceIfEmptyKeepsANamespaceItMayNotDelete(t *testing.T) {
	rr := &opRunner{output: func(c opCall) []byte {
		if opCanI(c) && opArgvMatch(c, "namespaces") {
			return []byte("no\n")
		}
		return nil
	}}
	a := &App{Cfg: writeK8sDeployAllEnv(t, "false"), Platform: config.K8s, Runner: rr}
	opsK8sSilenced(a)
	_, errOut, err := opsK8sRun(t, func() error { return removeNamespaceIfEmpty(a, k8sCluster(a)) })
	if err != nil {
		t.Fatalf("removeNamespaceIfEmpty err = %v, want nil: keeping the namespace is not a failure", err)
	}
	for _, c := range rr.calls {
		if len(c.args) > 1 && c.args[0] == "delete" && c.args[1] == "namespace" {
			t.Errorf("a namespace this identity may not delete was deleted:\n%s", rr.dump())
		}
	}
	for _, want := range []string{"is empty but was kept", "kubectl delete namespace solace"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr = %q, want it to carry %q", errOut, want)
		}
	}
	if strings.Contains(errOut, "Delete the now-empty namespace") {
		t.Errorf("a deletion this identity cannot carry out must not be offered:\n%s", errOut)
	}
}

// TestOpsK8sRemoveNamespaceIfEmptyKeepsTheNamespaceInDoubt pins when the namespace is
// offered for deletion, which is only when nothing else is in it. A Kubernetes namespace is
// never removed and a listing that cannot run keeps the namespace -- both on every path,
// --no-prompt included, because a namespace wrongly reported empty is cascade-deleted while
// one wrongly reported occupied merely stays. An empty one is still a question: a no or an
// unanswerable prompt keeps it, and only a yes deletes it.
func TestOpsK8sRemoveNamespaceIfEmptyKeepsTheNamespaceInDoubt(t *testing.T) {
	cases := []struct {
		name      string
		namespace string
		listFails bool
		set       func(*App)
		deleted   bool
		noCalls   bool
		stderr    string
	}{
		{"a Kubernetes namespace is never removed, even under --no-prompt", "kube-system", false, opsK8sSilenced, false, true,
			`namespace "kube-system" is a Kubernetes namespace and is never removed by this tool`},
		{"a listing that cannot run keeps the namespace, even under --no-prompt", "solace", true, opsK8sSilenced, false, false,
			"could not list what is left in namespace"},
		{"an empty namespace the operator declines stays", "solace", false, opsK8sAnswer("n\n"), false, false,
			"Delete the now-empty namespace solace? [y/N] "},
		{"an empty namespace with no terminal and no --no-prompt stays and names the flag", "solace", false, opsK8sUnattended, false, false,
			"refusing to delete the now-empty namespace solace without confirmation; pass --no-prompt to proceed"},
		{"an empty namespace the operator accepts is deleted", "solace", false, opsK8sAnswer("y\n"), true, false,
			"Delete the now-empty namespace solace? [y/N] "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := writeK8sDeployAllEnv(t, "false")
			cfg.K8s.Namespace = tc.namespace
			rr := &opRunner{}
			if tc.listFails {
				rr.fail = func(c opCall) error {
					if c.method == "Output" && opArgvMatch(c, "persistentvolumeclaims,secrets,configmaps") {
						return errors.New("injected listing failure")
					}
					return nil
				}
			}
			a := &App{Cfg: cfg, Platform: config.K8s, Runner: rr}
			tc.set(a)
			_, errOut, err := opsK8sRun(t, func() error { return removeNamespaceIfEmpty(a, k8sCluster(a)) })
			if err != nil {
				t.Fatalf("removeNamespaceIfEmpty err = %v, want nil: keeping the namespace is not a failure", err)
			}
			deleted := false
			for _, c := range rr.calls {
				if len(c.args) > 1 && c.args[0] == "delete" && c.args[1] == "namespace" {
					deleted = true
				}
			}
			if deleted != tc.deleted {
				t.Errorf("namespace deleted = %v, want %v:\n%s", deleted, tc.deleted, rr.dump())
			}
			if tc.noCalls && len(rr.calls) != 0 {
				t.Errorf("a protected namespace must not even be listed:\n%s", rr.dump())
			}
			if !strings.Contains(errOut, tc.stderr) {
				t.Errorf("stderr = %q, want it to carry %q", errOut, tc.stderr)
			}
		})
	}
}

// TestOpsK8sClusterConfirmIsTheDowngradeGateWithNoWayAround covers the one question the
// Kubernetes package asks, as k8sCluster wires it: whether to install an operator OLDER than
// the cluster-scoped one already running, which rolls back every broker it watches. The
// wiring must reach confirmDowngrade and not the removal gate, so --no-prompt is no side
// door -- it neither answers the question nor lets an unattended run through -- and only a
// yes typed by a human proceeds. Nobody answering at all refuses and says to re-run
// interactively, not to pass a flag this command does not have.
func TestOpsK8sClusterConfirmIsTheDowngradeGateWithNoWayAround(t *testing.T) {
	const question = "Install an operator older than the one running?"
	const refusal = "refusing to downgrade the operator without confirmation; re-run interactively to confirm the downgrade"
	cases := []struct {
		name    string
		app     *App
		want    bool
		stderr  string
		refused bool
	}{
		{"a yes at the prompt lets the downgrade through",
			&App{Interactive: func() bool { return true }, PromptIn: strings.NewReader("y\n")}, true, question + " [y/N] ", false},
		{"a no at the prompt declines without a refusal line",
			&App{Interactive: func() bool { return true }, PromptIn: strings.NewReader("n\n")}, false, question + " [y/N] ", false},
		{"--no-prompt does not answer a question that is asked",
			&App{noPrompt: true, Interactive: func() bool { return true }, PromptIn: strings.NewReader("n\n")}, false, question + " [y/N] ", false},
		{"--no-prompt does not let an unattended run through",
			&App{noPrompt: true, Interactive: func() bool { return false }}, false, refusal, true},
		{"nobody answering refuses and says to re-run interactively",
			&App{Interactive: func() bool { return true }, PromptIn: strings.NewReader("")}, false, refusal, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := k8sCluster(tc.app)
			if c.Confirm == nil {
				t.Fatal("k8sCluster did not wire Cluster.Confirm: the operator downgrade would be declined for everyone")
			}
			var got bool
			errOut := captureStderr(t, func() { got = c.Confirm(question) })
			if got != tc.want {
				t.Errorf("Confirm = %v, want %v", got, tc.want)
			}
			if !strings.Contains(errOut, tc.stderr) {
				t.Errorf("stderr = %q, want it to carry %q", errOut, tc.stderr)
			}
			if refused := strings.Contains(errOut, "refusing"); refused != tc.refused {
				t.Errorf("a refusal line was printed = %v, want %v:\n%s", refused, tc.refused, errOut)
			}
		})
	}
}

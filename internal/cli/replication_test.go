package cli

import (
	"strings"
	"testing"

	"solace/internal/broker"
	"solace/internal/config"
)

// replication_test.go covers the wiring internal/cli owns for data replication: which
// mechanism a site is reached over, and where the MATE's password comes from. The
// channels themselves are tested in internal/broker and internal/k8s; what is at stake
// here is the dispatch, which is the one place the two can be confused for each other.

// replEnv is a DR pair whose two sites are deliberately reached DIFFERENTLY -- one over
// its own cluster CLI, one over SEMP -- so a dispatch that ignored `via` and always took
// one branch would fail on the other site rather than pass on both.
const replEnv = "redundancy:\n  enabled: false\n" +
	"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n" +
	"semp:\n  adminPass: " + smokeAdminPass + "\n" +
	"kubernetes:\n  name: dev-broker\n  namespace: solace\n" +
	"  storage:\n    class: standard\n    msgNodeSize: 30Gi\n" +
	"replication:\n" +
	"  sites:\n" +
	"    - virtualRouterName: \"v:sg1\"\n" +
	"      routerNames: [sg1]\n" +
	"      endpoints:\n" +
	"        - { host: 10.160.132.1, port: 55443, transport: ssl }\n" +
	"      via:\n" +
	"        kubernetes:\n" +
	"          command: kubectl\n" +
	"          namespace: solace-sg\n" +
	"          name: sg-broker\n" +
	"    - virtualRouterName: \"v:dr1\"\n" +
	"      routerNames: [dr1]\n" +
	"      endpoints:\n" +
	"        - { host: 10.150.132.1, port: 55443, transport: ssl }\n" +
	"      via:\n" +
	"        semp:\n" +
	"          host: 10.150.132.1\n" +
	"          port: 1943\n" +
	"          tls: true\n" +
	"          pass: dr-admin-pw\n" +
	"  vpns:\n" +
	"    - { name: ORDERS, activeAt: \"v:sg1\" }\n"

// replApp builds an App over replEnv with a recording runner, the shape every test here
// needs: no cluster, and every kubectl call captured.
func replApp(t *testing.T, r *opRunner) *App {
	t.Helper()
	return &App{Cfg: loadDirect(t, replEnv, config.K8s), Platform: config.K8s, Runner: r}
}

func siteNamed(t *testing.T, a *App, vrn string) config.ReplSite {
	t.Helper()
	for _, s := range a.Cfg.Replication.Sites {
		if s.VirtualRouterName == vrn {
			return s
		}
	}
	t.Fatalf("fixture has no site %q", vrn)
	return config.ReplSite{}
}

// TestMateChannelPicksTheMechanismTheSiteDeclares is the dispatch itself. `via` holds
// exactly one child and the key present IS the mechanism, so there is nothing to read
// but the block -- and a site driven over the wrong one would report on a broker nobody
// asked about.
func TestMateChannelPicksTheMechanismTheSiteDeclares(t *testing.T) {
	a := replApp(t, &opRunner{})

	kube, err := mateChannel(a, k8sOps(a), siteNamed(t, a, "v:sg1"))
	if err != nil {
		t.Fatalf("kubernetes site: %v", err)
	}
	// The description names the SITE's namespace and the pod built from the SITE's
	// broker name -- neither of which is this env file's (solace/dev-broker).
	if got := kube.Describe(); !strings.Contains(got, "solace-sg") || !strings.Contains(got, "sg-broker") {
		t.Errorf("kubernetes channel describes itself as %q, want the SITE's namespace and broker name", got)
	}

	semp, err := mateChannel(a, k8sOps(a), siteNamed(t, a, "v:dr1"))
	if err != nil {
		t.Fatalf("semp site: %v", err)
	}
	if got := semp.Describe(); !strings.Contains(got, "SEMP") || !strings.Contains(got, "10.150.132.1:1943") {
		t.Errorf("semp channel describes itself as %q, want the declared SEMP endpoint", got)
	}
}

// TestMateChannelRefusesASiteWithNoVia keeps `via` optional at LOAD and mandatory HERE:
// an env file that only ever runs the local-only `configure` command must still load,
// but a switch has to write the role at both sites, so an unreachable mate is refused
// before anything is touched.
func TestMateChannelRefusesASiteWithNoVia(t *testing.T) {
	a := replApp(t, &opRunner{})
	bare := siteNamed(t, a, "v:dr1")
	bare.Via = config.ReplVia{}

	_, err := mateChannel(a, k8sOps(a), bare)
	if err == nil {
		t.Fatal("a site with no via: block cannot be reached and must be refused")
	}
	if !strings.Contains(err.Error(), "v:dr1") {
		t.Errorf("error %q does not name the site it is about", err)
	}
}

// TestMateSEMPPasswordUsesTheLiteral proves the declared password is used as-is and that
// resolving it costs no cluster round trip.
func TestMateSEMPPasswordUsesTheLiteral(t *testing.T) {
	r := &opRunner{}
	a := replApp(t, r)

	got, err := mateSEMPPassword(a, siteNamed(t, a, "v:dr1"))
	if err != nil {
		t.Fatalf("mateSEMPPassword: %v", err)
	}
	if got != "dr-admin-pw" {
		t.Errorf("password = %q, want the site's declared value", got)
	}
	if len(r.calls) != 0 {
		t.Errorf("a literal password needs no command, but %d ran: %+v", len(r.calls), r.calls)
	}
}

// TestMateSEMPPasswordIsTheMatesNotThisBrokers is the misconfiguration this whole
// resolution exists to prevent. A DR site is a DIFFERENT broker: sending this
// deployment's semp.adminPass would fail at best, and if the two happened to match it
// would hide the mistake until the day they diverged.
func TestMateSEMPPasswordIsTheMatesNotThisBrokers(t *testing.T) {
	a := replApp(t, &opRunner{})
	got, err := mateSEMPPassword(a, siteNamed(t, a, "v:dr1"))
	if err != nil {
		t.Fatalf("mateSEMPPassword: %v", err)
	}
	if got == a.Cfg.SEMP.AdminPass {
		t.Error("the mate is dialled with ITS OWN password; this deployment's admin password must not be reused")
	}
}

// TestMateSEMPPasswordReadsTheNamedSecret covers the passSecret branch: the value comes
// back decoded, and the read names the MATE's Secret rather than this deployment's
// adminSecret.
func TestMateSEMPPasswordReadsTheNamedSecret(t *testing.T) {
	// "s3cret\n" -- base64 as a Secret stores it, trailing newline and all.
	r := &opRunner{output: func(opCall) []byte { return []byte("czNjcmV0Cg==\n") }}
	a := replApp(t, r)

	site := siteNamed(t, a, "v:dr1")
	site.Via.SEMP.Pass = ""
	site.Via.SEMP.PassSecret = &config.ReplPassSecret{
		Namespace: "solace-dr", Name: "dr-creds", Key: "admin-password"}

	got, err := mateSEMPPassword(a, site)
	if err != nil {
		t.Fatalf("mateSEMPPassword: %v", err)
	}
	if got != "s3cret" {
		t.Errorf("password = %q, want the decoded Secret value", got)
	}
	if len(r.calls) != 1 {
		t.Fatalf("made %d calls, want exactly one secret read: %+v", len(r.calls), r.calls)
	}
	argv := strings.Join(r.calls[0].args, " ")
	if !strings.Contains(argv, "dr-creds") || !strings.Contains(argv, "solace-dr") {
		t.Errorf("read %q, want the MATE's Secret", argv)
	}
	if strings.Contains(argv, a.Cfg.K8s.AdminSecret) {
		t.Errorf("read %q, want the mate's Secret rather than this deployment's adminSecret", argv)
	}
}

// TestMateSEMPPasswordRefusesASiteWithNoPassword names all three ways to supply one.
// There is deliberately no stdin prompt: promptLine echoes, and a no-echo read is a
// dependency decision this feature does not need to take.
func TestMateSEMPPasswordRefusesASiteWithNoPassword(t *testing.T) {
	a := replApp(t, &opRunner{})
	site := siteNamed(t, a, "v:dr1")
	site.Via.SEMP.Pass = ""

	_, err := mateSEMPPassword(a, site)
	if err == nil {
		t.Fatal("a SEMP site with no password cannot be dialled and must be refused")
	}
	for _, want := range []string{"pass", "passEnv", "passSecret", "v:dr1"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q, so it does not say how to fix the file", err, want)
		}
	}
}

// TestConfirmReplicationConfigNamesTheBiggerHammer proves the gate states the bigger of
// its two hazards even though it is asked BEFORE ConfigureReplication has read anything
// from the broker: phase 1, when it turns out to run, stops replication on EVERY VPN
// this broker is replicating, not only the listed ones the per-VPN lines name below the
// warning.
func TestConfirmReplicationConfigNamesTheBiggerHammer(t *testing.T) {
	a := replApp(t, &opRunner{})
	a.Interactive = func() bool { return true }
	a.PromptIn = strings.NewReader("yes\n")
	self := siteNamed(t, a, "v:sg1")
	mate := siteNamed(t, a, "v:dr1")

	var ok bool
	stderr := captureStderr(t, func() {
		ok = confirmReplicationConfig(a, self, mate)
	})
	if !ok {
		t.Fatal("an exact \"yes\" must confirm")
	}
	if !strings.Contains(stderr, "EVERY VPN this broker is replicating") {
		t.Errorf("stderr = %q, want the bigger-hammer warning naming every VPN, not only the listed ones", stderr)
	}
	if !strings.Contains(stderr, `message-vpn "ORDERS" -> active`) {
		t.Errorf("stderr = %q, want the listed VPN's own planned role too", stderr)
	}
}

// TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply is the other side of the
// gate, and the side that matters: this is the command whose phase 1 stops replication on
// every VPN the broker replicates, so what happens when the operator says NO has to be
// pinned, not assumed.
//
// Two properties, and the second is the one nothing covered. A lenient "y" must NOT pass
// -- this takes the exact-`yes` bar `--delete-data` sets, because a habitual y is not the
// answer that should reach it. And a refusal must stop the run BEFORE the broker is
// touched at all, which is judged by the transport recording no call whatsoever rather
// than by the command merely returning nil.
func TestConfirmReplicationConfigIsStrictAndRefusalStopsTheApply(t *testing.T) {
	for _, tc := range []struct {
		name, answer string
		wantOK       bool
	}{
		{"exact yes proceeds", "yes\n", true},
		{"lenient y is refused", "y\n", false},
		{"explicit no is refused", "no\n", false},
		{"an unanswerable prompt is refused", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := &opRunner{}
			a := replApp(t, rr)
			a.Interactive = func() bool { return tc.answer != "" }
			a.PromptIn = strings.NewReader(tc.answer)
			self := siteNamed(t, a, "v:sg1")
			mate := siteNamed(t, a, "v:dr1")

			var ok bool
			captureStderr(t, func() { ok = confirmReplicationConfig(a, self, mate) })
			if ok != tc.wantOK {
				t.Fatalf("confirmReplicationConfig = %v, want %v", ok, tc.wantOK)
			}
			// A gate that answers the question must not itself have reached the broker,
			// whichever way it answered: it runs before ConfigureReplication is called.
			if len(rr.calls) != 0 {
				t.Errorf("the confirmation reached the broker before it was answered:\n%s", rr.dump())
			}
		})
	}
}

// TestReportReplicationConfigMateSkippedSaysNothingWasWritten proves the report never
// claims the mate was applied when phase 1 was skipped -- res.Mate is what would have
// been sent, not what was, and printing it unconditionally would misdescribe a run that
// wrote nothing and stopped no VPN.
func TestReportReplicationConfigMateSkippedSaysNothingWasWritten(t *testing.T) {
	res := &broker.ReplicationConfigResult{
		MateApplied: false,
		Mate:        []string{`replication mate virtual-router-name "v:dr1"`},
		Enabled:     []string{"ORDERS"},
		Status:      "status text",
	}
	stdout := captureStdout(t, func() { reportReplicationConfig(res) })
	if strings.Contains(stdout, "mate addresses applied") {
		t.Errorf("stdout = %q, must not claim the mate was applied when phase 1 was skipped", stdout)
	}
	if !strings.Contains(stdout, "already matched the env file") {
		t.Errorf("stdout = %q, want it to say the mate already matched", stdout)
	}
	if !strings.Contains(stdout, "replication newly enabled") {
		t.Errorf("stdout = %q, want the renamed row for newly-enabled VPNs", stdout)
	}
}

// TestReportReplicationConfigMateAppliedNamesWhatPhase1Stopped proves that, when phase 1
// DID run, the report names every VPN it stopped -- the half of the blast radius
// res.Enabled cannot show, since phase 1's shutdown covers unlisted VPNs too.
func TestReportReplicationConfigMateAppliedNamesWhatPhase1Stopped(t *testing.T) {
	res := &broker.ReplicationConfigResult{
		MateApplied:    true,
		Mate:           []string{`replication mate virtual-router-name "v:dr1"`},
		StoppedForMate: []string{"default", "vpn-01"},
		Status:         "status text",
	}
	stdout := captureStdout(t, func() { reportReplicationConfig(res) })
	if !strings.Contains(stdout, "mate addresses applied") {
		t.Errorf("stdout = %q, want the applied row when phase 1 ran", stdout)
	}
	if !strings.Contains(stdout, "default, vpn-01") {
		t.Errorf("stdout = %q, want the VPNs phase 1 stopped named", stdout)
	}
}

// TestResolveSitesRefusesAFileWithNoReplication stops before the broker is asked
// anything: with no replication: section there is no DR pair, and reading a router name
// to match against an empty list would be a round trip whose answer cannot matter.
func TestResolveSitesRefusesAFileWithNoReplication(t *testing.T) {
	r := &opRunner{}
	cfg := loadDirect(t, "redundancy:\n  enabled: false\n"+
		"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n"+
		"semp:\n  adminPass: "+smokeAdminPass+"\n"+
		"kubernetes:\n  name: dev-broker\n  namespace: solace\n"+
		"  storage:\n    class: standard\n    msgNodeSize: 30Gi\n", config.K8s)
	a := &App{Cfg: cfg, Platform: config.K8s, Runner: r}

	if _, _, err := resolveSites(a, k8sOps(a), config.Primary); err == nil {
		t.Fatal("an env file with no replication: section has no DR pair to act on")
	}
	if len(r.calls) != 0 {
		t.Errorf("nothing should be asked of the broker first, but %d call(s) ran: %+v", len(r.calls), r.calls)
	}
}

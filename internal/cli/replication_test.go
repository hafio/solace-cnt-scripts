package cli

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"solace/internal/broker"
	"solace/internal/config"
)

// replication_test.go covers the wiring internal/cli owns for data replication: which
// mechanism a site is reached over, and where the MATE's password comes from. The
// channels themselves are tested in internal/broker and internal/k8s; what is at stake
// here is the dispatch, which is the one place the two can be confused for each other.
//
// The second half drives the two ops that compose all of it -- the switch and the local
// convergence -- over a scripted pair, because the composition is the part nothing else
// owns: which site resolves as "this one", which write lands where, and what a refusal or
// a declined question leaves behind.

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
	if strings.Contains(argv, a.Cfg.AdminSecretName()) {
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

// --- the two ops over a scripted pair ----------------------------------------------------
//
// opPerformReplication and opConfigureReplication compose what the tests above cover one
// piece at a time: they ask the LOCAL broker which site it is, build the channel to the
// other one, gate on a confirmation, and hand both to internal/broker. What is pinned below
// is the composition -- which site is demoted before which is promoted, what a refusal
// leaves written, and what the operator is asked first. The local broker is the REAL
// channel over an opRunner that answers each script by name; the far site is a double,
// which is what the op's mateChannelFunc seam exists to accept.

// replBanner is what `cli -Apes` echoes ahead of any output on a software broker. It is
// where a channel learns the broker's type, so every canned reply read by type carries it.
const replBanner = "Solace Event Broker Software Enterprise Version 10.26.0.8827\n" +
	"license-software\n"

// replShowMate renders a software broker's `show replication` for a mate called vrn that is
// dialled at addr over ssl: the layout internal/broker's reader parses, with the empty
// plain-text and compressed headings a real report prints.
func replShowMate(vrn, addr string) string {
	return "\nReplication Interface:\n" +
		"Replication Mate:                  " + vrn + "\n" +
		"    Plain Text:\n\n" +
		"    Compressed:\n\n" +
		"    SSL:                           " + addr + "\n\n" +
		"ConfigSync:\n" +
		"  Bridge:\n" +
		"    Admin State:                   Enabled\n"
}

// replUp is a VPN with replication enabled and the given role -- the only shape a switch
// will touch.
func replUp(role broker.ReplRole) broker.VPNRepl {
	return broker.VPNRepl{Admin: broker.AdminEnabled, Role: role}
}

// replVPNTable renders `show message-vpn * replication` for state: the column rule and
// header of a real capture, one row per VPN, every flag but the admin and config state left
// at n/a. It is what the local broker's reads are built from.
func replVPNTable(state map[string]broker.VPNRepl) string {
	names := make([]string, 0, len(state))
	for name := range state {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	fmt.Fprintf(&b, "\n%-32s A C B R Q S M T\n", "Message VPN")
	b.WriteString(strings.Repeat("-", 32) + " - - - - - - - -\n")
	for _, name := range names {
		admin, role := byte('-'), byte('-')
		switch state[name].Admin {
		case broker.AdminEnabled:
			admin = 'U'
		case broker.AdminShutdown:
			admin = 'D'
		}
		switch state[name].Role {
		case broker.RoleActive:
			role = 'A'
		case broker.RoleStandby:
			role = 'S'
		}
		fmt.Fprintf(&b, "%-32s %c %c - - - - N A\n", name, admin, role)
	}
	return b.String()
}

// replMate is a MATE channel the test scripts: every answer comes from its fields and every
// call is recorded. It is what opPerformReplication's mateChannelFunc seam exists to accept
// -- the real channel needs a second cluster or a second broker's SEMP endpoint.
type replMate struct {
	// site is the name Describe gives the channel.
	site string
	// vrn is what ShowReplication reports this broker's own mate to be. The preflight wants
	// it to be the OTHER site's virtualRouterName.
	vrn string
	// state is what ReadRoles reports and what SetRole changes, so a demotion nobody
	// sabotaged reads back as applied.
	state map[string]broker.VPNRepl
	// status is the text Status returns.
	status string
	// notPrimary makes PrimaryActive report a site that is running on its backup.
	notPrimary bool
	// errs fails a whole method, keyed by its name: Preflight, ShowReplication,
	// PrimaryActive, SetRole or Status.
	errs map[string]error
	// onRead runs at the start of every ReadRoles with that call's 1-based number. It can
	// fail the read, or move state to model something else writing between two of the run's
	// reads.
	onRead func(n int) error
	// afterSet runs after every SetRole that took.
	afterSet func(vpn string, role broker.ReplRole)
	// calls records every call in order; a SetRole is recorded as "SetRole <vpn>=<role>".
	calls []string
	// readVPNs records the list each ReadRoles was asked about.
	readVPNs [][]string
}

func (m *replMate) Describe() string {
	return "site " + m.site
}

func (m *replMate) Preflight(context.Context) error {
	m.calls = append(m.calls, "Preflight")
	return m.errs["Preflight"]
}

func (m *replMate) ShowReplication(context.Context) (broker.MateConfig, error) {
	m.calls = append(m.calls, "ShowReplication")
	if err := m.errs["ShowReplication"]; err != nil {
		return broker.MateConfig{}, err
	}
	return broker.MateConfig{VirtualRouterName: m.vrn}, nil
}

func (m *replMate) PrimaryActive(context.Context) (bool, error) {
	m.calls = append(m.calls, "PrimaryActive")
	if err := m.errs["PrimaryActive"]; err != nil {
		return false, err
	}
	return !m.notPrimary, nil
}

func (m *replMate) Status(context.Context) (string, error) {
	m.calls = append(m.calls, "Status")
	if err := m.errs["Status"]; err != nil {
		return "", err
	}
	return m.status, nil
}

func (m *replMate) ReadRoles(_ context.Context, vpns []string) (map[string]broker.VPNRepl, error) {
	m.calls = append(m.calls, "ReadRoles")
	m.readVPNs = append(m.readVPNs, vpns)
	if m.onRead != nil {
		if err := m.onRead(len(m.readVPNs)); err != nil {
			return nil, err
		}
	}
	// Narrowed the way a real channel narrows: a nil list means every VPN, and a listed one
	// the broker does not have is simply absent from the answer.
	listed := func(name string) bool {
		if vpns == nil {
			return true
		}
		for _, v := range vpns {
			if v == name {
				return true
			}
		}
		return false
	}
	out := map[string]broker.VPNRepl{}
	for name, st := range m.state {
		if listed(name) {
			out[name] = st
		}
	}
	return out, nil
}

func (m *replMate) SetRole(_ context.Context, vpn string, role broker.ReplRole) error {
	m.calls = append(m.calls, "SetRole "+vpn+"="+string(role))
	if err := m.errs["SetRole"]; err != nil {
		return err
	}
	m.state[vpn] = broker.VPNRepl{Admin: broker.AdminEnabled, Role: role}
	if m.afterSet != nil {
		m.afterSet(vpn, role)
	}
	return nil
}

// replRig is one run's whole world: an App over replEnv whose local broker -- site v:sg1,
// the one the router name "sg1" resolves to -- is the REAL channel over a recording runner
// that answers each script by name, and a scripted double for the mate, v:dr1.
type replRig struct {
	// app carries the prompt seams: they answer an exact "yes" once, and asked counts how
	// many times the question reached them.
	app *App
	// rr is the runner standing in for the local broker.
	rr *opRunner
	// mate is the far site.
	mate *replMate
	// local is the per-VPN state the local broker reports. A role written through the
	// runner changes it, so a promotion that took reads back as one.
	local map[string]broker.VPNRepl
	// router is the router name the local broker reports.
	router string
	// redundancy is what its `show redundancy` says; the default is a primary holding
	// activity.
	redundancy string
	// mateVRN and mateAddr are the mate the local broker holds, as its `show replication`
	// reports it. The defaults are exactly what replEnv asks of site v:sg1, so the mate
	// reads as already converged.
	mateVRN  string
	mateAddr string
	// rejectSetRole makes the local broker refuse every role write the way `cli -Apes`
	// does while still exiting 0: the refusal is in the transcript and nowhere else.
	rejectSetRole bool
	// asked counts how many times a confirmation reached the prompt seam.
	asked int
	// mateCalls counts how many times the op asked for a channel to the mate and mateAsked
	// is the site it asked for; mateErr, when set, is what the builder fails with.
	mateCalls int
	mateAsked string
	mateErr   error
	// events lists every role write at either site, in the order it was made.
	events []string
}

// replNewRig builds the world for a pair whose ORDERS VPN holds localRole at this broker and
// mateRole at the mate. The file wants ORDERS active at v:sg1, so (standby, active) is a
// switch to perform and (active, standby) is a pair that already matches.
func replNewRig(t *testing.T, localRole, mateRole broker.ReplRole) *replRig {
	t.Helper()
	g := &replRig{
		rr:         &opRunner{},
		router:     "sg1",
		redundancy: healthyShowRD,
		mateVRN:    "v:dr1",
		mateAddr:   "10.150.132.1:55443",
	}
	g.local = map[string]broker.VPNRepl{"ORDERS": replUp(localRole)}
	g.mate = &replMate{
		site:   "v:dr1",
		vrn:    "v:sg1",
		state:  map[string]broker.VPNRepl{"ORDERS": replUp(mateRole)},
		status: "MATE-STATUS-TEXT",
	}
	g.mate.afterSet = func(vpn string, role broker.ReplRole) {
		g.events = append(g.events, "v:dr1 "+vpn+"="+string(role))
	}
	g.rr.output = g.answer
	g.rr.fail = g.observe
	g.app = replApp(t, g.rr)
	g.app.Interactive = func() bool {
		g.asked++
		return true
	}
	g.app.PromptIn = strings.NewReader("yes\n")
	return g
}

// answer is the local broker: each script is answered by the name it was uploaded under,
// which is the one thing about a call that does not depend on how it was dressed in
// kubectl. The names are internal/broker's own constants, spelled out because they are
// unexported there.
func (g *replRig) answer(c opCall) []byte {
	switch c.method {
	case "Output":
		switch {
		case opArgvMatch(c, ".repl-router-name.cli"):
			return []byte("!   Router: \"" + g.router + "\"\n")
		case opArgvMatch(c, ".repl-show-replication.cli"), opArgvMatch(c, ".repl-status.cli"):
			return []byte(replBanner + replShowMate(g.mateVRN, g.mateAddr))
		case opArgvMatch(c, ".repl-show-vpn.cli"):
			return []byte(replBanner + replVPNTable(g.local))
		case opArgvMatch(c, ".repl-show-redundancy.cli"):
			return []byte(replBanner + g.redundancy)
		case opArgvMatch(c, ".repl-set-role.cli"):
			if g.rejectSetRole {
				return []byte(replBanner + "Error: the broker refused this line\n")
			}
			return []byte(replBanner + "sg1> \n")
		}
	case "OutputInput":
		if opArgvMatch(c, "repl-configure-mate") || opArgvMatch(c, "repl-configure-vpns") {
			return []byte(replBanner + "sg1> \n")
		}
	}
	return nil
}

// observe is the runner's fail hook used as an observer, which it may be because it is
// consulted for every call and an error is only propagated when it is non-nil. A role
// script's body rides the upload that precedes the CLI run, so that is where a write is
// seen: it is recorded, and unless the broker is set to refuse it, applied to what the
// local broker reports next.
func (g *replRig) observe(c opCall) error {
	if c.method != "RunInput" || !opArgvMatch(c, ".repl-set-role.cli") {
		return nil
	}
	var vpn, role string
	for _, line := range strings.Split(c.stdin, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "message-vpn "):
			vpn = strings.Trim(strings.TrimPrefix(line, "message-vpn "), `"`)
		case strings.HasPrefix(line, "state "):
			role = strings.TrimPrefix(line, "state ")
		}
	}
	g.events = append(g.events, "v:sg1 "+vpn+"="+role)
	if !g.rejectSetRole {
		g.local[vpn] = replUp(broker.ReplRole(role))
	}
	return nil
}

// mateFor is the mateChannelFunc the switch is handed: it records who it was asked to reach
// and returns the double, or the configured failure.
func (g *replRig) mateFor(_ *App, _ *broker.Ops, site config.ReplSite) (broker.MateChannel, error) {
	g.mateCalls++
	g.mateAsked = site.VirtualRouterName
	if g.mateErr != nil {
		return nil, g.mateErr
	}
	return g.mate, nil
}

// perform runs the switch as `broker perform data-replication` does, capturing both streams.
func (g *replRig) perform(t *testing.T) (stdout, stderr string, err error) {
	t.Helper()
	stdout = captureStdout(t, func() {
		stderr = captureStderr(t, func() {
			err = opPerformReplication(g.app, k8sOps(g.app), config.Primary, g.mateFor)
		})
	})
	return stdout, stderr, err
}

// configure runs the local convergence as `broker configure data-replication` does.
func (g *replRig) configure(t *testing.T) (stdout, stderr string, err error) {
	t.Helper()
	stdout = captureStdout(t, func() {
		stderr = captureStderr(t, func() {
			err = opConfigureReplication(g.app, k8sOps(g.app), config.Primary)
		})
	})
	return stdout, stderr, err
}

// wantNoWrites fails the test if a role was written anywhere: at the mate, through the
// local runner, or recorded by the observer. Three places, because a refusal that is only
// checked in one of them proves nothing about the other two.
func (g *replRig) wantNoWrites(t *testing.T) {
	t.Helper()
	if len(g.events) != 0 {
		t.Errorf("a role was written although the run refused: %v", g.events)
	}
	for _, c := range g.mate.calls {
		if strings.HasPrefix(c, "SetRole") {
			t.Errorf("the mate received %q although the run refused", c)
		}
	}
	if g.rr.hasCall(".repl-set-role.cli") {
		t.Errorf("the local broker was sent a role script although the run refused:\n%s", g.rr.dump())
	}
}

// wantOnlyTheNameRead fails the test unless the local broker was asked nothing but its own
// router name: the run stopped before it read any replication state, let alone wrote some.
func (g *replRig) wantOnlyTheNameRead(t *testing.T) {
	t.Helper()
	if len(g.rr.calls) == 0 {
		t.Error("the broker was never asked its router name")
	}
	for _, c := range g.rr.calls {
		if !opArgvMatch(c, ".repl-router-name.cli") {
			t.Errorf("the run went on to touch the broker after the router-name read:\n%s", g.rr.dump())
			return
		}
	}
}

// body is the script a RunCLI call carried on stdin, found by the script name its skeleton
// names, or "" when no such call was made.
func (g *replRig) body(script string) string {
	for _, c := range g.rr.calls {
		if c.method == "OutputInput" && opArgvMatch(c, script) {
			return c.stdin
		}
	}
	return ""
}

// TestResolveSitesIsDecidedByWhatTheBrokerReports is the property the byte-identical block
// rests on: the file never says which site a broker IS, the broker does. The same env file
// run at the DR broker therefore resolves to the opposite pair, and the node asked is the one
// the role names -- a wrong guess here points a switchover at the wrong broker.
func TestResolveSitesIsDecidedByWhatTheBrokerReports(t *testing.T) {
	for _, tc := range []struct {
		router   string
		wantSelf string
		wantMate string
	}{
		{"sg1", "v:sg1", "v:dr1"},
		{"dr1", "v:dr1", "v:sg1"},
	} {
		t.Run(tc.router, func(t *testing.T) {
			rr := &opRunner{}
			rr.output = func(c opCall) []byte {
				if opArgvMatch(c, ".repl-router-name.cli") {
					return []byte("!   Router: \"" + tc.router + "\"\n")
				}
				return nil
			}
			a := replApp(t, rr)

			var self, mate config.ReplSite
			var err error
			stderr := captureStderr(t, func() {
				self, mate, err = resolveSites(a, k8sOps(a), config.Backup)
			})
			if err != nil {
				t.Fatalf("resolveSites: %v", err)
			}
			if self.VirtualRouterName != tc.wantSelf || mate.VirtualRouterName != tc.wantMate {
				t.Errorf("a broker reporting %q resolved to self %q and mate %q, want %q and %q",
					tc.router, self.VirtualRouterName, mate.VirtualRouterName, tc.wantSelf, tc.wantMate)
			}
			if want := fmt.Sprintf("this broker reports router name %q", tc.router); !strings.Contains(stderr, want) {
				t.Errorf("stderr = %q, want the name the broker reported to be narrated: %q", stderr, want)
			}
			if len(rr.calls) == 0 {
				t.Fatal("the broker was never asked its router name")
			}
			for _, c := range rr.calls {
				argv := strings.Join(c.args, " ")
				if !strings.Contains(argv, "dev-broker-pubsubplus-b-0") || strings.Contains(argv, "-p-0") {
					t.Errorf("call %s %s went to the wrong node: the role was backup", c.name, argv)
				}
			}
		})
	}
}

// TestResolveSitesRefusesAnUnreadableOrUnknownRouterName covers the two ways the broker's
// answer can fail to identify it. Neither falls back to a guess: a name that cannot be read
// stops the run with the read's own error, and a name the file does not declare prints what
// was read beside every declared name so the difference is visible.
func TestResolveSitesRefusesAnUnreadableOrUnknownRouterName(t *testing.T) {
	t.Run("a router name that cannot be read", func(t *testing.T) {
		rr := &opRunner{fail: opFailOn(".repl-router-name.cli")}
		a := replApp(t, rr)

		var err error
		captureStderr(t, func() {
			_, _, err = resolveSites(a, k8sOps(a), config.Primary)
		})
		if err == nil {
			t.Fatal("a router name that cannot be read must stop the run, not fall back to a guess")
		}
		if !strings.Contains(err.Error(), "repl-router-name") {
			t.Errorf("error %q does not say which read failed", err)
		}
	})

	t.Run("a name that matches no site", func(t *testing.T) {
		rr := &opRunner{}
		rr.output = func(c opCall) []byte {
			if opArgvMatch(c, ".repl-router-name.cli") {
				return []byte("!   Router: \"elsewhere\"\n")
			}
			return nil
		}
		a := replApp(t, rr)

		var err error
		captureStderr(t, func() {
			_, _, err = resolveSites(a, k8sOps(a), config.Primary)
		})
		if err == nil {
			t.Fatal("a broker that is in neither site's routerNames must be refused, not assigned one")
		}
		for _, want := range []string{`"elsewhere"`, "sg1, dr1"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q does not mention %q, so the difference is not visible", err, want)
			}
		}
	})
}

// TestOpPerformReplicationMovesTheRoleDemoteFirst is the whole switch, end to end through
// the op: ORDERS is active at the mate and the file wants it here, so it must be demoted at
// the MATE and only then promoted at THIS broker. That order is the safety property, and what
// the composition adds to internal/broker's own ordering tests is that the two writes land at
// the right SITES -- the demotion on the channel mateChannelFunc supplied, the promotion on
// the local one -- and that a VPN the file does not list is neither asked about nor touched.
func TestOpPerformReplicationMovesTheRoleDemoteFirst(t *testing.T) {
	g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
	// OTHER is active at BOTH sites. Were the switch to look at it, it would refuse as a
	// split brain; it is not the file's to manage, so it must be left alone.
	g.local["OTHER"] = replUp(broker.RoleActive)
	g.mate.state["OTHER"] = replUp(broker.RoleActive)

	stdout, stderr, err := g.perform(t)
	if err != nil {
		t.Fatalf("opPerformReplication: %v\nstderr:\n%s", err, stderr)
	}

	if got, want := strings.Join(g.events, ", "), "v:dr1 ORDERS=standby, v:sg1 ORDERS=active"; got != want {
		t.Errorf("role writes = %q, want %q: the demotion at the mate, then the promotion here", got, want)
	}
	if g.local["ORDERS"].Role != broker.RoleActive || g.mate.state["ORDERS"].Role != broker.RoleStandby {
		t.Errorf("end state: here %v, mate %v; want ORDERS active here and standby at the mate",
			g.local["ORDERS"], g.mate.state["ORDERS"])
	}
	if g.local["OTHER"].Role != broker.RoleActive || g.mate.state["OTHER"].Role != broker.RoleActive {
		t.Error("a VPN the file does not list was changed")
	}
	if g.mateCalls != 1 || g.mateAsked != "v:dr1" {
		t.Errorf("the mate channel was requested %d time(s) for %q, want once, for the MATE v:dr1",
			g.mateCalls, g.mateAsked)
	}
	if len(g.mate.readVPNs) == 0 {
		t.Error("the mate's roles were never read")
	}
	for i, vpns := range g.mate.readVPNs {
		if strings.Join(vpns, ",") != "ORDERS" {
			t.Errorf("read #%d asked about %v, want exactly the listed VPN: a nil list reaches every VPN on the broker",
				i+1, vpns)
		}
	}

	for _, want := range []string{
		`this broker reports router name "sg1"`,
		"Move the replication role for 1 message-VPN(s): ORDERS?",
		`message-vpn "ORDERS" will be demoted, confirmed, then promoted`,
	} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr = %q, want it to contain %q", stderr, want)
		}
	}
	demoted := strings.Index(stderr, `demoting message-vpn "ORDERS" at v:dr1`)
	promoted := strings.Index(stderr, `promoting message-vpn "ORDERS" at v:sg1`)
	if demoted < 0 || promoted < 0 || demoted > promoted {
		t.Errorf("stderr = %q, want the demotion narrated at v:dr1 before the promotion at v:sg1", stderr)
	}

	dr := strings.Index(stdout, "Replication status: v:dr1")
	sg := strings.Index(stdout, "Replication status: v:sg1")
	if dr < 0 || sg < 0 || dr > sg || !strings.Contains(stdout, "MATE-STATUS-TEXT") {
		t.Errorf("stdout = %q, want both sites' own status pages, in site-name order", stdout)
	}
}

// TestOpPerformReplicationAsksNothingWhenThePairAlreadyMatches is the no-op: a pair already
// where the file says leaves nothing to move, so the run names the VPNs it found already
// correct, prints both sites' status and exits WITHOUT the exact-`yes` question. Asking a
// destructive question about a run that destroys nothing teaches an operator to answer it
// without reading, which is the opposite of what the gate is for.
func TestOpPerformReplicationAsksNothingWhenThePairAlreadyMatches(t *testing.T) {
	g := replNewRig(t, broker.RoleActive, broker.RoleStandby)

	stdout, stderr, err := g.perform(t)
	if err != nil {
		t.Fatalf("opPerformReplication: %v", err)
	}
	if g.asked != 0 {
		t.Errorf("the confirmation was asked %d time(s) about a run that moves nothing", g.asked)
	}
	g.wantNoWrites(t)
	for _, want := range []string{
		"already correct at both sites: ORDERS",
		"nothing to move; the pair already matches the env file",
	} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr = %q, want it to say %q", stderr, want)
		}
	}
	for _, site := range []string{"v:dr1", "v:sg1"} {
		if !strings.Contains(stdout, "Replication status: "+site) {
			t.Errorf("stdout = %q, want %s's status page even though nothing moved", stdout, site)
		}
	}
}

// TestOpPerformReplicationDeclinedConfirmationWritesNothing proves the op honours the gate
// it asks: anything but an exact `yes` returns without a single role being written at either
// site. A lenient `y` is the case that matters -- this question interrupts message flow, so a
// habitual y must not reach it -- and a stdin that ends before any reply is a decision nobody
// made, which is refused and says how to proceed unattended.
func TestOpPerformReplicationDeclinedConfirmationWritesNothing(t *testing.T) {
	for _, tc := range []struct {
		name   string
		answer string
	}{
		{"a lenient y is not the exact yes", "y\n"},
		{"an explicit no", "no\n"},
		{"stdin ends unanswered", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
			g.app.PromptIn = strings.NewReader(tc.answer)

			stdout, stderr, err := g.perform(t)
			if err != nil {
				t.Fatalf("a declined confirmation is an answer, not a failure: %v", err)
			}
			if g.asked != 1 {
				t.Errorf("the question was asked %d time(s), want once", g.asked)
			}
			g.wantNoWrites(t)
			if strings.Contains(stdout, "Replication status") {
				t.Errorf("stdout = %q, want no status page for a run that moved nothing", stdout)
			}
			if tc.answer == "" && !strings.Contains(stderr, "--no-prompt") {
				t.Errorf("stderr = %q, want the refusal to name --no-prompt", stderr)
			}
		})
	}
}

// TestOpPerformReplicationNoPromptSkipsOnlyTheQuestion pins what --no-prompt is for: it
// answers the confirmation and nothing else. The switch proceeds without reading stdin, but
// every preflight check still runs and still refuses -- an unattended run is not an
// unchecked one.
func TestOpPerformReplicationNoPromptSkipsOnlyTheQuestion(t *testing.T) {
	t.Run("the switch proceeds without asking", func(t *testing.T) {
		g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
		g.app.noPrompt = true
		g.app.PromptIn = strings.NewReader("")

		_, _, err := g.perform(t)
		if err != nil {
			t.Fatalf("opPerformReplication --no-prompt: %v", err)
		}
		if g.asked != 0 {
			t.Errorf("the question was asked %d time(s) under --no-prompt", g.asked)
		}
		if got, want := strings.Join(g.events, ", "), "v:dr1 ORDERS=standby, v:sg1 ORDERS=active"; got != want {
			t.Errorf("role writes = %q, want %q", got, want)
		}
	})

	t.Run("the preflight still refuses", func(t *testing.T) {
		g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
		g.app.noPrompt = true
		g.mate.notPrimary = true

		_, _, err := g.perform(t)
		if err == nil || !strings.Contains(err.Error(), "revert-activity") {
			t.Fatalf("err = %v, want the preflight to refuse a site on its backup even under --no-prompt", err)
		}
		g.wantNoWrites(t)
	})
}

// TestOpPerformReplicationStopsWhenTheMateChannelCannotBeBuilt covers a mate this tool has no
// way to reach -- a site with no via: block, say. The builder's own error comes back
// unchanged, because it already says what to fix, and the run stops with the local broker
// asked nothing beyond its own name: no preflight read, no state, no write.
func TestOpPerformReplicationStopsWhenTheMateChannelCannotBeBuilt(t *testing.T) {
	g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
	errNoVia := errors.New(`replication site "v:dr1" has no via: block, so this tool cannot reach it`)
	g.mateErr = errNoVia

	_, _, err := g.perform(t)
	if !errors.Is(err, errNoVia) {
		t.Fatalf("err = %v, want the builder's own error, which says what to fix", err)
	}
	if g.mateAsked != "v:dr1" {
		t.Errorf("the channel was requested for %q, want the MATE v:dr1", g.mateAsked)
	}
	g.wantOnlyTheNameRead(t)
	g.wantNoWrites(t)
	if g.asked != 0 {
		t.Errorf("the confirmation was asked %d time(s) before the mate could even be reached", g.asked)
	}
}

// TestOpPerformReplicationRefusesBeforeTheFirstWrite walks the preflight's refusals through
// the op. Each one has to name what to do, and each leaves the pair exactly as it found it:
// nothing written at either site, no confirmation asked, and not even a VPN's state read,
// because the preflight is a gate in front of the work rather than a step inside it.
func TestOpPerformReplicationRefusesBeforeTheFirstWrite(t *testing.T) {
	refused := func(t *testing.T, g *replRig, want ...string) {
		t.Helper()
		_, _, err := g.perform(t)
		if err == nil {
			t.Fatal("a pair that fails the preflight must not be switched")
		}
		for _, w := range want {
			if !strings.Contains(err.Error(), w) {
				t.Errorf("error %q does not mention %q, so it does not say what to do", err, w)
			}
		}
		g.wantNoWrites(t)
		if g.asked != 0 {
			t.Errorf("the confirmation was asked %d time(s) after the preflight had refused", g.asked)
		}
		if len(g.mate.readVPNs) != 0 || g.rr.hasCall(".repl-show-vpn.cli") {
			t.Errorf("VPN state was read after the preflight had refused:\n%s", g.rr.dump())
		}
	}

	t.Run("the mate is running on its backup", func(t *testing.T) {
		g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
		g.mate.notPrimary = true
		refused(t, g, "v:dr1", "revert-activity")
	})

	t.Run("this site is running on its backup", func(t *testing.T) {
		g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
		g.redundancy = standbyShowRD
		refused(t, g, "v:sg1", "revert-activity")
	})

	t.Run("the mate is unreachable", func(t *testing.T) {
		g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
		g.mate.errs = map[string]error{"Preflight": errors.New("cannot reach site v:dr1 over semp: connection refused")}
		refused(t, g, "connection refused")
	})

	t.Run("the brokers and the file disagree about which pair this is", func(t *testing.T) {
		g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
		g.mate.vrn = "v:somewhere-else"
		refused(t, g, "v:somewhere-else", "configure data-replication")
	})
}

// TestOpPerformReplicationRefusesAnUnusablePlan covers what the op does with the state it
// reads. A VPN active at both sites, missing, or shut down is a state this tool did not
// create and will not repair mid-switch, and a site that cannot be read cannot be planned
// against: each is refused with the pair untouched and nothing asked.
func TestOpPerformReplicationRefusesAnUnusablePlan(t *testing.T) {
	refused := func(t *testing.T, g *replRig, want ...string) {
		t.Helper()
		_, _, err := g.perform(t)
		if err == nil {
			t.Fatal("a pair that cannot be planned against must not be switched")
		}
		for _, w := range want {
			if !strings.Contains(err.Error(), w) {
				t.Errorf("error %q does not mention %q, so it does not say what to do", err, w)
			}
		}
		g.wantNoWrites(t)
		if g.asked != 0 {
			t.Errorf("the confirmation was asked %d time(s) about a plan that was refused", g.asked)
		}
	}

	t.Run("a VPN active at both sites", func(t *testing.T) {
		g := replNewRig(t, broker.RoleActive, broker.RoleActive)
		refused(t, g, "ORDERS", "BOTH", "v:sg1", "v:dr1")
	})

	t.Run("a listed VPN missing at the mate", func(t *testing.T) {
		g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
		delete(g.mate.state, "ORDERS")
		refused(t, g, "do not exist", "ORDERS at v:dr1", "configure data-replication")
	})

	t.Run("replication shut down at the mate", func(t *testing.T) {
		g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
		g.mate.state["ORDERS"] = broker.VPNRepl{Admin: broker.AdminShutdown, Role: broker.RoleStandby}
		refused(t, g, "shut down", "ORDERS at v:dr1")
	})

	t.Run("a site whose roles cannot be read", func(t *testing.T) {
		g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
		g.mate.onRead = func(int) error { return errors.New("connection reset by peer") }
		refused(t, g, "connection reset by peer")
	})
}

// TestOpPerformReplicationStopsWhenTheSwitchFailsPartway covers the three ways the ordered
// write can break after the preflight passed. In each, nothing is promoted on stale or
// unconfirmed evidence, and the direction of the failure is the documented one: a VPN left
// standby at BOTH sites is unavailable, not corrupted.
func TestOpPerformReplicationStopsWhenTheSwitchFailsPartway(t *testing.T) {
	t.Run("the mate refuses the demotion", func(t *testing.T) {
		g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
		errDemote := errors.New("the mate refused the demotion")
		g.mate.errs = map[string]error{"SetRole": errDemote}

		_, _, err := g.perform(t)
		if !errors.Is(err, errDemote) {
			t.Fatalf("err = %v, want the demotion's own error", err)
		}
		if g.rr.hasCall(".repl-set-role.cli") {
			t.Errorf("this broker was promoted after the demotion failed:\n%s", g.rr.dump())
		}
		if len(g.events) != 0 {
			t.Errorf("role writes = %v, want none: the one attempted write was refused", g.events)
		}
	})

	t.Run("the mate's state moves between the confirmation and the promotion", func(t *testing.T) {
		g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
		// The third read the mate answers is the one immediately before promoting; something
		// else has promoted ORDERS there again by then.
		g.mate.onRead = func(n int) error {
			if n == 3 {
				g.mate.state["ORDERS"] = replUp(broker.RoleActive)
			}
			return nil
		}

		_, _, err := g.perform(t)
		if err == nil || !strings.Contains(err.Error(), "since the demotion was confirmed") {
			t.Fatalf("err = %v, want a refusal to promote against state that moved", err)
		}
		if g.rr.hasCall(".repl-set-role.cli") {
			t.Errorf("this broker was promoted against stale state:\n%s", g.rr.dump())
		}
	})

	t.Run("this broker rejects the promotion", func(t *testing.T) {
		g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
		g.rejectSetRole = true

		_, _, err := g.perform(t)
		if err == nil || !strings.Contains(err.Error(), "rejected") || !strings.Contains(err.Error(), "ORDERS") {
			t.Fatalf("err = %v, want the rejected promotion named", err)
		}
		if g.local["ORDERS"].Role != broker.RoleStandby || g.mate.state["ORDERS"].Role != broker.RoleStandby {
			t.Errorf("end state: here %v, mate %v; want ORDERS standby at BOTH sites",
				g.local["ORDERS"], g.mate.state["ORDERS"])
		}
	})
}

// TestOpPerformReplicationFailsWhenAFinalReadDisagreesWithTheFile makes the end state the
// exit code, through the op. Every write takes and every read-back before the promotion
// agrees; it is the FINAL read that finds ORDERS active at the mate again, as if an operator
// at its CLI had promoted it back. That is a failure and not a warning -- the whole point of
// the command is that the end state is known -- and the status pages are printed before the
// verdict so the operator can see what the brokers say.
func TestOpPerformReplicationFailsWhenAFinalReadDisagreesWithTheFile(t *testing.T) {
	g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
	// The fourth read the mate answers is the verification: the plan, the demotion check
	// and the pre-promotion re-read come first.
	g.mate.onRead = func(n int) error {
		if n == 4 {
			g.mate.state["ORDERS"] = replUp(broker.RoleActive)
		}
		return nil
	}

	stdout, _, err := g.perform(t)
	if err == nil {
		t.Fatal("a pair that ends up where the file does not say must fail the run, not warn")
	}
	if want := "ORDERS at v:dr1 reads active, want standby"; !strings.Contains(err.Error(), want) {
		t.Errorf("error %q does not name what disagrees: want %q", err, want)
	}
	if !strings.Contains(stdout, "MATE-STATUS-TEXT") {
		t.Errorf("stdout = %q, want both sites' status printed even though the run failed", stdout)
	}
}

// TestOpPerformReplicationAddressesTheNodeTheRoleNames proves the role handed to the op
// reaches every read and write of the local site. The replication role config-syncs to the
// rest of the HA group, so the op writes to ONE node -- the one --pod names -- and a call
// that strayed to another pod would be a write to a node nobody chose.
func TestOpPerformReplicationAddressesTheNodeTheRoleNames(t *testing.T) {
	g := replNewRig(t, broker.RoleStandby, broker.RoleActive)

	var err error
	captureStdout(t, func() {
		captureStderr(t, func() {
			err = opPerformReplication(g.app, k8sOps(g.app), config.Backup, g.mateFor)
		})
	})
	if err != nil {
		t.Fatalf("opPerformReplication: %v", err)
	}
	if len(g.rr.calls) == 0 {
		t.Fatal("the local broker was never touched")
	}
	for _, c := range g.rr.calls {
		argv := strings.Join(c.args, " ")
		if !strings.Contains(argv, "dev-broker-pubsubplus-b-0") || strings.Contains(argv, "-p-0") {
			t.Errorf("call %s %s did not go to the node the role names", c.name, argv)
		}
	}
}

// TestVerifyReplicationSwitchMakesTheEndStateTheExitCode drives the verification directly,
// because its branches are what decide whether a finished switch is a success. Both sites are
// re-read and printed, and EVERY listed VPN is checked against the file -- ORDERS belongs at
// v:sg1, so a converged pair is active there and standby at v:dr1. The picture is printed
// before the verdict, and a site that cannot be read fails the verdict rather than being
// skipped, since a verdict from half a pair is no verdict.
func TestVerifyReplicationSwitchMakesTheEndStateTheExitCode(t *testing.T) {
	a := replApp(t, &opRunner{})
	vpns := []string{"ORDERS"}
	pair := func(sg, dr broker.ReplRole) (map[string]broker.MateChannel, *replMate, *replMate) {
		s := &replMate{state: map[string]broker.VPNRepl{"ORDERS": replUp(sg)}, status: "SG-STATUS"}
		d := &replMate{state: map[string]broker.VPNRepl{"ORDERS": replUp(dr)}, status: "DR-STATUS"}
		return map[string]broker.MateChannel{"v:sg1": s, "v:dr1": d}, s, d
	}

	t.Run("a converged pair passes and both pictures are printed", func(t *testing.T) {
		chans, _, _ := pair(broker.RoleActive, broker.RoleStandby)

		var err error
		stdout := captureStdout(t, func() {
			err = verifyReplicationSwitch(a, chans, broker.SwitchPlan{}, vpns)
		})
		if err != nil {
			t.Fatalf("a pair where the file says it should be must verify: %v", err)
		}
		for _, want := range []string{"SG-STATUS", "DR-STATUS"} {
			if !strings.Contains(stdout, want) {
				t.Errorf("stdout = %q, want %q printed", stdout, want)
			}
		}
	})

	t.Run("a pair still where it started fails after the picture is printed", func(t *testing.T) {
		chans, _, _ := pair(broker.RoleStandby, broker.RoleActive)

		var err error
		stdout := captureStdout(t, func() {
			err = verifyReplicationSwitch(a, chans, broker.SwitchPlan{}, vpns)
		})
		if err == nil {
			t.Fatal("a VPN not where the file says it should be must fail the run")
		}
		// Sorted, so the same broker state reads the same way twice.
		want := "ORDERS at v:dr1 reads active, want standby; ORDERS at v:sg1 reads standby, want active"
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not list every disagreement as %q", err, want)
		}
		if !strings.Contains(stdout, "SG-STATUS") || !strings.Contains(stdout, "DR-STATUS") {
			t.Errorf("stdout = %q, want the picture printed before the verdict", stdout)
		}
	})

	t.Run("a VPN that vanished during the run is not where the file says", func(t *testing.T) {
		chans, _, d := pair(broker.RoleActive, broker.RoleStandby)
		delete(d.state, "ORDERS")

		var err error
		captureStdout(t, func() {
			err = verifyReplicationSwitch(a, chans, broker.SwitchPlan{}, vpns)
		})
		if err == nil || !strings.Contains(err.Error(), "ORDERS at v:dr1") {
			t.Fatalf("err = %v, want the VPN that is gone named at its site", err)
		}
		if strings.Contains(err.Error(), "ORDERS at v:sg1") {
			t.Errorf("error %q blames the site that is correct", err)
		}
	})

	t.Run("a site that cannot be read fails the verdict and prints nothing", func(t *testing.T) {
		chans, _, d := pair(broker.RoleActive, broker.RoleStandby)
		d.onRead = func(int) error { return errors.New("connection reset by peer") }

		var err error
		stdout := captureStdout(t, func() {
			err = verifyReplicationSwitch(a, chans, broker.SwitchPlan{}, vpns)
		})
		if err == nil || !strings.Contains(err.Error(), "connection reset by peer") {
			t.Fatalf("err = %v, want the read's own error", err)
		}
		if stdout != "" {
			t.Errorf("stdout = %q, want nothing: the pair was not fully read", stdout)
		}
	})

	t.Run("a status page that cannot be read fails the verdict", func(t *testing.T) {
		chans, _, d := pair(broker.RoleActive, broker.RoleStandby)
		d.errs = map[string]error{"Status": errors.New("show replication timed out")}

		var err error
		captureStdout(t, func() {
			err = verifyReplicationSwitch(a, chans, broker.SwitchPlan{}, vpns)
		})
		if err == nil || !strings.Contains(err.Error(), "show replication timed out") {
			t.Fatalf("err = %v, want the status read's own error", err)
		}
	})
}

// TestReportReplicationStatusPrintsEachSiteInNameOrderAndStopsAtTheFirstFailure pins the
// report's two properties. The sites print in name order whatever order the map held them in
// -- a Go map ranges randomly, and two runs against one pair must read the same -- each
// followed by the broker's own text verbatim. And a site whose picture cannot be read stops
// the report there rather than being skipped, so a page is never mistaken for the whole pair.
func TestReportReplicationStatusPrintsEachSiteInNameOrderAndStopsAtTheFirstFailure(t *testing.T) {
	t.Run("name order", func(t *testing.T) {
		sg := &replMate{status: "SG-BODY"}
		dr := &replMate{status: "DR-BODY"}
		chans := map[string]broker.MateChannel{"v:sg1": sg, "v:dr1": dr}

		var err error
		stdout := captureStdout(t, func() {
			err = reportReplicationStatus(chans)
		})
		if err != nil {
			t.Fatalf("reportReplicationStatus: %v", err)
		}
		last := -1
		for _, want := range []string{"Replication status: v:dr1", "DR-BODY", "Replication status: v:sg1", "SG-BODY"} {
			i := strings.Index(stdout, want)
			if i < 0 || i < last {
				t.Fatalf("stdout = %q, want %q after everything printed before it", stdout, want)
			}
			last = i
		}
	})

	t.Run("a site that cannot be read stops the report", func(t *testing.T) {
		errDown := errors.New("show replication timed out")
		sg := &replMate{status: "SG-BODY"}
		dr := &replMate{errs: map[string]error{"Status": errDown}}
		chans := map[string]broker.MateChannel{"v:sg1": sg, "v:dr1": dr}

		var err error
		stdout := captureStdout(t, func() {
			err = reportReplicationStatus(chans)
		})
		if !errors.Is(err, errDown) {
			t.Fatalf("err = %v, want the failing site's own error", err)
		}
		if stdout != "" || len(sg.calls) != 0 {
			t.Errorf("stdout = %q, later site calls = %v; want the report to stop at the first failure",
				stdout, sg.calls)
		}
	})
}

// TestConfirmReplicationSwitchNamesEveryMoveAndIsStrict covers the switch's own gate with
// more than one VPN, which the end-to-end runs cannot show: every move is named, sorted, in
// the narration and in the question, and the bar is the exact `yes` -- a lenient y does not
// pass, and a stdin that ends unanswered is refused naming --no-prompt. The gate runs before
// anything is written, so the broker is never touched whichever way it answers.
func TestConfirmReplicationSwitchNamesEveryMoveAndIsStrict(t *testing.T) {
	both := func(role broker.ReplRole) map[string]broker.VPNRepl {
		return map[string]broker.VPNRepl{"ORDERS": replUp(role), "PAYMENTS": replUp(role)}
	}
	vpns := []config.ReplVPN{{Name: "PAYMENTS", ActiveAt: "v:sg1"}, {Name: "ORDERS", ActiveAt: "v:sg1"}}
	plan, err := broker.BuildSwitchPlan(vpns,
		broker.SiteState{Site: "v:sg1", State: both(broker.RoleStandby)},
		broker.SiteState{Site: "v:dr1", State: both(broker.RoleActive)})
	if err != nil {
		t.Fatalf("BuildSwitchPlan: %v", err)
	}

	for _, tc := range []struct {
		name   string
		answer string
		want   bool
	}{
		{"an exact yes proceeds", "yes\n", true},
		{"a lenient y does not", "y\n", false},
		{"stdin ends unanswered", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := &opRunner{}
			a := replApp(t, rr)
			a.Interactive = func() bool { return true }
			a.PromptIn = strings.NewReader(tc.answer)

			var ok bool
			stderr := captureStderr(t, func() {
				ok = confirmReplicationSwitch(a, plan)
			})
			if ok != tc.want {
				t.Fatalf("confirmReplicationSwitch = %v, want %v", ok, tc.want)
			}
			orders := strings.Index(stderr, `message-vpn "ORDERS" will be demoted, confirmed, then promoted`)
			payments := strings.Index(stderr, `message-vpn "PAYMENTS" will be demoted, confirmed, then promoted`)
			if orders < 0 || payments < 0 || orders > payments {
				t.Errorf("stderr = %q, want both moves named, sorted, before the question", stderr)
			}
			if !strings.Contains(stderr, "2 message-VPN(s): ORDERS, PAYMENTS") {
				t.Errorf("stderr = %q, want the count and the sorted list in the question", stderr)
			}
			if tc.answer == "" && !strings.Contains(stderr, "--no-prompt") {
				t.Errorf("stderr = %q, want the refusal to name --no-prompt", stderr)
			}
			if len(rr.calls) != 0 {
				t.Errorf("the gate reached the broker before it was answered:\n%s", rr.dump())
			}
		})
	}
}

// TestMateChannelRefusesASEMPSiteWithNoPassword is the production builder's half of the
// password rule: a SEMP site whose password cannot be resolved yields no channel at all, and
// says which site and how to fix it, rather than dialling the mate with an empty credential.
// It is a usage error -- the env file is what is wrong -- and costs no cluster round trip.
func TestMateChannelRefusesASEMPSiteWithNoPassword(t *testing.T) {
	r := &opRunner{}
	a := replApp(t, r)
	site := siteNamed(t, a, "v:dr1")
	site.Via.SEMP.Pass = ""

	ch, err := mateChannel(a, k8sOps(a), site)
	if err == nil {
		t.Fatal("a SEMP site with no password cannot be dialled, so no channel to it may be built")
	}
	if ch != nil {
		t.Errorf("a refused site still produced a channel: %s", ch.Describe())
	}
	if !isUsage(err) {
		t.Errorf("error %q is not a usage error, but the env file is what is wrong", err)
	}
	for _, want := range []string{"v:dr1", "passSecret"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q, so it does not say how to fix the file", err, want)
		}
	}
	if len(r.calls) != 0 {
		t.Errorf("refusing a site needs no command, but %d ran: %+v", len(r.calls), r.calls)
	}
}

// TestReportReplicationConfigNamesWhatWasClearedAndShutDown covers the two rows the other
// report tests leave out. The mate addresses a convergence CLEARED are named beside the ones
// it applied -- an operator reading "applied" alone cannot tell that the broker used to dial
// somewhere else -- and every replicating VPN the file does not list is named as shut down,
// which is the part of the blast radius the roles above it cannot show.
func TestReportReplicationConfigNamesWhatWasClearedAndShutDown(t *testing.T) {
	removed := []string{`no replication mate connect-via "10.9.9.9:55443"`, "no replication mate virtual-router-name"}
	applied := []string{`replication mate virtual-router-name "v:dr1"`}
	res := &broker.ReplicationConfigResult{
		MateApplied: true,
		Removed:     removed,
		Mate:        applied,
		Disabled:    []string{"legacy", "scratch"},
		Status:      "status text",
	}

	stdout := captureStdout(t, func() { reportReplicationConfig(res) })
	for _, want := range []string{
		"mate addresses cleared",
		strings.Join(removed, "; "),
		"mate addresses applied",
		"replication SHUT DOWN (not in the file)",
		"legacy, scratch",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
	// Nothing was stopped for the mate and nothing was newly enabled, so neither row exists.
	for _, absent := range []string{"replication stopped to change the mate", "newly enabled"} {
		if strings.Contains(stdout, absent) {
			t.Errorf("stdout = %q, must not carry a %q row for something that did not happen", stdout, absent)
		}
	}
	if strings.Index(stdout, "status text") < strings.Index(stdout, "SHUT DOWN") {
		t.Errorf("stdout = %q, want the broker's own report after the table", stdout)
	}
}

// TestOpConfigureReplicationAppliesTheRoleTheBrokersSiteDerives is the byte-identical block
// at work. The same env file, run at each end, hands the broker the OPPOSITE role for the
// same VPN -- active where the file says it lives, standby at the other site -- because the
// broker decides which site it is, and the mate it renders and compares against is the OTHER
// entry. A mate that already matches means phase 1 is skipped and nothing is stopped, and
// configure never reaches for the mate at all.
func TestOpConfigureReplicationAppliesTheRoleTheBrokersSiteDerives(t *testing.T) {
	for _, tc := range []struct {
		name     string
		router   string
		self     string
		mate     string
		mateAddr string
		role     string
	}{
		{"at the site the file makes active", "sg1", "v:sg1", "v:dr1", "10.150.132.1:55443", "active"},
		{"at the DR site, from the very same file", "dr1", "v:dr1", "v:sg1", "10.160.132.1:55443", "standby"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := replNewRig(t, broker.RoleStandby, broker.RoleStandby)
			g.router, g.mateVRN, g.mateAddr = tc.router, tc.mate, tc.mateAddr

			stdout, stderr, err := g.configure(t)
			if err != nil {
				t.Fatalf("opConfigureReplication: %v\nstderr:\n%s", err, stderr)
			}
			if g.asked != 1 {
				t.Errorf("the confirmation was asked %d time(s), want once", g.asked)
			}
			site := fmt.Sprintf("this broker is site %s; its mate is %s", tc.self, tc.mate)
			planned := fmt.Sprintf("message-vpn %q -> %s", "ORDERS", tc.role)
			for _, want := range []string{site, planned} {
				if !strings.Contains(stderr, want) {
					t.Errorf("stderr = %q, want it to contain %q", stderr, want)
				}
			}

			other := "standby"
			if tc.role == "standby" {
				other = "active"
			}
			body := g.body("repl-configure-vpns")
			if !strings.Contains(body, "state "+tc.role) || strings.Contains(body, "state "+other) {
				t.Errorf("phase 2 body = %q, want ORDERS given role %s and not %s", body, tc.role, other)
			}
			if g.rr.hasCall("repl-configure-mate") {
				t.Error("the mate already matches the file, so phase 1 must not run: it stops replication everywhere")
			}
			if g.mateCalls != 0 || len(g.mate.calls) != 0 || g.rr.hasCall("curl") {
				t.Errorf("configure reached for the mate, which it must never do:\n%s", g.rr.dump())
			}
			for _, want := range []string{"already matched the env file", "Replication Mate:"} {
				if !strings.Contains(stdout, want) {
					t.Errorf("stdout = %q, want it to contain %q", stdout, want)
				}
			}
		})
	}
}

// TestOpConfigureReplicationConvergesTheMateAndReportsWhatItStopped is the disruptive half.
// The broker holds a different mate than the file asks of this site, so phase 1 runs: the old
// address and virtual-router-name are cleared, the OTHER site's are applied -- never this
// broker's own, which is what a swapped (self, mate) pair would render -- and every VPN it
// had to stop is named in the report, listed or not, beside the ones the file does not list.
func TestOpConfigureReplicationConvergesTheMateAndReportsWhatItStopped(t *testing.T) {
	g := replNewRig(t, broker.RoleStandby, broker.RoleStandby)
	g.local["legacy"] = replUp(broker.RoleStandby)
	g.mateVRN, g.mateAddr = "v:old-dr", "10.9.9.9:55443"

	stdout, stderr, err := g.configure(t)
	if err != nil {
		t.Fatalf("opConfigureReplication: %v\nstderr:\n%s", err, stderr)
	}

	mateBody := g.body("repl-configure-mate")
	for _, want := range []string{
		`no replication mate connect-via "10.9.9.9:55443"`,
		"no replication mate virtual-router-name",
		`replication mate connect-via "10.150.132.1:55443" "ssl"`,
		`replication mate virtual-router-name "v:dr1"`,
	} {
		if !strings.Contains(mateBody, want) {
			t.Errorf("phase 1 body is missing %q:\n%s", want, mateBody)
		}
	}
	if strings.Contains(mateBody, "10.160.132.1") {
		t.Errorf("phase 1 body carries this broker's OWN address, so it was rendered for the wrong mate:\n%s", mateBody)
	}
	if g.body("repl-configure-vpns") == "" {
		t.Error("phase 2 never ran, so the listed VPN was never given its role back")
	}
	for _, want := range []string{
		"Replication configuration applied",
		"mate addresses cleared",
		"mate addresses applied",
		"replication stopped to change the mate",
		"ORDERS, legacy",
		"replication SHUT DOWN (not in the file)",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
}

// TestOpConfigureReplicationDeclinedLeavesTheBrokerUntouched proves a refusal stops the
// apply before the broker is read, let alone written. The gate is asked after the router name
// -- it needs the site -- and before anything else, so the one read that preceded it is the
// only thing the broker has seen. A decline is an answer and exits clean, prints no report
// for a run that applied nothing, and an unanswered question names --no-prompt.
func TestOpConfigureReplicationDeclinedLeavesTheBrokerUntouched(t *testing.T) {
	for _, tc := range []struct {
		name   string
		answer string
	}{
		{"a lenient y is not the exact yes", "y\n"},
		{"stdin ends unanswered", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := replNewRig(t, broker.RoleStandby, broker.RoleStandby)
			g.app.PromptIn = strings.NewReader(tc.answer)

			stdout, stderr, err := g.configure(t)
			if err != nil {
				t.Fatalf("a declined confirmation is an answer, not a failure: %v", err)
			}
			if g.asked != 1 {
				t.Errorf("the question was asked %d time(s), want once", g.asked)
			}
			g.wantOnlyTheNameRead(t)
			if stdout != "" {
				t.Errorf("stdout = %q, want no report for a run that applied nothing", stdout)
			}
			if tc.answer == "" && !strings.Contains(stderr, "--no-prompt") {
				t.Errorf("stderr = %q, want the refusal to name --no-prompt", stderr)
			}
		})
	}
}

// TestOpConfigureReplicationRefusesAnUnknownBrokerBeforeAsking covers a broker the file does
// not describe -- the env file of the wrong deployment, say. The router name it reports
// matches no site, which is refused with the names side by side, and the operator is never
// asked to confirm a convergence of a broker the tool cannot place.
func TestOpConfigureReplicationRefusesAnUnknownBrokerBeforeAsking(t *testing.T) {
	g := replNewRig(t, broker.RoleStandby, broker.RoleStandby)
	g.router = "nowhere"

	_, _, err := g.configure(t)
	if err == nil {
		t.Fatal("a broker that is in neither site's routerNames must be refused")
	}
	for _, want := range []string{`"nowhere"`, "sg1, dr1"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
	if g.asked != 0 {
		t.Errorf("the confirmation was asked %d time(s) about a broker that could not be placed", g.asked)
	}
	g.wantOnlyTheNameRead(t)
}

// TestOpConfigureReplicationFailureReportsNothingApplied proves an error from the apply is
// returned as it is and is not buried under a report: the listed VPN this broker does not
// have is refused before anything is written, names the VPN, and prints no "applied" page.
// The question was asked first, so this is the refusal an operator reads after saying yes.
func TestOpConfigureReplicationFailureReportsNothingApplied(t *testing.T) {
	g := replNewRig(t, broker.RoleStandby, broker.RoleStandby)
	g.local = map[string]broker.VPNRepl{"default": replUp(broker.RoleStandby)}

	stdout, _, err := g.configure(t)
	if err == nil {
		t.Fatal("a listed VPN the broker does not have must be refused")
	}
	for _, want := range []string{"ORDERS", "nothing was written"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
	if stdout != "" {
		t.Errorf("stdout = %q, want no report after a failure", stdout)
	}
	if g.rr.hasCall("repl-configure") {
		t.Errorf("a configuration script was sent although the apply was refused:\n%s", g.rr.dump())
	}
	if g.asked != 1 {
		t.Errorf("the question was asked %d time(s), want once", g.asked)
	}
}

// TestOpK8sReplicationOpsReadThePodFromTheFlag covers the two Kubernetes entry points. --pod
// is resolved before anything is asked of a cluster: an unknown word is a usage error naming
// it, with no command issued, and a valid one reaches every call the op makes -- the
// replication role config-syncs within the HA group, so the node chosen here is the node
// written to.
func TestOpK8sReplicationOpsReadThePodFromTheFlag(t *testing.T) {
	for _, op := range []struct {
		name string
		run  func(*App) error
	}{
		{"configure", opK8sConfigureReplication},
		{"perform", opK8sPerformReplication},
	} {
		t.Run(op.name, func(t *testing.T) {
			t.Run("an unknown pod is a usage error", func(t *testing.T) {
				g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
				g.app.pod = "bogus"

				err := op.run(g.app)
				if err == nil || !isUsage(err) || !strings.Contains(err.Error(), "bogus") {
					t.Fatalf("err = %v, want a usage error naming the bad pod", err)
				}
				if len(g.rr.calls) != 0 {
					t.Errorf("a bad --pod reached the cluster:\n%s", g.rr.dump())
				}
			})

			t.Run("the pod flag reaches the broker", func(t *testing.T) {
				g := replNewRig(t, broker.RoleStandby, broker.RoleActive)
				g.app.pod = "backup"
				// A name matching no site ends the run right after the first read, which is
				// enough to see which node that read went to.
				g.router = "nowhere"

				var err error
				captureStderr(t, func() {
					err = op.run(g.app)
				})
				if err == nil {
					t.Fatal("a broker the file does not describe must be refused")
				}
				if len(g.rr.calls) == 0 {
					t.Fatal("the broker was never asked its router name")
				}
				for _, c := range g.rr.calls {
					argv := strings.Join(c.args, " ")
					if !strings.Contains(argv, "dev-broker-pubsubplus-b-0") || strings.Contains(argv, "-p-0") {
						t.Errorf("call %s %s did not go to the backup pod --pod named", c.name, argv)
					}
				}
			})
		})
	}
}

// TestOpCtrReplicationOpsRunOnTheContainerEngine covers the docker and podman entry points.
// They have no --pod: a container host runs one broker, so the op is the node-local one and
// every call goes through the container engine, never a cluster CLI. The site resolution
// still comes first, so a broker the file does not describe is refused before anything else.
func TestOpCtrReplicationOpsRunOnTheContainerEngine(t *testing.T) {
	env := "redundancy:\n  enabled: false\n  primary:\n    name: pri-host\n" +
		"image:\n  repo: solace-pubsub-standard\n  tag: \"10.10.1.128\"\n" +
		"semp:\n  adminPass: " + smokeAdminPass + "\n" +
		"docker: {}\npodman: {}\n" +
		replEnv[strings.Index(replEnv, "replication:\n"):]
	cfg := loadDirect(t, env, config.Docker)

	for _, op := range []struct {
		name string
		run  func(*App) error
	}{
		{"configure", opCtrConfigureReplication},
		{"perform", opCtrPerformReplication},
	} {
		t.Run(op.name, func(t *testing.T) {
			rr := &opRunner{}
			rr.output = func(c opCall) []byte {
				if opArgvMatch(c, ".repl-router-name.cli") {
					return []byte("!   Router: \"nowhere\"\n")
				}
				return nil
			}
			a := &App{Cfg: cfg, Platform: config.Docker, Runner: rr}

			var err error
			captureStderr(t, func() {
				err = op.run(a)
			})
			if err == nil || !strings.Contains(err.Error(), `"nowhere"`) {
				t.Fatalf("err = %v, want the unknown router name refused by name", err)
			}
			if len(rr.calls) == 0 {
				t.Fatal("the broker was never asked its router name")
			}
			if rr.calls[0].name != "docker" || rr.hasCall("kubectl") {
				t.Errorf("the router name was read through %q, want the container engine only:\n%s",
					rr.calls[0].name, rr.dump())
			}
		})
	}
}

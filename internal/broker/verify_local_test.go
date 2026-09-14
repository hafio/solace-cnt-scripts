package broker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"solace/internal/config"
)

// These tests exercise the primary-host HA operations (LocalRole, LeaderLocal,
// RedundancyCoordinated) over the shared fakeTransport from broker_test.go. The
// transport is node-local, so a scripted `show redundancy` sequence stands in
// for this host's view across successive poll iterations, and the mate SEMP
// channel is answered by the responder's curl branch. PollInterval is 0
// (newTestOps leaves it at its zero value), so nothing sleeps.

// localCfg is a redundancy-group config with a named node table for role
// detection, the mate addresses the SEMP channel needs, and admin credentials.
func localCfg(redundancy string) *config.Config {
	return &config.Config{
		SEMP: config.SEMP{AdminPass: "adminpw"},
		Redundancy: config.Redundancy{
			Enabled: redundancy,
			Primary: config.Node{Name: "pri-host", Addr: "10.0.0.11"},
			Backup:  config.Node{Name: "bkp-host", Addr: "10.0.0.12"},
			Monitor: config.Node{Name: "mon-host", Addr: "10.0.0.13"},
		},
	}
}

// newLocalOps builds an Ops over ft with a fixed detected hostname.
func newLocalOps(t *testing.T, redundancy, host string, ft *fakeTransport) (*Ops, *bytes.Buffer) {
	t.Helper()
	o, buf := newTestOps(t, localCfg(redundancy), ft)
	o.Hostname = func() (string, error) { return host, nil }
	return o, buf
}

// rd renders a minimal `show redundancy` output for the given fields.
func rd(role, cfgStatus, rdStatus, activity string) string {
	return "Configuration Status : " + cfgStatus +
		"\nRedundancy Status : " + rdStatus +
		"\nActive-Standby Role : " + role +
		"\nADB Link To Mate : Up\nADB Hello To Mate : Up" +
		"\nActivity Status : " + activity + "\n"
}

// sempOK is a canned mate SEMP reply: one 2xx status line and a successful
// execute-result, satisfying both the preflight and the revert-activity check.
const sempOK = "HTTP/1.1 200 OK\r\nContent-Type: text/xml\r\n\r\n" +
	"<rpc-reply semp-version=\"soltr/10_0\"><execute-result code=\"ok\"/></rpc-reply>"

// isCurl reports whether argv is a mate SEMP curl exec.
func isCurl(argv []string) bool { return len(argv) > 0 && argv[0] == "curl" }

// curlCalls returns every recorded mate SEMP exec, in order.
func curlCalls(ft *fakeTransport) []recOutput {
	var calls []recOutput
	for _, out := range ft.outputs {
		if isCurl(out.argv) {
			calls = append(calls, out)
		}
	}
	return calls
}

// seqTransport answers show-rd with seq[i] (advancing i), the mate SEMP channel
// with sempOK, and nil for every other CLI exec (release/no-release). It returns
// a pointer to the consumed count so a test can assert the whole sequence was
// walked. Once the sequence is exhausted it repeats the last entry rather than
// panicking, so a stuck poll surfaces as a wrong consumed-count instead of an
// index panic.
func seqTransport(seq []string) (*fakeTransport, *int) {
	i := 0
	ft := &fakeTransport{}
	ft.responder = func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if isCurl(argv) {
			return []byte(sempOK), nil
		}
		if !matchCLI(argv, "show-rd") {
			return nil, nil
		}
		out := seq[len(seq)-1]
		if i < len(seq) {
			out = seq[i]
			i++
		}
		return []byte(out), nil
	}
	return ft, &i
}

// --- LocalRole -------------------------------------------------------------

func TestLocalRole(t *testing.T) {
	cases := []struct {
		arg, host string
		want      config.Role
		wantErr   bool
	}{
		{"backup", "anything", config.Backup, false},   // explicit arg wins
		{"m", "pri-host", config.Monitor, false},        // explicit short form wins over host
		{"", "pri-host", config.Primary, false},         // detect primary
		{"", "bkp-host", config.Backup, false},          // detect backup
		{"", "mon-host.example.com", config.Monitor, false}, // FQDN vs short name
		{"", "stranger", "", true},                      // no match -> loud error
		{"nonsense", "pri-host", "", true},              // bad explicit arg
		{"", "PRI-HOST", config.Primary, false},         // mixed case: host upper, config lower
	}
	for _, tc := range cases {
		o, _ := newLocalOps(t, "true", tc.host, &fakeTransport{})
		got, err := o.LocalRole(tc.arg)
		if tc.wantErr {
			if err == nil {
				t.Errorf("LocalRole(arg=%q host=%q) = %q, want error", tc.arg, tc.host, got)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Errorf("LocalRole(arg=%q host=%q) = %q,%v want %q", tc.arg, tc.host, got, err, tc.want)
		}
	}
}

// TestLocalRoleCaseInsensitiveConfigUpper pins the other direction of the fold:
// a configured node name in upper case must still match a lower-case hostname
// (hostMatches folds both the full-string and short-name compares, not just
// one side of one of them).
func TestLocalRoleCaseInsensitiveConfigUpper(t *testing.T) {
	cfg := localCfg("true")
	cfg.Redundancy.Primary.Name = "PRI-HOST"
	o, _ := newTestOps(t, cfg, &fakeTransport{})
	o.Hostname = func() (string, error) { return "pri-host", nil }
	got, err := o.LocalRole("")
	if err != nil || got != config.Primary {
		t.Errorf("LocalRole with upper-case config name = %q,%v want primary,nil", got, err)
	}
}

// TestLocalRoleAmbiguousHostname pins the new ambiguity guard: two node table
// entries sharing the same configured name must fail loud, naming both matched
// roles, rather than silently returning whichever the table lists first.
func TestLocalRoleAmbiguousHostname(t *testing.T) {
	cfg := localCfg("true")
	cfg.Redundancy.Backup.Name = cfg.Redundancy.Primary.Name // collision: same name configured for two roles
	o, _ := newTestOps(t, cfg, &fakeTransport{})
	o.Hostname = func() (string, error) { return cfg.Redundancy.Primary.Name, nil }
	_, err := o.LocalRole("")
	if err == nil {
		t.Fatal("LocalRole should fail loud when the hostname matches more than one configured role")
	}
	if !strings.Contains(err.Error(), "primary") || !strings.Contains(err.Error(), "backup") {
		t.Errorf("LocalRole ambiguous err = %v, want it to name both primary and backup", err)
	}
}

// TestLocalRoleExplicitArgSkipsAmbiguityCheck pins the escape hatch: an explicit
// role argument must win and skip detection entirely, even when the node table
// is ambiguous enough that detection alone would fail.
func TestLocalRoleExplicitArgSkipsAmbiguityCheck(t *testing.T) {
	cfg := localCfg("true")
	cfg.Redundancy.Backup.Name = cfg.Redundancy.Primary.Name
	o, _ := newTestOps(t, cfg, &fakeTransport{})
	o.Hostname = func() (string, error) { return cfg.Redundancy.Primary.Name, nil }
	got, err := o.LocalRole("backup")
	if err != nil || got != config.Backup {
		t.Errorf("LocalRole explicit arg over an ambiguous table = %q,%v want backup,nil", got, err)
	}
}

// --- DetectRole: the address pass ------------------------------------------
//
// A name cannot cover every host. A cloud instance reports something like
// "ip-10-0-0-12" while the env file names the broker "bkp-host", so the routername and
// the OS hostname are legitimately unrelated and the address is the only thing both
// ends agree on. These pin the second pass, and that it stays second.

// withAddrs is an Ops whose hostname matches nothing in the node table, carrying the
// given interface addresses.
func withAddrs(t *testing.T, host string, addrs ...string) *Ops {
	t.Helper()
	o, _ := newLocalOps(t, "true", host, &fakeTransport{})
	set := make(map[string]bool, len(addrs))
	for _, a := range addrs {
		set[a] = true
	}
	o.LocalAddrs = func() (map[string]bool, error) { return set, nil }
	return o
}

func TestDetectRoleByAddress(t *testing.T) {
	o := withAddrs(t, "ip-10-0-0-12", "127.0.0.1", "10.0.0.12")
	role, how, err := o.DetectRole()
	if err != nil || role != config.Backup {
		t.Fatalf("DetectRole = %q,%v want backup,nil -- the address is the only match", role, err)
	}
	if !strings.Contains(how, "address") || !strings.Contains(how, "10.0.0.12") {
		t.Errorf("how = %q, want it to say the match was by address, and which one", how)
	}
}

// TestDetectRoleNameBeatsAddress: a configured name always wins. The name is what the
// broker is called; an address can be shared by more than one interface, and a host
// carrying the backup's address while named after the primary is a misconfiguration to
// surface elsewhere -- not a reason to deploy the backup here.
func TestDetectRoleNameBeatsAddress(t *testing.T) {
	o := withAddrs(t, "pri-host", "10.0.0.12") // named primary, numbered backup
	role, how, err := o.DetectRole()
	if err != nil || role != config.Primary {
		t.Fatalf("DetectRole = %q,%v want primary,nil -- the name pass runs first", role, err)
	}
	if !strings.Contains(how, "hostname") {
		t.Errorf("how = %q, want it to say the match was by hostname", how)
	}
}

// TestDetectRoleAmbiguousAddress is the address half of the ambiguity guard: two node
// entries sharing an address must fail loud rather than return whichever comes first.
func TestDetectRoleAmbiguousAddress(t *testing.T) {
	o, _ := newLocalOps(t, "true", "stranger", &fakeTransport{})
	o.Cfg.Redundancy.Backup.Addr = o.Cfg.Redundancy.Primary.Addr
	o.LocalAddrs = func() (map[string]bool, error) {
		return map[string]bool{o.Cfg.Redundancy.Primary.Addr: true}, nil
	}
	_, _, err := o.DetectRole()
	if err == nil {
		t.Fatal("an address configured for two roles must fail loud")
	}
	for _, want := range []string{"primary", "backup", "address"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("ambiguous-address error must name %q, got: %v", want, err)
		}
	}
}

// TestDetectRoleNoMatchNamesBothPasses: the error is the operator's whole instruction,
// so it has to say that BOTH lookups were tried -- otherwise someone adds an addr to an
// env file that already has one and wonders why nothing changed.
func TestDetectRoleNoMatchNamesBothPasses(t *testing.T) {
	o := withAddrs(t, "stranger", "192.168.1.5")
	_, _, err := o.DetectRole()
	if err == nil {
		t.Fatal("a host matching neither a name nor an address must fail loud")
	}
	for _, want := range []string{"stranger", "redundancy.*.name", "redundancy.*.addr", "primary|backup|monitor"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("no-match error must contain %q, got: %v", want, err)
		}
	}
}

// TestDetectRoleAddrsError: the interface read can fail (a locked-down container, a
// permission-restricted host). That must surface as its own error naming the hostname
// that did not match, not as a bare "no match" that sends the operator looking at the
// node table.
func TestDetectRoleAddrsError(t *testing.T) {
	o, _ := newLocalOps(t, "true", "stranger", &fakeTransport{})
	o.LocalAddrs = func() (map[string]bool, error) { return nil, errors.New("no interfaces") }
	_, _, err := o.DetectRole()
	if err == nil || !strings.Contains(err.Error(), "no interfaces") || !strings.Contains(err.Error(), "stranger") {
		t.Errorf("DetectRole addr-read error = %v, want it to carry both the cause and the hostname", err)
	}
}

// TestDefaultLocalAddrs covers the unseamed path: the real interface list must come back
// as BARE addresses, since that is what an env file carries. net.InterfaceAddrs returns
// CIDRs ("10.0.0.11/24"), and leaving the mask on would make every addr match fail with
// nothing to show for it.
func TestDefaultLocalAddrs(t *testing.T) {
	addrs, err := defaultLocalAddrs()
	if err != nil {
		t.Skipf("interface enumeration unavailable here: %v", err)
	}
	for a := range addrs {
		if strings.Contains(a, "/") {
			t.Errorf("defaultLocalAddrs returned %q, want the bare address without the prefix length", a)
		}
	}
}

// TestLocalAddrsFallsBackToTheDefault closes the seam's nil branch: an Ops built without
// one still reads the machine, which is the production path.
func TestLocalAddrsFallsBackToTheDefault(t *testing.T) {
	o, _ := newTestOps(t, localCfg("true"), &fakeTransport{})
	o.LocalAddrs = nil
	if _, err := o.localAddrs(); err != nil {
		t.Skipf("interface enumeration unavailable here: %v", err)
	}
}

// --- LeaderLocal -----------------------------------------------------------

func TestLeaderLocalStandaloneSkips(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newLocalOps(t, "false", "pri-host", ft)
	if err := o.LeaderLocal(context.Background(), ""); err != nil {
		t.Fatalf("LeaderLocal standalone error: %v", err)
	}
	if len(ft.outputs) != 0 || len(ft.uploads) != 0 {
		t.Error("LeaderLocal must make no calls in standalone mode")
	}
}

func TestLeaderLocalRejectsNonPrimary(t *testing.T) {
	for _, host := range []string{"bkp-host", "mon-host"} {
		ft := &fakeTransport{}
		o, _ := newLocalOps(t, "true", host, ft)
		if err := o.LeaderLocal(context.Background(), ""); err == nil {
			t.Errorf("LeaderLocal on %q should fail loud", host)
		}
		if len(ft.uploads) != 0 {
			t.Errorf("LeaderLocal on %q must not upload before the guard", host)
		}
	}
	// Explicit backup arg is rejected the same way.
	ft := &fakeTransport{}
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	if err := o.LeaderLocal(context.Background(), "backup"); err == nil {
		t.Error("LeaderLocal with explicit backup arg should fail loud")
	}
}

func TestLeaderLocalSuccess(t *testing.T) {
	healthy := rd("Primary", "Enabled", "Up", "Local Active")
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		switch {
		case isCurl(argv):
			return []byte(sempOK), nil
		case matchCLI(argv, "show-rd"):
			return []byte(healthy), nil
		case matchCLI(argv, "assert-leader"):
			return []byte("l1\nl2\nSync Complete\n"), nil
		}
		return nil, nil
	}}
	o, buf := newLocalOps(t, "true", "pri-host", ft)
	if err := o.LeaderLocal(context.Background(), ""); err != nil {
		t.Fatalf("LeaderLocal error: %v", err)
	}
	if !uploadedForRole(ft, config.Primary, cliScriptPath("assert-leader")) {
		t.Error("LeaderLocal should assert leadership on the primary")
	}
	if !strings.Contains(buf.String(), "Sync Complete") {
		t.Errorf("LeaderLocal output = %q", buf.String())
	}
	if got := len(curlCalls(ft)); got != 2 {
		t.Errorf("LeaderLocal made %d mate SEMP calls, want 2 (preflight + revert-activity)", got)
	}
}

// TestLeaderLocalRevertsMateFirst pins the k8s Leader order: the mate
// revert-activity POST (the curl carrying a data line) happens before the
// first show-rd poll read, so a backup still holding activity is told to hand
// it back before the primary is polled for health.
func TestLeaderLocalRevertsMateFirst(t *testing.T) {
	healthy := rd("Primary", "Enabled", "Up", "Local Active")
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		switch {
		case isCurl(argv):
			return []byte(sempOK), nil
		case matchCLI(argv, "show-rd"):
			return []byte(healthy), nil
		case matchCLI(argv, "assert-leader"):
			return []byte("Sync Complete\n"), nil
		}
		return nil, nil
	}}
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	if err := o.LeaderLocal(context.Background(), ""); err != nil {
		t.Fatalf("LeaderLocal error: %v", err)
	}
	postIdx, showIdx := -1, -1
	for i, out := range ft.outputs {
		if postIdx == -1 && isCurl(out.argv) && strings.Contains(out.stdin, "data = ") {
			postIdx = i
		}
		if showIdx == -1 && matchCLI(out.argv, "show-rd") {
			showIdx = i
		}
	}
	if postIdx == -1 || showIdx == -1 || postIdx > showIdx {
		t.Errorf("mate revert POST at call %d, first show-rd at call %d; want the POST first", postIdx, showIdx)
	}
}

// TestLeaderLocalMateUnreachableWarnsAndContinues pins the deliberate
// downgrade: an unreachable mate SEMP service skips the revert-activity step
// with a warning -- the leader assertion itself is local -- instead of failing
// the whole command.
func TestLeaderLocalMateUnreachableWarnsAndContinues(t *testing.T) {
	healthy := rd("Primary", "Enabled", "Up", "Local Active")
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		switch {
		case isCurl(argv):
			return nil, errors.New("no route to host")
		case matchCLI(argv, "show-rd"):
			return []byte(healthy), nil
		case matchCLI(argv, "assert-leader"):
			return []byte("Sync Complete\n"), nil
		}
		return nil, nil
	}}
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	var logs []string
	o.Log = func(format string, args ...any) { logs = append(logs, fmt.Sprintf(format, args...)) }
	if err := o.LeaderLocal(context.Background(), ""); err != nil {
		t.Fatalf("LeaderLocal with unreachable mate error: %v, want warn-and-continue", err)
	}
	if !strings.Contains(strings.Join(logs, "\n"), "skipping the revert-activity step") {
		t.Errorf("LeaderLocal logs = %q, want the skip warning", logs)
	}
	for _, out := range curlCalls(ft) {
		if strings.Contains(out.stdin, "data = ") {
			t.Error("LeaderLocal must not POST revert-activity after the preflight failed")
		}
	}
}

// TestLeaderLocalMateRPCErrorFails is the other side of the downgrade: a
// REACHABLE mate refusing the revert-activity RPC is a real error, not a skip,
// and stops LeaderLocal before any local poll.
func TestLeaderLocalMateRPCErrorFails(t *testing.T) {
	rejected := "HTTP/1.1 200 OK\r\n\r\n<rpc-reply><execute-result code=\"fail\" reason=\"denied\"/></rpc-reply>"
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, in []byte) ([]byte, error) {
		switch {
		case isCurl(argv) && strings.Contains(string(in), "data = "):
			return []byte(rejected), nil
		case isCurl(argv):
			return []byte(sempOK), nil
		}
		return nil, nil
	}}
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	err := o.LeaderLocal(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "rejected the revert-activity") {
		t.Errorf("LeaderLocal mate-RPC-fail err = %v, want the rejected-RPC error", err)
	}
	for _, out := range ft.outputs {
		if matchCLI(out.argv, "show-rd") {
			t.Error("LeaderLocal must not poll after the mate rejected the RPC")
		}
	}
}

func TestLeaderLocalTimeoutDumpsDetail(t *testing.T) {
	down := rd("Primary", "Enabled", "Down", "Local Active")
	sawDetail := false
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		switch {
		case isCurl(argv):
			return []byte(sempOK), nil
		case matchCLI(argv, "show-redundancy-detail"):
			sawDetail = true
			return []byte("detail dump\n"), nil
		}
		return []byte(down), nil
	}}
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	o.PollAttempts = 2
	if err := o.LeaderLocal(context.Background(), ""); err == nil {
		t.Error("LeaderLocal should time out when redundancy never recovers")
	}
	if !sawDetail {
		t.Error("LeaderLocal should dump show-redundancy-detail on timeout")
	}
}

// --- RedundancyCoordinated guards -------------------------------------------

func TestRedundancyCoordinatedStandaloneSkips(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newLocalOps(t, "false", "pri-host", ft)
	if err := o.RedundancyCoordinated(context.Background(), ""); err != nil {
		t.Fatalf("RedundancyCoordinated standalone error: %v", err)
	}
	if len(ft.outputs) != 0 {
		t.Error("RedundancyCoordinated must make no calls in standalone mode")
	}
}

func TestRedundancyCoordinatedRejectsNonPrimary(t *testing.T) {
	for _, host := range []string{"bkp-host", "mon-host"} {
		ft := &fakeTransport{}
		o, _ := newLocalOps(t, "true", host, ft)
		if err := o.RedundancyCoordinated(context.Background(), ""); err == nil {
			t.Errorf("RedundancyCoordinated on %q should fail loud", host)
		}
		if len(ft.outputs) != 0 {
			t.Errorf("RedundancyCoordinated on %q must reject before any call", host)
		}
	}
	// Explicit backup arg is rejected the same way.
	ft := &fakeTransport{}
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	if err := o.RedundancyCoordinated(context.Background(), "backup"); err == nil {
		t.Error("RedundancyCoordinated with explicit backup arg should fail loud")
	}
}

func TestRedundancyCoordinatedBadRoleArg(t *testing.T) {
	o, _ := newLocalOps(t, "true", "pri-host", &fakeTransport{})
	if err := o.RedundancyCoordinated(context.Background(), "nonsense"); err == nil {
		t.Error("RedundancyCoordinated should propagate a bad explicit role arg")
	}
}

func TestRedundancyCoordinatedInitialShowError(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		if matchCLI(argv, "show-rd") {
			return nil, fmt.Errorf("cli unreachable")
		}
		return nil, nil
	}}
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	if err := o.RedundancyCoordinated(context.Background(), ""); err == nil {
		t.Error("RedundancyCoordinated should propagate an initial show redundancy error")
	}
}

func TestRedundancyCoordinatedUnhealthyPrimary(t *testing.T) {
	seq := []string{rd("Primary", "Enabled", "Down", "Local Active")} // active but redundancy Down
	ft, _ := seqTransport(seq)
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	if err := o.RedundancyCoordinated(context.Background(), ""); err == nil {
		t.Error("RedundancyCoordinated should fail loud when redundancy is not healthy")
	}
	if len(curlCalls(ft)) != 0 {
		t.Error("RedundancyCoordinated must not touch the mate when the primary is unhealthy")
	}
	if ft.hasUpload(cliScriptPath("release")) {
		t.Error("RedundancyCoordinated must not release activity when redundancy is unhealthy")
	}
}

// TestRedundancyCoordinatedMissingMateIP pins the loud refusal when the config
// cannot address the mate: the error names redundancy.backup.addr and nothing has been
// released.
func TestRedundancyCoordinatedMissingMateIP(t *testing.T) {
	seq := []string{rd("Primary", "Enabled", "Up", "Local Active")}
	ft, _ := seqTransport(seq)
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	o.Cfg.Redundancy.Backup.Addr = ""
	err := o.RedundancyCoordinated(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "redundancy.backup.addr") {
		t.Errorf("RedundancyCoordinated missing-IP err = %v, want it to name redundancy.backup.addr", err)
	}
	if ft.hasUpload(cliScriptPath("release")) {
		t.Error("RedundancyCoordinated must not release activity without a mate address")
	}
}

// TestRedundancyCoordinatedPreflightUnreachable pins the ordering that keeps a
// failed run harmless: the mate is reached BEFORE the first mutation, so an
// unreachable SEMP service aborts while the group is still undisturbed.
func TestRedundancyCoordinatedPreflightUnreachable(t *testing.T) {
	healthy := rd("Primary", "Enabled", "Up", "Local Active")
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		switch {
		case isCurl(argv):
			return nil, errors.New("connection refused")
		case matchCLI(argv, "show-rd"):
			return []byte(healthy), nil
		}
		return nil, nil
	}}
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	err := o.RedundancyCoordinated(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "mate SEMP preflight") {
		t.Errorf("RedundancyCoordinated preflight err = %v, want the preflight error", err)
	}
	if ft.hasUpload(cliScriptPath("release")) {
		t.Error("RedundancyCoordinated must not release activity after a failed preflight")
	}
}

// --- RedundancyCoordinated happy paths ----------------------------------------

func TestRedundancyCoordinatedSuccess(t *testing.T) {
	seq := []string{
		rd("Primary", "Enabled", "Up", "Local Active"),           // initial: active + healthy
		rd("Primary", "Enabled-Released", "Down", "Mate Active"), // released
		rd("Primary", "Enabled", "Up", "Mate Active"),            // un-released, backup active
		rd("Primary", "Enabled", "Up", "Mate Active"),            // re-read: backup confirmed active
		rd("Primary", "Enabled", "Up", "Local Active"),           // reverted back
	}
	ft, consumed := seqTransport(seq)
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	// The mate channel prefers TLS only when the broker has a server certificate
	// to serve it with, so this end-to-end flow states that precondition to assert
	// the https URLs below; without it the same run is correctly plaintext.
	o.Cfg.TLS = serverCert
	if err := o.RedundancyCoordinated(context.Background(), ""); err != nil {
		t.Fatalf("RedundancyCoordinated error: %v", err)
	}
	if *consumed != len(seq) {
		t.Errorf("coordinated run consumed %d/%d show-rd reads", *consumed, len(seq))
	}
	if !uploadedForRole(ft, config.Primary, cliScriptPath("release")) ||
		!uploadedForRole(ft, config.Primary, cliScriptPath("no-release")) {
		t.Error("coordinated run should release then un-release activity on the primary")
	}
	calls := curlCalls(ft)
	if len(calls) != 2 {
		t.Fatalf("coordinated run made %d mate SEMP calls, want 2 (preflight + revert-activity)", len(calls))
	}
	preflight, post := calls[0], calls[1]
	if got := preflight.argv[len(preflight.argv)-1]; got != "https://10.0.0.12:1943/SEMP/v2/monitor" {
		t.Errorf("preflight URL = %q", got)
	}
	if got := post.argv[len(post.argv)-1]; got != "https://10.0.0.12:1943/SEMP" {
		t.Errorf("revert-activity URL = %q", got)
	}
	if !strings.Contains(post.stdin, `data = "<rpc><admin><redundancy><revert-activity/></redundancy></admin></rpc>"`) {
		t.Errorf("revert-activity POST stdin = %q, want the SEMP v1 RPC body as a data line", post.stdin)
	}
	for _, c := range calls {
		if !strings.Contains(c.stdin, `user = "admin:adminpw"`) {
			t.Errorf("mate SEMP call stdin = %q, want the admin credentials as a user line", c.stdin)
		}
		for _, tok := range c.argv {
			if strings.Contains(tok, "adminpw") || strings.Contains(tok, "<rpc>") {
				t.Errorf("mate SEMP argv token %q leaks the credentials or RPC body", tok)
			}
		}
	}
}

// TestRedundancyCoordinatedStandbyPrimaryStillReverts covers the standby-start
// arm: a primary that is already standby skips the release walk but still
// reverts the mate and waits for activity to come home -- same as k8s.
func TestRedundancyCoordinatedStandbyPrimaryStillReverts(t *testing.T) {
	seq := []string{
		rd("Primary", "Enabled", "Up", "Mate Active"),  // initial: standby (mate active)
		rd("Primary", "Enabled", "Up", "Mate Active"),  // re-read: backup confirmed active
		rd("Primary", "Enabled", "Up", "Local Active"), // reverted back
	}
	ft, consumed := seqTransport(seq)
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	if err := o.RedundancyCoordinated(context.Background(), "primary"); err != nil {
		t.Fatalf("RedundancyCoordinated standby-start error: %v", err)
	}
	if *consumed != len(seq) {
		t.Errorf("standby-start consumed %d/%d show-rd reads", *consumed, len(seq))
	}
	if ft.hasUpload(cliScriptPath("release")) {
		t.Error("standby-start must not release activity")
	}
	if got := len(curlCalls(ft)); got != 2 {
		t.Errorf("standby-start made %d mate SEMP calls, want 2", got)
	}
}

// TestRedundancyCoordinatedNeverTargetsBackupRole pins the node-local-transport
// invariant: the coordinated flow must never pass config.Backup to the
// transport, because on a container host that silently re-reads the LOCAL
// broker under the wrong label instead of failing. The guard it adds is on the
// CLI-side calls (showRD/release/no-release), which use the flow's own resolved
// role; the curl leg is enforced by sempCurl hardcoding config.Primary, so no
// test could catch it regressing here.
func TestRedundancyCoordinatedNeverTargetsBackupRole(t *testing.T) {
	seq := []string{
		rd("Primary", "Enabled", "Up", "Local Active"),
		rd("Primary", "Enabled-Released", "Down", "Mate Active"),
		rd("Primary", "Enabled", "Up", "Mate Active"),
		rd("Primary", "Enabled", "Up", "Mate Active"),
		rd("Primary", "Enabled", "Up", "Local Active"),
	}
	ft, _ := seqTransport(seq)
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	if err := o.RedundancyCoordinated(context.Background(), ""); err != nil {
		t.Fatalf("RedundancyCoordinated error: %v", err)
	}
	for _, out := range ft.outputs {
		if out.role != config.Primary {
			t.Errorf("exec %v targeted role %q, want every call on config.Primary", out.argv, out.role)
		}
	}
	for _, up := range ft.uploads {
		if up.role != config.Primary {
			t.Errorf("upload to %q targeted role %q, want every call on config.Primary", up.dest, up.role)
		}
	}
}

// --- RedundancyCoordinated error / timeout branches ---------------------------

// TestRedundancyCoordinatedNeitherActive closes the takeover assert: when the
// local view reports no active mate after the release walk (or on a
// standby-start), the run stops loud before reverting anything.
func TestRedundancyCoordinatedNeitherActive(t *testing.T) {
	seq := []string{rd("Primary", "Enabled", "Up", "Nothing Active")} // healthy but nobody active
	ft, _ := seqTransport(seq)
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	err := o.RedundancyCoordinated(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "appears to be active") {
		t.Errorf("RedundancyCoordinated neither-active err = %v, want the neither-active error", err)
	}
	if got := len(curlCalls(ft)); got != 1 {
		t.Errorf("neither-active made %d mate SEMP calls, want 1 (preflight only, no revert POST)", got)
	}
}

// TestRedundancyCoordinatedMateRPCNotOK closes the revert-activity ok-check: a
// 2xx reply whose body lacks code="ok" fails loud and surfaces the mate's reply.
func TestRedundancyCoordinatedMateRPCNotOK(t *testing.T) {
	rejected := "HTTP/1.1 200 OK\r\n\r\n<rpc-reply><execute-result code=\"fail\" reason=\"denied\"/></rpc-reply>"
	seq := []string{
		rd("Primary", "Enabled", "Up", "Mate Active"), // standby-start, straight to the revert
		rd("Primary", "Enabled", "Up", "Mate Active"),
	}
	ft, _ := seqTransport(seq)
	base := ft.responder
	ft.responder = func(role config.Role, argv []string, in []byte) ([]byte, error) {
		if isCurl(argv) && strings.Contains(string(in), "data = ") {
			return []byte(rejected), nil
		}
		return base(role, argv, in)
	}
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	err := o.RedundancyCoordinated(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "rejected the revert-activity") {
		t.Errorf("RedundancyCoordinated RPC-not-ok err = %v, want the rejected-RPC error", err)
	}
}

// TestRedundancyCoordinatedRevertTimeout closes the final poll guard: if the
// primary never observes activity coming home after the revert, the run
// reports the bounded timeout rather than declaring success.
func TestRedundancyCoordinatedRevertTimeout(t *testing.T) {
	seq := []string{
		rd("Primary", "Enabled", "Up", "Mate Active"), // standby-start
		rd("Primary", "Enabled", "Up", "Mate Active"), // re-read; last entry repeats forever
	}
	ft, _ := seqTransport(seq)
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	o.PollAttempts = 2
	if err := o.RedundancyCoordinated(context.Background(), ""); err == nil {
		t.Error("RedundancyCoordinated should time out when activity never returns to the primary")
	}
	if got := len(curlCalls(ft)); got != 2 {
		t.Errorf("revert-timeout made %d mate SEMP calls, want 2 (the revert was sent)", got)
	}
}

// TestRedundancyCoordinatedReleaseError closes the release RunCLI guard: a
// failed release exec must stop before any show-rd poll is attempted.
func TestRedundancyCoordinatedReleaseError(t *testing.T) {
	healthy := rd("Primary", "Enabled", "Up", "Local Active")
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		switch {
		case isCurl(argv):
			return []byte(sempOK), nil
		case matchCLI(argv, "release"):
			return nil, errors.New("release boom")
		case matchCLI(argv, "show-rd"):
			return []byte(healthy), nil
		}
		return nil, nil
	}}
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	if err := o.RedundancyCoordinated(context.Background(), ""); err == nil {
		t.Error("RedundancyCoordinated should return the release RunCLI error")
	}
	showReads := 0
	for _, out := range ft.outputs {
		if matchCLI(out.argv, "show-rd") {
			showReads++
		}
	}
	if showReads != 1 {
		t.Errorf("saw %d show-rd reads, want only the initial one (no poll after release fails)", showReads)
	}
}

// TestRedundancyCoordinatedReleasedTimeout closes the "released to the Backup"
// poll guard: if the release never converges, no-release must never be sent --
// that would un-release a release that never took.
func TestRedundancyCoordinatedReleasedTimeout(t *testing.T) {
	seq := []string{rd("Primary", "Enabled", "Up", "Local Active")} // never Enabled-Released
	ft, _ := seqTransport(seq)
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	o.PollAttempts = 2
	if err := o.RedundancyCoordinated(context.Background(), ""); err == nil {
		t.Error("RedundancyCoordinated should time out waiting to be released")
	}
	if ft.hasUpload(cliScriptPath("no-release")) {
		t.Error("RedundancyCoordinated must not send no-release when released never became true")
	}
}

// TestRedundancyCoordinatedNoReleaseError closes the second RunCLI guard: a
// failed no-release exec must surface rather than proceeding to the
// un-released poll.
func TestRedundancyCoordinatedNoReleaseError(t *testing.T) {
	seq := []string{
		rd("Primary", "Enabled", "Up", "Local Active"),
		rd("Primary", "Enabled-Released", "Down", "Mate Active"),
	}
	ft, _ := seqTransport(seq)
	base := ft.responder
	ft.responder = func(role config.Role, argv []string, in []byte) ([]byte, error) {
		if matchCLI(argv, "no-release") {
			return nil, errors.New("no-release boom")
		}
		return base(role, argv, in)
	}
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	if err := o.RedundancyCoordinated(context.Background(), ""); err == nil {
		t.Error("RedundancyCoordinated should return the no-release RunCLI error")
	}
}

// TestRedundancyCoordinatedUnreleasedTimeout closes the "un-released" poll
// guard: if the un-release never converges, the run must report the timeout
// rather than reverting the mate on stale state.
func TestRedundancyCoordinatedUnreleasedTimeout(t *testing.T) {
	seq := []string{
		rd("Primary", "Enabled", "Up", "Local Active"),
		rd("Primary", "Enabled-Released", "Down", "Mate Active"), // stays released forever
	}
	ft, _ := seqTransport(seq)
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	o.PollAttempts = 2
	if err := o.RedundancyCoordinated(context.Background(), ""); err == nil {
		t.Error("RedundancyCoordinated should time out waiting to be un-released")
	}
	for _, out := range curlCalls(ft) {
		if strings.Contains(out.stdin, "data = ") {
			t.Error("RedundancyCoordinated must not revert the mate when un-release never converged")
		}
	}
}

// --- pure / detection arms -------------------------------------------------

func TestLocalRoleDefaultHostname(t *testing.T) {
	o, _ := newTestOps(t, localCfg("true"), &fakeTransport{})
	o.Hostname = nil // force the os.Hostname default
	// The real host name will not match the pri/bkp/mon table, so detection fails loud.
	if _, err := o.LocalRole(""); err == nil {
		t.Error("LocalRole with the default os.Hostname should fail loud off the node table")
	}
}

// --- LocalRole hostname error ------------------------------------------------

// TestLocalRoleHostnameError closes LocalRole's hostname-read error branch: if
// the injected Hostname func fails, LocalRole must wrap and return the error
// rather than falling through to the node-table match against a garbage/empty
// host.
func TestLocalRoleHostnameError(t *testing.T) {
	o, _ := newTestOps(t, localCfg("true"), &fakeTransport{})
	o.Hostname = func() (string, error) { return "", errors.New("boom") }
	_, err := o.LocalRole("")
	if err == nil || !strings.Contains(err.Error(), "detect node role") {
		t.Errorf("LocalRole hostname error = %v, want wrapped %q", err, "detect node role")
	}
}

// --- LeaderLocal error branches -----------------------------------------------

// TestLeaderLocalBadRoleArg mirrors TestRedundancyCoordinatedBadRoleArg: an
// invalid explicit role arg must stop LeaderLocal before the primary-only guard
// even runs.
func TestLeaderLocalBadRoleArg(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	if err := o.LeaderLocal(context.Background(), "nonsense"); err == nil {
		t.Error("LeaderLocal should propagate a bad explicit role arg")
	}
	if len(ft.outputs) != 0 || len(ft.uploads) != 0 {
		t.Error("LeaderLocal must make no transport calls on a bad role arg")
	}
}

// TestLeaderLocalPollCondError mirrors the k8s TestLeaderPollCondError,
// closing an asymmetry between the two implementations: a transport error
// mid-poll must abort immediately (the detail dump still runs), not be
// silently retried as "not yet healthy".
func TestLeaderLocalPollCondError(t *testing.T) {
	sawDetail := false
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		switch {
		case isCurl(argv):
			return []byte(sempOK), nil
		case matchCLI(argv, "show-rd"):
			return nil, errors.New("show boom")
		case matchCLI(argv, "show-redundancy-detail"):
			sawDetail = true
			return []byte("detail\n"), nil
		}
		return nil, nil
	}}
	o, _ := newLocalOps(t, "true", "pri-host", ft)
	if err := o.LeaderLocal(context.Background(), ""); err == nil {
		t.Error("LeaderLocal should return the poll condition error")
	}
	if !sawDetail {
		t.Error("LeaderLocal should still dump show-redundancy-detail after the error")
	}
}

// TestLeaderLocalAssertLeaderError closes LeaderLocal's final assert-leader
// error branch: after a healthy poll, a failing assert-leader exec must stop
// the function and never call o.show.
func TestLeaderLocalAssertLeaderError(t *testing.T) {
	healthy := rd("Primary", "Enabled", "Up", "Local Active")
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		switch {
		case isCurl(argv):
			return []byte(sempOK), nil
		case matchCLI(argv, "show-rd"):
			return []byte(healthy), nil
		case matchCLI(argv, "assert-leader"):
			return nil, errors.New("assert boom")
		}
		return nil, nil
	}}
	o, buf := newLocalOps(t, "true", "pri-host", ft)
	if err := o.LeaderLocal(context.Background(), ""); err == nil {
		t.Error("LeaderLocal should return the assert-leader error")
	}
	if buf.Len() != 0 {
		t.Errorf("LeaderLocal must not show output when assert-leader fails, got %q", buf.String())
	}
}

// --- exported surface ---------------------------------------------------------

// TestMateActivityStateReadsTheMateColumn covers the parser a live probe leans on
// before it sends anything: MateActivityState must say "the mate holds activity"
// only when the primary's own `show redundancy` says so, because a probe that
// misreads this would send revert-activity to an ACTIVE mate and fail a real HA
// group over. It is the same reading RedundancyCoordinated does, so the two
// cannot drift.
func TestMateActivityStateReadsTheMateColumn(t *testing.T) {
	for _, tc := range []struct {
		name string
		out  string
		want bool
	}{
		{"mate active", "Activity Status : Mate Active\r\n", true},
		{"local active", "Activity Status : Local Active\r\n", false},
		{"standby, nobody named", "Activity Status : Standby\n", false},
		{"empty output", "", false},
		{"mate active named on an unrelated label", "Some Other Field : Mate Active\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := MateActivityState(tc.out); got != tc.want {
				t.Errorf("MateActivityState(%q) = %v, want %v", tc.out, got, tc.want)
			}
		})
	}
}

// TestShowRedundancyIsReadOnly proves the exported wrapper is exactly showRD --
// one `show redundancy` CLI script on the named role and nothing else. A probe
// calls it to decide whether a mutation is safe, so it must not itself mutate.
func TestShowRedundancyIsReadOnly(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, _ []string, _ []byte) ([]byte, error) {
		return []byte("Activity Status : Mate Active\n"), nil
	}}
	o, _ := newTestOps(t, &config.Config{}, ft)
	out, err := o.ShowRedundancy(context.Background(), config.Backup)
	if err != nil {
		t.Fatalf("ShowRedundancy error: %v", err)
	}
	if !MateActivityState(out) {
		t.Errorf("ShowRedundancy returned %q, which MateActivityState should read as mate-active", out)
	}
	for _, c := range ft.outputs {
		if !matchCLI(c.argv, "show-rd") {
			t.Errorf("ShowRedundancy issued a non-show-redundancy command: %v", c.argv)
		}
	}
}

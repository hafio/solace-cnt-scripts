package broker

import (
	"context"
	"errors"
	"strings"
	"testing"

	"solace/internal/config"
)

// matechannel_test.go covers the CLI-driven MateChannel. The runner is a func, so these
// tests need no transport and no broker -- they assert exactly which scripts a channel
// runs and what it makes of the output.

// fakeRun records every script a channel runs and replies from a table keyed by the
// script NAME, so a test only has to supply the output it cares about.
type fakeRun struct {
	calls   []string // script names, in order
	scripts []string // script bodies, in order
	reply   map[string]string
	err     error
}

func (f *fakeRun) runner() CLIRunner {
	return func(_ context.Context, name, script string) ([]byte, error) {
		f.calls = append(f.calls, name)
		f.scripts = append(f.scripts, script)
		if f.err != nil {
			return nil, f.err
		}
		return []byte(f.reply[name]), nil
	}
}

// softwareBanner is what `cli -Apes` echoes ahead of any output on a software broker.
// Every reply carries it, because that is where the channel learns the broker's type.
const softwareBanner = "Solace Event Broker Software Enterprise Version 10.26.0.8827\n" +
	"license-software\n"

func mateReplies() map[string]string {
	return map[string]string{
		scriptShowReplication: softwareBanner + showReplicationSoftware,
		scriptShowVPN:         softwareBanner + showVPNReplication,
		scriptSetRole:         softwareBanner + "xps-ps-01> \n",
		scriptShowRedundancy:  softwareBanner + redundancyLocalActive,
	}
}

// The `show redundancy` shapes PrimaryActive reads. The `<label> : <value>` form is the
// one field/countContains expect and the one every other redundancy fixture in this
// package uses -- field returns everything after the FIRST ": " UNTRIMMED, so a
// column-padded `Configuration Status:    Enabled` yields "    Enabled" and matches
// nothing. Activity is one line per virtual router, which is why countContains needs the
// label and the value on the SAME line.
const (
	redundancyLocalActive = "Configuration Status : Enabled\nRedundancy Status : Up\n" +
		"Active-Standby Role : Primary\nActivity Status : Local Active\nActivity Status : Mate Active\n"

	// The same pair with the BACKUP holding activity: no Local Active line at all.
	redundancyMateActive = "Configuration Status : Enabled\nRedundancy Status : Up\n" +
		"Active-Standby Role : Primary\nActivity Status : Mate Active\nActivity Status : Mate Active\n"

	// A standalone broker: no HA group to hold activity for.
	redundancyShutdown = "Configuration Status : Shutdown\nRedundancy Status : Down\n"
)

// TestMateChannelPrimaryActive proves the gate reads real `show redundancy` output, and
// that a STANDALONE broker passes rather than being refused for failing to hold activity
// it cannot hold -- it is its own only node.
func TestMateChannelPrimaryActive(t *testing.T) {
	cases := []struct {
		name   string
		report string
		want   bool
	}{
		{"primary holds activity", redundancyLocalActive, true},
		{"backup holds activity", redundancyMateActive, false},
		{"standalone has no HA group", redundancyShutdown, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakeRun{reply: mateReplies()}
			f.reply[scriptShowRedundancy] = softwareBanner + tc.report
			c := NewCLIMate(f.runner(), "site v:broker-b")
			got, err := c.PrimaryActive(context.Background())
			if err != nil {
				t.Fatalf("PrimaryActive: %v", err)
			}
			if got != tc.want {
				t.Errorf("PrimaryActive = %v, want %v", got, tc.want)
			}
		})
	}
}

func newTestMate(t *testing.T) (MateChannel, *fakeRun) {
	t.Helper()
	f := &fakeRun{reply: mateReplies()}
	return NewCLIMate(f.runner(), "site v:broker-b"), f
}

// TestMateChannelPreflightLearnsTheBrokerType proves the preflight is a real read, and
// that the type comes from the BROKER rather than from configuration.
//
// It has to be discovered: the mate-address grammar and the `show replication` layout
// both differ between software and an appliance, and the env file deliberately does not
// say which a site is -- a `type:` key would be a second source of truth whose only
// effect when wrong is to emit CLI the target rejects.
func TestMateChannelPreflightLearnsTheBrokerType(t *testing.T) {
	c, f := newTestMate(t)
	if err := c.Preflight(context.Background()); err != nil {
		t.Fatalf("Preflight: %v", err)
	}
	if len(f.calls) != 1 {
		t.Errorf("Preflight ran %v, want exactly one read", f.calls)
	}

	// Cached: a second call costs no extra round trip.
	if _, err := c.ShowReplication(context.Background()); err != nil {
		t.Fatalf("ShowReplication: %v", err)
	}
	if n := strings.Count(strings.Join(f.calls, ","), "repl-show-replication"); n != 2 {
		t.Errorf("ran repl-show-replication %d times; the type is cached, so the second call is "+
			"the read itself and not a re-probe", n)
	}
}

// TestMateChannelPreflightRefusesAnUnknownType proves a banner that names neither
// platform stops the run.
//
// Guessing would be worse than failing: the grammar is not a preference, and a wrong
// choice emits `connect-port` lines to a software broker or six `connect-via` lines to
// an appliance, either of which the target rejects.
func TestMateChannelPreflightRefusesAnUnknownType(t *testing.T) {
	f := &fakeRun{reply: map[string]string{"repl-show-replication": "no banner here\n"}}
	c := NewCLIMate(f.runner(), "site v:broker-b")
	err := c.Preflight(context.Background())
	if err == nil {
		t.Fatal("a banner naming neither platform must stop the run")
	}
	if !strings.Contains(err.Error(), "site v:broker-b") {
		t.Errorf("the error must name the site it is about, got: %v", err)
	}
}

// TestMateChannelPreflightReportsTransportFailure proves an unreachable site fails at
// the preflight, which is what every flow calls BEFORE its first write so the pair is
// still undisturbed when it stops.
func TestMateChannelPreflightReportsTransportFailure(t *testing.T) {
	f := &fakeRun{err: errors.New("connection refused")}
	c := NewCLIMate(f.runner(), "site v:broker-b")
	err := c.Preflight(context.Background())
	if err == nil || !strings.Contains(err.Error(), "connection refused") {
		t.Fatalf("want the transport error passed through, got: %v", err)
	}
	if !strings.Contains(err.Error(), "cannot reach site v:broker-b") {
		t.Errorf("the error must name what could not be reached, got: %v", err)
	}
}

// TestMateChannelShowReplication proves the report is parsed in the grammar the banner
// selected, end to end through the channel.
func TestMateChannelShowReplication(t *testing.T) {
	c, _ := newTestMate(t)
	got, err := c.ShowReplication(context.Background())
	if err != nil {
		t.Fatalf("ShowReplication: %v", err)
	}
	if !SameEndpoints(got.Endpoints, mateSoftware().Endpoints) {
		t.Errorf("endpoints = %+v, want %+v", got.Endpoints, mateSoftware().Endpoints)
	}
	if got.VirtualRouterName != "v:broker-b" {
		t.Errorf("mate = %q", got.VirtualRouterName)
	}
}

// TestMateChannelReadRolesNarrowsToTheListedVPNs is the blast-radius guard.
//
// The broker reports EVERY VPN it has -- six in this table, including `#config-sync` and
// two the env file never mentions. A caller handed the whole table could act on VPNs
// nobody asked about, which is exactly what `replication.vpns` exists to bound.
func TestMateChannelReadRolesNarrowsToTheListedVPNs(t *testing.T) {
	c, _ := newTestMate(t)
	got, err := c.ReadRoles(context.Background(), []string{"default", "vpn-01"})
	if err != nil {
		t.Fatalf("ReadRoles: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d VPNs, want only the two asked about: %+v", len(got), got)
	}
	// Queue included: both are admin-UP, and only one of them has a queue -- which is
	// the whole reason the re-enable keyword reads this column instead of enablement.
	if got["default"] != (VPNRepl{Admin: AdminEnabled, Role: RoleActive, Queue: QueueDown}) {
		t.Errorf("default = %+v", got["default"])
	}
	if got["vpn-01"] != (VPNRepl{Admin: AdminEnabled, Role: RoleStandby, Queue: QueueNA}) {
		t.Errorf("vpn-01 = %+v", got["vpn-01"])
	}

	// A VPN the broker does not have is ABSENT rather than an error: the caller knows
	// what it asked for and reports the difference with the context to explain it.
	missing, err := c.ReadRoles(context.Background(), []string{"default", "nope"})
	if err != nil {
		t.Fatalf("ReadRoles: %v", err)
	}
	if _, ok := missing["nope"]; ok {
		t.Error("a VPN the broker does not have must be absent from the map")
	}
	if len(missing) != 1 {
		t.Errorf("got %+v, want just default", missing)
	}
}

// TestMateChannelSetRole proves the role command is the bare-keyword form, quotes the
// VPN name, and turns paging off like every other script.
func TestMateChannelSetRole(t *testing.T) {
	c, f := newTestMate(t)
	if err := c.SetRole(context.Background(), "A VPN WITH SPACES", RoleActive); err != nil {
		t.Fatalf("SetRole: %v", err)
	}
	script := f.scripts[len(f.scripts)-1]
	for _, want := range []string{
		"home\nno paging\n",
		`message-vpn "A VPN WITH SPACES"`,
		"state active",
	} {
		if !strings.Contains(script, want) {
			t.Errorf("script missing %q:\n%s", want, script)
		}
	}
	// `state "active"` is how a capture ECHOES it; the grammar is a bare keyword.
	if strings.Contains(script, `state "active"`) {
		t.Errorf("the role operand is a bare keyword, not quoted:\n%s", script)
	}
}

// TestMateChannelSetRoleRefusesNA proves n/a cannot be written. It is a READING -- what
// a VPN that does not replicate reports -- and treating it as a settable state would
// turn a parse result straight back into a command.
func TestMateChannelSetRoleRefusesNA(t *testing.T) {
	c, f := newTestMate(t)
	err := c.SetRole(context.Background(), "default", RoleNA)
	if err == nil {
		t.Fatal("n/a is a reading, not a state to write")
	}
	for _, name := range f.calls {
		if name == "repl-set-role" {
			t.Error("a refused role must not reach the broker at all")
		}
	}
}

// TestMateChannelSetRoleDetectsRejection is the exit-0 guard.
//
// `cli -Apes` exits 0 even when the broker rejects a line, so the transcript is the only
// evidence there is -- the fact the whole import driver exists to work around. A role
// change that silently did nothing would leave a switchover believing it had moved.
func TestMateChannelSetRoleDetectsRejection(t *testing.T) {
	for _, kw := range failKeywords {
		t.Run(kw, func(t *testing.T) {
			f := &fakeRun{reply: mateReplies()}
			f.reply["repl-set-role"] = softwareBanner + "configure\n" + kw + " something\n"
			c := NewCLIMate(f.runner(), "site v:broker-b")
			err := c.SetRole(context.Background(), "default", RoleStandby)
			if err == nil {
				t.Fatalf("a transcript carrying %q must fail despite exit 0", kw)
			}
			if !strings.Contains(err.Error(), "default") {
				t.Errorf("the error must name the VPN, got: %v", err)
			}
		})
	}
}

// TestMateChannelRejectionScanIsTailOnly proves the scan reads the END of a transcript.
//
// Scanning the whole thing would scan the configuration itself, where an object
// legitimately named `error-events` matches `error:` -- which is why the driver reads a
// tail, and why this shares that rule rather than inventing a second one.
func TestMateChannelRejectionScanIsTailOnly(t *testing.T) {
	head := "error: this is old output from far above\n" + strings.Repeat("filler\n", driverTailLines+5)
	f := &fakeRun{reply: mateReplies()}
	f.reply["repl-set-role"] = softwareBanner + head
	c := NewCLIMate(f.runner(), "site v:broker-b")
	if err := c.SetRole(context.Background(), "default", RoleStandby); err != nil {
		t.Errorf("a keyword outside the tail must not fail the command: %v", err)
	}
}

// TestMateChannelStatusIsBothReports proves Status returns the raw text of both `show`
// commands -- and that it is DISPLAY only, never a source of decisions.
func TestMateChannelStatusIsBothReports(t *testing.T) {
	c, _ := newTestMate(t)
	got, err := c.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !strings.Contains(got, "Replication Mate:") {
		t.Error("Status must carry the mate report")
	}
	if !strings.Contains(got, "Message VPN") {
		t.Error("Status must carry the per-VPN table")
	}
}

// TestMateChannelDescribeNamesTheTarget proves every channel can say what it is, so an
// error from a two-site flow says WHICH site it is about.
func TestMateChannelDescribeNamesTheTarget(t *testing.T) {
	c, _ := newTestMate(t)
	if c.Describe() != "site v:broker-b" {
		t.Errorf("Describe = %q", c.Describe())
	}
}

// TestLocalMateRunsThroughOps proves the local channel is the same implementation over
// the transport the ops already hold, addressed at the role given -- always a site's
// PRIMARY, since the replication role config-syncs to that site's HA backup.
func TestLocalMateRunsThroughOps(t *testing.T) {
	var roles []config.Role
	ft := &fakeTransport{responder: func(role config.Role, argv []string, _ []byte) ([]byte, error) {
		roles = append(roles, role)
		if len(argv) > 0 && strings.HasSuffix(argv[0], "cli") {
			return []byte(softwareBanner + showReplicationSoftware), nil
		}
		return nil, nil
	}}
	o, _ := newTestOps(t, localCfg("true"), ft)
	c := o.NewLocalMate(config.Primary)
	if err := c.Preflight(context.Background()); err != nil {
		t.Fatalf("Preflight through Ops: %v", err)
	}
	if c.Describe() == "" {
		t.Error("the local channel must describe itself too")
	}
	for _, r := range roles {
		if r != config.Primary {
			t.Errorf("addressed role %q, want the primary only", r)
		}
	}
}

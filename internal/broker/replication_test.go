package broker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"solace/internal/config"
)

// replication_test.go covers the mate renderers and the two `show replication` readers.
//
// The fixtures below are EMBEDDED rather than read from semp/, because the whole semp/
// tree is gitignored: every real capture is absent on a fresh clone and in CI, so a test
// that only read them would silently skip exactly where it matters most. Their layout --
// the column the values sit in, the empty transport headings, the continuation line --
// is transcribed from the real reports; only the hosts and addresses are synthetic, so
// no production address is committed. The real captures are still exercised, by the
// skip-when-absent tests at the end, which is what keeps the transcription honest.

// showReplicationSoftware mirrors semp/show-replication-sw.out: no Connect-Via and no
// Connect-Ports heading, transports owning host:port LISTS, unconfigured transports
// printing their heading and nothing else, and the echoed prompt the CLI leaves behind.
const showReplicationSoftware = `
Replication Interface:
Replication Mate:                  v:broker-b
    Plain Text:

    Compressed:

    SSL:                           broker-b:55443
                                   broker-b.cluster_default:55443

ConfigSync:
  Bridge:
    Admin State:                   Enabled
    State:                         up
    Authentication:
      Pre-Shared Key:              Configured
      Insecure Upgrade Mode:       Disabled
    Compressed:                    No
    Message Spool:
      Window Size:                 65535
    Retry Delay:                   3

broker-a> `

// showReplicationAppliance mirrors semp/show-replication-appl.out: one Connect-Via, a
// Connect-Ports group holding a port per transport, and a top-level SSL cipher section
// that must not be mistaken for a transport heading.
const showReplicationAppliance = `
Replication Interface:
Replication Mate:                  v:appl-b
  Connect-Via:                     10.0.0.9
  Connect-Ports:
    Plain Text:                    55555
    Compressed:                    55003
    SSL:                           55443
SSL:
  Default Cipher Suite List:       Yes
  Cipher Suites:                   ECDHE-RSA-AES256-GCM-SHA384
                                   AES256-SHA
  Trusted Common Names:

ConfigSync:
  Bridge:
    Admin State:                   Enabled
    State:                         up
`

// mateSoftware is the pair the software fixture describes.
func mateSoftware() MateConfig {
	return MateConfig{
		VirtualRouterName: "v:broker-b",
		Endpoints: []config.ReplEndpoint{
			{Host: "broker-b", Port: 55443, Transport: config.TransportSSL},
			{Host: "broker-b.cluster_default", Port: 55443, Transport: config.TransportSSL},
		},
	}
}

// mateAppliance is the pair the appliance fixture describes.
func mateAppliance() MateConfig {
	return MateConfig{
		VirtualRouterName: "v:appl-b",
		Endpoints: []config.ReplEndpoint{
			{Host: "10.0.0.9", Port: 55555},
			{Host: "10.0.0.9", Port: 55003, Transport: config.TransportCompressed},
			{Host: "10.0.0.9", Port: 55443, Transport: config.TransportSSL},
		},
	}
}

func wantLines(t *testing.T, got []string, err error, want ...string) {
	t.Helper()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("rendered:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// TestRenderMateSoftware proves the software grammar: one connect-via per endpoint with
// the port inside the address, the router name standing alone, and NO keyword on a
// plain-text entry -- the CLI's own default, and what keeps two spellings of one thing
// out of the emitted script.
func TestRenderMateSoftware(t *testing.T) {
	got, err := RenderMate(BrokerSoftware, mateSoftware())
	wantLines(t, got, err,
		`replication mate connect-via "broker-b:55443" "ssl"`,
		`replication mate connect-via "broker-b.cluster_default:55443" "ssl"`,
		`replication mate virtual-router-name "v:broker-b"`,
	)

	t.Run("plain text carries no keyword", func(t *testing.T) {
		m := MateConfig{VirtualRouterName: "v:b", Endpoints: []config.ReplEndpoint{{Host: "h", Port: 55555}}}
		got, err := RenderMate(BrokerSoftware, m)
		wantLines(t, got, err,
			`replication mate connect-via "h:55555"`,
			`replication mate virtual-router-name "v:b"`,
		)
	})

	// Two endpoints on one transport is the software case an appliance cannot express,
	// and the reason the model is a flat list rather than hosts x ports.
	t.Run("two addresses on one transport", func(t *testing.T) {
		got, err := RenderMate(BrokerSoftware, mateSoftware())
		if err != nil {
			t.Fatalf("render: %v", err)
		}
		if n := strings.Count(strings.Join(got, "\n"), "connect-via"); n != 2 {
			t.Errorf("got %d connect-via lines, want exactly 2 -- not a cross product of hosts and ports", n)
		}
	})
}

// TestRenderMateAppliance proves the appliance grammar: a connect-port per transport,
// the single address hung off the router-name line, and never a port on that address.
func TestRenderMateAppliance(t *testing.T) {
	got, err := RenderMate(BrokerAppliance, mateAppliance())
	wantLines(t, got, err,
		`replication mate connect-port 55555`,
		`replication mate connect-port 55003 compressed`,
		`replication mate connect-port 55443 ssl`,
		`replication mate virtual-router-name "v:appl-b" connect-via "10.0.0.9"`,
	)

	// The appliance grammar has no `plain-text` keyword at all, so a bare connect-port
	// is the plain-text case by definition rather than a shorthand.
	if strings.Contains(strings.Join(got, "\n"), "plain-text") {
		t.Error("the appliance grammar has no plain-text keyword; a bare connect-port is that case")
	}
	if strings.Contains(strings.Join(got, "\n"), `connect-via "10.0.0.9:`) {
		t.Error("a port on an appliance connect-via is illegal and must never be emitted")
	}
}

// TestRenderMateApplianceRefusesUnrenderable proves the two sets an appliance cannot
// express are refused by name rather than quietly truncated. Dropping either would point
// replication at an address the operator never chose.
func TestRenderMateApplianceRefusesUnrenderable(t *testing.T) {
	t.Run("two distinct hosts", func(t *testing.T) {
		_, err := RenderMate(BrokerAppliance, mateSoftware()) // two different hostnames
		if err == nil {
			t.Fatal("an appliance has ONE connect-via; two hosts must be refused")
		}
		for _, want := range []string{"broker-b", "broker-b.cluster_default"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q must name the offending host %q", err, want)
			}
		}
	})

	t.Run("two endpoints on one transport", func(t *testing.T) {
		m := MateConfig{VirtualRouterName: "v:b", Endpoints: []config.ReplEndpoint{
			{Host: "h", Port: 55443, Transport: config.TransportSSL},
			{Host: "h", Port: 55444, Transport: config.TransportSSL},
		}}
		if _, err := RenderMate(BrokerAppliance, m); err == nil {
			t.Fatal("connect-port is not repeatable per transport; two ssl ports must be refused")
		}
	})
}

// TestRenderMateNeedsABrokerType proves an unknown type is an error rather than a
// default. The grammar is not a preference -- guessing it emits CLI the target rejects.
func TestRenderMateNeedsABrokerType(t *testing.T) {
	if _, err := RenderMate(BrokerUnknown, mateAppliance()); err == nil {
		t.Error("rendering without a broker type must be refused, not defaulted")
	}
	if _, err := ParseShowReplication(BrokerUnknown, []byte(showReplicationAppliance)); err == nil {
		t.Error("reading `show replication` without a broker type must be refused")
	}
}

// TestRenderMateRejectsEmpty covers the two inputs that would render a mate pointing
// nowhere.
func TestRenderMateRejectsEmpty(t *testing.T) {
	if _, err := RenderMate(BrokerSoftware, MateConfig{Endpoints: mateSoftware().Endpoints}); err == nil {
		t.Error("a mate with no virtual-router-name must be refused")
	}
	if _, err := RenderMate(BrokerSoftware, MateConfig{VirtualRouterName: "v:b"}); err == nil {
		t.Error("a mate with no endpoints must be refused")
	}
}

// TestParseShowReplicationSoftware proves the transport-grouped layout is read
// correctly: the continuation line is a second address for the SAME transport, and the
// empty Plain Text / Compressed headings are NOT endpoints -- a heading prints whether
// or not that transport is configured.
func TestParseShowReplicationSoftware(t *testing.T) {
	got, err := ParseShowReplication(BrokerSoftware, []byte(showReplicationSoftware))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.VirtualRouterName != "v:broker-b" {
		t.Errorf("mate = %q, want v:broker-b", got.VirtualRouterName)
	}
	if len(got.Endpoints) != 2 {
		t.Fatalf("got %d endpoints, want 2 -- the empty Plain Text and Compressed headings are not "+
			"endpoints: %+v", len(got.Endpoints), got.Endpoints)
	}
	if !SameEndpoints(got.Endpoints, mateSoftware().Endpoints) {
		t.Errorf("endpoints = %+v, want %+v", got.Endpoints, mateSoftware().Endpoints)
	}
	// The host half carries dots and an underscore (a container network alias in the
	// real capture), which is why the address splits on the LAST colon.
	if got.Endpoints[1].Host != "broker-b.cluster_default" {
		t.Errorf("continuation host = %q; the split must be on the last colon", got.Endpoints[1].Host)
	}
}

// TestParseShowReplicationAppliance proves the one-host layout is read correctly, and
// that the top-level SSL cipher section is not mistaken for a transport heading -- it
// carries the word SSL and sits outside Connect-Ports.
func TestParseShowReplicationAppliance(t *testing.T) {
	got, err := ParseShowReplication(BrokerAppliance, []byte(showReplicationAppliance))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.VirtualRouterName != "v:appl-b" {
		t.Errorf("mate = %q, want v:appl-b", got.VirtualRouterName)
	}
	if !SameEndpoints(got.Endpoints, mateAppliance().Endpoints) {
		t.Errorf("endpoints = %+v, want %+v", got.Endpoints, mateAppliance().Endpoints)
	}
	for _, e := range got.Endpoints {
		if e.Host != "10.0.0.9" {
			t.Errorf("endpoint %+v: every appliance endpoint shares the single Connect-Via host", e)
		}
	}
}

// TestParseShowReplicationWrongPlatformFails is the guard against a silent empty result.
// The two layouts are different enough that each reader run on the other's report must
// fail loudly rather than report a mate that simply has no addresses -- a caller would
// read that as "replication is unconfigured" and carry on.
func TestParseShowReplicationWrongPlatformFails(t *testing.T) {
	if _, err := ParseShowReplication(BrokerSoftware, []byte(showReplicationAppliance)); err == nil {
		t.Error("the software reader must fail on an appliance report, not return an empty mate")
	}
	_, err := ParseShowReplication(BrokerAppliance, []byte(showReplicationSoftware))
	if err == nil {
		t.Fatal("the appliance reader must fail on a software report, not return an empty mate")
	}
	if !strings.Contains(err.Error(), "Connect-Via") {
		t.Errorf("the error should say what was missing, got: %v", err)
	}
}

// TestParseShowReplicationUnconfigured proves a broker with no mate parses to an empty
// MateConfig rather than an error. That is the ordinary state of most brokers, and it is
// distinguishable from the wrong-platform case only because the mate is unnamed too.
func TestParseShowReplicationUnconfigured(t *testing.T) {
	const none = "\nReplication Interface:\nReplication Mate:\n\nConfigSync:\n  Bridge:\n"
	for _, bt := range []BrokerType{BrokerSoftware, BrokerAppliance} {
		got, err := ParseShowReplication(bt, []byte(none))
		if err != nil {
			t.Errorf("%s: an unconfigured broker must parse, not error: %v", bt, err)
		}
		if got.VirtualRouterName != "" || len(got.Endpoints) != 0 {
			t.Errorf("%s: want an empty MateConfig, got %+v", bt, got)
		}
	}
}

// TestMateRoundTripsThroughItsOwnGrammar is the end-to-end claim: what a broker reports
// re-renders to what a broker would be told. Both directions, both platforms.
func TestMateRoundTripsThroughItsOwnGrammar(t *testing.T) {
	t.Run("appliance", func(t *testing.T) {
		parsed, err := ParseShowReplication(BrokerAppliance, []byte(showReplicationAppliance))
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		got, err := RenderMate(BrokerAppliance, parsed)
		wantLines(t, got, err,
			`replication mate connect-port 55555`,
			`replication mate connect-port 55003 compressed`,
			`replication mate connect-port 55443 ssl`,
			`replication mate virtual-router-name "v:appl-b" connect-via "10.0.0.9"`,
		)
	})
	t.Run("software", func(t *testing.T) {
		parsed, err := ParseShowReplication(BrokerSoftware, []byte(showReplicationSoftware))
		if err != nil {
			t.Fatalf("parse: %v", err)
		}
		got, err := RenderMate(BrokerSoftware, parsed)
		wantLines(t, got, err,
			`replication mate connect-via "broker-b:55443" "ssl"`,
			`replication mate connect-via "broker-b.cluster_default:55443" "ssl"`,
			`replication mate virtual-router-name "v:broker-b"`,
		)
	})
}

// TestSameEndpointsNormalisesPlainText proves the comparison treats an absent transport
// and an explicit plainText as one thing. Import's diff is directional, so a keyword this
// tool omits and a broker echoes back would otherwise read as a missing line on every
// plain-text endpoint -- a false failure on every run.
func TestSameEndpointsNormalisesPlainText(t *testing.T) {
	bare := []config.ReplEndpoint{{Host: "h", Port: 55555}}
	spelled := []config.ReplEndpoint{{Host: "h", Port: 55555, Transport: config.TransportPlainText}}
	if !SameEndpoints(bare, spelled) {
		t.Error("an omitted transport and an explicit plainText are the same endpoint")
	}
	// Order must not matter either: the broker prints what it holds, not what was sent.
	a := mateAppliance().Endpoints
	b := []config.ReplEndpoint{a[2], a[0], a[1]}
	if !SameEndpoints(a, b) {
		t.Error("endpoint comparison must ignore order")
	}
	if SameEndpoints(bare, []config.ReplEndpoint{{Host: "h", Port: 55555, Transport: config.TransportSSL}}) {
		t.Error("a different transport is a different endpoint")
	}
}

// TestRenderMateRemovals proves the apply is convergent rather than additive. Without
// these lines, pointing a broker at a new mate would leave the old addresses in place:
// a second run would not be a no-op, and the broker would keep dialling somewhere nobody
// named.
func TestRenderMateRemovals(t *testing.T) {
	t.Run("software names the address being cleared", func(t *testing.T) {
		have := mateSoftware()
		want := MateConfig{VirtualRouterName: "v:broker-b", Endpoints: have.Endpoints[:1]}
		got, err := RenderMateRemovals(BrokerSoftware, have, want)
		wantLines(t, got, err, `no replication mate connect-via "broker-b.cluster_default:55443"`)
	})

	t.Run("appliance names only the transport", func(t *testing.T) {
		have := mateAppliance()
		want := MateConfig{VirtualRouterName: "v:appl-b", Endpoints: have.Endpoints[:1]}
		got, err := RenderMateRemovals(BrokerAppliance, have, want)
		wantLines(t, got, err,
			`no replication mate connect-port compressed`,
			`no replication mate connect-port ssl`,
		)
	})

	t.Run("nothing to do when converged", func(t *testing.T) {
		got, err := RenderMateRemovals(BrokerSoftware, mateSoftware(), mateSoftware())
		if err != nil {
			t.Fatalf("removals: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("a converged broker needs no removals, got %v", got)
		}
	})

	t.Run("router name cleared only when it changes", func(t *testing.T) {
		have := mateSoftware()
		moved := MateConfig{VirtualRouterName: "v:elsewhere", Endpoints: have.Endpoints}
		got, err := RenderMateRemovals(BrokerSoftware, have, moved)
		if err != nil {
			t.Fatalf("removals: %v", err)
		}
		if strings.Join(got, "\n") != "no replication mate virtual-router-name" {
			t.Errorf("removals = %v, want just the router-name clear", got)
		}
	})
}

// TestTransportVocabularyIsThreeWords pins the decision that `encrypted` is not accepted.
// The request schema offers it as a fourth keyword, but neither report emits it and the
// reply schema's six-entry software cap is three transports times two addresses -- which
// only holds if the vocabulary is three words. Mapping it silently would hide a broker
// that really did behave differently.
func TestTransportVocabularyIsThreeWords(t *testing.T) {
	for cli, want := range map[string]string{
		"":             config.TransportPlainText,
		"plain-text":   config.TransportPlainText,
		`"plain-text"`: config.TransportPlainText, // a capture quotes it
		"compressed":   config.TransportCompressed,
		"ssl":          config.TransportSSL,
	} {
		got, err := schemaTransport(cli)
		if err != nil || got != want {
			t.Errorf("schemaTransport(%q) = %q, %v; want %q", cli, got, err, want)
		}
	}
	if _, err := schemaTransport("encrypted"); err == nil {
		t.Error("`encrypted` must be refused by name rather than folded into ssl")
	}
	if _, err := cliTransport("encrypted"); err == nil {
		t.Error("`encrypted` must not be renderable from the schema side either")
	}
}

// showVPNReplication mirrors semp/show-message-vpn-replication.out: the legend, the
// fixed-width flag table, a system VPN reporting N/A for everything, and -- the case
// this whole parser is shaped around -- two VPN names containing spaces.
const showVPNReplication = `
Flags Legend:
A - Admin State (U=Up, D=Down, -=N/A)
C - Config State (A=Active, S=Standby, -=N/A)
B - Local Bridge State (U=Up, Q=Queue Unbound, D=Down, -=N/A)
R - Remote Bridge State (U=Up, D=Down, -=N/A)
Q - Queue State (U=Up, D=Down, -=N/A)
S - Sync Replication Eligible (Y=Yes, N=No, -=N/A)
M - Reject Msg When Sync Ineligible (Y=Yes, N=No, -=N/A)
T - Transaction Replication Mode (A=Async, S=Sync, -=N/A)

Message VPN                      A C B R Q S M T
-------------------------------- - - - - - - - -
#config-sync                     - - - - - - - -
A VPN WITH LONG NAME AND SPACES  D S - - - - N A
VPN WITH MANY QUEUES             D S - - - - N A
default                          U A - D D N Y A
vpn-01                           U S U - - - Y A
xps-ps-01                        D S - - - - N A
`

// TestParseVPNReplication is the spaces-in-names test above all else.
//
// Splitting this table on whitespace -- the obvious reading -- renames
// `A VPN WITH LONG NAME AND SPACES` to `A` AND shifts every flag one column, so the
// admin state is read out of the name. Three of the six rows here would be wrong. The
// column rule is what defines the fields, and it is read rather than assumed.
func TestParseVPNReplication(t *testing.T) {
	got, err := ParseVPNReplication([]byte(showVPNReplication))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// The Queue column is read too, and these values are the fixture's own `Q` column
	// read across -- which is what makes this table the evidence for the rule the
	// re-enable keyword depends on: `default` is admin-UP WITH a queue and `vpn-01` is
	// admin-UP with NONE, so queue existence cannot be derived from enablement.
	want := map[string]VPNRepl{
		"#config-sync":                    {Admin: AdminNA, Role: RoleNA, Queue: QueueNA},
		"A VPN WITH LONG NAME AND SPACES": {Admin: AdminShutdown, Role: RoleStandby, Queue: QueueNA},
		"VPN WITH MANY QUEUES":            {Admin: AdminShutdown, Role: RoleStandby, Queue: QueueNA},
		"default":                         {Admin: AdminEnabled, Role: RoleActive, Queue: QueueDown},
		"vpn-01":                          {Admin: AdminEnabled, Role: RoleStandby, Queue: QueueNA},
		"xps-ps-01":                       {Admin: AdminShutdown, Role: RoleStandby, Queue: QueueNA},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d VPNs, want %d: %+v", len(got), len(want), got)
	}
	for name, w := range want {
		g, ok := got[name]
		if !ok {
			t.Errorf("VPN %q missing; a name with spaces must survive the fixed-width split", name)
			continue
		}
		if g != w {
			t.Errorf("VPN %q = %+v, want %+v", name, g, w)
		}
	}
	if _, split := got["A"]; split {
		t.Error(`a VPN called "A" means the table was split on whitespace, not by column`)
	}
}

// TestVPNReplStateIsTwoFacts proves the pair is read as two independent facts. `default`
// here is enabled AND active; the space-named VPNs are shut down but still report
// standby, which is the case that makes reading the role alone wrong -- a role is
// present whether or not replication was ever turned on.
func TestVPNReplStateIsTwoFacts(t *testing.T) {
	got, err := ParseVPNReplication([]byte(showVPNReplication))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if v := got["VPN WITH MANY QUEUES"]; v.Role != RoleStandby || v.Admin != AdminShutdown {
		t.Errorf("%+v: a shut-down VPN still reports a role, sitting at its default", v)
	}
	if got["VPN WITH MANY QUEUES"].Replicating() {
		t.Error("a shut-down VPN is not replicating, whatever role it reports")
	}
	if !got["default"].Replicating() {
		t.Error("default is enabled and active, so it is replicating")
	}
	// N/A must stay N/A. Folding it into standby would let a switch promote against a
	// VPN it cannot actually see.
	if v := got["#config-sync"]; v.Admin != AdminNA || v.Role != RoleNA {
		t.Errorf("#config-sync = %+v, want n/a for both", v)
	}
	if got["#config-sync"].Replicating() {
		t.Error("an n/a VPN is not replicating")
	}
}

// TestParseVPNReplicationReadsColumnOrderFromTheHeader proves the flags are located by
// their header letter rather than by position.
//
// If Solace ever reorders or inserts a flag column, reading position 1 as admin state
// and position 2 as role would swap the two silently, with both values still looking
// valid -- the one misreading that could send a switchover the wrong way. Here C comes
// before A, and the values must follow the letters.
func TestParseVPNReplicationReadsColumnOrderFromTheHeader(t *testing.T) {
	const swapped = `
Message VPN                      C A B
-------------------------------- - - -
default                          A U -
`
	got, err := ParseVPNReplication([]byte(swapped))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if v := got["default"]; v.Admin != AdminEnabled || v.Role != RoleActive {
		t.Errorf("default = %+v, want admin enabled and role active: the columns are located by "+
			"their header letter, not by position", v)
	}
}

// TestParseVPNReplicationRejectsUnreadable covers the reports that must fail loudly
// rather than return a partial map a caller would act on.
func TestParseVPNReplicationRejectsUnreadable(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"no column rule", "Message VPN   A C\ndefault       U A\n", "no column rule"},
		{"no A or C columns", "Message VPN   B R\n-----------   - -\ndefault       U D\n", "no A (Admin State)"},
		{"unknown admin flag", "Message VPN   A C\n-----------   - -\ndefault       X A\n", "not one of U, D or -"},
		{"unknown role flag", "Message VPN   A C\n-----------   - -\ndefault       U Z\n", "not one of A, S or -"},
		{"no rows", "Message VPN   A C\n-----------   - -\n", "listed no message-VPNs"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseVPNReplication([]byte(tc.in))
			if err == nil {
				t.Fatalf("expected an error containing %q", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to mention %q", err, tc.want)
			}
		})
	}
}

// TestParseVPNReplicationIgnoresTheEchoedPrompt proves a trailing CLI prompt is not read
// as a VPN. It is too short to reach the flag columns, which is how it is told apart --
// and in the real capture there is a VPN NAMED after the router, so matching the prompt
// by its text would have dropped a real row.
func TestParseVPNReplicationIgnoresTheEchoedPrompt(t *testing.T) {
	got, err := ParseVPNReplication([]byte(showVPNReplication + "\nxps-ps-01> "))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(got) != 6 {
		t.Errorf("got %d VPNs, want 6: the echoed prompt is not a row", len(got))
	}
	if v, ok := got["xps-ps-01"]; !ok || v.Role != RoleStandby {
		t.Errorf("the VPN named after the router is a real row and must survive: %+v", v)
	}
}

// TestParseVPNReplicationIgnoresRepeatedHeaders proves a paginated report does not
// produce a phantom VPN.
//
// Left in, the repeated column RULE parses as a row: its name column trims to a run of
// dashes, and its flag columns are literal `-`, which decode to n/a for both. The result
// is a VPN named "--------------------------------" in the map, reported as existing but
// not replicating -- a row no caller asked about and no broker has.
func TestParseVPNReplicationIgnoresRepeatedHeaders(t *testing.T) {
	const paginated = `
Message VPN                      A C B R Q S M T
-------------------------------- - - - - - - - -
default                          U A - D D N Y A

Message VPN                      A C B R Q S M T
-------------------------------- - - - - - - - -
vpn-01                           U S U - - - Y A
`
	got, err := ParseVPNReplication([]byte(paginated))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d VPNs, want 2 (default and vpn-01): %+v", len(got), got)
	}
	for name := range got {
		if strings.HasPrefix(name, "--") {
			t.Errorf("the repeated column rule parsed as a VPN named %q", name)
		}
		if name == "Message VPN" {
			t.Error("the repeated header parsed as a VPN")
		}
	}
}

// TestParseVPNReplicationToleratesALongPrompt is the other half of the prompt story.
//
// TestParseVPNReplicationIgnoresTheEchoedPrompt covers a prompt too SHORT to reach the
// flag columns. A broker whose router name is longer than the name column produces one
// that DOES reach them, and without the gutter check its characters decode as flags --
// or more often fail to, turning an ordinary successful report into a hard parse error.
// The gutter -- the space that always precedes a flag column in a padded row -- is what
// tells the two apart.
func TestParseVPNReplicationToleratesALongPrompt(t *testing.T) {
	long := showVPNReplication + "\nsolace-prod-primary-frankfurt-dc-a1> "
	got, err := ParseVPNReplication([]byte(long))
	if err != nil {
		t.Fatalf("a long echoed prompt must not fail the parse: %v", err)
	}
	if len(got) != 6 {
		t.Errorf("got %d VPNs, want 6: the prompt is not a row: %+v", len(got), got)
	}
}

// --- real captures ------------------------------------------------------------
//
// semp/ is gitignored, so these SKIP when the capture is absent rather than failing a
// fresh clone. They assert the SHAPE the embedded fixtures above claim to mirror, which
// is what stops the transcription drifting away from the reports it was copied from.

func loadShowReplication(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "semp", name))
	if err != nil {
		t.Skipf("%s not available (%v)", name, err)
	}
	return raw
}

// TestRealApplianceShowReplicationParses proves the embedded appliance fixture mirrors
// the real report: one host, and a port for each of the three transports.
func TestRealApplianceShowReplicationParses(t *testing.T) {
	got, err := ParseShowReplication(BrokerAppliance, loadShowReplication(t, "show-replication-appl.out"))
	if err != nil {
		t.Fatalf("parse the real appliance report: %v", err)
	}
	if got.VirtualRouterName == "" {
		t.Error("the real appliance report names a mate")
	}
	if len(got.Endpoints) != 3 {
		t.Errorf("got %d endpoints, want 3 (one per transport): %+v", len(got.Endpoints), got.Endpoints)
	}
	hosts := map[string]bool{}
	for _, e := range got.Endpoints {
		hosts[e.Host] = true
	}
	if len(hosts) != 1 {
		t.Errorf("an appliance has ONE Connect-Via, got %d hosts", len(hosts))
	}
	// It must also re-render, which is the property the renderer's refusals protect.
	if _, err := RenderMate(BrokerAppliance, got); err != nil {
		t.Errorf("a real appliance report must re-render in its own grammar: %v", err)
	}
}

// TestRealSoftwareShowReplicationParses proves the embedded software fixture mirrors the
// real report: two ssl addresses on one port, and no plain-text or compressed endpoint
// despite both headings being printed.
func TestRealSoftwareShowReplicationParses(t *testing.T) {
	got, err := ParseShowReplication(BrokerSoftware, loadShowReplication(t, "show-replication-sw.out"))
	if err != nil {
		t.Fatalf("parse the real software report: %v", err)
	}
	if len(got.Endpoints) != 2 {
		t.Fatalf("got %d endpoints, want 2 -- the empty headings must not become endpoints: %+v",
			len(got.Endpoints), got.Endpoints)
	}
	for _, e := range got.Endpoints {
		if e.Transport != config.TransportSSL {
			t.Errorf("endpoint %+v: the real report configures only ssl", e)
		}
	}
	if got.Endpoints[0].Port != got.Endpoints[1].Port {
		t.Error("the real report's two ssl addresses share one port, which is what makes the endpoint " +
			"list a list of triples rather than hosts times ports")
	}
	if _, err := RenderMate(BrokerSoftware, got); err != nil {
		t.Errorf("a real software report must re-render in its own grammar: %v", err)
	}
	// The same set is unrenderable on an appliance, since the two addresses differ.
	if _, err := RenderMate(BrokerAppliance, got); err == nil {
		t.Error("two distinct hosts must be refused by the appliance renderer")
	}
}

// TestRealVPNReplicationParses proves the embedded flag table mirrors the real report,
// and in particular that the real report really does carry VPN names with spaces -- the
// premise the whole fixed-width reader rests on. SKIPS when absent.
func TestRealVPNReplicationParses(t *testing.T) {
	got, err := ParseVPNReplication(loadShowReplication(t, "show-message-vpn-replication.out"))
	if err != nil {
		t.Fatalf("parse the real report: %v", err)
	}
	spaced := 0
	for name := range got {
		if strings.Contains(name, " ") {
			spaced++
		}
	}
	if spaced == 0 {
		t.Error("the real report carries no VPN name with a space; the fixed-width reader exists " +
			"because it does, so either the capture changed or the premise needs rechecking")
	}
	if _, ok := got["default"]; !ok {
		t.Error("every broker has a default VPN, so it must appear in the table")
	}
}

// TestRealApplianceCaptureMateRoundTrips closes the loop against the `show
// current-config` capture: the report and the configuration must describe the same mate,
// so parsing the report and rendering it must reproduce the capture's own lines.
func TestRealApplianceCaptureMateRoundTrips(t *testing.T) {
	report := loadShowReplication(t, "show-replication-appl.out")
	c := loadApplianceCapture(t)

	var want []string
	for _, b := range c.Blocks {
		if b.Section != "Configure Replication" {
			continue
		}
		for _, l := range b.Lines {
			if strings.HasPrefix(strings.TrimSpace(l), "replication mate ") {
				want = append(want, strings.TrimSpace(l))
			}
		}
	}
	if len(want) == 0 {
		t.Skip("this appliance capture configures no replication mate")
	}

	parsed, err := ParseShowReplication(BrokerAppliance, report)
	if err != nil {
		t.Fatalf("parse the real appliance report: %v", err)
	}
	got, err := RenderMate(BrokerAppliance, parsed)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("rendering the report does not reproduce the capture's own lines\n"+
			"--- rendered ---\n%s\n--- capture ---\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

package broker

import (
	"context"
	"strings"
	"testing"

	"solace/internal/config"
)

// sempmate_test.go covers the SEMP v1 leg. Every request and reply shape asserted here
// was read off semp/semp-rpc-soltr.xsd and semp-rpc-reply-soltr.xsd, so a test failing
// means the code drifted from the broker's own schema -- not that a guess changed.

// sempReply wraps a body in the HTTP envelope `curl -is` returns, which is what sempCurl
// hands back.
func sempReply(body string) string {
	return "HTTP/1.1 200 OK\r\nContent-Type: text/xml\r\n\r\n" + body
}

const sempOKResult = `<rpc-reply semp-version="broker/10.26.0.8827"><execute-result code="ok"/></rpc-reply>`

// sempSite is a site reached over SEMP, with TLS and its own password.
func sempSite() config.ReplSite {
	return config.ReplSite{
		VirtualRouterName: "v:sol-dr",
		RouterNames:       []string{"sol-dr1"},
		Via: config.ReplVia{SEMP: &config.ReplViaSEMP{
			Host: "10.0.1.1", Port: 1943, TLS: true, Insecure: true,
		}},
	}
}

// newSEMPTestMate builds the channel over a recording transport whose reply is fixed.
func newSEMPTestMate(t *testing.T, reply string) (MateChannel, *fakeTransport) {
	t.Helper()
	ft := &fakeTransport{responder: func(_ config.Role, _ []string, _ []byte) ([]byte, error) {
		return []byte(sempReply(reply)), nil
	}}
	o, _ := newTestOps(t, localCfg("true"), ft)
	m, err := o.NewSEMPMate(sempSite(), "dr-site-password")
	if err != nil {
		t.Fatalf("NewSEMPMate: %v", err)
	}
	return m, ft
}

// TestSEMPMateURLAndCredential proves the channel dials the SITE's endpoint with the
// SITE's password -- not this deployment's.
//
// A DR mate is a different broker. Sending semp.adminPass to it would fail at best and,
// if the two happened to match, hide the misconfiguration until the day they diverged.
func TestSEMPMateURLAndCredential(t *testing.T) {
	m, ft := newSEMPTestMate(t, sempOKResult)
	if err := m.Preflight(context.Background()); err != nil {
		t.Fatalf("Preflight: %v", err)
	}
	calls := curlCalls(ft)
	if len(calls) != 1 {
		t.Fatalf("made %d calls, want 1", len(calls))
	}
	if got := calls[0].argv[len(calls[0].argv)-1]; got != "https://10.0.1.1:1943/SEMP" {
		t.Errorf("url = %q, want the site's own https endpoint", got)
	}
	if !strings.Contains(calls[0].stdin, `user = "admin:dr-site-password"`) {
		t.Errorf("stdin must carry the SITE's credential: %q", calls[0].stdin)
	}
	if strings.Contains(calls[0].stdin, "adminpw") {
		t.Errorf("the LOCAL broker's password must never reach a DR site: %q", calls[0].stdin)
	}
	if !strings.Contains(calls[0].stdin, "insecure") {
		t.Errorf("insecure: true must reach curl: %q", calls[0].stdin)
	}
	assertNoPasswordInArgv(t, ft, "dr-site-password")
}

// TestSEMPMateRequestShapes pins every request body against the request XSD.
//
// These are the paths a wrong guess would break invisibly: the write has NO <admin>
// wrapper (rpc > admin has no message-vpn child at all), vpn-name comes FIRST because the
// schema puts it in an xs:sequence ahead of the sub-command choice, and the role operand
// is an EMPTY element rather than text.
func TestSEMPMateRequestShapes(t *testing.T) {
	cases := []struct {
		name string
		run  func(MateChannel) error
		want string
	}{
		{"show replication", func(c MateChannel) error {
			_, err := c.ShowReplication(context.Background())
			return err
		}, "<rpc><show><replication/></show></rpc>"},
		{"show message-vpn replication", func(c MateChannel) error {
			_, err := c.ReadRoles(context.Background(), nil)
			return err
		}, "<rpc><show><message-vpn><vpn-name>*</vpn-name><replication/></message-vpn></show></rpc>"},
		{"show redundancy", func(c MateChannel) error {
			_, err := c.PrimaryActive(context.Background())
			return err
		}, "<rpc><show><redundancy/></show></rpc>"},
		{"set role active", func(c MateChannel) error {
			return c.SetRole(context.Background(), "ORDERS", RoleActive)
		}, "<rpc><message-vpn><vpn-name>ORDERS</vpn-name><replication><state><active/></state>" +
			"</replication></message-vpn></rpc>"},
		{"set role standby", func(c MateChannel) error {
			return c.SetRole(context.Background(), "ORDERS", RoleStandby)
		}, "<rpc><message-vpn><vpn-name>ORDERS</vpn-name><replication><state><standby/></state>" +
			"</replication></message-vpn></rpc>"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, ft := newSEMPTestMate(t, sempOKResult+emptyReplicationReply)
			if err := tc.run(m); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			calls := curlCalls(ft)
			body := calls[len(calls)-1].stdin
			if !strings.Contains(body, tc.want) {
				t.Errorf("request body:\n%s\nmust contain\n%s", body, tc.want)
			}
			if strings.Contains(body, "<admin>") {
				t.Error("there is no <admin> wrapper on any replication path; rpc > admin has no " +
					"message-vpn child at all")
			}
		})
	}
}

// emptyReplicationReply is a well-formed reply with no mate, so a shape test can run
// without also exercising the parser.
const emptyReplicationReply = ""

// TestSEMPMateEscapesTheVPNName proves a hand-built body is escaped.
//
// The bodies are assembled as strings because the schema's element ORDER is load-bearing,
// which makes escaping this code's job. A real VPN name is operator-chosen, and one
// carrying `&` or `<` would otherwise produce a malformed document the broker rejects --
// or, worse, a different one than intended.
func TestSEMPMateEscapesTheVPNName(t *testing.T) {
	m, ft := newSEMPTestMate(t, sempOKResult)
	if err := m.SetRole(context.Background(), `A & B <x>`, RoleActive); err != nil {
		t.Fatalf("SetRole: %v", err)
	}
	body := curlCalls(ft)[0].stdin
	if strings.Contains(body, "<vpn-name>A & B <x></vpn-name>") {
		t.Errorf("the VPN name reached the body unescaped:\n%s", body)
	}
	if !strings.Contains(body, "&amp;") {
		t.Errorf("the ampersand must be escaped:\n%s", body)
	}
}

// TestSEMPMateChecksExecuteResult proves a 200 OK is not taken as success.
//
// `execute-result` is a SIBLING of <rpc> under <rpc-reply> and carries its verdict in the
// `code` ATTRIBUTE, so a reply can be HTTP 200 and still say fail -- which is exactly the
// shape a rejected role change takes.
func TestSEMPMateChecksExecuteResult(t *testing.T) {
	const failed = `<rpc-reply semp-version="broker/10.26.0.8827">` +
		`<execute-result code="fail" reason="not in a valid state"/></rpc-reply>`
	m, _ := newSEMPTestMate(t, failed)
	err := m.SetRole(context.Background(), "ORDERS", RoleActive)
	if err == nil {
		t.Fatal("a reply saying code=fail must fail the command despite HTTP 200")
	}
	if !strings.Contains(err.Error(), "not in a valid state") {
		t.Errorf("the broker's own reason must survive: %v", err)
	}
}

// TestSEMPMateReadsBothMateShapes is the advantage this leg has over the CLI one.
//
// `show replication` over the CLI prints two different LAYOUTS and needs the broker type
// to choose a parser. Its SEMP reply models both as sibling structures under one <mate>:
// an appliance fills <connect-via> with a <connect-ports> trio, software fills
// <mate-list>. Whichever is populated says which broker it is, so nothing is discovered.
func TestSEMPMateReadsBothMateShapes(t *testing.T) {
	t.Run("appliance", func(t *testing.T) {
		const reply = `<rpc-reply semp-version="broker/10.26.0.8827"><rpc><show><replication>
<interface></interface>
<mate><router-name>v:appl-b</router-name>
<connect-via><address>10.0.0.9</address>
<connect-ports><uncompressed>55555</uncompressed><compressed>55003</compressed><ssl>55443</ssl></connect-ports>
</connect-via></mate>
</replication></show></rpc><execute-result code="ok"/></rpc-reply>`
		m, _ := newSEMPTestMate(t, reply)
		got, err := m.ShowReplication(context.Background())
		if err != nil {
			t.Fatalf("ShowReplication: %v", err)
		}
		if got.VirtualRouterName != "v:appl-b" {
			t.Errorf("mate = %q", got.VirtualRouterName)
		}
		if !SameEndpoints(got.Endpoints, mateAppliance().Endpoints) {
			t.Errorf("endpoints = %+v, want %+v", got.Endpoints, mateAppliance().Endpoints)
		}
	})

	t.Run("software", func(t *testing.T) {
		const reply = `<rpc-reply semp-version="broker/10.26.0.8827"><rpc><show><replication>
<mate><router-name>v:broker-b</router-name>
<mate-list>
<connect-via><address>broker-b:55443</address><transport>ssl</transport></connect-via>
<connect-via><address>broker-b.cluster_default:55443</address><transport>ssl</transport></connect-via>
</mate-list></mate>
</replication></show></rpc><execute-result code="ok"/></rpc-reply>`
		m, _ := newSEMPTestMate(t, reply)
		got, err := m.ShowReplication(context.Background())
		if err != nil {
			t.Fatalf("ShowReplication: %v", err)
		}
		if !SameEndpoints(got.Endpoints, mateSoftware().Endpoints) {
			t.Errorf("endpoints = %+v, want %+v", got.Endpoints, mateSoftware().Endpoints)
		}
	})

	// A port of zero is the transport not being configured, not port 0 -- the field is
	// an unsignedInt with no way to say "absent".
	t.Run("unconfigured appliance transports are not endpoints", func(t *testing.T) {
		const reply = `<rpc-reply><rpc><show><replication><mate><router-name>v:x</router-name>
<connect-via><address>10.0.0.9</address>
<connect-ports><uncompressed>0</uncompressed><compressed>0</compressed><ssl>55443</ssl></connect-ports>
</connect-via></mate></replication></show></rpc><execute-result code="ok"/></rpc-reply>`
		m, _ := newSEMPTestMate(t, reply)
		got, err := m.ShowReplication(context.Background())
		if err != nil {
			t.Fatalf("ShowReplication: %v", err)
		}
		if len(got.Endpoints) != 1 || got.Endpoints[0].Transport != config.TransportSSL {
			t.Errorf("endpoints = %+v, want only the ssl one", got.Endpoints)
		}
	})
}

// TestSEMPMateReadsRolesTriState proves the reply enumerations decode, including `n/a`.
//
// Folding n/a into standby would let a switch promote against a VPN it cannot see, which
// is why both fields are tri-state here as they are on the CLI leg.
func TestSEMPMateReadsRolesTriState(t *testing.T) {
	const reply = `<rpc-reply><rpc><show><message-vpn><replication><message-vpns>
<message-vpn><vpn-name>default</vpn-name><admin-state>enabled</admin-state><config-state>active</config-state></message-vpn>
<message-vpn><vpn-name>A VPN WITH SPACES</vpn-name><admin-state>shutdown</admin-state><config-state>standby</config-state></message-vpn>
<message-vpn><vpn-name>#config-sync</vpn-name><admin-state>n/a</admin-state><config-state>n/a</config-state></message-vpn>
</message-vpns></replication></message-vpn></show></rpc><execute-result code="ok"/></rpc-reply>`
	m, _ := newSEMPTestMate(t, reply)
	got, err := m.ReadRoles(context.Background(), nil)
	if err != nil {
		t.Fatalf("ReadRoles: %v", err)
	}
	want := map[string]VPNRepl{
		"default":           {Admin: AdminEnabled, Role: RoleActive},
		"A VPN WITH SPACES": {Admin: AdminShutdown, Role: RoleStandby},
		"#config-sync":      {Admin: AdminNA, Role: RoleNA},
	}
	for name, w := range want {
		if got[name] != w {
			t.Errorf("%q = %+v, want %+v", name, got[name], w)
		}
	}
	// A name with spaces is element TEXT here, so unlike the CLI's fixed-width table it
	// needs no column arithmetic -- but it must still survive.
	if _, ok := got["A VPN WITH SPACES"]; !ok {
		t.Error("a VPN name with spaces must survive the XML decode")
	}

	t.Run("narrows to the listed VPNs", func(t *testing.T) {
		m, _ := newSEMPTestMate(t, reply)
		only, err := m.ReadRoles(context.Background(), []string{"default"})
		if err != nil {
			t.Fatalf("ReadRoles: %v", err)
		}
		if len(only) != 1 {
			t.Errorf("got %+v, want only the VPN asked about", only)
		}
	})

	t.Run("an unknown enumeration fails loudly", func(t *testing.T) {
		bad := strings.Replace(reply, "<config-state>active</config-state>",
			"<config-state>sideways</config-state>", 1)
		m, _ := newSEMPTestMate(t, bad)
		if _, err := m.ReadRoles(context.Background(), nil); err == nil {
			t.Error("a config-state outside the enumeration must not be guessed at")
		}
	})
}

// TestSEMPMatePrimaryActive proves the activity read, which is per VIRTUAL ROUTER rather
// than per node -- a chassis reports one activity for its primary virtual router and
// another for its backup.
func TestSEMPMatePrimaryActive(t *testing.T) {
	rd := func(configStatus, primaryActivity string) string {
		return `<rpc-reply><rpc><show><redundancy>` +
			`<config-status>` + configStatus + `</config-status>` +
			`<virtual-routers><primary><status><activity>` + primaryActivity +
			`</activity></status></primary></virtual-routers>` +
			`</redundancy></show></rpc><execute-result code="ok"/></rpc-reply>`
	}
	cases := []struct {
		name  string
		reply string
		want  bool
	}{
		{"primary holds activity", rd("Enabled", "Local Active"), true},
		{"backup holds activity", rd("Enabled", "Mate Active"), false},
		{"standalone has no HA group", rd("Shutdown", ""), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, _ := newSEMPTestMate(t, tc.reply)
			got, err := m.PrimaryActive(context.Background())
			if err != nil {
				t.Fatalf("PrimaryActive: %v", err)
			}
			if got != tc.want {
				t.Errorf("PrimaryActive = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestSEMPMateRejectsANonSEMPSite proves a site reached another way is refused here
// rather than silently driven over the wrong mechanism.
func TestSEMPMateRejectsANonSEMPSite(t *testing.T) {
	o, _ := newTestOps(t, localCfg("true"), &fakeTransport{})
	s := sempSite()
	s.Via = config.ReplVia{Kubernetes: &config.ReplViaKube{
		Command: config.Command{"kubectl"}, Namespace: "n", Name: "b"}}
	if _, err := o.NewSEMPMate(s, "p"); err == nil {
		t.Error("a site declaring via.kubernetes must not be built as a SEMP channel")
	}
}

// TestSEMPMateSetRoleRefusesNA proves n/a is a reading, not a state to write -- the same
// guard the CLI leg has, so the two channels cannot diverge on it.
func TestSEMPMateSetRoleRefusesNA(t *testing.T) {
	m, ft := newSEMPTestMate(t, sempOKResult)
	if err := m.SetRole(context.Background(), "ORDERS", RoleNA); err == nil {
		t.Fatal("n/a must not be writable")
	}
	if len(curlCalls(ft)) != 0 {
		t.Error("a refused role must not reach the broker at all")
	}
}

// TestSEMPMatePlainTextURL proves tls: false produces http rather than https, which is
// the declared-not-inferred choice the site block exists to make explicit.
func TestSEMPMatePlainTextURL(t *testing.T) {
	s := sempSite()
	s.Via.SEMP.TLS = false
	s.Via.SEMP.Insecure = false
	s.Via.SEMP.Port = 8080
	ft := &fakeTransport{responder: func(_ config.Role, _ []string, _ []byte) ([]byte, error) {
		return []byte(sempReply(sempOKResult)), nil
	}}
	o, _ := newTestOps(t, localCfg("true"), ft)
	m, err := o.NewSEMPMate(s, "p")
	if err != nil {
		t.Fatalf("NewSEMPMate: %v", err)
	}
	if err := m.Preflight(context.Background()); err != nil {
		t.Fatalf("Preflight: %v", err)
	}
	calls := curlCalls(ft)
	if got := calls[0].argv[len(calls[0].argv)-1]; got != "http://10.0.1.1:8080/SEMP" {
		t.Errorf("url = %q, want plain http on the declared port", got)
	}
	if strings.Contains(calls[0].stdin, "insecure") {
		t.Errorf("a plaintext leg needs no TLS options: %q", calls[0].stdin)
	}
}

package broker

import (
	"context"
	"errors"
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

// --- failure and boundary paths -------------------------------------------------------

// covSEMPMate builds the channel over a recording transport whose reply the test scripts
// per request. The responder is handed the curl config the call put on stdin, so it can
// answer by what was asked -- which is how one test makes the first of two requests succeed
// and the second fail.
func covSEMPMate(t *testing.T, respond func(stdin string) ([]byte, error)) (MateChannel, *fakeTransport) {
	t.Helper()
	ft := &fakeTransport{responder: func(_ config.Role, _ []string, stdin []byte) ([]byte, error) {
		return respond(string(stdin))
	}}
	o, _ := newTestOps(t, localCfg("true"), ft)
	m, err := o.NewSEMPMate(sempSite(), "dr-site-password")
	if err != nil {
		t.Fatalf("NewSEMPMate: %v", err)
	}
	return m, ft
}

// TestSEMPMateDescribeNamesTheSiteAndEndpoint proves the channel can say which site it
// reaches and where. A two-site flow reports errors through this, so an operator reading
// one knows which side of the pair it is about.
func TestSEMPMateDescribeNamesTheSiteAndEndpoint(t *testing.T) {
	m, _ := newSEMPTestMate(t, sempOKResult)
	if got, want := m.Describe(), "site v:sol-dr over SEMP at 10.0.1.1:1943"; got != want {
		t.Errorf("Describe = %q, want %q", got, want)
	}
}

// TestSEMPMateEveryCallSurfacesATransportFailure proves a curl that cannot run -- the
// container is gone, the exec was refused -- reaches the caller from every operation with
// its cause intact and the endpoint named, and without the site's password. The credential
// rides stdin precisely so that no error, log line or argv can carry it, and a failure path
// is where a careless message would break that.
func TestSEMPMateEveryCallSurfacesATransportFailure(t *testing.T) {
	boom := errors.New("connection refused")
	calls := []struct {
		name string
		run  func(MateChannel) error
	}{
		{"Preflight", func(c MateChannel) error { return c.Preflight(context.Background()) }},
		{"ShowReplication", func(c MateChannel) error {
			_, err := c.ShowReplication(context.Background())
			return err
		}},
		{"ReadRoles", func(c MateChannel) error {
			_, err := c.ReadRoles(context.Background(), nil)
			return err
		}},
		{"PrimaryActive", func(c MateChannel) error {
			_, err := c.PrimaryActive(context.Background())
			return err
		}},
		{"Status", func(c MateChannel) error {
			_, err := c.Status(context.Background())
			return err
		}},
		{"SetRole", func(c MateChannel) error {
			return c.SetRole(context.Background(), "ORDERS", RoleActive)
		}},
	}
	for _, tc := range calls {
		t.Run(tc.name, func(t *testing.T) {
			m, ft := covSEMPMate(t, func(string) ([]byte, error) { return nil, boom })
			err := tc.run(m)
			if !errors.Is(err, boom) {
				t.Fatalf("err = %v, want the transport cause kept", err)
			}
			if !strings.Contains(err.Error(), "https://10.0.1.1:1943/SEMP") {
				t.Errorf("err = %v, want the endpoint named", err)
			}
			if strings.Contains(err.Error(), "dr-site-password") {
				t.Errorf("err = %v, the site's password must never reach an error", err)
			}
			assertNoPasswordInArgv(t, ft, "dr-site-password")
		})
	}
}

// TestSEMPMateRefusesAReplyThatIsNotASuccessfulSEMPResponse covers the three shapes of a
// reply that must not be read as an answer: a non-2xx status (the credential, host, port or
// a firewall is wrong), no HTTP status at all (something other than a broker answered), and
// a 200 whose body is not the XML the schema describes. Each says what to check, and the
// body of a failed reply is never echoed -- a broker may repeat what it was sent.
func TestSEMPMateRefusesAReplyThatIsNotASuccessfulSEMPResponse(t *testing.T) {
	cases := []struct {
		name  string
		reply string
		want  string
	}{
		{"a non-2xx status", "HTTP/1.1 401 Unauthorized\r\nContent-Type: text/plain\r\n\r\nbad credentials",
			"HTTP/1.1 401 Unauthorized"},
		{"no HTTP status at all", "something that is not a broker", "<no HTTP response from mate>"},
		{"a 200 whose body is not XML", sempReply("this is not xml <<<"), "could not parse"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, _ := covSEMPMate(t, func(string) ([]byte, error) { return []byte(tc.reply), nil })
			err := m.Preflight(context.Background())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want it to contain %q", err, tc.want)
			}
			if !strings.Contains(err.Error(), "cannot reach site v:sol-dr") {
				t.Errorf("err = %v, want the preflight to name what it could not reach", err)
			}
			if strings.Contains(err.Error(), "bad credentials") {
				t.Errorf("err = %v, the body of a failed reply must not be echoed", err)
			}
		})
	}
}

// TestSEMPMateStatusIsBothReportsInOrder proves Status returns the mate report and then the
// per-VPN report, each as the broker sent it, for display only. They are two requests -- the
// first asks for `show replication`, the second for the per-VPN table -- and the order is the
// order an operator reads them in.
func TestSEMPMateStatusIsBothReportsInOrder(t *testing.T) {
	const mate = `<rpc-reply><rpc><show><replication><mate><router-name>v:sol-b</router-name></mate>` +
		`</replication></show></rpc><execute-result code="ok"/></rpc-reply>`
	const vpns = `<rpc-reply><rpc><show><message-vpn><replication><message-vpns>` +
		`<message-vpn><vpn-name>ORDERS</vpn-name></message-vpn></message-vpns></replication>` +
		`</message-vpn></show></rpc><execute-result code="ok"/></rpc-reply>`
	m, ft := covSEMPMate(t, func(stdin string) ([]byte, error) {
		if strings.Contains(stdin, "message-vpn") {
			return []byte(sempReply(vpns)), nil
		}
		return []byte(sempReply(mate)), nil
	})

	got, err := m.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if got != mate+"\n"+vpns {
		t.Errorf("Status = %q, want the mate report, a newline, then the per-VPN report", got)
	}
	if n := len(curlCalls(ft)); n != 2 {
		t.Errorf("Status made %d requests, want exactly the two reports", n)
	}
}

// TestSEMPMateStatusFailsOnEitherReport proves a failed report fails Status outright and
// returns no text. Handing back one half of a two-part picture would show an operator a mate
// with no VPN table and nothing to say the second request never succeeded. A failed first
// request also means the second is never sent.
func TestSEMPMateStatusFailsOnEitherReport(t *testing.T) {
	const ok = `<rpc-reply><execute-result code="ok"/></rpc-reply>`
	const refused = "HTTP/1.1 500 Internal Server Error\r\n\r\n"
	cases := []struct {
		name      string
		failMatch string
		wantCalls int
	}{
		{"the mate report fails", "<show><replication/>", 1},
		{"the per-VPN report fails", "<vpn-name>*</vpn-name>", 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, ft := covSEMPMate(t, func(stdin string) ([]byte, error) {
				if strings.Contains(stdin, tc.failMatch) {
					return []byte(refused), nil
				}
				return []byte(sempReply(ok)), nil
			})
			got, err := m.Status(context.Background())
			if err == nil || !strings.Contains(err.Error(), "HTTP/1.1 500") {
				t.Fatalf("Status = %q, %v; want the failed report's status in the error", got, err)
			}
			if got != "" {
				t.Errorf("Status returned %q alongside its error, want no partial report", got)
			}
			if n := len(curlCalls(ft)); n != tc.wantCalls {
				t.Errorf("Status made %d requests, want %d", n, tc.wantCalls)
			}
		})
	}
}

// TestSEMPMateShowReplicationRefusesAReplyItCannotRead covers the replies that parse as XML
// but do not describe a mate this tool can use. Each fails naming the site and what was
// wrong, rather than yielding a mate with a silently missing endpoint -- which a caller
// would converge a broker onto -- and one case is NOT a failure: an appliance connect-via
// with no address contributes nothing and is skipped.
func TestSEMPMateShowReplicationRefusesAReplyItCannotRead(t *testing.T) {
	wrap := func(mate string) string {
		return `<rpc-reply><rpc><show><replication><mate>` + mate +
			`</mate></replication></show></rpc><execute-result code="ok"/></rpc-reply>`
	}
	t.Run("a port that is not a number", func(t *testing.T) {
		const mate = `<router-name>v:appl-b</router-name><connect-via><address>10.0.0.9</address>` +
			`<connect-ports><ssl>not-a-port</ssl></connect-ports></connect-via>`
		m, _ := newSEMPTestMate(t, wrap(mate))
		_, err := m.ShowReplication(context.Background())
		if err == nil || !strings.Contains(err.Error(), "reading `show replication`") {
			t.Errorf("err = %v, want the unreadable reply named", err)
		}
	})

	cases := []struct {
		name string
		mate string
		want string
	}{
		{"a transport outside the vocabulary",
			`<router-name>v:b</router-name><mate-list><connect-via><address>b:55443</address>` +
				`<transport>encrypted</transport></connect-via></mate-list>`,
			`unknown replication transport keyword "encrypted"`},
		{"an address that is not host:port",
			`<router-name>v:b</router-name><mate-list><connect-via><address>broker-b</address>` +
				`<transport>ssl</transport></connect-via></mate-list>`,
			"which is not host:port"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, _ := newSEMPTestMate(t, wrap(tc.mate))
			_, err := m.ShowReplication(context.Background())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want it to contain %q", err, tc.want)
			}
			if !strings.Contains(err.Error(), "site v:sol-dr") {
				t.Errorf("err = %v, want the site named", err)
			}
		})
	}

	t.Run("an appliance connect-via with no address is skipped", func(t *testing.T) {
		const mate = `<router-name>v:appl-b</router-name><connect-via><address> </address>` +
			`<connect-ports><ssl>55443</ssl></connect-ports></connect-via>`
		m, _ := newSEMPTestMate(t, wrap(mate))
		got, err := m.ShowReplication(context.Background())
		if err != nil {
			t.Fatalf("ShowReplication: %v", err)
		}
		if got.VirtualRouterName != "v:appl-b" || len(got.Endpoints) != 0 {
			t.Errorf("got %+v, want the mate's name and no endpoint for an address that is blank", got)
		}
	})
}

// TestSEMPMateReadRolesRefusesAnUnknownAdminState is the admin-state half of the tri-state
// rule, which TestSEMPMateReadsRolesTriState checks for the config-state only. A value
// outside enabled, shutdown and n/a is a refusal naming the VPN, never a guess: reading it
// as shutdown would let a switch skip a VPN it cannot see, and as enabled would let it
// promote against one that is not replicating.
func TestSEMPMateReadRolesRefusesAnUnknownAdminState(t *testing.T) {
	const reply = `<rpc-reply><rpc><show><message-vpn><replication><message-vpns>` +
		`<message-vpn><vpn-name>ORDERS</vpn-name><admin-state>sideways</admin-state>` +
		`<config-state>active</config-state></message-vpn>` +
		`</message-vpns></replication></message-vpn></show></rpc><execute-result code="ok"/></rpc-reply>`
	m, _ := newSEMPTestMate(t, reply)
	got, err := m.ReadRoles(context.Background(), nil)
	if err == nil {
		t.Fatalf("ReadRoles = %+v, want a refusal of an admin-state outside the enumeration", got)
	}
	for _, want := range []string{`message-vpn "ORDERS"`, `admin-state "sideways"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, want it to mention %q", err, want)
		}
	}
	if got != nil {
		t.Errorf("ReadRoles returned %+v alongside its error, want nothing to act on", got)
	}
}

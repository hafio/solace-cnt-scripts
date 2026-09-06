package broker

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"solace/internal/config"
)

// These tests cover the SEMP-over-HTTP mate control channel (semp.go) in
// isolation: port resolution, curl config rendering, reply parsing, and the
// preflight/revert-activity request shapes over the shared fakeTransport. The
// coordinated flows that consume the channel are covered in
// verify_local_test.go.

// --- port resolution ---------------------------------------------------------

// serverCert is the pair that makes the broker serve a TLS SEMP listener at all.
// Port resolution is gated on it, so every TLS-expecting case has to set it --
// which is the point: an env file without it gets plaintext.
var serverCert = config.TLS{Cert: "server.crt", CertKey: "server.key"}

// TestSempPortHostMode pins that host networking exposes the broker's own ports
// directly on the mate host's interface, so unlike bridge mode there is no
// mapping to be missing -- the only question is whether the broker has a TLS
// listener to reach.
func TestSempPortHostMode(t *testing.T) {
	for _, mode := range []string{"", "host"} {
		cfg := &config.Config{TLS: serverCert, Docker: config.DockerConfig{Network: config.Network{Mode: mode}}}
		got, tls, err := sempPort(cfg, config.Docker)
		if err != nil || got != 1943 || !tls {
			t.Errorf("sempPort(mode=%q, cert set) = %d,%v,%v want 1943,true", mode, got, tls, err)
		}
	}
}

// TestSempPortWithoutCertificateStaysPlaintext is the guard against breaking an
// HA group that works today. Preferring TLS is only correct when the broker
// actually serves it: without tls.cert/tls.certKey the 1943 listener never comes
// up, so aiming the mate channel at it would fail every coordinated redundancy
// and leader step on a plaintext deployment that was working fine.
func TestSempPortWithoutCertificateStaysPlaintext(t *testing.T) {
	t.Run("host mode", func(t *testing.T) {
		cfg := &config.Config{Docker: config.DockerConfig{Network: config.Network{Mode: "host"}}}
		got, tls, err := sempPort(cfg, config.Docker)
		if err != nil || got != 8080 || tls {
			t.Errorf("sempPort(host, no cert) = %d,%v,%v want 8080,false", got, tls, err)
		}
	})
	t.Run("bridge mapping both ports", func(t *testing.T) {
		cfg := &config.Config{Podman: config.PodmanConfig{Network: config.Network{
			Mode: "bridge", Ports: []string{"18080:8080", "18943:1943"}}}}
		got, tls, err := sempPort(cfg, config.Podman)
		if err != nil || got != 18080 || tls {
			t.Errorf("sempPort(bridge, no cert) = %d,%v,%v want 18080,false -- 1943 is mapped but nothing "+
				"listens on it without a certificate", got, tls, err)
		}
	})
	t.Run("bridge mapping only the TLS port says what is missing", func(t *testing.T) {
		cfg := &config.Config{Podman: config.PodmanConfig{Network: config.Network{
			Mode: "bridge", Ports: []string{"18943:1943"}}}}
		_, _, err := sempPort(cfg, config.Podman)
		if err == nil {
			t.Fatal("mapping only 1943 with no certificate leaves no reachable SEMP port")
		}
		if !strings.Contains(err.Error(), "tls.cert") {
			t.Errorf("error = %v, want it to name the missing certificate rather than advise mapping 1943", err)
		}
	})
}

// TestSempPortBridgePrefersTLS pins that a bridge mapping exposing 1943 wins
// even when 8080 is also mapped -- TLS is tried first.
func TestSempPortBridgePrefersTLS(t *testing.T) {
	cases := []struct {
		ports []string
		want  int
	}{
		{[]string{"2222:2222", "18943:1943"}, 18943},
		{[]string{"1943:1943/tcp"}, 1943},
		{[]string{"127.0.0.1:18943:1943"}, 18943}, // host-IP-qualified publish
		{[]string{"18080:8080", "18943:1943"}, 18943},
	}
	for _, tc := range cases {
		cfg := &config.Config{TLS: serverCert, Podman: config.PodmanConfig{Network: config.Network{Mode: "bridge", Ports: tc.ports}}}
		got, tls, err := sempPort(cfg, config.Podman)
		if err != nil || got != tc.want || !tls {
			t.Errorf("sempPort(ports=%v) = %d,%v,%v want %d,true", tc.ports, got, tls, err, tc.want)
		}
	}
}

// TestSempPortBridgeFallsBackToPlaintext pins the last-resort branch: no TLS
// mapping, so the plaintext 8080 mapping is used instead.
func TestSempPortBridgeFallsBackToPlaintext(t *testing.T) {
	cases := []struct {
		ports []string
		want  int
	}{
		{[]string{"2222:2222", "18080:8080"}, 18080},
		{[]string{"8080:8080/tcp"}, 8080},
		{[]string{"127.0.0.1:18080:8080"}, 18080}, // host-IP-qualified publish
	}
	for _, tc := range cases {
		cfg := &config.Config{TLS: serverCert, Podman: config.PodmanConfig{Network: config.Network{Mode: "bridge", Ports: tc.ports}}}
		got, tls, err := sempPort(cfg, config.Podman)
		if err != nil || got != tc.want || tls {
			t.Errorf("sempPort(ports=%v) = %d,%v,%v want %d,false", tc.ports, got, tls, err, tc.want)
		}
	}
}

func TestSempPortBridgeRefusesNoMapping(t *testing.T) {
	cfg := &config.Config{Docker: config.DockerConfig{Network: config.Network{
		Mode:  "bridge",
		Ports: []string{"55555:55555", "8080:8080/udp"}, // no tcp SEMP mapping either way
	}}}
	_, _, err := sempPort(cfg, config.Docker)
	if err == nil || !strings.Contains(err.Error(), "network.ports") {
		t.Errorf("sempPort err = %v, want a loud refusal naming network.ports", err)
	}
}

// --- curl config rendering -----------------------------------------------------

func TestCurlConfigLineEscapesQuotes(t *testing.T) {
	cases := map[string]string{
		"plain":        "user = \"plain\"\n",
		`with "quote"`: "user = \"with \\\"quote\\\"\"\n",
		`back\slash`:   "user = \"back\\\\slash\"\n",
	}
	for value, want := range cases {
		if got := curlConfigLine("user", value); got != want {
			t.Errorf("curlConfigLine(user, %q) = %q, want %q", value, got, want)
		}
	}
}

// --- reply parsing -------------------------------------------------------------

func TestSempV1OKMatchesExecuteResultOK(t *testing.T) {
	replies := []string{
		`<rpc-reply><execute-result code="ok"/></rpc-reply>`,
		`<rpc-reply semp-version="soltr/10_0"><execute-result reason="" code="ok"/></rpc-reply>`,
	}
	for _, r := range replies {
		if !sempV1OK([]byte(r)) {
			t.Errorf("sempV1OK(%q) = false, want true", r)
		}
	}
}

func TestSempV1OKRejectsErrorReply(t *testing.T) {
	replies := []string{
		`<rpc-reply><execute-result code="fail" reason="denied"/></rpc-reply>`,
		`<rpc-reply>no execute-result at all</rpc-reply>`,
		``,
	}
	for _, r := range replies {
		if sempV1OK([]byte(r)) {
			t.Errorf("sempV1OK(%q) = true, want false", r)
		}
	}
}

func TestHTTPBodyStripsHeaders(t *testing.T) {
	cases := map[string]string{
		"HTTP/1.1 200 OK\r\nContent-Type: text/xml\r\n\r\n<rpc-reply/>": "<rpc-reply/>",
		"HTTP/1.1 200 OK\nContent-Type: text/xml\n\n<rpc-reply/>":       "<rpc-reply/>",
		"no header separator at all":                                    "no header separator at all",
	}
	for raw, want := range cases {
		if got := httpBody([]byte(raw)); got != want {
			t.Errorf("httpBody(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestLastHTTPStatus(t *testing.T) {
	raw := "HTTP/1.1 100 Continue\r\n\r\nHTTP/1.1 401 Unauthorized\r\n\r\nbody"
	if got := lastHTTPStatus(raw); got != "HTTP/1.1 401 Unauthorized" {
		t.Errorf("lastHTTPStatus = %q", got)
	}
	if got := lastHTTPStatus("no status here"); got != "<no HTTP response from mate>" {
		t.Errorf("lastHTTPStatus(no status) = %q", got)
	}
}

// --- preflight -----------------------------------------------------------------

// TestMateSEMPPreflightSuccess covers host networking with a server certificate
// and no tls.cas: the broker serves a TLS SEMP listener, so the mate is reached
// over https on the TLS port with certificate verification disabled and a
// warning naming tls.cas as the fix. The certificate is set explicitly because
// it is the precondition -- without it the answer is plaintext, which
// TestMateSEMPPreflightPlaintextWarns covers.
func TestMateSEMPPreflightSuccess(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		return []byte(sempOK), nil
	}}
	o, _ := newTestOps(t, localCfg("yes"), ft)
	o.Cfg.TLS = serverCert
	var logs []string
	o.Log = func(format string, args ...any) { logs = append(logs, fmt.Sprintf(format, args...)) }
	if err := o.MateSEMPPreflight(context.Background()); err != nil {
		t.Fatalf("MateSEMPPreflight error: %v", err)
	}
	calls := curlCalls(ft)
	if len(calls) != 1 || calls[0].argv[len(calls[0].argv)-1] != "https://10.0.0.12:1943/SEMP/v2/monitor" {
		t.Errorf("preflight calls = %v, want one GET of the mate's monitor endpoint over TLS", calls)
	}
	if !strings.Contains(calls[0].stdin, "insecure") {
		t.Errorf("preflight stdin = %q, want the insecure option when tls.cas is unset", calls[0].stdin)
	}
	if !strings.Contains(strings.Join(logs, "\n"), "tls.cas") {
		t.Errorf("preflight logs = %q, want a warning naming tls.cas", logs)
	}
	assertNoPasswordInArgv(t, ft, "adminpw")
}

// TestMateSEMPPreflightPlaintextWarns is the other half, and the one that keeps
// existing HA groups working: with no server certificate the broker has no TLS
// listener, so the mate is reached over http on 8080 exactly as before this
// change -- but the operator is told, before the request goes out, that the
// admin credentials are crossing the network unencrypted.
func TestMateSEMPPreflightPlaintextWarns(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		return []byte(sempOK), nil
	}}
	o, _ := newTestOps(t, localCfg("yes"), ft) // no TLS configured
	var logs []string
	o.Log = func(format string, args ...any) { logs = append(logs, fmt.Sprintf(format, args...)) }
	if err := o.MateSEMPPreflight(context.Background()); err != nil {
		t.Fatalf("MateSEMPPreflight error: %v", err)
	}
	calls := curlCalls(ft)
	if len(calls) != 1 || calls[0].argv[len(calls[0].argv)-1] != "http://10.0.0.12:8080/SEMP/v2/monitor" {
		t.Errorf("preflight calls = %v, want one plaintext GET on 8080 when no certificate is configured", calls)
	}
	if strings.Contains(calls[0].stdin, "insecure") {
		t.Errorf("preflight stdin = %q, want no TLS options on a plaintext request", calls[0].stdin)
	}
	joined := strings.Join(logs, "\n")
	if !strings.Contains(joined, "unencrypted") {
		t.Errorf("preflight logs = %q, want a warning that the credentials are sent unencrypted", logs)
	}
	assertNoPasswordInArgv(t, ft, "adminpw")
}

func TestMateSEMPPreflightMissingIP(t *testing.T) {
	ft := &fakeTransport{}
	o, _ := newTestOps(t, localCfg("yes"), ft)
	o.Cfg.Nodes.Backup.IP = ""
	err := o.MateSEMPPreflight(context.Background())
	if err == nil || !strings.Contains(err.Error(), "nodes.backup.ip") {
		t.Errorf("MateSEMPPreflight missing-IP err = %v, want it to name nodes.backup.ip", err)
	}
	if len(ft.outputs) != 0 {
		t.Error("MateSEMPPreflight must refuse before any transport call without a mate address")
	}
}

func TestMateSEMPPreflightNon2xx(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		return []byte("HTTP/1.1 401 Unauthorized\r\n\r\n"), nil
	}}
	o, _ := newTestOps(t, localCfg("yes"), ft)
	err := o.MateSEMPPreflight(context.Background())
	if err == nil || !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "nodes.backup.ip") {
		t.Errorf("MateSEMPPreflight non-2xx err = %v, want the status and the actionable causes", err)
	}
}

func TestMateSEMPPreflightTransportError(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		return nil, errors.New("connection refused")
	}}
	o, _ := newTestOps(t, localCfg("yes"), ft)
	err := o.MateSEMPPreflight(context.Background())
	if err == nil || !strings.Contains(err.Error(), "connection refused") ||
		!strings.Contains(err.Error(), "host-to-host") {
		t.Errorf("MateSEMPPreflight transport err = %v, want the cause and the host-to-host hint", err)
	}
}

// --- revert-activity -------------------------------------------------------------

// TestMateRevertActivitySuccess covers host networking with a server certificate
// and no tls.cas: TLS preferred, verification disabled, and a warning naming
// tls.cas as the fix -- matching TestMateSEMPPreflightSuccess.
func TestMateRevertActivitySuccess(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		return []byte(sempOK), nil
	}}
	o, _ := newTestOps(t, localCfg("yes"), ft)
	o.Cfg.TLS = serverCert
	var logs []string
	o.Log = func(format string, args ...any) { logs = append(logs, fmt.Sprintf(format, args...)) }
	if err := o.MateRevertActivity(context.Background()); err != nil {
		t.Fatalf("MateRevertActivity error: %v", err)
	}
	calls := curlCalls(ft)
	if len(calls) != 1 {
		t.Fatalf("MateRevertActivity made %d calls, want 1 (callers own the preflight)", len(calls))
	}
	post := calls[0]
	wantArgv := []string{"curl", "-is", "-K", "-", "https://10.0.0.12:1943/SEMP"}
	if len(post.argv) != len(wantArgv) {
		t.Fatalf("MateRevertActivity argv = %v, want %v", post.argv, wantArgv)
	}
	for i := range wantArgv {
		if post.argv[i] != wantArgv[i] {
			t.Fatalf("MateRevertActivity argv = %v, want %v", post.argv, wantArgv)
		}
	}
	if !strings.Contains(post.stdin, `user = "admin:adminpw"`) ||
		!strings.Contains(post.stdin, `data = "<rpc><admin><redundancy><revert-activity/></redundancy></admin></rpc>"`) ||
		!strings.Contains(post.stdin, "insecure") {
		t.Errorf("MateRevertActivity stdin = %q, want the user, data and insecure config lines", post.stdin)
	}
	if !strings.Contains(strings.Join(logs, "\n"), "tls.cas") {
		t.Errorf("MateRevertActivity logs = %q, want a warning naming tls.cas", logs)
	}
	assertNoPasswordInArgv(t, ft, "adminpw")
}

// assertNoPasswordInArgv is the S3 boundary check every mate-SEMP path shares:
// the admin credentials (and, where applicable, the RPC body) ride stdin
// only -- an argv token carrying either would surface in process listings and
// the Echo runner records.
func assertNoPasswordInArgv(t *testing.T, ft *fakeTransport, pass string) {
	t.Helper()
	for _, out := range ft.outputs {
		for _, tok := range out.argv {
			if strings.Contains(tok, pass) || strings.Contains(tok, "<rpc>") {
				t.Errorf("argv token %q leaks the credentials or RPC body", tok)
			}
		}
	}
}

// TestMateRevertActivityCredsAndBodyNeverInArgv is the S3 boundary check: the
// admin credentials and the RPC body ride stdin only -- an argv token carrying
// either would surface in process listings and the Echo runner records.
func TestMateRevertActivityCredsAndBodyNeverInArgv(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		return []byte(sempOK), nil
	}}
	o, _ := newTestOps(t, localCfg("yes"), ft)
	if err := o.MateRevertActivity(context.Background()); err != nil {
		t.Fatalf("MateRevertActivity error: %v", err)
	}
	assertNoPasswordInArgv(t, ft, "adminpw")
}

func TestMateRevertActivityNon2xx(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		return []byte("HTTP/1.1 500 Internal Server Error\r\n\r\n<rpc-reply><execute-result code=\"ok\"/></rpc-reply>"), nil
	}}
	o, _ := newTestOps(t, localCfg("yes"), ft)
	err := o.MateRevertActivity(context.Background())
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Errorf("MateRevertActivity non-2xx err = %v, want the HTTP status (a non-2xx ok body must not pass)", err)
	}
}

func TestMateRevertActivityRPCNotOK(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		return []byte("HTTP/1.1 200 OK\r\n\r\n<rpc-reply><execute-result code=\"fail\" reason=\"denied\"/></rpc-reply>"), nil
	}}
	o, _ := newTestOps(t, localCfg("yes"), ft)
	err := o.MateRevertActivity(context.Background())
	if err == nil || !strings.Contains(err.Error(), `code="fail"`) {
		t.Errorf("MateRevertActivity RPC-not-ok err = %v, want the mate's header-stripped reply", err)
	}
	if err != nil && strings.Contains(err.Error(), "HTTP/1.1 200") {
		t.Errorf("MateRevertActivity err = %v, want the headers stripped from the dumped reply", err)
	}
}

// TestMateRevertActivityBridgePlaintextOnly pins the last-resort branch: a
// bridge mapping only 8080 (no TLS port mapped at all) keeps today's plain
// http:// on the mapped host port, but now warns before the request that the
// admin credentials are crossing the wire unencrypted.
//
// The server certificate IS configured here deliberately: it isolates the port
// mapping as the deciding factor, so this proves the fallback comes from 1943
// being unmapped rather than from the certificate gate that
// TestSempPortWithoutCertificateStaysPlaintext covers separately.
func TestMateRevertActivityBridgePlaintextOnly(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		return []byte(sempOK), nil
	}}
	o, _ := newTestOps(t, localCfg("yes"), ft)
	o.Platform = config.Docker
	o.Cfg.TLS = serverCert
	o.Cfg.Docker.Network = config.Network{Mode: "bridge", Ports: []string{"18080:8080"}}
	var logs []string
	o.Log = func(format string, args ...any) { logs = append(logs, fmt.Sprintf(format, args...)) }
	if err := o.MateRevertActivity(context.Background()); err != nil {
		t.Fatalf("MateRevertActivity bridge error: %v", err)
	}
	post := curlCalls(ft)[0]
	if got := post.argv[len(post.argv)-1]; got != "http://10.0.0.12:18080/SEMP" {
		t.Errorf("bridge revert-activity URL = %q, want the mapped host port", got)
	}
	if strings.Contains(post.stdin, "cacert") || strings.Contains(post.stdin, "insecure") {
		t.Errorf("plaintext stdin = %q, want no TLS options", post.stdin)
	}
	joined := strings.Join(logs, "\n")
	if !strings.Contains(joined, "PLAINTEXT") || !strings.Contains(joined, "1943") {
		t.Errorf("plaintext logs = %q, want a warning naming the unencrypted channel and the TLS port fix", logs)
	}
	assertNoPasswordInArgv(t, ft, "adminpw")
}

// TestMateRevertActivityBridgeTLSNoCA pins the "TLS mapped, no tls.cas"
// branch: a bridge mapping exposing 1943 is preferred over a plaintext 8080
// mapping, but with no CAs configured curl's certificate verification is
// disabled (a self-signed broker certificate is the normal case here) and a
// warning names tls.cas as the way to turn verification on.
func TestMateRevertActivityBridgeTLSNoCA(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		return []byte(sempOK), nil
	}}
	o, _ := newTestOps(t, localCfg("yes"), ft)
	o.Platform = config.Docker
	o.Cfg.TLS = serverCert
	o.Cfg.Docker.Network = config.Network{Mode: "bridge", Ports: []string{"18080:8080", "18943:1943"}}
	var logs []string
	o.Log = func(format string, args ...any) { logs = append(logs, fmt.Sprintf(format, args...)) }
	if err := o.MateRevertActivity(context.Background()); err != nil {
		t.Fatalf("MateRevertActivity bridge TLS error: %v", err)
	}
	post := curlCalls(ft)[0]
	if got := post.argv[len(post.argv)-1]; got != "https://10.0.0.12:18943/SEMP" {
		t.Errorf("bridge TLS revert-activity URL = %q, want the mapped TLS host port", got)
	}
	if !strings.Contains(post.stdin, "insecure") {
		t.Errorf("bridge TLS stdin = %q, want the insecure option with no tls.cas configured", post.stdin)
	}
	if !strings.Contains(strings.Join(logs, "\n"), "tls.cas") {
		t.Errorf("bridge TLS logs = %q, want a warning naming tls.cas", logs)
	}
	assertNoPasswordInArgv(t, ft, "adminpw")
}

// TestMateRevertActivityBridgeTLSWithCA pins the fully-verified branch: a
// bridge mapping exposing 1943 plus tls.cas configured verifies the mate's
// certificate (--cacert) instead of disabling verification, and no warning is
// needed since the channel is both encrypted and verified.
func TestMateRevertActivityBridgeTLSWithCA(t *testing.T) {
	ft := &fakeTransport{responder: func(_ config.Role, argv []string, _ []byte) ([]byte, error) {
		return []byte(sempOK), nil
	}}
	o, _ := newTestOps(t, localCfg("yes"), ft)
	o.Platform = config.Docker
	o.Cfg.TLS = serverCert
	o.Cfg.Docker.Network = config.Network{Mode: "bridge", Ports: []string{"18943:1943"}}
	o.Cfg.TLS.CAs = []string{"/etc/solace/ca.pem"}
	var logs []string
	o.Log = func(format string, args ...any) { logs = append(logs, fmt.Sprintf(format, args...)) }
	if err := o.MateRevertActivity(context.Background()); err != nil {
		t.Fatalf("MateRevertActivity bridge TLS+CA error: %v", err)
	}
	post := curlCalls(ft)[0]
	if got := post.argv[len(post.argv)-1]; got != "https://10.0.0.12:18943/SEMP" {
		t.Errorf("bridge TLS+CA revert-activity URL = %q, want the mapped TLS host port", got)
	}
	if !strings.Contains(post.stdin, `cacert = "/etc/solace/ca.pem"`) {
		t.Errorf("bridge TLS+CA stdin = %q, want the cacert option naming tls.cas", post.stdin)
	}
	if strings.Contains(post.stdin, "insecure") {
		t.Errorf("bridge TLS+CA stdin = %q, want verification enabled, not insecure", post.stdin)
	}
	if len(logs) != 0 {
		t.Errorf("bridge TLS+CA logs = %v, want no warning: the channel is encrypted and verified", logs)
	}
	assertNoPasswordInArgv(t, ft, "adminpw")
}

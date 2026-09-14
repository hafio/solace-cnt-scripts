package broker

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"solace/internal/config"
)

// This file is the SEMP-over-HTTP control channel to the redundancy mate on the
// container platforms. The container transport is node-local (one broker per
// host), so a coordinated HA operation started on the primary host has exactly
// one way to touch the backup: an HTTP request this host's own broker container
// sends to the mate's SEMP service, addressed by redundancy.backup.addr. Every helper
// here execs curl with role config.Primary -- the only role the container
// transport reaches -- and reaches the mate by URL, never by role. Credentials
// and request bodies ride a curl config file on stdin (curl -K -), never argv
// (S3), mirroring Login.

// defaultSEMPPort is the broker's own plaintext SEMP port, the fallback used
// only when no TLS port can be resolved (TLS is preferred; plaintext with a
// warning is the last resort); host networking (the HA default) binds it
// directly on the mate host's interface.
const defaultSEMPPort = 8080

// defaultTLSSEMPPort is the broker's own TLS SEMP port. k8s already names this
// same 1943 in an unexported table (config.defaultK8sPorts's "tls-semp=1943"),
// but that function is internal to package config, so this is its own named
// constant rather than a second, unrelated import boundary for one integer.
const defaultTLSSEMPPort = 1943

// sempPort resolves the port the mate's SEMP service listens on as seen from
// this host, preferring TLS: host networking exposes both of the broker's own
// ports directly on the mate host's interface, so it always prefers the TLS
// one; a bridge network only exposes what network.ports maps, so the host
// side of the entry whose container side is 1943 (TLS SEMP) is tried first,
// falling back to the entry mapping 8080 (plaintext SEMP) when TLS is not
// mapped. Only single-port entries are recognized -- a range
// mapping ("18080-18090:8080-8090") may well cover one of them, but resolving
// it needs offset math this deliberately does not attempt, so the refusal
// names that limit. A bridge ports list mapping neither port is a loud
// refusal rather than a guess: guessing would either fail closed against a
// healthy broker or knock on an unrelated service.
func sempPort(cfg *config.Config, p config.Platform) (port int, tls bool, err error) {
	// A broker serves TLS SEMP only when it HAS a server certificate: without
	// tls.cert the 1943 listener does not come up, so preferring it would send
	// every mate call at a closed port and break an HA group that works today over
	// plaintext. This is decided from config rather than by probing because the
	// alternative -- try TLS, fall back on failure -- would have to retry the
	// MUTATION half too, and re-sending `redundancy revert-activity` after an
	// ambiguous failure is not a thing to do to a live redundancy pair.
	tlsConfigured := cfg.TLS.Cert != "" && cfg.TLS.CertKey != ""

	net := cfg.NetworkBlock(p)
	if net.Mode != "bridge" {
		if tlsConfigured {
			return defaultTLSSEMPPort, true, nil
		}
		return defaultSEMPPort, false, nil
	}
	if port, ok := bridgeHostPort(net.Ports, strconv.Itoa(defaultTLSSEMPPort)); ok && tlsConfigured {
		return port, true, nil
	}
	if port, ok := bridgeHostPort(net.Ports, strconv.Itoa(defaultSEMPPort)); ok {
		return port, false, nil
	}
	if _, ok := bridgeHostPort(net.Ports, strconv.Itoa(defaultTLSSEMPPort)); ok && !tlsConfigured {
		// The one combination the generic message below would misdescribe: the TLS
		// port IS mapped, so "map 1943" is unhelpful advice -- what is missing is
		// the certificate that would make the broker listen on it.
		return 0, false, fmt.Errorf("cannot reach the mate's SEMP port: %s.network.ports maps container port 1943 "+
			"(TLS SEMP) but tls.cert/tls.certKey are not configured, so the broker serves no TLS listener there; "+
			"either configure the server certificate or also map 8080 (plaintext SEMP)", string(p))
	}
	return 0, false, fmt.Errorf("cannot determine the mate's SEMP port: %s.network.mode is bridge but network.ports has "+
		"no single-port entry mapping container port 1943 (TLS SEMP) or 8080 (plaintext SEMP); add one (e.g. "+
		"\"1943:1943\") so the coordinated redundancy/leader steps can reach the mate over TLS -- a port RANGE "+
		"covering either is not recognized here", string(p))
}

// bridgeHostPort returns the host-side port of the bridge-mode network.ports
// entry whose container side is containerPort, or false if none maps it.
func bridgeHostPort(ports []string, containerPort string) (int, bool) {
	for _, entry := range ports {
		parts := strings.Split(entry, ":")
		if len(parts) < 2 {
			continue
		}
		if strings.TrimSuffix(parts[len(parts)-1], "/tcp") != containerPort {
			continue
		}
		if port, err := strconv.Atoi(parts[len(parts)-2]); err == nil {
			return port, true
		}
	}
	return 0, false
}

// curlConfigLine renders one `key = "value"` line of a curl config file (-K -),
// escaping backslashes and double quotes. Today's values carry neither, but the
// escaping lives here rather than being trusted at each call site so a future
// value with a quote cannot silently corrupt the config.
func curlConfigLine(key, value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return key + ` = "` + value + `"` + "\n"
}

// curlConfigFlag renders one boolean curl config line (e.g. "insecure"), the
// shape a no-argument option takes in a curl -K file -- unlike
// curlConfigLine's `key = "value"` form, which a boolean option does not take.
func curlConfigFlag(key string) string { return key + "\n" }

// Credential is the login one SEMP request is made with. It is a parameter rather
// than a field read off the config because sempCurl now serves two different brokers:
// the HA mate, which shares this deployment's admin password, and a REPLICATION mate,
// which is a separate broker whose password this env file states separately.
//
// Passing it explicitly is what stops the second case silently sending the first case's
// password: a DR site reached with the local broker's credentials would fail to
// authenticate at best, and at worst succeed because the two happen to match, hiding a
// misconfiguration until the day they diverge.
type Credential struct {
	User string
	Pass string
}

// LocalAdmin is this deployment's own admin login, for the HA paths that have always
// used it. AdminUser is a constant because it is the BROKER's name for the account, not
// a schema field (config.AdminUser).
func (o *Ops) LocalAdmin() Credential {
	return Credential{User: config.AdminUser, Pass: o.Cfg.SEMP.AdminPass}
}

// sempCurl execs curl in THIS host's broker container against url, with cred plus the
// caller's extra config lines (e.g. a POST body) on stdin via `curl -K -` -- nothing
// secret reaches argv or an echoed command (S3).
//
// The role is always config.Primary: the request is issued FROM this host, and which
// broker it reaches is decided by the URL alone. That is what lets one primitive serve
// both an HA mate on the same rack and a replication mate across a WAN.
//
// The `user` line is written HERE and nowhere else. A caller could otherwise pass its
// own through extraLines and rely on curl's behaviour for a repeated key, which is
// unspecified for this case and would make which password was sent depend on ordering.
func (o *Ops) sempCurl(ctx context.Context, url string, cred Credential, extraLines ...string) ([]byte, error) {
	cfg := curlConfigLine("user", cred.User+":"+cred.Pass) + strings.Join(extraLines, "")
	out, err := o.T.OutputInput(ctx, config.Primary, []byte(cfg), "curl", "-is", "-K", "-", url)
	if err != nil {
		return nil, fmt.Errorf("SEMP request to %s failed: %w", url, err)
	}
	return out, nil
}

// mateTarget is how a coordinated flow reaches the mate's SEMP service: the
// base URL, any extra curl -K lines the TLS leg needs (cacert/insecure), and a
// warning the caller must log BEFORE issuing the request whenever the
// admin credentials are about to cross the wire less than fully verified.
type mateTarget struct {
	url      string
	curlOpts []string
	warn     string
}

// mateSEMPTarget resolves redundancy.backup.addr and the mate's SEMP port into a
// mateTarget, preferring TLS: without it, the admin password crosses the
// network to the mate in cleartext. A mapped TLS port with tls.cas configured
// verifies the mate's certificate against them; a mapped TLS port with no CAs
// configured falls back to curl's insecure mode instead of refusing outright -- a
// self-signed broker certificate is the normal case here, and every existing
// HA container deployment would otherwise break. It falls back to today's
// plaintext port when the broker serves no TLS listener at all -- no server
// certificate configured, or on a bridge network no mapping for 1943; unlike Login,
// which really is hardcoded to http://localhost where nothing leaves the
// machine, this request crosses the network between hosts, so the plaintext
// case always carries a warning naming the fix.
func (o *Ops) mateSEMPTarget() (mateTarget, error) {
	ip := o.Cfg.Redundancy.Backup.Addr
	if ip == "" {
		return mateTarget{}, fmt.Errorf("redundancy.backup.addr is not set; the coordinated redundancy/leader steps need it to reach " +
			"the mate's SEMP service (see env/sample.yaml)")
	}
	port, tls, err := sempPort(o.Cfg, o.Platform)
	if err != nil {
		return mateTarget{}, err
	}
	if !tls {
		return mateTarget{
			url: fmt.Sprintf("http://%s:%d", ip, port),
			warn: fmt.Sprintf("sending the admin credentials to the mate's SEMP service at %s:%d over PLAINTEXT HTTP "+
				"-- they cross the network unencrypted. To encrypt this channel the broker needs a server "+
				"certificate (tls.cert/tls.certKey), and on a bridge network container port 1943 (TLS SEMP) must "+
				"be mapped in %s.network.ports (e.g. \"1943:1943\")",
				ip, port, string(o.Platform)),
		}, nil
	}
	// The TLS leg does not verify the mate's certificate, and tls.cas cannot make
	// it: those are paths on THIS host, while the curl that would read them is
	// exec'd inside the broker container (sempCurl), where the tool mounts no CA
	// material at all -- domain CAs reach the broker through
	// `config apply domain-certs`, which installs them into the broker's own trust
	// store rather than onto a filesystem curl can point at.
	//
	// This used to pass tls.cas[0] to `cacert`, which made one field mean a host
	// path to every other reader and an in-broker path to this one. It could not
	// have worked: a relative value resolved against curl's working directory
	// inside the container, and once host paths resolve against the env file's
	// directory it is an absolute host path that certainly does not exist there.
	// So the honest shape is one branch, always warning, rather than a verified
	// branch that never verified.
	return mateTarget{
		url:      fmt.Sprintf("https://%s:%d", ip, port),
		curlOpts: []string{curlConfigFlag("insecure")},
		warn: fmt.Sprintf("not verifying the mate's TLS certificate at %s:%d -- the admin credentials are "+
			"encrypted in transit but the mate's identity is not checked. tls.cas cannot verify this leg: "+
			"curl runs inside the broker container, where those host files are not mounted", ip, port),
	}, nil
}

// MateSEMPPreflight is the read-only reachability check every coordinated flow
// runs before its first mutation: the same GET Login sends to localhost, aimed
// at the mate. It is exported so a caller can run it standalone --
// the claim that this path and the port resolution behind it are right has never
// been checked against a live broker, and reaching it only through
// RedundancyCoordinated would mean failing over a real HA group to find out.
//
// A healthy HA group does NOT imply this passes -- redundancy only
// needs its own ports (8300-8302, 8741, 55555) open between the hosts, so SEMP
// can be firewalled while failover works; finding that out here, while the
// group is still undisturbed, is the point. One attempt, no retry: repeated
// failed logins can trip the broker's brute-force lockout on the mate.
func (o *Ops) MateSEMPPreflight(ctx context.Context) error {
	target, err := o.mateSEMPTarget()
	if err != nil {
		return err
	}
	if target.warn != "" {
		o.progress().Warn("%s", target.warn)
	}
	url := target.url + "/SEMP/v2/monitor"
	out, err := o.sempCurl(ctx, url, o.LocalAdmin(), target.curlOpts...)
	if err != nil {
		return fmt.Errorf("mate SEMP preflight: %w -- check that the SEMP port is reachable host-to-host (a working "+
			"HA group does not need it open) and that redundancy.backup.addr is the mate's address", err)
	}
	if !anyHTTP2xx(string(out)) {
		return fmt.Errorf("mate SEMP preflight: login to %s failed (%s) -- check the admin credentials, "+
			"redundancy.backup.addr and any firewall between the hosts", url, lastHTTPStatus(string(out)))
	}
	return nil
}

// MateRevertActivity sends the one cross-host mutation the coordinated flows
// need (everything else runs on the primary, matching the k8s sequences):
// admin-level `redundancy revert-activity` on the mate, over SEMP v1.
//
// Exported so it can be driven against a mate that is ALREADY
// standby -- the case revertActivityMateBody's own comment flags as unverified,
// and the only way to learn whether this RPC is idempotent without moving
// activity on a live group. A caller doing that must confirm the mate is standby
// first (ShowRedundancy + MateActivityState); sending it to an active mate is a
// real failover.
//
// Callers run MateSEMPPreflight before their own first mutation -- not here, so
// the check happens while the group is still undisturbed. Single shot, never
// inside poll(): retrying a failed login risks the mate's brute-force lockout,
// and the caller's own show-redundancy poll is what confirms the revert took.
func (o *Ops) MateRevertActivity(ctx context.Context) error {
	target, err := o.mateSEMPTarget()
	if err != nil {
		return err
	}
	if target.warn != "" {
		o.progress().Warn("%s", target.warn)
	}
	url := target.url + "/SEMP"
	out, err := o.sempCurl(ctx, url, o.LocalAdmin(),
		append(target.curlOpts, curlConfigLine("data", revertActivityMateBody()))...)
	if err != nil {
		return err
	}
	if !anyHTTP2xx(string(out)) {
		return fmt.Errorf("mate revert-activity request to %s failed (%s) -- check the admin credentials, "+
			"redundancy.backup.addr and any firewall between the hosts", url, lastHTTPStatus(string(out)))
	}
	if !sempV1OK(out) {
		return fmt.Errorf("mate rejected the revert-activity RPC; its reply:\n%s", httpBody(out))
	}
	return nil
}

// revertActivityMateBody is the SEMP v1 RPC equivalent of the CLI's admin-level
// `redundancy revert-activity` (revertActivityConfigureScript), POSTed to the
// mate's /SEMP endpoint.
//
// Three of the four unknowns here are now settled against the broker's own schemas
// (semp/semp-rpc-soltr.xsd and semp-rpc-reply-soltr.xsd, captured from 10.26.0.8827):
//
//   - the request path is exactly rpc > admin > redundancy > revert-activity;
//   - omitting the semp-version attribute is schema-legal, not a gamble -- the request
//     schema declares it `use="optional"` with the broker's own version as the default,
//     which is also what keeps the body quote-free inside a curl config `data` line;
//   - the reply carries <execute-result code="ok|fail"> with optional reason/reasonCode,
//     so sempV1OK's match on code="ok" is the right success test.
//
// STILL NEEDS VERIFICATION ON A LIVE BROKER: the /SEMP HTTP path itself (the XSDs
// describe the payload, not the endpoint), and whether reverting a mate that is
// ALREADY standby returns code="ok" (idempotent) or a benign failure needing a
// carve-out here.
func revertActivityMateBody() string {
	return "<rpc><admin><redundancy><revert-activity/></redundancy></admin></rpc>"
}

// sempV1OKRe matches the <execute-result code="ok"/> element a successful SEMP
// v1 reply carries, tolerating other attributes before code.
var sempV1OKRe = regexp.MustCompile(`<execute-result[^>]*\bcode="ok"`)

// sempV1OK reports whether a raw SEMP v1 reply (headers included) reports success.
func sempV1OK(raw []byte) bool { return sempV1OKRe.Match(raw) }

// anyHTTP2xx reports whether any HTTP status line in a raw `curl -is` capture
// is 2xx, tolerating a preceding 100 Continue or redirect line like Login does.
func anyHTTP2xx(out string) bool {
	for _, l := range httpStatusLines(out) {
		if isHTTP2xx(l) {
			return true
		}
	}
	return false
}

// lastHTTPStatus is the final HTTP status line of a raw capture, for error
// messages; a reply with none reads as "<no HTTP response from mate>".
func lastHTTPStatus(out string) string {
	lines := httpStatusLines(out)
	if len(lines) == 0 {
		return "<no HTTP response from mate>"
	}
	return lines[len(lines)-1]
}

// httpBody strips the status line and headers `curl -is` prepends, so an error
// message shows only the broker's own reply.
func httpBody(raw []byte) string {
	s := string(raw)
	if i := strings.Index(s, "\r\n\r\n"); i >= 0 {
		return s[i+4:]
	}
	if i := strings.Index(s, "\n\n"); i >= 0 {
		return s[i+2:]
	}
	return s
}

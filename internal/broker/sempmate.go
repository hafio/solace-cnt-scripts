package broker

import (
	"context"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"solace/internal/config"
)

// sempmate.go reaches a replication mate over SEMP v1, for a site this tool cannot exec
// into.
//
// Every request and reply path here is READ OFF THE BROKER'S OWN SCHEMAS
// (semp/semp-rpc-soltr.xsd and semp-rpc-reply-soltr.xsd, captured from 10.26.0.8827)
// rather than inferred, and each is recorded beside the thing it shapes. A path that
// skips a level is XML the broker rejects; one that is subtly wrong is worse.
//
// The transport is `sempCurl`: curl exec'd inside THIS host's own broker container, with
// the credential and body on stdin via `curl -K -`. So the mate needs no kubeconfig and
// no second exec path -- only a reachable SEMP endpoint.
//
// ONE ADVANTAGE OVER THE CLI LEG: this channel needs no broker type. `show replication`
// over the CLI prints two different layouts, but its SEMP reply models both as sibling
// structures under one `mate` element -- an appliance fills `connect-via`, software fills
// `mate-list` -- so the reply says which it is and nothing has to be discovered.

// sempMate is a MateChannel over SEMP v1.
type sempMate struct {
	ops  *Ops
	url  string     // base, e.g. https://10.0.1.1:1943
	cred Credential // the MATE's admin login, not this deployment's
	opts []string   // extra curl -K lines (insecure, etc.)
	what string
}

// NewSEMPMate builds the channel to one replication site over SEMP.
//
// pass is the mate's own admin password, already resolved from the site's
// pass/passEnv/passSecret -- a DR site is a different broker, and sending this
// deployment's password to it would fail at best and, if the two happened to match, hide
// the misconfiguration until the day they diverged.
func (o *Ops) NewSEMPMate(site config.ReplSite, pass string) (MateChannel, error) {
	s := site.Via.SEMP
	if s == nil {
		return nil, fmt.Errorf("replication site %q is not reached over SEMP", site.VirtualRouterName)
	}
	scheme := "http"
	if s.TLS {
		scheme = "https"
	}
	var opts []string
	if s.Insecure {
		opts = append(opts, curlConfigFlag("insecure"))
	}
	return &sempMate{
		ops:  o,
		url:  fmt.Sprintf("%s://%s:%d", scheme, s.Host, s.Port),
		cred: Credential{User: config.AdminUser, Pass: pass},
		opts: opts,
		what: fmt.Sprintf("site %s over SEMP at %s:%d", site.VirtualRouterName, s.Host, s.Port),
	}, nil
}

func (m *sempMate) Describe() string { return m.what }

// rpc posts one SEMP v1 request and returns the reply body.
//
// `semp-version` is deliberately omitted from the request: the schema declares it
// use="optional" with a default, so leaving it out lets one binary talk to brokers of
// different versions rather than asserting a version it cannot know.
func (m *sempMate) rpc(ctx context.Context, body string) ([]byte, error) {
	out, err := m.ops.sempCurl(ctx, m.url+"/SEMP", m.cred,
		append(append([]string{}, m.opts...), curlConfigLine("data", body))...)
	if err != nil {
		return nil, err
	}
	if !anyHTTP2xx(string(out)) {
		return nil, fmt.Errorf("SEMP request to %s failed (%s) -- check the credential, the host and "+
			"port, and any firewall between the two sites", m.what, lastHTTPStatus(string(out)))
	}
	raw := []byte(httpBody(out))
	// execute-result is a SIBLING of <rpc> under <rpc-reply>, and its status is the
	// `code` ATTRIBUTE. A reply can be HTTP 200 and still say fail, which is the whole
	// reason this is checked separately from the status line.
	var res sempExecuteResult
	if err := xml.Unmarshal(raw, &res); err != nil {
		return nil, fmt.Errorf("%s returned a reply this tool could not parse: %w", m.what, err)
	}
	if res.Result.Code != "ok" {
		return nil, fmt.Errorf("%s rejected the request: code=%q reason=%q", m.what,
			res.Result.Code, res.Result.Reason)
	}
	return raw, nil
}

// sempExecuteResult reads only the verdict, so every call can check it without knowing
// what the rest of the reply looks like.
type sempExecuteResult struct {
	XMLName xml.Name `xml:"rpc-reply"`
	Result  struct {
		Code   string `xml:"code,attr"`
		Reason string `xml:"reason,attr"`
	} `xml:"execute-result"`
}

// Preflight proves the whole path with a read that changes nothing, before any flow's
// first write.
func (m *sempMate) Preflight(ctx context.Context) error {
	if _, err := m.rpc(ctx, "<rpc><show><replication/></show></rpc>"); err != nil {
		return fmt.Errorf("cannot reach %s: %w", m.what, err)
	}
	return nil
}

// sempReplicationReply mirrors /rpc-reply/rpc/show/replication.
//
// The two mate shapes are siblings, exactly as the schema declares them
// (semp-rpc-reply-soltr.xsd:8290-8396): `connect-via` (maxOccurs=2, each carrying an
// address and a connect-ports trio) is the APPLIANCE form, and `mate-list/connect-via`
// (maxOccurs=6, each an address+transport pair) is the SOFTWARE form. Whichever is
// populated tells this code which broker it is talking to, so unlike the CLI leg nothing
// has to be discovered from a banner.
type sempReplicationReply struct {
	XMLName xml.Name `xml:"rpc-reply"`
	Mate    struct {
		RouterName string `xml:"router-name"`
		// APPLIANCE: one address, a port per transport.
		ConnectVia []struct {
			Address      string `xml:"address"`
			ConnectPorts struct {
				Uncompressed int `xml:"uncompressed"`
				Compressed   int `xml:"compressed"`
				SSL          int `xml:"ssl"`
			} `xml:"connect-ports"`
		} `xml:"connect-via"`
		// SOFTWARE: an address:port plus a transport label, repeated.
		MateList struct {
			ConnectVia []struct {
				Address   string `xml:"address"`
				Transport string `xml:"transport"`
			} `xml:"connect-via"`
		} `xml:"mate-list"`
	} `xml:"rpc>show>replication>mate"`
}

func (m *sempMate) ShowReplication(ctx context.Context) (MateConfig, error) {
	raw, err := m.rpc(ctx, "<rpc><show><replication/></show></rpc>")
	if err != nil {
		return MateConfig{}, err
	}
	var r sempReplicationReply
	if err := xml.Unmarshal(raw, &r); err != nil {
		return MateConfig{}, fmt.Errorf("%s: reading `show replication`: %w", m.what, err)
	}
	out := MateConfig{VirtualRouterName: strings.TrimSpace(r.Mate.RouterName)}

	// Software first: when mate-list is populated it is the authoritative shape, and an
	// appliance never fills it.
	for _, cv := range r.Mate.MateList.ConnectVia {
		t, tErr := schemaTransport(cv.Transport)
		if tErr != nil {
			return MateConfig{}, fmt.Errorf("%s: %w", m.what, tErr)
		}
		e, pErr := parseHostPort(cv.Address, t)
		if pErr != nil {
			return MateConfig{}, fmt.Errorf("%s: %w", m.what, pErr)
		}
		out.Endpoints = append(out.Endpoints, e)
	}
	if len(out.Endpoints) > 0 {
		return out, nil
	}
	for _, cv := range r.Mate.ConnectVia {
		host := strings.TrimSpace(cv.Address)
		if host == "" {
			continue
		}
		for _, p := range []struct {
			port      int
			transport string
		}{
			{cv.ConnectPorts.Uncompressed, config.TransportPlainText},
			{cv.ConnectPorts.Compressed, config.TransportCompressed},
			{cv.ConnectPorts.SSL, config.TransportSSL},
		} {
			// A port of zero is the transport not being configured, not port 0 --
			// the field is an unsignedInt with no "absent" representation.
			if p.port == 0 {
				continue
			}
			out.Endpoints = append(out.Endpoints, config.ReplEndpoint{
				Host: host, Port: p.port, Transport: p.transport,
			})
		}
	}
	return out, nil
}

// sempVPNReplicationReply mirrors
// /rpc-reply/rpc/show/message-vpn/replication/message-vpns/message-vpn.
//
// vpn-name, admin-state and config-state are all CHILD ELEMENTS, and config-state is the
// ROLE -- not a field called `role`, and not the `state` the write uses.
type sempVPNReplicationReply struct {
	XMLName xml.Name `xml:"rpc-reply"`
	VPNs    []struct {
		Name        string `xml:"vpn-name"`
		AdminState  string `xml:"admin-state"`
		ConfigState string `xml:"config-state"`
	} `xml:"rpc>show>message-vpn>replication>message-vpns>message-vpn"`
}

func (m *sempMate) ReadRoles(ctx context.Context, vpns []string) (map[string]VPNRepl, error) {
	// The wildcard is legal here because the SHOW operand is msgSvc_Name_param, which
	// permits `*`; the CONFIG operand used by SetRole is a different type that does not.
	raw, err := m.rpc(ctx,
		"<rpc><show><message-vpn><vpn-name>*</vpn-name><replication/></message-vpn></show></rpc>")
	if err != nil {
		return nil, err
	}
	var r sempVPNReplicationReply
	if err := xml.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("%s: reading replication roles: %w", m.what, err)
	}
	all := make(map[string]VPNRepl, len(r.VPNs))
	for _, v := range r.VPNs {
		admin, aErr := sempAdminState(v.AdminState)
		if aErr != nil {
			return nil, fmt.Errorf("%s: message-vpn %q: %w", m.what, v.Name, aErr)
		}
		role, rErr := sempReplRole(v.ConfigState)
		if rErr != nil {
			return nil, fmt.Errorf("%s: message-vpn %q: %w", m.what, v.Name, rErr)
		}
		all[strings.TrimSpace(v.Name)] = VPNRepl{Admin: admin, Role: role}
	}
	if vpns == nil {
		return all, nil
	}
	out := make(map[string]VPNRepl, len(vpns))
	for _, v := range vpns {
		if st, ok := all[v]; ok {
			out[v] = st
		}
	}
	return out, nil
}

// sempAdminState and sempReplRole decode the reply enumerations. Both are TRI-STATE:
// `n/a` is what a VPN that does not replicate reports, and folding it into standby would
// let a switch promote against a VPN it cannot actually see.
func sempAdminState(s string) (AdminState, error) {
	switch strings.TrimSpace(s) {
	case "enabled":
		return AdminEnabled, nil
	case "shutdown":
		return AdminShutdown, nil
	case "n/a", "n-a", "":
		return AdminNA, nil
	default:
		return "", fmt.Errorf("admin-state %q is not one of enabled, shutdown or n/a", s)
	}
}

func sempReplRole(s string) (ReplRole, error) {
	switch strings.TrimSpace(s) {
	case "active":
		return RoleActive, nil
	case "standby":
		return RoleStandby, nil
	case "n/a", "n-a", "":
		return RoleNA, nil
	default:
		return "", fmt.Errorf("config-state %q is not one of active, standby or n/a", s)
	}
}

// SetRole writes one VPN's replication role.
//
// There is NO <admin> wrapper: `rpc > admin` has no message-vpn child at all, so the
// path is rpc > message-vpn > vpn-name > replication > state > active|standby, and
// vpn-name must come FIRST because the schema puts it in an xs:sequence ahead of the
// sub-command choice.
func (m *sempMate) SetRole(ctx context.Context, vpn string, role ReplRole) error {
	if role != RoleActive && role != RoleStandby {
		return fmt.Errorf("cannot set message-vpn %q to replication role %q: only active and standby "+
			"are settable (n/a is a reading, not a state to write)", vpn, role)
	}
	body := "<rpc><message-vpn><vpn-name>" + xmlEscape(vpn) +
		"</vpn-name><replication><state><" + string(role) + "/></state></replication></message-vpn></rpc>"
	if _, err := m.rpc(ctx, body); err != nil {
		return fmt.Errorf("set message-vpn %q to %s on %s: %w", vpn, role, m.what, err)
	}
	return nil
}

// sempRedundancyReply mirrors /rpc-reply/rpc/show/redundancy.
//
// Activity is per VIRTUAL ROUTER, not per node: a chassis reports one activity for its
// primary virtual router and another for its backup. `Local Active` on the primary is
// what this tool means by "the primary HA node holds activity".
type sempRedundancyReply struct {
	XMLName        xml.Name `xml:"rpc-reply"`
	ConfigStatus   string   `xml:"rpc>show>redundancy>config-status"`
	VirtualRouters struct {
		Primary struct {
			Status struct {
				Activity string `xml:"activity"`
			} `xml:"status"`
		} `xml:"primary"`
	} `xml:"rpc>show>redundancy>virtual-routers"`
}

func (m *sempMate) PrimaryActive(ctx context.Context) (bool, error) {
	raw, err := m.rpc(ctx, "<rpc><show><redundancy/></show></rpc>")
	if err != nil {
		return false, err
	}
	var r sempRedundancyReply
	if err := xml.Unmarshal(raw, &r); err != nil {
		return false, fmt.Errorf("%s: reading redundancy: %w", m.what, err)
	}
	// config-status is an xs:string with no enumeration in the schema, so the value set
	// is not schema-provable -- but "Enabled" is what the CLI prints and what the rest
	// of this codebase already matches on (rdEnabledUp). Anything else is a broker with
	// no HA group, which is its own only node and so passes.
	if !strings.EqualFold(strings.TrimSpace(r.ConfigStatus), "Enabled") {
		return true, nil
	}
	return strings.Contains(r.VirtualRouters.Primary.Status.Activity, activityLocalActive), nil
}

// Status returns the two reports as raw XML, for DISPLAY only.
//
// It is XML rather than the CLI's tables because that is what this channel has, and
// nothing decides anything from it -- every verdict comes from the parsed methods.
func (m *sempMate) Status(ctx context.Context) (string, error) {
	mate, err := m.rpc(ctx, "<rpc><show><replication/></show></rpc>")
	if err != nil {
		return "", err
	}
	vpns, err := m.rpc(ctx,
		"<rpc><show><message-vpn><vpn-name>*</vpn-name><replication/></message-vpn></show></rpc>")
	if err != nil {
		return "", err
	}
	return string(mate) + "\n" + string(vpns), nil
}

// xmlEscape escapes a value going into a hand-built request body.
//
// The bodies here are assembled as strings rather than marshalled, because the schema's
// element ORDER is load-bearing and a struct would have to encode it anyway. That makes
// escaping this function's job: a real VPN name is operator-chosen and one carrying `&`
// or `<` would otherwise produce a malformed document.
func xmlEscape(s string) string {
	var b strings.Builder
	if err := xml.EscapeText(&b, []byte(s)); err != nil {
		// EscapeText only fails if the writer does, and strings.Builder never does.
		return strconv.Quote(s)
	}
	return b.String()
}

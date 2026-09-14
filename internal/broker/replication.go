package broker

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"solace/internal/config"
)

// replication.go renders and reads the broker-level replication mate configuration.
//
// THE TWO PLATFORMS SPELL IT DIFFERENTLY, at the same broker version, and both halves
// of this file are split down that line (TestApplianceReplicationGrammarDiffersFromSoftware
// is where the difference was first pinned):
//
//	software   replication mate connect-via "host:port" ["transport"]   one line per endpoint
//	           replication mate virtual-router-name "<v>"
//
//	appliance  replication mate connect-port <port> [compressed|ssl]    one line per endpoint
//	           replication mate virtual-router-name "<v>" connect-via "<host>"
//
// An appliance carries ONE address for the whole mate and a port per transport; software
// carries an address per entry with the port inside it, up to two per transport. So the
// model is a flat list of endpoints rather than hosts x ports: a real software capture
// holds two ssl entries on one port, which a cross product could only express by
// inventing four lines the broker never had. The SEMP v1 reply schema models the same
// split as two sibling structures -- an appliance `connect-via` with a `connect-ports`
// trio, and a software `mate-list` of (address, transport) pairs capped at six, which is
// three transports times two addresses.
//
// `show replication` is NOT normalised across the two, either -- it mirrors each
// grammar -- so the parsers are split the same way and chosen by the same BrokerType.

// MateConfig is one broker's view of its replication mate: what it is called, and every
// address it is dialled on. It is what a `show replication` reports, what gets rendered
// into the CLI, and what the two are compared as.
type MateConfig struct {
	// VirtualRouterName is the mate's own name, e.g. "v:tuas8csol1". It is carried
	// verbatim in both directions -- the env file states it rather than deriving it
	// from a router name, so nothing here has to know whether the "v:" convention holds.
	VirtualRouterName string
	Endpoints         []config.ReplEndpoint
}

// CLI transport keywords. These are the BROKER's words; the schema's own spellings are
// config.TransportPlainText and friends, and cliTransport is the only place the two meet.
const (
	cliPlainText  = "plain-text"
	cliCompressed = "compressed"
	cliSSL        = "ssl"
)

// cliTransport maps a schema transport to the broker's keyword. An empty schema value
// means plain text -- the CLI's own default for an unqualified entry.
func cliTransport(t string) (string, error) {
	switch t {
	case "", config.TransportPlainText:
		return cliPlainText, nil
	case config.TransportCompressed:
		return cliCompressed, nil
	case config.TransportSSL:
		return cliSSL, nil
	default:
		return "", fmt.Errorf("unknown replication transport %q", t)
	}
}

// schemaTransport is the reverse, for reading a capture or a `show` report. It accepts
// the broker's `plain-text` AND an absent token, because both spell the same thing: the
// keyword is optional and omitting it means plain text.
//
// `encrypted` is deliberately NOT accepted. The request schema offers it as a fourth
// keyword and it is the routing spelling of ssl, but neither report in hand ever emits
// it and the reply schema's six-entry cap only works if the vocabulary is three words.
// Mapping it silently would hide a broker that really did behave differently; refusing
// it by name is the fail-loud choice (S4a).
func schemaTransport(cli string) (string, error) {
	switch strings.Trim(strings.TrimSpace(cli), `"`) {
	case "", cliPlainText, "plaintext":
		return config.TransportPlainText, nil
	case cliCompressed:
		return config.TransportCompressed, nil
	case cliSSL:
		return config.TransportSSL, nil
	default:
		return "", fmt.Errorf("unknown replication transport keyword %q: expected %s, %s or %s "+
			"(an absent keyword means %s)", cli, cliPlainText, cliCompressed, cliSSL, cliPlainText)
	}
}

// RenderMate returns the CLI lines that point THIS broker at the given mate, in the
// grammar t uses. The caller supplies the LOCAL broker's type, since the grammar is a
// property of the broker being configured rather than of the mate being described.
func RenderMate(t BrokerType, m MateConfig) ([]string, error) {
	if m.VirtualRouterName == "" {
		return nil, fmt.Errorf("the replication mate has no virtual-router-name")
	}
	if len(m.Endpoints) == 0 {
		return nil, fmt.Errorf("replication mate %q has no endpoints", m.VirtualRouterName)
	}
	switch t {
	case BrokerSoftware:
		return renderMateSoftware(m)
	case BrokerAppliance:
		return renderMateAppliance(m)
	default:
		return nil, fmt.Errorf("cannot render replication for an unknown broker type: the CLI grammar "+
			"differs between software and appliance, so %q is not enough to choose one", string(t))
	}
}

// renderMateSoftware writes one connect-via per endpoint, the address carrying the port.
//
// The transport keyword is OMITTED for plain text rather than written out. Both spell
// the same thing to the broker, and omitting it matches the CLI's own default and the
// appliance capture's bare `connect-port 55555`; the comparison treats the two forms as
// equal (SameEndpoints) so a broker that echoes the keyword back is not read as drift.
func renderMateSoftware(m MateConfig) ([]string, error) {
	lines := make([]string, 0, len(m.Endpoints)+1)
	for _, e := range m.Endpoints {
		t, err := cliTransport(e.Transport)
		if err != nil {
			return nil, err
		}
		addr := fmt.Sprintf("%s:%d", e.Host, e.Port)
		if t == cliPlainText {
			lines = append(lines, fmt.Sprintf("replication mate connect-via %q", addr))
			continue
		}
		lines = append(lines, fmt.Sprintf("replication mate connect-via %q %q", addr, t))
	}
	lines = append(lines, fmt.Sprintf("replication mate virtual-router-name %q", m.VirtualRouterName))
	return lines, nil
}

// renderMateAppliance writes a connect-port per endpoint and hangs the single address
// off the router-name line.
//
// Two things it REFUSES rather than papers over, both because the appliance grammar
// cannot express them: more than one distinct host (there is exactly one connect-via,
// and it is an attribute of a single-valued setting), and two endpoints sharing a
// transport (connect-port is not repeatable per transport -- the reply schema models the
// three as an xs:all). Silently dropping either would point replication at an address
// the operator never chose.
func renderMateAppliance(m MateConfig) ([]string, error) {
	hosts := map[string]bool{}
	byTransport := map[string]string{}
	lines := make([]string, 0, len(m.Endpoints)+1)
	for _, e := range m.Endpoints {
		t, err := cliTransport(e.Transport)
		if err != nil {
			return nil, err
		}
		hosts[e.Host] = true
		if prev, dup := byTransport[t]; dup {
			return nil, fmt.Errorf("this is an appliance, which takes ONE %s port for the replication "+
				"mate, but %q declares two (%s and %s:%d). `replication mate connect-port` is not "+
				"repeatable per transport", t, m.VirtualRouterName, prev, e.Host, e.Port)
		}
		byTransport[t] = fmt.Sprintf("%s:%d", e.Host, e.Port)

		// A bare connect-port IS the plain-text case: the appliance grammar has no
		// `plain-text` keyword at all, so an unqualified line is not a shorthand.
		if t == cliPlainText {
			lines = append(lines, fmt.Sprintf("replication mate connect-port %d", e.Port))
			continue
		}
		lines = append(lines, fmt.Sprintf("replication mate connect-port %d %s", e.Port, t))
	}
	if len(hosts) != 1 {
		return nil, fmt.Errorf("this is an appliance, which reaches its replication mate at ONE "+
			"address, but %q declares %d distinct hosts (%s). A mate that is a software HA pair with "+
			"two hostnames cannot be expressed here", m.VirtualRouterName, len(hosts),
			strings.Join(sortedSet(hosts), ", "))
	}
	lines = append(lines, fmt.Sprintf("replication mate virtual-router-name %q connect-via %q",
		m.VirtualRouterName, sortedSet(hosts)[0]))
	return lines, nil
}

// RenderMateRemovals returns the lines that clear whatever the broker currently holds
// and the desired configuration does not.
//
// Without this the apply is additive, not convergent: pointing a broker at a new mate
// would leave the old addresses in place, so a second run would not be a no-op and the
// broker would end up dialling somewhere nobody named. The `no` forms are the ones the
// schemas give -- software names the addr-port being cleared, the appliance's
// connect-port no-form takes only the transport, and `no virtual-router-name` takes no
// operand at all.
//
// Removals come FIRST in an apply, so a changed virtual-router-name does not collide
// with the one already set.
func RenderMateRemovals(t BrokerType, have, want MateConfig) ([]string, error) {
	var lines []string
	for _, e := range have.Endpoints {
		if containsEndpoint(want.Endpoints, e) {
			continue
		}
		ct, err := cliTransport(e.Transport)
		if err != nil {
			return nil, err
		}
		switch t {
		case BrokerSoftware:
			lines = append(lines, fmt.Sprintf("no replication mate connect-via %q",
				fmt.Sprintf("%s:%d", e.Host, e.Port)))
		case BrokerAppliance:
			if ct == cliPlainText {
				lines = append(lines, "no replication mate connect-port")
				continue
			}
			lines = append(lines, "no replication mate connect-port "+ct)
		default:
			return nil, fmt.Errorf("cannot render replication removals for an unknown broker type")
		}
	}
	// The router name is cleared only when it is actually changing. On an appliance the
	// address rides this same line, so re-setting it is how a host change lands -- and
	// clearing it when nothing changed would drop replication for no reason.
	if have.VirtualRouterName != "" && have.VirtualRouterName != want.VirtualRouterName {
		lines = append(lines, "no replication mate virtual-router-name")
	}
	return lines, nil
}

// SameEndpoints reports whether two endpoint sets describe the same addresses, ignoring
// order and treating an absent transport as plain text.
//
// The plain-text normalisation is the load-bearing part. A capture may spell an
// unqualified entry either way -- both are valid to the broker -- and import's diff is
// directional, so a keyword this tool omits and a broker echoes back would otherwise read
// as a missing line on every plain-text endpoint.
func SameEndpoints(a, b []config.ReplEndpoint) bool {
	if len(a) != len(b) {
		return false
	}
	for _, e := range a {
		if !containsEndpoint(b, e) {
			return false
		}
	}
	return true
}

func containsEndpoint(set []config.ReplEndpoint, e config.ReplEndpoint) bool {
	for _, c := range set {
		if c.Host == e.Host && c.Port == e.Port && normTransport(c.Transport) == normTransport(e.Transport) {
			return true
		}
	}
	return false
}

// normTransport folds the two spellings of plain text into one.
func normTransport(t string) string {
	if t == "" {
		return config.TransportPlainText
	}
	return t
}

func sortedSet(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ParseShowReplication reads a `show replication` report into a MateConfig.
//
// The two platforms print DIFFERENT layouts -- this is not one report with cosmetic
// variation -- so t chooses the reader:
//
//	appliance                             software
//	  Replication Mate:  v:tuas8csol1       Replication Mate:  v:xps-ps-02
//	    Connect-Via:     10.150.132.1           Plain Text:
//	    Connect-Ports:                          Compressed:
//	      Plain Text:    55555                  SSL:   xps-ps-02:55443
//	      Compressed:    55003                         xps-ps-02.other:55443
//	      SSL:           55443
//
// An appliance names one host and a port per transport; software groups host:port LISTS
// under each transport heading, with continuation entries on their own line. On both, a
// transport that is not configured still prints its heading with nothing after it, so a
// heading is not evidence of an endpoint.
func ParseShowReplication(t BrokerType, out []byte) (MateConfig, error) {
	lines := strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n")
	switch t {
	case BrokerSoftware:
		return parseShowReplicationSoftware(lines)
	case BrokerAppliance:
		return parseShowReplicationAppliance(lines)
	default:
		return MateConfig{}, fmt.Errorf("cannot read `show replication` for an unknown broker type: the "+
			"report layout differs between software and appliance, so %q is not enough to choose one",
			string(t))
	}
}

// showTransportHeadings maps the report's own labels to schema transports. The reports
// label the encrypted transport "SSL" on both platforms; "Plain Text" is two words.
var showTransportHeadings = map[string]string{
	"plain text": config.TransportPlainText,
	"compressed": config.TransportCompressed,
	"ssl":        config.TransportSSL,
}

// mateLabel is the line naming the mate on both platforms.
const mateLabel = "Replication Mate:"

// parseShowReplicationSoftware reads the transport-grouped layout. A heading owns every
// following line indented DEEPER than itself, until a blank line or a line at its own
// level, which is how a second address for one transport is carried:
//
//	    SSL:                           broker-b:55443
//	                                   broker-b.cluster_default:55443
//
// INDENTATION is what separates a heading from a continuation, not the presence of a
// colon -- a continuation is a `host:port` address and has one too. Splitting on the
// colon instead read `broker-b.cluster_default:55443` as a heading named
// "broker-b.cluster_default", found it in no transport table, and dropped the second
// address silently. The same rule governs block boundaries in blocks.go, and for the
// same reason.
func parseShowReplicationSoftware(lines []string) (MateConfig, error) {
	var m MateConfig
	current := ""
	headIndent := 0
	for _, raw := range lines {
		line := strings.TrimRight(raw, " \t")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			current = "" // a blank line ends a transport group
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		// A continuation of the heading still in effect: deeper than the heading, and
		// therefore an address rather than a label of its own.
		if current != "" && indent > headIndent {
			e, err := parseHostPort(trimmed, current)
			if err != nil {
				return MateConfig{}, err
			}
			m.Endpoints = append(m.Endpoints, e)
			continue
		}
		current = "" // anything at or above the heading's level ends its group

		if v, ok := labelValue(trimmed, mateLabel); ok {
			m.VirtualRouterName = v
			continue
		}
		// `ConfigSync:` and anything after it is a different section; stop before it
		// so a bridge setting is never read as an address.
		if strings.HasPrefix(trimmed, "ConfigSync:") {
			break
		}
		label, rest, ok := splitLabel(trimmed)
		if !ok {
			continue
		}
		t, known := showTransportHeadings[strings.ToLower(label)]
		if !known {
			continue
		}
		current, headIndent = t, indent
		if rest == "" {
			continue // configured nowhere: the heading prints with no value
		}
		e, err := parseHostPort(rest, t)
		if err != nil {
			return MateConfig{}, err
		}
		m.Endpoints = append(m.Endpoints, e)
	}
	return m, nil
}

// parseShowReplicationAppliance reads the one-host layout: a Connect-Via naming the
// address, then a port per transport under Connect-Ports.
func parseShowReplicationAppliance(lines []string) (MateConfig, error) {
	var m MateConfig
	host := ""
	type portEntry struct {
		transport string
		port      int
	}
	var ports []portEntry
	inPorts := false
	for _, raw := range lines {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "ConfigSync:") {
			break
		}
		if v, ok := labelValue(trimmed, mateLabel); ok {
			m.VirtualRouterName = v
			inPorts = false
			continue
		}
		if v, ok := labelValue(trimmed, "Connect-Via:"); ok {
			host = v
			inPorts = false
			continue
		}
		if strings.HasPrefix(trimmed, "Connect-Ports:") {
			inPorts = true
			continue
		}
		if !inPorts {
			continue
		}
		label, rest, ok := splitLabel(trimmed)
		if !ok {
			continue
		}
		t, known := showTransportHeadings[strings.ToLower(label)]
		if !known {
			inPorts = false // out of the Connect-Ports group and into the next section
			continue
		}
		if rest == "" {
			continue // that transport is not configured
		}
		p, err := strconv.Atoi(rest)
		if err != nil {
			return MateConfig{}, fmt.Errorf("`show replication` reported %q as the %s port, which is "+
				"not a number", rest, label)
		}
		ports = append(ports, portEntry{t, p})
	}
	if host == "" {
		// A mate NAMED but with no Connect-Via is not an unconfigured broker -- it is
		// this reader on the wrong report. Software prints no Connect-Via at all, so
		// feeding it here would otherwise yield a cheerful empty result and a caller
		// that concludes the mate is unconfigured. Fail loudly instead (S4a).
		if m.VirtualRouterName != "" {
			return MateConfig{}, fmt.Errorf("`show replication` names mate %q but reports no "+
				"Connect-Via. An appliance always prints one; a report without it is the software "+
				"layout, which groups addresses under each transport instead", m.VirtualRouterName)
		}
		// Genuinely unconfigured: no mate, no address.
		return m, nil
	}
	for _, p := range ports {
		m.Endpoints = append(m.Endpoints, config.ReplEndpoint{Host: host, Port: p.port, Transport: p.transport})
	}
	return m, nil
}

// --- per-VPN replication state -------------------------------------------------

// AdminState is whether replication is ENABLED for a VPN, and ReplRole is which side of
// the pair it plays. They are two facts, not one, and must be read together: a shut-down
// VPN still reports a role, sitting at its default, so treating the role alone as the
// answer misreports every VPN on a broker where replication was never turned on.
//
// Both are tri-state. `n/a` is what a VPN that does not replicate reports, and collapsing
// it into "standby" would let a switch promote against a VPN it cannot actually see. The
// SEMP v1 reply schema models the same two enums, `admin-state` and `config-state`.
//
// WHAT "-" ACTUALLY IS, verified against a real broker: `#config-sync`, the internal
// message-VPN every broker carries, reports "-" in BOTH columns
// (semp/show-message-vpn-replication.out:14, `#config-sync                     - - - -
// - - - -`). It is neither a user-configured VPN nor the `default` VPN, and it does not
// exist as a configurable object at all -- it appears in zero lines of either
// `show current-config` capture, so `configure` cannot even enter its node. Nothing this
// tool writes ever names it: it is not something an operator lists in
// `replication.vpns`, and both `Replicating()` and `mateConvergenceShutdowns`
// (replicationops.go) REQUIRE `Admin == AdminEnabled` (`Replicating()` also requires a
// resolved role), which is false for it either way. So a listed VPN, or an unlisted one
// this tool acts on, NEVER reaches a code path
// keyed on RoleNA/AdminNA in practice -- the tri-state exists to describe `#config-sync`
// truthfully when it is read, not because a real, operator-managed VPN is expected to
// report it. Do not delete these two constants or their flag-map entries: doing so makes
// `ParseVPNReplication`/`ReadRoles` fail loudly against every real broker's
// `show message-vpn * replication`, because `#config-sync` is always in that report.
type AdminState string

const (
	AdminEnabled  AdminState = "enabled"
	AdminShutdown AdminState = "shutdown"
	// AdminNA is what `#config-sync` reports for admin state; see this type's own doc
	// comment for why no user-configured or default VPN can reach it.
	AdminNA AdminState = "n/a"
)

// ReplRole is a VPN's replication role at one site.
type ReplRole string

const (
	RoleActive  ReplRole = "active"
	RoleStandby ReplRole = "standby"
	// RoleNA is what `#config-sync` reports for config state, for the identical reason
	// AdminNA exists -- see AdminState's doc comment. The same capture row reports "-"
	// for both facts at once.
	RoleNA ReplRole = "n/a"
)

// internalConfigSyncVPN is the broker's own config-sync VPN, which appears in
// `show message-vpn * replication` on every broker and is NOT a message-VPN anyone
// configures. It is named once here because two separate rules key on it: it is the
// reason the n/a states above exist at all, and `configure dr` refuses it if an env file
// lists it, since a presence check against the broker's report cannot tell it apart from
// a real VPN -- it is genuinely present.
const internalConfigSyncVPN = "#config-sync"

// QueueState is whether the VPN's DATA REPLICATION QUEUE exists, read from the report's
// own `Q` column (legend: "Q - Queue State (U=Up, D=Down, -=N/A)").
//
// It is a THIRD fact, derivable from neither of the other two, and reading it is what
// decides which `no shutdown` form is legal. The broker's default is
// `fail-on-existing-queue` ("The data replication queue must not already exist") and the
// alternative is `force-use-existing-queue` ("The data replication queue must already
// exist") -- semp/appliance_cli_reference.html -- so a wrong guess is a REFUSED re-enable
// in either direction, not a harmless extra keyword.
//
// That it cannot be inferred from admin state is settled by the real capture, which is
// why this column is read rather than guessed at: semp/show-message-vpn-replication.out
// carries `vpn-01  U S U - - - Y A` -- admin-UP, standby, Q n/a, so an ENABLED VPN need
// not have a queue -- beside `default  U A - D D N Y A`, admin-UP, active, Q down. On
// that broker the queue tracks the ROLE, not the enablement, and code that read
// enablement would send `force-use-existing-queue` to vpn-01, which has no queue to use.
//
// QueueUnknown is what a report with no Q column yields. It is deliberately NOT folded
// into QueueNA: absent evidence and evidence of absence select opposite keywords here.
type QueueState string

const (
	QueueUp   QueueState = "up"
	QueueDown QueueState = "down"
	QueueNA   QueueState = "n/a"
	// QueueUnknown is the empty string ON PURPOSE, so it is also VPNRepl's zero value: a
	// state built without reading the column means the column was not read, which is
	// exactly what unknown says. Any other spelling would make a partially-built VPNRepl
	// claim the queue is absent, and absent selects the opposite keyword from unknown.
	QueueUnknown QueueState = ""
)

// Exists reports whether the report says a data replication queue is there. Up and Down
// are both existence -- a queue can be down without being absent, and only `-` means
// there is none. Unknown is FALSE, so a caller that cannot read the column falls back to
// the broker's own default rather than asserting something the report never said.
func (q QueueState) Exists() bool { return q == QueueUp || q == QueueDown }

// VPNRepl is one VPN's replication state at one broker.
type VPNRepl struct {
	Admin AdminState
	Role  ReplRole
	Queue QueueState
}

// Replicating reports whether this VPN is actually replicating -- enabled, with a real
// role. It is the guard every caller wants before acting on Role.
func (v VPNRepl) Replicating() bool {
	return v.Admin == AdminEnabled && (v.Role == RoleActive || v.Role == RoleStandby)
}

// adminFlags and roleFlags decode the single letters the report prints. The legend at the
// top of the report defines them: "A - Admin State (U=Up, D=Down, -=N/A)" and
// "C - Config State (A=Active, S=Standby, -=N/A)".
var adminFlags = map[byte]AdminState{'U': AdminEnabled, 'D': AdminShutdown, '-': AdminNA}
var roleFlags = map[byte]ReplRole{'A': RoleActive, 'S': RoleStandby, '-': RoleNA}

// queueFlags decodes the `Q` column, whose legend is "Q - Queue State (U=Up, D=Down,
// -=N/A)" -- the same alphabet as the admin column, and a separate map anyway so the two
// cannot be swapped by a careless edit and so a broker that adds a queue letter is a
// one-line change here rather than a silent misread through the admin table.
var queueFlags = map[byte]QueueState{'U': QueueUp, 'D': QueueDown, '-': QueueNA}

// vpnSeparatorRE matches the rule under the column headers: runs of dashes separated by
// single spaces, and nothing else. It is what gives the column spans.
var vpnSeparatorRE = regexp.MustCompile(`^-+(?: +-+)+ *$`)

// ParseVPNReplication reads `show message-vpn * replication` into per-VPN state.
//
// The report is a FIXED-WIDTH flag table, not whitespace-delimited:
//
//	Message VPN                      A C B R Q S M T
//	-------------------------------- - - - - - - - -
//	A VPN WITH LONG NAME AND SPACES  D S - - - - N A
//	default                          U A - D D N Y A
//
// A real VPN name contains spaces -- three of the six rows in the capture this was
// written against do -- so splitting on whitespace would rename
// `A VPN WITH LONG NAME AND SPACES` to `A` and read its flags out of the wrong columns.
// The separator line is what defines the columns, and it is read rather than assumed.
//
// The column ORDER is read from the header too, not hardcoded by position. A broker that
// reordered or inserted a flag column would otherwise have its admin state read as its
// role, silently and with both values still looking valid -- the one misreading that
// could send a switchover the wrong way.
func ParseVPNReplication(out []byte) (map[string]VPNRepl, error) {
	lines := strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n")

	sep := -1
	for i, l := range lines {
		if vpnSeparatorRE.MatchString(strings.TrimRight(l, " \t")) {
			sep = i
			break
		}
	}
	if sep < 1 {
		return nil, fmt.Errorf("`show message-vpn * replication` has no column rule: expected a line of " +
			"dashes under the headers, which is what defines the fixed-width columns")
	}

	spans := dashSpans(strings.TrimRight(lines[sep], " \t"))
	if len(spans) < 3 {
		return nil, fmt.Errorf("`show message-vpn * replication` has %d columns, want at least 3 "+
			"(the VPN name, plus the A and C flags)", len(spans))
	}
	header := lines[sep-1]
	adminCol, roleCol, queueCol := -1, -1, -1
	for i, sp := range spans[1:] {
		switch strings.TrimSpace(sliceSpan(header, sp)) {
		case "A":
			adminCol = i + 1
		case "C":
			roleCol = i + 1
		case "Q":
			// OPTIONAL, unlike A and C. A report without it yields QueueUnknown and the
			// caller falls back to the broker's own default keyword, rather than the
			// whole parse failing over a column only one decision needs.
			queueCol = i + 1
		}
	}
	if adminCol < 0 || roleCol < 0 {
		return nil, fmt.Errorf("`show message-vpn * replication` headers %q carry no A (Admin State) "+
			"and C (Config State) columns; reading the flags by position instead would risk reporting "+
			"one as the other", strings.TrimSpace(header))
	}

	states := map[string]VPNRepl{}
	headerText := strings.TrimSpace(header)
	for _, raw := range lines[sep+1:] {
		line := strings.TrimRight(raw, " \t")
		if strings.TrimSpace(line) == "" {
			continue
		}
		// A repeated header and rule, which a paginated report prints again partway
		// down. Left in, the RULE parses as a row: its name column trims to a run of
		// dashes and its flag columns are literal `-`, which decode to n/a -- so a
		// phantom VPN named "--------------------------------" joins the map, reported
		// as existing but not replicating.
		if vpnSeparatorRE.MatchString(line) || strings.TrimSpace(line) == headerText {
			continue
		}
		// Too short to carry the flag columns: the echoed CLI prompt the capture ends
		// with, or a wrapped continuation. Not a VPN row.
		if len(line) < spans[roleCol].end || len(line) < spans[adminCol].end {
			continue
		}
		// The GUTTER -- the character immediately before a flag column -- is always a
		// space in a real row, because the name column is padded out to the rule. A
		// line that overruns it is not a row at all: an echoed prompt on a broker whose
		// router name is longer than the name column reaches the flag columns and would
		// otherwise fail the whole parse on a garbage flag, turning ordinary output
		// into a hard error.
		//
		// Known limit: this also skips a row whose name is long enough to fill the
		// column with no gutter left. No Solace name seen here comes close, and
		// skipping is the safer half of the trade -- the alternative reads flags out of
		// the name.
		if !gutterClear(line, spans[adminCol]) || !gutterClear(line, spans[roleCol]) {
			continue
		}
		name := strings.TrimSpace(sliceSpan(line, spans[0]))
		if name == "" {
			continue
		}
		admin, ok := flagByte(sliceSpan(line, spans[adminCol]), adminFlags)
		if !ok {
			return nil, fmt.Errorf("message-vpn %q reports admin state %q, which is not one of U, D or -",
				name, strings.TrimSpace(sliceSpan(line, spans[adminCol])))
		}
		role, ok := flagByte(sliceSpan(line, spans[roleCol]), roleFlags)
		if !ok {
			return nil, fmt.Errorf("message-vpn %q reports config state %q, which is not one of A, S or -",
				name, strings.TrimSpace(sliceSpan(line, spans[roleCol])))
		}
		// The queue column is read best-effort and never fails the parse: it is absent
		// on a report that does not carry it, and a row can be too short to reach it or
		// carry a letter this build does not know. Each of those is QueueUnknown, which
		// selects the broker's own default keyword downstream -- the same answer as not
		// looking, rather than a hard error over a column A and C do not depend on.
		queue := QueueUnknown
		if queueCol >= 0 && len(line) >= spans[queueCol].end && gutterClear(line, spans[queueCol]) {
			if q, ok := flagByte(sliceSpan(line, spans[queueCol]), queueFlags); ok {
				queue = q
			}
		}
		states[name] = VPNRepl{Admin: admin, Role: role, Queue: queue}
	}
	if len(states) == 0 {
		return nil, fmt.Errorf("`show message-vpn * replication` listed no message-VPNs; every broker " +
			"has at least the default VPN, so an empty table means the report was not understood")
	}
	return states, nil
}

// colSpan is one fixed-width column of a flag table, as [start, end). It is named
// apart from inject.go's own `span`, which tracks a port-line extent -- two different
// notions of a range in one package.
type colSpan struct{ start, end int }

// dashSpans reads the column rule into spans.
func dashSpans(rule string) []colSpan {
	var out []colSpan
	i := 0
	for i < len(rule) {
		if rule[i] != '-' {
			i++
			continue
		}
		j := i
		for j < len(rule) && rule[j] == '-' {
			j++
		}
		out = append(out, colSpan{i, j})
		i = j
	}
	return out
}

// gutterClear reports whether the character before a column is the space that separates
// it from the previous one. It is the cheapest way to tell a table row from a line that
// merely reaches that far.
func gutterClear(line string, sp colSpan) bool {
	if sp.start == 0 {
		return true
	}
	if sp.start > len(line) {
		return false
	}
	c := line[sp.start-1]
	return c == ' ' || c == '\t'
}

// sliceSpan takes a column out of a line, tolerating a line shorter than the rule --
// the last column of the last row is not padded on every broker.
func sliceSpan(line string, sp colSpan) string {
	if sp.start >= len(line) {
		return ""
	}
	end := sp.end
	if end > len(line) {
		end = len(line)
	}
	return line[sp.start:end]
}

// flagByte decodes a one-character flag column.
func flagByte[T ~string](col string, table map[byte]T) (T, bool) {
	var zero T
	t := strings.TrimSpace(col)
	if len(t) != 1 {
		return zero, false
	}
	v, ok := table[t[0]]
	return v, ok
}

// labelValue returns the value after an exact label, and whether the line carried it.
func labelValue(line, label string) (string, bool) {
	if !strings.HasPrefix(line, label) {
		return "", false
	}
	return strings.TrimSpace(strings.TrimPrefix(line, label)), true
}

// splitLabel splits "Label:  value" into its halves. The report pads values into a
// column, so the value may be empty even when the label is present.
func splitLabel(line string) (label, value string, ok bool) {
	i := strings.Index(line, ":")
	if i < 0 {
		return "", "", false
	}
	return strings.TrimSpace(line[:i]), strings.TrimSpace(line[i+1:]), true
}

// parseHostPort reads a "host:port" address. The host half may itself be a name with
// dots and underscores (a container network alias, in one real capture), so the split is
// on the LAST colon rather than the first.
func parseHostPort(s, transport string) (config.ReplEndpoint, error) {
	s = strings.Trim(strings.TrimSpace(s), `"`)
	i := strings.LastIndex(s, ":")
	if i <= 0 || i == len(s)-1 {
		return config.ReplEndpoint{}, fmt.Errorf("`show replication` reported %q as a replication "+
			"address, which is not host:port", s)
	}
	p, err := strconv.Atoi(s[i+1:])
	if err != nil {
		return config.ReplEndpoint{}, fmt.Errorf("`show replication` reported %q as a replication "+
			"address, whose port is not a number", s)
	}
	return config.ReplEndpoint{Host: s[:i], Port: p, Transport: transport}, nil
}

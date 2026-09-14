package broker

import (
	"fmt"
	"strings"
)

// This file makes a captured port change safe to import into a broker that will
// refuse it while the service is up. Like scripts.go and blocks.go it is pure --
// no transport, no I/O -- so InjectShutdown is golden-testable against a real
// capture with no broker involved.
//
// THE SHAPE, observed in semp/all.cli's broker-level "! Configure X Service:"
// sections (lines 63-124): a service that changes its listen port on a NEWER
// broker is wrapped in a shutdown sandwich --
//
//	service semp shutdown plain-text
//	service semp listen-port 8080
//	no service semp shutdown plain-text
//
// -- but an OLDER broker emits the bare middle line with nothing around it, and
// the broker rejects a port change on a service that is not shut down. So the
// leading shutdown has to be injected for those captures. InjectShutdown is the
// transform that does it, and refuses to guess at the one thing it must not
// guess at: whether the service ends up enabled.
//
// THE TRAP: the trailing state genuinely varies per service and per broker, so a
// transform that always closes with "no ... shutdown" would silently ENABLE a
// service the artifact means to leave disabled. Evidence from semp/all.cli that
// a shutdown can be the broker's real, final word: line 168 "redundancy
// shutdown" (the router's own redundancy config, not the "service redundancy"
// port block), line 279 "config-sync shutdown", lines 313/328 the default
// ldap-profile/radius-profile's own "shutdown", and lines 195/199 the spool's
// defragment-files schedule/threshold shutdown -- all four end shut down, with
// no matching un-shutdown anywhere in the file, while the spool itself (line
// 186/200) and every service in the port table below end enabled only because
// that happens to be true of THIS broker. Every port-changing service in
// semp/all.cli ends with "no ... shutdown" (lines 67,70,83,88,93,103,118,124),
// but that is a property of this one capture, not a rule this function may rely
// on. So the only place InjectShutdown decides an enablement state at all is
// case 3 below, and it decides it to be a no-op against the TARGET, not a guess
// at the SOURCE's intent.
//
// THE FOUR CASES, checked per (section, service, transport):
//
//  1. A real leading "service X shutdown [transport]" already precedes the port
//     line in this block's section -> nothing to do. No duplicate is emitted
//     even if a trailing line also already exists (the normal, already-safe
//     case -- e.g. every service in semp/all.cli:63-124).
//  2. No leading shutdown, but a real trailing "no service X shutdown
//     [transport]" for the same key exists later in the section -> inject ONLY
//     the leading shutdown. The trailing line is the artifact's own and passes
//     through untouched.
//  3. No leading AND no trailing anywhere in the section -> injecting a bare
//     shutdown would leave the service down with nothing to bring it back,
//     which is an outage. targetState is asked for the TARGET's current
//     enablement; the leading shutdown is injected and a restore line
//     reproducing exactly that reading is appended, making the round trip a
//     no-op on enablement. If targetState reports unknown (or is nil -- "no
//     target reachable"), NOTHING is injected: the port line passes through
//     with a warning, because guessing here risks an outage either way
//     (guess enabled and the target meant it off, or the reverse).
//  4. The service is not in serviceShutdownStyle -> passed through unchanged
//     with a warning. The broker may reject it outright (cli -Apes always
//     exits 0, so that cannot be detected here); it will surface in the
//     post-import verify diff (diff.go) instead.
//
// GROUPING: several port lines for one service commonly share a single
// shutdown/un-shutdown pair rather than one pair each -- smf's four listen-port
// lines (semp/all.cli:79-82) sit inside one "service smf shutdown" /
// "no service smf shutdown" (lines 78,83), and web-transport (99-103) and
// health-check (121-124) do the same. Deciding per PORT LINE rather than per
// KEY would inject a redundant leading/trailing pair around every one of those
// and, worse, would briefly re-enable the service between two port changes that
// the source never separated. So a key's leading/trailing search spans from its
// FIRST port line to its LAST, and an injection (if any) happens exactly once at
// each end of that span.
//
// TWO PARSING RULES this file follows rather than trusting the Block fields, and
// both were bugs first:
//
//   - It reads Block.Opener(), NOT Lines[0]. A pragma sitting between a section
//     comment and an opener is held and prepended to Lines, so Lines[0] is a
//     "! pragma:" comment on exactly the blocks that carry one -- and such a block
//     then parsed as unrecognised and was skipped with no injection and no
//     warning.
//   - It strips a leading "no" itself. blocks.go now records that prefix in
//     Block.Negate and leaves Kind as the bare command, so Kind=="service" would
//     in fact match a negated opener today -- but the negation still has to be
//     read to tell a leading "service X shutdown" from a trailing
//     "no service X shutdown", which is the whole distinction this function turns
//     on. Doing it from the opener keeps that reading in one place.

// shutdownStyle is where the transport keyword sits in a service's shutdown and
// listen-port commands, when it appears at all. The two shapes observed in
// semp/all.cli never mix within one service.
type shutdownStyle int

const (
	// shutdownBare is "service X shutdown" / "no service X shutdown" -- the whole
	// service goes down for any of its ports, and the port line carries no
	// transport keyword relevant to the shutdown decision (a trailing suffix such
	// as smf's "compressed"/"ssl" or health-check's "ssl" is passed through
	// verbatim but does not select a different shutdown pair).
	shutdownBare shutdownStyle = iota
	// shutdownTransportAfter is "service X shutdown <transport>" with the
	// transport, when present, as the LAST token of the matching port line --
	// semp/all.cli:65-70 ("service semp shutdown plain-text" / "... shutdown
	// ssl", and the port lines' own trailing "ssl" at line 69).
	shutdownTransportAfter
	// shutdownTransportBefore is "service X <transport> shutdown" with the
	// transport as the token immediately after the service name on the matching
	// port line -- semp/all.cli:109-110 ("no service rest incoming shutdown" /
	// "... outgoing shutdown"); rest has no listen-port line in this capture, so
	// the port-line half of this shape is inferred by symmetry, not observed.
	shutdownTransportBefore
)

// serviceShutdownStyle is how each broker "service" this function knows about
// spells its shutdown command, keyed by the token right after "service" on the
// opener. A service absent from this table is unknown to InjectShutdown (case 4
// above): its port lines are passed through with a warning rather than guessed
// at, because a wrong guess here either leaves a service the artifact wanted
// running down, or shuts down a transport the injection had no business
// touching.
//
// Evidence, all from semp/all.cli's broker-level "! Configure X Service:"
// sections:
//
//	semp          shutdownTransportAfter  lines 65-70  (plain-text, ssl)
//	rest          shutdownTransportBefore lines 108-110 (incoming, outgoing)
//	smf           shutdownBare            lines 78,83; ports 79-82
//	mate-link     shutdownBare            lines 86,88; port 87
//	redundancy    shutdownBare            lines 91,93; port 92 (its port command
//	                                      is "first-listen-port", not
//	                                      "listen-port" -- see isPortCommand)
//	web-transport shutdownBare            lines 99,103; ports 100,102
//	mqtt          shutdownBare            line 113 (no port line in this
//	                                      capture -- broker-level mqtt exposes no
//	                                      listen-port command here)
//	amqp          shutdownBare            lines 116,118; port 117
//	health-check  shutdownBare            lines 121,124; ports 122-123
//
// Use ServiceShutdownStyles to read this table from a test without exporting
// shutdownStyle itself.
var serviceShutdownStyle = map[string]shutdownStyle{
	"semp":          shutdownTransportAfter,
	"rest":          shutdownTransportBefore,
	"smf":           shutdownBare,
	"mate-link":     shutdownBare,
	"redundancy":    shutdownBare,
	"web-transport": shutdownBare,
	"mqtt":          shutdownBare,
	"amqp":          shutdownBare,
	"health-check":  shutdownBare,
}

// String renders a shutdownStyle as the label ServiceShutdownStyles hands a
// test, so the exported accessor need not leak this unexported type.
func (s shutdownStyle) String() string {
	switch s {
	case shutdownTransportAfter:
		return "transport-after"
	case shutdownTransportBefore:
		return "transport-before"
	default:
		return "bare"
	}
}

// ServiceShutdownStyles returns the service -> style table InjectShutdown
// consults, as service name -> one of "bare", "transport-after",
// "transport-before" (shutdownStyle.String), so a test can assert the table's
// contents (which services are known, and how each spells its shutdown
// command) without a capture round-trip or a reference to the unexported
// shutdownStyle type.
func ServiceShutdownStyles() map[string]string {
	out := make(map[string]string, len(serviceShutdownStyle))
	for k, v := range serviceShutdownStyle {
		out[k] = v.String()
	}
	return out
}

// isPortCommand reports whether tok is a service's port-setting command. A
// suffix match (rather than an exact "listen-port") is what also catches
// redundancy's "first-listen-port" (semp/all.cli:92) without a second table.
func isPortCommand(tok string) bool {
	return strings.HasSuffix(tok, "listen-port")
}

// lineRole classifies one tokenised "service" line for a KNOWN service.
type lineRole int

const (
	roleOther    lineRole = iota // recognised service, but neither shutdown nor port
	roleShutdown                 // a shutdown or "no ... shutdown" control line
	rolePort                     // a listen-port / first-listen-port line
)

// serviceLine is what parseServiceLine reads out of one Block's opener.
type serviceLine struct {
	known       bool     // service appears in serviceShutdownStyle
	negated     bool     // opener began with "no"
	service     string   // the token right after "service"
	transport   string   // "" for a bare-style service or a bare-shaped line
	role        lineRole // meaningful only when known
	unknownPort bool     // an unknown service's line looks like a port change
}

// parseServiceLine tokenises b's opener into a serviceLine. It reads Lines[0]
// directly rather than trusting b.Kind, because a negated opener's Kind is the
// literal word "no" (see the "no" TRAP note above) -- checking b.Kind=="service"
// here would silently drop every trailing "no service ..." line. strings.Fields
// is what makes this tolerant of the double spaces real captures carry (e.g.
// semp/all.cli:80 `service smf listen-port "55003"  "compressed"`), and each
// field is unquoted so a bare and a quoted operand compare equal.
func parseServiceLine(b Block) serviceLine {
	if len(b.Lines) == 0 {
		return serviceLine{}
	}
	// Block.Opener(), not Lines[0]: a held pragma is prepended to Lines, so Lines[0]
	// is a `! pragma:` comment on exactly the blocks that carry one. Reading it as
	// the opener made those blocks parse as unrecognised and be skipped with no
	// injection and no warning -- silently, since an unrecognised line is a normal
	// thing for this function to see.
	f := fieldsUnquoted(b.Opener())
	negated := false
	if len(f) > 0 && f[0] == "no" {
		negated = true
		f = f[1:]
	}
	if len(f) < 2 || f[0] != "service" {
		return serviceLine{}
	}
	service := f[1]
	rest := f[2:]

	style, known := serviceShutdownStyle[service]
	if !known {
		return serviceLine{negated: negated, service: service, unknownPort: containsPortCommand(rest)}
	}
	role, transport := classifyServiceRest(style, service, rest)
	return serviceLine{known: true, negated: negated, service: service, transport: transport, role: role}
}

// injectedBlock builds a synthesised shutdown line that classifies exactly as the
// parser would, and inherits the REGION of the block it wraps.
//
// The region matters and its absence was a real bug: newBlock does not set Region
// (parseBody assigns it separately), so an injected block defaulted to the empty
// region and was then dropped by the renderer's region filter -- the injection
// silently did not reach the broker, while still appearing in the artifact the
// verification diff compares against.
func injectedBlock(around Block, section, opener string) Block {
	b := newBlock(section, 0, opener)
	b.Region = around.Region
	return b
}

// fieldsUnquoted splits a line on whitespace (collapsing runs of it, as
// strings.Fields already does) and strips a surrounding pair of double quotes
// from each token, so `"55003"` and `55003` tokenise the same way.
func fieldsUnquoted(line string) []string {
	fields := strings.Fields(line)
	out := make([]string, len(fields))
	for i, f := range fields {
		out[i] = strings.Trim(f, `"`)
	}
	return out
}

// containsPortCommand reports whether any token looks like a port-setting
// command. It is used only for an UNKNOWN service, where the style (and so the
// exact token position) is not known -- this is what makes an unknown service's
// port line discoverable as a warning instead of silently passed through with
// no signal at all.
func containsPortCommand(rest []string) bool {
	for _, t := range rest {
		if isPortCommand(t) {
			return true
		}
	}
	return false
}

// defaultTransport is the transport an untagged line means, for the services whose
// shutdown carries one.
//
// It exists because of a real mis-scoping. semp pairs "service semp shutdown
// plain-text" (semp/all.cli:65) with the UNTAGGED "service semp listen-port 8080"
// (:66), while its ssl half tags both (:68, :69). Reading the untagged port line as
// transport "" put it in a different key from the shutdown line that actually
// guards it, with two consequences: the existing, already-safe sandwich was not
// recognised, and a case-3 injection rendered the BARE "service semp shutdown" --
// semp's whole-service form -- which would have taken the ssl transport down too.
// That is precisely the granularity failure the four-case contract exists to avoid.
//
// Only shutdownTransportAfter services need an entry; rest (shutdownTransportBefore)
// tags both of its transports explicitly and has no untagged form in the capture.
var defaultTransport = map[string]string{
	"semp": "plain-text",
}

// classifyServiceRest reads the tokens after "service <name>" (already stripped
// of a leading "no") for a service whose style is known, and reports whether the
// line is a shutdown control line or a port-setting line, and the transport it
// carries. rest that matches neither shape -- e.g. semp's
// "cors allow-any-host" or "session-idle-timeout 3600" -- comes back roleOther
// and is passed through untouched regardless of style.
func classifyServiceRest(style shutdownStyle, service string, rest []string) (lineRole, string) {
	// An untagged line means the service default, not "no transport" -- see
	// defaultTransport for why conflating the two mis-scoped the injection.
	dflt := defaultTransport[service]
	switch style {
	case shutdownTransportAfter:
		// semp/all.cli:65 "shutdown plain-text", :66 "listen-port 8080",
		// :69 "listen-port 1943 ssl" -- transport, when present, trails.
		if len(rest) >= 1 && rest[0] == "shutdown" {
			if len(rest) >= 2 {
				return roleShutdown, rest[1]
			}
			return roleShutdown, dflt
		}
		if len(rest) >= 1 && isPortCommand(rest[0]) {
			if len(rest) >= 3 {
				return rolePort, rest[len(rest)-1]
			}
			return rolePort, dflt
		}

	case shutdownTransportBefore:
		// semp/all.cli:109 "incoming shutdown", :110 "outgoing shutdown" --
		// transport leads. No broker-level rest listen-port line is in the
		// capture to confirm the port-line shape; by symmetry with the shutdown
		// line it is assumed to lead the same way ("rest incoming listen-port
		// ..."), with a bare fallback kept for tolerance since it is unverified.
		if len(rest) >= 2 && rest[1] == "shutdown" {
			return roleShutdown, rest[0]
		}
		if len(rest) >= 2 && isPortCommand(rest[1]) {
			return rolePort, rest[0]
		}
		if len(rest) >= 1 && isPortCommand(rest[0]) {
			return rolePort, ""
		}

	default: // shutdownBare
		// semp/all.cli:78/83 (smf), 86/88 (mate-link), 91/93 (redundancy),
		// 99/103 (web-transport), 113 (mqtt), 116/118 (amqp), 121/124
		// (health-check) -- no transport component either line.
		if len(rest) == 1 && rest[0] == "shutdown" {
			return roleShutdown, ""
		}
		if len(rest) >= 1 && isPortCommand(rest[0]) {
			return rolePort, ""
		}
	}
	return roleOther, ""
}

// shutdownLine renders the shutdown (or, negated, un-shutdown) command for
// service+transport in the given style. transport is ignored for shutdownBare,
// matching every bare-style example in serviceShutdownStyle's evidence.
func shutdownLine(style shutdownStyle, service, transport string, negate bool) string {
	var body string
	switch style {
	case shutdownTransportAfter:
		if transport == "" {
			body = fmt.Sprintf("service %s shutdown", service)
		} else {
			body = fmt.Sprintf("service %s shutdown %s", service, transport)
		}
	case shutdownTransportBefore:
		if transport == "" {
			body = fmt.Sprintf("service %s shutdown", service)
		} else {
			body = fmt.Sprintf("service %s %s shutdown", service, transport)
		}
	default:
		body = fmt.Sprintf("service %s shutdown", service)
	}
	if negate {
		return "no " + body
	}
	return body
}

// svcKey identifies one shutdown sandwich: a service+transport within one
// section. Section is part of the key (rather than service+transport alone) to
// hold the search to "the same block/section" the task describes -- in every
// observed capture a section holds exactly one service, so this is currently
// equivalent to keying on service+transport, but it costs nothing to be precise
// about the scope rather than relying on that always being true.
type svcKey struct{ section, service, transport string }

// span tracks one key's port-line extent and the real (never injected)
// shutdown/un-shutdown lines the scan found for it, by block index in the input
// slice.
type span struct {
	firstPort, lastPort int // -1 until a non-negated port line is seen
	leading             []int
	trailing            []int
}

// InjectShutdown makes a captured "service X listen-port ..." change safe to
// import into a broker that requires the service to be shut down first, without
// ever inventing a final enablement state the artifact itself does not attest to
// (see the file-level comment for the four cases and why case 3 is the only one
// that decides anything, and decides it as a no-op against the target).
//
// targetState reads the target's current enablement for one service+transport;
// pass nil when no target is reachable (case 3 then always takes its "unknown"
// branch, warning rather than guessing). It is an injected seam precisely so the
// read-and-restore case is testable with no live broker.
//
// The returned error is reserved for a systemic failure this function cannot
// actually produce today: every Block a real ParseBlocks capture can hand it is
// either transformed or passed through with a warning, never aborted -- a single
// unrecognised or ambiguous line is skipped, not a reason to fail the whole
// transform (this package's fail-loud-vs-skip rule). It is part of the signature
// for symmetry with the other Capture-consuming transforms in this package and
// so a future genuinely-systemic check (e.g. an internally contradictory span)
// has somewhere to report to without a breaking signature change.
func InjectShutdown(blocks []Block, targetState func(service, transport string) (enabled bool, known bool)) ([]Block, []string, error) {
	read := targetState
	if read == nil {
		read = func(string, string) (bool, bool) { return false, false }
	}

	spans := map[svcKey]*span{}
	warnAt := map[int][]string{}

	for i, b := range blocks {
		li := parseServiceLine(b)

		if !li.known {
			if li.unknownPort {
				warnAt[i] = append(warnAt[i], fmt.Sprintf(
					"service %q is not in the shutdown table -- port line passed through unchanged: %q",
					li.service, strings.TrimSpace(b.Lines[0])))
			}
			continue
		}
		if li.role == roleOther {
			continue
		}

		k := svcKey{section: b.Section, service: li.service, transport: li.transport}
		sp := spans[k]
		if sp == nil {
			sp = &span{firstPort: -1, lastPort: -1}
			spans[k] = sp
		}
		switch li.role {
		case rolePort:
			if !li.negated {
				if sp.firstPort == -1 {
					sp.firstPort = i
				}
				sp.lastPort = i
			}
		case roleShutdown:
			if li.negated {
				sp.trailing = append(sp.trailing, i)
			} else {
				sp.leading = append(sp.leading, i)
			}
		}
	}

	leadInjectAt := map[int]Block{}
	trailInjectAt := map[int]Block{}

	for k, sp := range spans {
		if sp.firstPort == -1 {
			// No non-negated port line for this key -- e.g. rest's shutdown pair
			// in semp/all.cli:109-110, which has no matching broker-level
			// listen-port line in this capture. Nothing to protect, nothing to do.
			continue
		}
		style := serviceShutdownStyle[k.service]

		hasLeading := false
		for _, idx := range sp.leading {
			if idx < sp.firstPort {
				hasLeading = true
				break
			}
		}
		if hasLeading {
			continue // case 1: already sandwiched
		}

		hasTrailing := false
		for _, idx := range sp.trailing {
			if idx > sp.lastPort {
				hasTrailing = true
				break
			}
		}
		if hasTrailing {
			// case 2: the artifact's own trailing line excuses us from
			// synthesising one; only the leading needs injecting.
			leadInjectAt[sp.firstPort] = injectedBlock(blocks[sp.firstPort], k.section,
				shutdownLine(style, k.service, k.transport, false))
			continue
		}

		// case 3: neither end is real. Ask the target rather than guess.
		enabled, known := read(k.service, k.transport)
		if !known {
			warnAt[sp.firstPort] = append(warnAt[sp.firstPort], fmt.Sprintf(
				"%s has no shutdown sandwich and no reachable target state -- port line(s) passed through "+
					"unchanged; shut the service down manually first, or supply targetState", describeKey(k)))
			continue
		}
		leadInjectAt[sp.firstPort] = injectedBlock(blocks[sp.firstPort], k.section,
			shutdownLine(style, k.service, k.transport, false))
		trailInjectAt[sp.lastPort] = injectedBlock(blocks[sp.lastPort], k.section,
			shutdownLine(style, k.service, k.transport, enabled))
	}

	out := make([]Block, 0, len(blocks)+len(leadInjectAt)+len(trailInjectAt))
	for i, b := range blocks {
		if lead, ok := leadInjectAt[i]; ok {
			out = append(out, lead)
		}
		out = append(out, b)
		if trail, ok := trailInjectAt[i]; ok {
			out = append(out, trail)
		}
	}
	for i := range out {
		out[i].Index = i + 1
	}

	var warnings []string
	for i := range blocks {
		warnings = append(warnings, warnAt[i]...)
	}

	return out, warnings, nil
}

// describeKey names a span's identity for a warning message: the bare service
// for a shutdownBare key, or "service X transport Y" when a transport applies.
func describeKey(k svcKey) string {
	if k.transport == "" {
		return fmt.Sprintf("service %q", k.service)
	}
	return fmt.Sprintf("service %q transport %q", k.service, k.transport)
}

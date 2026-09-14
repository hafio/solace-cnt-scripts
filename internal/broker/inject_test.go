package broker

import (
	"reflect"
	"strings"
	"testing"
)

// inject_test.go covers InjectShutdown (inject.go) against the four-case
// contract in that file's own doc comment: a captured "service X listen-port
// ..." change is wrapped in a shutdown sandwich only when doing so cannot
// invent a final enablement state the artifact never attested to. Every
// fixture below is either transcribed verbatim from semp/all.cli (cited by
// line number, service-config lines only -- this run carries no credential)
// or built to match its shape by hand for a scenario the real capture does
// not happen to exercise (an un-sandwiched port line, an unknown service).
//
// realBrokerServiceSection is reused by more than one test, so it is parsed
// fresh via parseInjectCapture in each -- InjectShutdown never mutates its
// input, but a shared *parsed* slice would still invite one test's assertion
// to depend on another's read of it.

// parseInjectCapture wraps body (broker-level "! Configure X Service:" lines,
// real or hand-built to match their shape) in the minimal preamble and
// "! END" terminator ParseBlocks requires, and returns the parsed blocks.
// Going through the real parser -- rather than hand-building Block literals --
// is what keeps every case honest about Section assignment and the "no" TRAP
// newBlock/parseServiceLine both navigate (inject.go's file comment): a test
// that built Blocks by hand would only be asserting against its own guess at
// what ParseBlocks does with a negated opener, not the real thing.
func parseInjectCapture(t *testing.T, body string) []Block {
	t.Helper()
	const header = " ! BEGIN\n" +
		"! Command: \"show current-config all\"\n" +
		"! SEMP Schema Version: \"broker/10.26.0.8827\"\n" +
		"! Redacted: \"False\"\n" +
		"\n" +
		"home\n" +
		"enable\n" +
		"configure\n" +
		"\n"
	raw := header + body + "\n! END\n"
	c, err := ParseBlocks([]byte(raw))
	if err != nil {
		t.Fatalf("ParseBlocks(inject fixture): %v\nbody:\n%s", err, body)
	}
	return c.Blocks
}

// linesOf collects each block's opener (Lines[0]) for a compact diff against
// an expected sequence -- what changed is easier to see as a list of command
// strings than as a list of Block structs.
func linesOf(blocks []Block) []string {
	out := make([]string, len(blocks))
	for i, b := range blocks {
		out[i] = b.Lines[0]
	}
	return out
}

// enablementCounts tallies, per (section, service, transport), how many real
// shutdown lines a block slice carries: index 0 counts a leading
// "service X shutdown" and index 1 counts a trailing "no service X shutdown".
// TestInjectShutdownNeverChangesEnablement compares this before and after the
// transform instead of eyeballing one case -- the risk it guards is
// inject.go's own: a sandwich that always closed with "no ... shutdown" would
// silently enable a service the artifact means to leave disabled, and that
// would show up here as a trailing count that grew without the leading count
// growing to match.
func enablementCounts(blocks []Block) map[svcKey][2]int {
	out := map[svcKey][2]int{}
	for _, b := range blocks {
		li := parseServiceLine(b)
		if !li.known || li.role != roleShutdown {
			continue
		}
		k := svcKey{section: b.Section, service: li.service, transport: li.transport}
		c := out[k]
		if li.negated {
			c[1]++
		} else {
			c[0]++
		}
		out[k] = c
	}
	return out
}

// realBrokerServiceSection is semp/all.cli lines 64-124 verbatim: the whole
// broker-level service-configuration run, transcribed by hand per this
// package's rule that the real capture (which carries admin/LDAP/redundancy
// secrets elsewhere in the file) is never copied wholesale. None of these
// lines carry a credential. It gives TestInjectShutdownAlreadySandwichedPassesThroughUnchanged
// and TestInjectShutdownNeverChangesEnablement every style in
// serviceShutdownStyle at once: semp (transport-after, both transports),
// smf/mate-link/redundancy/web-transport/amqp/health-check (bare, several
// with the GROUPING shape of one shutdown pair around more than one port
// line), rest and mqtt (transport-before / bare with no broker-level port
// line at all, so nothing to protect).
const realBrokerServiceSection = `! Configure SEMP Service:
service semp shutdown plain-text
service semp listen-port 8080
no service semp shutdown plain-text
service semp shutdown ssl
service semp listen-port 1943 ssl
no service semp shutdown ssl
service semp cors allow-any-host
no service semp legacy-timeout
service semp session-idle-timeout 3600
service semp session-max-lifetime 43200

! Configure SMF Service:
service smf event connections thresholds set-percentage 80 clear-percentage 60
service smf shutdown
service smf listen-port "55555"
service smf listen-port "55003"  "compressed"
service smf listen-port "55556"  "routing"
service smf listen-port "55443"  "ssl"
no service smf shutdown

! Configure Matelink Service:
service mate-link shutdown
service mate-link listen-port "8741"
no service mate-link shutdown

! Configure Redundancy Service:
service redundancy shutdown
service redundancy first-listen-port "8300"
no service redundancy shutdown

! Configure SSL Service:
service ssl event connections thresholds set-percentage 80 clear-percentage 60

! Configure WEB Service
service web-transport shutdown
service web-transport listen-port 8008
no service web-transport web-url-suffix
service web-transport listen-port "1443"  "ssl"
no service web-transport shutdown

! Configure Virtual Hostnames:

! Configure REST Service:
service rest event outgoing connections thresholds set-percentage 80 clear-percentage 60
no service rest incoming shutdown
no service rest outgoing shutdown

! Configure MQTT Service:
no service mqtt shutdown

! Configure AMQP Service:
service amqp shutdown
service amqp listen-port "0" ssl
no service amqp shutdown

! Configure HealthCheck Service:
service health-check shutdown
service health-check listen-port 5550
service health-check listen-port 5553 ssl
no service health-check shutdown`

// TestInjectShutdownAlreadySandwichedPassesThroughUnchanged covers case 1:
// every port change in the real broker-level run already sits inside a real
// shutdown/un-shutdown pair (or, for rest/mqtt, has no broker-level port line
// at all), so nothing may be injected and no block's content may change --
// the risk is a false positive on "already safe" that duplicates a shutdown
// the artifact already carries.
//
// targetState is nil (case 3's "no target reachable" branch) because case 1
// must not need one: if detecting an existing sandwich required asking the
// target, a target-less run of this exact capture -- the common case, since
// it is already safe -- would needlessly warn or, worse, inject.
//
// Warnings are deliberately not asserted here. This fixture also exercises a
// separate, narrower gap around semp's untagged plain-text port line
// (TestInjectShutdownBUGPlainTextEmitsBareShutdown below); pinning an exact
// warning count in THIS test would make it fail the moment that gap is
// fixed, for a reason unrelated to the property this test protects.
func TestInjectShutdownAlreadySandwichedPassesThroughUnchanged(t *testing.T) {
	blocks := parseInjectCapture(t, realBrokerServiceSection)
	out, _, err := InjectShutdown(blocks, nil)
	if err != nil {
		t.Fatalf("InjectShutdown: %v", err)
	}
	wantLines := linesOf(blocks)
	gotLines := linesOf(out)
	if !reflect.DeepEqual(gotLines, wantLines) {
		t.Fatalf("already-sandwiched capture was rewritten:\ngot  %v\nwant %v", gotLines, wantLines)
	}
}

// TestInjectShutdownNeverChangesEnablement is the central guard: for every
// (section, service, transport) the real capture describes, the shutdown /
// un-shutdown counts implied by the blocks AFTER the transform must equal
// what the artifact specified BEFORE it, counted rather than eyeballed via
// enablementCounts. Run against the same real, already-sandwiched fixture as
// the test above, this is deliberately a second, independent way of stating
// that property (by tally, not by line-for-line diff) so a future change to
// the grouping/keying logic that happened to preserve line COUNT but not
// per-key PAIRING would still be caught.
func TestInjectShutdownNeverChangesEnablement(t *testing.T) {
	blocks := parseInjectCapture(t, realBrokerServiceSection)
	out, _, err := InjectShutdown(blocks, nil)
	if err != nil {
		t.Fatalf("InjectShutdown: %v", err)
	}
	before := enablementCounts(blocks)
	after := enablementCounts(out)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("enablement counts changed:\nbefore %v\nafter  %v", before, after)
	}
	if len(before) == 0 {
		t.Fatal("fixture produced no known shutdown-role lines; test is not exercising anything")
	}
}

// TestInjectShutdownInjectsLeadingWhenTrailingExists covers case 2: no
// leading shutdown, but a real trailing "no ... shutdown" for the same key
// exists later in the section. Only the leading line may be synthesised --
// the trailing line is the artifact's own and a correct transform must leave
// it byte-for-byte alone rather than, say, re-emitting an equivalent one.
// Uses semp's ssl transport (semp/all.cli:69-70 shape) rather than its
// plain-text one, which is covered separately (see the BUG test below).
func TestInjectShutdownInjectsLeadingWhenTrailingExists(t *testing.T) {
	blocks := parseInjectCapture(t, `! Configure SEMP Service:
service semp listen-port 1943 ssl
no service semp shutdown ssl`)
	out, warnings, err := InjectShutdown(blocks, nil)
	if err != nil {
		t.Fatalf("InjectShutdown: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	want := []string{
		"service semp shutdown ssl",       // injected leading
		"service semp listen-port 1943 ssl", // untouched port line
		"no service semp shutdown ssl",    // untouched, the artifact's own trailing
	}
	if got := linesOf(out); !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %v\nwant %v", got, want)
	}
}

// TestInjectShutdownCaseThreeRestoresTargetState covers case 3 in both
// directions -- the property the file comment calls "the whole point of the
// function": a sandwich-free port change must inject a restore line matching
// whatever the target reports, NOT always "no ... shutdown". Getting this
// backwards (or defaulting to "enabled") would silently turn a service the
// target means to leave disabled back on.
//
// Uses smf's four real listen-port lines with no wrapper at all
// (semp/all.cli:79-82, the "older broker" bare-middle-line shape the file
// comment opens with), which also exercises GROUPING: one pair is injected
// once, at the first and last port line of the run, not once per line.
func TestInjectShutdownCaseThreeRestoresTargetState(t *testing.T) {
	const smfPorts = `! Configure SMF Service:
service smf listen-port "55555"
service smf listen-port "55003"  "compressed"
service smf listen-port "55556"  "routing"
service smf listen-port "55443"  "ssl"`

	cases := []struct {
		name        string
		enabled     bool
		wantTrailer string
	}{
		{"target reports enabled -> restore un-shuts it down", true, "no service smf shutdown"},
		{"target reports disabled -> restore leaves it shut down", false, "service smf shutdown"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			blocks := parseInjectCapture(t, smfPorts)
			target := func(service, transport string) (bool, bool) {
				if service == "smf" && transport == "" {
					return c.enabled, true
				}
				return false, false
			}
			out, warnings, err := InjectShutdown(blocks, target)
			if err != nil {
				t.Fatalf("InjectShutdown: %v", err)
			}
			if len(warnings) != 0 {
				t.Fatalf("unexpected warnings: %v", warnings)
			}
			want := []string{
				"service smf shutdown", // injected leading, once, before the first port line
				`service smf listen-port "55555"`,
				`service smf listen-port "55003"  "compressed"`,
				`service smf listen-port "55556"  "routing"`,
				`service smf listen-port "55443"  "ssl"`,
				c.wantTrailer, // injected trailing, once, after the last port line
			}
			if got := linesOf(out); !reflect.DeepEqual(got, want) {
				t.Fatalf("got  %v\nwant %v", got, want)
			}
		})
	}
}

// TestInjectShutdownCaseThreeUnknownTargetLeavesPortUnchanged covers case 3's
// "unknown" branch, both ways it is reached (nil targetState, and a
// targetState that answers but reports known=false): the risk named in
// inject.go is an outage, so the safe failure here is to change NOTHING and
// warn, never to guess. Also carries two of the three real quoting/spacing
// variants named in the task -- amqp's quoted-port/unquoted-transport line
// (semp/all.cli:117) and redundancy's "first-listen-port" command, which
// isPortCommand must still recognise as a port line via its suffix match.
func TestInjectShutdownCaseThreeUnknownTargetLeavesPortUnchanged(t *testing.T) {
	cases := []struct {
		name        string
		section     string
		line        string
		serviceName string
		target      func(service, transport string) (bool, bool)
	}{
		{
			name:        "nil targetState",
			section:     "! Configure AMQP Service: ",
			line:        `service amqp listen-port "0" ssl`,
			serviceName: "amqp",
			target:      nil,
		},
		{
			name:        "targetState answers known=false",
			section:     "! Configure Redundancy Service: ",
			line:        `service redundancy first-listen-port "8300"`,
			serviceName: "redundancy",
			target:      func(string, string) (bool, bool) { return false, false },
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			blocks := parseInjectCapture(t, c.section+"\n"+c.line)
			out, warnings, err := InjectShutdown(blocks, c.target)
			if err != nil {
				t.Fatalf("InjectShutdown: %v", err)
			}
			if got := linesOf(out); !reflect.DeepEqual(got, []string{c.line}) {
				t.Fatalf("port line was rewritten with no known target state: got %v, want unchanged %v", got, []string{c.line})
			}
			if len(warnings) != 1 {
				t.Fatalf("want exactly one warning, got %v", warnings)
			}
			if !strings.Contains(warnings[0], c.serviceName) {
				t.Errorf("warning %q does not name the affected service %q", warnings[0], c.serviceName)
			}
			if !strings.Contains(warnings[0], "no reachable target state") {
				t.Errorf("warning %q does not explain why nothing was injected", warnings[0])
			}
		})
	}
}

// TestInjectShutdownUnknownServicePassesThroughWithWarning covers case 4: a
// service absent from serviceShutdownStyle is passed through unchanged with a
// warning that names it, rather than silently dropped (which would hide a
// port change the broker might reject) or guessed at (which risks the same
// outage case 3 exists to avoid).
func TestInjectShutdownUnknownServicePassesThroughWithWarning(t *testing.T) {
	const line = `service turbo-encabulator listen-port 9999`
	blocks := parseInjectCapture(t, "! Configure Turbo Encabulator Service: \n"+line)
	out, warnings, err := InjectShutdown(blocks, nil)
	if err != nil {
		t.Fatalf("InjectShutdown: %v", err)
	}
	if got := linesOf(out); !reflect.DeepEqual(got, []string{line}) {
		t.Fatalf("unknown-service port line was rewritten: got %v, want unchanged %v", got, []string{line})
	}
	if len(warnings) != 1 {
		t.Fatalf("want exactly one warning, got %v", warnings)
	}
	if !strings.Contains(warnings[0], "turbo-encabulator") {
		t.Errorf("warning %q does not name the unknown service", warnings[0])
	}
	if !strings.Contains(warnings[0], "not in the shutdown table") {
		t.Errorf("warning %q does not explain why the line was passed through", warnings[0])
	}
}

// TestInjectShutdownTransportGranularitySemp covers property 6 for semp's
// ssl transport: injecting ssl's shutdown must not touch a plain-text port
// change that already has its own, separate, real sandwich (semp/all.cli:
// 65-67) sitting right next to it in the same section. If InjectShutdown ever
// keyed a shutdown pair on service alone (ignoring transport), the plain-text
// trio here would either be treated as ssl's own leading/trailing pair (case
// 1 firing on the wrong lines) or get a second, redundant wrap of its own --
// either way the plain-text trio would not survive byte-for-byte, which is
// exactly what this test checks.
//
// The reverse direction -- injecting a FRESH shutdown for the plain-text
// transport without disturbing ssl -- is not provable this way: see
// TestInjectShutdownBUGPlainTextEmitsBareShutdown, which documents why.
func TestInjectShutdownTransportGranularitySemp(t *testing.T) {
	blocks := parseInjectCapture(t, `! Configure SEMP Service:
service semp shutdown plain-text
service semp listen-port 8080
no service semp shutdown plain-text
service semp listen-port 1943 ssl
no service semp shutdown ssl`)
	out, _, err := InjectShutdown(blocks, nil)
	if err != nil {
		t.Fatalf("InjectShutdown: %v", err)
	}
	want := []string{
		"service semp shutdown plain-text",
		"service semp listen-port 8080",
		"no service semp shutdown plain-text",
		"service semp shutdown ssl", // injected -- carries "ssl", not "plain-text" and not bare
		"service semp listen-port 1943 ssl",
		"no service semp shutdown ssl",
	}
	if got := linesOf(out); !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %v\nwant %v", got, want)
	}
}

// TestInjectShutdownTransportGranularityRest covers property 6 for rest's
// transport-before style: incoming and outgoing each need their own
// shutdown/restore around their own port line, with no leading/trailing
// sandwich in the capture to key off (rest has none in semp/all.cli --
// 108-110 are trailing-only, matched by
// TestInjectShutdownAlreadySandwichedPassesThroughUnchanged's fixture). Also
// gives transport-before its own case-3-both-directions check, mirroring
// TestInjectShutdownCaseThreeRestoresTargetState for transport-after.
func TestInjectShutdownTransportGranularityRest(t *testing.T) {
	blocks := parseInjectCapture(t, `! Configure REST Service:
service rest incoming listen-port 1900
service rest outgoing listen-port 1901`)
	target := func(service, transport string) (bool, bool) {
		switch {
		case service == "rest" && transport == "incoming":
			return true, true // ends enabled
		case service == "rest" && transport == "outgoing":
			return false, true // ends disabled
		}
		return false, false
	}
	out, warnings, err := InjectShutdown(blocks, target)
	if err != nil {
		t.Fatalf("InjectShutdown: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	want := []string{
		"service rest incoming shutdown",
		"service rest incoming listen-port 1900",
		"no service rest incoming shutdown", // incoming ends enabled
		"service rest outgoing shutdown",
		"service rest outgoing listen-port 1901",
		"service rest outgoing shutdown", // outgoing ends disabled -- NOT negated
	}
	if got := linesOf(out); !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %v\nwant %v", got, want)
	}
}

// TestInjectShutdownParsesRealQuotingVariants pins parseServiceLine's
// tokenisation directly against the three real quoting/spacing shapes the
// task names, so a regression here fails as a tokenisation bug rather than
// only surfacing indirectly as a missed sandwich several tests away. Lines
// transcribed verbatim from semp/all.cli (80, 69, 117).
func TestInjectShutdownParsesRealQuotingVariants(t *testing.T) {
	cases := []struct {
		name          string
		line          string
		wantService   string
		wantTransport string
	}{
		{"smf: double-quoted, double-spaced operands (all.cli:80)",
			`service smf listen-port "55003"  "compressed"`, "smf", ""},
		{"semp: unquoted port with a trailing transport keyword (all.cli:69)",
			`service semp listen-port 1943 ssl`, "semp", "ssl"},
		{"amqp: quoted port, unquoted trailing token (all.cli:117)",
			`service amqp listen-port "0" ssl`, "amqp", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := newBlock("Configure Service", 1, c.line)
			got := parseServiceLine(b)
			if !got.known {
				t.Fatalf("parseServiceLine(%q).known = false, want true", c.line)
			}
			if got.role != rolePort {
				t.Fatalf("parseServiceLine(%q).role = %v, want rolePort", c.line, got.role)
			}
			if got.service != c.wantService || got.transport != c.wantTransport {
				t.Errorf("parseServiceLine(%q) = {service:%q transport:%q}, want {%q %q}",
					c.line, got.service, got.transport, c.wantService, c.wantTransport)
			}
		})
	}
}

// TestInjectShutdownServiceShutdownStylesTable pins ServiceShutdownStyles'
// documented contents -- the exact set of services InjectShutdown knows and
// how each spells its shutdown command. Every entry here is exercised by a
// test above: semp by TestInjectShutdownAlreadySandwichedPassesThroughUnchanged,
// TestInjectShutdownInjectsLeadingWhenTrailingExists,
// TestInjectShutdownTransportGranularitySemp and the BUG test below; rest by
// TestInjectShutdownTransportGranularityRest; smf by
// TestInjectShutdownCaseThreeRestoresTargetState; redundancy by
// TestInjectShutdownCaseThreeUnknownTargetLeavesPortUnchanged; and
// mate-link/web-transport/mqtt/amqp/health-check by their place in
// realBrokerServiceSection, read by the first two tests in this file. A
// service silently added to or dropped from the map without a matching test
// would still fail this one, since it changes what the map returns.
func TestInjectShutdownServiceShutdownStylesTable(t *testing.T) {
	want := map[string]string{
		"semp":          "transport-after",
		"rest":          "transport-before",
		"smf":           "bare",
		"mate-link":     "bare",
		"redundancy":    "bare",
		"web-transport": "bare",
		"mqtt":          "bare",
		"amqp":          "bare",
		"health-check":  "bare",
	}
	got := ServiceShutdownStyles()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ServiceShutdownStyles() = %v, want %v", got, want)
	}
}

// TestInjectShutdownBUGPlainTextEmitsBareShutdown documents a real defect
// rather than intended behaviour (see this file's package-level convention,
// established in blocks_test.go: a test named BUG asserts the CORRECT
// behaviour and is expected to fail against the current implementation).
//
// EVIDENCE: semp/all.cli's own real sandwich for the untagged transport is
//
//	service semp shutdown plain-text   (line 65)
//	service semp listen-port 8080      (line 66)
//	no service semp shutdown plain-text (line 67)
//
// -- the broker requires the literal word "plain-text" on the shutdown
// command even though the matching port line carries no transport suffix at
// all. classifyServiceRest's shutdownTransportAfter case reads that "plain-text"
// token as a real transport value (transport="plain-text") for the shutdown
// line, but the untagged port line (fewer than 3 fields) reports
// transport="" -- so the two never share a svcKey, and a case-3 injection
// for the plain-text port asks shutdownLine for transport="", which its
// shutdownTransportAfter branch renders as the BARE "service semp shutdown"
// (inject.go's own shutdownLine: `if transport == "" { body = "service %s
// shutdown" }`). A bare "service semp shutdown" is semp's whole-service
// shutdown, not a plain-text-only one -- exactly the outage property 6
// (transport granularity) exists to prevent: changing the plain-text port
// would take the ssl transport down with it too.
//
// This is reachable with no ambiguity about "already sandwiched": the input
// below is a fresh, isolated plain-text port line with no wrapper at all
// (the "older broker" shape), so case 3 is the only path and targetState is
// answered, so no warning suppresses it.
func TestInjectShutdownUntaggedSempTransportNamesPlainText(t *testing.T) {
	blocks := parseInjectCapture(t, `! Configure SEMP Service:
service semp listen-port 8080`)
	// The stub keys on "plain-text", not "". That IS the fix under test: semp's
	// untagged `listen-port 8080` (semp/all.cli:66) is the plain-text transport, and
	// parsing it as "no transport" put it in a different svcKey from the
	// `service semp shutdown plain-text` line that guards it -- so the real sandwich
	// went unrecognised AND a case-3 injection rendered the bare, whole-service
	// `service semp shutdown`, taking the ssl transport down with it. See
	// defaultTransport in inject.go.
	target := func(service, transport string) (bool, bool) {
		if service == "semp" && transport == "plain-text" {
			return true, true
		}
		return false, false
	}
	out, warnings, err := InjectShutdown(blocks, target)
	if err != nil {
		t.Fatalf("InjectShutdown: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	// Correct behaviour: the injected pair names the plain-text transport
	// explicitly, matching the real broker's own syntax, and so leaves the
	// ssl transport (not present in this block at all, let alone shut down)
	// completely out of it.
	want := []string{
		"service semp shutdown plain-text",
		"service semp listen-port 8080",
		"no service semp shutdown plain-text",
	}
	if got := linesOf(out); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v -- a bare `service semp shutdown` here would take the ssl "+
			"transport down alongside plain-text, which is the granularity failure the "+
			"four-case contract exists to prevent", got, want)
	}
}

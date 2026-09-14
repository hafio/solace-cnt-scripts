package broker

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// transform_test.go covers internal/broker/transform.go: the replay
// transformations ported from semp/CurrentConfigParser.html, the operator's
// hand-built migration GUI.
//
// Where a rule exists because of something in the REAL capture, the test uses the
// real capture (semp/all.cli) rather than a hand-made block, because the point of
// porting these rules is that the hand-made version of the input is exactly what
// missed them for so long.

// loadRealCapture parses semp/all.cli, the ground-truth capture from a live broker.
// It skips rather than fails when absent: the file is gitignored (it carries real
// credentials), so a fresh clone legitimately does not have it.
func loadRealCapture(t *testing.T) *Capture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "semp", "all.cli"))
	if err != nil {
		t.Skipf("real capture not available (%v); the shape assertions below need it", err)
	}
	c, err := ParseBlocks(raw)
	if err != nil {
		t.Fatalf("parse the real capture: %v", err)
	}
	return c
}

// realPortLineRE finds the VPN-level service port lines rule #1 exists for.
var realPortLineRE = regexp.MustCompile(`^\s+(no )?service .*listen-port`)

// TestInjectVPNServiceShutdownCoversEveryRealPortLine is the regression for the
// biggest single gap between the app and the GUI: 48 VPN-level
// `service ... listen-port` lines in the real capture, and not one was handled.
//
// They were missed because they are not BLOCKS. A broker-level
// `service semp listen-port 8080` is its own column-0 block, so InjectShutdown can
// put a shutdown beside it; the VPN-level ones are indented children of a single
// `message-vpn "X"` block, so no amount of block-level work reaches them. The broker
// refuses a port change while the transport is up, and `cli -Apes` exits 0 when it
// does, so every one of these was a silent failure waiting to happen.
func TestInjectVPNServiceShutdownCoversEveryRealPortLine(t *testing.T) {
	c := loadRealCapture(t)

	var before int
	for _, b := range c.Blocks {
		for _, line := range b.Lines {
			if realPortLineRE.MatchString(line) {
				before++
			}
		}
	}
	if before == 0 {
		t.Fatal("sanity: the real capture carries no VPN-level listen-port lines")
	}

	got := InjectVPNServiceShutdown(c.Blocks)

	// Every port line must now be immediately preceded by a shutdown for its own
	// service AND transport. Anything less specific would take a sibling transport
	// down with it -- mqtt alone carries four in this capture.
	var checked int
	for _, b := range got {
		for i, line := range b.Lines {
			if !realPortLineRE.MatchString(line) {
				continue
			}
			checked++
			if i == 0 {
				t.Errorf("port line %q has nothing before it", line)
				continue
			}
			prev := strings.TrimSpace(b.Lines[i-1])
			if !strings.HasSuffix(prev, " shutdown") {
				t.Errorf("port line %q is not preceded by a shutdown, got %q", strings.TrimSpace(line), prev)
			}
		}
	}
	if checked != before {
		t.Errorf("checked %d port lines, capture has %d -- the transform lost some", checked, before)
	}
}

// TestInjectVPNServiceShutdownFormPerTransport pins the exact shutdown each shape
// gets. Transport granularity is the whole difficulty: taking a service down as a
// whole to change one transport's port would drop the others, and mqtt in the real
// capture has plain-text, ssl, websocket and websocket-secure all at once.
func TestInjectVPNServiceShutdownFormPerTransport(t *testing.T) {
	cases := []struct{ line, want string }{
		{"  no service rest incoming listen-port", "  service rest incoming plain-text shutdown"},
		{"  no service rest incoming listen-port ssl", "  service rest incoming ssl shutdown"},
		{"  service rest outgoing listen-port 9000", "  service rest outgoing plain-text shutdown"},
		{"  no service mqtt listen-port", "  service mqtt plain-text shutdown"},
		{"  no service mqtt listen-port ssl", "  service mqtt ssl shutdown"},
		{"  no service mqtt listen-port web", "  service mqtt websocket shutdown"},
		{"  no service mqtt listen-port ssl web", "  service mqtt websocket-secure shutdown"},
		{"  service mqtt listen-port 1883", "  service mqtt plain-text shutdown"},
		{`  service amqp listen-port "5672" ssl`, "  service amqp ssl shutdown"},
	}
	for _, tc := range cases {
		got := InjectVPNServiceShutdown([]Block{{Kind: "message-vpn", Lines: []string{
			`message-vpn "acme"`, tc.line, "  exit",
		}}})
		if len(got[0].Lines) != 4 {
			t.Errorf("%q: expected exactly one injected line, got %v", tc.line, got[0].Lines)
			continue
		}
		if got[0].Lines[1] != tc.want {
			t.Errorf("%q injected %q, want %q", tc.line, got[0].Lines[1], tc.want)
		}
	}
}

// TestInjectVPNServiceShutdownLeavesColumnZeroAlone keeps the two injectors apart.
// Broker-level port lines are InjectShutdown's, and they sit at column 0 -- this one
// must not also touch them, or every broker-level port change would get two
// shutdowns from two different rules.
func TestInjectVPNServiceShutdownLeavesColumnZeroAlone(t *testing.T) {
	in := []Block{{Kind: "service", Lines: []string{"service semp listen-port 8080"}}}
	got := InjectVPNServiceShutdown(in)
	if len(got[0].Lines) != 1 {
		t.Errorf("a column-0 port line was injected into: %v", got[0].Lines)
	}
}

// TestInjectVPNServiceShutdownSkipsAnExistingGuard covers the no-duplicate case: a
// capture that already shuts the transport down immediately before its port change
// needs nothing added.
func TestInjectVPNServiceShutdownSkipsAnExistingGuard(t *testing.T) {
	got := InjectVPNServiceShutdown([]Block{{Lines: []string{
		`message-vpn "acme"`,
		"  service mqtt plain-text shutdown",
		"  no service mqtt listen-port",
	}}})
	if len(got[0].Lines) != 3 {
		t.Errorf("a shutdown was injected despite one already guarding the port line: %v", got[0].Lines)
	}
}

// TestShutdownDefaultClientUsername covers the `default` client-username, which is
// the one object in a VPN that is edited rather than recreated: it cannot be
// deleted, so the teardown leaves it standing and the broker refuses several edits
// while it is enabled. The capture carries no shutdown of its own because on the
// SOURCE broker it was already in its final state.
func TestShutdownDefaultClientUsername(t *testing.T) {
	in := []Block{
		{Kind: "client-username", Name: "default", VPN: "acme", Lines: []string{
			`client-username "default" message-vpn "acme"`, "  no shutdown", "  exit"}},
		{Kind: "client-username", Name: "app1", VPN: "acme", Lines: []string{
			`client-username "app1" message-vpn "acme"`, "  no shutdown", "  exit"}},
	}
	got := ShutdownDefaultClientUsername(in)

	if got[0].Lines[1] != "  shutdown" {
		t.Errorf("default username lines = %v, want a shutdown right after the opener", got[0].Lines)
	}
	// The capture's own trailing state still follows, so this cannot leave it down.
	if got[0].Lines[2] != "  no shutdown" {
		t.Errorf("the artifact's own enablement line must survive, got %v", got[0].Lines)
	}
	if len(got[1].Lines) != 3 {
		t.Errorf("a non-default username must be untouched, got %v", got[1].Lines)
	}
}

// TestShutdownBridges covers the rule that deliberately changes the state the
// artifact asked for. A bridge comes up the moment it is un-shut and starts moving
// messages between two brokers -- so an import that brought one up would begin
// replicating traffic before anyone had looked at the result.
func TestShutdownBridges(t *testing.T) {
	got := ShutdownBridges([]Block{{Kind: "bridge", Name: "b1", Lines: []string{
		`create bridge "b1" message-vpn "acme"`,
		"  remote",
		"      no shutdown",
		"    exit",
		"  no shutdown",
		"  exit",
	}}})
	for _, line := range got[0].Lines {
		if strings.TrimSpace(line) == "no shutdown" {
			t.Errorf("a bridge was left able to come up: %v", got[0].Lines)
		}
	}
	if got[0].Lines[2] != "      shutdown" || got[0].Lines[4] != "  shutdown" {
		t.Errorf("indentation was not preserved: %v", got[0].Lines)
	}
}

// TestRewriteAdminUsername covers the admin CLI user, which already exists on every
// broker -- `create` is refused as a duplicate and every attribute under it is lost
// with the block.
func TestRewriteAdminUsername(t *testing.T) {
	got := RewriteAdminUsername([]Block{{Kind: "username", Name: "admin", Create: true, Lines: []string{
		`create username "admin" password xyz`, "  global-access-level admin", "  exit"}}})
	if got[0].Lines[0] != "username admin" {
		t.Errorf("opener = %q, want `username admin`", got[0].Lines[0])
	}
	if got[0].Create {
		t.Error("Create must be cleared with the rewrite")
	}
}

// TestRewriteAdminUsernameFindsTheOpenerPastAPragma is the trap Block.Opener() was
// added for: a held `! pragma:` line is prepended to Lines, so the opener is not
// always Lines[0]. Rewriting index 0 blindly would corrupt the pragma and leave the
// `create` in place.
func TestRewriteAdminUsernameFindsTheOpenerPastAPragma(t *testing.T) {
	got := RewriteAdminUsername([]Block{{Kind: "username", Name: "admin", Create: true, Lines: []string{
		"! pragma:interpreter:ignore-already-exists",
		`create username "admin" password xyz`,
		"  exit",
	}}})
	if got[0].Lines[0] != "! pragma:interpreter:ignore-already-exists" {
		t.Errorf("the held pragma was overwritten: %v", got[0].Lines)
	}
	if got[0].Lines[1] != "username admin" {
		t.Errorf("the opener past the pragma was not rewritten: %v", got[0].Lines)
	}
}

// TestReorderVPNProfilesPutsProfilesLast covers the dependency order the capture
// does not have: a client-username names the client-profile and acl-profile it
// uses, and the broker refuses a username referencing a profile it does not yet
// have.
func TestReorderVPNProfilesPutsProfilesLast(t *testing.T) {
	blk := func(kind, name, vpn string) Block {
		return Block{Region: RegionVPN, Kind: kind, Name: name, VPN: vpn, Lines: []string{kind + " " + name}}
	}
	in := []Block{
		blk("client-username", "app1", "acme"),
		blk("queue", "orders", "acme"),
		blk("client-profile", "p1", "acme"),
		blk("message-spool", "", "acme"),
		blk("acl-profile", "a1", "acme"),
		blk("client-username", "app2", "beta"),
		blk("queue", "events", "beta"),
	}
	got := ReorderVPNProfiles(in)

	var order []string
	for _, b := range got {
		order = append(order, b.Kind+"/"+b.VPN)
	}
	want := []string{
		"queue/acme", "message-spool/acme",
		"client-username/acme", "client-profile/acme", "acl-profile/acme",
		"queue/beta", "client-username/beta",
	}
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Errorf("order = %v,\nwant       %v -- profiles last, and never carried across a VPN boundary", order, want)
	}
}

// TestCollapseDuplicateLines covers the GUI's own dedupe. A repeated configuration
// line is at best noise and at worst a second `create` of something that now exists;
// blank lines are exempt because collapsing them only reflows the artifact.
func TestCollapseDuplicateLines(t *testing.T) {
	got := CollapseDuplicateLines([]Block{{Lines: []string{
		"a", "a", "b", "", "", "c", "b",
	}}})
	want := []string{"a", "b", "", "", "c", "b"}
	if strings.Join(got[0].Lines, "|") != strings.Join(want, "|") {
		t.Errorf("lines = %v, want %v", got[0].Lines, want)
	}
}

// TestClearExistingSyslogsOnlyForEntriesTheTargetHas covers both halves of a rule
// that pulls in opposite directions: `create syslog` for an entry that exists is
// rejected, but `no syslog` for one that does NOT exist is itself an error. Emitting
// the removal unconditionally would fail every import onto a broker with no syslog
// configured, which is most of them.
func TestClearExistingSyslogsOnlyForEntriesTheTargetHas(t *testing.T) {
	in := []Block{
		{Kind: kindSyslog, Name: "central", Create: true, Section: "Create Syslog",
			Lines: []string{`create syslog "central"`, "  exit"}},
		{Kind: kindSyslog, Name: "brand-new", Create: true, Section: "Create Syslog",
			Lines: []string{`create syslog "brand-new"`, "  exit"}},
	}
	got := ClearExistingSyslogs(in, []string{"central"})

	if len(got) != 3 {
		t.Fatalf("got %d blocks, want 3 (one removal, for the entry the target has): %+v", len(got), got)
	}
	if got[0].Opener() != `no syslog "central"` {
		t.Errorf("first block = %q, want the removal for the EXISTING entry", got[0].Opener())
	}
	// The operand is mirrored from the artifact, not re-quoted: a capture that
	// writes the name bare must get a bare removal, because no capture in hand
	// settles which form the broker emits and a wrong guess is a line that errors.
	bare := ClearExistingSyslogs([]Block{{
		Kind: kindSyslog, Name: "central", Create: true,
		Lines: []string{"create syslog central", "  exit"},
	}}, []string{"central"})
	if len(bare) != 2 || bare[0].Opener() != "no syslog central" {
		t.Errorf("bare-operand removal = %q, want `no syslog central` (the artifact's own rendering)",
			bare[0].Opener())
	}
	if got[1].Name != "central" || got[2].Name != "brand-new" {
		t.Errorf("the removal must come immediately before its own create, got %q then %q",
			got[1].Name, got[2].Name)
	}
	for _, b := range got {
		if b.Opener() == `no syslog "brand-new"` {
			t.Error("a removal was emitted for an entry the target does NOT have; that is itself an error on the broker")
		}
	}
}

// TestClearExistingSyslogsAgainstTheRealCapture is the end-to-end shape check, now
// that a real capture carries a syslog destination.
//
// It pins the three facts the rule depends on, all of which were guesses until the
// capture existed: the section header carries the name (so the rule is matched by
// PREFIX, not equality), the section is classified rather than falling through as
// unclassified (an unclassified one is never applied at all), and the operand is
// quoted -- which is what the emitted removal mirrors.
func TestClearExistingSyslogsAgainstTheRealCapture(t *testing.T) {
	c := loadRealCapture(t)

	names := c.SyslogNames()
	if len(names) == 0 {
		t.Skip("this capture configures no syslog destination")
	}

	var syslogSection string
	for _, b := range c.Blocks {
		if b.Kind == kindSyslog && b.Create {
			syslogSection = b.Section
			break
		}
	}
	if _, ok := RuleFor(syslogSection); !ok {
		t.Fatalf("section %q is UNCLASSIFIED, so the syslog entry would never be applied -- "+
			"the removal would be emitted and the create would not", syslogSection)
	}
	if rule, _ := RuleFor(syslogSection); rule.Disposition != Apply {
		t.Errorf("section %q is %q, want apply", syslogSection, rule.Disposition)
	}

	removalsFor := func(targetHas []string) []string {
		var out []string
		for _, b := range ClearExistingSyslogs(c.Blocks, targetHas) {
			if strings.HasPrefix(b.Opener(), "no syslog") {
				out = append(out, b.Opener())
			}
		}
		return out
	}

	// One removal per entry the target ACTUALLY has, and none for the others. The
	// capture carries several destinations, so this is also what proves the rule is
	// per-entry rather than all-or-nothing.
	var wantAll []string
	for _, n := range names {
		wantAll = append(wantAll, `no syslog "`+n+`"`)
	}
	if got := removalsFor(names); strings.Join(got, ",") != strings.Join(wantAll, ",") {
		t.Errorf("target has every entry: removals = %v, want %v", got, wantAll)
	}
	if got := removalsFor(names[:1]); len(got) != 1 || got[0] != wantAll[0] {
		t.Errorf("target has only %q: removals = %v, want exactly [%q]", names[0], got, wantAll[0])
	}

	// Target has none -> nothing added, because `no syslog` for a missing entry is
	// itself an error on the broker. This is the common case and the reason the rule
	// consults the target at all.
	if n := len(ClearExistingSyslogs(c.Blocks, nil)); n != len(c.Blocks) {
		t.Errorf("a target with no syslog entries gained %d block(s)", n-len(c.Blocks))
	}

	// A destination the TARGET has but the artifact does not is left alone: broker
	// scope is additive, and dropping a destination the target's own operator
	// configured would be a log-delivery outage nobody asked for.
	if got := removalsFor([]string{"a-destination-only-the-target-has"}); len(got) != 0 {
		t.Errorf("removals = %v, want none -- import must not touch a target-only destination", got)
	}

	// Idempotent: the removal is a BLOCK of its own placed before the create, so
	// without a guard a second pass prepends another one to the same create, and a
	// third the pass after. Re-running a failed import is the documented recovery,
	// so the second attempt has to apply the same script as the first. This escaped
	// notice until the whole-pipeline idempotency test was pointed at a POPULATED
	// target -- with an empty one, this transform does nothing at all.
	once := ClearExistingSyslogs(c.Blocks, names)
	twice := ClearExistingSyslogs(once, names)
	if len(twice) != len(once) {
		t.Errorf("ClearExistingSyslogs is not idempotent: %d blocks after one pass, %d after two",
			len(once), len(twice))
	}
}

// TestClearExistingSyslogsNoTargetEntriesIsANoOp covers the common case -- a target
// with no syslog configured at all -- which must add nothing.
func TestClearExistingSyslogsNoTargetEntriesIsANoOp(t *testing.T) {
	in := []Block{{Kind: kindSyslog, Name: "central", Create: true, Lines: []string{`create syslog "central"`}}}
	if got := ClearExistingSyslogs(in, nil); len(got) != 1 {
		t.Errorf("got %d blocks, want 1 unchanged", len(got))
	}
}

// TestSyslogNamesReadsThroughTheQuoteAwareParser pins where the target's syslog
// names come from: the parsed blocks, so a name containing spaces survives -- the
// same failure mode that filed an existing message-VPN as new.
func TestSyslogNamesReadsThroughTheQuoteAwareParser(t *testing.T) {
	c := &Capture{Blocks: []Block{
		{Kind: kindSyslog, Name: "central log", Create: true},
		{Kind: kindSyslog, Name: "central log", Create: true},
		{Kind: "queue", Name: "orders"},
	}}
	got := c.SyslogNames()
	if len(got) != 1 || got[0] != "central log" {
		t.Errorf("SyslogNames() = %v, want [\"central log\"] once", got)
	}
}

// tpsCapture parses semp/all-cfg-tps-org.cli -- a SECOND real capture, from an older
// broker (soltr/10_4_1VMR) than semp/all.cli. It carries the nested-object sections
// all.cli does not, which is how the gap they represent was found at all.
func tpsCapture(t *testing.T) *Capture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "semp", "all-cfg-tps-org.cli"))
	if err != nil {
		t.Skipf("second real capture not available (%v)", err)
	}
	c, err := ParseBlocks(raw)
	if err != nil {
		t.Fatalf("parse the TPS capture: %v", err)
	}
	return c
}

// TestNestedObjectsAreFoundDespiteTheParentKind is the discovery this whole family
// of rules rests on.
//
// A client CA is NOT a column-0 block: the section opens `authentication` and
// creates inside it, so Block.Kind is the PARENT and every rule keyed on Kind misses
// it -- the same shape that hid the VPN-level service lines. Domain CAs (`ssl`) and
// virtual hostnames (`service`) are identical in form.
func TestNestedObjectsAreFoundDespiteTheParentKind(t *testing.T) {
	c := tpsCapture(t)

	cas := c.nestedNames(verbClientCA)
	if len(cas) == 0 {
		t.Fatal("no client CAs found; nestedNames is not seeing objects created inside a parent block")
	}
	hosts := c.nestedNames(verbVirtualHostname)
	if len(hosts) == 0 {
		t.Fatal("no virtual hostnames found in a capture that has them")
	}

	// The parent is what Kind reports, which is exactly why Kind cannot be the key.
	for _, b := range c.Blocks {
		if strings.HasPrefix(b.Section, "Create Client Certificate Authority") {
			if b.Kind == verbClientCA {
				t.Errorf("Kind = %q; if this ever becomes true the nested rules can be simplified", b.Kind)
			}
			break
		}
	}

	// And both sections must be CLASSIFIED, or they are skipped and the import exits
	// non-zero without applying them.
	for _, sec := range []string{"Create Client Certificate Authority: \"G2-root\"", "Configure Virtual Hostnames"} {
		if _, ok := RuleFor(sec); !ok {
			t.Errorf("section %q is unclassified, so it would never be applied", sec)
		}
	}
}

// TestClearExistingNestedCoversAllThreeVerbs is the regression for three sections
// that shipped applied-but-UNGUARDED: a `create` for a name the target already has
// is rejected, no capture emits a removal (the broker's own `show current-config`
// writes only the creates), and two of the three carried a Reason claiming a
// delete-then-create that existed nowhere in the code.
//
// The removal form is VERIFIED against Solace's CLI reference
// (semp/appliance_cli_reference.html), which documents each as one command with both
// verbs -- `[create | no] ldap-profile <profile-name>` and the same shape for the
// two CAs -- so `no <verb> <name>` is documented, not inferred.
//
// It also covers the placement rule the LDAP group forced: a removal goes
// immediately BEFORE its own create at that create's own indent, which is the only
// placement correct at every depth.
func TestClearExistingNestedCoversAllThreeVerbs(t *testing.T) {
	c := tpsCapture(t)

	target := TargetState{
		ClientCAs:    c.nestedNames(verbClientCA),
		DomainCAs:    c.nestedNames(verbDomainCA),
		LDAPProfiles: c.nestedNames(verbLDAPProfile),
	}
	if len(target.ClientCAs) == 0 || len(target.LDAPProfiles) == 0 {
		t.Skip("this capture carries neither client CAs nor LDAP profiles")
	}

	got := ClearExistingNested(c.Blocks, target)

	// Every create of a name the target has is now immediately preceded by its own
	// removal, at the same indentation.
	for _, verb := range []string{verbClientCA, verbDomainCA, verbLDAPProfile} {
		re := nestedCreateRE(verb)
		var creates, guarded int
		for _, b := range got {
			for i, line := range b.Lines {
				m := re.FindStringSubmatch(line)
				if m == nil {
					continue
				}
				creates++
				want := m[1] + "no " + verb + " " + m[2]
				if i > 0 && b.Lines[i-1] == want {
					guarded++
				}
			}
		}
		if creates > 0 && guarded != creates {
			t.Errorf("%s: %d create(s), %d guarded by a removal -- an unguarded create for a "+
				"name the target already has is rejected", verb, creates, guarded)
		}
	}

	// Nothing is cleared for a target that has none of them: a removal for an absent
	// object is itself an error.
	for _, b := range ClearExistingNested(c.Blocks, TargetState{}) {
		for _, line := range b.Lines {
			if strings.HasPrefix(strings.TrimSpace(line), "no ") &&
				strings.Contains(line, "-certificate-authority") {
				t.Errorf("removal emitted for an empty target: %q", line)
			}
		}
	}
}

// TestClearExistingClientCAsIntersects pins the rule that differs from the virtual
// hostname one despite the identical shape: a client CA is removed only when the
// target HAS it and this artifact re-creates it.
//
// Without the target half the removal errors on an absent CA; without the artifact
// half the import deletes trust material the TARGET's operator added and this
// artifact never mentions -- and a client CA decides which clients the broker
// accepts, so removing one nobody asked about locks those clients out.
func TestClearExistingClientCAsIntersects(t *testing.T) {
	c := tpsCapture(t)
	artifact := c.nestedNames(verbClientCA)

	removals := func(blocks []Block) []string {
		var out []string
		for _, b := range blocks {
			for _, l := range b.Lines {
				if strings.HasPrefix(strings.TrimSpace(l), "no "+verbClientCA) {
					out = append(out, strings.TrimSpace(l))
				}
			}
		}
		return out
	}

	// Target has the first CA and one of its own that the artifact does not carry.
	got := removals(ClearExistingClientCAs(c.Blocks, []string{artifact[0], "a-ca-only-the-target-has"}))
	want := `no ` + verbClientCA + ` "` + artifact[0] + `"`
	if len(got) != 1 || got[0] != want {
		t.Errorf("removals = %v, want exactly [%q] -- the target's own CA must be left alone", got, want)
	}

	// Target has none: nothing to clear, because the removal would error.
	if r := removals(ClearExistingClientCAs(c.Blocks, nil)); len(r) != 0 {
		t.Errorf("removals = %v, want none when the target carries no client CA", r)
	}
}

// TestClearTargetVirtualHostnamesCoversBothHalves pins the OPPOSITE rule to the
// client CAs, and the division of labour that delivers it.
//
// Why it differs: a virtual hostname is a DNS name resolving to this broker, so one
// left behind from the target's previous life keeps answering for a name the
// imported configuration never claimed. So EVERY name the target has is cleared,
// artifact or not -- but by two mechanisms, because a removal has to be issued in
// the scope its object lives in. A name the artifact re-creates gets its removal
// immediately before that create (the only placement correct at every nesting
// depth); a name with no create has nothing to sit before, so TargetOnlyRemovals
// renders it opening the `service` scope itself.
//
// The test covers both together, because neither half alone is the property.
func TestClearTargetVirtualHostnamesCoversBothHalves(t *testing.T) {
	c := tpsCapture(t)
	inArtifact := c.nestedNames(verbVirtualHostname)
	if len(inArtifact) == 0 {
		t.Skip("this capture creates no virtual hostnames")
	}
	const orphan = "stale.example.test"
	target := append(append([]string{}, inArtifact...), orphan)

	// Half one: a name the artifact re-creates, cleared beside its own create.
	var beside bool
	for _, b := range ClearTargetVirtualHostnames(c.Blocks, target) {
		for i, l := range b.Lines {
			if strings.TrimSpace(l) == `no `+verbVirtualHostname+` "`+inArtifact[0]+`"` {
				if i+1 < len(b.Lines) && strings.Contains(b.Lines[i+1], `create `+verbVirtualHostname) {
					beside = true
				}
			}
		}
	}
	if !beside {
		t.Errorf("virtual hostname %q is re-created by the artifact but got no removal "+
			"immediately before its create", inArtifact[0])
	}

	// Half two: a name only the TARGET has, which has no create to sit before.
	orphans := strings.Join(TargetOnlyRemovals(c.Blocks, TargetState{VirtualHostnames: target}), "\n")
	if !strings.Contains(orphans, `no `+verbVirtualHostname+` "`+orphan+`"`) {
		t.Errorf("a virtual hostname the target has and the artifact does NOT was not cleared:\n%s", orphans)
	}
	// ...and the in-artifact ones must NOT be repeated there, or each is removed twice.
	if strings.Contains(orphans, inArtifact[0]) {
		t.Errorf("%q is cleared beside its create AND again as an orphan:\n%s", inArtifact[0], orphans)
	}
}

// TestPrepareForApplyIsIdempotentOnTheRealCapture is the safety net over the whole
// pipeline: running it twice must not keep adding lines. Several of these rules
// INSERT, and an insert that does not recognise its own previous output grows the
// script every time -- which on a re-run after a failed import would apply a
// different configuration than the first attempt did.
func TestPrepareForApplyIsIdempotentOnTheRealCapture(t *testing.T) {
	c := loadRealCapture(t)

	// A populated target, not an empty one: the clears are exactly the transforms
	// that INSERT, so running with them switched off would not exercise the
	// idempotency this test exists to protect.
	//
	// A populated target is NOT enough, and believing it was is what this setup got
	// wrong for a while. Every clear attaches its removal to a CREATE in the
	// ARTIFACT -- ClearExistingSyslogs keys on a `create syslog` block, and
	// clearNestedObjects on a `create <verb>` line inside one -- so an artifact with
	// no such create is a no-op however many names the target reports. Both halves
	// have to come from a capture that actually carries the object.
	//
	// semp/all.cli is re-taken from whichever live broker is available, and the
	// current one configures no syslog, no client CA and no virtual hostname at all,
	// so it can supply neither half for any of the three. semp/all-cfg-tps-org.cli
	// carries all three, so both halves are read from it -- still real capture data,
	// not a hand-made block.
	target := readTargetState(c)
	tps := tpsCapture(t)
	target.Syslogs = tps.SyslogNames()
	target.ClientCAs = tps.nestedNames(verbClientCA)
	target.VirtualHostnames = tps.nestedNames(verbVirtualHostname)
	if len(target.Syslogs) == 0 || len(target.ClientCAs) == 0 || len(target.VirtualHostnames) == 0 {
		t.Fatalf("target state is incomplete (syslogs=%d clientCAs=%d vhosts=%d); this test cannot "+
			"reach the clears with an empty one", len(target.Syslogs), len(target.ClientCAs), len(target.VirtualHostnames))
	}

	cleared := []string{kindSyslog, verbClientCA, verbVirtualHostname}
	creates := func(b Block, verb string) bool {
		for _, l := range b.Lines {
			if strings.HasPrefix(strings.TrimSpace(l), "create "+verb+" ") {
				return true
			}
		}
		return false
	}
	blocks := append([]Block{}, c.Blocks...)
	for _, b := range tps.Blocks {
		for _, verb := range cleared {
			if creates(b, verb) {
				blocks = append(blocks, b)
				break
			}
		}
	}

	once := PrepareForApply(blocks, target)
	twice := PrepareForApply(once, target)

	countLines := func(bs []Block) int {
		n := 0
		for _, b := range bs {
			n += len(b.Lines)
		}
		return n
	}
	countRemovals := func(bs []Block, verb string) int {
		n := 0
		for _, b := range bs {
			for _, l := range b.Lines {
				if strings.HasPrefix(strings.TrimSpace(l), "no "+verb+" ") {
					n++
				}
			}
		}
		return n
	}

	// Each clear must actually have INSERTED something. Without this the setup above
	// can go quietly dead again -- a fixture re-taken from a broker that configures
	// one of these objects differently would leave that clear on its no-op path, and
	// running a no-op twice is identical in a way that proves nothing. A total line
	// count cannot stand in for this: the inserting transforms that do NOT consult
	// the target (InjectVPNServiceShutdown alone adds one line per VPN-level port
	// line, and there are dozens) make the total grow whether any clear fired or not.
	for _, verb := range cleared {
		if n := countRemovals(once, verb); n == 0 {
			t.Fatalf("no `no %s` removal was inserted, so that clear took its no-op path and "+
				"running the pipeline twice says nothing about it", verb)
		}
	}
	// Idempotency, per clear as well as in total: one clear growing while another
	// shrank would net out to the same line count and slip past the total alone.
	for _, verb := range cleared {
		if a, b := countRemovals(once, verb), countRemovals(twice, verb); a != b {
			t.Errorf("the %s clear is not idempotent: %d removals after one pass, %d after two", verb, a, b)
		}
	}
	if a, b := countLines(once), countLines(twice); a != b {
		t.Errorf("PrepareForApply is not idempotent: %d lines after one pass, %d after two", a, b)
	}
}

// TestShutdownBridgesLinesAreExcusedByTheDiff is the regression for the worst bug the
// chunked apply shipped with: EVERY artifact containing a bridge failed verification
// after a perfectly clean apply.
//
// ShutdownBridges is the one transform that CHANGES a line rather than adding one --
// it applies `shutdown` where the artifact says `no shutdown`, deliberately, so an
// import never brings a bridge up. But ImportVerify diffs the UNTRANSFORMED artifact
// against the target, directionally, so it demanded the artifact's `no shutdown` be
// present on a target the import had just made say `shutdown`. It never could be.
//
// The line therefore belongs with the ignored ones, for the same reason a skipped
// section's lines do: it was never intended to land.
func TestShutdownBridgesLinesAreExcusedByTheDiff(t *testing.T) {
	bridge := Block{Section: "Create Bridge", Region: RegionVPN, Kind: kindBridge,
		Name: "b1", VPN: "acme", Lines: []string{
			`create bridge "b1" message-vpn "acme"`,
			"  remote",
			"      no shutdown",
			"    exit",
			"  no shutdown",
			"  exit",
		}}
	// What the target re-exports after the import: the bridge left shut down.
	applied := ShutdownBridges([]Block{bridge})

	p := &ImportPlan{Capture: &Capture{Blocks: []Block{bridge}}}
	d := DiffBlocks([]Block{bridge}, applied, importIgnore(p))
	if !d.OK() {
		t.Errorf("a clean apply of an artifact with a bridge reported a verification FAILURE: %+v -- "+
			"the artifact's `no shutdown` can never be on the target, because the import applied "+
			"`shutdown` on purpose", d)
	}

	// The exemption is narrow: it must not excuse a `no shutdown` anywhere else, or a
	// service the import genuinely failed to enable would read as clean.
	notABridge := Block{Kind: "client-username", Lines: []string{`client-username "u"`, "  no shutdown"}}
	if BridgeEnablementInverted(notABridge, "  no shutdown") {
		t.Error("the bridge exemption leaked to a non-bridge block")
	}
	if BridgeEnablementInverted(bridge, "  some-other-line") {
		t.Error("the bridge exemption matched a line that is not an enablement line")
	}
}

// TestTargetOnlyRemovalsWhenTheArtifactCreatesNone covers the case that made
// "clear the target's virtual hostnames" silently do nothing: clearNestedObjects can
// only place a removal inside a block that CREATES that verb, and an artifact with no
// virtual hostnames has no such block. That is precisely when the rule has to remove
// something -- the target carries names the artifact does not, and it is supposed to
// end up matching.
func TestTargetOnlyRemovalsWhenTheArtifactCreatesNone(t *testing.T) {
	noVHosts := []Block{{Kind: "console", Lines: []string{"create console"}}}
	got := TargetOnlyRemovals(noVHosts, TargetState{VirtualHostnames: []string{"stale.example.test"}})
	if len(got) == 0 {
		t.Fatal("no removal rendered; a target-only virtual hostname would survive the import")
	}
	joined := strings.Join(got, " | ")
	if !strings.Contains(joined, `no virtual-hostname "stale.example.test"`) {
		t.Errorf("removals = %q, want the target-only name", joined)
	}
	// It opens its own scope, since there is no create block to borrow one from.
	if got[0] != "service" || got[len(got)-1] != "  exit" {
		t.Errorf("removals = %q, want them wrapped in the `service` scope", joined)
	}

	// A name the artifact DOES create is left to clearNestedObjects, which places it
	// beside that create -- emitting it here too would remove it twice.
	withVHost := []Block{{Kind: "service", Lines: []string{
		"service", `  create virtual-hostname "kept.example.test"`, "    exit", "  exit"}}}
	if r := TargetOnlyRemovals(withVHost, TargetState{VirtualHostnames: []string{"kept.example.test"}}); len(r) != 0 {
		t.Errorf("removals = %q, want none for a name the artifact re-creates", r)
	}
	if r := TargetOnlyRemovals(noVHosts, TargetState{}); len(r) != 0 {
		t.Errorf("removals = %q, want none when the target has no virtual hostnames", r)
	}
}

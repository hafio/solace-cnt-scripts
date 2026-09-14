package broker

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// blocks_test.go exercises ParseBlocks and its helpers against
// testdata/currentconfig_sample.cli, a small hand-written fixture that
// reproduces the STRUCTURE of a real `show current-config all` capture
// (semp/all.cli, broker 10.26.0.8827, 33728 lines) with fake names and no
// credentials. semp/all.cli itself is never read by a test or copied into
// testdata: it carries real admin/LDAP/redundancy secrets and must not be
// committed or quoted, so every case below is traced against fixed line
// numbers in that file (cited per test) and reproduced by hand instead.
//
// Where a test's name says BUG, the assertion documents the CORRECT behaviour
// per blocks.go's own doc comments and is expected to fail against the current
// implementation -- this file owns only the test suite, not blocks.go, so the
// bug is reported rather than fixed.

// loadSample parses the fixture and fails the test on any parse error, so
// every other test here can assume a clean Capture.
func loadSample(t *testing.T) *Capture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "currentconfig_sample.cli"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	c, err := ParseBlocks(raw)
	if err != nil {
		t.Fatalf("ParseBlocks(fixture): %v", err)
	}
	return c
}

// blockByIndex returns the block with the given 1-based Index (the position
// ParseBlocks assigned it), or fails the test -- used instead of indexing
// c.Blocks directly so a test also cross-checks Index itself.
func blockByIndex(t *testing.T, blocks []Block, idx int) Block {
	t.Helper()
	for _, b := range blocks {
		if b.Index == idx {
			return b
		}
	}
	t.Fatalf("no block with Index %d among %d blocks", idx, len(blocks))
	return Block{}
}

// --- header / preamble -----------------------------------------------------

// TestParseBlocksHeaderAndPreamble covers the header fields ParseBlocks reads
// and the trap named in blocks.go's own comment: a real capture's first line
// is " ! BEGIN" WITH A LEADING SPACE (semp/all.cli line 1), and that must not
// stop the header from parsing.
func TestParseBlocksHeaderAndPreamble(t *testing.T) {
	c := loadSample(t)

	if c.Version != "99.99.9.9999" {
		t.Errorf("Version = %q, want %q", c.Version, "99.99.9.9999")
	}
	if c.SEMPSchema != "broker/99.99.9.9999" {
		t.Errorf("SEMPSchema = %q, want %q", c.SEMPSchema, "broker/99.99.9.9999")
	}
	if c.Type != BrokerSoftware {
		t.Errorf("Type = %q, want %q (from the \"broker/\" schema prefix)", c.Type, BrokerSoftware)
	}
	if c.Redacted {
		t.Error(`Redacted should be false for Redacted: "False"`)
	}

	if len(c.Preamble) < 17 {
		t.Fatalf("Preamble has %d lines, want at least 17", len(c.Preamble))
	}
	if c.Preamble[0] != " ! BEGIN" {
		t.Errorf("Preamble[0] = %q, want %q (leading space preserved)", c.Preamble[0], " ! BEGIN")
	}
	// The fixture's 17th line is the bare "configure" that ends the header
	// (mode preamble: home/enable/configure), so it must sit at index 16.
	if c.Preamble[16] != "configure" {
		t.Errorf("Preamble[16] = %q, want %q", c.Preamble[16], "configure")
	}
}

// TestParseBlocksPreambleHoldsNoOrphanPragma pins Capture.Preamble to exactly
// what its doc comment promises: everything up to and including `configure`,
// verbatim, and nothing more.
//
// The regression it guards against is specific. A pragma can sit between a
// "! Section:" comment and its opener -- the fixture's
// "! Configure SolOsPhy Interfaces:" / "! pragma:interpreter:ignore-already-exists"
// / "create interface ..." run, modelled on semp/all.cli lines 43-45 and 282-284.
// The section comment has already closed the previous block, so that pragma
// arrives with none open, and an earlier parser appended it to Preamble.
//
// That was not cosmetic. Preamble is re-emitted as a fixed prefix ahead of every
// block, so a pragma captured that way is relocated ahead of ALL of them rather
// than staying immediately before the one `create` line it guards -- which
// detaches `ignore-already-exists` from the line whose replayability depends on
// it. The parser now HOLDS such a pragma and attaches it to the block that
// follows; TestParseBlocksPragmaBetweenOpenerAndChildren asserts the other half,
// that it lands in that block's Lines in source order.
func TestParseBlocksPreambleHoldsNoOrphanPragma(t *testing.T) {
	c := loadSample(t)
	if len(c.Preamble) != 17 {
		t.Errorf("len(Preamble) = %d, want 17 (everything up to and including `configure`, "+
			"verbatim); extra lines = %v", len(c.Preamble), c.Preamble[17:])
	}
}

// --- BrokerType --------------------------------------------------------------

// TestBrokerTypeFromSchema covers every arm of brokerTypeFromSchema, since it
// gates cross-type import refusal: an appliance capture carries interface,
// VRF, DNS, clock and SNMP sections a software capture has none of, so getting
// this wrong lets a cross-type import through silently.
func TestBrokerTypeFromSchema(t *testing.T) {
	cases := []struct {
		schema string
		want   BrokerType
	}{
		{"broker/10.26.0.8827", BrokerSoftware},
		{"app/10.25.4.1", BrokerAppliance},
		{"soltr/10_11_1VMR", BrokerSoftware},
		{"soltr/10_11_1", BrokerAppliance},
		{"garbage", BrokerUnknown},       // no "/" at all
		{"unknown/1.2.3", BrokerUnknown}, // has "/" but an unrecognised prefix
		{"", BrokerUnknown},
	}
	for _, c := range cases {
		if got := brokerTypeFromSchema(c.schema); got != c.want {
			t.Errorf("brokerTypeFromSchema(%q) = %q, want %q", c.schema, got, c.want)
		}
	}
}

// --- opener parsing (newBlock) -----------------------------------------------

// TestNewBlockOpener is table-driven over the opener grammar
// `[create] <kind> ["<name>"] [message-vpn "<vpn>"] ...`. The cases that carry
// a real bug (a name-less message-vpn-qualified opener, and any quoted operand
// containing spaces) are pulled into their own tests below rather than folded
// in here, so a red assertion cannot hide a regression in the cases that do
// pass today.
func TestNewBlockOpener(t *testing.T) {
	cases := []struct {
		name             string
		opener           string
		kind, vname, vpn string
		create           bool
	}{
		{"message-vpn", `message-vpn "acme"`, "message-vpn", "acme", "acme", false},
		{"create message-vpn", `create message-vpn "acme"`, "message-vpn", "acme", "acme", true},
		{"vpn-scoped object", `client-username "app1" message-vpn "acme"`, "client-username", "app1", "acme", false},
		{"bare one-liner", `ssl crime-exploit-protection`, "ssl", "", "", false},
		{"double-spaced fields", `system scaling  scale  max-connections 1000`, "system", "", "", false},
	}
	for _, c := range cases {
		b := newBlock("section", 1, c.opener)
		if b.Kind != c.kind || b.Name != c.vname || b.VPN != c.vpn || b.Create != c.create {
			t.Errorf("%s: newBlock(%q) = {Kind:%q Name:%q VPN:%q Create:%v}, want {%q %q %q %v}",
				c.name, c.opener, b.Kind, b.Name, b.VPN, b.Create, c.kind, c.vname, c.vpn, c.create)
		}
	}
}

// TestNewBlockMessageSpoolQualifierHasNoOwnName covers an opener that carries
// ONLY a message-vpn qualifier and no separate name operand -- the real shape
// at semp/all.cli line 32743 (also 571, 592, 905, 926):
// `message-spool message-vpn "default"`. message-spool is a per-VPN singleton,
// not a named object, so Name should be "".
//
// BUG: firstQuoted(fields[1:]) cannot distinguish a `message-vpn "X"` scope
// qualifier from the block's own name operand, so when there is no OTHER
// quoted field it returns the qualifier's value as the name too. newBlock's
// Name ends up equal to VPN ("default") instead of "". Any consumer that keys
// off an empty Name to mean "this object has no identity of its own" (every
// bare `<kind> message-vpn "X"` opener, not just message-spool) gets an object
// that appears to be named after its own VPN.
func TestNewBlockMessageSpoolQualifierHasNoOwnName(t *testing.T) {
	const opener = `message-spool message-vpn "default"`
	b := newBlock("", 1, opener)
	if b.Kind != "message-spool" {
		t.Fatalf("newBlock(%q).Kind = %q, want %q", opener, b.Kind, "message-spool")
	}
	if b.VPN != "default" {
		t.Errorf("newBlock(%q).VPN = %q, want %q", opener, b.VPN, "default")
	}
	if b.Name != "" {
		t.Errorf("newBlock(%q).Name = %q, want \"\" -- message-spool has no name of its own, "+
			"only the message-vpn qualifier", opener, b.Name)
	}
}

// TestNewBlockVPNNameWithSpaces covers a message-vpn block whose own name
// contains spaces -- the real VPN at semp/all.cli lines 20 and 379,
// `create message-vpn "A VPN WITH LONG NAME AND SPACES"`. Name and VPN must
// round-trip the full quoted operand.
//
// BUG: newBlock splits the opener with strings.Fields before looking for a
// quoted operand, so a multi-word quoted name is fragmented into several
// fields and firstQuoted/vpnQualifier each return only its first word ("A").
// This is not hypothetical: semp/all.cli's own broker-level capture creates
// exactly this VPN, so parsing a real capture containing it would silently
// rename the VPN to "A" and, per TestNewBlockVPNQualifierWithSpaces below,
// attribute none of its VPN-scoped objects to it either.
func TestNewBlockVPNNameWithSpaces(t *testing.T) {
	const opener = `create message-vpn "A VPN WITH LONG NAME AND SPACES"`
	const want = "A VPN WITH LONG NAME AND SPACES"
	b := newBlock("", 1, opener)
	if b.Name != want {
		t.Errorf("newBlock(%q).Name = %q, want %q", opener, b.Name, want)
	}
	if b.VPN != want {
		t.Errorf("newBlock(%q).VPN = %q, want %q", opener, b.VPN, want)
	}
}

// TestNewBlockVPNQualifierWithSpaces is the client-username-side twin of
// TestNewBlockVPNNameWithSpaces: any object scoped to a multi-word VPN name is
// mis-attributed the same way, which is what would actually break the
// VPN-scope filter (export/import) on a real capture containing this VPN.
//
// BUG: see TestNewBlockVPNNameWithSpaces -- vpnQualifier truncates on the same
// strings.Fields split.
func TestNewBlockVPNQualifierWithSpaces(t *testing.T) {
	const opener = `client-username "app1" message-vpn "A VPN WITH LONG NAME AND SPACES"`
	const want = "A VPN WITH LONG NAME AND SPACES"
	b := newBlock("", 1, opener)
	if b.Name != "app1" {
		t.Errorf("newBlock(%q).Name = %q, want %q", opener, b.Name, "app1")
	}
	if b.VPN != want {
		t.Errorf("newBlock(%q).VPN = %q, want %q", opener, b.VPN, want)
	}
}

// --- section attribution -----------------------------------------------------

// TestParseBlocksSectionAttribution covers the "! Configure X" / "! Create X"
// -> Block.Section mapping: the "! " prefix and a trailing ":" are stripped,
// but a quoted operand inside the comment (which does not end in ":") survives
// untouched.
func TestParseBlocksSectionAttribution(t *testing.T) {
	c := loadSample(t)

	cases := []struct {
		index int
		want  string
	}{
		{1, `Create All: "Message Vpns"`},
		{3, "Configure System"},
		{6, `Configure ip vrf: "management"`}, // quoted operand: no trailing ":" to strip
	}
	for _, tc := range cases {
		b := blockByIndex(t, c.Blocks, tc.index)
		if b.Section != tc.want {
			t.Errorf("block %d Section = %q, want %q", tc.index, b.Section, tc.want)
		}
	}
}

// --- block boundaries ---------------------------------------------------------

// TestParseBlocksPragmaBetweenOpenerAndChildren covers the trap named in
// blocks.go: a "! pragma:..." line sitting at column 0 BETWEEN an opener and
// its indented children must not end the block, or the object silently loses
// every attribute after the pragma. Modelled on semp/all.cli lines 45-48
// (`create interface "intf0" ...` / pragma / os-physical-interface / exit).
//
// It also pins the OTHER half of the same trap, which a real capture puts
// immediately before this one: the OPENING pragma sits between the section
// comment and the opener, so the section comment has already closed the previous
// block and the pragma arrives with none open. It has to be HELD and attached to
// the block that follows -- an earlier parser sent it to Preamble instead, which
// detached `ignore-already-exists` from the `create` line it exists to make
// replayable. So both pragmas belong to this block, in source order.
func TestParseBlocksPragmaBetweenOpenerAndChildren(t *testing.T) {
	c := loadSample(t)
	b := blockByIndex(t, c.Blocks, 5)

	if b.Kind != "interface" || b.Name != "intf0" || !b.Create {
		t.Fatalf("block 5 = {Kind:%q Name:%q Create:%v}, want {interface intf0 true}", b.Kind, b.Name, b.Create)
	}
	want := []string{
		`! pragma:interpreter:ignore-already-exists`,
		`create interface "intf0" vrf management`,
		`! pragma:interpreter:no-ignore-already-exists`,
		`  os-physical-interface "eth0"`,
		`  exit`,
	}
	if len(b.Lines) != len(want) {
		t.Fatalf("block 5 Lines = %v, want %v", b.Lines, want)
	}
	for i, line := range want {
		if b.Lines[i] != line {
			t.Errorf("block 5 Lines[%d] = %q, want %q", i, b.Lines[i], line)
		}
	}

	// A parser that treated the pragma as a boundary would split this into two
	// (or more) objects instead of one.
	n := 0
	for _, blk := range c.Blocks {
		if blk.Kind == "interface" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("found %d %q blocks, want exactly 1 -- the pragma must not split the block", n, "interface")
	}
}

// TestParseBlocksNestedSubBlockOwnsBothExits is the critical case: block
// boundaries come from INDENTATION, never from matching "exit". The fixture's
// message-vpn block (index 7) nests two sub-blocks, "authentication" and
// "authorization", each closed by its OWN "exit" at indent 4 -- one level
// DEEPER than the outer message-vpn's own closing "exit" at indent 2. Modelled
// on bash/cliscripts-templates/disable-vpn-default.cli lines 5-51
// (message-vpn nesting authentication/authorization, each with its own exit
// before the outer one) and semp/all.cli lines 291-316 (authentication nesting
// basic/ldap-profile the same way). An exit-MATCHING parser would stop at the
// first nested "exit" and produce two or three objects here instead of one.
func TestParseBlocksNestedSubBlockOwnsBothExits(t *testing.T) {
	c := loadSample(t)
	b := blockByIndex(t, c.Blocks, 7)

	if b.Kind != "message-vpn" || b.Name != "simple-vpn" || b.VPN != "simple-vpn" {
		t.Fatalf("block 7 = {Kind:%q Name:%q VPN:%q}, want {message-vpn simple-vpn simple-vpn}",
			b.Kind, b.Name, b.VPN)
	}
	if len(b.Lines) != 10 {
		t.Fatalf("block 7 has %d lines, want 10 (opener + 8 body lines + the outer exit): %v",
			len(b.Lines), b.Lines)
	}
	exits := 0
	for _, line := range b.Lines {
		if strings.TrimSpace(line) == "exit" {
			exits++
		}
	}
	if exits != 3 {
		t.Errorf(`block 7 contains %d "exit" lines, want 3 (authentication's, authorization's, `+
			`and the outer message-vpn's own)`, exits)
	}

	// Exactly one object was produced for the whole message-vpn, not one per
	// nested exit. (The fixture also has a name-only "create message-vpn
	// simple-vpn / exit" pass at index 2, which is why this checks for body
	// content rather than just Kind+Name.)
	n := 0
	for _, blk := range c.Blocks {
		if blk.Kind == "message-vpn" && blk.Name == "simple-vpn" && len(blk.Lines) > 2 {
			n++
		}
	}
	if n != 1 {
		t.Errorf(`found %d message-vpn "simple-vpn" blocks with body content, want exactly 1`, n)
	}
}

// --- truncation / malformed-input guards --------------------------------------

// TestParseBlocksMissingEndTerminatorRefused covers the truncation guard:
// "! END" is the only evidence a capture is complete (`show current-config`
// writes to stdout, and a truncated transfer is still valid CLI), so its
// absence must be a loud refusal, not a partial parse that looks like success.
func TestParseBlocksMissingEndTerminatorRefused(t *testing.T) {
	const in = `! SEMP Schema Version: "broker/10.26.0.8827"

home
enable
configure

! Configure System:
system scaling  scale  max-connections 1000
`
	_, err := ParseBlocks([]byte(in))
	if err == nil {
		t.Fatal(`ParseBlocks should refuse a capture with no "! END" terminator`)
	}
	if !strings.Contains(err.Error(), "END") {
		t.Errorf(`error %q should name the missing "! END" terminator`, err)
	}
}

// TestParseBlocksIndentedLineWithNoOpenerRefused covers the other truncation
// guard: an indented line with no preceding opener means the capture began
// mid-block, which only a truncated or hand-edited artifact does. Refusing
// beats silently dropping the orphaned attributes.
func TestParseBlocksIndentedLineWithNoOpenerRefused(t *testing.T) {
	const in = `! SEMP Schema Version: "broker/10.26.0.8827"

home
enable
configure

  exit
! END
`
	_, err := ParseBlocks([]byte(in))
	if err == nil {
		t.Fatal("ParseBlocks should refuse an indented line with no preceding object")
	}
	if !strings.Contains(err.Error(), "no preceding object") {
		t.Errorf("error %q should say the indented line has no preceding object", err)
	}
}

// TestParseBlocksCRLFMatchesLF pins that a capture crossing a Windows host
// (\r\n line endings) parses identically to one that did not: splitLines
// strips a trailing \r from every line for exactly this reason, and this repo
// has a standing problem with fixtures flipping line endings, so it is
// asserted here rather than assumed.
func TestParseBlocksCRLFMatchesLF(t *testing.T) {
	lf, err := os.ReadFile(filepath.Join("testdata", "currentconfig_sample.cli"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	crlf := []byte(strings.ReplaceAll(string(lf), "\n", "\r\n"))

	wantCap, err := ParseBlocks(lf)
	if err != nil {
		t.Fatalf("ParseBlocks(LF): %v", err)
	}
	gotCap, err := ParseBlocks(crlf)
	if err != nil {
		t.Fatalf("ParseBlocks(CRLF): %v", err)
	}
	if !reflect.DeepEqual(wantCap, gotCap) {
		t.Errorf("CRLF capture differs from LF capture:\nLF:   %+v\nCRLF: %+v", wantCap, gotCap)
	}
}

package broker

import (
	"fmt"
	"regexp"
	"strings"
)

// Line-level transformations an artifact needs before it can be replayed, ported
// from semp/CurrentConfigParser.html -- the operator's hand-built GUI that has been
// preparing these captures for migration by hand.
//
// Everything here is a pure function over []Block, so all of it is testable against
// the real capture (semp/all.cli) with no broker.
//
// The distinction from inject.go matters. InjectShutdown works on BLOCKS: a
// broker-level `service semp listen-port 8080` is its own column-0 block, so a
// shutdown can be injected as a block beside it. The VPN-level equivalents are
// INDENTED CHILDREN of one `message-vpn "X"` block, so they are not blocks at all
// and no amount of block-level work reaches them. That is why they were missed:
// there are 48 of them in the real capture and none was handled.

// vpnServiceRE matches a VPN-level service port line and splits it into the parts
// the shutdown form is built from:
//
//	1 indent  2 "no "  3 rest-direction  4 service  5 " ssl"  6 " web"
//
// The port VALUE is optional because a real capture carries both forms -- the
// enabling `service mqtt listen-port 1883` and the clearing `no service mqtt
// listen-port` (semp/all.cli:442) -- and both change the port, so both need the
// service down first.
var vpnServiceRE = regexp.MustCompile(
	`^(\s+)(no )?service (?:rest (incoming|outgoing)|([a-z0-9-]+)) listen-port(?: "?[0-9]+"?)?( ssl)?( web)?\s*$`)

// InjectVPNServiceShutdown puts a `shutdown` for the right service AND TRANSPORT
// immediately before every VPN-level `service ... listen-port` line.
//
// The broker refuses a port change while the transport is up, and a real capture
// emits the two in the wrong order to replay: semp/all.cli:442-443 is
//
//	no service mqtt listen-port
//	service mqtt plain-text shutdown
//
// -- the port is cleared first and the shutdown lands after it, so a straight replay
// has the port change rejected. `cli -Apes` exits 0 regardless, so nothing says so.
//
// Transport granularity is the whole difficulty, and the reason this cannot be one
// blanket `service <x> shutdown`: mqtt alone carries four transports in that capture
// (plain-text, ssl, websocket, websocket-secure, lines 442-449) and taking the
// service down as a whole to change one port would drop the other three. The `web`
// suffix selects the websocket forms, `ssl` the secure one, and an untagged line
// means plain-text.
//
// Unlike InjectShutdown this never has to guess an enablement state: the capture's
// own trailing `service ... shutdown` / `no service ... shutdown` line follows and
// is passed through untouched, so the final state is always the artifact's. A
// duplicate shutdown (the capture already shutting that transport down) is harmless
// -- shutting down what is already down is a no-op -- but it is skipped anyway, so
// the applied script says what it means.
//
// The discriminator is INDENTATION, not the region: vpnServiceRE requires leading
// whitespace, and in the real capture every indented `service ... listen-port` line
// is inside a message-vpn block while every broker-level one sits at column 0 (where
// InjectShutdown already handles it). Gating on Region as well would add nothing and
// could only cause a silent skip if a block were ever mis-regioned.
func InjectVPNServiceShutdown(blocks []Block) []Block {
	out := make([]Block, 0, len(blocks))
	for _, b := range blocks {
		b.Lines = injectPortShutdownLines(b.Lines)
		out = append(out, b)
	}
	return out
}

// injectPortShutdownLines is InjectVPNServiceShutdown's per-block half.
func injectPortShutdownLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		m := vpnServiceRE.FindStringSubmatch(line)
		if m == nil {
			out = append(out, line)
			continue
		}
		shutdown := vpnShutdownLine(m)
		// Already guarded by the line immediately above: emit nothing.
		if len(out) == 0 || strings.TrimSpace(out[len(out)-1]) != strings.TrimSpace(shutdown) {
			out = append(out, shutdown)
		}
		out = append(out, line)
	}
	return out
}

// vpnShutdownLine builds the shutdown line for one matched port line, preserving
// the original's indentation so the result still parses as the same child.
func vpnShutdownLine(m []string) string {
	indent, restDir, service, ssl, web := m[1], m[3], m[4], m[5], m[6]

	name := service
	if restDir != "" {
		// rest names its direction where other services name nothing:
		// `service rest incoming ...`.
		name = "rest " + restDir
	}

	transport := "plain-text"
	switch {
	case web != "" && ssl != "":
		transport = "websocket-secure"
	case web != "":
		transport = "websocket"
	case ssl != "":
		transport = "ssl"
	}
	return fmt.Sprintf("%sservice %s %s shutdown", indent, name, transport)
}

// defaultClientUsernameRE matches the opener of the `default` client-username block
// inside a VPN.
var defaultClientUsernameRE = regexp.MustCompile(`^(\s*)(?:create )?client-username "default" message-vpn `)

// ShutdownDefaultClientUsername adds a `shutdown` as the first child of every
// `default` client-username block.
//
// The `default` client-username cannot be deleted -- confirmed on a live broker,
// and the reason the teardown capture simply omits it -- so unlike every other
// object in a VPN it is never recreated from scratch, it is EDITED in place. The
// broker refuses several of those edits while the username is enabled, and the
// capture carries no shutdown of its own because on the source broker it was
// already in its final state. The HTML adds this line for the same reason.
//
// The capture's own trailing `no shutdown` still follows and still decides the final
// state, so this cannot leave the username disabled.
func ShutdownDefaultClientUsername(blocks []Block) []Block {
	out := make([]Block, 0, len(blocks))
	for _, b := range blocks {
		if !defaultClientUsernameRE.MatchString(b.Opener()) {
			out = append(out, b)
			continue
		}
		// Only when it is not already there. Without this the transform is not
		// idempotent, and a re-run after a failed import -- which is the documented
		// recovery -- would apply a different script than the first attempt did.
		if at := openerIndex(b.Lines) + 1; at >= len(b.Lines) || !containsLine(b.Lines[at:at+1], "shutdown") {
			b.Lines = insertAfterOpener(b.Lines, "  shutdown")
		}
		out = append(out, b)
	}
	return out
}

// bridgeNoShutdownRE matches a `no shutdown` line at any depth inside a bridge.
var bridgeNoShutdownRE = regexp.MustCompile(`^(\s*)no shutdown\s*$`)

// BridgeEnablementInverted reports whether line is one ShutdownBridges deliberately
// applies the OPPOSITE of.
//
// The verification diff has to know. ImportVerify compares the untransformed
// ARTIFACT against a fresh read of the target, directionally -- every artifact line
// must be present on the target -- and ShutdownBridges is the one transform that
// changes a line rather than adding one. So the artifact's `no shutdown` can never
// be present: the import applied `shutdown` on purpose. Without this the diff
// reported a false failure on EVERY artifact containing a bridge, after a perfectly
// clean apply, which is the worst possible behaviour from the feature's only verdict.
//
// It belongs with the ignored lines rather than the checked ones for the same reason
// a skipped section's lines do: it was never intended to land.
func BridgeEnablementInverted(b Block, line string) bool {
	return b.Kind == kindBridge && bridgeNoShutdownRE.MatchString(line)
}

// ShutdownBridges provisions every bridge in the shut-down state, by turning each
// `no shutdown` inside a `create bridge` block into `shutdown`.
//
// This is the one transformation here that deliberately CHANGES the state the
// artifact asked for, and it is the HTML's oldest rule. A bridge comes up the
// instant it is un-shut, connects to the remote broker named in the capture, and
// starts moving messages -- so a bridge brought up by an import is an import that
// has begun replicating traffic between two brokers before anyone has looked at the
// result. The operator un-shuts bridges afterwards, deliberately.
//
// It applies at every depth because a bridge carries a nested remote block whose own
// `no shutdown` (HTML: the `      no shutdown` case) starts the remote connection.
func ShutdownBridges(blocks []Block) []Block {
	out := make([]Block, 0, len(blocks))
	for _, b := range blocks {
		if b.Kind != kindBridge {
			out = append(out, b)
			continue
		}
		lines := make([]string, len(b.Lines))
		for i, line := range b.Lines {
			if m := bridgeNoShutdownRE.FindStringSubmatch(line); m != nil {
				lines[i] = m[1] + "shutdown"
				continue
			}
			lines[i] = line
		}
		b.Lines = lines
		out = append(out, b)
	}
	return out
}

// adminUsernameRE matches the CLI admin's creation line.
var adminUsernameRE = regexp.MustCompile(`^(\s*)create username "admin" `)

// RewriteAdminUsername turns `create username "admin" ...` into `username admin`.
//
// The admin CLI user already exists on every broker -- it is what the operator's own
// kubectl/docker access reaches the CLI through -- so `create` is rejected as a
// duplicate and every attribute under it is lost with the block. Addressing the
// existing user instead is what the HTML does, and it is what makes the rest of the
// `Create Usernames` section (which the operator decided to apply in full) actually
// land.
func RewriteAdminUsername(blocks []Block) []Block {
	out := make([]Block, 0, len(blocks))
	for _, b := range blocks {
		m := adminUsernameRE.FindStringSubmatch(b.Opener())
		if m == nil {
			out = append(out, b)
			continue
		}
		lines := make([]string, len(b.Lines))
		copy(lines, b.Lines)
		lines[openerIndex(b.Lines)] = m[1] + "username admin"
		b.Lines = lines
		b.Create = false
		out = append(out, b)
	}
	return out
}

// profileKind reports whether a block is one of the three VPN objects that must be
// applied after the rest of their VPN.
func profileKind(b Block) bool {
	switch b.Kind {
	case "client-profile", "acl-profile", "client-username":
		return true
	}
	return false
}

// ReorderVPNProfiles moves every client-profile, acl-profile and client-username to
// the END of its VPN's blocks, preserving their relative order.
//
// A client-username names the client-profile and acl-profile it uses, and the broker
// refuses a username that references a profile it does not yet have. A capture emits
// them in its own order, which is not that dependency order -- so replaying it
// straight loses usernames, silently. The HTML solves this by extracting all three
// into holding variables and re-appending them at `! Enable Message Vpn`; this is
// the same reordering expressed over blocks.
//
// The relative order WITHIN the three kinds is preserved rather than sorted, because
// profiles can reference each other and the capture's order already satisfies that.
func ReorderVPNProfiles(blocks []Block) []Block {
	out := make([]Block, 0, len(blocks))

	flush := func(held []Block) []Block {
		out = append(out, held...)
		return nil
	}

	var held []Block
	current := ""
	for _, b := range blocks {
		if b.Region != RegionVPN {
			held = flush(held)
			current = ""
			out = append(out, b)
			continue
		}
		if b.VPN != current {
			// A new VPN begins: everything held for the previous one lands before it.
			held = flush(held)
			current = b.VPN
		}
		if profileKind(b) {
			held = append(held, b)
			continue
		}
		out = append(out, b)
	}
	flush(held)
	return out
}

// CollapseDuplicateLines removes a line that exactly repeats the one before it,
// within a block.
//
// The HTML does this at the stream level (`lines[ln] != lines[ln-1]`) because its
// own insertions can repeat a line the capture already had. The same is true here,
// and a repeated configuration line is at best noise in the applied script and at
// worst a second `create` of something that now exists.
//
// Blank lines are exempt: consecutive blanks are formatting, and collapsing them
// would reflow the artifact without changing what it configures.
func CollapseDuplicateLines(blocks []Block) []Block {
	out := make([]Block, 0, len(blocks))
	for _, b := range blocks {
		lines := make([]string, 0, len(b.Lines))
		for i, line := range b.Lines {
			if i > 0 && line == b.Lines[i-1] && strings.TrimSpace(line) != "" {
				continue
			}
			lines = append(lines, line)
		}
		b.Lines = lines
		out = append(out, b)
	}
	return out
}

// openerIndex is the position of a block's opener in Lines: held `! pragma:` comment
// lines are prepended, so it is not always zero (the bug Block.Opener() exists for).
func openerIndex(lines []string) int {
	for i, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "!") {
			return i
		}
	}
	return 0
}

// insertAfterOpener puts line directly after a block's opener.
func insertAfterOpener(lines []string, line string) []string {
	at := openerIndex(lines) + 1
	out := make([]string, 0, len(lines)+1)
	out = append(out, lines[:at]...)
	out = append(out, line)
	return append(out, lines[at:]...)
}

// ClearExistingSyslogs emits `no syslog "<name>"` before each `create syslog
// "<name>"` the artifact carries -- but ONLY for a name the TARGET already has.
//
// Both halves are load-bearing and they pull in opposite directions. A `create
// syslog` for an entry that already exists is rejected, so the removal has to come
// first; but `no syslog` for an entry that does NOT exist is itself an error, so
// emitting it unconditionally would fail every import onto a broker with no syslog
// configured -- which is most of them. Only the target's own configuration settles
// it, which is why existing comes from the capture ImportPlan already takes of the
// target rather than from a guess or a flag.
//
// The removal is `no syslog <name>` at configure level, per the operator. The
// mode commands that get there (`home`/`enable`/`configure`) already open every
// chunk -- preambleForApply emits them -- so nothing extra is needed here.
//
// The operand is reproduced EXACTLY as the artifact renders it, quotes and all,
// rather than re-quoted from Block.Name. A real capture writes
// `! Create Syslog: "test-syslog"` / `create syslog "test-syslog"`
// (semp/all.cli:239-240), so the quoted form is what this produces today -- but
// mirroring the source is what keeps that true without depending on it, and it
// cannot introduce a quoting style the broker does not itself produce.
//
// The section header carries the destination name, so each syslog entry gets its
// own `! Create Syslog: "<name>"` section and the rule for it is matched by PREFIX
// (sections.go's prefixSections).
func ClearExistingSyslogs(blocks []Block, existing []string) []Block {
	if len(existing) == 0 {
		return blocks
	}
	have := make(map[string]bool, len(existing))
	for _, n := range existing {
		have[n] = true
	}

	out := make([]Block, 0, len(blocks))
	for _, b := range blocks {
		if b.Kind == kindSyslog && b.Create && have[b.Name] {
			if m := syslogCreateRE.FindStringSubmatch(b.Opener()); m != nil {
				removal := fmt.Sprintf("%sno syslog %s", m[1], m[2])
				// Only when this transform has not already put it there. The removal
				// is a BLOCK of its own immediately before the create, so a re-run
				// would otherwise prepend a second one to the same create, and a
				// third the pass after -- and re-running a failed import is the
				// documented recovery, so the second attempt has to apply the same
				// script as the first.
				if len(out) == 0 || !containsLine([]string{out[len(out)-1].Opener()}, removal) {
					out = append(out, injectedBlock(b, b.Section, removal))
				}
			}
		}
		out = append(out, b)
	}
	return out
}

// syslogCreateRE captures a `create syslog` opener's indentation and its operand AS
// WRITTEN -- the quoted form when the capture quoted it, the bare word when it did
// not. See ClearExistingSyslogs for why the rendering is mirrored rather than chosen.
var syslogCreateRE = regexp.MustCompile(`^(\s*)create\s+syslog\s+("[^"]*"|\S+)`)

// --- nested objects: created INSIDE a parent block ---------------------------
//
// Several broker-level sections do not create their object at column 0. They open a
// scope and create inside it:
//
//	! Create Client Certificate Authority: "G2-inter"
//	authentication
//	  create client-certificate-authority "G2-inter"
//
// The same shape carries domain CAs (parent `ssl`) and virtual hostnames (parent
// `service`). Two consequences follow and both were missed.
//
// The Block's Kind is the PARENT (`authentication`), not the object -- so nothing
// keyed on Kind reaches these, the same way block-level rules could not reach the
// VPN-level service lines. And a `create <verb> "<name>"` for a name the target
// already has is REJECTED, so none of these three sections can be replayed onto a
// target carrying any of them without a removal first. No capture emits one: the
// broker's own `show current-config` writes only the creates, which is why the
// conflict is invisible until an import fails on it.
// kindBridge is the opener kind of a message-VPN bridge.
const kindBridge = "bridge"

const (
	verbClientCA        = "client-certificate-authority"
	verbDomainCA        = "domain-certificate-authority"
	verbVirtualHostname = "virtual-hostname"
	verbLDAPProfile     = "ldap-profile"
)

// nestedCreateRE matches `create <verb> <operand>` at any indent, capturing the
// indentation and the operand AS WRITTEN.
func nestedCreateRE(verb string) *regexp.Regexp {
	return regexp.MustCompile(`^(\s*)create\s+` + regexp.QuoteMeta(verb) + `\s+("[^"]*"|\S+)`)
}

// nestedNames lists the objects a capture creates under verb, in first-seen order.
// Used against the TARGET's capture to learn which names a removal would be valid
// for -- a `no <verb>` for a name that is not there is itself an error.
func (c *Capture) nestedNames(verb string) []string {
	re := nestedCreateRE(verb)
	var out []string
	seen := map[string]bool{}
	for _, b := range c.Blocks {
		for _, line := range b.Lines {
			m := re.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			name := strings.Trim(m[2], `"`)
			if name == "" || seen[name] {
				continue
			}
			seen[name] = true
			out = append(out, name)
		}
	}
	return out
}

// clearNestedObjects inserts `no <verb> "<name>"` for every name in clear, directly
// after the opener of each block that creates objects of that verb.
//
// Each removal is placed IMMEDIATELY BEFORE its own create, at that create's own
// indentation, which is the only placement that is correct at every nesting depth.
// The objects are not all two levels deep: a client CA, domain CA, LDAP profile and
// virtual hostname sit at indent 2 directly inside their parent scope, but an LDAP
// authorisation group sits at indent 6, inside `authentication > access-level >
// ldap`. Grouping the removals at the top of the block would have emitted a
// `no group ...` while still in the `authentication` scope, which is not where the
// group lives.
//
// It also means this function never has to invent a scope. A name the TARGET has
// that the artifact does not re-create has no create to sit before, and that case is
// handled separately (TargetOnlyRemovals), because only the clear-everything rules
// can produce one.
//
// VERIFIED against Solace's own CLI reference (semp/appliance_cli_reference.html),
// which documents each of these as a single command with both verbs:
//
//	[create | no] ldap-profile <profile-name>
//	[create | no] domain-certificate-authority <ca-name>
//	[create | no] client-certificate-authority <ca-name>
//	[create | no] syslog <name>
//	[create | no] virtual-hostname <name>
//
// so `no <verb> <name>` in the create's own scope is the documented form, not an
// inference from the shape of other objects.
func clearNestedObjects(blocks []Block, verb string, clear []string) []Block {
	if len(clear) == 0 {
		return blocks
	}
	want := make(map[string]bool, len(clear))
	for _, n := range clear {
		want[n] = true
	}
	re := nestedCreateRE(verb)

	out := make([]Block, 0, len(blocks))
	for _, b := range blocks {
		lines := make([]string, 0, len(b.Lines)+len(clear))
		changed := false
		for i, line := range b.Lines {
			m := re.FindStringSubmatch(line)
			if m == nil || !want[strings.Trim(m[2], `"`)] {
				lines = append(lines, line)
				continue
			}
			// Mirror the artifact's own rendering of the name, which is what carries
			// an LDAP group's escaped commas (`"CN=*X\\, Y,OU=..."`) through intact.
			removal := m[1] + "no " + verb + " " + m[2]
			// Not if this transform already put it there: a re-run after a failed
			// import is the documented recovery, so the second attempt has to build
			// the same script as the first.
			if i == 0 || !containsLine(b.Lines[i-1:i], removal) {
				lines = append(lines, removal)
				changed = true
			}
			lines = append(lines, line)
		}
		if changed {
			b.Lines = lines
		}
		out = append(out, b)
	}
	return out
}

// containsLine reports whether lines already carries an identical (trimmed) line,
// which is what keeps clearNestedObjects idempotent across a re-run.
func containsLine(lines []string, want string) bool {
	for _, l := range lines {
		if strings.TrimSpace(l) == strings.TrimSpace(want) {
			return true
		}
	}
	return false
}

// intersect returns the entries of a that are also in b, in a's order.
func intersect(a, b []string) []string {
	have := make(map[string]bool, len(b))
	for _, s := range b {
		have[s] = true
	}
	var out []string
	for _, s := range a {
		if have[s] {
			out = append(out, s)
		}
	}
	return out
}

// ClearExistingNested removes every nested object the TARGET already has and this
// artifact re-creates, across the three sections that create one.
//
// They are one rule because they are one failure: `create <verb> "<name>"` for a
// name the target already carries is REJECTED, and no capture emits a removal --
// the broker's own `show current-config` writes only the creates, so the conflict is
// invisible until an import fails on it. All three shipped applied-but-unguarded,
// and two of them carried a Reason claiming a delete-then-create that did not exist
// anywhere.
//
// The LDAP authorisation groups in `Create Authentication` are deliberately NOT here
// (operator-confirmed): `create group` inside `authentication > access-level >
// ldap` is not one of the objects that needs clearing, even though it reads like the
// same shape at a glance.
//
// The intersection is what keeps it safe in the other direction: a removal for a
// name the target does NOT have is itself an error, and clearing an object the
// target's own operator added and this artifact never mentions would delete trust
// material, an LDAP bind or an authorisation group nobody asked about.
func ClearExistingNested(blocks []Block, target TargetState) []Block {
	for _, e := range []struct {
		verb   string
		onHost []string
	}{
		{verbClientCA, target.ClientCAs},
		{verbDomainCA, target.DomainCAs},
		{verbLDAPProfile, target.LDAPProfiles},
	} {
		blocks = clearNestedObjects(blocks, e.verb,
			intersect(nestedArtifactNames(blocks, e.verb), e.onHost))
	}
	return blocks
}

// ClearTargetVirtualHostnames removes EVERY virtual hostname the target has, so it
// ends up matching the artifact rather than carrying the union of both.
//
// This one deliberately does not intersect with the artifact. A virtual hostname is
// a DNS name that resolves to this broker, so one left behind from the target's
// previous life keeps answering for a name the imported configuration never claimed
// -- which is the opposite failure from the client-CA case, and why the two sections
// get opposite rules despite the identical shape.
func ClearTargetVirtualHostnames(blocks []Block, targetNames []string) []Block {
	return clearNestedObjects(blocks, verbVirtualHostname, targetNames)
}

// nestedParent is the column-0 scope each nested object is created inside, and so
// the scope a removal has to be issued in.
var nestedParent = map[string]string{
	verbClientCA:        "authentication",
	verbDomainCA:        "ssl",
	verbVirtualHostname: "service",
	verbLDAPProfile:     "authentication",
}

// TargetOnlyRemovals renders removals for objects the TARGET carries that the
// artifact does not re-create, as a standalone script opening its own scope.
//
// clearNestedObjects can only place a removal inside a block that already creates
// objects of that verb -- that block is where the parent scope and the indentation
// come from. When the artifact creates NONE of them, there is no such block and the
// entire clear silently did nothing. That case is not hypothetical and it defeats
// the rule outright: a target carrying virtual hostnames, importing an artifact that
// has none, is exactly when "clear the target's first so it matches the artifact"
// has to remove something, and it removed nothing.
//
// Only the clear-EVERYTHING rules can produce these. The intersecting ones (client
// CAs) clear only names the artifact re-creates, so by construction every removal
// has a create to sit beside.
//
// The removal verb itself is documented (`[create | no] virtual-hostname <name>`,
// semp/appliance_cli_reference.html). What is NOT confirmed is this variant's need
// to open the `service` scope itself, since no capture emits a removal at all.
func TargetOnlyRemovals(blocks []Block, target TargetState) []string {
	var out []string
	// Virtual hostnames only: see above for why the intersecting rules cannot.
	names := target.VirtualHostnames
	if len(names) == 0 {
		return nil
	}
	have := (&Capture{Blocks: blocks}).nestedNames(verbVirtualHostname)
	artifactHas := make(map[string]bool, len(have))
	for _, n := range have {
		artifactHas[n] = true
	}

	var body []string
	for _, n := range names {
		if artifactHas[n] {
			continue // clearNestedObjects places this one beside its create
		}
		body = append(body, fmt.Sprintf("  no %s %q", verbVirtualHostname, n))
	}
	if len(body) == 0 {
		return nil
	}
	out = append(out, nestedParent[verbVirtualHostname])
	out = append(out, body...)
	return append(out, "  exit")
}

// TargetState is what the apply needs to know about the TARGET broker, read once at
// plan time from its own capture.
//
// Every field is something that CANNOT be decided from the artifact: whether a
// removal is valid depends on what the target already carries, and getting it wrong
// fails in both directions -- a missing removal makes the create a rejected
// duplicate, and a removal for an absent object is itself an error.
type TargetState struct {
	Syslogs          []string
	ClientCAs        []string
	VirtualHostnames []string
	DomainCAs        []string
	LDAPProfiles     []string
}

// readTargetState extracts everything the apply needs from a target capture.
func readTargetState(c *Capture) TargetState {
	return TargetState{
		Syslogs:          c.SyslogNames(),
		ClientCAs:        c.nestedNames(verbClientCA),
		VirtualHostnames: c.nestedNames(verbVirtualHostname),
		DomainCAs:        c.nestedNames(verbDomainCA),
		LDAPProfiles:     c.nestedNames(verbLDAPProfile),
	}
}

// PrepareForApply runs every replay transformation over an artifact's blocks, in the
// order they have to happen.
//
// target is what the TARGET already has: the transformations that consult it are the
// ones that cannot be decided from the artifact alone.
//
// Order is not arbitrary. The reorder moves whole blocks and must run before the
// duplicate collapse, which compares neighbouring lines; the admin rewrite must run
// before anything keys on Create; and the injections add lines the collapse then
// dedupes against what the capture already carried.
func PrepareForApply(blocks []Block, target TargetState) []Block {
	blocks = RewriteAdminUsername(blocks)
	blocks = ClearExistingSyslogs(blocks, target.Syslogs)

	blocks = ClearExistingNested(blocks, target)
	blocks = ClearTargetVirtualHostnames(blocks, target.VirtualHostnames)
	blocks = ShutdownDefaultClientUsername(blocks)
	blocks = ShutdownBridges(blocks)
	blocks = InjectVPNServiceShutdown(blocks)
	blocks = ReorderVPNProfiles(blocks)
	return CollapseDuplicateLines(blocks)
}

// nestedArtifactNames is nestedNames over a bare block slice, for the artifact side
// (which is []Block rather than a *Capture by the time the transforms run).
func nestedArtifactNames(blocks []Block, verb string) []string {
	return (&Capture{Blocks: blocks}).nestedNames(verb)
}

package broker

import "strings"

// This file is the baked-in classification of what `import-config` does with
// each broker-level section of a `show current-config` capture (blocks.go
// parses the capture; this file decides what happens to each Block it finds).
//
// It is baked in ON PURPOSE: there is no config key, no flag and no runtime
// decision file that could change one of these classifications. A decision
// file read at import time would be untrusted input choosing what executes --
// exactly what the execution guard in internal/config/execguard.go exists to
// prevent for the runtime/compose fields, and the same reasoning applies here.
// The rationale for every row is recorded in semp/import-decisions.md (the
// operator's own answers, against a real capture from broker 10.26.0.8827);
// docs/import.md is generated from Rules() below so the published docs cannot
// drift from what the code actually does.
//
// Line numbers cited in Reason strings and doc comments are semp/all.cli, the
// fixture the decisions were verified against -- its broker-level region runs
// lines 31-378; everything from line 379 (`! Just Create Message Vpn`) on is
// per-VPN and belongs to the VPN scope, not this table.

// Disposition is what import does with one section.
type Disposition string

const (
	Apply         Disposition = "apply"          // applied as-is
	ApplyFiltered Disposition = "apply-filtered"  // applied, minus specific lines
	Skip          Disposition = "skip"            // never applied, never prompted, no opt-in
	SkipAppliance Disposition = "skip-appliance"  // appliance-only; skipped on a software target
)

// SectionRule is the classification for one section, matched against the
// broker's own comment (Block.Section).
type SectionRule struct {
	Section     string            // matched against Block.Section
	Disposition Disposition       // apply / apply-filtered / skip / skip-appliance
	Reason      string            // one sentence, printed in the report and in docs/import.md
	DropLines   func(string) bool // ApplyFiltered only: true to drop this line
	Interrupts  string            // "" or what applying it interrupts, e.g. "all messaging"
	First       bool              // must be applied first, alone, in its own CLI invocation
}

// prefixSections lists the Section values matched by PREFIX instead of
// equality, because the broker appends a variable, deployment-chosen quoted
// operand to its own section comment for these -- e.g.
// `! Configure ip vrf: "management"` (semp/all.cli:50), so blocks.go's
// newBlock/parseBody (which strips only a single trailing ":") leaves
// Block.Section as `Configure ip vrf: "management"`, not `Configure ip vrf`.
// RuleFor matches such a Section by prefix rather than equality.
//
// Six of these carry a name an operator chose (an LDAP/RADIUS profile name, a
// domain CA name, an authentication profile name, a syslog destination name), so
// equality would need one rule per possible name. The odd one out is
// Create Redundancy PSK: its
// operand in this capture is `"#redundancy-group"` (semp/all.cli:340), and the
// leading "#" marks it as a Solace-internal singleton name rather than
// something an operator supplies -- in practice it never varies. It is still
// matched by prefix here, both for consistency with the rest of this list and
// because nothing in this table should depend on that staying true.
var prefixSections = []string{
	"Configure ip vrf",
	"Create LDAP Profile",
	"Create Radius Profile",
	"Create Domain Certificate Authority",
	"Create Client Certificate Authority",
	"Create Authentication",
	"Create Redundancy PSK",
	"Create Syslog",
}

// dropRoutingInterfaceLine drops the `interface "intf0"` operand out of a
// "Configure Routing" block (semp/all.cli:135). The name is host hardware, and
// the sections that create an interface by that name -- "Configure SolOsPhy
// Interfaces" (semp/all.cli:43-49) and "Configure ip vrf" (semp/all.cli:50-58)
// -- are themselves skip-appliance, so an import target may have no interface
// of that name to reference.
func dropRoutingInterfaceLine(line string) bool {
	return strings.TrimSpace(line) == `interface "intf0"`
}

// dropMessageSpoolLines drops the lines a "Create hardware" / "Message Spool"
// block uses to bounce or size the message spool itself -- the
// guaranteed-messaging data path -- while keeping its event thresholds and
// defragment schedule, which are ordinary broker settings:
//
//   - the shutdown/no-shutdown pair that brackets the whole block
//     (semp/all.cli:186, 200)
//   - `spool-sync mode` (semp/all.cli:187), a property of the HA pair the
//     target already has, not of the capture's source
//   - `max-spool-usage` (semp/all.cli:192), which on Kubernetes comes from the
//     CR and the PVC rather than from a CLI capture
//
// A plain substring match is enough for all three: "message-spool shutdown"
// matches both the shutdown and the no-shutdown line but not the unrelated
// `defragment-spool-files schedule shutdown` / `... threshold shutdown` lines
// (semp/all.cli:194, 197), which are the defragment schedule this rule keeps.
func dropMessageSpoolLines(line string) bool {
	t := strings.TrimSpace(line)
	return strings.Contains(t, "message-spool shutdown") ||
		strings.Contains(t, "spool-sync mode") ||
		strings.Contains(t, "max-spool-usage")
}

// sectionRules is the classification table, in the order its sections appear
// in semp/all.cli's broker-level region (lines 31-378). Rules() returns it
// as-is, so that order is also docs/import.md's row order.
//
// One section still has NO rule here: "Configure Routing DMR". It is empty in every
// capture in hand, so there is nothing to classify from.
//
// What must not be read into that is that it is empty on every broker. That claim
// was made here twice and was wrong twice: "Configure Virtual Hostnames" turned out
// to carry two real entries in semp/all-cfg-tps-org.cli:93, and "Configure Hostname"
// carries `hostname "<name>" defer` on an appliance
// (semp/show-currentconfig-all.out). Both are classified below now. An empty section
// proves nothing about any other broker, and the only reason this one is still
// unruled is that nobody has produced a capture where it is not empty.
//
// An unclassified section with CONTENT is handled correctly rather than silently:
// planSections skips it, warns by name in the plan (before the tear-down
// confirmation), and -- unlike a rule-based skip -- keeps its lines in the
// verification diff, so the import exits non-zero instead of dropping the
// configuration and reporting success. Classifying one properly needs a capture
// from a broker that actually uses it; until then the gap is loud, which is the
// point.
var sectionRules = []SectionRule{
	// semp/all.cli:31 -- On Kubernetes the CR sets connection/subscription
	// scaling at first boot and this is not durable, so replaying it here is
	// not a real fix.
	{
		Section:     "Configure System",
		Disposition: Skip,
		Reason: "On Kubernetes the CR sets connection/subscription scaling at first boot " +
			"(semp/all.cli:32) and this is not durable: it reverts on the next pod recreate.",
	},

	// semp/all.cli:39-63 -- appliance-only host/interface identity, empty in every
	// software capture. CONFIRMED against a real appliance
	// (semp/show-currentconfig-all.out, soltr/10_4_1): all of these carry real
	// content there -- Ethernet 26 lines, Lag 38, DNS 7, Clock 1, Clock
	// Synchronization 7, SNMP 40, and `ip vrf` TWICE ("management" and
	// "msg-backbone", which is also what validates the prefix match). Until that
	// capture existed these rules were reasoned from an empty section; now they are
	// evidence.
	//
	// "Configure ip vrf" is matched by prefix because the broker names the VRF in
	// its own section comment (see prefixSections).
	{
		Section:     "Configure Ethernet Interfaces",
		Disposition: SkipAppliance,
		Reason:      "Physical Ethernet interface configuration; empty on a software broker.",
	},
	{
		Section:     "Configure Lag Interfaces",
		Disposition: SkipAppliance,
		Reason:      "Link-aggregation interface configuration; empty on a software broker.",
	},
	{
		Section:     "Configure SolOsPhy Interfaces",
		Disposition: SkipAppliance,
		Reason: "Host NIC identity (semp/all.cli:45-48); on Kubernetes the interface " +
			"belongs to the pod, not to an imported artifact.",
	},
	{
		Section:     "Configure ip vrf",
		Disposition: SkipAppliance,
		Reason: "Host addressing under the management/matelink/msg-backbone VRFs " +
			"(semp/all.cli:50-58, 60, 62); matched by prefix since the broker names " +
			"the VRF in its own section comment.",
	},

	// semp/all.cli:64-74 -- SEMP is the channel import runs over.
	{
		Section:     "Configure SEMP Service",
		Disposition: Skip,
		Reason: "Lines 65-70 shut SEMP down to change its own port, and SEMP is the " +
			"channel import runs over, so applying this section would sever the " +
			"connection the import needs to finish.",
	},

	// semp/all.cli:76-124 -- messaging services and the data path: applied,
	// but each bounces something and the operator should know what.
	{
		Section:     "Configure SMF Service",
		Disposition: Apply,
		Interrupts:  "all messaging",
		Reason:      "Bounces SMF (semp/all.cli:77, 82) to apply its listen ports.",
	},

	// semp/all.cli:85-93 -- HA mate services: this capture is from a
	// standalone broker, so bouncing these buys nothing and only risks the
	// mate link (semp/import-decisions.md section C).
	{
		Section:     "Configure Matelink Service",
		Disposition: Skip,
		Reason:      "Bounces the HA mate link (semp/all.cli:86, 88) to apply its port.",
	},
	{
		Section:     "Configure Redundancy Service",
		Disposition: Skip,
		Reason:      "Bounces HA (semp/all.cli:91, 93) to apply its listen port.",
	},

	{
		Section:     "Configure SSL Service",
		Disposition: Apply,
		Reason:      "Connection-count thresholds only (semp/all.cli:96); no port change.",
	},
	{
		Section:     "Configure WEB Service",
		Disposition: Apply,
		Reason:      "Bounces web-transport (semp/all.cli:99, 103) to apply its ports.",
	},
	// semp/all.cli:105 "Configure Virtual Hostnames" has no rule: it is empty
	// in every software capture, a software-relevant no-op rather than
	// appliance-only, so there is nothing to classify (see the table comment).
	{
		Section:     "Configure REST Service",
		Disposition: Apply,
		Reason:      "Enables REST incoming/outgoing and their thresholds (semp/all.cli:108-110).",
	},
	{
		Section:     "Configure MQTT Service",
		Disposition: Apply,
		Reason:      "Starts MQTT back up (semp/all.cli:113).",
	},
	{
		Section:     "Configure AMQP Service",
		Disposition: Apply,
		Reason:      "Bounces AMQP (semp/all.cli:116, 118) to apply its port.",
	},

	// Empty in semp/all.cli, which is what once made this section look like a
	// permanent no-op; semp/all-cfg-tps-org.cli:93 carries two real virtual
	// hostnames. Same nested shape as the CA sections: opens `service` and creates
	// inside it.
	{
		Section:     "Configure Virtual Hostnames",
		Disposition: Apply,
		Reason: "Virtual hostnames and the message-VPN each is bound to, created " +
			"inside the `service` scope (semp/all-cfg-tps-org.cli:93-103). Every " +
			"virtual hostname already on the target is removed first, so the target " +
			"ends up matching the artifact rather than carrying the union: a virtual " +
			"hostname is a DNS name pointing at this broker, and a stale one left " +
			"behind resolves somewhere it should not.",
	},
	{
		Section:     "Configure HealthCheck Service",
		Disposition: Apply,
		Interrupts:  "the kubelet readiness probe",
		Reason: "Bounces the health-check service (semp/all.cli:121, 124); port 5550 " +
			"(semp/all.cli:122) is what render.go's healthPort targets, so a capture " +
			"with a different port fails every pod's readiness.",
	},

	// semp/all.cli:126 -- appliance-only; empty on a software broker.
	{
		Section:     "Configure DNS",
		Disposition: SkipAppliance,
		Reason:      "DNS server configuration; empty on a software broker.",
	},
	// semp/all.cli:128 "Configure Hostname" has no rule: same as Virtual
	// Hostnames above, empty and software-relevant rather than appliance-only.

	// semp/all.cli:130-131 -- node identity.
	{
		Section:     "Configure Router Name",
		Disposition: Skip,
		Reason: "Node identity (semp/all.cli:131); overwriting the target's router-name " +
			"breaks HA pairing, DMR and MNR.",
	},

	// Empty in both software captures, which is why it had no rule; an APPLIANCE
	// capture carries `hostname "tuas9csol1" defer`
	// (semp/show-currentconfig-all.out). Same class as Router Name directly above.
	{
		Section:     "Configure Hostname",
		Disposition: Skip,
		Reason: "The broker's own hostname (`hostname \"<name>\" defer`); node identity " +
			"like Router Name above, so copying the source's onto a target renames that " +
			"broker to its source.",
	},

	// semp/all.cli:133-140 -- apply-filtered: drops the one line that names
	// host hardware the skip-appliance sections above never created here.
	{
		Section:     "Configure Routing",
		Disposition: ApplyFiltered,
		DropLines:   dropRoutingInterfaceLine,
		Reason: "Drops `interface \"intf0\"` (semp/all.cli:135): that name is host " +
			"hardware, and the sections that create it are skip-appliance, so the " +
			"target may have no interface by that name.",
	},
	{
		Section:     "Configure Routing MNR",
		Disposition: Apply,
		Reason:      "CSPF SSL validation and queue thresholds (semp/all.cli:146-152); no shutdown.",
	},

	// semp/all.cli:156-169 -- the HA group itself.
	{
		Section:     "Configure Redundancy",
		Disposition: Skip,
		Reason: "Ends at `redundancy shutdown` (semp/all.cli:168) with no matching " +
			"un-shutdown, so a standalone capture applied to an HA pair leaves " +
			"redundancy down; it also sets `active-standby-role primary` " +
			"(semp/all.cli:159), which can split the group on the wrong node.",
	},

	// semp/all.cli:171-200 -- apply-filtered: "Create hardware" and "Disk" are
	// the two comments the broker emits ahead of "Message Spool"; "Create
	// hardware" carries no content on a software broker (its appliance-only
	// physical-hardware content, if any, would need the same filtering as
	// Message Spool, so it shares the rule), "Disk" is genuinely
	// appliance-only, and "Message Spool" is where this capture's actual
	// spool tuning lives.
	{
		Section:     "Create hardware",
		Disposition: ApplyFiltered,
		DropLines:   dropMessageSpoolLines,
		Reason:      "Empty on a software broker; classified with Message Spool for the case it is not.",
	},
	{
		Section:     "Disk",
		Disposition: SkipAppliance,
		Reason:      "Physical disk configuration; empty on a software broker.",
	},

	// Two more appliance-only `hardware` sections, found only once a real appliance
	// capture existed (semp/show-currentconfig-all.out). Of the four `hardware`
	// commands a capture can carry, `message-spool` appears on BOTH platforms (26
	// lines in each software capture, 25 here) while `disk`, `power-redundancy` and
	// `topic-routing` appear on the appliance alone -- which is what puts these two
	// beside Disk rather than beside Message Spool.
	{
		Section:     "Power Redundancy",
		Disposition: SkipAppliance,
		Reason: "Physical power-supply redundancy (`hardware power-redundancy \"1+1\"`); " +
			"an appliance chassis property with nothing to configure on a software broker.",
	},
	{
		Section:     "Topic Routing",
		Disposition: SkipAppliance,
		Reason: "`hardware topic-routing acl-topic-matching-mode`; issued as a hardware " +
			"command and present on no software capture, so a software target has " +
			"nowhere to apply it.",
	},
	{
		Section:     "Message Spool",
		Disposition: ApplyFiltered,
		DropLines:   dropMessageSpoolLines,
		Reason: "Applies the event thresholds and defragment schedule; drops " +
			"`max-spool-usage`, `spool-sync mode` and the shutdown/no-shutdown pair " +
			"(semp/all.cli:186, 187, 192, 200) that bounce the message spool -- the " +
			"guaranteed-messaging data path -- and that on Kubernetes are sized by " +
			"the CR and the PVC instead.",
	},
	// semp/all.cli:202 "Configure Routing DMR" has no rule: empty and
	// software-relevant, same as Virtual Hostnames and Hostname above.

	// semp/all.cli:204-221 -- apply, with a consequence recorded rather than
	// an objection: the operator overrode a skip recommendation for this one.
	{
		Section:     "Configure Replication",
		Disposition: Apply,
		Reason: "Carries a config-sync bridge pre-shared key (semp/all.cli:212) and " +
			"`mate virtual-router-name` (semp/all.cli:205), so the replication mate " +
			"must be re-imported to match.",
	},

	// semp/all.cli:223 -- appliance-only; empty on a software broker.
	{
		Section:     "Configure Clock",
		Disposition: SkipAppliance,
		Reason:      "Clock/timezone configuration; empty on a software broker.",
	},

	{
		Section:     "Configure Memory Event",
		Disposition: Apply,
		Reason:      "Event thresholds only (semp/all.cli:226-228).",
	},
	{
		Section:     "Configure Service",
		Disposition: Apply,
		Interrupts:  "msg-backbone, the data path",
		Reason:      "Starts msg-backbone back up (semp/all.cli:232) after its thresholds.",
	},
	{
		Section:     "Configure MQTT",
		Disposition: Apply,
		Reason:      "`mqtt retain max-memory 0` (semp/all.cli:235).",
	},

	// semp/all.cli:237 -- appliance-only; empty on a software broker.
	{
		Section:     "Configure Clock Synchronization",
		Disposition: SkipAppliance,
		Reason:      "NTP/clock-sync configuration; empty on a software broker.",
	},

	// semp/all.cli:239-247 -- one section PER syslog destination, so the header
	// carries a deployment-chosen quoted operand (`! Create Syslog: "test-syslog"`)
	// and it is matched by prefix (see prefixSections).
	{
		Section:     "Create Syslog",
		Disposition: Apply,
		Reason: "Syslog destinations: facilities and remote hosts with their transports " +
			"(semp/all.cli:239-247). Applied, and uniquely among these it needs the TARGET " +
			"consulted first -- a `create syslog` for an entry that already exists is " +
			"rejected, so ClearExistingSyslogs emits `no syslog <name>` ahead of it, but " +
			"only for a name the target actually has, because the removal is itself an " +
			"error when the entry does not exist.",
	},

	{
		Section:     "Configure Schedule",
		Disposition: Apply,
		Reason:      "Disables the autobackup schedule (semp/all.cli:240).",
	},
	{
		Section:     "Create Compression",
		Disposition: Apply,
		Reason:      "`compression mode optimize-for-size` (semp/all.cli:243).",
	},

	// semp/all.cli:245 -- appliance-only; empty on a software broker.
	{
		Section:     "Configure SNMP",
		Disposition: SkipAppliance,
		Reason:      "SNMP agent and trap configuration; empty on a software broker.",
	},

	{
		Section:     "Create console",
		Disposition: Apply,
		Reason:      "Console timeout and login banner (semp/all.cli:248-249).",
	},

	// semp/all.cli:251-260 -- the one section that ends the CLI session it
	// runs in, so it must be applied alone and first.
	{
		Section:     "Create logging",
		Disposition: Apply,
		First:       true,
		Interrupts:  "this CLI session",
		Reason: "Logging changes end the CLI session (semp/all.cli:251-260), so any " +
			"later command in the same invocation would never run; applying it " +
			"alone in its own RunCLI removes the hazard, since each RunCLI execs a " +
			"fresh cli process.",
	},

	{
		Section:     "Configure SSL",
		Disposition: Apply,
		Interrupts:  "live TLS sessions",
		Reason: "Cipher-suite and TLS-version changes (semp/all.cli:264-267) can drop " +
			"sessions already using the settings being replaced.",
	},

	// semp/all.cli:271-280 -- config-sync is what propagates VPN config to the
	// HA mate; line 279 shuts it down.
	{
		Section:     "Configure Config Sync",
		Disposition: Skip,
		Reason: "Line 279 shuts config-sync down, which is what propagates VPN " +
			"configuration to the HA mate.",
	},

	// semp/all.cli:282-374 -- credentials and authentication: apply, each with
	// its own risk noted; three of these are operator overrides of a skip
	// recommendation, and their Reason records the consequence rather than an
	// objection (semp/import-decisions.md section D).
	{
		Section:     "Create Usernames",
		Disposition: Apply,
		Reason: "Overwrites the target's CLI admin password with the source's " +
			"(semp/all.cli:284, 286); harmless to this tool, since the CLI channel " +
			"needs no broker credentials, but on Kubernetes adminCredentialsSecret " +
			"still holds the old value, so pods go out of readiness until " +
			"semp.adminPass is updated to match.",
	},
	{
		Section:     "Create LDAP Profile",
		Disposition: Apply,
		Reason: "LDAP profiles, created inside the `authentication` scope and carrying " +
			"an aes-password bind credential (semp/all-cfg-tps-org.cli:494). A `create " +
			"ldap-profile` for a name the target already has is rejected, so " +
			"ClearExistingNested emits `no ldap-profile <name>` first -- but only for " +
			"names the target carries AND this artifact re-creates, since a removal for " +
			"an absent profile is itself an error.",
	},
	{
		Section:     "Create Radius Profile",
		Disposition: Apply,
		Reason: "Named object, shut down in this capture " +
			"(semp/all.cli:325).",
	},
	{
		Section:     "Create Domain Certificate Authority",
		Disposition: Apply,
		Reason: "Domain CAs, created inside the `ssl` scope (semp/all.cli:334-337). A " +
			"`create domain-certificate-authority` for a name the target already has is " +
			"rejected, so ClearExistingNested emits `no domain-certificate-authority " +
			"<name>` first, for names the target carries AND this artifact re-creates. " +
			"Overlaps `broker configure domain-certs`, which also owns this surface.",
	},

	// Same nested shape as the domain CAs above: the section opens `authentication`
	// and creates inside it, so the Block's Kind is the PARENT and nothing keyed on
	// Kind reaches the object.
	{
		Section:     "Create Client Certificate Authority",
		Disposition: Apply,
		Reason: "Client CA trust material -- certificate, CRL and OCSP revocation " +
			"settings -- created inside the `authentication` scope " +
			"(semp/all-cfg-tps-org.cli:410-500). A `create` for a name the target " +
			"already has is rejected, so ClearExistingNested emits `no " +
			"client-certificate-authority <name>` first, but ONLY for names the target " +
			"carries AND this artifact re-creates: trust material the target's own " +
			"operator added is not this import's to remove.",
	},

	// semp/all.cli:340-343 -- the HA pre-shared key.
	{
		Section:     "Create Redundancy PSK",
		Disposition: Skip,
		Reason: "The HA pre-shared key (semp/all.cli:342); the Solace operator ignores " +
			"preSharedAuthKeySecret updates once a group exists, so this breaks the " +
			"group with no automatic repair.",
	},

	{
		Section:     "Configure Web Manager",
		Disposition: Apply,
		Reason: "Wizard/customization flags and the HTTP redirect port " +
			"(semp/all.cli:346-348).",
	},
	{
		Section:     "Create Authentication",
		Disposition: Apply,
		Reason: "Sets default access levels and the basic auth type " +
			"(semp/all.cli:365, 371); a wrong value can lock an operator out of the " +
			"broker, so review it before applying.",
	},
	{
		Section:     "Configure Management Message Vpn",
		Disposition: Apply,
		Reason:      "`no management-message-vpn` (semp/all.cli:377); cheap and reversible.",
	},
}

// RuleFor returns the classification for one captured section, matched against
// Block.Section. Equality is tried first, over every rule, so a prefix
// candidate can never shadow another rule's exact section name; only when no
// rule matches exactly is section checked against the prefixSections entries.
func RuleFor(section string) (SectionRule, bool) {
	for _, r := range sectionRules {
		if r.Section == section {
			return r, true
		}
	}
	for _, r := range sectionRules {
		if isPrefixSection(r.Section) && strings.HasPrefix(section, r.Section) {
			return r, true
		}
	}
	return SectionRule{}, false
}

// isPrefixSection reports whether section is one of prefixSections -- i.e.
// whether RuleFor should match it by prefix rather than by equality.
func isPrefixSection(section string) bool {
	for _, p := range prefixSections {
		if p == section {
			return true
		}
	}
	return false
}

// Rules returns every classified section, in the stable order sectionRules
// declares them, for the docs generator that produces docs/import.md.
func Rules() []SectionRule {
	return append([]SectionRule(nil), sectionRules...)
}

package broker

import (
	"fmt"
	"sort"
	"strings"
)

// The functions here are pure: each returns the exact Solace CLI script body a
// config/verify op uploads and executes. Keeping them pure (no transport, no
// I/O) makes them golden-testable in scripts_test.go, the same discipline as
// internal/render. They are faithful ports of the heredocs in the numbered bash
// scripts; comments cite the source line ranges.

// cliHome opens every generated script. `home` puts the session at a known level so the
// script does not depend on where a previous command left it, and `no paging` stops the
// broker paginating its output.
//
// Paging is not cosmetic here. A paginated report re-prints its column header and rule
// partway down, and this tool PARSES those reports: the repeated rule reads as a data row
// whose every field is a dash, which is how `show message-vpn * replication` grew a
// phantom VPN named after its own separator. Turning paging off at the source is the fix;
// the parsers' tolerance of a repeat is the backstop, not the other way round.
//
// It is one constant so the pair cannot drift apart, and so a new script cannot be
// written with `home` and without `no paging` -- which is how the scripts that needed
// this fix came to be missing it (operator, 2026-09-13).
const cliHome = "home\nno paging\n"

// showRedundancyScript is the one-line probe used by leader/redundancy polling
// (050 line 34, 061 line 25).
func showRedundancyScript() string { return "show redundancy\n" }

// showRedundancyDetailScript is the timeout diagnostic dumped by 050 (lines 66-67).
func showRedundancyDetailScript() string { return "no paging\nshow redundancy detail\n" }

// assertLeaderScript restores the Primary as config-sync leader for the router
// and all VPNs (050 lines 38-45).
func assertLeaderScript() string {
	return cliHome + `enable
admin
config-sync assert-leader router
config-sync assert-leader message-vpn *
show config-sync database
`
}

// revertActivityScript reverts activity back to the local node (050 lines 23-27).
func revertActivityScript() string {
	return cliHome + "enable\nadmin\nredundancy revert-activity\n"
}

// releaseActivityScript releases activity from the Primary (061 lines 30-34).
func releaseActivityScript() string {
	return cliHome + "enable\nconfigure\nredundancy release-activity\n"
}

// noReleaseActivityScript un-releases the Primary (061 lines 38-42).
func noReleaseActivityScript() string {
	return cliHome + "enable\nconfigure\nno redundancy release-activity\n"
}

// revertActivityConfigureScript reverts activity from the Backup during a
// redundancy test (061 lines 46-50). The trailing space after the command is
// preserved from the source script.
func revertActivityConfigureScript() string {
	return cliHome + "enable\nadmin\nredundancy revert-activity \n"
}

// serverCertScript applies the uploaded TLS server certificate (051 lines 40-43).
// dt is the date stamp (YYYY-MM-DD) that names the uploaded tls-<dt>.crt.key file.
func serverCertScript(dt string) string {
	return cliHome + fmt.Sprintf("enable\nconfigure\nssl server-certificate %s\nshow ssl server-certificate detail\n", serverCertFile(dt))
}

// serverCertFile is the in-broker filename the concatenated key+cert+CAs are
// uploaded as, and the name the CLI loads (051 lines 42, 51).
func serverCertFile(dt string) string { return "tls-" + dt + ".crt.key" }

// removeServerCertScript removes the TLS server certificate the broker presents.
// The form is the operator's, confirmed on a live broker:
//
//	home
//	enable
//	configure
//	no ssl server-certificate
//
// (cliHome adds `no paging` after that `home`, as it does for every script here.)
//
// Emitted verbatim, and it is the one script here that ends with NO confirming `show`.
// Every other one does, because a transcript that proves what changed is worth the extra
// line -- but `show ssl server-certificate detail` against a broker that now has no
// certificate is unverified, and a non-zero result there would report failure for a
// removal that already succeeded. The removal's own output is shown instead.
func removeServerCertScript() string {
	return cliHome + "enable\nconfigure\nno ssl server-certificate\n"
}

// domainCertsScript loads each domain certificate authority (052 lines 20-34).
// cas maps CA name -> certificate filename (already uploaded to the certs dir).
// CA names are emitted in sorted order for deterministic output (bash iterated a
// hash in unspecified order).
func domainCertsScript(cas map[string]string) string {
	var b strings.Builder
	b.WriteString(cliHome + "enable\nconfigure\nssl\n")
	for _, ca := range sortedKeys(cas) {
		fmt.Fprintf(&b, "create domain-certificate-authority %s\ncertificate file %s\nexit\n", ca, cas[ca])
	}
	b.WriteString("end\nshow domain-certificate-authority ca-name *\n")
	return b.String()
}

// removeDomainCertsScript deletes each domain certificate authority (150 lines
// 20-31). cas are the CA names to remove, emitted in the order given (the caller
// sorts them for determinism, since bash iterated a hash in unspecified order).
func removeDomainCertsScript(cas []string) string {
	var b strings.Builder
	b.WriteString(cliHome + "enable\nconfigure\n")
	for _, ca := range cas {
		fmt.Fprintf(&b, "no ssl domain-certificate-authority %s\n", ca)
	}
	b.WriteString(cliHome + "show domain-certificate-authority ca-name *\n")
	return b.String()
}

// disableDefaultVPNScript shuts the default message-VPN down, and nothing else.
//
// Deliberately narrower than its 053 ancestor, which also shut down basic and
// client-certificate authentication, all twelve services, and
// `ssl allow-downgrade-to-plain-text`. Shutting the VPN down already stops every
// service it fronts, so those edits bought nothing and made the op unreversible in
// practice: enableDefaultVPNScript would have had to restore plaintext downgrade to
// be a true inverse, which is a security regression, or leave the VPN half-configured,
// which is worse than not offering an enable at all. One line down, one line up.
//
// The default client-username is NOT touched here either -- that is
// disableDefaultUsersScript's job, and it covers every VPN rather than just this one.
func disableDefaultVPNScript() string {
	return cliHome + `enable
configure
message-vpn "default"
  shutdown
  exit
`
}

// enableDefaultVPNScript is disableDefaultVPNScript's exact inverse: it starts the
// default message-VPN back up. Because the disable side only ever shut the VPN down,
// this restores the broker to the state it was in beforehand without re-opening
// anything the operator hardened by hand.
func enableDefaultVPNScript() string {
	return cliHome + `enable
configure
message-vpn "default"
  no shutdown
  exit
`
}

// showVPNScript lists all message-VPNs. 053 (lines 58-61) wraps it in
// home/enable/configure; 054 (line 20) uses the bare form to parse VPN names.
func showVPNScript() string     { return cliHome + "enable\nconfigure\nshow message-vpn *\n" }
func showVPNBareScript() string { return cliHome + "show message-vpn *\n" }

// showRedundancyLocalScript is `show redundancy` with a script preamble, for the
// replication paths that need to know whether the node they reached holds activity.
//
// showRedundancyScript is the bare one-liner the HA polls use, where the session is
// already where it needs to be; this one opens properly because it is uploaded and run as
// its own script.
func showRedundancyLocalScript() string { return cliHome + "show redundancy\n" }

// showReplicationScript reports the broker's replication mate: its virtual-router-name
// and every address it is dialled on. It carries NO per-VPN role -- that is a different
// command (showVPNReplicationScript), confirmed against real output on both platforms.
func showReplicationScript() string { return cliHome + "show replication\n" }

// showVPNReplicationScript reports every message-VPN's replication admin-state and
// config-state as a fixed-width flag table (ParseVPNReplication).
//
// `replication` is a sub-command of `show message-vpn <vpn-name>`, so the plain
// `show message-vpn *` is a different reply that carries neither flag.
func showVPNReplicationScript() string { return cliHome + "show message-vpn * replication\n" }

// setReplicationRoleScript sets one VPN's replication role.
//
// The operand is a BARE keyword -- the grammar is `state {active | standby}` -- even
// though `show current-config` echoes it back quoted. The VPN name is quoted because real
// ones contain spaces, the same reason currentConfigScript quotes it.
//
// This is the only command in the feature that moves a role, and the only one whose
// ordering across two brokers matters; everything that decides WHEN to call it lives in
// the switch plan, not here.
func setReplicationRoleScript(vpn string, role ReplRole) string {
	return cliHome + fmt.Sprintf("enable\nconfigure\nmessage-vpn %q\n  replication\n    state %s\n    exit\n  exit\n",
		vpn, role)
}

// setReplicationEnabledScript turns replication on or off for one VPN. Enabling is
// `no shutdown` and disabling is `shutdown`, inside the VPN's replication node.
func setReplicationEnabledScript(vpn string, enabled bool) string {
	cmd := "shutdown"
	if enabled {
		cmd = "no shutdown"
	}
	return cliHome + fmt.Sprintf("enable\nconfigure\nmessage-vpn %q\n  replication\n    %s\n    exit\n  exit\n",
		vpn, cmd)
}

// currentConfigScript renders the `show current-config` capture that backs
// `broker perform export-config` and, with remove set, the teardown that
// `import-config` applies before rebuilding a VPN.
//
// `no paging` is mandatory and is the difference between the whole configuration
// and the first screen of it -- gatherConfigsScript opens the same way, for the
// same reason. It is a show command, so it needs neither `enable` nor
// `configure`.
//
// The three forms come straight from the broker's own grammar,
// `show current-config [all | message-vpn <name> [remove]]`:
//
//	vpn == ""   -> `show current-config all`, the whole broker including every VPN
//	vpn != ""   -> `show current-config message-vpn "<vpn>"`
//	remove      -> the broker generates the REMOVAL commands for that VPN instead
//
// remove is emitted only with a vpn, because the broker offers it only there --
// there is no whole-broker teardown generator, which is why broker-level import
// is a section classification rather than a delete-and-replace.
//
// The broker also offers a `redact` form, and this tool deliberately does not use
// it: redaction strips exactly the credential material an import has to put back,
// so a redacted capture is an artifact import would refuse. `Capture.Redacted` is
// still read from the header, because a hand-edited file could claim it and import
// should say so rather than apply a credential-free configuration.
//
// The vpn name is quoted because real VPN names contain spaces (a live capture
// carried "A VPN WITH LONG NAME AND SPACES"), and it is checked by
// validVPNName rather than validName -- validName's charset would reject exactly
// those legitimate names.
func currentConfigScript(vpn string, remove bool) string {
	var b strings.Builder
	b.WriteString(cliHome + "show current-config ")
	if vpn == "" {
		b.WriteString("all")
	} else {
		fmt.Fprintf(&b, "message-vpn %q", vpn)
	}
	if remove && vpn != "" {
		b.WriteString(" remove")
	}
	b.WriteString("\n")
	return b.String()
}

// disableDefaultUsersScript shuts down the "default" client-username in each of
// the given VPNs (054 lines 33-42).
func disableDefaultUsersScript(vpns []string) string {
	return defaultUsersScript(vpns, "shutdown")
}

// enableDefaultUsersScript is disableDefaultUsersScript's inverse: it starts the
// "default" client-username back up in each of the given VPNs.
func enableDefaultUsersScript(vpns []string) string {
	return defaultUsersScript(vpns, "no shutdown")
}

// defaultUsersScript is the shape both default-user scripts share. The only
// difference between them is the one command applied per VPN, so it lives in one
// place: a divergence in the surrounding home/enable/configure preamble or in the
// closing `show` would mean the enable and disable paths reported on different things.
func defaultUsersScript(vpns []string, command string) string {
	var b strings.Builder
	b.WriteString(cliHome + "enable\nconfigure\n")
	for _, vpn := range vpns {
		fmt.Fprintf(&b, "client-username default message-vpn %q\n%s\nexit\n", vpn, command)
	}
	b.WriteString("end\nshow client-username default message-vpn *\n")
	return b.String()
}

// productKeysScript applies each product key (057 lines 24-29).
func productKeysScript(keys []string) string { return productKeyScript(keys, "product-key") }

// removeProductKeysScript revokes each product key. `no product-key <key>` is the exact
// inverse of the apply form, confirmed on a live broker, which is why the two are one
// function: a divergence in the preamble would mean the apply and remove paths ran in
// different CLI contexts, and `product-key` outside `admin` is not a command.
func removeProductKeysScript(keys []string) string { return productKeyScript(keys, "no product-key") }

// productKeyScript is the shape both share. The preamble is the operator's own,
// confirmed on a live broker:
//
//	home
//	enable
//	admin
//	[no ]product-key <key>
//
// `admin` rather than `configure`: a product key is an admin-level action, and the same
// line under `configure` is not a command. `home` leads, so the script does not depend on
// where a session happened to be -- it was missing here and is not optional in the
// confirmed form.
func productKeyScript(keys []string, command string) string {
	var b strings.Builder
	b.WriteString(cliHome + "enable\nadmin\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "%s %s\n", command, k)
	}
	b.WriteString("show product-key\n")
	return b.String()
}

// parseVPNNames extracts message-VPN names from `show message-vpn *` output,
// porting the parser in 054 (lines 24-31): skip until a 30-dash separator, then
// take the first token of each subsequent non-comment line's first 32 columns.
func parseVPNNames(output string) []string {
	var vpns []string
	parsing := false
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimRight(raw, "\r")
		switch {
		case strings.HasPrefix(line, strings.Repeat("-", 30)):
			parsing = true
		case parsing && !strings.HasPrefix(line, "#"):
			col := line
			if len(col) > 32 {
				col = col[:32]
			}
			vpns = append(vpns, strings.Fields(col)...)
		}
	}
	return vpns
}

// gatherConfigsScript is the ~110-command show + gather-diagnostics run collected
// by 069 (lines 38-161). days sets days-of-history for gather-diagnostics.
func gatherConfigsScript(days int) string {
	var b strings.Builder
	b.WriteString(cliHome + "\n! some commands for specific to appliance vs software\n\n")
	for _, cmd := range gatherShowCommands {
		fmt.Fprintf(&b, "show %s > configs/cliout/show-%s.out\n", cmd.args, cmd.out)
	}
	fmt.Fprintf(&b, "\n! gather diagnostics '%d' days\nend\n"+cliHome+"enable\nadmin\ngather-diagnostics days-of-history '%d' no-encrypt\n", days, days)
	return b.String()
}

// zipConfigsScript is the in-broker helper that zips the collected show output
// (069 lines 163-169).
func zipConfigsScript() string {
	return `#!/bin/bash

cd /usr/sw/jail
rm -f gather-configs.zip
mv configs/cliout cli-out
zip gather-configs.zip -q -r cli-out/*
`
}

// showCmd pairs a `show` argument list with the output-file stem it is written to.
type showCmd struct{ args, out string }

// gatherShowCommands is the ordered show-command table from 069 (lines 43-154).
var gatherShowCommands = []showCmd{
	{"acl-profile *", "aclprofiles"},
	{"acl-profile * detail", "aclprofiles-detail"},
	{"alarm", "alarm"},
	{"authentication", "auth"},
	{"authentication access-level", "auth-access-level"},
	{"authentication access-level detail", "auth-access-level-detail"},
	{"backup", "backup"},
	{"bridge *", "bridges"},
	{"bridge * detail", "bridges-detail"},
	{"bridge * stats", "bridge-stats"},
	{"bridge * stats queues", "bridge-stats-queues"},
	{"cache-cluster * detail", "cachecluster"},
	{"cache-instance * detail", "cacheinstance"},
	{"client *", "clients"},
	{"client * detail", "clients-detail"},
	{"client-certificate-authority ca-name * cert", "client-cert-auth-cert"},
	{"client-certificate-authority ca-name * detail", "client-cert-auth-detail"},
	{"client-profile *", "clientprofile"},
	{"client-profile * detail", "clientprofile-detail"},
	{"client-username *", "client-username"},
	{"client-username * detail", "client-username-detail"},
	{"clock detail", "clock-detail"},
	{"cluster *", "cluster"},
	{"cluster * detail", "cluster-detail"},
	{"cluster * link * detail", "cluster-link-detail"},
	{"compression", "compression"},
	{"config-sync", "config-sync"},
	{"config-sync database", "config-sync-database"},
	{"config-sync database detail", "config-sync-database-detail"},
	{"cspf stats", "cspf-stats"},
	{"current-config all", "currentconfig-all"},
	{"current-config message-vpn *", "currentconfig-vpns"},
	{"debug lldp", "debug-lldp"},
	{"disk", "disk"},
	{"disk detail", "disk-detail"},
	{"distributed-cache * detail", "distributedcache"},
	{"dns", "dns"},
	{"domain-certificate-authority ca-name * cert", "domain-cert-auth"},
	{"hardware details", "hardware-details"},
	{"hardware post", "hardware-post"},
	{"hostname", "hostname"},
	{"interface detail", "interface-detail"},
	{"ip vrf management", "vrf-mgmt"},
	{"ip vrf msg-backbone", "vrf-msg-backbone"},
	{"jndi connection-factory * detail", "jndi-cf"},
	{"jndi queue * detail", "jndi-queues"},
	{"jndi summary", "jndi-summary"},
	{"jndi topic * detail", "jndi-topics"},
	{"kerberos keytab", "kerberose-keytab"},
	{"kerberos keytab detail", "kerberose-keytab-details"},
	{"ldap-profile * detail", "ldap-profile-detail"},
	{"logging command", "logging-command"},
	{"logging config", "logging-config"},
	{"logging debug", "logging-debug"},
	{"logging event", "logging-event"},
	{"memory", "memory"},
	{"message-spool detail", "message-spool-detail"},
	{"message-spool message-vpn * detail", "message-spool-vpn-detail"},
	{"message-spool rates", "message-spool-rates"},
	{"message-spool stats", "message-spool-stats"},
	{"message-vpn *", "vpns"},
	{"message-vpn * authorization", "vpn-auth"},
	{"message-vpn * authorization authorization-group *", "vpn-auth-authgroup"},
	{"message-vpn * authorization authorization-group * detail", "vpn-auth-authgroup-detail"},
	{"message-vpn * detail", "vpn-details"},
	{"message-vpn * dynamic-message-routing", "vpn-dmr"},
	{"message-vpn * dynamic-message-routing dmr-bridge *", "vpn-dmr-bridge"},
	{"message-vpn * mqtt", "vpn-mqtt"},
	{"message-vpn * mqtt mqtt-session *", "vpn-mqtt-session"},
	{"message-vpn * mqtt retain cache *", "vpn-mqtt-retain-cache"},
	{"message-vpn * replication", "vpns-replication"},
	{"message-vpn * replication detail", "vpns-repl-detail"},
	{"message-vpn * rest", "vpn-rest"},
	{"message-vpn * rest rest-delivery-point * detail", "vpn-rdp-detail"},
	{"message-vpn * service", "vpn-service"},
	{"mqtt", "mqtt"},
	{"oauth-profile * detail", "oauth-profile-detail"},
	{"product-key", "product-key"},
	{"queue *", "queues"},
	{"queue * detail", "queues-details"},
	{"redundancy", "redundancy"},
	{"redundancy detail", "redundancy-detail"},
	{"redundancy group", "redundancy-group"},
	{"replay-log *", "replay-log"},
	{"replicated-topic *", "replicated-topics"},
	{"replication", "replication"},
	{"router-name", "routername"},
	{"routing", "routing"},
	{"service", "service"},
	{"service semp", "service.semp"},
	{"service virtual-hostname *", "service-virtual-hostname"},
	{"service web-transport", "service-web-transport"},
	{"snmp", "snmp"},
	{"snmp trap *", "snmp-trap"},
	{"ssl allow-tls-version", "ssl-allowed-tls"},
	{"ssl certificate-files", "ssl-certificate-files"},
	{"ssl cipher-suite-list default", "ssl-cipher-default"},
	{"ssl cipher-suite-list management", "ssl-cipher-management"},
	{"ssl cipher-suite-list msg-backbone", "ssl-cipher-msg-backbone"},
	{"ssl cipher-suite-list ssh", "ssl-cipher-ssh"},
	{"ssl server-certificate", "ssl-server-certificate"},
	{"ssl server-certificate detail", "ssl-server-certificate-detail"},
	{"syslog", "syslog"},
	{"system post", "system-post"},
	{"system detail", "system"},
	{"system health", "system-health"},
	{"telemetry", "telemetry"},
	{"topic-endpoint * detail", "topicendpoints"},
	{"username *", "username"},
	{"username * detail", "username-detail"},
	{"version", "version"},
	{"web-manager", "web-manager"},
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

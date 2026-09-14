package broker

import (
	"strings"
	"testing"
)

func TestFixedScripts(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"show-redundancy", showRedundancyScript(), "show redundancy\n"},
		{"show-redundancy-detail", showRedundancyDetailScript(), "no paging\nshow redundancy detail\n"},
		{"revert-activity", revertActivityScript(), "home\nno paging\nenable\nadmin\nredundancy revert-activity\n"},
		{"release-activity", releaseActivityScript(), "home\nno paging\nenable\nconfigure\nredundancy release-activity\n"},
		{"no-release-activity", noReleaseActivityScript(), "home\nno paging\nenable\nconfigure\nno redundancy release-activity\n"},
		{"show-vpn", showVPNScript(), "home\nno paging\nenable\nconfigure\nshow message-vpn *\n"},
		{"show-vpn-bare", showVPNBareScript(), "home\nno paging\nshow message-vpn *\n"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}

// The revert during a redundancy test preserves a trailing space after the
// command (061); the leader-path revert does not (050). Guard both so a stray
// gofmt/whitespace change is caught.
func TestRevertActivityTrailingSpace(t *testing.T) {
	if got := revertActivityConfigureScript(); !strings.Contains(got, "redundancy revert-activity \n") {
		t.Errorf("revertActivityConfigureScript lost its trailing space: %q", got)
	}
	if strings.Contains(revertActivityScript(), "revert-activity \n") {
		t.Error("revertActivityScript should not have a trailing space after the command")
	}
}

func TestAssertLeaderScript(t *testing.T) {
	got := assertLeaderScript()
	for _, want := range []string{
		"config-sync assert-leader router\n",
		"config-sync assert-leader message-vpn *\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("assertLeaderScript missing %q", want)
		}
	}
	if !strings.HasSuffix(got, "show config-sync database\n") {
		t.Errorf("assertLeaderScript should end with the database show: %q", got)
	}
}

func TestServerCertScript(t *testing.T) {
	if got, want := serverCertFile("2026-07-31"), "tls-2026-07-31.crt.key"; got != want {
		t.Fatalf("serverCertFile = %q, want %q", got, want)
	}
	got := serverCertScript("2026-07-31")
	if !strings.HasPrefix(got, cliHome+"enable\nconfigure\n") {
		t.Errorf("serverCertScript prefix: %q", got)
	}
	if !strings.Contains(got, "ssl server-certificate tls-2026-07-31.crt.key\n") {
		t.Errorf("serverCertScript missing the load line: %q", got)
	}
	if !strings.HasSuffix(got, "show ssl server-certificate detail\n") {
		t.Errorf("serverCertScript suffix: %q", got)
	}
}

// TestDomainCertsScriptSorted pins that the caller's order is emitted as given --
// domainCertsScript no longer sorts internally (Ops.DomainCerts sorts before
// calling it) -- and that the `certificate file` operand is the CA NAME itself,
// not a separate filename: the upload destination is certPath(name), so the
// name IS the in-broker filename the operand has to resolve against.
func TestDomainCertsScriptSorted(t *testing.T) {
	got := domainCertsScript([]string{"alpha", "zeta"})
	if !strings.HasPrefix(got, "home\nno paging\nenable\nconfigure\nssl\n") {
		t.Errorf("domainCertsScript prefix: %q", got)
	}
	if !strings.HasSuffix(got, "end\nshow domain-certificate-authority ca-name *\n") {
		t.Errorf("domainCertsScript suffix: %q", got)
	}
	a := strings.Index(got, "create domain-certificate-authority alpha")
	z := strings.Index(got, "create domain-certificate-authority zeta")
	if a < 0 || z < 0 || a > z {
		t.Errorf("domainCertsScript did not preserve the given order (alpha=%d zeta=%d): %q", a, z, got)
	}
	if !strings.Contains(got, "create domain-certificate-authority alpha\ncertificate file alpha\nexit\n") {
		t.Errorf("domainCertsScript missing alpha block, or the operand is not the CA name: %q", got)
	}
}

func TestDisableDefaultUsersScriptQuoting(t *testing.T) {
	got := disableDefaultUsersScript([]string{"default", "my vpn"})
	if !strings.HasPrefix(got, cliHome+"enable\nconfigure\n") {
		t.Errorf("disableDefaultUsersScript prefix: %q", got)
	}
	// VPN names must be shell-safe quoted; a name with a space must stay one token.
	if !strings.Contains(got, `client-username default message-vpn "my vpn"`) {
		t.Errorf("disableDefaultUsersScript did not quote a spaced VPN name: %q", got)
	}
	if !strings.HasSuffix(got, "end\nshow client-username default message-vpn *\n") {
		t.Errorf("disableDefaultUsersScript suffix: %q", got)
	}
}

// TestProductKeysScript pins the apply form confirmed on a live broker, whole. Two
// details it exists to hold: the context is `admin`, not `configure` (the same line under
// configure is not a command), and `home` leads so the script does not depend on where a
// session happened to be. `home` was missing until the form was confirmed, and `no paging`
// beside it was missing until a paginated report was found parsing as data (cliHome).
func TestProductKeysScript(t *testing.T) {
	got := productKeysScript([]string{"KEY-1", "KEY-2"})
	want := cliHome + "enable\nadmin\nproduct-key KEY-1\nproduct-key KEY-2\nshow product-key\n"
	if got != want {
		t.Errorf("productKeysScript = %q, want %q", got, want)
	}
}

// TestRemoveProductKeysScript is the inverse, and TestProductKeyScriptsShareAPreamble is
// what keeps the two from drifting: a removal that reached `no product-key` from a
// different CLI context than the apply reached `product-key` from would fail in a way no
// unit test comparing only its own literal would catch.
func TestRemoveProductKeysScript(t *testing.T) {
	got := removeProductKeysScript([]string{"KEY-1", "KEY-2"})
	want := cliHome + "enable\nadmin\nno product-key KEY-1\nno product-key KEY-2\nshow product-key\n"
	if got != want {
		t.Errorf("removeProductKeysScript = %q, want %q", got, want)
	}
}

func TestProductKeyScriptsShareAPreamble(t *testing.T) {
	const preamble = cliHome + "enable\nadmin\n"
	for name, got := range map[string]string{
		"apply":  productKeysScript([]string{"KEY-1"}),
		"remove": removeProductKeysScript([]string{"KEY-1"}),
	} {
		if !strings.HasPrefix(got, preamble) {
			t.Errorf("%s script does not open with the confirmed preamble: %q", name, got)
		}
		if !strings.HasSuffix(got, "show product-key\n") {
			t.Errorf("%s script does not end by showing what it left behind: %q", name, got)
		}
	}
}

func TestDisableDefaultVPNScript(t *testing.T) {
	got := disableDefaultVPNScript()
	for _, want := range []string{`message-vpn "default"`, "shutdown"} {
		if !strings.Contains(got, want) {
			t.Errorf("disableDefaultVPNScript missing %q", want)
		}
	}
	// The script shuts the VPN down and stops. It used to also shut down the default
	// client-username, all twelve services, and basic/client-certificate auth, and to
	// turn OFF ssl allow-downgrade-to-plain-text.
	//
	// Those are separate concerns and two of them were actively wrong here: the default
	// client-username belongs to `configure default-users`, which shuts it down in EVERY
	// VPN rather than only this one; and disabling the plaintext downgrade is a
	// broker-wide TLS decision that has no business riding along with a VPN being taken
	// out of service -- it also made the operation unreversible, since `--enable` could
	// not know whether the downgrade had been on beforehand.
	for _, unwanted := range []string{
		`client-username "default"`,
		"no ssl allow-downgrade-to-plain-text",
		"service smf plain-text shutdown",
		"authentication",
	} {
		if strings.Contains(got, unwanted) {
			t.Errorf("disableDefaultVPNScript should shut the VPN down and nothing else, but carries %q:\n%s",
				unwanted, got)
		}
	}
}

// TestEnableDefaultVPNScript is the inverse, and the reason the disable script was
// narrowed: an operation that changed six things could not be undone by one that changes
// one. `no shutdown` restores exactly what `shutdown` took away.
func TestEnableDefaultVPNScript(t *testing.T) {
	got := enableDefaultVPNScript()
	for _, want := range []string{`message-vpn "default"`, "no shutdown"} {
		if !strings.Contains(got, want) {
			t.Errorf("enableDefaultVPNScript missing %q", want)
		}
	}
	// It must not re-open the plaintext downgrade: the disable script no longer closes
	// it, so restoring it here would turn an enable into a security change.
	if strings.Contains(got, "ssl allow-downgrade-to-plain-text") {
		t.Errorf("enableDefaultVPNScript must not touch the plaintext downgrade:\n%s", got)
	}
}

func TestParseVPNNames(t *testing.T) {
	// The parser reads the VPN name from the first 32 columns, so each data row
	// must pad the name well past column 32 before the next column begins.
	row := func(name string) string { return name + strings.Repeat(" ", 40-len(name)) + "Yes" }
	out := strings.Join([]string{
		"Flags Legend:",
		"Message VPN" + strings.Repeat(" ", 29) + "Enabled",
		strings.Repeat("-", 40),
		row("default"),
		row("myvpn"),
		"# a comment row is skipped",
		"",
		row("another"),
	}, "\r\n")
	got := parseVPNNames(out)
	want := []string{"default", "myvpn", "another"}
	if len(got) != len(want) {
		t.Fatalf("parseVPNNames = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("parseVPNNames[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParseVPNNamesNoSeparator(t *testing.T) {
	if got := parseVPNNames("no dashes here\njust text\n"); len(got) != 0 {
		t.Errorf("parseVPNNames without separator = %v, want empty", got)
	}
}

func TestGatherConfigsScript(t *testing.T) {
	got := gatherConfigsScript(7)
	if !strings.HasPrefix(got, "home\nno paging\n") {
		t.Errorf("gatherConfigsScript prefix: %q", got[:40])
	}
	if !strings.Contains(got, "show acl-profile * > configs/cliout/show-aclprofiles.out\n") {
		t.Error("gatherConfigsScript missing the first show command")
	}
	if !strings.Contains(got, "gather-diagnostics days-of-history '7' no-encrypt\n") {
		t.Error("gatherConfigsScript missing gather-diagnostics with days substituted")
	}
	if n := strings.Count(got, "> configs/cliout/show-"); n != len(gatherShowCommands) {
		t.Errorf("gatherConfigsScript emitted %d show lines, want %d", n, len(gatherShowCommands))
	}
}

func TestZipConfigsScript(t *testing.T) {
	got := zipConfigsScript()
	if !strings.Contains(got, "zip gather-configs.zip -q -r cli-out/*") {
		t.Errorf("zipConfigsScript missing the zip command: %q", got)
	}
}

// TestSortedKeys is GONE with sortedKeys(map[string]string): domainCertsScript's
// only caller now sorts a []string of CA names directly (Ops.DomainCerts), so
// there is no map left in this package for it to sort keys from.

// TestEveryScriptTurnsPagingOffAfterHome is the invariant behind the parsers.
//
// A paginated report re-prints its column header and rule partway down, and the rule
// then reads as a data row of all-dashes -- which is how `show message-vpn * replication`
// grew a phantom VPN. The parsers tolerate a repeat as a backstop, but the fix is to stop
// the broker paginating, and that means `no paging` immediately after every `home`.
//
// This walks every script generator rather than checking the ones that happened to be
// remembered, because the gap it closes was exactly a set of scripts written with `home`
// and without `no paging` (operator, 2026-09-13).
func TestEveryScriptTurnsPagingOffAfterHome(t *testing.T) {
	scripts := map[string]string{
		"assertLeader":            assertLeaderScript(),
		"revertActivity":          revertActivityScript(),
		"releaseActivity":         releaseActivityScript(),
		"noReleaseActivity":       noReleaseActivityScript(),
		"revertActivityConfigure": revertActivityConfigureScript(),
		"serverCert":              serverCertScript("2026-09-13"),
		"removeServerCert":        removeServerCertScript(),
		"domainCerts":             domainCertsScript([]string{"ca"}),
		"removeDomainCerts":       removeDomainCertsScript([]string{"ca"}),
		"disableDefaultVPN":       disableDefaultVPNScript(),
		"enableDefaultVPN":        enableDefaultVPNScript(),
		"showVPN":                 showVPNScript(),
		"showVPNBare":             showVPNBareScript(),
		"disableDefaultUsers":     disableDefaultUsersScript([]string{"default"}),
		"enableDefaultUsers":      enableDefaultUsersScript([]string{"default"}),
		"productKeys":             productKeysScript([]string{"key"}),
		"removeProductKeys":       removeProductKeysScript([]string{"key"}),
		"gatherConfigs":           gatherConfigsScript(1),
		"currentConfigAll":        currentConfigScript("", false),
		"currentConfigVPN":        currentConfigScript("default", true),
	}
	for name, body := range scripts {
		t.Run(name, func(t *testing.T) {
			lines := strings.Split(body, "\n")
			homes := 0
			for i, l := range lines {
				if strings.TrimSpace(l) != "home" {
					continue
				}
				homes++
				if i+1 >= len(lines) || strings.TrimSpace(lines[i+1]) != "no paging" {
					next := ""
					if i+1 < len(lines) {
						next = lines[i+1]
					}
					t.Errorf("line %d is `home` but the next line is %q, not `no paging`:\n%s",
						i+1, next, body)
				}
			}
			if homes == 0 {
				t.Errorf("no `home` at all; every execution block opens at a known level:\n%s", body)
			}
		})
	}
}

package broker

import (
	"os"
	"strings"
	"testing"
)

// The tests here cover the one thing a `cli -Apes` capture has that a script does
// not: the CLI transcript around it. cli -Apes echoes the login banner, the licence
// text and every prompt+command into stdout ahead of the script's own output, and
// a real export (2026-09-11) showed all of it landing in the artifact. Three
// consequences, each pinned below: the banner is where the target's broker type is
// read from (bannerType); the transcript must be cut off before a capture is parsed
// or replayed (stripTranscript); and an apply script must keep the pragma while
// dropping everything else the preamble picked up (preambleForApply).

// testBannerSoftware and testBannerAppliance are real CLI login banners: the
// software one from the 2026-09-11 export of xps-ps-01, the appliance one from a
// PubSub+ 3560. They carry no credentials.
var (
	testBannerSoftware = []byte("Solace Event Broker Software Enterprise Version 10.26.0.8827\n\n" +
		"This Solace product is proprietary software of\n" +
		"Solace Corporation. By accessing this Solace product\n" +
		"you are agreeing to the license terms and conditions\n" +
		"located at http://www.solace.com/license-software\n\n" +
		"Copyright 2004-2026 Solace Corporation. All rights reserved.\n\n" +
		"Operating Mode: Message Routing Node\n\n\n" +
		"xps-ps-01> home\n")
	testBannerAppliance = []byte("Solace PubSub+ 3560 Version 10.25.0.200\n\n" +
		"The System Software included in the Solace PubSub+ 3560 is proprietary\n" +
		"software of Solace Corporation. By using the Solace PubSub+ 3560\n" +
		"you are agreeing to the license terms and conditions located at\n" +
		"http://www.solace.com/license-hardware\n\n" +
		"Copyright 2004-2026 Solace Corporation. All rights reserved.\n")
)

// TestTranscriptBannerType pins the discriminator the cross-type refusal depends on.
// RISK: this replaced a `show version` scan whose expected schema string a real
// target never printed, so the first live import was refused on a valid target.
// Every arm is real banner text, and the URL-stripped arms prove the product-line
// fallback classifies a reworded banner too.
func TestTranscriptBannerType(t *testing.T) {
	without := func(b []byte, url string) []byte {
		return []byte(strings.ReplaceAll(string(b), url, "(url)"))
	}
	cases := []struct {
		name string
		in   []byte
		want BrokerType
	}{
		{"software banner", testBannerSoftware, BrokerSoftware},
		{"appliance banner", testBannerAppliance, BrokerAppliance},
		{"software, licence URL removed", without(testBannerSoftware, "http://www.solace.com/license-software"), BrokerSoftware},
		{"appliance, licence URL removed", without(testBannerAppliance, "http://www.solace.com/license-hardware"), BrokerAppliance},
		{"no banner", []byte("no banner in this transcript\n"), BrokerUnknown},
		{"empty", nil, BrokerUnknown},
	}
	for _, c := range cases {
		if got := bannerType(c.in); got != c.want {
			t.Errorf("%s: bannerType = %q, want %q", c.name, got, c.want)
		}
	}
}

// TestTranscriptStripCutsAtBegin pins that a capture is cut at the broker's own
// ` ! BEGIN` line, leading space included. RISK: replayed as a script, an echoed
// `xps-ps-01> home` is a rejected command, and cli -Apes exits 0 so the rejection
// is invisible -- this is what would have fed banner text to the CLI as the first
// lines of every teardown.
func TestTranscriptStripCutsAtBegin(t *testing.T) {
	body := " ! BEGIN\n! Command: \"show current-config all\"\n\nhome\nenable\nconfigure\n\n! END\n"
	in := append(append([]byte{}, testBannerSoftware...), []byte("xps-ps-01> show current-config all\n"+body)...)

	got, err := stripTranscript(in)
	if err != nil {
		t.Fatalf("stripTranscript: %v", err)
	}
	if string(got) != body {
		t.Errorf("did not cut at BEGIN:\n%s", got)
	}

	// Already clean -> unchanged, so it is safe on every capture, including one
	// that reached us without a transcript.
	again, err := stripTranscript(got)
	if err != nil || string(again) != body {
		t.Errorf("not idempotent: err=%v got=%q", err, again)
	}

	// No BEGIN -> refused. Whatever this output is, it is not show current-config.
	if _, err := stripTranscript(testBannerSoftware); err == nil {
		t.Error("accepted output with no `! BEGIN` header")
	}
}

// TestTranscriptPreambleForApply pins what an apply script keeps from a captured
// preamble. Two risks pull opposite ways: the banner and echoed prompts must NOT be
// replayed as commands, while `! pragma:interpreter:relaxed-db-restore` MUST be --
// an earlier filter that dropped every `!` line to lose the header lost the pragma
// that makes the interpreter tolerate a replayed configuration.
func TestTranscriptPreambleForApply(t *testing.T) {
	preamble := []string{
		"Solace Event Broker Software Enterprise Version 10.26.0.8827",
		"xps-ps-01> home",
		" ! BEGIN",
		"! Command: \"show current-config all\"",
		"!   SEMP Schema Version: \"broker/10.26.0.8827\"",
		"! pragma:interpreter:relaxed-db-restore",
		"",
		"home",
		"enable",
		"configure",
	}
	// `no paging` is INSERTED after the captured `home`, not taken from the artifact:
	// an apply that paginates stalls waiting for a keypress nobody is there to give, so
	// the setting is this tool's to make rather than whatever the exporting session had.
	want := []string{"! pragma:interpreter:relaxed-db-restore", "home", "no paging", "enable", "configure"}
	got := preambleForApply(preamble)
	if len(got) != len(want) {
		t.Fatalf("preambleForApply = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestTranscriptTeardownFixtureResetsNotDeletes parses a REAL
// `show current-config message-vpn default remove` (testdata/
// currentconfig_remove_default.cli, credential-free) and pins the property the
// teardown design rests on: the broker's generated removal never tries to delete
// an undeletable object. For the default VPN and its default client-profile,
// acl-profile and client-username it emits RESET blocks (`shutdown`, `no alias`,
// `no password`), not a negated opener. That is why this tool carries no
// reserved-name filter and no `default` branch -- and this test is what says so if
// a broker release changes it.
func TestTranscriptTeardownFixtureResetsNotDeletes(t *testing.T) {
	raw, err := os.ReadFile("testdata/currentconfig_remove_default.cli")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	c, err := ParseBlocks(raw)
	if err != nil {
		t.Fatalf("ParseBlocks(teardown): %v", err)
	}
	if len(c.Blocks) == 0 {
		t.Fatal("teardown parsed to no blocks")
	}

	for _, b := range c.Blocks {
		if b.Negate && b.Kind == kindMsgVPN {
			t.Errorf("teardown tries to delete a message-VPN outright: %q", b.Opener())
		}
		if b.Negate && b.Name == "default" {
			t.Errorf("teardown tries to delete an undeletable default object: %q", b.Opener())
		}
	}

	// The reset blocks it DOES emit must be present, or the fixture is not the
	// real shape and the assertions above pass vacuously.
	for _, kind := range []string{kindMsgVPN, "client-username", "client-profile", "acl-profile"} {
		found := false
		for _, b := range c.Blocks {
			if b.Kind == kind && b.Name == "default" && !b.Negate {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("no reset block for %s \"default\" in the teardown fixture", kind)
		}
	}
}

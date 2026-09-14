package config

import (
	"strings"
	"testing"
)

// TestDeriveCAName covers the naming rule for a certificate found by walking
// broker.domainCerts.dirs: the sanitised last directory element, an underscore, and the
// sanitised filename with its extension.
//
// The rule matters more than it looks, because the name it produces is used in THREE
// places at once -- the operand of the broker's `create domain-certificate-authority`,
// the filename the certificate is uploaded under inside the broker, and the key
// `--remove` deletes by. It must therefore be a legal CLI operand, a legal filename, and
// STABLE: a name that changed between two runs would create a second CA beside the first
// rather than update it.
func TestDeriveCAName(t *testing.T) {
	for _, tc := range []struct {
		name, dir, file, want, why string
	}{
		{
			name: "the ordinary case",
			dir:  "/opt/solace/prod-cas", file: "DigiCert Global Root G2.crt",
			want: "prod-cas_DigiCert_Global_Root_G2.crt",
			why:  "both parts sanitised separately, joined with one underscore",
		},
		{
			name: "a run of unsafe characters collapses to one underscore",
			dir:  "/opt/certs", file: "Some  CA   Root.pem",
			want: "certs_Some_CA_Root.pem",
			why: "one-for-one substitution would give a row of underscores that reads as noise " +
				"and makes two visibly different files look alike",
		},
		{
			// A dash is IN the safe set, so it survives -- only the spaces around it
			// collapse. Tempting to normalise the result further, and deliberately not
			// done: collapsing across the safe separators would make `a-b.pem` and
			// `a_b.pem` the same CA name, silently merging two different certificates.
			// A slightly ugly name is a far better outcome than that.
			name: "legal separators are preserved, not tidied away",
			dir:  "/opt/certs", file: "Some CA -- Root.pem",
			want: "certs_Some_CA_--_Root.pem",
			why: "a dash is legal in a CA name, so removing it would be this tool editing a " +
				"name the operator chose",
		},
		{
			name: "two names differing only in their separator stay different",
			dir:  "/opt/certs", file: "a-b.pem",
			want: "certs_a-b.pem",
			why:  "the companion to the case above: a_b.pem must not collapse onto this one",
		},
		{
			name: "the directory keeps two identical filenames apart",
			dir:  "/opt/partner-cas", file: "ca.pem",
			want: "partner-cas_ca.pem",
			why: "ca.pem under two configured dirs is two different certificates, so the " +
				"directory component is what stops them resolving to one CA",
		},
		{
			name: "a trailing separator does not become an empty component",
			dir:  "/opt/solace/prod-cas/", file: "root.crt",
			want: "prod-cas_root.crt",
		},
		{
			name: "windows separators split the same way",
			dir:  `C:\certs\prod-cas`, file: "root.crt",
			want: "prod-cas_root.crt",
			why: "an env file written on Windows and run on Linux is ordinary here, so the " +
				"split cannot depend on the OS running",
		},
		{
			name: "a windows short name keeps its tilde",
			dir:  "/opt/certs", file: "DIGICE~1.CRT",
			want: "certs_DIGICE_1.CRT",
			why:  "a tilde is not in the safe set, so it collapses like any other unsafe character",
		},
		{
			name: "a dotted name keeps every dot",
			dir:  "/opt/certs", file: "root.ca.bundle.pem",
			want: "certs_root.ca.bundle.pem",
			why:  "a period is legal in a CA name, so nothing about a multi-dot name needs rewriting",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DeriveCAName(tc.dir, tc.file)
			if err != nil {
				t.Fatalf("DeriveCAName(%q, %q) error: %v", tc.dir, tc.file, err)
			}
			if got != tc.want {
				t.Errorf("DeriveCAName(%q, %q) = %q, want %q%s",
					tc.dir, tc.file, got, tc.want, whySuffix(tc.why))
			}
			assertUsableCAName(t, got)
		})
	}
}

// TestDeriveCANameShortensTheStemOnly pins which part gives way when a name exceeds the
// broker's limit.
//
// The directory component is what disambiguates two certificates with the same filename,
// and the extension is what makes the name recognisable, so neither may be the part
// dropped. The stem keeps a leading slice of itself rather than becoming a hash, so the
// result still reads as the file it came from.
func TestDeriveCANameShortensTheStemOnly(t *testing.T) {
	dir := "/opt/prod-cas"
	file := strings.Repeat("VeryLongCertificateName", 5) + ".crt"

	got, err := DeriveCAName(dir, file)
	if err != nil {
		t.Fatalf("DeriveCAName: %v", err)
	}
	assertUsableCAName(t, got)
	if len(got) != CANameMax {
		t.Errorf("len = %d, want exactly %d: the stem should be cut to fill the budget, not under it",
			len(got), CANameMax)
	}
	if !strings.HasPrefix(got, "prod-cas_") {
		t.Errorf("%q lost the directory component, which is the part that keeps two same-named "+
			"certificates apart", got)
	}
	if !strings.HasSuffix(got, ".crt") {
		t.Errorf("%q lost its extension, which is what makes the name recognisable as a certificate", got)
	}
	if !strings.Contains(got, "VeryLong") {
		t.Errorf("%q does not read as the file it came from; the stem should keep a leading slice", got)
	}
}

// TestDeriveCANameRefusesWhatItCannotName pins the failures, which exist so nothing is
// silently dropped or silently merged.
func TestDeriveCANameRefusesWhatItCannotName(t *testing.T) {
	for _, tc := range []struct{ name, dir, file, wantSub string }{
		{
			name: "a directory with no usable characters",
			dir:  "/opt/@@@", file: "root.crt",
			wantSub: "has no name this tool can use",
		},
		{
			name: "a file with no usable characters",
			dir:  "/opt/certs", file: "@@@",
			wantSub: "has no name this tool can use",
		},
		{
			// The directory alone eats the whole budget, so there is nothing left to
			// shorten. Refusing beats emitting a name that is all directory.
			name: "a directory name longer than the limit",
			dir:  "/opt/" + strings.Repeat("d", CANameMax+10), file: "root.crt",
			wantSub: "cannot be shortened",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DeriveCAName(tc.dir, tc.file)
			if err == nil {
				t.Fatalf("DeriveCAName(%q, %q) = %q, want an error", tc.dir, tc.file, got)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("err = %v, want it to mention %q", err, tc.wantSub)
			}
			// Every refusal must say what to do about it, not merely that it failed.
			if !strings.Contains(err.Error(), "broker.domainCerts.files") {
				t.Errorf("err = %v, want it to name the escape hatch: an explicitly named "+
					"certificate is how an operator gets past a name this cannot derive", err)
			}
		})
	}
}

// TestMatchesCertExt covers which files in a directory are offered to the broker.
//
// A filter rather than "every file" is the safe default: a certificate directory
// routinely also holds a README, a private key or an editor backup, and uploading one of
// those as a certificate authority fails at best and installs something unintended at
// worst.
func TestMatchesCertExt(t *testing.T) {
	for _, tc := range []struct {
		file string
		exts []string
		want bool
		why  string
	}{
		{file: "root.crt", want: true, why: "a default extension"},
		{file: "root.pem", want: true},
		{file: "root.cer", want: true},
		{file: "README.md", want: false, why: "the reason the filter exists"},
		{file: "root.key", want: false, why: "a private key must never be offered as a CA"},
		{file: "root.crt.bak", want: false, why: "an editor backup is not a certificate"},
		{file: "root.CRT", want: true, why: "a certificate from a Windows tool is as likely to be upper case"},
		{file: "root.der", exts: []string{"der"}, want: true, why: "a caller's own set, dot omitted"},
		{file: "root.der", exts: []string{".der"}, want: true, why: "and with the dot written"},
		{file: "root.crt", exts: []string{"der"}, want: false, why: "a caller's set REPLACES the default"},
		{file: "root.der", exts: []string{" der ", ""}, want: true, why: "whitespace and empties tolerated"},
	} {
		t.Run(tc.file+"/"+strings.Join(tc.exts, ","), func(t *testing.T) {
			if got := MatchesCertExt(tc.file, tc.exts); got != tc.want {
				t.Errorf("MatchesCertExt(%q, %v) = %v, want %v%s",
					tc.file, tc.exts, got, tc.want, whySuffix(tc.why))
			}
		})
	}
}

// assertUsableCAName checks the three properties every derived name must have, since a
// name failing any of them breaks at a different layer: the CLI operand, the in-broker
// filename, or the broker's own length limit.
func assertUsableCAName(t *testing.T, name string) {
	t.Helper()
	if name == "" {
		t.Fatal("empty CA name")
	}
	if len(name) > CANameMax {
		t.Errorf("%q is %d characters, over the broker's limit of %d", name, len(name), CANameMax)
	}
	for _, r := range name {
		if !caNameSafe(r) {
			t.Errorf("%q contains %q, which is not legal in a CA name or as a filename", name, string(r))
		}
	}
}

func whySuffix(why string) string {
	if why == "" {
		return ""
	}
	return " -- " + why
}

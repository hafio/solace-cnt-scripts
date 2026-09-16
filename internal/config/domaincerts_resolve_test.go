package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeDirEntry is a minimal os.DirEntry so ResolveDomainCerts can be tested
// without any directory existing on the machine running the test.
type fakeDirEntry struct {
	name  string
	isDir bool
}

func (e fakeDirEntry) Name() string               { return e.name }
func (e fakeDirEntry) IsDir() bool                { return e.isDir }
func (e fakeDirEntry) Type() os.FileMode          { return 0 }
func (e fakeDirEntry) Info() (os.FileInfo, error) { return nil, fmt.Errorf("not implemented") }

func fakeFile(name string) fakeDirEntry { return fakeDirEntry{name: name} }
func fakeDir(name string) fakeDirEntry  { return fakeDirEntry{name: name, isDir: true} }

// fakeReader builds a DirReader over an in-memory map, so the resolver's walk
// is exercised with no filesystem at all -- the seam ResolveDomainCerts is
// built around, the same shape as container.Manager.Resolve/Geteuid.
func fakeReader(listings map[string][]os.DirEntry) DirReader {
	return func(dir string) ([]os.DirEntry, error) {
		entries, ok := listings[dir]
		if !ok {
			return nil, fmt.Errorf("no such directory: %s", dir)
		}
		return entries, nil
	}
}

// namesOf collects the resolved names in order, for a short assertion.
func namesOf(certs []DomainCert) []string {
	out := make([]string, len(certs))
	for i, c := range certs {
		out[i] = c.Name
	}
	return out
}

func TestResolveDomainCertsWalksOneDirDefaultExtensions(t *testing.T) {
	reader := fakeReader(map[string][]os.DirEntry{
		"/opt/prod-cas": {
			fakeFile("root.crt"),
			fakeFile("chain.PEM"),   // case-insensitive
			fakeFile("readme.txt"),  // not a cert extension: skipped
			fakeFile("key.pem.bak"), // not a cert extension: skipped
			fakeDir("subdir"),       // rule B: never walked
		},
	})
	dc := DomainCerts{Dirs: []CertDir{{Path: "/opt/prod-cas"}}}
	certs, err := ResolveDomainCerts(dc, reader)
	if err != nil {
		t.Fatalf("ResolveDomainCerts: %v", err)
	}
	want := []string{"prod-cas_chain.PEM", "prod-cas_root.crt"} // sorted by filename
	if got := namesOf(certs); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("names = %v, want %v", got, want)
	}
	dir := filepath.Join("/opt/prod-cas") + string(filepath.Separator)
	for _, c := range certs {
		if !strings.HasPrefix(c.Path, dir) {
			t.Errorf("cert %q has path %q, want it joined under %q", c.Name, c.Path, dir)
		}
	}
}

func TestResolveDomainCertsFileExtOverridesDefault(t *testing.T) {
	reader := fakeReader(map[string][]os.DirEntry{
		"/opt/der-cas": {fakeFile("root.der"), fakeFile("root.crt")},
	})
	dc := DomainCerts{Dirs: []CertDir{{Path: "/opt/der-cas", FileExt: "der"}}}
	certs, err := ResolveDomainCerts(dc, reader)
	if err != nil {
		t.Fatalf("ResolveDomainCerts: %v", err)
	}
	if got := namesOf(certs); len(got) != 1 || got[0] != "der-cas_root.der" {
		t.Errorf("names = %v, want exactly [der-cas_root.der] -- fileExt replaces the default, it does not add to it", got)
	}
}

func TestResolveDomainCertsUnreadableDirFailsLoudBeforeUpload(t *testing.T) {
	dc := DomainCerts{Dirs: []CertDir{{Path: "/opt/missing"}}}
	_, err := ResolveDomainCerts(dc, fakeReader(nil))
	if err == nil {
		t.Fatal("an unreadable dir must fail, not be skipped")
	}
	if !strings.Contains(err.Error(), "broker.domainCerts.dirs[0]") || !strings.Contains(err.Error(), "/opt/missing") {
		t.Errorf("error %q must name the field and the dir", err)
	}
}

func TestResolveDomainCertsFilesTakeFullPathVerbatim(t *testing.T) {
	dc := DomainCerts{Files: map[string]string{"my-ca": "/opt/partner-cas/some-ca.pem"}}
	certs, err := ResolveDomainCerts(dc, fakeReader(nil))
	if err != nil {
		t.Fatalf("ResolveDomainCerts: %v", err)
	}
	if len(certs) != 1 || certs[0].Name != "my-ca" || certs[0].Path != "/opt/partner-cas/some-ca.pem" {
		t.Errorf("certs = %+v, want the files entry carried through unchanged", certs)
	}
}

func TestResolveDomainCertsEmptyIsANoOp(t *testing.T) {
	certs, err := ResolveDomainCerts(DomainCerts{}, fakeReader(nil))
	if err != nil || len(certs) != 0 {
		t.Errorf("ResolveDomainCerts(empty) = (%v, %v), want (empty, nil)", certs, err)
	}
}

// TestResolveDomainCertsDuplicateAcrossDirs is the likely real-world collision:
// DeriveCAName uses only the directory's LAST element, so two dirs that end in
// the same segment collide on every file they share.
func TestResolveDomainCertsDuplicateAcrossDirs(t *testing.T) {
	reader := fakeReader(map[string][]os.DirEntry{
		"/site-a/prod-cas": {fakeFile("ca.pem")},
		"/site-b/prod-cas": {fakeFile("ca.pem")},
	})
	dc := DomainCerts{Dirs: []CertDir{{Path: "/site-a/prod-cas"}, {Path: "/site-b/prod-cas"}}}
	_, err := ResolveDomainCerts(dc, reader)
	if err == nil {
		t.Fatal("two dirs sharing a last element must collide on a shared filename")
	}
	// Two things make the expected text less obvious than it looks, and both are
	// about running the same suite on Windows and on Linux.
	//
	// filepath.Join, not a literal forward-slash string: the resolver joins the dir
	// and filename with filepath.Join, whose separator is OS-native.
	//
	// And %q, because that is how this repo renders a path in an error everywhere
	// (see `read tls CA %q` in internal/broker/config_ops.go). On Windows that
	// DOUBLES each backslash in the rendered message, so a raw substring built from
	// filepath.Join would never match there -- the assertion has to quote the same
	// way the message does rather than compare against the bare path.
	for _, want := range []string{
		fmt.Sprintf("%q", filepath.Join("/site-a/prod-cas", "ca.pem")),
		fmt.Sprintf("%q", filepath.Join("/site-b/prod-cas", "ca.pem")),
		"prod-cas_ca.pem",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q must name %s", err, want)
		}
	}
}

// TestResolveDomainCertsDuplicateAgainstExplicitFile covers the third collision
// shape: a derived name colliding with an operator-chosen broker.domainCerts.files
// key.
func TestResolveDomainCertsDuplicateAgainstExplicitFile(t *testing.T) {
	reader := fakeReader(map[string][]os.DirEntry{
		"/opt/prod-cas": {fakeFile("ca.pem")},
	})
	dc := DomainCerts{
		Dirs:  []CertDir{{Path: "/opt/prod-cas"}},
		Files: map[string]string{"prod-cas_ca.pem": "/opt/elsewhere/ca.pem"},
	}
	_, err := ResolveDomainCerts(dc, reader)
	if err == nil {
		t.Fatal("a derived name colliding with an explicit files key must be refused")
	}
	if !strings.Contains(err.Error(), "prod-cas_ca.pem") {
		t.Errorf("error %q must name the colliding CA name", err)
	}
}

// TestResolveDomainCertsRefusesDotDot pins the gap DeriveCAName's own charset does
// not close: '.' is CA-name-safe, so a stem ending in one derives a name
// containing "..", which broker.validName (the in-broker-filename guard) refuses
// at upload time. Catching it here means the error names the source path.
func TestResolveDomainCertsRefusesDotDot(t *testing.T) {
	reader := fakeReader(map[string][]os.DirEntry{
		"/opt/prod-cas": {fakeFile("ca..pem")},
	})
	dc := DomainCerts{Dirs: []CertDir{{Path: "/opt/prod-cas"}}}
	_, err := ResolveDomainCerts(dc, reader)
	if err == nil {
		t.Fatal("a derived name containing \"..\" must be refused before upload")
	}
	if !strings.Contains(err.Error(), "..") {
		t.Errorf("error %q must mention the offending sequence", err)
	}
}

func TestValidExplicitCANameRejects(t *testing.T) {
	cases := []struct {
		name, want string
	}{
		{"", "must not be empty"},
		{"bad name", "letters, digits, underscore, dash and period"},
		{"bad/name", "letters, digits, underscore, dash and period"},
		{"a..b", "must not contain"},
		{strings.Repeat("a", CANameMax+1), fmt.Sprintf("at most %d", CANameMax)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validExplicitCAName(tc.name)
			if err == nil {
				t.Fatalf("validExplicitCAName(%q) accepted, want refused", tc.name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q must contain %q", err, tc.want)
			}
		})
	}
}

func TestValidExplicitCANameAccepts(t *testing.T) {
	for _, name := range []string{"my-ca", "my.ca_1", strings.Repeat("a", CANameMax)} {
		if err := validExplicitCAName(name); err != nil {
			t.Errorf("validExplicitCAName(%q) = %v, want accepted", name, err)
		}
	}
}

func TestSplitFileExt(t *testing.T) {
	if got := splitFileExt(""); got != nil {
		t.Errorf("splitFileExt(\"\") = %v, want nil (so MatchesCertExt substitutes the default)", got)
	}
	got := splitFileExt("cer,crt, pem")
	want := []string{"cer", "crt", " pem"}
	if len(got) != len(want) {
		t.Fatalf("splitFileExt = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("splitFileExt[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestDefaultDirReaderReadsARealDirectory is the one test that touches a real
// filesystem, and it creates its own temp directory rather than depending on
// one already existing on the machine running the test.
func TestDefaultDirReaderReadsARealDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ca.pem"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write fixture file: %v", err)
	}
	entries, err := DefaultDirReader(dir)
	if err != nil {
		t.Fatalf("DefaultDirReader(%q) = %v", dir, err)
	}
	if len(entries) != 1 || entries[0].Name() != "ca.pem" {
		t.Errorf("entries = %v, want exactly [ca.pem]", entries)
	}
}

// --- CertDir.UnmarshalYAML ----------------------------------------------------

func TestCertDirUnmarshalScalar(t *testing.T) {
	var c Config
	if err := decodeStrict("broker:\n  domainCerts:\n    dirs:\n      - /opt/prod-cas\n", &c); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(c.Broker.DomainCerts.Dirs) != 1 || c.Broker.DomainCerts.Dirs[0].Path != "/opt/prod-cas" ||
		c.Broker.DomainCerts.Dirs[0].FileExt != "" {
		t.Errorf("Dirs = %+v", c.Broker.DomainCerts.Dirs)
	}
}

func TestCertDirUnmarshalMapping(t *testing.T) {
	var c Config
	doc := "broker:\n  domainCerts:\n    dirs:\n      - path: /opt/partner-cas\n        fileExt: cer,crt\n"
	if err := decodeStrict(doc, &c); err != nil {
		t.Fatalf("decode: %v", err)
	}
	d := c.Broker.DomainCerts.Dirs[0]
	if d.Path != "/opt/partner-cas" || d.FileExt != "cer,crt" {
		t.Errorf("CertDir = %+v", d)
	}
}

func TestCertDirUnmarshalMappingNeedsPath(t *testing.T) {
	var c Config
	doc := "broker:\n  domainCerts:\n    dirs:\n      - fileExt: cer\n"
	err := decodeStrict(doc, &c)
	if err == nil {
		t.Fatal("a dirs mapping entry with no path must fail to decode")
	}
	if !strings.Contains(err.Error(), "needs a path") {
		t.Errorf("error %q must say a path is required", err)
	}
}

// TestCertDirUnmarshalRejectsUnknownKey is the strictness Command's own
// UnmarshalYAML does not need: a custom UnmarshalYAML takes over decoding for
// the node, so yaml.v3's KnownFields(true) never reaches inside it, and a
// typo'd `fileExts:` would otherwise decode clean and silently keep the
// default extensions instead of failing loud the way every other typo here does.
func TestCertDirUnmarshalRejectsUnknownKey(t *testing.T) {
	var c Config
	doc := "broker:\n  domainCerts:\n    dirs:\n      - path: /opt/cas\n        fileExts: cer\n"
	err := decodeStrict(doc, &c)
	if err == nil {
		t.Fatal("a dirs mapping entry with an unknown key must fail to decode")
	}
	if !strings.Contains(err.Error(), "fileExts") {
		t.Errorf("error %q must name the offending key", err)
	}
}

func TestCertDirUnmarshalRejectsOtherKinds(t *testing.T) {
	var c Config
	doc := "broker:\n  domainCerts:\n    dirs:\n      - [nested, list]\n"
	err := decodeStrict(doc, &c)
	if err == nil {
		t.Fatal("a dirs entry that is neither a scalar nor a mapping must fail to decode")
	}
}

// TestDomainCertsNamesAreCheckedAtLoadNotAtUpload is the timing pin, and timing is
// the whole point of it.
//
// ResolveDomainCerts deliberately runs at the OP rather than at load, because it
// READS DIRECTORIES: a certificate directory that does not happen to exist on this
// machine must not fail a `deploy` or a `generate` that never touches domain
// certificates. The cost of that choice was that everything else about the block
// waited for the op too -- so a files key that was never a legal certificate
// authority name survived generate, deploy and verify, and surfaced only at the one
// command that uploads, which is the worst moment to learn the name was illegal.
//
// A key and an empty value are pure string facts about the env file. They need no
// filesystem, so they are checked at load, and this test is what stops them
// drifting back to the op.
func TestDomainCertsNamesAreCheckedAtLoadNotAtUpload(t *testing.T) {
	for _, tc := range []struct {
		name string
		doc  string
		want string
	}{
		{
			"an illegal character in the name",
			"broker:\n  domainCerts:\n    files:\n      \"bad name\": /certs/ca.pem\n",
			"only letters, digits, underscore, dash and period",
		},
		{
			"a name that is too long",
			"broker:\n  domainCerts:\n    files:\n      " + strings.Repeat("a", CANameMax+1) + ": /certs/ca.pem\n",
			"config.CANameMax",
		},
		{
			// Empty is NOT "will be defaulted" here, which is the asymmetry
			// CheckHostPath encodes for every other path field -- it returns nil on
			// an empty value precisely because unset means defaulted. Nothing
			// defaults a files entry, so an empty one can only ever fail at upload.
			"a name with no path behind it",
			"broker:\n  domainCerts:\n    files:\n      my-ca: \"\"\n",
			"has no path",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var c Config
			if err := decodeStrict(tc.doc, &c); err != nil {
				t.Fatalf("the document must decode -- the point is that VALIDATION catches it: %v", err)
			}
			c.ApplyDefaults(K8s)
			err := c.Validate(K8s)
			if err == nil {
				t.Fatal("a files entry this broken must fail at LOAD, not wait for the upload")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

// TestDomainCertsRenameErrorWinsOverTheNewBlock pins the ORDER of the two checks.
//
// A file still carrying the retired broker.domainCerts.folder has not written the
// new block yet, so it must hear that its key was renamed and reshaped -- not a
// complaint about the shape of something it never wrote.
func TestDomainCertsRenameErrorWinsOverTheNewBlock(t *testing.T) {
	var c Config
	doc := "broker:\n  domainCerts:\n    folder: certs\n    files:\n      \"bad name\": /certs/ca.pem\n"
	if err := decodeStrict(doc, &c); err != nil {
		t.Fatalf("decode: %v", err)
	}
	c.ApplyDefaults(K8s)
	err := c.Validate(K8s)
	if err == nil {
		t.Fatal("this document is invalid twice over and must fail")
	}
	if !strings.Contains(err.Error(), "was renamed to broker.domainCerts.dirs") {
		t.Errorf("error = %v, want the RENAME error: an operator who has not migrated the key yet cannot "+
			"act on a complaint about the block they have not written", err)
	}
}

// TestResolveDomainCertsSameFileReachedTwice covers the case the general
// duplicate-name message describes badly: when one file is reached by two entries,
// prev and path are the SAME string, so that message named one path twice and sent
// the operator hunting for a second certificate that does not exist.
func TestResolveDomainCertsSameFileReachedTwice(t *testing.T) {
	read := fakeReader(map[string][]os.DirEntry{
		"/certs": {fakeFile("ca.pem")},
	})
	dc := DomainCerts{Dirs: []CertDir{{Path: "/certs"}, {Path: "/certs"}}}
	_, err := ResolveDomainCerts(dc, read)
	if err == nil {
		t.Fatal("the same file reached twice must fail: it would be uploaded twice under one name")
	}
	for _, want := range []string{"is loaded twice", "Remove the duplicate dirs entry"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q must contain %q -- it must name the duplicated ENTRY, not report a "+
				"collision between two certificates that do not exist", err, want)
		}
	}
	// And the general message must NOT be what fired: it would print the one path
	// twice, which describes nothing.
	if strings.Contains(err.Error(), "two certificates resolve to") {
		t.Errorf("error %q used the two-certificate message for one file reached twice", err)
	}
}

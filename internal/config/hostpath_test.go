package config

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestHostPathCharsetIsDerivedFromTheTokenCharset is the reason hostPathUnsafe is
// built by strings.Map rather than typed out: two hand-written charsets drift, and
// the drift is invisible until an operator hits the one character that is in one
// list and not the other. This asserts the relationship, not the contents -- adding
// a character to unsafeTokenChars must widen the path set too, automatically.
func TestHostPathCharsetIsDerivedFromTheTokenCharset(t *testing.T) {
	for _, r := range hostPathAdmits {
		if strings.ContainsRune(hostPathUnsafe, r) {
			t.Errorf("hostPathUnsafe must not contain %q: a host path is not argv, and this character is "+
				"legitimate in one (Windows separator, 8.3 short name)", string(r))
		}
		// Each admitted character must actually be in the set it is removed FROM,
		// or the removal is decoration. A colon belongs in neither list: the token
		// charset never refused one, so "admitting" it here would assert nothing.
		if !strings.ContainsRune(unsafeTokenChars, r) {
			t.Errorf("hostPathAdmits lists %q, but unsafeTokenChars does not refuse it, so removing it is a "+
				"no-op that makes the derivation look stricter than it is", string(r))
		}
	}
	// Everything the token charset refuses and the path charset is not explicitly
	// forgiven must still be refused. This is the half that catches a new character
	// added to unsafeTokenChars alone.
	for _, r := range unsafeTokenChars {
		if strings.ContainsRune(hostPathAdmits, r) {
			continue
		}
		if !strings.ContainsRune(hostPathUnsafe, r) {
			t.Errorf("unsafeTokenChars contains %q but hostPathUnsafe does not, and it is not one of the "+
				"deliberately admitted characters: the two sets have drifted", string(r))
		}
	}
	if len(hostPathUnsafe) != len(unsafeTokenChars)-len(hostPathAdmits) {
		t.Errorf("hostPathUnsafe has %d characters, want %d (the token set minus exactly %q)",
			len(hostPathUnsafe), len(unsafeTokenChars)-len(hostPathAdmits), hostPathAdmits)
	}
	// A colon passes both gates, and that is deliberate rather than an oversight:
	// it is legitimate in a command token (`--server=https://host:6443`) and in a
	// Windows path (`C:\certs`). What a colon does to a mount LINE is a rendering
	// problem, handled where the line is written, not here.
	if strings.ContainsRune(hostPathUnsafe, ':') || strings.ContainsRune(unsafeTokenChars, ':') {
		t.Error("neither charset should refuse ':': it is legitimate in a flag value and in a Windows path")
	}
	// The dollar sign is the one worth naming individually: compose interpolates it
	// across the whole document, including a volume source.
	if !strings.ContainsRune(hostPathUnsafe, '$') {
		t.Error("hostPathUnsafe must refuse '$': compose interpolates it in a volume source")
	}
}

// TestCheckHostPathAccepts covers the shapes an operator legitimately writes,
// including the two Windows forms that a command token may not carry.
func TestCheckHostPathAccepts(t *testing.T) {
	for _, p := range []string{
		"",                        // empty is a per-field question, not this one
		"certs/tls.crt",           // relative, the common case
		"/opt/solace/data",        // absolute posix
		`C:\certs\tls.crt`,        // absolute windows, backslash and colon
		"C:/certs/tls.crt",        // absolute windows, forward slashes
		`C:\Users\RUNNER~1\certs`, // 8.3 short name: tilde, not leading
		"../shared/certs/tls.crt", // above the env file
		"diag-configs",            // a bare directory name
		"a.b_c-d/e.pem",           // punctuation a name grammar would refuse
	} {
		if err := CheckHostPath("tls.cert", p); err != nil {
			t.Errorf("CheckHostPath(%q) = %v, want accepted", p, err)
		}
	}
}

// TestCheckHostPathRejects covers each refusal with its own reason, so a message
// that stops naming the cause fails here rather than confusing an operator.
func TestCheckHostPathRejects(t *testing.T) {
	cases := []struct {
		path string
		want string // substring the message must carry
	}{
		{"certs/$HOME/tls.crt", "compose file or quadlet unit"},
		{"certs/tls$.crt", "compose file or quadlet unit"},
		{"certs/a;b.crt", "compose file or quadlet unit"},
		{"certs/a|b.crt", "compose file or quadlet unit"},
		{"certs/a*.crt", "compose file or quadlet unit"},
		{"certs/`cmd`.crt", "compose file or quadlet unit"},
		{"my certs/tls.crt", "source:target:options"},
		// U+00A0, a non-breaking space: unicode.IsSpace, but NOT isCtrl (which is
		// r < 0x20) and not a Cf formatting character. It is the case that proves
		// the whitespace branch handles more than the ASCII space. Written as an
		// escape so this source file stays plain ASCII.
		{"certs/\u00a0tls.crt", "source:target:options"},
		// A tab is caught one branch earlier, as a control character, because
		// isCtrl is r < 0x20. Pinned so the branch order is deliberate.
		{"certs/tls\t.crt", "control character"},
		{"certs/tls\n.crt", "control character"},
		{"~/certs/tls.crt", "does not expand"},
		{"~", "does not expand"},
	}
	for _, tc := range cases {
		err := CheckHostPath("tls.cert", tc.path)
		if err == nil {
			t.Errorf("CheckHostPath(%q) = nil, want refused", tc.path)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("CheckHostPath(%q) error %q must explain the refusal (want substring %q)",
				tc.path, err, tc.want)
		}
		if !strings.Contains(err.Error(), "tls.cert") {
			t.Errorf("CheckHostPath(%q) error %q must name the field", tc.path, err)
		}
	}
}

// TestLeadingTildeIsRefusedButAnEmbeddedOneIsNot pins the carve-out on its own,
// because the two cases differ only by the tilde's position and the reasons are
// unrelated: nothing in this tool expands a home directory, while an 8.3 short name
// is a real path a Windows runner hands us for its own temp directory.
func TestLeadingTildeIsRefusedButAnEmbeddedOneIsNot(t *testing.T) {
	if err := CheckHostPath("podman.baseDir", "~/solace"); err == nil {
		t.Error("a leading '~' must be refused: it would be joined onto the env file's directory and " +
			"produce a literal '~' segment, reported later as a missing file")
	}
	if err := CheckHostPath("podman.baseDir", `C:\Users\RUNNER~1\AppData\Local\Temp`); err != nil {
		t.Errorf("an 8.3 short name must be accepted, got %v", err)
	}
}

// TestIsAbsHostPath asks the question for both operating systems at once, which is
// the point: filepath.IsAbs alone answers only for the machine running the test, so
// the same env file would be judged differently depending on where it was loaded.
func TestIsAbsHostPath(t *testing.T) {
	for _, p := range []string{
		"/opt/solace/data",
		`\\server\share\certs`,
		`C:\certs`,
		"C:/certs",
		"c:/certs",
		"Z:/data",
	} {
		if !IsAbsHostPath(p) {
			t.Errorf("IsAbsHostPath(%q) = false, want true regardless of host OS", p)
		}
	}
	for _, p := range []string{
		"",
		"certs/tls.crt",
		"certs",
		"../certs",
		"./certs",
		"1:/certs", // not a drive letter
	} {
		if IsAbsHostPath(p) {
			t.Errorf("IsAbsHostPath(%q) = true, want false", p)
		}
	}
}

// TestBaseNameSplitsOnBothSeparators is the cross-platform agreement test. The
// value it returns becomes an in-broker filename and a Secret data key, so two
// answers for one env file name two different objects rather than differing
// cosmetically.
func TestBaseNameSplitsOnBothSeparators(t *testing.T) {
	cases := map[string]string{
		"certs/tls.crt":       "tls.crt",
		`certs\tls.crt`:       "tls.crt",
		`C:\certs\tls.crt`:    "tls.crt",
		"/opt/certs/ca.pem":   "ca.pem",
		"tls.crt":             "tls.crt",
		"":                    "",
		`mixed/sep\file.p12`:  "file.p12",
		`trailing/backslash\`: "",
	}
	for in, want := range cases {
		if got := BaseName(in); got != want {
			t.Errorf("BaseName(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestHasPathSeparator pins the name-or-path question that `cli --input` and the
// domain-certificate filenames both rest on.
func TestHasPathSeparator(t *testing.T) {
	for _, p := range []string{"a/b", `a\b`, "/abs", `C:\x`, "./x"} {
		if !HasPathSeparator(p) {
			t.Errorf("HasPathSeparator(%q) = false, want true", p)
		}
	}
	for _, p := range []string{"setup.cli", "", "name.with.dots", "name-with-dashes"} {
		if HasPathSeparator(p) {
			t.Errorf("HasPathSeparator(%q) = true, want false", p)
		}
	}
}

// TestValidateHostPathsChecksEveryPlatformsFields walks each platform and asserts
// the gate actually fires for every field it is supposed to cover. It uses a
// character the path charset refuses ('$') so a failure can only come from
// CheckHostPath, and it sets one field at a time so a message naming the wrong
// field is caught rather than being masked by a neighbour.
func TestValidateHostPathsChecksEveryPlatformsFields(t *testing.T) {
	// field name -> the setter that puts a bad value on that field.
	perPlatform := map[Platform]map[string]func(*Config){
		K8s: {
			"tls.cert":                 func(c *Config) { c.TLS.Cert = "certs/$bad.crt" },
			"tls.certKey":              func(c *Config) { c.TLS.CertKey = "certs/$bad.key" },
			"tls.cas[0]":               func(c *Config) { c.TLS.CAs = []string{"certs/$bad.pem"} },
			"broker.cliScriptsFolder":  func(c *Config) { c.Broker.CLIScriptsFolder = "cli$x" },
			"broker.diagDir":           func(c *Config) { c.Broker.DiagDir = "diag$x" },
			"broker.domainCerts.folder": func(c *Config) { c.Broker.DomainCerts.Folder = "certs$x" },
		},
		Docker: {
			"docker.composeFile":       func(c *Config) { c.Docker.ComposeFile = "com$pose.yml" },
			"docker.container.dataDir": func(c *Config) { c.Docker.Container.DataDir = "/opt/sol$ace" },
		},
		Podman: {
			"podman.quadletDir":        func(c *Config) { c.Podman.QuadletDir = "/etc/quad$let" },
			"podman.baseDir":           func(c *Config) { c.Podman.BaseDir = "/opt/base$dir" },
			"podman.container.dataDir": func(c *Config) { c.Podman.Container.DataDir = "/opt/sol$ace" },
		},
	}
	for p, fields := range perPlatform {
		for name, set := range fields {
			c := guardConfig(p)
			set(c)
			err := c.validateHostPaths(p)
			if err == nil {
				t.Errorf("%s: validateHostPaths accepted a '$' in %s, so that field is not gated", p, name)
				continue
			}
			if !strings.Contains(err.Error(), name) {
				t.Errorf("%s: validateHostPaths error %q must name the field %s", p, err, name)
			}
		}
	}
}

// TestValidateHostPathsMatchesTheRebaseList is the drift test between the two
// halves of the host-path story: a field that is rebased must also be gated. The
// reverse is allowed and deliberate -- podman.quadletDir is gated but never
// rebased, because the unit has to live where systemd scans.
func TestValidateHostPathsMatchesTheRebaseList(t *testing.T) {
	for _, p := range Platforms() {
		// Every rebased field, given a value the gate refuses, must make
		// validateHostPaths fail. If a field gains a rebase without a gate entry,
		// the bad value survives into an artifact instead.
		c := guardConfig(p)
		c.baseDir = "" // isolate: this test is about the gate, not the join
		rebased := map[string]func(*Config){
			"tls.cert":                  func(c *Config) { c.TLS.Cert = "a/$b" },
			"tls.certKey":               func(c *Config) { c.TLS.CertKey = "a/$b" },
			"tls.cas[0]":                func(c *Config) { c.TLS.CAs = []string{"a/$b"} },
			"broker.cliScriptsFolder":   func(c *Config) { c.Broker.CLIScriptsFolder = "a/$b" },
			"broker.diagDir":            func(c *Config) { c.Broker.DiagDir = "a/$b" },
			"broker.domainCerts.folder": func(c *Config) { c.Broker.DomainCerts.Folder = "a/$b" },
		}
		if p == Docker {
			rebased["docker.composeFile"] = func(c *Config) { c.Docker.ComposeFile = "a/$b" }
		}
		for name, set := range rebased {
			fresh := guardConfig(p)
			set(fresh)
			if err := fresh.validateHostPaths(p); err == nil {
				t.Errorf("%s: %s is rebased by rebaseHostPaths but not gated by validateHostPaths, so a "+
					"refused character would reach an artifact", p, name)
			}
		}
	}
}

// TestRebaseResolvesAgainstTheEnvFileDirectory covers the join itself: relative
// values move, absolute ones do not, and an empty value stays empty rather than
// becoming the base directory.
func TestRebaseResolvesAgainstTheEnvFileDirectory(t *testing.T) {
	c := guardConfig(Docker)
	c.baseDir = filepath.FromSlash("/srv/solace/env")
	c.TLS.Cert = "certs/tls.crt"
	c.TLS.CertKey = filepath.FromSlash("/abs/tls.key") // already absolute
	c.TLS.CAs = []string{"certs/ca.pem", filepath.FromSlash("/abs/ca2.pem")}
	c.Broker.DiagDir = "diag-configs"
	c.Broker.DomainCerts.Folder = "" // unset stays unset
	c.Docker.ComposeFile = "docker-compose.yml"

	c.rebaseHostPaths()

	want := map[string]string{
		"tls.cert":            filepath.Join("/srv/solace/env", "certs/tls.crt"),
		"tls.certKey":         filepath.FromSlash("/abs/tls.key"),
		"tls.cas[0]":          filepath.Join("/srv/solace/env", "certs/ca.pem"),
		"tls.cas[1]":          filepath.FromSlash("/abs/ca2.pem"),
		"broker.diagDir":      filepath.Join("/srv/solace/env", "diag-configs"),
		"docker.composeFile":  filepath.Join("/srv/solace/env", "docker-compose.yml"),
	}
	got := map[string]string{
		"tls.cert":           c.TLS.Cert,
		"tls.certKey":        c.TLS.CertKey,
		"tls.cas[0]":         c.TLS.CAs[0],
		"tls.cas[1]":         c.TLS.CAs[1],
		"broker.diagDir":     c.Broker.DiagDir,
		"docker.composeFile": c.Docker.ComposeFile,
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("after rebase %s = %q, want %q", k, got[k], w)
		}
	}
	if c.Broker.DomainCerts.Folder != "" {
		t.Errorf("an unset path must stay unset, got %q -- otherwise every omitted field silently becomes "+
			"the env file's own directory", c.Broker.DomainCerts.Folder)
	}
}

// TestRebaseIsANoOpWithoutABaseDir is the property internal/convert and every
// hand-built Config depend on: ApplyDefaults and Validate are called directly, with
// no path to derive a base from, and those paths must behave exactly as before.
func TestRebaseIsANoOpWithoutABaseDir(t *testing.T) {
	c := guardConfig(Docker)
	c.baseDir = ""
	c.TLS.Cert = "certs/tls.crt"
	c.Broker.DiagDir = "diag-configs"
	c.rebaseHostPaths()
	if c.TLS.Cert != "certs/tls.crt" || c.Broker.DiagDir != "diag-configs" {
		t.Errorf("with no base dir the paths must be untouched, got tls.cert=%q diagDir=%q",
			c.TLS.Cert, c.Broker.DiagDir)
	}
	if c.BaseDir() != "" {
		t.Errorf("BaseDir() = %q, want empty", c.BaseDir())
	}
}

// TestQuadletDirAndDataDirAreNotRebased pins the two deliberate exclusions, each
// for its own reason: the unit must live where systemd scans, and the data dir is
// what a recursive delete points at.
func TestQuadletDirAndDataDirAreNotRebased(t *testing.T) {
	c := guardConfig(Podman)
	c.baseDir = filepath.FromSlash("/srv/solace/env")
	c.Podman.QuadletDir = "relative/quadlet"
	c.Podman.Container.DataDir = "relative/data"
	c.Podman.BaseDir = "relative/base"
	c.rebaseHostPaths()
	if c.Podman.QuadletDir != "relative/quadlet" {
		t.Errorf("podman.quadletDir must not be rebased, got %q: systemd only reads its own scan "+
			"directories, so resolving a relative value would invent a location it never reads", c.Podman.QuadletDir)
	}
	if c.Podman.Container.DataDir != "relative/data" {
		t.Errorf("container.dataDir must not be rebased, got %q: it is required absolute instead, because it "+
			"is the target of `remove broker --delete-data`", c.Podman.Container.DataDir)
	}
	if c.Podman.BaseDir != "relative/base" {
		t.Errorf("podman.baseDir must not be rebased, got %q: it is required absolute instead, because it "+
			"receives a file holding a private key and resolving a relative value would invent a location "+
			"for that key the operator never named", c.Podman.BaseDir)
	}
}

// TestPodmanBaseDirIsRequiredAndAbsolute pins both halves of the one mandatory path
// key. It is mandatory rather than defaulted because it receives the
// server-certificate bundle, which contains a PRIVATE KEY, and where that lands on a
// host is the operator's decision. It is required absolute rather than rebased for
// the same reason container.dataDir is: it is a quadlet `Volume=` source, and podman
// reads a relative source as the name of a NAMED VOLUME, which would mount an empty
// volume over the certificate with no error at all.
func TestPodmanBaseDirIsRequiredAndAbsolute(t *testing.T) {
	// Unset is refused, and the message names the key.
	c := guardConfig(Podman)
	c.Podman.BaseDir = ""
	err := c.Validate(Podman)
	if err == nil {
		t.Fatal("an unset podman.baseDir must be refused: it is where the private key lands")
	}
	if !strings.Contains(err.Error(), "podman.baseDir") {
		t.Errorf("error %q must name the key", err)
	}

	// Relative is refused, and the message says why rather than only that it is invalid.
	for _, bad := range []string{"solace", "relative/base", "./base"} {
		c := guardConfig(Podman)
		c.Podman.BaseDir = bad
		err := c.Validate(Podman)
		if err == nil {
			t.Errorf("podman.baseDir %q must be refused", bad)
			continue
		}
		if !strings.Contains(err.Error(), "NAMED VOLUME") {
			t.Errorf("podman.baseDir %q error %q must explain the named-volume trap, which is what makes a "+
				"relative value dangerous rather than merely wrong", bad, err)
		}
	}

	// Absolute is accepted, in both spellings a real env file might use.
	for _, good := range []string{"/opt/solace", "/etc/solace"} {
		c := guardConfig(Podman)
		c.Podman.BaseDir = good
		if err := c.Validate(Podman); err != nil {
			t.Errorf("podman.baseDir %q must be accepted: %v", good, err)
		}
	}

	// Docker has no such key and must not acquire the requirement.
	d := guardConfig(Docker)
	d.Podman.BaseDir = ""
	if err := d.Validate(Docker); err != nil {
		t.Errorf("docker must not require podman.baseDir: a compose file can inline what a quadlet cannot: %v", err)
	}
}

// TestDataDirMustBeAbsolute pins the one host path that is REQUIRED absolute rather
// than resolved, and the message has to explain both halves of why: it is a bind
// mount source (podman reads a bare one as a named volume) and it is what
// `remove broker --delete-data` deletes recursively.
func TestDataDirMustBeAbsolute(t *testing.T) {
	for _, p := range []Platform{Docker, Podman} {
		// The default is absolute, so an untouched fixture must pass.
		if err := guardConfig(p).Validate(p); err != nil {
			t.Errorf("%s: the default dataDir must satisfy the absolute rule: %v", p, err)
		}
		for _, bad := range []string{"data", "relative/data", "./data"} {
			c := guardConfig(p)
			if p == Docker {
				c.Docker.Container.DataDir = bad
			} else {
				c.Podman.Container.DataDir = bad
			}
			err := c.Validate(p)
			if err == nil {
				t.Errorf("%s: dataDir %q must be refused: a relative bind-mount source is a named volume to "+
					"podman, and it is the target of a recursive delete", p, bad)
				continue
			}
			if !strings.Contains(err.Error(), "dataDir") {
				t.Errorf("%s: dataDir %q error %q must name the field", p, bad, err)
			}
			if !strings.Contains(err.Error(), "--delete-data") {
				t.Errorf("%s: dataDir %q error %q must say what makes this dangerous, not just that it is "+
					"invalid -- the recursive delete is the reason it is a refusal and not a rebase", p, bad, err)
			}
		}
	}
}

// TestDomainCertsFolderIsDefaulted covers a promise the sample env file had been
// making without the code keeping it: the file shows `folder: certs` under a header
// saying a commented-out key shows the default that applies when omitted, but nothing
// set it -- so an omitted folder made filepath.Join("", file) collapse to a bare
// filename resolved against whatever directory the command ran from.
func TestDomainCertsFolderIsDefaulted(t *testing.T) {
	for _, p := range Platforms() {
		c := &Config{}
		c.ApplyDefaults(p)
		if got := c.Broker.DomainCerts.Folder; got != "certs" {
			t.Errorf("%s: broker.domainCerts.folder = %q, want %q -- env/sample.yaml documents that default",
				p, got, "certs")
		}
	}
	// An explicit value still wins.
	c := &Config{}
	c.Broker.DomainCerts.Folder = "my-cas"
	c.ApplyDefaults(Docker)
	if c.Broker.DomainCerts.Folder != "my-cas" {
		t.Errorf("an explicit folder must not be overwritten, got %q", c.Broker.DomainCerts.Folder)
	}
}

// TestAdminUserIsFormatCheckedOnContainers is GONE with the admin.user key it pinned.
// It held a configurable admin username to identRE because the value reached a Secret
// key, a compose `target:`, a podman `target=` and the broker setting
// username_<u>_globalaccesslevel. The username is the constant config.AdminUser now, on
// every platform, so none of those positions carries operator input any more.

// TestContainerCertRequiresKey pins a deliberate TIGHTENING: an env file setting
// tls.cert without tls.certKey loads today and will stop loading on docker and
// podman.
//
// The reason it should stop is not tidiness. Until the server certificate moved onto
// a real delivery path, the container renderers mounted tls.cert ALONE and never
// read tls.certKey at all -- so a broker-owned TLS listener could only ever have
// worked if that single file already carried the key. Failing at load names the
// missing field; the previous behaviour produced a running broker whose TLS quietly
// did not work.
//
// Kubernetes now enforces the same pairing, for a different reason and from a different
// check -- see TestTLSCertAndKeyMustBeSetTogetherOnKubernetes. It used to be exempt only
// because applyK8sDefaults filled both fields together, so one could not arrive alone;
// that defaulting is gone.
func TestContainerCertRequiresKey(t *testing.T) {
	for _, p := range []Platform{Docker, Podman} {
		c := guardConfig(p)
		c.TLS.Cert = "certs/tls.crt" // no certKey
		err := c.Validate(p)
		if err == nil {
			t.Errorf("%s: tls.cert without tls.certKey must be refused: the container mounts ONE file "+
				"containing both halves", p)
			continue
		}
		if !strings.Contains(err.Error(), "tls.certKey") {
			t.Errorf("%s: error %q must name the missing field", p, err)
		}

		// And the reverse. Every container site keys on TLS.Cert alone -- the compose
		// secret, the quadlet Volume=, the configsync/matelink pairs -- so a lone certKey
		// used to pass validation and deploy a broker with TLS silently off: the same
		// failure this check exists to end, arrived at from the other side.
		rev := guardConfig(p)
		rev.TLS.CertKey = "certs/tls.key" // no cert
		err = rev.Validate(p)
		if err == nil {
			t.Errorf("%s: tls.certKey without tls.cert must be refused too, or TLS is silently off", p)
		} else if !strings.Contains(err.Error(), "tls.cert") {
			t.Errorf("%s: error %q must name the missing field", p, err)
		}

		// Both set is accepted.
		c2 := guardConfig(p)
		c2.TLS.Cert, c2.TLS.CertKey = "certs/tls.crt", "certs/tls.key"
		if err := c2.Validate(p); err != nil {
			t.Errorf("%s: both halves set must be accepted: %v", p, err)
		}

		// Neither set is accepted: TLS is optional, and a broker with no server
		// certificate is a legitimate deployment.
		if err := guardConfig(p).Validate(p); err != nil {
			t.Errorf("%s: no TLS at all must stay legal: %v", p, err)
		}
	}
}

// TestKubernetesNameIsBoundedByItsDerivedNames covers a name that passes on its own and
// still produces an object the cluster refuses.
//
// kubernetes.name is a DNS-1123 label, so 63 characters is legal for the name ITSELF.
// But the operator suffixes it into every object it creates, and Kubernetes copies a
// pod's name into the statefulset.kubernetes.io/pod-name label -- whose values stop at
// 63. So a 63-character name yields a 78-character pod name and a rejection naming a
// label the operator never wrote.
func TestKubernetesNameIsBoundedByItsDerivedNames(t *testing.T) {
	limit := maxDNSLabelLen - len(brokerPodSuffixShape)

	// Exactly at the limit is accepted.
	c := validK8sConfig()
	c.K8s.Name = strings.Repeat("a", limit)
	if err := c.Validate(K8s); err != nil {
		t.Errorf("a name of exactly %d characters must be accepted: %v", limit, err)
	}

	// One over is refused, and the message has to explain the derivation rather than
	// just asserting a number -- the operator did not choose the suffix.
	c2 := validK8sConfig()
	c2.K8s.Name = strings.Repeat("a", limit+1)
	err := c2.Validate(K8s)
	if err == nil {
		t.Fatalf("a name of %d characters must be refused: the derived pod name would exceed %d",
			limit+1, maxDNSLabelLen)
	}
	for _, want := range []string{"kubernetes.name", "pod name", "label"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q must mention %q so the cause is findable", err, want)
		}
	}

	// A 63-character name is a legal DNS-1123 label and must still be refused here,
	// which is the whole point: the label check alone was not enough.
	c3 := validK8sConfig()
	c3.K8s.Name = strings.Repeat("a", maxDNSLabelLen)
	if err := c3.Validate(K8s); err == nil {
		t.Error("a 63-character name passes the DNS-1123 label rule but must still be refused: its derived " +
			"pod name does not fit")
	}
}

// TestContainerNameMustBeAnEngineName covers the half of container.name's grammar the
// charset check did not: the FIRST character must be alphanumeric.
//
// `-broker` passes the charset and both engines refuse it -- but not before this tool
// has written it into a compose file and a quadlet unit, and not before it has reached
// argument positions where a leading dash reads as a flag rather than a name. A bound
// too, since container.name is the stem of every derived name and nothing else imposed
// one.
func TestContainerNameMustBeAnEngineName(t *testing.T) {
	for _, p := range []Platform{Docker, Podman} {
		for _, bad := range []string{"-broker", ".broker", "_broker", "-", "."} {
			c := validContainerConfig(p, "false")
			setContainerName(c, p, bad)
			err := c.Validate(p)
			if err == nil {
				t.Errorf("%s: container.name %q must be refused: both engines reject it, and a leading dash "+
					"reads as a flag on a command line", p, bad)
				continue
			}
			if !strings.Contains(err.Error(), "container.name") {
				t.Errorf("%s: error %q must name the field", p, err)
			}
		}
		// Legal engine names stay legal, including the ones a stricter grammar would
		// have refused for no reason.
		for _, good := range []string{"solace", "Solace", "sol.broker_1", "s", "0broker"} {
			c := validContainerConfig(p, "false")
			setContainerName(c, p, good)
			if err := c.Validate(p); err != nil {
				t.Errorf("%s: container.name %q must be accepted: %v", p, good, err)
			}
		}
		// Bounded.
		c := validContainerConfig(p, "false")
		setContainerName(c, p, "a"+strings.Repeat("b", maxContainerNameLen))
		if err := c.Validate(p); err == nil {
			t.Errorf("%s: a container.name longer than %d must be refused: it is the stem of every derived "+
				"name, and the host-side secret names are longer still", p, maxContainerNameLen)
		}
	}
}

// setContainerName writes container.name into whichever platform block p names.
func setContainerName(c *Config, p Platform, name string) {
	if p == Podman {
		c.Podman.Container.Name = name
		return
	}
	c.Docker.Container.Name = name
}

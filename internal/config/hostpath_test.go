package config

import (
	"fmt"
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
// including the two Windows forms that a command token may not carry, and every
// leading-tilde shape -- CheckHostPath no longer refuses one: expandHomePaths is
// what interprets it (and what refuses the unsupported `~user/...` form), so at
// the as-written charset gate a tilde is just another admitted character.
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
		"~",                       // expands to the home directory, alone
		"~/certs/tls.crt",         // expands to the home directory, forward slash
		`~\certs\tls.crt`,         // expands to the home directory, backslash
		"~bob/certs",              // the unsupported ~user form -- refused later, at expand time, not here
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

// TestExpandTilde covers expandTilde directly: a home-directory seam so no case
// depends on the environment the test happens to run in.
func TestExpandTilde(t *testing.T) {
	home := func() (string, error) { return "/home/op", nil }

	accept := []struct{ in, want string }{
		{"", ""},
		{"certs/tls.crt", "certs/tls.crt"}, // no leading tilde: untouched
		{"~", "/home/op"},                  // bare tilde
		{"~/certs/tls.crt", "/home/op/certs/tls.crt"},
		{`~\certs\tls.crt`, "/home/op/certs/tls.crt"}, // backslash form, ToSlash'd
		// An embedded tilde -- an 8.3 short name -- is not a LEADING one and must
		// survive completely untouched, backslashes included.
		{`C:\Users\RUNNER~1\AppData\Local\Temp`, `C:\Users\RUNNER~1\AppData\Local\Temp`},
	}
	for _, tc := range accept {
		got, err := expandTilde(tc.in, home)
		if err != nil {
			t.Errorf("expandTilde(%q) = %v, want accepted", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("expandTilde(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}

	t.Run("another user's home is refused, not guessed at", func(t *testing.T) {
		_, err := expandTilde("~bob/certs", home)
		if err == nil {
			t.Fatal("~bob/certs must be refused: this tool has no notion of another account's home")
		}
		if !strings.Contains(err.Error(), "another user's home") {
			t.Errorf("error %q must say why, not just that it failed", err)
		}
	})

	t.Run("an unresolvable home directory is an error", func(t *testing.T) {
		broken := func() (string, error) { return "", fmt.Errorf("no HOME set") }
		if _, err := expandTilde("~/certs", broken); err == nil {
			t.Error("an unresolvable home directory must fail rather than silently keep the literal ~")
		}
	})

	t.Run("a home directory containing whitespace is refused", func(t *testing.T) {
		// The real-world case: a Windows home like C:\Users\John Smith. The
		// charset gate already ran on the literal "~/certs" (which has no
		// space), so the expanded value has to be re-checked here or the space
		// would reach a compose/quadlet mount line this tool cannot delimit.
		spaced := func() (string, error) { return `C:\Users\John Smith`, nil }
		_, err := expandTilde("~/certs", spaced)
		if err == nil {
			t.Fatal("an expanded value containing whitespace must be refused")
		}
		if !strings.Contains(err.Error(), "whitespace") {
			t.Errorf("error %q must say why", err)
		}
	})
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

// TestHasPathSeparator pins the name-or-path question a bare CLI/shell script
// argument rests on (opK8sExecCLI/opCtrExecCLI and their shell-script siblings):
// a bare name resolves under broker.cliScriptsDir, a path is used as given.
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
			"tls.cert":                     func(c *Config) { c.TLS.Cert = "certs/$bad.crt" },
			"tls.certKey":                  func(c *Config) { c.TLS.CertKey = "certs/$bad.key" },
			"tls.cas[0]":                   func(c *Config) { c.TLS.CAs = []string{"certs/$bad.pem"} },
			"broker.cliScriptsDir":         func(c *Config) { c.Broker.CLIScriptsDir = "cli$x" },
			"broker.hostDiagnosticDir":     func(c *Config) { c.Broker.HostDiagnosticDir = "diag$x" },
			"broker.domainCerts.dirs[0]":   func(c *Config) { c.Broker.DomainCerts.Dirs = []CertDir{{Path: "certs$x"}} },
			"broker.domainCerts.files[ca]": func(c *Config) { c.Broker.DomainCerts.Files = map[string]string{"ca": "certs$x.pem"} },
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
			"tls.cert":                     func(c *Config) { c.TLS.Cert = "a/$b" },
			"tls.certKey":                  func(c *Config) { c.TLS.CertKey = "a/$b" },
			"tls.cas[0]":                   func(c *Config) { c.TLS.CAs = []string{"a/$b"} },
			"broker.cliScriptsDir":         func(c *Config) { c.Broker.CLIScriptsDir = "a/$b" },
			"broker.hostDiagnosticDir":     func(c *Config) { c.Broker.HostDiagnosticDir = "a/$b" },
			"broker.domainCerts.dirs[0]":   func(c *Config) { c.Broker.DomainCerts.Dirs = []CertDir{{Path: "a/$b"}} },
			"broker.domainCerts.files[ca]": func(c *Config) { c.Broker.DomainCerts.Files = map[string]string{"ca": "a/$b"} },
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
	c.Broker.HostDiagnosticDir = "diag-configs"
	c.Broker.DomainCerts.Dirs = []CertDir{{Path: "prod-cas"}, {Path: filepath.FromSlash("/abs/partner-cas")}}
	c.Broker.DomainCerts.Files = map[string]string{"my-ca": "certs/my-ca.pem"}
	c.Docker.ComposeFile = "docker-compose.yml"

	c.rebaseHostPaths()

	want := map[string]string{
		"tls.cert":                   filepath.Join("/srv/solace/env", "certs/tls.crt"),
		"tls.certKey":                filepath.FromSlash("/abs/tls.key"),
		"tls.cas[0]":                 filepath.Join("/srv/solace/env", "certs/ca.pem"),
		"tls.cas[1]":                 filepath.FromSlash("/abs/ca2.pem"),
		"broker.hostDiagnosticDir":   filepath.Join("/srv/solace/env", "diag-configs"),
		"docker.composeFile":         filepath.Join("/srv/solace/env", "docker-compose.yml"),
		"broker.domainCerts.dirs[0]": filepath.Join("/srv/solace/env", "prod-cas"),
		"broker.domainCerts.dirs[1]": filepath.FromSlash("/abs/partner-cas"),
		"broker.domainCerts.files":   filepath.Join("/srv/solace/env", "certs/my-ca.pem"),
	}
	got := map[string]string{
		"tls.cert":                   c.TLS.Cert,
		"tls.certKey":                c.TLS.CertKey,
		"tls.cas[0]":                 c.TLS.CAs[0],
		"tls.cas[1]":                 c.TLS.CAs[1],
		"broker.hostDiagnosticDir":   c.Broker.HostDiagnosticDir,
		"docker.composeFile":         c.Docker.ComposeFile,
		"broker.domainCerts.dirs[0]": c.Broker.DomainCerts.Dirs[0].Path,
		"broker.domainCerts.dirs[1]": c.Broker.DomainCerts.Dirs[1].Path,
		"broker.domainCerts.files":   c.Broker.DomainCerts.Files["my-ca"],
	}
	for k, w := range want {
		if got[k] != w {
			t.Errorf("after rebase %s = %q, want %q", k, got[k], w)
		}
	}
}

// TestRebaseIsANoOpWithoutABaseDir is the property internal/convert and every
// hand-built Config depend on: ApplyDefaults and Validate are called directly, with
// no path to derive a base from, and those paths must behave exactly as before.
func TestRebaseIsANoOpWithoutABaseDir(t *testing.T) {
	c := guardConfig(Docker)
	c.baseDir = ""
	c.TLS.Cert = "certs/tls.crt"
	c.Broker.HostDiagnosticDir = "diag-configs"
	c.rebaseHostPaths()
	if c.TLS.Cert != "certs/tls.crt" || c.Broker.HostDiagnosticDir != "diag-configs" {
		t.Errorf("with no base dir the paths must be untouched, got tls.cert=%q hostDiagnosticDir=%q",
			c.TLS.Cert, c.Broker.HostDiagnosticDir)
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

// TestContainerHostPathsRefuseATilde pins the fix for the gap that opened when
// CheckHostPath stopped refusing a leading '~' everywhere (expandHomePaths now
// handles it, for every field IT reaches): podman.quadletDir, podman.baseDir
// and <platform>.container.dataDir are deliberately NOT in that expansion list,
// because each names a path on the machine that runs the CONTAINER rather than
// the machine running this tool, and os.UserHomeDir() cannot answer for the
// former. Without a check here, a literal '~' -- an operator's typo, or the
// rootless quadletDir default's own xdgConfigHome fallback -- would silently
// reach the generated quadlet unit as a literal segment.
func TestContainerHostPathsRefuseATilde(t *testing.T) {
	cases := []struct {
		name  string
		p     Platform
		setup func(*Config)
		field string
	}{
		{"podman.quadletDir", Podman, func(c *Config) { c.Podman.QuadletDir = "~/quadlets" }, "podman.quadletDir"},
		{"podman.baseDir", Podman, func(c *Config) { c.Podman.BaseDir = "~/solace" }, "podman.baseDir"},
		{"podman.container.dataDir", Podman, func(c *Config) { c.Podman.Container.DataDir = "~/data" },
			"podman.container.dataDir"},
		{"docker.container.dataDir", Docker, func(c *Config) { c.Docker.Container.DataDir = "~/data" },
			"docker.container.dataDir"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := guardConfig(tc.p)
			tc.setup(c)
			err := c.validateHostPaths(tc.p)
			if err == nil {
				t.Fatalf("%s: a leading '~' must be refused, not silently left as a literal", tc.field)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Errorf("%s: error %q must name the field", tc.field, err)
			}
			if !strings.Contains(err.Error(), "not expanded for this field") {
				t.Errorf("%s: error %q must explain why (this field is on the container's host, not this "+
					"tool's), not just refuse", tc.field, err)
			}
		})
	}
	// docker.composeFile is expanded and rebased -- it IS read by this tool --
	// so it must NOT be refused here; it goes through the ordinary gate.
	dc := guardConfig(Docker)
	dc.Docker.ComposeFile = "~/compose.yml"
	if err := dc.validateHostPaths(Docker); err != nil {
		t.Errorf("docker.composeFile with a leading '~' must be accepted at validate time (expanded at "+
			"Load, not refused): %v", err)
	}
}

// TestDomainCertsDirsAreNotDefaulted replaces TestDomainCertsFolderIsDefaulted: the
// `folder` key it pinned defaulted to "certs", but its replacement, `dirs`,
// deliberately does NOT inherit that default. broker.configure domain-certs' own
// no-op ("no domain certificate authorities configured -- skipping") depends on
// dirs and files both being empty for a file that configures neither; defaulting
// dirs would turn that no-op into a hard failure (rule F: an unreadable configured
// dir is an error) for every deployment that has no ./certs directory beside its
// env file.
func TestDomainCertsDirsAreNotDefaulted(t *testing.T) {
	for _, p := range Platforms() {
		c := &Config{}
		c.ApplyDefaults(p)
		if len(c.Broker.DomainCerts.Dirs) != 0 {
			t.Errorf("%s: broker.domainCerts.dirs = %+v, want empty -- an unconfigured file must stay a no-op",
				p, c.Broker.DomainCerts.Dirs)
		}
		if len(c.Broker.DomainCerts.Files) != 0 {
			t.Errorf("%s: broker.domainCerts.files = %+v, want empty", p, c.Broker.DomainCerts.Files)
		}
	}
	// An explicit value still survives ApplyDefaults untouched.
	c := &Config{}
	c.Broker.DomainCerts.Dirs = []CertDir{{Path: "my-cas"}}
	c.ApplyDefaults(Docker)
	if len(c.Broker.DomainCerts.Dirs) != 1 || c.Broker.DomainCerts.Dirs[0].Path != "my-cas" {
		t.Errorf("an explicit dirs entry must not be overwritten, got %+v", c.Broker.DomainCerts.Dirs)
	}
}

// TestRenamedBrokerKeysFailLoud pins the remaining three retained-legacy-field
// migrations (the three .runtime keys have their own pin,
// TestRenamedRuntimeKeysFailLoud, in command_test.go): the old key decodes, is
// never defaulted, and Validate's error names the replacement.
func TestRenamedBrokerKeysFailLoud(t *testing.T) {
	cases := []struct {
		name string
		doc  string
		want string
	}{
		{"cliScriptsFolder", "broker:\n  cliScriptsFolder: cli\n", "broker.cliScriptsFolder was renamed to broker.cliScriptsDir"},
		{"diagDir", "broker:\n  diagDir: diag-configs\n", "broker.diagDir was renamed to broker.hostDiagnosticDir"},
		{"domainCerts.folder", "broker:\n  domainCerts:\n    folder: certs\n",
			"broker.domainCerts.folder was renamed to broker.domainCerts.dirs"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c Config
			if err := decodeStrict(tc.doc, &c); err != nil {
				t.Fatalf("the old key must still decode (that is the point): %v", err)
			}
			c.ApplyDefaults(K8s)
			err := c.Validate(K8s)
			if err == nil {
				t.Fatal("the old key must fail validation")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

// TestExpandHomePathsRunsBeforeRebaseAndDoesNotDependOnBaseDir is the regression
// pin for the whole feature: expandHomePaths must run so that a `~` value is
// already absolute by the time rebaseHostPaths sees it, and it must run even
// with baseDir == "" (the internal/convert / hand-built Config case), or the
// old bug -- a literal '~' segment joined onto the env file's directory --
// comes back.
func TestExpandHomePathsRunsBeforeRebaseAndDoesNotDependOnBaseDir(t *testing.T) {
	home := "/home/operator"
	for _, baseDir := range []string{"", filepath.FromSlash("/srv/solace/env")} {
		c := &Config{homeDir: func() (string, error) { return home, nil }}
		c.TLS.Cert = "~/certs/tls.crt"
		c.baseDir = baseDir

		if err := c.expandHomePaths(); err != nil {
			t.Fatalf("baseDir=%q: expandHomePaths: %v", baseDir, err)
		}
		c.rebaseHostPaths()

		want := "/home/operator/certs/tls.crt"
		if c.TLS.Cert != want {
			t.Errorf("baseDir=%q: tls.cert = %q, want %q", baseDir, c.TLS.Cert, want)
		}
		if strings.Contains(c.TLS.Cert, "~") {
			t.Errorf("baseDir=%q: tls.cert = %q still carries a literal '~'", baseDir, c.TLS.Cert)
		}
		if baseDir != "" && strings.Contains(c.TLS.Cert, baseDir) {
			t.Errorf("baseDir=%q: tls.cert = %q was joined onto baseDir; an expanded (absolute) "+
				"value must not be rebased", baseDir, c.TLS.Cert)
		}
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

// TestLoadExpandsATildeThroughTheWholePipeline is the WIRING test, and it is the one that
// was missing: every other tilde test calls expandTilde or expandHomePaths directly, so
// deleting the single call from Load left the whole suite green.
//
// That is a real gap rather than a stylistic one. The behaviour this feature replaced --
// CheckHostPath's outright refusal of a leading '~' -- was reached through Validate, whose
// wiring into Load is itself pinned. Swapping a refusal for an expansion without pinning
// the new path would have traded a tested behaviour for an untested one.
//
// It drives the REAL os.UserHomeDir rather than the injectable seam, because the seam is
// unexported and Load builds its own Config -- and because the seam is exactly what this
// test must not use: the thing under test is that Load calls the expansion at all.
// os.UserHomeDir reads HOME on unix and USERPROFILE on Windows, so both are set.
func TestLoadExpandsATildeThroughTheWholePipeline(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	path := writeTempYAML(t, `redundancy:
  enabled: "false"
image:
  repo: solace/broker
  tag: latest
semp:
  adminPass: s3cret
broker:
  cliScriptsDir: ~/scripts
kubernetes:
  name: mybroker
  namespace: sol-ns
  storage:
    msgNodeSize: 30Gi
`)

	c, err := Load(path, K8s)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if strings.Contains(c.Broker.CLIScriptsDir, "~") {
		t.Fatalf("broker.cliScriptsDir = %q: the tilde survived Load, so nothing expanded it",
			c.Broker.CLIScriptsDir)
	}
	// And it expanded to the HOME above rather than being rebased onto the env file's
	// own directory -- which is precisely what the refusal this replaced existed to
	// prevent, and what would happen again if expansion ran after the rebase.
	if !strings.HasPrefix(filepath.ToSlash(c.Broker.CLIScriptsDir), filepath.ToSlash(home)) {
		t.Errorf("broker.cliScriptsDir = %q, want it under the home directory %q -- a value "+
			"rebased onto the env file's directory instead is the original bug",
			c.Broker.CLIScriptsDir, home)
	}
	if !strings.HasSuffix(filepath.ToSlash(c.Broker.CLIScriptsDir), "/scripts") {
		t.Errorf("broker.cliScriptsDir = %q, want it to keep the path after the tilde",
			c.Broker.CLIScriptsDir)
	}
}

// TestExpandHomePathsCoversEveryFieldItClaims walks EVERY field expandHomePaths lists,
// which is the only way to catch the failure this pass is most likely to have.
//
// The risk is not a wrong expansion; it is a MISSING one. The function is a hand-written
// list of fields, so a field added to the schema later -- or one dropped from this list in
// a refactor -- simply never expands, and the value reaches the filesystem with a literal
// '~' in it. That surfaces as a missing file somewhere far from the env file, which is
// precisely the failure the refusal this feature replaced existed to prevent.
//
// Setting every field at once also pins the loops: the CA list, the certificate directory
// list and the files map are each iterated separately, and a loop that expanded only its
// first element would pass any test that set one value per field.
func TestExpandHomePathsCoversEveryFieldItClaims(t *testing.T) {
	const home = "/home/op"
	c := &Config{homeDir: func() (string, error) { return home, nil }}
	c.TLS.Cert = "~/tls/cert.pem"
	c.TLS.CertKey = "~/tls/key.pem"
	c.TLS.CAs = []string{"~/tls/ca1.pem", "~/tls/ca2.pem"}
	c.Broker.CLIScriptsDir = "~/scripts"
	c.Broker.HostDiagnosticDir = "~/diag"
	c.Docker.ComposeFile = "~/compose.yaml"
	c.Broker.DomainCerts.Dirs = []CertDir{{Path: "~/cas/a"}, {Path: "~/cas/b"}}
	c.Broker.DomainCerts.Files = map[string]string{
		"first":  "~/cas/one.pem",
		"second": "~/cas/two.pem",
	}

	if err := c.expandHomePaths(); err != nil {
		t.Fatalf("expandHomePaths: %v", err)
	}

	got := map[string]string{
		"tls.cert":                    c.TLS.Cert,
		"tls.certKey":                 c.TLS.CertKey,
		"tls.cas[0]":                  c.TLS.CAs[0],
		"tls.cas[1]":                  c.TLS.CAs[1],
		"broker.cliScriptsDir":        c.Broker.CLIScriptsDir,
		"broker.hostDiagnosticDir":    c.Broker.HostDiagnosticDir,
		"docker.composeFile":          c.Docker.ComposeFile,
		"broker.domainCerts.dirs[0]":  c.Broker.DomainCerts.Dirs[0].Path,
		"broker.domainCerts.dirs[1]":  c.Broker.DomainCerts.Dirs[1].Path,
		"broker.domainCerts.files[1]": c.Broker.DomainCerts.Files["first"],
		"broker.domainCerts.files[2]": c.Broker.DomainCerts.Files["second"],
	}
	for field, v := range got {
		if strings.Contains(v, "~") {
			t.Errorf("%s = %q: the tilde survived, so this field is not in expandHomePaths' list",
				field, v)
		}
		if !strings.HasPrefix(filepath.ToSlash(v), home) {
			t.Errorf("%s = %q, want it under %q", field, v, home)
		}
	}
}

// TestExpandHomePathsReportsWhichFieldFailed pins that a failure names the field, for each
// SHAPE of field the pass holds -- a plain one, a list element, and a map entry.
//
// The field name is the whole value of the error here. An operator sees the message, not
// the struct, and "starts with '~'" on its own would leave them searching an env file for
// which of a dozen path keys caused it.
func TestExpandHomePathsReportsWhichFieldFailed(t *testing.T) {
	// "~other" is the unsupported another-user's-home form, which is refused by name
	// rather than guessed at -- so it is the shape that drives the error path.
	for _, tc := range []struct {
		name, wantField string
		set             func(c *Config)
	}{
		{"a plain field", "broker.cliScriptsDir", func(c *Config) { c.Broker.CLIScriptsDir = "~other/x" }},
		{"a list element", "tls.cas[1]", func(c *Config) {
			c.TLS.CAs = []string{"/fine/ca.pem", "~other/ca.pem"}
		}},
		{"a certificate dir", "broker.domainCerts.dirs[0]", func(c *Config) {
			c.Broker.DomainCerts.Dirs = []CertDir{{Path: "~other/cas"}}
		}},
		{"a map entry, named by its CA name", "broker.domainCerts.files[my-ca]", func(c *Config) {
			c.Broker.DomainCerts.Files = map[string]string{"my-ca": "~other/ca.pem"}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := &Config{homeDir: func() (string, error) { return "/home/op", nil }}
			tc.set(c)
			err := c.expandHomePaths()
			if err == nil {
				t.Fatal("an unsupported ~user form must be refused rather than guessed at")
			}
			if !strings.Contains(err.Error(), tc.wantField) {
				t.Errorf("err = %v, want it to name the field %q -- an operator reads the message, "+
					"not the struct", err, tc.wantField)
			}
		})
	}
}

// TestExpandHomePathsSurfacesAnUnresolvableHome pins the remaining branch: when the home
// directory itself cannot be determined, that is an error naming the field rather than a
// silent fallthrough leaving the tilde in place.
func TestExpandHomePathsSurfacesAnUnresolvableHome(t *testing.T) {
	c := &Config{homeDir: func() (string, error) { return "", fmt.Errorf("no home for this user") }}
	c.Broker.CLIScriptsDir = "~/scripts"

	err := c.expandHomePaths()
	if err == nil {
		t.Fatal("an unresolvable home must fail rather than leave the tilde in the value")
	}
	if !strings.Contains(err.Error(), "broker.cliScriptsDir") {
		t.Errorf("err = %v, want it to name the field", err)
	}
	if strings.Contains(c.Broker.CLIScriptsDir, "~") && err == nil {
		t.Error("the value must not be left half-expanded on failure")
	}
}

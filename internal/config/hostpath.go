package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

// A host path is not a command token, and the difference is the whole reason this
// file exists beside execguard.go.
//
// A Command-typed field becomes argv. A host-path field becomes a value
// CONCATENATED into an artifact some other program parses -- a quadlet
// `Volume=<src>:<dst>:ro` line, a compose `- "<src>:<dst>:Z"` entry -- or the
// argument of an os.ReadFile on this machine. Neither ever reaches a shell, so the
// hazard is not injection; it is that a `$` in a compose scalar is interpolated by
// compose, and a `:` in an unquoted quadlet directive changes where the field
// splits. So the two charsets differ, and they differ in one direction only: a
// path admits three characters a token refuses.
//
// The set is DERIVED from unsafeTokenChars by removal rather than typed out again,
// which is what stops the two drifting apart when one is edited.
// TestHostPathCharsetIsDerivedFromTheTokenCharset pins the derivation itself.
//
// Two characters are removed:
//
//   - backslash, because `C:\certs\tls.crt` is a legitimate host path on a Windows
//     machine authoring or running this tool. Forcing forward slashes there would
//     be a rule with no purpose: the value is never argv[0], so the bare-name rule
//     that motivates the token ban does not apply.
//   - tilde, because `%TEMP%` on a GitHub Windows runner really is an 8.3 short name
//     (`C:\Users\RUNNER~1\...`). Refusing it made the sibling project's own Windows
//     CI fail on a path the operator never chose. See the leading-tilde carve-out in
//     CheckHostPath, which is a separate question from the charset.
//
// A colon needs no removal: unsafeTokenChars never contained one, because
// `--server=https://host:6443` is a legitimate command token. It is called out here
// because a colon in a host path is NOT harmless the way a backslash is -- both
// container artifacts write a mount as `source:target:options`, so a drive-letter
// path adds a field the engine did not expect. That is a rendering problem rather
// than a charset one, and it is handled where the mount line is written.
var hostPathUnsafe = strings.Map(func(r rune) rune {
	if strings.ContainsRune(hostPathAdmits, r) {
		return -1
	}
	return r
}, unsafeTokenChars)

// hostPathAdmits are the token-charset characters a host path is allowed to carry.
// Named so the test can assert the relationship against one definition instead of
// repeating the list.
const hostPathAdmits = `\~`

// CheckHostPath is the single definition of what a host filesystem path may be in
// an env file. It runs on the value AS WRITTEN, before any base-directory
// rebasing: the rule exists to police what the file says, which is the only thing
// this tool controls. Gating the rebased value would fail every load from a
// checkout whose own directory contains a space, which is not the env file's fault
// and would make this repository's own test suite unrunnable from such a path.
//
// An empty path is accepted here. Whether a given field may be empty is a
// per-field question the mandatory-field checks already answer, and answering it
// twice would produce two different messages for one mistake.
func CheckHostPath(field, p string) error {
	if p == "" {
		return nil
	}
	if j := strings.IndexFunc(p, isCtrl); j >= 0 {
		return fmt.Errorf("%s contains a control character (0x%02x) at offset %d: %q", field, p[j], j, p)
	}
	if j := strings.IndexFunc(p, isInvisible); j >= 0 {
		return fmt.Errorf("%s = %q contains an invisible formatting character at offset %d; it cannot be "+
			"seen in a review or a log line, so it is not allowed in a path", field, p, j)
	}
	// Whitespace is refused for a narrower reason than in a token: both container
	// artifacts write a mount as `<src>:<dst>:<opts>` on one line, the quadlet form
	// unquoted, so a space in the source is a field this tool cannot delimit.
	if j := strings.IndexFunc(p, isSpace); j >= 0 {
		return fmt.Errorf("%s = %q contains whitespace at offset %d; a mount line is written as "+
			"source:target:options on a single line, so a path with a space in it cannot be delimited",
			field, p, j)
	}
	if j := strings.IndexAny(p, hostPathUnsafe); j >= 0 {
		return fmt.Errorf("%s = %q contains %q, which is not allowed in a path: it is written verbatim into "+
			"the generated compose file or quadlet unit, where that character changes how the engine reads "+
			"the line", field, p, string(p[j]))
	}
	// A leading tilde is refused even though the charset admits the character.
	// Nothing in this tool expands it -- there is no home-directory expansion on
	// any path field -- so `~/certs/tls.crt` would be joined onto the env file's
	// directory and produce a literal `~` segment: a path that does not exist,
	// reported as a missing file somewhere far from the typo. An 8.3 short name
	// like RUNNER~1 carries its tilde in the middle and is unaffected.
	if strings.HasPrefix(p, "~") {
		return fmt.Errorf("%s = %q starts with '~', which this tool does not expand; write the path out in "+
			"full, or make it relative to the env file's own directory", field, p)
	}
	return nil
}

// IsAbsHostPath reports whether p is already absolute for SOME host, not for the
// machine currently running. filepath.IsAbs answers only for this OS, so on Linux
// it calls `C:/certs/tls.crt` relative and on Windows it calls `/opt/solace/data`
// relative -- and both of those are absolute where they will actually be used. An
// env file authored on one OS and run on the other is ordinary here, so the
// question has to be asked for both.
// It deliberately does NOT consult filepath.IsAbs. That would make the answer
// depend on the machine running, which is the opposite of this function's purpose:
// on Windows filepath.IsAbs("1:/certs") is true, because its volume-name check
// accepts any character before the colon, so the same env file would be rebased on
// Linux and left alone on Windows. Every shape filepath.IsAbs recognises is covered
// explicitly below -- a leading separator (posix roots and UNC alike) and a real
// drive letter -- so nothing is lost by answering it here.
func IsAbsHostPath(p string) bool {
	if p == "" {
		return false
	}
	if p[0] == '/' || p[0] == '\\' {
		return true
	}
	// A drive-letter root: C:\ or C:/, and the bare C: form, which names a
	// volume-relative path rather than a truly absolute one but is not something to
	// rebase either. The letter is checked: `1:/certs` is not a drive.
	if len(p) >= 2 && p[1] == ':' {
		c := p[0]
		return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
	}
	return false
}

// BaseName returns the last element of a path, splitting on BOTH separators
// regardless of the OS running.
//
// filepath.Base is OS-dependent and path.Base only ever splits on '/', so an env
// file authored on Windows with `certs\tls.crt` produced `certs\tls.crt` as the
// "basename" on Linux and `tls.crt` on Windows -- for the same env file. That
// basename is used as an in-broker filename and as a Secret data key, so the two
// answers are not cosmetically different, they name different objects. Three
// separate rules for this existed in the tree before this helper: path.Base in the
// copy-from paths, filepath.Base in the CLI-script upload, and a hand-rolled
// strings.ContainsAny(file, "/\\") in two command files.
func BaseName(p string) string {
	if i := strings.LastIndexAny(p, `/\`); i >= 0 {
		return p[i+1:]
	}
	return p
}

// HasPathSeparator reports whether p carries a directory component under either
// separator. It is the question "did the operator give me a name or a path?",
// which decides whether a value is resolved under a configured folder or used as
// typed, and it must answer the same way on both operating systems for the same
// env file.
func HasPathSeparator(p string) bool {
	return strings.ContainsAny(p, `/\`)
}

// BaseDir is the directory of the env file this Config was loaded from, or empty
// for a Config that never went through Load. Exported for error messages and for
// the tests that assert what the rebase did; nothing needs it to resolve a path,
// because rebaseHostPaths has already rewritten every field in place.
func (c *Config) BaseDir() string { return c.baseDir }

// rebaseHostPaths resolves every relative HOST path in the file against the
// directory the env file itself was found in, and it is the fix for a silent
// failure rather than a convenience.
//
// The one that mattered: podman reads a `Volume=` source with no leading
// separator as the name of a NAMED VOLUME, not as a path. A shipped
// `Volume=certs/tls.crt:...` therefore made podman create an empty volume and
// mount it over the certificate, with no error at any point -- the unit started,
// the broker started, and TLS was simply not what the operator configured. systemd
// starts a unit with no useful working directory, so there was never a cwd that
// would have made the relative form work either.
//
// Choosing the env file's directory over the process working directory is what
// makes a deployment reproducible: the same env file describes the same
// deployment whichever directory it is driven from, which is not true of a
// cwd-relative path.
//
// It runs AFTER Validate, deliberately. Validate's host-path charset gate polices
// what the FILE says, and it is the only thing this tool controls; running it on a
// rebased value would fail every load from a checkout whose own directory contains
// a space. So: validate as written, then normalise.
//
// An empty baseDir is a no-op, which is what keeps internal/convert and every
// hand-built Config behaving as before.
func (c *Config) rebaseHostPaths() {
	if c.baseDir == "" {
		return
	}
	// Every field here is a host path on the machine running this tool. What is
	// deliberately ABSENT matters as much as what is present:
	//
	//   podman.quadletDir       -- both defaults are already absolute, and the unit
	//                              must live where systemd scans. Accepting a
	//                              relative value by rebasing it would invent a
	//                              location systemd would never read.
	//   podman.baseDir          -- mandatory and required absolute instead. It is a
	//                              `Volume=` source (the named-volume trap again) and
	//                              it receives a file containing a PRIVATE KEY, so
	//                              resolving a relative value would invent a location
	//                              for that key which the operator never named.
	//   <platform>.container.dataDir
	//                           -- required absolute instead (validateContainer).
	//                              It is the target of a recursive delete on
	//                              `remove broker --delete-data`, and quietly
	//                              moving what gets deleted is not a fix.
	//   broker.domainCerts.files
	//                           -- each value is simultaneously the host-side
	//                              filename AND the in-broker name under
	//                              /usr/sw/jail/certs. validName already forbids a
	//                              separator, so there is nothing to rebase.
	//   broker.productKeys      -- Solace licence strings, not paths.
	for _, p := range []*string{
		&c.TLS.Cert,
		&c.TLS.CertKey,
		&c.Broker.CLIScriptsFolder,
		&c.Broker.DiagDir,
		&c.Broker.DomainCerts.Folder,
		&c.Docker.ComposeFile,
	} {
		*p = c.hostPath(*p)
	}
	for i := range c.TLS.CAs {
		c.TLS.CAs[i] = c.hostPath(c.TLS.CAs[i])
	}
}

// hostPath joins one relative path onto the base directory, leaving an empty or
// already-absolute value alone. IsAbsHostPath rather than filepath.IsAbs, because
// an env file authored on Windows and run on Linux (or the reverse) is ordinary
// here and `C:/certs/tls.crt` must not be joined onto anything.
func (c *Config) hostPath(p string) string {
	if p == "" || IsAbsHostPath(p) {
		return p
	}
	return filepath.Join(c.baseDir, p)
}

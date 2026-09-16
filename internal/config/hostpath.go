package config

import (
	"fmt"
	"os"
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
// One character is removed: the backslash, because `C:\certs\tls.crt` is a
// legitimate host path on a Windows machine authoring or running this tool. Forcing
// forward slashes there would be a rule with no purpose: the value is never argv[0],
// so the bare-name rule that motivates the token ban does not apply.
//
// The tilde is not removed here, because the token charset does not carry it: a command
// token's tilde is refused POSITIONALLY (execguard.go's checkToken refuses a leading one
// and admits an embedded one), which is the rule host paths already used. That is what
// lets an 8.3 short name through -- `C:\Users\RUNNER~1\...` is what a GitHub Windows
// runner really hands out.
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
//
// It is just the backslash. The tilde is not in unsafeTokenChars at all -- a command
// token's tilde is refused POSITIONALLY (checkToken refuses a leading one, admits an 8.3
// short name's embedded one) -- so there is nothing to remove here.
const hostPathAdmits = `\`

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
	// A LEADING tilde is refused, and this is the FAIL-CLOSED half of the expansion
	// feature rather than a rejection of the syntax: expandHomePaths runs over every
	// host-path field in this schema BEFORE Validate calls this, so an operator who
	// writes `~/solace/data` never reaches here -- the value arriving is already
	// their home directory. One surviving means the field is not in
	// expandHomePaths' list (a new key someone forgot) or the Config never went
	// through Load, and the alternative to failing is a literal `~` reaching a
	// quadlet `Volume=` line, a compose mount source, or the `rm -rf` behind
	// `broker remove --delete-data`, each of which would silently act on a
	// directory named `~` instead. The message therefore names the field.
	//
	// An EMBEDDED tilde is untouched, here and in expandTilde: `C:/Users/RUNNER~1/`
	// is an 8.3 short name and a perfectly ordinary path, and it is what expansion
	// itself produces on such a machine.
	if strings.HasPrefix(p, "~") {
		return fmt.Errorf("%s = %q still starts with '~', which no later step expands: it would be written "+
			"verbatim into a mount line or handed to a file operation, naming a directory literally called "+
			"'~'. Every path key in this schema has a leading '~' expanded to the home directory of the user "+
			"running this tool while the env file is loaded, so one surviving here means this configuration "+
			"was not loaded from a file -- write the path out in full", field, p)
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
	if isPathSep(p[0]) {
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
	if i := strings.LastIndexAny(p, pathSeparators); i >= 0 {
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
	return strings.ContainsAny(p, pathSeparators)
}

// isPathSep is the same question about ONE byte, for the callers examining a known
// position rather than searching. Both read the separator set from the one constant
// (execguard.go's pathSeparators), so a path rule and a command rule cannot come to
// disagree about what a separator is.
func isPathSep(b byte) bool { return strings.IndexByte(pathSeparators, b) >= 0 }

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
	//   broker.productKeys      -- Solace licence strings, not paths.
	//
	// broker.domainCerts.files and every broker.domainCerts.dirs[i].path ARE rebased,
	// below, in a loop of their own: a files value is a FULL host path (the CA name,
	// not the value, is the in-broker filename), and map values are not addressable,
	// so they cannot join the []*string loop a struct field can.
	for _, p := range []*string{
		&c.TLS.Cert,
		&c.TLS.CertKey,
		&c.Broker.CLIScriptsDir,
		&c.Broker.HostDiagnosticDir,
		&c.Docker.ComposeFile,
	} {
		*p = c.hostPath(*p)
	}
	for i := range c.TLS.CAs {
		c.TLS.CAs[i] = c.hostPath(c.TLS.CAs[i])
	}
	for i := range c.Broker.DomainCerts.Dirs {
		c.Broker.DomainCerts.Dirs[i].Path = c.hostPath(c.Broker.DomainCerts.Dirs[i].Path)
	}
	for ca, p := range c.Broker.DomainCerts.Files {
		c.Broker.DomainCerts.Files[ca] = c.hostPath(p)
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

// expandHomePaths expands a leading `~` in every host path THIS TOOL ITSELF
// READS off the machine it runs on. It is its own pass (Load calls it between
// Validate and rebaseHostPaths) rather than living inside rebaseHostPaths,
// deliberately: rebaseHostPaths is a no-op with no baseDir, which is exactly
// how internal/convert and every hand-built Config behave, and gating
// expansion behind that same guard would silently skip it for both -- leaving
// a literal `~` to reach a quadlet `Volume=` line, which is the failure this
// feature exists to end. Nothing here depends on baseDir.
//
// Order matters twice over, both directions:
//
//   - It runs BEFORE Validate/CheckHostPath. That is the opposite of where this
//     pass started, and the reason it moved is that a directory key must be
//     allowed to be relative to a home directory and still satisfy the checks
//     that require it to be ABSOLUTE: `dataDir: ~/solace/data` is not absolute by
//     IsAbsHostPath, so validated first it was refused before anything could
//     expand it. Running first means every host-path check -- the charset, the
//     absoluteness requirements on podman.baseDir and container.dataDir, and
//     CheckHostPath's own leading-`~` refusal -- sees the value that will
//     actually be used.
//     The cost, stated plainly: the charset gate now also polices a home
//     directory the env file did not name, so a home containing `'` or a space
//     fails the load. That is deliberate rather than accepted -- such a value
//     lands unquoted in a `Volume=` line -- and expandTilde reports the
//     whitespace case itself, with the resolved value in the message. (The
//     env file's own directory is still gated nowhere, because rebaseHostPaths
//     runs after Validate: an operator opts into expansion by writing `~`,
//     where the rebase happens to every relative path whether they think about
//     it or not.)
//   - It runs BEFORE rebaseHostPaths. An expanded value is already absolute, so
//     IsAbsHostPath leaves it alone in hostPath -- the join can no longer
//     happen. Expanding AFTER the join would look for a leading tilde in
//     "<baseDir>/~/certs", find none, and silently do nothing: the exact bug
//     this feature replaces, reintroduced one step later.
//
// EVERY host-path key in this schema is covered, which is what makes the rule
// something an operator can hold in their head -- `~` works in a path key, full
// stop -- and what lets CheckHostPath refuse a surviving one instead of having to
// decide per field whether it means anything.
//
// The three container-host keys (podman.quadletDir, podman.baseDir,
// <platform>.container.dataDir) were excluded at first, on the argument that they
// name a path on the machine running the CONTAINER while os.UserHomeDir answers for
// the machine running this TOOL. For docker and podman those are the same machine,
// and the code says so rather than assuming it: this process writes the quadlet unit
// to quadletDir with os.WriteFile and deletes it with os.Remove
// (container.Manager.writeArtifact / deletePodman), writes the server-certificate
// bundle under baseDir the same way, and creates, chowns and `rm -rf`s dataDir
// through local subprocesses of its own (Manager.PrepHost, purgeData) before the
// locally-invoked engine ever reads it as a bind-mount source. The rootless
// quadletDir DEFAULT is itself built from os.UserHomeDir (load.go's xdgConfigHome),
// so the field was already answering for this machine's home before any of this.
// The exclusion cost the operator the one platform where a home-relative directory
// is most natural, and bought nothing.
//
// One knock-on worth knowing: xdgConfigHome falls back to the literal string "~"
// when it cannot resolve a home directory, and that default now flows through this
// pass, which fails naming the unresolvable home -- where before it reached the
// charset gate and was refused for the wrong reason.
func (c *Config) expandHomePaths() error {
	home := c.home()
	expand := func(field string, p *string) error {
		v, err := expandTilde(*p, home)
		if err != nil {
			return fmt.Errorf("%s starts with '~': %w", field, err)
		}
		*p = v
		return nil
	}
	if err := expand("tls.cert", &c.TLS.Cert); err != nil {
		return err
	}
	if err := expand("tls.certKey", &c.TLS.CertKey); err != nil {
		return err
	}
	for i := range c.TLS.CAs {
		if err := expand(fmt.Sprintf("tls.cas[%d]", i), &c.TLS.CAs[i]); err != nil {
			return err
		}
	}
	if err := expand("broker.cliScriptsDir", &c.Broker.CLIScriptsDir); err != nil {
		return err
	}
	if err := expand("broker.hostDiagnosticDir", &c.Broker.HostDiagnosticDir); err != nil {
		return err
	}
	if err := expand("docker.composeFile", &c.Docker.ComposeFile); err != nil {
		return err
	}
	// The four container-host directories, expanded for the machine that actually
	// writes and deletes them (the paragraph above has the evidence). They are listed
	// unconditionally rather than per platform, the way every field above is: an
	// unset field is left alone by expandTilde, and a platform-shaped list here would
	// be a second enumeration of the schema to keep in step with validateHostPaths'.
	if err := expand("docker.container.dataDir", &c.Docker.Container.DataDir); err != nil {
		return err
	}
	if err := expand("podman.quadletDir", &c.Podman.QuadletDir); err != nil {
		return err
	}
	if err := expand("podman.baseDir", &c.Podman.BaseDir); err != nil {
		return err
	}
	if err := expand("podman.container.dataDir", &c.Podman.Container.DataDir); err != nil {
		return err
	}
	for i := range c.Broker.DomainCerts.Dirs {
		field := fmt.Sprintf("broker.domainCerts.dirs[%d]", i)
		if err := expand(field, &c.Broker.DomainCerts.Dirs[i].Path); err != nil {
			return err
		}
	}
	for ca, p := range c.Broker.DomainCerts.Files {
		v, err := expandTilde(p, home)
		if err != nil {
			return fmt.Errorf("broker.domainCerts.files[%s] starts with '~': %w", ca, err)
		}
		c.Broker.DomainCerts.Files[ca] = v
	}
	return nil
}

// home is the home-directory resolver this Config's expansion passes use: the test
// seam when one was installed, os.UserHomeDir otherwise. One accessor rather than
// the same two-line fallback in expandHomePaths and expandCommandHomes, so a Config
// built by hand cannot have one pass see the seam and the other the real machine.
func (c *Config) home() func() (string, error) {
	if c.homeDir != nil {
		return c.homeDir
	}
	return os.UserHomeDir
}

// expandTilde expands a leading tilde in a HOST PATH. The expansion itself is
// tildeHome's; what this wrapper adds is the whitespace re-check, whose message is
// about a mount line -- see expandTildeToken for the command-token twin, which is the
// same expansion under a different explanation.
//
// Re-checked here, and not only as written: CheckHostPath ran on the literal
// "~/...", which never carries whitespace -- but a real Windows home routinely does
// (`C:\Users\John Smith`), and that space would otherwise reach a compose/quadlet
// mount line this tool cannot delimit, silently.
func expandTilde(p string, homeDir func() (string, error)) (string, error) {
	expanded, err := tildeHome(p, homeDir)
	if err != nil {
		return "", err
	}
	if strings.IndexFunc(expanded, isSpace) >= 0 {
		return "", fmt.Errorf("expands to %q, which contains whitespace; a mount line is written as "+
			"source:target:options on a single line and cannot delimit a path with a space in it -- write "+
			"the path out in full instead of relying on '~' expansion", expanded)
	}
	return expanded, nil
}

// expandTildeToken expands a leading tilde in a COMMAND TOKEN (execguard.go's
// expandCommandHomes). It is expandTilde's twin and differs only in what it says
// about whitespace: an expanded home carrying a space would reach CheckCommand's own
// whitespace refusal, which explains itself by telling the operator to write each
// argument as its own list entry -- true of a token they typed, and misleading advice
// for a space that came out of their home directory rather than their env file.
func expandTildeToken(tok string, homeDir func() (string, error)) (string, error) {
	expanded, err := tildeHome(tok, homeDir)
	if err != nil {
		return "", err
	}
	if strings.IndexFunc(expanded, isSpace) >= 0 {
		return "", fmt.Errorf("expands to %q, which contains whitespace; a command is split into arguments on "+
			"whitespace, so this token cannot be delimited -- pass the path in full instead of relying on '~' "+
			"expansion", expanded)
	}
	return expanded, nil
}

// tildeHome expands a LEADING tilde only: exactly "~", or "~" immediately
// followed by '/' or '\\'. A tilde anywhere else in the value -- including one
// followed by any other character -- is left untouched, which is what keeps an
// 8.3 short name (`C:\Users\RUNNER~1\...`) working: CheckHostPath and checkToken
// both draw this same line between "leading" and "everywhere else", and this is
// the resolving half of that one rule. Anything starting `~` and
// followed by something other than a separator is the `~user/...` form (another
// user's home directory) and is refused by name rather than guessed at -- this
// tool has no notion of another account's home.
func tildeHome(p string, homeDir func() (string, error)) (string, error) {
	if p == "" || p[0] != '~' {
		return p, nil
	}
	rest := p[1:]
	if rest != "" && !isPathSep(rest[0]) {
		return "", fmt.Errorf("%q names another user's home directory (a `~user/...` form), which this "+
			"tool does not support -- write the path out in full", p)
	}
	home, err := homeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "", fmt.Errorf("the home directory of the user running this tool could not be resolved: %v", err)
	}
	// String concatenation, not filepath.Join: the result may be read by this
	// tool (fine on either OS) or written into a compose/quadlet artifact as a
	// path on the Linux host that runs the container, and filepath.Join would
	// apply this OS's own cleaning rules to a value built from two otherwise-
	// unrelated strings.
	//
	// The slash normalisation is a ReplaceAll and deliberately NOT
	// filepath.ToSlash, for the reason BaseName gives at the top of this file:
	// ToSlash rewrites only THIS OS's separator, so on Linux it is the identity
	// function -- and the backslashes here did not come from this OS. They came
	// out of the env file, which is routinely authored on Windows and run on
	// Linux. ToSlash therefore answered `~\certs\tls.crt` two different ways for
	// the same file, normalising it on Windows and leaving `<home>\certs\tls.crt`
	// on Linux, where that whole value is one filename. (load.go's xdgConfigHome
	// keeps ToSlash correctly: its separators come from a filepath.Join on the
	// machine running, so there is only ever one to rewrite.)
	// Whitespace in the resolved home is left to the two callers above: it is
	// refused by both, in the words that fit where the value is going.
	return strings.ReplaceAll(home+rest, `\`, "/"), nil
}

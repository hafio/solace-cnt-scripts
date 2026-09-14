package config

import (
	"fmt"
	"path"
	"strings"
)

// domaincerts.go derives the NAME a domain certificate authority is created under,
// for the certificates found by walking broker.domainCerts.dirs.
//
// The name does three jobs at once, which is why its rules are as strict as they are:
//
//   - it is the operand of the broker's own `create domain-certificate-authority <name>`
//     (scripts.go, domainCertsScript),
//   - it is the FILENAME the certificate is uploaded under inside the broker, beneath
//     the certs directory (broker.certPath),
//   - it is the key an operator writes in `files:` when naming one explicitly, and the
//     key `broker configure domain-certs --remove` deletes by.
//
// So it has to be simultaneously a legal CLI operand, a legal filename, and stable
// across runs -- a name that changed between two runs would create a second CA rather
// than update the first. Letters, digits, underscore, dash and period, capped at
// CANameMax, is the intersection that satisfies all three (operator, 2026-09-14).
//
// An explicitly named certificate under `files:` is NOT derived: the operator chose the
// name and it is used verbatim, checked against the same rules.

// CANameMax is the longest a domain CA name may be. The broker's own limit; a longer
// name is refused rather than silently truncated, because truncation is how two
// certificates quietly become one CA.
const CANameMax = 64

// caNameSafe reports whether r may appear in a CA name. Deliberately narrow: this set
// is safe as a CLI operand AND as a filename on every platform this tool runs on, which
// a wider set (spaces, colons, slashes) would not be.
func caNameSafe(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return true
	case r == '_', r == '-', r == '.':
		return true
	}
	return false
}

// sanitiseCAPart rewrites one component -- a directory name or a filename -- into the
// safe charset, collapsing each run of unsafe characters to a single underscore.
//
// COLLAPSING rather than one-for-one substitution is deliberate. A real certificate is
// called `DigiCert Global Root G2.crt`, and one-for-one on a name with a doubled space
// or a " - " separator yields runs of underscores that read as noise and make two
// visibly different files look alike. Leading and trailing underscores are trimmed for
// the same reason: they carry no information and only spend the length budget.
//
// The result can be empty (a component made entirely of unsafe characters), which the
// caller must treat as a failure rather than silently dropping the component.
func sanitiseCAPart(s string) string {
	var b strings.Builder
	lastUnsafe := false
	for _, r := range s {
		if caNameSafe(r) {
			b.WriteRune(r)
			lastUnsafe = false
			continue
		}
		if !lastUnsafe {
			b.WriteByte('_')
			lastUnsafe = true
		}
	}
	return strings.Trim(b.String(), "_")
}

// DeriveCAName builds the CA name for a certificate found under a directory: the
// sanitised last element of the directory, an underscore, and the sanitised filename
// with its extension (operator, 2026-09-14).
//
// The DIRECTORY component is what keeps two certificates apart when the same filename
// appears under two of the configured dirs -- `ca.pem` under `prod-cas` and under
// `partner-cas` are different certificates and must not resolve to one CA. It is not
// decoration, so it is never the part dropped when the name is too long.
//
// OVERFLOW. A derived name is usually far inside CANameMax, so the shortening below is
// the exception rather than the rule. When it is needed, the file STEM is what gives
// way: the directory component is the thing that disambiguates, and the extension is
// what makes the name recognisable as a certificate. The stem keeps a leading slice of
// itself rather than being replaced by a hash, so the result still reads as the file it
// came from. Two files whose shortened names collide are a DUPLICATE, reported by the
// caller as one -- that is the same outcome as any other duplicate and is why nothing
// here invents a numeric suffix to force uniqueness: a name nobody chose, silently
// pointing at one of two certificates, is worse than a refusal that names both paths.
//
// dir is the directory as configured, file the base name of a file directly inside it.
func DeriveCAName(dir, file string) (string, error) {
	last := sanitiseCAPart(lastPathElement(dir))
	name := sanitiseCAPart(file)
	if last == "" {
		return "", fmt.Errorf("the directory %q has no name this tool can use in a certificate "+
			"authority name: after removing the characters a CA name may not contain (only "+
			"letters, digits, underscore, dash and period are allowed) nothing is left. Rename "+
			"the directory, or name this certificate explicitly under broker.domainCerts.files", dir)
	}
	if name == "" {
		return "", fmt.Errorf("the file %q in %q has no name this tool can use in a certificate "+
			"authority name: after removing the characters a CA name may not contain (only "+
			"letters, digits, underscore, dash and period are allowed) nothing is left. Rename "+
			"the file, or name it explicitly under broker.domainCerts.files", file, dir)
	}

	full := last + "_" + name
	if len(full) <= CANameMax {
		return full, nil
	}

	// Too long: shorten the STEM, keeping the directory and the extension whole.
	stem, ext := splitCAExt(name)
	budget := CANameMax - len(last) - 1 - len(ext) // 1 for the joining underscore
	if budget < 1 {
		return "", fmt.Errorf("the certificate authority name for %q in %q cannot be shortened to "+
			"%d characters: the directory name and file extension alone need %d. Use a shorter "+
			"directory name, or name this certificate explicitly under broker.domainCerts.files",
			file, dir, CANameMax, len(last)+1+len(ext))
	}
	if budget > len(stem) {
		budget = len(stem)
	}
	// Trim any underscore the cut exposed, so the join does not read as a double
	// separator and two names differing only past the cut do not both end in "_".
	return last + "_" + strings.TrimRight(stem[:budget], "_") + ext, nil
}

// splitCAExt splits a sanitised name into stem and extension, where the extension is
// the final dot and what follows it.
//
// A leading dot is NOT an extension: `.hidden` is a name, not an empty stem with a
// `.hidden` extension, and treating it as the latter would leave nothing to shorten. A
// name with no dot at all has an empty extension, which the caller handles naturally.
func splitCAExt(name string) (stem, ext string) {
	i := strings.LastIndexByte(name, '.')
	if i <= 0 {
		return name, ""
	}
	return name[:i], name[i:]
}

// lastPathElement returns the final component of a directory path, splitting on BOTH
// separators and ignoring any trailing one.
//
// It does not use filepath.Base, for the reason BaseName records: filepath is
// OS-dependent, and an env file written on Windows and run on Linux is ordinary here,
// so `C:\certs\prod-cas\` must yield `prod-cas` on either machine.
func lastPathElement(dir string) string {
	d := strings.TrimRight(strings.ReplaceAll(dir, "\\", "/"), "/")
	if d == "" {
		return ""
	}
	return path.Base(d)
}

// defaultCertExts are the file extensions a directory walk treats as certificates when
// an entry does not name its own.
//
// A filter rather than "every file" is the safe default: a certificate directory
// routinely also holds a README, a private key, or an editor's backup, and uploading
// one of those to the broker as a certificate authority fails at best and installs
// something unintended at worst. An entry that needs a different set says so with its
// own fileExt.
var defaultCertExts = []string{".cer", ".crt", ".pem"}

// MatchesCertExt reports whether file carries one of exts, compared case-insensitively
// and tolerating whether the caller wrote the leading dot.
//
// Case-insensitive because a certificate handed over by a Windows tool is as likely to
// be `.CER` as `.cer`, and refusing it would be a papercut with no security value --
// the extension decides only whether this tool OFFERS the file to the broker, and the
// broker validates the content either way.
func MatchesCertExt(file string, exts []string) bool {
	if len(exts) == 0 {
		exts = defaultCertExts
	}
	lower := strings.ToLower(file)
	for _, e := range exts {
		e = strings.ToLower(strings.TrimSpace(e))
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		if strings.HasSuffix(lower, e) {
			return true
		}
	}
	return false
}

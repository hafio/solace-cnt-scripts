// Package convert turns a legacy bash env file -- the pre-Go `bash/env/<name>`
// format sourced by 000-env.sh -- into the unified YAML env file this CLI reads.
//
// It is a one-way migration aid, not a shell interpreter: it understands the
// assignment forms the env files actually use (scalars, indexed arrays,
// associative arrays, `declare`/`export` prefixes, `${VAR}` references to
// earlier assignments) and ignores everything else. Variables it cannot map are
// reported as warnings rather than dropped silently, so a hand-edited env file
// never loses a setting without saying so.
package convert

import (
	"fmt"
	"regexp"
	"strings"
)

// vars is a parsed bash env file: scalar assignments, indexed arrays, and
// associative arrays, keyed by variable name. seen preserves assignment order so
// the unmapped-variable warnings come out in file order.
type vars struct {
	scalar map[string]string
	array  map[string][]string
	assoc  map[string]map[string]string
	seen   []string
	used   map[string]bool
}

func newVars() *vars {
	return &vars{
		scalar: map[string]string{},
		array:  map[string][]string{},
		assoc:  map[string]map[string]string{},
		used:   map[string]bool{},
	}
}

// record notes a name in first-assignment order, and clears any value it held under a
// DIFFERENT type.
//
// A bash file may legitimately re-assign a name as another type (`X=a` then
// `X=(a b)`), and each type lives in its own map here. Without the clear, the stale
// entry survived in the old map and the accessors -- which check one map each -- could
// still answer from the assignment the file had superseded.
func (v *vars) record(name string) {
	_, wasScalar := v.scalar[name]
	_, wasArray := v.array[name]
	_, wasAssoc := v.assoc[name]
	delete(v.scalar, name)
	delete(v.array, name)
	delete(v.assoc, name)
	if wasScalar || wasArray || wasAssoc {
		return // already in seen, in its original position
	}
	v.seen = append(v.seen, name)
}

// s returns a scalar and marks the variable as mapped.
func (v *vars) s(name string) string {
	v.used[name] = true
	return v.scalar[name]
}

// l returns an indexed array and marks the variable as mapped. A scalar
// assignment is accepted as a one-element list, which is how a single-entry
// bash array is sometimes written.
func (v *vars) l(name string) []string {
	v.used[name] = true
	if a, ok := v.array[name]; ok {
		return a
	}
	if s, ok := v.scalar[name]; ok && s != "" {
		return []string{s}
	}
	return nil
}

// m returns an associative array and marks the variable as mapped.
func (v *vars) m(name string) map[string]string {
	v.used[name] = true
	return v.assoc[name]
}

// has reports whether the variable was assigned a non-empty value. It does not
// mark it as mapped -- callers use it only for platform detection.
func (v *vars) has(name string) bool {
	if s, ok := v.scalar[name]; ok && s != "" {
		return true
	}
	if a, ok := v.array[name]; ok && len(a) > 0 {
		return true
	}
	_, ok := v.assoc[name]
	return ok
}

// ignore marks a variable as mapped without reading it, for bash-only plumbing
// that has no YAML equivalent.
func (v *vars) ignore(names ...string) {
	for _, n := range names {
		v.used[n] = true
	}
}

// unmapped lists, in file order, every assigned variable no mapping consumed.
func (v *vars) unmapped() []string {
	var out []string
	for _, n := range v.seen {
		if !v.used[n] {
			out = append(out, n)
		}
	}
	return out
}

var (
	assignRE = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)=(.*)$`)
	assocRE  = regexp.MustCompile(`^\[(.+)\]=(.*)$`)
	refRE    = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)
)

// parse reads a bash env file into vars. Lines that are not assignments
// (functions, conditionals, blank lines, comments) are skipped: an env file is
// declarations, and anything else is bash the YAML schema has no place for.
func parse(src string) (*vars, error) {
	v := newVars()
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")

	for i := 0; i < len(lines); i++ {
		line, isAssoc := stripDecl(strings.TrimSpace(lines[i]))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		m := assignRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		name, rhs := m[1], m[2]

		if !strings.HasPrefix(rhs, "(") {
			v.record(name)
			v.scalar[name] = v.expandSegments(firstTokenSegments(rhs))
			continue
		}

		// An array body may span lines; accumulate until the closing paren.
		body := rhs[1:]
		for closeIdx(body) < 0 {
			i++
			if i >= len(lines) {
				return nil, fmt.Errorf("unterminated array assignment for %s: no closing ')'", name)
			}
			body += "\n" + lines[i]
		}
		body = body[:closeIdx(body)]

		words := tokenizeSegments(body)
		toks := make([]string, len(words))
		for j, w := range words {
			toks[j] = v.expandSegments(w)
		}
		v.record(name)
		if isAssoc || allAssocEntries(toks) {
			entries := map[string]string{}
			for _, t := range toks {
				e := assocRE.FindStringSubmatch(t)
				if e == nil {
					// A declared associative array whose entry is not `[key]=value` --
					// a typo'd bracket, a missing '=', a stray bare word. Silently
					// dropping it lost a real setting from the converted file with
					// nothing said, which is the opposite of what this converter
					// promises: it reads a file once, and every unconvertible thing in
					// it has to be named.
					return nil, fmt.Errorf("%s: entry %q is not the `[key]=value` form an associative "+
						"array takes -- fix it in the source file, or remove it", name, t)
				}
				entries[e[1]] = e[2]
			}
			v.assoc[name] = entries
			continue
		}
		v.array[name] = toks
	}
	return v, nil
}

// stripDecl removes any `declare`/`export`/`local` prefixes, reporting whether
// the declaration was an associative array (`declare -A`).
func stripDecl(line string) (string, bool) {
	assoc := false
	for {
		switch {
		case strings.HasPrefix(line, "declare -A "):
			line, assoc = strings.TrimSpace(line[len("declare -A "):]), true
		case strings.HasPrefix(line, "declare -a "):
			line = strings.TrimSpace(line[len("declare -a "):])
		case strings.HasPrefix(line, "declare "):
			line = strings.TrimSpace(line[len("declare "):])
		case strings.HasPrefix(line, "export "):
			line = strings.TrimSpace(line[len("export "):])
		case strings.HasPrefix(line, "local "):
			line = strings.TrimSpace(line[len("local "):])
		default:
			return line, assoc
		}
	}
}

// allAssocEntries reports whether every token has the `[key]=value` shape, which
// is how an associative array reads even without the `declare -A` prefix.
func allAssocEntries(toks []string) bool {
	if len(toks) == 0 {
		return false
	}
	for _, t := range toks {
		if !assocRE.MatchString(t) {
			return false
		}
	}
	return true
}

// closeIdx returns the index of the first `)` outside quotes and outside a
// comment, or -1 when the body is not yet terminated.
func closeIdx(s string) int {
	var q rune
	rs := []rune(s)
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		switch {
		case q != 0:
			if r == q {
				q = 0
			}
		case r == '\'' || r == '"':
			q = r
		case r == '#':
			for i < len(rs) && rs[i] != '\n' {
				i++
			}
		case r == ')':
			return i
		}
	}
	return -1
}

// segment is one contiguous run of a tokenized word that shares a quoting
// style. Bash expands `$VAR` inside double quotes and in bare text, but never
// inside single quotes, and a single word can mix both (`a'$b'"$c"`), so the
// "does $ expand here" fact has to travel with a run of text shorter than the
// whole word. literal marks a single-quoted run; expand() only substitutes
// into the non-literal ones, which is what stops a single-quoted secret like
// 'p$s3cret' from being corrupted by a $-reference it never asked for (B5).
type segment struct {
	text    string
	literal bool
}

// tokenizeSegments splits a bash word list into words, honoring single and
// double quotes and dropping `#` comments. Adjacent quoted and bare chunks
// join into one word, so `[CA-NAME]="cert.pem"` stays a single token; unlike
// the plain-string tokenizer this replaced, each word comes back as an
// ordered list of segments so expand() can skip the single-quoted ones
// instead of running over the whole (quote-stripped) word.
func tokenizeSegments(s string) [][]segment {
	var (
		words   [][]segment
		curWord []segment
		cur     strings.Builder
		curLit  bool
		active  bool // cur holds a segment, possibly still empty
		inWord  bool
		q       rune
	)
	flushSeg := func() {
		if active {
			curWord = append(curWord, segment{text: cur.String(), literal: curLit})
			cur.Reset()
			active = false
		}
	}
	flushWord := func() {
		flushSeg()
		if inWord {
			if curWord == nil {
				curWord = []segment{}
			}
			words = append(words, curWord)
			curWord = nil
			inWord = false
		}
	}
	rs := []rune(s)
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		switch {
		case q != 0:
			if r == q {
				q = 0
				continue
			}
			// Inside double quotes bash honours a backslash before exactly $, `,
			// " and \, and leaves it LITERAL before anything else. Both halves
			// matter -- stripping it unconditionally breaks two things:
			//
			//   - "C:\Users\me" lost its separators, silently corrupting any
			//     Windows path a legacy env file happened to double-quote;
			//   - "\$SECRET" produced a bare $SECRET in an expandable segment,
			//     so expand() substituted it and truncated the value -- the same
			//     silent secret loss the single-quote fix above exists to stop.
			//     The escape is written precisely to prevent that expansion, so
			//     the character it produces is emitted as its own LITERAL segment
			//     and can never reach the $VAR substitution.
			if q == '"' && r == '\\' && i+1 < len(rs) && strings.ContainsRune("$`\"\\", rs[i+1]) {
				i++
				flushSeg()
				curWord = append(curWord, segment{text: string(rs[i]), literal: true})
				curLit, active, inWord = false, true, true
				continue
			}
			cur.WriteRune(r)
			inWord = true
		case r == '\'' || r == '"':
			flushSeg() // the run before it may have had different literalness
			q = r
			curLit = r == '\''
			active = true // an empty "" or '' is still a word
			inWord = true
		case r == '#':
			flushWord()
			for i < len(rs) && rs[i] != '\n' {
				i++
			}
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			flushWord()
		default:
			if !active || curLit {
				flushSeg()
				curLit = false
				active = true
			}
			cur.WriteRune(r)
			inWord = true
		}
	}
	flushWord()
	return words
}

// firstTokenSegments returns the segments of a scalar right-hand side's first
// word, dropping any trailing comment (tokenizeSegments already stops a word
// at `#`). No words at all yields nil, which expandSegments renders as "".
func firstTokenSegments(rhs string) []segment {
	words := tokenizeSegments(rhs)
	if len(words) == 0 {
		return nil
	}
	return words[0]
}

// expand substitutes `$VAR` / `${VAR}` with an earlier scalar assignment, which
// is how the bash env files reference e.g. ${SOLBK_NS}. An unknown name expands
// to empty, exactly as bash would. It only ever sees expandable text --
// expandSegments is what keeps single-quoted segments away from it.
func (v *vars) expand(s string) string {
	if !strings.ContainsRune(s, '$') {
		return s
	}
	return refRE.ReplaceAllStringFunc(s, func(ref string) string {
		return v.scalar[strings.Trim(ref, "${}")]
	})
}

// expandSegments joins a word's segments into its final value, running expand
// over everything except the single-quoted runs, which are copied verbatim.
// That is the fix for B5: the old code discarded quoting before expand() ever
// ran, so a single-quoted PSK or password containing `$` was silently
// corrupted the moment it referenced (or merely resembled) a `$VAR` bash would
// have left alone.
func (v *vars) expandSegments(segs []segment) string {
	var b strings.Builder
	for _, seg := range segs {
		if seg.literal {
			b.WriteString(seg.text)
		} else {
			b.WriteString(v.expand(seg.text))
		}
	}
	return b.String()
}

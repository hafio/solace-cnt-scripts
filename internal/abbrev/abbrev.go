// Package abbrev owns the rules for every short form solace-util accepts.
//
// The tool takes abbreviations in four unrelated places -- command names
// (`rm br`), node roles (`p`), --platform values (`dk`), and flag shorthands
// (`-e`) -- and nothing rewrites argv: each one is expanded by whatever already
// owns that input, at the moment it reads it. What was missing was a single
// definition of what makes a short form legal, so each site invented its own
// and described itself in a hand-written sentence beside the table it was
// supposed to describe.
//
// This package is that definition. It owns the RULES, not the vocabulary: a
// Set is declared by the package whose words they are (config declares roles
// and platforms, cli declares command names), because a platform's canonical
// spelling is simultaneously the env file's section key and a command's is the
// cobra tree's. There is no registry and no init-time side effect -- a caller
// that needs several sets, such as the docs generator, names them itself.
//
// It imports nothing but the stdlib, which is what lets both config and cli
// use it -- the same leaf shape as internal/output.
package abbrev

import (
	"fmt"
	"strings"
)

// Entry is one canonical word and the short forms approved for it.
//
// Short may be empty, and that is the point of having an entry at all for a
// word with no abbreviation: a deliberate absence is then data -- it prints in
// the reference and a test can pin it -- rather than something a reader has to
// notice is missing. Note says why, for the absences and for the short forms
// whose safety is not self-evident.
type Entry struct {
	Canonical string
	Short     []string
	Note      string
}

// clone copies an Entry so a caller cannot reach into the Set's own slice.
func (e Entry) clone() Entry {
	out := e
	if e.Short != nil {
		out.Short = append([]string(nil), e.Short...)
	}
	return out
}

// Set is one namespace of approved short forms, in declaration order.
//
// Entries keep the order they were declared in rather than being sorted: that
// order is meaningful in every set here (redundancy order for roles, the order
// platforms are offered in), and it is what errors and the reference print.
type Set struct {
	kind    string
	entries []Entry
	expand  map[string]string
}

// New builds a Set, and panics if the declaration is not one it can honour.
//
// A panic is the house answer for this class: applyAliases, registerFlagCompletion
// and flagOnlyOn already treat their own wiring as a construction-time bug rather
// than a runtime condition, and failing here means every test that builds a Set is
// also a validation test. The rules are:
//
//   - a canonical name is non-empty and declared once;
//   - a short form is lowercase ASCII letters or digits, so it cannot carry a
//     shell metacharacter, and cannot be a second spelling that differs only in case;
//   - a short form is shorter than the word it stands for, because one that is not
//     is a synonym, and a synonym puts a second name for the same thing into
//     circulation;
//   - a short form is claimed once, and never shadows a canonical name in the
//     same set -- the word would then mean two things depending on which lookup
//     ran first.
func New(kind string, entries []Entry) *Set {
	if kind == "" {
		panic("abbrev.New: kind is empty")
	}
	s := &Set{
		kind:    kind,
		entries: make([]Entry, 0, len(entries)),
		expand:  make(map[string]string, len(entries)*2),
	}
	// Canonicals are claimed in a pass of their own, so "a short form may not
	// shadow a canonical" holds whatever order the entries are declared in.
	for _, e := range entries {
		if e.Canonical == "" {
			panic(fmt.Sprintf("abbrev %q: an entry has no canonical name", kind))
		}
		if _, dup := s.expand[e.Canonical]; dup {
			panic(fmt.Sprintf("abbrev %q: %q is declared twice", kind, e.Canonical))
		}
		s.expand[e.Canonical] = e.Canonical
	}
	for _, e := range entries {
		for _, short := range e.Short {
			checkShort(kind, e.Canonical, short)
			if owner, dup := s.expand[short]; dup {
				if owner == short {
					panic(fmt.Sprintf("abbrev %q: %q is a short form for %q and also a %s in its own right",
						kind, short, e.Canonical, kind))
				}
				panic(fmt.Sprintf("abbrev %q: %q is claimed by both %q and %q", kind, short, owner, e.Canonical))
			}
			s.expand[short] = e.Canonical
		}
		s.entries = append(s.entries, e.clone())
	}
	return s
}

// checkShort enforces the charset and the shorter-than rule on one short form.
func checkShort(kind, canonical, short string) {
	if short == "" {
		panic(fmt.Sprintf("abbrev %q: %q has an empty short form", kind, canonical))
	}
	if len(short) >= len(canonical) {
		panic(fmt.Sprintf("abbrev %q: short form %q is not shorter than %q", kind, short, canonical))
	}
	for _, r := range short {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			panic(fmt.Sprintf("abbrev %q: short form %q for %q must be lowercase letters or digits",
				kind, short, canonical))
		}
	}
}

// Kind names what this set abbreviates, for errors and section headings.
func (s *Set) Kind() string { return s.kind }

// Expand resolves a word to its canonical spelling. A canonical name resolves
// to itself, so a caller has one lookup rather than a membership test and a
// lookup, and anything else is reported unknown for the caller to phrase.
func (s *Set) Expand(word string) (string, bool) {
	c, ok := s.expand[word]
	return c, ok
}

// Short returns the short forms approved for a canonical name, or nil.
func (s *Set) Short(canonical string) []string {
	for _, e := range s.entries {
		if e.Canonical == canonical {
			return append([]string(nil), e.Short...)
		}
	}
	return nil
}

// Entries returns every entry in declaration order.
func (s *Set) Entries() []Entry {
	out := make([]Entry, len(s.entries))
	for i, e := range s.entries {
		out[i] = e.clone()
	}
	return out
}

// Names returns the canonical spellings -- what completion offers and errors teach.
func (s *Set) Names() []string {
	out := make([]string, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e.Canonical)
	}
	return out
}

// Shorts returns every approved short form, flattened, in declaration order.
func (s *Set) Shorts() []string {
	out := make([]string, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e.Short...)
	}
	return out
}

// List renders the set for an error message -- "kubernetes (kube), docker (dk)
// or podman (pm)" -- with each short form beside the word it stands for rather
// than as a second list to cross-reference.
func (s *Set) List() string {
	parts := make([]string, 0, len(s.entries))
	for _, e := range s.entries {
		if len(e.Short) == 0 {
			parts = append(parts, e.Canonical)
			continue
		}
		parts = append(parts, fmt.Sprintf("%s (%s)", e.Canonical, strings.Join(e.Short, ", ")))
	}
	if len(parts) < 2 {
		return strings.Join(parts, "")
	}
	return strings.Join(parts[:len(parts)-1], ", ") + " or " + parts[len(parts)-1]
}

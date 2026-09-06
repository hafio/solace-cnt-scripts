package abbrev

import (
	"strings"
	"testing"
)

// sample is the shape every set in the tool has: canonical words in a meaningful
// order, one short form each, one word deliberately carrying none.
func sample() *Set {
	return New("thing", []Entry{
		{Canonical: "kubernetes", Short: []string{"kube"}, Note: "the section key"},
		{Canonical: "docker", Short: []string{"dk"}},
		{Canonical: "podman", Short: []string{"pm"}},
		{Canonical: "start", Note: "no short form: ambiguous with stop"},
	})
}

// TestExpandResolvesBothSpellings pins the one lookup callers get: a canonical
// word resolves to itself, so a parser needs a membership test and a lookup no
// longer -- ParseRole and ParsePlatform both rely on that to stay a single branch.
func TestExpandResolvesBothSpellings(t *testing.T) {
	s := sample()
	for _, tc := range []struct {
		in   string
		want string
		ok   bool
	}{
		{in: "kubernetes", want: "kubernetes", ok: true},
		{in: "kube", want: "kubernetes", ok: true},
		{in: "dk", want: "docker", ok: true},
		{in: "start", want: "start", ok: true},
		{in: "k8s"},
		{in: "KUBE"},
		{in: " docker"},
		{in: ""},
	} {
		got, ok := s.Expand(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Errorf("Expand(%q) = %q, %v; want %q, %v", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

// TestOrderIsDeclarationOrder pins that nothing sorts the entries. The order is
// meaningful in every set the tool declares -- redundancy order for roles, the
// order platforms are offered in -- and it is the order errors and the reference
// print, so an alphabetical Names() would silently reword both.
func TestOrderIsDeclarationOrder(t *testing.T) {
	s := sample()
	wantNames := []string{"kubernetes", "docker", "podman", "start"}
	if got := s.Names(); strings.Join(got, ",") != strings.Join(wantNames, ",") {
		t.Errorf("Names() = %v, want %v", got, wantNames)
	}
	wantShorts := []string{"kube", "dk", "pm"}
	if got := s.Shorts(); strings.Join(got, ",") != strings.Join(wantShorts, ",") {
		t.Errorf("Shorts() = %v, want %v", got, wantShorts)
	}
	entries := s.Entries()
	if len(entries) != 4 {
		t.Fatalf("Entries() returned %d entries, want 4", len(entries))
	}
	if entries[0].Note != "the section key" {
		t.Errorf("Entries()[0].Note = %q, want the declared note", entries[0].Note)
	}
	if len(entries[3].Short) != 0 {
		t.Errorf("Entries()[3].Short = %v, want none: a word may be declared with no short form", entries[3].Short)
	}
}

// TestAccessorsReturnCopies stops a caller reaching into the set through a slice
// it was handed. The sets are package-level and shared by the parsers, the CLI
// wiring and the docs generator, so a mutation here would be a global one.
func TestAccessorsReturnCopies(t *testing.T) {
	s := sample()
	s.Entries()[0].Canonical = "hijacked"
	s.Entries()[0].Short[0] = "hijacked"
	s.Short("kubernetes")[0] = "hijacked"
	s.Names()[0] = "hijacked"
	if got, ok := s.Expand("kube"); !ok || got != "kubernetes" {
		t.Errorf("Expand(\"kube\") = %q, %v after mutating returned slices; want %q, true", got, ok, "kubernetes")
	}
	if got := s.Short("kubernetes"); len(got) != 1 || got[0] != "kube" {
		t.Errorf("Short(\"kubernetes\") = %v, want [kube]", got)
	}
}

// TestShortOfUnknownWord is what applyAliases relies on: it asks the set about
// every command in the tree, most of which have no entry, and must get nothing
// back rather than a panic.
func TestShortOfUnknownWord(t *testing.T) {
	if got := sample().Short("nothing-like-this"); got != nil {
		t.Errorf("Short of an undeclared word = %v, want nil", got)
	}
}

// TestList pins the error sentence exactly, because ParsePlatform's message is
// this rendering and TestParsePlatformErrorTeachesCanonical reads it back.
func TestList(t *testing.T) {
	for _, tc := range []struct {
		name    string
		entries []Entry
		want    string
	}{
		{
			name:    "three with short forms",
			entries: sample().Entries()[:3],
			want:    "kubernetes (kube), docker (dk) or podman (pm)",
		},
		{
			name:    "a word with no short form prints bare",
			entries: []Entry{{Canonical: "start"}, {Canonical: "stop"}},
			want:    "start or stop",
		},
		{
			name:    "one entry needs no conjunction",
			entries: []Entry{{Canonical: "docker", Short: []string{"dk"}}},
			want:    "docker (dk)",
		},
		{
			name:    "empty set renders nothing",
			entries: nil,
			want:    "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := New("thing", tc.entries).List(); got != tc.want {
				t.Errorf("List() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestKind(t *testing.T) {
	if got := sample().Kind(); got != "thing" {
		t.Errorf("Kind() = %q, want %q", got, "thing")
	}
}

// TestNewPanicsOnAWiringBug is the whole reason this package exists: every rule
// that makes a short form safe is checked once, here, at construction. A panic
// rather than an error is deliberate -- these are declarations in source, so the
// failure belongs at the moment the set is built, which makes every test that
// builds one a validation test.
func TestNewPanicsOnAWiringBug(t *testing.T) {
	for _, tc := range []struct {
		name    string
		kind    string
		entries []Entry
		want    string
	}{
		{
			name: "no kind",
			want: "kind is empty",
		},
		{
			name:    "empty canonical",
			kind:    "thing",
			entries: []Entry{{Canonical: ""}},
			want:    "no canonical name",
		},
		{
			name:    "canonical declared twice",
			kind:    "thing",
			entries: []Entry{{Canonical: "docker"}, {Canonical: "docker"}},
			want:    "declared twice",
		},
		{
			name: "one short form claimed by two words",
			kind: "thing",
			entries: []Entry{
				{Canonical: "docker", Short: []string{"dk"}},
				{Canonical: "darkroom", Short: []string{"dk"}},
			},
			want: "claimed by both",
		},
		{
			name: "short form shadows a canonical name declared later",
			kind: "thing",
			entries: []Entry{
				{Canonical: "docker", Short: []string{"pm"}},
				{Canonical: "pm"},
			},
			want: "in its own right",
		},
		{
			name:    "empty short form",
			kind:    "thing",
			entries: []Entry{{Canonical: "docker", Short: []string{""}}},
			want:    "empty short form",
		},
		{
			name:    "short form is not shorter",
			kind:    "thing",
			entries: []Entry{{Canonical: "docker", Short: []string{"dockers"}}},
			want:    "not shorter",
		},
		{
			name:    "short form the same length is a synonym",
			kind:    "thing",
			entries: []Entry{{Canonical: "docker", Short: []string{"moby42"}}},
			want:    "not shorter",
		},
		{
			name:    "uppercase short form",
			kind:    "thing",
			entries: []Entry{{Canonical: "docker", Short: []string{"DK"}}},
			want:    "lowercase letters or digits",
		},
		{
			name:    "short form carrying a shell metacharacter",
			kind:    "thing",
			entries: []Entry{{Canonical: "docker", Short: []string{"d;k"}}},
			want:    "lowercase letters or digits",
		},
		{
			name:    "short form with a space",
			kind:    "thing",
			entries: []Entry{{Canonical: "docker", Short: []string{"d k"}}},
			want:    "lowercase letters or digits",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("New(%q, %v) did not panic, want a panic naming %q", tc.kind, tc.entries, tc.want)
				}
				if msg, _ := r.(string); !strings.Contains(msg, tc.want) {
					t.Errorf("panic = %q, want it to name %q", r, tc.want)
				}
			}()
			New(tc.kind, tc.entries)
		})
	}
}

// TestDigitsAreAllowed keeps the charset from being letters-only by accident: a
// product word that ends in a number is exactly the case an abbreviation is worth
// having for, and refusing it would be a rule nobody chose.
func TestDigitsAreAllowed(t *testing.T) {
	s := New("thing", []Entry{{Canonical: "kubernetes", Short: []string{"k8s"}}})
	if got, ok := s.Expand("k8s"); !ok || got != "kubernetes" {
		t.Errorf("Expand(\"k8s\") = %q, %v; want %q, true", got, ok, "kubernetes")
	}
}

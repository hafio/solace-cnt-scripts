package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// The abbreviations are a convenience layer over a tree that is already settled, so
// what these tests protect is not the individual words but the two properties that
// make the words safe to hand out: an alias reaches exactly the command its long
// form does, and no alias can ever be ambiguous with a sibling.

// TestAliasesResolveToTheCanonicalCommand pins equivalence by resolution rather than
// by running anything: cobra's Find is what dispatch itself uses, so proving both
// paths land on the same *cobra.Command proves they behave identically without
// needing a cluster, an env file, or a runner.
//
// Every row abbreviates BOTH halves (`br rm`, not `br remove`), which is the case worth
// pinning: commandAbbrev is keyed by NAME and applied by a tree walk, so a word means the
// same thing at whatever depth it appears. `rm` under `broker` and `rm` under `operator`
// are two different commands reached by one abbreviation, and nothing in the table itself
// says so -- the walk is what makes it true.
func TestAliasesResolveToTheCanonicalCommand(t *testing.T) {
	for _, tc := range []struct{ alias, canonical []string }{
		{[]string{"br", "rm"}, []string{"broker", "remove"}},
		{[]string{"op", "rm"}, []string{"operator", "remove"}},
		{[]string{"br", "dp"}, []string{"broker", "deploy"}},
		{[]string{"op", "dp"}, []string{"operator", "deploy"}},
		{[]string{"br", "sts"}, []string{"broker", "status"}},
		{[]string{"op", "sts"}, []string{"operator", "status"}},
		{[]string{"br", "lg"}, []string{"broker", "logs"}},
		{[]string{"br", "vld"}, []string{"broker", "validate"}},
		{[]string{"op", "vld"}, []string{"operator", "validate"}},
		{[]string{"vld"}, []string{"validate"}},
		{[]string{"br", "pf", "assert-leader"}, []string{"broker", "perform", "assert-leader"}},
		{[]string{"br", "pf", "gd"}, []string{"broker", "perform", "gather-diagnostics"}},
		{[]string{"br", "cfg", "server-certs"}, []string{"broker", "configure", "server-certs"}},
		{[]string{"br", "rs"}, []string{"broker", "restart"}},
		{[]string{"br", "gen"}, []string{"broker", "generate"}},
		{[]string{"op", "gen"}, []string{"operator", "generate"}},
		{[]string{"cv"}, []string{"convert"}},
		{[]string{"ver"}, []string{"version"}},
		{[]string{"br", "sh"}, []string{"broker", "shell"}},
		{[]string{"br", "cp", "from"}, []string{"broker", "copy", "from"}},
	} {
		t.Run(strings.Join(tc.alias, " "), func(t *testing.T) {
			root := newRootCmd(&App{})
			viaAlias, _, err := root.Find(tc.alias)
			if err != nil {
				t.Fatalf("resolving %v: %v", tc.alias, err)
			}
			root2 := newRootCmd(&App{})
			viaName, _, err := root2.Find(tc.canonical)
			if err != nil {
				t.Fatalf("resolving %v: %v", tc.canonical, err)
			}
			if viaAlias.CommandPath() != viaName.CommandPath() {
				t.Errorf("%v resolved to %q, want the same command as %v (%q)",
					tc.alias, viaAlias.CommandPath(), tc.canonical, viaName.CommandPath())
			}
		})
	}
}

// TestAliasesDoNotCollide walks the real tree and proves no two siblings answer to
// the same word. applyAliases panics on a collision at construction time, so this
// is the second line of defence -- and the one that would still catch a collision
// introduced by a command's own hand-written Aliases rather than by the table.
func TestAliasesDoNotCollide(t *testing.T) {
	root := newRootCmd(&App{})
	var walk func(*cobra.Command)
	walk = func(parent *cobra.Command) {
		seen := map[string]string{}
		for _, c := range parent.Commands() {
			for _, word := range append([]string{c.Name()}, c.Aliases...) {
				if owner, dup := seen[word]; dup {
					t.Errorf("under %q, %q is claimed by both %q and %q",
						parent.CommandPath(), word, owner, c.Name())
				}
				seen[word] = c.Name()
			}
			walk(c)
		}
	}
	walk(root)
}

// TestEveryAliasEntryIsLive catches the quiet failure mode of a name-keyed table:
// a command gets renamed, its entry stops matching anything, and the abbreviation
// silently disappears while the table still claims to provide it.
func TestEveryAliasEntryIsLive(t *testing.T) {
	root := newRootCmd(&App{})
	live := map[string]bool{}
	var walk func(*cobra.Command)
	walk = func(parent *cobra.Command) {
		for _, c := range parent.Commands() {
			live[c.Name()] = true
			walk(c)
		}
	}
	walk(root)
	for _, name := range commandAbbrev.Names() {
		if !live[name] {
			t.Errorf("commandAbbrev has an entry for %q, which is not a command in the tree", name)
		}
	}
}

// TestNounGroupsRunNothing is the inverted form of what TestDangerousVerbsHaveNoBareAlias
// used to pin, and it is why the abbreviations stay safe to hand out.
//
// The tree used to be verb-first, so the danger was `rm` acting on its own and the guard was
// "a verb that owns objects has no RunE". It is noun-first now: `rm` is not reachable at the
// top level at all, and the two-token `br rm` IS the destructive command. What has to stay
// true is the same shape one level up -- the NOUN runs nothing, so `br` and `op` alone act
// on nothing, and the families beneath them (copy, configure, perform) are groups too.
//
// A RunE added to any of these later is what this catches. It also catches the subtler one:
// a group that lost its last subcommand still passes "has no RunE" while being useless, so
// HasSubCommands is checked as well.
func TestNounGroupsRunNothing(t *testing.T) {
	root := newRootCmd(&App{})
	for _, path := range [][]string{
		{"broker"},
		{"operator"},
		{"broker", "copy"},
		{"broker", "configure"},
		{"broker", "perform"},
	} {
		name := strings.Join(path, " ")
		c, _, err := root.Find(path)
		if err != nil {
			t.Fatalf("finding %q: %v", name, err)
		}
		if c.Annotations[groupAnnotation] != "true" {
			t.Errorf("%q is not a group; a noun that owns verbs must not act on its own", name)
		}
		if c.RunE != nil && !c.HasSubCommands() {
			t.Errorf("%q runs something and owns nothing", name)
		}
		if !c.HasSubCommands() {
			t.Errorf("%q has no subcommands, so there is no verb to name", name)
		}
	}
}

// TestGroupsRejectAnUnknownVerb pins the reason groups are runnable at all. Cobra answers
// a NON-runnable command by printing help and exiting 0 whatever arguments it got, so a
// mistyped verb on a destructive noun would report success having done nothing. Bare still
// prints help and succeeds; a word the noun does not know fails.
//
// The groups are the nouns now, so the mistyped word is the VERB: `broker remvoe` is the
// slip this catches, and reporting success for it is exactly as bad as it was when the
// mistyped word was the noun.
func TestGroupsRejectAnUnknownVerb(t *testing.T) {
	for _, group := range [][]string{
		{"broker"}, {"operator"}, {"broker", "copy"}, {"broker", "configure"}, {"broker", "perform"},
	} {
		name := strings.Join(group, " ")
		t.Run(name, func(t *testing.T) {
			if _, err := runRoot(t, group); err != nil {
				t.Errorf("bare %q err = %v, want help and success", name, err)
			}
			if _, err := runRoot(t, append(append([]string{}, group...), "bogus")); err == nil {
				t.Errorf("%q bogus err = nil, want a loud refusal naming the unknown word", name)
			}
		})
	}
}

// TestStartStopHaveNoAlias pins the one deliberate omission. `st` could mean start,
// stop or status, and the cost of guessing wrong between the first two is an
// outage, so none of them gets a two-letter form that could be confused.
//
// Both are declared in the set with no short form rather than left out of it, so
// the absence is checkable here and printed in docs/abbreviation.md instead of
// being an omission a reader has to spot. The recorded note is what a later reader
// finds when they wonder why -- an entry that lost it would read as an oversight.
func TestStartStopHaveNoAlias(t *testing.T) {
	notes := map[string]string{}
	for _, e := range commandAbbrev.Entries() {
		notes[e.Canonical] = e.Note
	}
	for _, name := range []string{"start", "stop"} {
		note, declared := notes[name]
		if !declared {
			t.Errorf("commandAbbrev has no entry for %q: declare it with no short form so the absence is deliberate", name)
			continue
		}
		if got := commandAbbrev.Short(name); len(got) > 0 {
			t.Errorf("commandAbbrev short forms for %q = %v, want none: it would be ambiguous with the other", name, got)
		}
		if note == "" {
			t.Errorf("commandAbbrev entry for %q has no note saying why it has no short form", name)
		}
	}
	// start/stop are verbs under each noun now, so the tree half of the check has to look
	// under both -- an alias added to only one of them would otherwise slip through.
	root := newRootCmd(&App{})
	for _, path := range [][]string{
		{"broker", "start"}, {"broker", "stop"},
		{"operator", "start"}, {"operator", "stop"},
	} {
		name := strings.Join(path, " ")
		c, _, err := root.Find(path)
		if err != nil {
			t.Fatalf("finding %q: %v", name, err)
		}
		if len(c.Aliases) > 0 {
			t.Errorf("%q has aliases %v, want none", name, c.Aliases)
		}
	}
}

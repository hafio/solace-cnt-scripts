package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"solace/internal/abbrev"
)

// commandAbbrev is the approved short form of every command NAME.
//
// Keyed by name, not by path, and applied by a tree walk: a word means the same
// thing wherever it appears, so `broker` is `br` under every verb that takes it and
// `operator` is `op` under every verb that takes it. That is the property being
// bought -- one abbreviation per concept, not one per command path. The charset,
// the shorter-than rule and the one-claim-per-word rule are abbrev's, shared with
// the role and --platform sets so a short form means the same kind of thing
// whichever of the four the reader has met before.
//
// `rollout`/`rollback` were considered as names and rejected; the verbs that
// replaced them (`deploy all`, `remove all`) inherit `dp`/`rm` from their verb, so
// `dp all` and `rm all` already work. `delete` is deliberately NOT a short form of
// `remove`: one removal word, everywhere.
//
// `start` and `stop` are declared with no short form rather than left out, so the
// absence is data -- it prints in docs/abbreviation.md and TestStartStopHaveNoAlias
// pins it -- instead of something a reader has to notice is missing.
var commandAbbrev = abbrev.New("command", []abbrev.Entry{
	// verbs
	{Canonical: "check", Short: []string{"ck"}},
	{Canonical: "config", Short: []string{"cfg"}},
	{Canonical: "convert", Short: []string{"cv"}},
	{Canonical: "copy", Short: []string{"cp"}},
	{Canonical: "deploy", Short: []string{"dp"}},
	{Canonical: "diagnostics", Short: []string{"diag"}},
	{Canonical: "examples", Short: []string{"eg"},
		Note: "the one noun at the top level: it emits an env file rather than acting on a deployment"},
	{Canonical: "generate", Short: []string{"gen"}},
	{Canonical: "logs", Short: []string{"lg"}},
	{Canonical: "prepare", Short: []string{"pre"}},
	{Canonical: "remove", Short: []string{"rm"},
		Note: "safe to hand out because `remove` takes a noun before it does anything: `rm` alone prints help"},
	{Canonical: "restart", Short: []string{"rs"}},
	{Canonical: "shell", Short: []string{"sh"}},
	{Canonical: "start",
		Note: "no short form: any two-letter form is ambiguous with `stop` and `status`, and that is the one slip that costs an outage"},
	{Canonical: "status", Short: []string{"sts"}},
	{Canonical: "stop",
		Note: "no short form, for the same reason as `start`"},
	{Canonical: "version", Short: []string{"ver"}},

	// nouns -- these ride along under every verb that takes them
	{Canonical: "broker", Short: []string{"br"}, Note: "works under every verb that takes a broker"},
	{Canonical: "operator", Short: []string{"op"}, Note: "works under every verb that takes the operator"},
})

// applyAliases attaches every approved short form to the tree, and panics if one
// would collide with a sibling's name or with another sibling's alias.
//
// The panic is the point: a collision is a wiring bug in this package, not a
// runtime condition, and cobra resolves a duplicate silently by taking whichever
// command it reaches first. Failing at construction means every test that builds
// the tree is also a collision test. It matches how registerFlagCompletion and
// flagOnlyOn already treat their own wiring bugs.
func applyAliases(root *cobra.Command) {
	var walk func(*cobra.Command)
	walk = func(parent *cobra.Command) {
		// taken maps every word already claimed among these siblings to the command
		// that claimed it, so the panic can name both sides of a collision.
		taken := map[string]string{}
		claim := func(c *cobra.Command, word string) {
			if owner, dup := taken[word]; dup {
				panic(fmt.Sprintf("alias collision under %q: %q is claimed by both %q and %q",
					parent.CommandPath(), word, owner, c.Name()))
			}
			taken[word] = c.Name()
		}
		for _, c := range parent.Commands() {
			claim(c, c.Name())
			for _, a := range c.Aliases {
				claim(c, a)
			}
		}
		for _, c := range parent.Commands() {
			for _, a := range commandAbbrev.Short(c.Name()) {
				claim(c, a)
				// Append rather than replace: a command may already carry an alias of
				// its own for a reason this table knows nothing about.
				c.Aliases = append(c.Aliases, a)
			}
			walk(c)
		}
	}
	walk(root)
}

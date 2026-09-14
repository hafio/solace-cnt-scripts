package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"solace/internal/abbrev"
)

// commandAbbrev is the approved short form of every command NAME.
//
// Keyed by name, not by path, and applied by a tree walk: a word means the same thing
// wherever it appears, so `deploy` is `dp` under both nouns and `validate` is `vld` at all
// three places it appears. That is the property being bought -- one abbreviation per
// concept, not one per command path. The charset, the shorter-than rule and the
// one-claim-per-word rule are abbrev's, shared with the role and --platform sets so a short
// form means the same kind of thing whichever of the four the reader has met before.
//
// The tree is noun-first, which is what makes the two-letter verbs safe: `broker` and
// `operator` are groups that run nothing, so `br` and `op` alone destroy nothing and `rm`
// only reaches a removal with a noun already spelled in front of it. `delete` is
// deliberately NOT a short form of `remove`: one removal word, everywhere.
//
// A word with NO short form is DECLARED rather than left out, so the absence is data -- it
// prints in docs/abbreviation.md and a test can pin it -- instead of something a reader has
// to notice is missing. There are two reasons a word has none: ambiguity (`start`/`stop`,
// where any two-letter form collides with each other or with `status`, and that is the one
// slip that costs an outage), or a name already short enough that an abbreviation would
// only be a second thing to remember.
var commandAbbrev = abbrev.New("command", []abbrev.Entry{
	// nouns -- the top level, and the reason the verbs below are safe to shorten
	{Canonical: "broker", Short: []string{"br"},
		Note: "the noun runs nothing on its own, so `br` alone cannot act"},
	{Canonical: "operator", Short: []string{"op"},
		Note: "same: `op` alone prints what the operator can be asked to do"},
	{Canonical: "examples", Short: []string{"eg"},
		Note: "emits an env file rather than acting on a deployment, so it sits at the top level"},
	{Canonical: "convert", Short: []string{"cv"}},
	{Canonical: "version", Short: []string{"ver"}},

	// verbs -- these ride under whichever noun takes them
	{Canonical: "validate", Short: []string{"vld"},
		Note: "works at the top level and under both nouns: the whole env file, or either half"},
	{Canonical: "deploy", Short: []string{"dp"}},
	{Canonical: "remove", Short: []string{"rm"},
		Note: "safe because the noun comes first: `rm` alone is not a command"},
	{Canonical: "generate", Short: []string{"gen"}},
	{Canonical: "restart", Short: []string{"rs"}},
	{Canonical: "status", Short: []string{"sts"}},
	{Canonical: "logs", Short: []string{"lg"}},
	{Canonical: "shell", Short: []string{"sh"}},
	{Canonical: "copy", Short: []string{"cp"}},
	{Canonical: "configure", Short: []string{"cfg"}},
	{Canonical: "perform", Short: []string{"pf"}},
	{Canonical: "start",
		Note: "no short form: any two-letter form is ambiguous with `stop` and `status`, and that is the one slip that costs an outage"},
	{Canonical: "stop",
		Note: "no short form, for the same reason as `start`"},
	{Canonical: "cli",
		Note: "no short form: three letters already, and `cl` would not read as anything"},

	// the two leaves short enough to earn a form of their own
	{Canonical: "gather-diagnostics", Short: []string{"gd"}},
	// `dr` is an INITIALISM of a hyphenated leaf, the same shape as gd, not a synonym
	// for a different word -- which is why `replication` is not the canonical name here.
	// It is also what the operator calls the feature.
	{Canonical: "data-replication", Short: []string{"dr"}},

	// the remaining leaves are spelled out. Each is rare enough that a short form would
	// be a second name to learn rather than a saving, and completion already types them.
	{Canonical: "server-certs", Note: "no short form: completion types it, and `sc` reads as nothing"},
	{Canonical: "domain-certs", Note: "no short form, for the same reason as server-certs"},
	{Canonical: "product-keys", Note: "no short form, for the same reason as server-certs"},
	{Canonical: "default-vpn", Note: "no short form: `dv` would not distinguish it from default-users"},
	{Canonical: "default-users", Note: "no short form, for the same reason as default-vpn"},
	{Canonical: "assert-leader", Note: "no short form: rare, and run deliberately"},
	{Canonical: "redundancy-test", Note: "no short form: invasive, so spelling it out is a feature"},
	{Canonical: "semp-login-check", Note: "no short form: rare, and run deliberately"},
	{Canonical: "cli-script", Note: "no short form: `cs` would not distinguish it from shell-script"},
	{Canonical: "shell-script", Note: "no short form: `sh` is already claimed by `shell`"},
	{Canonical: "export-config", Note: "no short form: completion types it, and `ec` reads as nothing"},
	{Canonical: "import-config",
		Note: "no short form: it tears down and rebuilds a message-VPN, losing the messages " +
			"spooled in its queues, so spelling it out is a feature -- the same reason " +
			"redundancy-test has none"},
	{Canonical: "from", Note: "no short form: it is already one syllable and reads as a direction"},
	{Canonical: "into", Note: "no short form, for the same reason as from"},
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

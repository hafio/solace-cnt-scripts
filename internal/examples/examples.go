// Package examples holds the env-file templates `solace-util examples` prints:
// one minimal starter per platform, plus the complete annotated schema.
//
// They are embedded rather than read from env/ because a released binary has no
// checkout to read from. //go:embed cannot reach outside its own package, so
// assets/full.yaml is the AUTHORITATIVE copy and env/sample.yaml -- which three
// other packages load as their shared fixture, and which every doc points at --
// is generated from it by examples_test.go's -update flag. Editing the sample
// directly fails that gate rather than drifting.
//
// Like internal/abbrev and internal/output this is a stdlib-only leaf, so any
// package can import it. It deliberately does not import internal/config: the
// templates are text, and examples_test.go is what proves the text still loads.
package examples

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed assets/kubernetes.yaml assets/docker.yaml assets/podman.yaml assets/full.yaml
var assets embed.FS

// FullName is the annotated-schema template: what bare `solace-util examples` emits, and
// the one name in this registry that is NOT a platform. Named rather than spelled at the
// call site so the CLI's default and this registry cannot disagree about it.
const FullName = "full"

// Example is one template: its name, the one line the listing shows beside it, and the
// bytes it prints. Every name but FullName is a platform's own section key, so a starter
// and the `docker:` block it emits are the same word rather than two spellings of one.
type Example struct {
	Name  string
	Short string
	Body  []byte
}

// all is declared in the order the listing shows them: the platform starters in
// config.Platforms() order, then the schema. Each starter's name IS the env
// file's platform section key, so `examples docker` and the `docker:` block it
// emits are the same word -- the platform's own name, not a third spelling of it.
var all = []struct {
	name  string
	short string
	file  string
}{
	{"kubernetes", "minimal standalone kubernetes env file", "assets/kubernetes.yaml"},
	{"docker", "minimal standalone docker env file", "assets/docker.yaml"},
	{"podman", "minimal standalone podman env file", "assets/podman.yaml"},
	{FullName, "every key the schema accepts, annotated with each default", "assets/full.yaml"},
}

// All returns every example in listing order.
func All() []Example {
	out := make([]Example, 0, len(all))
	for _, e := range all {
		out = append(out, Example{Name: e.name, Short: e.short, Body: read(e.file)})
	}
	return out
}

// Names returns the canonical names, in listing order.
func Names() []string {
	out := make([]string, 0, len(all))
	for _, e := range all {
		out = append(out, e.name)
	}
	return out
}

// Get returns one example by its exact name. The error names every alternative,
// because a wrong name here is a typo and the fix is one of four words.
func Get(name string) (Example, error) {
	for _, e := range all {
		if e.name == name {
			return Example{Name: e.name, Short: e.short, Body: read(e.file)}, nil
		}
	}
	return Example{}, fmt.Errorf("unknown example %q; available: %s", name, List())
}

// List renders the names for an error or a help string: "a, b, c or d".
func List() string {
	names := Names()
	if len(names) < 2 {
		return strings.Join(names, "")
	}
	return strings.Join(names[:len(names)-1], ", ") + " or " + names[len(names)-1]
}

// read returns the asset's bytes. embed.FS hands back a fresh slice per call, so
// a caller cannot corrupt the next reader's copy. A failure here is a wiring bug
// -- the file is embedded at compile time, so the only way to miss is a name that
// does not match the //go:embed pattern.
func read(file string) []byte {
	b, err := assets.ReadFile(file)
	if err != nil {
		panic(fmt.Sprintf("examples: embedded asset %q: %v", file, err))
	}
	return b
}

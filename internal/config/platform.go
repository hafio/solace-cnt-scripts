package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"solace/internal/abbrev"
)

// Platforms lists every platform, in the order they are offered to a user and
// reported in errors. It exists so no caller has to hand-write the set: a
// hand-written list is what silently goes stale when a platform is added.
func Platforms() []Platform { return []Platform{K8s, Docker, Podman} }

// platformShort names the short --platform spelling for each platform. They are
// abbreviations of a current name, not retired names kept alive: the canonical
// word is still the only thing the env file, the error messages and the
// completions use, so the short form saves typing without putting a second
// spelling of the schema into circulation.
var platformShort = map[Platform]string{
	K8s:    "kube",
	Docker: "dk",
	Podman: "pm",
}

var platformAbbrev = newPlatforms()

// newPlatforms builds the platform set by walking Platforms(), so a platform
// added there is in the set -- with no short form until one is declared --
// rather than quietly missing from it.
func newPlatforms() *abbrev.Set {
	entries := make([]abbrev.Entry, 0, len(Platforms()))
	for _, p := range Platforms() {
		e := abbrev.Entry{Canonical: string(p)}
		if s, ok := platformShort[p]; ok {
			e.Short = []string{s}
		}
		if p == K8s {
			e.Note = "`k8s` and `k8` are refused: the section key is the product's own word"
		}
		entries = append(entries, e)
	}
	return abbrev.New("platform", entries)
}

// PlatformAbbrev returns the approved --platform spellings, for the reference.
func PlatformAbbrev() *abbrev.Set { return platformAbbrev }

// ParsePlatform normalizes a --platform value. An empty string returns an empty
// platform and no error -- it means "not specified", and what that implies is
// the caller's decision (the CLI detects it from the env file; convert detects
// it from the legacy variable names).
//
// The error names the canonical spellings first, since those are what the env
// file's own section keys are, with each abbreviation beside the word it stands
// for rather than as a separate list to cross-reference. It is rendered from the
// set, so it cannot describe a spelling this function would reject.
func ParsePlatform(s string) (Platform, error) {
	if s == "" {
		return "", nil
	}
	name, ok := platformAbbrev.Expand(s)
	if !ok {
		return "", fmt.Errorf("invalid platform %q: expected %s", s, platformAbbrev.List())
	}
	return Platform(name), nil
}

// DetectPlatforms reports which platform sections the env file at path declares,
// in Platforms() order. It reads the RAW document rather than a decoded Config
// because a decoded one cannot answer the question: the platform sections are
// value structs, so an omitted section and an empty one are the same zero value,
// and ApplyDefaults writes into kubernetes.command whatever the platform is.
//
// A key that is present but empty still counts as declared. That is deliberate:
// `docker: {}` is how an env file says "this is a docker file" when every docker
// setting is happy with its default, and the container schema has no mandatory
// field to force the section into existence on its own.
func DetectPlatforms(path string) ([]Platform, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read env file %q: %w", path, err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		// Same diagnostic Load would give, including the legacy-bash-file hint:
		// this is the first thing to read the file, so it is where that mistake
		// now surfaces, and it must not degrade into a barer message than before.
		return nil, parseError(path, raw, err)
	}
	found := make([]Platform, 0, len(Platforms()))
	for _, p := range Platforms() {
		if _, ok := doc[string(p)]; ok {
			found = append(found, p)
		}
	}
	return found, nil
}

// JoinPlatforms renders a platform list for an error or a prompt.
func JoinPlatforms(ps []Platform) string {
	names := make([]string, len(ps))
	for i, p := range ps {
		names[i] = string(p)
	}
	return strings.Join(names, ", ")
}

package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// scalingTier is the compute footprint Solace publishes for one connection tier.
//
// CPU is a property of the tier, not a knob: sizing a broker by connection count
// and then sizing its CPU independently is how a 200k-connection broker ends up
// on 2 cores, so the env file states the tier and this table states the cores.
// Mem is only the tier's *default* -- kubernetes.msgNode.mem and the container blocks'
// mem both still override it, because memory headroom depends on the message
// mix in a way core count does not.
type scalingTier struct {
	cpu string // messagingNodeCpu (k8s); the container cpuset's core count (cpuSetRange); never settable
	mem string // Kubernetes quantity; containerMem rewrites it for the engines
}

// scalingTiers is the authoritative maxConnections -> footprint table, and
// doubles as the enum validateScalingTier checks against: these five are the
// only connection counts Solace publishes sizing for. A value between tiers is
// deliberately refused rather than rounded up -- rounding would silently
// provision cores the operator never asked for, and rounding down would
// under-size a broker that had already declared its load.
var scalingTiers = map[int]scalingTier{
	100:    {cpu: "2", mem: "3410Mi"},
	1000:   {cpu: "2", mem: "6898Mi"},
	10000:  {cpu: "4", mem: "12435Mi"},
	100000: {cpu: "8", mem: "30925Mi"},
	200000: {cpu: "12", mem: "52581Mi"},
}

// scalingTierList names scalingTiers' keys in ascending order for the error
// message. It is a literal for the same reason accessLevelList is: the message
// has to be stable, and this package avoids pulling in sort for a handful of
// items (see sortStrings). TestScalingTierListMatchesTable pins the two
// together so the list cannot drift from the table.
const scalingTierList = "100, 1000, 10000, 100000, 200000"

// tierFor looks up the footprint for maxConnections. ok is false for anything
// that is not a tier -- including the zero value, which is what a Config built
// in code rather than through Load carries. validateScalingTier is what turns
// that into a loud failure; this stays total so applyScalingTierDefaults (which
// necessarily runs before Validate) never has to guess.
func tierFor(maxConnections int) (scalingTier, bool) {
	t, ok := scalingTiers[maxConnections]
	return t, ok
}

// containerMemRE is docker's and podman's own memory syntax: an integer then one
// of b/k/m/g. It exists because this schema sits a Kubernetes quantity
// (kubernetes.msgNode.mem, "3410Mi") next to a container one that rejects that exact
// spelling, so the likeliest mistake is copying the k8s form across.
var containerMemRE = regexp.MustCompile(`(?i)^[0-9]+[bkmg]$`)

// containerMem rewrites a tier's Kubernetes quantity into the suffix docker's
// mem_limit and podman's Memory= accept. Both conventions count in binary units
// -- Kubernetes spells the mebibyte "Mi", the engines spell it "m" -- so this
// only ever drops a trailing "i", never rescales. It is not a general converter:
// it sees only the five fixed strings above, and anything without the trailing
// "i" passes through untouched.
func containerMem(k8sMem string) string {
	if strings.HasSuffix(k8sMem, "i") {
		return strings.ToLower(strings.TrimSuffix(k8sMem, "i"))
	}
	return k8sMem
}

// cpuSetRange turns a tier's core COUNT into the cpuset the engines take: the
// 0-based range over that many cores ("2" -> "0-1", "12" -> "0-11"). It is
// containerMem's counterpart for the other half of the footprint, and no more
// general: it sees only the five fixed strings in scalingTiers.
//
// Anything that is not a positive integer yields "", which setDefault reads as no
// default and both renderers then skip -- so a tier that ever spelled its cpu as
// a millicore quantity could not produce an invalid cpuset. A count of 1 yields
// "0", a single-cpu list both engines accept, rather than the legal but odd "0-0".
func cpuSetRange(cores string) string {
	n, err := strconv.Atoi(strings.TrimSpace(cores))
	if err != nil || n < 1 {
		return ""
	}
	if n == 1 {
		return "0"
	}
	return "0-" + strconv.Itoa(n-1)
}

// cpuSetCount is how many cpus a cpuSetRE-shaped value names: "0-3" is four,
// "0,2,4" is three, "0-3,8" is five. Only reachable after the regex has matched,
// so every half parses and no range runs backwards.
func cpuSetCount(s string) int {
	n := 0
	for _, part := range strings.Split(s, ",") {
		lo, hi, ok := strings.Cut(part, "-")
		if !ok {
			n++
			continue
		}
		l, _ := strconv.Atoi(lo)
		h, _ := strconv.Atoi(hi)
		n += h - l + 1
	}
	return n
}

// applyScalingTierDefaults derives the tier-fixed CPU and the tier-defaulted
// memory. ApplyDefaults calls it *after* the platform branches, which is the
// whole point: maxConnections only reaches its final value in those branches
// (they default it to 100 on k8s and 1000 on containers), so a tier lookup any
// earlier would read a zero.
//
// CPU is assigned unconditionally -- it is derived, so there is no user value to
// preserve. Memory keeps setDefault semantics because both mem keys override it.
//
// An out-of-tier maxConnections leaves both alone: validateScalingTier rejects
// it moments later, and inventing a footprint for a tier that does not exist
// would bury that error under a plausible-looking artifact.
// applyScalingDefaults fills the scaling knobs, taking the only two whose default
// differs by platform as arguments.
//
// Every knob reaches every platform -- a container gets each one as an environment
// variable, Kubernetes through the CR -- and the four below are deliberately the SAME
// number on both, so one env file sizes the same broker whichever platform runs it.
// Writing the two lists out separately is what let that claim quietly stop being true,
// so the difference is now the argument list and nothing else.
func (c *Config) applyScalingDefaults(maxConnections, maxSpoolUsageMB int) {
	setDefaultInt(&c.Scaling.MaxConnections, maxConnections)
	setDefaultInt(&c.Scaling.MaxSpoolUsageMB, maxSpoolUsageMB)
	setDefaultInt(&c.Scaling.MaxQueueMessages, 100)
	setDefaultInt(&c.Scaling.MaxBridges, 25)
	setDefaultInt(&c.Scaling.MaxSubscriptions, 50000)
	setDefaultInt(&c.Scaling.MaxGuaranteedMsgMB, 10)
}

func (c *Config) applyScalingTierDefaults(p Platform) {
	t, ok := tierFor(c.Scaling.MaxConnections)
	if !ok {
		return
	}
	// Still assigned unconditionally, and now read by the Kubernetes CR alone: the
	// container platforms cap cpu with a cpuset instead of a core count (below).
	c.Scaling.CPU = t.cpu
	if p == K8s {
		setDefault(&c.K8s.MsgNode.Mem, t.mem)
	}
	if p.IsContainer() {
		// Both blocks are filled whichever container platform is active, matching
		// applyContainerDefaults' existing parity for name/runUser/dataDir.
		mem := containerMem(t.mem)
		setDefault(&c.Docker.Container.Mem, mem)
		setDefault(&c.Podman.Container.Mem, mem)
		// The tier's cores as a cpuset, not as a count: --cpus= is a fractional
		// quota and a broker sized by whole cores wants the cores. Same setDefault
		// semantics as mem -- which cpus are free is the operator's knowledge.
		set := cpuSetRange(t.cpu)
		setDefault(&c.Docker.Container.CPUSet, set)
		setDefault(&c.Podman.Container.CPUSet, set)
	}
}

// validateScaling checks the scaling block, which is platform-independent: every
// knob reaches every platform, so this runs ahead of the platform switch in
// Validate rather than inside either platform validator.
func (c *Config) validateScaling() error {
	if _, ok := tierFor(c.Scaling.MaxConnections); !ok {
		return fmt.Errorf("scaling.maxConnections must be one of the supported scaling tiers %s (got: %d); "+
			"broker CPU is fixed by the tier, so a value between tiers has no published sizing to apply",
			scalingTierList, c.Scaling.MaxConnections)
	}
	if c.Scaling.MaxPool != 0 {
		// Removed rather than aliased: silently forwarding it would leave two keys
		// for one setting, and an env file setting both would have no defined
		// winner. That is NOT what the destination-name alias every other
		// setting gets (scalingKeys, below) does: setting both spellings of one
		// of those at once is a load failure naming both, so there is never an
		// undecided winner there. maxPool had no such mechanism, which is why it
		// was removed rather than added as a third spelling.
		return fmt.Errorf("scaling.maxPool was removed; it named the same broker setting as "+
			"scaling.maxSpoolUsageMB, which now feeds messagespool_maxspoolusage on every platform -- "+
			"rename it to scaling.maxSpoolUsageMB (got: %d)", c.Scaling.MaxPool)
	}
	return nil
}

// --- the dual-spelling allowlist ---------------------------------------------
//
// Every scaling setting is settable under either its friendly schema name or
// the destination broker setting this tool emits for it. Both write the same
// field. This table is the single source of truth for that pairing AND the
// closed allowlist Scaling.UnmarshalYAML decodes against -- see the doc
// comment on Scaling (config.go) for why a table is needed here at all: a
// custom UnmarshalYAML takes over decoding for the whole struct, and
// yaml.v3's KnownFields(true) (load.go) never reaches inside it.

// scalingKey is one YAML-settable scaling field.
type scalingKey struct {
	yamlName string              // the schema's own name, e.g. "maxConnections"
	dest     string              // the destination broker setting, or "" when there is none
	field    func(*Scaling) *int // the one typed field either spelling writes
}

// scalingKeys enumerates every mapped setting plus maxPool, in struct field
// order -- which is also the order scalingKeyList renders an error in.
//
// Every entry's destination is the name BOTH platforms emit: the key inside the
// CR's spec.systemScaling, and the environment variable name for docker and
// podman. Keeping those identical is what makes "the destination spelling" a
// single answerable thing rather than a per-platform one, and it is worth
// guarding: a setting whose two platforms disagreed would make the alias
// accepted here name something only one of them actually writes.
//
// That constraint is not free. A destination must also be a legal systemd
// Environment= name, which is alphanumeric and underscore only -- so a
// hyphenated broker setting cannot be carried this way at all, and a setting
// that needs one cannot simply be added to this table.
var scalingKeys = []scalingKey{
	{"maxConnections", "system_scaling_maxconnectioncount", func(s *Scaling) *int { return &s.MaxConnections }},
	{"maxQueueMessages", "system_scaling_maxqueuemessagecount", func(s *Scaling) *int { return &s.MaxQueueMessages }},
	{"maxSpoolUsageMB", "messagespool_maxspoolusage", func(s *Scaling) *int { return &s.MaxSpoolUsageMB }},
	{"maxKafkaBridge", "system_scaling_maxkafkabridgecount", func(s *Scaling) *int { return &s.MaxKafkaBridge }},
	{"maxKafkaConnections", "system_scaling_maxkafkabrokerconnectioncount", func(s *Scaling) *int { return &s.MaxKafkaConnections }},
	{"maxBridges", "system_scaling_maxbridgecount", func(s *Scaling) *int { return &s.MaxBridges }},
	{"maxSubscriptions", "system_scaling_maxsubscriptioncount", func(s *Scaling) *int { return &s.MaxSubscriptions }},
	{"maxGuaranteedMsgMB", "system_scaling_maxguaranteedmessagesize", func(s *Scaling) *int { return &s.MaxGuaranteedMsgMB }},
	// maxPool has no destination: the broker setting it once named is now
	// maxSpoolUsageMB's alone (validateScaling, above). It stays in the table
	// purely so a file carrying the removed key decodes far enough to reach
	// that explanation instead of dying here with a generic unknown-key error
	// -- the whole reason the field is retained at all (config.go).
	{"maxPool", "", func(s *Scaling) *int { return &s.MaxPool }},
}

// scalingDenylist names scaling-shaped keys that must never decode, because
// the field they would set is DERIVED rather than read from the file
// (Scaling.CPU, yaml:"-"). Checked ahead of the allowlist lookup so the
// message explains WHY, the way kubernetes.msgNode.cpu's removal does
// (validateK8s) -- a bare unknown-key message would read as a typo rather
// than a schema rule.
var scalingDenylist = map[string]string{
	"cpu": "scaling.cpu is fixed by the maxConnections tier and derived once ApplyDefaults runs; " +
		"there is no cpu key in this schema",
	"messagingNodeCpu": "messagingNodeCpu is fixed by the maxConnections tier and derived once ApplyDefaults " +
		"runs; there is no messagingNodeCpu key in this schema",
}

// scalingKeyIndex maps every accepted spelling -- friendly and destination --
// to the one scalingKey it writes.
func scalingKeyIndex() map[string]*scalingKey {
	idx := make(map[string]*scalingKey, 2*len(scalingKeys))
	for i := range scalingKeys {
		k := &scalingKeys[i]
		idx[k.yamlName] = k
		if k.dest != "" {
			idx[k.dest] = k
		}
	}
	return idx
}

// scalingKeyList renders every accepted spelling for an unknown-key error, in
// scalingKeys' own order -- so the message names every real way out for a
// typo instead of just refusing it, the way scalingTierList does for the tier
// enum above.
func scalingKeyList() string {
	parts := make([]string, len(scalingKeys))
	for i, k := range scalingKeys {
		if k.dest == "" {
			parts[i] = k.yamlName
			continue
		}
		parts[i] = fmt.Sprintf("%s (or %s)", k.yamlName, k.dest)
	}
	return strings.Join(parts, ", ")
}

// scalingSpelling records which YAML key first claimed a scalingKey, so a
// second spelling of the same setting can be refused naming both (rule: two
// spellings of one setting set at once fails to load, naming both keys).
type scalingSpelling struct {
	key  string
	line int
}

// UnmarshalYAML decodes the scaling: block by hand, off scalingKeys, because a
// custom UnmarshalYAML TAKES OVER decoding for the struct it is declared on
// and yaml.v3's KnownFields(true) (load.go) never reaches inside it. Four
// things establish that, read against gopkg.in/yaml.v3@v3.0.1: Decoder.Decode
// (yaml.go) is the ONLY place knownFields is copied onto a working decoder;
// callUnmarshaler (decode.go) hands this method the bare *yaml.Node and
// nothing else, so it cannot see that flag even if it wanted to; the only
// place an unknown key becomes an error is the default struct-mapping path
// (decode.go), gated on that same flag; and Node.Decode -- the only other way
// to decode a sub-node, which this method uses below -- builds a FRESH
// decoder with knownFields false. So without this method, "scaling:" would be
// OPEN PASSTHROUGH: `system_scaling_maxtransactedsessioncount: 500` (a real
// broker setting this tool does not map) or `maxConections: 1000` (a typo)
// would both decode clean, leave the real field at its zero value, and
// ApplyDefaults would silently size the broker at the default tier -- exactly
// the outcome KnownFields(true) exists to prevent everywhere else in this
// schema. This method reimplements that strictness by hand.
//
// decode.go's d.prepare checks for a null node BEFORE it ever looks at
// whether the target implements Unmarshaler, so a bare `scaling:` with
// nothing after it never reaches this method at all -- the field is left at
// its zero value exactly as an omitted key would leave it, with no special
// case needed here for that.
//
// Every key is looked up under EITHER spelling (scalingKeyIndex; both write
// the same field), refused by name when it is in scalingDenylist (a derived
// field with no YAML key at all), and refused as a collision when the setting
// is already claimed by a DIFFERENT spelling earlier in this same mapping --
// keyed by the field rather than the key text, since two spellings of one
// setting must never both be honoured with no defined winner (the problem
// that got maxPool removed instead of aliased, above). Every failure is
// collected into one *yaml.TypeError rather than returned as soon as it is
// found: callUnmarshaler splices a *TypeError's Errors into the document's
// own error list and keeps decoding, while any OTHER error type panics
// (yaml.go's fail/handleErr) and aborts the whole document immediately -- so
// a plain error here would hide every other mistake the rest of the file
// carries, where today a typo'd top-level key is reported alongside them.
func (s *Scaling) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return &yaml.TypeError{Errors: []string{
			fmt.Sprintf("line %d: scaling must be a mapping of settings to values", value.Line),
		}}
	}
	index := scalingKeyIndex()
	seen := make(map[*scalingKey]scalingSpelling, len(scalingKeys))
	var errs []string
	for i := 0; i+1 < len(value.Content); i += 2 {
		errs = append(errs, s.decodeScalingEntry(value.Content[i], value.Content[i+1], index, seen)...)
	}
	if len(errs) > 0 {
		return &yaml.TypeError{Errors: errs}
	}
	return nil
}

// decodeScalingEntry decodes one "key: value" pair of the scaling: mapping,
// returning zero or more error lines rather than stopping at the first --
// UnmarshalYAML collects every entry's errors into one *yaml.TypeError, the
// same way KnownFields(true) collects every unknown top-level key.
func (s *Scaling) decodeScalingEntry(
	keyNode, valNode *yaml.Node, index map[string]*scalingKey, seen map[*scalingKey]scalingSpelling,
) []string {
	name := keyNode.Value
	if reason, denied := scalingDenylist[name]; denied {
		return []string{fmt.Sprintf("line %d: scaling.%s: %s", keyNode.Line, name, reason)}
	}
	k, ok := index[name]
	if !ok {
		return []string{fmt.Sprintf("line %d: scaling.%s is not a scaling setting; expected one of: %s",
			keyNode.Line, name, scalingKeyList())}
	}
	if prev, dup := seen[k]; dup {
		if prev.key == name {
			return []string{fmt.Sprintf("line %d: scaling.%s is set twice (already at line %d)",
				keyNode.Line, name, prev.line)}
		}
		return []string{fmt.Sprintf("line %d: scaling.%s duplicates scaling.%s at line %d -- "+
			"they are two spellings of the same setting; set it under one name",
			keyNode.Line, name, prev.key, prev.line)}
	}
	seen[k] = scalingSpelling{name, keyNode.Line}
	if err := valNode.Decode(k.field(s)); err != nil {
		if te, ok := err.(*yaml.TypeError); ok {
			// Node.Decode's own errors carry a line but never the field: prefix
			// each one with the key so the operator does not have to guess which
			// setting an unadorned "cannot unmarshal ..." refers to.
			out := make([]string, len(te.Errors))
			for i, e := range te.Errors {
				out[i] = fmt.Sprintf("scaling.%s: %s", name, e)
			}
			return out
		}
		return []string{fmt.Sprintf("line %d: scaling.%s: %v", valNode.Line, name, err)}
	}
	return nil
}

package config

import (
	"fmt"
	"strings"

	"solace/internal/abbrev"
)

// Role is a broker node role. The single-letter forms (p|b|m) appear in k8s pod
// names; the long forms are the CLI-facing positional args.
type Role string

const (
	Primary Role = "p"
	Backup  Role = "b"
	Monitor Role = "m"
)

// roleTable is the one declaration of the role vocabulary, in redundancy order.
// The letter is not a separate spelling to keep in step: it IS the Role value,
// so the abbreviation and the internal form cannot drift by construction.
var roleTable = []struct {
	name string
	role Role
	note string
}{
	{"primary", Primary, "the default when the argument is omitted"},
	{"backup", Backup, ""},
	{"monitor", Monitor, ""},
}

var roleAbbrev, roleByName = newRoles()

// newRoles builds the role set and its reverse lookup from roleTable in one
// pass, so the parser and the reference cannot disagree about what exists.
func newRoles() (*abbrev.Set, map[string]Role) {
	entries := make([]abbrev.Entry, 0, len(roleTable))
	byName := make(map[string]Role, len(roleTable))
	for _, r := range roleTable {
		entries = append(entries, abbrev.Entry{
			Canonical: r.name,
			Short:     []string{r.role.Letter()},
			Note:      r.note,
		})
		byName[r.name] = r.role
	}
	return abbrev.New("role", entries), byName
}

// RoleAbbrev returns the approved role spellings, for the abbreviation reference.
func RoleAbbrev() *abbrev.Set { return roleAbbrev }

// ParseRole normalizes a role argument to its single-letter form, porting
// pick_pod. Accepts p|primary, b|backup, m|monitor; empty defaults to primary.
//
// The error teaches both spellings, rendered from the set rather than written
// out beside it: a hand-typed list is what silently outlives the table it
// describes.
func ParseRole(s string) (Role, error) {
	if s == "" {
		return Primary, nil
	}
	name, ok := roleAbbrev.Expand(s)
	if !ok {
		return "", fmt.Errorf("invalid node role %q (expected %s or %s)", s,
			strings.Join(roleAbbrev.Shorts(), "|"), strings.Join(roleAbbrev.Names(), "|"))
	}
	return roleByName[name], nil
}

// RoleNames returns the long role names in redundancy order, for completing the
// [role] positionals and --pod. ParseRole stays the only validator -- the p|b|m
// forms it also accepts are not worth suggesting, and both now read the same
// set, so the suggestion list cannot offer a word the parser rejects.
func RoleNames() []string { return roleAbbrev.Names() }

// Letter returns the single-letter role, matching pod-name suffixes.
func (r Role) Letter() string { return string(r) }

// NodeIdentity is a host's resolved Solace identity for container deployment,
// porting node_env/resolve_node's THIS_HOSTNAME/THIS_NODETYPE/THIS_ACTIVESTANDBY.
type NodeIdentity struct {
	Hostname      string // routername
	NodeType      string // message_routing | monitoring
	ActiveStandby string // primary | backup | "" (monitor / standalone)
}

// ResolveNode returns this host's identity for the given role, honoring
// redundancy mode. In HA (redundancy: yes) the role selects from the node table;
// in standalone (redundancy: no) there is one message_routing node named after
// the primary, with no active/standby role, and the role arg is ignored.
func (c *Config) ResolveNode(role Role) NodeIdentity {
	if !c.RedundancyEnabled() {
		return NodeIdentity{Hostname: c.Nodes.Primary.Name, NodeType: "message_routing"}
	}
	switch role {
	case Backup:
		return NodeIdentity{Hostname: c.Nodes.Backup.Name, NodeType: "message_routing", ActiveStandby: "backup"}
	case Monitor:
		return NodeIdentity{Hostname: c.Nodes.Monitor.Name, NodeType: "monitoring"}
	default:
		return NodeIdentity{Hostname: c.Nodes.Primary.Name, NodeType: "message_routing", ActiveStandby: "primary"}
	}
}

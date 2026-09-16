package broker

import (
	"context"
	"fmt"
	"strings"

	"solace/internal/config"
)

// MateChannel reaches ONE broker of a replication pair.
//
// It is deliberately NOT broker.Transport. That interface takes (ctx, role) and has no
// notion of a host: the k8s implementation derives a pod from cfg.K8s.Namespace, and the
// container ones are node-local and ignore the role entirely. A second SITE is not
// addressable through it, and widening a type every existing op depends on -- for one
// feature -- is a worse trade than a narrow interface beside it.
//
// Implementations are addressed by site, not by role. A site's HA group is ONE endpoint
// here, because the replication role config-syncs within a site (`HA: yes` on every
// replication command): reach the primary and the backup receives it.
//
// Config-sync is ASSUMED OPERATIONAL and deliberately never checked (operator-confirmed).
// It is off by default and it can be oper-down, so the assumption is not free -- but an
// HA group in service has it running, and probing it here would be this tool
// second-guessing an invariant the operator maintains. The fact is appliance-proven from
// the CLI reference's per-command CONFIG-SYNC blocks and verified on live software
// brokers, which is why both broker types are treated identically.
type MateChannel interface {
	// Preflight proves the channel works and mutates nothing. Every flow calls it on
	// BOTH channels before the first write, so an unreachable site stops the run while
	// the pair is still undisturbed rather than halfway through a switchover.
	Preflight(ctx context.Context) error

	// ShowReplication reads the mate configuration this broker holds.
	ShowReplication(ctx context.Context) (MateConfig, error)

	// ReadRoles reads every listed VPN's replication admin-state and config-state.
	// A VPN the broker does not have is absent from the map rather than an error --
	// the caller knows which VPNs it asked about and reports the difference itself.
	//
	// A NIL list means every VPN the broker has. That is what the local convergence
	// needs, since shutting replication down on a VPN the file does not list requires
	// knowing which unlisted VPNs are replicating; a switch always passes its own list,
	// so it can never reach a VPN nobody named.
	ReadRoles(ctx context.Context, vpns []string) (map[string]VPNRepl, error)

	// SetRole moves one VPN's replication role.
	SetRole(ctx context.Context, vpn string, role ReplRole) error

	// PrimaryActive reports whether the node this channel addresses currently holds
	// activity for its own HA group. A standalone broker is its own only node and
	// reports true.
	//
	// A switch requires it at BOTH sites (operator-confirmed). See
	// requirePrimaryActive for why that is a gate rather than a preference.
	PrimaryActive(ctx context.Context) (bool, error)

	// Status returns the raw text of the two `show` reports, for DISPLAY only.
	//
	// Nothing decides anything from it: every verdict comes from the parsed methods
	// above. Keeping the two apart means a display-format change in some future broker
	// version can alter what an operator reads and not what the tool does.
	Status(ctx context.Context) (string, error)

	// Describe names this channel for an error or a report ("site v:sol-b over semp").
	Describe() string
}

// CLIRunner uploads a CLI script and returns its transcript. It is the one thing a
// CLI-driven MateChannel needs, and it is a func rather than an interface so
// internal/k8s can supply its own without importing anything new: the kubectl
// implementation of this channel CANNOT live in this package, because internal/k8s
// already imports internal/broker and the reverse would be an import cycle.
type CLIRunner func(ctx context.Context, name, script string) ([]byte, error)

// Script names. They name both the uploaded file and the CLI's own argument, so they are
// constants rather than literals at each call: the two halves of one run cannot then
// name different files.
const (
	scriptShowReplication = "repl-show-replication"
	scriptShowVPN         = "repl-show-vpn"
	scriptSetRole         = "repl-set-role"
	scriptSetEnabled      = "repl-set-enabled"
	scriptShowRedundancy  = "repl-show-redundancy"
)

// cliMate drives one broker over its CLI. It backs both the local channel and the
// kubernetes one -- they differ only in which broker the runner reaches.
type cliMate struct {
	run   CLIRunner
	what  string     // for Describe
	typ   BrokerType // discovered from the login banner, cached
	typed bool
}

// NewCLIMate builds a MateChannel over a CLI runner. what names the target in messages.
func NewCLIMate(run CLIRunner, what string) MateChannel {
	return &cliMate{run: run, what: what}
}

// NewLocalMate is the channel to THIS deployment's own broker, over the transport the
// ops already hold. role is the node to address -- always the site's primary, since the
// role config-syncs to its HA backup.
//
// This CLIRunner goes through runCLIRead, RunCLI's unwrapped sibling, not RunCLI itself.
// Every method above but SetRole is a `show`-style read (ShowReplication, ReadRoles,
// PrimaryActive, Status, and Preflight's own brokerType probe), and per runCLIRead's own
// doc comment a read has nothing for RunCLI's stop-on-error wrapper to protect. SetRole is
// the one write on this channel, and it stays unwrapped too -- deliberately: see
// rejectionIn's doc comment for why its own tail scan is still sound over SetRole's short,
// single-command script without the wrapper. Wrapping SetRole would need the same
// skeleton-rendering helper on internal/k8s's own implementation of this channel (which
// cannot import this package's unexported runCLISkeleton, and cannot share code with it at
// all without an exported equivalent -- see CLIScriptPath/CLIArg for the existing instance
// of that same problem); that is a real follow-on, not done here.
func (o *Ops) NewLocalMate(role config.Role) MateChannel {
	return NewCLIMate(func(ctx context.Context, name, script string) ([]byte, error) {
		// readCLI, not runCLIRead: this channel's methods are called repeatedly
		// across a switchover, and each left its uploaded script on the broker.
		return o.readCLI(ctx, role, name, script)
	}, "this broker")
}

func (c *cliMate) Describe() string { return c.what }

// brokerType reads the type from the login banner `cli -Apes` echoes ahead of any
// output, and caches it.
//
// It has to be discovered rather than configured: the mate-address grammar and the
// `show replication` layout both differ between software and an appliance, and the env
// file deliberately does not state which a site is -- the broker knows, and a `type:` key
// would be a second source of truth whose only effect when wrong is to emit CLI the
// target rejects.
func (c *cliMate) brokerType(ctx context.Context) (BrokerType, error) {
	if c.typed {
		return c.typ, nil
	}
	out, err := c.run(ctx, scriptShowReplication, showReplicationScript())
	if err != nil {
		return BrokerUnknown, err
	}
	t := bannerType(out)
	if t == BrokerUnknown {
		return BrokerUnknown, fmt.Errorf("cannot tell whether %s is a software broker or an appliance: "+
			"its CLI login banner named neither. The replication grammar and the `show replication` "+
			"layout differ between the two, so there is nothing safe to assume", c.what)
	}
	c.typ, c.typed = t, true
	return t, nil
}

// Preflight runs the cheapest read that proves the whole path: the transport reaches the
// broker, the CLI runs, and the banner says what kind of broker it is -- which every
// later call needs anyway.
func (c *cliMate) Preflight(ctx context.Context) error {
	if _, err := c.brokerType(ctx); err != nil {
		return fmt.Errorf("cannot reach %s: %w", c.what, err)
	}
	return nil
}

func (c *cliMate) ShowReplication(ctx context.Context) (MateConfig, error) {
	t, err := c.brokerType(ctx)
	if err != nil {
		return MateConfig{}, err
	}
	out, err := c.run(ctx, scriptShowReplication, showReplicationScript())
	if err != nil {
		return MateConfig{}, fmt.Errorf("read replication config from %s: %w", c.what, err)
	}
	m, err := ParseShowReplication(t, out)
	if err != nil {
		return MateConfig{}, fmt.Errorf("%s: %w", c.what, err)
	}
	return m, nil
}

func (c *cliMate) ReadRoles(ctx context.Context, vpns []string) (map[string]VPNRepl, error) {
	out, err := c.run(ctx, scriptShowVPN, showVPNReplicationScript())
	if err != nil {
		return nil, fmt.Errorf("read replication roles from %s: %w", c.what, err)
	}
	all, err := ParseVPNReplication(out)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", c.what, err)
	}
	if vpns == nil {
		return all, nil
	}
	// Narrow to what was asked about. The broker reports every VPN it has, and a
	// caller acting on the whole table would reach VPNs the env file never named --
	// exactly the blast radius `replication.vpns` exists to bound.
	out2 := make(map[string]VPNRepl, len(vpns))
	for _, v := range vpns {
		if st, ok := all[v]; ok {
			out2[v] = st
		}
	}
	return out2, nil
}

func (c *cliMate) SetRole(ctx context.Context, vpn string, role ReplRole) error {
	if role != RoleActive && role != RoleStandby {
		return fmt.Errorf("cannot set message-vpn %q to replication role %q: only active and standby "+
			"are settable (n/a is a reading, not a state to write)", vpn, role)
	}
	out, err := c.run(ctx, scriptSetRole, setReplicationRoleScript(vpn, role))
	if err != nil {
		return fmt.Errorf("set message-vpn %q to %s on %s: %w", vpn, role, c.what, err)
	}
	// `cli -Apes` exits 0 even when the broker rejects a line, so the transcript is the
	// only evidence there is. This is the same scan import's driver makes, against the
	// same list -- one definition of what a rejection looks like.
	if bad := rejectionIn(out); bad != "" {
		return fmt.Errorf("%s rejected setting message-vpn %q to %s: the transcript carries %q",
			c.what, vpn, role, bad)
	}
	return nil
}

// PrimaryActive reads `show redundancy` on the node this channel addresses.
//
// A broker with redundancy shut down is STANDALONE: it is its own only node, there is no
// activity to hold, and reporting false would refuse every switch on a single-broker
// pair. So the question is only asked of a node that has an HA group.
func (c *cliMate) PrimaryActive(ctx context.Context) (bool, error) {
	out, err := c.run(ctx, scriptShowRedundancy, showRedundancyLocalScript())
	if err != nil {
		return false, fmt.Errorf("read redundancy from %s: %w", c.what, err)
	}
	s := string(out)
	if field(s, labelConfigStatus) != "Enabled" {
		return true, nil // standalone: no HA group, so this node is the only one
	}
	return activity(s, activityLocalActive) >= 1, nil
}

func (c *cliMate) Status(ctx context.Context) (string, error) {
	mate, err := c.run(ctx, scriptShowReplication, showReplicationScript())
	if err != nil {
		return "", fmt.Errorf("read replication status from %s: %w", c.what, err)
	}
	vpns, err := c.run(ctx, scriptShowVPN, showVPNReplicationScript())
	if err != nil {
		return "", fmt.Errorf("read replication status from %s: %w", c.what, err)
	}
	return string(mate) + "\n" + string(vpns), nil
}

// rejectionIn reports the first failure keyword in the tail of a transcript, or "".
//
// Tail rather than whole, for the reason driverTailLines gives: the broker abandons a
// script at the offending line, so a rejection is the last thing printed -- while
// scanning the whole transcript would scan the configuration itself, where an object
// legitimately named `error-events` would match.
//
// Its OPERATING ENVELOPE: this scan is sound only over output that STOPPED at the
// rejection, which is true in exactly two shapes. One is a script run through the
// broker's own `source script ... stop-on-error` (RunCLI's wrapper, driver.go's chunks),
// where the broker itself abandons the script at the first bad line, so a rejection is
// necessarily the last thing printed. The other is a single BOUNDED object short enough
// that "the tail" is the whole thing regardless -- SetRole's script here is one `state
// <role>` line inside a `replication` block, so there is nothing after a rejection for the
// tail window to miss even without stop-on-error. What this is NOT sound over is a long,
// unwrapped script: a rejection on line 50 of a 500-line script that keeps running past it
// (no stop-on-error) would scroll out of the last driverTailLines lines, and the scan would
// report a clean run that was not one.
func rejectionIn(out []byte) string {
	lines := strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n")
	if len(lines) > driverTailLines {
		lines = lines[len(lines)-driverTailLines:]
	}
	tail := strings.ToLower(strings.Join(lines, "\n"))
	for _, kw := range failKeywords {
		if strings.Contains(tail, strings.ToLower(kw)) {
			return kw
		}
	}
	return ""
}

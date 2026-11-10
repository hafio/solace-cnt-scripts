package container

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hafio/go-deploy-utility/podman"
	"github.com/hafio/go-deploy-utility/rootless"
	"github.com/hafio/go-deploy-utility/runner"
	"github.com/hafio/go-deploy-utility/systemd"

	"solace/internal/config"
	"solace/internal/engine"
)

// nearestWritableScript walks up from the configured data directory to the first
// path that exists, prints it, and reports whether it is writable. The walk is in
// the shell rather than in Go because the value is a path on the Linux host the
// broker runs on: this tool is routinely driven from Windows, where an os.Stat of
// "/opt/solace/data" answers a question about the wrong filesystem.
//
// stdout is the directory it settled on -- captured even when the test fails, so
// the refusal can name the directory whose ownership is actually wrong.
const nearestWritableScript = `d="$0"; while [ ! -e "$d" ] && [ "$d" != "/" ]; do d=$(dirname "$d"); done; printf %s "$d"; test -w "$d"`

// checkPodmanHost is the podman host-readiness block behind `validate` and the
// first thing `broker deploy` does. Every row is READ-ONLY and every failure is a
// refusal, never a fix: reading a prerequisite needs no privilege, but repairing
// several of them means editing host-wide configuration as root, which is exactly
// what a rootless deployment exists to avoid (the same argument limits.go and
// Preflight already make).
//
// It reports every row before returning, rather than stopping at the first, because
// `validate` promises one pass over everything it can see. The rows that need the
// user bus are SKIPPED rather than failed when the session itself is unusable --
// three failures with one cause are less useful than one plus two honest skips,
// which is the rule k8s.verifyRows follows for an unreachable API server.
// fix says whether this caller may repair what it finds. It is false from `validate`,
// whose own help promises a run that is "safe against a system you don't want to
// disturb", and true from PrepHost -- the same block, read twice over, rather than two
// definitions of what a ready host is. Exactly ONE row acts on it: linger, which needs
// no privilege for the invoking user and so belongs with the other unprivileged
// preparations prep already performs (mkdir, chown, podman unshare chown, podman login).
func (m *Manager) checkPodmanHost(ctx context.Context, fix bool) error {
	if m.P != config.Podman {
		return nil
	}
	r := m.report()
	if m.isEcho() {
		r.KVRow(reportKeyWidth, "podman host", "skipped (preview)")
		return nil
	}
	euid := m.Geteuid()
	if euid < 0 {
		// Windows: no meaningful euid, and therefore no rootless podman to probe.
		r.KVRow(reportKeyWidth, "podman host", "skipped (no POSIX euid here)")
		return nil
	}

	// The euid invariant gates everything after it: a rootless block probed as root
	// reads the WRONG user's linger and runtime directory. cli.prepare refuses a
	// mismatch before this block can run, so reaching it means a Manager built
	// outside the CLI -- it fails and probes nothing rather than answering about an
	// account the deploy will never use.
	if err := m.checkPodmanEUID(); err != nil {
		r.Fail("euid: %v", err)
		return err
	}
	r.OK("euid: %d, matching podman.rootful=%t", euid, m.Cfg.Podman.Rootful)
	if m.Cfg.Podman.Rootful {
		// Rootful podman's engine runs privileged: it owns the id mapping and its
		// units live under the system systemd instance. Its nofile ceiling is
		// checked too, but one level up -- it binds every engine (limits.go).
		return nil
	}
	return m.rootlessRows(ctx, euid, fix)
}

// rootlessRows is the rootless-only half of checkPodmanHost, split out so the
// gating above (platform, preview, euid) reads as the short guard it is. Every row
// runs and reports; the errors are joined so one pass names every problem.
func (m *Manager) rootlessRows(ctx context.Context, euid int, fix bool) error {
	r := m.report()
	// Resolved once: two rows hand an administrator a command naming this account,
	// and both must name it literally rather than as a substitution.
	user := m.hostUserName(ctx, euid)
	var errs []error
	sessionErr := m.ensureUserSession(ctx)
	if sessionErr != nil {
		r.Fail("user session: %v", sessionErr)
		errs = append(errs, sessionErr)
	} else {
		r.OK("user session: %s", m.sessionDesc)
	}

	// Independent of the bus: podman reads its own id mapping, and the data
	// directory is a plain filesystem fact.
	if err := m.checkIDMapping(ctx, user); err != nil {
		errs = append(errs, err)
	}
	if sessionErr == nil {
		if err := m.ensureLinger(ctx, euid, user, fix); err != nil {
			errs = append(errs, err)
		}
		if err := m.checkUserSystemd(ctx); err != nil {
			errs = append(errs, err)
		}
	} else {
		r.Skip("linger: skipped (no user session)")
		r.Skip("user systemd: skipped (no user session)")
	}
	if err := m.checkDataDir(ctx, user); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// ensureUserSession is the one thing this block SETS rather than reports. It runs
// before every other rootless row and before any `systemctl --user` call, because
// a missing XDG_RUNTIME_DIR makes the bus unreachable for a reason that has nothing
// to do with whether the host is ready.
//
// The rule is the shared rootless.UserSession: the runtime directory is proven to
// EXIST before either variable is exported, since a derived path to a directory
// that is not there is worse than no value at all -- it turns "you are under sudo,
// or this user has no session" into an obscure systemctl or podman failure several
// rows later. An INHERITED value is never overwritten: an operator who set one
// meant it. Only an empty variable is filled, and the report says which half came
// from where.
//
// The existence probe is `sh -c 'test -d "$0"' <dir>` through the runner, so it asks
// the host the runner reaches. Exit status 1 is "no such directory"; any other
// failure is a probe that could not run, which UserSession reports as such rather
// than as a missing directory.
//
// The mechanism is os.Setenv through a seam, not Cmd.Env: every holder but the compose
// manager wraps its runner in engine.NoEnv so that putting a value in a child's
// environment stays the compose-secrets path alone. Neither of these is a secret,
// and routing them through Cmd.Env would reopen the door that wrapper exists to close.
//
// It is memoised because every systemctl call funnels through it, and it must
// probe once per run rather than once per command.
func (m *Manager) ensureUserSession(ctx context.Context) error {
	if m.sessionOnce {
		return m.sessionErr
	}
	m.sessionOnce = true
	if !m.rootlessPodman() || m.isEcho() {
		return nil
	}
	euid := m.Geteuid()
	if euid < 0 {
		return nil
	}

	isDir := func(dir string) (bool, error) {
		_, err := engine.Output(ctx, m.R, runner.Cmd{Argv: []string{"sh", "-c", `test -d "$0"`, dir}})
		if err == nil {
			return true, nil
		}
		var st interface{ ExitCode() int }
		if errors.As(err, &st) && st.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	session, err := rootless.UserSession(euid, m.Getenv, isDir)
	if err != nil {
		m.sessionErr = err
		return m.sessionErr
	}
	for _, kv := range session.Env() {
		key, value, _ := strings.Cut(kv, "=")
		if err := m.Setenv(key, value); err != nil {
			m.sessionErr = fmt.Errorf("setting %s=%s: %w", key, value, err)
			return m.sessionErr
		}
	}
	m.sessionDesc = session.String()
	return nil
}

// checkIDMapping asserts that the container's runUser falls inside the subuid and
// subgid ranges this user actually has. Podman is asked rather than /etc/subuid
// read, because podman's own view is the authoritative one across the allocation
// backends (files, or an external source) and it is already reachable here.
//
// Without this the failure surfaces as `podman unshare chown` reporting a bare
// "invalid argument" during prep, which names neither the id nor the range.
func (m *Manager) checkIDMapping(ctx context.Context, user string) error {
	r := m.report()
	runUser := m.Cfg.ContainerBlock(m.P).RunUser
	uid, gid, numeric := runUserIDs(runUser)
	if !numeric {
		// config accepts account NAMES here (validateContainer), and resolving one
		// means reading the CONTAINER's passwd file, which does not exist yet.
		m.progress().Info("id mapping: runUser %q names accounts rather than numeric ids, so its subuid/subgid range is not checked", runUser)
		return nil
	}
	rt, err := m.runtime()
	if err != nil {
		return err
	}
	base := rt.Argv()
	for _, k := range []struct {
		file, field string
		gid         bool
		id          int
	}{
		{"subuid", "UIDMap", false, uid},
		{"subgid", "GIDMap", true, gid},
	} {
		out, err := engine.Output(ctx, m.R, podman.IDMap(base, k.gid))
		if err != nil {
			r.Fail("id mapping: cannot read this user's %s range from `podman info`", k.file)
			return fmt.Errorf("reading the %s range from `podman info` (--format .Host.IDMappings.%s): %w", k.file, k.field, err)
		}
		// This process already runs as the account the container will use, so an
		// empty list is an answer about THAT account and not a gap in what was
		// asked. `{{range}}` over an absent mapping yields nothing without erroring,
		// and this branch is rootless-only, so empty means the account genuinely has
		// no allocation -- the shape a `useradd --system` or a directory (LDAP/AD)
		// account arrives in. Reported, not skipped: podman then maps only uid 0, so
		// every non-zero runUser is unusable and the reason is invisible. Id 0
		// always passes (rootless.CheckIDMap): it is the invoking user itself.
		if err := rootless.CheckIDMap(string(out), k.file, k.id, "container.runUser", runUser); err != nil {
			ranges := strings.TrimSpace(string(out))
			if ranges == "" {
				r.Fail("id mapping: this user has no %s allocation at all", k.file)
			} else {
				r.Fail("id mapping: %s %d is outside this user's %s ranges (%s)", k.file, k.id, k.file, ranges)
			}
			return m.idMappingErr(ctx, user, k.id, err)
		}
	}
	r.OK("id mapping: runUser %s is inside this user's subuid/subgid ranges", runUser)
	return nil
}

// idMappingErr is the shared remedy for both subid failures -- none allocated, and
// allocated but too small. They differ only in the diagnosis, and the commands are
// the same either way, so the commands live in one place: rootless.SubIDFix, fed the
// range suggestSubIDRange computes from this host's own files.
//
// The account is named LITERALLY rather than as `$(id -un)`: the privileged line is
// meant to be forwarded to an administrator, and a command substitution would resolve
// to whoever ends up running it. The range has to reach the id itself, so the count
// is id+1 -- a uid of 1000 needs 1001, which is why the rootless default is a low id
// rather than the image's own 1000001 (that would need 1000002).
func (m *Manager) idMappingErr(ctx context.Context, user string, id int, diagnosis error) error {
	start, end, computed := m.suggestSubIDRange(ctx, id+1)
	// The reasoning behind the range, the accounts that arrive without one, and the
	// runUser: "0:0" alternative are all in docs/operations.md. An error says what to
	// run, not why.
	return fmt.Errorf("%w.\n%s", diagnosis, rootless.SubIDFix(user, start, end, computed))
}

// suggestSubIDRange proposes a range that does not overlap anything already
// allocated on this host (rootless.NextFreeRange). Overlap is the reason this is
// computed rather than hardcoded: two accounts sharing a subuid range map to the same
// host uids, which is a containment hole, and only the newer shadow-utils refuse to
// create one.
//
// Both files are world-readable, so this needs no privilege even though applying the
// result does. A read that fails is reported as such by the caller rather than being
// passed off as a computed answer -- the fallback is a convention, and saying so is
// what lets an administrator know to check it by hand.
//
// It cannot see ranges served by SSSD rather than the files (the `subid: sss` route),
// so on a host configured that way the proposal is still worth checking.
func (m *Manager) suggestSubIDRange(ctx context.Context, need int) (start, end int, computed bool) {
	out, err := engine.Output(ctx, m.R,
		runner.Cmd{Argv: []string{"sh", "-c", "cat /etc/subuid /etc/subgid 2>/dev/null"}})
	if err != nil {
		start, end = rootless.NextFreeRange("", need)
		return start, end, false
	}
	start, end = rootless.NextFreeRange(string(out), need)
	return start, end, true
}

// hostUserName resolves the account this process runs as, for the commands handed to
// an administrator. `id -un` rather than os/user, because its pure-Go fallback reads
// /etc/passwd and would miss exactly the directory accounts most likely to be here.
// The numeric uid is a usable fallback: usermod accepts it too.
func (m *Manager) hostUserName(ctx context.Context, euid int) string {
	out, err := engine.Output(ctx, m.R, runner.Cmd{Argv: []string{"id", "-un"}})
	if name := strings.TrimSpace(string(out)); err == nil && name != "" {
		return name
	}
	return strconv.Itoa(euid)
}

// runUserIDs splits a `uid[:gid]` runUser into numbers. config also accepts account
// names ("solace:solace"), which report numeric=false: those name accounts inside
// the container image, which cannot be resolved from here.
func runUserIDs(runUser string) (uid, gid int, numeric bool) {
	u, g, found := strings.Cut(runUser, ":")
	if !found {
		g = u
	}
	uid, err := strconv.Atoi(strings.TrimSpace(u))
	if err != nil {
		return 0, 0, false
	}
	gid, err = strconv.Atoi(strings.TrimSpace(g))
	if err != nil {
		return 0, 0, false
	}
	return uid, gid, true
}

// ensureLinger asserts that this user lingers, and on a mutating caller ENABLES it.
// Without it systemd tears the user's instance down at logout -- taking the broker
// with it -- and starts nothing at boot, despite the quadlet carrying
// WantedBy=default.target.
//
// This is the one row that repairs rather than refuses, and the reason is the
// privilege it needs, which is none. `loginctl enable-linger` with no username
// targets the CALLER's own account, which polkit's set-self-linger grants to an
// active session by default -- so it sits with the other unprivileged preparations
// prep already performs (mkdir, chown, podman unshare chown, podman login) rather
// than with the ones that would need root. `enable-linger <other-user>` is the
// variant that needs admin authentication, and this never issues it: the account
// is addressed by nothing at all, so it cannot name anybody else's.
//
// `validate` still only reports, because its own help promises a run that disturbs
// nothing -- hence fix, rather than a second opinion about what a ready host is.
//
// A loginctl that cannot be run at all is a failure rather than a skip: quadlet is
// a systemd generator, so a host where logind does not answer cannot run this
// deployment however the rest of the rows read. It is only reached once
// ensureUserSession has passed, so it can no longer fail merely for a missing bus
// address.
func (m *Manager) ensureLinger(ctx context.Context, euid int, user string, fix bool) error {
	r := m.report()
	out, err := engine.Output(ctx, m.R, systemd.ShowLinger(euid))
	if err != nil {
		r.Fail("linger: cannot read this user's linger state from loginctl")
		return fmt.Errorf("cannot read this user's linger state (`loginctl show-user %d --property=Linger`): %w\n"+
			"  podman quadlets are systemd units, so a host whose logind does not answer cannot run this deployment",
			euid, err)
	}
	// `Linger=yes` / `Linger=no`; anything else means a logind that answered in a
	// shape this does not read, which is reported rather than assumed good.
	state := systemd.ParseLinger(string(out))
	if state == "yes" {
		r.OK("linger: enabled for uid %d", euid)
		return nil
	}
	if !fix {
		r.Fail("linger: not enabled for uid %d (Linger=%s)", euid, state)
		// `broker deploy` enables this itself; the command is in the shared refusal
		// for anyone who wants it done without deploying. Why it matters is in
		// docs/operations.md.
		return rootless.CheckLinger(euid, state)
	}
	// No username argument: this enables it for the invoking user, which is the
	// variant that needs no privilege. Naming a user would be the one that does.
	m.logf("enabling linger for uid %d, so its user units survive logout and start at boot", euid)
	if err := engine.Run(ctx, m.R, systemd.EnableLinger()); err != nil {
		r.Fail("linger: could not enable it for uid %d", euid)
		// polkit normally grants this for your own account, so a refusal means an
		// inactive session or a tightened set-self-linger -- either way what is left
		// is the administrator's form, which is the only thing worth printing.
		return fmt.Errorf("cannot enable linger for uid %d (`loginctl enable-linger`): %w\n"+
			"  Ask an administrator to run:\n"+
			"    sudo loginctl enable-linger %s",
			euid, err, user)
	}
	r.OK("linger: enabled for uid %d (was %s)", euid, state)
	return nil
}

// checkUserSystemd proves the user's systemd instance actually answers, BEFORE the
// quadlet unit is written. Deploy's first systemctl call is daemon-reload, which
// happens after the unit is on disk -- so without this row a broken user session
// leaves an orphan unit behind and breaks Preflight's "nothing happened" promise.
func (m *Manager) checkUserSystemd(ctx context.Context) error {
	r := m.report()
	out, err := m.systemctlOutput(ctx, systemd.Systemctl(m.systemctlUser(), "show", "--property=Version"))
	if err != nil {
		r.Fail("user systemd: the user instance does not answer")
		return fmt.Errorf("this user's systemd instance does not answer (`systemctl --user show --property=Version`): %w\n"+
			"  podman quadlets are generated and started by it, so `broker deploy` would install a unit that "+
			"nothing can load. Start the user session (`systemctl --user start podman.socket`) and make it "+
			"survive logout (`loginctl enable-linger`) -- do NOT use sudo, a rootless deployment runs as this user",
			err)
	}
	r.OK("user systemd: answering (%s)", strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(string(out)), "Version=")))
	return nil
}

// checkDataDir proves the data directory can be created before PrepHost tries.
// container.dataDir defaults to /opt/solace/data on every platform, and that parent
// is root-owned, so the common first run as a non-root user fails at `mkdir -p` with
// nothing but "permission denied". runUser is not the issue here -- rootless picks a
// low id of its own -- the DIRECTORY is.
func (m *Manager) checkDataDir(ctx context.Context, user string) error {
	r := m.report()
	dir := m.Cfg.ContainerBlock(m.P).DataDir
	out, err := engine.Output(ctx, m.R, runner.Cmd{Argv: []string{"sh", "-c", nearestWritableScript, dir}})
	existing := strings.TrimSpace(string(out))
	if err == nil {
		r.OK("data dir: %s is creatable (%s is writable)", dir, existing)
		return nil
	}
	// A directory prep has already handed to the container fails the test above by
	// design: `podman unshare chown` moved it to a subuid this user cannot become.
	// Asking again INSIDE the namespace is the question that actually matters, and
	// it is what stops a correctly prepared host being flagged on every later run.
	if m.namespaceWritable(ctx, dir) {
		r.OK("data dir: %s is already owned by the container (writable in the user namespace)", dir)
		return nil
	}
	r.Fail("data dir: %s is not creatable (%s is not writable by this user)", dir, existing)
	// The chown handed over here is to THIS user, not to container.runUser: prep's
	// own `podman unshare chown` sets the in-namespace ownership afterwards, and a
	// host-side chown to the container's uid would name an id this user cannot use.
	return fmt.Errorf("container.dataDir %q cannot be created: %q is not writable by this user.\n"+
		"  Either set %s.container.dataDir to a directory you own:\n"+
		"    dataDir: ~/solace/data\n"+
		"  or ask an administrator to run:\n"+
		"    sudo mkdir -p %s && sudo chown %s %s",
		dir, existing, m.P, dir, user, dir)
}

// namespaceWritable reports whether dir is writable from inside rootless podman's
// user namespace -- that is, whether the container will be able to write there.
// Read-only, so `validate` may call it, and false for anything but rootless podman,
// where there is no namespace and the host-side answer was already the right one.
func (m *Manager) namespaceWritable(ctx context.Context, dir string) bool {
	if !m.rootlessPodman() {
		return false
	}
	return m.run(ctx, "unshare", "sh", "-c", `test -w "$0"`, dir) == nil
}

// dataDirHint is what a failed mkdir/chown adds to the engine's own message. The
// checkDataDir row above should make it unreachable in practice -- it is here for
// the race where the directory's ownership changes between the check and the
// create, and because an engine's bare "permission denied" says nothing about
// which setting to move.
//
// It ports the guidance the legacy bash carried (002-host-prep.sh) and the Go port
// dropped: dataDir defaults under /opt, which no unprivileged user can write to,
// and that default does not change when deploying rootless.
func (m *Manager) dataDirHint(dir string) string {
	if !m.rootlessPodman() {
		return ""
	}
	return fmt.Sprintf("\n  Set podman.container.dataDir to a directory you own (~/solace/data, say), "+
		"or ask an administrator to create %s and chown it to you", dir)
}

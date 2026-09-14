package broker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"solace/internal/config"
	"solace/internal/output"
)

// Ops runs the shared config/verify operations against a broker through a
// Transport. One Ops is built per command with the platform's transport injected
// (kubectl-exec for k8s, engine-exec for containers); the operations themselves
// are platform-agnostic. Log receives progress lines (routed to stderr by the
// caller so rendered stdout stays clean); a nil Log discards them.
type Ops struct {
	T   Transport
	Cfg *config.Config
	// Log is the RAW line sink for progress: one already-formatted line, emitted
	// verbatim to stderr by the caller. The `==> ` and `[TAG ] ` prefixes come
	// from the internal/output Sink built over it (progress below), never from a
	// call site. nil discards.
	Log func(format string, args ...any)
	Out io.Writer // user-facing command output; nil -> os.Stdout

	// Polling knobs for the HA state machines (leader, redundancy). New sets
	// sensible defaults; tests set PollInterval to 0 to avoid sleeping.
	PollInterval time.Duration
	PollAttempts int

	// Hostname resolves this host's name for node-local role detection
	// (LocalRole). New defaults it to os.Hostname; tests inject a fixed value.
	Hostname func() (string, error)
	// LocalAddrs reports this machine's own interface addresses, the second pass of
	// node-local role detection: a cloud host reports a name like "ip-10-0-0-11"
	// that no redundancy.*.name will ever match, and the address is then the only
	// thing the env file and the machine agree on. Defaulted in New; tests inject a
	// fixed set, since enumerating interfaces is machine state.
	LocalAddrs func() (map[string]bool, error)
	// Platform is the container platform this Ops runs against, read only by
	// the SEMP mate channel (semp.go) to resolve a bridge network's SEMP port
	// mapping. ctrOps sets it; the k8s wiring leaves it zero -- k8s never uses
	// the SEMP channel, its transport addresses either pod directly.
	Platform config.Platform
}

// New builds an Ops with default polling parameters (2s interval, 60 attempts --
// a bounded ceiling replacing the bash scripts' unbounded busy-waits) and
// os.Hostname for role detection.
func New(t Transport, cfg *config.Config, log func(string, ...any)) *Ops {
	return &Ops{
		T:            t,
		Cfg:          cfg,
		Log:          log,
		Out:          os.Stdout,
		PollInterval: 2 * time.Second,
		PollAttempts: 60,
		Hostname:     os.Hostname,
		LocalAddrs:   defaultLocalAddrs,
	}
}

// progress is the stderr Sink for this package's narration: phases through Step,
// leveled status through OK/Warn/Fail/Info. A nil Log discards.
func (o *Ops) progress() *output.Sink { return output.NewFunc(o.Log) }

// report is the stdout Sink for report bodies -- the outcome lines that belong
// with the broker output this package shows, rather than with the narration.
func (o *Ops) report() *output.Sink { return output.New(o.out()) }

// logf announces one phase of work (`==> ...`) via the injected line sink.
func (o *Ops) logf(format string, args ...any) { o.progress().Step(format, args...) }

// out returns the user-facing output sink, defaulting to os.Stdout.
func (o *Ops) out() io.Writer {
	if o.Out != nil {
		return o.Out
	}
	return os.Stdout
}

// show writes captured broker output to the user-facing sink. A write error to
// stdout is not actionable, so it is deliberately discarded.
func (o *Ops) show(b []byte) { _, _ = o.out().Write(b) }

// sleep waits PollInterval, honoring context cancellation. A zero interval (in
// tests) returns immediately without allocating a timer.
func (o *Ops) sleep(ctx context.Context) error {
	return o.wait(ctx, o.PollInterval)
}

// wait blocks for d, honoring context cancellation. A non-positive d returns
// immediately without allocating a timer.
func (o *Ops) wait(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// ErrCLIRejected is what RunCLI wraps when the BROKER refused a line, as opposed to
// the call failing to reach the broker at all. The two leave the broker in different
// states and a caller that reports them alike misinforms the operator: a rejection
// means the broker stopped exactly at that line, while an unreachable broker means
// the uploaded script may have run to completion, partly, or not at all, and nothing
// here can tell which.
//
// It is a sentinel rather than a phrase callers match on, because the only caller
// that needs the distinction (ConfigureReplication's two phases, replicationops.go)
// used to find it with strings.Contains against this very message -- so rewording the
// sentence below would silently have reclassified every rejection as unreachable,
// turning a report that names what landed into one that says it cannot be known.
// Nothing pinned that coupling, and errors.Is cannot drift.
var ErrCLIRejected = errors.New("the broker rejected a line in this script")

// RunCLI is the CLI-over-exec primitive every MUTATING config/verify step is built
// on: it runs body inside the broker through the broker's own `source script ...
// stop-on-error no-prompt`, in ONE exec, and turns a rejected line into a Go error.
//
// `cli -Apes` exits 0 even when the broker rejects a line -- a bad product key, a
// malformed certificate, a refused CA create -- so without this wrapper nothing
// about the process says whether the configuration landed. The callers this
// replaced either had no detection at all (ServerCert, DomainCerts) or scanned the
// whole transcript for a bare "error"/"fail" (ProductKeys, ExecCLI), which
// false-positives on an object legitimately named e.g. "error-events". Wrapping
// body in `source script ... stop-on-error` makes the BROKER itself stop at the
// first rejected line, so rejectionIn's tail scan is sound by construction: the
// rejection is the last thing the broker printed. That is the same mechanism
// driver.go's import applier uses, against the same failKeywords list -- one
// definition of what a rejection looks like, applied here in Go rather than a
// second copy of the grep inside the generated shell.
//
// The round trip is ONE Transport.OutputInput call: body rides stdin, exactly like
// Upload, so it never touches argv or a log; the generated skeleton (runCLISkeleton)
// writes it, writes the one-line wrapper that sources it with stop-on-error, execs
// `cli -Apes` against the wrapper, and removes all three broker-side files itself
// on exit -- so, unlike the old upload-then-exec-then-removeCLI shape, a caller
// needs no separate cleanup call. ServerCert and DomainCerts never called
// removeCLI at all and leaked a script on every run; every other caller called it
// as a second round trip. Both now get a self-cleaning single call for free.
//
// name must be a safe identifier (validName); it also seeds the broker-side
// filenames the skeleton writes (cliRunNames) as Go-baked literals rather than a
// shell-side substitution, so nothing in the skeleton is built from
// operator-controlled text except body, which never leaves stdin.
func (o *Ops) RunCLI(ctx context.Context, role config.Role, name, body string) ([]byte, error) {
	if err := validName("cli script name", name); err != nil {
		return nil, err
	}
	bodyName, wrapName, outName := cliRunNames(name)
	skeleton := runCLISkeleton(bodyName, wrapName, outName, true)
	out, err := o.T.OutputInput(ctx, role, []byte(body), "sh", "-c", skeleton)
	if err != nil {
		return out, fmt.Errorf("run cli script %q: %w", name, err)
	}
	// `cli -Apes` and the shell wrapping it both exit 0 regardless of what the
	// broker made of the script, so the transcript is the only evidence there is.
	// Sound here for the reason stated above: stop-on-error means a rejection, if
	// there is one, is the last thing in it.
	if bad := rejectionIn(out); bad != "" {
		return out, fmt.Errorf("%w: cli script %q -- the transcript carries %q; see the output above for detail",
			ErrCLIRejected, name, bad)
	}
	return out, nil
}

// runCLIRead is RunCLI's unwrapped sibling for `show`-style reads: upload body as
// the CLI's own hidden `.<name>.cli` (the old RunCLI shape, byte for byte) and run
// it directly, with no stop-on-error wrapper and no rejection scan.
//
// A read has nothing for stop-on-error to protect -- there is no later line a
// rejection could poison -- and wrapping one would change the transcript shape
// every parser here was tuned against real, unwrapped captures: bannerType,
// stripTranscript's `! BEGIN` search and driverTailLines' measured counts
// (importops.go), and the field()/countContains() scans verify_ops.go and
// verify_local.go make of `show redundancy` output. So every capture read
// (runCapture), every replication read (readMateConfig, readVPNReplication,
// replicationStatus, RouterName), and the cliMate CLIRunner behind NewLocalMate
// all go through this instead of RunCLI. The caller owns its own cleanup
// (removeCLI), exactly as before this change -- this function does not self-clean.
func (o *Ops) runCLIRead(ctx context.Context, role config.Role, name, body string) ([]byte, error) {
	if err := validName("cli script name", name); err != nil {
		return nil, err
	}
	if err := o.T.Upload(ctx, role, []byte(body), cliScriptPath(name)); err != nil {
		return nil, fmt.Errorf("upload cli script %q: %w", name, err)
	}
	out, err := o.T.Output(ctx, role, CLIBinary, "-Apes", cliArg(name))
	if err != nil {
		return out, fmt.Errorf("run cli script %q: %w", name, err)
	}
	return out, nil
}

// cliRunNames derives the three broker-side filenames RunCLI's skeleton writes and
// cleans up, all keyed off name (already validName-checked by the caller) so two
// concurrent RunCLI calls for different operations against the same broker cannot
// collide mid-write. They are Go-baked literals, always starting with a fixed
// prefix and drawing only from name's own restricted charset, never a shell-side
// substitution -- which is what keeps them out of injection range regardless of
// what name is.
func cliRunNames(name string) (body, wrap, out string) {
	const prefix = "solace-util-cli-"
	return prefix + name, prefix + "wrap-" + name, prefix + "out-" + name
}

// runCLISkeleton renders the POSIX /bin/sh script RunCLI (and ExecCLI) hand to
// `sh -c`.
//
// writeBody includes the `cat > body` step that reads the script off stdin --
// RunCLI's own shape. ExecCLI sets it false: its script is a local file already
// uploaded through UploadFile, which streams from disk rather than buffering the
// whole thing in memory, so its skeleton only has to write the wrapper, run it and
// clean up.
//
// `source script <body> stop-on-error no-prompt` is what makes the broker itself
// stop at the first rejected line -- VERIFIED against Solace's own CLI reference,
// the same command driver.go's driverWrapperName carries. `source script` resolves
// a BARE name inside the cliscripts directory, which is why body/wrap are written
// there under a plain name rather than the dot-and-.cli form `-Apes` itself takes.
func runCLISkeleton(bodyName, wrapName, outName string, writeBody bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "CLIDIR=%s\n", shQuote(CLIScriptsDir))
	fmt.Fprintf(&b, "CLI=%s\n", shQuote(CLIBinary))
	fmt.Fprintf(&b, "cleanup() { rm -f \"$CLIDIR/%s\" \"$CLIDIR/%s\" \"$CLIDIR/%s\"; }\n", bodyName, wrapName, outName)
	b.WriteString("trap cleanup EXIT INT TERM\n")
	if writeBody {
		fmt.Fprintf(&b, "cat > \"$CLIDIR/%s\"\n", bodyName)
	}
	fmt.Fprintf(&b, "printf 'source script %s stop-on-error no-prompt\\n' > \"$CLIDIR/%s\"\n", bodyName, wrapName)
	fmt.Fprintf(&b, "\"$CLI\" -Apes \"%s\" > \"$CLIDIR/%s\" 2>&1\n", wrapName, outName)
	fmt.Fprintf(&b, "cat \"$CLIDIR/%s\"\n", outName)
	return b.String()
}

// removeCLI deletes one or more uploaded `.<name>.cli` scripts, best-effort. It
// logs a warning on failure rather than aborting (cleanup is not fatal).
func (o *Ops) removeCLI(ctx context.Context, role config.Role, names ...string) {
	paths := make([]string, 0, len(names))
	for _, n := range names {
		paths = append(paths, cliScriptPath(n))
	}
	if err := o.rmPaths(ctx, role, paths...); err != nil {
		o.progress().Warn("cleanup of cli script(s) %v failed: %v", names, err)
	}
}

// removeFiles deletes uploaded files by absolute in-broker path, best-effort. It is
// removeCLI's counterpart for scripts that do NOT live in the cliscripts dir --
// ExecShellScript's, at the jail root.
func (o *Ops) removeFiles(ctx context.Context, role config.Role, paths ...string) {
	if err := o.rmPaths(ctx, role, paths...); err != nil {
		o.progress().Warn("cleanup of uploaded file(s) %v failed: %v", paths, err)
	}
}

// rmPaths is the one `rm -f` both cleanup helpers issue, so an added caller cannot
// introduce a second removal shape. Each caller words its own warning, because naming
// what was left behind is more useful than naming the paths it happened to be at.
func (o *Ops) rmPaths(ctx context.Context, role config.Role, paths ...string) error {
	if len(paths) == 0 {
		return nil
	}
	return o.T.Run(ctx, role, append([]string{"rm", "-f"}, paths...)...)
}

// skipIfStandalone reports whether an HA-gated step should no-op, returning true
// for standalone deployments (the "Standalone ... detected" branches of 050/061).
//
// It reports [SKIP], not [WARN]: running standalone is a supported choice, not a
// problem, and a warning the reader cannot act on is a warning that teaches them
// to ignore warnings.
func (o *Ops) skipIfStandalone(step string) bool {
	if o.Cfg.RedundancyEnabled() {
		return false
	}
	o.progress().Skip("%s is HA-only; standalone deployment.", step)
	return true
}

// field returns the value after the first ": " on the first output line that
// contains label, with CR stripped -- the Go form of the bash idiom
// `grep "<label>" | tr -d '\r'` followed by `${VAR#*: }`. Returns "" if absent.
func field(output, label string) string {
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimRight(raw, "\r")
		if strings.Contains(line, label) {
			if i := strings.Index(line, ": "); i >= 0 {
				return line[i+2:]
			}
			return ""
		}
	}
	return ""
}

// countContains counts output lines containing needle (CR stripped), the Go form
// of `grep <label> | grep -c <needle>` used by the redundancy activity checks.
func countContains(output, label, needle string) int {
	n := 0
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimRight(raw, "\r")
		if strings.Contains(line, label) && strings.Contains(line, needle) {
			n++
		}
	}
	return n
}

// nameRE constrains user-influenced identifiers that reach a shell or the CLI
// (cli script names, domain CA names, uploaded filenames) to a safe character
// set, the §3 boundary validation this port owns (the bash scripts had none).
var nameRE = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

func validName(kind, s string) error {
	if s == "" || strings.Contains(s, "..") || !nameRE.MatchString(s) {
		return fmt.Errorf("invalid %s %q: only letters, digits, '.', '_' and '-' are allowed", kind, s)
	}
	return nil
}

// validVPNName checks a message-VPN name before it is written into a quoted CLI
// operand (`show current-config message-vpn "<name>"`).
//
// It is deliberately NOT validName: a real broker capture carried the VPN
// "A VPN WITH LONG NAME AND SPACES", and nameRE would reject that legitimate
// name outright. So the rule is the narrower one that the quoting actually needs
// -- reject anything that could end the quoted operand or the line, and nothing
// else, because the rest of a VPN name's alphabet is Solace's to decide and not
// ours.
//
// A double quote would close the operand early and leave the remainder as bare
// CLI tokens in an already-elevated session; a control character would split the
// line into extra commands. Both are the S3 boundary the export guard owns.
// ValidVPNName exposes validVPNName to the CLI layer so a bad `--vpn` value is
// rejected as a USAGE error (exit 2) by the flag that carried it, instead of
// surfacing out of ExportConfig as a runtime failure (exit 1). The two codes exist
// so a wrapper can retry a transient cluster failure and refuse to retry a typo,
// and `--vpn "$VPN"` with VPN unset is a typo. ExportConfig still validates every
// name itself -- this is the flag's own gate, not a replacement for it.
func ValidVPNName(s string) error { return validVPNName(s) }

func validVPNName(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("invalid message-VPN name: must not be empty")
	}
	if strings.Contains(s, `"`) {
		return fmt.Errorf("invalid message-VPN name %q: it must not contain a double quote, "+
			"which would end the quoted CLI operand and run the rest as commands", s)
	}
	if i := strings.IndexFunc(s, func(r rune) bool { return r < 0x20 || r == 0x7f }); i >= 0 {
		return fmt.Errorf("invalid message-VPN name %q: contains a control character (0x%02x) at "+
			"offset %d; it must be a single line", s, s[i], i)
	}
	return nil
}

// validCLILine checks a value that is written verbatim into one line of a Solace
// CLI script. Unlike validName it does not constrain the character set -- a
// product key is an opaque vendor string and its alphabet is not ours to decide --
// but it rejects the one thing that changes the script's meaning: a control
// character. A newline would turn a single `product-key <k>` line into extra
// commands run in the already-elevated session (§3 boundary validation).
func validCLILine(kind, s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("invalid %s: must not be empty", kind)
	}
	if i := strings.IndexFunc(s, func(r rune) bool { return r < 0x20 || r == 0x7f }); i >= 0 {
		return fmt.Errorf("invalid %s %q: contains a control character (0x%02x) at offset %d; it must be a single line",
			kind, s, s[i], i)
	}
	return nil
}

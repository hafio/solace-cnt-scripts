package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"solace/internal/broker"
	"solace/internal/config"
)

// emit writes rendered bytes to stdout. Used by `gen` and `--gen` paths so the
// artifact is the command's only stdout (progress/warnings go to stderr).
func emit(b []byte) error {
	_, err := os.Stdout.Write(b)
	return err
}

// renderAnnotation marks a command that ONLY renders -- it never executes an
// external command, whatever flags it is given. `gen` is the whole set today. It
// exists so checkAllowCommand can refuse --allow-command where there is nothing to
// allow, the mirror of what genAnnotation does for the --gen-*-only trio.
const renderAnnotation = "solace_render_only"

// renderOnly tags a command as never executing, and returns it so registration can
// wrap inline: renderOnly(genCapable(...)).
func renderOnly(c *cobra.Command) *cobra.Command {
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}
	c.Annotations[renderAnnotation] = "true"
	return c
}

// withLong attaches extended help to a command built by leaf/roleLeaf/dispatchLeaf,
// which take a Short and nothing else. It returns the command so registration can
// wrap inline the way onlyOn and renderOnly do -- onlyOn(withLong(leaf(...), x), ...)
// -- rather than forcing a named variable and a separate assignment for every
// command that has something more to say than its one-line Short.
//
// A Long is a BRIEF summary -- one to four lines -- of what the command does and how
// to invoke it. Mechanics, constraints and history belong in docs/operations.md, not
// in help; a destructive command carries exactly one sentence saying what it
// destroys. The generator skips a Long equal to the Short (commanddoc_test.go), so a
// Long that would only restate the Short is not written at all.
func withLong(c *cobra.Command, long string) *cobra.Command {
	c.Long = long
	return c
}

// willExecute reports whether this invocation can reach an external command. Only
// the render-only commands cannot: they build an artifact from the env file and
// print it, touching nothing.
func (a *App) willExecute(cmd *cobra.Command) bool {
	return cmd.Annotations[renderAnnotation] != "true"
}

// addAllowCommandFlag wires --allow-command onto one command that executes. It is
// the operator's escape hatch for the execution-guard allowlist (config/execguard.go):
// a binary this tool does not drive by default -- a `microk8s kubectl`, a site
// wrapper -- runs only when the person at the keyboard names it, for that one
// invocation. It cannot approve a privilege-escalation wrapper at all: elevate this
// tool when you run it (`sudo solace-util ...`), never through an env file.
//
// It is a CLI flag and NOTHING else on purpose. There is no config key for it, no
// environment variable, and no binding layer that could give an env file a way to
// set it: an env file that could approve its own binary would make the allowlist
// decorative. wireExec adds it to each command that runs something rather than to
// root, so `solace-util convert --allow-command ...` is a usage error too.
func addAllowCommandFlag(c *cobra.Command, app *App) {
	c.Flags().StringArrayVar(&app.AllowCommand, "allow-command", nil,
		"approve one extra binary for the config's platform command, for this run only "+
			"(repeatable; a bare name, never a path). The env file cannot grant this")
	// No file completion: the value is a bare binary name, and offering paths would
	// coach exactly the mistake the help text above warns against.
	registerFlagCompletion(c, "allow-command", cobra.NoFileCompletions)
}

// checkAllowCommand rejects --allow-command on an invocation that cannot execute
// anything. Silently accepting it there would teach the flag as harmless boilerplate
// -- exactly the habit that gets it pasted into a wrapper script, where it then
// applies to runs that DO execute. Hand-rolled rather than cobra's flag groups for
// the same reason checkGenFlags is: the flag is declared on the platform command and
// validated against the leaf that inherited it, which lets the error name the leaf.
func checkAllowCommand(cmd *cobra.Command, app *App) error {
	if len(app.AllowCommand) == 0 || app.willExecute(cmd) {
		return nil
	}
	return usagef("--allow-command is only valid on a command that runs something, and %q renders "+
		"without executing; drop the flag", cmd.CommandPath())
}

// opFunc is a leaf handler that needs only the app context.
type opFunc func(*App) error

// roleOpFunc is a leaf handler parameterized by a broker node role (p|b|m).
type roleOpFunc func(*App, config.Role) error

// leaf builds a no-arg subcommand that dispatches straight to fn. It takes no
// arguments, so NoFileCompletions is what it should offer -- cobra's default is
// to fall back to filenames, which would be wrong for every command built here.
func leaf(app *App, use, short string, fn opFunc) *cobra.Command {
	return wireExec(app, &cobra.Command{
		Use:               use,
		Short:             short,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              func(*cobra.Command, []string) error { return fn(app) },
	})
}

// layer describes the one thing a removal keeps by default: the part that is
// expensive to recreate and impossible to get back. Both removals ask about theirs
// the same way, through addLayerFlags/confirmLayer, so learning the contract on one
// teaches the other.
type layer struct {
	flag  string // the flag that deletes it without asking
	what  string // what is kept, for the flag help and the report
	why   string // what deleting it costs, said at the moment of asking
	usage string // the flag's own help text
}

var (
	layerData = layer{
		flag: "delete-data",
		what: "persistent data",
		why:  "Kubernetes PVCs / the container data directory -- the broker's messages and configuration",
		usage: "delete the broker's persistent data too (Kubernetes PVCs / the container data " +
			"directory). Without it the data is kept",
	}
	layerCRD = layer{
		flag: "delete-crd",
		what: "the operator CRDs",
		why: "the CRDs are cluster-wide: deleting them cascade-deletes EVERY PubSubPlusEventBroker " +
			"in this cluster, including brokers this env file does not describe",
		usage: "delete the operator's CustomResourceDefinitions too. Without it they are kept, " +
			"so existing brokers survive",
	}
)

// addRemoveFlags wires the confirmation contract onto a command that destroys
// something. Every such command asks before it acts; --no-prompt answers yes to every
// question that is ASKED, so a script needs exactly one thing switched off rather than
// one per question.
//
// l is the retained layer, or nil for a command that has none. The two flags answer
// DIFFERENT questions and compose rather than conflict:
//
//	--delete-data / --delete-crd   RAISES the layer question. Without it the layer is
//	                               kept and no question about it is asked at all.
//	--no-prompt                    answers yes to whatever was asked.
//
// So --no-prompt alone keeps the data, because no data question was raised; a fully
// unattended wipe is `--delete-data --no-prompt`. That is the inverse of the earlier
// design, where the layer flag SILENCED a question that was always asked -- which made
// --delete-data both the request and its own confirmation.
//
// It is also why --no-prompt appears on commands that remove nothing (`broker restart`,
// `operator stop`): they ask before acting, so they need the silencer, and its help text
// therefore cannot say "the removal".
func addRemoveFlags(c *cobra.Command, app *App, l *layer) {
	c.Flags().BoolVar(&app.noPrompt, "no-prompt", false,
		"do not ask: answer yes to every question this command would ask. It does not RAISE "+
			"a question -- without a --delete-* flag the expensive layer is still kept")
	if l != nil {
		c.Flags().BoolVar(&app.deleteLayer, l.flag, false, l.usage)
	}
}

// addApplyRemoveFlags wires the direction pair for a `broker configure` leaf whose
// default is to apply: --apply is the default and accepted explicitly, --remove is the
// other way. Both at once is a usage error.
func addApplyRemoveFlags(c *cobra.Command, app *App) {
	c.Flags().BoolVar(&app.flagApply, "apply", false,
		"apply what the env file configures (the default; accepted explicitly so a script can say so)")
	c.Flags().BoolVar(&app.flagRemove, "remove", false,
		"remove it from the broker instead of applying it")
}

// addDisableEnableFlags wires the direction pair for a hardening leaf, whose default is
// to disable: --disable is the default and accepted explicitly, --enable reverses it.
func addDisableEnableFlags(c *cobra.Command, app *App) {
	c.Flags().BoolVar(&app.flagDisable, "disable", false,
		"shut it down (the default; accepted explicitly so a script can say so)")
	c.Flags().BoolVar(&app.flagEnable, "enable", false,
		"start it back up instead of shutting it down")
}

// wantRemove reports whether an --apply/--remove leaf should remove. It is the one place
// the pair is read, so two leaves cannot decide the precedence differently.
func wantRemove(a *App) (bool, error) {
	if a.flagApply && a.flagRemove {
		return false, usagef("--apply and --remove ask for opposite things; pass one or neither " +
			"(neither means --apply)")
	}
	return a.flagRemove, nil
}

// wantEnable reports whether a --disable/--enable leaf should enable.
func wantEnable(a *App) (bool, error) {
	if a.flagDisable && a.flagEnable {
		return false, usagef("--disable and --enable ask for opposite things; pass one or neither " +
			"(neither means --disable)")
	}
	return a.flagEnable, nil
}

// confirmLayer decides whether the retained layer goes with the removal.
//
//	--delete-*  --no-prompt  TTY   outcome
//	no          any          any   KEPT, nothing asked
//	yes         yes          any   deleted
//	yes         no           yes   strict prompt: an exact "yes" deletes, anything else keeps
//	yes         no           no    KEPT, with a loud warning naming --no-prompt
//
// The layer flag RAISES the question rather than silencing it, so asking for the deletion
// and confirming it stay two separate acts -- and an operator who never asked is never
// asked. Keeping is the answer in every direction that is not an explicit yes.
//
// The last row refuses the LAYER, not the command: the removal itself still proceeds. An
// abort partway through would leave a half-removed broker, which is worse than a reported
// keep, and it is the same shape confirmAction already takes on a non-TTY.
func confirmLayer(a *App, l layer) bool {
	if !a.deleteLayer {
		return false
	}
	if a.noPrompt {
		return true
	}
	yes, answered := false, false
	if interactive(a) {
		yes, answered = promptYes(promptSource(a), os.Stderr, fmt.Sprintf(
			"Also delete %s? %s.\nThis cannot be undone. Type 'yes' to delete, anything else keeps it: ",
			l.what, l.why))
	}
	if !answered {
		warn("refusing to delete %s without confirmation; pass --no-prompt to proceed -- %s is kept",
			l.what, l.what)
		return false
	}
	return yes
}

// stdinCanAnswer reports whether stdin is a stream a reply could arrive on.
//
// It asks only whether the descriptor is USABLE, never what kind it is, and that
// restraint is the whole lesson of two bugs. The original test was
// `isTTY(os.Stdin)` -- stdin is a character device -- and in a shell where stdin was
// a pipe it made every confirmation in this tool a silent refusal: `broker remove`,
// `broker restart`, the import tear-down gate and the `--out` overwrite question all
// declined without asking, on a terminal with the operator sitting at it. Reading the
// console device instead (`/dev/tty`, then `CONIN$`, the handles git and sudo reach
// for) fixed that and introduced something worse: in a mintty-style shell the
// keystrokes go into the stdin pipe and never reach the console input buffer, so the
// process blocked forever on a handle nothing would ever write to. A descriptor's
// type does not say whether a human is behind it. Only reading does.
//
// So the question is asked, the answer is read from stdin, and an EOF before any
// input is what "nobody is here" means -- decided by observation instead of by a
// guess that was wrong in both directions.
//
// The residual cost is a run whose stdin is an open pipe that nobody ever writes to
// and nobody closes: it now waits at the question rather than refusing it. That is
// what `--no-prompt` is for, and every refusal below names it.
func stdinCanAnswer() bool {
	_, err := os.Stdin.Stat()
	return err == nil
}

// interactive reports whether this run may ASK. It routes through the App's seam so
// a test can exercise the prompt branches that gate every destructive action; unset
// (the production case) it is stdinCanAnswer. Whether the question is actually
// ANSWERED is a separate fact, and only promptLine can report it.
func interactive(a *App) bool {
	if a != nil && a.Interactive != nil {
		return a.Interactive()
	}
	return stdinCanAnswer()
}

// promptSource is where a confirmation answer is read from -- the App's seam, or
// os.Stdin in production. It is stdin and nothing else: a prompt written to one
// stream and read from another is how the console-handle attempt came to hang.
func promptSource(a *App) io.Reader {
	if a == nil {
		return bufio.NewReader(os.Stdin)
	}
	// ONE reader for the whole run, memoised here.
	//
	// promptLine used to wrap the source in a fresh bufio.Reader per call and throw
	// it away, which silently ate every byte that reader had buffered past the first
	// newline. `printf 'y\nyes\n' | solace-util broker remove --delete-data` reads
	// all six bytes in one syscall, answers the delete with "y", and discards "yes\n"
	// -- so the layer confirmation that follows saw EOF, reported "unanswered", and
	// kept the data the operator had just authorised deleting. Any command asking
	// two questions in one run hit it.
	if a.promptReader == nil {
		src := io.Reader(os.Stdin)
		if a.PromptIn != nil {
			src = a.PromptIn
		}
		a.promptReader = bufio.NewReader(src)
	}
	return a.promptReader
}

// promptLine writes prompt to out and returns one trimmed line read from in.
//
// answered is false when the stream ended before any input arrived. That is NOT the
// same as a reply of "no": it means nobody was in a position to reply at all, and the
// callers report the two differently -- only the unanswered case earns the line
// naming --no-prompt. A final line with no trailing newline still counts as an answer.
func promptLine(in io.Reader, out io.Writer, prompt string) (text string, answered bool) {
	fmt.Fprint(out, prompt)
	// Reuse the caller's reader when it already is one (promptSource hands over the
	// run's single memoised reader) so nothing buffered past this line is lost; wrap
	// only a bare reader, which is what a test passing strings.NewReader gives.
	br, ok := in.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(in)
	}
	line, err := br.ReadString('\n')
	if line == "" && err != nil {
		// The prompt deliberately ends mid-line and EOF supplies no newline of its
		// own, so without this the refusal would be appended to the question.
		fmt.Fprintln(out)
		return "", false
	}
	return strings.TrimSpace(line), true
}

// promptYesNo returns true for y/yes (case-insensitive) -- the lenient form used to
// confirm a reversible delete. answered carries promptLine's meaning.
func promptYesNo(in io.Reader, out io.Writer, prompt string) (yes, answered bool) {
	line, ok := promptLine(in, out, prompt)
	if !ok {
		return false, false
	}
	switch strings.ToLower(line) {
	case "y", "yes":
		return true, true
	}
	return false, true
}

// promptYes returns true only for an exact "yes" (case-insensitive) -- the strict
// form required before an irreversible data purge.
func promptYes(in io.Reader, out io.Writer, prompt string) (yes, answered bool) {
	line, ok := promptLine(in, out, prompt)
	if !ok {
		return false, false
	}
	return strings.ToLower(line) == "yes", true
}

// confirmDelete gates every command that destroys something. --no-prompt confirms
// without asking; an interactive session is prompted [y/N]; a non-TTY without
// --no-prompt declines loudly, because nothing should be destroyed unattended
// without someone having said so on the command line.
func confirmDelete(a *App, what string) bool {
	return confirmAction(a, "Delete", "delete", what)
}

// confirmAction is confirmDelete's gate with the verb left to the caller, for the
// destructive actions that are not deletions. A container restart drops in-flight
// messaging the way a delete does and so earns the same gate -- but showing an
// operator the word "Delete" before bouncing a broker invites exactly the wrong
// conclusion about what is about to happen to their data, and a prompt that
// misdescribes its own action is worse than no prompt. title leads the question;
// lower is the same verb inside the refusal sentence.
func confirmAction(a *App, title, lower, what string) bool {
	if a.noPrompt {
		return true
	}
	yes, answered := false, false
	if interactive(a) {
		yes, answered = promptYesNo(promptSource(a), os.Stderr, fmt.Sprintf("%s %s? [y/N] ", title, what))
	}
	if !answered {
		warn("refusing to %s %s without confirmation; pass --no-prompt to proceed", lower, what)
		return false
	}
	return yes
}

// addExportFlags wires the scope flags onto `broker perform export-config`.
//
// Scope is expressed by the PRESENCE of a value rather than by a --scope enum,
// which is the same grammar --pod already uses: a value narrows, absence means the
// default set. A `--scope total|broker|vpn` enum plus a --vpn list would leave
// three overlaps to memorise (--scope vpn with no --vpn, --vpn with --scope total,
// --scope broker --vpn x); this leaves exactly one, and exportScope refuses it.
//
// --vpn takes no completer that reads the broker, and cannot: completion never
// loads the env file, which is what stops a TAB press from parsing untrusted YAML
// (TestCompletionNeverReadsTheEnvFile). So it is NoFileCompletions -- a VPN name is
// not a path, and falling back to filename completion would offer nonsense.
func addExportFlags(c *cobra.Command, app *App) {
	c.Flags().StringArrayVar(&app.vpns, "vpn", nil,
		"capture only this message-VPN, repeatable for several; omit it and every VPN is captured")
	c.Flags().BoolVar(&app.brokerOnly, "broker-only", false,
		"capture only broker-level configuration, and no message-VPN at all")
	registerFlagCompletion(c, "vpn", cobra.NoFileCompletions)
}

// exportScope reads the two scope flags as one decision, the way wantRemove and
// wantEnable already read the configure direction flags: the refusal lives here
// once rather than in each caller.
func exportScope(a *App) ([]string, bool, error) {
	if a.brokerOnly && len(a.vpns) > 0 {
		return nil, false, usagef("--broker-only excludes every message-VPN, so naming --vpn %s "+
			"asks for opposite things; pass one or neither (neither captures both)", a.vpns[0])
	}
	// The value itself is checked here, at the flag, so an unusable name is the exit
	// code that means "the request was wrong" rather than the one that means "the
	// broker failed". broker.ExportConfig checks again; this decides the exit code.
	for _, v := range a.vpns {
		if err := broker.ValidVPNName(v); err != nil {
			return nil, false, asUsage(err)
		}
	}
	return a.vpns, a.brokerOnly, nil
}

// confirmActionStrict is confirmAction with promptYes in place of promptYesNo: the
// same gate, but a lenient "y" is not enough and only an exact "yes" proceeds.
//
// It exists for one case that the three existing gates could not describe honestly.
// confirmLayer already sets this bar, and rightly -- but it gates on a.deleteLayer
// and its refusal says "refusing to DELETE", so reusing it for an import would
// misdescribe the action, which confirmAction's own comment says is worse than no
// prompt at all. And confirmAction's [y/N] is too weak here: importing over an
// existing message-VPN tears it down first, which destroys the messages spooled in
// every one of its queues. That is the same irreversible data loss --delete-data
// asks about, so it earns the same exact-"yes".
//
// --no-prompt still answers it. One silencer per command beats two flags whose
// overlap has to be memorised, so `import-config <file> --no-prompt` is what a
// fully unattended overwrite looks like -- reading exactly as
// `--delete-data --no-prompt` already does.
func confirmActionStrict(a *App, title, lower, what string) bool {
	if a.noPrompt {
		return true
	}
	yes, answered := false, false
	if interactive(a) {
		yes, answered = promptYes(promptSource(a), os.Stderr,
			fmt.Sprintf("%s %s?\nType 'yes' to proceed, anything else aborts: ", title, what))
	}
	if !answered {
		warn("refusing to %s %s without confirmation; pass --no-prompt to proceed", lower, what)
		return false
	}
	return yes
}

// addOutFlags wires --out/-o (and the --no-prompt that answers its one question) onto a
// command whose output is an artifact you keep rather than read.
//
// It exists because redirection is not portable. `broker generate > x.yaml` in Windows
// PowerShell 5.1 re-encodes this tool's plain ASCII as UTF-16LE with a BOM, and kubectl
// then refuses the file with an error that points at the YAML rather than at the shell --
// the corruption happens after this process has exited, so nothing inside it can prevent
// it. Writing the file here is the only fix, and it buys three smaller things too: the
// file appears only if the render SUCCEEDED (a shell truncates it before the command even
// runs), the secret-bearing Kubernetes stream stays out of terminal scrollback, and the
// file is created 0600 rather than inheriting whatever the shell would have used.
func addOutFlags(c *cobra.Command, app *App) {
	c.Flags().StringVarP(&app.out, "out", "o", "", "write the artifact to this file instead of stdout")
	// Meaningful only alongside --out, and accepted without it rather than refused: a
	// script that passes it unconditionally is not making a mistake worth an exit 2, and
	// nothing is silently skipped -- with no --out there is no file and no question.
	c.Flags().BoolVar(&app.noPrompt, "no-prompt", false,
		"answer yes to the --out overwrite question (no effect without --out)")
	registerFlagCompletion(c, "out", cobra.FixedCompletions(nil, cobra.ShellCompDirectiveDefault))
}

// emitOrWrite sends body to stdout, or to app.out when --out named a file.
//
// The overwrite question is asked the same way every destructive question in this tree is
// asked (confirmAction): an exact prompt when there is a terminal, yes under --no-prompt,
// and a refusal that names the flag when there is neither. `what` describes the artifact
// for the confirmation line, since "overwrite x.yaml?" is more useful when it says what is
// about to be written over it.
func emitOrWrite(a *App, body []byte, what string) error {
	if a.out == "" {
		return emit(body)
	}
	if _, err := os.Stat(a.out); err == nil {
		if !confirmAction(a, "Overwrite", "overwrite", fmt.Sprintf("%s with the %s", a.out, what)) {
			// Same verb confirmAction's own non-TTY warning uses, and no more: that line
			// has already named --no-prompt, and a declined prompt needs no explaining.
			return fmt.Errorf("refusing to overwrite %q", a.out)
		}
	}
	// 0o600 on every artifact this writes. On Kubernetes the stream carries the admin
	// password, the pre-shared key and the TLS private key base64-encoded, and a compose
	// file or quadlet unit names the secrets a broker reads -- neither is group-readable
	// material.
	if err := os.WriteFile(a.out, body, 0o600); err != nil {
		return fmt.Errorf("write %q: %w", a.out, err)
	}
	// WriteFile's perm argument is honoured only when it CREATES the file, so
	// re-exporting over an existing path would leave the artifact at whatever mode
	// that file already had -- 0644 under a default umask, world-readable, holding
	// the encrypted admin password, the config-sync pre-shared key and every client
	// secret. The chmod is what makes the 0600 above true on the overwrite path too.
	if err := os.Chmod(a.out, 0o600); err != nil {
		return fmt.Errorf("restrict %q to owner-only: %w", a.out, err)
	}
	step("wrote the %s to %s", what, a.out)
	return nil
}

// addRestartFlag wires --restart onto the deploy command. Deliberately separate
// from --no-prompt: bouncing a live broker to apply a changed artifact is its own
// explicit decision, the same way deleting its data is.
func addRestartFlag(c *cobra.Command, app *App) {
	c.Flags().BoolVar(&app.restart, "restart", false,
		"restart an already-running broker when the deploy artifact changed (otherwise you are asked, and a non-interactive run leaves it running)")
}

// confirmRestart asks whether a running broker may be bounced to apply a changed
// deploy artifact. A non-interactive session declines: the caller then leaves the
// new artifact in place and warns, so a scripted deploy never drops messaging
// traffic unattended.
// It takes the App so the prompt goes through the same seams as the other confirm
// helpers; ops_container wires it to Manager.Confirm as a closure.
func confirmRestart(a *App, question string) bool {
	if !interactive(a) {
		return false
	}
	// No warning on the unanswered path: this helper's callers already say what they
	// did instead (the artifact is left in place and the broker is not bounced), so a
	// second line here would only repeat it.
	yes, _ := promptYesNo(promptSource(a), os.Stderr, question+" [y/N] ")
	return yes
}

// confirmDowngrade gates installing an operator OLDER than the one the cluster
// already runs. Unlike confirmDelete/confirmAction, there is no --no-prompt
// escape here: `deploy operator` never registers that flag, because this
// decision is made by a human at a terminal or not at all. A non-interactive run
// therefore always declines, and the refusal says to re-run interactively rather
// than naming a flag this command does not offer -- the two versions being
// weighed were already shown by the warning printed just before this is asked.
func confirmDowngrade(a *App, question string) bool {
	yes, answered := false, false
	if interactive(a) {
		yes, answered = promptYesNo(promptSource(a), os.Stderr, question+" [y/N] ")
	}
	if !answered {
		warn("refusing to downgrade the operator without confirmation; re-run interactively to confirm the downgrade")
		return false
	}
	return yes
}

// firstArg / firstArgOr are gone with the positional arguments they read. No command
// takes a role or a platform as an argument any more -- the platform is --platform and the
// role is --pod -- and the file-path positionals that remain are read directly, since a
// missing one is an Args-validator failure rather than something to default.

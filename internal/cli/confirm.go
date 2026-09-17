package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// confirm.go is the whole prompt-and-confirmation stack, and the layer registry the
// destructive commands gate on.
//
// It is its own file because the rule is stated once, here, rather than re-derived per
// call site: whether a human is present is decided by READING, never by the kind of
// stream stdin is. The question goes to stderr, the answer is read from stdin, and an
// unanswered question is NOT a "no" -- a reply of `n` is a decision, EOF means nobody
// could make one, and only the second earns the refusal line naming --no-prompt. See
// App.Interactive/PromptIn for the seams, and stdinCanAnswer below for why the two
// obvious alternatives are both wrong.

// layer describes the one thing a removal keeps by default: the part that is expensive to
// recreate and impossible to get back. Both removals ask about theirs the same way,
// through addRemoveFlags and confirmLayer, so learning the contract on one teaches the
// other.
type layer struct {
	flag  string // the flag that deletes it without asking
	what  string // what is kept, for the flag help and the report
	why   string // what deleting it costs, said at the moment of asking
	usage string // the flag's own help text
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
	return confirmGate(a, promptYes, fmt.Sprintf(
		"Also delete %s? %s.\nThis cannot be undone. Type 'yes' to delete, anything else keeps it: ",
		l.what, l.why),
		fmt.Sprintf("refusing to delete %s without confirmation; pass --no-prompt to proceed -- %s is kept",
			l.what, l.what))
}

// confirmGate is the body every confirmation below shares: --no-prompt answers yes
// without asking, an interactive run is asked through the App's own seams, and a
// question nobody could answer refuses and says so. The callers differ ONLY in which
// prompt reader they hand in -- the lenient [y/N] or the exact-"yes" form -- and in
// the two sentences they word themselves, so the rule that decides is stated once and
// a change to it cannot land on one gate and miss another.
//
// refusal is pre-formatted rather than a format string with arguments, because what it
// describes is operator-supplied text (a file path, a container name) that can carry a
// percent sign.
//
// confirmDowngrade deliberately does NOT go through here: it has no --no-prompt escape
// at all, and the first line of this function is exactly that escape.
func confirmGate(a *App, ask func(io.Reader, io.Writer, string) (bool, bool), prompt, refusal string) bool {
	if a.noPrompt {
		return true
	}
	yes, answered := false, false
	if interactive(a) {
		yes, answered = ask(promptSource(a), os.Stderr, prompt)
	}
	if !answered {
		warn("%s", refusal)
		return false
	}
	return yes
}

// stdinCanAnswer reports whether stdin is a stream a reply could arrive on.
//
// It asks only whether the descriptor is USABLE, never what kind it is. A descriptor's
// type does not say whether a human is behind it, and the two obvious alternatives fail
// in opposite directions:
//
//   - `isTTY(os.Stdin)` refuses every confirmation in this tool, silently, in any shell
//     that hands it a pipe -- with the operator sitting at the terminal.
//   - Reading the console device instead (`/dev/tty`, then `CONIN$`, what git and sudo
//     reach for) blocks forever in a mintty-style shell, where the keystrokes go into
//     the stdin pipe and never reach the console input buffer.
//
// So the question is asked, the answer is read from stdin, and an EOF before any input
// is what "nobody is here" means.
//
// The residual cost is a run whose stdin is an open pipe that nobody ever writes to and
// nobody closes: it waits at the question rather than refusing it. That is what
// `--no-prompt` is for, and every refusal below names it.
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
	return confirmGate(a, promptYesNo, fmt.Sprintf("%s %s? [y/N] ", title, what),
		refuseUnconfirmed(lower, what))
}

// refuseUnconfirmed is the one sentence both action gates print when the question went
// unanswered. It names the verb the caller used, so the refusal describes the same act
// the question did.
func refuseUnconfirmed(lower, what string) string {
	return fmt.Sprintf("refusing to %s %s without confirmation; pass --no-prompt to proceed", lower, what)
}

// confirmActionStrict is confirmAction with promptYes in place of promptYesNo: the same
// gate, but a lenient "y" is not enough and only an exact "yes" proceeds.
//
// It exists for one case the other gates could not describe honestly. confirmLayer already
// sets this bar, and rightly -- but it gates on a.deleteLayer and its refusal says
// "refusing to DELETE", so reusing it for an import would misdescribe the action, which
// confirmAction's own comment says is worse than no prompt at all. And confirmAction's
// [y/N] is too weak here: importing over an existing message-VPN tears it down first,
// which destroys the messages spooled in every one of its queues. That is the same
// irreversible loss --delete-data asks about, so it earns the same exact "yes".
//
// --no-prompt still answers it, so `import-config <file> --no-prompt` is what a fully
// unattended overwrite looks like -- reading exactly as `--delete-data --no-prompt` does.
func confirmActionStrict(a *App, title, lower, what string) bool {
	return confirmGate(a, promptYes,
		fmt.Sprintf("%s %s?\nType 'yes' to proceed, anything else aborts: ", title, what),
		refuseUnconfirmed(lower, what))
}

// confirmRestart asks whether a running broker may be bounced to apply a changed deploy
// artifact. A non-interactive session declines: the caller then leaves the new artifact in
// place and warns, so a scripted deploy never drops messaging traffic unattended.
//
// It takes the App so the prompt goes through the same seams as the other confirm helpers;
// ops_container wires it to Manager.Confirm as a closure.
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

// No command takes a role or a platform as a positional argument: the platform is
// --platform and the role is --pod. The file-path positionals that remain are read
// directly, since a missing one is an Args-validator failure rather than a default.

var (
	layerData = layer{
		flag: "delete-data",
		what: "persistent data",
		why:  "Kubernetes PVCs / the container data directory's contents -- the broker's messages and configuration",
		usage: "delete the broker's persistent data too (Kubernetes PVCs, or the contents of the " +
			"container data directory). Without it the data is kept",
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

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
// It is its own file because the rule it implements cost two bugs to learn and is stated
// once, here, rather than re-derived per call site: whether a human is present is decided
// by READING, never by the kind of stream stdin is. The question goes to stderr, the
// answer is read from stdin, and an unanswered question is NOT a "no" -- a reply of `n` is
// a decision, EOF means nobody could make one, and only the second earns the refusal line
// naming --no-prompt. See App.Interactive/PromptIn for the seams and CLAUDE.md for the two
// bugs (a TTY check that made every confirmation a silent refusal, then a console-device
// read that hung on a handle nothing would write to).

type layer struct {
	flag  string // the flag that deletes it without asking
	what  string // what is kept, for the flag help and the report
	why   string // what deleting it costs, said at the moment of asking
	usage string // the flag's own help text
}

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

package cli

import (
	"fmt"

	"github.com/hafio/go-deploy-utility/confirm"
)

// confirm.go is the confirmation gates, built on the shared confirm package, and the
// layer registry the destructive commands gate on.
//
// It is its own file because the rule is stated once, in confirm.Prompter.Gate, rather
// than re-derived per call site: a human is present when stdin is a terminal (a
// character device). Without one, and without --no-prompt, a gate fails at once with an
// error naming --no-prompt, and a pipe is never read. A reply of `n` is a decision and
// declines quietly; end of input at a terminal is an error. See App.Interactive/PromptIn
// for the seams.

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

// prompter returns the run's single Prompter, built on first use from the App's seams.
func (a *App) prompter() *confirm.Prompter {
	if a == nil {
		return confirm.New(confirm.Options{})
	}
	if a.prompt == nil {
		a.prompt = confirm.New(confirm.Options{IsTerminal: a.Interactive, In: a.PromptIn})
	}
	return a.prompt
}

// confirmLayer decides whether the retained layer goes with the removal.
//
//	--delete-*  --no-prompt  TTY   outcome
//	no          any          any   KEPT, nothing asked
//	yes         yes          any   deleted
//	yes         no           yes   strict prompt: an exact "yes" deletes, anything else keeps
//	yes         no           no    error naming --no-prompt, before anything is deleted
//
// The layer flag RAISES the question rather than silencing it, so asking for the deletion
// and confirming it stay two separate acts -- and an operator who never asked is never
// asked. Keeping is the answer in every direction that is not an explicit yes.
//
// It is reached only after the command's own gate, so without a terminal that gate has
// already failed unless --no-prompt answered both.
func confirmLayer(a *App, l layer) (bool, error) {
	if !a.deleteLayer {
		return false, nil
	}
	return a.prompter().Gate(a.noPrompt, confirm.Strict, fmt.Sprintf(
		"Also delete %s? %s.\nThis cannot be undone. Type 'yes' to delete, anything else keeps it: ",
		l.what, l.why), "the deletion of "+l.what)
}

// confirmDelete gates every command that destroys something. --no-prompt confirms
// without asking; a terminal is prompted [y/N]; anything else is an error naming
// --no-prompt, because nothing should be destroyed unattended without someone having
// said so on the command line.
func confirmDelete(a *App, what string) (bool, error) {
	return confirmAction(a, "Delete", "delete", what)
}

// confirmAction is confirmDelete's gate with the verb left to the caller, for the
// destructive actions that are not deletions. A container restart drops in-flight
// messaging the way a delete does and so earns the same gate -- but showing an
// operator the word "Delete" before bouncing a broker invites exactly the wrong
// conclusion about what is about to happen to their data, and a prompt that
// misdescribes its own action is worse than no prompt. title leads the question;
// lower is the same verb inside the error sentence.
func confirmAction(a *App, title, lower, what string) (bool, error) {
	return a.prompter().Gate(a.noPrompt, confirm.Lenient, fmt.Sprintf("%s %s? [y/N] ", title, what),
		lower+" "+what)
}

// confirmActionStrict is confirmAction with the strict form: a lenient "y" is not
// enough and only an exact "yes" proceeds.
//
// It exists for one case the other gates could not describe honestly. confirmLayer already
// sets this bar, and rightly -- but it gates on a.deleteLayer and describes a deletion,
// so reusing it for an import would misdescribe the action, which confirmAction's own
// comment says is worse than no prompt at all. And confirmAction's [y/N] is too weak
// here: importing over an existing message-VPN tears it down first, which destroys the
// messages spooled in every one of its queues. That is the same irreversible loss
// --delete-data asks about, so it earns the same exact "yes".
//
// --no-prompt still answers it, so `import-config <file> --no-prompt` is what a fully
// unattended overwrite looks like -- reading exactly as `--delete-data --no-prompt` does.
func confirmActionStrict(a *App, title, lower, what string) (bool, error) {
	return a.prompter().Gate(a.noPrompt, confirm.Strict,
		fmt.Sprintf("%s %s?\nType 'yes' to proceed, anything else aborts: ", title, what),
		lower+" "+what)
}

// confirmRedeploy gates a container deploy that would restart or recreate a running
// broker because its artifact or a secret changed. --no-prompt approves without asking;
// a terminal is asked [y/N]; no terminal and no --no-prompt is an error naming
// --no-prompt, so a scripted deploy never drops messaging traffic unattended and never
// does so silently either. A reply of `n` is (false, nil): the deploy has written and
// restarted nothing, and says so. --restart is not read here: it pre-approves in the
// Manager, before this is asked.
//
// ops_container wires it to Manager.Confirm as a closure.
func confirmRedeploy(a *App, question string) (bool, error) {
	return a.prompter().Gate(a.noPrompt, confirm.Lenient, question+" [y/N] ", "the redeploy")
}

// confirmRestart asks whether a running Kubernetes broker may be rolled to serve a renewed
// server certificate. A session with no terminal declines: the caller then leaves the
// broker running and warns. It stays lenient where the container deploy's gate is not
// because --no-prompt is not a flag of `deploy` on Kubernetes, so an error naming it
// there would be one the operator cannot act on.
//
// It takes the App so the prompt goes through the same seams as the other confirm helpers.
func confirmRestart(a *App, question string) bool {
	p := a.prompter()
	if !p.Terminal() {
		return false
	}
	// No warning on the unanswered path: this helper's callers already say what they
	// did instead (the artifact is left in place and the broker is not bounced), so a
	// second line here would only repeat it.
	yes, _, err := p.Ask(question+" [y/N] ", confirm.Lenient)
	return err == nil && yes
}

// confirmDowngrade gates installing an operator OLDER than the one the cluster
// already runs. --no-prompt on `operator deploy` approves it, and only it: the
// two versions being weighed were already shown by the warning printed just before
// this is asked, and a run with neither a terminal nor the flag fails naming it.
func confirmDowngrade(a *App, question string) (bool, error) {
	return a.prompter().Gate(a.noPrompt, confirm.Lenient, question+" [y/N] ", "the operator downgrade")
}

// confirmScope asks whether the operator may watch EVERY namespace in the cluster. It
// deliberately ignores --no-prompt: widening the watch is a least-privilege decision a
// flag must not make for the operator, so it needs a terminal. Without one, or when the
// input ends unanswered, it is an error that names the env-file setting that chooses
// the watched namespaces instead.
func confirmScope(a *App, question string) (bool, error) {
	yes, answered, err := a.prompter().Ask(question+" [y/N] ", confirm.Lenient)
	if err == nil && !answered {
		err = confirm.Unanswered()
	}
	if err != nil {
		return false, fmt.Errorf("letting the operator watch every namespace needs a terminal, and --no-prompt does not "+
			"answer it: %w; name the namespaces to watch in kubernetes.operator.watchNamespaces in the env file instead", err)
	}
	return yes, nil
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

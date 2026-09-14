package cli

import (
	"fmt"
	"os"

	"solace/internal/broker"
	"solace/internal/config"
	"solace/internal/output"
)

// The shared bodies behind `broker perform export-config` and `import-config`.
//
// One body with two thin per-platform wrappers, the k8sLogin/ctrLogin shape: the
// flags are read in exactly one place, so the two platforms cannot disagree about
// what a scope means. The platform difference is only which entry type builds the
// broker.Ops and how the role is resolved -- both talk to the broker CLI over
// exec, so nothing else about them differs.

// runExport captures the broker's configuration and emits it as one artifact.
//
// The report goes to STDOUT only when --out has freed it. Without --out the
// artifact IS stdout, so a Section header there would corrupt a piped capture --
// while the narration on stderr is unconditional, so nothing essential is lost
// either way.
func runExport(a *App, o *broker.Ops, role config.Role) error {
	vpns, brokerOnly, err := exportScope(a)
	if err != nil {
		return err
	}
	body, err := o.ExportConfig(bg(), role, vpns, brokerOnly, nowStamp())
	if err != nil {
		return err
	}
	if a.out != "" {
		o.ExportReport(output.New(os.Stdout), vpns, brokerOnly, len(body))
	}
	return emitOrWrite(a, body, "configuration export")
}

// runImport applies an artifact back to the broker: plan, confirm, apply, verify.
//
// Plan and apply are separate calls because every confirmation lives in this
// package -- broker.Ops has no Confirm seam, deliberately -- so the question has to
// be asked between reading the target and writing to it. That is the same shape
// `operator remove` already uses: read the cluster, ask, then act.
func runImport(a *App, o *broker.Ops, role config.Role, file string) error {
	artifact, err := os.ReadFile(file)
	if err != nil {
		// The operator chose this path, so a bad one is a usage error rather than a
		// runtime failure: no retry and no different broker state would help.
		return usagef("read the configuration artifact %q: %v", file, err)
	}

	plan, err := o.ImportPlan(bg(), role, artifact)
	if err != nil {
		return err
	}
	plan.Report(output.New(os.Stdout))

	if !confirmImport(a, plan) {
		return nil
	}

	// Apply, report, and STOP if anything failed. The diff below runs only after a
	// fully successful apply.
	//
	// The report comes before the error return because it is the only statement of
	// what the broker is now holding -- which chunk failed, what the broker said
	// about it, and whether a torn-down VPN was rebuilt.
	//
	// Not verifying after a failure is deliberate. The apply stops at the failing
	// chunk, so every chunk after it never ran: a diff at that point would report
	// each of them as missing, which is true but says nothing -- dozens of rows
	// describing work that was never attempted, burying the one line that names the
	// actual failure. The operator already has the specific answer; a diff would
	// only make it harder to find.
	res, applyErr := o.ImportApply(bg(), role, plan)
	if res != nil {
		res.Report(output.New(os.Stdout))
	}
	if applyErr != nil {
		return applyErr
	}

	// A clean apply is not proof the configuration is THERE. `cli -Apes` exits 0 even
	// on a rejected line, and the driver's own check reads only the last lines of
	// each chunk -- so a line the broker refused without a recognised phrase is still
	// possible. The diff is what settles it, and its verdict is the command's exit
	// code.
	diff, err := o.ImportVerify(bg(), role, plan)
	if err != nil {
		return err
	}
	broker.RenderDiffResult(output.New(os.Stdout), diff)
	if !diff.OK() {
		return fmt.Errorf("import verification failed: %d block(s) did not match the artifact and "+
			"%d were absent from the broker -- see the report above",
			len(diff.Differing), len(diff.Missing))
	}
	return nil
}

// confirmImport asks the one question an import gets, at the strength the plan
// earns.
//
// Broker scope never asks: which broker-level sections are applied is a fixed
// classification published in docs/import.md, not a per-run decision, so there is
// nothing for an operator to answer. VPN scope asks once for the whole run --
// tearing down an existing VPN destroys the messages spooled in its queues, which
// is the irreversible loss --delete-data asks about, so it takes the same exact
// "yes". Creating VPNs the target does not have destroys nothing and takes [y/N].
func confirmImport(a *App, plan *broker.ImportPlan) bool {
	switch {
	case len(plan.Existing) > 0:
		return confirmActionStrict(a, "Tear down and rebuild", "import",
			fmt.Sprintf("%s, destroying every queue in %s and the messages spooled in them",
				plan.ExistingList(), pluralVPN(len(plan.Existing))))
	case len(plan.NewVPNs) > 0:
		return confirmAction(a, "Import", "import",
			fmt.Sprintf("%s into this broker", plan.NewList()))
	default:
		return true
	}
}

// pluralVPN keeps the confirmation sentence readable for one VPN and for several,
// rather than emitting "1 VPNs".
func pluralVPN(n int) string {
	if n == 1 {
		return "it"
	}
	return "them"
}

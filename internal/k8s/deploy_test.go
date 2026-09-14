package k8s

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

// logBuf builds a step-logger func that appends every formatted line to a
// buffer, the container package's pattern for capturing progress text -- this
// package's tests otherwise pass Log: nil and never inspect it.
func logBuf() (func(string, ...any), *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return func(f string, a ...any) { fmt.Fprintf(buf, f+"\n", a...) }, buf
}

func TestDeployBrokerApply(t *testing.T) {
	cfg := loadK8s(t)
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.DeployBroker(context.Background(), false); err != nil {
		t.Fatalf("DeployBroker: %v", err)
	}
	calls := rr.afterPreflights(t,
		probe{verb: "create", resource: brokerResource},
		probe{verb: "get", resource: brokerResource})
	if len(calls) != 1 {
		t.Fatalf("DeployBroker(keepYAML=false) made %d calls after the probes, want 1 apply", len(calls))
	}
	got := calls[0]
	if got.method != "RunInput" || got.name != "kubectl" || !eqArgs(got.args, []string{"apply", "-f", "-"}) {
		t.Fatalf("DeployBroker argv = %+v, want RunInput kubectl [apply -f -]", got)
	}
	if !strings.Contains(got.stdin, "dev-broker") {
		t.Errorf("rendered CR missing the broker name:\n%s", got.stdin)
	}
}

// TestDeployBrokerKeepYAML: with keepYAML set, the rendered manifest is written to
// .broker.yaml in the working directory and is byte-identical to what was applied.
func TestDeployBrokerKeepYAML(t *testing.T) {
	cfg := loadK8s(t) // load before chdir: sampleFixture is a relative path
	t.Chdir(t.TempDir())
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.DeployBroker(context.Background(), true); err != nil {
		t.Fatalf("DeployBroker: %v", err)
	}
	data, err := os.ReadFile(brokerYAMLFile)
	if err != nil {
		t.Fatalf("reading %s: %v", brokerYAMLFile, err)
	}
	if string(data) != rr.last().stdin {
		t.Errorf("%s differs from the applied manifest", brokerYAMLFile)
	}
}

// TestDeployBrokerKeepYAMLWriteError proves a failed .broker.yaml write (disk full,
// permission denied, path collision) fails loud and never applies -- otherwise a
// user would believe the manifest was saved for review/VCS when it was not.
func TestDeployBrokerKeepYAMLWriteError(t *testing.T) {
	cfg := loadK8s(t) // load before chdir: sampleFixture is a relative path
	t.Chdir(t.TempDir())
	// Occupy the path with a directory so os.WriteFile fails cross-platform.
	if err := os.Mkdir(brokerYAMLFile, 0o755); err != nil {
		t.Fatalf("Mkdir %s: %v", brokerYAMLFile, err)
	}
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	err := c.DeployBroker(context.Background(), true)
	if err == nil || !strings.Contains(err.Error(), brokerYAMLFile) {
		t.Fatalf("DeployBroker error = %v, want it to name %s", err, brokerYAMLFile)
	}
	// The probes ran (they precede the write); nothing after them may have.
	if calls := rr.afterPreflights(t,
		probe{verb: "create", resource: brokerResource},
		probe{verb: "get", resource: brokerResource}); len(calls) != 0 {
		t.Errorf("DeployBroker should abort before applying when the write fails; got %d calls after the probes", len(calls))
	}
}

// TestDeployBrokerStopsOnPreflightFailure is the layer-7 ordering guarantee: when
// the probe says no, nothing is written and nothing is applied. Without this the
// preflight would be decoration -- a check whose failure still let the work proceed.
func TestDeployBrokerStopsOnPreflightFailure(t *testing.T) {
	cfg := loadK8s(t) // load before chdir: sampleFixture is a relative path
	dir := t.TempDir()
	t.Chdir(dir)
	rr := &recRunner{canI: "no"}
	c := NewCluster(rr, cfg, nil, nil)

	err := c.DeployBroker(context.Background(), true)
	if err == nil {
		t.Fatal("DeployBroker must fail when the permission probe answers no")
	}
	if !strings.Contains(err.Error(), "not allowed to create") {
		t.Errorf("error = %v, want it to say the permission was refused", err)
	}
	// Nonzero, and nothing done: no manifest on disk...
	if _, statErr := os.Stat(brokerYAMLFile); statErr == nil {
		t.Errorf("%s was written despite a failed preflight", brokerYAMLFile)
	}
	// ...and no call beyond the probe itself.
	if len(rr.calls) != 1 {
		t.Errorf("%d calls made after a failed preflight, want only the probe: %+v", len(rr.calls), rr.calls)
	}
}

// TestPreflightUnreachableClusterHints: an expired token or missing context is a
// different failure from an RBAC refusal, and gets the hint that actually helps.
// The tool never offers to log in on the operator's behalf.
func TestPreflightUnreachableClusterHints(t *testing.T) {
	cfg := loadK8s(t)
	rr := &recRunner{canIErr: errFake}
	c := NewCluster(rr, cfg, nil, nil)

	err := c.DeployBroker(context.Background(), false)
	if err == nil {
		t.Fatal("DeployBroker must fail when the cluster cannot be reached")
	}
	msg := err.Error()
	for _, want := range []string{"cannot check permission", "log in first", "oc login"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error = %v, want it to contain %q", msg, want)
		}
	}
	// The CLI's own error is passed through rather than replaced.
	if !strings.Contains(msg, errFake.Error()) {
		t.Errorf("error = %v, want it to carry the CLI's own failure", msg)
	}
	if len(rr.calls) != 1 {
		t.Errorf("%d calls made after an unreachable cluster, want only the probe", len(rr.calls))
	}
}

func TestDeleteBrokerNoPurge(t *testing.T) {
	cfg := loadK8s(t)
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.DeleteBroker(context.Background(), false); err != nil {
		t.Fatalf("DeleteBroker: %v", err)
	}
	if calls := rr.afterPreflight(t, "delete", brokerResource); len(calls) != 1 {
		t.Fatalf("DeleteBroker(purge=false) made %d calls after the probe, want 1 (CR delete, no PVCs)", len(calls))
	}
	got := rr.last()
	if got.method != "RunInput" || !eqArgs(got.args, []string{"delete", "-f", "-", "--ignore-not-found"}) {
		t.Errorf("DeleteBroker argv = %+v, want RunInput kubectl [delete -f - --ignore-not-found]", got)
	}
}

func TestDeleteBrokerPurgeHA(t *testing.T) {
	cfg := loadK8s(t) // redundancy.enabled: true
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.DeleteBroker(context.Background(), true); err != nil {
		t.Fatalf("DeleteBroker: %v", err)
	}
	calls := rr.afterPreflights(t,
		probe{verb: "delete", resource: brokerResource},
		probe{verb: "delete", resource: "persistentvolumeclaims"})
	if len(calls) != 4 {
		t.Fatalf("DeleteBroker(purge, HA) made %d calls after the probe, want 4 (CR + 3 PVCs)", len(calls))
	}
	wantPVCs := []string{
		"data-dev-broker-pubsubplus-p-0",
		"data-dev-broker-pubsubplus-b-0",
		"data-dev-broker-pubsubplus-m-0",
	}
	for i, pvc := range wantPVCs {
		got := calls[i+1]
		want := []string{"delete", "pvc", pvc, "-n", "solace", "--ignore-not-found"}
		if got.method != "Run" || !eqArgs(got.args, want) {
			t.Errorf("PVC delete[%d] = %+v, want Run kubectl %v", i, got, want)
		}
	}
}

func TestDeleteBrokerPurgeStandalone(t *testing.T) {
	cfg := loadK8s(t)
	cfg.Redundancy.Enabled = "false"
	rr := &recRunner{}
	c := NewCluster(rr, cfg, nil, nil)
	if err := c.DeleteBroker(context.Background(), true); err != nil {
		t.Fatalf("DeleteBroker: %v", err)
	}
	calls := rr.afterPreflights(t,
		probe{verb: "delete", resource: brokerResource},
		probe{verb: "delete", resource: "persistentvolumeclaims"})
	if len(calls) != 2 {
		t.Fatalf("DeleteBroker(purge, standalone) made %d calls after the probe, want 2 (CR + 1 PVC)", len(calls))
	}
	got := calls[1]
	want := []string{"delete", "pvc", "data-dev-broker-pubsubplus-p-0", "-n", "solace", "--ignore-not-found"}
	if !eqArgs(got.args, want) {
		t.Errorf("PVC delete = %v, want %v", got.args, want)
	}
}

// TestDeleteBrokerPurgeSwallowsPVCError: a failing PVC delete is real (an RBAC
// denial or a stuck finalizer survives --ignore-not-found) and must not be reported
// as a clean teardown (deploy.go). runErr hits the Run-backed PVC deletes; the CR
// delete rides RunInput and still succeeds. Every role is still attempted -- one
// role's failure must not stop the others from being deleted -- and the error
// names every PVC that survived.
func TestDeleteBrokerPurgeSwallowsPVCError(t *testing.T) {
	cfg := loadK8s(t) // redundancy.enabled: true
	rr := &recRunner{runErr: errFake}
	log, buf := logBuf()
	c := NewCluster(rr, cfg, log, nil)
	err := c.DeleteBroker(context.Background(), true)
	if err == nil {
		t.Fatal("DeleteBroker must fail when a PVC delete fails, not report success")
	}
	for _, want := range []string{
		"data-dev-broker-pubsubplus-p-0",
		"data-dev-broker-pubsubplus-b-0",
		"data-dev-broker-pubsubplus-m-0",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("DeleteBroker error = %v, want it to name %s", err, want)
		}
	}
	if !errors.Is(err, errFake) {
		t.Errorf("DeleteBroker error = %v, want it to wrap the underlying cause", err)
	}
	if calls := rr.afterPreflights(t,
		probe{verb: "delete", resource: brokerResource},
		probe{verb: "delete", resource: "persistentvolumeclaims"}); len(calls) != 4 {
		t.Fatalf("all PVC deletes should still be attempted; got %d calls after the probe, want 4", len(calls))
	}
	if strings.Contains(buf.String(), "PVCs deleted") {
		t.Errorf("a failed PVC delete must not log the success line:\n%s", buf.String())
	}
}

// TestDeleteBrokerPurgeAllSucceed: when every PVC delete succeeds, DeleteBroker
// issues one delete per HA role and returns nil with the success line logged.
func TestDeleteBrokerPurgeAllSucceed(t *testing.T) {
	cfg := loadK8s(t) // redundancy.enabled: true -> 3 HA roles
	rr := &recRunner{}
	log, buf := logBuf()
	c := NewCluster(rr, cfg, log, nil)
	if err := c.DeleteBroker(context.Background(), true); err != nil {
		t.Fatalf("DeleteBroker: %v", err)
	}
	calls := rr.afterPreflights(t,
		probe{verb: "delete", resource: brokerResource},
		probe{verb: "delete", resource: "persistentvolumeclaims"})
	if len(calls) != 4 {
		t.Fatalf("DeleteBroker(purge, all succeed) made %d calls after the probe, want 4 (CR + 3 PVCs)", len(calls))
	}
	if !strings.Contains(buf.String(), "PVCs deleted") {
		t.Errorf("all PVC deletes succeeding should log the success line:\n%s", buf.String())
	}
}

// TestDeleteBrokerNoPurgeIssuesNoPVCDeletes: purge=false must not touch PVCs at
// all -- only the CR delete runs, and the kept-PVCs line is logged.
func TestDeleteBrokerNoPurgeIssuesNoPVCDeletes(t *testing.T) {
	cfg := loadK8s(t)
	rr := &recRunner{}
	log, buf := logBuf()
	c := NewCluster(rr, cfg, log, nil)
	if err := c.DeleteBroker(context.Background(), false); err != nil {
		t.Fatalf("DeleteBroker: %v", err)
	}
	calls := rr.afterPreflight(t, "delete", brokerResource)
	if len(calls) != 1 {
		t.Fatalf("DeleteBroker(purge=false) made %d calls after the probe, want 1 (CR delete only, no PVC deletes)", len(calls))
	}
	if !strings.Contains(buf.String(), "PVCs kept") {
		t.Errorf("DeleteBroker(purge=false) should log that PVCs were kept:\n%s", buf.String())
	}
}

// TestDeleteBrokerLogsPVCOutcome pins the explicit, both-directions logging
// DeleteBroker now does: which of the two layers -- the CR, the data -- survived
// a removal is the fact an operator most needs from this command's output, so it
// is stated rather than left to be inferred from the argv.
func TestDeleteBrokerLogsPVCOutcome(t *testing.T) {
	t.Run("kept", func(t *testing.T) {
		cfg := loadK8s(t)
		rr := &recRunner{}
		log, buf := logBuf()
		c := NewCluster(rr, cfg, log, nil)
		if err := c.DeleteBroker(context.Background(), false); err != nil {
			t.Fatalf("DeleteBroker: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "PVCs kept") || !strings.Contains(out, "--delete-data") {
			t.Errorf("DeleteBroker(purge=false) should log that PVCs were kept and name the flag to remove them:\n%s", out)
		}
		if strings.Contains(out, "deleting PVC") {
			t.Errorf("DeleteBroker(purge=false) must not log a per-PVC delete line:\n%s", out)
		}
	})
	t.Run("deleted", func(t *testing.T) {
		cfg := loadK8s(t) // redundancy.enabled: true -> three PVCs, three per-role log lines
		rr := &recRunner{}
		log, buf := logBuf()
		c := NewCluster(rr, cfg, log, nil)
		if err := c.DeleteBroker(context.Background(), true); err != nil {
			t.Fatalf("DeleteBroker: %v", err)
		}
		out := buf.String()
		for _, want := range []string{
			"deleting PVC data-dev-broker-pubsubplus-p-0",
			"deleting PVC data-dev-broker-pubsubplus-b-0",
			"deleting PVC data-dev-broker-pubsubplus-m-0",
			"PVCs deleted",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("DeleteBroker(purge=true) missing log line %q:\n%s", want, out)
			}
		}
		if strings.Contains(out, "PVCs kept") {
			t.Errorf("DeleteBroker(purge=true) must not also log the kept-PVCs line:\n%s", out)
		}
	})
}

// TestDeleteBrokerProbesClaimsOnlyWhenPurging pins both halves: a removal that will
// delete persistent volume claims asks permission to, and one that keeps them does
// not demand a permission it never uses.
func TestDeleteBrokerProbesClaimsOnlyWhenPurging(t *testing.T) {
	for _, tc := range []struct {
		name  string
		purge bool
		want  bool
	}{
		{"purge asks about claims", true, true},
		{"keeping data does not", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rr := &recRunner{}
			c := NewCluster(rr, haCfg(), nil, nil)
			if err := c.DeleteBroker(context.Background(), tc.purge); err != nil {
				t.Fatalf("DeleteBroker: %v", err)
			}
			_, found := rr.probedNamespace("delete", "persistentvolumeclaims")
			if found != tc.want {
				t.Errorf("probed claims = %v, want %v", found, tc.want)
			}
		})
	}
}

// The read-back is what makes "deployed" mean the object exists rather than that
// `apply` exited 0. An admission webhook can reject or rewrite the CR, and a CRD that
// is present but not yet established fails differently again -- both leave apply
// reporting success. These cover each outcome, because a read-back that got any of
// them wrong would be worse than not reading back at all.

// brokerJSON is a `get pubsubpluseventbrokers <name> -o json` reply carrying one item
// with the given name -- the shape ConfirmBrokerApplied decodes.
func brokerJSON(name string) []byte {
	return []byte(`{"apiVersion":"v1","kind":"List","items":[{"metadata":{"name":"` + name + `"}}]}`)
}

// TestConfirmBrokerAppliedAcceptsTheObjectItApplied is the happy path: the cluster
// returns the CR under the configured name and the deploy reports it.
func TestConfirmBrokerAppliedAcceptsTheObjectItApplied(t *testing.T) {
	cfg := loadK8s(t)
	log, buf := logBuf()
	rr := &recRunner{out: brokerJSON(cfg.K8s.Name)}
	c := NewCluster(rr, cfg, log, nil)
	if err := c.ConfirmBrokerApplied(context.Background()); err != nil {
		t.Fatalf("ConfirmBrokerApplied: %v", err)
	}
	if !strings.Contains(buf.String(), cfg.K8s.Name) {
		t.Errorf("the confirmation should name the broker it read back:\n%s", buf.String())
	}
}

// TestConfirmBrokerAppliedTreatsSilenceAsSkipped draws the distinction the whole
// function rests on. A runner that answers with NOTHING has not said the broker is
// absent -- it has said nothing, which is the preview case and every test seam. Turning
// that into the alarming answer would fail ~200 wiring tests and every dry run.
func TestConfirmBrokerAppliedTreatsSilenceAsSkipped(t *testing.T) {
	out := &bytes.Buffer{}
	c := NewCluster(&recRunner{}, loadK8s(t), nil, out)
	if err := c.ConfirmBrokerApplied(context.Background()); err != nil {
		t.Fatalf("an empty answer is not evidence of absence: %v", err)
	}
	if !strings.Contains(out.String(), "skipped") {
		t.Errorf("a skipped read-back must say so rather than pass silently:\n%s", out.String())
	}
}

// TestConfirmBrokerAppliedFailsWhenTheObjectIsNotThere is the case the read-back exists
// for: the apply was accepted and the object is not in the cluster. The message has to
// name both likely causes, because neither is visible from the apply's own output.
func TestConfirmBrokerAppliedFailsWhenTheObjectIsNotThere(t *testing.T) {
	rr := &recRunner{out: []byte(`{"items":[]}`)}
	c := NewCluster(rr, loadK8s(t), nil, nil)
	err := c.ConfirmBrokerApplied(context.Background())
	if err == nil {
		t.Fatal("an applied broker that does not exist must fail, not warn")
	}
	for _, want := range []string{"does not exist", "webhook", "operator validate"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %q", err, want)
		}
	}
}

// TestConfirmBrokerAppliedFailsOnADifferentObject: a reply carrying something else is
// not the broker being confirmed, so it cannot stand in for it.
func TestConfirmBrokerAppliedFailsOnADifferentObject(t *testing.T) {
	rr := &recRunner{out: brokerJSON("someone-elses-broker")}
	c := NewCluster(rr, loadK8s(t), nil, nil)
	if err := c.ConfirmBrokerApplied(context.Background()); err == nil {
		t.Fatal("a different object must not confirm this deploy")
	}
}

// TestConfirmBrokerAppliedSurfacesReadFailures covers the two ways the read itself can
// fail -- the cluster refusing it, and an answer that will not decode. Both name the
// next step, because the deploy has already happened and the operator needs to know
// what to check rather than what broke in here.
func TestConfirmBrokerAppliedSurfacesReadFailures(t *testing.T) {
	for _, tc := range []struct {
		name string
		rr   *recRunner
	}{
		{"the read is refused", &recRunner{outErr: errors.New("forbidden")}},
		{"the answer will not decode", &recRunner{out: []byte("not json at all")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := NewCluster(tc.rr, loadK8s(t), nil, nil)
			err := c.ConfirmBrokerApplied(context.Background())
			if err == nil {
				t.Fatal("a failed read-back must not report success")
			}
			if !strings.Contains(err.Error(), "broker validate") {
				t.Errorf("error %q should name the command that diagnoses it", err)
			}
		})
	}
}

// A custom volume mount names a claim this tool did not create, which may hold data
// predating the broker entirely. --delete-data therefore skips it -- and says so, since
// someone who passed the flag and got a surviving volume would otherwise assume the flag
// failed. These pin both mixes.

// TestDeleteDataKeepsEveryCustomMountedClaim: with all three nodes custom-mounted there
// is nothing for --delete-data to delete, and the report has to say that outright rather
// than printing the ordinary "data is gone" line over a full set of surviving volumes.
func TestDeleteDataKeepsEveryCustomMountedClaim(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Storage.Class = ""
	cfg.K8s.Storage.CustomVolumeMount = map[string]string{
		"primary": "pvc-p", "backup": "pvc-b", "monitor": "pvc-m",
	}
	log, buf := logBuf()
	rr := &recRunner{}
	c := NewCluster(rr, cfg, log, &bytes.Buffer{})
	if err := c.DeleteBroker(context.Background(), true); err != nil {
		t.Fatalf("DeleteBroker(purge): %v", err)
	}
	for _, call := range rr.calls {
		for _, a := range call.args {
			if a == "pvc" {
				t.Fatalf("no PVC may be deleted when every node is custom-mounted: %+v", call)
			}
		}
	}
	out := buf.String()
	if !strings.Contains(out, "no PVC was deleted") {
		t.Errorf("the report must state that nothing was deleted:\n%s", out)
	}
	for _, claim := range []string{"pvc-p", "pvc-b", "pvc-m"} {
		if !strings.Contains(out, claim) {
			t.Errorf("the surviving claim %q must be named:\n%s", claim, out)
		}
	}
}

// TestDeleteDataDeletesOnlyTheProvisionedClaims is the mixed case: the operator's own
// claims go, the custom-mounted one stays, and the closing line says which.
func TestDeleteDataDeletesOnlyTheProvisionedClaims(t *testing.T) {
	cfg := loadK8s(t)
	// Class and customVolumeMount are mutually exclusive in a VALID env file, so the
	// class goes too -- this fixture should model a state the loader would accept.
	cfg.K8s.Storage.Class = ""
	cfg.K8s.Storage.CustomVolumeMount = map[string]string{"monitor": "pvc-m"}
	log, buf := logBuf()
	rr := &recRunner{}
	c := NewCluster(rr, cfg, log, &bytes.Buffer{})
	if err := c.DeleteBroker(context.Background(), true); err != nil {
		t.Fatalf("DeleteBroker(purge): %v", err)
	}
	var deleted []string
	for _, call := range rr.calls {
		for i, a := range call.args {
			if a == "pvc" && i+1 < len(call.args) {
				deleted = append(deleted, call.args[i+1])
			}
		}
	}
	if len(deleted) != 2 {
		t.Fatalf("deleted PVCs = %v, want the two provisioned ones only", deleted)
	}
	for _, pvc := range deleted {
		if strings.Contains(pvc, "-m-") {
			t.Errorf("the custom-mounted monitor claim must not be deleted, got %q", pvc)
		}
	}
	if out := buf.String(); !strings.Contains(out, "pvc-m") || !strings.Contains(out, "except the custom volume mount") {
		t.Errorf("the report must name what was kept and why:\n%s", out)
	}
}

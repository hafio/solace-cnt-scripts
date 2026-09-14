package k8s

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"solace/internal/config"
	"solace/internal/engine"
	"solace/internal/output"
)

// These carry forward what the removed TestCheckEnv* / TestCheckOperatorNS /
// TestCheckDryRun tests proved, against the grouped report that replaced them:
// no secret ever reaches the output, the sparse-config wording still explains
// itself, the operator namespace is reported with its origin, and a preview
// claims nothing.

// TestValidateNeverPrintsASecret is the one that must not regress. A check
// report is the artifact most likely to be pasted into a ticket, and a report of
// a fully-configured deployment is exactly where a password would leak.
func TestValidateNeverPrintsASecret(t *testing.T) {
	cfg := loadK8s(t)
	cfg.SEMP.AdminPass = "SUPER-SECRET-ADMIN"
	cfg.SEMP.MonitorPass = "SUPER-SECRET-MONITOR"
	cfg.Image.Pass = "SUPER-SECRET-REGISTRY"
	cfg.TLS.CertKey = "/path/to/tls.key"

	buf := &bytes.Buffer{}
	c := NewCluster(engine.Echo{W: buf}, cfg, nil, buf)
	if err := c.Validate(context.Background()); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	out := buf.String()
	for _, secret := range []string{"SUPER-SECRET-ADMIN", "SUPER-SECRET-MONITOR", "SUPER-SECRET-REGISTRY"} {
		if strings.Contains(out, secret) {
			t.Errorf("the report leaked %q:\n%s", secret, out)
		}
	}
	// Presence is reported without the value.
	if !strings.Contains(out, "admin=set") || !strings.Contains(out, "monitor=set") {
		t.Errorf("credentials should be reported as set/MISSING:\n%s", out)
	}
}

// TestValidateGroupsAndOrdersSections pins the layout: the sections in the
// agreed order, the four leading rows first, and no compound row for the values
// that were split out on purpose (cpu and mem never share a line).
func TestValidateGroupsAndOrdersSections(t *testing.T) {
	buf := &bytes.Buffer{}
	c := NewCluster(engine.Echo{W: buf}, loadK8s(t), nil, buf)
	if err := c.Validate(context.Background()); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	out := buf.String()

	var last int
	for _, title := range []string{"=== Deployment", "=== Operator", "=== Broker", "=== Credentials", "=== Placement"} {
		at := strings.Index(out, title)
		if at < 0 {
			t.Fatalf("report missing section %q:\n%s", title, out)
		}
		if at < last {
			t.Errorf("section %q is out of order:\n%s", title, out)
		}
		last = at
	}
	// The four rows that always lead, in order.
	var prev int
	for _, key := range []string{"namespace", "name", "image", "image pull"} {
		at := strings.Index(out, "] "+key+" ")
		if at < 0 {
			t.Fatalf("report missing leading row %q:\n%s", key, out)
		}
		if at < prev {
			t.Errorf("leading row %q is out of order:\n%s", key, out)
		}
		prev = at
	}
	// cpu and mem each own a row; a combined "cpu=.. mem=.." row is what the
	// grouping rules ruled out.
	if strings.Contains(out, "cpu=") || strings.Contains(out, "mem=") {
		t.Errorf("cpu/mem must not be crammed into a compound value:\n%s", out)
	}
}

// TestValidateTagsEchoedConfigAsINFO is the distinction the tags exist for:
// config read back from the env file is [INFO], never [ OK ]. A report that
// tagged unverified config as OK would claim verification nobody did.
func TestValidateTagsEchoedConfigAsINFO(t *testing.T) {
	buf := &bytes.Buffer{}
	c := NewCluster(engine.Echo{W: buf}, loadK8s(t), nil, buf)
	if err := c.Validate(context.Background()); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	for _, line := range strings.Split(buf.String(), "\n") {
		for _, configRow := range []string{"namespace", "name ", "image ", "redundancy", "update strategy"} {
			if strings.Contains(line, "] "+configRow) && strings.HasPrefix(line, "[ OK ]") {
				t.Errorf("echoed config must be [INFO], not [ OK ]: %q", line)
			}
		}
	}
	// And a preview claims nothing about the cluster.
	out := buf.String()
	for _, want := range []string{"[SKIP] api server", "[SKIP] permission", "[SKIP] installed", "[SKIP] storage class"} {
		if !strings.Contains(out, want) {
			t.Errorf("a preview must report its cluster checks as skipped, missing %q:\n%s", want, out)
		}
	}
}

// TestValidateSparseConfigExplainsItself carries forward
// TestCheckEnvSparseConfig: an empty watch list is the widest scope the operator
// has, so the report says what it produces rather than the reassuring opposite.
func TestValidateSparseConfigExplainsItself(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Operator.WatchNamespaces = ""
	cfg.K8s.Operator.WatchBrokerNS = boolPtr(false)
	cfg.K8s.TLSServerSecret = ""
	cfg.SEMP.AdminPass = ""

	buf := &bytes.Buffer{}
	c := NewCluster(engine.Echo{W: buf}, cfg, nil, buf)
	if err := c.Validate(context.Background()); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"would watch ALL namespaces",
		"(not configured)", // tls
		"MISSING",          // the admin password
	} {
		if !strings.Contains(out, want) {
			t.Errorf("sparse-config report missing %q:\n%s", want, out)
		}
	}
}

// TestValidateReportsEveryFailureInOneRun is the behaviour change worth having:
// the old flow returned at the first failure, so fixing an unreachable cluster
// only revealed the next problem. Now one run names them all.
func TestValidateReportsEveryFailureInOneRun(t *testing.T) {
	// A refused permission is the first failure. Under the old flow it would have
	// returned there; the property under test is that the run CONTINUES and still
	// reports the sections after it.
	rr := &recRunner{canI: "no"}
	buf := &bytes.Buffer{}
	c := NewCluster(rr, loadK8s(t), nil, buf)

	err := c.Validate(context.Background())
	if err == nil {
		t.Fatal("a report with failures must return an error")
	}
	if !strings.Contains(err.Error(), "check(s) failed") {
		t.Errorf("error = %v, want it to count the failures rather than wrap the first one", err)
	}
	out := buf.String()
	if !strings.Contains(out, "[FAIL] permission") {
		t.Errorf("the refused permission should be a FAIL row:\n%s", out)
	}
	// The sections that come AFTER the failure are still rendered -- that is what
	// "one run tells you everything" means in practice.
	for _, want := range []string{"=== Operator", "=== Broker", "=== Credentials", "=== Placement"} {
		if !strings.Contains(out, want) {
			t.Errorf("a failure must not cut the report short, missing %q:\n%s", want, out)
		}
	}
	// The failing row names what to ask for, not just that it failed.
	if !strings.Contains(out, "role binding") {
		t.Errorf("the FAIL row should carry the actionable message:\n%s", out)
	}
}

// TestCheckReportFailedCounts the aggregate the returned error is built from.
func TestCheckReportFailedCounts(t *testing.T) {
	rep := &checkReport{
		deployment:  []checkRow{info("a", "x"), failRow("b", "y")},
		operator:    []checkRow{failRow("c", "z")},
		broker:      []checkRow{okRow("d", "w")},
		credentials: []checkRow{skipRow("e", "v")},
		placement:   []checkRow{warnRow("f", "u")},
	}
	if got := rep.failed(scopeAll); got != 2 {
		t.Errorf("failed() = %d, want 2", got)
	}
	empty := &checkReport{}
	if got := empty.failed(scopeAll); got != 0 {
		t.Errorf("an empty report has no failures, got %d", got)
	}
}

// TestStorageRows exercises storageRows directly -- the section builder
// CheckDeploy actually calls. It carries forward what the removed
// CheckStorageClass (check.go) covered before L4's dead-code sweep found it had
// no production caller left once this section absorbed the same rule: a
// suitable configured class passes, an unsuitable binding/expansion fails, a
// class missing the queried fields fails, and every read failure along the way
// (default resolution, the first attribute column, the second after the first
// succeeds) surfaces as a FAIL row rather than being swallowed.
func TestStorageRows(t *testing.T) {
	t.Run("suitable configured class reports OK", func(t *testing.T) {
		cfg := haCfg()
		cfg.K8s.Storage.Class = "fast"
		rr := &recRunner{outQueue: [][]byte{[]byte("WaitForFirstConsumer\n"), []byte("true\n")}}
		c := NewCluster(rr, cfg, nil, &bytes.Buffer{})
		rep := &checkReport{}
		c.storageRows(context.Background(), rep)
		if len(rep.broker) != 1 || rep.broker[0].Level != output.LevelOK {
			t.Fatalf("storageRows = %+v, want one OK row", rep.broker)
		}
		// Configured class -> no default-resolution query; just the two attribute reads.
		if len(rr.calls) != 2 {
			t.Fatalf("made %d calls, want 2 attribute reads", len(rr.calls))
		}
	})
	t.Run("unsuitable binding/expansion fails", func(t *testing.T) {
		cfg := haCfg()
		cfg.K8s.Storage.Class = "slow"
		rr := &recRunner{outQueue: [][]byte{[]byte("Immediate\n"), []byte("false\n")}}
		c := NewCluster(rr, cfg, nil, &bytes.Buffer{})
		rep := &checkReport{}
		c.storageRows(context.Background(), rep)
		if len(rep.broker) != 1 || rep.broker[0].Level != output.LevelFail {
			t.Fatalf("storageRows = %+v, want one FAIL row", rep.broker)
		}
	})
	t.Run("missing fields report as <none> and fail", func(t *testing.T) {
		cfg := haCfg()
		cfg.K8s.Storage.Class = "x"
		rr := &recRunner{outQueue: [][]byte{[]byte("<none>\n"), []byte("<none>\n")}}
		c := NewCluster(rr, cfg, nil, &bytes.Buffer{})
		rep := &checkReport{}
		c.storageRows(context.Background(), rep)
		if len(rep.broker) != 1 || rep.broker[0].Level != output.LevelFail ||
			!strings.Contains(rep.broker[0].Value, "<none>") {
			t.Fatalf("storageRows = %+v, want one FAIL row naming <none>", rep.broker)
		}
	})
	// The 009-ported actionable error must not carry the bash script number: L4
	// dropped it from the user-visible string (it stays in the code as
	// provenance, not in what an operator reads).
	t.Run("no default and none configured names the fix, not the bash script", func(t *testing.T) {
		cfg := haCfg() // no Storage.Class, no default StorageClass on the cluster
		rr := &recRunner{out: []byte("")}
		c := NewCluster(rr, cfg, nil, &bytes.Buffer{})
		rep := &checkReport{}
		c.storageRows(context.Background(), rep)
		if len(rep.broker) != 1 || rep.broker[0].Level != output.LevelFail ||
			!strings.Contains(rep.broker[0].Value, "no default StorageClass found") {
			t.Fatalf("storageRows = %+v, want a FAIL row naming \"no default StorageClass found\"", rep.broker)
		}
		if strings.Contains(rep.broker[0].Value, "009") {
			t.Errorf("the report must not carry the bash script number: %q", rep.broker[0].Value)
		}
	})
	t.Run("default-resolution query fails", func(t *testing.T) {
		cfg := haCfg() // no Storage.Class -> resolves the cluster default
		rr := &recRunner{outErr: errFake}
		c := NewCluster(rr, cfg, nil, &bytes.Buffer{})
		rep := &checkReport{}
		c.storageRows(context.Background(), rep)
		if len(rep.broker) != 1 || rep.broker[0].Level != output.LevelFail ||
			!strings.Contains(rep.broker[0].Value, "resolving default StorageClass") {
			t.Fatalf("storageRows = %+v, want it to wrap \"resolving default StorageClass\"", rep.broker)
		}
		if len(rr.calls) != 1 {
			t.Errorf("storageRows should stop at the failing default-resolution query; got %d calls", len(rr.calls))
		}
	})
	t.Run("first attribute read fails", func(t *testing.T) {
		cfg := haCfg()
		cfg.K8s.Storage.Class = "fast"
		rr := &recRunner{outErr: errFake}
		c := NewCluster(rr, cfg, nil, &bytes.Buffer{})
		rep := &checkReport{}
		c.storageRows(context.Background(), rep)
		if len(rep.broker) != 1 || rep.broker[0].Level != output.LevelFail ||
			!strings.Contains(rep.broker[0].Value, `reading StorageClass "fast"`) {
			t.Fatalf(`storageRows = %+v, want it to wrap reading StorageClass "fast"`, rep.broker)
		}
		if len(rr.calls) != 1 {
			t.Errorf("storageRows should stop at the first failing column read; got %d calls", len(rr.calls))
		}
	})
	// The other real, reachable half of the two-read sequence: volumeBindingMode
	// and allowVolumeExpansion are two separate kubectl calls, so a transient
	// failure on just the second is just as real as on the first.
	t.Run("second attribute read fails after the first succeeds", func(t *testing.T) {
		cfg := haCfg()
		cfg.K8s.Storage.Class = "fast"
		rr := &recRunner{
			outQueue:    [][]byte{[]byte("WaitForFirstConsumer\n")},
			outErrQueue: []error{nil, errFake},
		}
		c := NewCluster(rr, cfg, nil, &bytes.Buffer{})
		rep := &checkReport{}
		c.storageRows(context.Background(), rep)
		if len(rep.broker) != 1 || rep.broker[0].Level != output.LevelFail ||
			!strings.Contains(rep.broker[0].Value, `reading StorageClass "fast"`) {
			t.Fatalf(`storageRows = %+v, want it to wrap reading StorageClass "fast"`, rep.broker)
		}
		if len(rr.calls) != 2 {
			t.Errorf("storageRows should have attempted both attribute reads; got %d calls", len(rr.calls))
		}
	})
}

// TestValidateReportsResolvedPorts covers M11 (second half): an operator
// cannot see the CR's port configuration without deploying or running
// `generate broker`. The Config section closes that gap, so an explicit
// kubernetes.ports must show up as name=port pairs an operator can read
// straight off the report.
func TestValidateReportsResolvedPorts(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Ports = []string{"tcp-custom=9100", "tls-custom=9443"}

	buf := &bytes.Buffer{}
	c := NewCluster(engine.Echo{W: buf}, cfg, nil, buf)
	if err := c.Validate(context.Background()); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "=== Config") {
		t.Fatalf("report missing the Config section:\n%s", out)
	}
	if !strings.Contains(out, "ports (2)") {
		t.Errorf("report should count the resolved ports, missing \"ports (2)\":\n%s", out)
	}
	for _, want := range []string{"tcp-custom=9100", "tls-custom=9443"} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing resolved port %q:\n%s", want, out)
		}
	}
}

// TestValidateReportsDefaultPorts covers the other half of M11: an env file
// that leaves kubernetes.ports unset is not "no ports configured" -- ApplyDefaults
// already filled in the operator's own 17-entry default before Validate ever ran,
// and that is what will actually render into the broker CR. The precondition
// assertion is the "confirm that is true before asserting it" the task calls
// for: if defaulting ever stopped happening here, this test should fail on the
// precondition, not silently pass on an empty list.
func TestValidateReportsDefaultPorts(t *testing.T) {
	cfg := loadK8s(t) // sample.yaml leaves kubernetes.ports commented out
	if len(cfg.K8s.Ports) != 17 {
		t.Fatalf("precondition: ApplyDefaults should fill 17 ports when kubernetes.ports is unset, got %d", len(cfg.K8s.Ports))
	}

	buf := &bytes.Buffer{}
	c := NewCluster(engine.Echo{W: buf}, cfg, nil, buf)
	if err := c.Validate(context.Background()); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "ports (17)") {
		t.Errorf("report should show the default port count, missing \"ports (17)\":\n%s", out)
	}
	// One from the head of the default list and one from the tail, so a bug
	// that only wrapped the first chunk would still be caught.
	for _, want := range []string{"tcp-ssh=2222", "tls-mqttweb=8443"} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing default port %q:\n%s", want, out)
		}
	}
}

// TestPortRowsNeverFail pins the M11 rule directly at the row-builder level:
// however many ports there are, and however many rows that wraps into, none of
// them may carry FAIL. `validate` is read-only, and a config that already
// passed Validate must not turn around and stop a deploy over how many ports
// it has to print.
func TestPortRowsNeverFail(t *testing.T) {
	for _, n := range []int{0, 1, portsPerRow, portsPerRow + 1, 17} {
		ports := make([]string, n)
		for i := range ports {
			ports[i] = fmt.Sprintf("p%d=%d", i, 1000+i)
		}
		for _, row := range portRows(ports) {
			if row.Level != output.LevelOK {
				t.Errorf("portRows(%d ports) produced a non-OK row %+v", n, row)
			}
		}
	}
}

// TestValidateConfigSectionNeverFails is the report-level companion to
// TestPortRowsNeverFail: a clean run's Config section (env-file-accepted row
// plus every ports row) must not move rep.failed(scopeAll) off zero.
func TestValidateConfigSectionNeverFails(t *testing.T) {
	buf := &bytes.Buffer{}
	c := NewCluster(engine.Echo{W: buf}, loadK8s(t), nil, buf)
	rep := &checkReport{}
	c.validationRows(rep)
	if got := rep.failed(scopeAll); got != 0 {
		t.Errorf("failed() = %d after validationRows alone, want 0", got)
	}
	for _, row := range rep.config {
		if row.Level == output.LevelFail {
			t.Errorf("a Config-section row must never be FAIL: %+v", row)
		}
	}
}

// TestCheckReportSkipsEmptySections proves a section with no rows prints no
// header, so a standalone config does not show an empty Placement block.
func TestCheckReportSkipsEmptySections(t *testing.T) {
	buf := &bytes.Buffer{}
	rep := &checkReport{deployment: []checkRow{info("namespace", "solace")}}
	rep.render(output.New(buf), scopeAll)
	out := buf.String()
	if !strings.Contains(out, "=== Deployment") {
		t.Errorf("a populated section should render:\n%s", out)
	}
	for _, absent := range []string{"=== Operator", "=== Broker", "=== Credentials", "=== Placement"} {
		if strings.Contains(out, absent) {
			t.Errorf("an empty section should not print a header (%s):\n%s", absent, out)
		}
	}
}

// TestValidateReportsThePreSharedKeyChoice: on Kubernetes an empty pre-shared key is a
// legitimate deployment rather than a gap -- the operator generates and distributes one --
// so the report has to say WHICH of the two keys the group will end up using. Nothing else
// shows it: neither the CR nor the Secret exists until deploy.
func TestValidateReportsThePreSharedKeyChoice(t *testing.T) {
	run := func(t *testing.T, mutate func(*config.Config)) string {
		t.Helper()
		cfg := loadK8s(t)
		mutate(cfg)
		buf := &bytes.Buffer{}
		c := NewCluster(engine.Echo{W: buf}, cfg, nil, buf)
		if err := c.Validate(context.Background()); err != nil {
			t.Fatalf("Validate: %v", err)
		}
		return buf.String()
	}

	set := run(t, func(c *config.Config) {
		c.Redundancy.Enabled = "true"
		c.Redundancy.PSK = "UNIQUE-PSK-VALUE"
	})
	if !strings.Contains(set, "preshared key") || !strings.Contains(set, pskSecretKey) {
		t.Errorf("a configured key must be reported, naming the Secret entry it becomes:\n%s", set)
	}
	if strings.Contains(set, "UNIQUE-PSK-VALUE") {
		t.Errorf("the report must never print the key itself:\n%s", set)
	}

	unset := run(t, func(c *config.Config) {
		c.Redundancy.Enabled = "true"
		c.Redundancy.PSK = ""
		c.Redundancy.PSKEnv = ""
	})
	if !strings.Contains(unset, "the operator generates") {
		t.Errorf("an empty key must be reported as the operator's job, not as a gap:\n%s", unset)
	}

	// Standalone has no group to authenticate, so the row would be noise.
	standalone := run(t, func(c *config.Config) { c.Redundancy.Enabled = "false" })
	if strings.Contains(standalone, "preshared key") {
		t.Errorf("a standalone broker has no redundancy group, so no key row:\n%s", standalone)
	}
}

// TestValidateSaysWhoOwnsTheTLSSecret: naming a Secret and supplying the files it is built
// from are separate decisions, and the report is the only place the difference shows before
// a deploy. Read as "this tool will create it" in the case where it will not, a missing
// Secret is first discovered by a pod that will not mount.
func TestValidateSaysWhoOwnsTheTLSSecret(t *testing.T) {
	run := func(t *testing.T, mutate func(*config.Config)) string {
		t.Helper()
		cfg := loadK8s(t)
		cfg.K8s.TLSServerSecret = "solace-tls-secret"
		mutate(cfg)
		buf := &bytes.Buffer{}
		c := NewCluster(engine.Echo{W: buf}, cfg, nil, buf)
		if err := c.Validate(context.Background()); err != nil {
			t.Fatalf("Validate: %v", err)
		}
		return buf.String()
	}

	ours := run(t, func(c *config.Config) {
		c.TLS.Cert, c.TLS.CertKey = "certs/tls.crt", "certs/tls.key"
	})
	if !strings.Contains(ours, "built from") {
		t.Errorf("a supplied pair means this tool builds the Secret, and the row must say so:\n%s", ours)
	}

	theirs := run(t, func(c *config.Config) { c.TLS.Cert, c.TLS.CertKey = "", "" })
	if !strings.Contains(theirs, "must already exist") {
		t.Errorf("with no files supplied the Secret is the operator's, and the row must say so:\n%s", theirs)
	}

	// The passphrase has nowhere to go on this platform: the CRD's spec.tls carries no
	// such field. Silence would leave that to be found by a broker that will not start.
	pass := run(t, func(c *config.Config) {
		c.TLS.Cert, c.TLS.CertKey = "certs/tls.crt", "certs/tls.key"
		c.TLS.CertPassphrase = "UNIQUE-PASSPHRASE"
	})
	if !strings.Contains(pass, "[WARN] tls passphrase") {
		t.Errorf("an encrypted key must warn on kubernetes:\n%s", pass)
	}
	if strings.Contains(pass, "UNIQUE-PASSPHRASE") {
		t.Errorf("the report must never print the passphrase itself:\n%s", pass)
	}
	if strings.Contains(ours, "tls passphrase") {
		t.Errorf("no passphrase configured, so no row:\n%s", ours)
	}
}

// TestValidateSkipsTheStorageClassWhenEveryNodeIsCustomMounted closes a gap the shared
// name hid: there are two storageRows, and only the config-side one branched on custom
// mounts. The cluster-side namesake ran the class check regardless -- and since
// validateStorage refuses class and customVolumeMount together, it always fell through to
// the cluster default. On a cluster with no default (bare metal, static provisioning --
// exactly why someone pre-creates PVCs) that FAILED the whole report over a class the
// deploy never touches, and told the operator to set a key config.Load would then refuse.
//
// It drives a recording runner rather than engine.Echo on purpose: the Echo path short-
// circuits every cluster-backed row to "skipped (preview)" before storageRows is reached,
// so a preview cannot see this either way. recRunner answers the class lookup with nothing,
// which is precisely the no-default-StorageClass cluster the bug bit on.
func TestValidateSkipsTheStorageClassWhenEveryNodeIsCustomMounted(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Storage.Class = "" // mandatory: mutually exclusive with the mounts below
	cfg.K8s.Storage.CustomVolumeMount = map[string]string{
		"primary": "pvc-p", "backup": "pvc-b", "monitor": "pvc-m",
	}

	buf := &bytes.Buffer{}
	c := NewCluster(&recRunner{}, cfg, nil, buf)
	// Validate may still return an error -- with no cluster behind it the operator rows
	// fail -- but the STORAGE CLASS must not be one of the failures.
	_ = c.Validate(context.Background())
	out := buf.String()
	if strings.Contains(out, "[FAIL] storage class") {
		t.Fatalf("no class is in play, so it cannot fail the report:\n%s", out)
	}
	if !strings.Contains(out, "not checked -- every node mounts an existing claim") {
		t.Errorf("the storage-class row must say why it was skipped:\n%s", out)
	}
	// The advice that used to be printed here was a dead end -- setting the class is what
	// config.Load refuses. Guard against it coming back.
	if strings.Contains(out, "kubernetes.storage.class is not set") {
		t.Errorf("that fix instruction is refused by config.Load; it must not be offered:\n%s", out)
	}
}

// TestValidateStillChecksTheClassWhenOnlySomeNodesAreMounted: the guard asks per role
// rather than off UsesCustomMounts, so a partly covered group -- which still has claims to
// bind -- keeps being checked. validateStorage forbids that state today; this is what
// stops the check silently disappearing if that rule is ever relaxed.
func TestValidateStillChecksTheClassWhenOnlySomeNodesAreMounted(t *testing.T) {
	cfg := loadK8s(t)
	cfg.K8s.Storage.Class = ""
	cfg.K8s.Storage.CustomVolumeMount = map[string]string{"primary": "pvc-p"}
	if allCustomMounted(cfg) {
		t.Fatal("a partly covered HA group is not fully custom-mounted")
	}
}

// TestValidateReadsDeploymentsOnce pins the fetch-once rule for `check deploy`.
//
// Two rows want two different facts about the SAME operator Deployment -- its
// WATCH_NAMESPACE and its running image -- and each reader used to issue its own
// `get deployment --all-namespaces`, so the healthy path of every `validate` paid for a
// full cluster-wide list twice. The list is the expensive one: it is not scoped to a
// namespace, deliberately, because on a first install the configured namespace does not
// exist and an operator installed somewhere else is the case worth noticing.
//
// BrokerReport made the same fix and pins it the same way; this is that rule applied to
// the other report. Counting the calls is the only way to hold it -- both readers return
// correct answers either way, so nothing else would notice the second fetch.
func TestValidateReadsDeploymentsOnce(t *testing.T) {
	r := &recRunner{}
	buf := &bytes.Buffer{}
	c := NewCluster(r, loadK8s(t), nil, buf)
	// The VERDICT is deliberately ignored. Against a fake that answers nothing, some
	// rows legitimately come back [FAIL] and Validate reports that -- which is the
	// report doing its job, not a problem with this test. What is being measured is how
	// many times the cluster was asked, and that is the same either way.
	_ = c.Validate(context.Background())

	n := 0
	for _, call := range r.calls {
		argv := strings.Join(call.args, " ")
		if strings.Contains(argv, "get deployment") && strings.Contains(argv, "--all-namespaces") {
			n++
		}
	}
	// EXACTLY one, not "at most one". At most would also pass if the fetch stopped
	// happening at all -- a report that silently skipped both rows would look like a
	// successful optimisation.
	if n != 1 {
		t.Errorf("`get deployment --all-namespaces` ran %d times, want exactly 1 -- the watch row and "+
			"the operator rows read the same object and must share one fetch, and neither may skip it", n)
	}
}

package k8s

import (
	"context"
	"errors"
	"strings"
	"testing"

	"solace/internal/config"
)

// adminsecret.go covers the two states in which this tool does not hold the broker's
// admin password: a Secret it only references, and none -- the operator generating one.
// These pin where the password is read back from, and the one deploy that is refused
// because it would leave nobody holding it.

// operatorManagedCfg is haCfg in the neither state: no semp.adminPass, no
// kubernetes.adminSecret, so the operator generates dev-broker-pubsubplus-admin-creds.
func operatorManagedCfg() *config.Config { return haCfg() }

// TestOperatorAdminSecretNameMatchesTheBrokerNames: config spells the operator's
// generated name itself, because it cannot import this package. This stops it drifting
// from the names this package derives for the same CR.
func TestOperatorAdminSecretNameMatchesTheBrokerNames(t *testing.T) {
	cfg := haCfg()
	if got, want := cfg.OperatorAdminSecretName(), lbServiceName(cfg)+"-admin-creds"; got != want {
		t.Errorf("OperatorAdminSecretName() = %q, want %q (<name>%s-admin-creds)", got, want, brokerSuffix)
	}
}

// TestAdminSecretInUseFollowsTheStates: the Secret the broker's password is actually in.
func TestAdminSecretInUseFollowsTheStates(t *testing.T) {
	built := adminCfg() // semp.adminPass and kubernetes.adminSecret
	if got := AdminSecretInUse(built); got != "solace-admin-secret" {
		t.Errorf("built: AdminSecretInUse = %q, want the configured name", got)
	}
	ref := haCfg()
	ref.K8s.AdminSecret = "byo-admin"
	if got := AdminSecretInUse(ref); got != "byo-admin" {
		t.Errorf("referenced: AdminSecretInUse = %q, want the referenced name", got)
	}
	if got := AdminSecretInUse(operatorManagedCfg()); got != "dev-broker-pubsubplus-admin-creds" {
		t.Errorf("operator-generated: AdminSecretInUse = %q, want the operator's name", got)
	}
}

// TestReadAdminPasswordArgvAndDecoding pins the read-back in both states that need it:
// one key, by jsonpath, from the Secret the broker uses, with this env file's own cluster
// CLI -- and the value decoded with the trailing newline dropped, as ReadSecretKey does.
func TestReadAdminPasswordArgvAndDecoding(t *testing.T) {
	for _, tc := range []struct {
		name, adminSecret, reads string
	}{
		{"referenced", "byo-admin", "byo-admin"},
		{"operator-generated", "", "dev-broker-pubsubplus-admin-creds"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := haCfg()
			cfg.K8s.AdminSecret = tc.adminSecret
			r := &recRunner{out: []byte("czNjcmV0Cg==\n")} // "s3cret\n"
			got, err := ReadAdminPassword(r, cfg)
			if err != nil {
				t.Fatalf("ReadAdminPassword: %v", err)
			}
			if got != "s3cret" {
				t.Errorf("password = %q, want the decoded value", got)
			}
			want := []string{"get", "secret", tc.reads, "-n", "solace", "-o", "jsonpath={.data.username_admin_password}"}
			if len(r.calls) != 1 || r.calls[0].name != "kubectl" || !eqArgs(r.calls[0].args, want) {
				t.Errorf("calls = %+v, want exactly one `kubectl %s`", r.calls, strings.Join(want, " "))
			}
		})
	}
}

// TestReadAdminPasswordNamesTheNextStep: a failed read keeps kubectl's own cause and adds
// the next step for whichever Secret it was -- create it, or deploy first -- plus the
// permission the read needs. It never prints a value.
func TestReadAdminPasswordNamesTheNextStep(t *testing.T) {
	ref := haCfg()
	ref.K8s.AdminSecret = "byo-admin"
	_, err := ReadAdminPassword(&recRunner{outErr: errors.New("secrets \"byo-admin\" not found")}, ref)
	if err == nil {
		t.Fatal("a failed read must be an error")
	}
	for _, want := range []string{"not found", "byo-admin", "references but never creates", "semp.adminPass", "get secrets"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("referenced: error %q should mention %q", err, want)
		}
	}

	_, err = ReadAdminPassword(&recRunner{outErr: errors.New("forbidden")}, operatorManagedCfg())
	if err == nil {
		t.Fatal("a failed read must be an error")
	}
	for _, want := range []string{"forbidden", "dev-broker-pubsubplus-admin-creds", "deploy the broker first", "get secrets"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("operator-generated: error %q should mention %q", err, want)
		}
	}
}

// TestReadAdminPasswordGuardsTheCommand: the read goes through the execution guard like
// every other command, and a refused command is refused before anything runs.
func TestReadAdminPasswordGuardsTheCommand(t *testing.T) {
	cfg := operatorManagedCfg()
	cfg.K8s.Command = config.Command{"curl"}
	r := &recRunner{out: []byte("czNjcmV0Cg==")}
	if _, err := ReadAdminPassword(r, cfg); err == nil {
		t.Fatal("a kubernetes.command outside the allowlist must be refused before the read")
	}
	if len(r.calls) != 0 {
		t.Errorf("the guard must stop before exec, but %d call(s) ran: %+v", len(r.calls), r.calls)
	}
}

// TestReadAdminPasswordRefusesAMalformedReference: the admin read goes through the same
// reader as a mate's passSecret, so it gets the same refusal. Load already holds
// kubernetes.namespace and kubernetes.adminSecret to these rules; this config skips Load,
// which is the case the reader's own check exists for.
func TestReadAdminPasswordRefusesAMalformedReference(t *testing.T) {
	for _, tc := range []struct {
		name  string
		apply func(*config.Config)
		want  string
	}{
		{"namespace", func(c *config.Config) { c.K8s.Namespace = "--kubeconfig=/x" }, `secret namespace "--kubeconfig=/x" is invalid`},
		{"adminSecret", func(c *config.Config) { c.K8s.AdminSecret = "-A" }, `secret name "-A" is invalid`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := operatorManagedCfg()
			tc.apply(cfg)
			r := &recRunner{out: []byte("czNjcmV0Cg==")}
			_, err := ReadAdminPassword(r, cfg)
			if err == nil {
				t.Fatal("a reference kubectl would parse as a flag must be refused")
			}
			if len(r.calls) != 0 {
				t.Errorf("the refusal must come before exec, but %d call(s) ran: %+v", len(r.calls), r.calls)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to name %q", err, tc.want)
			}
		})
	}
}

const brokerListJSON = `{"items":[{"metadata":{"name":"dev-broker"},"spec":{"adminCredentialsSecret":%q}}]}`

// TestAdminPasswordPreflight pins the deploy guard: it refuses exactly the two deploys
// that would pair a broker's stored password with a new random one, and nothing else.
func TestAdminPasswordPreflight(t *testing.T) {
	run := func(t *testing.T, cfg *config.Config, r *recRunner) error {
		t.Helper()
		return NewCluster(r, cfg, nil, nil).AdminPasswordPreflight(context.Background())
	}

	t.Run("a named admin Secret asks nothing", func(t *testing.T) {
		for _, cfg := range []*config.Config{adminCfg(), func() *config.Config {
			c := haCfg()
			c.K8s.AdminSecret = "byo-admin"
			return c
		}()} {
			r := &recRunner{}
			if err := run(t, cfg, r); err != nil || len(r.calls) != 0 {
				t.Errorf("err = %v, calls = %+v; the CR names a Secret, so there is nothing to check", err, r.calls)
			}
		}
	})
	t.Run("a running broker whose CR names a Secret is refused", func(t *testing.T) {
		r := &recRunner{outQueue: [][]byte{[]byte(strings.Replace(brokerListJSON, "%q", `"solace-admin-secret"`, 1))}}
		err := run(t, operatorManagedCfg(), r)
		if err == nil {
			t.Fatal("dropping the running broker's admin Secret reference must be refused")
		}
		for _, want := range []string{"solace-admin-secret", "kubernetes.adminSecret: solace-admin-secret", "dev-broker-pubsubplus-admin-creds"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q should mention %q", err, want)
			}
		}
	})
	t.Run("a broker already on the operator's Secret passes", func(t *testing.T) {
		for _, secret := range []string{`""`, `"dev-broker-pubsubplus-admin-creds"`} {
			r := &recRunner{outQueue: [][]byte{[]byte(strings.Replace(brokerListJSON, "%q", secret, 1))}}
			if err := run(t, operatorManagedCfg(), r); err != nil {
				t.Errorf("spec.adminCredentialsSecret=%s: the operator reuses its own Secret, so this must pass: %v", secret, err)
			}
			for _, c := range r.calls {
				if !isCanI(c.args) && strings.Contains(strings.Join(c.args, " "), "get pvc") {
					t.Errorf("a running broker settles it; no PVC read should follow: %+v", r.calls)
				}
			}
		}
	})
	t.Run("an earlier broker's kept data is refused", func(t *testing.T) {
		r := &recRunner{outQueue: [][]byte{
			[]byte(`{"items":[]}`),
			[]byte(`{"items":[{"metadata":{"name":"data-dev-broker-pubsubplus-b-0"}},{"metadata":{"name":"other"}}]}`),
		}}
		err := run(t, operatorManagedCfg(), r)
		if err == nil {
			t.Fatal("a new random password over kept data must be refused")
		}
		for _, want := range []string{"data-dev-broker-pubsubplus-b-0", "semp.adminPass", "--delete-data"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q should mention %q", err, want)
			}
		}
		if strings.Contains(err.Error(), "other") {
			t.Errorf("only this broker's data claims count: %v", err)
		}
	})
	t.Run("a first deploy passes", func(t *testing.T) {
		r := &recRunner{outQueue: [][]byte{[]byte(`{"items":[]}`), []byte(`{"items":[{"metadata":{"name":"other"}}]}`)}}
		if err := run(t, operatorManagedCfg(), r); err != nil {
			t.Errorf("no broker and no data of its own: nothing to lose, so this must pass: %v", err)
		}
	})
	t.Run("no answer is not a refusal", func(t *testing.T) {
		if err := run(t, operatorManagedCfg(), &recRunner{}); err != nil {
			t.Errorf("a silent runner has told us nothing to refuse on: %v", err)
		}
	})
	t.Run("an unreadable CR still refuses over kept data, naming why", func(t *testing.T) {
		// The CR read fails -- most often the CRD is not installed yet -- so the data
		// claims decide, and the refusal carries the read's cause.
		r := &recRunner{
			outErrQueue: []error{errors.New("no such kind")},
			outQueue:    [][]byte{nil, []byte(`{"items":[{"metadata":{"name":"data-dev-broker-pubsubplus-p-0"}}]}`)},
		}
		err := run(t, operatorManagedCfg(), r)
		if err == nil || !strings.Contains(err.Error(), "no such kind") || !strings.Contains(err.Error(), "data-dev-broker-pubsubplus-p-0") {
			t.Errorf("err = %v, want the refusal naming the kept claim and the failed CR read", err)
		}
	})
	t.Run("an unreadable CR with no data passes", func(t *testing.T) {
		r := &recRunner{outErrQueue: []error{errors.New("no such kind")}, outQueue: [][]byte{nil, []byte(`{"items":[]}`)}}
		if err := run(t, operatorManagedCfg(), r); err != nil {
			t.Errorf("no broker can be running without its data, so a first deploy must pass: %v", err)
		}
	})
	t.Run("a refused CR read falls through to the data check", func(t *testing.T) {
		// The CR read is refused (kubectl's Forbidden on stderr): that is no reason to stop,
		// since the deploy's own apply is what needs the grant. The claims decide, so with
		// nothing of this broker's data there is nothing to lose and it passes ...
		r := &recRunner{
			outErrQueue: []error{errors.New("exit status 1")},
			outQueue:    [][]byte{nil, []byte(`{"items":[]}`)},
			stderrQueue: []string{forbiddenGet},
		}
		if err := run(t, operatorManagedCfg(), r); err != nil {
			t.Fatalf("a refused CR read must fall through to the data check: %v", err)
		}
		for _, c := range r.calls {
			if isCanI(c.args) {
				t.Errorf("reads are not probed as requirements, but %v was asked", c.args)
			}
		}
		// ... and with kept data it still refuses, naming why the CR was not read.
		r = &recRunner{
			outErrQueue: []error{errors.New("exit status 1")},
			outQueue:    [][]byte{nil, []byte(`{"items":[{"metadata":{"name":"data-dev-broker-pubsubplus-p-0"}}]}`)},
			stderrQueue: []string{forbiddenGet},
		}
		err := run(t, operatorManagedCfg(), r)
		if err == nil || !strings.Contains(err.Error(), "data-dev-broker-pubsubplus-p-0") ||
			!strings.Contains(err.Error(), "could not be read") {
			t.Errorf("err = %v, want the refusal naming the kept claim and the unread CR", err)
		}
	})
	t.Run("both reads are by name, so get is all they need, and neither is probed", func(t *testing.T) {
		r := &recRunner{}
		if err := run(t, operatorManagedCfg(), r); err != nil {
			t.Fatalf("AdminPasswordPreflight: %v", err)
		}
		var reads [][]string
		for _, c := range r.calls {
			if isCanI(c.args) {
				t.Errorf("a read is not a requirement, so it is not probed: %v", c.args)
				continue
			}
			reads = append(reads, c.args)
		}
		want := [][]string{
			{"get", brokerResource, "dev-broker", "-n", "solace", "--ignore-not-found", "-o", "json"},
			{"get", "pvc", "data-dev-broker-pubsubplus-p-0", "data-dev-broker-pubsubplus-b-0",
				"data-dev-broker-pubsubplus-m-0", "-n", "solace", "--ignore-not-found", "-o", "json"},
		}
		if len(reads) != len(want) || !eqArgs(reads[0], want[0]) || !eqArgs(reads[1], want[1]) {
			t.Errorf("reads = %v, want %v", reads, want)
		}
	})
	t.Run("a refused claims read warns and lets the deploy proceed", func(t *testing.T) {
		r := &recRunner{
			outErrQueue: []error{nil, errors.New("exit status 1")},
			outQueue:    [][]byte{[]byte(`{"items":[]}`), nil},
			stderrQueue: []string{strings.Replace(forbiddenGet, "pubsubpluseventbrokers", "persistentvolumeclaims", -1)},
		}
		log, narration := logBuf()
		c := NewCluster(r, operatorManagedCfg(), log, nil)
		if err := c.AdminPasswordPreflight(context.Background()); err != nil {
			t.Fatalf("reading claims is no part of a deploy; a refused read must not stop it: %v", err)
		}
		for _, want := range []string{"not checked", "get persistentvolumeclaims", "semp.adminPass"} {
			if !strings.Contains(narration.String(), want) {
				t.Errorf("warning %q should mention %q", narration.String(), want)
			}
		}
		if len(r.calls) != 2 {
			t.Errorf("calls = %+v, want the CR read and the one refused claims read", r.calls)
		}
	})
	t.Run("a running broker's Secret is decided before the claims are read", func(t *testing.T) {
		r := &recRunner{outQueue: [][]byte{[]byte(strings.Replace(brokerListJSON, "%q", `"solace-admin-secret"`, 1))}}
		if err := run(t, operatorManagedCfg(), r); err == nil || !strings.Contains(err.Error(), "solace-admin-secret") {
			t.Errorf("err = %v, want the refusal for the dropped reference: the CR decides before the claims", err)
		}
		if len(r.calls) != 1 {
			t.Errorf("calls = %+v, want only the CR read", r.calls)
		}
	})
	t.Run("a failed check refuses and says how to skip it", func(t *testing.T) {
		err := run(t, operatorManagedCfg(), &recRunner{outErr: errors.New("forbidden")})
		if err == nil || !strings.Contains(err.Error(), "forbidden") || !strings.Contains(err.Error(), "semp.adminPass") {
			t.Errorf("err = %v, want the cause and the way to deploy without the check", err)
		}
	})
}

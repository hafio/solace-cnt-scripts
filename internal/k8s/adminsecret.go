package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"solace/internal/config"
	"solace/internal/engine"
)

// adminsecret.go is where this deployment's admin password lives on Kubernetes when the
// env file does not carry it, and the one deploy that would lose it.
//
// config.Config.AdminSecretName decides what the CR names: a Secret this tool builds from
// semp.adminPass, one it only references, or none -- in which case the Solace operator
// generates <kubernetes.name>-pubsubplus-admin-creds with a random password. The last two
// leave this tool without the password, so the commands that authenticate read it back
// from the cluster (ReadAdminPassword), and `broker deploy` refuses the one transition
// that would leave nobody holding it (AdminPasswordPreflight).
//
// That transition exists because the broker reads its admin password from the Secret
// only when it boots on a FRESH data volume (Solace's configuration keys are evaluated
// "upon initial startup ... or after a reload to the default configuration", and the
// operator guide says changing the Secret later "will not result in password updates in
// the broker"). The operator's generated Secret is owned by the CR and deleted with it,
// while `broker remove` keeps the data by default -- so a redeploy onto that data would
// get a new random password the broker ignores.

// AdminSecretInUse is the Secret the broker's admin password is in: the one the CR names
// (config.Config.AdminSecretName), or, when it names none, the one the operator generates
// (config.Config.OperatorAdminSecretName).
func AdminSecretInUse(cfg *config.Config) string {
	if name := cfg.AdminSecretName(); name != "" {
		return name
	}
	return cfg.OperatorAdminSecretName()
}

// ReadAdminPassword reads the broker's admin password back out of AdminSecretInUse, for a
// command that must authenticate when the env file carries no semp.adminPass. It goes
// through readSecretKey, the reader replication already uses, and adds the next step that
// fits whichever Secret it was.
func ReadAdminPassword(r engine.Runner, cfg *config.Config) (string, error) {
	cmd, err := cfg.ClusterCommand()
	if err != nil {
		return "", err
	}
	name := AdminSecretInUse(cfg)
	pass, err := readSecretKey(r, cmd, cfg.K8s.Namespace, name, adminPassKey, fmt.Sprintf("broker %q", cfg.K8s.Name))
	if err == nil {
		return pass, nil
	}
	next := fmt.Sprintf("kubernetes.adminSecret names %s, which this tool references but never creates: it "+
		"must exist in namespace %s with the key %s -- or set semp.adminPass", name, cfg.K8s.Namespace, adminPassKey)
	if cfg.AdminSecretName() == "" {
		next = fmt.Sprintf("semp.adminPass and kubernetes.adminSecret are both unset, so the Solace operator "+
			"generates %s when it first reconciles the broker -- deploy the broker first, or set semp.adminPass",
			name)
	}
	return "", fmt.Errorf("%w\n  %s. Reading it needs `get secrets` in namespace %s", err, next, cfg.K8s.Namespace)
}

// LiveAdminSecret reads the admin Secret the running broker's CR names. found is false
// when no broker of this name exists in the namespace, or when the runner answered
// nothing at all (the preview and test seams). name is "" for a broker on the operator's
// generated Secret.
func (c *Cluster) LiveAdminSecret(ctx context.Context) (name string, found bool, err error) {
	raw, err := c.kubectlOutput(ctx, "get", brokerResource, "-n", c.ns(), "-o", "json")
	if err != nil {
		return "", false, err
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return "", false, nil
	}
	var brokers struct {
		Items []struct {
			Metadata objectMeta `json:"metadata"`
			Spec     struct {
				AdminCredentialsSecret string `json:"adminCredentialsSecret"`
			} `json:"spec"`
		} `json:"items"`
	}
	if err := json.Unmarshal(normalizeToList(raw), &brokers); err != nil {
		return "", false, fmt.Errorf("decode the brokers in namespace %q: %w", c.ns(), err)
	}
	for _, b := range brokers.Items {
		if b.Metadata.Name == c.Cfg.K8s.Name {
			return b.Spec.AdminCredentialsSecret, true, nil
		}
	}
	return "", false, nil
}

// AdminPasswordPreflight refuses the deploy that is certain to leave the broker with an
// admin password nobody holds. It applies only when the CR will name no admin Secret, so
// the operator generates one:
//
//   - the broker already runs with its password in a Secret the CR names. Dropping the
//     reference makes the operator generate a new random password the broker ignores.
//   - no broker exists, but an earlier one's data PVCs do. The operator would generate a
//     new random password for data that keeps the old one.
//
// Everything else passes: a first deploy, and a broker already on its generated Secret,
// which the operator reuses. A runner that answers nothing has told us nothing (the
// preview and test seams), so there is nothing to refuse on.
//
// The two reads are permission-probed first, so an expired login or a missing grant gets
// the preflight's own guidance rather than a message about passwords. A failed CR read
// after that is most often the CRD not being installed yet, where no broker can exist:
// it is not fatal on its own, and the data check decides -- a running broker always has
// its data claims, so a refusal still follows whenever there is something to lose.
// Claims under a custom volume mount are not checked: they exist before a first deploy,
// so their presence says nothing about earlier data (docs/configuration.md, "The admin
// Secret", says so).
func (c *Cluster) AdminPasswordPreflight(ctx context.Context) error {
	if c.Cfg.AdminSecretName() != "" {
		return nil
	}
	if err := c.PreflightAll(ctx,
		probe{verb: "list", resource: brokerResource},
		probe{verb: "list", resource: "persistentvolumeclaims"},
	); err != nil {
		return err
	}
	name, generated := c.Cfg.K8s.Name, c.Cfg.OperatorAdminSecretName()
	live, found, crErr := c.LiveAdminSecret(ctx)
	if crErr == nil && found {
		if live != "" && live != generated {
			return fmt.Errorf("broker %q runs with its admin password in the Secret %q, and this env file names no "+
				"admin Secret and no semp.adminPass: the CR would drop that reference and the operator would generate "+
				"%s with a new random password, which the broker ignores -- it read its password once, on a fresh "+
				"data volume -- so nothing would hold its working password.\n"+
				"  Keep the reference: set kubernetes.adminSecret: %s (or semp.adminPass to that password)",
				name, live, generated, live)
		}
		return nil
	}
	var pvcs struct {
		Items []struct {
			Metadata objectMeta `json:"metadata"`
		} `json:"items"`
	}
	if err := c.getJSON(ctx, &pvcs, "pvc", "-n", c.ns()); err != nil {
		return fmt.Errorf("check namespace %q for an earlier broker's data before letting the operator generate "+
			"its admin password: %w\n  Set semp.adminPass (or kubernetes.adminSecret) to deploy without this check",
			c.ns(), err)
	}
	want := map[string]bool{}
	for _, role := range HARoles(c.Cfg) {
		want[pvcName(c.Cfg, role)] = true
	}
	var kept []string
	for _, p := range pvcs.Items {
		if want[p.Metadata.Name] {
			kept = append(kept, p.Metadata.Name)
		}
	}
	if len(kept) == 0 {
		return nil
	}
	unread := ""
	if crErr != nil {
		unread = fmt.Sprintf(" (the broker CR could not be read to tell whether it is still running: %v)", crErr)
	}
	return fmt.Errorf("broker %q has data in namespace %q (%s)%s, and this env file names no admin Secret and no "+
		"semp.adminPass: the operator would generate %s with a new random password, but the data keeps the password "+
		"it was first deployed with -- the broker reads it only on a fresh data volume -- so nothing would hold its "+
		"working password, and in HA the standby and monitor pods would never become Ready.\n"+
		"  Set semp.adminPass to that broker's admin password (or kubernetes.adminSecret to a Secret holding it), "+
		"or, if the data is not needed, remove it first with `broker remove --delete-data`",
		name, c.ns(), strings.Join(kept, ", "), unread, generated)
}

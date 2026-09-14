package config

import (
	"fmt"
	"os"
	"regexp"
)

// envVarRE is the portable environment-variable name shape. The value of an *Env
// key is a NAME this process looks up, never a value and never anything that
// reaches a shell, so this check is about actionable errors rather than injection:
// it catches `passEnv: ${FOO}` and `passEnv: my secret` -- the two ways an
// operator mistakes the reference key for the value key -- before the lookup
// silently fails (§4a).
var envVarRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// secretRef pairs one secret field with the sibling *Env key that may supply it
// from the process environment instead of the env file.
type secretRef struct {
	field  string  // "semp.adminPass" -- the literal key, for actionable errors
	envKey string  // "semp.adminPassEnv" -- the reference key
	envVar string  // the variable that key names ("" when it is unset)
	value  *string // the secret field to fill
}

// secretRefs lists every secret field of the schema beside its *Env sibling. One
// list, so a secret added to the schema cannot silently miss out on environment
// referencing, and so every error message about it reads the same.
func (c *Config) secretRefs() []secretRef {
	refs := []secretRef{
		{"semp.adminPass", "semp.adminPassEnv", c.SEMP.AdminPassEnv, &c.SEMP.AdminPass},
		{"semp.monitorPass", "semp.monitorPassEnv", c.SEMP.MonitorPassEnv, &c.SEMP.MonitorPass},
		{"image.pass", "image.passEnv", c.Image.PassEnv, &c.Image.Pass},
		{"tls.certPassphrase", "tls.certPassphraseEnv", c.TLS.CertPassphraseEnv, &c.TLS.CertPassphrase},
		{"redundancy.psk", "redundancy.pskEnv", c.Redundancy.PSKEnv, &c.Redundancy.PSK},
	}
	for i := range c.SEMP.AdditionalUsers {
		u := &c.SEMP.AdditionalUsers[i]
		base := fmt.Sprintf("semp.additionalUsers[%d]", i)
		refs = append(refs, secretRef{base + ".password", base + ".passwordEnv", u.PasswordEnv, &u.Password})
	}
	// A replication site reached over SEMP carries the MATE's admin password, which is
	// a different broker's credential and often a different value from this file's
	// semp.adminPass. It belongs on this list for the ordinary reason: a secret the
	// schema accepts as a literal must also be referable from the environment, or the
	// only way to configure a DR pair is to write a password into a file that travels.
	for i := range c.Replication.Sites {
		s := c.Replication.Sites[i].Via.SEMP
		if s == nil {
			continue
		}
		base := fmt.Sprintf("replication.sites[%d].via.semp", i)
		refs = append(refs, secretRef{base + ".pass", base + ".passEnv", s.PassEnv, &s.Pass})
	}
	return refs
}

// resolveSecretRefs fills each secret field whose *Env sibling names an
// environment variable. Load runs it before defaults and validation, so
// everything downstream -- the renderers, both secret stores, the k8s Secret
// builders -- sees a plain resolved value and needs no knowledge of where it came
// from.
//
// Both keys set is a hard error rather than a precedence rule: guessing which one
// the operator meant is exactly the kind of silently-wrong password that only
// surfaces much later as a broker login failure. An unset or empty variable is the
// same kind of error: deploying a broker with a blank password is worse than
// refusing to deploy (§4a fail-loud). No message ever carries a value.
func (c *Config) resolveSecretRefs() error {
	for _, r := range c.secretRefs() {
		if r.envVar == "" {
			continue
		}
		if !envVarRE.MatchString(r.envVar) {
			return fmt.Errorf("%s %q is not a valid environment variable name: expected a letter or '_' "+
				"followed by letters, digits or '_' -- %s takes the variable NAME, %s takes the value itself",
				r.envKey, r.envVar, r.envKey, r.field)
		}
		if *r.value != "" {
			return fmt.Errorf("%s and %s are both set: use one or the other -- %s carries the value, "+
				"%s names the environment variable holding it", r.field, r.envKey, r.field, r.envKey)
		}
		v, set := os.LookupEnv(r.envVar)
		if v == "" {
			return fmt.Errorf("%s names environment variable %s, which is %s; export it before running",
				r.envKey, r.envVar, unsetOrEmpty(set))
		}
		*r.value = v
	}
	return nil
}

// unsetOrEmpty distinguishes the two failures, since the fix differs: a missing
// export against an export that produced nothing.
func unsetOrEmpty(set bool) string {
	if set {
		return "set but empty"
	}
	return "not set"
}

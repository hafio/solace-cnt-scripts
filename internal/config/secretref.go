package config

import (
	"fmt"
	"os"

	"github.com/hafio/go-deploy-utility/envref"
)

// secretRef pairs one secret field, as an envref.Ref carrying its literal key, its
// sibling *Env key and the two current values, with the pointer the resolved secret
// is written back to.
type secretRef struct {
	ref   envref.Ref
	value *string // the secret field to fill
}

// secretRefs lists every secret field of the schema beside its *Env sibling. One
// list, so a secret added to the schema cannot silently miss out on environment
// referencing, and so every error message about it reads the same.
func (c *Config) secretRefs() []secretRef {
	refs := []secretRef{
		newSecretRef("semp.adminPass", "semp.adminPassEnv", &c.SEMP.AdminPass, c.SEMP.AdminPassEnv),
		newSecretRef("semp.monitorPass", "semp.monitorPassEnv", &c.SEMP.MonitorPass, c.SEMP.MonitorPassEnv),
		newSecretRef("image.pass", "image.passEnv", &c.Image.Pass, c.Image.PassEnv),
		newSecretRef("tls.certPassphrase", "tls.certPassphraseEnv", &c.TLS.CertPassphrase, c.TLS.CertPassphraseEnv),
		newSecretRef("redundancy.psk", "redundancy.pskEnv", &c.Redundancy.PSK, c.Redundancy.PSKEnv),
	}
	for i := range c.SEMP.AdditionalUsers {
		u := &c.SEMP.AdditionalUsers[i]
		base := fmt.Sprintf("semp.additionalUsers[%d]", i)
		refs = append(refs, newSecretRef(base+".password", base+".passwordEnv", &u.Password, u.PasswordEnv))
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
		refs = append(refs, newSecretRef(base+".pass", base+".passEnv", &s.Pass, s.PassEnv))
	}
	return refs
}

// newSecretRef builds one entry from the two key names, the literal field and the
// *Env key's value.
func newSecretRef(field, envKey string, value *string, name string) secretRef {
	return secretRef{envref.Ref{Field: field, EnvKey: envKey, Literal: *value, Name: name}, value}
}

// resolveSecretRefs fills each secret field whose *Env sibling names an
// environment variable. Load runs it before defaults and validation, so
// everything downstream -- the renderers, both secret stores, the k8s Secret
// builders -- sees a plain resolved value and needs no knowledge of where it came
// from.
//
// The rules are the shared envref package's: both keys set is a hard error rather
// than a precedence rule, and an unset or empty variable is the same kind of error,
// since deploying a broker with a blank password is worse than refusing to deploy.
// No message carries a value, and a name that fails the name rules is not printed
// either, in case it is the secret itself. A field with no *Env key is left as
// written.
func (c *Config) resolveSecretRefs() error {
	for _, r := range c.secretRefs() {
		if r.ref.Name == "" {
			continue
		}
		v, err := r.ref.Resolve(os.LookupEnv)
		if err != nil {
			return err
		}
		*r.value = v
	}
	return nil
}

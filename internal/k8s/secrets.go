package k8s

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"solace/internal/config"
)

// secretManifest is a core/v1 Secret rendered to YAML with base64-encoded data,
// matching the shape of `kubectl create secret ... -o yaml`. Building the manifest
// in Go and applying it on stdin (`apply -f -`) is behavior-equivalent to the bash
// `create secret --from-literal=...` form (012) but keeps every secret value off
// the argv and out of an echoed command, which the bash form leaked (012:26,36,39,
// 43). §3 hardening.
type secretManifest struct {
	name      string
	namespace string
	typ       string
	data      map[string][]byte
}

// render emits the manifest. Data keys are sorted so the output is deterministic
// and golden-testable; kubectl is order-insensitive.
func (s secretManifest) render() []byte {
	var b strings.Builder
	b.WriteString("apiVersion: v1\n")
	b.WriteString("kind: Secret\n")
	b.WriteString("metadata:\n")
	b.WriteString("  name: " + s.name + "\n")
	b.WriteString("  namespace: " + s.namespace + "\n")
	b.WriteString("type: " + s.typ + "\n")
	b.WriteString("data:\n")
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString("  " + k + ": " + base64.StdEncoding.EncodeToString(s.data[k]) + "\n")
	}
	return []byte(b.String())
}

// pskSecretKey is the entry the operator reads a pre-shared key from. Fixed by the CRD
// (spec.preSharedAuthKeySecret's description names it), so it is spelled once here rather
// than derived from anything of ours.
const pskSecretKey = "preshared_auth_key"

// AdminSecret builds the Opaque secret holding broker credentials, porting the
// user-secret of 012:26-32: username_admin_password (mandatory) and an optional
// username_monitor_password. Fails loud on an empty admin password.
//
// admin.additionalUsers are deliberately NOT here, and the reason is the operator's:
// it reads only the two keys above out of this Secret, so extra username_<user>_password
// keys are ignored and including them wrote passwords into a Secret nothing ever read
// (verified against a live cluster). They get a Secret of their own instead --
// AdditionalUsersSecret, named in the CR's spec.extraEnvVarsSecret.
//
// The split is not tidiness. extraEnvVarsSecret is projected with envFrom, which turns
// EVERY key of the Secret it names into an environment variable: point it at this one and
// the admin and monitor passwords land in the pod environment too, to get the extra users
// in. A separate Secret puts exactly the additional users' material there and nothing else.
func AdminSecret(cfg *config.Config) ([]byte, error) {
	if cfg.SEMP.AdminPass == "" {
		return nil, fmt.Errorf("admin.pass must be set to build the admin secret")
	}
	if cfg.K8s.AdminSecret == "" {
		return nil, fmt.Errorf("kubernetes.adminSecret (the secret name) must be set")
	}
	data := map[string][]byte{
		"username_admin_password": []byte(cfg.SEMP.AdminPass),
	}
	if cfg.SEMP.MonitorPass != "" {
		data["username_monitor_password"] = []byte(cfg.SEMP.MonitorPass)
	}
	// The pre-shared key rides in the SAME Secret rather than one of its own: the CR
	// references it through a separate field (preSharedAuthKeySecret), so one Secret
	// object can serve both, and a second object would be another thing to create,
	// name, and clean up for one extra key.
	//
	// `preshared_auth_key` is the CRD's spelling, not ours -- the operator looks for
	// exactly that entry, so it is written as a constant rather than derived.
	if psk := cfg.Redundancy.PSK; psk != "" {
		data[pskSecretKey] = []byte(psk)
	}
	return secretManifest{
		name:      cfg.K8s.AdminSecret,
		namespace: cfg.K8s.Namespace,
		typ:       "Opaque",
		data:      data,
	}.render(), nil
}

// AdditionalUsersSecret builds the Opaque Secret behind admin.additionalUsers, which the
// broker CR names in spec.extraEnvVarsSecret. Returns nil when none are configured: the CR
// then omits the field, and a Secret with no data is not worth applying.
//
// Both halves of each user ride the environment here, which is the one place this platform
// departs from the container convention of mounting every secret as a file. The CRD has no
// way to mount an arbitrary Secret -- its spec offers extraEnvVars, extraEnvVarsCM and
// extraEnvVarsSecret, and no volume passthrough -- so username_<u>_passwordfilepath would
// name a path nothing creates. The env-var form is the only one the operator can deliver,
// and it is why config.validateAdditionalUsers holds these usernames to a stricter rule on
// Kubernetes than elsewhere: envFrom silently DROPS keys that are not valid variable names.
func AdditionalUsersSecret(cfg *config.Config) ([]byte, error) {
	if len(cfg.SEMP.AdditionalUsers) == 0 {
		return nil, nil
	}
	data := make(map[string][]byte, len(cfg.SEMP.AdditionalUsers)*2)
	for _, u := range cfg.SEMP.AdditionalUsers {
		if u.Password == "" {
			return nil, fmt.Errorf("semp.additionalUsers %q has no password to build a secret from", u.Username)
		}
		// The access level is not a secret, but it rides the same Secret because it has
		// to reach the broker as an environment variable too and extraEnvVarsSecret is
		// the only channel the CRD gives us. The container platforms put it in the
		// artifact instead, where it is not sensitive.
		data["username_"+u.Username+"_globalaccesslevel"] = []byte(u.AccessLevel)
		data["username_"+u.Username+"_password"] = []byte(u.Password)
	}
	return secretManifest{
		name:      cfg.AdditionalUsersSecretName(),
		namespace: cfg.K8s.Namespace,
		typ:       "Opaque",
		data:      data,
	}.render(), nil
}

// TLSSecret builds the kubernetes.io/tls secret from the configured server
// certificate, porting 012:39 / 051:32: tls.crt is the certificate followed by any
// trusted CAs (the bash `--cert <(cat cert cas)`), tls.key is the private key. Both
// files are read from disk here; the manifest is applied on stdin so the key never
// reaches an argv or an echoed command (§3).
func TLSSecret(cfg *config.Config) ([]byte, error) {
	if cfg.TLS.Cert == "" || cfg.TLS.CertKey == "" {
		return nil, fmt.Errorf("tls.cert and tls.certKey must both be set to build the TLS secret")
	}
	if cfg.K8s.TLSServerSecret == "" {
		return nil, fmt.Errorf("kubernetes.tlsServerSecret (the secret name) must be set")
	}
	crt, err := os.ReadFile(cfg.TLS.Cert)
	if err != nil {
		return nil, fmt.Errorf("read tls.cert %q: %w", cfg.TLS.Cert, err)
	}
	for _, ca := range cfg.TLS.CAs {
		caBytes, err := os.ReadFile(ca)
		if err != nil {
			return nil, fmt.Errorf("read tls CA %q: %w", ca, err)
		}
		crt = append(crt, caBytes...)
	}
	key, err := os.ReadFile(cfg.TLS.CertKey)
	if err != nil {
		return nil, fmt.Errorf("read tls.certKey %q: %w", cfg.TLS.CertKey, err)
	}
	return secretManifest{
		name:      cfg.K8s.TLSServerSecret,
		namespace: cfg.K8s.Namespace,
		typ:       "kubernetes.io/tls",
		data: map[string][]byte{
			"tls.crt": crt,
			"tls.key": key,
		},
	}.render(), nil
}

// dockerAuthEntry is one registry credential in a .dockerconfigjson payload. The
// field order (username, password, auth) is fixed by struct order so the rendered
// secret is deterministic and golden-testable.
type dockerAuthEntry struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Auth     string `json:"auth"`
}

// dockerConfigJSON is the .dockerconfigjson document of an image-pull secret.
type dockerConfigJSON struct {
	Auths map[string]dockerAuthEntry `json:"auths"`
}

// dockerRegistrySecret builds a kubernetes.io/dockerconfigjson pull secret named
// name in namespace ns for cfg.Image's registry/user/pass, porting 012:43. It backs
// both the broker image-pull secret (DockerRegistrySecret) and the operator's fixed
// "regcred" (operatorRegcred, 010:29), which live in different namespaces.
func dockerRegistrySecret(name, ns string, cfg *config.Config) ([]byte, error) {
	if name == "" {
		return nil, fmt.Errorf("image-pull secret name must not be empty")
	}
	auth := base64.StdEncoding.EncodeToString([]byte(cfg.Image.User + ":" + cfg.Image.Pass))
	payload, err := json.Marshal(dockerConfigJSON{
		Auths: map[string]dockerAuthEntry{
			cfg.Image.Registry: {
				Username: cfg.Image.User,
				Password: cfg.Image.Pass,
				Auth:     auth,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal dockerconfigjson: %w", err)
	}
	return secretManifest{
		name:      name,
		namespace: ns,
		typ:       "kubernetes.io/dockerconfigjson",
		data:      map[string][]byte{".dockerconfigjson": payload},
	}.render(), nil
}

// DockerRegistrySecret builds the broker image-pull secret -- named
// kubernetes.imagePullSecret when set, else the derived default
// (config.Config.ImagePullSecretName) -- in the broker namespace. Called only when
// ManagesImagePullSecret is true, so the derived name is always a real name here, never "".
func DockerRegistrySecret(cfg *config.Config) ([]byte, error) {
	return dockerRegistrySecret(cfg.ImagePullSecretName(), cfg.K8s.Namespace, cfg)
}

// operatorRegcredName is the fixed name of the operator's image-pull Secret. Named
// rather than inlined because OperatorDelete has to remove it BY NAME: it is applied
// separately from the bundle and used to be reaped along with the operator namespace,
// which is no longer deleted. Fixed rather than derived the way the broker's own pull
// secret is (config.Config.ImagePullSecretName) is the operator's own deliberate choice:
// the operator install is one thing shared by every env file in the cluster, so there is
// no per-file name to derive it from, and "regcred" stays the name whether or not this env
// file names its own kubernetes.imagePullSecret.
const operatorRegcredName = "regcred"

// operatorRegcred builds the operator's image-pull secret under the fixed name
// "regcred" in the operator namespace opNS (010:29).
func operatorRegcred(cfg *config.Config, opNS string) ([]byte, error) {
	return dockerRegistrySecret(operatorRegcredName, opNS, cfg)
}

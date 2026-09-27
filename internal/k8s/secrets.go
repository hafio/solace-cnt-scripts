package k8s

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"solace/internal/config"
	"solace/internal/render"
)

// secretManifest is a core/v1 Secret rendered to YAML with base64-encoded data,
// matching the shape of `kubectl create secret ... -o yaml`. Building the manifest
// in Go and applying it on stdin (`apply -f -`) is behavior-equivalent to the bash
// `create secret --from-literal=...` form (012) but keeps every secret value off
// the argv and out of an echoed command, which the bash form leaked (012:26,36,39,
// 43).
type secretManifest struct {
	name        string
	namespace   string
	typ         string
	annotations map[string]string
	data        map[string][]byte
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
	if len(s.annotations) > 0 {
		// Values quoted: an annotation value is a string, and a hex digest that
		// happens to be all digits would otherwise be read as a number.
		b.WriteString("  annotations:\n")
		names := make([]string, 0, len(s.annotations))
		for k := range s.annotations {
			names = append(names, k)
		}
		sort.Strings(names)
		for _, k := range names {
			b.WriteString("    " + k + ": " + strconv.Quote(s.annotations[k]) + "\n")
		}
	}
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

// adminPassKey is the credentials Secret's admin-password entry, the operator's spelling
// again (spec.adminCredentialsSecret's description names it). It is written here and read
// back by ReadAdminPassword, so it is spelled once for both.
const adminPassKey = "username_admin_password"

// AdminSecret builds the Opaque secret holding broker credentials, porting the
// user-secret of 012:26-32: username_admin_password and an optional
// username_monitor_password. It is built only under config.Config.ManagesAdminSecret --
// GenSecrets asks -- and fails loud on an empty admin password or no name to build under.
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
		return nil, fmt.Errorf("semp.adminPass must be set to build the admin secret")
	}
	name := cfg.AdminSecretName()
	if name == "" {
		// Unreachable from a loaded config: Load requires kubernetes.name, and the
		// password above then derives <kubernetes.name>-admin.
		return nil, fmt.Errorf("the admin secret has no name: set kubernetes.adminSecret, or " +
			"kubernetes.name so the default <kubernetes.name>-admin can be derived")
	}
	data := map[string][]byte{
		adminPassKey: []byte(cfg.SEMP.AdminPass),
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
		name:      name,
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
// reaches an argv or an echoed command.
func TLSSecret(cfg *config.Config) ([]byte, error) {
	if cfg.TLS.Cert == "" || cfg.TLS.CertKey == "" {
		return nil, fmt.Errorf("tls.cert and tls.certKey must both be set to build the TLS secret")
	}
	name := cfg.TLSServerSecretName()
	if name == "" {
		// Unreachable from a loaded config: Load requires kubernetes.name, and the pair above
		// then derives <kubernetes.name>-tls. Reaching it means a Config built in code with
		// neither a Secret name nor a kubernetes.name.
		return nil, fmt.Errorf("the TLS secret has no name: set kubernetes.tlsServerSecret, or " +
			"kubernetes.name so the default <kubernetes.name>-tls can be derived")
	}
	key, crt, err := readTLSMaterial(cfg)
	if err != nil {
		return nil, err
	}
	return secretManifest{
		name:        name,
		namespace:   cfg.K8s.Namespace,
		typ:         "kubernetes.io/tls",
		annotations: map[string]string{render.CertDigestLabel: tlsDigest(key, crt)},
		data: map[string][]byte{
			"tls.crt": crt,
			"tls.key": key,
		},
	}.render(), nil
}

// readTLSMaterial reads the TLS Secret's two entries off disk: the key, and the
// certificate with every tls.cas file appended, which is the chain tls.crt holds.
func readTLSMaterial(cfg *config.Config) (key, crt []byte, err error) {
	crt, err = os.ReadFile(cfg.TLS.Cert)
	if err != nil {
		return nil, nil, fmt.Errorf("read tls.cert %q: %w", cfg.TLS.Cert, err)
	}
	for _, ca := range cfg.TLS.CAs {
		caBytes, err := os.ReadFile(ca)
		if err != nil {
			return nil, nil, fmt.Errorf("read tls CA %q: %w", ca, err)
		}
		crt = append(crt, caBytes...)
	}
	key, err = os.ReadFile(cfg.TLS.CertKey)
	if err != nil {
		return nil, nil, fmt.Errorf("read tls.certKey %q: %w", cfg.TLS.CertKey, err)
	}
	return key, crt, nil
}

// tlsDigest is the render.CertDigestLabel value the TLS Secret carries: the sha256 of
// the key followed by the certificate chain as the Secret holds them, CAs included, so
// a renewed CA marks the change as surely as a renewed certificate. It is an
// annotation rather than a label because a label value stops at 63 characters and
// the hex digest is 64.
func tlsDigest(key, crt []byte) string {
	sum := sha256.Sum256(append(append([]byte{}, key...), crt...))
	return hex.EncodeToString(sum[:])
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
// separately from the bundle, and the operator namespace that would otherwise reap it is
// deliberately never deleted. Fixed rather than derived the way the broker's own pull
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

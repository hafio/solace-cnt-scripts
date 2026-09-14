// Package config defines the single unified YAML schema for every platform
// (kubernetes, docker, podman), plus loading, defaulting, and validation. It
// replaces the two bash bootstraps (000-env.sh and docker-podman/000-env.sh):
// shared identity/image/admin/tls/redundancy live at the top level;
// platform-specific knobs live under kubernetes/docker/podman, with the HA group under redundancy.
package config

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Platform selects which platform's defaults, validation, and renderers apply.
type Platform string

const (
	K8s    Platform = "kubernetes"
	Docker Platform = "docker"
	Podman Platform = "podman"
)

// IsContainer reports whether p is a host-container platform (docker or podman),
// which share the container config, host-prep, and env-pair generation.
func (p Platform) IsContainer() bool { return p == Docker || p == Podman }

// Config is the whole deserialized env file. Fields shared by all platforms sit
// at the top; platform sections hold the rest. Unknown-to-a-platform fields are
// simply ignored by that platform's renderer.
type Config struct {
	// Redundancy is the HA group: whether there is one, who is in it, and the
	// pre-shared key its members authenticate with. Unified across platforms -- the
	// k8s renderer emits enabled as a bool into the CR, the container renderer emits
	// the redundancy_*/configsync_* key=value set (or just redundancy_enable=no).
	Redundancy Redundancy `yaml:"redundancy"`

	// Timezone applies to every platform -- the k8s CR's timezone field and the
	// containers' TZ setting are the same knob. Unset on purpose: an omitted
	// value leaves the broker on the image default rather than pinning a region.
	Timezone string `yaml:"timezone"`

	Image Image `yaml:"image"`
	SEMP  SEMP  `yaml:"semp"`
	// LegacyAdmin catches the section's former name so a stale env file gets a message
	// naming the rename, not a bare unknown-field decode error. Rejected in Validate;
	// map[string]any rather than SEMP so its CONTENTS never have to decode -- the old
	// shape had keys this one does not.
	LegacyAdmin map[string]any `yaml:"admin"`
	TLS         TLS            `yaml:"tls"`
	Scaling     Scaling        `yaml:"scaling"`
	Replication Replication    `yaml:"replication"`
	Broker      Broker         `yaml:"broker"`

	K8s    K8sConfig    `yaml:"kubernetes"`
	Docker DockerConfig `yaml:"docker"`
	Podman PodmanConfig `yaml:"podman"`

	// extraAllowed holds the binaries the operator approved for this invocation
	// with --allow-command. Unexported on purpose: yaml.v3 cannot decode into an
	// unexported field, so no env file can widen its own allowlist -- that
	// authority stays with the person at the keyboard (see AllowCommands).
	extraAllowed map[string]bool

	// baseDir is the directory of the env file this Config was loaded from, and
	// every relative host path in the file resolves against it (rebaseHostPaths).
	// Unexported for the same reason extraAllowed is: yaml.v3 cannot decode into
	// it, so a file cannot declare its own base and thereby reach a path outside
	// the directory it was found in.
	//
	// Empty is the documented no-op. A Config built by hand or by
	// internal/convert never goes through Load, has no path to derive a base from,
	// and must keep behaving exactly as it did -- paths left as written.
	baseDir string
}

// AdditionalUsersSecretName is the Secret carrying admin.additionalUsers, which the broker
// CR names in spec.extraEnvVarsSecret: <kubernetes.name>-additional-users.
//
// Derived rather than configured, unlike adminSecret/tlsServerSecret/imagePullSecret. It
// holds nothing an operator would want to create themselves -- this tool is the only thing
// that writes it -- and it cannot collide: two brokers in one namespace would already
// collide on the CR name itself, and different namespaces cannot collide at all.
//
// It lives here rather than in internal/k8s because internal/render needs it too, to emit
// the CR field, and render cannot import k8s (k8s imports render). One definition, two
// callers, in the leaf they share.
func (c *Config) AdditionalUsersSecretName() string {
	return c.K8s.Name + "-additional-users"
}

// ManagesTLSSecret reports whether this tool builds the TLS Secret itself, which is true
// exactly when the env file supplies the material to build it from. Naming the Secret
// (kubernetes.tlsServerSecret) only says the broker should USE one: it may already exist,
// created by hand or by cert-manager, in which case the CR references it by name and
// nothing here reads, applies or deletes it.
func (c *Config) ManagesTLSSecret() bool {
	return c.TLS.Cert != "" || c.TLS.CertKey != ""
}

// RedundancyEnabled reports HA mode (redundancy.enabled: true). Container HA and k8s HA
// both key off this; HA-only steps (leader, redundancy verify) no-op otherwise.
func (c *Config) RedundancyEnabled() bool { return c.Redundancy.Enabled == "true" }

// Command is an external command line: argv[0] plus any leading arguments that
// precede every call's own. It exists so the platform CLI can be more than a
// binary name -- `oc`, `microk8s kubectl`, or a profile like
// `kubectl --kubeconfig /path/.kubeconfig-cluster` (bash/env/customer-sample:7).
//
// The bash bootstraps got this for free by expanding ${KUBE} and
// ${CONTAINER_RUNTIME} unquoted, which word-splits. Go's exec never splits and
// never involves a shell, so the split happens here instead.
type Command []string

// UnmarshalYAML accepts either a scalar, split on whitespace exactly as the
// bash bootstraps' unquoted expansion did, or an explicit sequence, which is the
// only way to express a token that itself contains a space (a binary path under
// "C:\Program Files\..."). Whitespace splitting deliberately does not honour
// embedded quotes -- neither did bash's word splitting.
func (c *Command) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var s string
		if err := value.Decode(&s); err != nil {
			return err
		}
		*c = strings.Fields(s)
	case yaml.SequenceNode:
		var parts []string
		if err := value.Decode(&parts); err != nil {
			return err
		}
		*c = parts
	default:
		return fmt.Errorf("line %d: a command must be a string (split on whitespace) "+
			"or a list of exact arguments", value.Line)
	}
	return nil
}

// Name is argv[0] -- the executable actually run.
func (c Command) Name() string {
	if len(c) == 0 {
		return ""
	}
	return c[0]
}

// Args prepends the configured leading arguments to a single call's own args. It
// always builds a new slice, so a caller's backing array is never aliased or
// appended into by a later call.
func (c Command) Args(extra ...string) []string {
	if len(c) <= 1 {
		return extra
	}
	out := make([]string, 0, len(c)-1+len(extra))
	out = append(out, c[1:]...)
	return append(out, extra...)
}

// String renders the command for reports and error messages.
func (c Command) String() string { return strings.Join(c, " ") }

// Image is the broker image reference and registry credentials, shared by every
// platform (k8s builds the regcred Secret from user/pass; docker and podman feed
// them to `<runtime> login`). The k8s-only pull knobs -- the pull-secret NAME and
// the pull policy -- live under kubernetes.* instead.
type Image struct {
	Repo     string `yaml:"repo"`     // SOLBK_IMAGE
	Tag      string `yaml:"tag"`      // SOLBK_IMG_TAG
	Registry string `yaml:"registry"` // IMAGEREPO_HOST (optional prefix)
	User     string `yaml:"user"`     // IMAGEREPO_USER
	Pass     string `yaml:"pass"`     // IMAGEREPO_PASS (secret)
	PassEnv  string `yaml:"passEnv"`  // env var holding pass instead
}

// Ref is the fully-qualified image reference, with the optional registry prefix.
func (i Image) Ref() string {
	if i.Registry != "" {
		return fmt.Sprintf("%s/%s:%s", i.Registry, i.Repo, i.Tag)
	}
	return fmt.Sprintf("%s:%s", i.Repo, i.Tag)
}

// TagVersion parses the leading major.minor of the image tag ("10.26.1.5" -> 10,
// 26). ok is false when the tag carries no version at all -- "latest", a digest,
// a bare codename -- and callers must treat that as *unknown*, never as old or
// new: a feature gate that guessed either way would be wrong half the time.
func (i Image) TagVersion() (major, minor int, ok bool) {
	parts := strings.Split(strings.TrimSpace(i.Tag), ".")
	if len(parts) < 2 {
		return 0, 0, false
	}
	major, ok = atoiPrefix(parts[0])
	if !ok {
		return 0, 0, false
	}
	minor, ok = atoiPrefix(parts[1])
	if !ok {
		return 0, 0, false
	}
	return major, minor, true
}

// AtLeast reports whether the image tag names a broker release at or above
// major.minor. known is false when the tag carries no version to compare.
func (i Image) AtLeast(major, minor int) (ok, known bool) {
	haveMajor, haveMinor, known := i.TagVersion()
	if !known {
		return false, false
	}
	if haveMajor != major {
		return haveMajor > major, true
	}
	return haveMinor >= minor, true
}

// atoiPrefix parses the digits at the start of s, so a tag component like "0-rc1"
// still yields 0. It fails when there are no leading digits at all.
func atoiPrefix(s string) (int, bool) {
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, false
	}
	n, err := strconv.Atoi(s[:end])
	if err != nil {
		return 0, false
	}
	return n, true
}

// SEMP holds the broker's management (SEMP/CLI) credentials. Passwords are secrets:
// never logged/echoed. Every secret field has a sibling *Env key naming an environment
// variable to read the value from instead, so an env file can be reviewed and shared
// without carrying a single secret (resolveSecretRefs).
//
// There is deliberately no username field. The built-in admin user is called `admin` on
// every platform: Kubernetes has no choice -- the operator reads the fixed
// username_admin_password key out of the credentials Secret -- and letting the container
// platforms name it something else bought a divergence nobody asked for, at the cost of a
// value that reached a Secret key, a compose `target:`, a podman `target=` and a broker
// setting. The monitor user is likewise fixed at `monitor`.
type SEMP struct {
	AdminPass       string           `yaml:"adminPass"`       // SOLBK_ADM_PASS (secret, mandatory)
	AdminPassEnv    string           `yaml:"adminPassEnv"`    // env var holding adminPass instead
	MonitorPass     string           `yaml:"monitorPass"`     // SOLBK_MON_PASS (k8s, secret)
	MonitorPassEnv  string           `yaml:"monitorPassEnv"`  // env var holding monitorPass instead
	AdditionalUsers []AdditionalUser `yaml:"additionalUsers"` // extra CLI users beyond admin/monitor
}

// AdminUser and MonitorUser are the broker's built-in management accounts. They are the
// broker's names, not this schema's, which is why they are constants rather than fields.
const (
	AdminUser   = "admin"
	MonitorUser = "monitor"
)

// AdditionalUser is one extra broker CLI user. It replaces the old
// admin.userPasswords "user=password" list: the access level was not expressible
// there, and a password is now referable through the environment like every other
// secret. Exactly one of Password/PasswordEnv must be set.
type AdditionalUser struct {
	Username    string `yaml:"username"`    // becomes username_<username>_password
	AccessLevel string `yaml:"accessLevel"` // admin|read-write|read-only|none
	Password    string `yaml:"password"`    // secret
	PasswordEnv string `yaml:"passwordEnv"` // env var holding password instead
}

// TLS is the broker server certificate + trusted CAs, shared by every platform.
// The name of the k8s Secret built from these files is kubernetes.tlsServerSecret.
type TLS struct {
	Cert    string   `yaml:"cert"`    // SOLBK_TLS_CERT
	CertKey string   `yaml:"certKey"` // SOLBK_TLS_CERTKEY
	CAs     []string `yaml:"cas"`     // SOLBK_TLS_CERTCAS
	// CertPassphrase unlocks an encrypted server-certificate key. It is a secret,
	// so on containers it is externalized like the admin password rather than
	// written into the deploy artifact. Empty means the key is not encrypted.
	CertPassphrase    string `yaml:"certPassphrase"`
	CertPassphraseEnv string `yaml:"certPassphraseEnv"` // env var holding certPassphrase instead
}

// Scaling is the broker's sizing, and every knob applies to every platform. They
// differ only in delivery: k8s writes them into the CR's spec.systemScaling,
// while docker and podman pass them to the container as environment variables
// under the same broker setting names (render.EnvPairs).
type Scaling struct {
	MaxConnections      int `yaml:"maxConnections"`      // system_scaling_maxconnectioncount
	MaxQueueMessages    int `yaml:"maxQueueMessages"`    // system_scaling_maxqueuemessagecount
	MaxSpoolUsageMB     int `yaml:"maxSpoolUsageMB"`     // messagespool_maxspoolusage / CR maxSpoolUsage
	MaxKafkaBridge      int `yaml:"maxKafkaBridge"`      // system_scaling_maxkafkabridgecount
	MaxKafkaConnections int `yaml:"maxKafkaConnections"` // system_scaling_maxkafkabrokerconnectioncount
	MaxBridges          int `yaml:"maxBridges"`          // system_scaling_maxbridgecount
	MaxSubscriptions    int `yaml:"maxSubscriptions"`    // system_scaling_maxsubscriptioncount
	MaxGuaranteedMsgMB  int `yaml:"maxGuaranteedMsgMB"`  // system_scaling_maxguaranteedmessagesize

	// MaxPool is retained so an env file carrying the removed maxPool fails with
	// an actionable error instead of a bare unknown-field decode error. It named
	// the same broker setting as MaxSpoolUsageMB -- one concept under two keys,
	// one per platform -- which is exactly what this block no longer has.
	MaxPool int `yaml:"maxPool"`

	// CPU is the broker CPU the MaxConnections tier fixes, derived in
	// ApplyDefaults once MaxConnections resolves (scaling.go). Not read from
	// YAML, so no env file can set it: k8s renders it as messagingNodeCpu,
	// docker and podman as their own CPU cap. It replaces the independently
	// settable kubernetes.msgNode.cpu, which could contradict the tier.
	CPU string `yaml:"-"`
}

// Replication is data replication: a DR pair of SEPARATE brokers, per message-VPN,
// each VPN active at one site and standby at the other. It is not Redundancy, which is
// one HA group inside a single deployment -- this block describes two of those.
//
// THE BLOCK IS BYTE-IDENTICAL AT BOTH SITES. Nothing in it is written from one site's
// point of view: a broker reads its own `show router-name`, finds itself in Sites, and
// whichever entry is not itself is its mate (Locate). So the DR site's env file carries
// this block verbatim, with nothing reversed or paired up -- which is the whole reason
// there is no `mate:` key and no per-site role.
//
// Replaces the dead REPL_* stub (mate/connSsl/psk), which nothing ever read.
// internal/convert no longer carries those variables over; it reports them as dropped.
type Replication struct {
	Sites []ReplSite `yaml:"sites"` // exactly 2, or none at all
	VPNs  []ReplVPN  `yaml:"vpns"`  // listed = replication enabled at both sites
}

// ReplSite is one end of the pair -- a single broker or a whole HA group. Three of its
// four fields say what the OTHER side needs to know about this one; only RouterNames
// says who this one is, which is what lets a single text be correct at both ends.
type ReplSite struct {
	// VirtualRouterName is the site's KEY: what ReplVPN.ActiveAt references, and the
	// literal operand the mate emits as `replication mate virtual-router-name "<v>"`.
	// Mandatory and never derived. An earlier design defaulted it to "v:" + the first
	// router name; stating it outright means the file holds the exact string the CLI
	// will be given, so nothing in the mate-address path is inferred.
	VirtualRouterName string `yaml:"virtualRouterName"`

	// RouterNames is what a broker ANSWERS TO -- matched against its own
	// `show router-name` so it can find itself. A list because a site can be an HA
	// pair and only ONE of its nodes appears in the virtual router name: run against
	// the other node and a single-name key would never match, leaving the broker
	// unable to locate itself in its own file.
	RouterNames []string `yaml:"routerNames"`

	// Endpoints is how the other BROKER dials this one. Never dialled by this tool --
	// it is rendered into the mate's CLI lines and nothing more. It is a flat list of
	// (host, port, transport) triples rather than hosts x ports, because the two are
	// not a cross product: a real capture carries two ssl entries on one port, which a
	// cross product could only express by inventing four more lines.
	Endpoints []ReplEndpoint `yaml:"endpoints"`

	// Via is how THIS TOOL reaches this site when it is the MATE. A site never reads
	// its own Via: running here, the tool takes its own access from the env file's
	// kubernetes:/docker:/podman: section like every other command. Only the
	// replication switch reads it, and only from the other entry -- so it is optional
	// at load and checked when a command actually needs it.
	Via ReplVia `yaml:"via"`
}

// Replication transports. The yaml spellings are this schema's; the CLI's own words are
// `plain-text`/`compressed`/`ssl` and the mapping lives in internal/broker, so nothing
// here has to know CLI syntax. There are exactly three: the request schema offers a
// fourth, `encrypted`, which is the DMR spelling of ssl and is deliberately not accepted
// -- the reply schema's six-entry cap is 3 transports x 2 addresses, which only works if
// those two words name one transport.
const (
	TransportPlainText  = "plainText"
	TransportCompressed = "compressed"
	TransportSSL        = "ssl"
)

// ReplEndpoint is one address the mate dials this site on. Transport may be empty,
// meaning plain text -- the CLI's own default, and how a capture spells an unqualified
// entry.
type ReplEndpoint struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	Transport string `yaml:"transport"` // plainText|compressed|ssl; empty means plainText
}

// ReplVia holds EXACTLY ONE mechanism. The key that is present is the discriminator,
// rather than a string naming one beside a separate block: that way a `via` cannot name
// a mechanism whose settings are missing, or name one while the other is filled. A whole
// class of mismatch stops being expressible instead of needing a rule to catch it.
//
// Pointers so an absent block is distinguishable from an empty one, which is what makes
// "exactly one" checkable at all.
type ReplVia struct {
	Kubernetes *ReplViaKube `yaml:"kubernetes"`
	SEMP       *ReplViaSEMP `yaml:"semp"`
}

// Set reports whether this site declares any mechanism.
func (v ReplVia) Set() bool { return v.Kubernetes != nil || v.SEMP != nil }

// ReplViaKube reaches the mate with its own cluster CLI. Command carries the cluster
// rather than a --context flag of this tool's: it then passes the same execution guard
// as kubernetes.runtime, so `kubectl --context dr` is checked by the machinery already
// in place and --allow-command stays the only door to widen the allowlist.
type ReplViaKube struct {
	Command   Command `yaml:"command"`   // THIS FIELD RUNS A BINARY (SiteCommand)
	Namespace string  `yaml:"namespace"` // the mate's namespace
	Name      string  `yaml:"name"`      // the mate's PubSubPlusEventBroker name
}

// ReplViaSEMP reaches the mate over SEMP, with curl exec'd inside THIS host's broker
// container -- so it needs no second kubeconfig and no new binary.
//
// Host is declared and never derived from Endpoints: replication runs over the message
// backbone, so a reachable replication endpoint proves nothing about SEMP reachability.
// TLS is declared for the same reason -- the local broker's own posture says nothing
// about a DR site's, and this hop is a WAN rather than a rack.
type ReplViaSEMP struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	TLS      bool   `yaml:"tls"`      // https rather than http; declared, never inferred
	Insecure bool   `yaml:"insecure"` // skip certificate verification (self-signed mate)

	// Exactly one of these supplies the mate's admin password. The username is always
	// config.AdminUser -- the broker's own name, not this schema's. There is no stdin
	// prompt: promptLine echoes, and a no-echo read is a dependency decision this
	// feature does not need to take.
	Pass       string          `yaml:"pass"`       // secret; discouraged, this file travels
	PassEnv    string          `yaml:"passEnv"`    // env var holding pass instead
	PassSecret *ReplPassSecret `yaml:"passSecret"` // a Kubernetes Secret holding it
}

// ReplPassSecret names a Kubernetes Secret holding the mate's admin password. It is read
// with that site's via.kubernetes.command when it has one, and this file's
// kubernetes.runtime otherwise -- so a site reached over SEMP can still keep its password
// in a cluster this machine can read.
type ReplPassSecret struct {
	Namespace string `yaml:"namespace"`
	Name      string `yaml:"name"`
	Key       string `yaml:"key"`
}

// ReplVPN is one replicated message-VPN. Listing it means replication is ENABLED for it
// at BOTH sites; a VPN replicating on the broker but absent from the list has its
// replication shut down. ActiveAt names the site that owns the active role, so the two
// sites derive opposite roles from the same sentence and two actives is not expressible.
type ReplVPN struct {
	Name     string `yaml:"name"`
	ActiveAt string `yaml:"activeAt"` // a site's virtualRouterName
}

// Configured reports whether the env file declares data replication at all. The block is
// optional: most deployments have no DR pair, and an absent section must not make every
// other command fail validation.
func (r Replication) Configured() bool { return len(r.Sites) > 0 || len(r.VPNs) > 0 }

// Locate finds this broker among the sites by the router name it reports, and returns
// its own entry and the mate's.
//
// All three non-happy paths are errors rather than guesses. No match usually means a
// typo, and the message prints what was read beside every name declared so the
// difference is visible. More than one match is a config error that Validate also
// prevents. There is deliberately no "assume the first": a wrong guess here points a
// switchover at the wrong broker.
func (r Replication) Locate(routerName string) (self, mate ReplSite, err error) {
	var found []int
	for i, s := range r.Sites {
		for _, n := range s.RouterNames {
			if n == routerName {
				found = append(found, i)
				break
			}
		}
	}
	switch len(found) {
	case 1:
		i := found[0]
		return r.Sites[i], r.Sites[1-i], nil
	case 0:
		var declared []string
		for _, s := range r.Sites {
			declared = append(declared, s.RouterNames...)
		}
		return ReplSite{}, ReplSite{}, fmt.Errorf("this broker reports router name %q, which is not in "+
			"replication.sites: the declared names are %s. Fix the name in the env file, or run this "+
			"command against the broker the file describes", routerName, strings.Join(declared, ", "))
	default:
		return ReplSite{}, ReplSite{}, fmt.Errorf("router name %q matches replication.sites[%d] (%s) and "+
			"replication.sites[%d] (%s): routerNames must not overlap", routerName,
			found[0], r.Sites[found[0]].VirtualRouterName, found[1], r.Sites[found[1]].VirtualRouterName)
	}
}

// Broker holds the post-deployment broker configuration every platform applies
// over the broker CLI, plus the local folders those operations read and write.
// It is platform-neutral on purpose: the container ops apply exactly the same
// domain certificates, product keys, and .cli scripts as the kubernetes ops do,
// and while these lived under kubernetes.* a container env file had to carry a
// kubernetes: section to reach them -- which made "which platform is this file
// for?" unanswerable from the file itself (DetectPlatforms).
type Broker struct {
	CLIScriptsFolder string      `yaml:"cliScriptsFolder"` // SOLBK_CLISCRIPTS_FOLDER
	DiagDir          string      `yaml:"diagDir"`          // SOLBK_DIAG_DIR
	ProductKeys      []string    `yaml:"productKeys"`      // SOLBK_PRODUCTKEYS
	DomainCerts      DomainCerts `yaml:"domainCerts"`
}

// K8sConfig holds everything specific to the operator-based Kubernetes deployment.
type K8sConfig struct {
	Runtime         Command `yaml:"runtime"`         // KUBE (default: kubectl)
	Name            string  `yaml:"name"`            // SOLBK_NAME
	Namespace       string  `yaml:"namespace"`       // SOLBK_NS
	AdminSecret     string  `yaml:"adminSecret"`     // SOLBK_USR_SECRET: Secret holding the admin/monitor creds
	TLSServerSecret string  `yaml:"tlsServerSecret"` // SOLBK_SVR_SECRET: TLS Secret built from tls.cert/certKey; enables the CR's TLS block
	ImagePullSecret string  `yaml:"imagePullSecret"` // IMAGEREPO_SECRET: dockerconfigjson Secret; enables imagePullSecrets
	// ImagePullPolicy is the k8s image pull policy: Always for a moving tag, Never
	// for an air-gapped cluster with the image preloaded. Empty keeps the CR's own
	// IfNotPresent, so an unset value renders exactly as before.
	ImagePullPolicy   string            `yaml:"imagePullPolicy"`
	UpdateStrategy    string            `yaml:"updateStrategy"` // automatedRolling|manualPodRestart
	ServiceAccount    string            `yaml:"serviceAccount"` // SOLBK_SVC_ACCOUNT (optional)
	Storage           Storage           `yaml:"storage"`
	MsgNode           Resources         `yaml:"msgNode"` // SOLBK_MSGNODE_CPU/MEM
	Operator          Operator          `yaml:"operator"`
	SecurityContext   PodSecurity       `yaml:"securityContext"`   // -> spec.securityContext
	ContainerSecurity ContainerSecurity `yaml:"containerSecurity"` // -> spec.brokerContainerSecurity
	PodAnnotations    map[string]string `yaml:"podAnnotations"`    // -> spec.podAnnotations
	PodLabels         map[string]string `yaml:"podLabels"`         // -> spec.podLabels
	Placement         Placement         `yaml:"placement"`
	LoadBalancer      LoadBalancer      `yaml:"loadBalancer"`
	Ports             []string          `yaml:"ports"` // SOLBK_PORTS "name=port[/proto]"
}

// Storage is how each broker node gets its data volume: provisioned from a StorageClass,
// or mounted from a PersistentVolumeClaim you created yourself.
//
// The two are mutually exclusive per file, not per role. `class` names a StorageClass for
// the operator to provision from; `customVolumeMount` hands it claims that already exist.
// A file that says both is describing two different storage stories, and which one wins
// would come down to which field the operator happened to read -- so validate refuses it.
//
// MsgNodeSize/MonNodeSize are the sizes the operator requests when it provisions. They are
// named "...Size" to separate them from kubernetes.msgNode, the resources block: the two
// were both spelled `msgNode` and read as the same word while meaning a volume size and a
// memory limit.
type Storage struct {
	Class       string `yaml:"class"`       // SOLBK_STORAGECLASS
	MsgNodeSize string `yaml:"msgNodeSize"` // SOLBK_STORAGE_MSGNODE
	MonNodeSize string `yaml:"monNodeSize"` // SOLBK_STORAGE_MONNODE

	// CustomVolumeMount maps a role to an EXISTING PersistentVolumeClaim, keyed by the
	// canonical role word (primary/backup/monitor). It becomes
	// spec.storage.customVolumeMount[] in the CR.
	//
	// The claims are yours: this tool never creates them and, crucially, never deletes
	// them -- `broker remove --delete-data` skips any role mounted this way and says so.
	// A claim named here can point at a volume holding data that predates this broker,
	// and there is no safe way for a tool to tell the difference.
	CustomVolumeMount map[string]string `yaml:"customVolumeMount"`
}

// CustomMountFor returns the claim configured for a role, if any. Roles are keyed by the
// canonical word, which is what the env file must use -- abbreviations are a CLI-input
// convenience and are refused in the file.
func (s Storage) CustomMountFor(r Role) (string, bool) {
	claim, ok := s.CustomVolumeMount[r.Word()]
	return claim, ok && claim != ""
}

// UsesCustomMounts reports whether any role is mounted from a claim the operator did not
// provision. It is what gates the storage-class rules and the removal's data handling.
func (s Storage) UsesCustomMounts() bool {
	for _, claim := range s.CustomVolumeMount {
		if claim != "" {
			return true
		}
	}
	return false
}

// Resources is the message-node resource block.
type Resources struct {
	// CPU is retained so an env file carrying the removed kubernetes.msgNode.cpu fails
	// with an actionable error instead of a bare unknown-field decode error. It
	// is never defaulted and never rendered: broker CPU is fixed by the scaling
	// tier (scaling.go, Scaling.CPU). validateK8s rejects any value here.
	CPU string `yaml:"cpu"`
	Mem string `yaml:"mem"` // SOLBK_MSGNODE_MEM (defaults to the scaling tier's memory)
}

// Operator is the cluster-scoped EventBroker Operator configuration.
type Operator struct {
	Image           string `yaml:"image"`           // SOLOP_IMAGE
	Namespace       string `yaml:"namespace"`       // SOLOP_NS (blank -> defaultOperatorNS)
	WatchNamespaces string `yaml:"watchNamespaces"` // SOLOP_WATCH_NS
	WatchBrokerNS   *bool  `yaml:"watchBrokerNs"`   // SOLOP_WATCH_SOLBK_NS (nil -> default true)
	CPU             string `yaml:"cpu"`             // SOLOP_CPU
	Mem             string `yaml:"mem"`             // SOLOP_MEM
}

// WatchBrokerNSEnabled reports whether the operator should also watch the
// broker namespace. Unset (nil) defaults to true, matching the bash default
// SOLOP_WATCH_SOLBK_NS=true; set it to false in YAML to opt out.
func (o Operator) WatchBrokerNSEnabled() bool {
	return o.WatchBrokerNS == nil || *o.WatchBrokerNS
}

// PodSecurity is the broker pod's securityContext. The ids are strings, not
// ints, so an explicit "0" (which asks OpenShift to auto-assign) stays
// distinguishable from an unset field -- the whole block is optional and is
// omitted from the CR when nothing is set.
type PodSecurity struct {
	RunAsUser string `yaml:"runAsUser"`
	FSGroup   string `yaml:"fsGroup"`
}

// Configured reports whether any field was set, which is what decides if the
// block reaches the CR at all.
func (s PodSecurity) Configured() bool { return s.RunAsUser != "" || s.FSGroup != "" }

// ContainerSecurity is the broker container's own security settings. Same
// optional-block rule as PodSecurity; ReadOnlyRootFilesystem is a pointer so an
// explicit false is not mistaken for "not configured".
type ContainerSecurity struct {
	RunAsUser              string `yaml:"runAsUser"`
	RunAsGroup             string `yaml:"runAsGroup"`
	ReadOnlyRootFilesystem *bool  `yaml:"readOnlyRootFilesystem"`
}

// Configured reports whether any field was set.
func (s ContainerSecurity) Configured() bool {
	return s.RunAsUser != "" || s.RunAsGroup != "" || s.ReadOnlyRootFilesystem != nil
}

// Placement controls broker pod scheduling (tolerations, node labels, anti-affinity).
type Placement struct {
	TolerationsPrimary []string `yaml:"tolerationsPrimary"`     // SOLBK_NODETOL_PRI
	TolerationsBackup  []string `yaml:"tolerationsBackup"`      // SOLBK_NODETOL_BKP
	TolerationsMonitor []string `yaml:"tolerationsMonitor"`     // SOLBK_NODETOL_MON
	LabelsPrimary      []string `yaml:"labelsPrimary"`          // SOLBK_NODELABEL_PRI
	LabelsBackup       []string `yaml:"labelsBackup"`           // SOLBK_NODELABEL_BKP
	LabelsMonitor      []string `yaml:"labelsMonitor"`          // SOLBK_NODELABEL_MON
	AntiAffinityNS     []string `yaml:"antiAffinityNamespaces"` // SOLBK_ANTIAFFINITY_NS
	AntiAffinityWeight int      `yaml:"antiAffinityWeight"`     // SOLBK_ANTIAFFINITY_WT

	// The blocks below are additive: unset, the rendered CR is exactly what
	// AntiAffinityNS/Weight and the label/toleration lists above produce, so an
	// existing env file is unaffected. They apply to every broker role -- per-role
	// affinity is not modelled, matching how AntiAffinityNS already behaves.
	NodeAffinity    NodeAffinity      `yaml:"nodeAffinity"`
	PodAffinity     []PodAffinityTerm `yaml:"podAffinity"`
	PodAntiAffinity []PodAffinityTerm `yaml:"podAntiAffinity"`
}

// NodeAffinity mirrors the Kubernetes nodeAffinity subset the operator passes
// through, rather than inventing a parallel spelling. Deliberately not modelled:
// matchFields, and OR-ed nodeSelectorTerms -- Required is one ANDed term.
type NodeAffinity struct {
	Preferred []WeightedNodeTerm `yaml:"preferred"`
	Required  []NodeMatchExpr    `yaml:"required"`
}

// Configured reports whether either list was set, which is what decides if a
// nodeAffinity block reaches the CR at all.
func (n NodeAffinity) Configured() bool { return len(n.Preferred) > 0 || len(n.Required) > 0 }

// WeightedNodeTerm is one weighted preference: every expression in Match must hold
// for the weight to apply.
type WeightedNodeTerm struct {
	Weight int             `yaml:"weight"` // 1-100
	Match  []NodeMatchExpr `yaml:"match"`
}

// NodeMatchExpr is a node-label match expression.
type NodeMatchExpr struct {
	Key      string   `yaml:"key"`
	Operator string   `yaml:"operator"` // In|NotIn|Exists|DoesNotExist|Gt|Lt
	Values   []string `yaml:"values"`   // required by In/NotIn/Gt/Lt
}

// PodAffinityTerm is one pod (anti-)affinity rule. Weight 0 makes it a required
// term; 1-100 makes it a preference with that weight. The pod selector is
// matchLabels only -- matchExpressions there is not modelled.
type PodAffinityTerm struct {
	Weight      int               `yaml:"weight"`
	TopologyKey string            `yaml:"topologyKey"`
	MatchLabels map[string]string `yaml:"matchLabels"`
	Namespaces  []string          `yaml:"namespaces"`
}

// LoadBalancer holds MetalLB / service-LB options.
type LoadBalancer struct {
	IP          string   `yaml:"ip"`          // SOLBK_LOADBALANCER_IP
	Annotations []string `yaml:"annotations"` // SOLBK_LOADBALANCER_ANOTN "key: value"
	IPPool      string   `yaml:"ipPool"`      // SOLBK_IPPOOL
}

// DomainCerts are trusted domain CA certificates loaded post-deploy.
type DomainCerts struct {
	Folder string            `yaml:"folder"` // SOLBK_DOMAINCERT_FOLDER
	Files  map[string]string `yaml:"files"`  // SOLBK_DOMAINCERT_FILES [CA-NAME]=filename
}

// DockerConfig holds docker-only deployment options plus the shared container block.
type DockerConfig struct {
	Runtime Command `yaml:"runtime"` // CONTAINER_RUNTIME override (default: docker)
	// Compose is the compose invocation, whose form differs per host: the modern
	// plugin is a runtime subcommand (`docker compose`), the standalone v1 binary
	// is its own executable (`docker-compose`). Unset defaults to the configured
	// runtime plus `compose`; this tool appends every argument after it.
	Compose     Command   `yaml:"compose"`
	ComposeFile string    `yaml:"composeFile"` // DOCKER_COMPOSE_FILE
	Network     Network   `yaml:"network"`
	Container   Container `yaml:"container"`
}

// PodmanConfig holds podman-only deployment options plus the shared container block.
type PodmanConfig struct {
	Runtime    Command `yaml:"runtime"`    // CONTAINER_RUNTIME override (default: podman)
	Rootless   bool    `yaml:"rootless"`   // PODMAN_ROOTLESS
	QuadletDir string  `yaml:"quadletDir"` // QUADLET_DIR override
	// BaseDir is the host directory for files THIS TOOL writes for podman, as
	// opposed to Config.BaseDir(), which is the directory the env FILE was loaded
	// from -- two different bases, so the doc comment says which.
	//
	// It exists because a quadlet unit cannot inline file content the way a compose
	// file can, so anything podman must read from a path has to be a real file
	// somewhere. Today that is the server-certificate bundle, which carries a
	// PRIVATE KEY -- which is why this is mandatory rather than defaulted: where a
	// private key lands on the host is the operator's decision, not a value this
	// tool picks. Deliberately separate from QuadletDir: the unit must live where
	// systemd scans, and a stray tls.pem in that directory is at best ignored.
	BaseDir   string    `yaml:"baseDir"`
	Network   Network   `yaml:"network"`
	Container Container `yaml:"container"`

	// Derived from Rootless in ApplyDefaults (not read from YAML).
	SystemctlUser string `yaml:"-"`
	WantedBy      string `yaml:"-"`
}

// Network is the container networking mode + published ports.
type Network struct {
	Mode  string   `yaml:"mode"`  // host|bridge (SOLBK_NETWORK_MODE)
	Ports []string `yaml:"ports"` // SOLBK_PORTS host:container (required for bridge)
}

// Container is the shared docker/podman container runtime settings.
type Container struct {
	Name    string `yaml:"name"`    // CONTAINER_NAME
	RunUser string `yaml:"runUser"` // SOLBK_RUN_USER uid:gid
	ShmSize string `yaml:"shmSize"` // SOLBK_SHM_SIZE
	// Mem is the container memory limit in docker's and podman's own b|k|m|g
	// suffix, NOT the Mi/Gi Kubernetes quantity kubernetes.msgNode.mem takes -- the
	// engines reject that spelling, so validateContainer catches it here rather
	// than letting compose fail at deploy. Defaults to the scaling tier's memory
	// (scaling.go). CPU has no counterpart: it is fixed by the tier, so there is
	// nothing here to override.
	Mem         string      `yaml:"mem"`
	DataDir     string      `yaml:"dataDir"` // SOLBK_DATA_DIR (host bind mount)
	Ulimits     Ulimits     `yaml:"ulimits"`
	HealthCheck HealthCheck `yaml:"healthCheck"`
}

// HealthCheck is the container engine's own probe against the broker, which
// upgrades `docker ps`/`compose ps` and podman's auto-restart from "the process is
// up" to "the broker is ready". It is opt-in; leaving the whole block out renders
// the artifacts unchanged.
//
// Enabled with no Cmd uses the built-in readiness probe (the broker's own
// /health-check/readiness endpoint). That endpoint only exists from
// HealthCheckMinMajor.HealthCheckMinMinor onward, so Validate refuses to enable it
// against an older -- or an unidentifiable -- image tag: an always-failing probe
// would mark the container permanently unhealthy, which under podman's
// auto-restart becomes a restart loop. Setting Cmd explicitly is the escape hatch
// and skips the version gate, since the probe is then the operator's own choice.
type HealthCheck struct {
	Enabled bool `yaml:"enabled"`
	// Cmd is the probe argv, run inside the container. Empty means the built-in
	// readiness probe. The quadlet form is a command line rather than an argv, so a
	// token containing a space is not representable there -- wrap such a probe in a
	// script instead.
	Cmd         []string `yaml:"cmd"`
	Interval    string   `yaml:"interval"`
	Timeout     string   `yaml:"timeout"`
	Retries     int      `yaml:"retries"`
	StartPeriod string   `yaml:"startPeriod"`
}

// The first broker release exposing /health-check/readiness, which the built-in
// probe depends on.
const (
	HealthCheckMinMajor = 10
	HealthCheckMinMinor = 26
)

// Ulimits are the container resource limits (soft:hard where applicable).
type Ulimits struct {
	NoFile  string `yaml:"nofile"`  // SOLBK_ULIMIT_NOFILE
	MemLock string `yaml:"memlock"` // SOLBK_ULIMIT_MEMLOCK
	Core    string `yaml:"core"`    // SOLBK_ULIMIT_CORE
}

// Redundancy is everything about the HA group: whether there is one, who is in it, and
// the key its members authenticate to each other with.
//
// The on/off switch lives here rather than as a separate top-level scalar because the two
// were always one decision -- `redundancy.enabled: true` with no node table is not a deployable
// state, and neither is a node table nobody reads.
type Redundancy struct {
	// Enabled is "true" or "false" as written; RedundancyEnabled is the reader. It stays
	// a string rather than a bool so a value that is neither gets this schema's own
	// actionable error instead of yaml's. Omitted means false: HA provisions three
	// brokers, so it is opted into, never inherited.
	Enabled string `yaml:"enabled"`

	Primary Node `yaml:"primary"` // message_routing (also the standalone broker)
	Backup  Node `yaml:"backup"`  // message_routing (redundancy only)
	Monitor Node `yaml:"monitor"` // monitoring (redundancy only)

	// PSK is the pre-shared key the group's members authenticate with, and how it
	// reaches the broker differs by platform even though the value does not.
	//
	// On docker and podman it is MANDATORY -- refused at load when empty, with the
	// `openssl rand -base64 32` command in the message -- and is externalized like every
	// other container secret: mounted at
	// /mnt/secrets/redundancy_authentication_presharedkey_key and read through the
	// setting's own *keyfilepath variant. Nothing generates it: a key one host invented
	// is a key the other two never see, so it is the operator's to make and to copy.
	//
	// On Kubernetes it is OPTIONAL. Left empty, the operator generates its own key and
	// spec.preSharedAuthKeySecret is omitted entirely; set, the value becomes the
	// `preshared_auth_key` entry of the credentials Secret this tool already builds, and
	// the CR points at that Secret. The CRD fixes that key name, so it is not ours to
	// choose.
	PSK    string `yaml:"psk"`    // SOLBK_REDUNDANCY_PSK (secret)
	PSKEnv string `yaml:"pskEnv"` // env var holding psk instead
}

// Node is one redundancy-group member.
type Node struct {
	// Name is the broker's routername, and the container's hostname on docker/podman.
	//
	// MANDATORY for all three members of an HA container group, and not defaultable:
	// each name is the KEY of that node's entry in the group table EVERY host renders
	// (redundancy_group_node_<name>_connectvia), so a host filling in one it invented
	// would build a table the other two disagree with and the group would never form.
	//
	// OPTIONAL in a standalone container deployment, where there is one node and it is
	// always this host: the CLI fills it from the host's own OS hostname at load
	// (Config.FillStandaloneNodeName), which is what the broker would have called
	// itself anyway. Not read at all on Kubernetes -- the operator names the pods after
	// kubernetes.name.
	Name string `yaml:"name"`
	// Addr is the host the other members reach this node at -- an IP or a resolvable
	// name. Named addr rather than ip because a DNS name has always been valid here and
	// calling it "ip" invited an address-shaped answer to a question that accepts both.
	Addr string `yaml:"addr"`
}

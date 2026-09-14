package config

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// Validate checks mandatory and enum fields for the given platform, mirroring
// the mandatory-vars and enum checks in the two bash bootstraps. It fails loud
// with an actionable message listing every offending field at once.
func (c *Config) Validate(p Platform) error {
	// The section was renamed, so an env file carrying the old name gets told that rather
	// than a bare "field admin not found in type config.Config". Two keys went with the
	// rename and are named too, since a mechanical `admin:` -> `semp:` edit leaves them
	// behind and the resulting error would otherwise be about the wrong thing.
	if c.LegacyAdmin != nil {
		return fmt.Errorf("the `admin:` section was renamed to `semp:` (it holds the broker's SEMP/CLI " +
			"credentials). Rename it, and inside it: `pass:` -> `adminPass:`, `passEnv:` -> `adminPassEnv:`. " +
			"`user:` was REMOVED -- the admin user is always 'admin' on every platform now, the way " +
			"Kubernetes already required. monitorPass, monitorPassEnv and additionalUsers keep their names")
	}

	// redundancy.enabled is a shared enum on every platform.
	switch c.Redundancy.Enabled {
	case "true", "false":
	default:
		return fmt.Errorf("redundancy.enabled must be 'true' or 'false' (got: %q)", c.Redundancy.Enabled)
	}

	// The platform CLI is user-supplied and reaches os/exec, so it goes through the
	// full execution guard (execguard.go) on every platform before anything can
	// run. This is the first of the two enforcement points; every executor re-runs
	// the same CheckCommand immediately before it builds argv, so a hostile env
	// file is inert even on a path that never reached Validate.
	if err := c.validateExecCommands(p); err != nil {
		return err
	}

	// Scaling applies to every platform -- k8s through the CR, containers through
	// the environment -- so it is checked once here (scaling.go).
	if err := c.validateScaling(); err != nil {
		return err
	}

	// These secrets are shared top-level fields (not platform-scoped), and each
	// reaches a consumer that cannot tolerate a control character in the value:
	// admin.pass and admin.monitorPass reach broker.sempCurl's curl config on
	// stdin, redundancy.psk and tls.certPassphrase travel the same way through the
	// container config path. Checked once here rather than per platform, since
	// the fields exist regardless of which platform ends up reading them.
	if err := c.validateCredentialChars(); err != nil {
		return err
	}

	// Replication is platform-neutral: the same DR pair is describable from a
	// kubernetes, docker or podman env file, and which platform reads it changes
	// nothing about whether the block is well formed. Checked once here, beside
	// scaling, rather than inside each platform's validator.
	if err := c.validateReplication(); err != nil {
		return err
	}

	// Host paths are the OTHER charset (hostpath.go). A command token becomes argv;
	// a host path becomes a value concatenated into a quadlet or compose line, so
	// the two rules differ and the difference is derived rather than duplicated.
	//
	// This runs on the values AS WRITTEN. Load rebases relative paths onto the env
	// file's directory only AFTER Validate returns, deliberately: the rule polices
	// what the FILE says, which is the only thing this tool controls, and checking a
	// rebased value would fail every load from a checkout whose own directory
	// happens to contain a space.
	if err := c.validateHostPaths(p); err != nil {
		return err
	}

	switch p {
	case K8s:
		return c.validateK8s()
	case Docker, Podman:
		return c.validateContainer(p)
	default:
		return fmt.Errorf("unknown platform %q", p)
	}
}

// validateHostPaths runs CheckHostPath over every field whose value is a path on
// the machine running this tool. The list is per platform because the fields are:
// only a container platform has a compose file or a data dir, and only kubernetes
// reads the certificate files into a Secret -- but tls.* and broker.* are
// platform-neutral and checked everywhere, since a container deployment reads the
// same certificate and the same script folder.
//
// TestValidateHostPathsCoversEveryHostPathField pins the list against
// rebaseHostPaths, so a field that gains a rebase without a check, or the reverse,
// fails rather than quietly diverging.
func (c *Config) validateHostPaths(p Platform) error {
	fields := []struct{ field, value string }{
		{"tls.cert", c.TLS.Cert},
		{"tls.certKey", c.TLS.CertKey},
		{"broker.cliScriptsFolder", c.Broker.CLIScriptsFolder},
		{"broker.diagDir", c.Broker.DiagDir},
		{"broker.domainCerts.folder", c.Broker.DomainCerts.Folder},
	}
	for i, ca := range c.TLS.CAs {
		fields = append(fields, struct{ field, value string }{fmt.Sprintf("tls.cas[%d]", i), ca})
	}
	switch p {
	case Docker:
		fields = append(fields,
			struct{ field, value string }{"docker.composeFile", c.Docker.ComposeFile},
			struct{ field, value string }{"docker.container.dataDir", c.Docker.Container.DataDir},
		)
	case Podman:
		fields = append(fields,
			// The quadlet dir is checked but never rebased: the unit has to live
			// where systemd scans, so a relative value is an operator error rather
			// than something to resolve helpfully.
			struct{ field, value string }{"podman.quadletDir", c.Podman.QuadletDir},
			// Likewise checked but never rebased, and required absolute above: it is
			// a `Volume=` source, and resolving a relative one would invent a
			// location for a private key that the operator did not name.
			struct{ field, value string }{"podman.baseDir", c.Podman.BaseDir},
			struct{ field, value string }{"podman.container.dataDir", c.Podman.Container.DataDir},
		)
	}
	for _, f := range fields {
		if err := CheckHostPath(f.field, f.value); err != nil {
			return err
		}
	}
	return nil
}

// validateProbeCommand checks the tokens of a command this tool does NOT execute:
// the container health-check probe, which is rendered into the compose/quadlet
// artifact and run by the container engine INSIDE the broker container. It never
// becomes argv on the operator's machine, so the execution guard's allowlist and
// subcommand rules would be meaningless here -- a probe is legitimately
// `sh -c 'curl ... || exit 1'`, and the engine, not this process, decides what it
// means. What still applies is the exec-boundary check the field always had: an
// empty argument, or a control character carried in from a converted bash file,
// can only ever fail obscurely (§4a).
//
// An empty Command is not an error: ApplyDefaults runs before Validate on every
// path and fills the platform default, so "empty" means "unset" exactly as it
// does for every setDefault field in this schema.
func validateProbeCommand(field string, cmd Command) error {
	for i, tok := range cmd {
		if tok == "" {
			return fmt.Errorf("%s[%d] is an empty argument; remove it or quote the intended value", field, i)
		}
		if j := strings.IndexFunc(tok, isCtrl); j >= 0 {
			return fmt.Errorf("%s[%d] contains a control character (0x%02x) at offset %d: %q",
				field, i, tok[j], j, tok)
		}
	}
	return nil
}

// isCtrl reports the ASCII control characters (including NUL and DEL), which can
// never legitimately appear in a command name or argument here.
func isCtrl(r rune) bool { return r < 0x20 || r == 0x7f }

// identRE constrains the identifiers that reach a rendered artifact in a
// structural position: a compose service key, a `container_name`, a systemd
// ContainerName=/HostName=, and the systemd Environment= keys built from node
// names. A colon, '=' or newline there produces a broken artifact instead of an
// error, so the check belongs here where it can name the field (§4a).
//
// It is a package-local copy of broker.nameRE / k8s.secretKeyUserRE, which are
// already the same expression: config sits below both in the import graph, and
// the house convention is one small copy per package over a shared micro-package.
var identRE = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// brokerUsernameRE is the broker's OWN grammar for a CLI username: it must start with a
// letter or '_', and the whole name is 1-32 characters. The tail keeps identRE's charset,
// which is what the derived secret names and broker settings are built from.
//
// Checked here rather than left to the broker, because the broker rejects it at CREATE
// time -- on a running deployment, after everything else has already been applied.
var brokerUsernameRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9._-]{0,31}$`)

// envVarNameRE is the C-identifier shape Kubernetes requires of an environment variable
// name. It is NARROWER than identRE, which admits '.', '_' and '-': those are legal in a
// Secret KEY but not in the variable envFrom projects it into, and the kubelet drops the
// ones that do not match rather than failing the pod.
var envVarNameRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// engineContainerNameRE is docker's and podman's own grammar for a container name:
// identRE's charset, plus the requirement that the FIRST character be alphanumeric.
// The same expression internal/container applies to a name the engine hands back
// (engineNameRE) -- one grammar, checked on the way in and on the way out.
var engineContainerNameRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// maxContainerNameLen bounds container.name. It is the stem of every derived name --
// the podman unit and service, and the host-side secret names, which add up to about
// twenty characters more -- so it is set well inside the engines' own limit rather
// than at it.
const maxContainerNameLen = 100

// runUserRE allows the container runtime's `uid[:gid]` form, which identRE alone
// would reject -- the default "0:0" contains a colon.
var runUserRE = regexp.MustCompile(`^[A-Za-z0-9._-]+(:[A-Za-z0-9._-]+)?$`)

// dnsLabelBodyRE is the Kubernetes DNS-1123 label charset and start/end rule
// (RFC 1123): lowercase alphanumerics and '-', starting and ending with an
// alphanumeric. It carries no length bound of its own because the two callers
// below cap it differently -- validDNSLabel at 63 (a Kubernetes object name),
// validPortName at 15 (a Service port name) -- so the length is checked beside
// each cap instead of baked into one regex neither could reuse.
//
// This is deliberately NOT identRE just above: identRE is the container-side
// rule ("^[A-Za-z0-9._-]+$"), which permits uppercase, '.' and '_' -- none of
// which Kubernetes accepts in a label -- and has no length bound at all. Reusing
// it here would let a value through that kubectl then rejects, or worse, one
// that splices into the hand-built Secret/namespace YAML these fields reach by
// string concatenation (k8s/secrets.go, k8s/prep.go) or the CR (render/render.go):
// a newline plus "---" adds an extra document to the applied stream, and a bare
// colon breaks the mapping. identRE stays exactly as it is for the container
// fields that already rely on its looser rule.
var dnsLabelBodyRE = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// maxDNSLabelLen is the general Kubernetes object-name limit (a Secret,
// namespace, or CR name).
const maxDNSLabelLen = 63

// brokerPodSuffixShape is the longest thing the operator appends to kubernetes.name
// when deriving an object name: the pod's "-pubsubplus-<role letter>-0". Only its
// LENGTH is used, so the letter here stands for any role.
//
// It is spelled out rather than imported because internal/config must not depend on
// internal/k8s -- config is the leaf every platform package reads. The copy is the
// smaller cost, and TestPodNameSuffixMatchesTheConfigBound in internal/k8s is what
// stops the two drifting: it derives a real pod name and checks the length this
// constant claims.
const brokerPodSuffixShape = "-pubsubplus-p-0"

// BrokerPodSuffixShape exposes that suffix so internal/k8s can assert its own
// derivation still matches the bound applied here. An accessor rather than an exported
// constant: nothing outside a test has any business reading it, and a function makes
// that obvious at the call site.
func BrokerPodSuffixShape() string { return brokerPodSuffixShape }

// maxDNSSubdomainLen is the Kubernetes limit for a DNS-1123 SUBDOMAIN -- the rule
// the API applies to most object names, Secrets included.
const maxDNSSubdomainLen = 253

// dnsSubdomainRE is one or more DNS-1123 labels joined by single dots. It is the
// rule Kubernetes really enforces on a Secret name, as opposed to the stricter
// single-label rule a namespace or Service name must meet -- see the two callers
// in validateK8s for which fields get which and why.
var dnsSubdomainRE = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)*$`)

// validDNSSubdomain checks one Kubernetes DNS-1123 subdomain, ignoring an empty
// value: every field that reaches it is optional, so an absent value stays legal
// and only a non-empty malformed one is rejected.
func validDNSSubdomain(field, value string) error {
	if value == "" {
		return nil
	}
	if len(value) > maxDNSSubdomainLen || !dnsSubdomainRE.MatchString(value) {
		return fmt.Errorf("%s %q is invalid: must be a Kubernetes DNS-1123 subdomain -- lowercase "+
			"alphanumerics, '-' and '.', each dot-separated part starting and ending with an "+
			"alphanumeric, at most %d characters", field, value, maxDNSSubdomainLen)
	}
	return nil
}

// validDNSLabel checks one Kubernetes DNS-1123 label, ignoring an empty value:
// kubernetes.namespace and kubernetes.name are mandatory and already reported by
// requireAll, and the remaining fields (the three secret names, operator.namespace)
// are optional -- an absent value must stay legal (§4a/rule 8), so only a
// non-empty malformed one is rejected here.
func validDNSLabel(field, value string) error {
	if value == "" {
		return nil
	}
	if len(value) > maxDNSLabelLen || !dnsLabelBodyRE.MatchString(value) {
		return fmt.Errorf("%s %q is invalid: must be a Kubernetes DNS-1123 label -- lowercase "+
			"alphanumerics and '-', starting and ending with an alphanumeric, at most %d characters",
			field, value, maxDNSLabelLen)
	}
	return nil
}

// maxPortNameLen is the Kubernetes Service port name limit -- shorter than the
// general 63-character object-name bound above, which is why it gets its own
// cap rather than sharing validDNSLabel's.
const maxPortNameLen = 15

// validPortName checks a kubernetes.ports entry's name: the same DNS-1123
// charset validDNSLabel checks, but capped at 15 characters instead of 63.
func validPortName(field, value string) error {
	if len(value) > maxPortNameLen || !dnsLabelBodyRE.MatchString(value) {
		return fmt.Errorf("%s %q is invalid: a Kubernetes port name must be a DNS-1123 label (lowercase "+
			"alphanumerics and '-', starting and ending with an alphanumeric) of at most %d characters",
			field, value, maxPortNameLen)
	}
	return nil
}

// validIdent checks one identifier, ignoring an empty value: emptiness is
// requireAll's job, and several of these fields are legitimately empty (the
// backup/monitor rows in standalone).
func validIdent(field, value string) error {
	if value == "" {
		return nil
	}
	if !identRE.MatchString(value) {
		return fmt.Errorf("%s %q is invalid: only letters, digits, '.', '_' and '-' are allowed "+
			"(it becomes a container name, host name and systemd/compose key)", field, value)
	}
	return nil
}

// keyValueEntries is a named list of user-supplied "key: value" fragments, kept as
// a slice rather than a map so a failure message is deterministic.
type keyValueEntries struct {
	field   string
	entries []string
}

// requireKeyValue checks that each entry carries the "key: value" shape the
// renderer emits as a YAML mapping entry. The renderer quotes both halves, so any
// character is safe once the shape holds -- what cannot be recovered is an entry
// with no key at all.
func requireKeyValue(groups []keyValueEntries) error {
	for _, g := range groups {
		for i, entry := range g.entries {
			key, _, ok := strings.Cut(entry, ":")
			if !ok || strings.TrimSpace(key) == "" {
				return fmt.Errorf("%s[%d] = %q is not a \"key: value\" entry; write it as `key: value`", g.field, i, entry)
			}
		}
	}
	return nil
}

func (c *Config) validateK8s() error {
	required := map[string]string{
		"kubernetes.name":      c.K8s.Name,
		"kubernetes.namespace": c.K8s.Namespace,
		"image.repo":           c.Image.Repo,
		"image.tag":            c.Image.Tag,
		"semp.adminPass":       c.SEMP.AdminPass, // hardening: no hardcoded default password
	}
	// A size is what the operator asks a StorageClass to provision. With custom volume
	// mounts the volumes already exist at whatever size they were created, so demanding
	// one would be demanding a number that changes nothing.
	if !c.K8s.Storage.UsesCustomMounts() {
		required["kubernetes.storage.msgNodeSize"] = c.K8s.Storage.MsgNodeSize
	}
	if missing := requireAll(required); len(missing) > 0 {
		return missingErr(missing)
	}
	if err := c.validateStorage(); err != nil {
		return err
	}
	// These reach a hand-built Secret/namespace manifest by string concatenation
	// (k8s/secrets.go, k8s/prep.go) or the CR (render/render.go), so they are
	// checked as Kubernetes DNS-1123 labels rather than merely non-empty. The
	// three secret names and operator.namespace are optional and skip an empty
	// value; namespace and name are already known non-empty at this point.
	// Two rules, because Kubernetes has two. A namespace must be a single LABEL,
	// and kubernetes.name must be one too -- it is suffixed into the StatefulSet,
	// Service and pod names the operator derives (k8s/names.go), which are Service
	// names and so cannot carry a dot.
	for _, f := range []struct{ field, value string }{
		{"kubernetes.namespace", c.K8s.Namespace},
		{"kubernetes.name", c.K8s.Name},
		{"kubernetes.operator.namespace", c.K8s.Operator.Namespace},
	} {
		if err := validDNSLabel(f.field, f.value); err != nil {
			return err
		}
	}
	// The DERIVED names, not just the typed one. kubernetes.name is suffixed into
	// every object the operator creates, and the longest of those is the pod:
	// <name>-pubsubplus-<role>-0. Kubernetes stamps a pod's own name into the
	// statefulset.kubernetes.io/pod-name LABEL, and a label VALUE is capped at 63 --
	// so a name that passes the check above on its own can still produce a pod the
	// cluster refuses, with an error naming a label nobody wrote.
	if n := len(c.K8s.Name) + len(brokerPodSuffixShape); n > maxDNSLabelLen {
		return fmt.Errorf("kubernetes.name %q is %d characters, which is too long once the operator's suffixes "+
			"are added: the derived pod name would be %d characters (%q), and Kubernetes copies a pod's name into "+
			"a label whose values stop at %d.\n  Use at most %d characters",
			c.K8s.Name, len(c.K8s.Name), n, c.K8s.Name+brokerPodSuffixShape, maxDNSLabelLen,
			maxDNSLabelLen-len(brokerPodSuffixShape))
	}
	// Secret names are the looser DNS-1123 SUBDOMAIN, which is what the Kubernetes
	// API actually enforces for them: dots are legal, up to 253 characters. Holding
	// them to the label rule would reject "prod.solace-admin-secret" -- a real
	// naming convention, accepted by kubectl today -- for no benefit, since a dot
	// is harmless in the hand-built YAML these reach. The injection risks that
	// motivated this check are a newline plus "---" and a bare colon, and the
	// subdomain rule excludes both just as firmly.
	for _, f := range []struct{ field, value string }{
		{"kubernetes.adminSecret", c.K8s.AdminSecret},
		{"kubernetes.tlsServerSecret", c.K8s.TLSServerSecret},
		{"kubernetes.imagePullSecret", c.K8s.ImagePullSecret},
	} {
		if err := validDNSSubdomain(f.field, f.value); err != nil {
			return err
		}
	}
	// Both halves or neither. This was enforced only on docker and podman while
	// applyK8sDefaults filled the pair together; now that naming the Secret no longer
	// invents the paths, a lone cert here would build a Secret with no tls.key in it --
	// which the operator mounts, and the broker then fails to start a listener over.
	if (c.TLS.Cert == "") != (c.TLS.CertKey == "") {
		return fmt.Errorf("tls.cert and tls.certKey must be set together (got cert=%q key=%q): "+
			"the TLS Secret carries both, and one without the other builds a Secret the broker cannot use. "+
			"To point at a Secret that already exists, set kubernetes.tlsServerSecret alone and leave both of these unset",
			c.TLS.Cert, c.TLS.CertKey)
	}
	// Material with nowhere to put it. The Secret this tool would build has no name, and
	// the CR would have no tls block, so the certificate would be silently unused.
	if c.ManagesTLSSecret() && c.K8s.TLSServerSecret == "" {
		return fmt.Errorf("tls.cert/tls.certKey are set but kubernetes.tlsServerSecret is not: " +
			"the Secret built from them needs a name, and the broker CR references it by that name")
	}
	if c.K8s.MsgNode.CPU != "" {
		// Removed rather than ignored: a stale cpu: in an env file is a sizing
		// decision the operator believes is in effect, so it has to be seen.
		return fmt.Errorf("kubernetes.msgNode.cpu was removed; broker CPU is fixed by the scaling tier and "+
			"derived from scaling.maxConnections (one of %s) -- drop the key. "+
			"kubernetes.msgNode.mem is unaffected: it still overrides the tier's default memory", scalingTierList)
	}
	switch c.K8s.UpdateStrategy {
	case "automatedRolling", "manualPodRestart":
	default:
		return fmt.Errorf("kubernetes.updateStrategy must be 'automatedRolling' or 'manualPodRestart' (got: %q)", c.K8s.UpdateStrategy)
	}
	switch c.K8s.ImagePullPolicy {
	case "", "Always", "IfNotPresent", "Never":
	default:
		return fmt.Errorf("kubernetes.imagePullPolicy must be 'Always', 'IfNotPresent' or 'Never' (got: %q)", c.K8s.ImagePullPolicy)
	}
	if err := c.validateK8sPorts(); err != nil {
		return err
	}
	pl := c.K8s.Placement
	if err := requireKeyValue([]keyValueEntries{
		{"kubernetes.loadBalancer.annotations", c.K8s.LoadBalancer.Annotations},
		{"kubernetes.placement.labelsPrimary", pl.LabelsPrimary},
		{"kubernetes.placement.labelsBackup", pl.LabelsBackup},
		{"kubernetes.placement.labelsMonitor", pl.LabelsMonitor},
	}); err != nil {
		return err
	}
	if err := c.validateAdditionalUsers(K8s); err != nil {
		return err
	}
	return validatePlacementAffinity(pl)
}

// validateStorage enforces the one-storage-story rule and checks the claims.
//
// A broker node gets its data volume one of two ways: provisioned by the operator from a
// StorageClass, or mounted from a PersistentVolumeClaim that already exists. A file naming
// both is describing two, and which one the operator honours is not something this tool
// should guess on the operator's behalf -- so it is refused rather than resolved.
func (c *Config) validateStorage() error {
	st := c.K8s.Storage
	if !st.UsesCustomMounts() {
		return nil
	}

	if st.Class != "" {
		return fmt.Errorf("kubernetes.storage.class and kubernetes.storage.customVolumeMount cannot both be set: "+
			"a storage class asks the operator to PROVISION a volume, a custom mount hands it one that already "+
			"exists, and the CRD does not say which wins.\n"+
			"  Drop kubernetes.storage.class to use the custom claim(s), or drop customVolumeMount to provision "+
			"from %q", st.Class)
	}

	// Keys are the canonical role words. An unknown key is a typo, and a typo here is the
	// dangerous kind: the role it was meant for falls back to default provisioning,
	// quietly, on a broker whose storage someone deliberately took control of.
	known := map[string]bool{}
	for _, r := range []Role{Primary, Backup, Monitor} {
		known[r.Word()] = true
	}
	for key, claim := range st.CustomVolumeMount {
		if !known[key] {
			return fmt.Errorf("kubernetes.storage.customVolumeMount.%s is not a role: expected one of %s",
				key, strings.Join(RoleNames(), ", "))
		}
		if strings.TrimSpace(claim) == "" {
			return fmt.Errorf("kubernetes.storage.customVolumeMount.%s is empty: name the PersistentVolumeClaim "+
				"to mount, or remove the key", key)
		}
	}

	// Every role in the redundancy group, or none. A half-covered HA group is far more
	// likely to be a mistyped key than a deliberate mix, and the failure mode is silent:
	// the uncovered node provisions from the cluster default and nothing says so.
	//
	// Roles OUTSIDE the group are ignored rather than refused -- a backup entry in a
	// standalone file describes a node that does not exist, which is harmless and is
	// exactly what a file switched from HA to standalone looks like.
	var missing []string
	for _, r := range c.redundancyRoles() {
		if _, ok := st.CustomMountFor(r); !ok {
			missing = append(missing, "kubernetes.storage.customVolumeMount."+r.Word())
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("customVolumeMount must cover every node in the redundancy group or none of them; "+
			"missing: %s.\n"+
			"  A role left out is provisioned from the cluster's default StorageClass instead, which is not "+
			"something a half-finished list should decide silently", strings.Join(missing, ", "))
	}
	return nil
}

// redundancyRoles is the set of broker nodes this config actually deploys.
func (c *Config) redundancyRoles() []Role {
	if c.RedundancyEnabled() {
		return []Role{Primary, Backup, Monitor}
	}
	return []Role{Primary}
}

// validateK8sPorts checks kubernetes.ports ("name=port[/proto]"), which
// render.parsePort turns into containerPort/servicePort/protocol/name fields
// spliced UNQUOTED into the broker CR's service.ports list -- a "name:port"
// typo (no '=') currently renders containerPort as YAML null and fails as an
// opaque kubectl error against a manifest the operator never sees. This mirrors
// parsePort's own parsing (render/render.go) rather than reimplementing it
// independently, so every shape parsePort accepts is either accepted here too or
// deliberately rejected below: the name must be a Kubernetes port name (a
// DNS-1123 label, capped at 15 rather than 63 -- see validPortName); the
// container port is mandatory, the service port defaults to it exactly as
// parsePort's own container/service split does; the protocol defaults to TCP
// and is otherwise accepted case-insensitively (parsePort never folds case
// itself, so an operator's "tcp"/"udp" must still validate). A container or
// service half that is present but non-numeric (parsePort would pass the raw
// string straight through as an unquantified YAML value) is rejected here
// rather than reaching the CR malformed. Port names and container ports must
// each be unique, so two entries cannot silently overwrite the same service
// port.
func (c *Config) validateK8sPorts() error {
	names := make(map[string]int, len(c.K8s.Ports))
	containerPorts := make(map[int]int, len(c.K8s.Ports))
	for i, entry := range c.K8s.Ports {
		field := fmt.Sprintf("kubernetes.ports[%d]", i)
		name, rest, ok := strings.Cut(entry, "=")
		if !ok {
			return fmt.Errorf("%s = %q must have the form name=port[/proto] (e.g. tcp-web=8008)", field, entry)
		}
		if err := validPortName(field+" name", name); err != nil {
			return err
		}
		proto := "TCP"
		if j := strings.LastIndex(rest, "/"); j >= 0 {
			proto = rest[j+1:]
			rest = rest[:j]
		}
		switch strings.ToUpper(proto) {
		case "TCP", "UDP":
		default:
			return fmt.Errorf("%s protocol %q must be TCP or UDP", field, proto)
		}
		container := rest
		service := rest
		if j := strings.Index(rest, ":"); j >= 0 {
			container = rest[:j]
			service = rest[j+1:]
		}
		containerN, err := validPortNumber(field+" container port", container)
		if err != nil {
			return err
		}
		if _, err := validPortNumber(field+" service port", service); err != nil {
			return err
		}
		if prev, dup := names[name]; dup {
			return fmt.Errorf("%s: port name %q is also used by kubernetes.ports[%d]; names must be unique", field, name, prev)
		}
		names[name] = i
		if prev, dup := containerPorts[containerN]; dup {
			return fmt.Errorf("%s: container port %d is also used by kubernetes.ports[%d]; container ports must be unique", field, containerN, prev)
		}
		containerPorts[containerN] = i
	}
	return nil
}

// validPortNumber parses one port half (container or service) and bounds it to
// the range a Kubernetes Service port actually accepts.
func validPortNumber(field, value string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s %q must be numeric", field, value)
	}
	if n < 1 || n > 65535 {
		return 0, fmt.Errorf("%s %d must be between 1 and 65535", field, n)
	}
	return n, nil
}

// accessLevels are the broker's global access levels, in increasing order of
// privilege. The value reaches the broker as a username_<user>_globalaccesslevel
// setting (containers) or a `global-access-level` CLI attribute (k8s), so an invalid
// one is a user the broker declines to create -- checked here where the field can be
// named.
var accessLevels = map[string]bool{
	"none": true, "read-only": true, "mesh-manager": true, "read-write": true, "admin": true,
}

// accessLevelList is the enum for error messages, in the same order.
const accessLevelList = "'none', 'read-only', 'mesh-manager', 'read-write' or 'admin'"

// foldToEnvVar upper-cases a username and folds every character an environment
// variable name cannot carry to '_' -- the same mapping render's
// ContainerSecret.EnvVar applies to the whole secret name when docker sources a
// compose secret from the host environment. It exists here only to detect the
// collision that mapping can create; the rendering itself stays in render, and
// config must not import it (render depends on config). A small package-local copy
// in the spirit of identRE, which likewise mirrors a regexp two other packages own.
func foldToEnvVar(name string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(name) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}
	return b.String()
}

// checkCredentialChars rejects a control character (unicode.IsControl: this
// covers Unicode's control category, not just the ASCII set isCtrl above
// checks, since a secret can carry any encoding) in one secret value. An empty
// value is left alone -- emptiness is requireAll's or the caller's own job, not
// this check's. The message names the field and the byte offset of the
// offending character but never the value itself, or even the character: these
// are secrets, and this repo's rule is that reports say set/MISSING and never
// echo one (§3).
func checkCredentialChars(field, value string) error {
	if i := strings.IndexFunc(value, unicode.IsControl); i >= 0 {
		return fmt.Errorf("%s contains a control character at byte offset %d; remove it "+
			"(the value is a secret and is not shown)", field, i)
	}
	return nil
}

// validateCredentialChars checks every top-level secret that reaches a curl
// config fed on stdin (broker.sempCurl -> curlConfigLine, semp.go) or an
// equivalent line-oriented consumer: curlConfigLine escapes a quote and a
// backslash but not a newline, so a credential carrying one breaks out of its
// line before the request is even sent. Rejecting it here -- rather than
// escaping harder in curlConfigLine -- means the value never reaches that
// shape at all. admin.additionalUsers passwords get the same check inline in
// validateAdditionalUsers, where the per-user field name is already at hand.
func (c *Config) validateCredentialChars() error {
	for _, f := range []struct{ field, value string }{
		{"semp.adminPass", c.SEMP.AdminPass},
		{"semp.monitorPass", c.SEMP.MonitorPass},
		{"redundancy.psk", c.Redundancy.PSK},
		{"tls.certPassphrase", c.TLS.CertPassphrase},
	} {
		if err := checkCredentialChars(f.field, f.value); err != nil {
			return err
		}
	}
	return nil
}

// replTransports is the transport vocabulary, and replTransportList is how an error
// spells it. Three words: `encrypted` is the DMR name for ssl and is refused by name
// rather than quietly accepted, so an env file cannot carry a fourth spelling the
// renderer would have to decide about.
var replTransports = map[string]bool{
	TransportPlainText:  true,
	TransportCompressed: true,
	TransportSSL:        true,
}

const replTransportList = "plainText, compressed or ssl"

// validateReplication checks the DR block. The whole section is optional -- most
// deployments have no pair -- but a partially written one is an error rather than a
// half-configured switchover waiting to happen.
//
// What is NOT checked here is a site's `via`: it says how to reach a site when it is
// the MATE, and only the switch command ever needs it. Requiring it at load would
// refuse a file that configures replication perfectly well for the local-only command,
// so the switch checks both sites' blocks in its own preflight instead, before any
// mutation. An EMPTY `via:` is accepted for the same reason and not as a preference: a
// block with neither child decodes to the zero value, which is indistinguishable from an
// absent key, so there is nothing here to tell the two apart (validateReplVia).
func (c *Config) validateReplication() error {
	r := c.Replication
	if !r.Configured() {
		return nil
	}
	if len(r.Sites) != 2 {
		return fmt.Errorf("replication.sites must hold exactly 2 entries, one per site (got: %d). "+
			"Replication is a pair; omit the whole replication: section if this broker has no DR mate",
			len(r.Sites))
	}

	seenRouter := make(map[string]int, 4)
	seenVRN := make(map[string]int, 2)
	for i, s := range r.Sites {
		field := fmt.Sprintf("replication.sites[%d]", i)
		if strings.TrimSpace(s.VirtualRouterName) == "" {
			return fmt.Errorf("%s.virtualRouterName must be set: it is the site's key, referenced by "+
				"replication.vpns[].activeAt, and the operand the mate is given as "+
				"`replication mate virtual-router-name`", field)
		}
		if j, dup := seenVRN[s.VirtualRouterName]; dup {
			return fmt.Errorf("%s.virtualRouterName %q is the same as replication.sites[%d]'s: the two "+
				"sites must be distinguishable", field, s.VirtualRouterName, j)
		}
		seenVRN[s.VirtualRouterName] = i

		if len(s.RouterNames) == 0 {
			return fmt.Errorf("%s.routerNames must list at least one name: it is matched against the "+
				"broker's own `show router-name` so it can find itself in this file. List every node "+
				"of the site's HA group, since the backup node reports its own name", field)
		}
		for k, n := range s.RouterNames {
			if strings.TrimSpace(n) == "" {
				return fmt.Errorf("%s.routerNames[%d] must not be empty", field, k)
			}
			if j, dup := seenRouter[n]; dup {
				return fmt.Errorf("%s.routerNames[%d] %q is already declared by replication.sites[%d]: "+
					"a router belongs to one site, and an overlap would leave a broker unable to tell "+
					"which end of the pair it is", field, k, n, j)
			}
			seenRouter[n] = i
		}

		if err := validateReplEndpoints(field, s.Endpoints); err != nil {
			return err
		}
		if err := validateReplVia(field, s.Via); err != nil {
			return err
		}
	}

	seenVPN := make(map[string]bool, len(r.VPNs))
	for i, v := range r.VPNs {
		field := fmt.Sprintf("replication.vpns[%d]", i)
		if strings.TrimSpace(v.Name) == "" {
			return fmt.Errorf("%s.name must be set", field)
		}
		if seenVPN[v.Name] {
			return fmt.Errorf("%s.name %q is listed twice; each VPN appears once", field, v.Name)
		}
		seenVPN[v.Name] = true
		if _, ok := seenVRN[v.ActiveAt]; !ok {
			return fmt.Errorf("%s.activeAt %q names no site: it must be one of the declared "+
				"virtualRouterName values (%s), not a router name", field, v.ActiveAt,
				strings.Join(sortedKeys(seenVRN), ", "))
		}
	}
	return nil
}

// validateReplEndpoints checks one site's mate addresses. The two-per-transport ceiling
// is the software grammar's; an appliance takes only one, but which platform a site runs
// is not in the file and must not be -- the broker knows, so the renderer refuses an
// unrenderable set at apply time, when the CLI banner has said what the target is.
func validateReplEndpoints(field string, eps []ReplEndpoint) error {
	if len(eps) == 0 {
		return fmt.Errorf("%s.endpoints must list at least one address: it is how the other broker "+
			"dials this site, and a site with none is unreachable for replication", field)
	}
	perTransport := make(map[string]int, 3)
	// seen keys the NORMALISED triple, so `plain-text` written out and left blank count
	// as the same address rather than two. A duplicate is refused rather than quietly
	// de-duplicated, because it is a typo in a file that is byte-identical at both sites
	// and the operator should see it once instead of wondering later why a count differs.
	//
	// It matters more than a tidiness rule: `broker configure data-replication` decides
	// whether the mate needs converging by comparing the file's endpoint set against what
	// the broker reports, and the broker reports each address ONCE. A file listing one
	// address twice therefore never compares equal, so every run would decide the mate
	// differs, shut replication down on every VPN on the broker to rewrite identical mate
	// lines, and do it again on the next run -- an outage per invocation, forever.
	seen := make(map[string]int, len(eps))
	for k, e := range eps {
		ef := fmt.Sprintf("%s.endpoints[%d]", field, k)
		if strings.TrimSpace(e.Host) == "" {
			return fmt.Errorf("%s.host must be set", ef)
		}
		if e.Port < 1 || e.Port > 65535 {
			return fmt.Errorf("%s.port must be between 1 and 65535 (got: %d)", ef, e.Port)
		}
		t := e.Transport
		if t == "" {
			t = TransportPlainText // the CLI's own default for an unqualified entry
		}
		if !replTransports[t] {
			return fmt.Errorf("%s.transport must be %s (got: %q). `encrypted` is the routing name for "+
				"the same thing; use ssl", ef, replTransportList, e.Transport)
		}
		key := fmt.Sprintf("%s|%d|%s", strings.TrimSpace(e.Host), e.Port, t)
		if first, dup := seen[key]; dup {
			return fmt.Errorf("%s repeats the address already given as %s.endpoints[%d] "+
				"(%s:%d over %s). Each endpoint must be distinct: the broker reports an address "+
				"once, so a repeated one makes `broker configure data-replication` read the mate "+
				"as different on every run and shut replication down on every VPN to rewrite the "+
				"same lines", ef, field, first, e.Host, e.Port, t)
		}
		seen[key] = k

		perTransport[t]++
		if perTransport[t] > 2 {
			return fmt.Errorf("%s: replication accepts at most 2 addresses per transport, and %q has "+
				"%d", field, t, perTransport[t])
		}
	}
	return nil
}

// validateReplVia checks a site's access block when it has one. Exactly one mechanism,
// because the key present IS the choice -- there is no separate name to disagree with it.
func validateReplVia(field string, v ReplVia) error {
	switch {
	case v.Kubernetes != nil && v.SEMP != nil:
		return fmt.Errorf("%s.via declares both kubernetes and semp: name one, since the key that is "+
			"present is the mechanism", field)
	case v.Kubernetes != nil:
		k := v.Kubernetes
		if len(k.Command) == 0 {
			return fmt.Errorf("%s.via.kubernetes.command must be set: it is the cluster CLI this tool "+
				"runs to reach that site (carry the cluster in it, e.g. `kubectl --context dr`)", field)
		}
		if strings.TrimSpace(k.Namespace) == "" {
			return fmt.Errorf("%s.via.kubernetes.namespace must be set", field)
		}
		if strings.TrimSpace(k.Name) == "" {
			return fmt.Errorf("%s.via.kubernetes.name must be set: the mate's PubSubPlusEventBroker "+
				"name, which its pod is named after", field)
		}
	case v.SEMP != nil:
		return validateReplViaSEMP(field, v.SEMP)
	}
	// A `via:` with neither child decodes to the zero value, indistinguishable from an
	// absent key -- so an empty block is accepted here and the switch command's own
	// preflight is what reports a site it cannot reach.
	return nil
}

// validateReplViaSEMP checks the SEMP leg, including the credential. There is no stdin
// prompt, so a missing source is a config error on every path rather than a question
// nobody can answer under --no-prompt.
func validateReplViaSEMP(field string, s *ReplViaSEMP) error {
	if strings.TrimSpace(s.Host) == "" {
		return fmt.Errorf("%s.via.semp.host must be set, and it is not derived from endpoints: "+
			"replication runs over the message backbone, so a reachable replication endpoint proves "+
			"nothing about SEMP reachability", field)
	}
	if s.Port < 1 || s.Port > 65535 {
		return fmt.Errorf("%s.via.semp.port must be between 1 and 65535 (got: %d)", field, s.Port)
	}
	set := 0
	if s.Pass != "" {
		set++
	}
	if s.PassEnv != "" {
		set++
	}
	if s.PassSecret != nil {
		set++
	}
	if set == 0 {
		return fmt.Errorf("%s.via.semp needs the mate's admin password: set pass, or point passEnv at "+
			"an environment variable, or name a Kubernetes Secret in passSecret. There is no prompt for "+
			"it", field)
	}
	if set > 1 {
		return fmt.Errorf("%s.via.semp declares more than one of pass, passEnv and passSecret: use one, "+
			"so which value reaches the mate is never a precedence question", field)
	}
	if s.PassSecret != nil {
		for _, f := range []struct{ key, value string }{
			{"namespace", s.PassSecret.Namespace},
			{"name", s.PassSecret.Name},
			{"key", s.PassSecret.Key},
		} {
			if strings.TrimSpace(f.value) == "" {
				return fmt.Errorf("%s.via.semp.passSecret.%s must be set", field, f.key)
			}
		}
	}
	// Same reasoning as validateCredentialChars: this value reaches a curl config on
	// stdin, where a newline breaks out of its line before the request is sent.
	return checkCredentialChars(field+".via.semp.pass", s.Pass)
}

// sortedKeys renders a set as a stable, readable list for an error message.
func sortedKeys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// validateAdditionalUsers checks the extra CLI users in semp.additionalUsers, which every
// platform carries but delivers differently: containers mount each password as a file,
// while Kubernetes puts both password and access level in the pod environment from a
// dedicated Secret (<kubernetes.name>-additional-users, projected via
// spec.extraEnvVarsSecret) -- so on both platforms the user exists from the broker's first
// boot, with nothing to run afterwards. Access level is required rather than defaulted --
// silently choosing someone's permissions is not a default worth having. p selects the
// platform-specific rules; only the k8s path constrains the USERNAME, since the kubelet's
// envFrom projection silently drops any Secret key that is not a valid environment
// variable name.
func (c *Config) validateAdditionalUsers(p Platform) error {
	seen := make(map[string]bool, len(c.SEMP.AdditionalUsers))
	folded := make(map[string]string, len(c.SEMP.AdditionalUsers))
	for i, u := range c.SEMP.AdditionalUsers {
		field := fmt.Sprintf("semp.additionalUsers[%d]", i)
		if strings.TrimSpace(u.Username) == "" {
			return fmt.Errorf("%s.username must be set", field)
		}
		if !brokerUsernameRE.MatchString(u.Username) {
			return fmt.Errorf("%s.username %q is invalid: it must start with a letter or '_', then letters, "+
				"digits, '.', '_' or '-', and be 1-32 characters. That is the broker's own rule, checked "+
				"here because the broker would otherwise reject it at CREATE time, on a running "+
				"deployment. The name also becomes the setting username_%s_password",
				field, u.Username, u.Username)
		}
		if u.Username == AdminUser || u.Username == MonitorUser {
			return fmt.Errorf("%s.username %q is a built-in user: admin has admin.pass and monitor has "+
				"semp.monitorPass -- additionalUsers is for users beyond those", field, u.Username)
		}
		if seen[u.Username] {
			return fmt.Errorf("%s.username %q is listed twice; each user appears once", field, u.Username)
		}
		seen[u.Username] = true
		// Two users differing only in separator style are distinct to the broker but
		// fold to ONE docker host variable name (render's ContainerSecret.EnvVar maps
		// every non-alphanumeric to '_'), which would feed one user's password to
		// both. Caught here, where both offending fields can be named, rather than
		// silently at deploy time.
		key := foldToEnvVar(u.Username)
		if other := folded[key]; other != "" {
			return fmt.Errorf("%s.username %q collides with %q: they differ only in '.', '_' or '-', "+
				"which become the same host environment variable (...%s...) for docker's compose secrets -- "+
				"rename one", field, u.Username, other, key)
		}
		folded[key] = u.Username
		if !accessLevels[u.AccessLevel] {
			return fmt.Errorf("%s.accessLevel must be %s (got: %q)", field, accessLevelList, u.AccessLevel)
		}
		if u.Password == "" {
			return fmt.Errorf("%s.password must not be empty; set it, or point %s.passwordEnv at an "+
				"environment variable holding it", field, field)
		}
		if err := checkCredentialChars(field+".password", u.Password); err != nil {
			return err
		}
		// Kubernetes delivers these users through a Secret that the CR names in
		// spec.extraEnvVarsSecret, so every key in it becomes an environment variable in
		// the broker container. The kubelet SKIPS keys that are not valid environment
		// variable names -- silently, bar one pod event -- so a username carrying '.' or
		// '-' would produce a user with no password, or no user at all, with nothing in
		// the deploy to say so. The Secret key itself would be legal; it is the envFrom
		// projection that drops it.
		//
		// Containers are unaffected: they mount the password as a FILE and only the
		// access level rides the environment, so the wider identRE above still holds
		// there. This is the one rule that differs by platform, which is why it is here
		// rather than in the shared loop above.
		if p == K8s && !envVarNameRE.MatchString(u.Username) {
			return fmt.Errorf("%s.username %q cannot be used on Kubernetes: it becomes the environment "+
				"variables username_%s_password and username_%s_globalaccesslevel, and Kubernetes "+
				"silently DROPS environment variables whose names are not letters, digits and "+
				"underscores (starting with a letter or underscore) -- the user would be created "+
				"without a password, or not at all. Rename it, or drop the '.'/'-'", field,
				u.Username, u.Username, u.Username)
		}
	}
	return nil
}

// nodeMatchOperators are the node-label match operators Kubernetes accepts. The
// value reaches the manifest unquoted as a bare enum, so it is checked here.
var nodeMatchOperators = map[string]bool{
	"In": true, "NotIn": true, "Exists": true, "DoesNotExist": true, "Gt": true, "Lt": true,
}

// validatePlacementAffinity checks the additive affinity blocks. It deliberately
// does not police weight bounds, matching the existing laxness on
// antiAffinityWeight -- what it catches is a value the API server would reject
// with a far less obvious message, or a term with no topology to spread over.
func validatePlacementAffinity(pl Placement) error {
	for i, term := range pl.NodeAffinity.Preferred {
		if err := validateMatchExprs(fmt.Sprintf("kubernetes.placement.nodeAffinity.preferred[%d].match", i), term.Match); err != nil {
			return err
		}
	}
	if err := validateMatchExprs("kubernetes.placement.nodeAffinity.required", pl.NodeAffinity.Required); err != nil {
		return err
	}
	for _, group := range []struct {
		field string
		terms []PodAffinityTerm
	}{
		{"kubernetes.placement.podAffinity", pl.PodAffinity},
		{"kubernetes.placement.podAntiAffinity", pl.PodAntiAffinity},
	} {
		for i, term := range group.terms {
			if strings.TrimSpace(term.TopologyKey) == "" {
				return fmt.Errorf("%s[%d].topologyKey must be set (e.g. kubernetes.io/hostname)", group.field, i)
			}
		}
	}
	return nil
}

func validateMatchExprs(field string, exprs []NodeMatchExpr) error {
	for i, e := range exprs {
		if strings.TrimSpace(e.Key) == "" {
			return fmt.Errorf("%s[%d].key must be set", field, i)
		}
		if !nodeMatchOperators[e.Operator] {
			return fmt.Errorf("%s[%d].operator %q is invalid: expected In, NotIn, Exists, DoesNotExist, Gt or Lt",
				field, i, e.Operator)
		}
		needsValues := e.Operator == "In" || e.Operator == "NotIn" || e.Operator == "Gt" || e.Operator == "Lt"
		if needsValues && len(e.Values) == 0 {
			return fmt.Errorf("%s[%d].values must not be empty for operator %s", field, i, e.Operator)
		}
	}
	return nil
}

// pskValue is the pre-shared key however it was supplied -- literally, or through the
// environment variable pskEnv names. Load resolves pskEnv into PSK before validation, so
// this reads PSK first and falls back to the variable's NAME only to answer "was one
// configured at all", which is what the container requirement actually asks.
func (c *Config) pskValue() string {
	if c.Redundancy.PSK != "" {
		return c.Redundancy.PSK
	}
	return c.Redundancy.PSKEnv
}

// validateContainerPSK enforces the one credential this tool will not invent.
//
// The asymmetry with Kubernetes is the operator's rather than a preference: there the Solace
// operator generates a key and distributes it to the pods it owns, so an empty one is a
// legitimate deployment. Nothing does that for three container hosts -- each reads its own
// env file -- so a key one host invented is a key the other two never see, and the group
// fails to form with no obvious cause.
//
// This tool used to generate one on the first HA deploy and write it back into the env file.
// That is gone: it only ever ran on one host, the value still had to be copied to the other
// two by hand, and a deploy that edits the file it was handed is a surprise on a file that
// may be version-controlled or templated. So the key is the operator's to make, and the
// message has to hand them the command rather than just naming the field.
func (c *Config) validateContainerPSK(p Platform) error {
	if strings.TrimSpace(c.pskValue()) != "" {
		return nil
	}
	return fmt.Errorf("redundancy.psk must be set when redundancy.enabled is true on %s.\n"+
		"  Generate one:  openssl rand -base64 32\n"+
		"  Then put the SAME value in the env file on all three hosts -- a group whose members "+
		"hold different keys never forms.\n"+
		"  (On Kubernetes this key is optional: the Solace operator generates and distributes "+
		"its own when it is left empty.)", platformKey(p))
}

func (c *Config) validateContainer(p Platform) error {
	req := map[string]string{
		"image.repo":     c.Image.Repo,
		"image.tag":      c.Image.Tag,
		"semp.adminPass": c.SEMP.AdminPass,
	}
	// redundancy.primary.name is NOT here. In standalone it is optional: there is one
	// node, it is always this host, and FillStandaloneNodeName supplies the host's own
	// hostname as the routername when the env file leaves it out. In HA it is mandatory,
	// below, with the other two.
	//
	// The backup/monitor rows + the addresses are required only for the HA group.
	if c.RedundancyEnabled() {
		// The pre-shared key gets its own error rather than joining req below, because the
		// generic "these fields must not be empty" tells an operator nothing about what to
		// do, and this is the one field where nothing else will do it for them.
		if err := c.validateContainerPSK(p); err != nil {
			return err
		}
		// Every name is required in HA, and cannot be defaulted the way the standalone
		// one is: each is the KEY of that node's entry in the group table EVERY host
		// renders (redundancy_group_node_<name>_connectvia). A host knows its own
		// hostname and no other machine's, so a guess here would build a table the other
		// two hosts do not agree with, and the group would never form.
		req["redundancy.primary.name"] = c.Redundancy.Primary.Name
		req["redundancy.primary.addr"] = c.Redundancy.Primary.Addr
		req["redundancy.backup.name"] = c.Redundancy.Backup.Name
		req["redundancy.backup.addr"] = c.Redundancy.Backup.Addr
		req["redundancy.monitor.name"] = c.Redundancy.Monitor.Name
		req["redundancy.monitor.addr"] = c.Redundancy.Monitor.Addr
	}
	// Data dir lives in the platform's container block.
	dataKey := "docker.container.dataDir"
	if p == Podman {
		dataKey = "podman.container.dataDir"
	}
	req[dataKey] = c.ContainerBlock(p).DataDir
	if p == Podman {
		// Mandatory, not defaulted: this directory receives the server-certificate
		// bundle, which contains a PRIVATE KEY, and where a private key lands on the
		// host is the operator's call rather than a path this tool picks for them.
		// A quadlet unit cannot inline file content, which is why podman needs such a
		// directory at all and docker does not.
		req["podman.baseDir"] = c.Podman.BaseDir
	}

	if missing := requireAll(req); len(missing) > 0 {
		return missingErr(missing)
	}

	if p == Podman && !IsAbsHostPath(c.Podman.BaseDir) {
		return fmt.Errorf("podman.baseDir %q must be an absolute path: it is the source of a quadlet `Volume=` "+
			"line, and podman reads a relative source as the name of a NAMED VOLUME -- so a relative value would "+
			"mount an empty volume over the server certificate with no error at all (e.g. /opt/solace)",
			c.Podman.BaseDir)
	}

	// The data dir must be ABSOLUTE, and it is the one host path that is required to
	// be rather than being resolved against the env file's directory like the rest
	// (hostpath.go, rebaseHostPaths).
	//
	// Two reasons, and the second is why it is a refusal instead of a rebase. It is
	// the host side of a bind mount, so a bare relative value is read by podman as
	// the name of a NAMED VOLUME and would mount an empty volume over the broker's
	// data with no error. And it is the target of `rm -rf` on
	// `remove broker --delete-data` (container.Manager.purgeData), so silently
	// changing which directory a recursive delete points at is not a fix -- the
	// operator has to be the one who says where their data lives.
	//
	// The default is /opt/solace/data, so nothing that works today is refused.
	// A certificate with no key cannot produce the bundle a container mounts, and
	// this is a deliberate TIGHTENING: an env file setting tls.cert alone loads
	// today. It should not, and the reason is worth stating -- until now the
	// container renderers mounted tls.cert ALONE, never reading tls.certKey, so a
	// broker-owned TLS listener could only have worked if that one file already
	// carried the key. Failing at load names the missing field; the old behaviour
	// produced a broker whose TLS silently did not work.
	//
	// Both directions, not just cert-without-key. Every container site keys on TLS.Cert
	// alone -- the compose secret, the quadlet Volume=, the configsync/matelink pairs --
	// so a lone certKey produces no error, no warning and no TLS: exactly the silent
	// failure this check exists to end, arrived at from the other side.
	if (c.TLS.Cert == "") != (c.TLS.CertKey == "") {
		return fmt.Errorf("tls.cert and tls.certKey must be set together (got cert=%q key=%q): on %s the "+
			"server certificate is delivered as ONE file containing the private key followed by the "+
			"certificate, so both halves are needed (set the missing one, or unset both to deploy "+
			"without broker TLS)", c.TLS.Cert, c.TLS.CertKey, platformKey(p))
	}

	if dir := c.ContainerBlock(p).DataDir; !IsAbsHostPath(dir) {
		return fmt.Errorf("%s %q must be an absolute path: it is the host side of a bind mount, and it is what "+
			"`remove broker --delete-data` deletes recursively -- a relative value would be read by podman as a "+
			"named volume and would leave a recursive delete pointing at whatever directory the command was run "+
			"from (e.g. /opt/solace/data)", dataKey, dir)
	}

	// These reach the compose/quadlet artifact in structural positions, so they are
	// format-checked here rather than being allowed to produce a broken artifact.
	//
	// admin.user is in the list because on a container platform it may be any string
	// (only kubernetes pins it to "admin"), and it is not merely a value there: it is
	// interpolated into the derived broker setting `username_<user>_globalaccesslevel`
	// and into the secret key `username_<user>_password`, which is simultaneously a
	// compose secret target and a podman `target=`. admin.additionalUsers usernames
	// have always been held to this grammar for exactly that reason; the built-in
	// admin reaches the same positions and was not.
	cb := c.ContainerBlock(p)
	for _, f := range []struct{ field, value string }{
		{platformKey(p) + ".container.name", cb.Name},
		{"redundancy.primary.name", c.Redundancy.Primary.Name},
		{"redundancy.backup.name", c.Redundancy.Backup.Name},
		{"redundancy.monitor.name", c.Redundancy.Monitor.Name},
	} {
		if err := validIdent(f.field, f.value); err != nil {
			return err
		}
	}
	// container.name additionally has to be a name the ENGINES accept, which is
	// narrower than identIdent's charset in one way that matters: the first character
	// must be alphanumeric. `-broker` passes the charset, and both docker and podman
	// refuse it -- but not before this tool has written it into a compose file and a
	// quadlet unit, and not before it has appeared in argument positions where a
	// leading dash reads as a flag rather than a name.
	//
	// A bound too, because nothing else imposes one and the derived host-side secret
	// names are longer still.
	if n := cb.Name; n != "" {
		if !engineContainerNameRE.MatchString(n) {
			return fmt.Errorf("%s.container.name %q is invalid: it must start with a letter or digit, then "+
				"letters, digits, '.', '_' or '-'. Both engines refuse a name starting with '-' or '.', and a "+
				"leading dash also reads as a flag wherever the name reaches a command line",
				platformKey(p), n)
		}
		if len(n) > maxContainerNameLen {
			return fmt.Errorf("%s.container.name %q is %d characters, at most %d: it is the stem of every "+
				"derived name, including the host-side secret names and the podman unit, which are longer still",
				platformKey(p), n, len(n), maxContainerNameLen)
		}
	}
	if u := cb.RunUser; u != "" && !runUserRE.MatchString(u) {
		return fmt.Errorf("%s.container.runUser %q is invalid: expected uid[:gid] using only letters, digits, '.', '_' and '-'",
			platformKey(p), u)
	}
	if m := cb.Mem; m != "" && !containerMemRE.MatchString(m) {
		// The likely mistake is copying kubernetes.msgNode.mem's Kubernetes quantity
		// across; the engines reject "Mi"/"Gi", and catching it here beats a
		// compose parse error at deploy time.
		return fmt.Errorf("%s.container.mem %q is invalid: docker and podman take an integer followed by "+
			"b, k, m or g (e.g. 6898m), not the Mi/Gi suffix kubernetes.msgNode.mem uses", platformKey(p), m)
	}
	if err := c.validateAdditionalUsers(p); err != nil {
		return err
	}

	if hc := cb.HealthCheck; hc.Enabled {
		if len(hc.Cmd) > 0 {
			// An explicit probe is the operator's own; it only gets the exec-boundary
			// check, not the version gate -- and not the execution guard, since it
			// runs inside the container rather than here (validateProbeCommand).
			if err := validateProbeCommand(platformKey(p)+".container.healthCheck.cmd", Command(hc.Cmd)); err != nil {
				return err
			}
		} else if err := c.checkHealthCheckVersion(p); err != nil {
			return err
		}
	}

	net := c.NetworkBlock(p)
	switch net.Mode {
	case "host":
	case "bridge":
		if len(net.Ports) == 0 {
			// ApplyDefaults fills this list, so reaching here means Validate ran on
			// its own (a hand-built config) -- say so rather than implying the user
			// must always list ports.
			return fmt.Errorf("%s.network.mode=bridge requires a non-empty ports list (host:container entries); "+
				"config.Load fills the default set automatically", platformKey(p))
		}
	default:
		return fmt.Errorf("%s.network.mode must be 'host' or 'bridge' (got: %q)", platformKey(p), net.Mode)
	}

	return nil
}

// checkHealthCheckVersion gates the built-in readiness probe on the broker
// release that first serves /health-check/readiness. An older broker has no such
// endpoint, so the probe would fail forever and report a healthy broker as
// unhealthy -- under podman's auto-restart, a restart loop. An unidentifiable tag
// is refused for the same reason: this cannot be verified, and guessing "new
// enough" is the dangerous direction. Both errors name the explicit-cmd escape
// hatch, which is the supported way to probe an older or custom-tagged broker.
func (c *Config) checkHealthCheckVersion(p Platform) error {
	const field = ".container.healthCheck"
	ok, known := c.Image.AtLeast(HealthCheckMinMajor, HealthCheckMinMinor)
	switch {
	case !known:
		return fmt.Errorf("%s%s is enabled with no cmd, which uses the built-in readiness probe, "+
			"but the broker version cannot be read from image.tag %q; the probe needs %d.%d or later -- "+
			"use a version-numbered tag, or set %s%s.cmd to your own probe",
			platformKey(p), field, c.Image.Tag, HealthCheckMinMajor, HealthCheckMinMinor, platformKey(p), field)
	case !ok:
		return fmt.Errorf("%s%s is enabled with no cmd, which uses the built-in readiness probe, "+
			"but image.tag %q is older than %d.%d, where /health-check/readiness was introduced -- "+
			"upgrade the broker, or set %s%s.cmd to a probe that image supports",
			platformKey(p), field, c.Image.Tag, HealthCheckMinMajor, HealthCheckMinMinor, platformKey(p), field)
	}
	return nil
}

func platformKey(p Platform) string {
	if p == Podman {
		return "podman"
	}
	return "docker"
}

func requireAll(fields map[string]string) []string {
	var missing []string
	for name, val := range fields {
		if strings.TrimSpace(val) == "" {
			missing = append(missing, name)
		}
	}
	return missing
}

func missingErr(missing []string) error {
	sort.Strings(missing)
	return fmt.Errorf("these fields must not be empty: %s", strings.Join(missing, ", "))
}

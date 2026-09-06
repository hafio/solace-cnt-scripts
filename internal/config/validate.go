package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Validate checks mandatory and enum fields for the given platform, mirroring
// the mandatory-vars and enum checks in the two bash bootstraps. It fails loud
// with an actionable message listing every offending field at once.
func (c *Config) Validate(p Platform) error {
	// Redundancy is a shared enum on every platform.
	switch c.Redundancy {
	case "yes", "no":
	default:
		return fmt.Errorf("redundancy must be 'yes' or 'no' (got: %q)", c.Redundancy)
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
	// stdin, nodes.psk and tls.certPassphrase travel the same way through the
	// container config path. Checked once here rather than per platform, since
	// the fields exist regardless of which platform ends up reading them.
	if err := c.validateCredentialChars(); err != nil {
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
	missing := requireAll(map[string]string{
		"kubernetes.name":            c.K8s.Name,
		"kubernetes.namespace":       c.K8s.Namespace,
		"image.repo":                 c.Image.Repo,
		"image.tag":                  c.Image.Tag,
		"kubernetes.storage.msgNode": c.K8s.Storage.MsgNode,
		"admin.pass":                 c.Admin.Pass, // hardening: no hardcoded default password
	})
	if len(missing) > 0 {
		return missingErr(missing)
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
	if c.K8s.MsgNode.CPU != "" {
		// Removed rather than ignored: a stale cpu: in an env file is a sizing
		// decision the operator believes is in effect, so it has to be seen.
		return fmt.Errorf("kubernetes.msgNode.cpu was removed; broker CPU is fixed by the scaling tier and "+
			"derived from scaling.maxConnections (one of %s) -- drop the key. "+
			"kubernetes.msgNode.mem is unaffected: it still overrides the tier's default memory", scalingTierList)
	}
	if u := c.Admin.User; u != "" && u != "admin" {
		// Rejected rather than ignored, for the same reason as msgNode.cpu above: this is
		// the login an operator believes is in effect. The operator reads the fixed
		// username_admin_password key out of the credentials Secret (k8s/secrets.go,
		// verified against a live cluster), and creates the admin user itself, so nothing
		// on this platform can honour another name. An unset value is skipped: ApplyDefaults
		// fills "admin", so empty means "will be defaulted" as it does for every other
		// setDefault field.
		return fmt.Errorf("admin.user %q is not supported on Kubernetes: the operator reads the fixed "+
			"username_admin_password key out of kubernetes.adminSecret, so the broker admin user is always "+
			"'admin' -- drop the key (it applies to docker and podman, where the username is yours to choose)", u)
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

// cliForbiddenPassword are the characters the broker's own CLI rejects in a
// `create username ... password ...` value. They are refused here rather than at
// config time on the cluster, so an env file that cannot be applied fails at load.
// Only k8s delivers a password through the CLI: on containers it is written to a
// mounted file, which has no such restriction, so this is checked per platform.
const cliForbiddenPassword = ":()\";'<>,`\\*&|"

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
		{"admin.pass", c.Admin.Pass},
		{"admin.monitorPass", c.Admin.MonitorPass},
		{"nodes.psk", c.Nodes.PSK},
		{"tls.certPassphrase", c.TLS.CertPassphrase},
	} {
		if err := checkCredentialChars(f.field, f.value); err != nil {
			return err
		}
	}
	return nil
}

// validateAdditionalUsers checks the extra CLI users, which every platform carries
// but delivers differently: containers create them at boot from a mounted secret
// file plus an access-level setting, while k8s creates them post-deployment through
// the broker CLI (`config additional-users`). Access level is required rather than
// defaulted -- silently choosing someone's permissions is not a default worth
// having. p selects the platform-specific rules; only the k8s path constrains the
// password, since only it puts the value on a CLI line.
func (c *Config) validateAdditionalUsers(p Platform) error {
	seen := make(map[string]bool, len(c.Admin.AdditionalUsers))
	folded := make(map[string]string, len(c.Admin.AdditionalUsers))
	for i, u := range c.Admin.AdditionalUsers {
		field := fmt.Sprintf("admin.additionalUsers[%d]", i)
		if strings.TrimSpace(u.Username) == "" {
			return fmt.Errorf("%s.username must be set", field)
		}
		if !identRE.MatchString(u.Username) {
			return fmt.Errorf("%s.username %q is invalid: only letters, digits, '.', '_' and '-' are allowed "+
				"(it becomes the secret name username_%s_password)", field, u.Username, u.Username)
		}
		if u.Username == "admin" || u.Username == "monitor" || u.Username == c.Admin.User {
			return fmt.Errorf("%s.username %q is a built-in user: admin has admin.pass and monitor has "+
				"admin.monitorPass -- additionalUsers is for users beyond those", field, u.Username)
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
		// k8s creates the user with `create username "<u>" password "<p>"`, and the
		// broker CLI rejects these characters in the value. The message names the
		// offending character but never the password (§3).
		if p == K8s {
			if i := strings.IndexAny(u.Password, cliForbiddenPassword); i >= 0 {
				return fmt.Errorf("%s.password contains %q, which the broker CLI rejects in a password; "+
					"on Kubernetes the user is created over the CLI, so none of %s may appear "+
					"(the value itself is not shown)", field, string(u.Password[i]), cliForbiddenPassword)
			}
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

func (c *Config) validateContainer(p Platform) error {
	req := map[string]string{
		"image.repo":         c.Image.Repo,
		"image.tag":          c.Image.Tag,
		"admin.pass":         c.Admin.Pass,
		"nodes.primary.name": c.Nodes.Primary.Name,
	}
	// The backup/monitor rows + primary IP are required only for the HA group.
	if c.RedundancyEnabled() {
		req["nodes.primary.ip"] = c.Nodes.Primary.IP
		req["nodes.backup.name"] = c.Nodes.Backup.Name
		req["nodes.backup.ip"] = c.Nodes.Backup.IP
		req["nodes.monitor.name"] = c.Nodes.Monitor.Name
		req["nodes.monitor.ip"] = c.Nodes.Monitor.IP
	}
	// Data dir lives in the platform's container block.
	dataKey := "docker.container.dataDir"
	if p == Podman {
		dataKey = "podman.container.dataDir"
	}
	req[dataKey] = c.ContainerBlock(p).DataDir

	if missing := requireAll(req); len(missing) > 0 {
		return missingErr(missing)
	}

	// These reach the compose/quadlet artifact in structural positions, so they are
	// format-checked here rather than being allowed to produce a broken artifact.
	cb := c.ContainerBlock(p)
	for _, f := range []struct{ field, value string }{
		{platformKey(p) + ".container.name", cb.Name},
		{"nodes.primary.name", c.Nodes.Primary.Name},
		{"nodes.backup.name", c.Nodes.Backup.Name},
		{"nodes.monitor.name", c.Nodes.Monitor.Name},
	} {
		if err := validIdent(f.field, f.value); err != nil {
			return err
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
	sortStrings(missing)
	return fmt.Errorf("these fields must not be empty: %s", strings.Join(missing, ", "))
}

// sortStrings is a tiny insertion sort to keep the missing-fields message stable
// without pulling in the sort package for a handful of items.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}

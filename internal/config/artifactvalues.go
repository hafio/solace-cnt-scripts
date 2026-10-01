package config

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// The values in this file reach a rendered artifact as they were written: a quadlet
// unit line (`Image=`, `PublishPort=`, `HealthInterval=`), a compose scalar
// (`image:`, `interval:`), or the broker CR. None of those sinks can carry a newline
// -- in a unit file it starts a new key or a new section, so `[Service]` plus
// `ExecStartPre=` becomes a command systemd runs as root; in YAML it starts a new
// key, or with `---` a new document `broker deploy` applies. So each value is held to
// the grammar its consumer actually accepts, which excludes every character that
// could restructure the artifact, rather than to a denylist of the ones known today.

// The OCI distribution reference grammar (github.com/distribution/reference), in
// the parts this schema splits a reference into.
const (
	refDomainComponent = `(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])`
	refDomain          = refDomainComponent + `(?:\.` + refDomainComponent + `)*(?::[0-9]+)?`
	refPathComponent   = `[a-z0-9]+(?:(?:[._]|__|-+)[a-z0-9]+)*`
)

// imageRegistryRE is a registry host with an optional port and path prefix
// (ghcr.io/solace), since Ref joins it to the repository with a '/'.
var imageRegistryRE = regexp.MustCompile(`^` + refDomain + `(?:/` + refPathComponent + `)*$`)

// imageRepoRE is a repository path, optionally led by its own registry host when
// image.registry is unset (localhost:5000/solace-pubsub-standard).
var imageRepoRE = regexp.MustCompile(`^(?:` + refDomain + `/)?` + refPathComponent + `(?:/` + refPathComponent + `)*$`)

// imageTagRE is the reference grammar's tag. A digest is not a tag, and Ref always
// joins with ':', so a digest cannot be expressed here at all.
var imageTagRE = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9._-]{0,127}$`)

// validateImage holds the three image fields to the reference grammar. Empty values
// are skipped: image.repo and image.tag are required per platform, image.registry is
// optional.
func (c *Config) validateImage() error {
	for _, f := range []struct {
		field, value, shape string
		re                  *regexp.Regexp
	}{
		{"image.registry", c.Image.Registry, "a registry host with an optional :port and path (e.g. " +
			"registry.example.com:5000 or ghcr.io/solace)", imageRegistryRE},
		{"image.repo", c.Image.Repo, "a lowercase repository path, optionally led by a registry host (e.g. " +
			"solace/solace-pubsub-standard)", imageRepoRE},
		{"image.tag", c.Image.Tag, "1-128 letters, digits, '_', '.' or '-', not starting with '.' or '-' " +
			"(e.g. 10.26.1.5); a digest is not supported", imageTagRE},
	} {
		if f.value != "" && !f.re.MatchString(f.value) {
			return fmt.Errorf("%s %q is not a valid image reference part: it must be %s. It is written into "+
				"the quadlet unit, the compose file and the broker CR as-is, where a newline or a space would "+
				"restructure them", f.field, f.value, f.shape)
		}
	}
	return nil
}

// portOrRange parses one side of a port mapping: a port, or lo-hi. It returns the
// number of ports it names.
func portOrRange(s string) (int, bool) {
	lo, hi, isRange := strings.Cut(s, "-")
	a, errA := strconv.Atoi(lo)
	if errA != nil || a < 1 || a > 65535 || lo != strconv.Itoa(a) {
		return 0, false
	}
	if !isRange {
		return 1, true
	}
	b, errB := strconv.Atoi(hi)
	if errB != nil || b < a || b > 65535 || hi != strconv.Itoa(b) {
		return 0, false
	}
	return b - a + 1, true
}

// validContainerPort checks one <platform>.network.ports entry against the form
// both engines' publish syntax shares: [ip:]host:container[/tcp|/udp], each side a
// port or a lo-hi range, and an IPv6 address in brackets. A container range needs a
// host range of the same size; a host range with one container port lets the engine
// pick a free host port from it. The engines accept more (an empty host port, a
// bare container port) -- this schema has always said host:container, so those
// stay out.
func validContainerPort(field, entry string) error {
	bad := func(why string) error {
		return fmt.Errorf("%s = %q is invalid: %s. Write it as host:container (8080:8080), optionally "+
			"bound to one address (127.0.0.1:8080:8080 or [::1]:8080:8080) and suffixed /tcp or /udp; either "+
			"side may be a range (55000-55010:55000-55010)", field, entry, why)
	}
	rest := entry
	if i := strings.LastIndex(rest, "/"); i >= 0 {
		switch rest[i+1:] {
		case "tcp", "udp":
		default:
			return bad(fmt.Sprintf("protocol %q must be tcp or udp", rest[i+1:]))
		}
		rest = rest[:i]
	}
	var ip string
	if strings.HasPrefix(rest, "[") {
		end := strings.Index(rest, "]:")
		if end < 0 {
			return bad("a bracketed IPv6 address must be followed by :host:container")
		}
		ip, rest = rest[1:end], rest[end+2:]
		if parsed := net.ParseIP(ip); parsed == nil || parsed.To4() != nil {
			return bad(fmt.Sprintf("%q is not an IPv6 address", ip))
		}
	}
	parts := strings.Split(rest, ":")
	switch {
	case len(parts) == 3 && ip == "":
		ip, parts = parts[0], parts[1:]
		if parsed := net.ParseIP(ip); parsed == nil || parsed.To4() == nil {
			return bad(fmt.Sprintf("%q is not an IPv4 address (an IPv6 one goes in brackets)", ip))
		}
	case len(parts) != 2:
		return bad("it needs exactly a host port and a container port")
	}
	hostN, ok := portOrRange(parts[0])
	if !ok {
		return bad(fmt.Sprintf("host side %q must be a port 1-65535 or a lo-hi range", parts[0]))
	}
	ctrN, ok := portOrRange(parts[1])
	if !ok {
		return bad(fmt.Sprintf("container side %q must be a port 1-65535 or a lo-hi range", parts[1]))
	}
	if ctrN > 1 && hostN != ctrN {
		return bad(fmt.Sprintf("a container range of %d ports needs a host range of the same size, not %d",
			ctrN, hostN))
	}
	return nil
}

// validHealthDuration checks one health-check timing: a Go duration, which is what
// both compose and quadlet parse, and never negative. Zero is allowed only where
// the engines take it to mean "none" (startPeriod); an interval or timeout of zero
// would never probe, or never wait for one.
func validHealthDuration(field, value string, zeroOK bool) error {
	if value == "" {
		return nil
	}
	d, err := time.ParseDuration(value)
	if err != nil || d < 0 || (d == 0 && !zeroOK) {
		floor := "greater than zero"
		if zeroOK {
			floor = "zero or more"
		}
		return fmt.Errorf("%s %q is invalid: it must be a duration %s, a number with a unit like 5s, 1m or "+
			"1m30s. It is written into the quadlet unit and the compose file as-is", field, value, floor)
	}
	return nil
}

// coreLimitRE is ulimits.core: -1 for unlimited, or a byte count as a plain decimal
// whole number with no sign, unit or leading zero.
var coreLimitRE = regexp.MustCompile(`^(?:-1|0|[1-9][0-9]*)$`)

// validCoreLimit checks ulimits.core against the one spelling docker, podman and
// systemd's LimitCORE= all take -- the -1 is rewritten to infinity for systemd
// (Container.SystemdCoreLimit) -- and that fits the int64 each of them parses it
// into. Anything else -- `unlimited`, a unit suffix, a newline carrying a key of
// its own -- is refused here rather than by the engine at deploy.
func validCoreLimit(field, value string) error {
	if coreLimitRE.MatchString(value) {
		if _, err := strconv.ParseInt(value, 10, 64); err == nil {
			return nil
		}
	}
	return fmt.Errorf("%s %q is invalid: it must be -1 for unlimited core dumps (the default, and Solace's "+
		"recommendation) or a size in bytes as a plain whole number, such as 0 to write none. It is written "+
		"into the quadlet unit and the compose file as-is", field, value)
}

// validateContainerArtifactValues checks the container block's values that the
// quadlet and compose renderers write as-is: every network.ports entry (checked
// whether or not bridge mode is on, since a file switched to bridge later carries
// the same list), the three health-check timings and the core-dump limit.
func (c *Config) validateContainerArtifactValues(p Platform) error {
	key := platformKey(p)
	for i, entry := range c.NetworkBlock(p).Ports {
		if err := validContainerPort(fmt.Sprintf("%s.network.ports[%d]", key, i), entry); err != nil {
			return err
		}
	}
	hc := c.ContainerBlock(p).HealthCheck
	for _, f := range []struct {
		name, value string
		zeroOK      bool
	}{
		{"interval", hc.Interval, false},
		{"timeout", hc.Timeout, false},
		{"startPeriod", hc.StartPeriod, true},
	} {
		if err := validHealthDuration(key+".container.healthCheck."+f.name, f.value, f.zeroOK); err != nil {
			return err
		}
	}
	// Empty is what a Config built in code carries; the renderers take ContainerCore
	// for it (Container.CoreLimit).
	if v := c.ContainerBlock(p).Ulimits.Core; v != "" {
		if err := validCoreLimit(key+".container.ulimits.core", v); err != nil {
			return err
		}
	}
	return nil
}

// operatorImageRE is kubernetes.operator.image, which carries its whole reference in
// one string: a repository path, an optional :tag and an optional @sha256 digest.
// image.registry, checked on its own, is prefixed onto it at render.
var operatorImageRE = regexp.MustCompile(`^(?:` + refDomain + `/)?` + refPathComponent + `(?:/` + refPathComponent +
	`)*(?::[A-Za-z0-9_][A-Za-z0-9._-]{0,127})?(?:@sha256:[a-f0-9]{64})?$`)

// k8sQuantityRE is a Kubernetes resource quantity in its everyday spelling: a
// decimal number with an optional decimal (m, k, M, G, T, P, E) or binary (Ki-Ei)
// suffix. Narrower than Kubernetes' own grammar, which also takes a sign and an
// exponent: neither is a size anyone means, and 1e3 is a YAML float besides.
var k8sQuantityRE = regexp.MustCompile(`^(?:[0-9]+(?:\.[0-9]+)?|\.[0-9]+)(?:m|k|M|G|T|P|E|Ki|Mi|Gi|Ti|Pi|Ei)?$`)

// validateK8sArtifactValues checks the Kubernetes values written as-is into YAML:
// the operator bundle's image, cpu, mem and watch list (text/template substitutes
// them raw into the manifests `operator deploy` applies alongside the operator's
// cluster-wide RBAC), and the CR's two storage sizes.
func (c *Config) validateK8sArtifactValues() error {
	op := c.K8s.Operator
	if op.Image != "" && !operatorImageRE.MatchString(op.Image) {
		return fmt.Errorf("kubernetes.operator.image %q is not a valid image reference: it must be a lowercase "+
			"repository path with an optional :tag and @sha256:<digest> (e.g. "+
			"solace/pubsubplus-eventbroker-operator:1.4.2). It is written into the operator bundle as-is, where a "+
			"newline or a space would restructure it", op.Image)
	}
	for _, f := range []struct{ field, value string }{
		{"kubernetes.operator.cpu", op.CPU},
		{"kubernetes.operator.mem", op.Mem},
		{"kubernetes.storage.msgNodeSize", c.K8s.Storage.MsgNodeSize},
		{"kubernetes.storage.monNodeSize", c.K8s.Storage.MonNodeSize},
	} {
		if f.value != "" && !k8sQuantityRE.MatchString(f.value) {
			return fmt.Errorf("%s %q is not a Kubernetes quantity: write a number with an optional unit, like "+
				"500m, 2, 512Mi or 30Gi (no sign, no exponent). It is written into the manifest as-is, where "+
				"anything else would be misread or restructure it", f.field, f.value)
		}
	}
	// The same trimming k8s.splitWatch applies, so what is checked is what reaches
	// WATCH_NAMESPACE. Each entry is a namespace, so it is held to the same label
	// rule kubernetes.namespace is.
	for i, ns := range strings.Split(op.WatchNamespaces, ",") {
		if err := validDNSLabel(fmt.Sprintf("kubernetes.operator.watchNamespaces entry %d", i+1),
			strings.TrimSpace(ns)); err != nil {
			return err
		}
	}
	return nil
}

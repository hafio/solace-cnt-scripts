package k8s

import (
	"context"
	"fmt"
	"strings"

	"solace/internal/output"
)

// This file is `check deploy` on Kubernetes: the grouped, tagged report an
// operator reads before deploying.
//
// The tag on each row is the point. Most of a check report is the env file read
// back -- useful, but not evidence of anything -- while a handful of rows cost a
// cluster round-trip and actually prove something. Rendering both the same way
// invited reading the whole report as "validated", when most of it was never
// checked at all. So:
//
//	[INFO]  echoed from the env file. Not checked.
//	[ OK ]  checked against the cluster, and good.
//	[FAIL]  checked, and it will stop a deploy.
//	[WARN]  checked, and worth knowing before you deploy.
//	[SKIP]  not checked, and the report says so rather than implying a pass.
//
// The second decision here: every check runs before any of them reports a
// failure, so one run tells you everything that is wrong. The old flow returned
// at the first failure, which meant fixing an unreachable cluster only to
// discover the StorageClass was wrong too.

// checkKeyWidth is the key column shared by every section of the check report. A
// constant rather than a computed width because the sections are built by
// different methods and must line up with each other.
const checkKeyWidth = 17

// checkRow is one report row: a verdict, a key, and a value.
type checkRow struct {
	Level output.Level
	Key   string
	Value string
}

// info/ok/fail/warn/skipped build a row at each level. They exist so the section
// builders below read as data rather than as a wall of struct literals.
func info(key, format string, a ...any) checkRow {
	return checkRow{output.LevelInfo, key, fmt.Sprintf(format, a...)}
}
func okRow(key, format string, a ...any) checkRow {
	return checkRow{output.LevelOK, key, fmt.Sprintf(format, a...)}
}
func failRow(key, format string, a ...any) checkRow {
	return checkRow{output.LevelFail, key, fmt.Sprintf(format, a...)}
}
func warnRow(key, format string, a ...any) checkRow {
	return checkRow{output.LevelWarn, key, fmt.Sprintf(format, a...)}
}
func skipRow(key, format string, a ...any) checkRow {
	return checkRow{output.LevelSkip, key, fmt.Sprintf(format, a...)}
}

// checkReport accumulates the rows of each section, so the verification steps can
// contribute to whichever section they belong in and the whole thing renders in a
// fixed order at the end.
type checkReport struct {
	config      []checkRow
	deployment  []checkRow
	operator    []checkRow
	broker      []checkRow
	credentials []checkRow
	placement   []checkRow
}

// failed reports whether any row will stop a deploy.
func (rep *checkReport) failed() int {
	n := 0
	for _, section := range [][]checkRow{rep.config, rep.deployment, rep.operator, rep.broker, rep.credentials, rep.placement} {
		for _, row := range section {
			if row.Level == output.LevelFail {
				n++
			}
		}
	}
	return n
}

// render writes the sections in the order an operator reads them: what is being
// deployed and whether the cluster will take it, then the operator it depends on,
// then the broker's own shape, then credentials, then placement.
func (rep *checkReport) render(r *output.Sink) {
	for _, section := range []struct {
		title string
		rows  []checkRow
	}{
		{"Config", rep.config},
		{"Deployment", rep.deployment},
		{"Operator", rep.operator},
		{"Broker", rep.broker},
		{"Credentials", rep.credentials},
		{"Placement", rep.placement},
	} {
		if len(section.rows) == 0 {
			continue
		}
		r.Section(section.title)
		for _, row := range section.rows {
			r.KVRowAt(row.Level, checkKeyWidth, row.Key, row.Value)
		}
		r.Line("")
	}
}

// CheckDeploy is the report behind `check deploy`. It confirms the env file was
// accepted and shows what it resolved to (the ports included), echoes the rest
// of the config, probes what needs a cluster, and returns an error naming how
// many checks failed -- after reporting all of them.
func (c *Cluster) CheckDeploy(ctx context.Context) error {
	rep := &checkReport{}
	c.validationRows(rep)
	c.configRows(rep)

	if c.isEcho() {
		// A preview still ISSUES the reads, so the Echo runner records the argv --
		// that is what makes the whole command tree assertable without a cluster.
		// What it cannot do is decode an answer, so each cluster-backed row says
		// it was skipped rather than being quietly omitted.
		_ = c.Reachable(ctx)
		_ = c.Preflight(ctx, "create", brokerResource)
		_, _ = c.resolveStorageClass(ctx)
		rep.deployment = append(rep.deployment,
			skipRow("api server", "skipped (preview)"),
			skipRow("permission", "skipped (preview)"))
		rep.operator = append(rep.operator, skipRow("installed", "skipped (preview)"))
		rep.broker = append(rep.broker, skipRow("storage class", "skipped (preview)"))
		rep.render(c.report())
		return nil
	}

	c.verifyRows(ctx, rep)
	rep.render(c.report())

	if n := rep.failed(); n > 0 {
		return fmt.Errorf("%d check(s) failed; the rows tagged [FAIL] above say what to fix", n)
	}
	return nil
}

// validationRows is the one section that is NOT an echo: config.Load already ran
// Validate over the whole env file before this command's report could even start
// building, so a malformed kubernetes.ports entry (or any other invalid field)
// would have failed loudly right there, before CheckDeploy ever runs. These rows
// say so plainly instead of leaving an operator to infer it, and report what
// Validate actually leaves resolved: cfg.K8s.Ports, already filled by
// ApplyDefaults when the env file left kubernetes.ports unset. That is real,
// already-checked information -- not merely echoed -- so unlike configRows below
// these are [ OK ], and deliberately can never be [FAIL]: `check deploy` is
// read-only and must not fail just for having ports to show.
func (c *Cluster) validationRows(rep *checkReport) {
	rep.config = append(rep.config,
		okRow("env file", "accepted -- an invalid field would have stopped this command before this report could run"))
	rep.config = append(rep.config, portRows(c.Cfg.K8s.Ports)...)
}

// portsPerRow groups the ports row into fixed-width chunks rather than one row
// per port. kubernetes.ports defaults to 17 entries (defaultK8sPorts); one row
// each would make this section alone longer than every other section in the
// report combined, and one unwrapped line would run far past what every other
// row in this report reads as (KVRowAt's tag plus the checkKeyWidth key column
// already costs 27 columns before the value even starts). Four
// "name=port" pairs (~18 characters apiece with the ", " separator) keeps each
// wrapped line under 100 columns, matching the width the rest of the report
// settles at, while still fitting the default 17 onto five lines instead of
// seventeen.
const portsPerRow = 4

// portRows renders ports as a leading count row plus its pairs wrapped across
// as many continuation rows as it takes. cfg.K8s.Ports entries are already in
// the exact "name=port[:service][/proto]" form render.parsePort decodes, so
// they are printed verbatim -- reformatting them here would risk the report
// showing something other than what actually gets rendered into the broker CR.
func portRows(ports []string) []checkRow {
	if len(ports) == 0 {
		return []checkRow{okRow("ports", "(none configured)")}
	}
	rows := make([]checkRow, 0, len(ports)/portsPerRow+1)
	for i := 0; i < len(ports); i += portsPerRow {
		end := i + portsPerRow
		if end > len(ports) {
			end = len(ports)
		}
		key := ""
		if i == 0 {
			key = fmt.Sprintf("ports (%d)", len(ports))
		}
		rows = append(rows, okRow(key, "%s", strings.Join(ports[i:end], ", ")))
	}
	return rows
}

// configRows is everything readable from the env file alone. Every row is [INFO]:
// it is what you asked for, not proof of anything.
func (c *Cluster) configRows(rep *checkReport) {
	cfg := c.Cfg

	mode := "standalone (single broker)"
	if cfg.RedundancyEnabled() {
		mode = "HA redundancy group (primary + backup + monitor)"
	}
	rep.deployment = append(rep.deployment,
		info("namespace", "%s", cfg.K8s.Namespace),
		info("name", "%s", cfg.K8s.Name),
		info("image", "%s", cfg.Image.Ref()),
		info("image pull", "secret=%s creds=%s",
			orNone(cfg.K8s.ImagePullSecret), setOrNone(cfg.Image.User != "" && cfg.Image.Pass != "")),
		info("redundancy", "%s", mode),
		info("cluster cmd", "%s", cfg.K8s.Runtime.String()),
	)

	rep.operator = append(rep.operator,
		info("image", "%s", operatorImage(cfg)),
		info("cpu", "%s", orNone(cfg.K8s.Operator.CPU)),
		info("mem", "%s", orNone(cfg.K8s.Operator.Mem)),
		// An empty watch list does not mean "the broker namespace" -- it means every
		// namespace in the cluster, the widest scope the operator has and the one
		// least likely to be chosen on purpose.
		info("watch scope", "%s", orValue(watchNamespace(cfg),
			"(empty -- the operator watches ALL namespaces)")),
	)

	rep.broker = append(rep.broker,
		info("update strategy", "%s", orNone(cfg.K8s.UpdateStrategy)),
		info("scaling tier", "%d connections", cfg.Scaling.MaxConnections),
		// MsgNode.Mem, not Storage.MsgNode: the first is the container's memory
		// limit (rendered as the CR's messagingNodeMemory), the second is the size
		// of its data volume. They are different numbers with similar names, and
		// reporting the disk size as "mem" is exactly the confusion this row exists
		// to prevent.
		info("cpu", "%s", orNone(cfg.Scaling.CPU)),
		info("mem", "%s", orNone(cfg.K8s.MsgNode.Mem)),
		info("message storage", "%s", cfg.K8s.Storage.MsgNode),
		info("monitor storage", "%s", orNone(cfg.K8s.Storage.MonNode)),
	)

	// user= is the literal "admin": the operator reads the fixed
	// username_admin_password key out of the credentials Secret, so that is the
	// broker's admin user whatever an env file says.
	rep.credentials = append(rep.credentials,
		info("admin user", "admin"),
		info("admin secret", "%s", orNone(cfg.K8s.AdminSecret)),
		info("passwords", "admin=%s monitor=%s",
			setOrMissing(cfg.Admin.Pass), setOrNone(cfg.Admin.MonitorPass != "")),
		info("additional users", "%d", len(cfg.Admin.AdditionalUsers)),
	)
	if cfg.K8s.TLSServerSecret != "" {
		rep.credentials = append(rep.credentials,
			info("tls secret", "%s", cfg.K8s.TLSServerSecret),
			info("tls files", "cert=%s key=%s cas=%d",
				orNone(cfg.TLS.Cert), setOrMissing(cfg.TLS.CertKey), len(cfg.TLS.CAs)))
	} else {
		rep.credentials = append(rep.credentials, info("tls", "(not configured)"))
	}

	p := cfg.K8s.Placement
	rep.placement = append(rep.placement,
		info("loadBalancer", "ip=%s pool=%s",
			orNone(cfg.K8s.LoadBalancer.IP), orNone(cfg.K8s.LoadBalancer.IPPool)),
		info("node labels", "p=%d b=%d m=%d",
			len(p.LabelsPrimary), len(p.LabelsBackup), len(p.LabelsMonitor)),
		info("anti-affinity", "%d namespace(s)", len(p.AntiAffinityNS)),
	)
}

// verifyRows runs the checks that need a cluster. Reachability comes first: with
// no API server every other probe would fail for the same reason, and three
// identical failures are less useful than one plus three honest skips.
func (c *Cluster) verifyRows(ctx context.Context, rep *checkReport) {
	if err := c.Reachable(ctx); err != nil {
		rep.deployment = append(rep.deployment, failRow("api server", "%v", err))
		const why = "skipped (cluster unreachable)"
		rep.deployment = append(rep.deployment, skipRow("permission", why))
		rep.operator = append(rep.operator, skipRow("installed", why))
		rep.broker = append(rep.broker, skipRow("storage class", why))
		return
	}
	rep.deployment = append(rep.deployment, okRow("api server", "reachable"))

	if err := c.Preflight(ctx, "create", brokerResource); err != nil {
		rep.deployment = append(rep.deployment, failRow("permission", "%v", err))
	} else {
		rep.deployment = append(rep.deployment, okRow("permission", "create %s", brokerResource))
	}

	c.operatorRows(ctx, rep)
	c.storageRows(ctx, rep)
}

// operatorRows reports the operator this deploy depends on: whether it is
// installed, where, and -- the row worth having -- which version is running,
// since a cluster can carry a different one from the env file and reconciliation
// follows the installed one.
func (c *Cluster) operatorRows(ctx context.Context, rep *checkReport) {
	ns, origin := c.operatorNSOrigin(ctx)
	rep.operator = append(rep.operator, okRow("namespace", "%s (%s)", ns, origin))

	if !c.OperatorInstalled(ctx) {
		// A WARNING, not a failure, and deliberately so: `check deploy` is
		// read-only, and a cluster where the operator is not installed YET is a
		// perfectly reasonable thing to be checking -- installing it is the next
		// command. So this reports the gap without failing the check itself. The
		// row is prominent where the old trailing warning was easy to miss.
		rep.operator = append(rep.operator, warnRow("installed",
			"not found -- `deploy operator` installs it, and `deploy broker` will not work without it"))
		return
	}
	rep.operator = append(rep.operator, okRow("installed", "yes"))

	running := c.installedOperatorImage(ctx)
	if running == "" {
		return
	}
	_, runningTag := imageTag(running)
	_, wantTag := imageTag(operatorImage(c.Cfg))
	switch cmp, ok := compareVersions(runningTag, wantTag); {
	case !ok || cmp == 0:
		rep.operator = append(rep.operator, okRow("running version", "%s", orNone(runningTag)))
	case cmp > 0:
		// The cluster is ahead. `deploy operator` would roll it back, which is
		// what the downgrade prompt guards -- worth knowing here too.
		rep.operator = append(rep.operator, warnRow("running version",
			"%s, NEWER than the configured %s -- `deploy operator` would downgrade it", runningTag, wantTag))
	default:
		rep.operator = append(rep.operator, warnRow("running version",
			"%s, older than the configured %s -- `deploy operator` would upgrade it", runningTag, wantTag))
	}
}

// storageRows validates the StorageClass the broker's claims will bind. It is the
// check that most often explains a broker whose pods never start.
func (c *Cluster) storageRows(ctx context.Context, rep *checkReport) {
	name, err := c.resolveStorageClass(ctx)
	if err != nil {
		rep.broker = append(rep.broker, failRow("storage class", "%v", err))
		return
	}
	if name == "" {
		// (009) is the pre-Go bash script this check ports from -- provenance
		// for tracing the port, not something a user reading this line needs.
		rep.broker = append(rep.broker, failRow("storage class",
			"no default StorageClass found and kubernetes.storage.class is not set"))
		return
	}
	binding, expansion, err := c.storageClassAttrs(ctx, name)
	if err != nil {
		rep.broker = append(rep.broker, failRow("storage class", "%v", err))
		return
	}
	// Report what was actually found, not a summary of what was wanted: when a
	// class is rejected, the values are the whole explanation.
	value := fmt.Sprintf("%s (volumeBindingMode=%s allowVolumeExpansion=%s)", name, binding, expansion)
	if err := storageClassSuitable(name, binding, expansion); err != nil {
		rep.broker = append(rep.broker, failRow("storage class", "%s -- %v", value, err))
		return
	}
	rep.broker = append(rep.broker, okRow("storage class", "%s", value))
}

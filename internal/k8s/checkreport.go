package k8s

import (
	"context"
	"fmt"
	"strings"

	"solace/internal/config"
	"solace/internal/output"
)

// This file is `validate` on Kubernetes: the grouped, tagged report an operator reads
// before deploying, in three views (whole file, broker half, operator half).
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

// validateScope selects which halves of the report render.
//
// There is ONE report builder and three views of it, rather than three builders. That is
// what makes `validate` provably equal to `broker validate` plus `operator validate`: the
// rows cannot drift between the scoped commands and the whole-file one, because there is
// only one set of them. A single env file describes both nouns, so the whole-file view is
// the one to run; the scoped views exist for when only half is interesting.
type validateScope int

const (
	scopeAll      validateScope = iota // both halves -- top-level `validate`
	scopeBroker                        // `broker validate`
	scopeOperator                      // `operator validate`
)

// wantsBroker and wantsOperator say which sections a scope renders. The
// operator-COVERAGE row (does the installed operator watch this broker's namespace?)
// belongs to the broker half, not the operator one: it is a precondition for deploying
// this broker, which is the question `broker validate` is asked.
func (s validateScope) wantsBroker() bool   { return s != scopeOperator }
func (s validateScope) wantsOperator() bool { return s != scopeBroker }

// sections returns the report's sections in the order an operator reads them, filtered to
// the scope: what is being deployed and whether the cluster will take it, then the
// operator it depends on, then the broker's own shape, then credentials, then placement.
func (rep *checkReport) sections(scope validateScope) []struct {
	title string
	rows  []checkRow
} {
	type section = struct {
		title string
		rows  []checkRow
	}
	var out []section
	if scope.wantsBroker() {
		out = append(out, section{"Config", rep.config}, section{"Deployment", rep.deployment})
	}
	if scope.wantsOperator() {
		out = append(out, section{"Operator", rep.operator})
	}
	if scope.wantsBroker() {
		out = append(out,
			section{"Broker", rep.broker},
			section{"Credentials", rep.credentials},
			section{"Placement", rep.placement})
	}
	return out
}

// failed counts the rows that will stop a deploy, within the scope being rendered. A
// failure in a section this view does not show must not fail this view: `operator
// validate` is not the place a bad StorageClass is reported.
func (rep *checkReport) failed(scope validateScope) int {
	n := 0
	for _, section := range rep.sections(scope) {
		for _, row := range section.rows {
			if row.Level == output.LevelFail {
				n++
			}
		}
	}
	return n
}

// render writes the scope's sections, each with the shared key width so the halves line
// up with each other when both are shown.
func (rep *checkReport) render(r *output.Sink, scope validateScope) {
	for _, section := range rep.sections(scope) {
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

// Validate is the whole-env-file report behind the top-level `validate`. A single env file
// describes both the broker and the operator, so this is the view to run.
func (c *Cluster) Validate(ctx context.Context) error {
	return c.validate(ctx, scopeAll)
}

// ValidateBroker is `broker validate`: everything about this broker, plus whether the
// installed operator actually covers its namespace -- which is a precondition for
// deploying it, not an operator detail.
func (c *Cluster) ValidateBroker(ctx context.Context) error {
	return c.validate(ctx, scopeBroker)
}

// ValidateOperator is `operator validate`: the operator's install state, version and
// watch scope, and nothing about this broker.
func (c *Cluster) ValidateOperator(ctx context.Context) error {
	return c.validate(ctx, scopeOperator)
}

// validate builds the whole report and renders the requested view. It confirms the env
// file was accepted and shows what it resolved to (the ports included), echoes the rest of
// the config, probes what needs a cluster, and returns an error naming how many checks
// failed -- after reporting all of them, so one run tells you everything that is wrong.
func (c *Cluster) validate(ctx context.Context, scope validateScope) error {
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
			skipRow("permission", "skipped (preview)"),
			skipRow("operator watch", "skipped (preview)"))
		rep.operator = append(rep.operator, skipRow("installed", "skipped (preview)"))
		rep.broker = append(rep.broker, skipRow("storage class", "skipped (preview)"))
		rep.render(c.report(), scope)
		return nil
	}

	c.verifyRows(ctx, rep)
	rep.render(c.report(), scope)

	if n := rep.failed(scope); n > 0 {
		return fmt.Errorf("%d check(s) failed; the rows tagged [FAIL] above say what to fix", n)
	}
	return nil
}

// validationRows is the one section that is NOT an echo: config.Load already ran
// Validate over the whole env file before this command's report could even start
// building, so a malformed kubernetes.ports entry (or any other invalid field)
// would have failed loudly right there, before this report ever runs. These rows
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
		// What THIS env file asks for, which is not necessarily what is running: the
		// operator is shared, so a deploy applies the union of this list and the
		// installed one. watchCoverageRow reports what the cluster actually says.
		//
		// An empty list does not mean "the broker namespace" -- it means every
		// namespace in the cluster, the widest scope the operator has and the one
		// least likely to be chosen on purpose.
		info("watch scope (env)", "%s", orValue(watchNamespace(cfg),
			"(empty -- the operator would watch ALL namespaces)")),
	)

	rep.broker = append(rep.broker,
		info("update strategy", "%s", orNone(cfg.K8s.UpdateStrategy)),
		info("scaling tier", "%d connections", cfg.Scaling.MaxConnections),
		// MsgNode.Mem, not Storage.MsgNodeSize: the first is the container's memory
		// limit (rendered as the CR's messagingNodeMemory), the second is the size
		// of its data volume. They are different numbers with similar names, and
		// reporting the disk size as "mem" is exactly the confusion this row exists
		// to prevent.
		info("cpu", "%s", orNone(cfg.Scaling.CPU)),
		info("mem", "%s", orNone(cfg.K8s.MsgNode.Mem)),
		info("message storage", "%s", orNone(cfg.K8s.Storage.MsgNodeSize)),
		info("monitor storage", "%s", orNone(cfg.K8s.Storage.MonNodeSize)),
	)
	rep.broker = append(rep.broker, storageRows(cfg)...)

	// user= is the literal "admin": the operator reads the fixed
	// username_admin_password key out of the credentials Secret, so that is the
	// broker's admin user whatever an env file says.
	rep.credentials = append(rep.credentials,
		info("admin user", "admin"),
		info("admin secret", "%s", orNone(cfg.K8s.AdminSecret)),
		info("passwords", "admin=%s monitor=%s",
			setOrMissing(cfg.SEMP.AdminPass), setOrNone(cfg.SEMP.MonitorPass != "")),
		additionalUsersRow(cfg),
	)
	// The pre-shared key is optional here and mandatory on containers, so an empty one
	// is a legitimate deployment rather than a gap -- but which of the two keys the
	// group will end up using is not visible anywhere else, since neither the CR nor
	// the Secret is written until deploy. Standalone has no group, so no row.
	if cfg.RedundancyEnabled() {
		rep.credentials = append(rep.credentials, pskRow(cfg))
	}
	if cfg.K8s.TLSServerSecret != "" {
		// Naming the Secret and supplying the files it is built from are separate
		// decisions, and the row has to say which one this env file made: otherwise
		// "tls secret: solace-tls-secret" reads as "this tool will create it" in the
		// case where it will not, and a missing Secret is only discovered by the
		// operator failing to mount it.
		origin := "built from tls.cert/tls.certKey"
		if !cfg.ManagesTLSSecret() {
			origin = "must already exist -- this env file supplies no cert/key"
		}
		rep.credentials = append(rep.credentials,
			info("tls secret", "%s (%s)", cfg.K8s.TLSServerSecret, origin),
			info("tls files", "cert=%s key=%s cas=%d",
				orNone(cfg.TLS.Cert), setOrMissing(cfg.TLS.CertKey), len(cfg.TLS.CAs)))
		if cfg.TLS.CertPassphrase != "" || cfg.TLS.CertPassphraseEnv != "" {
			// The CRD's spec.tls carries only serverTlsConfigSecret, certFilename,
			// certKeyFilename and enabled -- there is no passphrase field and no Secret
			// key the operator reads one from. An encrypted key therefore cannot be used
			// on this platform, and saying nothing would leave that to be discovered by
			// a broker that will not start.
			rep.credentials = append(rep.credentials, warnRow("tls passphrase",
				"set, but NOT applied on kubernetes -- the CRD has no passphrase field; "+
					"supply an unencrypted key, or decrypt it into the Secret yourself"))
		}
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
		rep.deployment = append(rep.deployment,
			skipRow("permission", why),
			skipRow("operator watch", why))
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

	// ONE cluster-wide `get deployment --all-namespaces`, read twice. The watch row
	// wants its WATCH_NAMESPACE and the operator rows want its running image, both off
	// the same object -- and each used to fetch it for itself, so every healthy
	// `validate` and `check deploy` paid for a full cluster-wide list twice. The error
	// is threaded rather than swallowed, because "could not be read" and "no operator
	// installed" are different answers and the watch row already says so.
	dep, depErr := c.findOperatorDeployment(ctx)
	c.watchCoverageRow(rep, dep, depErr)
	c.operatorRows(ctx, rep, dep)
	c.storageRows(ctx, rep)
}

// watchCoverageRow answers the question this report exists to answer about the operator:
// will the operator that is RUNNING reconcile this broker?
//
// It is in the Deployment section, not the Operator one, because it is a precondition for
// deploying this broker rather than a fact about the operator -- so `broker validate`
// shows it and `operator validate` does not.
//
// The value is read from the cluster, never rendered from config. The old report had a
// config-only "watch scope" row that stated what this env file WOULD apply, which is
// exactly the value that can silently differ from what is running once a second env file
// has deployed the operator. Every verdict here is a [WARN] rather than a [FAIL]: an
// operator that does not yet cover this namespace is fixed by `operator deploy`, and this
// command is also what you run before installing anything.
func (c *Cluster) watchCoverageRow(rep *checkReport, dep *deploymentItem, err error) {
	ns := c.ns()
	installed, allNS, found := watchFromDeployment(dep)
	switch {
	case err != nil:
		rep.deployment = append(rep.deployment, warnRow("operator watch",
			"could not be read: %v", err))
	case !found:
		rep.deployment = append(rep.deployment, warnRow("operator watch",
			"no operator is installed -- run `solace-util operator deploy`"))
	case allNS:
		rep.deployment = append(rep.deployment, okRow("operator watch",
			"ALL namespaces, so %q is covered", ns))
	case containsString(installed, ns):
		rep.deployment = append(rep.deployment, okRow("operator watch",
			"%s (%d namespace(s)) -- %q is covered", strings.Join(installed, ","), len(installed), ns))
	default:
		rep.deployment = append(rep.deployment, warnRow("operator watch",
			"the operator watches %s and NOT %q, so this broker would not be reconciled -- "+
				"`solace-util operator deploy` adds it", strings.Join(installed, ","), ns))
	}
}

// containsString is a membership test over a small slice. A loop rather than a map: these
// lists are a handful of namespaces long, and building a map to ask one question costs
// more than the scan.
func containsString(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// operatorRows reports the operator this deploy depends on: whether it is
// installed, where, and -- the row worth having -- which version is running,
// since a cluster can carry a different one from the env file and reconciliation
// follows the installed one.
// storageRows says where each broker node's data volume comes from.
//
// "which volume is this broker actually on" is the question worth answering before a
// deploy, and it is not readable off the env file at a glance once custom mounts are in
// play: a role named in customVolumeMount mounts an existing claim, everything else is
// provisioned. Reported per role so a claim meant for one node and typed under another is
// visible rather than inferred.
func storageRows(cfg *config.Config) []checkRow {
	if !cfg.K8s.Storage.UsesCustomMounts() {
		if cfg.K8s.Storage.Class != "" {
			return []checkRow{info("storage", "provisioned from StorageClass %q", cfg.K8s.Storage.Class)}
		}
		return []checkRow{info("storage", "provisioned from the cluster's default StorageClass")}
	}
	var rows []checkRow
	for _, role := range HARoles(cfg) {
		claim, ok := cfg.K8s.Storage.CustomMountFor(role)
		if !ok {
			// Unreachable while validateStorage demands all-or-none, and reported rather
			// than assumed away: if that rule is ever relaxed, this is the row that says
			// which node quietly went back to default provisioning.
			rows = append(rows, warnRow("storage "+role.Word(), "cluster default StorageClass (no custom mount)"))
			continue
		}
		rows = append(rows, info("storage "+role.Word(), "custom mount %s (never deleted by --delete-data)", claim))
	}
	return rows
}

// operatorRows takes the already-fetched operator Deployment rather than looking it up:
// verifyRows reads it once for both this and the watch row. It still takes a ctx,
// because the namespace origin and the CRD check are separate reads of their own.
func (c *Cluster) operatorRows(ctx context.Context, rep *checkReport, dep *deploymentItem) {
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

	running := imageFromDeployment(dep)
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

// allCustomMounted reports whether EVERY node in the group mounts an existing claim, so
// the operator provisions nothing and no StorageClass is in play.
//
// Asked per role rather than off UsesCustomMounts alone: validateStorage demands all-or-
// none today, so the two agree, but if that rule is ever relaxed a partly covered group
// still has claims to bind and must keep being checked.
func allCustomMounted(cfg *config.Config) bool {
	roles := HARoles(cfg)
	if len(roles) == 0 {
		return false
	}
	for _, role := range roles {
		if _, ok := cfg.K8s.Storage.CustomMountFor(role); !ok {
			return false
		}
	}
	return true
}

// storageRows validates the StorageClass the broker's claims will bind. It is the
// check that most often explains a broker whose pods never start.
//
// Unless there are no claims to bind. With every node on a customVolumeMount the operator
// provisions nothing, so there is no class in play: validateStorage refuses class and
// customVolumeMount together, so resolveStorageClass would fall through to the cluster
// default and FAIL the whole report over a class the deploy never touches -- advising a
// fix (`set kubernetes.storage.class`) that config.Load then refuses. The config-side
// storageRows above has always branched on this; the cluster-side namesake had not, and
// the shared name is what hid the gap.
func (c *Cluster) storageRows(ctx context.Context, rep *checkReport) {
	if allCustomMounted(c.Cfg) {
		rep.broker = append(rep.broker, info("storage class",
			"not checked -- every node mounts an existing claim, so nothing is provisioned"))
		return
	}
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

// additionalUsersRow reports admin.additionalUsers and where their credentials land.
//
// They ARE applied on Kubernetes now, through a Secret of their own that the CR names in
// spec.extraEnvVarsSecret -- so this is an [INFO] echo rather than the warning it used to
// be. What the row still has to say is that the passwords reach the broker as environment
// variables here, unlike every other secret this tool handles: the CRD has no way to mount
// an arbitrary Secret as files, so envFrom is the only channel, and someone reviewing where
// their credentials end up should not have to read the CRD to find that out. The container
// platforms are unaffected: they create these users at boot from the mounted password file.
//
// The replacement is planned as a second mounted Secret surfaced as environment variables
// (see broker.AdditionalUsers). Drop this warning back to an info row when that lands.
// pskRow says which pre-shared key an HA group will authenticate with. Set, the value
// becomes the preshared_auth_key entry of the credentials Secret and the CR points at it;
// empty, the CR omits preSharedAuthKeySecret entirely and the operator generates and
// distributes a key of its own. Both are supported, so this is [INFO] rather than a
// verdict -- but it is the only place the choice is stated before a deploy makes it.
func pskRow(cfg *config.Config) checkRow {
	if cfg.Redundancy.PSK == "" && cfg.Redundancy.PSKEnv == "" {
		return info("preshared key", "(not set) -- the operator generates and distributes its own")
	}
	return info("preshared key", "set -- %s in %s", pskSecretKey, orNone(cfg.K8s.AdminSecret))
}

func additionalUsersRow(cfg *config.Config) checkRow {
	n := len(cfg.SEMP.AdditionalUsers)
	if n == 0 {
		return info("additional users", "0")
	}
	return info("additional users",
		"%d in Secret %s -- passwords reach the broker as environment variables here, not as "+
			"mounted files (the CRD offers no volume passthrough)", n, cfg.AdditionalUsersSecretName())
}

# Critical review - solace-util (solace-cnt-scripts)

Date: 2026-11-11
Commit: b0941b1 on main
Previous report: none
Scope: all areas
Stacks: Go (core + go.md); bash (bash.md); GitHub Actions (github-actions.md); containers (containers.md); core only: PowerShell, HTML/JS
Graph: reused; 4065 nodes, 20953 links, 110 communities
Dynamic checks: none run; offered as a hand-over (gates are user-run under CLAUDE.md S1)

## Summary

| Severity | New | Still open | Fixed since last | Total open |
|---|---|---|---|---|
| Blocker | 4 | 0 | 0 | 4 |
| Major | 13 | 0 | 0 | 13 |
| Minor | 50 | 0 | 0 | 50 |
| Nit | 13 | 0 | 0 | 13 |
| Suspected (uncounted) | 10 | 0 | 0 | 10 |

The codebase is disciplined: the stated import layering holds in every file, the gates recorded today are
green (build, vet, test, scan; coverage 97.7%), and decisions are documented where they are implemented. The
problems cluster where a guard exists and one path goes around it. The sink-grammar invariant misses seven
broker CR fields and a watch list read back from the cluster, and one of those fields also reaches kubectl
as a bare positional. Several destructive paths act wider than what the operator was shown: a VPN-scoped
import clears every virtual hostname on the target, the docker removal empties the data directory without
confirming the container is gone, and the namespace emptiness gate does not list RBAC, network or custom
resources. The DR switch gate can pass on a site that is running on its backup. This is the first run, so
there is no trend.

## Top risks

1. CR-0001 [blocker] A VPN-scoped import deletes every virtual hostname on the target, silently -
internal/broker/importops.go:708-722. Importing a --vpn-scoped artifact removes every virtual hostname on
the target; no plan, prompt, result or verification row says so.
2. CR-0002 [blocker] Seven kubernetes.* values reach the broker CR unquoted with no load-time grammar
(YAML/--- injection) - internal/render/render.go:100-102. Env-file text can add keys or whole manifests to
what `broker deploy` applies and `broker remove` deletes.
3. CR-0003 [blocker] kubernetes.storage.class reaches kubectl argv as a bare positional, so config text can
inject kubectl flags - internal/k8s/check.go:46-57. A dash-led storage class becomes a kubectl flag
(`--kubeconfig=...`) during `validate`, which can run a credential plugin.
4. CR-0004 [blocker] WATCH_NAMESPACE read back from a running Deployment is rendered raw into the operator
bundle an admin applies - internal/k8s/watch.go:128-158. A decoy Deployment's WATCH_NAMESPACE value is
templated unescaped into the bundle a cluster admin applies.
5. CR-0005 [major] Docker broker remove empties the data dir under a broker compose down did not find -
internal/container/manager.go:999-1023. With another compose project in effect, `compose down` finds nothing
and exits 0, and --delete-data empties the live data directory.
6. CR-0006 [major] Namespace emptiness gate misses RBAC, NetworkPolicy, Ingress and CR objects; --no-prompt
then deletes the namespace - internal/cli/ops_k8s.go:790-823. A namespace holding only RoleBindings,
NetworkPolicies, Ingresses or custom resources reads as empty; --no-prompt then cascade-deletes it.
7. CR-0007 [major] Primary-active switch gate is inverted when the local site is addressed through its
backup node - internal/cli/replication.go:142-162. With `--pod backup` the gate passes when the backup holds
activity and refuses the healthy group.
8. CR-0008 [major] PrimaryActive reads Enabled-Released (or an unreadable report) as standalone and passes
the gate - internal/broker/matechannel.go:222-237. A primary in Enabled-Released, or an unreadable report,
is treated as standalone and the switch proceeds.
9. CR-0009 [major] An operator found in another namespace drives deploy and remove, which then write and
delete in the resolved namespace - internal/k8s/watch.go:90-97. `operator remove` deletes the ClusterRole
and binding of an operator its prompt never named; `operator deploy` installs a second one.
10. CR-0010 [major] Create Usernames chunk cannot replay onto a target that already has a non-admin CLI user
- internal/broker/transform.go:205-232. A re-run, or a target that already has the same CLI users, fails the
usernames chunk after the VPNs were torn down.

Further majors, in their area sections: CR-0011, CR-0012, CR-0013, CR-0014, CR-0015, CR-0016, CR-0017.

## Findings by area

### design

#### CR-0018 [minor] [new] platformOps is the single source of truth only for dispatchLeaf; 12 commands build it in RunE

Location: internal/cli/platform.go:288-304, internal/cli/platform.go:317-333,
internal/cli/commands.go:124-134, internal/cli/commands.go:165-175, internal/cli/commands.go:206-220,
internal/cli/commands.go:232-244, internal/cli/commands.go:259-269, internal/cli/commands.go:301-311,
internal/cli/commands.go:322-349, internal/cli/commands.go:607-667
Confidence: verified
Evidence:

    CLAUDE.md:101-102 says "the `platformOps` map is the single source of truth: its keys set the
    annotation help reads, its values are what runs." platform.go:291-293 says "the same entries that
    dispatch the call also tag the command, so help text and the pre-run refusal can never disagree".
    Only dispatchLeaf (317-321) does that: `return onlyOn(c, supported(impls)...)`. deploy, generate,
    restart, status, logs, cli, shell, copy from, copy into, import-config, cli-script and shell-script
    build `dispatch(platformOps(...), app)` inside RunE (e.g. commands.go:131-132) and never call
    onlyOn, so their annotation is absent, which means every platform. Step by step for the likely next
    change: give `broker copy into` a Kubernetes-only implementation, `platformOps(fn, nil)`. On docker,
    prepare passes it because the absent annotation means every platform, help shows no "(kubernetes
    only)", the env file loads and the preamble runs. Then dispatch (329-331) returns `no implementation
    for platform docker` with exit 1, the drift case platform.go:324-326 says "prepare already refuses".
    No test covers these 12 RunE-built maps (TestPlatformOpsCoversEveryPlatform tests only the builder).

Why it matters: The stated invariant holds for about half of the dispatching commands. The next
platform-specific change to one of the other half produces exactly the help/pre-run disagreement the map was
built to rule out.

Change: Route every dispatching command through one builder that both sets RunE from impls and calls
onlyOn(c, supported(impls)...). For example dispatchCmd(app, c, impls), plus an args variant taking
func([]string) map[config.Platform]opFunc for copy/import/cli-script. Alternatively, add a tree-walk test
asserting every wireExec command without the platform annotation dispatches on all of Platforms().

#### CR-0019 [minor] [new] A removed dry-run mode survives as 34 production branches keyed on the test-only Echo runner

Location: internal/engine/runner.go:122-129, internal/k8s/check.go:22-28, internal/k8s/preflight.go:22-25,
internal/k8s/preflight.go:62-70, internal/container/euid.go:31-40, internal/container/manager.go:134-136,
docs/operations.md:991-992, docs/operations.md:1283-1286, docs/operations.md:2055-2058,
docs/operations.md:1102-1106, internal/cli/context.go:57-64, internal/cli/context.go:168-173,
internal/container/euid.go:30-41, internal/container/manager.go:521-540, internal/k8s/deploy.go:62-88,
internal/container/preflight.go:22-25, internal/cli/cli_test.go:1786-1791, docs/test.md:939
Confidence: verified
Evidence:

    CLAUDE.md says "`runner.Echo` (from the shared module) is a TEST-ONLY runner installed through the
    `App.NewRunner` seam; nothing user-facing installs it." A grep finds no production construction of
    runner.Echo. Yet engine/runner.go:122-124 says "Callers use it to skip work that only makes sense
    against a real engine: file writes, host probes, questions a preview cannot answer". There are 34
    non-test call sites, for example: preflight.go:62-69 `if c.isEcho() { ... KVRow(..., "permission",
    "skipped (preview)"); return nil }`, the comment at preflight.go:24-25 "previewing without a cluster
    -- is already the Echo runner", and euid.go:38 `if engine.IsEcho(r) { return nil }`, which skips the
    podman rootful guard. The operator docs still describe the mode: operations.md:991 "A preview
    (`broker generate`, or any dry run) asks nothing", 1285 "A runner that answers nothing at all -- a
    dry run --", and 2057 "a preview (the Echo runner) never asks". All of that against 1104-1105:
    "There is no `--dry-run`".

Why it matters: Each echoRunner test through these sites exercises the preview branch, not the production
assertion: RBAC verdict, euid guard, secret read-back. Those branches are covered only where a bespoke stub
exists. The operator docs also promise a preview no user can reach. The same branches are what the tests
exercise, so the test-only runner is steering production control flow.

Change: Decide one way. Either delete the preview branches and give the tests a scripted runner (runnertest)
that answers like a permitted cluster or host, so they cover the production path. Or keep the branches but
call them test-only throughout: rename isEcho, drop 'preview'/'dry run' from the comments, and remove the
dry-run sentences from operations.md:991-992, 1285-1286 and 2057-2058.

### correctness

#### CR-0001 [blocker] [new] A VPN-scoped import deletes every virtual hostname on the target, silently

Location: internal/broker/importops.go:708-722, internal/broker/transform.go:602-646,
internal/broker/transform.go:663-672, internal/broker/importops.go:221-249,
internal/cli/exportconfig.go:117-129, internal/broker/importops.go:1072-1104,
internal/broker/diff.go:108-111, docs/operations.md:1506-1511
Confidence: verified
Evidence:

    importops.go:711 `if lines := TargetOnlyRemovals(blocks, p.Target); len(lines) > 0 {` runs for every
    import. Nothing checks the artifact's scope or region.
    transform.go:623-638: `names := target.VirtualHostnames` ... `have := (&Capture{Blocks:
    blocks}).nestedNames(verbVirtualHostname)` ... every target name the artifact does not create gets `
    no virtual-hostname %q`.
    transform.go:668: VirtualHostnames is read from the target's full `show current-config all`
    (readTarget, importops.go:535).
    importops.go:233-247: a `--vpn acme` export captures only `show current-config message-vpn "acme"`,
    so the artifact has no broker region and no `create virtual-hostname` line.
    Failure: the target has virtual hostname `orders.example.com` bound to VPN `orders`. The operator
    imports an artifact exported with `--vpn acme`. buildChunks adds the chunk `clear target-only
    objects` containing `service / no virtual-hostname "orders.example.com" / exit`, which runs before
    anything else. The hostname is gone.
    Nothing reports it. ImportPlan.Report (1072-1104) has no row for it, confirmImport
    (exportconfig.go:117-129) names only the VPN teardown, ImportResult.Report has no row for it, and
    ImportVerify's diff is directional (diff.go:108-111), so it cannot see a removal.

Why it matters: This deletes broker-level configuration that the artifact never described, and it touches
VPNs outside the import's scope. Clients that reach other VPNs through those SNI hostnames stop connecting.
The operator is not told at the plan, at the confirmation, in the result or in the verification. Even for
broker or total scope, the removals should be listed in the plan before they run.

Suggested patch (not applied):

```diff
--- a/internal/broker/importops.go
+++ b/internal/broker/importops.go
@@ -708,7 +708,9 @@ func (o *Ops) buildChunks(ctx context.Context, role config.Role, p *ImportPlan, blocks []Block, res *ImportResult) ([]chunk, error) {
 	// Objects the TARGET carries that the artifact does not re-create, removed before
 	// anything is applied. They have no create line to sit beside, so they cannot go
 	// through the block-level clear -- see TargetOnlyRemovals.
-	if lines := TargetOnlyRemovals(blocks, p.Target); len(lines) > 0 {
+	// Only an artifact that carries the broker level speaks for them: a VPN-scoped
+	// artifact says nothing about the target's virtual hostnames and must not clear them.
+	if lines := TargetOnlyRemovals(blocks, p.Target); len(lines) > 0 && hasRegion(blocks, RegionBroker) {
 		var b strings.Builder
 		for _, line := range preambleForApply(p.Capture.Preamble) {
 			b.WriteString(line)
@@ -1041,6 +1043,16 @@ func keepRegion(blocks []Block, keep Region) []Block {
 	}
 	return out
 }
 
+// hasRegion reports whether any block belongs to region.
+func hasRegion(blocks []Block, region Region) bool {
+	for _, b := range blocks {
+		if b.Region == region {
+			return true
+		}
+	}
+	return false
+}
+
 // describeScope renders the scope for the artifact's metadata.
 func describeScope(vpns []string, brokerOnly bool) string {
```

#### CR-0002 [blocker] [new] Seven kubernetes.* values reach the broker CR unquoted with no load-time grammar (YAML/--- injection)

Location: internal/render/render.go:100-102, internal/render/render.go:164-172,
internal/render/render.go:239-251, internal/render/render.go:410-419, internal/render/render.go:479-491,
internal/config/artifactvalues.go:269-303, internal/config/validate.go:764-794,
internal/k8s/deploy.go:29-47, internal/k8s/deploy.go:187, internal/k8s/cluster.go:189-197,
internal/config/artifactvalues_test.go:197-239, internal/render/render.go:239-258,
internal/render/render.go:407-419, internal/config/validate.go:785-794, internal/k8s/deploy.go:176-189
Confidence: verified
Evidence:

    CLAUDE.md: "A value written into an artifact as-is is held to its sink's grammar at load ... Either
    quote it at the sink or check it in internal/config/artifactvalues.go ... A new schema field that
    reaches an artifact does neither by default."
    These seven sinks do neither:
    render.go:102  fmt.Fprintf(&b, "    name: %s\n", c.K8s.ServiceAccount)
    render.go:167  fmt.Fprintf(&b, "    messagingNodeMemory: %s\n", c.K8s.MsgNode.Mem)
    render.go:171  fmt.Fprintf(&b, "    useStorageClass: %s\n", c.K8s.Storage.Class)
    render.go:249-250 "    - name: %s\n ... claimName: %s\n", role.Title(), claim
    render.go:418  fmt.Fprintf(b, "              - %s\n", ns)   (placement.antiAffinityNamespaces)
    render.go:485-490 "metallb...loadBalancerIPs: %s" lb.IP / "address-pool: %s" lb.IPPool
    On the config side, artifactvalues.go:281-286 checks only operator.cpu/mem and the two storage
    sizes. A grep of internal/config finds no check of ServiceAccount, MsgNode.Mem, Storage.Class
    (validate.go:770 only refuses it beside customVolumeMount), the customVolumeMount claim values
    (validate.go:790 checks non-empty only), LoadBalancer.IP/IPPool or AntiAffinityNS entries.
    Failure: an env file with `kubernetes: {storage: {class: "standard\n---\napiVersion:
    rbac.authorization.k8s.io/v1\nkind: ClusterRoleBinding ..."}}` decodes clean (yaml.v3 double-quoted
    escapes) and passes Validate. render.BrokerCR then emits a second document. deploy.go:38,46 pipes it
    to `kubectl apply -f -` (cluster.go:190), which applies every document in the stream. `broker
    remove` sends the same bytes to `kubectl delete -f -` (deploy.go:187), so whatever the injected
    document names is deleted. TestValidateK8sArtifactValues already runs artifactInjections (which
    includes "latest\n---\nkind: Secret"), but only against the five fields it lists.

Why it matters: This bypasses the stated sink-grammar invariant. A shared or untrusted env file can add
arbitrary manifests (RBAC, workloads) to what `broker deploy` applies with the operator's credentials, and
arbitrary deletions to `broker remove`. Per CLAUDE.md, docs/test.md's artifactvalues_test.go rows need
updating in the same change.

Suggested patch (not applied):

```diff
--- a/internal/config/artifactvalues.go
+++ b/internal/config/artifactvalues.go
@@ -281,6 +281,7 @@ func (c *Config) validateK8sArtifactValues() error {
 	for _, f := range []struct{ field, value string }{
 		{"kubernetes.operator.cpu", op.CPU},
 		{"kubernetes.operator.mem", op.Mem},
+		{"kubernetes.msgNode.mem", c.K8s.MsgNode.Mem},
 		{"kubernetes.storage.msgNodeSize", c.K8s.Storage.MsgNodeSize},
 		{"kubernetes.storage.monNodeSize", c.K8s.Storage.MonNodeSize},
 	} {
@@ -299,5 +300,38 @@ func (c *Config) validateK8sArtifactValues() error {
 			return err
 		}
 	}
+	// Object names the CR carries unquoted (render.BrokerCR), each held to the grammar
+	// the API gives that kind of name. storage.class also reaches `kubectl get sc
+	// <class>` as a positional (k8s/check.go), where a dash-led value is a flag.
+	for _, f := range []struct{ field, value string }{
+		{"kubernetes.storage.class", c.K8s.Storage.Class},
+		{"kubernetes.serviceAccount", c.K8s.ServiceAccount},
+		{"kubernetes.loadBalancer.ipPool", c.K8s.LoadBalancer.IPPool},
+	} {
+		if err := validDNSSubdomain(f.field, f.value); err != nil {
+			return err
+		}
+	}
+	for _, r := range []Role{Primary, Backup, Monitor} {
+		if claim, ok := c.K8s.Storage.CustomMountFor(r); ok {
+			if err := validDNSSubdomain("kubernetes.storage.customVolumeMount."+r.Word(), claim); err != nil {
+				return err
+			}
+		}
+	}
+	for i, ns := range c.K8s.Placement.AntiAffinityNS {
+		if err := validDNSLabel(fmt.Sprintf("kubernetes.placement.antiAffinityNamespaces[%d]", i), ns); err != nil {
+			return err
+		}
+	}
+	if ips := c.K8s.LoadBalancer.IP; ips != "" {
+		for _, ip := range strings.Split(ips, ",") {
+			if net.ParseIP(strings.TrimSpace(ip)) == nil {
+				return fmt.Errorf("kubernetes.loadBalancer.ip %q is not an IP address or a comma-separated list "+
+					"of them. It is written into the CR's Service annotations as-is, where anything else would be "+
+					"misread or restructure it", ips)
+			}
+		}
+	}
 	return nil
 }
--- a/internal/config/artifactvalues_test.go
+++ b/internal/config/artifactvalues_test.go
@@ -214,6 +214,14 @@ func TestValidateK8sArtifactValues(t *testing.T) {
 			[]string{"30Gi", "1Ti", "500G"}, []string{"30 Gi", "30GiB", "-30Gi"}},
 		{"kubernetes.storage.monNodeSize", func(c *Config, v string) { c.K8s.Storage.MonNodeSize = v },
 			[]string{"5Gi"}, []string{"5gb"}},
+		{"kubernetes.msgNode.mem", func(c *Config, v string) { c.K8s.MsgNode.Mem = v },
+			[]string{"6898Mi", "8Gi"}, []string{"8 Gi", "-8Gi"}},
+		{"kubernetes.storage.class", func(c *Config, v string) { c.K8s.Storage.Class = v },
+			[]string{"standard", "fast.ssd-1"}, []string{"Standard", "-A", "--kubeconfig=./evil"}},
+		{"kubernetes.serviceAccount", func(c *Config, v string) { c.K8s.ServiceAccount = v },
+			[]string{"solace-broker"}, []string{"Solace", "sa:x"}},
+		{"kubernetes.loadBalancer.ipPool", func(c *Config, v string) { c.K8s.LoadBalancer.IPPool = v },
+			[]string{"pool-a"}, []string{"pool a"}},
 	} {
 		for _, v := range f.good {
 			c := validK8sConfig()
```

#### CR-0003 [blocker] [new] kubernetes.storage.class reaches kubectl argv as a bare positional, so config text can inject kubectl flags

Location: internal/k8s/check.go:46-57, internal/k8s/check.go:78-81, internal/k8s/check.go:99-105,
internal/k8s/checkreport.go:404-435, internal/k8s/checkreport.go:621-645,
internal/config/validate.go:525-558, internal/config/validate.go:764-776
Confidence: verified
Evidence:

    check.go:79-80   if c.Cfg.K8s.Storage.Class != "" { return c.Cfg.K8s.Storage.Class, nil }
    check.go:100     out, err := c.kubectlOutput(ctx, "get", "sc", name, "-o",
    "custom-columns=V:"+field, "--no-headers")
    checkreport.go:432-433  if scope.wantsBroker() { c.storageRows(ctx, rep) }  (run by `validate` /
    `check deploy` once the API server is reachable)
    checkreport.go:627,640  name, err := c.resolveStorageClass(ctx) ... c.storageClassAttrs(ctx, name)
    The class has no grammar check anywhere in config (validate.go:770 only refuses it beside
    customVolumeMount). The project's own CheckSecretKeyRef comment (validate.go:527-533) names exactly
    this hazard: "a dash-led name or namespace is parsed by kubectl as a flag (`-A`, `--kubeconfig=/x`)
    ... It is 'config text must not choose what executes' applied to the arguments".
    Failure: an env file shipped in a repo next to a crafted kubeconfig sets `kubernetes.storage.class:
    --kubeconfig=./ops.kubeconfig`. The operator runs `solace-util validate -e that.yaml` to vet it.
    argv becomes `kubectl get sc --kubeconfig=./ops.kubeconfig -o custom-columns=... --no-headers`.
    pflag reads the token as a flag that overrides the operator's kubeconfig (last wins), and a
    `users[].user.exec` credential plugin in that file runs an arbitrary binary before the request.

Why it matters: This bypasses CLAUDE.md's first security invariant ("Config text must not be able to choose
what executes"). The execguard allowlist governs argv[0], and this hands argv[2] to the file author. The
load-time half is fixed by CR-0002's validDNSSubdomain on kubernetes.storage.class (a StorageClass name is a
DNS-1123 subdomain, so it cannot start with '-').

Change: Besides the load-time check in CR-0002's patch, re-check at the sink, because CLAUDE.md says
executors "must not assume config.Load ran". Export a config check, for example `func
CheckStorageClassName(v string) error { return validDNSSubdomain("kubernetes.storage.class", v) }` in
internal/config/validate.go, mirroring how CheckSecretKeyRef is enforced twice. In internal/k8s/check.go
resolveStorageClass, call it before returning c.Cfg.K8s.Storage.Class and return its error. Add a k8s test
with a recording runner asserting that a class of `--kubeconfig=./x` never reaches argv, and catalogue it in
docs/test.md.

#### CR-0004 [blocker] [new] WATCH_NAMESPACE read back from a running Deployment is rendered raw into the operator bundle an admin applies

Location: internal/k8s/watch.go:128-158, internal/k8s/watch.go:238-267, internal/k8s/watch.go:337-372,
internal/k8s/operator.go:65-83, internal/k8s/operator.go:247-290,
internal/k8s/assets/operator-1.4.2.yaml.tmpl:1929-1931, internal/config/artifactvalues.go:293-301
Confidence: verified
Evidence:

    watch.go:145-149  switch len(elsewhere) { case 0: return nil, "", nil; case 1: return elsewhere[0],
    "", nil }
    watch.go:244-245  if parsed := splitWatch(env.Value); len(parsed) > 0 { return parsed, false }
    (splitWatch only trims and de-duplicates)
    watch.go:365,371  merged, added := unionWatch(installed, desired) ... return watchPlan{value:
    strings.Join(merged, ",")}
    operator.go:258  manifest, err := RenderOperator(c.Cfg, opNS, plan.value)
    template:1929-1930  - name: WATCH_NAMESPACE / value: "{{.WatchNamespace}}"   (text/template, no
    escaping)
    artifactvalues.go:293-301 holds only the CONFIG entries to the DNS-label rule. Nothing checks
    entries that came from the cluster.
    Failure: an identity that may only create Deployments in its own namespace team-t creates a
    Deployment named pubsubplus-eventbroker-operator whose WATCH_NAMESPACE is: x, a double quote, a
    newline, ---, a ClusterRoleBinding of cluster-admin to team-t's ServiceAccount, then a trailing
    document that absorbs the rest of the template (written without commas, which splitWatch would split
    on). A cluster admin later runs the first `operator deploy`, with no operator in the resolved
    namespace. The cluster-wide list finds the decoy as the sole candidate elsewhere, reconcileWatch
    unions its entries, the quote closes the scalar, and operator.go:286 `kubectl apply -f -` applies
    the injected binding under the admin's identity. Without a newline, the decoy still decides which
    namespaces the new operator watches and which image the downgrade gate compares against.

Why it matters: Holding create on Deployments in any one namespace becomes cluster-admin through an admin's
routine operator deploy. This bypasses the sink-grammar invariant through data read back from the cluster.
The check belongs at the sink so that every caller, GenOperator included, is covered. CR-0009 closes the
decoy route, but the sink check is still needed.

Suggested patch (not applied):

```diff
--- a/internal/k8s/operator.go
+++ b/internal/k8s/operator.go
@@ -64,6 +64,16 @@
 // first (splitAfterNamespace), never a separately-invented one.
 func RenderOperator(cfg *config.Config, opNS, watch string) ([]byte, error) {
 	op := cfg.K8s.Operator
+	// watch may carry entries read back from a running operator (reconcileWatch), which no
+	// config check has seen, and the bundle writes it into a YAML scalar as-is. Each entry
+	// is a namespace, so it is held to the label rule kubernetes.namespace is.
+	for _, ns := range splitWatch(watch) {
+		if !kube.ValidLabel(ns) {
+			return nil, fmt.Errorf("refusing to render the operator bundle: %s entry %q is not a namespace "+
+				"name (a DNS-1123 label); nothing was applied -- check the installed operator's %s",
+				watchEnvVar, ns, watchEnvVar)
+		}
+	}
 	vars := operatorTmplVars{
 		Namespace:      opNS,
 		WatchNamespace: watch,
```

#### CR-0006 [major] [new] Namespace emptiness gate misses RBAC, NetworkPolicy, Ingress and CR objects; --no-prompt then deletes the namespace

Location: internal/cli/ops_k8s.go:790-823, internal/k8s/namespace.go:55-72, docs/operations.md:1270-1305,
internal/k8s/assets/operator-1.4.2.yaml.tmpl:9-23
Confidence: verified
Evidence:

    namespace.go:57  out, err := c.outKube(ctx, func(b []string) runner.Cmd { return nsgate.List(b, ns)
    })
    nsgate.List (shared, v0.1.2 nsgate.go:55) lists only: get
    all,persistentvolumeclaims,secrets,configmaps -n <ns> -o json
    ops_k8s.go:808-821  delete-namespaces probe, then confirmDelete(a, "the now-empty namespace "+ns...)
    (--no-prompt answers yes), then c.DeleteNamespace
    template:14-23  the PubSubPlusEventBroker CRD declares no `categories`, so the kind is not in `all`
    docs/operations.md:1282 ("The namespace is listed (`get
    all,persistentvolumeclaims,secrets,configmaps -o json`)") and the bullet after it: "Anything else
    blocks removal: a second Solace broker sharing the namespace blocks it"
    Failure: the namespace was provisioned with RoleBindings, NetworkPolicies, a ServiceAccount, an
    Ingress or a PodDisruptionBudget, or holds custom resources such as cert-manager objects, or a
    second PubSubPlusEventBroker CR that has no objects yet (operator stopped, or not watching this
    namespace). The `all` category lists none of these kinds, so Occupants comes back empty. `broker
    remove --no-prompt`, run by an identity with delete on namespaces, then cascade-deletes the
    namespace and all of it. Interactively, the prompt calls it "the now-empty namespace".

Why it matters: The documented promise that anything else blocks removal fails in the destructive direction.
The kind set lives in the shared nsgate package (see handoff_notes). This app chose to let --no-prompt
approve the namespace delete and to document the stronger guarantee.

Change: Short term, in internal/cli/ops_k8s.go removeNamespaceIfEmpty: stop letting --no-prompt approve the
namespace delete. Ask only at a terminal, as confirmScope does, and otherwise keep the namespace with a
warning. Also correct docs/operations.md 1282-1302 to name the kinds that are actually listed. Long term,
the shared nsgate.List should also list roles, rolebindings, serviceaccounts, networkpolicies, ingresses,
poddisruptionbudgets and pubsubpluseventbrokers, or derive its kinds from `api-resources --namespaced
--verbs=list`. Add a test in internal/cli where the listing holds only a RoleBinding and --no-prompt is set,
asserting the namespace is kept.

#### CR-0007 [major] [new] Primary-active switch gate is inverted when the local site is addressed through its backup node

Location: internal/cli/replication.go:142-162, internal/cli/replication.go:336-346,
internal/broker/matechannel.go:222-237, internal/broker/switchplan.go:126-162,
internal/cli/commands.go:515-527, internal/broker/verify_ops.go:177-186,
internal/broker/verify_local.go:163-170
Confidence: verified
Evidence:

    cli/replication.go:151-152: `chans := map[string]broker.MateChannel{ self.VirtualRouterName:
    o.NewLocalMate(role), ...` where role comes from --pod (336-341: `role, err := podRole(a) ...
    opPerformReplication(a, k8sOps(a), role, mateChannel)`); commands.go:526 `addPodFlag(c, app,
    config.K8s)`. On docker/podman 344-345 pass config.Primary, but matechannel.go:14-15 states the
    container transports "are node-local and ignore the role entirely", and unlike RedundancyCoordinated
    (verify_local.go:163-170) nothing checks this host is the primary.
    matechannel.go:233-236: PrimaryActive asks the ADDRESSED node `return activity(s,
    activityLocalActive) >= 1`. The codebase's own model (verify_ops.go:182-186): on the backup node
    `activity(bk, activityLocalActive) == 1` means "Backup node is active".
    help text commands.go:518: "Refuses unless both sites are on their primary HA node";
    switchplan.go:135-137: "When the backup is active BECAUSE the primary is down, the primary is the
    node this tool addresses -- so the run would fail anyway". Test cli/replication_test.go:1193-1213
    pins that --pod backup sends EVERY local call (show redundancy included) to pubsubplus-b-0.
    Failure: (1) k8s, primary pod down, backup holds activity. Operator runs `broker perform dr --pod
    backup` (the obvious move with the primary down). RouterName from b-0 matches routerNames [sg1,
    sg1b] (configuration.md:540 says list every node). b-0's Activity Status reads Local Active ->
    PrimaryActive true -> gate passes and the DR switch runs on a site that is on its backup. (2) Same
    on docker/podman run from the backup host. (3) Healthy group + --pod backup: b-0 shows only Mate
    Active -> refused with "the primary HA node is not active" and advice to revert activity that is
    already on the primary.

Why it matters: The one gate the help text promises (both sites on their primary) passes in exactly the
state it exists to refuse, and refuses the healthy state, whenever the local site is reached through its
backup -- which is how an operator reaches a site whose primary is down.

Change: Make the gate ask about the PRIMARY rather than the addressed node. Either (a) in
opPerformReplication refuse --pod other than primary on Kubernetes and, on docker/podman, resolve
o.LocalRole(a.pod) and refuse a non-primary host the way RedundancyCoordinated does
(verify_local.go:163-170); or (b) have cliMate.PrimaryActive read the node's own `Active-Standby Role`
(labelActiveStandbyRole) and return Local Active on a Primary, but (Mate Active and no Local Active) on a
Backup, refusing anything else. Fix the switchplan.go:135-137 comment and add a test where --pod backup and
the backup reports Local Active must be refused.

#### CR-0008 [major] [new] PrimaryActive reads Enabled-Released (or an unreadable report) as standalone and passes the gate

Location: internal/broker/matechannel.go:222-237, internal/broker/sempmate.go:317-334,
internal/broker/verify_ops.go:196-211, internal/broker/verify_local.go:230-245,
internal/broker/broker.go:308-322, docs/operations.md:1803-1807
Confidence: verified
Evidence:

    matechannel.go:233-235:
      if field(s, labelConfigStatus) != "Enabled" {
          return true, nil // standalone: no HA group, so this node is the only one
      }
    sempmate.go:330-332: `if !strings.EqualFold(strings.TrimSpace(r.ConfigStatus), "Enabled") { return
    true, nil }`.
    The tool's own model of a primary that released activity (verify_ops.go:205-207,
    verify_local.go:239-241): `field(out, labelConfigStatus) == "Enabled-Released" && field(out,
    labelRedundancyStatus) == "Down" && activity(out, activityMateActive) == 1` -- an HA group whose
    BACKUP holds activity.
    broker.go:311-321: field returns "" when the label is absent, and the local/kubectl reads go through
    readCLI with no rejection scan.
    Failure: primary was put into `redundancy release-activity` for maintenance -> backup active,
    primary reports Configuration Status "Enabled-Released" -> "Enabled-Released" != "Enabled" ->
    PrimaryActive returns true as if standalone -> requirePrimaryActive passes -> the switch runs on a
    site that is on its backup. Likewise any report the field scan cannot read (label missing, error
    transcript, trailing whitespace) is waved through as standalone.

Why it matters: docs/operations.md:1803-1807 promises the run refuses when a site is on its backup; an
admin-initiated failover (the most common way a site ends up there) and any unreadable report silently pass
instead, because the standalone branch is the fallback for everything that is not exactly "Enabled".

Suggested patch (not applied):

```diff
--- a/internal/broker/matechannel.go
+++ b/internal/broker/matechannel.go
@@ -230,7 +230,13 @@ func (c *cliMate) PrimaryActive(ctx context.Context) (bool, error) {
 		return false, fmt.Errorf("read redundancy from %s: %w", c.what, err)
 	}
 	s := string(out)
-	if field(s, labelConfigStatus) != "Enabled" {
+	status := strings.TrimSpace(field(s, labelConfigStatus))
+	if status == "" {
+		return false, fmt.Errorf("read redundancy from %s: `show redundancy` carries no %q line, so "+
+			"whether this site is on its primary cannot be told", c.what, labelConfigStatus)
+	}
+	// "Enabled-Released" is an HA group whose primary handed activity away, not a standalone.
+	if !strings.HasPrefix(status, "Enabled") {
 		return true, nil // standalone: no HA group, so this node is the only one
 	}
 	return activity(s, activityLocalActive) >= 1, nil
--- a/internal/broker/sempmate.go
+++ b/internal/broker/sempmate.go
@@ -330,3 +330,7 @@ func (m *sempMate) PrimaryActive(ctx context.Context) (bool, error) {
-	if !strings.EqualFold(strings.TrimSpace(r.ConfigStatus), "Enabled") {
+	status := strings.ToLower(strings.TrimSpace(r.ConfigStatus))
+	if status == "" {
+		return false, fmt.Errorf("%s: the `show redundancy` reply carries no config-status", m.what)
+	}
+	if !strings.HasPrefix(status, "enabled") {
 		return true, nil
 	}
```

#### CR-0009 [major] [new] An operator found in another namespace drives deploy and remove, which then write and delete in the resolved namespace

Location: internal/k8s/watch.go:90-97, internal/k8s/watch.go:128-158, internal/k8s/watch.go:394-431,
internal/k8s/operator.go:198-290, internal/k8s/operator.go:511-575,
internal/k8s/assets/operator-1.4.2.yaml.tmpl:1770-1895, internal/k8s/checkreport.go:493-499,
internal/cli/ops_k8s.go:590-593, docs/operations.md:1347-1360
Confidence: verified
Evidence:

    watch.go:148-149  case 1: return elsewhere[0], "", nil   (a sole operator outside the resolved
    namespace is returned as THE operator)
    operator.go:199,241,247  opNS := c.operatorNS(ctx) ... c.confirmNoDowngrade(dep) ...
    c.reconcileWatch(dep), then the bundle is applied into opNS (operator.go:286)
    watch.go:395,399,425  installedWatch(ctx) [the elsewhere dep] ... opNS := c.operatorNS(ctx) ...
    return c.OperatorDelete(ctx, deleteCRDs)
    operator.go:528,550  crds, _, rest := splitOperatorBundle(manifest) ... c.deleteStdin(ctx, rest)
    template:1770-1772,1885-1887  the cluster-scoped ClusterRole pubsubplus-eventbroker-operator-role
    and ClusterRoleBinding pubsubplus-eventbroker-operator-rolebinding are in `rest`
    checkreport.go:497-498  "`solace-util operator deploy` would install a second operator there"
    ops_k8s.go:591-592  the prompt names k8s.OperatorNamespace(a.Cfg), the resolved namespace
    Failure: env file A (operator.namespace: solace-operator) installed the operator to watch `solace`.
    Env file B describes the same broker but omits operator.namespace, so it resolves to
    pubsubplus-operator-system. `operator remove -e B` prompts about pubsubplus-operator-system.
    installedWatch returns the solace-operator Deployment, the subtraction leaves nothing, and
    OperatorDelete's `delete -f -` misses the namespaced documents (--ignore-not-found) but deletes the
    shared ClusterRole and ClusterRoleBinding by name. The running operator loses its cluster
    permissions while the tool reports success, and with --delete-crd and no brokers the CRDs go too.
    `operator deploy -e B` merges the other operator's watch list and installs a second operator,
    re-pointing the binding at the new ServiceAccount. docs/operations.md:1356-1360 calls the operator
    namespace "the only namespace a removal ever acts on", adding that "an operator installed elsewhere
    is not touched".

Why it matters: A removal breaks an operator that its prompt never named. A deploy knowingly creates the
second install that its own check report warns about. No test covers OperatorApply or OperatorRelease with a
sole operator elsewhere: watch_test.go:283 covers only the lookup.

Suggested patch (not applied):

```diff
--- a/internal/k8s/operator.go
+++ b/internal/k8s/operator.go
@@ -212,6 +212,9 @@ func (c *Cluster) OperatorApply(ctx context.Context) error {
 	dep, onlyNS, err := c.findOperatorDeployment(ctx)
 	if err != nil {
 		return fmt.Errorf("reading the installed operator (nothing was applied): %w", err)
 	}
+	if err := c.refuseOperatorElsewhere(ctx, dep); err != nil {
+		return err
+	}
 	// The cluster-wide list was refused and the operator namespace is empty: a first
 	// install by an identity that does not hold every permission the bundle's ClusterRole
--- a/internal/k8s/watch.go
+++ b/internal/k8s/watch.go
@@ -91,7 +91,10 @@ func (c *Cluster) installedWatch(ctx context.Context) (list []string, allNS, found bool, err error) {
 	dep, _, err := c.findOperatorDeployment(ctx)
 	if err != nil {
 		return nil, false, false, fmt.Errorf("reading the installed operator's watch scope: %w", err)
 	}
+	if err := c.refuseOperatorElsewhere(ctx, dep); err != nil {
+		return nil, false, false, err
+	}
 	list, allNS, found = watchFromDeployment(dep)
 	return list, allNS, found, nil
 }
@@ -217,3 +220,17 @@
 var errAmbiguousOperator = errors.New("cannot tell which Deployment is the operator")
 
+// refuseOperatorElsewhere stops a deploy or a removal whose operator Deployment was found
+// outside the namespace this env file resolves. Every write and delete targets the resolved
+// namespace, so acting on that reading would install a second operator -- re-pointing the
+// bundle's fixed-name ClusterRoleBinding -- or delete the cluster-scoped RBAC (and, with
+// --delete-crd, the CRDs) of the operator that is actually running.
+func (c *Cluster) refuseOperatorElsewhere(ctx context.Context, dep *deploymentItem) error {
+	opNS := c.operatorNS(ctx)
+	if dep == nil || dep.Metadata.Namespace == opNS {
+		return nil
+	}
+	return fmt.Errorf("the operator runs in namespace %s, but kubernetes.operator.namespace resolves to %s; "+
+		"nothing was changed. Set kubernetes.operator.namespace: %s", dep.Metadata.Namespace, opNS, dep.Metadata.Namespace)
+}
+
 // watchFromDeployment reads the watch scope off an already-fetched operator Deployment.
```

#### CR-0010 [major] [new] Create Usernames chunk cannot replay onto a target that already has a non-admin CLI user

Location: internal/broker/transform.go:205-232, internal/broker/transform.go:648-695,
internal/broker/importops.go:678-757, internal/broker/config_ops.go:223-230, docs/operations.md:497-503,
docs/operations.md:529-533, docs/operations.md:1539-1541
Confidence: verified
Evidence:

    transform.go:206 `var adminUsernameRE = regexp.MustCompile(`^(\s*)create username "admin" `)`. Only
    `admin` is rewritten from create into edit form.
    transform.go:655-661: TargetState carries Syslogs, ClientCAs, VirtualHostnames, DomainCAs and
    LDAPProfiles, but no usernames. PrepareForApply (684-695) has no clear or rewrite for any other
    `create username`.
    The repo states the broker behaviour three times: docs/operations.md:531 "The broker's `create
    username` fails if the user exists", config_ops.go:228, and transform.go:210-212 ("`create` is
    rejected as a duplicate").
    Order (importops.go:683-691, 745-757): teardowns, then create-VPN, then logging, then `sections:
    Create Usernames`, then broker sections, then VPN chunks.
    Failure 1: source and target were both deployed by this tool with `admin.additionalUsers: ops`,
    which is declared at boot (operations.md:497-503). The artifact carries `create username "ops" ...`.
    The broker rejects it and stop-on-error fails the chunk. The VPNs were already torn down and are not
    rebuilt.
    Failure 2: an import fails in any chunk after the usernames chunk. The documented recovery is a
    re-run (operations.md:1539-1541, "the chunks that already applied are applied again, which is
    safe"). The re-run's usernames chunk now hits the users the first run created and fails the same
    way, every time.

Why it matters: A normal migration between two brokers that carry the same CLI users, and the documented
re-run recovery, both end with the VPNs emptied and an import that cannot finish. The fix needs a
TargetState.Usernames read from the target capture, plus an edit-form rewrite (operands preserved) for every
artifact username the target already has. Removing the user first is not a fix: it would delete users the
operator's own tooling logs in with. The premise (the broker rejects a duplicate `create username`) is the
repo's own statement at internal/broker/transform.go:210-212 and docs/operations.md:531.

Change: Add `Usernames []string` to TargetState (transform.go:655-661), filled in readTargetState from the
target capture's `create username` blocks (Kind "username", Create). Then generalise RewriteAdminUsername
into a target-aware transform: for every artifact `create username "X"` where X is admin or is in
target.Usernames, rewrite the opener to the existing-user form while keeping its operands (see CR-0081). Do
not emit `no username`. Run it in PrepareForApply in place of RewriteAdminUsername, add transform_test.go
cases for a target that already has a non-admin user and for a re-run, and list them in docs/test.md.
Correct operations.md:1539-1541 if any chunk stays non-idempotent.

#### CR-0012 [major] [new] operator remove --delete-crd is silently not honoured (exit 0) whenever the operator is kept

Location: internal/cli/ops_k8s.go:611-630, internal/k8s/watch.go:394-431, docs/operations.md:1253-1257,
internal/k8s/watch_test.go:761-787
Confidence: verified
Evidence:

    ops_k8s.go:615-616 `case a.deleteLayer && brokersExist: deleteCRDs = true` and 622-624 asks the
    strict CRD question when no broker exists, then 630 `return c.OperatorRelease(bg(), deleteCRDs)`.
    watch.go only consults deleteCRDs on the full-delete branch (425 `return c.OperatorDelete(ctx,
    deleteCRDs)`). The other branches drop it:
      401-406 `!found` -> Warn("nothing to remove"); return nil
      408-414 `allNS` -> logf("operator kept..."); return nil
      428-430 narrowing -> Warn("...KEPT for those namespaces"); return c.setWatch(...)
    None mentions the CRDs. Failure: env file covers ns `a`, operator watches `a,b` (i.e. another team's
    broker lives in `b`). `operator remove --delete-crd` at a terminal: brokersExist -> deleteCRDs=true
    without asking -> narrowing branch -> watch narrowed, exit 0, CRDs untouched, nothing says so. With
    no brokers but the operator watching other namespaces, the operator is first asked to type 'yes' to
    delete the CRDs, then that yes is ignored. docs/operations.md:1255-1257 promises the opposite: "An
    explicit --delete-crd still runs into the refusal ... rather than being downgraded to 'kept': naming
    the flag earns a loud failure ..., not a silent no-op."
    TestOperatorReleaseLeavesAnOperatorThatWatchesEverything passes deleteCRDs=true and asserts nil, so
    the silent path is pinned rather than caught.

Why it matters: An explicitly requested and (sometimes) explicitly confirmed destructive layer is reported
as success while nothing happened -- the accepted-and-ignored shape this tree refuses elsewhere, and a
direct contradiction of the documented contract. Scripts that rely on exit 0 meaning the CRDs are gone will
be wrong.

Suggested patch (not applied):

```diff
--- a/internal/k8s/watch.go
+++ b/internal/k8s/watch.go
@@ -411,5 +411,9 @@ func (c *Cluster) OperatorRelease(ctx context.Context, deleteCRDs bool) error {
 		c.logf("operator kept. To remove it anyway, set kubernetes.operator.watchNamespaces to match "+
 			"what it watches, or delete it by hand: kubectl delete deployment %s -n %s",
 			operatorDeployment, opNS)
+		if deleteCRDs {
+			return fmt.Errorf("refusing --delete-crd: the operator watches ALL namespaces and stays installed, " +
+				"so its CRDs stay too.\n  Nothing was changed")
+		}
 		return nil
 	}
@@ -425,5 +429,10 @@ func (c *Cluster) OperatorRelease(ctx context.Context, deleteCRDs bool) error {
 		return c.OperatorDelete(ctx, deleteCRDs)
 	}
 
+	if deleteCRDs {
+		return fmt.Errorf("refusing --delete-crd: the operator also watches %s, so it stays installed and its "+
+			"CRDs cannot go with it.\n  Nothing was changed: re-run without --delete-crd to release this env "+
+			"file's namespaces", strings.Join(remaining, ","))
+	}
 	c.progress().Warn("the operator also watches %s, so it is KEPT for those namespaces.",
 		strings.Join(remaining, ","))
```

#### CR-0013 [major] [new] container.dataDir has no floor: '/' or '~' is accepted, chowned unprompted and emptied by --delete-data

Location: internal/config/validate.go:1543-1580, internal/config/hostpath.go:465-477,
internal/container/manager.go:417-443, internal/container/manager.go:1138-1156,
internal/cli/ops_container.go:74-85, internal/cli/confirm.go:55-62, internal/cli/confirm.go:164-171
Confidence: verified
Evidence:

    validate.go 1575: `if dir := c.ContainerBlock(p).DataDir; !IsAbsHostPath(dir) {` is the only shape
    check (CheckHostPath refuses charset only); hostpath.go 465-466: tildeHome expands 'exactly "~"' to
    the home directory.
    manager.go 419  Argv: []string{"mkdir", "-p", cb.DataDir}
    432  m.run(ctx, "unshare", "chmod", "775", cb.DataDir)
    438  m.run(ctx, "unshare", "chown", cb.RunUser, cb.DataDir)
    441  runner.Cmd{Argv: []string{"chown", cb.RunUser, cb.DataDir}}
    ops_container.go 81: PrepHost runs on every `broker deploy`, with no question.
    manager.go 1151: `find <dir> -mindepth 1 -delete`; confirm.go 60: the prompt reads 'Also delete
    persistent data? ...' and never shows the path.
    Step by step: `podman.container.dataDir: ~` (one character short of the sample's `~/solace/data`) on
    the rootless default -> expands to $HOME -> checkDataDir passes (home is writable) -> prep runs
    `podman unshare chmod 775 $HOME` and `podman unshare chown 1000:0 $HOME`, handing the home directory
    to a subuid and making it group-writable (sshd StrictModes then refuses key logins); a later `broker
    remove --delete-data` empties $HOME. `docker.container.dataDir: /` run as root -> `chown 1000001:0
    /` on every deploy.

Why it matters: The stated reason for accepting any absolute path ('the operator has to be the one who says
where their data lives', validate.go 1552-1553) does not cover values that cannot be anyone's data
directory, and the two destructive operations on it (an unprompted chown, an unscoped recursive delete) have
no other guard. CLAUDE.md 4a: 'a systemic misconfiguration fails the whole operation loudly at setup'.

Change: In validateContainer, after the IsAbsHostPath check, refuse a dataDir that is '/' (after trimming
trailing separators), has fewer than two path components ('/opt', '/home', 'C:/x'), or equals the home
directory the '~' expansion resolves to (c.home()), naming the field and the resolved path. Put the resolved
dataDir into the container --delete-data question (layerData is shared with k8s, so pass the path through
confirmLayer for containers). Add the refusal cases to the validate tests and docs/test.md, and a sentence
to docs/configuration.md's dataDir row.

#### CR-0015 [major] [new] convert silently empties ${VAR} references to unassigned variables and copies $(...)/backtick/${VAR:-x} literally

Location: internal/convert/parse.go:116-120, internal/convert/parse.go:138-143,
internal/convert/parse.go:362-391, internal/convert/parse.go:330-345, internal/convert/convert.go:63-94,
docs/configuration.md:743-750
Confidence: verified
Evidence:

    parse.go:119  refRE = `\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)`
    parse.go:364-372  "An unknown name expands to empty, exactly as bash would." ... return
    v.scalar[strings.Trim(ref, "${}")]
    parse.go:142  v.scalar[name] = v.expandSegments(firstTokenSegments(rhs))
    parse.go:335-336  whitespace ends the word (flushWord)
    No warning is emitted in either case. The docs promise: "A variable with no YAML equivalent is named
    on stderr, never dropped silently" and "Shell constructs beyond assignments are skipped"
    (configuration.md:745-749).
    Failures:
    (1) `SOLBK_DATA_DIR=${HOME}/solace/data` -> HOME is not assigned in the file, so the result is
    "/solace/data". That is absolute, so validateOutput passes and nothing is said. deploy bind-mounts
    /solace/data (an empty dir) instead of the real data, and `remove --delete-data` targets it. Bash
    took HOME from the environment.
    (2) `SOLBK_ADM_PASS=$(cat /etc/solace/adm)` -> the first word is "$(cat", refRE does not match, and
    adminPass becomes the literal "$(cat". It passes validateCredentialChars and becomes the broker's
    admin password.
    (3) `SOLBK_ADM_PASS=${ADM_PASS:-changeme}` -> refRE does not match the `:-` form, so the literal
    "${ADM_PASS:-changeme}" is carried over as the password.

Why it matters: The converter's purpose is a faithful migration with every loss named. Here a value comes
out wrong (a relocated data dir, a literal password) with no warning, and Load accepts the result.

Change: In internal/convert/parse.go, record per ASSIGNED variable name (never the value, since these may be
secrets): (a) each $VAR/${VAR} reference with no earlier scalar assignment in the file (expand currently
returns v.scalar[name] == ""), and (b) each non-literal segment that still contains "$(", a backtick or "${"
after refRE substitution (command substitution and ${VAR:-x}/${VAR:?x} forms). Have Convert append a warning
for each, e.g. "SOLBK_DATA_DIR references ${HOME}, which this file does not assign: bash read it from the
environment, convert wrote it as empty -- set the value by hand" and "SOLBK_ADM_PASS uses shell expansion
convert does not evaluate; its text was copied literally -- set it by hand (or use semp.adminPassEnv)". Add
convert_test cases for ${HOME}/x, $(cat f), `cat f` and ${X:-d}, and correct docs/configuration.md:743-749.

#### CR-0020 [minor] [new] kubernetes.ports accepts lowercase /tcp but renders it raw, and the CRD protocol enum is case-sensitive

Location: internal/config/validate.go:826-869, internal/render/render.go:215-223,
internal/render/render.go:1249-1263, internal/k8s/assets/operator-1.4.2.yaml.tmpl:263-270,
internal/config/config_test.go:1358-1377, internal/config/artifactvalues.go:139-146,
internal/config/validate.go:860-869, internal/render/render.go:1249-1256, internal/render/render.go:212-222,
internal/k8s/assets/operator-1.4.2.yaml.tmpl:1475-1482, internal/config/config_test.go:1358-1368,
solace-yaml-generator.html:726-737
Confidence: verified
Evidence:

    validate.go:865  switch strings.ToUpper(proto) { case "TCP", "UDP": ...  (case is folded only for
    the check)
    validate.go:837-838  "parsePort never folds case itself, so an operator's \"tcp\"/\"udp\" must still
    validate"
    render.go:1253  proto = rest[i+1:]          (raw)
    render.go:221   fmt.Fprintf(&b, "      protocol: %s\n", p.proto)
    operator-1.4.2.yaml.tmpl:263-270  protocol: ... enum: [TCP, UDP, SCTP]
    config_test.go:1367 pins "tcp-web=8008:9008/tcp" as valid.
    The same schema's <platform>.network.ports accepts ONLY lowercase (artifactvalues.go:140-141 `case
    "tcp", "udp"`), so lowercase is the natural spelling to copy across.
    Failure: kubernetes.ports: ["tcp-web=8008/tcp"] loads clean, the CR carries `protocol: tcp`, and
    `kubectl apply` is refused by CRD schema validation (Unsupported value "tcp") midway through `broker
    deploy`, after the Secrets were applied.

Why it matters: Validation accepts a value the downstream schema rejects, which defeats validateK8sPorts'
stated aim ("rejected here rather than reaching the CR malformed"). The validate.go:837-838 comment and a
render test need updating with the fix.

Suggested patch (not applied):

```diff
--- a/internal/render/render.go
+++ b/internal/render/render.go
@@ -1250,7 +1250,9 @@ func parsePort(entry string) portSpec {
 	name, rest := cut(entry, "=")
 	proto := "TCP"
 	if i := strings.LastIndex(rest, "/"); i >= 0 {
-		proto = rest[i+1:]
+		// The CRD's protocol enum (TCP|UDP|SCTP) is case-sensitive and config.validateK8sPorts
+		// accepts either case, so the CR always carries the upper-case spelling.
+		proto = strings.ToUpper(rest[i+1:])
 		rest = rest[:i]
 	}
 	container := rest
```

#### CR-0021 [minor] [new] replication via.kubernetes.namespace/name only checked non-empty, though they reach the mate's kubectl argv

Location: internal/config/validate.go:1167-1192, internal/config/validate.go:1225-1229,
internal/k8s/matechannel.go:56-98
Confidence: verified
Evidence:

    validate.go:1178-1184  if strings.TrimSpace(k.Namespace) == "" {...}  if strings.TrimSpace(k.Name)
    == "" {...}  (no grammar check)
    matechannel.go:60  pod := podNameFor(k.Name, config.Primary)
    matechannel.go:64  opts := kube.ExecOptions{Namespace: k.Namespace}
    matechannel.go:69,88,98  kube.Upload / kube.Exec(base, pod, opts, ...)
    shared kube.go:219  args = append(append(args, pod), nsArgs(o.Namespace)...)   (pod is a bare
    positional)
    Contrast validate.go:1225-1229: the sibling passSecret goes through CheckSecretKeyRef because "a
    dash-led name ... is parsed by kubectl as a flag". kubernetes.name/namespace get validDNSLabel plus
    the pod-suffix length bound (validate.go:645-666).
    Failure: `via.kubernetes.name: Solace` or a 60-character name loads clean and fails only during
    `perform dr`, as an opaque kubectl error against pod "Solace-pubsubplus-p-0". A dash-led name
    ("--as=x") reaches kubectl as a flag. The fixed suffix and the then-missing pod make the exec fail
    before any request, so this is a robustness gap rather than an exploit.

Why it matters: A typo surfaces mid-switchover instead of at load, and the argument-grammar rule the file
applies to passSecret is skipped here. A length bound against BrokerPodSuffixShape (as for kubernetes.name)
would also be worth adding.

Suggested patch (not applied):

```diff
--- a/internal/config/validate.go
+++ b/internal/config/validate.go
@@ -1181,5 +1181,14 @@ func validateReplVia(field string, v ReplVia) error {
 		if strings.TrimSpace(k.Name) == "" {
 			return fmt.Errorf("%s.via.kubernetes.name must be set: the mate's PubSubPlusEventBroker "+
 				"name, which its pod is named after", field)
 		}
+		// Both reach the mate's kubectl argv -- the namespace after -n, the name as the pod
+		// positional <name>-pubsubplus-p-0 -- so they get kubernetes.namespace's and
+		// kubernetes.name's rule, the way passSecret gets CheckSecretKeyRef.
+		if err := validDNSLabel(field+".via.kubernetes.namespace", k.Namespace); err != nil {
+			return err
+		}
+		if err := validDNSLabel(field+".via.kubernetes.name", k.Name); err != nil {
+			return err
+		}
 	case v.SEMP != nil:
```

#### CR-0022 [minor] [new] Standalone routername filled from os.Hostname after Validate, bypassing validIdent before quadlet HostName=

Location: internal/config/roles.go:120-138, internal/cli/context.go:182-230,
internal/config/validate.go:1592-1602, internal/render/render.go:893-904, internal/cli/context.go:162-167,
internal/cli/context.go:196-211, internal/config/roles.go:132-138, internal/config/validate.go:579-587,
internal/config/validate.go:1592-1601, internal/render/render.go:896-903
Confidence: verified
Evidence:

    roles.go:133-137  if c.RedundancyEnabled() || c.Redundancy.Primary.Name != "" || hostname == "" {
    return false }; c.Redundancy.Primary.Name = hostname
    context.go:200,208  name, err := a.hostname() ... a.Cfg.FillStandaloneNodeName(name)   (runs after
    Load/Validate)
    validate.go:1595  validIdent("redundancy.primary.name", ...) ran on the EMPTY value, which
    validIdent skips.
    render.go:896  fmt.Fprintf(&b, "Description=Solace PubSub+ Event Broker (%s, nodetype=%s)\n",
    id.Hostname, ...)
    render.go:903  fmt.Fprintf(&b, "HostName=%s\n", id.Hostname)      (no EscapePercent, unquoted)
    Failure: a host whose name is outside identRE (for example containing '%', which systemd expands as
    a specifier, or another character the env-file value would be refused for) gets it written raw into
    the quadlet unit. The routername also never passes the check an env-file value of the same field
    would.

Why it matters: The value is root-controlled, so the impact is low. Still, the sink-grammar rule ("held to
its sink's grammar at load") has a path around it: a value written into an artifact enters the Config after
the gate.

Change: Hold the hostname to the same rule before filling it: change config.FillStandaloneNodeName to return
(bool, error) and refuse a hostname that fails validIdent("redundancy.primary.name", hostname). In
cli.fillStandaloneNodeName, fail the command with that error plus "set redundancy.primary.name in the env
file" rather than filling it. Add roles/cli tests for a hostname carrying '%' and a newline.

#### CR-0023 [minor] [new] bannerType scans the whole capture, so configuration text can decide the target's broker type

Location: internal/broker/importops.go:63-72, internal/broker/importops.go:266-284,
internal/broker/importops.go:534-543, internal/broker/importops.go:480-498,
internal/broker/sections.go:427-431
Confidence: verified
Evidence:

    importops.go:278 `typ := bannerType(out)` is applied to the full `cli -Apes` output, which is banner
    + transcript + the whole `show current-config all`.
    importops.go:65-69: `case strings.Contains(s, "license-software"), strings.Contains(s, "Event Broker
    Software"): return BrokerSoftware` is checked FIRST.
    The configuration carries free operator text: descriptions, and the console login banner
    (sections.go:427-431, "Console timeout and login banner").
    Failure: an APPLIANCE target's login banner or any description mentions "Event Broker Software" or
    "license-software". readTarget returns BrokerSoftware for it, checkSameType (480-498) passes for a
    software artifact, and the cross-type import the refusal exists to stop proceeds.

Why it matters: The cross-type guard can be satisfied by configuration text instead of the banner. The
comment at 276-277 says the banner is what is read; the code reads everything.

Suggested patch (not applied):

```diff
--- a/internal/broker/importops.go
+++ b/internal/broker/importops.go
@@ -63,5 +63,11 @@ var applianceModelRE = regexp.MustCompile(`Solace PubSub\+ [0-9]{3,5} Version`)
 func bannerType(out []byte) BrokerType {
 	s := string(out)
+	// Only the banner ahead of the capture votes: everything after ` ! BEGIN` is the
+	// broker's configuration, which carries operator text (descriptions, the console
+	// login banner) that can spell either marker.
+	if i := strings.Index(s, "! BEGIN"); i >= 0 {
+		s = s[:i]
+	}
 	switch {
 	case strings.Contains(s, "license-software"), strings.Contains(s, "Event Broker Software"):
 		return BrokerSoftware
```

#### CR-0024 [minor] [new] promoteConfirmed reads once for the whole phase and promotes anything not reading active

Location: internal/broker/switchplan.go:396-443, internal/broker/switchplan.go:21-27,
docs/operations.md:1827-1834
Confidence: verified
Evidence:

    switchplan.go:396-397: "promoteConfirmed re-reads immediately before each promotion and promotes
    only what is still standby everywhere." Code: 411-418 read every site ONCE before the loop; 422 `if
    st[act.VPN].Role == RoleActive {` is the only test. File comment 25-26: "The re-read narrows that
    window to one read and one write".
    Failure: (1) with N moving VPNs the Nth promotion acts on a read that is N-1 SetRole round trips old
    (each a WAN kubectl exec or curl), not one read and one write. (2) a VPN missing from the fresh read
    (zero VPNRepl, Role ""), or one whose replication a concurrent `configure dr` just shut down (Admin
    shutdown, Role standby), passes the guard and is promoted, although the doc says only what is "still
    standby everywhere" is.

Why it matters: The documented safety property is stronger than the code; the absent-VPN case in particular
treats missing evidence as permission. Confirm with a test whose pre-promotion read omits the VPN at the
demoted site.

Suggested patch (not applied):

```diff
--- a/internal/broker/switchplan.go
+++ b/internal/broker/switchplan.go
@@ -394,5 +394,6 @@ func confirmDemotions(ctx context.Context, chans map[string]MateChannel, plan SwitchPlan,
 }
 
-// promoteConfirmed re-reads immediately before each promotion and promotes only what is
-// still standby everywhere.
+// promoteConfirmed re-reads both sites ONCE, immediately before the promotion phase, and
+// promotes only what that read shows still standby, with replication enabled, at every
+// site. A VPN missing from the read is not evidence that it is standby.
 //
@@ -421,11 +422,11 @@ func promoteConfirmed(ctx context.Context, chans map[string]MateChannel,
 		for site, st := range fresh {
-			if st[act.VPN].Role == RoleActive {
+			if cur, ok := st[act.VPN]; !ok || cur.Admin != AdminEnabled || cur.Role != RoleStandby {
 				moved = append(moved, site)
 			}
 		}
 		if len(moved) > 0 {
 			sort.Strings(moved)
-			return fmt.Errorf("message-vpn %q is active at %s again since the demotion was confirmed: "+
+			return fmt.Errorf("message-vpn %q is no longer standby at %s since the demotion was confirmed: "+
 				"something outside this run moved it. Refusing to promote it at %s -- that is how two "+
 				"actives happen. Nothing further has been changed",
 				act.VPN, strings.Join(moved, " and "), act.Site)
```

#### CR-0025 [minor] [new] configure dr gate never mentions that unlisted replicating VPNs are shut down when the mate already matches

Location: internal/cli/replication.go:86-103, internal/broker/replicationops.go:572-577,
internal/broker/replicationops.go:619-630
Confidence: verified
Evidence:

    replicationops.go:572-575: "UNLISTED-BUT-ENABLED means shut down ... That is destructive to message
    flow, which is why the command takes an exact-`yes` gate naming those VPNs before any of it runs."
    The gate (cli/replication.go:93-99) says only "if this broker's mate configuration does not already
    match the env file, REPLICATION is stopped on EVERY VPN ... and restored only for the VPNs below"
    and lists the LISTED VPNs. replicationVPNLines 619-630 shuts down every unlisted admin-enabled VPN
    whenever mateStopped is false -- the ordinary re-run where the mate already matches. (Area note:
    this is an interfaces/confirmation-text issue reported here because the comment claims a safety
    property the gate does not have.)
    Failure: operator re-runs configure on a converged pair after adding a VPN on the broker but
    forgetting the file -> prompt says nothing stops unless the mate differs -> `yes` -> that VPN's
    replication is shut down and stays down; its DR copy goes stale until a failover reveals it.

Why it matters: The exact-yes gate is the only point the operator is warned before replication to DR is
stopped for good, and in the common case it does not mention it while the code comment says it names those
VPNs.

Suggested patch (not applied):

```diff
--- a/internal/cli/replication.go
+++ b/internal/cli/replication.go
@@ -95,3 +95,5 @@ func confirmReplicationConfig(a *App, self, mate config.ReplSite) (bool, error) {
 		"themselves keep running and clients stay connected -- including ones " +
 		"this file does not list -- and restored only for the VPNs below")
+	step("whether or not it does, every VPN this broker is replicating that the file does " +
+		"NOT list has its replication shut down, and it stays down")
 	for _, v := range names {
```

#### CR-0026 [minor] [new] The kube context the prompt names is not pinned, so the deletes run against whatever context is current when each kubectl call starts

Location: internal/cli/context.go:253-266, internal/cli/ops_k8s.go:709-737, internal/cli/ops_k8s.go:835-851,
internal/k8s/cluster.go:123-131, docs/operations.md:1403-1416
Confidence: verified
Evidence:

    context.go:261-263  out, err := engine.Output(bg(), a.Runner, runner.Cmd{Argv: kc.Argv("config",
    "current-context")}) ... a.kubeContext = ctx   (stored for display only)
    ops_k8s.go:715  confirmDelete(a, k8sWhat(a, "broker "+a.Cfg.K8s.Name))  -> "... in namespace <ns>
    (context \"dev\")? [y/N]"
    ops_k8s.go:849-850  return fmt.Sprintf(" (context %q)", a.kubeContext)
    cluster.go:131  engine.Run(ctx, c.R, runner.Cmd{Argv: k.Argv(args...)})   (no --context; each
    kubectl process re-reads the kubeconfig current-context)
    docs/operations.md:1403  "every one of those prompts also says where it will act"
    Failure: `broker remove --delete-data` resolves the context "dev" and then waits at the [y/N]
    prompt, with no time limit. Meanwhile the operator runs `kubectl config use-context prod` (or
    kubectx) in another terminal, which rewrites the shared kubeconfig. They come back and answer yes,
    the layer question gets "yes", and PreflightRemoval, DeleteBroker (CR and PVCs) and DeleteSecrets
    all run against prod.

Why it matters: The prompt's context clause is the tool's stated defence against acting on the wrong
cluster, but nothing enforces it between the question and the action. The window includes unbounded human
think time, and the destructive path deletes the broker and its data.

Change: Pin the resolved context. After announceKubeContext resolves a non-empty context, and when
kubernetes.command does not already carry --context or --kubeconfig, add `--context <ctx>` to the guarded
cluster argv for the rest of the run. One way is an App-held override that Cluster.base/cmd,
kubectlTransport.base, readSecretKey and NewMateChannel's own-cluster path all apply. That makes every call
land in the context the prompt named. Add a test in internal/cli where the runner answers current-context
"dev", and assert that every later argv carries --context dev.

#### CR-0027 [minor] [new] The CRD cascade gate reads an empty answer as "no brokers" and deletes the CRDs

Location: internal/k8s/operator.go:437-478, internal/k8s/inspect.go:60-83,
internal/k8s/watch_test.go:822-840, internal/k8s/namespace.go:55-66
Confidence: verified
Evidence:

    operator.go:439-441  if err := c.getJSON(ctx, &list, getRead{...AllNamespaces...}); err != nil {
    return nil, ... }
    inspect.go:71-74  if len(strings.TrimSpace(string(raw))) == 0 { ... return stderr, nil }   (any
    runner, not only Echo)
    operator.go:469-470  if len(refs) == 0 { return nil }   -> OperatorDelete proceeds to
    c.deleteStdin(ctx, crds)
    watch_test.go:824-827  "the repeat of it right before they go reads empty too" (the test relies on
    an empty read passing the gate)
    Contrast: namespace.go:63-66 and nsgate.Occupants treat an empty listing as unknown and keep the
    namespace. operator.go:461-463 says "being unable to see what would be destroyed is not permission
    to destroy it".
    Failure: a kubernetes.command wrapper approved with --allow-command (or any non-Echo runner) exits 0
    with empty stdout for `get pubsubpluseventbrokers --all-namespaces -o json`. Both gate passes see
    zero brokers, and `operator remove --delete-crd` deletes the CRDs, cascading to every broker in the
    cluster.

Why it matters: The one gate in front of a cluster-wide cascade fails open on "no answer", unlike the
namespace gate and rbac.Check, which refuse an exit-0 answer they cannot read. A real kubectl list always
carries "items", so the fix costs nothing on the normal path. The test at watch_test.go:827 needs a third
queued `{"items":[]}`.

Suggested patch (not applied):

```diff
--- a/internal/k8s/operator.go
+++ b/internal/k8s/operator.go
@@ -439,4 +439,10 @@ func (c *Cluster) BrokerCRs(ctx context.Context) ([]string, error) {
 	if err := c.getJSON(ctx, &list, getRead{kube.GetOptions{AllNamespaces: true}, []string{"pubsubpluseventbrokers"}}); err != nil {
 		return nil, fmt.Errorf("listing broker custom resources across the cluster: %w", err)
 	}
+	// A real list always carries "items", even an empty one. Nothing at all is no answer --
+	// the preview aside -- and the CRD cascade gate must not read it as "no brokers".
+	if list.Items == nil && !c.isEcho() {
+		return nil, fmt.Errorf("listing broker custom resources across the cluster: the command answered " +
+			"nothing, so what exists is unknown")
+	}
 	refs := make([]string, 0, len(list.Items))
```

#### CR-0028 [minor] [new] `status broker --all` shows [ OK ] ready when ServiceReady or HAReady is unreported or Unknown

Location: internal/k8s/statusreport.go:471-510, internal/k8s/inspect.go:343-365,
internal/output/output.go:110-118
Confidence: verified
Evidence:

    output.go:113-117  LevelInfo = iota; LevelOK; LevelWarn; LevelSkip; LevelFail   (Info sorts below
    OK)
    inspect.go:345-346  a condition not published -> output.LevelInfo, "not reported by the operator"
    inspect.go:362-363  any status other than True/False (e.g. "Unknown") -> output.LevelInfo
    statusreport.go:491-500  lvl := output.LevelOK; detailText := "ready"; if !anyKnownCondition(...)
    {...} else { for ... if row.Level > lvl { ... } }
    Failure: the CR carries NoWarnings=True and ServiceReady=Unknown, or no ServiceReady at all.
    anyKnownCondition is true, the Info rows never exceed the OK seed, and the listing prints [ OK ]
    ready for a broker whose service readiness nobody reported.

Why it matters: The listing claims health without evidence, which is the case the M8 comment at
statusreport.go:478-490 set out to stop. It only closed the case where no conditions are reported at all.

Suggested patch (not applied):

```diff
--- a/internal/k8s/statusreport.go
+++ b/internal/k8s/statusreport.go
@@ -491,11 +491,21 @@ func (c *Cluster) ClusterReport(ctx context.Context, detail bool) error {
 		lvl := output.LevelOK
 		detailText := "ready"
 		if !anyKnownCondition(it.Status.Conditions) {
 			lvl, detailText = output.LevelInfo, "not yet reported by the operator"
 		} else {
+			// A row the operator did not report (or reported as neither True nor False) is
+			// LevelInfo, which sits below the OK seed and so never raised it: one known
+			// condition was enough to read "ready" with ServiceReady unreported.
+			unreported := ""
 			for _, row := range brokerConditionRows(it.Status.Conditions, ha) {
-				if row.Level > lvl {
+				switch {
+				case row.Level == output.LevelInfo && unreported == "":
+					unreported = row.Key + ": " + row.Value
+				case row.Level > lvl:
 					lvl, detailText = row.Level, row.Key+": "+row.Value
 				}
 			}
+			if lvl == output.LevelOK && unreported != "" {
+				lvl, detailText = output.LevelInfo, unreported
+			}
 		}
```

#### CR-0029 [minor] [new] runUser uid 0 is accepted on docker and rootful podman, running the broker as host root

Location: internal/config/validate.go:1625-1628, internal/config/validate.go:412-415,
docs/container-security.md:143-146, docs/operations.md:706-709, internal/render/render.go:1075,
internal/render/render.go:904-907
Confidence: verified
Evidence:

    Posture finding (the containers checklist files it under build); reported here as the closest of my
    areas. container-security.md 143-145: 'Two corollaries worth enforcing rather than documenting.
    First, **refuse uid 0 where the engine is privileged** -- accepting it means accepting host root,
    and a format check on `uid[:gid]` will not catch it.'
    validate.go 1625  if u := cb.RunUser; u != "" && !runUserRE.MatchString(u) {
    validate.go 415   var runUserRE = regexp.MustCompile(`^[A-Za-z0-9._-]+(:[A-Za-z0-9._-]+)?$`)
    render.go 1075   fmt.Fprintf(&b, "    user: %s\n", compose.Quote(cb.RunUser))
    render.go 904    fmt.Fprintf(&b, "User=%s\n", uid)
    Step by step: `docker.container.runUser: "0:0"` (or `root`, `0`) passes runUserRE, nothing else in
    config/validate or internal/container checks the uid, the compose file carries `user: "0:0"`, and
    dockerd starts the broker as host uid 0. operations.md 708 says exactly that ('container uid 0 would
    be **host root**') but only as the reason for the default.

Why it matters: A single env-file value silently widens the posture the project's own security document says
must be enforced, and nothing in the artifact tests can see it because the rendered value is whatever the
operator wrote.

Suggested patch (not applied):

```diff
--- a/internal/config/validate.go
+++ b/internal/config/validate.go
@@ -1625,5 +1625,15 @@ func (c *Config) validateContainer(p Platform) error {
 	if u := cb.RunUser; u != "" && !runUserRE.MatchString(u) {
 		return fmt.Errorf("%s.container.runUser %q is invalid: expected uid[:gid] using only letters, digits, '.', '_' and '-'",
 			platformKey(p), u)
 	}
+	// container-security.md rule 3: docker and rootful podman map container uids straight
+	// onto host uids, so uid 0 is host root. Rootless 0:0 is only the invoking user.
+	if p == Docker || (p == Podman && c.Podman.Rootful) {
+		uid, _, _ := strings.Cut(cb.RunUser, ":")
+		if n, err := strconv.Atoi(uid); uid == "root" || (err == nil && n == 0) {
+			return fmt.Errorf("%s.container.runUser %q runs the broker as uid 0, which on this engine is host "+
+				"root; use the image's own %q, or deploy rootless podman, where 0:0 is only the invoking user",
+				platformKey(p), cb.RunUser, ImageRunUser)
+		}
+	}
 	if m := cb.Mem; m != "" && !containerMemRE.MatchString(m) {
```

#### CR-0030 [minor] [new] dataDir reaches the quadlet Volume= line unescaped: '%' is a systemd specifier, ':' splits the mount

Location: internal/render/render.go:961-974, internal/render/render.go:985, internal/render/render.go:1150,
internal/config/hostpath.go:41-52, internal/config/hostpath.go:97-101, internal/config/validate.go:1575-1580
Confidence: verified
Evidence:

    CLAUDE.md: 'A value written into an artifact as-is is held to its sink's grammar at load ... Either
    quote it at the sink or check it in internal/config/artifactvalues.go'.
    render.go 985   fmt.Fprintf(&b, "Volume=%s:%s:Z\n", cb.DataDir, dataMount)   // no EscapePercent,
    unlike HealthCmd/HealthInterval at 970-974
    hostpath.go 41-46: 'a colon in a host path is NOT harmless ... both container artifacts write a
    mount as `source:target:options` ... it is handled where the mount line is written' -- it is not:
    985 and 1150 write it raw.
    hostpath.go 97: refusal set is execguard.UnsafeChars minus '\\' = "'`$;|&<>()*?[]{}#!" plus quotes
    -- neither '%' nor ':'.
    Step by step: dataDir `/srv/solace/%h` loads clean; prep's mkdir/chown and --delete-data's find act
    on the literal '%h' directory, while systemd expands %h in the quadlet Volume= to the unit user's
    home and the broker mounts that instead; an unknown specifier makes systemd reject the unit. A ':'
    anywhere in dataDir turns the mount into a different source/target on both engines, failing the
    start after prep already created and chowned the path.

Why it matters: Not a key injection (newline is refused), but the one free-text value in the unit escapes
the sink grammar the project requires, and the source comment claims a handling that does not exist.

Suggested patch (not applied):

```diff
--- a/internal/render/render.go
+++ b/internal/render/render.go
@@ -984,3 +984,3 @@ func Quadlet(c *config.Config, id config.NodeIdentity) []byte {
 	}
-	fmt.Fprintf(&b, "Volume=%s:%s:Z\n", cb.DataDir, dataMount)
+	fmt.Fprintf(&b, "Volume=%s:%s:Z\n", quadlet.EscapePercent(cb.DataDir), dataMount)
 	for _, pair := range EnvPairs(c, id) {
--- a/internal/config/validate.go
+++ b/internal/config/validate.go
@@ -1579,4 +1579,8 @@ func (c *Config) validateContainer(p Platform) error {
 			"from (e.g. /opt/solace/data)", dataKey, dir)
 	}
+	if dir := c.ContainerBlock(p).DataDir; strings.Contains(dir, ":") {
+		return fmt.Errorf("%s %q contains ':': both artifacts write the mount as source:target:options, so a "+
+			"colon in the source splits it into a different mount -- choose a path without one", dataKey, dir)
+	}
 
 	// These reach the compose/quadlet artifact in structural positions, so they are
```

#### CR-0031 [minor] [new] broker deploy mutates the host before the redeploy question; 'a no leaves the host exactly as it was' does not hold

Location: internal/cli/ops_container.go:74-85, internal/container/manager.go:399-449,
internal/container/limits.go:156-168, internal/container/rootless.go:361-375, docs/operations.md:935-939,
docs/operations.md:738-741, internal/container/manager.go:487-489, internal/container/manager.go:763-766,
internal/cli/commands.go:118-122, docs/operations.md:970-974
Confidence: verified
Evidence:

    ops_container.go 81-84: `if err := m.PrepHost(ctx); err != nil {` ... `return m.Deploy(ctx, role)`
    -- PrepHost runs before Deploy's compare/ask.
    manager.go 413 checkLimits(ctx, true) -> limits.go 159 `sysctl -w fs.nr_open=...` and 164 writes
    /etc/sysctl.d (as root); rootless.go 364 `loginctl enable-linger`; manager.go 432/438/441
    chmod/chown of dataDir to cb.RunUser (no -R); 448 registryLogin.
    operations.md 938-939: 'A "no" leaves the host exactly as it was -- no file written, no secret
    stored, no container touched'; 738-740: 'broker deploy re-chowns it on every run, so a redeploy
    heals it'.
    Step by step: change runUser on a running deployment and run `broker deploy`; the data directory is
    chowned to the new uid at 441 before the question; answer n -> the running broker (old uid) now has
    a top-level data directory it does not own. Answer y -> only the top-level directory changes owner
    (no -R), so what the broker wrote under the old uid keeps it, and 'heals it' holds only if the
    image's gid-0 group bits cover it (unverified).

Why it matters: The documented contract is stronger than the code; the runUser change is the case where the
gap is observable. The ownership part needs a host to confirm (deploy, let the broker write, change
runUser's uid, deploy, `find <dataDir> ! -user <new>`).

Change: Either run PrepHost's chown/chmod only after Deploy's approve (or skip it when the directory is
already owned by runUser) or narrow operations.md 938-939 to 'no artifact written, no secret stored, no
container touched' and list what prep does first; and either chown -R when the owner differs from runUser or
drop 'so a redeploy heals it' from operations.md 739.

#### CR-0032 [minor] [new] Bare `auto-complete` is a command-line error that exits 1, not 2

Location: internal/cli/completion.go:48-50, internal/cli/exit.go:8-19, internal/cli/root.go:32
Confidence: verified
Evidence:

    completion.go:48-50:
      RunE: func(cmd *cobra.Command, _ []string) error {
          return fmt.Errorf("%s needs a shell: bash, zsh, fish, or powershell", cmd.CommandPath())
    The error carries no usage marker, so exitcode.Code returns 1. `auto-complete tcsh` goes through the
    NoArgs validator that markUsageArgs wraps and exits 2. root.go:32 promises "Exit status: 0 worked, 2
    bad command line or env file, 1 anything else"; exit.go:13-19 documents only the mistyped top-level
    command as a known gap. TestCompletionNeedsAShell checks err != nil but not the code.

Why it matters: Two spellings of the same mistake exit differently, and a script cannot tell 'forgot the
shell' from a runtime failure.

Suggested patch (not applied):

```diff
--- a/internal/cli/completion.go
+++ b/internal/cli/completion.go
@@ -7,6 +7,7 @@ import (
 	"path/filepath"
 	"strings"
 
+	"github.com/hafio/go-deploy-utility/exitcode"
 	"github.com/spf13/cobra"
 
 	"solace/internal/config"
@@ -46,7 +47,7 @@ func newCompletionCmd() *cobra.Command {
 		// instead, and makes the bare command say what to pass. `--help` is unaffected:
 		// the help flag short-circuits ahead of both.
 		RunE: func(cmd *cobra.Command, _ []string) error {
-			return fmt.Errorf("%s needs a shell: bash, zsh, fish, or powershell", cmd.CommandPath())
+			return exitcode.Usagef("%s needs a shell: bash, zsh, fish, or powershell", cmd.CommandPath())
 		},
 	}
 	c.AddCommand(
```

#### CR-0033 [minor] [new] convert prints every secret to stdout, breaking the one-place secret-display invariant

Location: internal/cli/convert.go:26-28, internal/cli/convert.go:73-79
Confidence: verified
Evidence:

    CLAUDE.md: "Secret values are shown in exactly one place -- `generate` on Kubernetes, where the
    Secret manifest *is* the artifact. Everywhere else reports say `set`/`MISSING`". convert.go:26 "The
    output carries every secret from the source file verbatim"; 73-78: with no -o it only warns (`prefer
    -o <file> to a shared terminal`) and then emitOrWrite sends the YAML -- admin password, PSK,
    registry password -- to stdout (terminal scrollback, CI logs, dev-script captures). Reported here
    because the slot asked for every path a secret reaches output; it is a security-invariant drift
    rather than a logic bug.

Why it matters: Either the invariant is wrong or convert is; as written the house rule does not hold and the
only guard is a warning that is easy to miss in a pipeline.

Change: Prefer refusing stdout when the converted result carries any secret value unless -o is given (exit 2
naming -o), or rewrite secrets to their *Env sibling keys on stdout; otherwise add convert as a named second
exception in CLAUDE.md and docs/operations.md.

#### CR-0068 [nit] [new] InjectShutdown's unknown-service warning quotes Lines[0], which is a held pragma on exactly the blocks that carry one

Location: internal/broker/inject.go:211-245, internal/broker/inject.go:437-443
Confidence: verified
Evidence:

    inject.go:439-441: `fmt.Sprintf("service %q is not in the shutdown table -- port line passed through
    unchanged: %q", li.service, strings.TrimSpace(b.Lines[0]))`. For a block with a held `! pragma:` the
    warning quotes the pragma instead of the port line, which is the same Lines[0]-vs-Opener() trap
    described at 222-226. The doc comment at 211-217 still says the function reads Lines[0] because a
    negated Kind is "no". Both claims are stale: blocks.go:415-423 strips `no`, and the code calls
    Opener() at 227.

Why it matters: The warning names the wrong line, and the stale comment misdescribes the parser contract.

Suggested patch (not applied):

```diff
--- a/internal/broker/inject.go
+++ b/internal/broker/inject.go
@@ -438,7 +438,7 @@ func InjectShutdown(blocks []Block, targetState func(service, transport string) (enabled bool, known bool)) ([]Block, []string, error) {
 			if li.unknownPort {
 				warnAt[i] = append(warnAt[i], fmt.Sprintf(
 					"service %q is not in the shutdown table -- port line passed through unchanged: %q",
-					li.service, strings.TrimSpace(b.Lines[0])))
+					li.service, strings.TrimSpace(b.Opener())))
 			}
 			continue
 		}
```

#### CR-0069 [nit] [new] broker status --pod is accepted and ignored unless --detail is also given

Location: internal/cli/commands.go:232-254, internal/cli/ops_k8s.go:463-482
Confidence: verified
Evidence:

    commands.go:241 wraps status in withPodRole and 252 registers --pod ("Kubernetes: which pod").
    ops_k8s.go:466-481 uses role only in `c.DescribeBroker(ctx, role)` under --detail without --all;
    BrokerReport and ClusterReport take no role. So `broker status --pod backup` (or with --all)
    silently reports the same thing as without it -- the shape ops_k8s.go:406-408 calls "the
    accepted-and-silently-dropped shape this tree refuses everywhere else".

Why it matters: An operator asking for the backup's status is not told the flag did nothing.

Change: Refuse --pod on broker status unless --detail is set and --all is not (usage error naming --detail),
or make BrokerReport narrow to the role.

### resilience

#### CR-0005 [major] [new] Docker broker remove empties the data dir under a broker compose down did not find

Location: internal/container/manager.go:999-1023, internal/container/manager.go:1117-1136,
internal/container/manager.go:1148-1156, internal/container/manager.go:1025-1052,
internal/render/render.go:1049-1063, docs/operations.md:556-567
Confidence: verified
Evidence:

    render.go 1057-1062: the file's name: is used so that COMPOSE_PROJECT_NAME keeps working as an
    override; operations.md 563-566: 'leaves COMPOSE_PROJECT_NAME working as an override ... Export it
    before `broker deploy` if you need a specific project name'.
    manager.go 1120: '... and a project that is not there is not an error.'
    1135  return engine.Run(ctx, m.R, containercli.ComposeDown(c.Argv(), file, m.projectDir(), env))
    1012  if purge {
    1013      if err := m.purgeData(ctx); err != nil {
    1151  args := []string{dir, "-mindepth", "1", "-delete"}
    Step by step: (1) the operator follows the doc and deploys with COMPOSE_PROJECT_NAME=legacy
    exported; (2) later, from a shell without it, `broker remove --delete-data --no-prompt`; (3) compose
    down targets project 'solace', finds nothing, exits 0; (4) Delete treats that as 'container gone'
    and runs `find <dataDir> -mindepth 1 -delete` under the still-running broker container_name
    'solace'. Without --delete-data, remove still reports success while the broker keeps serving.
    Nothing between 1135 and 1013 asks the engine whether container.name still exists. The podman path
    has the same shape when container.name changed since deploy: `systemctl stop <new>.service` fails,
    `is-active` prints 'inactive', systemd.Stopped accepts it (1041), and the purge proceeds against the
    old unit's live data directory.

Why it matters: Irreversible loss of a live broker's spool and configuration on a path the docs themselves
recommend; podman got an explicit 'confirmed stopped' guard for exactly this risk, docker has none.

Suggested patch (not applied):

```diff
--- a/internal/container/manager.go
+++ b/internal/container/manager.go
@@ -1131,6 +1131,26 @@ func (m *Manager) deleteDocker(ctx context.Context) error {
 	if err != nil {
 		return err
 	}
 	file := render.Compose(m.Cfg, m.Cfg.ResolveNode(config.Primary))
-	return engine.Run(ctx, m.R, containercli.ComposeDown(c.Argv(), file, m.projectDir(), env))
+	if err := engine.Run(ctx, m.R, containercli.ComposeDown(c.Argv(), file, m.projectDir(), env)); err != nil {
+		return err
+	}
+	// compose down of a project it cannot find exits 0, so a container deployed under
+	// another project name (COMPOSE_PROJECT_NAME) survives it. Confirm before Delete
+	// goes on to empty the data directory under it.
+	base, err := m.base()
+	if err != nil {
+		return err
+	}
+	running, answered, probeErr := m.containerRunningKnown(ctx, base)
+	if !answered {
+		return fmt.Errorf("compose down finished, but whether container %s is gone cannot be confirmed (the `ps` "+
+			"probe failed); the data directory is left in place -- check by hand and re-run: %w", m.name(), probeErr)
+	}
+	if running {
+		return fmt.Errorf("compose down finished but container %s is still running: it was deployed under a "+
+			"different compose project (COMPOSE_PROJECT_NAME?). The data directory is left in place; export the "+
+			"project name it was deployed under and re-run", m.name())
+	}
+	return nil
 }
```

#### CR-0011 [major] [new] InjectShutdown warnings are produced only after the teardown has run, not in the plan

Location: internal/broker/importops.go:386-390, internal/broker/importops.go:611-630,
internal/broker/importops.go:1094-1103, internal/broker/importops_test.go:426-444,
internal/broker/inject.go:506-516, internal/cli/exportconfig.go:59-85
Confidence: verified
Evidence:

    importops.go:388 `plan.Sections, plan.Warnings = planSections(c.Blocks, targetType)`: plan warnings
    come only from the section classification.
    importops.go:625-629 `blocks, warn, err := InjectShutdown(p.Capture.Blocks, nil)` ... `res.Warnings
    = append(res.Warnings, warn...)`: InjectShutdown runs only inside ImportApply, after confirmImport
    (exportconfig.go:65) and in the same call that runs the teardowns.
    The comments say otherwise. importops.go:1094-1100 says these warnings were moved into the plan so
    that "a service InjectShutdown cannot safely bounce" is seen before the confirmation, and the test
    comment at importops_test.go:429-433 says the same. That test only feeds a hand-built
    unclassified-section warning.
    targetState is always nil (625), so case 3 always warns (inject.go:510-516): "port line(s) passed
    through unchanged; shut the service down manually first".
    Failure: an artifact from an older broker has a bare `service smf listen-port ...` line (the case
    inject.go exists for). The operator sees no warning and types `yes`. The teardowns run, then the
    broker refuses the port change on a running service (inject.go:21-23), so stop-on-error fails the
    `sections: broker` chunk. Every VPN is now torn down and not rebuilt. The warning appears only in
    the result report, telling the operator to do something "first". A re-run fails in the same place.

Why it matters: The one warning that predicts a guaranteed failed import reaches the operator after the
irreversible teardown, which defeats the exact-yes gate. InjectShutdown is pure, so its warnings can be
computed in ImportPlan. If any test asserts InjectShutdown warnings on ImportResult for a hand-built plan,
it needs updating.

Suggested patch (not applied):

```diff
--- a/internal/broker/importops.go
+++ b/internal/broker/importops.go
@@ -386,5 +386,12 @@ func (o *Ops) ImportPlan(ctx context.Context, role config.Role, artifact []byte) (*ImportPlan, error) {
 		}
 	}
 	plan.Sections, plan.Warnings = planSections(c.Blocks, targetType)
+	// InjectShutdown is pure and runs with a nil targetState at apply, so what it will
+	// warn about is known now -- and belongs before the confirmation, not after teardown.
+	_, warn, err := InjectShutdown(c.Blocks, nil)
+	if err != nil {
+		return nil, err
+	}
+	plan.Warnings = append(plan.Warnings, warn...)
 	return plan, nil
 }
@@ -622,9 +629,9 @@ func (o *Ops) ImportApply(ctx context.Context, role config.Role, p *ImportPlan) (*ImportResult, error) {
 	// p.Capture.Blocks is the artifact, and it is what ImportVerify diffs the target
 	// against; folding synthesised lines into it would make verification demand that
 	// the target contain lines the artifact never had.
-	blocks, warn, err := InjectShutdown(p.Capture.Blocks, nil)
+	// Its warnings are already in p.Warnings (ImportPlan), which res starts from.
+	blocks, _, err := InjectShutdown(p.Capture.Blocks, nil)
 	if err != nil {
 		return res, err
 	}
-	res.Warnings = append(res.Warnings, warn...)
 	blocks = PrepareForApply(blocks, p.Target)
```

#### CR-0014 [major] [new] Podman redeploy forgets a written-but-unapplied change and later reports 'unchanged, left running'

Location: internal/container/manager.go:791-856, internal/container/manager.go:685-718,
internal/container/manager.go:1193-1203, internal/container/manager_test.go:2269-2278,
docs/operations.md:957-986
Confidence: verified
Evidence:

    Change detection compares against what this tool WROTE, never against what the running service was
    started from:
    796  unitChanged := m.unitChanged(unit, body)        // 716-717: os.ReadFile(path) vs body
    802  changedSecrets := m.changedSecrets(ctx, base, secrets, replace)   // store vs config
    824  if err := m.writePodmanSecrets(ctx, base, changedSecrets, replace); err != nil {
    828  if _, err := m.writeArtifact(unit, body, "quadlet unit", 0o755); err != nil {
    834  if err := m.removeLegacyCertBundle(); err != nil { return err }
    838  if changed { ... systemd.DaemonReload(user) ...
    847  return m.systemctl(ctx, systemd.Restart(user, svc))
    853  m.progress().Info("%s is already active on this unit and these secrets -- unchanged, left
    running.", svc)
    Step by step: (1) running broker, operator bumps image.tag (or rotates a password) and approves the
    restart; (2) secrets and unit are written (824/828); (3) removeLegacyCertBundle fails -- e.g. EACCES
    on a bundle an earlier rootful build left -- and its own message (700-702) says 'remove it by hand
    and re-run'; or daemon-reload fails (TestManagerDeployPodmanDaemonReloadError only asserts the error
    propagates); (4) the operator re-runs deploy: unitChanged=false (file matches), every secret matches
    the store, so changed=false, active=true, and line 853 prints 'unchanged, left running' and exits 0
    while the service still runs the OLD unit and the container still holds the OLD secret copies. On a
    first deploy (not active) the re-run skips daemon-reload (838) and `systemctl start` asks for a
    service quadlet never generated. `broker restart` (1195-1203) issues no daemon-reload either. The
    documented lossy-probe branch (786-790, operations.md 984-986) lands in the same state: start is a
    no-op on the running unit, nothing is printed, and every later deploy says 'unchanged'.

Why it matters: Following the tool's own recovery instruction ('re-run') produces a success message while
the broker keeps the old image/config/credentials; docker does not have this hole because it reads the
applied state back from the running container's digest label.

Change: In deployPodman, derive 'changed' from applied state as well as written state: treat the unit as
changed when `systemctl [--user] show <svc> -p NeedDaemonReload --value` answers yes (covers a write that
was never reloaded), and treat secrets as changed when any secret's UpdatedAt (podman secret inspect) is
newer than the running container's StartedAt. Move removeLegacyCertBundle ahead of the first write so its
failure leaves nothing half-applied, and in the lossy-probe branch (!active because the probe did not
answer) say on stderr that a running service was not restarted. Add a manager_test case: daemon-reload
fails, then a second Deploy must reload and restart rather than report unchanged; catalogue it in
docs/test.md.

#### CR-0016 [major] [new] A switch that fails mid-plan returns a bare error: VPNs left standby at both sites are not named

Location: internal/broker/switchplan.go:324-356, internal/broker/switchplan.go:409-443,
internal/cli/replication.go:193-196, docs/operations.md:1864-1872,
internal/broker/switchplan_test.go:700-724
Confidence: verified
Evidence:

    switchplan.go:338-340 (demote): `if err := c.SetRole(ctx, act.VPN, act.Role); err != nil { return
    err }`; 413-416 (pre-promotion re-read): `return fmt.Errorf("re-reading %s before promoting: %w",
    site, err)`; 438-440 (promote): `if err := c.SetRole(...); err != nil { return err }`. Only
    confirmDemotions (387-389) says the VPNs are "standby at both sites -- unavailable, not corrupted".
    cli/replication.go:193-195: `if err := broker.ExecuteSwitchPlan(...); err != nil { return err }` --
    no status page, no verify, no recovery sentence.
    docs/operations.md:1866-1868: "Failing between the demote and the promote leaves those VPNs standby
    at BOTH sites ... and the report says so plainly"; switchplan_test.go:700 says the attempted
    promotion is recorded "so the report can say how far the run got" -- nothing reports it.
    Failure: plan moves ORDERS and PAYMENTS to v:dr1. ORDERS promoted; the kubectl exec for PAYMENTS
    drops over the WAN -> "set message-vpn \"PAYMENTS\" to active on site v:dr1 (...): run cli script
    ...: exit status 1". PAYMENTS (and every VPN after it) is standby at both sites, possibly active if
    the broker applied it before the drop; the operator gets one line about a transport error, no
    picture of either site, and no statement that clients of those VPNs are without an active site or
    that re-running resumes. Same when the 2nd of N demotions fails after the 1st landed.

Why it matters: Four of the five failure exits in the highest-stakes command leave an outage the error does
not mention, contradicting the documented contract; the operator must reconstruct state from stderr progress
lines mid-incident.

Suggested patch (not applied):

```diff
--- a/internal/cli/replication.go
+++ b/internal/cli/replication.go
@@ -193,4 +193,12 @@ func opPerformReplication(a *App, o *broker.Ops, role config.Role, mateFor mateChannelFunc) error {
 	if err := broker.ExecuteSwitchPlan(bg(), chans, plan, vpns, lineSink()); err != nil {
-		return err
+		// Where the run stopped is the operator's first question: show both sites as they
+		// are now (best-effort -- the switch error stays the verdict), then say what that
+		// state means and how to resume.
+		if serr := reportReplicationStatus(chans); serr != nil {
+			warn("could not read both sites after the failed switch: %v", serr)
+		}
+		return fmt.Errorf("%w\n\nThe switch stopped partway and nothing is rolled back. A VPN demoted but not yet "+
+			"promoted is standby at BOTH sites -- unavailable, not corrupted. Fix the cause and run this command "+
+			"again: the re-run skips what already moved and promotes the rest", err)
 	}
 	return verifyReplicationSwitch(a, chans, plan, vpns)
```

#### CR-0034 [minor] [new] Half of image.user/image.pass is silently discarded on Kubernetes and only refused at deploy on containers

Location: internal/config/config.go:190-221, internal/config/validate.go:104-109,
internal/config/validate.go:683-692, internal/config/validate.go:1386-1391,
internal/container/manager.go:456-463
Confidence: verified
Evidence:

    config.go:198-199  func (c *Config) ManagesImagePullSecret() bool { return c.Image.User != "" &&
    c.Image.Pass != "" }
    Nothing in Validate pairs the two (a grep finds no Image.User check in internal/config).
    manager.go:461-462  if user == "" || pass == "" { return fmt.Errorf("image.user and image.pass must
    both be set ...") }  (deploy time, containers only)
    Contrast validate.go:683-692, where tls.cert/certKey "must be set together" at load, and
    validate.go:1390-1391: "Refusing beats dropping them: a key the env file sets and the broker never
    receives is the silent kind of wrong."
    Failure: on kubernetes, `image: {user: robot}` with no pass/passEnv loads clean.
    ManagesImagePullSecret is false, so no Secret is built and the CR has no pullSecrets block. A
    private-registry pull then fails as ImagePullBackOff, and nothing names the dropped username. On
    docker/podman the same file passes `validate` and fails only at deploy.

Why it matters: This is a silent fallback the project's own rule forbids, and the TLS pair already applies
that rule. TestManagesImagePullSecretNeedsBothCredentials pins only the accessor, not load acceptance, so
the refusal adds no conflict. TestValidateCredentialControlChars still passes, because the control-character
check runs first.

Suggested patch (not applied):

```diff
--- a/internal/config/validate.go
+++ b/internal/config/validate.go
@@ -107,6 +107,20 @@ func (c *Config) Validate(p Platform) error {
 	if err := c.validateImage(); err != nil {
 		return err
 	}
+
+	// Registry credentials are a pair on every platform: docker and podman log in with
+	// both, and Kubernetes builds its pull Secret only from both (ManagesImagePullSecret).
+	// Half of one is a value the file sets and nothing uses, so it is refused here the
+	// way tls.cert/tls.certKey are. The user is named; the password is a secret.
+	if (c.Image.User == "") != (c.Image.Pass == "") {
+		pass := "MISSING"
+		if c.Image.Pass != "" {
+			pass = "set"
+		}
+		return fmt.Errorf("image.user and image.pass (or image.passEnv) must be set together (image.user %q, "+
+			"image.pass %s): one without the other builds no image-pull Secret on Kubernetes and fails "+
+			"`<command> login` on docker and podman. Set the missing one, or unset both", c.Image.User, pass)
+	}
 
 	// The timezone is the one free-text value every platform writes into its
 	// artifact (TZ in compose and quadlet, timezone in the CR), so it is held to an
```

#### CR-0035 [minor] [new] v0.2.x schema keys (redundancy scalar, nodes, nodes.*.ip, storage.msgNode/monNode) were dropped, not refused by name

Location: internal/config/config.go:30-84, internal/config/config.go:717-731,
internal/config/config.go:1221-1271, internal/config/validate.go:18-35
Confidence: verified
Evidence:

    CLAUDE.md: "A retired key is retained and refused by name. ... the old field still parses so a file
    carrying it fails loud naming its replacement instead of hitting a bare unknown-key error. Do the
    same for any future rename."
    `git show v0.2.1:env/sample.yaml`: line 31 `redundancy: yes`, 179-180 `msgNode: 30Gi` / `monNode:
    5Gi` (kubernetes.storage), 322-331 `nodes:` with `ip:`. These were released in v0.1.1-v0.2.1 and
    renamed in 18854f0 (v0.3.0).
    The current Config has no nodes field, Storage has only MsgNodeSize/MonNodeSize (config.go:719-720),
    Node has only name/addr (config.go:1266-1270), and Redundancy is a struct, so the old scalar
    `redundancy: yes` fails decoding. Only `admin:` was retained (LegacyAdmin, config.go:47-51, refused
    at validate.go:23-28).
    Failure: a v0.2.1 env file gets `field nodes not found in type config.Config`, `field msgNode not
    found in type config.Storage` and `cannot unmarshal !!str yes into config.Redundancy`. None of these
    says the key moved to redundancy.primary/backup/monitor (ip -> addr) or to msgNodeSize/monNodeSize.

Why it matters: The house rule exists so upgrades fail with an actionable rename message. Operators
upgrading from v0.2.x get bare decode errors instead, and docs/configuration.md does not record these
renames either.

Change: Retain and refuse, following the LegacyAdmin and maxPool pattern: add `LegacyNodes map[string]any
`yaml:"nodes"`` to Config, `LegacyMsgNode`/`LegacyMonNode string` (yaml msgNode/monNode) to Storage, and
`LegacyIP string `yaml:"ip"`` to Node. Refuse each in validateRenamedKeys, naming
redundancy.primary|backup|monitor (ip -> addr, psk -> redundancy.psk) and
kubernetes.storage.msgNodeSize/monNodeSize. For the scalar `redundancy: yes`, catch it before the struct
decode: either give Redundancy an UnmarshalYAML that returns the rename error on a ScalarNode and
re-implements KnownFields on a MappingNode (the scaling.go pattern, which CLAUDE.md requires), or sniff a
top-level scalar `redundancy` in parseError. Add the refusals to the test catalogue.

#### CR-0036 [minor] [new] convert's validateOutput skips Load's '~' expansion and wrongly warns that '~/...' paths will not load

Location: internal/convert/convert.go:137-151, internal/convert/convert.go:84-92,
internal/config/load.go:55-74, internal/config/hostpath.go:102-123
Confidence: verified
Evidence:

    convert.go:145-150  c, err := config.DecodeStrict(...); c.ApplyDefaults(p); return c.Validate(p)
    load.go:65-73  Load runs expandCommandHomes, resolveQuadletDir and expandHomePaths between
    ApplyDefaults and Validate.
    hostpath.go:116-121  a surviving leading '~' is refused: "one surviving here means this
    configuration was not loaded from a file -- write the path out in full".
    Failure: a legacy file with `SOLBK_DATA_DIR=~/solace/data` (bash expands '~' in assignments)
    converts to `dataDir: "~/solace/data"`. validateOutput refuses it and Convert prints "the converted
    file will not load as-is: ... write the path out in full". The file does load, because Load expands
    '~' first. KUBE="kubectl --kubeconfig ~/k" produces the same false warning through execguard.

Why it matters: The convert check exists to predict what Load will do (convert.go:141-143). It diverges in
the opposite direction from the one it guards against, so the user gets a false failure and advice to remove
a supported form.

Change: Give config one exported normalisation step used by both callers, for example `func (c *Config)
Prepare(p Platform) error` that runs ApplyDefaults, expandCommandHomes, resolveQuadletDir and
expandHomePaths in Load's order. Call it from Load (in place of those four lines) and from
convert.validateOutput (in place of ApplyDefaults) so the two cannot drift. Add a convert_test case with
SOLBK_DATA_DIR=~/solace/data that expects no 'will not load' warning.

#### CR-0037 [minor] [new] DecodeStrict reads only the first YAML document; anything after a second '---' is silently ignored

Location: internal/config/load.go:24-36, internal/config/platform.go:81-99
Confidence: verified
Evidence:

    load.go:30-35  dec := yaml.NewDecoder(...); dec.KnownFields(true); if err := dec.Decode(&c); err !=
    nil {...}; return &c, nil
    yaml.v3 v3.0.1 yaml.go:119-136: Decoder.Decode parses ONE node per call, so later documents stay
    unread. DetectPlatforms (platform.go:86-87) likewise uses yaml.Unmarshal, which reads the first
    document only.
    Failure: an env file whose second document carries `semp: {adminPass: ...}` (for example, two
    snippets concatenated with `---`) loads on kubernetes without it. ManagesAdminSecret is false, no
    Secret is built, and the operator generates a random admin password. No message says the second
    document was dropped.

Why it matters: A silent partial read of the configuration. Every other unread key in this schema fails loud
through KnownFields.

Suggested patch (not applied):

```diff
--- a/internal/config/load.go
+++ b/internal/config/load.go
@@ -2,7 +2,9 @@ package config
 
 import (
 	"bytes"
+	"errors"
 	"fmt"
+	"io"
 	"os"
 	"path/filepath"
 	"regexp"
@@ -32,5 +34,21 @@ func DecodeStrict(raw []byte) (*Config, error) {
 	if err := dec.Decode(&c); err != nil {
 		return nil, err
 	}
-	return &c, nil
+	// One document only. Decode reads the first and stops, so anything after a second
+	// `---` would otherwise be dropped unread -- a semp: block there is a password the
+	// broker never receives. An empty trailing document (a bare closing `---`) is fine.
+	for {
+		var rest yaml.Node
+		err := dec.Decode(&rest)
+		if errors.Is(err, io.EOF) {
+			return &c, nil
+		}
+		if err != nil {
+			return nil, err
+		}
+		if len(rest.Content) > 0 && rest.Content[0].ShortTag() != "!!null" {
+			return nil, fmt.Errorf("line %d: the env file holds a second YAML document (after `---`), which is "+
+				"never read; merge it into the first", rest.Content[0].Line)
+		}
+	}
 }
```

#### CR-0038 [minor] [new] Interrupting an import leaves the credential-bearing driver on the broker and the apply running unobserved

Location: internal/broker/importops.go:860-873, internal/broker/driver.go:157-168,
internal/cli/ops_k8s.go:43, internal/cli/exportconfig.go:82-88, internal/broker/transport.go:72-78,
docs/operations.md:1527-1530
Confidence: verified
Evidence:

    importops.go:865-871: `dest := shellScriptPath("solace-util-import")` ... `o.T.Upload(...)` ...
    `defer o.removeFiles(ctx, role, dest)`. The driver holds every chunk, "credentials included" (868).
    driver.go:161 `rm -f "$TMP" "$CLIDIR/$WRAPPER"` plus the chunk files. The in-broker trap never
    removes the driver script itself (`$0`).
    ops_k8s.go:43 `func bg() context.Context { return context.Background() }`. The only
    signal.NotifyContext in the repo is in watch.go:63. On Ctrl-C the Go runtime exits without running
    deferred calls, so removeFiles never runs.
    Failure: the operator hits Ctrl-C during `sh /usr/sw/jail/.solace-util-import`. The host process
    dies and no result report is printed. `/usr/sw/jail/.solace-util-import` stays on the broker with
    the full configuration, contrary to operations.md:1529-1530 ("the script deletes every file it wrote
    even if it is interrupted"). The exec has no TTY, so the remote driver most likely keeps applying,
    teardowns included, while the operator believes the run was aborted. A re-run (the documented
    recovery) then uploads to the same fixed paths while the old driver may still be reading them.

Why it matters: Credential-equivalent configuration is left on the broker filesystem, and a destructive
apply can carry on after the operator thinks it stopped. Whether the remote shell keeps running after
kubectl or docker exec dies is the suspected part; confirm it with a long import interrupted from the host.

Change: In renderDriver (driver.go:161), add `"$0"` to the cleanup rm list so the uploaded driver deletes
itself on EXIT/INT/TERM whatever the host does, and update the golden in driver_test.go
(TestRenderDriverExactOutput) and TestRenderDriverCleansUpEveryChunk. In runImport, run the import under
signal.NotifyContext(os.Interrupt) instead of bg(), so Ctrl-C cancels, runs the deferred removal, and prints
that the broker may still be applying. Consider a `mkdir` lock in the driver so a second run refuses while
one is active.

#### CR-0039 [minor] [new] Driver ignores failed chunk/wrapper writes and an empty transcript, so it can report OK for what never ran

Location: internal/broker/driver.go:147, internal/broker/driver.go:170-195,
internal/broker/broker.go:248-261
Confidence: verified
Evidence:

    driver.go:147 sets only `set -u`; there is no -e and no status check.
    driver.go:172 `cat > "$CLIDIR/<chunk>" <<'...'`, :182 `printf 'source script ...' >
    "$CLIDIR/$WRAPPER"` and :183 `"$CLI" -Apes "$WRAPPER" > "$TMP" 2>&1` are all unchecked.
    driver.go:184 `if tail -n 10 "$TMP" | grep -qiE ...; then <FAIL>` and otherwise :194 prints
    CHUNK-OK.
    Failure: the jail filesystem fills after the driver upload succeeds, which is a known Solace
    operational condition. A chunk heredoc is then truncated or empty, and the transcript cannot be
    written. `tail` sees no rejection phrase, so the driver prints `SOLACE-UTIL-CHUNK-OK i`. For a
    TEARDOWN chunk this matters most: importops.go:36-41 says the driver is the ONLY detector of a
    failed teardown, because the diff cannot see one. The VPN is reported as torn down
    (importops.go:652-654) when it was not.

Why it matters: This only bites on a full or read-only jail filesystem. When it does, the detector the
design relies on for teardowns fails silently in the success direction.

Change: In renderDriver, end the heredoc open line with `|| { echo 'SOLACE-UTIL-CHUNK-FAIL i'; exit 1; }`
(the heredoc body still follows that line), do the same for the wrapper printf, and before the tail scan add
`[ -s "$TMP" ] || { echo <DETAIL-BEGIN>; echo 'empty transcript'; echo <DETAIL-END>; echo
'SOLACE-UTIL-CHUNK-FAIL i'; exit 1; }` (a real `cli -Apes` always prints its banner). Update the golden in
driver_test.go and add a case for a write failure. runCLISkeleton (broker.go:248-261) has the same unchecked
writes.

#### CR-0040 [minor] [new] No deadline on any mate call: a stalled SEMP peer or kubectl exec hangs the switch

Location: internal/cli/ops_k8s.go:43, internal/cli/replication.go:161-196, internal/broker/semp.go:149-162,
internal/broker/sempmate.go:74-97, internal/k8s/matechannel.go:65-107
Confidence: verified
Evidence:

    ops_k8s.go:43 `func bg() context.Context { return context.Background() }`, used for every switch
    call (replication.go:161, 168, 193). semp.go:156-157 builds the curl config from `user` plus the
    caller's lines (sempmate.go:75-76 adds only `data` and optionally `insecure`) and runs `curl -is -K
    - url` -- no connect-timeout, no max-time. k8s/matechannel.go:73 and 102 `engine.Output(ctx, r,
    ...)` for kubectl exec with no --request-timeout. A grep of internal/ for
    WithTimeout/WithDeadline/max-time/connect-timeout returns nothing outside tests; the only signal
    handler is cli/watch.go:63.
    Failure: the mate's SEMP endpoint accepts the connection and stalls (overloaded broker, half-broken
    proxy) during the promote phase -> the run blocks indefinitely after logging "promoting message-vpn
    ...", with the moving VPNs standby at both sites; under --no-prompt in automation nothing ever
    interrupts it, and Ctrl-C loses the state report.

Why it matters: A DR switch runs precisely when the far site or the WAN is degraded; an unbounded wait there
turns a recoverable partial failure into an indefinite outage window with no error at all.

Change: Bound each mate call: wrap cliMate.run and sempMate.rpc (or opPerformReplication's calls) in
context.WithTimeout, and add `connect-timeout` and `max-time` curl config lines for the sempMate leg
(curlConfigLine in semp.go). On expiry the partial-failure report proposed in CR-0016 then tells the
operator where it stopped.

#### CR-0041 [minor] [new] PVC and CR deletes during `broker remove` have no timeout and run under context.Background

Location: internal/k8s/deploy.go:176-222, internal/cli/ops_k8s.go:41-43
Confidence: verified
Evidence:

    deploy.go:187  c.deleteStdin(ctx, render.BrokerCR(c.Cfg))   -> `delete -f - --ignore-not-found`
    (shared kube.Delete adds no --timeout)
    deploy.go:215  c.kubectl(ctx, "delete", "pvc", pvc, "-n", c.ns(), "--ignore-not-found")
    ops_k8s.go:43  func bg() context.Context { return context.Background() }
    kubectl delete waits by default (--wait=true), and with --timeout unset its wait is 168h.
    kubernetes.io/pvc-protection holds a claim while any pod mounts it.
    Failure: `broker remove --delete-data` after a node loss. The broker pod on the lost node stays
    Terminating, so `kubectl delete pvc` blocks with nothing after "deleting PVC ..." until the node
    returns, someone force-deletes the pod, or the operator presses Ctrl-C. The rollout waits are
    bounded (ops.go:149-151, "so a stuck pod fails loud instead of hanging"), but the deletes are not.

Why it matters: An unbounded wait on the teardown path is the hang this codebase already fixed for rollouts.
Interrupting it leaves the PVC marked for deletion and the Secrets in place, and nothing tells the operator
that.

Change: In internal/k8s/deploy.go, pass a bounded --timeout (kube.RolloutTimeout) to the PVC delete, and to
the CR delete through a deleteStdin variant. On a timeout, report that the claim is marked for deletion and
is waiting on its pod (not "persistent data survives"), and name `kubectl get pod <pod> -n <ns>` as the next
check. Update the argv assertions in deploy_test.go and the docs/test.md entries.

#### CR-0042 [minor] [new] Podman server-certificate secret (private key) is orphaned in the store once tls.* is dropped

Location: internal/render/render.go:740-751, internal/container/manager.go:1080-1115,
internal/container/manager.go:791-830, internal/container/manager.go:674-686,
internal/cli/ops_container.go:427-435
Confidence: verified
Evidence:

    render.go 740  if c.TLS.Cert != "" && c.TLS.CertKey != "" {  // the cert secret exists in the list
    only while TLS is configured
    manager.go 1081  secrets := render.ContainerSecrets(m.Cfg, m.P)  // removePodmanSecrets walks the
    CURRENT config
    manager.go 1104  "it holds the broker's PRIVATE KEY and must not be left in this host's secret
    store"
    manager.go 678-679 (removeLegacyCertBundle): 'The path does not depend on TLS being configured: a
    host may have dropped tls.cert since the file was written.'
    ops_container.go 434-435: 'removing the certificate is NOT optional and a failure there is fatal,
    because leaving a private key on the host is the worst outcome available.'
    Step by step: (1) deploy with tls.cert/certKey -> `<name>-tls-servercertificate` (key + cert) is
    created in podman's store; (2) TLS moves to a load balancer, tls.* is removed from the env file,
    redeploy -> deployPodman only compares/writes the current list (802, 824), the unit stops
    referencing the secret, the store keeps the key; (3) `broker remove` -> removePodmanSecrets iterates
    the current list, never names the cert secret, prints nothing about it. The key stays in podman's
    file driver (unencrypted, per container-security.md 213-215). The same applies to a dropped
    additional user's password, the PSK after HA is turned off, and a dropped passphrase.

Why it matters: The project treats a FAILED removal of this secret as fatal, yet the commonest way for it to
survive -- a config that no longer mentions it -- is silent; the legacy bundle path already handles this
exact case and the store path does not.

Change: Export the cert secret name from render (e.g. render.ServerCertSecretName(cfg, p) built from
container.name + certSuffix, independent of tls.*) and have removePodmanSecrets always `podman secret rm
--ignore` it, with the same fatal-on-failure rule; in deployPodman, when TLS is not configured, remove it
(rm --ignore) after the new unit is written. Consider the same name-based sweep for the PSK and passphrase
suffixes. Add a manager_test case (deploy with TLS, Delete with TLS unset -> the cert secret rm is still
issued) and catalogue it in docs/test.md.

#### CR-0043 [minor] [new] --watch turns a bad --pod into an endless retry loop that exits 0

Location: internal/cli/commands.go:232-245, internal/cli/watch.go:54-66, internal/cli/ops_k8s.go:869-872
Confidence: verified
Evidence:

    commands.go:240-242 resolves --pod INSIDE the refreshed function: `withWatch("broker status", func(a
    *App) error { return dispatch(platformOps(withPodRole(opK8sStatusBroker), ...), a) })`. withWatch
    (watch.go:56-59) validates only --interval before the loop. The pinned watch.loop (go-deploy-utility
    v0.1.2 watch/watch.go:141-150) warns on every pass error and never stops; ctx cancellation returns
    nil. So `solace-util broker status --watch --pod bogus` on Kubernetes prints `unknown role ... --
    trying again in 5s` forever, and Ctrl-C exits 0, while the same command without --watch exits 2
    (podRole wraps with exitcode.AsUsage, ops_k8s.go:869-872). watch.go:52-53 claims "The flags are read
    first either way".

Why it matters: A usage error becomes an indefinite loop and a success exit, breaking the exit-status
contract (2 = bad command line) for the watch path only.

Suggested patch (not applied):

```diff
--- a/internal/cli/commands.go
+++ b/internal/cli/commands.go
@@ -237,4 +237,10 @@ func newBrokerStatusCmd(app *App) *cobra.Command {
 		Args:              noRolePositional,
 		ValidArgsFunction: cobra.NoFileCompletions,
 		RunE: func(_ *cobra.Command, _ []string) error {
+			// --pod is otherwise read inside each refresh, where --watch would turn a bad value
+			// into a warning repeated until Ctrl-C (and exit 0). Resolve it once, up front, so it
+			// stays the usage error it is without --watch.
+			if _, err := podRole(app); err != nil {
+				return err
+			}
 			return withWatch("broker status", func(a *App) error {
```

#### CR-0044 [minor] [new] vulnjudge passes an empty or message-boundary-truncated govulncheck stream as clean

Location: internal/tools/vulnjudge/main.go:107-128, internal/tools/vulnjudge/main.go:194-206,
internal/tools/vulnjudge/main_test.go:111-123, scripts/dev.sh:183-193
Confidence: verified
Evidence:

    main.go:121-126:
      err := dec.Decode(&m)
      if errors.Is(err, io.EOF) { break }
    A zero-byte file (or whitespace, `null`, `{}`, or a stream cut cleanly after the config/progress
    messages) reaches 203-205 `no vulnerabilities called by this module` and exit 0. Nothing requires
    the `config` message govulncheck always emits first. main_test.go:111-112 states the intended
    invariant: "Truncated or non-JSON input is a systemic problem, not a clean scan: it must fail loudly
    rather than report success" -- but only mid-object truncation (ErrUnexpectedEOF) is caught.
    dev.sh:188 and dev.ps1:252 check govulncheck's own exit status, so the hole needs govulncheck to
    exit 0 with short output (e.g. a write error on a full disk it does not report).

Why it matters: The scan gate's last line of defence fails open on exactly the input it claims to reject;
cheap to close.

Suggested patch (not applied):

```diff
--- a/internal/tools/vulnjudge/main.go
+++ b/internal/tools/vulnjudge/main.go
@@ -46,6 +46,9 @@ type frame struct {
 // message is the subset of govulncheck's stream messages the decision needs;
 // `config` and `progress` entries decode to a zero value and are skipped.
 type message struct {
+	// Config is only checked for presence: govulncheck opens every stream with it, so a
+	// stream without one is empty or cut short, never a clean scan.
+	Config json.RawMessage `json:"config"`
 	OSV *struct {
 		ID      string `json:"id"`
 		Summary string `json:"summary"`
@@ -109,6 +112,7 @@ func judge(data []byte) (int, string) {
 	dec := json.NewDecoder(bytes.NewReader(bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))))
 
 	seen := map[string]*vuln{}
+	sawConfig := false
 	entry := func(id string) *vuln {
 		if v, ok := seen[id]; ok {
 			return v
@@ -129,6 +133,8 @@ func judge(data []byte) (int, string) {
 		}
 
 		switch {
+		case len(m.Config) > 0:
+			sawConfig = true
 		case m.OSV != nil:
 			entry(m.OSV.ID).summary = m.OSV.Summary
 		case m.Finding != nil:
@@ -160,6 +166,9 @@ func judge(data []byte) (int, string) {
 			}
 		}
 	}
+	if !sawConfig {
+		return 2, "vulnjudge: no govulncheck config message -- the stream is empty or was cut short, not a clean scan\n"
+	}
 
 	var fixable, nofix []*vuln
 	uncalled := 0
```

#### CR-0070 [nit] [new] convert warning names admin.additionalUsers; the key is semp.additionalUsers and admin: is refused

Location: internal/convert/convert.go:278-281, internal/config/validate.go:23-28
Confidence: verified
Evidence:

    convert.go:279  "SOLBK_USR_PASS became admin.additionalUsers with accessLevel: none -- "
    validate.go:24  "the `admin:` section was renamed to `semp:`"  (Load refuses `admin:`)
    convert.go:248 actually emits the users under d.section("semp", ...).

Why it matters: The warning points at a key the loader refuses, so following it literally breaks the file.

Suggested patch (not applied):

```diff
--- a/internal/convert/convert.go
+++ b/internal/convert/convert.go
@@ -278,3 +278,3 @@ func emitYAML(v *vars, p config.Platform, source string) (string, []string) {
 	if converted > 0 {
-		warns = append(warns, "SOLBK_USR_PASS became admin.additionalUsers with accessLevel: none -- "+
+		warns = append(warns, "SOLBK_USR_PASS became semp.additionalUsers with accessLevel: none -- "+
 			"the bash flow set no access level; give each user the level it actually needs")
```

#### CR-0071 [nit] [new] Errors over map-typed fields pick the reported entry in random map order

Location: internal/config/validate.go:146-155, internal/config/validate.go:296-299,
internal/config/validate.go:785-794, internal/config/hostpath.go:381-387
Confidence: verified
Evidence:

    validate.go:146-150 states the rule: "Go randomises map iteration -- an error that named a different
    one of two bad keys per run would make a bug report unreproducible" and sorts for that reason.
    But validate.go:296 `for ca, path := range c.Broker.DomainCerts.Files` (CheckHostPath),
    validate.go:785 `for key, claim := range st.CustomVolumeMount`, and hostpath.go:381 `for ca, p :=
    range c.Broker.DomainCerts.Files` (expandTilde) iterate unsorted and return on the first bad entry.

Why it matters: With two bad entries, the error names a different one from run to run, which the file itself
calls unreproducible.

Change: Iterate sorted keys at validate.go:296, validate.go:785 and hostpath.go:381, using the same
sort.Strings over the keys that validateDomainCerts does at validate.go:151-155.

#### CR-0072 [nit] [new] An empty broker.domainCerts.dirs path is accepted at load and fails only at the op

Location: internal/config/config.go:924-962, internal/config/validate.go:145-172,
internal/config/hostpath.go:73-76, internal/config/domaincerts.go:286-290
Confidence: verified
Evidence:

    config.go:939-943,956  `path: ""` sets havePath=true and Path "" (a scalar "" or a null entry also
    yields Path "").
    hostpath.go:74-75  CheckHostPath returns nil on "".
    validate.go:160-169 refuses an empty FILES value at load, because "the entry can only ever fail --
    at upload", but has no dirs counterpart.
    domaincerts.go:287-289  readDir("") -> "broker.domainCerts.dirs[0] () cannot be read: open : no such
    file or directory" at `broker configure domain-certs`.

Why it matters: This is the same deferred-typo cost the files check was added to remove, and the eventual
error is unclear.

Suggested patch (not applied):

```diff
--- a/internal/config/validate.go
+++ b/internal/config/validate.go
@@ -168,5 +168,12 @@ func (c *Config) validateDomainCerts() error {
 				"Give it one, or remove the entry", ca)
 		}
 	}
+	// The dirs twin of the empty files value above: nothing defaults it either, and an
+	// empty path only fails later, at the op, as "cannot be read: open : ...".
+	for i, d := range c.Broker.DomainCerts.Dirs {
+		if strings.TrimSpace(d.Path) == "" {
+			return fmt.Errorf("broker.domainCerts.dirs[%d] has no path: name a directory to walk, or remove the entry", i)
+		}
+	}
 	return nil
 }
```

#### CR-0073 [nit] [new] verifyReplicationSwitch drops the wrong-role verdict when a status page cannot be read

Location: internal/cli/replication.go:203-235
Confidence: verified
Evidence:

    226-228: `if err := reportReplicationStatus(chans); err != nil { return err }` runs before 229-233
    `if len(wrong) > 0 { ... return fmt.Errorf("the switch did not leave the pair where the env file
    says it should be: %s" ...`. When roles are wrong AND a Status read fails, the operator gets only
    the status read error and never learns which VPN is where.

Why it matters: The computed verdict is the most important line of a DR run and is discarded for a
display-only failure.

Suggested patch (not applied):

```diff
--- a/internal/cli/replication.go
+++ b/internal/cli/replication.go
@@ -226,9 +226,11 @@ func verifyReplicationSwitch(a *App, chans map[string]broker.MateChannel, plan broker.SwitchPlan,
-	if err := reportReplicationStatus(chans); err != nil {
-		return err
-	}
+	statusErr := reportReplicationStatus(chans)
 	if len(wrong) > 0 {
 		sort.Strings(wrong)
-		return fmt.Errorf("the switch did not leave the pair where the env file says it should be: %s",
-			strings.Join(wrong, "; "))
+		err := fmt.Errorf("the switch did not leave the pair where the env file says it should be: %s",
+			strings.Join(wrong, "; "))
+		if statusErr != nil {
+			return fmt.Errorf("%w (the status pages could not be printed either: %v)", err, statusErr)
+		}
+		return err
 	}
-	return nil
+	return statusErr
```

#### CR-0074 [nit] [new] Phase-1 rejection text says it shuts down each listed VPN and that all of them are now disabled

Location: internal/broker/replicationops.go:241-264, internal/broker/replicationops.go:316-329
Confidence: verified
Evidence:

    runMatePhase 247-248 shuts down every name in `stopped` = mateConvergenceShutdowns(state), i.e.
    every admin-enabled VPN, listed or not. replPhase1Rejected 319: "Phase 1 sends, in order: a
    replication shutdown for each listed VPN" and 323: "These have replication disabled" for the whole
    list, although 320-321 say "WHERE it stopped decides what is true now" (docs/operations.md:1756
    correctly says "every VPN it reached").

Why it matters: The error an operator reads mid-incident misstates which VPNs phase 1 touches and overstates
which are now down.

Suggested patch (not applied):

```diff
--- a/internal/broker/replicationops.go
+++ b/internal/broker/replicationops.go
@@ -318,7 +318,8 @@ func replPhase1Rejected(err error, stopped []string) error {
 		"Nothing is rolled back, and phase 2 (the per-VPN state) was never sent. Phase 1 "+
-		"sends, in order: a replication shutdown for each listed VPN, the mate removals, "+
+		"sends, in order: a replication shutdown for every VPN with replication enabled, the mate removals, "+
 		"then the new mate address lines -- so WHERE it stopped decides what is true now, "+
 		"and only the broker can say. If it stopped after the removals, this broker holds "+
 		"NO mate configuration at all.\n\n"+
-		"These have replication disabled and will NOT have it turned back on by this run: "+
+		"Phase 1 was to disable replication on these, and this run will NOT turn it back on for any of "+
+		"them -- each one it reached before it stopped is disabled now: "+
 		"%s. The VPNs themselves are untouched and their clients are still connected.\n\n"+
```

### quality

#### CR-0045 [minor] [new] Dead code with no production caller is kept covered by its own tests

Location: internal/k8s/ops.go:14-27, internal/k8s/operator.go:628-641,
internal/broker/verify_ops.go:265-272, internal/broker/semp.go:268-271, internal/broker/scripts.go:206-215,
internal/broker/replication.go:570-574, internal/render/render.go:513-514, internal/k8s/ops_test.go:14-58,
docs/test.md:2449, docs/test.md:2505-2506, internal/container/manager.go:49-50,
internal/container/manager.go:93, internal/cli/ops_container.go:48-56
Confidence: verified
Evidence:

    A scan of every production func for references outside its own declaration and comments:
    k8s/ops.go:17 `func (c *Cluster) Status` has no caller. `broker status` uses statusreport, and
    opK8sStatusOperator (ops_k8s.go:486-496) calls OperatorReport, not OperatorStatus (operator.go:630).
    Both are still tested (TestStatus, TestStatusFailureStopsEarly, TestOperatorStatus).
    verify_ops.go:270 `func backupActivityState` has no caller. Its only mention is semp.go:270, which
    tells callers to use it: "A caller doing that must confirm the mate is standby first (showRD +
    backupActivityState)". Covered by TestBackupActivityStateReadsTheMateColumn.
    scripts.go:208 `func setReplicationEnabledScript` has no caller and no test (uncovered statements).
    replication.go:572 `func (v VPNRepl) Replicating` is used only in replication_test.go:488-499. Its
    comment says "the guard every caller wants", but no caller uses it.
    render.go:514 `func (p EnvPair) Assignment` ("podman Environment=, env-file line") is used only by
    render_test.go:160.
    The changes that replaced them (for example 72a3a79 for ShowAll/Status) left the old functions and
    their tests behind.

Why it matters: Dead code inflates the test count and coverage, and comments point maintainers at helpers
nothing uses. No staticcheck/U1000 gate catches unused unexported functions (cross-reference for the cicd
reviewer).

Change: Delete Cluster.Status, Cluster.OperatorStatus, backupActivityState, setReplicationEnabledScript,
VPNRepl.Replicating and EnvPair.Assignment. Delete or repurpose their tests (TestStatus,
TestStatusFailureStopsEarly, TestOperatorStatus, TestBackupActivityStateReadsTheMateColumn, the Replicating
assertions; inline Assignment in render_test.go:160). Remove the matching docs/test.md rows and the Summary
counts. Fix the semp.go:268-271 comment. Re-run cov and state the moved total, as the floor rule requires.

#### CR-0046 [minor] [new] set/MISSING and orNone rendering is duplicated per package and has already drifted

Location: internal/k8s/check.go:111-125, internal/container/manager.go:1537-1552,
internal/broker/importops.go:1165-1170, internal/broker/importops.go:1077-1081
Confidence: verified
Evidence:

    k8s/check.go:111-124 and container/manager.go:1537-1551 each define byte-identical orValue, orNone
    ("(none)") and setOrMissing ("MISSING"/"set").
    broker/importops.go:1165-1169: `func orNone(s string) string { if s == "" { return "none" } ... }`.
    This is used for report values at 1077-1081 ("VPNs torn down", "VPNs created").
    CLAUDE.md: "One leaf package owns every stdout and stderr shape (`internal/output`), so a convention
    cannot drift per call site" and "Reports say `set`/`MISSING`".
    So an empty value reads "(none)" in validate/status reports and "none" in import reports, and the
    secret-presence vocabulary is defined twice.

Why it matters: The set/MISSING vocabulary is a stated security-reporting convention. Keeping two copies
plus a third divergent empty marker is the per-call-site drift the output package exists to prevent.

Change: Move setOrMissing and orNone (one spelling) into internal/output, for example output.SetOrMissing
and output.OrNone. Call them from k8s, container and broker, and delete the three local copies.

#### CR-0075 [nit] [new] Comments and catalogue rows describe retired flags, commands, tests and fields

Location: internal/cli/helpers.go:17-31, internal/cli/ops_k8s.go:573-574, internal/k8s/ops_test.go:35-38,
internal/config/config_test.go:2040-2044, internal/config/config_test.go:2063-2065, docs/test.md:606,
internal/config/load.go:15-38, internal/cli/commands.go:890-912, internal/cli/context.go:213-223,
internal/cli/ops_container.go:314-323, internal/cli/ops_container.go:418-425, internal/cli/helpers.go:17-28,
internal/cli/flags.go:250-252
Confidence: verified
Evidence:

    helpers.go:17 "Used by `gen` and `--gen` paths"; :27 "what genAnnotation does for the --gen-*-only
    trio"; :31 "renderOnly(genCapable(...))". None of genAnnotation, genCapable or --gen-* exists
    (CLAUDE.md: "There is no `--dry-run` and no `--gen-*` flags").
    ops_k8s.go:573-574 "with a gen flag it prints the rendered bundle only. It backs both `operator
    deploy` and `prep operator`". There is no gen flag and no prep noun.
    ops_test.go:35-37 "matching ShowAll's identical guard (already covered by
    TestShowAllWrapsGetError)". ShowAll and that test were removed (72a3a79).
    config_test.go:2063-2065 says the user password reaches "a CLI line (broker/scripts.go's `create
    username ... password ...`)". That string is no longer in scripts.go.
    test.md:606 lists four fields for TestValidateCredentialControlChars. The test
    (config_test.go:2040-2044) also covers image.pass.

Why it matters: Comments that describe removed behaviour mislead the next change in exactly the areas
CLAUDE.md marks as settled.

Change: Rewrite helpers.go:17-31 and ops_k8s.go:573-574 for the current generate/renderOnly design. Drop the
ShowAll sentence from ops_test.go:35-38. Fix config_test.go:2063-2065 to say Secret data value / mounted
file. Add image.pass to the test.md:606 row.

### tests

#### CR-0017 [major] [new] 15 real-capture tests skip in every gate because their fixtures are gitignored and absent

Location: internal/broker/appliance_test.go:18-33, internal/broker/transform_test.go:21-34,
internal/broker/transform_test.go:425-435, internal/broker/replication_test.go:625-637, docs/test.md:64-65,
.gitignore:34, scripts/logs/test.log:1-14
Confidence: verified
Evidence:

    docs/test.md:64-65 (house rule): "Every fixture a test reads must be committed. `bash/` is
    gitignored in its entirety, so no test may point at it -- a fresh CI checkout has no such files."
    .gitignore:34: `semp`
    appliance_test.go:18-19: "The capture is gitignored (it carries a production configuration), so
    every test here SKIPS when it is absent"
    appliance_test.go:24-27: raw, err := os.ReadFile(filepath.Join("..", "..", "semp",
    "show-currentconfig-all.out")); if err != nil { t.Skipf("appliance capture not available (%v)", err)
    }
    transform_test.go:27 / 430 and replication_test.go:635 skip the same way for semp/all.cli,
    semp/all-cfg-tps-org.cli and semp/show-*.out.
    `ls semp/` in the working tree that produced today's gate log: "No such file or directory".
    The 15 tests behind these loaders: TestApplianceCaptureIsReadAsAnAppliance,
    TestEveryApplianceBrokerSectionIsClassified, TestApplianceOnlySectionsCarryContent,
    TestApplianceReplicationGrammarDiffersFromSoftware, TestPrepareForApplyIsIdempotentOnTheRealCapture,
    TestInjectVPNServiceShutdownCoversEveryRealPortLine, TestClearExistingNestedCoversAllThreeVerbs,
    TestClearExistingSyslogsAgainstTheRealCapture, TestClearExistingClientCAsIntersects,
    TestClearTargetVirtualHostnamesCoversBothHalves, TestNestedObjectsAreFoundDespiteTheParentKind,
    TestRealApplianceShowReplicationParses, TestRealSoftwareShowReplicationParses,
    TestRealVPNReplicationParses, TestRealApplianceCaptureMateRoundTrips.
    test.log: "ok  solace/internal/broker 0.226s". The run is not -v, so 15 skips look the same as 15
    passes.
    Step by step: CI checks out a tree with no semp/, so every test above skips. The local gate tree has
    no semp/ either, so they skip there too. The test task reports ok, and the 97.7% total is measured
    without them.

Why it matters: The project's own comment says these tests exist because the hand-made fixtures "is exactly
what missed them for so long". Today none of them runs in any gate, so a regression in import
classification, transforms or replication parsing against real broker shapes passes CI and the local gate
silently. This also breaks the written fixture rule in test.md.

Change: Option 1: commit scrubbed copies of the captures under internal/broker/testdata (the loaders already
parse them; replace the credentials and hostnames) and turn the t.Skipf calls into t.Fatalf. Option 2, if
the captures cannot be committed: make the test task count skips (for example `go test -json` filtered for
Action=="skip", or `-v | grep -c -- '--- SKIP'`) and print the number in scripts/logs/test.log. Also add a
sentence to docs/test.md under 'Running the tests' saying these 15 tests run only where semp/ exists, and
remove the conflict with the rule at test.md:64.

#### CR-0047 [minor] [new] Error-path tests assert only err != nil and cannot tell which step failed or whether later writes ran

Location: internal/container/manager_test.go:2240-2309, internal/container/manager_test.go:2518-2561,
internal/broker/coverage_test.go:281-300, internal/broker/coverage_test.go:690-706,
internal/broker/coverage_test.go:851-906, internal/broker/coverage_test.go:954-969, docs/test.md:2740
Confidence: verified
Evidence:

    manager_test.go:2291-2299:
      m, _, _ := newCapMgr(cfg, config.Podman)
      m.Geteuid = func() int { return 0 } // rootless as root -> guard rejects
      if err := m.Deploy(context.Background(), config.Primary); err == nil {
        t.Fatal("Deploy should fail when the euid guard rejects the host")
      }
    test.md:2740 justifies this test with: "which `Deploy` keeps because a wrong-account deploy writes
    into the secret store before anything can be undone". The test never checks the store. The runner is
    discarded (`m, _, _`) and the error text is not read.
    The same err-only shape appears in TestManagerDeployPodmanStartError (failOn("start"), a substring
    match), TestManagerDeployPodmanMkdirError, TestManagerDeletePodmanDaemonReloadError,
    TestManagerDeleteDockerComposeDownError (failOn("down")) and TestManagerDeletePurgeError. In
    coverage_test.go: TestRedundancyBackupShowError (880-882), TestRedundancyRevertToPrimaryError
    (903-905), TestReleaseToBackupNoReleaseError (966-968), TestDisableDefaultUsersDisableError
    (703-705), TestServerCertBundleReadError/CAReadError (284-299).
    Failure scenario: move checkPodmanEUID after writePodmanSecrets in Deploy (manager.go:499-508 vs
    824). The guard still returns an error, so the test still passes, but the secret store has already
    been written as the wrong account. Another example: if TestRedundancyBackupShowError's scripted
    transcript drifts and releaseToBackup times out first, Redundancy errors on the timeout and the test
    passes without reaching the Backup read it claims to cover.
    Global CLAUDE.md S5: "Always cover security and error boundaries: bad input is rejected *and* the
    downstream call is not made".

Why it matters: These tests count toward the 97.7% total but cannot tell the intended failure branch from
any other error. A reordering regression at the executor-side euid boundary, the one the catalogue says
matters, would pass.

Suggested patch (not applied):

```diff
--- a/internal/container/manager_test.go
+++ b/internal/container/manager_test.go
@@ -2291,10 +2291,17 @@
 func TestManagerDeployPodmanEUIDGuardFails(t *testing.T) {
 	cfg := ctrCfg(config.Podman, "false")
 	cfg.Podman.Rootful = false
 	cfg.Podman.QuadletDir = t.TempDir()
-	m, _, _ := newCapMgr(cfg, config.Podman)
+	m, rr, _ := newCapMgr(cfg, config.Podman)
 	m.Geteuid = func() int { return 0 } // rootless as root -> guard rejects
-	if err := m.Deploy(context.Background(), config.Primary); err == nil {
-		t.Fatal("Deploy should fail when the euid guard rejects the host")
+	err := m.Deploy(context.Background(), config.Primary)
+	if err == nil || !strings.Contains(err.Error(), "podman.rootful") {
+		t.Fatalf("Deploy err = %v, want the euid guard's refusal naming podman.rootful", err)
+	}
+	for _, c := range rr.calls {
+		joined := c.name + " " + strings.Join(c.args, " ")
+		if strings.Contains(joined, "secret") || strings.Contains(joined, "systemctl") {
+			t.Errorf("nothing may reach the secret store or systemd once the guard refuses: %s", joined)
+		}
 	}
 }
```

#### CR-0048 [minor] [new] Secret-leak tests scan only stdout; stderr narration and returned errors are never checked

Location: internal/cli/cli_test.go:453-468, internal/cli/cli_test.go:2021-2032,
internal/cli/cli_test.go:2041-2084, internal/cli/cli_test.go:4117-4126, internal/cli/context.go:342-346
Confidence: verified
Evidence:

    runRootWith (453-466) captures stdout only: `out := captureStdout(t, func() {...
    root.SetErr(io.Discard) ...})`.
    2076-2082: `out, err := runCtr(t, path, full...)` ... `if strings.Contains(out, keyMaterial) {
    t.Errorf("%v leaked the server-certificate private key to stdout", args) }`. smokeAdminPass is in
    this fixture (2056) but is never checked.
    4119-4125 (TestSecretsNeverEchoed): checks `out` (stdout) only.
    Narration goes elsewhere: context.go:343-346 `warn`/`step` write to os.Stderr through progress(),
    and broker/k8s/container Log sinks are lineSink() -> os.Stderr.
    CLAUDE.md: "Secrets ride stdin, never argv, and are never echoed back." A future `warn("...%s",
    cfg.SEMP.AdminPass)` or an error that quotes a credential would pass all three tests.

Why it matters: The leak tests guard only one of the three channels an operator captures (stdout, stderr,
the final error line). The narration channel, where formatted values usually end up, is unguarded.

Suggested patch (not applied):

```diff
--- a/internal/cli/cli_test.go
+++ b/internal/cli/cli_test.go
@@ -2075,9 +2075,13 @@
 		full := append(append([]string{}, args...), "--platform", "docker")
-		out, err := runCtr(t, path, full...)
+		var out string
+		var err error
+		errOut := captureStderr(t, func() { out, err = runCtr(t, path, full...) })
 		if err != nil {
 			t.Fatalf("%v err = %v, want nil", args, err)
 		}
-		if strings.Contains(out, keyMaterial) {
-			t.Errorf("%v leaked the server-certificate private key to stdout", args)
+		for _, secret := range []string{keyMaterial, smokeAdminPass} {
+			if strings.Contains(out, secret) || strings.Contains(errOut, secret) {
+				t.Errorf("%v leaked a secret to stdout or stderr", args)
+			}
 		}
 	}
```

#### CR-0076 [nit] [new] Coverage-only tests in coverage_test.go assert constructor defaults and trivial edges

Location: internal/broker/coverage_test.go:91-171, internal/broker/coverage_test.go:219-230,
docs/test.md:1550-1564
Confidence: verified
Evidence:

    TestNewDefaults (91-110) re-asserts New's literal defaults (2s, 60, os.Stdout).
    TestOutDefaultsToStdout (112-122) and TestSleepElapses (150-155) do similar. TestShowWritesToStdout
    (124-146) swaps the global os.Stdout without t.Cleanup, so a panic in show leaves it swapped for
    later tests.
    TestSleepZeroIntervalReturnsCtxErr (157-162) asserts `err != nil` -> fail, i.e. that it returns nil,
    the opposite of its name. docs/test.md:1558: "A zero interval with a live context returns nil".
    TestFieldLabelWithoutColon/TestLastLinesEqualCount (219-230) test single edge rows. test.md:1550
    describes the file as "Branch coverage for the paths the happy-path tests ... cannot reach".

Why it matters: These add covered lines but check nothing about behaviour. The misnamed test misleads
readers.

Change: Rename TestSleepZeroIntervalReturnsCtxErr to TestSleepZeroIntervalReturnsAtOnce. Restore os.Stdout
through t.Cleanup in TestShowWritesToStdout. Fold the constructor and edge checks into the table tests of
the functions they exercise, or drop them and record the coverage move.

#### CR-0077 [nit] [new] Tests and code cross-reference unresolvable review IDs (M1, M4, M5, M11, M12, H2, H4, H5)

Location: internal/cli/cli_test.go:1326, internal/cli/cli_test.go:2907, internal/cli/cli_test.go:3468,
internal/config/config_test.go:1201, internal/config/config_test.go:1351,
internal/config/config_test.go:2020-2031, internal/config/config_test.go:2063,
internal/k8s/checkreport_test.go:470-527, internal/container/manager_test.go:1659,
internal/container/manager.go:873
Confidence: verified
Evidence:

    cli_test.go:2907 "TestContainerRoleDetectsFromHostname covers M1."; config_test.go:2031
    "TestValidateCredentialControlChars covers M12"; manager.go:873 "\"Running\" has to be ANSWERED, not
    assumed (M4)".
    None of these IDs appears anywhere in docs/ or README (git grep for M12/H4 returns nothing), so they
    point at a review that no longer exists.
    Global CLAUDE.md S5: "Cross-reference tests by filename + test name -- never ticket/PR/commit
    numbers; those rot."

Why it matters: A reader cannot resolve the reference, and the rule says to name the test or file instead.

Change: Replace each `covers Mn/Hn` with a one-clause statement of the behaviour, or with the name of the
test or file that holds the rationale. Drop `(M4)` from manager.go:873.

### cicd

#### CR-0049 [minor] [new] tag.yml has drifted from the init-devops template: any v* tag publishes a release that GitHub marks Latest

Location: .github/workflows/tag.yml:18-20, .github/workflows/tag.yml:38-40,
.github/workflows/tag.yml:53-108, .github/workflows/tag.yml:291-295, .github/workflows/ci.yml:64-71
Confidence: verified
Evidence:

    tag.yml:38-40 triggers on `tags: ["v*"]`. The plan job (53-108) validates only SHIP_IMAGE,
    BUILD_TARGETS and SHIP_LIBRARY; nothing checks the tag's shape. Release, 291-295: `gh release create
    "$TAG" --verify-tag --title "$TAG" --notes-file notes.md` with no --prerelease and no --latest. Diff
    against ~/.claude/skills/init-devops/assets/tag.yml (last copied here in 04c1b85 on 2026-11-05; the
    template changed 2026-11-11): the template's plan has an `id: tag` step that fails unless the tag
    matches '^v[0-9]+\.[0-9]+\.[0-9]+$', computes `latest`, and passes `--latest="$LATEST"` to gh
    release create. This copy has neither. The template ci.yml also runs `all cov scan`, while ci.yml:66
    and :71 run `all scan`. tag.yml:18 says 'ALL VERBATIM. Copy this file unchanged'; global S7 says
    'Workflow templates copied/deployed as part of init-devops skill, never hand-crafted'. Step by step:
    `git push origin v1.3.0-rc1`, or a typo such as `v1.3` -> plan passes -> gates pass -> a full
    release is published, GitHub makes it Latest, and its binaries report v1.3.0-rc1. A hotfix tag on an
    older line also becomes Latest.

Why it matters: A pre-release or mistyped tag ships as the current public release, and users who download
'latest' get it. The template already has the guard; this copy is a version behind.

Change: Copy ~/.claude/skills/init-devops/assets/tag.yml and ci.yml again, unchanged, into
.github/workflows/. This brings in the plan's vX.Y.Z tag check, the latest output, metadata-action flavor
latest=false and `--latest="$LATEST"`. Then decide whether ci.yml moves to `all cov scan`. If it does, the
scripts need the contract's COV_QUEUED handling first, so `all cov` runs the suite once; if it does not,
record the deviation in docs/developer.md. Keep the repo-specific race comment out of the verbatim file.

#### CR-0050 [minor] [new] The Windows CI leg runs dev.ps1 under pwsh 7, so the PowerShell 5.1 compatibility it claims is never tested

Location: .github/workflows/ci.yml:68-71, scripts/dev.ps1:12-13, scripts/dev.ps1:36-37,
scripts/dev.ps1:238-258
Confidence: verified
Evidence:

    ci.yml:70 `shell: pwsh` runs `./scripts/dev.ps1 all scan` under PowerShell 7. dev.ps1:12-13 says
    'Windows PowerShell 5.1 compatible: ASCII only, no &&, no ternary, and no Tee-Object on native
    commands', and the script has 5.1-only workarounds: 36-37 (stderr wrapped as NativeCommandError),
    238-243 (console codepage on native output), 256-258 (5.1's utf8 BOM breaking json.Decoder).
    CLAUDE.md also says output lines 'are read in PowerShell 5.1'. Step by step: a 7-only construct (&&,
    ternary, ??, -Encoding utf8NoBOM) or a 5.1 encoding regression lands in dev.ps1 -> the Windows leg
    passes under pwsh -> `scripts\dev.ps1` breaks for a user on stock Windows PowerShell.

Why it matters: Windows PowerShell 5.1 is the shell the script is written for, and CI never runs it.

Change: Make the Windows leg run `shell: powershell`, which is Windows PowerShell 5.1 and preinstalled on
windows-2025, or add a second 5.1 step. Make the change in the init-devops ci.yml template and copy it in
again, so this file stays verbatim.

### build

#### CR-0051 [minor] [new] Local gates build the shared module from ../go-deploy-utility through an untracked go.work, while CI and releases build v0.1.2, and nothing reports it

Location: go.work:1-5, .gitignore:58-60, go.mod:12, CLAUDE.md:31-33, scripts/dev.sh:45-53,
scripts/dev.ps1:68-77, docs/developer.md:169-176, go.mod:11-12, internal/cli/context.go:10-11,
internal/cli/flags.go:7, internal/cli/commands.go:5, scripts/dev.sh:4-6
Confidence: verified
Evidence:

    go.work:1-5 contains `go 1.27`, `use .` and `replace github.com/hafio/go-deploy-utility =>
    ../go-deploy-utility`; it is untracked (.gitignore:58-60). go.mod:12 requires v0.1.2. `git -C
    ../go-deploy-utility describe` gives v0.1.2-2-gc52b128; the two extra commits are docs and graph
    only, so the code is identical today. CLAUDE.md:32-33 says the go.work is for building against the
    checkout 'until a tag is pinned', and the tag is now pinned. dev.sh:45-53 and dev.ps1:68-77 handle
    GOTOOLCHAIN but neither sets, checks nor prints GOWORK. docs/developer.md never mentions go.work,
    and :175-176 says '`full` on a clean checkout before tagging is what catches a file you never
    committed' -- a checkout that git calls clean still has the ignored go.work. The pinned module's
    docs/consuming.md:52-53 says 'The workspace hides a wrong `require` line', and step 7 (:94-95) says
    to run `GOWORK=off` 'so each tool is checked against the published tag instead of this checkout'.
    Step by step: change ../go-deploy-utility, or bump the module without updating go.mod and go.sum ->
    local `all scan`/`full` and the recorded 97.7% cov floor go green against the checkout -> CI and the
    tag build v0.1.2 from go.sum, so they test different code or go red. No log line says the local run
    was in workspace mode.

Why it matters: A local green no longer means CI and the release build the same code, and the coverage floor
may be measured on code that never ships. Today the diff is docs only, so the risk is latent until the
checkout gets ahead of the pin in code.

Change: Pick one. (a) Delete go.work now that v0.1.2 is pinned, and update CLAUDE.md:31-33. (b) Keep it but
make it visible. In both dev.sh and dev.ps1, after the GOTOOLCHAIN block, read `go env GOWORK`. When it is
non-empty and not 'off', print one warn line naming the go.work file and saying that GOWORK=off runs the
gates as CI does, and write the same line into each task log. Add a short 'Workspace mode' note to
docs/developer.md Gates/Releases saying the pre-tag `full` runs with GOWORK=off.

#### CR-0052 [minor] [new] The vet task omits the gofmt -l check the dev-script contract requires, so no gate checks formatting

Location: scripts/dev.sh:101, scripts/dev.ps1:142, docs/developer.md:51
Confidence: verified
Evidence:

    dev.sh:101 `task_vet()  { cap go vet ./...; }`; dev.ps1:142 `function Task-vet  { return (Cap go vet
    ./...) }`; developer.md:51 documents vet as `go vet ./...` only. The dev-script contract
    (init-devops references/dev-script-contract.md:12) says: 'Go: `go vet ./...` and `gofmt -l` (it
    exits 0 either way, so any listed file fails the task; use the selected toolchain's gofmt, `$(go env
    GOROOT)/bin/gofmt`)'. No other task or workflow step runs gofmt. Step by step: commit an unformatted
    .go file -> `all scan` passes locally, in ci.yml and in tag.yml -> it ships.

Why it matters: Global S7 says the scripts follow the contract, and both scripts and CI have the same gap.
Before turning it on, run `gofmt -l` once, because the gate may go red on files that already exist.

Change: In both scripts, after `go vet ./...` succeeds, run "$(go env GOROOT)/bin/gofmt" -l over the
module's package directories, not `.` (the ignored tmp/ tree may hold stray .go files). Write the list into
the vet log and fail the task when it is non-empty. In dev.ps1 use the same GOROOT gofmt through Cap and
fail on non-empty output. Update the vet row in docs/developer.md:51 and the usage text in both scripts.

#### CR-0053 [minor] [new] Nothing checks that the -X version stamp lands, so a rename would ship binaries that report 'dev'

Location: scripts/dev.sh:138, scripts/dev.ps1:179, internal/cli/version.go:10-21,
.github/workflows/tag.yml:214-228
Confidence: verified
Evidence:

    dev.sh:138 and dev.ps1:179 pass `-ldflags "-s -w -X solace/internal/cli.version=..."`. version.go:21
    has `var version = "dev"`. `git grep cli.version` finds only the comment at version.go:12; no test
    ties the -X target to the variable. The Go linker silently ignores -X for a symbol that does not
    exist. tag.yml:214-228 builds and uploads without running the binary's `version`. The init-devops
    SKILL.md ('A wrong-version binary builds green', line 360) says: 'the release leg must compare the
    binary's output against the tag before publishing - a mismatch is a failed release, not a warning.'
    Step by step: rename `version`, move internal/cli, or change `module solace` -> every gate stays
    green -> every release binary prints `solace-util dev go1.27.4 ...`.

Why it matters: The path is correct today, but nothing would catch a refactor that breaks it, and version
output is what support requests depend on (version.go:32-33).

Change: Add a test in internal/cli that reads scripts/dev.sh and scripts/dev.ps1, extracts the -X target,
and asserts it equals '<module path from go.mod>/internal/cli.version'. Also, or instead: in build_one and
Build-One, when the target os/arch equals the host, run "$out" version and fail the task unless the second
field equals VERSION. Keep the two scripts mirrored and record the test in docs/test.md.

#### CR-0054 [minor] [new] dev.sh test aborts on bash older than 4.4 when SOLACE_RACE=0 (empty array under set -u)

Location: scripts/dev.sh:1, scripts/dev.sh:11, scripts/dev.sh:37-38, scripts/dev.sh:102
Confidence: verified
Evidence:

    dev.sh:11 `set -uo pipefail`; :37-38 leave `RACE_FLAG=()` empty when SOLACE_RACE=0; :102
    `task_test() { cap go test "${RACE_FLAG[@]}" -count=1 ./...; }`. Before bash 4.4, expanding an empty
    array with "${arr[@]}" under nounset is an unbound-variable error (fixed in bash 4.4). :1
    `#!/usr/bin/env bash` resolves to /bin/bash 3.2 on stock macOS, and docs/developer.md:28 and
    docs/test.md:22-24 list macOS as a dev.sh platform. Step by step: `SOLACE_RACE=0 ./scripts/dev.sh
    test` on macOS -> 'RACE_FLAG[@]: unbound variable' -> the non-interactive shell exits at once ->
    test.log has a header and no footer, and no [warn] line is printed. The usage text at :224 already
    guards the same array with `${RACE_FLAG[*]:-}`.

Why it matters: This hits only one platform and one opt-out, but it breaks the footer contract and fails
with no explanation.

Suggested patch (not applied):

```diff
--- a/scripts/dev.sh
+++ b/scripts/dev.sh
@@ -100,3 +100,3 @@
 task_tidy() { cap go mod tidy; }
 task_vet()  { cap go vet ./...; }
-task_test() { cap go test "${RACE_FLAG[@]}" -count=1 ./...; }
+task_test() { cap go test ${RACE_FLAG[@]+"${RACE_FLAG[@]}"} -count=1 ./...; }
```

### interfaces

#### CR-0055 [minor] [new] broker stop and perform redundancy-test run with no confirmation gate, while restart and operator stop are gated

Location: internal/cli/commands.go:186-198, internal/cli/commands.go:548-558, internal/cli/ops_k8s.go:401,
internal/cli/ops_k8s.go:537-542, internal/cli/ops_container.go:87-105,
internal/cli/ops_container.go:314-326, internal/cli/ops_k8s.go:633-646, internal/cli/aliases.go:55-58,
internal/cli/aliases.go:77, docs/operations.md:1362-1416, README.md:357-358,
internal/cli/commands.go:548-559, internal/cli/ops_container.go:324-326, internal/cli/confirm.go:72-78,
docs/operations.md:1362-1366, docs/operations.md:1398-1401
Confidence: verified
Evidence:

    commands.go:188 "This takes messaging down until `broker start` brings it back." (the help carries
    the one-sentence destroy line CLAUDE.md:78 requires of a destructive command), yet 195-197 builds
    stop with no addRemoveFlags, and ops_k8s.go:542 / ops_container.go:92 call ReplicasStop / Stop
    directly. commands.go:548-551 says redundancy-test "Fails the broker over to its mate and back for
    real ... A Ctrl-C partway through can leave the group failed over", and aliases.go:77 calls it
    "invasive". It has no gate either: ops_k8s.go:401 `func opK8sVerifyRedundancy(a *App) error { return
    k8sOps(a).Redundancy(bg()) }`, and ops_container.go:325 calls RedundancyCoordinated directly. By
    contrast ops_container.go:94-97 gates restart: "A restart drops every open client connection and any
    in-flight messaging the same way a delete would, so it takes the same gate". ops_k8s.go:640-646
    gates `operator stop`, which only freezes reconciliation. operations.md:1364-1366 gates restart
    "because it drops every in-flight connection", but 1400-1401 exempts stop "since nothing is
    deleted". On docker/podman a restart deletes nothing either, so that reason does not hold. Step by
    step: (1) the kubeconfig context has drifted to prod; (2) `solace-util broker stop -e dev.yaml` runs
    from cron or a terminal; (3) the preamble prints `==> kube-context: "prod"` and ReplicasStop scales
    prod to zero. No question is asked and no --no-prompt is needed. The context-in-prompt safeguard
    (operations.md:1403-1416) never fires, because no prompt exists. redundancy-test fails the primary
    over and back the same way.

Why it matters: The two verbs that disrupt traffic the most, one of them strictly worse than restart, are
the ones that run unasked and unattended. Meanwhile restart and operator stop, which do less damage, are
gated. The documented reason for exempting stop contradicts the reason the same section gives for gating
restart. Note: `broker stop` asking nothing is deliberate and documented (docs/operations.md:1400-1401,
'since nothing is deleted'); the finding is that the same reasoning would exempt restart, which is gated,
and that redundancy-test has no stated exemption.

Change: Add addRemoveFlags(c, app, nil) to newBrokerStopCmd and newPerformRedundancyTestCmd. Gate
opK8sStopBroker with confirmAction(a, "Stop", "stop", k8sWhat(a, "broker "+a.Cfg.K8s.Name)), and gate
opCtrStopBroker with confirmAction(a, "Stop", "stop", containerWhat(a)). For redundancy-test, ask only after
the standalone skip, the role check and the health/SEMP preflight have passed, and before the first
release-activity. That needs the Ops call split into a check and an act (the LeaderCheck/LeaderAssert
pattern), because internal/broker has no Confirm seam. Then update operations.md 'Every destructive command
confirms' and README 357-358, add the confirm tests and the TestFlagsRegistered rows to docs/test.md, and
run regen for docs/commands.md.

#### CR-0056 [minor] [new] redundancy-test parses --pod in internal/broker, skipping the usage marker and the role cross-check

Location: internal/cli/ops_container.go:324-326, internal/broker/verify_local.go:46-52,
internal/broker/verify_local.go:159-170, internal/cli/ops_container.go:501-519, docs/operations.md:43-51
Confidence: verified
Evidence:

    ops_container.go:325 `return ctrOps(a).RedundancyCoordinated(bg(), a.pod)` passes the raw flag
    through. verify_local.go:160-163 runs `if o.skipIfStandalone("redundancy-test") { return nil }` and
    then `role, err := o.LocalRole(roleArg)`, and 47-48 return `config.ParseRole(roleArg)` without
    exitcode.AsUsage. Every other container --pod reader goes through detectContainerRole
    (ops_container.go:506-518): it wraps the error in `exitcode.AsUsage(err)`, and when an explicit role
    disagrees with the host it warns "--pod says %s, but this host looks like the %s node".
    operations.md:46 lists "a bad `--pod`/`--platform`/`--since` value" as exit 2. Three failures
    follow. (1) An HA docker env file with `broker perform redundancy-test --pod bogus` exits 1. (2) A
    standalone env file with `--pod bogus` returns the [SKIP] and exits 0, accepting the bad value. (3)
    `--pod primary` on a host that detects as the backup gets no warning.

Why it matters: This one --pod reader breaks the documented exit contract (a script will retry a 1) and the
explicit-role cross-check that deploy, generate and assert-leader all apply.

Suggested patch (not applied):

```diff
--- a/internal/cli/ops_container.go
+++ b/internal/cli/ops_container.go
@@ -324,3 +324,9 @@
 func opCtrVerifyRedundancy(a *App) error {
-	return ctrOps(a).RedundancyCoordinated(bg(), a.pod)
+	// --pod through the CLI's own parser and cross-check, as deploy and assert-leader
+	// do: a bad value is a usage error, and a role that disagrees with the host warns.
+	role, err := detectContainerRole(a, a.pod, true, "broker perform redundancy-test")
+	if err != nil {
+		return err
+	}
+	return ctrOps(a).RedundancyCoordinated(bg(), role.Word())
 }
```

#### CR-0057 [minor] [new] Declining the --out overwrite prompt exits 1, against the house rule that n exits 0

Location: internal/cli/helpers.go:147-160, internal/cli/confirm.go:9-17,
internal/cli/ops_container.go:223-225, docs/operations.md:1374-1377
Confidence: verified
Evidence:

    helpers.go:153-159: `ok, err := confirmAction(a, "Overwrite", ...)`, then `if !ok { return
    fmt.Errorf("refusing to overwrite %q", a.out) }`. Gate returns (false, nil) only for an answered
    "n", because no-terminal comes back as err. So a decision becomes an error and main exits 1.
    CLAUDE.md:177-178 says "A reply of `n` is a decision and exits 0." ops_container.go:225 says "A
    declined or unanswered question returns nil, as every other asker in this tree does."
    operations.md:1376-1377 says "Declining at a terminal is an answer, not a failure: it prints nothing
    more and exits 0." Step by step: x.yaml exists -> `solace-util broker generate -o x.yaml` at a
    terminal -> answer n -> `Error: refusing to overwrite "x.yaml"`, exit 1. examples, convert and
    perform export-config share emitOrWrite and do the same. TestExportConfigOutOverwriteGate
    (docs/test.md:1205) covers no-terminal, yes and --no-prompt, but not a declined prompt.

Why it matters: This is the one asker in the tree that reports a decision as a failure. The exit code then
tells a wrapper something went wrong when the operator chose to keep the file.

Change: Return a sentinel such as errKeptExisting, or (written bool, err), from emitOrWrite. Callers map the
sentinel to nil and skip their post-write step lines: runExample's CHANGE-ME line and runConvert's review
line would otherwise claim a write that did not happen. Print an [INFO] "kept x.yaml; nothing was written".
Add the declined case to TestExportConfigOutOverwriteGate and TestGenerateToFileConfirmsAnOverwrite.

#### CR-0058 [minor] [new] --tail reaches kubectl/docker/podman unvalidated, and its help advertises `all`

Location: internal/cli/commands.go:818-840, internal/cli/ops_container.go:354-362, docs/operations.md:43-51
Confidence: verified
Evidence:

    commands.go:820 declares `"print only this many trailing lines (a count, or `all`)"` on a flag
    shared by broker logs and operator logs on every platform. 831 says "--tail is passed as given
    (kubectl validates it)", and 839 does `Tail: a.tail`. ops_container.go:360 also passes `Tail:
    a.tail` straight through. The pinned shared contract says otherwise: go-deploy-utility@v0.1.2
    kube/kube.go:250-251 "Tail and Since are passed as given, so the tool validates them first", and
    containercli/containercli.go:295-296 says the same. This app canonicalises --since (logSince) but
    never checks --tail. Step by step: `broker logs --tail ten` -> argv `... logs ... --tail ten` -> the
    child refuses it -> exit 1. operations.md:43-51 classifies a bad --since as 2 and does not mention
    --tail. `--tail all` works for docker, whose --tail is a string. kubectl and podman declare --tail
    as an integer, so the value the help advertises fails on two of the three platforms.

Why it matters: A bad value is reported as a runtime failure (retryable 1) instead of a usage error, which
breaks the shared module's contract. The `all` half depends on the external CLIs: confirm with `kubectl logs
<pod> --tail all` and `podman logs --tail all <ctr>`.

Change: Add a logTail(a) beside logSince that accepts "" or a non-negative integer and otherwise returns
exitcode.Usagef naming the value. Call it from logOptions and opCtrLogs. Change the usage to "(a count; omit
it for every line)", or map `all` per engine. Add the bad-value row to TestExitCodeContract and list --tail
in operations.md's exit-2 sentence. Regen docs/commands.md.

#### CR-0059 [minor] [new] validate promises every problem in one pass, but env-file errors stop at the first

Location: internal/cli/commands.go:109-116, internal/config/validate.go:15-126,
internal/config/validate.go:630-653, internal/cli/platform.go:187-189, README.md:286
Confidence: verified
Evidence:

    commands.go:110-111 says "Reports every problem it finds in one pass, rather than stopping at the
    first." validate.go:16-17 says "It fails loud with an actionable message listing every offending
    field at once." In the code, every step is `if err := ...; err != nil { return err }`
    (validate.go:33-125), and validateK8s returns at the first bad DNS label (650-652). Only
    requireAll's missing-field group is collected (630-631). Load runs in PreRunE (platform.go:187), so
    a file error ends `validate` with exit 2 before any report row prints. Step by step: an env file
    with a bad `timezone` and a bad `kubernetes.namespace` -> `solace-util validate` names only the
    first -> fix it and re-run to see the second. README:286 describes validate as "Check the whole env
    file".

Why it matters: The help text and the exported godoc describe behaviour the loader does not have. The fuller
fix is to collect errors with errors.Join in Validate; the patch below makes the text true at least.

Suggested patch (not applied):

```diff
--- a/internal/cli/commands.go
+++ b/internal/cli/commands.go
@@ -108,5 +108,5 @@
 
 const brokerValidateLong = "Read-only, and safe to run against a system you don't want to disturb.\n\n" +
-	"Reports every problem it finds in one pass, rather than stopping at the\n" +
-	"first."
+	"Reports every cluster or host problem it finds in one pass. The env file is\n" +
+	"checked first, at load, and stops at its first error."
 
--- a/internal/config/validate.go
+++ b/internal/config/validate.go
@@ -15,3 +15,4 @@
 // Validate checks mandatory and enum fields for the given platform, mirroring
-// the mandatory-vars and enum checks in the two bash bootstraps. It fails loud
-// with an actionable message listing every offending field at once.
+// the mandatory-vars and enum checks in the two bash bootstraps. It fails loud
+// at the first offending rule with an actionable message; only a group of missing
+// mandatory fields is listed together.
```

#### CR-0060 [minor] [new] Help and README text describe commands and behaviour that do not exist

Location: internal/cli/convert.go:28-34, docs/commands.md:1049, internal/cli/completion.go:48-50,
README.md:118-120, internal/cli/flags.go:221-224, internal/cli/ops_k8s.go:110-113,
internal/cli/root.go:28-31, internal/cli/commands.go:864-866, internal/cli/convert.go:29-31
Confidence: verified
Evidence:

    (a) convert.go:31 `"  solace-util check deploy -e prod.yaml"`: there is no `check` command, so the
    suggested next step is an unknown top-level command, which takes the documented exit-1 gap. The text
    is copied into generated docs/commands.md:1049. (b) README.md:119 says "`solace-util auto-complete`
    on its own prints all of this", but completion.go:49 returns `fmt.Errorf("%s needs a shell: ...")`.
    So bare auto-complete prints an error and exits 1, while `auto-complete tcsh` exits 2 through
    markUsageArgs. Both are wrong command lines. (c) flags.go:223 says --restart acts "for a changed
    artifact, secret or server certificate and also for an unchanged one". On Kubernetes,
    ops_k8s.go:110-112 `if !certChanged || !c.BrokerPodsExist(ctx) { return nil }`, so `broker deploy
    --restart` with an unchanged certificate restarts nothing. (d) root.go:29 says "Every command takes
    -e/--env <file>". examples, convert, version and auto-complete accept the persistent -e/--base-dir
    and silently ignore them, which commands.go:864-866 calls "worse than a refused one".

Why it matters: The help text and the README are the interface a runbook quotes, and these four statements
are false for the code as written. Run the regen task after this patch so docs/commands.md follows.

Suggested patch (not applied):

```diff
--- a/internal/cli/convert.go
+++ b/internal/cli/convert.go
@@ -28,7 +28,7 @@
 			"never commit it.\n\n" +
 			"  solace-util convert bash/env/prod -o prod.yaml\n" +
 			"  solace-util convert bash/env/prod --platform podman -o prod.yaml\n" +
-			"  solace-util check deploy -e prod.yaml",
+			"  solace-util validate -e prod.yaml",
 		Args:          cobra.ExactArgs(1),
 		SilenceUsage:  true,
 		SilenceErrors: true,
--- a/internal/cli/completion.go
+++ b/internal/cli/completion.go
@@ -7,6 +7,7 @@
 	"path/filepath"
 	"strings"
 
+	"github.com/hafio/go-deploy-utility/exitcode"
 	"github.com/spf13/cobra"
 
 	"solace/internal/config"
@@ -46,7 +47,7 @@
 		// instead, and makes the bare command say what to pass. `--help` is unaffected:
 		// the help flag short-circuits ahead of both.
 		RunE: func(cmd *cobra.Command, _ []string) error {
-			return fmt.Errorf("%s needs a shell: bash, zsh, fish, or powershell", cmd.CommandPath())
+			return exitcode.Usagef("%s needs a shell: bash, zsh, fish, or powershell", cmd.CommandPath())
 		},
 	}
 	c.AddCommand(
--- a/internal/cli/flags.go
+++ b/internal/cli/flags.go
@@ -221,4 +221,4 @@
 func addRestartFlag(c *cobra.Command, app *App) {
 	c.Flags().BoolVar(&app.restart, "restart", false,
-		"restart or recreate an already-running broker without asking, for a changed artifact, secret or server certificate and also for an unchanged one (otherwise docker/podman ask, and fail with no terminal unless --no-prompt is given; Kubernetes asks only about a renewed certificate)")
+		"restart or recreate an already-running broker without asking. Docker/podman: for a changed artifact, secret or server certificate and also for an unchanged one (otherwise they ask, and fail with no terminal unless --no-prompt is given). Kubernetes: only for a renewed server certificate, the one change it otherwise asks about")
 }
--- a/internal/cli/root.go
+++ b/internal/cli/root.go
@@ -28,4 +28,5 @@
 			"\n" +
-			"Every command takes -e/--env <file>; the platform comes from that file, or from\n" +
-			"--platform kubernetes|docker|podman when it declares more than one.\n" +
+			"Every broker, operator and validate command takes -e/--env <file>; the platform\n" +
+			"comes from that file, or from --platform kubernetes|docker|podman when it\n" +
+			"declares more than one.\n" +
 			"\n" +
--- a/README.md
+++ b/README.md
@@ -118,3 +118,3 @@
 
-`solace-util auto-complete` on its own prints all of this, and each shell's own `--help`
+`solace-util auto-complete --help` prints all of this, and each shell's own `--help`
 adds whatever that shell needs first (zsh wants `compinit` enabled; bash needs nothing).
```

### observability

#### CR-0061 [minor] [new] User-facing errors name retired verb-first commands (remove broker, deploy operator)

Location: internal/config/validate.go:1575-1580, internal/k8s/operatorversion.go:162-163, docs/test.md:761
Confidence: verified
Evidence:

    validate.go:1576-1577: `return fmt.Errorf("%s %q must be an absolute path: it is the host side of a
    bind mount, and it is what "+ "`remove broker --delete-data` empties -- ...`
    operatorversion.go:162-163: `return fmt.Errorf("operator downgrade declined; ... point that at the
    installed version, or skip `deploy operator`")`
    CLAUDE.md: "Noun then verb, and nothing implicit ... `broker deploy`, `operator remove` ... give it
    a word it does not own and it fails loud".
    docs/test.md:761 says this error names "`broker remove --delete-data`", which the code does not do.
    Step by step: a load error tells the operator about `remove broker --delete-data`. They type
    `solace-util remove broker ...`, and root does not own `remove`, so they get a usage error instead
    of the command they were pointed at.

Why it matters: The 'what to do next' half of these two error messages sends the operator to a command that
does not exist.

Suggested patch (not applied):

```diff
--- a/internal/config/validate.go
+++ b/internal/config/validate.go
@@ -1576,3 +1576,3 @@
 		return fmt.Errorf("%s %q must be an absolute path: it is the host side of a bind mount, and it is what "+
-			"`remove broker --delete-data` empties -- a relative value would be read by podman as a named "+
+			"`broker remove --delete-data` empties -- a relative value would be read by podman as a named "+
 			"volume and would leave a recursive delete pointing at whatever directory the command was run "+
--- a/internal/k8s/operatorversion.go
+++ b/internal/k8s/operatorversion.go
@@ -162,2 +162,2 @@
 	return fmt.Errorf("operator downgrade declined; nothing was applied. The operator already installed is " +
-		"newer than kubernetes.operator.image -- point that at the installed version, or skip `deploy operator`")
+		"newer than kubernetes.operator.image -- point that at the installed version, or skip `operator deploy`")
```

#### CR-0062 [minor] [new] The final error line uses a third stderr shape outside internal/output, and the docs show it differently

Location: main.go:20-25, internal/output/output.go:1-17, docs/operations.md:1963-1968,
docs/operations.md:2019-2021
Confidence: verified
Evidence:

    main.go:22-23: `if err != nil { fmt.Fprintln(os.Stderr, "Error:", err) }`
    CLAUDE.md: "One leaf package owns every stdout and stderr shape (`internal/output`) ... A stderr
    line is `==> <phase>` or `[TAG ] <msg>` and never both."
    output.go has no shape for this line, and nothing in docs or tests pins "Error:" (git grep for
    '"Error:' finds only main.go and broker transcript fixtures).
    docs/operations.md:1964-1965 documents the line as:
      [FAIL] nr_open: 65536 (need 1048576)
      error: fs.nr_open is 65536, needs 1048576.
    The binary prints "Error: ...", so the doc's lowercase `error:` matches no real output. The most
    important stderr line of a failed run follows a convention nobody defined, and anyone grepping logs
    for the documented spelling misses it.

Why it matters: Captured stderr is how a failed deploy gets diagnosed. Its terminal line is the one line
that bypasses the house format, and the docs give it a spelling the tool never prints.

Change: Either emit the final line through the output package (main may import internal/output, which is
stdlib-only), for example `output.New(os.Stderr).Fail("%v", err)`, and update the two operations.md
examples. Or keep "Error:" deliberately, document it in output.go's package comment as the one terminal
shape, and fix operations.md:1965 and 2020 to `Error:`.

### conventions

#### CR-0063 [minor] [new] Real env files under env/ are not gitignored; the old protection only covered the removed docker-podman/env layout

Location: .gitignore:9-12, .gitignore:25, .gitignore:43-48, README.md:135-151
Confidence: verified
Evidence:

    .gitignore:9 '# docker-podman: generated compose file + real env files (hold admin password / PSK)',
    :11 'docker-podman/env/*', :12 '!docker-podman/env/sample'. The docker-podman/ directory no longer
    exists. .gitignore:25 'env.yaml' matches only a file named exactly env.yaml. README.md:135
    'solace-util examples --platform kubernetes -o env/dev.yaml' and :145 'cp env/sample.yaml
    env/dev.yaml'. README.md:156 tells the user to set semp.adminPass in that file. `git check-ignore -v
    env/dev.yaml env/prod.yaml` matches nothing; only env/env.yaml is caught, by line 25. Step by step:
    follow the quick start in a checkout -> env/dev.yaml holds semp.adminPass, the redundancy PSK and
    the registry password -> `git add .` or `git add env` stages it -> a push publishes the credentials.
    By contrast, .gitignore:43-48 ignores /*.cli exports exactly because 'an artifact must never be
    committed by accident'.

Why it matters: When the layout moved from docker-podman/env to env/, the secret-bearing env files lost
their ignore rule. The README now tells users to create them in the checkout, so one broad `git add` can
commit an admin password or PSK.

Suggested patch (not applied):

```diff
--- a/.gitignore
+++ b/.gitignore
@@ -8,7 +8,7 @@
 
-# docker-podman: generated compose file + real env files (hold admin password / PSK)
-docker-podman/docker-compose.yml
-docker-podman/env/*
-!docker-podman/env/sample
+# Real env files hold the admin password, PSK and registry credentials; only the
+# generated, credential-free sample is tracked.
+/env/*
+!/env/sample.yaml
 
 # Go build / dev tooling (scripts/dev.sh|dev.ps1)
```

#### CR-0064 [minor] [new] The race flag differs between dev.sh and dev.ps1 and from the contract, so the Windows CI leg never runs -race

Location: scripts/dev.sh:35-38, scripts/dev.ps1:58-61, scripts/dev.sh:4, scripts/dev.ps1:4-5,
docs/developer.md:45-53, .github/workflows/ci.yml:59-63
Confidence: verified
Evidence:

    dev.sh:38 `if [[ "${SOLACE_RACE:-1}" != "0" ]]; then RACE_FLAG=(-race); fi` against dev.ps1:61 `if
    ($env:SOLACE_RACE -eq '1') { $RaceFlag = @('-race') }`. dev.sh:4, dev.ps1:4-5 and developer.md:45
    all say the scripts are 'behaviourally identical'. ci.yml:59-63 says the Windows leg is the non-race
    one because '-race needs cgo and a C toolchain'. Contract :13: 'Go runs them under the race detector
    (`CGO_ENABLED=1 go test -race -count=1 ./...`) wherever gcc is on PATH, both CI runners included;
    without one it runs without `-race` and says so in `test.log`'. Results: (1) the Windows leg of
    every tag never race-tests; (2) SOLACE_RACE=true or yes turns race on in dev.sh and leaves it off in
    dev.ps1; (3) dev.sh on a Linux host without gcc fails `test` outright (-race requires cgo) instead
    of running without it and logging that.

Why it matters: The divergence is documented, but its stated reason does not match the contract, and the
init-devops skeleton (dev_sh.skeleton:98) says gcc is on both CI runners. Confirm from the windows-2025
runner image manifest that MinGW gcc is on PATH.

Change: In both scripts, use the contract's rule: turn -race on when gcc is on PATH (`command -v gcc` /
`Get-Command gcc`), set CGO_ENABLED=1 for that run, and otherwise run without it and write 'go test: -race
skipped, no gcc on PATH' into test.log. Parse SOLACE_RACE the same way in both (for example 0 forces off, 1
forces on, unset means auto). Update the ci.yml comment, docs/developer.md:53 and docs/test.md:28-29.

#### CR-0065 [minor] [new] gen-cert.sh hides every openssl error, overwrites an existing CA without asking, and splits a scalar san into characters

Location: gen-cert.sh:19, gen-cert.sh:58, gen-cert.sh:66-72, gen-cert.sh:161-164, gen-cert.sh:170,
gen-cert.sh:219, gen-cert.sh:321-322, gen-cert.sh:334-366
Confidence: verified
Evidence:

    gen-cert.sh:336-341 runs `openssl req -x509 ... -keyout "${CA_PREFIX}.key" -out "${CA_PREFIX}.crt"
    ... 2>/dev/null`, then checkExit (66-69) prints only 'Failed'. Lines 354/355 and 365/366 do the
    same. :19 `DEF_CA_GENERATE="true"`, and nothing checks for an existing file before :337, so an
    existing CA.key and CA.crt in the working directory are overwritten. :161 `for s in (get(doc, "ca",
    "san") or []):` -- a scalar `san: host.example.com` is iterated per character into DNS:h,DNS:o,...
    :321-322 `-noenc` and :365 `-copy_extensions` need OpenSSL >= 3.0, and :170 `declare -gA` and :219
    `${__in,,}` need bash >= 4. On macOS (LibreSSL, bash 3.2) it fails, and with stderr discarded the
    user sees only 'Failed'. :58 points at https://github.com/hafio/solace-psg/tree/main/ssl-certs. `git
    grep gen-cert` finds no doc in this repo that mentions the script, and docs/developer.md's layout
    table (183-203) leaves it out. Global rule 4a: 'Errors are actionable: what failed, why ...'.

Why it matters: A tracked helper that can destroy an existing CA key and fails with no reason breaks the
actionable-errors rule. No doc in this repo explains how to use it.

Change: Capture openssl's stderr in a variable and print it when a step fails. Refuse to overwrite an
existing <ca-prefix>.key or .crt unless the user confirms or passes an explicit flag. In the Python parser,
reject a non-list `san`. Check `openssl version` (>= 3.0) and BASH_VERSINFO (>= 4) up front with a clear
message. Then document the script in README or developer.md, or move it to the repo its usage points at.

#### CR-0066 [minor] [new] docs/test.md and CLAUDE.md are out of date on the CI gate, the regen package list and the shared packages

Location: docs/test.md:20-26, docs/test.md:45-61, CLAUDE.md:31-33, scripts/dev.sh:121
Confidence: verified
Evidence:

    docs/test.md:26 calls `all` 'CI's gate', but CI runs `all scan` (ci.yml:66,71). test.md:45-47 says
    'Four packages carry golden files ... The `regen` task does all four', and 56-61 list render, k8s,
    convert and cli. regen actually walks six packages (dev.sh:121: examples, cli, convert, render, k8s,
    broker; developer.md:95-96). The omitted ./internal/examples must run first, per CLAUDE.md: '`regen`
    runs `./internal/examples` **first**'. CLAUDE.md:31's shared-code list leaves out runner (38
    imports), kube with kube/rbac and kube/nsgate, confirm and watch, all of which internal/ imports.
    Step by step: edit internal/examples/assets/full.yaml, then follow test.md's per-package commands ->
    the render and k8s goldens are rewritten from a stale env/sample.yaml.

Why it matters: S6 says docs stay in sync. The regen instructions, followed as written, leave goldens out of
date.

Change: In docs/test.md, change the CI row to `all scan`. Change 'Four packages' to the six regen walks, and
list `go test ./internal/examples -update` first and `./internal/broker` last. In CLAUDE.md:31, add runner,
confirm, kube (rbac, nsgate) and watch to the shared-code list.

#### CR-0067 [minor] [new] A script(1) terminal capture named 'typescript' is tracked at the repo root

Location: typescript:1-5
Confidence: verified
Evidence:

    typescript:1 'Script started on 2026-10-23 07:00:14+08:00 [TERM="xterm-256color" TTY="/dev/pts/0"
    COLUMNS="117" LINES="39"]'. Line 2 contains OSC and CSI escapes and the prompt '[hafio@erika
    solace-cnt-scripts]$ exit'. Line 5 'Script done ...'. It was added in f239d14, a feature commit
    about namespace-scoped identities. It is the only tracked file with control bytes (git grep), and
    `git ls-files --eol` reports it as `i/-text`. Global rule: plain ASCII only; checklist: no stray
    artifacts tracked.

Why it matters: It is an accidental artifact that carries escape sequences and a local user and host name,
and it would be ingested on the next graphify refresh unless excluded.

Change: Run `git rm typescript` and add `/typescript` to .gitignore.

#### CR-0078 [nit] [new] Non-ASCII bytes in two tracked text files

Location: .graphifyignore:1, internal/config/hostpath_test.go:113-114
Confidence: verified
Evidence:

    `LC_ALL=C git grep -P '[^\x00-\x7F]'` lists .graphifyignore:1 '# .graphifyignore <U+2014> files to
    exclude ...' (an em dash). It also lists hostpath_test.go:113 `{"certs/tls<U+200B>.crt", ...}` and
    :114 `{"certs/<U+202E>tls.crt", ...}`: a literal zero-width space and a literal right-to-left
    override in Go source. Global rule: 'Plain ASCII only -- everything you write: ... code, comments'.

Why it matters: Invisible and bidi characters in source cannot be seen in review and display unpredictably.
Go's \u escapes produce the same strings.

Change: Write the test strings as "certs/tls<U+200B>.crt" and "certs/<U+202E>tls.crt", and replace the em
dash in .graphifyignore:1 with '--'.

#### CR-0079 [nit] [new] README version sample shows the pre-patch toolchain, and the version command's Long exceeds four lines

Location: README.md:77-79, go.mod:5-9, internal/cli/version.go:32-36
Confidence: verified
Evidence:

    README.md:78 shows `solace-util v1.2.3 go1.27.0 linux/amd64`. go.mod:5-8 says go1.27.0 is what
    'shipped a net panic bug into release binaries', and releases now build with go1.27.4.
    version.go:32-36 sets a Long of five text lines that includes mechanics ('built by scripts/dev.sh or
    dev.ps1'). CLAUDE.md:76 says 'One to four lines saying what the command does and how to invoke it;
    mechanics and history belong in docs/operations.md'.

Why it matters: The example implies releases still ship the toolchain go.mod says is broken, and the Long
breaks a written help-text rule.

Change: Show go1.27.4 in README.md:78. Trim the version Long to at most four lines with no build mechanics,
then run regen so docs/commands.md follows.

#### CR-0080 [nit] [new] Stale ignore entries, and a future docs/review/ is covered by neither .graphifyignore nor the notice test

Location: .graphifyignore:10-15, .graphifyignore:33-35, .gitignore:10-12
Confidence: verified
Evidence:

    Paths named in the ignore files are absent from the tree: docker-podman/ (.gitignore:10-12),
    cliscripts-templates/nppBackup (.graphifyignore:11-14) and howto-deploy-solace.html
    (.graphifyignore:33-35). docs/review/ does not exist yet, but this review writes there. It is not in
    .graphifyignore, and the notice check (TestEveryDocCarriesTheSupportNotice -> noticetest default,
    noticetest.go:28-29 'Empty means README.md plus docs/*.md, non-recursively') never looks at a docs/
    subdirectory. CLAUDE.md:16-17 says 'Every doc opens with the README's notice ... a new doc needs it
    too'.

Why it matters: The next graphify refresh will ingest review reports as project content, and a doc placed
under docs/<sub>/ escapes the notice rule without any test failing.

Change: Add `docs/review/` to .graphifyignore. Either exempt docs/review in CLAUDE.md's notice rule, or pass
noticetest.Options.Files to cover subdirectories. Drop the dead ignore entries.

## Suspected (needs confirmation)

Area: correctness

#### CR-0081 [major] [new] RewriteAdminUsername discards every operand on the admin opener, so verification demands what was dropped

Location: internal/broker/transform.go:205-232, internal/broker/transform_test.go:199-208,
internal/broker/sections.go:467-476, internal/broker/importops.go:985-995,
internal/broker/importops.go:999-1031, internal/broker/diff.go:353-401
Confidence: suspected
Evidence:

    transform.go:206 `^(\s*)create username "admin" ` requires a space after the name, so it matches
    only an opener that has operands after it.
    transform.go:226 `lines[openerIndex(b.Lines)] = m[1] + "username admin"` replaces the whole opener,
    so every one of those operands is discarded.
    The repo's own test pins the loss. transform_test.go:200-203 feeds `create username "admin" password
    xyz` and wants exactly `username admin`.
    sections.go:470-475 says this section "Overwrites the target's CLI admin password with the
    source's", and operations.md:1636-1645 builds a Kubernetes hazard on that claim.
    ImportVerify (importops.go:994) diffs the UNtransformed artifact. importIgnore (1006-1030) has no
    exemption for this opener. The target's admin opener still carries the target's own operands, so the
    artifact opener is reported as Differing.
    Failure: source and target have different admin passwords. The import applies `username admin` with
    no password, so the password is not carried over, despite the docs. Verification then fails with
    `import verification failed`, and RenderDiffResult (diff.go:395-399) prints the artifact's admin
    opener, password operand included, to stdout.

Why it matters: Either the admin password the docs say is overwritten is silently not applied, or every
cross-broker import fails verification after a clean apply. Either way the feature's only verdict becomes
noise that operators learn to ignore. To confirm, check the admin opener in a real capture: the test fixture
models it as `create username "admin" password ...`.

Change: In RewriteAdminUsername (transform.go:216-232), keep what follows `create username "admin"` on the
opener instead of discarding it. Translate it into the existing-user form the broker accepts, after
confirming that form on a live broker against a real capture. Then either verify the result honestly, or
exempt exactly the rewritten opener in importIgnore (importops.go:999-1031) the way BridgeEnablementInverted
exempts bridge lines. Replace transform_test.go:199-208, which asserts the loss, with a test that the
password operand survives.

To confirm: The admin opener's shape in a real capture: whether `create username "admin"` carries operands
on the opener line. sections.go cites semp/all.cli:284-286, which is not in the tree.

Area: correctness

#### CR-0082 [major] [new] Destructive-prompt context clause can name the wrong cluster when kubernetes.command carries --context

Location: internal/cli/context.go:253-266, internal/cli/ops_k8s.go:835-851,
internal/config/config.go:288-292, docs/configuration.md:607-610, docs/operations.md:2068-2079
Confidence: suspected
Evidence:

    context.go:257-264:
      kc, err := a.Cfg.ClusterCommand()
      out, err := engine.Output(bg(), a.Runner, runner.Cmd{Argv: kc.Argv("config", "current-context")})
      ... a.kubeContext = ctx; step("kube-context: %q", ctx)
    config.go:288-291 Argv appends the call's args after every configured token, so `kubernetes.command:
    kubectl --context prod -n solace` (the documented example, configuration.md:608) runs `kubectl
    --context prod -n solace config current-context`. kubectl's `config current-context` reads the
    kubeconfig file's current-context and does not apply the global --context override, so it prints
    e.g. `dev` while every real call goes to `prod`. k8sWhat/k8sContext (ops_k8s.go:835-851) then put
    that value into every destructive prompt: `Delete broker x in namespace solace (context "dev")?
    [y/N]` -- for a delete that lands in prod. docs/operations.md:2077-2079 even tells the operator the
    tool has no --context of its own, overlooking the one in kubernetes.command. No test covers a
    --context-bearing command (TestK8sPromptsNameNamespaceAndContext injects a.kubeContext directly).

Why it matters: This clause exists specifically so an operator can catch a wrong-cluster delete; reporting
the kubeconfig default instead of the effective context turns it into a false reassurance, which
confirm.go:76-77 itself calls worse than no prompt. Confirm with `kubectl --context X config
current-context` against a kubeconfig whose current-context is Y (expected output: Y).

Change: In announceKubeContext, resolve the EFFECTIVE context: if kc carries a `--context <v>` /
`--context=<v>` token, report <v>; otherwise ask kubectl in a form that applies the override (e.g. `config
view --minify -o jsonpath={.current-context}`, which honours the global --context -- verify against the
supported kubectl/oc versions). Add a test with kubernetes.command `kubectl --context prod` asserting the
prompt names prod, and correct docs/operations.md:2077-2079.

To confirm: kubectl --context X config current-context against a kubeconfig whose current-context is Y;
printing Y confirms it.

Area: resilience

#### CR-0083 [major] [new] broker restart/start wait with `rollout status statefulset`, which kubectl refuses for OnDelete StatefulSets

Location: internal/k8s/ops.go:129-191, internal/cli/ops_k8s.go:116-134, docs/operations.md:123
Confidence: suspected
Evidence:

    ops.go:138  c.kubectl(ctx, "delete", "pod", "-n", c.ns(), pod, "--ignore-not-found")
    ops.go:143  c.waitRollout(ctx, "statefulset", sts, c.ns())   -> kube.RolloutStatus: rollout status
    statefulset/<sts> --timeout=300s
    ops.go:159-164  only a Forbidden stderr is downgraded; any other error is returned
    ops.go:186-188  ReplicasStart returns on the first failed wait: "%s did not become ready within %s:
    %w"
    ops.go:171  RestartRolling stops at the first failing role
    kubectl's StatefulSet status viewer returns "rollout status is only available for RollingUpdate
    strategy type" (exit 1) when spec.updateStrategy.type is OnDelete. The legacy
    replicas-start-broker.sh polled `get pods | grep Running` instead, and no test runs against a real
    StatefulSet.
    Failure, if the Solace operator renders its broker StatefulSets with OnDelete (it orders
    automatedRolling restarts itself): `broker restart` deletes the monitor pod, the wait fails at once,
    and the backup and primary are never restarted. `broker start` in HA scales the primary up and
    fails, leaving the backup and monitor at 0. The certificate-renewal restart after `broker deploy`
    stops after the first pod. Each failure is reported as "did not become ready within 300s".

Why it matters: The whole manualPodRestart and start flow would fail after acting, leaving a half-started or
half-restarted HA group. Confirm on a live operator-managed broker with: kubectl get sts <name>-pubsubplus-p
-n <ns> -o jsonpath='{.spec.updateStrategy.type}'.

Change: In internal/k8s/ops.go, wait on the role's pod rather than the StatefulSet rollout. After the delete
or scale, poll `get pod <pod> -o jsonpath={.status.conditions[?(@.type=="Ready")].status}` until it reads
True, treating NotFound as "not yet", or use `kubectl wait pod/<pod> --for=condition=Ready --timeout=300s`
after the pod reappears. Bound it by kube.RolloutTimeout and keep the Forbidden-to-warning rule. Add a test
whose runner fails `rollout status` with the OnDelete message and asserts that every role is still restarted
or scaled.

To confirm: kubectl get sts <name>-pubsubplus-p -n <ns> -o jsonpath='{.spec.updateStrategy.type}' on an
operator-managed broker; OnDelete confirms it.

Area: correctness

#### CR-0084 [minor] [new] Bare 'busy' keyword in the tail scan can fail a clean chunk deterministically

Location: internal/broker/driver.go:66-99, internal/broker/driver.go:184,
internal/broker/matechannel.go:269-281, internal/broker/transform.go:244-286
Confidence: suspected
Evidence:

    driver.go:94-99 failKeywords = {"invalid command", "error:", "command not found", "busy"}, matched
    case-insensitively by `grep -qiE` (184) over the last 10 transcript lines.
    driver.go:68-72 says the transcript echoes the configuration ("Scanning the whole transcript would
    instead scan the configuration itself"), so a SUCCESSFUL chunk's tail is its last few configuration
    lines.
    "busy" is a bare substring, not a phrase. ReorderVPNProfiles (transform.go:256-286) moves
    client-usernames and profiles to the end of each VPN chunk, so names like `client-username
    "busy-svc"` or a topic `.../busy/...` land in exactly that tail.
    Failure: CHUNK-FAIL, `exit 1`, every later chunk skipped, VPNs left torn down. The chunk content is
    deterministic, so the documented re-run fails identically. rejectionIn (matechannel.go:269-281)
    shares the same list for RunCLI.

Why it matters: A false rejection after the teardown can never be cleared by re-running. To confirm, capture
a real successful `source script ... stop-on-error` transcript and check whether the echoed commands appear
in its last 10 lines.

Change: Match the keywords only against lines the broker printed, not echoed commands: skip lines carrying
the CLI prompt echo (`<router>(...)# ` / `> `) in both the grep pipeline (driver.go:184) and rejectionIn.
Alternatively, anchor `busy` to the broker's actual rejection wording once a real one is captured. Add a
driver test whose successful tail contains an object named `busy-svc`.

To confirm: Capture a real successful `source script ... stop-on-error` transcript and check whether echoed
commands containing 'busy' can appear in its last 10 lines.

Area: correctness

#### CR-0085 [minor] [new] Credential-bearing artifact lines are printed to stdout by the failure detail and the verification report

Location: internal/broker/importops.go:1135-1147, internal/broker/driver.go:184-192,
internal/broker/diff.go:164-169, internal/broker/diff.go:395-401, internal/broker/sections.go:357-362,
internal/broker/sections.go:477-486, internal/broker/config_ops.go:270-280
Confidence: suspected
Evidence:

    importops.go:1142-1145 prints FailedDetail (the last 10 transcript lines, driver.go:189) line by
    line to stdout.
    diff.go:164-169: an ABSENT block puts every one of its lines into bd.Missing, and reportBlockLines
    (395-399) prints each to stdout.
    The artifact carries credential material in the broker's encrypted, replayable form: the config-sync
    PSK (sections.go:359), LDAP aes-password (481), CLI user passwords.
    The codebase's own rule at config_ops.go:276-277 is "The keyword only, never the line ... a CLI
    transcript can carry passwords". CLAUDE.md: "Secrets ride stdin, never argv, and are never echoed
    back. Reports say set/MISSING."
    Failure: an LDAP profile block absent after import prints its aes-password line, and a chunk
    rejected next to a password line echoes it in the tail. Either way it lands in CI and dev-script
    logs.

Why it matters: This is security-adjacent (filed under correctness as the closest area). The material is the
same material that makes the artifact 0600-sensitive. Confirm by checking which real capture lines carry
encrypted credentials.

Change: Add one masking helper in internal/broker that replaces the operand of credential-bearing lines
(password, aes-password, pre-shared-key, and their encrypted forms) with `***`. Apply it to FailedDetail
before ImportResult.Report prints it, and to bd.Missing lines in reportBlockLines, keeping the masked line
so the report stays actionable. Test both paths.

To confirm: Check which lines of a real capture carry encrypted credentials and whether the failure detail
and RenderDiffResult print them.

Area: correctness

#### CR-0086 [minor] [new] Repeated --vpn duplicates every block of that VPN in the artifact

Location: internal/cli/flags.go:109-133, internal/broker/importops.go:133-151,
internal/broker/importops.go:233-249
Confidence: suspected
Evidence:

    flags.go:109 registers `--vpn` with StringArrayVar, and exportScope (119-133) validates each value
    but does not dedupe.
    importops.go:234-247 captures once per list entry and concatenates: `combined.Blocks =
    append(combined.Blocks, c.Blocks...)`.
    Failure: `export-config --vpn "$A" --vpn "$B"` with A == B writes an artifact with every block of
    that VPN twice. On import, the create-VPN chunk and the VPN chunk replay each `create ...` twice.
    Unless the broker wraps the line in ignore-already-exists, the second create is rejected after the
    VPN was torn down.

Why it matters: The artifact exports cleanly and fails only at import, after the destructive phase. Confirm
with a doubled export against a broker.

Change: In exportScope (flags.go:127-132), drop repeated names while keeping first-seen order (or refuse a
repeat as a usage error naming it), and do the same in ExportConfig before captureScope. Add a test with a
doubled --vpn.

To confirm: Run an export with the same --vpn given twice against a broker and inspect the artifact for
duplicated blocks.

Area: correctness

#### CR-0087 [minor] [new] Unrecognised show replication parses as an unconfigured mate and triggers a site-wide replication stop

Location: internal/broker/replication.go:339-392, internal/broker/replication.go:449-461,
internal/broker/replicationops.go:38-43, internal/broker/replicationops.go:121-139,
internal/broker/replicationops.go:197-210
Confidence: suspected
Evidence:

    Software parser: with no `Replication Mate:` line anywhere the loop finishes and returns `m, nil`
    (replication.go:391) -- empty VRN, no endpoints. Appliance parser: host=="" and VRN=="" ->
    "Genuinely unconfigured: no mate, no address. return m, nil" (459-460). Contrast ParseVPNReplication
    (720-723), which errors because "an empty table means the report was not understood".
    ConfigureReplication then: 136 RenderMateRemovals(typ, have, want) clears nothing (have is empty);
    197 `mateApplied := have.VirtualRouterName != want.VirtualRouterName || ...` -> true; 206-207 shuts
    replication down on every admin-enabled VPN and re-sends the mate lines. replicationops.go:41-43:
    "that check is the only thing standing between an ordinary re-run and a site-wide replication stop."
    Failure: a broker release whose `show replication` heading differs -> every `configure dr` re-run
    stops replication on all VPNs; and because no removals are rendered the mate lines are additive, so
    on software already holding 2 ssl endpoints while the file wants 2 different ones the adds can
    exceed the per-transport cap, be rejected, and leave every VPN with replication disabled
    (replPhase1Rejected).

Why it matters: Parse failure is treated as a real state that authorises the most disruptive branch.
Suspected because it needs a real report without the label; confirm that an unconfigured broker still prints
`Replication Mate:` and then require it.

Change: In parseShowReplicationSoftware and parseShowReplicationAppliance record whether mateLabel was seen
and return an error ("`show replication` has no Replication Mate: line; the report was not understood") when
it was not, so ConfigureReplication fails before phase 1 instead of reading an unknown report as an empty
mate.

To confirm: Check that an unconfigured broker's `show replication` still prints the `Replication Mate:`
label.

Area: resilience

#### CR-0088 [minor] [new] First podman deploy pulls the broker image inside systemd's default start timeout

Location: internal/render/render.go:999-1014, internal/container/manager.go:837-848,
internal/container/manager.go:451-478
Confidence: suspected
Evidence:

    render.go 1000-1010: [Service] carries LimitNOFILE, LimitMEMLOCK, LimitCORE and Restart=always only
    -- no TimeoutStartSec.
    manager.go 844  return m.systemctl(ctx, systemd.Start(user, svc))  // waits for the service to come
    up
    manager.go 456-478: prep logs in to the registry but nothing pulls the image ahead of the start.
    Step by step: fresh host, image not cached -> quadlet's ExecStart `podman run` pulls the 1 GB+
    broker image while systemd counts the default 90 s start timeout -> on a slow registry or proxy the
    start job times out, `broker deploy` fails, and Restart=always retries in the background.

Why it matters: Suspected: depends on quadlet not raising TimeoutStartSec for the generated service and on
link speed. Confirm with `systemctl --user show solace.service -p TimeoutStartUSec` on a deployed host and a
deploy with the image removed over a throttled link.

Change: Pull explicitly in Deploy before the unit is started (`<runtime> pull <image.Ref()>` through m.run,
after registryLogin, skipped under Echo), or emit `TimeoutStartSec=900` in [Service] and pin it in the
quadlet goldens and TestQuadletKeysAreWithinTheFloor.

To confirm: systemctl --user show <name>.service -p TimeoutStartUSec on a deployed host, then a first deploy
with the image absent on a slow link.

Area: resilience

#### CR-0089 [minor] [new] Rootless podman on a cgroup v1 host: memory/cpuset caps are not checked for, and the printed remedy cannot fix it

Location: internal/container/limits.go:44-51, internal/container/limits.go:226-262,
internal/render/render.go:932-954, internal/container/rootless.go:85-119
Confidence: suspected
Evidence:

    render.go 949-954: rootless units always carry `PodmanArgs=--cpuset-cpus=` and
    `PodmanArgs=--memory=`.
    limits.go 228-230: the only related check is `missingControllers(props["DelegateControllers"])`,
    whose remedy (243-257) is a `Delegate=` drop-in.
    rootless.go 85-119: no row reads the cgroup version; a grep of internal/ and docs/ for cgroup v1/v2
    or CgroupsVersion finds nothing.
    Step by step: RHEL 8 (cgroup v1 by default) with the rootless default -> either the
    DelegateControllers row refuses with a Delegate= remedy that cannot delegate on v1, or (once the
    drop-in exists and systemd reports the list) the deploy proceeds and podman drops the resource
    limits for a rootless v1 container, so the broker runs without the tier's memory cap and cpuset.

Why it matters: Suspected: podman's v1-rootless behaviour (warn and ignore limits) is runtime state. Confirm
on a v1 host: `podman info --format '{{.Host.CgroupsVersion}}'` -> v1, deploy, then `podman inspect -f
'{{.HostConfig.Memory}} {{.HostConfig.CpusetCpus}}' solace`.

Change: Add a rootlessRows check of `podman info --format {{.Host.CgroupsVersion}}` (through the guarded
runtime) that refuses v1 with the actual remedy (boot with systemd.unified_cgroup_hierarchy=1, or
podman.rootful: true), and document cgroup v2 as a rootless prerequisite in operations.md's Rootless podman
prerequisites.

To confirm: On a cgroup v1 host: podman info --format '{{.Host.CgroupsVersion}}', deploy, then podman
inspect -f '{{.HostConfig.Memory}} {{.HostConfig.CpusetCpus}}' <name>.

Area: build

#### CR-0090 [minor] [new] The release target set is defined twice and nothing ties the two together; developer.md calls the local dist list the release targets

Location: scripts/dev.sh:31-33, scripts/dev.ps1:49-56, docs/developer.md:31-34,
.github/workflows/tag.yml:200-203, README.md:72-75
Confidence: suspected
Evidence:

    dev.sh:31-32 says 'Local convenience cross-compile set (`dist`). CI does not use it: tag.yml
    matrixes over `build` with TARGET_OS/TARGET_ARCH from BUILD_TARGETS instead.'
    docs/developer.md:32-33 says '`dist` cross-compiles all four release targets -- linux/amd64,
    linux/arm64, darwin/arm64, windows/amd64'. tag.yml:203 `target: ${{ fromJSON(vars.BUILD_TARGETS)
    }}`. README.md:74 'Download the release binary for your platform.' The set actually shipped lives
    only in a repo variable that no file records.

Why it matters: If BUILD_TARGETS differs from the four documented targets, the docs promise binaries the
release does not carry, or the other way round. To confirm, compare `gh variable get BUILD_TARGETS` with
DIST_TARGETS.

Change: State in docs/developer.md that BUILD_TARGETS is the release set and give its current value. Call
DIST_TARGETS a local convenience that should match it, not 'the release targets'. Optionally have the plan
job print the resolved targets in its ::notice line so every run records them.

To confirm: gh variable get BUILD_TARGETS compared with DIST_TARGETS in scripts/dev.sh:31-33.

## Dynamic checks

None run.

Read, not run (scripts/logs/, written 2026-11-11T16:17 +0800 against the code at this commit):

- build.log: OK. vet.log: OK. test.log: every package `ok` (the run is not -v, so skipped tests do not show).
- cov.log: total 97.7% of statements. scan.log: no vulnerabilities called by this module.

Offered for the user to run (nothing was installed or modified):

- `go test -count=1 -v ./internal/broker 2>&1 | grep -c -- '--- SKIP'` - settles the skip count behind CR-0017.
- `"$(go env GOROOT)/bin/gofmt" -l .` - sizes the formatting gap in CR-0052 before a gate is added.
- `GOWORK=off ./scripts/dev.sh all scan` - runs the gates against the pinned v0.1.2, as CI does (CR-0051).
- `gh variable get BUILD_TARGETS` - confirms or clears CR-0090.
- `kubectl --context X config current-context` with a kubeconfig whose current-context is Y - confirms or clears CR-0082.
- `kubectl get sts <name>-pubsubplus-p -n <ns> -o jsonpath='{.spec.updateStrategy.type}'` - confirms or clears CR-0083.
- shellcheck, actionlint, staticcheck: not installed on this host; not run.

## Fixed since last report

First run

## Accepted (not re-litigated)

None

## Coverage

- Areas reviewed: design, correctness, resilience, quality, tests, cicd, build, interfaces, observability,
  conventions; skipped: none
- Read closely: main.go; internal/cli/*.go; internal/config/*.go; internal/convert/*.go;
  internal/broker/*.go (scripts.go read at 186-425, the rest by grep); internal/k8s/*.go (checkreport.go at
  195-244 and 400-680); internal/container/*.go; internal/render/render.go; internal/engine/runner.go;
  internal/output/output.go; internal/examples/examples.go; internal/tools/vulnjudge/main.go;
  .github/workflows/ci.yml, tag.yml and dependabot.yml; scripts/dev.sh; scripts/dev.ps1; gen-cert.sh;
  go.mod; README.md; docs/developer.md; docs/container-security.md
- Sampled: the _test.go files (about 25 of 90 read in part, plus a scripted two-way check of all 1670 Test
  functions against docs/test.md); docs/operations.md and docs/configuration.md (the sections for each path
  reviewed); docs/test.md; docs/commands.md (by grep); the goldens (two container goldens in full, the rest
  by posture grep); internal/k8s/assets/operator-1.4.2.yaml.tmpl (kind, name and substitution lines);
  solace-yaml-generator.html (the hand-written JS for kubernetes.ports and joinPEM only)
- Not read: docs/parity.md, docs/abbreviation.md, docs/import.md beyond the cited rows, env/sample.yaml,
  graphify-out/, the rest of the generator page, and the convert, scaling, watch and exportconfig test files
- Stacks with core checklist only: PowerShell (scripts/dev.ps1), HTML/JS (solace-yaml-generator.html)
- Dynamic checks not run: all of them (see Dynamic checks); shellcheck, actionlint and staticcheck are not
  installed
- Limits:
  - The shared module github.com/hafio/go-deploy-utility was out of scope (a parallel session owns it). It
    was read only at the pinned v0.1.2 in the module cache, as context for how this app calls it. Defects
    seen inside it are the hand-off notes below; they have no CR ID and are not in the ledger.
  - The review ran as nine area reviewers (Claude Opus 5.5 subagents) with the lead merging and
    deduplicating. The lead opened the cited lines for all 4 blockers, all 13 majors and the 3 suspected
    majors. Minors and nits rest on the reviewers' reads; 8 of the 63 were spot-checked, and every cited
    path and line range was checked to exist.
  - One reviewer cited .github/workflows/tag.yml with line numbers offset by 82 (it read both workflow files
    as one stream). Those citations were corrected and re-checked against the file.
  - Recalibrated by the lead: CR-0055, CR-0026, CR-0029, CR-0042 were reported as major and filed as minor
    (deliberate and documented, or reachable only under unusual conditions); CR-0081 was reported as
    verified and filed as suspected (it depends on the shape of a real capture that is not in the tree).
  - No live broker, cluster or podman host was used. Findings that turn on broker behaviour rely on the
    repo's own statements about it and say so.
  - Not visible from here: branch protection and rulesets, and the BUILD_TARGETS and SHIP_* repository
    variables.
  - The graph was reused, not refreshed: it was built from 20ec2f70, and b0941b1 adds only docs and graph
    outputs.
  - Generator page: by reading, the hand-written JS and the Go renderers agree for input the form can
    produce. The manual browser cross-checks docs/developer.md asks for are still owed for kubernetes.ports
    (2026-10-02) and joinPEM (2026-10-23).
  - This report follows the review template and carries no support notice. The notice test globs docs/*.md
    without recursing, so docs/review/ is outside it; docs/review/ is also not in .graphifyignore, so the
    next graph refresh will ingest this report (CR-0080).
  - Hand-off notes for go-deploy-utility (paths are relative to the module at v0.1.2):
    - kube/kube.go:250-256 and containercli/containercli.go:295-300: Tail is documented as 'passed as given,
      so the tool validates them first', but the engines disagree on its grammar (docker accepts `all`,
      kubectl and podman take an integer). Every consumer has to re-derive the same rule; a shared Tail
      validator or per-engine normalisation would remove the trap this app fell into.
    - kube/kube.go:203-224 (Exec, and Upload which builds on it): the pod name is emitted as a bare
      positional before `--` with no refusal of a leading '-', so a dash-led pod is parsed by kubectl as a
      flag. A shared guard (refuse a leading '-', or a DNS-1123 check) would back up every caller's own
      validation.
    - kube/kube.go:203-224 -- Exec refuses only an empty pod and appends it as a bare positional ahead of
      nsArgs, so a pod name beginning with '-' is parsed by kubectl as a flag; this app feeds it
      podNameFor(site.Via.Kubernetes.Name, ...) where via.kubernetes.name is only checked non-empty
      (internal/config/validate.go:1181-1184), so a dash-led refusal in Exec (and Upload) would close it for
      every caller.
    - kube/nsgate/nsgate.go:52-56 -- List reads only all,persistentvolumeclaims,secrets,configmaps. Roles,
      RoleBindings, ServiceAccounts, NetworkPolicies, Ingresses, PodDisruptionBudgets, Leases and any CR
      outside the `all` category (including PubSubPlusEventBroker) are never listed, so Occupants can report
      a namespace holding them as empty, and both tools then offer it for a cascading delete.
    - kube/kube.go:104-112 -- ParseNamespace maps empty stdout to NamespaceAbsent. That is right for a real
      `--ignore-not-found` read, but a silent wrapper runner also reads as "absent" rather than "unknown",
      unlike nsgate.Occupants, which refuses an empty listing.
    - containercli/containercli.go:130-155 (ComposeDigest): the digest is an unsalted sha256 over the
      compose file plus every secret value, and consumers publish it as a container label readable with
      `docker inspect`; for a deployment whose only secret is a low-entropy admin password this is an
      offline-guessable hash (low impact since docker access is root-equivalent, but container-security.md
      203-205 promises nothing derived from a value lands on a label).
    - systemd/systemd.go:64-73 (Stopped): accepting 'inactive'/'unknown' as 'confirmed stopped' is also what
      a unit that never existed reports, so a teardown keyed on a renamed unit passes the guard; worth
      documenting on the helper so consumers do not treat it as proof that no container is using their data.
    - watch/watch.go:141-151 -- loop retries every refresh error forever with no way for a caller to mark
      one fatal (e.g. an exitcode usage error), so a deterministic mistake inside a refresh becomes an
      endless warning and exit 0 on Ctrl-C; consider stopping on exitcode.IsUsage or a Fatal marker.
    - confirm/terminal.go -- StdinIsTerminal treats any character device (including /dev/null) as a human;
      documented by the app, but the shared package could offer a real isatty check (term.IsTerminal) as an
      option for consumers that want /dev/null to count as unattended.
    - runner/runner.go:236-247 -- classify wraps a failed child as "<argv[0]>: exit status N". It names
      neither the subcommand nor what to do next, so every caller that returns the error bare (most of
      internal/k8s and internal/container here) ends with a final line like "Error: systemctl: exit status
      1". Consider including the first non-flag argument, or the full argv when -v is set.
    - rootless/rootless.go:46-59 -- not a defect: noted only because this app's executor-side test
      (internal/container/manager_test.go:2291-2300) relies on the message containing the key name. Keep the
      key in both refusal messages.
    - notice/noticetest/noticetest.go:28-29,112-113 (v0.1.2): the default Files set globs docs/*.md without
      recursing, so a consumer's doc in a docs/ subdirectory (e.g. docs/review/) escapes the 'every doc
      carries the notice' rule. Consider a recursive default or an explicit exclusion list.
    - docs/consuming.md (v0.1.2): the go.work toolchain-trap paragraph and the 'run gates with GOWORK=off
      once the tag is published' guidance exist only at unreleased HEAD (83fa381). Consumers pinned at
      v0.1.2 do not get them from the module cache, so the next tag should carry them.

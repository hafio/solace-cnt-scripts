// Package k8s implements the Kubernetes platform of the solace-util CLI: the operator
// bundle, cluster/broker lifecycle over kubectl, and the concrete broker.Transport
// that wraps `kubectl exec/cp`. It is the Go port of the numbered bash scripts at
// the repo root (001-069, 010-020 deploy, 110-120 delete, and the operational
// helpers). Cluster/operator operations run over an engine.Runner directly; runtime
// broker operations run through internal/broker over the transport in transport.go.
package k8s

import "solace/internal/config"

// brokerSuffix is the operator's fixed name suffix for every broker resource
// (StatefulSets, pods, PVCs, the LB service): <name>-pubsubplus[...].
const brokerSuffix = "-pubsubplus"

// podNameFor returns the broker pod for a role given the DEPLOYMENT NAME rather than a
// whole config: <name>-pubsubplus-<p|b|m>-0. The operator names pods off the per-role
// StatefulSet's single replica (050:30, enter-solace-cli.sh:18). Broker pods are
// single-container, so no `-c` is ever used.
//
// It takes a bare name because a replication MATE is a different deployment in a
// different cluster: it is described by its site's own `via.kubernetes` block, not by a
// *config.Config, so matechannel.go could not call podName and spelled the rule out by
// hand instead. Two spellings of one naming rule is how the rule drifts, and this is the
// file that exists so it lives once.
func podNameFor(name string, role config.Role) string {
	return name + brokerSuffix + "-" + role.Letter() + "-0"
}

// podName is podNameFor for THIS deployment, which is what almost every caller wants.
func podName(cfg *config.Config, role config.Role) string {
	return podNameFor(cfg.K8s.Name, role)
}

// pvcName returns the PersistentVolumeClaim backing a role's pod:
// data-<name>-pubsubplus-<role>-0 (120:65-69).
func pvcName(cfg *config.Config, role config.Role) string {
	return "data-" + cfg.K8s.Name + brokerSuffix + "-" + role.Letter() + "-0"
}

// stsName returns the StatefulSet for a role: <name>-pubsubplus-<role> (no `-0`
// suffix, unlike the pod/PVC) (replicas-start-broker.sh:17,24,29).
func stsName(cfg *config.Config, role config.Role) string {
	return cfg.K8s.Name + brokerSuffix + "-" + role.Letter()
}

// lbServiceName returns the load-balancer service: <name>-pubsubplus, with no role
// suffix -- one service fronts the active broker (desc-lb.sh:16).
func lbServiceName(cfg *config.Config) string {
	return cfg.K8s.Name + brokerSuffix
}

// HARoles returns the broker roles present in this deployment: [primary] for a
// standalone broker, [primary, backup, monitor] for a redundancy group. It bounds
// the per-role resource operations (PVC deletion, replica scaling, diagnostics).
func HARoles(cfg *config.Config) []config.Role {
	if cfg.RedundancyEnabled() {
		return []config.Role{config.Primary, config.Backup, config.Monitor}
	}
	return []config.Role{config.Primary}
}

// RestartOrder returns the roles in the order a manual pod bounce should follow:
// monitor first, then backup, then primary -- least message-routing impact first,
// with the node most likely to be serving traffic left for last. Standalone has
// only the one broker.
//
// The order is by configured role, not by which node is currently active: after a
// failover the config's "primary" may be the standby. Check `verify redundancy`
// first, or restart roles one at a time in the order you want.
func RestartOrder(cfg *config.Config) []config.Role {
	if cfg.RedundancyEnabled() {
		return []config.Role{config.Monitor, config.Backup, config.Primary}
	}
	return []config.Role{config.Primary}
}

// ProductKeyRoles returns the roles a product key is applied to: [primary] for a
// standalone broker, [primary, backup] for a redundancy group -- never the monitor,
// which carries no message spool (057).
func ProductKeyRoles(cfg *config.Config) []config.Role {
	if cfg.RedundancyEnabled() {
		return []config.Role{config.Primary, config.Backup}
	}
	return []config.Role{config.Primary}
}

// roleName is the long-form role word used in this package's own progress lines --
// "copying x from backup" reads where a bare letter would not. It lives beside the
// other role helpers rather than in the file that happens to call it; it moved here
// when the node-labelling prompt that first needed it was removed.
//
// It is a hand-rolled duplicate of config.Role.Word() and should be deleted in favour
// of it, but its remaining call sites (ops.go) are outside this change's file scope,
// so it stays until they move.
func roleName(role config.Role) string {
	switch role {
	case config.Backup:
		return "backup"
	case config.Monitor:
		return "monitor"
	default:
		return "primary"
	}
}

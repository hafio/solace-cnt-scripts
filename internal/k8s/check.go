package k8s

import (
	"context"
	"fmt"
	"strings"

	"solace/internal/engine"
)

// reportKeyWidth is the key column for the few standalone report lines that are
// not part of the grouped `check deploy` report (checkreport.go owns that, with
// its own width): Preflight when called on its own rather than through the report.
const reportKeyWidth = 15

// defaultSCJSONPath selects the name of the StorageClass annotated as cluster
// default, porting 009:18. The dots in the annotation key are backslash-escaped as
// jsonpath requires; the whole value is one kubectl argv token (no shell involved).
const defaultSCJSONPath = `jsonpath={.items[?(@.metadata.annotations.storageclass\.kubernetes\.io/is-default-class=="true")].metadata.name}`

// isEcho reports whether commands are being echoed rather than run (the Echo
// runner). The cluster-touching checks use this to skip assertions they cannot make
// without a live cluster, while still echoing the commands they would run. A real
// cluster never hits this path, so `no default StorageClass` still fails loud there.
func (c *Cluster) isEcho() bool {
	_, ok := c.R.(engine.Echo)
	return ok
}

// Check is the whole-env-file validation, and is the container-side Manager.Check's
// counterpart -- kept as the one name the CLI calls on both platforms, so the top-level
// `validate` dispatches to a single method name whichever platform it lands on. The report
// itself lives in checkreport.go, which also owns the two scoped views.
func (c *Cluster) Check(ctx context.Context) error { return c.Validate(ctx) }

// Reachable probes the API server (001's kubectl availability check, strengthened
// to an actual server round-trip). The Echo runner returns no error, so a preview
// passes this and shows the command it would have run.
func (c *Cluster) Reachable(ctx context.Context) error {
	if _, err := c.output(ctx, "version", "-o", "json"); err != nil {
		return fmt.Errorf("cannot reach the Kubernetes API server (check kubeconfig/context): %w", err)
	}
	return nil
}

// storageClassAttrs reads the two attributes that decide whether a class can back
// a broker. Split out from storageRows (checkreport.go) so the two kubectl reads
// and the rule they feed (storageClassSuitable) live in one place.
func (c *Cluster) storageClassAttrs(ctx context.Context, name string) (binding, expansion string, err error) {
	if binding, err = c.scColumn(ctx, name, ".volumeBindingMode"); err != nil {
		return "", "", fmt.Errorf("reading StorageClass %q: %w", name, err)
	}
	if expansion, err = c.scColumn(ctx, name, ".allowVolumeExpansion"); err != nil {
		return "", "", fmt.Errorf("reading StorageClass %q: %w", name, err)
	}
	return binding, expansion, nil
}

// storageClassSuitable is the rule itself: the PV has to land in the pod's zone
// (WaitForFirstConsumer) and the volume has to be growable (009:26-41).
func storageClassSuitable(name, binding, expansion string) error {
	if binding != "WaitForFirstConsumer" || expansion != "true" {
		return fmt.Errorf("StorageClass %q is unsuitable: need volumeBindingMode=WaitForFirstConsumer (got %q) and allowVolumeExpansion=true (got %q)",
			name, binding, expansion)
	}
	return nil
}

// resolveStorageClass returns the configured class if set, else the cluster default
// resolved by annotation (009:18). Multiple defaults are an error rather than a
// silent pick, since the choice would be non-deterministic.
func (c *Cluster) resolveStorageClass(ctx context.Context) (string, error) {
	if c.Cfg.K8s.Storage.Class != "" {
		return c.Cfg.K8s.Storage.Class, nil
	}
	out, err := c.output(ctx, "get", "sc", "-o", defaultSCJSONPath)
	if err != nil {
		return "", fmt.Errorf("resolving default StorageClass: %w", err)
	}
	fields := strings.Fields(string(out))
	if len(fields) > 1 {
		return "", fmt.Errorf("multiple default StorageClasses found (%s); set kubernetes.storage.class explicitly", strings.Join(fields, ", "))
	}
	if len(fields) == 1 {
		return fields[0], nil
	}
	return "", nil
}

// scColumn reads one field of a StorageClass via custom-columns. kubectl prints
// "<none>" (not empty) for an absent field, which is preserved so the caller can
// report it rather than mistaking it for a missing class.
func (c *Cluster) scColumn(ctx context.Context, name, field string) (string, error) {
	out, err := c.output(ctx, "get", "sc", name, "-o", "custom-columns=V:"+field, "--no-headers")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// --- small report formatters ------------------------------------------------

func orNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

func orValue(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func setOrMissing(s string) string {
	if s == "" {
		return "MISSING"
	}
	return "set"
}

func setOrNone(present bool) string {
	if present {
		return "set"
	}
	return "(none)"
}

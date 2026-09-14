package k8s

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// The operator is cluster-scoped and shared by every broker in the cluster, so
// installing an OLDER one than the cluster already runs is not a local mistake --
// it changes reconciliation behaviour for brokers this env file knows nothing
// about, and it is silent: `kubectl apply` downgrades an image without comment.
//
// So `deploy operator` reads what is already running, shows both versions, and
// ASKS. A deliberate rollback is a legitimate thing to want, which is why this is
// a question and not a refusal; doing one by accident is not, which is why an
// unattended run declines by default.

// imageTag splits a container image reference into its repository and tag. A
// digest pin (@sha256:...) or a tagless reference yields an empty tag, which the
// comparison treats as "not comparable" rather than guessing.
func imageTag(ref string) (repo, tag string) {
	if at := strings.Index(ref, "@"); at >= 0 {
		return ref[:at], "" // digest-pinned: no version to compare
	}
	slash := strings.LastIndex(ref, "/")
	colon := strings.LastIndex(ref, ":")
	if colon < 0 || colon < slash {
		return ref, "" // no tag, or the colon belonged to a registry:port
	}
	return ref[:colon], ref[colon+1:]
}

// compareVersions orders two dotted numeric versions: -1 if a < b, 0 if equal,
// +1 if a > b. ok is false when either side is not a dotted-numeric version --
// `latest`, a date stamp, a git sha -- because ordering those is guesswork and a
// guess here would produce a downgrade warning nobody can act on.
//
// A leading `v` is tolerated on either side; trailing pre-release text is not
// parsed, so 1.4.0-rc1 compares as 1.4.0. That is deliberate: the question this
// answers is "is the cluster ahead of us", and a release candidate of the same
// version is not the case worth shouting about.
func compareVersions(a, b string) (int, bool) {
	pa, oka := parseVersion(a)
	pb, okb := parseVersion(b)
	if !oka || !okb {
		return 0, false
	}
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y int
		if i < len(pa) {
			x = pa[i]
		}
		if i < len(pb) {
			y = pb[i]
		}
		if x != y {
			if x < y {
				return -1, true
			}
			return 1, true
		}
	}
	return 0, true
}

// parseVersion turns "1.4.2" or "v1.4.2-rc1" into [1 4 2].
func parseVersion(s string) ([]int, bool) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	if s == "" {
		return nil, false
	}
	// Drop any pre-release/build tail so 1.4.0-rc1 parses as 1.4.0.
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, false
		}
		out = append(out, n)
	}
	return out, true
}

// operatorVersionWarning compares the operator already running in the cluster
// against the one this config would apply, and returns the two comparison lines
// when the existing one is NEWER. It returns nil when there is nothing to say:
// no operator installed yet, the same version, an upgrade, or a pair of tags that
// cannot be ordered.
//
// The body is deliberately just the two versions. The question that follows is
// what carries the weight -- explaining at length what a downgrade means, to
// somebody who is about to be asked whether they want one, is text they will
// scroll past to reach the prompt.
func operatorVersionWarning(existingImage, newImage string) []string {
	_, existingTag := imageTag(existingImage)
	_, newTag := imageTag(newImage)
	if existingTag == "" || newTag == "" {
		return nil
	}
	cmp, ok := compareVersions(existingTag, newTag)
	if !ok || cmp <= 0 {
		return nil
	}
	return []string{
		"existing : " + existingTag,
		"new      : " + newTag,
	}
}

// installedOperatorImage reads the image the operator deployment is actually
// running, or "" when there is no operator to read (the normal first-install
// case, not an error).
//
// It reads across ALL namespaces rather than naming one, for two reasons. The
// first is noise: on a first install the operator namespace does not exist yet,
// and a namespaced get prints `Error from server (NotFound): namespaces
// "solace-operator" not found` straight to stderr -- an alarming line about a
// lookup whose failure this function is designed to shrug off. The second is
// correctness: the operator may be running somewhere other than where this env
// file expects, and that is precisely the case worth noticing before a deploy.
func (c *Cluster) installedOperatorImage(ctx context.Context) string {
	dep, err := c.findOperatorDeployment(ctx)
	if err != nil {
		return ""
	}
	return imageFromDeployment(dep)
}

// imageFromDeployment reads the running image off an already-fetched operator
// Deployment, so a caller that already has one does not fetch it again. A nil dep, or
// one whose containers name no image, is "" -- the same "nothing to read" this function
// has always returned rather than an error.
func imageFromDeployment(dep *deploymentItem) string {
	if dep == nil {
		return ""
	}
	for _, ct := range dep.Spec.Template.Spec.Containers {
		if ct.Image != "" {
			return ct.Image
		}
	}
	return ""
}

// confirmNoDowngrade is the gate OperatorApply passes through before it applies
// the bundle. It shows the two versions and asks; declining aborts the apply.
//
// It asks rather than merely warning because the operator is cluster-scoped: the
// blast radius is every broker in the cluster, including ones this env file has
// never heard of, and `kubectl apply` would roll the image back without comment.
// An unattended run (nil Confirm, or --no-prompt wired to decline) therefore does
// NOT downgrade by default -- the same direction every other destructive question
// in this tool takes.
//
// It returns nil when there is nothing to ask about: a preview, a first install,
// an upgrade, the same version, or tags that cannot be ordered.
func (c *Cluster) confirmNoDowngrade(ctx context.Context, opNS string) error {
	if c.isEcho() {
		return nil // a preview has no cluster to read
	}
	body := operatorVersionWarning(c.installedOperatorImage(ctx), operatorImage(c.Cfg))
	if body == nil {
		return nil
	}
	c.progress().Warning("operator downgrade", body...)
	if c.confirm("Downgrade the operator?") {
		return nil
	}
	return fmt.Errorf("operator downgrade declined; nothing was applied. The operator already installed is " +
		"newer than kubernetes.operator.image -- point that at the installed version, or skip `deploy operator`")
}

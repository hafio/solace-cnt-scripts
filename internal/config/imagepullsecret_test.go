package config

import (
	"strings"
	"testing"
)

// The image-pull Secret can come from either side, exactly like the TLS Secret
// (tlssecret_test.go): this tool builds it from image.user/image.pass, or the
// operator -- or a cluster admin, or external-secrets -- created it themselves
// and kubernetes.imagePullSecret only names which one the broker CR should
// reference. These pin the three states config.go's own doc comment spells
// out, and the middle one -- a name with no credentials behind it -- is the
// one a future reader would otherwise "fix" into a validation error: it
// deliberately stays legal, mirroring the rule ManagesTLSSecret already states
// for kubernetes.tlsServerSecret.

// pullSecretCfg is a valid Kubernetes config with the pull-secret fields under
// the caller's control.
func pullSecretCfg(name, user, pass string) *Config {
	c := validK8sConfig()
	c.K8s.ImagePullSecret = name
	c.Image.User = user
	c.Image.Pass = pass
	return c
}

// TestManagesImagePullSecretNeedsBothCredentials covers all four corners of
// user x pass: only the pair together makes the Secret this tool's to build --
// the same AND that ManagesTLSSecret uses for tls.cert/tls.certKey. A lone
// username with no password (or the reverse) cannot build a dockerconfigjson
// Secret, so it must read as "bring your own", never as "build with half a
// credential".
func TestManagesImagePullSecretNeedsBothCredentials(t *testing.T) {
	cases := []struct {
		name       string
		user, pass string
		want       bool
	}{
		{"both set", "reguser", "regpass", true},
		{"user only", "reguser", "", false},
		{"pass only", "", "regpass", false},
		{"neither", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := pullSecretCfg("", tc.user, tc.pass)
			if got := c.ManagesImagePullSecret(); got != tc.want {
				t.Errorf("ManagesImagePullSecret(user=%q, pass=%q) = %v, want %v", tc.user, tc.pass, got, tc.want)
			}
		})
	}
}

// TestImagePullSecretNameConfiguredNameWins: an operator-chosen name is never
// overridden by the derived default, whether or not this tool goes on to
// build the Secret behind it.
func TestImagePullSecretNameConfiguredNameWins(t *testing.T) {
	c := pullSecretCfg("custom-pull-secret", "reguser", "regpass")
	if got := c.ImagePullSecretName(); got != "custom-pull-secret" {
		t.Errorf("ImagePullSecretName() = %q, want the configured name", got)
	}
}

// TestImagePullSecretNameDerivedWhenCredentialsButNoName is the common case:
// image.user/image.pass with no kubernetes.imagePullSecret builds a Secret
// under a name derived from kubernetes.name, the same "<name>-<suffix>" shape
// AdditionalUsersSecretName already uses for the extra-users Secret.
func TestImagePullSecretNameDerivedWhenCredentialsButNoName(t *testing.T) {
	c := pullSecretCfg("", "reguser", "regpass")
	c.K8s.Name = "mybroker"
	want := "mybroker-image-pull"
	if got := c.ImagePullSecretName(); got != want {
		t.Errorf("ImagePullSecretName() = %q, want %q", got, want)
	}
}

// TestImagePullSecretNameBringYourOwn is the subtle middle state: a name with
// NO credentials behind it. That is not an invalid config to reject -- it is
// exactly the case ManagesTLSSecret's own doc comment describes for
// kubernetes.tlsServerSecret, a Secret created by hand or by a cluster admin
// (or external-secrets) that the CR should merely reference by name. The name
// must still come back, so render.go's pullSecrets block still points the CR
// at it; ManagesImagePullSecret must come back false, so nothing here tries to
// build, apply, or delete a Secret whose contents it was never given.
func TestImagePullSecretNameBringYourOwn(t *testing.T) {
	c := pullSecretCfg("operator-created-secret", "", "")
	if c.ManagesImagePullSecret() {
		t.Error("a name with no credentials must not read as this tool's Secret to build")
	}
	if got := c.ImagePullSecretName(); got != "operator-created-secret" {
		t.Errorf("ImagePullSecretName() = %q, want the configured name even with no credentials behind it", got)
	}
	if err := c.Validate(K8s); err != nil {
		t.Errorf("a pull-secret name with no credentials is a legal bring-your-own deployment: %v", err)
	}
}

// TestImagePullSecretNameEmptyWithNeither is the third state: no configured
// name and no credentials to derive one from. Returning "" here -- rather than
// inventing a name for a Secret that will never exist -- is what lets every
// reference site (render.go's CR block, the operator bundle's own
// imagePullSecrets, the "image pull" check-report row) stay a plain
// `if n := c.ImagePullSecretName(); n != ""` instead of also asking
// ManagesImagePullSecret.
func TestImagePullSecretNameEmptyWithNeither(t *testing.T) {
	c := pullSecretCfg("", "", "")
	if got := c.ImagePullSecretName(); got != "" {
		t.Errorf("ImagePullSecretName() = %q, want empty with no name and no credentials", got)
	}
	if err := c.Validate(K8s); err != nil {
		t.Errorf("no pull secret at all is the plainest deployment and must validate: %v", err)
	}
}

// TestImagePullSecretDerivedSuffixFitsWhereAdditionalUsersAlreadyFits pins the
// "no new length bound needed" reasoning as a RELATIONSHIP rather than a
// re-typed constant: kubernetes.name is already bounded, by the pod-name
// derivation TestKubernetesNameIsBoundedByItsDerivedNames covers, against the
// "-additional-users" suffix AdditionalUsersSecretName has appended for as
// long as that method has existed. "-image-pull" only has to be no longer
// than that suffix for every kubernetes.name that already fits one to fit the
// other too -- which is what makes a dedicated bound for the new suffix
// unnecessary rather than merely untested.
func TestImagePullSecretDerivedSuffixFitsWhereAdditionalUsersAlreadyFits(t *testing.T) {
	c := pullSecretCfg("", "reguser", "regpass")
	c.K8s.Name = "n" // any name; only the SUFFIX length is under test
	pullSuffix := strings.TrimPrefix(c.ImagePullSecretName(), c.K8s.Name)
	usersSuffix := strings.TrimPrefix(c.AdditionalUsersSecretName(), c.K8s.Name)
	if len(pullSuffix) >= len(usersSuffix) {
		t.Errorf("derived image-pull suffix %q (%d chars) must be shorter than the additional-users suffix %q "+
			"(%d chars): a longer suffix would need its own length bound, which this change deliberately does not add",
			pullSuffix, len(pullSuffix), usersSuffix, len(usersSuffix))
	}
}

// TestImagePullSecretCredentialsResolveFromPassEnv pins "or equivalent" at the
// level the owner asked for it: through a full Load, not a hand-built Config.
// image.passEnv is resolved into Image.Pass by resolveSecretRefs
// (secretref.go) during Load, so an env file supplying the registry password
// only as an env-var reference must still make ManagesImagePullSecret true and
// derive the same default name a literal image.pass would -- proving
// ManagesImagePullSecret needs no separate image.passEnv check of its own.
func TestImagePullSecretCredentialsResolveFromPassEnv(t *testing.T) {
	t.Setenv("SOLACE_TEST_REGISTRY_PASS", "from-env")
	// minimalK8s cannot be reused here: it already writes its own image:
	// section, and a second one in the body would collide with it. So this
	// mirrors minimalK8s's own fields by hand, plus the image.user/passEnv
	// pair under test.
	yaml := `redundancy:
  enabled: "false"
image:
  repo: solace/broker
  tag: latest
  user: reguser
  passEnv: SOLACE_TEST_REGISTRY_PASS
semp:
  adminPass: s3cret
kubernetes:
  name: mybroker
  namespace: sol-ns
  storage:
    msgNodeSize: 30Gi
`
	c, err := Load(writeTempYAML(t, yaml), K8s)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.Image.Pass != "from-env" {
		t.Fatalf("image.pass = %q, want the value of SOLACE_TEST_REGISTRY_PASS", c.Image.Pass)
	}
	if !c.ManagesImagePullSecret() {
		t.Error("passEnv-resolved credentials must count the same as a literal image.pass")
	}
	want := "mybroker-image-pull"
	if got := c.ImagePullSecretName(); got != want {
		t.Errorf("ImagePullSecretName() = %q, want %q", got, want)
	}
}

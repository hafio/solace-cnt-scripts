package container

import (
	"fmt"

	"solace/internal/broker"
	"solace/internal/config"
	"solace/internal/render"
)

// previewValue stands in for a file-backed secret under the Echo runner. It is
// non-empty on purpose: the entry must survive so the child-environment list still
// has exactly one variable per secret, and engine.MaskEnv prints it as *** anyway.
const previewValue = "(preview)"

// ResolveSecretValues fills in the secrets whose bytes are not in the env file.
//
// Every credential secret carries its own value already; the server certificate
// does not, because internal/render is pure and may not open a file. So this is the
// seam where the bundle is actually read, and it lives here because this is the
// package that already does host file I/O.
//
// preview substitutes a placeholder instead of reading, and it is not a convenience.
// `compose()` runs on six docker verbs, so without it every argv test of any of them
// would fail before the certificate files exist -- losing exactly the coverage
// prepareSecrets goes out of its way to keep for the pre-shared key. The placeholder keeps the ENTRY and blanks only the value:
// dropping the entry would break the deliberate "one child variable per secret"
// invariant.
//
// The read goes through broker.ServerCertBundle rather than concatenating here,
// because that function is the only place that refuses a tls.cert which already
// contains a private key -- a bundle with the key twice would otherwise reach the
// broker.
func ResolveSecretValues(cfg *config.Config, p config.Platform, preview bool) ([]render.ContainerSecret, error) {
	secrets := render.ContainerSecrets(cfg, p)
	for i := range secrets {
		if len(secrets[i].SourceFiles) == 0 {
			continue
		}
		if preview {
			secrets[i].Value = previewValue
			continue
		}
		bundle, err := broker.ServerCertBundle(cfg)
		if err != nil {
			// ServerCertBundle already names the offending file in its own wrap, so
			// this adds the secret and the env-file key rather than restating the path.
			return nil, fmt.Errorf("build the server certificate bundle for secret %q (from %s): %w",
				secrets[i].Name, secrets[i].ConfigKey, err)
		}
		secrets[i].Value = string(bundle)
	}
	return secrets, nil
}

package examples

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"solace/internal/config"
)

// update regenerates env/sample.yaml from assets/full.yaml. The sample is a
// golden, so this test is also the generator -- the same arrangement the cli,
// render and k8s packages use for theirs.
var update = flag.Bool("update", false, "regenerate env/sample.yaml from assets/full.yaml")

// samplePath is anchored on the package dir, which is `go test`'s cwd.
const samplePath = "../../env/sample.yaml"

// TestSampleYAMLMatchesTheFullExample is the drift gate between the template the
// binary prints and the copy committed at env/sample.yaml, which three packages
// load as their shared fixture and every doc points at. They must be byte
// identical: `examples full` is what a user without a checkout gets, and the
// sample is what a user with one is told to copy.
func TestSampleYAMLMatchesTheFullExample(t *testing.T) {
	full, err := Get("full")
	if err != nil {
		t.Fatalf("Get(full): %v", err)
	}

	if *update {
		if err := os.WriteFile(samplePath, full.Body, 0o644); err != nil {
			t.Fatalf("write %s: %v", samplePath, err)
		}
		return
	}
	want, err := os.ReadFile(samplePath)
	if err != nil {
		t.Fatalf("read %s (regenerate: go test ./internal/examples -update): %v", samplePath, err)
	}
	if !bytes.Equal(full.Body, want) {
		t.Errorf("%s is stale: it is GENERATED from internal/examples/assets/full.yaml -- "+
			"edit the asset, then regenerate with: go test ./internal/examples -update\n%s",
			samplePath, firstDiff(full.Body, want))
	}
}

// firstDiff names the first line that differs, so the failure says what changed
// instead of dumping two 340-line documents.
func firstDiff(got, want []byte) string {
	g, w := strings.Split(string(got), "\n"), strings.Split(string(want), "\n")
	for i := 0; i < len(g) || i < len(w); i++ {
		gl, wl := "", ""
		if i < len(g) {
			gl = g[i]
		}
		if i < len(w) {
			wl = w[i]
		}
		if gl != wl {
			return fmt.Sprintf("first difference at line %d:\n  asset:     %q\n  committed: %q", i+1, gl, wl)
		}
	}
	return ""
}

// write puts an example's bytes on disk so the real loader can read it, which is
// the only way to exercise DetectPlatforms (it reads the raw file, because a
// decoded Config cannot say which sections were present).
func write(t *testing.T, e Example) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), e.Name+".yaml")
	if err := os.WriteFile(path, e.Body, 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

// TestEveryExampleLoads is the reason these templates can be shipped: each one is
// loaded and validated under every platform it declares, so a schema change that
// invalidates an example fails this gate instead of a user's first deploy.
func TestEveryExampleLoads(t *testing.T) {
	for _, e := range All() {
		t.Run(e.Name, func(t *testing.T) {
			path := write(t, e)
			platforms, err := config.DetectPlatforms(path)
			if err != nil {
				t.Fatalf("DetectPlatforms: %v", err)
			}
			if len(platforms) == 0 {
				t.Fatal("example declares no platform section, so the CLI could never resolve one")
			}
			for _, p := range platforms {
				if _, err := config.Load(path, p); err != nil {
					t.Errorf("Load as %s: %v", p, err)
				}
			}
		})
	}
}

// TestStartersDeclareExactlyOnePlatform is what makes a starter usable with no
// flags: one declared section means resolvePlatform picks it silently, so
// `examples docker -o dev.yaml && deploy all -e dev.yaml` needs no --platform.
func TestStartersDeclareExactlyOnePlatform(t *testing.T) {
	for _, name := range []string{"kubernetes", "docker", "podman"} {
		t.Run(name, func(t *testing.T) {
			e, err := Get(name)
			if err != nil {
				t.Fatalf("Get(%s): %v", name, err)
			}
			platforms, err := config.DetectPlatforms(write(t, e))
			if err != nil {
				t.Fatalf("DetectPlatforms: %v", err)
			}
			if len(platforms) != 1 || string(platforms[0]) != name {
				t.Errorf("declared platforms = %v, want exactly [%s]", platforms, name)
			}
		})
	}
}

// TestFullDeclaresEveryPlatform pins the other half of that contract: the schema
// example populates all three sections so every key has one place to look it up,
// which is exactly why it is the one example that needs --platform.
func TestFullDeclaresEveryPlatform(t *testing.T) {
	e, err := Get("full")
	if err != nil {
		t.Fatalf("Get(full): %v", err)
	}
	platforms, err := config.DetectPlatforms(write(t, e))
	if err != nil {
		t.Fatalf("DetectPlatforms: %v", err)
	}
	if len(platforms) != len(config.Platforms()) {
		t.Errorf("declared platforms = %v, want all of %v", platforms, config.Platforms())
	}
}

// TestExamplesCarryNoRealSecret is the S3 check on shipped text: every template
// must load with a placeholder admin password, so an unedited copy fails at the
// broker rather than deploying with a password that is public in this repo.
func TestExamplesCarryNoRealSecret(t *testing.T) {
	for _, e := range All() {
		t.Run(e.Name, func(t *testing.T) {
			path := write(t, e)
			platforms, err := config.DetectPlatforms(path)
			if err != nil {
				t.Fatalf("DetectPlatforms: %v", err)
			}
			cfg, err := config.Load(path, platforms[0])
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if !strings.Contains(cfg.Admin.Pass, "CHANGE-ME") {
				t.Errorf("admin.pass = %q, want a CHANGE-ME placeholder", cfg.Admin.Pass)
			}
		})
	}
}

// TestGetRejectsAnUnknownName covers the error boundary: a typo is answered with
// every name it could have meant, since the fix is one of four words.
func TestGetRejectsAnUnknownName(t *testing.T) {
	_, err := Get("kube")
	if err == nil {
		t.Fatal("Get(kube) should fail: the package takes canonical names only")
	}
	for _, want := range Names() {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not name %q", err, want)
		}
	}
}

// TestEveryNameResolves walks the declared table against the embedded files, so a
// name added without its asset (or with a path the //go:embed pattern misses)
// fails here rather than panicking in front of a user.
func TestEveryNameResolves(t *testing.T) {
	names := Names()
	if len(names) != len(All()) {
		t.Fatalf("Names() = %d entries, All() = %d", len(names), len(All()))
	}
	for _, name := range names {
		e, err := Get(name)
		if err != nil {
			t.Fatalf("Get(%s): %v", name, err)
		}
		if len(e.Body) == 0 {
			t.Errorf("%s: empty body", name)
		}
		if e.Short == "" {
			t.Errorf("%s: no description for the listing", name)
		}
	}
}

// TestListNamesEveryExample pins the error/help phrasing on the table rather than
// on a list typed beside it, the same rule internal/abbrev's Set.List follows.
func TestListNamesEveryExample(t *testing.T) {
	got := List()
	for _, name := range Names() {
		if !strings.Contains(got, name) {
			t.Errorf("List() = %q, missing %q", got, name)
		}
	}
	if !strings.Contains(got, " or ") {
		t.Errorf("List() = %q, want the trailing %q form", got, " or ")
	}
}

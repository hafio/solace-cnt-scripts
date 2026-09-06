package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"solace/internal/config"
	"solace/internal/examples"
)

// TestExamplesListsWhatItHas: bare `examples` names every template on STDOUT, so
// `examples > list.txt` captures the list. It acts on nothing, which is why the
// bare form prints instead of failing like a verb group would.
func TestExamplesListsWhatItHas(t *testing.T) {
	out, err := runRoot(t, []string{"examples"})
	if err != nil {
		t.Fatalf("examples err = %v, want nil", err)
	}
	for _, name := range examples.Names() {
		if !strings.Contains(out, name) {
			t.Errorf("listing does not name %q:\n%s", name, out)
		}
	}
}

// TestExamplesEmitsToStdout covers the default sink for each name, and asserts the
// emitted body is the package's own bytes -- the artifact is this command's only
// stdout, exactly like `generate`.
func TestExamplesEmitsToStdout(t *testing.T) {
	for _, name := range examples.Names() {
		t.Run(name, func(t *testing.T) {
			out, err := runRoot(t, []string{"examples", name})
			if err != nil {
				t.Fatalf("examples %s err = %v, want nil", name, err)
			}
			ex, err := examples.Get(name)
			if err != nil {
				t.Fatalf("Get(%s): %v", name, err)
			}
			if out != string(ex.Body) {
				t.Errorf("examples %s stdout is not the template verbatim", name)
			}
		})
	}
}

// TestExamplesAcceptsAPlatformAbbreviation pins the CLI-wide rule that an
// abbreviation works wherever the word does: --platform dk and `examples dk` name
// the same platform, expanded by the same parser at the point of use.
func TestExamplesAcceptsAPlatformAbbreviation(t *testing.T) {
	for short, canonical := range map[string]string{"kube": "kubernetes", "dk": "docker", "pm": "podman"} {
		t.Run(short, func(t *testing.T) {
			got, err := runRoot(t, []string{"examples", short})
			if err != nil {
				t.Fatalf("examples %s err = %v, want nil", short, err)
			}
			want, err := runRoot(t, []string{"examples", canonical})
			if err != nil {
				t.Fatalf("examples %s err = %v, want nil", canonical, err)
			}
			if got != want {
				t.Errorf("examples %s and examples %s emit different templates", short, canonical)
			}
		})
	}
}

// TestExamplesWritesOutFile covers the --out branch: the file lands on disk, and
// the confirmation goes to stderr so stdout stays empty and pipe-safe.
func TestExamplesWritesOutFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dev.yaml")
	var out string
	stderr := captureStderr(t, func() {
		var err error
		out, err = runRoot(t, []string{"examples", "docker", "-o", path})
		if err != nil {
			t.Fatalf("examples docker -o err = %v, want nil", err)
		}
	})
	if out != "" {
		t.Errorf("stdout = %q, want nothing: the file is the output", out)
	}
	if !strings.Contains(stderr, path) {
		t.Errorf("stderr should name the file it wrote:\n%s", stderr)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	ex, _ := examples.Get("docker")
	if string(body) != string(ex.Body) {
		t.Error("--out did not write the template verbatim")
	}
	// The point of the whole command: what it writes is loadable as-is.
	if _, err := config.Load(path, config.Docker); err != nil {
		t.Errorf("the written example does not load: %v", err)
	}
}

// TestExamplesRefusesToOverwrite is the destructive-path guard: an existing --out
// path is kept, the error names the flag that would replace it, and --force is
// what actually replaces it.
func TestExamplesRefusesToOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dev.yaml")
	if err := os.WriteFile(path, []byte("keep me\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := runRoot(t, []string{"examples", "docker", "-o", path})
	if err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("err = %v, want a refusal naming --force", err)
	}
	if body, _ := os.ReadFile(path); string(body) != "keep me\n" {
		t.Error("the refused write still replaced the file")
	}
	if _, err := runRoot(t, []string{"examples", "docker", "-o", path, "--force"}); err != nil {
		t.Fatalf("examples docker -o --force err = %v, want nil", err)
	}
	if body, _ := os.ReadFile(path); string(body) == "keep me\n" {
		t.Error("--force did not replace the file")
	}
}

// TestExamplesRejectsAnUnknownName covers the error boundary: an unknown word is
// refused naming every alternative, and nothing is written.
func TestExamplesRejectsAnUnknownName(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dev.yaml")
	_, err := runRoot(t, []string{"examples", "k8s", "-o", path})
	if err == nil {
		t.Fatal("examples k8s should fail: neither an example nor an accepted platform word")
	}
	for _, name := range examples.Names() {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not name %q", err, name)
		}
	}
	if _, statErr := os.Stat(path); statErr == nil {
		t.Error("a rejected name still wrote the --out file")
	}
}

// TestExamplesNeedsNoEnvFile pins the reason this command is a bare literal rather
// than a leaf: it emits an env file, so it must work before one exists. A future
// wireExec added here would fail this instead of failing a first-time user.
func TestExamplesNeedsNoEnvFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist.yaml")
	for _, args := range [][]string{
		{"examples", "--env", missing},
		{"examples", "full", "--env", missing},
	} {
		out, err := runRoot(t, args)
		if err != nil {
			t.Errorf("%v err = %v, want nil", args, err)
		}
		if out == "" {
			t.Errorf("%v produced no output", args)
		}
	}
}

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"solace/internal/config"
	"solace/internal/examples"
)

// TestExamplesBareEmitsTheFullSchema: with no --platform, `examples` writes the complete
// annotated schema to STDOUT. There is no name list to print any more -- the templates are
// selected by --platform, which completion already enumerates, so a listing would be a
// second vocabulary for the same three words.
func TestExamplesBareEmitsTheFullSchema(t *testing.T) {
	out, err := runRoot(t, []string{"examples"})
	if err != nil {
		t.Fatalf("examples err = %v, want nil", err)
	}
	full, err := examples.Get("full")
	if err != nil {
		t.Fatalf("Get(full): %v", err)
	}
	if out != string(full.Body) {
		t.Error("bare examples is not the full template verbatim")
	}
}

// TestExamplesEmitsToStdout covers the default sink for each name, and asserts the
// emitted body is the package's own bytes -- the artifact is this command's only
// stdout, exactly like `generate`.
func TestExamplesEmitsToStdout(t *testing.T) {
	for _, name := range examples.Names() {
		t.Run(name, func(t *testing.T) {
			args := []string{"examples"}
			if name != "full" {
				args = append(args, "--platform", name)
			}
			out, err := runRoot(t, args)
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

// TestExamplesAcceptsAPlatformAbbreviation pins the CLI-wide rule that an abbreviation
// works wherever the word does. It is the same --platform flag every other command reads,
// but it resolves differently here and that is the point worth pinning: everywhere else
// --platform must name a section the env file declares, while `examples` has no env file
// to check against and expands the word with config.ParsePlatform alone.
func TestExamplesAcceptsAPlatformAbbreviation(t *testing.T) {
	for short, canonical := range map[string]string{"kube": "kubernetes", "dk": "docker", "pm": "podman"} {
		t.Run(short, func(t *testing.T) {
			got, err := runRoot(t, []string{"examples", "--platform", short})
			if err != nil {
				t.Fatalf("examples --platform %s err = %v, want nil", short, err)
			}
			want, err := runRoot(t, []string{"examples", "--platform", canonical})
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
		out, err = runRoot(t, []string{"examples", "--platform", "docker", "-o", path})
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

// TestExamplesRefusesToOverwrite is the destructive-path guard. An existing --out path is
// CONFIRMED rather than requiring a second flag: --force is gone, so the question is asked
// the way every other destructive question in this tree is asked, and --no-prompt is the
// one silencer. A non-interactive run without it keeps the file and says which flag would
// have proceeded -- overwriting an env file someone has already edited is the worst
// accident this command can have, and it must not happen unattended.
func TestExamplesRefusesToOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dev.yaml")
	if err := os.WriteFile(path, []byte("keep me\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := runRoot(t, []string{"examples", "--platform", "docker", "-o", path})
	if err == nil {
		t.Fatal("an existing --out path must not be replaced without confirmation")
	}
	if body, _ := os.ReadFile(path); string(body) != "keep me\n" {
		t.Error("the refused write still replaced the file")
	}
	if _, err := runRoot(t, []string{"examples", "--platform", "docker", "-o", path, "--no-prompt"}); err != nil {
		t.Fatalf("examples docker -o --no-prompt err = %v, want nil", err)
	}
	if body, _ := os.ReadFile(path); string(body) == "keep me\n" {
		t.Error("--no-prompt did not replace the file")
	}
	// The flag it replaced must be gone, not quietly still accepted.
	if _, err := runRoot(t, []string{"examples", "--platform", "docker", "-o", path, "--force"}); err == nil {
		t.Error("--force was removed; it must be an unknown flag, not a silent no-op")
	}
}

// TestExamplesRejectsAnUnknownPlatform covers the error boundary: an unknown word is
// refused naming every alternative, and nothing is written. `k8s` is the interesting value
// -- it is the spelling people reach for and is deliberately NOT accepted anywhere, so the
// refusal has to list what is.
func TestExamplesRejectsAnUnknownPlatform(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dev.yaml")
	_, err := runRoot(t, []string{"examples", "--platform", "k8s", "-o", path})
	if err == nil {
		t.Fatal("examples --platform k8s should fail: k8s is not an accepted platform word")
	}
	for _, name := range []string{"kubernetes", "docker", "podman"} {
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
		{"examples", "--platform", "docker", "--env", missing},
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

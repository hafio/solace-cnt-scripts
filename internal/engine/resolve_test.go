package engine

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hafio/go-deploy-utility/runner"
)

// tracePrefix is the per-call announcement --verbose installs. It carries the house
// `==> ` prefix so the line reads as output rather than as stray debug text.
const tracePrefix = "==> exec: "

// verboseExec builds the runner --verbose produces, announcing into a buffer.
func verboseExec(t *testing.T) (runner.Exec, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	return NewExec(&buf, true), &buf
}

// TestExecVerboseAnnouncesEveryCommand: under --verbose, the announcer installed on the
// runner writes the path the runner resolved and the arguments, quoted as one command
// line, before each command runs. The allowlist in config guarantees the NAME is one this
// tool drives; this line is what makes the LOCATION visible at the moment it matters. It
// goes to the injected writer (stderr in production) so it never contaminates a rendered
// artifact on stdout. The runner calls Announce with the resolved path right before it
// starts a command (the shared runner's own tests pin that), so what is pinned here is
// the line this tool makes of it, without starting a process.
func TestExecVerboseAnnouncesEveryCommand(t *testing.T) {
	e, buf := verboseExec(t)
	if e.Announce == nil {
		t.Fatal("a verbose runner has no announcer")
	}
	e.Announce("/usr/bin/kubectl", []string{"get", "pods", "-n", "a b"})
	want := tracePrefix + "/usr/bin/kubectl get pods -n 'a b'\n"
	if got := buf.String(); got != want {
		t.Fatalf("announcement = %q, want %q", got, want)
	}
	// Verbose means EVERY call, not the first: a trail with one entry per binary would
	// not answer "what exactly did this run issue?", which is why the flag exists.
	e.Announce("/usr/bin/kubectl", []string{"version"})
	if got := strings.Count(buf.String(), tracePrefix); got != 2 {
		t.Errorf("two calls announced %d lines, want 2", got)
	}
}

// TestExecIsSilentWithoutVerbose pins the default: the runner announces nothing, because
// the CLI lists the binaries an env file chose once in its preamble. Announcing per call
// as well put the same resolved path between report lines on every command, which read as
// debug output rather than as part of the report.
func TestExecIsSilentWithoutVerbose(t *testing.T) {
	var buf bytes.Buffer
	// nil Announce is the quiet default, not a hole that falls back to stderr, and the
	// bare runner.Exec{} a test or a hand-built runner produces has it too.
	if NewExec(&buf, false).Announce != nil {
		t.Error("the default runner has an announcer, want none")
	}
	if (runner.Exec{}).Announce != nil {
		t.Error("a bare runner.Exec has an announcer, want none")
	}
	if buf.Len() != 0 {
		t.Errorf("the default runner announced %q, want nothing", buf.String())
	}
}

// TestChildEnvNamesAreNotSystemVariables pins the layer-6 environment rule at the
// place the environment is actually built. Cmd.Env on the compose deploy is the ONLY
// path in this codebase that adds variables to a child (every other holder wraps its
// runner in NoEnv), and the values it carries are secrets -- so the
// question that matters is whether anything could name a variable the child's
// loader reads. Nothing here may be PATH, LD_PRELOAD or their relatives.
//
// The upstream half (that a config-derived secret name always keeps a fixed literal
// suffix, so it cannot fold to a bare PATH) is pinned in container's
// TestComposeSecretEnvNamesCannotBeSystemVars; this end asserts the invariant holds
// for whatever it is handed.
func TestChildEnvNamesAreNotSystemVariables(t *testing.T) {
	dangerous := []string{
		"PATH", "Path", "LD_PRELOAD", "LD_LIBRARY_PATH", "DYLD_INSERT_LIBRARIES",
		"PATHEXT", "SHELL", "IFS", "BASH_ENV", "ENV",
	}
	// The names this tool really passes, from render.ContainerSecret.EnvVar.
	passed := []string{
		"SOLACE_ADMIN_PASSWORD=x",
		"SOLACE_REDUNDANCY_PSK=y",
		"SOLACE_TLS_PASSPHRASE=z",
		"SOLACE_USER_APPUSER_PASSWORD=w",
	}
	for _, pair := range passed {
		name, _, _ := strings.Cut(pair, "=")
		for _, bad := range dangerous {
			if strings.EqualFold(name, bad) {
				t.Errorf("child environment would define %q, which the loader reads", name)
			}
		}
	}
	// And the masking that keeps values out of echoed output still holds for them.
	masked := runner.MaskEnv(passed)
	for _, secret := range []string{"x", "y", "z", "w"} {
		if strings.Contains(masked, "="+secret) {
			t.Errorf("runner.MaskEnv(%v) leaked a value: %q", passed, masked)
		}
	}
}

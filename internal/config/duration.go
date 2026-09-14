package config

import (
	"fmt"
	"time"
)

// CanonicalDuration parses an operator-supplied duration and returns it in Go's own
// canonical spelling.
//
// It exists so a free-form string never becomes an argv token. Both engines and
// kubectl take a duration for `logs --since`, and the value arrives from a command
// line, so the alternative to this is either trusting it or growing another charset
// check. Round-tripping through a trusted parser is better than either: what comes
// back is produced by time.Duration.String(), so it is provably digits plus the unit
// letters, and no consumer downstream has to police it.
//
// It lives in internal/config because that is the only package BOTH argv builders
// already import -- internal/k8s and internal/container each depend on it and neither
// depends on the other -- so this is the one place a shared helper creates no cycle.
// Canonicalising here rather than in the CLI layer also means Cluster.Logs and
// Manager.Logs can assume canonical input from any caller, which is the same argument
// the engine-returned name check makes: validate at the boundary rather than trusting
// every consumer.
//
// An empty value is not an error -- an unset flag means "no --since at all" -- and it
// comes back empty so a caller can test for it.
//
// One capability this deliberately narrows: docker and podman also accept an RFC3339
// timestamp for --since, and kubectl does not. Accepting only a duration keeps one
// spelling meaning one thing on every platform, at the cost of the timestamp form.
func CanonicalDuration(field, s string) (string, error) {
	if s == "" {
		return "", nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return "", fmt.Errorf("%s %q is not a duration: use a number with a unit, like 30s, 5m, 2h or 1h30m "+
			"(a timestamp is not accepted, so one spelling means the same thing on every platform)", field, s)
	}
	if d < 0 {
		return "", fmt.Errorf("%s %q is negative: it names how far back to read, so it must be positive", field, s)
	}
	return d.String(), nil
}

package config

import (
	"strings"
	"testing"
)

// TestCanonicalDuration is the round-trip that keeps a free-form string out of an
// argv. What comes back is produced by time.Duration.String(), so it is provably
// digits plus unit letters and no consumer downstream has to police it.
func TestCanonicalDuration(t *testing.T) {
	// Accepted, and normalised: the same span written differently comes back the same,
	// which is the property that makes the output predictable rather than merely safe.
	for in, want := range map[string]string{
		"":       "",
		"30s":    "30s",
		"5m":     "5m0s",
		"2h":     "2h0m0s",
		"1h30m":  "1h30m0s",
		"90m":    "1h30m0s",
		"1000ms": "1s",
	} {
		got, err := CanonicalDuration("--since", in)
		if err != nil {
			t.Errorf("CanonicalDuration(%q) = %v, want accepted", in, err)
			continue
		}
		if got != want {
			t.Errorf("CanonicalDuration(%q) = %q, want %q", in, got, want)
		}
	}

	// Refused, each naming the field and saying what a duration looks like.
	for _, bad := range []string{
		"abc",                  // not a duration at all
		"5",                    // no unit
		"2026-09-08T00:00:00Z", // a timestamp: the engines take one, kubectl does not
		"5 m",                  // whitespace
		"$(date)",              // the shape the charset gate would otherwise have to catch
		"-5m",                  // negative: it names how far BACK to read
	} {
		got, err := CanonicalDuration("--since", bad)
		if err == nil {
			t.Errorf("CanonicalDuration(%q) = %q, want refused", bad, got)
			continue
		}
		if !strings.Contains(err.Error(), "--since") {
			t.Errorf("CanonicalDuration(%q) error %q must name the field", bad, err)
		}
	}

	// The returned value can never need charset checking, which is the whole point.
	for _, in := range []string{"30s", "1h30m", "1000ms", "2h45m10s"} {
		got, _ := CanonicalDuration("--since", in)
		if err := CheckHostPath("--since", got); err != nil {
			t.Errorf("the canonical form %q should be inert by construction, but the charset gate refused it: %v",
				got, err)
		}
	}
}

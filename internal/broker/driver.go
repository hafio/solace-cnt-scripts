package broker

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// The apply half of `broker perform import-config`: an artifact is split into
// ordered CHUNKS, and one generated bash script applies them all inside the broker.
//
// Three problems shape this file, and each rules out the obvious approach.
//
//  1. `cli -Apes` exits 0 even when the broker REJECTS a line, so nothing about the
//     process says whether the configuration landed. The only in-band signal is the
//     transcript text, which is why the bash original (bash/059-execute-cli.sh) and
//     ExecCLI both grep it. Import used to do neither.
//  2. A rejected line does not stop the rest of a script, so one bad line used to
//     poison everything after it silently. The broker's own `source script ...
//     stop-on-error` fixes that at the source -- it is the broker, not this tool,
//     that stops at the first bad line.
//  3. Doing 1 and 2 per chunk from the host would mean an exec per chunk. On a real
//     broker that is dozens of `kubectl exec` round trips for one import. So the
//     loop is pushed INTO the broker: one upload of this script, one exec, and the
//     script reports back per chunk.
//
// What comes back is deliberately minimal. A CLI transcript repeats every credential
// in the configuration it applied, so success transcripts never leave the broker --
// only a per-chunk marker line. On FAILURE the tail is echoed back, because a chunk
// that failed with no detail is not actionable, and a `trap` deletes every temp file
// on the way out whatever happened.

// chunkFilePrefix is the in-broker basename each chunk is written to. Chunks are
// addressed by INDEX, never by their human name: a chunk is named after a
// message-VPN, and a real VPN is called `A VPN WITH LONG NAME AND SPACES`, which is
// not something to be interpolating into a shell filename or a marker line. The
// index is the only thing that crosses into bash; the app maps it back.
const chunkFilePrefix = "solace-util-chunk-"

// driverWrapperName is the one-line script handed to `cli -Apes`, rewritten before
// each chunk. Its whole content is a `source script` line, because that is what
// carries `stop-on-error`: `cli -Apes <chunk>` would run the chunk with no such
// option and keep going past a rejected line.
//
// VERIFIED against Solace's own CLI reference (semp/appliance_cli_reference.html):
//
//	COMMAND: source script <script-name> [stop-on-error] [no-prompt]
//	DESCRIPTION: Use this command to run a cli script.
//
// -- so both options are real and optional, and the operand is a bare script name
// resolved inside the cliscripts directory, which is why chunks are written there
// under a plain name rather than the dot-and-.cli form `cli -Apes` takes.
const driverWrapperName = "solace-util-run"

// Marker lines the driver prints and parseDriverOutput reads. They are the entire
// protocol between the two, so they are matched exactly rather than searched for.
const (
	markerChunkOK    = "SOLACE-UTIL-CHUNK-OK"
	markerChunkFail  = "SOLACE-UTIL-CHUNK-FAIL"
	markerDetailOpen = "SOLACE-UTIL-DETAIL-BEGIN"
	markerDetailShut = "SOLACE-UTIL-DETAIL-END"
	markerDone       = "SOLACE-UTIL-DONE"
)

// driverTailLines is how much of a chunk's transcript the failure check reads.
//
// The tail is the right place to look BECAUSE of stop-on-error: the broker abandons
// the chunk at the offending line, so the rejection is the last thing in the
// transcript. Scanning the whole transcript would instead scan the configuration
// itself, which is what makes a keyword scan a bad idea in general -- an object
// legitimately named `error-events` would match.
const driverTailLines = 10

// failKeywords are the phrases that mean the broker rejected something.
//
// This list is the operator's and is COMPLETE as it stands (confirmed 2026-09-13; it
// began as a placeholder on 2026-09-12). It is not provisional and no caller should
// treat it as approximate -- a new phrase is added here when real rejection output turns
// one up, as its own change. They live in one place, so that is a one-line edit and no
// caller has to be touched; `broker configure data-replication` scans with the same list
// for the same reason.
//
// A phrase MISSING from this list is a silent hole rather than a loud one: the chunk
// applies, the driver reports OK, and only the verification diff afterwards notices
// that the configuration is not there. That is the designed fallback -- the two
// detectors catch different things (see importops.go) -- but it is why this list
// wants to be generous rather than minimal.
//
// The phrases are deliberately specific rather than the bare `invalid` / `error`
// that bash/059-execute-cli.sh greps for. Measured against real captures --
// semp/all.cli (33,728 lines) and a production export (33,638) -- the phrase forms
// match nothing in configuration text, including all 917 `reject-msg-*` lines, so
// they carry no false-positive cost that a looser pattern would.
var failKeywords = []string{
	"invalid command",
	"error:",
	"command not found",
	"busy",
}

// chunk is one unit of application: an ordered CLI script with a name for the report.
type chunk struct {
	Name     string   // human name, for the report only -- never reaches bash
	Body     string   // the CLI script
	Sections []string // broker-level sections this chunk applies, for the report
	// TearsDown names the message-VPN this chunk destroys, empty for every other
	// chunk. The result records it only once the driver reports the chunk applied:
	// a captured teardown is not an executed one, and the difference is whether the
	// operator is told their spooled messages are gone.
	TearsDown string
}

// chunkResult is what the driver reported for one chunk.
type chunkResult struct {
	Index  int
	Name   string
	OK     bool
	Detail string // the failing transcript tail; empty on success
}

// heredocDelim terminates every chunk heredoc. It is quoted at the open (<<'...'),
// so bash performs NO expansion on the body -- a chunk full of `$`, backticks and
// quotes is written through byte for byte.
const heredocDelim = "SOLACE_UTIL_CHUNK_EOF"

// renderDriver builds the bash script that applies every chunk inside the broker.
//
// It is a pure function of the chunks so the whole thing is golden-testable: what
// this returns is exactly what runs against a production broker, and there is no
// second code path that a test could pass while production fails.
func renderDriver(chunks []chunk, cliScriptsDir, cliBinary string) (string, error) {
	for i, c := range chunks {
		// A body carrying the delimiter would end its own heredoc and the rest would
		// be read as bash. Nothing in `show current-config` output looks like this,
		// so refusing is right: it means the artifact is not what it claims to be.
		if strings.Contains(c.Body, heredocDelim) {
			return "", fmt.Errorf("chunk %d (%q) contains the heredoc delimiter %q, so it cannot be "+
				"written safely; this is not `show current-config` output", i, c.Name, heredocDelim)
		}
	}

	var b strings.Builder
	b.WriteString("#!/bin/sh\n")
	b.WriteString("# Generated by `solace-util broker perform import-config`. Do not edit.\n")
	b.WriteString("# Applies each chunk through the broker's own `source script ... stop-on-error`,\n")
	b.WriteString("# checks the transcript tail, and reports one marker line per chunk.\n")
	b.WriteString("set -u\n\n")

	fmt.Fprintf(&b, "CLIDIR=%s\n", shQuote(cliScriptsDir))
	fmt.Fprintf(&b, "CLI=%s\n", shQuote(cliBinary))
	fmt.Fprintf(&b, "WRAPPER=%s\n", shQuote(driverWrapperName))
	// One transcript file, reused by every chunk: a chunk that succeeded has nothing
	// worth keeping, so the next one overwrites it. Only the failing chunk's
	// transcript is ever read, and it is read before the trap removes it.
	b.WriteString("TMP=\"$CLIDIR/." + chunkFilePrefix + "out\"\n\n")

	// EXIT covers a normal return and an error return; INT/TERM cover the operator
	// interrupting the import. Without them a failed run would leave the artifact's
	// configuration -- credentials included -- sitting in cliscripts.
	b.WriteString("cleanup() {\n")
	b.WriteString("  rm -f \"$TMP\" \"$CLIDIR/$WRAPPER\"")
	for i := range chunks {
		// Double-quoted so $CLIDIR expands; the basename is a generated literal
		// (chunkName), never operator text, so there is nothing here to quote against.
		fmt.Fprintf(&b, " \"$CLIDIR/%s\"", chunkName(i))
	}
	b.WriteString("\n}\n")
	b.WriteString("trap cleanup EXIT INT TERM\n\n")

	for i, c := range chunks {
		fmt.Fprintf(&b, "# ---- chunk %d: %s ----\n", i, sanitizeComment(c.Name))
		fmt.Fprintf(&b, "cat > \"$CLIDIR/%s\" <<'%s'\n", chunkName(i), heredocDelim)
		b.WriteString(c.Body)
		if !strings.HasSuffix(c.Body, "\n") {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "%s\n", heredocDelim)

		// The wrapper is what carries stop-on-error. `source script` resolves a plain
		// name inside the cliscripts directory, which is why chunks are written there
		// under a bare name rather than the dot-and-.cli form `cli -Apes` takes.
		fmt.Fprintf(&b, "printf 'source script %s stop-on-error no-prompt\\n' > \"$CLIDIR/$WRAPPER\"\n", chunkName(i))
		b.WriteString("\"$CLI\" -Apes \"$WRAPPER\" > \"$TMP\" 2>&1\n")
		fmt.Fprintf(&b, "if tail -n %d \"$TMP\" | grep -qiE %s; then\n", driverTailLines, shQuote(keywordPattern()))
		// Detail before cleanup, and only on failure: a chunk that failed with no
		// detail is not actionable, but a successful chunk's transcript is pure
		// credential-bearing noise.
		fmt.Fprintf(&b, "  echo %s\n", shQuote(markerDetailOpen))
		fmt.Fprintf(&b, "  tail -n %d \"$TMP\"\n", driverTailLines)
		fmt.Fprintf(&b, "  echo %s\n", shQuote(markerDetailShut))
		fmt.Fprintf(&b, "  echo %s\n", shQuote(fmt.Sprintf("%s %d", markerChunkFail, i)))
		b.WriteString("  exit 1\n")
		b.WriteString("fi\n")
		fmt.Fprintf(&b, "echo %s\n\n", shQuote(fmt.Sprintf("%s %d", markerChunkOK, i)))
	}

	// A terminator, so a truncated transport read is distinguishable from a run that
	// genuinely applied every chunk. Without it, output cut short after the last OK
	// marker would read as complete success.
	fmt.Fprintf(&b, "echo %s\n", shQuote(markerDone))
	return b.String(), nil
}

// chunkName is the in-broker basename for chunk i.
func chunkName(i int) string { return fmt.Sprintf("%s%03d", chunkFilePrefix, i) }

// keywordPattern is the grep -E alternation the tail is matched against.
func keywordPattern() string {
	quoted := make([]string, len(failKeywords))
	for i, k := range failKeywords {
		quoted[i] = regexp.QuoteMeta(k)
	}
	return strings.Join(quoted, "|")
}

// shQuote renders s as a single-quoted shell word. A single quote inside is closed,
// escaped and reopened -- the standard POSIX form, and the only safe one, since
// there is no escape character inside single quotes.
func shQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// commentUnsafeRE strips everything a shell comment must not carry: a newline would
// end the comment and run the rest as code.
var commentUnsafeRE = regexp.MustCompile(`[^\x20-\x7e]`)

// sanitizeComment makes a chunk name safe to write into a `#` comment. The name is
// operator-facing only, so mangling an exotic one costs nothing; letting a newline
// through would let a VPN name inject shell.
func sanitizeComment(name string) string {
	return commentUnsafeRE.ReplaceAllString(name, "?")
}

// driverMarkerRE matches one OK/FAIL marker and captures its chunk index.
var driverMarkerRE = regexp.MustCompile(`^(` + markerChunkOK + `|` + markerChunkFail + `) ([0-9]+)$`)

// parseDriverOutput reads the driver's marker protocol back into per-chunk results.
//
// It is strict on purpose. This is the only place an import learns whether anything
// worked, so every ambiguity resolves to an ERROR rather than to success: output
// with no terminator was truncated, a marker naming a chunk that does not exist
// means the driver and the caller disagree about what ran, output with no
// markers at all means the driver never started (a shell that could not run it, a
// container without `bash`), and a terminator that arrives after fewer results than
// chunks -- or after results that are out of order or duplicated -- means the
// transport dropped or reordered a marker somewhere in the middle rather than at the
// end, which the truncation check alone cannot see. Any of those reported as "fine"
// would certify an import that did not happen -- a dropped OK for a TEARDOWN chunk in
// particular would omit a destroyed VPN from the report.
func parseDriverOutput(out []byte, chunks []chunk) ([]chunkResult, error) {
	var (
		results  []chunkResult
		detail   []string
		inDetail bool
		done     bool
	)
	for _, raw := range strings.Split(string(out), "\n") {
		line := strings.TrimRight(raw, "\r")
		switch {
		case line == markerDetailOpen:
			inDetail, detail = true, nil
			continue
		case line == markerDetailShut:
			inDetail = false
			continue
		case inDetail:
			detail = append(detail, line)
			continue
		case line == markerDone:
			done = true
			continue
		}
		m := driverMarkerRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		i, err := strconv.Atoi(m[2])
		if err != nil || i < 0 || i >= len(chunks) {
			return nil, fmt.Errorf("the import driver reported chunk %q, which is not one of the "+
				"%d chunks it was given; treat this run as having an unknown outcome and re-export "+
				"before retrying", m[2], len(chunks))
		}
		r := chunkResult{Index: i, Name: chunks[i].Name, OK: m[1] == markerChunkOK}
		if !r.OK {
			r.Detail = strings.TrimSpace(strings.Join(detail, "\n"))
		}
		detail = nil
		results = append(results, r)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("the import driver produced no chunk markers at all, so nothing is " +
			"known about what it applied; check that the broker container can run `/bin/sh`")
	}
	// A failing chunk stops the driver, so a run that ends without the terminator is
	// only correct when the last thing reported was that failure.
	if !done && results[len(results)-1].OK {
		return nil, fmt.Errorf("the import driver's output ended after chunk %q with no completion "+
			"marker, so it was cut short; the broker may be part-configured -- re-run the import",
			results[len(results)-1].Name)
	}
	// A terminator alone does not prove every chunk was reported: a marker dropped or
	// reordered in the MIDDLE of the run still leaves the last result intact and the
	// terminator present, which the truncation check above cannot catch. Every chunk
	// must appear exactly once, in order.
	if done {
		if len(results) != len(chunks) {
			return nil, fmt.Errorf("the import driver reported %d chunk result(s) but was given %d "+
				"chunks; treat this run as having an unknown outcome and re-export before retrying",
				len(results), len(chunks))
		}
		for i, r := range results {
			if r.Index != i {
				return nil, fmt.Errorf("the import driver reported chunk %d where chunk %d was "+
					"expected next -- a marker was skipped, duplicated or reordered; treat this run "+
					"as having an unknown outcome and re-export before retrying", r.Index, i)
			}
		}
	}
	return results, nil
}

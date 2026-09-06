package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Consenter gates every mutating probe step. It mirrors internal/cli's
// promptYes: an exact "yes", nothing looser, because the thing on the other side
// of the question is a live broker and "y" is too easy to type by reflex.
//
// The prompt seams (Interactive/In) exist for the same reason App.Interactive and
// App.PromptIn do -- a test cannot supply a terminal, and the branches guarding a
// real mutation are exactly the ones worth testing.
type Consenter struct {
	Yes         bool        // --yes: pre-approved, for an unattended run
	Interactive func() bool // nil -> isTTY(os.Stdin)
	In          io.Reader   // nil -> os.Stdin
	Err         io.Writer   // nil -> os.Stderr; where the question is asked
}

// Confirm asks whether to proceed with what. A non-interactive run without --yes
// declines: nobody is there to answer, and silently taking "yes" would let a cron
// job rotate a production secret because a flag was forgotten.
func (c Consenter) Confirm(what string) bool {
	if c.Yes {
		return true
	}
	if !c.interactive() {
		return false
	}
	w := c.errOut()
	fmt.Fprintf(w, "%s\nThis touches a LIVE environment. Type 'yes' to proceed, anything else skips it: ", what)
	line, _ := bufio.NewReader(c.reader()).ReadString('\n')
	return strings.TrimSpace(line) == "yes"
}

func (c Consenter) interactive() bool {
	if c.Interactive != nil {
		return c.Interactive()
	}
	return isTTY(os.Stdin)
}

func (c Consenter) reader() io.Reader {
	if c.In != nil {
		return c.In
	}
	return os.Stdin
}

func (c Consenter) errOut() io.Writer {
	if c.Err != nil {
		return c.Err
	}
	return os.Stderr
}

// isTTY reports whether f is a character device, the same test internal/cli uses
// to decide whether a run may ask anything at all.
func isTTY(f *os.File) bool {
	if f == nil {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

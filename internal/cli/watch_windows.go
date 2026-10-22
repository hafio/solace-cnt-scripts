//go:build windows

package cli

import (
	"os"

	"golang.org/x/sys/windows"
)

// enableVT asks this console to interpret escape sequences, which a Windows console --
// PowerShell 5.1's conhost among them -- does only once a program turns it on. It reports
// whether the console now does: false (stdout is not a console, or one too old to
// support it) makes --watch append its reports instead of redrawing, rather than print
// the escape bytes as text.
func enableVT() bool {
	h := windows.Handle(os.Stdout.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return false
	}
	if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING != 0 {
		return true
	}
	return windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING) == nil
}

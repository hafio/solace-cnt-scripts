//go:build !windows

package cli

// enableVT is a no-op outside Windows: a Unix terminal interprets escape sequences
// without being asked (see watch_windows.go for the console that does not).
func enableVT() bool { return true }

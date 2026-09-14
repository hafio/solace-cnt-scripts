// Command solace is a single CLI for deploying and operating Solace PubSub+
// Event Brokers on Kubernetes (via the EventBroker Operator), Docker, or Podman.
// It presents the same lifecycle verbs on every platform. Unsupported — not a
// Solace product.
package main

import (
	"fmt"
	"os"

	"solace/internal/cli"
)

// main reports, then exits with the code cli.ExitCode chose: 0 it worked, 2 the
// command line or the env file was wrong, 1 anything else. The classification
// lives in internal/cli (exit.go), beside the errors it classifies and in a
// package that has tests -- main stays the one line that cannot be tested.
func main() {
	err := cli.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
	os.Exit(cli.ExitCode(err))
}

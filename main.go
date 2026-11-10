// Command solace is a single CLI for deploying and operating Solace PubSub+
// Event Brokers on Kubernetes (via the EventBroker Operator), Docker, or Podman.
// It presents the same lifecycle verbs on every platform. Not a supported Solace
// product: supported only by Solace Professional Services.
package main

import (
	"fmt"
	"os"

	"github.com/hafio/go-deploy-utility/exitcode"

	"solace/internal/cli"
)

// main reports, then exits with the code exitcode.Code chose: 0 it worked, 2 the
// command line or the env file was wrong, 1 anything else. The classification
// lives in the shared exitcode package, which has tests; the errors it classifies
// are marked in internal/cli -- main stays the one line that cannot be tested.
func main() {
	err := cli.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
	os.Exit(exitcode.Code(err))
}

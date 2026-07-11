// Command rescheck quickly queries NHN Cloud compute and network resources.
// See resource_checker/README.md for the full CLI Mode spec.
package main

import (
	"os"

	"github.com/opp-13/nhn_iac/resource_checker/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}

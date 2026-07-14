// Command instsched keeps a configured set of NHN Cloud instances present
// on their own schedules: each instance has a cron expression saying when it
// must exist, plus a flag saying whether a deleted instance may be recreated
// via terraform. See instance_scheduler/README.md for the full CLI spec.
package main

import (
	"os"

	"github.com/opp-13/nhn_iac/instance_scheduler/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}

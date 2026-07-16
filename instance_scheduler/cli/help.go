package cli

// These texts follow the same format as resource_checker/cli/help.go: a
// verbatim `instsched ... --help` transcript for each command.

const rootHelp = `Usage: instsched [OPTIONS] COMMAND [ARGS]...
Keep NHN Cloud instances present on their own schedules.

Commands:
  check                      run a check-and-heal pass for one or all instances
  stop                       shut down one or all instances (for a stopSchedule)
  schedule                   manage the crontab entries that call 'check'/'stop'
  help                       show help for a command

Global options:
  -h, --help                 display this help and exit
  -v, --version              output version information and exit

Run 'instsched COMMAND --help' for more information on a command.
`

const checkHelp = `Usage: instsched check [OPTIONS]
Check configured instances and heal them as needed:
  - a SHUTOFF instance is started via 'rescheck compute run'
  - a deleted instance is recreated via 'terraform apply', but only when
    that instance has terraform: true in config.yaml
  - a deleted instance with terraform: false is reported only, never
    recreated
  - any other status (ERROR, BUILD, ...) is reported only

Instance status and start are done by calling the rescheck binary on PATH.

      --instance=NAME         check only this instance (default: every
                             configured instance — mainly for manual runs;
                             crontab entries created by 'schedule add'
                             always pass --instance)
      --config=PATH          path to config.yaml (default: ./config.yaml)
`

const stopHelp = `Usage: instsched stop [OPTIONS]
Shut down configured instances that are ACTIVE (the mirror of 'check', used
for a stopSchedule — e.g. powering instances off outside business hours):
  - an ACTIVE instance is stopped via 'rescheck compute shutdown'
  - an already-SHUTOFF instance is left alone
  - a deleted instance is reported only (nothing to stop)
  - any other status (ERROR, BUILD, ...) is reported only

terraform is never consulted here — stopping has nothing to do with
recreate-on-delete.

      --instance=NAME         stop only this instance (default: every
                             configured instance — mainly for manual runs;
                             crontab entries created by 'schedule add'
                             always pass --instance)
      --config=PATH          path to config.yaml (default: ./config.yaml)
`

const scheduleHelp = `Usage: instsched schedule COMMAND [OPTIONS]
Manage the crontab entries that call 'instsched check --instance NAME' (at
startSchedule) and, if configured, 'instsched stop --instance NAME' (at
stopSchedule). Entries are tagged with a marker comment so other crontab
lines are left untouched.

Commands:
  add                        install or replace an instance's crontab entries
  remove                     remove an instance's crontab entries
  list                       list managed crontab entries
`

const scheduleAddHelp = `Usage: instsched schedule add NAME [OPTIONS]
Install (or replace) the crontab entries for NAME: a "start" entry at
startSchedule (always), and a "stop" entry at stopSchedule (only if set in
config.yaml). Both entries call this binary and config.yaml by absolute
path, since cron runs jobs with a minimal PATH.

If crontab itself fails to run, you may not have permission — try again
with sudo.

      --config=PATH          path to config.yaml (default: ./config.yaml)
`

const scheduleRemoveHelp = `Usage: instsched schedule remove NAME
Remove the crontab entries (start and/or stop) for NAME, if any exist. It
is not an error for none to exist.
`

const scheduleListHelp = `Usage: instsched schedule list
List every instsched-managed crontab entry (name, kind — start/stop — and
schedule).
`

// version is set at build time via -ldflags "-X ...cli.version=...";
// it defaults to "dev" for local builds.
var version = "dev"

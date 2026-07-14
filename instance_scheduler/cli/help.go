package cli

// These texts follow the same format as resource_checker/cli/help.go: a
// verbatim `instsched ... --help` transcript for each command.

const rootHelp = `Usage: instsched [OPTIONS] COMMAND [ARGS]...
Keep NHN Cloud instances present on their own schedules.

Commands:
  check                      run a check-and-heal pass for one or all instances
  schedule                   manage the crontab entries that call 'check'
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

const scheduleHelp = `Usage: instsched schedule COMMAND [OPTIONS]
Manage the crontab entries that call 'instsched check --instance NAME' on
each instance's configured schedule. Entries are tagged with a marker
comment so other crontab lines are left untouched.

Commands:
  add                        install or replace a crontab entry
  remove                     remove a crontab entry
  list                       list managed crontab entries
`

const scheduleAddHelp = `Usage: instsched schedule add NAME [OPTIONS]
Install (or replace) the crontab entry for NAME, using its schedule from
config.yaml. The crontab entry calls this binary and config.yaml by
absolute path, since cron runs jobs with a minimal PATH.

If crontab itself fails to run, you may not have permission — try again
with sudo.

      --config=PATH          path to config.yaml (default: ./config.yaml)
`

const scheduleRemoveHelp = `Usage: instsched schedule remove NAME
Remove the crontab entry for NAME, if one exists. It is not an error for
no entry to exist.
`

const scheduleListHelp = `Usage: instsched schedule list
List every instsched-managed crontab entry (name and schedule).
`

// version is set at build time via -ldflags "-X ...cli.version=...";
// it defaults to "dev" for local builds.
var version = "dev"

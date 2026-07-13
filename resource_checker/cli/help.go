package cli

// These texts are kept in lockstep with resource_checker/README.md's
// "## CLI Mode" section (each is a verbatim copy of one of its bash blocks,
// minus the "$ rescheck ..." prompt line). If you change one, change both.

const rootHelp = `Usage: rescheck [OPTIONS] COMMAND [ARGS]...
Quickly query NHN Cloud compute and network resources from the CLI.

Commands:
  configure                  manage stored credentials and config
  compute                    query Compute (Nova) resources
  network                    query Network (Neutron) resources
  help                       show help for a command

Global options:
  -o, --output=FORMAT        table, json, or text (default: table)
      --config=PATH          path to config.yaml (default: ./config.yaml)
  -h, --help                 display this help and exit
  -v, --version              output version information and exit

Run 'rescheck COMMAND --help' for more information on a command.
`

const configureHelp = `Usage: rescheck configure COMMAND [OPTIONS]
Manage the credentials rescheck uses to call NHN Cloud.

Commands:
  set                        write credentials to the config file
  show                       print the current config (password masked)
`

const configureSetHelp = `Usage: rescheck configure set [OPTIONS]
Write NHN Cloud credentials to the config file (under nhn.auth).

      --tenant-id=ID          NHN Cloud tenant ID; if omitted, you will be
                             prompted
      --username=USER         NHN Cloud username; if omitted, you will be
                             prompted
      --password=PASSWORD     NHN Cloud password; if omitted, you will be
                             prompted (recommended — keeps it out of shell
                             history)
      --region=REGION         NHN Cloud region; if omitted, keeps the
                             current value (default: KR1 the first time)
      --config=PATH          config file to write to; if omitted, you will be
                             prompted (default: ./config.yaml, press Enter to
                             accept it)

Note: the config file holds credentials in plaintext. Never commit it —
make sure it is listed in .gitignore.
`

const configureShowHelp = `Usage: rescheck configure show [OPTIONS]
Print the resolved config. The password field is always shown as ******,
regardless of --output format.

  -o, --output=FORMAT        table, json, or text (default: table)
      --config=PATH          config file to read (default: ./config.yaml)
`

const computeHelp = `Usage: rescheck compute COMMAND [OPTIONS]
Query Compute (Nova) resources, and start or stop instances.

Commands:
  desc, describe               describe compute resources (see 'rescheck compute desc --help')
  run, start                   start instances
  shutdown, stop               stop instances
`

const computeRunHelp = `Usage: rescheck compute run NAME|ID|GLOB [OPTIONS]
Start instances. The target is an exact instance name or ID, or a shell
glob pattern (*, ?, [...]) matched against instance names. When a glob
is used, the matched instances are listed and you are asked to confirm.
Instances already ACTIVE are skipped.

  -y, --yes                  skip the confirmation prompt for glob targets
      --config=PATH          path to config.yaml (default: ./config.yaml)

Examples:
  rescheck compute run web-01        # exact name, runs immediately
  rescheck compute run 'test*'       # glob, asks y/N first
  rescheck compute run '*' -y        # everything, no prompt (careful!)
`

const computeShutdownHelp = `Usage: rescheck compute shutdown NAME|ID|GLOB [OPTIONS]
Stop instances. The target is an exact instance name or ID, or a shell
glob pattern (*, ?, [...]) matched against instance names. When a glob
is used, the matched instances are listed and you are asked to confirm.
Instances already SHUTOFF are skipped.

  -y, --yes                  skip the confirmation prompt for glob targets
      --config=PATH          path to config.yaml (default: ./config.yaml)

Examples:
  rescheck compute shutdown web-01        # exact name, runs immediately
  rescheck compute shutdown 'test*'       # glob, asks y/N first
  rescheck compute shutdown '*' -y        # everything, no prompt (careful!)
`

const computeDescHelp = `Usage: rescheck compute desc RESOURCE [OPTIONS]
Describe Compute (Nova) resources.

Resources:
  instance                    list instances
`

const computeDescInstanceHelp = `Usage: rescheck compute desc instance [OPTIONS]
List instances. By default only ACTIVE (running) instances are shown.

  -a, --all                  include every state (SHUTOFF, ERROR, BUILD, ...),
                             not just ACTIVE
  -l, --long                 show extended columns
  -o, --output=FORMAT        table, json, or text (default: table)

Default columns:
  NAME  STATUS  PRIVATE IP  FLOATING IP  SECURITY GROUPS

With --long, also shows:
  ID  FLAVOR  IMAGE  AVAILABILITY ZONE  CREATED
`

const networkHelp = `Usage: rescheck network COMMAND [OPTIONS]
Query Network (Neutron) resources.

Commands:
  desc, describe               describe network resources (see 'rescheck network desc --help')
`

const networkDescHelp = `Usage: rescheck network desc RESOURCE [OPTIONS]
Describe Network (Neutron) resources.

Resources:
  vpc                        list networks (VPCs)
  subnet                     list subnets
  security-group             list security groups and their rules
  floating-ip                list floating IPs
`

const networkDescVPCHelp = `Usage: rescheck network desc vpc [OPTIONS]

  -l, --long                 show extended columns
  -o, --output=FORMAT        table, json, or text (default: table)

Default columns:
  NAME  STATUS  EXTERNAL  SHARED  SUBNETS

With --long, also shows:
  ID  MTU  PROVIDER NETWORK TYPE  PROJECT ID  CREATED
`

const networkDescSubnetHelp = `Usage: rescheck network desc subnet [OPTIONS]

  -l, --long                 show extended columns
  -o, --output=FORMAT        table, json, or text (default: table)

Default columns:
  NAME  CIDR  GATEWAY IP  NETWORK  DHCP ENABLED

With --long, also shows:
  ID  ALLOCATION POOLS  DNS NAMESERVERS  IP VERSION  CREATED
`

const networkDescSecurityGroupHelp = `Usage: rescheck network desc security-group [OPTIONS]

  -l, --long                 show one row per rule instead of one per group
  -o, --output=FORMAT        table, json, or text (default: table)

Default columns:
  NAME  ID  DESCRIPTION (rule count summary, e.g. "3 inbound / 1 outbound")

With --long, columns become (one row per rule):
  SG NAME  DIRECTION  ETHERTYPE  PROTOCOL  PORT RANGE  REMOTE (CIDR/SG)
`

const networkDescFloatingIPHelp = `Usage: rescheck network desc floating-ip [OPTIONS]
Floating IPs are always listed regardless of status — unattached (spare)
IPs are useful to see too, so there is no -a/--all flag here.

  -l, --long                 show extended columns
  -o, --output=FORMAT        table, json, or text (default: table)

Default columns:
  FLOATING IP  FIXED IP  STATUS  ATTACHED INSTANCE

With --long, also shows:
  ID  FLOATING NETWORK  PORT ID  CREATED
`

// version is set at build time via -ldflags "-X ...cli.version=...";
// it defaults to "dev" for local builds.
var version = "dev"

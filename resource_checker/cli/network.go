package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strconv"

	"github.com/opp-13/nhn_iac/resource_checker/config"
	"github.com/opp-13/nhn_iac/resource_checker/nhncloud"
	"github.com/opp-13/nhn_iac/resource_checker/output"
)

func runNetwork(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, networkHelp)
		return 0
	}

	switch args[0] {
	case "-h", "--help", "help":
		fmt.Fprint(stdout, networkHelp)
		return 0
	case "desc", "describe":
		return runNetworkDesc(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "rescheck network: 알 수 없는 명령어 %q\n\n", args[0])
		fmt.Fprint(stderr, networkHelp)
		return 1
	}
}

func runNetworkDesc(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, networkDescHelp)
		return 0
	}

	switch args[0] {
	case "-h", "--help", "help":
		fmt.Fprint(stdout, networkDescHelp)
		return 0
	case "vpc":
		return runNetworkDescVPC(args[1:], stdout, stderr)
	case "subnet":
		return runNetworkDescSubnet(args[1:], stdout, stderr)
	case "security-group", "sg":
		return runNetworkDescSecurityGroup(args[1:], stdout, stderr)
	case "floating-ip", "fip":
		return runNetworkDescFloatingIP(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "rescheck network desc: 알 수 없는 리소스 %q\n\n", args[0])
		fmt.Fprint(stderr, networkDescHelp)
		return 1
	}
}

// networkFlags declares the -l/--long, -o/--output, --config flags shared by
// every "network desc <resource>" subcommand. Callers must still call fs.Parse.
func networkFlags(name string) (fs *flag.FlagSet, long *bool, format, configPath *string) {
	fs = flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	long = new(bool)
	format = new(string)
	configPath = new(string)
	fs.BoolVar(long, "l", false, "")
	fs.BoolVar(long, "long", false, "")
	fs.StringVar(format, "o", output.FormatTable, "")
	fs.StringVar(format, "output", output.FormatTable, "")
	fs.StringVar(configPath, "config", config.FindConfigPath(), "")
	return fs, long, format, configPath
}

func loadClients(configPath string) (*nhncloud.Clients, error) {
	cfg, _, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	return nhncloud.NewClients(context.Background(), cfg)
}

func runNetworkDescVPC(args []string, stdout, stderr io.Writer) int {
	fs, long, format, configPath := networkFlags("network desc vpc")
	if err := fs.Parse(args); err != nil {
		return helpOrError(err, stdout, stderr, networkDescVPCHelp)
	}
	if !output.Valid(*format) {
		fmt.Fprintf(stderr, "rescheck network desc vpc: 알 수 없는 --output %q\n", *format)
		return 1
	}

	clients, err := loadClients(*configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	networks, err := nhncloud.ListNetworks(context.Background(), clients.Network)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if *format == output.FormatJSON {
		return renderJSONOrError(stdout, stderr, networks)
	}

	var headers []string
	if *long {
		headers = []string{"ID", "NAME", "STATUS", "EXTERNAL", "SHARED", "SUBNETS", "MTU", "PROVIDER NETWORK TYPE", "PROJECT ID", "CREATED"}
	} else {
		headers = []string{"NAME", "STATUS", "EXTERNAL", "SHARED", "SUBNETS"}
	}
	rows := make([][]string, 0, len(networks))
	for _, n := range networks {
		subnets := strconv.Itoa(len(n.Subnets))
		if *long {
			rows = append(rows, []string{
				n.ID, n.Name, n.Status, strconv.FormatBool(n.External), strconv.FormatBool(n.Shared),
				subnets, strconv.Itoa(n.MTU), n.ProviderType, n.ProjectID, formatTime(n.Created),
			})
		} else {
			rows = append(rows, []string{n.Name, n.Status, strconv.FormatBool(n.External), strconv.FormatBool(n.Shared), subnets})
		}
	}
	return renderTableOrError(stdout, stderr, *format, headers, rows)
}

func runNetworkDescSubnet(args []string, stdout, stderr io.Writer) int {
	fs, long, format, configPath := networkFlags("network desc subnet")
	if err := fs.Parse(args); err != nil {
		return helpOrError(err, stdout, stderr, networkDescSubnetHelp)
	}
	if !output.Valid(*format) {
		fmt.Fprintf(stderr, "rescheck network desc subnet: 알 수 없는 --output %q\n", *format)
		return 1
	}

	clients, err := loadClients(*configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	subnets, err := nhncloud.ListSubnets(context.Background(), clients.Network)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if *format == output.FormatJSON {
		return renderJSONOrError(stdout, stderr, subnets)
	}

	var headers []string
	if *long {
		headers = []string{"ID", "NAME", "CIDR", "GATEWAY IP", "NETWORK", "DHCP ENABLED", "ALLOCATION POOLS", "DNS NAMESERVERS", "IP VERSION", "CREATED"}
	} else {
		headers = []string{"NAME", "CIDR", "GATEWAY IP", "NETWORK", "DHCP ENABLED"}
	}
	rows := make([][]string, 0, len(subnets))
	for _, s := range subnets {
		if *long {
			// gophercloud's Subnet doesn't unmarshal a creation timestamp
			// (its CreatedAt field is tagged json:"-"), so CREATED is always "-".
			rows = append(rows, []string{
				s.ID, s.Name, s.CIDR, s.GatewayIP, s.NetworkID, strconv.FormatBool(s.EnableDHCP),
				nhncloud.JoinOrDash(s.AllocationPools), nhncloud.JoinOrDash(s.DNSNameservers), strconv.Itoa(s.IPVersion), "-",
			})
		} else {
			rows = append(rows, []string{s.Name, s.CIDR, s.GatewayIP, s.NetworkID, strconv.FormatBool(s.EnableDHCP)})
		}
	}
	return renderTableOrError(stdout, stderr, *format, headers, rows)
}

func runNetworkDescSecurityGroup(args []string, stdout, stderr io.Writer) int {
	fs, long, format, configPath := networkFlags("network desc security-group")
	if err := fs.Parse(args); err != nil {
		return helpOrError(err, stdout, stderr, networkDescSecurityGroupHelp)
	}
	if !output.Valid(*format) {
		fmt.Fprintf(stderr, "rescheck network desc security-group: 알 수 없는 --output %q\n", *format)
		return 1
	}

	clients, err := loadClients(*configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	groups, err := nhncloud.ListSecurityGroups(context.Background(), clients.Network)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if *format == output.FormatJSON {
		return renderJSONOrError(stdout, stderr, groups)
	}

	if !*long {
		headers := []string{"NAME", "ID", "DESCRIPTION"}
		rows := make([][]string, 0, len(groups))
		for _, g := range groups {
			desc := g.Description
			if desc == "" {
				desc = nhncloud.RuleCountSummary(g.Rules)
			} else {
				desc = fmt.Sprintf("%s (%s)", desc, nhncloud.RuleCountSummary(g.Rules))
			}
			rows = append(rows, []string{g.Name, g.ID, desc})
		}
		return renderTableOrError(stdout, stderr, *format, headers, rows)
	}

	headers := []string{"SG NAME", "DIRECTION", "ETHERTYPE", "PROTOCOL", "PORT RANGE", "REMOTE"}
	var rows [][]string
	for _, g := range groups {
		for _, r := range g.Rules {
			rows = append(rows, []string{
				g.Name, r.Direction, r.EtherType, r.Protocol, portRange(r.PortRangeMin, r.PortRangeMax), dashIfEmpty(r.Remote),
			})
		}
	}
	return renderTableOrError(stdout, stderr, *format, headers, rows)
}

func runNetworkDescFloatingIP(args []string, stdout, stderr io.Writer) int {
	fs, long, format, configPath := networkFlags("network desc floating-ip")
	if err := fs.Parse(args); err != nil {
		return helpOrError(err, stdout, stderr, networkDescFloatingIPHelp)
	}
	if !output.Valid(*format) {
		fmt.Fprintf(stderr, "rescheck network desc floating-ip: 알 수 없는 --output %q\n", *format)
		return 1
	}

	clients, err := loadClients(*configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fips, err := nhncloud.ListFloatingIPs(context.Background(), clients.Compute, clients.Network)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if *format == output.FormatJSON {
		return renderJSONOrError(stdout, stderr, fips)
	}

	var headers []string
	if *long {
		headers = []string{"ID", "FLOATING IP", "FIXED IP", "STATUS", "ATTACHED INSTANCE", "FLOATING NETWORK", "PORT ID", "CREATED"}
	} else {
		headers = []string{"FLOATING IP", "FIXED IP", "STATUS", "ATTACHED INSTANCE"}
	}
	rows := make([][]string, 0, len(fips))
	for _, f := range fips {
		attached := dashIfEmpty(f.AttachedInstance)
		if *long {
			rows = append(rows, []string{
				f.ID, f.FloatingIP, dashIfEmpty(f.FixedIP), f.Status, attached,
				f.FloatingNetworkID, dashIfEmpty(f.PortID), formatTime(f.Created),
			})
		} else {
			rows = append(rows, []string{f.FloatingIP, dashIfEmpty(f.FixedIP), f.Status, attached})
		}
	}
	return renderTableOrError(stdout, stderr, *format, headers, rows)
}

func portRange(min, max int) string {
	if min == 0 && max == 0 {
		return "-"
	}
	if min == max {
		return strconv.Itoa(min)
	}
	return fmt.Sprintf("%d-%d", min, max)
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func renderTableOrError(stdout, stderr io.Writer, format string, headers []string, rows [][]string) int {
	if err := output.Render(stdout, format, headers, rows); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func renderJSONOrError(stdout, stderr io.Writer, v any) int {
	if err := output.RenderJSON(stdout, v); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

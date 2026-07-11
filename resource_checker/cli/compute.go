package cli

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/opp-13/nhn_iac/resource_checker/config"
	"github.com/opp-13/nhn_iac/resource_checker/nhncloud"
	"github.com/opp-13/nhn_iac/resource_checker/output"
)

func runCompute(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, computeHelp)
		return 0
	}

	switch args[0] {
	case "-h", "--help":
		fmt.Fprint(stdout, computeHelp)
		return 0
	case "desc", "describe":
		return runComputeDesc(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "rescheck compute: 알 수 없는 명령어 %q\n\n", args[0])
		fmt.Fprint(stderr, computeHelp)
		return 1
	}
}

func runComputeDesc(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, computeDescHelp)
		return 0
	}

	switch args[0] {
	case "-h", "--help":
		fmt.Fprint(stdout, computeDescHelp)
		return 0
	case "instance":
		return runComputeDescInstance(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "rescheck compute desc: 알 수 없는 리소스 %q\n\n", args[0])
		fmt.Fprint(stderr, computeDescHelp)
		return 1
	}
}

func runComputeDescInstance(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("compute desc instance", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var all, long bool
	var format, configPath string
	fs.BoolVar(&all, "a", false, "")
	fs.BoolVar(&all, "all", false, "")
	fs.BoolVar(&long, "l", false, "")
	fs.BoolVar(&long, "long", false, "")
	fs.StringVar(&format, "o", output.FormatTable, "")
	fs.StringVar(&format, "output", output.FormatTable, "")
	fs.StringVar(&configPath, "config", config.DefaultPath, "")

	if err := fs.Parse(expandShortBoolFlags(args, "al")); err != nil {
		return helpOrError(err, stdout, stderr, computeDescInstanceHelp)
	}
	if !output.Valid(format) {
		fmt.Fprintf(stderr, "rescheck compute desc instance: 알 수 없는 --output %q\n", format)
		return 1
	}

	cfg, _, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	ctx := context.Background()
	clients, err := nhncloud.NewClients(ctx, cfg)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	instances, err := nhncloud.ListInstances(ctx, clients.Compute, all)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	if format == output.FormatJSON {
		if err := output.RenderJSON(stdout, instances); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}

	headers, rows := instanceTable(instances, long)
	if err := output.Render(stdout, format, headers, rows); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func instanceTable(instances []nhncloud.Instance, long bool) (headers []string, rows [][]string) {
	if long {
		headers = []string{"ID", "NAME", "STATUS", "FLAVOR", "IMAGE", "PRIVATE IP", "FLOATING IP", "SECURITY GROUPS", "AVAILABILITY ZONE", "CREATED"}
	} else {
		headers = []string{"NAME", "STATUS", "PRIVATE IP", "FLOATING IP", "SECURITY GROUPS"}
	}

	for _, inst := range instances {
		privateIP := nhncloud.JoinOrDash(inst.PrivateIPs)
		floatingIP := nhncloud.JoinOrDash(inst.FloatingIPs)
		securityGroups := nhncloud.JoinOrDash(inst.SecurityGroups)

		if long {
			rows = append(rows, []string{
				inst.ID, inst.Name, inst.Status, inst.FlavorID, inst.ImageID,
				privateIP, floatingIP, securityGroups, inst.AvailabilityZone,
				formatTime(inst.Created),
			})
		} else {
			rows = append(rows, []string{
				inst.Name, inst.Status, privateIP, floatingIP, securityGroups,
			})
		}
	}
	return headers, rows
}

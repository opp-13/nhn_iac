package cli

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/gophercloud/gophercloud/v2"

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
	case "run", "start":
		return runComputePower(args[1:], powerStart, stdout, stderr)
	case "shutdown", "stop":
		return runComputePower(args[1:], powerStop, stdout, stderr)
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

// powerAction describes the two mutating instance operations rescheck
// supports (see .claude/rules/resource-checker.md for the carve-out).
type powerAction struct {
	name       string // subcommand name, for messages ("run" / "shutdown")
	verb       string // confirmation phrasing ("started" / "stopped")
	skipStatus string // instances already in this state are skipped
	help       string
	do         func(ctx context.Context, client *gophercloud.ServiceClient, id string) error
}

var (
	powerStart = powerAction{
		name:       "run",
		verb:       "started",
		skipStatus: "ACTIVE",
		help:       computeRunHelp,
		do:         nhncloud.StartInstance,
	}
	powerStop = powerAction{
		name:       "shutdown",
		verb:       "stopped",
		skipStatus: "SHUTOFF",
		help:       computeShutdownHelp,
		do:         nhncloud.StopInstance,
	}
)

// matchInstances resolves target against instances: a shell glob (contains
// * ? or [) matches instance names via path.Match; anything else must
// exactly match one instance's name or ID. Duplicate-name exact matches are
// an error so the wrong instance is never acted on.
func matchInstances(instances []nhncloud.Instance, target string) ([]nhncloud.Instance, error) {
	if strings.ContainsAny(target, "*?[") {
		var matched []nhncloud.Instance
		for _, inst := range instances {
			ok, err := path.Match(target, inst.Name)
			if err != nil {
				return nil, fmt.Errorf("잘못된 패턴 %q: %w", target, err)
			}
			if ok {
				matched = append(matched, inst)
			}
		}
		if len(matched) == 0 {
			return nil, fmt.Errorf("패턴 %q와 일치하는 인스턴스가 없습니다", target)
		}
		return matched, nil
	}

	var byName []nhncloud.Instance
	for _, inst := range instances {
		if inst.ID == target {
			return []nhncloud.Instance{inst}, nil
		}
		if inst.Name == target {
			byName = append(byName, inst)
		}
	}
	switch len(byName) {
	case 0:
		return nil, fmt.Errorf("이름 또는 ID가 %q인 인스턴스가 없습니다", target)
	case 1:
		return byName, nil
	default:
		ids := make([]string, 0, len(byName))
		for _, inst := range byName {
			ids = append(ids, inst.ID)
		}
		return nil, fmt.Errorf("이름 %q인 인스턴스가 %d개 있습니다 — ID로 지정하세요: %s",
			target, len(byName), strings.Join(ids, ", "))
	}
}

func runComputePower(args []string, action powerAction, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("compute "+action.name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var yes bool
	var configPath string
	fs.BoolVar(&yes, "y", false, "")
	fs.BoolVar(&yes, "yes", false, "")
	fs.StringVar(&configPath, "config", config.DefaultPath, "")

	// The stdlib flag package stops at the first positional argument, but
	// "rescheck compute shutdown 'test*' -y" (flag after the target) should
	// work too — re-parse the remainder after collecting each positional.
	var positional []string
	rest := args
	for {
		if err := fs.Parse(rest); err != nil {
			return helpOrError(err, stdout, stderr, action.help)
		}
		if fs.NArg() == 0 {
			break
		}
		positional = append(positional, fs.Arg(0))
		rest = fs.Args()[1:]
	}
	if len(positional) != 1 {
		fmt.Fprintf(stderr, "rescheck compute %s: 대상(NAME|ID|GLOB)을 하나 지정하세요\n\n", action.name)
		fmt.Fprint(stderr, action.help)
		return 1
	}
	target := positional[0]

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

	instances, err := nhncloud.ListInstances(ctx, clients.Compute, true)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	matched, err := matchInstances(instances, target)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	var targets, skipped []nhncloud.Instance
	for _, inst := range matched {
		if inst.Status == action.skipStatus {
			skipped = append(skipped, inst)
		} else {
			targets = append(targets, inst)
		}
	}
	for _, inst := range skipped {
		fmt.Fprintf(stdout, "skipped %s (already %s)\n", inst.Name, inst.Status)
	}
	if len(targets) == 0 {
		fmt.Fprintln(stdout, "nothing to do")
		return 0
	}

	// Only a glob can silently widen the blast radius, so only a glob asks.
	if strings.ContainsAny(target, "*?[") && !yes {
		fmt.Fprintf(stdout, "The following %d instances will be %s:\n", len(targets), action.verb)
		for _, inst := range targets {
			fmt.Fprintf(stdout, "  %s\t%s\n", inst.Name, inst.Status)
		}
		answer, err := promptLine(stdout, bufio.NewReader(os.Stdin), "Proceed? [y/N]: ")
		if err != nil {
			fmt.Fprintf(stderr, "rescheck compute %s: 확인 입력을 받을 수 없습니다 (-y로 확인을 생략할 수 있습니다): %v\n", action.name, err)
			return 1
		}
		switch strings.ToLower(answer) {
		case "y", "yes":
		default:
			fmt.Fprintln(stdout, "aborted")
			return 0
		}
	}

	code := 0
	for _, inst := range targets {
		if err := action.do(ctx, clients.Compute, inst.ID); err != nil {
			fmt.Fprintln(stderr, err)
			code = 1
			continue
		}
		fmt.Fprintf(stdout, "%s %s\n", action.verb, inst.Name)
	}
	return code
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

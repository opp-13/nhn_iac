// Package cli implements instsched's command-line dispatch: "check" (the
// one-shot command cron invokes per instance) and "schedule" (manage the
// crontab entries that call it).
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/opp-13/nhn_iac/instance_scheduler/config"
	"github.com/opp-13/nhn_iac/instance_scheduler/guard"
	"github.com/opp-13/nhn_iac/instance_scheduler/schedule"
)

// Run dispatches args (os.Args[1:]) and returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, rootHelp)
		return 0
	}

	switch args[0] {
	case "-h", "--help", "help":
		fmt.Fprint(stdout, rootHelp)
		return 0
	case "-v", "--version":
		fmt.Fprintf(stdout, "instsched %s\n", version)
		return 0
	case "check":
		return runCheck(args[1:], stdout, stderr)
	case "schedule":
		return runSchedule(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "instsched: 알 수 없는 명령어 %q\n\n", args[0])
		fmt.Fprint(stderr, rootHelp)
		return 1
	}
}

// helpOrError prints helpText to stdout and returns 0 if err is
// flag.ErrHelp (the user passed -h/--help), otherwise prints err and
// helpText to stderr and returns 1.
func helpOrError(err error, stdout, stderr io.Writer, helpText string) int {
	if errors.Is(err, flag.ErrHelp) {
		fmt.Fprint(stdout, helpText)
		return 0
	}
	fmt.Fprintln(stderr, err)
	fmt.Fprint(stderr, helpText)
	return 1
}

func runCheck(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var configPath, instance string
	fs.StringVar(&configPath, "config", config.FindConfigPath(), "")
	fs.StringVar(&instance, "instance", "", "")

	if err := fs.Parse(args); err != nil {
		return helpOrError(err, stdout, stderr, checkHelp)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if !cfg.Nhn.Instancescheduler.Enabled {
		fmt.Fprintln(stdout, "instsched: 비활성화됨 (nhn.Instancescheduler.enabled: false)")
		return 0
	}
	if len(cfg.Nhn.Instancescheduler.Instances) == 0 {
		fmt.Fprintln(stdout, "instsched: 설정된 인스턴스가 없습니다 (nhn.Instancescheduler.instances)")
		return 0
	}

	result, err := guard.Check(context.Background(), cfg, configPath, instance, nil)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	code := 0
	for _, ir := range result.Instances {
		switch ir.Action {
		case guard.ActionOK:
			fmt.Fprintf(stdout, "%s: OK\n", ir.Name)
		case guard.ActionStarted:
			fmt.Fprintf(stdout, "%s: STOPPED -> started\n", ir.Name)
		case guard.ActionRecreated:
			fmt.Fprintf(stdout, "%s: DELETED -> terraform apply 완료\n", ir.Name)
		case guard.ActionSkippedStatus:
			fmt.Fprintf(stdout, "%s: %s (자동 조치 대상 아님)\n", ir.Name, ir.Detail)
		case guard.ActionSkippedDeleted:
			fmt.Fprintf(stdout, "%s: DELETED (terraform 미설정 — 복구하지 않음)\n", ir.Name)
		case guard.ActionError:
			fmt.Fprintf(stderr, "%s: 오류 - %s\n", ir.Name, ir.Detail)
			code = 1
		}
	}
	return code
}

func runSchedule(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, scheduleHelp)
		return 0
	}

	switch args[0] {
	case "-h", "--help":
		fmt.Fprint(stdout, scheduleHelp)
		return 0
	case "add":
		return runScheduleAdd(args[1:], stdout, stderr)
	case "remove":
		return runScheduleRemove(args[1:], stdout, stderr)
	case "list":
		return runScheduleList(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "instsched schedule: 알 수 없는 명령어 %q\n\n", args[0])
		fmt.Fprint(stderr, scheduleHelp)
		return 1
	}
}

func runScheduleAdd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("schedule add", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var configPath string
	fs.StringVar(&configPath, "config", config.FindConfigPath(), "")
	if err := fs.Parse(args); err != nil {
		return helpOrError(err, stdout, stderr, scheduleAddHelp)
	}
	if fs.NArg() != 1 {
		return helpOrError(fmt.Errorf("instsched schedule add: 인스턴스 이름을 하나 지정하세요"), stdout, stderr, scheduleAddHelp)
	}
	name := fs.Arg(0)

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := schedule.Add(context.Background(), nil, cfg, configPath, name); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "%s: crontab에 등록했습니다\n", name)
	return 0
}

func runScheduleRemove(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("schedule remove", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return helpOrError(err, stdout, stderr, scheduleRemoveHelp)
	}
	if fs.NArg() != 1 {
		return helpOrError(fmt.Errorf("instsched schedule remove: 인스턴스 이름을 하나 지정하세요"), stdout, stderr, scheduleRemoveHelp)
	}
	name := fs.Arg(0)

	if err := schedule.Remove(context.Background(), nil, name); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "%s: crontab에서 제거했습니다\n", name)
	return 0
}

func runScheduleList(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("schedule list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return helpOrError(err, stdout, stderr, scheduleListHelp)
	}

	entries, err := schedule.List(context.Background(), nil)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if len(entries) == 0 {
		fmt.Fprintln(stdout, "등록된 스케줄이 없습니다")
		return 0
	}
	for _, e := range entries {
		fmt.Fprintf(stdout, "%s\t%s\n", e.Name, e.Schedule)
	}
	return 0
}

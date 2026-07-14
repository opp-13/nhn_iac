// Package schedule manages the current user's crontab entries that trigger
// `instsched check --instance <name>` at each instance's configured
// schedule. Entries are tagged with a marker comment so they can be found,
// replaced, or removed without disturbing any other crontab lines the user
// manages themselves.
package schedule

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/opp-13/nhn_iac/instance_scheduler/config"
)

// markerPrefix tags a comment line immediately above a crontab line this
// package manages, so Add/Remove/List can find their own entries among
// whatever else is in the user's crontab.
const markerPrefix = "# instsched:managed:"

// Runner executes name with args and returns combined stdout+stderr. Tests
// inject a fake Runner instead of shelling out to the real "crontab" binary.
type Runner func(ctx context.Context, stdin string, name string, args ...string) (output []byte, err error)

func defaultRunner(ctx context.Context, stdin string, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		var execErr *exec.Error
		if errors.As(err, &execErr) {
			return nil, fmt.Errorf("%s 바이너리를 찾을 수 없습니다 (PATH 확인): %w", name, err)
		}
		return nil, fmt.Errorf("%s %s 실행 실패: %s (권한 문제일 수 있습니다 — sudo로 다시 시도해 보세요)",
			name, strings.Join(args, " "), strings.TrimSpace(out.String()))
	}
	return out.Bytes(), nil
}

// Entry is one managed crontab line.
type Entry struct {
	Name     string
	Schedule string
	Command  string
}

// readCrontab returns the current user's crontab lines. "no crontab" is not
// an error — an empty/nonexistent crontab is treated as zero lines.
func readCrontab(ctx context.Context, run Runner) ([]string, error) {
	out, err := run(ctx, "", "crontab", "-l")
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no crontab") {
			return nil, nil
		}
		return nil, err
	}
	text := strings.TrimRight(string(out), "\n")
	if text == "" {
		return nil, nil
	}
	return strings.Split(text, "\n"), nil
}

func writeCrontab(ctx context.Context, run Runner, lines []string) error {
	content := strings.Join(lines, "\n")
	if content != "" {
		content += "\n"
	}
	_, err := run(ctx, content, "crontab", "-")
	return err
}

// stripManaged removes any existing marker+command pair for name from
// lines, returning the remaining lines.
func stripManaged(lines []string, name string) []string {
	marker := markerPrefix + name
	kept := make([]string, 0, len(lines))
	for i := 0; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == marker {
			i++ // also drop the command line right after the marker
			continue
		}
		kept = append(kept, lines[i])
	}
	return kept
}

// Add installs (or replaces) the crontab entry for the instance named name,
// using its configured schedule. The generated command uses absolute paths
// for both the instsched binary and configPath, since cron runs jobs with a
// minimal environment that can't be relied on to resolve relative paths or
// a bare "instsched" via PATH.
func Add(ctx context.Context, run Runner, cfg *config.Config, configPath, name string) error {
	if run == nil {
		run = defaultRunner
	}

	inst, ok := cfg.FindInstance(name)
	if !ok {
		return fmt.Errorf("설정에 없는 인스턴스입니다: %q", name)
	}
	if inst.Schedule == "" {
		return fmt.Errorf("인스턴스 %q에 schedule이 설정되어 있지 않습니다", name)
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("실행 파일 경로를 확인할 수 없습니다: %w", err)
	}
	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("config 경로를 확인할 수 없습니다: %w", err)
	}

	lines, err := readCrontab(ctx, run)
	if err != nil {
		return err
	}
	lines = stripManaged(lines, name)
	lines = append(lines, markerPrefix+name,
		fmt.Sprintf("%s %s check --instance %s --config %s", inst.Schedule, exe, name, absConfigPath))

	return writeCrontab(ctx, run, lines)
}

// Remove uninstalls the crontab entry for the instance named name, if any.
// It is not an error for no entry to exist.
func Remove(ctx context.Context, run Runner, name string) error {
	if run == nil {
		run = defaultRunner
	}

	lines, err := readCrontab(ctx, run)
	if err != nil {
		return err
	}
	stripped := stripManaged(lines, name)
	if len(stripped) == len(lines) {
		return nil // nothing registered for this instance
	}
	return writeCrontab(ctx, run, stripped)
}

// List returns every instsched-managed crontab entry.
func List(ctx context.Context, run Runner) ([]Entry, error) {
	if run == nil {
		run = defaultRunner
	}

	lines, err := readCrontab(ctx, run)
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for i := 0; i < len(lines); i++ {
		marker := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(marker, markerPrefix) || i+1 >= len(lines) {
			continue
		}
		name := strings.TrimPrefix(marker, markerPrefix)
		command := lines[i+1]
		fields := strings.SplitN(command, " ", 6)
		schedule := command
		if len(fields) >= 5 {
			schedule = strings.Join(fields[:5], " ")
		}
		entries = append(entries, Entry{Name: name, Schedule: schedule, Command: command})
		i++
	}
	return entries, nil
}

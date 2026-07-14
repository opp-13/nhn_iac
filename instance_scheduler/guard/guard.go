// Package guard implements instsched's check-and-heal pass: for each
// scheduled instance it asks the rescheck binary (resource_checker's CLI,
// via subprocess) whether the instance is ACTIVE, SHUTOFF, or missing, then
// starts it (rescheck compute run) or recreates it (terraform apply, only
// when the instance opts into that) as needed. instance_scheduler is an
// independent Go module and deliberately does not import resource_checker's
// packages — every cross-module call goes through a subprocess boundary
// instead.
package guard

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/opp-13/nhn_iac/instance_scheduler/config"
)

// identityEndpoint is NHN Cloud's OpenStack Identity v2 token endpoint,
// needed so `terraform apply` can authenticate via OS_AUTH_URL. This is the
// same value resource_checker/nhncloud hardcodes; it's duplicated here
// (rather than imported) because the two modules must build independently.
const identityEndpoint = "https://api-identity-infrastructure.nhncloudservice.com/v2.0"

// Runner executes name with args, optionally within dir and with extra
// environment variables appended to the current process's environment, and
// returns combined stdout+stderr. Tests inject a fake Runner instead of
// shelling out to real "rescheck"/"terraform" binaries.
type Runner func(ctx context.Context, dir string, env []string, name string, args ...string) (output []byte, err error)

func defaultRunner(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		var execErr *exec.Error
		if errors.As(err, &execErr) {
			return nil, fmt.Errorf("%s 바이너리를 찾을 수 없습니다 (PATH 확인): %w", name, err)
		}
		return nil, fmt.Errorf("%s %s 실행 실패: %s", name, strings.Join(args, " "), strings.TrimSpace(out.String()))
	}
	return out.Bytes(), nil
}

type instanceJSON struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// Action is the outcome of checking one scheduled instance.
type Action string

const (
	ActionOK            Action = "OK"
	ActionStarted       Action = "STARTED"
	ActionRecreated     Action = "RECREATED"
	ActionSkippedStatus Action = "SKIPPED"
	// ActionSkippedDeleted marks an instance that's missing from the cloud
	// but isn't configured for terraform recovery (Instance.Terraform ==
	// false) — reported only, never recreated.
	ActionSkippedDeleted Action = "SKIPPED_DELETED"
	ActionError          Action = "ERROR"
)

// InstanceResult is the outcome for one scheduled instance.
type InstanceResult struct {
	Name   string
	Action Action
	// Detail carries the cloud status for ActionSkippedStatus, or the error
	// message for ActionError. Empty otherwise.
	Detail string
}

type Result struct {
	Instances []InstanceResult
}

// HasErrors reports whether any instance ended in ActionError.
func (r Result) HasErrors() bool {
	for _, ir := range r.Instances {
		if ir.Action == ActionError {
			return true
		}
	}
	return false
}

// Check runs one check-and-heal pass. If targetName is non-empty, only that
// instance (which must be present in cfg) is checked — this is the mode
// cron uses, since each instance runs on its own schedule. If targetName is
// empty, every configured instance is checked in one pass (manual/testing
// use). configPath is forwarded to the rescheck subprocess (--config) so it
// reads the same config.yaml regardless of the caller's working directory.
// run is the subprocess runner; pass nil to use the real one.
func Check(ctx context.Context, cfg *config.Config, configPath, targetName string, run Runner) (Result, error) {
	if run == nil {
		run = defaultRunner
	}

	targets := cfg.Nhn.Instancescheduler.Instances
	if targetName != "" {
		inst, ok := cfg.FindInstance(targetName)
		if !ok {
			return Result{}, fmt.Errorf("설정에 없는 인스턴스입니다: %q", targetName)
		}
		targets = []config.Instance{inst}
	}

	instances, err := listInstances(ctx, run, configPath)
	if err != nil {
		return Result{}, err
	}
	byName := make(map[string]instanceJSON, len(instances))
	for _, inst := range instances {
		byName[inst.Name] = inst
	}

	var results []InstanceResult
	var missing []config.Instance
	for _, target := range targets {
		inst, ok := byName[target.Name]
		switch {
		case !ok:
			missing = append(missing, target)
		case inst.Status == "SHUTOFF":
			if _, err := startInstance(ctx, run, configPath, target.Name); err != nil {
				results = append(results, InstanceResult{Name: target.Name, Action: ActionError, Detail: err.Error()})
			} else {
				results = append(results, InstanceResult{Name: target.Name, Action: ActionStarted})
			}
		case inst.Status == "ACTIVE":
			results = append(results, InstanceResult{Name: target.Name, Action: ActionOK})
		default:
			results = append(results, InstanceResult{Name: target.Name, Action: ActionSkippedStatus, Detail: inst.Status})
		}
	}

	var toRecreate []config.Instance
	for _, inst := range missing {
		if inst.Terraform {
			toRecreate = append(toRecreate, inst)
		} else {
			results = append(results, InstanceResult{Name: inst.Name, Action: ActionSkippedDeleted})
		}
	}

	if len(toRecreate) > 0 {
		_, applyErr := applyTerraform(ctx, run, cfg.Nhn.Instancescheduler.TerraformDir, &cfg.Nhn.Auth)
		for _, inst := range toRecreate {
			if applyErr != nil {
				results = append(results, InstanceResult{Name: inst.Name, Action: ActionError, Detail: applyErr.Error()})
			} else {
				results = append(results, InstanceResult{Name: inst.Name, Action: ActionRecreated})
			}
		}
	}

	return Result{Instances: results}, nil
}

func listInstances(ctx context.Context, run Runner, configPath string) ([]instanceJSON, error) {
	out, err := run(ctx, "", nil, "rescheck", "compute", "desc", "instance", "-a", "-o", "json", "--config", configPath)
	if err != nil {
		return nil, fmt.Errorf("인스턴스 상태 조회 실패: %w", err)
	}
	var instances []instanceJSON
	if err := json.Unmarshal(out, &instances); err != nil {
		return nil, fmt.Errorf("rescheck 출력 파싱 실패: %w", err)
	}
	return instances, nil
}

func startInstance(ctx context.Context, run Runner, configPath, name string) ([]byte, error) {
	return run(ctx, "", nil, "rescheck", "compute", "run", name, "--config", configPath)
}

func applyTerraform(ctx context.Context, run Runner, dir string, auth *config.Auth) ([]byte, error) {
	env := []string{
		"OS_AUTH_URL=" + identityEndpoint,
		"OS_USERNAME=" + auth.Username,
		"OS_PASSWORD=" + auth.Password,
		"OS_TENANT_ID=" + auth.TenantID,
		"OS_REGION_NAME=" + auth.Region,
		"OS_IDENTITY_API_VERSION=2",
	}
	return run(ctx, dir, env, "terraform", "apply", "-auto-approve")
}

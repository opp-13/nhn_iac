package guard

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/opp-13/nhn_iac/instance_scheduler/config"
)

func testConfig(instances ...config.Instance) *config.Config {
	cfg := &config.Config{}
	cfg.Nhn.Auth = config.Auth{TenantID: "tenant", Region: "KR1", Username: "user", Password: "pass"}
	cfg.Nhn.Instancescheduler = config.Instancescheduler{
		Enabled:      true,
		TerraformDir: "./terraform",
		Instances:    instances,
	}
	return cfg
}

func inst(name string, terraform bool) config.Instance {
	return config.Instance{Name: name, Schedule: "0 9 * * *", Terraform: terraform}
}

func TestCheck_ActiveInstanceIsNoOp(t *testing.T) {
	cfg := testConfig(inst("web-01", true))
	var calls []string
	run := func(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
		calls = append(calls, strings.Join(append([]string{name}, args...), " "))
		return []byte(`[{"name":"web-01","status":"ACTIVE"}]`), nil
	}

	result, err := Check(context.Background(), cfg, "config.yaml", "", run)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(result.Instances) != 1 || result.Instances[0].Action != ActionOK {
		t.Fatalf("expected single ActionOK, got %+v", result.Instances)
	}
	if len(calls) != 1 {
		t.Fatalf("expected only the list call, got %v", calls)
	}
}

func TestCheck_StoppedInstanceIsStarted(t *testing.T) {
	cfg := testConfig(inst("web-01", true))
	var startCalled bool
	run := func(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
		if name == "rescheck" && len(args) > 0 && args[0] == "compute" && args[1] == "run" {
			startCalled = true
			if args[2] != "web-01" {
				t.Fatalf("expected start target web-01, got %v", args)
			}
			return nil, nil
		}
		return []byte(`[{"name":"web-01","status":"SHUTOFF"}]`), nil
	}

	result, err := Check(context.Background(), cfg, "config.yaml", "", run)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !startCalled {
		t.Fatal("expected rescheck compute run to be called")
	}
	if len(result.Instances) != 1 || result.Instances[0].Action != ActionStarted {
		t.Fatalf("expected ActionStarted, got %+v", result.Instances)
	}
}

func TestCheck_MissingInstanceWithTerraformTriggersApply(t *testing.T) {
	cfg := testConfig(inst("web-01", true), inst("db-01", true))
	var applyCalled bool
	var applyEnv []string
	run := func(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
		if name == "terraform" {
			applyCalled = true
			applyEnv = env
			if dir != "./terraform" {
				t.Fatalf("expected terraform dir ./terraform, got %q", dir)
			}
			return nil, nil
		}
		return []byte(`[{"name":"web-01","status":"ACTIVE"}]`), nil
	}

	result, err := Check(context.Background(), cfg, "config.yaml", "", run)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if !applyCalled {
		t.Fatal("expected terraform apply to be called for the missing instance")
	}
	hasAuthEnv := false
	for _, e := range applyEnv {
		if e == "OS_USERNAME=user" {
			hasAuthEnv = true
		}
	}
	if !hasAuthEnv {
		t.Fatalf("expected OS_USERNAME env var to be set, got %v", applyEnv)
	}

	var dbResult *InstanceResult
	for i := range result.Instances {
		if result.Instances[i].Name == "db-01" {
			dbResult = &result.Instances[i]
		}
	}
	if dbResult == nil || dbResult.Action != ActionRecreated {
		t.Fatalf("expected db-01 to be ActionRecreated, got %+v", result.Instances)
	}
}

func TestCheck_MissingInstanceWithoutTerraformIsSkipped(t *testing.T) {
	cfg := testConfig(inst("db-01", false))
	var applyCalled bool
	run := func(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
		if name == "terraform" {
			applyCalled = true
			return nil, nil
		}
		return []byte(`[]`), nil
	}

	result, err := Check(context.Background(), cfg, "config.yaml", "", run)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if applyCalled {
		t.Fatal("expected terraform apply NOT to be called for an instance without terraform: true")
	}
	if len(result.Instances) != 1 || result.Instances[0].Action != ActionSkippedDeleted {
		t.Fatalf("expected ActionSkippedDeleted, got %+v", result.Instances)
	}
}

func TestCheck_TerraformApplyFailureIsReportedAsError(t *testing.T) {
	cfg := testConfig(inst("db-01", true))
	run := func(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
		if name == "terraform" {
			return nil, errors.New("apply failed: quota exceeded")
		}
		return []byte(`[]`), nil
	}

	result, err := Check(context.Background(), cfg, "config.yaml", "", run)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(result.Instances) != 1 || result.Instances[0].Action != ActionError {
		t.Fatalf("expected ActionError, got %+v", result.Instances)
	}
	if !result.HasErrors() {
		t.Fatal("expected HasErrors() to be true")
	}
}

func TestCheck_UnknownStatusIsSkipped(t *testing.T) {
	cfg := testConfig(inst("web-01", true))
	run := func(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
		return []byte(`[{"name":"web-01","status":"ERROR"}]`), nil
	}

	result, err := Check(context.Background(), cfg, "config.yaml", "", run)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(result.Instances) != 1 || result.Instances[0].Action != ActionSkippedStatus || result.Instances[0].Detail != "ERROR" {
		t.Fatalf("expected ActionSkippedStatus with detail ERROR, got %+v", result.Instances)
	}
}

func TestCheck_TargetNameChecksOnlyThatInstance(t *testing.T) {
	cfg := testConfig(inst("web-01", true), inst("db-01", true))
	run := func(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
		return []byte(`[{"name":"web-01","status":"ACTIVE"},{"name":"db-01","status":"ACTIVE"}]`), nil
	}

	result, err := Check(context.Background(), cfg, "config.yaml", "db-01", run)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(result.Instances) != 1 || result.Instances[0].Name != "db-01" {
		t.Fatalf("expected only db-01 to be checked, got %+v", result.Instances)
	}
}

func TestCheck_UnknownTargetNameIsError(t *testing.T) {
	cfg := testConfig(inst("web-01", true))
	run := func(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
		t.Fatal("run should not be called when the target name is invalid")
		return nil, nil
	}

	if _, err := Check(context.Background(), cfg, "config.yaml", "does-not-exist", run); err == nil {
		t.Fatal("expected an error for an unconfigured instance name")
	}
}

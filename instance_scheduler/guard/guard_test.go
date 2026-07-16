package guard

import (
	"context"
	"errors"
	"path/filepath"
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
	return config.Instance{Name: name, StartSchedule: "0 9 * * *", Terraform: terraform}
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
	wantDir := filepath.Join(".", "terraform") // resolveTerraformDir(".../terraform", "config.yaml")
	run := func(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
		if name == "terraform" {
			applyCalled = true
			applyEnv = env
			if dir != wantDir {
				t.Fatalf("expected terraform dir %q, got %q", wantDir, dir)
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

func TestStop_ActiveInstanceIsStopped(t *testing.T) {
	cfg := testConfig(inst("web-01", true))
	var stopCalled bool
	run := func(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
		if name == "rescheck" && len(args) > 0 && args[0] == "compute" && args[1] == "shutdown" {
			stopCalled = true
			if args[2] != "web-01" {
				t.Fatalf("expected stop target web-01, got %v", args)
			}
			return nil, nil
		}
		return []byte(`[{"name":"web-01","status":"ACTIVE"}]`), nil
	}

	result, err := Stop(context.Background(), cfg, "config.yaml", "", run)
	if err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if !stopCalled {
		t.Fatal("expected rescheck compute shutdown to be called")
	}
	if len(result.Instances) != 1 || result.Instances[0].Action != ActionStopped {
		t.Fatalf("expected ActionStopped, got %+v", result.Instances)
	}
}

func TestStop_AlreadyStoppedIsOK(t *testing.T) {
	cfg := testConfig(inst("web-01", true))
	var stopCalled bool
	run := func(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
		if name == "rescheck" && len(args) > 0 && args[0] == "compute" && args[1] == "shutdown" {
			stopCalled = true
			return nil, nil
		}
		return []byte(`[{"name":"web-01","status":"SHUTOFF"}]`), nil
	}

	result, err := Stop(context.Background(), cfg, "config.yaml", "", run)
	if err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if stopCalled {
		t.Fatal("expected rescheck compute shutdown NOT to be called for an already-stopped instance")
	}
	if len(result.Instances) != 1 || result.Instances[0].Action != ActionOK {
		t.Fatalf("expected ActionOK, got %+v", result.Instances)
	}
}

func TestStop_DeletedInstanceHasNothingToStop(t *testing.T) {
	cfg := testConfig(inst("web-01", true))
	run := func(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
		return []byte(`[]`), nil
	}

	result, err := Stop(context.Background(), cfg, "config.yaml", "", run)
	if err != nil {
		t.Fatalf("Stop: %v", err)
	}
	if len(result.Instances) != 1 || result.Instances[0].Action != ActionSkippedDeleted {
		t.Fatalf("expected ActionSkippedDeleted, got %+v", result.Instances)
	}
}

func TestStop_UnknownTargetNameIsError(t *testing.T) {
	cfg := testConfig(inst("web-01", true))
	run := func(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
		t.Fatal("run should not be called when the target name is invalid")
		return nil, nil
	}

	if _, err := Stop(context.Background(), cfg, "config.yaml", "does-not-exist", run); err == nil {
		t.Fatal("expected an error for an unconfigured instance name")
	}
}

func TestResolveTerraformDir_RelativeIsAnchoredToConfigDir(t *testing.T) {
	got := resolveTerraformDir("instance_scheduler/terraform", "/home/user/.config/nhn_iac/config.yaml")
	want := filepath.Join("/home/user/.config/nhn_iac", "instance_scheduler/terraform")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolveTerraformDir_EmptyFallsBackToHomeDir(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome) // os.UserHomeDir reads this on Windows

	got := resolveTerraformDir("", "/home/user/.config/nhn_iac/config.yaml")
	want := filepath.Join(tmpHome, filepath.FromSlash(".config/nhn_iac/instance_scheduler/terraform"))
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolveTerraformDir_AbsoluteIsUnchanged(t *testing.T) {
	// filepath.Abs guarantees an OS-legitimate absolute path (on Windows,
	// filepath.IsAbs requires a volume — a bare "/opt/..." string isn't
	// absolute there), so build "abs" that way rather than hardcoding one.
	abs, err := filepath.Abs(filepath.Join("opt", "nhn_iac", "terraform"))
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	got := resolveTerraformDir(abs, "/home/user/.config/nhn_iac/config.yaml")
	if got != abs {
		t.Fatalf("got %q, want unchanged %q", got, abs)
	}
}

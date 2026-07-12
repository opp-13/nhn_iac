package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opp-13/nhn_iac/resource_checker/config"
)

func TestRunConfigureSetUsesFlatYAMLSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	var stdout, stderr bytes.Buffer
	code := runConfigureSet([]string{
		"--tenant-id", "t1", "--username", "u1", "--password", "p1", "--config", path,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("runConfigureSet() code = %d, stderr = %s", code, stderr.String())
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(data), "passwordCredentials") {
		t.Errorf("config.yaml should use the flat nhn.auth.{username,password} schema, got:\n%s", data)
	}
}

func TestRunConfigureSetPreservesRegionOnCredentialOnlyResave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	var stdout, stderr bytes.Buffer
	code := runConfigureSet([]string{
		"--tenant-id", "t1", "--username", "u1", "--password", "p1",
		"--region", "KR2", "--config", path,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("first runConfigureSet() code = %d, stderr = %s", code, stderr.String())
	}

	cfg, _, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Nhn.Auth.Region != "KR2" {
		t.Fatalf("Region after first save = %q, want %q", cfg.Nhn.Auth.Region, "KR2")
	}

	stdout.Reset()
	stderr.Reset()
	code = runConfigureSet([]string{
		"--tenant-id", "t1", "--username", "u1", "--password", "p2", "--config", path,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("second runConfigureSet() code = %d, stderr = %s", code, stderr.String())
	}

	cfg, _, err = config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Nhn.Auth.Region != "KR2" {
		t.Errorf("Region after credential-only resave = %q, want preserved %q", cfg.Nhn.Auth.Region, "KR2")
	}
	if cfg.Nhn.Auth.Password != "p2" {
		t.Errorf("Password after resave = %q, want %q", cfg.Nhn.Auth.Password, "p2")
	}
}

package config

import (
	"path/filepath"
	"testing"
)

func TestLoadMissingFileDefaultsRegion(t *testing.T) {
	cfg, raw, err := Load(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg.Nhn.Auth.Region != DefaultRegion {
		t.Errorf("Region = %q, want %q", cfg.Nhn.Auth.Region, DefaultRegion)
	}
	if len(raw) != 0 {
		t.Errorf("raw = %v, want empty map", raw)
	}
}

func TestSaveAuthThenLoadRoundTrips(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	if err := SaveAuth(path, nil, "tenant-1", DefaultRegion, "jdoe", "s3cret"); err != nil {
		t.Fatalf("SaveAuth() error = %v", err)
	}

	cfg, _, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Nhn.Auth.TenantID != "tenant-1" {
		t.Errorf("TenantID = %q, want %q", cfg.Nhn.Auth.TenantID, "tenant-1")
	}
	if cfg.Nhn.Auth.Username != "jdoe" {
		t.Errorf("Username = %q, want %q", cfg.Nhn.Auth.Username, "jdoe")
	}
	if cfg.Nhn.Auth.Password != "s3cret" {
		t.Errorf("Password = %q, want %q", cfg.Nhn.Auth.Password, "s3cret")
	}
	if cfg.Nhn.Auth.Region != DefaultRegion {
		t.Errorf("Region = %q, want default %q", cfg.Nhn.Auth.Region, DefaultRegion)
	}
}

func TestSaveAuthPreservesUnrelatedKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	_, raw, err := Load(path) // start from a fresh raw map
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	raw["nhn"] = map[string]any{
		"Resourcechecker": map[string]any{"enabled": true, "mode": "cli"},
		"Autoremover":     map[string]any{"enabled": true, "mode": "cli"},
	}
	raw["unrelated_top_level_key"] = "keep-me"

	if err := SaveAuth(path, raw, "tenant-2", "KR2", "alice", "hunter2"); err != nil {
		t.Fatalf("SaveAuth() error = %v", err)
	}

	cfg, rawAfter, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Nhn.Auth.TenantID != "tenant-2" {
		t.Errorf("TenantID = %q, want %q", cfg.Nhn.Auth.TenantID, "tenant-2")
	}
	if cfg.Nhn.Auth.Region != "KR2" {
		t.Errorf("Region = %q, want %q", cfg.Nhn.Auth.Region, "KR2")
	}
	if !cfg.Nhn.Resourcechecker.Enabled || cfg.Nhn.Resourcechecker.Mode != "cli" {
		t.Errorf("Resourcechecker = %+v, want enabled=true mode=cli", cfg.Nhn.Resourcechecker)
	}
	if rawAfter["unrelated_top_level_key"] != "keep-me" {
		t.Errorf("unrelated_top_level_key = %v, want %q", rawAfter["unrelated_top_level_key"], "keep-me")
	}
	nhn, _ := rawAfter["nhn"].(map[string]any)
	if _, ok := nhn["Autoremover"]; !ok {
		t.Errorf("nhn.Autoremover was dropped by SaveAuth, raw = %v", rawAfter)
	}
}

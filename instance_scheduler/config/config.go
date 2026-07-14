// Package config loads instsched's view of the shared config.yaml, under
// the nhn.auth / nhn.Instancescheduler keys. This mirrors (but does not
// import) resource_checker/config — the two modules are independent and
// each reads the same file on disk without sharing Go code.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultPath is used when the caller doesn't override --config and
	// FindConfigPath doesn't find a config.yaml anywhere above the current
	// directory (e.g. a fresh setup with no config file yet).
	DefaultPath = "./config.yaml"
	// DefaultRegion is used when nhn.auth.region is absent from config.yaml.
	DefaultRegion = "KR1"
)

// FindConfigPath looks for a file named "config.yaml" starting in the
// current working directory and walking up through parent directories, so
// instsched finds the repo's shared config.yaml even when run from a
// subdirectory (e.g. `go run .` inside instance_scheduler/ itself, one
// level below where config.yaml actually lives). If none is found by the
// filesystem root, it falls back to DefaultPath so Load's "missing file"
// branch still applies for a genuinely fresh setup.
func FindConfigPath() string {
	dir, err := os.Getwd()
	if err != nil {
		return DefaultPath
	}
	for {
		candidate := filepath.Join(dir, "config.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return DefaultPath
		}
		dir = parent
	}
}

type Auth struct {
	TenantID string `yaml:"tenantId"`
	Region   string `yaml:"region"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Instance is one protected instance: schedule says when it must exist
// (a cron expression), and Terraform says whether a deleted instance may be
// recreated via `terraform apply` in TerraformDir. When Terraform is false,
// a deleted instance is only reported, never recreated.
type Instance struct {
	Name      string `yaml:"name"`
	Schedule  string `yaml:"schedule"`
	Terraform bool   `yaml:"terraform"`
}

type Instancescheduler struct {
	Enabled      bool       `yaml:"enabled"`
	Mode         string     `yaml:"mode"`
	TerraformDir string     `yaml:"terraformDir"`
	Instances    []Instance `yaml:"instances"`
}

type Config struct {
	Nhn struct {
		Auth              Auth              `yaml:"auth"`
		Instancescheduler Instancescheduler `yaml:"Instancescheduler"`
	} `yaml:"nhn"`
}

// Load reads the config file at path. A missing file is not an error: Load
// returns a zero Config with region defaulted (Instancescheduler.Enabled
// stays false, so callers treat that the same as an explicit opt-out).
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
	case os.IsNotExist(err):
		cfg := &Config{}
		cfg.Nhn.Auth.Region = DefaultRegion
		return cfg, nil
	default:
		return nil, fmt.Errorf("config 파일을 읽을 수 없음 (%s): %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config 파일 파싱 실패 (%s): %w", path, err)
	}
	if cfg.Nhn.Auth.Region == "" {
		cfg.Nhn.Auth.Region = DefaultRegion
	}
	return cfg, nil
}

// FindInstance returns the configured instance named name, or false if it
// isn't listed under nhn.Instancescheduler.instances.
func (c *Config) FindInstance(name string) (Instance, bool) {
	for _, inst := range c.Nhn.Instancescheduler.Instances {
		if inst.Name == name {
			return inst, true
		}
	}
	return Instance{}, false
}

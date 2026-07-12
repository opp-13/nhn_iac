// Package config loads and saves rescheck's config.yaml, under the
// nhn.auth / nhn.Resourcechecker keys shared with the rest of this repo.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	// DefaultPath is used when the caller doesn't override --config.
	DefaultPath = "./config.yaml"
	// DefaultRegion is used when nhn.auth.region is absent from config.yaml.
	DefaultRegion = "KR1"
)

type Auth struct {
	TenantID string `yaml:"tenantId"`
	Region   string `yaml:"region"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Resourcechecker struct {
	Enabled bool   `yaml:"enabled"`
	Mode    string `yaml:"mode"`
}

type Config struct {
	Nhn struct {
		Auth            Auth            `yaml:"auth"`
		Resourcechecker Resourcechecker `yaml:"Resourcechecker"`
	} `yaml:"nhn"`
}

// Load reads the config file at path and returns the typed Config together
// with the raw document tree. The raw tree is needed by SaveAuth so it can
// round-trip sibling keys (nhn.Autoremover, etc.) that this package doesn't
// otherwise know about. A missing file is not an error: Load returns a zero
// Config with region defaulted.
func Load(path string) (*Config, map[string]any, error) {
	raw := map[string]any{}

	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return nil, nil, fmt.Errorf("config 파일 파싱 실패 (%s): %w", path, err)
		}
	case os.IsNotExist(err):
		// no existing file: start from an empty document
	default:
		return nil, nil, fmt.Errorf("config 파일을 읽을 수 없음 (%s): %w", path, err)
	}

	cfg := &Config{}
	if len(raw) > 0 {
		reencoded, err := yaml.Marshal(raw)
		if err != nil {
			return nil, nil, fmt.Errorf("config 재해석 실패: %w", err)
		}
		if err := yaml.Unmarshal(reencoded, cfg); err != nil {
			return nil, nil, fmt.Errorf("config 파일 파싱 실패 (%s): %w", path, err)
		}
	}
	if cfg.Nhn.Auth.Region == "" {
		cfg.Nhn.Auth.Region = DefaultRegion
	}
	return cfg, raw, nil
}

// SaveAuth writes tenantId/region/username/password under nhn.auth in the
// config file at path, preserving every other key already present in raw
// (including other nhn.* keys such as Resourcechecker/Autoremover). Callers
// are expected to resolve a non-empty region before calling this (e.g. the
// previously saved region, or DefaultRegion) so a credential-only update
// doesn't silently clear it.
func SaveAuth(path string, raw map[string]any, tenantID, region, username, password string) error {
	if raw == nil {
		raw = map[string]any{}
	}

	nhn, _ := raw["nhn"].(map[string]any)
	if nhn == nil {
		nhn = map[string]any{}
	}
	nhn["auth"] = map[string]any{
		"tenantId": tenantID,
		"region":   region,
		"username": username,
		"password": password,
	}
	raw["nhn"] = nhn

	out, err := yaml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("config 직렬화 실패: %w", err)
	}
	// 0o600: this file holds a plaintext credential.
	if err := os.WriteFile(path, out, 0o600); err != nil {
		return fmt.Errorf("config 파일 쓰기 실패 (%s): %w", path, err)
	}
	return nil
}

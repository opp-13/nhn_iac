package cli

import (
	"strings"
	"testing"

	"github.com/opp-13/nhn_iac/resource_checker/nhncloud"
)

func TestMatchInstances(t *testing.T) {
	instances := []nhncloud.Instance{
		{ID: "id-web", Name: "web-01", Status: "ACTIVE"},
		{ID: "id-test1", Name: "test-web", Status: "ACTIVE"},
		{ID: "id-test2", Name: "test-db", Status: "SHUTOFF"},
		{ID: "id-dup1", Name: "dup", Status: "ACTIVE"},
		{ID: "id-dup2", Name: "dup", Status: "ACTIVE"},
	}

	names := func(matched []nhncloud.Instance) []string {
		out := make([]string, 0, len(matched))
		for _, m := range matched {
			out = append(out, m.Name)
		}
		return out
	}

	t.Run("exact name", func(t *testing.T) {
		matched, err := matchInstances(instances, "web-01")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(matched) != 1 || matched[0].ID != "id-web" {
			t.Errorf("matched = %v, want [web-01]", names(matched))
		}
	})

	t.Run("exact ID", func(t *testing.T) {
		matched, err := matchInstances(instances, "id-test2")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(matched) != 1 || matched[0].Name != "test-db" {
			t.Errorf("matched = %v, want [test-db]", names(matched))
		}
	})

	t.Run("glob prefix", func(t *testing.T) {
		matched, err := matchInstances(instances, "test*")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(matched) != 2 {
			t.Errorf("matched = %v, want [test-web test-db]", names(matched))
		}
	})

	t.Run("glob everything", func(t *testing.T) {
		matched, err := matchInstances(instances, "*")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(matched) != len(instances) {
			t.Errorf("matched %d instances, want %d", len(matched), len(instances))
		}
	})

	t.Run("duplicate name errors", func(t *testing.T) {
		_, err := matchInstances(instances, "dup")
		if err == nil {
			t.Fatal("want error for duplicate name, got nil")
		}
		if !strings.Contains(err.Error(), "id-dup1") || !strings.Contains(err.Error(), "id-dup2") {
			t.Errorf("error should list candidate IDs, got: %v", err)
		}
	})

	t.Run("no match errors", func(t *testing.T) {
		if _, err := matchInstances(instances, "nope"); err == nil {
			t.Error("want error for unknown name, got nil")
		}
		if _, err := matchInstances(instances, "nope*"); err == nil {
			t.Error("want error for non-matching glob, got nil")
		}
	})

	t.Run("invalid pattern errors", func(t *testing.T) {
		if _, err := matchInstances(instances, "[unclosed"); err == nil {
			t.Error("want error for malformed pattern, got nil")
		}
	})
}

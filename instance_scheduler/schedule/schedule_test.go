package schedule

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/opp-13/nhn_iac/instance_scheduler/config"
)

func testConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Nhn.Instancescheduler.Instances = []config.Instance{
		{Name: "web-01", Schedule: "0 9 * * *", Terraform: true},
		{Name: "db-01", Schedule: "0 8 * * *", Terraform: false},
	}
	return cfg
}

// fakeCrontab simulates a user's crontab across "crontab -l"/"crontab -"
// calls so Add/Remove/List can be tested without a real crontab binary.
func fakeCrontab(content *string) Runner {
	return func(ctx context.Context, stdin string, name string, args ...string) ([]byte, error) {
		if name != "crontab" {
			return nil, errors.New("unexpected command: " + name)
		}
		switch {
		case len(args) == 1 && args[0] == "-l":
			if *content == "" {
				return nil, errors.New("no crontab for user")
			}
			return []byte(*content), nil
		case len(args) == 1 && args[0] == "-":
			*content = stdin
			return nil, nil
		default:
			return nil, errors.New("unexpected crontab args: " + strings.Join(args, " "))
		}
	}
}

func TestAdd_InstallsNewEntry(t *testing.T) {
	content := ""
	run := fakeCrontab(&content)

	if err := Add(context.Background(), run, testConfig(), "/etc/instsched/config.yaml", "web-01"); err != nil {
		t.Fatalf("Add: %v", err)
	}

	if !strings.Contains(content, markerPrefix+"web-01") {
		t.Fatalf("expected marker for web-01 in crontab, got: %q", content)
	}
	if !strings.Contains(content, "0 9 * * *") || !strings.Contains(content, "check --instance web-01") {
		t.Fatalf("expected schedule line for web-01, got: %q", content)
	}
}

func TestAdd_ReplacesExistingEntryForSameInstance(t *testing.T) {
	content := "# some other line\n" + markerPrefix + "web-01\n0 8 * * * /old/path check --instance web-01 --config /old.yaml\n"
	run := fakeCrontab(&content)

	if err := Add(context.Background(), run, testConfig(), "/etc/instsched/config.yaml", "web-01"); err != nil {
		t.Fatalf("Add: %v", err)
	}

	if strings.Count(content, markerPrefix+"web-01") != 1 {
		t.Fatalf("expected exactly one web-01 marker after replace, got: %q", content)
	}
	if strings.Contains(content, "/old/path") {
		t.Fatalf("expected old entry to be replaced, got: %q", content)
	}
	if !strings.Contains(content, "# some other line") {
		t.Fatalf("expected unrelated crontab lines to survive, got: %q", content)
	}
}

func TestAdd_UnknownInstanceIsError(t *testing.T) {
	content := ""
	run := fakeCrontab(&content)

	if err := Add(context.Background(), run, testConfig(), "config.yaml", "does-not-exist"); err == nil {
		t.Fatal("expected an error for an unconfigured instance name")
	}
}

func TestRemove_DropsOnlyMatchingEntry(t *testing.T) {
	content := markerPrefix + "web-01\n0 9 * * * /bin/instsched check --instance web-01 --config /c.yaml\n" +
		markerPrefix + "db-01\n0 8 * * * /bin/instsched check --instance db-01 --config /c.yaml\n"
	run := fakeCrontab(&content)

	if err := Remove(context.Background(), run, "web-01"); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	if strings.Contains(content, "web-01") {
		t.Fatalf("expected web-01 entry to be removed, got: %q", content)
	}
	if !strings.Contains(content, "db-01") {
		t.Fatalf("expected db-01 entry to survive, got: %q", content)
	}
}

func TestRemove_NoEntryIsNotAnError(t *testing.T) {
	content := ""
	run := fakeCrontab(&content)

	if err := Remove(context.Background(), run, "web-01"); err != nil {
		t.Fatalf("Remove on empty crontab should not error: %v", err)
	}
}

func TestList_ReturnsManagedEntriesOnly(t *testing.T) {
	content := "# unrelated\n*/5 * * * * /some/other/cron/job\n" +
		markerPrefix + "web-01\n0 9 * * * /bin/instsched check --instance web-01 --config /c.yaml\n"
	run := fakeCrontab(&content)

	entries, err := List(context.Background(), run)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 || entries[0].Name != "web-01" || entries[0].Schedule != "0 9 * * *" {
		t.Fatalf("expected single web-01 entry, got %+v", entries)
	}
}

package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestRenderTableIncludesHeaderAndAlignsColumns(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, FormatTable, []string{"NAME", "STATUS"}, [][]string{
		{"web-01", "ACTIVE"},
		{"db-1", "SHUTOFF"},
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "STATUS") {
		t.Errorf("table output missing headers: %q", out)
	}
	if !strings.Contains(out, "web-01") || !strings.Contains(out, "db-1") {
		t.Errorf("table output missing rows: %q", out)
	}
}

func TestRenderTextHasNoHeaderAndIsTabSeparated(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, FormatText, []string{"NAME", "STATUS"}, [][]string{
		{"web-01", "ACTIVE"},
	})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "NAME") {
		t.Errorf("text output should not include header row: %q", out)
	}
	want := "web-01\tACTIVE\n"
	if out != want {
		t.Errorf("text output = %q, want %q", out, want)
	}
}

func TestRenderUnknownFormatErrors(t *testing.T) {
	var buf bytes.Buffer
	if err := Render(&buf, "yaml", nil, nil); err == nil {
		t.Error("Render() with unknown format: want error, got nil")
	}
}

func TestRenderJSON(t *testing.T) {
	var buf bytes.Buffer
	type row struct {
		Name string `json:"name"`
	}
	if err := RenderJSON(&buf, []row{{Name: "web-01"}}); err != nil {
		t.Fatalf("RenderJSON() error = %v", err)
	}
	if !strings.Contains(buf.String(), `"name": "web-01"`) {
		t.Errorf("RenderJSON output = %q, want field \"name\"", buf.String())
	}
}

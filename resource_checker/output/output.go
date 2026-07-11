// Package output renders resource_checker query results as aligned tables,
// paste-friendly tab-separated text, or JSON.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

const (
	FormatTable = "table"
	FormatText  = "text"
	FormatJSON  = "json"
)

// Valid reports whether format is one of table/json/text.
func Valid(format string) bool {
	switch format {
	case FormatTable, FormatText, FormatJSON:
		return true
	default:
		return false
	}
}

// Render writes headers+rows as an aligned table (FormatTable) or as
// tab-separated lines with no header (FormatText, for pasting elsewhere).
// It does not handle FormatJSON — callers marshal their own typed structs
// via RenderJSON so JSON output keeps real field names instead of the
// display-only headers used here.
func Render(w io.Writer, format string, headers []string, rows [][]string) error {
	switch format {
	case FormatTable:
		tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
		fmt.Fprintln(tw, strings.Join(headers, "\t"))
		for _, row := range rows {
			fmt.Fprintln(tw, strings.Join(row, "\t"))
		}
		return tw.Flush()
	case FormatText:
		for _, row := range rows {
			fmt.Fprintln(w, strings.Join(row, "\t"))
		}
		return nil
	default:
		return fmt.Errorf("output: 지원하지 않는 포맷 %q (table|json|text 중 하나여야 함)", format)
	}
}

// RenderJSON writes v as an indented JSON array/object.
func RenderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

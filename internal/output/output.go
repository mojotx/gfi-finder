// Package output renders search results as a human-readable table or as
// JSON for scripting.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/mojotx/gfi-finder/internal/search"
)

// RenderTable writes issues to w as a simple aligned table.
func RenderTable(w io.Writer, issues []search.Issue) error {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "NUMBER\tTITLE\tURL\tLABELS"); err != nil {
		return err
	}
	for _, iss := range issues {
		if _, err := fmt.Fprintf(tw, "#%d\t%s\t%s\t%s\n", iss.Number, iss.Title, iss.URL, strings.Join(iss.Labels, ", ")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

// RenderJSON writes issues to w as indented JSON.
func RenderJSON(w io.Writer, issues []search.Issue) error {
	if issues == nil {
		issues = []search.Issue{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(issues)
}

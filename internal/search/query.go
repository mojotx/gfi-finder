// Package search builds and executes GitHub search queries for candidate
// "good first issue" style issues, and models the results.
package search

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var repoPattern = regexp.MustCompile(`^[^/\s]+/[^/\s]+$`)

// BuildQuery builds a GitHub search/issues query string that finds open
// issues in repo matching labels (OR'd), excluding issues that already have
// a linked pull request (via the -linked:pr qualifier). When unassigned is
// true, the no:assignee qualifier is added to exclude issues that already
// have an assignee.
func BuildQuery(repo string, labels []string, unassigned bool) (string, error) {
	if !repoPattern.MatchString(repo) {
		return "", fmt.Errorf("invalid repo %q: expected format OWNER/REPO", repo)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "repo:%s is:issue is:open", repo)
	if unassigned {
		b.WriteString(" no:assignee")
	}
	b.WriteString(" -linked:pr")

	if len(labels) > 0 {
		quoted := make([]string, len(labels))
		for i, l := range labels {
			quoted[i] = strconv.Quote(l)
		}
		fmt.Fprintf(&b, " label:%s", strings.Join(quoted, ","))
	}

	return b.String(), nil
}

// SplitRepo splits an OWNER/REPO string into its owner and repo name parts.
// The caller is expected to have already validated the format, e.g. via
// BuildQuery.
func SplitRepo(repo string) (owner, name string, err error) {
	parts := strings.SplitN(repo, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repo %q: expected format OWNER/REPO", repo)
	}
	return parts[0], parts[1], nil
}

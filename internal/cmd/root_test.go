package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mojotx/gfi-finder/internal/search"
)

type fakeSearcher struct {
	issues []search.Issue
	err    error

	gotQuery string
	gotLimit int
}

func (f *fakeSearcher) SearchIssues(query string, limit int) ([]search.Issue, error) {
	f.gotQuery = query
	f.gotLimit = limit
	return f.issues, f.err
}

type fakeChecker struct {
	linked map[int]bool
	err    error
}

func (f *fakeChecker) HasLinkedPR(owner, repo string, number int) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	return f.linked[number], nil
}

func TestRun_tableOutput(t *testing.T) {
	searcher := &fakeSearcher{issues: []search.Issue{
		{Number: 1, Title: "Fix bug", URL: "https://github.com/o/r/issues/1", Labels: []string{"bug"}},
	}}
	opts := &Options{Repo: "o/r", Limit: 10}

	var buf bytes.Buffer
	err := run(opts, searcher, nil, &buf)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "Fix bug")
	assert.Contains(t, searcher.gotQuery, "repo:o/r")
	assert.Equal(t, 10, searcher.gotLimit)
}

func TestRun_jsonOutput(t *testing.T) {
	searcher := &fakeSearcher{issues: []search.Issue{
		{Number: 1, Title: "Fix bug", URL: "https://github.com/o/r/issues/1", Labels: []string{"bug"}},
	}}
	opts := &Options{Repo: "o/r", JSON: true}

	var buf bytes.Buffer
	err := run(opts, searcher, nil, &buf)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"title": "Fix bug"`)
}

func TestRun_invalidRepo(t *testing.T) {
	opts := &Options{Repo: "not-a-valid-repo"}
	var buf bytes.Buffer
	err := run(opts, &fakeSearcher{}, nil, &buf)
	require.Error(t, err)
}

func TestRun_searchError(t *testing.T) {
	searcher := &fakeSearcher{err: errors.New("boom")}
	opts := &Options{Repo: "o/r"}
	var buf bytes.Buffer
	err := run(opts, searcher, nil, &buf)
	require.Error(t, err)
}

func TestRun_strictModeFiltersLinkedIssues(t *testing.T) {
	searcher := &fakeSearcher{issues: []search.Issue{
		{Number: 1, Title: "Not linked"},
		{Number: 2, Title: "Linked via mention"},
	}}
	checker := &fakeChecker{linked: map[int]bool{2: true}}
	opts := &Options{Repo: "o/r", StrictLinkCheck: true, JSON: true}

	var buf bytes.Buffer
	err := run(opts, searcher, checker, &buf)
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "Not linked")
	assert.NotContains(t, out, "Linked via mention")
}

func TestRun_strictModeCheckerError(t *testing.T) {
	searcher := &fakeSearcher{issues: []search.Issue{{Number: 1, Title: "Issue"}}}
	checker := &fakeChecker{err: errors.New("boom")}
	opts := &Options{Repo: "o/r", StrictLinkCheck: true}

	var buf bytes.Buffer
	err := run(opts, searcher, checker, &buf)
	require.Error(t, err)
}

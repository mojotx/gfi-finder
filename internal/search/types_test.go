package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleSearchResponse = `{
  "total_count": 2,
  "items": [
    {
      "number": 42,
      "title": "Fix flaky test",
      "html_url": "https://github.com/cli/cli/issues/42",
      "labels": [{"name": "good first issue"}, {"name": "bug"}]
    },
    {
      "number": 99,
      "title": "A pull request that snuck in",
      "html_url": "https://github.com/cli/cli/pull/99",
      "labels": [],
      "pull_request": {"url": "https://api.github.com/repos/cli/cli/pulls/99"}
    }
  ]
}`

func TestParseSearchResponse(t *testing.T) {
	issues, err := ParseSearchResponse([]byte(sampleSearchResponse))
	require.NoError(t, err)

	require.Len(t, issues, 1, "pull requests must be filtered out")
	assert.Equal(t, Issue{
		Number: 42,
		Title:  "Fix flaky test",
		URL:    "https://github.com/cli/cli/issues/42",
		Labels: []string{"good first issue", "bug"},
	}, issues[0])
}

func TestParseSearchResponse_invalidJSON(t *testing.T) {
	_, err := ParseSearchResponse([]byte("not json"))
	require.Error(t, err)
}

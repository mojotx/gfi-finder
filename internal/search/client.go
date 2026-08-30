package search

import (
	"fmt"
	"net/url"
)

const (
	perPage  = 100
	maxPages = 10 // GitHub's Search API caps results at 1000.
)

// DefaultLimit is used when no positive limit is supplied to SearchIssues.
const DefaultLimit = 30

// RESTGetter is the subset of api.RESTClient used by Client. It is
// satisfied by *github.com/cli/go-gh/v2/pkg/api.RESTClient.
type RESTGetter interface {
	Get(path string, resp interface{}) error
}

// Searcher finds candidate issues for a given search query.
type Searcher interface {
	SearchIssues(query string, limit int) ([]Issue, error)
}

// Client executes GitHub search/issues queries via the REST API.
type Client struct {
	getter RESTGetter
}

// NewClient returns a Client that issues requests through getter.
func NewClient(getter RESTGetter) *Client {
	return &Client{getter: getter}
}

// SearchIssues runs query against the GitHub search/issues API, paginating
// until limit results have been collected or the API is exhausted. If limit
// is not positive, DefaultLimit is used.
func (c *Client) SearchIssues(query string, limit int) ([]Issue, error) {
	if limit <= 0 {
		limit = DefaultLimit
	}

	var issues []Issue
	for page := 1; page <= maxPages && len(issues) < limit; page++ {
		path := fmt.Sprintf("search/issues?q=%s&per_page=%d&page=%d", url.QueryEscape(query), perPage, page)

		var resp searchResponse
		if err := c.getter.Get(path, &resp); err != nil {
			return nil, fmt.Errorf("search issues: %w", err)
		}

		issues = append(issues, resp.toIssues()...)
		if len(resp.Items) < perPage {
			break
		}
	}

	if len(issues) > limit {
		issues = issues[:limit]
	}
	return issues, nil
}

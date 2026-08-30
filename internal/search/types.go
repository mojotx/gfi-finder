package search

import "encoding/json"

// Issue is a candidate "good first issue" returned by a search.
type Issue struct {
	Number int      `json:"number"`
	Title  string   `json:"title"`
	URL    string   `json:"url"`
	Labels []string `json:"labels"`
}

// searchResponse models the relevant subset of the GitHub REST
// "Search issues and pull requests" API response.
type searchResponse struct {
	TotalCount int          `json:"total_count"`
	Items      []searchItem `json:"items"`
}

type searchItem struct {
	Number      int             `json:"number"`
	Title       string          `json:"title"`
	HTMLURL     string          `json:"html_url"`
	Labels      []searchLabel   `json:"labels"`
	PullRequest json.RawMessage `json:"pull_request,omitempty"`
}

type searchLabel struct {
	Name string `json:"name"`
}

// toIssues converts the raw response items to Issues, defensively dropping
// any item that is actually a pull request (the search query already
// constrains results to is:issue, but this guards against future changes).
func (r searchResponse) toIssues() []Issue {
	issues := make([]Issue, 0, len(r.Items))
	for _, item := range r.Items {
		if item.PullRequest != nil {
			continue
		}
		labels := make([]string, len(item.Labels))
		for i, l := range item.Labels {
			labels[i] = l.Name
		}
		issues = append(issues, Issue{
			Number: item.Number,
			Title:  item.Title,
			URL:    item.HTMLURL,
			Labels: labels,
		})
	}
	return issues
}

// ParseSearchResponse decodes a raw GitHub search/issues API response body
// into a slice of Issues.
func ParseSearchResponse(data []byte) ([]Issue, error) {
	var resp searchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.toIssues(), nil
}

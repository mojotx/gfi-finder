// Package linkcheck implements the strict-mode secondary pass that inspects
// an issue's GraphQL timeline for any pull request that has ever
// cross-referenced or formally connected to it, even without a closing
// keyword. This catches issues that the fast-mode -linked:pr search
// qualifier misses.
package linkcheck

import (
	_ "embed"
	"fmt"
)

//go:embed timeline.graphql
var timelineQuery string

// GraphQLDoer is the subset of api.GraphQLClient used by Client. It is
// satisfied by *github.com/cli/go-gh/v2/pkg/api.GraphQLClient.
type GraphQLDoer interface {
	Do(query string, variables map[string]interface{}, response interface{}) error
}

// Checker determines whether an issue has ever been referenced by a pull
// request, formally or informally.
type Checker interface {
	HasLinkedPR(owner, repo string, number int) (bool, error)
}

// Client checks issues' timelines via the GitHub GraphQL API.
type Client struct {
	doer GraphQLDoer
}

// NewClient returns a Client that issues requests through doer.
func NewClient(doer GraphQLDoer) *Client {
	return &Client{doer: doer}
}

type timelineTypedEntity struct {
	Typename string `json:"__typename"`
}

type timelineNode struct {
	Typename string               `json:"__typename"`
	Source   *timelineTypedEntity `json:"source,omitempty"`
	Subject  *timelineTypedEntity `json:"subject,omitempty"`
}

type timelineResponse struct {
	Repository struct {
		Issue struct {
			TimelineItems struct {
				Nodes []timelineNode `json:"nodes"`
			} `json:"timelineItems"`
		} `json:"issue"`
	} `json:"repository"`
}

// hasLinkedPR reports whether any timeline node references a pull request,
// either as the source of a cross-reference or the subject of a formal
// connection.
func hasLinkedPR(nodes []timelineNode) bool {
	for _, n := range nodes {
		if n.Source != nil && n.Source.Typename == "PullRequest" {
			return true
		}
		if n.Subject != nil && n.Subject.Typename == "PullRequest" {
			return true
		}
	}
	return false
}

// HasLinkedPR reports whether the issue numbered number in owner/repo has
// ever been cross-referenced or connected to a pull request.
func (c *Client) HasLinkedPR(owner, repo string, number int) (bool, error) {
	vars := map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}

	var resp timelineResponse
	if err := c.doer.Do(timelineQuery, vars, &resp); err != nil {
		return false, fmt.Errorf("check linked pull requests for issue #%d: %w", number, err)
	}

	return hasLinkedPR(resp.Repository.Issue.TimelineItems.Nodes), nil
}

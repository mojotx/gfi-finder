package linkcheck

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasLinkedPR_pureLogic(t *testing.T) {
	pr := &timelineTypedEntity{Typename: "PullRequest"}
	issue := &timelineTypedEntity{Typename: "Issue"}

	tests := []struct {
		name  string
		nodes []timelineNode
		want  bool
	}{
		{name: "no nodes", nodes: nil, want: false},
		{name: "unrelated cross-reference", nodes: []timelineNode{{Source: issue}}, want: false},
		{name: "cross-referenced by PR", nodes: []timelineNode{{Source: pr}}, want: true},
		{name: "connected to PR", nodes: []timelineNode{{Subject: pr}}, want: true},
		{
			name: "PR link among other events",
			nodes: []timelineNode{
				{Source: issue},
				{Subject: pr},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, hasLinkedPR(tt.nodes))
		})
	}
}

// roundTripFunc lets a plain function satisfy http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestGraphQLClient(t *testing.T, rt roundTripFunc) *api.GraphQLClient {
	t.Helper()
	client, err := api.NewGraphQLClient(api.ClientOptions{
		Host:      "github.com",
		AuthToken: "test-token",
		Transport: rt,
	})
	require.NoError(t, err)
	return client
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func TestClient_HasLinkedPR(t *testing.T) {
	body := `{"data":{"repository":{"issue":{"timelineItems":{"nodes":[
		{"__typename":"CrossReferencedEvent","source":{"__typename":"PullRequest"}}
	]}}}}}`

	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(body), nil
	})

	client := NewClient(newTestGraphQLClient(t, rt))
	linked, err := client.HasLinkedPR("cli", "cli", 42)
	require.NoError(t, err)
	assert.True(t, linked)
}

func TestClient_HasLinkedPR_noLink(t *testing.T) {
	body := `{"data":{"repository":{"issue":{"timelineItems":{"nodes":[]}}}}}`

	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(body), nil
	})

	client := NewClient(newTestGraphQLClient(t, rt))
	linked, err := client.HasLinkedPR("cli", "cli", 42)
	require.NoError(t, err)
	assert.False(t, linked)
}

func TestClient_HasLinkedPR_error(t *testing.T) {
	body := `{"errors":[{"message":"not found"}]}`

	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(body), nil
	})

	client := NewClient(newTestGraphQLClient(t, rt))
	_, err := client.HasLinkedPR("cli", "cli", 42)
	require.Error(t, err)
}

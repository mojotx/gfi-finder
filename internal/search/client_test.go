package search

import (
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// roundTripFunc lets a plain function satisfy http.RoundTripper, following
// the stubbing style used by cli/cli's own tests.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestRESTClient(t *testing.T, rt roundTripFunc) *api.RESTClient {
	t.Helper()
	client, err := api.NewRESTClient(api.ClientOptions{
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

func TestClient_SearchIssues_singlePage(t *testing.T) {
	var requests int
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		assert.Contains(t, req.URL.String(), "search/issues")
		assert.Contains(t, req.URL.RawQuery, "page=1")
		return jsonResponse(sampleSearchResponse), nil
	})

	client := NewClient(newTestRESTClient(t, rt))
	issues, err := client.SearchIssues("repo:cli/cli is:issue", 10)
	require.NoError(t, err)

	assert.Equal(t, 1, requests)
	require.Len(t, issues, 1)
	assert.Equal(t, 42, issues[0].Number)
}

func TestClient_SearchIssues_paginatesUntilShortPage(t *testing.T) {
	fullPage := makeSearchResponse(perPage, 1)
	shortPage := makeSearchResponse(5, perPage+1)

	var requests []string
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req.URL.RawQuery)
		if len(requests) == 1 {
			return jsonResponse(fullPage), nil
		}
		return jsonResponse(shortPage), nil
	})

	client := NewClient(newTestRESTClient(t, rt))
	issues, err := client.SearchIssues("repo:cli/cli is:issue", 1000)
	require.NoError(t, err)

	assert.Len(t, requests, 2, "expected pagination to stop after a short page")
	assert.Len(t, issues, perPage+5)
}

func TestClient_SearchIssues_respectsLimit(t *testing.T) {
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(makeSearchResponse(perPage, 1)), nil
	})

	client := NewClient(newTestRESTClient(t, rt))
	issues, err := client.SearchIssues("repo:cli/cli is:issue", 3)
	require.NoError(t, err)

	assert.Len(t, issues, 3)
}

func TestClient_SearchIssues_httpError(t *testing.T) {
	rt := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader(`{"message":"boom"}`)),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Request:    req,
		}, nil
	})

	client := NewClient(newTestRESTClient(t, rt))
	_, err := client.SearchIssues("repo:cli/cli is:issue", 10)
	require.Error(t, err)
}

// makeSearchResponse builds a synthetic search response JSON body with n
// items, numbered starting at startNumber.
func makeSearchResponse(n, startNumber int) string {
	var b strings.Builder
	b.WriteString(`{"total_count":0,"items":[`)
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(`{"number":`)
		b.WriteString(strconv.Itoa(startNumber + i))
		b.WriteString(`,"title":"issue","html_url":"https://github.com/cli/cli/issues/1","labels":[]}`)
	}
	b.WriteString(`]}`)
	return b.String()
}

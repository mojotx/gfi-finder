package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mojotx/gfi-finder/internal/search"
)

func TestRenderTable(t *testing.T) {
	var buf bytes.Buffer
	err := RenderTable(&buf, []search.Issue{
		{Number: 1, Title: "Fix bug", URL: "https://github.com/o/r/issues/1", Labels: []string{"bug", "good first issue"}},
	})
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "NUMBER")
	assert.Contains(t, out, "#1")
	assert.Contains(t, out, "Fix bug")
	assert.Contains(t, out, "https://github.com/o/r/issues/1")
	assert.Contains(t, out, "bug, good first issue")
}

func TestRenderTable_empty(t *testing.T) {
	var buf bytes.Buffer
	err := RenderTable(&buf, nil)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "NUMBER")
}

func TestRenderJSON(t *testing.T) {
	var buf bytes.Buffer
	issues := []search.Issue{
		{Number: 1, Title: "Fix bug", URL: "https://github.com/o/r/issues/1", Labels: []string{"bug"}},
	}
	err := RenderJSON(&buf, issues)
	require.NoError(t, err)

	var got []search.Issue
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.Equal(t, issues, got)
}

func TestRenderJSON_nilRendersEmptyArray(t *testing.T) {
	var buf bytes.Buffer
	err := RenderJSON(&buf, nil)
	require.NoError(t, err)
	assert.JSONEq(t, "[]", buf.String())
}

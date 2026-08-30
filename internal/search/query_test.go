package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildQuery(t *testing.T) {
	tests := []struct {
		name       string
		repo       string
		labels     []string
		unassigned bool
		want       string
		wantErr    bool
	}{
		{
			name:       "no labels, unassigned",
			repo:       "cli/cli",
			labels:     nil,
			unassigned: true,
			want:       `repo:cli/cli is:issue is:open no:assignee -linked:pr`,
		},
		{
			name:       "single label",
			repo:       "cli/cli",
			labels:     []string{"good first issue"},
			unassigned: true,
			want:       `repo:cli/cli is:issue is:open no:assignee -linked:pr label:"good first issue"`,
		},
		{
			name:       "multiple labels are OR'd",
			repo:       "cli/cli",
			labels:     []string{"help wanted", "good first issue"},
			unassigned: true,
			want:       `repo:cli/cli is:issue is:open no:assignee -linked:pr label:"help wanted","good first issue"`,
		},
		{
			name:       "allow assigned omits no:assignee",
			repo:       "cli/cli",
			labels:     nil,
			unassigned: false,
			want:       `repo:cli/cli is:issue is:open -linked:pr`,
		},
		{
			name:    "missing slash is invalid",
			repo:    "cli",
			wantErr: true,
		},
		{
			name:    "empty repo is invalid",
			repo:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildQuery(tt.repo, tt.labels, tt.unassigned)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSplitRepo(t *testing.T) {
	tests := []struct {
		name      string
		repo      string
		wantOwner string
		wantName  string
		wantErr   bool
	}{
		{name: "valid", repo: "cli/cli", wantOwner: "cli", wantName: "cli"},
		{name: "no slash", repo: "cli", wantErr: true},
		{name: "empty owner", repo: "/cli", wantErr: true},
		{name: "empty name", repo: "cli/", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, name, err := SplitRepo(tt.repo)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOwner, owner)
			assert.Equal(t, tt.wantName, name)
		})
	}
}

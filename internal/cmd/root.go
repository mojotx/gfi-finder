// Package cmd implements the gfi-finder command line interface.
package cmd

import (
	"io"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"

	"github.com/mojotx/gfi-finder/internal/linkcheck"
	"github.com/mojotx/gfi-finder/internal/output"
	"github.com/mojotx/gfi-finder/internal/search"
)

const longDescription = `gfi-finder finds "good first issue" candidates in a GitHub repository:
open issues matching the given labels that are unassigned and have no
linked pull request.

Fast mode (default) relies solely on GitHub's search "-linked:pr"
qualifier, which excludes issues formally linked to a pull request via
the "Development" sidebar (i.e. a PR using a closing keyword like
"Fixes #N"). This is cheap - one search request per page - but it will
not catch issues that were merely mentioned by a PR or commit that never
used a closing keyword.

Strict mode (--strict-link-check) additionally inspects each candidate
issue's GraphQL timeline for any pull request that has ever
cross-referenced or connected to it, even without a closing keyword.
This is slower (one extra request per candidate issue) so it only runs
against the fast-mode shortlist, never the whole repository.`

// Options holds the flag values for the root command.
type Options struct {
	Repo            string
	Labels          []string
	AllowAssigned   bool
	StrictLinkCheck bool
	Limit           int
	JSON            bool
}

// NewRootCmd builds the gfi-finder root command, wired to real GitHub API
// clients (picked up from the user's `gh auth login` session via go-gh).
func NewRootCmd() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:           "gfi-finder",
		Short:         `Find "good first issue" candidates in a GitHub repository`,
		Long:          longDescription,
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			restClient, err := api.NewRESTClient(api.ClientOptions{})
			if err != nil {
				return err
			}
			searcher := search.NewClient(restClient)

			var checker linkcheck.Checker
			if opts.StrictLinkCheck {
				gqlClient, err := api.NewGraphQLClient(api.ClientOptions{})
				if err != nil {
					return err
				}
				checker = linkcheck.NewClient(gqlClient)
			}

			return run(opts, searcher, checker, cmd.OutOrStdout())
		},
	}

	cmd.Flags().StringVar(&opts.Repo, "repo", "", "the OWNER/REPO to search (required)")
	_ = cmd.MarkFlagRequired("repo")
	cmd.Flags().StringSliceVarP(&opts.Labels, "label", "l", nil, "label to match, repeatable or comma-separated (OR'd)")
	cmd.Flags().BoolVar(&opts.AllowAssigned, "allow-assigned", false, "include issues that already have an assignee")
	cmd.Flags().BoolVar(&opts.StrictLinkCheck, "strict-link-check", false, "also exclude issues ever mentioned by a PR, not just formally linked ones (slower)")
	cmd.Flags().IntVar(&opts.Limit, "limit", search.DefaultLimit, "maximum number of issues to return")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "output as JSON instead of a table")

	return cmd
}

// run performs the search and, if requested, the strict-mode filtering
// pass, then renders the results to out. It is decoupled from cobra and the
// concrete API clients so it can be unit tested with fakes.
func run(opts *Options, searcher search.Searcher, checker linkcheck.Checker, out io.Writer) error {
	query, err := search.BuildQuery(opts.Repo, opts.Labels, !opts.AllowAssigned)
	if err != nil {
		return err
	}

	issues, err := searcher.SearchIssues(query, opts.Limit)
	if err != nil {
		return err
	}

	if opts.StrictLinkCheck {
		owner, name, err := search.SplitRepo(opts.Repo)
		if err != nil {
			return err
		}

		filtered := issues[:0]
		for _, iss := range issues {
			linked, err := checker.HasLinkedPR(owner, name, iss.Number)
			if err != nil {
				return err
			}
			if !linked {
				filtered = append(filtered, iss)
			}
		}
		issues = filtered
	}

	if opts.JSON {
		return output.RenderJSON(out, issues)
	}
	return output.RenderTable(out, issues)
}

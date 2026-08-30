# gfi-finder -- Find Good First Issues on GitHub

![gfi-finder mascot](assets/gfi-gopher.png)

[![CI](https://github.com/mojotx/gfi-finder/actions/workflows/ci.yml/badge.svg)](https://github.com/mojotx/gfi-finder/actions/workflows/ci.yml)
[![CodeQL](https://github.com/mojotx/gfi-finder/actions/workflows/codeql.yml/badge.svg)](https://github.com/mojotx/gfi-finder/actions/workflows/codeql.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mojotx/gfi-finder.svg)](https://pkg.go.dev/github.com/mojotx/gfi-finder)

`gfi-finder` finds "good first issue" candidates in a GitHub repository:
open issues matching your labels that are **unassigned** and have **no
linked pull request**. The name comes from its original use case, but any
label(s) work — `help wanted`, `bug`, `documentation`, or anything else
your project uses to flag actionable issues.

It uses [`github.com/cli/go-gh`](https://github.com/cli/go-gh) — the same
library the `gh` CLI itself uses — so it picks up your existing
`gh auth login` session for free. No separate token setup required, as
long as the [`gh` CLI](https://cli.github.com/) is installed and
authenticated (or a `GH_TOKEN`/`GITHUB_TOKEN` environment variable is set,
which go-gh also honors without requiring `gh` itself to be installed).

## How it works

GitHub's search API supports a `-linked:pr` qualifier that excludes issues
already connected to a pull request via the "Development" sidebar (i.e. a PR
using a closing keyword like `Fixes #N` or `Closes #N`, open or closed).
Combined with `no:assignee` and a comma-separated `label:"A","B"` list
(which GitHub OR's together), the whole search collapses into **one query**:

```
repo:OWNER/REPO is:issue is:open no:assignee -linked:pr label:"help wanted","good first issue"
```

No need to loop over every issue and inspect its timeline — the Search API
does that filtering for you.

## Two matching modes

### Fast mode (default)

Relies solely on the `-linked:pr` search qualifier. Cheap — one request per
page of results — but it only catches issues **formally** linked to a pull
request (i.e. a PR that used a closing keyword). It will not catch issues
that were merely *mentioned* by a PR or commit that never used a closing
keyword.

### Strict mode (`--strict-link-check`)

After the fast-mode shortlist is built, `gfi-finder` does a secondary pass
per candidate issue using the GraphQL `timelineItems` field
(`CROSS_REFERENCED_EVENT` and `CONNECTED_EVENT`) to also exclude issues that
were ever mentioned by *any* pull request, even without a formal closing
keyword. This is slower — one extra request per candidate issue — so it
only ever runs against the already-filtered shortlist, never the whole
repository.

This distinction is a real, non-obvious nuance in GitHub's data model: a PR
can reference an issue in its description or a commit message without
"closing" it, and the Search API's `-linked:pr` qualifier has no visibility
into that. Strict mode trades speed for a more complete picture.

## Installation

```sh
go install github.com/mojotx/gfi-finder@latest
```

### Authentication

`gfi-finder` needs a GitHub token to make API requests. The easiest way is
to install the [`gh` CLI](https://cli.github.com/) and run
`gh auth login` once; `gfi-finder` will pick up that session automatically.
Without it, you'll see:

```text
Error: authentication token not found for host github.com
```

If you'd rather not install `gh`, set a `GH_TOKEN` or `GITHUB_TOKEN`
environment variable to a [personal access
token](https://github.com/settings/tokens) instead.

## Usage

```sh
# Fast mode: candidates with a "good first issue" or "help wanted" label
gfi-finder --repo kubernetes/kubernetes --label "good first issue" --label "help wanted"

# Comma-separated labels work too (OR'd, same as repeating --label)
gfi-finder --repo kubernetes/kubernetes --label "good first issue,help wanted"

# Strict mode: also filter out issues merely mentioned by a PR
gfi-finder --repo kubernetes/kubernetes --label "good first issue" --strict-link-check

# Include issues that already have an assignee
gfi-finder --repo kubernetes/kubernetes --label "good first issue" --allow-assigned

# Machine-readable output for scripting
gfi-finder --repo kubernetes/kubernetes --label "good first issue" --json --limit 50
```

### Flags

| Flag                  | Default | Description                                                          |
| --------------------- | ------- | -------------------------------------------------------------------- |
| `--help`, `-h`        | —       | Help                                                                 |
| `--repo`              | —       | `OWNER/REPO` to search (required)                                    |
| `--label`, `-l`       | —       | Label to match; repeatable or comma-separated (OR'd)                 |
| `--allow-assigned`    | `false` | Include issues that already have an assignee                         |
| `--strict-link-check` | `false` | Also exclude issues ever mentioned by a PR, not just formally linked |
| `--limit`             | `30`    | Maximum number of issues to return                                   |
| `--json`              | `false` | Output as JSON instead of a table                                    |

## Rate limits and caching

The Search API is capped at **30 requests/minute** (authenticated) — much
stricter than the 5,000/hour core REST limit. That's fine for interactive,
single-repo use. If you plan to scan many repos or run this on a schedule,
add backoff/retry and consider caching results locally (e.g. a JSON file
keyed by repo + query + timestamp) to avoid hammering the API.

The strict-mode secondary pass consumes normal REST/GraphQL rate-limit
budget instead of the Search API's, so it scales differently — it's
opt-in and bounded by the size of the fast-mode shortlist, never the whole
repository.

## Building

```sh
go build -o gfi-finder .
```

## Testing

```sh
go test -race ./...
golangci-lint run --timeout=5m --allow-parallel-runners --max-same-issues 0 --max-issues-per-linter 0 ./...
```

## Non-goals (v1)

- Multi-repo / org-wide scanning
- A persistent daemon or scheduled runner
- A web UI

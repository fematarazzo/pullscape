# pullscape

Generates SVG cards showing a GitHub user's pull request history. Embed them in a profile README or anywhere that renders images.

## How it works

One endpoint fetches all PRs for a given user via the GitHub GraphQL API, processes them, and returns an SVG table. No database, no caching layer.

```
GET /api/github-pr-stats?username=<github_username>
```

## Parameters

| Parameter  | Default            | Description                                                                 |
|------------|--------------------|-----------------------------------------------------------------------------|
| `username` | required           | GitHub username                                                             |
| `theme`    | `dark`             | `dark` or `light`                                                           |
| `mode`     | PR list            | `repo-aggregate` to group by repository instead of listing individual PRs  |
| `status`   | `all`              | Filter by status: `merged`, `open`, `draft`, `closed`, or comma-separated  |
| `min_stars`| `0`                | Only include PRs from repos with at least this many stars                   |
| `limit`    | `10`               | Max rows to show                                                            |
| `sort`     | `status,stars_desc`| Sort fields, comma-separated (see below)                                    |
| `fields`   | all columns        | Columns to include, comma-separated (see below)                             |
| `stats`    | `all`              | Which summary stats to show in the header bar                               |

### Sort options

PR list mode: `stars_desc`, `stars_asc`, `created_date_desc`, `created_date_asc`, `status`

Repo aggregate mode: `stars_desc`, `stars_asc`, `merged_desc`, `merged_asc`, `merged_rate_desc`, `merged_rate_asc`

### Field options

PR list: `repo`, `stars`, `pr_title`, `pr_number`, `status`, `created_date`, `merged_date`

Repo aggregate: `repo`, `stars`, `pr_numbers`, `total`, `merged`, `open`, `draft`, `closed`, `merged_rate`

### Stats options

`total_pr`, `merged_pr`, `display_pr`, `repos_with_pr`, `repos_with_merged_pr`, `showing_repos`

## Examples

```
/api/github-pr-stats?username=torvalds&theme=light&limit=5
/api/github-pr-stats?username=torvalds&status=merged&sort=stars_desc&limit=10
/api/github-pr-stats?username=torvalds&mode=repo-aggregate&sort=merged_rate_desc
/api/github-pr-stats?username=torvalds&fields=repo,stars,status&stats=total_pr,merged_pr
```

## Running locally

```sh
cp .env.example .env
# add your GitHub token to .env
go run .
```

The server starts on port 8080 by default. Set `PORT` in `.env` to change it.

## Environment variables

| Variable       | Description                                                  |
|----------------|--------------------------------------------------------------|
| `GITHUB_TOKEN` | GitHub personal access token. Needs `read:user` scope.      |
| `PORT`         | Port to listen on. Defaults to `8080`.                       |

## Deployment

The repo includes a `Dockerfile` and `fly.toml` for deploying to fly.io.

```sh
fly launch
fly secrets set GITHUB_TOKEN=your_token_here
fly deploy
```

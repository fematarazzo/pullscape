# pullscape

Generates SVG cards showing a GitHub user's pull request history. Embed them in a profile README or anywhere that renders images.

**Live:** https://pullscape.fly.dev

## Usage

```
GET https://pullscape.fly.dev/api/github-pr-stats?username=<github_username>
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
https://pullscape.fly.dev/api/github-pr-stats?username=torvalds&theme=light&limit=5
https://pullscape.fly.dev/api/github-pr-stats?username=torvalds&status=merged&sort=stars_desc&limit=10
https://pullscape.fly.dev/api/github-pr-stats?username=torvalds&mode=repo-aggregate&sort=merged_rate_desc
https://pullscape.fly.dev/api/github-pr-stats?username=torvalds&fields=repo,stars,status&stats=total_pr,merged_pr
```

To embed in a GitHub README:

```md
![Pull Request Activity](https://pullscape.fly.dev/api/github-pr-stats?username=torvalds&theme=dark&limit=10)
```

## Running locally

```sh
cp .env.example .env
# add your GitHub token to .env
go run .
```

The server starts on port 8080 by default. Set `PORT` in `.env` to change it.

## Environment variables

| Variable                    | Required | Description                                                        |
|-----------------------------|----------|--------------------------------------------------------------------|
| `GITHUB_TOKEN`              | yes      | GitHub personal access token. Needs `read:user` scope.            |
| `PORT`                      | no       | Port to listen on. Defaults to `8080`.                             |
| `UPSTASH_REDIS_REST_URL`    | no       | Upstash Redis REST endpoint. Enables persistent cache across restarts. |
| `UPSTASH_REDIS_REST_TOKEN`  | no       | Upstash Redis REST token.                                          |

Without Redis the server falls back to an in-memory cache (wiped on restart).

## Caching

Responses are cached for 1 hour. Cache lookup order:

1. In-memory (`sync.Map`) — fastest, per-process
2. Redis — survives restarts; first request after a cold start still returns instantly
3. GitHub API — only on a full cache miss

## Deployment

The repo includes a `Dockerfile` and `fly.toml` for deploying to Fly.io.

```sh
fly apps create pullscape
fly secrets set GITHUB_TOKEN=your_token_here
fly deploy
```

To add persistent Redis cache (recommended):

```sh
fly redis create
fly secrets set UPSTASH_REDIS_REST_URL=https://... UPSTASH_REDIS_REST_TOKEN=...
```

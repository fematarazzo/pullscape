package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var githubClient = &http.Client{Timeout: 10 * time.Second}

const githubGraphQL = "https://api.github.com/graphql"

const prQuery = `
query GetUserPRs($username: String!, $first: Int!, $after: String) {
  user(login: $username) {
    pullRequests(first: $first, after: $after, orderBy: {field: CREATED_AT, direction: DESC}) {
      pageInfo { hasNextPage endCursor }
      nodes {
        id number title state
        createdAt mergedAt closedAt isDraft
        repository {
          name
          owner { login }
          stargazerCount
        }
        url
        timelineItems(itemTypes: [CLOSED_EVENT], last: 1) {
          nodes {
            ... on ClosedEvent {
              closer {
                __typename
                ... on PullRequest { state }
              }
            }
          }
        }
      }
    }
  }
}`

type rawPR struct {
	ID        string  `json:"id"`
	Number    int     `json:"number"`
	Title     string  `json:"title"`
	State     string  `json:"state"`
	CreatedAt string  `json:"createdAt"`
	MergedAt  *string `json:"mergedAt"`
	ClosedAt  *string `json:"closedAt"`
	IsDraft   bool    `json:"isDraft"`
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		StargazerCount int `json:"stargazerCount"`
	} `json:"repository"`
	URL           string `json:"url"`
	TimelineItems struct {
		Nodes []struct {
			Closer *struct {
				Typename string `json:"__typename"`
				State    string `json:"state"`
			} `json:"closer"`
		} `json:"nodes"`
	} `json:"timelineItems"`
}

type gqlResponse struct {
	Data struct {
		User *struct {
			PullRequests struct {
				PageInfo struct {
					HasNextPage bool    `json:"hasNextPage"`
					EndCursor   *string `json:"endCursor"`
				} `json:"pageInfo"`
				Nodes []rawPR `json:"nodes"`
			} `json:"pullRequests"`
		} `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func fetchAllPRs(token, username string) ([]rawPR, error) {
	var all []rawPR
	var cursor *string

	for {
		page, hasNext, next, err := fetchPage(token, username, cursor)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		if !hasNext {
			break
		}
		cursor = next
	}

	return all, nil
}

func firstN(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}

func fetchPage(token, username string, after *string) ([]rawPR, bool, *string, error) {
	body := map[string]any{
		"query": prQuery,
		"variables": map[string]any{
			"username": username,
			"first":    100,
			"after":    after,
		},
	}

	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", githubGraphQL, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "pullscape")

	resp, err := githubClient.Do(req)
	if err != nil {
		return nil, false, nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var result gqlResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, false, nil, fmt.Errorf("github returned HTTP %d with non-JSON body: %s", resp.StatusCode, firstN(data, 120))
	}

	if len(result.Errors) > 0 {
		return nil, false, nil, fmt.Errorf("github: %s", result.Errors[0].Message)
	}
	if result.Data.User == nil {
		return nil, false, nil, fmt.Errorf("user %q not found", username)
	}

	prs := result.Data.User.PullRequests
	return prs.Nodes, prs.PageInfo.HasNextPage, prs.PageInfo.EndCursor, nil
}

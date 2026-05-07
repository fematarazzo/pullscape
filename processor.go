package main

import (
	"sort"
	"strings"
	"time"
)

type PR struct {
	Repo        string
	Stars       int
	Title       string
	Number      int
	Status      string
	CreatedDate string
	MergedDate  string
	URL         string
}

type Stats struct {
	TotalPR         int
	MergedPR        int
	DisplayPR       int
	ReposWithPR     int
	ReposWithMerged int
	ShowingRepos    int
}

type RepoAggregate struct {
	Repo       string
	Stars      int
	PRNumbers  []int
	Total      int
	Merged     int
	Open       int
	Draft      int
	Closed     int
	MergedRate int
}

type Params struct {
	Username string
	Theme    string
	Status   string
	MinStars int
	Limit    int
	Sort     string
	Stats    string
	Fields   string
	Mode     string
}

func processPRs(raw []rawPR) []PR {
	prs := make([]PR, 0, len(raw))
	for _, r := range raw {
		status := prStatus(r)
		mergedDate := ""
		if r.MergedAt != nil {
			mergedDate = formatDate(*r.MergedAt)
		} else if (status == "merged" || status == "upstream") && r.ClosedAt != nil {
			mergedDate = formatDate(*r.ClosedAt)
		}
		prs = append(prs, PR{
			Repo:        r.Repository.Owner.Login + "/" + r.Repository.Name,
			Stars:       r.Repository.StargazerCount,
			Title:       r.Title,
			Number:      r.Number,
			Status:      status,
			CreatedDate: formatDate(r.CreatedAt),
			MergedDate:  mergedDate,
			URL:         r.URL,
		})
	}
	return prs
}

func prStatus(r rawPR) string {
	if r.IsDraft {
		return "draft"
	}
	if r.State == "MERGED" {
		return "merged"
	}
	if r.State == "OPEN" {
		return "open"
	}
	if r.State == "CLOSED" {
		switch closerType(r) {
		case "Commit":
			return "merged"
		case "PullRequest":
			return "upstream"
		}
	}
	return "closed"
}

func closerType(r rawPR) string {
	if len(r.TimelineItems.Nodes) > 0 {
		closer := r.TimelineItems.Nodes[0].Closer
		if closer != nil {
			if closer.Typename == "Commit" {
				return "Commit"
			}
			if closer.Typename == "PullRequest" && closer.State == "MERGED" {
				return "PullRequest"
			}
		}
	}
	if mergedExternallyByBot(r) {
		return "PullRequest"
	}
	return ""
}

// mergedExternallyByBot detects projects (Go, FreeBSD, etc.) that merge patches
// via external tools (Gerrit, Phabricator) and have a bot close the GitHub PR
// with a comment confirming the merge.
func mergedExternallyByBot(r rawPR) bool {
	if len(r.Comments.Nodes) == 0 {
		return false
	}
	c := r.Comments.Nodes[0]
	login := strings.ToLower(c.Author.Login)
	body := strings.ToLower(c.BodyText)
	isBot := strings.HasSuffix(login, "[bot]") || login == "gopherbot" || login == "bors"
	hasMergeMsg := strings.Contains(body, "has been merged") || strings.Contains(body, "is being closed because")
	return isBot && hasMergeMsg
}

func formatDate(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("2006-01-02")
}

func filterByStatus(prs []PR, status string) []PR {
	if status == "" || status == "all" {
		return prs
	}
	set := map[string]bool{}
	for _, s := range strings.Split(status, ",") {
		set[strings.TrimSpace(s)] = true
	}
	var result []PR
	for _, pr := range prs {
		if set[pr.Status] {
			result = append(result, pr)
		}
	}
	return result
}

func filterByMinStars(prs []PR, min int) []PR {
	if min <= 0 {
		return prs
	}
	var result []PR
	for _, pr := range prs {
		if pr.Stars >= min {
			result = append(result, pr)
		}
	}
	return result
}

func sortPRs(prs []PR, sortParam string) []PR {
	if sortParam == "" {
		return prs
	}
	fields := strings.Split(sortParam, ",")
	result := make([]PR, len(prs))
	copy(result, prs)
	sort.SliceStable(result, func(i, j int) bool {
		for _, f := range fields {
			cmp := comparePRs(result[i], result[j], strings.TrimSpace(f))
			if cmp != 0 {
				return cmp < 0
			}
		}
		return false
	})
	return result
}

func comparePRs(a, b PR, field string) int {
	switch field {
	case "stars_desc":
		return b.Stars - a.Stars
	case "stars_asc":
		return a.Stars - b.Stars
	case "created_date_desc":
		return strings.Compare(b.CreatedDate, a.CreatedDate)
	case "created_date_asc":
		return strings.Compare(a.CreatedDate, b.CreatedDate)
	case "status":
		return statusPriority(a.Status) - statusPriority(b.Status)
	}
	return 0
}

func statusPriority(s string) int {
	switch s {
	case "merged":
		return 0
	case "upstream":
		return 1
	case "open":
		return 2
	case "draft":
		return 3
	case "closed":
		return 4
	}
	return 5
}

func calcStats(all, display []PR) Stats {
	reposWithPR := map[string]bool{}
	reposWithMerged := map[string]bool{}
	merged := 0
	for _, pr := range all {
		reposWithPR[pr.Repo] = true
		if pr.Status == "merged" || pr.Status == "upstream" {
			merged++
			reposWithMerged[pr.Repo] = true
		}
	}
	showingRepos := map[string]bool{}
	for _, pr := range display {
		showingRepos[pr.Repo] = true
	}
	return Stats{
		TotalPR:         len(all),
		MergedPR:        merged,
		DisplayPR:       len(display),
		ReposWithPR:     len(reposWithPR),
		ReposWithMerged: len(reposWithMerged),
		ShowingRepos:    len(showingRepos),
	}
}

func aggregateByRepo(prs []PR) []RepoAggregate {
	repoMap := map[string]*RepoAggregate{}
	var order []string

	for _, pr := range prs {
		if _, ok := repoMap[pr.Repo]; !ok {
			repoMap[pr.Repo] = &RepoAggregate{Repo: pr.Repo, Stars: pr.Stars}
			order = append(order, pr.Repo)
		}
		r := repoMap[pr.Repo]
		r.PRNumbers = append(r.PRNumbers, pr.Number)
		r.Total++
		switch pr.Status {
		case "merged", "upstream":
			r.Merged++
		case "open":
			r.Open++
		case "draft":
			r.Draft++
		case "closed":
			r.Closed++
		}
	}

	result := make([]RepoAggregate, 0, len(order))
	for _, repo := range order {
		r := repoMap[repo]
		if r.Total > 0 {
			r.MergedRate = r.Merged * 100 / r.Total
		}
		sort.Sort(sort.Reverse(sort.IntSlice(r.PRNumbers)))
		result = append(result, *r)
	}
	return result
}

func sortRepos(repos []RepoAggregate, sortParam string) []RepoAggregate {
	if sortParam == "" {
		return repos
	}
	fields := strings.Split(sortParam, ",")
	result := make([]RepoAggregate, len(repos))
	copy(result, repos)
	sort.SliceStable(result, func(i, j int) bool {
		for _, f := range fields {
			cmp := compareRepos(result[i], result[j], strings.TrimSpace(f))
			if cmp != 0 {
				return cmp < 0
			}
		}
		return false
	})
	return result
}

func compareRepos(a, b RepoAggregate, field string) int {
	switch field {
	case "stars_desc":
		return b.Stars - a.Stars
	case "stars_asc":
		return a.Stars - b.Stars
	case "merged_desc":
		return b.Merged - a.Merged
	case "merged_asc":
		return a.Merged - b.Merged
	case "merged_rate_desc":
		return b.MergedRate - a.MergedRate
	case "merged_rate_asc":
		return a.MergedRate - b.MergedRate
	}
	return 0
}

func processData(raw []rawPR, params Params) ([]PR, Stats, []RepoAggregate) {
	prs := processPRs(raw)
	all := make([]PR, len(prs))
	copy(all, prs)

	prs = filterByMinStars(prs, params.MinStars)

	if params.Mode == "repo-aggregate" {
		repos := aggregateByRepo(prs)
		repos = sortRepos(repos, orDefault(params.Sort, "merged_desc,stars_desc"))
		limit := orDefaultInt(params.Limit, 10)
		if limit > 0 && limit < len(repos) {
			repos = repos[:limit]
		}
		displayRepoSet := map[string]bool{}
		for _, r := range repos {
			displayRepoSet[r.Repo] = true
		}
		var displayPRs []PR
		for _, pr := range prs {
			if displayRepoSet[pr.Repo] {
				displayPRs = append(displayPRs, pr)
			}
		}
		return nil, calcStats(all, displayPRs), repos
	}

	prs = filterByStatus(prs, orDefault(params.Status, "all"))
	prs = sortPRs(prs, orDefault(params.Sort, "status,stars_desc"))
	limit := orDefaultInt(params.Limit, 10)
	if limit > 0 && limit < len(prs) {
		prs = prs[:limit]
	}
	return prs, calcStats(all, prs), nil
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func orDefaultInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}

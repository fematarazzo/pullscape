package main

import (
	"bytes"
	"fmt"
	"html"
	"strconv"
	"strings"
)

type theme struct {
	bg, titleColor, headerBg, headerColor string
	borderColor, textColor                string
	rowEvenBg, rowOddBg                   string
	statsBg, statsColor, mergedColor      string
}

func getTheme(name string) theme {
	if name == "light" {
		return theme{
			bg: "#ffffff", titleColor: "#24292f",
			headerBg: "#f6f8fa", headerColor: "#24292f",
			borderColor: "#d0d7de", textColor: "#24292f",
			rowEvenBg: "#ffffff", rowOddBg: "#f6f8fa",
			statsBg: "#f6f8fa", statsColor: "#57606a",
			mergedColor: "rgba(130,80,223,1)",
		}
	}
	return theme{
		bg: "#0d1117", titleColor: "#f0f6fc",
		headerBg: "#21262d", headerColor: "#c9d1d9",
		borderColor: "#30363d", textColor: "#e6edf3",
		rowEvenBg: "#0d1117", rowOddBg: "#161b22",
		statsBg: "#161b22", statsColor: "#8b949e",
		mergedColor: "rgba(130,80,223,1)",
	}
}

type col struct {
	key   string
	label string
	width int
	align string
}

var prListCols = []col{
	{"repo", "Repository", 200, "left"},
	{"stars", "Stars", 80, "right"},
	{"pr_title", "PR Title", 250, "left"},
	{"pr_number", "PR #", 60, "center"},
	{"status", "Status", 80, "center"},
	{"created_date", "Created", 90, "center"},
	{"merged_date", "Merged", 90, "center"},
}

var repoCols = []col{
	{"repo", "Repository", 200, "left"},
	{"stars", "Stars", 80, "right"},
	{"pr_numbers", "PRs", 150, "left"},
	{"total", "Total", 60, "center"},
	{"merged", "Merged", 60, "center"},
	{"open", "Open", 50, "center"},
	{"draft", "Draft", 50, "center"},
	{"closed", "Closed", 60, "center"},
	{"merged_rate", "Merge Rate", 90, "center"},
}

var statusColors = map[string]string{
	"merged": "rgba(130,80,223,1)",
	"open":   "rgba(26,127,55,1)",
	"draft":  "rgba(89,99,110,1)",
	"closed": "rgba(209,36,47,1)",
}

var statusIcons = map[string]string{
	"merged": `<path d="M5.45 5.154A4.25 4.25 0 0 0 9.25 7.5h1.378a2.251 2.251 0 1 1 0 1.5H9.25A5.734 5.734 0 0 1 5 7.123v3.505a2.25 2.25 0 1 1-1.5 0V5.372a2.25 2.25 0 1 1 1.95-.218ZM4.25 13.5a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Zm8.5-4.5a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5ZM5 3.25a.75.75 0 1 0 0 .005V3.25Z"/>`,
	"open":   `<path d="M1.5 3.25a2.25 2.25 0 1 1 3 2.122v5.256a2.251 2.251 0 1 1-1.5 0V5.372A2.25 2.25 0 0 1 1.5 3.25Zm5.677-.177L9.573.677A.25.25 0 0 1 10 .854V2.5h1A2.5 2.5 0 0 1 13.5 5v5.628a2.251 2.251 0 1 1-1.5 0V5a1 1 0 0 0-1-1h-1v1.646a.25.25 0 0 1-.427.177L7.177 3.427a.25.25 0 0 1 0-.354ZM3.75 2.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5Zm0 9.5a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5Zm8.25.75a.75.75 0 1 0 1.5 0 .75.75 0 0 0-1.5 0Z"/>`,
	"draft":  `<path d="M3.25 1A2.25 2.25 0 0 1 4 5.372v5.256a2.251 2.251 0 1 1-1.5 0V5.372A2.25 2.25 0 0 1 3.25 1Zm9.5 14a2.25 2.25 0 1 1 0-4.5 2.25 2.25 0 0 1 0 4.5ZM2.5 3.25a.75.75 0 1 0 1.5 0 .75.75 0 0 0-1.5 0ZM3.25 12a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5Zm9.5 0a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5ZM9.5 7.25a1.75 1.75 0 1 1-3.5 0 1.75 1.75 0 0 1 3.5 0Z"/>`,
	"closed": `<path d="M3.25 1A2.25 2.25 0 0 1 4 5.372v5.256a2.251 2.251 0 1 1-1.5 0V5.372A2.25 2.25 0 0 1 3.25 1Zm9.5 5.5a.75.75 0 0 1 .75.75v3.378a2.251 2.251 0 1 1-1.5 0V7.25a.75.75 0 0 1 .75-.75Zm-2.03-4.53a.75.75 0 0 1 1.06 0l.97.97.97-.97a.749.749 0 0 1 1.06 1.06l-.97.97.97.97a.749.749 0 1 1-1.06 1.06l-.97-.97-.97.97a.749.749 0 1 1-1.06-1.06l.97-.97-.97-.97a.75.75 0 0 1 0-1.06ZM2.5 3.25a.75.75 0 1 0 1.5 0 .75.75 0 0 0-1.5 0ZM3.25 12a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5Zm9.5 0a.75.75 0 1 0 0 1.5.75.75 0 0 0 0-1.5Z"/>`,
}

const (
	rowH    = 35
	headerH = 40
	titleH  = 50
	pad     = 10
)

func generateSVG(username string, prs []PR, stats Stats, params Params, repos []RepoAggregate) string {
	t := getTheme(params.Theme)
	if params.Mode == "repo-aggregate" && repos != nil {
		return repoView(username, repos, stats, params, t)
	}
	return prListView(username, prs, stats, params, t)
}

func activeCols(param, defaultCols string, available []col) []col {
	if param == "" {
		param = defaultCols
	}
	set := map[string]bool{}
	for _, k := range strings.Split(param, ",") {
		set[strings.TrimSpace(k)] = true
	}
	var result []col
	for _, c := range available {
		if set[c.key] {
			result = append(result, c)
		}
	}
	if len(result) == 0 {
		return available
	}
	return result
}

func tableWidth(cols []col) int {
	w := 1
	for _, c := range cols {
		w += c.width + 1
	}
	return w
}

type statItem struct {
	label string
	value string
	color string
}

func buildStatItems(stats Stats, param, textColor, mergedColor string) []statItem {
	keys := strings.Split(orDefault(param, "all"), ",")
	if len(keys) == 1 && (keys[0] == "all" || keys[0] == "") {
		return []statItem{
			{"Total PRs", strconv.Itoa(stats.TotalPR), textColor},
			{"Merged PRs", strconv.Itoa(stats.MergedPR), mergedColor},
			{"Displayed", strconv.Itoa(stats.DisplayPR), textColor},
		}
	}
	lookup := map[string]statItem{
		"total_pr":             {"Total PRs", strconv.Itoa(stats.TotalPR), textColor},
		"merged_pr":            {"Merged PRs", strconv.Itoa(stats.MergedPR), mergedColor},
		"display_pr":           {"Displayed PRs", strconv.Itoa(stats.DisplayPR), textColor},
		"repos_with_pr":        {"Repos", strconv.Itoa(stats.ReposWithPR), textColor},
		"repos_with_merged_pr": {"Repos w/ Merges", strconv.Itoa(stats.ReposWithMerged), mergedColor},
		"showing_repos":        {"Showing Repos", strconv.Itoa(stats.ShowingRepos), textColor},
	}
	var result []statItem
	for _, k := range keys {
		if item, ok := lookup[strings.TrimSpace(k)]; ok {
			result = append(result, item)
		}
	}
	return result
}

func prListView(username string, prs []PR, stats Stats, params Params, t theme) string {
	cols := activeCols(params.Fields, "repo,stars,pr_title,pr_number,status,created_date,merged_date", prListCols)
	items := buildStatItems(stats, params.Stats, t.statsColor, t.mergedColor)

	tw := tableWidth(cols)
	statsH, effectiveW := statsLayout(len(items), tw)
	statsMargin := 0
	if statsH > 0 {
		statsMargin = 10
	}

	totalH := titleH + statsH + statsMargin + headerH + len(prs)*rowH + 20

	var b bytes.Buffer
	svgOpen(&b, effectiveW, totalH, t.bg)

	y := 20
	writeTitle(&b, username, effectiveW, y, t)
	y += titleH

	if statsH > 0 {
		writeStats(&b, items, effectiveW, y, t)
		y += statsH + statsMargin
	}

	tableX := (effectiveW - tw) / 2
	writeHeader(&b, cols, y, tableX, t)
	y += headerH

	for i, pr := range prs {
		writePRRow(&b, pr, cols, i, y, tableX, t)
		y += rowH
	}

	b.WriteString("</svg>")
	return b.String()
}

func repoView(username string, repos []RepoAggregate, stats Stats, params Params, t theme) string {
	cols := activeCols(params.Fields, "repo,stars,pr_numbers,total,merged,open,draft,closed,merged_rate", repoCols)
	items := buildStatItems(stats, params.Stats, t.statsColor, t.mergedColor)

	tw := tableWidth(cols)
	statsH, effectiveW := statsLayout(len(items), tw)
	statsMargin := 0
	if statsH > 0 {
		statsMargin = 10
	}

	totalH := titleH + statsH + statsMargin + headerH + len(repos)*rowH + 20

	var b bytes.Buffer
	svgOpen(&b, effectiveW, totalH, t.bg)

	y := 20
	writeTitle(&b, username, effectiveW, y, t)
	y += titleH

	if statsH > 0 {
		writeStats(&b, items, effectiveW, y, t)
		y += statsH + statsMargin
	}

	tableX := (effectiveW - tw) / 2
	writeHeader(&b, cols, y, tableX, t)
	y += headerH

	for i, repo := range repos {
		writeRepoRow(&b, repo, cols, i, y, tableX, t)
		y += rowH
	}

	b.WriteString("</svg>")
	return b.String()
}

func statsLayout(itemCount, tableW int) (height, width int) {
	if itemCount == 0 {
		return 0, tableW
	}
	w := itemCount*140 + pad*2
	if w < tableW {
		w = tableW
	}
	return 40, w
}

func svgOpen(b *bytes.Buffer, w, h int, bg string) {
	fmt.Fprintf(b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">`, w, h)
	fmt.Fprintf(b, `<rect width="%d" height="%d" fill="%s" rx="6"/>`, w, h, bg)
	b.WriteString(`<style>text{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",monospace;font-size:12px}</style>`)
}

func writeTitle(b *bytes.Buffer, username string, w, y int, t theme) {
	fmt.Fprintf(b, `<text x="%d" y="%d" text-anchor="middle" font-size="15" font-weight="600" fill="%s">%s's PR Stats</text>`,
		w/2, y+25, t.titleColor, html.EscapeString(username))
}

func writeStats(b *bytes.Buffer, items []statItem, w, y int, t theme) {
	fmt.Fprintf(b, `<rect x="0" y="%d" width="%d" height="40" fill="%s"/>`, y, w, t.statsBg)
	sectionW := w / len(items)
	for i, item := range items {
		cx := i*sectionW + sectionW/2
		fmt.Fprintf(b, `<text x="%d" y="%d" text-anchor="middle" font-size="11" fill="%s">%s: <tspan fill="%s" font-weight="600">%s</tspan></text>`,
			cx, y+25, t.statsColor, html.EscapeString(item.label), item.color, html.EscapeString(item.value))
	}
}

func writeHeader(b *bytes.Buffer, cols []col, y, tableX int, t theme) {
	tw := tableWidth(cols)
	fmt.Fprintf(b, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>`, tableX, y, tw, headerH, t.headerBg)

	x := tableX + 1
	for _, c := range cols {
		tx := alignX(x, c.width, c.align)
		fmt.Fprintf(b, `<text x="%d" y="%d" fill="%s" text-anchor="%s" font-weight="600">%s</text>`,
			tx, y+headerH/2+5, t.headerColor, svgAnchor(c.align), html.EscapeString(c.label))
		x += c.width + 1
	}

	drawBorders(b, cols, y, headerH, tableX, t.borderColor)
}

func writePRRow(b *bytes.Buffer, pr PR, cols []col, rowIdx, y, tableX int, t theme) {
	tw := tableWidth(cols)
	bg := t.rowEvenBg
	if rowIdx%2 == 1 {
		bg = t.rowOddBg
	}
	fmt.Fprintf(b, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>`, tableX, y, tw, rowH, bg)

	x := tableX + 1
	for _, c := range cols {
		cy := y + rowH/2 + 5
		switch c.key {
		case "status":
			writeIcon(b, pr.Status, x+c.width/2-7, y+rowH/2-7)
		case "pr_title":
			clip := fmt.Sprintf("c%d%s", rowIdx, c.key)
			fmt.Fprintf(b, `<clipPath id="%s"><rect x="%d" y="%d" width="%d" height="%d"/></clipPath>`, clip, x, y, c.width-pad, rowH)
			fmt.Fprintf(b, `<a href="%s"><text x="%d" y="%d" fill="%s" clip-path="url(#%s)">%s</text></a>`,
				html.EscapeString(pr.URL), x+pad, cy, t.textColor, clip, html.EscapeString(pr.Title))
		case "repo":
			clip := fmt.Sprintf("c%d%s", rowIdx, c.key)
			fmt.Fprintf(b, `<clipPath id="%s"><rect x="%d" y="%d" width="%d" height="%d"/></clipPath>`, clip, x, y, c.width-pad, rowH)
			fmt.Fprintf(b, `<text x="%d" y="%d" fill="%s" clip-path="url(#%s)">%s</text>`,
				x+pad, cy, t.textColor, clip, html.EscapeString(pr.Repo))
		default:
			tx := alignX(x, c.width, c.align)
			fmt.Fprintf(b, `<text x="%d" y="%d" fill="%s" text-anchor="%s">%s</text>`,
				tx, cy, t.textColor, svgAnchor(c.align), html.EscapeString(prCell(pr, c.key)))
		}
		x += c.width + 1
	}

	drawBorders(b, cols, y, rowH, tableX, t.borderColor)
}

func writeRepoRow(b *bytes.Buffer, repo RepoAggregate, cols []col, rowIdx, y, tableX int, t theme) {
	tw := tableWidth(cols)
	bg := t.rowEvenBg
	if rowIdx%2 == 1 {
		bg = t.rowOddBg
	}
	fmt.Fprintf(b, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>`, tableX, y, tw, rowH, bg)

	x := tableX + 1
	for _, c := range cols {
		cy := y + rowH/2 + 5
		color := t.textColor
		if c.key == "merged" {
			color = t.mergedColor
		}
		text := repoCell(repo, c.key)

		if c.key == "repo" || c.key == "pr_numbers" {
			clip := fmt.Sprintf("rc%d%s", rowIdx, c.key)
			fmt.Fprintf(b, `<clipPath id="%s"><rect x="%d" y="%d" width="%d" height="%d"/></clipPath>`, clip, x, y, c.width-pad, rowH)
			fmt.Fprintf(b, `<text x="%d" y="%d" fill="%s" clip-path="url(#%s)">%s</text>`,
				x+pad, cy, color, clip, html.EscapeString(text))
		} else {
			tx := alignX(x, c.width, c.align)
			fmt.Fprintf(b, `<text x="%d" y="%d" fill="%s" text-anchor="%s">%s</text>`,
				tx, cy, color, svgAnchor(c.align), html.EscapeString(text))
		}
		x += c.width + 1
	}

	drawBorders(b, cols, y, rowH, tableX, t.borderColor)
}

func drawBorders(b *bytes.Buffer, cols []col, y, h, tableX int, color string) {
	tw := tableWidth(cols)
	fmt.Fprintf(b, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="1"/>`,
		tableX, y+h, tableX+tw, y+h, color)
	x := tableX
	for _, c := range cols {
		fmt.Fprintf(b, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="1"/>`,
			x, y, x, y+h, color)
		x += c.width + 1
	}
	fmt.Fprintf(b, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="1"/>`,
		x, y, x, y+h, color)
}

func writeIcon(b *bytes.Buffer, status string, x, y int) {
	path, ok := statusIcons[status]
	if !ok {
		return
	}
	fmt.Fprintf(b, `<svg x="%d" y="%d" width="14" height="14" viewBox="0 0 16 16" fill="%s">%s</svg>`,
		x, y, statusColors[status], path)
}

func prCell(pr PR, key string) string {
	switch key {
	case "stars":
		return strconv.Itoa(pr.Stars)
	case "pr_number":
		return "#" + strconv.Itoa(pr.Number)
	case "created_date":
		return pr.CreatedDate
	case "merged_date":
		return pr.MergedDate
	}
	return ""
}

func repoCell(r RepoAggregate, key string) string {
	switch key {
	case "repo":
		return r.Repo
	case "stars":
		return strconv.Itoa(r.Stars)
	case "pr_numbers":
		nums := make([]string, len(r.PRNumbers))
		for i, n := range r.PRNumbers {
			nums[i] = "#" + strconv.Itoa(n)
		}
		return strings.Join(nums, ", ")
	case "total":
		return strconv.Itoa(r.Total)
	case "merged":
		return strconv.Itoa(r.Merged)
	case "open":
		return strconv.Itoa(r.Open)
	case "draft":
		return strconv.Itoa(r.Draft)
	case "closed":
		return strconv.Itoa(r.Closed)
	case "merged_rate":
		return strconv.Itoa(r.MergedRate) + "%"
	}
	return ""
}

func alignX(x, width int, align string) int {
	switch align {
	case "center":
		return x + width/2
	case "right":
		return x + width - pad
	default:
		return x + pad
	}
}

func svgAnchor(align string) string {
	switch align {
	case "center":
		return "middle"
	case "right":
		return "end"
	default:
		return "start"
	}
}

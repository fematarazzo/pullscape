package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type cacheEntry struct {
	svg string
	at  time.Time
}

var (
	svgCache   sync.Map
	cacheTTL   = 12 * time.Hour
	refreshing sync.Map
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	http.HandleFunc("/api/github-pr-stats", handler)
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		http.Error(w, "GITHUB_TOKEN not set", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query()
	username := q.Get("username")
	if username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	params := buildParams(q)
	cacheKey := r.URL.RawQuery

	// Memory cache never expires entries. Serve immediately; refresh in
	// background when older than cacheTTL, so responses are sub-ms regardless
	// of upstream GitHub latency. Keeps GitHub Camo's ~10s fetch timeout
	// from ever being an issue.
	if v, ok := svgCache.Load(cacheKey); ok {
		entry := v.(cacheEntry)
		age := time.Since(entry.at)
		source := "memory"
		if age >= cacheTTL {
			source = "memory-stale"
			go refresh(token, cacheKey, params)
		}
		log.Printf("status:200 source:%s user:%s age:%s time:%.3fs", source, username, age.Round(time.Second), time.Since(start).Seconds())
		writeSVG(w, entry.svg)
		return
	}

	// Cold miss: synchronous fetch from GitHub.
	svg, err := fetchAndRender(token, username, params)
	if err != nil {
		log.Printf("status:500 source:github user:%s time:%.3fs err:%s", username, time.Since(start).Seconds(), err)
		svgError(w, err.Error(), strings.Contains(err.Error(), "not found"))
		return
	}

	svgCache.Store(cacheKey, cacheEntry{svg: svg, at: time.Now()})
	log.Printf("status:200 source:github user:%s time:%.3fs", username, time.Since(start).Seconds())
	writeSVG(w, svg)
}

func refresh(token, cacheKey string, params Params) {
	if _, busy := refreshing.LoadOrStore(cacheKey, true); busy {
		return
	}
	defer refreshing.Delete(cacheKey)

	svg, err := fetchAndRender(token, params.Username, params)
	if err != nil {
		log.Printf("refresh failed user:%s err:%s", params.Username, err)
		return
	}
	svgCache.Store(cacheKey, cacheEntry{svg: svg, at: time.Now()})
	log.Printf("refresh ok user:%s", params.Username)
}

func fetchAndRender(token, username string, params Params) (string, error) {
	raw, err := fetchAllPRs(token, username)
	if err != nil {
		return "", err
	}
	prs, stats, repos := processData(raw, params)
	return generateSVG(username, prs, stats, params, repos), nil
}

func buildParams(q url.Values) Params {
	return Params{
		Username: q.Get("username"),
		Theme:    orDefault(q.Get("theme"), "dark"),
		Status:   q.Get("status"),
		MinStars: parseInt(q.Get("min_stars")),
		Limit:    parseInt(q.Get("limit")),
		Sort:     q.Get("sort"),
		Stats:    q.Get("stats"),
		Fields:   q.Get("fields"),
		Mode:     q.Get("mode"),
	}
}

func writeSVG(w http.ResponseWriter, svg string) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400, stale-while-revalidate=2592000")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, svg)
}

func svgError(w http.ResponseWriter, msg string, _ bool) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-store")
	fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="400" height="60"><rect width="400" height="60" fill="#0d1117" rx="6"/><text x="20" y="35" font-family="monospace" font-size="13" fill="#e6edf3">Error: %s</text></svg>`, html.EscapeString(msg))
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

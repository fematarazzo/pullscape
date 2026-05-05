package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
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
	svgCache sync.Map
	cacheTTL = 5 * time.Minute
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/api/github-pr-stats", handler)
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
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

	params := Params{
		Username: username,
		Theme:    orDefault(q.Get("theme"), "dark"),
		Status:   q.Get("status"),
		MinStars: parseInt(q.Get("min_stars")),
		Limit:    parseInt(q.Get("limit")),
		Sort:     q.Get("sort"),
		Stats:    q.Get("stats"),
		Fields:   q.Get("fields"),
		Mode:     q.Get("mode"),
	}

	cacheKey := r.URL.RawQuery
	if v, ok := svgCache.Load(cacheKey); ok {
		entry := v.(cacheEntry)
		if time.Since(entry.at) < cacheTTL {
			w.Header().Set("Content-Type", "image/svg+xml")
			w.Header().Set("Cache-Control", "public, max-age=300")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			fmt.Fprint(w, entry.svg)
			return
		}
	}

	raw, err := fetchAllPRs(token, username)
	if err != nil {
		svgError(w, err.Error(), strings.Contains(err.Error(), "not found"))
		return
	}

	prs, stats, repos := processData(raw, params)
	svg := generateSVG(username, prs, stats, params, repos)

	svgCache.Store(cacheKey, cacheEntry{svg: svg, at: time.Now()})

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, svg)
}

func svgError(w http.ResponseWriter, msg string, notFound bool) {
	status := http.StatusInternalServerError
	if notFound {
		status = http.StatusNotFound
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(status)
	fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="400" height="60"><rect width="400" height="60" fill="#0d1117" rx="6"/><text x="20" y="35" font-family="monospace" font-size="13" fill="#e6edf3">Error: %s</text></svg>`, html.EscapeString(msg))
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

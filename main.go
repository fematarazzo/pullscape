package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
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

	raw, err := fetchAllPRs(token, username)
	if err != nil {
		svgError(w, err.Error(), strings.Contains(err.Error(), "not found"))
		return
	}

	prs, stats, repos := processData(raw, params)
	svg := generateSVG(username, prs, stats, params, repos)

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache, must-revalidate")
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

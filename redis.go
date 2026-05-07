package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"
)

type redisClient struct {
	url   string
	token string
	http  *http.Client
}

func newRedisClient() *redisClient {
	url := os.Getenv("UPSTASH_REDIS_REST_URL")
	token := os.Getenv("UPSTASH_REDIS_REST_TOKEN")
	if url == "" || token == "" {
		return nil
	}
	return &redisClient{
		url:   url,
		token: token,
		http:  &http.Client{Timeout: 2 * time.Second},
	}
}

func (c *redisClient) get(key string) (string, bool) {
	body, _ := json.Marshal([]string{"GET", key})
	req, _ := http.NewRequest("POST", c.url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var result struct {
		Result *string `json:"result"`
	}
	if json.Unmarshal(data, &result) != nil || result.Result == nil {
		return "", false
	}
	return *result.Result, true
}

func (c *redisClient) set(key, value string, ttl time.Duration) {
	body, _ := json.Marshal([]any{"SET", key, value, "EX", int(ttl.Seconds())})
	req, _ := http.NewRequest("POST", c.url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
		http:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *redisClient) ping() error {
	body, _ := json.Marshal([]string{"PING"})
	req, _ := http.NewRequest("POST", c.url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	// First call to a Fly-managed Upstash over 6PN can be slow when the
	// endpoint is idle; give the startup ping its own generous timeout.
	pinger := &http.Client{Timeout: 15 * time.Second}
	resp, err := pinger.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, firstN(data, 200))
	}
	return nil
}

func (c *redisClient) get(key string) (string, bool) {
	body, _ := json.Marshal([]string{"GET", key})
	req, _ := http.NewRequest("POST", c.url, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		log.Printf("redis get failed: %v", err)
		return "", false
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Printf("redis get: status %d body:%s", resp.StatusCode, firstN(data, 200))
		return "", false
	}
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
		log.Printf("redis set failed: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		log.Printf("redis set: status %d body:%s", resp.StatusCode, firstN(data, 200))
	}
}

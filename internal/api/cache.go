package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type CacheEntry struct {
	Response  *SearchResponse `json:"response"`
	Timestamp time.Time       `json:"timestamp"`
}

type Cache struct {
	path    string
	entries map[string]CacheEntry
	mu      sync.RWMutex
}

func NewCache(path string) *Cache {
	c := &Cache{
		path:    path,
		entries: make(map[string]CacheEntry),
	}
	c.load()
	return c
}

func (c *Cache) Get(key string, ttl time.Duration) (*SearchResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	if time.Since(entry.Timestamp) > ttl {
		return nil, false
	}

	return entry.Response, true
}

func (c *Cache) Set(key string, resp *SearchResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = CacheEntry{
		Response:  resp,
		Timestamp: time.Now(),
	}
	c.save()
}

func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]CacheEntry)
	return os.Remove(c.path)
}

func (c *Cache) load() {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return
	}
	json.Unmarshal(data, &c.entries)
}

func (c *Cache) save() {
	os.MkdirAll(filepath.Dir(c.path), 0755)
	data, _ := json.Marshal(c.entries)
	os.WriteFile(c.path, data, 0644)
}

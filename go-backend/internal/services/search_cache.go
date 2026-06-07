package services

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type cacheEntry struct {
	results []GraphRAGResult
	expiry  time.Time
}

// SearchCache provides a thread-safe TTL cache for GraphRAG search results.
// It uses lazy eviction (expired entries cleaned on access) instead of a
// background goroutine, making it safe for tests.
type SearchCache struct {
	mu      sync.RWMutex
	store   map[string]cacheEntry
	ttl     time.Duration
	maxSize int
}

func NewSearchCache(ttl time.Duration, maxSize int) *SearchCache {
	return &SearchCache{
		store:   make(map[string]cacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

func cacheKey(userID uint, query string, limit int) string {
	return fmt.Sprintf("%d:%s:%d", userID, query, limit)
}

func (c *SearchCache) Get(userID uint, query string, limit int) ([]GraphRAGResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(userID, query, limit)
	entry, ok := c.store[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiry) {
		delete(c.store, key)
		return nil, false
	}
	return entry.results, true
}

func (c *SearchCache) Set(userID uint, query string, limit int, results []GraphRAGResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.store) >= c.maxSize {
		c.evictLocked()
	}

	key := cacheKey(userID, query, limit)
	c.store[key] = cacheEntry{
		results: results,
		expiry:  time.Now().Add(c.ttl),
	}
}

// Invalidate removes all cached entries for a given user.
func (c *SearchCache) Invalidate(userID uint) {
	c.mu.Lock()
	defer c.mu.Unlock()

	prefix := fmt.Sprintf("%d:", userID)
	for k := range c.store {
		if strings.HasPrefix(k, prefix) {
			delete(c.store, k)
		}
	}
}

// evictLocked removes expired entries and, if still over capacity,
// evicts the entry with the earliest expiry. Must be called with mu held.
func (c *SearchCache) evictLocked() {
	now := time.Now()
	for k, v := range c.store {
		if now.After(v.expiry) {
			delete(c.store, k)
		}
	}
	if len(c.store) >= c.maxSize {
		oldest := ""
		earliestExpiry := time.Time{}
		for k, v := range c.store {
			if oldest == "" || v.expiry.Before(earliestExpiry) {
				oldest = k
				earliestExpiry = v.expiry
			}
		}
		if oldest != "" {
			delete(c.store, oldest)
		}
	}
}

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MemoryCache provides in-memory caching with TTL as a fallback when Redis is unavailable.
// Suitable for single-process deployments. For multi-process, configure Redis.
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
}

type cacheItem struct {
	value     []byte
	expiresAt time.Time
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		items: make(map[string]*cacheItem),
	}
}

func (c *MemoryCache) Get(ctx context.Context, key string, dest interface{}) error {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		return fmt.Errorf("cache miss: %s", key)
	}
	if time.Now().After(item.expiresAt) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return fmt.Errorf("cache expired: %s", key)
	}
	return json.Unmarshal(item.value, dest)
}

func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	c.mu.Lock()
	c.items[key] = &cacheItem{
		value:     data,
		expiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()
	return nil
}

func (c *MemoryCache) Delete(ctx context.Context, key ...string) error {
	c.mu.Lock()
	for _, k := range key {
		delete(c.items, k)
	}
	c.mu.Unlock()
	return nil
}

func (c *MemoryCache) Invalidate(ctx context.Context, pattern string) error {
	c.mu.Lock()
	for k := range c.items {
		if matchPattern(k, pattern) {
			delete(c.items, k)
		}
	}
	c.mu.Unlock()
	return nil
}

// Simple glob pattern matching (* wildcard)
func matchPattern(s, pattern string) bool {
	if pattern == "*" {
		return true
	}
	// Support "prefix:*" pattern
	if len(pattern) > 1 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(s) >= len(prefix) && s[:len(prefix)] == prefix
	}
	return s == pattern
}

// Cache keys for different entities
const (
	ProductListKey   = "products:list:%s"
	ProductDetailKey = "product:%s"
	CategoryListKey  = "categories:%s"
	PostListKey      = "posts:list:%s"
	PostDetailKey    = "post:%s"
	EventListKey     = "events:list:%s"
)

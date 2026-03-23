package cache

import (
	"sync"
	"time"
)

// Item represents a cached item
type Item struct {
	Value      interface{}
	Expiration time.Time
}

// IsExpired checks if item is expired
func (i *Item) IsExpired() bool {
	if i.Expiration.IsZero() {
		return false
	}
	return time.Now().After(i.Expiration)
}

// Cache represents an in-memory cache
type Cache struct {
	mu         sync.RWMutex
	items      map[string]*Item
	defaultTTL time.Duration
}

// New creates a new cache instance
func New(defaultTTL time.Duration) *Cache {
	c := &Cache{
		items:      make(map[string]*Item),
		defaultTTL: defaultTTL,
	}

	// Start cleanup goroutine
	go c.cleanupLoop()

	return c
}

// Set adds an item to cache with default TTL
func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL adds an item to cache with custom TTL
func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &Item{
		Value:      value,
		Expiration: time.Now().Add(ttl),
	}
}

// Get retrieves an item from cache
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if item.IsExpired() {
		return nil, false
	}

	return item.Value, true
}

// Delete removes an item from cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear removes all items from cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*Item)
}

// Count returns number of items in cache
func (c *Cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// cleanupLoop periodically removes expired items
func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.deleteExpired()
	}
}

// deleteExpired removes all expired items
func (c *Cache) deleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.Expiration) {
			delete(c.items, key)
		}
	}
}

// Global cache instance
var global *Cache

// Global returns global cache instance
func Global() *Cache {
	if global == nil {
		global = New(5 * time.Minute)
	}
	return global
}

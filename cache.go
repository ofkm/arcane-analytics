package main

import (
	"sync"
	"time"
)

type CacheEntry[T any] struct {
	data      T
	timestamp time.Time
}

type TTLCache[T any] struct {
	entries map[string]CacheEntry[T]
	mutex   sync.RWMutex
	ttl     time.Duration
}

func NewTTLCache[T any](ttl time.Duration) *TTLCache[T] {
	return &TTLCache[T]{
		entries: make(map[string]CacheEntry[T]),
		ttl:     ttl,
	}
}

func (c *TTLCache[T]) Get(key string) (T, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.entries[key]
	if !exists || time.Since(entry.timestamp) > c.ttl {
		var zero T
		return zero, false
	}

	return entry.data, true
}

func (c *TTLCache[T]) Set(key string, value T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries[key] = CacheEntry[T]{
		data:      value,
		timestamp: time.Now(),
	}
}

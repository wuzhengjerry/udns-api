// Package cache 缓存模块
package cache

import (
	"context"
	"time"
)

var (
	cache Cache
)

// C 全局缓存对象, 默认使用
func C() Cache {
	if cache == nil {
		panic("global cache instance is nil")
	}
	return cache
}

// SetGlobal 设置全局缓存
func SetGlobal(c Cache) {
	cache = c
}

// Cache provides the interface for cache implementations.
type Cache interface {
	// SetDefaultTTL SetTTL set default ttl
	SetDefaultTTL(ttl time.Duration)
	// Put set cached value with key and expire time.
	Put(key string, val interface{}) error
	// PutWithTTL set cached value with key and expire time.
	PutWithTTL(key string, val interface{}, ttl time.Duration) error
	// Get get cached value by key.
	Get(key string, val interface{}) error
	// Delete delete cached value by key.
	Delete(key string) error
	// IsExist check if cached value exists or not.
	IsExist(key string) bool
	// ClearAll clear all cache.
	ClearAll() error
	// Incr increase cached int value by key, as a counter.
	Incr(key string) error
	// Decr decrease cached int value by key, as a counter.
	Decr(key string) error
	// Close close cache
	Close() error
	// WithContext 携带上下文
	WithContext(ctx context.Context) Cache
}

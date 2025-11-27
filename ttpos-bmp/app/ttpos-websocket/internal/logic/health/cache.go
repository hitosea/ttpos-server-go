// Package health 健康检查逻辑层
package health

import (
	"sync"
	"time"
)

// ResultCache 结果缓存
type ResultCache struct {
	mu     sync.RWMutex
	result *HealthResponse
	expiry time.Time
	ttl    time.Duration
}

// NewResultCache 创建结果缓存
func NewResultCache(ttl time.Duration) *ResultCache {
	return &ResultCache{
		ttl: ttl,
	}
}

// Get 获取缓存结果
func (c *ResultCache) Get() (*HealthResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.result == nil || time.Now().After(c.expiry) {
		return nil, false
	}

	return c.result, true
}

// Set 设置缓存结果
func (c *ResultCache) Set(result *HealthResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.result = result
	c.expiry = time.Now().Add(c.ttl)
}

// Clear 清空缓存
func (c *ResultCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.result = nil
	c.expiry = time.Time{}
}

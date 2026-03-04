package cache

import (
	"context"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
)

type goCache struct {
	cache *gocache.Cache
}

func newGoCache(_ Config) Cache {
	// 创建一个默认过期时间为5分钟，清理间隔为10分钟的缓存
	c := gocache.New(5*time.Minute, 10*time.Minute)
	return &goCache{
		cache: c,
	}
}

func (c *goCache) Set(key string, value interface{}, expiration time.Duration) error {
	c.cache.Set(key, value, expiration)
	return nil
}

func (c *goCache) Get(key string) (interface{}, bool) {
	return c.cache.Get(key)
}

func (c *goCache) GetBytes(key string) ([]byte, bool) {
	return nil, false
}

func (c *goCache) GetBatchBytes(keys []string) (map[string][]byte, []string) {
	result := make(map[string][]byte)
	var missedKeys []string

	for _, key := range keys {
		if val, found := c.cache.Get(key); found {
			switch v := val.(type) {
			case []byte:
				result[key] = v
			case string:
				result[key] = []byte(v)
			default:
				missedKeys = append(missedKeys, key)
			}
		} else {
			missedKeys = append(missedKeys, key)
		}
	}

	return result, missedKeys
}

func (c *goCache) Del(keys ...string) {
	for _, key := range keys {
		c.cache.Delete(key)
	}
}

func (c *goCache) GetClient() *redis.Client {
	return nil
}

func (c *goCache) GetClusterClient() *redis.ClusterClient {
	return nil
}

// ==================== WithContext 方法 - goCache 不支持追踪，忽略 ctx ====================

func (c *goCache) SetWithContext(_ context.Context, key string, value interface{}, expiration time.Duration) error {
	c.cache.Set(key, value, expiration)
	return nil
}

func (c *goCache) GetWithContext(_ context.Context, key string) (interface{}, bool) {
	return c.cache.Get(key)
}

func (c *goCache) GetBytesWithContext(_ context.Context, key string) ([]byte, bool) {
	return nil, false
}

func (c *goCache) GetBatchBytesWithContext(_ context.Context, keys []string) (map[string][]byte, []string) {
	return c.GetBatchBytes(keys)
}

func (c *goCache) DelWithContext(_ context.Context, keys ...string) {
	c.Del(keys...)
}

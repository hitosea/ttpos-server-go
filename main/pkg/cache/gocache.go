package cache

import (
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

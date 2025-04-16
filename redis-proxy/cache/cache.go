package cache

import (
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache 缓存接口
type Cache interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (interface{}, bool)
	Del(keys ...string)
	GetClient() *redis.Client
	GetClusterClient() *redis.ClusterClient
}

var Global Cache
var once sync.Once

// CacheType 缓存类型
type CacheType string

const (
	Redis CacheType = "redis"
)

func Init(cacheType CacheType, config Config) {
	once.Do(func() {
		Global = NewCache(cacheType, config)
	})
}

type Config struct {
	Host     string // 单节点时，正常填。集群时用,号分隔多个地址
	Port     string // 单节点时，正常填。集群时端口要求一样
	Password string
	DB       int
}

// NewCache 创建新的缓存实例
func NewCache(cacheType CacheType, config Config) Cache {
	return newRedisCache(config) // 默认使用redis-cache
}

func (c *redisCache) GetClient() *redis.Client {
	return c.client
}

func (c *redisCache) GetClusterClient() *redis.ClusterClient {
	return c.clusterClient
}

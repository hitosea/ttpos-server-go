package cache

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache 缓存接口
type Cache interface {
	Set(key string, value any, expiration time.Duration) error
	Get(key string) (any, bool)
	GetBytes(key string) ([]byte, bool)
	GetBatchBytes(keys []string) (map[string][]byte, []string)
	Del(keys ...string)
	GetClient() *redis.Client
	GetClusterClient() *redis.ClusterClient

	// WithContext 方法 - 支持链路追踪
	SetWithContext(ctx context.Context, key string, value any, expiration time.Duration) error
	GetWithContext(ctx context.Context, key string) (any, bool)
	GetBytesWithContext(ctx context.Context, key string) ([]byte, bool)
	GetBatchBytesWithContext(ctx context.Context, keys []string) (map[string][]byte, []string)
	DelWithContext(ctx context.Context, keys ...string)
}

// CacheType 缓存类型
type CacheType string

const (
	Redis   CacheType = "redis"
	GoCache CacheType = "go-cache"
)

type Config struct {
	Host     string // 单节点时，正常填。集群时用,号分隔多个地址
	Port     string // 单节点时，正常填。集群时端口要求一样
	Password string
	DB       int
}

// NewCache 创建新的缓存实例
func NewCache(cacheType CacheType, config Config) Cache {
	switch cacheType {
	case Redis:
		return newRedisCache(config)
	case GoCache:
		return newGoCache(config) // TODO 可能还是得换 sqlite或其他支持落盘的缓存引擎，否则离线版挂了的话，缓存就没有了
	default:
		return newRedisCache(config) // 默认使用redis-cache
	}
}

func (c *redisCache) GetClient() *redis.Client {
	return c.client
}

func (c *redisCache) GetClusterClient() *redis.ClusterClient {
	return c.clusterClient
}

var Global Cache
var once sync.Once

func Init(cacheType CacheType, config Config) {
	once.Do(func() {
		Global = NewCache(cacheType, config)
	})
}

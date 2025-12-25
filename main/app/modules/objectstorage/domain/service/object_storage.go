package service

import (
	"context"
	"sync"
	"time"

	"ttpos-server-go/app/modules/objectstorage/domain/entity"
	"ttpos-server-go/app/modules/objectstorage/domain/repository"
)

// IObjectStorage 对象存储领域服务接口
type IObjectStorage[T any] interface {
	// Get 获取对象，自动处理缓存查询和回填
	// key 格式：{company_uuid}:{object_type}:{object_uuid}
	Get(ctx context.Context, key string, query func() (T, error)) (T, error)

	// BatchGet 批量获取对象
	BatchGet(ctx context.Context, keys []string, query func([]string) (map[string]T, error)) (map[string]T, error)

	// Invalidate 使缓存失效
	Invalidate(ctx context.Context, key string) error

	// Update 更新缓存
	Update(ctx context.Context, key string, value T) error

	// Warmup 预热缓存
	Warmup(ctx context.Context, keys []string, query func([]string) (map[string]T, error)) error

	// InvalidateByCompany 按 company 粒度批量失效缓存
	InvalidateByCompany(ctx context.Context, companyUuid uint64) error

	// InvalidateByCompanyAndType 按 company + object_type 粒度批量失效缓存
	InvalidateByCompanyAndType(ctx context.Context, companyUuid uint64, objectType string) error

	// UpdateByCompany 按 company 粒度批量更新缓存
	UpdateByCompany(ctx context.Context, companyUuid uint64, objectType string, values map[string]T) error

	// PreloadWithConfig 配置映射自动关联注入（推荐方式）
	PreloadWithConfig(ctx context.Context, obj interface{}, associations []entity.Association) error
}

// Config 配置选项（泛型版本）
type Config[T any] struct {
	// TTL 缓存过期时间
	TTL time.Duration

	// DisableCache 是否禁用缓存（用于调试）
	DisableCache bool

	// KeyPrefix Key 前缀（自动包含 company UUID）
	KeyPrefix string

	// CacheLayer 三级缓存基础包实例
	CacheLayer repository.CacheLayer[T]

	// ttlMap 不同对象类型的 TTL 配置
	ttlMap map[string]time.Duration
	mu     sync.RWMutex
}

// SetTTL 为指定对象类型设置 TTL
func (c *Config[T]) SetTTL(objectType string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ttlMap == nil {
		c.ttlMap = make(map[string]time.Duration)
	}
	c.ttlMap[objectType] = ttl
}

// GetTTL 获取指定对象类型的 TTL，如果未配置则返回默认 TTL
func (c *Config[T]) GetTTL(objectType string) time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.ttlMap != nil {
		if ttl, ok := c.ttlMap[objectType]; ok {
			return ttl
		}
	}

	return c.TTL
}

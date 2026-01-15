package controller

import (
	goCtx "context"
	"fmt"

	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
)

// CacheVersionController 缓存版本控制器（泛型）
// 用于统一管理对象缓存的失效操作
type CacheVersionController struct {
	cache          cache.Cache
	versionManager *persistence.CacheVersionManager
	objectType     string // 对象类型
}

// NewCacheVersionController 创建缓存版本控制器
// 参数：
//   - cacheInstance: 缓存实例
//   - objectType: 对象类型
//   - defaultUuid: 默认 UUID（通常为 0，表示全局版本时间戳）
//
// 返回：
//   - *CacheVersionController: 缓存版本控制器实例
func NewCacheVersionController(cacheInstance cache.Cache, objectType string) *CacheVersionController {
	return &CacheVersionController{
		cache:          cacheInstance,
		versionManager: persistence.NewCacheVersionManager(cacheInstance),
		objectType:     objectType,
	}
}

// Invalidate 使对象缓存失效
// 参数：
//   - ctx: 上下文（用于提取 companyUuid）
//   - uuid: 对象的 UUID（如果为 0，则使用 defaultUuid；如果不为 0，则使用传入的 uuid）
//
// 返回：
//   - error: 错误信息
func (c *CacheVersionController) Invalidate(ctx goCtx.Context, uuid uint64) error {
	cctx := ctx.(context.Context)
	companyUuid := cctx.GetCompanyUuid()

	// 更新缓存版本时间戳（在二级缓存中记录对象的最新版本时间戳）
	// Key 格式：{cache_version_prefix}:{system_prefix}:{company_uuid}:{object_type}:{object_uuid}
	// object_uuid 为 0 表示全局版本时间戳，要更新所有的该类型对象缓存
	// object_uuid 不为 0 表示具体对象的版本时间戳，要更新具体对象的缓存
	// 版本时间戳的过期时间设置为 L2TTL（5分钟），与最长有效的缓存时间一致
	// 当版本时间戳过期时，GetCacheVersionTimestamp 会返回 (0, false)，
	// 表示缓存已过期，需要重新查询并设置新的版本时间戳
	if err := c.versionManager.UpdateCacheVersionTimestamp(companyUuid, c.objectType, uuid); err != nil {
		return fmt.Errorf("更新缓存版本时间戳失败: %w", err)
	}

	return nil
}

// InitCacheVersionController 初始化缓存版本控制器（单例模式）
// 参数：
//   - cacheInstance: 缓存实例
//   - objectType: 对象类型
//   - defaultUuid: 默认 UUID（通常为 0，表示全局版本时间戳）
//   - instance: 单例实例指针（用于存储控制器实例）
//   - once: sync.Once 实例（确保只初始化一次）
func InitCacheVersionController(
	cacheInstance cache.Cache,
	objectType string,
) *CacheVersionController {
	return NewCacheVersionController(cacheInstance, objectType)
}

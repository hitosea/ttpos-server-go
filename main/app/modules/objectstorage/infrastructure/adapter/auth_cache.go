package adapter

import (
	goCtx "context"
	"fmt"
	"time"

	"ttpos-server-go/app/modules/objectstorage/domain/repository"
	objectStoragePersistence "ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
)

// GetAuthStaffCache 获取员工信息缓存层（包含 Company 和 CompanySetting）
// 使用阶梯式 TTL：L1 缓存 1 分钟，L2 缓存 5 分钟
// 返回配置好的缓存层，service 层只需要调用 GET 方法并传入 query 函数
func GetAuthStaffCache[T any](underlyingCache cache.Cache) repository.CacheLayer[T] {
	// 创建缓存组配置（使用阶梯式 TTL）
	groupConfig := cache.GroupConfig{
		Name:             "object-storage-auth",
		EnableLocalCache: true,
		EnableRedisCache: true,
		NegativeTTL:      30 * time.Second,
		L1TTL:            1 * time.Minute, // L1 缓存 1 分钟
		L2TTL:            5 * time.Minute, // L2 缓存 5 分钟
	}

	// 使用缓存查询员工信息
	return GetOrCreateCacheLayer[T](
		groupConfig,
		underlyingCache,
		5*time.Minute, // 默认 TTL 5 分钟
		WithSingletonKey("auth_staff_cache_group"),
		WithL1TTL(1*time.Minute),
		WithL2TTL(5*time.Minute),
	)
}

// GetAuthDeskCache 获取桌台信息缓存层
// 使用阶梯式 TTL：L1 缓存 1 分钟，L2 缓存 5 分钟
// 返回配置好的缓存层，service 层只需要调用 GET 方法并传入 query 函数
func GetAuthDeskCache[T any](underlyingCache cache.Cache) repository.CacheLayer[T] {
	// 创建缓存组配置（使用阶梯式 TTL）
	groupConfig := cache.GroupConfig{
		Name:             "object-storage-auth-desk",
		EnableLocalCache: true,
		EnableRedisCache: true,
		NegativeTTL:      30 * time.Second,
		L1TTL:            1 * time.Minute, // L1 缓存 1 分钟
		L2TTL:            5 * time.Minute, // L2 缓存 5 分钟
	}

	// 使用对象存储缓存查询桌台信息
	return GetOrCreateCacheLayer[T](
		groupConfig,
		underlyingCache,
		5*time.Minute,
		WithSingletonKey("auth_desk_cache_group"),
		WithL1TTL(1*time.Minute),
		WithL2TTL(5*time.Minute),
	)
}

// BuildAuthStaffKey 构建员工信息缓存 key
func BuildAuthStaffKey(ctx goCtx.Context, staffUuid uint64) string {
	return objectStoragePersistence.BuildKey(ctx, "staff", staffUuid)
}

// BuildAuthDeskKey 构建桌台信息缓存 key
func BuildAuthDeskKey(ctx goCtx.Context, deviceUuid uint64) string {
	return objectStoragePersistence.BuildKey(ctx, "desk", deviceUuid)
}

// GetApiPermissionCache 获取 API 权限缓存层
// 使用阶梯式 TTL：L1 缓存 1 分钟，L2 缓存 5 分钟
// 返回配置好的缓存层，service 层只需要调用 GET 方法并传入 query 函数
func GetApiPermissionCache[T any](underlyingCache cache.Cache) repository.CacheLayer[T] {
	// 创建缓存组配置（使用阶梯式 TTL）
	groupConfig := cache.GroupConfig{
		Name:             "object-storage-api-permission",
		EnableLocalCache: true,
		EnableRedisCache: true,
		NegativeTTL:      30 * time.Second,
		L1TTL:            1 * time.Minute, // L1 缓存 1 分钟
		L2TTL:            5 * time.Minute, // L2 缓存 5 分钟
	}

	// 使用缓存查询 API 权限
	return GetOrCreateCacheLayer[T](
		groupConfig,
		underlyingCache,
		5*time.Minute, // 默认 TTL 5 分钟
		WithSingletonKey("api_permission_cache_group"),
		WithL1TTL(1*time.Minute),
		WithL2TTL(5*time.Minute),
	)
}

// BuildApiPermissionKey 构建 API 权限缓存 key
// Key 格式：ttpos4:{company_uuid}:api_permission:{staff_uuid}
func BuildApiPermissionKey(companyUuid, staffUuid uint64) string {
	return fmt.Sprintf("%s:%d:api_permission:%d", objectStoragePersistence.SystemPrefix, companyUuid, staffUuid)
}

// GetProductListCache 获取商品列表缓存层
// 使用阶梯式 TTL：L1 缓存 1 分钟，L2 缓存 5 分钟
// 返回配置好的缓存层，service 层只需要调用 GET 方法并传入 query 函数
func GetProductListCache[T any](underlyingCache cache.Cache) repository.CacheLayer[T] {
	// 创建缓存组配置（使用阶梯式 TTL）
	groupConfig := cache.GroupConfig{
		Name:             "object-storage-product-list",
		EnableLocalCache: true,
		EnableRedisCache: true,
		NegativeTTL:      30 * time.Second,
		L1TTL:            1 * time.Minute, // L1 缓存 1 分钟
		L2TTL:            5 * time.Minute, // L2 缓存 5 分钟
	}

	// 获取缓存层（使用单例模式，确保 L1 缓存可以跨请求共享）
	return GetOrCreateCacheLayer[T](
		groupConfig,
		underlyingCache,
		2*time.Minute, // 默认 TTL（用于负缓存等场景）
		WithSingletonKey("product_list_cache_group"), // 使用自定义 key
		WithL1TTL(1*time.Minute),                     // L1 缓存 1 分钟
		WithL2TTL(5*time.Minute),                     // L2 缓存 5 分钟
	)
}

// GetOrderObjectCache 获取订单相关对象缓存层
// 使用阶梯式 TTL：L1 缓存 1 分钟，L2 缓存 5 分钟
// 参数：
//   - underlyingCache: 底层缓存实例
//   - defaultTTL: 默认 TTL（用于负缓存等场景，不同对象类型可以设置不同的 TTL）
//
// 返回配置好的缓存层，repository 层只需要调用 GET 方法并传入 query 函数
func GetOrderObjectCache[T any](underlyingCache cache.Cache, defaultTTL time.Duration) repository.CacheLayer[T] {
	// 创建缓存组配置（使用阶梯式 TTL）
	groupConfig := cache.GroupConfig{
		Name:             "object-storage",
		EnableLocalCache: true,             // 开启 L1 本地缓存
		EnableRedisCache: true,             // 开启 L2 Redis 缓存
		NegativeTTL:      30 * time.Second, // 负缓存 30 秒
		L1TTL:            1 * time.Minute,  // L1 缓存 1 分钟（减少内存占用）
		L2TTL:            5 * time.Minute,  // L2 缓存 5 分钟（保持缓存命中率）
	}

	// 获取缓存层（使用单例模式，确保 L1 缓存可以跨请求共享）
	// 使用类型名称作为单例 key，确保每种对象类型只有一个 cacheGroup 实例
	return GetOrCreateCacheLayer[T](
		groupConfig,
		underlyingCache,
		defaultTTL,
		// 不指定 WithSingletonKey，使用类型名称作为 key
		WithL1TTL(1*time.Minute), // L1 缓存 1 分钟
		WithL2TTL(5*time.Minute), // L2 缓存 5 分钟
	)
}

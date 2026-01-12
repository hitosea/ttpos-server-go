package controller

import (
	goCtx "context"
	"fmt"
	"sync"
	"time"

	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"

	"github.com/redis/go-redis/v9"
)

// ProductListCacheController 商品列表缓存控制器
// 用于统一管理商品列表缓存的失效操作
type ProductListCacheController struct {
	cache cache.Cache
}

// NewProductListCacheController 创建商品列表缓存控制器
func NewProductListCacheController() *ProductListCacheController {
	return &ProductListCacheController{
		cache: cache.Global,
	}
}

// InvalidateProductListCache 使商品列表缓存失效
// 参数：
//   - ctx: 上下文（用于提取 companyUuid）
//
// 返回：
//   - error: 错误信息
func (c *ProductListCacheController) InvalidateProductListCache(ctx goCtx.Context) error {
	cctx := ctx.(context.Context)
	companyUuid := cctx.GetCompanyUuid()

	// 获取 Redis 客户端
	var client redis.UniversalClient
	if clusterClient := c.cache.GetClusterClient(); clusterClient != nil {
		client = clusterClient
	} else if redisClient := c.cache.GetClient(); redisClient != nil {
		client = redisClient
	} else {
		// 如果不是 Redis 缓存，无法使用模式匹配，直接返回
		return nil
	}

	// 构建缓存键模式
	// Key 格式：{system_prefix}:{company_uuid}:product_list:*
	pattern := fmt.Sprintf("%s:%d:%s:*", persistence.SystemPrefix, companyUuid, persistence.ObjectTypeProductList)

	// 扫描并删除所有匹配的键（L2 Redis 缓存）
	backgroundCtx := goCtx.Background()
	keys, err := cache.ScanRedisKeysDefault(backgroundCtx, client, pattern)
	if err != nil {
		return fmt.Errorf("扫描商品列表缓存键失败: %w", err)
	}

	if len(keys) > 0 {
		// 删除 L2 Redis 缓存
		c.cache.Del(keys...)
	}

	// 更新缓存版本时间戳（在二级缓存中记录对象的最新版本时间戳）
	// Key 格式：{system_prefix}:{company_uuid}:cacheversion_{object_type}
	// 版本时间戳的过期时间设置为 L2TTL（5分钟），与最长有效的缓存时间一致
	// 当版本时间戳过期时，GetCacheVersionTimestamp 会返回 (0, false)，
	// 表示缓存已过期，需要重新查询并设置新的版本时间戳
	if err := persistence.UpdateCacheVersionTimestamp(c.cache, companyUuid, persistence.ObjectTypeProductList, 5*time.Minute); err != nil {
		return fmt.Errorf("更新缓存版本时间戳失败: %w", err)
	}

	return nil
}

// 全局单例实例
var (
	productListCacheControllerInstance *ProductListCacheController
	productListCacheControllerOnce     sync.Once
)

// GetProductListCacheController 获取商品列表缓存控制器单例
func GetProductListCacheController() *ProductListCacheController {
	productListCacheControllerOnce.Do(func() {
		productListCacheControllerInstance = NewProductListCacheController()
	})
	return productListCacheControllerInstance
}

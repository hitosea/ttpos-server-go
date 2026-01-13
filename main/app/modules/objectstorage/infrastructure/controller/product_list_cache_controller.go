package controller

import (
	goCtx "context"
	"fmt"
	"sync"

	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
)

// ProductListCacheController 商品列表缓存控制器
// 用于统一管理商品列表缓存的失效操作
type ProductListCacheController struct {
	cache          cache.Cache
	versionManager *persistence.CacheVersionManager
}

// NewProductListCacheController 创建商品列表缓存控制器
func NewProductListCacheController() *ProductListCacheController {
	cacheInstance := cache.Global
	return &ProductListCacheController{
		cache:          cacheInstance,
		versionManager: persistence.NewCacheVersionManager(cacheInstance),
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

	// 更新缓存版本时间戳（在二级缓存中记录对象的最新版本时间戳）
	// Key 格式：{cache_version_prefix}:{system_prefix}:{company_uuid}:{object_type}:{object_uuid} ,object_uuid 为 0 表示全局版本时间戳,要更新所有的product_list缓存. 其他对象的objectUuid不为0,表示具体对象的版本时间戳,要更新具体对象的缓存.
	// 版本时间戳的过期时间设置为 L2TTL（5分钟），与最长有效的缓存时间一致
	// 当版本时间戳过期时，GetCacheVersionTimestamp 会返回 (0, false)，
	// 表示缓存已过期，需要重新查询并设置新的版本时间戳
	if err := c.versionManager.UpdateCacheVersionTimestamp(companyUuid, persistence.ObjectTypeProductList, 0); err != nil {
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

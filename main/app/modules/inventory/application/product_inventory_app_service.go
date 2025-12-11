package inventory

import (
	"encoding/json"
	"fmt"
	"time"
	domainService "ttpos-server-go/app/modules/inventory/domain/service"
	inventoryPersistence "ttpos-server-go/app/modules/inventory/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

const (
	// ProductInventoryCacheKeyPrefix 商品库存缓存键前缀
	ProductInventoryCacheKeyPrefix = "product_inventory:%d:%d" // company_uuid:product_bom_uuid
	// ProductInventoryCacheTTL 商品库存缓存过期时间（5分钟）
	// ProductInventoryCacheTTL = 5 * time.Minute
	ProductInventoryCacheTTL = 30 * time.Second
	// ProductPackageInventoryCacheKeyPrefix 商品包库存缓存键前缀
	ProductPackageInventoryCacheKeyPrefix = "product_package_inventory:%d:%d" // company_uuid:product_package_uuid
	// ProductPackageInventoryCacheTTL 商品包库存缓存过期时间（5分钟）
	// ProductPackageInventoryCacheTTL = 5 * time.Minute
	ProductPackageInventoryCacheTTL = 30 * time.Second
)

// ProductInventoryAppService 商品库存应用服务（带缓存）
type ProductInventoryAppService struct {
	domainService domainService.IProductInventoryDomainService
	cache         cache.Cache
	dbm           *database.DBManager
	enableCache   bool // 是否启用缓存，默认为 true
}

// NewProductInventoryAppService 创建商品库存应用服务
func NewProductInventoryAppService(
	domainService domainService.IProductInventoryDomainService,
	cache cache.Cache,
	dbm *database.DBManager,
) *ProductInventoryAppService {
	return NewProductInventoryAppServiceWithCacheOption(domainService, cache, dbm, true)
}

// NewProductInventoryAppServiceWithCacheOption 创建商品库存应用服务（可配置缓存开关）
func NewProductInventoryAppServiceWithCacheOption(
	domainService domainService.IProductInventoryDomainService,
	cache cache.Cache,
	dbm *database.DBManager,
	enableCache bool,
) *ProductInventoryAppService {
	return &ProductInventoryAppService{
		domainService: domainService,
		cache:         cache,
		dbm:           dbm,
		enableCache:   enableCache,
	}
}

// ProductInventoryAppServiceOption 商品库存应用服务选项
type ProductInventoryAppServiceOption struct {
	EnableCache bool // 是否启用缓存，默认为 true
}

// WithEnableCache 设置是否启用缓存
func WithEnableCache(enableCache bool) func(*ProductInventoryAppServiceOption) {
	return func(option *ProductInventoryAppServiceOption) {
		option.EnableCache = enableCache
	}
}

// NewProductInventoryAppServiceWithDependencies 创建商品库存应用服务（工厂方法）
// 接收 dbm 和 cache 作为参数，自动创建所需的依赖并返回 ProductInventoryAppService 实例
// 用于简化其他模块的使用，避免重复创建依赖
// opts: 可选参数，使用 WithEnableCache 设置缓存开关，默认为 true
func NewProductInventoryAppServiceWithDependencies(
	dbm *database.DBManager,
	cache cache.Cache,
	opts ...func(*ProductInventoryAppServiceOption),
) *ProductInventoryAppService {
	// 创建持久化层
	productBomRepo := inventoryPersistence.NewProductBomRepository(dbm)
	productPackageRepo := inventoryPersistence.NewProductPackageRepository(dbm)

	// 创建领域服务
	domainSvc := domainService.NewProductInventoryDomainService(productBomRepo, productPackageRepo)

	// 解析选项，设置默认值
	option := &ProductInventoryAppServiceOption{
		EnableCache: false, // 默认启用缓存
	}
	for _, opt := range opts {
		opt(option)
	}

	// 创建应用服务
	return NewProductInventoryAppServiceWithCacheOption(domainSvc, cache, dbm, option.EnableCache)
}

// GetProductInventory 获取商品库存（带缓存）
func (s *ProductInventoryAppService) GetProductInventory(
	ctx context.Context,
	productBomUuid uint64,
) (float64, error) {
	companyUuid := ctx.GetCompanyUuid()
	cacheKey := fmt.Sprintf(ProductInventoryCacheKeyPrefix, companyUuid, productBomUuid)

	// 1. 如果启用缓存，尝试从缓存获取
	if s.enableCache {
		if cached, exists := s.cache.Get(cacheKey); exists {
			if cachedStr, ok := cached.(string); ok {
				var inventory float64
				if err := json.Unmarshal([]byte(cachedStr), &inventory); err == nil {
					return inventory, nil
				}
			}
		}
	}

	// 2. 从领域服务获取
	inventory, err := s.domainService.GetProductInventory(ctx, productBomUuid)
	if err != nil {
		return 0, err
	}

	// 3. 如果启用缓存，写入缓存
	if s.enableCache {
		inventoryBytes, _ := json.Marshal(inventory)
		s.cache.Set(cacheKey, string(inventoryBytes), ProductInventoryCacheTTL)
	}

	return inventory, nil
}

// InvalidateProductInventoryCache 使商品库存缓存失效
func (s *ProductInventoryAppService) InvalidateProductInventoryCache(
	companyUuid uint64,
	productBomUuid uint64,
) {
	if !s.enableCache {
		return
	}
	cacheKey := fmt.Sprintf(ProductInventoryCacheKeyPrefix, companyUuid, productBomUuid)
	s.cache.Del(cacheKey)
}

// GetProductPackageInventory 获取商品包库存（带缓存）
// opts: 可选参数，使用 domainService.WithStrategy 设置策略
func (s *ProductInventoryAppService) GetProductPackageInventory(
	ctx context.Context,
	productPackageUuid uint64,
	opts ...func(option *domainService.GetProductPackageInventoryOption),
) (float64, error) {
	companyUuid := ctx.GetCompanyUuid()
	cacheKey := fmt.Sprintf(ProductPackageInventoryCacheKeyPrefix, companyUuid, productPackageUuid)

	// 1. 如果启用缓存，尝试从缓存获取
	if s.enableCache {
		if cached, exists := s.cache.Get(cacheKey); exists {
			if cachedStr, ok := cached.(string); ok {
				var inventory float64
				if err := json.Unmarshal([]byte(cachedStr), &inventory); err == nil {
					return inventory, nil
				}
			}
		}
	}

	// 2. 从领域服务获取
	inventory, err := s.domainService.GetProductPackageInventory(ctx, productPackageUuid, opts...)
	if err != nil {
		return 0, err
	}

	// 3. 如果启用缓存，写入缓存
	if s.enableCache {
		inventoryBytes, _ := json.Marshal(inventory)
		s.cache.Set(cacheKey, string(inventoryBytes), ProductPackageInventoryCacheTTL)
	}

	return inventory, nil
}

// InvalidateProductPackageInventoryCache 使商品包库存缓存失效
func (s *ProductInventoryAppService) InvalidateProductPackageInventoryCache(
	ctx context.Context,
	productPackageUuid uint64,
) error {
	if !s.enableCache {
		return nil
	}
	companyUuid := ctx.GetCompanyUuid()
	cacheKey := fmt.Sprintf(ProductPackageInventoryCacheKeyPrefix, companyUuid, productPackageUuid)
	s.cache.Del(cacheKey)
	return nil
}

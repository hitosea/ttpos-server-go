package repository

import (
	"fmt"
	"slices"
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/cache"

	"github.com/vmihailenco/msgpack"
	"gorm.io/gorm"
)

// 缓存键常量，避免重复的字符串格式化
const (
	CACHE_KEY_DESK                        = "{GORM_CACHE_DESK}:%v:%v"
	CACHE_KEY_BUFFET_CUSTOMER_TYPE_PRICES = "{GORM_CACHE_BUFFET_CUSTOMER_TYPE_PRICES}:%v:%v"
	CACHE_KEY_BUFFET_PACKAGE              = "{GORM_CACHE_BUFFET_PACKAGE}:%v:%v"
	CACHE_KEY_BUFFET_PRODUCTS             = "{GORM_CACHE_BUFFET_PRODUCTS}:%v:%v"
	CACHE_KEY_PRODUCT_PACKAGE             = "{GORM_CACHE_PRODUCT_PACKAGE_UUID}:%v:%v"
	CACHE_KEY_PRODUCT_BOM                 = "{GORM_CACHE_PRODUCT_BOM_UUID}:%v:%v"
	CACHE_KEY_PRODUCT_ATTRIBUTE           = "{GORM_CACHE_PRODUCT_ATTRIBUTE_UUID}:%v:%v"
)

// ICacheDataRepo 缓存数据
type ICacheDataRepo interface {
	GetCacheDesk(uuid uint64) (model.Desk, error)                                                         // 获取缓存桌台
	GetCacheBuffetCustomerTypePrices(buffetPackageUuid uint64) ([]model.BuffetCustomerTypePrice, error)   // 获取缓存自助餐套餐
	GetCacheBuffetPackage(uuid uint64) (model.BuffetPackage, error)                                       // 获取缓存自助餐套餐
	GetCacheBuffetProducts(buffetPackageUuid uint64) ([]model.BuffetProduct, error)                       // 获取缓存自助餐商品
	GetCacheAllProductPackageByUuids(productPackageUuids []uint64) ([]*model.ProductPackage, error)       // 获取缓存商品包
	GetCacheAllProductBomByUuids(productBomUuids []uint64) ([]*model.ProductBom, error)                   // 获取缓存商品包
	GetCacheAllProductAttributeByUuids(productAttributeUuids []uint64) ([]*model.ProductAttribute, error) // 获取缓存商品属性
}

func NewCacheDataRepo(db *gorm.DB) ICacheDataRepo {
	return NewCacheDataRepoImpl(db)
}

// NewCacheDataRepoImpl 创建新的缓存数据仓库实现
func NewCacheDataRepoImpl(db *gorm.DB) *CacheDataRepoImpl {
	return &CacheDataRepoImpl{db: db, companyUuid: GetCompanyUuid(db)}
}

type CacheDataRepoImpl struct {
	db          *gorm.DB
	companyUuid uint64
}

// GetCacheDesk 获取缓存桌台
func (r *CacheDataRepoImpl) GetCacheDesk(uuid uint64) (model.Desk, error) {
	cacheKey := fmt.Sprintf(CACHE_KEY_DESK, r.companyUuid, uuid)
	cacheValue, ok := cache.Global.GetBytes(cacheKey)
	if ok {
		desk := model.Desk{}
		err := msgpack.Unmarshal(cacheValue, &desk)
		if err != nil {
			return model.Desk{}, errors.WithMessage(err)
		}
		return desk, nil
	} else {
		desk, err := NewDeskRepo(r.db).GetDesk(
			CommonRepo.WhereByUuid(uuid),
		)
		if err != nil {
			return model.Desk{}, errors.WithMessage(err)
		}
		data, _ := msgpack.Marshal(desk)
		cache.Global.Set(cacheKey, data, 10*time.Hour)
		return desk, nil
	}
}

// GetCacheBuffetCustomerTypePrices 获取缓存的自助餐顾客类型价格
func (r *CacheDataRepoImpl) GetCacheBuffetCustomerTypePrices(buffetPackageUuid uint64) ([]model.BuffetCustomerTypePrice, error) {
	cacheKey := fmt.Sprintf(CACHE_KEY_BUFFET_CUSTOMER_TYPE_PRICES, r.companyUuid, buffetPackageUuid)

	// 尝试从缓存获取
	if cacheData, ok := cache.Global.GetBytes(cacheKey); ok {
		var prices []model.BuffetCustomerTypePrice
		err := msgpack.Unmarshal(cacheData, &prices)
		if err == nil && len(prices) > 0 {
			return prices, nil
		}
	}

	// 缓存未命中，从数据库查询
	prices, err := NewBuffetCustomerTypePricesRepo(r.db).GetBuffetCustomerTypePrices(
		CommonRepo.WhereByBuffetPackageUuid(buffetPackageUuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "MultiLanguageName",
			},
		),
	)

	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 存入缓存
	data, _ := msgpack.Marshal(prices)
	cache.Global.Set(cacheKey, data, 10*time.Hour)

	return prices, nil
}

// GetCacheBuffetPackage 获取自助餐套餐
func (r *CacheDataRepoImpl) GetCacheBuffetPackage(uuid uint64) (model.BuffetPackage, error) {
	cacheKey := fmt.Sprintf(CACHE_KEY_BUFFET_PACKAGE, r.companyUuid, uuid)
	cacheValue, ok := cache.Global.GetBytes(cacheKey)
	if ok {
		buffetPackage := model.BuffetPackage{}
		err := msgpack.Unmarshal(cacheValue, &buffetPackage)
		if err != nil {
			return model.BuffetPackage{}, errors.WithMessage(err)
		}
		return buffetPackage, nil
	} else {
		buffetPackage, err := NewBuffetPackageRepo(r.db).GetBuffetPackage(
			CommonRepo.WhereByUuid(uuid),
		)
		if err != nil {
			return model.BuffetPackage{}, errors.WithMessage(err)
		}
		data, _ := msgpack.Marshal(buffetPackage)
		cache.Global.Set(cacheKey, data, 6*time.Hour)
		return buffetPackage, nil
	}
}

// GetCacheBuffetProducts 获取缓存的自助餐商品
func (r *CacheDataRepoImpl) GetCacheBuffetProducts(buffetPackageUuid uint64) ([]model.BuffetProduct, error) {
	cacheKey := fmt.Sprintf(CACHE_KEY_BUFFET_PRODUCTS, r.companyUuid, buffetPackageUuid)

	// 尝试从缓存获取
	if cacheData, ok := cache.Global.GetBytes(cacheKey); ok {
		var products []model.BuffetProduct
		err := msgpack.Unmarshal(cacheData, &products)
		if err == nil && len(products) > 0 {
			return products, nil
		}
	}

	// 缓存未命中，从数据库查询
	products, err := NewBuffetProductRepo(r.db).GetBuffetProducts(
		CommonRepo.WhereByBuffetPackageUuid(buffetPackageUuid),
		CommonRepo.Preload(
			WithPreload{
				Query: "ProductPackage",
			},
		),
	)

	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 存入缓存
	data, _ := msgpack.Marshal(products)
	cache.Global.Set(cacheKey, data, 10*time.Hour)

	return products, nil
}

// GetCacheAllProductPackage 批量获取缓存的商品包
func (r *CacheDataRepoImpl) GetCacheAllProductPackageByUuids(productPackageUuids []uint64) ([]*model.ProductPackage, error) {
	if len(productPackageUuids) == 0 {
		return []*model.ProductPackage{}, nil
	}

	// 创建缓存键列表
	cacheKeys := make([]string, len(productPackageUuids))
	for i, uuid := range productPackageUuids {
		cacheKeys[i] = fmt.Sprintf(CACHE_KEY_PRODUCT_PACKAGE, r.companyUuid, uuid)
	}

	// 批量获取缓存
	cacheValuesMap, missedKeys := cache.Global.GetBatchBytes(cacheKeys)

	// 如果没有缓存未命中
	if len(missedKeys) == 0 {
		productPackages := make([]*model.ProductPackage, 0, len(productPackageUuids))

		// 遍历每个UUID，按顺序获取对应的商品包
		for _, uuid := range productPackageUuids {
			cacheKey := fmt.Sprintf(CACHE_KEY_PRODUCT_PACKAGE, r.companyUuid, uuid)
			if cacheValue, exists := cacheValuesMap[cacheKey]; exists {
				var productPackage *model.ProductPackage
				err := msgpack.Unmarshal(cacheValue, &productPackage)
				if err == nil && productPackage != nil {
					productPackages = append(productPackages, productPackage)
				}
			}
		}

		// 如果成功解析了所有商品包，则返回结果
		if len(productPackages) == len(productPackageUuids) {
			return productPackages, nil
		}
	}

	// 获取未命中缓存的UUID列表
	missedUuids := make([]uint64, 0)
	for _, key := range missedKeys {
		// 从缓存键中提取UUID
		var uuid uint64
		_, err := fmt.Sscanf(key, CACHE_KEY_PRODUCT_PACKAGE, r.companyUuid, &uuid)
		if err == nil && uuid > 0 {
			missedUuids = append(missedUuids, uuid)
		}
	}

	// 如果所有缓存都未命中，直接查询所有UUID
	if len(missedUuids) == len(productPackageUuids) {
		missedUuids = productPackageUuids
	}

	// 从数据库查询
	var dbOptions []DBOption
	if len(missedUuids) > 0 {
		dbOptions = append(dbOptions, CommonRepo.WhereInUuids(missedUuids))
	} else {
		dbOptions = append(dbOptions, CommonRepo.WhereInUuids(productPackageUuids))
	}

	dbOptions = append(dbOptions, CommonRepo.Preload(
		WithPreload{Query: "MultiLanguageName"},
		WithPreload{Query: "DineTax"},
		WithPreload{Query: "TakeoutTax"},
		WithPreload{Query: "ProductCategory"},
	))

	productPackages, err := NewProductPackageRepo(r.db).GetProductPackageList(dbOptions...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 批量存入缓存并筛选出请求的商品包
	var resultPackages []*model.ProductPackage
	for _, productPack := range productPackages {
		if slices.Contains(productPackageUuids, productPack.Uuid) {
			resultPackages = append(resultPackages, productPack)
		}
		cacheKey := fmt.Sprintf(CACHE_KEY_PRODUCT_PACKAGE, r.companyUuid, productPack.Uuid)
		data, _ := msgpack.Marshal(productPack)
		cache.Global.Set(cacheKey, data, 6*time.Hour)
	}

	return resultPackages, nil
}

// GetCacheProductBomByUuids 获取缓存的商品包
func (r *CacheDataRepoImpl) GetCacheAllProductBomByUuids(productBomUuids []uint64) ([]*model.ProductBom, error) {
	if len(productBomUuids) == 0 {
		return []*model.ProductBom{}, nil
	}

	// 创建缓存键列表
	cacheKeys := make([]string, len(productBomUuids))
	for i, uuid := range productBomUuids {
		cacheKeys[i] = fmt.Sprintf(CACHE_KEY_PRODUCT_BOM, r.companyUuid, uuid)
	}

	// 批量获取缓存
	cacheValuesMap, missedKeys := cache.Global.GetBatchBytes(cacheKeys)

	// 如果没有缓存未命中
	if len(missedKeys) == 0 {
		productBoms := make([]*model.ProductBom, 0, len(productBomUuids))

		// 遍历每个UUID，按顺序获取对应的商品包
		for _, uuid := range productBomUuids {
			cacheKey := fmt.Sprintf(CACHE_KEY_PRODUCT_BOM, r.companyUuid, uuid)
			if cacheValue, exists := cacheValuesMap[cacheKey]; exists {
				var productBom *model.ProductBom
				err := msgpack.Unmarshal(cacheValue, &productBom)
				if err == nil && productBom != nil {
					productBoms = append(productBoms, productBom)
				}
			}
		}

		// 如果成功解析了所有商品包，则返回结果
		if len(productBoms) == len(productBomUuids) {
			return productBoms, nil
		}
	}

	// 获取未命中缓存的UUID列表
	missedUuids := make([]uint64, 0)
	for _, key := range missedKeys {
		// 从缓存键中提取UUID
		var uuid uint64
		_, err := fmt.Sscanf(key, CACHE_KEY_PRODUCT_PACKAGE, r.companyUuid, &uuid)
		if err == nil && uuid > 0 {
			missedUuids = append(missedUuids, uuid)
		}
	}

	// 如果所有缓存都未命中，直接查询所有UUID
	if len(missedUuids) == len(productBomUuids) {
		missedUuids = productBomUuids
	}

	// 从数据库查询
	var dbOptions []DBOption
	if len(missedUuids) > 0 {
		dbOptions = append(dbOptions, CommonRepo.WhereInUuids(missedUuids))
	} else {
		dbOptions = append(dbOptions, CommonRepo.WhereInUuids(productBomUuids))
	}

	dbOptions = append(dbOptions, CommonRepo.Preload(
		WithPreload{Query: "ProductPackage"},
		WithPreload{Query: "ProductFlavor.MultiLanguageName"},
		WithPreload{Query: "ProductSauce.MultiLanguageName"},
		WithPreload{Query: "FlavorMaterials"},
		WithPreload{Query: "ProductSauce.SauceMaterials"},
	))

	productBoms, err := NewProductBomRepo(r.db).GetProductBoms(dbOptions...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 批量存入缓存并筛选出请求的商品包
	var resultBoms []*model.ProductBom
	for _, productBom := range productBoms {
		if slices.Contains(productBomUuids, productBom.Uuid) {
			resultBoms = append(resultBoms, productBom)
		}
		cacheKey := fmt.Sprintf(CACHE_KEY_PRODUCT_BOM, r.companyUuid, productBom.Uuid)
		data, _ := msgpack.Marshal(productBom)
		cache.Global.Set(cacheKey, data, 6*time.Hour)
	}

	return resultBoms, nil
}

// GetCacheAllProductAttributeByUuids 获取缓存的商品属性
func (r *CacheDataRepoImpl) GetCacheAllProductAttributeByUuids(productAttributeUuids []uint64) ([]*model.ProductAttribute, error) {
	if len(productAttributeUuids) == 0 {
		return []*model.ProductAttribute{}, nil
	}

	// 创建缓存键列表
	cacheKeys := make([]string, len(productAttributeUuids))
	for i, uuid := range productAttributeUuids {
		cacheKeys[i] = fmt.Sprintf(CACHE_KEY_PRODUCT_ATTRIBUTE, r.companyUuid, uuid)
	}

	// 批量获取缓存
	cacheValuesMap, missedKeys := cache.Global.GetBatchBytes(cacheKeys)

	// 如果没有缓存未命中
	if len(missedKeys) == 0 {
		productAttributes := make([]*model.ProductAttribute, 0, len(productAttributeUuids))

		// 遍历每个UUID，按顺序获取对应的商品属性
		for _, uuid := range productAttributeUuids {
			cacheKey := fmt.Sprintf(CACHE_KEY_PRODUCT_ATTRIBUTE, r.companyUuid, uuid)
			if cacheValue, exists := cacheValuesMap[cacheKey]; exists {
				var productAttribute *model.ProductAttribute
				err := msgpack.Unmarshal(cacheValue, &productAttribute)
				if err == nil && productAttribute != nil {
					productAttributes = append(productAttributes, productAttribute)
				}
			}
		}

		// 如果成功解析了所有商品属性，则返回结果
		if len(productAttributes) == len(productAttributeUuids) {
			return productAttributes, nil
		}
	}

	// 获取未命中缓存的UUID列表
	missedUuids := make([]uint64, 0)
	for _, key := range missedKeys {
		// 从缓存键中提取UUID
		var uuid uint64
		_, err := fmt.Sscanf(key, CACHE_KEY_PRODUCT_ATTRIBUTE, r.companyUuid, &uuid)
		if err == nil && uuid > 0 {
			missedUuids = append(missedUuids, uuid)
		}
	}

	// 如果所有缓存都未命中，直接查询所有UUID
	if len(missedUuids) == len(productAttributeUuids) {
		missedUuids = productAttributeUuids
	}

	// 从数据库查询
	var dbOptions []DBOption
	if len(missedUuids) > 0 {
		dbOptions = append(dbOptions, CommonRepo.WhereInUuids(missedUuids))
	} else {
		dbOptions = append(dbOptions, CommonRepo.WhereInUuids(productAttributeUuids))
	}

	dbOptions = append(dbOptions, CommonRepo.Preload(
		WithPreload{Query: "MultiLanguageName"},
	))

	productAttributes, err := NewProductAttributeRepo(r.db).GetProductAttributes(dbOptions...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	// 批量存入缓存并筛选出请求的商品属性
	var resultAttributes []*model.ProductAttribute
	for _, productAttribute := range productAttributes {
		if slices.Contains(productAttributeUuids, productAttribute.Uuid) {
			resultAttributes = append(resultAttributes, productAttribute)
		}
		cacheKey := fmt.Sprintf(CACHE_KEY_PRODUCT_ATTRIBUTE, r.companyUuid, productAttribute.Uuid)
		data, _ := msgpack.Marshal(productAttribute)
		cache.Global.Set(cacheKey, data, 6*time.Hour)
	}

	return resultAttributes, nil
}

package repository

import (
	"fmt"
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/cache"

	"github.com/vmihailenco/msgpack"
	"gorm.io/gorm"
)

// 缓存键常量，避免重复的字符串格式化
const (
	CACHE_KEY_DESK                        = "GORM_CACHE_DESK:%v"
	CACHE_KEY_BUFFET_CUSTOMER_TYPE_PRICES = "GORM_CACHE_BUFFET_CUSTOMER_TYPE_PRICES:%v"
	CACHE_KEY_BUFFET_PACKAGE              = "GORM_CACHE_BUFFET_PACKAGE:%v"
	CACHE_KEY_BUFFET_PRODUCTS             = "GORM_CACHE_BUFFET_PRODUCTS:%v"
	CACHE_KEY_PRODUCT_PACKAGE             = "GORM_CACHE_PRODUCT_PACKAGE_UUID:%v"
)

// ICacheDataRepo 缓存数据
type ICacheDataRepo interface {
	GetCacheDesk(uuid uint64) (model.Desk, error)                                                       // 获取缓存桌台
	GetCacheBuffetCustomerTypePrices(buffetPackageUuid uint64) ([]model.BuffetCustomerTypePrice, error) // 获取缓存自助餐套餐
	GetCacheBuffetPackage(uuid uint64) (model.BuffetPackage, error)                                     // 获取缓存自助餐套餐
	GetCacheBuffetProducts(buffetPackageUuid uint64) ([]model.BuffetProduct, error)                     // 获取缓存自助餐商品
	GetCacheAllProductPackage(productPackageUuid uint64) (*model.ProductPackage, error)                 // 获取缓存商品包
}

func NewCacheDataRepo(db *gorm.DB) ICacheDataRepo {
	return NewCacheDataRepoImpl(db)
}

// NewCacheDataRepoImpl 创建新的缓存数据仓库实现
func NewCacheDataRepoImpl(db *gorm.DB) *CacheDataRepoImpl {
	return &CacheDataRepoImpl{db: db}
}

type CacheDataRepoImpl struct {
	db *gorm.DB
}

// GetCacheDesk 获取缓存桌台
func (r *CacheDataRepoImpl) GetCacheDesk(uuid uint64) (model.Desk, error) {
	cacheKey := fmt.Sprintf("GORM_CACHE_DESK:%v", uuid)
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
	cacheKey := fmt.Sprintf("GORM_CACHE_BUFFET_CUSTOMER_TYPE_PRICES:%v", buffetPackageUuid)

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
	cacheKey := fmt.Sprintf(CACHE_KEY_BUFFET_PACKAGE, uuid)
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
	cacheKey := fmt.Sprintf("GORM_CACHE_BUFFET_PRODUCTS:%v", buffetPackageUuid)

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

// GetCacheAllProductPackage 获取缓存的商品包
func (r *CacheDataRepoImpl) GetCacheAllProductPackage(productPackageUuid uint64) (*model.ProductPackage, error) {
	cacheKey := fmt.Sprintf(CACHE_KEY_PRODUCT_PACKAGE, productPackageUuid)
	cacheValue, ok := cache.Global.GetBytes(cacheKey)
	if ok {
		var productPackage *model.ProductPackage
		err := msgpack.Unmarshal(cacheValue, &productPackage)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		return productPackage, nil
	} else {
		productPackages, err := NewProductPackageRepo(r.db).GetProductPackageList(
			CommonRepo.Preload(
				WithPreload{
					Query: "MultiLanguageName",
				},
				WithPreload{
					Query: "DineTax",
				},
				WithPreload{
					Query: "DineTax",
				},
				WithPreload{
					Query: "TakeoutTax",
				},
			),
		)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		// 批量存入缓存
		var productPackage *model.ProductPackage
		for _, productPack := range productPackages {
			if productPack.Uuid == productPackageUuid {
				productPackage = productPack
			}
			cacheKey := fmt.Sprintf(CACHE_KEY_PRODUCT_PACKAGE, productPack.Uuid)
			data, _ := msgpack.Marshal(productPack)
			cache.Global.Set(cacheKey, data, 6*time.Hour)
		}
		return productPackage, nil
	}
}

package repository

import (
	"fmt"
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/utils"

	"github.com/vmihailenco/msgpack"
	"gorm.io/gorm"
)

// ICacheDataRepo 缓存数据
type ICacheDataRepo interface {
	GetCacheDesk(uuid uint64) (model.Desk, error)                                                       // 获取缓存桌台
	GetCacheBuffetCustomerTypePrices(buffetPackageUuid uint64) ([]model.BuffetCustomerTypePrice, error) // 获取缓存自助餐套餐
	GetCacheBuffetPackage(uuid uint64) (model.BuffetPackage, error)                                     // 获取缓存自助餐套餐
	GetCacheBuffetProducts(buffetPackageUuid uint64) ([]model.BuffetProduct, error)                     // 获取缓存自助餐商品
	GetCacheAllProductPackage(productPackageUuid uint64) ([]*model.ProductPackage, error)               // 获取缓存商品包
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
	cacheValue, ok := cache.Global.Get(cacheKey)
	if ok {
		desk := model.Desk{}
		utils.JsonToStruct(cacheValue.(string), &desk)
		return desk, nil
	} else {
		desk, err := NewDeskRepo(r.db).GetDesk(
			CommonRepo.WhereByUuid(uuid),
		)
		if err != nil {
			return model.Desk{}, errors.WithMessage(err)
		}
		cache.Global.Set(cacheKey, utils.ToJson(desk), 10*time.Hour)
		return desk, nil
	}
}

// GetCacheBuffetCustomerTypePrices 获取缓存的自助餐顾客类型价格
func (r *CacheDataRepoImpl) GetCacheBuffetCustomerTypePrices(buffetPackageUuid uint64) ([]model.BuffetCustomerTypePrice, error) {
	cacheKey := fmt.Sprintf("GORM_CACHE_BUFFET_CUSTOMER_TYPE_PRICES:%v", buffetPackageUuid)

	// 尝试从缓存获取
	if cacheData, ok := cache.Global.Get(cacheKey); ok {
		var prices []model.BuffetCustomerTypePrice
		err := utils.JsonToStruct(cacheData.(string), &prices)
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
	cache.Global.Set(cacheKey, utils.ToJson(prices), 10*time.Hour)

	return prices, nil
}

// GetCacheBuffetPackage 获取自助餐套餐
func (r *CacheDataRepoImpl) GetCacheBuffetPackage(uuid uint64) (model.BuffetPackage, error) {
	cacheKey := fmt.Sprintf("GORM_CACHE_BUFFET_PACKAGE:%v", uuid)
	cacheValue, ok := cache.Global.Get(cacheKey)
	if ok {
		buffetPackage := model.BuffetPackage{}
		utils.JsonToStruct(cacheValue.(string), &buffetPackage)
		return buffetPackage, nil
	} else {
		buffetPackage, err := NewBuffetPackageRepo(r.db).GetBuffetPackage(
			CommonRepo.WhereByUuid(uuid),
		)
		if err != nil {
			return model.BuffetPackage{}, errors.WithMessage(err)
		}
		cache.Global.Set(cacheKey, utils.ToJson(buffetPackage), 6*time.Hour)
		return buffetPackage, nil
	}
}

// GetCacheBuffetProducts 获取缓存的自助餐商品
func (r *CacheDataRepoImpl) GetCacheBuffetProducts(buffetPackageUuid uint64) ([]model.BuffetProduct, error) {
	cacheKey := fmt.Sprintf("GORM_CACHE_BUFFET_PRODUCTS:%v", buffetPackageUuid)

	// 尝试从缓存获取
	if cacheData, ok := cache.Global.Get(cacheKey); ok {
		var products []model.BuffetProduct
		err := utils.JsonToStruct(cacheData.(string), &products)
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
	cache.Global.Set(cacheKey, utils.ToJson(products), 10*time.Hour)

	return products, nil
}

// GetCacheAllProductPackage 获取缓存的商品包
func (r *CacheDataRepoImpl) GetCacheAllProductPackage(productPackageUuid uint64) ([]*model.ProductPackage, error) {
	cacheKey := fmt.Sprintf("GORM_CACHE_PRODUCT_PACKAGE:%v", productPackageUuid)
	cacheValue, ok := cache.Global.Get(cacheKey)
	if ok {
		startTime := time.Now()
		var productPackages []*model.ProductPackage
		fmt.Println("GetCacheAllProductPackage", time.Since(startTime))
		fmt.Println("GetCacheAllProductPackage", cacheValue)
		// utils.JsonToStruct(cacheValue.(string), &productPackages)

		var productPackagesss []model.ProductPackage
		err := msgpack.Unmarshal(cacheValue.([]byte), &productPackagesss)
		if err != nil {
			fmt.Println("GetCacheAllProductPackage-err", err)
			return nil, errors.WithMessage(err)
		}
		fmt.Println("GetCacheAllProductPackage2", time.Since(startTime))
		return productPackages, nil
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
		data, _ := msgpack.Marshal(productPackages)
		cache.Global.Set(cacheKey, data, 10*time.Hour)
		return productPackages, nil
	}
}

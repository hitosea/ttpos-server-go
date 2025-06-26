package repository

import (
	"fmt"
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

// IBuffetPackageRepo 自助餐套餐
type IBuffetPackageRepo interface {
	GetBuffetPackage(opts ...DBOption) (model.BuffetPackage, error) // 获取自助餐套餐
	GetCacheBuffetPackage(uuid uint64) (model.BuffetPackage, error) // 获取缓存自助餐套餐
	WhereUuid(uuid uint64) DBOption                                 // 通过uuid查询
}

func NewBuffetPackageRepo(db *gorm.DB) IBuffetPackageRepo {
	return NewBuffetPackageRepoImpl(db)
}

// NewBuffetPackageRepoImpl 创建新的自助餐套餐仓库实现
func NewBuffetPackageRepoImpl(db *gorm.DB) *BuffetPackageRepoImpl {
	return &BuffetPackageRepoImpl{db: db}
}

type BuffetPackageRepoImpl struct {
	db *gorm.DB
}

// WhereUuid 通过uuid查询
func (r *BuffetPackageRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// GetBuffetPackage 获取自助餐套餐
func (r *BuffetPackageRepoImpl) GetBuffetPackage(opts ...DBOption) (model.BuffetPackage, error) {
	var buffetPackage model.BuffetPackage
	db := r.db.Model(&model.BuffetPackage{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&buffetPackage).Error
	return buffetPackage, errors.WithMessage(err)
}

// GetCacheBuffetPackage 获取自助餐套餐
func (r *BuffetPackageRepoImpl) GetCacheBuffetPackage(uuid uint64) (model.BuffetPackage, error) {
	cacheKey := fmt.Sprintf("BUFFET_PACKAGE_CACHE:%v", uuid)
	cacheValue, ok := cache.Global.Get(cacheKey)
	if ok {
		buffetPackage := model.BuffetPackage{}
		utils.JsonToStruct(cacheValue.(string), &buffetPackage)
		return buffetPackage, nil
	} else {
		buffetPackage, err := r.GetBuffetPackage(r.WhereUuid(uuid))
		if err != nil {
			return model.BuffetPackage{}, errors.WithMessage(err)
		}
		cache.Global.Set(cacheKey, utils.ToJson(buffetPackage), 6*time.Hour)
		return buffetPackage, nil
	}
}

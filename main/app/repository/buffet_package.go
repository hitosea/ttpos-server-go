package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IBuffetPackageRepo 自助餐套餐
type IBuffetPackageRepo interface {
	GetBuffetPackage(opts ...DBOption) (model.BuffetPackage, error) // 获取自助餐套餐
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

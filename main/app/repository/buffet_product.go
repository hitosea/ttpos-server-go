package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IBuffetProductRepo 自助餐商品
type IBuffetProductRepo interface {
	GetBuffetProducts(opts ...DBOption) ([]model.BuffetProduct, error) // 获取自助餐商品列表
}

func NewBuffetProductRepo(db *gorm.DB) IBuffetProductRepo {
	return NewBuffetProductRepoImpl(db)
}

// NewBuffetProductRepoImpl 创建新的自助餐商品仓库实现
func NewBuffetProductRepoImpl(db *gorm.DB) *BuffetProductRepoImpl {
	return &BuffetProductRepoImpl{db: db}
}

type BuffetProductRepoImpl struct {
	db *gorm.DB
}

// GetBuffetProducts 获取自助餐商品列表
func (r *BuffetProductRepoImpl) GetBuffetProducts(opts ...DBOption) ([]model.BuffetProduct, error) {
	var buffetProducts []model.BuffetProduct
	db := r.db.Model(&model.BuffetProduct{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&buffetProducts).Error
	return buffetProducts, errors.WithMessage(err)
}

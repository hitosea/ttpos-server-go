package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IBuffetCustomerTypePricesRepo 自助餐套餐
type IBuffetCustomerTypePricesRepo interface {
	GetBuffetCustomerTypePrices(opts ...DBOption) ([]model.BuffetCustomerTypePrice, error) // 获取自助餐套餐
}

func NewBuffetCustomerTypePricesRepo(db *gorm.DB) IBuffetCustomerTypePricesRepo {
	return NewBuffetCustomerTypePricesRepoImpl(db)
}

// NewBuffetCustomerTypePricesRepoImpl 创建新的自助餐套餐仓库实现
func NewBuffetCustomerTypePricesRepoImpl(db *gorm.DB) *BuffetCustomerTypePricesRepoImpl {
	return &BuffetCustomerTypePricesRepoImpl{db: db}
}

type BuffetCustomerTypePricesRepoImpl struct {
	db *gorm.DB
}

// GetBuffetCustomer
func (r *BuffetCustomerTypePricesRepoImpl) GetBuffetCustomerTypePrices(opts ...DBOption) ([]model.BuffetCustomerTypePrice, error) {
	var buffetCustomerTypePrices []model.BuffetCustomerTypePrice
	db := r.db.Model(&model.BuffetCustomerTypePrice{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&buffetCustomerTypePrices).Error
	return buffetCustomerTypePrices, errors.WithMessage(err)
}

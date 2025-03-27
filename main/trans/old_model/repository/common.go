package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/trans/old_model"

	"gorm.io/gorm"
)

// ICommonRepo 通用
type ICommonRepo interface {
	GetProductTax(productID uint64, taxType uint) (*old_model.ProductTax, error)
	GetProductSKUs(productID uint64) ([]*old_model.ProductSKU, error)
}

func NewCommonRepo(db *gorm.DB) ICommonRepo {
	return NewCommonRepoImpl(db)
}

// NewCommonRepoImpl 创建新的角色仓库实现
func NewCommonRepoImpl(db *gorm.DB) *CommonRepoImpl {
	return &CommonRepoImpl{db: db}
}

type CommonRepoImpl struct {
	db *gorm.DB
}

// GetProductTax 获取商品税
func (s *CommonRepoImpl) GetProductTax(productID uint64, taxType uint) (*old_model.ProductTax, error) {
	var productTax old_model.ProductTax
	if err := s.db.Where("product_id = ? AND product_tax_type = ?", productID, taxType).First(&productTax).Error; err != nil {
		return nil, err
	}
	return &productTax, nil
}

// GetProductSKU 获取商品规格
func (s *CommonRepoImpl) GetProductSKUs(productID uint64) ([]*old_model.ProductSKU, error) {
	var productSKUs []*old_model.ProductSKU
	if err := s.db.Where("product_id = ?", productID).Find(&productSKUs).Error; err != nil {
		return nil, errors.WithMessage(err)
	}
	return productSKUs, nil
}

package repository

import (
	"ttpos-server-go/app/errors"
	v1 "ttpos-server-go/trans/v1"

	"gorm.io/gorm"
)

// ICommonRepo 通用
type ICommonRepo interface {
	GetProductTax(productID uint64, taxType uint) (*v1.ProductTax, error)
	GetProductSKUs(productID uint64) ([]*v1.ProductSKU, error)
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
func (s *CommonRepoImpl) GetProductTax(productID uint64, taxType uint) (*v1.ProductTax, error) {
	var productTax v1.ProductTax
	if err := s.db.Where("product_id = ? AND product_tax_type = ?", productID, taxType).First(&productTax).Error; err != nil {
		return nil, err
	}
	return &productTax, nil
}

// GetProductSKU 获取商品规格
func (s *CommonRepoImpl) GetProductSKUs(productID uint64) ([]*v1.ProductSKU, error) {
	var productSKUs []*v1.ProductSKU
	if err := s.db.Preload("ProductSoldOut").Where("product_id = ?", productID).Find(&productSKUs).Error; err != nil {
		return nil, errors.WithMessage(err)
	}
	return productSKUs, nil
}

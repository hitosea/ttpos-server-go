package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductPackageAttributeRepo 产品包属性仓库接口
type IProductPackageAttributeRepo interface {
	GetProductPackageAttributeList(productPackageAttributeGroupId uint) ([]model.ProductPackageAttribute, error)                    // 获取产品包属性列表
	UpdateProductPackageAttribute(productPackageAttributeGroupId uint, productPackageAttribute model.ProductPackageAttribute) error // 更新产品包属性
	CreateProductPackageAttribute(productPackageAttribute model.ProductPackageAttribute) (uint64, error)                            // 创建产品包属性
	DeleteProductPackageAttribute(productPackageAttributeGroupId uint) error                                                        // 删除产品包属性
}

// NewProductPackageAttributeRepo 创建新的产品包属性仓库
func NewProductPackageAttributeRepo(db *gorm.DB) IProductPackageAttributeRepo {
	return NewProductPackageAttributeRepoImpl(db)
}

// NewProductPackageAttributeRepoImpl 创建新的赠品或免费订单原因仓库实现
func NewProductPackageAttributeRepoImpl(db *gorm.DB) *ProductPackageAttributeRepoImpl {
	return &ProductPackageAttributeRepoImpl{db: db}
}

type ProductPackageAttributeRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// GetProductPackageAttributeList 获取赠品或免费订单原因列表，排除逻辑删除的原因
func (r *ProductPackageAttributeRepoImpl) GetProductPackageAttributeList(productPackageAttributeGroupId uint) ([]model.ProductPackageAttribute, error) {
	var productPackageAttributes []model.ProductPackageAttribute
	err := r.db.Model(&model.ProductPackageAttribute{}).Where("product_package_attribute_group_id = ?", productPackageAttributeGroupId).Find(&productPackageAttributes).Error
	return productPackageAttributes, err
}

// UpdateProductPackageAttribute 更新产品包属性
func (r *ProductPackageAttributeRepoImpl) UpdateProductPackageAttribute(productPackageAttributeGroupId uint, productPackageAttribute model.ProductPackageAttribute) error {

	if err := r.db.Model(&model.ProductPackageAttribute{}).Where("product_package_attribute_group_id = ?", productPackageAttributeGroupId).Updates(productPackageAttribute).Error; err != nil {
		return err
	}

	return nil // 返回更新结果
}

// CreateProductPackageAttribute 创建产品包属性
func (r *ProductPackageAttributeRepoImpl) CreateProductPackageAttribute(productPackageAttribute model.ProductPackageAttribute) (uint64, error) {
	// 创建产品包属性
	if err := r.db.Model(&model.ProductPackageAttribute{}).Create(&productPackageAttribute).Error; err != nil {
		return 0, err
	}

	return productPackageAttribute.Uuid, nil // 返回产品包属性ID
}

// DeleteProductPackageAttribute 软删除产品包属性
func (r *ProductPackageAttributeRepoImpl) DeleteProductPackageAttribute(productPackageAttributeGroupId uint) error {
	return r.db.Model(&model.ProductPackageAttribute{}).Where("product_package_attribute_group_id = ?", productPackageAttributeGroupId).Update("delete_time", uint(time.Now().Unix())).Error
}

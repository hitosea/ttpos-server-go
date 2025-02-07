package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductPackageAttributeGroupRepo 产品包属性组仓库接口
type IProductPackageAttributeGroupRepo interface {
	GetProductPackageAttributeGroupList(productPackageId uint) ([]model.ProductPackageAttributeGroup, error)                         // 获取产品包属性组列表
	UpdateProductPackageAttributeGroup(productPackageId uint, productPackageAttributeGroup model.ProductPackageAttributeGroup) error // 更新产品包属性组
	CreateProductPackageAttributeGroup(productPackageAttributeGroup model.ProductPackageAttributeGroup) (uint, error)                // 创建产品包属性组
	DeleteProductPackageAttributeGroup(productPackageId uint) error                                                                  // 删除产品包属性组
}

// NewProductPackageAttributeGroupRepo 创建新的产品包属性组仓库
func NewProductPackageAttributeGroupRepo(db *gorm.DB) IProductPackageAttributeGroupRepo {
	return NewProductPackageAttributeGroupRepoImpl(db)
}

// NewProductPackageAttributeGroupRepoImpl 创建新的赠品或免费订单原因仓库实现
func NewProductPackageAttributeGroupRepoImpl(db *gorm.DB) *ProductPackageAttributeGroupRepoImpl {
	return &ProductPackageAttributeGroupRepoImpl{db: db}
}

type ProductPackageAttributeGroupRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// GetProductPackageAttributeGroupList 获取赠品或免费订单原因列表，排除逻辑删除的原因
func (r *ProductPackageAttributeGroupRepoImpl) GetProductPackageAttributeGroupList(productPackageId uint) ([]model.ProductPackageAttributeGroup, error) {
	var productPackageAttributeGroups []model.ProductPackageAttributeGroup
	err := r.db.Model(&model.ProductPackageAttributeGroup{}).Where("product_package_id = ?", productPackageId).Find(&productPackageAttributeGroups).Error
	return productPackageAttributeGroups, err
}

// UpdateProductPackageAttributeGroup 更新产品包属性组
func (r *ProductPackageAttributeGroupRepoImpl) UpdateProductPackageAttributeGroup(productPackageId uint, productPackageAttributeGroup model.ProductPackageAttributeGroup) error {

	if err := r.db.Model(&model.ProductPackageAttributeGroup{}).Where("product_package_id = ?", productPackageId).Updates(productPackageAttributeGroup).Error; err != nil {
		return err
	}

	return nil // 返回更新结果
}

// CreateProductPackageAttributeGroup 创建产品包属性组
func (r *ProductPackageAttributeGroupRepoImpl) CreateProductPackageAttributeGroup(productPackageAttributeGroup model.ProductPackageAttributeGroup) (uint, error) {
	// 创建产品包属性组
	if err := r.db.Create(&productPackageAttributeGroup).Error; err != nil {
		return 0, err
	}

	return productPackageAttributeGroup.ID, nil // 返回产品包属性组ID
}

// DeleteProductPackageAttributeGroup 软删除产品包属性组
func (r *ProductPackageAttributeGroupRepoImpl) DeleteProductPackageAttributeGroup(productPackageId uint) error {
	return r.db.Model(&model.ProductPackageAttributeGroup{}).Where("product_package_id = ?", productPackageId).Update("delete_time", uint(time.Now().Unix())).Error
}

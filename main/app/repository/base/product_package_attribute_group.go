package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// 产品包属性组仓库接口
type ProductPackageAttributeGroupRepoInterface interface {
	GetProductPackageAttributeGroupList(productPackageId uint) ([]model.ProductPackageAttributeGroup, error)                         // 获取产品包属性组列表
	UpdateProductPackageAttributeGroup(productPackageId uint, productPackageAttributeGroup model.ProductPackageAttributeGroup) error // 更新产品包属性组
	CreateProductPackageAttributeGroup(productPackageAttributeGroup model.ProductPackageAttributeGroup) (uint, error)                // 创建产品包属性组
	DeleteProductPackageAttributeGroup(productPackageId uint) error                                                                  // 删除产品包属性组
}

// 创建新的产品包属性组仓库
func NewProductPackageAttributeGroupRepo(db *gorm.DB) ProductPackageAttributeGroupRepoInterface {
	return NewProductPackageAttributeGroupRepoImpl(db)
}

// 创建新的赠品或免费订单原因仓库实现
func NewProductPackageAttributeGroupRepoImpl(db *gorm.DB) *ProductPackageAttributeGroupRepoImpl {
	return &ProductPackageAttributeGroupRepoImpl{db: db}
}

type ProductPackageAttributeGroupRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// 获取赠品或免费订单原因列表，排除逻辑删除的原因
func (r *ProductPackageAttributeGroupRepoImpl) GetProductPackageAttributeGroupList(productPackageId uint) ([]model.ProductPackageAttributeGroup, error) {
	var productPackageAttributeGroups []model.ProductPackageAttributeGroup
	err := r.db.Model(&model.ProductPackageAttributeGroup{}).Where("product_package_id = ?", productPackageId).Find(&productPackageAttributeGroups).Error
	return productPackageAttributeGroups, err
}

// 更新产品包属性组
func (r *ProductPackageAttributeGroupRepoImpl) UpdateProductPackageAttributeGroup(productPackageId uint, productPackageAttributeGroup model.ProductPackageAttributeGroup) error {

	if err := r.db.Model(&model.ProductPackageAttributeGroup{}).Where("product_package_id = ?", productPackageId).Updates(productPackageAttributeGroup).Error; err != nil {
		return err
	}

	return nil // 返回更新结果
}

// 创建产品包属性组
func (r *ProductPackageAttributeGroupRepoImpl) CreateProductPackageAttributeGroup(productPackageAttributeGroup model.ProductPackageAttributeGroup) (uint, error) {
	// 创建产品包属性组
	if err := r.db.Create(&productPackageAttributeGroup).Error; err != nil {
		return 0, err
	}

	return productPackageAttributeGroup.Id, nil // 返回产品包属性组ID
}

// 软删除产品包属性组
func (r *ProductPackageAttributeGroupRepoImpl) DeleteProductPackageAttributeGroup(productPackageId uint) error {
	return r.db.Model(&model.ProductPackageAttributeGroup{}).Where("product_package_id = ?", productPackageId).Update("delete_time", uint(time.Now().Unix())).Error
}

package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductAttributeGroupRepo 产品属性组仓库接口
type IProductAttributeGroupRepo interface {
	GetProductAttributeGroupList() ([]model.ProductAttributeGroup, error)                          // 获取产品属性组列表
	UpdateProductAttributeGroup(id uint, productAttributeGroup model.ProductAttributeGroup) error  // 更新产品属性组
	CreateProductAttributeGroup(productAttributeGroup model.ProductAttributeGroup) (uint64, error) // 创建产品属性组
	DeleteProductAttributeGroup(id uint) error                                                     // 删除产品属性组
}

// NewProductAttributeGroupRepo 创建新的产品属性组仓库
func NewProductAttributeGroupRepo(db *gorm.DB) IProductAttributeGroupRepo {
	return NewProductAttributeGroupRepoImpl(db)
}

// NewProductAttributeGroupRepoImpl 创建新的退菜原因仓库实现
func NewProductAttributeGroupRepoImpl(db *gorm.DB) *ProductAttributeGroupRepoImpl {
	return &ProductAttributeGroupRepoImpl{db: db}
}

type ProductAttributeGroupRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// GetProductAttributeGroupList 获取退菜原因列表，排除逻辑删除的退菜原因
func (r *ProductAttributeGroupRepoImpl) GetProductAttributeGroupList() ([]model.ProductAttributeGroup, error) {
	var productAttributeGroups []model.ProductAttributeGroup
	err := r.db.Model(&model.ProductAttributeGroup{}).Preload("MultiLanguageName").Where("delete_time = ?", 0).Find(&productAttributeGroups).Error
	return productAttributeGroups, err
}

// UpdateProductAttributeGroup 更新退菜原因
func (r *ProductAttributeGroupRepoImpl) UpdateProductAttributeGroup(id uint, productAttributeGroup model.ProductAttributeGroup) error {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	if err := tx.Model(&model.ProductAttributeGroup{}).Where("id = ?", id).Updates(productAttributeGroup).Error; err != nil {
		tx.Rollback() // 更新失败，回滚事务
		return errors.WithMessage(err)
	}

	if err := tx.Model(&productAttributeGroup.MultiLanguageName).Where("id = ?", productAttributeGroup.MultiLanguageNameUuid).Updates(productAttributeGroup.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 更新多语言名称失败，回滚事务
		return errors.WithMessage(err)
	}

	return tx.Commit().Error // 提交事务
}

// CreateProductAttributeGroup 创建产品属性组
func (r *ProductAttributeGroupRepoImpl) CreateProductAttributeGroup(productAttributeGroup model.ProductAttributeGroup) (uint64, error) {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	// 创建多语言名称
	if err := tx.Create(&productAttributeGroup.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 创建多语言名称失败，回滚事务
		return 0, err
	}

	// 创建产品属性组
	if err := tx.Create(&productAttributeGroup).Error; err != nil {
		tx.Rollback() // 创建失败，回滚事务
		return 0, err
	}

	return productAttributeGroup.Uuid, tx.Commit().Error // 提交事务
}

// DeleteProductAttributeGroup 软删除退菜原因
func (r *ProductAttributeGroupRepoImpl) DeleteProductAttributeGroup(id uint) error {
	return r.db.Model(&model.ProductAttributeGroup{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}

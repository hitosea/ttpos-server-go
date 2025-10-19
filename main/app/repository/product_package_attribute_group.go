package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductPackageAttributeGroupRepo interface {
	GetProductPackageAttributeGroup(opts ...DBOption) (*model.ProductPackageAttributeGroup, error)
	GetProductPackageAttributeGroups(opts ...DBOption) ([]model.ProductPackageAttributeGroup, error)
	CreateProductPackageAttributeGroups(productPackageAttributeGroups []model.ProductPackageAttributeGroup) error
	DeleteProductPackageAttributeGroup(opts ...DBOption) error
	UpdateProductPackageAttributeGroup(data map[string]any, opts ...DBOption) error
	DestroyProductPackageAttributeGroup(opts ...DBOption) error
}

type productPackageAttributeGroupRepoImpl struct {
	db *gorm.DB
}

func NewProductPackageAttributeGroupRepo(db *gorm.DB) IProductPackageAttributeGroupRepo {
	return &productPackageAttributeGroupRepoImpl{db: db}
}

func (r *productPackageAttributeGroupRepoImpl) GetProductPackageAttributeGroup(opts ...DBOption) (*model.ProductPackageAttributeGroup, error) {
	var productPackageAttributeGroup model.ProductPackageAttributeGroup
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productPackageAttributeGroup)
	if result.Error != nil {
		return nil, errors.WithMessage(result.Error)
	}

	return &productPackageAttributeGroup, nil
}

func (r *productPackageAttributeGroupRepoImpl) GetProductPackageAttributeGroups(opts ...DBOption) ([]model.ProductPackageAttributeGroup, error) {
	var productPackageAttributeGroups []model.ProductPackageAttributeGroup
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&productPackageAttributeGroups)
	if result.Error != nil {
		return nil, errors.WithMessage(result.Error)
	}

	return productPackageAttributeGroups, nil
}

func (r *productPackageAttributeGroupRepoImpl) CreateProductPackageAttributeGroups(productPackageAttributeGroups []model.ProductPackageAttributeGroup) error {
	// 如果productPackageAttributeGroups为空，则不创建
	if len(productPackageAttributeGroups) == 0 {
		return nil
	}
	// 清空关联对象
	list := make([]model.ProductPackageAttributeGroup, 0)
	for _, attributeGroup := range productPackageAttributeGroups {
		attributeGroup.SetNil()
		list = append(list, attributeGroup)
	}

	// 创建product_package_attribute_group表数据
	if err := r.db.Model(&model.ProductPackageAttributeGroup{}).Create(list).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *productPackageAttributeGroupRepoImpl) DeleteProductPackageAttributeGroup(opts ...DBOption) error {
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Model(&model.ProductPackageAttributeGroup{}).Update("delete_time", time.Now().Unix()).Error
}

func (r *productPackageAttributeGroupRepoImpl) UpdateProductPackageAttributeGroup(data map[string]any, opts ...DBOption) error {
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Model(&model.ProductPackageAttributeGroup{}).Updates(data).Error
}

// DestroyProductPackageAttributeGroup 销毁商品包属性组
func (r *productPackageAttributeGroupRepoImpl) DestroyProductPackageAttributeGroup(opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.ProductPackageAttributeGroup{}).Error
}

package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductPackageAttributeRepo interface {
	GetProductPackageAttribute(opts ...DBOption) (*model.ProductPackageAttribute, error)
	GetProductPackageAttributes(opts ...DBOption) ([]*model.ProductPackageAttribute, error)
	GetProductPackageAttributesByUuids(uuids []uint64) ([]*model.ProductPackageAttribute, error)
	CreateProductPackageAttributes(productPackageAttributes []model.ProductPackageAttribute) error
	DeleteProductPackageAttribute(opts ...DBOption) error
	UpdateProductPackageAttribute(data map[string]any, opts ...DBOption) error
	DestroyProductPackageAttribute(opts ...DBOption) error

	GetProductPackageAttributeGroupCount(attributeUuid uint64) ([]model.ProductPackageAttributeGroupCount, error)
}

type productPackageAttributeRepoImpl struct {
	db *gorm.DB
}

func NewProductPackageAttributeRepo(db *gorm.DB) IProductPackageAttributeRepo {
	return &productPackageAttributeRepoImpl{db: db}
}

func (r *productPackageAttributeRepoImpl) GetProductPackageAttribute(opts ...DBOption) (*model.ProductPackageAttribute, error) {
	var productPackageAttribute model.ProductPackageAttribute
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productPackageAttribute)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productPackageAttribute, nil
}

func (r *productPackageAttributeRepoImpl) GetProductPackageAttributes(opts ...DBOption) ([]*model.ProductPackageAttribute, error) {
	var productPackageAttributes []*model.ProductPackageAttribute
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&productPackageAttributes)
	if result.Error != nil {
		return nil, result.Error
	}

	return productPackageAttributes, nil
}

func (r *productPackageAttributeRepoImpl) GetProductPackageAttributesByUuids(uuids []uint64) ([]*model.ProductPackageAttribute, error) {
	productPackageAttributes, err := r.GetProductPackageAttributes(
		CommonRepo.WhereInUuids(uuids),
		CommonRepo.Preload(WithPreload{
			Query: "Attribute.MultiLanguageName",
		}),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productPackageAttributes, nil
}

func (r *productPackageAttributeRepoImpl) CreateProductPackageAttributes(productPackageAttributes []model.ProductPackageAttribute) error {
	// 如果productPackageAttributes为空，则不创建
	if len(productPackageAttributes) == 0 {
		return nil
	}
	// 清空关联对象
	list := make([]model.ProductPackageAttribute, 0)
	for _, attribute := range productPackageAttributes {
		attribute.SetNil()
		list = append(list, attribute)
	}

	// 创建product_package_attribute表数据
	if err := r.db.Model(&model.ProductPackageAttribute{}).Create(list).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *productPackageAttributeRepoImpl) DeleteProductPackageAttribute(opts ...DBOption) error {
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Model(&model.ProductPackageAttribute{}).Updates(map[string]any{
		"delete_time": time.Now().Unix(),
	}).Error
}

func (r *productPackageAttributeRepoImpl) UpdateProductPackageAttribute(data map[string]any, opts ...DBOption) error {
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	return db.Model(&model.ProductPackageAttribute{}).Updates(data).Error
}

// DestroyProductPackageAttribute 销毁商品包属性
func (r *productPackageAttributeRepoImpl) DestroyProductPackageAttribute(opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Delete(&model.ProductPackageAttribute{}).Error
}

func (r *productPackageAttributeRepoImpl) GetProductPackageAttributeGroupCount(attributeUuid uint64) ([]model.ProductPackageAttributeGroupCount, error) {
	var productPackageAttributeGroupCountList []model.ProductPackageAttributeGroupCount
	err := r.db.Model(&model.ProductPackageAttribute{}).Select("product_package_attribute_group_uuid, count(1) as related_attribute_uuid_count").
		Scopes(NotDeleted).Where("attribute_uuid = ?", attributeUuid).
		Group("product_package_attribute_group_uuid").Scan(&productPackageAttributeGroupCountList).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productPackageAttributeGroupCountList, nil
}

package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductAttributeRepo interface {
	GetProductAttribute(opts ...DBOption) (*model.ProductAttribute, error)
	GetProductAttributes(opts ...DBOption) ([]*model.ProductAttribute, error)
	GetProductAttributesByUuids(uuids []uint64) ([]*model.ProductAttribute, error)
	GetProductAttributeWithMultiLanguageName(uuid uint64) (*model.ProductAttribute, error)       // 根据UUID查询商品属性（包含MultiLanguageName）
	GetProductAttributesWithMultiLanguageName(uuids []uint64) ([]*model.ProductAttribute, error) // 批量根据UUID列表查询商品属性（包含MultiLanguageName）
}

type productAttributeRepoImpl struct {
	db *gorm.DB
}

func NewProductAttributeRepo(db *gorm.DB) IProductAttributeRepo {
	return &productAttributeRepoImpl{db: db}
}

func (r *productAttributeRepoImpl) GetProductAttribute(opts ...DBOption) (*model.ProductAttribute, error) {
	var productAttribute model.ProductAttribute
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productAttribute)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productAttribute, nil
}

func (r *productAttributeRepoImpl) GetProductAttributes(opts ...DBOption) ([]*model.ProductAttribute, error) {
	var productAttributes []*model.ProductAttribute
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&productAttributes)
	if result.Error != nil {
		return nil, result.Error
	}

	return productAttributes, nil
}

func (r *productAttributeRepoImpl) GetProductAttributesByUuids(uuids []uint64) ([]*model.ProductAttribute, error) {
	productAttributes, err := r.GetProductAttributes(
		CommonRepo.WhereInUuids(uuids),
		CommonRepo.Preload(WithPreload{
			Query: "Attribute.MultiLanguageName",
		}),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productAttributes, nil
}

// GetProductAttributeWithMultiLanguageName 根据UUID查询商品属性（包含MultiLanguageName）
func (r *productAttributeRepoImpl) GetProductAttributeWithMultiLanguageName(uuid uint64) (*model.ProductAttribute, error) {
	if uuid == 0 {
		return nil, errors.New("商品属性UUID不能为0")
	}
	attr, err := r.GetProductAttribute(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(WithPreload{
			Query: "MultiLanguageName",
		}),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "查询商品属性失败")
	}
	return attr, nil
}

// GetProductAttributesWithMultiLanguageName 批量根据UUID列表查询商品属性（包含MultiLanguageName）
func (r *productAttributeRepoImpl) GetProductAttributesWithMultiLanguageName(uuids []uint64) ([]*model.ProductAttribute, error) {
	if len(uuids) == 0 {
		return []*model.ProductAttribute{}, nil
	}
	// 过滤掉 0 值
	validUuids := make([]uint64, 0, len(uuids))
	for _, uuid := range uuids {
		if uuid > 0 {
			validUuids = append(validUuids, uuid)
		}
	}
	if len(validUuids) == 0 {
		return []*model.ProductAttribute{}, nil
	}
	productAttributes, err := r.GetProductAttributes(
		CommonRepo.WhereInUuids(validUuids),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(WithPreload{
			Query: "MultiLanguageName",
		}),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "批量查询商品属性失败")
	}
	return productAttributes, nil
}

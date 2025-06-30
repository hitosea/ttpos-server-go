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

package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductPackageAttributeRepo interface {
	GetProductPackageAttribute(opts ...DBOption) (*model.ProductPackageAttribute, error)
	GetProductPackageAttributes(opts ...DBOption) ([]*model.ProductPackageAttribute, error)
	GetProductPackageAttributesByUuids(uuids []uint64) ([]*model.ProductPackageAttribute, error)
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

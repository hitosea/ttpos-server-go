package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductBomRepo interface {
	GetProductBom(opts ...DBOption) (*model.ProductBom, error)
	GetProductBoms(opts ...DBOption) ([]*model.ProductBom, error)
	GetFlavorProductBomByUuid(uuid uint64) (*model.ProductBom, error)
	GetSauceProductBomsByUuids(uuids []uint64) ([]*model.ProductBom, error)
}

type productBomRepoImpl struct {
	db *gorm.DB
}

func NewProductBomRepo(db *gorm.DB) IProductBomRepo {
	return &productBomRepoImpl{db: db}
}

func (r *productBomRepoImpl) GetProductBom(opts ...DBOption) (*model.ProductBom, error) {
	var productBom model.ProductBom
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productBom)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productBom, nil
}

func (r *productBomRepoImpl) GetProductBoms(opts ...DBOption) ([]*model.ProductBom, error) {
	productBoms := make([]*model.ProductBom, 0)
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&productBoms)
	if result.Error != nil {
		return nil, result.Error
	}

	return productBoms, nil
}

func (r *productBomRepoImpl) GetFlavorProductBomByUuid(uuid uint64) (*model.ProductBom, error) {
	productBom, err := r.GetProductBom(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.Preload(WithPreload{
			Query: "ProductFlavor.MultiLanguageName",
		}),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBom, nil
}

func (r *productBomRepoImpl) GetSauceProductBomsByUuids(uuids []uint64) ([]*model.ProductBom, error) {
	productBoms, err := r.GetProductBoms(
		CommonRepo.WhereInUuids(uuids),
		CommonRepo.Preload(WithPreload{
			Query: "ProductSauce.MultiLanguageName",
		}),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBoms, nil
}

package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductBomCardRepo interface {
	IProductBomCardQueryRepo
	CreateProductBomCard(productBomCard model.ProductBomCard) error
	CreateProductBomCardMaterial(productBomCardMaterial model.RelatedMaterial) error
	UpdateProductBomCardMaterial(productBomCardMaterial model.RelatedMaterial) error
	CreateOrUpdateProductBomCardMaterial(productBomCardMaterial model.RelatedMaterial) error
}

type IProductBomCardQueryRepo interface {
	GetProductBomCard(opts ...DBOption) (*model.ProductBomCard, error)
	GetProductBomCardList(opts ...DBOption) ([]*model.ProductBomCard, error)
	GetProductBomCardMaterialByUuid(uuid uint64) (*model.RelatedMaterial, error)
}

type productBomCardRepoImpl struct {
	db *gorm.DB
}

func NewProductBomCardRepo(db *gorm.DB) IProductBomCardRepo {
	return &productBomCardRepoImpl{db: db}
}

func (r *productBomCardRepoImpl) GetProductBomCardList(opts ...DBOption) ([]*model.ProductBomCard, error) {
	var productBomCards []*model.ProductBomCard
	db := r.db

	db = db.Model(&model.ProductBomCard{})

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&productBomCards)
	if result.Error != nil {
		return nil, result.Error
	}

	return productBomCards, nil
}

func (r *productBomCardRepoImpl) GetProductBomCard(opts ...DBOption) (*model.ProductBomCard, error) {
	var productBomCard model.ProductBomCard
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productBomCard)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productBomCard, nil
}

func (r *productBomCardRepoImpl) CreateProductBomCard(productBomCard model.ProductBomCard) error {
	productBomCard.SetNil()
	result := r.db.Create(&productBomCard)
	if result.Error != nil {
		return errors.WithMessage(result.Error)
	}

	return nil
}

func (r *productBomCardRepoImpl) CreateOrUpdateProductBomCardMaterial(productBomCardMaterial model.RelatedMaterial) error {
	if productBomCardMaterial.Uuid == 0 {
		return r.CreateProductBomCardMaterial(productBomCardMaterial)
	} else {
		return r.UpdateProductBomCardMaterial(productBomCardMaterial)
	}
}

func (r *productBomCardRepoImpl) CreateProductBomCardMaterial(productBomCardMaterial model.RelatedMaterial) error {
	productBomCardMaterial.SetNil()
	result := r.db.Create(&productBomCardMaterial)
	if result.Error != nil {
		return errors.WithMessage(result.Error)
	}

	return nil
}

func (r *productBomCardRepoImpl) UpdateProductBomCardMaterial(productBomCardMaterial model.RelatedMaterial) error {
	result := r.db.Model(&model.RelatedMaterial{}).Where("uuid = ?", productBomCardMaterial.Uuid).Updates(productBomCardMaterial)
	if result.Error != nil {
		return errors.WithMessage(result.Error)
	}
	return nil
}

func (r *productBomCardRepoImpl) GetProductBomCardMaterialByUuid(uuid uint64) (*model.RelatedMaterial, error) {
	var productBomCardMaterial model.RelatedMaterial
	result := r.db.Where("uuid = ?", uuid).First(&productBomCardMaterial)
	if result.Error != nil {
		return nil, result.Error
	}
	return &productBomCardMaterial, nil
}

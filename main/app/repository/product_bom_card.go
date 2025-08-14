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
}

type IProductBomCardQueryRepo interface {
	GetProductBomCard(opts ...DBOption) (*model.ProductBomCard, error)
	GetProductBomCardList(opts ...DBOption) ([]*model.ProductBomCard, error)
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

func (r *productBomCardRepoImpl) CreateProductBomCardMaterial(productBomCardMaterial model.RelatedMaterial) error {
	productBomCardMaterial.SetNil()
	result := r.db.Create(&productBomCardMaterial)
	if result.Error != nil {
		return errors.WithMessage(result.Error)
	}

	return nil
}

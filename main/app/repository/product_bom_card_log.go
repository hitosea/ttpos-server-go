package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IProductBomCardLogRepo interface {
	IProductBomCardLogQueryRepo
	CreateProductBomCardLog(productBomCardLog model.ProductBomCardLog) error
}

type IProductBomCardLogQueryRepo interface {
	GetProductBomCardLog(opts ...DBOption) (*model.ProductBomCardLog, error)
	GetProductBomCardLogList(opts ...DBOption) ([]*model.ProductBomCardLog, error)
}

type productBomCardLogRepoImpl struct {
	db *gorm.DB
}

func NewProductBomCardLogRepo(db *gorm.DB) IProductBomCardLogRepo {
	return &productBomCardLogRepoImpl{db: db}
}

func (r *productBomCardLogRepoImpl) GetProductBomCardLogList(opts ...DBOption) ([]*model.ProductBomCardLog, error) {
	var productBomCardLogs []*model.ProductBomCardLog
	db := r.db

	db = db.Model(&model.ProductBomCardLog{})

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&productBomCardLogs)
	if result.Error != nil {
		return nil, result.Error
	}

	return productBomCardLogs, nil
}

func (r *productBomCardLogRepoImpl) GetProductBomCardLog(opts ...DBOption) (*model.ProductBomCardLog, error) {
	var productBomCardLog model.ProductBomCardLog
	db := r.db

	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&productBomCardLog)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productBomCardLog, nil
}

func (r *productBomCardLogRepoImpl) CreateProductBomCardLog(productBomCardLog model.ProductBomCardLog) error {
	result := r.db.Create(&productBomCardLog)
	if result.Error != nil {
		return errors.WithMessage(result.Error)
	}

	return nil
}

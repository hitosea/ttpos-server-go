package repository

import (
	"gorm.io/gorm"

	"ttpos-server-go/app/model"
)

type IProductPrinterRepo interface {
	WhereStatus(status int) DBOption
	GetProductPrinters(opts ...DBOption) ([]model.ProductPrinter, error) //
}

func NewProductPrinterRepo(db *gorm.DB) IProductPrinterRepo {
	return NewProductPrinterRepoImpl(db)
}

type productPrinterRepo struct {
	db *gorm.DB
}

func NewProductPrinterRepoImpl(db *gorm.DB) IProductPrinterRepo {
	return &productPrinterRepo{db: db}
}

func (r *productPrinterRepo) GetProductPrinters(opts ...DBOption) ([]model.ProductPrinter, error) {
	var printers []model.ProductPrinter
	db := r.db.Model(&model.ProductPrinter{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.Find(&printers).Debug().Order("create_time desc").Error
	return printers, err
}

func (r *productPrinterRepo) WhereStatus(status int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

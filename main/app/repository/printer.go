package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/model"
)

type IPrinterRepo interface {
	WhereUuid(uuid uint64) DBOption

	WithPrinterType() DBOption

	GetPrinter(opts ...DBOption) (model.Printer, error)
}

func NewPrinterRepo(db *gorm.DB) IPrinterRepo {
	return NewPrinterRepoImpl(db)
}

type printerRepo struct {
	db *gorm.DB
}

func NewPrinterRepoImpl(db *gorm.DB) IPrinterRepo {
	return &printerRepo{db: db}
}

func (r *printerRepo) GetPrinter(opts ...DBOption) (model.Printer, error) {

	var printer model.Printer
	db := r.db.Model(&model.Printer{}).Scopes(NotDeleted)

	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&printer).Error

	return printer, err
}
func (r *printerRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

func (r *printerRepo) WithPrinterType() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("PrinterType")
	}
}

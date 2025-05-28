package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IPrinterTypeRepo interface {
	GetRecordByKey(companyUuid uint64, key string) model.PrinterType
}

type printerTypeRepository struct {
	db *gorm.DB
}

func NewPrinterTypeRepository(db *gorm.DB) *printerTypeRepository {
	return &printerTypeRepository{db: db}
}

func (r *printerTypeRepository) GetRecordByKey(companyUuid uint64, key string) model.PrinterType {
	var PrinterType model.PrinterType
	r.db.Model(&model.PrinterType{}).Where("`key` = ?", key).First(&PrinterType)
	return PrinterType
}

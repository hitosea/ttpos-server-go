package repository

import (
	"websocket/model"
	"websocket/pkg/database"
)

type printerTypeRepository struct {
	dbm *database.DBManager
}

func NewPrinterTypeRepository(dbm *database.DBManager) *printerTypeRepository {
	return &printerTypeRepository{dbm: dbm}
}

func (r *printerTypeRepository) GetRecordByKey(companyUuid uint64, key string) model.PrinterType {
	var PrinterType model.PrinterType
	r.dbm.GetDB(companyUuid).Model(&model.PrinterType{}).Where("`key` = ?", key).First(&PrinterType)
	return PrinterType
}

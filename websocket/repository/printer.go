package repository

import (
	"websocket/model"
	"websocket/pkg/database"
)

type PrinterRepository struct {
	dbm *database.DBManager
}

func NewPrinterRepository(dbm *database.DBManager) *PrinterRepository {
	return &PrinterRepository{dbm: dbm}
}

func (r *PrinterRepository) GetUsbList(companyUuid uint64) []model.Printer {
	var Printers []model.Printer
	r.dbm.GetDB(companyUuid).Model(&model.Printer{}).Where("is_usb = ?", 1).Find(&Printers)
	return Printers
}

func (r *PrinterRepository) GetUsbListByStatus(companyUuid uint64, status int) []model.Printer {
	var Printers []model.Printer
	r.dbm.GetDB(companyUuid).Model(&model.Printer{}).Where("is_usb = ?", 1).Where("status = ?", status).Find(&Printers)
	return Printers
}

// 更新
func (r *PrinterRepository) Update(companyUuid uint64, id uint, vars map[string]interface{}) error {
	return r.dbm.GetDB(companyUuid).Model(&model.Printer{}).Where("id = ?", id).Updates(vars).Error
}

// 创建
func (r *PrinterRepository) Create(companyUuid uint64, Printer model.Printer) error {
	return r.dbm.GetDB(companyUuid).Create(&Printer).Error
}

// 删除
func (r *PrinterRepository) Delete(companyUuid uint64, id uint) error {
	return r.dbm.GetDB(companyUuid).Model(&model.Printer{}).Where("id = ?", id).Delete(&model.Printer{}).Error
}

package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPrinterCustomizeRepo 打印机定制
type IPrinterCustomizeRepo interface {
	GetPrinterCustomizeList(opts ...DBOption) ([]model.PrinterCustomize, error)
	GetPrinterCustomizeInfo(id uint64) (model.PrinterCustomize, error)
	CreatePrinterCustomize(printerCustomize model.PrinterCustomize) error
	UpdatePrinterCustomize(printerCustomize model.PrinterCustomize) error
	DeletePrinterCustomize(id uint64) error
}

type PrinterCustomizeRepoImpl struct {
	db *gorm.DB
}

func NewPrinterCustomizeRepo(db *gorm.DB) IPrinterCustomizeRepo {
	return &PrinterCustomizeRepoImpl{db: db}
}

// GetPrinterCustomizeList 获取所有打印机定制
func (r *PrinterCustomizeRepoImpl) GetPrinterCustomizeList(opts ...DBOption) ([]model.PrinterCustomize, error) {
	var printerCustomizes []model.PrinterCustomize
	db := r.db.Model(&model.PrinterCustomize{})
	for _, option := range opts {
		option(db)
	}
	err := db.Find(&printerCustomizes).Error
	return printerCustomizes, err
}

// GetPrinterCustomizeInfo 获取打印机定制详情
func (r *PrinterCustomizeRepoImpl) GetPrinterCustomizeInfo(id uint64) (model.PrinterCustomize, error) {
	var printerCustomize model.PrinterCustomize
	db := r.db.Model(&model.PrinterCustomize{}).Where("id = ?", id)
	err := db.First(&printerCustomize).Error
	return printerCustomize, err
}

// CreatePrinterCustomize 创建打印机定制
func (r *PrinterCustomizeRepoImpl) CreatePrinterCustomize(printerCustomize model.PrinterCustomize) error {
	if err := r.db.Create(&printerCustomize).Error; err != nil {
		return err
	}
	return nil
}

// UpdatePrinterCustomize 更新打印机定制
func (r *PrinterCustomizeRepoImpl) UpdatePrinterCustomize(printerCustomize model.PrinterCustomize) error {
	if err := r.db.Save(&printerCustomize).Error; err != nil {
		return err
	}
	return nil
}

// DeletePrinterCustomize 删除打印机定制
func (r *PrinterCustomizeRepoImpl) DeletePrinterCustomize(id uint64) error {
	if err := r.db.Delete(&model.PrinterCustomize{}).Where("id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

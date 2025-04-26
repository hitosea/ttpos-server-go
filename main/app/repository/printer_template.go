package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPrinterTemplateRepo 打印机模板
type IPrinterTemplateRepo interface {
	GetPrinterTemplateInfo(id uint64) (model.PrinterTemplate, error)
	CreatePrinterTemplate(printerTemplate model.PrinterTemplate) error
}

func NewPrinterTemplateRepo(db *gorm.DB) IPrinterTemplateRepo {
	return NewPrinterTemplateRepoImpl(db)
}

// NewPrinterTemplateRepoImpl 创建新的打印机模板仓库实现
func NewPrinterTemplateRepoImpl(db *gorm.DB) *PrinterTemplateRepoImpl {
	return &PrinterTemplateRepoImpl{db: db}
}

type PrinterTemplateRepoImpl struct {
	db *gorm.DB
}

// GetPrinterTemplateInfo 获取打印机模板详情
func (r *PrinterTemplateRepoImpl) GetPrinterTemplateInfo(id uint64) (model.PrinterTemplate, error) {
	var printerTemplate model.PrinterTemplate
	db := r.db.Model(&model.PrinterTemplate{}).Where("id = ?", id)
	err := db.First(&printerTemplate).Error
	return printerTemplate, err
}

// CreatePrinterTemplate 创建打印机模板
func (r *PrinterTemplateRepoImpl) CreatePrinterTemplate(printerTemplate model.PrinterTemplate) error {
	if err := r.db.Create(&printerTemplate).Error; err != nil {
		return err
	}
	return nil
}

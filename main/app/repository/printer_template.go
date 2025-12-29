package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPrinterTemplateRepo 打印机模板
type IPrinterTemplateRepo interface {
	GetPrinterTemplateInfo(id uint64) (model.PrinterTemplate, error)
	GetPrinterTemplateBasicInfo(id uint64) (template int, tmpUuid uint64, err error) // 只获取 Template 和 TmpUuid，不获取 TmpData
	CreatePrinterTemplate(printerTemplate model.PrinterTemplate) error
	// 获取打印菜单列表
	GetPrinterTemplates() ([]model.PrinterTemplate, error)             // 获取所有打印模版列表
	UpdatePrinterTemplate(printerTemplate model.PrinterTemplate) error // 更新打印机模板
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

// GetPrinterTemplateBasicInfo 只获取 Template 和 TmpUuid，不获取大字段 TmpData
func (r *PrinterTemplateRepoImpl) GetPrinterTemplateBasicInfo(id uint64) (template int, tmpUuid uint64, err error) {
	var result struct {
		Template int    `gorm:"column:template"`
		TmpUuid  uint64 `gorm:"column:tmp_uuid"`
	}
	err = r.db.Model(&model.PrinterTemplate{}).
		Select("template, tmp_uuid").
		Where("id = ?", id).
		First(&result).Error
	if err != nil {
		return 0, 0, err
	}
	return result.Template, result.TmpUuid, nil
}

// CreatePrinterTemplate 创建打印机模板
func (r *PrinterTemplateRepoImpl) CreatePrinterTemplate(printerTemplate model.PrinterTemplate) error {
	if err := r.db.Create(&printerTemplate).Error; err != nil {
		return err
	}
	return nil
}

// GetPrinterTemplates 获取所有打印模版列表
func (r *PrinterTemplateRepoImpl) GetPrinterTemplates() ([]model.PrinterTemplate, error) {
	var templates []model.PrinterTemplate
	err := r.db.Model(&model.PrinterTemplate{}).
		Where("delete_time = ?", 0). // 只查询未删除的记录
		Order("create_time DESC").   // 按创建时间倒序排列
		Find(&templates).Error
	return templates, err
}

// UpdatePrinterTemplate 更新打印机模板
func (r *PrinterTemplateRepoImpl) UpdatePrinterTemplate(printerTemplate model.PrinterTemplate) error {
	if err := r.db.Model(&model.PrinterTemplate{}).Where("id = ?", printerTemplate.ID).Updates(printerTemplate).Error; err != nil {
		return err
	}
	return nil
}

package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPrinterCustomizeRepo 打印机定制
type IPrinterCustomizeRepo interface {
	GetPrinterCustomizeList(opts ...DBOption) ([]model.PrinterCustomize, error)
	GetPrinterCustomizeInfo(uuid uint64) (model.PrinterCustomize, error)
	CreatePrinterCustomize(printerCustomize model.PrinterCustomize) error
	UpdatePrinterCustomize(printerCustomize model.PrinterCustomize) error
	DeletePrinterCustomize(uuid uint64) error
	UpdatePrinterCustomizeByTemplateId(templateId uint64) error // 按template_id更新is_use为0, 其他字段不变
	// 检查打印机定制名称是否存在
	CheckPrinterCustomizeNameExists(uuid uint64, name string) (bool, error)

	WhereByTemplateId(templateId uint64) DBOption
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
func (r *PrinterCustomizeRepoImpl) GetPrinterCustomizeInfo(uuid uint64) (model.PrinterCustomize, error) {
	var printerCustomize model.PrinterCustomize
	db := r.db.Model(&model.PrinterCustomize{}).Where("uuid = ?", uuid)
	err := db.First(&printerCustomize).Error
	return printerCustomize, err
}

// CheckPrinterCustomizeNameExists 检查打印机定制名称是否存在
func (r *PrinterCustomizeRepoImpl) CheckPrinterCustomizeNameExists(uuid uint64, name string) (bool, error) {
	var printerCustomize model.PrinterCustomize
	db := r.db.Model(&model.PrinterCustomize{}).Where("name = ? AND uuid != ?", name, uuid)
	err := db.First(&printerCustomize).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return printerCustomize.Uuid > 0, err
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
	if err := r.db.Model(&model.PrinterCustomize{}).Where("uuid = ?", printerCustomize.Uuid).Updates(&printerCustomize).Error; err != nil {
		return err
	}
	return nil
}

// DeletePrinterCustomize 删除打印机定制
func (r *PrinterCustomizeRepoImpl) DeletePrinterCustomize(uuid uint64) error {
	if err := r.db.Where("uuid = ?", uuid).Delete(&model.PrinterCustomize{}).Error; err != nil {
		return err
	}
	return nil
}

// WhereByTemplateId 根据模板ID查询
func (r *PrinterCustomizeRepoImpl) WhereByTemplateId(templateId uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("template_id = ?", templateId)
	}
}

// UpdatePrinterCustomizeByTemplateId 按template_id更新is_use为0, 其他字段不变
func (r *PrinterCustomizeRepoImpl) UpdatePrinterCustomizeByTemplateId(templateId uint64) error {
	if err := r.db.Model(&model.PrinterCustomize{}).Where("template_id = ?", templateId).Update("is_use", 0).Error; err != nil {
		return err
	}
	return nil
}

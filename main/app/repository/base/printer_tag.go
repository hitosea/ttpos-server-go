package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// 打印机标签
type PrinterTagRepoInterface interface {
	GetPrinterTagList() ([]model.PrinterTag, error)
	UpdatePrinterTag(id uint, printerTag model.PrinterTag) error
	CreatePrinterTag(printerTag model.PrinterTag) (uint, error)
	DeletePrinterTag(id uint) error
}

func NewPrinterTagRepo(db *gorm.DB) PrinterTagRepoInterface {
	return NewPrinterTagRepoImpl(db)
}

// 创建新的打印机标签仓库实现
func NewPrinterTagRepoImpl(db *gorm.DB) *PrinterTagRepoImpl {
	return &PrinterTagRepoImpl{db: db}
}

type PrinterTagRepoImpl struct {
	db *gorm.DB
}

// 获取打印机标签列表，排除逻辑删除的标签
func (r *PrinterTagRepoImpl) GetPrinterTagList() ([]model.PrinterTag, error) {
	var printerTags []model.PrinterTag
	err := r.db.Model(&model.PrinterTag{}).Where("delete_time = ?", 0).Find(&printerTags).Error
	return printerTags, err
}

// 更新打印机标签
func (r *PrinterTagRepoImpl) UpdatePrinterTag(id uint, printerTag model.PrinterTag) error {
	return r.db.Model(&model.PrinterTag{}).Where("id = ?", id).Updates(printerTag).Error
}

// 创建打印机标签
func (r *PrinterTagRepoImpl) CreatePrinterTag(printerTag model.PrinterTag) (uint, error) {
	return printerTag.Id, r.db.Create(&printerTag).Error
}

// 软删除打印机标签
func (r *PrinterTagRepoImpl) DeletePrinterTag(id uint) error {
	return r.db.Model(&model.PrinterTag{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}

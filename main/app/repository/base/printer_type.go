package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPrinterTypeRepo 打印机类型
type IPrinterTypeRepo interface {
	GetPrinterTypeList() ([]model.PrinterType, error)
	UpdatePrinterType(id uint, printerType model.PrinterType) error
	CreatePrinterType(printerType model.PrinterType) (uint, error)
	DeletePrinterType(id uint) error
}

func NewPrinterTypeRepo(db *gorm.DB) IPrinterTypeRepo {
	return NewPrinterTypeRepoImpl(db)
}

// NewPrinterTypeRepoImpl 创建新的打印机类型仓库实现
func NewPrinterTypeRepoImpl(db *gorm.DB) *PrinterTypeRepoImpl {
	return &PrinterTypeRepoImpl{db: db}
}

type PrinterTypeRepoImpl struct {
	db *gorm.DB
}

// GetPrinterTypeList 获取打印机类型列表，排除逻辑删除的打印机类型
func (r *PrinterTypeRepoImpl) GetPrinterTypeList() ([]model.PrinterType, error) {
	var printerTypes []model.PrinterType
	err := r.db.Model(&model.PrinterType{}).Where("delete_time = ?", 0).Find(&printerTypes).Error
	return printerTypes, errors.WithMessage(err)
}

// UpdatePrinterType 更新打印机类型
func (r *PrinterTypeRepoImpl) UpdatePrinterType(uuid uint, printerType model.PrinterType) error {
	return r.db.Model(&model.PrinterType{}).Where("uuid = ?", uuid).Updates(printerType).Error
}

// CreatePrinterType 创建打印机类型
func (r *PrinterTypeRepoImpl) CreatePrinterType(printerType model.PrinterType) (uint, error) {
	return printerType.ID, r.db.Create(&printerType).Error
}

// DeletePrinterType 软删除打印机类型
func (r *PrinterTypeRepoImpl) DeletePrinterType(uuid uint) error {
	return r.db.Model(&model.PrinterType{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}

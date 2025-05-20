package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPrinterRepo 打印机
type IPrinterRepo interface {
	GetPrinterList() ([]model.Printer, error)
	GetUsbPrinterList() ([]model.Printer, error)
	GetPrinter(uuid uint) (model.Printer, error)
	UpdatePrinter(id uint, printer model.Printer) error
	CreatePrinter(printer model.Printer) (uint, error)
	DeletePrinter(id uint) error
}

func NewPrinterRepo(db *gorm.DB) IPrinterRepo {
	return NewPrinterRepoImpl(db)
}

// NewPrinterRepoImpl 创建新的打印机仓库实现
func NewPrinterRepoImpl(db *gorm.DB) *PrinterRepoImpl {
	return &PrinterRepoImpl{db: db}
}

type PrinterRepoImpl struct {
	db *gorm.DB
}

// GetPrinterList 获取打印机列表，排除逻辑删除的打印机
func (r *PrinterRepoImpl) GetPrinterList() ([]model.Printer, error) {
	var printers []model.Printer
	err := r.db.Model(&model.Printer{}).Where("delete_time = ?", 0).Find(&printers).Error
	return printers, errors.WithMessage(err)
}

// UpdatePrinter 更新打印机
func (r *PrinterRepoImpl) UpdatePrinter(uuid uint, printer model.Printer) error {
	return r.db.Model(&model.Printer{}).Where("uuid = ?", uuid).Updates(printer).Error
}

// CreatePrinter 创建打印机
func (r *PrinterRepoImpl) CreatePrinter(printer model.Printer) (uint, error) {
	return printer.ID, r.db.Create(&printer).Error
}

// DeletePrinter 软删除打印机
func (r *PrinterRepoImpl) DeletePrinter(uuid uint) error {
	return r.db.Model(&model.Printer{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}

// 获取 usb 打印机，心跳时间超过5分钟的打印机
func (r *PrinterRepoImpl) GetUsbPrinterList() ([]model.Printer, error) {
	var printers []model.Printer
	currentTime := uint(time.Now().Unix())
	heartbeatTimeThreshold := currentTime - 300 // 5分钟 (300秒)
	//
	err := r.db.Model(&model.Printer{}).
		Where("delete_time = ?", 0).
		Where("is_usb = ?", 1).
		Where("status = ?", 0).
		Where("last_heartbeat_time < ?", heartbeatTimeThreshold).
		Find(&printers).Error
	//
	return printers, errors.WithMessage(err)
}

// GetPrinter 获取打印机
func (r *PrinterRepoImpl) GetPrinter(uuid uint) (model.Printer, error) {
	var printer model.Printer
	err := r.db.Model(&model.Printer{}).Where("uuid = ?", uuid).First(&printer).Error
	return printer, errors.WithMessage(err)
}

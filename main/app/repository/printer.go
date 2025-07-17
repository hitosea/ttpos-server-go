package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IPrinterRepo interface {
	WhereUuid(uuid uint64) DBOption
	WhereDeviceSn(deviceSn string) DBOption

	WithPrinterType() DBOption

	GetPrinter(opts ...DBOption) (model.Printer, error)

	GetUsbList() []model.Printer
	Update(id uint, vars map[string]interface{}) error
	UpdateBySourceDeviceSn(id uint, sourceDeviceSn string, vars map[string]interface{}) error
	Create(companyUuid uint64, printer model.Printer) error
}

func NewPrinterRepo(db *gorm.DB) IPrinterRepo {
	return NewPrinterRepoImpl(db)
}

type printerRepo struct {
	db *gorm.DB
}

func NewPrinterRepoImpl(db *gorm.DB) IPrinterRepo {
	return &printerRepo{db: db}
}

func (r *printerRepo) GetPrinter(opts ...DBOption) (model.Printer, error) {

	var printer model.Printer
	db := r.db.Model(&model.Printer{}).Scopes(NotDeleted)

	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&printer).Error

	return printer, errors.WithMessage(err)
}
func (r *printerRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

func (r *printerRepo) WhereDeviceSn(deviceSn string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("device_id = ?", deviceSn)
	}
}

func (r *printerRepo) WithPrinterType() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("PrinterType")
	}
}

func (r *printerRepo) GetUsbList() []model.Printer {
	var Printers []model.Printer
	r.db.Model(&model.Printer{}).Where("is_usb = ?", 1).Where("delete_time = ?", 0).Order("id DESC").Find(&Printers)
	return Printers
}

// Update 更新
func (r *printerRepo) Update(id uint, vars map[string]interface{}) error {
	return r.db.Model(&model.Printer{}).Where("id = ?", id).Updates(vars).Error
}

// UpdateBySourceDeviceSn 更新
func (r *printerRepo) UpdateBySourceDeviceSn(id uint, sourceDeviceSn string, vars map[string]interface{}) error {
	return r.db.Model(&model.Printer{}).Where("id = ?", id).Where("source_device_sn = ?", sourceDeviceSn).Updates(vars).Error
}

// Create 创建
func (r *printerRepo) Create(companyUuid uint64, printer model.Printer) error {
	return r.db.Create(&printer).Error
}

package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type LanPrinterScanRepository struct {
	db *gorm.DB
}

func NewLanPrinterScanRepository(db *gorm.DB) *LanPrinterScanRepository {
	return &LanPrinterScanRepository{db: db}
}

func (r *LanPrinterScanRepository) GetList() []model.LanPrinterScan {
	var LanPrinterScans []model.LanPrinterScan
	r.db.Model(&model.LanPrinterScan{}).Where("delete_time = ?", 0).Find(&LanPrinterScans)
	return LanPrinterScans
}

// GetListByDeviceSn 根据设备SN获取打印机列表
func (r *LanPrinterScanRepository) GetListByDeviceSn(deviceSn string) []model.LanPrinterScan {
	var LanPrinterScans []model.LanPrinterScan
	r.db.Model(&model.LanPrinterScan{}).
		Where("source_device_sn = ? AND delete_time = ?", deviceSn, 0).
		Find(&LanPrinterScans)
	return LanPrinterScans
}

// Update 更新打印机记录
func (r *LanPrinterScanRepository) Update(id uint, vars map[string]interface{}) error {
	return r.db.Model(&model.LanPrinterScan{}).Where("id = ?", id).Updates(vars).Error
}

// Create 创建打印机记录
func (r *LanPrinterScanRepository) Create(printer model.LanPrinterScan) error {
	return r.db.Create(&printer).Error
}

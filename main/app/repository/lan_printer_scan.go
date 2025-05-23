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

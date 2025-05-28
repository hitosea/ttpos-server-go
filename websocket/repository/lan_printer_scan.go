package repository

import (
	"websocket/model"
	"websocket/pkg/database"
)

type LanPrinterScanRepository struct {
	dbm *database.DBManager
}

func NewLanPrinterScanRepository(dbm *database.DBManager) *LanPrinterScanRepository {
	return &LanPrinterScanRepository{dbm: dbm}
}

func (r *LanPrinterScanRepository) GetList(companyUuid uint64) []model.LanPrinterScan {
	var LanPrinterScans []model.LanPrinterScan
	r.dbm.GetDB(companyUuid).Model(&model.LanPrinterScan{}).Where("delete_time = ?", 0).Find(&LanPrinterScans)
	return LanPrinterScans
}

func (r *LanPrinterScanRepository) GetListByStatus(companyUuid uint64, status int) []model.LanPrinterScan {
	var LanPrinterScans []model.LanPrinterScan
	r.dbm.GetDB(companyUuid).Model(&model.LanPrinterScan{}).Where("status = ?", status).Find(&LanPrinterScans)
	return LanPrinterScans
}

// 更新
func (r *LanPrinterScanRepository) Update(companyUuid uint64, id uint, vars map[string]interface{}) error {
	return r.dbm.GetDB(companyUuid).Model(&model.LanPrinterScan{}).Where("id = ?", id).Updates(vars).Error
}

// 更新
func (r *LanPrinterScanRepository) UpdateBySourceDeviceSn(companyUuid uint64, id uint, sourceDeviceSn string, vars map[string]interface{}) error {
	return r.dbm.GetDB(companyUuid).Model(&model.LanPrinterScan{}).Where("id = ?", id).Where("source_device_sn = ?", sourceDeviceSn).Updates(vars).Error
}

// 创建
func (r *LanPrinterScanRepository) Create(companyUuid uint64, LanPrinterScan model.LanPrinterScan) error {
	return r.dbm.GetDB(companyUuid).Create(&LanPrinterScan).Error
}

// 删除
func (r *LanPrinterScanRepository) Delete(companyUuid uint64, id uint) error {
	return r.dbm.GetDB(companyUuid).Model(&model.LanPrinterScan{}).Where("id = ?", id).Delete(&model.LanPrinterScan{}).Error
}

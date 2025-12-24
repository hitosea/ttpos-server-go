package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMaterialRepo 物品仓库接口
type ISaleOrderMaterialRepo interface {
	BatchInsertSaleOrderMaterial(saleOrderMaterials []*model.SaleOrderMaterial) error                        // 批量插入销售订单原料
	DeleteSaleOrderMaterial(saleBillUuid uint64) error                                                       // 删除销售订单原料（按 sale_bill_uuid）
	DeleteSaleOrderMaterialBySaleOrderUuid(saleOrderUuid uint64) error                                       // 删除销售订单原料（按 sale_order_uuid）
	GetSaleOrderMaterialBySaleOrderUuid(saleOrderUuid uint64) ([]*model.SaleOrderMaterial, error)            // 获取指定订单的材料记录
	GetSaleOrderMaterialByCreateTimeBetween(startTime, endTime int64) ([]*model.SaleOrderMaterial, error)    // 获取某时间范围内的销售订单原料（仅未统计的）
	GetSaleOrderMaterialByCreateTimeBetweenAll(startTime, endTime int64) ([]*model.SaleOrderMaterial, error) // 获取某时间范围内的销售订单原料（包含已统计的）
	UpdateSaleOrderMaterialIsSummarized(uuids []uint64) error                                                // 更新销售订单原料的统计状态
}

// NewSaleOrderMaterialRepo 创建新的销售订单原料仓库
func NewSaleOrderMaterialRepo(db *gorm.DB) ISaleOrderMaterialRepo {
	return NewSaleOrderMaterialRepoImpl(db)
}

// NewSaleOrderMaterialRepoImpl 创建新的销售订单原料仓库实现
func NewSaleOrderMaterialRepoImpl(db *gorm.DB) *SaleOrderMaterialRepoImpl {
	return &SaleOrderMaterialRepoImpl{db: db}
}

type SaleOrderMaterialRepoImpl struct {
	db *gorm.DB // 数据库连接
}

func (r *SaleOrderMaterialRepoImpl) BatchInsertSaleOrderMaterial(saleOrderMaterials []*model.SaleOrderMaterial) error {
	if len(saleOrderMaterials) == 0 {
		return nil
	}
	return r.db.Model(&model.SaleOrderMaterial{}).Create(&saleOrderMaterials).Error
}

// DeleteSaleOrderMaterial 删除销售订单原料（按 sale_bill_uuid）
func (r *SaleOrderMaterialRepoImpl) DeleteSaleOrderMaterial(saleBillUuid uint64) error {
	return r.db.Model(&model.SaleOrderMaterial{}).Where("sale_bill_uuid = ? AND delete_time = ?", saleBillUuid, constant.NotDeleted).Update("delete_time", time.Now().Unix()).Error
}

// DeleteSaleOrderMaterialBySaleOrderUuid 删除销售订单原料（按 sale_order_uuid）
func (r *SaleOrderMaterialRepoImpl) DeleteSaleOrderMaterialBySaleOrderUuid(saleOrderUuid uint64) error {
	return r.db.Model(&model.SaleOrderMaterial{}).Where("sale_order_uuid = ? AND delete_time = ?", saleOrderUuid, constant.NotDeleted).Update("delete_time", time.Now().Unix()).Error
}

// GetSaleOrderMaterialBySaleOrderUuid 获取指定订单的材料记录
func (r *SaleOrderMaterialRepoImpl) GetSaleOrderMaterialBySaleOrderUuid(saleOrderUuid uint64) ([]*model.SaleOrderMaterial, error) {
	var saleOrderMaterials []*model.SaleOrderMaterial
	err := r.db.Model(&model.SaleOrderMaterial{}).
		Where("sale_order_uuid = ? AND delete_time = ?", saleOrderUuid, constant.NotDeleted).
		Find(&saleOrderMaterials).Error
	return saleOrderMaterials, errors.WithMessage(err)
}

// GetSaleOrderMaterialByCreateTimeBetween 获取某时间范围内的销售订单原料（仅未统计的）
func (r *SaleOrderMaterialRepoImpl) GetSaleOrderMaterialByCreateTimeBetween(startTime, endTime int64) ([]*model.SaleOrderMaterial, error) {
	var saleOrderMaterials []*model.SaleOrderMaterial
	err := r.db.Model(&model.SaleOrderMaterial{}).Preload("Material.Unit").Where("create_time BETWEEN ? AND ? AND delete_time = 0 AND is_summarized = 0", startTime, endTime).Find(&saleOrderMaterials).Error
	return saleOrderMaterials, errors.WithMessage(err)
}

// GetSaleOrderMaterialByCreateTimeBetweenAll 获取某时间范围内的销售订单原料（包含已统计的）
func (r *SaleOrderMaterialRepoImpl) GetSaleOrderMaterialByCreateTimeBetweenAll(startTime, endTime int64) ([]*model.SaleOrderMaterial, error) {
	var saleOrderMaterials []*model.SaleOrderMaterial
	err := r.db.Model(&model.SaleOrderMaterial{}).Preload("Material.Unit").Where("create_time BETWEEN ? AND ? AND delete_time = 0", startTime, endTime).Find(&saleOrderMaterials).Error
	return saleOrderMaterials, errors.WithMessage(err)
}

func (r *SaleOrderMaterialRepoImpl) UpdateSaleOrderMaterialIsSummarized(uuids []uint64) error {
	return r.db.Model(&model.SaleOrderMaterial{}).Where("uuid in (?)", uuids).Update("is_summarized", 1).Error
}

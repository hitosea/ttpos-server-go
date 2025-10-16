package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMaterialRepo 物品仓库接口
type ISaleOrderMaterialRepo interface {
	BatchInsertSaleOrderMaterial(saleOrderMaterials []*model.SaleOrderMaterial) error // 批量插入销售订单原料
	DeleteSaleOrderMaterial(saleBillUuid uint64) error                                // 删除销售订单原料
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

// DeleteSaleOrderMaterial 删除销售订单原料
func (r *SaleOrderMaterialRepoImpl) DeleteSaleOrderMaterial(saleBillUuid uint64) error {
	return r.db.Model(&model.SaleOrderMaterial{}).Where("sale_bill_uuid = ? AND delete_time = ?", saleBillUuid, constant.NotDeleted).Update("delete_time", time.Now().Unix()).Error
}

package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IOrderProductBomRepo 定义销售订单商品BOM仓库接口
type IOrderProductBomRepo interface {
	CreateBatch(boms []*model.SaleOrderProductBom) error // 批量创建
	CreateSaleOrderProductBomRecord(model model.SaleOrderProductBom) error
	UpdateSaleOrderProductBomRecord(model model.SaleOrderProductBom) error
	UpdateOrCreateSaleOrderProductBomRecord(model model.SaleOrderProductBom) error
}

// orderProductBomRepo 销售订单商品BOM仓库
type orderProductBomRepo struct {
	db *gorm.DB
}

// NewOrderProductBomRepo 实例化销售订单商品BOM仓库
func NewOrderProductBomRepo(db *gorm.DB) IOrderProductBomRepo {
	return NewOrderProductBomRepoImpl(db)
}

// NewOrderProductBomRepoImpl 实例化销售订单商品BOM仓库实现
func NewOrderProductBomRepoImpl(db *gorm.DB) IOrderProductBomRepo {
	return &orderProductBomRepo{db: db}
}

// CreateBatch 批量创建
func (r *orderProductBomRepo) CreateBatch(boms []*model.SaleOrderProductBom) error {
	return r.db.Create(&boms).Error
}

// CreateSaleOrderProductBomRecord 创建销售订单商品BOM记录
func (r *orderProductBomRepo) CreateSaleOrderProductBomRecord(obj model.SaleOrderProductBom) error {
	obj.SetNil()
	return r.db.Model(&model.SaleOrderProductBom{}).Create(&obj).Error
}

// UpdateSaleOrderProductBomRecord 更新销售订单商品BOM记录
func (r *orderProductBomRepo) UpdateSaleOrderProductBomRecord(obj model.SaleOrderProductBom) error {
	// 如果标记商品需要更新才更新该商品
	if obj.GetUpdate() {
		obj.SetNil()
		return r.db.Model(&model.SaleOrderProductBom{}).Select("*").Where("uuid = ?", obj.Uuid).Updates(&obj).Error
	}
	return nil
}

// UpdateOrCreateSaleOrderProductBomRecord 修改或新建销售订单商品BOM记录
// 如果没有uuid则新建，否则修改
func (r *orderProductBomRepo) UpdateOrCreateSaleOrderProductBomRecord(obj model.SaleOrderProductBom) error {
	// 如果主键id为0或uuid为0则create，否则update
	obj.SetNil()
	if obj.NoPrimaryKey() {
		return r.CreateSaleOrderProductBomRecord(obj)
	}
	return r.UpdateSaleOrderProductBomRecord(obj)
}

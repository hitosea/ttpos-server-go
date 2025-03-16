package repository

import (
	"fmt"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IOrderProductBomRepo 定义销售订单商品BOM仓库接口
type IOrderProductBomRepo interface {
	CreateBatch(boms []*model.SaleOrderProductBom) error // 批量创建
	UpdateSaleOrderProductBomRecord(model model.SaleOrderProductBom) error
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
func (o *orderProductBomRepo) CreateBatch(boms []*model.SaleOrderProductBom) error {
	return o.db.Create(&boms).Error
}

// UpdateSaleOrderProductBomRecord 更新销售订单商品BOM记录
func (r *orderProductBomRepo) UpdateSaleOrderProductBomRecord(obj model.SaleOrderProductBom) error {
	// 如果标记商品需要更新才更新该商品
	if obj.GetUpdate() {
		fmt.Println("qqqqq更新BOM", obj.Name, obj.Price)
		obj.SetNil()
		return r.db.Model(&model.SaleOrderProductBom{}).Select("*").Where("uuid = ?", obj.Uuid).Updates(&obj).Error
	}
	return nil
}

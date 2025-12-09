package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ISaleOrderProductReasonRepo 销售订单商品原因
type ISaleOrderProductReasonRepo interface {
	CreateSaleOrderProductReasons(reasons []*model.SaleOrderProductReason) error
	DeleteFreeReason(saleOrderUuid uint64) error                                          // 删除某销售订单的所有免单原因
	DeleteOrderItemRemarkReasons(saleOrderUuid uint64, saleOrderProductUuid uint64) error // 删除某订单商品的所有备注预设原因（物理删除）
}

func NewSaleOrderProductReasonRepo(db *gorm.DB) ISaleOrderProductReasonRepo {
	return NewSaleOrderProductReasonRepoImpl(db)
}

// NewSaleOrderProductReasonRepoImpl 创建新的销售订单商品原因仓库实现
func NewSaleOrderProductReasonRepoImpl(db *gorm.DB) *saleOrderProductReasonRepoImpl {
	return &saleOrderProductReasonRepoImpl{db: db}
}

type saleOrderProductReasonRepoImpl struct {
	db *gorm.DB
}

// CreateSaleOrderProductReasons 创建销售订单商品原因
func (r *saleOrderProductReasonRepoImpl) CreateSaleOrderProductReasons(reasons []*model.SaleOrderProductReason) error {
	for _, reason := range reasons {
		reason.SetNil()
	}
	err := r.db.Model(&model.SaleOrderProductReason{}).Create(reasons).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// DeleteFreeReason 删除某销售订单的所有免单原因
func (r *saleOrderProductReasonRepoImpl) DeleteFreeReason(saleOrderUuid uint64) error {
	err := r.db.Model(&model.SaleOrderProductReason{}).Where("sale_order_uuid = ? AND free_reason_uuid != 0", saleOrderUuid).Update("delete_time", time.Now().Unix()).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// DeleteOrderItemRemarkReasons 删除某订单商品的所有备注预设原因（物理删除）
func (r *saleOrderProductReasonRepoImpl) DeleteOrderItemRemarkReasons(saleOrderUuid uint64, saleOrderProductUuid uint64) error {
	err := r.db.Where("sale_order_uuid = ? AND sale_order_product_uuid = ? AND order_item_remark_uuid > 0", saleOrderUuid, saleOrderProductUuid).
		Delete(&model.SaleOrderProductReason{}).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

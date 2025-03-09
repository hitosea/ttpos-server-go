package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ISaleOrderProductReasonRepo 销售订单商品原因
type ISaleOrderProductReasonRepo interface {
	CreateSaleOrderProductReasons(reasons []*model.SaleOrderProductReason) error
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

// CreateSaleOrderProductReason 创建销售订单商品原因
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

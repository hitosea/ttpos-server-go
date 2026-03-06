package repository

import (
	"database/sql"

	"gorm.io/gorm"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
)

type IReturnOrderRepo interface {
	IReturnOrderQueryRepo
	CreateReturnOrder(order model.ReturnOrder) (model.ReturnOrder, error)           // 创建退货单
	CreateReturnOrderRecord(order model.ReturnOrder) (uint64, error)                // 创建退货单
	UpdateReturnOrderRecordErpInvoiceName(uuid uint64, erpInvoiceName string) error // 更新退货单erp发票名
	CreateReturnOrderAmount(amounts []model.ReturnOrderAmount) error                // 创建退货金额
	CreateReturnOrderProduct(products []*model.ReturnOrderProduct) error            // 创建退货商品
	UpdateReturnOrder(opts []DBOption, order model.ReturnOrder) error
	UpdateReturnOrderAmount(opts []DBOption, amount model.ReturnOrderAmount) error
	SumRefundAmount(opts ...DBOption) float64 // 统计退款金额

	WhereUuid(uuid uint64) DBOption                                   // 通过uuid查询
	WhereMerchantRefundOrderNo(merchantRefundOrderNo string) DBOption // 通过商户退货单号查询
	WhereNotReverseSettlement() DBOption                              // 通过不是反结账查询

	WithReturnOrder() DBOption // 预加载退货单
	WithPaymentMethod() DBOption
}

// IReturnOrderQueryRepo 退货单查询仓库接口
type IReturnOrderQueryRepo interface {
	GetReturnOrder(opts ...DBOption) (model.ReturnOrder, error)               // 获取退货单
	GetReturnOrderAmount(opts ...DBOption) (model.ReturnOrderAmount, error)   // 获取退货金额
	CountReturnOrders(opts ...DBOption) (int64, error)                        // 统计退货单数量
	SumRefundAmountWithSaleOrderJoin(saleBillUuids []uint64) (float64, error) // 统计退款金额（关联销售订单）
}

func NewReturnOrderRepo(db *gorm.DB) IReturnOrderRepo {
	return NewReturnOrderRepoImpl(db)
}

type returnOrderRepo struct {
	db *gorm.DB
}

func NewReturnOrderRepoImpl(db *gorm.DB) IReturnOrderRepo {
	return &returnOrderRepo{db: db}
}

func (r *returnOrderRepo) CreateReturnOrderRecord(order model.ReturnOrder) (uint64, error) {
	order.SetNil()
	err := r.db.Model(&model.ReturnOrder{}).Create(&order).Error
	return order.Uuid, errors.WithMessage(err)
}

func (r *returnOrderRepo) UpdateReturnOrderRecordErpInvoiceName(uuid uint64, erpInvoiceName string) error {
	err := r.db.Model(&model.ReturnOrder{}).Where("uuid = ?", uuid).Update("erp_invoice_name", erpInvoiceName).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *returnOrderRepo) CreateReturnOrder(order model.ReturnOrder) (model.ReturnOrder, error) {
	err := r.db.Model(&model.ReturnOrder{}).Create(&order).Error
	return order, errors.WithMessage(err)
}

func (r *returnOrderRepo) CreateReturnOrderAmount(amounts []model.ReturnOrderAmount) error {
	for i := range amounts {
		amounts[i].SetNil()
	}
	return r.db.Model(&model.ReturnOrderAmount{}).Create(&amounts).Error
}

func (r *returnOrderRepo) CreateReturnOrderProduct(products []*model.ReturnOrderProduct) error {
	if len(products) == 0 {
		return nil // 避免gorm报错empty slice found
	}
	for _, product := range products {
		product.SetNil()
	}
	return r.db.Model(&model.ReturnOrderProduct{}).Create(products).Error
}

func (r *returnOrderRepo) GetReturnOrder(opts ...DBOption) (model.ReturnOrder, error) {
	var order model.ReturnOrder
	db := r.db
	for _, w := range opts {
		db = w(db)
	}
	err := db.First(&order).Error
	return order, errors.WithMessage(err)
}

func (r *returnOrderRepo) GetReturnOrderAmount(opts ...DBOption) (model.ReturnOrderAmount, error) {
	var amount model.ReturnOrderAmount
	db := r.db
	for _, w := range opts {
		db = w(db)
	}
	err := db.First(&amount).Error
	return amount, errors.WithMessage(err)
}

// SumRefundAmount 统计退款金额
func (r *returnOrderRepo) SumRefundAmount(opts ...DBOption) float64 {
	var amount sql.NullFloat64
	db := r.db
	for _, w := range opts {
		db = w(db)
	}
	db.Model(&model.ReturnOrder{}).Select("SUM(refund_amount) as amount").Find(&amount)
	return amount.Float64
}

func (r *returnOrderRepo) UpdateReturnOrder(opts []DBOption, order model.ReturnOrder) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(&order).Error
}

func (r *returnOrderRepo) UpdateReturnOrderAmount(opts []DBOption, amount model.ReturnOrderAmount) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(&amount).Error
}

func (r *returnOrderRepo) WhereMerchantRefundOrderNo(merchantRefundOrderNo string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("merchant_refund_order_no = ?", merchantRefundOrderNo)
	}
}

func (r *returnOrderRepo) WithSaleOrder() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("SaleOrder")
	}
}

func (r *returnOrderRepo) WithReturnOrder() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ReturnOrder")
	}
}

func (r *returnOrderRepo) WithPaymentMethod() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("PaymentMethod")
	}
}

// CountReturnOrders 统计退货单数量
func (r *returnOrderRepo) CountReturnOrders(opts ...DBOption) (int64, error) {
	var count int64
	db := r.db.Model(&model.ReturnOrder{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&count).Error
	return count, err
}

// SumRefundAmountWithSaleOrderJoin 统计退款金额（关联销售订单，过滤已完成订单）
func (r *returnOrderRepo) SumRefundAmountWithSaleOrderJoin(saleBillUuids []uint64) (float64, error) {
	var result struct{ Amount float64 }
	err := r.db.Table("ttpos_return_order AS ro").
		Select("COALESCE(SUM(ro.refund_amount), 0) AS amount").
		Joins("INNER JOIN ttpos_sale_order AS so ON ro.related_order_uuid = so.uuid AND so.delete_time = ? AND so.status = ?", constant.NotDeleted, constant.SaleOrderStatusFinish).
		Where("ro.delete_time = ?", constant.NotDeleted).
		Where("ro.related_order_type = ?", constant.ReturnOrderRelatedOrderTypeSaleOrder).
		Where("so.sale_bill_uuid IN (?)", saleBillUuids).
		Scan(&result).Error
	return result.Amount, err
}

func (r *returnOrderRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

func (r *returnOrderRepo) WhereNotReverseSettlement() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_reverse_settlement = 0")
	}
}

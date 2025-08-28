package repository

import (
	"database/sql"

	"gorm.io/gorm"

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

	WithReturnOrder() DBOption // 预加载退货单
	WithPaymentMethod() DBOption
}

// IReturnOrderQueryRepo 退货单查询仓库接口
type IReturnOrderQueryRepo interface {
	GetReturnOrder(opts ...DBOption) (model.ReturnOrder, error)             // 获取退货单
	GetReturnOrderAmount(opts ...DBOption) (model.ReturnOrderAmount, error) // 获取退货金额
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
	for i, _ := range amounts {
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

func (r *returnOrderRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPaymentOrderRepo 定义仓库接口
type IPaymentOrderRepo interface {
	WhereRelatedUuid(uuid uint64) DBOption
	WhereRelatedUuids(uuids []uint64) DBOption
	WhereStatus(status uint) DBOption
	WherePaymentTypeUuid(uuid uint64) DBOption
	WhereUuid(uuid uint64) DBOption

	WithPaymentMethod() DBOption       // 关联支付方式
	WithMemberRechargeOrder() DBOption // 关联会员充值订单

	GetPaymentOrder(opts ...DBOption) (model.PaymentOrder, error) // 获取详细信息
	GetPaymentOrderList(opts ...DBOption) ([]model.PaymentOrder, error)

	Create(order model.PaymentOrder) (model.PaymentOrder, error) // 创建支付订单
	Update(uuid uint64, vars map[string]any) error               // 更新支付订单金额
}

// paymentOrderRepo 仓库
type paymentOrderRepo struct {
	db *gorm.DB
}

// NewPaymentOrderRepo 创建新仓库
func NewPaymentOrderRepo(db *gorm.DB) IPaymentOrderRepo {
	return NewPaymentOrderRepoImpl(db)
}

// NewPaymentOrderRepoImpl 创建新仓库实现
func NewPaymentOrderRepoImpl(db *gorm.DB) IPaymentOrderRepo {
	return &paymentOrderRepo{db: db}
}

// GetPaymentOrder 获取详细信息
func (r *paymentOrderRepo) GetPaymentOrder(opts ...DBOption) (model.PaymentOrder, error) {
	var paymentOrder model.PaymentOrder
	db := r.db
	for _, w := range opts {
		db = w(db)
	}
	db = CommonRepo.WhereBySoftDelete()(db)
	err := db.Order("id asc").First(&paymentOrder).Error
	return paymentOrder, err
}

// GetPaymentOrderList 获取列表
func (r *paymentOrderRepo) GetPaymentOrderList(opts ...DBOption) ([]model.PaymentOrder, error) {
	var paymentOrders []model.PaymentOrder
	db := r.db
	for _, w := range opts {
		db = w(db)
	}
	db = CommonRepo.WhereBySoftDelete()(db)
	err := db.Order("id asc").Find(&paymentOrders).Error
	return paymentOrders, err
}

func (r *paymentOrderRepo) Create(order model.PaymentOrder) (model.PaymentOrder, error) {
	err := r.db.Model(&model.PaymentOrder{}).Create(&order).Error
	return order, err
}

// Update 更新支付订单
func (r *paymentOrderRepo) Update(uuid uint64, vars map[string]any) error {
	err := r.db.Model(&model.PaymentOrder{}).Where("uuid = ?", uuid).Updates(vars).Error
	return err
}

// WhereRelatedUuid 根据related_uuid(销售订单、充值订单)查询
func (r *paymentOrderRepo) WhereRelatedUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("related_uuid = ?", uuid)
	}
}

// WherePaymentTypeUuid 根据支付方式Uuid查询
func (r *paymentOrderRepo) WherePaymentTypeUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("payment_method_uuid = ?", uuid)
	}
}

// WhereUuid 根据支付订单Uuid查询
func (r *paymentOrderRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereRelatedUuids 根据related_uuids(销售订单、充值订单)查询
func (r *paymentOrderRepo) WhereRelatedUuids(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("related_uuid in (?)", uuids)
	}
}

// WhereStatus 根据状态查询
func (r *paymentOrderRepo) WhereStatus(status uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WithPaymentMethod 关联支付方式
func (r *paymentOrderRepo) WithPaymentMethod() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("PaymentMethod")
	}
}

// WithMemberRechargeOrder 关联会员充值订单
func (r *paymentOrderRepo) WithMemberRechargeOrder() DBOption { // 关联会员充值订单
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberRechargeOrder")
	}
}

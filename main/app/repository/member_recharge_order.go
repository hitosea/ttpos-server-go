package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/model"
)

type IMemberRechargeOrderRepo interface {
	WithPaymentOrders() DBOption                                                       // 预加载支付订单
	WithPaymentOrderPaymentMethod() DBOption                                           // 预加载支付订单.支付方式
	WhereUuid(uuid uint64) DBOption                                                    // Uuid条件
	WhereStatus(status int) DBOption                                                   // 状态条件
	GetRechargeOrder(opts ...DBOption) model.MemberRechargeOrder                       // 获取充值订单
	Create(rechargeOrder model.MemberRechargeOrder) (model.MemberRechargeOrder, error) // 创建充值订单
	Update(uuid uint64, vars map[string]any) error                                     // 更新充值订单
}

func NewMemberRechargeOrderRepo(db *gorm.DB) IMemberRechargeOrderRepo {
	return NewMemberRechargeOrderRepoImpl(db)
}

type MemberRechargeOrderRepo struct {
	db *gorm.DB
}

func NewMemberRechargeOrderRepoImpl(db *gorm.DB) *MemberRechargeOrderRepo {
	return &MemberRechargeOrderRepo{db: db}
}

// GetRechargeOrder 获取充值订单
func (r *MemberRechargeOrderRepo) GetRechargeOrder(opts ...DBOption) model.MemberRechargeOrder {
	var rechargeOrder model.MemberRechargeOrder
	db := r.db.Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Model(&model.MemberRechargeOrder{}).First(&rechargeOrder)
	return rechargeOrder
}

// WithPaymentOrders 预加载支付订单
func (r *MemberRechargeOrderRepo) WithPaymentOrders() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("PaymentOrders", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted)
		})
	}
}

// WhereUuid uuid 条件
func (r *MemberRechargeOrderRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereStatus status 条件
func (r *MemberRechargeOrderRepo) WhereStatus(status int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WithPaymentOrderPaymentMethod 预加载支付订单.支付方式
func (r *MemberRechargeOrderRepo) WithPaymentOrderPaymentMethod() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("PaymentOrders.PaymentMethod")
	}
}

// Update 修改充值订单
func (r *MemberRechargeOrderRepo) Update(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.MemberRechargeOrder{}).Where("uuid = ?", uuid).Updates(vars).Error
}

// Create 创建充值订单
func (r *MemberRechargeOrderRepo) Create(rechargeOrder model.MemberRechargeOrder) (model.MemberRechargeOrder, error) {
	err := r.db.Model(&model.MemberRechargeOrder{}).Create(&rechargeOrder).Error
	return rechargeOrder, err
}

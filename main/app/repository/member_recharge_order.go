package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
)

type IMemberRechargeOrderRepo interface {
	WithPaymentOrders() With             // 预加载支付订单
	WithPaymentOrderPaymentMethod() With // 预加载支付订单.支付方式

	GetByUuid(uuid uint64, withs ...With) model.MemberRechargeOrder                    // 根据uuid获取进行中的充值订单
	GetPendingRechargeOrder(withs ...With) model.MemberRechargeOrder                   // 获取进行中的充值订单
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

// GetPendingRechargeOrder 获取充值订单
func (r *MemberRechargeOrderRepo) GetPendingRechargeOrder(withs ...With) model.MemberRechargeOrder {
	var rechargeOrder model.MemberRechargeOrder
	db := r.db.Scopes(NotDeleted)
	handleWiths(db, withs).Model(&model.MemberRechargeOrder{}).Where("status = ?", constant.RechargeOrderStatusPending).First(&rechargeOrder)
	return rechargeOrder
}

// GetByUuid 获取充值订单
func (r *MemberRechargeOrderRepo) GetByUuid(uuid uint64, withs ...With) model.MemberRechargeOrder {
	var rechargeOrder model.MemberRechargeOrder
	db := r.db.Scopes(NotDeleted)
	handleWiths(db, withs).Model(&model.MemberRechargeOrder{}).Where("uuid = ? ", uuid).First(&rechargeOrder)
	return rechargeOrder
}

// WithPaymentOrders 预加载支付订单
func (r *MemberRechargeOrderRepo) WithPaymentOrders() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("PaymentOrders", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(NotDeleted)
		})
	}
}

// WithPaymentOrderPaymentMethod 预加载支付订单.支付方式
func (r *MemberRechargeOrderRepo) WithPaymentOrderPaymentMethod() With {
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

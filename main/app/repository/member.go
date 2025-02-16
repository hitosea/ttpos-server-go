package repository

import (
	"gorm.io/gorm"
	"time"
	"ttpos-server-go/app/constant"

	"ttpos-server-go/app/model"
)

type IMemberRepo interface {
	GetMemberLevels() []model.MemberLevel                                                           // 获取会员等级
	SearchMember(keyword string) []model.Member                                                     // 关键字搜索会员
	CreateMember(member model.Member) error                                                         // 添加会员
	CheckMemberExists(phone string) bool                                                            // 根据手机号检查是否存在
	CheckLevelExists(uuid uint64) bool                                                              // 根据Uuid检查等级是否存在
	GetByUuid(uuid uint64, withs ...With) model.Member                                              // 根据Uuid查询会员
	GetPendingRechargeOrder(uuid uint64, withs ...With) model.MemberRechargeOrder                   // 获取进行中的充值订单
	UpdateRechargeOrder(uuid uint64, vars map[string]any) error                                     // 修改充值订单
	CreateRechargeOrder(rechargeOrder model.MemberRechargeOrder) (model.MemberRechargeOrder, error) // 创建充值订单
	CancelPaymentOrder(paymentOrderUuid uint64) error                                               // 撤销充值支付订单

	WithMemberLevel() With    // Member 预加载会员等级
	WithMemberCard() With     // Member 预加载会员卡
	WithMemberCardType() With // Member 预加载会员卡.卡类型

	WithPaymentOrder() With              // MemberRechargeOrder 预加载充值订单
	WithMember() With                    // MemberRechargeOrder 预加载会员
	WithPaymentOrderPaymentMethod() With // MemberRechargeOrder 预加载充值订单支付方式
	WithPaidPaymentOrder() With          // MemberRechargeOrder 预加载已支付充值订单
}

func NewMemberRepo(db *gorm.DB) IMemberRepo {
	return NewMemberRepoImpl(db)
}

type MemberRepo struct {
	db *gorm.DB
}

func NewMemberRepoImpl(db *gorm.DB) *MemberRepo {
	return &MemberRepo{db: db}
}

// GetMemberLevels 获取会员等级
func (r *MemberRepo) GetMemberLevels() []model.MemberLevel {
	var levels []model.MemberLevel
	r.db.Scopes(NotDeleted).Model(&model.MemberLevel{}).Select("uuid, name, priority, create_time").Order("priority asc, create_time asc").Find(&levels)
	return levels
}

// SearchMember 关键字搜索会员
func (r *MemberRepo) SearchMember(keyword string) []model.Member {
	var members []model.Member
	keyword = Like(keyword)
	r.db.Scopes(NotDeleted).Model(&model.Member{}).Select("uuid, nickname, phone").Where("phone LIKE ? OR uuid LIKE ?", keyword, keyword).Find(&members)
	return members
}

// CreateMember 添加会员
func (r *MemberRepo) CreateMember(member model.Member) error {
	return r.db.Create(&member).Error
}

// CheckMemberExists 根据手机号检查是否存在
func (r *MemberRepo) CheckMemberExists(phone string) bool {
	var memberUuid uint64
	r.db.Scopes(NotDeleted).Model(&model.Member{}).Where("phone = ?", phone).Select("uuid").Scan(&memberUuid)
	return memberUuid > 0
}

// CheckLevelExists 根据uuid检查等级是否存在
func (r *MemberRepo) CheckLevelExists(uuid uint64) bool {
	var exists uint64
	r.db.Scopes(NotDeleted).Model(&model.MemberLevel{}).Where("uuid = ?", uuid).Select("uuid").Scan(&exists)
	return exists > 0
}

// GetByUuid 根据uuid查询会员
func (r *MemberRepo) GetByUuid(uuid uint64, withs ...With) model.Member {
	var member model.Member
	handleWiths(r.db, withs).Scopes(NotDeleted).Model(&model.Member{}).Where("uuid = ?", uuid).Debug().First(&member)
	return member
}

// GetPendingRechargeOrder 获取进行中的充值订单
func (r *MemberRepo) GetPendingRechargeOrder(uuid uint64, withs ...With) model.MemberRechargeOrder {
	var rechargeOrder model.MemberRechargeOrder
	builder := handleWiths(r.db, withs).Model(&model.MemberRechargeOrder{}).Where("status = ?", constant.RechargeOrderStatusPending)
	if uuid != 0 {
		builder.Where("uuid = ?", uuid)
	}
	return rechargeOrder
}

// UpdateRechargeOrder 修改充值订单
func (r *MemberRepo) UpdateRechargeOrder(uuid uint64, vars map[string]any) error {
	return r.db.Model(&model.MemberRechargeOrder{}).Where("uuid = ?", uuid).Updates(vars).Error
}

// CreateRechargeOrder 创建充值订单
func (r *MemberRepo) CreateRechargeOrder(rechargeOrder model.MemberRechargeOrder) (model.MemberRechargeOrder, error) {
	err := r.db.Model(&model.MemberRechargeOrder{}).Create(&rechargeOrder).Error
	return rechargeOrder, err
}

// CancelPaymentOrder 撤销充值订单
func (r *MemberRepo) CancelPaymentOrder(paymentOrderUuid uint64) error {
	err := r.db.Model(&model.PaymentOrder{}).Where("uuid = ?", paymentOrderUuid).Updates(map[string]any{
		"delete_time": time.Now().Unix(),
	}).Error
	return err
}

// WithMemberLevel Member 预加载会员等级
func (r *MemberRepo) WithMemberLevel() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberLevel")
	}
}

// WithMemberCard Member 预加载会员卡
func (r *MemberRepo) WithMemberCard() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberCard")
	}
}

// WithMemberCardType Member 预加载会员卡.卡类型
func (r *MemberRepo) WithMemberCardType() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberCard.MemberCardType")
	}
}

// WithPaymentOrder MemberRechargeOrder 预加载充值订单
func (r *MemberRepo) WithPaymentOrder() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("PaymentOrder")
	}
}

// WithPaidPaymentOrder MemberRechargeOrder 预加载已支付充值订单
func (r *MemberRepo) WithPaidPaymentOrder() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("PaymentOrder", "status = ?", constant.PaymentOrderStatusPaid)
	}
}

// WithMember MemberRechargeOrder 预加载会员
func (r *MemberRepo) WithMember() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Member")
	}
}

// WithPaymentOrderPaymentMethod MemberRechargeOrder 预加载充值订单支付方式
func (r *MemberRepo) WithPaymentOrderPaymentMethod() With {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("PaymentOrder.PaymentMethod")
	}
}

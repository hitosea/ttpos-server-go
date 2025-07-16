package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IMemberSaleOrderRepo interface {
	GetMemberSaleOrder(opts ...DBOption) (*model.MemberSaleOrder, error)
	GetMemberSaleOrderRecord(uuid uint64) (*model.MemberSaleOrder, error)             // 获取会员端销售订单记录
	CreateMemberSaleOrder(memberSaleOrder model.MemberSaleOrder) error                // 创建会员端销售订单
	UpdateMemberSaleOrderVerifiedPhoneStatus(memberSaleOrderUuid uint64) error        // 更新会员端销售订单的手机号验证状态为已验证
	UpdateMemberSaleOrderPendingPayment(memberSaleOrder *model.MemberSaleOrder) error // 更新会员端销售订单为待支付状态
}

func NewMemberSaleOrderRepo(db *gorm.DB) IMemberSaleOrderRepo {
	return NewMemberSaleOrderRepoImpl(db)
}

type MemberSaleOrderRepo struct {
	db *gorm.DB
}

func NewMemberSaleOrderRepoImpl(db *gorm.DB) *MemberSaleOrderRepo {
	return &MemberSaleOrderRepo{db: db}
}

func (r *MemberSaleOrderRepo) GetMemberSaleOrder(opts ...DBOption) (*model.MemberSaleOrder, error) {
	var memberSaleOrder model.MemberSaleOrder
	db := r.db.Model(&model.MemberSaleOrder{})

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&memberSaleOrder).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &memberSaleOrder, nil
}

func (r *MemberSaleOrderRepo) GetMemberSaleOrderRecord(uuid uint64) (*model.MemberSaleOrder, error) {
	memberSaleOrder, err := r.GetMemberSaleOrder(
		CommonRepo.WhereByUuid(uuid),
		CommonRepo.WhereBySoftDelete(),
		CommonRepo.Preload(
			WithPreload{
				Query: "Address",
				Args: []any{
					CommonRepo.DBOption(CommonRepo.WhereBySoftDelete()),
				},
			},
		),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return memberSaleOrder, nil
}

func (r *MemberSaleOrderRepo) CreateMemberSaleOrder(memberSaleOrder model.MemberSaleOrder) error {
	memberSaleOrder.SetNil()
	if err := r.db.Create(&memberSaleOrder).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateMemberSaleOrderVerifiedPhoneStatus 更新会员端销售订单的手机号验证状态为已验证
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderVerifiedPhoneStatus(memberSaleOrderUuid uint64) error {
	if err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrderUuid).Update("is_verified_phone", 1).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// UpdateMemberSaleOrderPendingPayment 更新会员端销售订单为待支付状态
func (r *MemberSaleOrderRepo) UpdateMemberSaleOrderPendingPayment(memberSaleOrder *model.MemberSaleOrder) error {
	if err := r.db.Model(&model.MemberSaleOrder{}).Where("uuid = ?", memberSaleOrder.Uuid).Updates(model.MemberSaleOrder{
		PaymentMethodUuid: memberSaleOrder.PaymentMethodUuid,            // 更新支付方式UUID
		Status:            constant.MemberSaleOrderStatusPendingPayment, // 更新订单状态为待支付
		Remark:            memberSaleOrder.Remark,                       // 更新订单备注
	}).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

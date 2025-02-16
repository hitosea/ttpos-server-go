package repository

import (
	"gorm.io/gorm"

	"ttpos-server-go/app/model"
)

type IMemberRechargeOperationRepo interface {
	Add(log model.MemberRechargeOrderOperationLog) error // 添加会员充值操作日志
}

func NewMemberRechargeOperationRepo(db *gorm.DB) IMemberRechargeOperationRepo {
	return NewMemberRechargeOperationRepoImpl(db)
}

type MemberRechargeOperationRepo struct {
	db *gorm.DB
}

func NewMemberRechargeOperationRepoImpl(db *gorm.DB) *MemberRechargeOperationRepo {
	return &MemberRechargeOperationRepo{db: db}
}

// Add 添加会员充值操作日志
func (r *MemberRechargeOperationRepo) Add(log model.MemberRechargeOrderOperationLog) error {
	return r.db.Create(&log).Error
}

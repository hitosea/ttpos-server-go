package repository

import (
	"gorm.io/gorm"

	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
)

type IMemberRechargeOperationRepo interface {
	AddLog(log model.MemberRechargeOrderOperationLog) error // 添加会员充值操作日志
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

// AddLog 添加会员充值操作日志
func (r *MemberRechargeOperationRepo) AddLog(log model.MemberRechargeOrderOperationLog) error {
	// 设置默认值
	if err := r.db.Model(&model.MemberRechargeOrderOperationLog{}).Create(&log).Error; err != nil {
		return errors.WithMessage(err)
	}
	// 添加异常日志
	NewMemberRechargeAbnormalRecordRepo(r.db).CreateRechargeAbnormalLog(log)
	// 返回
	return nil
}

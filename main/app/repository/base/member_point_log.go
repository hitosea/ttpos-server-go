package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMemberPointLogRepo 会员积分变动记录
type IMemberPointLogRepo interface {
	GetMemberPointLogList() ([]model.MemberPointLog, error)
	UpdateMemberPointLog(uuid uint, memberPointLog model.MemberPointLog) error
	CreateMemberPointLog(memberPointLog model.MemberPointLog) (uint64, error)
	DeleteMemberPointLog(uuid uint) error
}

func NewMemberPointLogRepo(db *gorm.DB) IMemberPointLogRepo {
	return NewMemberPointLogRepoImpl(db)
}

// NewMemberPointLogRepoImpl 创建新的会员积分变动记录仓库实现
func NewMemberPointLogRepoImpl(db *gorm.DB) *MemberPointLogRepo {
	return &MemberPointLogRepo{db: db}
}

type MemberPointLogRepo struct {
	db *gorm.DB
}

// GetMemberPointLogList 获取会员积分变动记录列表
func (r *MemberPointLogRepo) GetMemberPointLogList() ([]model.MemberPointLog, error) {
	var memberPointLogs []model.MemberPointLog
	err := r.db.Model(&model.MemberPointLog{}).Where("delete_time = ?", 0).Find(&memberPointLogs).Error
	return memberPointLogs, err
}

// UpdateMemberPointLog 更新会员积分变动记录
func (r *MemberPointLogRepo) UpdateMemberPointLog(uuid uint, memberPointLog model.MemberPointLog) error {
	if err := r.db.Model(&model.MemberPointLog{}).Where("uuid = ?", uuid).Updates(memberPointLog).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// CreateMemberPointLog 创建会员积分变动记录
func (r *MemberPointLogRepo) CreateMemberPointLog(memberPointLog model.MemberPointLog) (uint64, error) {
	// 创建会员积分变动记录
	if err := r.db.Create(&memberPointLog).Error; err != nil {
		return 0, err
	}
	return memberPointLog.Uuid, nil
}

// DeleteMemberPointLog 软删除会员积分变动记录
func (r *MemberPointLogRepo) DeleteMemberPointLog(uuid uint) error {
	return r.db.Model(&model.MemberPointLog{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}

package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMemberCardLogRepo 会员卡领取记录
type IMemberCardLogRepo interface {
	GetMemberCardLogList() ([]model.MemberCardLog, error)
	UpdateMemberCardLog(uuid uint, memberCardLog model.MemberCardLog) error
	CreateMemberCardLog(memberCardLog model.MemberCardLog) (uint64, error)
	DeleteMemberCardLog(uuid uint) error
}

func NewMemberCardLogRepo(db *gorm.DB) IMemberCardLogRepo {
	return NewMemberCardLogRepoImpl(db)
}

// NewMemberCardLogRepoImpl 创建新的会员卡领取记录仓库实现
func NewMemberCardLogRepoImpl(db *gorm.DB) *MemberCardLogRepo {
	return &MemberCardLogRepo{db: db}
}

type MemberCardLogRepo struct {
	db *gorm.DB
}

// GetMemberCardLogList 获取会员卡领取记录列表，排除逻辑删除的记录
func (r *MemberCardLogRepo) GetMemberCardLogList() ([]model.MemberCardLog, error) {
	var memberCardLogs []model.MemberCardLog
	err := r.db.Model(&model.MemberCardLog{}).Where("delete_time = ?", 0).Find(&memberCardLogs).Error
	return memberCardLogs, err
}

// UpdateMemberCardLog 更新会员卡领取记录
func (r *MemberCardLogRepo) UpdateMemberCardLog(uuid uint, memberCardLog model.MemberCardLog) error {
	if err := r.db.Model(&model.MemberCardLog{}).Where("uuid = ?", uuid).Updates(memberCardLog).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// CreateMemberCardLog 创建会员卡领取记录
func (r *MemberCardLogRepo) CreateMemberCardLog(memberCardLog model.MemberCardLog) (uint64, error) {
	// 创建会员卡领取记录
	if err := r.db.Create(&memberCardLog).Error; err != nil {
		return 0, err
	}
	return memberCardLog.Uuid, nil
}

// DeleteMemberCardLog 软删除会员卡领取记录
func (r *MemberCardLogRepo) DeleteMemberCardLog(uuid uint) error {
	return r.db.Model(&model.MemberCardLog{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}

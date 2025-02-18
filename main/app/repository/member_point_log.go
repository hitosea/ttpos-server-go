package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/model"
)

type IMemberPointLogRepo interface {
	Create(log model.MemberPointLog) (model.MemberPointLog, error) // 创建会员积分日志

}

func NewMemberPointLogRepo(db *gorm.DB) IMemberPointLogRepo {
	return NewMemberPointLogRepoImpl(db)
}

type MemberPointLogRepo struct {
	db *gorm.DB
}

func NewMemberPointLogRepoImpl(db *gorm.DB) *MemberPointLogRepo {
	return &MemberPointLogRepo{db: db}
}

// Create 创建会员积分日志
func (r *MemberPointLogRepo) Create(log model.MemberPointLog) (model.MemberPointLog, error) {
	err := r.db.Model(&model.MemberPointLog{}).Create(&log).Error
	return log, err
}

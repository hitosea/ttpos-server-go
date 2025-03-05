package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IMemberLevelLogRepo interface {
	Create(log model.MemberLevelLog) (model.MemberLevelLog, error) // 创建会员积分日志

}

func NewMemberLevelLogRepo(db *gorm.DB) IMemberLevelLogRepo {
	return NewMemberLevelLogRepoImpl(db)
}

type MemberLevelLogRepo struct {
	db *gorm.DB
}

func NewMemberLevelLogRepoImpl(db *gorm.DB) *MemberLevelLogRepo {
	return &MemberLevelLogRepo{db: db}
}

// Create 创建会员余额日志
func (r *MemberLevelLogRepo) Create(log model.MemberLevelLog) (model.MemberLevelLog, error) {
	err := r.db.Model(&model.MemberLevelLog{}).Create(&log).Error
	return log, errors.WithMessage(err)
}

package repository

import (
	"gorm.io/gorm"
	"ttpos-server-go/app/model"
)

type IMemberBalanceLogRepo interface {
	Create(log model.MemberBalanceLog) (model.MemberBalanceLog, error) // 创建会员积分日志

}

func NewMemberBalanceLogRepo(db *gorm.DB) IMemberBalanceLogRepo {
	return NewMemberBalanceLogRepoImpl(db)
}

type MemberBalanceLogRepo struct {
	db *gorm.DB
}

func NewMemberBalanceLogRepoImpl(db *gorm.DB) *MemberBalanceLogRepo {
	return &MemberBalanceLogRepo{db: db}
}

// Create 创建会员余额日志
func (r *MemberBalanceLogRepo) Create(log model.MemberBalanceLog) (model.MemberBalanceLog, error) {
	err := r.db.Model(&model.MemberBalanceLog{}).Create(&log).Error
	return log, err
}

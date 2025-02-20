package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMemberBalanceLogRepo 会员余额变动记录
type IMemberBalanceLogRepo interface {
	GetMemberBalanceLogList() ([]model.MemberBalanceLog, error)
	UpdateMemberBalanceLog(uuid uint, memberBalanceLog model.MemberBalanceLog) error
	CreateMemberBalanceLog(memberBalanceLog model.MemberBalanceLog) (uint64, error)
	DeleteMemberBalanceLog(uuid uint) error
}

func NewMemberBalanceLogRepo(db *gorm.DB) IMemberBalanceLogRepo {
	return NewMemberBalanceLogRepoImpl(db)
}

// NewMemberBalanceLogRepoImpl 创建新的会员余额变动记录仓库实现
func NewMemberBalanceLogRepoImpl(db *gorm.DB) *MemberBalanceLogRepo {
	return &MemberBalanceLogRepo{db: db}
}

type MemberBalanceLogRepo struct {
	db *gorm.DB
}

// GetMemberBalanceLogList 获取商品规格列表，排除逻辑删除的规格
func (r *MemberBalanceLogRepo) GetMemberBalanceLogList() ([]model.MemberBalanceLog, error) {
	var memberBalanceLogs []model.MemberBalanceLog
	err := r.db.Model(&model.MemberBalanceLog{}).Where("delete_time = ?", 0).Find(&memberBalanceLogs).Error
	return memberBalanceLogs, err
}

// UpdateMemberBalanceLog 更新自助餐客户类型
func (r *MemberBalanceLogRepo) UpdateMemberBalanceLog(uuid uint, memberBalanceLog model.MemberBalanceLog) error {
	if err := r.db.Model(&model.MemberBalanceLog{}).Where("uuid = ?", uuid).Updates(memberBalanceLog).Error; err != nil {
		return err
	}
	return nil
}

// CreateMemberBalanceLog 创建会员余额变动记录
func (r *MemberBalanceLogRepo) CreateMemberBalanceLog(memberBalanceLog model.MemberBalanceLog) (uint64, error) {
	// 创建会员余额变动记录
	if err := r.db.Create(&memberBalanceLog).Error; err != nil {
		return 0, err
	}
	return memberBalanceLog.Uuid, nil
}

// DeleteMemberBalanceLog 软删除会员余额变动记录
func (r *MemberBalanceLogRepo) DeleteMemberBalanceLog(uuid uint) error {
	return r.db.Model(&model.MemberBalanceLog{}).Where("uuid = ?", uuid).Update("delete_time", uint(time.Now().Unix())).Error
}

package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IMemberBalanceLogRepo interface {
	IMemberBalanceLogQueryRepo
	Create(log model.MemberBalanceLog) (model.MemberBalanceLog, error)            // 创建会员积分日志
	ReverseSettleDelete(memberUuid uint64, saleOrderUuid uint64, scene int) error // 反结账时删除充值记录
}

// IMemberBalanceLogQueryRepo 会员余额日志查询
type IMemberBalanceLogQueryRepo interface {
	GetMemberBalanceLogList(opts ...DBOption) ([]*model.MemberBalanceLog, error) // 获取会员余额日志列表
	GetMemberBalanceLogNotProcessed() ([]*model.MemberBalanceLog, error)         // 获取未处理的会员余额日志. 用于处理会员余额变动
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
	return log, errors.WithMessage(err)
}

// ReverseSettleDelete 反结账时删除充值记录
func (r *MemberBalanceLogRepo) ReverseSettleDelete(memberUuid uint64, saleOrderUuid uint64, scene int) error {
	return r.db.Model(&model.MemberBalanceLog{}).
		Where("member_uuid = ? AND related_uuid = ? AND scene = ?", memberUuid, saleOrderUuid, scene).
		Updates(map[string]any{
			"delete_time": time.Now().Unix(),
		}).Error
}

// GetMemberBalanceLogList 获取会员余额日志列表
func (r *MemberBalanceLogRepo) GetMemberBalanceLogList(opts ...DBOption) ([]*model.MemberBalanceLog, error) {
	var logs []*model.MemberBalanceLog
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	result := db.Find(&logs)
	if result.Error != nil {
		return nil, errors.WithMessage(result.Error)
	}
	return logs, nil
}

// GetMemberBalanceLogNotProcessed 获取未处理的会员余额日志. 用于处理会员余额变动
func (r *MemberBalanceLogRepo) GetMemberBalanceLogNotProcessed() ([]*model.MemberBalanceLog, error) {
	logs, err := r.GetMemberBalanceLogList(
		CommonRepo.WhereByProcessedNot(),
		CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return logs, nil
}

package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IMemberPointLogRepo interface {
	Create(log model.MemberPointLog) (model.MemberPointLog, error)       // 创建会员积分日志
	GetMemberPoint(opts ...DBOption) (*model.MemberPointLog, error)      // 获取会员积分日志
	GetMemberPointList(opts ...DBOption) ([]model.MemberPointLog, error) // 获取会员积分日志列表
	GetMemberPointLogNotProcessed() ([]model.MemberPointLog, error)      // 获取未处理的会员积分日志. 用于处理积分变动
	UpdateProcessed(uuids []uint64) error                                // 更新会员积分日志为已处理
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
	return log, errors.WithMessage(err)
}

func (r *MemberPointLogRepo) GetMemberPoint(opts ...DBOption) (*model.MemberPointLog, error) {
	var log model.MemberPointLog
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	result := db.First(&log)
	if result.Error != nil {
		return nil, errors.WithMessage(result.Error)
	}
	return &log, nil
}

func (r *MemberPointLogRepo) GetMemberPointList(opts ...DBOption) ([]model.MemberPointLog, error) {
	var logs []model.MemberPointLog
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

// GetMemberPointLogNotProcessed 获取未处理的会员积分日志. 用于处理积分变动
func (r *MemberPointLogRepo) GetMemberPointLogNotProcessed() ([]model.MemberPointLog, error) {
	logs, err := r.GetMemberPointList(
		CommonRepo.WhereByProcessedNot(),
		CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return logs, nil
}

// UpdateProcessed 更新会员积分日志为已处理
func (r *MemberPointLogRepo) UpdateProcessed(uuids []uint64) error {
	if err := r.db.Model(&model.MemberPointLog{}).Where("uuid IN (?)", uuids).Update("processed", constant.MemberPointLogProcessedSuccess).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

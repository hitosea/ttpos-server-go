package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IMemberPointLogRepo interface {
	IMemberPointLogQueryRepo
	Create(log model.MemberPointLog) (model.MemberPointLog, error) // 创建会员积分日志
	UpdateProcessed(uuids []uint64) error                          // 更新会员积分日志为已处理
}

// IMemberPointLogQueryRepo 会员积分日志查询
type IMemberPointLogQueryRepo interface {
	GetMemberPoint(opts ...DBOption) (*model.MemberPointLog, error)                          // 获取会员积分日志
	GetMemberPointList(opts ...DBOption) ([]model.MemberPointLog, error)                     // 获取会员积分日志列表
	PaginateGet(page, pageSize int, opts ...DBOption) ([]model.MemberPointLog, int64, error) // 分页获取会员积分日志列表
	GetMemberPointLogNotProcessed() ([]model.MemberPointLog, error)                          // 获取未处理的会员积分日志. 用于处理积分变动

	WhereByPositiveValue() DBOption // 根据正数积分查询
	WhereByNegativeValue() DBOption // 根据负数积分查询
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
	if err := r.db.Model(&model.MemberPointLog{}).Where("uuid IN (?)", uuids).Update("processed", constant.MemberPointLogOrBalanceProcessedSuccess).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

func (r *MemberPointLogRepo) PaginateGet(page int, pageSize int, opts ...DBOption) ([]model.MemberPointLog, int64, error) {
	var logs []model.MemberPointLog
	var total int64
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	db = db.Order("id DESC")

	if err := db.Model(&model.MemberPointLog{}).Count(&total).Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	return logs, total, nil
}

// WhereByPositiveValue 根据正数积分查询
func (r *MemberPointLogRepo) WhereByPositiveValue() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("value > 0")
	}
}

// WhereByNegativeValue 根据负数积分查询
func (r *MemberPointLogRepo) WhereByNegativeValue() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("value < 0")
	}
}

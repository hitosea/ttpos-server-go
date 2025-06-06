package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IMarketingActivityRecordRepo interface {
	// CreateRecord 创建奖励记录
	CreateRecord(record *model.MarketingActivityRecord) error
	// GetByActivityAndMember 根据活动ID和会员ID获取奖励记录
	GetByActivityAndMember(activityUuid, memberUuid uint64) (*model.MarketingActivityRecord, error)
	// GetRewardCount 获取奖励次数
	GetRewardCount(activityUuid uint64, memberUuid uint64) (int64, error)
}

// MarketingActivityRecordRepo 营销活动奖励记录仓库
type MarketingActivityRecordRepo struct {
	db *gorm.DB
}

// NewMarketingActivityRecordRepo 创建营销活动奖励记录仓库
func NewMarketingActivityRecordRepo(db *gorm.DB) IMarketingActivityRecordRepo {
	return &MarketingActivityRecordRepo{
		db: db,
	}
}

// CreateRecord 创建奖励记录
func (r *MarketingActivityRecordRepo) CreateRecord(record *model.MarketingActivityRecord) error {
	return r.db.Create(record).Error
}

// GetByActivityAndMember 根据活动ID和会员ID获取奖励记录
func (r *MarketingActivityRecordRepo) GetByActivityAndMember(activityUuid, memberUuid uint64) (*model.MarketingActivityRecord, error) {
	var record model.MarketingActivityRecord
	err := r.db.Where("activity_uuid = ? AND member_uuid = ? AND delete_time = 0", activityUuid, memberUuid).First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// GetRewardCount 获取奖励次数
func (r *MarketingActivityRecordRepo) GetRewardCount(activityUuid uint64, memberUuid uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.MarketingActivityRecord{}).Where("activity_uuid = ? AND member_uuid = ? AND delete_time = 0", activityUuid, memberUuid).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

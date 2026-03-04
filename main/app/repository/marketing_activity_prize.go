package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMarketingActivityPrizeRepo 营销活动奖品仓库接口
type IMarketingActivityPrizeRepo interface {
	UnscopedDeleteByActivityUuid(activityUuid uint64) error
}

// MarketingActivityPrizeRepo 营销活动奖品仓库
type MarketingActivityPrizeRepo struct {
	db *gorm.DB
}

// NewMarketingActivityPrizeRepo 创建营销活动奖品仓库
func NewMarketingActivityPrizeRepo(db *gorm.DB) IMarketingActivityPrizeRepo {
	return &MarketingActivityPrizeRepo{db: db}
}

// UnscopedDeleteByActivityUuid 根据活动UUID硬删除所有奖品
func (r *MarketingActivityPrizeRepo) UnscopedDeleteByActivityUuid(activityUuid uint64) error {
	return r.db.Unscoped().Where("activity_uuid = ?", activityUuid).Delete(&model.MarketingActivityPrize{}).Error
}

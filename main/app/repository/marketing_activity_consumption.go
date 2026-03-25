package repository

import (
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IMarketingActivityConsumptionRepo interface {
	// CreateConsumption 创建消费记录
	CreateConsumption(consumption *model.MarketingActivityConsumption) error
	// 创建与更新记录
	CreateOrUpdateConsumption(activityUuid, referrerUuid, consumerUuid uint64, amount float64) error
	// GetByActivityAndReferrer 根据活动ID、推荐人ID获取消费记录
	GetByActivityAndReferrer(activityUuid, referrerUuid uint64) (*model.MarketingActivityConsumption, error)
	// GetByActivityAndReferrerAndConsumer 根据活动ID、推荐人ID和消费者ID获取消费记录
	GetByActivityAndReferrerAndConsumer(activityUuid, referrerUuid, consumerUuid uint64) (*model.MarketingActivityConsumption, error)
	// UpdateAmount 更新消费金额
	UpdateAmount(uuid uint64, amount float64) error
	// GetByActivityAndReferrerConsumptionAmount 根据活动ID、推荐人ID获取总消费金额
	GetByActivityAndReferrerConsumptionAmount(activityUuid, referrerUuid uint64) (float64, error)
	// UpdateRewardStatus 更新奖励状态
	UpdateRewardStatus(activityUuid, referrerUuid uint64) error
}

// MarketingActivityConsumptionRepo 营销活动消费记录仓库
type MarketingActivityConsumptionRepo struct {
	db *gorm.DB
}

// NewMarketingActivityConsumptionRepo 创建营销活动消费记录仓库
func NewMarketingActivityConsumptionRepo(db *gorm.DB) IMarketingActivityConsumptionRepo {
	return &MarketingActivityConsumptionRepo{
		db: db,
	}
}

// CreateConsumption 创建消费记录
func (r *MarketingActivityConsumptionRepo) CreateConsumption(consumption *model.MarketingActivityConsumption) error {
	err := r.db.Create(consumption).Error
	if err != nil {
		return err
	}
	return nil
}

// CreateOrUpdateConsumption 创建与更新记录
func (r *MarketingActivityConsumptionRepo) CreateOrUpdateConsumption(activityUuid, referrerUuid, consumerUuid uint64, amount float64) error {
	// 查询是否存在记录
	consumption, err := r.GetByActivityAndReferrerAndConsumer(
		activityUuid,
		referrerUuid,
		consumerUuid,
	)
	if err != nil {
		return err
	}
	if consumption != nil {
		// 存在则更新金额
		err = r.UpdateAmount(
			consumption.Uuid,
			amount,
		)
		if err != nil {
			logger.Logger.Error("SubscribeCheckoutSaleOrderEvent process, UpdateAmount failed", zap.Any("consumption", consumption), zap.Error(err))
		}
	} else {
		// 不存在则新增
		err = r.CreateConsumption(&model.MarketingActivityConsumption{
			ActivityUuid:      activityUuid,
			ReferrerUuid:      referrerUuid,
			ConsumerUuid:      consumerUuid,
			ConsumptionAmount: amount,
		})
		if err != nil {
			logger.Logger.Error("SubscribeCheckoutSaleOrderEvent process, CreateConsumption failed", zap.Any("activityUuid", activityUuid), zap.Any("referrerUuid", referrerUuid), zap.Any("consumerUuid", consumerUuid), zap.Error(err))
		}
	}
	if err != nil {
		return err
	}
	return nil
}

// GetByActivityAndReferrer 根据活动ID、推荐人ID获取总消费金额
func (r *MarketingActivityConsumptionRepo) GetByActivityAndReferrerConsumptionAmount(activityUuid, referrerUuid uint64) (float64, error) {
	var amount float64
	err := r.db.Model(&model.MarketingActivityConsumption{}).Where("activity_uuid = ? AND referrer_uuid = ? AND delete_time = 0", activityUuid, referrerUuid).Select("COALESCE(SUM(consumption_amount), 0)").Scan(&amount).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return amount, nil
}

// GetByActivityAndReferrer 根据活动ID、推荐人ID获取消费记录
func (r *MarketingActivityConsumptionRepo) GetByActivityAndReferrer(activityUuid, referrerUuid uint64) (*model.MarketingActivityConsumption, error) {
	var consumption model.MarketingActivityConsumption
	err := r.db.Where("activity_uuid = ? AND referrer_uuid = ? AND delete_time = 0", activityUuid, referrerUuid, 0).First(&consumption).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &consumption, nil
}

// GetByActivityAndReferrerAndConsumer 根据活动ID、推荐人ID和消费者ID获取消费记录
func (r *MarketingActivityConsumptionRepo) GetByActivityAndReferrerAndConsumer(activityUuid, referrerUuid, consumerUuid uint64) (*model.MarketingActivityConsumption, error) {
	var consumption model.MarketingActivityConsumption
	err := r.db.Where("activity_uuid = ? AND referrer_uuid = ? AND consumer_uuid = ? AND delete_time = 0", activityUuid, referrerUuid, consumerUuid).First(&consumption).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &consumption, nil
}

// UpdateAmount 更新消费金额
func (r *MarketingActivityConsumptionRepo) UpdateAmount(uuid uint64, amount float64) error {
	return r.db.Model(&model.MarketingActivityConsumption{}).Where("uuid = ?", uuid).Update("consumption_amount", gorm.Expr("consumption_amount + ?", amount)).Error
}

// UpdateRewardStatus 更新奖励状态
func (r *MarketingActivityConsumptionRepo) UpdateRewardStatus(activityUuid, referrerUuid uint64) error {
	return r.db.Model(&model.MarketingActivityConsumption{}).
		Where("activity_uuid = ? AND referrer_uuid = ? AND delete_time = 0", activityUuid, referrerUuid).
		Update("reward_status", 1).Error
}

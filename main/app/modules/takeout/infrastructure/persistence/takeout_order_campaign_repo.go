package persistence

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/takeout/domain/model"

	"gorm.io/gorm"
)

// ITakeoutOrderCampaignRepo 外卖订单活动仓储接口
type ITakeoutOrderCampaignRepo interface {
	Create(campaign *model.TakeoutOrderCampaign) error
	BatchCreate(campaigns []*model.TakeoutOrderCampaign) error
	Update(campaign *model.TakeoutOrderCampaign) error
	GetByOrderUuid(orderUuid uint64, options ...DBOption) ([]*model.TakeoutOrderCampaign, error)
	Delete(uuid uint64) error
}

// TakeoutOrderCampaignRepoImpl 外卖订单活动仓储实现
type TakeoutOrderCampaignRepoImpl struct {
	db *gorm.DB
}

// NewTakeoutOrderCampaignRepo 创建 TakeoutOrderCampaignRepoImpl 实例
func NewTakeoutOrderCampaignRepo(db *gorm.DB) ITakeoutOrderCampaignRepo {
	return &TakeoutOrderCampaignRepoImpl{db: db}
}

// Create 创建活动信息
func (r *TakeoutOrderCampaignRepoImpl) Create(campaign *model.TakeoutOrderCampaign) error {
	if err := r.db.Create(campaign).Error; err != nil {
		return errors.WithMessage(err, "创建活动信息失败")
	}
	return nil
}

// BatchCreate 批量创建活动信息
func (r *TakeoutOrderCampaignRepoImpl) BatchCreate(campaigns []*model.TakeoutOrderCampaign) error {
	if len(campaigns) == 0 {
		return nil
	}
	if err := r.db.Create(&campaigns).Error; err != nil {
		return errors.WithMessage(err, "批量创建活动信息失败")
	}
	return nil
}

// Update 更新活动信息
func (r *TakeoutOrderCampaignRepoImpl) Update(campaign *model.TakeoutOrderCampaign) error {
	if err := r.db.Save(campaign).Error; err != nil {
		return errors.WithMessage(err, "更新活动信息失败")
	}
	return nil
}

// GetByOrderUuid 根据订单 UUID 获取活动列表
func (r *TakeoutOrderCampaignRepoImpl) GetByOrderUuid(orderUuid uint64, options ...DBOption) ([]*model.TakeoutOrderCampaign, error) {
	var campaigns []*model.TakeoutOrderCampaign
	db := r.db.Model(&model.TakeoutOrderCampaign{}).Where("delete_time = ?", constant.NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	err := db.Where("takeout_order_uuid = ?", orderUuid).Find(&campaigns).Error
	if err != nil {
		return nil, errors.WithMessage(err, "查询活动信息失败")
	}
	return campaigns, nil
}

// Delete 删除活动信息 (软删除)
func (r *TakeoutOrderCampaignRepoImpl) Delete(uuid uint64) error {
	if err := r.db.Model(&model.TakeoutOrderCampaign{}).Where("uuid = ?", uuid).Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error; err != nil {
		return errors.WithMessage(err, "删除活动信息失败")
	}
	return nil
}

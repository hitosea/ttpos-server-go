package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IFullReductionActivityRuleRepo 满减活动规则仓库接口
type IFullReductionActivityRuleRepo interface {
	Create(rule *model.FullReductionActivityRule) error
	Update(rule *model.FullReductionActivityRule, options ...DBOption) error
	GetByFullReductionActivityUuid(activityUuid uint64, options ...DBOption) ([]*model.FullReductionActivityRule, error)
	DeleteByFullReductionActivityUuid(activityUuid uint64) error
	Delete(uuid uint64) error
}

// NewFullReductionActivityRuleRepo 创建满减活动规则仓库
func NewFullReductionActivityRuleRepo(db *gorm.DB) IFullReductionActivityRuleRepo {
	return &FullReductionActivityRuleRepoImpl{db: db}
}

// FullReductionActivityRuleRepoImpl 满减活动规则仓库实现
type FullReductionActivityRuleRepoImpl struct {
	db *gorm.DB
}

// Create 创建满减活动规则
func (r *FullReductionActivityRuleRepoImpl) Create(rule *model.FullReductionActivityRule) error {
	return errors.WithMessage(r.db.Create(rule).Error)
}

// Update 更新满减活动规则
func (r *FullReductionActivityRuleRepoImpl) Update(rule *model.FullReductionActivityRule, options ...DBOption) error {
	db := r.db.Model(&model.FullReductionActivityRule{})
	for _, option := range options {
		db = option(db)
	}
	return errors.WithMessage(db.Where("uuid = ?", rule.Uuid).Updates(rule).Error)
}

// GetByFullReductionActivityUuid 根据活动UUID获取规则列表
func (r *FullReductionActivityRuleRepoImpl) GetByFullReductionActivityUuid(activityUuid uint64, options ...DBOption) ([]*model.FullReductionActivityRule, error) {
	var rules []*model.FullReductionActivityRule
	db := r.db.Model(&model.FullReductionActivityRule{}).
		Where("delete_time = ?", constant.NotDeleted).
		Where("full_reduction_activity_uuid = ?", activityUuid)

	for _, option := range options {
		db = option(db)
	}

	err := db.Order("threshold ASC").Find(&rules).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return rules, nil
}

// DeleteByFullReductionActivityUuid 根据活动UUID软删除所有规则
func (r *FullReductionActivityRuleRepoImpl) DeleteByFullReductionActivityUuid(activityUuid uint64) error {
	return errors.WithMessage(
		r.db.Model(&model.FullReductionActivityRule{}).
			Where("full_reduction_activity_uuid = ? AND delete_time = ?", activityUuid, constant.NotDeleted).
			Update("delete_time", time.Now().Unix()).Error,
	)
}

// Delete 软删除满减活动规则
func (r *FullReductionActivityRuleRepoImpl) Delete(uuid uint64) error {
	return errors.WithMessage(
		r.db.Model(&model.FullReductionActivityRule{}).
			Where("uuid = ? AND delete_time = ?", uuid, constant.NotDeleted).
			Update("delete_time", time.Now().Unix()).Error,
	)
}

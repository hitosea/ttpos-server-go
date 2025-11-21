package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IFullReductionActivityRepo 满减活动仓库接口
type IFullReductionActivityRepo interface {
	Create(activity *model.FullReductionActivity) error
	Update(activity *model.FullReductionActivity, options ...DBOption) error
	GetByUuid(uuid uint64, options ...DBOption) (*model.FullReductionActivity, error)
	GetList(options ...DBOption) ([]*model.FullReductionActivity, int64, error)
	Delete(uuid uint64) error

	// 选项方法
	WhereUuid(uuid uint64) DBOption
	WhereStatus(status string, now int64) DBOption
}

// NewFullReductionActivityRepo 创建满减活动仓库
func NewFullReductionActivityRepo(db *gorm.DB) IFullReductionActivityRepo {
	return &FullReductionActivityRepoImpl{db: db}
}

// FullReductionActivityRepoImpl 满减活动仓库实现
type FullReductionActivityRepoImpl struct {
	db *gorm.DB
}

// Create 创建满减活动
func (r *FullReductionActivityRepoImpl) Create(activity *model.FullReductionActivity) error {
	return errors.WithMessage(r.db.Create(activity).Error)
}

// Update 更新满减活动
func (r *FullReductionActivityRepoImpl) Update(activity *model.FullReductionActivity, options ...DBOption) error {
	db := r.db.Model(&model.FullReductionActivity{})
	for _, option := range options {
		db = option(db)
	}
	return errors.WithMessage(db.Where("uuid = ?", activity.Uuid).Updates(activity).Error)
}

// GetByUuid 根据UUID获取满减活动
func (r *FullReductionActivityRepoImpl) GetByUuid(uuid uint64, options ...DBOption) (*model.FullReductionActivity, error) {
	var activity model.FullReductionActivity
	db := r.db.Model(&model.FullReductionActivity{}).Where("delete_time = ?", constant.NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	err := db.Where("uuid = ?", uuid).
		Preload("Rules", "delete_time = ?", constant.NotDeleted).
		Preload("MultiLanguageName", "delete_time = ?", constant.NotDeleted).
		First(&activity).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.WithMessage(err)
	}

	return &activity, nil
}

// GetList 获取满减活动列表
func (r *FullReductionActivityRepoImpl) GetList(options ...DBOption) ([]*model.FullReductionActivity, int64, error) {
	var activities []*model.FullReductionActivity
	var total int64

	db := r.db.Model(&model.FullReductionActivity{}).Where("delete_time = ?", constant.NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	// 获取列表
	err := db.
		Preload("Rules", "delete_time = ?", constant.NotDeleted).
		Preload("MultiLanguageName", "delete_time = ?", constant.NotDeleted).
		Order("create_time DESC").
		Find(&activities).Error

	if err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	return activities, total, nil
}

// Delete 软删除满减活动
func (r *FullReductionActivityRepoImpl) Delete(uuid uint64) error {
	return errors.WithMessage(
		r.db.Model(&model.FullReductionActivity{}).
			Where("uuid = ? AND delete_time = ?", uuid, constant.NotDeleted).
			Update("delete_time", time.Now().Unix()).Error,
	)
}

// WhereUuid 根据UUID查询选项
func (r *FullReductionActivityRepoImpl) WhereUuid(uuid uint64) DBOption {
	return CommonRepo.WhereByUuid(uuid)
}

// WhereStatus 根据状态查询选项
func (r *FullReductionActivityRepoImpl) WhereStatus(status string, now int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		switch status {
		case "ongoing":
			// 进行中：当前时间在活动日期范围内，且未失效
			return db.Where("start_date <= ? AND end_date >= ? AND is_disabled = ?", now, now, 0)
		case "not_started":
			// 未开始：当前时间小于开始日期，且未失效
			return db.Where("start_date > ? AND is_disabled = ?", now, 0)
		case "ended":
			// 已结束：当前时间大于结束日期，或已失效
			return db.Where("end_date < ? OR is_disabled = ?", now, 1)
		default:
			// 全部：不添加额外条件
			return db
		}
	}
}


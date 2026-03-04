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
	UpdateByMap(uuid uint64, data map[string]interface{}) error
	UpdateActivity(activity *model.FullReductionActivity) error
	UpdateDisabled(activity *model.FullReductionActivity) error
	GetByUuid(uuid uint64, options ...DBOption) (*model.FullReductionActivity, error)
	GetList(options ...DBOption) ([]*model.FullReductionActivity, int64, error)
	Delete(uuid uint64) error

	// 同步相关方法
	SoftDeleteByHeadquarterExcludeUuids(headquarterUuid uint64, excludeUuids []uint64) error
	FindHeadquarterByUuidsWithRules(uuids []uint64) ([]model.FullReductionActivity, error)
	UnscopedDeleteByUuid(uuid uint64) error
	UnscopedDeleteByHeadquarter(headquarterUuid uint64) error
	FindByHeadquarter(headquarterUuid uint64) ([]model.FullReductionActivity, error)
	FindAllHeadquarterWithRules() ([]model.FullReductionActivity, error)

	// 选项方法
	WhereUuid(uuid uint64) DBOption
	WhereStatus(status string, now int64) DBOption
	Limit(limit int) DBOption
	Offset(offset int) DBOption
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

// Update 更新满减活动（不推荐，建议使用 UpdateByMap）
func (r *FullReductionActivityRepoImpl) Update(activity *model.FullReductionActivity, options ...DBOption) error {
	db := r.db.Model(&model.FullReductionActivity{})
	for _, option := range options {
		db = option(db)
	}
	// 使用 Updates 更新非零值字段
	// 注意：零值字段（如 IsAllDay=0）不会被更新，建议使用 UpdateByMap 方法
	return errors.WithMessage(db.Where("uuid = ?", activity.Uuid).Updates(activity).Error)
}

// UpdateByMap 使用 map 更新满减活动
// 使用 map 可以更新零值字段，如 IsAllDay 从 1 变成 0
func (r *FullReductionActivityRepoImpl) UpdateByMap(uuid uint64, data map[string]interface{}) error {
	return errors.WithMessage(
		r.db.Model(&model.FullReductionActivity{}).
			Where("uuid = ?", uuid).
			Updates(data).Error,
	)
}

// UpdateActivity 更新满减活动基本信息（推荐）
func (r *FullReductionActivityRepoImpl) UpdateActivity(activity *model.FullReductionActivity) error {
	return errors.WithMessage(
		r.db.Model(&model.FullReductionActivity{}).
			Where("uuid = ?", activity.Uuid).
			Updates(map[string]interface{}{
				"name":           activity.Name,
				"start_date":     activity.StartDate,
				"end_date":       activity.EndDate,
				"start_time":     activity.StartTime,
				"end_time":       activity.EndTime,
				"is_all_day":     activity.IsAllDay,
				"reduction_type": activity.ReductionType,
				"update_time":    activity.UpdateTime,
			}).Error,
	)
}

// UpdateDisabled 更新满减活动失效状态
func (r *FullReductionActivityRepoImpl) UpdateDisabled(activity *model.FullReductionActivity) error {
	return errors.WithMessage(
		r.db.Model(&model.FullReductionActivity{}).
			Where("uuid = ?", activity.Uuid).
			Updates(map[string]interface{}{
				"is_disabled": activity.IsDisabled,
				"update_time": activity.UpdateTime,
			}).Error,
	)
}

// GetByUuid 根据UUID获取满减活动
func (r *FullReductionActivityRepoImpl) GetByUuid(uuid uint64, options ...DBOption) (*model.FullReductionActivity, error) {
	var activity model.FullReductionActivity
	db := r.db.Model(&model.FullReductionActivity{}).Where("delete_time = ?", constant.NotDeleted)

	for _, option := range options {
		db = option(db)
	}

	err := db.Where("uuid = ?", uuid).
		Preload("Rules", func(db *gorm.DB) *gorm.DB {
			return db.Where("delete_time = ?", constant.NotDeleted).Order("reduction_amount ASC")
		}).
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

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.WithMessage(err)
	}

	for _, option := range options {
		db = option(db)
	}

	// 获取列表（应用分页）
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

// SoftDeleteByHeadquarterExcludeUuids 根据总部UUID软删除排除指定UUID的活动
func (r *FullReductionActivityRepoImpl) SoftDeleteByHeadquarterExcludeUuids(headquarterUuid uint64, excludeUuids []uint64) error {
	return r.db.Model(&model.FullReductionActivity{}).
		Where("headquarter_uuid = ?", headquarterUuid).
		Where("uuid NOT IN (?)", excludeUuids).
		Update("delete_time", time.Now().Unix()).Error
}

// FindHeadquarterByUuidsWithRules 根据UUID列表查询总部满减活动（含规则）
func (r *FullReductionActivityRepoImpl) FindHeadquarterByUuidsWithRules(uuids []uint64) ([]model.FullReductionActivity, error) {
	var activities []model.FullReductionActivity
	err := r.db.Where("headquarter_uuid = 0 AND delete_time = 0 AND uuid IN (?)", uuids).
		Preload("Rules").
		Find(&activities).Error
	if err != nil {
		return nil, err
	}
	return activities, nil
}

// UnscopedDeleteByUuid 根据UUID硬删除满减活动
func (r *FullReductionActivityRepoImpl) UnscopedDeleteByUuid(uuid uint64) error {
	return r.db.Unscoped().Where("uuid = ?", uuid).Delete(&model.FullReductionActivity{}).Error
}

// UnscopedDeleteByHeadquarter 根据总部UUID硬删除所有满减活动
func (r *FullReductionActivityRepoImpl) UnscopedDeleteByHeadquarter(headquarterUuid uint64) error {
	return r.db.Unscoped().Where("headquarter_uuid = ?", headquarterUuid).Delete(&model.FullReductionActivity{}).Error
}

// FindByHeadquarter 根据总部UUID查询满减活动
func (r *FullReductionActivityRepoImpl) FindByHeadquarter(headquarterUuid uint64) ([]model.FullReductionActivity, error) {
	var activities []model.FullReductionActivity
	err := r.db.Where("headquarter_uuid = ?", headquarterUuid).Find(&activities).Error
	if err != nil {
		return nil, err
	}
	return activities, nil
}

// FindAllHeadquarterWithRules 查询所有总部满减活动（含规则）
func (r *FullReductionActivityRepoImpl) FindAllHeadquarterWithRules() ([]model.FullReductionActivity, error) {
	var activities []model.FullReductionActivity
	err := r.db.Where("headquarter_uuid = 0 AND delete_time = 0").
		Preload("Rules").
		Find(&activities).Error
	if err != nil {
		return nil, err
	}
	return activities, nil
}

// WhereUuid 根据UUID查询选项
func (r *FullReductionActivityRepoImpl) WhereUuid(uuid uint64) DBOption {
	return CommonRepo.WhereByUuid(uuid)
}

// WhereStatus 根据状态查询选项
func (r *FullReductionActivityRepoImpl) WhereStatus(status string, now int64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		switch status {
		case constant.ActivityStatusInProgress:
			// 进行中：当前时间在活动日期范围内，且未失效
			return db.Where("start_date <= ? AND end_date >= ? AND is_disabled = ?", now, now, constant.No)
		case constant.ActivityStatusNotStarted:
			// 未开始：当前时间小于开始日期，且未失效
			return db.Where("start_date > ? AND is_disabled = ?", now, constant.No)
		case constant.ActivityStatusEnded:
			// 已结束：当前时间大于结束日期，或已失效
			return db.Where("end_date < ? OR is_disabled = ?", now, constant.Yes)
		default:
			// 全部：不添加额外条件
			return db
		}
	}
}

// Limit 限制查询数量
func (r *FullReductionActivityRepoImpl) Limit(limit int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	}
}

// Offset 偏移量
func (r *FullReductionActivityRepoImpl) Offset(offset int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset)
	}
}

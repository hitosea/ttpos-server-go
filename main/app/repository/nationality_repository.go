package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// INationalityRepo 国籍仓库接口
//
// 任务: story-order-source-nationality Phase 2.6
// 需求: R2.1-R2.7
//
// @version v2.10.0
type INationalityRepo interface {
	INationalityQueryRepo
	Create(nationality model.Nationality) (uint64, error)
	Update(nationality model.Nationality) error
	SoftDelete(uuid uint64) error
}

// INationalityQueryRepo 国籍查询接口
type INationalityQueryRepo interface {
	FindList() ([]model.Nationality, error)
	FindByUuid(uuid uint64) (*model.Nationality, error)
	FindByUuidWithDeleted(uuid uint64) (*model.Nationality, error)
	CountOrdersByNationalityUuid(uuid uint64) (int64, error)
}

// NewNationalityRepo 创建国籍仓库
func NewNationalityRepo(db *gorm.DB) INationalityRepo {
	return NewNationalityRepoImpl(db)
}

// NewNationalityRepoImpl 创建国籍仓库实现
func NewNationalityRepoImpl(db *gorm.DB) *NationalityRepoImpl {
	return &NationalityRepoImpl{db: db}
}

// NationalityRepoImpl 国籍仓库实现
type NationalityRepoImpl struct {
	db *gorm.DB
}

// FindList 获取国籍列表（JOIN多语言表）
func (r *NationalityRepoImpl) FindList() ([]model.Nationality, error) {
	var nationalities []model.Nationality
	err := r.db.Model(&model.Nationality{}).
		Preload("MultiLanguageName", "delete_time = ?", 0).
		Where("delete_time = ?", 0).
		Order("create_time DESC").
		Find(&nationalities).Error

	return nationalities, errors.WithMessage(err)
}

// FindByUuid 根据UUID查找国籍
func (r *NationalityRepoImpl) FindByUuid(uuid uint64) (*model.Nationality, error) {
	var nationality model.Nationality
	err := r.db.Model(&model.Nationality{}).
		Preload("MultiLanguageName", "delete_time = ?", 0).
		Where("uuid = ? AND delete_time = ?", uuid, 0).
		First(&nationality).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 未找到记录返回 nil，不报错
		}
		return nil, errors.WithMessage(err)
	}

	return &nationality, nil
}

// Create 创建国籍
func (r *NationalityRepoImpl) Create(nationality model.Nationality) (uint64, error) {
	if err := r.db.Model(&model.Nationality{}).Create(&nationality).Error; err != nil {
		return 0, errors.WithMessage(err)
	}
	return nationality.Uuid, nil
}

// Update 更新国籍
func (r *NationalityRepoImpl) Update(nationality model.Nationality) error {
	err := r.db.Model(&model.Nationality{}).
		Where("uuid = ? AND delete_time = ?", nationality.Uuid, 0).
		Updates(map[string]interface{}{
			"multi_language_name_uuid": nationality.MultiLanguageNameUuid,
			"sort":                     nationality.Sort,
			"status":                   nationality.Status,
			"update_time":              nationality.UpdateTime,
		}).Error

	return errors.WithMessage(err)
}

// SoftDelete 软删除国籍
func (r *NationalityRepoImpl) SoftDelete(uuid uint64) error {
	err := r.db.Model(&model.Nationality{}).
		Where("uuid = ? AND delete_time = ?", uuid, 0).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error

	return errors.WithMessage(err)
}

// FindByUuidWithDeleted 根据UUID查找国籍（包含已删除）
// 用于订单详情查询，保证历史订单仍可显示已删除的配置名称
func (r *NationalityRepoImpl) FindByUuidWithDeleted(uuid uint64) (*model.Nationality, error) {
	var nationality model.Nationality
	err := r.db.Model(&model.Nationality{}).
		Preload("MultiLanguageName").
		Where("uuid = ?", uuid).
		First(&nationality).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 未找到记录返回 nil，不报错
		}
		return nil, errors.WithMessage(err)
	}

	return &nationality, nil
}

// CountOrdersByNationalityUuid 统计使用该国籍的订单数量
func (r *NationalityRepoImpl) CountOrdersByNationalityUuid(uuid uint64) (int64, error) {
	var count int64
	err := r.db.Model(&model.SaleBill{}).
		Where("nationality_uuid = ? AND delete_time = ?", uuid, 0).
		Count(&count).Error

	return count, errors.WithMessage(err)
}

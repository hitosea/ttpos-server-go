package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMaterialCategoryVisibilityRepo 物品分类可见性配置仓库接口
type IMaterialCategoryVisibilityRepo interface {
	// 基础 CRUD
	GetList(opts ...DBOption) ([]model.MaterialCategoryVisibility, error)
	GetByUuid(uuid uint64) (*model.MaterialCategoryVisibility, error)
	Create(visibility *model.MaterialCategoryVisibility) error
	Update(visibility *model.MaterialCategoryVisibility) error
	Delete(uuid uint64) error

	// 查询方法
	GetByHeadquarterUuid(headquarterUuid uint64) ([]model.MaterialCategoryVisibility, error)

	// DBOption 方法
	WhereHeadquarterUuid(headquarterUuid uint64) DBOption
	WhereNotDeleted() DBOption
}

// NewMaterialCategoryVisibilityRepo 创建新的物品分类可见性配置仓库
func NewMaterialCategoryVisibilityRepo(db *gorm.DB) IMaterialCategoryVisibilityRepo {
	return &MaterialCategoryVisibilityRepoImpl{db: db}
}

// MaterialCategoryVisibilityRepoImpl 物品分类可见性配置仓库实现
type MaterialCategoryVisibilityRepoImpl struct {
	db *gorm.DB
}

// GetList 获取配置列表
func (r *MaterialCategoryVisibilityRepoImpl) GetList(opts ...DBOption) ([]model.MaterialCategoryVisibility, error) {
	var list []model.MaterialCategoryVisibility
	query := r.db.Model(&model.MaterialCategoryVisibility{}).Where("delete_time = 0")
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.Order("create_time DESC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// GetByUuid 根据UUID获取配置
func (r *MaterialCategoryVisibilityRepoImpl) GetByUuid(uuid uint64) (*model.MaterialCategoryVisibility, error) {
	var visibility model.MaterialCategoryVisibility
	err := r.db.Model(&model.MaterialCategoryVisibility{}).
		Where("uuid = ? AND delete_time = 0", uuid).
		First(&visibility).Error
	if err != nil {
		return nil, err
	}
	return &visibility, nil
}

// Create 创建配置
func (r *MaterialCategoryVisibilityRepoImpl) Create(visibility *model.MaterialCategoryVisibility) error {
	return r.db.Create(visibility).Error
}

// Update 更新配置
func (r *MaterialCategoryVisibilityRepoImpl) Update(visibility *model.MaterialCategoryVisibility) error {
	return r.db.Model(&model.MaterialCategoryVisibility{}).
		Where("uuid = ?", visibility.Uuid).
		Updates(map[string]any{
			"name":           visibility.Name,
			"is_all_roles":   visibility.IsAllRoles,
			"category_uuids": visibility.CategoryUuids,
			"role_configs":   visibility.RoleConfigs,
		}).Error
}

// Delete 删除配置（软删除）
func (r *MaterialCategoryVisibilityRepoImpl) Delete(uuid uint64) error {
	return r.db.Model(&model.MaterialCategoryVisibility{}).
		Where("uuid = ?", uuid).
		Update("delete_time", gorm.Expr("UNIX_TIMESTAMP()")).Error
}

// GetByHeadquarterUuid 根据总部UUID获取配置列表
func (r *MaterialCategoryVisibilityRepoImpl) GetByHeadquarterUuid(headquarterUuid uint64) ([]model.MaterialCategoryVisibility, error) {
	var list []model.MaterialCategoryVisibility
	err := r.db.Model(&model.MaterialCategoryVisibility{}).
		Where("headquarter_uuid = ? AND delete_time = 0", headquarterUuid).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// WhereHeadquarterUuid 过滤总部UUID
func (r *MaterialCategoryVisibilityRepoImpl) WhereHeadquarterUuid(headquarterUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("headquarter_uuid = ?", headquarterUuid)
	}
}

// WhereNotDeleted 过滤未删除
func (r *MaterialCategoryVisibilityRepoImpl) WhereNotDeleted() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("delete_time = 0")
	}
}

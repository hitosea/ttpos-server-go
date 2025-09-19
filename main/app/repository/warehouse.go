package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IWarehouseRepo 仓库Repository接口
type IWarehouseRepo interface {
	// 基础操作
	Create(warehouse *model.Warehouse) error
	Update(warehouse *model.Warehouse) error
	Delete(uuid uint64) error
	GetByUuid(uuid uint64, opts ...DBOption) (*model.Warehouse, error)

	// 查询操作
	GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.Warehouse, int64, error)
	IsCodeExists(code string, excludeUuid uint64) (bool, error)

	// 条件查询选项
	WhereNameOrCodeLike(name string) DBOption
	WhereType(warehouseType string) DBOption
	WhereStatus(status int) DBOption
	OrderByCreateTime(desc bool) DBOption
	UpdateIsDefault(uuid uint64) error
}

// WarehouseRepoImpl 仓库Repository实现
type WarehouseRepoImpl struct {
	db *gorm.DB
}

// NewWarehouseRepo 创建仓库Repository
func NewWarehouseRepo(db *gorm.DB) IWarehouseRepo {
	return &WarehouseRepoImpl{db: db}
}

// Create 创建仓库
func (r *WarehouseRepoImpl) Create(warehouse *model.Warehouse) error {
	return r.db.Create(warehouse).Error
}

// Update 更新仓库
func (r *WarehouseRepoImpl) Update(warehouse *model.Warehouse) error {
	return r.db.Save(warehouse).Error
}

// Delete 删除仓库（软删除）
func (r *WarehouseRepoImpl) Delete(uuid uint64) error {
	return r.db.Model(&model.Warehouse{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error
}

// GetByUuid 根据UUID获取仓库
func (r *WarehouseRepoImpl) GetByUuid(uuid uint64, opts ...DBOption) (*model.Warehouse, error) {
	var warehouse model.Warehouse
	query := r.db.Where("uuid = ?", uuid).Preload("MultiLanguageName").Scopes(NotDeleted)
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.First(&warehouse).Error
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// GetListWithPagination 分页获取仓库列表
func (r *WarehouseRepoImpl) GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.Warehouse, int64, error) {
	var warehouses []model.Warehouse
	var total int64
	query := r.db.Model(&model.Warehouse{}).Preload("MultiLanguageName").Scopes(NotDeleted).Debug()
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	// 获取总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	// 分页查询
	offset := (pageNo - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Find(&warehouses).Error
	return warehouses, total, err
}

// Count 获取仓库总数
func (r *WarehouseRepoImpl) Count(opts ...DBOption) (int64, error) {
	var count int64
	query := r.db.Model(&model.Warehouse{})
	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.Count(&count).Error
	return count, err
}

// IsCodeExists 检查编码是否存在
func (r *WarehouseRepoImpl) IsCodeExists(code string, excludeUuid uint64) (bool, error) {
	var count int64
	query := r.db.Model(&model.Warehouse{}).Scopes(NotDeleted).Where("code = ?", code)
	if excludeUuid > 0 {
		query = query.Where("uuid != ?", excludeUuid)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// IsNameExists 检查名称是否存在
func (r *WarehouseRepoImpl) IsNameExists(name string, excludeUuid uint64) (bool, error) {
	var count int64
	query := r.db.Model(&model.Warehouse{}).Where("name = ? AND delete_time = 0", name)
	if excludeUuid > 0 {
		query = query.Where("uuid != ?", excludeUuid)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// 以下是查询选项方法

// WhereNameLike 名称模糊查询条件
func (r *WarehouseRepoImpl) WhereNameOrCodeLike(name string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name LIKE ? OR code LIKE ?", "%"+name+"%", "%"+name+"%")
	}
}

// WhereType 类型条件
func (r *WarehouseRepoImpl) WhereType(warehouseType string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if warehouseType != "" {
			return db.Where("type = ?", warehouseType)
		}
		return db
	}
}

// WhereStatus 状态条件
func (r *WarehouseRepoImpl) WhereStatus(status int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// OrderByCreateTime 按创建时间排序
func (r *WarehouseRepoImpl) OrderByCreateTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("create_time DESC")
		}
		return db.Order("create_time ASC")
	}
}

// UpdateIsDefault 更新是否默认
func (r *WarehouseRepoImpl) UpdateIsDefault(uuid uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.Warehouse{}).Where("id > 0").Update("is_default", 0).Error
		if err != nil {
			return err
		}
		err = tx.Model(&model.Warehouse{}).Where("uuid = ?", uuid).Update("is_default", 1).Error
		if err != nil {
			return err
		}
		return nil
	})
}

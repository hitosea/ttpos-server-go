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
	GetById(id uint64, opts ...DBOption) (*model.Warehouse, error)
	GetByUuid(uuid uint64, opts ...DBOption) (*model.Warehouse, error)

	// 查询操作
	GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.Warehouse, int64, error)
	IsCodeExists(code string, excludeUuid uint64) (bool, error)
	GetDefaultWarehouse() (*model.Warehouse, error)
	GetTransitWarehouse() (*model.Warehouse, error)
	GetNormalWarehouse() ([]model.Warehouse, error)
	GetByCode(code string, opts ...DBOption) (*model.Warehouse, error)
	GetByErpCode(erpCode string, opts ...DBOption) (*model.Warehouse, error)
	Get(opts ...DBOption) ([]model.Warehouse, error)

	// 条件查询选项
	WhereNameOrCodeLike(name string) DBOption
	WhereType(warehouseType string) DBOption
	WhereHeadquarterUuid(headquarterUuid uint64) DBOption
	WhereIsHeadquarter(isHeadquarter bool) DBOption
	WhereStatus(status int) DBOption
	OrderByUpdateTime(desc bool) DBOption
	UpdateIsDefault(uuid uint64) error

	WithItems() DBOption
	WithMultiLanguageName() DBOption
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

// GetById 根据ID获取仓库
func (r *WarehouseRepoImpl) GetById(id uint64, opts ...DBOption) (*model.Warehouse, error) {
	var warehouse model.Warehouse
	query := r.db.Where("id = ?", id).Preload("MultiLanguageName").Preload("Items").Scopes(NotDeleted)
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

// GetByUuid 根据UUID获取仓库
func (r *WarehouseRepoImpl) GetByUuid(uuid uint64, opts ...DBOption) (*model.Warehouse, error) {
	var warehouse model.Warehouse
	query := r.db.Where("uuid = ?", uuid).Scopes(NotDeleted)
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
	query := r.db.Model(&model.Warehouse{}).Preload("MultiLanguageName").Preload("Items").Scopes(NotDeleted)
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

// GetDefaultWarehouse 获取默认仓库
func (r *WarehouseRepoImpl) GetDefaultWarehouse() (*model.Warehouse, error) {
	var warehouse model.Warehouse
	err := r.db.Model(&model.Warehouse{}).Where("is_default = 1").First(&warehouse).Error
	return &warehouse, err
}

// GetTransitWarehouse 获取中转仓库
func (r *WarehouseRepoImpl) GetTransitWarehouse() (*model.Warehouse, error) {
	var warehouse model.Warehouse
	err := r.db.Model(&model.Warehouse{}).Where("type = 'transit'").First(&warehouse).Error
	return &warehouse, err
}

// GetNormalWarehouse 获取普通仓库
func (r *WarehouseRepoImpl) GetNormalWarehouse() ([]model.Warehouse, error) {
	var warehouses []model.Warehouse
	err := r.db.Model(&model.Warehouse{}).Where("type = 'normal'").
		Preload("MultiLanguageName").
		Scopes(NotDeleted).
		Find(&warehouses).Error
	return warehouses, err
}

// GetByCode 根据编码获取仓库
func (r *WarehouseRepoImpl) GetByCode(code string, opts ...DBOption) (*model.Warehouse, error) {
	var warehouse model.Warehouse
	query := r.db.Where("code = ?", code).Scopes(NotDeleted)
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

// GetByErpCode 根据ERP编码获取仓库
func (r *WarehouseRepoImpl) GetByErpCode(erpCode string, opts ...DBOption) (*model.Warehouse, error) {
	var warehouse model.Warehouse
	query := r.db.Where("erp_code = ?", erpCode).Scopes(NotDeleted)
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

// WhereIsHeadquarter 是否总部条件
func (r *WarehouseRepoImpl) WhereIsHeadquarter(isHeadquarter bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if isHeadquarter {
			return db.Where("headquarter_uuid != 0")
		}
		return db.Where("headquarter_uuid = 0")
	}
}

// OrderByUpdateTime 按更新时间排序
func (r *WarehouseRepoImpl) OrderByUpdateTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("update_time DESC")
		} else {
			return db
		}
	}
}

// UpdateIsDefault 更新是否默认
func (r *WarehouseRepoImpl) UpdateIsDefault(uuid uint64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		updateTime := time.Now().Unix()
		if err := tx.Model(&model.Warehouse{}).Where("is_default = ?", 1).Updates(map[string]any{
			"is_default":  0,
			"update_time": updateTime - 1,
		}).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Warehouse{}).Where("uuid = ?", uuid).Updates(map[string]any{
			"is_default":  1,
			"update_time": updateTime,
		}).Error; err != nil {
			return err
		}
		// 更新所有物品的仓库uuid
		if err := tx.Model(&model.Material{}).Where("id > 0").Update("warehouse_uuid", uuid).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *WarehouseRepoImpl) WhereHeadquarterUuid(headquarterUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("headquarter_uuid = ?", headquarterUuid)
	}
}

func (r *WarehouseRepoImpl) Get(opts ...DBOption) ([]model.Warehouse, error) {
	var warehouses []model.Warehouse
	query := r.db.Model(&model.Warehouse{}).Scopes(NotDeleted)
	for _, opt := range opts {
		query = opt(query)
	}
	err := query.Find(&warehouses).Error
	return warehouses, err
}

func (r *WarehouseRepoImpl) WithItems() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Items")
	}
}

func (r *WarehouseRepoImpl) WithMultiLanguageName() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MultiLanguageName")
	}
}

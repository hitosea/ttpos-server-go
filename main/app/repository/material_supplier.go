package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMaterialSupplierRepo 原料供应商关联Repository接口
type IMaterialSupplierRepo interface {
	// 基础操作
	Create(materialSupplier *model.MaterialSupplier) error
	Update(uuid uint64, data map[string]any) error
	Delete(opts ...DBOption) error
	DeletePermanently(opts ...DBOption) error // 彻底删除
	GetByUuid(uuid uint64, opts ...DBOption) (*model.MaterialSupplier, error)
	GetByMaterialAndSupplier(materialUuid, supplierUuid uint64, opts ...DBOption) (*model.MaterialSupplier, error)

	// 查询操作
	GetList(opts ...DBOption) ([]model.MaterialSupplier, error)
	GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.MaterialSupplier, int64, error)
	GetByMaterialUuid(materialUuid uint64, opts ...DBOption) ([]model.MaterialSupplier, error)
	GetBySupplierUuid(supplierUuid uint64, opts ...DBOption) ([]model.MaterialSupplier, error)
	GetByMaterialCode(materialCode string, opts ...DBOption) ([]model.MaterialSupplier, error)
	GetBySupplierErpCode(supplierErpCode string, opts ...DBOption) ([]model.MaterialSupplier, error)

	// 条件查询选项
	WhereMaterialUuid(materialUuid uint64) DBOption
	WhereSupplierUuid(supplierUuid uint64) DBOption
	WhereMaterialCode(materialCode string) DBOption
	WhereSupplierErpCode(supplierErpCode string) DBOption
	WhereHeadquarterUuid(headquarterUuid uint64) DBOption
	WhereIsHeadquarter(isHeadquarter bool) DBOption
	WhereNotDeleted() DBOption
	OrderByCreateTime(desc bool) DBOption
	WithMaterial() DBOption
	WithSupplier() DBOption
}

// MaterialSupplierRepoImpl 原料供应商关联Repository实现
type MaterialSupplierRepoImpl struct {
	db *gorm.DB
}

// NewMaterialSupplierRepo 创建原料供应商关联Repository
func NewMaterialSupplierRepo(db *gorm.DB) IMaterialSupplierRepo {
	return &MaterialSupplierRepoImpl{db: db}
}

// Create 创建原料供应商关联
func (r *MaterialSupplierRepoImpl) Create(materialSupplier *model.MaterialSupplier) error {
	return r.db.Create(materialSupplier).Error
}

// Update 更新原料供应商关联
func (r *MaterialSupplierRepoImpl) Update(uuid uint64, data map[string]any) error {
	return r.db.Model(&model.MaterialSupplier{}).Where("uuid = ? AND delete_time = 0", uuid).Updates(data).Error
}

// Delete 删除原料供应商关联（软删除）
func (r *MaterialSupplierRepoImpl) Delete(opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Model(&model.MaterialSupplier{}).Update("delete_time", time.Now().Unix()).Error
}

// DeletePermanently 彻底删除原料供应商关联
func (r *MaterialSupplierRepoImpl) DeletePermanently(opts ...DBOption) error {
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Model(&model.MaterialSupplier{}).Delete(&model.MaterialSupplier{}).Error
}

// GetByUuid 根据UUID获取原料供应商关联
func (r *MaterialSupplierRepoImpl) GetByUuid(uuid uint64, opts ...DBOption) (*model.MaterialSupplier, error) {
	var materialSupplier model.MaterialSupplier
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Where("uuid = ?", uuid).First(&materialSupplier).Error
	if err != nil {
		return nil, err
	}
	return &materialSupplier, nil
}

// GetByMaterialAndSupplier 根据原料UUID和供应商UUID获取关联
func (r *MaterialSupplierRepoImpl) GetByMaterialAndSupplier(materialUuid, supplierUuid uint64, opts ...DBOption) (*model.MaterialSupplier, error) {
	var materialSupplier model.MaterialSupplier
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Where("material_uuid = ? AND supplier_uuid = ?", materialUuid, supplierUuid).First(&materialSupplier).Error
	if err != nil {
		return nil, err
	}
	return &materialSupplier, nil
}

// GetList 获取原料供应商关联列表
func (r *MaterialSupplierRepoImpl) GetList(opts ...DBOption) ([]model.MaterialSupplier, error) {
	var list []model.MaterialSupplier
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&list).Error
	return list, err
}

// GetListWithPagination 分页获取原料供应商关联列表
func (r *MaterialSupplierRepoImpl) GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.MaterialSupplier, int64, error) {
	var list []model.MaterialSupplier
	var total int64

	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}

	// 统计总数
	err := db.Model(&model.MaterialSupplier{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageNo - 1) * pageSize
	err = db.Offset(offset).Limit(pageSize).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// GetByMaterialUuid 根据原料UUID获取关联列表
func (r *MaterialSupplierRepoImpl) GetByMaterialUuid(materialUuid uint64, opts ...DBOption) ([]model.MaterialSupplier, error) {
	var list []model.MaterialSupplier
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Where("material_uuid = ?", materialUuid).Find(&list).Error
	return list, err
}

// GetBySupplierUuid 根据供应商UUID获取关联列表
func (r *MaterialSupplierRepoImpl) GetBySupplierUuid(supplierUuid uint64, opts ...DBOption) ([]model.MaterialSupplier, error) {
	var list []model.MaterialSupplier
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Where("supplier_uuid = ?", supplierUuid).Find(&list).Error
	return list, err
}

// GetByMaterialCode 根据原料编码获取关联列表
func (r *MaterialSupplierRepoImpl) GetByMaterialCode(materialCode string, opts ...DBOption) ([]model.MaterialSupplier, error) {
	var list []model.MaterialSupplier
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Where("material_code = ?", materialCode).Find(&list).Error
	return list, err
}

// GetBySupplierErpCode 根据供应商ERP编码获取关联列表
func (r *MaterialSupplierRepoImpl) GetBySupplierErpCode(supplierErpCode string, opts ...DBOption) ([]model.MaterialSupplier, error) {
	var list []model.MaterialSupplier
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Where("supplier_erp_code = ?", supplierErpCode).Find(&list).Error
	return list, err
}

// WhereMaterialUuid 按原料UUID筛选
func (r *MaterialSupplierRepoImpl) WhereMaterialUuid(materialUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("material_uuid = ?", materialUuid)
	}
}

// WhereSupplierUuid 按供应商UUID筛选
func (r *MaterialSupplierRepoImpl) WhereSupplierUuid(supplierUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("supplier_uuid = ?", supplierUuid)
	}
}

// WhereMaterialCode 按原料编码筛选
func (r *MaterialSupplierRepoImpl) WhereMaterialCode(materialCode string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("material_code = ?", materialCode)
	}
}

// WhereSupplierErpCode 按供应商ERP编码筛选
func (r *MaterialSupplierRepoImpl) WhereSupplierErpCode(supplierErpCode string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("supplier_erp_code = ?", supplierErpCode)
	}
}

// WhereHeadquarterUuid 按总部UUID筛选
func (r *MaterialSupplierRepoImpl) WhereHeadquarterUuid(headquarterUuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("headquarter_uuid = ?", headquarterUuid)
	}
}

// WhereIsHeadquarter 按是否总部筛选
func (r *MaterialSupplierRepoImpl) WhereIsHeadquarter(isHeadquarter bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if isHeadquarter {
			return db.Where("headquarter_uuid != 0")
		}
		return db.Where("headquarter_uuid = 0")
	}
}

// WhereNotDeleted 筛选未删除的记录
func (r *MaterialSupplierRepoImpl) WhereNotDeleted() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("delete_time = 0")
	}
}

// OrderByCreateTime 按创建时间排序
func (r *MaterialSupplierRepoImpl) OrderByCreateTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("create_time DESC")
		}
		return db.Order("create_time ASC")
	}
}

// WithMaterial 预加载原料信息
func (r *MaterialSupplierRepoImpl) WithMaterial() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Material")
	}
}

// WithSupplier 预加载供应商信息
func (r *MaterialSupplierRepoImpl) WithSupplier() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("Supplier")
	}
}

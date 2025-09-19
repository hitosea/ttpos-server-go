package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ISupplierRepo 供应商Repository接口
type ISupplierRepo interface {
	// 基础操作
	Create(supplier *model.Supplier) error
	Update(supplier *model.Supplier) error
	Delete(uuid uint64) error
	GetByUuid(uuid uint64, opts ...DBOption) (*model.Supplier, error)

	// 查询操作
	GetList(opts ...DBOption) ([]model.Supplier, error)
	GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.Supplier, int64, error)
	IsNameExists(name string, excludeUuid uint64) (bool, error)
	IsCodeExists(code string, excludeUuid uint64) (bool, error)

	// 条件查询选项
	WhereNameOrCodeLike(name string) DBOption
	OrderByCreateTime(desc bool) DBOption
	OrderByName(desc bool) DBOption
}

// SupplierRepoImpl 供应商Repository实现
type SupplierRepoImpl struct {
	db *gorm.DB
}

// NewSupplierRepo 创建供应商Repository
func NewSupplierRepo(db *gorm.DB) ISupplierRepo {
	return &SupplierRepoImpl{db: db}
}

// Create 创建供应商
func (r *SupplierRepoImpl) Create(supplier *model.Supplier) error {
	return r.db.Create(supplier).Error
}

// Update 更新供应商
func (r *SupplierRepoImpl) Update(supplier *model.Supplier) error {
	return r.db.Model(supplier).Select("name", "code", "status", "address", "contact_name", "contact_phone").Where("uuid = ?", supplier.Uuid).Updates(supplier).Error
}

// Delete 软删除供应商
func (r *SupplierRepoImpl) Delete(uuid uint64) error {
	return r.db.Model(&model.Supplier{}).
		Where("uuid = ?", uuid).
		Update("delete_time", time.Now().Unix()).Error
}

// GetByUuid 根据UUID查询供应商
func (r *SupplierRepoImpl) GetByUuid(uuid uint64, opts ...DBOption) (*model.Supplier, error) {
	var supplier model.Supplier
	db := r.db.Model(&model.Supplier{}).Where("uuid = ?", uuid)
	db = r.applyOptions(db, opts...)
	err := db.First(&supplier).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

// GetList 获取供应商列表
func (r *SupplierRepoImpl) GetList(opts ...DBOption) ([]model.Supplier, error) {
	var suppliers []model.Supplier
	db := r.db.Model(&model.Supplier{})
	db = r.applyOptions(db, opts...)
	err := db.Find(&suppliers).Error
	return suppliers, err
}

// GetListWithPagination 分页获取供应商列表
func (r *SupplierRepoImpl) GetListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.Supplier, int64, error) {
	var suppliers []model.Supplier
	var total int64

	// 构建基础查询
	db := r.db.Model(&model.Supplier{}).Scopes(NotDeleted)
	db = r.applyOptions(db, opts...)

	// 获取总数
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (pageNo - 1) * pageSize
	err = db.Offset(offset).Limit(pageSize).Find(&suppliers).Error
	if err != nil {
		return nil, 0, err
	}

	return suppliers, total, nil
}

// IsNameExists 检查供应商名称是否存在
func (r *SupplierRepoImpl) IsNameExists(name string, excludeUuid uint64) (bool, error) {
	var count int64
	query := r.db.Model(&model.Supplier{}).
		Where("name = ? AND delete_time = ?", name, 0)

	if excludeUuid != 0 {
		query = query.Where("uuid != ?", excludeUuid)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IsCodeExists 检查供应商编码是否存在
func (r *SupplierRepoImpl) IsCodeExists(code string, excludeUuid uint64) (bool, error) {
	var count int64
	query := r.db.Model(&model.Supplier{}).
		Where("code = ? AND delete_time = ?", code, 0)

	if excludeUuid != 0 {
		query = query.Where("uuid != ?", excludeUuid)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// applyOptions 应用查询选项
func (r *SupplierRepoImpl) applyOptions(db *gorm.DB, opts ...DBOption) *gorm.DB {
	for _, opt := range opts {
		db = opt(db)
	}
	return db
}

// 查询选项实现

// WhereNameLike 名称模糊匹配条件
func (r *SupplierRepoImpl) WhereNameOrCodeLike(name string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name LIKE ? OR code LIKE ?", "%"+name+"%", "%"+name+"%")
	}
}

// WhereNotDeleted 未删除条件
func (r *SupplierRepoImpl) WhereNotDeleted() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("delete_time = ?", 0)
	}
}

// OrderByCreateTime 按创建时间排序
func (r *SupplierRepoImpl) OrderByCreateTime(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("create_time DESC")
		}
		return db.Order("create_time ASC")
	}
}

// OrderByName 按名称排序
func (r *SupplierRepoImpl) OrderByName(desc bool) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		if desc {
			return db.Order("name DESC")
		}
		return db.Order("name ASC")
	}
}

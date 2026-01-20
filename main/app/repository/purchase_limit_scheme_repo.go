package repository

import (
	"time"

	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPurchaseLimitSchemeRepo 限购方案数据访问接口
type IPurchaseLimitSchemeRepo interface {
	// Create 创建限购方案
	Create(scheme *model.PurchaseLimitScheme) error

	// Update 更新限购方案
	Update(scheme *model.PurchaseLimitScheme) error

	// GetByUuid 根据UUID查询限购方案
	GetByUuid(uuid uint64, options ...DBOption) (*model.PurchaseLimitScheme, error)

	// GetList 查询限购方案列表
	GetList(options ...DBOption) ([]*model.PurchaseLimitScheme, int64, error)

	// Delete 软删除限购方案
	Delete(uuid uint64) error

	// 选项方法
	WhereUuid(uuid uint64) DBOption
	WhereStatus(status int8) DBOption
	WhereName(name string) DBOption
	WhereNameLike(name string) DBOption
	Paginate(pageNo, pageSize int) DBOption
}

type purchaseLimitSchemeRepoImpl struct {
	db *gorm.DB // ✅ 只持有 db 实例
}

// NewPurchaseLimitSchemeRepo 创建限购方案仓储实例
func NewPurchaseLimitSchemeRepo(db *gorm.DB) IPurchaseLimitSchemeRepo {
	return &purchaseLimitSchemeRepoImpl{db: db}
}

// Create 创建限购方案
func (r *purchaseLimitSchemeRepoImpl) Create(scheme *model.PurchaseLimitScheme) error {
	return r.db.Create(scheme).Error
}

// Update 更新限购方案
func (r *purchaseLimitSchemeRepoImpl) Update(scheme *model.PurchaseLimitScheme) error {
	return r.db.Save(scheme).Error
}

// GetByUuid 根据UUID查询限购方案
func (r *purchaseLimitSchemeRepoImpl) GetByUuid(uuid uint64, options ...DBOption) (*model.PurchaseLimitScheme, error) {
	var scheme model.PurchaseLimitScheme
	db := r.db.Where("delete_time = ?", 0)

	for _, option := range options {
		db = option(db)
	}

	if err := db.Where("uuid = ?", uuid).First(&scheme).Error; err != nil {
		return nil, err
	}
	return &scheme, nil
}

// GetList 查询限购方案列表
func (r *purchaseLimitSchemeRepoImpl) GetList(options ...DBOption) ([]*model.PurchaseLimitScheme, int64, error) {
	var list []*model.PurchaseLimitScheme
	var total int64

	db := r.db.Where("delete_time = ?", 0)

	for _, option := range options {
		db = option(db)
	}

	if err := db.Model(&model.PurchaseLimitScheme{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Delete 软删除限购方案
func (r *purchaseLimitSchemeRepoImpl) Delete(uuid uint64) error {
	return r.db.Model(&model.PurchaseLimitScheme{}).
		Where("uuid = ?", uuid).
		Update("delete_time", time.Now().Unix()).Error
}

// 选项方法
func (r *purchaseLimitSchemeRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

func (r *purchaseLimitSchemeRepoImpl) WhereStatus(status int8) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

func (r *purchaseLimitSchemeRepoImpl) WhereName(name string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name = ?", name)
	}
}

func (r *purchaseLimitSchemeRepoImpl) WhereNameLike(name string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name LIKE ?", "%"+name+"%")
	}
}

func (r *purchaseLimitSchemeRepoImpl) Paginate(pageNo, pageSize int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		offset := (pageNo - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

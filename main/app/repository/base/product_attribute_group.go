package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductAttributeGroupRepo 产品属性组仓库接口
type IProductAttributeGroupRepo interface {
	GetProductAttributeGroupList() ([]model.ProductAttributeGroup, error)                          // 获取产品属性组列表
	GetProductAttributeGroupListWithAttributes() ([]model.ProductAttributeGroup, error)            // 获取属性组列表（含属性及多语言）
	UpdateProductAttributeGroup(id uint, productAttributeGroup model.ProductAttributeGroup) error  // 更新产品属性组
	CreateProductAttributeGroup(productAttributeGroup model.ProductAttributeGroup) (uint64, error) // 创建产品属性组
	CreateRecord(group *model.ProductAttributeGroup) error                                         // 仅创建属性组记录（不含多语言）
	CreateBatch(groups []model.ProductAttributeGroup) error
	GetMaxSort() (int, error) // 获取最大排序值
	GetMaxSortWithScopes(scopes ...func(*gorm.DB) *gorm.DB) (int, error)
	UpdateProductAttributeGroupData(data map[string]any, opts ...func(*gorm.DB) *gorm.DB) error
	HardDeleteByHeadquarter() error
	PluckHeadquarterUuids() ([]uint64, error)  // 获取所有总部属性组UUID（不过滤delete_time）
	DeleteProductAttributeGroup(id uint) error // 删除产品属性组
}

// NewProductAttributeGroupRepo 创建新的产品属性组仓库
func NewProductAttributeGroupRepo(db *gorm.DB) IProductAttributeGroupRepo {
	return NewProductAttributeGroupRepoImpl(db)
}

// NewProductAttributeGroupRepoImpl 创建新的退菜原因仓库实现
func NewProductAttributeGroupRepoImpl(db *gorm.DB) *ProductAttributeGroupRepoImpl {
	return &ProductAttributeGroupRepoImpl{db: db}
}

type ProductAttributeGroupRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// GetProductAttributeGroupList 获取退菜原因列表，排除逻辑删除的退菜原因
func (r *ProductAttributeGroupRepoImpl) GetProductAttributeGroupList() ([]model.ProductAttributeGroup, error) {
	var productAttributeGroups []model.ProductAttributeGroup
	err := r.db.Model(&model.ProductAttributeGroup{}).Preload("MultiLanguageName").Where("delete_time = ?", 0).Find(&productAttributeGroups).Error
	return productAttributeGroups, errors.WithMessage(err)
}

// GetProductAttributeGroupListWithAttributes 获取属性组列表（含属性及多语言）
func (r *ProductAttributeGroupRepoImpl) GetProductAttributeGroupListWithAttributes() ([]model.ProductAttributeGroup, error) {
	var groups []model.ProductAttributeGroup
	err := r.db.Model(&model.ProductAttributeGroup{}).
		Preload("MultiLanguageName").
		Preload("ProductAttributes").
		Preload("ProductAttributes.MultiLanguageName").
		Find(&groups).Error
	return groups, errors.WithMessage(err)
}

// UpdateProductAttributeGroup 更新退菜原因
func (r *ProductAttributeGroupRepoImpl) UpdateProductAttributeGroup(id uint, productAttributeGroup model.ProductAttributeGroup) error {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	if err := tx.Model(&model.ProductAttributeGroup{}).Where("id = ?", id).Updates(productAttributeGroup).Error; err != nil {
		tx.Rollback() // 更新失败，回滚事务
		return errors.WithMessage(err)
	}

	if err := tx.Model(&productAttributeGroup.MultiLanguageName).Where("id = ?", productAttributeGroup.MultiLanguageNameUuid).Updates(productAttributeGroup.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 更新多语言名称失败，回滚事务
		return errors.WithMessage(err)
	}

	return tx.Commit().Error // 提交事务
}

// CreateProductAttributeGroup 创建产品属性组
func (r *ProductAttributeGroupRepoImpl) CreateProductAttributeGroup(productAttributeGroup model.ProductAttributeGroup) (uint64, error) {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	// 创建多语言名称
	if err := tx.Create(&productAttributeGroup.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 创建多语言名称失败，回滚事务
		return 0, errors.WithMessage(err)
	}

	// 创建产品属性组
	if err := tx.Create(&productAttributeGroup).Error; err != nil {
		tx.Rollback() // 创建失败，回滚事务
		return 0, errors.WithMessage(err)
	}

	return productAttributeGroup.Uuid, tx.Commit().Error // 提交事务
}

// CreateRecord 仅创建属性组记录（不含多语言名称创建）
func (r *ProductAttributeGroupRepoImpl) CreateRecord(group *model.ProductAttributeGroup) error {
	return errors.WithMessage(r.db.Create(group).Error)
}

// GetMaxSort 获取最大排序值
func (r *ProductAttributeGroupRepoImpl) GetMaxSort() (int, error) {
	var maxSort int
	err := r.db.Model(&model.ProductAttributeGroup{}).Select("IFNULL(MAX(sort), 0)").Scan(&maxSort).Error
	return maxSort, errors.WithMessage(err)
}

func (r *ProductAttributeGroupRepoImpl) CreateBatch(groups []model.ProductAttributeGroup) error {
	if len(groups) == 0 {
		return nil
	}
	return r.db.Create(&groups).Error
}

func (r *ProductAttributeGroupRepoImpl) GetMaxSortWithScopes(scopes ...func(*gorm.DB) *gorm.DB) (int, error) {
	var maxSort int
	err := r.db.Model(&model.ProductAttributeGroup{}).Scopes(scopes...).Select("IFNULL(MAX(sort), 0)").Scan(&maxSort).Error
	return maxSort, err
}

func (r *ProductAttributeGroupRepoImpl) UpdateProductAttributeGroupData(data map[string]any, opts ...func(*gorm.DB) *gorm.DB) error {
	db := r.db.Model(&model.ProductAttributeGroup{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(data).Error
}

func (r *ProductAttributeGroupRepoImpl) HardDeleteByHeadquarter() error {
	return r.db.Where("headquarter_uuid > 0").Delete(&model.ProductAttributeGroup{}).Error
}

// PluckHeadquarterUuids 获取所有总部属性组UUID（不过滤delete_time，与同步清理逻辑一致）
func (r *ProductAttributeGroupRepoImpl) PluckHeadquarterUuids() ([]uint64, error) {
	var uuids []uint64
	err := r.db.Model(&model.ProductAttributeGroup{}).Where("headquarter_uuid > 0").Pluck("uuid", &uuids).Error
	return uuids, err
}

// DeleteProductAttributeGroup 软删除退菜原因
func (r *ProductAttributeGroupRepoImpl) DeleteProductAttributeGroup(id uint) error {
	return r.db.Model(&model.ProductAttributeGroup{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}

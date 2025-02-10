package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductUnitRepo 商品规格
type IProductUnitRepo interface {
	GetProductUnitList() ([]model.ProductUnit, error)
	UpdateProductUnit(id uint, productUnit model.ProductUnit) error
	CreateProductUnit(productUnit model.ProductUnit) (uint64, error)
	DeleteProductUnit(id uint) error
}

func NewProductUnitRepo(db *gorm.DB) IProductUnitRepo {
	return NewProductUnitRepoImpl(db)
}

// NewProductUnitRepoImpl 创建新的商品规格仓库实现
func NewProductUnitRepoImpl(db *gorm.DB) *ProductUnitRepoImpl {
	return &ProductUnitRepoImpl{db: db}
}

type ProductUnitRepoImpl struct {
	db *gorm.DB
}

// GetProductUnitList 获取商品规格列表，排除逻辑删除的规格
func (r *ProductUnitRepoImpl) GetProductUnitList() ([]model.ProductUnit, error) {
	var productUnits []model.ProductUnit
	err := r.db.Model(&model.ProductUnit{}).Preload("MultiLanguageName").Where("delete_time = ?", 0).Find(&productUnits).Error
	return productUnits, err
}

// UpdateProductUnit 更新商品规格
func (r *ProductUnitRepoImpl) UpdateProductUnit(id uint, productUnit model.ProductUnit) error {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	if err := tx.Model(&model.ProductUnit{}).Where("id = ?", id).Updates(productUnit).Error; err != nil {
		tx.Rollback() // 更新失败，回滚事务
		return err
	}

	if err := tx.Model(&productUnit.MultiLanguageName).Where("id = ?", productUnit.MultiLanguageNameUuid).Updates(productUnit.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 更新多语言名称失败，回滚事务
		return err
	}

	return tx.Commit().Error // 提交事务
}

// CreateProductUnit 创建商品规格
func (r *ProductUnitRepoImpl) CreateProductUnit(productUnit model.ProductUnit) (uint64, error) {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	// 创建多语言名称
	if err := tx.Create(&productUnit.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 创建多语言名称失败，回滚事务
		return 0, err
	}

	// 创建商品规格
	if err := tx.Create(&productUnit).Error; err != nil {
		tx.Rollback() // 创建失败，回滚事务
		return 0, err
	}

	return productUnit.Uuid, tx.Commit().Error // 提交事务
}

// DeleteProductUnit 软删除商品规格
func (r *ProductUnitRepoImpl) DeleteProductUnit(id uint) error {
	return r.db.Model(&model.ProductUnit{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}

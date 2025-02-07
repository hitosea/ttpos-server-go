package base

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMaterialRepo 原料仓库接口
type IMaterialRepo interface {
	GetMaterialList() ([]model.Material, error)
	UpdateMaterial(id uint, material model.Material) error
	CreateMaterial(material model.Material) (uint, error)
	DeleteMaterial(id uint) error
}

// NewMaterialRepo 创建新的原料仓库
func NewMaterialRepo(db *gorm.DB) IMaterialRepo {
	return NewMaterialRepoImpl(db)
}

// NewMaterialRepoImpl 创建新的原料仓库实现
func NewMaterialRepoImpl(db *gorm.DB) *MaterialRepoImpl {
	return &MaterialRepoImpl{db: db}
}

type MaterialRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// GetMaterialList 获取原料列表
func (r *MaterialRepoImpl) GetMaterialList() ([]model.Material, error) {
	var materials []model.Material
	err := r.db.Model(&model.Material{}).Preload("MultiLanguageName").Where("delete_time = ?", 0).Find(&materials).Error
	return materials, err
}

// UpdateMaterial 更新原料
func (r *MaterialRepoImpl) UpdateMaterial(id uint, material model.Material) error {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	if err := tx.Model(&model.Material{}).Where("id = ?", id).Updates(material).Error; err != nil {
		tx.Rollback() // 更新失败，回滚事务
		return err
	}

	if err := tx.Model(&material.MultiLanguageName).Where("id = ?", material.MultiLanguageNameUuid).Updates(material.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 更新多语言名称失败，回滚事务
		return err
	}

	return tx.Commit().Error // 提交事务
}

// CreateMaterial 创建原料
func (r *MaterialRepoImpl) CreateMaterial(material model.Material) (uint, error) {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	// 创建多语言名称
	if err := tx.Create(&material.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 创建多语言名称失败，回滚事务
		return 0, err
	}

	// 创建原料
	if err := tx.Create(&material).Error; err != nil {
		tx.Rollback() // 创建失败，回滚事务
		return 0, err
	}

	return material.Id, tx.Commit().Error // 提交事务
}

// DeleteMaterial 删除原料
func (r *MaterialRepoImpl) DeleteMaterial(id uint) error {
	return r.db.Model(&model.Material{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}

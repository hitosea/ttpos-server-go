package repository

import (
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// MaterialCategoryRepoImpl 实现 IMaterialCategoryRepo
type MaterialCategoryRepoImpl struct {
	db *gorm.DB
}

func NewMaterialCategoryRepoImpl(db *gorm.DB) *MaterialCategoryRepoImpl {
	return &MaterialCategoryRepoImpl{db: db}
}

// 获取原料类别列表
func (r *MaterialCategoryRepoImpl) GetMaterialCategoryList() ([]model.MaterialCategory, error) {
	var categories []model.MaterialCategory
	err := r.db.Model(&model.MaterialCategory{}).Find(&categories).Error
	return categories, err
}

// 更新原料类别
func (r *MaterialCategoryRepoImpl) UpdateMaterialCategory(id uint, materialCategory model.MaterialCategory) error {
	return r.db.Model(&model.MaterialCategory{}).Where("id = ?", id).Updates(materialCategory).Error
}

// 创建原料类别
func (r *MaterialCategoryRepoImpl) CreateMaterialCategory(materialCategory model.MaterialCategory) (uint, error) {
	err := r.db.Create(&materialCategory).Error
	return materialCategory.Id, err
}

// 软删除原料类别
func (r *MaterialCategoryRepoImpl) DeleteMaterialCategory(id uint) error {
	return r.db.Model(&model.MaterialCategory{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}

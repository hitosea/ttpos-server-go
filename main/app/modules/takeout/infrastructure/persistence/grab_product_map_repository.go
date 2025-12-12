package persistence

import (
	"ttpos-server-go/app/model"
	menuRepo "ttpos-server-go/app/modules/takeout/domain/menu/repository"

	"gorm.io/gorm"
)

// productMapRepository 商品映射仓储实现
type productMapRepository struct{}

// NewProductMapRepository 创建映射仓储
func NewProductMapRepository() menuRepo.IProductMapRepository {
	return &productMapRepository{}
}

// GetBySourceId 根据来源平台和商品ID获取映射
func (r *productMapRepository) GetBySourceId(db *gorm.DB, source, sourceProductId string) (model.ProductMap, error) {
	var m model.ProductMap
	err := db.Where("source = ? AND source_product_id = ? AND delete_time = 0", source, sourceProductId).First(&m).Error
	return m, err
}

// Create 创建映射
func (r *productMapRepository) Create(db *gorm.DB, m *model.ProductMap) error {
	return db.Create(m).Error
}

// UpdateById 更新映射
func (r *productMapRepository) UpdateById(db *gorm.DB, id uint, fields map[string]interface{}) error {
	return db.Model(&model.ProductMap{}).Where("id = ?", id).Updates(fields).Error
}

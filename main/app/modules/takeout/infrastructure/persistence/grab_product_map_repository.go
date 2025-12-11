package persistence

import (
	"ttpos-server-go/app/model"
	menuRepo "ttpos-server-go/app/modules/takeout/domain/menu/repository"

	"gorm.io/gorm"
)

// grabProductMapRepository Grab 商品映射仓储实现
type grabProductMapRepository struct{}

// NewGrabProductMapRepository 创建映射仓储
func NewGrabProductMapRepository() menuRepo.IGrabProductMapRepository {
	return &grabProductMapRepository{}
}

// GetByGrabId 获取映射
func (r *grabProductMapRepository) GetByGrabId(db *gorm.DB, grabProductId string) (model.GrabProductMap, error) {
	var m model.GrabProductMap
	err := db.Where("grab_product_id = ? AND delete_time = 0", grabProductId).First(&m).Error
	return m, err
}

// Create 创建映射
func (r *grabProductMapRepository) Create(db *gorm.DB, m *model.GrabProductMap) error {
	return db.Create(m).Error
}

// UpdateById 更新映射
func (r *grabProductMapRepository) UpdateById(db *gorm.DB, id uint, fields map[string]interface{}) error {
	return db.Model(&model.GrabProductMap{}).Where("id = ?", id).Updates(fields).Error
}

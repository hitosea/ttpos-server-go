package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductMapRepository 外卖平台商品与店内商品映射仓储接口
type IProductMapRepository interface {
	GetBySourceId(db *gorm.DB, source, sourceProductId string) (model.ProductMap, error)
	Create(db *gorm.DB, m *model.ProductMap) error
	UpdateById(db *gorm.DB, id uint, fields map[string]interface{}) error
}

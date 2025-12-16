package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IGrabProductMapRepository Grab 商品与店内商品映射仓储接口
type IGrabProductMapRepository interface {
	GetByGrabId(db *gorm.DB, grabProductId string) (model.GrabProductMap, error)
	Create(db *gorm.DB, m *model.GrabProductMap) error
	UpdateById(db *gorm.DB, id uint, fields map[string]interface{}) error
}

package model

import (
	"time"
)

// WarehouseItem 仓库商品库存表 `ttpos_warehouse_item`
type WarehouseItem struct {
	BaseModel
	Uuid          uint64  `json:"uuid" gorm:"type:bigint;default:0;comment:UUID;uniqueIndex:unique_uuid"`
	WarehouseUuid uint64  `json:"warehouse_uuid" gorm:"type:bigint;default:0;comment:仓库UUID;index:idx_warehouse_uuid"`
	MaterialUuid  uint64  `json:"material_uuid" gorm:"type:bigint;default:0;comment:商品UUID;index:idx_material_uuid"`
	MaterialCode  string  `json:"material_code" gorm:"type:varchar(255);default:'';comment:商品编码;index:idx_material_code"`
	Stock         float64 `json:"stock" gorm:"type:decimal(14,2);default:0.00;comment:库存数量"`
	ReservedStock float64 `json:"reserved_stock" gorm:"type:decimal(14,2);default:0.00;comment:预留库存数量"`
	CreateTime    int     `json:"create_time" gorm:"type:int;default:0;comment:创建时间"`
	UpdateTime    int     `json:"update_time" gorm:"type:int;default:0;comment:更新时间"`
	DeleteTime    int     `json:"delete_time" gorm:"type:int;default:0;comment:删除时间"`

	// 关联模型
	Warehouse *Warehouse `gorm:"foreignKey:WarehouseUuid;references:Uuid" json:"warehouse,omitempty"`
}

// TableName 返回表名
func (WarehouseItem) TableName() string {
	return "ttpos_warehouse_item"
}

// BeforeCreate 创建前钩子
func (w *WarehouseItem) BeforeCreate() error {
	now := int(time.Now().Unix())
	w.CreateTime = now
	w.UpdateTime = now
	return nil
}

// BeforeUpdate 更新前钩子
func (w *WarehouseItem) BeforeUpdate() error {
	w.UpdateTime = int(time.Now().Unix())
	return nil
}

package model

// WarehouseItem 仓库商品库存表
type WarehouseItem struct {
	BaseModel
	Uuid          uint64  `json:"uuid" gorm:"type:bigint;default:0;comment:UUID;uniqueIndex:unique_uuid"`
	WarehouseUuid uint64  `json:"warehouse_uuid" gorm:"type:bigint;default:0;comment:仓库UUID;index:idx_warehouse_uuid"`
	MaterialUuid  uint64  `json:"material_uuid" gorm:"type:bigint;default:0;comment:商品UUID;index:idx_material_uuid"`
	MaterialCode  string  `json:"material_code" gorm:"type:varchar(255);default:'';comment:商品编码;index:idx_material_code"`
	Stock         float64 `json:"stock" gorm:"type:decimal(14,8);default:0.00;comment:库存数量"`
	ReservedStock float64 `json:"reserved_stock" gorm:"type:decimal(14,8);default:0.00;comment:预留库存数量"`
	Valuation     float64 `json:"valuation" gorm:"type:decimal(14,8);default:0.00;comment:估值单价"`
}

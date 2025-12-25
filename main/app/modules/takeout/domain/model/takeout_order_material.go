package model

import "ttpos-server-go/app/model"

// TakeoutOrderMaterial 外卖订单原料 `ttpos_takeout_order_material`
type TakeoutOrderMaterial struct {
	BaseModel
	// 基础标识字段
	TakeoutOrderUuid  uint64  `gorm:"column:takeout_order_uuid;type:bigint(20);default:0;comment:外卖订单ID" json:"takeout_order_uuid"`
	MaterialUuid      uint64  `gorm:"column:material_uuid;type:bigint(20);default:0;comment:原料ID" json:"material_uuid"`
	WarehouseUuid     uint64  `gorm:"column:warehouse_uuid;type:bigint(20);default:0;comment:仓库ID" json:"warehouse_uuid"`
	Num               float64 `gorm:"column:num;type:decimal(12,2);default:0;comment:数量,原料的实际使用数量" json:"num"`
	StaffShiftLogUuid uint64  `gorm:"column:staff_shift_log_uuid;type:bigint(20);default:0;comment:员工班次记录ID" json:"staff_shift_log_uuid"`
	IsSummarized      int     `gorm:"column:is_summarized;type:int(11);default:0;comment:是否已经统计,0-未统计 1-已统计" json:"is_summarized"`

	Material *model.Material `gorm:"foreignKey:MaterialUuid;references:Uuid" json:"material,omitempty"`
}

func (m *TakeoutOrderMaterial) TableName() string {
	return "ttpos_takeout_order_material"
}

func (m *TakeoutOrderMaterial) SetNil() {
	m.Material = nil
}

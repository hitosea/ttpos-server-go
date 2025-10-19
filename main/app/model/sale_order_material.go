package model

// SaleOrderMaterial 销售订单原料 `ttpos_sale_order_material`
type SaleOrderMaterial struct {
	BaseModel
	// 基础标识字段
	SaleOrderUuid     uint64  `gorm:"column:sale_order_uuid;type:bigint(20);default:0;comment:销售订单ID" json:"sale_order_uuid"`
	SaleBillUuid      uint64  `gorm:"column:sale_bill_uuid;type:bigint(20);default:0;comment:销售账单ID" json:"sale_bill_uuid"`
	MaterialUuid      uint64  `gorm:"column:material_uuid;type:bigint(20);default:0;comment:原料ID" json:"material_uuid"`
	Num               float64 `gorm:"column:num;type:decimal(12,2);default:0;comment:数量,原料的实际使用数量" json:"num"`
	StaffShiftLogUuid uint64  `gorm:"column:staff_shift_log_uuid;type:bigint(20);default:0;comment:员工班次记录ID" json:"staff_shift_log_uuid"`

	Material *Material `gorm:"foreignKey:MaterialUuid;references:Uuid"`
}

func (model *SaleOrderMaterial) TableName() string {
	return "ttpos_sale_order_material"
}

func (model *SaleOrderMaterial) SetNil() {
	model.Material = nil
}

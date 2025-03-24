package model

// SaleBillOperationRecord 桌台账单操作记录 `ttpos_sale_bill_operation_record`
type SaleBillOperationRecord struct {
	BaseModel
	// 基本信息
	Data   string `gorm:"column:data;comment:操作来源 cashier-收银 assistant-助手 shop-商家后台" json:"data"`
	Source string `gorm:"column:source;comment:操作来源 cashier-收银 assistant-助手 shop-商家后台" json:"source"`
	Action string `gorm:"column:action;comment:操作行为" json:"action"`
	Remark string `gorm:"column:remark;comment:备注" json:"remark"`
	// 关联ID字段
	SaleBillUuid  uint64 `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;not null;default:0;comment:销售账单ID" json:"sale_bill_uuid"`
	SaleOrderUuid uint64 `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:销售订单ID" json:"sale_order_uuid"`
	OperatorUuid  uint64 `gorm:"column:operator_uuid;type:bigint(20) unsigned;not null;default:0;comment:操作员ID" json:"operator_uuid"`

	Operator Staff `gorm:"foreignKey:OperatorUuid;references:uuid"`
}

func (model *SaleBillOperationRecord) SetNil() {

}

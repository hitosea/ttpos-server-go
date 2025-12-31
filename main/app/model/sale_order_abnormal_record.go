package model

// SaleOrderAbnormalRecord 订单异常记录 `ttpos_sale_order_abnormal_record`
type SaleOrderAbnormalRecord struct {
	BaseModel
	// 基本信息
	SaleBillUuid  uint64 `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;not null;default:0;comment:销售账单ID" json:"sale_bill_uuid"`
	SaleOrderUuid uint64 `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:销售订单ID" json:"sale_order_uuid"`
	DutyNo        string `gorm:"column:duty_no;type:varchar(64);not null;default:'';comment:当班编号" json:"duty_no"`
	Action        string `gorm:"column:action;type:varchar(150);not null;default:'';comment:行为" json:"action"`
	SubAction     string `gorm:"column:sub_action;type:varchar(150);not null;default:'';comment:自定义子行为" json:"sub_action"`
	Sign          string `gorm:"column:sign;type:varchar(2000);not null;default:'';comment:操作签名" json:"sign"`
	Remark        string `gorm:"column:remark;type:varchar(255);not null;default:'';comment:备注" json:"remark"`
	CashierUuid   uint64 `gorm:"column:cashier_uuid;type:bigint(20) unsigned;not null;default:0;comment:收银员ID" json:"cashier_uuid"`
}

func (model *SaleOrderAbnormalRecord) SetNil() {

}

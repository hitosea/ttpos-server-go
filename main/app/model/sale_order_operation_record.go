package model

// SaleOrderOperationRecord 订单操作记录 `ttpos_sale_order_operation_record`
type SaleOrderOperationRecord struct {
	BaseModel

	Source        string `gorm:"column:source;type:varchar(255);comment:操作来源 cashier-收银端 assistant-点餐助手 shop-商家后台 h5-扫码点餐;NOT NULL" json:"source"`
	Action        string `gorm:"column:action;type:varchar(150);comment:操作行为;NOT NULL" json:"action"`
	Data          string `gorm:"column:data;type:text;comment:数据;NOT NULL" json:"data"`
	Remark        string `gorm:"column:remark;type:varchar(255);comment:备注;NOT NULL" json:"remark"`
	SaleBillUuid  uint64 `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;default:0;comment:销售账单ID;NOT NULL" json:"sale_bill_uuid"`
	SaleOrderUuid uint64 `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;default:0;comment:销售订单ID;NOT NULL" json:"sale_order_uuid"`
	H5OrderUuid   uint64 `gorm:"column:h5_order_uuid;type:bigint(20) unsigned;default:0;comment:h5订单Uuid;NOT NULL" json:"h5_order_uuid"`
	OperatorUuid  uint64 `gorm:"column:operator_uuid;type:bigint(20) unsigned;default:0;comment:操作员ID;NOT NULL" json:"operator_uuid"`

	dutyNo string `gorm:"-" json:"duty_no,omitempty"`

	Operator Staff `gorm:"foreignKey:OperatorUuid;references:uuid"`
}

func (model *SaleOrderOperationRecord) SetNil() {

}

func (model *SaleOrderOperationRecord) SetDutyNo(dutyNo string) {
	model.dutyNo = dutyNo
}

func (model *SaleOrderOperationRecord) GetDutyNo() string {
	return model.dutyNo
}

package model

// CostCardCorrectionLog 成本卡修正操作日志表 `ttpos_cost_card_correction_log`
type CostCardCorrectionLog struct {
	BaseModel
	CorrectionUuid uint64 `gorm:"column:correction_uuid;type:bigint(20) unsigned;not null;default:0;comment:修正操作UUID;index" json:"correction_uuid"`
	OrderUuid      uint64 `gorm:"column:order_uuid;type:bigint(20) unsigned;not null;default:0;comment:订单UUID;index" json:"order_uuid"`
	OrderNo        string `gorm:"column:order_no;type:varchar(100);not null;default:'';comment:订单号" json:"order_no"`
	OperatorUuid  uint64 `gorm:"column:operator_uuid;type:bigint(20) unsigned;not null;default:0;comment:操作人UUID;index" json:"operator_uuid"`
	OperatorName   string `gorm:"column:operator_name;type:varchar(100);not null;default:'';comment:操作人姓名" json:"operator_name"`
	OperationType  string `gorm:"column:operation_type;type:varchar(50);not null;default:'';comment:操作类型(preview/execute);index" json:"operation_type"`
	Status         string `gorm:"column:status;type:varchar(20);not null;default:'';comment:状态(success/failed);index" json:"status"`
	Message        string `gorm:"column:message;type:varchar(500);not null;default:'';comment:消息" json:"message"`
	Details        string `gorm:"column:details;type:text;comment:详细信息(JSON格式)" json:"details"`
}

// TableName 指定表名
func (CostCardCorrectionLog) TableName() string {
	return "ttpos_cost_card_correction_log"
}


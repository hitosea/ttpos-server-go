package model

// StockLossAnnotation 报损单批注表
type StockLossAnnotation struct {
	BaseModel
	StockLossUuid uint64 `gorm:"column:stock_loss_uuid;not null;default:0;index:idx_stock_loss_uuid;comment:报损单UUID" json:"stock_loss_uuid"`
	Action        string `gorm:"column:action;type:varchar(20);not null;default:'';comment:操作类型 submit/resubmit/approve/reject" json:"action"`
	Content       string `gorm:"column:content;type:text;comment:批注内容" json:"content"`
	OperatorUuid  uint64 `gorm:"column:operator_uuid;not null;default:0;comment:操作人UUID" json:"operator_uuid"`
	OperatorName  string `gorm:"column:operator_name;type:varchar(100);not null;default:'';comment:操作人姓名" json:"operator_name"`

	// 关联报损单
	StockLoss *StockLoss `gorm:"foreignKey:StockLossUuid;references:Uuid"`
}

// TableName 指定表名
func (StockLossAnnotation) TableName() string {
	return "ttpos_stock_loss_annotation"
}

// 批注操作类型常量
const (
	StockLossActionSubmit   = "submit"   // 提交
	StockLossActionResubmit = "resubmit" // 重新提交
	StockLossActionApprove  = "approve"  // 审核通过
	StockLossActionReject   = "reject"   // 驳回
)

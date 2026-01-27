package model

// StockReconciliationAnnotation 盘点单批注表
type StockReconciliationAnnotation struct {
	BaseModel
	StockReconciliationUuid uint64 `gorm:"column:stock_reconciliation_uuid;not null;default:0;index:idx_stock_reconciliation_uuid;comment:盘点单UUID" json:"stock_reconciliation_uuid"`
	AnnotationType          int    `gorm:"column:annotation_type;not null;default:0;comment:批注类型 1-重新发起 2-驳回 3-通过" json:"annotation_type"`
	Content                 string `gorm:"column:content;type:text;comment:批注内容" json:"content"`
}

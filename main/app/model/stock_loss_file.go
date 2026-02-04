package model

// StockLossFile 报损单附件关联表
type StockLossFile struct {
	BaseModel
	StockLossUuid uint64 `gorm:"column:stock_loss_uuid;not null;default:0;index:idx_stock_loss_uuid;comment:报损单UUID" json:"stock_loss_uuid"`
	FileUuid      uint64 `gorm:"column:file_uuid;not null;default:0;index:idx_file_uuid;comment:文件UUID" json:"file_uuid"`
	SortOrder     int    `gorm:"column:sort_order;not null;default:0;comment:排序顺序" json:"sort_order"`

	// 关联报损单
	StockLoss *StockLoss `gorm:"foreignKey:StockLossUuid;references:Uuid"`
	// 关联文件
	File *File `gorm:"foreignKey:FileUuid;references:Uuid"`
}

// TableName 指定表名
func (StockLossFile) TableName() string {
	return "ttpos_stock_loss_file"
}

// SetNil 设置关联为空
func (slf *StockLossFile) SetNil() {
	slf.StockLoss = nil
	slf.File = nil
}

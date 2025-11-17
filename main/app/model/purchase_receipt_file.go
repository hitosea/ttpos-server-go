package model

// PurchaseReceiptFile 收货单附件表 ttpos_purchase_receipt_file
type PurchaseReceiptFile struct {
	BaseModel
	ReceiptOrderUuid uint64 `gorm:"column:receipt_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:收货单UUID;index:idx_receipt_order_uuid" json:"receipt_order_uuid"`
	FileUuid         uint64 `gorm:"column:file_uuid;type:bigint(20) unsigned;not null;default:0;comment:文件UUID;index:idx_file_uuid" json:"file_uuid"`
	SortOrder        int    `gorm:"column:sort_order;type:int(11);not null;default:0;comment:排序顺序" json:"sort_order"`

	// 关联关系
	ReceiptOrder *PurchaseReceiptOrder `gorm:"foreignKey:ReceiptOrderUuid;references:Uuid" json:"receipt_order,omitempty"`
	File         *File                 `gorm:"foreignKey:FileUuid;references:Uuid" json:"file,omitempty"`
}

// TableName 指定表名
func (PurchaseReceiptFile) TableName() string {
	return "ttpos_purchase_receipt_file"
}

// SetNil 设置为空
func (prf *PurchaseReceiptFile) SetNil() {
	prf.ReceiptOrder = nil
	prf.File = nil
}

package model

// TransferOrderFile 调拨单附件表 ttpos_transfer_order_file
type TransferOrderFile struct {
	BaseModel
	TransferOrderUuid uint64 `gorm:"column:transfer_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:调拨单UUID;index:idx_transfer_order_uuid" json:"transfer_order_uuid"`
	FileUuid          uint64 `gorm:"column:file_uuid;type:bigint(20) unsigned;not null;default:0;comment:文件UUID;index:idx_file_uuid" json:"file_uuid"`
	SortOrder         int    `gorm:"column:sort_order;type:int(11);not null;default:0;comment:排序顺序" json:"sort_order"`

	// 关联关系
	TransferOrder *TransferOrder `gorm:"foreignKey:TransferOrderUuid;references:Uuid" json:"transfer_order,omitempty"`
	File          *File          `gorm:"foreignKey:FileUuid;references:Uuid" json:"file,omitempty"`
}

// TableName 指定表名
func (TransferOrderFile) TableName() string {
	return "ttpos_transfer_order_file"
}

// SetNil 设置为空
func (tof *TransferOrderFile) SetNil() {
	tof.TransferOrder = nil
	tof.File = nil
}

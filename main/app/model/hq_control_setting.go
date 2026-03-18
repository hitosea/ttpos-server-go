package model

// HqControlSetting 总部控制设置（KV 表，仅总部可写，子店通过总部 DB 读取）
type HqControlSetting struct {
	BaseModel
	FieldType   string `gorm:"column:field_type;type:varchar(50);NOT NULL;comment:'字段类型: dine_shelf/takeout_shelf/takeout_price/negative_stock'" json:"field_type"`
	ControlMode int    `gorm:"column:control_mode;type:int(11);default:0;NOT NULL;comment:'控制模式: 0-分开控制 1-统一控制'" json:"control_mode"`
}

func (HqControlSetting) TableName() string {
	return GetTableName("hq_control_setting")
}

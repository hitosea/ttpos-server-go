package model

// TakeoutOrderUpdateLog 外卖订单更新日志表
type TakeoutOrderUpdateLog struct {
	BaseModel
	TakeoutOrderUuid uint64 `gorm:"column:takeout_order_uuid;type:bigint(20) unsigned;default:0;comment:外卖订单UUID;NOT NULL" json:"takeout_order_uuid"`
	OldData          string `gorm:"column:old_data;type:text;comment:更新前订单数据(JSON格式,包含订单主表、商品、修饰符等完整数据)" json:"old_data"`
	NewData          string `gorm:"column:new_data;type:text;comment:更新后订单数据(JSON格式,包含订单主表、商品、修饰符等完整数据)" json:"new_data"`
}

// TableName 表名
func (*TakeoutOrderUpdateLog) TableName() string {
	return "ttpos_takeout_order_update_log"
}

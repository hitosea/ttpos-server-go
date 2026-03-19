package model

// HqFieldOverride 总部字段覆盖追踪表 `ttpos_hq_field_override`
// 记录子店对总部来源商品/物品的可修改字段的覆盖情况
type HqFieldOverride struct {
	BaseModel
	EntityType string `gorm:"column:entity_type;type:varchar(50);NOT NULL;comment:'实体类型: product/product_takeout/material'" json:"entity_type"`
	EntityUuid uint64 `gorm:"column:entity_uuid;type:bigint(20) unsigned;NOT NULL;comment:'实体UUID'" json:"entity_uuid"`
	FieldType  string `gorm:"column:field_type;type:varchar(50);NOT NULL;comment:'字段类型: dine_shelf/takeout_shelf/takeout_price/safety_stock/negative_stock'" json:"field_type"`
}

// TableName 指定表名
func (HqFieldOverride) TableName() string {
	return GetTableName("hq_field_override")
}

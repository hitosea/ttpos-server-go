package model

// PurchaseLimitSchemeItem 限购方案物品配置表 ttpos_purchase_limit_scheme_item
type PurchaseLimitSchemeItem struct {
	BaseModel
	SchemeUuid      uint64  `gorm:"column:scheme_uuid;type:bigint(20) unsigned;not null;default:0;comment:'限购方案UUID';index:idx_scheme_material" json:"scheme_uuid"`
	MaterialCode    string  `gorm:"column:material_code;type:varchar(50);not null;default:'';comment:'物品编码';index:idx_scheme_material" json:"material_code"`
	UnitCode        string  `gorm:"column:unit_code;type:varchar(20);not null;default:'';comment:'单位编码'" json:"unit_code"`
	QuotaLimit      float64 `gorm:"column:quota_limit;type:decimal(20,8);not null;default:0;comment:'最大采购数量（0=不限制）'" json:"quota_limit"`
	MinQuotaLimit   float64 `gorm:"column:min_quota_limit;type:decimal(20,8);not null;default:0;comment:'最小采购数量（0=不限制）'" json:"min_quota_limit"`
	IsAllowPurchase string  `gorm:"column:is_allow_purchase;type:varchar(10);not null;default:'yes';comment:'是否允许采购 yes/no'" json:"is_allow_purchase"`
}

// TableName 指定表名
func (PurchaseLimitSchemeItem) TableName() string {
	return "ttpos_purchase_limit_scheme_item"
}

package model

// PurchaseLimitSchemeShop 限购方案门店配置表 ttpos_purchase_limit_scheme_shop
type PurchaseLimitSchemeShop struct {
	BaseModel
	SchemeUuid  uint64 `gorm:"column:scheme_uuid;type:bigint(20) unsigned;not null;default:0;comment:'限购方案UUID';index:idx_scheme_company" json:"scheme_uuid"`
	CompanyUuid uint64 `gorm:"column:company_uuid;type:bigint(20) unsigned;not null;default:0;comment:'门店UUID';index:idx_scheme_company" json:"company_uuid"`
}

// TableName 指定表名
func (PurchaseLimitSchemeShop) TableName() string {
	return "ttpos_purchase_limit_scheme_shop"
}

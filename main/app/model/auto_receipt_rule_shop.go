package model

// AutoReceiptRuleShop 品采自动收货规则门店子表（saas主库）
type AutoReceiptRuleShop struct {
	BaseModel
	RuleUuid        uint64 `gorm:"column:rule_uuid;type:bigint(20) unsigned;not null;default:0;comment:关联规则主表UUID" json:"rule_uuid"`
	ShopCompanyUuid uint64 `gorm:"column:shop_company_uuid;type:bigint(20) unsigned;not null;default:0;comment:门店company_uuid" json:"shop_company_uuid"`
}

func (AutoReceiptRuleShop) TableName() string {
	return "ttpos_auto_receipt_rule_shop"
}

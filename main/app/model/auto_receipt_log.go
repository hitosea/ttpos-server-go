package model

// AutoReceiptLog 品采自动收货记录表（saas主库）
type AutoReceiptLog struct {
	BaseModel
	HeadquarterCompanyUuid uint64 `gorm:"column:headquarter_company_uuid;type:bigint(20) unsigned;not null;default:0;comment:总部company_uuid（租户隔离）" json:"headquarter_company_uuid"`
	RuleUuid               uint64 `gorm:"column:rule_uuid;type:bigint(20) unsigned;not null;default:0;comment:触发规则UUID" json:"rule_uuid"`
	ShopCompanyUuid        uint64 `gorm:"column:shop_company_uuid;type:bigint(20) unsigned;not null;default:0;comment:门店company_uuid" json:"shop_company_uuid"`
	ReceiptOrderUuid       uint64 `gorm:"column:receipt_order_uuid;type:bigint(20) unsigned;not null;default:0;comment:收货单UUID" json:"receipt_order_uuid"`
	ReceiptOrderNo         string `gorm:"column:receipt_order_no;type:varchar(100);not null;default:'';comment:收货单号" json:"receipt_order_no"`
	ReceiptErpOrderNo      string `gorm:"column:receipt_erp_order_no;type:varchar(255);not null;default:'';comment:收货单ERP单号" json:"receipt_erp_order_no"`
	ReceiptTime            int64  `gorm:"column:receipt_time;type:int(10);not null;default:0;comment:收货日期（当天0点时间戳，筛选用）" json:"receipt_time"`
}

func (AutoReceiptLog) TableName() string {
	return "ttpos_auto_receipt_log"
}

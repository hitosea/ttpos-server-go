package model

// AutoReceiptRule 品采自动收货规则主表（saas主库）
type AutoReceiptRule struct {
	BaseModel
	HeadquarterCompanyUuid uint64 `gorm:"column:headquarter_company_uuid;type:bigint(20) unsigned;not null;default:0;comment:总部company_uuid（租户隔离）" json:"headquarter_company_uuid"`
	Name                   string `gorm:"column:name;type:text;comment:规则名称（多语言JSON）" json:"name"`
	WarehouseErpCode       string `gorm:"column:warehouse_erp_code;type:varchar(100);not null;default:'';comment:发货仓库ERP编码" json:"warehouse_erp_code"`
	DelayDays              int    `gorm:"column:delay_days;type:int(11);not null;default:0;comment:DN发送后N天自动收货（0=当天24:00）" json:"delay_days"`
	Status                 *int   `gorm:"column:status;type:tinyint(4);not null;default:1;comment:状态：1=启用，0=禁用" json:"status"`
}

func (AutoReceiptRule) TableName() string {
	return "ttpos_auto_receipt_rule"
}

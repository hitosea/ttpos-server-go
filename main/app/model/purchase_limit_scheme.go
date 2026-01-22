package model

// PurchaseLimitScheme 限购方案表 ttpos_purchase_limit_scheme
type PurchaseLimitScheme struct {
	BaseModel
	Name            string `gorm:"column:name;type:text;not null;comment:'方案名称'" json:"name"`
	Status          int8   `gorm:"column:status;type:tinyint(1);not null;default:1;comment:'状态：0=关闭，1=开启';index:idx_status" json:"status"`
	ApplyToAllShops int8   `gorm:"column:apply_to_all_shops;type:tinyint(1);not null;default:1;comment:'是否适用所有门店：0=否，1=是'" json:"apply_to_all_shops"`
	DailyLimit      int    `gorm:"column:daily_limit;type:int(10);not null;default:0;comment:'每日申请次数限制（0=不限制）'" json:"daily_limit"`
	Weekdays        string `gorm:"column:weekdays;type:varchar(20);not null;default:'';comment:'星期配置（逗号分隔，1-7表示周一到周日）'" json:"weekdays"`
}

// TableName 指定表名
func (PurchaseLimitScheme) TableName() string {
	return "ttpos_purchase_limit_scheme"
}

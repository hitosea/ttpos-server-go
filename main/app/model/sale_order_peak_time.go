package model

// SaleOrderPeakTime 销售订单高峰时间 `ttpos_sale_order_peak_time`
type SaleOrderPeakTime struct {
	// 基础字段
	BaseModel
	// 日期
	Date   int64   `gorm:"column:date;type:int(11);default:0;comment:日期（天）" json:"date"`
	Hour   int64   `gorm:"column:hour;type:int(11);default:0;comment:小时" json:"hour"`
	Num    int     `gorm:"column:num;type:int(11);default:0;comment:订单数" json:"num"`
	Amount float64 `gorm:"column:amount;type:decimal(12,2);default:0.00;comment:订单金额" json:"amount"`
	// 关联ID
	CashierUuid uint64 `gorm:"column:cashier_uuid;type:bigint(20);not null;default:0;comment:收银员ID" json:"cashier_uuid"`
}

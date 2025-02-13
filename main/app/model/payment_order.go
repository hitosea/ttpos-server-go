package model

// PaymentMethod 支付方式 `ttpos_payment_method`
type PaymentMethod struct {
	BaseModel
	Name                 string  `gorm:"default:'';column:name;comment:支付方式名称" json:"name"`
	PaymentName          string  `gorm:"default:'';column:payment_name;comment:支付名称" json:"payment_name"`
	LogoFileUuid         uint64  `gorm:"default:0;column:logo_file_uuid;comment:logo图片UUID" json:"logo_file_uuid"`
	QrCodeFileUuid       uint64  `gorm:"default:0;column:qrcode_file_uuid;comment:二维码图片UUID" json:"qrcode_file_uuid"`
	FeePercent           float64 `gorm:"default:0.00;column:fee_percent;comment:手续费百分比" json:"fee_percent"`
	OrderBy              int     `gorm:"default:0;column:order_by;comment:排序" json:"order_by"`
	IsShowCashier        bool    `gorm:"default:1;column:is_show_cashier;comment:是否在收银员界面显示" json:"is_show_cashier"`
	IsShowAssistant      bool    `gorm:"default:1;column:is_show_assistant;comment:是否在助手界面显示" json:"is_show_assistant"`
	IsShowMemberRecharge bool    `gorm:"default:1;column:is_show_member_recharge;comment:是否在会员充值界面显示" json:"is_show_member_recharge"`
	Source               uint    `gorm:"default:0;column:source;comment:来源 0-系统 1-手动 2-LianLianPay" json:"source"`
	Status               uint    `gorm:"default:0;column:status;comment:状态 0-禁用 1-启用" json:"status"`
}

// PaymentOrder 支付订单 ttpos_payment_order
type PaymentOrder struct {
	BaseModel
	PaymentMethodName string  `gorm:"column:payment_method_name;type:varchar(255);default:'';comment:支付方式名称" json:"payment_method_name"`
	PaymentFeePercent float64 `gorm:"column:payment_fee_percent;type:decimal(5,2);default:0.00;comment:支付手续费百分比" json:"payment_fee_percent"`
	CurrencyUnit      string  `gorm:"column:currency_unit;type:varchar(10);default:'';comment:货币单位" json:"currency_unit"`
	PaymentAmount     float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0.00;comment:支付金额" json:"payment_amount"`
	Amount            float64 `gorm:"column:amount;type:decimal(12,2);default:0.00;comment:金额" json:"amount"`
	TransactionNumber string  `gorm:"column:transaction_number;type:varchar(255);default:'';comment:交易号" json:"transaction_number"`
	Status            uint    `gorm:"column:status;type:tinyint(1);default:0;comment:支付状态, 0-未支付 1-已支付 2-已退款" json:"status"`

	// 关联ID字段
	PaymentMethodUuid uint64 `gorm:"column:payment_method_uuid;type:bigint(20);default:0;comment:支付方式ID" json:"payment_method_uuid"`
	SaleOrderUuid     uint64 `gorm:"column:sale_order_uuid;type:bigint(20);default:0;comment:销售订单ID" json:"sale_order_uuid"`

	// 关联字段
	PaymentMethod PaymentMethod `gorm:"foreignKey:PaymentMethodUuid;references:uuid"`
}

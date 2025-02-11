package model

// 支付方式 PaymentMethod ttpos_payment_method
type PaymentMethod struct {
	ID                   uint    `gorm:"primaryKey;AUTO_INCREMENT;comment:自增ID" json:"id"`
	Uuid                 uint64  `gorm:"default:0;comment:支付方式ID" json:"uuid"`
	Name                 string  `gorm:"default:'';comment:支付方式名称" json:"name"`
	PaymentName          string  `gorm:"default:'';comment:支付名称" json:"payment_name"`
	LogoUrl              string  `gorm:"default:'';comment:logo图片URL" json:"logo_url"`
	QrCodeUrl            string  `gorm:"default:'';comment:二维码图片URL" json:"qrcode_url"`
	FeePercent           float64 `gorm:"default:0.00;comment:手续费百分比" json:"fee_percent"`
	OrderBy              int     `gorm:"default:0;comment:排序" json:"order_by"`
	IsShowCashier        bool    `gorm:"default:1;comment:是否在收银员界面显示" json:"is_show_cashier"`
	IsShowAssistant      bool    `gorm:"default:1;comment:是否在助手界面显示" json:"is_show_assistant"`
	IsShowMemberRecharge bool    `gorm:"default:1;comment:是否在会员充值界面显示" json:"is_show_member_recharge"`
	Source               uint    `gorm:"default:0;comment:来源 0-系统 1-手动 2-LianLianPay" json:"source"`
	Status               uint    `gorm:"default:0;comment:状态 0-禁用 1-启用" json:"status"`
	CreateTime           int     `gorm:"autoCreateTime;comment:创建时间（时间戳）" json:"create_time"`
	Provider             string  `gorm:"default:'';comment:提供者" json:"provider"`
}

// 支付订单 PaymentOrder ttpos_payment_order
type PaymentOrder struct {
	ID                uint    `gorm:"column:id;type:int(10);primary_key;AUTO_INCREMENT;comment:主键id" json:"id"`
	Uuid              uint64  `gorm:"column:uuid;type:bigint(20);default:0;comment:支付订单ID" json:"uuid"`
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

	// 时间相关字段
	CreateTime uint `gorm:"autoCreateTime;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime uint `gorm:"autoUpdateTime;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime uint `gorm:"column:delete_time;type:int(10);default:0;comment:删除时间（时间戳）" json:"delete_time"`

	// 关联字段
	PaymentMethod PaymentMethod `gorm:"foreignKey:PaymentMethodUuid;references:uuid"`
}

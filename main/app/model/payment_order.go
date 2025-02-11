package model

// 支付订单 PaymentOrder
type PaymentOrder struct {
	ID                uint    `gorm:"column:id;type:int(10);primary_key;AUTO_INCREMENT;comment:主键id" json:"id"`
	Uuid              uint64  `gorm:"column:uuid;type:bigint(20);default:0;comment:支付订单ID" json:"uuid"`
	PaymentTypeName   string  `gorm:"column:payment_type_name;type:varchar(255);default:'';comment:支付类型名称" json:"payment_type_name"`
	PaymentTypeUuid   uint64  `gorm:"column:payment_type_uuid;type:bigint(20);default:0;comment:支付类型ID" json:"payment_type_uuid"`
	PaymentFeePercent float64 `gorm:"column:payment_fee_percent;type:decimal(5,2);default:0.00;comment:支付手续费百分比" json:"payment_fee_percent"`
	SaleOrderUuid     uint64  `gorm:"column:sale_order_uuid;type:bigint(20);default:0;comment:销售订单ID" json:"sale_order_uuid"`
	CurrencyUnit      string  `gorm:"column:currency_unit;type:varchar(10);default:'';comment:货币单位" json:"currency_unit"`
	PaymentAmount     float64 `gorm:"column:payment_amount;type:decimal(12,2);default:0.00;comment:支付金额" json:"payment_amount"`
	Amount            float64 `gorm:"column:amount;type:decimal(12,2);default:0.00;comment:金额" json:"amount"`
	TransactionNumber string  `gorm:"column:transaction_number;type:varchar(255);default:'';comment:交易号" json:"transaction_number"`
	Status            uint    `gorm:"column:status;type:tinyint(1);default:0;comment:支付状态, 0-未支付 1-已支付 2-已退款" json:"status"`
	CreateTime        uint    `gorm:"autoCreateTime;comment:创建时间（时间戳）" json:"create_time"`
	UpdateTime        uint    `gorm:"autoUpdateTime;comment:更新时间（时间戳）" json:"update_time"`
	DeleteTime        uint    `gorm:"column:delete_time;type:int(10);default:0;comment:删除时间（时间戳）" json:"delete_time"`
}

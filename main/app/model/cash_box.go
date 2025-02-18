package model

// CashBox 钱箱表 ttpos_cash_box
type CashBox struct {
	BaseModel
	Name            string  `gorm:"column:name;type:varchar(255);comment:名称;NOT NULL" json:"name"`
	Balance         float64 `gorm:"column:balance;type:decimal(12,2);default:0.00;comment:钱箱余额;NOT NULL" json:"balance"`
	PreviousBalance float64 `gorm:"column:previous_balance;type:decimal(12,2);default:0.00;comment:上一班遗留备用金;NOT NULL" json:"previous_balance"`
	CashWithdrawal  float64 `gorm:"column:cash_withdrawal;type:decimal(12,2);default:0.00;comment:中途取出金额;NOT NULL" json:"cash_withdrawal"`
	CashDeposit     float64 `gorm:"column:cash_deposit;type:decimal(12,2);default:0.00;comment:中途存入金额;NOT NULL" json:"cash_deposit"`
}

// CashBoxLog 钱箱存取记录表 ttpos_cash_box_log
type CashBoxLog struct {
	BaseModel
	Type                  int     `gorm:"column:type;type:tinyint(1);default:0;comment:类型 1-取现 2-存现;NOT NULL" json:"type"`
	Scene                 int     `gorm:"column:scene;type:tinyint(1);default:0;comment:场景 1-支付 2-退货退款 3-取消付款 4-中途取出 5-中途存入;NOT NULL" json:"scene"`
	Amount                float64 `gorm:"column:amount;type:decimal(12,2);default:0.00;comment:金额;NOT NULL" json:"amount"`
	Remark                string  `gorm:"column:remark;type:varchar(255);comment:备注;NOT NULL" json:"remark"`
	PaymentBillUuid       uint64  `gorm:"column:payment_bill_uuid;type:bigint(20) unsigned;default:0;comment:付款单ID,场景为1时必填;NOT NULL" json:"payment_bill_uuid"`
	ReturnOrderUuid       uint64  `gorm:"column:return_order_uuid;type:bigint(20) unsigned;default:0;comment:退货单ID,场景为2时必填;NOT NULL" json:"return_order_uuid"`
	RefundOrderAmountUuid uint64  `gorm:"column:refund_order_amount_uuid;type:bigint(20) unsigned;default:0;comment:退款单金额ID,场景为3时必填;NOT NULL" json:"refund_order_amount_uuid"`
}

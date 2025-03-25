package model

import "github.com/shopspring/decimal"

// CashBox 钱箱表 ttpos_cash_box
type CashBox struct {
	BaseModel
	Name            string  `gorm:"column:name;type:varchar(255);comment:名称;NOT NULL" json:"name"`
	Balance         float64 `gorm:"column:balance;type:decimal(12,2);default:0.00;comment:钱箱余额;NOT NULL" json:"balance"`
	FrozenBalance   float64 `gorm:"column:frozen_balance;type:decimal(12,2);default:0.00;comment:冻结金额。冻结金额不能使用，在前端显示为已扣除或已增加。冻结金额可为负数。钱箱余额=钱箱余额+冻结金额;NOT NULL" json:"frozen_balance"`
	PreviousBalance float64 `gorm:"column:previous_balance;type:decimal(12,2);default:0.00;comment:上一班遗留备用金;NOT NULL" json:"previous_balance"`
	CashWithdrawal  float64 `gorm:"column:cash_withdrawal;type:decimal(12,2);default:0.00;comment:中途取出金额;NOT NULL" json:"cash_withdrawal"`
	CashDeposit     float64 `gorm:"column:cash_deposit;type:decimal(12,2);default:0.00;comment:中途存入金额;NOT NULL" json:"cash_deposit"`
}

// GetBalance 获取钱箱余额. 钱箱余额=钱箱余额+冻结金额
func (c *CashBox) GetBalance() float64 {
	return decimal.NewFromFloat(c.Balance).Add(decimal.NewFromFloat(c.FrozenBalance)).Truncate(2).InexactFloat64()
}

// CashBoxLog 钱箱存取记录表 ttpos_cash_box_log
type CashBoxLog struct {
	BaseModel
	Scene                 int     `gorm:"column:scene;type:tinyint(1);default:0;comment:场景 1-销售订单支付 2-退货退款 3-取消付款 4-中途取出 5-中途存入 6-会员充值 7-结账找零 8-交班;NOT NULL" json:"scene"`
	Amount                float64 `gorm:"column:amount;type:decimal(12,2);default:0.00;comment:金额.变动的金额NOT NULL" json:"amount"`
	Remark                string  `gorm:"column:remark;type:varchar(255);comment:备注;NOT NULL" json:"remark"`
	Processed             int     `gorm:"column:processed;type:tinyint(1);default:0;comment:是否已处理,0-未处理 1-已处理. 用于处理钱箱余额变动，修改钱箱的余额并清0冻结的余额;NOT NULL" json:"processed"`
	RelatedUuid           uint64  `gorm:"column:related_uuid;type:bigint(20) unsigned;default:0;comment:关联的充值订单、销售订单ID,场景为1、6时必填;NOT NULL" json:"related_uuid"`
	ReturnOrderUuid       uint64  `gorm:"column:return_order_uuid;type:bigint(20) unsigned;default:0;comment:退货单ID,场景为2时必填;NOT NULL" json:"return_order_uuid"`
	RefundOrderAmountUuid uint64  `gorm:"column:refund_order_amount_uuid;type:bigint(20) unsigned;default:0;comment:退款单金额ID,场景为3时必填;NOT NULL" json:"refund_order_amount_uuid"`
}

package model

import "database/sql"

// StatisticsShiftSaleRefundAmount 当班用餐订单退款金额
type StatisticsShiftSaleRefundAmount struct {
	ShiftNo         string          `json:"shift_no"`
	RefundAmount    sql.NullFloat64 `json:"refund_amount"`
	RefundTaxAmount sql.NullFloat64 `json:"refund_tax_amount"`
}

// StatisticsShiftRechargeRefundAmount 当班充值订单退款金额
type StatisticsShiftRechargeRefundAmount struct {
	ShiftNo         string          `json:"shift_no"`
	RefundAmount    sql.NullFloat64 `json:"refund_amount"`
	RefundTaxAmount sql.NullFloat64 `json:"refund_tax_amount"`
}

// StatisticsPaymentMethodAmount 支付方式累计收入
type StatisticsPaymentMethodAmount struct {
	PaymentName  string          `gorm:"column:payment_name;comment:支付方式名称"`
	PaymentCode  int             `gorm:"column:payment_code;comment:支付方式编码"`
	PayAmount    sql.NullFloat64 `gorm:"column:pay_amount;comment:累计支付金额"`
	RefundAmount sql.NullFloat64 `gorm:"column:refund_amount;comment:累计退款金额"`
}

// StatisticsSaleFreeAmount 当班用餐订单免单金额
type StatisticsSaleFreeAmount struct {
	ShiftNo    string          `json:"shift_no"`
	FreeAmount sql.NullFloat64 `json:"free_amount"`
}

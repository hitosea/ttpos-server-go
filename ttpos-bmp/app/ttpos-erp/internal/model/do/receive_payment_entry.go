// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ReceivePaymentEntry is the golang structure of table erp_receive_payment_entry for DAO operations like Where/Data.
type ReceivePaymentEntry struct {
	g.Meta           `orm:"table:erp_receive_payment_entry, do:true"`
	Id               any // ID
	SaleOrderUuid    any // TTPOS订单UUID
	ModeOfPayment    any // 支付方式
	PaymentEntryName any // ERP Payment Entry名称
	Docstatus        any // 文档状态
	PaidAmount       any // 支付金额
	SiteCode         any // ERP site code
	ReqBody          any // 请求文本
	RespBody         any // 响应文本
	RetryCount       any // 重试次数
	CreatedAt        any // 创建时间
	UpdatedAt        any // 更新时间
}

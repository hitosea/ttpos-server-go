// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ReceiveSalesInvoice is the golang structure of table erp_receive_sales_invoice for DAO operations like Where/Data.
type ReceiveSalesInvoice struct {
	g.Meta            `orm:"table:erp_receive_sales_invoice, do:true"`
	Id                any // ID
	OrderNo           any // TTPOS订单号
	SaleOrderUuid     any // TTPOS订单UUID（幂等键）
	PosProfile        any // POS Profile名称
	PostingDatetime   any // 过账时间戳
	Docstatus         any // 文档状态: 0=Draft 1=Submitted 2=Cancelled
	SalesInvoiceName  any // ERP Sales Invoice名称
	PaymentEntryNames any // ERP Payment Entry名称(JSON)
	SiteCode          any // ERP site code
	ReqMessage        any // 请求数据(base64)
	RespMessage       any // 响应数据(base64)
	ReqBody           any // 请求文本
	RespBody          any // 响应文本
	RetryCount        any // 重试次数
	MqMsgId           any // 最后一次MQ消息ID
	CreatedAt         any // 创建时间
	UpdatedAt         any // 更新时间
}

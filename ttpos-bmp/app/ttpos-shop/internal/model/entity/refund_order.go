// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// RefundOrder is the golang structure for table refund_order.
type RefundOrder struct {
	Id               uint    `json:"id"               orm:"id"                 description:"自增ID"`              // 自增ID
	Uuid             uint64  `json:"uuid"             orm:"uuid"               description:"退款单唯一标识符"`          // 退款单唯一标识符
	SaleOrderUuid    uint64  `json:"saleOrderUuid"    orm:"sale_order_uuid"    description:"销售订单ID"`            // 销售订单ID
	SaleOrderNo      string  `json:"saleOrderNo"      orm:"sale_order_no"      description:"销售订单号"`             // 销售订单号
	PaymentOrderUuid uint64  `json:"paymentOrderUuid" orm:"payment_order_uuid" description:"支付单ID"`             // 支付单ID
	RefundType       int     `json:"refundType"       orm:"refund_type"        description:"退款类型,1-反结账,2-取消付款"` // 退款类型,1-反结账,2-取消付款
	Amount           float64 `json:"amount"           orm:"amount"             description:"退款金额"`              // 退款金额
	Reason           string  `json:"reason"           orm:"reason"             description:"退款原因"`              // 退款原因
	Status           int     `json:"status"           orm:"status"             description:"退款状态"`              // 退款状态
	CreateTime       uint    `json:"createTime"       orm:"create_time"        description:"创建时间(时间戳)"`         // 创建时间(时间戳)
	UpdateTime       uint    `json:"updateTime"       orm:"update_time"        description:"更新时间(时间戳)"`         // 更新时间(时间戳)
	DeleteTime       uint    `json:"deleteTime"       orm:"delete_time"        description:"删除时间(时间戳)"`         // 删除时间(时间戳)
}

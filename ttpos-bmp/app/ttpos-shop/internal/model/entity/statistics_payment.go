// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// StatisticsPayment is the golang structure for table statistics_payment.
type StatisticsPayment struct {
	Id                uint    `json:"id"                orm:"id"                  description:"自增ID"`     // 自增ID
	Uuid              uint64  `json:"uuid"              orm:"uuid"                description:"UUID"`     // UUID
	SaleBillUuid      uint64  `json:"saleBillUuid"      orm:"sale_bill_uuid"      description:"销售单UUID"`  // 销售单UUID
	SaleOrderUuid     uint64  `json:"saleOrderUuid"     orm:"sale_order_uuid"     description:"销售订单UUID"` // 销售订单UUID
	DutyNo            string  `json:"dutyNo"            orm:"duty_no"             description:"当班编号"`     // 当班编号
	DeskUuid          uint64  `json:"deskUuid"          orm:"desk_uuid"           description:"桌台UUID"`   // 桌台UUID
	PaymentMethodUuid uint64  `json:"paymentMethodUuid" orm:"payment_method_uuid" description:"支付方式UUID"` // 支付方式UUID
	PaymentAmount     float64 `json:"paymentAmount"     orm:"payment_amount"      description:"支付金额"`     // 支付金额
	RefundAmount      float64 `json:"refundAmount"      orm:"refund_amount"       description:"退款金额"`     // 退款金额
	CompleteTime      uint    `json:"completeTime"      orm:"complete_time"       description:"完成时间"`     // 完成时间
	CreateTime        uint    `json:"createTime"        orm:"create_time"         description:"创建时间"`     // 创建时间
	UpdateTime        uint    `json:"updateTime"        orm:"update_time"         description:"更新时间"`     // 更新时间
	DeleteTime        uint    `json:"deleteTime"        orm:"delete_time"         description:"删除时间"`     // 删除时间
}

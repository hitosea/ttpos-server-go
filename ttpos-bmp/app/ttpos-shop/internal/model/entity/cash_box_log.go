// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// CashBoxLog is the golang structure for table cash_box_log.
type CashBoxLog struct {
	Id                    uint    `json:"id"                    orm:"id"                       description:"自增ID"`                                                  // 自增ID
	Uuid                  uint64  `json:"uuid"                  orm:"uuid"                     description:"钱箱ID"`                                                  // 钱箱ID
	Scene                 int     `json:"scene"                 orm:"scene"                    description:"场景 1-销售订单支付 2-退货退款 3-取消付款 4-中途取出 5-中途存入 6-会员充值 7-结账找零"` // 场景 1-销售订单支付 2-退货退款 3-取消付款 4-中途取出 5-中途存入 6-会员充值 7-结账找零
	Amount                float64 `json:"amount"                orm:"amount"                   description:"金额"`                                                    // 金额
	Remark                string  `json:"remark"                orm:"remark"                   description:"备注"`                                                    // 备注
	Processed             int     `json:"processed"             orm:"processed"                description:"是否已处理,0-未处理 1-已处理. 用于处理钱箱余额变动，修改钱箱的余额并清0冻结的余额"`         // 是否已处理,0-未处理 1-已处理. 用于处理钱箱余额变动，修改钱箱的余额并清0冻结的余额
	RelatedUuid           uint64  `json:"relatedUuid"           orm:"related_uuid"             description:"关联的充值订单、销售订单ID,场景为1、6时必填"`                              // 关联的充值订单、销售订单ID,场景为1、6时必填
	ReturnOrderUuid       uint64  `json:"returnOrderUuid"       orm:"return_order_uuid"        description:"退货单ID,场景为2时必填"`                                         // 退货单ID,场景为2时必填
	RefundOrderAmountUuid uint64  `json:"refundOrderAmountUuid" orm:"refund_order_amount_uuid" description:"退款单金额ID,场景为3时必填"`                                       // 退款单金额ID,场景为3时必填
	CreateTime            uint    `json:"createTime"            orm:"create_time"              description:"创建时间(时间戳)"`                                             // 创建时间(时间戳)
	UpdateTime            uint    `json:"updateTime"            orm:"update_time"              description:"更新时间(时间戳)"`                                             // 更新时间(时间戳)
	DeleteTime            uint    `json:"deleteTime"            orm:"delete_time"              description:"删除时间(时间戳)"`                                             // 删除时间(时间戳)
}

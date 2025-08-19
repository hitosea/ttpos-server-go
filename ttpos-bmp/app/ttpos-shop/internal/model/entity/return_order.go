// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ReturnOrder is the golang structure for table return_order.
type ReturnOrder struct {
	Id                  uint    `json:"id"                  orm:"id"                    description:"自增ID"`                         // 自增ID
	Uuid                uint64  `json:"uuid"                orm:"uuid"                  description:"退货单唯一标识符"`                     // 退货单唯一标识符
	RelatedOrderType    uint    `json:"relatedOrderType"    orm:"related_order_type"    description:"关联订单类型：0-销售订单；1-充值订单"`         // 关联订单类型：0-销售订单；1-充值订单
	RelatedOrderUuid    uint64  `json:"relatedOrderUuid"    orm:"related_order_uuid"    description:"关联订单ID"`                       // 关联订单ID
	RelatedOrderNo      string  `json:"relatedOrderNo"      orm:"related_order_no"      description:"关联订单号"`                        // 关联订单号
	LlReturnOrderId     string  `json:"llReturnOrderId"     orm:"ll_return_order_id"    description:"连连退款订单ID, 用来重新发起退款"`           // 连连退款订单ID, 用来重新发起退款
	IsReverseSettlement uint    `json:"isReverseSettlement" orm:"is_reverse_settlement" description:"是否反结账：0-否；1-是"`                // 是否反结账：0-否；1-是
	ReturnType          uint    `json:"returnType"          orm:"return_type"           description:"退货类型,1-整单退货,2-部分退货"`           // 退货类型,1-整单退货,2-部分退货
	RefundAmount        float64 `json:"refundAmount"        orm:"refund_amount"         description:"退款金额,包括税额"`                    // 退款金额,包括税额
	Unit                string  `json:"unit"                orm:"unit"                  description:"货币单位"`                         // 货币单位
	RefundTaxAmount     float64 `json:"refundTaxAmount"     orm:"refund_tax_amount"     description:"退款税额"`                         // 退款税额
	RefundReason        string  `json:"refundReason"        orm:"refund_reason"         description:"退款原因"`                         // 退款原因
	BankCode            string  `json:"bankCode"            orm:"bank_code"             description:"银行编码 - 当存在QR PromptPay的时候需要传"` // 银行编码 - 当存在QR PromptPay的时候需要传
	AccountNo           string  `json:"accountNo"           orm:"account_no"            description:"账号 - 当存在QR PromptPay的时候需要传"`   // 账号 - 当存在QR PromptPay的时候需要传
	AccountName         string  `json:"accountName"         orm:"account_name"          description:"账户名称 - 当存在QR PromptPay的时候需要传"` // 账户名称 - 当存在QR PromptPay的时候需要传
	CreateTime          uint    `json:"createTime"          orm:"create_time"           description:"创建时间(时间戳)"`                    // 创建时间(时间戳)
	UpdateTime          uint    `json:"updateTime"          orm:"update_time"           description:"更新时间(时间戳)"`                    // 更新时间(时间戳)
	DeleteTime          uint    `json:"deleteTime"          orm:"delete_time"           description:"删除时间(时间戳)"`                    // 删除时间(时间戳)
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ReturnOrder is the golang structure of table ttpos_return_order for DAO operations like Where/Data.
type ReturnOrder struct {
	g.Meta              `orm:"table:ttpos_return_order, do:true"`
	Id                  interface{} // 自增ID
	Uuid                interface{} // 退货单唯一标识符
	RelatedOrderType    interface{} // 关联订单类型：0-销售订单；1-充值订单
	RelatedOrderUuid    interface{} // 关联订单ID
	RelatedOrderNo      interface{} // 关联订单号
	LlReturnOrderId     interface{} // 连连退款订单ID, 用来重新发起退款
	IsReverseSettlement interface{} // 是否反结账：0-否；1-是
	ReturnType          interface{} // 退货类型,1-整单退货,2-部分退货
	RefundAmount        interface{} // 退款金额,包括税额
	Unit                interface{} // 货币单位
	RefundTaxAmount     interface{} // 退款税额
	RefundReason        interface{} // 退款原因
	BankCode            interface{} // 银行编码 - 当存在QR PromptPay的时候需要传
	AccountNo           interface{} // 账号 - 当存在QR PromptPay的时候需要传
	AccountName         interface{} // 账户名称 - 当存在QR PromptPay的时候需要传
	CreateTime          interface{} // 创建时间(时间戳)
	UpdateTime          interface{} // 更新时间(时间戳)
	DeleteTime          interface{} // 删除时间(时间戳)
}

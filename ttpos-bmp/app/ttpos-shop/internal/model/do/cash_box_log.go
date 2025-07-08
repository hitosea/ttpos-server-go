// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CashBoxLog is the golang structure of table ttpos_cash_box_log for DAO operations like Where/Data.
type CashBoxLog struct {
	g.Meta                `orm:"table:ttpos_cash_box_log, do:true"`
	Id                    interface{} // 自增ID
	Uuid                  interface{} // 钱箱ID
	Scene                 interface{} // 场景 1-销售订单支付 2-退货退款 3-取消付款 4-中途取出 5-中途存入 6-会员充值 7-结账找零
	Amount                interface{} // 金额
	Remark                interface{} // 备注
	Processed             interface{} // 是否已处理,0-未处理 1-已处理. 用于处理钱箱余额变动，修改钱箱的余额并清0冻结的余额
	RelatedUuid           interface{} // 关联的充值订单、销售订单ID,场景为1、6时必填
	ReturnOrderUuid       interface{} // 退货单ID,场景为2时必填
	RefundOrderAmountUuid interface{} // 退款单金额ID,场景为3时必填
	CreateTime            interface{} // 创建时间(时间戳)
	UpdateTime            interface{} // 更新时间(时间戳)
	DeleteTime            interface{} // 删除时间(时间戳)
}

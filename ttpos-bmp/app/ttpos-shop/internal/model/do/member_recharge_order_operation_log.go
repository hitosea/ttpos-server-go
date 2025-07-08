// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MemberRechargeOrderOperationLog is the golang structure of table ttpos_member_recharge_order_operation_log for DAO operations like Where/Data.
type MemberRechargeOrderOperationLog struct {
	g.Meta            `orm:"table:ttpos_member_recharge_order_operation_log, do:true"`
	Id                interface{} // 自增ID
	Uuid              interface{} // 会员充值订单操作日志ID
	OperatorName      interface{} // 操作员姓名
	OperatorEmail     interface{} // 操作员电子邮件
	Client            interface{} // 客户端信息
	Message           interface{} // 消息内容
	Action            interface{} // 操作
	Data              interface{} // 数据
	RechargeOrderUuid interface{} // 充值订单ID
	CreateTime        interface{} // 创建时间(时间戳)
	UpdateTime        interface{} // 更新时间(时间戳)
	DeleteTime        interface{} // 删除时间(时间戳)
}

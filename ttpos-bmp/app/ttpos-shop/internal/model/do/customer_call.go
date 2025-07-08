// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// CustomerCall is the golang structure of table ttpos_customer_call for DAO operations like Where/Data.
type CustomerCall struct {
	g.Meta     `orm:"table:ttpos_customer_call, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // 客户呼叫记录ID
	DeskUuid   interface{} // 桌台ID
	DeskNo     interface{} // 桌台编号,不随后台改变
	CallType   interface{} // 呼叫类型(1服务员,2结账)
	Status     interface{} // 状态,0-unhandled未处理 1-handled已处理
	IsSend     interface{} // 消息发送状态 0-否 1-是
	CreateTime interface{} // 创建时间(时间戳)
	UpdateTime interface{} // 更新时间(时间戳)
	DeleteTime interface{} // 删除时间(时间戳)
}

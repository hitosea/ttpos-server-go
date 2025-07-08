// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// WebSocketMsg is the golang structure of table ttpos_web_socket_msg for DAO operations like Where/Data.
type WebSocketMsg struct {
	g.Meta       `orm:"table:ttpos_web_socket_msg, do:true"`
	Id           interface{} //
	Uid          interface{} // 设备uid
	Type         interface{} // 消息类型
	Msg          interface{} // 详细消息
	Status       interface{} // 状态 0-未读 1-已读
	IsOffline    interface{} // 是否离线消息 0-否 1-是
	Remark       interface{} // 备注
	SourceClient interface{} // 来源客户端
	CompanyUuid  interface{} // 集团ID
	CreateTime   interface{} // 创建时间
	UpdateTime   interface{} // 更新时间
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MessageSendLog is the golang structure of table message_send_log for DAO operations like Where/Data.
type MessageSendLog struct {
	g.Meta       `orm:"table:message_send_log, do:true"`
	Id           interface{} // 日志ID
	MessageUuid  interface{} // 消息UUID
	SendTime     interface{} // 发送时间
	SendResult   interface{} // 发送结果(0-失败,1-成功)
	ErrorMessage interface{} // 错误信息
	RequestData  interface{} // 请求数据
	ResponseData interface{} // 响应数据
	CreatedAt    interface{} // 创建时间
}

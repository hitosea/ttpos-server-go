// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// WebsocketMsg is the golang structure of table websocket_msg for DAO operations like Where/Data.
type WebsocketMsg struct {
	g.Meta       `orm:"table:websocket_msg, do:true"`
	Id           any // 主键ID
	CompanyUuid  any // 公司UUID
	Uid          any // 用户/设备标识
	Msg          any // 消息内容（JSON格式）
	Type         any // 消息类型：heartbeat/order/notification/system/broadcast
	SourceClient any // 来源客户端：pos/tablet/kitchen/h5/mobile
	Status       any // 消息状态：0-待发送，1-发送中，2-发送成功，3-发送失败
	IsOffline    any // 是否离线消息：0-在线消息，1-离线消息
	CreateTime   any // 创建时间
	UpdateTime   any // 更新时间
	DeleteTime   any // 删除时间（软删除）
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// WebsocketMsg WebSocket消息记录表
type WebsocketMsg struct {
	g.Meta       `orm:"table:ttpos_websocket_msg, do:true"`
	Id           interface{} //
	Uuid         interface{} //
	CompanyUuid  interface{} //
	Uid          interface{} //
	Msg          interface{} //
	Type         interface{} //
	SourceClient interface{} //
	Status       interface{} //
	IsOffline    interface{} //
	CreateTime   interface{} //
	UpdateTime   interface{} //
	DeleteTime   interface{} //
}

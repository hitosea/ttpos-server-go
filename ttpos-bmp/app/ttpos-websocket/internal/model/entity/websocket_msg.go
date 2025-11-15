// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// WebsocketMsg is the golang structure for table websocket_msg.
type WebsocketMsg struct {
	Id           int64  `json:"id"           orm:"id"            description:"主键ID"`                                               // 主键ID
	CompanyUuid  int64  `json:"companyUuid"  orm:"company_uuid"  description:"公司UUID"`                                             // 公司UUID
	Uid          string `json:"uid"          orm:"uid"           description:"用户/设备标识"`                                            // 用户/设备标识
	Msg          string `json:"msg"          orm:"msg"           description:"消息内容（JSON格式）"`                                       // 消息内容（JSON格式）
	Type         string `json:"type"         orm:"type"          description:"消息类型：heartbeat/order/notification/system/broadcast"` // 消息类型：heartbeat/order/notification/system/broadcast
	SourceClient string `json:"sourceClient" orm:"source_client" description:"来源客户端：pos/tablet/kitchen/h5/mobile"`                 // 来源客户端：pos/tablet/kitchen/h5/mobile
	Status       int    `json:"status"       orm:"status"        description:"消息状态：0-待发送，1-发送中，2-发送成功，3-发送失败"`                     // 消息状态：0-待发送，1-发送中，2-发送成功，3-发送失败
	IsOffline    int    `json:"isOffline"    orm:"is_offline"    description:"是否离线消息：0-在线消息，1-离线消息"`                               // 是否离线消息：0-在线消息，1-离线消息
	CreateTime   int    `json:"createTime"   orm:"create_time"   description:"创建时间"`                                               // 创建时间
	UpdateTime   int    `json:"updateTime"   orm:"update_time"   description:"更新时间"`                                               // 更新时间
	DeleteTime   int    `json:"deleteTime"   orm:"delete_time"   description:"删除时间（软删除）"`                                          // 删除时间（软删除）
}

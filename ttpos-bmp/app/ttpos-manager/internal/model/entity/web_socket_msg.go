// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// WebSocketMsg is the golang structure for table web_socket_msg.
type WebSocketMsg struct {
	Id           uint   `json:"id"           orm:"id"            description:""`               //
	Uid          string `json:"uid"          orm:"uid"           description:"设备uid"`          // 设备uid
	Type         string `json:"type"         orm:"type"          description:"消息类型"`           // 消息类型
	Msg          string `json:"msg"          orm:"msg"           description:"详细消息"`           // 详细消息
	Status       int    `json:"status"       orm:"status"        description:"状态 0-未读 1-已读"`   // 状态 0-未读 1-已读
	IsOffline    int    `json:"isOffline"    orm:"is_offline"    description:"是否离线消息 0-否 1-是"` // 是否离线消息 0-否 1-是
	Remark       string `json:"remark"       orm:"remark"        description:"备注"`             // 备注
	SourceClient string `json:"sourceClient" orm:"source_client" description:"来源客户端"`          // 来源客户端
	CompanyUuid  int64  `json:"companyUuid"  orm:"company_uuid"  description:"集团ID"`           // 集团ID
	CreateTime   int    `json:"createTime"   orm:"create_time"   description:"创建时间"`           // 创建时间
	UpdateTime   int    `json:"updateTime"   orm:"update_time"   description:"更新时间"`           // 更新时间
}

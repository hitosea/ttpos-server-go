// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// WebsocketMsg WebSocket消息记录表
type WebsocketMsg struct {
	Id           uint   `json:"id"            description:""`
	Uuid         uint64 `json:"uuid"          description:""`
	CompanyUuid  uint64 `json:"company_uuid"  description:""`
	Uid          string `json:"uid"           description:""`
	Msg          string `json:"msg"           description:""`
	Type         string `json:"type"          description:""`
	SourceClient string `json:"source_client" description:""`
	Status       int    `json:"status"        description:""`
	IsOffline    int    `json:"is_offline"    description:""`
	CreateTime   uint   `json:"create_time"   description:""`
	UpdateTime   uint   `json:"update_time"   description:""`
	DeleteTime   uint   `json:"delete_time"   description:""`
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// CustomerCall is the golang structure for table customer_call.
type CustomerCall struct {
	Id         uint   `json:"id"         orm:"id"          description:"自增ID"`                           // 自增ID
	Uuid       uint64 `json:"uuid"       orm:"uuid"        description:"客户呼叫记录ID"`                       // 客户呼叫记录ID
	DeskUuid   uint64 `json:"deskUuid"   orm:"desk_uuid"   description:"桌台ID"`                           // 桌台ID
	DeskNo     string `json:"deskNo"     orm:"desk_no"     description:"桌台编号,不随后台改变"`                    // 桌台编号,不随后台改变
	CallType   uint   `json:"callType"   orm:"call_type"   description:"呼叫类型(1服务员,2结账)"`                 // 呼叫类型(1服务员,2结账)
	Status     int    `json:"status"     orm:"status"      description:"状态,0-unhandled未处理 1-handled已处理"` // 状态,0-unhandled未处理 1-handled已处理
	IsSend     int    `json:"isSend"     orm:"is_send"     description:"消息发送状态 0-否 1-是"`                 // 消息发送状态 0-否 1-是
	CreateTime uint   `json:"createTime" orm:"create_time" description:"创建时间(时间戳)"`                      // 创建时间(时间戳)
	UpdateTime uint   `json:"updateTime" orm:"update_time" description:"更新时间(时间戳)"`                      // 更新时间(时间戳)
	DeleteTime uint   `json:"deleteTime" orm:"delete_time" description:"删除时间(时间戳)"`                      // 删除时间(时间戳)
}

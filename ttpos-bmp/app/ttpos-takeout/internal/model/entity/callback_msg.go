// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// CallbackMsg is the golang structure for table callback_msg.
type CallbackMsg struct {
	Id             int64       `json:"id"             orm:"id"              description:"主键"`                     // 主键
	CreatedAt      *gtime.Time `json:"createdAt"      orm:"created_at"      description:"创建时间"`                   // 创建时间
	UpdatedAt      *gtime.Time `json:"updatedAt"      orm:"updated_at"      description:"修改时间"`                   // 修改时间
	DeletedAt      *gtime.Time `json:"deletedAt"      orm:"deleted_at"      description:"软删除"`                    // 软删除
	Uuid           string      `json:"uuid"           orm:"uuid"            description:"全局唯一ID"`                 // 全局唯一ID
	TakeoutRefNo   string      `json:"takeoutRefNo"   orm:"takeout_ref_no"  description:"外送系统订单号，如skootar.jobId"` // 外送系统订单号，如skootar.jobId
	Content        string      `json:"content"        orm:"content"         description:"消息内容"`                   // 消息内容
	StatusDatetime *gtime.Time `json:"statusDatetime" orm:"status_datetime" description:"状态变更时间"`                 // 状态变更时间
}

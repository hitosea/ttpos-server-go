// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MemberLevelLog is the golang structure for table member_level_log.
type MemberLevelLog struct {
	Id         uint   `json:"id"         orm:"id"           description:"自增ID"`                   // 自增ID
	Uuid       uint64 `json:"uuid"       orm:"uuid"         description:"日志ID"`                   // 日志ID
	MemberUuid uint64 `json:"memberUuid" orm:"member_uuid"  description:"会员ID"`                   // 会员ID
	OldLevelId uint64 `json:"oldLevelId" orm:"old_level_id" description:"变更前的等级id"`               // 变更前的等级id
	NewLevelId uint64 `json:"newLevelId" orm:"new_level_id" description:"变更后的等级id"`               // 变更后的等级id
	ChangeType uint   `json:"changeType" orm:"change_type"  description:"变更类型(10后台管理员设置 20自动升级)"` // 变更类型(10后台管理员设置 20自动升级)
	Remark     string `json:"remark"     orm:"remark"       description:"管理员备注"`                  // 管理员备注
	CreateTime uint   `json:"createTime" orm:"create_time"  description:"创建时间(时间戳)"`              // 创建时间(时间戳)
	UpdateTime uint   `json:"updateTime" orm:"update_time"  description:"更新时间(时间戳)"`              // 更新时间(时间戳)
	DeleteTime uint   `json:"deleteTime" orm:"delete_time"  description:"删除时间(时间戳)"`              // 删除时间(时间戳)
}

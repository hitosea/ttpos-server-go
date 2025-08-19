// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MemberLevelLog is the golang structure of table ttpos_member_level_log for DAO operations like Where/Data.
type MemberLevelLog struct {
	g.Meta     `orm:"table:ttpos_member_level_log, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // 日志ID
	MemberUuid interface{} // 会员ID
	OldLevelId interface{} // 变更前的等级id
	NewLevelId interface{} // 变更后的等级id
	ChangeType interface{} // 变更类型(10后台管理员设置 20自动升级)
	Remark     interface{} // 管理员备注
	CreateTime interface{} // 创建时间(时间戳)
	UpdateTime interface{} // 更新时间(时间戳)
	DeleteTime interface{} // 删除时间(时间戳)
}

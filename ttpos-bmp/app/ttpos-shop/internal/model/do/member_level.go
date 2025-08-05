// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MemberLevel is the golang structure of table ttpos_member_level for DAO operations like Where/Data.
type MemberLevel struct {
	g.Meta       `orm:"table:ttpos_member_level, do:true"`
	Id           interface{} // 自增ID
	Uuid         interface{} // 会员等级ID
	Name         interface{} // 等级名称
	OpenMoney    interface{} // 是否开放累计消费额升级，0-否 1-是
	UpgradeMoney interface{} // 升级条件，累计消费额
	OpenPoint    interface{} // 是否开放累计积分升级，0-否 1-是
	UpgradePoint interface{} // 升级条件，累计积分
	Discount     interface{} // 等级权益,百分比折扣,单位%, 如80%为打8折，discount值为0.8
	Priority     interface{} // 等级权重，越大等级越高
	IsDefault    interface{} // 是否默认, 1-是 0-否
	Remark       interface{} // 备注
	CreateTime   interface{} // 创建时间(时间戳)
	UpdateTime   interface{} // 更新时间(时间戳)
	DeleteTime   interface{} // 删除时间(时间戳)
}

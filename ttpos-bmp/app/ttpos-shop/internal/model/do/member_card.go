// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MemberCard is the golang structure of table ttpos_member_card for DAO operations like Where/Data.
type MemberCard struct {
	g.Meta       `orm:"table:ttpos_member_card, do:true"`
	Id           interface{} // 自增ID
	Uuid         interface{} // 会员卡ID
	CardTypeUuid interface{} // 会员卡类型ID
	MemberUuid   interface{} // 会员ID
	ExpireTime   interface{} // 截止日期(时间戳)
	Discount     interface{} // 折扣,单位%, 如80%为打8折，discount值为0.8 .不随后台改变,按领取时的折扣。后续会员卡类型折扣改变时,不改变此字段
	CreateTime   interface{} // 创建时间(时间戳),领取时间
	UpdateTime   interface{} // 更新时间(时间戳)
	DeleteTime   interface{} // 删除时间(时间戳)
}

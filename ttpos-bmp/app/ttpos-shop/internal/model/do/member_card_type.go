// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MemberCardType is the golang structure of table ttpos_member_card_type for DAO operations like Where/Data.
type MemberCardType struct {
	g.Meta       `orm:"table:ttpos_member_card_type, do:true"`
	Id           interface{} // 自增ID
	Uuid         interface{} // 会员卡类型ID
	Name         interface{} // 会员卡类型名称
	Expire       interface{} // 有效期限,单位:月, 0为永久有效
	Price        interface{} // 价格
	Discount     interface{} // 折扣,单位%
	Sort         interface{} // 排序
	Status       interface{} // 状态, 0-开启 1-关闭
	OpenPoint    interface{} // 开卡赠送积分,0-否 1-是
	OpenPointNum interface{} // 开卡赠送积分数
	OpenMoney    interface{} // 开卡赠送余额,0-否 1-是
	OpenMoneyNum interface{} // 开卡赠送余额数
	Describe     interface{} // 使用须知
	CreateTime   interface{} // 创建时间(时间戳)
	UpdateTime   interface{} // 更新时间(时间戳)
	DeleteTime   interface{} // 删除时间(时间戳)
}

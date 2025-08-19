// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MemberCardLog is the golang structure of table ttpos_member_card_log for DAO operations like Where/Data.
type MemberCardLog struct {
	g.Meta             `orm:"table:ttpos_member_card_log, do:true"`
	Id                 interface{} // 自增ID
	Uuid               interface{} // 会员卡领取记录ID
	Price              interface{} // 价格,会员卡价格,不随后台改变,记录领取时的价格
	Discount           interface{} // 折扣,单位%,不随后台改变,记录领取时的折扣
	Expire             interface{} // 有效期限,单位:月, 0为永久有效,不随后台改变,记录领取时的有效期限
	MemberName         interface{} // 会员名称,不随后台改变,当无法用member_uuid获取会员信息时,用此字段
	MemberPhone        interface{} // 会员电话,不随后台改变,当无法用member_uuid获取会员信息时,用此字段
	MemberNo           interface{} // 会员编号,不随后台改变,当无法用member_uuid获取会员信息时,用此字段
	MemberCardTypeName interface{} // 会员卡类型名称,不随后台改变,当无法用member_card_type_uuid获取会员卡类型信息时,用此字段
	MemberCardTypeUuid interface{} // 会员卡类型ID
	MemberUuid         interface{} // 会员ID
	CreateTime         interface{} // 创建时间(时间戳)
	UpdateTime         interface{} // 更新时间(时间戳)
	DeleteTime         interface{} // 删除时间(时间戳)
}

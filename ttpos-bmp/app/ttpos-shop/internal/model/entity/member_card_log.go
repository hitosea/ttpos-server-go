// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MemberCardLog is the golang structure for table member_card_log.
type MemberCardLog struct {
	Id                 uint    `json:"id"                 orm:"id"                    description:"自增ID"`                                                    // 自增ID
	Uuid               uint64  `json:"uuid"               orm:"uuid"                  description:"会员卡领取记录ID"`                                               // 会员卡领取记录ID
	Price              float64 `json:"price"              orm:"price"                 description:"价格,会员卡价格,不随后台改变,记录领取时的价格"`                                // 价格,会员卡价格,不随后台改变,记录领取时的价格
	Discount           float64 `json:"discount"           orm:"discount"              description:"折扣,单位%,不随后台改变,记录领取时的折扣"`                                  // 折扣,单位%,不随后台改变,记录领取时的折扣
	Expire             int     `json:"expire"             orm:"expire"                description:"有效期限,单位:月, 0为永久有效,不随后台改变,记录领取时的有效期限"`                     // 有效期限,单位:月, 0为永久有效,不随后台改变,记录领取时的有效期限
	MemberName         string  `json:"memberName"         orm:"member_name"           description:"会员名称,不随后台改变,当无法用member_uuid获取会员信息时,用此字段"`                 // 会员名称,不随后台改变,当无法用member_uuid获取会员信息时,用此字段
	MemberPhone        string  `json:"memberPhone"        orm:"member_phone"          description:"会员电话,不随后台改变,当无法用member_uuid获取会员信息时,用此字段"`                 // 会员电话,不随后台改变,当无法用member_uuid获取会员信息时,用此字段
	MemberNo           string  `json:"memberNo"           orm:"member_no"             description:"会员编号,不随后台改变,当无法用member_uuid获取会员信息时,用此字段"`                 // 会员编号,不随后台改变,当无法用member_uuid获取会员信息时,用此字段
	MemberCardTypeName string  `json:"memberCardTypeName" orm:"member_card_type_name" description:"会员卡类型名称,不随后台改变,当无法用member_card_type_uuid获取会员卡类型信息时,用此字段"` // 会员卡类型名称,不随后台改变,当无法用member_card_type_uuid获取会员卡类型信息时,用此字段
	MemberCardTypeUuid uint64  `json:"memberCardTypeUuid" orm:"member_card_type_uuid" description:"会员卡类型ID"`                                                 // 会员卡类型ID
	MemberUuid         uint64  `json:"memberUuid"         orm:"member_uuid"           description:"会员ID"`                                                    // 会员ID
	CreateTime         uint    `json:"createTime"         orm:"create_time"           description:"创建时间(时间戳)"`                                               // 创建时间(时间戳)
	UpdateTime         uint    `json:"updateTime"         orm:"update_time"           description:"更新时间(时间戳)"`                                               // 更新时间(时间戳)
	DeleteTime         uint    `json:"deleteTime"         orm:"delete_time"           description:"删除时间(时间戳)"`                                               // 删除时间(时间戳)
}

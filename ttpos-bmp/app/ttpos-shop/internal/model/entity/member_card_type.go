// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MemberCardType is the golang structure for table member_card_type.
type MemberCardType struct {
	Id           uint    `json:"id"           orm:"id"             description:"自增ID"`              // 自增ID
	Uuid         uint64  `json:"uuid"         orm:"uuid"           description:"会员卡类型ID"`           // 会员卡类型ID
	Name         string  `json:"name"         orm:"name"           description:"会员卡类型名称"`           // 会员卡类型名称
	Expire       int     `json:"expire"       orm:"expire"         description:"有效期限,单位:月, 0为永久有效"` // 有效期限,单位:月, 0为永久有效
	Price        float64 `json:"price"        orm:"price"          description:"价格"`                // 价格
	Discount     float64 `json:"discount"     orm:"discount"       description:"折扣,单位%"`            // 折扣,单位%
	Sort         int     `json:"sort"         orm:"sort"           description:"排序"`                // 排序
	Status       int     `json:"status"       orm:"status"         description:"状态, 0-开启 1-关闭"`     // 状态, 0-开启 1-关闭
	OpenPoint    int     `json:"openPoint"    orm:"open_point"     description:"开卡赠送积分,0-否 1-是"`    // 开卡赠送积分,0-否 1-是
	OpenPointNum float64 `json:"openPointNum" orm:"open_point_num" description:"开卡赠送积分数"`           // 开卡赠送积分数
	OpenMoney    int     `json:"openMoney"    orm:"open_money"     description:"开卡赠送余额,0-否 1-是"`    // 开卡赠送余额,0-否 1-是
	OpenMoneyNum float64 `json:"openMoneyNum" orm:"open_money_num" description:"开卡赠送余额数"`           // 开卡赠送余额数
	Describe     string  `json:"describe"     orm:"describe"       description:"使用须知"`              // 使用须知
	CreateTime   uint    `json:"createTime"   orm:"create_time"    description:"创建时间(时间戳)"`         // 创建时间(时间戳)
	UpdateTime   uint    `json:"updateTime"   orm:"update_time"    description:"更新时间(时间戳)"`         // 更新时间(时间戳)
	DeleteTime   uint    `json:"deleteTime"   orm:"delete_time"    description:"删除时间(时间戳)"`         // 删除时间(时间戳)
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MemberCard is the golang structure for table member_card.
type MemberCard struct {
	Id           uint    `json:"id"           orm:"id"             description:"自增ID"`                                                               // 自增ID
	Uuid         uint64  `json:"uuid"         orm:"uuid"           description:"会员卡ID"`                                                              // 会员卡ID
	CardTypeUuid uint64  `json:"cardTypeUuid" orm:"card_type_uuid" description:"会员卡类型ID"`                                                            // 会员卡类型ID
	MemberUuid   uint64  `json:"memberUuid"   orm:"member_uuid"    description:"会员ID"`                                                               // 会员ID
	ExpireTime   int     `json:"expireTime"   orm:"expire_time"    description:"截止日期(时间戳)"`                                                          // 截止日期(时间戳)
	Discount     float64 `json:"discount"     orm:"discount"       description:"折扣,单位%, 如80%为打8折，discount值为0.8 .不随后台改变,按领取时的折扣。后续会员卡类型折扣改变时,不改变此字段"` // 折扣,单位%, 如80%为打8折，discount值为0.8 .不随后台改变,按领取时的折扣。后续会员卡类型折扣改变时,不改变此字段
	CreateTime   uint    `json:"createTime"   orm:"create_time"    description:"创建时间(时间戳),领取时间"`                                                     // 创建时间(时间戳),领取时间
	UpdateTime   uint    `json:"updateTime"   orm:"update_time"    description:"更新时间(时间戳)"`                                                          // 更新时间(时间戳)
	DeleteTime   uint    `json:"deleteTime"   orm:"delete_time"    description:"删除时间(时间戳)"`                                                          // 删除时间(时间戳)
}

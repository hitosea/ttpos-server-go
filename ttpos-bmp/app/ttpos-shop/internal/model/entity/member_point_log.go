// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MemberPointLog is the golang structure for table member_point_log.
type MemberPointLog struct {
	Id          uint    `json:"id"          orm:"id"           description:"自增ID"`                                                                // 自增ID
	Uuid        uint64  `json:"uuid"        orm:"uuid"         description:"积分变动记录ID"`                                                            // 积分变动记录ID
	MemberUuid  uint64  `json:"memberUuid"  orm:"member_uuid"  description:"会员ID"`                                                                // 会员ID
	Scene       int     `json:"scene"       orm:"scene"        description:"场景,10-用户充值 20-订单赠送 30-管理员操作 40-退款扣除 60-订单反结账 70-充值赠送 80-充值反结账 90-扣减"` // 场景,10-用户充值 20-订单赠送 30-管理员操作 40-退款扣除 60-订单反结账 70-充值赠送 80-充值反结账 90-扣减
	Value       float64 `json:"value"       orm:"value"        description:"数值,负数:减积分 正数:加积分"`                                                    // 数值,负数:减积分 正数:加积分
	Describe    string  `json:"describe"    orm:"describe"     description:"变动描述"`                                                                // 变动描述
	RelatedUuid uint64  `json:"relatedUuid" orm:"related_uuid" description:"关联uuid. 表示积分变动记录关联的业务订单ID,可能是销售订单、充值订单、退款单、退货单退款金额"`                  // 关联uuid. 表示积分变动记录关联的业务订单ID,可能是销售订单、充值订单、退款单、退货单退款金额
	Processed   int     `json:"processed"   orm:"processed"    description:"是否已处理,0-未处理 1-已处理. 用于处理积分变动，修改会员的积分并清0冻结的积分"`                         // 是否已处理,0-未处理 1-已处理. 用于处理积分变动，修改会员的积分并清0冻结的积分
	CreateTime  uint    `json:"createTime"  orm:"create_time"  description:"创建时间(时间戳)"`                                                           // 创建时间(时间戳)
	UpdateTime  uint    `json:"updateTime"  orm:"update_time"  description:"更新时间(时间戳)"`                                                           // 更新时间(时间戳)
	DeleteTime  uint    `json:"deleteTime"  orm:"delete_time"  description:"删除时间(时间戳)"`                                                           // 删除时间(时间戳)
}

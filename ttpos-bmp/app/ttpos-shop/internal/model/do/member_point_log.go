// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MemberPointLog is the golang structure of table ttpos_member_point_log for DAO operations like Where/Data.
type MemberPointLog struct {
	g.Meta      `orm:"table:ttpos_member_point_log, do:true"`
	Id          interface{} // 自增ID
	Uuid        interface{} // 积分变动记录ID
	MemberUuid  interface{} // 会员ID
	Scene       interface{} // 场景,10-用户充值 20-订单赠送 30-管理员操作 40-退款扣除 60-订单反结账 70-充值赠送 80-充值反结账 90-扣减
	Value       interface{} // 数值,负数:减积分 正数:加积分
	Describe    interface{} // 变动描述
	RelatedUuid interface{} // 关联uuid. 表示积分变动记录关联的业务订单ID,可能是销售订单、充值订单、退款单、退货单退款金额
	Processed   interface{} // 是否已处理,0-未处理 1-已处理. 用于处理积分变动，修改会员的积分并清0冻结的积分
	CreateTime  interface{} // 创建时间(时间戳)
	UpdateTime  interface{} // 更新时间(时间戳)
	DeleteTime  interface{} // 删除时间(时间戳)
}

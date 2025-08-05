// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Member is the golang structure of table ttpos_member for DAO operations like Where/Data.
type Member struct {
	g.Meta                       `orm:"table:ttpos_member, do:true"`
	Id                           interface{} // 自增ID
	Uuid                         interface{} // 会员ID
	MemberNo                     interface{} // 会员编号
	Nickname                     interface{} // 昵称
	Gender                       interface{} // 性别,0-女 1-男 2-未知
	Phone                        interface{} // 电话号码
	Password                     interface{} // 密码
	Birthday                     interface{} // 生日,时间戳
	Point                        interface{} // 积分
	FrozenPoint                  interface{} // 冻结积分。冻结积分不能使用，在前端显示为已扣除或已增加。冻结积分可为负数。积分余额=积分+冻结积分
	AccumulatedConsumptionAmount interface{} // 累计消费金额
	ConsumptionCount             interface{} // 消费次数
	Balance                      interface{} // 余额
	FrozenBalance                interface{} // 冻结余额。冻结余额不能使用，在前端显示为已扣除或已增加。冻结余额可为负数。会员余额=余额+冻结余额
	GiftBalance                  interface{} // 赠送账户余额
	FrozenGiftBalance            interface{} // 冻结赠送账户余额。冻结赠送账户余额不能使用，在前端显示为已扣除或已增加。冻结赠送账户余额可为负数。赠送账户余额=赠送账户余额+冻结赠送账户余额
	AccumulatedRechargeAmount    interface{} // 累计充值金额
	MemberLevelUuid              interface{} // 会员等级ID
	MemberCardUuid               interface{} // 会员卡片ID
	CreateTime                   interface{} // 创建时间(时间戳)
	UpdateTime                   interface{} // 更新时间(时间戳)
	DeleteTime                   interface{} // 删除时间(时间戳)
}

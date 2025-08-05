// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MemberBalanceLog is the golang structure of table ttpos_member_balance_log for DAO operations like Where/Data.
type MemberBalanceLog struct {
	g.Meta        `orm:"table:ttpos_member_balance_log, do:true"`
	Id            interface{} // 自增ID
	Uuid          interface{} // 余额变动记录ID
	MemberUuid    interface{} // 会员ID
	Scene         interface{} // 场景,10-用户充值 20-用户消费 30-管理员操作 40-订单退款 50-余额提现 60-订单反结账 70-充值反结账 80-充值退款 90-销售订单支付扣减
	SaleOrderUuid interface{} // 销售订单ID
	Money         interface{} // 变动金额,负数:减余额 整数:加余额
	GiftMoney     interface{} // 变动赠送金额
	Describe      interface{} // 变动描述
	Processed     interface{} // 是否已处理,0-未处理 1-已处理. 用于处理会员余额变动，修改会员的余额并清0冻结的余额
	RelatedUuid   interface{} // 关联uuid. 表示余额变动记录关联的业务订单ID,可能是销售订单(场景90)、充值订单(场景10)、退款单(场景80)、退货单退款金额(场景40)
	CreateTime    interface{} // 创建时间(时间戳)
	UpdateTime    interface{} // 更新时间(时间戳)
	DeleteTime    interface{} // 删除时间(时间戳)
}

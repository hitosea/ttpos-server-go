// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MemberBalanceLog is the golang structure for table member_balance_log.
type MemberBalanceLog struct {
	Id            uint    `json:"id"            orm:"id"              description:"自增ID"`                                                                              // 自增ID
	Uuid          uint64  `json:"uuid"          orm:"uuid"            description:"余额变动记录ID"`                                                                          // 余额变动记录ID
	MemberUuid    uint64  `json:"memberUuid"    orm:"member_uuid"     description:"会员ID"`                                                                              // 会员ID
	Scene         int     `json:"scene"         orm:"scene"           description:"场景,10-用户充值 20-用户消费 30-管理员操作 40-订单退款 50-余额提现 60-订单反结账 70-充值反结账 80-充值退款 90-销售订单支付扣减"` // 场景,10-用户充值 20-用户消费 30-管理员操作 40-订单退款 50-余额提现 60-订单反结账 70-充值反结账 80-充值退款 90-销售订单支付扣减
	SaleOrderUuid uint64  `json:"saleOrderUuid" orm:"sale_order_uuid" description:"销售订单ID"`                                                                            // 销售订单ID
	Money         float64 `json:"money"         orm:"money"           description:"变动金额,负数:减余额 整数:加余额"`                                                                // 变动金额,负数:减余额 整数:加余额
	GiftMoney     float64 `json:"giftMoney"     orm:"gift_money"      description:"变动赠送金额"`                                                                            // 变动赠送金额
	Describe      string  `json:"describe"      orm:"describe"        description:"变动描述"`                                                                              // 变动描述
	Processed     int     `json:"processed"     orm:"processed"       description:"是否已处理,0-未处理 1-已处理. 用于处理会员余额变动，修改会员的余额并清0冻结的余额"`                                     // 是否已处理,0-未处理 1-已处理. 用于处理会员余额变动，修改会员的余额并清0冻结的余额
	RelatedUuid   uint64  `json:"relatedUuid"   orm:"related_uuid"    description:"关联uuid. 表示余额变动记录关联的业务订单ID,可能是销售订单(场景90)、充值订单(场景10)、退款单(场景80)、退货单退款金额(场景40)"`        // 关联uuid. 表示余额变动记录关联的业务订单ID,可能是销售订单(场景90)、充值订单(场景10)、退款单(场景80)、退货单退款金额(场景40)
	CreateTime    uint    `json:"createTime"    orm:"create_time"     description:"创建时间(时间戳)"`                                                                         // 创建时间(时间戳)
	UpdateTime    uint    `json:"updateTime"    orm:"update_time"     description:"更新时间(时间戳)"`                                                                         // 更新时间(时间戳)
	DeleteTime    uint    `json:"deleteTime"    orm:"delete_time"     description:"删除时间(时间戳)"`                                                                         // 删除时间(时间戳)
}

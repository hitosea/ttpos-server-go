// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Member is the golang structure for table member.
type Member struct {
	Id                           uint    `json:"id"                           orm:"id"                             description:"自增ID"`                                                                    // 自增ID
	Uuid                         uint64  `json:"uuid"                         orm:"uuid"                           description:"会员ID"`                                                                    // 会员ID
	MemberNo                     string  `json:"memberNo"                     orm:"member_no"                      description:"会员编号"`                                                                    // 会员编号
	Nickname                     string  `json:"nickname"                     orm:"nickname"                       description:"昵称"`                                                                      // 昵称
	Gender                       int     `json:"gender"                       orm:"gender"                         description:"性别,0-女 1-男 2-未知"`                                                         // 性别,0-女 1-男 2-未知
	Phone                        string  `json:"phone"                        orm:"phone"                          description:"电话号码"`                                                                    // 电话号码
	Password                     string  `json:"password"                     orm:"password"                       description:"密码"`                                                                      // 密码
	Birthday                     int     `json:"birthday"                     orm:"birthday"                       description:"生日,时间戳"`                                                                  // 生日,时间戳
	Point                        float64 `json:"point"                        orm:"point"                          description:"积分"`                                                                      // 积分
	FrozenPoint                  float64 `json:"frozenPoint"                  orm:"frozen_point"                   description:"冻结积分。冻结积分不能使用，在前端显示为已扣除或已增加。冻结积分可为负数。积分余额=积分+冻结积分"`                       // 冻结积分。冻结积分不能使用，在前端显示为已扣除或已增加。冻结积分可为负数。积分余额=积分+冻结积分
	AccumulatedConsumptionAmount float64 `json:"accumulatedConsumptionAmount" orm:"accumulated_consumption_amount" description:"累计消费金额"`                                                                  // 累计消费金额
	ConsumptionCount             int     `json:"consumptionCount"             orm:"consumption_count"              description:"消费次数"`                                                                    // 消费次数
	Balance                      float64 `json:"balance"                      orm:"balance"                        description:"余额"`                                                                      // 余额
	FrozenBalance                float64 `json:"frozenBalance"                orm:"frozen_balance"                 description:"冻结余额。冻结余额不能使用，在前端显示为已扣除或已增加。冻结余额可为负数。会员余额=余额+冻结余额"`                       // 冻结余额。冻结余额不能使用，在前端显示为已扣除或已增加。冻结余额可为负数。会员余额=余额+冻结余额
	GiftBalance                  float64 `json:"giftBalance"                  orm:"gift_balance"                   description:"赠送账户余额"`                                                                  // 赠送账户余额
	FrozenGiftBalance            float64 `json:"frozenGiftBalance"            orm:"frozen_gift_balance"            description:"冻结赠送账户余额。冻结赠送账户余额不能使用，在前端显示为已扣除或已增加。冻结赠送账户余额可为负数。赠送账户余额=赠送账户余额+冻结赠送账户余额"` // 冻结赠送账户余额。冻结赠送账户余额不能使用，在前端显示为已扣除或已增加。冻结赠送账户余额可为负数。赠送账户余额=赠送账户余额+冻结赠送账户余额
	AccumulatedRechargeAmount    float64 `json:"accumulatedRechargeAmount"    orm:"accumulated_recharge_amount"    description:"累计充值金额"`                                                                  // 累计充值金额
	MemberLevelUuid              uint64  `json:"memberLevelUuid"              orm:"member_level_uuid"              description:"会员等级ID"`                                                                  // 会员等级ID
	MemberCardUuid               uint64  `json:"memberCardUuid"               orm:"member_card_uuid"               description:"会员卡片ID"`                                                                  // 会员卡片ID
	CreateTime                   uint    `json:"createTime"                   orm:"create_time"                    description:"创建时间(时间戳)"`                                                               // 创建时间(时间戳)
	UpdateTime                   uint    `json:"updateTime"                   orm:"update_time"                    description:"更新时间(时间戳)"`                                                               // 更新时间(时间戳)
	DeleteTime                   uint    `json:"deleteTime"                   orm:"delete_time"                    description:"删除时间(时间戳)"`                                                               // 删除时间(时间戳)
}

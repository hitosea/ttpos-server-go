// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MarketingActivityPrize is the golang structure of table ttpos_marketing_activity_prize for DAO operations like Where/Data.
type MarketingActivityPrize struct {
	g.Meta       `orm:"table:ttpos_marketing_activity_prize, do:true"`
	Id           interface{} //
	Uuid         interface{} // 礼品唯一ID
	ActivityUuid interface{} // 活动uuid
	PrizeType    interface{} // 奖品类型
	PrizeUuid    interface{} // 奖品uuid
	CreateTime   interface{} // 创建时间
	UpdateTime   interface{} // 更新时间
	DeleteTime   interface{} // 删除时间
}

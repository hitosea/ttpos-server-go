// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MarketingActivityConsumption is the golang structure of table ttpos_marketing_activity_consumption for DAO operations like Where/Data.
type MarketingActivityConsumption struct {
	g.Meta            `orm:"table:ttpos_marketing_activity_consumption, do:true"`
	Id                interface{} //
	Uuid              interface{} // 消费记录唯一ID
	ActivityUuid      interface{} // 活动uuid
	ReferrerUuid      interface{} // 推荐人uuid
	ConsumerUuid      interface{} // 消费人uuid
	ConsumptionAmount interface{} // 消费金额
	RewardAmount      interface{} // 奖励金额
	RewardStatus      interface{} // 奖励状态 0未发放 1已发放
	CreateTime        interface{} // 创建时间
	UpdateTime        interface{} // 更新时间
	DeleteTime        interface{} // 删除时间
}

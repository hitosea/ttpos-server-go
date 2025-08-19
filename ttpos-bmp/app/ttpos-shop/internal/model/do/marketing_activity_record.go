// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MarketingActivityRecord is the golang structure of table ttpos_marketing_activity_record for DAO operations like Where/Data.
type MarketingActivityRecord struct {
	g.Meta         `orm:"table:ttpos_marketing_activity_record, do:true"`
	Id             interface{} //
	Uuid           interface{} // 记录唯一ID
	ActivityUuid   interface{} // 活动uuid
	PrizeUuid      interface{} // 奖品uuid
	MemberUuid     interface{} // 会员uuid
	RewardCount    interface{} // 已获得奖励次数
	LastRewardTime interface{} // 最后一次获得奖励时间
	CreateTime     interface{} // 创建时间
	UpdateTime     interface{} // 更新时间
	DeleteTime     interface{} // 删除时间
}

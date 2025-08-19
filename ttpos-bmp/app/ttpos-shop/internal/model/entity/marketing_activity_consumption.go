// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MarketingActivityConsumption is the golang structure for table marketing_activity_consumption.
type MarketingActivityConsumption struct {
	Id                uint    `json:"id"                orm:"id"                 description:""`               //
	Uuid              int64   `json:"uuid"              orm:"uuid"               description:"消费记录唯一ID"`       // 消费记录唯一ID
	ActivityUuid      int64   `json:"activityUuid"      orm:"activity_uuid"      description:"活动uuid"`         // 活动uuid
	ReferrerUuid      int64   `json:"referrerUuid"      orm:"referrer_uuid"      description:"推荐人uuid"`        // 推荐人uuid
	ConsumerUuid      int64   `json:"consumerUuid"      orm:"consumer_uuid"      description:"消费人uuid"`        // 消费人uuid
	ConsumptionAmount float64 `json:"consumptionAmount" orm:"consumption_amount" description:"消费金额"`           // 消费金额
	RewardAmount      float64 `json:"rewardAmount"      orm:"reward_amount"      description:"奖励金额"`           // 奖励金额
	RewardStatus      int     `json:"rewardStatus"      orm:"reward_status"      description:"奖励状态 0未发放 1已发放"` // 奖励状态 0未发放 1已发放
	CreateTime        int     `json:"createTime"        orm:"create_time"        description:"创建时间"`           // 创建时间
	UpdateTime        int     `json:"updateTime"        orm:"update_time"        description:"更新时间"`           // 更新时间
	DeleteTime        int     `json:"deleteTime"        orm:"delete_time"        description:"删除时间"`           // 删除时间
}

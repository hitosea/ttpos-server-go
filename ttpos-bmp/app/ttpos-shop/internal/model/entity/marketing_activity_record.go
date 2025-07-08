// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MarketingActivityRecord is the golang structure for table marketing_activity_record.
type MarketingActivityRecord struct {
	Id             uint  `json:"id"             orm:"id"               description:""`           //
	Uuid           int64 `json:"uuid"           orm:"uuid"             description:"记录唯一ID"`     // 记录唯一ID
	ActivityUuid   int64 `json:"activityUuid"   orm:"activity_uuid"    description:"活动uuid"`     // 活动uuid
	PrizeUuid      int64 `json:"prizeUuid"      orm:"prize_uuid"       description:"奖品uuid"`     // 奖品uuid
	MemberUuid     int64 `json:"memberUuid"     orm:"member_uuid"      description:"会员uuid"`     // 会员uuid
	RewardCount    int   `json:"rewardCount"    orm:"reward_count"     description:"已获得奖励次数"`    // 已获得奖励次数
	LastRewardTime int   `json:"lastRewardTime" orm:"last_reward_time" description:"最后一次获得奖励时间"` // 最后一次获得奖励时间
	CreateTime     int   `json:"createTime"     orm:"create_time"      description:"创建时间"`       // 创建时间
	UpdateTime     int   `json:"updateTime"     orm:"update_time"      description:"更新时间"`       // 更新时间
	DeleteTime     int   `json:"deleteTime"     orm:"delete_time"      description:"删除时间"`       // 删除时间
}

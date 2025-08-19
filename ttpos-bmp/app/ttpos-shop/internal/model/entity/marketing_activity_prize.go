// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MarketingActivityPrize is the golang structure for table marketing_activity_prize.
type MarketingActivityPrize struct {
	Id           uint  `json:"id"           orm:"id"            description:""`       //
	Uuid         int64 `json:"uuid"         orm:"uuid"          description:"礼品唯一ID"` // 礼品唯一ID
	ActivityUuid int64 `json:"activityUuid" orm:"activity_uuid" description:"活动uuid"` // 活动uuid
	PrizeType    int   `json:"prizeType"    orm:"prize_type"    description:"奖品类型"`   // 奖品类型
	PrizeUuid    int64 `json:"prizeUuid"    orm:"prize_uuid"    description:"奖品uuid"` // 奖品uuid
	CreateTime   int   `json:"createTime"   orm:"create_time"   description:"创建时间"`   // 创建时间
	UpdateTime   int   `json:"updateTime"   orm:"update_time"   description:"更新时间"`   // 更新时间
	DeleteTime   int   `json:"deleteTime"   orm:"delete_time"   description:"删除时间"`   // 删除时间
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MarketingActivity is the golang structure for table marketing_activity.
type MarketingActivity struct {
	Id                    uint    `json:"id"                    orm:"id"                       description:""`                 //
	Uuid                  int64   `json:"uuid"                  orm:"uuid"                     description:"活动唯一ID"`           // 活动唯一ID
	Name                  string  `json:"name"                  orm:"name"                     description:"活动名称"`             // 活动名称
	Type                  int     `json:"type"                  orm:"type"                     description:"活动类型 0邀请有礼 1积分商城"` // 活动类型 0邀请有礼 1积分商城
	MultiLanguageNameUuid int64   `json:"multiLanguageNameUuid" orm:"multi_language_name_uuid" description:"活动名称多语言uuid"`      // 活动名称多语言uuid
	Description           string  `json:"description"           orm:"description"              description:"活动描述"`             // 活动描述
	MultiLanguageDescUuid int64   `json:"multiLanguageDescUuid" orm:"multi_language_desc_uuid" description:"活动文案多语言uuid"`      // 活动文案多语言uuid
	StartTime             int     `json:"startTime"             orm:"start_time"               description:"活动开始时间"`           // 活动开始时间
	EndTime               int     `json:"endTime"               orm:"end_time"                 description:"活动结束时间"`           // 活动结束时间
	RewardConditionAmount float64 `json:"rewardConditionAmount" orm:"reward_condition_amount"  description:"奖励条件金额"`           // 奖励条件金额
	IsOpenRewardLimit     int     `json:"isOpenRewardLimit"     orm:"is_open_reward_limit"     description:"是否开启奖励次数限制 0否 1是"` // 是否开启奖励次数限制 0否 1是
	RewardLimit           int     `json:"rewardLimit"           orm:"reward_limit"             description:"奖励次数限制"`           // 奖励次数限制
	IsInvalid             int     `json:"isInvalid"             orm:"is_invalid"               description:"是否失效 0否 1是"`       // 是否失效 0否 1是
	ImageBase64           string  `json:"imageBase64"           orm:"image_base64"             description:"活动图片base64"`       // 活动图片base64
	CreateTime            int     `json:"createTime"            orm:"create_time"              description:"创建时间"`             // 创建时间
	UpdateTime            int     `json:"updateTime"            orm:"update_time"              description:"更新时间"`             // 更新时间
	DeleteTime            int     `json:"deleteTime"            orm:"delete_time"              description:"删除时间"`             // 删除时间
}

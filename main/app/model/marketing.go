package model

import "time"

// MarketingActivity 营销活动模型
type MarketingActivity struct {
	BaseModel
	Name                  string  `gorm:"column:name;type:varchar(2500);default:'';comment:活动名称" json:"name"`
	Type                  int     `gorm:"column:type;type:tinyint(1);default:0;comment:活动类型 0邀请有礼 1积分商城" json:"type"`
	MultiLanguageNameUuid uint64  `gorm:"column:multi_language_name_uuid;type:biginteger;default:0;comment:活动名称多语言uuid" json:"multi_language_name_uuid"`
	Description           string  `gorm:"column:description;type:varchar(5000);default:'';comment:活动文案" json:"description"`
	MultiLanguageDescUuid uint64  `gorm:"column:multi_language_desc_uuid;type:biginteger;default:0;comment:活动文案多语言uuid" json:"multi_language_desc_uuid"`
	StartTime             int     `gorm:"column:start_time;type:int;default:0;comment:活动开始时间" json:"start_time"`
	EndTime               int     `gorm:"column:end_time;type:int;default:0;comment:活动结束时间" json:"end_time"`
	RewardConditionAmount float64 `gorm:"column:reward_condition_amount;type:decimal(14,2);default:0;comment:奖励条件金额" json:"reward_condition_amount"`
	IsOpenRewardLimit     int     `gorm:"column:is_open_reward_limit;type:tinyint(1);default:0;comment:是否开启奖励次数限制 0否 1是" json:"is_open_reward_limit"`
	RewardLimit           int64   `gorm:"column:reward_limit;type:int;default:0;comment:奖励次数限制" json:"reward_limit"`
	IsInvalid             int     `gorm:"column:is_invalid;type:tinyint(1);default:0;comment:是否失效 0否 1是" json:"is_invalid"`
	ImageBase64           string  `gorm:"column:image_base64;type:text;comment:活动图片base64" json:"image_base64"`

	Prizes            []*MarketingActivityPrize  `gorm:"foreignKey:ActivityUuid;references:Uuid" json:"prizes"`
	Records           []*MarketingActivityRecord `gorm:"foreignKey:ActivityUuid;references:Uuid" json:"records"`
	MultiLanguageName *MultiLanguageName         `gorm:"foreignKey:Uuid;references:MultiLanguageNameUuid" json:"multi_language_name"`
	MultiLanguageDesc *MultiLanguageName         `gorm:"foreignKey:Uuid;references:MultiLanguageDescUuid" json:"multi_language_desc"`
}

// 是否正在进行中
func (m *MarketingActivity) IsValid() bool {
	return m.IsInvalid == 0 && time.Now().Unix() > int64(m.StartTime) && time.Now().Unix() < int64(m.EndTime)
}

// MarketingActivityPrize 营销活动奖品模型
type MarketingActivityPrize struct {
	Uuid         uint64 `gorm:"column:uuid;type:biginteger;default:0;comment:礼品唯一ID" json:"uuid"`
	ActivityUuid uint64 `gorm:"column:activity_uuid;type:biginteger;default:0;comment:活动uuid" json:"activity_uuid"`
	PrizeType    int    `gorm:"column:prize_type;type:tinyint(1);default:0;comment:奖品类型 1优惠券 2未知" json:"prize_type"`
	PrizeUuid    uint64 `gorm:"column:prize_uuid;type:biginteger;default:0;comment:奖品uuid" json:"prize_uuid"`
	CreateTime   int    `gorm:"column:create_time;type:int;default:0;comment:创建时间" json:"create_time"`
	UpdateTime   int    `gorm:"column:update_time;type:int;default:0;comment:更新时间" json:"update_time"`
	DeleteTime   int    `gorm:"column:delete_time;type:int;default:0;comment:删除时间" json:"delete_time"`
}

// MarketingActivityRecord 营销活动记录模型
type MarketingActivityRecord struct {
	BaseModel
	ActivityUuid   uint64 `gorm:"column:activity_uuid;type:biginteger;default:0;comment:活动uuid" json:"activity_uuid"`
	PrizeUuid      uint64 `gorm:"column:prize_uuid;type:biginteger;default:0;comment:奖品uuid" json:"prize_uuid"`
	MemberUuid     uint64 `gorm:"column:member_uuid;type:biginteger;default:0;comment:会员uuid" json:"member_uuid"`
	RewardCount    int    `gorm:"column:reward_count;type:int;default:0;comment:已获得奖励次数" json:"reward_count"`
	LastRewardTime int64  `gorm:"column:last_reward_time;type:int;default:0;comment:最后一次获得奖励时间" json:"last_reward_time"`
}

// 营销活动消费记录表
type MarketingActivityConsumption struct {
	BaseModel
	ActivityUuid      uint64  `gorm:"column:activity_uuid;type:biginteger;default:0;comment:活动uuid" json:"activity_uuid"`
	ReferrerUuid      uint64  `gorm:"column:referrer_uuid;type:biginteger;default:0;comment:推荐人uuid" json:"referrer_uuid"`
	ConsumerUuid      uint64  `gorm:"column:consumer_uuid;type:biginteger;default:0;comment:消费人uuid" json:"consumer_uuid"`
	ConsumptionAmount float64 `gorm:"column:consumption_amount;type:decimal(14,2);default:0;comment:消费金额" json:"consumption_amount"`
	RewardAmount      float64 `gorm:"column:reward_amount;type:decimal(14,2);default:0;comment:奖励金额" json:"reward_amount"`
	RewardStatus      int     `gorm:"column:reward_status;type:int;default:0;comment:奖励状态 0未发放 1已发放" json:"reward_status"`
	CreateTime        int     `gorm:"column:create_time;type:int;default:0;comment:创建时间" json:"create_time"`
	UpdateTime        int     `gorm:"column:update_time;type:int;default:0;comment:更新时间" json:"update_time"`
	DeleteTime        int     `gorm:"column:delete_time;type:int;default:0;comment:删除时间" json:"delete_time"`
}

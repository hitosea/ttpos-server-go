package model

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/utils"
)

// MarketingActivity 营销活动模型 `ttpos_marketing_activity`
type MarketingActivity struct {
	BaseModel
	Name                  string  `gorm:"column:name;type:varchar(2500);default:'';comment:活动名称" json:"name"`
	Type                  int     `gorm:"column:type;type:tinyint(1);default:0;comment:活动类型 0邀请有礼 1积分商城" json:"type"`
	MultiLanguageNameUuid uint64  `gorm:"column:multi_language_name_uuid;type:biginteger;default:0;comment:活动名称多语言uuid" json:"multi_language_name_uuid"`
	Description           string  `gorm:"column:description;type:varchar(5000);default:'';comment:活动文案" json:"description"`
	MultiLanguageDescUuid uint64  `gorm:"column:multi_language_desc_uuid;type:biginteger;default:0;comment:活动文案多语言uuid" json:"multi_language_desc_uuid"`
	StartTime             int     `gorm:"column:start_time;type:int;default:0;comment:活动开始时间" json:"start_time"`
	EndTime               int     `gorm:"column:end_time;type:int;default:0;comment:活动结束时间" json:"end_time"`
	RewardType            int     `gorm:"column:reward_type;type:tinyint(1);default:0;comment:奖励类型 0优惠券 1积分" json:"reward_type"`
	RewardValue           float64 `gorm:"column:reward_value;type:decimal(14,2);default:0;comment:奖励值" json:"reward_value"`
	IsSendSms             int     `gorm:"column:is_send_sms;type:tinyint(1);default:0;comment:是否发送短信通知 0否 1是" json:"is_send_sms"`
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

/**
 * 获取状态文字
 */
func (m *MarketingActivity) GetStatus() int {
	if m.IsInvalid == 1 {
		return 2
	}
	now := time.Now().Unix()
	if int64(m.StartTime) > now {
		return 0
	} else if int64(m.EndTime) < now {
		return 2
	} else {
		return 1
	}
}

// 是否正在进行中
func (m *MarketingActivity) IsValid() bool {
	return m.IsInvalid == 0 && time.Now().Unix() > int64(m.StartTime) && time.Now().Unix() < int64(m.EndTime)
}

// MarketingActivityPrize 营销活动奖品模型 `ttpos_marketing_activity_prize`
type MarketingActivityPrize struct {
	BaseModel
	ActivityUuid uint64           `gorm:"column:activity_uuid;type:biginteger;default:0;comment:活动uuid" json:"activity_uuid"`
	PrizeType    int              `gorm:"column:prize_type;type:tinyint(1);default:0;comment:奖品类型 1优惠券 2未知" json:"prize_type"`
	PrizeUuid    uint64           `gorm:"column:prize_uuid;type:biginteger;default:0;comment:奖品uuid" json:"prize_uuid"`
	Coupon       *MarketingCoupon `gorm:"foreignKey:Uuid;references:PrizeUuid" json:"coupon"`
}

// MarketingActivityRecord 营销活动记录模型 `ttpos_marketing_activity_record`
type MarketingActivityRecord struct {
	BaseModel
	ActivityUuid   uint64  `gorm:"column:activity_uuid;type:biginteger;default:0;comment:活动uuid" json:"activity_uuid"`
	PrizeUuid      uint64  `gorm:"column:prize_uuid;type:biginteger;default:0;comment:奖品uuid" json:"prize_uuid"`
	MemberUuid     uint64  `gorm:"column:member_uuid;type:biginteger;default:0;comment:会员uuid" json:"member_uuid"`
	RewardCount    int     `gorm:"column:reward_count;type:int;default:0;comment:已获得奖励次数" json:"reward_count"`
	RewardValue    float64 `gorm:"column:reward_value;type:decimal(14,2);default:0;comment:奖励值" json:"reward_value"`
	LastRewardTime int64   `gorm:"column:last_reward_time;type:int;default:0;comment:最后一次获得奖励时间" json:"last_reward_time"`
}

// 营销活动消费记录表 `ttpos_marketing_activity_consumption`
type MarketingActivityConsumption struct {
	BaseModel
	ActivityUuid      uint64  `gorm:"column:activity_uuid;type:biginteger;default:0;comment:活动uuid" json:"activity_uuid"`
	ReferrerUuid      uint64  `gorm:"column:referrer_uuid;type:biginteger;default:0;comment:推荐人uuid" json:"referrer_uuid"`
	ConsumerUuid      uint64  `gorm:"column:consumer_uuid;type:biginteger;default:0;comment:消费人uuid" json:"consumer_uuid"`
	ConsumptionAmount float64 `gorm:"column:consumption_amount;type:decimal(14,2);default:0;comment:消费金额" json:"consumption_amount"`
	RewardAmount      float64 `gorm:"column:reward_amount;type:decimal(14,2);default:0;comment:奖励金额" json:"reward_amount"`
	RewardStatus      int     `gorm:"column:reward_status;type:int;default:0;comment:奖励状态 0未发放 1已发放" json:"reward_status"`
}

// MarketingCoupon 会员营销-优惠券表 `ttpos_marketing_coupon`
type MarketingCoupon struct {
	BaseModel
	Name           string  `gorm:"column:name;type:varchar(50);comment:优惠券名称" json:"name"`
	Sort           int     `gorm:"column:sort;type:int(11);default:0;comment:排序, 1-99" json:"sort"`
	Type           string  `gorm:"column:type;type:varchar(20);comment:优惠券类型: deduction - 抵扣券" json:"type"`
	DeductionType  string  `gorm:"column:deduction_type;type:varchar(20);comment:抵扣类型: taxed - 税后抵扣" json:"deduction_type"`
	Amount         float64 `gorm:"column:amount;type:decimal(14,2);default:0.00;comment:优惠券金额" json:"amount"`
	Count          int     `gorm:"column:count;type:int(11);default:0;comment:优惠券数量, 最大999999" json:"count"`
	DayStartTime   string  `gorm:"column:day_start_time;type:varchar(5);comment:每日适用时段开始时间, hh:mm 格式" json:"day_start_time"`
	DayEndTime     string  `gorm:"column:day_end_time;type:varchar(5);comment:每日适用时段结束时间, hh:mm 格式" json:"day_end_time"`
	Requirement    string  `gorm:"column:requirement;type:varchar(20);comment:获得优惠券所需条件: none - 都可以获取; marketing - 营销活动" json:"requirement"`
	ValidStartTime int     `gorm:"column:valid_start_time;type:int(11);default:0;comment:优惠券有效开始时间, requirement = none 时有效" json:"valid_start_time"`
	ValidEndTime   int     `gorm:"column:valid_end_time;type:int(11);default:0;comment:优惠券有效结束时间, requirement = none 时有效" json:"valid_end_time"`
	ValidDays      int     `gorm:"column:valid_days;type:int(11);default:0;comment:领取优惠券后n天内有效, requirement = marketing 时有效" json:"valid_days"`
	Status         int     `gorm:"column:status;type:int(11);default:1;comment:优惠券状态 0禁用 1开启" json:"status"`
}

// 判断优惠券是否可用
// nowTime 订单点击结账时该商家的时区实时时间，格式：HH:mm
func (model *MarketingCoupon) IsAvailable(nowTime string) error {
	if model.Requirement != constant.CouponRequirementNone {
		return errors.WithMessage(errors.New("无效优惠券"))
	}
	if model.IsExpire() {
		return errors.WithMessage(errors.New("优惠券已过期"))
	}
	if model.IsCountZero() {
		return errors.WithMessage(errors.New("优惠券已用完"))
	}
	if model.IsDelete() {
		return errors.WithMessage(errors.New("优惠券已删除"))
	}
	// 不在使用时间区间内
	inRange, err := utils.IsTimeInRange(nowTime, model.DayStartTime, model.DayEndTime)
	if err != nil {
		return errors.WithMessage(err)
	}
	if !inRange {
		return errors.WithMessage(errors.New("优惠券不在使用时间区间内"))
	}
	return nil
}

// 判断优惠券是否过期
func (model *MarketingCoupon) IsExpire() bool {
	nowTimestamp := time.Now().Unix()
	if model.ValidStartTime <= int(nowTimestamp) && int(nowTimestamp) <= model.ValidEndTime {
		return false
	}
	return true
}

// 判断优惠券是否已经使用
func (model *MarketingCoupon) IsCountZero() bool {
	return model.Count <= 0
}

// MarketingCouponRecord 会员营销-优惠券表 `ttpos_marketing_coupon_record`
type MarketingCouponRecord struct {
	BaseModel
	CouponUuid   uint64 `gorm:"column:coupon_uuid;type:bigint(20);default:0;comment:优惠券唯一ID" json:"coupon_uuid"`
	SerialNo     string `gorm:"column:serial_no;type:varchar(255);comment:记录编号, yyMMddhhmmssxxxx, 比如2506061456550001这样, 后四位是0000到9999依次递增, 循环使用" json:"serial_no"`
	ActivityUuid uint64 `gorm:"column:activity_uuid;type:bigint(20);default:0;comment:活动uuid" json:"activity_uuid"`
	MemberUuid   uint64 `gorm:"column:member_uuid;type:bigint(20);default:0;comment:会员唯一ID" json:"member_uuid"`
	Type         int    `gorm:"column:type;type:int(11);default:1;comment:记录类型：1-首次添加、2-调整添加、3-调整扣减、4-反结账退还、5、奖励领取（冻结）、6、核销扣减" json:"type"`
	Count        int    `gorm:"column:count;type:int(11);default:0;comment:变动数量" json:"count"`
	LeftCount    int    `gorm:"column:left_count;type:int(11);default:0;comment:剩余有效张数" json:"left_count"`
}

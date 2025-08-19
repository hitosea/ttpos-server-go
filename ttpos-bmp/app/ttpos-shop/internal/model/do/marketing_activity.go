// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// MarketingActivity is the golang structure of table ttpos_marketing_activity for DAO operations like Where/Data.
type MarketingActivity struct {
	g.Meta                `orm:"table:ttpos_marketing_activity, do:true"`
	Id                    interface{} //
	Uuid                  interface{} // 活动唯一ID
	Name                  interface{} // 活动名称
	Type                  interface{} // 活动类型 0邀请有礼 1积分商城
	MultiLanguageNameUuid interface{} // 活动名称多语言uuid
	Description           interface{} // 活动描述
	MultiLanguageDescUuid interface{} // 活动文案多语言uuid
	StartTime             interface{} // 活动开始时间
	EndTime               interface{} // 活动结束时间
	RewardConditionAmount interface{} // 奖励条件金额
	IsOpenRewardLimit     interface{} // 是否开启奖励次数限制 0否 1是
	RewardLimit           interface{} // 奖励次数限制
	IsInvalid             interface{} // 是否失效 0否 1是
	ImageBase64           interface{} // 活动图片base64
	CreateTime            interface{} // 创建时间
	UpdateTime            interface{} // 更新时间
	DeleteTime            interface{} // 删除时间
}

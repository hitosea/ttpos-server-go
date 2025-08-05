// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// BuffetPackage is the golang structure for table buffet_package.
type BuffetPackage struct {
	Id                    uint   `json:"id"                    orm:"id"                       description:"自增ID"`           // 自增ID
	Uuid                  uint64 `json:"uuid"                  orm:"uuid"                     description:"自助餐套餐ID"`        // 自助餐套餐ID
	Name                  string `json:"name"                  orm:"name"                     description:"自助餐套餐名称"`        // 自助餐套餐名称
	MultiLanguageNameUuid uint64 `json:"multiLanguageNameUuid" orm:"multi_language_name_uuid" description:"多语言名称ID"`        // 多语言名称ID
	Sort                  int    `json:"sort"                  orm:"sort"                     description:"排序顺序"`           // 排序顺序
	TaxUuid               uint64 `json:"taxUuid"               orm:"tax_uuid"                 description:"税收ID"`           // 税收ID
	IsLimitTime           int    `json:"isLimitTime"           orm:"is_limit_time"            description:"是否限时, 0-否 1-是"`  // 是否限时, 0-否 1-是
	LimitTime             int    `json:"limitTime"             orm:"limit_time"               description:"限时时间(分钟)"`       // 限时时间(分钟)
	CanCombined           int    `json:"canCombined"           orm:"can_combined"             description:"是否可合并, 0-否 1-是"` // 是否可合并, 0-否 1-是
	NonOrderingTime       int    `json:"nonOrderingTime"       orm:"non_ordering_time"        description:"平板不可下单时间(分钟)"`   // 平板不可下单时间(分钟)
	ReminderOrderTime     int    `json:"reminderOrderTime"     orm:"reminder_order_time"      description:"平板提醒不可下单时间(分钟)"` // 平板提醒不可下单时间(分钟)
	Status                int    `json:"status"                orm:"status"                   description:"状态 0-禁用 1-启用"`   // 状态 0-禁用 1-启用
	CreateTime            uint   `json:"createTime"            orm:"create_time"              description:"创建时间(时间戳)"`      // 创建时间(时间戳)
	UpdateTime            uint   `json:"updateTime"            orm:"update_time"              description:"更新时间(时间戳)"`      // 更新时间(时间戳)
	DeleteTime            uint   `json:"deleteTime"            orm:"delete_time"              description:"删除时间(时间戳)"`      // 删除时间(时间戳)
}

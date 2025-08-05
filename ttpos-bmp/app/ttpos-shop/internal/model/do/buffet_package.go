// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// BuffetPackage is the golang structure of table ttpos_buffet_package for DAO operations like Where/Data.
type BuffetPackage struct {
	g.Meta                `orm:"table:ttpos_buffet_package, do:true"`
	Id                    interface{} // 自增ID
	Uuid                  interface{} // 自助餐套餐ID
	Name                  interface{} // 自助餐套餐名称
	MultiLanguageNameUuid interface{} // 多语言名称ID
	Sort                  interface{} // 排序顺序
	TaxUuid               interface{} // 税收ID
	IsLimitTime           interface{} // 是否限时, 0-否 1-是
	LimitTime             interface{} // 限时时间(分钟)
	CanCombined           interface{} // 是否可合并, 0-否 1-是
	NonOrderingTime       interface{} // 平板不可下单时间(分钟)
	ReminderOrderTime     interface{} // 平板提醒不可下单时间(分钟)
	Status                interface{} // 状态 0-禁用 1-启用
	CreateTime            interface{} // 创建时间(时间戳)
	UpdateTime            interface{} // 更新时间(时间戳)
	DeleteTime            interface{} // 删除时间(时间戳)
}

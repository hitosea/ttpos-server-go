// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ReturnFoodReason is the golang structure of table ttpos_return_food_reason for DAO operations like Where/Data.
type ReturnFoodReason struct {
	g.Meta                `orm:"table:ttpos_return_food_reason, do:true"`
	Id                    interface{} // 自增ID
	Uuid                  interface{} // 退菜原因ID
	Name                  interface{} // 名称
	MultiLanguageNameUuid interface{} // 多语言名称ID
	CreateTime            interface{} // 创建时间(时间戳)
	UpdateTime            interface{} // 更新时间(时间戳)
	DeleteTime            interface{} // 删除时间(时间戳)
}

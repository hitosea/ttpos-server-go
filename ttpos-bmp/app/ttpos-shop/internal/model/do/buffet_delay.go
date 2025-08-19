// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// BuffetDelay is the golang structure of table ttpos_buffet_delay for DAO operations like Where/Data.
type BuffetDelay struct {
	g.Meta     `orm:"table:ttpos_buffet_delay, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // 自助餐加钟价格ID
	Name       interface{} // 自助餐加钟商品名称
	DelayTime  interface{} // 加钟时间(分钟)
	Price      interface{} // 价格
	Status     interface{} // 状态 0-禁用 1-启用
	CreateTime interface{} // 创建时间(时间戳)
	UpdateTime interface{} // 更新时间(时间戳)
	DeleteTime interface{} // 删除时间(时间戳)
}

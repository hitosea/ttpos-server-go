// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeskType is the golang structure of table ttpos_desk_type for DAO operations like Where/Data.
type DeskType struct {
	g.Meta     `orm:"table:ttpos_desk_type, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // 餐桌类型ID
	Name       interface{} // 餐桌类型名称
	Sort       interface{} // 排序序号
	RangeMin   interface{} // 最少人数
	RangeMax   interface{} // 最多人数
	CreateTime interface{} // 创建时间(时间戳)
	UpdateTime interface{} // 更新时间(时间戳)
	DeleteTime interface{} // 删除时间(时间戳)
}

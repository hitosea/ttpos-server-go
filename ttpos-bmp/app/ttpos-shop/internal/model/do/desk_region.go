// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// DeskRegion is the golang structure of table ttpos_desk_region for DAO operations like Where/Data.
type DeskRegion struct {
	g.Meta     `orm:"table:ttpos_desk_region, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // 餐桌区域ID
	Name       interface{} // 餐桌区域名称
	Sort       interface{} // 排序序号
	CreateTime interface{} // 创建时间(时间戳)
	UpdateTime interface{} // 更新时间(时间戳)
	DeleteTime interface{} // 删除时间(时间戳)
}

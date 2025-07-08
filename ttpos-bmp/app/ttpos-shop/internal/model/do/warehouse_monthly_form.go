// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// WarehouseMonthlyForm is the golang structure of table ttpos_warehouse_monthly_form for DAO operations like Where/Data.
type WarehouseMonthlyForm struct {
	g.Meta     `orm:"table:ttpos_warehouse_monthly_form, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // 月度报表ID
	Year       interface{} // 年
	Month      interface{} // 月
	Scene      interface{} // 记录类型,0-月初 1-月末
	Stock      interface{} // 库存
	CreateTime interface{} // 创建时间(时间戳)
	UpdateTime interface{} // 更新时间(时间戳)
	DeleteTime interface{} // 删除时间(时间戳)
}

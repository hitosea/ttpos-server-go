// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Supplier is the golang structure of table ttpos_supplier for DAO operations like Where/Data.
type Supplier struct {
	g.Meta       `orm:"table:ttpos_supplier, do:true"`
	Id           interface{} // 自增ID
	Uuid         interface{} // 供应商ID
	Name         interface{} // 供应商名称
	Address      interface{} // 供应商地址
	ContactName  interface{} // 联系人姓名
	ContactPhone interface{} // 联系人电话
	Position     interface{} // 职位
	StaffUuid    interface{} // 员工ID, 采购负责人
	CreateTime   interface{} // 创建时间(时间戳)
	UpdateTime   interface{} // 更新时间(时间戳)
	DeleteTime   interface{} // 删除时间(时间戳)
}

// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// BuffetCustomerType is the golang structure of table ttpos_buffet_customer_type for DAO operations like Where/Data.
type BuffetCustomerType struct {
	g.Meta     `orm:"table:ttpos_buffet_customer_type, do:true"`
	Id         interface{} // 自增ID
	Uuid       interface{} // 自助餐客户类型ID
	Name       interface{} // 自助餐客户类型名称
	CreateTime interface{} // 创建时间(时间戳)
	UpdateTime interface{} // 更新时间(时间戳)
	DeleteTime interface{} // 删除时间(时间戳)
}

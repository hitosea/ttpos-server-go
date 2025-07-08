// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// BuffetCustomerTypePrice is the golang structure of table ttpos_buffet_customer_type_price for DAO operations like Where/Data.
type BuffetCustomerTypePrice struct {
	g.Meta            `orm:"table:ttpos_buffet_customer_type_price, do:true"`
	Id                interface{} // 自增ID
	Uuid              interface{} // 自助餐顾客类型价格ID
	BuffetPackageUuid interface{} // 自助餐套餐ID
	CustomerTypeUuid  interface{} // 客户类型ID
	Price             interface{} // 价格
	CreateTime        interface{} // 创建时间(时间戳)
	UpdateTime        interface{} // 更新时间(时间戳)
	DeleteTime        interface{} // 删除时间(时间戳)
}

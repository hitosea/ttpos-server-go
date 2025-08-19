// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductSauce is the golang structure of table ttpos_product_sauce for DAO operations like Where/Data.
type ProductSauce struct {
	g.Meta                `orm:"table:ttpos_product_sauce, do:true"`
	Id                    interface{} // 自增ID
	Uuid                  interface{} // 商品小料ID
	Name                  interface{} // 名称
	MultiLanguageNameUuid interface{} // 多语言名称ID
	Price                 interface{} // 价格
	CreateTime            interface{} // 创建时间(时间戳)
	UpdateTime            interface{} // 更新时间(时间戳)
	DeleteTime            interface{} // 删除时间(时间戳)
}

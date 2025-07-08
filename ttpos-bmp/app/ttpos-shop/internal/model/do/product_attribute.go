// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductAttribute is the golang structure of table ttpos_product_attribute for DAO operations like Where/Data.
type ProductAttribute struct {
	g.Meta                `orm:"table:ttpos_product_attribute, do:true"`
	Id                    interface{} // 自增ID
	Uuid                  interface{} // 商品属性ID
	Name                  interface{} // 名称
	MultiLanguageNameUuid interface{} // 多语言名称ID
	AttributeGroupUuid    interface{} // 属性组ID
	CreateTime            interface{} // 创建时间(时间戳)
	UpdateTime            interface{} // 更新时间(时间戳)
	DeleteTime            interface{} // 删除时间(时间戳)
}
